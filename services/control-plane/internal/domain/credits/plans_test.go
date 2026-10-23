package credits

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPlanOffersMatchThePRDInTomanAndIRR(t *testing.T) {
	t.Parallel()

	// Monthly prices from the PRD, in toman, before VAT.
	wantToman := map[string]int64{
		"developer": 5_000_000,
		"team":      15_000_000,
		"growth":    45_000_000,
		"business":  150_000_000,
	}
	wantBonus := map[string]string{"developer": "0", "team": "10", "growth": "20", "business": "26.67"}

	require.Len(t, PlanOffers, len(wantToman))

	for _, offer := range PlanOffers {
		require.Equal(t, wantToman[offer.Slug], offer.MonthlyPriceToman(), offer.Slug)
		require.Equal(t, wantToman[offer.Slug]*10, offer.MonthlyPriceIRR(), "%s: IRR must be exactly 10× toman", offer.Slug)
		require.Equal(t, offer.MonthlyPriceIRR()*12*9/10, offer.YearlyPriceIRR(), offer.Slug)
		require.Equal(t, wantBonus[offer.Slug], offer.BonusPercent().String(), offer.Slug)
		require.GreaterOrEqual(t, offer.MonthlyCredits, offer.BaseCredits, "discounts are bonus credits, never a lower list price")
	}
}

func TestCurrentPlanSliceDeliversMonthly(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 1, 31, 10, 0, 0, 0, time.UTC)
	yearEnd := start.AddDate(1, 0, 0)

	slice, ok := CurrentPlanSlice(start, yearEnd, start)
	require.True(t, ok)
	require.Equal(t, PlanSlice{Start: start, End: start.AddDate(0, 1, 0)}, slice)

	mid := start.AddDate(0, 5, 3)
	slice, ok = CurrentPlanSlice(start, yearEnd, mid)
	require.True(t, ok)
	require.Equal(t, start.AddDate(0, 5, 0), slice.Start)
	require.Equal(t, start.AddDate(0, 6, 0), slice.End)

	boundary := start.AddDate(0, 6, 0)
	slice, ok = CurrentPlanSlice(start, yearEnd, boundary)
	require.True(t, ok)
	require.Equal(t, boundary, slice.Start, "a slice boundary belongs to the next slice")

	last, ok := CurrentPlanSlice(start, yearEnd, yearEnd.Add(-time.Second))
	require.True(t, ok)
	require.Equal(t, yearEnd, last.End)

	_, ok = CurrentPlanSlice(start, yearEnd, yearEnd)
	require.False(t, ok, "the period is over")

	_, ok = CurrentPlanSlice(start, yearEnd, start.Add(-time.Second))
	require.False(t, ok)
}

func TestCurrentPlanSliceClipsAShortPeriod(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 20)

	slice, ok := CurrentPlanSlice(start, end, start.AddDate(0, 0, 3))
	require.True(t, ok)
	require.Equal(t, PlanSlice{Start: start, End: end}, slice)
}

func TestCurrentPlanSliceIgnoresTheLocalTimeZone(t *testing.T) {
	t.Parallel()

	// 31 January 22:00 UTC is already 1 February in Tehran, so stepping months
	// in the local zone would move every slice boundary and its grant key.
	tehran := time.FixedZone("IRST", 3*60*60+30*60)
	start := time.Date(2026, 1, 31, 22, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)
	now := start.AddDate(0, 1, 5)

	utcSlice, ok := CurrentPlanSlice(start, end, now)
	require.True(t, ok)

	localSlice, ok := CurrentPlanSlice(start.In(tehran), end.In(tehran), now)
	require.True(t, ok)
	require.Equal(t, utcSlice, localSlice)
	require.Equal(t, PlanGrantKey("sub", utcSlice.Start), PlanGrantKey("sub", localSlice.Start))
}

func TestUserCapAllows(t *testing.T) {
	t.Parallel()

	require.True(t, UserCapAllows(2, 1, 3))
	require.False(t, UserCapAllows(3, 1, 3))
	require.True(t, UserCapAllows(3, 0, 3), "accepting a counted invitation takes no new seat")
	require.False(t, UserCapAllows(5, 0, 3), "grandfathered tenants cannot add users")
}

func TestUserLimitParsesEntitlementsAndCapsCreditTrials(t *testing.T) {
	t.Parallel()

	cases := []struct {
		raw         string
		trialing    bool
		wantLimit   int
		wantLimited bool
	}{
		{raw: "10", wantLimit: 10, wantLimited: true},
		{raw: "0", wantLimit: 0, wantLimited: true},
		{raw: "null"},
		{raw: ""},
		{raw: "-1"},
		{raw: `"custom"`},
		{raw: "30", trialing: true, wantLimit: TrialMaxUsers, wantLimited: true},
		{raw: "2", trialing: true, wantLimit: 2, wantLimited: true},
		{raw: "null", trialing: true, wantLimit: TrialMaxUsers, wantLimited: true},
	}

	for _, tc := range cases {
		limit, limited := UserLimit(tc.raw, tc.trialing, TrialMaxUsers)
		require.Equal(t, tc.wantLimited, limited, tc.raw)
		require.Equal(t, tc.wantLimit, limit, tc.raw)
	}
}
