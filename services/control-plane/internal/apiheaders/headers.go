// Package apiheaders centralizes control-plane header names and lookup helpers.
package apiheaders

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// Standard and legacy HTTP header names used by the control plane.
const (
	HeaderTraceID              = "x-godarai-trace-id"
	LegacyHeaderTraceID        = "x-shahrah-trace-id"
	HeaderTenant               = "x-godarai-tenant"
	LegacyHeaderTenant         = "x-shahrah-tenant"
	HeaderCacheConfig          = "x-godarai-cache-config"
	LegacyHeaderCacheConfig    = "x-shahrah-cache-config"
	HeaderCacheStatus          = "x-godarai-cache-status"
	HeaderCachedTraceID        = "x-godarai-cached-trace-id"
	HeaderCacheSimilarity      = "x-godarai-cache-similarity-score"
	HeaderCostUSD              = "x-godarai-cost-usd"
	HeaderCostIRR              = "x-godarai-cost-irr"
	LegacyHeaderCostUSD        = "x-shahrah-cost-usd"
	LegacyHeaderCostIRR        = "x-shahrah-cost-irr"
	HeaderGuardrails           = "x-godarai-guardrails"
	LegacyHeaderGuardrails     = "x-shahrah-guardrails"
	HeaderGuardrailsScope      = "x-godarai-guardrails-scope"
	LegacyHeaderGuardrailScope = "x-shahrah-guardrails-scope"
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
	return GetFirst(c, HeaderTraceID, LegacyHeaderTraceID, "x-trace-id")
}

// GetTenant returns the tenant header value.
func GetTenant(c *fiber.Ctx) string {
	return GetFirst(c, HeaderTenant, LegacyHeaderTenant)
}

// GetCacheConfig returns the cache configuration header value.
func GetCacheConfig(c *fiber.Ctx) string {
	return GetFirst(c, HeaderCacheConfig, LegacyHeaderCacheConfig)
}

// GetGuardrails returns the guardrails header value.
func GetGuardrails(c *fiber.Ctx) string {
	return GetFirst(c, HeaderGuardrails, LegacyHeaderGuardrails, "x-tfy-guardrails")
}

// GetGuardrailsScope returns the guardrails scope header value.
func GetGuardrailsScope(c *fiber.Ctx) string {
	return GetFirst(c, HeaderGuardrailsScope, LegacyHeaderGuardrailScope, "x-tfy-guardrails-scope")
}
