package admin

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	billingdomain "github.com/Delpak/shahrah/services/control-plane/internal/domain/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

// walletTopUpFake implements only the wallet calls that top-ups make.
type walletTopUpFake struct {
	walletStore
	intent    *platform.WalletTopUpIntent
	created   int
	completed int
	failed    int
}

func (f *walletTopUpFake) GetLatestExchangeRate(context.Context) (*platform.ExchangeRate, error) {
	return &platform.ExchangeRate{USDtoIRR: 1000}, nil
}

func (f *walletTopUpFake) GetWalletTopUpSettings(context.Context) (*platform.WalletTopUpSettings, error) {
	return &platform.WalletTopUpSettings{MinAmountUSD: 1, MaxAmountUSD: 1000}, nil
}

func (f *walletTopUpFake) CreateWalletTopUpIntent(_ context.Context, p billingdomain.CreateTopUpIntentParams) (*platform.WalletTopUpIntent, error) {
	f.created++
	f.intent = &platform.WalletTopUpIntent{PublicID: "topup-1", Provider: p.Provider, Status: platform.WalletTopUpStatusPending, ReturnURL: p.ReturnURL}

	return f.intent, nil
}

func (f *walletTopUpFake) AttachWalletTopUpGatewayData(_ context.Context, _, authority, _ string, _ map[string]interface{}) error {
	f.intent.Authority = authority

	return nil
}

func (f *walletTopUpFake) GetWalletTopUpIntentByAuthority(_ context.Context, authority string) (*platform.WalletTopUpIntent, error) {
	if f.intent == nil || f.intent.Authority != authority {
		return nil, billingdomain.ErrTopUpIntentNotFound
	}

	return f.intent, nil
}

func (f *walletTopUpFake) CompleteWalletTopUpByAuthority(_ context.Context, _ string, _ map[string]interface{}, _, _, _, _ string) (*platform.WalletTopUpIntent, bool, error) {
	if f.intent.Status == platform.WalletTopUpStatusCompleted {
		return f.intent, true, nil
	}

	f.completed++
	f.intent.Status = platform.WalletTopUpStatusCompleted

	return f.intent, false, nil
}

func (f *walletTopUpFake) MarkWalletTopUpFailedByAuthority(context.Context, string, string, map[string]interface{}) error {
	f.failed++

	return nil
}

func walletTopUpApp(store *walletTopUpFake, payment PaymentConfig) *fiber.App {
	h := &Handlers{billingState: billingState{wallets: store, payment: payment.withDefaults()}}
	app := fiber.New()
	app.Get("/api/v1/wallet/topup/callback/:provider", h.WalletTopUpCallback)

	tenantRoutes := app.Group("/api/v1", func(c *fiber.Ctx) error {
		c.Locals(tenant.ContextKey{}, &platform.Tenant{ID: "tenant_a", Slug: "a", Status: "active"})

		return c.Next()
	})
	tenantRoutes.Post("/wallet/topup/intents", h.CreateWalletTopUpIntent)

	return app
}

func postWalletTopUp(t *testing.T, app *fiber.App, body string) int {
	t.Helper()

	req := httptest.NewRequest(fiber.MethodPost, "/api/v1/wallet/topup/intents", bytes.NewBufferString(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	return resp.StatusCode
}

// callWalletTopUpCallback returns the response headers, which carry the redirect target.
func callWalletTopUpCallback(t *testing.T, app *fiber.App, target string) http.Header {
	t.Helper()

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, target, http.NoBody))
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	return resp.Header
}

func TestWalletTopUpRefusesUnavailableProviderBeforeCreatingIntent(t *testing.T) {
	store := &walletTopUpFake{}
	app := walletTopUpApp(store, PaymentConfig{Production: true})

	status := postWalletTopUp(t, app, `{"amount_usd": 10}`)

	require.NotEqual(t, http.StatusCreated, status)
	require.Zero(t, store.created, "no pending intent may be written for a refused provider")
}

func TestWalletTopUpProductionCallbackNeverCreditsMock(t *testing.T) {
	store := &walletTopUpFake{intent: &platform.WalletTopUpIntent{
		Provider: "mock", Authority: "mock_abc", Status: platform.WalletTopUpStatusPending, ReturnURL: "http://example.com/billing",
	}}
	app := walletTopUpApp(store, PaymentConfig{Production: true})

	callWalletTopUpCallback(t, app, "/api/v1/wallet/topup/callback/mock?authority=mock_abc&status=ok")

	require.Zero(t, store.completed)
	require.Equal(t, 1, store.failed)
}

func TestWalletTopUpLocalCallbackCreditsOnce(t *testing.T) {
	store := &walletTopUpFake{intent: &platform.WalletTopUpIntent{
		Provider: "mock", Authority: "mock_abc", Status: platform.WalletTopUpStatusPending, ReturnURL: "http://example.com/billing",
	}}
	app := walletTopUpApp(store, PaymentConfig{})

	first := callWalletTopUpCallback(t, app, "/api/v1/wallet/topup/callback/mock?authority=mock_abc&status=ok")
	second := callWalletTopUpCallback(t, app, "/api/v1/wallet/topup/callback/mock?authority=mock_abc&status=ok")

	require.Equal(t, 1, store.completed, "a repeated callback must not credit the wallet again")
	require.Contains(t, first.Get("Location"), "status=success")
	require.Contains(t, second.Get("Location"), "status=success")
}

func TestWalletTopUpLateFailureCallbackKeepsCompletedTopUp(t *testing.T) {
	store := &walletTopUpFake{intent: &platform.WalletTopUpIntent{
		Provider: "mock", Authority: "mock_abc", Status: platform.WalletTopUpStatusCompleted, ReturnURL: "http://example.com/billing",
	}}
	app := walletTopUpApp(store, PaymentConfig{})

	resp := callWalletTopUpCallback(t, app, "/api/v1/wallet/topup/callback/mock?authority=mock_abc&status=failed")

	require.Zero(t, store.failed)
	require.Contains(t, resp.Get("Location"), "status=success")
}

func TestPaymentReturnURLMustBeOnThisSite(t *testing.T) {
	wallets := &walletTopUpFake{}
	app := walletTopUpApp(wallets, PaymentConfig{})

	require.Equal(t, http.StatusBadRequest, postWalletTopUp(t, app, `{"amount_usd": 10, "return_url": "https://evil.example/steal"}`))
	require.Zero(t, wallets.created)

	creditStore := &creditPackStoreFake{}
	creditApp := creditPackApp(creditStore, "zarinpal")

	req := httptest.NewRequest(fiber.MethodPost, "/api/v1/billing/credits/orders",
		bytes.NewBufferString(`{"pack_slug":"credits-5000","provider":"zarinpal","return_url":"https://evil.example/steal"}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := creditApp.Test(req)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	require.Empty(t, creditStore.gotProvider)
}
