package proxy

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/config"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/Delpak/shahrah/services/control-plane/internal/testcache"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestXAIDeferredCompletionLifecycleIsOwnedAndExactlyOnce(t *testing.T) {
	resources := &responseOperationResources{}
	deployment := platform.Deployment{
		ID: "dep_xai", TenantID: "tenant_a", ModelName: "grok-4.5", Provider: "xai", Model: "grok-4.5",
		APIKey: "test-key", ProviderAccountID: "xai_acct_1", BillingMode: platform.BillingModeWalletBilled,
	}
	deployments := responseOperationDeployments{deployments: []platform.Deployment{deployment}}
	mem, err := testcache.New(t, zap.NewNop())
	require.NoError(t, err)
	rt, err := router.NewRouter(nil, config.RouterSettings{CooldownTime: 60, NumRetries: 1}, mem, deployments, zap.NewNop())
	require.NoError(t, err)

	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       []byte(`{"request_id":"deferred_1"}`),
	}}
	handler := NewChatCompletionHandlerWithDependencies(ChatCompletionDependencies{Base: BaseHandlerDependencies{
		Router: rt, Tunnel: tunnel, Billing: &xaiResponsesBudgeter{}, Wallets: &testWallets{},
		ProviderResources: resources, Deployments: deployments, RequestLogger: &testLogs{}, SpendLogger: &testLogs{}, Logger: zap.NewNop(),
	}})

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(auth.AuthContextKey{}, &platform.AuthContext{TenantID: "tenant_a", KeyID: "key_1", TokenHash: "hash_1", Role: "team_admin"})
		c.Locals(tenant.ContextKey{}, &platform.Tenant{ID: "tenant_a", Slug: "a", Status: "active"})

		return c.Next()
	})
	app.Post("/v1/chat/completions", handler.Handle)
	app.Get("/v1/chat/deferred-completion/:request_id", handler.HandleXAIDeferredGet)

	create := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(`{"model":"grok-4.5","messages":[{"role":"user","content":"hello"}],"deferred":true}`))
	create.Header.Set("Content-Type", "application/json")

	createResp, err := app.Test(create)
	defer closeResponseBody(t, createResp)

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, createResp.StatusCode)

	mapping, ok := resources.mappings["deferred_1"]
	require.True(t, ok)
	assert.Equal(t, "xai", mapping.Provider)
	assert.Equal(t, "xai_acct_1", mapping.ProviderAccountID)
	assert.Equal(t, "pending", mapping.Status)
	require.NotNil(t, mapping.ExpiresAt)

	tunnel.resp = &platform.TunnelResponse{StatusCode: http.StatusAccepted, Body: nil}
	pending := httptest.NewRequest(http.MethodGet, "/v1/chat/deferred-completion/deferred_1", http.NoBody)

	pendingResp, err := app.Test(pending)
	defer closeResponseBody(t, pendingResp)

	require.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, pendingResp.StatusCode)
	assert.Equal(t, platform.ProviderOperationXAIDeferredChatGet, tunnel.lastReq.Operation)
	assert.Equal(t, "xai_acct_1", tunnel.lastReq.Deployment.ProviderAccountID)

	tunnel.resp = &platform.TunnelResponse{
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       []byte(`{"id":"chatcmpl_1","object":"chat.completion","model":"grok-4.5","choices":[],"usage":{"prompt_tokens":3,"completion_tokens":2,"total_tokens":5,"cost_in_usd_ticks":10000000}}`),
	}
	ready := httptest.NewRequest(http.MethodGet, "/v1/chat/deferred-completion/deferred_1", http.NoBody)

	readyResp, err := app.Test(ready)
	defer closeResponseBody(t, readyResp)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, readyResp.StatusCode)
	assert.Equal(t, "retrieved", resources.mappings["deferred_1"].Status)

	again := httptest.NewRequest(http.MethodGet, "/v1/chat/deferred-completion/deferred_1", http.NoBody)

	againResp, err := app.Test(again)
	defer closeResponseBody(t, againResp)

	require.NoError(t, err)
	assert.Equal(t, http.StatusGone, againResp.StatusCode)
}
