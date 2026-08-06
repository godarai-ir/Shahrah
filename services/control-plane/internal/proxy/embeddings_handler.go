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

// EmbeddingsHandler serves embeddings requests.
type EmbeddingsHandler struct {
	llmRouter  *router.Router
	pipeline   *RequestPipeline
	billing    budgeter
	logs       usageLogger
	logger     *zap.Logger
	emitMetric func(input metricsobs.BuildInput)
}

type embeddingValidationError struct {
	Message string
	Param   string
	Code    string
}

func (e *embeddingValidationError) Error() string {
	return e.Message
}

// NewEmbeddingsHandler creates a new embeddings handler.
func NewEmbeddingsHandler(llmRouter *router.Router, tunnel platform.TunnelTransport, billing budgeter, wallets walletStore, logs usageLogger, logger *zap.Logger, emitMetric ...func(input metricsobs.BuildInput)) *EmbeddingsHandler {
	var metricEmitter func(input metricsobs.BuildInput)
	if len(emitMetric) > 0 {
		metricEmitter = emitMetric[0]
	}

	return NewEmbeddingsHandlerWithDependencies(BaseHandlerDependencies{
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

// NewEmbeddingsHandlerWithDependencies creates an embeddings handler from explicit dependencies.
func NewEmbeddingsHandlerWithDependencies(deps BaseHandlerDependencies) *EmbeddingsHandler {
	return &EmbeddingsHandler{
		llmRouter:  deps.Router,
		pipeline:   deps.requestPipeline(),
		billing:    deps.Billing,
		logs:       deps.logger(),
		logger:     deps.Logger,
		emitMetric: deps.metricEmitter(),
	}
}

// Handle processes an embeddings request.
func (h *EmbeddingsHandler) Handle(c *fiber.Ctx) error {
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

	req, err := parseEmbeddingRequest(originalBody)
	if err != nil {
		if vErr, ok := err.(*embeddingValidationError); ok {
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

	modelAccess := modelAccessForType(authInfo, requestedModel, platform.ModelTypeEmbedding)
	if !modelAccess.Allowed {
		if modelAccess.UnsupportedVMMode {
			return writeUnsupportedModelModeError(c, requestedModel, platform.ModelTypeEmbedding, modelAccess.ExposedModes)
		}

		denyReason := denyReasonForModel(authInfo)
		emitDenyCounter(h.logger, denyReason, requestedModel, currentTenant.ID)

		const message = "Model access denied or unsupported for embeddings"
		logPreProviderRejection(h.logger, h.logs, c, authInfo, requestedModel, "/v1/embeddings", "embeddings", fiber.StatusForbidden, "authorization_error", message, denyReason, startTime)

		return writeAPIError(c, fiber.StatusForbidden, message, "authorization_error")
	}

	authStart := time.Now()
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "auth.validate", "internal", authStart, fiber.StatusOK, map[string]interface{}{"model": requestedModel})

	estimateReq := req
	estimateReq.Model = billingEstimateModel(requestedModel)
	quotaStart := time.Now()

	estimated, err := h.billing.EstimateEmbeddingCost(c.Context(), estimateReq)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, "Failed to estimate request cost", "billing_error")
	}

	rateLimit, err := h.billing.CheckRateLimit(c.Context(), authInfo, requestedModel, estimated.TotalTokens)
	setRateLimitHeaders(c, rateLimit)

	if err != nil {
		recordRateLimitRejection(c, h.logger, h.logs, h.emitMetric, authInfo, requestedModel, "/v1/embeddings", "embeddings", startTime, quotaStart, rateLimit, err)
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

	if !supportsEmbeddingsProvider(dep.Provider) {
		return writeAPIError(c, fiber.StatusNotImplemented, fmt.Sprintf("Embeddings API is not enabled for provider: %s", dep.Provider), "invalid_request_error")
	}

	if isMistralProvider(dep.Provider) && !mistralEndpointEnabled("embeddings") {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Mistral AI embeddings are disabled", "provider_disabled")
	}

	if isGoogleGeminiProvider(dep.Provider) {
		if endpointErr := validateGoogleGeminiRuntimeEndpoint(
			estimated.ModelRuntimeProfile,
			"embeddings",
		); endpointErr != nil {
			return writeAPIError(
				c,
				fiber.StatusBadRequest,
				endpointErr.Error(),
				"invalid_request_error",
			)
		}
	}

	if isMistralProvider(dep.Provider) {
		if endpointErr := validateMistralJSONParameters(estimated.ModelRuntimeProfile, "embeddings", originalBody); endpointErr != nil {
			return writeAPIError(c, fiber.StatusBadRequest, endpointErr.Error(), "invalid_request_error")
		}

		if endpointErr := validateMistralEmbeddingRequest(originalBody); endpointErr != nil {
			return writeAPIError(c, fiber.StatusBadRequest, endpointErr.Error(), "invalid_request_error")
		}
	}

	estimated, err = applyDeploymentPricingOverride(c.Context(), h.billing, estimated, dep)
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

	tunnelReq := platform.TunnelRequest{
		RequestID:  requestID,
		TenantID:   currentTenant.ID,
		TraceID:    requestID,
		Stream:     false,
		Operation:  platform.ProviderOperationOpenAIEmbeddings,
		Deployment: toContractDeployment(dep),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: payload,
		RequestMetadata: map[string]interface{}{
			"tenant_slug":            apiheaders.GetTenant(c),
			"model":                  requestedModel,
			"provider_endpoint_path": "/embeddings",
			"input_count":            embeddingInputCount(req.Input),
			"encoding_format":        embeddingEncodingValue(req.EncodingFormat),
		},
	}

	tunnelReq.ModelRuntimeProfiles, err = runtimeProfilesFromEstimate(dep.Provider, outboundModel, estimated)
	if err != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, err.Error(), "runtime_profile_unavailable")
	}

	if req.Dimensions != nil {
		tunnelReq.RequestMetadata["dimensions"] = *req.Dimensions
	}

	tunnelReq.RequestMetadata = enrichRequestMetadataForDeployment(tunnelReq.RequestMetadata, tunnelReq.Deployment)

	if !budgetReservationAttempted {
		budgetReservation, err = h.pipeline.Reserve(c.Context(), budgetRequestFromHTTP(c, authInfo, estimated, requestID, deploymentBillingMode(tunnelReq.Deployment)))
		if err != nil {
			return writeBudgetCheckError(c, err)
		}
	}

	err = h.handleSync(c, req, originalBody, authInfo, estimated, tunnelReq, budgetReservation, startTime, tenantSlug)
	if err == nil {
		releaseBudgetOnExit = false
	}

	return err
}

func (h *EmbeddingsHandler) handleSync(
	c *fiber.Ctx,
	req contractsopenai.EmbeddingRequest,
	originalBody []byte,
	authInfo *platform.AuthContext,
	estimated platform.UsageCostBreakdown,
	tunnelReq platform.TunnelRequest,
	budgetReservation *platform.BudgetReservation,
	start time.Time,
	tenantSlug string,
) error {
	providerStart := time.Now()
	providerAttrs := map[string]interface{}{
		"provider":        tunnelReq.Deployment.Provider,
		"request_payload": buildTraceEmbeddingRequestPayload(req),
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

	var response contractsopenai.EmbeddingResponse

	unmarshalErr := json.Unmarshal(resp.Body, &response)
	if unmarshalErr != nil {
		providerAttrs["error"] = "failed to decode provider response"
		providerAttrs["response_payload"] = buildTraceProviderErrorPayload(fiber.StatusBadGateway, resp.Body)
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, fiber.StatusBadGateway, providerAttrs)

		return writeAPIError(c, fiber.StatusBadGateway, "Failed to decode provider response", "api_error")
	}

	providerAttrs["response_payload"] = buildTraceEmbeddingResponsePayload(response)
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, resp.StatusCode, providerAttrs)

	finalCost, err := h.billing.FinalizeFromUsage(c.Context(), billingModelForDeployment(req.Model, tunnelReq.Deployment), response.Usage, estimated)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, "Failed to finalize usage", "billing_error")
	}

	if debitErr := h.pipeline.Settle(c.Context(), SettlementRequest{Reservation: budgetReservation, TenantID: authInfo.TenantID, Actual: finalCost, RequestID: tunnelReq.RequestID, BillingMode: deploymentBillingMode(tunnelReq.Deployment)}); debitErr != nil {
		return writeBudgetCheckError(c, debitErr)
	}

	responseTimeMS := int(time.Since(start).Milliseconds())
	vectorDimensions := embeddingResponseVectorDimensions(response)
	logEntry := buildRequestLogEntry(c, authInfo, tunnelReq.Deployment, req.Model, "/v1/embeddings", "embeddings", "success", resp.StatusCode, start, responseTimeMS)
	logEntry.RequestID = tunnelReq.RequestID
	logEntry.TraceID = tunnelReq.TraceID
	logEntry.CostUSD = finalCost.TotalCostUSD
	logEntry.CostIRR = finalCost.TotalCostIRR
	logEntry.InputTokens = finalCost.InputTokens
	logEntry.OutputTokens = finalCost.OutputTokens
	logEntry.TotalTokens = finalCost.TotalTokens
	logEntry.RequestBody = buildEmbeddingRequestLogPayload(req, originalBody)
	logEntry.ResponseBody = buildEmbeddingResponseLogPayload(response, resp.Body)
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
		RequestMetadata: enrichPricingMetadata(withEmbeddingMetadata(tunnelReq.RequestMetadata, vectorDimensions), authInfo, req.Model),
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
		RequestType:        "embeddings",
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
			"endpoint":            "/v1/embeddings",
			"key_type":            authInfo.KeyType,
			"team_id":             authInfo.TeamID,
			"org_unit_id":         authInfo.OrgUnitID,
			"tenant_slug":         tenantSlug,
			"input_count":         embeddingInputCount(req.Input),
			"encoding_format":     embeddingEncodingValue(req.EncodingFormat),
			"vector_dimensions":   vectorDimensions,
			"requested_dimension": req.Dimensions,
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
		"response_payload": buildTraceEmbeddingResponsePayload(response),
	})

	return err
}

func (h *EmbeddingsHandler) emitModelMetric(input metricsobs.BuildInput) {
	if h.emitMetric == nil {
		return
	}

	h.emitMetric(input)
}

func buildEmbeddingRequestLogPayload(req contractsopenai.EmbeddingRequest, originalBody []byte) map[string]interface{} {
	return buildJSONRequestLogPayload(originalBody, map[string]interface{}{
		"model": req.Model,
		"input": req.Input,
	})
}

func buildEmbeddingResponseLogPayload(resp contractsopenai.EmbeddingResponse, body []byte) map[string]interface{} {
	return buildJSONResponseLogPayload(body, resp)
}

func parseEmbeddingRequest(body []byte) (contractsopenai.EmbeddingRequest, error) {
	var (
		req contractsopenai.EmbeddingRequest
		raw map[string]json.RawMessage
	)
	if err := json.Unmarshal(body, &raw); err != nil {
		return req, fmt.Errorf("invalid request body")
	}

	if err := json.Unmarshal(body, &req); err != nil {
		return req, fmt.Errorf("invalid request body")
	}

	req.Model = strings.TrimSpace(req.Model)
	req.EncodingFormat = strings.TrimSpace(strings.ToLower(req.EncodingFormat))
	req.User = strings.TrimSpace(req.User)

	if _, ok := raw["stream"]; ok {
		return req, apiFieldError("stream is not supported for embeddings", "stream", "unsupported_parameter")
	}

	for _, field := range []string{"messages", "temperature", "max_tokens", "tools", "tool_choice", "max_output_tokens"} {
		if _, ok := raw[field]; ok {
			return req, apiFieldError(fmt.Sprintf("%s is not supported for embeddings", field), field, "unsupported_parameter")
		}
	}

	if req.Model == "" {
		return req, apiFieldError("model is required", "model", "invalid_model")
	}

	normalizedInput, err := normalizeEmbeddingInput(req.Input)
	if err != nil {
		return req, err
	}

	req.Input = normalizedInput
	if req.EncodingFormat == "" {
		req.EncodingFormat = "float"
	}

	if req.EncodingFormat != "float" && req.EncodingFormat != "base64" {
		return req, apiFieldError("encoding_format must be either 'float' or 'base64'", "encoding_format", "invalid_encoding_format")
	}

	if req.Dimensions != nil {
		if *req.Dimensions <= 0 {
			return req, apiFieldError("dimensions must be a positive integer", "dimensions", "invalid_dimensions")
		}

		if !modelSupportsCustomDimensions(req.Model) {
			return req, apiFieldError("dimensions is not supported for the selected embedding model", "dimensions", "unsupported_parameter")
		}
	}

	return req, nil
}

func normalizeEmbeddingInput(input interface{}) (interface{}, error) {
	switch v := input.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return nil, apiFieldError("input is required and must be a non-empty string or a non-empty array of strings", "input", "invalid_input")
		}

		return v, nil
	case []interface{}:
		if len(v) == 0 {
			return nil, apiFieldError("input is required and must be a non-empty string, token array, or array of inputs", "input", "invalid_input")
		}

		if allEmbeddingTokens(v) {
			return v, nil
		}

		allStrings := true
		allTokenArrays := true

		for _, item := range v {
			if s, ok := item.(string); !ok || strings.TrimSpace(s) == "" {
				allStrings = false
			}

			nested, ok := item.([]interface{})
			if !ok || len(nested) == 0 || !allEmbeddingTokens(nested) {
				allTokenArrays = false
			}
		}

		if allStrings || allTokenArrays {
			return v, nil
		}

		return nil, apiFieldError("input must be a string, array of strings, token array, or array of token arrays", "input", "unsupported_input_type")
	default:
		if input == nil {
			return nil, apiFieldError("input is required and must be a non-empty string, token array, or array of inputs", "input", "invalid_input")
		}

		return nil, apiFieldError("input must be a string, array of strings, token array, or array of token arrays", "input", "unsupported_input_type")
	}
}

func allEmbeddingTokens(values []interface{}) bool {
	for _, value := range values {
		number, ok := value.(float64)
		if !ok || number < 0 || number != float64(int64(number)) {
			return false
		}
	}

	return len(values) > 0
}

func apiFieldError(message string, param string, code string) error {
	return &embeddingValidationError{Message: message, Param: param, Code: code}
}

func supportsEmbeddingsProvider(provider string) bool {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "openai", "azure", "groq", "mistral", "mistral-ai", "togetherai", "openrouter", "gemini", "google-gemini", "mock":
		return true
	default:
		return false
	}
}

func modelSupportsCustomDimensions(model string) bool {
	model = providerModelID(model)

	return model == "text-embedding-3-small" ||
		model == "text-embedding-3-large" ||
		model == "codestral-embed-2505" ||
		model == "codestral-embed"
}

func embeddingInputCount(input interface{}) int {
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

func embeddingEncodingValue(v string) string {
	if strings.TrimSpace(v) == "" {
		return "float"
	}

	return v
}

func embeddingResponseVectorDimensions(resp contractsopenai.EmbeddingResponse) []int {
	out := make([]int, 0, len(resp.Data))
	for _, item := range resp.Data {
		var values []float64
		if err := json.Unmarshal(item.Embedding, &values); err == nil {
			out = append(out, len(values))
			continue
		}

		out = append(out, 0)
	}

	return out
}

func withEmbeddingMetadata(input map[string]interface{}, vectorDimensions []int) map[string]interface{} {
	out := map[string]interface{}{}
	for k, v := range input {
		out[k] = v
	}

	out["vector_dimensions"] = vectorDimensions

	return out
}
