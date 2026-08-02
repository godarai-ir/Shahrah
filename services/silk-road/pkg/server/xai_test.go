package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func TestXAIRegistryUsesDedicatedAdapter(t *testing.T) {
	adapter, err := newProviderRegistry().adapterFor("xai")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := adapter.(*xaiAdapter); !ok {
		t.Fatalf("expected dedicated xAI adapter, got %T", adapter)
	}
}

func TestXAIDispatchForwardsAllowlistedConversationHeader(t *testing.T) {
	var gotAuth, gotConversationID string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("unexpected path %s", r.URL.Path)
			return
		}
		gotAuth = r.Header.Get("Authorization")
		gotConversationID = r.Header.Get("x-grok-conv-id")
		w.Header().Set("x-request-id", "xai-request-1")
		_, _ = io.WriteString(w, `{"id":"chatcmpl_xai","usage":{"cost_in_usd_ticks":125000}}`)
	}))
	defer server.Close()

	payload := xaiTestPayload("grok-4.5", platform.ProviderOperationChatCompletions, `{"model":"grok-4.5-latest","messages":[{"role":"user","content":"hi"}],"reasoning_effort":"low"}`)
	payload.Deployment.APIBase = server.URL
	payload.Deployment.APIKey = "xai-secret"
	payload.Headers = map[string]string{
		"Content-Type":   "application/json",
		"X-Grok-Conv-Id": "conversation-123",
	}

	response, err := newXAIAdapter().Dispatch(context.Background(), server.Client(), payload)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if gotAuth != "Bearer xai-secret" || gotConversationID != "conversation-123" {
		t.Fatalf("unexpected forwarded headers auth=%q conversation=%q", gotAuth, gotConversationID)
	}
	if response.Headers.Get("x-request-id") != "xai-request-1" {
		t.Fatalf("expected xAI request id to be preserved")
	}
}

func TestXAIDispatchFailsClosedWithoutRuntimeProfile(t *testing.T) {
	payload := xaiTestPayload("grok-4.5", platform.ProviderOperationChatCompletions, `{"model":"grok-4.5","messages":[]}`)
	payload.ModelRuntimeProfiles = nil
	_, err := newXAIAdapter().Dispatch(context.Background(), http.DefaultClient, payload)
	if err == nil || !strings.Contains(err.Error(), "xai_runtime_profile_missing") {
		t.Fatalf("expected missing profile error, got %v", err)
	}
}

func TestXAIChatRejectsModelSpecificUnsupportedParameters(t *testing.T) {
	for name, test := range map[string]struct {
		body string
		want string
	}{
		"4.5 cannot disable reasoning": {
			body: `{"model":"grok-4.5","messages":[],"reasoning_effort":"none"}`,
			want: "xai_unsupported_reasoning_effort",
		},
		"logprobs are rejected": {
			body: `{"model":"grok-4.5","messages":[],"logprobs":true}`,
			want: "xai_unsupported_parameter: logprobs",
		},
		"presence penalty is rejected": {
			body: `{"model":"grok-4.5","messages":[],"presence_penalty":0.2}`,
			want: "xai_unsupported_parameter: presence_penalty",
		},
		"developer role is probe gated": {
			body: `{"model":"grok-4.5","messages":[{"role":"developer","content":"policy"}]}`,
			want: "xai_unsupported_message_role: developer",
		},
		"undocumented auto service tier is rejected": {
			body: `{"model":"grok-4.5","messages":[],"service_tier":"auto"}`,
			want: "xai_unsupported_service_tier: auto",
		},
		"unknown server tool is rejected": {
			body: `{"model":"grok-4.5","messages":[],"tools":[{"type":"unknown_search"}]}`,
			want: "xai_unsupported_tool",
		},
	} {
		t.Run(name, func(t *testing.T) {
			payload := xaiTestPayload("grok-4.5", platform.ProviderOperationChatCompletions, test.body)
			_, err := newXAIAdapter().Dispatch(context.Background(), http.DefaultClient, payload)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("expected %q, got %v", test.want, err)
			}
		})
	}
}

func TestXAIImageEditsRequireJSONAndVariationsAreRejected(t *testing.T) {
	payload := xaiTestPayload("grok-imagine-image-quality", platform.ProviderOperationOpenAIImagesEdits, `{"model":"grok-imagine-image-quality","prompt":"edit"}`)
	payload.Headers = map[string]string{"Content-Type": "multipart/form-data; boundary=test"}
	_, err := newXAIAdapter().Dispatch(context.Background(), http.DefaultClient, payload)
	if err == nil || !strings.Contains(err.Error(), "xai_image_edits_require_json") {
		t.Fatalf("expected JSON-only edit error, got %v", err)
	}

	payload.Operation = platform.ProviderOperationOpenAIImagesVariations
	_, err = newXAIAdapter().Dispatch(context.Background(), http.DefaultClient, payload)
	if err == nil || !strings.Contains(err.Error(), "xai_unsupported_operation") {
		t.Fatalf("expected variations rejection, got %v", err)
	}
}

func TestXAIResponsesEnforcesModelSpecificRolesAndParameters(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "developer role remains probe gated",
			body: `{"model":"grok-4.5","input":[{"role":"developer","content":"policy"}]}`,
			want: "xai_unsupported_message_role: developer",
		},
		{
			name: "4.5 presence penalty is rejected",
			body: `{"model":"grok-4.5","input":"hello","presence_penalty":0.2}`,
			want: "xai_unsupported_parameter: presence_penalty",
		},
		{
			name: "unknown server tool is rejected",
			body: `{"model":"grok-4.5","input":"hello","tools":[{"type":"unknown_search"}]}`,
			want: "xai_unsupported_tool: unknown_search",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payload := xaiTestPayload("grok-4.5", platform.ProviderOperationOpenAIResponsesCreate, test.body)
			_, err := newXAIAdapter().Dispatch(context.Background(), http.DefaultClient, payload)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("expected %q, got %v", test.want, err)
			}
		})
	}
}

func TestXAIImageAdapterEnforcesControls(t *testing.T) {
	tests := []struct {
		name      string
		operation platform.ProviderOperation
		body      string
		want      string
	}{
		{name: "invalid resolution", operation: platform.ProviderOperationOpenAIImagesGenerations, body: `{"model":"grok-imagine-image-quality","prompt":"hi","resolution":"4k"}`, want: "xai_unsupported_resolution"},
		{name: "invalid count", operation: platform.ProviderOperationOpenAIImagesGenerations, body: `{"model":"grok-imagine-image-quality","prompt":"hi","n":0}`, want: "xai_invalid_image_count"},
		{name: "streaming rejected", operation: platform.ProviderOperationOpenAIImagesGenerations, body: `{"model":"grok-imagine-image-quality","prompt":"hi","stream":true}`, want: "xai_image_streaming_unsupported"},
		{name: "edit reference required", operation: platform.ProviderOperationOpenAIImagesEdits, body: `{"model":"grok-imagine-image-quality","prompt":"hi"}`, want: "xai_image_reference_required"},
		{name: "too many references", operation: platform.ProviderOperationOpenAIImagesEdits, body: `{"model":"grok-imagine-image-quality","prompt":"hi","images":[{"url":"https://example.test/1.png"},{"url":"https://example.test/2.png"},{"url":"https://example.test/3.png"},{"url":"https://example.test/4.png"}]}`, want: "xai_too_many_image_references"},
		{name: "unknown parameter", operation: platform.ProviderOperationOpenAIImagesGenerations, body: `{"model":"grok-imagine-image-quality","prompt":"hi","quality":"hd"}`, want: "xai_unsupported_image_parameter: quality"},
		{name: "ambiguous reference", operation: platform.ProviderOperationOpenAIImagesEdits, body: `{"model":"grok-imagine-image-quality","prompt":"hi","image":{"url":"https://example.test/a.png","file_id":"file_a"}}`, want: "xai_invalid_image_reference"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payload := xaiTestPayload("grok-imagine-image-quality", test.operation, test.body)
			_, err := newXAIAdapter().Dispatch(context.Background(), http.DefaultClient, payload)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("expected %q, got %v", test.want, err)
			}
		})
	}
}

func TestXAIImageStorageOptionsEnforceProviderContract(t *testing.T) {
	for name, test := range map[string]struct {
		body string
		want string
	}{
		"filename is required": {
			body: `{"expires_after":3600}`,
			want: "xai_storage_filename_required",
		},
		"file expiry is bounded": {
			body: `{"filename":"image.png","expires_after":3599}`,
			want: "xai_invalid_storage_expiry",
		},
		"public URL cannot outlive file": {
			body: `{"filename":"image.png","expires_after":3600,"public_url":{"expires_after":7200}}`,
			want: "xai_public_url_outlives_file",
		},
		"unknown options are rejected": {
			body: `{"filename":"image.png","future_option":true}`,
			want: "xai_unsupported_storage_option",
		},
	} {
		t.Run(name, func(t *testing.T) {
			err := validateXAIImageStorageOptions([]byte(test.body))
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("expected %q, got %v", test.want, err)
			}
		})
	}
	for _, body := range []string{
		`{"filename":"image.png"}`,
		`{"filename":"image.png","public_url":true}`,
		`{"filename":"image.png","expires_after":7200,"public_url":{"expires_after":3600}}`,
	} {
		if err := validateXAIImageStorageOptions([]byte(body)); err != nil {
			t.Fatalf("valid storage_options rejected: %v", err)
		}
	}
}

func TestXAIRejectsArbitraryCompatibleOperations(t *testing.T) {
	payload := xaiTestPayload("grok-4.5", platform.ProviderOperationOpenAICompletions, `{"model":"grok-4.5"}`)
	_, err := newXAIAdapter().Dispatch(context.Background(), http.DefaultClient, payload)
	if err == nil || !strings.Contains(err.Error(), "xai_unsupported_operation") {
		t.Fatalf("expected fixed operation rejection, got %v", err)
	}
}

func TestXAIRawFileUploadAndPublicURLUseFixedPaths(t *testing.T) {
	requests := make(chan string, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		requests <- r.Method + " " + r.URL.Path + " " + string(body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{}`)
	}))
	defer server.Close()

	adapter := newXAIAdapter().(*xaiAdapter)
	upload := platform.TunnelRequest{
		Operation:  platform.ProviderOperationOpenAIFilesCreate,
		Deployment: platform.Deployment{Provider: "xai", APIBase: server.URL, APIKey: "secret"},
		Headers:    map[string]string{"Content-Type": "multipart/form-data; boundary=test"},
	}
	response, err := adapter.DispatchRaw(context.Background(), server.Client(), upload, strings.NewReader("raw-file"), int64(len("raw-file")))
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()

	publicURL := platform.TunnelRequest{
		Operation:      platform.ProviderOperationXAIFilePublicURLCreate,
		OperationInput: platform.ProviderOperationInput{ResourceID: "file_xai"},
		Deployment:     platform.Deployment{Provider: "xai", APIBase: server.URL, APIKey: "secret"},
		Headers:        map[string]string{"Content-Type": "application/json"},
		Body:           []byte(`{"expires_after":3600}`),
	}
	response, err = adapter.Dispatch(context.Background(), server.Client(), publicURL)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()

	if got := <-requests; got != "POST /files raw-file" {
		t.Fatalf("unexpected upload request %q", got)
	}
	if got := <-requests; got != `POST /files/file_xai/public-url {"expires_after":3600}` {
		t.Fatalf("unexpected public URL request %q", got)
	}
}

func TestXAIBatchOperationsUseNativeRequestAndCancelPaths(t *testing.T) {
	requests := make(chan string, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests <- r.Method + " " + r.URL.RequestURI()
		_, _ = io.WriteString(w, `{}`)
	}))
	defer server.Close()

	add := xaiTestPayload("grok-4.3", platform.ProviderOperationXAIBatchRequestsAdd, `{"batch_requests":[{"batch_request_id":"one","batch_request":{"chat_get_completion":{"model":"grok-4.3","messages":[]}}}]}`)
	add.Deployment.APIBase = server.URL
	add.OperationInput.ResourceID = "batch_1"
	response, err := newXAIAdapter().Dispatch(context.Background(), server.Client(), add)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()

	cancel := platform.TunnelRequest{
		Operation:      platform.ProviderOperationXAIBatchesCancel,
		OperationInput: platform.ProviderOperationInput{ResourceID: "batch_1"},
		Deployment:     platform.Deployment{Provider: "xai", APIBase: server.URL},
	}
	response, err = newXAIAdapter().Dispatch(context.Background(), server.Client(), cancel)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()

	if got := <-requests; got != "POST /batches/batch_1/requests" {
		t.Fatalf("unexpected add request path %q", got)
	}
	if got := <-requests; got != "POST /batches/batch_1:cancel" {
		t.Fatalf("unexpected cancel path %q", got)
	}
}

func TestXAIRetriesOnlyIdempotentGETOperations(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = io.WriteString(w, `{}`)
	}))
	defer server.Close()

	payload := platform.TunnelRequest{
		Operation:      platform.ProviderOperationOpenAIFilesGet,
		OperationInput: platform.ProviderOperationInput{ResourceID: "file_1"},
		Deployment:     platform.Deployment{Provider: "xai", APIBase: server.URL},
	}
	response, err := newXAIAdapter().Dispatch(context.Background(), server.Client(), payload)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if attempts != 2 {
		t.Fatalf("expected one safe retry, got %d attempts", attempts)
	}

	attempts = 0
	payload.Operation = platform.ProviderOperationXAIFilePublicURLCreate
	payload.Body = []byte(`{}`)
	response, err = newXAIAdapter().Dispatch(context.Background(), server.Client(), payload)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if attempts != 1 || response.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("POST must not be retried: attempts=%d status=%d", attempts, response.StatusCode)
	}
}

func xaiTestPayload(model string, operation platform.ProviderOperation, body string) platform.TunnelRequest {
	binding := xaiTestRuntimeProfile(model)
	return platform.TunnelRequest{
		Operation: operation,
		Deployment: platform.Deployment{
			Provider: "xai",
			Model:    model,
		},
		ModelRuntimeProfiles: []platform.ModelRuntimeProfileBinding{binding},
		Headers:              map[string]string{"Content-Type": "application/json"},
		Body:                 []byte(body),
	}
}

func xaiTestRuntimeProfile(model string) platform.ModelRuntimeProfileBinding {
	isImage := strings.HasPrefix(model, "grok-imagine-image")
	aliases := []string(nil)
	endpoints := []string{"models", "chat_completions", "deferred_chat_completions", "responses", "responses_compact", "files", "file_public_urls", "batches", "batch_requests", "batch_results"}
	transports := []string{"http", "sse"}
	efforts := []string{"low", "medium", "high"}
	tools := []string{"function_calling", "web_search", "x_search", "code_execution", "file_search", "mcp"}
	modalities := platform.ModelRuntimeModalities{Input: []string{"text", "image"}, Output: []string{"text"}}
	thinking := platform.ModelRuntimeThinkingProfile{DefaultMode: "adaptive", AllowedModes: []string{"adaptive"}, EffortValues: efforts, AlwaysEnabled: true}
	imageControls := platform.ModelRuntimeImageControls{}
	batchEndpoints := []string{"chat_completions"}
	if isImage {
		endpoints = []string{"models", "image_generation_models", "image_generations", "image_edits", "files", "file_public_urls", "batches", "batch_requests", "batch_results"}
		transports = []string{"http"}
		tools = nil
		modalities = platform.ModelRuntimeModalities{Input: []string{"text", "image"}, Output: []string{"image"}}
		thinking = platform.ModelRuntimeThinkingProfile{DefaultMode: "disabled", AllowedModes: []string{"disabled"}}
		imageControls = platform.ModelRuntimeImageControls{
			AspectRatios:           []string{"1:1", "16:9", "auto"},
			Resolutions:            []string{"1k", "2k"},
			ResponseFormats:        []string{"url", "b64_json"},
			MaximumOutputs:         10,
			MaximumReferenceImages: 3,
			EditTransport:          "json",
			FilesInput:             true,
			StorageOptions:         true,
		}
		batchEndpoints = []string{"image_generations", "image_edits"}
	}
	if model == "grok-4.5" {
		aliases = []string{"grok-4.5-latest", "grok-build-latest"}
	}
	profile := platform.ModelRuntimeProfile{
		Version:         3,
		Aliases:         aliases,
		Endpoints:       endpoints,
		Transports:      transports,
		Lifecycle:       "active",
		Thinking:        thinking,
		Sampling:        platform.ModelRuntimeSamplingProfile{Parameters: []string{"temperature", "top_p"}},
		Tools:           platform.ModelRuntimeToolsProfile{Custom: !isImage, Server: !isImage, Parallel: !isImage, StructuredOutput: !isImage, Supported: tools},
		Modalities:      modalities,
		RequestControls: platform.ModelRuntimeRequestControls{ServiceTierValues: []string{"default", "priority"}, InferenceGeoValues: []string{"global"}},
		RequestHeaders: platform.ModelRuntimeRequestHeaders{
			Allowed:    []string{"x-grok-conv-id"},
			ByEndpoint: map[string][]string{"chat_completions": {"x-grok-conv-id"}},
		},
		DeferredCompletions: platform.ModelRuntimeDeferredProfile{Eligible: !isImage, RetentionSeconds: 86400, MaximumRetrievals: 1},
		ImageControls:       imageControls,
		Batch: platform.ModelRuntimeBatchProfile{
			Eligible:           model != "grok-4.5",
			Discounted:         model == "grok-4.3",
			SupportedEndpoints: batchEndpoints,
			MaxRequestBytes:    25 * 1024 * 1024,
		},
		UnsupportedParameters: []string{"logprobs", "top_logprobs", "presence_penalty", "frequency_penalty", "stop"},
		Limits:                platform.ModelRuntimeLimits{MessagesBytes: 32 * 1024 * 1024, BatchBytes: 25 * 1024 * 1024, FileBytes: 48 * 1024 * 1024},
	}
	return platform.ModelRuntimeProfileBinding{
		Provider:        "xai",
		CatalogRevision: "xai-test-revision",
		CanonicalID:     "xai/" + model,
		Slug:            model,
		ProfileVersion:  3,
		Profile:         profile,
		SourceURL:       "https://docs.x.ai/developers/models",
		EffectiveFrom:   time.Now().UTC(),
	}
}
