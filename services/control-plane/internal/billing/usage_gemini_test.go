package billing

import "testing"

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
