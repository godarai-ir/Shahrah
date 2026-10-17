package credits

import (
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/shopspring/decimal"
)

// ForecastWindow is how far back the depletion forecast looks.
const ForecastWindow = 7 * 24 * time.Hour

// forecastConfidentSpan is the shortest usage history a forecast is sure of.
const forecastConfidentSpan = 48 * time.Hour

// Forecast estimates when availableMC runs out at the rate consumedMC was spent
// since firstUsage, which is the first consumption inside the forecast window.
// The rate spreads over the time since firstUsage, at least one hour, so a new
// tenant's first requests do not read as a full week of usage. expiresBy, when
// set, is when the last spendable credit expires; the credits run out by then
// even without spending.
func Forecast(availableMC, consumedMC int64, firstUsage, now time.Time, expiresBy *time.Time) platform.CreditForecast {
	forecast := usageForecast(availableMC, consumedMC, firstUsage, now)
	if forecast.DepletesAt != nil {
		forecast.LimitedBy = platform.CreditForecastLimitedByUsage
	}

	if expiresBy != nil && availableMC > 0 && (forecast.DepletesAt == nil || expiresBy.Before(*forecast.DepletesAt)) {
		expiry := *expiresBy
		forecast.DepletesAt = &expiry
		forecast.LimitedBy = platform.CreditForecastLimitedByExpiry
	}

	return forecast
}

func usageForecast(availableMC, consumedMC int64, firstUsage, now time.Time) platform.CreditForecast {
	forecast := platform.CreditForecast{
		Status:              platform.CreditForecastNoUsage,
		WindowDays:          int(ForecastWindow / (24 * time.Hour)),
		ConsumedCredits:     Format(Millicredits(max(consumedMC, 0))),
		DailyAverageCredits: Format(0),
	}

	if consumedMC <= 0 {
		return forecast
	}

	windowStart := now.Add(-ForecastWindow)
	if firstUsage.IsZero() || firstUsage.Before(windowStart) {
		firstUsage = windowStart
	}

	span := max(now.Sub(firstUsage), time.Hour)

	forecast.Status = platform.CreditForecastOK
	if span < forecastConfidentSpan {
		forecast.Status = platform.CreditForecastUncertain
	}

	perDay := decimal.NewFromInt(consumedMC).Mul(decimal.NewFromInt(int64(24 * time.Hour))).Div(decimal.NewFromInt(int64(span)))
	forecast.DailyAverageCredits = Format(Millicredits(perDay.Round(0).IntPart()))
	depletesAt := now

	if availableMC > 0 {
		left := decimal.NewFromInt(availableMC).Div(perDay).Mul(decimal.NewFromInt(int64(24 * time.Hour)))
		depletesAt = now.Add(time.Duration(left.IntPart()))
	}

	forecast.DepletesAt = &depletesAt

	return forecast
}
