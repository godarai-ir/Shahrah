package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

const (
	geminiNativeBaseURL = "https://generativelanguage.googleapis.com"
	geminiOpenAIBaseURL = "https://generativelanguage.googleapis.com/v1beta/openai"
)

type geminiAdapter struct {
	compat *openAICompatibleAdapter
}

func newGeminiAdapter() providerAdapter {
	return &geminiAdapter{compat: newOpenAICompatibleAdapter(geminiOpenAIBaseURL)}
}

func (a *geminiAdapter) Dispatch(ctx context.Context, client *http.Client, payload platform.TunnelRequest) (*adapterResponse, error) {
	return a.dispatch(ctx, client, payload, bytes.NewReader(payload.Body), int64(len(payload.Body)), false)
}

func (a *geminiAdapter) DispatchRaw(
	ctx context.Context,
	client *http.Client,
	payload platform.TunnelRequest,
	body io.Reader,
	contentLength int64,
) (*adapterResponse, error) {
	return a.dispatch(ctx, client, payload, body, contentLength, true)
}

func (a *geminiAdapter) dispatch(
	ctx context.Context,
	client *http.Client,
	payload platform.TunnelRequest,
	body io.Reader,
	contentLength int64,
	raw bool,
) (*adapterResponse, error) {
	model := strings.TrimPrefix(strings.TrimSpace(payload.Deployment.Model), "models/")
	if model != "" && !isSupportedGeminiModel(model) {
		return nil, fmt.Errorf("google_gemini_model_not_allowlisted: %s", model)
	}

	if err := validateGeminiRuntimeProfile(payload, model); err != nil {
		return nil, err
	}

	if isGeminiCompatibleOperation(payload.Operation) {
		if payload.Operation == platform.ProviderOperationOpenAIModelsGet {
			requestedModel := strings.TrimPrefix(strings.TrimSpace(payload.OperationInput.ResourceID), "models/")
			if requestedModel == "" || !isSupportedGeminiModel(requestedModel) {
				return nil, fmt.Errorf("google_gemini_model_not_allowlisted: %s", requestedModel)
			}
		}

		if err := validateGeminiCompatibleModelOperation(model, payload.Operation); err != nil {
			return nil, err
		}

		compatPayload := payload

		compatPayload.Deployment.APIBase = normalizeGeminiOpenAIBaseURL(payload.Deployment.APIBase)
		if auth := payload.Deployment.ProviderAuth; auth != nil && auth.OAuthAccessToken != "" {
			resolved := *auth
			resolved.Mode = string(platform.ProviderAuthModeAPIKey)
			resolved.APIKey = auth.OAuthAccessToken
			compatPayload.Deployment.ProviderAuth = &resolved
		}

		var (
			response *adapterResponse
			err      error
		)

		if raw {
			response, err = a.compat.DispatchRaw(ctx, client, compatPayload, body, contentLength)
		} else {
			response, err = a.compat.Dispatch(ctx, client, compatPayload)
		}

		if err != nil {
			return nil, err
		}

		if payload.Operation == platform.ProviderOperationOpenAIModelsList && response.StatusCode < http.StatusBadRequest {
			return filterGeminiOpenAIModelListResponse(response)
		}

		return response, nil
	}

	if !isGeminiNativeOperation(payload.Operation) {
		return nil, fmt.Errorf("google_gemini_unsupported_operation: %s", payload.Operation)
	}

	if err := validateGeminiNativeModelOperation(model, payload.Operation); err != nil {
		return nil, err
	}

	method, path, err := geminiOperationTarget(payload.Operation, payload.OperationInput, model)
	if err != nil {
		return nil, err
	}

	target, err := resolveBaseURL(normalizeGeminiNativeBaseURL(payload.Deployment.APIBase), geminiNativeBaseURL)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(ctx, method, joinURL(target, path), body)
	if err != nil {
		return nil, err
	}

	if contentLength >= 0 {
		request.ContentLength = contentLength
	}

	if raw {
		if contentType := strings.TrimSpace(headerValue(payload.Headers, "content-type")); contentType != "" {
			request.Header.Set("Content-Type", contentType)
		} else if contentLength != 0 && payload.Operation != platform.ProviderOperationGeminiFilesCreate {
			request.Header.Set("Content-Type", "application/json")
		}
	} else if contentLength != 0 {
		request.Header.Set("Content-Type", "application/json")
	}

	if payload.Stream || payload.Operation == platform.ProviderOperationGeminiStreamGenerateContent {
		request.Header.Set("Accept", "text/event-stream")
	}

	if authErr := applyGeminiAuth(request, payload.Deployment, false); authErr != nil {
		return nil, authErr
	}

	response, err := client.Do(request) //nolint:bodyclose // The server handler owns and closes the body transferred through adapterResponse.
	if err != nil {
		return nil, err
	}

	result := &adapterResponse{StatusCode: response.StatusCode, Headers: response.Header, Body: response.Body}
	if payload.Operation == platform.ProviderOperationGeminiModelsList && response.StatusCode < http.StatusBadRequest {
		return filterGeminiModelListResponse(result)
	}

	return result, nil
}

func isSupportedGeminiModel(model string) bool {
	return platform.IsSupportedGoogleGeminiModel(model)
}

func validateGeminiRuntimeProfile(payload platform.TunnelRequest, model string) error {
	endpoint := geminiRuntimeEndpoint(payload.Operation)
	if endpoint == "" || model == "" {
		return nil
	}

	binding, ok := platform.FindModelRuntimeProfile(payload.ModelRuntimeProfiles, model)
	if !ok {
		return fmt.Errorf("google_gemini_runtime_profile_missing: %s", model)
	}

	if !strings.EqualFold(binding.Provider, "google-gemini") && !strings.EqualFold(binding.Provider, "gemini") {
		return fmt.Errorf("google_gemini_runtime_profile_provider_mismatch")
	}

	if !binding.Profile.AllowsEndpoint(endpoint) {
		return fmt.Errorf("google_gemini_unsupported_model_operation: %s does not allow %s", model, endpoint)
	}

	transport := "http"
	if payload.Stream || payload.Operation == platform.ProviderOperationGeminiStreamGenerateContent {
		transport = "sse"
	}

	if !binding.Profile.AllowsTransport(transport) {
		return fmt.Errorf("google_gemini_unsupported_transport: %s", transport)
	}

	return nil
}

func geminiRuntimeEndpoint(operation platform.ProviderOperation) string {
	switch operation {
	case "", platform.ProviderOperationChatCompletions:
		return "chat_completions"
	case platform.ProviderOperationOpenAIEmbeddings:
		return "embeddings"
	case platform.ProviderOperationOpenAIImagesGenerations:
		return "image_generations"
	case platform.ProviderOperationGeminiGenerateContent:
		return "generate_content"
	case platform.ProviderOperationGeminiStreamGenerateContent:
		return "stream_generate_content"
	case platform.ProviderOperationGeminiCountTokens:
		return "count_tokens"
	case platform.ProviderOperationGeminiEmbedContent:
		return "embed_content"
	case platform.ProviderOperationGeminiBatchEmbedContents:
		return "batch_embed_contents"
	case platform.ProviderOperationGeminiInteractionsCreate, platform.ProviderOperationGeminiInteractionsGet,
		platform.ProviderOperationGeminiInteractionsDelete, platform.ProviderOperationGeminiInteractionsCancel:
		return "interactions"
	case platform.ProviderOperationGeminiCachesCreate, platform.ProviderOperationGeminiCachesList,
		platform.ProviderOperationGeminiCachesGet, platform.ProviderOperationGeminiCachesUpdate,
		platform.ProviderOperationGeminiCachesDelete:
		return "cached_contents"
	case platform.ProviderOperationGeminiBatchesCreate:
		return "batch_generate_content"
	default:
		return ""
	}
}

func isGeminiCompatibleOperation(operation platform.ProviderOperation) bool {
	switch operation {
	case "", platform.ProviderOperationChatCompletions, platform.ProviderOperationOpenAIEmbeddings,
		platform.ProviderOperationOpenAIImagesGenerations, platform.ProviderOperationOpenAIModelsList,
		platform.ProviderOperationOpenAIModelsGet:
		return true
	default:
		return false
	}
}

func validateGeminiCompatibleModelOperation(model string, operation platform.ProviderOperation) error {
	if (operation == "" || operation == platform.ProviderOperationChatCompletions) &&
		(model == "gemini-embedding-2" || strings.Contains(model, "-image") || strings.Contains(model, "-tts-")) {
		return fmt.Errorf("google_gemini_unsupported_model_operation: %s does not support chat completions", model)
	}

	if operation == platform.ProviderOperationOpenAIEmbeddings && model != "gemini-embedding-2" {
		return fmt.Errorf("google_gemini_unsupported_model_operation: %s does not support embeddings", model)
	}

	if operation == platform.ProviderOperationOpenAIImagesGenerations && model != "gemini-3-pro-image" && model != "gemini-3.1-flash-image" && model != "gemini-3.1-flash-lite-image" {
		return fmt.Errorf("google_gemini_unsupported_model_operation: %s does not support image generation", model)
	}

	return nil
}

func validateGeminiNativeModelOperation(model string, operation platform.ProviderOperation) error {
	if model == "" {
		return nil
	}

	textModel := isGeminiTextModel(model)
	interactionModel := textModel || model == "gemini-3-pro-image" || model == "gemini-3.1-flash-image" || model == "gemini-3.1-flash-tts-preview"

	switch operation {
	case platform.ProviderOperationGeminiGenerateContent, platform.ProviderOperationGeminiStreamGenerateContent,
		platform.ProviderOperationGeminiCountTokens, platform.ProviderOperationGeminiBatchesCreate:
		if model == "gemini-embedding-2" {
			return fmt.Errorf("google_gemini_unsupported_model_operation: %s is embedding-only", model)
		}
	case platform.ProviderOperationGeminiEmbedContent, platform.ProviderOperationGeminiBatchEmbedContents:
		if model != "gemini-embedding-2" {
			return fmt.Errorf("google_gemini_unsupported_model_operation: %s is not an embedding model", model)
		}
	case platform.ProviderOperationGeminiInteractionsCreate, platform.ProviderOperationGeminiInteractionsGet,
		platform.ProviderOperationGeminiInteractionsDelete, platform.ProviderOperationGeminiInteractionsCancel:
		if !interactionModel {
			return fmt.Errorf("google_gemini_unsupported_model_operation: %s does not support interactions", model)
		}
	case platform.ProviderOperationGeminiCachesCreate, platform.ProviderOperationGeminiCachesList,
		platform.ProviderOperationGeminiCachesGet, platform.ProviderOperationGeminiCachesUpdate,
		platform.ProviderOperationGeminiCachesDelete:
		if !textModel {
			return fmt.Errorf("google_gemini_unsupported_model_operation: %s does not support explicit context caching", model)
		}
	}

	return nil
}

func isGeminiTextModel(model string) bool {
	switch model {
	case "gemini-2.5-flash-lite", "gemini-2.5-flash", "gemini-2.5-pro",
		"gemini-3.1-pro-preview", "gemini-3.5-flash", "gemini-3.5-flash-lite",
		"gemini-3.6-flash", "gemini-3.7-flash":
		return true
	default:
		return false
	}
}

func isGeminiNativeOperation(operation platform.ProviderOperation) bool {
	switch operation {
	case platform.ProviderOperationGeminiGenerateContent, platform.ProviderOperationGeminiStreamGenerateContent,
		platform.ProviderOperationGeminiCountTokens, platform.ProviderOperationGeminiEmbedContent,
		platform.ProviderOperationGeminiBatchEmbedContents, platform.ProviderOperationGeminiInteractionsCreate,
		platform.ProviderOperationGeminiInteractionsGet, platform.ProviderOperationGeminiInteractionsDelete,
		platform.ProviderOperationGeminiInteractionsCancel, platform.ProviderOperationGeminiModelsList,
		platform.ProviderOperationGeminiModelsGet, platform.ProviderOperationGeminiFilesCreate,
		platform.ProviderOperationGeminiFilesList, platform.ProviderOperationGeminiFilesGet,
		platform.ProviderOperationGeminiFilesDelete, platform.ProviderOperationGeminiCachesCreate,
		platform.ProviderOperationGeminiCachesList, platform.ProviderOperationGeminiCachesGet,
		platform.ProviderOperationGeminiCachesUpdate, platform.ProviderOperationGeminiCachesDelete,
		platform.ProviderOperationGeminiBatchesCreate, platform.ProviderOperationGeminiBatchesList,
		platform.ProviderOperationGeminiBatchesGet, platform.ProviderOperationGeminiBatchesCancel:
		return true
	default:
		return false
	}
}

func geminiOperationTarget(operation platform.ProviderOperation, input platform.ProviderOperationInput, model string) (string, string, error) {
	modelPath := func(action string) (string, error) {
		if model == "" || !isSupportedGeminiModel(model) {
			return "", fmt.Errorf("google_gemini_model_required_or_not_allowlisted")
		}

		return "/v1beta/models/" + url.PathEscape(model) + ":" + action, nil
	}
	resourcePath := func(prefix, suffix string) (string, error) {
		id := strings.TrimSpace(input.ResourceID)
		if id == "" {
			return "", fmt.Errorf("google_gemini_resource_id_required")
		}

		id = strings.TrimPrefix(id, prefix)

		return "/v1beta/" + prefix + url.PathEscape(id) + suffix, nil
	}

	switch operation {
	case platform.ProviderOperationGeminiGenerateContent:
		path, err := modelPath("generateContent")
		return http.MethodPost, path, err
	case platform.ProviderOperationGeminiStreamGenerateContent:
		path, err := modelPath("streamGenerateContent")
		return http.MethodPost, path + "?alt=sse", err
	case platform.ProviderOperationGeminiCountTokens:
		path, err := modelPath("countTokens")
		return http.MethodPost, path, err
	case platform.ProviderOperationGeminiEmbedContent:
		if model != "gemini-embedding-2" {
			return "", "", fmt.Errorf("google_gemini_unsupported_model_operation: embedContent")
		}

		path, err := modelPath("embedContent")

		return http.MethodPost, path, err
	case platform.ProviderOperationGeminiBatchEmbedContents:
		if model != "gemini-embedding-2" {
			return "", "", fmt.Errorf("google_gemini_unsupported_model_operation: batchEmbedContents")
		}

		path, err := modelPath("batchEmbedContents")

		return http.MethodPost, path, err
	case platform.ProviderOperationGeminiInteractionsCreate:
		return http.MethodPost, "/v1beta/interactions", nil
	case platform.ProviderOperationGeminiInteractionsGet:
		path, err := resourcePath("interactions/", "")
		return http.MethodGet, path, err
	case platform.ProviderOperationGeminiInteractionsDelete:
		path, err := resourcePath("interactions/", "")
		return http.MethodDelete, path, err
	case platform.ProviderOperationGeminiInteractionsCancel:
		path, err := resourcePath("interactions/", ":cancel")
		return http.MethodPost, path, err
	case platform.ProviderOperationGeminiModelsList:
		return http.MethodGet, "/v1beta/models", nil
	case platform.ProviderOperationGeminiModelsGet:
		id := strings.TrimPrefix(strings.TrimSpace(input.ResourceID), "models/")
		if !isSupportedGeminiModel(id) {
			return "", "", fmt.Errorf("google_gemini_model_not_allowlisted: %s", id)
		}

		return http.MethodGet, "/v1beta/models/" + url.PathEscape(id), nil
	case platform.ProviderOperationGeminiFilesCreate:
		return http.MethodPost, "/upload/v1beta/files", nil
	case platform.ProviderOperationGeminiFilesList:
		return http.MethodGet, "/v1beta/files", nil
	case platform.ProviderOperationGeminiFilesGet:
		path, err := resourcePath("files/", "")
		return http.MethodGet, path, err
	case platform.ProviderOperationGeminiFilesDelete:
		path, err := resourcePath("files/", "")
		return http.MethodDelete, path, err
	case platform.ProviderOperationGeminiCachesCreate:
		return http.MethodPost, "/v1beta/cachedContents", nil
	case platform.ProviderOperationGeminiCachesList:
		return http.MethodGet, "/v1beta/cachedContents", nil
	case platform.ProviderOperationGeminiCachesGet:
		path, err := resourcePath("cachedContents/", "")
		return http.MethodGet, path, err
	case platform.ProviderOperationGeminiCachesUpdate:
		path, err := resourcePath("cachedContents/", "")
		return http.MethodPatch, path, err
	case platform.ProviderOperationGeminiCachesDelete:
		path, err := resourcePath("cachedContents/", "")
		return http.MethodDelete, path, err
	case platform.ProviderOperationGeminiBatchesCreate:
		path, err := modelPath("batchGenerateContent")
		return http.MethodPost, path, err
	case platform.ProviderOperationGeminiBatchesList:
		return http.MethodGet, "/v1beta/batches", nil
	case platform.ProviderOperationGeminiBatchesGet:
		path, err := resourcePath("batches/", "")
		return http.MethodGet, path, err
	case platform.ProviderOperationGeminiBatchesCancel:
		path, err := resourcePath("batches/", ":cancel")
		return http.MethodPost, path, err
	default:
		return "", "", fmt.Errorf("google_gemini_unsupported_operation: %s", operation)
	}
}

func applyGeminiAuth(request *http.Request, deployment platform.Deployment, compatible bool) error {
	if deployment.ProviderAuth != nil && deployment.ProviderAuth.OAuthAccessToken != "" {
		request.Header.Set("Authorization", "Bearer "+deployment.ProviderAuth.OAuthAccessToken)
		return nil
	}

	apiKey := deploymentAPIKey(deployment)
	if apiKey == "" {
		return fmt.Errorf("google_gemini_credentials_missing")
	}

	if compatible {
		request.Header.Set("Authorization", "Bearer "+apiKey)
	} else {
		request.Header.Set("x-goog-api-key", apiKey)
	}

	return nil
}

func normalizeGeminiNativeBaseURL(value string) string {
	base := strings.TrimRight(strings.TrimSpace(value), "/")

	lower := strings.ToLower(base)
	for _, suffix := range []string{"/v1beta/openai", "/v1/openai", "/v1beta", "/v1"} {
		if strings.HasSuffix(lower, suffix) {
			return base[:len(base)-len(suffix)]
		}
	}

	return base
}

func normalizeGeminiOpenAIBaseURL(value string) string {
	base := strings.TrimRight(strings.TrimSpace(value), "/")
	if base == "" {
		return geminiOpenAIBaseURL
	}

	lower := strings.ToLower(base)
	if strings.HasSuffix(lower, "/openai") {
		return base
	}

	if strings.HasSuffix(lower, "/v1") || strings.HasSuffix(lower, "/v1beta") {
		return base + "/openai"
	}

	return base + "/v1beta/openai"
}

func filterGeminiModelListResponse(response *adapterResponse) (*adapterResponse, error) {
	body, err := io.ReadAll(response.Body)
	closeBody(response.Body)

	if err != nil {
		return nil, err
	}

	var envelope struct {
		Models        []json.RawMessage `json:"models"`
		NextPageToken string            `json:"nextPageToken,omitempty"`
	}
	if decodeErr := json.Unmarshal(body, &envelope); decodeErr != nil {
		return nil, fmt.Errorf("decode Gemini models response: %w", decodeErr)
	}

	filtered := make([]json.RawMessage, 0, len(envelope.Models))
	for _, item := range envelope.Models {
		var model struct {
			Name string `json:"name"`
		}
		if json.Unmarshal(item, &model) == nil && isSupportedGeminiModel(model.Name) {
			filtered = append(filtered, item)
		}
	}

	result, err := json.Marshal(struct {
		Models        []json.RawMessage `json:"models"`
		NextPageToken string            `json:"nextPageToken,omitempty"`
	}{Models: filtered, NextPageToken: envelope.NextPageToken})
	if err != nil {
		return nil, err
	}

	response.Headers.Del("Content-Length")
	response.Body = io.NopCloser(bytes.NewReader(result))

	return response, nil
}

func filterGeminiOpenAIModelListResponse(response *adapterResponse) (*adapterResponse, error) {
	body, err := io.ReadAll(response.Body)
	closeBody(response.Body)

	if err != nil {
		return nil, err
	}

	var envelope struct {
		Object string            `json:"object,omitempty"`
		Data   []json.RawMessage `json:"data"`
	}
	if decodeErr := json.Unmarshal(body, &envelope); decodeErr != nil {
		return nil, fmt.Errorf("decode Gemini OpenAI models response: %w", decodeErr)
	}

	filtered := make([]json.RawMessage, 0, len(envelope.Data))
	for _, item := range envelope.Data {
		var model struct {
			ID string `json:"id"`
		}
		if json.Unmarshal(item, &model) == nil && isSupportedGeminiModel(model.ID) {
			filtered = append(filtered, item)
		}
	}

	result, err := json.Marshal(struct {
		Object string            `json:"object,omitempty"`
		Data   []json.RawMessage `json:"data"`
	}{Object: envelope.Object, Data: filtered})
	if err != nil {
		return nil, err
	}

	response.Headers.Del("Content-Length")
	response.Body = io.NopCloser(bytes.NewReader(result))

	return response, nil
}
