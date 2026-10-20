package storage

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	billingdomain "github.com/Delpak/shahrah/services/control-plane/internal/domain/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/seats"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

const seatOrderColumns = `
	o.order_id::text, o.public_id::text, o.tenant_id::text, o.kind, o.seat_type, o.quantity, o.billing_interval,
	o.period_start, o.period_end, o.remaining_days, o.period_days,
	o.price_irr::bigint, o.tax_percent::text, o.tax_irr::bigint, o.payable_irr::bigint,
	COALESCE(o.assign_user_id::text, ''), COALESCE(o.upgrade_request_id::text, ''), o.renew_block_ids::text[],
	o.provider, o.status, COALESCE(o.authority, ''), COALESCE(o.payment_url, ''), COALESCE(o.ref_id, ''),
	COALESCE(o.failure_reason, ''), COALESCE(o.callback_url, ''), COALESCE(o.return_url, ''),
	o.created_at, o.completed_at`

func scanSeatOrder(row pgx.Row) (*platform.SeatOrder, error) {
	var order platform.SeatOrder

	err := row.Scan(&order.ID, &order.PublicID, &order.TenantID, &order.Kind, &order.Seat, &order.Quantity, &order.BillingInterval,
		&order.PeriodStart, &order.PeriodEnd, &order.RemainingDays, &order.PeriodDays,
		&order.PriceIRR, &order.TaxPercent, &order.TaxIRR, &order.PayableIRR,
		&order.AssignUserID, &order.RequestID, &order.RenewBlockIDs,
		&order.Provider, &order.Status, &order.Authority, &order.PaymentURL, &order.RefID,
		&order.FailureReason, &order.CallbackURL, &order.ReturnURL, &order.CreatedAt, &order.CompletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, seats.ErrOrderNotFound
	}

	if err != nil {
		return nil, err
	}

	return &order, nil
}

// paidSeatType parses a seat type that tenants can buy.
func paidSeatType(raw string) (seats.Type, error) {
	seatType, err := seats.ParseType(raw)
	if err != nil || !seatType.Paid() {
		return "", seats.ErrInvalidSeatType
	}

	return seatType, nil
}

// QuoteSeats prices buying paid seats from now to the end of the tenant's
// current billing period.
func (s *SeatRepository) QuoteSeats(ctx context.Context, tenantID string, req platform.SeatQuoteRequest, now time.Time) (*platform.SeatQuote, error) {
	return s.quoteSeats(ctx, s.pool, tenantID, req, now.UTC())
}

func (s *SeatRepository) quoteSeats(ctx context.Context, q interface {
	rowQuerier
	rowsQuerier
}, tenantID string, req platform.SeatQuoteRequest, now time.Time,
) (*platform.SeatQuote, error) {
	seatType, err := paidSeatType(req.Seat)
	if err != nil {
		return nil, err
	}

	if req.Quantity < 1 || req.Quantity > seats.MaxQuantityPerOrder {
		return nil, seats.ErrInvalidQuantity
	}

	sub, err := currentSeatSubscription(ctx, q, tenantID)
	if err != nil {
		return nil, err
	}

	if !sub.Found || sub.Status != "active" || !sub.PeriodEnd.After(now) {
		return nil, seats.ErrSeatsUnavailable
	}

	unitPrice, listPrice, err := seatUnitPeriodPrice(ctx, q, seatType, sub.Interval)
	if err != nil {
		return nil, err
	}

	proration := seats.Prorate(unitPrice, sub.PeriodStart, sub.PeriodEnd, now)

	taxPercent, err := creditTaxPercent(ctx, q)
	if err != nil {
		return nil, err
	}

	price := proration.AmountIRR * int64(req.Quantity)
	tax := seatTaxIRR(price, taxPercent)

	return &platform.SeatQuote{
		Seat: string(seatType), Quantity: req.Quantity, BillingInterval: seats.NormalizeInterval(sub.Interval),
		PeriodStart: sub.PeriodStart, PeriodEnd: sub.PeriodEnd,
		RemainingDays: proration.RemainingDays, PeriodDays: proration.PeriodDays,
		ListPriceIRRPerCredit: listPrice, UnitPeriodPriceIRR: unitPrice,
		PriceIRR: price, TaxPercent: taxPercent.StringFixed(2), TaxIRR: tax, PayableIRR: price + tax,
	}, nil
}

// seatUnitPeriodPrice is one seat for a whole billing period at the current
// credit list price.
func seatUnitPeriodPrice(ctx context.Context, q interface {
	rowQuerier
	rowsQuerier
}, seatType seats.Type, interval string,
) (int64, int64, error) {
	types, err := loadSeatTypes(ctx, q)
	if err != nil {
		return 0, 0, err
	}

	row, ok := types[seatType]
	if !ok || !row.Active {
		return 0, 0, seats.ErrInvalidSeatType
	}

	listPrice, err := creditListPriceIRR(ctx, q)
	if err != nil {
		return 0, 0, err
	}

	unitPrice, err := seats.PeriodPriceIRR(row.offer(), listPrice, interval)

	return unitPrice, listPrice, err
}

func seatTaxIRR(priceIRR int64, taxPercent decimal.Decimal) int64 {
	if taxPercent.IsNegative() {
		return 0
	}

	return decimal.NewFromInt(priceIRR).Mul(taxPercent).Div(decimal.NewFromInt(100)).Round(0).IntPart()
}

// quoteSeatRenewal prices carrying the tenant's paid seats that lapse before
// the current period ends to that end. It fails with ErrRenewalNotDue when
// every paid seat already runs to the period end.
func (s *SeatRepository) quoteSeatRenewal(ctx context.Context, q interface {
	rowQuerier
	rowsQuerier
}, tenantID string, sub seatSubscription, now time.Time,
) (*platform.SeatRenewalQuote, error) {
	if !sub.Found || sub.Status != "active" {
		return nil, seats.ErrSeatsUnavailable
	}

	rows, err := q.Query(ctx, `
		SELECT block_id::text, seat_type, COALESCE(renew_quantity, quantity), paid_through
		FROM tenant_seat_blocks
		WHERE tenant_id = $1::uuid AND paid_through > $2 AND paid_through < $3
		  AND COALESCE(renew_quantity, quantity) > 0
		ORDER BY paid_through, seat_type, block_id
	`, tenantID, now, sub.PeriodEnd)
	if err != nil {
		return nil, err
	}

	lines, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (platform.SeatRenewalLine, error) {
		var line platform.SeatRenewalLine

		scanErr := row.Scan(&line.BlockID, &line.Seat, &line.Quantity, &line.PaidThrough)

		return line, scanErr
	})
	if err != nil {
		return nil, err
	}

	if len(lines) == 0 {
		return nil, seats.ErrRenewalNotDue
	}

	taxPercent, err := creditTaxPercent(ctx, q)
	if err != nil {
		return nil, err
	}

	quote := &platform.SeatRenewalQuote{Lines: lines, PeriodEnd: sub.PeriodEnd, TaxPercent: taxPercent.StringFixed(2)}

	for _, line := range lines {
		unitPrice, _, priceErr := seatUnitPeriodPrice(ctx, q, seats.Type(line.Seat), sub.Interval)
		if priceErr != nil {
			return nil, priceErr
		}

		// The renewal covers paid_through to the period end, at the share of
		// the period it spans.
		quote.PriceIRR += seats.Prorate(unitPrice, sub.PeriodStart, sub.PeriodEnd, line.PaidThrough).AmountIRR * int64(line.Quantity)
	}

	quote.TaxIRR = seatTaxIRR(quote.PriceIRR, taxPercent)
	quote.PayableIRR = quote.PriceIRR + quote.TaxIRR

	return quote, nil
}

// QuoteSeatRenewal prices renewing the tenant's lapsing paid seats.
func (s *SeatRepository) QuoteSeatRenewal(ctx context.Context, tenantID string, now time.Time) (*platform.SeatRenewalQuote, error) {
	sub, err := currentSeatSubscription(ctx, s.pool, tenantID)
	if err != nil {
		return nil, err
	}

	return s.quoteSeatRenewal(ctx, s.pool, tenantID, sub, now.UTC())
}

// UpdateSeatRenewal sets how many seats of a type the next renewal keeps.
// Seats not kept lapse when their block's paid period ends.
func (s *SeatRepository) UpdateSeatRenewal(ctx context.Context, tenantID string, req platform.UpdateSeatRenewalRequest, now time.Time) (*platform.SeatSummary, error) {
	seatType, err := paidSeatType(req.Seat)
	if err != nil {
		return nil, err
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	rows, err := tx.Query(ctx, `
		SELECT block_id::text, quantity FROM tenant_seat_blocks
		WHERE tenant_id = $1::uuid AND seat_type = $2 AND paid_through > $3
		ORDER BY paid_through DESC, created_at DESC
		FOR UPDATE
	`, tenantID, string(seatType), now.UTC())
	if err != nil {
		return nil, err
	}

	type block struct {
		id       string
		quantity int
	}

	blocks, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (block, error) {
		var item block

		return item, row.Scan(&item.id, &item.quantity)
	})
	if err != nil {
		return nil, err
	}

	held := 0
	for _, item := range blocks {
		held += item.quantity
	}

	if req.Quantity < 0 || req.Quantity > held {
		return nil, seats.ErrInvalidRenewal
	}

	// Keep the longest-paid blocks first.
	keep := req.Quantity
	for _, item := range blocks {
		renew := min(item.quantity, keep)
		keep -= renew

		if _, err := tx.Exec(ctx, `
			UPDATE tenant_seat_blocks SET renew_quantity = NULLIF($2, quantity), updated_at = NOW()
			WHERE block_id = $1::uuid
		`, item.id, renew); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return s.GetSeatSummary(ctx, tenantID, true, now)
}

// CreateSeatOrder records a seat purchase or renewal waiting for payment.
func (s *SeatRepository) CreateSeatOrder(ctx context.Context, tenantID string, req platform.CreateSeatOrderRequest, callbackURL, returnURL, actor string, now time.Time) (*platform.SeatOrder, error) {
	now = now.UTC()

	var billingUnit string
	if err := s.pool.QueryRow(ctx, `SELECT billing_unit FROM tenants WHERE tenant_id = $1::uuid`, tenantID).Scan(&billingUnit); err != nil || billingUnit != credits.BillingUnitCredit {
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}

		return nil, billingdomain.ErrCreditBillingNotEnabled
	}

	order := orderDraft{provider: req.Provider, callbackURL: callbackURL, returnURL: returnURL, actor: actor}

	switch strings.TrimSpace(req.Kind) {
	case "", platform.SeatOrderPurchase:
		quote, err := s.QuoteSeats(ctx, tenantID, platform.SeatQuoteRequest{Seat: req.Seat, Quantity: req.Quantity}, now)
		if err != nil {
			return nil, err
		}

		if err := s.validateOrderTargets(ctx, tenantID, req, seats.Type(quote.Seat)); err != nil {
			return nil, err
		}

		order.kind, order.seat, order.quantity = platform.SeatOrderPurchase, quote.Seat, quote.Quantity
		order.interval, order.periodStart, order.periodEnd = quote.BillingInterval, quote.PeriodStart, quote.PeriodEnd
		order.remainingDays, order.periodDays = quote.RemainingDays, quote.PeriodDays
		order.listPrice, order.unitPrice = quote.ListPriceIRRPerCredit, quote.UnitPeriodPriceIRR
		order.price, order.taxPercent, order.tax, order.payable = quote.PriceIRR, quote.TaxPercent, quote.TaxIRR, quote.PayableIRR
		order.assignUserID, order.requestID = strings.TrimSpace(req.AssignUserID), strings.TrimSpace(req.RequestID)
	case platform.SeatOrderRenewal:
		sub, err := currentSeatSubscription(ctx, s.pool, tenantID)
		if err != nil {
			return nil, err
		}

		renewal, err := s.quoteSeatRenewal(ctx, s.pool, tenantID, sub, now)
		if err != nil {
			return nil, err
		}

		listPrice, err := creditListPriceIRR(ctx, s.pool)
		if err != nil {
			return nil, err
		}

		order.kind, order.seat = platform.SeatOrderRenewal, renewal.Lines[0].Seat
		order.interval, order.periodStart, order.periodEnd = seats.NormalizeInterval(sub.Interval), sub.PeriodStart, sub.PeriodEnd
		order.periodDays = seats.Prorate(1, sub.PeriodStart, sub.PeriodEnd, sub.PeriodStart).PeriodDays
		order.remainingDays = order.periodDays
		order.listPrice = listPrice
		order.price, order.taxPercent, order.tax, order.payable = renewal.PriceIRR, renewal.TaxPercent, renewal.TaxIRR, renewal.PayableIRR

		for _, line := range renewal.Lines {
			order.quantity += line.Quantity
			order.renewBlockIDs = append(order.renewBlockIDs, line.BlockID)
		}
	default:
		return nil, seats.ErrInvalidSeatType
	}

	return order.insert(ctx, s, tenantID)
}

// validateOrderTargets checks the member and request an order will act on.
func (s *SeatRepository) validateOrderTargets(ctx context.Context, tenantID string, req platform.CreateSeatOrderRequest, seatType seats.Type) error {
	if userID := strings.TrimSpace(req.AssignUserID); userID != "" {
		var current string

		err := s.pool.QueryRow(ctx, `
			SELECT COALESCE(a.seat_type, 'unassigned')
			FROM users u
			LEFT JOIN tenant_seat_assignments a ON a.tenant_id = u.tenant_id AND a.user_id = u.user_id
			WHERE u.tenant_id = $1::uuid AND u.user_id::text = $2 AND COALESCE(u.active, true)
		`, tenantID, userID).Scan(&current)
		if errors.Is(err, pgx.ErrNoRows) {
			return seats.ErrMemberNotFound
		}

		if err != nil {
			return err
		}

		if _, err := seats.DecideChange(seats.Type(current), seatType, true); err != nil {
			return err
		}
	}

	if requestID := strings.TrimSpace(req.RequestID); requestID != "" {
		var status string

		err := s.pool.QueryRow(ctx, `
			SELECT status FROM seat_upgrade_requests WHERE tenant_id = $1::uuid AND request_id::text = $2
		`, tenantID, requestID).Scan(&status)
		if errors.Is(err, pgx.ErrNoRows) {
			return seats.ErrRequestNotFound
		}

		if err != nil {
			return err
		}

		if status != platform.SeatRequestPending {
			return seats.ErrRequestClosed
		}
	}

	return nil
}

// orderDraft is a priced seat order before it is stored.
type orderDraft struct {
	kind, seat, interval                      string
	quantity, remainingDays, periodDays       int
	periodStart, periodEnd                    time.Time
	listPrice, unitPrice, price, tax, payable int64
	taxPercent                                string
	assignUserID, requestID                   string
	renewBlockIDs                             []string
	provider, callbackURL, returnURL, actor   string
}

func (d orderDraft) insert(ctx context.Context, s *SeatRepository, tenantID string) (*platform.SeatOrder, error) {
	status := platform.CreditPackOrderPending

	if d.renewBlockIDs == nil {
		d.renewBlockIDs = []string{}
	}

	return scanSeatOrder(s.pool.QueryRow(ctx, `
		WITH inserted AS (
			INSERT INTO seat_orders (
				tenant_id, kind, seat_type, quantity, billing_interval, period_start, period_end,
				remaining_days, period_days, list_price_irr_per_credit, unit_period_price_irr,
				price_irr, tax_percent, tax_irr, payable_irr,
				assign_user_id, upgrade_request_id, renew_block_ids,
				provider, status, callback_url, return_url, created_by
			)
			VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13::numeric, $14, $15,
			        NULLIF($16, '')::uuid, NULLIF($17, '')::uuid, $18::uuid[],
			        $19, $20, NULLIF($21, ''), NULLIF($22, ''), NULLIF($23, ''))
			RETURNING *
		)
		SELECT `+seatOrderColumns+` FROM inserted o
	`, tenantID, d.kind, d.seat, d.quantity, d.interval, d.periodStart, d.periodEnd,
		d.remainingDays, d.periodDays, d.listPrice, d.unitPrice,
		d.price, d.taxPercent, d.tax, d.payable,
		d.assignUserID, d.requestID, d.renewBlockIDs,
		d.provider, status, d.callbackURL, d.returnURL, d.actor))
}

// AttachSeatOrderGatewayData stores the gateway authority of a pending order.
func (s *SeatRepository) AttachSeatOrderGatewayData(ctx context.Context, publicID, authority, paymentURL string, raw map[string]interface{}) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE seat_orders
		SET authority = $2, payment_url = $3, raw_payload = raw_payload || $4::jsonb, updated_at = NOW()
		WHERE public_id::text = $1 AND status = 'PENDING'
	`, publicID, authority, paymentURL, marshalJSONBestEffort(raw))
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return seats.ErrOrderNotPayable
	}

	return nil
}

// GetSeatOrderByAuthority returns the order a gateway callback refers to.
func (s *SeatRepository) GetSeatOrderByAuthority(ctx context.Context, authority string) (*platform.SeatOrder, error) {
	return scanSeatOrder(s.pool.QueryRow(ctx, `SELECT `+seatOrderColumns+` FROM seat_orders o WHERE o.authority = $1`, authority))
}

// ListSeatOrders returns the tenant's recent seat orders.
func (s *SeatRepository) ListSeatOrders(ctx context.Context, tenantID string, limit int) ([]platform.SeatOrder, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	rows, err := s.pool.Query(ctx, `
		SELECT `+seatOrderColumns+` FROM seat_orders o
		WHERE o.tenant_id = $1::uuid ORDER BY o.created_at DESC LIMIT $2
	`, tenantID, limit)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (platform.SeatOrder, error) {
		order, scanErr := scanSeatOrder(row)
		if scanErr != nil {
			return platform.SeatOrder{}, scanErr
		}

		return *order, nil
	})
}

// MarkSeatOrderFailedByAuthority records a failed or canceled payment. A
// completed order is never downgraded.
func (s *SeatRepository) MarkSeatOrderFailedByAuthority(ctx context.Context, authority, reason string, raw map[string]interface{}) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE seat_orders
		SET status = CASE WHEN status = 'PENDING' THEN 'FAILED' ELSE status END,
		    failure_reason = CASE WHEN status = 'PENDING' THEN $2 ELSE failure_reason END,
		    raw_payload = raw_payload || $3::jsonb, updated_at = NOW()
		WHERE authority = $1
	`, authority, reason, marshalJSONBestEffort(raw))
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return seats.ErrOrderNotFound
	}

	return nil
}

// CompleteSeatOrderByAuthority applies a paid seat order: a purchase adds a
// block of seats paid through the period end and gives one to the member it
// was bought for; a renewal moves the renewed blocks to the period end. A
// repeated callback for a completed order changes nothing and reports duplicate.
func (s *SeatRepository) CompleteSeatOrderByAuthority(ctx context.Context, authority string, verified billingdomain.PaymentVerification, now time.Time) (*platform.SeatOrder, bool, error) {
	now = now.UTC()

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, false, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	order, err := scanSeatOrder(tx.QueryRow(ctx, `SELECT `+seatOrderColumns+` FROM seat_orders o WHERE o.authority = $1 FOR UPDATE`, authority))
	if err != nil {
		return nil, false, err
	}

	if order.Status == platform.CreditPackOrderCompleted {
		return order, true, tx.Commit(ctx)
	}

	if order.Status != platform.CreditPackOrderPending {
		return nil, false, seats.ErrOrderNotPayable
	}

	blockID, err := applySeatOrderTx(ctx, tx, order, now)
	if err != nil {
		return nil, false, err
	}

	if _, updateErr := tx.Exec(ctx, `
		UPDATE seat_orders
		SET status = 'COMPLETED', block_id = NULLIF($2, '')::uuid, completed_at = $3, failure_reason = NULL,
		    ref_id = NULLIF($4, ''), gateway_txn_id = NULLIF($5, ''), card_mask = NULLIF($6, ''),
		    payer_phone = NULLIF($7, ''), raw_payload = raw_payload || $8::jsonb, updated_at = NOW()
		WHERE order_id = $1::uuid
	`, order.ID, blockID, now, verified.RefID, verified.GatewayTxnID, verified.CardMask, verified.PayerPhone,
		marshalJSONBestEffort(verified.Raw)); updateErr != nil {
		return nil, false, updateErr
	}

	if auditErr := insertFinancialAuditTx(ctx, tx, order.TenantID, "seat_order.completed", platform.SupportActor{ID: "system", Role: order.Provider},
		"seat_order", order.PublicID, "seat_order:"+order.ID, map[string]interface{}{
			"kind": order.Kind, "seat": order.Seat, "quantity": order.Quantity, "price_irr": order.PriceIRR,
			"payable_irr": order.PayableIRR, "block_id": blockID,
		}); auditErr != nil {
		return nil, false, auditErr
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		return nil, false, commitErr
	}

	s.afterSeatChange(ctx, order.TenantID, order.AssignUserID, now)

	completed, err := s.GetSeatOrderByAuthority(ctx, authority)

	return completed, false, err
}

// applySeatOrderTx changes the tenant's seat blocks for a paid order and
// returns the purchased block id.
func applySeatOrderTx(ctx context.Context, tx pgx.Tx, order *platform.SeatOrder, now time.Time) (string, error) {
	if order.Kind == platform.SeatOrderRenewal {
		return "", renewSeatBlocksTx(ctx, tx, order)
	}

	var blockID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO tenant_seat_blocks (tenant_id, seat_type, quantity, paid_through, source_order_id)
		VALUES ($1::uuid, $2, $3, $4, $5::uuid)
		RETURNING block_id::text
	`, order.TenantID, order.Seat, order.Quantity, order.PeriodEnd, order.ID).Scan(&blockID); err != nil {
		return "", err
	}

	if order.AssignUserID == "" {
		return blockID, nil
	}

	// The seats are paid for whatever happens next, so a member who can no
	// longer take one (they left, or already upgraded) only leaves it in stock.
	if _, err := tx.Exec(ctx, `SAVEPOINT seat_order_assign`); err != nil {
		return "", err
	}

	if _, err := setMemberSeatTx(ctx, tx, order.TenantID, order.AssignUserID, seats.Type(order.Seat), "seat_order", now); err != nil {
		if _, rollbackErr := tx.Exec(ctx, `ROLLBACK TO SAVEPOINT seat_order_assign`); rollbackErr != nil {
			return "", rollbackErr
		}

		return blockID, nil
	}

	if order.RequestID != "" {
		if err := closeSeatRequestTx(ctx, tx, order.TenantID, order.RequestID, order.AssignUserID, platform.SeatRequestApproved, "seat_order", "", now); err != nil &&
			!errors.Is(err, seats.ErrRequestClosed) && !errors.Is(err, seats.ErrRequestNotFound) {
			return "", err
		}
	}

	return blockID, nil
}

// renewSeatBlocksTx moves renewed blocks to the order's period end. A block
// renewing fewer seats than it holds is split: the kept seats move on, the
// rest lapse at the block's old date.
func renewSeatBlocksTx(ctx context.Context, tx pgx.Tx, order *platform.SeatOrder) error {
	for _, blockID := range order.RenewBlockIDs {
		var quantity, renew int

		err := tx.QueryRow(ctx, `
			SELECT quantity, COALESCE(renew_quantity, quantity) FROM tenant_seat_blocks
			WHERE block_id = $1::uuid AND tenant_id = $2::uuid FOR UPDATE
		`, blockID, order.TenantID).Scan(&quantity, &renew)
		if errors.Is(err, pgx.ErrNoRows) {
			continue
		}

		if err != nil {
			return err
		}

		if renew == 0 {
			continue
		}

		if renew < quantity {
			if _, err := tx.Exec(ctx, `
				INSERT INTO tenant_seat_blocks (tenant_id, seat_type, quantity, paid_through, source_order_id)
				SELECT tenant_id, seat_type, $2, paid_through, source_order_id FROM tenant_seat_blocks WHERE block_id = $1::uuid
			`, blockID, quantity-renew); err != nil {
				return err
			}
		}

		if _, err := tx.Exec(ctx, `
			UPDATE tenant_seat_blocks
			SET quantity = $2, renew_quantity = NULL, paid_through = GREATEST(paid_through, $3), updated_at = NOW()
			WHERE block_id = $1::uuid
		`, blockID, renew, order.PeriodEnd); err != nil {
			return err
		}
	}

	return nil
}
