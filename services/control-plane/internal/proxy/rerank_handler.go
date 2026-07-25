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

// RerankHandler serves Cohere-compatible rerank requests.
type RerankHandler struct {
	llmRouter  *router.Router
	tunnel     platform.TunnelTransport
	billing    budgeter
	wallets    walletStore
	logs       usageLogger
	logger     *zap.Logger
	emitMetric func(input metricsobs.BuildInput)
}

type rerankValidationError struct {
	Message string
	Param   string
	Code    string
}

func (e *rerankValidationError) Error() string {
	return e.Message
}

// NewRerankHandler creates a new rerank handler.
func NewRerankHandler(llmRouter *router.Router, tunnel platform.TunnelTransport, billing budgeter, wallets walletStore, logs usageLogger, logger *zap.Logger, emitMetric ...func(input metricsobs.BuildInput)) *RerankHandler {
	var metricEmitter func(input metricsobs.BuildInput)
	if len(emitMetric) > 0 {
		metricEmitter = emitMetric[0]
	}

	return NewRerankHandlerWithDependencies(BaseHandlerDependencies{
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

// NewRerankHandlerWithDependencies creates a rerank handler from explicit dependencies.
func NewRerankHandlerWithDependencies(deps BaseHandlerDependencies) *RerankHandler {
	return &RerankHandler{
		llmRouter:  deps.Router,
		tunnel:     deps.Tunnel,
		billing:    deps.Billing,
		wallets:    deps.Wallets,
		logs:       deps.logger(),
		logger:     deps.Logger,
		emitMetric: deps.metricEmitter(),
	}
}

// Handle processes a Cohere-compatible rerank request.
func (h *RerankHandler) Handle(c *fiber.Ctx) error {
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

	req, err := parseRerankRequest(originalBody)
	if err != nil {
		if vErr, ok := err.(*rerankValidationError); ok {
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

	modelAccess := modelAccessForType(authInfo, requestedModel, platform.ModelTypeRerank)
	if !modelAccess.Allowed {
		if modelAccess.UnsupportedVMMode {
			return writeUnsupportedModelModeError(c, requestedModel, platform.ModelTypeRerank, modelAccess.ExposedModes)
		}

		denyReason := denyReasonForModel(authInfo)
		emitDenyCounter(h.logger, denyReason, requestedModel, currentTenant.ID)

		const message = "Model access denied or unsupported for rerank"
		logPreProviderRejection(h.logger, h.logs, c, authInfo, requestedModel, "/v1/rerank", "rerank", fiber.StatusForbidden, "authorization_error", message, denyReason, startTime)

		return writeAPIError(c, fiber.StatusForbidden, message, "authorization_error")
	}

	authStart := time.Now()
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "auth.validate", "internal", authStart, fiber.StatusOK, map[string]interface{}{"model": requestedModel})

	estimateReq := contractsopenai.EmbeddingRequest{
		Model: providerModelID(requestedModel),
		Input: rerankEstimateInput(req),
	}
	quotaStart := time.Now()

	estimated, err := h.billing.EstimateEmbeddingCost(c.Context(), estimateReq)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, "Failed to estimate request cost", "billing_error")
	}

	rateLimit, err := h.billing.CheckRateLimit(c.Context(), authInfo, requestedModel, estimated.TotalTokens)
	setRateLimitHeaders(c, rateLimit)

	if err != nil {
		recordRateLimitRejection(c, h.logger, h.logs, h.emitMetric, authInfo, requestedModel, "/v1/rerank", "rerank", startTime, quotaStart, rateLimit, err)
		return writeAPIError(c, rateLimitErrorStatus(err), err.Error(), "rate_limit_error")
	}

	requestID := tracing.TraceIDFromContext(c)
	if requestID == "" {
		requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
	}

	var (
		budgetReservation     *platform.BudgetReservation
		budgetResponseHandled bool
	)

	budgetReservationAttempted := false
	if !shouldDeferBudgetReservationUntilRoute(requestedModel) {
		budgetReservationAttempted = true

		budgetReservation, budgetResponseHandled, err = reserveBudgetForRequest(c, h.billing, authInfo, estimated, requestID)
		if err != nil || budgetResponseHandled {
			return err
		}
	}

	releaseBudgetOnExit := true
	defer func() {
		if releaseBudgetOnExit {
			releaseBudgetBestEffort(c.Context(), h.logger, h.billing, budgetReservation, "request_not_committed")
		}
	}()

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
		if handled, dependencyErr := writeRoutingDependencyUnavailable(c, h.logger, err, requestedModel, currentTenant.ID); handled {
			return dependencyErr
		}

		emitDenyCounter(h.logger, denyReasonUnknownModel, requestedModel, currentTenant.ID)

		return writeAPIError(c, fiber.StatusServiceUnavailable, fmt.Sprintf("No available deployment for model: %s", requestedModel), "routing_error")
	}

	if !supportsRerankProvider(dep.Provider) {
		return writeAPIError(c, fiber.StatusNotImplemented, fmt.Sprintf("Rerank API is not enabled for provider: %s", dep.Provider), "invalid_request_error")
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
			"provider_endpoint_path": rerankProviderEndpointPath(dep.Provider),
			"document_count":         len(req.Documents),
			"top_n":                  req.TopN,
		},
	}
	tunnelReq.RequestMetadata = enrichRequestMetadataForDeployment(tunnelReq.RequestMetadata, tunnelReq.Deployment)

	if !budgetReservationAttempted {
		budgetReservation, budgetResponseHandled, err = reserveBudgetForRequest(c, h.billing, authInfo, estimated, requestID, deploymentBillingMode(tunnelReq.Deployment))
		if err != nil || budgetResponseHandled {
			return err
		}
	}

	err = h.handleSync(c, req, originalBody, authInfo, estimated, tunnelReq, budgetReservation, startTime, tenantSlug)
	if err == nil {
		releaseBudgetOnExit = false
	}

	return err
}

func (h *RerankHandler) handleSync(c *fiber.Ctx, req contractsopenai.RerankRequest, originalBody []byte, authInfo *platform.AuthContext, estimated platform.UsageCostBreakdown, tunnelReq platform.TunnelRequest, budgetReservation *platform.BudgetReservation, start time.Time, tenantSlug string) error {
	providerStart := time.Now()
	providerAttrs := map[string]interface{}{
		"provider":        tunnelReq.Deployment.Provider,
		"request_payload": buildTraceRerankRequestPayload(req),
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

	var response contractsopenai.RerankResponse
	if decodeErr := json.Unmarshal(resp.Body, &response); decodeErr != nil {
		providerAttrs["error"] = "failed to decode provider response"
		providerAttrs["response_payload"] = buildTraceProviderErrorPayload(fiber.StatusBadGateway, resp.Body)
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, fiber.StatusBadGateway, providerAttrs)

		return writeAPIError(c, fiber.StatusBadGateway, "Failed to decode provider response", "api_error")
	}

	providerAttrs["response_payload"] = buildTraceRerankResponsePayload(response)
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, resp.StatusCode, providerAttrs)

	finalCost, err := h.billing.FinalizeFromUsage(c.Context(), billingModelForDeployment(req.Model, tunnelReq.Deployment), response.Usage, estimated)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, "Failed to finalize usage", "billing_error")
	}

	if debitErr := commitBudgetOrDebit(c.Context(), h.logger, h.billing, h.wallets, budgetReservation, authInfo.TenantID, finalCost, tunnelReq.RequestID, deploymentBillingMode(tunnelReq.Deployment)); debitErr != nil {
		return writeBudgetCheckError(c, debitErr)
	}

	responseTimeMS := int(time.Since(start).Milliseconds())
	logEntry := buildRequestLogEntry(c, authInfo, tunnelReq.Deployment, req.Model, "/v1/rerank", "rerank", "success", resp.StatusCode, start, responseTimeMS)
	logEntry.RequestID = tunnelReq.RequestID
	logEntry.TraceID = tunnelReq.TraceID
	logEntry.CostUSD = finalCost.TotalCostUSD
	logEntry.CostIRR = finalCost.TotalCostIRR
	logEntry.InputTokens = finalCost.InputTokens
	logEntry.OutputTokens = finalCost.OutputTokens
	logEntry.TotalTokens = finalCost.TotalTokens
	logEntry.RequestBody = buildRerankRequestLogPayload(req, originalBody)
	logEntry.ResponseBody = buildJSONResponseLogPayload(resp.Body, response)
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
		ProviderName:       tunnelReq.Deployment.Provider,
		ProviderModelName:  tunnelReq.Deployment.Model,
		RequestType:        "rerank",
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
			"endpoint":       "/v1/rerank",
			"key_type":       authInfo.KeyType,
			"team_id":        authInfo.TeamID,
			"org_unit_id":    authInfo.OrgUnitID,
			"tenant_slug":    tenantSlug,
			"document_count": len(req.Documents),
			"top_n":          req.TopN,
		}, tunnelReq.Deployment),
	})

	forwardProviderHeaders(c, resp.Headers)
	setTraceHeaders(c, tunnelReq.TraceID)
	setCostHeaders(c, finalCost)

	responseStart := time.Now()
	err = c.Status(resp.StatusCode).JSON(response)
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "response.write", "internal", responseStart, resp.StatusCode, map[string]interface{}{
		"response_payload": buildTraceRerankResponsePayload(response),
	})

	return err
}

func buildRerankRequestLogPayload(req contractsopenai.RerankRequest, originalBody []byte) map[string]interface{} {
	return buildJSONRequestLogPayload(originalBody, map[string]interface{}{
		"model":     req.Model,
		"query":     req.Query,
		"documents": req.Documents,
	})
}

func (h *RerankHandler) emitModelMetric(input metricsobs.BuildInput) {
	if h.emitMetric == nil {
		return
	}

	h.emitMetric(input)
}

func parseRerankRequest(body []byte) (contractsopenai.RerankRequest, error) {
	var (
		req contractsopenai.RerankRequest
		raw map[string]json.RawMessage
	)
	if err := json.Unmarshal(body, &raw); err != nil {
		return req, fmt.Errorf("invalid request body")
	}

	if err := json.Unmarshal(body, &req); err != nil {
		return req, fmt.Errorf("invalid request body")
	}

	req.Model = strings.TrimSpace(req.Model)
	req.Query = strings.TrimSpace(req.Query)

	if _, ok := raw["stream"]; ok {
		return req, rerankFieldError("stream is not supported for rerank", "stream", "unsupported_parameter")
	}

	for _, field := range []string{"messages", "input", "prompt", "tools", "tool_choice", "temperature", "top_p", "max_tokens", "max_output_tokens", "encoding_format", "dimensions"} {
		if _, ok := raw[field]; ok {
			return req, rerankFieldError(fmt.Sprintf("%s is not supported for rerank", field), field, "unsupported_parameter")
		}
	}

	if req.Model == "" {
		return req, rerankFieldError("model is required", "model", "invalid_model")
	}

	if req.Query == "" {
		return req, rerankFieldError("query is required", "query", "invalid_query")
	}

	if len(req.Documents) == 0 {
		return req, rerankFieldError("documents must be a non-empty array of strings", "documents", "invalid_documents")
	}

	if len(req.Documents) > 1000 {
		return req, rerankFieldError("documents must contain at most 1000 items", "documents", "invalid_documents")
	}

	for i, doc := range req.Documents {
		if strings.TrimSpace(doc) == "" {
			return req, rerankFieldError(fmt.Sprintf("documents[%d] must be a non-empty string", i), "documents", "invalid_documents")
		}
	}

	if req.TopN != nil {
		if *req.TopN < 1 {
			return req, rerankFieldError("top_n must be greater than or equal to 1", "top_n", "invalid_top_n")
		}

		if *req.TopN > len(req.Documents) {
			return req, rerankFieldError("top_n must be less than or equal to the number of documents", "top_n", "invalid_top_n")
		}
	}

	if req.MaxTokensPerDoc != nil && *req.MaxTokensPerDoc < 1 {
		return req, rerankFieldError("max_tokens_per_doc must be greater than or equal to 1", "max_tokens_per_doc", "invalid_max_tokens_per_doc")
	}

	if req.Priority != nil && (*req.Priority < 0 || *req.Priority > 999) {
		return req, rerankFieldError("priority must be between 0 and 999", "priority", "invalid_priority")
	}

	return req, nil
}

func rerankFieldError(message string, param string, code string) error {
	return &rerankValidationError{Message: message, Param: param, Code: code}
}

func rerankEstimateInput(req contractsopenai.RerankRequest) []string {
	out := make([]string, 0, len(req.Documents)+1)
	out = append(out, req.Query)
	out = append(out, req.Documents...)

	return out
}

func supportsRerankProvider(provider string) bool {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "cohere", "mock":
		return true
	default:
		return false
	}
}

func rerankProviderEndpointPath(provider string) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "cohere":
		return "/v2/rerank"
	default:
		return "/rerank"
	}
}
