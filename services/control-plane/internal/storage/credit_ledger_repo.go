package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	// ErrInvalidCreditGrant reports a grant request that cannot be written.
	ErrInvalidCreditGrant = errors.New("invalid credit grant")
)

// CreditGrant describes credits added to a tenant as a new lot.
type CreditGrant struct {
	TenantID       string
	Kind           credits.EntryKind
	Source         credits.LotSource
	AmountMC       int64
	ExpiresAt      *time.Time
	IdempotencyKey string
	OccurredAt     time.Time
	// PurchasePriceIRR and PricePerCreditIRR are set for paid lots only.
	PurchasePriceIRR  *int64
	PricePerCreditIRR string
	SourceRef         map[string]interface{}
	Reason            string
	Actor             string
}

// CreditGrantResult reports the outcome of a grant.
type CreditGrantResult struct {
	EntryID      string
	LotID        string
	Duplicate    bool
	RepaidDebtMC int64
}

// CreditBalance is a tenant's credit balance projection.
type CreditBalance struct {
	TenantID    string
	AvailableMC int64
	DebtMC      int64
	LedgerSumMC int64
	UpdatedAt   time.Time
}

// creditEntry is one credit_ledger row.
type creditEntry struct {
	EntryID            string
	TenantID           string
	Kind               credits.EntryKind
	AmountMC           int64
	IdempotencyKey     string
	OccurredAt         time.Time
	LotID              string
	RequestID          string
	TraceID            string
	Model              string
	UserID             string
	TeamID             string
	Product            string
	BillingMode        string
	PricingRevision    string
	RateCardVersion    int
	InternalCostUSD    string
	CreditEquivalentMC *int64
	SpendLogID         *int64
	Reason             string
	Actor              string
}

// GetCreditBalance returns the tenant's balance projection; a tenant with no
// ledger activity has a zero balance.
func (s *CreditLedgerRepository) GetCreditBalance(ctx context.Context, tenantID string) (CreditBalance, error) {
	balance := CreditBalance{TenantID: tenantID}

	err := s.pool.QueryRow(ctx, `
		SELECT available_mc, debt_mc, ledger_sum_mc, updated_at
		FROM credit_tenant_balances
		WHERE tenant_id = $1::uuid
	`, tenantID).Scan(&balance.AvailableMC, &balance.DebtMC, &balance.LedgerSumMC, &balance.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return balance, nil
	}

	return balance, err
}

// GrantCredits adds a lot, pays outstanding debt from it first, and records the
// grant in the ledger. Repeating the same idempotency key grants nothing.
func (s *CreditLedgerRepository) GrantCredits(ctx context.Context, grant CreditGrant) (CreditGrantResult, error) {
	if err := validateCreditGrant(grant); err != nil {
		return CreditGrantResult{}, err
	}

	if grant.OccurredAt.IsZero() {
		grant.OccurredAt = time.Now().UTC()
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return CreditGrantResult{}, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	result, err := grantCreditsTx(ctx, tx, grant)
	if err != nil {
		return CreditGrantResult{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return CreditGrantResult{}, err
	}

	s.invalidateCreditTenant(ctx, grant.TenantID, "")

	return result, nil
}

func validateCreditGrant(grant CreditGrant) error {
	switch {
	case strings.TrimSpace(grant.TenantID) == "":
		return fmt.Errorf("%w: tenant is required", ErrInvalidCreditGrant)
	case grant.AmountMC <= 0:
		return fmt.Errorf("%w: amount must be positive", ErrInvalidCreditGrant)
	case strings.TrimSpace(grant.IdempotencyKey) == "":
		return fmt.Errorf("%w: idempotency key is required", ErrInvalidCreditGrant)
	case !credits.ValidLotSource(grant.Source):
		return fmt.Errorf("%w: unknown lot source %q", ErrInvalidCreditGrant, grant.Source)
	}

	switch grant.Kind {
	case credits.EntryPurchase, credits.EntryPlanGrant, credits.EntryPromoGrant, credits.EntryRefund:
		return nil
	case credits.EntryAdjust, credits.EntryConvert:
		if strings.TrimSpace(grant.Reason) == "" {
			return fmt.Errorf("%w: adjustments need a reason", ErrInvalidCreditGrant)
		}

		return nil
	default:
		return fmt.Errorf("%w: kind %q cannot grant credits", ErrInvalidCreditGrant, grant.Kind)
	}
}

// grantCreditsTx writes a grant inside tx. The caller commits.
func grantCreditsTx(ctx context.Context, tx pgx.Tx, grant CreditGrant) (CreditGrantResult, error) {
	entryID := uuid.NewString()

	claimed, err := claimCreditKeyTx(ctx, tx, grant.IdempotencyKey, entryID, grant.TenantID, grant.OccurredAt)
	if err != nil {
		return CreditGrantResult{}, err
	}

	if !claimed {
		return existingCreditGrantTx(ctx, tx, grant.IdempotencyKey)
	}

	state, err := lockCreditTenantTx(ctx, tx, grant.TenantID)
	if err != nil {
		return CreditGrantResult{}, err
	}

	lotID := uuid.NewString()
	if entryErr := insertCreditEntryTx(ctx, tx, creditEntry{
		EntryID: entryID, TenantID: grant.TenantID, Kind: grant.Kind, AmountMC: grant.AmountMC,
		IdempotencyKey: grant.IdempotencyKey, OccurredAt: grant.OccurredAt, LotID: lotID,
		Reason: grant.Reason, Actor: grant.Actor,
	}); entryErr != nil {
		return CreditGrantResult{}, entryErr
	}

	if lotErr := insertCreditLotTx(ctx, tx, lotID, entryID, grant); lotErr != nil {
		return CreditGrantResult{}, lotErr
	}

	repaid, repayAllocations, err := repayCreditDebtTx(ctx, tx, grant.TenantID, lotID, grant.AmountMC, state.DebtMC)
	if err != nil {
		return CreditGrantResult{}, err
	}

	allocations := make([]creditAllocation, 0, 1+len(repayAllocations))
	allocations = append(allocations, creditAllocation{EntryID: entryID, LotID: lotID, AmountMC: grant.AmountMC})
	allocations = append(allocations, repayAllocations...)

	if err := insertCreditAllocationsTx(ctx, tx, grant.TenantID, grant.OccurredAt, allocations); err != nil {
		return CreditGrantResult{}, err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO credit_lot_balances (lot_id, tenant_id, remaining_mc, expires_at, source_rank, lot_created_at)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6)
	`, lotID, grant.TenantID, grant.AmountMC-repaid, grant.ExpiresAt, credits.SourceRank(grant.Source), grant.OccurredAt); err != nil {
		return CreditGrantResult{}, err
	}

	if err := applyCreditTenantDeltaTx(ctx, tx, grant.TenantID, grant.AmountMC-repaid, -repaid, grant.AmountMC); err != nil {
		return CreditGrantResult{}, err
	}

	return CreditGrantResult{EntryID: entryID, LotID: lotID, RepaidDebtMC: repaid}, nil
}

func insertCreditLotTx(ctx context.Context, tx pgx.Tx, lotID, entryID string, grant CreditGrant) error {
	sourceRef := grant.SourceRef
	if sourceRef == nil {
		sourceRef = map[string]interface{}{}
	}

	_, err := tx.Exec(ctx, `
		INSERT INTO credit_lots (
			lot_id, tenant_id, source, amount_mc, expires_at, purchase_price_irr,
			price_per_credit_irr, source_ref, grant_entry_id, created_at
		)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, NULLIF($7, '')::numeric, $8, $9::uuid, $10)
	`, lotID, grant.TenantID, string(grant.Source), grant.AmountMC, grant.ExpiresAt, grant.PurchasePriceIRR,
		grant.PricePerCreditIRR, mustJSON(sourceRef), entryID, grant.OccurredAt)

	return err
}

func existingCreditGrantTx(ctx context.Context, tx pgx.Tx, key string) (CreditGrantResult, error) {
	result := CreditGrantResult{Duplicate: true}

	err := tx.QueryRow(ctx, `
		SELECT k.entry_id::text, COALESCE(l.lot_id::text, '')
		FROM credit_ledger_keys k
		LEFT JOIN credit_ledger l ON l.entry_id = k.entry_id AND l.occurred_at = k.occurred_at
		WHERE k.idempotency_key = $1
	`, key).Scan(&result.EntryID, &result.LotID)

	return result, err
}

// repayCreditDebtTx moves outstanding debt onto a new lot, oldest debt first.
// It returns the repaid amount and the allocations that record it.
func repayCreditDebtTx(ctx context.Context, tx pgx.Tx, tenantID, lotID string, grantMC, debtMC int64) (int64, []creditAllocation, error) {
	if debtMC <= 0 {
		return 0, nil, nil
	}

	rows, err := tx.Query(ctx, `
		SELECT entry_id::text, (-SUM(amount_mc))::bigint
		FROM credit_ledger_allocations
		WHERE tenant_id = $1::uuid AND lot_id IS NULL
		GROUP BY entry_id
		HAVING SUM(amount_mc) < 0
		ORDER BY MIN(occurred_at), entry_id
	`, tenantID)
	if err != nil {
		return 0, nil, err
	}

	debts, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (credits.DebtEntry, error) {
		var debt credits.DebtEntry

		scanErr := row.Scan(&debt.EntryID, &debt.OutstandingMC)

		return debt, scanErr
	})
	if err != nil {
		return 0, nil, err
	}

	repayments, left := credits.RepayDebt(debts, grantMC)

	allocations := make([]creditAllocation, 0, len(repayments)*2)
	for _, repayment := range repayments {
		allocations = append(allocations,
			creditAllocation{EntryID: repayment.EntryID, AmountMC: repayment.AmountMC},
			creditAllocation{EntryID: repayment.EntryID, LotID: lotID, AmountMC: -repayment.AmountMC},
		)
	}

	return grantMC - left, allocations, nil
}

// creditTenantState is the locked balance row of one tenant.
type creditTenantState struct {
	AvailableMC int64
	DebtMC      int64
	LedgerSumMC int64
}

// lockCreditTenantTx serializes every ledger write of one tenant. Writers take
// the idempotency key first, then this lock, then lot rows, so they never deadlock.
func lockCreditTenantTx(ctx context.Context, tx pgx.Tx, tenantID string) (creditTenantState, error) {
	if _, err := tx.Exec(ctx, `
		INSERT INTO credit_tenant_balances (tenant_id) VALUES ($1::uuid)
		ON CONFLICT (tenant_id) DO NOTHING
	`, tenantID); err != nil {
		return creditTenantState{}, err
	}

	var state creditTenantState

	err := tx.QueryRow(ctx, `
		SELECT available_mc, debt_mc, ledger_sum_mc
		FROM credit_tenant_balances
		WHERE tenant_id = $1::uuid
		FOR UPDATE
	`, tenantID).Scan(&state.AvailableMC, &state.DebtMC, &state.LedgerSumMC)

	return state, err
}

// claimCreditKeyTx reserves an idempotency key and reports whether this call won it.
func claimCreditKeyTx(ctx context.Context, tx pgx.Tx, key, entryID, tenantID string, occurredAt time.Time) (bool, error) {
	tag, err := tx.Exec(ctx, `
		INSERT INTO credit_ledger_keys (idempotency_key, entry_id, tenant_id, occurred_at)
		VALUES ($1, $2::uuid, $3::uuid, $4)
		ON CONFLICT (idempotency_key) DO NOTHING
	`, key, entryID, tenantID, occurredAt)
	if err != nil {
		return false, err
	}

	return tag.RowsAffected() == 1, nil
}

func insertCreditEntryTx(ctx context.Context, tx pgx.Tx, entry creditEntry) error {
	var rateCardVersion interface{}
	if entry.RateCardVersion > 0 {
		rateCardVersion = entry.RateCardVersion
	}

	_, err := tx.Exec(ctx, `
		INSERT INTO credit_ledger (
			entry_id, tenant_id, kind, amount_mc, idempotency_key, occurred_at, lot_id,
			request_id, trace_id, model, user_id, team_id, product, billing_mode,
			pricing_revision, rate_card_version, internal_cost_usd, credit_equivalent_mc, spend_log_id,
			reason, actor
		)
		VALUES (
			$1::uuid, $2::uuid, $3, $4, $5, $6, NULLIF($7, '')::uuid,
			NULLIF($8, ''), NULLIF($9, ''), NULLIF($10, ''), NULLIF($11, '')::uuid, NULLIF($12, '')::uuid,
			NULLIF($13, ''), NULLIF($14, ''),
			NULLIF($15, ''), $16, NULLIF($17, ''), $18, $19,
			NULLIF($20, ''), NULLIF($21, '')
		)
	`, entry.EntryID, entry.TenantID, string(entry.Kind), entry.AmountMC, entry.IdempotencyKey, entry.OccurredAt, entry.LotID,
		entry.RequestID, entry.TraceID, entry.Model, entry.UserID, entry.TeamID,
		entry.Product, entry.BillingMode,
		entry.PricingRevision, rateCardVersion, entry.InternalCostUSD, entry.CreditEquivalentMC, entry.SpendLogID,
		entry.Reason, entry.Actor)

	return err
}

// creditAllocation is one credit_ledger_allocations row. An empty LotID is debt.
type creditAllocation struct {
	EntryID  string
	LotID    string
	AmountMC int64
}

func insertCreditAllocationsTx(ctx context.Context, tx pgx.Tx, tenantID string, occurredAt time.Time, allocations []creditAllocation) error {
	if len(allocations) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, allocation := range allocations {
		batch.Queue(`
			INSERT INTO credit_ledger_allocations (allocation_id, entry_id, tenant_id, lot_id, amount_mc, occurred_at)
			VALUES ($1::uuid, $2::uuid, $3::uuid, NULLIF($4, '')::uuid, $5, $6)
		`, uuid.NewString(), allocation.EntryID, tenantID, allocation.LotID, allocation.AmountMC, occurredAt)
	}

	return tx.SendBatch(ctx, batch).Close()
}

// applyCreditTenantDeltaTx updates the tenant projection after a locked write.
func applyCreditTenantDeltaTx(ctx context.Context, tx pgx.Tx, tenantID string, availableDelta, debtDelta, ledgerDelta int64) error {
	_, err := tx.Exec(ctx, `
		UPDATE credit_tenant_balances
		SET available_mc = available_mc + $2,
		    debt_mc = debt_mc + $3,
		    ledger_sum_mc = ledger_sum_mc + $4,
		    updated_at = NOW()
		WHERE tenant_id = $1::uuid
	`, tenantID, availableDelta, debtDelta, ledgerDelta)

	return err
}

// debitCreditLotsTx spends amountMC from the tenant's lots valid at occurredAt,
// in consumption order, and returns the allocations, including any debt.
func debitCreditLotsTx(ctx context.Context, tx pgx.Tx, tenantID, entryID string, amountMC int64, occurredAt time.Time) ([]creditAllocation, int64, error) {
	if amountMC <= 0 {
		return nil, 0, nil
	}

	rows, err := tx.Query(ctx, `
		SELECT lot_id::text, remaining_mc, expires_at, source_rank, lot_created_at
		FROM credit_lot_balances
		WHERE tenant_id = $1::uuid AND remaining_mc > 0 AND (expires_at IS NULL OR expires_at > $2)
		ORDER BY expires_at ASC NULLS LAST, source_rank, lot_created_at, lot_id
		FOR UPDATE
	`, tenantID, occurredAt)
	if err != nil {
		return nil, 0, err
	}

	lots, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (credits.SpendableLot, error) {
		var lot credits.SpendableLot

		scanErr := row.Scan(&lot.ID, &lot.RemainingMC, &lot.ExpiresAt, &lot.SourceRank, &lot.CreatedAt)

		return lot, scanErr
	})
	if err != nil {
		return nil, 0, err
	}

	credits.SortForConsumption(lots)

	debits := credits.AllocateDebit(lots, amountMC)
	allocations := make([]creditAllocation, 0, len(debits))

	var debtMC int64

	for _, debit := range debits {
		allocations = append(allocations, creditAllocation{EntryID: entryID, LotID: debit.LotID, AmountMC: debit.AmountMC})
		if debit.LotID == "" {
			debtMC -= debit.AmountMC

			continue
		}

		if _, err := tx.Exec(ctx, `
			UPDATE credit_lot_balances SET remaining_mc = remaining_mc + $2, updated_at = NOW()
			WHERE lot_id = $1::uuid
		`, debit.LotID, debit.AmountMC); err != nil {
			return nil, 0, err
		}
	}

	return allocations, debtMC, nil
}
