package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func TestHealthProxy(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("ok"))
	}))
	defer upstream.Close()

	handler := NewHandler(upstream.URL, nil)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, rec.Code)
	}
	if strings.TrimSpace(rec.Body.String()) != "ok" {
		t.Fatalf("unexpected body: %q", rec.Body.String())
	}
}

func TestDispatchForwardsTraceAndBody(t *testing.T) {
	t.Parallel()

	var gotTraceID string
	var gotRequestID string
	var gotReq platform.TunnelRequest

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dispatch" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		gotTraceID = r.Header.Get("X-GodarAI-Trace-ID")
		gotRequestID = r.Header.Get("X-Shahrah-Request-ID")

		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&gotReq); err != nil {
			t.Fatalf("decode request body: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"status":"forwarded"}`))
	}))
	defer upstream.Close()

	handler := NewHandler(upstream.URL, nil)
	payload := platform.TunnelRequest{
		RequestID: "req_123",
		TraceID:   "trace_123",
		Body:      []byte(`{"hello":"world"}`),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/dispatch", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}
	if gotTraceID != payload.TraceID {
		t.Fatalf("expected trace id %q, got %q", payload.TraceID, gotTraceID)
	}
	if gotRequestID != payload.RequestID {
		t.Fatalf("expected request id %q, got %q", payload.RequestID, gotRequestID)
	}
	if string(gotReq.Body) != string(payload.Body) {
		t.Fatalf("expected body %q, got %q", string(payload.Body), string(gotReq.Body))
	}
}

func TestDispatchStreamingFlushesSSE(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dispatch" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
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

	handler := NewHandler(upstream.URL, nil)
	payload := platform.TunnelRequest{
		RequestID: "req_stream",
		TraceID:   "trace_stream",
		Stream:    true,
		Body:      []byte(`{"hello":"world"}`),
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

func TestDispatchStreamingRelaysMultipleEventsWithoutTruncation(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dispatch" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
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

	handler := NewHandler(upstream.URL, nil)
	payload := platform.TunnelRequest{
		RequestID: "req_stream_multi",
		TraceID:   "trace_stream_multi",
		Stream:    true,
		Body:      []byte(`{"hello":"world"}`),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/dispatch", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !rec.Flushed {
		t.Fatal("expected streaming response to flush SSE data")
	}
	bodyText := rec.Body.String()
	if !strings.Contains(bodyText, "event: response.created") || !strings.Contains(bodyText, "data: [DONE]") {
		t.Fatalf("expected full SSE relay without truncation, got %q", bodyText)
	}
}
