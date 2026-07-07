package proxy

import (
	"context"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type budgetReservationManager interface {
	ReserveBudget(ctx context.Context, auth *platform.AuthContext, breakdown platform.UsageCostBreakdown, requestID string) (*platform.BudgetReservation, error)
	CommitBudget(ctx context.Context, reservation *platform.BudgetReservation, breakdown platform.UsageCostBreakdown) error
	ReleaseBudget(ctx context.Context, reservation *platform.BudgetReservation, reason string) error
}

func reserveBudgetForRequest(c *fiber.Ctx, billing budgeter, authInfo *platform.AuthContext, estimated platform.UsageCostBreakdown, requestID string) (*platform.BudgetReservation, error) {
	if manager, ok := billing.(budgetReservationManager); ok {
		reservation, err := manager.ReserveBudget(c.Context(), authInfo, estimated, requestID)
		if err != nil {
			return nil, writeBudgetCheckError(c, err)
		}
		return reservation, nil
	}

	budgetErr := billing.CheckBudget(c.Context(), authInfo, estimated)
	if budgetErr != nil {
		return nil, writeBudgetCheckError(c, budgetErr)
	}
	return nil, nil
}

func commitBudgetOrDebit(ctx context.Context, logger *zap.Logger, billing budgeter, wallets walletStore, reservation *platform.BudgetReservation, tenantID string, breakdown platform.UsageCostBreakdown, requestID string) error {
	if manager, ok := billing.(budgetReservationManager); ok && reservation != nil {
		return manager.CommitBudget(ctx, reservation, breakdown)
	}
	if wallets == nil {
		return nil
	}
	_, err := wallets.DebitWallet(ctx, tenantID, breakdown, requestID)
	if err != nil && logger != nil {
		logger.Warn("failed to debit wallet", zap.Error(err), zap.String("request_id", requestID))
	}
	return err
}

func releaseBudgetBestEffort(ctx context.Context, logger *zap.Logger, billing budgeter, reservation *platform.BudgetReservation, reason string) {
	if reservation == nil {
		return
	}
	manager, ok := billing.(budgetReservationManager)
	if !ok {
		return
	}
	if err := manager.ReleaseBudget(ctx, reservation, reason); err != nil && logger != nil {
		logger.Warn("failed to release budget reservation", zap.Error(err), zap.String("request_id", reservation.ID), zap.String("reason", reason))
	}
}
