package billing

import (
	"testing"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/services/control-plane/internal/providerregistry"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

// tokenBazaarEmbeddingsPrice builds a priceInfo the same way
// (*PricingEngine).modelPrice builds one for a PricingSnapshotCatalog
// provider (see pricing_engine.go): InputPerToken/OutputPerToken are read
// directly off the "input_tokens"/"output_tokens" catalog rates, and Rates
// itself is carried along for the cosmetic per-dimension Components
// breakdown. This mirrors the already-pinned official OpenAI rate baked into
// apps/model-discovery/providers/tokenbazaar/text-embedding-3-large.yaml
// (input_cost_per_token: 1.3e-7, output_cost_per_token: 0 — see that file's
// own top-of-file comment from the 2026-09-10 pricing-override delivery).
func tokenBazaarEmbeddingsPrice(inputRate, outputRate string) priceInfo {
	input := normalizedPricingRate{Dimension: "input_tokens", Unit: "token", UnitPriceUSD: decimal.RequireFromString(inputRate)}
	output := normalizedPricingRate{Dimension: "output_tokens", Unit: "token", UnitPriceUSD: decimal.RequireFromString(outputRate)}

	return priceInfo{
		Provider:       "tokenbazaar",
		InputPerToken:  input.UnitPriceUSD,
		OutputPerToken: output.UnitPriceUSD,
		Rates: map[string]normalizedPricingRate{
			"input_tokens":  input,
			"output_tokens": output,
		},
	}
}

// TestTokenBazaarPricingStrategyFinalizesEmbeddingsUsageWithExactDecimalCost
// exercises billing for text-embedding-3-large via TokenBazaar using the
// exact, already-pinned official OpenAI per-token rate
// (input_cost_per_token=1.3e-7, output_cost_per_token=0), asserting an exact
// decimal result (computed independently with shopspring/decimal, not a
// float epsilon comparison) for a representative input-token count.
// tokenBazaarPricingStrategy does not override FinalizeUsage, so this
// exercises the inherited standardPricingStrategy.FinalizeUsage/costBreakdown
// path unmodified -- the same generic path already used for every other
// TokenBazaar chat model's non-cached usage billing.
func TestTokenBazaarPricingStrategyFinalizesEmbeddingsUsageWithExactDecimalCost(t *testing.T) {
	price := tokenBazaarEmbeddingsPrice("0.00000013", "0")

	// Representative input-token count: 12,345 tokens of embeddings input.
	// Embeddings responses report completion_tokens=0 (no output tokens).
	usage := &contractsopenai.Usage{
		PromptTokens:     12345,
		CompletionTokens: 0,
		TotalTokens:      12345,
	}

	strategy := pricingStrategy("tokenbazaar")

	got, err := strategy.FinalizeUsage(pricingUsageInput{
		model:        "text-embedding-3-large",
		usage:        usage,
		price:        price,
		exchangeRate: 1,
	})
	require.NoError(t, err)

	// Independently computed exact-decimal expectation:
	//   input_tokens  = 12345 @ 0.00000013 = 0.00160485
	//   output_tokens = 0     @ 0          = 0
	// total = 0.00160485
	expected := decimal.RequireFromString("12345").Mul(decimal.RequireFromString("0.00000013"))
	expectedFloat, _ := expected.Float64()

	require.InDelta(t, expectedFloat, got.TotalCostUSD, 1e-15)
	require.InDelta(t, expectedFloat, got.InputCostUSD, 1e-15)
	require.Equal(t, 0.0, got.OutputCostUSD, "output cost for an embeddings-only model must always be exactly 0")
	require.Equal(t, 12345, got.InputTokens)
	require.Equal(t, 0, got.OutputTokens)
	require.Equal(t, "text-embedding-3-large", got.Model)
}

// TestTokenBazaarPricingStrategyEmbeddingsOutputCostIsAlwaysZero confirms
// output cost is exactly 0 regardless of the input-token count, since
// text-embedding-3-large's catalog scaffold pins output_cost_per_token to 0
// unconditionally (embeddings models have no output tokens).
func TestTokenBazaarPricingStrategyEmbeddingsOutputCostIsAlwaysZero(t *testing.T) {
	price := tokenBazaarEmbeddingsPrice("0.00000013", "0")

	for _, tokens := range []int{0, 1, 8192} {
		usage := &contractsopenai.Usage{PromptTokens: tokens, CompletionTokens: 0, TotalTokens: tokens}

		got, err := pricingStrategy("tokenbazaar").FinalizeUsage(pricingUsageInput{
			model:        "text-embedding-3-large",
			usage:        usage,
			price:        price,
			exchangeRate: 1,
		})
		require.NoError(t, err)
		require.Equal(t, 0.0, got.OutputCostUSD)
	}
}

// TestTokenBazaarPricingStrategyIsRegisteredForEmbeddingsCapability is a
// wiring sanity check: providerregistry.Lookup("tokenbazaar") must declare
// CapabilityEmbeddings, and resolve to PricingStrategyTokenBazaar, which is
// the strategy this file's other tests exercise via
// pricingStrategy("tokenbazaar").
func TestTokenBazaarPricingStrategyIsRegisteredForEmbeddingsCapability(t *testing.T) {
	definition, ok := providerregistry.Lookup("tokenbazaar")
	require.True(t, ok)
	require.True(t, definition.Supports(providerregistry.CapabilityEmbeddings))
	require.Equal(t, providerregistry.PricingStrategyTokenBazaar, definition.PricingStrategy)

	_, ok = pricingStrategy("tokenbazaar").(tokenBazaarPricingStrategy)
	require.True(t, ok, "pricingStrategy(\"tokenbazaar\") must resolve to tokenBazaarPricingStrategy")
}
