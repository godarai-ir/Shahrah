package server

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func TestRegistryContainsEveryDatabaseProviderMarkedRuntimeReady(t *testing.T) {
	t.Parallel()
	registry := newProviderRegistry()
	providers := []string{
		"anthropic", "azure-open-ai", "cerebras",
		"cohere", "databricks", "deepinfra", "google-gemini", "groq", "mistral-ai",
		"openai", "openrouter", "together-ai", "xai",
	}
	for _, provider := range providers {
		provider := provider
		t.Run(provider, func(t *testing.T) {
			t.Parallel()
			if _, err := registry.adapterFor(provider); err != nil {
				t.Fatalf("runtime-ready provider %q has no adapter: %v", provider, err)
			}
		})
	}
}

func TestVideoAdapterSignsAWSRequestsWithSigV4(t *testing.T) {
	t.Parallel()
	var authorization, sessionToken, contentHash string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization = r.Header.Get("Authorization")
		sessionToken = r.Header.Get("X-Amz-Security-Token")
		contentHash = r.Header.Get("X-Amz-Content-Sha256")
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{}`)
	}))
	defer upstream.Close()

	adapter := newVideoOnlyAdapter(upstream.URL, "bearer")
	response, err := adapter.Dispatch(context.Background(), upstream.Client(), platform.TunnelRequest{
		RequestID: "video-1",
		Deployment: platform.Deployment{
			Provider: "aws-bedrock", APIBase: upstream.URL, RegionName: "us-west-2",
			ProviderAuth: &platform.ResolvedProviderAuth{
				Mode: string(platform.ProviderAuthModeAWSSigV4), AWSAccessKeyID: "AKID",
				AWSSecretAccessKey: "secret", AWSSessionToken: "session",
			},
		},
		Body: []byte(`{"modelId":"test"}`),
		RequestMetadata: map[string]interface{}{
			"provider_operation": "video", "provider_endpoint_path": "/model/test/async-invoke",
		},
	})
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	defer response.Body.Close()
	if !strings.HasPrefix(authorization, "AWS4-HMAC-SHA256 Credential=AKID/") {
		t.Fatalf("unexpected AWS Authorization header %q", authorization)
	}
	if sessionToken != "session" || len(contentHash) != 64 {
		t.Fatalf("missing SigV4 headers: session=%q hash=%q", sessionToken, contentHash)
	}
}

func TestVideoAdapterUsesAndCachesGoogleServiceAccountToken(t *testing.T) {
	t.Parallel()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	pkcs8, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatal(err)
	}
	privatePEM := string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8}))

	var tokenCalls atomic.Int32
	var apiCalls atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			tokenCalls.Add(1)
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"access_token":"google-token","expires_in":3600}`)
			return
		}
		apiCalls.Add(1)
		if got := r.Header.Get("Authorization"); got != "Bearer google-token" {
			t.Errorf("unexpected Google Authorization header %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{}`)
	}))
	defer upstream.Close()

	adapter := newVideoOnlyAdapter(upstream.URL, "bearer")
	payload := platform.TunnelRequest{
		Deployment: platform.Deployment{
			Provider: "google-vertex", APIBase: upstream.URL, RegionName: "us-central1",
			ProviderAuth: &platform.ResolvedProviderAuth{
				Mode: string(platform.ProviderAuthModeGoogleServiceAccount), GoogleClientEmail: "svc@example.test",
				GooglePrivateKey: privatePEM, GoogleTokenURI: upstream.URL + "/token",
			},
		},
		Body: []byte(`{}`),
		RequestMetadata: map[string]interface{}{
			"provider_operation": "video", "provider_endpoint_path": "/v1/video",
		},
	}
	for range 2 {
		response, dispatchErr := adapter.Dispatch(context.Background(), upstream.Client(), payload)
		if dispatchErr != nil {
			t.Fatalf("dispatch: %v", dispatchErr)
		}
		_ = response.Body.Close()
	}
	if tokenCalls.Load() != 1 || apiCalls.Load() != 2 {
		t.Fatalf("expected one token request and two API requests, got token=%d api=%d", tokenCalls.Load(), apiCalls.Load())
	}
}

func TestAzureOpenAIUsesDocumentedAPIKeyHeaderAndDeploymentRoute(t *testing.T) {
	t.Parallel()
	var gotHeader, gotPath, gotVersion string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("api-key")
		gotPath = r.URL.Path
		gotVersion = r.URL.Query().Get("api-version")
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{}`)
	}))
	defer upstream.Close()

	adapter := newAzureOpenAIAdapter()
	response, err := adapter.Dispatch(context.Background(), upstream.Client(), platform.TunnelRequest{
		Deployment: platform.Deployment{
			Provider:   "azure-open-ai",
			Model:      "gpt-4o-production",
			APIBase:    upstream.URL,
			APIVersion: "2024-10-21",
			ProviderAuth: &platform.ResolvedProviderAuth{
				Mode: "api_key", APIKey: "azure-secret",
			},
		},
		Body: []byte(`{"model":"gpt-4o-production","messages":[]}`),
	})
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	defer response.Body.Close()
	if gotHeader != "azure-secret" {
		t.Fatalf("unexpected api-key header %q", gotHeader)
	}
	if gotPath != "/openai/deployments/gpt-4o-production/chat/completions" {
		t.Fatalf("unexpected Azure deployment path %q", gotPath)
	}
	if gotVersion != "2024-10-21" {
		t.Fatalf("unexpected api-version %q", gotVersion)
	}
}

func TestDocumentedOptionalHeadersAreProviderScoped(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodPost, "https://example.com", nil)
	applyDocumentedProviderHeaders(req, platform.Deployment{
		Provider: "openrouter",
		ProviderHeaders: map[string]string{
			"HTTP-Referer":       "https://godarai.example",
			"X-OpenRouter-Title": "GodarAI",
			"OpenAI-Project":     "must-not-cross-provider-boundary",
		},
	})
	if req.Header.Get("HTTP-Referer") != "https://godarai.example" || req.Header.Get("X-OpenRouter-Title") != "GodarAI" {
		t.Fatal("OpenRouter documented headers were not applied")
	}
	if req.Header.Get("OpenAI-Project") != "" {
		t.Fatal("OpenAI-only header crossed the provider boundary")
	}
}
