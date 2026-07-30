package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRunProducesSanitizedPassingModelMatrix(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(
		w http.ResponseWriter,
		request *http.Request,
	) {
		if request.Header.Get("x-api-key") != "virtual-secret" {
			t.Error("missing virtual key")
		}
		w.Header().Set("request-id", "req-safe")
		w.Header().Set("Content-Type", "application/json")
		switch {
		case request.Method == http.MethodGet &&
			strings.HasPrefix(request.URL.Path, "/v1/models/"):
			_, _ = io.WriteString(w, `{"id":"claude-sonnet-5","max_input_tokens":1000000,"max_tokens":128000}`)
		case request.URL.Path == "/v1/messages/count_tokens":
			_, _ = io.WriteString(w, `{"input_tokens":8}`)
		case request.URL.Path == "/v1/chat/completions":
			body, _ := io.ReadAll(request.Body)
			if !strings.Contains(string(body), `"type":"web_search_20260318"`) {
				t.Errorf("readiness matrix must use the model-profile web search version: %s", body)
			}
			_, _ = io.WriteString(w, `{"id":"chat","choices":[{"message":{"content":"OK"}}]}`)
		case request.URL.Path == "/v1/messages":
			body, _ := io.ReadAll(request.Body)
			if strings.Contains(string(body), `"stream":true`) {
				w.Header().Set("Content-Type", "text/event-stream")
				_, _ = io.WriteString(
					w,
					"event: message_start\ndata: {\"type\":\"message_start\",\"message\":{\"model\":\"claude-sonnet-5\"}}\n\n"+
						"event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n",
				)
				return
			}
			_, _ = io.WriteString(w, `{"id":"msg","type":"message","model":"claude-sonnet-5","content":[{"type":"text","text":"OK"}]}`)
		default:
			http.NotFound(w, request)
		}
	}))
	defer server.Close()

	result := run(context.Background(), server.Client(), config{
		baseURL:    server.URL,
		virtualKey: "virtual-secret",
		authLabel:  "api_key",
		models:     []string{"anthropic/claude-sonnet-5"},
		timeout:    time.Minute,
	})
	if result.Summary.Failed != 0 ||
		result.Summary.Passed != 7 ||
		result.Summary.Blocked != 1 {
		t.Fatalf("unexpected summary: %#v", result.Summary)
	}
	if len(result.Models) != 1 {
		t.Fatalf("unexpected model report count: %d", len(result.Models))
	}
	for _, check := range result.Models[0].Checks {
		if check.RequestIDHash == "" || check.RequestIDHash == "req-safe" {
			t.Fatalf("request id was not safely hashed for %s", check.Name)
		}
		if strings.Contains(check.Detail, "virtual-secret") {
			t.Fatal("secret leaked to report detail")
		}
	}
}
