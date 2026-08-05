package platformcredentials

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return fn(req) }

func TestDocumentedProviderProbesUseExactAuthentication(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		record     EncryptedCredential
		secrets    map[string]string
		headerName string
		header     string
		urlPart    string
	}{
		{
			name:    "anthropic x-api-key and version",
			record:  EncryptedCredential{Binding: Binding{ProviderSlug: "anthropic", AuthMode: "api_key"}, APIVersion: "2023-06-01"},
			secrets: map[string]string{"api_key": "anthropic-secret"}, headerName: "x-api-key", header: "anthropic-secret", urlPart: "/v1/models?limit=1",
		},
		{
			name:    "gemini x-goog-api-key",
			record:  EncryptedCredential{Binding: Binding{ProviderSlug: "google-gemini", AuthMode: "api_key"}, BaseURL: "https://generativelanguage.googleapis.com/v1beta/openai"},
			secrets: map[string]string{"api_key": "gemini-secret"}, headerName: "x-goog-api-key", header: "gemini-secret", urlPart: "/v1beta/models?pageSize=1",
		},
		{
			name:    "azure api-key",
			record:  EncryptedCredential{Binding: Binding{ProviderSlug: "azure-open-ai", AuthMode: "api_key"}, BaseURL: "https://resource.openai.azure.com", APIVersion: "2024-10-21"},
			secrets: map[string]string{"api_key": "azure-secret"}, headerName: "api-key", header: "azure-secret", urlPart: "/openai/models?api-version=2024-10-21",
		},
		{
			name:    "databricks personal access token",
			record:  EncryptedCredential{Binding: Binding{ProviderSlug: "databricks", AuthMode: "personal_access_token"}, BaseURL: "https://workspace.cloud.databricks.com/serving-endpoints"},
			secrets: map[string]string{"access_token": "databricks-token"}, headerName: "Authorization", header: "Bearer databricks-token", urlPart: "/api/2.0/serving-endpoints",
		},
		{
			name:    "sambanova custom base URL",
			record:  EncryptedCredential{Binding: Binding{ProviderSlug: "sambanova", AuthMode: "api_key"}, BaseURL: "https://sambastack.example.com/openai/v1"},
			secrets: map[string]string{"api_key": "sambanova-secret"}, headerName: "Authorization", header: "Bearer sambanova-secret", urlPart: "/openai/v1/models",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var request *http.Request

			client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				request = req
				return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{}`)), Request: req}, nil
			})}

			ok, category, _ := NewHTTPProber(client).Probe(context.Background(), test.record, test.secrets)
			if !ok || category != "ok" {
				t.Fatalf("probe failed: ok=%v category=%q", ok, category)
			}

			if request == nil || request.Header.Get(test.headerName) != test.header {
				t.Fatalf("unexpected %s header", test.headerName)
			}

			if !strings.Contains(request.URL.String(), test.urlPart) {
				t.Fatalf("unexpected probe URL %q", request.URL.String())
			}
		})
	}
}

func TestOpenAIProbePreservesConnectionMetadataHeaders(t *testing.T) {
	t.Parallel()

	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Header.Get("OpenAI-Organization") != "org-test" || req.Header.Get("OpenAI-Project") != "project-test" {
			t.Fatalf("OpenAI connection metadata headers were not forwarded: %#v", req.Header)
		}

		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{}`)), Request: req}, nil
	})}

	ok, category, _ := NewHTTPProber(client).Probe(context.Background(), EncryptedCredential{
		Binding:      Binding{ProviderSlug: "openai", AuthMode: "api_key"},
		Organization: "org-test",
		Project:      "project-test",
	}, map[string]string{"api_key": "openai-secret"})
	if !ok || category != "ok" {
		t.Fatalf("probe failed: ok=%v category=%q", ok, category)
	}
}

func TestUnsupportedProviderProbeDoesNotPerformNetworkIO(t *testing.T) {
	t.Parallel()

	calls := 0
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		calls++
		return nil, nil
	})}

	ok, category, _ := NewHTTPProber(client).Probe(context.Background(), EncryptedCredential{
		Binding: Binding{ProviderSlug: "deepgram", AuthMode: "api_key"},
	}, map[string]string{"api_key": "secret"})
	if ok || category != "unsupported" || calls != 0 {
		t.Fatalf("unsupported probe must fail without network I/O: ok=%v category=%q calls=%d", ok, category, calls)
	}
}

func TestHTTPProberClassifiesProviderResponses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		status   int
		success  bool
		category string
	}{
		{name: "accepted", status: http.StatusOK, success: true, category: "ok"},
		{name: "invalid credential", status: http.StatusUnauthorized, category: "invalid_credential"},
		{name: "forbidden", status: http.StatusForbidden, category: "forbidden"},
		{name: "quota", status: http.StatusTooManyRequests, success: true, category: "quota_exhausted"},
		{name: "billing", status: http.StatusPaymentRequired, success: true, category: "quota_exhausted"},
		{name: "not found", status: http.StatusNotFound, category: "deployment_not_found"},
		{name: "provider error", status: http.StatusInternalServerError, category: "provider"},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: test.status, Header: make(http.Header), Body: http.NoBody, Request: req}, nil
			})}

			success, category, _ := NewHTTPProber(client).Probe(context.Background(), EncryptedCredential{
				Binding: Binding{ProviderSlug: "openai", AuthMode: "api_key"},
			}, map[string]string{"api_key": "secret"})
			if success != test.success || category != test.category {
				t.Fatalf("unexpected classification: success=%v category=%q", success, category)
			}
		})
	}
}

func TestHTTPProberRejectsUnsafeBaseURLBeforeNetworkIO(t *testing.T) {
	t.Parallel()

	calls := 0
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		calls++
		return nil, nil
	})}

	ok, category, _ := NewHTTPProber(client).Probe(context.Background(), EncryptedCredential{
		Binding: Binding{ProviderSlug: "openai", AuthMode: "api_key"},
		BaseURL: "https://127.0.0.1/v1",
	}, map[string]string{"api_key": "secret"})
	if ok || category != "configuration" || calls != 0 {
		t.Fatalf("unsafe URL must fail without network I/O: ok=%v category=%q calls=%d", ok, category, calls)
	}
}

func TestDocumentedHTTPProbeCoverageForRuntimeReadyDirectAuthModes(t *testing.T) {
	t.Parallel()

	tests := []EncryptedCredential{
		{Binding: Binding{ProviderSlug: "anthropic", AuthMode: "api_key"}},
		{Binding: Binding{ProviderSlug: "azure-open-ai", AuthMode: "api_key"}, BaseURL: "https://resource.openai.azure.com"},
		{Binding: Binding{ProviderSlug: "cerebras", AuthMode: "api_key"}},
		{Binding: Binding{ProviderSlug: "cohere", AuthMode: "api_key"}},
		{Binding: Binding{ProviderSlug: "databricks", AuthMode: "personal_access_token"}, BaseURL: "https://workspace.cloud.databricks.com/serving-endpoints/example"},
		{Binding: Binding{ProviderSlug: "deepinfra", AuthMode: "api_key"}},
		{Binding: Binding{ProviderSlug: "google-gemini", AuthMode: "api_key"}},
		{Binding: Binding{ProviderSlug: "groq", AuthMode: "api_key"}},
		{Binding: Binding{ProviderSlug: "mistral-ai", AuthMode: "api_key"}},
		{Binding: Binding{ProviderSlug: "moonshot", AuthMode: "api_key"}},
		{Binding: Binding{ProviderSlug: "openai", AuthMode: "api_key"}},
		{Binding: Binding{ProviderSlug: "openrouter", AuthMode: "api_key"}},
		{Binding: Binding{ProviderSlug: "together-ai", AuthMode: "api_key"}},
		{Binding: Binding{ProviderSlug: "xai", AuthMode: "api_key"}},
	}

	for _, record := range tests {
		record := record
		name := record.ProviderSlug + "/" + record.AuthMode
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, ok := documentedProbe(record); !ok {
				t.Fatalf("runtime-ready provider auth mode has no documented probe: %s", name)
			}
		})
	}
}
