package proxy

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	metricsobs "github.com/Delpak/shahrah/services/control-plane/internal/observability/metrics"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// ResponsesHandler serves OpenAI's responses API for OpenAI-native models.
type ResponsesHandler struct {
	llmRouter  *router.Router
	tunnel     platform.TunnelTransport
	billing    budgeter
	wallets    walletStore
	logs       usageLogger
	logger     *zap.Logger
	emitMetric func(input metricsobs.BuildInput)
}

// NewResponsesHandler creates a new responses handler.
func NewResponsesHandler(llmRouter *router.Router, tunnel platform.TunnelTransport, billing budgeter, wallets walletStore, logs usageLogger, logger *zap.Logger, emitMetric ...func(input metricsobs.BuildInput)) *ResponsesHandler {
	var metricEmitter func(input metricsobs.BuildInput)
	if len(emitMetric) > 0 {
		metricEmitter = emitMetric[0]
	}
	return &ResponsesHandler{
		llmRouter:  llmRouter,
		tunnel:     tunnel,
		billing:    billing,
		wallets:    wallets,
		logs:       logs,
		logger:     logger,
		emitMetric: metricEmitter,
	}
}

// Handle processes an OpenAI responses request.
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

	var req contractsopenai.ResponseRequest
	if err := c.BodyParser(&req); err != nil {
		return writeAPIError(c, fiber.StatusBadRequest, "Invalid request body", "invalid_request_error")
	}
	if strings.TrimSpace(req.Model) == "" {
		return writeAPIError(c, fiber.StatusBadRequest, "Model is required", "invalid_request_error")
	}
	requestedModel := normalizeModelIdentifier(req.Model)
	routeCandidates := modelRoutingCandidates(requestedModel)
	if len(routeCandidates) == 0 {
		return writeAPIError(c, fiber.StatusBadRequest, "Model is required", "invalid_request_error")
	}
	tracing.SetModelForTrace(c, requestedModel)
	if !hasResponseInput(req.Input, req.Instructions) {
		return writeAPIError(c, fiber.StatusBadRequest, "input is required", "invalid_request_error")
	}
	if req.Temperature != nil && (*req.Temperature < 0 || *req.Temperature > 2) {
		return writeAPIError(c, fiber.StatusBadRequest, "temperature must be between 0 and 2", "invalid_request_error")
	}
	if req.TopP != nil && (*req.TopP < 0 || *req.TopP > 1) {
		return writeAPIError(c, fiber.StatusBadRequest, "top_p must be between 0 and 1", "invalid_request_error")
	}
	if req.MaxOutputTokens != nil && *req.MaxOutputTokens < 1 {
		return writeAPIError(c, fiber.StatusBadRequest, "max_output_tokens must be at least 1", "invalid_request_error")
	}
	modelAccess := modelAccessForType(authInfo, requestedModel, platform.ModelTypeChat)
	if !modelAccess.Allowed {
		if modelAccess.UnsupportedVMMode {
			return writeUnsupportedModelModeError(c, requestedModel, platform.ModelTypeChat, modelAccess.ExposedModes)
		}
		emitDenyCounter(h.logger, denyReasonForModel(authInfo), requestedModel, currentTenant.ID)
		return writeAPIError(c, fiber.StatusForbidden, "Model access denied or unsupported for responses", "authorization_error")
	}

	authStart := time.Now()
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "auth.validate", "internal", authStart, fiber.StatusOK, map[string]interface{}{"model": requestedModel})

	estimateReq := responseRequestToChatCompletion(req, requestedModel)
	quotaStart := time.Now()
	estimated, err := h.billing.EstimateCost(c.Context(), estimateReq)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, "Failed to estimate request cost", "billing_error")
	}
	rateLimit, err := h.billing.CheckRateLimit(c.Context(), authInfo, requestedModel, estimated.TotalTokens)
	setRateLimitHeaders(c, rateLimit)
	if err != nil {
		return writeAPIError(c, rateLimitErrorStatus(err), err.Error(), "rate_limit_error")
	}
	if err := h.billing.CheckBudget(c.Context(), authInfo, estimated); err != nil {
		return writeAPIError(c, fiber.StatusPaymentRequired, err.Error(), "budget_exceeded")
	}
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "quota.check", "internal", quotaStart, fiber.StatusOK, map[string]interface{}{"estimated_tokens": estimated.TotalTokens})

	if err := h.tunnel.Healthy(c.Context()); err != nil {
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
	if !supportsResponsesProvider(dep.Provider) {
		return writeAPIError(c, fiber.StatusNotImplemented, fmt.Sprintf("Responses API is not enabled for provider: %s", dep.Provider), "invalid_request_error")
	}
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "router.select_model", "internal", routeStart, fiber.StatusOK, map[string]interface{}{"provider": dep.Provider, "deployment_id": dep.ID})

	outboundReq := req
	outboundModel := strings.TrimSpace(dep.Model)
	if outboundModel == "" || outboundModel == "*" {
		outboundModel = providerModelID(requestedModel)
	}
	outboundReq.Model = outboundModel

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
		Stream:     req.Stream,
		Deployment: toContractDeployment(dep),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: payload,
		RequestMetadata: map[string]interface{}{
			"tenant_slug":            strings.TrimSpace(c.Get("x-shahrah-tenant")),
			"model":                  requestedModel,
			"provider_endpoint_path": "/responses",
		},
	}

	if req.Stream {
		return h.handleStream(c, req, authInfo, estimated, tunnelReq, startTime, tenantSlug)
	}
	return h.handleSync(c, req, authInfo, estimated, tunnelReq, startTime, tenantSlug)
}

func (h *ResponsesHandler) handleSync(c *fiber.Ctx, req contractsopenai.ResponseRequest, authInfo *platform.AuthContext, estimated platform.UsageCostBreakdown, tunnelReq platform.TunnelRequest, start time.Time, tenantSlug string) error {
	providerStart := time.Now()
	providerAttrs := map[string]interface{}{
		"provider":        tunnelReq.Deployment.Provider,
		"request_payload": buildTraceResponseRequestPayload(req),
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

	var response contractsopenai.Response
	if err := json.Unmarshal(resp.Body, &response); err != nil {
		providerAttrs["error"] = "failed to decode provider response"
		providerAttrs["response_payload"] = buildTraceProviderErrorPayload(fiber.StatusBadGateway, resp.Body)
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, fiber.StatusBadGateway, providerAttrs)
		return writeAPIError(c, fiber.StatusBadGateway, "Failed to decode provider response", "api_error")
	}
	providerAttrs["response_payload"] = buildTraceResponsePayload(response)
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, resp.StatusCode, providerAttrs)

	finalCost, err := h.billing.FinalizeFromUsage(c.Context(), billingModelForDeployment(req.Model, tunnelReq.Deployment), responseUsageToOpenAI(response.Usage), estimated)
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
		ResponseBody:   map[string]interface{}{"id": response.ID},
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
		Metadata: map[string]interface{}{
			"endpoint":    "/v1/responses",
			"key_type":    authInfo.KeyType,
			"team_id":     authInfo.TeamID,
			"org_unit_id": authInfo.OrgUnitID,
			"tenant_slug": tenantSlug,
		},
	})

	forwardProviderHeaders(c, resp.Headers)
	c.Set("x-shahrah-trace-id", tunnelReq.TraceID)
	c.Set("x-shahrah-cost-usd", fmt.Sprintf("%.8f", finalCost.TotalCostUSD))
	c.Set("x-shahrah-cost-irr", fmt.Sprintf("%.2f", finalCost.TotalCostIRR))
	responseStart := time.Now()
	err = c.Status(resp.StatusCode).JSON(response)
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "response.write", "internal", responseStart, resp.StatusCode, map[string]interface{}{
		"response_payload": buildTraceResponsePayload(response),
	})
	return err
}

func (h *ResponsesHandler) handleStream(c *fiber.Ctx, req contractsopenai.ResponseRequest, authInfo *platform.AuthContext, estimated platform.UsageCostBreakdown, tunnelReq platform.TunnelRequest, start time.Time, tenantSlug string) error {
	providerStart := time.Now()
	stream, err := h.tunnel.Stream(c.Context(), tunnelReq)
	if err != nil {
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, fiber.StatusBadGateway, map[string]interface{}{
			"provider":        tunnelReq.Deployment.Provider,
			"request_payload": buildTraceResponseRequestPayload(req),
			"stream":          true,
			"error":           err.Error(),
		})
		return writeAPIError(c, fiber.StatusBadGateway, fmt.Sprintf("Tunnel error: %v", err), "api_error")
	}
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, fiber.StatusOK, map[string]interface{}{
		"provider":        tunnelReq.Deployment.Provider,
		"request_payload": buildTraceResponseRequestPayload(req),
		"stream":          true,
	})

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("x-shahrah-trace-id", tunnelReq.TraceID)
	forwardProviderHeaders(c, stream.Headers())

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		defer stream.Close()

		for {
			select {
			case chunk, ok := <-stream.Chunks():
				if !ok {
					finalCost, _ := h.billing.FinalizeFromUsage(c.Context(), billingModelForDeployment(req.Model, tunnelReq.Deployment), nil, estimated)
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
						Metadata: map[string]interface{}{
							"endpoint":    "/v1/responses",
							"key_type":    authInfo.KeyType,
							"team_id":     authInfo.TeamID,
							"org_unit_id": authInfo.OrgUnitID,
							"tenant_slug": tenantSlug,
						},
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
			}
		}
	})
	responseStart := time.Now()
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "response.write", "internal", responseStart, fiber.StatusOK, map[string]interface{}{"stream": true})
	return nil
}

func (h *ResponsesHandler) emitModelMetric(input metricsobs.BuildInput) {
	if h.emitMetric == nil {
		return
	}
	h.emitMetric(input)
}

func supportsResponsesProvider(provider string) bool {
	return strings.EqualFold(strings.TrimSpace(provider), "openai")
}

func hasResponseInput(input interface{}, instructions string) bool {
	if strings.TrimSpace(instructions) != "" {
		return true
	}
	switch v := input.(type) {
	case nil:
		return false
	case string:
		return strings.TrimSpace(v) != ""
	case []interface{}:
		return len(v) > 0
	case map[string]interface{}:
		return len(v) > 0
	default:
		return true
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
	return &contractsopenai.Usage{
		PromptTokens:     usage.InputTokens,
		CompletionTokens: usage.OutputTokens,
		TotalTokens:      usage.TotalTokens,
	}
}
