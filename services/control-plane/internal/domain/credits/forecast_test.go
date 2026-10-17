package credits

import (
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/stretchr/testify/require"
)

func TestForecastWithoutUsageHasNoEstimate(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)
	forecast := Forecast(500_000, 0, time.Time{}, now, nil)

	require.Equal(t, platform.CreditForecastNoUsage, forecast.Status)
	require.Equal(t, 7, forecast.WindowDays)
	require.Equal(t, "0.00", forecast.ConsumedCredits)
	require.Nil(t, forecast.DepletesAt)
}

func TestForecastSpreadsAWeekOfUsage(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)
	// 70 credits over a full week is 10 a day; 30 left lasts 3 days.
	forecast := Forecast(30_000, 70_000, now.Add(-30*24*time.Hour), now, nil)

	require.Equal(t, platform.CreditForecastOK, forecast.Status)
	require.Equal(t, "70.00", forecast.ConsumedCredits)
	require.Equal(t, "10.00", forecast.DailyAverageCredits)
	require.NotNil(t, forecast.DepletesAt)
	require.Equal(t, now.Add(72*time.Hour), *forecast.DepletesAt)
}

func TestForecastIsUncertainOnShortHistory(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)
	// 5 credits in the last 12 hours is 10 a day.
	forecast := Forecast(100_000, 5_000, now.Add(-12*time.Hour), now, nil)

	require.Equal(t, platform.CreditForecastUncertain, forecast.Status)
	require.Equal(t, "10.00", forecast.DailyAverageCredits)
	require.Equal(t, now.Add(10*24*time.Hour), *forecast.DepletesAt)
}

func TestForecastWithNothingLeftDepletesNow(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)
	forecast := Forecast(0, 7_000, now.Add(-7*24*time.Hour), now, nil)

	require.Equal(t, now, *forecast.DepletesAt)
}

func TestForecastStopsAtExpiry(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)
	expiry := now.Add(14 * 24 * time.Hour)

	// 10 a day lasts 47 days, but the credits expire in 14.
	forecast := Forecast(470_000, 70_000, now.Add(-7*24*time.Hour), now, &expiry)
	require.Equal(t, expiry, *forecast.DepletesAt)
	require.Equal(t, platform.CreditForecastLimitedByExpiry, forecast.LimitedBy)

	// Spending runs out first.
	forecast = Forecast(30_000, 70_000, now.Add(-7*24*time.Hour), now, &expiry)
	require.Equal(t, now.Add(72*time.Hour), *forecast.DepletesAt)
	require.Equal(t, platform.CreditForecastLimitedByUsage, forecast.LimitedBy)

	// Without usage the expiry is still the end.
	forecast = Forecast(500_000, 0, time.Time{}, now, &expiry)
	require.Equal(t, platform.CreditForecastNoUsage, forecast.Status)
	require.Equal(t, expiry, *forecast.DepletesAt)
}
