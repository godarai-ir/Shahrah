package billing

import (
	"testing"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestXAIImageCostBreakdownSelectsResolutionAndReferenceRates(t *testing.T) {
	rate := func(dimension, value string, condition map[string]interface{}) normalizedPricingRate {
		return normalizedPricingRate{Dimension: dimension, Unit: "image", UnitPriceUSD: decimal.RequireFromString(value), Condition: condition}
	}
	input := rate("input_images", "0.01", nil)
	oneK := rate("output_images", "0.05", map[string]interface{}{"resolution": "1k"})
	twoK := rate("output_images", "0.07", map[string]interface{}{"resolution": "2k"})
	price := priceInfo{
		Provider: "xai",
		Rates:    map[string]normalizedPricingRate{"input_images": input},
		RateList: []normalizedPricingRate{input, oneK, twoK},
	}

	got, err := xAIImageCostBreakdown("grok-imagine-image-quality", 0, 0, 0, price, 1, map[string]interface{}{
		"image_count": 2,
		"n":           3,
		"resolution":  "2k",
	})
	require.NoError(t, err)
	require.InDelta(t, 0.23, got.TotalCostUSD, 1e-12)
	require.InDelta(t, 0.02, got.InputCostUSD, 1e-12)
	require.InDelta(t, 0.21, got.OutputCostUSD, 1e-12)
	require.Equal(t, "3", got.Quantities["output_images"])
}

func TestReconcileXAIProviderCostUsesExactTicks(t *testing.T) {
	breakdown, err := reconcileXAIProviderCost(
		xAITextEstimateForTest(),
		123456789,
		850000,
	)
	require.NoError(t, err)
	require.InDelta(t, 0.0123456789, breakdown.TotalCostUSD, 1e-14)
	require.EqualValues(t, int64(123456789), breakdown.Metadata["provider_cost_in_usd_ticks"])
	require.Equal(t, true, breakdown.Metadata["provider_cost_exact"])
	require.Equal(t, "0.0123456789", breakdown.Metadata["provider_cost_usd"])
	require.Equal(t, "0.0120456789", breakdown.Metadata["provider_cost_reconciliation_delta_usd"])
	require.Equal(t, "provider_cost_reconciliation", breakdown.Components[len(breakdown.Components)-1].Dimension)
}

func TestReconcileXAIProviderCostRejectsMissingTicksForImageUsage(t *testing.T) {
	_, err := reconcileXAIProviderCost(platform.UsageCostBreakdown{
		Model: "grok-imagine-image-quality", TotalTokens: 0, TotalCostUSD: 0.05,
	}, 0, 850000)
	require.ErrorIs(t, err, ErrPricingUnavailable)
}

func TestXAITextPricingUses200KLongContextThreshold(t *testing.T) {
	price := xAITextPriceForTest()
	usage := &contractsopenai.Usage{PromptTokens: 200000, CompletionTokens: 10, TotalTokens: 200010}
	breakdown, err := openAIUsageCostBreakdown("grok-4.5", usage, price, 1, nil)
	require.NoError(t, err)
	require.Equal(t, "long_context_input_tokens", breakdown.Components[0].Dimension)
	require.InDelta(t, 0.80012, breakdown.TotalCostUSD, 1e-12)
}

func TestXAITextPricingSupportsPriorityLongContext(t *testing.T) {
	price := xAITextPriceForTest()
	usage := &contractsopenai.Usage{
		PromptTokens: 200000, CompletionTokens: 10, TotalTokens: 200010,
		ServiceTier: "priority",
	}
	breakdown, err := openAIUsageCostBreakdown("grok-4.5", usage, price, 1, nil)
	require.NoError(t, err)
	require.Equal(t, "speed_long_context_input_tokens", breakdown.Components[0].Dimension)
	require.InDelta(t, 1.60024, breakdown.TotalCostUSD, 1e-12)
}

func xAITextPriceForTest() priceInfo {
	rate := func(dimension, value string, condition map[string]interface{}) normalizedPricingRate {
		return normalizedPricingRate{Dimension: dimension, Unit: "token", UnitPriceUSD: decimal.RequireFromString(value), Condition: condition}
	}
	return priceInfo{Provider: "xai", RateList: []normalizedPricingRate{
		rate("input_tokens", "0.000002", nil),
		rate("output_tokens", "0.000006", nil),
		rate("long_context_input_tokens", "0.000004", map[string]interface{}{"min_prompt_tokens": 200000}),
		rate("long_context_output_tokens", "0.000012", map[string]interface{}{"min_prompt_tokens": 200000}),
		rate("speed_input_tokens", "0.000004", map[string]interface{}{"service_tier": "priority"}),
		rate("speed_output_tokens", "0.000012", map[string]interface{}{"service_tier": "priority"}),
		rate("speed_long_context_input_tokens", "0.000008", map[string]interface{}{"service_tier": "priority", "min_prompt_tokens": 200000}),
		rate("speed_long_context_output_tokens", "0.000024", map[string]interface{}{"service_tier": "priority", "min_prompt_tokens": 200000}),
	}}
}

func xAITextEstimateForTest() platform.UsageCostBreakdown {
	return platform.UsageCostBreakdown{
		InputTokens: 10, OutputTokens: 5, TotalTokens: 15,
		InputCostUSD: 0.0001, OutputCostUSD: 0.0002, TotalCostUSD: 0.0003,
	}
}
