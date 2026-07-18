package server

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
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

type deepgramAudioTranscriptionsAdapter struct{}

type googleVertexSpeechAdapter struct{}

type audioMultipartPayload struct {
	Fields      map[string][]string
	File        []byte
	ContentType string
	Filename    string
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

	target, err := resolveBaseURL(payload.Deployment.APIBase, "https://api.deepgram.com")
	if err != nil {
		return nil, err
	}

	endpoint, err := deepgramListenURL(target, audioProviderModel(payload.Deployment.Model), form)
	if err != nil {
		return errorAdapterResponse(http.StatusBadRequest, err.Error()), nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(form.File))
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
		return &adapterResponse{StatusCode: resp.StatusCode, Headers: resp.Header.Clone(), Body: resp.Body}, nil
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

	endpoint, err := googleVertexRecognizeURL(payload.Deployment)
	if err != nil {
		return errorAdapterResponse(http.StatusBadRequest, err.Error()), nil
	}

	body, err := googleVertexRecognizeBody(payload.Deployment, form)
	if err != nil {
		return errorAdapterResponse(http.StatusBadRequest, err.Error()), nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
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
		return &adapterResponse{StatusCode: resp.StatusCode, Headers: resp.Header.Clone(), Body: resp.Body}, nil
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
	if strings.TrimSpace(firstAudioFormValue(form.Fields, "language")) == "" {
		return errorAdapterResponse(http.StatusBadRequest, "language is required for cohere audio transcriptions"), nil
	}

	target, err := resolveBaseURL(payload.Deployment.APIBase, "https://api.cohere.com")
	if err != nil {
		return nil, err
	}

	body, contentType, err := cohereAudioTranscriptionsBody(payload.Deployment, form)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, joinURL(target, "/v2/audio/transcriptions"), bytes.NewReader(body))
	if err != nil {
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
	return &adapterResponse{StatusCode: resp.StatusCode, Headers: resp.Header.Clone(), Body: resp.Body}, nil
}

func parseAudioMultipartPayload(payload platform.TunnelRequest) (audioMultipartPayload, error) {
	contentType := strings.TrimSpace(payload.Headers["Content-Type"])
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return audioMultipartPayload{}, fmt.Errorf("invalid multipart content type: %w", err)
	}
	if !strings.EqualFold(mediaType, "multipart/form-data") || strings.TrimSpace(params["boundary"]) == "" {
		return audioMultipartPayload{}, fmt.Errorf("multipart/form-data content type is required")
	}

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
		data, err := io.ReadAll(part)
		if err != nil {
			return audioMultipartPayload{}, fmt.Errorf("read multipart field %q: %w", name, err)
		}

		if name == "file" {
			out.File = data
			out.Filename = part.FileName()
			out.ContentType = strings.TrimSpace(part.Header.Get("Content-Type"))
			continue
		}

		out.Fields[name] = append(out.Fields[name], string(data))
	}

	if len(out.File) == 0 {
		return audioMultipartPayload{}, fmt.Errorf("file is required")
	}

	return out, nil
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

func cohereAudioTranscriptionsBody(deployment platform.Deployment, form audioMultipartPayload) ([]byte, string, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	for key, values := range form.Fields {
		if key == "model" || key == "model[]" || strings.HasPrefix(key, "model[") {
			continue
		}
		for _, value := range values {
			if err := writer.WriteField(key, value); err != nil {
				_ = writer.Close()
				return nil, "", err
			}
		}
	}
	if err := writer.WriteField("model", stripProviderModelPrefix(deployment.Model, "cohere")); err != nil {
		_ = writer.Close()
		return nil, "", err
	}

	part, err := createAudioFilePart(writer, "file", form)
	if err != nil {
		_ = writer.Close()
		return nil, "", err
	}
	if _, err := part.Write(form.File); err != nil {
		_ = writer.Close()
		return nil, "", err
	}
	if err := writer.Close(); err != nil {
		return nil, "", err
	}

	return buf.Bytes(), writer.FormDataContentType(), nil
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

func googleVertexRecognizeBody(deployment platform.Deployment, form audioMultipartPayload) ([]byte, error) {
	config := map[string]interface{}{}
	if model := strings.TrimSpace(stripProviderModelPrefix(deployment.Model, "google-vertex")); model != "" {
		config["model"] = model
	}
	if language := strings.TrimSpace(firstAudioFormValue(form.Fields, "language")); language != "" {
		config["languageCodes"] = []string{language}
	}

	body := map[string]interface{}{
		"content": base64.StdEncoding.EncodeToString(form.File),
	}
	if len(config) > 0 {
		body["config"] = config
	}

	return json.Marshal(body)
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

	headers := cloneJSONHeaders(resp.Header)
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

	headers := cloneJSONHeaders(resp.Header)
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

func cloneJSONHeaders(headers http.Header) http.Header {
	out := headers.Clone()
	out.Set("Content-Type", "application/json")
	return out
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
