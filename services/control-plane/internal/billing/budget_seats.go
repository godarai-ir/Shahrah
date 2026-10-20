package billing

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/seats"
	"github.com/Delpak/shahrah/services/control-plane/internal/runtimeprofile"
)

// Seat budget targets: a member's seat credits, and their monthly pool limit.
const (
	budgetTargetSeatCredit    = "seat_credit"
	budgetTargetSeatPoolLimit = "seat_pool_limit"
	budgetScopeSeat           = "seat"
)

// SeatCreditStore reads what seat admission needs beyond the runtime profile.
type SeatCreditStore interface {
	// SeatWindowAllowanceMC seeds a member's seat window.
	SeatWindowAllowanceMC(ctx context.Context, tenantID, userID, windowID string) (int64, error)
	// MemberSeatType resolves a member the compiled profile does not know yet,
	// such as one who joined after it was compiled.
	MemberSeatType(ctx context.Context, tenantID, userID string) (string, error)
}

// seatAdmission is how one request relates to the member's seat.
type seatAdmission struct {
	applies     bool
	userID      string
	seat        seats.Type
	enforced    bool
	poolAccess  bool
	poolLimitMC *int64
}

// seatAdmissionFor decides whether a request spends seat credits: chat usage
// of a known member, paid by the platform, in a tenant with seats. A member
// without a seat in a tenant without a subscription shares the pool.
func (s *BudgetService) seatAdmissionFor(ctx context.Context, auth *platform.AuthContext, profile *runtimeprofile.Profile, billingMode string, now time.Time) (seatAdmission, error) {
	if s.seatCredits == nil || auth == nil || profile == nil || profile.Seats == nil ||
		auth.CreditProduct != platform.CreditProductLibreChat || strings.TrimSpace(auth.UserID) == "" ||
		billingMode == platform.BillingModeExternalBilled {
		return seatAdmission{}, nil
	}

	userID := strings.TrimSpace(auth.UserID)

	seatType, known := profile.Seats.Members[userID]
	if !known {
		resolved, err := s.seatCredits.MemberSeatType(ctx, auth.TenantID, userID)
		if err != nil {
			return seatAdmission{}, fmt.Errorf("%w: seat lookup failed: %v", ErrBudgetUnavailable, err)
		}

		seatType = resolved
	}

	seat, err := seats.ParseType(seatType)
	if err != nil {
		seat = seats.Unassigned
	}

	enforced := seats.Enforced(profile.Seats.SubscriptionStatus)
	if seat == seats.Unassigned && !enforced {
		return seatAdmission{}, nil
	}

	admission := seatAdmission{
		applies: true, userID: userID, seat: seat, enforced: enforced, poolAccess: seats.PoolAccess(seat, enforced),
	}

	if limit, ok := profile.Seats.PoolLimitMC[string(seat)]; ok && enforced && seat.Paid() {
		admission.poolLimitMC = &limit
	}

	return admission, nil
}

// reserveSeatTargets admits a chat request against the member's seat. While
// the seat has credits only the seat window is reserved, and any overflow is
// charged to the pool at commit. Once it is spent, members with pool access
// fall through to the pool and their pool limit; others are refused.
func (s *BudgetService) reserveSeatTargets(ctx context.Context, tenantID string, admission seatAdmission, reservationID string, now time.Time) ([]budgetReservationTarget, error) {
	if admission.seat == seats.Unassigned {
		return nil, seatRefusal(seats.BlockSeatUnassigned)
	}

	seatTarget := s.seatCreditTarget(tenantID, admission.userID, reservationID, now)

	result, err := s.reserveBudgetTarget(ctx, s.budgetWindows, seatTarget, reservationID, now)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, ErrBudgetUnavailable
	}

	if result.Allowed {
		return []budgetReservationTarget{seatTarget}, nil
	}

	if result.State != "exhausted" {
		return nil, ErrBudgetUnavailable
	}

	if !admission.poolAccess {
		return nil, seatRefusal(seats.BlockSeatCreditsExhausted)
	}

	targets := []budgetReservationTarget{s.creditPoolTarget(tenantID, reservationID, 0, now)}
	if admission.poolLimitMC != nil {
		targets = append(targets, s.seatPoolLimitTarget(tenantID, admission.userID, reservationID, *admission.poolLimitMC, now))
	}

	reserved, err := s.reserveCreditTargets(ctx, targets, reservationID, now)
	if err != nil {
		return nil, withSeatRefusalCode(err)
	}

	return reserved, nil
}

func (s *BudgetService) seatCreditTarget(tenantID, userID, reservationID string, now time.Time) budgetReservationTarget {
	windowID, windowEnd := credits.PoolWindow(now)

	return budgetReservationTarget{
		tenantID: tenantID,
		BudgetReservationTarget: platform.BudgetReservationTarget{
			Kind:           budgetTargetSeatCredit,
			Scope:          budgetScopeSeat,
			ScopeValue:     userID,
			WindowID:       windowID,
			WindowKey:      seats.WindowKey(tenantID, userID, windowID),
			ReservationKey: seats.ReservationKey(tenantID, userID, reservationID),
			CreatedAt:      now,
		},
		expiresAt:    windowEnd,
		ttl:          s.windowTTL(windowEnd),
		seedSeatUser: userID,
	}
}

func (s *BudgetService) seatPoolLimitTarget(tenantID, userID, reservationID string, limitMC int64, now time.Time) budgetReservationTarget {
	windowID, windowEnd := credits.PoolWindow(now)
	windowKey := seats.PoolLimitWindowKey(tenantID, userID, windowID)

	return budgetReservationTarget{
		tenantID: tenantID,
		BudgetReservationTarget: platform.BudgetReservationTarget{
			Kind:           budgetTargetSeatPoolLimit,
			Scope:          budgetTargetSeatPoolLimit,
			ScopeValue:     userID,
			WindowID:       windowID,
			WindowKey:      windowKey,
			ReservationKey: credits.PolicyReservationKey(tenantID, reservationID, shortHash(windowKey)),
			CreatedAt:      now,
		},
		allowanceIRR: float64(limitMC),
		expiresAt:    windowEnd,
		ttl:          s.windowTTL(windowEnd),
	}
}

// seedSeatWindow seeds a member's seat window from their personal credits.
func (s *BudgetService) seedSeatWindow(ctx context.Context, budgetCache BudgetWindowCache, target budgetReservationTarget) error {
	allowanceMC, err := s.seatCredits.SeatWindowAllowanceMC(ctx, target.tenantID, target.seedSeatUser, target.WindowID)
	if err != nil {
		return fmt.Errorf("%w: seat window seed failed: %v", ErrBudgetUnavailable, err)
	}

	result, err := budgetCache.SeedBudgetWindow(ctx, cache.BudgetWindowSeed{
		WindowKey: target.WindowKey, AllowanceIRR: float64(allowanceMC),
		ExpiresAt: target.expiresAt, TTL: target.ttl, Version: target.expiresAt.Unix(),
	})
	if err != nil {
		return fmt.Errorf("%w: seat window seed failed: %v", ErrBudgetUnavailable, err)
	}

	if result == nil || !result.Allowed {
		return ErrBudgetUnavailable
	}

	return nil
}

func seatRefusal(reason seats.BlockReason) error {
	return &BudgetExceededError{Scope: budgetScopeSeat, Code: string(reason)}
}

// withSeatRefusalCode names the seat block reason of a refusal on the pool
// side of a chat request.
func withSeatRefusalCode(err error) error {
	var exceeded *BudgetExceededError
	if !errors.As(err, &exceeded) || exceeded.Code != "" {
		return err
	}

	switch exceeded.Scope {
	case BudgetScopeTenantBalance:
		exceeded.Code = string(seats.BlockPoolExhausted)
	case budgetTargetSeatPoolLimit:
		exceeded.Code = string(seats.BlockPersonalLimitReached)
	default:
		exceeded.Code = string(seats.BlockBudgetRuleReached)
	}

	return err
}

// BudgetErrorCode returns the seat block reason of a budget refusal, or "".
func BudgetErrorCode(err error) string {
	var exceeded *BudgetExceededError
	if errors.As(err, &exceeded) {
		return exceeded.Code
	}

	return ""
}

// commitSeatOverflow charges what a chat request took beyond the member's
// seat credits to the pool and the member's pool limit, for the windows the
// reservation did not reserve. The seat window was committed already.
func (s *BudgetService) commitSeatOverflow(ctx context.Context, reservation *platform.BudgetReservation, overflowMC int64, now time.Time) error {
	if overflowMC <= 0 {
		return nil
	}

	recorder, ok := s.budgetWindows.(budgetSpendRecorder)
	if !ok {
		return errors.New("budget windows cannot record spend")
	}

	var targets []budgetReservationTarget

	if !reservationHasTarget(reservation, budgetTargetTenantCredit) {
		targets = append(targets, s.creditPoolTarget(reservation.TenantID, reservation.ID, overflowMC, now))
	}

	if reservation.SeatPoolLimitMC != nil && !reservationHasTarget(reservation, budgetTargetSeatPoolLimit) {
		targets = append(targets, s.seatPoolLimitTarget(reservation.TenantID, reservation.SeatUserID, reservation.ID, *reservation.SeatPoolLimitMC, now))
	}

	var recordErrors []error

	for _, target := range targets {
		if _, err := s.recordSideUsageTarget(ctx, recorder, target, overflowMC, now); err != nil {
			recordErrors = append(recordErrors, err)
		}
	}

	return errors.Join(recordErrors...)
}

func reservationHasTarget(reservation *platform.BudgetReservation, kind string) bool {
	for _, target := range reservation.Targets {
		if target.Kind == kind {
			return true
		}
	}

	return false
}
