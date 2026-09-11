package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func TestDeepSeekAdapterUsesFixedChatContractAndOpaqueIsolation(t *testing.T) {
	t.Setenv("DEEPSEEK_USER_ID_HMAC_SECRET", "test-only-isolation-secret")

	var gotMethod, gotPath, gotAuthorization, gotBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		gotMethod, gotPath = request.Method, request.URL.Path
		gotAuthorization = request.Header.Get("Authorization")

		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
			return
		}

		gotBody = string(body)

		w.Header().Set("Content-Type", "text/event-stream")

		if _, err := io.WriteString(w, ": keep-alive\n\ndata: [DONE]\n\n"); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	defer server.Close()

	payload := deepSeekTestPayload("tenant-a", "deepseek-v4-flash")
	adapter := &deepSeekAdapter{compat: &openAICompatibleAdapter{
		defaultBaseURL: server.URL,
		authHeader:     "Authorization",
	}}

	response, err := adapter.Dispatch(context.Background(), server.Client(), payload)
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}

	if gotMethod != http.MethodPost || gotPath != "/chat/completions" {
		t.Fatalf("unexpected upstream request: %s %s", gotMethod, gotPath)
	}

	if gotAuthorization != "Bearer deepseek-secret" {
		t.Fatalf("unexpected authorization: %q", gotAuthorization)
	}

	if !strings.Contains(gotBody, `"user_id":"gdr_`) || strings.Contains(gotBody, "tenant-a") {
		t.Fatalf("request body does not contain an opaque user_id: %s", gotBody)
	}

	if string(responseBody) != ": keep-alive\n\ndata: [DONE]\n\n" {
		t.Fatalf("SSE body was not preserved: %q", responseBody)
	}
}

func TestDeepSeekIsolationSeparatesTenantsAndFailsClosed(t *testing.T) {
	first, err := deepSeekOpaqueUserID("tenant-a", "account-1", "secret")
	if err != nil {
		t.Fatalf("first opaque ID: %v", err)
	}

	second, err := deepSeekOpaqueUserID("tenant-b", "account-1", "secret")
	if err != nil {
		t.Fatalf("second opaque ID: %v", err)
	}

	if first == second || strings.Contains(first, "tenant") || !deepSeekUserIDPattern.MatchString(first) {
		t.Fatalf("opaque IDs are not safely separated: first=%q second=%q", first, second)
	}

	if _, err := deepSeekOpaqueUserID("tenant-a", "account-1", ""); err == nil || !strings.Contains(err.Error(), "isolation_unavailable") {
		t.Fatalf("missing isolation secret must fail closed: %v", err)
	}
}

func TestDeepSeekAdapterRejectsUnsupportedOperations(t *testing.T) {
	payload := deepSeekTestPayload("tenant-a", "deepseek-v4-pro")

	payload.Operation = platform.ProviderOperationOpenAIResponsesGet
	if err := validateDeepSeekDispatch(payload); err == nil || !strings.Contains(err.Error(), "unsupported_operation") {
		t.Fatalf("Responses resources must be rejected: %v", err)
	}
}

func TestDeepSeekAdapterAcceptsBYOKGatewayHosts(t *testing.T) {
	for _, raw := range []string{
		"https://api.deepseek.com",
		"https://api.deepseek.com/v1",
		"https://api.deepseek.com.evil.example",
		"https://my-gateway.example.com/deepseek/v1",
	} {
		if err := validateDeepSeekBaseURL(raw); err != nil {
			t.Fatalf("BYOK API base rejected: %s: %v", raw, err)
		}
	}
}

func TestDeepSeekAdapterRejectsUnsafeHosts(t *testing.T) {
	for _, raw := range []string{
		"https://127.0.0.1/v1",
		"https://localhost/v1",
		"http://api.deepseek.com",
		"https://user:pass@api.deepseek.com",
		"https://api.deepseek.com/v1?foo=bar",
	} {
		if err := validateDeepSeekBaseURL(raw); err == nil {
			t.Fatalf("unsafe API base accepted: %s", raw)
		}
	}
}

func TestDeepSeekAdapterSupportsResponsesSSEWithoutDoneSentinel(t *testing.T) {
	t.Setenv("DEEPSEEK_USER_ID_HMAC_SECRET", "test-only-isolation-secret")

	const upstreamBody = "event: response.completed\ndata: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_1\",\"model\":\"deepseek-v4-pro\",\"status\":\"completed\"}}\n\n"

	var gotPath, gotBody string

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		gotPath = request.URL.Path

		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)

			return
		}

		gotBody = string(body)

		writer.Header().Set("Content-Type", "text/event-stream")

		if _, err := io.WriteString(writer, upstreamBody); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	defer server.Close()

	payload := deepSeekTestPayload("tenant-a", "deepseek-v4-pro")
	payload.Operation = platform.ProviderOperationOpenAIResponsesCreate
	payload.Body = []byte(`{"model":"deepseek-v4-pro","input":"hello","stream":true}`)
	adapter := &deepSeekAdapter{compat: &openAICompatibleAdapter{
		defaultBaseURL: server.URL,
		authHeader:     "Authorization",
	}}

	response, err := adapter.Dispatch(context.Background(), server.Client(), payload)
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}

	if gotPath != "/responses" {
		t.Fatalf("unexpected Responses path: %s", gotPath)
	}

	if !strings.Contains(gotBody, `"user":"gdr_`) || strings.Contains(gotBody, "tenant-a") || strings.Contains(gotBody, "user_id") {
		t.Fatalf("Responses request does not contain a safe managed user: %s", gotBody)
	}

	if string(responseBody) != upstreamBody || strings.Contains(string(responseBody), "[DONE]") {
		t.Fatalf("Responses SSE body was changed: %q", responseBody)
	}
}

func TestDeepSeekBYOKDoesNotRequireSharedCredentialIsolationSecret(t *testing.T) {
	payload := deepSeekTestPayload("tenant-a", "deepseek-v4-flash")
	payload.Deployment.CredentialSource = platform.CredentialSourceUserBYOK

	body, err := deepSeekRequestBody(payload)
	if err != nil {
		t.Fatalf("BYOK request body: %v", err)
	}

	if strings.Contains(string(body), "user_id") {
		t.Fatalf("BYOK request unexpectedly contains user_id: %s", body)
	}
}

func TestDeepSeekRejectsNullJSONBody(t *testing.T) {
	t.Parallel()

	payload := deepSeekTestPayload("tenant-a", "deepseek-v4-flash")
	payload.Body = []byte("null")

	if err := validateDeepSeekDispatch(payload); err == nil || !strings.Contains(err.Error(), "deepseek_invalid_json") {
		t.Fatalf("null request body must be rejected, got %v", err)
	}
}

func TestDeepSeekAdapterRetriesOnlyModelsList(t *testing.T) {
	var modelListAttempts atomic.Int32

	modelsServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		attempt := modelListAttempts.Add(1)
		if attempt == 1 {
			writer.WriteHeader(http.StatusServiceUnavailable)

			return
		}

		writer.Header().Set("Content-Type", "application/json")

		_, err := io.WriteString(writer, `{"data":[{"id":"deepseek-v4-flash"}]}`)
		if err != nil {
			t.Errorf("write models response: %v", err)
		}
	}))
	defer modelsServer.Close()

	payload := deepSeekTestPayload("tenant-a", "deepseek-v4-flash")
	payload.Operation = platform.ProviderOperationOpenAIModelsList
	adapter := &deepSeekAdapter{compat: &openAICompatibleAdapter{
		defaultBaseURL: modelsServer.URL,
		authHeader:     "Authorization",
	}}

	response, err := adapter.Dispatch(context.Background(), modelsServer.Client(), payload)
	if err != nil {
		t.Fatalf("models dispatch: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK || modelListAttempts.Load() != 2 {
		t.Fatalf("models retry status=%d attempts=%d", response.StatusCode, modelListAttempts.Load())
	}

	var chatAttempts atomic.Int32

	chatServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		chatAttempts.Add(1)
		writer.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer chatServer.Close()

	adapter = &deepSeekAdapter{compat: &openAICompatibleAdapter{
		defaultBaseURL: chatServer.URL,
		authHeader:     "Authorization",
	}}
	chatPayload := deepSeekTestPayload("tenant-a", "deepseek-v4-flash")
	chatPayload.Deployment.CredentialSource = platform.CredentialSourceUserBYOK

	response, err = adapter.Dispatch(context.Background(), chatServer.Client(), chatPayload)
	if err != nil {
		t.Fatalf("chat dispatch: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusServiceUnavailable || chatAttempts.Load() != 1 {
		t.Fatalf("billable chat was retried: status=%d attempts=%d", response.StatusCode, chatAttempts.Load())
	}
}

func TestDeepSeekAdapterPreservesReasoningToolLoopBodies(t *testing.T) {
	for _, model := range []string{"deepseek-v4-flash", "deepseek-v4-pro"} {
		for _, thinking := range []string{"enabled", "disabled"} {
			for _, stream := range []bool{false, true} {
				name := model + "_" + thinking
				if stream {
					name += "_stream"
				} else {
					name += "_sync"
				}

				t.Run(name, func(t *testing.T) {
					requestBody := `{"model":"` + model + `","messages":[` +
						`{"role":"assistant","reasoning_content":"opaque reasoning","tool_calls":[{"id":"call_1","type":"function","function":{"name":"lookup","arguments":"{\"id\":7}"}}]},` +
						`{"role":"tool","tool_call_id":"call_1","content":"tool result"}],` +
						`"thinking":{"type":"` + thinking + `"},"stream":` + fmt.Sprint(stream) + `}`

					responseBody := `{"model":"` + model + `","choices":[{"index":0,"message":{"role":"assistant","reasoning_content":"next reasoning","content":"answer"},"finish_reason":"stop"}]}`

					contentType := "application/json"
					if stream {
						contentType = "text/event-stream"
						responseBody = ": keep-alive\n\ndata: " + responseBody + "\n\ndata: [DONE]\n\n"
					}

					server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
						body, err := io.ReadAll(request.Body)
						if err != nil {
							t.Errorf("read request body: %v", err)

							return
						}

						if string(body) != requestBody {
							t.Errorf("request body changed:\nwant %s\ngot  %s", requestBody, body)
						}

						writer.Header().Set("Content-Type", contentType)

						if _, err := io.WriteString(writer, responseBody); err != nil {
							t.Errorf("write response: %v", err)
						}
					}))
					defer server.Close()

					payload := deepSeekTestPayload("tenant-a", model)
					payload.Stream = stream
					payload.Body = []byte(requestBody)
					payload.Deployment.CredentialSource = platform.CredentialSourceUserBYOK
					adapter := &deepSeekAdapter{compat: &openAICompatibleAdapter{
						defaultBaseURL: server.URL,
						authHeader:     "Authorization",
					}}

					response, err := adapter.Dispatch(context.Background(), server.Client(), payload)
					if err != nil {
						t.Fatalf("dispatch: %v", err)
					}
					defer response.Body.Close()

					got, err := io.ReadAll(response.Body)
					if err != nil {
						t.Fatalf("read response: %v", err)
					}

					if string(got) != responseBody {
						t.Fatalf("response body changed:\nwant %q\ngot  %q", responseBody, got)
					}
				})
			}
		}
	}
}

func TestDeepSeekAdapterPropagatesCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	adapter := newDeepSeekAdapter()
	payload := deepSeekTestPayload("tenant-a", "deepseek-v4-flash")
	payload.Deployment.CredentialSource = platform.CredentialSourceUserBYOK

	_, err := adapter.Dispatch(ctx, http.DefaultClient, payload)
	if err == nil || !strings.Contains(err.Error(), context.Canceled.Error()) {
		t.Fatalf("cancellation was not propagated: %v", err)
	}
}

func deepSeekTestPayload(tenantID, model string) platform.TunnelRequest {
	profileVersion := 1
	endpoints := []string{"models", "chat_completions"}
	parameters := map[string][]string(nil)

	if model == "deepseek-v4-flash" || model == "deepseek-v4-pro" {
		profileVersion = map[string]int{
			"deepseek-v4-flash": 3,
			"deepseek-v4-pro":   2,
		}[model]

		endpoints = append(endpoints, "responses")
		parameters = map[string][]string{
			"responses": {"input", "instructions", "max_output_tokens", "reasoning", "stream", "temperature", "top_p", "top_logprobs", "text", "tools", "tool_choice"},
		}
	}

	return platform.TunnelRequest{
		TenantID: tenantID, Stream: true,
		Operation: platform.ProviderOperationChatCompletions,
		Deployment: platform.Deployment{
			Provider: "deepseek", Model: model, APIKey: "deepseek-secret",
			CredentialSource:  platform.CredentialSourcePlatformManaged,
			ProviderAccountID: "account-1",
		},
		Body: []byte(`{"model":"` + model + `","messages":[{"role":"user","content":"hello"}],"stream":true}`),
		ModelRuntimeProfiles: []platform.ModelRuntimeProfileBinding{{
			Provider: "deepseek", CatalogRevision: "deepseek-test", CanonicalID: "deepseek/" + model,
			Slug: model, ProfileVersion: profileVersion, EffectiveFrom: time.Now().UTC(),
			Profile: platform.ModelRuntimeProfile{
				Version: profileVersion, Endpoints: endpoints, Transports: []string{"http", "sse"},
				Lifecycle: map[string]string{
					"deepseek-v4-flash": "preview",
					"deepseek-v4-pro":   "active",
				}[model],
				Thinking:   platform.ModelRuntimeThinkingProfile{DefaultMode: "adaptive", AllowedModes: []string{"adaptive", "disabled"}, EffortValues: []string{"low", "medium", "high", "xhigh", "max"}},
				Sampling:   platform.ModelRuntimeSamplingProfile{Parameters: []string{"temperature", "top_p"}, DisallowedWithThinking: []string{"temperature", "top_p"}},
				Tools:      platform.ModelRuntimeToolsProfile{Custom: true, Supported: []string{"function_calling"}, StructuredOutput: true},
				Modalities: platform.ModelRuntimeModalities{Input: []string{"text"}, Output: []string{"text"}},
				Parameters: parameters,
				Limits: platform.ModelRuntimeLimits{
					ContextWindow: 1048576, MaxInputTokens: 1048576, MaxOutputTokens: 393216,
					MessagesBytes: 32 * 1024 * 1024, BatchBytes: 256 * 1024 * 1024, FileBytes: 500 * 1024 * 1024,
				},
			},
		}},
		Headers: map[string]string{"Content-Type": "application/json"},
	}
}
