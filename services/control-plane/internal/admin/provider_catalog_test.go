package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapDiscoveryModels_PreservesGatewayAndUIFields(t *testing.T) {
	inputPrice := 0.4
	outputPrice := 1.2
	contextWindow := 128000

	items := mapDiscoveryModels([]modelDiscoveryModel{
		{
			CanonicalID:      "gpt-4.1-mini",
			Name:             "GPT-4.1 Mini",
			Mode:             []string{"chat", "embedding"},
			ContextWindow:    &contextWindow,
			Capabilities:     []string{"chat", "reasoning"},
			InputModalities:  []string{"text"},
			OutputModalities: []string{"text"},
			Pricing: &modelDiscoveryPricing{
				InputPerMillionTokens:  &inputPrice,
				OutputPerMillionTokens: &outputPrice,
			},
		},
	})

	require.Len(t, items, 1)
	assert.Equal(t, "gpt-4.1-mini", items[0].Model)
	assert.Equal(t, "GPT-4.1 Mini", items[0].DisplayName)
	assert.Equal(t, []string{"chat", "embedding"}, items[0].Modes)
	assert.Equal(t, []platform.ModelType{platform.ModelTypeChat, platform.ModelTypeEmbedding}, items[0].Types)
	assert.Equal(t, &contextWindow, items[0].ContextWindow)
	assert.Equal(t, []string{"text"}, items[0].InputModalities)
	assert.Equal(t, []string{"text"}, items[0].OutputModalities)
	assert.Equal(t, []string{"chat", "reasoning"}, items[0].Capabilities)
	assert.Equal(t, []string{"chat", "reasoning"}, items[0].Features)
	require.NotNil(t, items[0].PublicPricing)
	assert.Equal(t, &inputPrice, items[0].PublicPricing.InputUSDPer1M)
	assert.Equal(t, &outputPrice, items[0].PublicPricing.OutputUSDPer1M)
}

func TestResolveProviderAccountModelSelections_ExpandsCatalogMetadata(t *testing.T) {
	inputPrice := 0.15
	outputPrice := 0.6

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/providers/openai/models" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"canonicalId":"gpt-4.1-mini","name":"GPT-4.1 Mini","mode":["chat"],"capabilities":["chat","function_calling"],"inputModalities":["text"],"outputModalities":["text"],"pricing":{"inputPerMillionTokens":0.15,"outputPerMillionTokens":0.6}}]}`))
	}))
	defer server.Close()

	t.Setenv("SHAHRAH_MODEL_DISCOVERY_URL", server.URL)
	invalidateDiscoveryCache()

	items, invalid, invalidCapabilities, err := resolveProviderAccountModelSelections(context.Background(), "openai", []platform.UpsertTenantProviderAccountModelSelection{
		{
			Model:                 "gpt-4.1-mini",
			EffectiveCapabilities: []platform.ModelType{platform.ModelTypeChat},
			PricingMode:           platform.PricingModePublic,
		},
	})
	require.NoError(t, err)
	require.Empty(t, invalid)
	require.Empty(t, invalidCapabilities)
	require.Len(t, items, 1)

	assert.Equal(t, "gpt-4.1-mini", items[0].Model)
	assert.True(t, items[0].Enabled)
	assert.NotEmpty(t, items[0].Types)
	assert.NotEmpty(t, items[0].InputModalities)
	assert.NotEmpty(t, items[0].OutputModalities)
	assert.NotNil(t, items[0].PublicPricing)
	assert.Equal(t, &inputPrice, items[0].PublicPricing.InputUSDPer1M)
	assert.Equal(t, &outputPrice, items[0].PublicPricing.OutputUSDPer1M)
	assert.Equal(t, platform.PricingModePublic, items[0].PricingMode)
}

func TestResolveProviderAccountModelSelections_RejectsInvalidCapabilities(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/providers/openai/models" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(
			`{"data":[` +
				`{"canonicalId":"gpt-4.1-mini","name":"GPT-4.1 Mini","mode":["chat","embedding"],"capabilities":["chat"],"inputModalities":["text"],"outputModalities":["text"]},` +
				`{"canonicalId":"gpt-4.1","name":"GPT-4.1","mode":["chat"],"capabilities":["chat"],"inputModalities":["text"],"outputModalities":["text"]}` +
				`]}`,
		))
	}))
	defer server.Close()

	t.Setenv("SHAHRAH_MODEL_DISCOVERY_URL", server.URL)
	invalidateDiscoveryCache()

	items, invalid, invalidCapabilities, err := resolveProviderAccountModelSelections(context.Background(), "openai", []platform.UpsertTenantProviderAccountModelSelection{
		{
			Model:                 "gpt-4.1-mini",
			EffectiveCapabilities: []platform.ModelType{platform.ModelTypeEmbedding},
			PricingMode:           platform.PricingModePublic,
		},
		{
			Model:                 "gpt-4.1",
			EffectiveCapabilities: []platform.ModelType{platform.ModelTypeRerank},
			PricingMode:           platform.PricingModePublic,
		},
		{
			Model:                 "gpt-4.1-mini-2",
			EffectiveCapabilities: []platform.ModelType{platform.ModelTypeRerank},
			PricingMode:           platform.PricingModePublic,
		},
	})
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, []string{"gpt-4.1-mini-2"}, invalid)
	require.Contains(t, invalidCapabilities, "gpt-4.1")
	assert.Equal(t, []string{"rerank"}, invalidCapabilities["gpt-4.1"])
	assert.Equal(t, []platform.ModelType{platform.ModelTypeEmbedding}, items[0].Types)
}

func TestResolveProviderAccountModelSelections_AcceptsResponsesAliasAsCompletion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/providers/openai/models" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"canonicalId":"gpt-4.1-nano","name":"GPT-4.1 Nano","mode":["chat","responses"],"capabilities":["chat"],"inputModalities":["text"],"outputModalities":["text"]}]}`))
	}))
	defer server.Close()

	t.Setenv("SHAHRAH_MODEL_DISCOVERY_URL", server.URL)
	invalidateDiscoveryCache()

	items, invalid, invalidCapabilities, err := resolveProviderAccountModelSelections(context.Background(), "openai", []platform.UpsertTenantProviderAccountModelSelection{
		{
			Model:                 "gpt-4.1-nano",
			EffectiveCapabilities: []platform.ModelType{platform.ModelTypeChat, "responses"},
			PricingMode:           platform.PricingModePublic,
		},
	})
	require.NoError(t, err)
	require.Empty(t, invalid)
	require.Empty(t, invalidCapabilities)
	require.Len(t, items, 1)
	assert.Equal(t, []platform.ModelType{platform.ModelTypeChat, platform.ModelTypeCompletion}, items[0].Types)
}

func TestDiscoveryProviders_FailsWithoutFallbackWhenDiscoveryUnavailable(t *testing.T) {
	t.Setenv("SHAHRAH_MODEL_DISCOVERY_URL", "http://127.0.0.1:1")
	invalidateDiscoveryCache()

	items, err := discoveryProviders(context.Background())
	require.Error(t, err)
	require.Nil(t, items)
	assert.ErrorIs(t, err, errDiscoveryUnavailable)
}

func TestDiscoveryFetch_DoesNotForwardMasterKey(t *testing.T) {
	var authorization string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	t.Setenv("SHAHRAH_DISCOVERY_TOKEN", "")
	t.Setenv("MASTER_KEY", "must-not-leave-control-plane")

	_, status, _, err := discoveryFetch(context.Background(), server.URL)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)
	assert.Empty(t, authorization)
}

func TestDiscoveryFetch_ForwardsExplicitDiscoveryToken(t *testing.T) {
	var authorization string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization = r.Header.Get("Authorization")

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	t.Setenv("SHAHRAH_DISCOVERY_TOKEN", "explicit-discovery-service-token")
	t.Setenv("MASTER_KEY", "must-not-leave-control-plane")

	_, status, _, err := discoveryFetch(context.Background(), server.URL)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, status)
	assert.Equal(t, "Bearer explicit-discovery-service-token", authorization)
}

func TestDiscoveryProviderExists_NormalizesLegacyProviderSlug(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/providers" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"provider":"perplexity-ai","display_name":"Perplexity","enabled":true,"auth_modes":["api_key"],"supports_bring_your_own_key":true,"supports_virtual_keys":true}]}`))
	}))
	defer server.Close()

	t.Setenv("SHAHRAH_MODEL_DISCOVERY_URL", server.URL)
	invalidateDiscoveryCache()

	exists, err := discoveryProviderExists(context.Background(), "perplexity")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestMapDiscoveryProviders_PreservesModelCount(t *testing.T) {
	items := mapDiscoveryProviders([]discoveryProvider{{
		Slug:       "anthropic",
		Name:       "Anthropic",
		ModelCount: 12,
		Enabled:    true,
	}})

	require.Len(t, items, 1)
	assert.Equal(t, 12, items[0].ModelCount)
}

func TestDiscoveryProviderModels_NormalizesLegacyProviderSlug(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/providers/perplexity-ai/models":
			http.NotFound(w, r)
		case r.URL.Path == "/models" && r.URL.Query().Get("provider") == "perplexity-ai":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":[{"canonicalId":"sonar","name":"Sonar","mode":["chat"],"capabilities":["chat"],"inputModalities":["text"],"outputModalities":["text"]}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	t.Setenv("SHAHRAH_MODEL_DISCOVERY_URL", server.URL)
	invalidateDiscoveryCache()

	items, err := discoveryProviderModels(context.Background(), "perplexity")
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "perplexity-ai", items[0].Provider)
	assert.Equal(t, "sonar", items[0].Model)
}

func TestDecodeProvidersPayload_UsesDiscoveryShapeWhenCanonicalFieldsMissing(t *testing.T) {
	body := []byte(`[{"slug":"cerebras","name":"Cerebras","logoUrl":"/api/v1/providers/assets/provider-logos/cerebras.png","enabled":true,"supportsByok":true,"supportsRegionalRouting":true,"models":[{"canonicalId":"cerebras/llama","name":"Llama"}]}]`)

	items, err := decodeProvidersPayload(body)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "cerebras", items[0].Provider)
	assert.Equal(t, "Cerebras", items[0].DisplayName)
	assert.Equal(t, "/api/v1/providers/assets/provider-logos/cerebras.png", items[0].LogoURL)
}

func TestDecodeProvidersPayload_AllowsEmptyArrays(t *testing.T) {
	items, err := decodeProvidersPayload([]byte(`[]`))
	require.NoError(t, err)
	require.Empty(t, items)
}

func TestDiscoveryProviders_EmptyCatalogReturnsDiscoveryUnavailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/providers" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	t.Setenv("SHAHRAH_MODEL_DISCOVERY_URL", server.URL)
	invalidateDiscoveryCache()

	items, err := discoveryProviders(context.Background())
	require.Error(t, err)
	require.Nil(t, items)
	assert.ErrorIs(t, err, errDiscoveryUnavailable)
	assert.Contains(t, err.Error(), "empty provider catalog")
}

func TestNormalizeProviders_ResolvesRelativeLogoURLAgainstDiscoveryBase(t *testing.T) {
	items := normalizeProviders([]platform.ProviderCatalogProvider{
		{
			Provider:    "openai",
			DisplayName: "OpenAI",
			LogoURL:     "/api/v1/providers/assets/provider-logos/openai.png",
			Enabled:     true,
		},
	})

	require.Len(t, items, 1)
	assert.Equal(t, "/api/v1/providers/assets/provider-logos/openai.png", items[0].LogoURL)
}

func TestDecodeProviderModelsPayload_UsesDiscoveryShapeWhenCanonicalFieldsMissing(t *testing.T) {
	body := []byte(`{"data":[{"canonicalId":"cerebras/llama-4-scout","name":"Llama 4 Scout","mode":["chat"],"capabilities":["chat"],"inputModalities":["text"],"outputModalities":["text"]}]}`)

	items, err := decodeProviderModelsPayload(body)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "cerebras/llama-4-scout", items[0].Model)
	assert.Equal(t, "Llama 4 Scout", items[0].DisplayName)
	assert.Equal(t, []string{"chat"}, items[0].Modes)
	assert.Equal(t, []platform.ModelType{platform.ModelTypeChat}, items[0].Types)
}

func TestDiscoveryProviderModels_FallbackQueryUsesPaginationLimit100(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/providers/azure-ai-foundry/models":
			http.NotFound(w, r)
		case r.URL.Path == "/models" && r.URL.Query().Get("provider") == "azure-ai-foundry" && r.URL.Query().Get("limit") == "100" && r.URL.Query().Get("page") == "1":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":[{"canonicalId":"azure-ai-foundry/model-a","name":"Model A","mode":["chat"],"capabilities":["chat"],"inputModalities":["text"],"outputModalities":["text"]}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	t.Setenv("SHAHRAH_MODEL_DISCOVERY_URL", server.URL)
	invalidateDiscoveryCache()

	items, err := discoveryProviderModels(context.Background(), "azure-ai-foundry")
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "azure-ai-foundry", items[0].Provider)
	assert.Equal(t, "azure-ai-foundry/model-a", items[0].Model)
}
