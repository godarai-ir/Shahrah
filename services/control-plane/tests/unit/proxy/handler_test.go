package proxy_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/apiheaders"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	mcache "github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/config"
	"github.com/Delpak/shahrah/services/control-plane/internal/guardrails"
	metricsobs "github.com/Delpak/shahrah/services/control-plane/internal/observability/metrics"
	"github.com/Delpak/shahrah/services/control-plane/internal/proxy"
	"github.com/Delpak/shahrah/services/control-plane/internal/responsecache"
	llmrouter "github.com/Delpak/shahrah/services/control-plane/internal/router"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// mockTunnelTransport implements platform.TunnelTransport for testing.
type mockTunnelTransport struct {
	response    *platform.TunnelResponse
	stream      platform.TunnelStream
	proxyErr    error
	streamErr   error
	lastRequest platform.TunnelRequest
	proxyCalls  int
}

func (m *mockTunnelTransport) Proxy(ctx context.Context, req platform.TunnelRequest) (*platform.TunnelResponse, error) {
	m.lastRequest = req
	m.proxyCalls++

	return m.response, m.proxyErr
}

func TestChatCompletionHandler_SimpleCacheHitSkipsProviderAndWallet(t *testing.T) {
	logger := zap.NewNop()
	store, err := mcache.NewMemoryCache(logger)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })
	r := newTestRouter(t, logger)
	providerBody, err := json.Marshal(contractsopenai.ChatCompletionResponse{
		ID: "chatcmpl-cache", Object: "chat.completion", Model: "gpt-4",
		Choices: []contractsopenai.Choice{{Index: 0, Message: &contractsopenai.ChatMessage{Role: "assistant", Content: "cached"}, FinishReason: "stop"}},
		Usage:   &contractsopenai.Usage{PromptTokens: 2, CompletionTokens: 1, TotalTokens: 3},
	})
	require.NoError(t, err)

	tunnel := &mockTunnelTransport{response: &platform.TunnelResponse{StatusCode: http.StatusOK, Body: providerBody}}
	billing := &mockBilling{cost: platform.UsageCostBreakdown{InputTokens: 2, OutputTokens: 1, TotalTokens: 3, TotalCostUSD: 0.01, TotalCostIRR: 8500, ExchangeRate: 850000}}
	wallet := &mockWalletStore{}
	logs := &mockLogger{}
	metricInputs := make([]metricsobs.BuildInput, 0, 2)
	handler := proxy.NewChatCompletionHandler(r, tunnel, billing, wallet, logs, logger, func(input metricsobs.BuildInput) {
		metricInputs = append(metricInputs, input)
	})
	service, err := responsecache.NewService(store, nil, responsecache.Config{
		Enabled: true, DefaultTTL: time.Minute, MinTTL: time.Second, MaxTTL: time.Hour, MaxResponseBytes: 1024,
	})
	require.NoError(t, err)
	handler.SetResponseCache(service)
	app := setupTestApp(t, handler)

	makeRequest := func() *testResponse {
		req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"gpt-4","messages":[{"role":"user","content":"hello"}]}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set(apiheaders.HeaderCacheConfig, `{"type":"simple","ttl":60}`)
		resp, requestErr := app.Test(req)
		require.NoError(t, requestErr)
		t.Cleanup(func() { closeResponseBody(t, resp) })

		return &testResponse{Response: resp}
	}

	first := makeRequest()
	assert.Equal(t, http.StatusOK, first.StatusCode)
	assert.Equal(t, "miss", first.Header.Get(apiheaders.HeaderCacheStatus))

	second := makeRequest()
	assert.Equal(t, http.StatusOK, second.StatusCode)
	assert.Equal(t, "hit", second.Header.Get(apiheaders.HeaderCacheStatus))
	assert.NotEmpty(t, second.Header.Get(apiheaders.HeaderCachedTraceID))
	assert.Equal(t, 1, tunnel.proxyCalls)
	assert.Len(t, wallet.debitCalls, 1)
	require.Len(t, logs.loggedSpends, 2)
	assert.Zero(t, logs.loggedSpends[1].CostUSD)
	require.Len(t, metricInputs, 2)
	assert.Equal(t, "simple", metricInputs[1].CacheMode)
	assert.Equal(t, "hit", metricInputs[1].CacheStatus)
	assert.Equal(t, "exact", metricInputs[1].CacheMatch)
	assert.Equal(t, 0.01, metricInputs[1].AvoidedCostUSD)
	assert.Equal(t, 8500.0, metricInputs[1].AvoidedCostIRR)
	assert.Equal(t, 2, metricInputs[1].AvoidedInputTokens)
}

func (m *mockTunnelTransport) Stream(ctx context.Context, req platform.TunnelRequest) (platform.TunnelStream, error) {
	m.lastRequest = req
	return m.stream, m.streamErr
}

func (m *mockTunnelTransport) Healthy(ctx context.Context) error { return nil }

// mockBilling implements budgeter interface for testing.
type mockBilling struct {
	cost          platform.UsageCostBreakdown
	estimateErr   error
	rateLimitErr  error
	rateLimitRes  *mcache.RateLimitResult
	budgetErr     error
	estimateModel string
}

func (m *mockBilling) EstimateCost(ctx context.Context, req contractsopenai.ChatCompletionRequest) (platform.UsageCostBreakdown, error) {
	m.estimateModel = req.Model
	return m.cost, m.estimateErr
}

func (m *mockBilling) EstimateEmbeddingCost(ctx context.Context, req contractsopenai.EmbeddingRequest) (platform.UsageCostBreakdown, error) {
	m.estimateModel = req.Model
	return m.cost, m.estimateErr
}

func (m *mockBilling) CheckRateLimit(ctx context.Context, authCtx *platform.AuthContext, model string, estimatedTokens int) (*mcache.RateLimitResult, error) {
	return m.rateLimitRes, m.rateLimitErr
}

func (m *mockBilling) CheckBudget(ctx context.Context, authCtx *platform.AuthContext, breakdown platform.UsageCostBreakdown) error {
	return m.budgetErr
}

func (m *mockBilling) FinalizeFromUsage(ctx context.Context, model string, usage *contractsopenai.Usage, estimated platform.UsageCostBreakdown) (platform.UsageCostBreakdown, error) {
	return m.cost, nil
}

func (m *mockBilling) FinalizeImageUsage(ctx context.Context, model string, usage *contractsopenai.Usage, providerCostUSDTicks *int64, estimated platform.UsageCostBreakdown) (platform.UsageCostBreakdown, error) {
	return m.cost, nil
}

// mockWalletStore implements walletStore interface.
type mockWalletStore struct {
	transaction *platform.Transaction
	err         error
	debitCalls  []platform.UsageCostBreakdown
}

func (m *mockWalletStore) DebitWallet(ctx context.Context, tenantID string, breakdown platform.UsageCostBreakdown, requestID string) (*platform.Transaction, error) {
	m.debitCalls = append(m.debitCalls, breakdown)
	return m.transaction, m.err
}

// mockLogger implements usageLogger interface.
type mockLogger struct {
	loggedEntries    []platform.RequestLogEntry
	loggedSpends     []platform.SpendLog
	virtualModels    map[string]*platform.VirtualModel
	tenantGuardrails []platform.TenantGuardrail
}

func (m *mockLogger) LogRequest(ctx context.Context, entry platform.RequestLogEntry) error {
	m.loggedEntries = append(m.loggedEntries, entry)
	return nil
}

func (m *mockLogger) LogSpend(ctx context.Context, log platform.SpendLog) error {
	m.loggedSpends = append(m.loggedSpends, log)
	return nil
}

func (m *mockLogger) GetVirtualModelByAlias(ctx context.Context, tenantID string, alias string) (*platform.VirtualModel, error) {
	if m.virtualModels != nil {
		if vm := m.virtualModels[alias]; vm != nil {
			return vm, nil
		}
	}

	return nil, errors.New("virtual model not found")
}

func (m *mockLogger) ListEnabledTenantGuardrailsBySlugs(ctx context.Context, tenantID string, slugs []string) ([]platform.TenantGuardrail, error) {
	wanted := map[string]bool{}
	for _, slug := range slugs {
		wanted[strings.TrimSpace(slug)] = true
	}

	out := make([]platform.TenantGuardrail, 0, len(m.tenantGuardrails))
	for _, item := range m.tenantGuardrails {
		if item.TenantID != tenantID || !item.Enabled || !wanted[item.Slug] {
			continue
		}

		out = append(out, item)
	}

	return out, nil
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

func newTestRouter(t *testing.T, logger *zap.Logger) *llmrouter.Router {
	t.Helper()

	memCache, err := mcache.NewMemoryCache(logger)
	require.NoError(t, err)

	modelList := []config.ModelDeployment{
		{
			ModelName: "gpt-4",
			ProviderEndpoint: config.ProviderEndpointConfig{
				Provider: "openai",
				Model:    "openai/gpt-4",
				RPM:      1000,
			},
			ModelInfo: map[string]interface{}{"id": "dep-1"},
		},
	}

	settings := config.RouterSettings{
		RoutingStrategy: "simple-shuffle",
		CooldownTime:    60,
		NumRetries:      3,
	}

	r, err := llmrouter.NewRouter(modelList, settings, memCache, nil, logger)
	require.NoError(t, err)

	return r
}

func setupTestApp(t *testing.T, handler *proxy.ChatCompletionHandler) *fiber.App {
	t.Helper()

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(auth.AuthContextKey{}, &platform.AuthContext{
			TokenHash:     "test-hash",
			TenantID:      "tenant-1",
			AllowedModels: []string{"*"},
		})

		return c.Next()
	})
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(tenant.ContextKey{}, &platform.Tenant{
			ID:   "tenant-1",
			Slug: "test-tenant",
		})

		return c.Next()
	})
	app.Post("/v1/chat/completions", handler.Handle)

	return app
}

func TestChatCompletionHandler_RoutesFQNThroughWildcardAsOfficialProviderModel(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, err := mcache.NewMemoryCache(logger)
	require.NoError(t, err)

	r, err := llmrouter.NewRouter([]config.ModelDeployment{
		{
			ModelName: "openai:openai:gpt-4.1",
			ProviderEndpoint: config.ProviderEndpointConfig{
				Provider: "openai",
				Model:    "openai/*",
				APIKey:   "sk-provider",
				RPM:      1000,
			},
			ModelInfo: map[string]interface{}{"id": "dep-wildcard"},
		},
	}, config.RouterSettings{RoutingStrategy: "simple-shuffle", CooldownTime: 60}, memCache, nil, logger)
	require.NoError(t, err)

	validResponse := contractsopenai.ChatCompletionResponse{
		ID:      "chatcmpl-123",
		Object:  "chat.completion",
		Created: 1234567890,
		Model:   "gpt-4.1",
		Choices: []contractsopenai.Choice{{
			Index: 0,
			Message: &contractsopenai.ChatMessage{
				Role:    "assistant",
				Content: "Hello!",
			},
			FinishReason: "stop",
		}},
		Usage: &contractsopenai.Usage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}
	respBody, _ := json.Marshal(validResponse)
	mTunnel := &mockTunnelTransport{
		response: &platform.TunnelResponse{
			StatusCode: http.StatusOK,
			Headers:    map[string]string{},
			Body:       respBody,
		},
	}
	mBilling := &mockBilling{cost: platform.UsageCostBreakdown{TotalCostUSD: 0.001, TotalTokens: 15}}
	mWallet := &mockWalletStore{}
	mLogs := &mockLogger{}

	handler := proxy.NewChatCompletionHandlerWithDependencies(proxy.ChatCompletionDependencies{
		Base: proxy.BaseHandlerDependencies{
			Router:        r,
			Tunnel:        mTunnel,
			Billing:       mBilling,
			Wallets:       mWallet,
			RequestLogger: mLogs,
			SpendLogger:   mLogs,
			Logger:        logger,
		},
		VirtualModelStore: mLogs,
	})
	app := setupTestApp(t, handler)

	payload, _ := json.Marshal(map[string]interface{}{
		"model":    "openai:openai:gpt-4.1",
		"messages": []interface{}{map[string]interface{}{"role": "user", "content": "Hello"}},
	})
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "gpt-4.1", mBilling.estimateModel)

	var outbound contractsopenai.ChatCompletionRequest
	require.NoError(t, json.Unmarshal(mTunnel.lastRequest.Body, &outbound))
	assert.Equal(t, "gpt-4.1", outbound.Model)
	assert.Equal(t, "openai", mTunnel.lastRequest.Deployment.Provider)
	assert.Equal(t, "sk-provider", mTunnel.lastRequest.Deployment.APIKey)
}

func TestChatCompletionHandler_VirtualModelSendsSelectedProviderModel(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, err := mcache.NewMemoryCache(logger)
	require.NoError(t, err)

	r, err := llmrouter.NewRouter([]config.ModelDeployment{
		{
			ModelName: "*",
			ProviderEndpoint: config.ProviderEndpointConfig{
				Provider: "openai",
				Model:    "openai/*",
				APIKey:   "sk-openai-wildcard",
				RPM:      1000,
			},
			ModelInfo: map[string]interface{}{"id": "dep-wildcard"},
		},
		{
			ModelName: "mistral-ai:mistral-ai-eager-harbor:ministral-8b-2512",
			ProviderEndpoint: config.ProviderEndpointConfig{
				Provider: "mistral-ai",
				Model:    "mistral-ai/ministral-8b-2512",
				APIKey:   "sk-mistral-provider",
				RPM:      1000,
			},
			ModelInfo: map[string]interface{}{"id": "dep-mistral"},
		},
	}, config.RouterSettings{RoutingStrategy: "simple-shuffle", CooldownTime: 60}, memCache, nil, logger)
	require.NoError(t, err)

	validResponse := contractsopenai.ChatCompletionResponse{
		ID:      "chatcmpl-virtual",
		Object:  "chat.completion",
		Created: 1234567890,
		Model:   "ministral-8b-2512",
		Choices: []contractsopenai.Choice{{
			Index: 0,
			Message: &contractsopenai.ChatMessage{
				Role:    "assistant",
				Content: "Hello from Mistral",
			},
			FinishReason: "stop",
		}},
		Usage: &contractsopenai.Usage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}
	respBody, _ := json.Marshal(validResponse)
	mTunnel := &mockTunnelTransport{
		response: &platform.TunnelResponse{
			StatusCode: http.StatusOK,
			Headers:    map[string]string{},
			Body:       respBody,
		},
	}
	mBilling := &mockBilling{cost: platform.UsageCostBreakdown{TotalCostUSD: 0.001, TotalTokens: 15}}
	mWallet := &mockWalletStore{}
	mLogs := &mockLogger{
		virtualModels: map[string]*platform.VirtualModel{
			"mistral": {
				ID:              "vm-mistral",
				TenantID:        "tenant-1",
				Alias:           "mistral",
				RoutingStrategy: "priority",
				Enabled:         true,
				Targets: []platform.VirtualModelTarget{{
					ID:       "target-1",
					Order:    1,
					Provider: "mistral-ai",
					Model:    "mistral-ai:mistral-ai-eager-harbor:ministral-8b-2512",
					Policy: platform.VirtualModelPolicy{
						Retry: platform.VirtualModelRetryPolicy{
							Attempts:    1,
							StatusCodes: []int{429, 500},
						},
						Fallback: platform.VirtualModelFallbackPolicy{
							StatusCodes: []int{429, 500},
							Candidate:   true,
						},
					},
				}},
			},
		},
	}

	handler := proxy.NewChatCompletionHandlerWithDependencies(proxy.ChatCompletionDependencies{
		Base: proxy.BaseHandlerDependencies{
			Router:        r,
			Tunnel:        mTunnel,
			Billing:       mBilling,
			Wallets:       mWallet,
			RequestLogger: mLogs,
			SpendLogger:   mLogs,
			Logger:        logger,
		},
		VirtualModelStore: mLogs,
	})
	app := setupTestApp(t, handler)

	payload, _ := json.Marshal(map[string]interface{}{
		"model":    "mistral",
		"messages": []interface{}{map[string]interface{}{"role": "user", "content": "Hello"}},
	})
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var outbound contractsopenai.ChatCompletionRequest
	require.NoError(t, json.Unmarshal(mTunnel.lastRequest.Body, &outbound))
	assert.Equal(t, "ministral-8b-2512", outbound.Model)
	assert.Equal(t, "mistral-ai", mTunnel.lastRequest.Deployment.Provider)
	assert.Equal(t, "ministral-8b-2512", mTunnel.lastRequest.Deployment.Model)
	assert.Equal(t, "sk-mistral-provider", mTunnel.lastRequest.Deployment.APIKey)
}

func TestChatCompletionHandler_ShouldRejectEmptyMessages(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	r := newTestRouter(t, logger)
	mTunnel := &mockTunnelTransport{}
	mBilling := &mockBilling{cost: platform.UsageCostBreakdown{TotalCostUSD: 0.001, TotalTokens: 100}}
	mWallet := &mockWalletStore{}
	mLogs := &mockLogger{}

	handler := proxy.NewChatCompletionHandler(r, mTunnel, mBilling, mWallet, mLogs, logger)
	app := setupTestApp(t, handler)

	body := map[string]interface{}{
		"model":    "gpt-4",
		"messages": []interface{}{},
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var apiErr contractsopenai.APIError

	err = json.NewDecoder(resp.Body).Decode(&apiErr)
	require.NoError(t, err)
	assert.Contains(t, apiErr.Error.Message, "messages array must not be empty")
}

func TestChatCompletionHandler_ShouldRejectMissingMessageRole(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	r := newTestRouter(t, logger)
	mTunnel := &mockTunnelTransport{}
	mBilling := &mockBilling{cost: platform.UsageCostBreakdown{TotalCostUSD: 0.001, TotalTokens: 100}}
	mWallet := &mockWalletStore{}
	mLogs := &mockLogger{}

	handler := proxy.NewChatCompletionHandler(r, mTunnel, mBilling, mWallet, mLogs, logger)
	app := setupTestApp(t, handler)

	body := map[string]interface{}{
		"model": "gpt-4",
		"messages": []interface{}{
			map[string]interface{}{"content": "Hello"},
		},
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var apiErr contractsopenai.APIError

	err = json.NewDecoder(resp.Body).Decode(&apiErr)
	require.NoError(t, err)
	assert.Contains(t, apiErr.Error.Message, "message[0].role is required")
}

func TestChatCompletionHandler_ShouldRejectInvalidTemperature(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	r := newTestRouter(t, logger)
	mTunnel := &mockTunnelTransport{}
	mBilling := &mockBilling{cost: platform.UsageCostBreakdown{TotalCostUSD: 0.001, TotalTokens: 100}}
	mWallet := &mockWalletStore{}
	mLogs := &mockLogger{}

	handler := proxy.NewChatCompletionHandler(r, mTunnel, mBilling, mWallet, mLogs, logger)
	app := setupTestApp(t, handler)

	temps := []float64{-0.1, 2.1, 3.0}
	for _, temp := range temps {
		t.Run("", func(t *testing.T) {
			body := map[string]interface{}{
				"model":       "gpt-4",
				"messages":    []interface{}{map[string]interface{}{"role": "user", "content": "Hello"}},
				"temperature": temp,
			}
			payload, _ := json.Marshal(body)

			req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(payload))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			defer closeResponseBody(t, resp)

			require.NoError(t, err)
			assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

			var apiErr contractsopenai.APIError

			err = json.NewDecoder(resp.Body).Decode(&apiErr)
			require.NoError(t, err)
			assert.Contains(t, apiErr.Error.Message, "temperature must be between 0 and 2")
		})
	}
}

func TestChatCompletionHandler_ShouldRejectInvalidTopP(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	r := newTestRouter(t, logger)
	mTunnel := &mockTunnelTransport{}
	mBilling := &mockBilling{cost: platform.UsageCostBreakdown{TotalCostUSD: 0.001, TotalTokens: 100}}
	mWallet := &mockWalletStore{}
	mLogs := &mockLogger{}

	handler := proxy.NewChatCompletionHandler(r, mTunnel, mBilling, mWallet, mLogs, logger)
	app := setupTestApp(t, handler)

	topPValues := []float64{-0.1, 1.1, 1.5}
	for _, topP := range topPValues {
		t.Run("", func(t *testing.T) {
			body := map[string]interface{}{
				"model":    "gpt-4",
				"messages": []interface{}{map[string]interface{}{"role": "user", "content": "Hello"}},
				"top_p":    topP,
			}
			payload, _ := json.Marshal(body)

			req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(payload))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			defer closeResponseBody(t, resp)

			require.NoError(t, err)
			assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

			var apiErr contractsopenai.APIError

			err = json.NewDecoder(resp.Body).Decode(&apiErr)
			require.NoError(t, err)
			assert.Contains(t, apiErr.Error.Message, "top_p must be between 0 and 1")
		})
	}
}

func TestChatCompletionHandler_ShouldRejectInvalidMaxTokens(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	r := newTestRouter(t, logger)
	mTunnel := &mockTunnelTransport{}
	mBilling := &mockBilling{cost: platform.UsageCostBreakdown{TotalCostUSD: 0.001, TotalTokens: 100}}
	mWallet := &mockWalletStore{}
	mLogs := &mockLogger{}

	handler := proxy.NewChatCompletionHandler(r, mTunnel, mBilling, mWallet, mLogs, logger)
	app := setupTestApp(t, handler)

	body := map[string]interface{}{
		"model":      "gpt-4",
		"messages":   []interface{}{map[string]interface{}{"role": "user", "content": "Hello"}},
		"max_tokens": 0,
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var apiErr contractsopenai.APIError

	err = json.NewDecoder(resp.Body).Decode(&apiErr)
	require.NoError(t, err)
	assert.Contains(t, apiErr.Error.Message, "max_tokens must be at least 1")
}

func TestChatCompletionHandler_DeniesWhenAllowlistEnforcedAndEmpty(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	r := newTestRouter(t, logger)
	mTunnel := &mockTunnelTransport{}
	mBilling := &mockBilling{cost: platform.UsageCostBreakdown{TotalCostUSD: 0.001, TotalTokens: 100}}
	mWallet := &mockWalletStore{}
	mLogs := &mockLogger{}

	handler := proxy.NewChatCompletionHandler(r, mTunnel, mBilling, mWallet, mLogs, logger)
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(auth.AuthContextKey{}, &platform.AuthContext{
			TokenHash:             "test-hash",
			TenantID:              "tenant-1",
			AllowedModels:         []string{},
			EnforceModelAllowlist: true,
		})

		return c.Next()
	})
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(tenant.ContextKey{}, &platform.Tenant{ID: "tenant-1", Slug: "test-tenant"})
		return c.Next()
	})
	app.Post("/v1/chat/completions", handler.Handle)

	body := map[string]interface{}{
		"model":    "gpt-4",
		"messages": []interface{}{map[string]interface{}{"role": "user", "content": "Hello"}},
	}
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	require.Len(t, mLogs.loggedEntries, 1)
	entry := mLogs.loggedEntries[0]
	assert.Equal(t, "blocked", entry.Status)
	assert.Equal(t, fiber.StatusForbidden, entry.StatusCode)
	assert.Equal(t, "authorization_error", entry.ErrorType)
	assert.Equal(t, "gpt-4", entry.Model)
	assert.Equal(t, "/v1/chat/completions", entry.Route)
	assert.Equal(t, "chat_completions", entry.RequestType)
	assert.Equal(t, "model_access", entry.Metadata["denial_source"])
	assert.Equal(t, "provider_account_policy", entry.Metadata["rejection_reason"])
	assert.Zero(t, mTunnel.proxyCalls)
}

func TestChatCompletionHandler_RejectsVirtualModelUnsupportedMode(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	r := newTestRouter(t, logger)
	handler := proxy.NewChatCompletionHandler(
		r,
		&mockTunnelTransport{},
		&mockBilling{cost: platform.UsageCostBreakdown{TotalCostUSD: 0.001, TotalTokens: 100}},
		&mockWalletStore{},
		&mockLogger{},
		logger,
	)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(auth.AuthContextKey{}, &platform.AuthContext{
			TokenHash:     "test-hash",
			TenantID:      "tenant-1",
			AllowedModels: []string{"creative/myCustomModel"},
			AllowedModelTypes: map[string][]string{
				"creative/myCustomModel": {"moderation"},
			},
			AllowedModelPricingMode: map[string]string{
				"creative/myCustomModel": "virtual",
			},
		})

		return c.Next()
	})
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(tenant.ContextKey{}, &platform.Tenant{ID: "tenant-1", Slug: "test-tenant"})
		return c.Next()
	})
	app.Post("/v1/chat/completions", handler.Handle)

	payload, _ := json.Marshal(map[string]interface{}{
		"model":    "creative/myCustomModel",
		"messages": []interface{}{map[string]interface{}{"role": "user", "content": "Hello"}},
	})
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var apiErr contractsopenai.APIError
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&apiErr))
	assert.Equal(t, "unsupported_model_mode", apiErr.Error.Code)
	assert.Equal(t, "chat", apiErr.Error.Details["requested_mode"])
	assert.Equal(t, "creative/myCustomModel", apiErr.Error.Details["model"])
	assert.Equal(t, "requested mode is not available on this virtual model", apiErr.Error.Details["reason"])
}

func TestChatCompletionHandler_ShouldTransformProviderError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	r := newTestRouter(t, logger)

	errorBody := []byte("Internal Server Error - upstream failure")
	mTunnel := &mockTunnelTransport{
		response: &platform.TunnelResponse{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"X-Ratelimit-Limit":     "100",
				"X-Ratelimit-Remaining": "0",
				"X-Trace-Id":            "trace-123",
			},
			Body: errorBody,
		},
	}
	mBilling := &mockBilling{cost: platform.UsageCostBreakdown{TotalCostUSD: 0.001, TotalTokens: 100}}
	mWallet := &mockWalletStore{}
	mLogs := &mockLogger{}

	handler := proxy.NewChatCompletionHandler(r, mTunnel, mBilling, mWallet, mLogs, logger)
	app := setupTestApp(t, handler)

	body := map[string]interface{}{
		"model":    "gpt-4",
		"messages": []interface{}{map[string]interface{}{"role": "user", "content": "Hello"}},
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	var apiErr contractsopenai.APIError

	err = json.NewDecoder(resp.Body).Decode(&apiErr)
	require.NoError(t, err)
	assert.Equal(t, "server_error", apiErr.Error.Type)
	assert.Contains(t, apiErr.Error.Message, "Internal Server Error")

	// Verify provider headers were forwarded
	assert.Equal(t, "100", resp.Header.Get("X-Ratelimit-Limit"))
	assert.Equal(t, "0", resp.Header.Get("X-Ratelimit-Remaining"))
	assert.Equal(t, "trace-123", resp.Header.Get("X-Trace-Id"))
}

func TestChatCompletionHandler_LogsProviderAuthenticationFailures(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	r := newTestRouter(t, logger)

	errorBody := []byte(`{"error":{"message":"invalid x-api-key","type":"authentication_error"}}`)
	mTunnel := &mockTunnelTransport{
		response: &platform.TunnelResponse{
			StatusCode: http.StatusUnauthorized,
			Headers:    map[string]string{},
			Body:       errorBody,
		},
	}
	mBilling := &mockBilling{cost: platform.UsageCostBreakdown{TotalCostUSD: 0.001, TotalTokens: 100}}
	mWallet := &mockWalletStore{}
	mLogs := &mockLogger{}

	handler := proxy.NewChatCompletionHandler(r, mTunnel, mBilling, mWallet, mLogs, logger)
	app := setupTestApp(t, handler)

	body := map[string]interface{}{
		"model":    "gpt-4",
		"messages": []interface{}{map[string]interface{}{"role": "user", "content": "Hello"}},
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	require.Len(t, mLogs.loggedEntries, 1)

	entry := mLogs.loggedEntries[0]
	assert.Equal(t, "failed", entry.Status)
	assert.Equal(t, http.StatusUnauthorized, entry.StatusCode)
	assert.Equal(t, "authentication_error", entry.ErrorType)
	assert.Equal(t, "/v1/chat/completions", entry.Route)
	assert.Equal(t, "chat_completions", entry.RequestType)
	assert.Equal(t, "gpt-4", entry.Model)
	assert.Equal(t, map[string]interface{}{"model": "gpt-4"}, entry.RequestBody)
	assert.Equal(t, map[string]interface{}{"error": "invalid x-api-key"}, entry.ResponseBody)
	assert.Empty(t, mLogs.loggedSpends)
}

func TestChatCompletionHandler_ShouldTransformRateLimitError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	r := newTestRouter(t, logger)

	mTunnel := &mockTunnelTransport{
		response: &platform.TunnelResponse{
			StatusCode: http.StatusTooManyRequests,
			Headers:    map[string]string{},
			Body:       []byte("Rate limit exceeded"),
		},
	}
	mBilling := &mockBilling{cost: platform.UsageCostBreakdown{TotalCostUSD: 0.001, TotalTokens: 100}}
	mWallet := &mockWalletStore{}
	mLogs := &mockLogger{}

	handler := proxy.NewChatCompletionHandler(r, mTunnel, mBilling, mWallet, mLogs, logger)
	app := setupTestApp(t, handler)

	body := map[string]interface{}{
		"model":    "gpt-4",
		"messages": []interface{}{map[string]interface{}{"role": "user", "content": "Hello"}},
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)

	var apiErr contractsopenai.APIError

	err = json.NewDecoder(resp.Body).Decode(&apiErr)
	require.NoError(t, err)
	assert.Equal(t, "rate_limit_error", apiErr.Error.Type)
}

func TestChatCompletionHandler_ShouldPassThroughValidProviderHeaders(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	r := newTestRouter(t, logger)

	validResponse := contractsopenai.ChatCompletionResponse{
		ID:      "chatcmpl-123",
		Object:  "chat.completion",
		Created: 1234567890,
		Model:   "gpt-4",
		Choices: []contractsopenai.Choice{
			{
				Index: 0,
				Message: &contractsopenai.ChatMessage{
					Role:    "assistant",
					Content: "Hello!",
				},
				FinishReason: "stop",
			},
		},
		Usage: &contractsopenai.Usage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}
	respBody, _ := json.Marshal(validResponse)

	mTunnel := &mockTunnelTransport{
		response: &platform.TunnelResponse{
			StatusCode: http.StatusOK,
			Headers: map[string]string{
				"X-Ratelimit-Limit":        "100",
				"X-Ratelimit-Remaining":    "99",
				"X-Request-Id":             "req-456",
				"X-Trace-Id":               "trace-789",
				"X-Correlation-Id":         "corr-012",
				"Cf-Ray":                   "abc123",
				"Traceparent":              "00-123-456-01",
				"X-Some-Other-Header":      "should-not-pass",
				"X-Internal-Secret-Header": "should-not-pass",
			},
			Body: respBody,
		},
	}
	mBilling := &mockBilling{cost: platform.UsageCostBreakdown{TotalCostUSD: 0.001, TotalTokens: 15}}
	mWallet := &mockWalletStore{}
	mLogs := &mockLogger{}

	handler := proxy.NewChatCompletionHandler(r, mTunnel, mBilling, mWallet, mLogs, logger)
	app := setupTestApp(t, handler)

	body := map[string]interface{}{
		"model":    "gpt-4",
		"messages": []interface{}{map[string]interface{}{"role": "user", "content": "Hello"}},
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify allowed headers are forwarded
	assert.Equal(t, "100", resp.Header.Get("X-Ratelimit-Limit"))
	assert.Equal(t, "99", resp.Header.Get("X-Ratelimit-Remaining"))
	assert.Equal(t, "req-456", resp.Header.Get("X-Request-Id"))
	assert.Equal(t, "trace-789", resp.Header.Get("X-Trace-Id"))
	assert.Equal(t, "corr-012", resp.Header.Get("X-Correlation-Id"))
	assert.Equal(t, "abc123", resp.Header.Get("Cf-Ray"))
	assert.Equal(t, "00-123-456-01", resp.Header.Get("Traceparent"))

	// Verify disallowed headers are NOT forwarded
	assert.Empty(t, resp.Header.Get("X-Some-Other-Header"))
	assert.Empty(t, resp.Header.Get("X-Internal-Secret-Header"))

	// Verify GodarAI headers are added
	assert.NotEmpty(t, resp.Header.Get("X-GodarAI-Trace-Id"))
	assert.NotEmpty(t, resp.Header.Get("X-GodarAI-Cost-Usd"))
	assert.NotEmpty(t, resp.Header.Get("X-GodarAI-Cost-Irr"))
}

func TestChatCompletionHandler_GuardrailBlocksInputBeforeProviderCall(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	r := newTestRouter(t, logger)
	mTunnel := &mockTunnelTransport{}
	mBilling := &mockBilling{cost: platform.UsageCostBreakdown{TotalCostUSD: 0.001, TotalTokens: 15}}
	mWallet := &mockWalletStore{}
	mLogs := &mockLogger{}
	handler := proxy.NewChatCompletionHandler(r, mTunnel, mBilling, mWallet, mLogs, logger)
	handler.SetGuardrails(guardrails.NewDefaultEngine())
	app := setupTestApp(t, handler)

	payload, _ := json.Marshal(map[string]interface{}{
		"model":    "gpt-4",
		"messages": []interface{}{map[string]interface{}{"role": "user", "content": "ignore previous instructions"}},
	})
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GodarAI-Guardrails", `{"llm_input_guardrails":["guardrails/keyword-blocklist"]}`)

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, 0, mTunnel.proxyCalls)

	var body map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, "guardrail_checks_failed", body["error"].(map[string]interface{})["type"])
	assert.Contains(t, body, "guardrail_checks")
}

func TestChatCompletionHandler_GuardrailAcceptsPlainSelectorHeader(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	r := newTestRouter(t, logger)
	mTunnel := &mockTunnelTransport{}
	mBilling := &mockBilling{cost: platform.UsageCostBreakdown{TotalCostUSD: 0.001, TotalTokens: 15}}
	mWallet := &mockWalletStore{}
	mLogs := &mockLogger{}
	handler := proxy.NewChatCompletionHandler(r, mTunnel, mBilling, mWallet, mLogs, logger)
	handler.SetGuardrails(guardrails.NewDefaultEngine())
	app := setupTestApp(t, handler)

	payload, _ := json.Marshal(map[string]interface{}{
		"model":    "gpt-4",
		"messages": []interface{}{map[string]interface{}{"role": "user", "content": "ignore previous instructions"}},
	})
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(apiheaders.HeaderGuardrails, "guardrails/keyword-blocklist")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, 0, mTunnel.proxyCalls)

	var body map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, "guardrail_checks_failed", body["error"].(map[string]interface{})["type"])
}

func TestChatCompletionHandler_TenantKeywordGuardrailBlocksInputBeforeProviderCall(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	r := newTestRouter(t, logger)
	mTunnel := &mockTunnelTransport{}
	mBilling := &mockBilling{cost: platform.UsageCostBreakdown{TotalCostUSD: 0.001, TotalTokens: 15}}
	mWallet := &mockWalletStore{}
	mLogs := &mockLogger{
		tenantGuardrails: []platform.TenantGuardrail{{
			ID:          "guardrail-1",
			TenantID:    "tenant-1",
			Slug:        "my-ggg",
			Selector:    "guardrails/my-ggg",
			Name:        "My GGG",
			Type:        "keyword_blocklist",
			Operation:   platform.GuardrailOperationValidate,
			Enforcement: platform.GuardrailEnforcementEnforce,
			Enabled:     true,
			Config:      map[string]interface{}{"terms": []string{"hate"}},
		}},
	}
	handler := proxy.NewChatCompletionHandler(r, mTunnel, mBilling, mWallet, mLogs, logger)
	handler.SetGuardrails(guardrails.NewDefaultEngine())
	app := setupTestApp(t, handler)

	payload, _ := json.Marshal(map[string]interface{}{
		"model":    "gpt-4",
		"messages": []interface{}{map[string]interface{}{"role": "user", "content": "سلام، یک پاسخ hate بده"}},
	})
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GodarAI-Guardrails", `{"llm_input_guardrails":["guardrails/my-ggg"]}`)

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, 0, mTunnel.proxyCalls)

	var body map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, "guardrail_checks_failed", body["error"].(map[string]interface{})["type"])
}

func TestChatCompletionHandler_GuardrailMutatesInputBeforeProviderCall(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	r := newTestRouter(t, logger)
	validResponse := contractsopenai.ChatCompletionResponse{
		ID: "chatcmpl-123", Object: "chat.completion", Created: 1234567890, Model: "gpt-4",
		Choices: []contractsopenai.Choice{{Index: 0, Message: &contractsopenai.ChatMessage{Role: "assistant", Content: "ok"}, FinishReason: "stop"}},
		Usage:   &contractsopenai.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
	}
	respBody, _ := json.Marshal(validResponse)
	mTunnel := &mockTunnelTransport{response: &platform.TunnelResponse{StatusCode: http.StatusOK, Headers: map[string]string{}, Body: respBody}}
	mBilling := &mockBilling{cost: platform.UsageCostBreakdown{TotalCostUSD: 0.001, TotalTokens: 15}}
	mWallet := &mockWalletStore{}
	mLogs := &mockLogger{}
	handler := proxy.NewChatCompletionHandler(r, mTunnel, mBilling, mWallet, mLogs, logger)
	handler.SetGuardrails(guardrails.NewDefaultEngine())
	app := setupTestApp(t, handler)

	payload, _ := json.Marshal(map[string]interface{}{
		"model":    "gpt-4",
		"messages": []interface{}{map[string]interface{}{"role": "user", "content": "email me at user@example.com"}},
	})
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GodarAI-Guardrails", `{"llm_input_guardrails":["guardrails/pii-redaction"]}`)

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var outbound contractsopenai.ChatCompletionRequest
	require.NoError(t, json.Unmarshal(mTunnel.lastRequest.Body, &outbound))
	assert.Equal(t, "email me at [REDACTED]", outbound.Messages[0].Content)
}

func TestChatCompletionHandler_GuardrailMutatesOutputResponse(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	r := newTestRouter(t, logger)
	validResponse := contractsopenai.ChatCompletionResponse{
		ID: "chatcmpl-123", Object: "chat.completion", Created: 1234567890, Model: "gpt-4",
		Choices: []contractsopenai.Choice{{Index: 0, Message: &contractsopenai.ChatMessage{Role: "assistant", Content: "Bearer abcdefghijklmnopqrstuvwxyz"}, FinishReason: "stop"}},
		Usage:   &contractsopenai.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
	}
	respBody, _ := json.Marshal(validResponse)
	mTunnel := &mockTunnelTransport{response: &platform.TunnelResponse{StatusCode: http.StatusOK, Headers: map[string]string{}, Body: respBody}}
	mBilling := &mockBilling{cost: platform.UsageCostBreakdown{TotalCostUSD: 0.001, TotalCostIRR: 850, TotalTokens: 15}}
	mWallet := &mockWalletStore{}
	mLogs := &mockLogger{}
	handler := proxy.NewChatCompletionHandler(r, mTunnel, mBilling, mWallet, mLogs, logger)
	handler.SetGuardrails(guardrails.NewDefaultEngine())
	app := setupTestApp(t, handler)

	payload, _ := json.Marshal(map[string]interface{}{
		"model":    "gpt-4",
		"messages": []interface{}{map[string]interface{}{"role": "user", "content": "hello"}},
	})
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(apiheaders.HeaderGuardrails, `{"output_guardrails":["guardrails/secrets-detection"]}`)

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var completion contractsopenai.ChatCompletionResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&completion))
	require.NotNil(t, completion.Choices[0].Message)
	assert.Equal(t, "[REDACTED]", completion.Choices[0].Message.Content)
}

func TestChatCompletionHandler_GuardrailRejectsStreamingWithOutputGuardrails(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	r := newTestRouter(t, logger)
	mTunnel := &mockTunnelTransport{}
	mBilling := &mockBilling{cost: platform.UsageCostBreakdown{TotalCostUSD: 0.001, TotalTokens: 15}}
	mWallet := &mockWalletStore{}
	mLogs := &mockLogger{}
	handler := proxy.NewChatCompletionHandler(r, mTunnel, mBilling, mWallet, mLogs, logger)
	handler.SetGuardrails(guardrails.NewDefaultEngine())
	app := setupTestApp(t, handler)

	payload, _ := json.Marshal(map[string]interface{}{
		"model":    "gpt-4",
		"stream":   true,
		"messages": []interface{}{map[string]interface{}{"role": "user", "content": "hello"}},
	})
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GodarAI-Guardrails", `{"llm_output_guardrails":["guardrails/secrets-detection"]}`)

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, 0, mTunnel.proxyCalls)
}

func TestChatCompletionHandler_ShouldSetRateLimitHeaders(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	r := newTestRouter(t, logger)

	validResponse := contractsopenai.ChatCompletionResponse{
		ID:      "chatcmpl-123",
		Object:  "chat.completion",
		Created: 1234567890,
		Model:   "gpt-4",
		Choices: []contractsopenai.Choice{
			{
				Index: 0,
				Message: &contractsopenai.ChatMessage{
					Role:    "assistant",
					Content: "Hello!",
				},
				FinishReason: "stop",
			},
		},
		Usage: &contractsopenai.Usage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}
	respBody, _ := json.Marshal(validResponse)

	mTunnel := &mockTunnelTransport{
		response: &platform.TunnelResponse{
			StatusCode: http.StatusOK,
			Headers:    map[string]string{},
			Body:       respBody,
		},
	}
	mBilling := &mockBilling{
		cost: platform.UsageCostBreakdown{TotalCostUSD: 0.001, TotalTokens: 15},
		rateLimitRes: &mcache.RateLimitResult{
			Allowed:    false,
			Current:    2,
			Limit:      10,
			ResetAt:    time.Now().Add(time.Minute),
			RetryAfter: true,
		},
	}
	mWallet := &mockWalletStore{}
	mLogs := &mockLogger{}

	handler := proxy.NewChatCompletionHandler(r, mTunnel, mBilling, mWallet, mLogs, logger)
	app := setupTestApp(t, handler)

	body := map[string]interface{}{
		"model":    "gpt-4",
		"messages": []interface{}{map[string]interface{}{"role": "user", "content": "Hello"}},
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "10", resp.Header.Get("X-Ratelimit-Limit"))
	assert.Equal(t, "8", resp.Header.Get("X-Ratelimit-Remaining"))
	assert.NotEmpty(t, resp.Header.Get("X-Ratelimit-Reset"))
	assert.NotEmpty(t, resp.Header.Get("Retry-After"))
}

func TestChatCompletionHandler_ShouldAcceptValidTemperature(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	r := newTestRouter(t, logger)

	validResponse := contractsopenai.ChatCompletionResponse{
		ID:      "chatcmpl-123",
		Object:  "chat.completion",
		Created: 1234567890,
		Model:   "gpt-4",
		Choices: []contractsopenai.Choice{
			{
				Index: 0,
				Message: &contractsopenai.ChatMessage{
					Role:    "assistant",
					Content: "Hello!",
				},
				FinishReason: "stop",
			},
		},
		Usage: &contractsopenai.Usage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}
	respBody, _ := json.Marshal(validResponse)

	mTunnel := &mockTunnelTransport{
		response: &platform.TunnelResponse{
			StatusCode: http.StatusOK,
			Headers:    map[string]string{},
			Body:       respBody,
		},
	}
	mBilling := &mockBilling{cost: platform.UsageCostBreakdown{TotalCostUSD: 0.001, TotalTokens: 15}}
	mWallet := &mockWalletStore{}
	mLogs := &mockLogger{}

	handler := proxy.NewChatCompletionHandler(r, mTunnel, mBilling, mWallet, mLogs, logger)
	app := setupTestApp(t, handler)

	validTemps := []float64{0, 0.5, 1.0, 1.5, 2.0}
	for _, temp := range validTemps {
		t.Run("", func(t *testing.T) {
			body := map[string]interface{}{
				"model":       "gpt-4",
				"messages":    []interface{}{map[string]interface{}{"role": "user", "content": "Hello"}},
				"temperature": temp,
			}
			payload, _ := json.Marshal(body)

			req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(payload))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			defer closeResponseBody(t, resp)

			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestChatCompletionHandler_ShouldAcceptValidTopP(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	r := newTestRouter(t, logger)

	validResponse := contractsopenai.ChatCompletionResponse{
		ID:      "chatcmpl-123",
		Object:  "chat.completion",
		Created: 1234567890,
		Model:   "gpt-4",
		Choices: []contractsopenai.Choice{
			{
				Index: 0,
				Message: &contractsopenai.ChatMessage{
					Role:    "assistant",
					Content: "Hello!",
				},
				FinishReason: "stop",
			},
		},
		Usage: &contractsopenai.Usage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}
	respBody, _ := json.Marshal(validResponse)

	mTunnel := &mockTunnelTransport{
		response: &platform.TunnelResponse{
			StatusCode: http.StatusOK,
			Headers:    map[string]string{},
			Body:       respBody,
		},
	}
	mBilling := &mockBilling{cost: platform.UsageCostBreakdown{TotalCostUSD: 0.001, TotalTokens: 15}}
	mWallet := &mockWalletStore{}
	mLogs := &mockLogger{}

	handler := proxy.NewChatCompletionHandler(r, mTunnel, mBilling, mWallet, mLogs, logger)
	app := setupTestApp(t, handler)

	validTopPs := []float64{0, 0.5, 1.0}
	for _, topP := range validTopPs {
		t.Run("", func(t *testing.T) {
			body := map[string]interface{}{
				"model":    "gpt-4",
				"messages": []interface{}{map[string]interface{}{"role": "user", "content": "Hello"}},
				"top_p":    topP,
			}
			payload, _ := json.Marshal(body)

			req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(payload))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			defer closeResponseBody(t, resp)

			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestChatCompletionHandler_AcceptsMultimodalChatContent(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	r := newTestRouter(t, logger)

	validResponse := contractsopenai.ChatCompletionResponse{
		ID:      "chatcmpl-123",
		Object:  "chat.completion",
		Created: 1234567890,
		Model:   "gpt-4",
		Choices: []contractsopenai.Choice{{
			Index: 0,
			Message: &contractsopenai.ChatMessage{
				Role:    "assistant",
				Content: "It is a small icon.",
			},
			FinishReason: "stop",
		}},
		Usage: &contractsopenai.Usage{PromptTokens: 20, CompletionTokens: 6, TotalTokens: 26},
	}
	respBody, _ := json.Marshal(validResponse)
	mTunnel := &mockTunnelTransport{response: &platform.TunnelResponse{StatusCode: http.StatusOK, Headers: map[string]string{}, Body: respBody}}
	mBilling := &mockBilling{cost: platform.UsageCostBreakdown{TotalCostUSD: 0.001, TotalTokens: 26}}
	mWallet := &mockWalletStore{}
	mLogs := &mockLogger{}

	handler := proxy.NewChatCompletionHandler(r, mTunnel, mBilling, mWallet, mLogs, logger)
	app := setupTestApp(t, handler)

	payload := []byte(`{
		"model":"gpt-4",
		"messages":[{
			"role":"user",
			"content":[
				{"type":"text","text":"Describe this image."},
				{"type":"image_url","image_url":{"url":"data:image/png;base64,AAAA"}}
			]
		}]
	}`)
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(mTunnel.lastRequest.Body), `"image_url"`)
}

func TestChatCompletionHandler_AcceptsFileChatContent(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	r := newTestRouter(t, logger)

	validResponse := contractsopenai.ChatCompletionResponse{
		ID:      "chatcmpl-123",
		Object:  "chat.completion",
		Created: 1234567890,
		Model:   "gpt-4",
		Choices: []contractsopenai.Choice{{
			Index: 0,
			Message: &contractsopenai.ChatMessage{
				Role:    "assistant",
				Content: "The PDF is a short guide.",
			},
			FinishReason: "stop",
		}},
		Usage: &contractsopenai.Usage{PromptTokens: 80, CompletionTokens: 8, TotalTokens: 88},
	}
	respBody, _ := json.Marshal(validResponse)
	mTunnel := &mockTunnelTransport{response: &platform.TunnelResponse{StatusCode: http.StatusOK, Headers: map[string]string{}, Body: respBody}}
	mBilling := &mockBilling{cost: platform.UsageCostBreakdown{TotalCostUSD: 0.001, TotalTokens: 88}}
	mWallet := &mockWalletStore{}
	mLogs := &mockLogger{}

	handler := proxy.NewChatCompletionHandler(r, mTunnel, mBilling, mWallet, mLogs, logger)
	app := setupTestApp(t, handler)

	payload := []byte(`{
		"model":"gpt-4",
		"messages":[{
			"role":"user",
			"content":[
				{"type":"text","text":"Summarize this PDF."},
				{"type":"file","file":{"filename":"guide.pdf","file_data":"data:application/pdf;base64,JVBERi0x"}}
			]
		}]
	}`)
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(mTunnel.lastRequest.Body), `"type":"file"`)
	assert.Contains(t, string(mTunnel.lastRequest.Body), `"file_data":"data:application/pdf;base64,JVBERi0x"`)
}

func TestChatCompletionHandler_RoundTripsOpenAICompatibleTools(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	r := newTestRouter(t, logger)

	validResponse := contractsopenai.ChatCompletionResponse{
		ID:      "chatcmpl-tools",
		Object:  "chat.completion",
		Created: 1234567890,
		Model:   "gpt-4",
		Choices: []contractsopenai.Choice{{
			Index: 0,
			Message: &contractsopenai.ChatMessage{
				Role: "assistant",
				ToolCalls: []contractsopenai.ToolCall{{
					ID:   "call_123",
					Type: "function",
					Function: &contractsopenai.FunctionCall{
						Name:      "get_order_status",
						Arguments: `{"order_id":"A-1024"}`,
					},
				}},
			},
			FinishReason: "tool_calls",
		}},
		Usage: &contractsopenai.Usage{PromptTokens: 40, CompletionTokens: 12, TotalTokens: 52},
	}
	respBody, _ := json.Marshal(validResponse)
	mTunnel := &mockTunnelTransport{response: &platform.TunnelResponse{StatusCode: http.StatusOK, Headers: map[string]string{}, Body: respBody}}
	mBilling := &mockBilling{cost: platform.UsageCostBreakdown{TotalCostUSD: 0.001, TotalTokens: 52}}
	mWallet := &mockWalletStore{}
	mLogs := &mockLogger{}

	handler := proxy.NewChatCompletionHandler(r, mTunnel, mBilling, mWallet, mLogs, logger)
	app := setupTestApp(t, handler)

	payload := []byte(`{
		"model":"gpt-4",
		"parallel_tool_calls":false,
		"messages":[
			{"role":"user","content":"What is order A-1024?"},
			{"role":"assistant","tool_calls":[{"id":"call_prev","type":"function","function":{"name":"get_order_status","arguments":"{\"order_id\":\"A-1023\"}"}}]},
			{"role":"tool","tool_call_id":"call_prev","content":"{\"status\":\"delivered\"}"}
		],
		"tools":[{
			"type":"function",
			"function":{
				"name":"get_order_status",
				"description":"Get order status.",
				"strict":true,
				"parameters":{"type":"object","properties":{"order_id":{"type":"string"}},"required":["order_id"],"additionalProperties":false}
			}
		}],
		"tool_choice":{"type":"function","function":{"name":"get_order_status"}}
	}`)
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	outbound := string(mTunnel.lastRequest.Body)
	assert.Contains(t, outbound, `"tools"`)
	assert.Contains(t, outbound, `"tool_choice"`)
	assert.Contains(t, outbound, `"tool_calls"`)
	assert.Contains(t, outbound, `"tool_call_id":"call_prev"`)
	assert.Contains(t, outbound, `"parallel_tool_calls":false`)
	assert.Contains(t, outbound, `"strict":true`)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	bodyText := string(body)
	assert.Contains(t, bodyText, `"tool_calls"`)
	assert.Contains(t, bodyText, `"finish_reason":"tool_calls"`)
	assert.Contains(t, bodyText, `"name":"get_order_status"`)
}

func TestChatCompletionHandler_RoundTripsStructuredOutputsResponseFormat(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	r := newTestRouter(t, logger)

	validResponse := contractsopenai.ChatCompletionResponse{
		ID:      "chatcmpl-structured",
		Object:  "chat.completion",
		Created: 1234567890,
		Model:   "gpt-4",
		Choices: []contractsopenai.Choice{{
			Index: 0,
			Message: &contractsopenai.ChatMessage{
				Role:    "assistant",
				Content: `{"summary":"ok","priority":"high"}`,
			},
			FinishReason: "stop",
		}},
		Usage: &contractsopenai.Usage{PromptTokens: 40, CompletionTokens: 12, TotalTokens: 52},
	}
	respBody, _ := json.Marshal(validResponse)
	mTunnel := &mockTunnelTransport{response: &platform.TunnelResponse{StatusCode: http.StatusOK, Headers: map[string]string{}, Body: respBody}}
	mBilling := &mockBilling{cost: platform.UsageCostBreakdown{TotalCostUSD: 0.001, TotalTokens: 52}}
	mWallet := &mockWalletStore{}
	mLogs := &mockLogger{}

	handler := proxy.NewChatCompletionHandler(r, mTunnel, mBilling, mWallet, mLogs, logger)
	app := setupTestApp(t, handler)

	payload := []byte(`{
		"model":"gpt-4",
		"messages":[{"role":"user","content":"Summarize this ticket."}],
		"response_format":{
			"type":"json_schema",
			"json_schema":{
				"name":"support_ticket",
				"strict":true,
				"schema":{
					"type":"object",
					"properties":{"summary":{"type":"string"},"priority":{"type":"string","enum":["low","medium","high"]}},
					"required":["summary","priority"],
					"additionalProperties":false
				}
			}
		}
	}`)
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	outbound := string(mTunnel.lastRequest.Body)
	assert.Contains(t, outbound, `"response_format"`)
	assert.Contains(t, outbound, `"type":"json_schema"`)
	assert.Contains(t, outbound, `"name":"support_ticket"`)
	assert.Contains(t, outbound, `"strict":true`)
	assert.Contains(t, outbound, `"additionalProperties":false`)
}

func TestChatCompletionHandler_PreservesStructuredOutputRefusal(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	r := newTestRouter(t, logger)

	validResponse := contractsopenai.ChatCompletionResponse{
		ID:      "chatcmpl-refusal",
		Object:  "chat.completion",
		Created: 1234567890,
		Model:   "gpt-4",
		Choices: []contractsopenai.Choice{{
			Index: 0,
			Message: &contractsopenai.ChatMessage{
				Role:    "assistant",
				Refusal: "I cannot help with that request.",
			},
			FinishReason: "stop",
		}},
		Usage: &contractsopenai.Usage{PromptTokens: 40, CompletionTokens: 12, TotalTokens: 52},
	}
	respBody, _ := json.Marshal(validResponse)
	mTunnel := &mockTunnelTransport{response: &platform.TunnelResponse{StatusCode: http.StatusOK, Headers: map[string]string{}, Body: respBody}}
	mBilling := &mockBilling{cost: platform.UsageCostBreakdown{TotalCostUSD: 0.001, TotalTokens: 52}}
	mWallet := &mockWalletStore{}
	mLogs := &mockLogger{}

	handler := proxy.NewChatCompletionHandler(r, mTunnel, mBilling, mWallet, mLogs, logger)
	app := setupTestApp(t, handler)

	payload := []byte(`{
		"model":"gpt-4",
		"messages":[{"role":"user","content":"Extract the data."}],
		"response_format":{"type":"json_object"}
	}`)
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), `"refusal":"I cannot help with that request."`)
}

func TestChatCompletionHandler_RoundTripsOpenAICompatibleExtendedThinking(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	r := newTestRouter(t, logger)

	providerBody := []byte(`{
		"id":"chatcmpl-thinking",
		"object":"chat.completion",
		"created":1234567890,
		"model":"gpt-5",
		"choices":[{
			"index":0,
			"message":{
				"role":"assistant",
				"content":"Final answer",
				"reasoning_content":"private reasoning",
				"thinking_blocks":[{"type":"thinking","thinking":"private reasoning","signature":"sig_123"}]
			},
			"finish_reason":"stop"
		}],
		"usage":{"prompt_tokens":40,"completion_tokens":12,"total_tokens":52}
	}`)
	mTunnel := &mockTunnelTransport{response: &platform.TunnelResponse{StatusCode: http.StatusOK, Headers: map[string]string{}, Body: providerBody}}
	mBilling := &mockBilling{cost: platform.UsageCostBreakdown{TotalCostUSD: 0.001, TotalTokens: 52}}
	mWallet := &mockWalletStore{}
	mLogs := &mockLogger{}

	handler := proxy.NewChatCompletionHandler(r, mTunnel, mBilling, mWallet, mLogs, logger)
	app := setupTestApp(t, handler)

	payload := []byte(`{
		"model":"gpt-4",
		"reasoning_effort":"high",
		"messages":[
			{"role":"user","content":"Solve this."},
			{"role":"assistant","content":"Previous answer","reasoning_content":"previous reasoning","thinking_blocks":[{"type":"thinking","thinking":"previous reasoning","signature":"sig_prev"}]},
			{"role":"user","content":"Continue."}
		]
	}`)
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	outbound := string(mTunnel.lastRequest.Body)
	assert.Contains(t, outbound, `"reasoning_effort":"high"`)
	assert.Contains(t, outbound, `"thinking_blocks"`)
	assert.Contains(t, outbound, `"signature":"sig_prev"`)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	bodyText := string(body)
	assert.Contains(t, bodyText, `"reasoning_content":"private reasoning"`)
	assert.Contains(t, bodyText, `"thinking_blocks"`)
	assert.Contains(t, bodyText, `"signature":"sig_123"`)
}

func TestChatCompletionHandler_RejectsExtendedThinkingForUnsupportedProvider(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, err := mcache.NewMemoryCache(logger)
	require.NoError(t, err)
	r, err := llmrouter.NewRouter([]config.ModelDeployment{{
		ModelName: "command-r",
		ProviderEndpoint: config.ProviderEndpointConfig{
			Provider: "cohere",
			Model:    "cohere/command-r",
			RPM:      1000,
		},
	}}, config.RouterSettings{RoutingStrategy: "simple-shuffle", CooldownTime: 60}, memCache, nil, logger)
	require.NoError(t, err)

	mTunnel := &mockTunnelTransport{}
	mBilling := &mockBilling{cost: platform.UsageCostBreakdown{TotalCostUSD: 0.001, TotalTokens: 52}}
	mWallet := &mockWalletStore{}
	mLogs := &mockLogger{}
	handler := proxy.NewChatCompletionHandler(r, mTunnel, mBilling, mWallet, mLogs, logger)
	app := setupTestApp(t, handler)

	payload := []byte(`{
		"model":"command-r",
		"messages":[{"role":"user","content":"Solve this."}],
		"reasoning_effort":"high"
	}`)
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	assert.Zero(t, mTunnel.proxyCalls)

	var apiErr contractsopenai.APIError
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&apiErr))
	assert.Contains(t, apiErr.Error.Message, "does not support extended thinking")
}

func TestChatCompletionHandler_RejectsImageInputForUnsupportedProvider(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, err := mcache.NewMemoryCache(logger)
	require.NoError(t, err)
	r, err := llmrouter.NewRouter([]config.ModelDeployment{{
		ModelName: "command-r",
		ProviderEndpoint: config.ProviderEndpointConfig{
			Provider: "cohere",
			Model:    "cohere/command-r",
			RPM:      1000,
		},
	}}, config.RouterSettings{RoutingStrategy: "simple-shuffle", CooldownTime: 60}, memCache, nil, logger)
	require.NoError(t, err)

	mTunnel := &mockTunnelTransport{}
	mBilling := &mockBilling{cost: platform.UsageCostBreakdown{TotalCostUSD: 0.001, TotalTokens: 26}}
	mWallet := &mockWalletStore{}
	mLogs := &mockLogger{}
	handler := proxy.NewChatCompletionHandler(r, mTunnel, mBilling, mWallet, mLogs, logger)
	app := setupTestApp(t, handler)

	payload := []byte(`{
		"model":"command-r",
		"messages":[{
			"role":"user",
			"content":[
				{"type":"text","text":"Describe this image."},
				{"type":"image_url","image_url":{"url":"data:image/png;base64,AAAA"}}
			]
		}]
	}`)
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var apiErr contractsopenai.APIError
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&apiErr))
	assert.Contains(t, apiErr.Error.Message, "does not support image input")
}

func TestChatCompletionHandler_RejectsFileInputForUnsupportedProvider(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, err := mcache.NewMemoryCache(logger)
	require.NoError(t, err)
	r, err := llmrouter.NewRouter([]config.ModelDeployment{{
		ModelName: "command-r",
		ProviderEndpoint: config.ProviderEndpointConfig{
			Provider: "cohere",
			Model:    "cohere/command-r",
			RPM:      1000,
		},
	}}, config.RouterSettings{RoutingStrategy: "simple-shuffle", CooldownTime: 60}, memCache, nil, logger)
	require.NoError(t, err)

	mTunnel := &mockTunnelTransport{}
	mBilling := &mockBilling{cost: platform.UsageCostBreakdown{TotalCostUSD: 0.001, TotalTokens: 88}}
	mWallet := &mockWalletStore{}
	mLogs := &mockLogger{}
	handler := proxy.NewChatCompletionHandler(r, mTunnel, mBilling, mWallet, mLogs, logger)
	app := setupTestApp(t, handler)

	payload := []byte(`{
		"model":"command-r",
		"messages":[{
			"role":"user",
			"content":[
				{"type":"text","text":"Summarize this PDF."},
				{"type":"file","file":{"filename":"guide.pdf","file_data":"data:application/pdf;base64,JVBERi0x"}}
			]
		}]
	}`)
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var apiErr contractsopenai.APIError
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&apiErr))
	assert.Contains(t, apiErr.Error.Message, "does not support file input")
}

func TestChatCompletionHandler_RejectsFileIDForGeminiAdapter(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, err := mcache.NewMemoryCache(logger)
	require.NoError(t, err)
	r, err := llmrouter.NewRouter([]config.ModelDeployment{{
		ModelName: "gemini-1.5-flash",
		ProviderEndpoint: config.ProviderEndpointConfig{
			Provider: "gemini",
			Model:    "gemini/gemini-1.5-flash",
			RPM:      1000,
		},
	}}, config.RouterSettings{RoutingStrategy: "simple-shuffle", CooldownTime: 60}, memCache, nil, logger)
	require.NoError(t, err)

	mTunnel := &mockTunnelTransport{}
	mBilling := &mockBilling{cost: platform.UsageCostBreakdown{TotalCostUSD: 0.001, TotalTokens: 88}}
	mWallet := &mockWalletStore{}
	mLogs := &mockLogger{}
	handler := proxy.NewChatCompletionHandler(r, mTunnel, mBilling, mWallet, mLogs, logger)
	app := setupTestApp(t, handler)

	payload := []byte(`{
		"model":"gemini-1.5-flash",
		"messages":[{
			"role":"user",
			"content":[
				{"type":"text","text":"Summarize this file."},
				{"type":"file","file":{"file_id":"file_123"}}
			]
		}]
	}`)
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var apiErr contractsopenai.APIError
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&apiErr))
	assert.Contains(t, apiErr.Error.Message, "file.file_id is not supported")
}

func TestChatCompletionHandler_RejectsToolsForNativeProviderWithoutAdapter(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, err := mcache.NewMemoryCache(logger)
	require.NoError(t, err)
	r, err := llmrouter.NewRouter([]config.ModelDeployment{{
		ModelName: "claude-3-haiku",
		ProviderEndpoint: config.ProviderEndpointConfig{
			Provider: "anthropic",
			Model:    "anthropic/claude-3-haiku-20240307",
			RPM:      1000,
		},
	}}, config.RouterSettings{RoutingStrategy: "simple-shuffle", CooldownTime: 60}, memCache, nil, logger)
	require.NoError(t, err)

	mTunnel := &mockTunnelTransport{}
	mBilling := &mockBilling{cost: platform.UsageCostBreakdown{TotalCostUSD: 0.001, TotalTokens: 52}}
	mWallet := &mockWalletStore{}
	mLogs := &mockLogger{}
	handler := proxy.NewChatCompletionHandler(r, mTunnel, mBilling, mWallet, mLogs, logger)
	app := setupTestApp(t, handler)

	payload := []byte(`{
		"model":"claude-3-haiku",
		"messages":[{"role":"user","content":"What is order A-1024?"}],
		"tools":[{"type":"function","function":{"name":"get_order_status","parameters":{"type":"object"}}}]
	}`)
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	assert.Zero(t, mTunnel.proxyCalls)

	var apiErr contractsopenai.APIError
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&apiErr))
	assert.Contains(t, apiErr.Error.Message, "does not support tools")
}

func TestChatCompletionHandler_RejectsResponseFormatForNativeProviderWithoutAdapter(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, err := mcache.NewMemoryCache(logger)
	require.NoError(t, err)
	r, err := llmrouter.NewRouter([]config.ModelDeployment{{
		ModelName: "claude-3-haiku",
		ProviderEndpoint: config.ProviderEndpointConfig{
			Provider: "anthropic",
			Model:    "anthropic/claude-3-haiku-20240307",
			RPM:      1000,
		},
	}}, config.RouterSettings{RoutingStrategy: "simple-shuffle", CooldownTime: 60}, memCache, nil, logger)
	require.NoError(t, err)

	mTunnel := &mockTunnelTransport{}
	mBilling := &mockBilling{cost: platform.UsageCostBreakdown{TotalCostUSD: 0.001, TotalTokens: 52}}
	mWallet := &mockWalletStore{}
	mLogs := &mockLogger{}
	handler := proxy.NewChatCompletionHandler(r, mTunnel, mBilling, mWallet, mLogs, logger)
	app := setupTestApp(t, handler)

	payload := []byte(`{
		"model":"claude-3-haiku",
		"messages":[{"role":"user","content":"Return JSON."}],
		"response_format":{"type":"json_object"}
	}`)
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	assert.Zero(t, mTunnel.proxyCalls)

	var apiErr contractsopenai.APIError
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&apiErr))
	assert.Contains(t, apiErr.Error.Message, "does not support response_format")
}

func TestChatCompletionHandler_RejectsInvalidResponseFormatType(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	r := newTestRouter(t, logger)

	mTunnel := &mockTunnelTransport{}
	mBilling := &mockBilling{cost: platform.UsageCostBreakdown{TotalCostUSD: 0.001, TotalTokens: 52}}
	mWallet := &mockWalletStore{}
	mLogs := &mockLogger{}
	handler := proxy.NewChatCompletionHandler(r, mTunnel, mBilling, mWallet, mLogs, logger)
	app := setupTestApp(t, handler)

	payload := []byte(`{
		"model":"gpt-4",
		"messages":[{"role":"user","content":"Return JSON."}],
		"response_format":{"type":"xml"}
	}`)
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	assert.Zero(t, mTunnel.proxyCalls)

	var apiErr contractsopenai.APIError
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&apiErr))
	assert.Contains(t, apiErr.Error.Message, `response_format.type "xml" is not supported`)
}

func TestChatCompletionHandler_StreamSuccessRelaysChunksWithoutPanic(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	r := newTestRouter(t, logger)

	stream := newFakeTunnelStream(map[string]string{"content-type": "text/event-stream"}, 2, 0)
	stream.chunks <- platform.TunnelChunk{Data: []byte("data: {\"id\":\"chatcmpl-stream\",\"object\":\"chat.completion.chunk\",\"created\":1234567890,\"model\":\"gpt-4\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\"Salam\"}}]}\n\n")}

	stream.chunks <- platform.TunnelChunk{Data: []byte("data: [DONE]\n\n"), IsDone: true}

	close(stream.chunks)

	mTunnel := &mockTunnelTransport{stream: stream}
	mBilling := &mockBilling{cost: platform.UsageCostBreakdown{TotalCostUSD: 0.001, TotalCostIRR: 850, TotalTokens: 15}}
	mWallet := &mockWalletStore{}
	mLogs := &mockLogger{}

	handler := proxy.NewChatCompletionHandler(r, mTunnel, mBilling, mWallet, mLogs, logger)
	app := setupTestApp(t, handler)

	payload, _ := json.Marshal(map[string]interface{}{
		"model":    "gpt-4",
		"stream":   true,
		"messages": []interface{}{map[string]interface{}{"role": "user", "content": "Hello"}},
	})
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, int((5 * time.Second).Milliseconds()))
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	bodyText := string(body)
	assert.Contains(t, bodyText, `"content":"Salam"`)
	assert.Contains(t, bodyText, "data: [DONE]")
	assert.True(t, stream.closed)
	require.Len(t, mWallet.debitCalls, 1)
	require.Len(t, mLogs.loggedSpends, 1)
	assert.Equal(t, "success", mLogs.loggedSpends[0].Status)
}

func TestChatCompletionHandler_StreamCloseFinalizesWithoutPanic(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	r := newTestRouter(t, logger)

	stream := newFakeTunnelStream(map[string]string{"content-type": "text/event-stream"}, 1, 0)
	stream.chunks <- platform.TunnelChunk{Data: []byte("data: {\"id\":\"chatcmpl-stream\",\"object\":\"chat.completion.chunk\",\"created\":1234567890,\"model\":\"gpt-4\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\"Salam\"}}]}\n\n")}

	close(stream.chunks)

	mTunnel := &mockTunnelTransport{stream: stream}
	mBilling := &mockBilling{cost: platform.UsageCostBreakdown{TotalCostUSD: 0.001, TotalCostIRR: 850, TotalTokens: 15}}
	mWallet := &mockWalletStore{}
	mLogs := &mockLogger{}

	handler := proxy.NewChatCompletionHandler(r, mTunnel, mBilling, mWallet, mLogs, logger)
	app := setupTestApp(t, handler)

	payload, _ := json.Marshal(map[string]interface{}{
		"model":    "gpt-4",
		"stream":   true,
		"messages": []interface{}{map[string]interface{}{"role": "user", "content": "Hello"}},
	})
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, int((5 * time.Second).Milliseconds()))
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), `"content":"Salam"`)
	assert.NotContains(t, strings.ToLower(string(body)), "panic")
	assert.Len(t, mWallet.debitCalls, 1)
	assert.Len(t, mLogs.loggedSpends, 1)
	assert.True(t, stream.closed)
}

func TestChatCompletionHandler_StreamErrorWritesStreamingErrorEvent(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	r := newTestRouter(t, logger)

	stream := newFakeTunnelStream(map[string]string{"content-type": "text/event-stream"}, 0, 1)
	stream.errors <- errors.New("upstream stream failed")

	mTunnel := &mockTunnelTransport{stream: stream}
	mBilling := &mockBilling{cost: platform.UsageCostBreakdown{TotalCostUSD: 0.001, TotalCostIRR: 850, TotalTokens: 15}}
	mWallet := &mockWalletStore{}
	mLogs := &mockLogger{}

	handler := proxy.NewChatCompletionHandler(r, mTunnel, mBilling, mWallet, mLogs, logger)
	app := setupTestApp(t, handler)

	payload, _ := json.Marshal(map[string]interface{}{
		"model":    "gpt-4",
		"stream":   true,
		"messages": []interface{}{map[string]interface{}{"role": "user", "content": "Hello"}},
	})
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, int((5 * time.Second).Milliseconds()))
	defer closeResponseBody(t, resp)

	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), `"type":"streaming_error"`)
	assert.True(t, stream.closed)
}

// Ensure mock types satisfy interfaces at compile time.
var _ platform.TunnelTransport = (*mockTunnelTransport)(nil)
