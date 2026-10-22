package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/seats"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

// SeatRepository persists member seats: the paid seats a tenant holds, who has
// which seat, upgrade requests, and seat orders. Seat credits themselves are
// personal lots in the credit ledger.
type SeatRepository struct{ *walletState }

// seatsInvalidateReason is the runtime-profile invalidation reason of a seat change.
const seatsInvalidateReason = "seats_changed"

// seatSubscription is a tenant's current subscription, if any: the free plan
// or the seat subscription that paid seats follow.
type seatSubscription struct {
	ID          string
	PlanSlug    string
	Status      string
	Interval    string
	PeriodStart time.Time
	PeriodEnd   time.Time
	Found       bool
}

// running reports whether the tenant has a seat subscription whose period is
// still going at now. Paid seats bought then follow its period.
func (sub seatSubscription) running(now time.Time) bool {
	return sub.Found && sub.PlanSlug == seats.SeatsPlanSlug && sub.Status == "active" && sub.PeriodEnd.After(now)
}

func currentSeatSubscription(ctx context.Context, q rowQuerier, tenantID string) (seatSubscription, error) {
	var sub seatSubscription

	err := q.QueryRow(ctx, `
		SELECT ts.subscription_id::text, p.slug, ts.status, COALESCE(ts.billing_interval, 'monthly'),
		       ts.current_period_start, ts.current_period_end
		FROM tenant_subscriptions ts
		JOIN commercial_plans p ON p.plan_id = ts.plan_id
		WHERE ts.tenant_id = $1::uuid AND ts.status IN ('trialing', 'active', 'past_due')
		ORDER BY ts.updated_at DESC
		LIMIT 1
	`, tenantID).Scan(&sub.ID, &sub.PlanSlug, &sub.Status, &sub.Interval, &sub.PeriodStart, &sub.PeriodEnd)
	if errors.Is(err, pgx.ErrNoRows) {
		return seatSubscription{}, nil
	}

	if err != nil {
		return seatSubscription{}, err
	}

	sub.Found = true
	sub.PeriodStart, sub.PeriodEnd = sub.PeriodStart.UTC(), sub.PeriodEnd.UTC()

	return sub, nil
}

// seatTypeRow is one row of seat_types.
type seatTypeRow struct {
	Type           seats.Type
	MonthlyGrantMC int64
	Lifetime       bool
	Discount       decimal.Decimal
	YearlyDiscount decimal.Decimal
	PoolAccess     bool
	Active         bool
}

func (r seatTypeRow) offer() seats.Offer {
	return seats.Offer{
		Type: r.Type, MonthlyCredits: r.MonthlyGrantMC / credits.MillicreditsPerCredit, Lifetime: r.Lifetime,
		DiscountPercent: r.Discount.IntPart(), YearlyDiscountPercent: r.YearlyDiscount.IntPart(), PoolAccess: r.PoolAccess,
	}
}

type rowsQuerier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

func loadSeatTypes(ctx context.Context, q rowsQuerier) (map[seats.Type]seatTypeRow, error) {
	rows, err := q.Query(ctx, `
		SELECT seat_type, monthly_grant_mc, lifetime, discount_percent::text, yearly_discount_percent::text, pool_access, active
		FROM seat_types ORDER BY rank
	`)
	if err != nil {
		return nil, err
	}

	items, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (seatTypeRow, error) {
		var (
			item             seatTypeRow
			seatType         string
			discount, yearly string
		)

		if scanErr := row.Scan(&seatType, &item.MonthlyGrantMC, &item.Lifetime, &discount, &yearly, &item.PoolAccess, &item.Active); scanErr != nil {
			return item, scanErr
		}

		item.Type = seats.Type(seatType)

		parsed, parseErr := decimal.NewFromString(discount)
		if parseErr != nil {
			return item, parseErr
		}

		item.Discount = parsed
		item.YearlyDiscount, parseErr = decimal.NewFromString(yearly)

		return item, parseErr
	})
	if err != nil {
		return nil, err
	}

	out := make(map[seats.Type]seatTypeRow, len(items))
	for _, item := range items {
		out[item.Type] = item
	}

	return out, nil
}

// seatStock is the paid seats of one type a tenant holds and has assigned.
type seatStock struct {
	Purchased int
	Assigned  int
	Renewing  int
}

func (s seatStock) available() int {
	return max(s.Purchased-s.Assigned, 0)
}

// loadSeatStock counts the tenant's active paid seats per type and how many
// active members hold one.
func loadSeatStock(ctx context.Context, q rowsQuerier, tenantID string, now time.Time) (map[seats.Type]seatStock, error) {
	rows, err := q.Query(ctx, `
		SELECT seat_type,
		       COALESCE(SUM(quantity), 0)::int,
		       COALESCE(SUM(COALESCE(renew_quantity, quantity)), 0)::int,
		       0
		FROM tenant_seat_blocks
		WHERE tenant_id = $1::uuid AND paid_through > $2
		GROUP BY seat_type
		UNION ALL
		SELECT a.seat_type, 0, 0, COUNT(1)::int
		FROM tenant_seat_assignments a
		JOIN users u ON u.user_id = a.user_id AND u.tenant_id = a.tenant_id AND COALESCE(u.active, true)
		WHERE a.tenant_id = $1::uuid AND a.seat_type IN ('pro', 'max')
		GROUP BY a.seat_type
	`, tenantID, now)
	if err != nil {
		return nil, err
	}

	out := map[seats.Type]seatStock{seats.Pro: {}, seats.Max: {}}

	_, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (struct{}, error) {
		var (
			seatType                      string
			purchased, renewing, assigned int
		)

		if scanErr := row.Scan(&seatType, &purchased, &renewing, &assigned); scanErr != nil {
			return struct{}{}, scanErr
		}

		stock := out[seats.Type(seatType)]
		stock.Purchased += purchased
		stock.Renewing += renewing
		stock.Assigned += assigned
		out[seats.Type(seatType)] = stock

		return struct{}{}, nil
	})

	return out, err
}

// GetSeatProfile returns the seat state compiled into the tenant's runtime
// profile: who holds which seat, and when seat limits start to apply.
func (s *SeatRepository) GetSeatProfile(ctx context.Context, tenantID string) (*platform.SeatProfile, error) {
	sub, err := currentSeatSubscription(ctx, s.pool, tenantID)
	if err != nil {
		return nil, err
	}

	profile := &platform.SeatProfile{SubscriptionStatus: sub.Status, Members: map[string]string{}, PoolLimitMC: map[string]int64{}}

	var proLimit, maxLimit *int64

	err = s.pool.QueryRow(ctx, `
		SELECT seat_pool_limit_pro_mc, seat_pool_limit_max_mc
		FROM credit_tenant_settings WHERE tenant_id = $1::uuid
	`, tenantID).Scan(&proLimit, &maxLimit)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	if proLimit != nil {
		profile.PoolLimitMC[platform.SeatPro] = *proLimit
	}

	if maxLimit != nil {
		profile.PoolLimitMC[platform.SeatMax] = *maxLimit
	}

	rows, err := s.pool.Query(ctx, `
		SELECT a.user_id::text, a.seat_type
		FROM tenant_seat_assignments a
		JOIN users u ON u.user_id = a.user_id AND u.tenant_id = a.tenant_id AND COALESCE(u.active, true)
		WHERE a.tenant_id = $1::uuid
	`, tenantID)
	if err != nil {
		return nil, err
	}

	_, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (struct{}, error) {
		var userID, seatType string

		scanErr := row.Scan(&userID, &seatType)
		profile.Members[userID] = seatType

		return struct{}{}, scanErr
	})

	return profile, err
}

// GetSeatSummary returns the tenant's seat overview.
func (s *SeatRepository) GetSeatSummary(ctx context.Context, tenantID string, now time.Time) (*platform.SeatSummary, error) {
	now = now.UTC()

	sub, err := currentSeatSubscription(ctx, s.pool, tenantID)
	if err != nil {
		return nil, err
	}

	usage, err := tenantSeatUsageTx(ctx, s.pool, tenantID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, seats.ErrMemberNotFound
	}

	if err != nil {
		return nil, err
	}

	summary := &platform.SeatSummary{
		SubscriptionStatus: sub.Status, SubscriptionKind: seats.FreePlanSlug, PendingInvitations: usage.Pending,
	}

	if sub.running(now) {
		summary.SubscriptionKind = seats.SeatsPlanSlug
		summary.BillingInterval = seats.NormalizeInterval(sub.Interval)
		summary.PeriodStart, summary.PeriodEnd = &sub.PeriodStart, &sub.PeriodEnd
	}

	if usage.FreeLimited {
		limit := usage.FreeLimit
		summary.FreeLimit = &limit
	}

	if scanErr := s.pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(1) FROM tenant_seat_assignments a
			 JOIN users u ON u.user_id = a.user_id AND u.tenant_id = a.tenant_id AND COALESCE(u.active, true)
			 WHERE a.tenant_id = $1::uuid AND a.seat_type = 'free')::int,
			(SELECT COUNT(1) FROM users u
			 WHERE u.tenant_id = $1::uuid AND COALESCE(u.active, true)
			   AND NOT EXISTS (SELECT 1 FROM tenant_seat_assignments a WHERE a.tenant_id = u.tenant_id AND a.user_id = u.user_id))::int,
			(SELECT COUNT(1) FROM seat_upgrade_requests WHERE tenant_id = $1::uuid AND status = 'pending')::int
	`, tenantID).Scan(&summary.FreeUsed, &summary.Unassigned, &summary.PendingRequests); scanErr != nil {
		return nil, scanErr
	}

	summary.Enforced = seats.Enforced(sub.Status)

	stock, err := loadSeatStock(ctx, s.pool, tenantID, now)
	if err != nil {
		return nil, err
	}

	for _, seatType := range seats.PaidTypes {
		item := stock[seatType]
		summary.Paid = append(summary.Paid, platform.SeatTypeCount{
			Type: string(seatType), Purchased: item.Purchased, Assigned: item.Assigned,
			Available: item.available(), Renewing: item.Renewing,
		})
	}

	summary.Offers, err = s.seatOffers(ctx)
	if err != nil {
		return nil, err
	}

	taxPercent, err := creditTaxPercent(ctx, s.pool)
	if err != nil {
		return nil, err
	}

	summary.TaxPercent = taxPercent.StringFixed(2)

	if sub.running(now) {
		renewal, renewalErr := s.quoteSeatRenewal(ctx, s.pool, tenantID, sub, now)
		if renewalErr == nil {
			summary.RenewalDue = renewal
		} else if !errors.Is(renewalErr, seats.ErrRenewalNotDue) {
			return nil, renewalErr
		}
	}

	return summary, nil
}

// seatOffers prices every seat type at the current credit list price.
func (s *SeatRepository) seatOffers(ctx context.Context) ([]platform.SeatTypeOffer, error) {
	types, err := loadSeatTypes(ctx, s.pool)
	if err != nil {
		return nil, err
	}

	listPrice, err := creditListPriceIRR(ctx, s.pool)
	if err != nil {
		return nil, err
	}

	out := make([]platform.SeatTypeOffer, 0, len(types))

	for _, seatType := range []seats.Type{seats.Free, seats.Pro, seats.Max} {
		row, ok := types[seatType]
		if !ok {
			continue
		}

		offer := row.offer()

		monthly, monthlyErr := seats.PeriodPriceIRR(offer, listPrice, "monthly")
		if monthlyErr != nil {
			return nil, monthlyErr
		}

		yearly, yearlyErr := seats.PeriodPriceIRR(offer, listPrice, "yearly")
		if yearlyErr != nil {
			return nil, yearlyErr
		}

		out = append(out, platform.SeatTypeOffer{
			Type: string(seatType), MonthlyCredits: offer.MonthlyCredits, Lifetime: row.Lifetime, PoolAccess: row.PoolAccess,
			DiscountPercent: row.Discount.StringFixed(2), YearlyDiscountPercent: row.YearlyDiscount.StringFixed(2), MonthlyPriceIRR: monthly, YearlyPriceIRR: yearly, Active: row.Active,
		})
	}

	return out, nil
}

// MemberSeatType resolves a member's seat for chat admission when the runtime
// profile does not know them yet.
func (s *SeatRepository) MemberSeatType(ctx context.Context, tenantID, userID string) (string, error) {
	var seatType string

	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE((SELECT a.seat_type FROM tenant_seat_assignments a
		                 WHERE a.tenant_id = $1::uuid AND a.user_id::text = $2), 'unassigned')
	`, strings.TrimSpace(tenantID), strings.TrimSpace(userID)).Scan(&seatType)

	return seatType, err
}

// ListSeatTypes returns every seat type with its current price (support admin).
func (s *SeatRepository) ListSeatTypes(ctx context.Context) ([]platform.SeatTypeOffer, error) {
	return s.seatOffers(ctx)
}

// UpdateSeatType changes a seat type's quota, discount, or availability.
// Discounts are whole percentages so seat prices stay whole rials.
func (s *SeatRepository) UpdateSeatType(ctx context.Context, seatType string, req platform.SeatTypeUpdate, actor string) ([]platform.SeatTypeOffer, error) {
	parsed, err := seats.ParseType(seatType)
	if err != nil || parsed == seats.Unassigned {
		return nil, seats.ErrInvalidSeatType
	}

	var grantMC *int64

	if req.MonthlyCredits != nil {
		if *req.MonthlyCredits <= 0 {
			return nil, fmt.Errorf("%w: monthly credits must be positive", seats.ErrInvalidSeatType)
		}

		value := *req.MonthlyCredits * credits.MillicreditsPerCredit
		grantMC = &value
	}

	discount, err := wholePercent(req.DiscountPercent)
	if err != nil {
		return nil, err
	}

	yearly, err := wholePercent(req.YearlyDiscountPercent)
	if err != nil {
		return nil, err
	}

	tag, err := s.pool.Exec(ctx, `
		UPDATE seat_types
		SET monthly_grant_mc = COALESCE($2, monthly_grant_mc),
		    discount_percent = COALESCE($3::numeric, discount_percent),
		    yearly_discount_percent = COALESCE($6::numeric, yearly_discount_percent),
		    active = COALESCE($4, active),
		    updated_by = NULLIF($5, ''), updated_at = NOW()
		WHERE seat_type = $1
	`, string(parsed), grantMC, discount, req.Active, actor, yearly)
	if err != nil {
		return nil, err
	}

	if tag.RowsAffected() == 0 {
		return nil, seats.ErrInvalidSeatType
	}

	return s.seatOffers(ctx)
}

// wholePercent validates an optional discount: a whole percentage below 100,
// so seat prices stay whole rials.
func wholePercent(raw *string) (*string, error) {
	if raw == nil {
		return nil, nil
	}

	value, err := decimal.NewFromString(strings.TrimSpace(*raw))
	if err != nil || value.IsNegative() || value.GreaterThanOrEqual(decimal.NewFromInt(100)) || !value.Equal(value.Truncate(0)) {
		return nil, fmt.Errorf("%w: discount must be a whole percentage below 100", seats.ErrInvalidSeatType)
	}

	text := value.String()

	return &text, nil
}

// seatMemberColumns selects one member's seat and credit usage. $1 is the
// tenant, $2 the time "now", and $3 the calendar-month window id.
const seatMemberColumns = `
	u.user_id::text, COALESCE(u.email, ''), TRIM(COALESCE(u.first_name, '') || ' ' || COALESCE(u.last_name, '')),
	COALESCE(u.role, 'member'), COALESCE(u.active, true),
	COALESCE(a.seat_type, 'unassigned'), COALESCE(a.pending_seat_type, ''), a.pending_effective_at, a.assigned_at,
	EXISTS (SELECT 1 FROM tenant_free_seat_grants g WHERE g.tenant_id = u.tenant_id AND g.user_id = u.user_id),
	COALESCE((SELECT SUM(l.amount_mc) FROM credit_lots l JOIN credit_lot_balances b ON b.lot_id = l.lot_id
	          WHERE l.tenant_id = u.tenant_id AND l.owner_user_id = u.user_id
	            AND (b.expires_at IS NULL OR b.expires_at > $2)), 0)::bigint,
	COALESCE((SELECT SUM(b.remaining_mc) FROM credit_lot_balances b
	          WHERE b.tenant_id = u.tenant_id AND b.owner_user_id = u.user_id AND b.remaining_mc > 0
	            AND (b.expires_at IS NULL OR b.expires_at > $2)), 0)::bigint,
	COALESCE(w.pool_settled_mc, 0)`

const seatMemberJoins = `
	FROM users u
	LEFT JOIN tenant_seat_assignments a ON a.tenant_id = u.tenant_id AND a.user_id = u.user_id
	LEFT JOIN credit_user_window_settlements w ON w.tenant_id = u.tenant_id AND w.user_id = u.user_id AND w.window_id = $3`

func scanSeatMember(row pgx.Row, poolLimits map[string]int64) (platform.SeatMember, error) {
	var (
		item                           platform.SeatMember
		grantedMC, remainingMC, poolMC int64
	)

	if err := row.Scan(&item.UserID, &item.Email, &item.Name, &item.Role, &item.Active,
		&item.Seat, &item.PendingSeat, &item.PendingEffectiveAt, &item.AssignedAt, &item.FreeSeatUsed,
		&grantedMC, &remainingMC, &poolMC); err != nil {
		return item, err
	}

	item.PersonalGrantedCredits = credits.Format(credits.Millicredits(grantedMC))
	item.PersonalRemainingCredits = credits.Format(credits.Millicredits(remainingMC))
	item.PoolUsedCredits = credits.Format(credits.Millicredits(poolMC))

	if limit, ok := poolLimits[item.Seat]; ok {
		item.PoolLimitCredits = formatOptionalCredits(&limit)
	}

	return item, nil
}

func (s *SeatRepository) seatPoolLimits(ctx context.Context, tenantID string) (map[string]int64, error) {
	var proLimit, maxLimit *int64

	err := s.pool.QueryRow(ctx, `
		SELECT seat_pool_limit_pro_mc, seat_pool_limit_max_mc FROM credit_tenant_settings WHERE tenant_id = $1::uuid
	`, tenantID).Scan(&proLimit, &maxLimit)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	out := map[string]int64{}
	if proLimit != nil {
		out[platform.SeatPro] = *proLimit
	}

	if maxLimit != nil {
		out[platform.SeatMax] = *maxLimit
	}

	return out, nil
}

// ListSeatMembers lists the tenant's active members with their seats and usage.
func (s *SeatRepository) ListSeatMembers(ctx context.Context, tenantID, query, seatFilter string, limit, offset int, now time.Time) (*platform.SeatMemberList, error) {
	now = now.UTC()
	windowID, _ := credits.PoolWindow(now)
	query = strings.ToLower(strings.TrimSpace(query))
	like := "%" + query + "%"

	if limit <= 0 || limit > 500 {
		limit = 100
	}

	poolLimits, err := s.seatPoolLimits(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	filter := `
		WHERE u.tenant_id = $1::uuid AND COALESCE(u.active, true)
		  AND ($4 = '' OR LOWER(COALESCE(u.email, '')) LIKE $5
		       OR LOWER(COALESCE(u.first_name, '') || ' ' || COALESCE(u.last_name, '')) LIKE $5)
		  AND ($6 = '' OR COALESCE(a.seat_type, 'unassigned') = $6)`

	list := &platform.SeatMemberList{Items: []platform.SeatMember{}}

	// $2 (now) only appears in the select columns, so the count query pins its type.
	if countErr := s.pool.QueryRow(ctx, `SELECT COUNT(1) `+seatMemberJoins+filter+` AND $2::timestamptz IS NOT NULL`,
		tenantID, now, windowID, query, like, strings.TrimSpace(seatFilter)).Scan(&list.Total); countErr != nil {
		return nil, countErr
	}

	rows, err := s.pool.Query(ctx, `SELECT `+seatMemberColumns+seatMemberJoins+filter+`
		ORDER BY CASE COALESCE(a.seat_type, 'unassigned') WHEN 'max' THEN 0 WHEN 'pro' THEN 1 WHEN 'free' THEN 2 ELSE 3 END,
		         LOWER(COALESCE(u.email, '')), u.user_id
		LIMIT $7 OFFSET $8`,
		tenantID, now, windowID, query, like, strings.TrimSpace(seatFilter), limit, max(offset, 0))
	if err != nil {
		return nil, err
	}

	list.Items, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (platform.SeatMember, error) {
		return scanSeatMember(row, poolLimits)
	})

	return list, err
}

// GetSeatMember returns one member with their seat and usage.
func (s *SeatRepository) GetSeatMember(ctx context.Context, tenantID, userID string, now time.Time) (*platform.SeatMember, error) {
	now = now.UTC()
	windowID, _ := credits.PoolWindow(now)

	poolLimits, err := s.seatPoolLimits(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	item, err := scanSeatMember(s.pool.QueryRow(ctx, `SELECT `+seatMemberColumns+seatMemberJoins+`
		WHERE u.tenant_id = $1::uuid AND u.user_id::text = $4`, tenantID, now, windowID, strings.TrimSpace(userID)), poolLimits)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, seats.ErrMemberNotFound
	}

	if err != nil {
		return nil, err
	}

	return &item, nil
}

// memberSeatState is a member's seat inside a locked seat change.
type memberSeatState struct {
	Active   bool
	Seat     seats.Type
	FreeUsed bool
}

func lockMemberSeatTx(ctx context.Context, tx pgx.Tx, tenantID, userID string) (memberSeatState, error) {
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext('users_cap:' || $1))`, tenantID); err != nil {
		return memberSeatState{}, err
	}

	var (
		state    memberSeatState
		seatType string
	)

	err := tx.QueryRow(ctx, `
		SELECT COALESCE(u.active, true), COALESCE(a.seat_type, 'unassigned'),
		       EXISTS (SELECT 1 FROM tenant_free_seat_grants g WHERE g.tenant_id = u.tenant_id AND g.user_id = u.user_id)
		FROM users u
		LEFT JOIN tenant_seat_assignments a ON a.tenant_id = u.tenant_id AND a.user_id = u.user_id
		WHERE u.tenant_id = $1::uuid AND u.user_id::text = $2
		FOR UPDATE OF u
	`, tenantID, userID).Scan(&state.Active, &seatType, &state.FreeUsed)
	if errors.Is(err, pgx.ErrNoRows) {
		return memberSeatState{}, seats.ErrMemberNotFound
	}

	state.Seat = seats.Type(seatType)

	return state, err
}

// SetMemberSeat changes a member's seat. Upgrades and the free seat apply at
// once; paid downgrades wait for the end of the billing period. A paid upgrade
// takes a seat from the tenant's stock and fails with ErrSeatPurchaseRequired
// when none is free. requestID, when set, marks that upgrade request approved.
func (s *SeatRepository) SetMemberSeat(ctx context.Context, tenantID, userID string, target seats.Type, actor, requestID string, now time.Time) (*platform.SetMemberSeatResult, error) {
	now = now.UTC()
	userID = strings.TrimSpace(userID)

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	change, err := setMemberSeatTx(ctx, tx, tenantID, userID, target, actor, now)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(requestID) != "" {
		if closeErr := closeSeatRequestTx(ctx, tx, tenantID, requestID, userID, platform.SeatRequestApproved, actor, "", now); closeErr != nil {
			return nil, closeErr
		}
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		return nil, commitErr
	}

	s.afterSeatChange(ctx, tenantID, userID, now)

	member, err := s.GetSeatMember(ctx, tenantID, userID, now)
	if err != nil {
		return nil, err
	}

	return &platform.SetMemberSeatResult{Member: *member, Timing: string(change.Timing)}, nil
}

// setMemberSeatTx applies one seat change inside tx.
func setMemberSeatTx(ctx context.Context, tx pgx.Tx, tenantID, userID string, target seats.Type, actor string, now time.Time) (seats.Change, error) {
	state, err := lockMemberSeatTx(ctx, tx, tenantID, userID)
	if err != nil {
		return seats.Change{}, err
	}

	if !state.Active {
		return seats.Change{}, seats.ErrMemberNotFound
	}

	change, err := seats.DecideChange(state.Seat, target, state.FreeUsed)
	if err != nil {
		return seats.Change{}, err
	}

	sub, err := currentSeatSubscription(ctx, tx, tenantID)
	if err != nil {
		return seats.Change{}, err
	}

	if availableErr := ensureSeatAvailableTx(ctx, tx, tenantID, change, now); availableErr != nil {
		return seats.Change{}, availableErr
	}

	// Without a running seat subscription there is nothing to wait for.
	if change.Timing == seats.AtPeriodEnd && !sub.running(now) {
		change.Timing = seats.Immediate
	}

	if change.Timing == seats.AtPeriodEnd {
		_, err = tx.Exec(ctx, `
			UPDATE tenant_seat_assignments
			SET pending_seat_type = $3, pending_effective_at = $4, updated_at = NOW()
			WHERE tenant_id = $1::uuid AND user_id = $2::uuid
		`, tenantID, userID, string(target), sub.PeriodEnd)

		return change, err
	}

	return change, applySeatTx(ctx, tx, tenantID, userID, target, actor)
}

// ensureSeatAvailableTx checks the tenant has room for the change's new seat.
// A paid seat needs a free seat of its type in stock; without one the admin
// buys it, which starts the seat subscription when there is none yet.
func ensureSeatAvailableTx(ctx context.Context, tx pgx.Tx, tenantID string, change seats.Change, now time.Time) error {
	switch {
	case change.To == seats.Free:
		usage, err := tenantSeatUsageTx(ctx, tx, tenantID)
		if err != nil {
			return err
		}

		var freeHeld int
		if err := tx.QueryRow(ctx, `
			SELECT COUNT(1) FROM tenant_seat_assignments a
			JOIN users u ON u.user_id = a.user_id AND u.tenant_id = a.tenant_id AND COALESCE(u.active, true)
			WHERE a.tenant_id = $1::uuid AND a.seat_type = 'free'
		`, tenantID).Scan(&freeHeld); err != nil {
			return err
		}

		if usage.FreeLimited && freeHeld >= usage.FreeLimit {
			return seats.ErrFreeSeatsFull
		}
	case change.NeedsPaidSeat:
		stock, err := loadSeatStock(ctx, tx, tenantID, now)
		if err != nil {
			return err
		}

		if stock[change.To].available() <= 0 {
			return seats.ErrSeatPurchaseRequired
		}
	}

	return nil
}

// applySeatTx gives a member a seat now, clearing any scheduled change.
func applySeatTx(ctx context.Context, tx pgx.Tx, tenantID, userID string, target seats.Type, actor string) error {
	if target == seats.Unassigned {
		_, err := tx.Exec(ctx, `DELETE FROM tenant_seat_assignments WHERE tenant_id = $1::uuid AND user_id = $2::uuid`, tenantID, userID)

		return err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO tenant_seat_assignments (tenant_id, user_id, seat_type, assigned_by)
		VALUES ($1::uuid, $2::uuid, $3, NULLIF($4, ''))
		ON CONFLICT (tenant_id, user_id) DO UPDATE
		SET seat_type = EXCLUDED.seat_type, pending_seat_type = NULL, pending_effective_at = NULL,
		    assigned_by = EXCLUDED.assigned_by, assigned_at = NOW(), updated_at = NOW()
	`, tenantID, userID, string(target), actor); err != nil {
		return err
	}

	if target != seats.Free {
		return nil
	}

	_, err := tx.Exec(ctx, `
		INSERT INTO tenant_free_seat_grants (tenant_id, user_id) VALUES ($1::uuid, $2::uuid)
		ON CONFLICT DO NOTHING
	`, tenantID, userID)

	return err
}

// CancelPendingSeatChange keeps a member's current seat.
func (s *SeatRepository) CancelPendingSeatChange(ctx context.Context, tenantID, userID string, now time.Time) (*platform.SeatMember, error) {
	tag, err := s.pool.Exec(ctx, `
		UPDATE tenant_seat_assignments
		SET pending_seat_type = NULL, pending_effective_at = NULL, updated_at = NOW()
		WHERE tenant_id = $1::uuid AND user_id::text = $2 AND pending_seat_type IS NOT NULL
	`, tenantID, strings.TrimSpace(userID))
	if err != nil {
		return nil, err
	}

	if tag.RowsAffected() == 0 {
		return nil, seats.ErrNoPendingChange
	}

	return s.GetSeatMember(ctx, tenantID, userID, now)
}

// afterSeatChange recompiles the tenant's runtime profile and grants any seat
// credits the change made due. Both are best effort: the grant worker and the
// profile manager catch up on failure.
func (s *walletState) afterSeatChange(ctx context.Context, tenantID, userID string, now time.Time) {
	if s.runtimeProfileInvalidator != nil {
		s.runtimeProfileInvalidator.Invalidate(ctx, tenantID, seatsInvalidateReason)
	}

	if userID == "" {
		return
	}

	if _, err := s.grantMemberSeatCredits(ctx, tenantID, userID, now); err != nil && s.logger != nil {
		s.logger.Warn("seat credit grant after seat change failed")
	}
}
