package proxy

import (
	"errors"

	"github.com/Delpak/shahrah/services/control-plane/internal/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/gofiber/fiber/v2"
)

func rateLimitErrorStatus(err error) int {
	switch {
	case errors.Is(err, billing.ErrSubscriptionInactive):
		return fiber.StatusForbidden
	case errors.Is(err, billing.ErrBudgetExceeded):
		return fiber.StatusPaymentRequired
	default:
		return fiber.StatusTooManyRequests
	}
}

func writeBudgetCheckError(c *fiber.Ctx, err error) error {
	status, message, code := budgetCheckErrorResponse(err)
	return writeAPIError(c, status, message, code)
}

func budgetCheckErrorResponse(err error) (int, string, string) {
	if errors.Is(err, billing.ErrBudgetExceeded) || errors.Is(err, storage.ErrInsufficientWalletBalance) {
		return fiber.StatusPaymentRequired, err.Error(), "budget_exceeded"
	}

	return fiber.StatusServiceUnavailable, "Billing budget check unavailable", "billing_unavailable"
}
