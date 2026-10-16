package credits

import (
	"time"

	"github.com/shopspring/decimal"
)

// ListPriceIRRPerCredit is the launch list price of one credit in IRR. The
// credit price policy (Phase 2d) replaces it with the active approved revision.
const ListPriceIRRPerCredit = ListPriceTomanPerCredit * IRRPerToman

// PackOffer is one add-on pack of the credit pricing PRD.
type PackOffer struct {
	Slug            string
	Credits         int64
	DiscountPercent int64
}

// PackOffers are the launch add-on packs. Their PRD prices are 1,000,000,
// 4,500,000, and 16,000,000 toman at the launch list price.
var PackOffers = []PackOffer{
	{Slug: "credits-1000", Credits: 1_000, DiscountPercent: 0},
	{Slug: "credits-5000", Credits: 5_000, DiscountPercent: 10},
	{Slug: "credits-20000", Credits: 20_000, DiscountPercent: 20},
}

// PackQuote is the price of a pack at one list price and VAT rate.
type PackQuote struct {
	ListPriceIRRPerCredit int64
	PriceIRR              int64
	TaxPercent            decimal.Decimal
	TaxIRR                int64
	PayableIRR            int64
}

// QuotePack prices a pack: credits at the list price less the pack discount,
// plus VAT rounded to the nearest rial.
func QuotePack(credits int64, discountPercent decimal.Decimal, listPriceIRRPerCredit int64, taxPercent decimal.Decimal) PackQuote {
	hundred := decimal.NewFromInt(100)
	gross := decimal.NewFromInt(credits).Mul(decimal.NewFromInt(listPriceIRRPerCredit))
	price := gross.Mul(hundred.Sub(discountPercent)).Div(hundred).Round(0)

	if taxPercent.IsNegative() {
		taxPercent = decimal.Zero
	}

	tax := price.Mul(taxPercent).Div(hundred).Round(0)

	return PackQuote{
		ListPriceIRRPerCredit: listPriceIRRPerCredit,
		PriceIRR:              price.IntPart(),
		TaxPercent:            taxPercent,
		TaxIRR:                tax.IntPart(),
		PayableIRR:            price.Add(tax).IntPart(),
	}
}

// PricePerCreditIRR is what one credit of a paid lot cost, before VAT.
func PricePerCreditIRR(priceIRR, credits int64) decimal.Decimal {
	if credits <= 0 {
		return decimal.Zero
	}

	return decimal.NewFromInt(priceIRR).Div(decimal.NewFromInt(credits)).Round(4)
}

// PackExpiry is when credits bought at `purchasedAt` expire.
func PackExpiry(purchasedAt time.Time, validityMonths int) time.Time {
	if validityMonths <= 0 {
		validityMonths = 12
	}

	return purchasedAt.AddDate(0, validityMonths, 0)
}

// OverageChargeIRR is the IRR charge for `debtMC` millicredits of overage at
// the list price, rounded up to the next rial.
func OverageChargeIRR(debtMC, listPriceIRRPerCredit int64) int64 {
	if debtMC <= 0 {
		return 0
	}

	exact := decimal.NewFromInt(debtMC).Mul(decimal.NewFromInt(listPriceIRRPerCredit)).Div(decimal.NewFromInt(MillicreditsPerCredit))

	return exact.Ceil().IntPart()
}

// OverageKey is the idempotency key of the overage settlement before a plan slice.
func OverageKey(subscriptionID string, sliceStart time.Time) string {
	return "overage:" + subscriptionID + ":" + sliceStart.UTC().Format(time.RFC3339)
}
