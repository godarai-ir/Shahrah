package server

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func TestHealth(t *testing.T) {
	t.Parallel()

	handler := NewHandler(nil)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"service":"silk-road"`) {
		t.Fatalf("unexpected health response body: %q", rec.Body.String())
	}
}

func TestDispatchForwardsToDeploymentAPIBase(t *testing.T) {
	t.Parallel()

	var gotAuth string
	var gotAccept string
	var gotBody string
	var gotPath string

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotAccept = r.Header.Get("Accept")
		gotPath = r.URL.Path
		defer r.Body.Close()
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		gotBody = string(bodyBytes)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"chatcmpl_1","object":"chat.completion","choices":[]}`))
	}))
	defer upstream.Close()

	handler := NewHandler(nil)
	payload := platform.TunnelRequest{
		Stream: true,
		Deployment: platform.Deployment{
			Provider: "openai",
			APIBase:  upstream.URL,
			APIKey:   "provider-secret",
		},
		Body: []byte(`{"model":"gpt-4o-mini","messages":[{"role":"user","content":"hi"}]}`),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/dispatch", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if gotPath != "/chat/completions" {
		t.Fatalf("expected path /chat/completions, got %q", gotPath)
	}
	if gotAuth != "Bearer provider-secret" {
		t.Fatalf("unexpected Authorization header: %q", gotAuth)
	}
	if gotAccept != "text/event-stream" {
		t.Fatalf("unexpected Accept header: %q", gotAccept)
	}
	if !strings.Contains(gotBody, `"model":"gpt-4o-mini"`) {
		t.Fatalf("unexpected forwarded body: %q", gotBody)
	}
}

func TestDispatchUsesOpenAIEndpointOverride(t *testing.T) {
	t.Parallel()

	var gotPath string

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"resp_1","object":"response","output":[]}`))
	}))
	defer upstream.Close()

	handler := NewHandler(nil)
	payload := platform.TunnelRequest{
		Deployment: platform.Deployment{
			Provider: "openai",
			APIBase:  upstream.URL,
			APIKey:   "provider-secret",
		},
		RequestMetadata: map[string]interface{}{
			"provider_endpoint_path": "/responses",
		},
		Body: []byte(`{"model":"gpt-4o","input":"hi"}`),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/dispatch", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if gotPath != "/responses" {
		t.Fatalf("expected path /responses, got %q", gotPath)
	}
}

func TestDispatchUsesAnthropicAdapterForSync(t *testing.T) {
	t.Parallel()

	var gotPath string
	var gotAPIKey string
	var gotVersion string
	var body map[string]any

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAPIKey = r.Header.Get("x-api-key")
		gotVersion = r.Header.Get("anthropic-version")
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"id":"msg_123",
			"type":"message",
			"role":"assistant",
			"model":"claude-3-haiku-20240307",
			"content":[{"type":"text","text":"سلام"}],
			"usage":{"input_tokens":7,"output_tokens":3},
			"stop_reason":"end_turn"
		}`))
	}))
	defer upstream.Close()

	handler := NewHandler(nil)
	payload := platform.TunnelRequest{
		Deployment: platform.Deployment{
			Provider: "anthropic",
			APIBase:  upstream.URL,
			APIKey:   "sk-ant-test",
		},
		Body: []byte(`{"model":"claude-3-haiku-20240307","messages":[{"role":"system","content":"be brief"},{"role":"user","content":"hi"}]}`),
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/dispatch", strings.NewReader(string(raw)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}
	if gotPath != "/v1/messages" {
		t.Fatalf("expected anthropic path /v1/messages, got %q", gotPath)
	}
	if gotAPIKey != "sk-ant-test" {
		t.Fatalf("unexpected x-api-key header: %q", gotAPIKey)
	}
	if gotVersion != anthropicDefaultVersion {
		t.Fatalf("unexpected anthropic-version: %q", gotVersion)
	}
	if body["model"] != "claude-3-haiku-20240307" {
		t.Fatalf("unexpected upstream model: %#v", body["model"])
	}

	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if out["object"] != "chat.completion" {
		t.Fatalf("unexpected response object: %#v", out["object"])
	}
}

func TestDispatchUsesAnthropicAdapterForStream(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(strings.Join([]string{
			"event: message_start",
			`data: {"type":"message_start","message":{"id":"msg_123","model":"claude-3-haiku-20240307"}}`,
			"",
			"event: content_block_delta",
			`data: {"type":"content_block_delta","index":0,"delta":{"text":"Salam"}}`,
			"",
			"event: message_delta",
			`data: {"type":"message_delta","delta":{"stop_reason":"end_turn"}}`,
			"",
		}, "\n")))
	}))
	defer upstream.Close()

	handler := NewHandler(nil)
	payload := platform.TunnelRequest{
		Stream: true,
		Deployment: platform.Deployment{
			Provider: "anthropic",
			APIBase:  upstream.URL,
			APIKey:   "sk-ant-test",
		},
		Body: []byte(`{"model":"claude-3-haiku-20240307","messages":[{"role":"user","content":"hi"}],"stream":true}`),
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/dispatch", strings.NewReader(string(raw)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("unexpected content type: %q", ct)
	}

	streamText := rec.Body.String()
	if !strings.Contains(streamText, `"object":"chat.completion.chunk"`) {
		t.Fatalf("expected openai stream chunk, got %q", streamText)
	}
	if !strings.Contains(streamText, `"content":"Salam"`) {
		t.Fatalf("expected text delta in stream, got %q", streamText)
	}
	if !strings.Contains(streamText, "data: [DONE]") {
		t.Fatalf("expected stream terminator, got %q", streamText)
	}

	scanner := bufio.NewScanner(strings.NewReader(streamText))
	found := false
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), `"finish_reason":"end_turn"`) {
			found = true
			break
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan stream: %v", err)
	}
	if !found {
		t.Fatalf("expected finish reason in stream, got %q", streamText)
	}
}

func TestCanonicalAnthropicModel(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"claude-3-haiku-20240307":                                            "claude-3-haiku-20240307",
		"anthropic/claude-3-haiku-20240307":                                  "claude-3-haiku-20240307",
		"anthropic:anthropic-sunny-zephyr:anthropic/claude-3-haiku-20240307": "claude-3-haiku-20240307",
		"anthropic:anthropic-sunny-zephyr:claude-3-haiku-20240307":           "claude-3-haiku-20240307",
	}

	for input, want := range cases {
		if got := canonicalAnthropicModel(input); got != want {
			t.Fatalf("canonicalAnthropicModel(%q) = %q, want %q", input, got, want)
		}
	}
}
