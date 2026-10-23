package storage

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/seats"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const seatRequestColumns = `
	r.request_id::text, r.tenant_id::text, r.user_id::text, COALESCE(u.email, ''),
	TRIM(COALESCE(u.first_name, '') || ' ' || COALESCE(u.last_name, '')),
	r.current_seat_type, r.requested_seat_type, COALESCE(r.reason, ''), COALESCE(r.block_reason, ''), r.status,
	COALESCE(r.decided_by, ''), COALESCE(r.decision_note, ''), r.decided_at, r.created_at`

const seatRequestJoins = `
	FROM seat_upgrade_requests r
	LEFT JOIN users u ON u.user_id = r.user_id AND u.tenant_id = r.tenant_id`

func scanSeatRequest(row pgx.Row) (*platform.SeatUpgradeRequest, error) {
	var item platform.SeatUpgradeRequest

	err := row.Scan(&item.ID, &item.TenantID, &item.UserID, &item.Email, &item.Name,
		&item.CurrentSeat, &item.RequestedSeat, &item.Reason, &item.BlockReason, &item.Status,
		&item.DecidedBy, &item.DecisionNote, &item.DecidedAt, &item.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, seats.ErrRequestNotFound
	}

	if err != nil {
		return nil, err
	}

	return &item, nil
}

// CreateSeatUpgradeRequest records a member asking for a better seat.
func (s *SeatRepository) CreateSeatUpgradeRequest(ctx context.Context, tenantID, userID string, req platform.CreateSeatUpgradeRequest) (*platform.SeatUpgradeRequest, error) {
	requested, err := paidSeatType(req.Seat)
	if err != nil {
		return nil, err
	}

	var (
		enabled bool
		current string
	)

	err = s.pool.QueryRow(ctx, `
		SELECT COALESCE((SELECT seat_upgrade_requests_enabled FROM credit_tenant_settings WHERE tenant_id = $1::uuid), true),
		       COALESCE(a.seat_type, 'unassigned')
		FROM users u
		LEFT JOIN tenant_seat_assignments a ON a.tenant_id = u.tenant_id AND a.user_id = u.user_id
		WHERE u.tenant_id = $1::uuid AND u.user_id::text = $2 AND COALESCE(u.active, true)
	`, tenantID, strings.TrimSpace(userID)).Scan(&enabled, &current)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, seats.ErrMemberNotFound
	}

	if err != nil {
		return nil, err
	}

	if !enabled {
		return nil, seats.ErrRequestsDisabled
	}

	if requested.Rank() <= seats.Type(current).Rank() {
		return nil, seats.ErrRequestNotUpgrade
	}

	reason := strings.TrimSpace(req.Reason)
	if len([]rune(reason)) > 500 {
		reason = string([]rune(reason)[:500])
	}

	item, err := scanSeatRequest(s.pool.QueryRow(ctx, `
		WITH inserted AS (
			INSERT INTO seat_upgrade_requests (tenant_id, user_id, current_seat_type, requested_seat_type, reason, block_reason)
			VALUES ($1::uuid, $2::uuid, $3, $4, NULLIF($5, ''), NULLIF($6, ''))
			RETURNING *
		)
		SELECT `+seatRequestColumns+` FROM inserted r
		LEFT JOIN users u ON u.user_id = r.user_id AND u.tenant_id = r.tenant_id
	`, tenantID, strings.TrimSpace(userID), current, string(requested), reason, strings.TrimSpace(req.BlockReason)))

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return nil, seats.ErrRequestOpen
	}

	return item, err
}

// ListSeatUpgradeRequests returns the tenant's requests, newest first. An
// empty status lists every request.
func (s *SeatRepository) ListSeatUpgradeRequests(ctx context.Context, tenantID, status string, limit int) ([]platform.SeatUpgradeRequest, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	rows, err := s.pool.Query(ctx, `SELECT `+seatRequestColumns+seatRequestJoins+`
		WHERE r.tenant_id = $1::uuid AND ($2 = '' OR r.status = $2)
		ORDER BY CASE r.status WHEN 'pending' THEN 0 ELSE 1 END, r.created_at DESC
		LIMIT $3`, tenantID, strings.TrimSpace(status), limit)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (platform.SeatUpgradeRequest, error) {
		item, scanErr := scanSeatRequest(row)
		if scanErr != nil {
			return platform.SeatUpgradeRequest{}, scanErr
		}

		return *item, nil
	})
}

// GetSeatUpgradeRequest returns one request of the tenant.
func (s *SeatRepository) GetSeatUpgradeRequest(ctx context.Context, tenantID, requestID string) (*platform.SeatUpgradeRequest, error) {
	return scanSeatRequest(s.pool.QueryRow(ctx, `SELECT `+seatRequestColumns+seatRequestJoins+`
		WHERE r.tenant_id = $1::uuid AND r.request_id::text = $2`, tenantID, strings.TrimSpace(requestID)))
}

// DenySeatUpgradeRequest closes a pending request without changing the seat.
func (s *SeatRepository) DenySeatUpgradeRequest(ctx context.Context, tenantID, requestID, actor, note string, now time.Time) (*platform.SeatUpgradeRequest, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	if err := closeSeatRequestTx(ctx, tx, tenantID, requestID, "", platform.SeatRequestDenied, actor, note, now); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return s.GetSeatUpgradeRequest(ctx, tenantID, requestID)
}

// closeSeatRequestTx decides a pending request. userID, when set, must be the
// request's member.
func closeSeatRequestTx(ctx context.Context, tx pgx.Tx, tenantID, requestID, userID, status, actor, note string, now time.Time) error {
	var (
		owner, current string
	)

	err := tx.QueryRow(ctx, `
		SELECT user_id::text, status FROM seat_upgrade_requests
		WHERE tenant_id = $1::uuid AND request_id::text = $2 FOR UPDATE
	`, tenantID, strings.TrimSpace(requestID)).Scan(&owner, &current)
	if errors.Is(err, pgx.ErrNoRows) {
		return seats.ErrRequestNotFound
	}

	if err != nil {
		return err
	}

	if userID != "" && !strings.EqualFold(owner, strings.TrimSpace(userID)) {
		return seats.ErrRequestNotFound
	}

	if current != platform.SeatRequestPending {
		return seats.ErrRequestClosed
	}

	_, err = tx.Exec(ctx, `
		UPDATE seat_upgrade_requests
		SET status = $3, decided_by = NULLIF($4, ''), decision_note = NULLIF($5, ''), decided_at = $6
		WHERE tenant_id = $1::uuid AND request_id::text = $2
	`, tenantID, strings.TrimSpace(requestID), status, actor, strings.TrimSpace(note), now)

	return err
}

// GetSeatStatus reports a member's own seat for the gateway's /v1/budget/me
// and GodarChat. The block reason is computed from the ledger, so it can lag
// in-flight requests by one settlement pass.
func (s *SeatRepository) GetSeatStatus(ctx context.Context, tenantID, userID string, now time.Time) (*platform.SeatStatus, error) {
	now = now.UTC()
	windowID, _ := credits.PoolWindow(now)

	member, err := s.GetSeatMember(ctx, tenantID, userID, now)
	if err != nil {
		return nil, err
	}

	sub, err := currentSeatSubscription(ctx, s.pool, tenantID)
	if err != nil {
		return nil, err
	}

	var (
		requestsEnabled bool
		pending         string
	)

	if err := s.pool.QueryRow(ctx, `
		SELECT COALESCE((SELECT seat_upgrade_requests_enabled FROM credit_tenant_settings WHERE tenant_id = $1::uuid), true),
		       COALESCE((SELECT requested_seat_type FROM seat_upgrade_requests
		                 WHERE tenant_id = $1::uuid AND user_id::text = $2 AND status = 'pending' LIMIT 1), '')
	`, tenantID, strings.TrimSpace(userID)).Scan(&requestsEnabled, &pending); err != nil {
		return nil, err
	}

	seatType := seats.Type(member.Seat)
	enforced := seats.Enforced(sub.Status)

	status := &platform.SeatStatus{
		Enforced: enforced, Type: member.Seat,
		MonthlyCredits:     credits.Format(credits.Millicredits(seatType.MonthlyGrantMC())),
		RemainingCredits:   member.PersonalRemainingCredits,
		PoolAccess:         seats.PoolAccess(seatType, enforced),
		PoolLimitCredits:   member.PoolLimitCredits,
		PoolUsedCredits:    member.PoolUsedCredits,
		PendingRequest:     pending,
		UpgradeRequestable: requestsEnabled && pending == "" && seatType != seats.Max,
	}

	if offer, ok := seatType.Offer(); ok {
		status.Lifetime = offer.Lifetime
	}

	if seatType.Paid() && sub.running(now) {
		if slice, ok := credits.CurrentPlanSlice(sub.PeriodStart, sub.PeriodEnd, now); ok {
			renews := slice.End
			status.RenewsAt = &renews
		}
	}

	if enforced {
		reason, reasonErr := s.seatBlockReason(ctx, tenantID, userID, windowID, seatType, member)
		if reasonErr != nil {
			return nil, reasonErr
		}

		status.BlockReason = string(reason)
	}

	return status, nil
}

func (s *SeatRepository) seatBlockReason(ctx context.Context, tenantID, userID, windowID string, seatType seats.Type, member *platform.SeatMember) (seats.BlockReason, error) {
	if seatType == seats.Unassigned {
		return seats.BlockSeatUnassigned, nil
	}

	remainingMC, err := credits.ParseCreditAmountMC(member.PersonalRemainingCredits)
	if err == nil && remainingMC > 0 {
		return "", nil
	}

	if !seats.PoolAccess(seatType, true) {
		return seats.BlockSeatCreditsExhausted, nil
	}

	if member.PoolLimitCredits != nil {
		limitMC, limitErr := credits.ParseCreditAmountMC(*member.PoolLimitCredits)
		usedMC, usedErr := s.SeatPoolUsedMC(ctx, tenantID, userID, windowID)

		if limitErr == nil && usedErr == nil && usedMC >= limitMC {
			return seats.BlockPersonalLimitReached, nil
		}
	}

	poolMC, err := s.CreditPoolAllowanceMCForSeats(ctx, tenantID, windowID)
	if err != nil {
		return "", err
	}

	if poolMC <= 0 {
		return seats.BlockPoolExhausted, nil
	}

	return "", nil
}

// CreditPoolAllowanceMCForSeats is the tenant pool's remaining allowance as
// the ledger sees it, for status reporting.
func (s *SeatRepository) CreditPoolAllowanceMCForSeats(ctx context.Context, tenantID, windowID string) (int64, error) {
	pool, err := s.creditPoolAllowance(ctx, s.pool, tenantID, windowID)
	if err != nil {
		return 0, err
	}

	return pool.AllowanceMC() - pool.SettledMC, nil
}
