package billing

import (
	"testing"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestZAIPricingUsesPublishedCacheAndTokenDimensions(t *testing.T) {
	price := priceInfo{Revision: "zai-test", Rates: map[string]normalizedPricingRate{
		"input_tokens":      zaiRate("input_tokens", "0.0000014"),
		"cache_read_tokens": zaiRate("cache_read_tokens", "0.00000026"),
		"output_tokens":     zaiRate("output_tokens", "0.0000044"),
	}}
	usage := &contractsopenai.Usage{
		PromptTokens: 10, CompletionTokens: 2, TotalTokens: 12,
		PromptTokensDetails: &contractsopenai.UsageTokenDetails{CachedTokens: 4},
	}

	breakdown, err := (zaiPricingStrategy{}).FinalizeUsage(pricingUsageInput{
		model: "zai:acct:glm-5.2", usage: usage, price: price, exchangeRate: 1,
	})
	require.NoError(t, err)
	require.Equal(t, 4, breakdown.CacheReadInputTokens)
	require.Len(t, breakdown.Components, 3)
}

func TestZAIPricingUsesGLM53FlashConfirmedRates(t *testing.T) {
	price := priceInfo{Revision: "zai-test", Rates: map[string]normalizedPricingRate{
		"input_tokens":      zaiRate("input_tokens", "0.00000015"),
		"cache_read_tokens": zaiRate("cache_read_tokens", "0.00000003"),
		"output_tokens":     zaiRate("output_tokens", "0.0000005"),
	}}
	usage := &contractsopenai.Usage{
		PromptTokens: 1000, CompletionTokens: 200, TotalTokens: 1200,
		PromptTokensDetails: &contractsopenai.UsageTokenDetails{CachedTokens: 300},
	}

	breakdown, err := (zaiPricingStrategy{}).FinalizeUsage(pricingUsageInput{
		model: "zai:acct:glm-5.3-flash", usage: usage, price: price, exchangeRate: 1,
	})
	require.NoError(t, err)
	require.Equal(t, 300, breakdown.CacheReadInputTokens)
	require.Len(t, breakdown.Components, 3)
	require.Greater(t, breakdown.TotalCostUSD, 0.0)
}

func TestZAIPricingUsesGLM53ConfirmedRates(t *testing.T) {
	price := priceInfo{Revision: "zai-test", Rates: map[string]normalizedPricingRate{
		"input_tokens":      zaiRate("input_tokens", "0.0000014"),
		"cache_read_tokens": zaiRate("cache_read_tokens", "0.00000026"),
		"output_tokens":     zaiRate("output_tokens", "0.0000044"),
	}}
	usage := &contractsopenai.Usage{
		PromptTokens: 1000, CompletionTokens: 200, TotalTokens: 1200,
		PromptTokensDetails: &contractsopenai.UsageTokenDetails{CachedTokens: 300},
	}

	breakdown, err := (zaiPricingStrategy{}).FinalizeUsage(pricingUsageInput{
		model: "zai:acct:glm-5.3", usage: usage, price: price, exchangeRate: 1,
	})
	require.NoError(t, err)
	require.Equal(t, 300, breakdown.CacheReadInputTokens)
	require.Len(t, breakdown.Components, 3)
	require.Greater(t, breakdown.TotalCostUSD, 0.0)
}

func TestZAIPricingCompletenessRequiresCacheOnlyForChat(t *testing.T) {
	rates := map[string]normalizedPricingRate{
		"input_tokens":  zaiRate("input_tokens", "0.00000003"),
		"output_tokens": zaiRate("output_tokens", "0.00000003"),
	}

	require.True(t, zaiPricingDimensionsComplete(
		platform.ModelRuntimeProfile{Endpoints: []string{"layout_parsing"}},
		rates,
	))
	require.False(t, zaiPricingDimensionsComplete(
		platform.ModelRuntimeProfile{Endpoints: []string{"chat_completions"}},
		rates,
	))
}

func zaiRate(dimension, value string) normalizedPricingRate {
	return normalizedPricingRate{
		Dimension: dimension, Unit: "token", UnitPriceUSD: decimal.RequireFromString(value),
		SourceURL: "https://docs.z.ai/guides/overview/pricing",
	}
}
