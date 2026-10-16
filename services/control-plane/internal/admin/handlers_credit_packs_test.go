package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	billingdomain "github.com/Delpak/shahrah/services/control-plane/internal/domain/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type creditPackStoreFake struct {
	createErr   error
	order       *platform.CreditPackOrder
	attached    string
	completed   int
	failed      string
	gotProvider string
}

func (f *creditPackStoreFake) ListCreditPacks(context.Context) ([]platform.CreditPack, error) {
	return []platform.CreditPack{{Slug: "credits-5000", Credits: 5000, PriceIRR: 45_000_000, PayableIRR: 45_000_000}}, nil
}

func (f *creditPackStoreFake) CreateCreditPackOrder(_ context.Context, tenantID string, req platform.CreateCreditPackOrderRequest, callbackURL, returnURL, _ string) (*platform.CreditPackOrder, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}

	f.gotProvider = req.Provider
	status := platform.CreditPackOrderPending

	if req.Provider == platform.CreditPackProviderBankTransfer {
		status = platform.CreditPackOrderAwaitingTransfer
	}

	f.order = &platform.CreditPackOrder{
		ID: "order-1", PublicID: "public-1", TenantID: tenantID, PackSlug: req.PackSlug, Credits: 5000,
		PayableIRR: 45_000_000, Provider: req.Provider, Status: status, CallbackURL: callbackURL, ReturnURL: returnURL,
	}

	return f.order, nil
}

func (f *creditPackStoreFake) AttachCreditPackOrderGatewayData(_ context.Context, _, authority, _ string, _ map[string]interface{}) error {
	f.attached = authority
	f.order.Authority = authority

	return nil
}

func (f *creditPackStoreFake) GetCreditPackOrder(context.Context, string, string) (*platform.CreditPackOrder, error) {
	return f.order, nil
}

func (f *creditPackStoreFake) GetCreditPackOrderByAuthority(_ context.Context, authority string) (*platform.CreditPackOrder, error) {
	if f.order == nil || f.order.Authority != authority {
		return nil, billingdomain.ErrCreditPackOrderNotFound
	}

	return f.order, nil
}

func (f *creditPackStoreFake) ListCreditPackOrders(context.Context, string, int) ([]platform.CreditPackOrder, error) {
	return []platform.CreditPackOrder{}, nil
}

func (f *creditPackStoreFake) CompleteCreditPackOrderByAuthority(_ context.Context, _ string, _ billingdomain.PaymentVerification) (*platform.CreditPackOrder, bool, error) {
	f.completed++
	f.order.Status = platform.CreditPackOrderCompleted

	return f.order, false, nil
}

func (f *creditPackStoreFake) MarkCreditPackOrderFailedByAuthority(_ context.Context, _, reason string, _ map[string]interface{}) error {
	f.failed = reason

	return nil
}

func (f *creditPackStoreFake) GetCreditSettings(context.Context, string) (*platform.CreditSettings, error) {
	return &platform.CreditSettings{BillingUnit: "credit"}, nil
}

func (f *creditPackStoreFake) UpdateCreditSettings(context.Context, string, platform.UpdateCreditSettingsRequest, platform.SupportActor) (*platform.CreditSettings, error) {
	return nil, billingdomain.ErrOverageNotAllowed
}

func creditPackApp(store *creditPackStoreFake, defaultProvider string) *fiber.App {
	h := &Handlers{billingState: billingState{creditPacks: store, payment: PaymentConfig{DefaultSubscriptionProvider: defaultProvider}.withDefaults()}}
	app := fiber.New()
	app.Get("/api/v1/credits/packs/payment/callback/:provider", h.CreditPackPaymentCallback)

	tenantRoutes := app.Group("/api/v1", func(c *fiber.Ctx) error {
		c.Locals(auth.AuthContextKey{}, &platform.AuthContext{Role: "member", UserID: "user-1", Permissions: []string{"billing:manage"}})
		c.Locals(tenant.ContextKey{}, &platform.Tenant{ID: "tenant_a", Slug: "a", Status: "active"})

		return c.Next()
	})
	tenantRoutes.Post("/billing/credits/orders", h.CreateCreditPackOrder)
	tenantRoutes.Put("/billing/credits/settings", h.UpdateCreditSettings)

	return app
}

func postCreditPackOrder(t *testing.T, app *fiber.App, body string) (int, map[string]interface{}) {
	t.Helper()

	req := httptest.NewRequest(fiber.MethodPost, "/api/v1/billing/credits/orders", bytes.NewBufferString(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := app.Test(req)
	require.NoError(t, err)

	var payload map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&payload))
	require.NoError(t, resp.Body.Close())

	return resp.StatusCode, payload
}

func TestCreditPackOrderAndCallbackComplete(t *testing.T) {
	store := &creditPackStoreFake{}
	app := creditPackApp(store, "mock")

	status, payload := postCreditPackOrder(t, app, `{"pack_slug":"credits-5000"}`)
	require.Equal(t, fiber.StatusCreated, status)
	require.Equal(t, "mock", store.gotProvider, "falls back to the configured default provider")
	require.NotEmpty(t, store.attached)

	paymentURL, ok := payload["payment_url"].(string)
	require.True(t, ok)
	require.Contains(t, paymentURL, creditPackCallbackPath+"mock")

	// The URL names a different provider; the order's own provider verifies.
	callback := httptest.NewRequest(fiber.MethodGet, creditPackCallbackPath+"zarinpal?authority="+url.QueryEscape(store.attached), http.NoBody)
	callbackResp, err := app.Test(callback)
	require.NoError(t, err)
	require.NoError(t, callbackResp.Body.Close())
	require.Equal(t, fiber.StatusFound, callbackResp.StatusCode)
	require.Equal(t, 1, store.completed)
	assert.Contains(t, callbackResp.Header.Get("Location"), "payment_status=success")
}

func TestCreditPackCallbackRecordsFailedPayments(t *testing.T) {
	store := &creditPackStoreFake{}
	app := creditPackApp(store, "mock")

	status, _ := postCreditPackOrder(t, app, `{"pack_slug":"credits-5000","provider":"mock"}`)
	require.Equal(t, fiber.StatusCreated, status)

	callback := httptest.NewRequest(fiber.MethodGet, creditPackCallbackPath+"mock?status=canceled&authority="+url.QueryEscape(store.attached), http.NoBody)
	callbackResp, err := app.Test(callback)
	require.NoError(t, err)
	require.NoError(t, callbackResp.Body.Close())
	require.Equal(t, fiber.StatusFound, callbackResp.StatusCode)
	require.Zero(t, store.completed)
	require.NotEmpty(t, store.failed)
	assert.Contains(t, callbackResp.Header.Get("Location"), "payment_status=failed")
}

func TestCreditPackOrderRefusesTheMockGatewayOutsideMockEnvironments(t *testing.T) {
	store := &creditPackStoreFake{}
	app := creditPackApp(store, "zarinpal")

	for _, provider := range []string{"mock", "unknown"} {
		status, payload := postCreditPackOrder(t, app, `{"pack_slug":"credits-5000","provider":"`+provider+`"}`)
		require.Equal(t, fiber.StatusUnprocessableEntity, status, provider)
		require.Equal(t, "payment_provider_unavailable", payload["error"].(map[string]interface{})["code"], provider)
	}

	status, payload := postCreditPackOrder(t, app, `{"pack_slug":"credits-5000","provider":"bank_transfer"}`)
	require.Equal(t, fiber.StatusCreated, status)
	require.Equal(t, platform.CreditPackOrderAwaitingTransfer, payload["status"])
	require.Empty(t, store.attached, "bank transfers never reach a gateway")
}

func TestCreditPackErrorsAreCoded(t *testing.T) {
	store := &creditPackStoreFake{createErr: billingdomain.ErrCreditBillingNotEnabled}
	app := creditPackApp(store, "mock")

	status, payload := postCreditPackOrder(t, app, `{"pack_slug":"credits-5000"}`)
	require.Equal(t, fiber.StatusConflict, status)
	require.Equal(t, "credit_billing_not_enabled", payload["error"].(map[string]interface{})["code"])

	req := httptest.NewRequest(fiber.MethodPut, "/api/v1/billing/credits/settings", bytes.NewBufferString(`{"overage_enabled":true}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	settingsResp, err := app.Test(req, int((5 * time.Second).Milliseconds()))
	require.NoError(t, err)
	require.NoError(t, settingsResp.Body.Close())
	require.Equal(t, fiber.StatusForbidden, settingsResp.StatusCode)
}
