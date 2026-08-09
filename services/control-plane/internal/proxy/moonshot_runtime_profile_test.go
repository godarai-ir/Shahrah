package proxy

import (
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
	require.ErrorContains(t, validateMoonshotChatRuntimeProfile(
		contractsopenai.ChatCompletionRequest{
			Messages: []contractsopenai.ChatMessage{{Role: "user", Content: "hello"}},
			Thinking: &contractsopenai.ThinkingConfig{Type: "enabled"},
		},
		&k3,
	), "does not accept the thinking parameter")

	k27 := moonshotProxyRuntimeProfile("kimi-k2.7-code")
	require.ErrorContains(t, validateMoonshotChatRuntimeProfile(
		contractsopenai.ChatCompletionRequest{
			Messages: []contractsopenai.ChatMessage{{Role: "user", Content: "hello"}},
			Thinking: &contractsopenai.ThinkingConfig{Type: "disabled"},
		},
		&k27,
	), "always has thinking enabled")
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
}

func TestMoonshotRuntimeProfileAllowsK3VideoOnly(t *testing.T) {
	request := contractsopenai.ChatCompletionRequest{
		Messages: []contractsopenai.ChatMessage{{
			Role: "user",
			Content: []contractsopenai.ChatContentPart{{
				Type: "video_url", VideoURL: &contractsopenai.ChatVideoURL{URL: "https://example.test/video.mp4"},
			}},
		}},
	}

	k3 := moonshotProxyRuntimeProfile("kimi-k3")
	require.NoError(t, validateMoonshotChatRuntimeProfile(request, &k3))

	k26 := moonshotProxyRuntimeProfile("kimi-k2.6")
	require.ErrorContains(t, validateMoonshotChatRuntimeProfile(request, &k26), "video input")
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
		modalities = []string{"text", "image"}
		thinking = platform.ModelRuntimeThinkingProfile{
			DefaultMode: "adaptive", AllowedModes: []string{"adaptive", "disabled"},
		}
	}

	profile := platform.ModelRuntimeProfile{
		Version: 1, Endpoints: []string{"models", "chat_completions"},
		Transports: []string{"http", "sse"}, Lifecycle: "active", Thinking: thinking,
		Sampling: platform.ModelRuntimeSamplingProfile{Parameters: []string{"temperature", "top_p"}},
		Tools: platform.ModelRuntimeToolsProfile{
			Custom: true, Parallel: true, StructuredOutput: true,
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
