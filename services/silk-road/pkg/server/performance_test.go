package server

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

const benchmarkPayloadSize = 1024 * 1024

type benchmarkResponseWriter struct {
	header http.Header
	status int
}

func (w *benchmarkResponseWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}

	return w.header
}

func (w *benchmarkResponseWriter) WriteHeader(status int) {
	w.status = status
}

func (w *benchmarkResponseWriter) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}

	return len(body), nil
}

func (w *benchmarkResponseWriter) ReadFrom(reader io.Reader) (int64, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}

	return io.Copy(io.Discard, reader)
}

func (w *benchmarkResponseWriter) Flush() {}

type benchmarkReader struct {
	reader bytes.Reader
}

func (r *benchmarkReader) reset(body []byte) {
	r.reader.Reset(body)
}

func (r *benchmarkReader) Read(body []byte) (int, error) {
	return r.reader.Read(body)
}

func BenchmarkRelayBodyNonStreaming(b *testing.B) {
	payload := bytes.Repeat([]byte("x"), benchmarkPayloadSize)
	metrics := newHTTPMetrics("benchmark")
	handler := metrics.wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		var reader benchmarkReader
		reader.reset(payload)

		if err := relayBody(w, &reader, false); err != nil {
			b.Fatalf("relay body: %v", err)
		}
	}))
	request := &http.Request{Method: http.MethodPost, URL: &url.URL{Path: "/dispatch/raw"}}
	writer := &benchmarkResponseWriter{}

	b.ReportAllocs()
	b.SetBytes(int64(len(payload)))
	b.ResetTimer()

	for range b.N {
		writer.status = 0
		handler.ServeHTTP(writer, request)
	}
}

func BenchmarkRelayBodyStreaming(b *testing.B) {
	payload := bytes.Repeat([]byte("x"), benchmarkPayloadSize)
	writer := &benchmarkResponseWriter{}

	b.ReportAllocs()
	b.SetBytes(int64(len(payload)))
	b.ResetTimer()

	for range b.N {
		var reader benchmarkReader
		reader.reset(payload)

		writer.status = 0
		if err := relayBody(writer, &reader, true); err != nil {
			b.Fatalf("relay body: %v", err)
		}
	}
}

func BenchmarkHTTPMetricsParallel(b *testing.B) {
	metrics := newHTTPMetrics("benchmark")
	handler := metrics.wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	request := &http.Request{Method: http.MethodPost, URL: &url.URL{Path: "/dispatch/raw"}}

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		writer := &benchmarkResponseWriter{}
		for pb.Next() {
			writer.status = 0
			handler.ServeHTTP(writer, request)
		}
	})
}

func BenchmarkAnthropicSSEAdaptation(b *testing.B) {
	var source strings.Builder
	source.WriteString("event: message_start\ndata: {\"type\":\"message_start\",\"message\":{\"id\":\"msg_benchmark\",\"model\":\"claude-benchmark\",\"usage\":{\"input_tokens\":10}}}\n\n")

	for range 100 {
		source.WriteString("event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"benchmark token\"}}\n\n")
	}

	source.WriteString("event: message_delta\ndata: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"end_turn\"},\"usage\":{\"output_tokens\":100}}\n\n")
	source.WriteString("event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n")
	stream := source.String()

	b.ReportAllocs()
	b.SetBytes(int64(len(stream)))
	b.ResetTimer()

	for range b.N {
		response := &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(stream)),
		}

		adapted := adaptAnthropicStreamResponse(response, "claude-benchmark")
		if _, err := io.Copy(io.Discard, adapted.Body); err != nil {
			b.Fatalf("copy adapted stream: %v", err)
		}

		if err := adapted.Body.Close(); err != nil {
			b.Fatalf("close adapted stream: %v", err)
		}
	}
}

func BenchmarkConvertAnthropicResponse(b *testing.B) {
	response := anthropicResponse{
		ID:    "msg_benchmark",
		Model: "claude-benchmark",
	}

	response.Content = make([]anthropicContentBlock, 1000)
	for index := range response.Content {
		response.Content[index] = anthropicContentBlock{Type: "text", Text: "benchmark response block"}
	}

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		converted := convertToOpenAIResponse(response, "claude-benchmark")
		if converted.ID == "" {
			b.Fatal("converted response is empty")
		}
	}
}

func BenchmarkNormalizeMistralRequestWithoutAliases(b *testing.B) {
	body := []byte(`{"model":"mistral-large","messages":[{"role":"user","content":"benchmark"}]}`)

	b.ReportAllocs()
	b.SetBytes(int64(len(body)))
	b.ResetTimer()

	for range b.N {
		if _, err := normalizeMistralJSONRequest(platform.ProviderOperationChatCompletions, body); err != nil {
			b.Fatalf("normalize Mistral request: %v", err)
		}
	}
}

func BenchmarkParseAudioMultipartBuffered(b *testing.B) {
	body, contentType := benchmarkAudioMultipartBody(b, 4*1024*1024)
	payload := platform.TunnelRequest{
		Headers: map[string]string{"Content-Type": contentType},
		Body:    body,
	}

	b.ReportAllocs()
	b.SetBytes(int64(len(body)))
	b.ResetTimer()

	for range b.N {
		form, err := parseAudioMultipartPayload(payload)
		if err != nil {
			b.Fatalf("parse buffered multipart: %v", err)
		}

		if len(form.File) == 0 {
			b.Fatal("parsed file is empty")
		}
	}
}

func BenchmarkParseAudioMultipartStreaming(b *testing.B) {
	body, contentType := benchmarkAudioMultipartBody(b, 4*1024*1024)
	payload := platform.TunnelRequest{
		Headers: map[string]string{"Content-Type": contentType},
	}

	b.ReportAllocs()
	b.SetBytes(int64(len(body)))
	b.ResetTimer()

	for range b.N {
		_, file, err := parseAudioMultipartStream(payload, bytes.NewReader(body))
		if err != nil {
			b.Fatalf("parse streaming multipart: %v", err)
		}

		if _, err := io.Copy(io.Discard, file); err != nil {
			b.Fatalf("consume streaming file: %v", err)
		}
	}
}

func benchmarkAudioMultipartBody(b *testing.B, size int) ([]byte, string) {
	b.Helper()

	var body bytes.Buffer

	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("language", "fa"); err != nil {
		b.Fatalf("write language: %v", err)
	}

	if err := writer.WriteField("model", "benchmark"); err != nil {
		b.Fatalf("write model: %v", err)
	}

	part, err := writer.CreateFormFile("file", "benchmark.wav")
	if err != nil {
		b.Fatalf("create file: %v", err)
	}

	if _, err := io.CopyN(part, zeroReader{}, int64(size)); err != nil {
		b.Fatalf("write file: %v", err)
	}

	if err := writer.Close(); err != nil {
		b.Fatalf("close multipart writer: %v", err)
	}

	return body.Bytes(), writer.FormDataContentType()
}

type zeroReader struct{}

func (zeroReader) Read(buffer []byte) (int, error) {
	clear(buffer)

	return len(buffer), nil
}
