package proxy

import (
	"context"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/billing"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

type staticBudgetStatus struct {
	err            error
	includeBalance bool
}

func (s *staticBudgetStatus) BudgetStatus(_ context.Context, _ *platform.AuthContext, includeBalance bool) (*platform.BudgetStatus, error) {
	s.includeBalance = includeBalance
	if s.err != nil {
		return nil, s.err
	}

	return &platform.BudgetStatus{Object: platform.BudgetStatusObject, Unit: platform.CreditUnit, Rules: []platform.BudgetRuleStatus{}}, nil
}

func budgetMeApp(statuses BudgetStatusReader, authInfo *platform.AuthContext) *fiber.App {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		if authInfo != nil {
			c.Locals(auth.AuthContextKey{}, authInfo)
		}

		return NewBudgetMeHandler(statuses)(c)
	})

	return app
}

func TestBudgetMeShowsTheBalanceOnlyToBillingReaders(t *testing.T) {
	cases := []struct {
		name string
		auth *platform.AuthContext
		want bool
	}{
		{name: "virtual key", auth: &platform.AuthContext{TenantID: "t", Role: "virtual_key"}, want: false},
		{name: "member", auth: &platform.AuthContext{TenantID: "t", Role: "member"}, want: false},
		{name: "billing reader", auth: &platform.AuthContext{TenantID: "t", Role: "member", Permissions: []string{"billing:read"}}, want: true},
		{name: "tenant admin", auth: &platform.AuthContext{TenantID: "t", Role: "tenant_admin"}, want: true},
	}

	for _, tc := range cases {
		statuses := &staticBudgetStatus{}
		status, body := decodeErrorBody(t, budgetMeApp(statuses, tc.auth))

		require.Equal(t, fiber.StatusOK, status, tc.name)
		require.Equal(t, platform.BudgetStatusObject, body["object"], tc.name)
		require.Equal(t, tc.want, statuses.includeBalance, tc.name)
	}
}

func TestBudgetMeErrors(t *testing.T) {
	status, _ := decodeErrorBody(t, budgetMeApp(&staticBudgetStatus{}, nil))
	require.Equal(t, fiber.StatusUnauthorized, status)

	status, body := decodeErrorBody(t, budgetMeApp(&staticBudgetStatus{err: billing.ErrCreditBillingRequired}, &platform.AuthContext{TenantID: "t"}))
	require.Equal(t, fiber.StatusConflict, status)
	require.Equal(t, "credit_billing_required", body["error"].(map[string]interface{})["code"])

	status, _ = decodeErrorBody(t, budgetMeApp(&staticBudgetStatus{err: billing.ErrBudgetUnavailable}, &platform.AuthContext{TenantID: "t"}))
	require.Equal(t, fiber.StatusServiceUnavailable, status)
}
