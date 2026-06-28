// Package proxy hosts OpenAI-compatible request handlers for the control plane.
package proxy

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/apiheaders"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	metricsobs "github.com/Delpak/shahrah/services/control-plane/internal/observability/metrics"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// ModerationsHandler serves moderation requests.
type ModerationsHandler struct {
	llmRouter  *router.Router
	tunnel     platform.TunnelTransport
	billing    budgeter
	wallets    walletStore
	logs       usageLogger
	logger     *zap.Logger
	emitMetric func(input metricsobs.BuildInput)
}

type moderationValidationError struct {
	Message string
	Param   string
	Code    string
}

func (e *moderationValidationError) Error() string {
	return e.Message
}

// NewModerationsHandler creates a new moderations handler.
func NewModerationsHandler(llmRouter *router.Router, tunnel platform.TunnelTransport, billing budgeter, wallets walletStore, logs usageLogger, logger *zap.Logger, emitMetric ...func(input metricsobs.BuildInput)) *ModerationsHandler {
	var metricEmitter func(input metricsobs.BuildInput)
	if len(emitMetric) > 0 {
		metricEmitter = emitMetric[0]
	}

	return NewModerationsHandlerWithDependencies(BaseHandlerDependencies{
		Router:        llmRouter,
		Tunnel:        tunnel,
		Billing:       billing,
		Wallets:       wallets,
		RequestLogger: logs,
		SpendLogger:   logs,
		Logger:        logger,
		EmitMetric:    metricEmitter,
	})
}

// NewModerationsHandlerWithDependencies creates a moderations handler from explicit dependencies.
func NewModerationsHandlerWithDependencies(deps BaseHandlerDependencies) *ModerationsHandler {
	return &ModerationsHandler{
		llmRouter:  deps.Router,
		tunnel:     deps.Tunnel,
		billing:    deps.Billing,
		wallets:    deps.Wallets,
		logs:       deps.logger(),
		logger:     deps.Logger,
		emitMetric: deps.metricEmitter(),
	}
}

// Handle processes a moderations request.
func (h *ModerationsHandler) Handle(c *fiber.Ctx) error {
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

	req, err := parseModerationRequest(c.Body())
	if err != nil {
		if vErr, ok := err.(*moderationValidationError); ok {
			return writeAPIErrorDetail(c, fiber.StatusBadRequest, vErr.Message, "invalid_request_error", vErr.Param, vErr.Code)
		}

		return writeAPIError(c, fiber.StatusBadRequest, err.Error(), "invalid_request_error")
	}

	requestedModel := normalizeModelIdentifier(req.Model)

	routeCandidates := modelRoutingCandidates(requestedModel)
	if len(routeCandidates) == 0 {
		return writeAPIErrorDetail(c, fiber.StatusBadRequest, "model is required", "invalid_request_error", "model", "invalid_model")
	}

	tracing.SetModelForTrace(c, requestedModel)

	modelAccess := modelAccessForType(authInfo, requestedModel, platform.ModelTypeModeration)
	if !modelAccess.Allowed {
		if modelAccess.UnsupportedVMMode {
			return writeUnsupportedModelModeError(c, requestedModel, platform.ModelTypeModeration, modelAccess.ExposedModes)
		}

		emitDenyCounter(h.logger, denyReasonForModel(authInfo), requestedModel, currentTenant.ID)

		return writeAPIError(c, fiber.StatusForbidden, "Model access denied or unsupported for moderations", "authorization_error")
	}

	authStart := time.Now()
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "auth.validate", "internal", authStart, fiber.StatusOK, map[string]interface{}{"model": requestedModel})

	estimateReq := contractsopenai.EmbeddingRequest{
		Model: providerModelID(requestedModel),
		Input: req.Input,
	}
	quotaStart := time.Now()

	estimated, err := h.billing.EstimateEmbeddingCost(c.Context(), estimateReq)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, "Failed to estimate request cost", "billing_error")
	}

	rateLimit, err := h.billing.CheckRateLimit(c.Context(), authInfo, requestedModel, estimated.TotalTokens)
	setRateLimitHeaders(c, rateLimit)

	if err != nil {
		return writeAPIError(c, rateLimitErrorStatus(err), err.Error(), "rate_limit_error")
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

	routeStart := time.Now()

	var dep *router.Deployment
	for _, candidate := range routeCandidates {
		dep, err = h.llmRouter.Route(c.Context(), currentTenant.ID, candidate)
		if err == nil {
			break
		}
	}

	if err != nil || dep == nil {
		emitDenyCounter(h.logger, denyReasonUnknownModel, requestedModel, currentTenant.ID)
		return writeAPIError(c, fiber.StatusServiceUnavailable, fmt.Sprintf("No available deployment for model: %s", requestedModel), "routing_error")
	}

	if !supportsModerationsProvider(dep.Provider) {
		return writeAPIError(c, fiber.StatusNotImplemented, fmt.Sprintf("Moderations API is not enabled for provider: %s", dep.Provider), "invalid_request_error")
	}

	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "router.select_model", "internal", routeStart, fiber.StatusOK, map[string]interface{}{"provider": dep.Provider, "deployment_id": dep.ID})

	outboundReq := req

	outboundModel := strings.TrimSpace(dep.Model)
	if outboundModel == "" || outboundModel == "*" {
		outboundModel = providerModelID(requestedModel)
	}

	outboundReq.Model = providerModelID(outboundModel)

	payload, err := json.Marshal(outboundReq)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadRequest, "Failed to encode request", "invalid_request_error")
	}

	requestID := tracing.TraceIDFromContext(c)
	if requestID == "" {
		requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
	}

	tunnelReq := platform.TunnelRequest{
		RequestID:  requestID,
		TenantID:   currentTenant.ID,
		TraceID:    requestID,
		Stream:     false,
		Deployment: toContractDeployment(dep),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: payload,
		RequestMetadata: map[string]interface{}{
			"tenant_slug":            apiheaders.GetTenant(c),
			"model":                  requestedModel,
			"provider_endpoint_path": "/moderations",
			"input_count":            moderationInputCount(req.Input),
		},
	}

	return h.handleSync(c, req, authInfo, estimated, tunnelReq, startTime, tenantSlug)
}

func (h *ModerationsHandler) handleSync(c *fiber.Ctx, req contractsopenai.ModerationRequest, authInfo *platform.AuthContext, estimated platform.UsageCostBreakdown, tunnelReq platform.TunnelRequest, start time.Time, tenantSlug string) error {
	providerStart := time.Now()
	providerAttrs := map[string]interface{}{
		"provider":        tunnelReq.Deployment.Provider,
		"request_payload": buildTraceModerationRequestPayload(req),
	}

	resp, err := h.tunnel.Proxy(c.Context(), tunnelReq)
	if err != nil {
		providerAttrs["error"] = err.Error()
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, fiber.StatusBadGateway, providerAttrs)

		return writeAPIError(c, fiber.StatusBadGateway, fmt.Sprintf("Tunnel error: %v", err), "api_error")
	}

	if resp.StatusCode >= httpErrorFloor {
		providerAttrs["response_payload"] = buildTraceProviderErrorPayload(resp.StatusCode, resp.Body)
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, resp.StatusCode, providerAttrs)

		return transformProviderError(c, resp.StatusCode, resp.Body, resp.Headers)
	}

	var response contractsopenai.ModerationResponse

	unmarshalErr := json.Unmarshal(resp.Body, &response)
	if unmarshalErr != nil {
		providerAttrs["error"] = "failed to decode provider response"
		providerAttrs["response_payload"] = buildTraceProviderErrorPayload(fiber.StatusBadGateway, resp.Body)
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, fiber.StatusBadGateway, providerAttrs)

		return writeAPIError(c, fiber.StatusBadGateway, "Failed to decode provider response", "api_error")
	}

	providerAttrs["response_payload"] = buildTraceModerationResponsePayload(response)
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, resp.StatusCode, providerAttrs)

	finalCost, err := h.billing.FinalizeFromUsage(c.Context(), billingModelForDeployment(req.Model, tunnelReq.Deployment), nil, estimated)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, "Failed to finalize usage", "billing_error")
	}

	_, debitErr := h.wallets.DebitWallet(c.Context(), authInfo.TenantID, finalCost, tunnelReq.RequestID)
	if debitErr != nil {
		return writeAPIError(c, fiber.StatusPaymentRequired, debitErr.Error(), "budget_exceeded")
	}

	responseTimeMS := int(time.Since(start).Milliseconds())
	logRequestBestEffort(h.logger, h.logs, c.Context(), platform.RequestLogEntry{
		RequestID:      tunnelReq.RequestID,
		TokenHash:      authInfo.TokenHash,
		TenantID:       authInfo.TenantID,
		Model:          req.Model,
		Status:         "success",
		CostUSD:        finalCost.TotalCostUSD,
		ResponseTimeMS: responseTimeMS,
		RequestBody:    map[string]interface{}{"model": req.Model, "input_count": moderationInputCount(req.Input)},
		ResponseBody:   map[string]interface{}{"id": response.ID, "result_count": len(response.Results), "flagged_count": moderationFlaggedCount(response)},
	})
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
		ProviderName:       tunnelReq.Deployment.Provider,
		ProviderModelName:  tunnelReq.Deployment.Model,
		RequestType:        "moderations",
		UserID:             authInfo.UserID,
		VirtualAccountID:   authInfo.KeyID,
		VirtualAccountName: authInfo.KeyAlias,
		TeamIDs:            nonEmptySlice([]string{authInfo.TeamID}),
		StatusCode:         resp.StatusCode,
		LatencyMS:          responseTimeMS,
		InputTokens:        finalCost.InputTokens,
		OutputTokens:       finalCost.OutputTokens,
		CostUSD:            finalCost.TotalCostUSD,
		Metadata: map[string]interface{}{
			"endpoint":    "/v1/moderations",
			"key_type":    authInfo.KeyType,
			"team_id":     authInfo.TeamID,
			"org_unit_id": authInfo.OrgUnitID,
			"tenant_slug": tenantSlug,
			"input_count": moderationInputCount(req.Input),
		},
	})

	forwardProviderHeaders(c, resp.Headers)
	setTraceHeaders(c, tunnelReq.TraceID)
	setCostHeaders(c, finalCost)

	responseStart := time.Now()
	err = c.Status(resp.StatusCode).JSON(response)
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "response.write", "internal", responseStart, resp.StatusCode, map[string]interface{}{
		"response_payload": buildTraceModerationResponsePayload(response),
	})

	return err
}

func (h *ModerationsHandler) emitModelMetric(input metricsobs.BuildInput) {
	if h.emitMetric == nil {
		return
	}

	h.emitMetric(input)
}

func parseModerationRequest(body []byte) (contractsopenai.ModerationRequest, error) {
	var (
		req contractsopenai.ModerationRequest
		raw map[string]json.RawMessage
	)
	if err := json.Unmarshal(body, &raw); err != nil {
		return req, fmt.Errorf("invalid request body")
	}

	if err := json.Unmarshal(body, &req); err != nil {
		return req, fmt.Errorf("invalid request body")
	}

	req.Model = strings.TrimSpace(req.Model)
	req.User = strings.TrimSpace(req.User)

	if _, ok := raw["stream"]; ok {
		return req, moderationFieldError("stream is not supported for moderations", "stream", "unsupported_parameter")
	}

	for _, field := range []string{"messages", "prompt", "tools", "tool_choice", "temperature", "top_p", "max_tokens", "max_output_tokens", "encoding_format", "dimensions"} {
		if _, ok := raw[field]; ok {
			return req, moderationFieldError(fmt.Sprintf("%s is not supported for moderations", field), field, "unsupported_parameter")
		}
	}

	if req.Model == "" {
		return req, moderationFieldError("model is required", "model", "invalid_model")
	}

	normalizedInput, err := normalizeModerationInput(req.Input)
	if err != nil {
		return req, err
	}

	req.Input = normalizedInput

	return req, nil
}

func normalizeModerationInput(input interface{}) (interface{}, error) {
	switch v := input.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return nil, moderationFieldError("input is required and must be a non-empty string or a non-empty array of strings", "input", "invalid_input")
		}

		return v, nil
	case []interface{}:
		if len(v) == 0 {
			return nil, moderationFieldError("input is required and must be a non-empty string or a non-empty array of strings", "input", "invalid_input")
		}

		out := make([]string, 0, len(v))
		for _, item := range v {
			s, ok := item.(string)
			if !ok {
				return nil, moderationFieldError("input must be a string or array of strings", "input", "unsupported_input_type")
			}

			if strings.TrimSpace(s) == "" {
				return nil, moderationFieldError("input array items must be non-empty strings", "input", "invalid_input")
			}

			out = append(out, s)
		}

		return out, nil
	default:
		if input == nil {
			return nil, moderationFieldError("input is required and must be a non-empty string or a non-empty array of strings", "input", "invalid_input")
		}

		return nil, moderationFieldError("input must be a string or array of strings", "input", "unsupported_input_type")
	}
}

func moderationFieldError(message string, param string, code string) error {
	return &moderationValidationError{Message: message, Param: param, Code: code}
}

func moderationInputCount(input interface{}) int {
	switch v := input.(type) {
	case string:
		return 1
	case []string:
		return len(v)
	case []interface{}:
		return len(v)
	default:
		return 0
	}
}

func moderationFlaggedCount(resp contractsopenai.ModerationResponse) int {
	count := 0

	for _, result := range resp.Results {
		if result.Flagged {
			count++
		}
	}

	return count
}

func supportsModerationsProvider(provider string) bool {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "openai", "openrouter":
		return true
	default:
		return false
	}
}
