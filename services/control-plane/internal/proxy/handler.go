package proxy

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
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
	llmRouter  *router.Router
	tunnel     platform.TunnelTransport
	billing    budgeter
	wallets    walletStore
	logs       usageLogger
	logger     *zap.Logger
	vmResolver virtualmodels.VirtualModelResolver
	vmEngine   virtualmodels.RoutingEngine
	vmExec     virtualmodels.FallbackExecutor
}

// NewChatCompletionHandler creates a new handler.
func NewChatCompletionHandler(llmRouter *router.Router, tunnel platform.TunnelTransport, billing budgeter, wallets walletStore, logs usageLogger, logger *zap.Logger) *ChatCompletionHandler {
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

	var req contractsopenai.ChatCompletionRequest
	if err := c.BodyParser(&req); err != nil {
		return writeAPIError(c, fiber.StatusBadRequest, "Invalid request body", "invalid_request_error")
	}
	if req.Model == "" {
		return writeAPIError(c, fiber.StatusBadRequest, "Model is required", "invalid_request_error")
	}
	tracing.SetModelForTrace(c, req.Model)
	if len(req.Messages) == 0 {
		return writeAPIError(c, fiber.StatusBadRequest, "messages array must not be empty", "invalid_request_error")
	}
	for i, msg := range req.Messages {
		if msg.Role == "" {
			return writeAPIError(c, fiber.StatusBadRequest, fmt.Sprintf("message[%d].role is required", i), "invalid_request_error")
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
	if !isModelAllowed(authInfo, req.Model) {
		emitDenyCounter(h.logger, denyReasonForModel(authInfo), req.Model, currentTenant.ID)
		return writeAPIError(c, fiber.StatusForbidden, "Model access denied", "authorization_error")
	}
	authStart := time.Now()
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "auth.validate", "internal", authStart, fiber.StatusOK, map[string]interface{}{"model": req.Model})

	quotaStart := time.Now()
	estimated, err := h.billing.EstimateCost(c.Context(), req)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, "Failed to estimate request cost", "billing_error")
	}

	rateLimit, err := h.billing.CheckRateLimit(c.Context(), authInfo, req.Model, estimated.TotalTokens)
	setRateLimitHeaders(c, rateLimit)
	if err != nil {
		status := fiber.StatusTooManyRequests
		if err == billing.ErrBudgetExceeded {
			status = fiber.StatusPaymentRequired
		}
		return writeAPIError(c, status, err.Error(), "rate_limit_error")
	}
	if err := h.billing.CheckBudget(c.Context(), authInfo, estimated); err != nil {
		return writeAPIError(c, fiber.StatusPaymentRequired, err.Error(), "budget_exceeded")
	}
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "quota.check", "internal", quotaStart, fiber.StatusOK, map[string]interface{}{"estimated_tokens": estimated.TotalTokens})

	if err := h.tunnel.Healthy(c.Context()); err != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Tunnel path unavailable", "tunnel_unavailable")
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadRequest, "Failed to encode request", "invalid_request_error")
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
		Body: payload,
		RequestMetadata: map[string]interface{}{
			"tenant_slug": currentTenant.Slug,
			"model":       req.Model,
		},
	}

	routeStart := time.Now()
	deployment, err := h.llmRouter.Route(c.Context(), currentTenant.ID, req.Model)
	isVirtual := false
	if err != nil && h.vmResolver != nil && h.vmEngine != nil && h.vmExec != nil {
		vm, vmErr := h.vmResolver.Resolve(c.Context(), currentTenant.ID, req.Model)
		if vmErr == nil {
			targets, selErr := h.vmEngine.SelectTargets(vm)
			if selErr == nil && !req.Stream {
				resp, selectedDep, fallbackCount, failureReason, execErr := h.vmExec.ExecuteSync(c.Context(), targets, func(callCtx context.Context, dep *router.Deployment) (*platform.TunnelResponse, error) {
					tunnelReq := baseReq
					tunnelReq.Deployment = toContractDeployment(dep)
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
		emitDenyCounter(h.logger, denyReasonUnknownModel, req.Model, currentTenant.ID)
		return writeAPIError(c, fiber.StatusServiceUnavailable, fmt.Sprintf("No available deployment for model: %s", req.Model), "routing_error")
	}
	if !isVirtual {
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "router.select_model", "internal", routeStart, fiber.StatusOK, map[string]interface{}{"provider": deployment.Provider, "deployment_id": deployment.ID})
	}
	tunnelReq := baseReq
	tunnelReq.Deployment = toContractDeployment(deployment)

	if req.Stream {
		return h.handleStream(c, req, authInfo, estimated, tunnelReq, startTime)
	}
	return h.handleSync(c, req, authInfo, estimated, tunnelReq, startTime)
}

func (h *ChatCompletionHandler) handleSync(c *fiber.Ctx, req contractsopenai.ChatCompletionRequest, authInfo *platform.AuthContext, estimated platform.UsageCostBreakdown, tunnelReq platform.TunnelRequest, start time.Time) error {
	providerStart := time.Now()
	var resp *platform.TunnelResponse
	if vmResp, ok := c.Locals("vm_resp").(*platform.TunnelResponse); ok && vmResp != nil {
		resp = vmResp
	} else {
		var err error
		resp, err = h.tunnel.Proxy(c.Context(), tunnelReq)
		if err != nil {
			h.llmRouter.MarkFailed(c.Context(), &router.Deployment{ID: tunnelReq.Deployment.ID}, err.Error())
			return writeAPIError(c, fiber.StatusBadGateway, fmt.Sprintf("Tunnel error: %v", err), "api_error")
		}
	}
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, resp.StatusCode, map[string]interface{}{"provider": tunnelReq.Deployment.Provider})

	if resp.StatusCode >= httpErrorFloor {
		return transformProviderError(c, resp.StatusCode, resp.Body, resp.Headers)
	}

	var completion contractsopenai.ChatCompletionResponse
	if err := json.Unmarshal(resp.Body, &completion); err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, "Failed to decode provider response", "api_error")
	}

	finalCost, err := h.billing.FinalizeFromUsage(c.Context(), req.Model, completion.Usage, estimated)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, "Failed to finalize usage", "billing_error")
	}
	if _, err := h.wallets.DebitWallet(c.Context(), authInfo.TenantID, finalCost, tunnelReq.RequestID); err != nil {
		return writeAPIError(c, fiber.StatusPaymentRequired, err.Error(), "budget_exceeded")
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
		RequestMetadata: tunnelReq.RequestMetadata,
		CompletionStart: start,
		CompletionEnd:   time.Now(),
	})

	// Forward allowed provider response headers
	forwardProviderHeaders(c, resp.Headers)

	c.Set("x-shahrah-trace-id", tunnelReq.TraceID)
	c.Set("x-shahrah-cost-usd", fmt.Sprintf("%.8f", finalCost.TotalCostUSD))
	c.Set("x-shahrah-cost-irr", fmt.Sprintf("%.2f", finalCost.TotalCostIRR))
	responseStart := time.Now()
	err = c.Status(resp.StatusCode).JSON(completion)
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "response.write", "internal", responseStart, resp.StatusCode, nil)
	return err
}

func (h *ChatCompletionHandler) handleStream(c *fiber.Ctx, req contractsopenai.ChatCompletionRequest, authInfo *platform.AuthContext, estimated platform.UsageCostBreakdown, tunnelReq platform.TunnelRequest, start time.Time) error {
	stream, err := h.tunnel.Stream(c.Context(), tunnelReq)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, fmt.Sprintf("Tunnel error: %v", err), "api_error")
	}

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("x-shahrah-trace-id", tunnelReq.TraceID)

	// Forward provider headers from stream
	forwardProviderHeaders(c, stream.Headers())

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		defer stream.Close()

		for {
			select {
			case chunk, ok := <-stream.Chunks():
				if !ok {
					finalCost, _ := h.billing.FinalizeFromUsage(c.Context(), req.Model, nil, estimated)
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
						RequestMetadata: tunnelReq.RequestMetadata,
						CompletionStart: start,
						CompletionEnd:   time.Now(),
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
			case <-c.Context().Done():
				return
			}
		}
	})

	return nil
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
