package proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Delpak/shahrah/services/control-plane/internal/testcache"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/config"
	metricsobs "github.com/Delpak/shahrah/services/control-plane/internal/observability/metrics"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newModerationsTestApp(t *testing.T, modelName string, provider string, deploymentModel string, tunnel *captureTunnel, logs *testLogs, allowedTypes []string, emitMetric ...func(input metricsobs.BuildInput)) (*fiber.App, *tracing.Queue) {
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

	if allowedTypes == nil {
		allowedTypes = []string{"moderation"}
	}

	h := NewModerationsHandler(rt, tunnel, &testBudgeter{}, logs, zap.NewNop(), emitMetric...)
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
				modelName: allowedTypes,
			},
		})
		c.Locals(tenant.ContextKey{}, &platform.Tenant{
			ID:     "tenant_a",
			Slug:   "a",
			Status: "active",
		})

		return c.Next()
	})
	app.Post("/v1/moderations", h.Handle)

	return app, q
}

func TestModerationsHandlerSingleStringInput(t *testing.T) {
	payload := contractsopenai.ModerationResponse{
		ID:    "modr_1",
		Model: "omni-moderation-latest",
		Results: []contractsopenai.ModerationResult{{
			Flagged:        false,
			Categories:     map[string]bool{"hate": false, "violence": false},
			CategoryScores: map[string]float64{"hate": 0.001, "violence": 0.002},
		}},
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       body,
	}}
	logs := &testLogs{}
	app, _ := newModerationsTestApp(t, "openai:acct:omni-moderation-latest", "openai", "omni-moderation-latest", tunnel, logs, nil)

	req := httptest.NewRequest("POST", "/v1/moderations", bytes.NewBufferString(`{"model":"openai:acct:omni-moderation-latest","input":"hello"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, "/moderations", tunnel.lastReq.RequestMetadata["provider_endpoint_path"])

	var outbound contractsopenai.ModerationRequest

	err = json.Unmarshal(tunnel.lastReq.Body, &outbound)
	require.NoError(t, err)
	assert.Equal(t, "omni-moderation-latest", outbound.Model)
	assert.Equal(t, "hello", outbound.Input)

	require.Len(t, logs.loggedEntries, 1)
	assert.Equal(t, map[string]interface{}{"model": "openai:acct:omni-moderation-latest", "input": "hello"}, logs.loggedEntries[0].RequestBody)
	assert.Equal(t, "modr_1", logs.loggedEntries[0].ResponseBody["id"])
	assert.Equal(t, "omni-moderation-latest", logs.loggedEntries[0].ResponseBody["model"])
	assert.NotContains(t, logs.loggedEntries[0].ResponseBody, "result_count")
	assert.NotContains(t, logs.loggedEntries[0].ResponseBody, "flagged_count")

	require.Len(t, logs.loggedSpends, 1)
	assert.Equal(t, "/moderations", logs.loggedSpends[0].RequestMetadata["provider_endpoint_path"])
	assert.Equal(t, 1, logs.loggedSpends[0].RequestMetadata["input_count"])
}

func TestModerationsHandlerPreservesFutureRequestAndResponseFields(t *testing.T) {
	providerBody := []byte(`{"id":"modr_future","model":"omni-moderation-latest","results":[],"future_response":{"taxonomy":"v2"}}`)
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json; charset=utf-8"},
		Body:       providerBody,
	}}
	app, _ := newModerationsTestApp(t, "omni-moderation-latest", "openai", "omni-moderation-latest", tunnel, nil, nil)

	req := httptest.NewRequest(
		"POST",
		"/v1/moderations",
		bytes.NewBufferString(`{"model":"omni-moderation-latest","input":"hello","future_request":{"policy":"strict"}}`),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	require.NoError(t, err)
	defer closeResponseBody(t, resp)

	var outbound map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(tunnel.lastReq.Body, &outbound))
	require.JSONEq(t, `{"policy":"strict"}`, string(outbound["future_request"]))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, string(providerBody), string(body))
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
}

func TestNormalizeModerationInputAcceptsTextAndImageItems(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{"type": "text", "text": "check this"},
		map[string]interface{}{"type": "image_url", "image_url": map[string]interface{}{"url": "https://example.test/image.png"}},
	}

	normalized, err := normalizeModerationInput(input)
	require.NoError(t, err)
	assert.Equal(t, input, normalized)
}

func TestModerationsHandlerBatchInputPreservesOrder(t *testing.T) {
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       []byte(`{"id":"modr_1","model":"omni-moderation-latest","results":[{"flagged":false,"categories":{"hate":false},"category_scores":{"hate":0.001}}]}`),
	}}
	app, _ := newModerationsTestApp(t, "omni-moderation-latest", "openai", "omni-moderation-latest", tunnel, nil, nil)

	req := httptest.NewRequest("POST", "/v1/moderations", bytes.NewBufferString(`{"model":"omni-moderation-latest","input":["first","second"]}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var outbound contractsopenai.ModerationRequest

	err = json.Unmarshal(tunnel.lastReq.Body, &outbound)
	require.NoError(t, err)
	assert.Equal(t, []interface{}{"first", "second"}, outbound.Input)
}

func TestModerationsHandlerValidationErrors(t *testing.T) {
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{StatusCode: fiber.StatusOK, Headers: map[string]string{"content-type": "application/json"}, Body: []byte(`{}`)}}
	app, _ := newModerationsTestApp(t, "omni-moderation-latest", "openai", "omni-moderation-latest", tunnel, nil, nil)

	cases := []struct {
		name  string
		body  string
		param string
		code  string
	}{
		{name: "missing model", body: `{"input":"hello"}`, param: "model", code: "invalid_model"},
		{name: "missing input", body: `{"model":"omni-moderation-latest"}`, param: "input", code: "invalid_input"},
		{name: "empty string", body: `{"model":"omni-moderation-latest","input":""}`, param: "input", code: "invalid_input"},
		{name: "empty array", body: `{"model":"omni-moderation-latest","input":[]}`, param: "input", code: "invalid_input"},
		{name: "object unsupported", body: `{"model":"omni-moderation-latest","input":{"type":"video","video":"hello"}}`, param: "input", code: "unsupported_input_type"},
		{name: "array object unsupported", body: `{"model":"omni-moderation-latest","input":[{"type":"video","video":"hello"}]}`, param: "input", code: "unsupported_input_type"},
		{name: "stream unsupported", body: `{"model":"omni-moderation-latest","input":"hello","stream":true}`, param: "stream", code: "unsupported_parameter"},
		{name: "messages unsupported", body: `{"model":"omni-moderation-latest","input":"hello","messages":[]}`, param: "messages", code: "unsupported_parameter"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/v1/moderations", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			defer closeResponseBody(t, resp)

			require.NoError(t, err)
			assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

			var apiErr contractsopenai.APIError

			err = json.NewDecoder(resp.Body).Decode(&apiErr)
			require.NoError(t, err)
			assert.Equal(t, tc.param, apiErr.Error.Param)
			assert.Equal(t, tc.code, apiErr.Error.Code)
		})
	}
}

func TestModerationsHandlerRejectsNonModerationModel(t *testing.T) {
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       []byte(`{}`),
	}}
	app, _ := newModerationsTestApp(t, "gpt-4o-mini", "openai", "gpt-4o-mini", tunnel, nil, []string{"chat"})

	req := httptest.NewRequest("POST", "/v1/moderations", bytes.NewBufferString(`{"model":"gpt-4o-mini","input":"hello"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestModerationsHandlerProviderErrorNormalized(t *testing.T) {
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusBadRequest,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       []byte(`{"error":{"message":"upstream rejected input","type":"invalid_request_error"}}`),
	}}
	app, _ := newModerationsTestApp(t, "omni-moderation-latest", "openai", "omni-moderation-latest", tunnel, nil, nil)

	req := httptest.NewRequest("POST", "/v1/moderations", bytes.NewBufferString(`{"model":"omni-moderation-latest","input":"hello"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var apiErr contractsopenai.APIError

	err = json.NewDecoder(resp.Body).Decode(&apiErr)
	require.NoError(t, err)
	assert.Equal(t, "upstream rejected input", apiErr.Error.Message)
}

func TestModerationsHandlerRejectsUnsupportedProvider(t *testing.T) {
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       []byte(`{}`),
	}}
	app, _ := newModerationsTestApp(t, "omni-moderation-latest", "anthropic", "omni-moderation-latest", tunnel, nil, nil)

	req := httptest.NewRequest("POST", "/v1/moderations", bytes.NewBufferString(`{"model":"omni-moderation-latest","input":"hello"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusNotImplemented, resp.StatusCode)
}

func TestModerationTracePayloadDoesNotIncludeRawInput(t *testing.T) {
	payload := buildTraceModerationRequestPayload(contractsopenai.ModerationRequest{
		Model: "omni-moderation-latest",
		Input: "sensitive text to moderate",
		User:  "user-1",
	})
	raw, err := json.Marshal(payload)
	require.NoError(t, err)
	assert.NotContains(t, string(raw), "sensitive text to moderate")
	assert.Contains(t, string(raw), "input_hashes")
	assert.Contains(t, string(raw), "input_lengths")
}

func TestModerationsHandlerEmitsMetricMetadata(t *testing.T) {
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       []byte(`{"id":"modr_1","model":"omni-moderation-latest","results":[{"flagged":true,"categories":{"violence":true},"category_scores":{"violence":0.99}}]}`),
	}}

	var metric metricsobs.BuildInput

	app, _ := newModerationsTestApp(t, "omni-moderation-latest", "openai", "omni-moderation-latest", tunnel, nil, nil, func(input metricsobs.BuildInput) {
		metric = input
	})

	req := httptest.NewRequest("POST", "/v1/moderations", bytes.NewBufferString(`{"model":"omni-moderation-latest","input":"hello"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, "moderations", metric.RequestType)
	assert.Equal(t, "/v1/moderations", metric.Metadata["endpoint"])
	assert.Equal(t, 1, metric.Metadata["input_count"])
}

func TestModerationsHandlerTraceLifecycle(t *testing.T) {
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       []byte(`{"id":"modr_1","model":"omni-moderation-latest","results":[{"flagged":false,"categories":{"hate":false},"category_scores":{"hate":0.001}}]}`),
	}}
	app, q := newModerationsTestApp(t, "omni-moderation-latest", "openai", "omni-moderation-latest", tunnel, nil, nil)

	req := httptest.NewRequest("POST", "/v1/moderations", strings.NewReader(`{"model":"omni-moderation-latest","input":"hello"}`))
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
