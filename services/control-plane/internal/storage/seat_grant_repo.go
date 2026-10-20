package storage

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/seats"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

// seatGrantCandidate is a member whose seat may be due credits.
type seatGrantCandidate struct {
	TenantID    string
	UserID      string
	Seat        seats.Type
	PeriodStart time.Time
	PeriodEnd   time.Time
}

// SeatWindowAllowanceMC is the allowance of a member's seat window: their
// personal credits left, plus what the ledger already settled from them in the
// window (the window counts it as committed too).
func (s *SeatRepository) SeatWindowAllowanceMC(ctx context.Context, tenantID, userID, windowID string) (int64, error) {
	return seatWindowAllowance(ctx, s.pool, tenantID, userID, windowID, time.Now().UTC())
}

func seatWindowAllowance(ctx context.Context, q rowQuerier, tenantID, userID, windowID string, now time.Time) (int64, error) {
	var allowance int64

	err := q.QueryRow(ctx, `
		SELECT
			COALESCE((SELECT SUM(b.remaining_mc) FROM credit_lot_balances b
			          WHERE b.tenant_id = $1::uuid AND b.owner_user_id::text = $2 AND b.remaining_mc > 0
			            AND (b.expires_at IS NULL OR b.expires_at > $4)), 0)
			+ COALESCE((SELECT w.settled_mc FROM credit_user_window_settlements w
			            WHERE w.tenant_id = $1::uuid AND w.user_id::text = $2 AND w.window_id = $3), 0)
	`, strings.TrimSpace(tenantID), strings.TrimSpace(userID), strings.TrimSpace(windowID), now).Scan(&allowance)

	return allowance, err
}

// SeatPoolUsedMC is how much pool credit the ledger settled for a member in a
// calendar-month window.
func (s *SeatRepository) SeatPoolUsedMC(ctx context.Context, tenantID, userID, windowID string) (int64, error) {
	var used int64

	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE((SELECT pool_settled_mc FROM credit_user_window_settlements
		                 WHERE tenant_id = $1::uuid AND user_id::text = $2 AND window_id = $3), 0)
	`, tenantID, userID, windowID).Scan(&used)

	return used, err
}

// syncSeatWindow re-seeds a member's current seat window after their personal
// credits changed. A window that does not exist yet is seeded on first use.
func (s *walletState) syncSeatWindow(ctx context.Context, tenantID, userID string) {
	liveCache, ok := s.entitlementCache.(budgetWindowStateCache)
	if !ok || liveCache == nil || s.pool == nil {
		return
	}

	now := time.Now().UTC()
	windowID, windowEnd := credits.PoolWindow(now)
	windowKey := seats.WindowKey(tenantID, userID, windowID)

	snapshot, err := liveCache.GetBudgetWindow(ctx, windowKey)
	if err != nil || snapshot == nil {
		return
	}

	allowance, err := seatWindowAllowance(ctx, s.pool, tenantID, userID, windowID, now)
	if err != nil {
		s.logSeatWindowSyncFailure(err, tenantID)

		return
	}

	ttl := time.Until(windowEnd.Add(s.budgetWindowTTLGrace))
	if ttl <= 0 {
		ttl = s.budgetWindowTTLGrace
	}

	if _, err := liveCache.SeedBudgetWindow(ctx, cache.BudgetWindowSeed{
		WindowKey: windowKey, AllowanceIRR: float64(allowance), ExpiresAt: windowEnd, TTL: ttl, Version: windowEnd.Unix(),
	}); err != nil {
		s.logSeatWindowSyncFailure(err, tenantID)
	}
}

func (s *walletState) logSeatWindowSyncFailure(err error, tenantID string) {
	if s.logger != nil {
		s.logger.Warn("failed to re-seed seat window", zap.Error(err), zap.String("tenant_id", tenantID))
	}
}

// RunSeatLifecycle is one pass of the seat worker: it gives members without a
// seat their free seat, applies downgrades that are due, removes seats whose
// paid block lapsed, and grants the seat credits that are due.
func (s *SeatRepository) RunSeatLifecycle(ctx context.Context, now time.Time) (seats.LifecycleResult, error) {
	now = now.UTC()

	var result seats.LifecycleResult

	changed := map[string]struct{}{}
	defer s.invalidateSeatTenants(ctx, changed)

	assigned, err := s.autoAssignFreeSeats(ctx, now, changed)
	if err != nil {
		return result, err
	}

	result.AutoAssigned = assigned

	applied, err := s.collectSeatTenants(ctx, changed, `
		WITH due AS (
			DELETE FROM tenant_seat_assignments
			WHERE pending_seat_type = 'unassigned' AND pending_effective_at <= $1
			RETURNING tenant_id
		), moved AS (
			UPDATE tenant_seat_assignments
			SET seat_type = pending_seat_type, pending_seat_type = NULL, pending_effective_at = NULL,
			    assigned_by = 'scheduled', assigned_at = $1, updated_at = NOW()
			WHERE pending_seat_type = 'pro' AND pending_effective_at <= $1
			RETURNING tenant_id
		)
		SELECT tenant_id::text FROM due UNION ALL SELECT tenant_id::text FROM moved
	`, now)
	if err != nil {
		return result, err
	}

	result.DowngradesApplied = applied

	// A block that lapsed without renewal leaves more members holding a paid
	// seat than the tenant pays for; the most recently assigned lose it first.
	lapsed, err := s.collectSeatTenants(ctx, changed, `
		WITH held AS (
			SELECT tenant_id, seat_type, SUM(quantity)::int AS purchased
			FROM tenant_seat_blocks WHERE paid_through > $1
			GROUP BY tenant_id, seat_type
		), ranked AS (
			SELECT a.tenant_id, a.user_id,
			       ROW_NUMBER() OVER (PARTITION BY a.tenant_id, a.seat_type ORDER BY a.assigned_at, a.user_id) AS position,
			       COALESCE(h.purchased, 0) AS purchased
			FROM tenant_seat_assignments a
			JOIN users u ON u.user_id = a.user_id AND u.tenant_id = a.tenant_id AND COALESCE(u.active, true)
			LEFT JOIN held h ON h.tenant_id = a.tenant_id AND h.seat_type = a.seat_type
			WHERE a.seat_type IN ('pro', 'max')
		)
		DELETE FROM tenant_seat_assignments a
		USING ranked r
		WHERE a.tenant_id = r.tenant_id AND a.user_id = r.user_id AND r.position > r.purchased
		RETURNING a.tenant_id::text
	`, now)
	if err != nil {
		return result, err
	}

	result.LapsedAssignments = lapsed

	candidates, err := s.listSeatGrantCandidates(ctx, "", "")
	if err != nil {
		return result, err
	}

	types, err := loadSeatTypes(ctx, s.pool)
	if err != nil {
		return result, err
	}

	for _, candidate := range candidates {
		grantedMC, grantErr := s.grantSeatCandidate(ctx, candidate, types, now)
		if grantErr != nil {
			return result, grantErr
		}

		if grantedMC > 0 {
			result.Granted++
			result.GrantedMC += grantedMC
		}
	}

	return result, nil
}

// grantMemberSeatCredits grants whatever seat credits one member is due now.
func (s *walletState) grantMemberSeatCredits(ctx context.Context, tenantID, userID string, now time.Time) (int64, error) {
	candidates, err := listSeatGrantCandidatesQ(ctx, s.pool, tenantID, userID)
	if err != nil || len(candidates) == 0 {
		return 0, err
	}

	types, err := loadSeatTypes(ctx, s.pool)
	if err != nil {
		return 0, err
	}

	return s.grantSeatCandidate(ctx, candidates[0], types, now.UTC())
}

func (s *SeatRepository) listSeatGrantCandidates(ctx context.Context, tenantID, userID string) ([]seatGrantCandidate, error) {
	return listSeatGrantCandidatesQ(ctx, s.pool, tenantID, userID)
}

// listSeatGrantCandidatesQ lists active members holding a seat in credit
// tenants with an active subscription; trialing tenants share their trial pool.
func listSeatGrantCandidatesQ(ctx context.Context, q rowsQuerier, tenantID, userID string) ([]seatGrantCandidate, error) {
	rows, err := q.Query(ctx, `
		SELECT a.tenant_id::text, a.user_id::text, a.seat_type, ts.current_period_start, ts.current_period_end
		FROM tenant_seat_assignments a
		JOIN users u ON u.user_id = a.user_id AND u.tenant_id = a.tenant_id AND COALESCE(u.active, true)
		JOIN tenants t ON t.tenant_id = a.tenant_id AND t.billing_unit = $3
		JOIN LATERAL (
			SELECT status, current_period_start, current_period_end FROM tenant_subscriptions
			WHERE tenant_id = a.tenant_id AND status IN ('trialing', 'active', 'past_due')
			ORDER BY updated_at DESC LIMIT 1
		) ts ON ts.status = 'active'
		WHERE ($1 = '' OR a.tenant_id::text = $1) AND ($2 = '' OR a.user_id::text = $2)
		ORDER BY a.tenant_id, a.user_id
	`, strings.TrimSpace(tenantID), strings.TrimSpace(userID), credits.BillingUnitCredit)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (seatGrantCandidate, error) {
		var (
			item     seatGrantCandidate
			seatType string
		)

		scanErr := row.Scan(&item.TenantID, &item.UserID, &seatType, &item.PeriodStart, &item.PeriodEnd)
		item.Seat = seats.Type(seatType)

		return item, scanErr
	})
}

// grantSeatCandidate grants a member's free seat once, or the monthly slice of
// a paid seat. Keys make every grant happen at most once.
func (s *walletState) grantSeatCandidate(ctx context.Context, candidate seatGrantCandidate, types map[seats.Type]seatTypeRow, now time.Time) (int64, error) {
	row, ok := types[candidate.Seat]
	if !ok || row.MonthlyGrantMC <= 0 {
		return 0, nil
	}

	ref := map[string]interface{}{"seat": string(candidate.Seat), "user_id": candidate.UserID}
	grant := CreditGrant{
		TenantID: candidate.TenantID, AmountMC: row.MonthlyGrantMC, OccurredAt: now,
		OwnerUserID: candidate.UserID, SourceRef: ref,
	}

	if row.Lifetime {
		grant.Kind, grant.Source = credits.EntryPromoGrant, credits.LotSourceSeatFree
		grant.IdempotencyKey = seats.FreeGrantKey(candidate.TenantID, candidate.UserID)
	} else {
		slice, inPeriod := credits.CurrentPlanSlice(candidate.PeriodStart, candidate.PeriodEnd, now)
		if !inPeriod {
			return 0, nil
		}

		expiresAt := slice.End
		ref["slice_start"], ref["slice_end"] = slice.Start, slice.End
		grant.Kind, grant.Source, grant.ExpiresAt = credits.EntryPlanGrant, credits.LotSourceSeat, &expiresAt
		grant.IdempotencyKey = seats.GrantKey(candidate.TenantID, candidate.UserID, candidate.Seat, slice.Start)
	}

	result, err := s.grantPersonalCredits(ctx, grant)
	if err != nil || result.Duplicate {
		return 0, err
	}

	s.syncSeatWindow(ctx, candidate.TenantID, candidate.UserID)

	return grant.AmountMC, nil
}

// grantPersonalCredits writes a personal grant. The tenant's pool window is
// unaffected, so only the member's seat window needs a refresh.
func (s *walletState) grantPersonalCredits(ctx context.Context, grant CreditGrant) (CreditGrantResult, error) {
	if err := validateCreditGrant(grant); err != nil {
		return CreditGrantResult{}, err
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

	return result, tx.Commit(ctx)
}

// autoAssignFreeSeats gives the free seat to active members of credit tenants
// who have no seat and never had the free seat, while the plan has free seats
// left. It covers members created outside invitations, such as tenant owners.
func (s *SeatRepository) autoAssignFreeSeats(ctx context.Context, now time.Time, changed map[string]struct{}) (int, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT u.tenant_id::text, u.user_id::text
		FROM users u
		JOIN tenants t ON t.tenant_id = u.tenant_id AND t.billing_unit = $1
		WHERE COALESCE(u.active, true)
		  AND NOT EXISTS (SELECT 1 FROM tenant_seat_assignments a WHERE a.tenant_id = u.tenant_id AND a.user_id = u.user_id)
		  AND NOT EXISTS (SELECT 1 FROM tenant_free_seat_grants g WHERE g.tenant_id = u.tenant_id AND g.user_id = u.user_id)
		ORDER BY u.tenant_id, u.created_at, u.user_id
	`, credits.BillingUnitCredit)
	if err != nil {
		return 0, err
	}

	members, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) ([2]string, error) {
		var item [2]string

		return item, row.Scan(&item[0], &item[1])
	})
	if err != nil {
		return 0, err
	}

	assigned := 0

	for _, member := range members {
		ok, assignErr := s.assignFreeSeatIfRoom(ctx, member[0], member[1], now)
		if assignErr != nil {
			return assigned, assignErr
		}

		if ok {
			assigned++
			changed[member[0]] = struct{}{}
		}
	}

	return assigned, nil
}

func (s *SeatRepository) assignFreeSeatIfRoom(ctx context.Context, tenantID, userID string, now time.Time) (bool, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return false, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	_, err = setMemberSeatTx(ctx, tx, tenantID, userID, seats.Free, "system", now)
	if errors.Is(err, seats.ErrFreeSeatsFull) || errors.Is(err, seats.ErrFreeSeatUsed) || errors.Is(err, seats.ErrSameSeat) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, tx.Commit(ctx)
}

// collectSeatTenants runs a statement that returns the tenant id of every
// seat it changed, and adds them to changed.
func (s *SeatRepository) collectSeatTenants(ctx context.Context, changed map[string]struct{}, sql string, args ...any) (int, error) {
	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return 0, err
	}

	tenants, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return 0, err
	}

	for _, tenantID := range tenants {
		changed[tenantID] = struct{}{}
	}

	return len(tenants), nil
}

// invalidateSeatTenants recompiles the profiles of tenants whose seats changed.
func (s *SeatRepository) invalidateSeatTenants(ctx context.Context, changed map[string]struct{}) {
	if s.runtimeProfileInvalidator == nil {
		return
	}

	for tenantID := range changed {
		s.runtimeProfileInvalidator.Invalidate(ctx, tenantID, seatsInvalidateReason)
	}
}
