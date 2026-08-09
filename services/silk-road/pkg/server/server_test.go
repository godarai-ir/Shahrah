package server

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func TestHealth(t *testing.T) {
	t.Parallel()

	handler := NewHandler(nil)
	req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if !strings.Contains(rec.Body.String(), `"service":"silk-road"`) {
		t.Fatalf("unexpected health response body: %q", rec.Body.String())
	}
}

func TestLiveAndReady(t *testing.T) {
	t.Parallel()

	handler := NewHandler(nil)

	for _, path := range []string{"/live", "/ready"} {
		req := httptest.NewRequest(http.MethodGet, path, http.NoBody)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("%s: expected status %d, got %d", path, http.StatusOK, rec.Code)
		}
	}
}

func TestMetricsEndpointReportsRequests(t *testing.T) {
	t.Parallel()

	handler := NewHandler(nil)
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/live", http.NoBody))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", http.NoBody))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "shahrah_transport_http_requests_total") ||
		!strings.Contains(body, `service="silk-road"`) ||
		!strings.Contains(body, `route="live"`) {
		t.Fatalf("expected silk-road request metrics, got %q", body)
	}
}

func TestDispatchForwardsToDeploymentAPIBase(t *testing.T) {
	t.Parallel()

	var (
		gotAuth   string
		gotAccept string
		gotBody   string
		gotPath   string
	)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotAccept = r.Header.Get("Accept")

		gotPath = r.URL.Path
		defer r.Body.Close()

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
			return
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
			t.Errorf("read body: %v", err)
			return
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
		Body: []byte(`{
			"model":"gpt-4o-mini",
			"parallel_tool_calls":false,
			"messages":[
				{"role":"user","content":"status?"},
				{"role":"tool","tool_call_id":"call_prev","content":"{\"status\":\"delivered\"}"}
			],
			"response_format":{"type":"json_schema","json_schema":{"name":"order_status","strict":true,"schema":{"type":"object","properties":{"status":{"type":"string"}},"required":["status"],"additionalProperties":false}}},
			"tools":[{"type":"function","function":{"name":"get_order_status","strict":true,"parameters":{"type":"object"}}}],
			"tool_choice":{"type":"function","function":{"name":"get_order_status"}}
		}`),
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

	var (
		gotPath            string
		gotClientRequestID string
	)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotClientRequestID = r.Header.Get("X-Client-Request-Id")
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
		Headers: map[string]string{
			"X-Client-Request-Id": "client_req_1",
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

	if gotClientRequestID != "client_req_1" {
		t.Fatalf("expected X-Client-Request-Id to be forwarded, got %q", gotClientRequestID)
	}
}

func TestDispatchUsesCohereRerankAdapter(t *testing.T) {
	t.Parallel()

	var (
		gotAuth        string
		gotContentType string
		gotPath        string
		gotBody        string
	)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotContentType = r.Header.Get("Content-Type")

		gotPath = r.URL.Path
		defer r.Body.Close()

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
			return
		}

		gotBody = string(bodyBytes)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"rerank_1","results":[{"index":1,"relevance_score":0.99}]}`))
	}))
	defer upstream.Close()

	handler := NewHandler(nil)
	payload := platform.TunnelRequest{
		Deployment: platform.Deployment{
			Provider: "cohere",
			APIBase:  upstream.URL,
			APIKey:   "cohere-key",
		},
		Headers: map[string]string{"Content-Type": "application/json"},
		RequestMetadata: map[string]interface{}{
			"provider_endpoint_path": "/v2/rerank",
		},
		Body: []byte(`{"model":"rerank-multilingual-v3.0","query":"q","documents":["a","b"],"top_n":1}`),
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

	if gotPath != "/v2/rerank" {
		t.Fatalf("expected Cohere rerank path, got %q", gotPath)
	}

	if gotAuth != "Bearer cohere-key" {
		t.Fatalf("unexpected Authorization header: %q", gotAuth)
	}

	if gotContentType != "application/json" {
		t.Fatalf("unexpected content type: %q", gotContentType)
	}

	for _, want := range []string{`"model":"rerank-multilingual-v3.0"`, `"query":"q"`, `"documents":["a","b"]`, `"top_n":1`} {
		if !strings.Contains(gotBody, want) {
			t.Fatalf("forwarded body missing %s: %s", want, gotBody)
		}
	}

	if !strings.Contains(rec.Body.String(), `"results"`) || !strings.Contains(rec.Body.String(), `"relevance_score":0.99`) {
		t.Fatalf("unexpected response: %s", rec.Body.String())
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
	t.Setenv("SILK_ROAD_MOCK_ALLOWED_TENANT_IDS", "tenant-loadtest")
	t.Setenv("SILK_ROAD_MOCK_ALLOWED_DEPLOYMENT_IDS", "deployment-loadtest")

	upstreamCalls := 0
	handler := NewHandler(&http.Client{
		Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			upstreamCalls++
			return nil, fmt.Errorf("mock provider attempted an upstream request")
		}),
	})
	payload := platform.TunnelRequest{
		TenantID:   "tenant-loadtest",
		Deployment: platform.Deployment{ID: "deployment-loadtest", Provider: "mock", Model: "mock-model"},
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

	if upstreamCalls != 0 {
		t.Fatalf("expected no upstream requests, got %d", upstreamCalls)
	}
}

func TestDispatchMockProviderFailsClosedOutsideAllowlist(t *testing.T) {
	t.Setenv("SILK_ROAD_ENABLE_MOCK_PROVIDER", "true")
	t.Setenv("SILK_ROAD_MOCK_LATENCY_MS", "0")
	t.Setenv("SILK_ROAD_MOCK_ALLOWED_TENANT_IDS", "tenant-loadtest")
	t.Setenv("SILK_ROAD_MOCK_ALLOWED_DEPLOYMENT_IDS", "deployment-loadtest")

	handler := NewHandler(nil)
	tests := []struct {
		name       string
		tenantID   string
		deployment string
	}{
		{name: "different tenant", tenantID: "tenant-customer", deployment: "deployment-loadtest"},
		{name: "different deployment", tenantID: "tenant-loadtest", deployment: "deployment-customer"},
		{name: "missing tenant", deployment: "deployment-loadtest"},
		{name: "missing deployment", tenantID: "tenant-loadtest"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := platform.TunnelRequest{
				TenantID: tt.tenantID,
				Deployment: platform.Deployment{
					ID:       tt.deployment,
					Provider: "mock",
					Model:    "mock-model",
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

			if rec.Code != http.StatusForbidden {
				t.Fatalf("expected status %d, got %d with body %q", http.StatusForbidden, rec.Code, rec.Body.String())
			}

			if !strings.Contains(rec.Body.String(), `"code":"mock_provider_access_denied"`) {
				t.Fatalf("unexpected mock denial: %q", rec.Body.String())
			}
		})
	}
}

func TestDispatchMockProviderRerank(t *testing.T) {
	t.Setenv("SILK_ROAD_ENABLE_MOCK_PROVIDER", "true")
	t.Setenv("SILK_ROAD_MOCK_LATENCY_MS", "0")
	t.Setenv("SILK_ROAD_MOCK_ALLOWED_TENANT_IDS", "tenant-loadtest")
	t.Setenv("SILK_ROAD_MOCK_ALLOWED_DEPLOYMENT_IDS", "deployment-rerank")

	handler := NewHandler(nil)
	payload := platform.TunnelRequest{
		TenantID:   "tenant-loadtest",
		Deployment: platform.Deployment{ID: "deployment-rerank", Provider: "mock", Model: "mock-rerank"},
		RequestMetadata: map[string]interface{}{
			"provider_endpoint_path": "/rerank",
		},
		Body: []byte(`{"model":"mock-rerank","query":"q","documents":["a","b"],"top_n":1}`),
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
		t.Fatalf("expected status %d, got %d with body %q", http.StatusOK, rec.Code, rec.Body.String())
	}

	if got := rec.Header().Get("X-Shahrah-Mock-Provider"); got != "true" {
		t.Fatalf("expected mock header, got %q", got)
	}

	if !strings.Contains(rec.Body.String(), `"id":"rerank_mock"`) || !strings.Contains(rec.Body.String(), `"results"`) {
		t.Fatalf("unexpected mock rerank response: %q", rec.Body.String())
	}
}

func TestDispatchMockProviderEmbeddings(t *testing.T) {
	t.Setenv("SILK_ROAD_ENABLE_MOCK_PROVIDER", "true")
	t.Setenv("SILK_ROAD_MOCK_LATENCY_MS", "0")
	t.Setenv("SILK_ROAD_MOCK_ALLOWED_TENANT_IDS", "tenant-loadtest")
	t.Setenv("SILK_ROAD_MOCK_ALLOWED_DEPLOYMENT_IDS", "deployment-embedding")

	handler := NewHandler(nil)
	payload := platform.TunnelRequest{
		TenantID:   "tenant-loadtest",
		Deployment: platform.Deployment{ID: "deployment-embedding", Provider: "mock", Model: "mock-embedding"},
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

	var (
		gotContentType string
		gotPath        string
	)

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
		Operation:            platform.ProviderOperationOpenAIImagesGenerations,
		Deployment:           platform.Deployment{Provider: "xai", Model: "grok-imagine-image", APIBase: upstream.URL, APIKey: "provider-secret"},
		ModelRuntimeProfiles: []platform.ModelRuntimeProfileBinding{xaiTestRuntimeProfile("grok-imagine-image")},
		Headers:              map[string]string{"Content-Type": "application/json"},
		Body:                 []byte(`{"model":"grok-imagine-image","prompt":"hi"}`),
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

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestDispatchOpenAICompatiblePreservesMultipartContentTypeForImages(t *testing.T) {
	t.Parallel()

	const multipartType = "multipart/form-data; boundary=shahrah-test-boundary"

	var (
		gotContentType string
		gotPath        string
	)

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

func TestDispatchOpenAICompatibleForwardsAudioTranscriptionsMultipart(t *testing.T) {
	t.Parallel()

	const multipartType = "multipart/form-data; boundary=shahrah-audio-boundary"

	var (
		gotAuth        string
		gotContentType string
		gotPath        string
		gotBody        string
	)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotContentType = r.Header.Get("Content-Type")

		gotPath = r.URL.Path
		defer r.Body.Close()

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
			return
		}

		gotBody = string(bodyBytes)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"text":"سلام"}`))
	}))
	defer upstream.Close()

	handler := NewHandler(nil)
	payload := platform.TunnelRequest{
		Deployment: platform.Deployment{Provider: "deepinfra", APIBase: upstream.URL, APIKey: "provider-secret"},
		Headers:    map[string]string{"Content-Type": multipartType},
		RequestMetadata: map[string]interface{}{
			"provider_endpoint_path": "/audio/transcriptions",
		},
		Body: []byte(strings.Join([]string{
			"--shahrah-audio-boundary",
			`Content-Disposition: form-data; name="model"`,
			"",
			"whisper-large-v3",
			"--shahrah-audio-boundary",
			`Content-Disposition: form-data; name="file"; filename="sample.wav"`,
			"Content-Type: audio/wav",
			"",
			"fake-audio",
			"--shahrah-audio-boundary--",
			"",
		}, "\r\n")),
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

	if gotPath != "/audio/transcriptions" {
		t.Fatalf("expected audio transcription path, got %q", gotPath)
	}

	if gotContentType != multipartType {
		t.Fatalf("unexpected content type: %q", gotContentType)
	}

	if gotAuth != "Bearer provider-secret" {
		t.Fatalf("unexpected Authorization header: %q", gotAuth)
	}

	for _, want := range []string{`name="model"`, "whisper-large-v3", `name="file"; filename="sample.wav"`, "fake-audio"} {
		if !strings.Contains(gotBody, want) {
			t.Fatalf("forwarded body missing %s: %s", want, gotBody)
		}
	}
}

func TestDispatchDeepgramAudioTranscriptionsAdapter(t *testing.T) {
	t.Parallel()

	var (
		gotAuth        string
		gotContentType string
		gotPath        string
		gotModel       string
		gotBody        string
	)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotContentType = r.Header.Get("Content-Type")
		gotPath = r.URL.Path

		gotModel = r.URL.Query().Get("model")
		defer r.Body.Close()

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
			return
		}

		gotBody = string(bodyBytes)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{\"results\":{\"channels\":[{\"alternatives\":[{\"transcript\":\"سلام از دیپ\u200cگرام\"}]}]}}"))
	}))
	defer upstream.Close()

	handler := NewHandler(nil)
	payload := platform.TunnelRequest{
		Deployment: platform.Deployment{Provider: "deepgram", APIBase: upstream.URL, APIKey: "deepgram-secret", Model: "deepgram/nova-3-general"},
		Headers:    map[string]string{"Content-Type": "multipart/form-data; boundary=shahrah-audio-boundary"},
		RequestMetadata: map[string]interface{}{
			"provider_endpoint_path": "/audio/transcriptions",
		},
		Body: audioMultipartPayloadBody("shahrah-audio-boundary", map[string]string{"language": "fa"}, "sample.wav", "audio/wav", "fake-audio"),
	}
	req := dispatchRequest(t, payload)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}

	if gotPath != "/v1/listen" {
		t.Fatalf("expected Deepgram listen path, got %q", gotPath)
	}

	if gotModel != "nova-3-general" {
		t.Fatalf("expected stripped Deepgram model, got %q", gotModel)
	}

	if gotAuth != "Token deepgram-secret" {
		t.Fatalf("unexpected Authorization header: %q", gotAuth)
	}

	if gotContentType != "audio/wav" {
		t.Fatalf("unexpected content type: %q", gotContentType)
	}

	if gotBody != "fake-audio" {
		t.Fatalf("unexpected body: %q", gotBody)
	}

	if !strings.Contains(rec.Body.String(), "\"text\":\"سلام از دیپ\u200cگرام\"") {
		t.Fatalf("expected normalized transcript response, got %s", rec.Body.String())
	}
}

func TestRawDispatchFallsBackForAdapterWithoutStreamingInterface(t *testing.T) {
	t.Parallel()

	receivedBody := ""

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer closeBody(r.Body)

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read provider body: %v", err)
			return
		}

		receivedBody = string(body)

		w.Header().Set("Content-Type", "application/json")
		time.Sleep(50 * time.Millisecond)

		if _, writeErr := w.Write([]byte(`{"results":{"channels":[{"alternatives":[{"transcript":"ok"}]}]}}`)); writeErr != nil {
			t.Errorf("write provider response: %v", writeErr)
		}
	}))
	defer upstream.Close()

	const boundary = "raw-audio-boundary"

	rawBody := audioMultipartPayloadBody(boundary, nil, "sample.wav", "audio/wav", "raw-audio")

	metadata, err := json.Marshal(platform.TunnelRequest{
		Deployment: platform.Deployment{
			Provider: "deepgram",
			APIBase:  upstream.URL,
			APIKey:   "provider-key",
			Model:    "deepgram/nova-3-general",
		},
		Headers: map[string]string{"Content-Type": "multipart/form-data; boundary=" + boundary},
		RequestMetadata: map[string]interface{}{
			"provider_endpoint_path": "/audio/transcriptions",
		},
	})
	if err != nil {
		t.Fatalf("marshal metadata: %v", err)
	}

	var prefix [4]byte
	binary.BigEndian.PutUint32(prefix[:], uint32(len(metadata)))
	req := httptest.NewRequest(
		http.MethodPost,
		"/dispatch/raw",
		io.MultiReader(bytes.NewReader(prefix[:]), bytes.NewReader(metadata), bytes.NewReader(rawBody)),
	)
	req.ContentLength = int64(len(prefix) + len(metadata) + len(rawBody))
	recorder := httptest.NewRecorder()
	NewHandler(&http.Client{Timeout: 10 * time.Millisecond}).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}

	if receivedBody != "raw-audio" {
		t.Fatalf("expected fallback adapter body %q, got %q", "raw-audio", receivedBody)
	}
}

func TestDispatchCohereAudioTranscriptionsAdapter(t *testing.T) {
	t.Parallel()

	var (
		gotAuth            string
		gotPath            string
		gotModel           string
		gotLanguage        string
		gotFile            string
		gotFileContentType string
	)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")

		gotPath = r.URL.Path
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Errorf("parse multipart: %v", err)
			return
		}

		gotModel = r.FormValue("model")
		gotLanguage = r.FormValue("language")

		file, header, err := r.FormFile("file")
		if err != nil {
			t.Errorf("read file: %v", err)
			return
		}

		gotFileContentType = header.Header.Get("Content-Type")

		defer file.Close()

		bodyBytes, err := io.ReadAll(file)
		if err != nil {
			t.Errorf("read file body: %v", err)
			return
		}

		gotFile = string(bodyBytes)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"text":"cohere ok"}`))
	}))
	defer upstream.Close()

	handler := NewHandler(nil)
	payload := platform.TunnelRequest{
		Deployment: platform.Deployment{Provider: "cohere", APIBase: upstream.URL, APIKey: "cohere-secret", Model: "cohere/cohere-transcribe-03-2026"},
		Headers:    map[string]string{"Content-Type": "multipart/form-data; boundary=shahrah-cohere-boundary"},
		RequestMetadata: map[string]interface{}{
			"provider_endpoint_path": "/audio/transcriptions",
		},
		Body: audioMultipartPayloadBody("shahrah-cohere-boundary", map[string]string{"language": "fa"}, "sample.wav", "audio/wav", "fake-audio"),
	}
	req := dispatchRequest(t, payload)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}

	if gotPath != "/v2/audio/transcriptions" {
		t.Fatalf("expected Cohere transcription path, got %q", gotPath)
	}

	if gotAuth != "Bearer cohere-secret" {
		t.Fatalf("unexpected Authorization header: %q", gotAuth)
	}

	if gotModel != "cohere-transcribe-03-2026" {
		t.Fatalf("expected stripped Cohere model, got %q", gotModel)
	}

	if gotLanguage != "fa" {
		t.Fatalf("expected language fa, got %q", gotLanguage)
	}

	if gotFile != "fake-audio" {
		t.Fatalf("unexpected file body: %q", gotFile)
	}

	if gotFileContentType != "audio/wav" {
		t.Fatalf("unexpected file content type: %q", gotFileContentType)
	}

	if !strings.Contains(rec.Body.String(), `"text":"cohere ok"`) {
		t.Fatalf("expected Cohere body passthrough, got %s", rec.Body.String())
	}
}

func TestDispatchGoogleVertexAudioTranscriptionsAdapter(t *testing.T) {
	t.Parallel()

	var (
		gotAuth     string
		gotPath     string
		gotModel    string
		gotLanguage []interface{}
		gotContent  string
	)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")

		gotPath = r.URL.Path
		defer r.Body.Close()

		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
			return
		}

		config, _ := body["config"].(map[string]interface{})
		gotModel, _ = config["model"].(string)
		gotLanguage, _ = config["languageCodes"].([]interface{})
		gotContent, _ = body["content"].(string)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"results":[{"alternatives":[{"transcript":"سلام"}]},{"alternatives":[{"transcript":"دنیا"}]}]}`))
	}))
	defer upstream.Close()

	handler := NewHandler(nil)
	payload := platform.TunnelRequest{
		Deployment: platform.Deployment{
			Provider:     "google-vertex",
			APIBase:      upstream.URL,
			APIKey:       "google-token",
			Model:        "google-vertex/google/chirp-3",
			ProjectID:    "project-1",
			RegionName:   "global",
			RecognizerID: "recognizer-1",
		},
		Headers: map[string]string{"Content-Type": "multipart/form-data; boundary=shahrah-google-boundary"},
		RequestMetadata: map[string]interface{}{
			"provider_endpoint_path": "/audio/transcriptions",
		},
		Body: audioMultipartPayloadBody("shahrah-google-boundary", map[string]string{"language": "fa-IR"}, "sample.wav", "audio/wav", "fake-audio"),
	}
	req := dispatchRequest(t, payload)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}

	if gotPath != "/projects/project-1/locations/global/recognizers/recognizer-1:recognize" {
		t.Fatalf("expected Google recognize path, got %q", gotPath)
	}

	if gotAuth != "Bearer google-token" {
		t.Fatalf("unexpected Authorization header: %q", gotAuth)
	}

	if gotModel != "google/chirp-3" {
		t.Fatalf("expected Google model, got %q", gotModel)
	}

	if len(gotLanguage) != 1 || gotLanguage[0] != "fa-IR" {
		t.Fatalf("expected languageCodes [fa-IR], got %#v", gotLanguage)
	}

	if gotContent != base64.StdEncoding.EncodeToString([]byte("fake-audio")) {
		t.Fatalf("unexpected base64 content: %q", gotContent)
	}

	if !strings.Contains(rec.Body.String(), `"text":"سلام\nدنیا"`) {
		t.Fatalf("expected normalized Google transcript response, got %s", rec.Body.String())
	}
}

func TestDispatchGoogleVertexAudioTranscriptionsAdapterAcceptsRecognizerResourceName(t *testing.T) {
	t.Parallel()

	var gotPath string

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"results":[{"alternatives":[{"transcript":"ok"}]}]}`))
	}))
	defer upstream.Close()

	handler := NewHandler(nil)
	payload := platform.TunnelRequest{
		Deployment: platform.Deployment{
			Provider:     "google-vertex",
			APIBase:      upstream.URL,
			APIKey:       "google-token",
			Model:        "google-vertex/google/chirp-3",
			RecognizerID: "projects/project-1/locations/global/recognizers/recognizer-1",
		},
		Headers: map[string]string{"Content-Type": "multipart/form-data; boundary=shahrah-google-boundary"},
		RequestMetadata: map[string]interface{}{
			"provider_endpoint_path": "/audio/transcriptions",
		},
		Body: audioMultipartPayloadBody("shahrah-google-boundary", map[string]string{"language": "fa-IR"}, "sample.wav", "audio/wav", "fake-audio"),
	}
	req := dispatchRequest(t, payload)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}

	if gotPath != "/projects/project-1/locations/global/recognizers/recognizer-1:recognize" {
		t.Fatalf("expected Google recognize path, got %q", gotPath)
	}
}

func TestDispatchCohereAudioTranscriptionsRequiresLanguage(t *testing.T) {
	t.Parallel()

	handler := NewHandler(nil)
	payload := platform.TunnelRequest{
		Deployment: platform.Deployment{Provider: "cohere", APIKey: "cohere-secret", Model: "cohere/cohere-transcribe-03-2026"},
		Headers:    map[string]string{"Content-Type": "multipart/form-data; boundary=shahrah-cohere-boundary"},
		RequestMetadata: map[string]interface{}{
			"provider_endpoint_path": "/audio/transcriptions",
		},
		Body: audioMultipartPayloadBody("shahrah-cohere-boundary", nil, "sample.wav", "audio/wav", "fake-audio"),
	}
	req := dispatchRequest(t, payload)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusBadRequest, rec.Code, rec.Body.String())
	}

	if !strings.Contains(rec.Body.String(), "language is required") {
		t.Fatalf("expected language validation error, got %s", rec.Body.String())
	}
}

func TestDispatchOpenAICompatibleForwardsAudioSpeechJSON(t *testing.T) {
	t.Parallel()

	var (
		gotAuth        string
		gotContentType string
		gotPath        string
		gotBody        string
	)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotContentType = r.Header.Get("Content-Type")

		gotPath = r.URL.Path
		defer r.Body.Close()

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
			return
		}

		gotBody = string(bodyBytes)

		w.Header().Set("Content-Type", "audio/mpeg")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("mp3"))
	}))
	defer upstream.Close()

	handler := NewHandler(nil)
	payload := platform.TunnelRequest{
		Deployment: platform.Deployment{Provider: "openai", APIBase: upstream.URL, APIKey: "provider-secret"},
		Headers:    map[string]string{"Content-Type": "application/json"},
		RequestMetadata: map[string]interface{}{
			"provider_endpoint_path": "/audio/speech",
		},
		Body: []byte(`{"model":"gpt-4o-mini-tts","input":"hello","voice":"alloy"}`),
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

	if gotPath != "/audio/speech" {
		t.Fatalf("expected audio speech path, got %q", gotPath)
	}

	if gotContentType != "application/json" {
		t.Fatalf("unexpected content type: %q", gotContentType)
	}

	if gotAuth != "Bearer provider-secret" {
		t.Fatalf("unexpected Authorization header: %q", gotAuth)
	}

	if !strings.Contains(gotBody, `"model":"gpt-4o-mini-tts"`) || !strings.Contains(gotBody, `"voice":"alloy"`) {
		t.Fatalf("unexpected forwarded body: %s", gotBody)
	}
}

func TestDispatchUsesAnthropicAdapterForSync(t *testing.T) {
	t.Parallel()

	var (
		gotPath    string
		gotAPIKey  string
		gotVersion string
		body       map[string]any
	)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAPIKey = r.Header.Get("x-api-key")

		gotVersion = r.Header.Get("anthropic-version")
		defer r.Body.Close()

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode request: %v", err)
			return
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
		ModelRuntimeProfiles: []platform.ModelRuntimeProfileBinding{
			anthropicTestRuntimeProfile("claude-3-haiku-20240307"),
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

func TestDispatchAnthropicAdapterMapsExtendedThinking(t *testing.T) {
	t.Parallel()

	var body map[string]any

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode request: %v", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"id":"msg_thinking",
			"type":"message",
			"role":"assistant",
			"model":"claude-3-7-sonnet",
			"content":[
				{"type":"thinking","thinking":"private reasoning","signature":"sig_123"},
				{"type":"redacted_thinking","data":"opaque_blob"},
				{"type":"text","text":"Final answer"}
			],
			"usage":{"input_tokens":12,"output_tokens":6},
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
		ModelRuntimeProfiles: []platform.ModelRuntimeProfileBinding{
			anthropicTestRuntimeProfile("anthropic/claude-3-7-sonnet"),
		},
		Body: []byte(`{
			"model":"anthropic/claude-3-7-sonnet",
			"reasoning_effort":"high",
			"max_tokens":4000,
			"messages":[
				{"role":"user","content":"solve"},
				{"role":"assistant","content":"previous","thinking_blocks":[{"type":"thinking","thinking":"previous reasoning","signature":"sig_prev"}]},
				{"role":"user","content":"continue"}
			]
		}`),
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

	thinking := body["thinking"].(map[string]any)
	if thinking["type"] != "enabled" || thinking["budget_tokens"].(float64) != 3200 {
		t.Fatalf("unexpected anthropic thinking config: %#v", thinking)
	}

	messages := body["messages"].([]any)
	assistantContent := messages[1].(map[string]any)["content"].([]any)

	replayedThinking := assistantContent[0].(map[string]any)
	if replayedThinking["type"] != "thinking" || replayedThinking["signature"] != "sig_prev" {
		t.Fatalf("thinking block was not replayed for anthropic: %#v", replayedThinking)
	}

	response := rec.Body.String()
	for _, want := range []string{`"content":"Final answer"`, `"reasoning_content":"private reasoning"`, `"thinking_blocks"`, `"signature":"sig_123"`, `"type":"redacted_thinking"`, `"data":"opaque_blob"`} {
		if !strings.Contains(response, want) {
			t.Fatalf("response missing %s: %s", want, response)
		}
	}
}

func TestDispatchAnthropicAdapterConvertsImageDataURL(t *testing.T) {
	t.Parallel()

	var body map[string]any

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode request: %v", err)
			return
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
		ModelRuntimeProfiles: []platform.ModelRuntimeProfileBinding{
			anthropicTestRuntimeProfile("claude-3-haiku-20240307"),
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
			t.Errorf("decode request: %v", err)
			return
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
		ModelRuntimeProfiles: []platform.ModelRuntimeProfileBinding{
			anthropicTestRuntimeProfile("claude-3-haiku-20240307"),
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

func TestDispatchGeminiAdapterUsesOfficialOpenAICompatibility(t *testing.T) {
	t.Parallel()

	var (
		gotPath          string
		gotAuthorization string
		gotQueryKey      string
		body             map[string]any
	)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuthorization = r.Header.Get("Authorization")

		gotQueryKey = r.URL.Query().Get("key")
		defer r.Body.Close()

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode request: %v", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"chatcmpl-google","object":"chat.completion","choices":[{"index":0,"message":{"role":"assistant","content":"small icon"},"finish_reason":"stop"}],"usage":{"prompt_tokens":12,"completion_tokens":2,"total_tokens":14}}`))
	}))
	defer upstream.Close()

	handler := NewHandler(nil)
	payload := platform.TunnelRequest{
		Deployment: platform.Deployment{
			Provider: "gemini",
			APIBase:  upstream.URL + "/v1beta",
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

	if gotPath != "/v1beta/openai/chat/completions" {
		t.Fatalf("unexpected gemini path: %q", gotPath)
	}

	if gotAuthorization != "Bearer gemini-key" || gotQueryKey != "" {
		t.Fatalf("Gemini OpenAI compatibility must use bearer auth only; header=%q query=%q", gotAuthorization, gotQueryKey)
	}

	messages := body["messages"].([]any)
	parts := messages[0].(map[string]any)["content"].([]any)

	imageURL := parts[1].(map[string]any)["image_url"].(map[string]any)
	if imageURL["url"] != "data:image/png;base64,AAAA" {
		t.Fatalf("unexpected OpenAI-compatible Gemini image: %#v", imageURL)
	}

	if !strings.Contains(rec.Body.String(), `"object":"chat.completion"`) {
		t.Fatalf("unexpected response: %s", rec.Body.String())
	}
}

func TestDispatchGeminiAdapterPreservesDocumentedAudioPayload(t *testing.T) {
	t.Parallel()

	var (
		gotPath string
		body    map[string]any
	)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		defer r.Body.Close()

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode request: %v", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"chatcmpl-google","object":"chat.completion","choices":[{"index":0,"message":{"role":"assistant","content":"short guide"},"finish_reason":"stop"}],"usage":{"prompt_tokens":20,"completion_tokens":3,"total_tokens":23}}`))
	}))
	defer upstream.Close()

	handler := NewHandler(nil)
	payload := platform.TunnelRequest{
		Deployment: platform.Deployment{
			Provider: "gemini",
			APIBase:  upstream.URL + "/v1beta",
			APIKey:   "gemini-key",
		},
		Body: []byte(`{"model":"gemini/gemini-3.1-flash-lite","messages":[{"role":"user","content":[{"type":"text","text":"transcribe"},{"type":"input_audio","input_audio":{"data":"UklGRg==","format":"wav"}}]}]}`),
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

	if gotPath != "/v1beta/openai/chat/completions" {
		t.Fatalf("unexpected gemini path: %q", gotPath)
	}

	messages := body["messages"].([]any)
	parts := messages[0].(map[string]any)["content"].([]any)

	inputAudio := parts[1].(map[string]any)["input_audio"].(map[string]any)
	if inputAudio["data"] != "UklGRg==" || inputAudio["format"] != "wav" {
		t.Fatalf("unexpected OpenAI-compatible Gemini audio: %#v", inputAudio)
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
			`data: {"type":"content_block_delta","index":0,"delta":{"type":"thinking_delta","thinking":"Reasoning"}}`,
			"",
			"event: content_block_delta",
			`data: {"type":"content_block_delta","index":0,"delta":{"type":"signature_delta","signature":"sig_stream"}}`,
			"",
			"event: content_block_delta",
			`data: {"type":"content_block_delta","index":0,"delta":{"text":"Salam"}}`,
			"",
			"event: message_delta",
			`data: {"type":"message_delta","delta":{"stop_reason":"end_turn"}}`,
			"",
			"event: message_stop",
			`data: {"type":"message_stop"}`,
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
		ModelRuntimeProfiles: []platform.ModelRuntimeProfileBinding{
			anthropicTestRuntimeProfile("claude-3-haiku-20240307"),
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

	if !strings.Contains(streamText, `"reasoning_content":"Reasoning"`) {
		t.Fatalf("expected reasoning delta in stream, got %q", streamText)
	}

	if !strings.Contains(streamText, `"signature":"sig_stream"`) {
		t.Fatalf("expected signature delta in stream, got %q", streamText)
	}

	if !strings.Contains(streamText, "data: [DONE]") {
		t.Fatalf("expected stream terminator, got %q", streamText)
	}

	scanner := bufio.NewScanner(strings.NewReader(streamText))
	found := false

	for scanner.Scan() {
		if strings.Contains(scanner.Text(), `"finish_reason":"stop"`) {
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
			t.Error("response writer does not support flushing")
			return
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

func TestRelayBodyOrAbortPreservesUpstreamReadFailure(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	reader := &terminalErrorReader{data: []byte("data: partial\n\n")}

	defer func() {
		recovered := recover()
		if recovered != http.ErrAbortHandler {
			t.Fatalf("expected http.ErrAbortHandler, got %#v", recovered)
		}

		if recorder.Body.String() != "data: partial\n\n" {
			t.Fatalf("expected partial response to be preserved, got %q", recorder.Body.String())
		}
	}()

	relayBodyOrAbort(recorder, reader, true, "req_stream_failure", "trace_stream_failure")
	t.Fatal("expected relayBodyOrAbort to abort the downstream response")
}

type terminalErrorReader struct {
	data []byte
	sent bool
}

func (r *terminalErrorReader) Read(buffer []byte) (int, error) {
	if r.sent {
		return 0, io.ErrUnexpectedEOF
	}

	r.sent = true

	return copy(buffer, r.data), nil
}

func TestDispatchOpenAIResponsesStreamingRelaysMultipleEventsAndDone(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/responses" {
			t.Errorf("expected /responses path, got %q", got)
			return
		}

		if got := r.Header.Get("Accept"); got != "text/event-stream" {
			t.Errorf("unexpected Accept header: %q", got)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")

		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Error("response writer does not support flushing")
			return
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

func dispatchRequest(t *testing.T, payload platform.TunnelRequest) *http.Request {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/dispatch", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")

	return req
}

func audioMultipartPayloadBody(boundary string, fields map[string]string, filename string, contentType string, fileBody string) []byte {
	lines := make([]string, 0, 12+len(fields)*4)
	for key, value := range fields {
		lines = append(lines,
			"--"+boundary,
			fmt.Sprintf("Content-Disposition: form-data; name=%q", key),
			"",
			value,
		)
	}

	lines = append(lines,
		"--"+boundary,
		fmt.Sprintf("Content-Disposition: form-data; name=\"file\"; filename=%q", filename),
		"Content-Type: "+contentType,
		"",
		fileBody,
		"--"+boundary+"--",
		"",
	)

	return []byte(strings.Join(lines, "\r\n"))
}
