package billing

import (
	"testing"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	typedmetadata "github.com/Delpak/shahrah/services/control-plane/internal/metadata"
	"github.com/Delpak/shahrah/services/control-plane/internal/providerregistry"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func tokenBazaarImageRate(dimension, value string) normalizedPricingRate {
	return normalizedPricingRate{
		Dimension:    dimension,
		Unit:         "token",
		UnitPriceUSD: decimal.RequireFromString(value),
	}
}

// TestTokenBazaarPricingStrategyRoutesGPTImage2ToOpenAIExactDecimalBreakdown
// exercises billing for gpt-image-2 via TokenBazaar using the exact,
// already-pinned official OpenAI per-token/per-image-token rates from
// apps/model-discovery/providers/tokenbazaar/gpt-image-2.yaml
// (input_cost_per_token=0.000005, input_cost_per_image_token=0.000008,
// output_cost_per_token=0.00003, cache_read_input_token_cost=0.00000125,
// cache_read_input_image_token_cost=0.000002), asserting an exact decimal
// result (computed independently with shopspring/decimal, not a float
// epsilon comparison).
func TestTokenBazaarPricingStrategyRoutesGPTImage2ToOpenAIExactDecimalBreakdown(t *testing.T) {
	price := priceInfo{
		Provider: "tokenbazaar",
		Rates: map[string]normalizedPricingRate{
			"input_tokens":            tokenBazaarImageRate("input_tokens", "0.000005"),
			"image_input_tokens":      tokenBazaarImageRate("image_input_tokens", "0.000008"),
			"cache_read_tokens":       tokenBazaarImageRate("cache_read_tokens", "0.00000125"),
			"image_cache_read_tokens": tokenBazaarImageRate("image_cache_read_tokens", "0.000002"),
			"image_output_tokens":     tokenBazaarImageRate("image_output_tokens", "0.00003"),
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

	strategy := pricingStrategy("tokenbazaar")

	got, err := strategy.FinalizeImage(pricingImageInput{
		pricingUsageInput: pricingUsageInput{
			model:        "gpt-image-2",
			usage:        usage,
			price:        price,
			exchangeRate: 1,
			metadata:     typedmetadata.Wrap(nil),
		},
		inputTokens:  usage.PromptTokens,
		outputTokens: usage.CompletionTokens,
		totalTokens:  usage.TotalTokens,
	})
	require.NoError(t, err)

	// Independently computed exact-decimal expectation:
	//   input_tokens      = 110 - 10 = 100  @ 0.000005  = 0.0005
	//   image_input_tokens= 220 - 20 = 200  @ 0.000008  = 0.0016
	//   cache_read_tokens = 10             @ 0.00000125 = 0.0000125
	//   image_cache_read  = 20             @ 0.000002   = 0.00004
	//   image_output      = 300            @ 0.00003    = 0.009
	// total = 0.0111525
	expected := decimal.RequireFromString("0.0005").
		Add(decimal.RequireFromString("0.0016")).
		Add(decimal.RequireFromString("0.0000125")).
		Add(decimal.RequireFromString("0.00004")).
		Add(decimal.RequireFromString("0.009"))

	expectedFloat, _ := expected.Float64()
	require.InDelta(t, expectedFloat, got.TotalCostUSD, 1e-12)
	require.Equal(t, "gpt-image-2", got.Model)
}

// TestTokenBazaarPricingStrategyRoutesGeminiImageToGoogleGeminiExactDecimalBreakdown
// exercises billing for gemini-3-pro-image via TokenBazaar using the exact,
// already-pinned official Google Gemini per-token and per-image-size-tier
// rates from apps/model-discovery/providers/tokenbazaar/gemini-3-pro-image.yaml
// (input_cost_per_token=0.000002, output_cost_per_image_1k=0.134), for the
// "1k" size tier.
func TestTokenBazaarPricingStrategyRoutesGeminiImageToGoogleGeminiExactDecimalBreakdown(t *testing.T) {
	price := priceInfo{
		Provider: "tokenbazaar",
		Rates: map[string]normalizedPricingRate{
			"input_tokens":     tokenBazaarImageRate("input_tokens", "0.000002"),
			"output_images_1k": tokenBazaarImageRate("output_images_1k", "0.134"),
		},
	}

	metadata := typedmetadata.Wrap(nil)
	metadata.Set(billingMetadataSize, "1024x1024")

	strategy := pricingStrategy("tokenbazaar")

	got, err := strategy.FinalizeImage(pricingImageInput{
		pricingUsageInput: pricingUsageInput{
			model:        "gemini-3-pro-image",
			price:        price,
			exchangeRate: 1,
			metadata:     metadata,
		},
		inputTokens:  500,
		outputTokens: 0,
		totalTokens:  500,
	})
	require.NoError(t, err)

	// Independently computed exact-decimal expectation:
	//   input_tokens = 500 @ 0.000002 = 0.001
	//   output_images_1k = 1 image @ 0.134 = 0.134
	// total = 0.135
	expected := decimal.RequireFromString("0.001").Add(decimal.RequireFromString("0.134"))
	expectedFloat, _ := expected.Float64()

	require.InDelta(t, expectedFloat, got.TotalCostUSD, 1e-12)
	require.Equal(t, "gemini-3-pro-image", got.Model)
}

// TestTokenBazaarImageModelFamilyClassifiesAllFourInScopeModels asserts each
// of the 4 in-scope image models is classified to the correct direct-
// provider cost-dimension family, and that a non-image TokenBazaar model
// (or an unrelated model id) is not misclassified into either family.
func TestTokenBazaarImageModelFamilyClassifiesAllFourInScopeModels(t *testing.T) {
	tests := map[string]string{
		"gpt-image-2":                 "openai",
		"gemini-3-pro-image":          "google-gemini",
		"gemini-3.1-flash-image":      "google-gemini",
		"gemini-3.1-flash-lite-image": "google-gemini",
		"claude-sonnet-5":             "",
		"text-embedding-3-large":      "",
	}

	for model, want := range tests {
		if got := tokenBazaarImageModelFamily(model); got != want {
			t.Errorf("tokenBazaarImageModelFamily(%q) = %q, want %q", model, got, want)
		}
	}
}

// TestTokenBazaarPricingStrategyIsRegisteredForTokenBazaarProvider is a
// wiring sanity check: providerregistry.Lookup("tokenbazaar") must resolve
// to PricingStrategyTokenBazaar, which is the strategy this file's other
// tests exercise via pricingStrategy("tokenbazaar").
func TestTokenBazaarPricingStrategyIsRegisteredForTokenBazaarProvider(t *testing.T) {
	definition, ok := providerregistry.Lookup("tokenbazaar")
	require.True(t, ok)
	require.Equal(t, providerregistry.PricingStrategyTokenBazaar, definition.PricingStrategy)

	_, ok = pricingStrategy("tokenbazaar").(tokenBazaarPricingStrategy)
	require.True(t, ok, "pricingStrategy(\"tokenbazaar\") must resolve to tokenBazaarPricingStrategy")
}
