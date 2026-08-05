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

type rawProviderAdapter interface {
	DispatchRaw(ctx context.Context, client *http.Client, payload platform.TunnelRequest, body io.Reader, contentLength int64) (*adapterResponse, error)
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
			"":                 openAI,
			"openai":           openAI,
			"ai21":             newOpenAICompatibleAdapter("https://api.ai21.com/studio/v1"),
			"azure-ai-foundry": newOpenAICompatibleAdapterWithAuth("", "api-key", false),
			"azure-open-ai":    newAzureOpenAIAdapter(),
			"azure":            newAzureOpenAIAdapter(),
			"cerebras":         newOpenAICompatibleAdapter("https://api.cerebras.ai/v1"),
			"databricks":       newOpenAICompatibleAdapter(""),
			"groq":             newOpenAICompatibleAdapter("https://api.groq.com/openai/v1"),
			"mistral":          newMistralAdapter(),
			"mistral-ai":       newMistralAdapter(),
			"moonshot":         newMoonshotAdapter(),
			"openrouter":       newOpenAICompatibleAdapter("https://openrouter.ai/api/v1"),
			"perplexity":       newOpenAICompatibleAdapter("https://api.perplexity.ai"),
			"perplexityai":     newOpenAICompatibleAdapter("https://api.perplexity.ai"),
			"perplexity-ai":    newOpenAICompatibleAdapter("https://api.perplexity.ai"),
			"sambanova":        newOpenAICompatibleAdapter(""),
			"deepinfra":        newOpenAICompatibleAdapter("https://api.deepinfra.com/v1/openai"),
			"deepseek":         newDeepSeekAdapter(),
			"fireworks":        newOpenAICompatibleAdapter(""),
			"togetherai":       newOpenAICompatibleAdapter("https://api.together.xyz/v1"),
			"together-ai":      newOpenAICompatibleAdapter("https://api.together.xyz/v1"),
			"xai":              newXAIAdapter(),
			"cohere":           newCohereAdapter(),
			"deepgram":         newDeepgramAudioTranscriptionsAdapter(),
			"google-vertex":    newGoogleVertexSpeechAdapter(),
			"vertex":           newGoogleVertexSpeechAdapter(),
			"vertex-ai":        newGoogleVertexSpeechAdapter(),
			"vertexai":         newGoogleVertexSpeechAdapter(),
			"anthropic":        newAnthropicAdapter(),
			"gemini":           newGeminiAdapter(),
			"google-gemini":    newGeminiAdapter(),
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
