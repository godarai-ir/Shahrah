package tunnel

import (
	"bytes"
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
	receivedBody := ""

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		hits++

		metadata, body := decodeRawTunnelRequest(t, request.Body)
		if len(metadata.Body) != 0 {
			t.Errorf("expected body to be omitted from metadata")
		}

		receivedBody = string(body)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer upstream.Close()

	transport := NewHTTPTransport(upstream.URL, time.Second)
	request := platform.TunnelRequest{
		Deployment: platform.Deployment{Provider: "openai"},
		Body:       []byte(`{"model":"gpt-test"}`),
	}

	if _, err := transport.Proxy(t.Context(), request); err != nil {
		t.Fatalf("OpenAI request failed: %v", err)
	}

	if hits != 1 {
		t.Fatalf("OpenAI request did not reach the tunnel endpoint")
	}

	if receivedBody != string(request.Body) {
		t.Fatalf("expected raw request body %q, got %q", request.Body, receivedBody)
	}
}

func TestHTTPTransportUsesConfiguredSharedConnectionPool(t *testing.T) {
	transport := NewHTTPTransportWithConfig("http://silk-road", HTTPTransportConfig{
		RequestTimeout:      time.Second,
		MaxIdleConns:        300,
		MaxIdleConnsPerHost: 200,
		MaxConnsPerHost:     250,
		IdleConnTimeout:     45 * time.Second,
	})

	roundTripper, ok := transport.client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("expected HTTP transport, got %T", transport.client.Transport)
	}

	if transport.streamClient.Transport != roundTripper {
		t.Fatal("expected sync and streaming clients to share a connection pool")
	}

	if roundTripper.MaxIdleConns != 300 || roundTripper.MaxIdleConnsPerHost != 200 ||
		roundTripper.MaxConnsPerHost != 250 || roundTripper.IdleConnTimeout != 45*time.Second {
		t.Fatalf("unexpected pool config: %+v", roundTripper)
	}
}

func TestProxyRawStreamsLengthPrefixedEnvelope(t *testing.T) {
	var (
		received     platform.TunnelRequest
		receivedBody string
	)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/dispatch/raw" {
			t.Errorf("unexpected path: %s", request.URL.Path)
			return
		}

		var prefix [4]byte
		if _, err := io.ReadFull(request.Body, prefix[:]); err != nil {
			t.Error(err)
			return
		}

		metadata := make([]byte, binary.BigEndian.Uint32(prefix[:]))
		if _, err := io.ReadFull(request.Body, metadata); err != nil {
			t.Error(err)
			return
		}

		if err := json.Unmarshal(metadata, &received); err != nil {
			t.Error(err)
			return
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
	defer closeBody(response.Body)

	output, _ := io.ReadAll(response.Body)
	if received.Operation != platform.ProviderOperationAnthropicFilesCreate ||
		receivedBody != "large-body" || string(output) != "download" {
		t.Fatalf("raw tunnel mismatch: operation=%s body=%q output=%q", received.Operation, receivedBody, output)
	}
}

func TestProxyHonorsConfiguredRequestTimeout(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.Copy(io.Discard, r.Body); err != nil {
			return
		}

		time.Sleep(100 * time.Millisecond)

		if _, err := w.Write([]byte(`{}`)); err != nil {
			return
		}
	}))
	defer upstream.Close()

	transport := NewHTTPTransport(upstream.URL, 10*time.Millisecond)

	_, err := transport.Proxy(t.Context(), platform.TunnelRequest{Body: []byte(`{"model":"gpt-test"}`)})
	if err == nil {
		t.Fatal("expected non-streaming proxy request to honor its configured timeout")
	}
}

func TestProxyRawPreservesLongRunningRawOperation(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.Copy(io.Discard, r.Body); err != nil {
			return
		}

		time.Sleep(50 * time.Millisecond)

		if _, err := w.Write([]byte("raw-response")); err != nil {
			return
		}
	}))
	defer upstream.Close()

	transport := NewHTTPTransport(upstream.URL, 10*time.Millisecond)

	response, err := transport.ProxyRaw(t.Context(), platform.TunnelRequest{}, http.NoBody, 0)
	if err != nil {
		t.Fatalf("expected raw operation to outlive the normal request timeout: %v", err)
	}
	defer closeBody(response.Body)

	body, err := io.ReadAll(response.Body)
	if err != nil || string(body) != "raw-response" {
		t.Fatalf("unexpected raw response body %q: %v", body, err)
	}
}

func TestStreamIgnoresClientTimeoutWhileBodyIsOpen(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dispatch/raw" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			return
		}

		_, _ = decodeRawTunnelRequest(t, r.Body)

		w.Header().Set("Content-Type", "text/event-stream")

		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Error("response writer does not support flushing")
			return
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

	var (
		payload   platform.TunnelRequest
		requestID string
		traceID   string
	)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		payload, _ = decodeRawTunnelRequest(t, r.Body)
		requestID = r.Header.Get("X-Shahrah-Request-ID")
		traceID = r.Header.Get("X-GodarAI-Trace-ID")

		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: ok\n\n"))
	}))
	defer upstream.Close()

	transport := NewHTTPTransport(upstream.URL, time.Second)

	stream, err := transport.Stream(context.Background(), platform.TunnelRequest{
		RequestID: "req_stream_headers",
		TraceID:   "trace_stream_headers",
	})
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

	if requestID != "req_stream_headers" || traceID != "trace_stream_headers" {
		t.Fatalf("expected request correlation headers, got request_id=%q trace_id=%q", requestID, traceID)
	}
}

func decodeRawTunnelRequest(t *testing.T, body io.Reader) (platform.TunnelRequest, []byte) {
	t.Helper()

	var prefix [4]byte
	if _, err := io.ReadFull(body, prefix[:]); err != nil {
		t.Errorf("read raw tunnel prefix: %v", err)

		return platform.TunnelRequest{}, nil
	}

	metadata := make([]byte, binary.BigEndian.Uint32(prefix[:]))
	if _, err := io.ReadFull(body, metadata); err != nil {
		t.Errorf("read raw tunnel metadata: %v", err)

		return platform.TunnelRequest{}, nil
	}

	var payload platform.TunnelRequest
	if err := json.Unmarshal(metadata, &payload); err != nil {
		t.Errorf("decode raw tunnel metadata: %v", err)

		return platform.TunnelRequest{}, nil
	}

	raw, err := io.ReadAll(body)
	if err != nil {
		t.Errorf("read raw tunnel body: %v", err)
	}

	return payload, raw
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
			t.Error("response writer does not support flushing")
			return
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

func TestStreamCloseCancelsBlockedProducer(t *testing.T) {
	requestCanceled := make(chan struct{})

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = decodeRawTunnelRequest(t, r.Body)

		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: first\n\n"))

		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Error("response writer does not support flushing")
			return
		}

		flusher.Flush()

		<-r.Context().Done()
		close(requestCanceled)
	}))
	defer upstream.Close()

	transport := NewHTTPTransport(upstream.URL, time.Second)

	stream, err := transport.Stream(t.Context(), platform.TunnelRequest{})
	if err != nil {
		t.Fatalf("stream request failed: %v", err)
	}

	select {
	case <-stream.Chunks():
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for first stream chunk")
	}

	if err := stream.Close(); err != nil {
		t.Fatalf("close stream: %v", err)
	}

	select {
	case <-requestCanceled:
	case <-time.After(time.Second):
		t.Fatal("closing stream did not cancel upstream request")
	}
}

func BenchmarkTunnelRequestEnvelopeEncoding(b *testing.B) {
	body := bytes.Repeat([]byte(`{"message":"performance"}`), 4096)
	request := platform.TunnelRequest{
		RequestID:  "request-benchmark",
		TraceID:    "trace-benchmark",
		Deployment: platform.Deployment{Provider: "openai", Model: "gpt-test"},
		Body:       body,
	}

	b.Run("legacy-json-base64", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(body)))

		for b.Loop() {
			encoded, err := json.Marshal(request)
			if err != nil || len(encoded) == 0 {
				b.Fatalf("encode legacy envelope: %v", err)
			}
		}
	})

	b.Run("raw-metadata", func(b *testing.B) {
		metadata := request
		metadata.Body = nil

		b.ReportAllocs()
		b.SetBytes(int64(len(body)))

		for b.Loop() {
			encoded, err := json.Marshal(metadata)
			if err != nil || len(encoded) == 0 {
				b.Fatalf("encode raw metadata: %v", err)
			}
		}
	})
}
