// Package proxy hosts OpenAI-compatible request handlers for the control plane.
package proxy

import (
	"bufio"
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

// ResponsesHandler serves the OpenAI-compatible Responses API.
type ResponsesHandler struct {
	llmRouter    *router.Router
	pipeline     *RequestPipeline
	billing      budgeter
	resources    providerNativeResourceStore
	deployments  responseDeploymentLister
	providerAuth responseProviderAuthResolver
	logs         usageLogger
	logger       *zap.Logger
	emitMetric   func(input metricsobs.BuildInput)
}

// NewResponsesHandlerWithDependencies creates a responses handler from explicit dependencies.
func NewResponsesHandlerWithDependencies(deps BaseHandlerDependencies) *ResponsesHandler {
	return &ResponsesHandler{
		llmRouter:    deps.Router,
		pipeline:     deps.requestPipeline(),
		billing:      deps.Billing,
		resources:    deps.ProviderResources,
		deployments:  deps.Deployments,
		providerAuth: deps.ResponseProviderAuth,
		logs:         deps.logger(),
		logger:       deps.Logger,
		emitMetric:   deps.metricEmitter(),
	}
}

// NewResponsesHandler creates a new responses handler.
func NewResponsesHandler(llmRouter *router.Router, tunnel platform.TunnelTransport, billing budgeter, wallets walletStore, logs usageLogger, logger *zap.Logger, emitMetric ...func(input metricsobs.BuildInput)) *ResponsesHandler {
	var metricEmitter func(input metricsobs.BuildInput)
	if len(emitMetric) > 0 {
		metricEmitter = emitMetric[0]
	}

	return NewResponsesHandlerWithDependencies(BaseHandlerDependencies{
		Router:              llmRouter,
		Tunnel:              tunnel,
		Billing:             billing,
		BillingCapabilities: compatibilityBillingCapabilities(),
		Wallets:             wallets,
		RequestLogger:       logs,
		SpendLogger:         logs,
		Logger:              logger,
		EmitMetric:          metricEmitter,
	})
}

// Handle processes an OpenAI-compatible Responses request.
func (h *ResponsesHandler) Handle(c *fiber.Ctx) error {
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

	var req contractsopenai.ResponseRequest
	if err := json.Unmarshal(originalBody, &req); err != nil {
		return writeAPIError(c, fiber.StatusBadRequest, "Invalid request body", "invalid_request_error")
	}

	modelWasProvided := strings.TrimSpace(req.Model) != ""
	requestedModel := normalizeModelIdentifier(req.Model)

	resolvedDeployment, effectiveModel, upstreamPreviousResponseID, err := h.resolveResponseCreateDeployment(
		c.Context(),
		currentTenant.ID,
		requestedModel,
		strings.TrimSpace(req.PreviousResponseID),
		strings.TrimSpace(c.Get("x-godarai-provider-account")),
	)
	if err != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, err.Error(), "routing_error")
	}

	requestedModel = effectiveModel
	req.Model = requestedModel

	if requestedModel == "" {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Unable to determine the model for Responses routing and billing", "routing_error")
	}

	tracing.SetModelForTrace(c, requestedModel)

	modelAccess := modelAccessForType(authInfo, requestedModel, platform.ModelTypeResponses)
	if !modelAccess.Allowed {
		if modelAccess.UnsupportedVMMode {
			return writeUnsupportedModelModeError(c, requestedModel, platform.ModelTypeResponses, modelAccess.ExposedModes)
		}

		denyReason := denyReasonForModel(authInfo)
		emitDenyCounter(h.logger, denyReason, requestedModel, currentTenant.ID)

		const message = "Model access denied or unsupported for responses"
		logPreProviderRejection(h.logger, h.logs, c, authInfo, requestedModel, "/v1/responses", "responses", fiber.StatusForbidden, "authorization_error", message, denyReason, startTime)

		return writeAPIError(c, fiber.StatusForbidden, message, "authorization_error")
	}

	authStart := time.Now()
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "auth.validate", "internal", authStart, fiber.StatusOK, map[string]interface{}{"model": requestedModel})

	estimateReq := responseRequestToChatCompletion(req, requestedModel)
	estimateReq.Model = billingEstimateModel(requestedModel)
	quotaStart := time.Now()

	estimated, err := h.billing.EstimateCost(c.Context(), estimateReq)
	if err != nil {
		return writeBillingEstimateError(c, err)
	}

	estimated.Metadata = req.Metadata

	rateLimit, err := h.billing.CheckRateLimit(c.Context(), authInfo, requestedModel, estimated.TotalTokens)
	setRateLimitHeaders(c, rateLimit)

	if err != nil {
		recordRateLimitRejection(c, h.logger, h.logs, h.emitMetric, authInfo, requestedModel, "/v1/responses", "responses", startTime, quotaStart, rateLimit, err)
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

	var (
		dep              *router.Deployment
		tunnelDeployment platform.Deployment
	)
	if resolvedDeployment != nil {
		tunnelDeployment = *resolvedDeployment
	} else {
		for _, candidate := range modelRoutingCandidates(requestedModel) {
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

		tunnelDeployment = toContractDeployment(dep)
	}

	if !supportsResponsesCreateProvider(tunnelDeployment.Provider) {
		return writeAPIError(c, fiber.StatusNotImplemented, fmt.Sprintf("Responses API is not enabled for provider: %s", tunnelDeployment.Provider), "invalid_request_error")
	}

	if dep != nil {
		estimated, err = applyDeploymentPricingOverride(c.Context(), h.pipeline, estimated, dep)
		if err != nil {
			return writeBillingEstimateError(c, err)
		}
	}

	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "router.select_model", "internal", routeStart, fiber.StatusOK, map[string]interface{}{"provider": tunnelDeployment.Provider, "deployment_id": tunnelDeployment.ID})

	payload := originalBody

	if modelWasProvided {
		outboundModel := strings.TrimSpace(tunnelDeployment.Model)
		if outboundModel == "" || outboundModel == "*" {
			outboundModel = providerModelID(requestedModel)
		}

		payload, err = rewriteOpenAIJSONModel(payload, outboundModel)
		if err != nil {
			return writeAPIError(c, fiber.StatusBadRequest, "Failed to encode request", "invalid_request_error")
		}
	}

	if upstreamPreviousResponseID != "" && upstreamPreviousResponseID != strings.TrimSpace(req.PreviousResponseID) {
		payload, err = rewriteOpenAIJSONStringField(payload, "previous_response_id", upstreamPreviousResponseID)
		if err != nil {
			return writeAPIError(c, fiber.StatusBadRequest, "Failed to encode request", "invalid_request_error")
		}
	}

	if strings.EqualFold(strings.TrimSpace(tunnelDeployment.Provider), "xai") {
		payload, err = rewriteXAIFileReferences(c.Context(), h.resources, currentTenant.ID, tunnelDeployment.ProviderAccountID, payload)
		if err != nil {
			return writeXAIFileReferenceError(c, err)
		}
	}

	tunnelReq := platform.TunnelRequest{
		RequestID:  requestID,
		TenantID:   currentTenant.ID,
		TraceID:    requestID,
		Stream:     req.Stream,
		Operation:  platform.ProviderOperationOpenAIResponsesCreate,
		Deployment: tunnelDeployment,
		Headers:    responseProviderHeaders(c),
		Body:       payload,
		RequestMetadata: map[string]interface{}{
			"tenant_slug":            apiheaders.GetTenant(c),
			"model":                  requestedModel,
			"provider_endpoint_path": "/responses",
		},
	}
	tunnelReq.RequestMetadata = enrichRequestMetadataForDeployment(tunnelReq.RequestMetadata, tunnelReq.Deployment)

	runtimeModel := strings.TrimSpace(tunnelReq.Deployment.Model)
	if runtimeModel == "" || runtimeModel == "*" {
		runtimeModel = providerModelID(requestedModel)
	}

	tunnelReq.ModelRuntimeProfiles, err = runtimeProfilesFromEstimate(
		tunnelReq.Deployment.Provider,
		runtimeModel,
		estimated,
	)
	if err != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, err.Error(), "runtime_profile_unavailable")
	}

	if isDeepSeekProvider(tunnelReq.Deployment.Provider) {
		if len(tunnelReq.ModelRuntimeProfiles) != 1 {
			return writeAPIError(c, fiber.StatusServiceUnavailable, "DeepSeek model runtime profile is unavailable", "runtime_profile_unavailable")
		}

		validationErr := validateDeepSeekResponsesRuntimeProfile(payload, &tunnelReq.ModelRuntimeProfiles[0])
		if validationErr != nil {
			return writeAPIError(c, fiber.StatusBadRequest, validationErr.Error(), "invalid_request_error")
		}
	}

	if !budgetReservationAttempted {
		budgetReservation, err = h.pipeline.Reserve(c.Context(), budgetRequestFromHTTP(c, authInfo, estimated, requestID, deploymentBillingMode(tunnelReq.Deployment)))
		if err != nil {
			return writeBudgetCheckError(c, err)
		}
	}

	if req.Stream {
		releaseBudgetOnExit = false
		return h.handleStream(c, req, authInfo, estimated, tunnelReq, budgetReservation, startTime, tenantSlug)
	}

	err = h.handleSync(c, req, originalBody, authInfo, estimated, tunnelReq, budgetReservation, startTime, tenantSlug)
	if err == nil {
		releaseBudgetOnExit = false
	}

	return err
}

func (h *ResponsesHandler) handleSync(
	c *fiber.Ctx,
	req contractsopenai.ResponseRequest,
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
		"request_payload": buildProviderTraceResponseRequestPayload(tunnelReq.Deployment.Provider, req),
	}

	resp, err := h.pipeline.Dispatch(c.Context(), tunnelReq)
	if err != nil {
		providerAttrs["error"] = err.Error()
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, fiber.StatusBadGateway, providerAttrs)

		return writeAPIError(c, fiber.StatusBadGateway, fmt.Sprintf("Tunnel error: %v", err), "api_error")
	}

	if resp.StatusCode >= httpErrorFloor {
		providerAttrs["response_payload"] = buildProviderTraceErrorPayload(tunnelReq.Deployment.Provider, resp.StatusCode, resp.Body)
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, resp.StatusCode, providerAttrs)

		return writeOpenAIProviderError(c, resp.StatusCode, resp.Body, resp.Headers)
	}

	var response contractsopenai.Response

	unmarshalErr := json.Unmarshal(resp.Body, &response)
	if unmarshalErr != nil {
		providerAttrs["error"] = "failed to decode provider response"
		providerAttrs["response_payload"] = buildProviderTraceErrorPayload(tunnelReq.Deployment.Provider, fiber.StatusBadGateway, resp.Body)
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, fiber.StatusBadGateway, providerAttrs)

		return writeAPIError(c, fiber.StatusBadGateway, "Failed to decode provider response", "api_error")
	}

	providerAttrs["response_payload"] = buildProviderTraceResponsePayload(tunnelReq.Deployment.Provider, response)
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, resp.StatusCode, providerAttrs)
	h.recordResponseMapping(c.Context(), authInfo.TenantID, tunnelReq.Deployment, response.ID, response.Status, response.Model)

	finalUsage := responseUsageWithToolsToOpenAI(response)
	if finalUsage != nil {
		finalUsage.ServiceTier = response.ServiceTier
	}

	finalCost, err := h.billing.FinalizeFromUsage(c.Context(), billingModelForDeployment(req.Model, tunnelReq.Deployment), finalUsage, estimated)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, "Failed to finalize usage", "billing_error")
	}

	tunnelReq.RequestMetadata = withXAIUsageMetadata(tunnelReq.RequestMetadata, tunnelReq.Deployment.Provider, finalUsage)

	if debitErr := h.pipeline.Settle(c.Context(), SettlementRequest{Reservation: budgetReservation, TenantID: authInfo.TenantID, Actual: finalCost, RequestID: tunnelReq.RequestID, BillingMode: deploymentBillingMode(tunnelReq.Deployment)}); debitErr != nil {
		return writeBudgetCheckError(c, debitErr)
	}

	responseTimeMS := int(time.Since(start).Milliseconds())
	logEntry := buildRequestLogEntry(c, authInfo, tunnelReq.Deployment, req.Model, "/v1/responses", "responses", "success", resp.StatusCode, start, responseTimeMS)
	logEntry.RequestID = tunnelReq.RequestID
	logEntry.TraceID = tunnelReq.TraceID
	logEntry.CostUSD = finalCost.TotalCostUSD
	logEntry.CostIRR = finalCost.TotalCostIRR
	logEntry.InputTokens = finalCost.InputTokens
	logEntry.OutputTokens = finalCost.OutputTokens
	logEntry.TotalTokens = finalCost.TotalTokens
	logEntry.RequestBody = buildProviderResponseRequestLogPayload(tunnelReq.Deployment.Provider, req, originalBody)
	logEntry.ResponseBody = buildProviderResponseLogPayload(tunnelReq.Deployment.Provider, resp.Body, response)
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
		RequestType:        "responses",
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
			"endpoint":    "/v1/responses",
			"key_type":    authInfo.KeyType,
			"team_id":     authInfo.TeamID,
			"org_unit_id": authInfo.OrgUnitID,
			"tenant_slug": tenantSlug,
		}, tunnelReq.Deployment),
	})

	forwardProviderHeaders(c, resp.Headers)
	setTraceHeaders(c, tunnelReq.TraceID)
	setCostHeaders(c, finalCost)

	responseStart := time.Now()

	c.Status(resp.StatusCode)
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	err = c.Send(resp.Body)
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "response.write", "internal", responseStart, resp.StatusCode, map[string]interface{}{
		"response_payload": buildProviderTraceResponsePayload(tunnelReq.Deployment.Provider, response),
	})

	return err
}

// rewriteOpenAIJSONModel changes only the routing-owned model field. Using
// RawMessage values preserves every other official or future OpenAI field
// instead of round-tripping through the gateway's necessarily incomplete DTO.
func rewriteOpenAIJSONModel(body []byte, model string) ([]byte, error) {
	return rewriteOpenAIJSONStringField(body, "model", providerModelID(model))
}

func rewriteOpenAIJSONStringField(body []byte, field string, value string) ([]byte, error) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(body, &object); err != nil {
		return nil, err
	}

	encodedValue, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	object[field] = encodedValue

	return json.Marshal(object)
}

func buildResponseRequestLogPayload(req contractsopenai.ResponseRequest, originalBody []byte) map[string]interface{} {
	return buildJSONRequestLogPayload(originalBody, map[string]interface{}{
		"model": req.Model,
		"input": req.Input,
	})
}

func buildProviderResponseRequestLogPayload(provider string, req contractsopenai.ResponseRequest, originalBody []byte) map[string]interface{} {
	if isDeepSeekProvider(provider) {
		return buildProviderTraceResponseRequestPayload(provider, req)
	}

	return buildResponseRequestLogPayload(req, originalBody)
}

func buildProviderResponseLogPayload(provider string, body []byte, response contractsopenai.Response) map[string]interface{} {
	if isDeepSeekProvider(provider) {
		return buildProviderTraceResponsePayload(provider, response)
	}

	return buildJSONResponseLogPayload(body, response)
}

func (h *ResponsesHandler) handleStream(c *fiber.Ctx, req contractsopenai.ResponseRequest, authInfo *platform.AuthContext, estimated platform.UsageCostBreakdown, tunnelReq platform.TunnelRequest, budgetReservation *platform.BudgetReservation, start time.Time, tenantSlug string) error {
	providerStart := time.Now()

	stream, err := h.pipeline.DispatchStream(c.Context(), tunnelReq)
	if err != nil {
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, fiber.StatusBadGateway, map[string]interface{}{
			"provider":        tunnelReq.Deployment.Provider,
			"request_payload": buildProviderTraceResponseRequestPayload(tunnelReq.Deployment.Provider, req),
			"stream":          true,
			"error":           err.Error(),
		})

		h.pipeline.Release(c.Context(), budgetReservation, "stream_open_failed")

		return writeAPIError(c, fiber.StatusBadGateway, fmt.Sprintf("Tunnel error: %v", err), "api_error")
	}

	if statusCode := stream.StatusCode(); statusCode >= httpErrorFloor {
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, statusCode, map[string]interface{}{
			"provider":        tunnelReq.Deployment.Provider,
			"request_payload": buildProviderTraceResponseRequestPayload(tunnelReq.Deployment.Provider, req),
			"stream":          true,
			"status_code":     statusCode,
		})
		h.pipeline.Release(c.Context(), budgetReservation, "provider_http_error")

		return writeOpenAITunnelStreamHTTPError(c, h.logger, stream)
	}

	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, fiber.StatusOK, map[string]interface{}{
		"provider":        tunnelReq.Deployment.Provider,
		"request_payload": buildProviderTraceResponseRequestPayload(tunnelReq.Deployment.Provider, req),
		"stream":          true,
	})

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	setTraceHeaders(c, tunnelReq.TraceID)
	forwardProviderHeaders(c, stream.Headers())

	requestCtx := c.Context()
	requestCtx.SetBodyStreamWriter(func(w *bufio.Writer) {
		defer closeStreamBestEffort(h.logger, stream)

		budgetCommitted := false

		var terminalResponse *contractsopenai.Response

		defer func() {
			if !budgetCommitted {
				h.pipeline.Release(requestCtx, budgetReservation, "stream_not_committed")
			}
		}()

		finalizeAndLogSuccess := func() {
			if budgetCommitted {
				return
			}

			if isDeepSeekProvider(tunnelReq.Deployment.Provider) && terminalResponse == nil {
				h.logger.Warn(
					"deepseek Responses stream closed without a terminal event",
					zap.String("request_id", tunnelReq.RequestID),
				)

				framed, frameErr := frameOpenAIResponseStreamError(
					fmt.Errorf("provider stream closed without a terminal Responses event"),
				)
				if frameErr != nil {
					h.logger.Warn("failed to frame incomplete DeepSeek Responses stream", zap.Error(frameErr))

					return
				}

				if writeErr := writeBufferedBytes(w, framed); writeErr != nil {
					h.logger.Warn("failed to write incomplete DeepSeek Responses stream error", zap.Error(writeErr))
				}

				return
			}

			billingModel := billingModelForDeployment(req.Model, tunnelReq.Deployment)

			var finalUsage *contractsopenai.Usage
			if terminalResponse != nil {
				finalUsage = responseUsageWithToolsToOpenAI(*terminalResponse)
				if finalUsage != nil {
					finalUsage.ServiceTier = terminalResponse.ServiceTier
				}
			}

			finalCost, finalizeErr := h.billing.FinalizeFromUsage(requestCtx, billingModel, finalUsage, estimated)
			if finalizeErr != nil {
				h.logger.Warn("failed to finalize streaming usage", zap.Error(finalizeErr), zap.String("request_id", tunnelReq.RequestID))

				finalCost = estimated
				finalCost.Model = billingModel
			}

			tunnelReq.RequestMetadata = withXAIUsageMetadata(tunnelReq.RequestMetadata, tunnelReq.Deployment.Provider, finalUsage)

			if debitErr := h.pipeline.Settle(requestCtx, SettlementRequest{Reservation: budgetReservation, TenantID: authInfo.TenantID, Actual: finalCost, RequestID: tunnelReq.RequestID, BillingMode: deploymentBillingMode(tunnelReq.Deployment)}); debitErr != nil {
				h.logger.Warn("failed to commit budget for streaming request", zap.Error(debitErr), zap.String("request_id", tunnelReq.RequestID))
			} else {
				budgetCommitted = true
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
				ProviderName:       tunnelReq.Deployment.Provider,
				ProviderModelName:  tunnelReq.Deployment.Model,
				RequestType:        "responses",
				UserID:             authInfo.UserID,
				VirtualAccountID:   authInfo.KeyID,
				VirtualAccountName: authInfo.KeyAlias,
				TeamIDs:            nonEmptySlice([]string{authInfo.TeamID}),
				StatusCode:         fiber.StatusOK,
				LatencyMS:          int(time.Since(start).Milliseconds()),
				InputTokens:        finalCost.InputTokens,
				OutputTokens:       finalCost.OutputTokens,
				CostUSD:            finalCost.TotalCostUSD,
				Metadata: enrichMetricMetadataForDeployment(map[string]interface{}{
					"endpoint":    "/v1/responses",
					"key_type":    authInfo.KeyType,
					"team_id":     authInfo.TeamID,
					"org_unit_id": authInfo.OrgUnitID,
					"tenant_slug": tenantSlug,
				}, tunnelReq.Deployment),
			})
		}

		relayTunnelStreamWithErrorFramer(
			stream,
			w,
			finalizeAndLogSuccess,
			h.logger,
			tunnelReq.RequestID,
			"responses",
			frameOpenAIResponseStreamError,
			h.newResponseStreamObserver(
				requestCtx,
				authInfo.TenantID,
				tunnelReq.Deployment,
				func(response contractsopenai.Response) {
					if isTerminalResponseStatus(response.Status) {
						responseCopy := response
						terminalResponse = &responseCopy
					}
				},
			),
		)
	})

	responseStart := time.Now()
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "response.write", "internal", responseStart, fiber.StatusOK, map[string]interface{}{"stream": true})

	return nil
}

func isTerminalResponseStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "completed", "incomplete", "failed":
		return true
	default:
		return false
	}
}

func (h *ResponsesHandler) emitModelMetric(input metricsobs.BuildInput) {
	if h.emitMetric == nil {
		return
	}

	h.emitMetric(input)
}

func supportsResponsesCreateProvider(provider string) bool {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "openai", "xai", "deepseek", "mock":
		return true
	default:
		return false
	}
}

func supportsResponseResourcesProvider(provider string) bool {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "openai", "xai":
		return true
	default:
		return false
	}
}

func responseRequestToChatCompletion(req contractsopenai.ResponseRequest, model string) contractsopenai.ChatCompletionRequest {
	content := flattenResponseInput(req.Input)
	if content == "" {
		content = strings.TrimSpace(req.Instructions)
	}

	messages := []contractsopenai.ChatMessage{}
	if strings.TrimSpace(req.Instructions) != "" {
		messages = append(messages, contractsopenai.ChatMessage{Role: "system", Content: strings.TrimSpace(req.Instructions)})
	}

	if strings.TrimSpace(content) != "" {
		messages = append(messages, contractsopenai.ChatMessage{Role: "user", Content: content})
	}

	if len(messages) == 0 {
		messages = append(messages, contractsopenai.ChatMessage{Role: "user", Content: " "})
	}

	return contractsopenai.ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		Stream:      req.Stream,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		MaxTokens:   req.MaxOutputTokens,
		Tools:       req.Tools,
		ToolChoice:  req.ToolChoice,
		User:        req.User,
		ServiceTier: req.ServiceTier,
	}
}

func flattenResponseInput(input interface{}) string {
	switch v := input.(type) {
	case string:
		return strings.TrimSpace(v)
	case []interface{}:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			if text := flattenResponseInput(item); text != "" {
				parts = append(parts, text)
			}
		}

		return strings.TrimSpace(strings.Join(parts, "\n"))
	case map[string]interface{}:
		if text, ok := v["text"].(string); ok && strings.TrimSpace(text) != "" {
			return strings.TrimSpace(text)
		}

		if content, ok := v["content"]; ok {
			return flattenResponseInput(content)
		}

		if inputText, ok := v["input_text"].(string); ok && strings.TrimSpace(inputText) != "" {
			return strings.TrimSpace(inputText)
		}

		return ""
	default:
		return ""
	}
}

func responseUsageToOpenAI(usage *contractsopenai.ResponseUsage) *contractsopenai.Usage {
	if usage == nil {
		return nil
	}

	result := &contractsopenai.Usage{
		PromptTokens:           usage.InputTokens,
		CompletionTokens:       usage.OutputTokens,
		TotalTokens:            usage.TotalTokens,
		CostInUSDTicks:         usage.CostInUSDTicks,
		NumServerSideToolsUsed: usage.NumServerSideToolsUsed,
	}
	if usage.InputDetails != nil {
		cachedTokens := usage.InputDetails.CachedTokens

		if cachedTokens < 0 {
			cachedTokens = 0
		}

		if cachedTokens > usage.InputTokens {
			cachedTokens = usage.InputTokens
		}

		result.PromptTokensDetails = &contractsopenai.UsageTokenDetails{
			CachedTokens:     cachedTokens,
			CacheWriteTokens: usage.InputDetails.CacheWriteTokens,
		}
		result.PromptCacheHitTokens = cachedTokens
		result.PromptCacheMissTokens = usage.InputTokens - cachedTokens
	}

	if usage.OutputDetails != nil {
		result.CompletionTokensDetails = &contractsopenai.UsageTokenDetails{
			ReasoningTokens: usage.OutputDetails.ReasoningTokens,
		}
	}

	return result
}

func responseUsageWithToolsToOpenAI(
	response contractsopenai.Response,
) *contractsopenai.Usage {
	usage := responseUsageToOpenAI(response.Usage)
	if usage == nil {
		return nil
	}

	serverTools := map[string]int{}

	for _, item := range response.Output {
		switch strings.ToLower(strings.TrimSpace(item.Type)) {
		case "web_search_call":
			serverTools["web_search_requests"]++
		case "x_search_call":
			serverTools["x_search_requests"]++
		case "code_execution_call", "code_interpreter_call":
			serverTools["code_execution_requests"]++
		case "attachment_search_call":
			serverTools["attachment_search_requests"]++
		case "collections_search_call":
			serverTools["collections_search_requests"]++
		case "file_search_call":
			serverTools["file_search_requests"]++
		}
	}

	if len(serverTools) > 0 {
		usage.ServerToolUse = serverTools
	}

	return usage
}
