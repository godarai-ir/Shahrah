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

var anthropicReadyModels = []string{
	"claude-haiku-4-5",
	"claude-haiku-4-5-20251001",
	"claude-opus-4-1",
	"claude-opus-4-1-20250805",
	"claude-opus-4-5",
	"claude-opus-4-5-20251101",
	"claude-opus-4-6",
	"claude-opus-4-7",
	"claude-sonnet-4-5",
	"claude-sonnet-4-5-20250929",
	"claude-sonnet-4-6",
}

func TestAnthropicPricingSnapshotRequiresEveryCoreDimensionForReadyModels(
	t *testing.T,
) {
	now := time.Now().UTC()

	payload := discoveryRuntimeSnapshot{
		Provider:    "anthropic",
		Revision:    "revision-1",
		GeneratedAt: now,
	}
	for _, model := range anthropicReadyModels {
		payload.Models = append(payload.Models, discoveryRuntimeModel{
			CanonicalID: "anthropic/" + model,
			Slug:        model,
			Status:      "active",
			RuntimeProfile: completeDiscoveryRuntimeProfile(
				now,
				"manual",
				[]string{"manual", "disabled"},
			),
			PricingRates: completeAnthropicRates(),
		})
	}

	snapshot, err := normalizeAnthropicPricingSnapshot(payload)
	require.NoError(t, err)
	require.Len(t, snapshot.Models, len(anthropicReadyModels)*2)

	for _, model := range anthropicReadyModels {
		rates := snapshot.Models[model].Rates
		require.Contains(t, rates, "input_tokens")
		require.Contains(t, rates, "output_tokens")
		require.Contains(t, rates, "cache_write_5m_tokens")
		require.Contains(t, rates, "cache_write_1h_tokens")
		require.Contains(t, rates, "cache_read_tokens")
		require.Contains(t, rates, "batch_input_tokens")
		require.Contains(t, rates, "batch_output_tokens")
		require.Contains(t, rates, "web_search_requests")
		require.Contains(t, rates, "code_execution_container_hours")
	}
}

func TestAnthropicPricingSnapshotFailsClosedForMissingOrStaleRates(
	t *testing.T,
) {
	now := time.Now().UTC()
	payload := discoveryRuntimeSnapshot{
		Provider:    "anthropic",
		Revision:    "revision-1",
		GeneratedAt: now,
		Models: []discoveryRuntimeModel{{
			CanonicalID: "anthropic/claude-sonnet-4-6",
			Slug:        "claude-sonnet-4-6",
			RuntimeProfile: completeDiscoveryRuntimeProfile(
				now,
				"adaptive",
				[]string{"adaptive", "manual", "disabled"},
			),
			PricingRates: []discoveryPricingRate{{
				Dimension:    "input_tokens",
				Unit:         "token",
				UnitPriceUSD: "0.000003",
				Currency:     "USD",
				Region:       "global",
				SourceURL:    "https://platform.claude.com/docs/en/about-claude/pricing",
			}},
		}},
	}
	_, err := normalizeAnthropicPricingSnapshot(payload)
	require.ErrorIs(t, err, ErrPricingUnavailable)

	manager := &AnthropicPricingSnapshotManager{now: time.Now}
	manager.state.Store(&anthropicPricingSnapshot{
		Revision:    "stale",
		GeneratedAt: time.Now().Add(-anthropicPricingMaxAge - time.Second),
		Models:      map[string]anthropicPricingModel{},
	})
	_, _, err = manager.price("claude-sonnet-4-6")
	require.ErrorIs(t, err, ErrPricingUnavailable)
}

func TestAnthropicPricingSnapshotFailsClosedWithoutValidRuntimeProfile(
	t *testing.T,
) {
	payload := discoveryRuntimeSnapshot{
		Provider:    "anthropic",
		Revision:    "revision-profile-required",
		GeneratedAt: time.Now().UTC(),
		Models: []discoveryRuntimeModel{{
			CanonicalID:  "anthropic/claude-opus-4-7",
			Slug:         "claude-opus-4-7",
			Status:       "active",
			PricingRates: completeAnthropicRates(),
		}},
	}

	_, err := normalizeAnthropicPricingSnapshot(payload)
	require.ErrorIs(t, err, ErrPricingUnavailable)
}

func TestAnthropicEstimatePinsRuntimeProfileToPricingRevision(t *testing.T) {
	now := time.Now().UTC()
	snapshot, err := normalizeAnthropicPricingSnapshot(
		discoveryRuntimeSnapshot{
			Provider:    "anthropic",
			Revision:    "revision-profile-1",
			GeneratedAt: now,
			Models: []discoveryRuntimeModel{{
				CanonicalID: "anthropic/claude-opus-4-7",
				Slug:        "claude-opus-4-7",
				Status:      "active",
				RuntimeProfile: completeDiscoveryRuntimeProfile(
					now,
					"adaptive",
					[]string{"adaptive"},
				),
				PricingRates: completeAnthropicRates(),
			}},
		},
	)
	require.NoError(t, err)

	manager := &AnthropicPricingSnapshotManager{now: time.Now}
	manager.state.Store(snapshot)
	service := &Service{
		PricingEngine: &PricingEngine{
			defaultRate:      platform.ExchangeRate{USDtoIRR: 1},
			anthropicPricing: manager,
		},
	}

	estimated, err := service.EstimateCost(
		context.Background(),
		contractsopenai.ChatCompletionRequest{
			Model: "anthropic/claude-opus-4-7",
		},
	)

	require.NoError(t, err)
	require.NotNil(t, estimated.ModelRuntimeProfile)
	require.Equal(
		t,
		estimated.PricingRevision,
		estimated.ModelRuntimeProfile.CatalogRevision,
	)
	require.Equal(
		t,
		"adaptive",
		estimated.ModelRuntimeProfile.Profile.Thinking.DefaultMode,
	)
}

func TestAnthropicFastModeUsesPremiumRatesForEstimateAndUsage(t *testing.T) {
	service, _ := anthropicBillingTestService(t)

	estimated, err := service.EstimateCost(
		context.Background(),
		contractsopenai.ChatCompletionRequest{
			Model: "anthropic/claude-opus-4-7",
			Speed: "fast",
		},
	)
	require.NoError(t, err)
	require.Equal(t, "speed_input_tokens", estimated.Components[0].Dimension)
	require.Equal(t, "fast", estimated.PricingConditions["speed"])

	final, err := service.FinalizeFromUsage(
		context.Background(),
		"anthropic/claude-opus-4-7",
		&contractsopenai.Usage{
			PromptTokens:     100,
			CompletionTokens: 10,
			TotalTokens:      110,
			Speed:            "fast",
		},
		estimated,
	)
	require.NoError(t, err)
	require.Equal(t, "speed_input_tokens", final.Components[0].Dimension)
	require.Equal(t, "speed_output_tokens", final.Components[1].Dimension)

	combined, err := service.EstimateCost(
		context.Background(),
		contractsopenai.ChatCompletionRequest{
			Model:        "anthropic/claude-opus-4-7",
			Speed:        "fast",
			InferenceGeo: "us",
		},
	)
	require.NoError(t, err)
	require.Equal(
		t,
		"data_residency_speed_input_tokens",
		combined.Components[0].Dimension,
	)
}

func TestAnthropicFinalizeUsesPinnedDecimalSnapshotAndCacheDimensions(
	t *testing.T,
) {
	now := time.Now().UTC()
	snapshot, err := normalizeAnthropicPricingSnapshot(
		discoveryRuntimeSnapshot{
			Provider:    "anthropic",
			Revision:    "revision-1",
			GeneratedAt: now,
			Models: []discoveryRuntimeModel{{
				CanonicalID: "anthropic/claude-sonnet-4-6",
				Slug:        "claude-sonnet-4-6",
				Status:      "active",
				RuntimeProfile: completeDiscoveryRuntimeProfile(
					now,
					"adaptive",
					[]string{"adaptive", "manual", "disabled"},
				),
				PricingRates: completeAnthropicRates(),
			}},
		},
	)
	require.NoError(t, err)

	manager := &AnthropicPricingSnapshotManager{now: time.Now}
	manager.state.Store(snapshot)
	service := &Service{
		PricingEngine: &PricingEngine{
			defaultRate:      platform.ExchangeRate{USDtoIRR: 1},
			anthropicPricing: manager,
		},
	}

	estimated, err := service.EstimateCost(
		context.Background(),
		contractsopenai.ChatCompletionRequest{
			Model: "anthropic/claude-sonnet-4-6",
		},
	)
	require.NoError(t, err)
	require.Equal(t, "revision-1", estimated.PricingRevision)

	// A catalog update during the request must not change finalization.
	newSnapshot := *snapshot
	newSnapshot.Revision = "revision-2"
	newSnapshot.Models = map[string]anthropicPricingModel{}
	manager.state.Store(&newSnapshot)

	final, err := service.FinalizeFromUsage(
		context.Background(),
		"anthropic/claude-sonnet-4-6",
		&contractsopenai.Usage{
			PromptTokens:               1000,
			CompletionTokens:           100,
			TotalTokens:                1130,
			CacheCreation5mInputTokens: 10,
			CacheReadInputTokens:       20,
		},
		estimated,
	)
	require.NoError(t, err)
	require.Equal(t, "revision-1", final.PricingRevision)
	require.True(
		t,
		decimal.NewFromFloat(final.TotalCostUSD).Equal(
			decimal.RequireFromString("0.0045435"),
		),
	)
	require.Len(t, final.Components, 4)
}

func TestAnthropicFinalizeAuditsFreeCodeExecutionWithoutHourlyConversion(
	t *testing.T,
) {
	service, estimated := anthropicBillingTestService(t)
	estimated.PricingConditions = map[string]string{
		platform.AnthropicCodeExecutionPricingCondition: platform.AnthropicCodeExecutionFreeWithCurrentWebTools,
	}

	final, err := service.FinalizeFromUsage(
		context.Background(),
		"anthropic/claude-opus-4-7",
		&contractsopenai.Usage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
			ServerToolUse: map[string]int{
				"code_execution_requests": 2,
			},
		},
		estimated,
	)

	require.NoError(t, err)
	require.Equal(t, map[string]int{"code_execution_requests": 2}, final.ServerToolUse)
	require.Equal(t, "revision-1", final.PricingRevision)
	require.Len(t, final.Components, 3)
	require.Equal(t, "code_execution_requests", final.Components[0].Dimension)
	require.Equal(t, "2", final.Components[0].Quantity)
	require.Equal(t, "request", final.Components[0].Unit)
	require.Equal(t, "0", final.Components[0].UnitPriceUSD)
	require.Equal(t, "0", final.Components[0].CostUSD)
	require.Equal(
		t,
		platform.AnthropicCodeExecutionFreeWithCurrentWebTools,
		final.Components[0].Condition[platform.AnthropicCodeExecutionPricingCondition],
	)
}

func TestAnthropicFinalizeFailsClosedForUnattributedCodeExecution(t *testing.T) {
	service, estimated := anthropicBillingTestService(t)

	_, err := service.FinalizeFromUsage(
		context.Background(),
		"anthropic/claude-opus-4-7",
		&contractsopenai.Usage{
			ServerToolUse: map[string]int{
				"code_execution_requests": 1,
			},
		},
		estimated,
	)

	require.ErrorIs(t, err, ErrPricingUnavailable)
	require.Contains(t, err.Error(), "code execution")
}

func TestAnthropicFinalizeFailsClosedWhenFreeCodeExecutionSourceRateIsMissing(
	t *testing.T,
) {
	service, estimated := anthropicBillingTestService(t)
	delete(
		estimated.PricingSnapshot.Rates,
		"code_execution_container_hours",
	)
	estimated.PricingConditions = map[string]string{
		platform.AnthropicCodeExecutionPricingCondition: platform.AnthropicCodeExecutionFreeWithCurrentWebTools,
	}

	_, err := service.FinalizeFromUsage(
		context.Background(),
		"anthropic/claude-opus-4-7",
		&contractsopenai.Usage{
			ServerToolUse: map[string]int{
				"code_execution_requests": 1,
			},
		},
		estimated,
	)

	require.ErrorIs(t, err, ErrPricingUnavailable)
	require.Contains(t, err.Error(), "code execution")
}

func anthropicBillingTestService(
	t *testing.T,
) (*Service, platform.UsageCostBreakdown) {
	t.Helper()

	snapshot, err := normalizeAnthropicPricingSnapshot(
		discoveryRuntimeSnapshot{
			Provider:    "anthropic",
			Revision:    "revision-1",
			GeneratedAt: time.Now().UTC(),
			Models: []discoveryRuntimeModel{{
				CanonicalID: "anthropic/claude-opus-4-7",
				Slug:        "claude-opus-4-7",
				Status:      "active",
				RuntimeProfile: completeDiscoveryRuntimeProfile(
					time.Now().UTC(),
					"adaptive",
					[]string{"adaptive"},
				),
				PricingRates: completeAnthropicRates(),
			}},
		},
	)
	require.NoError(t, err)

	manager := &AnthropicPricingSnapshotManager{now: time.Now}
	manager.state.Store(snapshot)
	service := &Service{
		PricingEngine: &PricingEngine{
			defaultRate:      platform.ExchangeRate{USDtoIRR: 1},
			anthropicPricing: manager,
		},
	}
	estimated, err := service.EstimateCost(
		context.Background(),
		contractsopenai.ChatCompletionRequest{
			Model: "anthropic/claude-opus-4-7",
		},
	)
	require.NoError(t, err)

	return service, estimated
}

func completeAnthropicRates() []discoveryPricingRate {
	source := "https://platform.claude.com/docs/en/about-claude/pricing"
	values := map[string]string{
		"input_tokens":                       "0.000003",
		"output_tokens":                      "0.000015",
		"cache_write_5m_tokens":              "0.00000375",
		"cache_write_1h_tokens":              "0.000006",
		"cache_read_tokens":                  "0.0000003",
		"batch_input_tokens":                 "0.0000015",
		"batch_output_tokens":                "0.0000075",
		"speed_input_tokens":                 "0.000010",
		"speed_output_tokens":                "0.000050",
		"speed_cache_write_5m_tokens":        "0.0000125",
		"speed_cache_write_1h_tokens":        "0.000020",
		"speed_cache_read_tokens":            "0.000001",
		"data_residency_speed_input_tokens":  "0.000011",
		"data_residency_speed_output_tokens": "0.000055",
		"web_search_requests":                "0.01",
		"code_execution_container_hours":     "0.05",
	}

	out := make([]discoveryPricingRate, 0, len(values))
	for dimension, price := range values {
		unit := "token"
		if dimension == "web_search_requests" {
			unit = "request"
		}

		if dimension == "code_execution_container_hours" {
			unit = "container_hour"
		}

		out = append(out, discoveryPricingRate{
			Dimension:    dimension,
			Unit:         unit,
			UnitPriceUSD: price,
			Currency:     "USD",
			Region:       "global",
			SourceURL:    source,
		})
	}

	return out
}

func completeDiscoveryRuntimeProfile(
	effectiveFrom time.Time,
	defaultThinking string,
	allowedThinking []string,
) *discoveryRuntimeProfileRevision {
	return &discoveryRuntimeProfileRevision{
		Version: 1,
		Profile: platform.ModelRuntimeProfile{
			Version:   1,
			Endpoints: []string{"messages", "count_tokens", "message_batches"},
			Thinking: platform.ModelRuntimeThinkingProfile{
				DefaultMode:  defaultThinking,
				AllowedModes: allowedThinking,
				EffortValues: []string{"low", "medium", "high", "max"},
			},
			AssistantPrefill: false,
			Sampling: platform.ModelRuntimeSamplingProfile{
				Parameters:             []string{"temperature", "top_p", "top_k"},
				DisallowedWithThinking: []string{"temperature", "top_p", "top_k"},
			},
			Tools: platform.ModelRuntimeToolsProfile{
				Custom: true, Parallel: true, StructuredOutput: true,
			},
			Modalities: platform.ModelRuntimeModalities{
				Input: []string{"text", "image", "pdf"}, Output: []string{"text"},
			},
			RequestControls: platform.ModelRuntimeRequestControls{
				ServiceTierValues:  []string{"auto", "standard_only"},
				InferenceGeoValues: []string{"global", "us"},
				SpeedValues:        []string{"standard"},
			},
			BetaHeaders: []string{"files-api-2025-04-14"},
			Limits: platform.ModelRuntimeLimits{
				ContextWindow: 1_000_000, MaxOutputTokens: 128_000,
				MessagesBytes: 32 * 1024 * 1024,
				BatchBytes:    256 * 1024 * 1024,
				FileBytes:     500 * 1024 * 1024,
			},
		},
		SourceURL:     "https://platform.claude.com/docs/en/about-claude/models/overview",
		EffectiveFrom: effectiveFrom,
	}
}
