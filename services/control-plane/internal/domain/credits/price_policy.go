package credits

import (
	"time"

	"github.com/shopspring/decimal"
)

// Credit price policy (PRD): the toman price of a credit follows the dollar
// only when the 7-day average leaves a ±10% band around the base rate.
const (
	// LaunchBaseUSDToman is the dollar rate the launch list price was set at.
	LaunchBaseUSDToman = 270_000
	// PriceBandPercent is the half-width of the band around the base rate.
	PriceBandPercent = 10
	// PriceRoundingToman is the step new list prices are rounded to.
	PriceRoundingToman = 50
	// PriceAverageDays is how many daily samples the average needs.
	PriceAverageDays = 7
	// PlanPriceNoticeDays is how far ahead plan price changes are announced.
	PlanPriceNoticeDays = 30
	// MaxListPriceToman bounds a list price support sets by hand, well inside
	// the integer column and far above any plausible credit price.
	MaxListPriceToman = 1_000_000
)

// RateSample is one observation of the dollar rate in toman.
type RateSample struct {
	At       time.Time
	USDToman decimal.Decimal
}

// SevenDayAverage averages the samples of the PriceAverageDays UTC days before
// now (today excluded, since it is incomplete). Each day must have at least one
// sample, so a feed outage never produces a proposal from partial data.
func SevenDayAverage(samples []RateSample, now time.Time) (decimal.Decimal, bool) {
	today := now.UTC().Truncate(24 * time.Hour)
	from := today.AddDate(0, 0, -PriceAverageDays)
	days := map[time.Time]bool{}
	sum := decimal.Zero
	count := 0

	for _, sample := range samples {
		at := sample.At.UTC()
		if at.Before(from) || !at.Before(today) || !sample.USDToman.IsPositive() {
			continue
		}

		days[at.Truncate(24*time.Hour)] = true
		sum = sum.Add(sample.USDToman)
		count++
	}

	if len(days) < PriceAverageDays {
		return decimal.Zero, false
	}

	return sum.Div(decimal.NewFromInt(int64(count))).Round(2), true
}

// PriceProposal is a new list price triggered by the band.
type PriceProposal struct {
	PriceToman   int64
	BaseUSDToman decimal.Decimal
}

// ProposePrice returns a new list price when the average is outside the band:
// price × (average / base), rounded to the nearest PriceRoundingToman, with the
// average as the new base. Example: base 270,000, price 1,000, average 300,000
// gives 1,100.
func ProposePrice(priceToman int64, baseUSDToman, averageUSDToman decimal.Decimal) (PriceProposal, bool) {
	if priceToman <= 0 || !baseUSDToman.IsPositive() || !averageUSDToman.IsPositive() {
		return PriceProposal{}, false
	}

	band := decimal.NewFromInt(PriceBandPercent).Div(decimal.NewFromInt(100))
	low := baseUSDToman.Mul(decimal.NewFromInt(1).Sub(band))
	high := baseUSDToman.Mul(decimal.NewFromInt(1).Add(band))

	if averageUSDToman.GreaterThanOrEqual(low) && averageUSDToman.LessThanOrEqual(high) {
		return PriceProposal{}, false
	}

	exact := decimal.NewFromInt(priceToman).Mul(averageUSDToman).Div(baseUSDToman)
	step := decimal.NewFromInt(PriceRoundingToman)
	rounded := exact.Div(step).Round(0).Mul(step).IntPart()

	if rounded < PriceRoundingToman {
		rounded = PriceRoundingToman
	}

	if rounded == priceToman {
		return PriceProposal{}, false
	}

	return PriceProposal{PriceToman: rounded, BaseUSDToman: averageUSDToman}, true
}

// ScalePriceIRR moves an IRR price from one credit list price to another,
// rounded to the rial. Plan prices are a number of base credits at the list
// price, so this keeps any custom plan price proportional.
func ScalePriceIRR(amountIRR, fromPriceToman, toPriceToman int64) int64 {
	if fromPriceToman <= 0 {
		return amountIRR
	}

	return decimal.NewFromInt(amountIRR).Mul(decimal.NewFromInt(toPriceToman)).Div(decimal.NewFromInt(fromPriceToman)).Round(0).IntPart()
}
