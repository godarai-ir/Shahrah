package credits

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func rateSamples(now time.Time, days int, toman int64) []RateSample {
	samples := make([]RateSample, 0, days*2)
	for day := 1; day <= days; day++ {
		at := now.UTC().Truncate(24*time.Hour).AddDate(0, 0, -day)
		samples = append(samples,
			RateSample{At: at.Add(2 * time.Hour), USDToman: decimal.NewFromInt(toman - 1_000)},
			RateSample{At: at.Add(14 * time.Hour), USDToman: decimal.NewFromInt(toman + 1_000)},
		)
	}

	return samples
}

func TestSevenDayAverageNeedsEveryDay(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 10, 13, 9, 0, 0, 0, time.UTC)

	avg, ok := SevenDayAverage(rateSamples(now, 7, 300_000), now)
	require.True(t, ok)
	require.Equal(t, "300000", avg.String())

	samples := rateSamples(now, 7, 300_000)
	samples = append(samples, RateSample{At: now, USDToman: decimal.NewFromInt(900_000)})
	avg, ok = SevenDayAverage(samples, now)
	require.True(t, ok)
	require.Equal(t, "300000", avg.String(), "today's partial samples are excluded")

	_, ok = SevenDayAverage(rateSamples(now, 6, 300_000), now)
	require.False(t, ok, "a missing day blocks the average")

	_, ok = SevenDayAverage(nil, now)
	require.False(t, ok)
}

func TestProposePriceFollowsTheBand(t *testing.T) {
	t.Parallel()

	base := decimal.NewFromInt(LaunchBaseUSDToman)

	proposal, ok := ProposePrice(1_000, base, decimal.NewFromInt(300_000))
	require.True(t, ok)
	require.Equal(t, int64(1_100), proposal.PriceToman, "the PRD example: 300k gives 1,100 toman")
	require.Equal(t, "300000", proposal.BaseUSDToman.String())

	_, ok = ProposePrice(1_000, base, decimal.NewFromInt(297_000))
	require.False(t, ok, "+10% is still inside the band")

	_, ok = ProposePrice(1_000, base, decimal.NewFromInt(243_000))
	require.False(t, ok, "-10% is still inside the band")

	proposal, ok = ProposePrice(1_000, base, decimal.NewFromInt(240_000))
	require.True(t, ok)
	require.Equal(t, int64(900), proposal.PriceToman, "889 rounds to 900")

	_, ok = ProposePrice(1_000, base, decimal.Zero)
	require.False(t, ok)
}

func TestScalePriceIRR(t *testing.T) {
	t.Parallel()

	require.Equal(t, int64(55_000_000), ScalePriceIRR(50_000_000, 1_000, 1_100))
	require.Equal(t, int64(594_000_000), ScalePriceIRR(540_000_000, 1_000, 1_100))
	require.Equal(t, int64(7), ScalePriceIRR(7, 0, 1_100))
}
