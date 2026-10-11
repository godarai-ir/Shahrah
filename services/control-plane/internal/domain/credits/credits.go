// Package credits converts provider USD cost into GodarAI credits.
//
// A credit is a stable consumption unit: credits = 500 × official provider USD
// cost. Amounts are stored as int64 millicredits and every request total is
// rounded up once to 0.01 credit. All arithmetic is decimal; float input is only
// accepted through the explicit fallback path for legacy breakdowns.
package credits

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/shopspring/decimal"
)

// Millicredits is the storage unit for credit amounts (1 credit = 1,000 mc).
type Millicredits int64

// Source identifies which cost input produced a credit charge.
type Source string

const (
	// MillicreditsPerCredit is the number of millicredits in one credit.
	MillicreditsPerCredit = 1000
	// RoundingStep is the charge granularity: 0.01 credit.
	RoundingStep Millicredits = 10

	// SourceComponents means the charge was computed from decimal cost components.
	SourceComponents Source = "components"
	// SourceFloatFallback means the breakdown had no components and the float
	// USD total was used. It is acceptable for shadow mode only.
	SourceFloatFallback Source = "float_fallback"
	// SourceNoCost means the request carried no billable cost (for example a cache hit).
	SourceNoCost Source = "no_cost"
)

// creditsPerUSD is the fixed conversion factor. It is internal and must not be
// exposed through public APIs.
const creditsPerUSD = 500

var (
	// ErrInvalidCost reports a cost value that is not a non-negative decimal.
	ErrInvalidCost = errors.New("credit cost input is invalid")

	millicreditsPerUSD = decimal.NewFromInt(creditsPerUSD * MillicreditsPerCredit)
	roundingStep       = decimal.NewFromInt(int64(RoundingStep))
	millicreditsScale  = decimal.NewFromInt(MillicreditsPerCredit)
)

// Charge is the credit result for one priced request or side call.
type Charge struct {
	Millicredits    Millicredits
	Exact           decimal.Decimal
	Source          Source
	RateCardVersion int
}

// Credits returns the charge formatted with two decimals.
func (c Charge) Credits() string {
	return Format(c.Millicredits)
}

// ExactFromUSD converts a USD amount to exact (unrounded) millicredits.
func ExactFromUSD(usd, multiplier decimal.Decimal) decimal.Decimal {
	return usd.Mul(millicreditsPerUSD).Mul(multiplier)
}

// RoundUp rounds exact millicredits up to the next 0.01 credit.
func RoundUp(exact decimal.Decimal) Millicredits {
	if !exact.IsPositive() {
		return 0
	}

	steps := exact.Div(roundingStep).Ceil()

	return Millicredits(steps.Mul(roundingStep).IntPart())
}

// FromBreakdown computes the credit charge for a priced usage breakdown.
func FromBreakdown(breakdown platform.UsageCostBreakdown, card RateCard) (Charge, error) {
	multiplier := card.Multiplier(breakdownModelKey(breakdown))

	if len(breakdown.Components) == 0 {
		if breakdown.TotalCostUSD == 0 {
			return Charge{Exact: decimal.Zero, Source: SourceNoCost, RateCardVersion: card.Version}, nil
		}

		if breakdown.TotalCostUSD < 0 {
			return Charge{}, fmt.Errorf("%w: negative total cost", ErrInvalidCost)
		}

		exact := ExactFromUSD(decimal.NewFromFloat(breakdown.TotalCostUSD), multiplier)

		return Charge{
			Millicredits:    RoundUp(exact),
			Exact:           exact,
			Source:          SourceFloatFallback,
			RateCardVersion: card.Version,
		}, nil
	}

	totalUSD, err := ComponentsUSD(breakdown.Components)
	if err != nil {
		return Charge{}, err
	}

	exact := ExactFromUSD(totalUSD, multiplier)

	return Charge{
		Millicredits:    RoundUp(exact),
		Exact:           exact,
		Source:          SourceComponents,
		RateCardVersion: card.Version,
	}, nil
}

// breakdownModelKey returns the multiplier key for a priced breakdown.
func breakdownModelKey(breakdown platform.UsageCostBreakdown) string {
	if breakdown.PricingSnapshot != nil && breakdown.PricingSnapshot.Provider != "" {
		return ModelKey(breakdown.PricingSnapshot.Provider, breakdown.Model)
	}

	return NormalizeModel(breakdown.Model)
}

// ComponentsUSD sums the decimal USD cost of every usage component.
func ComponentsUSD(components []platform.UsageCostComponent) (decimal.Decimal, error) {
	total := decimal.Zero

	for _, component := range components {
		raw := strings.TrimSpace(component.CostUSD)
		if raw == "" {
			return decimal.Zero, fmt.Errorf("%w: %s has no cost", ErrInvalidCost, component.Dimension)
		}

		cost, err := decimal.NewFromString(raw)
		if err != nil || cost.IsNegative() {
			return decimal.Zero, fmt.Errorf("%w: %s", ErrInvalidCost, component.Dimension)
		}

		total = total.Add(cost)
	}

	return total, nil
}

// FromAvoidedUSD converts an avoided (cache hit) USD cost to millicredits for
// analytics. It rounds up like a charge so savings match what would be billed.
func FromAvoidedUSD(usd float64, model string, card RateCard) Millicredits {
	if usd <= 0 {
		return 0
	}

	return RoundUp(ExactFromUSD(decimal.NewFromFloat(usd), card.Multiplier(model)))
}

// Format renders millicredits as a credit amount with exactly two decimals.
func Format(amount Millicredits) string {
	return decimal.NewFromInt(int64(amount)).Div(millicreditsScale).StringFixed(2)
}

// FormatExact renders exact millicredits as credits with up to the given
// number of decimals, trimming trailing zeros below two decimals.
func FormatExact(exact decimal.Decimal, places int32) string {
	value := exact.Div(millicreditsScale).Round(places)
	if value.Equal(value.Round(2)) {
		return value.StringFixed(2)
	}

	return value.String()
}
