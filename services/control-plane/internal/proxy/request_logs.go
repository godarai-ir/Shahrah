package proxy

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/apiheaders"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

const streamResponseLogCaptureLimit = 32 * 1024

type streamResponseLogCapture struct {
	data      []byte
	truncated bool
}

func (capture *streamResponseLogCapture) Append(data []byte) {
	if capture == nil || len(data) == 0 {
		return
	}

	remaining := streamResponseLogCaptureLimit - len(capture.data)
	if remaining <= 0 {
		capture.truncated = true
		return
	}

	if len(data) > remaining {
		capture.data = append(capture.data, data[:remaining]...)
		capture.truncated = true

		return
	}

	capture.data = append(capture.data, data...)
}

func (capture *streamResponseLogCapture) Payload(outcome string) map[string]interface{} {
	payload := map[string]interface{}{
		"stream":  true,
		"outcome": strings.TrimSpace(outcome),
	}

	if capture != nil && len(capture.data) > 0 {
		payload["raw_stream"] = string(capture.data)
	}

	if capture != nil && capture.truncated {
		payload["truncated"] = true
	}

	return payload
}

func logStreamingRequestBestEffort(
	logger *zap.Logger,
	logs usageLogger,
	ctx context.Context,
	base platform.RequestLogEntry,
	status string,
	statusCode int,
	errorType string,
	responseBody map[string]interface{},
	finalCost *platform.UsageCostBreakdown,
) {
	entry := base
	entry.Status = strings.TrimSpace(status)
	entry.StatusCode = statusCode
	entry.ErrorType = strings.TrimSpace(errorType)
	entry.ResponseTimeMS = int(time.Since(entry.StartedAt).Milliseconds())
	entry.ResponseBody = responseBody

	if len(base.Metadata) > 0 {
		entry.Metadata = make(map[string]interface{}, len(base.Metadata)+1)
		for key, value := range base.Metadata {
			entry.Metadata[key] = value
		}
	} else {
		entry.Metadata = make(map[string]interface{}, 1)
	}

	entry.Metadata["stream"] = true

	if finalCost != nil {
		entry.CostUSD = finalCost.TotalCostUSD
		entry.CostIRR = finalCost.TotalCostIRR
		entry.InputTokens = finalCost.InputTokens
		entry.OutputTokens = finalCost.OutputTokens
		entry.TotalTokens = finalCost.TotalTokens
	}

	logRequestBestEffort(logger, logs, ctx, entry)
}

func buildRequestLogEntry(
	c *fiber.Ctx,
	authInfo *platform.AuthContext,
	deployment platform.Deployment,
	requestedModel string,
	route string,
	requestType string,
	status string,
	statusCode int,
	start time.Time,
	durationMS int,
) platform.RequestLogEntry {
	entry := platform.RequestLogEntry{
		RequestID:      traceIDFromFiber(c),
		TraceID:        traceIDFromFiber(c),
		Model:          strings.TrimSpace(requestedModel),
		Provider:       strings.TrimSpace(deployment.Provider),
		ProviderModel:  strings.TrimSpace(deployment.Model),
		Route:          strings.TrimSpace(route),
		RequestType:    strings.TrimSpace(requestType),
		Status:         strings.TrimSpace(status),
		StatusCode:     statusCode,
		ResponseTimeMS: durationMS,
		StartedAt:      start.UTC(),
	}

	if authInfo != nil {
		entry.TokenHash = strings.TrimSpace(authInfo.TokenHash)
		entry.TenantID = strings.TrimSpace(authInfo.TenantID)
		entry.OrganizationID = strings.TrimSpace(authInfo.OrganizationID)
		entry.OrgUnitID = strings.TrimSpace(authInfo.OrgUnitID)
		entry.TeamID = strings.TrimSpace(authInfo.TeamID)
		entry.ActorID = strings.TrimSpace(authInfo.UserID)

		entry.ActorType = "user"
		if entry.ActorID == "" {
			entry.ActorID = strings.TrimSpace(authInfo.VirtualUserID)
			if entry.ActorID != "" {
				entry.ActorType = "virtual_user"
			}
		}

		if snapshot, ok := authInfo.VirtualModelSnapshots[requestedModel]; ok {
			entry.VirtualModel = strings.TrimSpace(snapshot.Alias)
		}
	}

	if c != nil {
		if value, ok := c.Locals("vm_fallback_count").(int); ok {
			entry.FallbackCount = value
		}

		if value, ok := c.Locals("vm_failure_reason").(string); ok {
			entry.FallbackStatus = fallbackStatusFromReason(value, entry.FallbackCount)
		}

		entry.CacheStatus = cacheMetricString(c, "response_cache_status")
		entry.CacheMode = cacheMetricString(c, "response_cache_type")
	}

	entry.LoadBalanceStatus = "inactive"
	if entry.FallbackStatus == "" {
		entry.FallbackStatus = fallbackStatusFromReason("", entry.FallbackCount)
	}

	if hasDeploymentBillingContext(deployment) {
		entry.Metadata = enrichDeploymentBillingMetadata(entry.Metadata, deployment)
	}

	return entry
}

func traceIDFromFiber(c *fiber.Ctx) string {
	if c == nil {
		return ""
	}

	return strings.TrimSpace(c.GetRespHeader(apiheaders.HeaderTraceID))
}

func fallbackStatusFromReason(reason string, count int) string {
	if strings.TrimSpace(reason) != "" || count > 0 {
		return "active"
	}

	return "inactive"
}

func logPreProviderRejection(
	logger *zap.Logger,
	logs usageLogger,
	c *fiber.Ctx,
	authInfo *platform.AuthContext,
	requestedModel string,
	route string,
	requestType string,
	statusCode int,
	errorType string,
	message string,
	reason string,
	start time.Time,
) {
	entry := buildRequestLogEntry(
		c,
		authInfo,
		platform.Deployment{},
		requestedModel,
		route,
		requestType,
		"blocked",
		statusCode,
		start,
		int(time.Since(start).Milliseconds()),
	)
	requestID := tracing.TraceIDFromContext(c)
	entry.RequestID = requestID
	entry.TraceID = requestID
	entry.ErrorType = strings.TrimSpace(errorType)
	entry.RequestBody = map[string]interface{}{"model": strings.TrimSpace(requestedModel)}
	entry.ResponseBody = map[string]interface{}{
		"error": map[string]interface{}{
			"message": strings.TrimSpace(message),
			"type":    strings.TrimSpace(errorType),
		},
	}
	entry.Metadata = map[string]interface{}{
		"denial_source":    "model_access",
		"rejection_stage":  "pre_provider",
		"rejection_reason": strings.TrimSpace(reason),
	}

	logRequestBestEffort(logger, logs, c.Context(), entry)
}

func buildJSONRequestLogPayload(body []byte, overrides map[string]interface{}) map[string]interface{} {
	payload := map[string]interface{}{}

	if err := json.Unmarshal(body, &payload); err != nil {
		payload["raw_body"] = string(body)
	}

	for key, value := range overrides {
		payload[key] = value
	}

	return payload
}

func buildJSONResponseLogPayload(body []byte, fallback interface{}) map[string]interface{} {
	payload := map[string]interface{}{}
	if err := json.Unmarshal(body, &payload); err == nil {
		return payload
	}

	if fallback == nil {
		return payload
	}

	encoded, err := json.Marshal(fallback)
	if err != nil {
		return payload
	}

	if err := json.Unmarshal(encoded, &payload); err != nil {
		return map[string]interface{}{"value": fallback}
	}

	return payload
}

// buildProviderHTTPErrorLogPayload preserves the provider response byte-for-byte
// alongside its decoded JSON fields. Error responses to streaming requests are
// not guaranteed to be JSON (some providers return SSE frames, HTML, or plain
// text), so relying on JSON decoding alone can silently discard the only useful
// diagnostic content.
func buildProviderHTTPErrorLogPayload(body []byte, stream bool) map[string]interface{} {
	payload := buildJSONResponseLogPayload(body, nil)
	payload["outcome"] = "provider_http_error"
	payload["stream"] = stream

	if len(body) > 0 {
		payload["raw_body"] = string(body)
	}

	return payload
}
