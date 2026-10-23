package admin

import (
	"context"
	"encoding/json"
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
}

func (m entitlementResolverMock) ResolveEffectiveEntitlements(context.Context, string) (*platform.EffectiveEntitlementsResponse, error) {
	return m.effective, nil
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

func TestActiveSubscriptionGateRejectsLapsedPlans(t *testing.T) {
	cases := map[string]int{"active": fiber.StatusNoContent, "trialing": fiber.StatusNoContent, "expired": fiber.StatusForbidden, "past_due": fiber.StatusForbidden}

	for status, want := range cases {
		effective := activeEntitlements(nil)
		effective.Subscription.Status = status
		h := &Handlers{billingState: billingState{entitlements: entitlementResolverMock{effective: effective}}}
		app := fiber.New()
		app.Post("/invite", func(c *fiber.Ctx) error {
			if err := h.requireActiveSubscription(c, "tenant-1"); err != nil {
				return err
			}

			return c.SendStatus(fiber.StatusNoContent)
		})

		resp, err := app.Test(httptest.NewRequest(fiber.MethodPost, "/invite", http.NoBody))
		require.NoError(t, err)
		assert.Equal(t, want, resp.StatusCode, status)
		closeResponseBody(t, resp)
	}
}

func TestPlanUserLimitReachedIsClientSafe(t *testing.T) {
	app := fiber.New()
	app.Post("/invite", planUserLimitReached)

	resp, err := app.Test(httptest.NewRequest(fiber.MethodPost, "/invite", http.NoBody))
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)

	var body struct {
		Error map[string]string `json:"error"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, "plan_user_limit_reached", body.Error["code"])
	assert.NotRegexp(t, `\d`, body.Error["message"], "no counts or limits leak to the client")
}

func activeEntitlements(features map[string]interface{}) *platform.EffectiveEntitlementsResponse {
	return &platform.EffectiveEntitlementsResponse{
		Subscription: &platform.TenantSubscription{Status: "active"},
		Features:     features,
	}
}
