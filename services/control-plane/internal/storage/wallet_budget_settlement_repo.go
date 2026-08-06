package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
)

// ApplyWalletBudgetSettlement mirrors committed budget-window spend into the ledger wallet balance.
func (s *WalletRepository) ApplyWalletBudgetSettlement(ctx context.Context, snapshot cache.BudgetWindowSnapshot) error {
	tenantID, windowID, err := parseBudgetWindowIdentity(snapshot.WindowKey)
	if err != nil {
		return err
	}

	committedIRR := roundedIRR(snapshot.CommittedIRR)
	if committedIRR <= 0 {
		return nil
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer rollbackTxBestEffort(ctx, tx)

	if _, err := tx.Exec(ctx, `
		INSERT INTO wallet_budget_window_settlements (
			tenant_id, window_id, settled_irr, last_observed_committed_irr, last_observed_at
		)
		VALUES ($1::uuid, $2, 0, 0, NOW())
		ON CONFLICT (tenant_id, window_id) DO NOTHING
	`, tenantID, windowID); err != nil {
		return err
	}

	var settledIRR int64
	if err := tx.QueryRow(ctx, `
		SELECT COALESCE(ROUND(settled_irr)::bigint, 0)
		FROM wallet_budget_window_settlements
		WHERE tenant_id::text = $1 AND window_id = $2
		FOR UPDATE
	`, tenantID, windowID).Scan(&settledIRR); err != nil {
		return err
	}

	deltaIRR := committedIRR - settledIRR
	now := time.Now().UTC()

	if deltaIRR > 0 {
		metadata := map[string]interface{}{
			"budget_window_key":       snapshot.WindowKey,
			"budget_window_id":        windowID,
			"budget_window_committed": committedIRR,
			"budget_window_settled":   settledIRR,
		}
		if _, err := appendWalletTransactionTx(ctx, tx, walletTransactionParams{
			TenantID:       tenantID,
			Type:           "consumption",
			AmountIRR:      deltaIRR,
			Status:         "successful",
			Description:    fmt.Sprintf("budget window settlement %s", windowID),
			Metadata:       metadata,
			SourceType:     "budget_window",
			SourceID:       snapshot.WindowKey,
			IdempotencyKey: fmt.Sprintf("budget_window_settlement:%s:%s:%d", tenantID, windowID, committedIRR),
			CompletedAt:    now,
		}); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(ctx, `
		UPDATE wallet_budget_window_settlements
		SET settled_irr = $3,
		    last_observed_committed_irr = $3,
		    last_observed_at = $4,
		    last_settled_at = CASE WHEN $5 THEN $4 ELSE last_settled_at END,
		    updated_at = NOW()
		WHERE tenant_id::text = $1 AND window_id = $2
	`, tenantID, windowID, committedIRR, now, deltaIRR > 0); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	s.invalidateWalletCache(ctx, tenantID)

	return nil
}

func parseBudgetWindowIdentity(windowKey string) (tenantID string, windowID string, err error) {
	parts := strings.Split(strings.TrimSpace(windowKey), ":")
	if len(parts) != 5 || parts[0] != "budget" || parts[1] != "tenant" || parts[3] != "window" {
		return "", "", fmt.Errorf("invalid budget window key %q", windowKey)
	}

	if strings.TrimSpace(parts[2]) == "" || strings.TrimSpace(parts[4]) == "" {
		return "", "", fmt.Errorf("invalid budget window key %q", windowKey)
	}

	return strings.TrimSpace(parts[2]), strings.TrimSpace(parts[4]), nil
}
