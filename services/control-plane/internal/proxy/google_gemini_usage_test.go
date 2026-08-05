package proxy

import (
	"encoding/json"
	"testing"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/stretchr/testify/require"
)

func TestNormalizeGoogleGeminiUsageIsProviderScoped(t *testing.T) {
	t.Parallel()

	usage := &contractsopenai.Usage{PromptTokens: 6, CompletionTokens: 4, TotalTokens: 130}
	repair := normalizeGoogleGeminiUsage("openai", usage)

	require.False(t, repair.applied)
	require.Equal(t, 4, usage.CompletionTokens)
	require.Nil(t, usage.CompletionTokensDetails)
}

func TestGoogleGeminiStreamUsageNormalizerRepairsSplitUsageFrame(t *testing.T) {
	t.Parallel()

	normalizer := newGoogleGeminiStreamUsageNormalizer(googleGeminiProvider)
	first, repairs := normalizer.Add([]byte(`event: message
data: {"id":"chatcmpl-gemini","object":"chat.completion.chunk","future_field":{"preserved":true},"choices":[],"usage":{"prompt_tokens":6,`))
	require.Empty(t, first)
	require.Empty(t, repairs)

	second, repairs := normalizer.Add([]byte(`"completion_tokens":4,"total_tokens":130}}

`))
	require.Len(t, repairs, 1)
	require.True(t, repairs[0].applied)
	require.Equal(t, 120, repairs[0].inferredReasoningTokens)
	require.Contains(t, string(second), "event: message")

	frame, rest, ok := takeSSEFrame(string(second))
	require.True(t, ok)
	require.Empty(t, rest)

	var envelope map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(streamFrameData(frame)), &envelope))
	require.JSONEq(t, `{"preserved":true}`, string(envelope["future_field"]))

	var usage contractsopenai.Usage
	require.NoError(t, json.Unmarshal(envelope["usage"], &usage))
	require.Equal(t, 6, usage.PromptTokens)
	require.Equal(t, 124, usage.CompletionTokens)
	require.Equal(t, 130, usage.TotalTokens)
	require.NotNil(t, usage.CompletionTokensDetails)
	require.Equal(t, 120, usage.CompletionTokensDetails.ReasoningTokens)
}

func TestGoogleGeminiStreamUsageNormalizerLeavesOtherProvidersUntouched(t *testing.T) {
	t.Parallel()

	data := []byte("data: {\"usage\":{\"prompt_tokens\":6,\"completion_tokens\":4,\"total_tokens\":130}}\n\n")
	normalizer := newGoogleGeminiStreamUsageNormalizer("openai")
	normalized, repairs := normalizer.Add(data)

	require.Equal(t, data, normalized)
	require.Empty(t, repairs)
}

func TestGoogleGeminiStreamUsageNormalizerFlushesIncompleteDataUnchanged(t *testing.T) {
	t.Parallel()

	normalizer := newGoogleGeminiStreamUsageNormalizer(googleGeminiProvider)
	data := []byte(`data: {"usage":{"prompt_tokens":6`)
	normalized, repairs := normalizer.Add(data)

	require.Empty(t, normalized)
	require.Empty(t, repairs)
	require.Equal(t, data, normalizer.Flush())
	require.Empty(t, normalizer.Flush())
}
