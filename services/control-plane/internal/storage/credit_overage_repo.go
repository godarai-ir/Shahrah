package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

// settleOverageTx invoices a tenant's debt as overage before a new plan slice,
// when the tenant enabled overage. The debt is billed in IRR at the list price
// and cleared by a fully allocated overage lot, so the new slice is not used to
// repay it. Without overage the slice repays the debt instead. The caller holds
// the tenant lock.
func settleOverageTx(ctx context.Context, tx pgx.Tx, sub creditSubscription, sliceStart time.Time, debtMC int64, now time.Time) (int64, error) {
	if debtMC <= 0 {
		return 0, nil
	}

	var enabled bool
	if err := tx.QueryRow(ctx, `
		SELECT COALESCE((SELECT overage_enabled FROM credit_tenant_settings WHERE tenant_id = $1::uuid), false)
	`, sub.TenantID).Scan(&enabled); err != nil || !enabled {
		return 0, err
	}

	listPrice, err := creditListPriceIRR(ctx, tx)
	if err != nil {
		return 0, err
	}

	chargeIRR := credits.OverageChargeIRR(debtMC, listPrice)
	key := credits.OverageKey(sub.SubscriptionID, sliceStart)

	lot, err := grantCreditsTx(ctx, tx, CreditGrant{
		TenantID: sub.TenantID, Kind: credits.EntryPurchase, Source: credits.LotSourceOverage,
		AmountMC: debtMC, IdempotencyKey: key, OccurredAt: now, PurchasePriceIRR: &chargeIRR,
		PricePerCreditIRR: decimal.NewFromInt(listPrice).String(),
		SourceRef:         map[string]interface{}{"subscription_id": sub.SubscriptionID, "slice_start": sliceStart},
	})
	if err != nil || lot.Duplicate {
		return 0, err
	}

	if err := insertOverageInvoiceTx(ctx, tx, sub, sliceStart, debtMC, chargeIRR, listPrice, lot.LotID); err != nil {
		return 0, err
	}

	return debtMC, nil
}

// insertOverageInvoiceTx issues the IRR invoice for invoiced overage. It has no
// subscription_id so it never collides with the period-end hybrid invoice.
func insertOverageInvoiceTx(ctx context.Context, tx pgx.Tx, sub creditSubscription, sliceStart time.Time, debtMC, chargeIRR, listPrice int64, lotID string) error {
	invoiceID := uuid.NewString()
	number := fmt.Sprintf("OVG-%s-%s", sliceStart.UTC().Format("200601"), strings.Split(sub.SubscriptionID, "-")[0])
	periodStart := sliceStart.AddDate(0, -1, 0)

	if _, err := tx.Exec(ctx, `
		INSERT INTO invoices (
			invoice_id, invoice_number, tenant_id, status, issued_at, period_start, period_end,
			billing_profile_snapshot, raw_usage_irr, final_amount_irr, metadata
		)
		VALUES (
			$1::uuid, $2, $3::uuid, 'issued', NOW(), $4, $5,
			COALESCE((SELECT jsonb_build_object('profile_type', profile_type, 'data', data)
			          FROM billing_profiles WHERE tenant_id = $3::uuid), '{}'::jsonb),
			$6, $6,
			jsonb_build_object('settlement_model', 'credit_overage', 'subscription_id', $7::text,
			                   'overage_mc', $8::bigint, 'list_price_irr_per_credit', $9::bigint, 'lot_id', $10::text)
		)
	`, invoiceID, number, sub.TenantID, periodStart, sliceStart, chargeIRR, sub.SubscriptionID, debtMC, listPrice, lotID); err != nil {
		return err
	}

	_, err := tx.Exec(ctx, `
		INSERT INTO invoice_line_items (line_item_id, invoice_id, item_type, provider_name, description, quantity, amount_irr, metadata)
		VALUES ($1::uuid, $2::uuid, 'credit_overage', '', 'Credit overage at list price', $3::numeric, $4,
		        jsonb_build_object('overage_mc', $5::bigint, 'list_price_irr_per_credit', $6::bigint))
	`, uuid.NewString(), invoiceID, credits.Format(credits.Millicredits(debtMC)), chargeIRR, debtMC, listPrice)

	return err
}
