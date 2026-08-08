package server

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func TestHealthIsLocal(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("liveness must not call upstream: %s", r.URL.Path)
		return
	}))
	defer upstream.Close()

	handler := NewHandler(upstream.URL, nil)
	for _, path := range []string{"/health", "/live"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("%s: expected status %d, got %d", path, http.StatusOK, rec.Code)
		}
	}
}

func TestReadinessChecksTunnelServer(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ready" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ready","service":"tunnel-server"}`))
	}))
	defer upstream.Close()

	handler := NewHandler(upstream.URL, nil)
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestReadinessFailsWhenTunnelServerIsUnavailable(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "not ready", http.StatusServiceUnavailable)
	}))
	defer upstream.Close()

	handler := NewHandler(upstream.URL, nil)
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, rec.Code)
	}
}

func TestMetricsEndpointReportsRequests(t *testing.T) {
	t.Parallel()

	handler := NewHandler("http://127.0.0.1:1", nil)
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/live", nil))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "shahrah_transport_http_requests_total") ||
		!strings.Contains(body, `service="tunnel-client"`) ||
		!strings.Contains(body, `route="live"`) {
		t.Fatalf("expected tunnel-client request metrics, got %q", body)
	}
}

func TestDispatchForwardsTraceAndBody(t *testing.T) {
	t.Parallel()

	var gotTraceID string
	var gotRequestID string
	var gotReq platform.TunnelRequest

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dispatch" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			return
		}
		gotTraceID = r.Header.Get("X-GodarAI-Trace-ID")
		gotRequestID = r.Header.Get("X-Shahrah-Request-ID")

		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&gotReq); err != nil {
			t.Errorf("decode request body: %v", err)
			return
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
			t.Errorf("unexpected path: %s", r.URL.Path)
			return
		}
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

func TestDispatchStreamingRelaysMultipleEventsWithoutTruncation(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dispatch" {
			t.Errorf("unexpected path: %s", r.URL.Path)
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

func TestRawDispatchIsNotCutOffByRegularClientTimeout(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(50 * time.Millisecond)

		if _, err := w.Write([]byte("raw-response")); err != nil {
			return
		}
	}))
	defer upstream.Close()

	handler := NewHandler(upstream.URL, &http.Client{Timeout: 10 * time.Millisecond})
	request := httptest.NewRequest(http.MethodPost, "/dispatch/raw", strings.NewReader("raw-envelope"))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK || recorder.Body.String() != "raw-response" {
		t.Fatalf("expected long-running raw response, got status=%d body=%q", recorder.Code, recorder.Body.String())
	}
}
