package proxy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/billing"
	"github.com/gofiber/fiber/v2"
)

type rejectingBudgetManager struct {
	testBudgeter
}

func (b *rejectingBudgetManager) ReserveBudget(context.Context, *platform.AuthContext, platform.UsageCostBreakdown, string) (*platform.BudgetReservation, error) {
	return nil, billing.ErrBudgetExceeded
}

func (b *rejectingBudgetManager) CommitBudget(context.Context, *platform.BudgetReservation, platform.UsageCostBreakdown) error {
	return nil
}

func (b *rejectingBudgetManager) ReleaseBudget(context.Context, *platform.BudgetReservation, string) error {
	return nil
}

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

func TestRequestPipelineReserveErrorCanBeWrittenAtHTTPEdge(t *testing.T) {
	app := fiber.New()
	reachedProvider := false
	manager := &rejectingBudgetManager{}
	pipeline := NewRequestPipeline(nil, nil, manager, BillingCapabilities{
		BudgetReservations:      manager,
		BudgetReservationPolicy: CapabilityFailClosed,
	}, nil)

	app.Post("/", func(c *fiber.Ctx) error {
		_, err := pipeline.Reserve(c.Context(), budgetRequestFromHTTP(
			c,
			&platform.AuthContext{TenantID: "tenant-1"},
			platform.UsageCostBreakdown{TotalCostIRR: 1200},
			"request-1",
		))
		if err != nil {
			return writeBudgetCheckError(c, err)
		}

		reachedProvider = true

		return c.SendStatus(fiber.StatusNoContent)
	})

	response, err := app.Test(httptest.NewRequest(http.MethodPost, "/", http.NoBody))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusPaymentRequired {
		t.Fatalf("expected status 402, got %d", response.StatusCode)
	}

	if reachedProvider {
		t.Fatal("provider path was reached after budget response was written")
	}
}
