package proxy

import (
	"context"
	"testing"

	"github.com/Delpak/shahrah/services/control-plane/internal/billing"
	"github.com/gofiber/fiber/v2"
)

func TestBudgetCheckErrorResponseMapsBudgetExceededToPaymentRequired(t *testing.T) {
	status, message, code := budgetCheckErrorResponse(billing.ErrBudgetExceeded)

	if status != fiber.StatusPaymentRequired {
		t.Fatalf("expected status 402, got %d", status)
	}

	if message != billing.ErrBudgetExceeded.Error() {
		t.Fatalf("expected budget message, got %q", message)
	}

	if code != "budget_exceeded" {
		t.Fatalf("expected budget_exceeded code, got %q", code)
	}
}

func TestBudgetCheckErrorResponseSanitizesStoreErrorsAsUnavailable(t *testing.T) {
	status, message, code := budgetCheckErrorResponse(context.DeadlineExceeded)

	if status != fiber.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", status)
	}

	if message != "Billing budget check unavailable" {
		t.Fatalf("expected sanitized message, got %q", message)
	}

	if code != "billing_unavailable" {
		t.Fatalf("expected billing_unavailable code, got %q", code)
	}
}

func TestBudgetCheckErrorResponseMapsBudgetUnavailableToServiceUnavailable(t *testing.T) {
	status, message, code := budgetCheckErrorResponse(billing.ErrBudgetUnavailable)

	if status != fiber.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", status)
	}

	if message != "Billing budget check unavailable" {
		t.Fatalf("expected sanitized message, got %q", message)
	}

	if code != "billing_unavailable" {
		t.Fatalf("expected billing_unavailable code, got %q", code)
	}
}
