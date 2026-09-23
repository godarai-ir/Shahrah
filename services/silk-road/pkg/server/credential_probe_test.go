package server

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

type credentialProbeRoundTripper struct {
	request *http.Request
}

func (t *credentialProbeRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	t.request = request

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(`{"models":[]}`)),
		Request:    request,
	}, nil
}

func TestDispatchCredentialProbeBuildsProviderRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		provider    string
		wantURL     string
		apiKey      string
		headerName  string
		headerValue string
	}{
		{
			name:        "Gemini",
			provider:    "google-gemini",
			wantURL:     "https://generativelanguage.googleapis.com/v1beta/models?pageSize=1",
			apiKey:      "test-gemini-key",
			headerName:  "x-goog-api-key",
			headerValue: "test-gemini-key",
		},
		{
			name:        "Moonshot",
			provider:    "moonshot",
			wantURL:     "https://api.moonshot.ai/v1/models",
			apiKey:      "test-moonshot-key",
			headerName:  "Authorization",
			headerValue: "Bearer test-moonshot-key",
		},
		{
			name:        "DeepSeek",
			provider:    "deepseek",
			wantURL:     "https://api.deepseek.com/models",
			apiKey:      "test-deepseek-key",
			headerName:  "Authorization",
			headerValue: "Bearer test-deepseek-key",
		},
		{
			name:        "xAI key status",
			provider:    "xai",
			wantURL:     "https://api.x.ai/v1/api-key",
			apiKey:      "test-xai-key",
			headerName:  "Authorization",
			headerValue: "Bearer test-xai-key",
		},
		{
			name:        "TokenBazaar",
			provider:    "tokenbazaar",
			wantURL:     "https://zx1.deepwl.net/v1/models",
			apiKey:      "test-tokenbazaar-key",
			headerName:  "Authorization",
			headerValue: "Bearer test-tokenbazaar-key",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			transport := &credentialProbeRoundTripper{}

			response, err := dispatchCredentialProbe(
				context.Background(),
				&http.Client{Transport: transport},
				platform.TunnelRequest{
					Operation: platform.ProviderOperationCredentialProbe,
					Deployment: platform.Deployment{
						Provider: test.provider,
						ProviderAuth: &platform.ResolvedProviderAuth{
							Mode: string(platform.ProviderAuthModeAPIKey), APIKey: test.apiKey,
						},
					},
				},
			)
			if err != nil {
				t.Fatalf("dispatchCredentialProbe: %v", err)
			}

			t.Cleanup(func() {
				if closeErr := response.Body.Close(); closeErr != nil {
					t.Errorf("close response body: %v", closeErr)
				}
			})

			if response.StatusCode != http.StatusOK {
				t.Fatalf("status = %d", response.StatusCode)
			}

			if transport.request == nil {
				t.Fatal("credential probe request was not sent")
			}

			if got := transport.request.Method; got != http.MethodGet {
				t.Fatalf("method = %q", got)
			}

			if got := transport.request.URL.String(); got != test.wantURL {
				t.Fatalf("URL = %q", got)
			}

			if got := transport.request.Header.Get(test.headerName); got != test.headerValue {
				t.Fatalf("%s = %q", test.headerName, got)
			}
		})
	}
}

func TestResolveCredentialProbeRejectsUnsafeCustomURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		provider string
		apiBase  string
	}{
		{name: "SambaNova IP", provider: "sambanova", apiBase: "https://127.0.0.1/openai/v1"},
		{name: "Moonshot lookalike", provider: "moonshot", apiBase: "https://api.moonshot.ai.evil.example/v1"},
		{name: "TokenBazaar lookalike", provider: "tokenbazaar", apiBase: "https://zx1.deepwl.net.evil.example"},
		{name: "TokenBazaar unconfirmed alias", provider: "tokenbazaar", apiBase: "https://api.tokenbazaar.ai"},
		{name: "DeepSeek IP", provider: "deepseek", apiBase: "https://127.0.0.1/v1"},
		{name: "DeepSeek localhost", provider: "deepseek", apiBase: "https://localhost/v1"},
		{name: "Z.ai IP", provider: "zai", apiBase: "https://127.0.0.1/api/paas/v4"},
		{name: "Z.ai localhost", provider: "zai", apiBase: "https://localhost/api/paas/v4"},
		{name: "unknown provider", provider: "unknown", apiBase: "https://api.example.com/v1"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := resolveCredentialProbe(platform.Deployment{
				Provider: test.provider,
				APIBase:  test.apiBase,
			})
			if err == nil {
				t.Fatal("expected unsafe or unsupported credential probe to be rejected")
			}
		})
	}
}

func TestResolveCredentialProbeAllowsBYOKGatewayHosts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		provider string
		apiBase  string
	}{
		{name: "DeepSeek lookalike", provider: "deepseek", apiBase: "https://api.deepseek.com.evil.example"},
		{name: "DeepSeek gateway", provider: "deepseek", apiBase: "https://my-gateway.example.com/deepseek/v1"},
		{name: "Z.ai lookalike", provider: "zai", apiBase: "https://api.z.ai.evil.example/api/paas/v4"},
		{name: "Z.ai wrong path", provider: "zai", apiBase: "https://api.z.ai/coding/paas/v4"},
		{name: "Z.ai gateway", provider: "zai", apiBase: "https://my-gateway.example.com/zai/v4"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if _, err := resolveCredentialProbe(platform.Deployment{
				Provider: test.provider,
				APIBase:  test.apiBase,
			}); err != nil {
				t.Fatalf("expected BYOK gateway host to be accepted: %v", err)
			}
		})
	}
}

func TestZAICredentialProbeIsExplicitBillablePOST(t *testing.T) {
	transport := &credentialProbeRoundTripper{}

	response, err := dispatchCredentialProbe(
		context.Background(),
		&http.Client{Transport: transport},
		platform.TunnelRequest{Deployment: platform.Deployment{
			Provider: "zai",
			ProviderAuth: &platform.ResolvedProviderAuth{
				Mode: string(platform.ProviderAuthModeAPIKey), APIKey: "test-zai-key",
			},
		}},
	)
	if err != nil {
		t.Fatalf("dispatch Z.ai credential probe: %v", err)
	}

	t.Cleanup(func() {
		if closeErr := response.Body.Close(); closeErr != nil {
			t.Errorf("close response body: %v", closeErr)
		}
	})

	if transport.request.Method != http.MethodPost ||
		transport.request.URL.String() != "https://api.z.ai/api/paas/v4/chat/completions" {
		t.Fatalf("unexpected Z.ai probe target: %s %s", transport.request.Method, transport.request.URL)
	}

	body, readErr := io.ReadAll(transport.request.Body)
	if readErr != nil {
		t.Fatalf("read Z.ai probe body: %v", readErr)
	}

	for _, expected := range []string{`"model":"glm-4.7-flashx"`, `"max_tokens":1`, `"do_sample":false`} {
		if !strings.Contains(string(body), expected) {
			t.Fatalf("probe body does not contain %s: %s", expected, body)
		}
	}
}
