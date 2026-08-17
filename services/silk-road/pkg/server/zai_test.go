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

	payload.Deployment.APIBase = "https://api.z.ai.evil.example/api/paas/v4"
	if _, _, err := validateZAIDispatch(payload); err == nil || !strings.Contains(err.Error(), "invalid_api_base") {
		t.Fatalf("expected unsafe base error, got %v", err)
	}
}

func TestZAIReasoningEffortCompatibilityMapping(t *testing.T) {
	body := map[string]json.RawMessage{"reasoning_effort": json.RawMessage(`"minimal"`)}
	normalizeZAIReasoningEffort(body)

	if _, exists := body["reasoning_effort"]; exists || string(body["thinking"]) != `{"type":"disabled"}` {
		t.Fatalf("minimal effort mapping = %s", body)
	}

	body = map[string]json.RawMessage{"reasoning_effort": json.RawMessage(`"medium"`)}
	normalizeZAIReasoningEffort(body)

	if string(body["reasoning_effort"]) != `"high"` {
		t.Fatalf("medium effort mapping = %s", body["reasoning_effort"])
	}

	body = map[string]json.RawMessage{"reasoning_effort": json.RawMessage(`"xhigh"`)}
	normalizeZAIReasoningEffort(body)

	if string(body["reasoning_effort"]) != `"max"` {
		t.Fatalf("xhigh effort mapping = %s", body["reasoning_effort"])
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
