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
		gotTraceID = r.Header.Get("X-Shahrah-Trace-ID")
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
