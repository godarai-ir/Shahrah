package billing

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
)

// CheckBudget validates wallet balance constraints for compatibility callers.
func (s *BudgetService) CheckBudget(
	ctx context.Context,
	auth *platform.AuthContext,
	breakdown platform.UsageCostBreakdown,
) error {
	if auth == nil {
		return ErrBudgetExceeded
	}

	reservationID := fmt.Sprintf("compat:%s:%d", strings.TrimSpace(auth.TokenHash), time.Now().UnixNano())

	reservation, err := s.ReserveBudget(ctx, auth, breakdown, reservationID)
	if err != nil {
		return err
	}

	if reservation != nil {
		if err := s.ReleaseBudget(ctx, reservation, "compat_check_only"); err != nil {
			return err
		}
	}

	return nil
}

// ReserveBudget reserves the estimated request credits from the tenant credit
// pool and the matching budget rule windows.
func (s *BudgetService) ReserveBudget(
	ctx context.Context,
	auth *platform.AuthContext,
	breakdown platform.UsageCostBreakdown,
	requestID string,
) (*platform.BudgetReservation, error) {
	if auth == nil || strings.TrimSpace(auth.TenantID) == "" {
		return nil, ErrBudgetExceeded
	}

	profile, err := s.tenantProfile(ctx, strings.TrimSpace(auth.TenantID))
	if err != nil {
		return nil, err
	}

	return s.reserveCreditBudget(ctx, auth, breakdown, budgetReservationID(requestID), profile, platform.BillingModeWalletBilled)
}

// CommitBudget charges a reservation with the request's actual credits.
func (s *BudgetService) CommitBudget(
	ctx context.Context,
	reservation *platform.BudgetReservation,
	breakdown platform.UsageCostBreakdown,
) error {
	if reservation == nil {
		return nil
	}

	if s.budgetWindows == nil {
		return ErrBudgetUnavailable
	}

	return s.commitCreditBudget(ctx, reservation, breakdown)
}

// ReleaseBudget releases a reservation that did not become chargeable usage.
func (s *BudgetService) ReleaseBudget(
	ctx context.Context,
	reservation *platform.BudgetReservation,
	_ string,
) error {
	if reservation == nil {
		return nil
	}

	budgetCache := s.budgetWindows
	if budgetCache == nil {
		return ErrBudgetUnavailable
	}

	targets := reservation.Targets
	if len(targets) == 0 {
		targets = []platform.BudgetReservationTarget{{
			WindowKey:      reservation.WindowKey,
			ReservationKey: reservation.ReservationKey,
		}}
	}

	now := time.Now().UTC()
	for _, target := range targets {
		result, err := budgetCache.ReleaseBudget(ctx, cache.BudgetReleaseRequest{
			WindowKey:      target.WindowKey,
			ReservationKey: target.ReservationKey,
			Now:            now,
		})
		if err != nil {
			return fmt.Errorf("%w: release failed: %v", ErrBudgetUnavailable, err)
		}

		if result == nil || !result.Allowed {
			return ErrBudgetUnavailable
		}
	}

	return nil
}

func budgetWindowForTime(now time.Time) (string, time.Time) {
	utc := now.UTC()
	start := time.Date(utc.Year(), utc.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)

	return start.Format("200601"), end
}

func budgetReservationID(requestID string) string {
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return fmt.Sprintf("reservation:%d", time.Now().UnixNano())
	}

	return requestID
}
