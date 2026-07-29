package billing

import (
	"context"
	"testing"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/stretchr/testify/require"
)

func TestCatalogPricingSnapshotSupportsGoogleGeminiWithoutStaticRegistry(t *testing.T) {
	now := time.Date(2026, 7, 29, 0, 0, 0, 0, time.UTC)
	profile := completeDiscoveryRuntimeProfile(
		now.Add(-time.Hour),
		"manual",
		[]string{"manual"},
	)
	profile.Profile.Endpoints = []string{"models", "chat_completions"}
	profile.Profile.Modalities = platform.ModelRuntimeModalities{
		Input:  []string{"text", "image", "audio", "video"},
		Output: []string{"text"},
	}
	profile.Profile.Thinking.EffortValues = []string{
		"none", "minimal", "low", "medium", "high",
	}
	profile.Profile.RequestControls = platform.ModelRuntimeRequestControls{
		ServiceTierValues:  []string{"standard", "batch", "flex", "priority"},
		InferenceGeoValues: []string{"global"},
	}
	rates := []discoveryPricingRate{
		{
			Dimension: "input_tokens", Unit: "token",
			UnitPriceUSD: "0.00000025", Currency: "USD", Region: "global",
			SourceURL: "https://ai.google.dev/gemini-api/docs/pricing",
		},
		{
			Dimension: "output_tokens", Unit: "token",
			UnitPriceUSD: "0.0000015", Currency: "USD", Region: "global",
			SourceURL: "https://ai.google.dev/gemini-api/docs/pricing",
		},
		{
			Dimension: "batch_input_tokens", Unit: "token",
			UnitPriceUSD: "0.000000125", Currency: "USD", Region: "global",
			Condition: map[string]interface{}{"service_tier": "batch"},
			SourceURL: "https://ai.google.dev/gemini-api/docs/pricing",
		},
		{
			Dimension: "flex_input_tokens", Unit: "token",
			UnitPriceUSD: "0.000000125", Currency: "USD", Region: "global",
			Condition: map[string]interface{}{"service_tier": "flex"},
			SourceURL: "https://ai.google.dev/gemini-api/docs/pricing",
		},
		{
			Dimension: "flex_output_tokens", Unit: "token",
			UnitPriceUSD: "0.00000075", Currency: "USD", Region: "global",
			Condition: map[string]interface{}{"service_tier": "flex"},
			SourceURL: "https://ai.google.dev/gemini-api/docs/pricing",
		},
		{
			Dimension: "cache_read_tokens", Unit: "token",
			UnitPriceUSD: "0.000000025", Currency: "USD", Region: "global",
			SourceURL: "https://ai.google.dev/gemini-api/docs/pricing",
		},
	}
	snapshot, err := normalizeCatalogPricingSnapshot(
		discoveryRuntimeSnapshot{
			Provider: "google-gemini", Revision: "gemini-revision-1",
			GeneratedAt: now,
			Models: []discoveryRuntimeModel{{
				CanonicalID:    "google-gemini/gemini-3.1-flash-lite",
				Slug:           "gemini-3.1-flash-lite",
				Status:         "active",
				RuntimeProfile: profile,
				PricingRates:   rates,
			}},
		},
		"google-gemini",
	)
	require.NoError(t, err)

	manager := &CatalogPricingSnapshotManager{
		provider: "google-gemini",
		now:      func() time.Time { return now },
	}
	manager.state.Store(snapshot)
	service := &Service{
		defaultRate: platform.ExchangeRate{USDtoIRR: 1},
		catalogPricing: map[string]*CatalogPricingSnapshotManager{
			"google-gemini": manager,
		},
	}

	estimated, err := service.EstimateCost(
		context.Background(),
		contractsopenai.ChatCompletionRequest{
			Model: "google-gemini/gemini-3.1-flash-lite",
		},
	)
	require.NoError(t, err)
	require.Equal(t, "google-gemini", estimated.PricingSnapshot.Provider)
	require.Equal(t, "gemini-revision-1", estimated.PricingRevision)
	require.NotNil(t, estimated.ModelRuntimeProfile)
	require.Contains(t, estimated.PricingSnapshot.Rates, "batch_input_tokens")

	flexEstimated, err := service.EstimateCost(
		context.Background(),
		contractsopenai.ChatCompletionRequest{
			Model:       "google-gemini/gemini-3.1-flash-lite",
			ServiceTier: "flex",
		},
	)
	require.NoError(t, err)
	require.Equal(t, "flex", flexEstimated.ServiceTier)
	require.Equal(t, "flex_input_tokens", flexEstimated.Components[0].Dimension)

	final, err := service.FinalizeFromUsage(
		context.Background(),
		"google-gemini/gemini-3.1-flash-lite",
		&contractsopenai.Usage{
			PromptTokens:     100,
			CompletionTokens: 20,
			TotalTokens:      120,
			ServiceTier:      "standard",
			PromptTokensDetails: &contractsopenai.UsageTokenDetails{
				CachedTokens: 40,
			},
		},
		estimated,
	)
	require.NoError(t, err)
	require.Equal(t, "standard", final.ServiceTier)
	require.Equal(t, 40, final.CacheReadInputTokens)
	require.Equal(t, []string{
		"cache_read_tokens",
		"input_tokens",
		"output_tokens",
	}, []string{
		final.Components[0].Dimension,
		final.Components[1].Dimension,
		final.Components[2].Dimension,
	})
}

func TestCatalogPricingSnapshotRejectsDuplicateDimensions(t *testing.T) {
	now := time.Date(2026, 7, 29, 0, 0, 0, 0, time.UTC)
	profile := completeDiscoveryRuntimeProfile(
		now.Add(-time.Hour),
		"manual",
		[]string{"manual"},
	)
	profile.Profile.Endpoints = []string{"models", "chat_completions"}
	profile.Profile.RequestControls.InferenceGeoValues = []string{"global"}
	duplicate := discoveryPricingRate{
		Dimension: "input_tokens", Unit: "token",
		UnitPriceUSD: "0.000001", Currency: "USD", Region: "global",
		SourceURL: "https://ai.google.dev/gemini-api/docs/pricing",
	}
	_, err := normalizeCatalogPricingSnapshot(
		discoveryRuntimeSnapshot{
			Provider: "google-gemini", Revision: "revision",
			GeneratedAt: now,
			Models: []discoveryRuntimeModel{{
				CanonicalID: "google-gemini/test", Slug: "test",
				RuntimeProfile: profile,
				PricingRates:   []discoveryPricingRate{duplicate, duplicate},
			}},
		},
		"google-gemini",
	)
	require.ErrorIs(t, err, ErrPricingUnavailable)
}
