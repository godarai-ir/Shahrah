package server

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

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func TestGeminiNativeUsesNativePathAndAPIKeyHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertEqual(t, "/v1beta/models/gemini-3.7-flash:generateContent", r.URL.Path)
		assertEqual(t, "native-secret", r.Header.Get("x-goog-api-key"))
		assertEqual(t, "", r.Header.Get("Authorization"))

		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	response, err := newGeminiAdapter().Dispatch(context.Background(), server.Client(), geminiTestPayload(server.URL, platform.ProviderOperationGeminiGenerateContent, "gemini-3.7-flash"))
	assertNoError(t, err)
	assertEqual(t, http.StatusOK, response.StatusCode)
	_ = response.Body.Close()
}

func TestGeminiCompatibilityUsesOpenAIPathAndBearerAPIKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertEqual(t, "/v1beta/openai/chat/completions", r.URL.Path)
		assertEqual(t, "Bearer compat-secret", r.Header.Get("Authorization"))
		assertEqual(t, "", r.Header.Get("x-goog-api-key"))

		_, _ = w.Write([]byte(`{"choices":[]}`))
	}))
	defer server.Close()

	payload := geminiTestPayload(server.URL, platform.ProviderOperationChatCompletions, "gemini-3.5-flash")
	payload.Deployment.ProviderAuth.APIKey = "compat-secret"
	response, err := newGeminiAdapter().Dispatch(context.Background(), server.Client(), payload)
	assertNoError(t, err)

	_ = response.Body.Close()
}

func TestGeminiRawTunnelSupportsChatCompletions(t *testing.T) {
	t.Parallel()

	requestBody := []byte(`{"model":"gemini-3.5-flash-lite","messages":[{"role":"user","content":"hello"}],"stream":true}`)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer closeBody(r.Body)

		assertEqual(t, "/v1beta/openai/chat/completions", r.URL.Path)
		assertEqual(t, "Bearer compat-secret", r.Header.Get("Authorization"))
		assertEqual(t, "application/json", r.Header.Get("Content-Type"))

		body, err := io.ReadAll(r.Body)
		assertNoError(t, err)
		assertEqual(t, string(requestBody), string(body))

		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"choices\":[]}\n\n"))
	}))
	defer upstream.Close()

	payload := geminiTestPayload(upstream.URL, platform.ProviderOperationChatCompletions, "gemini-3.5-flash-lite")
	payload.Body = nil
	payload.Stream = true
	payload.Headers = map[string]string{"Content-Type": "application/json"}
	payload.Deployment.ProviderAuth.APIKey = "compat-secret"

	metadata, err := json.Marshal(payload)
	assertNoError(t, err)

	var prefix [4]byte
	binary.BigEndian.PutUint32(prefix[:], uint32(len(metadata)))

	request := httptest.NewRequest(
		http.MethodPost,
		"/dispatch/raw",
		io.MultiReader(bytes.NewReader(prefix[:]), bytes.NewReader(metadata), bytes.NewReader(requestBody)),
	)
	request.ContentLength = int64(len(prefix) + len(metadata) + len(requestBody))

	recorder := httptest.NewRecorder()
	NewHandler(upstream.Client()).ServeHTTP(recorder, request)

	assertEqual(t, http.StatusOK, recorder.Code)
	assertEqual(t, "data: {\"choices\":[]}\n\n", recorder.Body.String())
}

func TestGeminiRawSupportsCompatibleAndNativeOperations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		operation     platform.ProviderOperation
		model         string
		resourceID    string
		requestBody   string
		contentType   string
		wantPath      string
		wantAuthName  string
		wantAuthValue string
		responseBody  string
	}{
		{
			name:          "chat completions",
			operation:     platform.ProviderOperationChatCompletions,
			model:         "gemini-3.5-flash-lite",
			requestBody:   `{"model":"gemini-3.5-flash-lite","messages":[{"role":"user","content":"hello"}]}`,
			contentType:   "application/json",
			wantPath:      "/v1beta/openai/chat/completions",
			wantAuthName:  "Authorization",
			wantAuthValue: "Bearer native-secret",
			responseBody:  `{}`,
		},
		{
			name:          "embeddings",
			operation:     platform.ProviderOperationOpenAIEmbeddings,
			model:         "gemini-embedding-2",
			requestBody:   `{"model":"gemini-embedding-2","input":"hello"}`,
			contentType:   "application/json",
			wantPath:      "/v1beta/openai/embeddings",
			wantAuthName:  "Authorization",
			wantAuthValue: "Bearer native-secret",
			responseBody:  `{}`,
		},
		{
			name:          "image generation",
			operation:     platform.ProviderOperationOpenAIImagesGenerations,
			model:         "gemini-3.1-flash-image",
			requestBody:   `{"model":"gemini-3.1-flash-image","prompt":"hello"}`,
			contentType:   "application/json",
			wantPath:      "/v1beta/openai/images/generations",
			wantAuthName:  "Authorization",
			wantAuthValue: "Bearer native-secret",
			responseBody:  `{}`,
		},
		{
			name:          "model list",
			operation:     platform.ProviderOperationOpenAIModelsList,
			wantPath:      "/v1beta/openai/models",
			wantAuthName:  "Authorization",
			wantAuthValue: "Bearer native-secret",
			responseBody:  `{"object":"list","data":[]}`,
		},
		{
			name:          "model get",
			operation:     platform.ProviderOperationOpenAIModelsGet,
			resourceID:    "gemini-3.6-flash",
			wantPath:      "/v1beta/openai/models/gemini-3.6-flash",
			wantAuthName:  "Authorization",
			wantAuthValue: "Bearer native-secret",
			responseBody:  `{}`,
		},
		{
			name:          "native generate content",
			operation:     platform.ProviderOperationGeminiGenerateContent,
			model:         "gemini-3.6-flash",
			requestBody:   `{"contents":[{"parts":[{"text":"hello"}]}]}`,
			contentType:   "application/json",
			wantPath:      "/v1beta/models/gemini-3.6-flash:generateContent",
			wantAuthName:  "x-goog-api-key",
			wantAuthValue: "native-secret",
			responseBody:  `{}`,
		},
		{
			name:          "native file upload",
			operation:     platform.ProviderOperationGeminiFilesCreate,
			requestBody:   "raw-file",
			contentType:   "application/octet-stream",
			wantPath:      "/upload/v1beta/files",
			wantAuthName:  "x-goog-api-key",
			wantAuthValue: "native-secret",
			responseBody:  `{}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer closeBody(r.Body)

				assertEqual(t, test.wantPath, r.URL.Path)
				assertEqual(t, test.wantAuthValue, r.Header.Get(test.wantAuthName))
				assertEqual(t, test.contentType, r.Header.Get("Content-Type"))

				body, err := io.ReadAll(r.Body)
				assertNoError(t, err)
				assertEqual(t, test.requestBody, string(body))

				_, _ = w.Write([]byte(test.responseBody))
			}))
			defer upstream.Close()

			payload := geminiTestPayload(upstream.URL, test.operation, test.model)
			payload.OperationInput.ResourceID = test.resourceID
			payload.Headers = map[string]string{"Content-Type": test.contentType}

			adapter, ok := newGeminiAdapter().(rawProviderAdapter)
			if !ok {
				t.Fatal("Gemini adapter does not implement raw dispatch")
			}

			response, err := adapter.DispatchRaw(
				context.Background(),
				upstream.Client(),
				payload,
				strings.NewReader(test.requestBody),
				int64(len(test.requestBody)),
			)
			assertNoError(t, err)

			_, err = io.ReadAll(response.Body)
			assertNoError(t, err)
			assertNoError(t, response.Body.Close())
		})
	}
}

func TestGeminiOAuthUsesBearerForNativeAndCompatibility(t *testing.T) {
	paths := make([]string, 0, 2)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		assertEqual(t, "Bearer short-lived", r.Header.Get("Authorization"))
		assertEqual(t, "", r.Header.Get("x-goog-api-key"))

		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	for _, operation := range []platform.ProviderOperation{platform.ProviderOperationGeminiCountTokens, platform.ProviderOperationChatCompletions} {
		payload := geminiTestPayload(server.URL, operation, "gemini-3.6-flash")
		payload.Deployment.ProviderAuth = &platform.ResolvedProviderAuth{Mode: string(platform.ProviderAuthModeOAuthAccessToken), OAuthAccessToken: "short-lived"}
		response, err := newGeminiAdapter().Dispatch(context.Background(), server.Client(), payload)
		assertNoError(t, err)

		_ = response.Body.Close()
	}

	assertEqual(t, "/v1beta/models/gemini-3.6-flash:countTokens"+","+"/v1beta/openai/chat/completions", strings.Join(paths, ","))
}

func TestGeminiFailsClosedForUnknownModelAndUnsupportedCombination(t *testing.T) {
	adapter := newGeminiAdapter()
	_, err := adapter.Dispatch(context.Background(), http.DefaultClient, geminiTestPayload("https://example.test", platform.ProviderOperationChatCompletions, "gemini-latest"))
	assertErrorContains(t, err, "google_gemini_model_not_allowlisted")
	_, err = adapter.Dispatch(context.Background(), http.DefaultClient, geminiTestPayload("https://example.test", platform.ProviderOperationGeminiInteractionsCreate, "gemini-3.1-flash-lite-image"))
	assertErrorContains(t, err, "google_gemini_unsupported_model_operation")
	_, err = adapter.Dispatch(context.Background(), http.DefaultClient, geminiTestPayload("https://example.test", platform.ProviderOperationOpenAIEmbeddings, "gemini-3.5-flash"))
	assertErrorContains(t, err, "does not support embeddings")
}

func TestGeminiModelListFiltersNonAllowlistedModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"models":[{"name":"models/gemini-3.6-flash"},{"name":"models/gemini-2.5-pro"}],"nextPageToken":"next"}`))
	}))
	defer server.Close()

	response, err := newGeminiAdapter().Dispatch(context.Background(), server.Client(), geminiTestPayload(server.URL, platform.ProviderOperationGeminiModelsList, ""))
	assertNoError(t, err)
	body, err := io.ReadAll(response.Body)
	assertNoError(t, err)

	if !strings.Contains(string(body), "gemini-3.6-flash") || !strings.Contains(string(body), "nextPageToken") || strings.Contains(string(body), "gemini-2.5-pro") {
		t.Fatalf("unexpected filtered response: %s", body)
	}
}

func TestGeminiOpenAIModelListFiltersNonAllowlistedModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"object":"list","data":[{"id":"gemini-3.6-flash"},{"id":"gemini-2.5-pro"}]}`))
	}))
	defer server.Close()

	response, err := newGeminiAdapter().Dispatch(context.Background(), server.Client(), geminiTestPayload(server.URL, platform.ProviderOperationOpenAIModelsList, ""))
	assertNoError(t, err)
	body, err := io.ReadAll(response.Body)
	assertNoError(t, err)

	if !strings.Contains(string(body), "gemini-3.6-flash") || strings.Contains(string(body), "gemini-2.5-pro") {
		t.Fatalf("unexpected filtered response: %s", body)
	}
}

func TestGeminiOpenAIModelGetRejectsNonAllowlistedModel(t *testing.T) {
	payload := geminiTestPayload("https://example.test", platform.ProviderOperationOpenAIModelsGet, "")
	payload.OperationInput.ResourceID = "gemini-2.5-pro"
	_, err := newGeminiAdapter().Dispatch(context.Background(), http.DefaultClient, payload)
	assertErrorContains(t, err, "google_gemini_model_not_allowlisted")
}

func TestGeminiBaseURLNormalizationDoesNotDuplicateVersionOrOpenAI(t *testing.T) {
	assertEqual(t, "https://generativelanguage.googleapis.com", normalizeGeminiNativeBaseURL("https://generativelanguage.googleapis.com/v1beta/openai/"))
	assertEqual(t, "https://generativelanguage.googleapis.com/v1beta/openai", normalizeGeminiOpenAIBaseURL("https://generativelanguage.googleapis.com/v1beta"))
	assertEqual(t, "https://proxy.example/base/v1beta/openai", normalizeGeminiOpenAIBaseURL("https://proxy.example/base"))
}

func assertNoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatal(err)
	}
}
func assertEqual[T comparable](t *testing.T, expected, actual T) {
	t.Helper()

	if expected != actual {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
}
func assertErrorContains(t *testing.T, err error, value string) {
	t.Helper()

	if err == nil || !strings.Contains(err.Error(), value) {
		t.Fatalf("expected error containing %q, got %v", value, err)
	}
}

func geminiTestPayload(baseURL string, operation platform.ProviderOperation, model string) platform.TunnelRequest {
	body, _ := json.Marshal(map[string]interface{}{"model": model, "messages": []interface{}{}})

	payload := platform.TunnelRequest{
		Operation: operation,
		Body:      body,
		Deployment: platform.Deployment{
			Provider:     "google-gemini",
			Model:        model,
			APIBase:      strings.TrimRight(baseURL, "/"),
			ProviderAuth: &platform.ResolvedProviderAuth{Mode: string(platform.ProviderAuthModeAPIKey), APIKey: "native-secret"},
		},
	}
	if model != "" {
		payload.ModelRuntimeProfiles = []platform.ModelRuntimeProfileBinding{{
			Provider: "google-gemini", CanonicalID: "google-gemini/" + model, Slug: model,
			Profile: platform.ModelRuntimeProfile{
				Endpoints:  []string{"chat_completions", "embeddings", "image_generations", "generate_content", "stream_generate_content", "count_tokens", "embed_content", "batch_embed_contents", "interactions", "cached_contents", "batch_generate_content"},
				Transports: []string{"http", "sse"},
			},
		}}
	}

	return payload
}
