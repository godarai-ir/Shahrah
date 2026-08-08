package server

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func TestAudioAdaptersImplementRawStreaming(t *testing.T) {
	t.Parallel()

	registry := newProviderRegistry()
	for _, provider := range []string{"deepgram", "cohere", "google-vertex"} {
		adapter, err := registry.adapterFor(provider)
		if err != nil {
			t.Fatalf("resolve %s adapter: %v", provider, err)
		}

		if _, ok := adapter.(rawProviderAdapter); !ok {
			t.Fatalf("%s adapter does not implement rawProviderAdapter", provider)
		}
	}
}

func TestParseAudioMultipartStreamPreservesFieldsAndStreamsFile(t *testing.T) {
	t.Parallel()

	body, contentType := testAudioMultipartBody(t, []byte("streamed audio"))
	payload := platform.TunnelRequest{Headers: map[string]string{"Content-Type": contentType}}

	form, file, err := parseAudioMultipartStream(payload, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("parse streaming multipart: %v", err)
	}

	if got := firstAudioFormValue(form.Fields, "language"); got != "fa" {
		t.Fatalf("language = %q, want fa", got)
	}

	if form.Filename != "sample.wav" {
		t.Fatalf("filename = %q, want sample.wav", form.Filename)
	}

	fileBody, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("read streamed file: %v", err)
	}

	if string(fileBody) != "streamed audio" {
		t.Fatalf("file body = %q, want streamed audio", fileBody)
	}
}

func TestParseAudioMultipartStreamBoundsFormMetadata(t *testing.T) {
	t.Parallel()

	var body bytes.Buffer

	writer := multipart.NewWriter(&body)
	for index := range maxAudioFormFields + 1 {
		if err := writer.WriteField(fmt.Sprintf("field-%d", index), "value"); err != nil {
			t.Fatalf("write field: %v", err)
		}
	}

	file, createErr := writer.CreateFormFile("file", "sample.wav")
	if createErr != nil {
		t.Fatalf("create file field: %v", createErr)
	}

	_, writeErr := io.WriteString(file, "audio")
	if writeErr != nil {
		t.Fatalf("write file field: %v", writeErr)
	}

	closeErr := writer.Close()
	if closeErr != nil {
		t.Fatalf("close multipart writer: %v", closeErr)
	}

	payload := platform.TunnelRequest{
		Headers: map[string]string{"Content-Type": writer.FormDataContentType()},
	}

	_, _, err := parseAudioMultipartStream(payload, bytes.NewReader(body.Bytes()))
	if err == nil || !strings.Contains(err.Error(), "exceeds 64 fields") {
		t.Fatalf("expected bounded form metadata error, got %v", err)
	}
}

func TestGoogleVertexRecognizeBodyStreamsValidJSON(t *testing.T) {
	t.Parallel()

	form := audioMultipartPayload{Fields: map[string][]string{"language": {"fa"}}}

	body, err := googleVertexRecognizeBody(
		platform.Deployment{Model: "google-vertex/chirp"},
		form,
		strings.NewReader("audio"),
	)
	if err != nil {
		t.Fatalf("create Vertex request body: %v", err)
	}
	defer closeBody(body)

	var decoded struct {
		Content string `json:"content"`
		Config  struct {
			Model         string   `json:"model"`
			LanguageCodes []string `json:"languageCodes"`
		} `json:"config"`
	}
	if err := json.NewDecoder(body).Decode(&decoded); err != nil {
		t.Fatalf("decode Vertex request body: %v", err)
	}

	if decoded.Content != base64.StdEncoding.EncodeToString([]byte("audio")) {
		t.Fatalf("content = %q", decoded.Content)
	}

	if decoded.Config.Model != "chirp" || len(decoded.Config.LanguageCodes) != 1 || decoded.Config.LanguageCodes[0] != "fa" {
		t.Fatalf("unexpected Vertex config: %+v", decoded.Config)
	}
}

func TestCohereAudioTranscriptionsBodyStreamsMultipart(t *testing.T) {
	t.Parallel()

	form := audioMultipartPayload{
		Fields:      map[string][]string{"language": {"fa"}},
		Filename:    "sample.wav",
		ContentType: "audio/wav",
	}

	body, contentType := cohereAudioTranscriptionsBody(
		platform.Deployment{Model: "cohere/command-a-transcribe"},
		form,
		strings.NewReader("audio"),
	)
	defer closeBody(body)

	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		t.Fatalf("parse content type: %v", err)
	}

	reader := multipart.NewReader(body, params["boundary"])
	fields := make(map[string]string)

	for {
		part, partErr := reader.NextPart()
		if partErr == io.EOF {
			break
		}

		if partErr != nil {
			t.Fatalf("read multipart part: %v", partErr)
		}

		partBody, readErr := io.ReadAll(part)
		if readErr != nil {
			t.Fatalf("read multipart field: %v", readErr)
		}

		fields[part.FormName()] = string(partBody)
	}

	if fields["language"] != "fa" || fields["model"] != "command-a-transcribe" || fields["file"] != "audio" {
		t.Fatalf("unexpected multipart fields: %#v", fields)
	}
}

func TestCohereRawDispatchFallsBackForRerank(t *testing.T) {
	t.Parallel()

	requestBody := `{"model":"rerank-v3","query":"q","documents":["a"]}`

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/rerank" {
			t.Errorf("upstream path = %q, want /v2/rerank", r.URL.Path)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read upstream body: %v", err)
			return
		}

		if string(body) != requestBody {
			t.Errorf("upstream body = %q, want %q", body, requestBody)
		}

		w.Header().Set("Content-Type", "application/json")

		if _, err := io.WriteString(w, `{"results":[]}`); err != nil {
			t.Errorf("write upstream response: %v", err)
		}
	}))
	defer upstream.Close()

	adapter := &cohereAdapter{}
	payload := platform.TunnelRequest{
		Deployment: platform.Deployment{Provider: "cohere", APIBase: upstream.URL},
		RequestMetadata: map[string]interface{}{
			"provider_endpoint_path": "/v2/rerank",
		},
	}

	response, err := adapter.DispatchRaw(
		t.Context(),
		upstream.Client(),
		payload,
		strings.NewReader(requestBody),
		int64(len(requestBody)),
	)
	if err != nil {
		t.Fatalf("dispatch raw rerank: %v", err)
	}
	defer closeBody(response.Body)

	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.StatusCode)
	}
}

func TestGoogleVertexRecognizeBodyPreservesCopyAndCloseErrors(t *testing.T) {
	t.Parallel()

	copyErr := errors.New("read audio")
	closeErr := errors.New("write encoded audio")
	reader := &errorAfterDataReader{data: []byte("a"), err: copyErr}
	writer := &errorAfterBytesWriter{remaining: len(`{"content":"`), err: closeErr}

	err := writeGoogleVertexRecognizeBody(writer, reader, nil, false)
	if !errors.Is(err, copyErr) {
		t.Fatalf("error %v does not preserve copy error", err)
	}

	if !errors.Is(err, closeErr) {
		t.Fatalf("error %v does not preserve encoder close error", err)
	}
}

func TestReadBufferedProviderBodyRejectsKnownOversizeBody(t *testing.T) {
	t.Parallel()

	_, err := readBufferedProviderBody(strings.NewReader(""), maxBufferedProviderBodyBytes+1)
	if err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("expected body size error, got %v", err)
	}
}

func TestDispatchRejectsKnownOversizeEnvelopeWithoutReadingBody(t *testing.T) {
	t.Parallel()

	handler := NewHandler(nil)
	request := httptest.NewRequest(http.MethodPost, "/dispatch", strings.NewReader("{}"))
	request.ContentLength = maxJSONTunnelEnvelopeBytes + 1
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusRequestEntityTooLarge)
	}
}

func testAudioMultipartBody(t *testing.T, fileBody []byte) ([]byte, string) {
	t.Helper()

	var body bytes.Buffer

	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("language", "fa"); err != nil {
		t.Fatalf("write language: %v", err)
	}

	file, err := writer.CreateFormFile("file", "sample.wav")
	if err != nil {
		t.Fatalf("create file field: %v", err)
	}

	if _, err := file.Write(fileBody); err != nil {
		t.Fatalf("write file field: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	return body.Bytes(), writer.FormDataContentType()
}

type errorAfterDataReader struct {
	data []byte
	err  error
}

func (r *errorAfterDataReader) Read(buffer []byte) (int, error) {
	if len(r.data) == 0 {
		return 0, r.err
	}

	n := copy(buffer, r.data)
	r.data = r.data[n:]

	return n, nil
}

type errorAfterBytesWriter struct {
	remaining int
	err       error
}

func (w *errorAfterBytesWriter) Write(buffer []byte) (int, error) {
	if len(buffer) <= w.remaining {
		w.remaining -= len(buffer)

		return len(buffer), nil
	}

	return 0, w.err
}
