package billing

import (
	"testing"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestCatalogPricingProviderRecognizesDeepSeekIdentity(t *testing.T) {
	for _, model := range []string{
		"deepseek:primary:deepseek-v4-flash",
		"deepseek/deepseek-v4-pro",
	} {
		require.Equal(t, "deepseek", catalogPricingProvider(model))
	}
}

func TestDeepSeekFlashUsagePricingUsesExactCacheDimensions(t *testing.T) {
	price := deepSeekPriceForTest("0.00000014", "0.0000000028", "0.00000028")
	usage := &contractsopenai.Usage{
		PromptTokens: 1000, CompletionTokens: 200, TotalTokens: 1200,
		PromptCacheHitTokens: 400, PromptCacheMissTokens: 600,
	}

	breakdown, err := deepSeekUsageCostBreakdown("deepseek-v4-flash", usage, price, 1)
	require.NoError(t, err)
	require.Equal(t, 400, breakdown.CacheReadInputTokens)
	require.InDelta(t, 0.00014112, breakdown.TotalCostUSD, 1e-15)
	require.Len(t, breakdown.Components, 3)
}

func TestDeepSeekProUsagePricingUsesExactCacheDimensions(t *testing.T) {
	price := deepSeekPriceForTest("0.000000435", "0.000000003625", "0.00000087")
	usage := &contractsopenai.Usage{
		PromptTokens: 1000, CompletionTokens: 200, TotalTokens: 1200,
		PromptCacheHitTokens: 400, PromptCacheMissTokens: 600,
	}

	breakdown, err := deepSeekUsageCostBreakdown("deepseek-v4-pro", usage, price, 1)
	require.NoError(t, err)
	require.InDelta(t, 0.00043645, breakdown.TotalCostUSD, 1e-15)
}

func TestDeepSeekUsagePricingFailsClosedOnCacheMismatch(t *testing.T) {
	price := deepSeekPriceForTest("0.00000014", "0.0000000028", "0.00000028")
	usage := &contractsopenai.Usage{
		PromptTokens: 10, PromptCacheHitTokens: 3, PromptCacheMissTokens: 6,
	}

	_, err := deepSeekUsageCostBreakdown("deepseek-v4-flash", usage, price, 1)
	require.ErrorIs(t, err, ErrPricingUnavailable)
}

func TestDeepSeekPricingDimensionsRejectZeroAndConditionalRates(t *testing.T) {
	rates := map[string]normalizedPricingRate{
		"input_tokens":      {Dimension: "input_tokens", Unit: "token", UnitPriceUSD: decimal.RequireFromString("0.00000014"), SourceURL: "https://api-docs.deepseek.com/quick_start/pricing"},
		"cache_read_tokens": {Dimension: "cache_read_tokens", Unit: "token", UnitPriceUSD: decimal.RequireFromString("0.0000000028"), SourceURL: "https://api-docs.deepseek.com/quick_start/pricing"},
		"output_tokens":     {Dimension: "output_tokens", Unit: "token", UnitPriceUSD: decimal.RequireFromString("0.00000028"), SourceURL: "https://api-docs.deepseek.com/quick_start/pricing"},
	}
	require.True(t, deepSeekPricingDimensionsComplete(rates))

	rate := rates["cache_read_tokens"]
	rate.UnitPriceUSD = decimal.Zero
	rates["cache_read_tokens"] = rate
	require.False(t, deepSeekPricingDimensionsComplete(rates))

	rates = deepSeekPriceForTest("0.00000014", "0.0000000028", "0.00000028").Rates
	rate = rates["input_tokens"]
	rate.Condition = map[string]interface{}{
		"time_window": "09:00-12:00",
		"timezone":    "Asia/Shanghai",
		"multiplier":  "2",
	}
	rates["input_tokens"] = rate
	require.False(t, deepSeekPricingDimensionsComplete(rates), "unannounced peak pricing must remain inactive")
}

func deepSeekPriceForTest(input, cached, output string) priceInfo {
	source := "https://api-docs.deepseek.com/quick_start/pricing"
	rates := map[string]normalizedPricingRate{
		"input_tokens":      {Dimension: "input_tokens", Unit: "token", UnitPriceUSD: decimal.RequireFromString(input), SourceURL: source},
		"cache_read_tokens": {Dimension: "cache_read_tokens", Unit: "token", UnitPriceUSD: decimal.RequireFromString(cached), SourceURL: source},
		"output_tokens":     {Dimension: "output_tokens", Unit: "token", UnitPriceUSD: decimal.RequireFromString(output), SourceURL: source},
	}

	return priceInfo{Provider: "deepseek", Rates: rates}
}
