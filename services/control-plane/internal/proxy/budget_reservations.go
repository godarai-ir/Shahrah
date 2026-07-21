package proxy

import (
	"context"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type budgetReservationManager interface {
	ReserveBudget(ctx context.Context, auth *platform.AuthContext, breakdown platform.UsageCostBreakdown, requestID string) (*platform.BudgetReservation, error)
	CommitBudget(ctx context.Context, reservation *platform.BudgetReservation, breakdown platform.UsageCostBreakdown) error
	ReleaseBudget(ctx context.Context, reservation *platform.BudgetReservation, reason string) error
}

func reserveBudgetForRequest(c *fiber.Ctx, billing budgeter, authInfo *platform.AuthContext, estimated platform.UsageCostBreakdown, requestID string, billingMode ...string) (*platform.BudgetReservation, error) {
	if firstBillingMode(billingMode) == platform.BillingModeExternalBilled {
		return nil, nil
	}

	estimated = enrichBudgetEvaluationContext(c, estimated)
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

func enrichBudgetEvaluationContext(c *fiber.Ctx, estimated platform.UsageCostBreakdown) platform.UsageCostBreakdown {
	if c == nil {
		return estimated
	}

	if strings.TrimSpace(estimated.Endpoint) == "" {
		estimated.Endpoint = c.Path()
	}

	if strings.TrimSpace(estimated.RequestType) == "" {
		estimated.RequestType = budgetRequestTypeFromPath(c.Path())
	}

	return estimated
}

func budgetRequestTypeFromPath(path string) string {
	normalized := strings.Trim(strings.ToLower(path), "/")
	normalized = strings.TrimPrefix(normalized, "v1/")

	return strings.ReplaceAll(normalized, "/", "_")
}

func commitBudgetOrDebit(ctx context.Context, logger *zap.Logger, billing budgeter, wallets walletStore, reservation *platform.BudgetReservation, tenantID string, breakdown platform.UsageCostBreakdown, requestID string, billingMode ...string) error {
	if firstBillingMode(billingMode) == platform.BillingModeExternalBilled {
		releaseBudgetBestEffort(ctx, logger, billing, reservation, "external_billed")
		return nil
	}

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

func firstBillingMode(values []string) string {
	if len(values) == 0 {
		return platform.BillingModeWalletBilled
	}

	mode := strings.TrimSpace(values[0])
	if mode == "" {
		return platform.BillingModeWalletBilled
	}

	return mode
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
