package credits

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestConvertIRRToCreditsRoundsDownAndKeepsTheResidue(t *testing.T) {
	t.Parallel()

	// 1 credit = 1,000 toman = 10,000 IRR.
	creditsMC, residue := ConvertIRRToCreditsMC(123_456, 1_000)
	require.Equal(t, int64(12_000), creditsMC)
	require.Equal(t, int64(3_456), residue)

	creditsMC, residue = ConvertIRRToCreditsMC(9_999, 1_000)
	require.Zero(t, creditsMC)
	require.Equal(t, int64(9_999), residue)

	creditsMC, residue = ConvertIRRToCreditsMC(-50, 1_000)
	require.Zero(t, creditsMC)
	require.Zero(t, residue)
}

func TestConvertIRRToCreditsIsNotOffByTen(t *testing.T) {
	t.Parallel()

	// 50,000,000 IRR is 5,000,000 toman: 5,000 credits at 1,000 toman, never 500 or 50,000.
	creditsMC, residue := ConvertIRRToCreditsMC(50_000_000, ListPriceTomanPerCredit)
	require.Equal(t, int64(5_000_000), creditsMC)
	require.Zero(t, residue)
}

func TestConvertIRRLimitRoundsUp(t *testing.T) {
	t.Parallel()

	require.Equal(t, int64(2_000), ConvertIRRLimitToCreditsMC(decimal.NewFromInt(10_001), 1_000))
	require.Equal(t, int64(1_000), ConvertIRRLimitToCreditsMC(decimal.NewFromInt(10_000), 1_000))
	require.Zero(t, ConvertIRRLimitToCreditsMC(decimal.Zero, 1_000))
}

func TestCreditsToIRRRoundsDown(t *testing.T) {
	t.Parallel()

	require.Equal(t, int64(12_340), CreditsToIRR(1_234, 1_000), "1.234 credits is 12,340 IRR")
	require.Equal(t, int64(10), CreditsToIRR(1, 1_000), "one millicredit is 10 IRR")
	require.Equal(t, int64(10_000), CreditsToIRR(1_000, 1_000))
	require.Zero(t, CreditsToIRR(-1, 1_000))
}

func TestConvertKeyIsPerSequenceAndPart(t *testing.T) {
	t.Parallel()

	require.Equal(t, "convert:t1:2:lot", ConvertKey(" t1 ", 2, "lot"))
	require.NotEqual(t, ConvertKey("t1", 1, "lot"), ConvertKey("t1", 2, "lot"))
}
