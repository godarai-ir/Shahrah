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
	"testing"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/testcache"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/apiheaders"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/config"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/Delpak/shahrah/services/control-plane/internal/responsecache"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type captureTunnel struct {
	resp      *platform.TunnelResponse
	responses []*platform.TunnelResponse
	stream    platform.TunnelStream
	streams   []platform.TunnelStream
	streamErr error
	lastReq   platform.TunnelRequest
	requests  []platform.TunnelRequest
}

type fixedResponsesCacheEmbedder struct{}

func (e *fixedResponsesCacheEmbedder) Embed(_ context.Context, texts []string) ([][]float32, error) {
	vectors := make([][]float32, len(texts))
	for index := range texts {
		vectors[index] = []float32{1, 0}
	}

	return vectors, nil
}

func (e *fixedResponsesCacheEmbedder) Identity() responsecache.EmbeddingIdentity {
	return responsecache.EmbeddingIdentity{
		Provider:   "test",
		Model:      "fixed",
		Revision:   "v1",
		Dimensions: 2,
	}
}

type responsesVirtualModelStore struct {
	model *platform.VirtualModel
}

func (s responsesVirtualModelStore) GetVirtualModelByAlias(_ context.Context, _, alias string) (*platform.VirtualModel, error) {
	if s.model == nil || s.model.Alias != alias {
		return nil, nil
	}

	return s.model, nil
}

type responsesVirtualModelResolver struct {
	model *platform.VirtualModel
	calls int
}

func (r *responsesVirtualModelResolver) Resolve(_ context.Context, _, _ string) (*platform.VirtualModel, error) {
	r.calls++

	return r.model, nil
}

type trackingResponsesRoutingEngine struct {
	calls int
}

func (e *trackingResponsesRoutingEngine) SelectTargets(_ *platform.VirtualModel) ([]platform.VirtualModelTarget, error) {
	e.calls++

	return nil, nil
}

type xaiResponsesBudgeter struct{ testBudgeter }

type deepSeekResponsesBudgeter struct{ testBudgeter }

type targetAwareResponsesBudgeter struct {
	testBudgeter
	costs             map[string]platform.UsageCostBreakdown
	estimateErrs      map[string]error
	estimateModels    []string
	budgetEstimate    *platform.UsageCostBreakdown
	finalizedEstimate *platform.UsageCostBreakdown
}

func (b *xaiResponsesBudgeter) EstimateCost(ctx context.Context, req contractsopenai.ChatCompletionRequest) (platform.UsageCostBreakdown, error) {
	estimated, err := b.testBudgeter.EstimateCost(ctx, req)
	if err != nil {
		return estimated, err
	}

	binding := xaiProxyTestRuntimeProfile()
	estimated.ModelRuntimeProfile = &binding
	estimated.PricingRevision = binding.CatalogRevision

	return estimated, nil
}

func (b *deepSeekResponsesBudgeter) EstimateCost(ctx context.Context, req contractsopenai.ChatCompletionRequest) (platform.UsageCostBreakdown, error) {
	estimated, err := b.testBudgeter.EstimateCost(ctx, req)
	if err != nil {
		return estimated, err
	}

	binding := deepSeekProxyRuntimeProfile(req.Model)
	estimated.ModelRuntimeProfile = &binding
	estimated.PricingRevision = binding.CatalogRevision

	return estimated, nil
}

func (b *targetAwareResponsesBudgeter) EstimateCost(
	ctx context.Context,
	req contractsopenai.ChatCompletionRequest,
) (platform.UsageCostBreakdown, error) {
	b.estimateModels = append(b.estimateModels, req.Model)

	if err, ok := b.estimateErrs[req.Model]; ok {
		return platform.UsageCostBreakdown{}, err
	}

	if cost, ok := b.costs[req.Model]; ok {
		return cost, nil
	}

	return b.testBudgeter.EstimateCost(ctx, req)
}

func (b *targetAwareResponsesBudgeter) CheckBudget(
	_ context.Context,
	_ *platform.AuthContext,
	estimated platform.UsageCostBreakdown,
) error {
	copyEstimate := estimated
	b.budgetEstimate = &copyEstimate

	return nil
}

func (b *targetAwareResponsesBudgeter) FinalizeFromUsage(
	ctx context.Context,
	model string,
	usage *contractsopenai.Usage,
	estimated platform.UsageCostBreakdown,
) (platform.UsageCostBreakdown, error) {
	copyEstimate := estimated
	b.finalizedEstimate = &copyEstimate

	return b.testBudgeter.FinalizeFromUsage(ctx, model, usage, estimated)
}

func (t *captureTunnel) Proxy(_ context.Context, req platform.TunnelRequest) (*platform.TunnelResponse, error) {
	t.lastReq = req

	t.requests = append(t.requests, req)
	if len(t.responses) > 0 {
		response := t.responses[0]
		t.responses = t.responses[1:]

		return response, nil
	}

	return t.resp, nil
}

func (t *captureTunnel) Stream(_ context.Context, req platform.TunnelRequest) (platform.TunnelStream, error) {
	t.lastReq = req

	t.requests = append(t.requests, req)
	if len(t.streams) > 0 {
		stream := t.streams[0]
		t.streams = t.streams[1:]

		return stream, t.streamErr
	}

	if t.stream != nil || t.streamErr != nil {
		return t.stream, t.streamErr
	}

	return nil, assert.AnError
}

func (t *captureTunnel) Healthy(_ context.Context) error { return nil }

func newResponsesTestApp(t *testing.T, provider string, tunnel *captureTunnel, logsOpt ...*testLogs) (*fiber.App, *tracing.Queue) {
	t.Helper()

	app, queue := newResponsesTestAppWithCache(t, provider, tunnel, nil, logsOpt...)

	return app, queue
}

func newResponsesTestAppWithCache(
	t *testing.T,
	provider string,
	tunnel *captureTunnel,
	cacheService *responsecache.Service,
	logsOpt ...*testLogs,
) (*fiber.App, *tracing.Queue) {
	t.Helper()

	mem, err := testcache.New(t, zap.NewNop())
	require.NoError(t, err)

	modelName := "gpt-4o"

	var billing budgeter = &testBudgeter{}

	if strings.EqualFold(provider, "xai") {
		modelName = "grok-4.5"
		billing = &xaiResponsesBudgeter{}
	}

	if strings.EqualFold(provider, "deepseek") {
		modelName = "deepseek-v4-flash"
		billing = &deepSeekResponsesBudgeter{}
	}

	deployments := []config.ModelDeployment{{
		ModelName: modelName,
		ProviderEndpoint: config.ProviderEndpointConfig{
			Provider: provider,
			Model:    modelName,
			APIKey:   "test-key",
		},
	}}
	if strings.EqualFold(provider, "deepseek") {
		deployments = append(deployments, config.ModelDeployment{
			ModelName: "deepseek-v4-pro",
			ProviderEndpoint: config.ProviderEndpointConfig{
				Provider: provider,
				Model:    "deepseek-v4-pro",
				APIKey:   "test-key",
			},
		})
	}

	if strings.EqualFold(provider, "openai") {
		deployments = append(deployments, config.ModelDeployment{
			ModelName: "gpt-6-sol",
			ProviderEndpoint: config.ProviderEndpointConfig{
				Provider: provider,
				Model:    "gpt-6-sol",
				APIKey:   "test-key",
			},
		})
	}

	rt, err := router.NewRouter(deployments, config.RouterSettings{CooldownTime: 60, NumRetries: 1}, mem, nil, zap.NewNop())
	require.NoError(t, err)

	logs := &testLogs{}
	if len(logsOpt) > 0 && logsOpt[0] != nil {
		logs = logsOpt[0]
	}

	h := NewResponsesHandlerWithDependencies(BaseHandlerDependencies{
		Router:              rt,
		Tunnel:              tunnel,
		Billing:             billing,
		BillingCapabilities: compatibilityBillingCapabilities(),
		RequestLogger:       logs,
		SpendLogger:         logs,
		Logger:              zap.NewNop(),
		ResponseCache:       cacheService,
	})
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

func newResponsesCacheService(t *testing.T, embedder responsecache.Embedder) *responsecache.Service {
	t.Helper()

	store, err := testcache.New(t, zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	service, err := responsecache.NewService(store, embedder, responsecache.Config{
		Enabled:               true,
		DefaultTTL:            time.Minute,
		MinTTL:                time.Second,
		MaxTTL:                time.Hour,
		MaxResponseBytes:      1 << 20,
		MaxSemanticCandidates: 10,
	})
	require.NoError(t, err)

	return service
}

func newVirtualResponsesTestApp(t *testing.T, tunnel *captureTunnel, logs *testLogs) *fiber.App {
	t.Helper()

	mem, err := testcache.New(t, zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, mem.Close()) })

	rt, err := router.NewRouter([]config.ModelDeployment{{
		ModelName: "openai:production:gpt-4o",
		ProviderEndpoint: config.ProviderEndpointConfig{
			Provider: "openai",
			Model:    "openai/gpt-4o",
			APIKey:   "test-key",
		},
	}}, config.RouterSettings{CooldownTime: 60, NumRetries: 1}, mem, nil, zap.NewNop())
	require.NoError(t, err)

	virtualModel := &platform.VirtualModel{
		ID:              "vm-hi",
		TenantID:        "tenant_a",
		Alias:           "hi-vm",
		RoutingStrategy: "priority",
		Enabled:         true,
		ExposedModes:    []string{"chat", "responses"},
		Targets: []platform.VirtualModelTarget{{
			ID:       "target-openai",
			Order:    1,
			Provider: "openai",
			Model:    "openai:production:gpt-4o",
			Policy: platform.VirtualModelPolicy{
				Retry:    platform.VirtualModelRetryPolicy{Attempts: 1, StatusCodes: []int{429}},
				Fallback: platform.VirtualModelFallbackPolicy{StatusCodes: []int{429}, Candidate: true},
			},
		}},
	}
	handler := NewResponsesHandlerWithDependencies(BaseHandlerDependencies{
		Router:            rt,
		Tunnel:            tunnel,
		Billing:           &testBudgeter{},
		RequestLogger:     logs,
		SpendLogger:       logs,
		Logger:            zap.NewNop(),
		VirtualModelStore: responsesVirtualModelStore{model: virtualModel},
	})

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(auth.AuthContextKey{}, &platform.AuthContext{
			TenantID:  "tenant_a",
			KeyID:     "key_1",
			KeyType:   "pat",
			Role:      "team_admin",
			TokenHash: "hash_1",
		})
		c.Locals(tenant.ContextKey{}, &platform.Tenant{ID: "tenant_a", Slug: "a", Status: "active"})

		return c.Next()
	})
	app.Post("/v1/responses", handler.Handle)
	app.Post("/v1/responses/compact", handler.HandleCompact)
	app.Post("/v1/responses/input_tokens", handler.HandleInputTokens)

	return app
}

func newRuntimeProfileVirtualResponsesTestApp(
	t *testing.T,
	tunnel *captureTunnel,
	billing budgeter,
	targets []platform.VirtualModelTarget,
) *fiber.App {
	t.Helper()

	mem, err := testcache.New(t, zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, mem.Close()) })

	rt, err := router.NewRouter([]config.ModelDeployment{
		{
			ModelName: "deepseek:deepseek-ember-harbor:deepseek-v4-flash",
			ProviderEndpoint: config.ProviderEndpointConfig{
				Provider: "deepseek",
				Model:    "deepseek/deepseek-v4-flash",
				APIKey:   "deepseek-key",
			},
			ModelInfo: map[string]interface{}{"id": "dep-deepseek"},
		},
		{
			ModelName: "openai:openai-calm-oasis:gpt-5.4-nano",
			ProviderEndpoint: config.ProviderEndpointConfig{
				Provider: "openai",
				Model:    "openai/gpt-5.4-nano",
				APIKey:   "openai-key",
			},
			ModelInfo: map[string]interface{}{"id": "dep-openai"},
		},
		{
			ModelName: "xai:xai-silent-river:grok-4.5",
			ProviderEndpoint: config.ProviderEndpointConfig{
				Provider: "xai",
				Model:    "xai/grok-4.5",
				APIKey:   "xai-key",
			},
			ModelInfo: map[string]interface{}{"id": "dep-xai"},
		},
	}, config.RouterSettings{CooldownTime: 60, NumRetries: 1}, mem, nil, zap.NewNop())
	require.NoError(t, err)

	virtualModel := &platform.VirtualModel{
		ID:              "vm-hi",
		TenantID:        "tenant_a",
		Alias:           "hi-vm",
		RoutingStrategy: "priority",
		Enabled:         true,
		ExposedModes:    []string{"chat", "responses"},
		Targets:         targets,
	}
	logs := &testLogs{}
	handler := NewResponsesHandlerWithDependencies(BaseHandlerDependencies{
		Router:            rt,
		Tunnel:            tunnel,
		Billing:           billing,
		RequestLogger:     logs,
		SpendLogger:       logs,
		Logger:            zap.NewNop(),
		VirtualModelStore: responsesVirtualModelStore{model: virtualModel},
	})

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(auth.AuthContextKey{}, &platform.AuthContext{
			TenantID:  "tenant_a",
			KeyID:     "key_1",
			KeyType:   "pat",
			Role:      "team_admin",
			TokenHash: "hash_1",
		})
		c.Locals(tenant.ContextKey{}, &platform.Tenant{ID: "tenant_a", Slug: "a", Status: "active"})

		return c.Next()
	})
	app.Post("/v1/responses", handler.Handle)

	return app
}

func newDirectDeepSeekResponsesTestApp(
	t *testing.T,
	tunnel *captureTunnel,
	billing budgeter,
) *fiber.App {
	t.Helper()

	mem, err := testcache.New(t, zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, mem.Close()) })

	rt, err := router.NewRouter([]config.ModelDeployment{{
		ModelName: "deepseek:deepseek-ember-harbor:deepseek-v4-flash",
		ProviderEndpoint: config.ProviderEndpointConfig{
			Provider: "deepseek",
			Model:    "deepseek/deepseek-v4-flash",
			APIKey:   "deepseek-key",
		},
		ModelInfo: map[string]interface{}{"id": "dep-deepseek"},
	}}, config.RouterSettings{CooldownTime: 60, NumRetries: 1}, mem, nil, zap.NewNop())
	require.NoError(t, err)

	handler := NewResponsesHandler(
		rt,
		tunnel,
		billing,
		&testLogs{},
		zap.NewNop(),
	)
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(auth.AuthContextKey{}, &platform.AuthContext{
			TenantID:  "tenant_a",
			KeyID:     "key_1",
			KeyType:   "pat",
			Role:      "team_admin",
			TokenHash: "hash_1",
		})
		c.Locals(tenant.ContextKey{}, &platform.Tenant{ID: "tenant_a", Slug: "a", Status: "active"})

		return c.Next()
	})
	app.Post("/v1/responses", handler.Handle)

	return app
}

func runtimeProfileVirtualTarget(id string, order int, provider, model string) platform.VirtualModelTarget {
	return platform.VirtualModelTarget{
		ID:       id,
		Order:    order,
		Provider: provider,
		Model:    model,
		Policy: platform.VirtualModelPolicy{
			Retry: platform.VirtualModelRetryPolicy{
				Attempts:    1,
				StatusCodes: []int{429, 500, 502, 503},
			},
			Fallback: platform.VirtualModelFallbackPolicy{
				Candidate:   true,
				StatusCodes: []int{429, 500, 502, 503, 504},
			},
		},
	}
}

func successfulResponsesTunnelResponse(t *testing.T, model string) *platform.TunnelResponse {
	t.Helper()

	body, err := json.Marshal(contractsopenai.Response{
		ID:        "resp_virtual",
		Object:    "response",
		CreatedAt: time.Now().Unix(),
		Model:     model,
		Status:    "completed",
		Usage: &contractsopenai.ResponseUsage{
			InputTokens:  2,
			OutputTokens: 1,
			TotalTokens:  3,
		},
	})
	require.NoError(t, err)

	return &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       body,
	}
}

func profiledResponsesCost(
	model string,
	costIRR float64,
	binding platform.ModelRuntimeProfileBinding,
) platform.UsageCostBreakdown {
	return platform.UsageCostBreakdown{
		Model:               model,
		InputTokens:         2,
		OutputTokens:        1,
		TotalTokens:         3,
		TotalCostIRR:        costIRR,
		PricingRevision:     binding.CatalogRevision,
		ModelRuntimeProfile: &binding,
	}
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
	logs := &testLogs{}
	app, q := newResponsesTestApp(t, "openai", tunnel, logs)

	req := httptest.NewRequest("POST", "/v1/responses", bytes.NewBufferString(`{"model":"gpt-4o","input":"hello"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

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
	require.Len(t, logs.loggedEntries, 1)
	assert.Equal(t, map[string]interface{}{"model": "gpt-4o", "input": "hello"}, logs.loggedEntries[0].RequestBody)
	assert.Equal(t, "resp_1", logs.loggedEntries[0].ResponseBody["id"])
	assert.Equal(t, "response", logs.loggedEntries[0].ResponseBody["object"])
	assert.NotContains(t, logs.loggedEntries[0].ResponseBody, "result_count")
}

func TestResponsesHandlerSimpleCacheHitSkipsProviderAndWallet(t *testing.T) {
	providerBody := []byte(`{
		"id":"resp_cache",
		"object":"response",
		"created_at":1700000000,
		"model":"gpt-4o",
		"status":"completed",
		"output":[],
		"usage":{"input_tokens":3,"output_tokens":2,"total_tokens":5},
		"future_response_field":{"preserved":true}
	}`)
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       providerBody,
	}}
	logs := &testLogs{}
	cacheService := newResponsesCacheService(t, nil)
	app, _ := newResponsesTestAppWithCache(t, "openai", tunnel, cacheService, logs)

	makeRequest := func() (int, http.Header, []byte) {
		req := httptest.NewRequest(
			http.MethodPost,
			"/v1/responses",
			bytes.NewBufferString(`{"model":"gpt-4o","input":"hello","future_request_field":{"mode":"official"}}`),
		)
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		req.Header.Set(apiheaders.HeaderCacheConfig, `{"type":"simple","ttl":60}`)

		resp, err := app.Test(req)
		require.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		statusCode := resp.StatusCode
		headers := resp.Header.Clone()
		closeResponseBody(t, resp)

		return statusCode, headers, body
	}

	firstStatus, firstHeaders, firstBody := makeRequest()
	assert.Equal(t, fiber.StatusOK, firstStatus)
	assert.Equal(t, "miss", firstHeaders.Get(apiheaders.HeaderCacheStatus))
	assert.JSONEq(t, string(providerBody), string(firstBody))

	secondStatus, secondHeaders, secondBody := makeRequest()
	assert.Equal(t, fiber.StatusOK, secondStatus)
	assert.Equal(t, "hit", secondHeaders.Get(apiheaders.HeaderCacheStatus))
	assert.NotEmpty(t, secondHeaders.Get(apiheaders.HeaderCachedTraceID))
	assert.JSONEq(t, string(providerBody), string(secondBody))
	assert.Len(t, tunnel.requests, 1)
	require.Len(t, logs.loggedSpends, 2)
	assert.Zero(t, logs.loggedSpends[1].CostUSD)
	require.Len(t, logs.loggedEntries, 2)
	assert.Equal(t, "hit", logs.loggedEntries[1].CacheStatus)
}

func TestResponsesHandlerSmartCacheReturnsSemanticHit(t *testing.T) {
	tunnel := &captureTunnel{resp: successfulResponsesTunnelResponse(t, "gpt-4o")}
	cacheService := newResponsesCacheService(t, &fixedResponsesCacheEmbedder{})
	app, _ := newResponsesTestAppWithCache(t, "openai", tunnel, cacheService)

	makeRequest := func(input string) http.Header {
		body, err := json.Marshal(map[string]interface{}{
			"model":        "gpt-4o",
			"input":        input,
			"instructions": "be helpful",
		})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(body))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		req.Header.Set(apiheaders.HeaderCacheConfig, `{"type":"smart","ttl":60,"similarity_threshold":0.9}`)

		resp, err := app.Test(req)
		require.NoError(t, err)

		headers := resp.Header.Clone()
		closeResponseBody(t, resp)

		return headers
	}

	firstHeaders := makeRequest("reset my password")
	assert.Equal(t, "miss", firstHeaders.Get(apiheaders.HeaderCacheStatus))

	secondHeaders := makeRequest("password recovery")
	assert.Equal(t, "hit", secondHeaders.Get(apiheaders.HeaderCacheStatus))
	assert.NotEmpty(t, secondHeaders.Get(apiheaders.HeaderCacheSimilarity))
	assert.Len(t, tunnel.requests, 1)
}

func TestResponsesHandlerRejectsInvalidCacheConfig(t *testing.T) {
	tunnel := &captureTunnel{resp: successfulResponsesTunnelResponse(t, "gpt-4o")}
	cacheService := newResponsesCacheService(t, nil)
	app, _ := newResponsesTestAppWithCache(t, "openai", tunnel, cacheService)
	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/responses",
		bytes.NewBufferString(`{"model":"gpt-4o","input":"hello"}`),
	)
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	req.Header.Set(apiheaders.HeaderCacheConfig, `{"type":"unknown"}`)

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	assert.Empty(t, tunnel.requests)
}

func TestResponsesHandlerBypassesCacheForBackgroundRequests(t *testing.T) {
	tunnel := &captureTunnel{resp: successfulResponsesTunnelResponse(t, "gpt-4o")}
	cacheService := newResponsesCacheService(t, nil)
	app, _ := newResponsesTestAppWithCache(t, "openai", tunnel, cacheService)

	makeRequest := func() http.Header {
		req := httptest.NewRequest(
			http.MethodPost,
			"/v1/responses",
			bytes.NewBufferString(`{"model":"gpt-4o","input":"hello","background":true}`),
		)
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		req.Header.Set(apiheaders.HeaderCacheConfig, `{"type":"simple","ttl":60}`)

		resp, err := app.Test(req)
		require.NoError(t, err)

		headers := resp.Header.Clone()
		closeResponseBody(t, resp)

		return headers
	}

	assert.Equal(t, "bypass", makeRequest().Get(apiheaders.HeaderCacheStatus))
	assert.Equal(t, "bypass", makeRequest().Get(apiheaders.HeaderCacheStatus))
	assert.Len(t, tunnel.requests, 2)
}

func TestResponsesHandlerVirtualModelBudgetDeferralDoesNotSelectTargets(t *testing.T) {
	resolver := &responsesVirtualModelResolver{model: &platform.VirtualModel{Alias: "hi-vm"}}
	engine := &trackingResponsesRoutingEngine{}
	handler := &ResponsesHandler{
		vmResolver: resolver,
		vmEngine:   engine,
	}

	deferred := handler.shouldDeferResponsesBudgetForVirtualModel(
		context.Background(),
		nil,
		"tenant_a",
		"hi-vm",
	)

	assert.True(t, deferred)
	assert.Equal(t, 1, resolver.calls)
	assert.Zero(t, engine.calls)
}

func TestResponsesHandlerVirtualModelSyncRoutesTargetWithoutAliasDeployment(t *testing.T) {
	providerResponse, err := json.Marshal(contractsopenai.Response{
		ID:        "resp_vm",
		Object:    "response",
		CreatedAt: time.Now().Unix(),
		Model:     "gpt-4o",
		Status:    "completed",
		Usage:     &contractsopenai.ResponseUsage{InputTokens: 1, OutputTokens: 1, TotalTokens: 2},
	})
	require.NoError(t, err)

	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       providerResponse,
	}}
	logs := &testLogs{}
	app := newVirtualResponsesTestApp(t, tunnel, logs)
	req := httptest.NewRequest("POST", "/v1/responses", bytes.NewBufferString(`{"model":"hi-vm","input":"hello"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	require.NoError(t, err)
	defer closeResponseBody(t, resp)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode, string(body))
	assert.NotEqual(t, "hi-vm", tunnel.lastReq.Deployment.ModelName)

	var outbound map[string]interface{}
	require.NoError(t, json.Unmarshal(tunnel.lastReq.Body, &outbound))
	assert.NotEqual(t, "hi-vm", outbound["model"])
	assert.Contains(t, outbound["model"], "gpt-4o")

	entries := logs.requestEntries()
	require.NotEmpty(t, entries)
	assert.Equal(t, "hi-vm", entries[len(entries)-1].VirtualModel)
}

func TestResponsesHandlerVirtualModelStreamRoutesTargetWithoutAliasDeployment(t *testing.T) {
	stream := newFakeTunnelStream(map[string]string{"content-type": "text/event-stream"}, 1, 0)
	stream.chunks <- platform.TunnelChunk{
		Data:   []byte("event: response.completed\ndata: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_vm_stream\",\"model\":\"gpt-4o\",\"status\":\"completed\",\"usage\":{\"input_tokens\":1,\"output_tokens\":1,\"total_tokens\":2}}}\n\n"),
		IsDone: true,
	}

	close(stream.chunks)

	tunnel := &captureTunnel{stream: stream}
	logs := &testLogs{}
	app := newVirtualResponsesTestApp(t, tunnel, logs)
	req := httptest.NewRequest("POST", "/v1/responses", bytes.NewBufferString(`{"model":"hi-vm","input":"hello","stream":true}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, int((5 * time.Second).Milliseconds()))

	require.NoError(t, err)
	defer closeResponseBody(t, resp)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode, string(body))
	assert.Contains(t, string(body), "response.completed")
	assert.NotEqual(t, "hi-vm", tunnel.lastReq.Deployment.ModelName)

	var outbound map[string]interface{}
	require.NoError(t, json.Unmarshal(tunnel.lastReq.Body, &outbound))
	assert.NotEqual(t, "hi-vm", outbound["model"])
	assert.Contains(t, outbound["model"], "gpt-4o")
	assert.True(t, stream.closed)
}

func TestResponsesHandlerVirtualModelUsesTargetSpecificDeepSeekRuntimeProfile(t *testing.T) {
	const (
		deepSeekModel = "deepseek:deepseek-ember-harbor:deepseek-v4-flash"
		openAIModel   = "openai:openai-calm-oasis:gpt-5.4-nano"
	)

	deepSeekProfile := deepSeekProxyRuntimeProfile("deepseek-v4-flash")
	billing := &targetAwareResponsesBudgeter{costs: map[string]platform.UsageCostBreakdown{
		"hi-vm":       {Model: "hi-vm", TotalTokens: 3, TotalCostIRR: 100},
		deepSeekModel: profiledResponsesCost(deepSeekModel, 200, deepSeekProfile),
		openAIModel:   {Model: openAIModel, TotalTokens: 3, TotalCostIRR: 150},
	}}
	tunnel := &captureTunnel{resp: successfulResponsesTunnelResponse(t, "deepseek-v4-flash")}
	app := newRuntimeProfileVirtualResponsesTestApp(t, tunnel, billing, []platform.VirtualModelTarget{
		runtimeProfileVirtualTarget("target-deepseek", 1, "deepseek", deepSeekModel),
		runtimeProfileVirtualTarget("target-openai", 2, "openai", openAIModel),
	})
	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/responses",
		bytes.NewBufferString(`{"model":"hi-vm","instructions":"Be concise.","input":"hello"}`),
	)
	request.Header.Set("Content-Type", "application/json")

	response, err := app.Test(request)

	require.NoError(t, err)
	defer closeResponseBody(t, response)

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, response.StatusCode, string(body))
	require.Len(t, tunnel.requests, 1)
	assert.Equal(t, "deepseek", tunnel.requests[0].Deployment.Provider)
	assert.Equal(t, "deepseek-v4-flash", tunnel.requests[0].Deployment.Model)
	require.Len(t, tunnel.requests[0].ModelRuntimeProfiles, 1)
	assert.Equal(t, "deepseek-v4-flash", tunnel.requests[0].ModelRuntimeProfiles[0].Slug)
	assert.Contains(t, billing.estimateModels, "hi-vm")
	assert.Contains(t, billing.estimateModels, deepSeekModel)
	require.NotNil(t, billing.budgetEstimate)
	assert.Equal(t, deepSeekModel, billing.budgetEstimate.Model)
	require.NotNil(t, billing.finalizedEstimate)
	assert.Equal(t, deepSeekModel, billing.finalizedEstimate.Model)

	var outbound contractsopenai.ResponseRequest
	require.NoError(t, json.Unmarshal(tunnel.requests[0].Body, &outbound))
	assert.Equal(t, "deepseek-v4-flash", outbound.Model)
}

func TestResponsesHandlerVirtualModelStreamsWithTargetSpecificDeepSeekRuntimeProfile(t *testing.T) {
	const deepSeekModel = "deepseek:deepseek-ember-harbor:deepseek-v4-flash"

	deepSeekProfile := deepSeekProxyRuntimeProfile("deepseek-v4-flash")
	billing := &targetAwareResponsesBudgeter{costs: map[string]platform.UsageCostBreakdown{
		"hi-vm":       {Model: "hi-vm", TotalTokens: 3, TotalCostIRR: 100},
		deepSeekModel: profiledResponsesCost(deepSeekModel, 200, deepSeekProfile),
	}}

	stream := newFakeTunnelStream(map[string]string{"content-type": "text/event-stream"}, 1, 0)
	stream.chunks <- platform.TunnelChunk{
		Data:   []byte("event: response.completed\ndata: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_ds_stream\",\"object\":\"response\",\"model\":\"deepseek-v4-flash\",\"status\":\"completed\",\"usage\":{\"input_tokens\":2,\"output_tokens\":1,\"total_tokens\":3}}}\n\n"),
		IsDone: true,
	}

	close(stream.chunks)

	tunnel := &captureTunnel{stream: stream}
	app := newRuntimeProfileVirtualResponsesTestApp(t, tunnel, billing, []platform.VirtualModelTarget{
		runtimeProfileVirtualTarget("target-deepseek", 1, "deepseek", deepSeekModel),
	})
	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/responses",
		bytes.NewBufferString(`{"model":"hi-vm","input":"hello","stream":true}`),
	)
	request.Header.Set("Content-Type", "application/json")

	response, err := app.Test(request, int((5 * time.Second).Milliseconds()))

	require.NoError(t, err)
	defer closeResponseBody(t, response)

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, response.StatusCode, string(body))
	assert.Contains(t, string(body), "response.completed")
	require.Len(t, tunnel.requests, 1)
	require.Len(t, tunnel.requests[0].ModelRuntimeProfiles, 1)
	assert.Equal(t, "deepseek-v4-flash", tunnel.requests[0].ModelRuntimeProfiles[0].Slug)
	require.NotNil(t, billing.finalizedEstimate)
	assert.Equal(t, deepSeekModel, billing.finalizedEstimate.Model)
	assert.True(t, stream.closed)
}

func TestResponsesHandlerVirtualModelFallsBackWhenDeepSeekPreparationFails(t *testing.T) {
	const (
		deepSeekModel = "deepseek:deepseek-ember-harbor:deepseek-v4-flash"
		openAIModel   = "openai:openai-calm-oasis:gpt-5.4-nano"
	)

	billing := &targetAwareResponsesBudgeter{costs: map[string]platform.UsageCostBreakdown{
		"hi-vm":       {Model: "hi-vm", TotalTokens: 3, TotalCostIRR: 100},
		deepSeekModel: {Model: deepSeekModel, TotalTokens: 3, TotalCostIRR: 200},
		openAIModel:   {Model: openAIModel, TotalTokens: 3, TotalCostIRR: 150},
	}}
	tunnel := &captureTunnel{resp: successfulResponsesTunnelResponse(t, "gpt-5.4-nano")}
	app := newRuntimeProfileVirtualResponsesTestApp(t, tunnel, billing, []platform.VirtualModelTarget{
		runtimeProfileVirtualTarget("target-deepseek", 1, "deepseek", deepSeekModel),
		runtimeProfileVirtualTarget("target-openai", 2, "openai", openAIModel),
	})
	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/responses",
		bytes.NewBufferString(`{"model":"hi-vm","input":"hello"}`),
	)
	request.Header.Set("Content-Type", "application/json")

	response, err := app.Test(request)

	require.NoError(t, err)
	defer closeResponseBody(t, response)

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, response.StatusCode, string(body))
	require.Len(t, tunnel.requests, 1)
	assert.Equal(t, "openai", tunnel.requests[0].Deployment.Provider)
	assert.Contains(t, billing.estimateModels, deepSeekModel)
	assert.Contains(t, billing.estimateModels, openAIModel)
	require.NotNil(t, billing.budgetEstimate)
	assert.Equal(t, openAIModel, billing.budgetEstimate.Model)
	require.NotNil(t, billing.finalizedEstimate)
	assert.Equal(t, openAIModel, billing.finalizedEstimate.Model)
}

func TestResponsesHandlerVirtualModelFallsBackFromDeepSeekProviderFailure(t *testing.T) {
	const (
		deepSeekModel = "deepseek:deepseek-ember-harbor:deepseek-v4-flash"
		openAIModel   = "openai:openai-calm-oasis:gpt-5.4-nano"
	)

	deepSeekProfile := deepSeekProxyRuntimeProfile("deepseek-v4-flash")
	billing := &targetAwareResponsesBudgeter{costs: map[string]platform.UsageCostBreakdown{
		"hi-vm":       {Model: "hi-vm", TotalTokens: 3, TotalCostIRR: 100},
		deepSeekModel: profiledResponsesCost(deepSeekModel, 200, deepSeekProfile),
		openAIModel:   {Model: openAIModel, TotalTokens: 3, TotalCostIRR: 150},
	}}
	tunnel := &captureTunnel{responses: []*platform.TunnelResponse{
		{StatusCode: fiber.StatusInternalServerError, Body: []byte(`{"error":{"message":"deepseek unavailable"}}`)},
		successfulResponsesTunnelResponse(t, "gpt-5.4-nano"),
	}}
	app := newRuntimeProfileVirtualResponsesTestApp(t, tunnel, billing, []platform.VirtualModelTarget{
		runtimeProfileVirtualTarget("target-deepseek", 1, "deepseek", deepSeekModel),
		runtimeProfileVirtualTarget("target-openai", 2, "openai", openAIModel),
	})
	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/responses",
		bytes.NewBufferString(`{"model":"hi-vm","input":"hello"}`),
	)
	request.Header.Set("Content-Type", "application/json")

	response, err := app.Test(request)

	require.NoError(t, err)
	defer closeResponseBody(t, response)

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, response.StatusCode, string(body))
	require.Len(t, tunnel.requests, 2)
	assert.Equal(t, "deepseek", tunnel.requests[0].Deployment.Provider)
	assert.Equal(t, "openai", tunnel.requests[1].Deployment.Provider)
	require.NotNil(t, billing.budgetEstimate)
	assert.Equal(t, deepSeekModel, billing.budgetEstimate.Model)
	require.NotNil(t, billing.finalizedEstimate)
	assert.Equal(t, openAIModel, billing.finalizedEstimate.Model)
}

func TestResponsesHandlerVirtualModelUsesTargetSpecificXAIRuntimeProfile(t *testing.T) {
	const xaiModel = "xai:xai-silent-river:grok-4.5"

	xaiProfile := xaiProxyTestRuntimeProfile()
	billing := &targetAwareResponsesBudgeter{costs: map[string]platform.UsageCostBreakdown{
		"hi-vm":  {Model: "hi-vm", TotalTokens: 3, TotalCostIRR: 100},
		xaiModel: profiledResponsesCost(xaiModel, 250, xaiProfile),
	}}
	tunnel := &captureTunnel{resp: successfulResponsesTunnelResponse(t, "grok-4.5")}
	app := newRuntimeProfileVirtualResponsesTestApp(t, tunnel, billing, []platform.VirtualModelTarget{
		runtimeProfileVirtualTarget("target-xai", 1, "xai", xaiModel),
	})
	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/responses",
		bytes.NewBufferString(`{"model":"hi-vm","input":"hello"}`),
	)
	request.Header.Set("Content-Type", "application/json")

	response, err := app.Test(request)

	require.NoError(t, err)
	defer closeResponseBody(t, response)

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, response.StatusCode, string(body))
	require.Len(t, tunnel.requests, 1)
	assert.Equal(t, "xai", tunnel.requests[0].Deployment.Provider)
	require.Len(t, tunnel.requests[0].ModelRuntimeProfiles, 1)
	assert.Equal(t, "grok-4.5", tunnel.requests[0].ModelRuntimeProfiles[0].Slug)
	require.NotNil(t, billing.finalizedEstimate)
	assert.Equal(t, xaiModel, billing.finalizedEstimate.Model)
}

func TestResponsesHandlerDirectDeepSeekStillFailsClosedWithoutRuntimeProfile(t *testing.T) {
	const deepSeekModel = "deepseek:deepseek-ember-harbor:deepseek-v4-flash"

	billing := &targetAwareResponsesBudgeter{costs: map[string]platform.UsageCostBreakdown{
		deepSeekModel: {Model: deepSeekModel, TotalTokens: 3, TotalCostIRR: 200},
	}}
	tunnel := &captureTunnel{resp: successfulResponsesTunnelResponse(t, "deepseek-v4-flash")}
	app := newDirectDeepSeekResponsesTestApp(t, tunnel, billing)
	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/responses",
		bytes.NewBufferString(`{"model":"deepseek:deepseek-ember-harbor:deepseek-v4-flash","input":"hello"}`),
	)
	request.Header.Set("Content-Type", "application/json")

	response, err := app.Test(request)

	require.NoError(t, err)
	defer closeResponseBody(t, response)

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusServiceUnavailable, response.StatusCode, string(body))
	assert.Contains(t, string(body), "runtime_profile_unavailable")
	assert.Empty(t, tunnel.requests)
}

func TestResponsesHandlerVirtualModelCompactRoutesTargetWithoutAliasDeployment(t *testing.T) {
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       []byte(`{"id":"resp_compact_vm","object":"response.compaction"}`),
	}}
	logs := &testLogs{}
	app := newVirtualResponsesTestApp(t, tunnel, logs)
	req := httptest.NewRequest(
		"POST",
		"/v1/responses/compact",
		bytes.NewBufferString(`{"model":"hi-vm","input":"hello"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	require.NoError(t, err)
	defer closeResponseBody(t, resp)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode, string(body))
	assert.NotEqual(t, "hi-vm", tunnel.lastReq.Deployment.ModelName)

	var outbound map[string]interface{}
	require.NoError(t, json.Unmarshal(tunnel.lastReq.Body, &outbound))
	assert.NotEqual(t, "hi-vm", outbound["model"])
	assert.Contains(t, outbound["model"], "gpt-4o")
}

func TestResponseUsagePreservesCacheWriteTokensForBilling(t *testing.T) {
	usage := responseUsageToOpenAI(&contractsopenai.ResponseUsage{
		InputTokens:  100,
		OutputTokens: 10,
		TotalTokens:  110,
		InputDetails: &contractsopenai.ResponseInputTokenDetails{
			CachedTokens:     20,
			CacheWriteTokens: 30,
		},
	})

	require.NotNil(t, usage)
	require.NotNil(t, usage.PromptTokensDetails)
	assert.Equal(t, 20, usage.PromptTokensDetails.CachedTokens)
	assert.Equal(t, 30, usage.PromptTokensDetails.CacheWriteTokens)
	assert.Equal(t, 20, usage.PromptCacheHitTokens)
	assert.Equal(t, 80, usage.PromptCacheMissTokens)
}

func TestResponsesHandlerForwardsTextFormat(t *testing.T) {
	payload := contractsopenai.Response{
		ID:        "resp_structured",
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
				Text: `{"name":"science fair"}`,
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
	app, _ := newResponsesTestApp(t, "openai", tunnel)

	req := httptest.NewRequest("POST", "/v1/responses", bytes.NewBufferString(`{
		"model":"gpt-4o",
		"input":"Extract the event.",
		"text":{
			"format":{
				"type":"json_schema",
				"name":"event",
				"strict":true,
				"schema":{"type":"object","properties":{"name":{"type":"string"}},"required":["name"],"additionalProperties":false}
			}
		}
	}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var outbound contractsopenai.ResponseRequest
	require.NoError(t, json.Unmarshal(tunnel.lastReq.Body, &outbound))
	require.NotNil(t, outbound.Text)
	raw, err := json.Marshal(outbound.Text)
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"format"`)
	assert.Contains(t, string(raw), `"type":"json_schema"`)
	assert.Contains(t, string(raw), `"strict":true`)
	assert.Contains(t, string(raw), `"additionalProperties":false`)
}

func TestResponsesHandlerPreservesUnknownOfficialFieldsLosslessly(t *testing.T) {
	upstreamBody := []byte(`{"id":"resp_future","object":"response","model":"gpt-4o","status":"completed","usage":{"input_tokens":1,"output_tokens":1,"total_tokens":2},"future_response_field":{"enabled":true}}`)
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       upstreamBody,
	}}
	app, _ := newResponsesTestApp(t, "openai", tunnel)

	req := httptest.NewRequest("POST", "/v1/responses", bytes.NewBufferString(`{"model":"gpt-4o","input":"hello","future_request_field":{"mode":"official"}}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	var outbound map[string]interface{}
	require.NoError(t, json.Unmarshal(tunnel.lastReq.Body, &outbound))
	assert.Equal(t, map[string]interface{}{"mode": "official"}, outbound["future_request_field"])

	received, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.JSONEq(t, string(upstreamBody), string(received))
	assert.Contains(t, string(received), `"future_response_field"`)
}

// GPT-6 request features documented at
// https://developers.openai.com/api/docs/guides/latest-model must reach OpenAI
// byte-for-byte: async tools, configuration_update items, reasoning mode and
// context, and prompt_cache_options with explicit cache breakpoints.
func TestResponsesHandlerRelaysGPT6RequestFeaturesLosslessly(t *testing.T) {
	requestBody := `{
		"model":"gpt-6-sol",
		"previous_response_id":null,
		"reasoning":{"effort":"medium","mode":"pro","context":"all_turns"},
		"prompt_cache_options":{"ttl":"30m","mode":"explicit"},
		"service_tier":"fast",
		"tools":[
			{"type":"function","name":"lookup_price","async":true,"parameters":{"type":"object","properties":{"task_handle":{"type":"string"}},"required":["task_handle"]}},
			{"type":"custom","name":"run_job","async":true},
			{"type":"function","name":"wait_for_tasks","async":false,"parameters":{"type":"object","properties":{}}}
		],
		"input":[
			{"role":"developer","content":[{"type":"input_text","text":"Stable instructions.","prompt_cache_breakpoint":{"mode":"explicit"}}]},
			{"type":"function_call_output","call_id":"call_widget","output":"{\"price_cents\":1200}"},
			{"type":"configuration_update","reasoning":{"effort":"high"}},
			{"role":"user","content":"Compare the prices."}
		]
	}`
	upstreamBody := []byte(`{
		"id":"resp_gpt6","object":"response","model":"gpt-6-sol","status":"completed","service_tier":"priority",
		"output":[{"type":"function_call","id":"fc_1","call_id":"call_gadget","name":"lookup_price",
			"arguments":"{\"task_handle\":\"gadget_price_1\"}","async":true,"status":"completed"}],
		"usage":{"input_tokens":10,"output_tokens":2,"total_tokens":12}
	}`)
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       upstreamBody,
	}}
	app, _ := newResponsesTestApp(t, "openai", tunnel)

	req := httptest.NewRequest("POST", "/v1/responses", bytes.NewBufferString(requestBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	var inbound, outbound map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(requestBody), &inbound))
	require.NoError(t, json.Unmarshal(tunnel.lastReq.Body, &outbound))
	assert.Equal(t, inbound, outbound)

	received, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.JSONEq(t, string(upstreamBody), string(received))
}

func TestResponsesHandlerPreservesOpenAIErrorContract(t *testing.T) {
	upstreamBody := []byte(`{"error":{"message":"bad tool","type":"invalid_request_error","param":"tools[0].type","code":"invalid_tool","future_error_field":{"retryable":false}}}`)
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusBadRequest,
		Headers: map[string]string{
			"x-request-id":                   "openai_req_error",
			"x-ratelimit-remaining-requests": "42",
		},
		Body: upstreamBody,
	}}
	app, _ := newResponsesTestApp(t, "openai", tunnel)

	req := httptest.NewRequest("POST", "/v1/responses", bytes.NewBufferString(`{"model":"gpt-4o","input":"hello"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, "openai_req_error", resp.Header.Get("X-Request-Id"))
	assert.Equal(t, "42", resp.Header.Get("X-Ratelimit-Remaining-Requests"))

	received, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.JSONEq(t, string(upstreamBody), string(received))
}

func TestResponsesHandlerLogsVirtualModelProviderValidationErrorExactlyOnce(t *testing.T) {
	upstreamBody := []byte(`{"error":{"message":"Unsupported parameter: 'messages'. In the Responses API, this parameter has moved to 'input'.","type":"invalid_request_error","param":null,"code":"unsupported_parameter"}}`)
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusBadRequest,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       upstreamBody,
	}}
	logs := &testLogs{}
	app := newVirtualResponsesTestApp(t, tunnel, logs)

	requestBody := `{"model":"hi-vm","messages":[{"role":"user","content":"در مورد علوم computer یک جمله بگو"}]}`
	request := httptest.NewRequest("POST", "/v1/responses", bytes.NewBufferString(requestBody))
	request.Header.Set("Content-Type", "application/json")

	response, err := app.Test(request)
	require.NoError(t, err)

	defer closeResponseBody(t, response)

	require.Equal(t, fiber.StatusBadRequest, response.StatusCode)

	responseBody, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	assert.JSONEq(t, string(upstreamBody), string(responseBody))

	entries := logs.requestEntries()
	require.Len(t, entries, 1)

	entry := entries[0]
	assert.Equal(t, "failed", entry.Status)
	assert.Equal(t, fiber.StatusBadRequest, entry.StatusCode)
	assert.Equal(t, "invalid_request_error", entry.ErrorType)
	assert.Equal(t, "hi-vm", entry.Model)
	assert.Equal(t, "hi-vm", entry.VirtualModel)
	assert.Contains(t, entry.RequestBody, "messages")

	errorPayload, ok := entry.ResponseBody["error"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "unsupported_parameter", errorPayload["code"])
	assert.Contains(t, errorPayload, "param")
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
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusNotImplemented, resp.StatusCode)
}

func TestDeepSeekResponsesCapabilityIsCreateOnly(t *testing.T) {
	assert.True(t, supportsResponsesCreateProvider("deepseek"))
	assert.False(t, supportsResponseResourcesProvider("deepseek"))
	assert.True(t, supportsResponseResourcesProvider("openai"))
	assert.True(t, supportsResponseResourcesProvider("xai"))
}

func TestResponsesHandlerSupportsXAIWithPinnedProfileAndProviderFields(t *testing.T) {
	upstreamBody := []byte(
		`{"id":"resp_xai","object":"response","model":"grok-4.5","status":"completed",` +
			`"output":[{"type":"x_search_call","id":"tool_1","provider_extension":{"keep":true}}],` +
			`"usage":{"input_tokens":3,"output_tokens":2,"total_tokens":5,` +
			`"cost_in_usd_ticks":123456789,"num_server_side_tools_used":1},` +
			`"xai_extension":{"keep":true}}`,
	)
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       upstreamBody,
	}}
	app, _ := newResponsesTestApp(t, "xai", tunnel)

	req := httptest.NewRequest("POST", "/v1/responses", bytes.NewBufferString(`{"model":"grok-4.5","input":"hello","reasoning":{"effort":"low"}}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	require.Len(t, tunnel.lastReq.ModelRuntimeProfiles, 1)
	assert.Equal(t, "xai", tunnel.lastReq.ModelRuntimeProfiles[0].Provider)
	assert.True(t, tunnel.lastReq.ModelRuntimeProfiles[0].MatchesModel("grok-4.5"))
	assert.Equal(t, platform.ProviderOperationOpenAIResponsesCreate, tunnel.lastReq.Operation)
	assert.Contains(t, string(tunnel.lastReq.Body), `"reasoning":{"effort":"low"}`)

	received, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.JSONEq(t, string(upstreamBody), string(received))
	assert.Contains(t, string(received), `"provider_extension":{"keep":true}`)
	assert.Contains(t, string(received), `"cost_in_usd_ticks":123456789`)
	assert.Contains(t, string(received), `"num_server_side_tools_used":1`)
}

func TestResponsesHandlerSupportsCurrentDeepSeekModelsAndRedactsLogs(t *testing.T) {
	for _, model := range []string{"deepseek-v4-flash", "deepseek-v4-pro"} {
		t.Run(model, func(t *testing.T) {
			upstreamBody := []byte(
				`{"id":"resp_ds","object":"response","model":"` + model + `","status":"completed",` +
					`"output":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"private answer"}]}],` +
					`"usage":{"input_tokens":10,"input_tokens_details":{"cached_tokens":4},"output_tokens":3,` +
					`"output_tokens_details":{"reasoning_tokens":2},"total_tokens":13},"store":false,"previous_response_id":null}`,
			)
			tunnel := &captureTunnel{resp: &platform.TunnelResponse{
				StatusCode: fiber.StatusOK,
				Headers:    map[string]string{"content-type": "application/json"},
				Body:       upstreamBody,
			}}
			logs := &testLogs{}
			app, queue := newResponsesTestApp(t, "deepseek", tunnel, logs)

			request := httptest.NewRequest("POST", "/v1/responses", bytes.NewBufferString(`{"model":"`+model+`","input":"private prompt","reasoning":{"effort":"high"}}`))
			request.Header.Set("Content-Type", "application/json")

			response, err := app.Test(request)
			defer closeResponseBody(t, response)

			require.NoError(t, err)
			require.Equal(t, fiber.StatusOK, response.StatusCode)
			require.Len(t, tunnel.lastReq.ModelRuntimeProfiles, 1)
			assert.Equal(t, platform.ProviderOperationOpenAIResponsesCreate, tunnel.lastReq.Operation)
			assert.Equal(t, model, tunnel.lastReq.ModelRuntimeProfiles[0].Slug)
			assert.Contains(t, string(tunnel.lastReq.Body), `"input":"private prompt"`)

			require.Len(t, logs.loggedEntries, 1)
			requestLog, err := json.Marshal(logs.loggedEntries[0].RequestBody)
			require.NoError(t, err)
			responseLog, err := json.Marshal(logs.loggedEntries[0].ResponseBody)
			require.NoError(t, err)
			assert.NotContains(t, string(requestLog), "private prompt")
			assert.NotContains(t, string(responseLog), "private answer")
			assert.Contains(t, string(requestLog), `"redacted":true`)

			for _, event := range queue.DequeueBatch(20) {
				encoded, marshalErr := json.Marshal(event.Attributes)
				require.NoError(t, marshalErr)
				assert.NotContains(t, string(encoded), "private prompt")
				assert.NotContains(t, string(encoded), "private answer")
			}
		})
	}
}

func TestResponsesHandlerCanonicalizesProviderPrefixedDeploymentModel(t *testing.T) {
	mem, err := testcache.New(t, zap.NewNop())
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
	h := NewResponsesHandler(rt, tunnel, &testBudgeter{}, &testLogs{}, zap.NewNop())

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
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var outbound contractsopenai.ResponseRequest

	err = json.Unmarshal(tunnel.lastReq.Body, &outbound)
	require.NoError(t, err)
	assert.Equal(t, "gpt-4o-mini", outbound.Model)
}

func TestResponsesHandlerStreamWithoutTerminalEventEmitsOrderedError(t *testing.T) {
	stream := newFakeTunnelStream(map[string]string{"content-type": "text/event-stream"}, 2, 0)
	stream.chunks <- platform.TunnelChunk{Data: []byte("event: response.created\n")}

	stream.chunks <- platform.TunnelChunk{Data: []byte("data: {\"type\":\"response.output_text.delta\",\"delta\":\"Salam\",\"sequence_number\":7}\n\n"), IsDone: true}

	close(stream.chunks)

	tunnel := &captureTunnel{stream: stream}
	logs := &testLogs{}
	cacheService := newResponsesCacheService(t, nil)
	app, _ := newResponsesTestAppWithCache(t, "openai", tunnel, cacheService, logs)

	req := httptest.NewRequest("POST", "/v1/responses", bytes.NewBufferString(`{"model":"gpt-4o","input":"hello","stream":true}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(apiheaders.HeaderCacheConfig, `{"type":"simple","ttl":60}`)

	resp, err := app.Test(req, int((5 * time.Second).Milliseconds()))
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))
	assert.Equal(t, "bypass", resp.Header.Get(apiheaders.HeaderCacheStatus))
	assert.Equal(t, "/responses", tunnel.lastReq.RequestMetadata["provider_endpoint_path"])

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	bodyText := string(body)
	assert.Contains(t, bodyText, "event: response.created")
	assert.Contains(t, bodyText, `response.output_text.delta`)
	assert.Contains(t, bodyText, "event: error\n")
	assert.Contains(t, bodyText, `"sequence_number":8`)
	assert.True(t, stream.closed)

	entries := logs.requestEntries()
	require.Len(t, entries, 1)
	assert.Equal(t, "failed", entries[0].Status)
	assert.Equal(t, fiber.StatusOK, entries[0].StatusCode)
}

func TestResponsesHandlerStreamSuccessRequiresTerminalEvent(t *testing.T) {
	stream := newFakeTunnelStream(map[string]string{"content-type": "text/event-stream"}, 2, 0)
	stream.chunks <- platform.TunnelChunk{Data: []byte("event: response.created\ndata: {\"type\":\"response.created\",\"sequence_number\":0,\"response\":{\"id\":\"resp_success\",\"object\":\"response\",\"model\":\"gpt-4o\",\"status\":\"in_progress\"}}\n\n")}

	completedEvent := "event: response.completed\n" +
		"data: {\"type\":\"response.completed\",\"sequence_number\":1," +
		"\"response\":{\"id\":\"resp_success\",\"object\":\"response\",\"model\":\"gpt-4o\"," +
		"\"status\":\"completed\",\"usage\":{\"input_tokens\":1,\"output_tokens\":1,\"total_tokens\":2}}}\n\n"
	stream.chunks <- platform.TunnelChunk{Data: []byte(completedEvent), IsDone: true}

	close(stream.chunks)

	tunnel := &captureTunnel{stream: stream}
	logs := &testLogs{}
	app, _ := newResponsesTestApp(t, "openai", tunnel, logs)

	req := httptest.NewRequest("POST", "/v1/responses", bytes.NewBufferString(`{"model":"gpt-4o","input":"hello","stream":true}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, int((5 * time.Second).Milliseconds()))
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), `response.completed`)
	assert.NotContains(t, string(body), "event: error\n")

	entries := logs.requestEntries()
	require.Len(t, entries, 1)
	assert.Equal(t, "success", entries[0].Status)
}

func TestResponsesHandlerStreamErrorWritesStreamingErrorEvent(t *testing.T) {
	stream := newFakeTunnelStream(map[string]string{"content-type": "text/event-stream"}, 0, 1)
	stream.errors <- errors.New("responses stream failed")

	tunnel := &captureTunnel{stream: stream}
	logs := &testLogs{}
	app, _ := newResponsesTestApp(t, "openai", tunnel, logs)

	req := httptest.NewRequest("POST", "/v1/responses", bytes.NewBufferString(`{"model":"gpt-4o","input":"hello","stream":true}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, int((5 * time.Second).Milliseconds()))
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "event: error\n")
	assert.Contains(t, string(body), `"type":"error"`)
	assert.Contains(t, string(body), `"code":null`)
	assert.Contains(t, string(body), `"param":null`)
	assert.Contains(t, string(body), `"sequence_number":0`)
	assert.True(t, stream.closed)

	entries := logs.requestEntries()
	require.Len(t, entries, 1)
	assert.Equal(t, "failed", entries[0].Status)
	assert.Equal(t, fiber.StatusOK, entries[0].StatusCode)
	assert.Equal(t, "streaming_error", entries[0].ErrorType)
}

func TestDeepSeekResponsesStreamRequiresTerminalEvent(t *testing.T) {
	stream := newFakeTunnelStream(map[string]string{"content-type": "text/event-stream"}, 1, 0)
	stream.chunks <- platform.TunnelChunk{
		Data:   []byte("event: response.output_text.delta\ndata: {\"type\":\"response.output_text.delta\",\"delta\":\"partial\"}\n\n"),
		IsDone: true,
	}

	logs := &testLogs{}
	tunnel := &captureTunnel{stream: stream}
	app, _ := newResponsesTestApp(t, "deepseek", tunnel, logs)

	request := httptest.NewRequest(
		"POST",
		"/v1/responses",
		bytes.NewBufferString(`{"model":"deepseek-v4-flash","input":"hello","stream":true}`),
	)
	request.Header.Set("Content-Type", "application/json")

	response, err := app.Test(request, int((5 * time.Second).Milliseconds()))
	defer closeResponseBody(t, response)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, response.StatusCode)

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "event: error\n")
	assert.Contains(t, string(body), "without a terminal Responses event")
	assert.Empty(t, logs.loggedSpends)
	assert.True(t, stream.closed)

	entries := logs.requestEntries()
	require.Len(t, entries, 1)
	assert.Equal(t, "failed", entries[0].Status)
	assert.Equal(t, "stream_incomplete", entries[0].ErrorType)
}

func TestResponsesHandlerStreamCloseFinalizesWithoutPanic(t *testing.T) {
	stream := newFakeTunnelStream(map[string]string{"content-type": "text/event-stream"}, 1, 0)
	stream.chunks <- platform.TunnelChunk{Data: []byte(
		"event: response.completed\n" +
			"data: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_done\",\"model\":\"gpt-4o\"," +
			"\"status\":\"completed\",\"service_tier\":\"standard\",\"usage\":{\"input_tokens\":7," +
			"\"output_tokens\":3,\"total_tokens\":10,\"input_tokens_details\":{\"cached_tokens\":2," +
			"\"cache_write_tokens\":1}}}}\n\n",
	)}

	close(stream.chunks)

	logs := &testLogs{}

	mem, err := testcache.New(t, zap.NewNop())
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

	h := NewResponsesHandler(rt, &captureTunnel{stream: stream}, &testBudgeter{}, logs, zap.NewNop())
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
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), `response.completed`)
	assert.NotContains(t, strings.ToLower(string(body)), "panic")
	assert.Len(t, logs.loggedSpends, 1)
	assert.Len(t, logs.loggedEntries, 1)
	assert.Equal(t, "success", logs.loggedEntries[0].Status)
	assert.Equal(t, 7, logs.loggedSpends[0].InputTokens)
	assert.Equal(t, 3, logs.loggedSpends[0].OutputTokens)
	assert.Equal(t, 10, logs.loggedSpends[0].TotalTokens)
	assert.True(t, stream.closed)
	assert.Equal(t, "/responses", logs.loggedSpends[0].RequestMetadata["provider_endpoint_path"])
}
