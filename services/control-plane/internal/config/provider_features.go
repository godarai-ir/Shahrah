package config

import (
	"os"
	"strconv"
	"strings"
)

// ProviderFeatureSettings is the immutable snapshot for provider endpoint rollout flags.
type ProviderFeatureSettings struct {
	MistralEnabled               bool
	MistralEndpoints             map[string]bool
	MistralFilesEnabled          bool
	MistralBatchEnabled          bool
	MistralAgentsEnabled         bool
	MistralConversationsEnabled  bool
	MistralConversationStreaming bool
	AnthropicNativeEnabled       bool
	AnthropicFilesEnabled        bool
	AnthropicBatchesEnabled      bool
	AnthropicAccountIDs          []string
	ZAIEnabled                   bool
	ZAILayoutParsingEnabled      bool
	// TokenBazaarEnabled is a kill switch, independent of every other
	// provider's flag, and defaults to false (not production-ready; see
	// plans/tokenbazaar-production-readiness.md).
	TokenBazaarEnabled bool
	// TokenBazaarImagesEnabled is TokenBazaar's own image-dispatch kill
	// switch (POST /v1/images/generations, POST /v1/images/edits),
	// independent of TokenBazaarEnabled (chat) so the two capabilities can be
	// canaried and rolled back independently, per the "one model's readiness
	// never unlocks another" rule in
	// plans/tokenbazaar-production-readiness.md. Defaults to false.
	TokenBazaarImagesEnabled bool
	// TokenBazaarEmbeddingsEnabled is TokenBazaar's own embeddings-dispatch
	// kill switch (POST /v1/embeddings), independent of both
	// TokenBazaarEnabled (chat) and TokenBazaarImagesEnabled (images), per
	// the same "one model's readiness never unlocks another" rule. Defaults
	// to false; the embeddings contract itself is inferred, not
	// first-party-confirmed, so this must stay off until a live probe
	// confirms it (see plans/tokenbazaar-production-readiness.md release
	// gates).
	TokenBazaarEmbeddingsEnabled bool
	// TokenBazaarResponsesEnabled is TokenBazaar's own Responses-API-dispatch
	// kill switch (POST /v1/responses, create-only -- TokenBazaar's own docs
	// state previous_response_id, built-in tools, and background: true are
	// not supported), independent of TokenBazaarEnabled (chat),
	// TokenBazaarImagesEnabled (images), and TokenBazaarEmbeddingsEnabled
	// (embeddings), per the same "one model's readiness never unlocks
	// another" rule. Defaults to false; streaming for this route is a lower
	// evidence tier than chat_completions (see
	// plans/tokenbazaar-production-readiness.md release gates).
	TokenBazaarResponsesEnabled bool
}

// LoadProviderFeatureSettings reads provider rollout flags at the configuration edge.
func LoadProviderFeatureSettings() ProviderFeatureSettings {
	endpoints := make(map[string]bool)

	for _, endpoint := range []string{
		"audio_speech",
		"audio_transcriptions",
		"chat_completions",
		"embeddings",
		"fim_completions",
		"moderations",
		"ocr",
	} {
		key := "SHAHRAH_MISTRAL_" + strings.ToUpper(endpoint) + "_ENABLED"
		endpoints[endpoint] = strictBooleanEnv(key, true)
	}

	return ProviderFeatureSettings{
		MistralEnabled:               strictBooleanEnv("SHAHRAH_MISTRAL_ENABLED", true),
		MistralEndpoints:             endpoints,
		MistralFilesEnabled:          strictBooleanEnv("SHAHRAH_MISTRAL_FILES_ENABLED", false),
		MistralBatchEnabled:          strictBooleanEnv("SHAHRAH_MISTRAL_BATCH_ENABLED", false),
		MistralAgentsEnabled:         strictBooleanEnv("SHAHRAH_MISTRAL_AGENTS_ENABLED", false),
		MistralConversationsEnabled:  strictBooleanEnv("SHAHRAH_MISTRAL_CONVERSATIONS_ENABLED", false),
		MistralConversationStreaming: strictBooleanEnv("SHAHRAH_MISTRAL_CONVERSATIONS_STREAMING_ENABLED", false),
		AnthropicNativeEnabled:       strictBooleanEnv("ANTHROPIC_NATIVE_API_ENABLED", true),
		AnthropicFilesEnabled:        strictBooleanEnv("ANTHROPIC_FILES_API_ENABLED", true),
		AnthropicBatchesEnabled:      strictBooleanEnv("ANTHROPIC_BATCH_API_ENABLED", true),
		AnthropicAccountIDs:          csvEnv("ANTHROPIC_NATIVE_ACCOUNT_ALLOWLIST"),
		ZAIEnabled:                   strictBooleanEnv("SHAHRAH_ZAI_ENABLED", true),
		ZAILayoutParsingEnabled:      strictBooleanEnv("SHAHRAH_ZAI_LAYOUT_PARSING_ENABLED", true),
		TokenBazaarEnabled:           strictBooleanEnv("SHAHRAH_TOKENBAZAAR_ENABLED", false),
		TokenBazaarImagesEnabled:     strictBooleanEnv("SHAHRAH_TOKENBAZAAR_IMAGES_ENABLED", false),
		TokenBazaarEmbeddingsEnabled: strictBooleanEnv("SHAHRAH_TOKENBAZAAR_EMBEDDINGS_ENABLED", false),
		TokenBazaarResponsesEnabled:  strictBooleanEnv("SHAHRAH_TOKENBAZAAR_RESPONSES_ENABLED", false),
	}
}

func strictBooleanEnv(key string, fallback bool) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false
	}

	return value
}
