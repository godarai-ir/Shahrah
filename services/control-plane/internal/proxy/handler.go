// Package proxy hosts OpenAI-compatible request handlers for the control plane.
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
	"github.com/Delpak/shahrah/services/control-plane/internal/apiheaders"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/guardrails"
	metricsobs "github.com/Delpak/shahrah/services/control-plane/internal/observability/metrics"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/Delpak/shahrah/services/control-plane/internal/prompts"
	"github.com/Delpak/shahrah/services/control-plane/internal/responsecache"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
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

type guardrailResolver interface {
	ListEnabledTenantGuardrailsBySlugs(ctx context.Context, tenantID string, slugs []string) ([]platform.TenantGuardrail, error)
}

type promptResolver interface {
	GetPromptVersionByFQN(ctx context.Context, tenantID string, fqn string) (*platform.PromptVersion, error)
}

// ChatCompletionHandler handles chat completion requests.
// ChatCompletionHandler serves chat completion requests.
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
	guardrails    *guardrails.Engine
	guardrailRepo guardrailResolver
	prompts       promptResolver
}

// SetResponseCache enables opt-in response caching without changing existing constructor callers.
func (h *ChatCompletionHandler) SetResponseCache(service *responsecache.Service) {
	h.responseCache = service
}

// SetGuardrails enables per-request LLM guardrail evaluation.
func (h *ChatCompletionHandler) SetGuardrails(engine *guardrails.Engine) {
	h.guardrails = engine
}

// NewChatCompletionHandlerWithDependencies creates a chat handler from explicit dependencies.
func NewChatCompletionHandlerWithDependencies(deps ChatCompletionDependencies) *ChatCompletionHandler {
	resolver, engine, exec := newChatVirtualModelDeps(deps.VirtualModelStore)

	return &ChatCompletionHandler{
		llmRouter:     deps.Base.Router,
		tunnel:        deps.Base.Tunnel,
		billing:       deps.Base.Billing,
		wallets:       deps.Base.Wallets,
		logs:          deps.Base.logger(),
		logger:        deps.Base.Logger,
		vmResolver:    resolver,
		vmEngine:      engine,
		vmExec:        exec,
		emitMetric:    deps.Base.metricEmitter(),
		guardrailRepo: deps.GuardrailResolver,
		prompts:       deps.PromptResolver,
	}
}

// NewChatCompletionHandler creates a new chat completion handler.
func NewChatCompletionHandler(llmRouter *router.Router, tunnel platform.TunnelTransport, billing budgeter, wallets walletStore, logs usageLogger, logger *zap.Logger, emitMetric ...func(input metricsobs.BuildInput)) *ChatCompletionHandler {
	var metricEmitter func(input metricsobs.BuildInput)
	if len(emitMetric) > 0 {
		metricEmitter = emitMetric[0]
	}

	return NewChatCompletionHandlerWithDependencies(ChatCompletionDependencies{
		Base: BaseHandlerDependencies{
			Router:        llmRouter,
			Tunnel:        tunnel,
			Billing:       billing,
			Wallets:       wallets,
			RequestLogger: logs,
			SpendLogger:   logs,
			Logger:        logger,
			EmitMetric:    metricEmitter,
		},
	})
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

	if err := h.applyPromptVersion(c, currentTenant.ID, &req); err != nil {
		return err
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

	if handled, err := validateChatCompletionRequest(c, req); handled || err != nil {
		return err
	}

	modelAccess := modelAccessForType(authInfo, requestedModel, platform.ModelTypeChat)
	if !modelAccess.Allowed {
		if modelAccess.UnsupportedVMMode {
			return writeUnsupportedModelModeError(c, requestedModel, platform.ModelTypeChat, modelAccess.ExposedModes)
		}

		emitDenyCounter(h.logger, denyReasonForModel(authInfo), requestedModel, currentTenant.ID)

		return writeAPIError(c, fiber.StatusForbidden, "Model access denied or unsupported for chat completions", "authorization_error")
	}

	guardrailSelection, err := h.parseGuardrails(c)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadRequest, err.Error(), "invalid_request_error")
	}

	if h.guardrails != nil && !guardrailSelection.Empty() {
		guardrailOverrides, resolveErr := h.resolveTenantGuardrails(c.Context(), currentTenant.ID, guardrailSelection)
		if resolveErr != nil {
			return writeAPIError(c, fiber.StatusBadGateway, "Failed to resolve guardrails", "guardrail_resolution_error")
		}

		c.Locals("guardrail_overrides", guardrailOverrides)
		inputResult, inputErr := h.guardrails.ApplyInputWithGuardrails(req, guardrailSelection, guardrailOverrides)
		h.recordGuardrailChecks(c, inputResult.Checks)

		if inputErr != nil {
			return writeGuardrailError(c, fiber.StatusBadRequest, "Request blocked by GodarAI guardrail", inputResult.Checks)
		}

		req = inputResult.Request
	}

	if req.Stream && len(guardrailSelection.LLMOutput) > 0 {
		return writeAPIError(c, fiber.StatusBadRequest, "Streaming with output guardrails is not supported in this MVP", "invalid_request_error")
	}

	var cacheConfig *responsecache.RequestConfig

	if rawCacheConfig := apiheaders.GetCacheConfig(c); rawCacheConfig != "" && h.responseCache != nil {
		parsed, parseErr := h.responseCache.ParseHeader(rawCacheConfig)
		if parseErr != nil {
			return writeAPIError(c, fiber.StatusBadRequest, parseErr.Error(), "invalid_request_error")
		}

		cacheConfig = &parsed

		if !h.responseCache.Enabled() || req.Stream {
			c.Set(apiheaders.HeaderCacheStatus, "bypass")

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

	cachePending, cacheRelease, cacheHandled, cacheErr := h.handleResponseCacheLookup(c, req, authInfo, currentTenant.ID, requestedModel, cacheConfig, startTime, tenantSlug)
	if cacheRelease != nil {
		defer cacheRelease()
	}

	if cacheErr != nil || cacheHandled {
		return cacheErr
	}

	budgetErr := h.billing.CheckBudget(c.Context(), authInfo, estimated)
	if budgetErr != nil {
		return writeAPIError(c, fiber.StatusPaymentRequired, budgetErr.Error(), "budget_exceeded")
	}

	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "quota.check", "internal", quotaStart, fiber.StatusOK, map[string]interface{}{"estimated_tokens": estimated.TotalTokens})

	tunnelErr := h.tunnel.Healthy(c.Context())
	if tunnelErr != nil {
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
			"tenant_slug": apiheaders.GetTenant(c),
			"model":       requestedModel,
		}),
	}
	baseReq.RequestMetadata = enrichPromptMetadata(c, baseReq.RequestMetadata)

	routeStart := time.Now()

	deployment, isVirtual, routeErr := h.tryRouteVirtualModel(c, currentTenant.ID, authInfo, requestedModel, req, resolvedModel, baseReq, routeStart)
	if routeErr != nil {
		return routeErr
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
		return writeAPIError(c, fiber.StatusBadRequest, fmt.Sprintf("Provider %s does not support image input through GodarAI chat completions", deployment.Provider), "invalid_request_error")
	}

	validateErr := validateProviderImageInput(req, deployment.Provider)
	if validateErr != nil {
		return writeAPIError(c, fiber.StatusBadRequest, validateErr.Error(), "invalid_request_error")
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

	return h.handleSync(c, req, authInfo, estimated, tunnelReq, startTime, tenantSlug, cachePending, guardrailSelection)
}

func buildChatCompletionPayload(req contractsopenai.ChatCompletionRequest, deploymentModel string, fallbackModel string) ([]byte, error) {
	outboundReq := req

	outboundModel := strings.TrimSpace(deploymentModel)
	if outboundModel == "" || outboundModel == "*" {
		outboundModel = fallbackModel
	}

	outboundReq.Model = outboundModel
	outboundReq.PromptVersionFQN = ""
	outboundReq.PromptVariables = nil

	return json.Marshal(outboundReq)
}

func validateChatCompletionRequest(c *fiber.Ctx, req contractsopenai.ChatCompletionRequest) (bool, error) {
	if len(req.Messages) == 0 {
		return true, writeAPIError(c, fiber.StatusBadRequest, "messages array must not be empty", "invalid_request_error")
	}

	for i, msg := range req.Messages {
		if msg.Role == "" {
			return true, writeAPIError(c, fiber.StatusBadRequest, fmt.Sprintf("message[%d].role is required", i), "invalid_request_error")
		}

		if err := contractsopenai.ValidateChatMessageContent(msg.Content); err != nil {
			return true, writeAPIError(c, fiber.StatusBadRequest, fmt.Sprintf("message[%d].content: %v", i, err), "invalid_request_error")
		}
	}

	if req.Temperature != nil && (*req.Temperature < 0 || *req.Temperature > 2) {
		return true, writeAPIError(c, fiber.StatusBadRequest, "temperature must be between 0 and 2", "invalid_request_error")
	}

	if req.TopP != nil && (*req.TopP < 0 || *req.TopP > 1) {
		return true, writeAPIError(c, fiber.StatusBadRequest, "top_p must be between 0 and 1", "invalid_request_error")
	}

	if req.MaxTokens != nil && *req.MaxTokens < 1 {
		return true, writeAPIError(c, fiber.StatusBadRequest, "max_tokens must be at least 1", "invalid_request_error")
	}

	return false, nil
}

func (h *ChatCompletionHandler) applyPromptVersion(c *fiber.Ctx, tenantID string, req *contractsopenai.ChatCompletionRequest) error {
	if strings.TrimSpace(req.PromptVersionFQN) == "" {
		return nil
	}

	if h.prompts == nil {
		return writeAPIError(c, fiber.StatusNotImplemented, "Prompt store is not configured", "invalid_request_error")
	}

	promptVersion, err := h.prompts.GetPromptVersionByFQN(c.Context(), tenantID, req.PromptVersionFQN)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadRequest, err.Error(), "invalid_request_error")
	}

	rendered, err := prompts.Render(*promptVersion, req.PromptVariables, req.Messages, req.Model)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadRequest, err.Error(), "invalid_request_error")
	}

	req.Model = rendered.Model
	req.Messages = rendered.Messages
	c.Locals("prompt_metadata", map[string]interface{}{
		"prompt_id":         rendered.PromptID,
		"prompt_version_id": rendered.PromptVersionID,
		"prompt_fqn":        rendered.PromptFQN,
		"prompt_version":    promptVersion.Version,
	})

	return nil
}

func (h *ChatCompletionHandler) handleResponseCacheLookup(
	c *fiber.Ctx,
	req contractsopenai.ChatCompletionRequest,
	authInfo *platform.AuthContext,
	tenantID string,
	requestedModel string,
	cacheConfig *responsecache.RequestConfig,
	startTime time.Time,
	tenantSlug string,
) (*responsecache.Pending, func(), bool, error) {
	if cacheConfig == nil {
		return nil, nil, false, nil
	}

	keyID := strings.TrimSpace(authInfo.KeyID)
	if keyID == "" {
		keyID = strings.TrimSpace(authInfo.TokenHash)
	}

	cacheScope := responsecache.Scope{
		TenantID:     tenantID,
		KeyID:        keyID,
		Model:        requestedModel,
		ModelVersion: h.virtualModelCacheVersion(c.Context(), authInfo, tenantID, requestedModel),
	}
	cacheResult, cachePending, lookupErr := h.responseCache.Lookup(c.Context(), cacheScope, req, *cacheConfig)

	c.Set(apiheaders.HeaderCacheStatus, cacheResult.Status)
	c.Locals("response_cache_status", cacheResult.Status)
	c.Locals("response_cache_type", cacheConfig.Type)
	c.Locals("response_cache_embedding_latency_ms", cacheResult.EmbeddingLatencyMS)

	if lookupErr != nil {
		h.logger.Warn("Response cache lookup failed", zap.String("cache_type", cacheConfig.Type), zap.Error(lookupErr))
	} else if cacheResult.Entry != nil {
		return cachePending, nil, true, h.handleCacheHit(c, req, authInfo, cacheResult, startTime, tenantSlug)
	}

	if cacheResult.Status == "bypass" {
		return cachePending, nil, false, nil
	}

	release, waited, acquireErr := h.responseCache.AcquireMiss(c.Context(), cachePending)
	if acquireErr != nil {
		return cachePending, nil, false, writeAPIError(c, fiber.StatusRequestTimeout, "Request canceled while waiting for response cache", "request_canceled")
	}

	if !waited {
		return cachePending, release, false, nil
	}

	secondResult, secondPending, secondErr := h.responseCache.Lookup(c.Context(), cacheScope, req, *cacheConfig)
	cachePending = secondPending

	if secondErr == nil && secondResult.Entry != nil {
		c.Set(apiheaders.HeaderCacheStatus, "hit")
		return cachePending, release, true, h.handleCacheHit(c, req, authInfo, secondResult, startTime, tenantSlug)
	}

	return cachePending, release, false, nil
}

func (h *ChatCompletionHandler) tryRouteVirtualModel(
	c *fiber.Ctx,
	tenantID string,
	authInfo *platform.AuthContext,
	requestedModel string,
	req contractsopenai.ChatCompletionRequest,
	resolvedModel string,
	baseReq platform.TunnelRequest,
	routeStart time.Time,
) (*router.Deployment, bool, error) {
	if req.Stream || h.vmEngine == nil || h.vmExec == nil {
		return nil, false, nil
	}

	if snapshotVM, ok := buildVirtualModelFromSnapshot(authInfo, requestedModel); ok {
		deployment, err := h.executeVirtualModelRoute(c, tenantID, snapshotVM, req, resolvedModel, baseReq, routeStart, true)
		if err != nil || deployment != nil {
			return deployment, deployment != nil, err
		}
	}

	if h.vmResolver == nil {
		return nil, false, nil
	}

	vm, resolveErr := h.vmResolver.Resolve(c.Context(), tenantID, requestedModel)
	if resolveErr == nil {
		deployment, err := h.executeVirtualModelRoute(c, tenantID, vm, req, resolvedModel, baseReq, routeStart, false)

		return deployment, deployment != nil, err
	}

	return nil, false, nil
}

func (h *ChatCompletionHandler) executeVirtualModelRoute(
	c *fiber.Ctx,
	tenantID string,
	vm *platform.VirtualModel,
	req contractsopenai.ChatCompletionRequest,
	resolvedModel string,
	baseReq platform.TunnelRequest,
	routeStart time.Time,
	fromSnapshot bool,
) (*router.Deployment, error) {
	targets, selErr := h.vmEngine.SelectTargets(vm)
	if selErr != nil {
		targets = nil
	}

	if targets == nil {
		return nil, nil
	}

	resp, selectedDep, fallbackCount, failureReason, execErr := h.vmExec.ExecuteSync(c.Context(), targets, func(callCtx context.Context, dep *router.Deployment) (*platform.TunnelResponse, error) {
		dep = h.resolveVirtualDeployment(callCtx, tenantID, dep)
		tunnelReq := baseReq
		tunnelReq.Deployment = toContractDeployment(dep)

		payload, payloadErr := buildChatCompletionPayload(req, dep.Model, resolvedModel)
		if payloadErr != nil {
			return nil, payloadErr
		}

		tunnelReq.Body = payload

		return h.tunnel.Proxy(callCtx, tunnelReq)
	})
	if execErr == nil && resp != nil && selectedDep != nil {
		h.recordVirtualModelRoute(c, vm, selectedDep, resp, fallbackCount, failureReason, routeStart, fromSnapshot)
		return selectedDep, nil
	}

	if errors.Is(execErr, virtualmodels.ErrFallbackExhausted) {
		return nil, writeAPIError(c, fiber.StatusBadGateway, "Virtual model fallback exhausted", "fallback_exhausted")
	}

	return nil, nil
}

func (h *ChatCompletionHandler) recordVirtualModelRoute(
	c *fiber.Ctx,
	vm *platform.VirtualModel,
	selectedDep *router.Deployment,
	resp *platform.TunnelResponse,
	fallbackCount int,
	failureReason string,
	routeStart time.Time,
	fromSnapshot bool,
) {
	c.Locals("vm_resp", resp)
	c.Locals("vm_fallback_count", fallbackCount)
	c.Locals("vm_failure_reason", failureReason)

	metadata := map[string]interface{}{
		"virtual_model":     vm.Alias,
		"routing_strategy":  vm.RoutingStrategy,
		"selected_provider": selectedDep.Provider,
		"selected_model":    selectedDep.Model,
		"fallback_count":    fallbackCount,
		"failure_reason":    failureReason,
	}
	if fromSnapshot {
		metadata["snapshot"] = true
	}

	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "router.select_virtual_model", "internal", routeStart, fiber.StatusOK, metadata)
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

func (h *ChatCompletionHandler) handleSync(
	c *fiber.Ctx,
	req contractsopenai.ChatCompletionRequest,
	authInfo *platform.AuthContext,
	estimated platform.UsageCostBreakdown,
	tunnelReq platform.TunnelRequest,
	start time.Time,
	tenantSlug string,
	cachePending *responsecache.Pending,
	guardrailSelection guardrails.RequestSelection,
) error {
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
		logProviderFailureRequest(h.logger, h.logs, c, authInfo, tunnelReq, req.Model, "/v1/chat/completions", "chat_completions", resp.StatusCode, resp.Body, start)

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

	_, debitErr := h.wallets.DebitWallet(c.Context(), authInfo.TenantID, finalCost, tunnelReq.RequestID)
	if debitErr != nil {
		return writeAPIError(c, fiber.StatusPaymentRequired, debitErr.Error(), "budget_exceeded")
	}

	if h.guardrails != nil && len(guardrailSelection.LLMOutput) > 0 {
		guardrailOverrides := map[string]guardrails.Guardrail(nil)
		if overrides, ok := c.Locals("guardrail_overrides").(map[string]guardrails.Guardrail); ok {
			guardrailOverrides = overrides
		}

		outputResult, outputErr := h.guardrails.ApplyOutputWithGuardrails(completion, guardrailSelection, guardrailOverrides)
		h.recordGuardrailChecks(c, outputResult.Checks)

		if outputErr != nil {
			responseTimeMS := int(time.Since(start).Milliseconds())
			logEntry := buildRequestLogEntry(c, authInfo, tunnelReq.Deployment, req.Model, "/v1/chat/completions", "chat_completions", "blocked", fiber.StatusBadRequest, start, responseTimeMS)
			logEntry.RequestID = tunnelReq.RequestID
			logEntry.TraceID = tunnelReq.TraceID
			logEntry.CostUSD = finalCost.TotalCostUSD
			logEntry.CostIRR = finalCost.TotalCostIRR
			logEntry.InputTokens = finalCost.InputTokens
			logEntry.OutputTokens = finalCost.OutputTokens
			logEntry.TotalTokens = finalCost.TotalTokens
			logEntry.RequestBody = map[string]interface{}{"model": req.Model}
			logEntry.ResponseBody = map[string]interface{}{"guardrail_status": "blocked"}
			logEntry.ErrorType = "guardrail_blocked"
			logRequestBestEffort(h.logger, h.logs, c.Context(), logEntry)
			logSpendBestEffort(h.logger, h.logs, c.Context(), platform.SpendLog{
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
				Status:          "blocked",
				RequestMetadata: enrichPricingMetadata(tunnelReq.RequestMetadata, authInfo, req.Model),
				CompletionStart: start,
				CompletionEnd:   time.Now(),
			})

			return writeGuardrailError(c, fiber.StatusBadRequest, "Response blocked by GodarAI guardrail", outputResult.Checks)
		}

		if outputResult.Response != nil {
			completion = *outputResult.Response
		}
	}

	if cachePending != nil && h.responseCache != nil {
		cachedBody, marshalErr := json.Marshal(completion)
		if marshalErr == nil {
			if cacheErr := h.responseCache.Store(c.Context(), cachePending, cachedBody, tunnelReq.TraceID, tunnelReq.Deployment.Provider, tunnelReq.Deployment.Model, finalCost, "gateway-pricing-v1"); cacheErr != nil {
				h.logger.Warn("Response cache store failed", zap.Error(cacheErr))
			}
		}
	}

	responseTimeMS := int(time.Since(start).Milliseconds())
	logEntry := buildRequestLogEntry(c, authInfo, tunnelReq.Deployment, req.Model, "/v1/chat/completions", "chat_completions", "success", resp.StatusCode, start, responseTimeMS)
	logEntry.RequestID = tunnelReq.RequestID
	logEntry.TraceID = tunnelReq.TraceID
	logEntry.CostUSD = finalCost.TotalCostUSD
	logEntry.CostIRR = finalCost.TotalCostIRR
	logEntry.InputTokens = finalCost.InputTokens
	logEntry.OutputTokens = finalCost.OutputTokens
	logEntry.TotalTokens = finalCost.TotalTokens
	logEntry.RequestBody = map[string]interface{}{"model": req.Model}
	logEntry.ResponseBody = map[string]interface{}{"id": completion.ID}
	logRequestBestEffort(h.logger, h.logs, c.Context(), logEntry)
	logSpendBestEffort(h.logger, h.logs, c.Context(), platform.SpendLog{
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
		CacheMode:          cacheMetricString(c, "response_cache_type"),
		CacheStatus:        cacheMetricString(c, "response_cache_status"),
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

	setTraceHeaders(c, tunnelReq.TraceID)
	setCostHeaders(c, finalCost)

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
	logEntry := buildRequestLogEntry(c, authInfo, platform.Deployment{Provider: entry.Provider, Model: entry.ProviderModel}, req.Model, "/v1/chat/completions", "chat_completions", "success", fiber.StatusOK, start, responseTimeMS)
	logEntry.RequestID = requestID
	logEntry.TraceID = entry.TraceID
	logEntry.RequestBody = map[string]interface{}{"model": req.Model}
	logEntry.ResponseBody = map[string]interface{}{"cache_status": "hit"}
	logEntry.CacheStatus = "hit"
	logRequestBestEffort(h.logger, h.logs, c.Context(), logEntry)
	logSpendBestEffort(h.logger, h.logs, c.Context(), platform.SpendLog{
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
		CostUSD:   0,
		CacheMode: result.CacheType, CacheStatus: "hit", CacheMatch: cacheMatchForResult(result),
		CacheSimilarity: cacheSimilarityForResult(result), CacheOriginRequestID: entry.TraceID,
		AvoidedInputTokens: entry.CostSnapshot.InputTokens, AvoidedOutputTokens: entry.CostSnapshot.OutputTokens,
		AvoidedCostUSD: entry.CostSnapshot.CostUSD, AvoidedCostIRR: entry.CostSnapshot.CostIRR,
		PricingVersion: entry.CostSnapshot.PricingVersion,
		Metadata:       metadata,
	})
	c.Set("Content-Type", fiber.MIMEApplicationJSON)
	c.Set(apiheaders.HeaderCacheStatus, "hit")
	c.Set(apiheaders.HeaderCachedTraceID, entry.TraceID)

	if result.Semantic {
		c.Set(apiheaders.HeaderCacheSimilarity, fmt.Sprintf("%.6f", result.Similarity))
	}

	setTraceHeaders(c, requestID)
	setCostHeaders(c, platform.UsageCostBreakdown{})

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
	setTraceHeaders(c, tunnelReq.TraceID)

	// Forward provider headers from stream
	forwardProviderHeaders(c, stream.Headers())
	virtualModelName := modelMetricVirtualName(c, req.Model)

	requestCtx := c.Context()
	requestCtx.SetBodyStreamWriter(func(w *bufio.Writer) {
		defer closeStreamBestEffort(h.logger, stream)

		for {
			select {
			case chunk, ok := <-stream.Chunks():
				if !ok {
					billingModel := billingModelForDeployment(req.Model, tunnelReq.Deployment)

					finalCost, finalizeErr := h.billing.FinalizeFromUsage(requestCtx, billingModel, nil, estimated)
					if finalizeErr != nil {
						h.logger.Warn("failed to finalize streaming usage", zap.Error(finalizeErr), zap.String("request_id", tunnelReq.RequestID))

						finalCost = estimated
						finalCost.Model = billingModel
					}

					if _, debitErr := h.wallets.DebitWallet(requestCtx, authInfo.TenantID, finalCost, tunnelReq.RequestID); debitErr != nil {
						h.logger.Warn("failed to debit wallet for streaming request", zap.Error(debitErr), zap.String("request_id", tunnelReq.RequestID))
					}

					logSpendBestEffort(h.logger, h.logs, requestCtx, platform.SpendLog{
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
					if writeErr := writeBufferedBytes(w, chunk.Data); writeErr != nil {
						h.logger.Warn("failed to write streaming chunk", zap.Error(writeErr), zap.String("request_id", tunnelReq.RequestID))
						return
					}
				}

				if chunk.IsDone {
					return
				}
			case err, ok := <-stream.Errors():
				if ok && err != nil {
					errorPayload, marshalErr := json.Marshal(contractsopenai.APIError{
						Error: contractsopenai.ErrorInfo{
							Message: err.Error(),
							Type:    "streaming_error",
						},
					})
					if marshalErr != nil {
						h.logger.Warn("failed to encode streaming error payload", zap.Error(marshalErr), zap.String("request_id", tunnelReq.RequestID))
						return
					}

					if writeErr := writeBufferedBytes(w, bytes.Join([][]byte{[]byte("data: "), errorPayload, []byte("\n\n")}, nil)); writeErr != nil {
						h.logger.Warn("failed to write streaming error payload", zap.Error(writeErr), zap.String("request_id", tunnelReq.RequestID))
					}
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

func enrichPromptMetadata(c *fiber.Ctx, metadata map[string]interface{}) map[string]interface{} {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	if c == nil {
		return metadata
	}

	if raw, ok := c.Locals("prompt_metadata").(map[string]interface{}); ok {
		for key, value := range raw {
			metadata[key] = value
		}
	}

	return metadata
}

func cacheMetricString(c *fiber.Ctx, key string) string {
	if c == nil {
		return ""
	}

	value := ""
	if localValue, ok := c.Locals(key).(string); ok {
		value = localValue
	}

	return value
}

func cacheMatchForResult(result responsecache.Result) string {
	if result.Semantic {
		return "semantic"
	}

	return "exact"
}

func cacheSimilarityForResult(result responsecache.Result) *float64 {
	if !result.Semantic {
		return nil
	}

	score := result.Similarity

	return &score
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

func logRequestBestEffort(logger *zap.Logger, logs usageLogger, ctx context.Context, entry platform.RequestLogEntry) {
	if err := logs.LogRequest(ctx, entry); err != nil && logger != nil && !errors.Is(err, storage.ErrLogStoreUnavailable) {
		logger.Warn("failed to log request", zap.Error(err), zap.String("request_id", entry.RequestID))
	}
}

func logProviderFailureRequest(
	logger *zap.Logger,
	logs usageLogger,
	c *fiber.Ctx,
	authInfo *platform.AuthContext,
	tunnelReq platform.TunnelRequest,
	requestedModel string,
	route string,
	requestType string,
	statusCode int,
	responseBody []byte,
	start time.Time,
) {
	responseTimeMS := int(time.Since(start).Milliseconds())
	logEntry := buildRequestLogEntry(c, authInfo, tunnelReq.Deployment, requestedModel, route, requestType, "failed", statusCode, start, responseTimeMS)
	logEntry.RequestID = tunnelReq.RequestID
	logEntry.TraceID = tunnelReq.TraceID
	logEntry.RequestBody = map[string]interface{}{"model": requestedModel}
	logEntry.ResponseBody = map[string]interface{}{"error": parseGenericError(responseBody)}

	if errType, ok := ErrorTypeMap[statusCode]; ok && strings.TrimSpace(errType) != "" {
		logEntry.ErrorType = errType
	} else {
		logEntry.ErrorType = "api_error"
	}

	logRequestBestEffort(logger, logs, c.Context(), logEntry)
}

func logSpendBestEffort(logger *zap.Logger, logs usageLogger, ctx context.Context, spendLog platform.SpendLog) {
	if err := logs.LogSpend(ctx, spendLog); err != nil && logger != nil {
		logger.Warn("failed to log spend", zap.Error(err), zap.String("request_id", spendLog.RequestID))
	}
}

func closeStreamBestEffort(logger *zap.Logger, stream interface{ Close() error }) {
	if err := stream.Close(); err != nil && logger != nil {
		logger.Warn("failed to close stream", zap.Error(err))
	}
}

func writeBufferedBytes(w *bufio.Writer, data []byte) error {
	if _, err := w.Write(data); err != nil {
		return err
	}

	return w.Flush()
}

func writeBufferedString(w *bufio.Writer, data string) error {
	if _, err := w.WriteString(data); err != nil {
		return err
	}

	return w.Flush()
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
