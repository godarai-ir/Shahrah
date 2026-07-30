package proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

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

func newEmbeddingsTestApp(t *testing.T, modelName string, provider string, deploymentModel string, tunnel *captureTunnel, logs *testLogs) (*fiber.App, *tracing.Queue) {
	t.Helper()

	mem, err := cache.NewMemoryCache(zap.NewNop())
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

	h := NewEmbeddingsHandler(rt, tunnel, &testBudgeter{}, &testWallets{}, logs, zap.NewNop())
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
				modelName: {"embedding"},
			},
		})
		c.Locals(tenant.ContextKey{}, &platform.Tenant{
			ID:     "tenant_a",
			Slug:   "a",
			Status: "active",
		})

		return c.Next()
	})
	app.Post("/v1/embeddings", h.Handle)

	return app, q
}

func TestEmbeddingsHandlerSingleStringInput(t *testing.T) {
	payload := contractsopenai.EmbeddingResponse{
		Object: "list",
		Model:  "text-embedding-3-small",
		Data: []contractsopenai.Embedding{{
			Object:    "embedding",
			Index:     0,
			Embedding: json.RawMessage(`[0.1,0.2,0.3]`),
		}},
		Usage: &contractsopenai.Usage{PromptTokens: 8, TotalTokens: 8},
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       body,
	}}
	logs := &testLogs{}
	app, q := newEmbeddingsTestApp(t, "openai:acct:text-embedding-3-small", "openai", "text-embedding-3-small", tunnel, logs)

	req := httptest.NewRequest("POST", "/v1/embeddings", bytes.NewBufferString(`{"model":"openai:acct:text-embedding-3-small","input":"hello","dimensions":1536}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, "/embeddings", tunnel.lastReq.RequestMetadata["provider_endpoint_path"])

	var outbound contractsopenai.EmbeddingRequest

	err = json.Unmarshal(tunnel.lastReq.Body, &outbound)
	require.NoError(t, err)
	assert.Equal(t, "text-embedding-3-small", outbound.Model)
	assert.Equal(t, "hello", outbound.Input)

	require.Len(t, logs.loggedEntries, 1)
	assert.Equal(t, map[string]interface{}{
		"model":      "openai:acct:text-embedding-3-small",
		"input":      "hello",
		"dimensions": float64(1536),
	}, logs.loggedEntries[0].RequestBody)
	assert.NotContains(t, logs.loggedEntries[0].RequestBody, "encoding_format")
	assert.Equal(t, "list", logs.loggedEntries[0].ResponseBody["object"])
	assert.Equal(t, "text-embedding-3-small", logs.loggedEntries[0].ResponseBody["model"])
	assert.NotContains(t, logs.loggedEntries[0].ResponseBody, "embedding_count")
	responseData, ok := logs.loggedEntries[0].ResponseBody["data"].([]interface{})
	require.True(t, ok)
	require.Len(t, responseData, 1)
	responseItem, ok := responseData[0].(map[string]interface{})
	require.True(t, ok)
	embedding, ok := responseItem["embedding"].([]interface{})
	require.True(t, ok)
	assert.Len(t, embedding, 3)

	usage, ok := logs.loggedEntries[0].ResponseBody["usage"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(8), usage["prompt_tokens"])
	assert.Equal(t, float64(8), usage["total_tokens"])

	require.Len(t, logs.loggedSpends, 1)
	assert.Equal(t, "/embeddings", logs.loggedSpends[0].RequestMetadata["provider_endpoint_path"])

	events := q.DequeueBatch(20)

	var sawProvider bool

	for _, ev := range events {
		if ev.Name == "provider.call" {
			sawProvider = true
		}
	}

	assert.True(t, sawProvider)
}

func TestEmbeddingsHandlerPreservesFutureRequestAndResponseFields(t *testing.T) {
	providerBody := []byte(`{"object":"list","model":"text-embedding-3-small","data":[],"usage":{"prompt_tokens":1,"total_tokens":1},"future_response":{"revision":2}}`)
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json; charset=utf-8"},
		Body:       providerBody,
	}}
	app, _ := newEmbeddingsTestApp(t, "text-embedding-3-small", "openai", "text-embedding-3-small", tunnel, nil)

	req := httptest.NewRequest(
		"POST",
		"/v1/embeddings",
		bytes.NewBufferString(`{"model":"text-embedding-3-small","input":"hello","future_request":{"mode":"exact"}}`),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer closeResponseBody(t, resp)

	var outbound map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(tunnel.lastReq.Body, &outbound))
	require.JSONEq(t, `{"mode":"exact"}`, string(outbound["future_request"]))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, string(providerBody), string(body))
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
}

func TestNormalizeEmbeddingInputAcceptsTokenArrays(t *testing.T) {
	for _, input := range []interface{}{
		[]interface{}{float64(1), float64(2), float64(3)},
		[]interface{}{
			[]interface{}{float64(1), float64(2)},
			[]interface{}{float64(3)},
		},
	} {
		normalized, err := normalizeEmbeddingInput(input)
		require.NoError(t, err)
		assert.Equal(t, input, normalized)
	}
}

func TestNormalizeEmbeddingInputRejectsMixedTokenAndTextArrays(t *testing.T) {
	_, err := normalizeEmbeddingInput([]interface{}{"hello", float64(2)})
	require.Error(t, err)
}

func TestEmbeddingsHandlerBatchInputPreservesOrder(t *testing.T) {
	payload := contractsopenai.EmbeddingResponse{
		Object: "list",
		Model:  "text-embedding-3-small",
		Data: []contractsopenai.Embedding{
			{Object: "embedding", Index: 0, Embedding: json.RawMessage(`[0.1,0.2]`)},
			{Object: "embedding", Index: 1, Embedding: json.RawMessage(`[0.3,0.4]`)},
		},
		Usage: &contractsopenai.Usage{PromptTokens: 12, TotalTokens: 12},
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       body,
	}}
	app, _ := newEmbeddingsTestApp(t, "text-embedding-3-small", "openai", "text-embedding-3-small", tunnel, nil)

	req := httptest.NewRequest("POST", "/v1/embeddings", bytes.NewBufferString(`{"model":"text-embedding-3-small","input":["first","second"]}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var decoded contractsopenai.EmbeddingResponse

	err = json.NewDecoder(resp.Body).Decode(&decoded)
	require.NoError(t, err)
	require.Len(t, decoded.Data, 2)
	assert.Equal(t, 0, decoded.Data[0].Index)
	assert.Equal(t, 1, decoded.Data[1].Index)
}

func TestEmbeddingsHandlerValidationErrors(t *testing.T) {
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{StatusCode: fiber.StatusOK, Headers: map[string]string{"content-type": "application/json"}, Body: []byte(`{}`)}}
	app, _ := newEmbeddingsTestApp(t, "text-embedding-3-small", "openai", "text-embedding-3-small", tunnel, nil)

	cases := []struct {
		name   string
		body   string
		status int
		param  string
	}{
		{name: "missing model", body: `{"input":"hello"}`, status: fiber.StatusBadRequest, param: "model"},
		{name: "missing input", body: `{"model":"text-embedding-3-small"}`, status: fiber.StatusBadRequest, param: "input"},
		{name: "empty string", body: `{"model":"text-embedding-3-small","input":""}`, status: fiber.StatusBadRequest, param: "input"},
		{name: "empty array", body: `{"model":"text-embedding-3-small","input":[]}`, status: fiber.StatusBadRequest, param: "input"},
		{name: "bad encoding", body: `{"model":"text-embedding-3-small","input":"hello","encoding_format":"hex"}`, status: fiber.StatusBadRequest, param: "encoding_format"},
		{name: "bad dimensions", body: `{"model":"text-embedding-3-small","input":"hello","dimensions":0}`, status: fiber.StatusBadRequest, param: "dimensions"},
		{name: "mixed token array unsupported", body: `{"model":"text-embedding-3-small","input":[1,"two",3]}`, status: fiber.StatusBadRequest, param: "input"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/v1/embeddings", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			defer closeResponseBody(t, resp)

			require.NoError(t, err)
			assert.Equal(t, tc.status, resp.StatusCode)

			var apiErr contractsopenai.APIError

			err = json.NewDecoder(resp.Body).Decode(&apiErr)
			require.NoError(t, err)
			assert.Equal(t, tc.param, apiErr.Error.Param)
		})
	}
}

func TestEmbeddingsHandlerProviderErrorNormalized(t *testing.T) {
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusBadRequest,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       []byte(`{"error":{"message":"upstream rejected dimensions","type":"invalid_request_error"}}`),
	}}
	app, _ := newEmbeddingsTestApp(t, "text-embedding-3-small", "openai", "text-embedding-3-small", tunnel, nil)

	req := httptest.NewRequest("POST", "/v1/embeddings", bytes.NewBufferString(`{"model":"text-embedding-3-small","input":"hello"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var apiErr contractsopenai.APIError

	err = json.NewDecoder(resp.Body).Decode(&apiErr)
	require.NoError(t, err)
	assert.Equal(t, "upstream rejected dimensions", apiErr.Error.Message)
}

func TestEmbeddingsHandlerRejectsUnsupportedProvider(t *testing.T) {
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       []byte(`{}`),
	}}
	app, _ := newEmbeddingsTestApp(t, "text-embedding-3-small", "anthropic", "text-embedding-3-small", tunnel, nil)

	req := httptest.NewRequest("POST", "/v1/embeddings", bytes.NewBufferString(`{"model":"text-embedding-3-small","input":"hello"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusNotImplemented, resp.StatusCode)
}

func TestEmbeddingsHandlerTraceLifecycle(t *testing.T) {
	payload := contractsopenai.EmbeddingResponse{
		Object: "list",
		Model:  "text-embedding-3-small",
		Data: []contractsopenai.Embedding{{
			Object:    "embedding",
			Index:     0,
			Embedding: json.RawMessage(`[0.1,0.2]`),
		}},
		Usage: &contractsopenai.Usage{PromptTokens: 5, TotalTokens: 5},
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       body,
	}}
	app, q := newEmbeddingsTestApp(t, "text-embedding-3-small", "openai", "text-embedding-3-small", tunnel, nil)

	req := httptest.NewRequest("POST", "/v1/embeddings", bytes.NewBufferString(`{"model":"text-embedding-3-small","input":"hello","user":"user-1"}`))
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
}

func TestEmbeddingsHandlerUsageRecordIncludesVectorDimensions(t *testing.T) {
	payload := contractsopenai.EmbeddingResponse{
		Object: "list",
		Model:  "text-embedding-3-small",
		Data: []contractsopenai.Embedding{{
			Object:    "embedding",
			Index:     0,
			Embedding: json.RawMessage(`[0.1,0.2,0.3,0.4]`),
		}},
		Usage: &contractsopenai.Usage{PromptTokens: 7, TotalTokens: 7},
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       body,
	}}
	logs := &testLogs{}
	app, _ := newEmbeddingsTestApp(t, "text-embedding-3-small", "openai", "text-embedding-3-small", tunnel, logs)

	req := httptest.NewRequest("POST", "/v1/embeddings", bytes.NewBufferString(`{"model":"text-embedding-3-small","input":["one","two"]}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	require.Len(t, logs.loggedSpends, 1)
	assert.Equal(t, []int{4}, logs.loggedSpends[0].RequestMetadata["vector_dimensions"])
	assert.Equal(t, 7, logs.loggedSpends[0].InputTokens)
}
