package tunnel

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func TestOpenAIRequestsReachTunnelWithoutReleaseGate(t *testing.T) {
	hits := 0
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits++
		_, _ = w.Write([]byte(`{}`))
	}))
	defer upstream.Close()

	transport := NewHTTPTransport(upstream.URL, time.Second)
	request := platform.TunnelRequest{
		Deployment: platform.Deployment{Provider: "openai"},
	}

	if _, err := transport.Proxy(t.Context(), request); err != nil {
		t.Fatalf("OpenAI request failed: %v", err)
	}
	if hits != 1 {
		t.Fatalf("OpenAI request did not reach the tunnel endpoint")
	}
}

func TestProxyRawStreamsLengthPrefixedEnvelope(t *testing.T) {
	var received platform.TunnelRequest
	var receivedBody string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/dispatch/raw" {
			t.Fatalf("unexpected path: %s", request.URL.Path)
		}
		var prefix [4]byte
		if _, err := io.ReadFull(request.Body, prefix[:]); err != nil {
			t.Fatal(err)
		}
		metadata := make([]byte, binary.BigEndian.Uint32(prefix[:]))
		if _, err := io.ReadFull(request.Body, metadata); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(metadata, &received); err != nil {
			t.Fatal(err)
		}
		raw, _ := io.ReadAll(request.Body)
		receivedBody = string(raw)
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write([]byte("download"))
	}))
	defer upstream.Close()

	transport := NewHTTPTransport(upstream.URL, time.Second)
	response, err := transport.ProxyRaw(t.Context(), platform.TunnelRequest{
		Operation:  platform.ProviderOperationAnthropicFilesCreate,
		Deployment: platform.Deployment{Provider: "anthropic"},
	}, strings.NewReader("large-body"), int64(len("large-body")))
	if err != nil {
		t.Fatalf("ProxyRaw: %v", err)
	}
	defer response.Body.Close()
	output, _ := io.ReadAll(response.Body)
	if received.Operation != platform.ProviderOperationAnthropicFilesCreate ||
		receivedBody != "large-body" || string(output) != "download" {
		t.Fatalf("raw tunnel mismatch: operation=%s body=%q output=%q", received.Operation, receivedBody, output)
	}
}

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

	for chunk := range stream.Chunks() {
		_ = chunk
	}

	if !payload.Stream {
		t.Fatal("expected stream flag to be forced on outbound tunnel request")
	}
}

func TestStreamReturnsErrorStreamForHTTPFailure(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Request-Id", "req_error")
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"error":"upstream failed"}`))
	}))
	defer upstream.Close()

	transport := NewHTTPTransport(upstream.URL, time.Second)

	stream, err := transport.Stream(context.Background(), platform.TunnelRequest{})
	if err != nil {
		t.Fatalf("stream request failed: %v", err)
	}
	defer stream.Close()

	assertedHeaders := stream.Headers()
	if assertedHeaders["X-Request-Id"] != "req_error" {
		t.Fatalf("expected propagated error header, got %#v", assertedHeaders)
	}

	if stream.StatusCode() != http.StatusBadGateway {
		t.Fatalf("expected upstream status %d, got %d", http.StatusBadGateway, stream.StatusCode())
	}

	chunks := stream.Chunks()

	chunk, ok := <-chunks
	if !ok {
		t.Fatal("expected one error chunk")
	}

	if !chunk.IsDone {
		t.Fatal("expected error chunk to mark stream done")
	}

	if !strings.Contains(string(chunk.Data), `upstream failed`) {
		t.Fatalf("unexpected error chunk body: %q", string(chunk.Data))
	}
}

func TestStreamPreservesHeadersAndMultiEventFrames(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")

		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("response writer does not support flushing")
		}

		w.WriteHeader(http.StatusOK)

		for _, frame := range []string{
			"event: response.created\n",
			"data: {\"type\":\"response.created\"}\n\n",
			"data: {\"type\":\"response.output_text.delta\",\"delta\":\"Salam\"}\n\n",
		} {
			_, _ = w.Write([]byte(frame))

			flusher.Flush()
		}
	}))
	defer upstream.Close()

	transport := NewHTTPTransport(upstream.URL, time.Second)

	stream, err := transport.Stream(context.Background(), platform.TunnelRequest{})
	if err != nil {
		t.Fatalf("stream request failed: %v", err)
	}
	defer stream.Close()

	if got := stream.Headers()["Content-Type"]; got != "text/event-stream" {
		t.Fatalf("unexpected content type header: %q", got)
	}

	if got := stream.Headers()["Cache-Control"]; got != "no-cache" {
		t.Fatalf("unexpected cache-control header: %q", got)
	}

	var sb strings.Builder
	for chunk := range stream.Chunks() {
		sb.Write(chunk.Data)

		if chunk.IsDone {
			break
		}
	}

	body := sb.String()
	if !strings.Contains(body, "event: response.created") {
		t.Fatalf("expected event frame, got %q", body)
	}

	if !strings.Contains(body, `response.output_text.delta`) {
		t.Fatalf("expected multi-event data frame, got %q", body)
	}
}
