package proxy

import (
	"encoding/json"
	"testing"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/stretchr/testify/require"
)

func TestBuildChatCompletionPayloadPreservesFutureOpenAIFields(t *testing.T) {
	var request contractsopenai.ChatCompletionRequest
	require.NoError(t, json.Unmarshal([]byte(`{
		"model":"public-model",
		"messages":[{"role":"user","content":"hello"}],
		"max_completion_tokens":123,
		"prediction":{"type":"content","content":"expected"},
		"future_parameter":{"revision":2},
		"metadata":{"customer_tag":"preserved"},
		"prompt_version_fqn":"internal",
		"prompt_variables":{"name":"value"}
	}`), &request))

	payload, err := buildChatCompletionPayload(request, "gpt-4.1", "")
	require.NoError(t, err)

	var got map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(payload, &got))
	require.JSONEq(t, `"gpt-4.1"`, string(got["model"]))
	require.JSONEq(t, `123`, string(got["max_completion_tokens"]))
	require.JSONEq(t, `{"type":"content","content":"expected"}`, string(got["prediction"]))
	require.JSONEq(t, `{"revision":2}`, string(got["future_parameter"]))
	require.JSONEq(t, `{"customer_tag":"preserved"}`, string(got["metadata"]))
	require.NotContains(t, got, "prompt_version_fqn")
	require.NotContains(t, got, "prompt_variables")
}

// GPT-6 Sol and Luna support Chat Completions function calling only with
// reasoning_effort "none"; the gateway forwards the combination and the other
// documented GPT-6 fields unchanged and leaves enforcement to OpenAI.
func TestBuildChatCompletionPayloadPreservesGPT6Fields(t *testing.T) {
	var request contractsopenai.ChatCompletionRequest
	require.NoError(t, json.Unmarshal([]byte(`{
		"model":"public-model",
		"messages":[{"role":"user","content":"What is the weather in Paris?"}],
		"reasoning_effort":"none",
		"service_tier":"fast",
		"prompt_cache_options":{"ttl":"30m"},
		"tools":[{"type":"function","function":{"name":"get_weather","parameters":{"type":"object","properties":{"city":{"type":"string"}},"required":["city"]}}}],
		"temperature":0.2
	}`), &request))

	payload, err := buildChatCompletionPayload(request, "gpt-6-luna", "")
	require.NoError(t, err)

	var got map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(payload, &got))
	require.JSONEq(t, `"gpt-6-luna"`, string(got["model"]))
	require.JSONEq(t, `"none"`, string(got["reasoning_effort"]))
	require.JSONEq(t, `"fast"`, string(got["service_tier"]))
	require.JSONEq(t, `{"ttl":"30m"}`, string(got["prompt_cache_options"]))
	require.JSONEq(t, `0.2`, string(got["temperature"]))
	require.JSONEq(
		t,
		`[{"type":"function","function":{"name":"get_weather","parameters":{"type":"object","properties":{"city":{"type":"string"}},"required":["city"]}}}]`,
		string(got["tools"]),
	)
}

func TestBuildChatCompletionPayloadDoesNotInventMetadataOrStoreForProviders(t *testing.T) {
	t.Parallel()

	request := contractsopenai.ChatCompletionRequest{
		Model:    "public-model",
		Messages: []contractsopenai.ChatMessage{{Role: "user", Content: "hello"}},
	}

	for _, providerModel := range []string{"openai/gpt-5.6-terra", "google/gemini-3.6-flash"} {
		t.Run(providerModel, func(t *testing.T) {
			t.Parallel()

			payload, err := buildChatCompletionPayload(request, providerModel, "")
			require.NoError(t, err)

			var got map[string]json.RawMessage
			require.NoError(t, json.Unmarshal(payload, &got))
			require.NotContains(t, got, "metadata")
			require.NotContains(t, got, "store")
		})
	}
}
