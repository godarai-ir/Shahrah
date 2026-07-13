package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

type providerAdapter interface {
	Dispatch(ctx context.Context, client *http.Client, payload platform.TunnelRequest) (*adapterResponse, error)
}

type adapterResponse struct {
	StatusCode int
	Headers    http.Header
	Body       io.ReadCloser
}

type providerRegistry struct {
	adapters map[string]providerAdapter
}

func newProviderRegistry() *providerRegistry {
	openAI := newOpenAICompatibleAdapter("https://api.openai.com/v1")

	registry := &providerRegistry{
		adapters: map[string]providerAdapter{
			"":             openAI,
			"openai":       openAI,
			"groq":         newOpenAICompatibleAdapter("https://api.groq.com/openai/v1"),
			"mistral":      newOpenAICompatibleAdapter("https://api.mistral.ai/v1"),
			"mistral-ai":   newOpenAICompatibleAdapter("https://api.mistral.ai/v1"),
			"openrouter":   newOpenAICompatibleAdapter(""),
			"perplexity":   newOpenAICompatibleAdapter(""),
			"perplexityai": newOpenAICompatibleAdapter(""),
			"sambanova":    newOpenAICompatibleAdapter(""),
			"togetherai":   newOpenAICompatibleAdapter("https://api.together.ai/v1"),
			"together-ai":  newOpenAICompatibleAdapter("https://api.together.ai/v1"),
			"xai":          newOpenAICompatibleAdapter(""),
			"cohere":       newCohereAdapter(),
			"anthropic":    newAnthropicAdapter(),
			"gemini":       newGeminiAdapter(),
		},
	}
	if mockProviderEnabled() {
		registry.adapters["mock"] = newMockAdapter(loadMockConfigFromEnv())
	}

	return registry
}

func (r *providerRegistry) adapterFor(provider string) (providerAdapter, error) {
	normalized := normalizeProvider(provider)
	adapter, ok := r.adapters[normalized]
	if !ok {
		return nil, fmt.Errorf("unsupported provider %q", provider)
	}
	return adapter, nil
}

func normalizeProvider(provider string) string {
	return strings.ToLower(strings.TrimSpace(provider))
}
