package storage

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	billingdomain "github.com/Delpak/shahrah/services/control-plane/internal/domain/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

// ListCreditPackVersions returns every version of every pack.
func (s *CreditPackRepository) ListCreditPackVersions(ctx context.Context) ([]platform.CreditPackVersion, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT slug, version, credits, discount_percent::text, validity_months, status, sort_order,
		       COALESCE(created_by, ''), created_at
		FROM credit_packs ORDER BY sort_order, slug, version DESC
	`)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (platform.CreditPackVersion, error) {
		var pack platform.CreditPackVersion

		scanErr := row.Scan(&pack.Slug, &pack.Version, &pack.Credits, &pack.DiscountPercent, &pack.ValidityMonths,
			&pack.Status, &pack.SortOrder, &pack.CreatedBy, &pack.CreatedAt)

		return pack, scanErr
	})
}

// CreateCreditPackVersion drafts the next version of a pack.
func (s *CreditPackRepository) CreateCreditPackVersion(ctx context.Context, req platform.CreateCreditPackVersionRequest, actor platform.SupportActor) (*platform.CreditPackVersion, error) {
	slug := strings.ToLower(strings.TrimSpace(req.Slug))

	discount, err := decimal.NewFromString(strings.TrimSpace(req.DiscountPercent))
	if err != nil || slug == "" || req.Credits <= 0 || discount.IsNegative() || discount.GreaterThanOrEqual(decimal.NewFromInt(100)) ||
		req.ValidityMonths <= 0 || req.ValidityMonths > 36 {
		return nil, billingdomain.ErrInvalidCreditPack
	}

	var pack platform.CreditPackVersion

	err = s.pool.QueryRow(ctx, `
		INSERT INTO credit_packs (slug, version, credits, discount_percent, validity_months, status, sort_order, created_by)
		SELECT $1::text, COALESCE(MAX(version), 0) + 1, $2, $3::numeric, $4, 'draft', $5, NULLIF($6, '')
		FROM credit_packs WHERE slug = $1::text
		RETURNING slug, version, credits, discount_percent::text, validity_months, status, sort_order, COALESCE(created_by, ''), created_at
	`, slug, req.Credits, discount.String(), req.ValidityMonths, req.SortOrder, actor.ID).Scan(
		&pack.Slug, &pack.Version, &pack.Credits, &pack.DiscountPercent, &pack.ValidityMonths,
		&pack.Status, &pack.SortOrder, &pack.CreatedBy, &pack.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &pack, nil
}

// PublishCreditPackVersion makes one version the purchasable one. Orders keep
// the version and price they were placed at.
func (s *CreditPackRepository) PublishCreditPackVersion(ctx context.Context, slug string, version int, actor platform.SupportActor) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer rollbackTxBestEffort(ctx, tx)

	if _, execErr := tx.Exec(ctx, `
		UPDATE credit_packs SET status = 'archived', updated_at = NOW() WHERE slug = $1 AND status = 'published'
	`, slug); execErr != nil {
		return execErr
	}

	tag, err := tx.Exec(ctx, `
		UPDATE credit_packs SET status = 'published', updated_at = NOW() WHERE slug = $1 AND version = $2 AND status = 'draft'
	`, slug, version)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return billingdomain.ErrCreditPackVersionNotFound
	}

	if auditErr := insertFinancialAuditTx(ctx, tx, "", "credit_pack.published", actor, "credit_pack", slug, "",
		map[string]interface{}{"slug": slug, "version": version}); auditErr != nil {
		return auditErr
	}

	return tx.Commit(ctx)
}

// GetTenantCreditOverview summarizes a tenant's balance and settings.
func (s *CreditLedgerRepository) GetTenantCreditOverview(ctx context.Context, tenantID string) (*platform.TenantCreditOverview, error) {
	settings, err := (&CreditPackRepository{walletState: s.walletState}).GetCreditSettings(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	balance, err := s.GetCreditBalance(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	return &platform.TenantCreditOverview{
		TenantID:         tenantID,
		BillingUnit:      settings.BillingUnit,
		AvailableCredits: credits.Format(credits.Millicredits(balance.AvailableMC)),
		DebtCredits:      credits.Format(credits.Millicredits(balance.DebtMC)),
		LedgerSumCredits: credits.Format(credits.Millicredits(balance.LedgerSumMC)),
		Settings:         *settings,
	}, nil
}

// ListTenantCreditLedger pages a tenant's ledger, newest first. before, when
// set, returns entries strictly older than it.
func (s *CreditLedgerRepository) ListTenantCreditLedger(ctx context.Context, tenantID string, limit int, before *time.Time) ([]platform.CreditLedgerEntry, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	rows, err := s.pool.Query(ctx, `
		SELECT entry_id::text, kind, amount_mc, occurred_at, COALESCE(lot_id::text, ''), COALESCE(request_id, ''),
		       COALESCE(model, ''), COALESCE(user_id::text, ''), COALESCE(team_id::text, ''), COALESCE(product, ''),
		       COALESCE(billing_mode, ''), credit_equivalent_mc, COALESCE(reason, ''), COALESCE(actor, '')
		FROM credit_ledger
		WHERE tenant_id = $1::uuid AND ($2::timestamptz IS NULL OR occurred_at < $2)
		ORDER BY occurred_at DESC, entry_id DESC
		LIMIT $3
	`, tenantID, before, limit)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (platform.CreditLedgerEntry, error) {
		var (
			entry        platform.CreditLedgerEntry
			amountMC     int64
			equivalentMC *int64
		)

		scanErr := row.Scan(&entry.EntryID, &entry.Kind, &amountMC, &entry.OccurredAt, &entry.LotID, &entry.RequestID,
			&entry.Model, &entry.UserID, &entry.TeamID, &entry.Product, &entry.BillingMode, &equivalentMC, &entry.Reason, &entry.Actor)
		entry.AmountCredits = formatSignedCredits(amountMC)

		if equivalentMC != nil {
			formatted := credits.Format(credits.Millicredits(*equivalentMC))
			entry.CreditEquivalent = &formatted
		}

		return entry, scanErr
	})
}

// ListTenantCreditLots returns a tenant's lots, spendable ones first.
func (s *CreditLedgerRepository) ListTenantCreditLots(ctx context.Context, tenantID string) ([]platform.CreditLot, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT l.lot_id::text, l.source, l.amount_mc, COALESCE(b.remaining_mc, 0), l.expires_at,
		       l.purchase_price_irr::bigint, l.price_per_credit_irr::text, l.source_ref, l.created_at
		FROM credit_lots l
		LEFT JOIN credit_lot_balances b ON b.lot_id = l.lot_id
		WHERE l.tenant_id = $1::uuid
		ORDER BY (COALESCE(b.remaining_mc, 0) > 0) DESC, l.expires_at ASC NULLS LAST, l.created_at DESC
		LIMIT 500
	`, tenantID)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (platform.CreditLot, error) {
		var (
			lot                   platform.CreditLot
			amountMC, remainingMC int64
			sourceRef             []byte
		)

		scanErr := row.Scan(&lot.LotID, &lot.Source, &amountMC, &remainingMC, &lot.ExpiresAt,
			&lot.PurchasePriceIRR, &lot.PricePerCreditIRR, &sourceRef, &lot.CreatedAt)
		lot.AmountCredits = credits.Format(credits.Millicredits(amountMC))
		lot.RemainingCredits = credits.Format(credits.Millicredits(remainingMC))
		lot.SourceRef = parseJSONMap(sourceRef)

		return lot, scanErr
	})
}

// ListCreditLotAllocations returns every movement into or out of one lot.
func (s *CreditLedgerRepository) ListCreditLotAllocations(ctx context.Context, tenantID, lotID string) ([]platform.CreditLotAllocation, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT a.entry_id::text, l.kind, a.amount_mc, a.occurred_at
		FROM credit_ledger_allocations a
		JOIN credit_ledger l ON l.entry_id = a.entry_id
		WHERE a.tenant_id = $1::uuid AND a.lot_id = $2::uuid
		ORDER BY a.occurred_at, a.allocation_id
	`, tenantID, lotID)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (platform.CreditLotAllocation, error) {
		var (
			allocation platform.CreditLotAllocation
			amountMC   int64
		)

		scanErr := row.Scan(&allocation.EntryID, &allocation.Kind, &amountMC, &allocation.OccurredAt)
		allocation.AmountCredits = formatSignedCredits(amountMC)

		return allocation, scanErr
	})
}

func formatSignedCredits(amountMC int64) string {
	if amountMC < 0 {
		return "-" + credits.Format(credits.Millicredits(-amountMC))
	}

	return credits.Format(credits.Millicredits(amountMC))
}

// AdjustCredits corrects a tenant's balance with an audited adjust entry. A
// positive amount becomes a lot; a negative one is spent from lots in the usual
// order, with any shortfall recorded as debt. The idempotency key is the
// caller's, so a retried request adjusts once.
func (s *CreditLedgerRepository) AdjustCredits(ctx context.Context, tenantID string, req platform.CreditAdjustmentRequest, actor platform.SupportActor) (*platform.TenantCreditOverview, error) {
	amount, err := decimal.NewFromString(strings.TrimSpace(req.AmountCredits))
	reason := strings.TrimSpace(req.Reason)
	key := strings.TrimSpace(req.IdempotencyKey)

	if err != nil || amount.IsZero() || !amount.Equal(amount.Round(2)) || reason == "" || key == "" {
		return nil, billingdomain.ErrInvalidCreditAdjustment
	}

	amountMC := amount.Mul(decimal.NewFromInt(credits.MillicreditsPerCredit)).IntPart()
	key = "adjust:" + tenantID + ":" + key
	now := time.Now().UTC()

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	var applied bool

	if amountMC > 0 {
		var granted CreditGrantResult

		granted, err = grantCreditsTx(ctx, tx, CreditGrant{
			TenantID: tenantID, Kind: credits.EntryAdjust, Source: credits.LotSourceAdjust, AmountMC: amountMC,
			ExpiresAt: req.ExpiresAt, IdempotencyKey: key, OccurredAt: now, Reason: reason, Actor: actor.ID,
			SourceRef: map[string]interface{}{"actor": actor.ID},
		})
		applied = err == nil && !granted.Duplicate
	} else {
		applied, err = debitAdjustmentTx(ctx, tx, tenantID, -amountMC, key, reason, actor.ID, now)
	}

	if err != nil {
		return nil, err
	}

	if !applied {
		// A retried request: the first one already adjusted and audited.
		return s.GetTenantCreditOverview(ctx, tenantID)
	}

	if auditErr := insertFinancialAuditTx(ctx, tx, tenantID, "credit.adjusted", actor, "credit_ledger", "", key,
		map[string]interface{}{"amount_mc": amountMC, "reason": reason}); auditErr != nil {
		return nil, auditErr
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		return nil, commitErr
	}

	s.invalidateCreditTenant(ctx, tenantID, "")

	return s.GetTenantCreditOverview(ctx, tenantID)
}

func debitAdjustmentTx(ctx context.Context, tx pgx.Tx, tenantID string, debitMC int64, key, reason, actor string, now time.Time) (bool, error) {
	entryID := uuid.NewString()

	claimed, err := claimCreditKeyTx(ctx, tx, key, entryID, tenantID, now)
	if err != nil || !claimed {
		return false, err
	}

	if _, lockErr := lockCreditTenantTx(ctx, tx, tenantID); lockErr != nil {
		return false, lockErr
	}

	if entryErr := insertCreditEntryTx(ctx, tx, creditEntry{
		EntryID: entryID, TenantID: tenantID, Kind: credits.EntryAdjust, AmountMC: -debitMC,
		IdempotencyKey: key, OccurredAt: now, Reason: reason, Actor: actor,
	}); entryErr != nil {
		return false, entryErr
	}

	allocations, debtMC, err := debitCreditLotsTx(ctx, tx, tenantID, entryID, debitMC, now)
	if err != nil {
		return false, err
	}

	if allocErr := insertCreditAllocationsTx(ctx, tx, tenantID, now, allocations); allocErr != nil {
		return false, allocErr
	}

	return true, applyCreditTenantDeltaTx(ctx, tx, tenantID, -(debitMC - debtMC), debtMC, -debitMC)
}

// SetTenantBillingUnitAudited switches a tenant between IRR and credit billing
// with an audit event.
func (s *CreditLedgerRepository) SetTenantBillingUnitAudited(ctx context.Context, tenantID string, req platform.SetBillingUnitRequest, actor platform.SupportActor) error {
	unit := strings.TrimSpace(req.BillingUnit)
	reason := strings.TrimSpace(req.Reason)

	if unit != credits.BillingUnitIRR && unit != credits.BillingUnitCredit {
		return errors.New("billing unit must be irr or credit")
	}

	if reason == "" {
		return billingdomain.ErrDecisionReasonRequired
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer rollbackTxBestEffort(ctx, tx)

	var previous string
	if scanErr := tx.QueryRow(ctx, `
		SELECT billing_unit FROM tenants WHERE tenant_id = $1::uuid FOR UPDATE
	`, tenantID).Scan(&previous); scanErr != nil {
		return scanErr
	}

	if _, execErr := tx.Exec(ctx, `UPDATE tenants SET billing_unit = $2 WHERE tenant_id = $1::uuid`, tenantID, unit); execErr != nil {
		return execErr
	}

	if auditErr := insertFinancialAuditTx(ctx, tx, tenantID, "credit.billing_unit_changed", actor, "tenant", tenantID, "",
		map[string]interface{}{"from": previous, "to": unit, "reason": reason}); auditErr != nil {
		return auditErr
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	// The billing unit is compiled into the runtime profile, which decides
	// how the gateway reserves this tenant's requests.
	s.invalidateCreditTenant(ctx, tenantID, "billing_unit.changed")

	return nil
}
