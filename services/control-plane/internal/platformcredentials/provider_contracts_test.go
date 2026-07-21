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
			record:  EncryptedCredential{Binding: Binding{ProviderSlug: "google-gemini", AuthMode: "api_key"}},
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
