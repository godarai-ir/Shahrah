package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func TestMoonshotRegistryUsesDedicatedAdapter(t *testing.T) {
	adapter, err := newProviderRegistry().adapterFor("moonshot")
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := adapter.(*moonshotAdapter); !ok {
		t.Fatalf("expected dedicated Moonshot adapter, got %T", adapter)
	}
}

func TestMoonshotDispatchUsesCanonicalChatEndpointAndNormalizesK3(t *testing.T) {
	var gotPath, gotAuthorization string

	var gotBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		gotPath = request.URL.Path
		gotAuthorization = request.Header.Get("Authorization")

		var readErr error

		gotBody, readErr = io.ReadAll(request.Body)
		if readErr != nil {
			t.Errorf("read request: %v", readErr)
		}

		w.Header().Set("x-request-id", "moonshot-request-1")

		if _, writeErr := io.WriteString(w, `{"id":"chatcmpl_moonshot"}`); writeErr != nil {
			t.Errorf("write response: %v", writeErr)
		}
	}))
	defer server.Close()

	payload := moonshotTestPayload("kimi-k3", `{
		"model":"kimi-k3","messages":[{"role":"user","content":"hello"}],
		"reasoning_effort":"low","thinking":{"type":"enabled"},
		"temperature":7.25,"top_p":-0.5,
		"presence_penalty":0,"frequency_penalty":0,
		"stop":["END"],"logprobs":true,"top_logprobs":20
	}`)
	payload.Deployment.APIBase = server.URL
	payload.Deployment.APIKey = "moonshot-secret"

	response, err := newMoonshotAdapter().Dispatch(context.Background(), server.Client(), payload)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if closeErr := response.Body.Close(); closeErr != nil {
			t.Errorf("close response: %v", closeErr)
		}
	}()

	if gotPath != "/chat/completions" || gotAuthorization != "Bearer moonshot-secret" {
		t.Fatalf("unexpected upstream request path=%q auth=%q", gotPath, gotAuthorization)
	}

	if response.Headers.Get("x-request-id") != "moonshot-request-1" {
		t.Fatal("expected provider request ID to be preserved")
	}

	for _, name := range []string{"stop", "logprobs", "top_logprobs"} {
		if !strings.Contains(string(gotBody), `"`+name+`"`) {
			t.Fatalf("%s was not relayed: %s", name, gotBody)
		}
	}

	for _, name := range []string{"temperature", "top_p", "thinking", "presence_penalty", "frequency_penalty"} {
		if strings.Contains(string(gotBody), `"`+name+`"`) {
			t.Fatalf("%s was not removed: %s", name, gotBody)
		}
	}
}

func TestNormalizeMoonshotK26PreservesDisabledThinkingWithKeepAll(t *testing.T) {
	body := decodeMoonshotTestBody(t, `{"thinking":{"type":"disabled","keep":"all"}}`)

	changed, err := normalizeMoonshotChatBody(body, "kimi-k2.6")
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}

	if changed {
		t.Fatalf("valid K2.6 thinking controls changed: %s", mustMarshalMoonshotTestBody(t, body))
	}

	assertMoonshotRawValue(t, body, "thinking", `{"type":"disabled","keep":"all"}`)
}

func TestNormalizeMoonshotChatBodyUsesPerModelCanonicalControls(t *testing.T) {
	tests := []struct {
		name            string
		model           string
		body            string
		wantTemperature string
		wantTopP        string
		wantThinking    string
	}{
		{
			name: "K2.7 code", model: "kimi-k2.7-code",
			body: `{"temperature":0.1,"top_p":1}`,
		},
		{
			name: "K2.7 code highspeed", model: "kimi-k2.7-code-highspeed",
			body: `{"temperature":"client-default","top_p":-1}`,
		},
		{
			name: "K2.7 adaptive thinking", model: "kimi-k2.7-code",
			body:         `{"thinking":{"type":"adaptive"}}`,
			wantThinking: `{"type":"enabled"}`,
		},
		{
			name: "K2.6 adaptive", model: "kimi-k2.6",
			body:         `{"thinking":{"type":"adaptive"},"temperature":0.1,"top_p":1}`,
			wantThinking: `{"type":"enabled"}`,
		},
		{
			name: "K2.6 disabled", model: "kimi-k2.6",
			body:         `{"thinking":{"type":"disabled"},"temperature":0.1,"top_p":1}`,
			wantThinking: `{"type":"disabled"}`,
		},
		{
			name: "K2.6 adaptive preserved thinking", model: "kimi-k2.6",
			body:         `{"thinking":{"type":"adaptive","keep":"all"}}`,
			wantThinking: `{"type":"enabled","keep":"all"}`,
		},
		{
			name: "K2.7 preserved thinking", model: "kimi-k2.7-code",
			body:         `{"thinking":{"type":"enabled","keep":"all"},"temperature":1}`,
			wantThinking: `{"type":"enabled","keep":"all"}`,
		},
		{
			name: "K3", model: "kimi-k3",
			body: `{"thinking":{"type":"adaptive","keep":"all"},"temperature":0.1,"top_p":1}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := decodeMoonshotTestBody(t, test.body)

			changed, err := normalizeMoonshotChatBody(body, test.model)
			if err != nil {
				t.Fatalf("normalize: %v", err)
			}

			if !changed {
				t.Fatal("expected request normalization")
			}

			assertMoonshotRawValue(t, body, "temperature", test.wantTemperature)
			assertMoonshotRawValue(t, body, "top_p", test.wantTopP)
			assertMoonshotRawValue(t, body, "thinking", test.wantThinking)
		})
	}
}

func TestNormalizeMoonshotChatBodyRemovesEquivalentFixedControls(t *testing.T) {
	body := decodeMoonshotTestBody(t, `{
		"n":1,"presence_penalty":0,"frequency_penalty":0,
		"temperature":1,"top_p":0.95
	}`)

	changed, err := normalizeMoonshotChatBody(body, "kimi-k2.7-code")
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}

	if !changed || len(body) != 0 {
		t.Fatalf("equivalent fixed controls remain: %s", mustMarshalMoonshotTestBody(t, body))
	}
}

func TestNormalizeMoonshotChatBodyPreservesAbsentControlsAndRemovesNulls(t *testing.T) {
	body := decodeMoonshotTestBody(t, `{
		"temperature":null,"top_p":null,"thinking":null,"reasoning_effort":null,
		"n":null,"presence_penalty":null,"frequency_penalty":null,
		"top_k":null,"seed":null,"random_seed":null,"stop":null,
		"logit_bias":null,"logprobs":null,"top_logprobs":null,"tool_choice":null,
		"max_tokens":null,"max_completion_tokens":null,"parallel_tool_calls":null
	}`)

	changed, err := normalizeMoonshotChatBody(body, "kimi-k2.7-code")
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}

	if !changed {
		t.Fatal("expected null controls to be removed")
	}

	if len(body) != 0 {
		t.Fatalf("null controls remain: %s", mustMarshalMoonshotTestBody(t, body))
	}

	body = decodeMoonshotTestBody(t, `{"model":"kimi-k2.7-code"}`)

	changed, err = normalizeMoonshotChatBody(body, "kimi-k2.7-code")
	if err != nil {
		t.Fatalf("normalize absent controls: %v", err)
	}

	if changed {
		t.Fatalf("absent controls were injected: %s", mustMarshalMoonshotTestBody(t, body))
	}

	body = decodeMoonshotTestBody(t, `{"temperature":null}`)

	changed, err = normalizeMoonshotChatBody(body, "future-moonshot-model")
	if err != nil {
		t.Fatalf("normalize future model: %v", err)
	}

	if changed || string(body["temperature"]) != "null" {
		t.Fatalf("future model controls changed: %s", mustMarshalMoonshotTestBody(t, body))
	}
}

func TestMoonshotRejectsModelSpecificParametersBeforeDispatch(t *testing.T) {
	tests := []struct {
		name  string
		model string
		body  string
		want  string
	}{
		{name: "K3 rejects disabled thinking", model: "kimi-k3", body: `{"model":"kimi-k3","messages":[{"role":"user","content":"hi"}],"thinking":{"type":"disabled"}}`, want: "moonshot_k3_thinking_disable_unsupported"},
		{name: "K3 rejects thinking budget", model: "kimi-k3", body: `{"model":"kimi-k3","messages":[{"role":"user","content":"hi"}],"thinking":{"type":"enabled","budget_tokens":1024}}`, want: "moonshot_k3_thinking_controls_unsupported"},
		{name: "K2.7 rejects disabled thinking", model: "kimi-k2.7-code", body: `{"model":"kimi-k2.7-code","messages":[{"role":"user","content":"hi"}],"thinking":{"type":"disabled"}}`, want: "moonshot_k2_7_thinking_always_enabled"},
		{name: "K2.7 rejects reasoning effort", model: "kimi-k2.7-code", body: `{"model":"kimi-k2.7-code","messages":[{"role":"user","content":"hi"}],"reasoning_effort":"low"}`, want: "moonshot_unsupported_parameter: reasoning_effort"},
		{name: "K2.7 required tool choice", model: "kimi-k2.7-code", body: `{"model":"kimi-k2.7-code","messages":[{"role":"user","content":"hi"}],"tool_choice":"required"}`, want: "moonshot_tool_choice_required_unsupported"},
		{name: "K2.6 rejects thinking budget", model: "kimi-k2.6", body: `{"model":"kimi-k2.6","messages":[{"role":"user","content":"hi"}],"thinking":{"type":"enabled","budget_tokens":1024}}`, want: "moonshot_k2_6_thinking_controls_unsupported"},
		{name: "K3 rejects specific tool choice while thinking", model: "kimi-k3", body: `{"model":"kimi-k3","messages":[{"role":"user","content":"hi"}],"tool_choice":{"type":"function","function":{"name":"lookup"}}}`, want: "moonshot_specific_tool_choice_incompatible_with_thinking"},
		{name: "K2.6 rejects specific tool choice while thinking", model: "kimi-k2.6", body: `{"model":"kimi-k2.6","messages":[{"role":"user","content":"hi"}],"tool_choice":{"type":"function","function":{"name":"lookup"}}}`, want: "moonshot_specific_tool_choice_incompatible_with_thinking"},
		{name: "Moonshot rejects multiple choices", model: "kimi-k2.7-code", body: `{"model":"kimi-k2.7-code","messages":[{"role":"user","content":"hi"}],"n":2}`, want: "moonshot_n_must_equal_one"},
		{name: "Moonshot rejects presence penalty", model: "kimi-k2.7-code", body: `{"model":"kimi-k2.7-code","messages":[{"role":"user","content":"hi"}],"presence_penalty":0.5}`, want: "moonshot_presence_penalty_must_equal_zero"},
		{name: "Moonshot rejects undocumented top k", model: "kimi-k2.7-code", body: `{"model":"kimi-k2.7-code","messages":[{"role":"user","content":"hi"}],"top_k":40}`, want: "moonshot_unsupported_parameter: top_k"},
		{name: "Moonshot rejects undocumented seed", model: "kimi-k2.7-code", body: `{"model":"kimi-k2.7-code","messages":[{"role":"user","content":"hi"}],"seed":42}`, want: "moonshot_unsupported_parameter: seed"},
		{name: "Moonshot rejects undocumented parallel tool control", model: "kimi-k2.7-code", body: `{"model":"kimi-k2.7-code","messages":[{"role":"user","content":"hi"}],"parallel_tool_calls":true}`, want: "moonshot_unsupported_parameter: parallel_tool_calls"},
		{name: "Moonshot rejects invalid stop", model: "kimi-k2.7-code", body: `{"model":"kimi-k2.7-code","messages":[{"role":"user","content":"hi"}],"stop":{"value":"END"}}`, want: "moonshot_invalid_parameter: stop"},
		{name: "Moonshot rejects invalid logprobs", model: "kimi-k2.7-code", body: `{"model":"kimi-k2.7-code","messages":[{"role":"user","content":"hi"}],"logprobs":"yes"}`, want: "moonshot_invalid_parameter: logprobs"},
		{name: "Moonshot requires logprobs for top logprobs", model: "kimi-k2.7-code", body: `{"model":"kimi-k2.7-code","messages":[{"role":"user","content":"hi"}],"top_logprobs":5}`, want: "moonshot_top_logprobs_requires_logprobs"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payload := moonshotTestPayload(test.model, test.body)

			_, err := newMoonshotAdapter().Dispatch(context.Background(), http.DefaultClient, payload)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("expected %q, got %v", test.want, err)
			}
		})
	}
}

func TestMoonshotLegacyModelsDoNotInheritCurrentKimiFixedControls(t *testing.T) {
	body := decodeMoonshotTestBody(t, `{"n":5,"presence_penalty":1,"frequency_penalty":-1}`)

	if err := validateMoonshotFixedCommon(body, "moonshot-v1-128k"); err != nil {
		t.Fatalf("legacy Moonshot V1 controls were rejected: %v", err)
	}
}

func TestMoonshotAllowsSpecificToolChoiceWhenK26ThinkingIsDisabled(t *testing.T) {
	body := decodeMoonshotTestBody(t, `{
		"thinking":{"type":"disabled","keep":"all"},
		"tool_choice":{"type":"function","function":{"name":"lookup"}}
	}`)

	if err := validateMoonshotToolChoice(body, "kimi-k2.6"); err != nil {
		t.Fatalf("valid K2.6 tool choice was rejected: %v", err)
	}
}

func TestMoonshotOutputLimitValidatesPositiveValuesWithoutKnownMaximum(t *testing.T) {
	body := decodeMoonshotTestBody(t, `{"max_completion_tokens":0}`)

	err := validateMoonshotOutputLimit(body, 0)
	if err == nil || !strings.Contains(err.Error(), "moonshot_invalid_parameter") {
		t.Fatalf("expected invalid output limit, got %v", err)
	}
}

func decodeMoonshotTestBody(t *testing.T, value string) map[string]json.RawMessage {
	t.Helper()

	var body map[string]json.RawMessage
	if err := json.Unmarshal([]byte(value), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}

	return body
}

func assertMoonshotRawValue(
	t *testing.T,
	body map[string]json.RawMessage,
	parameter string,
	want string,
) {
	t.Helper()

	got, present := body[parameter]
	if want == "" {
		if present {
			t.Fatalf("%s = %s, want absent", parameter, got)
		}

		return
	}

	if !present || string(got) != want {
		t.Fatalf("%s = %s, want %s", parameter, got, want)
	}
}

func mustMarshalMoonshotTestBody(t *testing.T, body map[string]json.RawMessage) []byte {
	t.Helper()

	value, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}

	return value
}

func TestMoonshotFailsClosedWithoutRuntimeProfile(t *testing.T) {
	payload := moonshotTestPayload("kimi-k3", `{"model":"kimi-k3","messages":[{"role":"user","content":"hi"}]}`)
	payload.ModelRuntimeProfiles = nil

	_, err := newMoonshotAdapter().Dispatch(context.Background(), http.DefaultClient, payload)
	if err == nil || !strings.Contains(err.Error(), "moonshot_runtime_profile_missing") {
		t.Fatalf("expected missing profile error, got %v", err)
	}
}

func moonshotTestPayload(model string, body string) platform.TunnelRequest {
	return platform.TunnelRequest{
		Operation:  platform.ProviderOperationChatCompletions,
		Deployment: platform.Deployment{Provider: "moonshot", Model: model},
		ModelRuntimeProfiles: []platform.ModelRuntimeProfileBinding{
			moonshotTestRuntimeProfile(model),
		},
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    []byte(body),
	}
}

func moonshotTestRuntimeProfile(model string) platform.ModelRuntimeProfileBinding {
	modalities := []string{"text"}
	thinking := platform.ModelRuntimeThinkingProfile{
		DefaultMode: "adaptive", AllowedModes: []string{"adaptive"}, AlwaysEnabled: true,
	}
	efforts := []string(nil)
	unsupported := []string{"reasoning_effort"}

	if model == "kimi-k3" {
		modalities = []string{"text", "image", "video"}
		efforts = []string{"low", "high", "max"}
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

	thinking.EffortValues = efforts

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
