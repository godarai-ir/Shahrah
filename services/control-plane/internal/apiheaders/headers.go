// Package apiheaders centralizes control-plane header names and lookup helpers.
package apiheaders

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// HTTP header names used by the control plane.
const (
	HeaderTraceID                 = "x-godarai-trace-id"
	HeaderTenant                  = "x-godarai-tenant"
	HeaderCacheConfig             = "x-godarai-cache-config"
	HeaderCacheStatus             = "x-godarai-cache-status"
	HeaderCachedTraceID           = "x-godarai-cached-trace-id"
	HeaderCacheSimilarity         = "x-godarai-cache-similarity-score"
	HeaderCostUSD                 = "x-godarai-cost-usd"
	HeaderCostIRR                 = "x-godarai-cost-irr"
	HeaderCreditsUsed             = "x-godarai-credits-used"
	HeaderStreamTelemetry         = "x-godarai-include-stream-telemetry"
	HeaderPlaygroundObservability = "x-godarai-playground-observability"
	HeaderGuardrails              = "x-godarai-guardrails"
	HeaderGuardrailsScope         = "x-godarai-guardrails-scope"
	HeaderGuardrailReview         = "x-godarai-guardrail-review"
	HeaderGuardrailChecks         = "x-godarai-guardrail-checks"
	HeaderUserID                  = "x-godarai-user-id"
	LegacyHeaderUserID            = "x-shahrah-user-id"
)

// GetFirst returns the first non-empty header value among the provided keys.
func GetFirst(c *fiber.Ctx, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(c.Get(key)); value != "" {
			return value
		}
	}

	return ""
}

// GetTraceID returns the trace identifier header value.
func GetTraceID(c *fiber.Ctx) string {
	return GetFirst(c, HeaderTraceID, "x-trace-id")
}

// GetTenant returns the tenant header value.
func GetTenant(c *fiber.Ctx) string {
	return GetFirst(c, HeaderTenant)
}

// GetUserID returns the delegated tenant-admin user id.
func GetUserID(c *fiber.Ctx) string {
	return GetFirst(c, HeaderUserID, LegacyHeaderUserID)
}

// GetCacheConfig returns the cache configuration header value.
func GetCacheConfig(c *fiber.Ctx) string {
	return GetFirst(c, HeaderCacheConfig)
}

// GetPlaygroundObservability returns the internal Playground observability context.
func GetPlaygroundObservability(c *fiber.Ctx) string {
	return GetFirst(c, HeaderPlaygroundObservability)
}

// GetGuardrails returns the guardrails header value.
func GetGuardrails(c *fiber.Ctx) string {
	return GetFirst(c, HeaderGuardrails)
}

// GetGuardrailsScope returns the guardrails scope header value.
func GetGuardrailsScope(c *fiber.Ctx) string {
	return GetFirst(c, HeaderGuardrailsScope)
}
