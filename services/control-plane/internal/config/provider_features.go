package config

import (
	"os"
	"strconv"
	"strings"
)

// ProviderFeatureSettings is the immutable snapshot for provider endpoint rollout flags.
type ProviderFeatureSettings struct {
	MistralFilesEnabled          bool
	MistralBatchEnabled          bool
	MistralAgentsEnabled         bool
	MistralConversationsEnabled  bool
	MistralConversationStreaming bool
	AnthropicAccountIDs          []string
	VideoCanaryTenantIDs         []string
}

// LoadProviderFeatureSettings reads provider rollout flags at the configuration edge.
func LoadProviderFeatureSettings() ProviderFeatureSettings {
	return ProviderFeatureSettings{
		MistralFilesEnabled:          strictBooleanEnv("SHAHRAH_MISTRAL_FILES_ENABLED", false),
		MistralBatchEnabled:          strictBooleanEnv("SHAHRAH_MISTRAL_BATCH_ENABLED", false),
		MistralAgentsEnabled:         strictBooleanEnv("SHAHRAH_MISTRAL_AGENTS_ENABLED", false),
		MistralConversationsEnabled:  strictBooleanEnv("SHAHRAH_MISTRAL_CONVERSATIONS_ENABLED", false),
		MistralConversationStreaming: strictBooleanEnv("SHAHRAH_MISTRAL_CONVERSATIONS_STREAMING_ENABLED", false),
		AnthropicAccountIDs:          csvEnv("ANTHROPIC_NATIVE_ACCOUNT_ALLOWLIST"),
		VideoCanaryTenantIDs:         csvEnv("SHAHRAH_VIDEOS_CANARY_TENANT_IDS"),
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
