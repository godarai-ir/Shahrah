package billing

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
)

// identityScopes are the applies-per scopes that one caller has a single
// value for. Rules applied per model or per metadata value have one window per
// value and are not reported.
var identityScopes = map[string]bool{
	string(platform.BudgetAppliesPerTenant):         true,
	string(platform.BudgetAppliesPerUser):           true,
	string(platform.BudgetAppliesPerTeam):           true,
	string(platform.BudgetAppliesPerVirtualAccount): true,
}

// BudgetStatus reports the credit budget rules that apply to the caller and,
// when includeBalance is set, the tenant's spendable balance. It reads only
// the cached windows and the compiled profile, never the ledger.
func (s *BudgetService) BudgetStatus(ctx context.Context, auth *platform.AuthContext, includeBalance bool) (*platform.BudgetStatus, error) {
	if auth == nil || strings.TrimSpace(auth.TenantID) == "" {
		return nil, ErrBudgetUnavailable
	}

	profile, err := s.tenantProfile(ctx, auth.TenantID)
	if err != nil {
		return nil, err
	}

	if profile == nil {
		return nil, ErrBudgetUnavailable
	}

	inspector, ok := s.budgetWindows.(cache.BudgetWindowInspector)
	if !ok {
		return nil, ErrBudgetUnavailable
	}

	now := time.Now().UTC()
	status := &platform.BudgetStatus{Object: platform.BudgetStatusObject, Unit: platform.CreditUnit, Rules: []platform.BudgetRuleStatus{}}

	for _, rule := range profile.BudgetRules {
		ruleStatus, applies, statusErr := s.ruleStatus(ctx, inspector, rule, auth, now)
		if statusErr != nil {
			return nil, statusErr
		}

		if applies {
			status.Rules = append(status.Rules, ruleStatus)
		}
	}

	if includeBalance {
		balance, balanceErr := s.balanceStatus(ctx, inspector, auth.TenantID, now)
		if balanceErr != nil {
			return nil, balanceErr
		}

		status.TenantBalance = balance
	}

	if reader, ok := s.seatCredits.(seatStatusReader); ok && strings.TrimSpace(auth.UserID) != "" && profile.Seats != nil {
		seat, seatErr := reader.GetSeatStatus(ctx, auth.TenantID, auth.UserID, now)
		if seatErr == nil {
			status.Seat = seat
		}
	}

	return status, nil
}

// seatStatusReader reports a member's own seat.
type seatStatusReader interface {
	GetSeatStatus(ctx context.Context, tenantID, userID string, now time.Time) (*platform.SeatStatus, error)
}

func (s *BudgetService) ruleStatus(ctx context.Context, inspector cache.BudgetWindowInspector, rule platform.BudgetLimitRule, auth *platform.AuthContext, now time.Time) (platform.BudgetRuleStatus, bool, error) {
	if !rule.Enabled || rule.LimitMC == nil || *rule.LimitMC <= 0 {
		return platform.BudgetRuleStatus{}, false, nil
	}

	scope, scopeValue, ok := ruleIdentityScope(rule, auth)
	if !ok || !budgetRuleAppliesToIdentity(rule, auth) {
		return platform.BudgetRuleStatus{}, false, nil
	}

	// A rule with an unknown window unit is never enforced, so it is not reported.
	windowID, windowEnd, unitErr := budgetPolicyWindowForTime(rule.Unit, now)
	if unitErr != nil {
		return platform.BudgetRuleStatus{}, false, nil //nolint:nilerr // An unenforced rule is not an error for the caller.
	}

	ruleID := ruleIdentifier(rule)
	windowKey := credits.PolicyWindowKey(auth.TenantID, ruleID, shortHash(scope+":"+scopeValue), windowID)

	snapshot, err := inspector.GetBudgetWindow(ctx, windowKey)
	if err != nil {
		return platform.BudgetRuleStatus{}, false, errors.Join(ErrBudgetUnavailable, err)
	}

	var spentMC, reservedMC int64
	if snapshot != nil {
		spentMC, reservedMC = int64(snapshot.CommittedIRR), int64(snapshot.ReservedIRR)
	}

	onExceed := rule.OnExceed
	if onExceed == "" {
		onExceed = platform.BudgetOnExceedBlock
	}

	return platform.BudgetRuleStatus{
		RuleID:    ruleID,
		Name:      rule.Name,
		Scope:     scope,
		Limit:     credits.Format(credits.Millicredits(*rule.LimitMC)),
		Spent:     credits.Format(credits.Millicredits(spentMC)),
		Remaining: credits.Format(credits.Millicredits(max(0, *rule.LimitMC-spentMC-reservedMC))),
		WindowEnd: windowEnd,
		OnExceed:  string(onExceed),
	}, true, nil
}

func (s *BudgetService) balanceStatus(ctx context.Context, inspector cache.BudgetWindowInspector, tenantID string, now time.Time) (*platform.BudgetBalanceStatus, error) {
	windowID, windowEnd := credits.PoolWindow(now)

	snapshot, err := inspector.GetBudgetWindow(ctx, credits.PoolWindowKey(tenantID, windowID))
	if err != nil {
		return nil, errors.Join(ErrBudgetUnavailable, err)
	}

	var remainingMC, reservedMC int64

	switch {
	case snapshot != nil:
		remainingMC = int64(snapshot.AllowanceIRR - snapshot.CommittedIRR - snapshot.ReservedIRR)
		reservedMC = int64(snapshot.ReservedIRR)
	case s.creditPools != nil:
		// No request this month yet: nothing is committed or reserved.
		if remainingMC, err = s.creditPools.CreditPoolAllowanceMC(ctx, tenantID, windowID); err != nil {
			return nil, errors.Join(ErrBudgetUnavailable, err)
		}
	default:
		return nil, ErrBudgetUnavailable
	}

	return &platform.BudgetBalanceStatus{
		Remaining: credits.Format(credits.Millicredits(max(0, remainingMC))),
		Reserved:  credits.Format(credits.Millicredits(reservedMC)),
		WindowEnd: windowEnd,
	}, nil
}

// ruleIdentityScope returns the rule's applies-per scope and the caller's value
// for it, when that scope identifies the caller.
func ruleIdentityScope(rule platform.BudgetLimitRule, auth *platform.AuthContext) (string, string, bool) {
	scope := string(platform.BudgetAppliesPerTenant)
	if rule.BudgetAppliesPer != nil && strings.TrimSpace(*rule.BudgetAppliesPer) != "" {
		scope = strings.TrimSpace(*rule.BudgetAppliesPer)
	}

	if !identityScopes[scope] {
		return "", "", false
	}

	value, known := budgetContextValue(scope, auth, platform.UsageCostBreakdown{})
	if !known || strings.TrimSpace(value) == "" {
		return "", "", false
	}

	return scope, value, true
}

// budgetRuleAppliesToIdentity evaluates only a rule's identity conditions.
// Conditions on the request (model, metadata, endpoint) cannot be known
// without a request, so a rule that matches the caller is reported whatever
// those conditions say.
func budgetRuleAppliesToIdentity(rule platform.BudgetLimitRule, auth *platform.AuthContext) bool {
	for key, expected := range rule.Conditions {
		switch key {
		case "combinator":
			continue
		case "conditions":
			if !structuredIdentityConditionsMatch(expected, auth) {
				return false
			}
		default:
			actual, known := budgetContextValue(key, auth, platform.UsageCostBreakdown{})
			if known && isIdentityConditionKey(key) && !conditionValueMatches(actual, expected) {
				return false
			}
		}
	}

	return true
}

func structuredIdentityConditionsMatch(raw interface{}, auth *platform.AuthContext) bool {
	conditions, ok := raw.([]interface{})
	if !ok {
		return true
	}

	for _, item := range conditions {
		condition, ok := item.(map[string]interface{})
		if !ok || strings.TrimSpace(asString(condition["field"])) != "subject" {
			continue
		}

		matches, err := budgetStructuredConditionMatches(condition, auth, platform.UsageCostBreakdown{})
		if err == nil && !matches {
			return false
		}
	}

	return true
}

func isIdentityConditionKey(key string) bool {
	switch strings.TrimSpace(key) {
	case "tenant_id", "user_id", "team_id", "virtual_account_id", "virtual_account",
		"key_id", "key_alias", "key_type", "role", "virtual_user_id",
		string(platform.BudgetAppliesPerUser), string(platform.BudgetAppliesPerTeam),
		string(platform.BudgetAppliesPerVirtualAccount):
		return true
	default:
		return false
	}
}
