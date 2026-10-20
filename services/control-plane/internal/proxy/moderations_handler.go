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
	pipeline   *RequestPipeline
	billing    budgeter
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
func NewModerationsHandler(llmRouter *router.Router, tunnel platform.TunnelTransport, billing budgeter, logs usageLogger, logger *zap.Logger, emitMetric ...func(input metricsobs.BuildInput)) *ModerationsHandler {
	var metricEmitter func(input metricsobs.BuildInput)
	if len(emitMetric) > 0 {
		metricEmitter = emitMetric[0]
	}

	return NewModerationsHandlerWithDependencies(BaseHandlerDependencies{
		Router:              llmRouter,
		Tunnel:              tunnel,
		Billing:             billing,
		BillingCapabilities: compatibilityBillingCapabilities(),
		RequestLogger:       logs,
		SpendLogger:         logs,
		Logger:              logger,
		EmitMetric:          metricEmitter,
	})
}

// NewModerationsHandlerWithDependencies creates a moderations handler from explicit dependencies.
func NewModerationsHandlerWithDependencies(deps BaseHandlerDependencies) *ModerationsHandler {
	return &ModerationsHandler{
		llmRouter:  deps.Router,
		pipeline:   deps.requestPipeline(),
		billing:    deps.Billing,
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

	originalBody := append([]byte(nil), c.Body()...)
	requestLog := newRequestLogLifecycle(h.logger, h.logs, c, authInfo, "/v1/moderations", "moderations", startTime)
	requestLog.SetRequestBody(buildJSONRequestLogPayload(originalBody, nil))

	defer requestLog.FinalizeFromFiber()

	req, err := parseModerationRequest(originalBody)
	if err != nil {
		if vErr, ok := err.(*moderationValidationError); ok {
			return writeAPIErrorDetail(c, fiber.StatusBadRequest, vErr.Message, "invalid_request_error", vErr.Param, vErr.Code)
		}

		return writeAPIError(c, fiber.StatusBadRequest, err.Error(), "invalid_request_error")
	}

	requestedModel := normalizeModelIdentifier(req.Model)
	requestLog.SetModel(requestedModel)
	requestLog.SetRequestBody(buildModerationRequestLogPayload(req, originalBody))

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

		denyReason := denyReasonForModel(authInfo)
		emitDenyCounter(h.logger, denyReason, requestedModel, currentTenant.ID)

		const message = "Model access denied or unsupported for moderations"
		logPreProviderRejection(h.logger, h.logs, c, authInfo, requestedModel, "/v1/moderations", "moderations", fiber.StatusForbidden, "authorization_error", message, denyReason, startTime)
		requestLog.MarkFinalized()

		return writeAPIError(c, fiber.StatusForbidden, message, "authorization_error")
	}

	authStart := time.Now()
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "auth.validate", "internal", authStart, fiber.StatusOK, map[string]interface{}{"model": requestedModel})

	estimateReq := contractsopenai.EmbeddingRequest{
		Model: billingEstimateModel(requestedModel),
		Input: req.Input,
	}
	quotaStart := time.Now()

	estimated, err := h.billing.EstimateEmbeddingCost(c.Context(), estimateReq)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, "Failed to estimate request cost", "billing_error")
	}

	rateLimit, err := h.billing.CheckRateLimit(c.Context(), authInfo, requestedModel)
	setRateLimitHeaders(c, rateLimit)

	if err != nil {
		if recordRateLimitRejection(c, h.logger, h.logs, h.emitMetric, authInfo, requestedModel, "/v1/moderations", "moderations", startTime, quotaStart, rateLimit, err) {
			requestLog.MarkFinalized()
		}

		return writeAPIError(c, rateLimitErrorStatus(err), err.Error(), "rate_limit_error")
	}

	requestID := tracing.TraceIDFromContext(c)
	if requestID == "" {
		requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
	}

	var (
		budgetReservation *platform.BudgetReservation
	)

	budgetReservationAttempted := false
	if !shouldDeferBudgetReservationUntilRoute(requestedModel) {
		budgetReservationAttempted = true

		budgetReservation, err = h.pipeline.Reserve(c.Context(), budgetRequestFromHTTP(c, authInfo, estimated, requestID))
		if err != nil {
			return writeBudgetCheckError(c, err)
		}
	}

	releaseBudgetOnExit := true
	defer func() {
		if releaseBudgetOnExit {
			h.pipeline.Release(c.Context(), budgetReservation, "request_not_committed")
		}
	}()

	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "quota.check", "internal", quotaStart, fiber.StatusOK, map[string]interface{}{"estimated_tokens": estimated.TotalTokens})

	tunnelErr := h.pipeline.EnsureReady(c.Context())
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
		if handled, dependencyErr := writeRoutingDependencyUnavailable(c, h.logger, err, requestedModel, currentTenant.ID); handled {
			return dependencyErr
		}

		emitDenyCounter(h.logger, denyReasonUnknownModel, requestedModel, currentTenant.ID)

		return writeAPIError(c, fiber.StatusServiceUnavailable, fmt.Sprintf("No available deployment for model: %s", requestedModel), "routing_error")
	}

	if !supportsModerationsProvider(dep.Provider) {
		return writeAPIError(c, fiber.StatusNotImplemented, fmt.Sprintf("Moderations API is not enabled for provider: %s", dep.Provider), "invalid_request_error")
	}

	estimated, err = applyDeploymentPricingOverride(c.Context(), h.pipeline, estimated, dep)
	if err != nil {
		return writeBillingEstimateError(c, err)
	}

	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "router.select_model", "internal", routeStart, fiber.StatusOK, map[string]interface{}{"provider": dep.Provider, "deployment_id": dep.ID})

	outboundModel := strings.TrimSpace(dep.Model)
	if outboundModel == "" || outboundModel == "*" {
		outboundModel = providerModelID(requestedModel)
	}

	payload, err := rewriteOpenAIJSONModel(
		originalBody,
		providerModelID(outboundModel),
	)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadRequest, "Failed to encode request", "invalid_request_error")
	}

	operation := platform.ProviderOperationOpenAIModerations
	providerPath := "/moderations"

	if isMistralProvider(dep.Provider) && strings.TrimSuffix(c.Path(), "/") == "/v1/chat/moderations" {
		operation = platform.ProviderOperationMistralChatModerations
		providerPath = "/chat/moderations"
	}

	if isMistralProvider(dep.Provider) {
		profileEndpoint := "moderations"
		if operation == platform.ProviderOperationMistralChatModerations {
			profileEndpoint = "chat_moderations"
		}

		if endpointErr := validateMistralJSONParameters(estimated.ModelRuntimeProfile, profileEndpoint, originalBody); endpointErr != nil {
			return writeAPIError(c, fiber.StatusBadRequest, endpointErr.Error(), "invalid_request_error")
		}
	}

	tunnelReq := platform.TunnelRequest{
		RequestID:  requestID,
		TenantID:   currentTenant.ID,
		TraceID:    requestID,
		Stream:     false,
		Operation:  operation,
		Deployment: toContractDeployment(dep),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: payload,
		RequestMetadata: map[string]interface{}{
			"tenant_slug":            apiheaders.GetTenant(c),
			"model":                  requestedModel,
			"provider_endpoint_path": providerPath,
			"input_count":            moderationInputCount(req.Input),
		},
	}

	tunnelReq.ModelRuntimeProfiles, err = runtimeProfilesFromEstimate(dep.Provider, outboundModel, estimated)
	if err != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, err.Error(), "runtime_profile_unavailable")
	}

	tunnelReq.RequestMetadata = enrichRequestMetadataForDeployment(tunnelReq.RequestMetadata, tunnelReq.Deployment)
	requestLog.SetDeployment(tunnelReq.Deployment)
	requestLog.SetRequestIDs(tunnelReq.RequestID, tunnelReq.TraceID)

	if !budgetReservationAttempted {
		budgetReservation, err = h.pipeline.Reserve(c.Context(), budgetRequestFromHTTP(c, authInfo, estimated, requestID, deploymentBillingMode(tunnelReq.Deployment)))
		if err != nil {
			return writeBudgetCheckError(c, err)
		}
	}

	err = h.handleSync(c, req, originalBody, authInfo, estimated, tunnelReq, budgetReservation, startTime, tenantSlug, requestLog)
	if err == nil {
		releaseBudgetOnExit = false
	}

	return err
}

func (h *ModerationsHandler) handleSync(
	c *fiber.Ctx,
	req contractsopenai.ModerationRequest,
	originalBody []byte,
	authInfo *platform.AuthContext,
	estimated platform.UsageCostBreakdown,
	tunnelReq platform.TunnelRequest,
	budgetReservation *platform.BudgetReservation,
	start time.Time,
	tenantSlug string,
	requestLog *requestLogLifecycle,
) error {
	providerStart := time.Now()
	providerAttrs := map[string]interface{}{
		"provider":        tunnelReq.Deployment.Provider,
		"request_payload": buildTraceModerationRequestPayload(req),
	}

	resp, err := h.pipeline.Dispatch(c.Context(), tunnelReq)
	if err != nil {
		providerAttrs["error"] = err.Error()
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, fiber.StatusBadGateway, providerAttrs)

		return writeAPIError(c, fiber.StatusBadGateway, fmt.Sprintf("Tunnel error: %v", err), "api_error")
	}

	if resp.StatusCode >= httpErrorFloor {
		providerAttrs["response_payload"] = buildTraceProviderErrorPayload(resp.StatusCode, resp.Body)
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, resp.StatusCode, providerAttrs)

		return writeCompatibleProviderError(c, tunnelReq.Deployment.Provider, resp.StatusCode, resp.Body, resp.Headers)
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

	if debitErr := h.pipeline.Settle(c.Context(), SettlementRequest{Reservation: budgetReservation, TenantID: authInfo.TenantID, Actual: finalCost, RequestID: tunnelReq.RequestID, BillingMode: deploymentBillingMode(tunnelReq.Deployment)}); debitErr != nil {
		return writeBudgetCheckError(c, debitErr)
	}

	responseTimeMS := int(time.Since(start).Milliseconds())
	logEntry := buildRequestLogEntry(c, authInfo, tunnelReq.Deployment, req.Model, "/v1/moderations", "moderations", "success", resp.StatusCode, start, responseTimeMS)
	logEntry.RequestID = tunnelReq.RequestID
	logEntry.TraceID = tunnelReq.TraceID
	logEntry.CostUSD = finalCost.TotalCostUSD
	logEntry.CostIRR = finalCost.TotalCostIRR
	logEntry.InputTokens = finalCost.InputTokens
	logEntry.OutputTokens = finalCost.OutputTokens
	logEntry.TotalTokens = finalCost.TotalTokens
	logEntry.RequestBody = buildModerationRequestLogPayload(req, originalBody)
	logEntry.ResponseBody = buildJSONResponseLogPayload(resp.Body, response)
	requestLog.FinalizeEntry(c.Context(), logEntry)
	logSpendBestEffort(h.logger, h.logs, c.Context(), platform.SpendLog{
		RequestID:       tunnelReq.RequestID,
		TokenHash:       authInfo.TokenHash,
		TenantID:        authInfo.TenantID,
		TeamID:          authInfo.TeamID,
		UserID:          authInfo.UserID,
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
		Metadata: enrichMetricMetadataForDeployment(map[string]interface{}{
			"endpoint":    "/v1/moderations",
			"key_type":    authInfo.KeyType,
			"team_id":     authInfo.TeamID,
			"tenant_slug": tenantSlug,
			"input_count": moderationInputCount(req.Input),
		}, tunnelReq.Deployment),
	})

	forwardProviderHeaders(c, resp.Headers)
	setTraceHeaders(c, tunnelReq.TraceID)
	setCostHeaders(c, finalCost)

	responseStart := time.Now()

	if contentType := strings.TrimSpace(resp.Headers["Content-Type"]); contentType != "" {
		c.Set(fiber.HeaderContentType, contentType)
	} else {
		c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	}

	err = c.Status(resp.StatusCode).Send(resp.Body)
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "response.write", "internal", responseStart, resp.StatusCode, map[string]interface{}{
		"response_payload": buildTraceModerationResponsePayload(response),
	})

	return err
}

func buildModerationRequestLogPayload(req contractsopenai.ModerationRequest, originalBody []byte) map[string]interface{} {
	return buildJSONRequestLogPayload(originalBody, map[string]interface{}{
		"model": req.Model,
		"input": req.Input,
	})
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

		for _, item := range v {
			switch typed := item.(type) {
			case string:
				if strings.TrimSpace(typed) == "" {
					return nil, moderationFieldError("input array items must be non-empty", "input", "invalid_input")
				}
			case map[string]interface{}:
				if err := validateModerationMultimodalItem(typed); err != nil {
					return nil, err
				}
			default:
				return nil, moderationFieldError("input must contain text or image input items", "input", "unsupported_input_type")
			}
		}

		return v, nil
	case map[string]interface{}:
		if err := validateModerationMultimodalItem(v); err != nil {
			return nil, err
		}

		return v, nil
	default:
		if input == nil {
			return nil, moderationFieldError("input is required and must be a non-empty string or a non-empty array of strings", "input", "invalid_input")
		}

		return nil, moderationFieldError("input must be text, an image input item, or an array of those items", "input", "unsupported_input_type")
	}
}

func validateModerationMultimodalItem(item map[string]interface{}) error {
	itemType, _ := item["type"].(string) //nolint:errcheck // Optional metadata uses the zero value when absent or mismatched.
	switch strings.TrimSpace(itemType) {
	case "text":
		text, _ := item["text"].(string) //nolint:errcheck // Optional metadata uses the zero value when absent or mismatched.
		if strings.TrimSpace(text) == "" {
			return moderationFieldError("text input items require non-empty text", "input", "invalid_input")
		}
	case "image_url":
		switch value := item["image_url"].(type) {
		case string:
			if strings.TrimSpace(value) == "" {
				return moderationFieldError("image_url input items require an image URL", "input", "invalid_input")
			}
		case map[string]interface{}:
			url, _ := value["url"].(string) //nolint:errcheck // Optional metadata uses the zero value when absent or mismatched.
			if strings.TrimSpace(url) == "" {
				return moderationFieldError("image_url input items require an image URL", "input", "invalid_input")
			}
		default:
			return moderationFieldError("image_url input items require an image URL", "input", "invalid_input")
		}
	default:
		return moderationFieldError("moderation input item type must be text or image_url", "input", "unsupported_input_type")
	}

	return nil
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

func supportsModerationsProvider(provider string) bool {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "openai", "openrouter", "mistral", "mistral-ai", "mock":
		return true
	default:
		return false
	}
}
