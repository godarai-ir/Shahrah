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

func TestGeminiAdapterCarriesDocumentedOpenAICompatibleEndpointsAndStreaming(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		path     string
		stream   bool
		wantPath string
	}{
		{name: "chat stream", stream: true, wantPath: "/v1beta/openai/chat/completions"},
		{name: "embeddings", path: "/embeddings", wantPath: "/v1beta/openai/embeddings"},
		{name: "images", path: "/images/generations", wantPath: "/v1beta/openai/images/generations"},
		{name: "videos", path: "/videos", wantPath: "/v1beta/openai/videos"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var gotPath, gotAuthorization, gotAccept string
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				gotAuthorization = r.Header.Get("Authorization")
				gotAccept = r.Header.Get("Accept")
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, `{}`)
			}))
			defer upstream.Close()

			metadata := map[string]interface{}{}
			if tt.path != "" {
				metadata["provider_endpoint_path"] = tt.path
			}
			adapter := newGeminiAdapter()
			response, err := adapter.Dispatch(context.Background(), upstream.Client(), platform.TunnelRequest{
				Stream: tt.stream,
				Deployment: platform.Deployment{
					Provider: "google-gemini",
					APIBase:  upstream.URL + "/v1beta",
					APIKey:   "gemini-secret",
				},
				Body:            []byte(`{"model":"gemini-3.1-flash-lite"}`),
				RequestMetadata: metadata,
			})
			if err != nil {
				t.Fatalf("dispatch: %v", err)
			}
			defer response.Body.Close()
			if gotPath != tt.wantPath {
				t.Fatalf("path=%q want=%q", gotPath, tt.wantPath)
			}
			if gotAuthorization != "Bearer gemini-secret" {
				t.Fatalf("authorization=%q", gotAuthorization)
			}
			if tt.stream && gotAccept != "text/event-stream" {
				t.Fatalf("stream accept=%q", gotAccept)
			}
		})
	}
}
