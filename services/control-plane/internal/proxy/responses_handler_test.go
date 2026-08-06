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

type captureTunnel struct {
	resp      *platform.TunnelResponse
	stream    platform.TunnelStream
	streamErr error
	lastReq   platform.TunnelRequest
}

type xaiResponsesBudgeter struct{ testBudgeter }

type deepSeekResponsesBudgeter struct{ testBudgeter }

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

	binding := deepSeekProxyRuntimeProfile("deepseek-v4-flash")
	estimated.ModelRuntimeProfile = &binding
	estimated.PricingRevision = binding.CatalogRevision

	return estimated, nil
}

func (t *captureTunnel) Proxy(_ context.Context, req platform.TunnelRequest) (*platform.TunnelResponse, error) {
	t.lastReq = req
	return t.resp, nil
}

func (t *captureTunnel) Stream(_ context.Context, req platform.TunnelRequest) (platform.TunnelStream, error) {
	t.lastReq = req
	if t.stream != nil || t.streamErr != nil {
		return t.stream, t.streamErr
	}

	return nil, assert.AnError
}

func (t *captureTunnel) Healthy(_ context.Context) error { return nil }

func newResponsesTestApp(t *testing.T, provider string, tunnel *captureTunnel, logsOpt ...*testLogs) (*fiber.App, *tracing.Queue) {
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

	rt, err := router.NewRouter([]config.ModelDeployment{{
		ModelName: modelName,
		ProviderEndpoint: config.ProviderEndpointConfig{
			Provider: provider,
			Model:    modelName,
			APIKey:   "test-key",
		},
	}}, config.RouterSettings{CooldownTime: 60, NumRetries: 1}, mem, nil, zap.NewNop())
	require.NoError(t, err)

	logs := &testLogs{}
	if len(logsOpt) > 0 && logsOpt[0] != nil {
		logs = logsOpt[0]
	}

	h := NewResponsesHandler(rt, tunnel, billing, &testWallets{}, logs, zap.NewNop())
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

func TestResponsesHandlerSupportsDeepSeekFlashCreateAndRedactsLogs(t *testing.T) {
	upstreamBody := []byte(
		`{"id":"resp_ds","object":"response","model":"deepseek-v4-flash","status":"completed",` +
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

	request := httptest.NewRequest("POST", "/v1/responses", bytes.NewBufferString(`{"model":"deepseek-v4-flash","input":"private prompt","reasoning":{"effort":"high"}}`))
	request.Header.Set("Content-Type", "application/json")

	response, err := app.Test(request)
	defer closeResponseBody(t, response)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, response.StatusCode)
	require.Len(t, tunnel.lastReq.ModelRuntimeProfiles, 1)
	assert.Equal(t, platform.ProviderOperationOpenAIResponsesCreate, tunnel.lastReq.Operation)
	assert.Equal(t, "deepseek-v4-flash", tunnel.lastReq.ModelRuntimeProfiles[0].Slug)
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
}

func TestResponsesHandlerRejectsDeepSeekProBeforeDispatch(t *testing.T) {
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       []byte(`{}`),
	}}
	app, _ := newResponsesTestApp(t, "deepseek", tunnel)

	request := httptest.NewRequest("POST", "/v1/responses", bytes.NewBufferString(`{"model":"deepseek-v4-pro","input":"hello"}`))
	request.Header.Set("Content-Type", "application/json")

	response, err := app.Test(request)
	defer closeResponseBody(t, response)

	require.NoError(t, err)
	assert.NotEqual(t, fiber.StatusOK, response.StatusCode)
	assert.Empty(t, tunnel.lastReq.Operation)
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
	defer closeResponseBody(t, resp)

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
	defer closeResponseBody(t, resp)

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

	wallets := &testWallets{}
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
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), `response.completed`)
	assert.NotContains(t, strings.ToLower(string(body)), "panic")
	assert.Len(t, wallets.debits, 1)
	assert.Len(t, logs.loggedSpends, 1)
	assert.Equal(t, 7, logs.loggedSpends[0].InputTokens)
	assert.Equal(t, 3, logs.loggedSpends[0].OutputTokens)
	assert.Equal(t, 10, logs.loggedSpends[0].TotalTokens)
	assert.True(t, stream.closed)
	assert.Equal(t, "/responses", logs.loggedSpends[0].RequestMetadata["provider_endpoint_path"])
}
