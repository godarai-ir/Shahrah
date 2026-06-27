package apiheaders

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

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
	HeaderGuardrails           = "x-godarai-guardrails"
	LegacyHeaderGuardrails     = "x-shahrah-guardrails"
	HeaderGuardrailsScope      = "x-godarai-guardrails-scope"
	LegacyHeaderGuardrailScope = "x-shahrah-guardrails-scope"
)

func GetFirst(c *fiber.Ctx, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(c.Get(key)); value != "" {
			return value
		}
	}
	return ""
}

func GetTraceID(c *fiber.Ctx) string {
	return GetFirst(c, HeaderTraceID, LegacyHeaderTraceID, "x-trace-id")
}

func GetTenant(c *fiber.Ctx) string {
	return GetFirst(c, HeaderTenant, LegacyHeaderTenant)
}

func GetCacheConfig(c *fiber.Ctx) string {
	return GetFirst(c, HeaderCacheConfig, LegacyHeaderCacheConfig)
}

func GetGuardrails(c *fiber.Ctx) string {
	return GetFirst(c, HeaderGuardrails, LegacyHeaderGuardrails, "x-tfy-guardrails")
}

func GetGuardrailsScope(c *fiber.Ctx) string {
	return GetFirst(c, HeaderGuardrailsScope, LegacyHeaderGuardrailScope, "x-tfy-guardrails-scope")
}
