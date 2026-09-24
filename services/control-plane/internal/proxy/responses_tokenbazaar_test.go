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

// tokenBazaarProxyTestRuntimeProfile mirrors deepSeekProxyRuntimeProfile's
// known-valid shape (see deepseek_runtime_profile_test.go) so
// platform.ModelRuntimeProfileBinding.Validate() passes, but for the
// TokenBazaar provider and one of the 6 in-scope Responses models.
func tokenBazaarProxyTestRuntimeProfile(model string) platform.ModelRuntimeProfileBinding {
	return platform.ModelRuntimeProfileBinding{
		Provider:        "tokenbazaar",
		CatalogRevision: "tokenbazaar-test-revision",
		CanonicalID:     "tokenbazaar/" + model,
		Slug:            model,
		ProfileVersion:  1,
		EffectiveFrom:   time.Now().UTC(),
		Profile: platform.ModelRuntimeProfile{
			Version:    1,
			Endpoints:  []string{"models", "chat_completions", "responses"},
			Transports: []string{"http", "sse"},
			Lifecycle:  "active",
			Thinking: platform.ModelRuntimeThinkingProfile{
				DefaultMode: "adaptive", AllowedModes: []string{"adaptive", "disabled"},
				EffortValues: []string{"low", "medium", "high"},
			},
			Sampling: platform.ModelRuntimeSamplingProfile{
				Parameters: []string{"temperature", "top_p"},
			},
			Tools: platform.ModelRuntimeToolsProfile{
				Custom: true, StructuredOutput: true, Supported: []string{"function_calling"},
			},
			Modalities: platform.ModelRuntimeModalities{Input: []string{"text"}, Output: []string{"text"}},
			Parameters: map[string][]string{
				"responses": {"input", "instructions", "max_output_tokens", "reasoning", "stream", "temperature", "top_p", "text", "tools", "tool_choice"},
			},
			Limits: platform.ModelRuntimeLimits{
				ContextWindow: 400000, MaxInputTokens: 400000, MaxOutputTokens: 128000,
				MessagesBytes: 32 * 1024 * 1024, BatchBytes: 256 * 1024 * 1024, FileBytes: 500 * 1024 * 1024,
			},
		},
	}
}

// tokenBazaarResponsesBudgeter attaches a TokenBazaar runtime profile to every
// estimate. A non-nil endpoints value replaces the profile's default endpoint
// list.
type tokenBazaarResponsesBudgeter struct {
	testBudgeter

	endpoints []string
}

func (b *tokenBazaarResponsesBudgeter) EstimateCost(ctx context.Context, req contractsopenai.ChatCompletionRequest) (platform.UsageCostBreakdown, error) {
	estimated, err := b.testBudgeter.EstimateCost(ctx, req)
	if err != nil {
		return estimated, err
	}

	binding := tokenBazaarProxyTestRuntimeProfile(req.Model)
	if b.endpoints != nil {
		binding.Profile.Endpoints = append([]string(nil), b.endpoints...)
		for endpoint := range binding.Profile.Parameters {
			if !binding.Profile.AllowsEndpoint(endpoint) {
				delete(binding.Profile.Parameters, endpoint)
			}
		}
	}

	estimated.ModelRuntimeProfile = &binding
	estimated.PricingRevision = binding.CatalogRevision

	return estimated, nil
}

// newTokenBazaarResponsesTestApp builds a Responses test app routed to a
// single "tokenbazaar" deployment for "gpt-5.5".
func newTokenBazaarResponsesTestApp(
	t *testing.T,
	tunnel *captureTunnel,
) (*fiber.App, *testLogs) {
	t.Helper()

	return newTokenBazaarResponsesTestAppForModel(t, tunnel, "gpt-5.5", nil)
}

// newTokenBazaarResponsesTestAppForModel builds a Responses test app routed to
// a single "tokenbazaar" deployment for model. A non-nil endpoints value
// replaces the runtime profile's endpoint list.
func newTokenBazaarResponsesTestAppForModel(
	t *testing.T,
	tunnel *captureTunnel,
	model string,
	endpoints []string,
) (*fiber.App, *testLogs) {
	t.Helper()

	mem, err := testcache.New(t, zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, mem.Close()) })

	rt, err := router.NewRouter([]config.ModelDeployment{{
		ModelName: model,
		ProviderEndpoint: config.ProviderEndpointConfig{
			Provider: "tokenbazaar",
			Model:    model,
			APIKey:   "test-key",
		},
	}}, config.RouterSettings{CooldownTime: 60, NumRetries: 1}, mem, nil, zap.NewNop())
	require.NoError(t, err)

	logs := &testLogs{}

	handler := NewResponsesHandlerWithDependencies(BaseHandlerDependencies{
		Router:              rt,
		Tunnel:              tunnel,
		Billing:             &tokenBazaarResponsesBudgeter{endpoints: endpoints},
		BillingCapabilities: compatibilityBillingCapabilities(),
		Wallets:             &testWallets{},
		RequestLogger:       logs,
		SpendLogger:         logs,
		Logger:              zap.NewNop(),
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

	return app, logs
}

// TestTokenBazaarResponsesCapabilityIsCreateOnly mirrors
// TestDeepSeekResponsesCapabilityIsCreateOnly: TokenBazaar must be added to
// supportsResponsesCreateProvider but MUST NOT be added to
// supportsResponseResourcesProvider -- the stateful GET/DELETE/CANCEL/
// InputItems/Compact/InputTokens surface TokenBazaar does not offer.
func TestTokenBazaarResponsesCapabilityIsCreateOnly(t *testing.T) {
	assert.True(t, supportsResponsesCreateProvider("tokenbazaar"))
	assert.False(t, supportsResponseResourcesProvider("tokenbazaar"))
	assert.True(t, supportsResponseResourcesProvider("openai"))
	assert.True(t, supportsResponseResourcesProvider("xai"))
}

// TestResponsesHandlerSupportsTokenBazaarWithReasoningEffortPassthrough
// confirms reasoning.effort -- documented only on /v1/responses, never on
// chat completions -- reaches the tunnel request body unmodified for a
// TokenBazaar deployment, mirroring
// TestResponsesHandlerSupportsXAIWithPinnedProfileAndProviderFields. The
// generic request/response translation in this file has no TokenBazaar- or
// even OpenAI-specific branch that would strip or rewrite this field; it
// passes through purely because TokenBazaar is in the
// supportsResponsesCreateProvider allowlist.
func TestResponsesHandlerSupportsTokenBazaarWithReasoningEffortPassthrough(t *testing.T) {
	upstreamBody := []byte(
		`{"id":"resp_tb","object":"response","model":"gpt-5.5","status":"completed",` +
			`"output":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"hi"}]}],` +
			`"usage":{"input_tokens":3,"output_tokens":2,"total_tokens":5}}`,
	)
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       upstreamBody,
	}}
	app, _ := newTokenBazaarResponsesTestApp(t, tunnel)

	req := httptest.NewRequest(
		"POST",
		"/v1/responses",
		bytes.NewBufferString(`{"model":"gpt-5.5","input":"hello","reasoning":{"effort":"low"}}`),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	require.Len(t, tunnel.lastReq.ModelRuntimeProfiles, 1)
	assert.Equal(t, "tokenbazaar", tunnel.lastReq.ModelRuntimeProfiles[0].Provider)
	assert.True(t, tunnel.lastReq.ModelRuntimeProfiles[0].MatchesModel("gpt-5.5"))
	assert.Equal(t, platform.ProviderOperationOpenAIResponsesCreate, tunnel.lastReq.Operation)
	assert.Contains(t, string(tunnel.lastReq.Body), `"reasoning":{"effort":"low"}`)
}

// TestResponsesHandlerRejectsGeminiWithoutResponsesEndpoint covers a Gemini
// model reached through TokenBazaar. The official Gemini API has no Responses
// endpoint, so the catalog profile omits it and the request must fail with a
// 400 before anything is dispatched to Silk Road.
func TestResponsesHandlerRejectsGeminiWithoutResponsesEndpoint(t *testing.T) {
	tunnel := &captureTunnel{}
	app, _ := newTokenBazaarResponsesTestAppForModel(
		t,
		tunnel,
		"gemini-2.5-pro",
		[]string{"models", "chat_completions", "generate_content", "stream_generate_content"},
	)

	req := httptest.NewRequest(
		"POST",
		"/v1/responses",
		bytes.NewBufferString(`{"model":"gemini-2.5-pro","input":"hello"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, string(body))
	assert.Contains(t, string(body), "does not support the Responses API")
	assert.Empty(t, tunnel.requests)
	assert.Empty(t, tunnel.lastReq.Operation)
}

func TestRuntimeProfilesAllowResponses(t *testing.T) {
	withResponses := tokenBazaarProxyTestRuntimeProfile("gpt-5.5")
	withoutResponses := tokenBazaarProxyTestRuntimeProfile("gemini-2.5-pro")
	withoutResponses.Profile.Endpoints = []string{"chat_completions", "generate_content"}

	assert.True(t, runtimeProfilesAllowResponses(nil))
	assert.True(t, runtimeProfilesAllowResponses([]platform.ModelRuntimeProfileBinding{withResponses}))
	assert.False(t, runtimeProfilesAllowResponses([]platform.ModelRuntimeProfileBinding{withoutResponses}))
	assert.False(t, runtimeProfilesAllowResponses([]platform.ModelRuntimeProfileBinding{withResponses, withoutResponses}))
}
