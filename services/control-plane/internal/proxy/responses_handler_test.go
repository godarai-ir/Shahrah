package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/config"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type captureTunnel struct {
	resp    *platform.TunnelResponse
	lastReq platform.TunnelRequest
}

func (t *captureTunnel) Proxy(ctx context.Context, req platform.TunnelRequest) (*platform.TunnelResponse, error) {
	t.lastReq = req
	return t.resp, nil
}

func (t *captureTunnel) Stream(ctx context.Context, req platform.TunnelRequest) (platform.TunnelStream, error) {
	t.lastReq = req
	return nil, assert.AnError
}

func (t *captureTunnel) Healthy(ctx context.Context) error { return nil }

func newResponsesTestApp(t *testing.T, provider string, tunnel *captureTunnel) (*fiber.App, *tracing.Queue) {
	t.Helper()

	mem, err := cache.NewMemoryCache(zap.NewNop())
	require.NoError(t, err)
	rt, err := router.NewRouter([]config.ModelDeployment{{
		ModelName: "gpt-4o",
		ProviderEndpoint: config.ProviderEndpointConfig{
			Provider: provider,
			Model:    "gpt-4o",
			APIKey:   "test-key",
		},
	}}, config.RouterSettings{CooldownTime: 60, NumRetries: 1}, mem, nil, zap.NewNop())
	require.NoError(t, err)

	h := NewResponsesHandler(rt, tunnel, &testBudgeter{}, &testWallets{}, &testLogs{}, zap.NewNop())
	q := tracing.NewQueue(64)
	s := tracing.NewSampler(1.0, 1500)

	app := fiber.New()
	app.Use(tracing.Middleware(q, s, 1500))
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(auth.AuthContextKey{}, &platform.AuthContext{
			TenantID:  "tenant_a",
			KeyID:     "key_1",
			KeyType:   "pat",
			Role:      "team_admin",
			TokenHash: "hash_1",
		})
		c.Locals(tenant.ContextKey{}, &platform.Tenant{
			ID:     "tenant_a",
			Slug:   "a",
			Status: "active",
		})
		return c.Next()
	})
	app.Post("/v1/responses", h.Handle)
	return app, q
}

func TestResponsesHandlerSyncSuccess(t *testing.T) {
	payload := contractsopenai.Response{
		ID:        "resp_1",
		Object:    "response",
		CreatedAt: time.Now().Unix(),
		Model:     "gpt-4o",
		Status:    "completed",
		Output: []contractsopenai.ResponseOutput{{
			ID:     "msg_1",
			Type:   "message",
			Role:   "assistant",
			Status: "completed",
			Content: []contractsopenai.ResponseOutputContent{{
				Type: "output_text",
				Text: "hi",
			}},
		}},
		Usage: &contractsopenai.ResponseUsage{
			InputTokens:  3,
			OutputTokens: 2,
			TotalTokens:  5,
		},
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       body,
	}}
	app, q := newResponsesTestApp(t, "openai", tunnel)

	req := httptest.NewRequest("POST", "/v1/responses", bytes.NewBufferString(`{"model":"gpt-4o","input":"hello"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, "/responses", tunnel.lastReq.RequestMetadata["provider_endpoint_path"])

	events := q.DequeueBatch(20)
	names := make([]string, 0, len(events))
	for _, ev := range events {
		names = append(names, ev.Name)
	}
	assert.Contains(t, names, "gateway.request")
	assert.Contains(t, names, "provider.call")
	assert.Contains(t, names, "response.write")
}

func TestResponsesHandlerRejectsNonOpenAIProvider(t *testing.T) {
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       []byte(`{}`),
	}}
	app, _ := newResponsesTestApp(t, "anthropic", tunnel)

	req := httptest.NewRequest("POST", "/v1/responses", bytes.NewBufferString(`{"model":"gpt-4o","input":"hello"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusNotImplemented, resp.StatusCode)
}
