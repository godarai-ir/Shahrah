package server

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func TestAnthropicRetryHonorsRetryAfterForIdempotentRequest(t *testing.T) {
	var attempts atomic.Int32
	var bodies []string
	upstream := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, request *http.Request) {
			attempts.Add(1)
			body, _ := io.ReadAll(request.Body)
			bodies = append(bodies, string(body))
			if attempts.Load() == 1 {
				w.Header().Set("retry-after", "0")
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			_, _ = io.WriteString(w, `{"ok":true}`)
		},
	))
	defer upstream.Close()

	request, err := http.NewRequestWithContext(
		t.Context(),
		http.MethodPost,
		upstream.URL,
		bytes.NewReader([]byte(`{"probe":true}`)),
	)
	if err != nil {
		t.Fatal(err)
	}
	response, err := doAnthropicRequest(
		t.Context(),
		upstream.Client(),
		request,
		true,
	)
	if err != nil {
		t.Fatalf("doAnthropicRequest: %v", err)
	}
	defer response.Body.Close()
	if attempts.Load() != 2 {
		t.Fatalf("expected two attempts, got %d", attempts.Load())
	}
	if len(bodies) != 2 || bodies[0] != bodies[1] {
		t.Fatalf("request body was not replayed: %#v", bodies)
	}
}

func TestAnthropicMessagesAreNeverRetried(t *testing.T) {
	var attempts atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			attempts.Add(1)
			w.WriteHeader(http.StatusInternalServerError)
		},
	))
	defer upstream.Close()

	binding := anthropicTestRuntimeProfile("claude-opus-4-7")
	response, err := (&anthropicAdapter{}).Dispatch(
		t.Context(),
		upstream.Client(),
		platform.TunnelRequest{
			Body: []byte(`{
				"model":"claude-opus-4-7",
				"messages":[{"role":"user","content":"hello"}]
			}`),
			Deployment: platform.Deployment{
				Provider: "anthropic",
				APIBase:  upstream.URL,
				APIKey:   "sk-test",
			},
			ModelRuntimeProfiles: []platform.ModelRuntimeProfileBinding{
				binding,
			},
		},
	)
	if err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	defer response.Body.Close()
	if attempts.Load() != 1 {
		t.Fatalf("messages request retried %d times", attempts.Load())
	}
}

func TestAnthropicRetryOperationAllowlist(t *testing.T) {
	if !anthropicOperationIsIdempotent(
		platform.ProviderOperationAnthropicCountTokens,
	) {
		t.Fatal("count_tokens must be retryable")
	}
	for _, operation := range []platform.ProviderOperation{
		platform.ProviderOperationAnthropicMessages,
		platform.ProviderOperationAnthropicFilesCreate,
		platform.ProviderOperationAnthropicBatchesCreate,
		platform.ProviderOperationAnthropicBatchesCancel,
	} {
		if anthropicOperationIsIdempotent(operation) {
			t.Fatalf("operation %q must not be retried", operation)
		}
	}
}
