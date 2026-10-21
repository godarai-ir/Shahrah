package storage

import (
	"context"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	billingdomain "github.com/Delpak/shahrah/services/control-plane/internal/domain/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/seats"
	"github.com/jackc/pgx/v5"
)

// seatTransferHold is how long a seat order waits for its receipt. Its price and
// period are fixed when it is created, so an old order is canceled, not paid.
const seatTransferHold = 72 * time.Hour

// checkSeatTransferSubmit says whether a receipt may be filed for the order.
func checkSeatTransferSubmit(order *platform.SeatOrder, amountIRR int64, now time.Time) error {
	if order.Status != platform.CreditPackOrderAwaitingTransfer || now.Sub(order.CreatedAt) > seatTransferHold {
		return seats.ErrOrderNotPayable
	}

	if amountIRR != order.PayableIRR {
		return billingdomain.ErrCreditPackTransferMismatch
	}

	return nil
}

// checkSeatTransferApprove says whether a reviewed receipt may complete the
// order. Seats whose paid period has already ended would be paid but worthless.
func checkSeatTransferApprove(order *platform.SeatOrder, amountIRR int64, now time.Time) error {
	if order.Status != platform.CreditPackOrderAwaitingTransfer || order.PayableIRR != amountIRR || !order.PeriodEnd.After(now) {
		return seats.ErrOrderNotPayable
	}

	return nil
}

// createSeatBankTransfer records a bank transfer that pays a seat order. The
// amount must match the order exactly, so review never has to judge partial or
// excess payments.
func (s *FinancialRepository) createSeatBankTransfer(ctx context.Context, tenantID string, req platform.CreateBankTransferRequest, actor string) (*platform.BankTransferRequest, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	order, err := scanSeatOrder(tx.QueryRow(ctx, `
		SELECT `+seatOrderColumns+` FROM seat_orders o
		WHERE o.tenant_id = $1::uuid AND o.public_id::text = $2
		FOR UPDATE
	`, tenantID, strings.TrimSpace(req.SeatOrderID)))
	if err != nil {
		return nil, err
	}

	if checkErr := checkSeatTransferSubmit(order, req.AmountIRR, time.Now()); checkErr != nil {
		return nil, checkErr
	}

	return insertOrderBankTransfer(ctx, tx, tenantID, req, actor, "seat_order_id", order.PublicID, order.ID)
}

// approveSeatBankTransferTx completes the seat order a reviewed transfer pays,
// instead of crediting the IRR wallet.
func (s *FinancialRepository) approveSeatBankTransferTx(ctx context.Context, tx pgx.Tx, item *platform.BankTransferRequest, orderID, idempotency string, actor platform.SupportActor) (*platform.BankTransferRequest, error) {
	order, err := scanSeatOrder(tx.QueryRow(ctx, `
		SELECT `+seatOrderColumns+` FROM seat_orders o WHERE o.order_id = $1::uuid FOR UPDATE
	`, orderID))
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	if checkErr := checkSeatTransferApprove(order, item.AmountIRR, now); checkErr != nil {
		return nil, checkErr
	}

	blockID, err := applySeatOrderTx(ctx, tx, order, now)
	if err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx, `
		UPDATE seat_orders
		SET status = 'COMPLETED', block_id = NULLIF($2, '')::uuid, completed_at = $3, failure_reason = NULL,
		    ref_id = NULLIF($4, ''), updated_at = NOW()
		WHERE order_id = $1::uuid
	`, order.ID, blockID, now, item.TrackingNumber); err != nil {
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
		map[string]interface{}{"seat_order_id": order.PublicID, "block_id": blockID}); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	s.afterSeatChange(ctx, order.TenantID, order.AssignUserID, now)

	return s.GetBankTransferRequest(ctx, item.TenantID, item.ID)
}
