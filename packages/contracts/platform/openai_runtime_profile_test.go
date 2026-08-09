package platform

import (
	"encoding/json"
	"testing"
)

func TestOpenAIRuntimeProfileV3JSONParity(t *testing.T) {
	const payload = `{
		"version":3,
		"endpoints":["models","chat_completions","responses","responses_input_tokens","responses_compact","conversations","uploads","vector_stores","containers","batches"],
		"transports":["http","sse"],
		"lifecycle":"active",
		"thinking":{"defaultMode":"manual","allowedModes":["manual"],"manualDeprecated":false,"effortValues":["none","low","medium","high","xhigh","max"],"reasoningModes":["standard","pro"],"contextValues":["auto","current_turn","all_turns"],"alwaysEnabled":false},
		"assistantPrefill":false,
		"fineTuning":false,
		"predictedOutputs":false,
		"sampling":{"parameters":["temperature","top_p"],"disallowedWithThinking":[]},
		"tools":{"custom":true,"server":true,"parallel":true,"structuredOutput":true,"supported":["function_calling","web_search","file_search","tool_search","image_generation","code_interpreter","hosted_shell","apply_patch","skills","computer_use","mcp"],"preferredVersions":{}},
		"modalities":{"input":["text","image"],"output":["text"]},
		"requestControls":{"serviceTierValues":["standard","batch","flex","priority","fast"],"inferenceGeoValues":["global"],"speedValues":[],"effectiveSpeed":{}},
		"betaHeaders":[],
		"betaCapabilities":[],
		"referenceRateLimits":[{"tier":1,"scope":"project","dynamic":false,"requestsPerMinute":500,"tokensPerMinute":500000,"batchQueueTokens":1500000}],
		"minimumRetentionDays":0,
		"zdrEligible":false,
		"limits":{"contextWindow":1050000,"maxInputTokens":0,"maxInputCharacters":0,"maxInputBytes":0,"maxOutputTokens":128000,"messagesBytes":33554432,"batchBytes":268435456,"fileBytes":524288000}
	}`

	var profile ModelRuntimeProfile
	if err := json.Unmarshal([]byte(payload), &profile); err != nil {
		t.Fatal(err)
	}

	if err := profile.Validate(); err != nil {
		t.Fatal(err)
	}

	if profile.Version != 3 || !profile.AllowsEndpoint("responses") || !profile.AllowsTransport("sse") || !profile.AllowsTool("tool_search") {
		t.Fatalf("OpenAI runtime profile parity failed: %+v", profile)
	}

	if len(profile.ReferenceRateLimits) != 1 || profile.ReferenceRateLimits[0].BatchQueueTokens != 1500000 {
		t.Fatalf("OpenAI rate-limit parity failed: %+v", profile.ReferenceRateLimits)
	}
}
