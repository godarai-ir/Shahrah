package proxy

import (
	"encoding/json"
	"testing"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/stretchr/testify/require"
)

func TestDeepSeekRuntimeProfileAcceptsDocumentedModes(t *testing.T) {
	binding := deepSeekProxyRuntimeProfile("deepseek-v4-flash")
	req := contractsopenai.ChatCompletionRequest{
		Messages:        []contractsopenai.ChatMessage{{Role: "user", Content: "hello"}},
		ReasoningEffort: "max",
		Thinking:        &contractsopenai.ThinkingConfig{Type: "enabled"},
		ResponseFormat:  map[string]interface{}{"type": "json_object"},
	}

	require.NoError(t, validateDeepSeekChatRuntimeProfile(req, &binding))

	temperature := 0.7
	req.Thinking = &contractsopenai.ThinkingConfig{Type: "disabled"}
	req.ReasoningEffort = ""
	req.Temperature = &temperature
	require.NoError(t, validateDeepSeekChatRuntimeProfile(req, &binding))

	req.Tools = []contractsopenai.ToolDefinition{{
		Type: "function",
		Function: &contractsopenai.FunctionDef{
			Name: "lookup",
		},
	}}
	req.ToolChoice = "required"
	require.NoError(t, validateDeepSeekChatRuntimeProfile(req, &binding))
}

func TestDeepSeekRuntimeProfilePassesSamplingAndRejectsGatewayManagedParameters(t *testing.T) {
	binding := deepSeekProxyRuntimeProfile("deepseek-v4-pro")
	temperature := 0.7
	req := contractsopenai.ChatCompletionRequest{
		Messages:       []contractsopenai.ChatMessage{{Role: "user", Content: "hello"}},
		Temperature:    &temperature,
		ResponseFormat: map[string]interface{}{"type": "json_schema"},
		AdditionalFields: map[string]json.RawMessage{
			"user_id": json.RawMessage(`"customer@example.com"`),
		},
	}

	require.ErrorContains(t, validateDeepSeekChatRuntimeProfile(req, &binding), "json_object")
	req.Temperature = nil
	req.ResponseFormat = nil
	require.ErrorContains(t, validateDeepSeekChatRuntimeProfile(req, &binding), "reserved")

	delete(req.AdditionalFields, "user_id")
	req.Temperature = &temperature
	require.NoError(t, validateDeepSeekChatRuntimeProfile(req, &binding))
}

func TestDeepSeekRuntimeProfileRejectsUnsupportedRolesModalitiesLimitsAndTools(t *testing.T) {
	binding := deepSeekProxyRuntimeProfile("deepseek-v4-flash")
	maxTokens := binding.Profile.Limits.MaxOutputTokens + 1

	tests := []struct {
		name    string
		mutate  func(*contractsopenai.ChatCompletionRequest)
		message string
	}{
		{
			name: "role",
			mutate: func(req *contractsopenai.ChatCompletionRequest) {
				req.Messages[0].Role = "developer"
			},
			message: "message role",
		},
		{
			name: "image",
			mutate: func(req *contractsopenai.ChatCompletionRequest) {
				req.Messages[0].Content = []interface{}{map[string]interface{}{
					"type": "image_url",
					"image_url": map[string]interface{}{
						"url": "https://example.invalid/image.png",
					},
				}}
			},
			message: "text input only",
		},
		{
			name: "output limit",
			mutate: func(req *contractsopenai.ChatCompletionRequest) {
				req.MaxTokens = &maxTokens
			},
			message: "output tokens",
		},
		{
			name: "strict tool",
			mutate: func(req *contractsopenai.ChatCompletionRequest) {
				strict := true
				req.Tools = []contractsopenai.ToolDefinition{{
					Type: "function",
					Function: &contractsopenai.FunctionDef{
						Name:   "lookup",
						Strict: &strict,
					},
				}}
			},
			message: "strict tool mode",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := contractsopenai.ChatCompletionRequest{
				Messages: []contractsopenai.ChatMessage{{Role: "user", Content: "hello"}},
				Thinking: &contractsopenai.ThinkingConfig{Type: "disabled"},
			}
			test.mutate(&req)

			require.ErrorContains(t, validateDeepSeekChatRuntimeProfile(req, &binding), test.message)
		})
	}
}

func TestDeepSeekRuntimeProfileEnforcesCatalogThinkingModes(t *testing.T) {
	binding := deepSeekProxyRuntimeProfile("deepseek-v4-pro")
	binding.Profile.Thinking.DefaultMode = "disabled"
	binding.Profile.Thinking.AllowedModes = []string{"disabled"}
	req := contractsopenai.ChatCompletionRequest{
		Messages: []contractsopenai.ChatMessage{{Role: "user", Content: "hello"}},
		Thinking: &contractsopenai.ThinkingConfig{Type: "enabled"},
	}

	require.ErrorContains(t, validateDeepSeekChatRuntimeProfile(req, &binding), "thinking mode")
}

func TestDeepSeekResponsesRuntimeProfileSupportsBothCurrentModels(t *testing.T) {
	for _, model := range []string{"deepseek-v4-flash", "deepseek-v4-pro"} {
		t.Run(model, func(t *testing.T) {
			binding := deepSeekProxyRuntimeProfile(model)
			body := []byte(`{"model":"` + model + `","input":[{"role":"user","content":[{"type":"input_text","text":"hello"}]}],"reasoning":{"effort":"high"},"tools":[{"type":"function","name":"lookup","parameters":{"type":"object"}}],"tool_choice":"auto","stream":true}`)

			require.NoError(t, validateDeepSeekResponsesRuntimeProfile(body, &binding))
		})
	}
}

func TestDeepSeekResponsesRuntimeProfileRejectsStateAndMisleadingControls(t *testing.T) {
	binding := deepSeekProxyRuntimeProfile("deepseek-v4-flash")

	for _, test := range []struct {
		name    string
		body    string
		message string
	}{
		{name: "state", body: `{"model":"deepseek-v4-flash","input":"hello","store":true}`, message: "store"},
		{name: "identity", body: `{"model":"deepseek-v4-flash","input":"hello","user":"email@example.com"}`, message: "user"},
		{name: "image", body: `{"model":"deepseek-v4-flash","input":[{"type":"message","role":"user","content":[{"type":"input_image","image_url":"data:image/png;base64,secret"}]}]}`, message: "input_image"},
		{name: "hosted tool", body: `{"model":"deepseek-v4-flash","input":"hello","tools":[{"type":"web_search"}]}`, message: "web_search"},
		{name: "reasoning summary", body: `{"model":"deepseek-v4-flash","input":"hello","reasoning":{"effort":"high","summary":"auto"}}`, message: "summary"},
	} {
		t.Run(test.name, func(t *testing.T) {
			require.ErrorContains(t, validateDeepSeekResponsesRuntimeProfile([]byte(test.body), &binding), test.message)
		})
	}
}

func deepSeekProxyRuntimeProfile(model string) platform.ModelRuntimeProfileBinding {
	profileVersion := 1
	endpoints := []string{"models", "chat_completions"}
	parameters := map[string][]string(nil)

	if model == "deepseek-v4-flash" || model == "deepseek-v4-pro" {
		profileVersion = map[string]int{
			"deepseek-v4-flash": 3,
			"deepseek-v4-pro":   2,
		}[model]

		endpoints = append(endpoints, "responses")
		parameters = map[string][]string{
			"responses": {"input", "instructions", "max_output_tokens", "reasoning", "stream", "temperature", "top_p", "top_logprobs", "text", "tools", "tool_choice"},
		}
	}

	return platform.ModelRuntimeProfileBinding{
		Provider: "deepseek", CatalogRevision: "deepseek-test-revision",
		CanonicalID: "deepseek/" + model, Slug: model, ProfileVersion: profileVersion,
		EffectiveFrom: time.Now().UTC(),
		Profile: platform.ModelRuntimeProfile{
			Version: profileVersion, Endpoints: endpoints,
			Transports: []string{"http", "sse"}, Lifecycle: map[string]string{
				"deepseek-v4-flash": "preview",
				"deepseek-v4-pro":   "active",
			}[model],
			Thinking: platform.ModelRuntimeThinkingProfile{
				DefaultMode: "adaptive", AllowedModes: []string{"adaptive", "disabled"},
				EffortValues: []string{"low", "medium", "high", "xhigh", "max"},
			},
			Sampling: platform.ModelRuntimeSamplingProfile{
				Parameters:             []string{"temperature", "top_p"},
				DisallowedWithThinking: []string{"temperature", "top_p"},
			},
			Tools: platform.ModelRuntimeToolsProfile{
				Custom: true, StructuredOutput: true, Supported: []string{"function_calling"},
			},
			Modalities: platform.ModelRuntimeModalities{Input: []string{"text"}, Output: []string{"text"}},
			Parameters: parameters,
			Limits: platform.ModelRuntimeLimits{
				ContextWindow: 1048576, MaxInputTokens: 1048576, MaxOutputTokens: 393216,
				MessagesBytes: 32 * 1024 * 1024, BatchBytes: 256 * 1024 * 1024, FileBytes: 500 * 1024 * 1024,
			},
		},
	}
}
