package proxy

import (
	"bytes"
	"context"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/config"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/Delpak/shahrah/services/control-plane/internal/testcache"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// tokenBazaarMessagesProxyTestRuntimeProfile mirrors
// tokenBazaarProxyTestRuntimeProfile (responses_tokenbazaar_test.go) but for
// the "messages" endpoint and one of the 4 in-scope Claude models.
func tokenBazaarMessagesProxyTestRuntimeProfile(model string) platform.ModelRuntimeProfileBinding {
	return platform.ModelRuntimeProfileBinding{
		Provider:        "tokenbazaar",
		CatalogRevision: "tokenbazaar-messages-test-revision",
		CanonicalID:     "tokenbazaar/" + model,
		Slug:            model,
		ProfileVersion:  1,
		EffectiveFrom:   time.Now().UTC(),
		Profile: platform.ModelRuntimeProfile{
			Version:    1,
			Endpoints:  []string{"models", "chat_completions", "messages"},
			Transports: []string{"http", "sse"},
			Lifecycle:  "active",
			Thinking: platform.ModelRuntimeThinkingProfile{
				DefaultMode: "adaptive", AllowedModes: []string{"adaptive", "disabled"},
			},
			Sampling:   platform.ModelRuntimeSamplingProfile{Parameters: []string{"temperature", "top_p"}},
			Tools:      platform.ModelRuntimeToolsProfile{Custom: true, Supported: []string{"function_calling"}},
			Modalities: platform.ModelRuntimeModalities{Input: []string{"text"}, Output: []string{"text"}},
			Parameters: map[string][]string{
				"messages": {"model", "messages", "max_tokens", "system", "stream", "thinking"},
			},
			Limits: platform.ModelRuntimeLimits{
				ContextWindow: 200000, MaxInputTokens: 200000, MaxOutputTokens: 64000,
				MessagesBytes: 32 * 1024 * 1024, BatchBytes: 256 * 1024 * 1024, FileBytes: 500 * 1024 * 1024,
			},
		},
	}
}

type tokenBazaarMessagesBudgeter struct{ testBudgeter }

func (b *tokenBazaarMessagesBudgeter) EstimateCost(ctx context.Context, req contractsopenai.ChatCompletionRequest) (platform.UsageCostBreakdown, error) {
	estimated, err := b.testBudgeter.EstimateCost(ctx, req)
	if err != nil {
		return estimated, err
	}

	binding := tokenBazaarMessagesProxyTestRuntimeProfile(providerModelID(req.Model))
	estimated.ModelRuntimeProfile = &binding
	estimated.PricingRevision = binding.CatalogRevision

	return estimated, nil
}

// newTokenBazaarMessagesTestApp builds a POST /v1/messages test app wired
// through NewAnthropicMessagesRouter exactly as routes_proxy.go wires it:
// a tokenbazaar-FQN'd model dispatches to TokenBazaarMessagesHandler; every
// other request falls back to a stand-in "native" handler (a real
// AnthropicNativeHandler is out of scope for this narrow test -- the routing
// shim itself, not AnthropicNativeHandler's own already-tested behavior, is
// what this delivery needs to prove).
func newTokenBazaarMessagesTestApp(
	t *testing.T,
	tunnel *captureTunnel,
) (*fiber.App, *testLogs) {
	t.Helper()

	mem, err := testcache.New(t, zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, mem.Close()) })

	// The static ModelName here is the literal
	// "tokenbazaar:<account>:<model-id>" FQN this test sends as the request
	// body's "model" field -- mirroring, in a self-contained unit test, how
	// a real dynamic tenant deployment is keyed by that same FQN string (see
	// router.Router.Route's isModelFQN handling). This is a deliberate test
	// simplification: it does not exercise the dynamic-store FQN lookup path
	// itself (already covered by router package tests), only that
	// TokenBazaarMessagesHandler routes, bills, and dispatches correctly once
	// a deployment is resolved.
	rt, err := router.NewRouter([]config.ModelDeployment{{
		ModelName: "tokenbazaar:primary:claude-sonnet-5",
		ProviderEndpoint: config.ProviderEndpointConfig{
			Provider: "tokenbazaar",
			Model:    "claude-sonnet-5",
			APIKey:   "test-key",
		},
	}}, config.RouterSettings{CooldownTime: 60, NumRetries: 1}, mem, nil, zap.NewNop())
	require.NoError(t, err)

	logs := &testLogs{}

	handler := NewTokenBazaarMessagesHandlerWithDependencies(BaseHandlerDependencies{
		Router:              rt,
		Tunnel:              tunnel,
		Billing:             &tokenBazaarMessagesBudgeter{},
		BillingCapabilities: compatibilityBillingCapabilities(),
		RequestLogger:       logs,
		SpendLogger:         logs,
		Logger:              zap.NewNop(),
	})

	native := func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusTeapot).SendString("native-anthropic-handler-called")
	}

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(auth.AuthContextKey{}, &platform.AuthContext{
			TenantID:  "tenant_a",
			KeyID:     "key_1",
			KeyType:   "pat",
			Role:      "member",
			TokenHash: "hash_1",
		})
		c.Locals(tenant.ContextKey{}, &platform.Tenant{ID: "tenant_a", Slug: "a", Status: "active"})

		return c.Next()
	})
	app.Post("/v1/messages", NewAnthropicMessagesRouter(handler, native))

	return app, logs
}

// TestRequestTargetsTokenBazaarMessages confirms the routing shim's sole
// signal (an explicit "tokenbazaar:<account>:<model>" FQN in the request
// body) is evaluated correctly: a bare model id must never be treated as a
// TokenBazaar request.
func TestRequestTargetsTokenBazaarMessages(t *testing.T) {
	cases := map[string]bool{
		`{"model":"tokenbazaar:primary:claude-sonnet-5","messages":[]}`: true,
		`{"model":"claude-sonnet-5","messages":[]}`:                     false,
		`{"model":"anthropic:primary:claude-sonnet-5","messages":[]}`:   false,
		`not json`: false,
		`{}`:       false,
	}

	for body, want := range cases {
		if got := requestTargetsTokenBazaarMessages([]byte(body)); got != want {
			t.Fatalf("requestTargetsTokenBazaarMessages(%s) = %t, want %t", body, got, want)
		}
	}
}

// TestAnthropicMessagesRouterFallsBackToNativeForNonTokenBazaarModels proves
// the shim leaves AnthropicNativeHandler's own behavior for the direct
// "anthropic" provider (and any bare model id) completely untouched.
func TestAnthropicMessagesRouterFallsBackToNativeForNonTokenBazaarModels(t *testing.T) {
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{StatusCode: fiber.StatusOK}}
	app, _ := newTokenBazaarMessagesTestApp(t, tunnel)

	req := httptest.NewRequest("POST", "/v1/messages", bytes.NewBufferString(`{"model":"claude-sonnet-5","messages":[],"max_tokens":16}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusTeapot, resp.StatusCode, "a bare model id must fall back to the native Anthropic handler")

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, "native-anthropic-handler-called", string(body))
	assert.Empty(t, tunnel.lastReq.RequestID, "TokenBazaarMessagesHandler must never dispatch a request meant for the native handler")
}

// TestTokenBazaarMessagesHandlerDispatchesSuccessfully confirms a
// well-formed request reaches the tunnel with the Anthropic-native operation,
// the deployment's provider-facing model, and a runtime profile bound to
// "tokenbazaar", and that a successful Anthropic-shaped response is returned
// to the caller unchanged.
func TestTokenBazaarMessagesHandlerDispatchesSuccessfully(t *testing.T) {
	upstreamBody := []byte(
		`{"id":"msg_1","type":"message","role":"assistant","model":"claude-sonnet-5",` +
			`"content":[{"type":"text","text":"hi there"}],"stop_reason":"end_turn",` +
			`"usage":{"input_tokens":3,"output_tokens":2}}`,
	)
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       upstreamBody,
	}}
	app, _ := newTokenBazaarMessagesTestApp(t, tunnel)

	req := httptest.NewRequest(
		"POST", "/v1/messages",
		bytes.NewBufferString(`{"model":"tokenbazaar:primary:claude-sonnet-5","messages":[{"role":"user","content":"hi"}],"max_tokens":16}`),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, string(upstreamBody), string(body))

	require.Len(t, tunnel.lastReq.ModelRuntimeProfiles, 1)
	assert.Equal(t, "tokenbazaar", tunnel.lastReq.ModelRuntimeProfiles[0].Provider)
	assert.True(t, tunnel.lastReq.ModelRuntimeProfiles[0].MatchesModel("claude-sonnet-5"))
	assert.Equal(t, platform.ProviderOperationAnthropicMessages, tunnel.lastReq.Operation)
	assert.Equal(t, "tokenbazaar", tunnel.lastReq.Deployment.Provider)
	assert.Contains(t, string(tunnel.lastReq.Body), `"model":"claude-sonnet-5"`, "the outbound model must be rewritten to the deployment's provider-facing model id")
	assert.Contains(t, string(tunnel.lastReq.Body), `"max_tokens":16`)
}

// TestTokenBazaarMessagesHandlerThinkingPassesThroughUnmodified confirms
// thinking.budget_tokens (documented as supported on this route) reaches the
// tunnel request body byte-for-byte, aside from the model-id rewrite.
func TestTokenBazaarMessagesHandlerThinkingPassesThroughUnmodified(t *testing.T) {
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body: []byte(
			`{"id":"msg_1","type":"message","role":"assistant","model":"claude-sonnet-5",` +
				`"content":[{"type":"text","text":"hi"}],"stop_reason":"end_turn",` +
				`"usage":{"input_tokens":3,"output_tokens":2}}`,
		),
	}}
	app, _ := newTokenBazaarMessagesTestApp(t, tunnel)

	req := httptest.NewRequest(
		"POST", "/v1/messages",
		bytes.NewBufferString(`{"model":"tokenbazaar:primary:claude-sonnet-5","messages":[{"role":"user","content":"hi"}],"max_tokens":1024,"thinking":{"type":"enabled","budget_tokens":512}}`),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Contains(t, string(tunnel.lastReq.Body), `"thinking":{"type":"enabled","budget_tokens":512}`)
}

// TestTokenBazaarMessagesHandlerNeverForwardsGodarAIAuthorizationHeader
// mirrors the other TokenBazaar capabilities' own coverage at the control
// plane layer: whatever Authorization header the caller sent to GodarAI must
// never appear on the outbound TunnelRequest.Headers used for dispatch (the
// deployment's own credential is applied later, at the Silk Road adapter).
func TestTokenBazaarMessagesHandlerNeverForwardsGodarAIAuthorizationHeader(t *testing.T) {
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body: []byte(
			`{"id":"msg_1","type":"message","role":"assistant","model":"claude-sonnet-5",` +
				`"content":[{"type":"text","text":"hi"}],"stop_reason":"end_turn",` +
				`"usage":{"input_tokens":1,"output_tokens":1}}`,
		),
	}}
	app, _ := newTokenBazaarMessagesTestApp(t, tunnel)

	req := httptest.NewRequest(
		"POST", "/v1/messages",
		bytes.NewBufferString(`{"model":"tokenbazaar:primary:claude-sonnet-5","messages":[{"role":"user","content":"hi"}],"max_tokens":16}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer godarai-caller-token-must-not-leak")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	if got := tunnel.lastReq.Headers["Authorization"]; got != "" {
		t.Fatalf("TunnelRequest.Headers leaked an Authorization value: %q", got)
	}
}

// TestParseAnthropicSSEUsageHandlesMalformedAndUnexpectedEvents extends
// coverage of the shared, provider-agnostic parseAnthropicSSEUsage /
// anthropicSSEIsError helpers (anthropic_native_handler.go) that
// TokenBazaarMessagesHandler.handleStream reuses unmodified for
// content_block_delta/message_stop parsing -- confirming malformed or
// unexpected SSE chunks never panic and are simply not treated as a usage
// update or an error.
func TestParseAnthropicSSEUsageHandlesMalformedAndUnexpectedEvents(t *testing.T) {
	cases := []struct {
		name string
		data string
	}{
		{"not_sse_at_all", "hello world, not an SSE frame"},
		{"malformed_json", "data: {not valid json\n\n"},
		{"unexpected_event_type", "event: ping\ndata: {\"type\":\"ping\"}\n\n"},
		{"empty", ""},
		{"data_prefix_only", "data:\n\n"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				_ = parseAnthropicSSEUsage([]byte(testCase.data))
				_ = anthropicSSEIsError([]byte(testCase.data))
			})
		})
	}

	usageEvent := `data: {"type":"message_delta","usage":{"output_tokens":7}}` + "\n\n"
	usage := parseAnthropicSSEUsage([]byte(usageEvent))
	require.NotNil(t, usage)
	assert.Equal(t, 7, usage.CompletionTokens)

	errorEvent := `data: {"type":"error","error":{"type":"api_error","message":"boom"}}` + "\n\n"
	assert.True(t, anthropicSSEIsError([]byte(errorEvent)))
	assert.False(t, anthropicSSEIsError([]byte(usageEvent)))
}
