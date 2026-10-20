// Package seats holds the rules of member seats: the personal credit quota a
// tenant member gets for GodarChat, on top of the tenant's shared credit pool.
// See docs/PRD صندلی‌ها و اعتبار شخصی گدارAI.md.
package seats

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/shopspring/decimal"
)

// Type is a member's seat. A member without a seat is Unassigned.
type Type string

// Seat types. Free, Pro, and Max are stored in tenant_seat_assignments; a
// member with no assignment row is Unassigned.
const (
	Free       Type = "free"
	Pro        Type = "pro"
	Max        Type = "max"
	Unassigned Type = "unassigned"
)

// Offer is the quota and price rule of one seat type.
type Offer struct {
	Type Type
	// MonthlyCredits is granted every monthly slice, or once for a lifetime seat.
	MonthlyCredits int64
	// Lifetime seats are granted once per member and tenant and never renew.
	Lifetime bool
	// DiscountPercent is the discount on the credit list price.
	DiscountPercent int64
	// YearlyDiscountPercent is the extra discount of yearly billing. Max gets
	// less than Pro so its yearly price stays above the margin floor.
	YearlyDiscountPercent int64
	// PoolAccess lets the member spend from the tenant pool after the seat's credits.
	PoolAccess bool
}

// Offers are the launch seat types.
var Offers = map[Type]Offer{
	Free: {Type: Free, MonthlyCredits: 500, Lifetime: true},
	Pro:  {Type: Pro, MonthlyCredits: 2_000, DiscountPercent: 10, YearlyDiscountPercent: 10, PoolAccess: true},
	Max:  {Type: Max, MonthlyCredits: 8_000, DiscountPercent: 20, YearlyDiscountPercent: 5, PoolAccess: true},
}

// PaidTypes lists the seat types tenants buy, in rank order.
var PaidTypes = []Type{Pro, Max}

var (
	// ErrInvalidSeatType is returned for an unknown seat type.
	ErrInvalidSeatType = errors.New("seat type must be free, pro, max, or unassigned")
	// ErrSameSeat is returned when a member already has the requested seat.
	ErrSameSeat = errors.New("member already has this seat")
	// ErrBackToFree is returned when a paid seat would move to a free seat.
	ErrBackToFree = errors.New("a paid seat cannot move back to a free seat")
	// ErrFreeSeatUsed is returned when a member already had their free seat.
	ErrFreeSeatUsed = errors.New("member already used their free seat")
	// ErrInvalidQuantity is returned for a seat purchase of less than one seat.
	ErrInvalidQuantity = errors.New("seat quantity must be between 1 and 1000")
	// ErrIntervalNotSupported is returned for a billing interval seats cannot use.
	ErrIntervalNotSupported = errors.New("seats are billed monthly or yearly")
)

// MaxQuantityPerOrder bounds one seat order.
const MaxQuantityPerOrder = 1000

// ParseType normalizes a seat type. An empty value is Unassigned.
func ParseType(raw string) (Type, error) {
	value := Type(strings.ToLower(strings.TrimSpace(raw)))
	switch value {
	case "", "none":
		return Unassigned, nil
	case Free, Pro, Max, Unassigned:
		return value, nil
	default:
		return "", ErrInvalidSeatType
	}
}

// Paid reports whether a seat type is bought.
func (t Type) Paid() bool {
	return t == Pro || t == Max
}

// Rank orders seat types from no seat to Max.
func (t Type) Rank() int {
	switch t {
	case Free:
		return 1
	case Pro:
		return 2
	case Max:
		return 3
	default:
		return 0
	}
}

// Offer returns the seat type's offer; Unassigned has none.
func (t Type) Offer() (Offer, bool) {
	offer, ok := Offers[t]

	return offer, ok
}

// MonthlyGrantMC is the seat's monthly (or lifetime) grant in millicredits.
func (t Type) MonthlyGrantMC() int64 {
	offer, ok := t.Offer()
	if !ok {
		return 0
	}

	return offer.MonthlyCredits * credits.MillicreditsPerCredit
}

// MonthlyPriceIRR is one seat-month before VAT: the seat's credits at the list
// price less the seat discount.
func MonthlyPriceIRR(offer Offer, listPriceIRRPerCredit int64) int64 {
	if offer.Lifetime || offer.MonthlyCredits <= 0 || listPriceIRRPerCredit <= 0 {
		return 0
	}

	return offer.MonthlyCredits * listPriceIRRPerCredit * (100 - offer.DiscountPercent) / 100
}

// PeriodPriceIRR is one seat for a whole billing period before VAT. Yearly
// billing gets the seat type's yearly discount; credits still arrive monthly.
func PeriodPriceIRR(offer Offer, listPriceIRRPerCredit int64, interval string) (int64, error) {
	monthly := MonthlyPriceIRR(offer, listPriceIRRPerCredit)

	switch NormalizeInterval(interval) {
	case "monthly":
		return monthly, nil
	case "yearly":
		return monthly * 12 * (100 - offer.YearlyDiscountPercent) / 100, nil
	default:
		return 0, ErrIntervalNotSupported
	}
}

// NormalizeInterval maps interval aliases to monthly or yearly.
func NormalizeInterval(interval string) string {
	switch strings.ToLower(strings.TrimSpace(interval)) {
	case "", "month", "monthly":
		return "monthly"
	case "year", "yearly", "annual", "annually":
		return "yearly"
	default:
		return strings.ToLower(strings.TrimSpace(interval))
	}
}

// Proration is the share of a billing period a mid-period purchase pays for.
type Proration struct {
	RemainingDays int
	PeriodDays    int
	AmountIRR     int64
}

// Prorate charges fullPriceIRR for the days left in [periodStart, periodEnd)
// at now. A partial day counts as a whole day, and the amount is rounded up to
// the rial, so a seat bought on the last day still costs one day.
func Prorate(fullPriceIRR int64, periodStart, periodEnd, now time.Time) Proration {
	periodDays := wholeDays(periodEnd.Sub(periodStart))
	if periodDays <= 0 || fullPriceIRR <= 0 {
		return Proration{}
	}

	if now.Before(periodStart) {
		now = periodStart
	}

	remaining := wholeDays(periodEnd.Sub(now))
	if remaining <= 0 {
		return Proration{PeriodDays: periodDays}
	}

	remaining = min(remaining, periodDays)
	amount := decimal.NewFromInt(fullPriceIRR).Mul(decimal.NewFromInt(int64(remaining))).
		Div(decimal.NewFromInt(int64(periodDays))).Ceil()

	return Proration{RemainingDays: remaining, PeriodDays: periodDays, AmountIRR: amount.IntPart()}
}

func wholeDays(d time.Duration) int {
	if d <= 0 {
		return 0
	}

	days := d / (24 * time.Hour)
	if d%(24*time.Hour) != 0 {
		days++
	}

	return int(days)
}

// Timing says when a seat change takes effect.
type Timing string

// Seat change timings.
const (
	Immediate   Timing = "immediate"
	AtPeriodEnd Timing = "period_end"
)

// Change is a decided seat change.
type Change struct {
	From   Type
	To     Type
	Timing Timing
	// NeedsPaidSeat is set when the change takes a paid seat from the tenant's stock.
	NeedsPaidSeat bool
}

// DecideChange applies the seat change rules: upgrades are immediate, paid
// downgrades wait for the end of the billing period, a paid seat never goes
// back to free, and a member gets the free seat once.
func DecideChange(from, to Type, freeSeatUsed bool) (Change, error) {
	if _, err := ParseType(string(from)); err != nil {
		return Change{}, err
	}

	if _, err := ParseType(string(to)); err != nil {
		return Change{}, err
	}

	change := Change{From: from, To: to, Timing: Immediate}

	switch {
	case from == to:
		return Change{}, ErrSameSeat
	case to == Free && from.Paid():
		return Change{}, ErrBackToFree
	case to == Free && freeSeatUsed:
		return Change{}, ErrFreeSeatUsed
	case to.Rank() > from.Rank():
		change.NeedsPaidSeat = to.Paid()
	case from.Paid():
		change.Timing = AtPeriodEnd
	}

	return change, nil
}

// Enforced reports whether seat limits apply to a tenant: they apply to every
// active or trialing subscription.
func Enforced(subscriptionStatus string) bool {
	switch strings.TrimSpace(subscriptionStatus) {
	case "active", "trialing":
		return true
	default:
		return false
	}
}

// PoolAccess reports whether a member's chat usage may fall through to the
// tenant pool. Before enforcement everyone keeps pool access.
func PoolAccess(seat Type, enforced bool) bool {
	if !enforced {
		return true
	}

	offer, ok := seat.Offer()

	return ok && offer.PoolAccess
}

// BlockReason is the client-safe reason a chat request was refused.
type BlockReason string

// Block reasons, in the order clients report them.
const (
	BlockSeatUnassigned       BlockReason = "seat_unassigned"
	BlockSeatCreditsExhausted BlockReason = "seat_credits_exhausted"
	BlockPersonalLimitReached BlockReason = "personal_limit_reached"
	BlockBudgetRuleReached    BlockReason = "budget_rule_reached"
	BlockPoolExhausted        BlockReason = "pool_exhausted"
)

// Grant keys make each seat grant happen at most once.

// GrantKey is the idempotency key of one monthly paid-seat grant. It names the
// seat type, so a member who upgrades mid-slice gets the higher seat's full
// quota for that slice.
func GrantKey(tenantID, userID string, seat Type, sliceStart time.Time) string {
	return fmt.Sprintf("seat_grant:%s:%s:%s:%s", tenantID, userID, seat, sliceStart.UTC().Format(time.RFC3339))
}

// FreeGrantKey is the idempotency key of a member's one free-seat grant.
func FreeGrantKey(tenantID, userID string) string {
	return fmt.Sprintf("seat_free:%s:%s", tenantID, userID)
}

// WindowKey is the Redis admission window of a member's seat credits for one
// calendar month, alongside the tenant pool window of the same month.
func WindowKey(tenantID, userID, windowID string) string {
	return fmt.Sprintf("credit:seat:%s:%s:window:%s", strings.TrimSpace(tenantID), strings.TrimSpace(userID), strings.TrimSpace(windowID))
}

// ReservationKey is the reservation of one request in a seat window.
func ReservationKey(tenantID, userID, reservationID string) string {
	return fmt.Sprintf("credit:seat:%s:%s:reservation:%s", strings.TrimSpace(tenantID), strings.TrimSpace(userID), strings.TrimSpace(reservationID))
}

// PoolLimitWindowKey is the Redis window of a member's pool usage under the
// tenant's default personal limit for their seat type.
func PoolLimitWindowKey(tenantID, userID, windowID string) string {
	return fmt.Sprintf("credit:seat_pool:%s:%s:window:%s", strings.TrimSpace(tenantID), strings.TrimSpace(userID), strings.TrimSpace(windowID))
}

// Overflow is the part of amountMC committed to a window that went past its
// allowance, given the window's committed total after the commit.
func Overflow(allowanceMC, committedAfterMC, amountMC int64) int64 {
	before := committedAfterMC - amountMC
	over := max(0, committedAfterMC-allowanceMC) - max(0, before-allowanceMC)

	return min(max(over, 0), amountMC)
}

// LifecycleResult reports one pass of the seat lifecycle worker.
type LifecycleResult struct {
	AutoAssigned      int
	DowngradesApplied int
	LapsedAssignments int
	Granted           int
	GrantedMC         int64
}
