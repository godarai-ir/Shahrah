package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

// creditBudgetManager records which budget calls the pipeline makes.
type creditBudgetManager struct {
	externalReservation *platform.BudgetReservation
	reserved            int
	externalReserved    int
	committed           int
	released            int
}

func (m *creditBudgetManager) ReserveBudget(context.Context, *platform.AuthContext, platform.UsageCostBreakdown, string) (*platform.BudgetReservation, error) {
	m.reserved++

	return &platform.BudgetReservation{ID: "r"}, nil
}

func (m *creditBudgetManager) ReserveExternalBilledBudget(context.Context, *platform.AuthContext, platform.UsageCostBreakdown, string) (*platform.BudgetReservation, error) {
	m.externalReserved++

	return m.externalReservation, nil
}

func (m *creditBudgetManager) CommitBudget(context.Context, *platform.BudgetReservation, platform.UsageCostBreakdown) error {
	m.committed++

	return nil
}

func (m *creditBudgetManager) ReleaseBudget(context.Context, *platform.BudgetReservation, string) error {
	m.released++

	return nil
}

func TestPipelineReservesExternalBilledRequestsOfCreditTenants(t *testing.T) {
	manager := &creditBudgetManager{
		externalReservation: &platform.BudgetReservation{ID: "byok", Unit: credits.BillingUnitCredit},
	}
	pipeline := NewRequestPipeline(nil, nil, &testBudgeter{}, BillingCapabilities{
		BudgetReservations:      manager,
		BudgetReservationPolicy: CapabilityFailClosed,
	}, nil, nil)

	reservation, err := pipeline.Reserve(t.Context(), BudgetRequest{
		Auth: &platform.AuthContext{TenantID: "t"}, RequestID: "req", BillingMode: platform.BillingModeExternalBilled,
	})
	require.NoError(t, err)
	require.Equal(t, "byok", reservation.ID)
	require.Equal(t, 1, manager.externalReserved)
	require.Zero(t, manager.reserved, "BYOK never takes the wallet-billed path")

	require.NoError(t, pipeline.Settle(t.Context(), SettlementRequest{Reservation: reservation, BillingMode: platform.BillingModeExternalBilled}))
	require.Equal(t, 1, manager.committed, "a credit BYOK reservation is committed, not released")

	irr := &platform.BudgetReservation{ID: "irr"}
	require.NoError(t, pipeline.Settle(t.Context(), SettlementRequest{Reservation: irr, BillingMode: platform.BillingModeExternalBilled}))
	require.Equal(t, 1, manager.released, "an IRR BYOK reservation is still only released")
}

func decodeErrorBody(t *testing.T, app *fiber.App) (int, map[string]interface{}) {
	t.Helper()

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", http.NoBody))
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(raw, &body))

	return resp.StatusCode, body
}

func TestBudgetErrorsExposeOnlyCodeAndScope(t *testing.T) {
	refusal := errors.Join(&billing.BudgetExceededError{Scope: "team"}, errors.New("release failed: redis 10.0.0.7 timeout"))

	writers := map[string]func(*fiber.Ctx, error) error{
		"openai":    writeBudgetCheckError,
		"anthropic": writeAnthropicBudgetError,
		"gemini":    writeGeminiBudgetError,
	}

	for name, write := range writers {
		t.Run(name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/", func(c *fiber.Ctx) error { return write(c, refusal) })

			status, body := decodeErrorBody(t, app)
			require.Equal(t, fiber.StatusPaymentRequired, status)

			inner, ok := body["error"].(map[string]interface{})
			require.True(t, ok)
			require.Equal(t, "team", inner["scope"])
			require.Equal(t, "budget exceeded", inner["message"], "no internal error text reaches the client")

			if name == "gemini" {
				require.Equal(t, "budget_exceeded", inner["reason"])
			} else {
				require.Equal(t, "budget_exceeded", inner["code"])
				require.Equal(t, "budget_exceeded", inner["type"])
			}
		})
	}
}

func TestSettlementErrorMessageHidesInternalText(t *testing.T) {
	require.Equal(t, "budget exceeded", settlementErrorMessage(errors.Join(billing.ErrBudgetExceeded, errors.New("window credit:tenant:x"))))
	require.Equal(t, "Billing settlement failed", settlementErrorMessage(errors.New("commit failed: dial tcp 10.0.0.7:6379")))
}
