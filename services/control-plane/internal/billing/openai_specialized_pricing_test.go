package billing

import (
	"testing"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestOpenAIImageCostBreakdownPricesModalitiesAndCachesSeparately(t *testing.T) {
	rate := func(dimension, value string) normalizedPricingRate {
		return normalizedPricingRate{
			Dimension:    dimension,
			Unit:         "token",
			UnitPriceUSD: decimal.RequireFromString(value),
		}
	}
	price := priceInfo{
		Provider: "openai",
		Rates: map[string]normalizedPricingRate{
			"input_tokens":            rate("input_tokens", "0.000005"),
			"image_input_tokens":      rate("image_input_tokens", "0.000008"),
			"cache_read_tokens":       rate("cache_read_tokens", "0.00000125"),
			"image_cache_read_tokens": rate("image_cache_read_tokens", "0.000002"),
			"image_output_tokens":     rate("image_output_tokens", "0.00003"),
		},
	}
	usage := &contractsopenai.Usage{
		PromptTokens:     330,
		CompletionTokens: 300,
		TotalTokens:      630,
		PromptTokensDetails: &contractsopenai.UsageTokenDetails{
			TextTokens:        110,
			ImageTokens:       220,
			CachedTextTokens:  10,
			CachedImageTokens: 20,
		},
	}

	got, err := openAIImageCostBreakdown(
		"gpt-image-2",
		usage,
		usage.PromptTokens,
		usage.CompletionTokens,
		usage.TotalTokens,
		price,
		1,
		nil,
		"standard",
	)
	require.NoError(t, err)
	require.InDelta(t, 0.0111525, got.TotalCostUSD, 0.0000000001)
	require.Len(t, got.Components, 5)
}

func TestOpenAIImageCostBreakdownSelectsBatchDimensions(t *testing.T) {
	rate := func(dimension, value string) normalizedPricingRate {
		return normalizedPricingRate{
			Dimension: dimension, Unit: "token",
			UnitPriceUSD: decimal.RequireFromString(value),
		}
	}
	price := priceInfo{
		Provider: "openai",
		Rates: map[string]normalizedPricingRate{
			"batch_input_tokens":            rate("batch_input_tokens", "0.0000025"),
			"batch_image_input_tokens":      rate("batch_image_input_tokens", "0.000004"),
			"batch_cache_read_tokens":       rate("batch_cache_read_tokens", "0.000000625"),
			"batch_image_cache_read_tokens": rate("batch_image_cache_read_tokens", "0.000001"),
			"batch_image_output_tokens":     rate("batch_image_output_tokens", "0.000015"),
		},
	}
	usage := &contractsopenai.Usage{
		PromptTokens: 330, CompletionTokens: 300, TotalTokens: 630,
		PromptTokensDetails: &contractsopenai.UsageTokenDetails{
			TextTokens: 110, ImageTokens: 220,
			CachedTextTokens: 10, CachedImageTokens: 20,
		},
	}

	got, err := openAIImageCostBreakdown(
		"gpt-image-2", usage, 330, 300, 630, price, 1, nil, "batch",
	)
	require.NoError(t, err)
	require.InDelta(t, 0.00557625, got.TotalCostUSD, 0.0000000001)
	require.Equal(t, "batch", got.ServiceTier)
	require.Len(t, got.Components, 5)
}
