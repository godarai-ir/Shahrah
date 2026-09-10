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

type tokenBazaarResponsesBudgeter struct{ testBudgeter }

func (b *tokenBazaarResponsesBudgeter) EstimateCost(ctx context.Context, req contractsopenai.ChatCompletionRequest) (platform.UsageCostBreakdown, error) {
	estimated, err := b.testBudgeter.EstimateCost(ctx, req)
	if err != nil {
		return estimated, err
	}

	binding := tokenBazaarProxyTestRuntimeProfile(req.Model)
	estimated.ModelRuntimeProfile = &binding
	estimated.PricingRevision = binding.CatalogRevision

	return estimated, nil
}

// newTokenBazaarResponsesTestApp builds a Responses test app routed to a
// single "tokenbazaar" deployment for "gpt-5.5", with an injectable
// ProviderFeatureConfig so tests can exercise
// SHAHRAH_TOKENBAZAAR_RESPONSES_ENABLED both on and off.
func newTokenBazaarResponsesTestApp(
	t *testing.T,
	tunnel *captureTunnel,
	features ProviderFeatureConfig,
) (*fiber.App, *testLogs) {
	t.Helper()

	mem, err := testcache.New(t, zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, mem.Close()) })

	rt, err := router.NewRouter([]config.ModelDeployment{{
		ModelName: "gpt-5.5",
		ProviderEndpoint: config.ProviderEndpointConfig{
			Provider: "tokenbazaar",
			Model:    "gpt-5.5",
			APIKey:   "test-key",
		},
	}}, config.RouterSettings{CooldownTime: 60, NumRetries: 1}, mem, nil, zap.NewNop())
	require.NoError(t, err)

	logs := &testLogs{}

	handler := NewResponsesHandlerWithDependencies(BaseHandlerDependencies{
		Router:              rt,
		Tunnel:              tunnel,
		Billing:             &tokenBazaarResponsesBudgeter{},
		BillingCapabilities: compatibilityBillingCapabilities(),
		Wallets:             &testWallets{},
		RequestLogger:       logs,
		SpendLogger:         logs,
		Logger:              zap.NewNop(),
		ProviderFeatures:    features,
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

// TestResponsesHandlerTokenBazaarResponsesDisabledByDefault confirms
// TokenBazaar Responses dispatch fails closed (503 provider_disabled) when
// TokenBazaarResponsesEnabled is not explicitly set -- the zero-value
// ProviderFeatureConfig -- and that the tunnel is never reached while the
// flag is off.
func TestResponsesHandlerTokenBazaarResponsesDisabledByDefault(t *testing.T) {
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       []byte(`{"id":"resp_1","object":"response","status":"completed"}`),
	}}
	app, _ := newTokenBazaarResponsesTestApp(t, tunnel, ProviderFeatureConfig{})

	req := httptest.NewRequest("POST", "/v1/responses", bytes.NewBufferString(`{"model":"gpt-5.5","input":"hello"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)
	assert.Empty(t, tunnel.lastReq.RequestID, "the tunnel must never be dispatched to while the responses flag is off")

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "provider_disabled")
	assert.Contains(t, string(body), "tokenbazaar responses dispatch is disabled")
}

// TestResponsesHandlerTokenBazaarResponsesFlagIsolatedFromOtherThreeFlags
// confirms SHAHRAH_TOKENBAZAAR_RESPONSES_ENABLED is evaluated independently
// of SHAHRAH_TOKENBAZAAR_ENABLED (chat), SHAHRAH_TOKENBAZAAR_IMAGES_ENABLED
// (images), and SHAHRAH_TOKENBAZAAR_EMBEDDINGS_ENABLED (embeddings) in both
// directions, mirroring
// TestEmbeddingsHandlerTokenBazaarEmbeddingsFlagIsolatedFromChatAndImagesFlags.
func TestResponsesHandlerTokenBazaarResponsesFlagIsolatedFromOtherThreeFlags(t *testing.T) {
	t.Run("responses on, chat/images/embeddings off: passes the responses gate", func(t *testing.T) {
		tunnel := &captureTunnel{resp: &platform.TunnelResponse{
			StatusCode: fiber.StatusOK,
			Headers:    map[string]string{"content-type": "application/json"},
			Body:       []byte(`{"id":"resp_1","object":"response","status":"completed","model":"gpt-5.5","usage":{"input_tokens":1,"output_tokens":1,"total_tokens":2}}`),
		}}
		app, _ := newTokenBazaarResponsesTestApp(t, tunnel, ProviderFeatureConfig{
			Configured:                   true,
			TokenBazaarEnabled:           false,
			TokenBazaarImagesEnabled:     false,
			TokenBazaarEmbeddingsEnabled: false,
			TokenBazaarResponsesEnabled:  true,
		})

		req := httptest.NewRequest("POST", "/v1/responses", bytes.NewBufferString(`{"model":"gpt-5.5","input":"hello"}`))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		defer closeResponseBody(t, resp)

		require.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		// The request may still fail later for an unrelated reason, but it
		// must never be rejected specifically because responses dispatch is
		// disabled -- that is the one thing this subtest isolates.
		assert.NotContains(t, string(body), "tokenbazaar responses dispatch is disabled")
	})

	t.Run("chat/images/embeddings on, responses off: still fails closed", func(t *testing.T) {
		tunnel := &captureTunnel{resp: &platform.TunnelResponse{
			StatusCode: fiber.StatusOK,
			Headers:    map[string]string{"content-type": "application/json"},
			Body:       []byte(`{"id":"resp_1","object":"response","status":"completed"}`),
		}}
		app, _ := newTokenBazaarResponsesTestApp(t, tunnel, ProviderFeatureConfig{
			Configured:                   true,
			TokenBazaarEnabled:           true,
			TokenBazaarImagesEnabled:     true,
			TokenBazaarEmbeddingsEnabled: true,
			TokenBazaarResponsesEnabled:  false,
		})

		req := httptest.NewRequest("POST", "/v1/responses", bytes.NewBufferString(`{"model":"gpt-5.5","input":"hello"}`))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		defer closeResponseBody(t, resp)

		require.NoError(t, err)
		assert.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)
		assert.Empty(t, tunnel.lastReq.RequestID, "the tunnel must never be dispatched to while the responses flag is off, even with chat/images/embeddings on")

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), "provider_disabled")
	})
}

// TestResponsesHandlerSupportsTokenBazaarWithReasoningEffortPassthrough
// confirms, once the flag is enabled, that reasoning.effort -- documented
// only on /v1/responses, never on chat completions -- reaches the tunnel
// request body unmodified for a TokenBazaar deployment, mirroring
// TestResponsesHandlerSupportsXAIWithPinnedProfileAndProviderFields. The
// generic request/response translation in this file has no TokenBazaar- or
// even OpenAI-specific branch that would strip or rewrite this field; it
// passes through purely because TokenBazaar is now in the
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
	app, _ := newTokenBazaarResponsesTestApp(t, tunnel, ProviderFeatureConfig{
		Configured:                  true,
		TokenBazaarResponsesEnabled: true,
	})

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
