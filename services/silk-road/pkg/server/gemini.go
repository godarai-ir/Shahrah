package server

import "strings"

const geminiOpenAIBaseURL = "https://generativelanguage.googleapis.com/v1beta/openai"

// newGeminiAdapter uses Google's documented OpenAI compatibility surface. This
// preserves the gateway's OpenAI contract and lets the shared adapter carry
// streaming, tools, structured output, reasoning, embeddings, and image
// generation without maintaining a lossy provider-specific translation.
func newGeminiAdapter() providerAdapter {
	adapter := newOpenAICompatibleAdapter(geminiOpenAIBaseURL).(*openAICompatibleAdapter)
	adapter.baseURLNormalizer = normalizeGeminiOpenAIBaseURL
	return adapter
}

func normalizeGeminiOpenAIBaseURL(value string) string {
	base := strings.TrimRight(strings.TrimSpace(value), "/")
	lower := strings.ToLower(base)
	if strings.HasSuffix(lower, "/openai") {
		return base
	}
	if strings.HasSuffix(lower, "/v1") || strings.HasSuffix(lower, "/v1beta") {
		return base + "/openai"
	}
	return base
}
