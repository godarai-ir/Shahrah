package billing

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	metricsobs "github.com/Delpak/shahrah/services/control-plane/internal/observability/metrics"
	"github.com/Delpak/shahrah/services/control-plane/internal/runtimeprofile"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// BudgetScopeTenantBalance is the error scope of a refused request when the
// tenant's credit balance, rather than a budget rule, ran out.
const BudgetScopeTenantBalance = "tenant_balance"

// alertDedupeTTL outlives the longest rule window. The dedupe key names its
// window, so a longer TTL never suppresses the next window's alerts.
const alertDedupeTTL = 32 * 24 * time.Hour

// BudgetExceededError reports which budget refused a request. It matches
// ErrBudgetExceeded, and Scope is safe to show clients.
type BudgetExceededError struct {
	Scope string
}

func (e *BudgetExceededError) Error() string {
	return "budget exceeded"
}

// Is lets errors.Is(err, ErrBudgetExceeded) match.
func (e *BudgetExceededError) Is(target error) bool {
	return target == ErrBudgetExceeded
}

// BudgetErrorScope returns the client-safe scope of a budget refusal, or "".
func BudgetErrorScope(err error) string {
	var exceeded *BudgetExceededError
	if errors.As(err, &exceeded) {
		return exceeded.Scope
	}

	return ""
}

// CreditPoolStore loads the allowance of a credit tenant's pool window.
type CreditPoolStore interface {
	CreditPoolAllowanceMC(ctx context.Context, tenantID, windowID string) (int64, error)
}

// budgetSpendRecorder adds already-served spend to a window without admission.
type budgetSpendRecorder interface {
	RecordBudgetSpend(ctx context.Context, req cache.BudgetSpendRequest) (*cache.BudgetWindowResult, error)
}

// isCreditTenant reports whether the compiled profile bills in credits.
func isCreditTenant(profile *runtimeprofile.Profile) bool {
	return profile != nil && strings.TrimSpace(profile.Tenant.BillingUnit) == credits.BillingUnitCredit
}

// ReserveExternalBilledBudget reserves an externally billed (BYOK) request. IRR
// tenants are not reserved, as before. Credit tenants reserve the full credit
// equivalent against their budget rules and only the platform fee against
// their balance.
func (s *BudgetService) ReserveExternalBilledBudget(
	ctx context.Context,
	auth *platform.AuthContext,
	breakdown platform.UsageCostBreakdown,
	requestID string,
) (*platform.BudgetReservation, error) {
	if auth == nil || strings.TrimSpace(auth.TenantID) == "" {
		return nil, ErrBudgetExceeded
	}

	profile, err := s.tenantProfile(ctx, auth.TenantID)
	if err != nil {
		return nil, err
	}

	if !isCreditTenant(profile) {
		return nil, nil
	}

	return s.reserveCreditBudget(ctx, auth, breakdown, budgetReservationID(requestID), profile, platform.BillingModeExternalBilled)
}

func (s *BudgetService) tenantProfile(ctx context.Context, tenantID string) (*runtimeprofile.Profile, error) {
	if s.runtimeProfiles == nil {
		return nil, nil
	}

	profile, err := s.runtimeProfiles.GetSnapshot(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("%w: profile budget rules unavailable: %v", ErrBudgetUnavailable, err)
	}

	return profile, nil
}

func (s *BudgetService) reserveCreditBudget(
	ctx context.Context,
	auth *platform.AuthContext,
	breakdown platform.UsageCostBreakdown,
	reservationID string,
	profile *runtimeprofile.Profile,
	billingMode string,
) (*platform.BudgetReservation, error) {
	// Serving a credit tenant without credit enforcement would let it spend
	// without limit, so it is refused until enforcement is turned on.
	if !s.creditEnforcementEnabled || s.budgetWindows == nil || s.creditPools == nil {
		return nil, ErrBudgetUnavailable
	}

	tenantID := strings.TrimSpace(auth.TenantID)
	now := time.Now().UTC()

	estimateMC, err := s.creditAmountMC(breakdown)
	if err != nil {
		return nil, err
	}

	reservation := &platform.BudgetReservation{
		ID: reservationID, TenantID: tenantID, Unit: credits.BillingUnitCredit, BillingMode: billingMode, CreatedAt: now,
	}

	poolMC := estimateMC

	if billingMode == platform.BillingModeExternalBilled {
		fee := byokFeePercent(profile)
		reservation.BYOKFeePercent = fee.String()
		poolMC = credits.BYOKFeeMC(estimateMC, fee)
	}

	targets := append([]budgetReservationTarget{s.creditPoolTarget(tenantID, reservationID, poolMC, now)},
		s.creditPolicyTargets(profile, auth, breakdown, reservationID, estimateMC, now)...)

	reserved, err := s.reserveCreditTargets(ctx, targets, reservationID, now)
	if err != nil {
		return nil, err
	}

	reservation.Targets = make([]platform.BudgetReservationTarget, 0, len(reserved))
	for _, target := range reserved {
		reservation.Targets = append(reservation.Targets, target.BudgetReservationTarget)
	}

	primary := reservation.Targets[0]
	reservation.WindowID, reservation.WindowKey, reservation.ReservationKey = primary.WindowID, primary.WindowKey, primary.ReservationKey

	return reservation, nil
}

// reserveCreditTargets reserves blocking targets and seeds alert-only rule
// windows, which are charged at commit without an admission check.
func (s *BudgetService) reserveCreditTargets(ctx context.Context, targets []budgetReservationTarget, reservationID string, now time.Time) ([]budgetReservationTarget, error) {
	reserved := make([]budgetReservationTarget, 0, len(targets))

	for _, target := range targets {
		if target.OnExceed == string(platform.BudgetOnExceedAlert) {
			if err := s.seedBudgetTarget(ctx, s.budgetWindows, target); err != nil {
				return nil, errors.Join(err, s.releaseBudgetTargetsBestEffort(ctx, s.budgetWindows, reserved, now))
			}

			reserved = append(reserved, target)

			continue
		}

		result, err := s.reserveBudgetTarget(ctx, s.budgetWindows, target, reservationID, now)
		if err == nil && result == nil {
			err = ErrBudgetUnavailable
		}

		if err == nil && !result.Allowed {
			err = ErrBudgetUnavailable

			if result.State == "exhausted" {
				scope := creditTargetScope(target)
				metricsobs.ObserveCreditBudgetRefusal(scope)
				err = &BudgetExceededError{Scope: scope}
			}
		}

		if err != nil {
			return nil, errors.Join(err, s.releaseBudgetTargetsBestEffort(ctx, s.budgetWindows, reserved, now))
		}

		reserved = append(reserved, target)
	}

	return reserved, nil
}

func creditTargetScope(target budgetReservationTarget) string {
	if target.Kind == budgetTargetTenantCredit {
		return BudgetScopeTenantBalance
	}

	return target.Scope
}

func (s *BudgetService) creditPoolTarget(tenantID, reservationID string, amountMC int64, now time.Time) budgetReservationTarget {
	windowID, windowEnd := credits.PoolWindow(now)

	return budgetReservationTarget{
		tenantID: tenantID,
		BudgetReservationTarget: platform.BudgetReservationTarget{
			Kind:           budgetTargetTenantCredit,
			Scope:          string(platform.BudgetAppliesPerTenant),
			ScopeValue:     tenantID,
			WindowID:       windowID,
			WindowKey:      credits.PoolWindowKey(tenantID, windowID),
			ReservationKey: credits.PoolReservationKey(tenantID, reservationID),
			EstimatedIRR:   float64(amountMC),
			EstimatedMC:    amountMC,
			CreatedAt:      now,
		},
		expiresAt:      windowEnd,
		ttl:            s.windowTTL(windowEnd),
		seedCreditPool: true,
	}
}

// creditPolicyTargets returns one target per enabled credit-limited rule that
// matches the request. Rules that cannot be evaluated are skipped, never
// enforced against the wrong scope.
func (s *BudgetService) creditPolicyTargets(profile *runtimeprofile.Profile, auth *platform.AuthContext, breakdown platform.UsageCostBreakdown, reservationID string, amountMC int64, now time.Time) []budgetReservationTarget {
	if profile == nil || len(profile.BudgetRules) == 0 {
		return nil
	}

	targets := make([]budgetReservationTarget, 0, len(profile.BudgetRules))

	for _, rule := range profile.BudgetRules {
		if !rule.Enabled || rule.LimitMC == nil || *rule.LimitMC <= 0 {
			continue
		}

		target, ok := s.creditPolicyTarget(rule, auth, breakdown, reservationID, amountMC, now)
		if ok {
			targets = append(targets, target)
		}
	}

	return targets
}

func (s *BudgetService) creditPolicyTarget(rule platform.BudgetLimitRule, auth *platform.AuthContext, breakdown platform.UsageCostBreakdown, reservationID string, amountMC int64, now time.Time) (budgetReservationTarget, bool) {
	matches, err := budgetRuleMatches(rule, auth, breakdown)
	if err != nil || !matches {
		return budgetReservationTarget{}, false
	}

	scope, scopeValue, err := budgetRuleScope(rule, auth, breakdown)
	if err != nil {
		return budgetReservationTarget{}, false
	}

	windowID, windowEnd, err := budgetPolicyWindowForTime(rule.Unit, now)
	if err != nil {
		return budgetReservationTarget{}, false
	}

	ruleID := ruleIdentifier(rule)
	windowKey := credits.PolicyWindowKey(auth.TenantID, ruleID, shortHash(scope+":"+scopeValue), windowID)

	onExceed := rule.OnExceed
	if onExceed == "" {
		onExceed = platform.BudgetOnExceedBlock
	}

	return budgetReservationTarget{
		tenantID: auth.TenantID,
		BudgetReservationTarget: platform.BudgetReservationTarget{
			Kind:           budgetTargetPolicy,
			RuleID:         ruleID,
			Scope:          scope,
			ScopeValue:     scopeValue,
			WindowID:       windowID,
			WindowKey:      windowKey,
			ReservationKey: credits.PolicyReservationKey(auth.TenantID, reservationID, shortHash(windowKey)),
			EstimatedIRR:   float64(amountMC),
			EstimatedMC:    amountMC,
			OnExceed:       string(onExceed),
			CreatedAt:      now,
		},
		allowanceIRR: float64(*rule.LimitMC),
		expiresAt:    windowEnd,
		ttl:          s.windowTTL(windowEnd),
	}, true
}

func (s *BudgetService) windowTTL(windowEnd time.Time) time.Duration {
	ttl := time.Until(windowEnd.Add(s.budgetWindowTTLGrace))
	if ttl <= 0 {
		ttl = s.budgetWindowTTLGrace
	}

	return ttl
}

func (s *BudgetService) seedCreditPoolWindow(ctx context.Context, budgetCache BudgetWindowCache, target budgetReservationTarget) error {
	allowanceMC, err := s.creditPools.CreditPoolAllowanceMC(ctx, target.tenantID, target.WindowID)
	if err != nil {
		return fmt.Errorf("%w: credit pool seed failed: %v", ErrBudgetUnavailable, err)
	}

	result, err := budgetCache.SeedBudgetWindow(ctx, cache.BudgetWindowSeed{
		WindowKey:    target.WindowKey,
		AllowanceIRR: float64(allowanceMC),
		ExpiresAt:    target.expiresAt,
		TTL:          target.ttl,
		Version:      target.expiresAt.Unix(),
	})
	if err != nil {
		return fmt.Errorf("%w: credit pool seed failed: %v", ErrBudgetUnavailable, err)
	}

	if result == nil || !result.Allowed {
		return ErrBudgetUnavailable
	}

	return nil
}

// commitCreditBudget commits the actual credits of a served request. The usage
// already happened, so commits are forced: spend beyond an allowance is
// recorded as overage instead of refused, and the ledger debits it as debt.
func (s *BudgetService) commitCreditBudget(ctx context.Context, reservation *platform.BudgetReservation, breakdown platform.UsageCostBreakdown) error {
	actualMC, err := s.creditAmountMC(breakdown)
	if err != nil {
		return err
	}

	poolMC := actualMC

	if reservation.BillingMode == platform.BillingModeExternalBilled {
		fee, parseErr := decimal.NewFromString(reservation.BYOKFeePercent)
		if parseErr != nil {
			fee = credits.DefaultBYOKFeePercent
		}

		poolMC = credits.BYOKFeeMC(actualMC, fee)
	}

	thresholds := s.ruleAlertThresholds(ctx, reservation.TenantID)
	now := time.Now().UTC()

	var commitErrors []error

	for _, target := range reservation.Targets {
		amountMC := actualMC
		if target.Kind == budgetTargetTenantCredit {
			amountMC = poolMC
		}

		result, commitErr := s.commitCreditTarget(ctx, target, amountMC, now)
		if commitErr != nil {
			commitErrors = append(commitErrors, commitErr)

			continue
		}

		if target.Kind == budgetTargetPolicy {
			s.reportCrossedThresholds(ctx, reservation.TenantID, target, thresholds[target.RuleID], amountMC, result)
		}
	}

	if len(commitErrors) > 0 {
		return fmt.Errorf("%w: commit failed: %v", ErrBudgetUnavailable, errors.Join(commitErrors...))
	}

	return nil
}

func (s *BudgetService) commitCreditTarget(ctx context.Context, target platform.BudgetReservationTarget, amountMC int64, now time.Time) (*cache.BudgetWindowResult, error) {
	var (
		result *cache.BudgetWindowResult
		err    error
	)

	if target.OnExceed == string(platform.BudgetOnExceedAlert) {
		recorder, ok := s.budgetWindows.(budgetSpendRecorder)
		if !ok {
			return nil, errors.New("budget windows cannot record spend")
		}

		result, err = recorder.RecordBudgetSpend(ctx, cache.BudgetSpendRequest{WindowKey: target.WindowKey, Amount: float64(amountMC), Now: now})
	} else {
		result, err = s.budgetWindows.CommitBudget(ctx, cache.BudgetCommitRequest{
			WindowKey: target.WindowKey, ReservationKey: target.ReservationKey, ActualIRR: float64(amountMC), Force: true, Now: now,
		})
	}

	if err != nil {
		return nil, err
	}

	if result == nil || !result.Allowed {
		state := "unavailable"
		if result != nil {
			state = result.State
		}

		return nil, fmt.Errorf("window %s: %s", target.WindowKey, state)
	}

	if result.CommittedIRR > result.AllowanceIRR {
		metricsobs.ObserveCreditBudgetOverage(min(amountMC, int64(result.CommittedIRR-result.AllowanceIRR)))
	}

	return result, nil
}

// ruleAlertThresholds maps rule ids to their alert thresholds.
func (s *BudgetService) ruleAlertThresholds(ctx context.Context, tenantID string) map[string][]int {
	profile, err := s.tenantProfile(ctx, tenantID)
	if err != nil || profile == nil {
		return nil
	}

	out := make(map[string][]int, len(profile.BudgetRules))
	for _, rule := range profile.BudgetRules {
		thresholds, normalizeErr := credits.NormalizeAlertThresholds(rule.AlertThresholds)
		if normalizeErr == nil {
			out[ruleIdentifier(rule)] = thresholds
		}
	}

	return out
}

// reportCrossedThresholds records each alert threshold the commit crossed.
// There is no outbound notification channel (product decision P1), so an alert
// is a structured log line plus a metric, deduplicated per window.
func (s *BudgetService) reportCrossedThresholds(ctx context.Context, tenantID string, target platform.BudgetReservationTarget, thresholds []int, amountMC int64, result *cache.BudgetWindowResult) {
	afterMC := int64(result.CommittedIRR)

	for _, threshold := range credits.CrossedThresholds(thresholds, int64(result.AllowanceIRR), afterMC-amountMC, afterMC) {
		if s.cache != nil {
			first, err := s.cache.Increment(ctx, credits.AlertDedupeKey(target.WindowKey, threshold), 1, alertDedupeTTL)
			if err == nil && first > 1 {
				continue
			}
		}

		metricsobs.ObserveCreditBudgetAlert(threshold, target.OnExceed)
		s.logger.Warn("budget_alert",
			zap.String("tenant_id", tenantID), zap.String("rule_id", target.RuleID),
			zap.String("scope", target.Scope), zap.String("scope_value", target.ScopeValue),
			zap.String("window_id", target.WindowID), zap.Int("threshold_percent", threshold),
			zap.String("on_exceed", target.OnExceed))
	}
}

// creditAmountMC converts a priced breakdown to millicredits with the current
// rate card, rounded up once.
func (s *BudgetService) creditAmountMC(breakdown platform.UsageCostBreakdown) (int64, error) {
	card := credits.IdentityRateCard()
	if s.creditRateCards != nil {
		card = s.creditRateCards.CurrentRateCard()
	}

	charge, err := credits.FromBreakdown(breakdown, card)
	if err != nil {
		return 0, fmt.Errorf("%w: credit pricing failed: %v", ErrBudgetUnavailable, err)
	}

	return int64(charge.Millicredits), nil
}

// byokFeePercent reads credits.byok_fee_percent from the compiled entitlements.
func byokFeePercent(profile *runtimeprofile.Profile) decimal.Decimal {
	if profile == nil || profile.Entitlements == nil {
		return credits.DefaultBYOKFeePercent
	}

	raw, ok := profile.Entitlements.Features[credits.FeatureBYOKFeePercent]
	if !ok || raw == nil {
		return credits.DefaultBYOKFeePercent
	}

	fee, err := decimal.NewFromString(strings.TrimSpace(asString(raw)))
	if err != nil || fee.IsNegative() || fee.GreaterThan(decimal.NewFromInt(100)) {
		return credits.DefaultBYOKFeePercent
	}

	return fee
}
