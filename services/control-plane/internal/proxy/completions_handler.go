package proxy

import (
	"bufio"
	"bytes"
	"context"
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

// CompletionsHandler handles legacy text completion requests.
type CompletionsHandler struct {
	llmRouter  *router.Router
	tunnel     platform.TunnelTransport
	billing    budgeter
	wallets    walletStore
	logs       usageLogger
	logger     *zap.Logger
	emitMetric func(input metricsobs.BuildInput)
}

// NewCompletionsHandlerWithDependencies creates a completions handler from explicit dependencies.
func NewCompletionsHandlerWithDependencies(deps BaseHandlerDependencies) *CompletionsHandler {
	return &CompletionsHandler{
		llmRouter:  deps.Router,
		tunnel:     deps.Tunnel,
		billing:    deps.Billing,
		wallets:    deps.Wallets,
		logs:       deps.logger(),
		logger:     deps.Logger,
		emitMetric: deps.metricEmitter(),
	}
}

// NewCompletionsHandler creates a new legacy completions handler.
func NewCompletionsHandler(llmRouter *router.Router, tunnel platform.TunnelTransport, billing budgeter, wallets walletStore, logs usageLogger, logger *zap.Logger, emitMetric ...func(input metricsobs.BuildInput)) *CompletionsHandler {
	var metricEmitter func(input metricsobs.BuildInput)
	if len(emitMetric) > 0 {
		metricEmitter = emitMetric[0]
	}

	return NewCompletionsHandlerWithDependencies(BaseHandlerDependencies{
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

// Handle processes a legacy text completion request.
func (h *CompletionsHandler) Handle(c *fiber.Ctx) error {
	startTime := time.Now()

	originalBody := append([]byte(nil), c.Body()...)

	var req contractsopenai.CompletionRequest
	if err := json.Unmarshal(originalBody, &req); err != nil {
		return writeAPIError(c, fiber.StatusBadRequest, "invalid request body", "invalid_request_error")
	}

	// Validate required fields
	if req.Model == "" {
		return writeAPIError(c, fiber.StatusBadRequest, "model is required", "invalid_request_error")
	}

	requestedModel := normalizeModelIdentifier(req.Model)

	routeCandidates := modelRoutingCandidates(requestedModel)
	if len(routeCandidates) == 0 {
		return writeAPIError(c, fiber.StatusBadRequest, "model is required", "invalid_request_error")
	}

	resolvedModel := providerModelID(requestedModel)
	tracing.SetModelForTrace(c, requestedModel)

	if req.Prompt == "" {
		return writeAPIError(c, fiber.StatusBadRequest, "prompt is required", "invalid_request_error")
	}

	// Get auth context from middleware
	authInfo := auth.GetAuthInfo(c)
	if authInfo == nil {
		return writeAPIError(c, fiber.StatusUnauthorized, "authentication required", "authentication_error")
	}

	// Get tenant context
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return writeAPIError(c, fiber.StatusBadRequest, "tenant resolution required", "invalid_request_error")
	}

	authInfo.TenantID = currentTenant.ID
	tracing.SetTenantForTrace(c, currentTenant.ID)
	tenantSlug := strings.TrimSpace(currentTenant.Slug)

	modelAccess := modelAccessForType(authInfo, requestedModel, platform.ModelTypeCompletion)
	if !modelAccess.Allowed {
		if modelAccess.UnsupportedVMMode {
			return writeUnsupportedModelModeError(c, requestedModel, platform.ModelTypeCompletion, modelAccess.ExposedModes)
		}

		denyReason := denyReasonForModel(authInfo)
		emitDenyCounter(h.logger, denyReason, requestedModel, currentTenant.ID)

		const message = "model access denied or unsupported for completions"
		logPreProviderRejection(h.logger, h.logs, c, authInfo, requestedModel, "/v1/completions", "completions", fiber.StatusForbidden, "authorization_error", message, denyReason, startTime)

		return writeAPIError(c, fiber.StatusForbidden, message, "authorization_error")
	}

	authStart := time.Now()
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "auth.validate", "internal", authStart, fiber.StatusOK, map[string]interface{}{"model": requestedModel})

	// Transform to chat completion format internally
	chatReq := contractsopenai.ChatCompletionRequest{
		Model:       resolvedModel,
		Messages:    []contractsopenai.ChatMessage{{Role: "user", Content: req.Prompt}},
		Stream:      req.Stream,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
	}

	// Estimate cost using the same billing logic
	estimated, err := h.billing.EstimateCost(c.Context(), chatReq)
	if err != nil {
		h.logger.Error("failed to estimate cost", zap.Error(err))

		estimated = platform.UsageCostBreakdown{
			Model:        req.Model,
			TotalTokens:  512,
			TotalCostUSD: 0.001,
			TotalCostIRR: 850,
		}
	}

	quotaStart := time.Now()
	// Check rate limit
	rateLimit, err := h.billing.CheckRateLimit(c.Context(), authInfo, requestedModel, estimated.TotalTokens)
	setRateLimitHeaders(c, rateLimit)

	if err != nil {
		recordRateLimitRejection(c, h.logger, h.logs, h.emitMetric, authInfo, requestedModel, "/v1/completions", "completions", startTime, quotaStart, rateLimit, err)
		return writeAPIError(c, rateLimitErrorStatus(err), err.Error(), "rate_limit_error")
	}

	requestID := tracing.TraceIDFromContext(c)
	if requestID == "" {
		requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
	}

	var budgetReservation *platform.BudgetReservation
	var budgetResponseHandled bool

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

	// Route to deployment
	routeStart := time.Now()

	var (
		dep *router.Deployment
	)
	for _, candidate := range routeCandidates {
		dep, err = h.llmRouter.Route(c.Context(), currentTenant.ID, candidate)
		if err == nil {
			break
		}
	}

	if err != nil {
		emitDenyCounter(h.logger, denyReasonUnknownModel, requestedModel, currentTenant.ID)
		return writeAPIError(c, fiber.StatusInternalServerError, "no available deployments", "internal_error")
	}

	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "router.select_model", "internal", routeStart, fiber.StatusOK, map[string]interface{}{"provider": dep.Provider, "deployment_id": dep.ID})

	outboundModel := strings.TrimSpace(dep.Model)
	if outboundModel == "" || outboundModel == "*" {
		outboundModel = resolvedModel
	}

	chatReq.Model = outboundModel

	tunnelReq := platform.TunnelRequest{
		RequestID:  requestID,
		TenantID:   currentTenant.ID,
		TraceID:    requestID,
		Stream:     req.Stream,
		Deployment: toContractDeployment(dep),
		Headers:    map[string]string{"Authorization": "Bearer " + dep.Params.APIKey},
		Body:       mustMarshalJSON(chatReq),
		RequestMetadata: map[string]interface{}{
			"tenant_slug":            apiheaders.GetTenant(c),
			"model":                  requestedModel,
			"provider_endpoint_path": "/completions",
		},
	}
	tunnelReq.RequestMetadata = enrichRequestMetadataForDeployment(tunnelReq.RequestMetadata, tunnelReq.Deployment)

	if !budgetReservationAttempted {
		budgetReservation, budgetResponseHandled, err = reserveBudgetForRequest(c, h.billing, authInfo, estimated, requestID, deploymentBillingMode(tunnelReq.Deployment))
		if err != nil || budgetResponseHandled {
			return err
		}
	}

	// Execute request
	if req.Stream {
		releaseBudgetOnExit = false
		return h.handleStream(c, chatReq, req, originalBody, authInfo, estimated, tunnelReq, budgetReservation, startTime, tenantSlug)
	}

	err = h.handleSync(c, chatReq, originalBody, req, authInfo, estimated, tunnelReq, budgetReservation, startTime, tenantSlug)
	if err == nil {
		releaseBudgetOnExit = false
	}

	return err
}

func (h *CompletionsHandler) handleSync(
	c *fiber.Ctx,
	req contractsopenai.ChatCompletionRequest,
	originalBody []byte,
	clientReq contractsopenai.CompletionRequest,
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
		"request_payload": buildTraceChatRequestPayload(req),
	}

	resp, err := h.tunnel.Proxy(c.Context(), tunnelReq)
	if err != nil {
		providerAttrs["error"] = err.Error()
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, fiber.StatusBadGateway, providerAttrs)

		return writeAPIError(c, fiber.StatusBadGateway, "provider request failed", "bad_gateway")
	}

	if resp.StatusCode >= httpErrorFloor {
		providerAttrs["response_payload"] = buildTraceProviderErrorPayload(resp.StatusCode, resp.Body)
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, resp.StatusCode, providerAttrs)

		return transformProviderError(c, resp.StatusCode, resp.Body, resp.Headers)
	}

	forwardProviderHeaders(c, resp.Headers)

	// Transform chat completion response to legacy completion format
	var chatResp contractsopenai.ChatCompletionResponse

	unmarshalErr := json.Unmarshal(resp.Body, &chatResp)
	if unmarshalErr != nil {
		providerAttrs["error"] = "failed to decode provider response"
		providerAttrs["response_payload"] = buildTraceProviderErrorPayload(fiber.StatusBadGateway, resp.Body)
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, fiber.StatusBadGateway, providerAttrs)

		return c.Status(fiber.StatusOK).Send(resp.Body)
	}

	// Convert to legacy format
	completionResp := contractsopenai.CompletionResponse{
		ID:      chatResp.ID,
		Object:  "text_completion",
		Created: chatResp.Created,
		Model:   chatResp.Model,
	}

	choices := make([]contractsopenai.CompletionChoice, 0, len(chatResp.Choices))
	for _, choice := range chatResp.Choices {
		if choice.Message != nil {
			choices = append(choices, contractsopenai.CompletionChoice{
				Text:         contractsopenai.ChatContentText(choice.Message.Content),
				Index:        choice.Index,
				FinishReason: choice.FinishReason,
			})
		}
	}

	completionResp.Choices = choices
	completionResp.Usage = chatResp.Usage
	providerAttrs["response_payload"] = buildTraceCompletionResponsePayload(completionResp)
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, resp.StatusCode, providerAttrs)

	// Finalize billing
	finalCost, finalizeErr := h.billing.FinalizeFromUsage(c.Context(), req.Model, chatResp.Usage, estimated)
	if finalizeErr != nil {
		return writeAPIError(c, fiber.StatusBadGateway, "failed to finalize usage", "billing_error")
	}

	if debitErr := commitBudgetOrDebit(c.Context(), h.logger, h.billing, h.wallets, budgetReservation, authInfo.TenantID, finalCost, tunnelReq.RequestID, deploymentBillingMode(tunnelReq.Deployment)); debitErr != nil {
		return writeBudgetCheckError(c, debitErr)
	}

	setCostHeaders(c, finalCost)
	setTraceHeaders(c, tunnelReq.RequestID)

	// Log asynchronously
	responseTimeMS := int(time.Since(start).Milliseconds())
	go func() {
		logEntry := buildRequestLogEntry(nil, authInfo, tunnelReq.Deployment, req.Model, "/v1/completions", "completions", "success", resp.StatusCode, start, responseTimeMS)
		logEntry.RequestID = tunnelReq.RequestID
		logEntry.TraceID = tunnelReq.TraceID
		logEntry.CostUSD = finalCost.TotalCostUSD
		logEntry.CostIRR = finalCost.TotalCostIRR
		logEntry.InputTokens = finalCost.InputTokens
		logEntry.OutputTokens = finalCost.OutputTokens
		logEntry.TotalTokens = finalCost.TotalTokens
		logEntry.Status = fmt.Sprintf("%d", resp.StatusCode)
		logEntry.RequestBody = buildCompletionRequestLogPayload(clientReq, originalBody)
		logEntry.ResponseBody = buildJSONResponseLogPayload(nil, completionResp)
		logRequestBestEffort(h.logger, h.logs, context.Background(), logEntry)
		logSpendBestEffort(h.logger, h.logs, context.Background(), platform.SpendLog{
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
			RequestType:        "completions",
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
				"endpoint":    "/v1/completions",
				"key_type":    authInfo.KeyType,
				"team_id":     authInfo.TeamID,
				"org_unit_id": authInfo.OrgUnitID,
				"tenant_slug": tenantSlug,
			}, tunnelReq.Deployment),
		})
	}()

	responseStart := time.Now()
	err = c.Status(resp.StatusCode).JSON(completionResp)
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "response.write", "internal", responseStart, resp.StatusCode, map[string]interface{}{
		"response_payload": buildTraceCompletionResponsePayload(completionResp),
	})

	return err
}

func buildCompletionRequestLogPayload(req contractsopenai.CompletionRequest, originalBody []byte) map[string]interface{} {
	return buildJSONRequestLogPayload(originalBody, map[string]interface{}{
		"model":  req.Model,
		"prompt": req.Prompt,
	})
}

func (h *CompletionsHandler) handleStream(c *fiber.Ctx, req contractsopenai.ChatCompletionRequest, clientReq contractsopenai.CompletionRequest, originalBody []byte, authInfo *platform.AuthContext, estimated platform.UsageCostBreakdown, tunnelReq platform.TunnelRequest, budgetReservation *platform.BudgetReservation, start time.Time, tenantSlug string) error {
	requestLogBase := buildRequestLogEntry(c, authInfo, tunnelReq.Deployment, req.Model, "/v1/completions", "completions", "failed", 0, start, 0)
	requestLogBase.RequestID = tunnelReq.RequestID
	requestLogBase.TraceID = tunnelReq.TraceID
	requestLogBase.RequestBody = buildCompletionRequestLogPayload(clientReq, originalBody)

	providerStart := time.Now()

	stream, err := h.tunnel.Stream(c.Context(), tunnelReq)
	if err != nil {
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, fiber.StatusBadGateway, map[string]interface{}{
			"provider":        tunnelReq.Deployment.Provider,
			"request_payload": buildTraceChatRequestPayload(req),
			"stream":          true,
			"error":           err.Error(),
		})

		releaseBudgetBestEffort(c.Context(), h.logger, h.billing, budgetReservation, "stream_open_failed")
		logStreamingRequestBestEffort(h.logger, h.logs, context.Background(), requestLogBase, "failed", fiber.StatusBadGateway, "stream_open_error", map[string]interface{}{
			"stream":  true,
			"outcome": "open_failed",
			"error":   map[string]interface{}{"message": err.Error(), "type": "stream_open_error"},
		}, nil)

		return writeAPIError(c, fiber.StatusBadGateway, "stream failed", "bad_gateway")
	}
	if statusCode := stream.StatusCode(); statusCode >= httpErrorFloor {
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, statusCode, map[string]interface{}{
			"provider":        tunnelReq.Deployment.Provider,
			"request_payload": buildTraceChatRequestPayload(req),
			"stream":          true,
			"status_code":     statusCode,
		})
		releaseBudgetBestEffort(c.Context(), h.logger, h.billing, budgetReservation, "provider_http_error")

		return writeTunnelStreamHTTPError(c, h.logger, stream, func(statusCode int, body []byte) {
			responsePayload := buildJSONResponseLogPayload(body, map[string]interface{}{"stream": true, "outcome": "provider_http_error"})
			responsePayload["stream"] = true
			responsePayload["outcome"] = "provider_http_error"
			errorType := ErrorTypeMap[statusCode]
			if strings.TrimSpace(errorType) == "" {
				errorType = "api_error"
			}
			logStreamingRequestBestEffort(h.logger, h.logs, context.Background(), requestLogBase, "failed", statusCode, errorType, responsePayload, nil)
		})
	}

	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, fiber.StatusOK, map[string]interface{}{
		"provider":        tunnelReq.Deployment.Provider,
		"request_payload": buildTraceChatRequestPayload(req),
		"stream":          true,
	})

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no")
	setTraceHeaders(c, tunnelReq.TraceID)

	// Forward provider headers from stream
	forwardProviderHeaders(c, stream.Headers())

	requestCtx := c.Context()
	requestCtx.SetBodyStreamWriter(func(w *bufio.Writer) {
		defer closeStreamBestEffort(h.logger, stream)
		responseCapture := &streamResponseLogCapture{}
		logStreamFailure := func(errorType string, message string) {
			payload := responseCapture.Payload("failed")
			payload["error"] = map[string]interface{}{"message": message, "type": errorType}
			logStreamingRequestBestEffort(h.logger, h.logs, context.Background(), requestLogBase, "failed", fiber.StatusOK, errorType, payload, nil)
		}

		budgetCommitted := false
		defer func() {
			if !budgetCommitted {
				releaseBudgetBestEffort(requestCtx, h.logger, h.billing, budgetReservation, "stream_not_committed")
			}
		}()

		finalizeAndLogSuccess := func() {
			if budgetCommitted {
				return
			}

			finalCost, finalizeErr := h.billing.FinalizeFromUsage(requestCtx, req.Model, nil, estimated)
			if finalizeErr != nil {
				h.logger.Warn("failed to finalize streaming usage", zap.Error(finalizeErr), zap.String("request_id", tunnelReq.RequestID))

				finalCost = estimated
				finalCost.Model = req.Model
			}

			if debitErr := commitBudgetOrDebit(requestCtx, h.logger, h.billing, h.wallets, budgetReservation, authInfo.TenantID, finalCost, tunnelReq.RequestID, deploymentBillingMode(tunnelReq.Deployment)); debitErr != nil {
				h.logger.Warn("failed to commit budget for streaming request", zap.Error(debitErr), zap.String("request_id", tunnelReq.RequestID))
			} else {
				budgetCommitted = true
			}

			logSpendBestEffort(h.logger, h.logs, requestCtx, platform.SpendLog{
				RequestID:       tunnelReq.RequestID,
				TokenHash:       authInfo.TokenHash,
				TenantID:        authInfo.TenantID,
				Model:           req.Model,
				InputTokens:     finalCost.InputTokens,
				OutputTokens:    finalCost.OutputTokens,
				TotalTokens:     finalCost.TotalTokens,
				CostUSD:         finalCost.TotalCostUSD,
				CostIRR:         finalCost.TotalCostIRR,
				ResponseTimeMS:  int(time.Since(start).Milliseconds()),
				Status:          "success",
				RequestMetadata: enrichPricingMetadata(tunnelReq.RequestMetadata, authInfo, req.Model),
			})
			logStreamingRequestBestEffort(h.logger, h.logs, context.Background(), requestLogBase, "success", fiber.StatusOK, "", responseCapture.Payload("completed"), &finalCost)
			h.emitModelMetric(metricsobs.BuildInput{
				TS:                 time.Now().UTC(),
				TenantID:           authInfo.TenantID,
				RequestID:          tunnelReq.RequestID,
				ModelName:          req.Model,
				ProviderName:       tunnelReq.Deployment.Provider,
				ProviderModelName:  tunnelReq.Deployment.Model,
				RequestType:        "completions",
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
					"endpoint":    "/v1/completions",
					"key_type":    authInfo.KeyType,
					"team_id":     authInfo.TeamID,
					"org_unit_id": authInfo.OrgUnitID,
					"tenant_slug": tenantSlug,
				}, tunnelReq.Deployment),
			})
		}

		for {
			select {
			case chunk, ok := <-stream.Chunks():
				if !ok {
					finalizeAndLogSuccess()
					return
				}

				// Transform chat chunk to legacy completion format
				if len(chunk.Data) > 0 {
					raw := bytes.TrimSpace(chunk.Data)
					if bytes.Equal(raw, []byte("data: [DONE]")) || bytes.Equal(raw, []byte("[DONE]")) {
						doneMarker := []byte("data: [DONE]\n\n")
						responseCapture.Append(doneMarker)
						if writeErr := writeBufferedBytes(w, doneMarker); writeErr != nil {
							h.logger.Warn("failed to write done marker", zap.Error(writeErr), zap.String("request_id", tunnelReq.RequestID))
							logStreamFailure("stream_write_error", writeErr.Error())
							return
						}

						finalizeAndLogSuccess()

						return
					}

					normalized := raw
					if bytes.HasPrefix(normalized, []byte("data: ")) {
						normalized = bytes.TrimSpace(bytes.TrimPrefix(normalized, []byte("data: ")))
					}

					var chatChunk contractsopenai.StreamChunk
					if err := json.Unmarshal(normalized, &chatChunk); err != nil {
						if !bytes.HasSuffix(chunk.Data, []byte("\n\n")) {
							chunk.Data = append(chunk.Data, []byte("\n\n")...)
						}

						responseCapture.Append(chunk.Data)
						if writeErr := writeBufferedBytes(w, chunk.Data); writeErr != nil {
							h.logger.Warn("failed to forward raw stream chunk", zap.Error(writeErr), zap.String("request_id", tunnelReq.RequestID))
							logStreamFailure("stream_write_error", writeErr.Error())
							return
						}

						continue
					}

					// Extract text from chat delta
					var text string

					for _, choice := range chatChunk.Choices {
						if choice.Delta != nil {
							text = contractsopenai.ChatContentText(choice.Delta.Content)
							break
						}
					}

					if text != "" {
						legacyChunk := map[string]interface{}{
							"id":      chatChunk.ID,
							"object":  "text_completion",
							"created": chatChunk.Created,
							"model":   chatChunk.Model,
							"choices": []map[string]interface{}{
								{"text": text, "index": 0, "finish_reason": nil},
							},
						}

						data, marshalErr := json.Marshal(legacyChunk)
						if marshalErr != nil {
							h.logger.Warn("failed to encode legacy completion chunk", zap.Error(marshalErr), zap.String("request_id", tunnelReq.RequestID))
							logStreamFailure("stream_encoding_error", marshalErr.Error())
							return
						}

						legacyEvent := bytes.Join([][]byte{[]byte("data: "), data, []byte("\n\n")}, nil)
						responseCapture.Append(legacyEvent)
						if writeErr := writeBufferedBytes(w, legacyEvent); writeErr != nil {
							h.logger.Warn("failed to write legacy completion chunk", zap.Error(writeErr), zap.String("request_id", tunnelReq.RequestID))
							logStreamFailure("stream_write_error", writeErr.Error())
							return
						}
					}

					if chunk.IsDone {
						doneMarker := []byte("data: [DONE]\n\n")
						responseCapture.Append(doneMarker)
						if writeErr := writeBufferedBytes(w, doneMarker); writeErr != nil {
							h.logger.Warn("failed to write final done marker", zap.Error(writeErr), zap.String("request_id", tunnelReq.RequestID))
							logStreamFailure("stream_write_error", writeErr.Error())
							return
						}

						finalizeAndLogSuccess()

						return
					}
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
						logStreamFailure("streaming_error", err.Error())
						return
					}

					streamErrorEvent := bytes.Join([][]byte{[]byte("data: "), errorPayload, []byte("\n\n")}, nil)
					responseCapture.Append(streamErrorEvent)
					if writeErr := writeBufferedBytes(w, streamErrorEvent); writeErr != nil {
						h.logger.Warn("failed to write streaming error payload", zap.Error(writeErr), zap.String("request_id", tunnelReq.RequestID))
					}
					logStreamFailure("streaming_error", err.Error())
				} else {
					logStreamFailure("stream_closed_without_completion", "provider stream closed without a completion marker")
				}

				return
			}
		}
	})

	responseStart := time.Now()
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "response.write", "internal", responseStart, fiber.StatusOK, map[string]interface{}{"stream": true})

	return nil
}

func (h *CompletionsHandler) emitModelMetric(input metricsobs.BuildInput) {
	if h.emitMetric == nil {
		return
	}

	h.emitMetric(input)
}

// setCostHeaders sets cost-related response headers.
func setCostHeaders(c *fiber.Ctx, cost platform.UsageCostBreakdown) {
	c.Set(apiheaders.HeaderCostUSD, fmt.Sprintf("%.8f", cost.TotalCostUSD))
	c.Set(apiheaders.HeaderCostIRR, fmt.Sprintf("%.2f", cost.TotalCostIRR))
}

func setTraceHeaders(c *fiber.Ctx, traceID string) {
	c.Set(apiheaders.HeaderTraceID, traceID)
}

// mustMarshalJSON marshals to JSON or panics.
func mustMarshalJSON(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return data
}
