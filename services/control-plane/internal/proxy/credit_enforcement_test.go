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
	"github.com/Delpak/shahrah/services/control-plane/internal/guardrails"
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
	}, nil)

	reservation, err := pipeline.Reserve(t.Context(), BudgetRequest{
		Auth: &platform.AuthContext{TenantID: "t"}, RequestID: "req", BillingMode: platform.BillingModeExternalBilled,
	})
	require.NoError(t, err)
	require.Equal(t, "byok", reservation.ID)
	require.Equal(t, 1, manager.externalReserved)
	require.Zero(t, manager.reserved, "BYOK never takes the wallet-billed path")

	require.NoError(t, pipeline.Settle(t.Context(), SettlementRequest{Reservation: reservation, BillingMode: platform.BillingModeExternalBilled}))
	require.Equal(t, 1, manager.committed, "a BYOK reservation is committed, not released")
	require.Zero(t, manager.released)
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

// sideUsageBudgetManager is a budget manager that also records side usage.
type sideUsageBudgetManager struct {
	creditBudgetManager
	recordedMC int64
	breakdown  platform.UsageCostBreakdown
}

func (m *sideUsageBudgetManager) RecordSideUsage(_ context.Context, _ *platform.AuthContext, breakdown platform.UsageCostBreakdown, amountMC int64) error {
	m.recordedMC += amountMC
	m.breakdown = breakdown

	return nil
}

func TestChargeSideUsageWritesAGuardrailSpendLog(t *testing.T) {
	manager := &sideUsageBudgetManager{}
	pipeline := NewRequestPipeline(nil, nil, &testBudgeter{}, BillingCapabilities{BudgetReservations: manager}, nil)
	logs := &testLogs{}
	auth := &platform.AuthContext{TenantID: "t", UserID: "u", TeamID: "team", TokenHash: "h"}

	usage := &guardrails.JudgeUsage{}
	pipeline.chargeJudgeUsage(t.Context(), logs, auth, "req-1", "trace-1", "gpt-test", "guardrail_input", usage)
	require.Empty(t, logs.loggedSpends, "no judge call, no charge")

	pipeline.ChargeSideUsage(t.Context(), logs, SideUsageCharge{
		Auth: auth, RequestID: "req-1", TraceID: "trace-1", Stage: "guardrail_input",
		Product: platform.CreditProductGuardrail, Model: "jev", RequestModel: "gpt-test", RequestType: "chat_completions",
		Calls: 2, PriceMC: credits.GuardrailJudgeCallMC,
	})

	require.Equal(t, int64(2_000), manager.recordedMC, "two judge calls cost two credits")
	require.Equal(t, "gpt-test", manager.breakdown.Model, "rules match the originating request")
	require.Len(t, logs.loggedSpends, 1)

	log := logs.loggedSpends[0]
	require.Equal(t, "req-1:guardrail_input", log.RequestID, "the side call never replaces the request's own spend log")
	require.Equal(t, "trace-1", log.RequestMetadata["trace_id"])
	require.Equal(t, "u", log.UserID)
	require.Equal(t, int64(2_000), log.Credits.Millicredits)
	require.Equal(t, platform.CreditProductGuardrail, log.Credits.Product)
	require.Zero(t, log.CostIRR, "side usage never touches IRR billing")
}
