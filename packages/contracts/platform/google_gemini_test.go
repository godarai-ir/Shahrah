package platform

import (
	"encoding/json"
	"testing"
)

func TestSupportedGoogleGeminiModelIDsAreExact(t *testing.T) {
	expected := []string{
		"gemini-3.1-pro-preview", "gemini-3.5-flash", "gemini-3.5-flash-lite",
		"gemini-3.6-flash", "gemini-3-pro-image", "gemini-3.1-flash-image",
		"gemini-3.1-flash-lite-image", "gemini-3.1-flash-tts-preview", "gemini-embedding-2",
	}
	if len(SupportedGoogleGeminiModelIDs) != len(expected) {
		t.Fatalf("unexpected model count %d", len(SupportedGoogleGeminiModelIDs))
	}
	for index, model := range expected {
		if SupportedGoogleGeminiModelIDs[index] != model || !IsSupportedGoogleGeminiModel(model) {
			t.Fatalf("allowlist mismatch at %d", index)
		}
	}
	for _, model := range []string{"gemini-latest", "gemini-3.1-pro-preview-customtools", "models/veo-3.1"} {
		if IsSupportedGoogleGeminiModel(model) {
			t.Fatalf("unexpected model %q", model)
		}
	}
}

func TestGoogleGeminiRuntimeProfileV2JSONParity(t *testing.T) {
	const payload = `{
		"version":2,
		"endpoints":["models","interactions","generate_content","stream_generate_content","count_tokens","batches","batch_generate_content","files","cached_contents","chat_completions"],
		"transports":["http","sse"],
		"lifecycle":"active",
		"thinking":{"defaultMode":"manual","allowedModes":["manual"],"manualDeprecated":false,"effortValues":["medium","high"],"reasoningModes":[],"contextValues":[],"alwaysEnabled":false},
		"assistantPrefill":false,
		"fineTuning":false,
		"predictedOutputs":false,
		"sampling":{"parameters":[],"disallowedWithThinking":["temperature","top_p","top_k"]},
		"tools":{"custom":true,"server":true,"parallel":true,"structuredOutput":true,"supported":["function_calling","google_search","google_maps","url_context","file_search","code_execution","computer_use","mcp"],"preferredVersions":{}},
		"modalities":{"input":["text","image","audio","video","pdf"],"output":["text"]},
		"requestControls":{"serviceTierValues":["standard","batch","flex","priority"],"inferenceGeoValues":["global"],"speedValues":[],"effectiveSpeed":{}},
		"betaHeaders":[],
		"betaCapabilities":[],
		"referenceRateLimits":[{"usageTier":"tier_1","scope":"project","dynamic":true,"spendUsdPerTenMinutes":10,"concurrentBatchRequests":100,"inputFileBytes":2000000000,"fileStorageBytes":20000000000}],
		"minimumRetentionDays":0,
		"zdrEligible":false,
		"limits":{"contextWindow":1048576,"maxInputTokens":1048576,"maxOutputTokens":65536,"messagesBytes":33554432,"batchBytes":2000000000,"fileBytes":2000000000}
	}`
	var profile ModelRuntimeProfile
	if err := json.Unmarshal([]byte(payload), &profile); err != nil {
		t.Fatal(err)
	}
	if err := profile.Validate(); err != nil {
		t.Fatal(err)
	}
	if profile.Version != 2 || !profile.AllowsEndpoint("interactions") || !profile.AllowsTransport("sse") || !profile.AllowsTool("google_maps") {
		t.Fatalf("Gemini runtime profile parity failed: %+v", profile)
	}
	if len(profile.ReferenceRateLimits) != 1 || profile.ReferenceRateLimits[0].FileStorageBytes != 20000000000 {
		t.Fatalf("Gemini rate-limit parity failed: %+v", profile.ReferenceRateLimits)
	}
}
