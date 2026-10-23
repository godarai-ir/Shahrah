package credits

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/shopspring/decimal"
)

// DefaultAlertThresholds are the budget-rule alert thresholds, in percent of
// the limit, used when a rule sets none.
var DefaultAlertThresholds = []int{80, 100}

// DefaultOverageCeilingCredits bounds overage for tenants that enabled it with
// no cap, so enforcement never allows unlimited spend.
const DefaultOverageCeilingCredits = 100_000

var (
	// ErrInvalidCreditAmount is returned for a credit amount that is not a
	// positive decimal with at most two decimals.
	ErrInvalidCreditAmount = errors.New("credit amount must be a positive number with at most two decimals")
	// ErrInvalidAlertThreshold is returned for a threshold outside 1..1000 percent.
	ErrInvalidAlertThreshold = errors.New("alert thresholds must be whole percentages between 1 and 1000")
)

// PoolWindowKey is the Redis window of a credit tenant's balance for one
// calendar month. The credit: prefix keeps it apart from IRR wallet windows.
func PoolWindowKey(tenantID, windowID string) string {
	return fmt.Sprintf("credit:tenant:%s:window:%s", strings.TrimSpace(tenantID), strings.TrimSpace(windowID))
}

// PoolReservationKey is the reservation of one request in the pool window.
func PoolReservationKey(tenantID, reservationID string) string {
	return fmt.Sprintf("credit:tenant:%s:reservation:%s", strings.TrimSpace(tenantID), strings.TrimSpace(reservationID))
}

// PolicyWindowKey is the Redis window of one budget rule and scope value.
func PolicyWindowKey(tenantID, ruleID, scopeHash, windowID string) string {
	return fmt.Sprintf("credit_policy:%s:rule:%s:scope:%s:window:%s",
		strings.TrimSpace(tenantID), strings.TrimSpace(ruleID), strings.TrimSpace(scopeHash), strings.TrimSpace(windowID))
}

// PolicyReservationKey is the reservation of one request in a rule window.
func PolicyReservationKey(tenantID, reservationID, targetHash string) string {
	return fmt.Sprintf("credit_reservation:%s:%s:%s",
		strings.TrimSpace(tenantID), strings.TrimSpace(reservationID), strings.TrimSpace(targetHash))
}

// AlertDedupeKey marks one alert threshold as already reported for a window.
func AlertDedupeKey(windowKey string, threshold int) string {
	return fmt.Sprintf("credit_alert:%s:%d", strings.TrimSpace(windowKey), threshold)
}

// PoolWindow returns the calendar-month window id and its end for a time.
func PoolWindow(now time.Time) (string, time.Time) {
	utc := now.UTC()
	start := time.Date(utc.Year(), utc.Month(), 1, 0, 0, 0, 0, time.UTC)

	return start.Format("200601"), start.AddDate(0, 1, 0)
}

// PoolAllowance holds the inputs of a credit pool window's allowance.
type PoolAllowance struct {
	AvailableMC    int64
	DebtMC         int64
	SettledMC      int64
	OverageEnabled bool
	// OverageCapMC is the tenant's overage cap; nil means no cap.
	OverageCapMC     *int64
	OverageCeilingMC int64
}

// AllowanceMC is the pool window allowance. Consumption already settled into
// the ledger in this window is added back, because the window also counts it
// as committed: remaining = balance − reserved − (committed − settled).
func (p PoolAllowance) AllowanceMC() int64 {
	allowance := p.AvailableMC - p.DebtMC + p.SettledMC + p.overageHeadroomMC()
	if allowance < 0 {
		return 0
	}

	return allowance
}

func (p PoolAllowance) overageHeadroomMC() int64 {
	if !p.OverageEnabled {
		return 0
	}

	limit := p.OverageCeilingMC
	if p.OverageCapMC != nil {
		limit = *p.OverageCapMC
	}

	if limit <= p.DebtMC {
		return 0
	}

	return limit - p.DebtMC
}

// ParseCreditAmountMC parses a positive credit amount with at most two
// decimals, such as "12.50", into millicredits.
func ParseCreditAmountMC(raw string) (int64, error) {
	value, err := decimal.NewFromString(strings.TrimSpace(raw))
	if err != nil || !value.IsPositive() || !value.Equal(value.Round(2)) {
		return 0, ErrInvalidCreditAmount
	}

	return value.Mul(decimal.NewFromInt(MillicreditsPerCredit)).IntPart(), nil
}

// NormalizeAlertThresholds validates thresholds and returns them sorted and
// deduplicated. An empty list means the defaults.
func NormalizeAlertThresholds(thresholds []int) ([]int, error) {
	if len(thresholds) == 0 {
		return append([]int(nil), DefaultAlertThresholds...), nil
	}

	seen := make(map[int]bool, len(thresholds))
	out := make([]int, 0, len(thresholds))

	for _, threshold := range thresholds {
		if threshold < 1 || threshold > 1000 {
			return nil, ErrInvalidAlertThreshold
		}

		if seen[threshold] {
			continue
		}

		seen[threshold] = true
		out = append(out, threshold)
	}

	sort.Ints(out)

	return out, nil
}

// CrossedThresholds returns the thresholds that spend crossed when it grew from
// beforeMC to afterMC against limitMC.
func CrossedThresholds(thresholds []int, limitMC, beforeMC, afterMC int64) []int {
	if limitMC <= 0 || afterMC <= beforeMC {
		return nil
	}

	var crossed []int

	for _, threshold := range thresholds {
		mark := decimal.NewFromInt(limitMC).Mul(decimal.NewFromInt(int64(threshold))).Div(decimal.NewFromInt(100))
		if decimal.NewFromInt(beforeMC).LessThan(mark) && !decimal.NewFromInt(afterMC).LessThan(mark) {
			crossed = append(crossed, threshold)
		}
	}

	return crossed
}

var (
	// ErrInvalidOnExceed is returned for an unknown exceed behavior.
	ErrInvalidOnExceed = errors.New("on_exceed must be block or alert")
	// ErrInconsistentOnExceed is returned when block_on_budget_exceed and
	// on_exceed disagree in one request.
	ErrInconsistentOnExceed = errors.New("block_on_budget_exceed contradicts on_exceed")
)

// ResolveOnExceed decides a rule's exceed behavior from the new on_exceed
// field and the legacy block_on_budget_exceed flag. Either may be omitted; the
// legacy flag maps true to block and false to alert. current is the stored
// value used when both are omitted.
func ResolveOnExceed(onExceed *platform.BudgetOnExceed, block *bool, current platform.BudgetOnExceed) (platform.BudgetOnExceed, error) {
	if onExceed == nil {
		switch {
		case block == nil && current != "":
			return current, nil
		case block == nil || *block:
			return platform.BudgetOnExceedBlock, nil
		default:
			return platform.BudgetOnExceedAlert, nil
		}
	}

	value := platform.BudgetOnExceed(strings.TrimSpace(string(*onExceed)))
	switch value {
	case platform.BudgetOnExceedBlock, platform.BudgetOnExceedAlert:
	default:
		return "", ErrInvalidOnExceed
	}

	if block != nil && *block != OnExceedBlocks(value) {
		return "", ErrInconsistentOnExceed
	}

	return value, nil
}

// OnExceedBlocks reports whether an exceed behavior can refuse a request, which
// is what the legacy block_on_budget_exceed flag means.
func OnExceedBlocks(onExceed platform.BudgetOnExceed) bool {
	return onExceed != platform.BudgetOnExceedAlert
}
