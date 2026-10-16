package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

const (
	creditSettlementCursorID   = "spend_logs"
	creditSettlementLockKey    = "credit_consume_settlement"
	creditByokFeeFeatureKey    = "credits.byok_fee_percent"
	defaultCreditSettleBatch   = 500
	maxCreditSettleBatch       = 5000
	creditSettlementBillingKey = "external_billed"
)

// creditConsumption is one billable spend log of a credit tenant.
type creditConsumption struct {
	SpendLogID      int64
	TenantID        string
	RequestID       string
	Model           string
	UserID          string
	TeamID          string
	OrgUnitID       string
	CreditsMC       int64
	Product         string
	BillingMode     string
	RateCardVersion int
	CostUSD         string
	PricingRevision string
	TraceID         string
	OccurredAt      time.Time
}

// SettleCreditConsumption turns committed spend logs of credit tenants into
// consume entries. It reads past the cursor, skipping rows younger than
// settleDelay so a slow writer's lower id is not passed over, and advances the
// cursor only past rows it handled. A row that fails stops the pass at that row.
func (s *CreditLedgerRepository) SettleCreditConsumption(ctx context.Context, batchSize int, settleDelay time.Duration) (credits.SettlementResult, error) {
	batchSize = clampCreditBatch(batchSize)

	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return credits.SettlementResult{}, err
	}
	defer conn.Release()

	locked, err := tryCreditSessionLock(ctx, conn, creditSettlementLockKey)
	if err != nil || !locked {
		return credits.SettlementResult{}, err
	}
	defer s.unlockCreditSessionLock(conn, creditSettlementLockKey)

	result := credits.SettlementResult{Locked: true}

	cursor, err := readCreditSettlementCursor(ctx, conn)
	if err != nil {
		return result, err
	}

	result.LastSpendLogID = cursor

	windowEnd, scanned, err := creditSettlementWindow(ctx, conn, cursor, batchSize, settleDelay)
	if err != nil || scanned == 0 {
		return result, err
	}

	rows, err := listCreditConsumptions(ctx, conn, `s.id > $1 AND s.id <= $2`, cursor, windowEnd)
	if err != nil {
		return result, err
	}

	fees := newCreditFeeResolver(conn)
	for _, row := range rows {
		outcome, settleErr := s.settleOne(ctx, conn, row, fees)
		if settleErr != nil {
			result.LastSpendLogID = row.SpendLogID - 1

			if cursorErr := writeCreditSettlementCursor(ctx, conn, row.SpendLogID-1); cursorErr != nil {
				return result, errors.Join(settleErr, cursorErr)
			}

			return result, fmt.Errorf("settle spend log %d: %w", row.SpendLogID, settleErr)
		}

		addCreditSettleOutcome(&result, outcome)
	}

	if err := writeCreditSettlementCursor(ctx, conn, windowEnd); err != nil {
		return result, err
	}

	result.Scanned = scanned
	result.LastSpendLogID = windowEnd

	return result, nil
}

// SweepCreditConsumption re-settles billable spend logs created since `since`
// and already behind the cursor. Settlement is idempotent, so this only adds the
// rows the cursor passed over, and Settled reports how many it recovered.
func (s *CreditLedgerRepository) SweepCreditConsumption(ctx context.Context, since time.Time) (credits.SettlementResult, error) {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return credits.SettlementResult{}, err
	}
	defer conn.Release()

	locked, err := tryCreditSessionLock(ctx, conn, creditSettlementLockKey)
	if err != nil || !locked {
		return credits.SettlementResult{}, err
	}
	defer s.unlockCreditSessionLock(conn, creditSettlementLockKey)

	result := credits.SettlementResult{Locked: true}

	cursor, err := readCreditSettlementCursor(ctx, conn)
	if err != nil {
		return result, err
	}

	result.LastSpendLogID = cursor

	rows, err := listCreditConsumptions(ctx, conn, `s.id <= $1 AND s.created_at >= $2`, cursor, since)
	if err != nil {
		return result, err
	}

	fees := newCreditFeeResolver(conn)
	for _, row := range rows {
		outcome, settleErr := s.settleOne(ctx, conn, row, fees)
		if settleErr != nil {
			return result, fmt.Errorf("sweep spend log %d: %w", row.SpendLogID, settleErr)
		}

		result.Scanned++
		addCreditSettleOutcome(&result, outcome)
	}

	return result, nil
}

type creditSettleOutcome struct {
	settled bool
	debtMC  int64
}

func addCreditSettleOutcome(r *credits.SettlementResult, outcome creditSettleOutcome) {
	if !outcome.settled {
		r.KeyConflicts++

		return
	}

	r.Settled++
	r.DebtMC += outcome.debtMC
}

func (s *CreditLedgerRepository) settleOne(ctx context.Context, conn *pgxpool.Conn, row creditConsumption, fees *creditFeeResolver) (creditSettleOutcome, error) {
	amountMC := row.CreditsMC

	var equivalentMC *int64

	if row.BillingMode == creditSettlementBillingKey {
		feePercent, err := fees.percent(ctx, row.TenantID)
		if err != nil {
			return creditSettleOutcome{}, err
		}

		equivalent := row.CreditsMC
		equivalentMC = &equivalent
		amountMC = credits.BYOKFeeMC(row.CreditsMC, feePercent)
	}

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return creditSettleOutcome{}, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	entryID := uuid.NewString()
	key := credits.ConsumeKey(row.TenantID, row.RequestID, row.Product, row.SpendLogID)

	claimed, err := claimCreditKeyTx(ctx, tx, key, entryID, row.TenantID, row.OccurredAt)
	if err != nil || !claimed {
		return creditSettleOutcome{}, err
	}

	if _, lockErr := lockCreditTenantTx(ctx, tx, row.TenantID); lockErr != nil {
		return creditSettleOutcome{}, lockErr
	}

	spendLogID := row.SpendLogID
	if entryErr := insertCreditEntryTx(ctx, tx, creditEntry{
		EntryID: entryID, TenantID: row.TenantID, Kind: credits.EntryConsume, AmountMC: -amountMC,
		IdempotencyKey: key, OccurredAt: row.OccurredAt, RequestID: row.RequestID, TraceID: row.TraceID,
		Model: row.Model, UserID: row.UserID, TeamID: row.TeamID, OrgUnitID: row.OrgUnitID,
		Product: row.Product, BillingMode: row.BillingMode, PricingRevision: row.PricingRevision,
		RateCardVersion: row.RateCardVersion, InternalCostUSD: row.CostUSD,
		CreditEquivalentMC: equivalentMC, SpendLogID: &spendLogID,
	}); entryErr != nil {
		return creditSettleOutcome{}, entryErr
	}

	allocations, debtMC, err := debitCreditLotsTx(ctx, tx, row.TenantID, entryID, amountMC, row.OccurredAt)
	if err != nil {
		return creditSettleOutcome{}, err
	}

	if err := insertCreditAllocationsTx(ctx, tx, row.TenantID, row.OccurredAt, allocations); err != nil {
		return creditSettleOutcome{}, err
	}

	if err := applyCreditTenantDeltaTx(ctx, tx, row.TenantID, -(amountMC - debtMC), debtMC, -amountMC); err != nil {
		return creditSettleOutcome{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return creditSettleOutcome{}, err
	}

	return creditSettleOutcome{settled: true, debtMC: debtMC}, nil
}

func clampCreditBatch(batchSize int) int {
	if batchSize <= 0 {
		return defaultCreditSettleBatch
	}

	return min(batchSize, maxCreditSettleBatch)
}

// creditSettlementWindow returns the last spend log id of the next window and
// the window size. Rows younger than settleDelay are left for a later pass.
func creditSettlementWindow(ctx context.Context, conn *pgxpool.Conn, cursor int64, batchSize int, settleDelay time.Duration) (int64, int, error) {
	var (
		windowEnd *int64
		scanned   int
	)

	err := conn.QueryRow(ctx, `
		SELECT MAX(id), COUNT(1)::int
		FROM (
			SELECT id FROM spend_logs
			WHERE id > $1 AND created_at < NOW() - make_interval(secs => $3)
			ORDER BY id
			LIMIT $2
		) scanned
	`, cursor, batchSize, settleDelay.Seconds()).Scan(&windowEnd, &scanned)
	if err != nil || windowEnd == nil {
		return cursor, 0, err
	}

	return *windowEnd, scanned, nil
}

func listCreditConsumptions(ctx context.Context, conn *pgxpool.Conn, filter string, args ...interface{}) ([]creditConsumption, error) {
	rows, err := conn.Query(ctx, `
		SELECT s.id, s.tenant_id::text, COALESCE(s.request_id, ''), COALESCE(s.model, ''),
		       COALESCE(s.user_id::text, ''), COALESCE(s.team_id::text, ''), COALESCE(s.org_unit_id::text, ''),
		       s.credits_mc, COALESCE(NULLIF(s.product, ''), '`+platform.CreditProductGateway+`'),
		       COALESCE(s.billing_mode, ''), COALESCE(s.rate_card_version, 0),
		       COALESCE(s.cost_usd::text, ''), COALESCE(s.pricing_snapshot_json->>'revision', ''),
		       COALESCE(s.request_metadata->>'trace_id', ''), COALESCE(s.created_at, NOW())
		FROM spend_logs s
		JOIN tenants t ON t.tenant_id = s.tenant_id AND t.billing_unit = '`+credits.BillingUnitCredit+`'
		WHERE `+filter+` AND s.credits_mc > 0
		ORDER BY s.id
	`, args...)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (creditConsumption, error) {
		var item creditConsumption

		scanErr := row.Scan(&item.SpendLogID, &item.TenantID, &item.RequestID, &item.Model,
			&item.UserID, &item.TeamID, &item.OrgUnitID, &item.CreditsMC, &item.Product,
			&item.BillingMode, &item.RateCardVersion, &item.CostUSD, &item.PricingRevision,
			&item.TraceID, &item.OccurredAt)
		item.OccurredAt = item.OccurredAt.UTC()

		return item, scanErr
	})
}

func readCreditSettlementCursor(ctx context.Context, conn *pgxpool.Conn) (int64, error) {
	var cursor int64

	err := conn.QueryRow(ctx, `SELECT last_spend_log_id FROM credit_settlement_cursors WHERE cursor_id = $1`, creditSettlementCursorID).Scan(&cursor)
	if errors.Is(err, pgx.ErrNoRows) {
		return startCreditSettlementCursor(ctx, conn)
	}

	return cursor, err
}

// startCreditSettlementCursor starts settlement at the newest spend log, so
// usage from before the ledger existed is never charged in credits.
func startCreditSettlementCursor(ctx context.Context, conn *pgxpool.Conn) (int64, error) {
	var cursor int64

	err := conn.QueryRow(ctx, `
		INSERT INTO credit_settlement_cursors (cursor_id, last_spend_log_id)
		VALUES ($1, (SELECT COALESCE(MAX(id), 0) FROM spend_logs))
		ON CONFLICT (cursor_id) DO UPDATE SET cursor_id = EXCLUDED.cursor_id
		RETURNING last_spend_log_id
	`, creditSettlementCursorID).Scan(&cursor)

	return cursor, err
}

func writeCreditSettlementCursor(ctx context.Context, conn *pgxpool.Conn, cursor int64) error {
	_, err := conn.Exec(ctx, `
		UPDATE credit_settlement_cursors
		SET last_spend_log_id = GREATEST(last_spend_log_id, $2), updated_at = NOW()
		WHERE cursor_id = $1
	`, creditSettlementCursorID, cursor)

	return err
}

func tryCreditSessionLock(ctx context.Context, conn *pgxpool.Conn, name string) (bool, error) {
	var locked bool

	err := conn.QueryRow(ctx, `SELECT pg_try_advisory_lock(hashtext($1))`, name).Scan(&locked)

	return locked, err
}

func (s *CreditLedgerRepository) unlockCreditSessionLock(conn *pgxpool.Conn, name string) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if _, err := conn.Exec(ctx, `SELECT pg_advisory_unlock(hashtext($1))`, name); err != nil {
		s.logger.Warn("credit ledger lock release failed; closing the connection", zap.String("lock", name), zap.Error(err))

		// Closing the session releases its advisory locks.
		if closeErr := conn.Conn().Close(ctx); closeErr != nil {
			s.logger.Warn("credit ledger connection close failed", zap.String("lock", name), zap.Error(closeErr))
		}
	}
}

// creditFeeResolver caches each tenant's BYOK fee percent for one pass.
type creditFeeResolver struct {
	conn  *pgxpool.Conn
	cache map[string]decimal.Decimal
}

func newCreditFeeResolver(conn *pgxpool.Conn) *creditFeeResolver {
	return &creditFeeResolver{conn: conn, cache: map[string]decimal.Decimal{}}
}

// percent resolves credits.byok_fee_percent: an active tenant override first,
// then the current plan's entitlement, then the global default.
func (r *creditFeeResolver) percent(ctx context.Context, tenantID string) (decimal.Decimal, error) {
	if cached, ok := r.cache[tenantID]; ok {
		return cached, nil
	}

	var raw *string

	err := r.conn.QueryRow(ctx, `
		SELECT COALESCE(
			(SELECT o.value_json::text FROM tenant_entitlement_overrides o
			 WHERE o.tenant_id = $1::uuid AND o.feature_key = $2 AND o.active
			   AND (o.expires_at IS NULL OR o.expires_at > NOW())
			   AND jsonb_typeof(o.value_json) = 'number'
			 LIMIT 1),
			(SELECT pe.value_json::text FROM tenant_subscriptions ts
			 JOIN plan_entitlements pe ON pe.version_id = ts.version_id AND pe.feature_key = $2
			 WHERE ts.tenant_id = $1::uuid AND ts.status IN ('trialing', 'active', 'past_due')
			   AND jsonb_typeof(pe.value_json) = 'number'
			 ORDER BY ts.updated_at DESC LIMIT 1)
		)
	`, tenantID, creditByokFeeFeatureKey).Scan(&raw)
	if err != nil {
		return decimal.Zero, err
	}

	fee := credits.DefaultBYOKFeePercent

	if raw != nil {
		parsed, parseErr := decimal.NewFromString(strings.TrimSpace(*raw))
		if parseErr == nil && !parsed.IsNegative() && parsed.LessThanOrEqual(decimal.NewFromInt(100)) {
			fee = parsed
		}
	}

	r.cache[tenantID] = fee

	return fee, nil
}
