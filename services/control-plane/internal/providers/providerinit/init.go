package providerinit

import (
	// Import provider registry and adapters
	"github.com/Delpak/shahrah/services/control-plane/internal/providers"

	// Import Wave 1 adapters to register them
	_ "github.com/Delpak/shahrah/services/control-plane/internal/providers/azure"
	_ "github.com/Delpak/shahrah/services/control-plane/internal/providers/groq"
	_ "github.com/Delpak/shahrah/services/control-plane/internal/providers/mistral"
	_ "github.com/Delpak/shahrah/services/control-plane/internal/providers/openai"
	_ "github.com/Delpak/shahrah/services/control-plane/internal/providers/togetherai"
)

// Init registers all supported provider adapters
func Init() {
	// Register Wave 1 adapters (they import themselves or are registered by init())
	// This function can be called from main to ensure all providers are registered
	_ = providers.Supported // Reference to ensure package is loaded
}

// SupportedProviders returns all supported provider names
func SupportedProviders() []string {
	return []string{
		// Wave 1
		"openai",
		"azure",
		"groq",
		"mistral",
		"togetherai",
		// Wave 2 (not yet implemented)
		// "anthropic",
		// "bedrock",
		// "vertexai",
		// "gemini",
		// Wave 3
		// "cohere",
	}
}
