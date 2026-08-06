package proxy

import (
	"context"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

type budgetReservationManager interface {
	ReserveBudget(ctx context.Context, auth *platform.AuthContext, breakdown platform.UsageCostBreakdown, requestID string) (*platform.BudgetReservation, error)
	CommitBudget(ctx context.Context, reservation *platform.BudgetReservation, breakdown platform.UsageCostBreakdown) error
	ReleaseBudget(ctx context.Context, reservation *platform.BudgetReservation, reason string) error
}
