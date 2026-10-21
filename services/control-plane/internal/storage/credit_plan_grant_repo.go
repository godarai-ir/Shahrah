package storage

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/jackc/pgx/v5"
)

// creditSubscription is a current subscription of a credit tenant.
type creditSubscription struct {
	SubscriptionID string
	TenantID       string
	Status         string
	PeriodStart    time.Time
	PeriodEnd      time.Time
	MonthlyCredits string
}

// GrantDuePlanCredits grants, for every active subscription of a credit
// tenant, the monthly slice of plan credits that contains now. Trialing
// tenants get no plan grant: their free seat grants are the trial credits.
// Keys make each grant happen at most once, so the pass is safe to repeat.
func (s *CreditLedgerRepository) GrantDuePlanCredits(ctx context.Context, now time.Time) (credits.GrantResult, error) {
	now = now.UTC()

	subscriptions, err := s.listCreditSubscriptions(ctx, now)
	if err != nil {
		return credits.GrantResult{}, err
	}

	result := credits.GrantResult{}

	for _, sub := range subscriptions {
		grant, ok := planGrantFor(sub, now)
		if !ok {
			continue
		}

		granted, overageMC, grantErr := s.grantDue(ctx, sub, grant)
		if grantErr != nil {
			return result, grantErr
		}

		result.OverageMC += overageMC

		if granted.Duplicate {
			result.Duplicates++

			continue
		}

		result.Granted++
		result.GrantedMC += grant.AmountMC
	}

	return result, nil
}

// grantDue writes one due grant. A plan slice first settles the tenant's debt
// as invoiced overage when overage is enabled; otherwise the grant repays it.
func (s *CreditLedgerRepository) grantDue(ctx context.Context, sub creditSubscription, grant CreditGrant) (CreditGrantResult, int64, error) {
	sliceStart, isSlice := grant.SourceRef["slice_start"].(time.Time)
	if grant.Kind != credits.EntryPlanGrant || !isSlice {
		result, err := s.GrantCredits(ctx, grant)

		return result, 0, err
	}

	if err := validateCreditGrant(grant); err != nil {
		return CreditGrantResult{}, 0, err
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return CreditGrantResult{}, 0, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	var granted bool
	if keyErr := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM credit_ledger_keys WHERE idempotency_key = $1)`, grant.IdempotencyKey).Scan(&granted); keyErr != nil {
		return CreditGrantResult{}, 0, keyErr
	}

	if granted {
		return CreditGrantResult{Duplicate: true}, 0, tx.Commit(ctx)
	}

	state, err := lockCreditTenantTx(ctx, tx, sub.TenantID)
	if err != nil {
		return CreditGrantResult{}, 0, err
	}

	overageMC, err := settleOverageTx(ctx, tx, sub, sliceStart, state.DebtMC, grant.OccurredAt)
	if err != nil {
		return CreditGrantResult{}, 0, err
	}

	result, err := grantCreditsTx(ctx, tx, grant)
	if err != nil {
		return CreditGrantResult{}, 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return CreditGrantResult{}, 0, err
	}

	s.invalidateCreditTenant(ctx, sub.TenantID, "")

	return result, overageMC, nil
}

// planGrantFor decides the grant a subscription is due at now, if any.
func planGrantFor(sub creditSubscription, now time.Time) (CreditGrant, bool) {
	ref := map[string]interface{}{"subscription_id": sub.SubscriptionID}

	monthly, err := strconv.ParseInt(strings.TrimSpace(sub.MonthlyCredits), 10, 64)
	if err != nil || monthly <= 0 {
		return CreditGrant{}, false
	}

	slice, ok := credits.CurrentPlanSlice(sub.PeriodStart, sub.PeriodEnd, now)
	if !ok {
		return CreditGrant{}, false
	}

	ref["slice_start"] = slice.Start
	ref["slice_end"] = slice.End
	expiresAt := slice.End

	return CreditGrant{
		TenantID: sub.TenantID, Kind: credits.EntryPlanGrant, Source: credits.LotSourcePlan,
		AmountMC: monthly * credits.MillicreditsPerCredit, ExpiresAt: &expiresAt, OccurredAt: now,
		IdempotencyKey: credits.PlanGrantKey(sub.SubscriptionID, slice.Start), SourceRef: ref,
	}, true
}

func (s *CreditLedgerRepository) listCreditSubscriptions(ctx context.Context, now time.Time) ([]creditSubscription, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT ts.subscription_id::text, ts.tenant_id::text, ts.status,
		       ts.current_period_start, ts.current_period_end,
		       COALESCE(
		           (SELECT o.value_json::text FROM tenant_entitlement_overrides o
		            WHERE o.tenant_id = ts.tenant_id AND o.feature_key = $2 AND o.active
		              AND (o.expires_at IS NULL OR o.expires_at > $1)
		            LIMIT 1),
		           (SELECT pe.value_json::text FROM plan_entitlements pe
		            WHERE pe.version_id = ts.version_id AND pe.feature_key = $2),
		           'null')
		FROM tenant_subscriptions ts
		JOIN tenants t ON t.tenant_id = ts.tenant_id AND t.billing_unit = $3
		WHERE ts.status = 'active'
		  AND ts.current_period_start <= $1 AND ts.current_period_end > $1
		ORDER BY ts.subscription_id
	`, now, credits.FeatureMonthlyGrant, credits.BillingUnitCredit)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (creditSubscription, error) {
		var sub creditSubscription

		scanErr := row.Scan(&sub.SubscriptionID, &sub.TenantID, &sub.Status, &sub.PeriodStart, &sub.PeriodEnd, &sub.MonthlyCredits)

		return sub, scanErr
	})
}
