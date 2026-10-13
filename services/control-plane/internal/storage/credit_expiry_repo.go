package storage

import (
	"context"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const defaultCreditExpiryBatch = 200

// ExpireCreditLots writes an expire entry for the remaining credits of every lot
// whose expiry has passed. The expire:{lot_id} key makes repeated and concurrent
// runs safe.
func (s *CreditLedgerRepository) ExpireCreditLots(ctx context.Context, now time.Time, limit int) (credits.ExpiryResult, error) {
	if limit <= 0 {
		limit = defaultCreditExpiryBatch
	}

	rows, err := s.pool.Query(ctx, `
		SELECT lot_id::text, tenant_id::text, expires_at
		FROM credit_lot_balances
		WHERE remaining_mc > 0 AND expires_at IS NOT NULL AND expires_at <= $1
		ORDER BY expires_at, lot_id
		LIMIT $2
	`, now, limit)
	if err != nil {
		return credits.ExpiryResult{}, err
	}

	type dueLot struct {
		lotID, tenantID string
		expiresAt       time.Time
	}

	due, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (dueLot, error) {
		var lot dueLot

		scanErr := row.Scan(&lot.lotID, &lot.tenantID, &lot.expiresAt)

		return lot, scanErr
	})
	if err != nil {
		return credits.ExpiryResult{}, err
	}

	result := credits.ExpiryResult{}
	seen := map[string]bool{}

	for _, lot := range due {
		expiredMC, expireErr := s.expireLot(ctx, lot.lotID, lot.tenantID, lot.expiresAt.UTC())
		if expireErr != nil {
			return result, expireErr
		}

		if expiredMC == 0 {
			continue
		}

		result.ExpiredLots++
		result.ExpiredMC += expiredMC

		if !seen[lot.tenantID] {
			seen[lot.tenantID] = true
			result.TenantIDs = append(result.TenantIDs, lot.tenantID)
		}
	}

	return result, nil
}

func (s *CreditLedgerRepository) expireLot(ctx context.Context, lotID, tenantID string, expiresAt time.Time) (int64, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	entryID := uuid.NewString()
	key := credits.ExpireKey(lotID)

	claimed, err := claimCreditKeyTx(ctx, tx, key, entryID, tenantID, expiresAt)
	if err != nil || !claimed {
		return 0, err
	}

	if _, err := lockCreditTenantTx(ctx, tx, tenantID); err != nil {
		return 0, err
	}

	var remainingMC int64
	if err := tx.QueryRow(ctx, `
		SELECT remaining_mc FROM credit_lot_balances WHERE lot_id = $1::uuid FOR UPDATE
	`, lotID).Scan(&remainingMC); err != nil {
		return 0, err
	}

	if remainingMC <= 0 {
		// Spent between the scan and the lock: nothing to expire, so release the key.
		return 0, nil
	}

	if err := insertCreditEntryTx(ctx, tx, creditEntry{
		EntryID: entryID, TenantID: tenantID, Kind: credits.EntryExpire, AmountMC: -remainingMC,
		IdempotencyKey: key, OccurredAt: expiresAt, LotID: lotID,
	}); err != nil {
		return 0, err
	}

	if err := insertCreditAllocationsTx(ctx, tx, tenantID, expiresAt, []creditAllocation{
		{EntryID: entryID, LotID: lotID, AmountMC: -remainingMC},
	}); err != nil {
		return 0, err
	}

	if _, err := tx.Exec(ctx, `
		UPDATE credit_lot_balances SET remaining_mc = 0, updated_at = NOW() WHERE lot_id = $1::uuid
	`, lotID); err != nil {
		return 0, err
	}

	if err := applyCreditTenantDeltaTx(ctx, tx, tenantID, -remainingMC, 0, -remainingMC); err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	return remainingMC, nil
}

// EnsureCreditLedgerPartitions creates the monthly ledger partitions from the
// previous month through monthsAhead months after now and returns how many it created.
func (s *CreditLedgerRepository) EnsureCreditLedgerPartitions(ctx context.Context, now time.Time, monthsAhead int) (int, error) {
	var created int

	err := s.pool.QueryRow(ctx, `SELECT credit_ledger_ensure_partitions($1, $2)`, now, monthsAhead).Scan(&created)

	return created, err
}

// ReconcileCreditLedger compares the ledger, the allocations, and both
// projections in one consistent snapshot. It only reports; it never fixes.
func (s *CreditLedgerRepository) ReconcileCreditLedger(ctx context.Context) ([]credits.ReconciliationDiff, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	diffs, err := reconcileCreditTenantsTx(ctx, tx)
	if err != nil {
		return nil, err
	}

	more, err := reconcileCreditDetailTx(ctx, tx)
	if err != nil {
		return nil, err
	}

	diffs = append(diffs, more...)

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return diffs, nil
}

// reconcileCreditTenantsTx checks each tenant projection against the ledger
// sum, the lot projections, and the debt recorded in allocations.
func reconcileCreditTenantsTx(ctx context.Context, tx pgx.Tx) ([]credits.ReconciliationDiff, error) {
	rows, err := tx.Query(ctx, `
		WITH ledger AS (
			SELECT tenant_id, SUM(amount_mc)::bigint AS total FROM credit_ledger GROUP BY tenant_id
		), lots AS (
			SELECT tenant_id, SUM(remaining_mc)::bigint AS total FROM credit_lot_balances GROUP BY tenant_id
		), debt AS (
			SELECT tenant_id, (-SUM(amount_mc))::bigint AS total
			FROM credit_ledger_allocations WHERE lot_id IS NULL GROUP BY tenant_id
		), tenants AS (
			SELECT tenant_id FROM credit_tenant_balances
			UNION SELECT tenant_id FROM ledger
		)
		SELECT t.tenant_id::text,
		       COALESCE(l.total, 0), COALESCE(b.ledger_sum_mc, 0),
		       COALESCE(lt.total, 0), COALESCE(b.available_mc, 0),
		       COALESCE(d.total, 0), COALESCE(b.debt_mc, 0)
		FROM tenants t
		LEFT JOIN ledger l ON l.tenant_id = t.tenant_id
		LEFT JOIN lots lt ON lt.tenant_id = t.tenant_id
		LEFT JOIN debt d ON d.tenant_id = t.tenant_id
		LEFT JOIN credit_tenant_balances b ON b.tenant_id = t.tenant_id
		ORDER BY t.tenant_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var diffs []credits.ReconciliationDiff

	for rows.Next() {
		var (
			tenantID                      string
			ledgerSum, projectedSum       int64
			lotSum, projectedAvailable    int64
			allocationDebt, projectedDebt int64
		)

		if err := rows.Scan(&tenantID, &ledgerSum, &projectedSum, &lotSum, &projectedAvailable, &allocationDebt, &projectedDebt); err != nil {
			return nil, err
		}

		diffs = appendCreditDiff(diffs, tenantID, credits.CheckLedgerSum, ledgerSum, projectedSum)
		diffs = appendCreditDiff(diffs, tenantID, credits.CheckAvailable, lotSum, projectedAvailable)
		diffs = appendCreditDiff(diffs, tenantID, credits.CheckDebt, allocationDebt, projectedDebt)
		diffs = appendCreditDiff(diffs, tenantID, credits.CheckLedgerSum, ledgerSum, projectedAvailable-projectedDebt)
	}

	return diffs, rows.Err()
}

// reconcileCreditDetailTx checks every lot projection against its allocations,
// every entry against its allocation sum, and that no row fell into a DEFAULT
// partition.
func reconcileCreditDetailTx(ctx context.Context, tx pgx.Tx) ([]credits.ReconciliationDiff, error) {
	checks := []struct {
		check string
		query string
	}{
		{credits.CheckLotProjection, `
			SELECT b.tenant_id::text, COALESCE(SUM(a.total), 0)::bigint, SUM(b.remaining_mc)::bigint
			FROM credit_lot_balances b
			LEFT JOIN (
				SELECT lot_id, SUM(amount_mc) AS total FROM credit_ledger_allocations
				WHERE lot_id IS NOT NULL GROUP BY lot_id
			) a ON a.lot_id = b.lot_id
			WHERE COALESCE(a.total, 0) <> b.remaining_mc
			GROUP BY b.tenant_id`},
		{credits.CheckEntryAllocations, `
			SELECT l.tenant_id::text, SUM(l.amount_mc)::bigint, COALESCE(SUM(a.total), 0)::bigint
			FROM credit_ledger l
			LEFT JOIN (
				SELECT entry_id, SUM(amount_mc) AS total FROM credit_ledger_allocations GROUP BY entry_id
			) a ON a.entry_id = l.entry_id
			WHERE COALESCE(a.total, 0) <> l.amount_mc
			GROUP BY l.tenant_id`},
		{credits.CheckDefaultPartition, `
			SELECT '', 0::bigint, ((SELECT COUNT(1) FROM credit_ledger_default)
			                     + (SELECT COUNT(1) FROM credit_ledger_allocations_default))::bigint`},
	}

	var diffs []credits.ReconciliationDiff

	for _, check := range checks {
		rows, err := tx.Query(ctx, check.query)
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			var (
				tenantID         string
				expected, actual int64
			)

			if err := rows.Scan(&tenantID, &expected, &actual); err != nil {
				rows.Close()

				return nil, err
			}

			diffs = appendCreditDiff(diffs, tenantID, check.check, expected, actual)
		}

		rows.Close()

		if err := rows.Err(); err != nil {
			return nil, err
		}
	}

	return diffs, nil
}

func appendCreditDiff(diffs []credits.ReconciliationDiff, tenantID, check string, expected, actual int64) []credits.ReconciliationDiff {
	if expected == actual {
		return diffs
	}

	return append(diffs, credits.ReconciliationDiff{TenantID: tenantID, Check: check, ExpectedMC: expected, ActualMC: actual})
}
