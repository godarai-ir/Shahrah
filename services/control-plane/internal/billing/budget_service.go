package billing

import (
	"context"
	"encoding/json"
	"errors"
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

// ReserveBudget reserves the estimated request cost from the tenant credit window and matching budget policy windows.
func (s *BudgetService) ReserveBudget(
	ctx context.Context,
	auth *platform.AuthContext,
	breakdown platform.UsageCostBreakdown,
	requestID string,
) (*platform.BudgetReservation, error) {
	if auth == nil || strings.TrimSpace(auth.TenantID) == "" {
		return nil, ErrBudgetExceeded
	}

	budgetCache := s.budgetWindows
	if budgetCache == nil {
		return nil, ErrBudgetUnavailable
	}

	tenantID := strings.TrimSpace(auth.TenantID)
	now := time.Now().UTC()
	reservationID := budgetReservationID(requestID)

	targets, err := s.budgetReservationTargets(ctx, auth, breakdown, reservationID, now)
	if err != nil {
		return nil, err
	}

	reserved := make([]budgetReservationTarget, 0, len(targets))
	for _, target := range targets {
		result, reserveErr := s.reserveBudgetTarget(ctx, budgetCache, target, reservationID, now)
		if reserveErr != nil {
			releaseErr := s.releaseBudgetTargetsBestEffort(ctx, budgetCache, reserved, now)

			return nil, errors.Join(reserveErr, releaseErr)
		}

		if result == nil {
			releaseErr := s.releaseBudgetTargetsBestEffort(ctx, budgetCache, reserved, now)

			return nil, errors.Join(ErrBudgetUnavailable, releaseErr)
		}

		if !result.Allowed {
			releaseErr := s.releaseBudgetTargetsBestEffort(ctx, budgetCache, reserved, now)
			if result.State == "exhausted" {
				return nil, errors.Join(ErrBudgetExceeded, releaseErr)
			}

			return nil, errors.Join(ErrBudgetUnavailable, releaseErr)
		}

		reserved = append(reserved, target)
	}

	if len(reserved) == 0 {
		return nil, ErrBudgetUnavailable
	}

	publicTargets := make([]platform.BudgetReservationTarget, 0, len(reserved))
	for _, target := range reserved {
		publicTargets = append(publicTargets, target.BudgetReservationTarget)
	}

	primary := publicTargets[0]

	return &platform.BudgetReservation{
		ID:             reservationID,
		TenantID:       tenantID,
		WindowID:       primary.WindowID,
		WindowKey:      primary.WindowKey,
		ReservationKey: primary.ReservationKey,
		EstimatedIRR:   breakdown.TotalCostIRR,
		CreatedAt:      now,
		Targets:        publicTargets,
	}, nil
}

// CommitBudget finalizes a reservation using the actual provider usage cost.
func (s *BudgetService) CommitBudget(
	ctx context.Context,
	reservation *platform.BudgetReservation,
	breakdown platform.UsageCostBreakdown,
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
		result, err := budgetCache.CommitBudget(ctx, cache.BudgetCommitRequest{
			WindowKey:      target.WindowKey,
			ReservationKey: target.ReservationKey,
			ActualIRR:      breakdown.TotalCostIRR,
			Now:            now,
		})
		if err != nil {
			return fmt.Errorf("%w: commit failed: %v", ErrBudgetUnavailable, err)
		}

		if result == nil || !result.Allowed {
			if result != nil && result.State == "exhausted" {
				return ErrBudgetExceeded
			}

			return ErrBudgetUnavailable
		}
	}

	return nil
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

func (s *BudgetService) seedBudgetWindow(
	ctx context.Context,
	budgetCache BudgetWindowCache,
	tenantID string,
	windowID string,
	windowKey string,
	windowEnd time.Time,
	ttl time.Duration,
) error {
	wallet, err := s.getWallet(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("%w: seed failed: %v", ErrBudgetUnavailable, err)
	}

	result, err := budgetCache.SeedBudgetWindow(ctx, cache.BudgetWindowSeed{
		WindowKey:    windowKey,
		AllowanceIRR: float64(wallet.BalanceIRR),
		ExpiresAt:    windowEnd,
		TTL:          ttl,
		Version:      windowEnd.Unix(),
	})
	if err != nil {
		return fmt.Errorf("%w: seed failed: %v", ErrBudgetUnavailable, err)
	}

	if result == nil || !result.Allowed {
		return ErrBudgetUnavailable
	}

	return nil
}

func (s *BudgetService) getWallet(ctx context.Context, tenantID string) (*platform.Wallet, error) {
	if s == nil || s.store == nil {
		return nil, errors.New("wallet store unavailable")
	}

	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" || s.cache == nil || s.walletCacheTTL <= 0 {
		return s.store.GetWallet(ctx, tenantID)
	}

	key := walletCacheKey(tenantID)
	if wallet, ok := s.cachedWallet(ctx, key); ok {
		return wallet, nil
	}

	wallet, err := s.store.GetWallet(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	s.storeWallet(ctx, key, wallet)

	return cloneWallet(wallet), nil
}

func (s *BudgetService) cachedWallet(ctx context.Context, key string) (*platform.Wallet, bool) {
	value, ok := s.cache.Get(ctx, key)
	if !ok {
		return nil, false
	}

	switch typed := value.(type) {
	case string:
		return walletFromJSON([]byte(typed))
	case []byte:
		return walletFromJSON(typed)
	case platform.Wallet:
		return cloneWallet(&typed), true
	case *platform.Wallet:
		return cloneWallet(typed), typed != nil
	default:
		s.cache.Delete(ctx, key)

		return nil, false
	}
}

func (s *BudgetService) storeWallet(ctx context.Context, key string, wallet *platform.Wallet) {
	if wallet == nil {
		return
	}

	raw, err := json.Marshal(wallet)
	if err != nil {
		return
	}

	s.cache.Set(ctx, key, string(raw), s.walletCacheTTL)
}

func walletFromJSON(raw []byte) (*platform.Wallet, bool) {
	var wallet platform.Wallet
	if err := json.Unmarshal(raw, &wallet); err != nil {
		return nil, false
	}

	return &wallet, true
}

func cloneWallet(input *platform.Wallet) *platform.Wallet {
	if input == nil {
		return nil
	}

	output := *input

	return &output
}

func walletCacheKey(tenantID string) string {
	return fmt.Sprintf("wallet:tenant:%s", strings.TrimSpace(tenantID))
}

func budgetWindowForTime(now time.Time) (string, time.Time) {
	utc := now.UTC()
	start := time.Date(utc.Year(), utc.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)

	return start.Format("200601"), end
}

func budgetWindowKey(tenantID string, windowID string) string {
	return fmt.Sprintf("budget:tenant:%s:window:%s", strings.TrimSpace(tenantID), strings.TrimSpace(windowID))
}

func budgetReservationKey(tenantID string, reservationID string) string {
	return fmt.Sprintf("budget:tenant:%s:reservation:%s", strings.TrimSpace(tenantID), strings.TrimSpace(reservationID))
}

func budgetReservationID(requestID string) string {
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return fmt.Sprintf("reservation:%d", time.Now().UnixNano())
	}

	return requestID
}
