package server

import (
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
			APIBase: upstream.URL,
			APIKey:  "provider-secret",
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
