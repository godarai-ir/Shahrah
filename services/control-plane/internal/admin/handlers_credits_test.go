package admin

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

type creditUsageStoreMock struct {
	tenantID   string
	start, end time.Time
}

func (m *creditUsageStoreMock) GetCreditUsageSummary(
	_ context.Context,
	tenantID string,
	start time.Time,
	end time.Time,
) (*platform.CreditUsageSummary, error) {
	m.tenantID, m.start, m.end = tenantID, start, end

	return &platform.CreditUsageSummary{
		Available:    true,
		ShadowMode:   true,
		TotalCredits: "12.50",
		ByModel:      []platform.CreditUsageItem{{Key: "anthropic/claude-sonnet-5", Credits: "12.50", RequestCount: 3}},
		ByTeam:       []platform.CreditUsageItem{},
		ByProduct:    []platform.CreditUsageItem{},
	}, nil
}

func creditSummaryApp(h *Handlers, permissions []string) *fiber.App {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(auth.AuthContextKey{}, &platform.AuthContext{Role: "member", Permissions: permissions})
		c.Locals(tenant.ContextKey{}, &platform.Tenant{ID: "tenant_a", Slug: "a", Status: "active"})

		return c.Next()
	})
	app.Get("/api/v1/billing/credits/summary", h.GetCreditUsageSummary)

	return app
}

func getCreditSummary(t *testing.T, app *fiber.App, query string) (int, platform.CreditUsageSummary) {
	t.Helper()

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/api/v1/billing/credits/summary"+query, http.NoBody))
	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	var summary platform.CreditUsageSummary
	if resp.StatusCode == fiber.StatusOK {
		require.NoError(t, json.Unmarshal(body, &summary))
	}

	return resp.StatusCode, summary
}

func TestGetCreditUsageSummaryUsesTenantAndPreset(t *testing.T) {
	t.Parallel()

	store := &creditUsageStoreMock{}
	h := &Handlers{billingState: billingState{creditUsage: store}}

	status, summary := getCreditSummary(t, creditSummaryApp(h, []string{"billing:read"}), "?preset=last_3_months")
	require.Equal(t, fiber.StatusOK, status)
	require.Equal(t, "tenant_a", store.tenantID)
	require.True(t, store.start.Before(store.end))
	require.Equal(t, "last_3_months", summary.Preset)
	require.Equal(t, "12.50", summary.TotalCredits)
	require.True(t, summary.Available)
}

func TestGetCreditUsageSummaryWithoutUsageStore(t *testing.T) {
	t.Parallel()

	h := &Handlers{}

	status, summary := getCreditSummary(t, creditSummaryApp(h, []string{"billing:read"}), "")
	require.Equal(t, fiber.StatusOK, status)
	require.False(t, summary.Available)
	require.Equal(t, "current_month", summary.Preset)
	require.Equal(t, "0.00", summary.TotalCredits)
}

func TestGetCreditUsageSummaryRejectsInvalidPresetAndMissingPermission(t *testing.T) {
	t.Parallel()

	h := &Handlers{billingState: billingState{creditUsage: &creditUsageStoreMock{}}}

	status, _ := getCreditSummary(t, creditSummaryApp(h, []string{"billing:read"}), "?preset=forever")
	require.Equal(t, fiber.StatusBadRequest, status)

	status, _ = getCreditSummary(t, creditSummaryApp(h, nil), "")
	require.Equal(t, fiber.StatusForbidden, status)
}
