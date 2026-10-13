package credits

import (
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/shopspring/decimal"
)

// BenchmarkFromBreakdown measures the per-request shadow credit cost for a
// realistic multi-component breakdown (Gate 1 budget: well under 1 ms).
func BenchmarkFromBreakdown(b *testing.B) {
	card := RateCard{Version: 2, Multipliers: map[string]decimal.Decimal{"xai/grok-4.7": decimal.RequireFromString("1.1")}}
	breakdown := platform.UsageCostBreakdown{
		Model:           "claude-sonnet-5",
		PricingSnapshot: &platform.UsagePricingSnapshot{Provider: "anthropic"},
		Components: []platform.UsageCostComponent{
			{Dimension: "input_tokens", CostUSD: "0.004"},
			{Dimension: "output_tokens", CostUSD: "0.005"},
			{Dimension: "cache_read_input_tokens", CostUSD: "0.0000412"},
			{Dimension: "cache_creation_5m_input_tokens", CostUSD: "0.00075"},
			{Dimension: "web_search_requests", CostUSD: "0.01"},
		},
	}

	b.ReportAllocs()

	for b.Loop() {
		if _, err := FromBreakdown(breakdown, card); err != nil {
			b.Fatal(err)
		}
	}
}
