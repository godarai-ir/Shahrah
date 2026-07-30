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
	require.NotContains(t, got, "prompt_version_fqn")
	require.NotContains(t, got, "prompt_variables")
}
