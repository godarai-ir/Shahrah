package proxy

import (
	"encoding/json"
	"testing"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/stretchr/testify/require"
)

func TestMoonshotRuntimeProfileEnforcesPerModelThinking(t *testing.T) {
	k3 := moonshotProxyRuntimeProfile("kimi-k3")
	require.NoError(t, validateMoonshotChatRuntimeProfile(
		contractsopenai.ChatCompletionRequest{
			Messages:        []contractsopenai.ChatMessage{{Role: "user", Content: "hello"}},
			ReasoningEffort: "high",
		},
		&k3,
	))
	require.NoError(t, validateMoonshotChatRuntimeProfile(
		contractsopenai.ChatCompletionRequest{
			Messages: []contractsopenai.ChatMessage{{Role: "user", Content: "hello"}},
			Thinking: &contractsopenai.ThinkingConfig{Type: "enabled"},
		},
		&k3,
	))
	require.NoError(t, validateMoonshotChatRuntimeProfile(
		contractsopenai.ChatCompletionRequest{
			Messages: []contractsopenai.ChatMessage{{Role: "user", Content: "hello"}},
			Thinking: &contractsopenai.ThinkingConfig{Type: "adaptive"},
		},
		&k3,
	))
	require.ErrorContains(t, validateMoonshotChatRuntimeProfile(
		contractsopenai.ChatCompletionRequest{
			Messages: []contractsopenai.ChatMessage{{Role: "user", Content: "hello"}},
			Thinking: &contractsopenai.ThinkingConfig{Type: "disabled"},
		},
		&k3,
	), "does not support disabling thinking")

	k27 := moonshotProxyRuntimeProfile("kimi-k2.7-code")
	require.NoError(t, validateMoonshotChatRuntimeProfile(
		contractsopenai.ChatCompletionRequest{
			Messages: []contractsopenai.ChatMessage{{Role: "user", Content: "hello"}},
			Thinking: &contractsopenai.ThinkingConfig{Type: "adaptive"},
		},
		&k27,
	))
	require.ErrorContains(t, validateMoonshotChatRuntimeProfile(
		contractsopenai.ChatCompletionRequest{
			Messages: []contractsopenai.ChatMessage{{Role: "user", Content: "hello"}},
			Thinking: &contractsopenai.ThinkingConfig{Type: "disabled"},
		},
		&k27,
	), "always has thinking enabled")
}

func TestMoonshotRuntimeProfileRejectsIntentChangingFixedParameters(t *testing.T) {
	binding := moonshotProxyRuntimeProfile("kimi-k2.7-code")

	for _, test := range []struct {
		name       string
		additional map[string]json.RawMessage
		want       string
	}{
		{
			name: "multiple choices",
			additional: map[string]json.RawMessage{
				"n": json.RawMessage(`2`),
			},
			want: "n must equal 1",
		},
		{
			name: "presence penalty",
			additional: map[string]json.RawMessage{
				"presence_penalty": json.RawMessage(`0.5`),
			},
			want: "presence_penalty must equal 0",
		},
		{
			name: "frequency penalty",
			additional: map[string]json.RawMessage{
				"frequency_penalty": json.RawMessage(`-0.5`),
			},
			want: "frequency_penalty must equal 0",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			require.ErrorContains(t, validateMoonshotChatRuntimeProfile(
				contractsopenai.ChatCompletionRequest{
					Messages:         []contractsopenai.ChatMessage{{Role: "user", Content: "hello"}},
					AdditionalFields: test.additional,
				},
				&binding,
			), test.want)
		})
	}

	require.NoError(t, validateMoonshotChatRuntimeProfile(
		contractsopenai.ChatCompletionRequest{
			Messages: []contractsopenai.ChatMessage{{Role: "user", Content: "hello"}},
			AdditionalFields: map[string]json.RawMessage{
				"n":                 json.RawMessage(`null`),
				"presence_penalty":  json.RawMessage(`0`),
				"frequency_penalty": json.RawMessage(`null`),
			},
		},
		&binding,
	))

	require.NoError(t, validateMoonshotFixedAdditionalFields(
		map[string]json.RawMessage{
			"n":                 json.RawMessage(`5`),
			"presence_penalty":  json.RawMessage(`1`),
			"frequency_penalty": json.RawMessage(`-1`),
		},
		"moonshot-v1-128k",
	))
}

func TestMoonshotRuntimeProfilePassesSamplingAndEnforcesToolChoice(t *testing.T) {
	binding := moonshotProxyRuntimeProfile("kimi-k2.6")
	wrongTemperature := 1.0
	require.NoError(t, validateMoonshotChatRuntimeProfile(
		contractsopenai.ChatCompletionRequest{
			Messages:    []contractsopenai.ChatMessage{{Role: "user", Content: "hello"}},
			Thinking:    &contractsopenai.ThinkingConfig{Type: "disabled"},
			Temperature: &wrongTemperature,
		},
		&binding,
	))

	require.ErrorContains(t, validateMoonshotChatRuntimeProfile(
		contractsopenai.ChatCompletionRequest{
			Messages:   []contractsopenai.ChatMessage{{Role: "user", Content: "hello"}},
			ToolChoice: "required",
		},
		&binding,
	), "tool_choice=required")

	specificTool := map[string]interface{}{
		"type": "function",
		"function": map[string]interface{}{
			"name": "lookup",
		},
	}
	require.ErrorContains(t, validateMoonshotChatRuntimeProfile(
		contractsopenai.ChatCompletionRequest{
			Messages:   []contractsopenai.ChatMessage{{Role: "user", Content: "hello"}},
			ToolChoice: specificTool,
		},
		&binding,
	), "incompatible with thinking enabled")
	require.NoError(t, validateMoonshotChatRuntimeProfile(
		contractsopenai.ChatCompletionRequest{
			Messages:   []contractsopenai.ChatMessage{{Role: "user", Content: "hello"}},
			Thinking:   &contractsopenai.ThinkingConfig{Type: "disabled"},
			ToolChoice: specificTool,
		},
		&binding,
	))
}

func TestMoonshotRuntimeProfileAllowsVideoForCurrentModels(t *testing.T) {
	request := contractsopenai.ChatCompletionRequest{
		Messages: []contractsopenai.ChatMessage{{
			Role: "user",
			Content: []contractsopenai.ChatContentPart{{
				Type: "video_url", VideoURL: &contractsopenai.ChatVideoURL{URL: "https://example.test/video.mp4"},
			}},
		}},
	}

	for _, model := range []string{"kimi-k3", "kimi-k2.7-code", "kimi-k2.7-code-highspeed", "kimi-k2.6"} {
		binding := moonshotProxyRuntimeProfile(model)
		require.NoError(t, validateMoonshotChatRuntimeProfile(request, &binding), model)
	}
}

func TestMoonshotRuntimeProfileValidatesDocumentedDecodingControls(t *testing.T) {
	binding := moonshotProxyRuntimeProfile("kimi-k2.7-code")

	for _, test := range []struct {
		name       string
		additional map[string]json.RawMessage
		want       string
	}{
		{name: "top k", additional: map[string]json.RawMessage{"top_k": json.RawMessage(`40`)}, want: "does not support top_k"},
		{name: "seed", additional: map[string]json.RawMessage{"seed": json.RawMessage(`42`)}, want: "does not support seed"},
		{name: "parallel tool calls", additional: map[string]json.RawMessage{"parallel_tool_calls": json.RawMessage(`true`)}, want: "does not support parallel_tool_calls"},
		{name: "invalid logprobs", additional: map[string]json.RawMessage{"logprobs": json.RawMessage(`"yes"`)}, want: "logprobs must be a boolean"},
		{name: "top logprobs range", additional: map[string]json.RawMessage{"logprobs": json.RawMessage(`true`), "top_logprobs": json.RawMessage(`21`)}, want: "between 0 and 20"},
		{name: "top logprobs dependency", additional: map[string]json.RawMessage{"top_logprobs": json.RawMessage(`5`)}, want: "requires logprobs=true"},
		{name: "invalid stop", additional: map[string]json.RawMessage{"stop": json.RawMessage(`{"value":"END"}`)}, want: "stop must be a string"},
	} {
		t.Run(test.name, func(t *testing.T) {
			require.ErrorContains(t, validateMoonshotChatRuntimeProfile(
				contractsopenai.ChatCompletionRequest{
					Messages:         []contractsopenai.ChatMessage{{Role: "user", Content: "hello"}},
					AdditionalFields: test.additional,
				},
				&binding,
			), test.want)
		})
	}

	require.NoError(t, validateMoonshotChatRuntimeProfile(
		contractsopenai.ChatCompletionRequest{
			Messages: []contractsopenai.ChatMessage{{Role: "user", Content: "hello"}},
			AdditionalFields: map[string]json.RawMessage{
				"logprobs":     json.RawMessage(`true`),
				"top_logprobs": json.RawMessage(`20`),
				"stop":         json.RawMessage(`["END"]`),
			},
		},
		&binding,
	))
}

func moonshotProxyRuntimeProfile(model string) platform.ModelRuntimeProfileBinding {
	modalities := []string{"text"}
	thinking := platform.ModelRuntimeThinkingProfile{
		DefaultMode: "adaptive", AllowedModes: []string{"adaptive"}, AlwaysEnabled: true,
	}
	unsupported := []string{"reasoning_effort"}

	if model == "kimi-k3" {
		modalities = []string{"text", "image", "video"}
		thinking.EffortValues = []string{"low", "high", "max"}
		unsupported = nil
	}

	if model == "kimi-k2.6" {
		modalities = []string{"text", "image", "video"}
		thinking = platform.ModelRuntimeThinkingProfile{
			DefaultMode: "adaptive", AllowedModes: []string{"adaptive", "disabled"},
		}
	}

	if model == "kimi-k2.7-code" || model == "kimi-k2.7-code-highspeed" {
		modalities = []string{"text", "image", "video"}
	}

	profile := platform.ModelRuntimeProfile{
		Version: 1, Endpoints: []string{"models", "chat_completions"},
		Transports: []string{"http", "sse"}, Lifecycle: "active", Thinking: thinking,
		Sampling: platform.ModelRuntimeSamplingProfile{Parameters: []string{"temperature", "top_p"}},
		Tools: platform.ModelRuntimeToolsProfile{
			Custom: true, Parallel: false, StructuredOutput: true,
			Supported: []string{"function_calling"},
		},
		Modalities: platform.ModelRuntimeModalities{Input: modalities, Output: []string{"text"}},
		RequestControls: platform.ModelRuntimeRequestControls{
			ServiceTierValues: []string{"auto"}, InferenceGeoValues: []string{"global"},
		},
		UnsupportedParameters: unsupported,
		Limits: platform.ModelRuntimeLimits{
			ContextWindow: 262144, MessagesBytes: 32 * 1024 * 1024,
			BatchBytes: 256 * 1024 * 1024, FileBytes: 500 * 1024 * 1024,
		},
	}
	if model == "kimi-k3" {
		profile.Sampling.Parameters = nil
		profile.Limits.ContextWindow = 1048576
		profile.Limits.MaxOutputTokens = 1048576
	}

	return platform.ModelRuntimeProfileBinding{
		Provider: "moonshot", CatalogRevision: "moonshot-test-revision",
		CanonicalID: "moonshot/" + model, Slug: model, ProfileVersion: 1,
		Profile: profile, SourceURL: "https://platform.kimi.ai/docs/models",
		EffectiveFrom: time.Now().UTC(),
	}
}
