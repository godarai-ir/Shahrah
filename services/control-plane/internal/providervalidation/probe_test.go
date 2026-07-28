package providervalidation

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func TestSupportedProviderContractsBuildExactProbeRequests(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		request     Request
		expectedURL string
		headerName  string
		headerValue string
	}{
		{
			name: "anthropic",
			request: Request{
				Provider: "anthropic", AuthMode: "api_key",
				Config:  map[string]interface{}{"api_version": "2023-06-01"},
				Secrets: map[string]string{"api_key": "anthropic-secret"},
			},
			expectedURL: "https://api.anthropic.com/v1/models?limit=1",
			headerName:  "x-api-key", headerValue: "anthropic-secret",
		},
		{
			name: "azure openai",
			request: Request{
				Provider: "azure-open-ai", AuthMode: "api_key",
				Config: map[string]interface{}{
					"endpoint":    "https://resource.openai.azure.com",
					"api_version": "2024-10-21",
				},
				Secrets: map[string]string{"api_key": "azure-secret"},
			},
			expectedURL: "https://resource.openai.azure.com/openai/models?api-version=2024-10-21",
			headerName:  "api-key", headerValue: "azure-secret",
		},
		{
			name:        "cerebras",
			request:     apiKeyRequest("cerebras", "cerebras-secret"),
			expectedURL: "https://api.cerebras.ai/v1/models",
			headerName:  "Authorization", headerValue: "Bearer cerebras-secret",
		},
		{
			name: "cohere",
			request: Request{
				Provider: "cohere", AuthMode: "api_key",
				Secrets: map[string]string{"api_key": "cohere-secret"},
			},
			expectedURL: "https://api.cohere.com/v1/models",
			headerName:  "Authorization", headerValue: "Bearer cohere-secret",
		},
		{
			name: "databricks",
			request: Request{
				Provider: "databricks", AuthMode: "personal_access_token",
				Config: map[string]interface{}{
					"serving_endpoint_url": "https://workspace.cloud.databricks.com/serving-endpoints/model/invocations",
				},
				Secrets: map[string]string{"access_token": "databricks-secret"},
			},
			expectedURL: "https://workspace.cloud.databricks.com/api/2.0/serving-endpoints",
			headerName:  "Authorization", headerValue: "Bearer databricks-secret",
		},
		{
			name:        "deepinfra",
			request:     apiKeyRequest("deepinfra", "deepinfra-secret"),
			expectedURL: "https://api.deepinfra.com/v1/openai/models",
			headerName:  "Authorization", headerValue: "Bearer deepinfra-secret",
		},
		{
			name: "google gemini",
			request: Request{
				Provider: "google-gemini", AuthMode: "api_key",
				Config: map[string]interface{}{
					"base_url":    "https://generativelanguage.googleapis.com/v1beta",
					"api_version": "v1beta",
				},
				Secrets: map[string]string{"api_key": "gemini-secret"},
			},
			expectedURL: "https://generativelanguage.googleapis.com/v1beta/models?pageSize=1",
			headerName:  "x-goog-api-key", headerValue: "gemini-secret",
		},
		{
			name:        "groq",
			request:     apiKeyRequest("groq", "groq-secret"),
			expectedURL: "https://api.groq.com/openai/v1/models",
			headerName:  "Authorization", headerValue: "Bearer groq-secret",
		},
		{
			name:        "mistral canonical slug",
			request:     apiKeyRequest("mistral-ai", "mistral-secret"),
			expectedURL: "https://api.mistral.ai/v1/models",
			headerName:  "Authorization", headerValue: "Bearer mistral-secret",
		},
		{
			name:        "openai",
			request:     apiKeyRequest("openai", "openai-secret"),
			expectedURL: "https://api.openai.com/v1/models",
			headerName:  "Authorization", headerValue: "Bearer openai-secret",
		},
		{
			name:        "openrouter authenticated key metadata",
			request:     apiKeyRequest("openrouter", "openrouter-secret"),
			expectedURL: "https://openrouter.ai/api/v1/key",
			headerName:  "Authorization", headerValue: "Bearer openrouter-secret",
		},
		{
			name:        "together canonical slug and domain",
			request:     apiKeyRequest("together-ai", "together-secret"),
			expectedURL: "https://api.together.xyz/v1/models",
			headerName:  "Authorization", headerValue: "Bearer together-secret",
		},
		{
			name:        "xai",
			request:     apiKeyRequest("xai", "xai-secret"),
			expectedURL: "https://api.x.ai/v1/models",
			headerName:  "Authorization", headerValue: "Bearer xai-secret",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var captured *http.Request
			client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				captured = req

				return testHTTPResponse(http.StatusOK, `{}`), nil
			})}

			result := NewProber(client).Probe(context.Background(), test.request)
			if !result.Success || result.Category != "ok" {
				t.Fatalf("probe failed: result=%#v", result)
			}
			if captured == nil {
				t.Fatal("expected a provider request")
			}
			if captured.Method != http.MethodGet {
				t.Fatalf("unexpected method %q", captured.Method)
			}
			if captured.URL.String() != test.expectedURL {
				t.Fatalf("unexpected URL %q, want %q", captured.URL.String(), test.expectedURL)
			}
			if got := captured.Header.Get(test.headerName); got != test.headerValue {
				t.Fatalf("unexpected %s header %q", test.headerName, got)
			}
		})
	}
}

func TestProviderAliasesResolveToCanonicalContracts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		provider string
		want     string
	}{
		{provider: "gemini", want: "google-gemini"},
		{provider: "mistral", want: "mistral-ai"},
		{provider: "togetherai", want: "together-ai"},
		{provider: "azure", want: "azure-open-ai"},
	}

	for _, test := range tests {
		if got := NormalizeProvider(test.provider); got != test.want {
			t.Fatalf("NormalizeProvider(%q)=%q, want %q", test.provider, got, test.want)
		}
		if !Supports(test.provider, "api_key") {
			t.Fatalf("expected alias %q to resolve to a supported contract", test.provider)
		}
	}
}

func TestUnsupportedAndInvalidConfigurationDoNotPerformNetworkIO(t *testing.T) {
	t.Parallel()

	calls := 0
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		calls++

		return testHTTPResponse(http.StatusOK, `{}`), nil
	})}
	prober := NewProber(client)

	unsupported := prober.Probe(context.Background(), Request{
		Provider: "deepgram", AuthMode: "api_key", Secrets: map[string]string{"api_key": "secret"},
	})
	if !unsupported.Unsupported || unsupported.Category != "unsupported" {
		t.Fatalf("unexpected unsupported result: %#v", unsupported)
	}

	invalid := prober.Probe(context.Background(), Request{
		Provider: "openai", AuthMode: "api_key",
		Config:  map[string]interface{}{"base_url": "https://127.0.0.1/v1"},
		Secrets: map[string]string{"api_key": "secret"},
	})
	if invalid.Success || invalid.Category != "configuration" {
		t.Fatalf("unexpected invalid-configuration result: %#v", invalid)
	}
	if calls != 0 {
		t.Fatalf("expected no network calls, got %d", calls)
	}
}

func TestProbeClassifiesQuotaAsAuthenticatedAndRedactsErrors(t *testing.T) {
	t.Parallel()

	status := http.StatusTooManyRequests
	body := `{"error":{"message":"quota for sk-secret-value-1234"}}`
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return testHTTPResponse(status, body), nil
	})}
	prober := NewProber(client)
	request := apiKeyRequest("openai", "sk-secret-value-1234")

	quota := prober.Probe(context.Background(), request)
	if !quota.Success || quota.Category != "quota_exhausted" {
		t.Fatalf("unexpected quota result: %#v", quota)
	}
	if strings.Contains(quota.Message, "sk-secret-value-1234") {
		t.Fatalf("result leaked the credential: %q", quota.Message)
	}

	status = http.StatusUnauthorized
	invalid := prober.Probe(context.Background(), request)
	if invalid.Success || invalid.Category != "invalid_credential" {
		t.Fatalf("unexpected invalid result: %#v", invalid)
	}
	if strings.Contains(invalid.Message, "sk-secret-value-1234") {
		t.Fatalf("result leaked the credential: %q", invalid.Message)
	}
}

func TestProbeDoesNotFollowProviderRedirects(t *testing.T) {
	t.Parallel()

	calls := 0
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		calls++

		response := testHTTPResponse(http.StatusFound, `{}`)
		response.Header.Set("Location", "http://127.0.0.1/internal")

		return response, nil
	})}

	result := NewProber(client).Probe(
		context.Background(),
		apiKeyRequest("openai", "openai-secret"),
	)
	if result.Success || result.Category != "provider" {
		t.Fatalf("unexpected redirect result: %#v", result)
	}
	if calls != 1 {
		t.Fatalf("provider redirect caused %d requests; want exactly 1", calls)
	}
}

func apiKeyRequest(provider, secret string) Request {
	return Request{
		Provider: provider,
		AuthMode: "api_key",
		Secrets:  map[string]string{"api_key": secret},
	}
}

func testHTTPResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}
