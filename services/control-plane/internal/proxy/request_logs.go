package proxy

import (
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/gofiber/fiber/v2"
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

	return entry
}

func traceIDFromFiber(c *fiber.Ctx) string {
	if c == nil {
		return ""
	}
	return strings.TrimSpace(c.GetRespHeader("x-shahrah-trace-id"))
}

func fallbackStatusFromReason(reason string, count int) string {
	if strings.TrimSpace(reason) != "" || count > 0 {
		return "active"
	}
	return "inactive"
}
