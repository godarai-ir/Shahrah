package server

import (
	"context"
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

func TestMoonshotDispatchUsesCanonicalChatEndpointAndBearerAuth(t *testing.T) {
	var gotPath, gotAuthorization string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		gotPath = request.URL.Path
		gotAuthorization = request.Header.Get("Authorization")
		w.Header().Set("x-request-id", "moonshot-request-1")

		if _, writeErr := io.WriteString(w, `{"id":"chatcmpl_moonshot"}`); writeErr != nil {
			t.Errorf("write response: %v", writeErr)
		}
	}))
	defer server.Close()

	payload := moonshotTestPayload("kimi-k3", `{"model":"kimi-k3","messages":[{"role":"user","content":"hello"}],"reasoning_effort":"low"}`)
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
}

func TestMoonshotRejectsModelSpecificParametersBeforeDispatch(t *testing.T) {
	tests := []struct {
		name  string
		model string
		body  string
		want  string
	}{
		{name: "K3 rejects thinking", model: "kimi-k3", body: `{"model":"kimi-k3","messages":[{"role":"user","content":"hi"}],"thinking":{"type":"enabled"}}`, want: "moonshot_k3_thinking_parameter_unsupported"},
		{name: "K2.7 rejects reasoning effort", model: "kimi-k2.7-code", body: `{"model":"kimi-k2.7-code","messages":[{"role":"user","content":"hi"}],"reasoning_effort":"low"}`, want: "moonshot_unsupported_parameter: reasoning_effort"},
		{name: "K2.7 fixed temperature", model: "kimi-k2.7-code", body: `{"model":"kimi-k2.7-code","messages":[{"role":"user","content":"hi"}],"temperature":0.7}`, want: "moonshot_temperature_must_equal_1"},
		{name: "K2.6 non-thinking temperature", model: "kimi-k2.6", body: `{"model":"kimi-k2.6","messages":[{"role":"user","content":"hi"}],"thinking":{"type":"disabled"},"temperature":1}`, want: "moonshot_temperature_must_equal_0.6"},
		{name: "K2.7 required tool choice", model: "kimi-k2.7-code", body: `{"model":"kimi-k2.7-code","messages":[{"role":"user","content":"hi"}],"tool_choice":"required"}`, want: "moonshot_tool_choice_required_unsupported"},
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
		modalities = []string{"text", "image"}
		thinking = platform.ModelRuntimeThinkingProfile{
			DefaultMode: "adaptive", AllowedModes: []string{"adaptive", "disabled"},
		}
	}

	thinking.EffortValues = efforts

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
