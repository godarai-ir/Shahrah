package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type budgetWindowStateCache interface {
	cache.BudgetWindowInspector
	SeedBudgetWindow(ctx context.Context, seed cache.BudgetWindowSeed) (*cache.BudgetWindowResult, error)
}

type budgetWindowSettlementState struct {
	SettledIRR    int64
	LastSettledAt *time.Time
}

func (s *Store) decorateWalletWithLiveState(ctx context.Context, wallet *platform.Wallet) (*platform.Wallet, error) {
	if wallet == nil {
		return nil, nil
	}

	now := time.Now().UTC()
	wallet.LedgerBalanceIRR = wallet.BalanceIRR
	wallet.AvailableBalanceIRR = wallet.BalanceIRR
	wallet.SyncStatus = "ledger_only"
	wallet.AsOf = &now

	liveCache, ok := s.entitlementCache.(cache.BudgetWindowInspector)
	if !ok || liveCache == nil {
		return wallet, nil
	}

	windowID, _ := currentBudgetWindowForTime(now)
	windowKey := currentBudgetWindowKey(wallet.TenantID, windowID)

	snapshot, err := liveCache.GetBudgetWindow(ctx, windowKey)
	if err != nil {
		return nil, fmt.Errorf("get live budget window: %w", err)
	}

	if snapshot == nil {
		return wallet, nil
	}

	settlement, err := s.getBudgetWindowSettlement(ctx, wallet.TenantID, windowID)
	if err != nil {
		return nil, fmt.Errorf("get budget window settlement: %w", err)
	}

	reservedIRR := roundedIRR(snapshot.ReservedIRR)
	committedIRR := roundedIRR(snapshot.CommittedIRR)

	committedUnsettledIRR := committedIRR - settlement.SettledIRR
	if committedUnsettledIRR < 0 {
		committedUnsettledIRR = 0
	}

	availableIRR := roundedIRR(snapshot.RemainingIRR)
	if availableIRR < 0 {
		availableIRR = 0
	}

	wallet.WindowID = windowID
	wallet.ReservedIRR = reservedIRR
	wallet.CommittedUnsettledIRR = committedUnsettledIRR
	wallet.AvailableBalanceIRR = availableIRR

	switch {
	case committedUnsettledIRR > 0 || reservedIRR > 0:
		wallet.SyncStatus = "pending_settlement"

		switch {
		case settlement.LastSettledAt != nil:
			wallet.SyncLagSeconds = int64(maxDuration(0, time.Since(settlement.LastSettledAt.UTC())).Seconds())
		case !snapshot.LastUpdatedAt.IsZero():
			wallet.SyncLagSeconds = int64(maxDuration(0, time.Since(snapshot.LastUpdatedAt)).Seconds())
		}
	default:
		wallet.SyncStatus = "synchronized"
		wallet.SyncLagSeconds = 0
	}

	return wallet, nil
}

func (s *Store) getBudgetWindowSettlement(ctx context.Context, tenantID string, windowID string) (*budgetWindowSettlementState, error) {
	if s == nil || s.pool == nil {
		return &budgetWindowSettlementState{}, nil
	}

	var (
		settled      int64
		lastSyncedAt *time.Time
	)

	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(ROUND(settled_irr)::bigint, 0), last_settled_at
		FROM wallet_budget_window_settlements
		WHERE tenant_id::text = $1 AND window_id = $2
	`, strings.TrimSpace(tenantID), strings.TrimSpace(windowID)).Scan(&settled, &lastSyncedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return &budgetWindowSettlementState{}, nil
	}

	if err != nil {
		return nil, err
	}

	return &budgetWindowSettlementState{
		SettledIRR:    settled,
		LastSettledAt: lastSyncedAt,
	}, nil
}

func (s *Store) syncCurrentBudgetWindowAllowance(ctx context.Context, tenantID string) {
	liveCache, ok := s.entitlementCache.(budgetWindowStateCache)
	if !ok || liveCache == nil {
		return
	}

	now := time.Now().UTC()
	windowID, windowEnd := currentBudgetWindowForTime(now)
	windowKey := currentBudgetWindowKey(tenantID, windowID)

	snapshot, err := liveCache.GetBudgetWindow(ctx, windowKey)
	if err != nil || snapshot == nil {
		return
	}

	settlement, err := s.getBudgetWindowSettlement(ctx, tenantID, windowID)
	if err != nil {
		return
	}

	wallet, err := s.getWalletLedger(ctx, tenantID)
	if err != nil || wallet == nil {
		return
	}

	ttlGrace := secondsEnvOrDefault("BILLING_BUDGET_WINDOW_GRACE_SECONDS", 86400*time.Second)

	ttl := time.Until(windowEnd.Add(ttlGrace))
	if ttl <= 0 {
		ttl = ttlGrace
	}

	allowance := float64(wallet.BalanceIRR + settlement.SettledIRR)

	_, err = liveCache.SeedBudgetWindow(ctx, cache.BudgetWindowSeed{
		WindowKey:    windowKey,
		AllowanceIRR: allowance,
		ExpiresAt:    windowEnd,
		TTL:          ttl,
		Version:      windowEnd.Unix(),
	})
	if err != nil && s.logger != nil {
		s.logger.Warn("failed to seed current budget window", zap.Error(err), zap.String("tenant_id", tenantID))
	}
}

func currentBudgetWindowForTime(now time.Time) (string, time.Time) {
	utc := now.UTC()
	start := time.Date(utc.Year(), utc.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)

	return start.Format("200601"), end
}

func currentBudgetWindowKey(tenantID string, windowID string) string {
	return fmt.Sprintf("budget:tenant:%s:window:%s", strings.TrimSpace(tenantID), strings.TrimSpace(windowID))
}

func maxDuration(left time.Duration, right time.Duration) time.Duration {
	if right > left {
		return right
	}

	return left
}
