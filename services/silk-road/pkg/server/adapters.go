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
	openAI := newVideoCapableAdapter(
		newOpenAICompatibleAdapter("https://api.openai.com/v1"),
		"https://api.openai.com/v1",
		"bearer",
	)

	registry := &providerRegistry{
		adapters: map[string]providerAdapter{
			"":                 openAI,
			"openai":           openAI,
			"ai21":             newOpenAICompatibleAdapter("https://api.ai21.com/studio/v1"),
			"azure-ai-foundry": newOpenAICompatibleAdapterWithAuth("", "api-key", false),
			"azure-open-ai":    newVideoCapableAdapter(newAzureOpenAIAdapter(), "", "api-key"),
			"azure":            newVideoCapableAdapter(newAzureOpenAIAdapter(), "", "api-key"),
			"cerebras":         newOpenAICompatibleAdapter("https://api.cerebras.ai/v1"),
			"databricks":       newOpenAICompatibleAdapter(""),
			"groq":             newOpenAICompatibleAdapter("https://api.groq.com/openai/v1"),
			"mistral":          newOpenAICompatibleAdapter("https://api.mistral.ai/v1"),
			"mistral-ai":       newOpenAICompatibleAdapter("https://api.mistral.ai/v1"),
			"openrouter":       newOpenAICompatibleAdapter("https://openrouter.ai/api/v1"),
			"perplexity":       newOpenAICompatibleAdapter("https://api.perplexity.ai"),
			"perplexityai":     newOpenAICompatibleAdapter("https://api.perplexity.ai"),
			"perplexity-ai":    newOpenAICompatibleAdapter("https://api.perplexity.ai"),
			"sambanova":        newOpenAICompatibleAdapter(""),
			"deepinfra":        newVideoCapableAdapter(newOpenAICompatibleAdapter("https://api.deepinfra.com/v1/openai"), "https://api.deepinfra.com", "bearer"),
			"fireworks":        newOpenAICompatibleAdapter(""),
			"togetherai":       newVideoCapableAdapter(newOpenAICompatibleAdapter("https://api.together.xyz/v1"), "https://api.together.xyz/v1", "bearer"),
			"together-ai":      newVideoCapableAdapter(newOpenAICompatibleAdapter("https://api.together.xyz/v1"), "https://api.together.xyz/v1", "bearer"),
			"xai":              newVideoCapableAdapter(newOpenAICompatibleAdapter("https://api.x.ai/v1"), "https://api.x.ai/v1", "bearer"),
			"cohere":           newCohereAdapter(),
			"deepgram":         newDeepgramAudioTranscriptionsAdapter(),
			"google-vertex":    newVideoCapableAdapter(newGoogleVertexSpeechAdapter(), "", "bearer"),
			"vertex":           newVideoCapableAdapter(newGoogleVertexSpeechAdapter(), "", "bearer"),
			"vertex-ai":        newVideoCapableAdapter(newGoogleVertexSpeechAdapter(), "", "bearer"),
			"vertexai":         newVideoCapableAdapter(newGoogleVertexSpeechAdapter(), "", "bearer"),
			"anthropic":        newAnthropicAdapter(),
			"gemini":           newVideoCapableAdapter(newGeminiAdapter(), "https://generativelanguage.googleapis.com/v1beta", "google-api-key"),
			"google-gemini":    newVideoCapableAdapter(newGeminiAdapter(), "https://generativelanguage.googleapis.com/v1beta", "google-api-key"),
			"aws-bedrock":      newVideoOnlyAdapter("", "bearer"),
			"bedrock":          newVideoOnlyAdapter("", "bearer"),
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
