package billing

import (
	"testing"
	"time"

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

	breakdown, err := deepSeekUsageCostBreakdown("deepseek-v4-flash", usage, price, 1, nil)
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

	breakdown, err := deepSeekUsageCostBreakdown("deepseek-v4-pro", usage, price, 1, nil)
	require.NoError(t, err)
	require.InDelta(t, 0.00043645, breakdown.TotalCostUSD, 1e-15)
}

func TestDeepSeekUsagePricingFailsClosedOnCacheMismatch(t *testing.T) {
	price := deepSeekPriceForTest("0.00000014", "0.0000000028", "0.00000028")
	usage := &contractsopenai.Usage{
		PromptTokens: 10, PromptCacheHitTokens: 3, PromptCacheMissTokens: 6,
	}

	_, err := deepSeekUsageCostBreakdown("deepseek-v4-flash", usage, price, 1, nil)
	require.ErrorIs(t, err, ErrPricingUnavailable)
}

func TestDeepSeekPricingSelectsPinnedUTCPeriod(t *testing.T) {
	price := deepSeekPriceForTest("0.00000132", "0.000000044", "0.00000396")
	for i := range price.RateList {
		rate := price.RateList[i]
		rate.UnitPriceUSD = rate.UnitPriceUSD.Div(decimal.NewFromInt(2))
		rate.Condition = map[string]interface{}{"pricing_period": "off_peak"}
		price.RateList = append(price.RateList, rate)
	}

	usage := &contractsopenai.Usage{
		PromptTokens: 1000, CompletionTokens: 200, TotalTokens: 1200,
		PromptCacheHitTokens: 400, PromptCacheMissTokens: 600,
	}

	peakConditions := deepSeekPricingConditionsAt(
		time.Date(2026, time.August, 16, 2, 0, 0, 0, time.UTC),
		price.RateList,
	)
	peak, err := deepSeekUsageCostBreakdown("deepseek-v4-pro", usage, price, 1, peakConditions)
	require.NoError(t, err)
	require.Equal(t, "peak", peak.PricingConditions["pricing_period"])

	offPeakConditions := deepSeekPricingConditionsAt(
		time.Date(2026, time.August, 16, 5, 0, 0, 0, time.UTC),
		price.RateList,
	)
	offPeak, err := deepSeekUsageCostBreakdown("deepseek-v4-pro", usage, price, 1, offPeakConditions)
	require.NoError(t, err)
	require.Equal(t, "off_peak", offPeak.PricingConditions["pricing_period"])
	require.InDelta(t, peak.TotalCostUSD/2, offPeak.TotalCostUSD, 1e-15)
}

func TestDeepSeekCatalogSnapshotAcceptsOffPeakRateOverrides(t *testing.T) {
	now := time.Date(2026, time.August, 16, 16, 0, 0, 0, time.UTC)
	profile := completeDiscoveryRuntimeProfile(
		now.Add(-time.Hour),
		"adaptive",
		[]string{"adaptive", "disabled"},
	)
	profile.Profile.Endpoints = []string{"models", "chat_completions", "responses"}

	rate := func(dimension, value string, condition map[string]interface{}) discoveryPricingRate {
		return discoveryPricingRate{
			Dimension: dimension, Unit: "token", UnitPriceUSD: value,
			Currency: "USD", Region: "global", Condition: condition,
			SourceURL: "https://api-docs.deepseek.com/quick_start/pricing",
		}
	}
	rates := []discoveryPricingRate{
		rate("input_tokens", "0.00000132", nil),
		rate("cache_read_tokens", "0.000000044", nil),
		rate("output_tokens", "0.00000396", nil),
		rate("input_tokens", "0.00000066", map[string]interface{}{"pricing_period": "off_peak"}),
		rate("cache_read_tokens", "0.000000022", map[string]interface{}{"pricing_period": "off_peak"}),
		rate("output_tokens", "0.00000198", map[string]interface{}{"pricing_period": "off_peak"}),
	}

	snapshot, err := normalizeCatalogPricingSnapshot(
		discoveryRuntimeSnapshot{
			Provider: "deepseek", Revision: "deepseek-time-pricing",
			GeneratedAt: now,
			Models: []discoveryRuntimeModel{{
				CanonicalID: "deepseek/deepseek-v4-pro", Slug: "deepseek-v4-pro",
				Status: "active", RuntimeProfile: profile, PricingRates: rates,
			}},
		},
		"deepseek",
	)
	require.NoError(t, err)
	require.Len(t, snapshot.Models["deepseek-v4-pro"].RateList, 6)
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

	rateList := make([]normalizedPricingRate, 0, len(rates))
	for _, dimension := range []string{"input_tokens", "cache_read_tokens", "output_tokens"} {
		rateList = append(rateList, rates[dimension])
	}

	return priceInfo{Provider: "deepseek", Rates: rates, RateList: rateList}
}
