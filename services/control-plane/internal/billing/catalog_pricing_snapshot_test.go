package billing

import (
	"context"
	"testing"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/shopspring/decimal"
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
				CanonicalID:    "google-gemini/gemini-3.1-pro-preview",
				Slug:           "gemini-3.1-pro-preview",
				Status:         "preview",
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
			Model: "google-gemini/gemini-3.1-pro-preview",
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
			Model:       "google-gemini/gemini-3.1-pro-preview",
			ServiceTier: "flex",
		},
	)
	require.NoError(t, err)
	require.Equal(t, "flex", flexEstimated.ServiceTier)
	require.Equal(t, "flex_input_tokens", flexEstimated.Components[0].Dimension)

	final, err := service.FinalizeFromUsage(
		context.Background(),
		"google-gemini/gemini-3.1-pro-preview",
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

func TestCatalogPricingModelStatusBillability(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		status   string
		billable bool
	}{
		{name: "active", status: "active", billable: true},
		{name: "preview", status: "preview", billable: true},
		{name: "normalized preview", status: " Preview ", billable: true},
		{name: "deprecated", status: "deprecated", billable: false},
		{name: "unavailable", status: "unavailable", billable: false},
		{name: "missing", status: "", billable: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.billable, catalogPricingModelStatusBillable(tt.status))
		})
	}
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

func TestOpenAIRateSelectorIsConditionalAndFailsOnAmbiguity(t *testing.T) {
	rates := []normalizedPricingRate{ //nolint:prealloc // A literal is clearer before the conditional test-only append below.
		{
			Dimension: "input_tokens", Unit: "token",
			UnitPriceUSD: decimal.RequireFromString("0.000001"),
		},
		{
			Dimension: "input_tokens", Unit: "token",
			UnitPriceUSD: decimal.RequireFromString("0.000002"),
			Condition:    map[string]interface{}{"service_tier": "priority"},
		},
	}

	selected, err := selectOpenAIRate(
		rates,
		"input_tokens",
		100,
		"priority",
		map[string]string{"service_tier": "priority"},
	)
	require.NoError(t, err)
	require.Equal(t, "0.000002", selected.UnitPriceUSD.String())

	rates = append(rates, normalizedPricingRate{
		Dimension: "input_tokens", Unit: "token",
		UnitPriceUSD: decimal.RequireFromString("0.000003"),
		Condition:    map[string]interface{}{"service_tier": "priority"},
	})
	_, err = selectOpenAIRate(
		rates,
		"input_tokens",
		100,
		"priority",
		map[string]string{"service_tier": "priority"},
	)
	require.ErrorIs(t, err, ErrPricingRateUnavailable)
}

func TestOpenAIGPT54NanoRejectsUnsupportedPriorityTier(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	profile := completeDiscoveryRuntimeProfile(
		now.Add(-time.Hour),
		"manual",
		[]string{"manual"},
	)
	profile.Profile.Endpoints = []string{"models", "chat_completions", "responses"}
	profile.Profile.RequestControls = platform.ModelRuntimeRequestControls{
		ServiceTierValues:  []string{"standard", "batch", "flex"},
		InferenceGeoValues: []string{"global"},
	}
	rates := []discoveryPricingRate{
		{Dimension: "input_tokens", Unit: "token", UnitPriceUSD: "0.0000002", Currency: "USD", Region: "global", SourceURL: "https://openai.com/api/pricing/"},
		{Dimension: "output_tokens", Unit: "token", UnitPriceUSD: "0.00000125", Currency: "USD", Region: "global", SourceURL: "https://openai.com/api/pricing/"},
		{Dimension: "batch_input_tokens", Unit: "token", UnitPriceUSD: "0.0000001", Currency: "USD", Region: "global", SourceURL: "https://openai.com/api/pricing/", Condition: map[string]interface{}{"service_tier": "batch"}},
		{Dimension: "batch_output_tokens", Unit: "token", UnitPriceUSD: "0.000000625", Currency: "USD", Region: "global", SourceURL: "https://openai.com/api/pricing/", Condition: map[string]interface{}{"service_tier": "batch"}},
		{Dimension: "flex_input_tokens", Unit: "token", UnitPriceUSD: "0.0000001", Currency: "USD", Region: "global", SourceURL: "https://openai.com/api/pricing/", Condition: map[string]interface{}{"service_tier": "flex"}},
		{Dimension: "flex_output_tokens", Unit: "token", UnitPriceUSD: "0.000000625", Currency: "USD", Region: "global", SourceURL: "https://openai.com/api/pricing/", Condition: map[string]interface{}{"service_tier": "flex"}},
	}
	snapshot, err := normalizeCatalogPricingSnapshot(
		discoveryRuntimeSnapshot{
			Provider: "openai", Revision: "openai-gpt-5.4-nano",
			GeneratedAt: now,
			Models: []discoveryRuntimeModel{{
				CanonicalID: "openai/gpt-5.4-nano", Slug: "gpt-5.4-nano",
				Status: "active", RuntimeProfile: profile, PricingRates: rates,
			}},
		},
		"openai",
	)
	require.NoError(t, err)

	manager := &CatalogPricingSnapshotManager{
		provider: "openai",
		now:      func() time.Time { return now },
	}
	manager.state.Store(snapshot)
	service := &Service{
		defaultRate: platform.ExchangeRate{USDtoIRR: 1},
		catalogPricing: map[string]*CatalogPricingSnapshotManager{
			"openai": manager,
		},
	}
	maxOutputTokens := 32
	_, err = service.EstimateCost(
		context.Background(),
		contractsopenai.ChatCompletionRequest{
			Model: "openai:openai-calm-oasis:gpt-5.4-nano",
			Messages: []contractsopenai.ChatMessage{{
				Role: "user", Content: "Reply with exactly OPENAI_TIER_OK.",
			}},
			ServiceTier: "priority",
			MaxTokens:   &maxOutputTokens,
		},
	)
	require.ErrorIs(t, err, ErrUnsupportedServiceTier)
	require.NotErrorIs(t, err, ErrPricingUnavailable)
	require.NotErrorIs(t, err, ErrPricingRateUnavailable)

	var tierErr *UnsupportedServiceTierError
	require.ErrorAs(t, err, &tierErr)
	require.Equal(t, "gpt-5.4-nano", tierErr.Model)
	require.Equal(t, "priority", tierErr.Tier)
}

func TestOpenAIUsagePricingMeasuresCacheAndAudioSeparately(t *testing.T) {
	rates := []normalizedPricingRate{
		{Dimension: "input_tokens", Unit: "token", UnitPriceUSD: decimal.RequireFromString("0.000001")},
		{Dimension: "output_tokens", Unit: "token", UnitPriceUSD: decimal.RequireFromString("0.000002")},
		{Dimension: "cache_read_tokens", Unit: "token", UnitPriceUSD: decimal.RequireFromString("0.0000005")},
		{Dimension: "audio_input_tokens", Unit: "token", UnitPriceUSD: decimal.RequireFromString("0.00001")},
		{Dimension: "audio_output_tokens", Unit: "token", UnitPriceUSD: decimal.RequireFromString("0.00002")},
	}
	breakdown, err := openAIUsageCostBreakdown(
		"openai:account:gpt-audio",
		&contractsopenai.Usage{
			PromptTokens: 100, CompletionTokens: 20, TotalTokens: 120,
			PromptTokensDetails: &contractsopenai.UsageTokenDetails{
				CachedTokens: 30,
				AudioTokens:  10,
			},
			CompletionTokensDetails: &contractsopenai.UsageTokenDetails{
				AudioTokens: 5,
			},
		},
		priceInfo{
			Provider: "openai", Revision: "openai-1",
			GeneratedAt: time.Now(), RateList: rates,
		},
		1,
		nil,
	)
	require.NoError(t, err)
	require.Equal(t, 30, breakdown.CacheReadInputTokens)
	require.Len(t, breakdown.Components, 5)
	require.Equal(t, "openai-1", breakdown.PricingRevision)
}

func TestOpenAIUsagePricingUses272KBoundaryAndTierSpecificLongRates(t *testing.T) {
	rates := []normalizedPricingRate{
		{Dimension: "input_tokens", Unit: "token", UnitPriceUSD: decimal.RequireFromString("0.000005")},
		{Dimension: "output_tokens", Unit: "token", UnitPriceUSD: decimal.RequireFromString("0.000030")},
		{Dimension: "long_context_input_tokens", Unit: "token", UnitPriceUSD: decimal.RequireFromString("0.000010"), Condition: map[string]interface{}{"min_prompt_tokens": 272001}},
		{Dimension: "long_context_output_tokens", Unit: "token", UnitPriceUSD: decimal.RequireFromString("0.000045"), Condition: map[string]interface{}{"min_prompt_tokens": 272001}},
		{Dimension: "batch_long_context_input_tokens", Unit: "token", UnitPriceUSD: decimal.RequireFromString("0.000005"), Condition: map[string]interface{}{"service_tier": "batch", "min_prompt_tokens": 272001}},
		{Dimension: "batch_long_context_output_tokens", Unit: "token", UnitPriceUSD: decimal.RequireFromString("0.0000225"), Condition: map[string]interface{}{"service_tier": "batch", "min_prompt_tokens": 272001}},
	}
	price := priceInfo{Provider: "openai", Revision: "openai-long", GeneratedAt: time.Now(), RateList: rates}

	standard, err := openAIUsageCostBreakdown(
		"openai/gpt-5.5",
		&contractsopenai.Usage{PromptTokens: 272001, CompletionTokens: 1, TotalTokens: 272002},
		price,
		1,
		nil,
	)
	require.NoError(t, err)
	require.Equal(t, "long_context_input_tokens", standard.Components[0].Dimension)

	batch, err := openAIUsageCostBreakdown(
		"openai/gpt-5.5",
		&contractsopenai.Usage{PromptTokens: 272001, CompletionTokens: 1, TotalTokens: 272002, ServiceTier: "batch"},
		price,
		1,
		map[string]string{"service_tier": "batch"},
	)
	require.NoError(t, err)
	require.Equal(t, "batch_long_context_input_tokens", batch.Components[0].Dimension)
}

func TestOpenAIFastAliasAndUndiscountedCachedInput(t *testing.T) {
	rates := []normalizedPricingRate{
		{Dimension: "speed_input_tokens", Unit: "token", UnitPriceUSD: decimal.RequireFromString("0.000030"), Condition: map[string]interface{}{"service_tier": "priority"}},
		{Dimension: "speed_output_tokens", Unit: "token", UnitPriceUSD: decimal.RequireFromString("0.000180"), Condition: map[string]interface{}{"service_tier": "priority"}},
	}
	breakdown, err := openAIUsageCostBreakdown(
		"openai/gpt-5.5-pro",
		&contractsopenai.Usage{
			PromptTokens: 100, CompletionTokens: 1, TotalTokens: 101, ServiceTier: "fast",
			PromptTokensDetails: &contractsopenai.UsageTokenDetails{CachedTokens: 40},
		},
		priceInfo{Provider: "openai", Revision: "openai-fast", GeneratedAt: time.Now(), RateList: rates},
		1,
		map[string]string{"service_tier": "fast"},
	)
	require.NoError(t, err)
	require.Equal(t, "priority", breakdown.ServiceTier)
	require.Len(t, breakdown.Components, 2)
	require.Equal(t, "speed_input_tokens", breakdown.Components[0].Dimension)
	require.Equal(t, "100", breakdown.Components[0].Quantity)

	_, err = openAIUsageCostBreakdown(
		"openai/gpt-5.5",
		&contractsopenai.Usage{PromptTokens: 272001, CompletionTokens: 1, ServiceTier: "fast"},
		priceInfo{Provider: "openai", Revision: "openai-fast", GeneratedAt: time.Now(), RateList: rates},
		1,
		nil,
	)
	require.ErrorIs(t, err, ErrUnsupportedServiceTier)
}

func TestApplyPricingOverridePinsPrivateRevisionAndRecalculates(t *testing.T) {
	service := &Service{}
	estimated := platform.UsageCostBreakdown{
		ExchangeRate:    2,
		PricingRevision: "openai-public-1",
		Components: []platform.UsageCostComponent{{
			Dimension: "input_tokens", Quantity: "100", Unit: "token",
			UnitPriceUSD: "0.000001", CostUSD: "0.0001",
		}},
		PricingSnapshot: &platform.UsagePricingSnapshot{
			Provider: "openai", Revision: "openai-public-1", GeneratedAt: time.Now(),
			Rates: map[string]platform.UsageCostComponent{
				"input_tokens": {
					Dimension: "input_tokens", Unit: "token",
					UnitPriceUSD: "0.000001",
				},
			},
			RateList: []platform.UsageCostComponent{{
				Dimension: "input_tokens", Unit: "token",
				UnitPriceUSD: "0.000001",
			}},
		},
	}
	actual, err := service.ApplyPricingOverride(
		context.Background(),
		estimated,
		&platform.ModelPrice{Rates: []platform.ModelPricingRate{{
			Key: "input", Dimension: "input_tokens", Unit: "token",
			UnitPriceUSD: "0.000002",
		}}},
	)
	require.NoError(t, err)
	require.Equal(t, 0.0002, actual.TotalCostUSD)
	require.Equal(t, 0.0004, actual.TotalCostIRR)
	require.Contains(t, actual.PricingRevision, "openai-public-1:private:")
	require.Equal(t, "0.000002", actual.Components[0].UnitPriceUSD)
}
