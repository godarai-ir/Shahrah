package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
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
