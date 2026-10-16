package credits

import (
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// Prices in the PRD are in toman; the database stores IRR. Every identifier
// below names its unit, so a 10× slip is visible and caught by tests.
const (
	// ListPriceTomanPerCredit is the launch list price of one credit.
	ListPriceTomanPerCredit = 1000
	// IRRPerToman converts toman to the rial amounts stored in the database.
	IRRPerToman = 10
	// YearlyDiscountPercent is the annual-plan discount (the PRD maximum).
	YearlyDiscountPercent = 10

	// TrialCredits, TrialDays, and TrialMaxUsers define the credit trial.
	TrialCredits  = 500
	TrialDays     = 14
	TrialMaxUsers = 3
)

// Entitlement keys of the credit plans.
const (
	FeatureMonthlyGrant    = "credits.monthly_grant"
	FeatureBonusPercent    = "credits.bonus_percent"
	FeatureOverageAllowed  = "credits.overage_allowed"
	FeatureRolloverAllowed = "credits.rollover_allowed"
	FeatureBYOKFeePercent  = "credits.byok_fee_percent"
	FeatureTrialGrant      = "credits.trial_grant"
	FeatureTrialDays       = "credits.trial_days"
	FeatureUsersMax        = "users.max"
)

// PlanOffer is one self-serve plan of the credit pricing PRD.
type PlanOffer struct {
	Slug     string
	MaxUsers int
	// BaseCredits is what the monthly price buys at the list price; the plan
	// grants MonthlyCredits, and the difference is the bonus.
	BaseCredits    int64
	MonthlyCredits int64
}

// PlanOffers are the self-serve plans. Enterprise is priced per contract
// through tenant entitlement overrides.
var PlanOffers = []PlanOffer{
	{Slug: "developer", MaxUsers: 3, BaseCredits: 5_000, MonthlyCredits: 5_000},
	{Slug: "team", MaxUsers: 10, BaseCredits: 15_000, MonthlyCredits: 16_500},
	{Slug: "growth", MaxUsers: 30, BaseCredits: 45_000, MonthlyCredits: 54_000},
	{Slug: "business", MaxUsers: 100, BaseCredits: 150_000, MonthlyCredits: 190_000},
}

// MonthlyPriceToman is the monthly plan price before VAT.
func (p PlanOffer) MonthlyPriceToman() int64 {
	return p.BaseCredits * ListPriceTomanPerCredit
}

// MonthlyPriceIRR is the monthly plan price stored in plan_prices.amount_minor.
func (p PlanOffer) MonthlyPriceIRR() int64 {
	return p.MonthlyPriceToman() * IRRPerToman
}

// YearlyPriceIRR is twelve months at the yearly discount.
func (p PlanOffer) YearlyPriceIRR() int64 {
	return p.MonthlyPriceIRR() * 12 * (100 - YearlyDiscountPercent) / 100
}

// BonusPercent is the extra credits over the list price, rounded to 2 places.
func (p PlanOffer) BonusPercent() decimal.Decimal {
	if p.BaseCredits <= 0 {
		return decimal.Zero
	}

	extra := decimal.NewFromInt(p.MonthlyCredits - p.BaseCredits)

	return extra.Mul(decimal.NewFromInt(100)).Div(decimal.NewFromInt(p.BaseCredits)).Round(2)
}

// PlanSlice is the monthly delivery window of plan credits.
type PlanSlice struct {
	Start time.Time
	End   time.Time
}

// CurrentPlanSlice returns the monthly slice of a subscription period that
// contains now. Slices start on the period start and repeat monthly, so annual
// plans deliver credits each month; the last slice ends with the period.
// Months are counted in UTC so every worker derives the same slice, and the
// same idempotency key, whatever its local time zone.
func CurrentPlanSlice(periodStart, periodEnd, now time.Time) (PlanSlice, bool) {
	periodStart, periodEnd = periodStart.UTC(), periodEnd.UTC()
	if !periodEnd.After(periodStart) || now.Before(periodStart) || !now.Before(periodEnd) {
		return PlanSlice{}, false
	}

	start := periodStart

	for months := 1; ; months++ {
		next := periodStart.AddDate(0, months, 0)
		if !next.Before(periodEnd) {
			return PlanSlice{Start: start, End: periodEnd}, true
		}

		if next.After(now) {
			return PlanSlice{Start: start, End: next}, true
		}

		start = next
	}
}

// TrialExpiry is when trial credits expire: the trial's end, capped at `days`
// after it started.
func TrialExpiry(trialStart time.Time, trialEnd *time.Time, days int) time.Time {
	if days <= 0 {
		days = TrialDays
	}

	limit := trialStart.AddDate(0, 0, days)
	if trialEnd != nil && trialEnd.Before(limit) {
		return *trialEnd
	}

	return limit
}

// UserLimit resolves the active-user cap from the users.max entitlement JSON.
// A missing, null, or negative value means unlimited. A trialing credit tenant
// is further capped at trialMaxUsers.
func UserLimit(raw string, trialingCreditTenant bool, trialMaxUsers int) (int, bool) {
	limit, limited := parseUserLimit(raw)
	if !trialingCreditTenant || trialMaxUsers <= 0 {
		return limit, limited
	}

	if !limited || trialMaxUsers < limit {
		return trialMaxUsers, true
	}

	return limit, true
}

func parseUserLimit(raw string) (int, bool) {
	value, err := decimal.NewFromString(strings.TrimSpace(raw))
	if err != nil || value.IsNegative() {
		return 0, false
	}

	return int(value.IntPart()), true
}

// UserCapAllows reports whether a tenant using `used` seats may take `adding`
// more under `limit`. Tenants already over the limit are grandfathered: they
// keep their users but cannot add more.
func UserCapAllows(used, adding, limit int) bool {
	return used+adding <= limit
}
