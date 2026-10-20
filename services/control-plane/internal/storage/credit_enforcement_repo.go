package storage

import (
	"context"
	"strings"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

// CreditPoolAllowanceMC returns the allowance of a credit tenant's pool window:
// what the tenant can still spend, plus consumption the ledger already settled
// in that window (the window counts it as committed too).
func (s *CreditLedgerRepository) CreditPoolAllowanceMC(ctx context.Context, tenantID, windowID string) (int64, error) {
	pool, err := s.creditPoolAllowance(ctx, s.pool, tenantID, windowID)
	if err != nil {
		return 0, err
	}

	return pool.AllowanceMC(), nil
}

func (s *walletState) creditPoolAllowance(ctx context.Context, q rowQuerier, tenantID, windowID string) (credits.PoolAllowance, error) {
	pool := credits.PoolAllowance{OverageCeilingMC: s.creditOverageCeilingMC}

	var overageEnabled bool

	if err := q.QueryRow(ctx, `
		SELECT COALESCE(b.available_mc, 0) - COALESCE((
		           SELECT SUM(lb.remaining_mc) FROM credit_lot_balances lb
		           WHERE lb.tenant_id = t.tenant_id AND lb.owner_user_id IS NOT NULL AND lb.remaining_mc > 0
		       ), 0),
		       COALESCE(b.debt_mc, 0), COALESCE(w.settled_mc, 0),
		       COALESCE(cs.overage_enabled, false), cs.overage_cap_mc
		FROM tenants t
		LEFT JOIN credit_tenant_balances b ON b.tenant_id = t.tenant_id
		LEFT JOIN credit_budget_window_settlements w ON w.tenant_id = t.tenant_id AND w.window_id = $2
		LEFT JOIN credit_tenant_settings cs ON cs.tenant_id = t.tenant_id
		WHERE t.tenant_id = $1::uuid
	`, strings.TrimSpace(tenantID), strings.TrimSpace(windowID)).Scan(
		&pool.AvailableMC, &pool.DebtMC, &pool.SettledMC, &overageEnabled, &pool.OverageCapMC,
	); err != nil {
		return credits.PoolAllowance{}, err
	}

	if overageEnabled {
		raw, err := resolveTenantFeatureJSON(ctx, q, tenantID, credits.FeatureOverageAllowed)
		if err != nil {
			return credits.PoolAllowance{}, err
		}

		pool.OverageEnabled = strings.TrimSpace(raw) == "true"
	}

	return pool, nil
}

// addCreditWindowSettledTx records consumption the ledger settled in the pool
// window of its occurrence.
func addCreditWindowSettledTx(ctx context.Context, tx pgx.Tx, tenantID string, occurredAt time.Time, amountMC int64) error {
	if amountMC <= 0 {
		return nil
	}

	windowID, _ := credits.PoolWindow(occurredAt)

	_, err := tx.Exec(ctx, `
		INSERT INTO credit_budget_window_settlements (tenant_id, window_id, settled_mc)
		VALUES ($1::uuid, $2, $3)
		ON CONFLICT (tenant_id, window_id) DO UPDATE
		SET settled_mc = credit_budget_window_settlements.settled_mc + EXCLUDED.settled_mc, updated_at = NOW()
	`, tenantID, windowID, amountMC)

	return err
}

// addUserWindowSettledTx records what the ledger settled for one member in the
// calendar-month window of its occurrence: from their personal lots, and from
// the pool on their behalf.
func addUserWindowSettledTx(ctx context.Context, tx pgx.Tx, tenantID, userID string, occurredAt time.Time, personalMC, poolMC int64) error {
	personalMC, poolMC = max(personalMC, 0), max(poolMC, 0)
	if personalMC == 0 && poolMC == 0 {
		return nil
	}

	windowID, _ := credits.PoolWindow(occurredAt)

	_, err := tx.Exec(ctx, `
		INSERT INTO credit_user_window_settlements (tenant_id, user_id, window_id, settled_mc, pool_settled_mc)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5)
		ON CONFLICT (tenant_id, user_id, window_id) DO UPDATE
		SET settled_mc = credit_user_window_settlements.settled_mc + EXCLUDED.settled_mc,
		    pool_settled_mc = credit_user_window_settlements.pool_settled_mc + EXCLUDED.pool_settled_mc,
		    updated_at = NOW()
	`, tenantID, userID, windowID, personalMC, poolMC)

	return err
}

// syncCreditPoolWindows re-seeds the current pool window of each tenant after
// its balance changed. Windows that do not exist yet are seeded on first use.
func (s *walletState) syncCreditPoolWindows(ctx context.Context, tenantIDs map[string]struct{}) {
	for tenantID := range tenantIDs {
		s.syncCreditPoolWindow(ctx, tenantID)
	}
}

func (s *walletState) syncCreditPoolWindow(ctx context.Context, tenantID string) {
	liveCache, ok := s.entitlementCache.(budgetWindowStateCache)
	if !ok || liveCache == nil || s.pool == nil {
		return
	}

	now := time.Now().UTC()
	windowID, windowEnd := credits.PoolWindow(now)
	windowKey := credits.PoolWindowKey(tenantID, windowID)

	snapshot, err := liveCache.GetBudgetWindow(ctx, windowKey)
	if err != nil || snapshot == nil {
		return
	}

	pool, err := s.creditPoolAllowance(ctx, s.pool, tenantID, windowID)
	if err != nil {
		s.logCreditWindowSyncFailure(err, tenantID)

		return
	}

	ttl := time.Until(windowEnd.Add(s.budgetWindowTTLGrace))
	if ttl <= 0 {
		ttl = s.budgetWindowTTLGrace
	}

	if _, err := liveCache.SeedBudgetWindow(ctx, cache.BudgetWindowSeed{
		WindowKey:    windowKey,
		AllowanceIRR: float64(pool.AllowanceMC()),
		ExpiresAt:    windowEnd,
		TTL:          ttl,
		Version:      windowEnd.Unix(),
	}); err != nil {
		s.logCreditWindowSyncFailure(err, tenantID)
	}
}

func (s *walletState) logCreditWindowSyncFailure(err error, tenantID string) {
	if s.logger != nil {
		s.logger.Warn("failed to re-seed credit pool window", zap.Error(err), zap.String("tenant_id", tenantID))
	}
}

// invalidateCreditTenant refreshes everything that caches a credit tenant's
// state after its balance or billing unit changed.
func (s *walletState) invalidateCreditTenant(ctx context.Context, tenantID, reason string) {
	s.syncCreditPoolWindow(ctx, tenantID)

	if s.runtimeProfileInvalidator != nil && reason != "" {
		s.runtimeProfileInvalidator.Invalidate(ctx, tenantID, reason)
	}
}
