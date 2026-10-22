package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/testcache"

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

func (b *testBudgeter) EstimateCost(_ context.Context, req contractsopenai.ChatCompletionRequest) (platform.UsageCostBreakdown, error) {
	return platform.UsageCostBreakdown{Model: req.Model, TotalTokens: 32, InputTokens: 16, OutputTokens: 16, TotalCostUSD: 0.001, TotalCostIRR: 850}, nil
}
func (b *testBudgeter) EstimateEmbeddingCost(_ context.Context, req contractsopenai.EmbeddingRequest) (platform.UsageCostBreakdown, error) {
	return platform.UsageCostBreakdown{Model: req.Model, TotalTokens: 16, InputTokens: 16, OutputTokens: 0, TotalCostUSD: 0.001, TotalCostIRR: 850}, nil
}
func (b *testBudgeter) CheckRateLimit(_ context.Context, _ *platform.AuthContext, _ string) (*cache.RateLimitResult, error) {
	return &cache.RateLimitResult{Allowed: true, Current: 1, Limit: 100, ResetAt: time.Now().Add(time.Minute)}, nil
}
func (b *testBudgeter) CheckBudget(_ context.Context, _ *platform.AuthContext, _ platform.UsageCostBreakdown) error {
	return nil
}
func (b *testBudgeter) FinalizeFromUsage(_ context.Context, _ string, usage *contractsopenai.Usage, estimated platform.UsageCostBreakdown) (platform.UsageCostBreakdown, error) {
	if usage != nil {
		estimated.InputTokens = usage.PromptTokens
		estimated.OutputTokens = usage.CompletionTokens
		estimated.TotalTokens = usage.TotalTokens
	}

	return estimated, nil
}

func (b *testBudgeter) FinalizeImageUsage(_ context.Context, _ string, usage *contractsopenai.Usage, providerCostUSDTicks *int64, estimated platform.UsageCostBreakdown) (platform.UsageCostBreakdown, error) {
	if usage != nil {
		estimated.InputTokens = usage.PromptTokens
		estimated.OutputTokens = usage.CompletionTokens

		estimated.TotalTokens = usage.TotalTokens
		if estimated.TotalTokens == 0 {
			estimated.TotalTokens = estimated.InputTokens + estimated.OutputTokens
		}
	}

	if providerCostUSDTicks != nil {
		estimated.TotalCostUSD = float64(*providerCostUSDTicks) / 10_000_000_000.0
		estimated.TotalCostIRR = estimated.TotalCostUSD * 850000
	}

	return estimated, nil
}

func (b *testBudgeter) PriceMeteredUsage(_ context.Context, model, _ string, _ string, estimated platform.UsageCostBreakdown) (platform.UsageCostBreakdown, error) {
	if !strings.Contains(strings.ToLower(model), "mistral") && !strings.Contains(strings.ToLower(model), "voxtral") && !strings.Contains(strings.ToLower(model), "codestral") {
		return estimated, nil
	}

	const revision = "mistral-test-revision"

	estimated.PricingRevision = revision
	estimated.ModelRuntimeProfile = &platform.ModelRuntimeProfileBinding{
		Provider: "mistral-ai", CatalogRevision: revision,
		CanonicalID: providerModelID(model), Slug: providerModelID(model), ProfileVersion: 2,
		EffectiveFrom: time.Now().UTC(),
		Profile: platform.ModelRuntimeProfile{
			Version: 2, Endpoints: []string{"chat_completions", "fim_completions", "embeddings", "moderations", "chat_moderations", "audio_speech", "audio_transcriptions", "ocr", "agents", "conversations", "files", "batches"},
			Transports: []string{"http", "sse"}, Lifecycle: "active",
			Thinking:   platform.ModelRuntimeThinkingProfile{DefaultMode: "disabled", AllowedModes: []string{"disabled"}},
			Sampling:   platform.ModelRuntimeSamplingProfile{Parameters: []string{"temperature", "top_p"}},
			Tools:      platform.ModelRuntimeToolsProfile{Custom: true, Parallel: true, StructuredOutput: true},
			Modalities: platform.ModelRuntimeModalities{Input: []string{"text", "audio", "image", "pdf", "document"}, Output: []string{"text", "audio", "embedding", "document", "moderation_result"}},
			Parameters: map[string][]string{
				"audio_transcriptions": {"model", "file", "file_url", "file_id", "language", "temperature", "stream", "diarize", "context_bias", "timestamp_granularities"},
				"audio_speech":         {"model", "input", "metadata", "stream", "prompt_cache_key", "voice_id", "ref_audio", "response_format"},
			},
			Limits: platform.ModelRuntimeLimits{MessagesBytes: 33554432, BatchBytes: 536870912, FileBytes: 536870912},
		},
	}

	return estimated, nil
}

type testLogs struct {
	mu            sync.RWMutex
	loggedEntries []platform.RequestLogEntry
	loggedSpends  []platform.SpendLog
}

func (l *testLogs) LogRequest(_ context.Context, entry platform.RequestLogEntry) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.loggedEntries = append(l.loggedEntries, entry)

	return nil
}

func (l *testLogs) LogSpend(_ context.Context, log platform.SpendLog) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.loggedSpends = append(l.loggedSpends, log)

	return nil
}

func (l *testLogs) requestEntries() []platform.RequestLogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return append([]platform.RequestLogEntry(nil), l.loggedEntries...)
}

type fakeTunnelStream struct {
	statusCode int
	headers    map[string]string
	chunks     chan platform.TunnelChunk
	errors     chan error
	closed     bool
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
		statusCode: http.StatusOK,
		headers:    headers,
		chunks:     make(chan platform.TunnelChunk, chunkBuf),
		errors:     make(chan error, errorBuf),
	}
}

func (s *fakeTunnelStream) Chunks() <-chan platform.TunnelChunk { return s.chunks }
func (s *fakeTunnelStream) Errors() <-chan error                { return s.errors }
func (s *fakeTunnelStream) Headers() map[string]string          { return s.headers }
func (s *fakeTunnelStream) StatusCode() int                     { return s.statusCode }
func (s *fakeTunnelStream) Close() error {
	s.closed = true
	return nil
}

type testTunnel struct {
	resp   *platform.TunnelResponse
	err    error
	stream platform.TunnelStream
}

func (t *testTunnel) Proxy(_ context.Context, _ platform.TunnelRequest) (*platform.TunnelResponse, error) {
	return t.resp, t.err
}
func (t *testTunnel) Stream(_ context.Context, _ platform.TunnelRequest) (platform.TunnelStream, error) {
	if t.stream != nil {
		return t.stream, nil
	}

	return nil, assert.AnError
}
func (t *testTunnel) Healthy(_ context.Context) error { return nil }

func newCompletionsTestApp(t *testing.T, tunnelResp *platform.TunnelResponse, logsOpt ...*testLogs) (*fiber.App, *tracing.Queue) {
	t.Helper()

	mem, err := testcache.New(t, zap.NewNop())
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

	logs := &testLogs{}
	if len(logsOpt) > 0 && logsOpt[0] != nil {
		logs = logsOpt[0]
	}

	h := NewCompletionsHandler(rt, &testTunnel{resp: tunnelResp}, &testBudgeter{}, logs, zap.NewNop())
	q := tracing.NewQueue(64)
	s := tracing.NewSampler(1.0, 1500)

	app := fiber.New()
	app.Use(tracing.Middleware(q, s, 1500))
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(auth.AuthContextKey{}, &platform.AuthContext{
			TenantID:  "tenant_a",
			KeyID:     "key_1",
			KeyType:   "pat",
			Role:      "member",
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

func newChatCompletionsTestApp(t *testing.T, tunnelResp *platform.TunnelResponse, logs *testLogs) (*fiber.App, *tracing.Queue) {
	t.Helper()

	mem, err := testcache.New(t, zap.NewNop())
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

	if logs == nil {
		logs = &testLogs{}
	}

	h := NewChatCompletionHandler(rt, &testTunnel{resp: tunnelResp}, &testBudgeter{}, logs, zap.NewNop())
	q := tracing.NewQueue(64)
	s := tracing.NewSampler(1.0, 1500)

	app := fiber.New()
	app.Use(tracing.Middleware(q, s, 1500))
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(auth.AuthContextKey{}, &platform.AuthContext{
			TenantID:  "tenant_a",
			KeyID:     "key_1",
			KeyType:   "pat",
			Role:      "member",
			TokenHash: "hash_1",
			AllowedModelTypes: map[string][]string{
				"gpt-4": {"chat"},
			},
		})
		c.Locals(tenant.ContextKey{}, &platform.Tenant{
			ID:     "tenant_a",
			Slug:   "a",
			Status: "active",
		})

		return c.Next()
	})
	app.Post("/v1/chat/completions", h.Handle)

	return app, q
}

func TestChatCompletionsLogsFullRequestAndResponsePayload(t *testing.T) {
	payload := contractsopenai.ChatCompletionResponse{
		ID:      "chatcmpl_log",
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   "gpt-4",
		Choices: []contractsopenai.Choice{{
			Index:        0,
			Message:      &contractsopenai.ChatMessage{Role: "assistant", Content: "سلام"},
			FinishReason: "stop",
		}},
		Usage: &contractsopenai.Usage{PromptTokens: 4, CompletionTokens: 3, TotalTokens: 7},
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	logs := &testLogs{}
	app, _ := newChatCompletionsTestApp(t, &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       body,
	}, logs)

	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBufferString(`{"model":"gpt-4","messages":[{"role":"user","content":"سلام"}],"temperature":0.2}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.Len(t, logs.loggedEntries, 1)
	assert.Equal(t, "gpt-4", logs.loggedEntries[0].RequestBody["model"])
	assert.Equal(t, float64(0.2), logs.loggedEntries[0].RequestBody["temperature"])
	requestMessages, ok := logs.loggedEntries[0].RequestBody["messages"].([]contractsopenai.ChatMessage)
	require.True(t, ok)
	require.Len(t, requestMessages, 1)
	assert.Equal(t, "سلام", requestMessages[0].Content)
	assert.Equal(t, "chatcmpl_log", logs.loggedEntries[0].ResponseBody["id"])
	assert.Equal(t, "chat.completion", logs.loggedEntries[0].ResponseBody["object"])
	assert.NotContains(t, logs.loggedEntries[0].ResponseBody, "embedding_count")
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

	logs := &testLogs{}
	app, q := newCompletionsTestApp(t, &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       body,
	}, logs)

	req := httptest.NewRequest("POST", "/v1/completions", bytes.NewBufferString(`{"model":"gpt-4","prompt":"hello"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

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

	require.Eventually(t, func() bool { return len(logs.requestEntries()) == 1 }, time.Second, 10*time.Millisecond)

	entry := logs.requestEntries()[0]
	assert.Equal(t, map[string]interface{}{"model": "gpt-4", "prompt": "hello"}, entry.RequestBody)
	assert.Equal(t, "chatcmpl_1", entry.ResponseBody["id"])
	assert.Equal(t, "text_completion", entry.ResponseBody["object"])
	assert.NotContains(t, entry.ResponseBody, "result_count")
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
	defer closeResponseBody(t, resp)

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

func newCompletionsStreamingApp(t *testing.T, tunnel platform.TunnelTransport, logs *testLogs) *fiber.App {
	t.Helper()

	mem, err := testcache.New(t, zap.NewNop())
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

	h := NewCompletionsHandler(rt, tunnel, &testBudgeter{}, logs, zap.NewNop())
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(auth.AuthContextKey{}, &platform.AuthContext{
			TenantID:  "tenant_a",
			KeyID:     "key_1",
			KeyAlias:  "test-key",
			KeyType:   "pat",
			Role:      "member",
			TokenHash: "hash_1",
			UserID:    "user_1",
			TeamID:    "team_1",
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

	logs := &testLogs{}
	app := newCompletionsStreamingApp(t, &testTunnel{stream: stream}, logs)

	req := httptest.NewRequest("POST", "/v1/completions", bytes.NewBufferString(`{"model":"gpt-4","prompt":"hello","stream":true}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, int((5 * time.Second).Milliseconds()))
	defer closeResponseBody(t, resp)

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
	assert.Len(t, logs.loggedSpends, 1)
	assert.Equal(t, "success", logs.loggedSpends[0].Status)
	assert.Equal(t, "user_1", logs.loggedSpends[0].UserID, "the credit ledger attributes usage to the user")
	assert.Equal(t, "team_1", logs.loggedSpends[0].TeamID)
	require.Len(t, logs.loggedEntries, 1)
	assert.Equal(t, "success", logs.loggedEntries[0].Status)
	assert.Equal(t, fiber.StatusOK, logs.loggedEntries[0].StatusCode)
	assert.Equal(t, "/v1/completions", logs.loggedEntries[0].Route)
	assert.Equal(t, true, logs.loggedEntries[0].RequestBody["stream"])
	assert.Equal(t, "completed", logs.loggedEntries[0].ResponseBody["outcome"])
	assert.Contains(t, logs.loggedEntries[0].ResponseBody["raw_stream"], `"object":"text_completion"`)
}

func TestCompletionsHandlerStreamErrorWritesStreamingErrorEvent(t *testing.T) {
	stream := newFakeTunnelStream(map[string]string{"content-type": "text/event-stream"}, 0, 1)
	stream.errors <- errors.New("provider stream exploded")

	logs := &testLogs{}
	app := newCompletionsStreamingApp(t, &testTunnel{stream: stream}, logs)

	req := httptest.NewRequest("POST", "/v1/completions", bytes.NewBufferString(`{"model":"gpt-4","prompt":"hello","stream":true}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, int((5 * time.Second).Milliseconds()))
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), `"type":"streaming_error"`)
	assert.True(t, stream.closed)
	require.Len(t, logs.loggedEntries, 1)
	assert.Equal(t, "failed", logs.loggedEntries[0].Status)
	assert.Equal(t, fiber.StatusOK, logs.loggedEntries[0].StatusCode)
	assert.Equal(t, "streaming_error", logs.loggedEntries[0].ErrorType)
	assert.Equal(t, "failed", logs.loggedEntries[0].ResponseBody["outcome"])
}

func TestCompletionsHandlerStreamCloseFinalizesWithoutPanic(t *testing.T) {
	stream := newFakeTunnelStream(map[string]string{"content-type": "text/event-stream"}, 1, 0)
	stream.chunks <- platform.TunnelChunk{Data: []byte("data: {\"id\":\"chatcmpl-stream\",\"object\":\"chat.completion.chunk\",\"created\":123,\"model\":\"gpt-4\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\"Salam\"}}]}\n\n")}

	close(stream.chunks)

	logs := &testLogs{}
	app := newCompletionsStreamingApp(t, &testTunnel{stream: stream}, logs)

	req := httptest.NewRequest("POST", "/v1/completions", bytes.NewBufferString(`{"model":"gpt-4","prompt":"hello","stream":true}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, int((5 * time.Second).Milliseconds()))
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), `"text":"Salam"`)
	assert.NotContains(t, string(body), "panic")
	assert.Len(t, logs.loggedSpends, 1)
	assert.True(t, stream.closed)
	assert.Equal(t, "gpt-4", logs.loggedSpends[0].Model)
	assert.Equal(t, "success", logs.loggedSpends[0].Status)
}

func TestCompletionsHandlerStreamPassesThroughUnparseableChunk(t *testing.T) {
	stream := newFakeTunnelStream(map[string]string{"content-type": "text/event-stream"}, 2, 0)
	stream.chunks <- platform.TunnelChunk{Data: []byte("event: response.created\n")}

	stream.chunks <- platform.TunnelChunk{Data: []byte("data: {\"raw\":true}\n\n"), IsDone: true}

	close(stream.chunks)

	app := newCompletionsStreamingApp(t, &testTunnel{stream: stream}, &testLogs{})

	req := httptest.NewRequest("POST", "/v1/completions", bytes.NewBufferString(`{"model":"gpt-4","prompt":"hello","stream":true}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, int((5 * time.Second).Milliseconds()))
	defer closeResponseBody(t, resp)

	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	bodyText := string(body)
	assert.True(t, strings.Contains(bodyText, "event: response.created") || strings.Contains(bodyText, "data: {\"raw\":true}"))
}

type completionsDeploymentStore struct {
	deployments []platform.Deployment
}

func (s completionsDeploymentStore) ListDeployments(_ context.Context, _ string) ([]platform.Deployment, error) {
	return append([]platform.Deployment(nil), s.deployments...), nil
}

type budgetDenyingBudgeter struct {
	testBudgeter
	checkBudgetCalls int
}

func (b *budgetDenyingBudgeter) CheckBudget(_ context.Context, _ *platform.AuthContext, _ platform.UsageCostBreakdown) error {
	b.checkBudgetCalls++
	return errors.New("wallet balance required")
}

func TestCompletionsHandlerBYOKProviderAccountSkipsWalletBudgetAndDebit(t *testing.T) {
	payload := contractsopenai.ChatCompletionResponse{
		ID:      "chatcmpl_byok",
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   "gpt-4o-mini",
		Choices: []contractsopenai.Choice{
			{Index: 0, Message: &contractsopenai.ChatMessage{Role: "assistant", Content: "hi"}},
		},
		Usage: &contractsopenai.Usage{PromptTokens: 4, CompletionTokens: 2, TotalTokens: 6},
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       body,
	}}
	billing := &budgetDenyingBudgeter{}
	store := completionsDeploymentStore{deployments: []platform.Deployment{{
		ID:                  "provider-account:acct_1:gpt-4o-mini",
		TenantID:            "tenant_a",
		ModelName:           "openai:personal-openai:gpt-4o-mini",
		Provider:            "openai",
		Model:               "gpt-4o-mini",
		APIKey:              "user-openai-key",
		ProviderAccountID:   "acct_1",
		ProviderAccountName: "personal-openai",
	}}}

	mem, err := testcache.New(t, zap.NewNop())
	require.NoError(t, err)
	rt, err := router.NewRouter(nil, config.RouterSettings{CooldownTime: 60, NumRetries: 1}, mem, store, zap.NewNop())
	require.NoError(t, err)

	h := NewCompletionsHandler(rt, tunnel, billing, &testLogs{}, zap.NewNop())
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(auth.AuthContextKey{}, &platform.AuthContext{
			TenantID:  "tenant_a",
			KeyID:     "key_1",
			KeyType:   "pat",
			Role:      "member",
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

	req := httptest.NewRequest("POST", "/v1/completions", bytes.NewBufferString(`{"model":"openai:personal-openai:gpt-4o-mini","prompt":"hello"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Zero(t, billing.checkBudgetCalls)
	require.NotNil(t, tunnel.lastReq.Deployment.ProviderAuth)
	assert.Equal(t, "user-openai-key", tunnel.lastReq.Deployment.ProviderAuth.APIKey)
	assert.Empty(t, tunnel.lastReq.Deployment.APIKey)
	assert.Equal(t, platform.CredentialSourceUserBYOK, tunnel.lastReq.RequestMetadata["credential_source"])
	assert.Equal(t, platform.BillingModeExternalBilled, tunnel.lastReq.RequestMetadata["billing_mode"])
	assert.Equal(t, "acct_1", tunnel.lastReq.RequestMetadata["provider_account_id"])
	assert.Equal(t, "personal-openai", tunnel.lastReq.RequestMetadata["provider_account_name"])
}

func TestCompletionsHandlerBYOKProviderAccountStreamLogsSpendOnDone(t *testing.T) {
	stream := newFakeTunnelStream(map[string]string{"content-type": "text/event-stream"}, 2, 0)
	stream.chunks <- platform.TunnelChunk{Data: []byte("data: {\"id\":\"chatcmpl-stream\",\"object\":\"chat.completion.chunk\",\"created\":123,\"model\":\"gpt-4o-mini\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\"Salam\"}}]}\n\n"), IsDone: true}

	close(stream.chunks)

	billing := &budgetDenyingBudgeter{}
	logs := &testLogs{}
	store := completionsDeploymentStore{deployments: []platform.Deployment{{
		ID:                  "provider-account:acct_1:gpt-4o-mini",
		TenantID:            "tenant_a",
		ModelName:           "openai:personal-openai:gpt-4o-mini",
		Provider:            "openai",
		Model:               "gpt-4o-mini",
		APIKey:              "user-openai-key",
		ProviderAccountID:   "acct_1",
		ProviderAccountName: "personal-openai",
	}}}

	mem, err := testcache.New(t, zap.NewNop())
	require.NoError(t, err)
	rt, err := router.NewRouter(nil, config.RouterSettings{CooldownTime: 60, NumRetries: 1}, mem, store, zap.NewNop())
	require.NoError(t, err)

	h := NewCompletionsHandler(rt, &testTunnel{stream: stream}, billing, logs, zap.NewNop())
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(auth.AuthContextKey{}, &platform.AuthContext{
			TenantID:  "tenant_a",
			KeyID:     "key_1",
			KeyType:   "pat",
			Role:      "member",
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

	req := httptest.NewRequest("POST", "/v1/completions", bytes.NewBufferString(`{"model":"openai:personal-openai:gpt-4o-mini","prompt":"hello","stream":true}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, int((5 * time.Second).Milliseconds()))
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Zero(t, billing.checkBudgetCalls)
	require.Len(t, logs.loggedSpends, 1)
	assert.Equal(t, platform.BillingModeExternalBilled, logs.loggedSpends[0].RequestMetadata["billing_mode"])
	assert.Equal(t, platform.CredentialSourceUserBYOK, logs.loggedSpends[0].RequestMetadata["credential_source"])
	assert.Equal(t, "acct_1", logs.loggedSpends[0].RequestMetadata["provider_account_id"])
}
