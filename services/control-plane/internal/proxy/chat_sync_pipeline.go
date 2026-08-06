package proxy

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/guardrails"
	metricsobs "github.com/Delpak/shahrah/services/control-plane/internal/observability/metrics"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	"github.com/gofiber/fiber/v2"
)

type chatProviderResult struct {
	tunnelRequest platform.TunnelRequest
	response      *platform.TunnelResponse
	completion    contractsopenai.ChatCompletionResponse
	finalCost     platform.UsageCostBreakdown
	usageRepair   googleGeminiUsageRepair
	handled       bool
}

func (h *ChatCompletionHandler) executeChatProvider(
	c *fiber.Ctx,
	request contractsopenai.ChatCompletionRequest,
	originalBody []byte,
	authInfo *platform.AuthContext,
	estimated platform.UsageCostBreakdown,
	tunnelRequest platform.TunnelRequest,
	reservation *platform.BudgetReservation,
	startedAt time.Time,
) (chatProviderResult, error) {
	providerStart := time.Now()
	providerAttrs := enrichPlaygroundObservability(c, map[string]interface{}{
		"provider":        tunnelRequest.Deployment.Provider,
		"request_payload": buildProviderTraceChatRequestPayload(tunnelRequest.Deployment.Provider, request),
	})

	response, err := h.dispatchChatProvider(c, tunnelRequest)
	if err != nil {
		providerAttrs["error"] = err.Error()
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, fiber.StatusBadGateway, providerAttrs)
		h.llmRouter.MarkFailed(c.Context(), &router.Deployment{ID: tunnelRequest.Deployment.ID}, err.Error())

		return chatProviderResult{handled: true}, writeAPIError(c, fiber.StatusBadGateway, fmt.Sprintf("Tunnel error: %v", err), "api_error")
	}

	if response == nil {
		providerAttrs["error"] = "tunnel returned an empty response"
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, fiber.StatusBadGateway, providerAttrs)

		return chatProviderResult{handled: true}, writeAPIError(c, fiber.StatusBadGateway, "Tunnel returned an empty response", "api_error")
	}

	if response.StatusCode >= httpErrorFloor {
		providerAttrs["response_payload"] = buildProviderTraceErrorPayload(tunnelRequest.Deployment.Provider, response.StatusCode, response.Body)
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, response.StatusCode, providerAttrs)
		logProviderFailureRequest(h.logger, h.logs, c, authInfo, tunnelRequest, request.Model, "/v1/chat/completions", "chat_completions", response.StatusCode, response.Body, startedAt)

		return chatProviderResult{handled: true}, transformProviderError(c, response.StatusCode, response.Body, response.Headers)
	}

	if strings.EqualFold(strings.TrimSpace(tunnelRequest.Deployment.Provider), "xai") && chatRequestDeferred(request) {
		err = h.handleXAIDeferredCreate(c, request, originalBody, authInfo, estimated, tunnelRequest, reservation, response, startedAt)

		return chatProviderResult{handled: true}, err
	}

	var completion contractsopenai.ChatCompletionResponse
	if err = json.Unmarshal(response.Body, &completion); err != nil {
		providerAttrs["error"] = "failed to decode provider response"
		providerAttrs["response_payload"] = buildProviderTraceErrorPayload(tunnelRequest.Deployment.Provider, fiber.StatusBadGateway, response.Body)
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, fiber.StatusBadGateway, providerAttrs)

		return chatProviderResult{handled: true}, writeAPIError(c, fiber.StatusBadGateway, "Failed to decode provider response", "api_error")
	}

	if completion.Usage != nil {
		if tier := responseHeaderValue(response.Headers, "x-gemini-service-tier"); tier != "" {
			completion.Usage.ServiceTier = tier
		} else if completion.Usage.ServiceTier == "" {
			completion.Usage.ServiceTier = completion.ServiceTier
		}
	}

	usageRepair := normalizeGoogleGeminiUsage(tunnelRequest.Deployment.Provider, completion.Usage)
	logGoogleGeminiUsageRepair(h.logger, tunnelRequest.RequestID, usageRepair)

	providerAttrs["response_payload"] = buildProviderTraceChatResponsePayload(tunnelRequest.Deployment.Provider, completion)
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, response.StatusCode, providerAttrs)

	billingModel := billingModelForDeployment(request.Model, tunnelRequest.Deployment)

	finalCost, err := h.billing.FinalizeFromUsage(c.Context(), billingModel, completion.Usage, estimated)
	if err != nil {
		return chatProviderResult{handled: true}, writeAPIError(c, fiber.StatusBadGateway, "Failed to finalize usage", "billing_error")
	}

	tunnelRequest.RequestMetadata = withXAIUsageMetadata(tunnelRequest.RequestMetadata, tunnelRequest.Deployment.Provider, completion.Usage)

	err = h.pipeline.Settle(c.Context(), SettlementRequest{
		Reservation: reservation,
		TenantID:    authInfo.TenantID,
		Actual:      finalCost,
		RequestID:   tunnelRequest.RequestID,
		BillingMode: deploymentBillingMode(tunnelRequest.Deployment),
	})
	if err != nil {
		return chatProviderResult{handled: true}, writeBudgetCheckError(c, err)
	}

	return chatProviderResult{
		tunnelRequest: tunnelRequest,
		response:      response,
		completion:    completion,
		finalCost:     finalCost,
		usageRepair:   usageRepair,
	}, nil
}

func (h *ChatCompletionHandler) dispatchChatProvider(
	c *fiber.Ctx,
	tunnelRequest platform.TunnelRequest,
) (*platform.TunnelResponse, error) {
	if response, ok := c.Locals("vm_resp").(*platform.TunnelResponse); ok && response != nil {
		return response, nil
	}

	return h.pipeline.Dispatch(c.Context(), tunnelRequest)
}

func (h *ChatCompletionHandler) applyChatOutputGuardrails(
	c *fiber.Ctx,
	request contractsopenai.ChatCompletionRequest,
	originalBody []byte,
	authInfo *platform.AuthContext,
	tunnelRequest platform.TunnelRequest,
	selection guardrails.RequestSelection,
	completion contractsopenai.ChatCompletionResponse,
	finalCost platform.UsageCostBreakdown,
	startedAt time.Time,
) (contractsopenai.ChatCompletionResponse, bool, error) {
	if h.guardrails == nil || len(selection.LLMOutput) == 0 {
		return completion, false, nil
	}

	guardrailOverrides := map[string]guardrails.Guardrail(nil)
	if overrides, ok := c.Locals("guardrail_overrides").(map[string]guardrails.Guardrail); ok {
		guardrailOverrides = overrides
	}

	outputResult, err := h.guardrails.ApplyOutputWithGuardrails(completion, selection, guardrailOverrides)
	h.recordGuardrailChecks(c, outputResult.Checks)

	if err != nil {
		h.observeBlockedChat(c, request, originalBody, authInfo, tunnelRequest, finalCost, startedAt)

		return contractsopenai.ChatCompletionResponse{}, true, writeGuardrailError(c, fiber.StatusBadRequest, "Response blocked by GodarAI guardrail", outputResult.Checks)
	}

	if outputResult.Response != nil {
		completion = *outputResult.Response
	}

	return completion, false, nil
}

func (h *ChatCompletionHandler) observeBlockedChat(
	c *fiber.Ctx,
	request contractsopenai.ChatCompletionRequest,
	originalBody []byte,
	authInfo *platform.AuthContext,
	tunnelRequest platform.TunnelRequest,
	finalCost platform.UsageCostBreakdown,
	startedAt time.Time,
) {
	responseTimeMS := int(time.Since(startedAt).Milliseconds())
	logEntry := buildRequestLogEntry(c, authInfo, tunnelRequest.Deployment, request.Model, "/v1/chat/completions", "chat_completions", "blocked", fiber.StatusBadRequest, startedAt, responseTimeMS)
	logEntry.RequestID = tunnelRequest.RequestID
	logEntry.TraceID = tunnelRequest.TraceID
	logEntry.CostUSD = finalCost.TotalCostUSD
	logEntry.CostIRR = finalCost.TotalCostIRR
	logEntry.InputTokens = finalCost.InputTokens
	logEntry.OutputTokens = finalCost.OutputTokens
	logEntry.TotalTokens = finalCost.TotalTokens
	logEntry.RequestBody = buildProviderChatRequestLogPayload(tunnelRequest.Deployment.Provider, request, originalBody)
	logEntry.ResponseBody = map[string]interface{}{"guardrail_status": "blocked"}
	logEntry.ErrorType = "guardrail_blocked"
	logRequestBestEffort(h.logger, h.logs, c.Context(), logEntry)
	logSpendBestEffort(h.logger, h.logs, c.Context(), platform.SpendLog{
		RequestID:       tunnelRequest.RequestID,
		TokenHash:       authInfo.TokenHash,
		TenantID:        authInfo.TenantID,
		OrganizationID:  authInfo.OrganizationID,
		TeamID:          authInfo.TeamID,
		Model:           request.Model,
		InputTokens:     finalCost.InputTokens,
		OutputTokens:    finalCost.OutputTokens,
		TotalTokens:     finalCost.TotalTokens,
		CostUSD:         finalCost.TotalCostUSD,
		CostIRR:         finalCost.TotalCostIRR,
		ResponseTimeMS:  responseTimeMS,
		Status:          "blocked",
		RequestMetadata: enrichPricingMetadata(tunnelRequest.RequestMetadata, authInfo, request.Model),
		UsageBreakdown:  &finalCost,
		PricingSnapshot: finalCost.PricingSnapshot,
		CompletionStart: startedAt,
		CompletionEnd:   time.Now(),
	})
}

func (h *ChatCompletionHandler) observeSuccessfulChat(
	c *fiber.Ctx,
	request contractsopenai.ChatCompletionRequest,
	originalBody []byte,
	authInfo *platform.AuthContext,
	tunnelRequest platform.TunnelRequest,
	result chatProviderResult,
	startedAt time.Time,
	tenantSlug string,
) {
	responseTimeMS := int(time.Since(startedAt).Milliseconds())
	logEntry := buildRequestLogEntry(c, authInfo, tunnelRequest.Deployment, request.Model, "/v1/chat/completions", "chat_completions", "success", result.response.StatusCode, startedAt, responseTimeMS)
	logEntry.RequestID = tunnelRequest.RequestID
	logEntry.TraceID = tunnelRequest.TraceID
	logEntry.CostUSD = result.finalCost.TotalCostUSD
	logEntry.CostIRR = result.finalCost.TotalCostIRR
	logEntry.InputTokens = result.finalCost.InputTokens
	logEntry.OutputTokens = result.finalCost.OutputTokens
	logEntry.TotalTokens = result.finalCost.TotalTokens
	logEntry.RequestBody = buildProviderChatRequestLogPayload(tunnelRequest.Deployment.Provider, request, originalBody)

	if result.usageRepair.applied {
		logEntry.ResponseBody = buildProviderChatResponseLogPayload(tunnelRequest.Deployment.Provider, nil, result.completion)
	} else {
		logEntry.ResponseBody = buildProviderChatResponseLogPayload(tunnelRequest.Deployment.Provider, result.response.Body, result.completion)
	}

	logRequestBestEffort(h.logger, h.logs, c.Context(), logEntry)
	logSpendBestEffort(h.logger, h.logs, c.Context(), platform.SpendLog{
		RequestID:       tunnelRequest.RequestID,
		TokenHash:       authInfo.TokenHash,
		TenantID:        authInfo.TenantID,
		OrganizationID:  authInfo.OrganizationID,
		TeamID:          authInfo.TeamID,
		Model:           request.Model,
		InputTokens:     result.finalCost.InputTokens,
		OutputTokens:    result.finalCost.OutputTokens,
		TotalTokens:     result.finalCost.TotalTokens,
		CostUSD:         result.finalCost.TotalCostUSD,
		CostIRR:         result.finalCost.TotalCostIRR,
		ResponseTimeMS:  responseTimeMS,
		Status:          "success",
		RequestMetadata: enrichPricingMetadata(tunnelRequest.RequestMetadata, authInfo, request.Model),
		UsageBreakdown:  &result.finalCost,
		PricingSnapshot: result.finalCost.PricingSnapshot,
		CompletionStart: startedAt,
		CompletionEnd:   time.Now(),
	})
	h.emitModelMetric(metricsobs.BuildInput{
		TS:                 time.Now().UTC(),
		TenantID:           authInfo.TenantID,
		RequestID:          tunnelRequest.RequestID,
		ModelName:          request.Model,
		VirtualModelName:   modelMetricVirtualName(c, request.Model),
		ProviderName:       tunnelRequest.Deployment.Provider,
		ProviderModelName:  tunnelRequest.Deployment.Model,
		RequestType:        "chat_completions",
		UserID:             authInfo.UserID,
		VirtualAccountID:   authInfo.KeyID,
		VirtualAccountName: authInfo.KeyAlias,
		TeamIDs:            nonEmptySlice([]string{authInfo.TeamID}),
		TeamNames:          []string{},
		StatusCode:         result.response.StatusCode,
		LatencyMS:          responseTimeMS,
		InputTokens:        result.finalCost.InputTokens,
		OutputTokens:       result.finalCost.OutputTokens,
		CostUSD:            result.finalCost.TotalCostUSD,
		CacheMode:          cacheMetricString(c, "response_cache_type"),
		CacheStatus:        cacheMetricString(c, "response_cache_status"),
		Metadata: enrichMetricMetadataForDeployment(copyPlaygroundObservability(enrichResponseCacheMetadata(c, map[string]interface{}{
			"endpoint":    "/v1/chat/completions",
			"key_type":    authInfo.KeyType,
			"team_id":     authInfo.TeamID,
			"org_unit_id": authInfo.OrgUnitID,
			"tenant_slug": tenantSlug,
		}), tunnelRequest.RequestMetadata), tunnelRequest.Deployment),
	})
}
