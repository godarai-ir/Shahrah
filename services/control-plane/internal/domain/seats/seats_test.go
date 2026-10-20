package seats

import (
	"testing"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/stretchr/testify/require"
)

func TestSeatPricesAtLaunchListPrice(t *testing.T) {
	list := int64(credits.ListPriceIRRPerCredit)

	tests := []struct {
		seat     Type
		interval string
		wantIRR  int64
	}{
		// 2,000 credits × 1,000 toman × 90% = 1,800,000 toman.
		{Pro, "monthly", 18_000_000},
		{Pro, "yearly", 18_000_000 * 12 * 90 / 100},
		// 8,000 credits × 1,000 toman × 80% = 6,400,000 toman.
		{Max, "monthly", 64_000_000},
		{Max, "annual", 64_000_000 * 12 * 90 / 100},
		{Free, "monthly", 0},
	}

	for _, tt := range tests {
		got, err := PeriodPriceIRR(Offers[tt.seat], list, tt.interval)
		require.NoError(t, err)
		require.Equal(t, tt.wantIRR, got, "%s %s", tt.seat, tt.interval)
	}

	_, err := PeriodPriceIRR(Offers[Pro], list, "quarterly")
	require.ErrorIs(t, err, ErrIntervalNotSupported)
}

func TestSeatPricesFollowListPrice(t *testing.T) {
	got := MonthlyPriceIRR(Offers[Pro], 12_000)
	require.Equal(t, int64(2_000*12_000*90/100), got)
}

func TestProrate(t *testing.T) {
	start := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 30)

	tests := []struct {
		name          string
		now           time.Time
		wantRemaining int
		wantAmount    int64
	}{
		{"period start pays in full", start, 30, 3_000},
		{"before the period pays in full", start.Add(-time.Hour), 30, 3_000},
		{"half way", start.AddDate(0, 0, 15), 15, 1_500},
		{"partial day counts as a day", start.AddDate(0, 0, 29).Add(time.Hour), 1, 100},
		{"rounds up to the rial", start.AddDate(0, 0, 20), 10, 1_000},
		{"after the period pays nothing", end.Add(time.Minute), 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Prorate(3_000, start, end, tt.now)
			require.Equal(t, tt.wantRemaining, got.RemainingDays)
			require.Equal(t, tt.wantAmount, got.AmountIRR)
		})
	}

	got := Prorate(1_000, start, start.AddDate(0, 0, 3), start.AddDate(0, 0, 1))
	require.Equal(t, int64(667), got.AmountIRR, "2/3 of 1000 rounds up")
}

func TestDecideChange(t *testing.T) {
	tests := []struct {
		from, to Type
		freeUsed bool
		want     Change
		wantErr  error
	}{
		{from: Unassigned, to: Free, want: Change{Unassigned, Free, Immediate, false}},
		{from: Unassigned, to: Free, freeUsed: true, wantErr: ErrFreeSeatUsed},
		{from: Unassigned, to: Pro, want: Change{Unassigned, Pro, Immediate, true}},
		{from: Free, to: Pro, freeUsed: true, want: Change{Free, Pro, Immediate, true}},
		{from: Free, to: Max, freeUsed: true, want: Change{Free, Max, Immediate, true}},
		{from: Pro, to: Max, want: Change{Pro, Max, Immediate, true}},
		{from: Max, to: Pro, want: Change{Max, Pro, AtPeriodEnd, false}},
		{from: Pro, to: Unassigned, want: Change{Pro, Unassigned, AtPeriodEnd, false}},
		{from: Max, to: Unassigned, want: Change{Max, Unassigned, AtPeriodEnd, false}},
		{from: Free, to: Unassigned, freeUsed: true, want: Change{Free, Unassigned, Immediate, false}},
		{from: Pro, to: Free, wantErr: ErrBackToFree},
		{from: Max, to: Free, wantErr: ErrBackToFree},
		{from: Pro, to: Pro, wantErr: ErrSameSeat},
		{from: "gold", to: Pro, wantErr: ErrInvalidSeatType},
	}

	for _, tt := range tests {
		got, err := DecideChange(tt.from, tt.to, tt.freeUsed)
		if tt.wantErr != nil {
			require.ErrorIs(t, err, tt.wantErr, "%s -> %s", tt.from, tt.to)

			continue
		}

		require.NoError(t, err, "%s -> %s", tt.from, tt.to)
		require.Equal(t, tt.want, got, "%s -> %s", tt.from, tt.to)
	}
}

func TestEnforcedAndPoolAccess(t *testing.T) {
	now := time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)
	later := now.Add(time.Hour)
	earlier := now.Add(-time.Hour)

	require.False(t, Enforced(false, "active", nil, now), "flag off")
	require.False(t, Enforced(true, "trialing", nil, now), "trial keeps the pool")
	require.False(t, Enforced(true, "active", &later, now), "grace period")
	require.True(t, Enforced(true, "active", &earlier, now))
	require.True(t, Enforced(true, "active", nil, now))

	require.True(t, PoolAccess(Free, false))
	require.True(t, PoolAccess(Unassigned, false))
	require.False(t, PoolAccess(Free, true))
	require.False(t, PoolAccess(Unassigned, true))
	require.True(t, PoolAccess(Pro, true))
	require.True(t, PoolAccess(Max, true))
}

func TestParseType(t *testing.T) {
	for _, tt := range []struct {
		raw  string
		want Type
	}{{"", Unassigned}, {" PRO ", Pro}, {"max", Max}, {"free", Free}, {"none", Unassigned}} {
		got, err := ParseType(tt.raw)
		require.NoError(t, err)
		require.Equal(t, tt.want, got)
	}

	_, err := ParseType("gold")
	require.ErrorIs(t, err, ErrInvalidSeatType)
}

func TestGrantAmounts(t *testing.T) {
	require.Equal(t, int64(500_000), Free.MonthlyGrantMC())
	require.Equal(t, int64(2_000_000), Pro.MonthlyGrantMC())
	require.Equal(t, int64(8_000_000), Max.MonthlyGrantMC())
	require.Zero(t, Unassigned.MonthlyGrantMC())
}

func TestOverflow(t *testing.T) {
	require.Equal(t, int64(0), Overflow(100, 80, 30), "fully inside")
	require.Equal(t, int64(10), Overflow(100, 110, 30), "crosses the allowance")
	require.Equal(t, int64(30), Overflow(100, 150, 30), "already past")
	require.Equal(t, int64(30), Overflow(0, 30, 30), "no allowance")
}

func TestKeys(t *testing.T) {
	slice := time.Date(2026, 10, 1, 0, 0, 0, 0, time.FixedZone("IRST", 12600))
	require.Equal(t, "seat_grant:t:u:pro:2026-09-30T20:30:00Z", GrantKey("t", "u", Pro, slice))
	require.Equal(t, "seat_free:t:u", FreeGrantKey("t", "u"))
	require.Equal(t, "credit:seat:t:u:window:202610", WindowKey(" t", "u ", "202610"))
}
