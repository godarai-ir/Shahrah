package proxy

import (
	"testing"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/stretchr/testify/require"
)

func TestAnthropicChatValidationUsesPinnedRuntimeProfile(t *testing.T) {
	binding := testAnthropicRuntimeProfileBinding("claude-future-model")
	binding.Profile.Thinking.DefaultMode = "adaptive"
	binding.Profile.Thinking.AllowedModes = []string{"adaptive"}
	binding.Profile.AssistantPrefill = false
	binding.Profile.RequestControls.SpeedValues = []string{"fast"}
	request := contractsopenai.ChatCompletionRequest{
		Model:           "claude-future-model",
		ReasoningEffort: "high",
		Speed:           "fast",
		Messages: []contractsopenai.ChatMessage{{
			Role: "user", Content: "test",
		}},
	}

	require.NoError(
		t,
		validateAnthropicChatRuntimeProfile(request, &binding),
	)

	request.Messages = []contractsopenai.ChatMessage{{
		Role: "assistant", Content: "prefill",
	}}
	require.ErrorContains(
		t,
		validateAnthropicChatRuntimeProfile(request, &binding),
		"assistant",
	)

	request.Messages = []contractsopenai.ChatMessage{{
		Role: "user", Content: "test",
	}}
	topK := 10
	request.TopK = &topK
	binding.Profile.Sampling.Parameters = []string{"temperature", "top_p"}
	require.ErrorContains(
		t,
		validateAnthropicChatRuntimeProfile(request, &binding),
		"top_k",
	)
}

func TestValidateThinkingConfigAcceptsAdaptiveForProfileValidation(
	t *testing.T,
) {
	require.NoError(
		t,
		validateThinkingConfig(
			&contractsopenai.ThinkingConfig{Type: "adaptive"},
		),
	)
}

func TestAnthropicOpenAIToolPricingPolicyPinsFreeCodeExecution(
	t *testing.T,
) {
	request := contractsopenai.ChatCompletionRequest{
		Tools: []contractsopenai.ToolDefinition{
			{Type: "web_search_20260209", Name: "web_search"},
			{Type: "code_execution_20260521", Name: "code_execution"},
		},
	}

	conditions, err := anthropicOpenAIToolPricingPolicy(request)

	require.NoError(t, err)
	require.Equal(
		t,
		platform.AnthropicCodeExecutionFreeWithCurrentWebTools,
		conditions[platform.AnthropicCodeExecutionPricingCondition],
	)
}

func TestAnthropicOpenAIToolPricingPolicyRejectsBillableCodeExecution(
	t *testing.T,
) {
	request := contractsopenai.ChatCompletionRequest{
		Tools: []contractsopenai.ToolDefinition{{
			Type: "code_execution_20260521", Name: "code_execution",
		}},
	}

	_, err := anthropicOpenAIToolPricingPolicy(request)

	require.ErrorContains(t, err, "billable")
}

func TestAnthropicEffectiveSpeedMetadataRecordsFallback(t *testing.T) {
	metadata := withAnthropicEffectiveSpeedMetadata(
		map[string]interface{}{"caller": "test"},
		"fast",
		"standard",
	)

	require.Equal(t, "fast", metadata["requested_speed"])
	require.Equal(t, "standard", metadata["effective_speed"])
	require.Equal(t, "test", metadata["caller"])
}

func TestRuntimeProfilesFromEstimateRequiresSamePinnedRevision(t *testing.T) {
	binding := testAnthropicRuntimeProfileBinding("claude-opus-4-7")
	estimated := platform.UsageCostBreakdown{
		PricingRevision:     "different-revision",
		ModelRuntimeProfile: &binding,
	}

	_, err := runtimeProfilesFromEstimate(
		"anthropic",
		"claude-opus-4-7",
		estimated,
	)

	require.ErrorContains(t, err, "invalid")
}

func TestRuntimeProfilesFromEstimateAcceptsGoogleGeminiAlias(t *testing.T) {
	binding := testGoogleGeminiRuntimeProfileBinding(
		"gemini-3.1-pro-preview",
		[]string{"chat_completions"},
	)
	estimated := platform.UsageCostBreakdown{
		PricingRevision:     binding.CatalogRevision,
		ModelRuntimeProfile: &binding,
	}

	profiles, err := runtimeProfilesFromEstimate(
		"gemini",
		"gemini-3.1-pro-preview",
		estimated,
	)

	require.NoError(t, err)
	require.Len(t, profiles, 1)
}

func TestGoogleGeminiChatValidationRejectsNonChatModel(t *testing.T) {
	binding := testGoogleGeminiRuntimeProfileBinding(
		"deep-research-preview-04-2026",
		[]string{"interactions"},
	)
	request := contractsopenai.ChatCompletionRequest{
		Messages: []contractsopenai.ChatMessage{{
			Role: "user", Content: "research",
		}},
	}

	require.ErrorContains(
		t,
		validateGoogleGeminiChatRuntimeProfile(request, &binding),
		"chat_completions",
	)
}

func TestGoogleGeminiChatValidationAcceptsDocumentedProfile(t *testing.T) {
	binding := testGoogleGeminiRuntimeProfileBinding(
		"gemini-3.1-pro-preview",
		[]string{"chat_completions", "batches"},
	)
	request := contractsopenai.ChatCompletionRequest{
		ServiceTier: "flex",
		Messages: []contractsopenai.ChatMessage{{
			Role: "user", Content: "test",
		}},
		Tools: []contractsopenai.ToolDefinition{{
			Type: "function", Name: "lookup",
		}},
	}

	require.NoError(
		t,
		validateGoogleGeminiChatRuntimeProfile(request, &binding),
	)
}

func testAnthropicRuntimeProfileBinding(
	model string,
) platform.ModelRuntimeProfileBinding {
	return platform.ModelRuntimeProfileBinding{
		Provider:        "anthropic",
		CatalogRevision: "runtime-revision",
		CanonicalID:     "anthropic/" + model,
		Slug:            model,
		ProfileVersion:  1,
		EffectiveFrom:   time.Now().UTC(),
		Profile: platform.ModelRuntimeProfile{
			Version:   1,
			Endpoints: []string{"messages", "count_tokens", "message_batches"},
			Thinking: platform.ModelRuntimeThinkingProfile{
				DefaultMode:  "manual",
				AllowedModes: []string{"manual", "disabled"},
				EffortValues: []string{"low", "medium", "high", "max"},
			},
			AssistantPrefill: true,
			Sampling: platform.ModelRuntimeSamplingProfile{
				Parameters:             []string{"temperature", "top_p", "top_k"},
				DisallowedWithThinking: []string{"temperature", "top_p", "top_k"},
			},
			Tools: platform.ModelRuntimeToolsProfile{
				Custom: true, Server: true, Parallel: true, StructuredOutput: true,
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
	}
}

func testGoogleGeminiRuntimeProfileBinding(
	model string,
	endpoints []string,
) platform.ModelRuntimeProfileBinding {
	return platform.ModelRuntimeProfileBinding{
		Provider:        "google-gemini",
		CatalogRevision: "gemini-runtime-revision",
		CanonicalID:     "google-gemini/" + model,
		Slug:            model,
		ProfileVersion:  1,
		EffectiveFrom:   time.Now().UTC(),
		Profile: platform.ModelRuntimeProfile{
			Version:   1,
			Endpoints: endpoints,
			Thinking: platform.ModelRuntimeThinkingProfile{
				DefaultMode:  "disabled",
				AllowedModes: []string{"disabled"},
			},
			Sampling: platform.ModelRuntimeSamplingProfile{
				Parameters: []string{"temperature", "top_p", "top_k"},
			},
			Tools: platform.ModelRuntimeToolsProfile{
				Custom: true, Parallel: true, StructuredOutput: true,
			},
			Modalities: platform.ModelRuntimeModalities{
				Input:  []string{"text", "image", "audio"},
				Output: []string{"text"},
			},
			RequestControls: platform.ModelRuntimeRequestControls{
				ServiceTierValues: []string{
					"standard", "flex", "priority",
				},
				InferenceGeoValues: []string{"global"},
			},
			Limits: platform.ModelRuntimeLimits{
				ContextWindow: 1_000_000, MaxOutputTokens: 64_000,
				MessagesBytes: 32 * 1024 * 1024,
				BatchBytes:    256 * 1024 * 1024,
				FileBytes:     500 * 1024 * 1024,
			},
		},
	}
}
