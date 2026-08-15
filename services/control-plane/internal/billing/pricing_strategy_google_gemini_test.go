package billing

import (
	"testing"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestGoogleGeminiFlatPricingKeepsBaseRatesAbove200K(t *testing.T) {
	price := googleGeminiPriceForTest(map[string]string{
		"input_tokens":  "0.0000001",
		"output_tokens": "0.0000004",
	})
	usage := &contractsopenai.Usage{
		PromptTokens: 250001, CompletionTokens: 10, TotalTokens: 250011,
	}

	breakdown, err := googleGeminiUsageCostBreakdown(
		"gemini-2.5-flash-lite",
		usage,
		price,
		1,
		nil,
	)
	require.NoError(t, err)
	require.Equal(t, "input_tokens", breakdown.Components[0].Dimension)
	require.Equal(t, "output_tokens", breakdown.Components[1].Dimension)
	require.InDelta(t, 0.0250041, breakdown.TotalCostUSD, 1e-12)
}

func TestGoogleGeminiBatchPricingUsesBatchRates(t *testing.T) {
	price := googleGeminiPriceForTest(map[string]string{
		"batch_input_tokens":  "0.00000005",
		"batch_output_tokens": "0.0000002",
	})
	usage := &contractsopenai.Usage{
		PromptTokens: 100, CompletionTokens: 20, TotalTokens: 120,
		ServiceTier: "batch",
	}

	breakdown, err := googleGeminiUsageCostBreakdown(
		"gemini-2.5-flash-lite",
		usage,
		price,
		1,
		map[string]string{"service_tier": "batch"},
	)
	require.NoError(t, err)
	require.Equal(t, "batch", breakdown.ServiceTier)
	require.Equal(t, "batch_input_tokens", breakdown.Components[0].Dimension)
	require.Equal(t, "batch_output_tokens", breakdown.Components[1].Dimension)
	require.InDelta(t, 0.000009, breakdown.TotalCostUSD, 1e-12)
}

func TestGoogleGeminiTieredPricingUsesLongContextRates(t *testing.T) {
	price := googleGeminiPriceForTest(map[string]string{
		"input_tokens":               "0.00000125",
		"output_tokens":              "0.00001",
		"long_context_input_tokens":  "0.0000025",
		"long_context_output_tokens": "0.000015",
	})
	usage := &contractsopenai.Usage{
		PromptTokens: 200001, CompletionTokens: 10, TotalTokens: 200011,
	}

	breakdown, err := googleGeminiUsageCostBreakdown(
		"gemini-2.5-pro",
		usage,
		price,
		1,
		nil,
	)
	require.NoError(t, err)
	require.Equal(t, "long_context_input_tokens", breakdown.Components[0].Dimension)
	require.Equal(t, "long_context_output_tokens", breakdown.Components[1].Dimension)
}

func TestGoogleGeminiPriorityPricingUsesAudioInputRate(t *testing.T) {
	price := googleGeminiPriceForTest(map[string]string{
		"priority_input_tokens":       "0.00000018",
		"priority_audio_input_tokens": "0.00000054",
		"priority_output_tokens":      "0.00000072",
	})
	usage := &contractsopenai.Usage{
		PromptTokens: 10, CompletionTokens: 2, TotalTokens: 12,
		ServiceTier: "priority",
		PromptTokensDetails: &contractsopenai.UsageTokenDetails{
			AudioTokens: 3,
		},
	}

	breakdown, err := googleGeminiUsageCostBreakdown(
		"gemini-2.5-flash-lite",
		usage,
		price,
		1,
		nil,
	)
	require.NoError(t, err)
	require.Equal(t, "priority_audio_input_tokens", breakdown.Components[0].Dimension)
	require.Equal(t, "3", breakdown.Components[0].Quantity)
	require.InDelta(t, 0.00000432, breakdown.TotalCostUSD, 1e-12)
}

func googleGeminiPriceForTest(values map[string]string) priceInfo {
	rates := make(map[string]normalizedPricingRate, len(values))
	for dimension, value := range values {
		rates[dimension] = normalizedPricingRate{
			Dimension:    dimension,
			Unit:         "token",
			UnitPriceUSD: decimal.RequireFromString(value),
		}
	}

	return priceInfo{Provider: "google-gemini", Rates: rates}
}
