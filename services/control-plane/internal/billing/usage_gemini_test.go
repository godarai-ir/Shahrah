package billing

import (
	"testing"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/stretchr/testify/require"
)

func TestGoogleGeminiTokenMapperPreservesNativeUsageDetails(t *testing.T) {
	usage := googleGeminiTokenMapper(map[string]interface{}{
		"usageMetadata": map[string]interface{}{
			"promptTokenCount": float64(210000), "candidatesTokenCount": float64(100),
			"thoughtsTokenCount": float64(25), "cachedContentTokenCount": float64(1000),
			"totalTokenCount": float64(210125), "trafficType": "PRIORITY",
			"promptTokensDetails": []interface{}{
				map[string]interface{}{"modality": "AUDIO", "tokenCount": float64(12)},
				map[string]interface{}{"modality": "IMAGE", "tokenCount": float64(8)},
			},
			"candidatesTokensDetails": []interface{}{
				map[string]interface{}{"modality": "IMAGE", "tokenCount": float64(80)},
			},
		},
	})
	if usage.PromptTokens != 210000 || usage.CompletionTokens != 125 || usage.TotalTokens != 210125 {
		t.Fatalf("unexpected native usage: %+v", usage)
	}

	if usage.CacheReadInputTokens != 1000 || usage.ServiceTier != "priority" {
		t.Fatalf("missing cache/tier usage: %+v", usage)
	}

	if usage.PromptTokensDetails.AudioTokens != 12 || usage.PromptTokensDetails.ImageTokens != 8 ||
		usage.CompletionTokensDetails.ImageTokens != 80 || usage.CompletionTokensDetails.ReasoningTokens != 25 {
		t.Fatalf("missing modality/thought usage: %+v", usage)
	}
}

func TestGoogleGeminiTokenMapperNeverFabricatesMissingUsage(t *testing.T) {
	usage := googleGeminiTokenMapper(map[string]interface{}{})
	if usage.PromptTokens != 0 || usage.CompletionTokens != 0 || usage.TotalTokens != 0 {
		t.Fatalf("missing upstream usage must remain unknown, got %+v", usage)
	}
}

func TestGoogleGeminiTokenMapperNormalizesCompatibleThinkingUsage(t *testing.T) {
	usage := googleGeminiTokenMapper(map[string]interface{}{
		"prompt_tokens":     float64(6),
		"completion_tokens": float64(4),
		"total_tokens":      float64(130),
	})

	require.Equal(t, 6, usage.PromptTokens)
	require.Equal(t, 124, usage.CompletionTokens)
	require.Equal(t, 130, usage.TotalTokens)
	require.NotNil(t, usage.CompletionTokensDetails)
	require.Equal(t, 120, usage.CompletionTokensDetails.ReasoningTokens)
}

func TestNormalizeGoogleGeminiOpenAIUsage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                  string
		usage                 *contractsopenai.Usage
		wantChanged           bool
		wantInferredReasoning int
		wantCompletionTokens  int
		wantReasoningTokens   int
		wantCompletionDetails bool
	}{
		{
			name: "repairs hidden thinking tokens",
			usage: &contractsopenai.Usage{
				PromptTokens: 6, CompletionTokens: 4, TotalTokens: 130,
			},
			wantChanged:           true,
			wantInferredReasoning: 120,
			wantCompletionTokens:  124,
			wantReasoningTokens:   120,
			wantCompletionDetails: true,
		},
		{
			name: "leaves consistent usage unchanged",
			usage: &contractsopenai.Usage{
				PromptTokens: 6, CompletionTokens: 124, TotalTokens: 130,
			},
			wantCompletionTokens: 124,
		},
		{
			name: "preserves larger existing reasoning detail",
			usage: &contractsopenai.Usage{
				PromptTokens: 6, CompletionTokens: 4, TotalTokens: 130,
				CompletionTokensDetails: &contractsopenai.UsageTokenDetails{ReasoningTokens: 121},
			},
			wantChanged:           true,
			wantInferredReasoning: 120,
			wantCompletionTokens:  124,
			wantReasoningTokens:   121,
			wantCompletionDetails: true,
		},
		{
			name:                 "leaves nil usage unchanged",
			wantCompletionTokens: 0,
		},
		{
			name: "rejects total below prompt",
			usage: &contractsopenai.Usage{
				PromptTokens: 10, CompletionTokens: 2, TotalTokens: 9,
			},
			wantCompletionTokens: 2,
		},
		{
			name: "rejects negative values",
			usage: &contractsopenai.Usage{
				PromptTokens: -1, CompletionTokens: 2, TotalTokens: 1,
			},
			wantCompletionTokens: 2,
		},
		{
			name: "does not lower completion above canonical total",
			usage: &contractsopenai.Usage{
				PromptTokens: 6, CompletionTokens: 125, TotalTokens: 130,
			},
			wantCompletionTokens: 125,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			inferredReasoning, changed := NormalizeGoogleGeminiOpenAIUsage(test.usage)
			require.Equal(t, test.wantChanged, changed)
			require.Equal(t, test.wantInferredReasoning, inferredReasoning)

			if test.usage == nil {
				return
			}

			require.Equal(t, test.wantCompletionTokens, test.usage.CompletionTokens)

			if !test.wantCompletionDetails {
				require.Nil(t, test.usage.CompletionTokensDetails)

				return
			}

			require.NotNil(t, test.usage.CompletionTokensDetails)
			require.Equal(t, test.wantReasoningTokens, test.usage.CompletionTokensDetails.ReasoningTokens)
		})
	}
}

func TestNormalizeGoogleGeminiOpenAIUsageIsIdempotent(t *testing.T) {
	t.Parallel()

	usage := &contractsopenai.Usage{PromptTokens: 6, CompletionTokens: 4, TotalTokens: 130}
	inferredReasoning, changed := NormalizeGoogleGeminiOpenAIUsage(usage)
	require.True(t, changed)
	require.Equal(t, 120, inferredReasoning)

	first := *usage
	firstDetails := *usage.CompletionTokensDetails

	inferredReasoning, changed = NormalizeGoogleGeminiOpenAIUsage(usage)
	require.False(t, changed)
	require.Zero(t, inferredReasoning)
	require.Equal(t, first, *usage)
	require.Equal(t, firstDetails, *usage.CompletionTokensDetails)
}
