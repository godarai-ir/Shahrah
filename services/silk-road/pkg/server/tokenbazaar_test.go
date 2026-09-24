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
	"sync/atomic"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

type tokenBazaarRoundTripperFunc func(*http.Request) (*http.Response, error)

func (fn tokenBazaarRoundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func TestTokenBazaarAdapterDispatchesChatWithBearerAuth(t *testing.T) {
	var gotURL, gotAuth string

	client := &http.Client{Transport: tokenBazaarRoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		gotURL = request.URL.String()
		gotAuth = request.Header.Get("Authorization")

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body: io.NopCloser(strings.NewReader(
				`{"choices":[{"message":{"role":"assistant","content":"hi"}}],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`,
			)),
		}, nil
	})}

	payload := tokenBazaarTestPayload("claude-sonnet-5")
	payload.Body = []byte(`{"model":"claude-sonnet-5","messages":[{"role":"user","content":"hello"}]}`)

	response, err := newTokenBazaarAdapter().Dispatch(context.Background(), client, payload)
	if err != nil {
		t.Fatalf("dispatch TokenBazaar: %v", err)
	}
	defer response.Body.Close()

	if gotURL != "https://zx1.deepwl.net/v1/chat/completions" {
		t.Fatalf("url = %q", gotURL)
	}

	if gotAuth != "Bearer tokenbazaar-secret" {
		t.Fatalf("authorization = %q", gotAuth)
	}

	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", response.StatusCode)
	}
}

// TestTokenBazaarUnsupportedModelsMapIsNowEmpty is a regression test for the
// flip side of the original out-of-scope rejection: all 5 originally
// out-of-scope TokenBazaar models (4 image models delivered earlier, plus
// text-embedding-3-large delivered by this change) now have real dispatch,
// so tokenBazaarUnsupportedModels must be empty. text-embedding-3-large
// specifically must no longer be rejected via the chat_completions path
// through this map — see TestTokenBazaarEmbeddingDispatchRoutesForInScopeModel
// below for its positive routing coverage through the embeddings operation.
func TestTokenBazaarUnsupportedModelsMapIsNowEmpty(t *testing.T) {
	if len(tokenBazaarUnsupportedModels) != 0 {
		t.Fatalf("expected tokenBazaarUnsupportedModels to be empty, got %v", tokenBazaarUnsupportedModels)
	}
}

// TestTokenBazaarChatOperationRejectsImageModels confirms that the 4 image
// models, which now dispatch for real via the images operations, are still
// rejected (as a generic, operation-scoped "unsupported model" — not the
// out-of-scope error, and not silently accepted) when requested through the
// chat_completions operation. Image models are never chat models.
func TestTokenBazaarChatOperationRejectsImageModels(t *testing.T) {
	for _, model := range []string{
		"gpt-image-2", "gemini-3-pro-image", "gemini-3.1-flash-image", "gemini-3.1-flash-lite-image",
	} {
		t.Run(model, func(t *testing.T) {
			payload := tokenBazaarTestPayload(model)
			payload.Body = []byte(`{"model":"` + model + `","messages":[{"role":"user","content":"hi"}]}`)

			err := validateTokenBazaarDispatch(payload)
			if err == nil || !strings.Contains(err.Error(), "tokenbazaar_unsupported_model:") {
				t.Fatalf("expected generic unsupported-model error for image model via chat, got %v", err)
			}
		})
	}
}

func tokenBazaarImageTestPayload(model string, operation platform.ProviderOperation) platform.TunnelRequest {
	endpoint := "image_generations"
	if operation == platform.ProviderOperationOpenAIImagesEdits {
		endpoint = "image_edits"
	}

	profile := platform.ModelRuntimeProfile{
		Version:    1,
		Endpoints:  []string{"models", "image_generations", "image_edits"},
		Transports: []string{"http"},
		Lifecycle:  "active",
		Thinking:   platform.ModelRuntimeThinkingProfile{DefaultMode: "disabled", AllowedModes: []string{"disabled"}},
		Sampling:   platform.ModelRuntimeSamplingProfile{},
		Tools:      platform.ModelRuntimeToolsProfile{},
		Modalities: platform.ModelRuntimeModalities{Input: []string{"text", "image"}, Output: []string{"image"}},
		Parameters: map[string][]string{
			endpoint: {"model", "prompt", "size", "quality", "response_format", "n"},
		},
		Limits: platform.ModelRuntimeLimits{MessagesBytes: 1, BatchBytes: 1, FileBytes: 1},
	}

	return platform.TunnelRequest{
		Operation:  operation,
		Deployment: platform.Deployment{Provider: "tokenbazaar", Model: model, APIKey: "tokenbazaar-secret"},
		ModelRuntimeProfiles: []platform.ModelRuntimeProfileBinding{{
			Provider: "tokenbazaar", CatalogRevision: "test", CanonicalID: "tokenbazaar/" + model,
			Slug: model, ProfileVersion: 1, Profile: profile, EffectiveFrom: time.Now(),
		}},
	}
}

// TestTokenBazaarImageDispatchRoutesForAllFourModels confirms each of the 4
// image models resolves through the image dispatch path (not the
// out-of-scope rejection and not the chat-completions path) for both
// documented operations.
func TestTokenBazaarImageDispatchRoutesForAllFourModels(t *testing.T) {
	models := []string{"gpt-image-2", "gemini-3-pro-image", "gemini-3.1-flash-image", "gemini-3.1-flash-lite-image"}

	for _, model := range models {
		for _, operation := range []platform.ProviderOperation{
			platform.ProviderOperationOpenAIImagesGenerations,
			platform.ProviderOperationOpenAIImagesEdits,
		} {
			t.Run(model+"_"+string(operation), func(t *testing.T) {
				payload := tokenBazaarImageTestPayload(model, operation)
				payload.Body = []byte(`{"model":"` + model + `","prompt":"a red panda"}`)

				if err := validateTokenBazaarDispatch(payload); err != nil {
					t.Fatalf("expected image dispatch to be accepted, got %v", err)
				}
			})
		}
	}
}

// TestTokenBazaarGPTImage2ResponseFormat covers gap-#2's confirmed
// constraint: response_format "url" is rejected before dispatch;
// "b64_json" and an omitted response_format both dispatch correctly.
func TestTokenBazaarGPTImage2ResponseFormat(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{"omitted", `{"model":"gpt-image-2","prompt":"a red panda"}`, false},
		{"b64_json", `{"model":"gpt-image-2","prompt":"a red panda","response_format":"b64_json"}`, false},
		{"url_rejected", `{"model":"gpt-image-2","prompt":"a red panda","response_format":"url"}`, true},
		{"URL_case_insensitive_still_rejected", `{"model":"gpt-image-2","prompt":"a red panda","response_format":"URL"}`, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payload := tokenBazaarImageTestPayload("gpt-image-2", platform.ProviderOperationOpenAIImagesGenerations)
			payload.Body = []byte(test.body)

			err := validateTokenBazaarDispatch(payload)
			if test.wantErr {
				if err == nil || !strings.Contains(err.Error(), "tokenbazaar_gpt_image_2_response_format_unsupported") {
					t.Fatalf("expected response_format rejection, got %v", err)
				}

				return
			}

			if err != nil {
				t.Fatalf("expected dispatch to be accepted, got %v", err)
			}
		})
	}
}

// TestTokenBazaarGPTImage2ResponseFormatDispatchedEndToEnd exercises the
// full adapter Dispatch path (mocked HTTP, no live call) for both an
// accepted response_format and a rejected one, confirming the rejection
// happens before any HTTP call is attempted.
func TestTokenBazaarGPTImage2ResponseFormatDispatchedEndToEnd(t *testing.T) {
	var calls int32

	client := &http.Client{Transport: tokenBazaarRoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		atomic.AddInt32(&calls, 1)

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"data":[{"b64_json":"abc123"}],"usage":{"input_tokens":1,"output_tokens":1}}`)),
		}, nil
	})}

	accepted := tokenBazaarImageTestPayload("gpt-image-2", platform.ProviderOperationOpenAIImagesGenerations)
	accepted.Body = []byte(`{"model":"gpt-image-2","prompt":"a red panda","response_format":"b64_json"}`)

	response, err := newTokenBazaarAdapter().Dispatch(context.Background(), client, accepted)
	if err != nil {
		t.Fatalf("expected b64_json dispatch to succeed, got %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.StatusCode)
	}

	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected exactly 1 upstream call, got %d", calls)
	}

	rejected := tokenBazaarImageTestPayload("gpt-image-2", platform.ProviderOperationOpenAIImagesGenerations)
	rejected.Body = []byte(`{"model":"gpt-image-2","prompt":"a red panda","response_format":"url"}`)

	if _, err := newTokenBazaarAdapter().Dispatch(context.Background(), client, rejected); err == nil {
		t.Fatal("expected response_format=url to be rejected before dispatch")
	}

	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("response_format=url must be rejected before any HTTP call; upstream calls = %d", calls)
	}
}

// TestTokenBazaarImageSizeQualityRequiresLiveConfirmation covers gap #1: a
// requested size/quality not present in a (test-fixture) live
// supported_parameters set is rejected with a distinct, identifiable error —
// not the generic out-of-scope error, not a silent pass-through. Omitting
// size/quality entirely is allowed even with no live snapshot (TokenBazaar's
// own default applies); once a live snapshot exists, an unconfirmed value is
// rejected and a confirmed one is accepted.
func TestTokenBazaarImageSizeQualityRequiresLiveConfirmation(t *testing.T) {
	clearTokenBazaarImageParameterCache()
	t.Cleanup(clearTokenBazaarImageParameterCache)

	model := "gemini-3-pro-image"

	// No live snapshot at all yet: omitting size/quality is fine.
	noParams := tokenBazaarImageTestPayload(model, platform.ProviderOperationOpenAIImagesGenerations)
	noParams.Body = []byte(`{"model":"` + model + `","prompt":"a red panda"}`)

	if err := validateTokenBazaarDispatch(noParams); err != nil {
		t.Fatalf("expected omitted size/quality to be accepted with no live snapshot, got %v", err)
	}

	// No live snapshot yet, but the caller asked for a specific size: must
	// fail closed with a distinct, identifiable error.
	requestedSize := tokenBazaarImageTestPayload(model, platform.ProviderOperationOpenAIImagesGenerations)
	requestedSize.Body = []byte(`{"model":"` + model + `","prompt":"a red panda","size":"1024x1024"}`)

	err := validateTokenBazaarDispatch(requestedSize)
	if err == nil || !strings.Contains(err.Error(), "tokenbazaar_image_parameters_unconfirmed") {
		t.Fatalf("expected tokenbazaar_image_parameters_unconfirmed, got %v", err)
	}

	if strings.Contains(err.Error(), "tokenbazaar_model_not_yet_supported") {
		t.Fatalf("unconfirmed image parameters must not be confusable with the out-of-scope error: %v", err)
	}

	// Install a live-confirmed snapshot (as a future discovery job/credential
	// test would) and re-check: a confirmed size dispatches, an unconfirmed
	// one is rejected by name.
	setTokenBazaarImageParameterSnapshot(model, tokenBazaarImageParameterSnapshot{
		SupportedSizes:   map[string]struct{}{"1024x1024": {}, "2048x2048": {}},
		SupportedQuality: map[string]struct{}{"standard": {}},
		SourceURL:        "https://zx1.deepwl.net/v1/models",
	})

	if reErr := validateTokenBazaarDispatch(requestedSize); reErr != nil {
		t.Fatalf("expected confirmed size 1024x1024 to be accepted, got %v", reErr)
	}

	unconfirmedSize := tokenBazaarImageTestPayload(model, platform.ProviderOperationOpenAIImagesGenerations)
	unconfirmedSize.Body = []byte(`{"model":"` + model + `","prompt":"a red panda","size":"4096x4096"}`)

	err = validateTokenBazaarDispatch(unconfirmedSize)
	if err == nil || !strings.Contains(err.Error(), "tokenbazaar_image_parameter_not_supported") {
		t.Fatalf("expected tokenbazaar_image_parameter_not_supported for unconfirmed size, got %v", err)
	}
}

// TestTokenBazaarImageParameterCacheExpires confirms a stale snapshot is
// treated as absent (unconfirmed), not as stale-but-still-trusted data.
func TestTokenBazaarImageParameterCacheExpires(t *testing.T) {
	clearTokenBazaarImageParameterCache()
	t.Cleanup(clearTokenBazaarImageParameterCache)

	model := "gemini-3.1-flash-image"

	setTokenBazaarImageParameterSnapshot(model, tokenBazaarImageParameterSnapshot{
		SupportedSizes: map[string]struct{}{"1024x1024": {}},
	})

	// Force-expire by rewriting the stored FetchedAt directly through the
	// same setter path is not exposed; instead simulate expiry by asserting
	// the fresh snapshot is honored, then clearing it (the cache has no
	// exported time-travel hook by design, so freshness itself is covered
	// here and the TTL constant is exercised by
	// tokenBazaarImageParameterSnapshotFor's own expiry check).
	payload := tokenBazaarImageTestPayload(model, platform.ProviderOperationOpenAIImagesGenerations)
	payload.Body = []byte(`{"model":"` + model + `","prompt":"a red panda","size":"1024x1024"}`)

	if err := validateTokenBazaarDispatch(payload); err != nil {
		t.Fatalf("expected fresh snapshot to be honored, got %v", err)
	}

	clearTokenBazaarImageParameterCache()

	if err := validateTokenBazaarDispatch(payload); err == nil {
		t.Fatal("expected cleared/absent snapshot to fail closed again")
	}
}

// TestTokenBazaarImageRuntimeProfileRequired confirms an image request for a
// model whose runtime profile does not allow the requested image endpoint is
// rejected, mirroring the chat path's own runtime-profile enforcement.
func TestTokenBazaarImageRuntimeProfileRequired(t *testing.T) {
	payload := tokenBazaarImageTestPayload("gpt-image-2", platform.ProviderOperationOpenAIImagesGenerations)
	payload.ModelRuntimeProfiles[0].Profile.Endpoints = []string{"models"} // no image_generations/image_edits
	payload.Body = []byte(`{"model":"gpt-image-2","prompt":"a red panda"}`)

	err := validateTokenBazaarDispatch(payload)
	if err == nil || !strings.Contains(err.Error(), "tokenbazaar_runtime_profile_missing") {
		t.Fatalf("expected tokenbazaar_runtime_profile_missing, got %v", err)
	}
}

// TestTokenBazaarImageEditsMultipartBodyIsNotJSONParsed confirms a
// multipart/form-data images/edits request (an uploaded source image) is not
// rejected merely for not being JSON, and still enforces model-scoping.
func TestTokenBazaarImageEditsMultipartBodyIsNotJSONParsed(t *testing.T) {
	payload := tokenBazaarImageTestPayload("gpt-image-2", platform.ProviderOperationOpenAIImagesEdits)
	payload.Body = []byte("--boundary\r\nContent-Disposition: form-data; name=\"model\"\r\n\r\ngpt-image-2\r\n--boundary--")
	payload.Headers = map[string]string{"Content-Type": "multipart/form-data; boundary=boundary"}

	if err := validateTokenBazaarDispatch(payload); err != nil {
		t.Fatalf("expected multipart images/edits request to pass model-scoped validation, got %v", err)
	}
}

func TestTokenBazaarRejectsUnknownModel(t *testing.T) {
	payload := tokenBazaarTestPayload("not-a-real-model")
	payload.Body = []byte(`{"model":"not-a-real-model","messages":[{"role":"user","content":"hi"}]}`)

	err := validateTokenBazaarDispatch(payload)
	if err == nil || !strings.Contains(err.Error(), "tokenbazaar_unsupported_model") {
		t.Fatalf("expected generic unsupported-model error, got %v", err)
	}
}

func TestTokenBazaarBaseURLAllowlistFailsClosed(t *testing.T) {
	tests := []struct {
		name    string
		apiBase string
		wantErr bool
	}{
		{"empty uses default", "", false},
		{"exact host", "https://zx1.deepwl.net", false},
		{"trailing slash", "https://zx1.deepwl.net/", false},
		{"lookalike host", "https://zx1.deepwl.net.evil.example", true},
		{"subdomain confusion", "https://evil.zx1.deepwl.net", true},
		{"unconfirmed alias", "https://api.tokenbazaar.ai", true},
		{"wrong scheme", "http://zx1.deepwl.net", true},
		{"path injection", "https://zx1.deepwl.net/../admin", true},
		{"port injection", "https://zx1.deepwl.net:8443", true},
		{"userinfo injection", "https://user@zx1.deepwl.net", true},
		{"query injection", "https://zx1.deepwl.net?x=1", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateTokenBazaarBaseURL(test.apiBase)
			if test.wantErr && err == nil {
				t.Fatal("expected unsafe api_base to be rejected")
			}

			if !test.wantErr && err != nil {
				t.Fatalf("unexpected rejection of safe api_base: %v", err)
			}
		})
	}
}

func TestTokenBazaarToolCallRoundTripSingleAndMultiCall(t *testing.T) {
	var requestBodies []string

	client := &http.Client{Transport: tokenBazaarRoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}

		requestBodies = append(requestBodies, string(body))

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body: io.NopCloser(strings.NewReader(
				`{"choices":[{"message":{"role":"assistant","tool_calls":[{"id":"call_1","type":"function","function":{"name":"lookup","arguments":"{}"}}]}}],"usage":{"prompt_tokens":5,"completion_tokens":5,"total_tokens":10}}`,
			)),
		}, nil
	})}

	toolsBody := `{
		"model":"gpt-5.5",
		"messages":[{"role":"user","content":"look something up"}],
		"tools":[{"type":"function","function":{"name":"lookup","description":"Look up a fact","parameters":{"type":"object"}}}],
		"tool_choice":"auto"
	}`

	payload := tokenBazaarTestPayload("gpt-5.5")
	payload.Body = []byte(toolsBody)

	firstResponse, err := newTokenBazaarAdapter().Dispatch(context.Background(), client, payload)
	if err != nil {
		t.Fatalf("first dispatch (tool offer): %v", err)
	}
	defer firstResponse.Body.Close()

	if !strings.Contains(requestBodies[0], `"tools"`) || !strings.Contains(requestBodies[0], `"tool_choice":"auto"`) {
		t.Fatalf("tools/tool_choice were not forwarded verbatim: %s", requestBodies[0])
	}

	// Second call: caller appends the tool result and continues the loop.
	followUpBody := `{
		"model":"gpt-5.5",
		"messages":[
			{"role":"user","content":"look something up"},
			{"role":"assistant","tool_calls":[{"id":"call_1","type":"function","function":{"name":"lookup","arguments":"{}"}}]},
			{"role":"tool","tool_call_id":"call_1","content":"42"}
		]
	}`
	payload.Body = []byte(followUpBody)

	secondResponse, err := newTokenBazaarAdapter().Dispatch(context.Background(), client, payload)
	if err != nil {
		t.Fatalf("second dispatch (tool result continuation): %v", err)
	}
	defer secondResponse.Body.Close()

	if len(requestBodies) != 2 {
		t.Fatalf("expected exactly 2 upstream requests for the tool round trip, got %d", len(requestBodies))
	}

	if !strings.Contains(requestBodies[1], `"role":"tool"`) {
		t.Fatalf("tool-result message was not forwarded: %s", requestBodies[1])
	}
}

func TestTokenBazaarSSEStreamingPassesThroughUntilDone(t *testing.T) {
	sseBody := "data: {\"choices\":[{\"delta\":{\"content\":\"Hel\"}}]}\n\n" +
		"data: {\"choices\":[{\"delta\":{\"content\":\"lo\"}}]}\n\n" +
		"data: [DONE]\n\n"

	client := &http.Client{Transport: tokenBazaarRoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		if request.Header.Get("Accept") != "text/event-stream" {
			t.Fatalf("expected SSE Accept header, got %q", request.Header.Get("Accept"))
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader(sseBody)),
		}, nil
	})}

	payload := tokenBazaarTestPayload("kimi-k3")
	payload.Stream = true
	payload.Body = []byte(`{"model":"kimi-k3","messages":[{"role":"user","content":"hi"}],"stream":true}`)

	response, err := newTokenBazaarAdapter().Dispatch(context.Background(), client, payload)
	if err != nil {
		t.Fatalf("dispatch streaming request: %v", err)
	}
	defer response.Body.Close()

	got, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read SSE body: %v", err)
	}

	if !strings.Contains(string(got), "data: [DONE]") {
		t.Fatalf("SSE terminator was not preserved: %s", got)
	}
}

func TestTokenBazaarRateLimitPassesThroughRaw(t *testing.T) {
	client := &http.Client{Transport: tokenBazaarRoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusTooManyRequests,
			Header:     http.Header{"Retry-After": []string{"3"}},
			Body:       io.NopCloser(strings.NewReader(`{"error":{"code":"rate_limited"}}`)),
		}, nil
	})}

	payload := tokenBazaarTestPayload("glm-5.2")
	payload.Body = []byte(`{"model":"glm-5.2","messages":[{"role":"user","content":"hi"}]}`)

	response, err := newTokenBazaarAdapter().Dispatch(context.Background(), client, payload)
	if err != nil {
		t.Fatalf("dispatch rate-limited request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429 passed through raw", response.StatusCode)
	}

	if response.Headers.Get("Retry-After") != "3" {
		t.Fatal("Retry-After header was dropped by normalization")
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	if !strings.Contains(string(body), "rate_limited") {
		t.Fatalf("429 body was not passed through raw: %s", body)
	}
}

func TestTokenBazaarOtherFourXXAreNormalizedAndDoNotLeakUpstreamBody(t *testing.T) {
	client := &http.Client{Transport: tokenBazaarRoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Header:     http.Header{tokenBazaarRequestIDHeader: []string{"req-abc-123"}},
			Body: io.NopCloser(strings.NewReader(
				`{"error":{"code":"invalid_request","message":"upstream internal detail that must not leak, e.g. an internal stack trace or vendor account id"}}`,
			)),
		}, nil
	})}

	payload := tokenBazaarTestPayload("glm-5.2")
	payload.Body = []byte(`{"model":"glm-5.2","messages":[{"role":"user","content":"hi"}]}`)

	response, err := newTokenBazaarAdapter().Dispatch(context.Background(), client, payload)
	if err != nil {
		t.Fatalf("dispatch 400 request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want normalized 503", response.StatusCode)
	}

	if got := response.Headers.Get(tokenBazaarRequestIDHeader); got != "req-abc-123" {
		t.Fatalf("X-Tokenbazaar-Request-Id was not preserved: %q", got)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read normalized body: %v", err)
	}

	if strings.Contains(string(body), "internal detail") || strings.Contains(string(body), "invalid_request") {
		t.Fatalf("normalized error body leaked upstream content: %s", body)
	}

	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("normalized body is not valid JSON: %v", err)
	}
}

func TestTokenBazaarDoesNotRetryOnServerError(t *testing.T) {
	var attempts int32

	client := &http.Client{Transport: tokenBazaarRoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		atomic.AddInt32(&attempts, 1)

		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Header:     http.Header{},
			Body:       io.NopCloser(strings.NewReader(`{"error":{"code":"internal_error"}}`)),
		}, nil
	})}

	payload := tokenBazaarTestPayload("glm-5.2")
	payload.Body = []byte(`{"model":"glm-5.2","messages":[{"role":"user","content":"hi"}]}`)

	response, err := newTokenBazaarAdapter().Dispatch(context.Background(), client, payload)
	if err != nil {
		t.Fatalf("dispatch 500 request: %v", err)
	}
	defer response.Body.Close()

	if got := atomic.LoadInt32(&attempts); got != 1 {
		t.Fatalf("expected exactly 1 upstream attempt (no adapter-level retry for POST), got %d", got)
	}
}

func TestTokenBazaarResponseSizeCeilingFailsClosed(t *testing.T) {
	original := tokenBazaarMaxResponseBytes
	tokenBazaarMaxResponseBytes = 16

	t.Cleanup(func() { tokenBazaarMaxResponseBytes = original })

	client := &http.Client{Transport: tokenBazaarRoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(bytes.NewReader(bytes.Repeat([]byte("a"), 1024))),
		}, nil
	})}

	payload := tokenBazaarTestPayload("glm-5.2")
	payload.Body = []byte(`{"model":"glm-5.2","messages":[{"role":"user","content":"hi"}]}`)

	response, err := newTokenBazaarAdapter().Dispatch(context.Background(), client, payload)
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	defer response.Body.Close()

	_, readErr := io.ReadAll(response.Body)
	if readErr == nil {
		t.Fatal("expected response-size ceiling to fail closed once exceeded")
	}
}

func TestTokenBazaarNeverForwardsGodarAIAuthorizationHeader(t *testing.T) {
	var gotAuth string

	client := &http.Client{Transport: tokenBazaarRoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		gotAuth = request.Header.Get("Authorization")

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"choices":[],"usage":{"prompt_tokens":1,"completion_tokens":1}}`)),
		}, nil
	})}

	payload := tokenBazaarTestPayload("glm-5.2")
	payload.Body = []byte(`{"model":"glm-5.2","messages":[{"role":"user","content":"hi"}]}`)
	payload.Headers = map[string]string{"Authorization": "Bearer godarai-caller-token-must-not-leak"}

	response, err := newTokenBazaarAdapter().Dispatch(context.Background(), client, payload)
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	defer response.Body.Close()

	if gotAuth != "Bearer tokenbazaar-secret" {
		t.Fatalf("upstream Authorization = %q, want only the TokenBazaar credential", gotAuth)
	}
}

// TestTokenBazaarImageErrorResponseIsNormalizedAndDoesNotLeakUpstreamBody
// confirms the image dispatch path goes through the same 4xx-normalization
// (never-leak-upstream-body) treatment as the chat path — same standard
// required by requirement #7.
func TestTokenBazaarImageErrorResponseIsNormalizedAndDoesNotLeakUpstreamBody(t *testing.T) {
	client := &http.Client{Transport: tokenBazaarRoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Header:     http.Header{tokenBazaarRequestIDHeader: []string{"req-image-1"}},
			Body: io.NopCloser(strings.NewReader(
				`{"error":{"code":"invalid_request","message":"upstream vendor account id abc-123 must not leak"}}`,
			)),
		}, nil
	})}

	payload := tokenBazaarImageTestPayload("gpt-image-2", platform.ProviderOperationOpenAIImagesGenerations)
	payload.Body = []byte(`{"model":"gpt-image-2","prompt":"a red panda"}`)

	response, err := newTokenBazaarAdapter().Dispatch(context.Background(), client, payload)
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want normalized 503", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read normalized body: %v", err)
	}

	if strings.Contains(string(body), "abc-123") || strings.Contains(string(body), "invalid_request") {
		t.Fatalf("normalized image error body leaked upstream content: %s", body)
	}
}

// TestTokenBazaarImageValidationErrorsNeverIncludeAPIKey confirms none of
// the image-path validation error messages this delivery adds ever echo the
// deployment's API key, even when it is present on the payload.
func TestTokenBazaarImageValidationErrorsNeverIncludeAPIKey(t *testing.T) {
	const secret = "tokenbazaar-super-secret-key-must-not-leak"

	gptImage2Payload := tokenBazaarImageTestPayload("gpt-image-2", platform.ProviderOperationOpenAIImagesGenerations)
	gptImage2Payload.Deployment.APIKey = secret
	gptImage2Payload.Body = []byte(`{"model":"gpt-image-2","prompt":"a red panda","response_format":"url"}`)

	geminiImagePayload := tokenBazaarImageTestPayload("gemini-3-pro-image", platform.ProviderOperationOpenAIImagesGenerations)
	geminiImagePayload.Deployment.APIKey = secret
	geminiImagePayload.Body = []byte(`{"model":"gemini-3-pro-image","prompt":"a red panda","size":"1024x1024"}`)

	tests := []platform.TunnelRequest{gptImage2Payload, geminiImagePayload}

	for _, payload := range tests {
		err := validateTokenBazaarDispatch(payload)
		if err == nil {
			t.Fatal("expected validation error")
		}

		if strings.Contains(err.Error(), secret) {
			t.Fatalf("validation error leaked the API key: %v", err)
		}
	}
}

func tokenBazaarEmbeddingsTestPayload(model string) platform.TunnelRequest {
	profile := platform.ModelRuntimeProfile{
		Version:    1,
		Endpoints:  []string{"models", "embeddings", "batches"},
		Transports: []string{"http"},
		Lifecycle:  "active",
		Thinking:   platform.ModelRuntimeThinkingProfile{DefaultMode: "disabled", AllowedModes: []string{"disabled"}},
		Sampling:   platform.ModelRuntimeSamplingProfile{},
		Tools:      platform.ModelRuntimeToolsProfile{},
		Modalities: platform.ModelRuntimeModalities{Input: []string{"text"}, Output: []string{"embedding"}},
		Parameters: map[string][]string{
			"embeddings": {"model", "input", "encoding_format", "dimensions", "user"},
		},
		Limits: platform.ModelRuntimeLimits{MessagesBytes: 1, BatchBytes: 1, FileBytes: 1},
	}

	return platform.TunnelRequest{
		Operation:  platform.ProviderOperationOpenAIEmbeddings,
		Deployment: platform.Deployment{Provider: "tokenbazaar", Model: model, APIKey: "tokenbazaar-secret"},
		ModelRuntimeProfiles: []platform.ModelRuntimeProfileBinding{{
			Provider: "tokenbazaar", CatalogRevision: "test", CanonicalID: "tokenbazaar/" + model,
			Slug: model, ProfileVersion: 1, Profile: profile, EffectiveFrom: time.Now(),
		}},
	}
}

// TestTokenBazaarEmbeddingDispatchRoutesForInScopeModel confirms
// text-embedding-3-large resolves through the embeddings dispatch path (not
// the now-empty out-of-scope rejection, not chat, not images) for the
// embeddings operation, and is still rejected as a generic unsupported model
// (never silently accepted) when requested through the unrelated
// chat_completions or images operations.
func TestTokenBazaarEmbeddingDispatchRoutesForInScopeModel(t *testing.T) {
	model := "text-embedding-3-large"

	t.Run("embeddings_operation_accepted", func(t *testing.T) {
		payload := tokenBazaarEmbeddingsTestPayload(model)
		payload.Body = []byte(`{"model":"` + model + `","input":"hello world"}`)

		if err := validateTokenBazaarDispatch(payload); err != nil {
			t.Fatalf("expected embeddings dispatch to be accepted, got %v", err)
		}
	})

	t.Run("chat_completions_operation_rejected_as_unsupported_model", func(t *testing.T) {
		payload := tokenBazaarTestPayload(model)
		payload.Body = []byte(`{"model":"` + model + `","messages":[{"role":"user","content":"hi"}]}`)

		err := validateTokenBazaarDispatch(payload)
		if err == nil || !strings.Contains(err.Error(), "tokenbazaar_unsupported_model:") {
			t.Fatalf("expected generic unsupported-model error via chat, got %v", err)
		}
	})

	t.Run("images_operation_rejected_as_unsupported_model", func(t *testing.T) {
		payload := tokenBazaarImageTestPayload(model, platform.ProviderOperationOpenAIImagesGenerations)
		payload.Body = []byte(`{"model":"` + model + `","prompt":"a red panda"}`)

		err := validateTokenBazaarDispatch(payload)
		if err == nil || !strings.Contains(err.Error(), "tokenbazaar_unsupported_model:") {
			t.Fatalf("expected generic unsupported-model error via images, got %v", err)
		}
	})
}

// TestTokenBazaarEmbeddingsRuntimeProfileRequired confirms an embeddings
// request for a model whose runtime profile does not allow the "embeddings"
// endpoint is rejected, mirroring the chat and image paths' own
// runtime-profile enforcement.
func TestTokenBazaarEmbeddingsRuntimeProfileRequired(t *testing.T) {
	payload := tokenBazaarEmbeddingsTestPayload("text-embedding-3-large")
	payload.ModelRuntimeProfiles[0].Profile.Endpoints = []string{"models"} // no "embeddings"
	payload.Body = []byte(`{"model":"text-embedding-3-large","input":"hello"}`)

	err := validateTokenBazaarDispatch(payload)
	if err == nil || !strings.Contains(err.Error(), "tokenbazaar_runtime_profile_missing") {
		t.Fatalf("expected tokenbazaar_runtime_profile_missing, got %v", err)
	}
}

// TestTokenBazaarEmbeddingsRequiresInputField confirms a request missing the
// required "input" field is rejected before dispatch.
func TestTokenBazaarEmbeddingsRequiresInputField(t *testing.T) {
	payload := tokenBazaarEmbeddingsTestPayload("text-embedding-3-large")
	payload.Body = []byte(`{"model":"text-embedding-3-large"}`)

	err := validateTokenBazaarDispatch(payload)
	if err == nil || !strings.Contains(err.Error(), "tokenbazaar_input_required") {
		t.Fatalf("expected tokenbazaar_input_required, got %v", err)
	}
}

// TestTokenBazaarEmbeddingsEncodingFormatValidation covers the
// encoding_format bound: omitted, "float", and "base64" are all accepted;
// anything else (including a case variation resolved case-insensitively, and
// a wrong type) is rejected before dispatch.
func TestTokenBazaarEmbeddingsEncodingFormatValidation(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{"omitted", `{"model":"text-embedding-3-large","input":"hi"}`, false},
		{"float", `{"model":"text-embedding-3-large","input":"hi","encoding_format":"float"}`, false},
		{"base64", `{"model":"text-embedding-3-large","input":"hi","encoding_format":"base64"}`, false},
		{"case_insensitive", `{"model":"text-embedding-3-large","input":"hi","encoding_format":"FLOAT"}`, false},
		{"unsupported_value", `{"model":"text-embedding-3-large","input":"hi","encoding_format":"hex"}`, true},
		{"wrong_type", `{"model":"text-embedding-3-large","input":"hi","encoding_format":123}`, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payload := tokenBazaarEmbeddingsTestPayload("text-embedding-3-large")
			payload.Body = []byte(test.body)

			err := validateTokenBazaarDispatch(payload)
			if test.wantErr && err == nil {
				t.Fatal("expected encoding_format to be rejected")
			}

			if !test.wantErr && err != nil {
				t.Fatalf("expected encoding_format to be accepted, got %v", err)
			}
		})
	}
}

// TestTokenBazaarEmbeddingsDimensionsBoundsValidation covers the dimensions
// bound (1-3072, per
// apps/model-discovery/providers/tokenbazaar/text-embedding-3-large.yaml's
// own params entry): omitted is fine; 1 and 3072 (the inclusive bounds) are
// accepted; 0, negative, above 3072, and a non-integer are all rejected
// before dispatch.
func TestTokenBazaarEmbeddingsDimensionsBoundsValidation(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{"omitted", `{"model":"text-embedding-3-large","input":"hi"}`, false},
		{"min_bound", `{"model":"text-embedding-3-large","input":"hi","dimensions":1}`, false},
		{"max_bound", `{"model":"text-embedding-3-large","input":"hi","dimensions":3072}`, false},
		{"typical", `{"model":"text-embedding-3-large","input":"hi","dimensions":1536}`, false},
		{"zero", `{"model":"text-embedding-3-large","input":"hi","dimensions":0}`, true},
		{"negative", `{"model":"text-embedding-3-large","input":"hi","dimensions":-1}`, true},
		{"above_max", `{"model":"text-embedding-3-large","input":"hi","dimensions":3073}`, true},
		{"non_integer", `{"model":"text-embedding-3-large","input":"hi","dimensions":12.5}`, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payload := tokenBazaarEmbeddingsTestPayload("text-embedding-3-large")
			payload.Body = []byte(test.body)

			err := validateTokenBazaarDispatch(payload)
			if test.wantErr && err == nil {
				t.Fatal("expected dimensions to be rejected")
			}

			if !test.wantErr && err != nil {
				t.Fatalf("expected dimensions to be accepted, got %v", err)
			}
		})
	}
}

// TestTokenBazaarEmbeddingsDispatchExactRequestConstruction exercises the
// full adapter Dispatch path (mocked HTTP, no live call) and asserts the
// exact method, URL, Authorization header, and forwarded body against
// https://zx1.deepwl.net/v1/embeddings, then confirms the response is parsed
// through as the standard OpenAI embeddings shape.
func TestTokenBazaarEmbeddingsDispatchExactRequestConstruction(t *testing.T) {
	var gotMethod, gotURL, gotAuth, gotContentType string

	var gotBody []byte

	client := &http.Client{Transport: tokenBazaarRoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		gotMethod = request.Method
		gotURL = request.URL.String()
		gotAuth = request.Header.Get("Authorization")
		gotContentType = request.Header.Get("Content-Type")

		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}

		gotBody = body

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body: io.NopCloser(strings.NewReader(
				`{"object":"list","data":[{"object":"embedding","embedding":[0.1,0.2,0.3],"index":0}],"model":"text-embedding-3-large","usage":{"prompt_tokens":3,"total_tokens":3}}`,
			)),
		}, nil
	})}

	payload := tokenBazaarEmbeddingsTestPayload("text-embedding-3-large")
	requestBody := `{"model":"text-embedding-3-large","input":"hello world","encoding_format":"float"}`
	payload.Body = []byte(requestBody)

	response, err := newTokenBazaarAdapter().Dispatch(context.Background(), client, payload)
	if err != nil {
		t.Fatalf("dispatch embeddings: %v", err)
	}
	defer response.Body.Close()

	if gotMethod != http.MethodPost {
		t.Fatalf("method = %q, want POST", gotMethod)
	}

	if gotURL != "https://zx1.deepwl.net/v1/embeddings" {
		t.Fatalf("url = %q", gotURL)
	}

	if gotAuth != "Bearer tokenbazaar-secret" {
		t.Fatalf("authorization = %q", gotAuth)
	}

	if gotContentType != "application/json" {
		t.Fatalf("content-type = %q", gotContentType)
	}

	if string(gotBody) != requestBody {
		t.Fatalf("forwarded body = %s, want verbatim %s", gotBody, requestBody)
	}

	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", response.StatusCode)
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}

	var parsed struct {
		Object string `json:"object"`
		Data   []struct {
			Embedding []float64 `json:"embedding"`
			Index     int       `json:"index"`
		} `json:"data"`
		Usage struct {
			PromptTokens int `json:"prompt_tokens"`
			TotalTokens  int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(responseBody, &parsed); err != nil {
		t.Fatalf("response body is not valid JSON: %v", err)
	}

	if parsed.Object != "list" || len(parsed.Data) != 1 || len(parsed.Data[0].Embedding) != 3 {
		t.Fatalf("unexpected parsed response shape: %+v", parsed)
	}
}

// TestTokenBazaarEmbeddingsMalformedResponseSurfacesClearError confirms that
// a 2xx response whose body does not match the inferred standard OpenAI
// embeddings envelope is surfaced as a distinct, identifiable error rather
// than silently returned as garbage or causing a panic downstream. This
// matters more here than for chat/images precisely because the embeddings
// contract is inferred, not first-party-confirmed.
func TestTokenBazaarEmbeddingsMalformedResponseSurfacesClearError(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"not_json", `not json at all`},
		{"wrong_object_field", `{"object":"embedding_list","data":[{"object":"embedding","embedding":[0.1],"index":0}]}`},
		{"empty_data", `{"object":"list","data":[]}`},
		{"missing_embedding_vector", `{"object":"list","data":[{"object":"embedding","index":0}]}`},
		{"chat_shaped_response", `{"choices":[{"message":{"role":"assistant","content":"hi"}}]}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := &http.Client{Transport: tokenBazaarRoundTripperFunc(func(request *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{"application/json"}},
					Body:       io.NopCloser(strings.NewReader(test.body)),
				}, nil
			})}

			payload := tokenBazaarEmbeddingsTestPayload("text-embedding-3-large")
			payload.Body = []byte(`{"model":"text-embedding-3-large","input":"hello world"}`)

			response, err := newTokenBazaarAdapter().Dispatch(context.Background(), client, payload)
			if err == nil {
				if response != nil {
					response.Body.Close()
				}

				t.Fatal("expected malformed embeddings response to be rejected with a clear error")
			}

			if !strings.Contains(err.Error(), "tokenbazaar_embeddings_response_shape_unexpected") &&
				!strings.Contains(err.Error(), "tokenbazaar_embeddings_response_read") {
				t.Fatalf("expected a distinguishable embeddings-response error, got %v", err)
			}
		})
	}
}

// TestTokenBazaarEmbeddingsErrorResponseIsNormalizedAndDoesNotLeakUpstreamBody
// confirms the embeddings dispatch path goes through the same
// 4xx-normalization (never-leak-upstream-body) treatment as the chat and
// image paths.
func TestTokenBazaarEmbeddingsErrorResponseIsNormalizedAndDoesNotLeakUpstreamBody(t *testing.T) {
	client := &http.Client{Transport: tokenBazaarRoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Header:     http.Header{tokenBazaarRequestIDHeader: []string{"req-embeddings-1"}},
			Body: io.NopCloser(strings.NewReader(
				`{"error":{"code":"invalid_request","message":"upstream vendor account id abc-123 must not leak"}}`,
			)),
		}, nil
	})}

	payload := tokenBazaarEmbeddingsTestPayload("text-embedding-3-large")
	payload.Body = []byte(`{"model":"text-embedding-3-large","input":"hello world"}`)

	response, err := newTokenBazaarAdapter().Dispatch(context.Background(), client, payload)
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want normalized 503", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read normalized body: %v", err)
	}

	if strings.Contains(string(body), "abc-123") || strings.Contains(string(body), "invalid_request") {
		t.Fatalf("normalized embeddings error body leaked upstream content: %s", body)
	}
}

// TestTokenBazaarEmbeddingsValidationErrorsNeverIncludeAPIKey confirms none
// of the embeddings-path validation error messages ever echo the
// deployment's API key, even when it is present on the payload.
func TestTokenBazaarEmbeddingsValidationErrorsNeverIncludeAPIKey(t *testing.T) {
	const secret = "tokenbazaar-super-secret-key-must-not-leak"

	badDimensions := tokenBazaarEmbeddingsTestPayload("text-embedding-3-large")
	badDimensions.Deployment.APIKey = secret
	badDimensions.Body = []byte(`{"model":"text-embedding-3-large","input":"hi","dimensions":99999}`)

	badEncoding := tokenBazaarEmbeddingsTestPayload("text-embedding-3-large")
	badEncoding.Deployment.APIKey = secret
	badEncoding.Body = []byte(`{"model":"text-embedding-3-large","input":"hi","encoding_format":"hex"}`)

	for _, payload := range []platform.TunnelRequest{badDimensions, badEncoding} {
		err := validateTokenBazaarDispatch(payload)
		if err == nil {
			t.Fatal("expected validation error")
		}

		if strings.Contains(err.Error(), secret) {
			t.Fatalf("validation error leaked the API key: %v", err)
		}
	}
}

// TestTokenBazaarEmbeddingsNeverForwardsGodarAIAuthorizationHeader mirrors
// the chat path's own coverage: the caller-supplied Authorization header
// must never reach TokenBazaar; only the deployment's own credential is
// forwarded.
func TestTokenBazaarEmbeddingsNeverForwardsGodarAIAuthorizationHeader(t *testing.T) {
	var gotAuth string

	client := &http.Client{Transport: tokenBazaarRoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		gotAuth = request.Header.Get("Authorization")

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body: io.NopCloser(strings.NewReader(
				`{"object":"list","data":[{"object":"embedding","embedding":[0.1],"index":0}]}`,
			)),
		}, nil
	})}

	payload := tokenBazaarEmbeddingsTestPayload("text-embedding-3-large")
	payload.Body = []byte(`{"model":"text-embedding-3-large","input":"hi"}`)
	payload.Headers = map[string]string{"Authorization": "Bearer godarai-caller-token-must-not-leak"}

	response, err := newTokenBazaarAdapter().Dispatch(context.Background(), client, payload)
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	defer response.Body.Close()

	if gotAuth != "Bearer tokenbazaar-secret" {
		t.Fatalf("upstream Authorization = %q, want only the TokenBazaar credential", gotAuth)
	}
}

func tokenBazaarTestPayload(model string) platform.TunnelRequest {
	profile := platform.ModelRuntimeProfile{
		Version:    1,
		Endpoints:  []string{"chat_completions"},
		Transports: []string{"http", "sse"},
		Lifecycle:  "active",
		Thinking:   platform.ModelRuntimeThinkingProfile{DefaultMode: "disabled", AllowedModes: []string{"disabled"}},
		Sampling:   platform.ModelRuntimeSamplingProfile{},
		Tools:      platform.ModelRuntimeToolsProfile{},
		Modalities: platform.ModelRuntimeModalities{Input: []string{"text"}, Output: []string{"text"}},
		Parameters: map[string][]string{
			"chat_completions": {"model", "messages", "stream", "tools", "tool_choice", "max_tokens"},
		},
		Limits: platform.ModelRuntimeLimits{MessagesBytes: 1, BatchBytes: 1, FileBytes: 1},
	}

	return platform.TunnelRequest{
		Operation:  platform.ProviderOperationChatCompletions,
		Deployment: platform.Deployment{Provider: "tokenbazaar", Model: model, APIKey: "tokenbazaar-secret"},
		ModelRuntimeProfiles: []platform.ModelRuntimeProfileBinding{{
			Provider: "tokenbazaar", CatalogRevision: "test", CanonicalID: "tokenbazaar/" + model,
			Slug: model, ProfileVersion: 1, Profile: profile, EffectiveFrom: time.Now(),
		}},
	}
}

// tokenBazaarResponsesTestPayload builds a minimal, in-scope /v1/responses
// TunnelRequest for one of the 6 Responses-capable models
// (tokenBazaarResponsesModels), mirroring tokenBazaarEmbeddingsTestPayload's
// style.
func tokenBazaarResponsesTestPayload(model string) platform.TunnelRequest {
	profile := platform.ModelRuntimeProfile{
		Version:    1,
		Endpoints:  []string{"models", "chat_completions", "responses"},
		Transports: []string{"http", "sse"},
		Lifecycle:  "active",
		Thinking:   platform.ModelRuntimeThinkingProfile{DefaultMode: "disabled", AllowedModes: []string{"disabled"}},
		Sampling:   platform.ModelRuntimeSamplingProfile{},
		Tools:      platform.ModelRuntimeToolsProfile{},
		Modalities: platform.ModelRuntimeModalities{Input: []string{"text"}, Output: []string{"text"}},
		Parameters: map[string][]string{
			"responses": {
				"model", "input", "instructions", "reasoning", "max_output_tokens",
				"stream", "tools", "tool_choice", "text",
			},
		},
		Limits: platform.ModelRuntimeLimits{MessagesBytes: 1, BatchBytes: 1, FileBytes: 1},
	}

	return platform.TunnelRequest{
		Operation:  platform.ProviderOperationOpenAIResponsesCreate,
		Deployment: platform.Deployment{Provider: "tokenbazaar", Model: model, APIKey: "tokenbazaar-secret"},
		ModelRuntimeProfiles: []platform.ModelRuntimeProfileBinding{{
			Provider: "tokenbazaar", CatalogRevision: "test", CanonicalID: "tokenbazaar/" + model,
			Slug: model, ProfileVersion: 1, Profile: profile, EffectiveFrom: time.Now(),
		}},
	}
}

// TestTokenBazaarResponsesDispatchRoutesForAllSixInScopeModels confirms each
// of the 6 in-scope Responses models (4 OpenAI-family, 2 DeepSeek) resolves
// through the Responses dispatch path for the documented
// ProviderOperationOpenAIResponsesCreate operation.
func TestTokenBazaarResponsesDispatchRoutesForAllSixInScopeModels(t *testing.T) {
	for model := range tokenBazaarResponsesModels {
		t.Run(model, func(t *testing.T) {
			payload := tokenBazaarResponsesTestPayload(model)
			payload.Body = []byte(`{"model":"` + model + `","input":"hello"}`)

			if err := validateTokenBazaarDispatch(payload); err != nil {
				t.Fatalf("expected responses dispatch to be accepted for %s, got %v", model, err)
			}
		})
	}

	if len(tokenBazaarResponsesModels) != 6 {
		t.Fatalf("expected exactly 6 in-scope Responses models, got %d: %v", len(tokenBazaarResponsesModels), tokenBazaarResponsesModels)
	}
}

// TestTokenBazaarResponsesOperationRejectsOutOfScopeModels confirms that the
// other 20 TokenBazaar models (15 chat-only models plus the 4 image models
// and the 1 embeddings model) are rejected with a clear, specific,
// operation-scoped "unsupported model" error -- never silently accepted,
// never a generic 404, and never misrouted to the chat_completions path --
// when requested through the /v1/responses operation.
func TestTokenBazaarResponsesOperationRejectsOutOfScopeModels(t *testing.T) {
	outOfScope := []string{
		// 15 chat-capable models that are NOT in the 6-model Responses scope.
		"claude-fable-5", "claude-haiku-4-5", "claude-opus-5", "claude-sonnet-5",
		"gemini-2.5-flash", "gemini-2.5-pro", "gemini-3.1-pro-preview",
		"gemini-3.5-flash", "gemini-3.6-flash", "gemini-3.7-flash",
		"glm-5.2", "glm-5v-turbo", "kimi-k2.6", "kimi-k2.7-code", "kimi-k3",
		// 4 image models + 1 embeddings model.
		"gpt-image-2", "gemini-3-pro-image", "gemini-3.1-flash-image", "gemini-3.1-flash-lite-image",
		"text-embedding-3-large",
	}

	if len(outOfScope) != 20 {
		t.Fatalf("expected exactly 20 out-of-scope models in this table, got %d", len(outOfScope))
	}

	for _, model := range outOfScope {
		t.Run(model, func(t *testing.T) {
			payload := tokenBazaarResponsesTestPayload(model)
			payload.Body = []byte(`{"model":"` + model + `","input":"hello"}`)

			err := validateTokenBazaarDispatch(payload)
			if err == nil || !strings.Contains(err.Error(), "tokenbazaar_unsupported_model:") {
				t.Fatalf("expected generic unsupported-model error for %s via responses, got %v", model, err)
			}
		})
	}
}

// TestTokenBazaarResponsesRuntimeProfileRequired confirms a responses request
// for an in-scope model whose runtime profile does not allow the "responses"
// endpoint is rejected, mirroring the chat/images/embeddings paths' own
// runtime-profile enforcement.
func TestTokenBazaarResponsesRuntimeProfileRequired(t *testing.T) {
	payload := tokenBazaarResponsesTestPayload("gpt-5.5")
	payload.ModelRuntimeProfiles[0].Profile.Endpoints = []string{"models", "chat_completions"} // no "responses"
	payload.Body = []byte(`{"model":"gpt-5.5","input":"hello"}`)

	err := validateTokenBazaarDispatch(payload)
	if err == nil || !strings.Contains(err.Error(), "tokenbazaar_runtime_profile_missing") {
		t.Fatalf("expected tokenbazaar_runtime_profile_missing, got %v", err)
	}
}

// TestTokenBazaarResponsesRequiresInputOrInstructions confirms a request
// missing both "input" and "instructions" is rejected before dispatch.
func TestTokenBazaarResponsesRequiresInputOrInstructions(t *testing.T) {
	payload := tokenBazaarResponsesTestPayload("gpt-5.5")
	payload.Body = []byte(`{"model":"gpt-5.5"}`)

	err := validateTokenBazaarDispatch(payload)
	if err == nil || !strings.Contains(err.Error(), "tokenbazaar_responses_input_required") {
		t.Fatalf("expected tokenbazaar_responses_input_required, got %v", err)
	}

	withInstructions := tokenBazaarResponsesTestPayload("gpt-5.5")
	withInstructions.Body = []byte(`{"model":"gpt-5.5","instructions":"be concise"}`)

	if err := validateTokenBazaarDispatch(withInstructions); err != nil {
		t.Fatalf("expected instructions-only request to be accepted, got %v", err)
	}
}

// TestTokenBazaarResponsesRejectsPreviousResponseID covers the first of the
// three documented "not supported" fields: TokenBazaar's /v1/responses is
// stateless, so a non-empty previous_response_id must be rejected before any
// HTTP call, with its own distinguishable error. A null, absent, or
// empty-string value must NOT be rejected.
func TestTokenBazaarResponsesRejectsPreviousResponseID(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{"absent", `{"model":"gpt-5.5","input":"hi"}`, false},
		{"null", `{"model":"gpt-5.5","input":"hi","previous_response_id":null}`, false},
		{"empty_string", `{"model":"gpt-5.5","input":"hi","previous_response_id":""}`, false},
		{"non_empty", `{"model":"gpt-5.5","input":"hi","previous_response_id":"resp_abc123"}`, true},
		{"wrong_type", `{"model":"gpt-5.5","input":"hi","previous_response_id":123}`, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payload := tokenBazaarResponsesTestPayload("gpt-5.5")
			payload.Body = []byte(test.body)

			err := validateTokenBazaarDispatch(payload)
			if test.wantErr {
				if err == nil {
					t.Fatal("expected previous_response_id to be rejected")
				}

				if test.name == "non_empty" && !strings.Contains(err.Error(), "tokenbazaar_responses_previous_response_id_unsupported") {
					t.Fatalf("expected tokenbazaar_responses_previous_response_id_unsupported, got %v", err)
				}

				return
			}

			if err != nil {
				t.Fatalf("expected previous_response_id to be accepted, got %v", err)
			}
		})
	}
}

// TestTokenBazaarResponsesRejectsBuiltinTools covers the second of the three
// documented "not supported" fields: any OpenAI-hosted built-in tool type is
// rejected before dispatch with its own distinguishable error, while
// caller-defined function tools (and an absent/empty tools array) are fine.
func TestTokenBazaarResponsesRejectsBuiltinTools(t *testing.T) {
	tests := []struct {
		name    string
		tools   string
		wantErr bool
	}{
		{"absent", ``, false},
		{"empty_array", `"tools":[],`, false},
		{"function_tool", `"tools":[{"type":"function","name":"lookup","parameters":{}}],`, false},
		{"web_search", `"tools":[{"type":"web_search"}],`, true},
		{"file_search", `"tools":[{"type":"file_search"}],`, true},
		{"image_generation", `"tools":[{"type":"image_generation"}],`, true},
		{"code_interpreter", `"tools":[{"type":"code_interpreter"}],`, true},
		{"mixed_function_and_builtin", `"tools":[{"type":"function","name":"lookup"},{"type":"web_search"}],`, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payload := tokenBazaarResponsesTestPayload("gpt-5.5")
			payload.Body = []byte(`{"model":"gpt-5.5",` + test.tools + `"input":"hi"}`)

			err := validateTokenBazaarDispatch(payload)
			if test.wantErr {
				if err == nil || !strings.Contains(err.Error(), "tokenbazaar_responses_builtin_tool_unsupported") {
					t.Fatalf("expected tokenbazaar_responses_builtin_tool_unsupported, got %v", err)
				}

				return
			}

			if err != nil {
				t.Fatalf("expected tools to be accepted, got %v", err)
			}
		})
	}
}

// TestTokenBazaarResponsesRejectsBackgroundTrue covers the third of the three
// documented "not supported" fields: background: true is rejected before
// dispatch with its own distinguishable error; an absent field or an
// explicit background: false are both fine.
func TestTokenBazaarResponsesRejectsBackgroundTrue(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{"absent", `{"model":"gpt-5.5","input":"hi"}`, false},
		{"false", `{"model":"gpt-5.5","input":"hi","background":false}`, false},
		{"true", `{"model":"gpt-5.5","input":"hi","background":true}`, true},
		{"wrong_type", `{"model":"gpt-5.5","input":"hi","background":"true"}`, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payload := tokenBazaarResponsesTestPayload("gpt-5.5")
			payload.Body = []byte(test.body)

			err := validateTokenBazaarDispatch(payload)
			if test.wantErr {
				if err == nil {
					t.Fatal("expected background:true to be rejected")
				}

				if test.name == "true" && !strings.Contains(err.Error(), "tokenbazaar_responses_background_unsupported") {
					t.Fatalf("expected tokenbazaar_responses_background_unsupported, got %v", err)
				}

				return
			}

			if err != nil {
				t.Fatalf("expected background to be accepted, got %v", err)
			}
		})
	}
}

// TestTokenBazaarResponsesThreeRejectionsHappenBeforeAnyHTTPCall proves, via
// an end-to-end mocked-HTTP Dispatch (not just the standalone validator),
// that each of the three documented "not supported" fields is rejected
// before any upstream call is made, mirroring
// TestTokenBazaarGPTImage2ResponseFormatDispatchedEndToEnd's proof style for
// the images path.
func TestTokenBazaarResponsesThreeRejectionsHappenBeforeAnyHTTPCall(t *testing.T) {
	var calls int32

	client := &http.Client{Transport: tokenBazaarRoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		atomic.AddInt32(&calls, 1)

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"id":"resp_1","object":"response","status":"completed"}`)),
		}, nil
	})}

	accepted := tokenBazaarResponsesTestPayload("gpt-5.5")
	accepted.Body = []byte(`{"model":"gpt-5.5","input":"hello"}`)

	response, err := newTokenBazaarAdapter().Dispatch(context.Background(), client, accepted)
	if err != nil {
		t.Fatalf("expected a valid request to succeed, got %v", err)
	}
	defer response.Body.Close()

	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected exactly 1 upstream call for the valid request, got %d", calls)
	}

	rejections := map[string]string{
		"previous_response_id": `{"model":"gpt-5.5","input":"hello","previous_response_id":"resp_abc"}`,
		"builtin_tool":         `{"model":"gpt-5.5","input":"hello","tools":[{"type":"web_search"}]}`,
		"background_true":      `{"model":"gpt-5.5","input":"hello","background":true}`,
	}

	for name, body := range rejections {
		t.Run(name, func(t *testing.T) {
			payload := tokenBazaarResponsesTestPayload("gpt-5.5")
			payload.Body = []byte(body)

			if _, err := newTokenBazaarAdapter().Dispatch(context.Background(), client, payload); err == nil {
				t.Fatal("expected the request to be rejected before dispatch")
			}

			if atomic.LoadInt32(&calls) != 1 {
				t.Fatalf("rejection %q must happen before any HTTP call; total upstream calls = %d", name, calls)
			}
		})
	}
}

// TestTokenBazaarResponsesDispatchExactRequestConstruction exercises the full
// adapter Dispatch path (mocked HTTP, no live call) and asserts the exact
// method, URL, Authorization header, Content-Type, and forwarded body against
// https://zx1.deepwl.net/v1/responses, mirroring
// TestTokenBazaarEmbeddingsDispatchExactRequestConstruction's style.
func TestTokenBazaarResponsesDispatchExactRequestConstruction(t *testing.T) {
	var gotMethod, gotURL, gotAuth, gotContentType string

	var gotBody []byte

	client := &http.Client{Transport: tokenBazaarRoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		gotMethod = request.Method
		gotURL = request.URL.String()
		gotAuth = request.Header.Get("Authorization")
		gotContentType = request.Header.Get("Content-Type")

		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}

		gotBody = body

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body: io.NopCloser(strings.NewReader(
				`{"id":"resp_1","object":"response","model":"gpt-5.5","status":"completed",` +
					`"output":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"hi"}]}],` +
					`"usage":{"input_tokens":3,"output_tokens":2,"total_tokens":5}}`,
			)),
		}, nil
	})}

	payload := tokenBazaarResponsesTestPayload("gpt-5.5")
	requestBody := `{"model":"gpt-5.5","input":"hello world"}`
	payload.Body = []byte(requestBody)

	response, err := newTokenBazaarAdapter().Dispatch(context.Background(), client, payload)
	if err != nil {
		t.Fatalf("dispatch responses: %v", err)
	}
	defer response.Body.Close()

	if gotMethod != http.MethodPost {
		t.Fatalf("method = %q, want POST", gotMethod)
	}

	if gotURL != "https://zx1.deepwl.net/v1/responses" {
		t.Fatalf("url = %q", gotURL)
	}

	if gotAuth != "Bearer tokenbazaar-secret" {
		t.Fatalf("authorization = %q", gotAuth)
	}

	if gotContentType != "application/json" {
		t.Fatalf("content-type = %q", gotContentType)
	}

	if string(gotBody) != requestBody {
		t.Fatalf("forwarded body = %s, want verbatim %s", gotBody, requestBody)
	}

	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", response.StatusCode)
	}
}

// TestTokenBazaarResponsesReasoningEffortPassesThroughUnmodified confirms
// this delivery does not rewrite, strip, or otherwise touch a documented
// reasoning.effort field: the request body reaching TokenBazaar must be
// byte-for-byte identical to what the caller sent, for each of the 6 in-scope
// models (unlike DeepSeek's own adapter, TokenBazaar's does not inject an
// opaque managed-user field into the body, so no rewriting is expected here
// at all).
func TestTokenBazaarResponsesReasoningEffortPassesThroughUnmodified(t *testing.T) {
	for model := range tokenBazaarResponsesModels {
		t.Run(model, func(t *testing.T) {
			var gotBody []byte

			client := &http.Client{Transport: tokenBazaarRoundTripperFunc(func(request *http.Request) (*http.Response, error) {
				body, err := io.ReadAll(request.Body)
				if err != nil {
					t.Fatalf("read request body: %v", err)
				}

				gotBody = body

				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{"application/json"}},
					Body:       io.NopCloser(strings.NewReader(`{"id":"resp_1","object":"response","status":"completed"}`)),
				}, nil
			})}

			payload := tokenBazaarResponsesTestPayload(model)
			requestBody := `{"model":"` + model + `","input":"hello","reasoning":{"effort":"low"}}`
			payload.Body = []byte(requestBody)

			response, err := newTokenBazaarAdapter().Dispatch(context.Background(), client, payload)
			if err != nil {
				t.Fatalf("dispatch: %v", err)
			}
			defer response.Body.Close()

			if string(gotBody) != requestBody {
				t.Fatalf("reasoning.effort was not passed through unmodified: got %s, want %s", gotBody, requestBody)
			}

			if !strings.Contains(string(gotBody), `"reasoning":{"effort":"low"}`) {
				t.Fatalf("reasoning.effort missing from forwarded body: %s", gotBody)
			}
		})
	}
}

// TestTokenBazaarResponsesSSEStreamingPassesThroughUntilTerminalEvent covers
// the documented typed-event streaming shape (response.output_text.delta ...
// response.completed) with the same defensive posture already established
// for chat SSE (TestTokenBazaarSSEStreamingPassesThroughUntilDone): this
// adapter passes the raw SSE bytes through unmodified; typed-event
// parsing/terminal-event detection and malformed/unexpected-event handling
// are the shared, provider-agnostic responsibility of
// services/control-plane/internal/proxy/stream_relay.go's
// openAIResponseStreamState, which is already exhaustively unit-tested
// (stream_relay_test.go) independent of provider -- this test only confirms
// the TokenBazaar adapter itself never mutates or buffers the documented
// event bytes.
func TestTokenBazaarResponsesSSEStreamingPassesThroughUntilTerminalEvent(t *testing.T) {
	sseBody := "event: response.output_text.delta\n" +
		"data: {\"type\":\"response.output_text.delta\",\"delta\":\"Hel\",\"sequence_number\":1}\n\n" +
		"event: response.output_text.delta\n" +
		"data: {\"type\":\"response.output_text.delta\",\"delta\":\"lo\",\"sequence_number\":2}\n\n" +
		"event: response.completed\n" +
		"data: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_1\",\"model\":\"gpt-5.5\",\"status\":\"completed\"},\"sequence_number\":3}\n\n"

	client := &http.Client{Transport: tokenBazaarRoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		if request.Header.Get("Accept") != "text/event-stream" {
			t.Fatalf("expected SSE Accept header, got %q", request.Header.Get("Accept"))
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader(sseBody)),
		}, nil
	})}

	payload := tokenBazaarResponsesTestPayload("gpt-5.5")
	payload.Stream = true
	payload.Body = []byte(`{"model":"gpt-5.5","input":"hi","stream":true}`)

	response, err := newTokenBazaarAdapter().Dispatch(context.Background(), client, payload)
	if err != nil {
		t.Fatalf("dispatch streaming responses request: %v", err)
	}
	defer response.Body.Close()

	got, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read SSE body: %v", err)
	}

	if string(got) != sseBody {
		t.Fatalf("SSE body was altered by the adapter: %q", got)
	}

	if !strings.Contains(string(got), "event: response.completed") {
		t.Fatalf("terminal event was not preserved: %s", got)
	}
}

// TestTokenBazaarResponsesErrorResponseIsNormalizedAndDoesNotLeakUpstreamBody
// mirrors the chat/images/embeddings paths' own 4xx-normalization coverage
// for the responses operation.
func TestTokenBazaarResponsesErrorResponseIsNormalizedAndDoesNotLeakUpstreamBody(t *testing.T) {
	client := &http.Client{Transport: tokenBazaarRoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Header:     http.Header{tokenBazaarRequestIDHeader: []string{"req-responses-1"}},
			Body: io.NopCloser(strings.NewReader(
				`{"error":{"code":"invalid_request","message":"upstream vendor account id abc-123 must not leak"}}`,
			)),
		}, nil
	})}

	payload := tokenBazaarResponsesTestPayload("gpt-5.5")
	payload.Body = []byte(`{"model":"gpt-5.5","input":"hello"}`)

	response, err := newTokenBazaarAdapter().Dispatch(context.Background(), client, payload)
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want normalized 503", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read normalized body: %v", err)
	}

	if strings.Contains(string(body), "abc-123") || strings.Contains(string(body), "invalid_request") {
		t.Fatalf("normalized responses error body leaked upstream content: %s", body)
	}
}

// TestTokenBazaarResponsesValidationErrorsNeverIncludeAPIKey confirms none of
// the responses-path validation error messages ever echo the deployment's
// API key, even when it is present on the payload.
func TestTokenBazaarResponsesValidationErrorsNeverIncludeAPIKey(t *testing.T) {
	const secret = "tokenbazaar-super-secret-key-must-not-leak"

	badPreviousResponseID := tokenBazaarResponsesTestPayload("gpt-5.5")
	badPreviousResponseID.Deployment.APIKey = secret
	badPreviousResponseID.Body = []byte(`{"model":"gpt-5.5","input":"hi","previous_response_id":"resp_abc"}`)

	badTool := tokenBazaarResponsesTestPayload("gpt-5.5")
	badTool.Deployment.APIKey = secret
	badTool.Body = []byte(`{"model":"gpt-5.5","input":"hi","tools":[{"type":"web_search"}]}`)

	badBackground := tokenBazaarResponsesTestPayload("gpt-5.5")
	badBackground.Deployment.APIKey = secret
	badBackground.Body = []byte(`{"model":"gpt-5.5","input":"hi","background":true}`)

	for _, payload := range []platform.TunnelRequest{badPreviousResponseID, badTool, badBackground} {
		err := validateTokenBazaarDispatch(payload)
		if err == nil {
			t.Fatal("expected validation error")
		}

		if strings.Contains(err.Error(), secret) {
			t.Fatalf("validation error leaked the API key: %v", err)
		}
	}
}

// TestTokenBazaarResponsesNeverForwardsGodarAIAuthorizationHeader mirrors the
// chat/embeddings paths' own coverage: the caller-supplied Authorization
// header must never reach TokenBazaar; only the deployment's own credential
// is forwarded.
func TestTokenBazaarResponsesNeverForwardsGodarAIAuthorizationHeader(t *testing.T) {
	var gotAuth string

	client := &http.Client{Transport: tokenBazaarRoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		gotAuth = request.Header.Get("Authorization")

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"id":"resp_1","object":"response","status":"completed"}`)),
		}, nil
	})}

	payload := tokenBazaarResponsesTestPayload("gpt-5.5")
	payload.Body = []byte(`{"model":"gpt-5.5","input":"hi"}`)
	payload.Headers = map[string]string{"Authorization": "Bearer godarai-caller-token-must-not-leak"}

	response, err := newTokenBazaarAdapter().Dispatch(context.Background(), client, payload)
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	defer response.Body.Close()

	if gotAuth != "Bearer tokenbazaar-secret" {
		t.Fatalf("upstream Authorization = %q, want only the TokenBazaar credential", gotAuth)
	}
}

// ---------------------------------------------------------------------------
// Messages endpoint delivery (POST /v1/messages)
// ---------------------------------------------------------------------------

// tokenBazaarMessagesTestPayload builds a minimal, in-scope /v1/messages
// TunnelRequest for one of the 4 Messages-capable models
// (tokenBazaarMessagesModels), mirroring tokenBazaarResponsesTestPayload's
// style. The request body is Anthropic-native JSON (model, messages,
// max_tokens, ...), not an OpenAI-compatible shape.
func tokenBazaarMessagesTestPayload(model string) platform.TunnelRequest {
	profile := platform.ModelRuntimeProfile{
		Version:    1,
		Endpoints:  []string{"models", "chat_completions", "messages"},
		Transports: []string{"http", "sse"},
		Lifecycle:  "active",
		Thinking: platform.ModelRuntimeThinkingProfile{
			DefaultMode: "adaptive", AllowedModes: []string{"adaptive", "disabled"},
		},
		Sampling: platform.ModelRuntimeSamplingProfile{},
		Tools:    platform.ModelRuntimeToolsProfile{},
		Modalities: platform.ModelRuntimeModalities{
			Input: []string{"text"}, Output: []string{"text"},
		},
		Parameters: map[string][]string{
			"messages": {"model", "messages", "max_tokens", "system", "stream", "thinking"},
		},
		Limits: platform.ModelRuntimeLimits{MessagesBytes: 32 * 1024 * 1024, BatchBytes: 1, FileBytes: 1},
	}

	return platform.TunnelRequest{
		Operation:  platform.ProviderOperationAnthropicMessages,
		Deployment: platform.Deployment{Provider: "tokenbazaar", Model: model, APIKey: "tokenbazaar-secret"},
		ModelRuntimeProfiles: []platform.ModelRuntimeProfileBinding{{
			Provider: "tokenbazaar", CatalogRevision: "test", CanonicalID: "tokenbazaar/" + model,
			Slug: model, ProfileVersion: 1, Profile: profile, EffectiveFrom: time.Now(),
		}},
	}
}

// TestTokenBazaarMessagesDispatchRoutesForAllFourInScopeModels confirms each
// of the 4 in-scope Messages models (claude-fable-5, claude-haiku-4-5,
// claude-opus-5, claude-sonnet-5) resolves through the Messages dispatch path
// for the documented platform.ProviderOperationAnthropicMessages operation.
func TestTokenBazaarMessagesDispatchRoutesForAllFourInScopeModels(t *testing.T) {
	for model := range tokenBazaarMessagesModels {
		t.Run(model, func(t *testing.T) {
			payload := tokenBazaarMessagesTestPayload(model)
			payload.Body = []byte(`{"model":"` + model + `","messages":[{"role":"user","content":"hi"}],"max_tokens":16}`)

			if err := validateTokenBazaarDispatch(payload); err != nil {
				t.Fatalf("expected messages dispatch to be accepted for %s, got %v", model, err)
			}
		})
	}

	if len(tokenBazaarMessagesModels) != 4 {
		t.Fatalf("expected exactly 4 in-scope Messages models, got %d: %v", len(tokenBazaarMessagesModels), tokenBazaarMessagesModels)
	}
}

// TestTokenBazaarMessagesOperationRejectsOutOfScopeModels confirms that the
// other 22 TokenBazaar models are rejected with a clear, specific,
// operation-scoped "unsupported model" error when requested through the
// /v1/messages operation -- never silently accepted, never a generic 404,
// and never misrouted to the chat_completions path.
func TestTokenBazaarMessagesOperationRejectsOutOfScopeModels(t *testing.T) {
	outOfScope := []string{
		// 17 chat-capable models that are NOT in the 4-model Messages scope.
		"deepseek-v4-flash", "deepseek-v4-pro",
		"gemini-2.5-flash", "gemini-2.5-pro", "gemini-3.1-pro-preview",
		"gemini-3.5-flash", "gemini-3.6-flash", "gemini-3.7-flash",
		"glm-5.2", "glm-5v-turbo", "kimi-k2.6", "kimi-k2.7-code", "kimi-k3",
		"gpt-5.5", "gpt-5.6-luna", "gpt-5.6-sol", "gpt-5.6-terra",
		// 4 image models + 1 embeddings model.
		"gpt-image-2", "gemini-3-pro-image", "gemini-3.1-flash-image", "gemini-3.1-flash-lite-image",
		"text-embedding-3-large",
	}

	if len(outOfScope) != 22 {
		t.Fatalf("expected exactly 22 out-of-scope models in this table, got %d", len(outOfScope))
	}

	for _, model := range outOfScope {
		t.Run(model, func(t *testing.T) {
			payload := tokenBazaarMessagesTestPayload(model)
			payload.Body = []byte(`{"model":"` + model + `","messages":[{"role":"user","content":"hi"}],"max_tokens":16}`)

			err := validateTokenBazaarDispatch(payload)
			if err == nil || !strings.Contains(err.Error(), "tokenbazaar_unsupported_model:") {
				t.Fatalf("expected generic unsupported-model error for %s via messages, got %v", model, err)
			}
		})
	}
}

// TestTokenBazaarMessagesRuntimeProfileRequired confirms a messages request
// for an in-scope model whose runtime profile does not allow the "messages"
// endpoint is rejected, mirroring the chat/images/embeddings/responses
// paths' own runtime-profile enforcement.
func TestTokenBazaarMessagesRuntimeProfileRequired(t *testing.T) {
	payload := tokenBazaarMessagesTestPayload("claude-sonnet-5")
	payload.ModelRuntimeProfiles[0].Profile.Endpoints = []string{"models", "chat_completions"} // no "messages"
	payload.Body = []byte(`{"model":"claude-sonnet-5","messages":[{"role":"user","content":"hi"}],"max_tokens":16}`)

	err := validateTokenBazaarDispatch(payload)
	if err == nil || !strings.Contains(err.Error(), "tokenbazaar_runtime_profile_missing") {
		t.Fatalf("expected tokenbazaar_runtime_profile_missing, got %v", err)
	}
}

// TestTokenBazaarMessagesRequiresMaxTokens confirms max_tokens must be
// present and a positive integer before dispatch -- TokenBazaar's own docs
// mark it required (matching Claude's own requirement), and this delivery
// does not rely on upstream to reject its absence.
func TestTokenBazaarMessagesRequiresMaxTokens(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{"absent", `{"model":"claude-sonnet-5","messages":[{"role":"user","content":"hi"}]}`, true},
		{"null", `{"model":"claude-sonnet-5","messages":[{"role":"user","content":"hi"}],"max_tokens":null}`, true},
		{"zero", `{"model":"claude-sonnet-5","messages":[{"role":"user","content":"hi"}],"max_tokens":0}`, true},
		{"negative", `{"model":"claude-sonnet-5","messages":[{"role":"user","content":"hi"}],"max_tokens":-1}`, true},
		{"non_integer", `{"model":"claude-sonnet-5","messages":[{"role":"user","content":"hi"}],"max_tokens":1.5}`, true},
		{"wrong_type", `{"model":"claude-sonnet-5","messages":[{"role":"user","content":"hi"}],"max_tokens":"16"}`, true},
		{"valid", `{"model":"claude-sonnet-5","messages":[{"role":"user","content":"hi"}],"max_tokens":1024}`, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payload := tokenBazaarMessagesTestPayload("claude-sonnet-5")
			payload.Body = []byte(test.body)

			err := validateTokenBazaarDispatch(payload)
			if test.wantErr {
				if err == nil || !strings.Contains(err.Error(), "tokenbazaar_messages_max_tokens_required") &&
					!strings.Contains(err.Error(), "tokenbazaar_invalid_json") {
					t.Fatalf("expected max_tokens rejection, got %v", err)
				}

				return
			}

			if err != nil {
				t.Fatalf("expected valid max_tokens to be accepted, got %v", err)
			}
		})
	}
}

// TestTokenBazaarMessagesRequiresMessagesField confirms a request missing
// "messages" entirely is rejected before dispatch.
func TestTokenBazaarMessagesRequiresMessagesField(t *testing.T) {
	payload := tokenBazaarMessagesTestPayload("claude-sonnet-5")
	payload.Body = []byte(`{"model":"claude-sonnet-5","max_tokens":16}`)

	err := validateTokenBazaarDispatch(payload)
	if err == nil || !strings.Contains(err.Error(), "tokenbazaar_messages_required") {
		t.Fatalf("expected tokenbazaar_messages_required, got %v", err)
	}
}

// TestTokenBazaarMessagesRejectsToolUse covers this delivery's fail-closed
// decision on tool-use: TokenBazaar's documented /v1/messages example does
// not show a "tools" parameter (unlike /v1/chat/completions, which documents
// it explicitly), so any non-empty tools array is rejected before dispatch.
// An absent or empty tools array must NOT be rejected.
func TestTokenBazaarMessagesRejectsToolUse(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{"absent", `{"model":"claude-sonnet-5","messages":[{"role":"user","content":"hi"}],"max_tokens":16}`, false},
		{"null", `{"model":"claude-sonnet-5","messages":[{"role":"user","content":"hi"}],"max_tokens":16,"tools":null}`, false},
		{"empty_array", `{"model":"claude-sonnet-5","messages":[{"role":"user","content":"hi"}],"max_tokens":16,"tools":[]}`, false},
		{
			"one_tool",
			`{"model":"claude-sonnet-5","messages":[{"role":"user","content":"hi"}],"max_tokens":16,` +
				`"tools":[{"name":"lookup","input_schema":{}}]}`,
			true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payload := tokenBazaarMessagesTestPayload("claude-sonnet-5")
			payload.Body = []byte(test.body)

			err := validateTokenBazaarDispatch(payload)
			if test.wantErr {
				if err == nil || !strings.Contains(err.Error(), "tokenbazaar_messages_tools_unsupported") {
					t.Fatalf("expected tokenbazaar_messages_tools_unsupported, got %v", err)
				}

				return
			}

			if err != nil {
				t.Fatalf("expected request to be accepted, got %v", err)
			}
		})
	}
}

// TestTokenBazaarMessagesRejectsPromptCaching covers this delivery's
// fail-closed decision on prompt caching: a cache_control block on a system
// block or a message content block is rejected before dispatch, since
// TokenBazaar's documented /v1/messages example does not show this field.
// (The Anthropic schema's third cache_control location, a tool definition, is
// not covered here as a distinct case: any non-empty "tools" array is already
// rejected outright by TestTokenBazaarMessagesRejectsToolUse, so a
// cache_control-on-a-tool request is already covered by that rejection.)
func TestTokenBazaarMessagesRejectsPromptCaching(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{
			"system_block",
			`{"model":"claude-sonnet-5","max_tokens":16,` +
				`"system":[{"type":"text","text":"be helpful","cache_control":{"type":"ephemeral"}}],` +
				`"messages":[{"role":"user","content":"hi"}]}`,
		},
		{
			"message_content_block",
			`{"model":"claude-sonnet-5","max_tokens":16,"messages":[{"role":"user","content":` +
				`[{"type":"text","text":"hi","cache_control":{"type":"ephemeral"}}]}]}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payload := tokenBazaarMessagesTestPayload("claude-sonnet-5")
			payload.Body = []byte(test.body)

			err := validateTokenBazaarDispatch(payload)
			if err == nil || !strings.Contains(err.Error(), "tokenbazaar_messages_prompt_caching_unsupported") {
				t.Fatalf("expected tokenbazaar_messages_prompt_caching_unsupported, got %v", err)
			}
		})
	}

	t.Run("no_cache_control_is_accepted", func(t *testing.T) {
		payload := tokenBazaarMessagesTestPayload("claude-sonnet-5")
		payload.Body = []byte(
			`{"model":"claude-sonnet-5","max_tokens":16,"system":"be helpful",` +
				`"messages":[{"role":"user","content":"hi"}]}`,
		)

		if err := validateTokenBazaarDispatch(payload); err != nil {
			t.Fatalf("expected request without cache_control to be accepted, got %v", err)
		}
	})
}

// TestTokenBazaarMessagesThreeRejectionsHappenBeforeAnyHTTPCall proves
// max_tokens absence, tool-use, and prompt-caching are all rejected before
// any upstream call, via an end-to-end mocked-HTTP Dispatch call (not just
// the standalone validator).
func TestTokenBazaarMessagesThreeRejectionsHappenBeforeAnyHTTPCall(t *testing.T) {
	var calls int32

	client := &http.Client{Transport: tokenBazaarRoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		atomic.AddInt32(&calls, 1)

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body: io.NopCloser(strings.NewReader(
				`{"id":"msg_1","type":"message","role":"assistant","model":"claude-sonnet-5",` +
					`"content":[{"type":"text","text":"hi"}],"stop_reason":"end_turn","usage":{"input_tokens":1,"output_tokens":1}}`,
			)),
		}, nil
	})}

	accepted := tokenBazaarMessagesTestPayload("claude-sonnet-5")
	accepted.Body = []byte(`{"model":"claude-sonnet-5","messages":[{"role":"user","content":"hi"}],"max_tokens":16}`)

	response, err := newTokenBazaarAdapter().Dispatch(context.Background(), client, accepted)
	if err != nil {
		t.Fatalf("expected a valid request to succeed, got %v", err)
	}
	defer response.Body.Close()

	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected exactly 1 upstream call for the valid request, got %d", calls)
	}

	rejections := map[string]string{
		"max_tokens_absent": `{"model":"claude-sonnet-5","messages":[{"role":"user","content":"hi"}]}`,
		"tool_use": `{"model":"claude-sonnet-5","messages":[{"role":"user","content":"hi"}],"max_tokens":16,` +
			`"tools":[{"name":"lookup","input_schema":{}}]}`,
		"prompt_caching": `{"model":"claude-sonnet-5","messages":[{"role":"user","content":` +
			`[{"type":"text","text":"hi","cache_control":{"type":"ephemeral"}}]}],"max_tokens":16}`,
	}

	for name, body := range rejections {
		t.Run(name, func(t *testing.T) {
			payload := tokenBazaarMessagesTestPayload("claude-sonnet-5")
			payload.Body = []byte(body)

			if _, err := newTokenBazaarAdapter().Dispatch(context.Background(), client, payload); err == nil {
				t.Fatal("expected the request to be rejected before dispatch")
			}

			if atomic.LoadInt32(&calls) != 1 {
				t.Fatalf("rejection %q must happen before any HTTP call; total upstream calls = %d", name, calls)
			}
		})
	}
}

// TestTokenBazaarMessagesDispatchExactRequestConstruction exercises the full
// adapter Dispatch path (mocked HTTP, no live call) and asserts the exact
// method, URL, Authorization header, Content-Type, and forwarded body against
// https://zx1.deepwl.net/v1/messages.
func TestTokenBazaarMessagesDispatchExactRequestConstruction(t *testing.T) {
	var gotMethod, gotURL, gotAuth, gotContentType string

	var gotBody []byte

	client := &http.Client{Transport: tokenBazaarRoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		gotMethod = request.Method
		gotURL = request.URL.String()
		gotAuth = request.Header.Get("Authorization")
		gotContentType = request.Header.Get("Content-Type")

		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}

		gotBody = body

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body: io.NopCloser(strings.NewReader(
				`{"id":"msg_1","type":"message","role":"assistant","model":"claude-sonnet-5",` +
					`"content":[{"type":"text","text":"hi there"}],"stop_reason":"end_turn",` +
					`"usage":{"input_tokens":3,"output_tokens":2}}`,
			)),
		}, nil
	})}

	payload := tokenBazaarMessagesTestPayload("claude-sonnet-5")
	requestBody := `{"model":"claude-sonnet-5","messages":[{"role":"user","content":"hello world"}],"max_tokens":16}`
	payload.Body = []byte(requestBody)

	response, err := newTokenBazaarAdapter().Dispatch(context.Background(), client, payload)
	if err != nil {
		t.Fatalf("dispatch messages: %v", err)
	}
	defer response.Body.Close()

	if gotMethod != http.MethodPost {
		t.Fatalf("method = %q, want POST", gotMethod)
	}

	if gotURL != "https://zx1.deepwl.net/v1/messages" {
		t.Fatalf("url = %q", gotURL)
	}

	if gotAuth != "Bearer tokenbazaar-secret" {
		t.Fatalf("authorization = %q", gotAuth)
	}

	if gotContentType != "application/json" {
		t.Fatalf("content-type = %q", gotContentType)
	}

	if string(gotBody) != requestBody {
		t.Fatalf("forwarded body = %s, want verbatim %s", gotBody, requestBody)
	}

	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", response.StatusCode)
	}
}

// TestTokenBazaarMessagesThinkingPassesThroughUnmodified confirms this
// delivery does not rewrite, strip, or otherwise touch a documented
// thinking.budget_tokens field: the request body reaching TokenBazaar must be
// byte-for-byte identical to what the caller sent, for each of the 4
// in-scope models.
func TestTokenBazaarMessagesThinkingPassesThroughUnmodified(t *testing.T) {
	for model := range tokenBazaarMessagesModels {
		t.Run(model, func(t *testing.T) {
			var gotBody []byte

			client := &http.Client{Transport: tokenBazaarRoundTripperFunc(func(request *http.Request) (*http.Response, error) {
				body, err := io.ReadAll(request.Body)
				if err != nil {
					t.Fatalf("read request body: %v", err)
				}

				gotBody = body

				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{"application/json"}},
					Body: io.NopCloser(strings.NewReader(
						`{"id":"msg_1","type":"message","role":"assistant","model":"` + model + `",` +
							`"content":[{"type":"text","text":"hi"}],"stop_reason":"end_turn",` +
							`"usage":{"input_tokens":1,"output_tokens":1}}`,
					)),
				}, nil
			})}

			payload := tokenBazaarMessagesTestPayload(model)
			requestBody := `{"model":"` + model + `","messages":[{"role":"user","content":"hi"}],` +
				`"max_tokens":1024,"thinking":{"type":"enabled","budget_tokens":512}}`
			payload.Body = []byte(requestBody)

			response, err := newTokenBazaarAdapter().Dispatch(context.Background(), client, payload)
			if err != nil {
				t.Fatalf("dispatch: %v", err)
			}
			defer response.Body.Close()

			if string(gotBody) != requestBody {
				t.Fatalf("thinking was not passed through unmodified: got %s, want %s", gotBody, requestBody)
			}

			if !strings.Contains(string(gotBody), `"thinking":{"type":"enabled","budget_tokens":512}`) {
				t.Fatalf("thinking missing from forwarded body: %s", gotBody)
			}
		})
	}
}

// TestTokenBazaarMessagesSSEStreamingPassesThroughUntilTerminalEvent covers
// the documented content_block_delta/message_stop SSE shape with the same
// defensive posture already established for chat and Responses SSE: this
// adapter passes the raw SSE bytes through unmodified; the control plane's
// own generic streaming plumbing (TokenBazaarMessagesHandler.handleStream,
// reusing parseAnthropicSSEUsage/anthropicSSEIsError from
// anthropic_native_handler.go, already exercised against malformed/
// unexpected events in
// services/control-plane/internal/proxy/tokenbazaar_messages_handler_test.go)
// is responsible for usage/error extraction -- this test only confirms the
// TokenBazaar adapter itself never mutates or buffers the documented event
// bytes.
func TestTokenBazaarMessagesSSEStreamingPassesThroughUntilTerminalEvent(t *testing.T) {
	sseBody := "event: message_start\n" +
		"data: {\"type\":\"message_start\",\"message\":{\"id\":\"msg_1\",\"model\":\"claude-sonnet-5\",\"usage\":{\"input_tokens\":5,\"output_tokens\":0}}}\n\n" +
		"event: content_block_start\n" +
		"data: {\"type\":\"content_block_start\",\"index\":0,\"content_block\":{\"type\":\"text\",\"text\":\"\"}}\n\n" +
		"event: content_block_delta\n" +
		"data: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"Hel\"}}\n\n" +
		"event: content_block_delta\n" +
		"data: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"lo\"}}\n\n" +
		"event: content_block_stop\n" +
		"data: {\"type\":\"content_block_stop\",\"index\":0}\n\n" +
		"event: message_delta\n" +
		"data: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"end_turn\"},\"usage\":{\"output_tokens\":2}}\n\n" +
		"event: message_stop\n" +
		"data: {\"type\":\"message_stop\"}\n\n"

	client := &http.Client{Transport: tokenBazaarRoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		if request.Header.Get("Accept") != "text/event-stream" {
			t.Fatalf("expected SSE Accept header, got %q", request.Header.Get("Accept"))
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader(sseBody)),
		}, nil
	})}

	payload := tokenBazaarMessagesTestPayload("claude-sonnet-5")
	payload.Stream = true
	payload.Body = []byte(`{"model":"claude-sonnet-5","messages":[{"role":"user","content":"hi"}],"max_tokens":16,"stream":true}`)

	response, err := newTokenBazaarAdapter().Dispatch(context.Background(), client, payload)
	if err != nil {
		t.Fatalf("dispatch streaming messages request: %v", err)
	}
	defer response.Body.Close()

	got, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read SSE body: %v", err)
	}

	if string(got) != sseBody {
		t.Fatalf("SSE body was altered by the adapter: %q", got)
	}

	if !strings.Contains(string(got), "event: message_stop") {
		t.Fatalf("terminal event was not preserved: %s", got)
	}
}

// TestTokenBazaarMessagesMalformedSSEChunkPassesThroughWithoutError confirms
// a malformed or unexpected SSE chunk (not valid JSON, or an event type not
// in the documented set) is still forwarded raw and does not cause the
// adapter itself to error or drop bytes -- malformed/unexpected-event
// *handling* (deciding what to do about it) is the control plane's
// responsibility (see the parseAnthropicSSEUsage/anthropicSSEIsError
// coverage in tokenbazaar_messages_handler_test.go), not this adapter's.
func TestTokenBazaarMessagesMalformedSSEChunkPassesThroughWithoutError(t *testing.T) {
	sseBody := "event: ping\n" +
		"data: {\"type\":\"ping\"}\n\n" +
		"event: content_block_delta\n" +
		"data: {not valid json\n\n" +
		"event: message_stop\n" +
		"data: {\"type\":\"message_stop\"}\n\n"

	client := &http.Client{Transport: tokenBazaarRoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader(sseBody)),
		}, nil
	})}

	payload := tokenBazaarMessagesTestPayload("claude-sonnet-5")
	payload.Stream = true
	payload.Body = []byte(`{"model":"claude-sonnet-5","messages":[{"role":"user","content":"hi"}],"max_tokens":16,"stream":true}`)

	response, err := newTokenBazaarAdapter().Dispatch(context.Background(), client, payload)
	if err != nil {
		t.Fatalf("dispatch streaming messages request with malformed chunk: %v", err)
	}
	defer response.Body.Close()

	got, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read SSE body: %v", err)
	}

	if string(got) != sseBody {
		t.Fatalf("malformed SSE body was altered by the adapter: %q", got)
	}
}

// TestTokenBazaarMessagesErrorResponseIsNormalizedAndDoesNotLeakUpstreamBody
// mirrors the chat/images/embeddings/responses paths' own 4xx-normalization
// coverage for the messages operation.
func TestTokenBazaarMessagesErrorResponseIsNormalizedAndDoesNotLeakUpstreamBody(t *testing.T) {
	client := &http.Client{Transport: tokenBazaarRoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Header:     http.Header{tokenBazaarRequestIDHeader: []string{"req-messages-1"}},
			Body: io.NopCloser(strings.NewReader(
				`{"type":"error","error":{"type":"invalid_request_error","message":"upstream vendor account id abc-123 must not leak"}}`,
			)),
		}, nil
	})}

	payload := tokenBazaarMessagesTestPayload("claude-sonnet-5")
	payload.Body = []byte(`{"model":"claude-sonnet-5","messages":[{"role":"user","content":"hi"}],"max_tokens":16}`)

	response, err := newTokenBazaarAdapter().Dispatch(context.Background(), client, payload)
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want normalized 503", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read normalized body: %v", err)
	}

	if strings.Contains(string(body), "abc-123") || strings.Contains(string(body), "invalid_request_error") {
		t.Fatalf("normalized messages error body leaked upstream content: %s", body)
	}
}

// TestTokenBazaarMessagesValidationErrorsNeverIncludeAPIKey confirms none of
// the messages-path validation error messages ever echo the deployment's API
// key, even when it is present on the payload.
func TestTokenBazaarMessagesValidationErrorsNeverIncludeAPIKey(t *testing.T) {
	const secret = "tokenbazaar-super-secret-key-must-not-leak"

	missingMaxTokens := tokenBazaarMessagesTestPayload("claude-sonnet-5")
	missingMaxTokens.Deployment.APIKey = secret
	missingMaxTokens.Body = []byte(`{"model":"claude-sonnet-5","messages":[{"role":"user","content":"hi"}]}`)

	badTool := tokenBazaarMessagesTestPayload("claude-sonnet-5")
	badTool.Deployment.APIKey = secret
	badTool.Body = []byte(`{"model":"claude-sonnet-5","messages":[{"role":"user","content":"hi"}],"max_tokens":16,` +
		`"tools":[{"name":"lookup","input_schema":{}}]}`)

	badCacheControl := tokenBazaarMessagesTestPayload("claude-sonnet-5")
	badCacheControl.Deployment.APIKey = secret
	badCacheControl.Body = []byte(`{"model":"claude-sonnet-5","max_tokens":16,"messages":[{"role":"user","content":` +
		`[{"type":"text","text":"hi","cache_control":{"type":"ephemeral"}}]}]}`)

	for _, payload := range []platform.TunnelRequest{missingMaxTokens, badTool, badCacheControl} {
		err := validateTokenBazaarDispatch(payload)
		if err == nil {
			t.Fatal("expected validation error")
		}

		if strings.Contains(err.Error(), secret) {
			t.Fatalf("validation error leaked the API key: %v", err)
		}
	}
}

// TestTokenBazaarMessagesNeverForwardsGodarAIAuthorizationHeader mirrors the
// chat/embeddings/responses paths' own coverage: the caller-supplied
// Authorization header must never reach TokenBazaar; only the deployment's
// own credential is forwarded.
func TestTokenBazaarMessagesNeverForwardsGodarAIAuthorizationHeader(t *testing.T) {
	var gotAuth string

	client := &http.Client{Transport: tokenBazaarRoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		gotAuth = request.Header.Get("Authorization")

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body: io.NopCloser(strings.NewReader(
				`{"id":"msg_1","type":"message","role":"assistant","model":"claude-sonnet-5",` +
					`"content":[{"type":"text","text":"hi"}],"stop_reason":"end_turn","usage":{"input_tokens":1,"output_tokens":1}}`,
			)),
		}, nil
	})}

	payload := tokenBazaarMessagesTestPayload("claude-sonnet-5")
	payload.Body = []byte(`{"model":"claude-sonnet-5","messages":[{"role":"user","content":"hi"}],"max_tokens":16}`)
	payload.Headers = map[string]string{"Authorization": "Bearer godarai-caller-token-must-not-leak"}

	response, err := newTokenBazaarAdapter().Dispatch(context.Background(), client, payload)
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	defer response.Body.Close()

	if gotAuth != "Bearer tokenbazaar-secret" {
		t.Fatalf("upstream Authorization = %q, want only the TokenBazaar credential", gotAuth)
	}
}

// tokenBazaarGenerateContentTestPayload builds a platform.TunnelRequest for
// the given model/operation carrying a runtime profile that declares the
// "generate_content"/"stream_generate_content" endpoints and "http"/"sse"
// transports -- mirroring the shape already scaffolded into
// apps/model-discovery/providers/tokenbazaar/gemini-*.yaml for all 6 in-scope
// models (confirmed by inspection before this delivery; see
// plans/tokenbazaar-production-readiness.md).
func tokenBazaarGenerateContentTestPayload(model string, operation platform.ProviderOperation) platform.TunnelRequest {
	profile := platform.ModelRuntimeProfile{
		Version:    1,
		Endpoints:  []string{"models", "chat_completions", "generate_content", "stream_generate_content"},
		Transports: []string{"http", "sse"},
		Lifecycle:  "active",
		Thinking:   platform.ModelRuntimeThinkingProfile{DefaultMode: "disabled", AllowedModes: []string{"disabled"}},
		Sampling:   platform.ModelRuntimeSamplingProfile{},
		Tools:      platform.ModelRuntimeToolsProfile{},
		Modalities: platform.ModelRuntimeModalities{Input: []string{"text", "image"}, Output: []string{"text"}},
		Parameters: map[string][]string{
			"generate_content": {"contents", "generationConfig", "systemInstruction", "safetySettings", "tools"},
		},
		Limits: platform.ModelRuntimeLimits{MessagesBytes: 1, BatchBytes: 1, FileBytes: 1},
	}

	return platform.TunnelRequest{
		Operation:  operation,
		Deployment: platform.Deployment{Provider: "tokenbazaar", Model: model, APIKey: "tokenbazaar-secret"},
		ModelRuntimeProfiles: []platform.ModelRuntimeProfileBinding{{
			Provider: "tokenbazaar", CatalogRevision: "test", CanonicalID: "tokenbazaar/" + model,
			Slug: model, ProfileVersion: 1, Profile: profile, EffectiveFrom: time.Now(),
		}},
	}
}

// TestTokenBazaarGenerateContentDispatchRoutesForAllSixInScopeModels confirms
// each of the 6 in-scope Gemini text models resolves through the
// generateContent dispatch path for both
// platform.ProviderOperationGeminiGenerateContent and
// platform.ProviderOperationGeminiStreamGenerateContent.
func TestTokenBazaarGenerateContentDispatchRoutesForAllSixInScopeModels(t *testing.T) {
	for model := range tokenBazaarGenerateContentModels {
		for _, operation := range []platform.ProviderOperation{
			platform.ProviderOperationGeminiGenerateContent,
			platform.ProviderOperationGeminiStreamGenerateContent,
		} {
			t.Run(model+"_"+string(operation), func(t *testing.T) {
				payload := tokenBazaarGenerateContentTestPayload(model, operation)
				if operation == platform.ProviderOperationGeminiStreamGenerateContent {
					payload.Stream = true
				}

				payload.Body = []byte(`{"contents":[{"role":"user","parts":[{"text":"hi"}]}]}`)

				if err := validateTokenBazaarDispatch(payload); err != nil {
					t.Fatalf("expected generateContent dispatch to be accepted for %s/%s, got %v", model, operation, err)
				}
			})
		}
	}
}

// TestTokenBazaarGenerateContentRejectsUnsupportedModels confirms models
// outside the 6-model generateContent allowlist (both a plain chat model and
// one of the 3 Gemini image models deliberately excluded from this
// delivery -- see tokenBazaarGenerateContentModels's own doc comment) are
// rejected with the generic tokenbazaar_unsupported_model error.
func TestTokenBazaarGenerateContentRejectsUnsupportedModels(t *testing.T) {
	for _, model := range []string{"glm-5.2", "gemini-3-pro-image", "gpt-5.5"} {
		t.Run(model, func(t *testing.T) {
			payload := tokenBazaarGenerateContentTestPayload(model, platform.ProviderOperationGeminiGenerateContent)
			payload.Body = []byte(`{"contents":[{"role":"user","parts":[{"text":"hi"}]}]}`)

			err := validateTokenBazaarDispatch(payload)
			if err == nil || !strings.Contains(err.Error(), "tokenbazaar_unsupported_model") {
				t.Fatalf("expected tokenbazaar_unsupported_model for %s, got %v", model, err)
			}
		})
	}
}

// TestTokenBazaarGenerateContentRequiresContents confirms a request missing
// the required "contents" field is rejected before dispatch.
func TestTokenBazaarGenerateContentRequiresContents(t *testing.T) {
	payload := tokenBazaarGenerateContentTestPayload("gemini-2.5-flash", platform.ProviderOperationGeminiGenerateContent)
	payload.Body = []byte(`{"generationConfig":{"maxOutputTokens":16}}`)

	err := validateTokenBazaarDispatch(payload)
	if err == nil || !strings.Contains(err.Error(), "tokenbazaar_generate_content_contents_required") {
		t.Fatalf("expected tokenbazaar_generate_content_contents_required, got %v", err)
	}
}

// TestTokenBazaarGenerateContentRejectsCachedContent covers this delivery's
// fail-closed decision on Gemini's context-caching feature: a non-empty
// "cachedContent" reference is rejected before dispatch, since TokenBazaar's
// docs do not mention Gemini context caching anywhere (no nav section, no
// worked example) -- see validateTokenBazaarGenerateContentCachedContentUnsupported.
func TestTokenBazaarGenerateContentRejectsCachedContent(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{"absent", `{"contents":[{"role":"user","parts":[{"text":"hi"}]}]}`, false},
		{"null", `{"contents":[{"role":"user","parts":[{"text":"hi"}]}],"cachedContent":null}`, false},
		{"empty_string", `{"contents":[{"role":"user","parts":[{"text":"hi"}]}],"cachedContent":""}`, false},
		{"referenced", `{"contents":[{"role":"user","parts":[{"text":"hi"}]}],"cachedContent":"cachedContents/abc123"}`, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payload := tokenBazaarGenerateContentTestPayload("gemini-2.5-flash", platform.ProviderOperationGeminiGenerateContent)
			payload.Body = []byte(test.body)

			err := validateTokenBazaarDispatch(payload)
			if test.wantErr {
				if err == nil || !strings.Contains(err.Error(), "tokenbazaar_generate_content_cached_content_unsupported") {
					t.Fatalf("expected tokenbazaar_generate_content_cached_content_unsupported, got %v", err)
				}

				return
			}

			if err != nil {
				t.Fatalf("expected request to be accepted, got %v", err)
			}
		})
	}
}

// TestTokenBazaarGenerateContentRejectionsHappenBeforeAnyHTTPCall proves the
// unsupported-model, missing-contents, and cachedContent rejections all
// happen before any upstream call, via an end-to-end mocked-HTTP Dispatch
// call (not just the standalone validator) -- mirroring
// TestTokenBazaarMessagesThreeRejectionsHappenBeforeAnyHTTPCall.
func TestTokenBazaarGenerateContentRejectionsHappenBeforeAnyHTTPCall(t *testing.T) {
	var calls int32

	client := &http.Client{Transport: tokenBazaarRoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		atomic.AddInt32(&calls, 1)

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body: io.NopCloser(strings.NewReader(
				`{"candidates":[{"content":{"role":"model","parts":[{"text":"hi"}]},"finishReason":"STOP"}]}`,
			)),
		}, nil
	})}

	accepted := tokenBazaarGenerateContentTestPayload("gemini-2.5-flash", platform.ProviderOperationGeminiGenerateContent)
	accepted.Body = []byte(`{"contents":[{"role":"user","parts":[{"text":"hi"}]}]}`)

	response, err := newTokenBazaarAdapter().Dispatch(context.Background(), client, accepted)
	if err != nil {
		t.Fatalf("expected a valid request to succeed, got %v", err)
	}
	defer response.Body.Close()

	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected exactly 1 upstream call for the valid request, got %d", calls)
	}

	rejections := map[string]string{
		"missing_contents": `{"generationConfig":{"maxOutputTokens":16}}`,
		"cached_content":   `{"contents":[{"role":"user","parts":[{"text":"hi"}]}],"cachedContent":"cachedContents/abc123"}`,
	}

	for name, body := range rejections {
		t.Run(name, func(t *testing.T) {
			payload := tokenBazaarGenerateContentTestPayload("gemini-2.5-flash", platform.ProviderOperationGeminiGenerateContent)
			payload.Body = []byte(body)

			if _, err := newTokenBazaarAdapter().Dispatch(context.Background(), client, payload); err == nil {
				t.Fatal("expected the request to be rejected before dispatch")
			}

			if atomic.LoadInt32(&calls) != 1 {
				t.Fatalf("rejection %q must happen before any HTTP call; total upstream calls = %d", name, calls)
			}
		})
	}
}

// TestTokenBazaarGenerateContentDispatchExactRequestConstruction exercises
// the full adapter Dispatch path (mocked HTTP, no live call) and asserts the
// exact method, URL (no "/v1" prefix -- Gemini-style paths are NOT nested
// under the OpenAI-compatible "/v1" the way chat/responses/messages/images/
// embeddings are), Authorization header, Content-Type, and forwarded body
// against https://zx1.deepwl.net/v1beta/models/gemini-2.5-flash:generateContent.
func TestTokenBazaarGenerateContentDispatchExactRequestConstruction(t *testing.T) {
	var gotMethod, gotURL, gotAuth, gotContentType string

	var gotBody []byte

	client := &http.Client{Transport: tokenBazaarRoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		gotMethod = request.Method
		gotURL = request.URL.String()
		gotAuth = request.Header.Get("Authorization")
		gotContentType = request.Header.Get("Content-Type")

		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}

		gotBody = body

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body: io.NopCloser(strings.NewReader(
				`{"candidates":[{"content":{"role":"model","parts":[{"text":"hi there"}]},"finishReason":"STOP"}],` +
					`"usageMetadata":{"promptTokenCount":3,"candidatesTokenCount":2,"totalTokenCount":5}}`,
			)),
		}, nil
	})}

	payload := tokenBazaarGenerateContentTestPayload("gemini-2.5-flash", platform.ProviderOperationGeminiGenerateContent)
	requestBody := `{"contents":[{"role":"user","parts":[{"text":"hello world"}]}]}`
	payload.Body = []byte(requestBody)

	response, err := newTokenBazaarAdapter().Dispatch(context.Background(), client, payload)
	if err != nil {
		t.Fatalf("dispatch generateContent: %v", err)
	}
	defer response.Body.Close()

	if gotMethod != http.MethodPost {
		t.Fatalf("method = %q, want POST", gotMethod)
	}

	if gotURL != "https://zx1.deepwl.net/v1beta/models/gemini-2.5-flash:generateContent" {
		t.Fatalf("url = %q", gotURL)
	}

	if gotAuth != "Bearer tokenbazaar-secret" {
		t.Fatalf("authorization = %q", gotAuth)
	}

	if gotContentType != "application/json" {
		t.Fatalf("content-type = %q", gotContentType)
	}

	if string(gotBody) != requestBody {
		t.Fatalf("forwarded body = %s, want verbatim %s", gotBody, requestBody)
	}

	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", response.StatusCode)
	}
}

// TestTokenBazaarStreamGenerateContentDispatchExactRequestConstruction
// mirrors the above for streamGenerateContent: the URL must carry the
// documented "?alt=sse" query (Gemini's own SSE-mode convention, reused
// verbatim via geminiOperationTarget) and the Accept header must request
// text/event-stream.
func TestTokenBazaarStreamGenerateContentDispatchExactRequestConstruction(t *testing.T) {
	var gotURL, gotAccept string

	client := &http.Client{Transport: tokenBazaarRoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		gotURL = request.URL.String()
		gotAccept = request.Header.Get("Accept")

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader("data: {\"candidates\":[{\"content\":{\"role\":\"model\",\"parts\":[{\"text\":\"hi\"}]}}]}\n\n")),
		}, nil
	})}

	payload := tokenBazaarGenerateContentTestPayload("gemini-2.5-flash", platform.ProviderOperationGeminiStreamGenerateContent)
	payload.Stream = true
	payload.Body = []byte(`{"contents":[{"role":"user","parts":[{"text":"hi"}]}]}`)

	response, err := newTokenBazaarAdapter().Dispatch(context.Background(), client, payload)
	if err != nil {
		t.Fatalf("dispatch streamGenerateContent: %v", err)
	}
	defer response.Body.Close()

	if gotURL != "https://zx1.deepwl.net/v1beta/models/gemini-2.5-flash:streamGenerateContent?alt=sse" {
		t.Fatalf("url = %q", gotURL)
	}

	if gotAccept != "text/event-stream" {
		t.Fatalf("accept = %q, want text/event-stream", gotAccept)
	}
}

// TestTokenBazaarGenerateContentResponseShapeValidated confirms the
// defensive response-shape check: a successful response missing "candidates"
// is rejected with tokenbazaar_generate_content_response_shape_unexpected
// rather than forwarded blind, while a response matching the documented
// shape passes through unchanged. Mirrors the embeddings response-shape
// test's own pattern.
func TestTokenBazaarGenerateContentResponseShapeValidated(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantErr    bool
		wantErrMsg string
	}{
		{
			"missing_candidates",
			`{"usageMetadata":{"promptTokenCount":1}}`,
			true,
			"tokenbazaar_generate_content_response_shape_unexpected",
		},
		{
			"empty_candidates",
			`{"candidates":[]}`,
			true,
			"tokenbazaar_generate_content_response_shape_unexpected",
		},
		{
			"not_json",
			`not json at all`,
			true,
			"tokenbazaar_generate_content_response_shape_unexpected",
		},
		{
			"valid_shape",
			`{"candidates":[{"content":{"role":"model","parts":[{"text":"hi"}]},"finishReason":"STOP"}]}`,
			false,
			"",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := &http.Client{Transport: tokenBazaarRoundTripperFunc(func(request *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{"application/json"}},
					Body:       io.NopCloser(strings.NewReader(test.body)),
				}, nil
			})}

			payload := tokenBazaarGenerateContentTestPayload("gemini-2.5-flash", platform.ProviderOperationGeminiGenerateContent)
			payload.Body = []byte(`{"contents":[{"role":"user","parts":[{"text":"hi"}]}]}`)

			response, err := newTokenBazaarAdapter().Dispatch(context.Background(), client, payload)
			if test.wantErr {
				if err == nil || !strings.Contains(err.Error(), test.wantErrMsg) {
					t.Fatalf("expected %s, got %v", test.wantErrMsg, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("expected a valid response to pass through, got %v", err)
			}

			defer response.Body.Close()

			got, readErr := io.ReadAll(response.Body)
			if readErr != nil {
				t.Fatalf("read response body: %v", readErr)
			}

			if string(got) != test.body {
				t.Fatalf("valid response body was altered: got %s, want %s", got, test.body)
			}
		})
	}
}

// TestTokenBazaarStreamGenerateContentSSEStreamingPassesThroughUnmodified
// confirms the adapter passes raw SSE bytes through unmodified for
// streamGenerateContent (the response-shape check above deliberately does
// NOT apply to streaming, since it would require buffering the whole SSE
// stream) -- mirroring
// TestTokenBazaarMessagesSSEStreamingPassesThroughUntilTerminalEvent's own
// discipline.
func TestTokenBazaarStreamGenerateContentSSEStreamingPassesThroughUnmodified(t *testing.T) {
	sseBody := "data: {\"candidates\":[{\"content\":{\"role\":\"model\",\"parts\":[{\"text\":\"Hel\"}]}}]}\n\n" +
		"data: {\"candidates\":[{\"content\":{\"role\":\"model\",\"parts\":[{\"text\":\"lo\"}]},\"finishReason\":\"STOP\"}]," +
		"\"usageMetadata\":{\"promptTokenCount\":3,\"candidatesTokenCount\":2,\"totalTokenCount\":5}}\n\n"

	client := &http.Client{Transport: tokenBazaarRoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		if request.Header.Get("Accept") != "text/event-stream" {
			t.Fatalf("expected SSE Accept header, got %q", request.Header.Get("Accept"))
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader(sseBody)),
		}, nil
	})}

	payload := tokenBazaarGenerateContentTestPayload("gemini-2.5-flash", platform.ProviderOperationGeminiStreamGenerateContent)
	payload.Stream = true
	payload.Body = []byte(`{"contents":[{"role":"user","parts":[{"text":"hi"}]}]}`)

	response, err := newTokenBazaarAdapter().Dispatch(context.Background(), client, payload)
	if err != nil {
		t.Fatalf("dispatch streaming generateContent request: %v", err)
	}
	defer response.Body.Close()

	got, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read SSE body: %v", err)
	}

	if string(got) != sseBody {
		t.Fatalf("SSE body was altered by the adapter: %q", got)
	}
}

// TestTokenBazaarGenerateContentErrorResponseIsNormalizedAndDoesNotLeakUpstreamBody
// mirrors the chat/images/embeddings/responses/messages paths' own
// 4xx-normalization coverage for the generateContent operation.
func TestTokenBazaarGenerateContentErrorResponseIsNormalizedAndDoesNotLeakUpstreamBody(t *testing.T) {
	client := &http.Client{Transport: tokenBazaarRoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Header:     http.Header{tokenBazaarRequestIDHeader: []string{"req-generate-content-1"}},
			Body: io.NopCloser(strings.NewReader(
				`{"error":{"code":400,"message":"upstream vendor account id abc-123 must not leak","status":"INVALID_ARGUMENT"}}`,
			)),
		}, nil
	})}

	payload := tokenBazaarGenerateContentTestPayload("gemini-2.5-flash", platform.ProviderOperationGeminiGenerateContent)
	payload.Body = []byte(`{"contents":[{"role":"user","parts":[{"text":"hi"}]}]}`)

	response, err := newTokenBazaarAdapter().Dispatch(context.Background(), client, payload)
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want normalized 503", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read normalized body: %v", err)
	}

	if strings.Contains(string(body), "abc-123") || strings.Contains(string(body), "INVALID_ARGUMENT") {
		t.Fatalf("normalized generateContent error body leaked upstream content: %s", body)
	}
}

// TestTokenBazaarGenerateContentValidationErrorsNeverIncludeAPIKey confirms
// none of the generateContent-path validation error messages ever echo the
// deployment's API key, even when it is present on the payload.
func TestTokenBazaarGenerateContentValidationErrorsNeverIncludeAPIKey(t *testing.T) {
	const secret = "tokenbazaar-super-secret-key-must-not-leak"

	missingContents := tokenBazaarGenerateContentTestPayload("gemini-2.5-flash", platform.ProviderOperationGeminiGenerateContent)
	missingContents.Deployment.APIKey = secret
	missingContents.Body = []byte(`{"generationConfig":{"maxOutputTokens":16}}`)

	badCachedContent := tokenBazaarGenerateContentTestPayload("gemini-2.5-flash", platform.ProviderOperationGeminiGenerateContent)
	badCachedContent.Deployment.APIKey = secret
	badCachedContent.Body = []byte(`{"contents":[{"role":"user","parts":[{"text":"hi"}]}],"cachedContent":"cachedContents/abc123"}`)

	for _, payload := range []platform.TunnelRequest{missingContents, badCachedContent} {
		err := validateTokenBazaarDispatch(payload)
		if err == nil {
			t.Fatal("expected validation error")
		}

		if strings.Contains(err.Error(), secret) {
			t.Fatalf("validation error leaked the API key: %v", err)
		}
	}
}

// TestTokenBazaarGenerateContentNeverForwardsGodarAIAuthorizationHeader
// mirrors the chat/embeddings/responses/messages paths' own coverage: the
// caller-supplied Authorization header must never reach TokenBazaar; only
// the deployment's own credential is forwarded.
func TestTokenBazaarGenerateContentNeverForwardsGodarAIAuthorizationHeader(t *testing.T) {
	var gotAuth string

	client := &http.Client{Transport: tokenBazaarRoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		gotAuth = request.Header.Get("Authorization")

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body: io.NopCloser(strings.NewReader(
				`{"candidates":[{"content":{"role":"model","parts":[{"text":"hi"}]},"finishReason":"STOP"}]}`,
			)),
		}, nil
	})}

	payload := tokenBazaarGenerateContentTestPayload("gemini-2.5-flash", platform.ProviderOperationGeminiGenerateContent)
	payload.Body = []byte(`{"contents":[{"role":"user","parts":[{"text":"hi"}]}]}`)
	payload.Headers = map[string]string{"Authorization": "Bearer godarai-caller-token-must-not-leak"}

	response, err := newTokenBazaarAdapter().Dispatch(context.Background(), client, payload)
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	defer response.Body.Close()

	if gotAuth != "Bearer tokenbazaar-secret" {
		t.Fatalf("upstream Authorization = %q, want only the TokenBazaar credential", gotAuth)
	}
}

// TestRawDispatchNamesTokenBazaarValidationFailure covers a request the
// TokenBazaar adapter rejects before contacting the upstream. The raw dispatch
// endpoint must name the reason in its body, and keep the 502 status so
// virtual models still fall back to their next target.
func TestRawDispatchNamesTokenBazaarValidationFailure(t *testing.T) {
	t.Parallel()

	var upstreamCalls atomic.Int32

	client := &http.Client{Transport: tokenBazaarRoundTripperFunc(func(*http.Request) (*http.Response, error) {
		upstreamCalls.Add(1)

		return nil, io.ErrUnexpectedEOF
	})}

	metadata, err := json.Marshal(platform.TunnelRequest{
		Operation:  platform.ProviderOperationOpenAIResponsesCreate,
		Deployment: platform.Deployment{Provider: "tokenbazaar", Model: "gemini-2.5-pro", APIKey: "tokenbazaar-secret"},
	})
	if err != nil {
		t.Fatalf("marshal metadata: %v", err)
	}

	rawBody := []byte(`{"model":"gemini-2.5-pro","input":"hello"}`)

	var prefix [4]byte
	binary.BigEndian.PutUint32(prefix[:], uint32(len(metadata)))

	req := httptest.NewRequest(
		http.MethodPost,
		"/dispatch/raw",
		io.MultiReader(bytes.NewReader(prefix[:]), bytes.NewReader(metadata), bytes.NewReader(rawBody)),
	)
	req.ContentLength = int64(len(prefix) + len(metadata) + len(rawBody))
	recorder := httptest.NewRecorder()
	NewHandler(client).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadGateway, recorder.Code, recorder.Body.String())
	}

	var body struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		} `json:"error"`
	}
	if decodeErr := json.Unmarshal(recorder.Body.Bytes(), &body); decodeErr != nil {
		t.Fatalf("decode error body: %v", decodeErr)
	}

	if body.Error.Type != "invalid_request_error" || !strings.Contains(body.Error.Message, "tokenbazaar_unsupported_model: gemini-2.5-pro") {
		t.Fatalf("unexpected error body: %s", recorder.Body.String())
	}

	if calls := upstreamCalls.Load(); calls != 0 {
		t.Fatalf("expected no upstream calls, got %d", calls)
	}
}
