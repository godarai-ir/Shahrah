package server

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

const audioTranscriptionsPath = "/audio/transcriptions"

const (
	maxAudioFileBytes      = 25 * 1024 * 1024
	maxAudioFormFieldBytes = 64 * 1024
	maxAudioFormFields     = 64
	maxAudioFormBytes      = 1024 * 1024
)

type deepgramAudioTranscriptionsAdapter struct{}

type googleVertexSpeechAdapter struct{}

type audioMultipartPayload struct {
	Fields      map[string][]string
	File        []byte
	ContentType string
	Filename    string
}

type audioFormBudget struct {
	bytes  int
	fields int
}

func newDeepgramAudioTranscriptionsAdapter() providerAdapter {
	return &deepgramAudioTranscriptionsAdapter{}
}

func newGoogleVertexSpeechAdapter() providerAdapter {
	return &googleVertexSpeechAdapter{}
}

func (a *deepgramAudioTranscriptionsAdapter) Dispatch(ctx context.Context, client *http.Client, payload platform.TunnelRequest) (*adapterResponse, error) {
	if openAICompatiblePath(payload) != audioTranscriptionsPath {
		return errorAdapterResponse(http.StatusBadRequest, "deepgram adapter only supports audio transcriptions"), nil
	}
	if payload.Stream {
		return errorAdapterResponse(http.StatusBadRequest, "streaming audio transcriptions are not supported for provider: deepgram"), nil
	}

	form, err := parseAudioMultipartPayload(payload)
	if err != nil {
		return errorAdapterResponse(http.StatusBadRequest, err.Error()), nil
	}

	return dispatchDeepgramAudioTranscription(ctx, client, payload, form, bytes.NewReader(form.File))
}

func (a *deepgramAudioTranscriptionsAdapter) DispatchRaw(
	ctx context.Context,
	client *http.Client,
	payload platform.TunnelRequest,
	body io.Reader,
	_ int64,
) (*adapterResponse, error) {
	if openAICompatiblePath(payload) != audioTranscriptionsPath {
		return errorAdapterResponse(http.StatusBadRequest, "deepgram adapter only supports audio transcriptions"), nil
	}

	if payload.Stream {
		return errorAdapterResponse(http.StatusBadRequest, "streaming audio transcriptions are not supported for provider: deepgram"), nil
	}

	form, file, err := parseAudioMultipartStream(payload, body)
	if err != nil {
		return errorAdapterResponse(http.StatusBadRequest, err.Error()), nil
	}

	return dispatchDeepgramAudioTranscription(ctx, client, payload, form, file)
}

func dispatchDeepgramAudioTranscription(
	ctx context.Context,
	client *http.Client,
	payload platform.TunnelRequest,
	form audioMultipartPayload,
	file io.Reader,
) (*adapterResponse, error) {
	target, err := resolveBaseURL(payload.Deployment.APIBase, "https://api.deepgram.com")
	if err != nil {
		return nil, err
	}

	endpoint, err := deepgramListenURL(target, audioProviderModel(payload.Deployment.Model), form)
	if err != nil {
		return errorAdapterResponse(http.StatusBadRequest, err.Error()), nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, file)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", audioFileContentType(form))
	if apiKey := deploymentAPIKey(payload.Deployment); apiKey != "" {
		req.Header.Set("Authorization", "Token "+apiKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return &adapterResponse{StatusCode: resp.StatusCode, Headers: resp.Header, Body: resp.Body}, nil
	}

	return normalizeDeepgramTranscriptionResponse(resp)
}

func (a *googleVertexSpeechAdapter) Dispatch(ctx context.Context, client *http.Client, payload platform.TunnelRequest) (*adapterResponse, error) {
	if openAICompatiblePath(payload) != audioTranscriptionsPath {
		return errorAdapterResponse(http.StatusBadRequest, "google vertex speech adapter only supports audio transcriptions"), nil
	}
	if payload.Stream {
		return errorAdapterResponse(http.StatusBadRequest, "streaming audio transcriptions are not supported for provider: google-vertex"), nil
	}

	form, err := parseAudioMultipartPayload(payload)
	if err != nil {
		return errorAdapterResponse(http.StatusBadRequest, err.Error()), nil
	}

	return dispatchGoogleVertexTranscription(ctx, client, payload, form, bytes.NewReader(form.File))
}

func (a *googleVertexSpeechAdapter) DispatchRaw(
	ctx context.Context,
	client *http.Client,
	payload platform.TunnelRequest,
	body io.Reader,
	_ int64,
) (*adapterResponse, error) {
	if openAICompatiblePath(payload) != audioTranscriptionsPath {
		return errorAdapterResponse(http.StatusBadRequest, "google vertex speech adapter only supports audio transcriptions"), nil
	}

	if payload.Stream {
		return errorAdapterResponse(http.StatusBadRequest, "streaming audio transcriptions are not supported for provider: google-vertex"), nil
	}

	form, file, err := parseAudioMultipartStream(payload, body)
	if err != nil {
		return errorAdapterResponse(http.StatusBadRequest, err.Error()), nil
	}

	return dispatchGoogleVertexTranscription(ctx, client, payload, form, file)
}

func dispatchGoogleVertexTranscription(
	ctx context.Context,
	client *http.Client,
	payload platform.TunnelRequest,
	form audioMultipartPayload,
	file io.Reader,
) (*adapterResponse, error) {
	endpoint, err := googleVertexRecognizeURL(payload.Deployment)
	if err != nil {
		return errorAdapterResponse(http.StatusBadRequest, err.Error()), nil
	}

	body, err := googleVertexRecognizeBody(payload.Deployment, form, file)
	if err != nil {
		return errorAdapterResponse(http.StatusBadRequest, err.Error()), nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, body)
	if err != nil {
		closeBody(body)

		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if token := deploymentBearerCredential(payload.Deployment); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return &adapterResponse{StatusCode: resp.StatusCode, Headers: resp.Header, Body: resp.Body}, nil
	}

	return normalizeGoogleVertexTranscriptionResponse(resp)
}

func cohereAudioTranscriptionsDispatch(ctx context.Context, client *http.Client, payload platform.TunnelRequest) (*adapterResponse, error) {
	if payload.Stream {
		return errorAdapterResponse(http.StatusBadRequest, "streaming audio transcriptions are not supported for provider: cohere"), nil
	}

	form, err := parseAudioMultipartPayload(payload)
	if err != nil {
		return errorAdapterResponse(http.StatusBadRequest, err.Error()), nil
	}

	return dispatchCohereAudioTranscription(ctx, client, payload, form, bytes.NewReader(form.File))
}

func cohereAudioTranscriptionsDispatchRaw(
	ctx context.Context,
	client *http.Client,
	payload platform.TunnelRequest,
	body io.Reader,
) (*adapterResponse, error) {
	if payload.Stream {
		return errorAdapterResponse(http.StatusBadRequest, "streaming audio transcriptions are not supported for provider: cohere"), nil
	}

	form, file, err := parseAudioMultipartStream(payload, body)
	if err != nil {
		return errorAdapterResponse(http.StatusBadRequest, err.Error()), nil
	}

	return dispatchCohereAudioTranscription(ctx, client, payload, form, file)
}

func dispatchCohereAudioTranscription(
	ctx context.Context,
	client *http.Client,
	payload platform.TunnelRequest,
	form audioMultipartPayload,
	file io.Reader,
) (*adapterResponse, error) {
	if strings.TrimSpace(firstAudioFormValue(form.Fields, "language")) == "" {
		return errorAdapterResponse(http.StatusBadRequest, "language is required for cohere audio transcriptions"), nil
	}

	target, err := resolveBaseURL(payload.Deployment.APIBase, "https://api.cohere.com")
	if err != nil {
		return nil, err
	}

	body, contentType := cohereAudioTranscriptionsBody(payload.Deployment, form, file)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, joinURL(target, "/v2/audio/transcriptions"), body)
	if err != nil {
		closeBody(body)

		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	if apiKey := deploymentAPIKey(payload.Deployment); apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	applyDocumentedProviderHeaders(req, payload.Deployment)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return &adapterResponse{StatusCode: resp.StatusCode, Headers: resp.Header, Body: resp.Body}, nil
}

func parseAudioMultipartPayload(payload platform.TunnelRequest) (audioMultipartPayload, error) {
	contentType := strings.TrimSpace(headerValue(payload.Headers, "content-type"))
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return audioMultipartPayload{}, fmt.Errorf("invalid multipart content type: %w", err)
	}
	if !strings.EqualFold(mediaType, "multipart/form-data") || strings.TrimSpace(params["boundary"]) == "" {
		return audioMultipartPayload{}, fmt.Errorf("multipart/form-data content type is required")
	}

	var budget audioFormBudget

	reader := multipart.NewReader(bytes.NewReader(payload.Body), params["boundary"])
	out := audioMultipartPayload{Fields: map[string][]string{}}

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return audioMultipartPayload{}, fmt.Errorf("invalid multipart request body: %w", err)
		}

		name := strings.TrimSpace(part.FormName())
		if name == "" {
			continue
		}

		limit := int64(maxAudioFormFieldBytes)
		if name == "file" {
			limit = maxAudioFileBytes
		}

		data, err := readLimitedAudioPart(part, limit)
		if err != nil {
			return audioMultipartPayload{}, fmt.Errorf("read multipart field %q: %w", name, err)
		}

		if name == "file" {
			out.File = data
			out.Filename = part.FileName()
			out.ContentType = strings.TrimSpace(part.Header.Get("Content-Type"))
			continue
		}

		if err := appendAudioFormField(out.Fields, &budget, name, data); err != nil {
			return audioMultipartPayload{}, err
		}
	}

	if len(out.File) == 0 {
		return audioMultipartPayload{}, fmt.Errorf("file is required")
	}

	return out, nil
}

// parseAudioMultipartStream relies on the tunnel producer writing form metadata
// before the file part so the file can be relayed without buffering it first.
func parseAudioMultipartStream(
	payload platform.TunnelRequest,
	body io.Reader,
) (audioMultipartPayload, io.Reader, error) {
	contentType := strings.TrimSpace(headerValue(payload.Headers, "content-type"))

	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return audioMultipartPayload{}, nil, fmt.Errorf("invalid multipart content type: %w", err)
	}

	if !strings.EqualFold(mediaType, "multipart/form-data") || strings.TrimSpace(params["boundary"]) == "" {
		return audioMultipartPayload{}, nil, fmt.Errorf("multipart/form-data content type is required")
	}

	var budget audioFormBudget

	reader := multipart.NewReader(body, params["boundary"])
	out := audioMultipartPayload{Fields: map[string][]string{}}

	for {
		part, partErr := reader.NextPart()
		if partErr == io.EOF {
			return audioMultipartPayload{}, nil, fmt.Errorf("file is required")
		}

		if partErr != nil {
			return audioMultipartPayload{}, nil, fmt.Errorf("invalid multipart request body: %w", partErr)
		}

		name := strings.TrimSpace(part.FormName())
		if name == "" {
			continue
		}

		if name == "file" {
			var firstByte [1]byte
			if _, readErr := io.ReadFull(part, firstByte[:]); readErr != nil {
				return audioMultipartPayload{}, nil, fmt.Errorf("file is required")
			}

			out.Filename = part.FileName()
			out.ContentType = strings.TrimSpace(part.Header.Get("Content-Type"))
			file := io.MultiReader(bytes.NewReader(firstByte[:]), part)

			return out, &hardLimitReader{reader: file, remaining: maxAudioFileBytes}, nil
		}

		data, readErr := readLimitedAudioPart(part, maxAudioFormFieldBytes)
		if readErr != nil {
			return audioMultipartPayload{}, nil, fmt.Errorf("read multipart field %q: %w", name, readErr)
		}

		if err := appendAudioFormField(out.Fields, &budget, name, data); err != nil {
			return audioMultipartPayload{}, nil, err
		}
	}
}

func appendAudioFormField(
	fields map[string][]string,
	budget *audioFormBudget,
	name string,
	data []byte,
) error {
	budget.fields++
	if budget.fields > maxAudioFormFields {
		return fmt.Errorf("multipart form exceeds %d fields", maxAudioFormFields)
	}

	budget.bytes += len(name) + len(data)
	if budget.bytes > maxAudioFormBytes {
		return fmt.Errorf("multipart form fields exceed %d bytes", maxAudioFormBytes)
	}

	fields[name] = append(fields[name], string(data))

	return nil
}

func readLimitedAudioPart(part io.Reader, maximum int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(part, maximum+1))
	if err != nil {
		return nil, err
	}

	if int64(len(data)) > maximum {
		return nil, fmt.Errorf("multipart field exceeds %d bytes", maximum)
	}

	return data, nil
}

func deepgramListenURL(base string, model string, form audioMultipartPayload) (string, error) {
	model = strings.TrimSpace(stripProviderModelPrefix(model, "deepgram"))
	if model == "" {
		return "", fmt.Errorf("model is required")
	}

	endpoint, err := url.Parse(joinURL(base, "/v1/listen"))
	if err != nil {
		return "", err
	}
	query := endpoint.Query()
	query.Set("model", model)
	for _, key := range []string{"language"} {
		if value := strings.TrimSpace(firstAudioFormValue(form.Fields, key)); value != "" {
			query.Set(key, value)
		}
	}
	endpoint.RawQuery = query.Encode()

	return endpoint.String(), nil
}

func cohereAudioTranscriptionsBody(
	deployment platform.Deployment,
	form audioMultipartPayload,
	file io.Reader,
) (io.ReadCloser, string) {
	pipeReader, pipeWriter := io.Pipe()
	writer := multipart.NewWriter(pipeWriter)
	contentType := writer.FormDataContentType()

	go func() {
		writeErr := writeCohereAudioMultipart(writer, deployment, form, file)
		if writeErr == nil {
			writeErr = writer.Close()
		}

		if writeErr != nil {
			_ = pipeWriter.CloseWithError(writeErr)

			return
		}

		closeBody(pipeWriter)
	}()

	return pipeReader, contentType
}

func writeCohereAudioMultipart(
	writer *multipart.Writer,
	deployment platform.Deployment,
	form audioMultipartPayload,
	file io.Reader,
) error {
	for key, values := range form.Fields {
		if key == "model" || key == "model[]" || strings.HasPrefix(key, "model[") {
			continue
		}
		for _, value := range values {
			if err := writer.WriteField(key, value); err != nil {
				return err
			}
		}
	}
	if err := writer.WriteField("model", stripProviderModelPrefix(deployment.Model, "cohere")); err != nil {
		return err
	}

	part, err := createAudioFilePart(writer, "file", form)
	if err != nil {
		return err
	}

	_, err = io.Copy(part, file)

	return err
}

func googleVertexRecognizeURL(deployment platform.Deployment) (string, error) {
	recognizerID := strings.TrimSpace(deployment.RecognizerID)
	if recognizerID == "" {
		recognizerID = strings.TrimSpace(deployment.EndpointID)
	}
	if recognizerID == "" {
		return "", fmt.Errorf("recognizer_id is required for google-vertex audio transcriptions")
	}

	base, err := resolveBaseURL(deployment.APIBase, "https://speech.googleapis.com/v2")
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(strings.TrimLeft(recognizerID, "/"), "projects/") {
		resourcePath := "/" + strings.TrimLeft(recognizerID, "/")
		if !strings.HasSuffix(resourcePath, ":recognize") {
			resourcePath += ":recognize"
		}
		return joinURL(base, resourcePath), nil
	}

	projectID := strings.TrimSpace(deployment.ProjectID)
	if projectID == "" {
		return "", fmt.Errorf("project_id is required for google-vertex audio transcriptions")
	}

	location := strings.TrimSpace(deployment.RegionName)
	if location == "" {
		return "", fmt.Errorf("region_name is required for google-vertex audio transcriptions")
	}

	path := fmt.Sprintf(
		"/projects/%s/locations/%s/recognizers/%s:recognize",
		url.PathEscape(projectID),
		url.PathEscape(location),
		url.PathEscape(recognizerID),
	)

	return joinURL(base, path), nil
}

func createAudioFilePart(writer *multipart.Writer, field string, form audioMultipartPayload) (io.Writer, error) {
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, escapeMultipartQuoted(field), escapeMultipartQuoted(safeAudioFilename(form))))
	header.Set("Content-Type", audioFileContentType(form))

	return writer.CreatePart(header)
}

func googleVertexRecognizeBody(
	deployment platform.Deployment,
	form audioMultipartPayload,
	file io.Reader,
) (io.ReadCloser, error) {
	config := map[string]interface{}{}
	if model := strings.TrimSpace(stripProviderModelPrefix(deployment.Model, "google-vertex")); model != "" {
		config["model"] = model
	}
	if language := strings.TrimSpace(firstAudioFormValue(form.Fields, "language")); language != "" {
		config["languageCodes"] = []string{language}
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	pipeReader, pipeWriter := io.Pipe()
	go func() {
		writeErr := writeGoogleVertexRecognizeBody(pipeWriter, file, configJSON, len(config) > 0)
		if writeErr != nil {
			_ = pipeWriter.CloseWithError(writeErr)

			return
		}

		closeBody(pipeWriter)
	}()

	return pipeReader, nil
}

func writeGoogleVertexRecognizeBody(
	w io.Writer,
	file io.Reader,
	configJSON []byte,
	hasConfig bool,
) error {
	if _, err := io.WriteString(w, `{"content":"`); err != nil {
		return err
	}

	encoder := base64.NewEncoder(base64.StdEncoding, w)
	if _, err := io.Copy(encoder, file); err != nil {
		return errors.Join(err, encoder.Close())
	}

	if err := encoder.Close(); err != nil {
		return err
	}

	if hasConfig {
		if _, err := io.WriteString(w, `","config":`); err != nil {
			return err
		}

		if _, err := w.Write(configJSON); err != nil {
			return err
		}
	}

	_, err := io.WriteString(w, `}`)

	return err
}

func normalizeDeepgramTranscriptionResponse(resp *http.Response) (*adapterResponse, error) {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	text := deepgramTranscriptText(body)
	out, err := json.Marshal(map[string]interface{}{"text": text})
	if err != nil {
		return nil, err
	}

	headers := normalizeJSONHeaders(resp.Header)
	return &adapterResponse{StatusCode: resp.StatusCode, Headers: headers, Body: io.NopCloser(bytes.NewReader(out))}, nil
}

func normalizeGoogleVertexTranscriptionResponse(resp *http.Response) (*adapterResponse, error) {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	text := googleVertexTranscriptText(body)
	out, err := json.Marshal(map[string]interface{}{"text": text})
	if err != nil {
		return nil, err
	}

	headers := normalizeJSONHeaders(resp.Header)
	return &adapterResponse{StatusCode: resp.StatusCode, Headers: headers, Body: io.NopCloser(bytes.NewReader(out))}, nil
}

func deepgramTranscriptText(body []byte) string {
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return ""
	}

	results, _ := raw["results"].(map[string]interface{})
	channels, _ := results["channels"].([]interface{})
	if len(channels) == 0 {
		return ""
	}
	channel, _ := channels[0].(map[string]interface{})
	alternatives, _ := channel["alternatives"].([]interface{})
	if len(alternatives) == 0 {
		return ""
	}
	alternative, _ := alternatives[0].(map[string]interface{})
	text, _ := alternative["transcript"].(string)

	return text
}

func googleVertexTranscriptText(body []byte) string {
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return ""
	}

	results, _ := raw["results"].([]interface{})
	parts := make([]string, 0, len(results))
	for _, item := range results {
		result, _ := item.(map[string]interface{})
		alternatives, _ := result["alternatives"].([]interface{})
		if len(alternatives) == 0 {
			continue
		}
		alternative, _ := alternatives[0].(map[string]interface{})
		if transcript, ok := alternative["transcript"].(string); ok && strings.TrimSpace(transcript) != "" {
			parts = append(parts, transcript)
		}
	}

	return strings.Join(parts, "\n")
}

func errorAdapterResponse(status int, message string) *adapterResponse {
	body, _ := json.Marshal(map[string]interface{}{
		"error": map[string]interface{}{
			"message": message,
			"type":    "invalid_request_error",
		},
	})

	return &adapterResponse{
		StatusCode: status,
		Headers:    http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewReader(body)),
	}
}

func normalizeJSONHeaders(headers http.Header) http.Header {
	headers.Set("Content-Type", "application/json")

	return headers
}

func audioProviderModel(model string) string {
	model = strings.TrimSpace(model)
	if model == "" || model == "*" {
		return ""
	}

	return model
}

func deploymentBearerCredential(deployment platform.Deployment) string {
	if deployment.ProviderAuth != nil {
		switch deployment.ProviderAuth.Mode {
		case "api_key", "personal_access_token", "bearer", "google_service_account":
			return strings.TrimSpace(deployment.ProviderAuth.APIKey)
		default:
			return ""
		}
	}

	return strings.TrimSpace(deployment.APIKey)
}

func stripProviderModelPrefix(model string, provider string) string {
	model = strings.TrimSpace(model)
	prefix := strings.Trim(strings.ToLower(provider), "/") + "/"
	if strings.HasPrefix(strings.ToLower(model), prefix) {
		return strings.TrimSpace(model[len(prefix):])
	}

	return model
}

func firstAudioFormValue(values map[string][]string, key string) string {
	for _, candidate := range []string{key, key + "[]"} {
		if rawValues := values[candidate]; len(rawValues) > 0 {
			return rawValues[0]
		}
	}

	return ""
}

func audioFileContentType(form audioMultipartPayload) string {
	if contentType := strings.TrimSpace(form.ContentType); contentType != "" {
		return contentType
	}

	return "application/octet-stream"
}

func safeAudioFilename(form audioMultipartPayload) string {
	if filename := strings.TrimSpace(form.Filename); filename != "" {
		return filename
	}

	return "audio"
}

func escapeMultipartQuoted(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, `"`, `\"`)
	value = strings.ReplaceAll(value, "\r", " ")
	value = strings.ReplaceAll(value, "\n", " ")

	return value
}
