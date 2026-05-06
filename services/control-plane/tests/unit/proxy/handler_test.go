package proxy_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	mcache "github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/config"
	"github.com/Delpak/shahrah/services/control-plane/internal/proxy"
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
}

func (m *mockTunnelTransport) Proxy(ctx context.Context, req platform.TunnelRequest) (*platform.TunnelResponse, error) {
	m.lastRequest = req
	return m.response, m.proxyErr
}

func (m *mockTunnelTransport) Stream(ctx context.Context, req platform.TunnelRequest) (platform.TunnelStream, error) {
	m.lastRequest = req
	return m.stream, m.streamErr
}

func (m *mockTunnelTransport) Healthy(ctx context.Context) error { return nil }

// mockBilling implements budgeter interface for testing.
type mockBilling struct {
	cost         platform.UsageCostBreakdown
	estimateErr  error
	rateLimitErr error
	rateLimitRes *mcache.RateLimitResult
	budgetErr    error
}

func (m *mockBilling) EstimateCost(ctx context.Context, req contractsopenai.ChatCompletionRequest) (platform.UsageCostBreakdown, error) {
	return m.cost, m.estimateErr
}

func (m *mockBilling) CheckRateLimit(ctx context.Context, auth *platform.AuthContext, model string, estimatedTokens int) (*mcache.RateLimitResult, error) {
	return m.rateLimitRes, m.rateLimitErr
}

func (m *mockBilling) CheckBudget(ctx context.Context, auth *platform.AuthContext, breakdown platform.UsageCostBreakdown) error {
	return m.budgetErr
}

func (m *mockBilling) FinalizeFromUsage(ctx context.Context, model string, usage *contractsopenai.Usage, estimated platform.UsageCostBreakdown) (platform.UsageCostBreakdown, error) {
	return m.cost, nil
}

// mockWalletStore implements walletStore interface.
type mockWalletStore struct {
	transaction *platform.Transaction
	err         error
}

func (m *mockWalletStore) DebitWallet(ctx context.Context, tenantID string, breakdown platform.UsageCostBreakdown, requestID string) (*platform.Transaction, error) {
	return m.transaction, m.err
}

// mockLogger implements usageLogger interface.
type mockLogger struct {
	loggedEntries []platform.RequestLogEntry
	loggedSpends  []platform.SpendLog
}

func (m *mockLogger) LogRequest(ctx context.Context, entry platform.RequestLogEntry) error {
	m.loggedEntries = append(m.loggedEntries, entry)
	return nil
}

func (m *mockLogger) LogSpend(ctx context.Context, log platform.SpendLog) error {
	m.loggedSpends = append(m.loggedSpends, log)
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
				Model: "openai/gpt-4",
				RPM:   1000,
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
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
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

	// Verify Shahrah headers are added
	assert.NotEmpty(t, resp.Header.Get("X-Shahrah-Trace-Id"))
	assert.NotEmpty(t, resp.Header.Get("X-Shahrah-Cost-Usd"))
	assert.NotEmpty(t, resp.Header.Get("X-Shahrah-Cost-Irr"))
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
			Allowed: true,
			Current: 2,
			Limit:   10,
			ResetAt: time.Now().Add(time.Minute),
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
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "10", resp.Header.Get("X-Ratelimit-Limit"))
	assert.Equal(t, "8", resp.Header.Get("X-Ratelimit-Remaining"))
	assert.NotEmpty(t, resp.Header.Get("X-Ratelimit-Reset"))
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
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

// Ensure mock types satisfy interfaces at compile time.
var _ platform.TunnelTransport = (*mockTunnelTransport)(nil)
