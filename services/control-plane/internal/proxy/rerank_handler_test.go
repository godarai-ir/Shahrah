package proxy

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/Delpak/shahrah/services/control-plane/internal/testcache"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/config"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newRerankTestApp(t *testing.T, modelName string, provider string, deploymentModel string, tunnel *captureTunnel, logs *testLogs) (*fiber.App, *tracing.Queue) {
	t.Helper()

	mem, err := testcache.New(t, zap.NewNop())
	require.NoError(t, err)
	rt, err := router.NewRouter([]config.ModelDeployment{{
		ModelName: modelName,
		ProviderEndpoint: config.ProviderEndpointConfig{
			Provider: provider,
			Model:    deploymentModel,
			APIKey:   "test-key",
		},
	}}, config.RouterSettings{CooldownTime: 60, NumRetries: 1}, mem, nil, zap.NewNop())
	require.NoError(t, err)

	if logs == nil {
		logs = &testLogs{}
	}

	h := NewRerankHandler(rt, tunnel, &testBudgeter{}, logs, zap.NewNop())
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
			AllowedModelTypes: map[string][]string{
				modelName: {"rerank"},
			},
		})
		c.Locals(tenant.ContextKey{}, &platform.Tenant{
			ID:     "tenant_a",
			Slug:   "a",
			Status: "active",
		})

		return c.Next()
	})
	app.Post("/v1/rerank", h.Handle)

	return app, q
}

func TestRerankHandlerSuccessRoutesToCohereV2Rerank(t *testing.T) {
	payload := contractsopenai.RerankResponse{
		ID: "rerank_123",
		Results: []contractsopenai.RerankResult{
			{Index: 0, RelevanceScore: 0.98},
			{Index: 2, RelevanceScore: 0.41},
		},
		Usage: &contractsopenai.Usage{PromptTokens: 20, TotalTokens: 20},
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       body,
	}}
	logs := &testLogs{}
	app, q := newRerankTestApp(t, "cohere:acct:rerank-multilingual-v3.0", "cohere", "rerank-multilingual-v3.0", tunnel, logs)

	req := httptest.NewRequest("POST", "/v1/rerank", bytes.NewBufferString(`{
		"model":"cohere:acct:rerank-multilingual-v3.0",
		"query":"تهران پایتخت کدام کشور است؟",
		"documents":["تهران پایتخت ایران است.","آسمان امروز آبی است.","مشهد یکی از شهرهای بزرگ ایران است."],
		"top_n":2
	}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, "/v2/rerank", tunnel.lastReq.RequestMetadata["provider_endpoint_path"])
	assert.Equal(t, 3, tunnel.lastReq.RequestMetadata["document_count"])

	var outbound contractsopenai.RerankRequest
	require.NoError(t, json.Unmarshal(tunnel.lastReq.Body, &outbound))
	assert.Equal(t, "rerank-multilingual-v3.0", outbound.Model)
	assert.Equal(t, 3, len(outbound.Documents))
	require.NotNil(t, outbound.TopN)
	assert.Equal(t, 2, *outbound.TopN)

	var decoded contractsopenai.RerankResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&decoded))
	require.Len(t, decoded.Results, 2)
	assert.Equal(t, 0, decoded.Results[0].Index)

	require.Len(t, logs.loggedSpends, 1)
	assert.Equal(t, "/v2/rerank", logs.loggedSpends[0].RequestMetadata["provider_endpoint_path"])
	require.Len(t, logs.loggedEntries, 1)
	assert.Equal(t, "cohere:acct:rerank-multilingual-v3.0", logs.loggedEntries[0].RequestBody["model"])
	assert.Equal(t, "تهران پایتخت کدام کشور است؟", logs.loggedEntries[0].RequestBody["query"])
	requestDocuments, ok := logs.loggedEntries[0].RequestBody["documents"].([]string)
	require.True(t, ok)
	require.Len(t, requestDocuments, 3)
	assert.Equal(t, "rerank_123", logs.loggedEntries[0].ResponseBody["id"])
	responseResults, ok := logs.loggedEntries[0].ResponseBody["results"].([]interface{})
	require.True(t, ok)
	require.Len(t, responseResults, 2)
	assert.NotContains(t, logs.loggedEntries[0].ResponseBody, "result_count")

	events := q.DequeueBatch(20)

	var sawProvider bool

	for _, ev := range events {
		if ev.Name == "provider.call" {
			sawProvider = true
		}
	}

	assert.True(t, sawProvider)
}

func TestRerankHandlerValidationErrors(t *testing.T) {
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{StatusCode: fiber.StatusOK, Headers: map[string]string{"content-type": "application/json"}, Body: []byte(`{"results":[]}`)}}
	app, _ := newRerankTestApp(t, "rerank-multilingual-v3.0", "cohere", "rerank-multilingual-v3.0", tunnel, nil)

	cases := []struct {
		name  string
		body  string
		param string
	}{
		{name: "missing model", body: `{"query":"q","documents":["a"]}`, param: "model"},
		{name: "missing query", body: `{"model":"rerank-multilingual-v3.0","documents":["a"]}`, param: "query"},
		{name: "empty docs", body: `{"model":"rerank-multilingual-v3.0","query":"q","documents":[]}`, param: "documents"},
		{name: "bad top_n", body: `{"model":"rerank-multilingual-v3.0","query":"q","documents":["a"],"top_n":0}`, param: "top_n"},
		{name: "unsupported stream", body: `{"model":"rerank-multilingual-v3.0","query":"q","documents":["a"],"stream":true}`, param: "stream"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/v1/rerank", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			defer closeResponseBody(t, resp)

			require.NoError(t, err)
			assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

			var apiErr contractsopenai.APIError
			require.NoError(t, json.NewDecoder(resp.Body).Decode(&apiErr))
			assert.Equal(t, tc.param, apiErr.Error.Param)
		})
	}
}

func TestRerankHandlerRejectsUnsupportedProvider(t *testing.T) {
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{StatusCode: fiber.StatusOK, Headers: map[string]string{"content-type": "application/json"}, Body: []byte(`{"results":[]}`)}}
	app, _ := newRerankTestApp(t, "rerank-multilingual-v3.0", "openai", "rerank-multilingual-v3.0", tunnel, nil)

	req := httptest.NewRequest("POST", "/v1/rerank", bytes.NewBufferString(`{"model":"rerank-multilingual-v3.0","query":"q","documents":["a"]}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusNotImplemented, resp.StatusCode)
}
