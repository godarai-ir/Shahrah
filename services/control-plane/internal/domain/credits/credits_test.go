package credits

import (
	"errors"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/shopspring/decimal"
)

func perMillion(usd string) decimal.Decimal {
	return decimal.RequireFromString(usd).Div(decimal.NewFromInt(1_000_000))
}

// TestRatePer1KMatchesPRDRateCard pins the PRD sample rate card (credits per 1K tokens).
func TestRatePer1KMatchesPRDRateCard(t *testing.T) {
	t.Parallel()

	card := IdentityRateCard()
	tests := []struct {
		name      string
		perMillUS string
		want      string
	}{
		{name: "opus 5.5 input", perMillUS: "4", want: "2.00"},
		{name: "opus 5.5 output", perMillUS: "20", want: "10.00"},
		{name: "opus 5.5 cache read", perMillUS: "0.2", want: "0.10"},
		{name: "kimi k3 input", perMillUS: "3", want: "1.50"},
		{name: "kimi k3 output", perMillUS: "15", want: "7.50"},
		{name: "kimi k3 cache read", perMillUS: "0.3", want: "0.15"},
		{name: "sonnet 5 input", perMillUS: "2", want: "1.00"},
		{name: "sonnet 5 output", perMillUS: "10", want: "5.00"},
		{name: "grok 4.7 output", perMillUS: "6", want: "3.00"},
		{name: "grok 4.7 cache read", perMillUS: "0.5", want: "0.25"},
		{name: "light input", perMillUS: "0.25", want: "0.125"},
		{name: "light output", perMillUS: "2", want: "1.00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := FormatExact(card.RatePer1K("model", perMillion(tt.perMillUS)), 4)
			if got != tt.want {
				t.Fatalf("RatePer1K = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestRatePerUnitWebSearch(t *testing.T) {
	t.Parallel()

	card := IdentityRateCard()

	tests := map[string]string{"0.01": "5.00", "0.005": "2.50"}
	for usd, want := range tests {
		if got := FormatExact(card.RatePerUnit("m", decimal.RequireFromString(usd)), 2); got != want {
			t.Fatalf("RatePerUnit(%s) = %s, want %s", usd, got, want)
		}
	}
}

// TestTypicalMessageMatchesPRD pins the 2,000 input / 500 output examples.
func TestTypicalMessageMatchesPRD(t *testing.T) {
	t.Parallel()

	card := IdentityRateCard()
	tests := []struct {
		name          string
		input, output string
		want          string
	}{
		{name: "light", input: "0.25", output: "2", want: "0.75"},
		{name: "sonnet 5", input: "2", output: "10", want: "4.50"},
		{name: "opus 5.5", input: "4", output: "20", want: "9.00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := Format(card.TypicalMessage("m", perMillion(tt.input), perMillion(tt.output)))
			if got != tt.want {
				t.Fatalf("TypicalMessage = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestRoundUpOncePerRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		exact string
		want  Millicredits
	}{
		{exact: "0", want: 0},
		{exact: "-5", want: 0},
		{exact: "0.0001", want: 10},
		{exact: "2", want: 10},
		{exact: "10", want: 10},
		{exact: "10.0000001", want: 20},
		{exact: "4500", want: 4500},
	}

	for _, tt := range tests {
		if got := RoundUp(decimal.RequireFromString(tt.exact)); got != tt.want {
			t.Fatalf("RoundUp(%s) = %d, want %d", tt.exact, got, tt.want)
		}
	}
}

func TestFromBreakdownSumsComponentsBeforeRounding(t *testing.T) {
	t.Parallel()

	// Three components of 1 mc each must round once (to 10 mc), not three times (30 mc).
	breakdown := platform.UsageCostBreakdown{
		Model: "anthropic/claude-sonnet-5",
		Components: []platform.UsageCostComponent{
			{Dimension: "input_tokens", CostUSD: "0.000002"},
			{Dimension: "output_tokens", CostUSD: "0.000002"},
			{Dimension: "cache_read_input_tokens", CostUSD: "0.000002"},
		},
	}

	charge, err := FromBreakdown(breakdown, IdentityRateCard())
	if err != nil {
		t.Fatalf("FromBreakdown: %v", err)
	}

	if charge.Millicredits != 10 || charge.Source != SourceComponents {
		t.Fatalf("charge = %+v, want 10 mc from components", charge)
	}

	if charge.Credits() != "0.01" {
		t.Fatalf("Credits() = %s, want 0.01", charge.Credits())
	}
}

func TestFromBreakdownSonnetTypicalMessage(t *testing.T) {
	t.Parallel()

	breakdown := platform.UsageCostBreakdown{
		Model:        "anthropic/claude-sonnet-5",
		TotalCostUSD: 0.009,
		Components: []platform.UsageCostComponent{
			{Dimension: "input_tokens", CostUSD: "0.004"},
			{Dimension: "output_tokens", CostUSD: "0.005"},
		},
	}

	charge, err := FromBreakdown(breakdown, IdentityRateCard())
	if err != nil {
		t.Fatalf("FromBreakdown: %v", err)
	}

	if charge.Millicredits != 4500 || charge.Credits() != "4.50" {
		t.Fatalf("charge = %+v, want 4.50 credits", charge)
	}
}

func TestFromBreakdownAppliesModelMultiplier(t *testing.T) {
	t.Parallel()

	card := RateCard{
		Version: 3,
		Multipliers: map[string]decimal.Decimal{
			"xai/grok-4.7": decimal.RequireFromString("1.2"),
		},
	}
	breakdown := platform.UsageCostBreakdown{
		Model:      " XAI/Grok-4.7 ",
		Components: []platform.UsageCostComponent{{Dimension: "input_tokens", CostUSD: "0.01"}},
	}

	charge, err := FromBreakdown(breakdown, card)
	if err != nil {
		t.Fatalf("FromBreakdown: %v", err)
	}

	if charge.Millicredits != 6000 || charge.RateCardVersion != 3 {
		t.Fatalf("charge = %+v, want 6000 mc on rc-3", charge)
	}

	if card.Label() != "rc-3" {
		t.Fatalf("Label() = %s", card.Label())
	}
}

func TestFromBreakdownFloatFallback(t *testing.T) {
	t.Parallel()

	charge, err := FromBreakdown(platform.UsageCostBreakdown{TotalCostUSD: 0.018}, IdentityRateCard())
	if err != nil {
		t.Fatalf("FromBreakdown: %v", err)
	}

	if charge.Source != SourceFloatFallback || charge.Millicredits != 9000 {
		t.Fatalf("charge = %+v, want float fallback of 9000 mc", charge)
	}
}

func TestFromBreakdownRejectsInvalidComponents(t *testing.T) {
	t.Parallel()

	tests := map[string]platform.UsageCostComponent{
		"empty":    {Dimension: "input_tokens"},
		"garbage":  {Dimension: "input_tokens", CostUSD: "abc"},
		"negative": {Dimension: "input_tokens", CostUSD: "-0.1"},
	}

	for name, component := range tests {
		breakdown := platform.UsageCostBreakdown{Components: []platform.UsageCostComponent{component}}
		if _, err := FromBreakdown(breakdown, IdentityRateCard()); !errors.Is(err, ErrInvalidCost) {
			t.Fatalf("%s: err = %v, want ErrInvalidCost", name, err)
		}
	}

	if _, err := FromBreakdown(platform.UsageCostBreakdown{TotalCostUSD: -1}, IdentityRateCard()); !errors.Is(err, ErrInvalidCost) {
		t.Fatalf("negative float total: err = %v, want ErrInvalidCost", err)
	}
}

// TestChargeProperties checks that charges are monotonic, step-aligned and never below exact.
func TestChargeProperties(t *testing.T) {
	t.Parallel()

	unit := perMillion("3")
	previous := Millicredits(0)

	for tokens := int64(0); tokens <= 20_000; tokens += 37 {
		exact := ExactFromUSD(unit.Mul(decimal.NewFromInt(tokens)), decimal.NewFromInt(1))
		charged := RoundUp(exact)

		if charged%RoundingStep != 0 {
			t.Fatalf("tokens=%d charged=%d is not step aligned", tokens, charged)
		}

		if decimal.NewFromInt(int64(charged)).LessThan(exact) {
			t.Fatalf("tokens=%d charged=%d below exact %s", tokens, charged, exact)
		}

		if charged < previous {
			t.Fatalf("tokens=%d charged=%d not monotonic (previous %d)", tokens, charged, previous)
		}

		previous = charged
	}
}

func TestFromAvoidedUSD(t *testing.T) {
	t.Parallel()

	if got := FromAvoidedUSD(0.009, "m", IdentityRateCard()); got != 4500 {
		t.Fatalf("FromAvoidedUSD = %d, want 4500", got)
	}

	if got := FromAvoidedUSD(0, "m", IdentityRateCard()); got != 0 {
		t.Fatalf("FromAvoidedUSD(0) = %d, want 0", got)
	}
}

func TestFormat(t *testing.T) {
	t.Parallel()

	tests := map[Millicredits]string{0: "0.00", 10: "0.01", 4500: "4.50", 190_000_000: "190000.00"}
	for amount, want := range tests {
		if got := Format(amount); got != want {
			t.Fatalf("Format(%d) = %s, want %s", amount, got, want)
		}
	}
}
