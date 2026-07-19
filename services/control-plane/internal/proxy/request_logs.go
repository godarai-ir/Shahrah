package proxy

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/apiheaders"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

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
	_ = json.Unmarshal(body, &payload)
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
	_ = json.Unmarshal(encoded, &payload)
	return payload
}
