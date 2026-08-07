package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type entitlementResolverMock struct {
	effective *platform.EffectiveEntitlementsResponse
	userCount int
}

func (m entitlementResolverMock) ResolveEffectiveEntitlements(context.Context, string) (*platform.EffectiveEntitlementsResponse, error) {
	return m.effective, nil
}

func (m entitlementResolverMock) CountActiveTenantUsers(context.Context, string) (int, error) {
	return m.userCount, nil
}

func TestEntitlementGateRejectsDisabledVAT(t *testing.T) {
	h := &Handlers{billingState: billingState{entitlements: entitlementResolverMock{effective: activeEntitlements(map[string]interface{}{platform.FeatureVATEnabled: false})}}}
	app := fiber.New()
	app.Post("/vat", func(c *fiber.Ctx) error { return h.requireVATEnabled(c, "tenant-1") })

	resp, err := app.Test(httptest.NewRequest(fiber.MethodPost, "/vat", http.NoBody))
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestEntitlementGateAllowsTeamRateLimitRules(t *testing.T) {
	effective := activeEntitlements(map[string]interface{}{platform.FeatureRateLimitRulesEnabled: true})
	effective.Subscription.Status = "trialing"
	h := &Handlers{billingState: billingState{entitlements: entitlementResolverMock{effective: effective}}}
	app := fiber.New()
	app.Post("/rules", func(c *fiber.Ctx) error {
		if err := h.requireEntitlementEnabled(c, "tenant-1", platform.FeatureRateLimitRulesEnabled); err != nil {
			return err
		}

		return c.SendStatus(fiber.StatusNoContent)
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodPost, "/rules", http.NoBody))
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
}

func TestEntitlementLimitUsesEffectiveOverrideValue(t *testing.T) {
	h := &Handlers{billingState: billingState{entitlements: entitlementResolverMock{effective: activeEntitlements(map[string]interface{}{platform.FeatureUsersMax: float64(10)})}}}
	app := fiber.New()
	app.Post("/users/:used", func(c *fiber.Ctx) error {
		used := 9
		if c.Params("used") == "10" {
			used = 10
		}

		if err := h.requireEntitlementLimitAvailable(c, "tenant-1", platform.FeatureUsersMax, used, 1); err != nil {
			return err
		}

		return c.SendStatus(fiber.StatusNoContent)
	})

	allowed, err := app.Test(httptest.NewRequest(fiber.MethodPost, "/users/9", http.NoBody))
	defer closeResponseBody(t, allowed)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusNoContent, allowed.StatusCode)

	denied, err := app.Test(httptest.NewRequest(fiber.MethodPost, "/users/10", http.NoBody))
	defer closeResponseBody(t, denied)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, denied.StatusCode)
}

func activeEntitlements(features map[string]interface{}) *platform.EffectiveEntitlementsResponse {
	return &platform.EffectiveEntitlementsResponse{
		Subscription: &platform.TenantSubscription{Status: "active"},
		Features:     features,
	}
}
