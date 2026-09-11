package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

type zaiRoundTripperFunc func(*http.Request) (*http.Response, error)

func (fn zaiRoundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func TestZAIAdapterDispatchesChatWithBearerAuth(t *testing.T) {
	var gotPath, gotAuth string

	client := &http.Client{Transport: zaiRoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		gotPath = request.URL.Path
		gotAuth = request.Header.Get("Authorization")

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"choices":[],"usage":{"prompt_tokens":1,"completion_tokens":1}}`)),
		}, nil
	})}
	payload := zaiTestPayload("glm-5.2", "chat_completions")
	payload.Body = []byte(`{"model":"glm-5.2","messages":[{"role":"user","content":"hello"}],"stream":false}`)

	response, err := newZAIAdapter().Dispatch(context.Background(), client, payload)
	if err != nil {
		t.Fatalf("dispatch Z.ai: %v", err)
	}
	defer response.Body.Close()

	if gotPath != "/api/paas/v4/chat/completions" {
		t.Fatalf("path = %q", gotPath)
	}

	if gotAuth != "Bearer zai-secret" {
		t.Fatalf("authorization = %q", gotAuth)
	}
}

func TestZAIAdapterFailsClosedForUnsafeAndDisputedContracts(t *testing.T) {
	payload := zaiTestPayload("glm-5v-turbo", "chat_completions")

	payload.Body = []byte(`{"model":"glm-5v-turbo","messages":[],"tools":[]}`)
	if _, _, err := validateZAIDispatch(payload); err == nil || !strings.Contains(err.Error(), "unsupported_parameter: tools") {
		t.Fatalf("expected gated tools error, got %v", err)
	}

	payload = zaiTestPayload("glm-ocr", "layout_parsing")
	payload.Operation = platform.ProviderOperationZAILayoutParsing

	payload.Body = []byte(`{"model":"glm-ocr","file":"https://example.com/a.pdf","start_page_id":1,"end_page_id":31}`)
	if _, _, err := validateZAIDispatch(payload); err == nil || !strings.Contains(err.Error(), "page_range_invalid") {
		t.Fatalf("expected page limit error, got %v", err)
	}

	payload.Body = []byte(`{"model":"glm-ocr","file":"https://example.com/a.pdf","start_page_id":1,"end_page_id":2}`)

	payload.Deployment.APIBase = "https://127.0.0.1/api/paas/v4"
	if _, _, err := validateZAIDispatch(payload); err == nil || !strings.Contains(err.Error(), "invalid_api_base") {
		t.Fatalf("expected unsafe base error, got %v", err)
	}

	payload.Deployment.APIBase = "https://localhost/api/paas/v4"
	if _, _, err := validateZAIDispatch(payload); err == nil || !strings.Contains(err.Error(), "invalid_api_base") {
		t.Fatalf("expected unsafe base error, got %v", err)
	}
}

func TestZAIAdapterRoutesGLM53FlashToChatCompletions(t *testing.T) {
	payload := zaiTestPayload("glm-5.3-flash", "chat_completions")
	payload.Body = []byte(`{"model":"glm-5.3-flash","messages":[{"role":"user","content":"hello"}]}`)

	endpoint, _, err := validateZAIDispatch(payload)
	if err != nil {
		t.Fatalf("validate glm-5.3-flash dispatch: %v", err)
	}

	if endpoint != "chat_completions" {
		t.Fatalf("endpoint = %q", endpoint)
	}
}

func TestZAIAdapterRejectsDisabledThinkingForAlwaysEnabledModels(t *testing.T) {
	payload := zaiTestPayload("glm-5.3-flash", "chat_completions")
	payload.ModelRuntimeProfiles[0].Profile.Thinking = platform.ModelRuntimeThinkingProfile{
		DefaultMode: "adaptive", AllowedModes: []string{"adaptive"}, AlwaysEnabled: true,
	}
	payload.ModelRuntimeProfiles[0].Profile.Parameters["chat_completions"] = append(
		payload.ModelRuntimeProfiles[0].Profile.Parameters["chat_completions"], "thinking",
	)
	payload.Body = []byte(`{"model":"glm-5.3-flash","messages":[{"role":"user","content":"hello"}],"thinking":{"type":"disabled"}}`)

	if _, _, err := validateZAIDispatch(payload); err == nil || !strings.Contains(err.Error(), "zai_thinking_disabled_unsupported") {
		t.Fatalf("expected disabled-thinking rejection at the adapter layer, got %v", err)
	}

	payload.Body = []byte(`{"model":"glm-5.3-flash","messages":[{"role":"user","content":"hello"}],"thinking":{"type":"enabled"}}`)
	if _, _, err := validateZAIDispatch(payload); err != nil {
		t.Fatalf("enabled thinking rejected for an always-enabled-thinking model: %v", err)
	}
}

func TestZAIAdapterRoutesGLM53ToChatCompletions(t *testing.T) {
	payload := zaiTestPayload("glm-5.3", "chat_completions")
	payload.ModelRuntimeProfiles[0].Profile.Thinking = platform.ModelRuntimeThinkingProfile{
		DefaultMode: "adaptive", AllowedModes: []string{"adaptive"}, AlwaysEnabled: true,
		EffortValues: []string{"low", "high", "max"},
	}
	payload.ModelRuntimeProfiles[0].Profile.Parameters["chat_completions"] = append(
		payload.ModelRuntimeProfiles[0].Profile.Parameters["chat_completions"], "thinking", "reasoning_effort",
	)
	payload.Body = []byte(`{"model":"glm-5.3","messages":[{"role":"user","content":"hello"}],"thinking":{"type":"enabled"},"reasoning_effort":"low"}`)

	endpoint, encoded, err := validateZAIDispatch(payload)
	if err != nil {
		t.Fatalf("validate glm-5.3 dispatch: %v", err)
	}

	if endpoint != "chat_completions" {
		t.Fatalf("endpoint = %q", endpoint)
	}

	if !strings.Contains(string(encoded), `"reasoning_effort":"low"`) {
		t.Fatalf("glm-5.3 low reasoning_effort was not preserved: %s", encoded)
	}

	payload.Body = []byte(`{"model":"glm-5.3","messages":[{"role":"user","content":"hello"}],"thinking":{"type":"disabled"}}`)
	if _, _, err := validateZAIDispatch(payload); err == nil || !strings.Contains(err.Error(), "zai_thinking_disabled_unsupported") {
		t.Fatalf("expected disabled-thinking rejection for glm-5.3, got %v", err)
	}
}

func TestZAIThinkingValueRespectsAlwaysEnabledModels(t *testing.T) {
	toggleable := platform.ModelRuntimeProfile{
		Thinking: platform.ModelRuntimeThinkingProfile{
			DefaultMode: "adaptive", AllowedModes: []string{"adaptive", "disabled"},
		},
	}
	alwaysEnabled := platform.ModelRuntimeProfile{
		Thinking: platform.ModelRuntimeThinkingProfile{
			DefaultMode: "adaptive", AllowedModes: []string{"adaptive"}, AlwaysEnabled: true,
		},
	}

	if err := validateZAIThinkingValue(toggleable, map[string]json.RawMessage{
		"thinking": json.RawMessage(`{"type":"disabled"}`),
	}); err != nil {
		t.Fatalf("thinking.type=disabled rejected for a toggleable Z.ai model: %v", err)
	}

	if err := validateZAIThinkingValue(alwaysEnabled, map[string]json.RawMessage{
		"thinking": json.RawMessage(`{"type":"enabled"}`),
	}); err != nil {
		t.Fatalf("thinking.type=enabled rejected for an always-enabled-thinking Z.ai model: %v", err)
	}

	if err := validateZAIThinkingValue(alwaysEnabled, map[string]json.RawMessage{
		"thinking": json.RawMessage(`{"type":"disabled"}`),
	}); err == nil {
		t.Fatal("thinking.type=disabled was accepted for an always-enabled-thinking Z.ai model")
	}

	if err := validateZAIThinkingValue(alwaysEnabled, map[string]json.RawMessage{}); err != nil {
		t.Fatalf("omitted thinking was rejected for an always-enabled-thinking Z.ai model: %v", err)
	}
}

func TestZAIAdapterAllowsBYOKGatewayHosts(t *testing.T) {
	for _, raw := range []string{
		"https://api.z.ai/api/paas/v4",
		"https://my-gateway.example.com/zai/v4",
		"https://api.z.ai.evil.example/api/paas/v4",
	} {
		if err := validateZAIBaseURL(raw); err != nil {
			t.Fatalf("BYOK API base rejected: %s: %v", raw, err)
		}
	}
}

func TestZAIReasoningEffortCompatibilityMapping(t *testing.T) {
	toggleable := platform.ModelRuntimeProfile{
		Thinking: platform.ModelRuntimeThinkingProfile{
			DefaultMode: "adaptive", AllowedModes: []string{"adaptive", "disabled"},
			EffortValues: []string{"none", "minimal", "low", "medium", "high", "xhigh", "max"},
		},
	}

	body := map[string]json.RawMessage{"reasoning_effort": json.RawMessage(`"minimal"`)}
	normalizeZAIReasoningEffort(toggleable, body)

	if _, exists := body["reasoning_effort"]; exists || string(body["thinking"]) != `{"type":"disabled"}` {
		t.Fatalf("minimal effort mapping = %s", body)
	}

	body = map[string]json.RawMessage{"reasoning_effort": json.RawMessage(`"low"`)}
	normalizeZAIReasoningEffort(toggleable, body)

	if string(body["reasoning_effort"]) != `"high"` {
		t.Fatalf("low effort mapping for glm-5.2 = %s", body["reasoning_effort"])
	}

	body = map[string]json.RawMessage{"reasoning_effort": json.RawMessage(`"medium"`)}
	normalizeZAIReasoningEffort(toggleable, body)

	if string(body["reasoning_effort"]) != `"high"` {
		t.Fatalf("medium effort mapping = %s", body["reasoning_effort"])
	}

	body = map[string]json.RawMessage{"reasoning_effort": json.RawMessage(`"xhigh"`)}
	normalizeZAIReasoningEffort(toggleable, body)

	if string(body["reasoning_effort"]) != `"max"` {
		t.Fatalf("xhigh effort mapping = %s", body["reasoning_effort"])
	}
}

// TestZAIReasoningEffortPreservesLowForGLM53 proves that glm-5.3's
// documented low/high/max reasoning_effort enum is not silently collapsed
// the way the other zai chat models' wider effort vocabulary is: "low" is a
// genuinely distinct, natively supported tier for this model alone.
func TestZAIReasoningEffortPreservesLowForGLM53(t *testing.T) {
	alwaysEnabled := platform.ModelRuntimeProfile{
		Thinking: platform.ModelRuntimeThinkingProfile{
			DefaultMode: "adaptive", AllowedModes: []string{"adaptive"}, AlwaysEnabled: true,
			EffortValues: []string{"low", "high", "max"},
		},
	}

	body := map[string]json.RawMessage{"reasoning_effort": json.RawMessage(`"low"`)}
	normalizeZAIReasoningEffort(alwaysEnabled, body)

	if string(body["reasoning_effort"]) != `"low"` {
		t.Fatalf("glm-5.3 low effort was rewritten: %s", body["reasoning_effort"])
	}

	body = map[string]json.RawMessage{"reasoning_effort": json.RawMessage(`"high"`)}
	normalizeZAIReasoningEffort(alwaysEnabled, body)

	if string(body["reasoning_effort"]) != `"high"` {
		t.Fatalf("glm-5.3 high effort was rewritten: %s", body["reasoning_effort"])
	}
}

func TestZAIReasoningEffortNeverDisablesAlwaysThinkingModels(t *testing.T) {
	alwaysEnabled := platform.ModelRuntimeProfile{
		Thinking: platform.ModelRuntimeThinkingProfile{
			DefaultMode: "adaptive", AllowedModes: []string{"adaptive"}, AlwaysEnabled: true,
		},
	}

	body := map[string]json.RawMessage{"reasoning_effort": json.RawMessage(`"minimal"`)}
	normalizeZAIReasoningEffort(alwaysEnabled, body)

	if _, exists := body["thinking"]; exists {
		t.Fatalf("glm-5.3-flash-style effort mapping injected a disabled thinking control: %s", body)
	}

	if _, exists := body["reasoning_effort"]; exists {
		t.Fatalf("glm-5.3-flash-style unsupported low-effort alias was kept: %s", body)
	}
}

func TestZAIChatContractRejectsUnsupportedStructuredAndToolControls(t *testing.T) {
	body := map[string]json.RawMessage{
		"response_format": json.RawMessage(`{"type":"json_schema","json_schema":{}}`),
	}
	if err := validateZAIResponseFormat(body); err == nil {
		t.Fatal("Z.ai JSON Schema response format was accepted")
	}

	body = map[string]json.RawMessage{
		"stream":      json.RawMessage(`false`),
		"tool_stream": json.RawMessage(`true`),
		"tools":       json.RawMessage(`[{"type":"function","function":{"name":"lookup","description":"Lookup","parameters":{"type":"object"}}}]`),
	}
	if err := validateZAIToolStream(body); err == nil {
		t.Fatal("tool_stream without stream=true was accepted")
	}

	body["stream"] = json.RawMessage(`true`)
	if err := validateZAIToolStream(body); err != nil {
		t.Fatalf("valid streamed tool request rejected: %v", err)
	}

	if err := validateZAITools(body); err != nil {
		t.Fatalf("valid Z.ai tool rejected: %v", err)
	}
}

func TestZAIUserIDIsOpaqueAndTenantIsolated(t *testing.T) {
	bodyA := map[string]json.RawMessage{"user_id": json.RawMessage(`"person@example.com"`)}
	bodyB := map[string]json.RawMessage{"user_id": json.RawMessage(`"person@example.com"`)}
	payloadA := zaiTestPayload("glm-5.2", "chat_completions")
	payloadA.TenantID = "tenant-a"
	payloadA.Deployment.ProviderAccountID = "shared-account"
	payloadB := payloadA
	payloadB.TenantID = "tenant-b"

	normalizeZAIUserID(payloadA, bodyA, true)
	normalizeZAIUserID(payloadB, bodyB, true)

	var userA, userB string
	if err := json.Unmarshal(bodyA["user_id"], &userA); err != nil {
		t.Fatalf("decode tenant A user ID: %v", err)
	}

	if err := json.Unmarshal(bodyB["user_id"], &userB); err != nil {
		t.Fatalf("decode tenant B user ID: %v", err)
	}

	if len(userA) != 64 || len(userB) != 64 || !strings.HasPrefix(userA, "gdr_") || !strings.HasPrefix(userB, "gdr_") {
		t.Fatalf("unexpected opaque user IDs: %q %q", userA, userB)
	}

	if userA == userB || strings.Contains(userA, "person") || strings.Contains(userB, "person") {
		t.Fatalf("user IDs are not tenant-isolated and opaque: %q %q", userA, userB)
	}
}

func zaiTestPayload(model, endpoint string) platform.TunnelRequest {
	profile := platform.ModelRuntimeProfile{
		Version:    1,
		Endpoints:  []string{endpoint},
		Transports: []string{"http", "sse"},
		Lifecycle:  "active",
		Thinking:   platform.ModelRuntimeThinkingProfile{DefaultMode: "disabled", AllowedModes: []string{"disabled"}},
		Sampling:   platform.ModelRuntimeSamplingProfile{},
		Tools:      platform.ModelRuntimeToolsProfile{},
		Modalities: platform.ModelRuntimeModalities{Input: []string{"text"}, Output: []string{"text"}},
		Parameters: map[string][]string{
			endpoint: {"model", "messages", "stream", "file", "start_page_id", "end_page_id"},
		},
		Limits: platform.ModelRuntimeLimits{MessagesBytes: 1, BatchBytes: 1, FileBytes: 1},
	}

	return platform.TunnelRequest{
		Operation:  platform.ProviderOperationChatCompletions,
		Deployment: platform.Deployment{Provider: "zai", Model: model, APIKey: "zai-secret"},
		ModelRuntimeProfiles: []platform.ModelRuntimeProfileBinding{{
			Provider: "zai", CatalogRevision: "test", CanonicalID: "zai/" + model,
			Slug: model, ProfileVersion: 1, Profile: profile, EffectiveFrom: time.Now(),
		}},
	}
}
