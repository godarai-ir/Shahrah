package storage

import (
	"testing"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/domain/seats"
	"github.com/stretchr/testify/require"
)

func TestSeatPurchasePeriod(t *testing.T) {
	now := time.Date(2026, time.October, 22, 18, 0, 0, 0, time.UTC)

	t.Run("free plan starts a new period instead of following its open end", func(t *testing.T) {
		// The free plan's period ends in 9999; prorating against it priced a
		// seat for ~292 years of days (the time.Duration limit).
		free := seatSubscription{
			Found: true, PlanSlug: seats.FreePlanSlug, Status: "active", Interval: "monthly",
			PeriodStart: now.AddDate(0, 0, -3), PeriodEnd: seats.FreePeriodEnd,
		}

		period, err := seatPurchasePeriod(free, "monthly", now)
		require.NoError(t, err)
		require.True(t, period.starts)
		require.Equal(t, now, period.start)
		// 30 Mehr to 30 Aban 1405: one Jalali month of 30 days.
		require.Equal(t, time.Date(2026, time.November, 21, 18, 0, 0, 0, time.UTC), period.end)

		proration := seats.Prorate(1_000_000, period.start, period.end, now)
		require.Equal(t, 30, proration.PeriodDays)
		require.Equal(t, 30, proration.RemainingDays)
	})

	t.Run("running seat subscription keeps its period", func(t *testing.T) {
		running := seatSubscription{
			Found: true, PlanSlug: seats.SeatsPlanSlug, Status: "active", Interval: "monthly",
			PeriodStart: now.AddDate(0, 0, -10), PeriodEnd: now.AddDate(0, 0, 20),
		}

		period, err := seatPurchasePeriod(running, "", now)
		require.NoError(t, err)
		require.False(t, period.starts)
		require.Equal(t, running.PeriodEnd, period.end)

		_, err = seatPurchasePeriod(running, "yearly", now)
		require.ErrorIs(t, err, seats.ErrIntervalLocked)
	})
}
