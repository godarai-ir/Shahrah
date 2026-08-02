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

func TestDispatchCredentialProbeBuildsGeminiRequest(t *testing.T) {
	t.Parallel()

	transport := &credentialProbeRoundTripper{}
	response, err := dispatchCredentialProbe(context.Background(), &http.Client{Transport: transport}, platform.TunnelRequest{
		Operation: platform.ProviderOperationCredentialProbe,
		Deployment: platform.Deployment{
			Provider: "google-gemini",
			ProviderAuth: &platform.ResolvedProviderAuth{
				Mode: string(platform.ProviderAuthModeAPIKey), APIKey: "gemini-secret",
			},
		},
	})
	if err != nil {
		t.Fatalf("dispatchCredentialProbe: %v", err)
	}
	defer response.Body.Close()

	if got := transport.request.URL.String(); got != "https://generativelanguage.googleapis.com/v1beta/models?pageSize=1" {
		t.Fatalf("URL = %q", got)
	}
	if got := transport.request.Header.Get("x-goog-api-key"); got != "gemini-secret" {
		t.Fatalf("x-goog-api-key = %q", got)
	}
}

func TestResolveCredentialProbeRejectsUnsafeCustomURL(t *testing.T) {
	t.Parallel()

	_, err := resolveCredentialProbe(platform.Deployment{
		Provider: "sambanova",
		APIBase:  "https://127.0.0.1/openai/v1",
	})
	if err == nil {
		t.Fatal("expected unsafe URL to be rejected")
	}
}
