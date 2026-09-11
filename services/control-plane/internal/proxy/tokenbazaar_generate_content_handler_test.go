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

// tokenBazaarGenerateContentProxyTestRuntimeProfile mirrors
// tokenBazaarMessagesProxyTestRuntimeProfile but for the
// "generate_content"/"stream_generate_content" endpoints and one of the 6
// in-scope Gemini text models.
func tokenBazaarGenerateContentProxyTestRuntimeProfile(model string) platform.ModelRuntimeProfileBinding {
	return platform.ModelRuntimeProfileBinding{
		Provider:        "tokenbazaar",
		CatalogRevision: "tokenbazaar-generate-content-test-revision",
		CanonicalID:     "tokenbazaar/" + model,
		Slug:            model,
		ProfileVersion:  1,
		EffectiveFrom:   time.Now().UTC(),
		Profile: platform.ModelRuntimeProfile{
			Version:    1,
			Endpoints:  []string{"models", "chat_completions", "generate_content", "stream_generate_content"},
			Transports: []string{"http", "sse"},
			Lifecycle:  "active",
			Thinking:   platform.ModelRuntimeThinkingProfile{DefaultMode: "disabled", AllowedModes: []string{"disabled"}},
			Sampling:   platform.ModelRuntimeSamplingProfile{Parameters: []string{"temperature", "top_p"}},
			Tools:      platform.ModelRuntimeToolsProfile{Custom: true, Supported: []string{"function_calling"}},
			Modalities: platform.ModelRuntimeModalities{Input: []string{"text", "image"}, Output: []string{"text"}},
			Parameters: map[string][]string{
				"generate_content": {"contents", "generationConfig", "systemInstruction"},
			},
			Limits: platform.ModelRuntimeLimits{
				ContextWindow: 1048576, MaxInputTokens: 1048576, MaxOutputTokens: 65536,
				MessagesBytes: 32 * 1024 * 1024, BatchBytes: 1, FileBytes: 1,
			},
		},
	}
}

type tokenBazaarGenerateContentBudgeter struct{ testBudgeter }

func (b *tokenBazaarGenerateContentBudgeter) EstimateCost(ctx context.Context, req contractsopenai.ChatCompletionRequest) (platform.UsageCostBreakdown, error) {
	estimated, err := b.testBudgeter.EstimateCost(ctx, req)
	if err != nil {
		return estimated, err
	}

	binding := tokenBazaarGenerateContentProxyTestRuntimeProfile(providerModelID(req.Model))
	estimated.ModelRuntimeProfile = &binding
	estimated.PricingRevision = binding.CatalogRevision

	return estimated, nil
}

// newTokenBazaarGenerateContentTestApp builds a
// POST /v1beta/models/*:generateContent test app wired through
// NewGeminiGenerateContentRouter exactly as routes_proxy.go wires it: a
// tokenbazaar-FQN'd {model} path segment dispatches to
// TokenBazaarGenerateContentHandler; every other request falls back to a
// stand-in "native" handler (a real GeminiNativeHandler is out of scope for
// this narrow test -- the routing shim itself, not GeminiNativeHandler's own
// already-tested behavior, is what this delivery needs to prove).
func newTokenBazaarGenerateContentTestApp(
	t *testing.T,
	tunnel *captureTunnel,
) (*fiber.App, *testLogs) {
	t.Helper()

	mem, err := testcache.New(t, zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, mem.Close()) })

	rt, err := router.NewRouter([]config.ModelDeployment{{
		ModelName: "tokenbazaar:primary:gemini-2.5-flash",
		ProviderEndpoint: config.ProviderEndpointConfig{
			Provider: "tokenbazaar",
			Model:    "gemini-2.5-flash",
			APIKey:   "test-key",
		},
	}}, config.RouterSettings{CooldownTime: 60, NumRetries: 1}, mem, nil, zap.NewNop())
	require.NoError(t, err)

	logs := &testLogs{}

	handler := NewTokenBazaarGenerateContentHandlerWithDependencies(BaseHandlerDependencies{
		Router:              rt,
		Tunnel:              tunnel,
		Billing:             &tokenBazaarGenerateContentBudgeter{},
		BillingCapabilities: compatibilityBillingCapabilities(),
		Wallets:             &testWallets{},
		RequestLogger:       logs,
		SpendLogger:         logs,
		Logger:              zap.NewNop(),
	})

	native := func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusTeapot).SendString("native-gemini-handler-called")
	}

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
	app.Post("/v1beta/models/*", NewGeminiGenerateContentRouter(handler, native))

	return app, logs
}

// TestParseGeminiModelActionPath confirms the path parser splits on the LAST
// colon (correct for a TokenBazaar-FQN'd model, which itself contains two
// colons), not the first.
func TestParseGeminiModelActionPath(t *testing.T) {
	cases := []struct {
		path       string
		wantModel  string
		wantAction string
		wantOK     bool
	}{
		{"/v1beta/models/gemini-2.5-flash:generateContent", "gemini-2.5-flash", "generateContent", true},
		{"/v1beta/models/tokenbazaar:primary:gemini-2.5-flash:generateContent", "tokenbazaar:primary:gemini-2.5-flash", "generateContent", true},
		{"/v1beta/models/tokenbazaar:primary:gemini-2.5-flash:streamGenerateContent", "tokenbazaar:primary:gemini-2.5-flash", "streamGenerateContent", true},
		{"/v1beta/models/gemini-2.5-flash:countTokens", "gemini-2.5-flash", "countTokens", true},
		{"/v1beta/models/gemini-2.5-flash", "", "", false},
		{"/v1beta/models", "", "", false},
		{"/v1/chat/completions", "", "", false},
	}

	for _, testCase := range cases {
		model, action, ok := parseGeminiModelActionPath(testCase.path)
		if ok != testCase.wantOK || model != testCase.wantModel || action != testCase.wantAction {
			t.Fatalf(
				"parseGeminiModelActionPath(%q) = (%q, %q, %t); want (%q, %q, %t)",
				testCase.path, model, action, ok, testCase.wantModel, testCase.wantAction, testCase.wantOK,
			)
		}
	}
}

// TestRequestTargetsTokenBazaarGenerateContent confirms the routing shim's
// sole signal (an explicit "tokenbazaar:<account>:<model>" FQN in the
// {model} path segment of a generateContent/streamGenerateContent request)
// is evaluated correctly: a bare model id, a non-generateContent action, and
// a direct "google-gemini" FQN must never be treated as a TokenBazaar
// request.
func TestRequestTargetsTokenBazaarGenerateContent(t *testing.T) {
	cases := map[string]bool{
		"/v1beta/models/tokenbazaar:primary:gemini-2.5-flash:generateContent":       true,
		"/v1beta/models/tokenbazaar:primary:gemini-2.5-flash:streamGenerateContent": true,
		"/v1beta/models/gemini-2.5-flash:generateContent":                           false,
		"/v1beta/models/google-gemini:primary:gemini-2.5-flash:generateContent":     false,
		"/v1beta/models/tokenbazaar:primary:gemini-2.5-flash:countTokens":           false,
		"/v1beta/models/gemini-2.5-flash":                                           false,
		"/v1beta/models":                                                            false,
	}

	for path, want := range cases {
		if got := requestTargetsTokenBazaarGenerateContent(path); got != want {
			t.Fatalf("requestTargetsTokenBazaarGenerateContent(%s) = %t, want %t", path, got, want)
		}
	}
}

// TestGeminiGenerateContentRouterFallsBackToNativeForNonTokenBazaarModels
// proves the shim leaves GeminiNativeHandler's own behavior for the direct
// "google-gemini" provider (and any bare model id) completely untouched.
func TestGeminiGenerateContentRouterFallsBackToNativeForNonTokenBazaarModels(t *testing.T) {
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{StatusCode: fiber.StatusOK}}
	app, _ := newTokenBazaarGenerateContentTestApp(t, tunnel)

	req := httptest.NewRequest("POST", "/v1beta/models/gemini-2.5-flash:generateContent", bytes.NewBufferString(`{"contents":[{"role":"user","parts":[{"text":"hi"}]}]}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusTeapot, resp.StatusCode, "a bare model id must fall back to the native Gemini handler")

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, "native-gemini-handler-called", string(body))
	assert.Empty(t, tunnel.lastReq.RequestID, "TokenBazaarGenerateContentHandler must never dispatch a request meant for the native handler")
}

// TestTokenBazaarGenerateContentHandlerDispatchesSuccessfully confirms a
// well-formed request reaches the tunnel with the Gemini generateContent
// operation, the deployment's provider-facing model, and a runtime profile
// bound to "tokenbazaar", and that a successful Gemini-shaped response is
// returned to the caller unchanged.
func TestTokenBazaarGenerateContentHandlerDispatchesSuccessfully(t *testing.T) {
	upstreamBody := []byte(
		`{"candidates":[{"content":{"role":"model","parts":[{"text":"hi there"}]},"finishReason":"STOP"}],` +
			`"usageMetadata":{"promptTokenCount":3,"candidatesTokenCount":2,"totalTokenCount":5}}`,
	)
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       upstreamBody,
	}}
	app, _ := newTokenBazaarGenerateContentTestApp(t, tunnel)

	req := httptest.NewRequest(
		"POST", "/v1beta/models/tokenbazaar:primary:gemini-2.5-flash:generateContent",
		bytes.NewBufferString(`{"contents":[{"role":"user","parts":[{"text":"hi"}]}]}`),
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
	assert.True(t, tunnel.lastReq.ModelRuntimeProfiles[0].MatchesModel("gemini-2.5-flash"))
	assert.Equal(t, platform.ProviderOperationGeminiGenerateContent, tunnel.lastReq.Operation)
	assert.Equal(t, "tokenbazaar", tunnel.lastReq.Deployment.Provider)
	assert.Equal(t, "gemini-2.5-flash", tunnel.lastReq.Deployment.Model, "the outbound model must be propagated via Deployment.Model since generateContent has no body \"model\" field")
	assert.False(t, tunnel.lastReq.Stream)
}

// TestTokenBazaarStreamGenerateContentHandlerSetsStreamOperation confirms a
// streamGenerateContent request is dispatched with Stream=true and the
// distinct stream operation.
func TestTokenBazaarStreamGenerateContentHandlerSetsStreamOperation(t *testing.T) {
	stream := newFakeTunnelStream(nil, 1, 1)
	close(stream.chunks)

	tunnel := &captureTunnel{streams: []platform.TunnelStream{stream}}
	app, _ := newTokenBazaarGenerateContentTestApp(t, tunnel)

	req := httptest.NewRequest(
		"POST", "/v1beta/models/tokenbazaar:primary:gemini-2.5-flash:streamGenerateContent",
		bytes.NewBufferString(`{"contents":[{"role":"user","parts":[{"text":"hi"}]}]}`),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	require.True(t, tunnel.lastReq.Stream)
	assert.Equal(t, platform.ProviderOperationGeminiStreamGenerateContent, tunnel.lastReq.Operation)
}

// TestTokenBazaarGenerateContentHandlerNeverForwardsGodarAIAuthorizationHeader
// mirrors the other TokenBazaar capabilities' own coverage at the control
// plane layer: whatever Authorization header the caller sent to GodarAI must
// never appear on the outbound TunnelRequest.Headers used for dispatch (the
// deployment's own credential is applied later, at the Silk Road adapter).
func TestTokenBazaarGenerateContentHandlerNeverForwardsGodarAIAuthorizationHeader(t *testing.T) {
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body: []byte(
			`{"candidates":[{"content":{"role":"model","parts":[{"text":"hi"}]},"finishReason":"STOP"}]}`,
		),
	}}
	app, _ := newTokenBazaarGenerateContentTestApp(t, tunnel)

	req := httptest.NewRequest(
		"POST", "/v1beta/models/tokenbazaar:primary:gemini-2.5-flash:generateContent",
		bytes.NewBufferString(`{"contents":[{"role":"user","parts":[{"text":"hi"}]}]}`),
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
