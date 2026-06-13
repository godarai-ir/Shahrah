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

type testBudgeter struct{}

func (b *testBudgeter) EstimateCost(ctx context.Context, req contractsopenai.ChatCompletionRequest) (platform.UsageCostBreakdown, error) {
	return platform.UsageCostBreakdown{Model: req.Model, TotalTokens: 32, InputTokens: 16, OutputTokens: 16, TotalCostUSD: 0.001, TotalCostIRR: 850}, nil
}
func (b *testBudgeter) CheckRateLimit(ctx context.Context, auth *platform.AuthContext, model string, estimatedTokens int) (*cache.RateLimitResult, error) {
	return &cache.RateLimitResult{Allowed: true, Current: 1, Limit: 100, ResetAt: time.Now().Add(time.Minute)}, nil
}
func (b *testBudgeter) CheckBudget(ctx context.Context, auth *platform.AuthContext, breakdown platform.UsageCostBreakdown) error {
	return nil
}
func (b *testBudgeter) FinalizeFromUsage(ctx context.Context, model string, usage *contractsopenai.Usage, estimated platform.UsageCostBreakdown) (platform.UsageCostBreakdown, error) {
	return estimated, nil
}

type walletDebitCall struct {
	TenantID  string
	Breakdown platform.UsageCostBreakdown
	RequestID string
}

type testWallets struct {
	debits []walletDebitCall
}

func (w *testWallets) DebitWallet(ctx context.Context, tenantID string, breakdown platform.UsageCostBreakdown, requestID string) (*platform.Transaction, error) {
	w.debits = append(w.debits, walletDebitCall{
		TenantID:  tenantID,
		Breakdown: breakdown,
		RequestID: requestID,
	})
	return &platform.Transaction{}, nil
}

type testLogs struct {
	loggedEntries []platform.RequestLogEntry
	loggedSpends  []platform.SpendLog
}

func (l *testLogs) LogRequest(ctx context.Context, entry platform.RequestLogEntry) error {
	l.loggedEntries = append(l.loggedEntries, entry)
	return nil
}
func (l *testLogs) LogSpend(ctx context.Context, log platform.SpendLog) error {
	l.loggedSpends = append(l.loggedSpends, log)
	return nil
}

type fakeTunnelStream struct {
	headers map[string]string
	chunks  chan platform.TunnelChunk
	errors  chan error
	closed  bool
}

func newFakeTunnelStream(headers map[string]string, chunkBuf int, errorBuf int) *fakeTunnelStream {
	if headers == nil {
		headers = map[string]string{}
	}
	if chunkBuf < 0 {
		chunkBuf = 0
	}
	if errorBuf < 0 {
		errorBuf = 0
	}
	return &fakeTunnelStream{
		headers: headers,
		chunks:  make(chan platform.TunnelChunk, chunkBuf),
		errors:  make(chan error, errorBuf),
	}
}

func (s *fakeTunnelStream) Chunks() <-chan platform.TunnelChunk { return s.chunks }
func (s *fakeTunnelStream) Errors() <-chan error                { return s.errors }
func (s *fakeTunnelStream) Headers() map[string]string          { return s.headers }
func (s *fakeTunnelStream) Close() error {
	s.closed = true
	return nil
}

type testTunnel struct {
	resp   *platform.TunnelResponse
	err    error
	stream platform.TunnelStream
}

func (t *testTunnel) Proxy(ctx context.Context, req platform.TunnelRequest) (*platform.TunnelResponse, error) {
	return t.resp, t.err
}
func (t *testTunnel) Stream(ctx context.Context, req platform.TunnelRequest) (platform.TunnelStream, error) {
	if t.stream != nil {
		return t.stream, nil
	}
	return nil, assert.AnError
}
func (t *testTunnel) Healthy(ctx context.Context) error { return nil }

func newCompletionsTestApp(t *testing.T, tunnelResp *platform.TunnelResponse) (*fiber.App, *tracing.Queue) {
	t.Helper()

	mem, err := cache.NewMemoryCache(zap.NewNop())
	require.NoError(t, err)
	rt, err := router.NewRouter([]config.ModelDeployment{{
		ModelName: "gpt-4",
		ProviderEndpoint: config.ProviderEndpointConfig{
			Provider: "openai",
			Model:    "gpt-4",
			APIKey:   "test-key",
		},
	}}, config.RouterSettings{CooldownTime: 60, NumRetries: 1}, mem, nil, zap.NewNop())
	require.NoError(t, err)

	h := NewCompletionsHandler(rt, &testTunnel{resp: tunnelResp}, &testBudgeter{}, &testWallets{}, &testLogs{}, zap.NewNop())
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
	app.Post("/v1/completions", h.Handle)
	return app, q
}

func TestCompletionsTracingLifecycle_SyncSuccess(t *testing.T) {
	payload := contractsopenai.ChatCompletionResponse{
		ID:      "chatcmpl_1",
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   "gpt-4",
		Choices: []contractsopenai.Choice{
			{Index: 0, Message: &contractsopenai.ChatMessage{Role: "assistant", Content: "hi"}},
		},
		Usage: &contractsopenai.Usage{PromptTokens: 3, CompletionTokens: 2, TotalTokens: 5},
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)
	app, q := newCompletionsTestApp(t, &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       body,
	})

	req := httptest.NewRequest("POST", "/v1/completions", bytes.NewBufferString(`{"model":"gpt-4","prompt":"hello"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	events := q.DequeueBatch(20)
	names := make([]string, 0, len(events))
	for _, ev := range events {
		names = append(names, ev.Name)
	}
	assert.Contains(t, names, "gateway.request")
	assert.Contains(t, names, "auth.validate")
	assert.Contains(t, names, "quota.check")
	assert.Contains(t, names, "router.select_model")
	assert.Contains(t, names, "provider.call")
	assert.Contains(t, names, "response.write")
}

func TestCompletionsTracingLifecycle_ProviderErrorIncludesProviderSpan(t *testing.T) {
	app, q := newCompletionsTestApp(t, &platform.TunnelResponse{
		StatusCode: fiber.StatusBadGateway,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       []byte(`{"error":{"message":"upstream failed","type":"api_error"}}`),
	})

	req := httptest.NewRequest("POST", "/v1/completions", bytes.NewBufferString(`{"model":"gpt-4","prompt":"hello"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadGateway, resp.StatusCode)

	events := q.DequeueBatch(20)
	var sawProvider bool
	for _, ev := range events {
		if ev.Name == "provider.call" {
			sawProvider = true
			assert.Equal(t, fiber.StatusBadGateway, ev.StatusCode)
		}
	}
	assert.True(t, sawProvider)
}

func newCompletionsStreamingApp(t *testing.T, tunnel platform.TunnelTransport, wallets *testWallets, logs *testLogs) *fiber.App {
	t.Helper()

	mem, err := cache.NewMemoryCache(zap.NewNop())
	require.NoError(t, err)
	rt, err := router.NewRouter([]config.ModelDeployment{{
		ModelName: "gpt-4",
		ProviderEndpoint: config.ProviderEndpointConfig{
			Provider: "openai",
			Model:    "gpt-4",
			APIKey:   "test-key",
		},
	}}, config.RouterSettings{CooldownTime: 60, NumRetries: 1}, mem, nil, zap.NewNop())
	require.NoError(t, err)

	h := NewCompletionsHandler(rt, tunnel, &testBudgeter{}, wallets, logs, zap.NewNop())
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
	app.Post("/v1/completions", h.Handle)
	return app
}

func TestCompletionsHandlerStreamTransformsChatChunksAndDone(t *testing.T) {
	stream := newFakeTunnelStream(map[string]string{"content-type": "text/event-stream"}, 3, 0)
	stream.chunks <- platform.TunnelChunk{Data: []byte("data: {\"id\":\"chatcmpl-stream\",\"object\":\"chat.completion.chunk\",\"created\":123,\"model\":\"gpt-4\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\"Salam\"}}]}\n\n")}
	stream.chunks <- platform.TunnelChunk{Data: []byte("data: [DONE]\n\n")}
	close(stream.chunks)

	wallets := &testWallets{}
	logs := &testLogs{}
	app := newCompletionsStreamingApp(t, &testTunnel{stream: stream}, wallets, logs)

	req := httptest.NewRequest("POST", "/v1/completions", bytes.NewBufferString(`{"model":"gpt-4","prompt":"hello","stream":true}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, int((5 * time.Second).Milliseconds()))
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))
	assert.Equal(t, "no", resp.Header.Get("X-Accel-Buffering"))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	bodyText := string(body)
	assert.Contains(t, bodyText, `"object":"text_completion"`)
	assert.Contains(t, bodyText, `"text":"Salam"`)
	assert.True(t, stream.closed)
	assert.Empty(t, wallets.debits)
	assert.Empty(t, logs.loggedSpends)
}

func TestCompletionsHandlerStreamErrorWritesStreamingErrorEvent(t *testing.T) {
	stream := newFakeTunnelStream(map[string]string{"content-type": "text/event-stream"}, 0, 1)
	stream.errors <- errors.New("provider stream exploded")

	app := newCompletionsStreamingApp(t, &testTunnel{stream: stream}, &testWallets{}, &testLogs{})

	req := httptest.NewRequest("POST", "/v1/completions", bytes.NewBufferString(`{"model":"gpt-4","prompt":"hello","stream":true}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, int((5 * time.Second).Milliseconds()))
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), `"type":"streaming_error"`)
	assert.True(t, stream.closed)
}

func TestCompletionsHandlerStreamCloseFinalizesWithoutPanic(t *testing.T) {
	stream := newFakeTunnelStream(map[string]string{"content-type": "text/event-stream"}, 1, 0)
	stream.chunks <- platform.TunnelChunk{Data: []byte("data: {\"id\":\"chatcmpl-stream\",\"object\":\"chat.completion.chunk\",\"created\":123,\"model\":\"gpt-4\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\"Salam\"}}]}\n\n")}
	close(stream.chunks)

	wallets := &testWallets{}
	logs := &testLogs{}
	app := newCompletionsStreamingApp(t, &testTunnel{stream: stream}, wallets, logs)

	req := httptest.NewRequest("POST", "/v1/completions", bytes.NewBufferString(`{"model":"gpt-4","prompt":"hello","stream":true}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, int((5 * time.Second).Milliseconds()))
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), `"text":"Salam"`)
	assert.NotContains(t, string(body), "panic")
	assert.Len(t, wallets.debits, 1)
	assert.Len(t, logs.loggedSpends, 1)
	assert.True(t, stream.closed)
	assert.Equal(t, "tenant_a", wallets.debits[0].TenantID)
	assert.Equal(t, "gpt-4", logs.loggedSpends[0].Model)
	assert.Equal(t, "success", logs.loggedSpends[0].Status)
}

func TestCompletionsHandlerStreamPassesThroughUnparseableChunk(t *testing.T) {
	stream := newFakeTunnelStream(map[string]string{"content-type": "text/event-stream"}, 2, 0)
	stream.chunks <- platform.TunnelChunk{Data: []byte("event: response.created\n")}
	stream.chunks <- platform.TunnelChunk{Data: []byte("data: {\"raw\":true}\n\n"), IsDone: true}
	close(stream.chunks)

	app := newCompletionsStreamingApp(t, &testTunnel{stream: stream}, &testWallets{}, &testLogs{})

	req := httptest.NewRequest("POST", "/v1/completions", bytes.NewBufferString(`{"model":"gpt-4","prompt":"hello","stream":true}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, int((5 * time.Second).Milliseconds()))
	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	bodyText := string(body)
	assert.True(t, strings.Contains(bodyText, "event: response.created") || strings.Contains(bodyText, "data: {\"raw\":true}"))
}
