// Package proxy hosts OpenAI-compatible request handlers for the control plane.
package proxy

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

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
	audioSpeechPath        = "/audio/speech"
	audioSpeechRequestType = "audio_speech"
)

// AudioSpeechHandler serves OpenAI-compatible text-to-speech requests.
type AudioSpeechHandler struct {
	llmRouter  *router.Router
	tunnel     platform.TunnelTransport
	billing    budgeter
	wallets    walletStore
	logs       usageLogger
	logger     *zap.Logger
	emitMetric func(input metricsobs.BuildInput)
}

type audioSpeechMetadata struct {
	Model          string
	Stream         bool
	ResponseFormat string
	StreamFormat   string
	Metadata       map[string]interface{}
}

// NewAudioSpeechHandler creates a new audio speech handler.
func NewAudioSpeechHandler(llmRouter *router.Router, tunnel platform.TunnelTransport, billing budgeter, wallets walletStore, logs usageLogger, logger *zap.Logger, emitMetric ...func(input metricsobs.BuildInput)) *AudioSpeechHandler {
	var metricEmitter func(input metricsobs.BuildInput)
	if len(emitMetric) > 0 {
		metricEmitter = emitMetric[0]
	}

	return NewAudioSpeechHandlerWithDependencies(BaseHandlerDependencies{
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

// NewAudioSpeechHandlerWithDependencies creates an audio speech handler from explicit dependencies.
func NewAudioSpeechHandlerWithDependencies(deps BaseHandlerDependencies) *AudioSpeechHandler {
	return &AudioSpeechHandler{
		llmRouter:  deps.Router,
		tunnel:     deps.Tunnel,
		billing:    deps.Billing,
		wallets:    deps.Wallets,
		logs:       deps.logger(),
		logger:     deps.Logger,
		emitMetric: deps.metricEmitter(),
	}
}

// Handle processes an OpenAI-compatible text-to-speech request.
func (h *AudioSpeechHandler) Handle(c *fiber.Ctx) error {
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

	reqMeta, err := extractAudioSpeechMetadata(body, contentType)
	if err != nil {
		param, code := audioSpeechValidationErrorDetail(err)
		return writeAPIErrorDetail(c, fiber.StatusBadRequest, err.Error(), "invalid_request_error", param, code)
	}

	requestedModel := normalizeModelIdentifier(reqMeta.Model)

	routeCandidates := modelRoutingCandidates(requestedModel)
	if len(routeCandidates) == 0 {
		return writeAPIErrorDetail(c, fiber.StatusBadRequest, "model is required", "invalid_request_error", "model", "invalid_model")
	}

	tracing.SetModelForTrace(c, requestedModel)

	modelAccess := modelAccessForType(authInfo, requestedModel, platform.ModelTypeTextToSpeech)
	if !modelAccess.Allowed {
		if modelAccess.UnsupportedVMMode {
			return writeUnsupportedModelModeError(c, requestedModel, platform.ModelTypeTextToSpeech, modelAccess.ExposedModes)
		}

		denyReason := denyReasonForModel(authInfo)
		emitDenyCounter(h.logger, denyReason, requestedModel, currentTenant.ID)

		const message = "Model access denied or unsupported for text to speech"
		logPreProviderRejection(h.logger, h.logs, c, authInfo, requestedModel, "/v1"+audioSpeechPath, audioSpeechRequestType, fiber.StatusForbidden, "authorization_error", message, denyReason, startTime)

		return writeAPIError(c, fiber.StatusForbidden, message, "authorization_error")
	}

	authStart := time.Now()
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "auth.validate", "internal", authStart, fiber.StatusOK, map[string]interface{}{"model": requestedModel})

	estimated := estimateAudioSpeechCostFallback(requestedModel, reqMeta.Metadata)
	estimated.Metadata = reqMeta.Metadata
	estimated, err = priceMeteredUsage(
		c.Context(),
		h.billing,
		billingEstimateModel(requestedModel),
		"speech_input_characters",
		strconv.Itoa(audioSpeechInputCharacters(reqMeta.Metadata)),
		estimated,
	)
	if err != nil {
		return writeBillingEstimateError(c, err)
	}
	quotaStart := time.Now()

	rateLimit, err := h.billing.CheckRateLimit(c.Context(), authInfo, requestedModel, estimated.TotalTokens)
	setRateLimitHeaders(c, rateLimit)

	if err != nil {
		recordRateLimitRejection(c, h.logger, h.logs, h.emitMetric, authInfo, requestedModel, "/v1"+audioSpeechPath, audioSpeechRequestType, startTime, quotaStart, rateLimit, err)
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

	if !supportsAudioSpeechProvider(dep.Provider) {
		return writeAPIError(c, fiber.StatusNotImplemented, fmt.Sprintf("Text to speech API is not enabled for provider: %s", dep.Provider), "invalid_request_error")
	}

	estimated, err = applyDeploymentPricingOverride(c.Context(), h.billing, estimated, dep)
	if err != nil {
		return writeBillingEstimateError(c, err)
	}
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "router.select_model", "internal", routeStart, fiber.StatusOK, map[string]interface{}{"provider": dep.Provider, "deployment_id": dep.ID})

	outboundModel := providerOutboundModel(requestedModel, dep)

	outboundBody, err := rewriteOpenAIJSONModel(body, outboundModel)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadRequest, "Invalid JSON request body", "invalid_request_error")
	}

	requestMetadata := map[string]interface{}{
		"tenant_slug":               apiheaders.GetTenant(c),
		"model":                     requestedModel,
		"provider_endpoint_path":    audioSpeechPath,
		"operation":                 "speech",
		"usage_reconciliation_mode": "audio_speech_estimated",
	}
	for k, v := range reqMeta.Metadata {
		requestMetadata[k] = v
	}

	tunnelReq := platform.TunnelRequest{
		RequestID:  requestID,
		TenantID:   currentTenant.ID,
		TraceID:    requestID,
		Stream:     reqMeta.Stream,
		Operation:  platform.ProviderOperationOpenAIAudioSpeech,
		Deployment: toContractDeployment(dep),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body:            outboundBody,
		RequestMetadata: requestMetadata,
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

func (h *AudioSpeechHandler) handleSync(
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
	if rawTunnel, ok := h.tunnel.(platform.TunnelRawTransport); ok {
		return h.handleRawSync(c, rawTunnel, requestedModel, authInfo, estimated, tunnelReq, budgetReservation, start, tenantSlug, originalBody, originalContentType)
	}

	providerStart := time.Now()
	providerAttrs := map[string]interface{}{
		"provider":        tunnelReq.Deployment.Provider,
		"request_payload": buildTraceAudioSpeechRequestPayload(originalBody, originalContentType, tunnelReq.RequestMetadata),
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

	responseSummary := buildTraceAudioSpeechResponsePayload(resp.Body, resp.Headers)
	providerAttrs["response_payload"] = responseSummary
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, resp.StatusCode, providerAttrs)

	responseMetadata := withAudioSpeechResponseMetadata(tunnelReq.RequestMetadata, resp.Body, resp.Headers)

	finalCost, err := h.billing.FinalizeFromUsage(c.Context(), billingModelForDeployment(requestedModel, tunnelReq.Deployment), audioSpeechUsageToOpenAI(resp.Body), estimated)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, "Failed to finalize text to speech usage", "billing_error")
	}
	finalCost, err = priceMeteredUsage(
		c.Context(),
		h.billing,
		billingModelForDeployment(requestedModel, tunnelReq.Deployment),
		"speech_input_characters",
		strconv.Itoa(audioSpeechInputCharacters(tunnelReq.RequestMetadata)),
		finalCost,
	)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, "Failed to finalize text to speech metered usage", "billing_error")
	}

	if debitErr := commitBudgetOrDebit(c.Context(), h.logger, h.billing, h.wallets, budgetReservation, authInfo.TenantID, finalCost, tunnelReq.RequestID, deploymentBillingMode(tunnelReq.Deployment)); debitErr != nil {
		return writeBudgetCheckError(c, debitErr)
	}

	responseTimeMS := int(time.Since(start).Milliseconds())
	logEntry := buildRequestLogEntry(c, authInfo, tunnelReq.Deployment, requestedModel, "/v1"+audioSpeechPath, audioSpeechRequestType, "success", resp.StatusCode, start, responseTimeMS)
	logEntry.RequestID = tunnelReq.RequestID
	logEntry.TraceID = tunnelReq.TraceID
	logEntry.CostUSD = finalCost.TotalCostUSD
	logEntry.CostIRR = finalCost.TotalCostIRR
	logEntry.InputTokens = finalCost.InputTokens
	logEntry.OutputTokens = finalCost.OutputTokens
	logEntry.TotalTokens = finalCost.TotalTokens
	logEntry.RequestBody = buildTraceAudioSpeechRequestPayload(originalBody, originalContentType, tunnelReq.RequestMetadata)
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
		RequestType:        audioSpeechRequestType,
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
			"endpoint":        "/v1" + audioSpeechPath,
			"key_type":        authInfo.KeyType,
			"team_id":         authInfo.TeamID,
			"org_unit_id":     authInfo.OrgUnitID,
			"tenant_slug":     tenantSlug,
			"response_format": tunnelReq.RequestMetadata["response_format"],
			"voice_type":      tunnelReq.RequestMetadata["voice_type"],
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

func (h *AudioSpeechHandler) handleRawSync(
	c *fiber.Ctx,
	rawTunnel platform.TunnelRawTransport,
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
		"request_payload": buildTraceAudioSpeechRequestPayload(originalBody, originalContentType, tunnelReq.RequestMetadata),
		"raw_stream":      true,
	}
	resp, err := rawTunnel.ProxyRaw(c.Context(), tunnelReq, bytes.NewReader(tunnelReq.Body), int64(len(tunnelReq.Body)))
	if err != nil {
		providerAttrs["error"] = err.Error()
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, fiber.StatusBadGateway, providerAttrs)
		return writeAPIError(c, fiber.StatusBadGateway, fmt.Sprintf("Tunnel error: %v", err), "api_error")
	}
	if resp == nil || resp.Body == nil {
		return writeAPIError(c, fiber.StatusBadGateway, "Provider returned an empty speech response", "api_error")
	}

	if resp.StatusCode >= httpErrorFloor {
		defer resp.Body.Close()
		errorBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
		providerAttrs["response_payload"] = buildTraceProviderErrorPayload(resp.StatusCode, errorBody)
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, resp.StatusCode, providerAttrs)
		return writeCompatibleProviderError(c, tunnelReq.Deployment.Provider, resp.StatusCode, errorBody, resp.Headers)
	}

	responseSummary := buildTraceAudioSpeechResponsePayload(nil, resp.Headers)
	providerAttrs["response_payload"] = responseSummary
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, resp.StatusCode, providerAttrs)

	finalCost, err := priceMeteredUsage(
		c.Context(),
		h.billing,
		billingModelForDeployment(requestedModel, tunnelReq.Deployment),
		"speech_input_characters",
		strconv.Itoa(audioSpeechInputCharacters(tunnelReq.RequestMetadata)),
		estimated,
	)
	if err != nil {
		_ = resp.Body.Close()
		return writeAPIError(c, fiber.StatusBadGateway, "Failed to finalize text to speech metered usage", "billing_error")
	}
	if debitErr := commitBudgetOrDebit(c.Context(), h.logger, h.billing, h.wallets, budgetReservation, authInfo.TenantID, finalCost, tunnelReq.RequestID, deploymentBillingMode(tunnelReq.Deployment)); debitErr != nil {
		_ = resp.Body.Close()
		return writeBudgetCheckError(c, debitErr)
	}

	responseTimeMS := int(time.Since(start).Milliseconds())
	responseMetadata := withAudioSpeechResponseMetadata(tunnelReq.RequestMetadata, nil, resp.Headers)
	logEntry := buildRequestLogEntry(c, authInfo, tunnelReq.Deployment, requestedModel, "/v1"+audioSpeechPath, audioSpeechRequestType, "success", resp.StatusCode, start, responseTimeMS)
	logEntry.RequestID = tunnelReq.RequestID
	logEntry.TraceID = tunnelReq.TraceID
	logEntry.CostUSD = finalCost.TotalCostUSD
	logEntry.CostIRR = finalCost.TotalCostIRR
	logEntry.InputTokens = finalCost.InputTokens
	logEntry.OutputTokens = finalCost.OutputTokens
	logEntry.TotalTokens = finalCost.TotalTokens
	logEntry.RequestBody = buildTraceAudioSpeechRequestPayload(originalBody, originalContentType, tunnelReq.RequestMetadata)
	logEntry.ResponseBody = responseSummary
	logRequestBestEffort(h.logger, h.logs, c.Context(), logEntry)
	logSpendBestEffort(h.logger, h.logs, c.Context(), platform.SpendLog{
		RequestID: tunnelReq.RequestID, TokenHash: authInfo.TokenHash,
		TenantID: authInfo.TenantID, OrganizationID: authInfo.OrganizationID, TeamID: authInfo.TeamID,
		Model: requestedModel, InputTokens: finalCost.InputTokens, OutputTokens: finalCost.OutputTokens,
		TotalTokens: finalCost.TotalTokens, CostUSD: finalCost.TotalCostUSD, CostIRR: finalCost.TotalCostIRR,
		ResponseTimeMS: responseTimeMS, Status: "success",
		RequestMetadata: enrichPricingMetadata(responseMetadata, authInfo, requestedModel),
		CompletionStart: start, CompletionEnd: time.Now(),
	})
	h.emitModelMetric(metricsobs.BuildInput{
		TS: time.Now().UTC(), TenantID: authInfo.TenantID, RequestID: tunnelReq.RequestID,
		ModelName: requestedModel, ProviderName: tunnelReq.Deployment.Provider, ProviderModelName: tunnelReq.Deployment.Model,
		RequestType: audioSpeechRequestType, UserID: authInfo.UserID, VirtualAccountID: authInfo.KeyID,
		VirtualAccountName: authInfo.KeyAlias, TeamIDs: nonEmptySlice([]string{authInfo.TeamID}),
		StatusCode: resp.StatusCode, LatencyMS: responseTimeMS, InputTokens: finalCost.InputTokens,
		OutputTokens: finalCost.OutputTokens, CostUSD: finalCost.TotalCostUSD,
		Metadata: enrichMetricMetadataForDeployment(map[string]interface{}{
			"endpoint": "/v1" + audioSpeechPath, "key_type": authInfo.KeyType,
			"team_id": authInfo.TeamID, "org_unit_id": authInfo.OrgUnitID, "tenant_slug": tenantSlug,
			"response_format": tunnelReq.RequestMetadata["response_format"], "voice_type": tunnelReq.RequestMetadata["voice_type"],
			"raw_stream": true,
		}, tunnelReq.Deployment),
	})

	forwardProviderHeaders(c, resp.Headers)
	if contentType := responseContentType(resp.Headers); contentType != "" {
		c.Set(fiber.HeaderContentType, contentType)
	}
	setTraceHeaders(c, tunnelReq.TraceID)
	setCostHeaders(c, finalCost)
	c.Status(resp.StatusCode)
	requestCtx := c.Context()
	requestCtx.SetBodyStreamWriter(func(writer *bufio.Writer) {
		defer resp.Body.Close()
		_, _ = io.CopyBuffer(writer, resp.Body, make([]byte, 32*1024))
		_ = writer.Flush()
	})

	return nil
}

func (h *AudioSpeechHandler) handleStream(
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
			"request_payload": buildTraceAudioSpeechRequestPayload(originalBody, originalContentType, tunnelReq.RequestMetadata),
			"stream":          true,
			"error":           err.Error(),
		})

		releaseBudgetBestEffort(c.Context(), h.logger, h.billing, budgetReservation, "stream_open_failed")

		return writeAPIError(c, fiber.StatusBadGateway, fmt.Sprintf("Tunnel error: %v", err), "api_error")
	}

	if statusCode := stream.StatusCode(); statusCode >= httpErrorFloor {
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, statusCode, map[string]interface{}{
			"provider":        tunnelReq.Deployment.Provider,
			"request_payload": buildTraceAudioSpeechRequestPayload(originalBody, originalContentType, tunnelReq.RequestMetadata),
			"stream":          true,
			"status_code":     statusCode,
		})
		releaseBudgetBestEffort(c.Context(), h.logger, h.billing, budgetReservation, "provider_http_error")

		return writeTunnelStreamHTTPError(c, h.logger, stream)
	}

	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, fiber.StatusOK, map[string]interface{}{
		"provider":        tunnelReq.Deployment.Provider,
		"request_payload": buildTraceAudioSpeechRequestPayload(originalBody, originalContentType, tunnelReq.RequestMetadata),
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
		defer func() {
			if !budgetCommitted {
				releaseBudgetBestEffort(requestCtx, h.logger, h.billing, budgetReservation, "stream_not_committed")
			}
		}()

		commitEstimatedBudget := func() {
			if budgetCommitted {
				return
			}

			finalCost, finalizeErr := h.billing.FinalizeFromUsage(requestCtx, billingModelForDeployment(requestedModel, tunnelReq.Deployment), nil, estimated)
			if finalizeErr != nil {
				h.logger.Warn("failed to finalize streaming text to speech usage", zap.Error(finalizeErr), zap.String("request_id", tunnelReq.RequestID))

				finalCost = estimated
				finalCost.Model = billingModelForDeployment(requestedModel, tunnelReq.Deployment)
			}

			if debitErr := commitBudgetOrDebit(requestCtx, h.logger, h.billing, h.wallets, budgetReservation, authInfo.TenantID, finalCost, tunnelReq.RequestID, deploymentBillingMode(tunnelReq.Deployment)); debitErr != nil {
				h.logger.Warn("failed to commit budget for streaming text to speech request", zap.Error(debitErr), zap.String("request_id", tunnelReq.RequestID))
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
				RequestType:        audioSpeechRequestType,
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
					"endpoint":      "/v1" + audioSpeechPath,
					"key_type":      authInfo.KeyType,
					"team_id":       authInfo.TeamID,
					"org_unit_id":   authInfo.OrgUnitID,
					"tenant_slug":   tenantSlug,
					"stream_format": "sse",
				}, tunnelReq.Deployment),
			})

			budgetCommitted = true
		}

		relayTunnelStream(stream, w, commitEstimatedBudget, h.logger, tunnelReq.RequestID, "text_to_speech")
	})

	responseStart := time.Now()
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "response.write", "internal", responseStart, fiber.StatusOK, map[string]interface{}{"stream": true})

	return nil
}

func (h *AudioSpeechHandler) emitModelMetric(input metricsobs.BuildInput) {
	if h.emitMetric == nil {
		return
	}

	h.emitMetric(input)
}

func extractAudioSpeechMetadata(body []byte, contentType string) (audioSpeechMetadata, error) {
	metadata := audioSpeechMetadata{
		ResponseFormat: "mp3",
		Metadata: map[string]interface{}{
			"content_type": safeContentTypeFamily(contentType),
		},
	}

	if !isJSONContentType(contentType) {
		return metadata, fmt.Errorf("application/json content type is required")
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil || raw == nil {
		return metadata, fmt.Errorf("invalid JSON request body")
	}

	model := strings.TrimSpace(stringFromMap(raw, "model"))
	if model == "" {
		return metadata, fmt.Errorf("model is required")
	}

	metadata.Model = model

	input, ok := raw["input"].(string)

	input = strings.TrimSpace(input)
	if !ok || input == "" {
		return metadata, fmt.Errorf("input is required")
	}

	inputLength := utf8.RuneCountInString(input)
	if inputLength > 4096 {
		return metadata, fmt.Errorf("input must be at most 4096 characters")
	}

	metadata.Metadata["input_length"] = inputLength
	metadata.Metadata["input_hash"] = hashTracePreview(input)

	if err := fillAudioSpeechVoiceMetadata(metadata.Metadata, raw["voice"]); err != nil {
		return metadata, err
	}

	if instructions, ok := raw["instructions"].(string); ok && strings.TrimSpace(instructions) != "" {
		instructionsLength := utf8.RuneCountInString(instructions)
		if instructionsLength > 4096 {
			return metadata, fmt.Errorf("instructions must be at most 4096 characters")
		}

		metadata.Metadata["instructions_length"] = instructionsLength
		metadata.Metadata["instructions_hash"] = hashTracePreview(instructions)
	}

	responseFormat := strings.ToLower(strings.TrimSpace(stringFromMap(raw, "response_format")))
	if responseFormat == "" {
		responseFormat = "mp3"
	}

	if !validAudioSpeechResponseFormat(responseFormat) {
		return metadata, fmt.Errorf("response_format must be one of mp3, opus, aac, flac, wav, or pcm")
	}

	metadata.ResponseFormat = responseFormat
	metadata.Metadata["response_format"] = responseFormat

	if speed, ok, err := numberFromMap(raw, "speed"); err != nil {
		return metadata, fmt.Errorf("speed must be a number")
	} else if ok {
		if speed < 0.25 || speed > 4 {
			return metadata, fmt.Errorf("speed must be between 0.25 and 4")
		}

		metadata.Metadata["speed"] = speed
	}

	streamFormat := strings.ToLower(strings.TrimSpace(stringFromMap(raw, "stream_format")))
	if streamFormat != "" {
		if streamFormat != "sse" && streamFormat != "audio" {
			return metadata, fmt.Errorf("stream_format must be sse or audio")
		}

		metadata.StreamFormat = streamFormat
		metadata.Stream = streamFormat == "sse"
		metadata.Metadata["stream_format"] = streamFormat
		metadata.Metadata["stream"] = metadata.Stream
	}

	return metadata, nil
}

func fillAudioSpeechVoiceMetadata(metadata map[string]interface{}, value interface{}) error {
	switch v := value.(type) {
	case string:
		voice := strings.TrimSpace(v)
		if voice == "" {
			return fmt.Errorf("voice is required")
		}

		metadata["voice_type"] = "builtin"
		metadata["voice"] = voice

		return nil
	case map[string]interface{}:
		id := strings.TrimSpace(stringFromMap(v, "id"))
		if id == "" {
			return fmt.Errorf("voice.id is required")
		}

		metadata["voice_type"] = "custom"
		metadata["voice_id_hash"] = hashTracePreview(id)

		return nil
	default:
		return fmt.Errorf("voice is required")
	}
}

func stringFromMap(values map[string]interface{}, key string) string {
	if len(values) == 0 {
		return ""
	}

	value, ok := values[key].(string)
	if !ok {
		return ""
	}

	return value
}

func numberFromMap(values map[string]interface{}, key string) (float64, bool, error) {
	value, exists := values[key]
	if !exists || value == nil {
		return 0, false, nil
	}

	switch v := value.(type) {
	case float64:
		return v, true, nil
	case int:
		return float64(v), true, nil
	case int64:
		return float64(v), true, nil
	case json.Number:
		parsed, err := v.Float64()
		return parsed, true, err
	default:
		return 0, true, fmt.Errorf("not a number")
	}
}

func validAudioSpeechResponseFormat(format string) bool {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "mp3", "opus", "aac", "flac", "wav", "pcm":
		return true
	default:
		return false
	}
}

func audioSpeechValidationErrorDetail(err error) (string, string) {
	message := ""
	if err != nil {
		message = strings.ToLower(strings.TrimSpace(err.Error()))
	}

	switch {
	case strings.Contains(message, "content type"):
		return "content_type", "unsupported_content_type"
	case strings.Contains(message, "json"):
		return "request", "invalid_json"
	case strings.Contains(message, "model"):
		return "model", "invalid_model"
	case strings.Contains(message, "input"):
		return "input", "invalid_input"
	case strings.Contains(message, "voice"):
		return "voice", "invalid_voice"
	case strings.Contains(message, "instructions"):
		return "instructions", "invalid_instructions"
	case strings.Contains(message, "response_format"):
		return "response_format", "invalid_response_format"
	case strings.Contains(message, "speed"):
		return "speed", "invalid_speed"
	case strings.Contains(message, "stream_format"):
		return "stream_format", "invalid_stream_format"
	default:
		return "request", "invalid_request"
	}
}

func estimateAudioSpeechCostFallback(model string, metadata map[string]interface{}) platform.UsageCostBreakdown {
	inputTokens := 0
	if inputLength, ok := metadata["input_length"]; ok {
		inputTokens += estimateTextTokensFromLength(inputLength)
	}

	if instructionsLength, ok := metadata["instructions_length"]; ok {
		inputTokens += estimateTextTokensFromLength(instructionsLength)
	}

	return platform.UsageCostBreakdown{
		Model:       model,
		RequestType: audioSpeechRequestType,
		Endpoint:    "/v1" + audioSpeechPath,
		InputTokens: inputTokens,
		TotalTokens: inputTokens,
	}
}

func audioSpeechInputCharacters(metadata map[string]interface{}) int {
	switch value := metadata["input_length"].(type) {
	case int:
		return value
	case int64:
		return int(value)
	case float64:
		return int(value)
	case json.Number:
		if parsed, err := value.Int64(); err == nil {
			return int(parsed)
		}
	}

	return 0
}

func audioSpeechUsageToOpenAI(body []byte) *contractsopenai.Usage {
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

func withAudioSpeechResponseMetadata(input map[string]interface{}, body []byte, headers map[string]string) map[string]interface{} {
	out := map[string]interface{}{}
	for k, v := range input {
		out[k] = v
	}

	if contentType := responseContentType(headers); contentType != "" {
		out["response_content_type"] = safeContentTypeFamily(contentType)
	}

	out["response_summary"] = buildTraceAudioSpeechResponsePayload(body, headers)

	return out
}

func supportsAudioSpeechProvider(provider string) bool {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "", "openai", "openrouter", "deepinfra", "fireworks", "togetherai", "together-ai", "sambanova", "mock":
		return true
	default:
		return false
	}
}
