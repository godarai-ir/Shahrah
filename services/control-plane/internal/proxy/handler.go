package proxy

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	metricsobs "github.com/Delpak/shahrah/services/control-plane/internal/observability/metrics"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/Delpak/shahrah/services/control-plane/internal/responsecache"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/Delpak/shahrah/services/control-plane/internal/virtualmodels"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// HTTP status constant used by sync.go for error floor checks.
const httpErrorFloor = 400

type budgeter interface {
	EstimateCost(ctx context.Context, req contractsopenai.ChatCompletionRequest) (platform.UsageCostBreakdown, error)
	EstimateEmbeddingCost(ctx context.Context, req contractsopenai.EmbeddingRequest) (platform.UsageCostBreakdown, error)
	CheckRateLimit(ctx context.Context, auth *platform.AuthContext, model string, estimatedTokens int) (*cache.RateLimitResult, error)
	CheckBudget(ctx context.Context, auth *platform.AuthContext, breakdown platform.UsageCostBreakdown) error
	FinalizeFromUsage(ctx context.Context, model string, usage *contractsopenai.Usage, estimated platform.UsageCostBreakdown) (platform.UsageCostBreakdown, error)
}

type walletStore interface {
	DebitWallet(ctx context.Context, tenantID string, breakdown platform.UsageCostBreakdown, requestID string) (*platform.Transaction, error)
}

type usageLogger interface {
	LogRequest(ctx context.Context, entry platform.RequestLogEntry) error
	LogSpend(ctx context.Context, log platform.SpendLog) error
}

// ChatCompletionHandler handles chat completion requests.
type ChatCompletionHandler struct {
	llmRouter     *router.Router
	tunnel        platform.TunnelTransport
	billing       budgeter
	wallets       walletStore
	logs          usageLogger
	logger        *zap.Logger
	vmResolver    virtualmodels.VirtualModelResolver
	vmEngine      virtualmodels.RoutingEngine
	vmExec        virtualmodels.FallbackExecutor
	emitMetric    func(input metricsobs.BuildInput)
	responseCache *responsecache.Service
}

// SetResponseCache enables opt-in response caching without changing existing constructor callers.
func (h *ChatCompletionHandler) SetResponseCache(service *responsecache.Service) {
	h.responseCache = service
}

// NewChatCompletionHandler creates a new handler.
func NewChatCompletionHandler(llmRouter *router.Router, tunnel platform.TunnelTransport, billing budgeter, wallets walletStore, logs usageLogger, logger *zap.Logger, emitMetric ...func(input metricsobs.BuildInput)) *ChatCompletionHandler {
	var resolver virtualmodels.VirtualModelResolver
	var engine virtualmodels.RoutingEngine
	var exec virtualmodels.FallbackExecutor
	if repo, ok := logs.(interface {
		GetVirtualModelByAlias(ctx context.Context, tenantID string, alias string) (*platform.VirtualModel, error)
	}); ok {
		resolver = virtualmodels.NewResolver(repo)
		engine = virtualmodels.NewPriorityRoutingEngine()
		exec = virtualmodels.NewExecutor(virtualmodels.NewDefaultErrorNormalizer())
	}
	var metricEmitter func(input metricsobs.BuildInput)
	if len(emitMetric) > 0 {
		metricEmitter = emitMetric[0]
	}
	return &ChatCompletionHandler{
		llmRouter:  llmRouter,
		tunnel:     tunnel,
		billing:    billing,
		wallets:    wallets,
		logs:       logs,
		logger:     logger,
		vmResolver: resolver,
		vmEngine:   engine,
		vmExec:     exec,
		emitMetric: metricEmitter,
	}
}

// Handle processes a chat completion request.
func (h *ChatCompletionHandler) Handle(c *fiber.Ctx) error {
	startTime := time.Now()
	authInfo := auth.GetAuthInfo(c)
	if authInfo == nil {
		return writeAPIError(c, fiber.StatusUnauthorized, "Authentication required", "authentication_error")
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return writeAPIError(c, fiber.StatusBadRequest, "Tenant resolution required", "invalid_request_error")
	}
	authInfo.TenantID = currentTenant.ID
	tracing.SetTenantForTrace(c, currentTenant.ID)
	tenantSlug := strings.TrimSpace(currentTenant.Slug)

	var req contractsopenai.ChatCompletionRequest
	if err := c.BodyParser(&req); err != nil {
		return writeAPIError(c, fiber.StatusBadRequest, "Invalid request body", "invalid_request_error")
	}
	if req.Model == "" {
		return writeAPIError(c, fiber.StatusBadRequest, "Model is required", "invalid_request_error")
	}
	requestedModel := normalizeModelIdentifier(req.Model)
	routeCandidates := modelRoutingCandidates(requestedModel)
	if len(routeCandidates) == 0 {
		return writeAPIError(c, fiber.StatusBadRequest, "Model is required", "invalid_request_error")
	}
	resolvedModel := providerModelID(requestedModel)
	tracing.SetModelForTrace(c, requestedModel)
	if len(req.Messages) == 0 {
		return writeAPIError(c, fiber.StatusBadRequest, "messages array must not be empty", "invalid_request_error")
	}
	for i, msg := range req.Messages {
		if msg.Role == "" {
			return writeAPIError(c, fiber.StatusBadRequest, fmt.Sprintf("message[%d].role is required", i), "invalid_request_error")
		}
		if err := contractsopenai.ValidateChatMessageContent(msg.Content); err != nil {
			return writeAPIError(c, fiber.StatusBadRequest, fmt.Sprintf("message[%d].content: %v", i, err), "invalid_request_error")
		}
	}
	if req.Temperature != nil && (*req.Temperature < 0 || *req.Temperature > 2) {
		return writeAPIError(c, fiber.StatusBadRequest, "temperature must be between 0 and 2", "invalid_request_error")
	}
	if req.TopP != nil && (*req.TopP < 0 || *req.TopP > 1) {
		return writeAPIError(c, fiber.StatusBadRequest, "top_p must be between 0 and 1", "invalid_request_error")
	}
	if req.MaxTokens != nil && *req.MaxTokens < 1 {
		return writeAPIError(c, fiber.StatusBadRequest, "max_tokens must be at least 1", "invalid_request_error")
	}
	modelAccess := modelAccessForType(authInfo, requestedModel, platform.ModelTypeChat)
	if !modelAccess.Allowed {
		if modelAccess.UnsupportedVMMode {
			return writeUnsupportedModelModeError(c, requestedModel, platform.ModelTypeChat, modelAccess.ExposedModes)
		}
		emitDenyCounter(h.logger, denyReasonForModel(authInfo), requestedModel, currentTenant.ID)
		return writeAPIError(c, fiber.StatusForbidden, "Model access denied or unsupported for chat completions", "authorization_error")
	}
	var cacheConfig *responsecache.RequestConfig
	if rawCacheConfig := strings.TrimSpace(c.Get("x-shahrah-cache-config")); rawCacheConfig != "" && h.responseCache != nil {
		parsed, parseErr := h.responseCache.ParseHeader(rawCacheConfig)
		if parseErr != nil {
			return writeAPIError(c, fiber.StatusBadRequest, parseErr.Error(), "invalid_request_error")
		}
		cacheConfig = &parsed
		if !h.responseCache.Enabled() || req.Stream {
			c.Set("x-shahrah-cache-status", "bypass")
			cacheConfig = nil
		}
	}
	authStart := time.Now()
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "auth.validate", "internal", authStart, fiber.StatusOK, map[string]interface{}{"model": requestedModel})

	quotaStart := time.Now()
	estimateReq := req
	estimateReq.Model = resolvedModel
	estimated, err := h.billing.EstimateCost(c.Context(), estimateReq)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, "Failed to estimate request cost", "billing_error")
	}

	rateLimit, err := h.billing.CheckRateLimit(c.Context(), authInfo, requestedModel, estimated.TotalTokens)
	setRateLimitHeaders(c, rateLimit)
	if err != nil {
		return writeAPIError(c, rateLimitErrorStatus(err), err.Error(), "rate_limit_error")
	}
	var cachePending *responsecache.Pending
	if cacheConfig != nil {
		keyID := strings.TrimSpace(authInfo.KeyID)
		if keyID == "" {
			keyID = strings.TrimSpace(authInfo.TokenHash)
		}
		cacheScope := responsecache.Scope{
			TenantID:     currentTenant.ID,
			KeyID:        keyID,
			Model:        requestedModel,
			ModelVersion: h.virtualModelCacheVersion(c.Context(), authInfo, currentTenant.ID, requestedModel),
		}
		cacheResult, pending, lookupErr := h.responseCache.Lookup(c.Context(), cacheScope, req, *cacheConfig)
		cachePending = pending
		c.Set("x-shahrah-cache-status", cacheResult.Status)
		c.Locals("response_cache_status", cacheResult.Status)
		c.Locals("response_cache_type", cacheConfig.Type)
		c.Locals("response_cache_embedding_latency_ms", cacheResult.EmbeddingLatencyMS)
		if lookupErr != nil {
			h.logger.Warn("Response cache lookup failed", zap.String("cache_type", cacheConfig.Type), zap.Error(lookupErr))
		} else if cacheResult.Entry != nil {
			return h.handleCacheHit(c, req, authInfo, cacheResult, startTime, tenantSlug)
		}
		if cacheResult.Status != "bypass" {
			release, waited, acquireErr := h.responseCache.AcquireMiss(c.Context(), cachePending)
			if acquireErr != nil {
				return writeAPIError(c, fiber.StatusRequestTimeout, "Request canceled while waiting for response cache", "request_canceled")
			}
			defer release()
			if waited {
				secondResult, secondPending, secondErr := h.responseCache.Lookup(c.Context(), cacheScope, req, *cacheConfig)
				cachePending = secondPending
				if secondErr == nil && secondResult.Entry != nil {
					c.Set("x-shahrah-cache-status", "hit")
					return h.handleCacheHit(c, req, authInfo, secondResult, startTime, tenantSlug)
				}
			}
		}
	}
	if err := h.billing.CheckBudget(c.Context(), authInfo, estimated); err != nil {
		return writeAPIError(c, fiber.StatusPaymentRequired, err.Error(), "budget_exceeded")
	}
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "quota.check", "internal", quotaStart, fiber.StatusOK, map[string]interface{}{"estimated_tokens": estimated.TotalTokens})

	if err := h.tunnel.Healthy(c.Context()); err != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Tunnel path unavailable", "tunnel_unavailable")
	}

	requestID := tracing.TraceIDFromContext(c)
	if requestID == "" {
		requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
	}

	baseReq := platform.TunnelRequest{
		RequestID: requestID,
		TenantID:  currentTenant.ID,
		TraceID:   requestID,
		Stream:    req.Stream,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		RequestMetadata: enrichResponseCacheMetadata(c, map[string]interface{}{
			"tenant_slug": strings.TrimSpace(c.Get("x-shahrah-tenant")),
			"model":       requestedModel,
		}),
	}

	routeStart := time.Now()
	var (
		deployment *router.Deployment
	)
	isVirtual := false
	if h.vmEngine != nil && h.vmExec != nil && !req.Stream {
		if snapshotVM, ok := buildVirtualModelFromSnapshot(authInfo, requestedModel); ok {
			targets, selErr := h.vmEngine.SelectTargets(snapshotVM)
			if selErr == nil {
				resp, selectedDep, fallbackCount, failureReason, execErr := h.vmExec.ExecuteSync(c.Context(), targets, func(callCtx context.Context, dep *router.Deployment) (*platform.TunnelResponse, error) {
					dep = h.resolveVirtualDeployment(callCtx, currentTenant.ID, dep)
					tunnelReq := baseReq
					tunnelReq.Deployment = toContractDeployment(dep)
					payload, err := buildChatCompletionPayload(req, dep.Model, resolvedModel)
					if err != nil {
						return nil, err
					}
					tunnelReq.Body = payload
					return h.tunnel.Proxy(callCtx, tunnelReq)
				})
				if execErr == nil && resp != nil && selectedDep != nil {
					isVirtual = true
					deployment = selectedDep
					c.Locals("vm_resp", resp)
					c.Locals("vm_fallback_count", fallbackCount)
					c.Locals("vm_failure_reason", failureReason)
					tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "router.select_virtual_model", "internal", routeStart, fiber.StatusOK, map[string]interface{}{
						"virtual_model":     snapshotVM.Alias,
						"selected_provider": selectedDep.Provider,
						"selected_model":    selectedDep.Model,
						"fallback_count":    fallbackCount,
						"failure_reason":    failureReason,
						"snapshot":          true,
					})
				} else if errors.Is(execErr, virtualmodels.ErrFallbackExhausted) {
					return writeAPIError(c, fiber.StatusBadGateway, "Virtual model fallback exhausted", "fallback_exhausted")
				}
			}
		}
	}
	if deployment == nil && h.vmResolver != nil && h.vmEngine != nil && h.vmExec != nil && !req.Stream {
		vm, vmErr := h.vmResolver.Resolve(c.Context(), currentTenant.ID, requestedModel)
		if vmErr == nil {
			targets, selErr := h.vmEngine.SelectTargets(vm)
			if selErr == nil {
				resp, selectedDep, fallbackCount, failureReason, execErr := h.vmExec.ExecuteSync(c.Context(), targets, func(callCtx context.Context, dep *router.Deployment) (*platform.TunnelResponse, error) {
					dep = h.resolveVirtualDeployment(callCtx, currentTenant.ID, dep)
					tunnelReq := baseReq
					tunnelReq.Deployment = toContractDeployment(dep)
					payload, err := buildChatCompletionPayload(req, dep.Model, resolvedModel)
					if err != nil {
						return nil, err
					}
					tunnelReq.Body = payload
					return h.tunnel.Proxy(callCtx, tunnelReq)
				})
				if execErr == nil && resp != nil && selectedDep != nil {
					isVirtual = true
					deployment = selectedDep
					c.Locals("vm_resp", resp)
					c.Locals("vm_fallback_count", fallbackCount)
					c.Locals("vm_failure_reason", failureReason)
					tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "router.select_virtual_model", "internal", routeStart, fiber.StatusOK, map[string]interface{}{
						"virtual_model":     vm.Alias,
						"selected_provider": selectedDep.Provider,
						"selected_model":    selectedDep.Model,
						"fallback_count":    fallbackCount,
						"failure_reason":    failureReason,
					})
				} else if errors.Is(execErr, virtualmodels.ErrFallbackExhausted) {
					return writeAPIError(c, fiber.StatusBadGateway, "Virtual model fallback exhausted", "fallback_exhausted")
				}
			}
		}
	}
	if deployment == nil {
		for _, candidate := range routeCandidates {
			deployment, err = h.llmRouter.Route(c.Context(), currentTenant.ID, candidate)
			if err == nil {
				break
			}
		}
	}
	if deployment == nil {
		emitDenyCounter(h.logger, denyReasonUnknownModel, requestedModel, currentTenant.ID)
		return writeAPIError(c, fiber.StatusServiceUnavailable, fmt.Sprintf("No available deployment for model: %s", requestedModel), "routing_error")
	}
	if chatRequestHasImage(req) && !supportsChatImageProvider(deployment.Provider) {
		return writeAPIError(c, fiber.StatusBadRequest, fmt.Sprintf("Provider %s does not support image input through Shahrah chat completions", deployment.Provider), "invalid_request_error")
	}
	if err := validateProviderImageInput(req, deployment.Provider); err != nil {
		return writeAPIError(c, fiber.StatusBadRequest, err.Error(), "invalid_request_error")
	}
	if !isVirtual {
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "router.select_model", "internal", routeStart, fiber.StatusOK, map[string]interface{}{"provider": deployment.Provider, "deployment_id": deployment.ID})
	}
	payload, err := buildChatCompletionPayload(req, deployment.Model, resolvedModel)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadRequest, "Failed to encode request", "invalid_request_error")
	}
	tunnelReq := baseReq
	tunnelReq.Body = payload
	tunnelReq.Deployment = toContractDeployment(deployment)

	if req.Stream {
		return h.handleStream(c, req, authInfo, estimated, tunnelReq, startTime, tenantSlug)
	}
	return h.handleSync(c, req, authInfo, estimated, tunnelReq, startTime, tenantSlug, cachePending)
}

func buildChatCompletionPayload(req contractsopenai.ChatCompletionRequest, deploymentModel string, fallbackModel string) ([]byte, error) {
	outboundReq := req
	outboundModel := strings.TrimSpace(deploymentModel)
	if outboundModel == "" || outboundModel == "*" {
		outboundModel = fallbackModel
	}
	outboundReq.Model = outboundModel
	return json.Marshal(outboundReq)
}

func (h *ChatCompletionHandler) resolveVirtualDeployment(ctx context.Context, tenantID string, dep *router.Deployment) *router.Deployment {
	if h == nil || h.llmRouter == nil || dep == nil {
		return dep
	}
	modelName := strings.TrimSpace(dep.ModelName)
	if modelName == "" {
		return dep
	}
	resolved, err := h.llmRouter.Route(ctx, tenantID, modelName)
	if err != nil || resolved == nil {
		return dep
	}
	*dep = *resolved
	return dep
}

func (h *ChatCompletionHandler) handleSync(c *fiber.Ctx, req contractsopenai.ChatCompletionRequest, authInfo *platform.AuthContext, estimated platform.UsageCostBreakdown, tunnelReq platform.TunnelRequest, start time.Time, tenantSlug string, cachePending *responsecache.Pending) error {
	providerStart := time.Now()
	providerAttrs := map[string]interface{}{
		"provider":        tunnelReq.Deployment.Provider,
		"request_payload": buildTraceChatRequestPayload(req),
	}
	var resp *platform.TunnelResponse
	if vmResp, ok := c.Locals("vm_resp").(*platform.TunnelResponse); ok && vmResp != nil {
		resp = vmResp
	} else {
		var err error
		resp, err = h.tunnel.Proxy(c.Context(), tunnelReq)
		if err != nil {
			providerAttrs["error"] = err.Error()
			tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, fiber.StatusBadGateway, providerAttrs)
			h.llmRouter.MarkFailed(c.Context(), &router.Deployment{ID: tunnelReq.Deployment.ID}, err.Error())
			return writeAPIError(c, fiber.StatusBadGateway, fmt.Sprintf("Tunnel error: %v", err), "api_error")
		}
	}

	if resp.StatusCode >= httpErrorFloor {
		providerAttrs["response_payload"] = buildTraceProviderErrorPayload(resp.StatusCode, resp.Body)
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, resp.StatusCode, providerAttrs)
		return transformProviderError(c, resp.StatusCode, resp.Body, resp.Headers)
	}

	var completion contractsopenai.ChatCompletionResponse
	if err := json.Unmarshal(resp.Body, &completion); err != nil {
		providerAttrs["error"] = "failed to decode provider response"
		providerAttrs["response_payload"] = buildTraceProviderErrorPayload(fiber.StatusBadGateway, resp.Body)
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, fiber.StatusBadGateway, providerAttrs)
		return writeAPIError(c, fiber.StatusBadGateway, "Failed to decode provider response", "api_error")
	}
	providerAttrs["response_payload"] = buildTraceChatResponsePayload(completion)
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, resp.StatusCode, providerAttrs)

	billingModel := billingModelForDeployment(req.Model, tunnelReq.Deployment)
	finalCost, err := h.billing.FinalizeFromUsage(c.Context(), billingModel, completion.Usage, estimated)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, "Failed to finalize usage", "billing_error")
	}
	if _, err := h.wallets.DebitWallet(c.Context(), authInfo.TenantID, finalCost, tunnelReq.RequestID); err != nil {
		return writeAPIError(c, fiber.StatusPaymentRequired, err.Error(), "budget_exceeded")
	}
	if cachePending != nil && h.responseCache != nil {
		cachedBody, marshalErr := json.Marshal(completion)
		if marshalErr == nil {
			if cacheErr := h.responseCache.Store(c.Context(), cachePending, cachedBody, tunnelReq.TraceID, tunnelReq.Deployment.Provider, tunnelReq.Deployment.Model); cacheErr != nil {
				h.logger.Warn("Response cache store failed", zap.Error(cacheErr))
			}
		}
	}

	responseTimeMS := int(time.Since(start).Milliseconds())
	_ = h.logs.LogRequest(c.Context(), platform.RequestLogEntry{
		RequestID:      tunnelReq.RequestID,
		TokenHash:      authInfo.TokenHash,
		TenantID:       authInfo.TenantID,
		Model:          req.Model,
		Status:         "success",
		CostUSD:        finalCost.TotalCostUSD,
		ResponseTimeMS: responseTimeMS,
		RequestBody:    map[string]interface{}{"model": req.Model},
		ResponseBody:   map[string]interface{}{"id": completion.ID},
	})
	_ = h.logs.LogSpend(c.Context(), platform.SpendLog{
		RequestID:       tunnelReq.RequestID,
		TokenHash:       authInfo.TokenHash,
		TenantID:        authInfo.TenantID,
		OrganizationID:  authInfo.OrganizationID,
		TeamID:          authInfo.TeamID,
		Model:           req.Model,
		InputTokens:     finalCost.InputTokens,
		OutputTokens:    finalCost.OutputTokens,
		TotalTokens:     finalCost.TotalTokens,
		CostUSD:         finalCost.TotalCostUSD,
		CostIRR:         finalCost.TotalCostIRR,
		ResponseTimeMS:  responseTimeMS,
		Status:          "success",
		RequestMetadata: enrichPricingMetadata(tunnelReq.RequestMetadata, authInfo, req.Model),
		CompletionStart: start,
		CompletionEnd:   time.Now(),
	})
	h.emitModelMetric(metricsobs.BuildInput{
		TS:                 time.Now().UTC(),
		TenantID:           authInfo.TenantID,
		RequestID:          tunnelReq.RequestID,
		ModelName:          req.Model,
		VirtualModelName:   modelMetricVirtualName(c, req.Model),
		ProviderName:       tunnelReq.Deployment.Provider,
		ProviderModelName:  tunnelReq.Deployment.Model,
		RequestType:        "chat_completions",
		UserID:             authInfo.UserID,
		VirtualAccountID:   authInfo.KeyID,
		VirtualAccountName: authInfo.KeyAlias,
		TeamIDs:            nonEmptySlice([]string{authInfo.TeamID}),
		TeamNames:          []string{},
		StatusCode:         resp.StatusCode,
		LatencyMS:          responseTimeMS,
		InputTokens:        finalCost.InputTokens,
		OutputTokens:       finalCost.OutputTokens,
		CostUSD:            finalCost.TotalCostUSD,
		Metadata: enrichResponseCacheMetadata(c, map[string]interface{}{
			"endpoint":    "/v1/chat/completions",
			"key_type":    authInfo.KeyType,
			"team_id":     authInfo.TeamID,
			"org_unit_id": authInfo.OrgUnitID,
			"tenant_slug": tenantSlug,
		}),
	})

	// Forward allowed provider response headers
	forwardProviderHeaders(c, resp.Headers)

	c.Set("x-shahrah-trace-id", tunnelReq.TraceID)
	c.Set("x-shahrah-cost-usd", fmt.Sprintf("%.8f", finalCost.TotalCostUSD))
	c.Set("x-shahrah-cost-irr", fmt.Sprintf("%.2f", finalCost.TotalCostIRR))
	responseStart := time.Now()
	err = c.Status(resp.StatusCode).JSON(completion)
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "response.write", "internal", responseStart, resp.StatusCode, map[string]interface{}{
		"response_payload": buildTraceChatResponsePayload(completion),
	})
	return err
}

func (h *ChatCompletionHandler) handleCacheHit(c *fiber.Ctx, req contractsopenai.ChatCompletionRequest, authInfo *platform.AuthContext, result responsecache.Result, start time.Time, tenantSlug string) error {
	entry := result.Entry
	requestID := tracing.TraceIDFromContext(c)
	if requestID == "" {
		requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
	}
	responseTimeMS := int(time.Since(start).Milliseconds())
	metadata := map[string]interface{}{
		"endpoint": "/v1/chat/completions", "cache_status": "hit", "cache_similarity": result.Similarity,
		"cache_type": result.CacheType, "cache_embedding_latency_ms": result.EmbeddingLatencyMS,
		"cached_trace_id": entry.TraceID, "tenant_slug": tenantSlug, "key_type": authInfo.KeyType,
	}
	_ = h.logs.LogRequest(c.Context(), platform.RequestLogEntry{
		RequestID: requestID, TokenHash: authInfo.TokenHash, TenantID: authInfo.TenantID, Model: req.Model,
		Status: "success", CostUSD: 0, ResponseTimeMS: responseTimeMS,
		RequestBody: map[string]interface{}{"model": req.Model}, ResponseBody: map[string]interface{}{"cache_status": "hit"},
	})
	_ = h.logs.LogSpend(c.Context(), platform.SpendLog{
		RequestID: requestID, TokenHash: authInfo.TokenHash, TenantID: authInfo.TenantID,
		OrganizationID: authInfo.OrganizationID, TeamID: authInfo.TeamID, Model: req.Model,
		CostUSD: 0, CostIRR: 0, ResponseTimeMS: responseTimeMS, Status: "success",
		RequestMetadata: metadata, CompletionStart: start, CompletionEnd: time.Now(),
	})
	h.emitModelMetric(metricsobs.BuildInput{
		TS: time.Now().UTC(), TenantID: authInfo.TenantID, RequestID: requestID, ModelName: req.Model,
		ProviderName: entry.Provider, ProviderModelName: entry.ProviderModel, RequestType: "chat_completions",
		UserID: authInfo.UserID, VirtualAccountID: authInfo.KeyID, VirtualAccountName: authInfo.KeyAlias,
		TeamIDs: nonEmptySlice([]string{authInfo.TeamID}), StatusCode: fiber.StatusOK, LatencyMS: responseTimeMS,
		CostUSD: 0, Metadata: metadata,
	})
	c.Set("Content-Type", fiber.MIMEApplicationJSON)
	c.Set("x-shahrah-cache-status", "hit")
	c.Set("x-shahrah-cached-trace-id", entry.TraceID)
	if result.Semantic {
		c.Set("x-shahrah-cache-similarity-score", fmt.Sprintf("%.6f", result.Similarity))
	}
	c.Set("x-shahrah-trace-id", requestID)
	c.Set("x-shahrah-cost-usd", "0.00000000")
	c.Set("x-shahrah-cost-irr", "0.00")
	return c.Status(fiber.StatusOK).Send(entry.Response)
}

func (h *ChatCompletionHandler) handleStream(c *fiber.Ctx, req contractsopenai.ChatCompletionRequest, authInfo *platform.AuthContext, estimated platform.UsageCostBreakdown, tunnelReq platform.TunnelRequest, start time.Time, tenantSlug string) error {
	providerStart := time.Now()
	stream, err := h.tunnel.Stream(c.Context(), tunnelReq)
	if err != nil {
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, fiber.StatusBadGateway, map[string]interface{}{
			"provider":        tunnelReq.Deployment.Provider,
			"request_payload": buildTraceChatRequestPayload(req),
			"stream":          true,
			"error":           err.Error(),
		})
		return writeAPIError(c, fiber.StatusBadGateway, fmt.Sprintf("Tunnel error: %v", err), "api_error")
	}
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, fiber.StatusOK, map[string]interface{}{
		"provider":        tunnelReq.Deployment.Provider,
		"request_payload": buildTraceChatRequestPayload(req),
		"stream":          true,
	})

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("x-shahrah-trace-id", tunnelReq.TraceID)

	// Forward provider headers from stream
	forwardProviderHeaders(c, stream.Headers())
	virtualModelName := modelMetricVirtualName(c, req.Model)

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		defer stream.Close()

		for {
			select {
			case chunk, ok := <-stream.Chunks():
				if !ok {
					billingModel := billingModelForDeployment(req.Model, tunnelReq.Deployment)
					finalCost, _ := h.billing.FinalizeFromUsage(c.Context(), billingModel, nil, estimated)
					_, _ = h.wallets.DebitWallet(c.Context(), authInfo.TenantID, finalCost, tunnelReq.RequestID)
					_ = h.logs.LogSpend(c.Context(), platform.SpendLog{
						RequestID:       tunnelReq.RequestID,
						TokenHash:       authInfo.TokenHash,
						TenantID:        authInfo.TenantID,
						OrganizationID:  authInfo.OrganizationID,
						TeamID:          authInfo.TeamID,
						Model:           req.Model,
						InputTokens:     finalCost.InputTokens,
						OutputTokens:    finalCost.OutputTokens,
						TotalTokens:     finalCost.TotalTokens,
						CostUSD:         finalCost.TotalCostUSD,
						CostIRR:         finalCost.TotalCostIRR,
						ResponseTimeMS:  int(time.Since(start).Milliseconds()),
						Status:          "success",
						RequestMetadata: enrichPricingMetadata(tunnelReq.RequestMetadata, authInfo, req.Model),
						CompletionStart: start,
						CompletionEnd:   time.Now(),
					})
					h.emitModelMetric(metricsobs.BuildInput{
						TS:                 time.Now().UTC(),
						TenantID:           authInfo.TenantID,
						RequestID:          tunnelReq.RequestID,
						ModelName:          req.Model,
						VirtualModelName:   virtualModelName,
						ProviderName:       tunnelReq.Deployment.Provider,
						ProviderModelName:  tunnelReq.Deployment.Model,
						RequestType:        "chat_completions",
						UserID:             authInfo.UserID,
						VirtualAccountID:   authInfo.KeyID,
						VirtualAccountName: authInfo.KeyAlias,
						TeamIDs:            nonEmptySlice([]string{authInfo.TeamID}),
						StatusCode:         fiber.StatusOK,
						LatencyMS:          int(time.Since(start).Milliseconds()),
						InputTokens:        finalCost.InputTokens,
						OutputTokens:       finalCost.OutputTokens,
						CostUSD:            finalCost.TotalCostUSD,
						Metadata: map[string]interface{}{
							"endpoint":    "/v1/chat/completions",
							"key_type":    authInfo.KeyType,
							"team_id":     authInfo.TeamID,
							"org_unit_id": authInfo.OrgUnitID,
							"tenant_slug": tenantSlug,
						},
					})
					return
				}
				if len(chunk.Data) > 0 {
					_, _ = w.Write(chunk.Data)
					_ = w.Flush()
				}
				if chunk.IsDone {
					return
				}
			case err, ok := <-stream.Errors():
				if ok && err != nil {
					errorPayload, _ := json.Marshal(contractsopenai.APIError{
						Error: contractsopenai.ErrorInfo{
							Message: err.Error(),
							Type:    "streaming_error",
						},
					})
					_, _ = w.Write(bytes.Join([][]byte{[]byte("data: "), errorPayload, []byte("\n\n")}, nil))
					_ = w.Flush()
				}
				return
			}
		}
	})
	responseStart := time.Now()
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "response.write", "internal", responseStart, fiber.StatusOK, map[string]interface{}{"stream": true})

	return nil
}

func (h *ChatCompletionHandler) emitModelMetric(input metricsobs.BuildInput) {
	if h.emitMetric == nil {
		return
	}
	h.emitMetric(input)
}

func modelMetricVirtualName(c *fiber.Ctx, fallback string) string {
	if c == nil {
		return ""
	}
	if v, ok := c.Locals("vm_failure_reason").(string); ok && v != "" {
		_ = v
	}
	if v, ok := c.Locals("vm_resp").(*platform.TunnelResponse); ok && v != nil {
		return fallback
	}
	return ""
}

func nonEmptySlice(in []string) []string {
	out := make([]string, 0, len(in))
	for _, v := range in {
		v = strings.TrimSpace(v)
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

func enrichResponseCacheMetadata(c *fiber.Ctx, metadata map[string]interface{}) map[string]interface{} {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	if c == nil {
		return metadata
	}
	if value := c.Locals("response_cache_status"); value != nil {
		metadata["cache_status"] = value
	}
	if value := c.Locals("response_cache_type"); value != nil {
		metadata["cache_type"] = value
	}
	if value := c.Locals("response_cache_embedding_latency_ms"); value != nil {
		metadata["cache_embedding_latency_ms"] = value
	}
	return metadata
}

func billingModelForDeployment(requestedModel string, deployment platform.Deployment) string {
	model := strings.TrimSpace(deployment.Model)
	if model != "" {
		return model
	}
	return strings.TrimSpace(requestedModel)
}

func (h *ChatCompletionHandler) virtualModelCacheVersion(ctx context.Context, authInfo *platform.AuthContext, tenantID, model string) string {
	if authInfo == nil {
		return ""
	}
	if snapshot, ok := authInfo.VirtualModelSnapshots[model]; ok {
		return fmt.Sprintf("%s:%d", snapshot.VirtualModelID, snapshot.Version)
	}
	if h.vmResolver != nil {
		if virtualModel, err := h.vmResolver.Resolve(ctx, tenantID, model); err == nil && virtualModel != nil {
			return fmt.Sprintf("%s:%d", virtualModel.ID, virtualModel.Version)
		}
	}
	return ""
}

func toContractDeployment(dep *router.Deployment) platform.Deployment {
	return platform.Deployment{
		ID:         dep.ID,
		TenantID:   dep.TenantID,
		ModelName:  dep.ModelName,
		Provider:   dep.Provider,
		Model:      dep.Model,
		APIBase:    dep.Params.APIBase,
		APIKey:     dep.Params.APIKey,
		APIVersion: dep.Params.APIVersion,
		RegionName: dep.Params.RegionName,
		RPM:        dep.Params.RPM,
		TPM:        dep.Params.TPM,
		Weight:     dep.Weight,
		TimeoutSec: dep.Params.Timeout,
	}
}

func setRateLimitHeaders(c *fiber.Ctx, result *cache.RateLimitResult) {
	if result == nil {
		return
	}

	remaining := result.Limit - result.Current
	if remaining < 0 {
		remaining = 0
	}
	c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", result.Limit))
	c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
	c.Set("X-RateLimit-Reset", fmt.Sprintf("%d", result.ResetAt.Unix()))
}

func chatRequestHasImage(req contractsopenai.ChatCompletionRequest) bool {
	for _, msg := range req.Messages {
		if contractsopenai.ChatContentHasImage(msg.Content) {
			return true
		}
	}
	return false
}

func supportsChatImageProvider(provider string) bool {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "", "openai", "azure", "groq", "mistral", "mistral-ai", "openrouter", "perplexity", "perplexityai", "sambanova", "togetherai", "together-ai", "xai", "anthropic", "gemini":
		return true
	default:
		return false
	}
}

func validateProviderImageInput(req contractsopenai.ChatCompletionRequest, provider string) error {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "anthropic", "gemini":
	default:
		return nil
	}
	for msgIndex, msg := range req.Messages {
		parts, ok, err := contractsopenai.ChatContentParts(msg.Content)
		if err != nil {
			return fmt.Errorf("message[%d].content: %v", msgIndex, err)
		}
		if !ok {
			continue
		}
		for partIndex, part := range parts {
			if part.Type != "image_url" {
				continue
			}
			if _, _, err := contractsopenai.ParseImageDataURL(part.ImageURL.URL); err != nil {
				return fmt.Errorf("message[%d].content part[%d].image_url.url: %v", msgIndex, partIndex, err)
			}
		}
	}
	return nil
}
