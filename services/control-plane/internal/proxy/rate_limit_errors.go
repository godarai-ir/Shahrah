package proxy

import (
	"errors"

	"github.com/Delpak/shahrah/services/control-plane/internal/billing"
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
