package storage

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	billingdomain "github.com/Delpak/shahrah/services/control-plane/internal/domain/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

const creditPackOrderColumns = `
	o.order_id::text, o.public_id::text, o.tenant_id::text, o.pack_slug, o.pack_version, o.credits,
	o.validity_months, o.price_irr::bigint, o.tax_percent::text, o.tax_irr::bigint, o.payable_irr::bigint,
	o.provider, o.status, COALESCE(o.authority, ''), COALESCE(o.payment_url, ''), COALESCE(o.ref_id, ''),
	COALESCE(o.failure_reason, ''), COALESCE(o.callback_url, ''), COALESCE(o.return_url, ''),
	COALESCE(o.lot_id::text, ''), o.created_at, o.completed_at`

func scanCreditPackOrder(row pgx.Row) (*platform.CreditPackOrder, error) {
	var order platform.CreditPackOrder

	err := row.Scan(&order.ID, &order.PublicID, &order.TenantID, &order.PackSlug, &order.PackVersion, &order.Credits,
		&order.ValidityMonths, &order.PriceIRR, &order.TaxPercent, &order.TaxIRR, &order.PayableIRR,
		&order.Provider, &order.Status, &order.Authority, &order.PaymentURL, &order.RefID,
		&order.FailureReason, &order.CallbackURL, &order.ReturnURL, &order.CreditsLotID, &order.CreatedAt, &order.CompletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, billingdomain.ErrCreditPackOrderNotFound
	}

	if err != nil {
		return nil, err
	}

	return &order, nil
}

// creditTaxPercent is the VAT rate applied to purchases, shared with wallet top-ups.
func creditTaxPercent(ctx context.Context, q rowQuerier) (decimal.Decimal, error) {
	var raw string

	err := q.QueryRow(ctx, `
		SELECT COALESCE((SELECT tax_percent::numeric::text FROM wallet_topup_settings WHERE settings_id = true), '0')
	`).Scan(&raw)
	if err != nil {
		return decimal.Zero, err
	}

	return decimal.NewFromString(raw)
}

// ListCreditPacks returns the published packs priced at the current list price.
func (s *CreditPackRepository) ListCreditPacks(ctx context.Context) ([]platform.CreditPack, error) {
	listPrice, err := creditListPriceIRR(ctx, s.pool)
	if err != nil {
		return nil, err
	}

	taxPercent, err := creditTaxPercent(ctx, s.pool)
	if err != nil {
		return nil, err
	}

	rows, err := s.pool.Query(ctx, `
		SELECT slug, version, credits, discount_percent::text, validity_months
		FROM credit_packs
		WHERE status = 'published'
		ORDER BY sort_order, credits
	`)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (platform.CreditPack, error) {
		var pack platform.CreditPack

		if scanErr := row.Scan(&pack.Slug, &pack.Version, &pack.Credits, &pack.DiscountPercent, &pack.ValidityMonths); scanErr != nil {
			return pack, scanErr
		}

		discount, parseErr := decimal.NewFromString(pack.DiscountPercent)
		if parseErr != nil {
			return pack, parseErr
		}

		quote := credits.QuotePack(pack.Credits, discount, listPrice, taxPercent)
		pack.DiscountPercent = discount.StringFixed(2)
		pack.PriceIRR, pack.TaxIRR, pack.PayableIRR = quote.PriceIRR, quote.TaxIRR, quote.PayableIRR
		pack.TaxPercent = quote.TaxPercent.StringFixed(2)

		return pack, nil
	})
}

// CreateCreditPackOrder prices a published pack for a credit tenant and records
// the order. Bank-transfer orders wait for an approved transfer; gateway orders
// wait for the payment callback.
func (s *CreditPackRepository) CreateCreditPackOrder(ctx context.Context, tenantID string, req platform.CreateCreditPackOrderRequest, callbackURL, returnURL, actor string) (*platform.CreditPackOrder, error) {
	var billingUnit string

	err := s.pool.QueryRow(ctx, `SELECT billing_unit FROM tenants WHERE tenant_id = $1::uuid`, tenantID).Scan(&billingUnit)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && billingUnit != credits.BillingUnitCredit) {
		return nil, billingdomain.ErrCreditBillingNotEnabled
	}

	if err != nil {
		return nil, err
	}

	var (
		packID, discountRaw string
		version, validity   int
		packCredits         int64
	)

	err = s.pool.QueryRow(ctx, `
		SELECT pack_id::text, version, credits, discount_percent::text, validity_months
		FROM credit_packs WHERE slug = $1 AND status = 'published'
	`, strings.TrimSpace(req.PackSlug)).Scan(&packID, &version, &packCredits, &discountRaw, &validity)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, billingdomain.ErrCreditPackNotFound
	}

	if err != nil {
		return nil, err
	}

	quote, err := s.quoteCreditPack(ctx, packCredits, discountRaw)
	if err != nil {
		return nil, err
	}

	status := platform.CreditPackOrderPending
	if req.Provider == platform.CreditPackProviderBankTransfer {
		status = platform.CreditPackOrderAwaitingTransfer
	}

	return scanCreditPackOrder(s.pool.QueryRow(ctx, `
		WITH inserted AS (
			INSERT INTO credit_pack_orders (
				tenant_id, pack_id, pack_slug, pack_version, credits, validity_months,
				list_price_irr_per_credit, discount_percent, price_irr, tax_percent, tax_irr, payable_irr,
				provider, status, callback_url, return_url, created_by
			)
			VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8::numeric, $9, $10::numeric, $11, $12,
			        $13, $14, NULLIF($15, ''), NULLIF($16, ''), NULLIF($17, ''))
			RETURNING *
		)
		SELECT `+creditPackOrderColumns+` FROM inserted o
	`, tenantID, packID, strings.TrimSpace(req.PackSlug), version, packCredits, validity,
		quote.ListPriceIRRPerCredit, discountRaw, quote.PriceIRR, quote.TaxPercent.String(), quote.TaxIRR, quote.PayableIRR,
		req.Provider, status, callbackURL, returnURL, actor))
}

func (s *CreditPackRepository) quoteCreditPack(ctx context.Context, packCredits int64, discountRaw string) (credits.PackQuote, error) {
	discount, err := decimal.NewFromString(discountRaw)
	if err != nil {
		return credits.PackQuote{}, err
	}

	listPrice, err := creditListPriceIRR(ctx, s.pool)
	if err != nil {
		return credits.PackQuote{}, err
	}

	taxPercent, err := creditTaxPercent(ctx, s.pool)
	if err != nil {
		return credits.PackQuote{}, err
	}

	return credits.QuotePack(packCredits, discount, listPrice, taxPercent), nil
}

// AttachCreditPackOrderGatewayData stores the gateway authority of a pending order.
func (s *CreditPackRepository) AttachCreditPackOrderGatewayData(ctx context.Context, publicID, authority, paymentURL string, raw map[string]interface{}) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE credit_pack_orders
		SET authority = $2, payment_url = $3, raw_payload = raw_payload || $4::jsonb, updated_at = NOW()
		WHERE public_id::text = $1 AND status = 'PENDING'
	`, publicID, authority, paymentURL, marshalJSONBestEffort(raw))
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return billingdomain.ErrCreditPackOrderNotFound
	}

	return nil
}

// GetCreditPackOrderByAuthority finds an order by its gateway authority.
func (s *CreditPackRepository) GetCreditPackOrderByAuthority(ctx context.Context, authority string) (*platform.CreditPackOrder, error) {
	return scanCreditPackOrder(s.pool.QueryRow(ctx, `SELECT `+creditPackOrderColumns+` FROM credit_pack_orders o WHERE o.authority = $1`, authority))
}

// GetCreditPackOrder returns one of the tenant's orders by its public id.
func (s *CreditPackRepository) GetCreditPackOrder(ctx context.Context, tenantID, publicID string) (*platform.CreditPackOrder, error) {
	return scanCreditPackOrder(s.pool.QueryRow(ctx, `
		SELECT `+creditPackOrderColumns+` FROM credit_pack_orders o
		WHERE o.tenant_id = $1::uuid AND o.public_id::text = $2
	`, tenantID, strings.TrimSpace(publicID)))
}

// ListCreditPackOrders returns the tenant's most recent orders.
func (s *CreditPackRepository) ListCreditPackOrders(ctx context.Context, tenantID string, limit int) ([]platform.CreditPackOrder, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	rows, err := s.pool.Query(ctx, `
		SELECT `+creditPackOrderColumns+` FROM credit_pack_orders o
		WHERE o.tenant_id = $1::uuid
		ORDER BY o.created_at DESC, o.order_id DESC
		LIMIT $2
	`, tenantID, limit)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (platform.CreditPackOrder, error) {
		order, scanErr := scanCreditPackOrder(row)
		if scanErr != nil {
			return platform.CreditPackOrder{}, scanErr
		}

		return *order, nil
	})
}

// CompleteCreditPackOrderByAuthority grants the credits of a paid gateway order.
// A repeated callback for a completed order grants nothing and reports duplicate.
func (s *CreditPackRepository) CompleteCreditPackOrderByAuthority(ctx context.Context, authority string, verified billingdomain.PaymentVerification) (*platform.CreditPackOrder, bool, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, false, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	order, err := scanCreditPackOrder(tx.QueryRow(ctx, `SELECT `+creditPackOrderColumns+` FROM credit_pack_orders o WHERE o.authority = $1 FOR UPDATE`, authority))
	if err != nil {
		return nil, false, err
	}

	if order.Status == platform.CreditPackOrderCompleted {
		return order, true, tx.Commit(ctx)
	}

	if order.Status != platform.CreditPackOrderPending {
		return nil, false, billingdomain.ErrCreditPackOrderNotPayable
	}

	if _, grantErr := completeCreditPackOrderTx(ctx, tx, order, time.Now().UTC()); grantErr != nil {
		return nil, false, grantErr
	}

	if _, updateErr := tx.Exec(ctx, `
		UPDATE credit_pack_orders
		SET ref_id = NULLIF($2, ''), gateway_txn_id = NULLIF($3, ''), card_mask = NULLIF($4, ''),
		    payer_phone = NULLIF($5, ''), raw_payload = raw_payload || $6::jsonb, updated_at = NOW()
		WHERE order_id = $1::uuid
	`, order.ID, verified.RefID, verified.GatewayTxnID, verified.CardMask, verified.PayerPhone, marshalJSONBestEffort(verified.Raw)); updateErr != nil {
		return nil, false, updateErr
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		return nil, false, commitErr
	}

	completed, err := s.GetCreditPackOrderByAuthority(ctx, authority)

	return completed, false, err
}

// MarkCreditPackOrderFailedByAuthority records a failed or canceled payment.
// A completed order is never downgraded.
func (s *CreditPackRepository) MarkCreditPackOrderFailedByAuthority(ctx context.Context, authority, reason string, raw map[string]interface{}) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE credit_pack_orders
		SET status = CASE WHEN status = 'PENDING' THEN 'FAILED' ELSE status END,
		    failure_reason = CASE WHEN status = 'PENDING' THEN $2 ELSE failure_reason END,
		    raw_payload = raw_payload || $3::jsonb, updated_at = NOW()
		WHERE authority = $1
	`, authority, reason, marshalJSONBestEffort(raw))
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return billingdomain.ErrCreditPackOrderNotFound
	}

	return nil
}

// completeCreditPackOrderTx grants a paid order's credits as a 12-month add-on
// lot and marks the order completed. The purchase:{order_id} key makes it
// impossible to grant one order twice.
func completeCreditPackOrderTx(ctx context.Context, tx pgx.Tx, order *platform.CreditPackOrder, now time.Time) (string, error) {
	expiresAt := credits.PackExpiry(now, order.ValidityMonths)
	priceIRR := order.PriceIRR

	grant := CreditGrant{
		TenantID: order.TenantID, Kind: credits.EntryPurchase, Source: credits.LotSourcePurchaseAddon,
		AmountMC: order.Credits * credits.MillicreditsPerCredit, ExpiresAt: &expiresAt, OccurredAt: now,
		IdempotencyKey: credits.PurchaseKey(order.ID), PurchasePriceIRR: &priceIRR,
		PricePerCreditIRR: credits.PricePerCreditIRR(order.PriceIRR, order.Credits).String(),
		SourceRef: map[string]interface{}{
			"order_id": order.PublicID, "pack_slug": order.PackSlug, "pack_version": order.PackVersion, "provider": order.Provider,
		},
	}
	if err := validateCreditGrant(grant); err != nil {
		return "", err
	}

	result, err := grantCreditsTx(ctx, tx, grant)
	if err != nil {
		return "", err
	}

	if _, err := tx.Exec(ctx, `
		UPDATE credit_pack_orders
		SET status = 'COMPLETED', lot_id = $2::uuid, completed_at = $3, failure_reason = NULL, updated_at = NOW()
		WHERE order_id = $1::uuid
	`, order.ID, result.LotID, now); err != nil {
		return "", err
	}

	if err := insertFinancialAuditTx(ctx, tx, order.TenantID, "credit_pack.purchased", platform.SupportActor{ID: "system", Role: order.Provider},
		"credit_pack_order", order.PublicID, grant.IdempotencyKey, map[string]interface{}{
			"lot_id": result.LotID, "credits": order.Credits, "price_irr": order.PriceIRR, "payable_irr": order.PayableIRR,
		}); err != nil {
		return "", err
	}

	return result.LotID, nil
}
