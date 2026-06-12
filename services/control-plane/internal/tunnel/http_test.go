package tunnel

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func TestStreamIgnoresClientTimeoutWhileBodyIsOpen(t *testing.T) {
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
		_, _ = w.Write([]byte("data: first\n\n"))
		flusher.Flush()

		time.Sleep(50 * time.Millisecond)

		_, _ = w.Write([]byte("data: second\n\n"))
		flusher.Flush()
	}))
	defer upstream.Close()

	transport := NewHTTPTransport(upstream.URL, 20*time.Millisecond)
	stream, err := transport.Stream(context.Background(), platform.TunnelRequest{
		RequestID: "req_stream_timeout",
		TraceID:   "trace_stream_timeout",
		Stream:    true,
	})
	if err != nil {
		t.Fatalf("stream request failed: %v", err)
	}
	defer stream.Close()

	var sb strings.Builder
	timeout := time.After(300 * time.Millisecond)

	for {
		select {
		case chunk, ok := <-stream.Chunks():
			if !ok {
				if !strings.Contains(sb.String(), "data: first") || !strings.Contains(sb.String(), "data: second") {
					t.Fatalf("expected both stream chunks, got %q", sb.String())
				}
				return
			}
			sb.Write(chunk.Data)
			if chunk.IsDone {
				if !strings.Contains(sb.String(), "data: first") || !strings.Contains(sb.String(), "data: second") {
					t.Fatalf("expected both stream chunks, got %q", sb.String())
				}
				return
			}
		case err := <-stream.Errors():
			if err != nil {
				t.Fatalf("unexpected stream error: %v; collected=%q", err, sb.String())
			}
		case <-timeout:
			t.Fatalf("timed out waiting for stream chunks; collected=%q", sb.String())
		}
	}
}

func TestStreamSetsRequestStreamFlag(t *testing.T) {
	t.Parallel()

	var payload platform.TunnelRequest
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: ok\n\n"))
	}))
	defer upstream.Close()

	transport := NewHTTPTransport(upstream.URL, time.Second)
	stream, err := transport.Stream(context.Background(), platform.TunnelRequest{})
	if err != nil {
		t.Fatalf("stream request failed: %v", err)
	}
	defer stream.Close()

	for range stream.Chunks() {
	}

	if !payload.Stream {
		t.Fatal("expected stream flag to be forced on outbound tunnel request")
	}
}
