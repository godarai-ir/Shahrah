package billing

import (
	"testing"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

// gpt6SolRates mirrors the pricing rates imported from the gpt-6-sol catalog
// YAML, using the published https://developers.openai.com/api/docs/pricing
// values converted to USD per token.
func gpt6SolRates() []normalizedPricingRate {
	rate := func(dimension, price string, condition map[string]interface{}) normalizedPricingRate {
		return normalizedPricingRate{
			Dimension: dimension, Unit: "token",
			UnitPriceUSD: decimal.RequireFromString(price),
			Condition:    condition,
		}
	}
	long := map[string]interface{}{"min_prompt_tokens": 272001}
	tierCondition := func(tier string, longContext bool) map[string]interface{} {
		condition := map[string]interface{}{"service_tier": tier}
		if longContext {
			condition["min_prompt_tokens"] = 272001
		}

		return condition
	}

	return []normalizedPricingRate{
		rate("input_tokens", "0.000002", nil),
		rate("cache_read_tokens", "0.0000002", nil),
		rate("cache_write_tokens", "0.0000025", nil),
		rate("output_tokens", "0.00001", nil),
		rate("long_context_input_tokens", "0.000004", long),
		rate("long_context_cache_read_tokens", "0.0000004", long),
		rate("long_context_cache_write_tokens", "0.000005", long),
		rate("long_context_output_tokens", "0.000015", long),
		rate("batch_input_tokens", "0.000001", tierCondition("batch", false)),
		rate("batch_output_tokens", "0.000005", tierCondition("batch", false)),
		rate("batch_long_context_input_tokens", "0.000002", tierCondition("batch", true)),
		rate("batch_long_context_output_tokens", "0.0000075", tierCondition("batch", true)),
		rate("flex_input_tokens", "0.000001", tierCondition("flex", false)),
		rate("flex_output_tokens", "0.000005", tierCondition("flex", false)),
		rate("speed_input_tokens", "0.000004", tierCondition("priority", false)),
		rate("speed_cache_read_tokens", "0.0000004", tierCondition("priority", false)),
		rate("speed_cache_write_tokens", "0.000005", tierCondition("priority", false)),
		rate("speed_output_tokens", "0.00002", tierCondition("priority", false)),
		rate("speed_long_context_input_tokens", "0.000008", tierCondition("priority", true)),
		rate("speed_long_context_cache_read_tokens", "0.0000008", tierCondition("priority", true)),
		rate("speed_long_context_cache_write_tokens", "0.00001", tierCondition("priority", true)),
		rate("speed_long_context_output_tokens", "0.00003", tierCondition("priority", true)),
	}
}

func gpt6SolPrice() priceInfo {
	return priceInfo{
		Provider: "openai", Revision: "openai-gpt-6-sol",
		GeneratedAt: time.Now(), RateList: gpt6SolRates(),
	}
}

func breakdownDimensions(breakdown platform.UsageCostBreakdown) []string {
	dimensions := make([]string, 0, len(breakdown.Components))
	for _, component := range breakdown.Components {
		dimensions = append(dimensions, component.Dimension)
	}

	return dimensions
}

func TestOpenAIPricingBillsExactly272KPromptAtShortContextRates(t *testing.T) {
	breakdown, err := openAIUsageCostBreakdown(
		"openai/gpt-6-sol",
		&contractsopenai.Usage{PromptTokens: 272000, CompletionTokens: 1000, TotalTokens: 273000},
		gpt6SolPrice(),
		1,
		nil,
	)
	require.NoError(t, err)
	require.Equal(t, []string{"input_tokens", "output_tokens"}, breakdownDimensions(breakdown))
	require.InDelta(t, 272000*0.000002+1000*0.00001, breakdown.TotalCostUSD, 1e-12)
}

func TestOpenAIGPT6SolPricingAcrossServiceTiers(t *testing.T) {
	details := &contractsopenai.UsageTokenDetails{CachedTokens: 100000, CacheWriteTokens: 50000}

	tests := []struct {
		name         string
		serviceTier  string
		promptTokens int
		wantTier     string
		wantInput    string
		wantTotalUSD float64
	}{
		{
			name: "standard short context", promptTokens: 200000, wantTier: "standard",
			wantInput:    "input_tokens",
			wantTotalUSD: 50000*0.000002 + 100000*0.0000002 + 50000*0.0000025 + 1000*0.00001,
		},
		{
			name: "standard long context", promptTokens: 300000, wantTier: "standard",
			wantInput:    "long_context_input_tokens",
			wantTotalUSD: 150000*0.000004 + 100000*0.0000004 + 50000*0.000005 + 1000*0.000015,
		},
		{
			name: "fast short context", serviceTier: "fast", promptTokens: 200000, wantTier: "priority",
			wantInput:    "speed_input_tokens",
			wantTotalUSD: 50000*0.000004 + 100000*0.0000004 + 50000*0.000005 + 1000*0.00002,
		},
		{
			name: "fast long context", serviceTier: "fast", promptTokens: 300000, wantTier: "priority",
			wantInput:    "speed_long_context_input_tokens",
			wantTotalUSD: 150000*0.000008 + 100000*0.0000008 + 50000*0.00001 + 1000*0.00003,
		},
		{
			name: "priority long context", serviceTier: "priority", promptTokens: 300000, wantTier: "priority",
			wantInput:    "speed_long_context_input_tokens",
			wantTotalUSD: 150000*0.000008 + 100000*0.0000008 + 50000*0.00001 + 1000*0.00003,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			breakdown, err := openAIUsageCostBreakdown(
				"openai/gpt-6-sol",
				&contractsopenai.Usage{
					PromptTokens: test.promptTokens, CompletionTokens: 1000,
					TotalTokens: test.promptTokens + 1000, ServiceTier: test.serviceTier,
					PromptTokensDetails: details,
				},
				gpt6SolPrice(),
				1,
				nil,
			)
			require.NoError(t, err)
			require.Equal(t, test.wantTier, breakdown.ServiceTier)
			require.InDelta(t, test.wantTotalUSD, breakdown.TotalCostUSD, 1e-12)
			require.Contains(t, breakdownDimensions(breakdown), test.wantInput)
		})
	}
}

func TestOpenAIGPT6SolBatchLongContextUsesBatchRates(t *testing.T) {
	breakdown, err := openAIUsageCostBreakdown(
		"openai/gpt-6-sol",
		&contractsopenai.Usage{PromptTokens: 300000, CompletionTokens: 1000, TotalTokens: 301000, ServiceTier: "batch"},
		gpt6SolPrice(),
		1,
		nil,
	)
	require.NoError(t, err)
	require.Equal(t, "batch", breakdown.ServiceTier)
	require.InDelta(t, 300000*0.000002+1000*0.0000075, breakdown.TotalCostUSD, 1e-12)
}
