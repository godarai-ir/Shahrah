package billing

import (
	"errors"
	"testing"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

// tokenBazaarClaudePrice mirrors the pinned catalog rates in
// apps/model-discovery/providers/tokenbazaar/claude-sonnet-5.yaml
// (input_cost_per_token=0.00000285, cache_read_input_token_cost=0.000000285),
// built the way (*PricingEngine).modelPrice builds a priceInfo for a
// PricingSnapshotCatalog provider.
func tokenBazaarClaudePrice() priceInfo {
	rate := func(dimension, value string) normalizedPricingRate {
		return normalizedPricingRate{
			Dimension:    dimension,
			Unit:         "token",
			UnitPriceUSD: decimal.RequireFromString(value),
		}
	}

	input := rate("input_tokens", "0.00000285")
	output := rate("output_tokens", "0.00001425")

	return priceInfo{
		Provider:       "tokenbazaar",
		InputPerToken:  input.UnitPriceUSD,
		OutputPerToken: output.UnitPriceUSD,
		Rates: map[string]normalizedPricingRate{
			"input_tokens":      input,
			"output_tokens":     output,
			"cache_read_tokens": rate("cache_read_tokens", "0.000000285"),
		},
	}
}

// TestTokenBazaarPricingStrategyBillsAnthropicStyleCacheReadTokens pins the
// behavior that the flat inherited breakdown got wrong: on the Anthropic
// Messages envelope, usage.input_tokens EXCLUDES cached tokens, which are
// reported separately as cache_read_input_tokens. Pricing only
// prompt_tokens/completion_tokens therefore billed every cached token at zero.
func TestTokenBazaarPricingStrategyBillsAnthropicStyleCacheReadTokens(t *testing.T) {
	price := tokenBazaarClaudePrice()

	// Shape produced by proxy.parseAnthropicNativeUsage for a response with a
	// 100k-token cached system prompt: input_tokens counts only the uncached
	// remainder.
	usage := &contractsopenai.Usage{
		PromptTokens:         50,
		CompletionTokens:     500,
		CacheReadInputTokens: 100000,
	}

	got, err := pricingStrategy("tokenbazaar").FinalizeUsage(pricingUsageInput{
		model:        "claude-sonnet-5",
		usage:        usage,
		price:        price,
		exchangeRate: 1,
	})
	require.NoError(t, err)

	//   input_tokens      =     50 @ 0.00000285  = 0.0001425
	//   output_tokens     =    500 @ 0.00001425  = 0.007125
	//   cache_read_tokens = 100000 @ 0.000000285 = 0.0285
	expected := decimal.RequireFromString("0.0001425").
		Add(decimal.RequireFromString("0.007125")).
		Add(decimal.RequireFromString("0.0285"))

	expectedFloat, _ := expected.Float64()
	require.InDelta(t, expectedFloat, got.TotalCostUSD, 1e-12)
	require.Equal(t, 100000, got.CacheReadInputTokens)

	// Regression guard: the flat path billed only input+output and dropped the
	// cache-read tokens entirely.
	flat, err := standardPricingStrategy{}.FinalizeUsage(pricingUsageInput{
		model:        "claude-sonnet-5",
		usage:        usage,
		price:        price,
		exchangeRate: 1,
	})
	require.NoError(t, err)
	require.Greater(t, got.TotalCostUSD, flat.TotalCostUSD)
}

// TestTokenBazaarPricingStrategyDoesNotDoubleCountOpenAIStyleCachedTokens
// covers the other convention: the OpenAI and Gemini envelopes report cached
// tokens INSIDE prompt_tokens, so the cached portion must be subtracted from
// the input dimension before being priced at the cache-read rate.
func TestTokenBazaarPricingStrategyDoesNotDoubleCountOpenAIStyleCachedTokens(t *testing.T) {
	price := tokenBazaarClaudePrice()

	usage := &contractsopenai.Usage{
		PromptTokens:        1000,
		CompletionTokens:    500,
		PromptTokensDetails: &contractsopenai.UsageTokenDetails{CachedTokens: 800},
	}

	got, err := pricingStrategy("tokenbazaar").FinalizeUsage(pricingUsageInput{
		model:        "gpt-5.5",
		usage:        usage,
		price:        price,
		exchangeRate: 1,
	})
	require.NoError(t, err)

	//   input_tokens      = 1000 - 800 = 200 @ 0.00000285  = 0.00057
	//   cache_read_tokens =         800     @ 0.000000285 = 0.000228
	//   output_tokens     =         500     @ 0.00001425  = 0.007125
	expected := decimal.RequireFromString("0.00057").
		Add(decimal.RequireFromString("0.000228")).
		Add(decimal.RequireFromString("0.007125"))

	expectedFloat, _ := expected.Float64()
	require.InDelta(t, expectedFloat, got.TotalCostUSD, 1e-12)
}

// TestTokenBazaarPricingStrategyLeavesUncachedUsageOnStandardPath asserts the
// override is additive only: usage with no cached dimension -- every
// embeddings response, and every uncached completion -- must bill exactly as
// the inherited standard path did before.
func TestTokenBazaarPricingStrategyLeavesUncachedUsageOnStandardPath(t *testing.T) {
	price := tokenBazaarClaudePrice()

	usage := &contractsopenai.Usage{PromptTokens: 12345, CompletionTokens: 678, TotalTokens: 13023}

	input := pricingUsageInput{model: "claude-sonnet-5", usage: usage, price: price, exchangeRate: 1}

	got, err := pricingStrategy("tokenbazaar").FinalizeUsage(input)
	require.NoError(t, err)

	want, err := standardPricingStrategy{}.FinalizeUsage(input)
	require.NoError(t, err)

	require.Equal(t, want.TotalCostUSD, got.TotalCostUSD)
	require.Equal(t, want.InputCostUSD, got.InputCostUSD)
	require.Equal(t, want.OutputCostUSD, got.OutputCostUSD)
}

// TestTokenBazaarPricingStrategyFailsClosedWhenCacheReadRateMissing asserts a
// pricing gap refuses the request rather than silently billing cached tokens
// at zero again -- the same posture mistralUsageCostBreakdown takes.
func TestTokenBazaarPricingStrategyFailsClosedWhenCacheReadRateMissing(t *testing.T) {
	price := tokenBazaarClaudePrice()
	delete(price.Rates, "cache_read_tokens")

	_, err := pricingStrategy("tokenbazaar").FinalizeUsage(pricingUsageInput{
		model:        "claude-sonnet-5",
		usage:        &contractsopenai.Usage{PromptTokens: 50, CompletionTokens: 500, CacheReadInputTokens: 100000},
		price:        price,
		exchangeRate: 1,
	})
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrPricingUnavailable))
}

// TestTokenBazaarPricingStrategyRejectsCachedTokensExceedingPromptTokens
// guards the inclusive-convention arithmetic against a malformed upstream
// usage record that would otherwise produce a negative input quantity.
func TestTokenBazaarPricingStrategyRejectsCachedTokensExceedingPromptTokens(t *testing.T) {
	_, err := pricingStrategy("tokenbazaar").FinalizeUsage(pricingUsageInput{
		model: "gpt-5.5",
		usage: &contractsopenai.Usage{
			PromptTokens:        100,
			CompletionTokens:    10,
			PromptTokensDetails: &contractsopenai.UsageTokenDetails{CachedTokens: 500},
		},
		price:        tokenBazaarClaudePrice(),
		exchangeRate: 1,
	})
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrPricingUnavailable))
}
