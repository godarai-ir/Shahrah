package proxy

import (
	"errors"
	"strings"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/services/control-plane/internal/billing"
	billingdomain "github.com/Delpak/shahrah/services/control-plane/internal/domain/billing"
	"github.com/gofiber/fiber/v2"
)

const (
	budgetExceededCode       = "budget_exceeded"
	billingUnavailableCode   = "billing_unavailable"
	billingUnavailableReason = "Billing budget check unavailable"
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

// writeBudgetCheckError writes an OpenAI-compatible budget error. It exposes
// only a stable code and the refusing scope, never amounts or internal text.
func writeBudgetCheckError(c *fiber.Ctx, err error) error {
	status, message, code := budgetCheckErrorResponse(err)

	return c.Status(status).JSON(contractsopenai.APIError{
		Error: contractsopenai.ErrorInfo{
			Message: message,
			Type:    code,
			Code:    code,
			Scope:   budgetErrorScope(err),
		},
	})
}

// writeAnthropicBudgetError writes a budget error in the Anthropic error shape.
func writeAnthropicBudgetError(c *fiber.Ctx, err error) error {
	status, message, code := budgetCheckErrorResponse(err)

	body := fiber.Map{"type": code, "code": code, "message": message}
	if scope := budgetErrorScope(err); scope != "" {
		body["scope"] = scope
	}

	return c.Status(status).JSON(fiber.Map{
		"type":       "error",
		"error":      body,
		"request_id": strings.TrimSpace(c.GetRespHeader("request-id")),
	})
}

// writeGeminiBudgetError writes a budget error in the Gemini error shape.
func writeGeminiBudgetError(c *fiber.Ctx, err error) error {
	status, message, code := budgetCheckErrorResponse(err)

	body := fiber.Map{"code": status, "status": strings.ToUpper(code), "message": message, "reason": code}
	if scope := budgetErrorScope(err); scope != "" {
		body["scope"] = scope
	}

	return c.Status(status).JSON(fiber.Map{"error": body})
}

// seatBlockMessages are the client-safe messages of chat requests refused by
// a member's seat. GodarChat shows its own localized text per code.
var seatBlockMessages = map[string]string{
	"seat_unassigned":        "You have no seat in this organization yet",
	"seat_credits_exhausted": "Your free seat credits are used up",
	"personal_limit_reached": "You reached your personal usage limit for this month",
	"budget_rule_reached":    "A budget set for you or your team is used up",
	"pool_exhausted":         "Your organization has no credits left",
}

func budgetCheckErrorResponse(err error) (int, string, string) {
	if code := billing.BudgetErrorCode(err); code != "" && isBudgetExceeded(err) {
		message, ok := seatBlockMessages[code]
		if !ok {
			message = billing.ErrBudgetExceeded.Error()
		}

		return fiber.StatusPaymentRequired, message, code
	}

	if isBudgetExceeded(err) {
		return fiber.StatusPaymentRequired, billing.ErrBudgetExceeded.Error(), budgetExceededCode
	}

	return fiber.StatusServiceUnavailable, billingUnavailableReason, billingUnavailableCode
}

// settlementErrorMessage is the client-safe text of a failed settlement.
func settlementErrorMessage(err error) string {
	if isBudgetExceeded(err) {
		return billing.ErrBudgetExceeded.Error()
	}

	return "Billing settlement failed"
}

func isBudgetExceeded(err error) bool {
	return errors.Is(err, billing.ErrBudgetExceeded) || errors.Is(err, billingdomain.ErrInsufficientWalletBalance)
}

// budgetErrorScope is the scope of a credit budget refusal; IRR refusals have
// none, because the IRR path does not track which window refused.
func budgetErrorScope(err error) string {
	if !isBudgetExceeded(err) {
		return ""
	}

	return billing.BudgetErrorScope(err)
}
