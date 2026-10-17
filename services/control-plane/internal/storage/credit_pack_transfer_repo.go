package storage

import (
	"context"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	billingdomain "github.com/Delpak/shahrah/services/control-plane/internal/domain/billing"
	"github.com/jackc/pgx/v5"
)

// createCreditPackBankTransfer records a bank transfer that pays a credit pack
// order. The amount must match the order exactly, so review never has to judge
// partial or excess payments.
func (s *FinancialRepository) createCreditPackBankTransfer(ctx context.Context, tenantID string, req platform.CreateBankTransferRequest, actor string) (*platform.BankTransferRequest, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	order, err := scanCreditPackOrder(tx.QueryRow(ctx, `
		SELECT `+creditPackOrderColumns+` FROM credit_pack_orders o
		WHERE o.tenant_id = $1::uuid AND o.public_id::text = $2
		FOR UPDATE
	`, tenantID, strings.TrimSpace(req.CreditPackOrderID)))
	if err != nil {
		return nil, err
	}

	if order.Status != platform.CreditPackOrderAwaitingTransfer {
		return nil, billingdomain.ErrCreditPackOrderNotPayable
	}

	if req.AmountIRR != order.PayableIRR {
		return nil, billingdomain.ErrCreditPackTransferMismatch
	}

	metadata := req.Metadata
	if metadata == nil {
		metadata = map[string]interface{}{}
	}

	metadata["credit_pack_order_id"] = order.PublicID

	item, err := scanBankTransferRequest(tx.QueryRow(ctx, `
		INSERT INTO bank_transfer_requests (
			tenant_id, amount_irr, tracking_number, bank_transaction_at, receipt_file_url,
			status, metadata, created_by, credit_pack_order_id
		)
		VALUES ($1::uuid, $2, $3, $4, NULLIF($5, ''), 'pending_review', $6, NULLIF($7, ''), $8::uuid)
		RETURNING bank_transfer_request_id::text, public_id::text, tenant_id::text, amount_irr::bigint,
		          tracking_number, bank_transaction_at, COALESCE(receipt_file_url, ''), status,
		          COALESCE(rejection_reason, ''), COALESCE(wallet_tx_id::text, ''), COALESCE(reviewed_by, ''),
		          reviewed_at, metadata, COALESCE(created_by, ''), created_at, updated_at
	`, tenantID, req.AmountIRR, strings.TrimSpace(req.TrackingNumber), req.BankTransactionAt, strings.TrimSpace(req.ReceiptFileURL),
		marshalJSONBestEffort(metadata), strings.TrimSpace(actor), order.ID))
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return item, nil
}

// approveCreditPackBankTransferTx completes the credit pack order a reviewed
// transfer pays, instead of crediting the IRR wallet.
func (s *FinancialRepository) approveCreditPackBankTransferTx(ctx context.Context, tx pgx.Tx, item *platform.BankTransferRequest, orderID, idempotency string, actor platform.SupportActor) (*platform.BankTransferRequest, error) {
	order, err := scanCreditPackOrder(tx.QueryRow(ctx, `
		SELECT `+creditPackOrderColumns+` FROM credit_pack_orders o WHERE o.order_id = $1::uuid FOR UPDATE
	`, orderID))
	if err != nil {
		return nil, err
	}

	if order.Status != platform.CreditPackOrderAwaitingTransfer || order.PayableIRR != item.AmountIRR {
		return nil, billingdomain.ErrCreditPackOrderNotPayable
	}

	lotID, err := completeCreditPackOrderTx(ctx, tx, order, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx, `
		UPDATE credit_pack_orders SET ref_id = $2, updated_at = NOW() WHERE order_id = $1::uuid
	`, order.ID, item.TrackingNumber); err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx, `
		UPDATE bank_transfer_requests
		SET status = 'approved', reviewed_by = NULLIF($2, ''), reviewed_at = NOW(), updated_at = NOW()
		WHERE bank_transfer_request_id::text = $1
	`, item.ID, actor.ID); err != nil {
		return nil, err
	}

	if err := insertFinancialAuditTx(ctx, tx, item.TenantID, "bank_transfer.approved", actor, "bank_transfer_request", item.ID, idempotency,
		map[string]interface{}{"credit_pack_order_id": order.PublicID, "lot_id": lotID}); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	s.invalidateCreditTenant(ctx, item.TenantID, "")

	return s.GetBankTransferRequest(ctx, item.TenantID, item.ID)
}
