package billing

import (
	"testing"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestCatalogPricingProviderRecognizesMoonshotIdentity(t *testing.T) {
	for _, model := range []string{
		"moonshot:primary:kimi-k3",
		"moonshot/kimi-k2.7-code",
		"moonshot/kimi-k2.7-code-highspeed",
		"moonshot/kimi-k2.6",
	} {
		require.Equal(t, "moonshot", catalogPricingProvider(model))
	}
}

func TestMoonshotUsagePricingSeparatesCachedInput(t *testing.T) {
	price := moonshotPriceForTest("0.000003", "0.0000003", "0.000015")
	usage := &contractsopenai.Usage{
		PromptTokens: 1000, CompletionTokens: 200, TotalTokens: 1200,
		PromptTokensDetails: &contractsopenai.UsageTokenDetails{CachedTokens: 400},
	}

	breakdown, err := moonshotUsageCostBreakdown("kimi-k3", usage, price, 1)
	require.NoError(t, err)
	require.Equal(t, 400, breakdown.CacheReadInputTokens)
	require.InDelta(t, 0.00492, breakdown.TotalCostUSD, 1e-12)
	require.Len(t, breakdown.Components, 3)
	require.Equal(t, "input_tokens", breakdown.Components[0].Dimension)
	require.Equal(t, "cache_read_tokens", breakdown.Components[1].Dimension)
	require.Equal(t, "output_tokens", breakdown.Components[2].Dimension)
}

func TestMoonshotUsagePricingRejectsInvalidCacheAccounting(t *testing.T) {
	price := moonshotPriceForTest("0.00000095", "0.00000019", "0.000004")
	usage := &contractsopenai.Usage{
		PromptTokens:        10,
		PromptTokensDetails: &contractsopenai.UsageTokenDetails{CachedTokens: 11},
	}

	_, err := moonshotUsageCostBreakdown("kimi-k2.7-code", usage, price, 1)
	require.ErrorIs(t, err, ErrPricingUnavailable)
}

func TestMoonshotPricingDimensionsFailClosed(t *testing.T) {
	input := normalizedPricingRate{Dimension: "input_tokens", Unit: "token", UnitPriceUSD: decimal.RequireFromString("0.000001")}
	output := normalizedPricingRate{Dimension: "output_tokens", Unit: "token", UnitPriceUSD: decimal.RequireFromString("0.000002")}

	require.False(t, moonshotPricingDimensionsComplete(map[string]normalizedPricingRate{
		"input_tokens": input, "output_tokens": output,
	}))

	cache := normalizedPricingRate{Dimension: "cache_read_tokens", Unit: "token", UnitPriceUSD: decimal.RequireFromString("0.0000001")}
	require.True(t, moonshotPricingDimensionsComplete(map[string]normalizedPricingRate{
		"input_tokens": input, "cache_read_tokens": cache, "output_tokens": output,
	}))
}

func moonshotPriceForTest(input, cached, output string) priceInfo {
	rates := map[string]normalizedPricingRate{
		"input_tokens": {
			Dimension: "input_tokens", Unit: "token", UnitPriceUSD: decimal.RequireFromString(input),
		},
		"cache_read_tokens": {
			Dimension: "cache_read_tokens", Unit: "token", UnitPriceUSD: decimal.RequireFromString(cached),
		},
		"output_tokens": {
			Dimension: "output_tokens", Unit: "token", UnitPriceUSD: decimal.RequireFromString(output),
		},
	}

	return priceInfo{Provider: "moonshot", Rates: rates}
}
