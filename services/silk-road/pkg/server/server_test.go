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

func TestDispatchForwardsOpenAICompatibleToolsPayload(t *testing.T) {
	t.Parallel()

	var gotBody string

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		gotBody = string(bodyBytes)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"id":"chatcmpl_tools",
			"object":"chat.completion",
			"created":1234567890,
			"model":"gpt-4o-mini",
			"choices":[{"index":0,"message":{"role":"assistant","tool_calls":[{"id":"call_123","type":"function","function":{"name":"get_order_status","arguments":"{\"order_id\":\"A-1024\"}"}}]},"finish_reason":"tool_calls"}],
			"usage":{"prompt_tokens":20,"completion_tokens":8,"total_tokens":28}
		}`))
	}))
	defer upstream.Close()

	handler := NewHandler(nil)
	payload := platform.TunnelRequest{
		Deployment: platform.Deployment{
			Provider: "openai",
			APIBase:  upstream.URL,
			APIKey:   "provider-secret",
		},
		Body: []byte(`{"model":"gpt-4o-mini","parallel_tool_calls":false,"messages":[{"role":"user","content":"status?"},{"role":"tool","tool_call_id":"call_prev","content":"{\"status\":\"delivered\"}"}],"response_format":{"type":"json_schema","json_schema":{"name":"order_status","strict":true,"schema":{"type":"object","properties":{"status":{"type":"string"}},"required":["status"],"additionalProperties":false}}},"tools":[{"type":"function","function":{"name":"get_order_status","strict":true,"parameters":{"type":"object"}}}],"tool_choice":{"type":"function","function":{"name":"get_order_status"}}}`),
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
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}
	for _, want := range []string{`"tools"`, `"tool_choice"`, `"tool_call_id":"call_prev"`, `"parallel_tool_calls":false`, `"response_format"`, `"type":"json_schema"`, `"strict":true`} {
		if !strings.Contains(gotBody, want) {
			t.Fatalf("forwarded body missing %s: %s", want, gotBody)
		}
	}
	if !strings.Contains(rec.Body.String(), `"tool_calls"`) || !strings.Contains(rec.Body.String(), `"finish_reason":"tool_calls"`) {
		t.Fatalf("unexpected response: %s", rec.Body.String())
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

func TestDispatchMockProviderDisabledByDefault(t *testing.T) {
	t.Setenv("SILK_ROAD_ENABLE_MOCK_PROVIDER", "")

	handler := NewHandler(nil)
	payload := platform.TunnelRequest{
		Deployment: platform.Deployment{Provider: "mock", Model: "mock-model"},
		Body:       []byte(`{"model":"gpt-4o-mini","messages":[{"role":"user","content":"hi"}]}`),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/dispatch", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d with body %q", http.StatusBadRequest, rec.Code, rec.Body.String())
	}
}

func TestDispatchMockProviderEnabled(t *testing.T) {
	t.Setenv("SILK_ROAD_ENABLE_MOCK_PROVIDER", "true")
	t.Setenv("SILK_ROAD_MOCK_LATENCY_MS", "0")

	handler := NewHandler(nil)
	payload := platform.TunnelRequest{
		Deployment: platform.Deployment{Provider: "mock", Model: "mock-model"},
		Body:       []byte(`{"model":"gpt-4o-mini","messages":[{"role":"user","content":"hi"}]}`),
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
		t.Fatalf("expected status %d, got %d with body %q", http.StatusOK, rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("X-Shahrah-Mock-Provider"); got != "true" {
		t.Fatalf("expected mock header, got %q", got)
	}
	if !strings.Contains(rec.Body.String(), `"object":"chat.completion"`) {
		t.Fatalf("unexpected mock response: %q", rec.Body.String())
	}
}

func TestDispatchMockProviderEmbeddings(t *testing.T) {
	t.Setenv("SILK_ROAD_ENABLE_MOCK_PROVIDER", "true")
	t.Setenv("SILK_ROAD_MOCK_LATENCY_MS", "0")

	handler := NewHandler(nil)
	payload := platform.TunnelRequest{
		Deployment: platform.Deployment{Provider: "mock", Model: "mock-embedding"},
		RequestMetadata: map[string]interface{}{
			"provider_endpoint_path": "/embeddings",
		},
		Body: []byte(`{"model":"text-embedding-3-small","input":["one","two"]}`),
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
		t.Fatalf("expected status %d, got %d with body %q", http.StatusOK, rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"object":"list"`) {
		t.Fatalf("unexpected mock response: %q", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"index":1`) {
		t.Fatalf("expected two embeddings, got %q", rec.Body.String())
	}
}

func TestDispatchOpenAICompatiblePreservesJSONContentTypeForImages(t *testing.T) {
	t.Parallel()

	var gotContentType string
	var gotPath string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"created":1,"data":[]}`))
	}))
	defer upstream.Close()

	handler := NewHandler(nil)
	payload := platform.TunnelRequest{
		Deployment: platform.Deployment{Provider: "xai", APIBase: upstream.URL, APIKey: "provider-secret"},
		Headers:    map[string]string{"Content-Type": "application/json"},
		RequestMetadata: map[string]interface{}{
			"provider_endpoint_path": "/images/generations",
		},
		Body: []byte(`{"model":"grok-imagine-image","prompt":"hi"}`),
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
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}
	if gotPath != "/images/generations" {
		t.Fatalf("expected image path, got %q", gotPath)
	}
	if gotContentType != "application/json" {
		t.Fatalf("unexpected content type: %q", gotContentType)
	}
}

func TestDispatchOpenAICompatiblePreservesMultipartContentTypeForImages(t *testing.T) {
	t.Parallel()

	const multipartType = "multipart/form-data; boundary=shahrah-test-boundary"
	var gotContentType string
	var gotPath string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"created":1,"data":[]}`))
	}))
	defer upstream.Close()

	handler := NewHandler(nil)
	payload := platform.TunnelRequest{
		Deployment: platform.Deployment{Provider: "openai", APIBase: upstream.URL, APIKey: "provider-secret"},
		Headers:    map[string]string{"Content-Type": multipartType},
		RequestMetadata: map[string]interface{}{
			"provider_endpoint_path": "/images/edits",
		},
		Body: []byte("--shahrah-test-boundary\r\nContent-Disposition: form-data; name=\"model\"\r\n\r\ngpt-image-1\r\n--shahrah-test-boundary--\r\n"),
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
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}
	if gotPath != "/images/edits" {
		t.Fatalf("expected image path, got %q", gotPath)
	}
	if gotContentType != multipartType {
		t.Fatalf("unexpected content type: %q", gotContentType)
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

func TestDispatchAnthropicAdapterConvertsImageDataURL(t *testing.T) {
	t.Parallel()

	var body map[string]any
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
			"content":[{"type":"text","text":"small icon"}],
			"usage":{"input_tokens":12,"output_tokens":2},
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
		Body: []byte(`{"model":"claude-3-haiku-20240307","messages":[{"role":"user","content":[{"type":"text","text":"describe"},{"type":"image_url","image_url":{"url":"data:image/png;base64,AAAA"}}]}]}`),
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
	messages := body["messages"].([]any)
	content := messages[0].(map[string]any)["content"].([]any)
	image := content[1].(map[string]any)
	source := image["source"].(map[string]any)
	if image["type"] != "image" || source["type"] != "base64" || source["media_type"] != "image/png" || source["data"] != "AAAA" {
		t.Fatalf("unexpected anthropic image block: %#v", image)
	}
}

func TestDispatchAnthropicAdapterConvertsPDFFileDataURL(t *testing.T) {
	t.Parallel()

	var body map[string]any
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
			"content":[{"type":"text","text":"short guide"}],
			"usage":{"input_tokens":20,"output_tokens":3},
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
		Body: []byte(`{"model":"claude-3-haiku-20240307","messages":[{"role":"user","content":[{"type":"text","text":"summarize"},{"type":"file","file":{"filename":"guide.pdf","file_data":"data:application/pdf;base64,JVBERi0x"}}]}]}`),
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
	messages := body["messages"].([]any)
	content := messages[0].(map[string]any)["content"].([]any)
	document := content[1].(map[string]any)
	source := document["source"].(map[string]any)
	if document["type"] != "document" || source["type"] != "base64" || source["media_type"] != "application/pdf" || source["data"] != "JVBERi0x" {
		t.Fatalf("unexpected anthropic document block: %#v", document)
	}
}

func TestDispatchGeminiAdapterConvertsImageDataURL(t *testing.T) {
	t.Parallel()

	var gotPath string
	var body map[string]any
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"candidates":[{"content":{"role":"model","parts":[{"text":"small icon"}]},"finishReason":"STOP"}],
			"usageMetadata":{"promptTokenCount":12,"candidatesTokenCount":2,"totalTokenCount":14}
		}`))
	}))
	defer upstream.Close()

	handler := NewHandler(nil)
	payload := platform.TunnelRequest{
		Deployment: platform.Deployment{
			Provider: "gemini",
			APIBase:  upstream.URL,
			APIKey:   "gemini-key",
		},
		Body: []byte(`{"model":"gemini/gemini-1.5-flash","messages":[{"role":"user","content":[{"type":"text","text":"describe"},{"type":"image_url","image_url":{"url":"data:image/png;base64,AAAA"}}]}]}`),
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
	if gotPath != "/models/gemini-1.5-flash:generateContent" {
		t.Fatalf("unexpected gemini path: %q", gotPath)
	}
	contents := body["contents"].([]any)
	parts := contents[0].(map[string]any)["parts"].([]any)
	inlineData := parts[1].(map[string]any)["inline_data"].(map[string]any)
	if inlineData["mime_type"] != "image/png" || inlineData["data"] != "AAAA" {
		t.Fatalf("unexpected gemini inline data: %#v", inlineData)
	}
	if !strings.Contains(rec.Body.String(), `"object":"chat.completion"`) {
		t.Fatalf("unexpected response: %s", rec.Body.String())
	}
}

func TestDispatchGeminiAdapterConvertsPDFFileDataURL(t *testing.T) {
	t.Parallel()

	var gotPath string
	var body map[string]any
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"candidates":[{"content":{"role":"model","parts":[{"text":"short guide"}]},"finishReason":"STOP"}],
			"usageMetadata":{"promptTokenCount":20,"candidatesTokenCount":3,"totalTokenCount":23}
		}`))
	}))
	defer upstream.Close()

	handler := NewHandler(nil)
	payload := platform.TunnelRequest{
		Deployment: platform.Deployment{
			Provider: "gemini",
			APIBase:  upstream.URL,
			APIKey:   "gemini-key",
		},
		Body: []byte(`{"model":"gemini/gemini-1.5-flash","messages":[{"role":"user","content":[{"type":"text","text":"summarize"},{"type":"file","file":{"filename":"guide.pdf","file_data":"data:application/pdf;base64,JVBERi0x"}}]}]}`),
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
	if gotPath != "/models/gemini-1.5-flash:generateContent" {
		t.Fatalf("unexpected gemini path: %q", gotPath)
	}
	contents := body["contents"].([]any)
	parts := contents[0].(map[string]any)["parts"].([]any)
	inlineData := parts[1].(map[string]any)["inline_data"].(map[string]any)
	if inlineData["mime_type"] != "application/pdf" || inlineData["data"] != "JVBERi0x" {
		t.Fatalf("unexpected gemini inline data: %#v", inlineData)
	}
	if !strings.Contains(rec.Body.String(), `"object":"chat.completion"`) {
		t.Fatalf("unexpected response: %s", rec.Body.String())
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

func TestDispatchStreamingFlushesSSE(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("response writer does not support flushing")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("data: hello\n\n"))
		flusher.Flush()
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
		RequestMetadata: map[string]interface{}{
			"provider_endpoint_path": "/responses",
		},
		Body: []byte(`{"model":"gpt-4o-mini","input":"hello","stream":true}`),
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
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}
	if !rec.Flushed {
		t.Fatal("expected streaming response to flush SSE data")
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("unexpected content type: %q", ct)
	}
	if !strings.Contains(rec.Body.String(), "data: hello") {
		t.Fatalf("expected stream body to be relayed, got %q", rec.Body.String())
	}
}

func TestDispatchOpenAIResponsesStreamingRelaysMultipleEventsAndDone(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/responses" {
			t.Fatalf("expected /responses path, got %q", got)
		}
		if got := r.Header.Get("Accept"); got != "text/event-stream" {
			t.Fatalf("unexpected Accept header: %q", got)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("response writer does not support flushing")
		}
		w.WriteHeader(http.StatusOK)
		for _, frame := range []string{
			"event: response.created\n",
			"data: {\"type\":\"response.created\"}\n\n",
			"data: [DONE]\n\n",
		} {
			_, _ = w.Write([]byte(frame))
			flusher.Flush()
		}
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
		RequestMetadata: map[string]interface{}{
			"provider_endpoint_path": "/responses",
		},
		Body: []byte(`{"model":"gpt-4o-mini","input":"hello","stream":true}`),
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
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}
	bodyText := rec.Body.String()
	if !strings.Contains(bodyText, "event: response.created") || !strings.Contains(bodyText, "data: [DONE]") {
		t.Fatalf("expected complete SSE relay, got %q", bodyText)
	}
	if !rec.Flushed {
		t.Fatal("expected streaming response to flush SSE data")
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
