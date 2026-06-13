package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"strings"
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
	resp      *platform.TunnelResponse
	stream    platform.TunnelStream
	streamErr error
	lastReq   platform.TunnelRequest
}

func (t *captureTunnel) Proxy(ctx context.Context, req platform.TunnelRequest) (*platform.TunnelResponse, error) {
	t.lastReq = req
	return t.resp, nil
}

func (t *captureTunnel) Stream(ctx context.Context, req platform.TunnelRequest) (platform.TunnelStream, error) {
	t.lastReq = req
	if t.stream != nil || t.streamErr != nil {
		return t.stream, t.streamErr
	}
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

func TestResponsesHandlerCanonicalizesProviderPrefixedDeploymentModel(t *testing.T) {
	mem, err := cache.NewMemoryCache(zap.NewNop())
	require.NoError(t, err)
	rt, err := router.NewRouter([]config.ModelDeployment{{
		ModelName: "gpt-4o-mini",
		ProviderEndpoint: config.ProviderEndpointConfig{
			Provider: "openai",
			Model:    "openai/gpt-4o-mini",
			APIKey:   "test-key",
		},
	}}, config.RouterSettings{CooldownTime: 60, NumRetries: 1}, mem, nil, zap.NewNop())
	require.NoError(t, err)

	payload := contractsopenai.Response{
		ID:        "resp_2",
		Object:    "response",
		CreatedAt: time.Now().Unix(),
		Model:     "gpt-4o-mini",
		Status:    "completed",
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       body,
	}}
	h := NewResponsesHandler(rt, tunnel, &testBudgeter{}, &testWallets{}, &testLogs{}, zap.NewNop())

	app := fiber.New()
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

	req := httptest.NewRequest("POST", "/v1/responses", bytes.NewBufferString(`{"model":"gpt-4o-mini","input":"hello"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var outbound contractsopenai.ResponseRequest
	err = json.Unmarshal(tunnel.lastReq.Body, &outbound)
	require.NoError(t, err)
	assert.Equal(t, "gpt-4o-mini", outbound.Model)
}

func TestResponsesHandlerStreamSuccess(t *testing.T) {
	stream := newFakeTunnelStream(map[string]string{"content-type": "text/event-stream"}, 2, 0)
	stream.chunks <- platform.TunnelChunk{Data: []byte("event: response.created\n")}
	stream.chunks <- platform.TunnelChunk{Data: []byte("data: {\"type\":\"response.output_text.delta\",\"delta\":\"Salam\"}\n\n"), IsDone: true}
	close(stream.chunks)

	tunnel := &captureTunnel{stream: stream}
	app, _ := newResponsesTestApp(t, "openai", tunnel)

	req := httptest.NewRequest("POST", "/v1/responses", bytes.NewBufferString(`{"model":"gpt-4o","input":"hello","stream":true}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, int((5 * time.Second).Milliseconds()))
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))
	assert.Equal(t, "/responses", tunnel.lastReq.RequestMetadata["provider_endpoint_path"])

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	bodyText := string(body)
	assert.Contains(t, bodyText, "event: response.created")
	assert.Contains(t, bodyText, `response.output_text.delta`)
	assert.True(t, stream.closed)
}

func TestResponsesHandlerStreamErrorWritesStreamingErrorEvent(t *testing.T) {
	stream := newFakeTunnelStream(map[string]string{"content-type": "text/event-stream"}, 0, 1)
	stream.errors <- errors.New("responses stream failed")

	tunnel := &captureTunnel{stream: stream}
	app, _ := newResponsesTestApp(t, "openai", tunnel)

	req := httptest.NewRequest("POST", "/v1/responses", bytes.NewBufferString(`{"model":"gpt-4o","input":"hello","stream":true}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, int((5 * time.Second).Milliseconds()))
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), `"type":"streaming_error"`)
	assert.True(t, stream.closed)
}

func TestResponsesHandlerStreamCloseFinalizesWithoutPanic(t *testing.T) {
	stream := newFakeTunnelStream(map[string]string{"content-type": "text/event-stream"}, 1, 0)
	stream.chunks <- platform.TunnelChunk{Data: []byte("data: {\"type\":\"response.output_text.delta\",\"delta\":\"Salam\"}\n\n")}
	close(stream.chunks)

	wallets := &testWallets{}
	logs := &testLogs{}

	mem, err := cache.NewMemoryCache(zap.NewNop())
	require.NoError(t, err)
	rt, err := router.NewRouter([]config.ModelDeployment{{
		ModelName: "gpt-4o",
		ProviderEndpoint: config.ProviderEndpointConfig{
			Provider: "openai",
			Model:    "gpt-4o",
			APIKey:   "test-key",
		},
	}}, config.RouterSettings{CooldownTime: 60, NumRetries: 1}, mem, nil, zap.NewNop())
	require.NoError(t, err)

	h := NewResponsesHandler(rt, &captureTunnel{stream: stream}, &testBudgeter{}, wallets, logs, zap.NewNop())
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(auth.AuthContextKey{}, &platform.AuthContext{
			TenantID:  "tenant_a",
			KeyID:     "key_1",
			KeyAlias:  "test-key",
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

	req := httptest.NewRequest("POST", "/v1/responses", bytes.NewBufferString(`{"model":"gpt-4o","input":"hello","stream":true}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, int((5 * time.Second).Milliseconds()))
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), `response.output_text.delta`)
	assert.NotContains(t, strings.ToLower(string(body)), "panic")
	assert.Len(t, wallets.debits, 1)
	assert.Len(t, logs.loggedSpends, 1)
	assert.True(t, stream.closed)
	assert.Equal(t, "/responses", logs.loggedSpends[0].RequestMetadata["provider_endpoint_path"])
}
