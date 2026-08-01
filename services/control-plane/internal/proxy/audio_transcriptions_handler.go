// Package proxy hosts OpenAI-compatible request handlers for the control plane.
package proxy

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"mime/multipart"
	"path/filepath"
	"strconv"
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

const (
	audioTranscriptionsPath                     = "/audio/transcriptions"
	mistralAudioTranscriptionReservationSeconds = 60 * 60
)

// AudioTranscriptionsHandler serves OpenAI-compatible speech-to-text requests.
type AudioTranscriptionsHandler struct {
	llmRouter  *router.Router
	tunnel     platform.TunnelTransport
	billing    budgeter
	wallets    walletStore
	logs       usageLogger
	logger     *zap.Logger
	emitMetric func(input metricsobs.BuildInput)
}

type audioTranscriptionMetadata struct {
	Model          string
	Stream         bool
	ResponseFormat string
	Metadata       map[string]interface{}
}

// NewAudioTranscriptionsHandler creates a new audio transcriptions handler.
func NewAudioTranscriptionsHandler(llmRouter *router.Router, tunnel platform.TunnelTransport, billing budgeter, wallets walletStore, logs usageLogger, logger *zap.Logger, emitMetric ...func(input metricsobs.BuildInput)) *AudioTranscriptionsHandler {
	var metricEmitter func(input metricsobs.BuildInput)
	if len(emitMetric) > 0 {
		metricEmitter = emitMetric[0]
	}

	return NewAudioTranscriptionsHandlerWithDependencies(BaseHandlerDependencies{
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

// NewAudioTranscriptionsHandlerWithDependencies creates an audio transcriptions handler from explicit dependencies.
func NewAudioTranscriptionsHandlerWithDependencies(deps BaseHandlerDependencies) *AudioTranscriptionsHandler {
	return &AudioTranscriptionsHandler{
		llmRouter:  deps.Router,
		tunnel:     deps.Tunnel,
		billing:    deps.Billing,
		wallets:    deps.Wallets,
		logs:       deps.logger(),
		logger:     deps.Logger,
		emitMetric: deps.metricEmitter(),
	}
}

// Handle processes an OpenAI-compatible audio transcription request.
func (h *AudioTranscriptionsHandler) Handle(c *fiber.Ctx) error {
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

	contentType := strings.TrimSpace(c.Get(fiber.HeaderContentType))
	body := append([]byte(nil), c.Body()...)

	reqMeta, err := extractAudioTranscriptionMetadata(c, contentType)
	if err != nil {
		param, code := audioTranscriptionValidationErrorDetail(err)
		return writeAPIErrorDetail(c, fiber.StatusBadRequest, err.Error(), "invalid_request_error", param, code)
	}

	requestedModel := normalizeModelIdentifier(reqMeta.Model)

	routeCandidates := modelRoutingCandidates(requestedModel)
	if len(routeCandidates) == 0 {
		return writeAPIErrorDetail(c, fiber.StatusBadRequest, "model is required", "invalid_request_error", "model", "invalid_model")
	}

	tracing.SetModelForTrace(c, requestedModel)

	modelAccess := modelAccessForType(authInfo, requestedModel, platform.ModelTypeAudioTranscription)
	if !modelAccess.Allowed {
		if modelAccess.UnsupportedVMMode {
			return writeUnsupportedModelModeError(c, requestedModel, platform.ModelTypeAudioTranscription, modelAccess.ExposedModes)
		}

		denyReason := denyReasonForModel(authInfo)
		emitDenyCounter(h.logger, denyReason, requestedModel, currentTenant.ID)

		const message = "Model access denied or unsupported for audio transcriptions"
		logPreProviderRejection(h.logger, h.logs, c, authInfo, requestedModel, "/v1"+audioTranscriptionsPath, "audio_transcriptions", fiber.StatusForbidden, "authorization_error", message, denyReason, startTime)

		return writeAPIError(c, fiber.StatusForbidden, message, "authorization_error")
	}

	authStart := time.Now()
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "auth.validate", "internal", authStart, fiber.StatusOK, map[string]interface{}{"model": requestedModel})

	estimated := estimateAudioTranscriptionCostFallback(requestedModel, reqMeta.Metadata)
	estimated.Metadata = reqMeta.Metadata

	isMistralTranscriptionModel := strings.HasPrefix(strings.ToLower(providerModelID(requestedModel)), "voxtral-mini-")
	if isMistralTranscriptionModel {
		estimated, err = priceMeteredUsage(
			c.Context(),
			h.billing,
			billingEstimateModel(requestedModel),
			"audio_input_seconds",
			strconv.Itoa(mistralAudioTranscriptionReservationSeconds),
			estimated,
		)
		if err != nil {
			return writeBillingEstimateError(c, err)
		}
	}

	quotaStart := time.Now()

	rateLimit, err := h.billing.CheckRateLimit(c.Context(), authInfo, requestedModel, estimated.TotalTokens)
	setRateLimitHeaders(c, rateLimit)

	if err != nil {
		recordRateLimitRejection(c, h.logger, h.logs, h.emitMetric, authInfo, requestedModel, "/v1"+audioTranscriptionsPath, "audio_transcriptions", startTime, quotaStart, rateLimit, err)
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

	if healthErr := h.tunnel.Healthy(c.Context()); healthErr != nil {
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

	providerSupport := audioTranscriptionsProviderSupport(dep.Provider)
	if !providerSupport.Supported {
		return writeAPIError(c, fiber.StatusNotImplemented, providerSupport.Message, "invalid_request_error")
	}

	if isMistralProvider(dep.Provider) && !mistralEndpointEnabled("audio_transcriptions") {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Mistral AI audio transcriptions are disabled", "provider_disabled")
	}

	if reqMeta.Stream && !providerSupport.Streaming {
		return writeAPIError(c, fiber.StatusBadRequest, fmt.Sprintf("Streaming audio transcriptions are not supported for provider: %s", dep.Provider), "invalid_request_error")
	}

	if isMistralProvider(dep.Provider) {
		if !isMistralTranscriptionModel {
			return writeAPIError(c, fiber.StatusBadRequest, "selected Mistral model does not support audio transcription", "invalid_request_error")
		}

		var endpointErr error
		if isJSONContentType(contentType) {
			endpointErr = validateMistralJSONParameters(estimated.ModelRuntimeProfile, "audio_transcriptions", body)
		} else {
			endpointErr = validateMistralMultipartParameters(estimated.ModelRuntimeProfile, "audio_transcriptions", c)
		}

		if endpointErr != nil {
			return writeAPIError(c, fiber.StatusBadRequest, endpointErr.Error(), "invalid_request_error")
		}
	}

	estimated, err = applyDeploymentPricingOverride(c.Context(), h.billing, estimated, dep)
	if err != nil {
		return writeBillingEstimateError(c, err)
	}

	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "router.select_model", "internal", routeStart, fiber.StatusOK, map[string]interface{}{"provider": dep.Provider, "deployment_id": dep.ID})

	outboundModel := providerOutboundModel(requestedModel, dep)

	var outboundBody []byte

	outboundContentType := contentType
	if isJSONContentType(contentType) {
		outboundBody, err = rewriteOpenAIJSONModel(body, outboundModel)
	} else {
		outboundBody, outboundContentType, err = rewriteMultipartModel(c, contentType, outboundModel)
	}

	if err != nil {
		return writeAPIError(c, fiber.StatusBadRequest, "Invalid audio transcription request body", "invalid_request_error")
	}

	requestMetadata := map[string]interface{}{
		"tenant_slug":               apiheaders.GetTenant(c),
		"model":                     requestedModel,
		"provider_endpoint_path":    audioTranscriptionsPath,
		"provider_adapter_kind":     providerSupport.Kind,
		"operation":                 "transcriptions",
		"usage_reconciliation_mode": "audio_usage_reconciliation",
	}
	for k, v := range reqMeta.Metadata {
		requestMetadata[k] = v
	}

	tunnelReq := platform.TunnelRequest{
		RequestID:  requestID,
		TenantID:   currentTenant.ID,
		TraceID:    requestID,
		Stream:     reqMeta.Stream,
		Operation:  platform.ProviderOperationOpenAIAudioTranscriptions,
		Deployment: toContractDeployment(dep),
		Headers: map[string]string{
			"Content-Type": outboundContentType,
		},
		Body:            outboundBody,
		RequestMetadata: requestMetadata,
	}

	tunnelReq.ModelRuntimeProfiles, err = runtimeProfilesFromEstimate(dep.Provider, outboundModel, estimated)
	if err != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, err.Error(), "runtime_profile_unavailable")
	}

	tunnelReq.RequestMetadata = enrichRequestMetadataForDeployment(tunnelReq.RequestMetadata, tunnelReq.Deployment)

	if !budgetReservationAttempted {
		budgetReservation, budgetResponseHandled, err = reserveBudgetForRequest(c, h.billing, authInfo, estimated, requestID, deploymentBillingMode(tunnelReq.Deployment))
		if err != nil || budgetResponseHandled {
			return err
		}
	}

	if reqMeta.Stream {
		releaseBudgetOnExit = false
		return h.handleStream(c, requestedModel, authInfo, estimated, tunnelReq, budgetReservation, startTime, tenantSlug, body, contentType)
	}

	err = h.handleSync(c, requestedModel, authInfo, estimated, tunnelReq, budgetReservation, startTime, tenantSlug, body, contentType)
	if err == nil && c.Response().StatusCode() < httpErrorFloor {
		releaseBudgetOnExit = false
	}

	return err
}

func (h *AudioTranscriptionsHandler) handleSync(
	c *fiber.Ctx,
	requestedModel string,
	authInfo *platform.AuthContext,
	estimated platform.UsageCostBreakdown,
	tunnelReq platform.TunnelRequest,
	budgetReservation *platform.BudgetReservation,
	start time.Time,
	tenantSlug string,
	originalBody []byte,
	originalContentType string,
) error {
	providerStart := time.Now()
	providerAttrs := map[string]interface{}{
		"provider":        tunnelReq.Deployment.Provider,
		"request_payload": buildTraceAudioTranscriptionRequestPayload(originalBody, originalContentType, tunnelReq.RequestMetadata),
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

		return writeCompatibleProviderError(c, tunnelReq.Deployment.Provider, resp.StatusCode, resp.Body, resp.Headers)
	}

	responseSummary := buildTraceAudioTranscriptionResponsePayload(resp.Body, resp.Headers)
	providerAttrs["response_payload"] = responseSummary
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, resp.StatusCode, providerAttrs)

	responseMetadata := withAudioTranscriptionResponseMetadata(tunnelReq.RequestMetadata, resp.Body, resp.Headers)

	finalCost := estimated
	if !isMistralProvider(tunnelReq.Deployment.Provider) {
		finalCost, err = h.billing.FinalizeFromUsage(c.Context(), billingModelForDeployment(requestedModel, tunnelReq.Deployment), audioTranscriptionUsageToOpenAI(resp.Body), estimated)
		if err != nil {
			return writeAPIError(c, fiber.StatusBadGateway, "Failed to finalize audio transcription usage", "billing_error")
		}
	}

	if minutes, ok := audioTranscriptionMinutes(resp.Body); ok {
		dimension, quantity := audioTranscriptionMeteredQuantity(tunnelReq.Deployment.Provider, minutes)

		finalCost, err = priceMeteredUsage(
			c.Context(),
			h.billing,
			billingModelForDeployment(requestedModel, tunnelReq.Deployment),
			dimension,
			quantity,
			finalCost,
		)
		if err != nil {
			return writeAPIError(c, fiber.StatusBadGateway, "Failed to finalize audio transcription metered usage", "billing_error")
		}
	} else if isMistralProvider(tunnelReq.Deployment.Provider) {
		return writeAPIError(c, fiber.StatusBadGateway, "Mistral transcription response is missing prompt_audio_seconds usage", "billing_error")
	}

	if debitErr := commitBudgetOrDebit(c.Context(), h.logger, h.billing, h.wallets, budgetReservation, authInfo.TenantID, finalCost, tunnelReq.RequestID, deploymentBillingMode(tunnelReq.Deployment)); debitErr != nil {
		return writeBudgetCheckError(c, debitErr)
	}

	responseTimeMS := int(time.Since(start).Milliseconds())
	logEntry := buildRequestLogEntry(c, authInfo, tunnelReq.Deployment, requestedModel, "/v1"+audioTranscriptionsPath, "audio_transcriptions", "success", resp.StatusCode, start, responseTimeMS)
	logEntry.RequestID = tunnelReq.RequestID
	logEntry.TraceID = tunnelReq.TraceID
	logEntry.CostUSD = finalCost.TotalCostUSD
	logEntry.CostIRR = finalCost.TotalCostIRR
	logEntry.InputTokens = finalCost.InputTokens
	logEntry.OutputTokens = finalCost.OutputTokens
	logEntry.TotalTokens = finalCost.TotalTokens
	logEntry.RequestBody = buildTraceAudioTranscriptionRequestPayload(originalBody, originalContentType, tunnelReq.RequestMetadata)
	logEntry.ResponseBody = responseSummary
	logRequestBestEffort(h.logger, h.logs, c.Context(), logEntry)
	logSpendBestEffort(h.logger, h.logs, c.Context(), platform.SpendLog{
		RequestID:       tunnelReq.RequestID,
		TokenHash:       authInfo.TokenHash,
		TenantID:        authInfo.TenantID,
		OrganizationID:  authInfo.OrganizationID,
		TeamID:          authInfo.TeamID,
		Model:           requestedModel,
		InputTokens:     finalCost.InputTokens,
		OutputTokens:    finalCost.OutputTokens,
		TotalTokens:     finalCost.TotalTokens,
		CostUSD:         finalCost.TotalCostUSD,
		CostIRR:         finalCost.TotalCostIRR,
		ResponseTimeMS:  responseTimeMS,
		Status:          "success",
		RequestMetadata: enrichPricingMetadata(responseMetadata, authInfo, requestedModel),
		CompletionStart: start,
		CompletionEnd:   time.Now(),
	})
	h.emitModelMetric(metricsobs.BuildInput{
		TS:                 time.Now().UTC(),
		TenantID:           authInfo.TenantID,
		RequestID:          tunnelReq.RequestID,
		ModelName:          requestedModel,
		ProviderName:       tunnelReq.Deployment.Provider,
		ProviderModelName:  tunnelReq.Deployment.Model,
		RequestType:        "audio_transcriptions",
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
			"endpoint":        "/v1" + audioTranscriptionsPath,
			"key_type":        authInfo.KeyType,
			"team_id":         authInfo.TeamID,
			"org_unit_id":     authInfo.OrgUnitID,
			"tenant_slug":     tenantSlug,
			"response_format": tunnelReq.RequestMetadata["response_format"],
			"file_count":      tunnelReq.RequestMetadata["file_count"],
		}, tunnelReq.Deployment),
	})

	forwardProviderHeaders(c, resp.Headers)

	if contentType := responseContentType(resp.Headers); contentType != "" {
		c.Set(fiber.HeaderContentType, contentType)
	}

	setTraceHeaders(c, tunnelReq.TraceID)
	setCostHeaders(c, finalCost)

	responseStart := time.Now()
	err = c.Status(resp.StatusCode).Send(resp.Body)
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "response.write", "internal", responseStart, resp.StatusCode, map[string]interface{}{
		"response_payload": responseSummary,
	})

	return err
}

func (h *AudioTranscriptionsHandler) handleStream(
	c *fiber.Ctx,
	requestedModel string,
	authInfo *platform.AuthContext,
	estimated platform.UsageCostBreakdown,
	tunnelReq platform.TunnelRequest,
	budgetReservation *platform.BudgetReservation,
	start time.Time,
	tenantSlug string,
	originalBody []byte,
	originalContentType string,
) error {
	providerStart := time.Now()

	stream, err := h.tunnel.Stream(c.Context(), tunnelReq)
	if err != nil {
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, fiber.StatusBadGateway, map[string]interface{}{
			"provider":        tunnelReq.Deployment.Provider,
			"request_payload": buildTraceAudioTranscriptionRequestPayload(originalBody, originalContentType, tunnelReq.RequestMetadata),
			"stream":          true,
			"error":           err.Error(),
		})

		releaseBudgetBestEffort(c.Context(), h.logger, h.billing, budgetReservation, "stream_open_failed")

		return writeAPIError(c, fiber.StatusBadGateway, fmt.Sprintf("Tunnel error: %v", err), "api_error")
	}

	if statusCode := stream.StatusCode(); statusCode >= httpErrorFloor {
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, statusCode, map[string]interface{}{
			"provider":        tunnelReq.Deployment.Provider,
			"request_payload": buildTraceAudioTranscriptionRequestPayload(originalBody, originalContentType, tunnelReq.RequestMetadata),
			"stream":          true,
			"status_code":     statusCode,
		})
		releaseBudgetBestEffort(c.Context(), h.logger, h.billing, budgetReservation, "provider_http_error")

		return writeTunnelStreamHTTPError(c, h.logger, stream)
	}

	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, fiber.StatusOK, map[string]interface{}{
		"provider":        tunnelReq.Deployment.Provider,
		"request_payload": buildTraceAudioTranscriptionRequestPayload(originalBody, originalContentType, tunnelReq.RequestMetadata),
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

		streamUsage := &audioTranscriptionStreamUsage{}

		budgetCommitted := false
		defer func() {
			if !budgetCommitted {
				releaseBudgetBestEffort(requestCtx, h.logger, h.billing, budgetReservation, "stream_not_committed")
			}
		}()

		commitEstimatedBudget := func() {
			if budgetCommitted {
				return
			}

			finalCost := estimated

			var finalizeErr error
			if !isMistralProvider(tunnelReq.Deployment.Provider) {
				finalCost, finalizeErr = h.billing.FinalizeFromUsage(requestCtx, billingModelForDeployment(requestedModel, tunnelReq.Deployment), nil, estimated)
				if finalizeErr != nil {
					h.logger.Warn("failed to finalize streaming audio transcription usage", zap.Error(finalizeErr), zap.String("request_id", tunnelReq.RequestID))

					finalCost = estimated
					finalCost.Model = billingModelForDeployment(requestedModel, tunnelReq.Deployment)
				}
			}

			if minutes, ok := streamUsage.Minutes(); ok {
				dimension, quantity := audioTranscriptionMeteredQuantity(tunnelReq.Deployment.Provider, minutes)

				finalCost, finalizeErr = priceMeteredUsage(
					requestCtx,
					h.billing,
					billingModelForDeployment(requestedModel, tunnelReq.Deployment),
					dimension,
					quantity,
					finalCost,
				)
				if finalizeErr != nil {
					h.logger.Warn("failed to finalize streaming audio transcription metered usage", zap.Error(finalizeErr), zap.String("request_id", tunnelReq.RequestID))
				}
			}

			if debitErr := commitBudgetOrDebit(requestCtx, h.logger, h.billing, h.wallets, budgetReservation, authInfo.TenantID, finalCost, tunnelReq.RequestID, deploymentBillingMode(tunnelReq.Deployment)); debitErr != nil {
				h.logger.Warn("failed to commit budget for streaming audio transcription request", zap.Error(debitErr), zap.String("request_id", tunnelReq.RequestID))
				return
			}

			responseTimeMS := int(time.Since(start).Milliseconds())
			logSpendBestEffort(h.logger, h.logs, requestCtx, platform.SpendLog{
				RequestID:       tunnelReq.RequestID,
				TokenHash:       authInfo.TokenHash,
				TenantID:        authInfo.TenantID,
				OrganizationID:  authInfo.OrganizationID,
				TeamID:          authInfo.TeamID,
				Model:           requestedModel,
				InputTokens:     finalCost.InputTokens,
				OutputTokens:    finalCost.OutputTokens,
				TotalTokens:     finalCost.TotalTokens,
				CostUSD:         finalCost.TotalCostUSD,
				CostIRR:         finalCost.TotalCostIRR,
				ResponseTimeMS:  responseTimeMS,
				Status:          "success",
				RequestMetadata: enrichPricingMetadata(tunnelReq.RequestMetadata, authInfo, requestedModel),
				CompletionStart: start,
				CompletionEnd:   time.Now(),
			})
			h.emitModelMetric(metricsobs.BuildInput{
				TS:                 time.Now().UTC(),
				TenantID:           authInfo.TenantID,
				RequestID:          tunnelReq.RequestID,
				ModelName:          requestedModel,
				ProviderName:       tunnelReq.Deployment.Provider,
				ProviderModelName:  tunnelReq.Deployment.Model,
				RequestType:        "audio_transcriptions",
				UserID:             authInfo.UserID,
				VirtualAccountID:   authInfo.KeyID,
				VirtualAccountName: authInfo.KeyAlias,
				TeamIDs:            nonEmptySlice([]string{authInfo.TeamID}),
				StatusCode:         fiber.StatusOK,
				LatencyMS:          responseTimeMS,
				InputTokens:        finalCost.InputTokens,
				OutputTokens:       finalCost.OutputTokens,
				CostUSD:            finalCost.TotalCostUSD,
				Metadata: enrichMetricMetadataForDeployment(map[string]interface{}{
					"endpoint":    "/v1" + audioTranscriptionsPath,
					"key_type":    authInfo.KeyType,
					"team_id":     authInfo.TeamID,
					"org_unit_id": authInfo.OrgUnitID,
					"tenant_slug": tenantSlug,
					"stream":      true,
				}, tunnelReq.Deployment),
			})

			budgetCommitted = true
		}

		relayTunnelStream(
			stream,
			w,
			commitEstimatedBudget,
			h.logger,
			tunnelReq.RequestID,
			"audio_transcription",
			streamUsage.Observe,
		)
	})

	responseStart := time.Now()
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "response.write", "internal", responseStart, fiber.StatusOK, map[string]interface{}{"stream": true})

	return nil
}

func audioTranscriptionMeteredQuantity(provider, minutes string) (string, string) {
	if !isMistralProvider(provider) {
		return "audio_input_minutes", minutes
	}

	value, err := strconv.ParseFloat(strings.TrimSpace(minutes), 64)
	if err != nil || value < 0 {
		return "audio_input_seconds", "0"
	}

	return "audio_input_seconds", strconv.FormatFloat(value*60, 'f', -1, 64)
}

func (h *AudioTranscriptionsHandler) emitModelMetric(input metricsobs.BuildInput) {
	if h.emitMetric == nil {
		return
	}

	h.emitMetric(input)
}

func extractAudioTranscriptionMetadata(c *fiber.Ctx, contentType string) (audioTranscriptionMetadata, error) {
	metadata := audioTranscriptionMetadata{
		ResponseFormat: "json",
		Metadata: map[string]interface{}{
			"content_type": safeContentTypeFamily(contentType),
		},
	}

	if isJSONContentType(contentType) {
		return extractMistralJSONAudioTranscriptionMetadata(c.Body(), metadata)
	}

	if !isMultipartContentType(contentType) {
		return metadata, fmt.Errorf("multipart/form-data or application/json content type is required")
	}

	form, err := c.MultipartForm()
	if err != nil || form == nil {
		return metadata, fmt.Errorf("invalid multipart request body")
	}

	model := strings.TrimSpace(firstFormValue(form.Value, "model"))
	if model == "" {
		return metadata, fmt.Errorf("model is required")
	}

	metadata.Model = model

	fileHeaders := form.File["file"]
	modelID := strings.ToLower(providerModelID(model))
	isMistral := strings.HasPrefix(modelID, "voxtral-mini-")
	fileURL := strings.TrimSpace(firstFormValue(form.Value, "file_url"))
	fileID := strings.TrimSpace(firstFormValue(form.Value, "file_id"))

	sourceCount := 0
	if len(fileHeaders) > 0 {
		sourceCount++
	}

	if fileURL != "" {
		sourceCount++
	}

	if fileID != "" {
		sourceCount++
	}

	if (!isMistral && len(fileHeaders) == 0) || (isMistral && sourceCount != 1) {
		if isMistral {
			return metadata, fmt.Errorf("exactly one of file, file_url, or file_id is required")
		}

		return metadata, fmt.Errorf("file is required")
	}

	fillAudioTranscriptionMetadata(metadata.Metadata, form)

	if fileURL != "" {
		metadata.Metadata["file_source"] = "url"
		metadata.Metadata["file_url_hash"] = hashTracePreview(fileURL)
	}

	if fileID != "" {
		metadata.Metadata["file_source"] = "file_id"
		metadata.Metadata["file_id_hash"] = hashTracePreview(fileID)
	}

	if err := validateOpenAIAudioTranscriptionModelFields(model, form); err != nil {
		return metadata, err
	}

	responseFormat := strings.ToLower(strings.TrimSpace(firstFormValue(form.Value, "response_format")))
	if responseFormat == "" {
		responseFormat = "json"
	}

	if !validAudioTranscriptionResponseFormat(responseFormat) {
		return metadata, fmt.Errorf("response_format must be one of json, text, srt, verbose_json, vtt, or diarized_json")
	}

	metadata.ResponseFormat = responseFormat
	metadata.Metadata["response_format"] = responseFormat

	if temperature := strings.TrimSpace(firstFormValue(form.Value, "temperature")); temperature != "" {
		parsed, parseErr := strconv.ParseFloat(temperature, 64)
		if parseErr != nil || parsed < 0 || parsed > 1 {
			return metadata, fmt.Errorf("temperature must be between 0 and 1")
		}

		metadata.Metadata["temperature"] = parsed
	}

	if rawStream := strings.TrimSpace(firstFormValue(form.Value, "stream")); rawStream != "" {
		stream, parseErr := strconv.ParseBool(rawStream)
		if parseErr != nil {
			return metadata, fmt.Errorf("stream must be a boolean")
		}

		metadata.Stream = stream
		metadata.Metadata["stream"] = stream
	}

	granularities := audioFormArrayValues(form.Value, "timestamp_granularities")
	if len(granularities) > 0 {
		for _, value := range granularities {
			if value != "word" && value != "segment" {
				return metadata, fmt.Errorf("timestamp_granularities values must be word or segment")
			}
		}

		if !isMistral && responseFormat != "verbose_json" {
			return metadata, fmt.Errorf("timestamp_granularities requires response_format=verbose_json")
		}

		metadata.Metadata["timestamp_granularities"] = granularities
	}

	return metadata, nil
}

func extractMistralJSONAudioTranscriptionMetadata(body []byte, metadata audioTranscriptionMetadata) (audioTranscriptionMetadata, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil || raw == nil {
		return metadata, fmt.Errorf("invalid JSON request body")
	}

	model := strings.TrimSpace(stringFromMap(raw, "model"))
	if model == "" {
		return metadata, fmt.Errorf("model is required")
	}

	if !strings.HasPrefix(strings.ToLower(providerModelID(model)), "voxtral-mini-") {
		return metadata, fmt.Errorf("application/json content type is only supported for Mistral audio transcription models")
	}

	metadata.Model = model

	fileURL, fileURLExists := raw["file_url"]

	fileID, fileIDExists := raw["file_id"]
	if _, exists := raw["file"]; exists {
		return metadata, fmt.Errorf("file uploads require multipart/form-data")
	}

	urlValue, urlOK := fileURL.(string)
	idValue, idOK := fileID.(string)
	urlValue = strings.TrimSpace(urlValue)
	idValue = strings.TrimSpace(idValue)

	if (fileURLExists && fileURL != nil && !urlOK) || (fileIDExists && fileID != nil && !idOK) {
		return metadata, fmt.Errorf("file_url and file_id must be strings")
	}

	if (urlValue == "") == (idValue == "") {
		return metadata, fmt.Errorf("exactly one of file_url or file_id is required for a JSON request")
	}

	if urlValue != "" {
		metadata.Metadata["file_source"] = "url"
		metadata.Metadata["file_url_hash"] = hashTracePreview(urlValue)
	} else {
		metadata.Metadata["file_source"] = "file_id"
		metadata.Metadata["file_id_hash"] = hashTracePreview(idValue)
	}

	for _, key := range []string{"language"} {
		if value, exists := raw[key]; exists && value != nil {
			text, ok := value.(string)
			if !ok {
				return metadata, fmt.Errorf("%s must be a string", key)
			}

			if text = strings.TrimSpace(text); text != "" {
				metadata.Metadata[key] = text
			}
		}
	}

	for _, key := range []string{"diarize", "stream"} {
		if value, exists := raw[key]; exists && value != nil {
			flag, ok := value.(bool)
			if !ok {
				return metadata, fmt.Errorf("%s must be a boolean", key)
			}

			metadata.Metadata[key] = flag
			if key == "stream" {
				metadata.Stream = flag
			}
		}
	}

	if value, exists := raw["temperature"]; exists && value != nil {
		parsed, ok := value.(float64)
		if !ok || parsed < 0 || parsed > 1 {
			return metadata, fmt.Errorf("temperature must be between 0 and 1")
		}

		metadata.Metadata["temperature"] = parsed
	}

	for _, key := range []string{"context_bias", "timestamp_granularities"} {
		value, exists := raw[key]
		if !exists || value == nil {
			continue
		}

		items, ok := value.([]interface{})
		if !ok {
			return metadata, fmt.Errorf("%s must be an array of strings", key)
		}

		values := make([]string, 0, len(items))
		for _, item := range items {
			text, ok := item.(string)
			if !ok || strings.TrimSpace(text) == "" {
				return metadata, fmt.Errorf("%s must be an array of strings", key)
			}

			text = strings.TrimSpace(text)
			if key == "timestamp_granularities" && text != "word" && text != "segment" {
				return metadata, fmt.Errorf("timestamp_granularities values must be word or segment")
			}

			values = append(values, text)
		}

		metadata.Metadata[key] = values
	}

	metadata.Metadata["response_format"] = metadata.ResponseFormat

	return metadata, nil
}

func fillAudioTranscriptionMetadata(metadata map[string]interface{}, form *multipart.Form) {
	if language := strings.TrimSpace(firstFormValue(form.Value, "language")); language != "" {
		metadata["language"] = language
	}

	if prompt := firstFormValue(form.Value, "prompt"); prompt != "" {
		metadata["prompt_length"] = len(prompt)
		metadata["prompt_hash"] = hashTracePreview(prompt)
	}

	if languages := audioFormArrayValues(form.Value, "languages"); len(languages) > 0 {
		metadata["languages"] = languages
	}

	if keywords := audioFormArrayValues(form.Value, "keywords"); len(keywords) > 0 {
		metadata["keywords_count"] = len(keywords)
	}

	for _, key := range []string{"chunking_strategy"} {
		if value := strings.TrimSpace(firstFormValue(form.Value, key)); value != "" {
			metadata[key] = summarizeAudioConfigValue(value)
		}
	}

	include := audioFormArrayValues(form.Value, "include")
	if len(include) > 0 {
		metadata["include"] = include
	}

	knownNames := audioFormArrayValues(form.Value, "known_speaker_names")
	if len(knownNames) > 0 {
		metadata["known_speaker_names_count"] = len(knownNames)
	}

	knownRefs := audioFormArrayValues(form.Value, "known_speaker_references")
	if len(knownRefs) > 0 {
		metadata["known_speaker_references_count"] = len(knownRefs)
	}

	fileCount := 0
	for _, files := range form.File {
		fileCount += len(files)
	}

	metadata["file_count"] = fileCount
	metadata["audio_file_count"] = len(form.File["file"])

	if len(form.File["file"]) > 0 {
		header := form.File["file"][0]
		metadata["file_size_bytes"] = header.Size

		metadata["file_name_hash"] = hashTracePreview(header.Filename)
		if ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(header.Filename)), "."); ext != "" {
			metadata["file_extension"] = ext
		}

		if contentType := strings.TrimSpace(header.Header.Get("Content-Type")); contentType != "" {
			metadata["file_content_type"] = safeContentTypeFamily(contentType)
		}
	}
}

func validateOpenAIAudioTranscriptionModelFields(
	model string,
	form *multipart.Form,
) error {
	if form == nil {
		return nil
	}

	modelID := strings.ToLower(strings.TrimSpace(providerModelID(model)))
	if modelID == "gpt-transcribe" ||
		modelID == "gpt-4o-mini-transcribe" ||
		modelID == "gpt-4o-transcribe-diarize" {
		for _, header := range form.File["file"] {
			if header != nil && header.Size > 25*1024*1024 {
				return fmt.Errorf("file must not exceed 25 MB")
			}
		}
	}

	if modelID == "gpt-transcribe" {
		language := strings.TrimSpace(firstFormValue(form.Value, "language"))

		languages := audioFormArrayValues(form.Value, "languages")
		if language != "" && len(languages) > 0 {
			return fmt.Errorf("language and languages cannot be used together for gpt-transcribe")
		}

		for _, keyword := range audioFormArrayValues(form.Value, "keywords") {
			if strings.ContainsAny(keyword, "<>\r\n") {
				return fmt.Errorf("keywords values cannot contain <, >, carriage returns, or line feeds")
			}
		}
	}

	if modelID == "gpt-4o-transcribe-diarize" {
		if strings.TrimSpace(firstFormValue(form.Value, "prompt")) != "" {
			return fmt.Errorf("prompt is not supported by gpt-4o-transcribe-diarize")
		}

		names := audioFormArrayValues(form.Value, "known_speaker_names")

		references := audioFormArrayValues(form.Value, "known_speaker_references")
		if len(names) > 4 || len(references) > 4 {
			return fmt.Errorf("known speaker references support at most four speakers")
		}

		if len(names) != len(references) {
			return fmt.Errorf("known_speaker_names and known_speaker_references must have the same length")
		}
	}

	return nil
}

func firstFormValue(values map[string][]string, key string) string {
	if len(values) == 0 {
		return ""
	}

	for _, candidate := range []string{key, key + "[]"} {
		rawValues := values[candidate]
		if len(rawValues) > 0 {
			return rawValues[0]
		}
	}

	return ""
}

func audioFormArrayValues(values map[string][]string, key string) []string {
	out := []string{}

	for candidate, rawValues := range values {
		if candidate != key && candidate != key+"[]" && !strings.HasPrefix(candidate, key+"[") {
			continue
		}

		for _, raw := range rawValues {
			value := strings.TrimSpace(raw)
			if value != "" {
				out = append(out, value)
			}
		}
	}

	return out
}

func summarizeAudioConfigValue(value string) interface{} {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	if strings.HasPrefix(value, "{") {
		var raw map[string]interface{}
		if err := json.Unmarshal([]byte(value), &raw); err == nil {
			out := map[string]interface{}{}
			if t, ok := raw["type"].(string); ok {
				out["type"] = strings.TrimSpace(t)
			}

			for _, key := range []string{"prefix_padding_ms", "silence_duration_ms", "threshold"} {
				if v, ok := raw[key]; ok {
					out[key] = v
				}
			}

			if len(out) > 0 {
				return out
			}
		}
	}

	return value
}

func validAudioTranscriptionResponseFormat(format string) bool {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json", "text", "srt", "verbose_json", "vtt", "diarized_json":
		return true
	default:
		return false
	}
}

func audioTranscriptionValidationErrorDetail(err error) (string, string) {
	message := ""
	if err != nil {
		message = strings.ToLower(strings.TrimSpace(err.Error()))
	}

	switch {
	case strings.Contains(message, "content type"):
		return "content_type", "unsupported_content_type"
	case strings.Contains(message, "model"):
		return "model", "invalid_model"
	case strings.Contains(message, "file"):
		return "file", "invalid_file"
	case strings.Contains(message, "response_format"):
		return "response_format", "invalid_response_format"
	case strings.Contains(message, "timestamp_granularities"):
		return "timestamp_granularities", "invalid_timestamp_granularities"
	case strings.Contains(message, "temperature"):
		return "temperature", "invalid_temperature"
	case strings.Contains(message, "stream"):
		return "stream", "invalid_stream"
	case strings.Contains(message, "languages"):
		return "languages", "invalid_languages"
	case strings.Contains(message, "keywords"):
		return "keywords", "invalid_keywords"
	case strings.Contains(message, "known_speaker"):
		return "known_speaker_references", "invalid_known_speakers"
	case strings.Contains(message, "prompt"):
		return "prompt", "unsupported_parameter"
	default:
		return "request", "invalid_request"
	}
}

func estimateAudioTranscriptionCostFallback(model string, metadata map[string]interface{}) platform.UsageCostBreakdown {
	inputTokens := 0
	if promptLength, ok := metadata["prompt_length"]; ok {
		inputTokens += estimateTextTokensFromLength(promptLength)
	}

	if size, ok := metadata["file_size_bytes"]; ok {
		inputTokens += estimateAudioTokensFromFileSize(size)
	}

	return platform.UsageCostBreakdown{
		Model:       model,
		RequestType: "audio_transcriptions",
		Endpoint:    "/v1" + audioTranscriptionsPath,
		InputTokens: inputTokens,
		TotalTokens: inputTokens,
	}
}

func estimateAudioTokensFromFileSize(raw interface{}) int {
	var size int64

	switch v := raw.(type) {
	case int64:
		size = v
	case int:
		size = int64(v)
	case float64:
		size = int64(v)
	case json.Number:
		parsed, err := v.Int64()
		if err != nil {
			return 0
		}

		size = parsed
	default:
		return 0
	}

	if size <= 0 {
		return 0
	}

	return int(math.Ceil(float64(size) / 32000.0))
}

func audioTranscriptionUsageToOpenAI(body []byte) *contractsopenai.Usage {
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil
	}

	usage, ok := raw["usage"].(map[string]interface{})
	if !ok {
		return nil
	}

	inputTokens := intFromAudioUsage(usage, "input_tokens")
	if inputTokens == 0 {
		inputTokens = intFromAudioUsage(usage, "prompt_tokens")
	}

	outputTokens := intFromAudioUsage(usage, "output_tokens")
	if outputTokens == 0 {
		outputTokens = intFromAudioUsage(usage, "completion_tokens")
	}

	totalTokens := intFromAudioUsage(usage, "total_tokens")
	if totalTokens == 0 {
		totalTokens = inputTokens + outputTokens
	}

	if inputTokens == 0 && outputTokens == 0 && totalTokens == 0 {
		return nil
	}

	return &contractsopenai.Usage{
		PromptTokens:     inputTokens,
		CompletionTokens: outputTokens,
		TotalTokens:      totalTokens,
	}
}

func audioTranscriptionMinutes(body []byte) (string, bool) {
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return "", false
	}

	usage, ok := raw["usage"].(map[string]interface{})
	if !ok {
		return "", false
	}

	seconds, ok := floatFromAudioUsage(usage, "prompt_audio_seconds")
	if !ok {
		seconds, ok = floatFromAudioUsage(usage, "seconds")
	}

	if !ok || seconds < 0 {
		return "", false
	}

	return strconv.FormatFloat(seconds/60, 'f', -1, 64), true
}

func floatFromAudioUsage(usage map[string]interface{}, key string) (float64, bool) {
	switch value := usage[key].(type) {
	case float64:
		return value, true
	case int:
		return float64(value), true
	case int64:
		return float64(value), true
	case json.Number:
		parsed, err := value.Float64()
		return parsed, err == nil
	default:
		return 0, false
	}
}

const maxAudioTranscriptionStreamUsageBytes = 1 << 20

type audioTranscriptionStreamUsage struct {
	body []byte
}

func (usage *audioTranscriptionStreamUsage) Observe(chunk []byte) {
	if usage == nil || len(chunk) == 0 ||
		len(usage.body) >= maxAudioTranscriptionStreamUsageBytes {
		return
	}

	remaining := maxAudioTranscriptionStreamUsageBytes - len(usage.body)
	if len(chunk) > remaining {
		chunk = chunk[:remaining]
	}

	usage.body = append(usage.body, chunk...)
}

func (usage *audioTranscriptionStreamUsage) Minutes() (string, bool) {
	if usage == nil {
		return "", false
	}

	for _, line := range bytes.Split(usage.body, []byte{'\n'}) {
		line = bytes.TrimSpace(line)
		if !bytes.HasPrefix(line, []byte("data:")) {
			continue
		}

		payload := bytes.TrimSpace(bytes.TrimPrefix(line, []byte("data:")))
		if minutes, ok := audioTranscriptionMinutes(payload); ok {
			return minutes, true
		}
	}

	return "", false
}

func intFromAudioUsage(usage map[string]interface{}, key string) int {
	switch v := usage[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	case json.Number:
		n, err := v.Int64()
		if err != nil {
			return 0
		}

		return int(n)
	default:
		return 0
	}
}

func withAudioTranscriptionResponseMetadata(input map[string]interface{}, body []byte, headers map[string]string) map[string]interface{} {
	out := map[string]interface{}{}
	for k, v := range input {
		out[k] = v
	}

	for k, v := range audioTranscriptionUsageMetadata(body) {
		out[k] = v
	}

	if contentType := responseContentType(headers); contentType != "" {
		out["response_content_type"] = safeContentTypeFamily(contentType)
	}

	out["response_summary"] = buildTraceAudioTranscriptionResponsePayload(body, headers)

	return out
}

func audioTranscriptionUsageMetadata(body []byte) map[string]interface{} {
	out := map[string]interface{}{}

	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return out
	}

	usage, ok := raw["usage"].(map[string]interface{})
	if !ok {
		return out
	}

	if usageType, ok := usage["type"].(string); ok && strings.TrimSpace(usageType) != "" {
		out["usage_type"] = strings.TrimSpace(usageType)
	}

	for _, key := range []string{"input_tokens", "output_tokens", "total_tokens", "seconds"} {
		if v, ok := usage[key]; ok {
			out["usage_"+key] = v
		}
	}

	return out
}

func responseContentType(headers map[string]string) string {
	if len(headers) == 0 {
		return ""
	}

	for key, value := range headers {
		if strings.EqualFold(key, "content-type") {
			return strings.TrimSpace(value)
		}
	}

	return ""
}

type audioTranscriptionsProviderSupportInfo struct {
	Supported bool
	Streaming bool
	Message   string
	Kind      string
}

func audioTranscriptionsProviderSupport(provider string) audioTranscriptionsProviderSupportInfo {
	normalized := strings.ToLower(strings.TrimSpace(provider))
	info := audioTranscriptionsProviderSupportInfo{
		Message: fmt.Sprintf("Audio transcriptions API is not enabled for provider: %s", provider),
	}

	switch normalized {
	case "", "openai", "azure", "azure-open-ai", "groq", "mistral", "mistral-ai", "togetherai", "together-ai", "openrouter", "sambanova", "deepinfra", "fireworks", "mock":
		info.Supported = true
		info.Streaming = true
		info.Kind = "openai_compatible"
	case "deepgram", "cohere", "google-vertex", "vertex", "vertex-ai", "vertexai":
		info.Supported = true
		info.Kind = "native"
	case "aws-bedrock", "bedrock":
		info.Message = "Audio transcriptions API is not enabled for provider: aws-bedrock; Nova Sonic is a realtime speech-to-speech model, not uploaded-file transcription"
	}

	return info
}
