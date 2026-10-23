package credits

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestPackOffersMatchThePRDInTomanAndIRR(t *testing.T) {
	t.Parallel()

	wantToman := map[string]int64{"credits-1000": 1_000_000, "credits-5000": 4_500_000, "credits-20000": 16_000_000}

	require.Equal(t, int64(10_000), int64(ListPriceIRRPerCredit), "1 credit = 1,000 toman = 10,000 IRR")

	for _, offer := range PackOffers {
		quote := QuotePack(offer.Credits, decimal.NewFromInt(offer.DiscountPercent), ListPriceIRRPerCredit, decimal.Zero)
		require.Equal(t, wantToman[offer.Slug]*IRRPerToman, quote.PriceIRR, "%s: IRR must be exactly 10x toman", offer.Slug)
		require.Equal(t, quote.PriceIRR, quote.PayableIRR)
	}
}

func TestQuotePackAddsVATRoundedToTheRial(t *testing.T) {
	t.Parallel()

	quote := QuotePack(5_000, decimal.NewFromInt(10), ListPriceIRRPerCredit, decimal.NewFromInt(10))
	require.Equal(t, int64(45_000_000), quote.PriceIRR)
	require.Equal(t, int64(4_500_000), quote.TaxIRR)
	require.Equal(t, int64(49_500_000), quote.PayableIRR)

	odd := QuotePack(3, decimal.RequireFromString("33.33"), 1_001, decimal.RequireFromString("9.5"))
	require.Equal(t, int64(2_002), odd.PriceIRR, "3003 x 66.67% rounds to the rial")
	require.Equal(t, int64(190), odd.TaxIRR)
	require.Equal(t, odd.PriceIRR+odd.TaxIRR, odd.PayableIRR)

	require.Zero(t, QuotePack(1, decimal.Zero, 10, decimal.NewFromInt(-5)).TaxIRR, "negative VAT is ignored")
}

func TestPricePerCreditAndExpiry(t *testing.T) {
	t.Parallel()

	require.Equal(t, "9000", PricePerCreditIRR(45_000_000, 5_000).String())
	require.Equal(t, "3333.3333", PricePerCreditIRR(10_000, 3).String())
	require.True(t, PricePerCreditIRR(1, 0).IsZero())

	bought := time.Date(2026, 10, 13, 9, 0, 0, 0, time.UTC)
	require.Equal(t, bought.AddDate(1, 0, 0), PackExpiry(bought, 12))
	require.Equal(t, bought.AddDate(1, 0, 0), PackExpiry(bought, 0))
}

func TestOverageChargeRoundsUpToTheRial(t *testing.T) {
	t.Parallel()

	require.Equal(t, int64(2_500_000), OverageChargeIRR(250_000, ListPriceIRRPerCredit))
	require.Equal(t, int64(1), OverageChargeIRR(1, 999))
	require.Zero(t, OverageChargeIRR(0, ListPriceIRRPerCredit))
	require.Equal(t, "overage:sub-1:2026-10-01T00:00:00Z", OverageKey("sub-1", time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)))
}
