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
