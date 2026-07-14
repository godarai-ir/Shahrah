// Package billing handles cost estimation and budget enforcement.
package billing

import (
	"fmt"
	"strings"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
)

// ModelPricing defines pricing for a specific model.
type ModelPricing struct {
	Provider    string  // provider name (openai, anthropic, etc.)
	Model       string  // model name
	InputPrice  float64 // price per 1M input tokens (in USD)
	OutputPrice float64 // price per 1M output tokens (in USD)
	Currency    string  // pricing currency (USD)
	LastUpdated string  // last updated date
}

// PricingRegistry contains all model pricing data.
var PricingRegistry = map[string]ModelPricing{
	// ========== OPENAI ==========
	// GPT-4o Family
	"openai/gpt-4o": {
		Provider: "openai", Model: "gpt-4o",
		InputPrice: 5.00, OutputPrice: 15.00, Currency: "USD",
	},
	"openai/gpt-4o-mini": {
		Provider: "openai", Model: "gpt-4o-mini",
		InputPrice: 0.15, OutputPrice: 0.60, Currency: "USD",
	},
	// GPT-4 Turbo
	"openai/gpt-4-turbo": {
		Provider: "openai", Model: "gpt-4-turbo",
		InputPrice: 10.00, OutputPrice: 30.00, Currency: "USD",
	},
	// GPT-4
	"openai/gpt-4": {
		Provider: "openai", Model: "gpt-4",
		InputPrice: 30.00, OutputPrice: 60.00, Currency: "USD",
	},
	// GPT-3.5 Turbo
	"openai/gpt-3.5-turbo": {
		Provider: "openai", Model: "gpt-3.5-turbo",
		InputPrice: 0.50, OutputPrice: 1.50, Currency: "USD",
	},
	"openai/text-embedding-3-small": {
		Provider: "openai", Model: "text-embedding-3-small",
		InputPrice: 0.02, OutputPrice: 0.00, Currency: "USD",
	},
	"openai/text-embedding-3-large": {
		Provider: "openai", Model: "text-embedding-3-large",
		InputPrice: 0.13, OutputPrice: 0.00, Currency: "USD",
	},
	// GPT Image family. InputPrice is text input; OutputPrice is image output.
	"openai/gpt-image-2": {
		Provider: "openai", Model: "gpt-image-2",
		InputPrice: 5.00, OutputPrice: 30.00, Currency: "USD",
	},
	"openai/gpt-image-1.5": {
		Provider: "openai", Model: "gpt-image-1.5",
		InputPrice: 5.00, OutputPrice: 32.00, Currency: "USD",
	},
	"openai/gpt-image-1-mini": {
		Provider: "openai", Model: "gpt-image-1-mini",
		InputPrice: 2.00, OutputPrice: 8.00, Currency: "USD",
	},
	"openai/gpt-image-1": {
		Provider: "openai", Model: "gpt-image-1",
		InputPrice: 5.00, OutputPrice: 40.00, Currency: "USD",
	},

	// ========== ANTHROPIC ==========
	// Claude 3.5 Family
	"anthropic/claude-3.5-sonnet-20241022": {
		Provider: "anthropic", Model: "claude-3.5-sonnet-20241022",
		InputPrice: 3.00, OutputPrice: 15.00, Currency: "USD",
	},
	"anthropic/claude-3.5-sonnet-20240620": {
		Provider: "anthropic", Model: "claude-3.5-sonnet-20240620",
		InputPrice: 3.00, OutputPrice: 15.00, Currency: "USD",
	},
	// Claude 3 Family
	"anthropic/claude-3-opus-20240229": {
		Provider: "anthropic", Model: "claude-3-opus-20240229",
		InputPrice: 15.00, OutputPrice: 75.00, Currency: "USD",
	},
	"anthropic/claude-3-sonnet-20240229": {
		Provider: "anthropic", Model: "claude-3-sonnet-20240229",
		InputPrice: 3.00, OutputPrice: 15.00, Currency: "USD",
	},
	"anthropic/claude-3-haiku-20240307": {
		Provider: "anthropic", Model: "claude-3-haiku-20240307",
		InputPrice: 0.25, OutputPrice: 1.25, Currency: "USD",
	},

	// ========== GROQ ==========
	// Llama family
	"groq/llama-3.1-70b-versatile": {
		Provider: "groq", Model: "llama-3.1-70b-versatile",
		InputPrice: 0.59, OutputPrice: 0.79, Currency: "USD",
	},
	"groq/llama-3.1-405b-reasoning": {
		Provider: "groq", Model: "llama-3.1-405b-reasoning",
		InputPrice: 3.50, OutputPrice: 3.50, Currency: "USD",
	},
	"groq/llama-3.1-8b-instant": {
		Provider: "groq", Model: "llama-3.1-8b-instant",
		InputPrice: 0.05, OutputPrice: 0.08, Currency: "USD",
	},
	"groq/llama3-70b-8192": {
		Provider: "groq", Model: "llama3-70b-8192",
		InputPrice: 0.59, OutputPrice: 0.79, Currency: "USD",
	},
	"groq/llama3-8b-8192": {
		Provider: "groq", Model: "llama3-8b-8192",
		InputPrice: 0.05, OutputPrice: 0.08, Currency: "USD",
	},
	// Mixtral
	"groq/mixtral-8x7b-32768": {
		Provider: "groq", Model: "mixtral-8x7b-32768",
		InputPrice: 0.24, OutputPrice: 0.24, Currency: "USD",
	},

	// ========== MISTRAL ==========
	"mistral/mistral-large-latest": {
		Provider: "mistral", Model: "mistral-large-latest",
		InputPrice: 2.00, OutputPrice: 6.00, Currency: "USD",
	},
	"mistral/mistral-small-latest": {
		Provider: "mistral", Model: "mistral-small-latest",
		InputPrice: 0.20, OutputPrice: 0.60, Currency: "USD",
	},
	"mistral/mistral-tiny": {
		Provider: "mistral", Model: "mistral-tiny",
		InputPrice: 0.14, OutputPrice: 0.14, Currency: "USD",
	},

	// ========== TOGETHER AI ==========
	// Llama family on Together
	"togetherai/meta-llama/Llama-3.1-70B-Instruct": {
		Provider: "togetherai", Model: "meta-llama/Llama-3.1-70B-Instruct",
		InputPrice: 0.88, OutputPrice: 0.88, Currency: "USD",
	},
	"togetherai/meta-llama/Llama-3.1-8B-Instruct": {
		Provider: "togetherai", Model: "meta-llama/Llama-3.1-8B-Instruct",
		InputPrice: 0.22, OutputPrice: 0.22, Currency: "USD",
	},
	"togetherai/meta-llama/Llama-3-70B-Instruct": {
		Provider: "togetherai", Model: "meta-llama/Llama-3-70B-Instruct",
		InputPrice: 0.90, OutputPrice: 0.90, Currency: "USD",
	},
	// Qwen
	"togetherai/Qwen/Qwen2-72B-Instruct": {
		Provider: "togetherai", Model: "Qwen/Qwen2-72B-Instruct",
		InputPrice: 1.00, OutputPrice: 1.00, Currency: "USD",
	},
	// Yi
	"togetherai/01-ai/Yi-34B": {
		Provider: "togetherai", Model: "01-ai/Yi-34B",
		InputPrice: 0.80, OutputPrice: 0.80, Currency: "USD",
	},

	// ========== GOOGLE GEMINI ==========
	"gemini/gemini-1.5-pro": {
		Provider: "gemini", Model: "gemini-1.5-pro",
		InputPrice: 1.25, OutputPrice: 5.00, Currency: "USD",
	},
	"gemini/gemini-1.5-flash": {
		Provider: "gemini", Model: "gemini-1.5-flash",
		InputPrice: 0.075, OutputPrice: 0.30, Currency: "USD",
	},
	"gemini/gemini-1.5-flash-8b": {
		Provider: "gemini", Model: "gemini-1.5-flash-8b",
		InputPrice: 0.0375, OutputPrice: 0.15, Currency: "USD",
	},
	"gemini/gemini-pro": {
		Provider: "gemini", Model: "gemini-pro",
		InputPrice: 0.50, OutputPrice: 1.50, Currency: "USD",
	},

	// ========== COHERE ==========
	"cohere/command-r-plus": {
		Provider: "cohere", Model: "command-r-plus",
		InputPrice: 3.00, OutputPrice: 15.00, Currency: "USD",
	},
	"cohere/command-r": {
		Provider: "cohere", Model: "command-r",
		InputPrice: 0.50, OutputPrice: 1.50, Currency: "USD",
	},
	"cohere/command": {
		Provider: "cohere", Model: "command",
		InputPrice: 1.00, OutputPrice: 1.00, Currency: "USD",
	},

	// ========== AWS BEDROCK (approximate) ==========
	// Note: Bedrock pricing varies by region and configuration
	"bedrock/anthropic.claude-3.5-sonnet-20241022-v1:0": {
		Provider: "bedrock", Model: "anthropic.claude-3.5-sonnet-20241022-v1:0",
		InputPrice: 3.00, OutputPrice: 15.00, Currency: "USD",
	},
	"bedrock/anthropic.claude-3-opus-20240229-v1:0": {
		Provider: "bedrock", Model: "anthropic.claude-3-opus-20240229-v1:0",
		InputPrice: 15.00, OutputPrice: 75.00, Currency: "USD",
	},
	"bedrock/anthropic.claude-3-sonnet-20240229-v1:0": {
		Provider: "bedrock", Model: "anthropic.claude-3-sonnet-20240229-v1:0",
		InputPrice: 3.00, OutputPrice: 15.00, Currency: "USD",
	},
	"bedrock/anthropic.claude-3-haiku-20240307-v1:0": {
		Provider: "bedrock", Model: "anthropic.claude-3-haiku-20240307-v1:0",
		InputPrice: 0.25, OutputPrice: 1.25, Currency: "USD",
	},
}

// GetPricing retrieves pricing for a provider/model combination.
func GetPricing(provider, model string) (ModelPricing, bool) {
	// Try exact match first
	key := provider + "/" + model
	if pricing, ok := PricingRegistry[key]; ok {
		return pricing, true
	}

	// Try prefix match (provider with model substring)
	lowerProvider := strings.ToLower(provider)

	for k, pricing := range PricingRegistry {
		parts := strings.SplitN(k, "/", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], lowerProvider) {
			// Check if model contains the requested model or vice versa
			if strings.Contains(strings.ToLower(parts[1]), strings.ToLower(model)) ||
				strings.Contains(strings.ToLower(model), strings.ToLower(parts[1])) {
				return pricing, false // found but not exact
			}
		}
	}

	// Return zero pricing as fallback
	return ModelPricing{
		Provider:    provider,
		Model:       model,
		InputPrice:  0,
		OutputPrice: 0,
		Currency:    "USD",
	}, false
}

// CalculateCost calculates the cost for a given usage.
func CalculateCost(provider, model string, promptTokens, completionTokens int) (float64, bool) {
	pricing, found := GetPricing(provider, model)

	inputCost := float64(promptTokens) / 1_000_000 * pricing.InputPrice
	outputCost := float64(completionTokens) / 1_000_000 * pricing.OutputPrice

	return inputCost + outputCost, found
}

// GetAllProviders returns list of all providers in the registry.
func GetAllProviders() []string {
	seen := make(map[string]bool)
	providers := []string{}

	for key := range PricingRegistry {
		parts := strings.SplitN(key, "/", 2)
		if len(parts) == 2 && !seen[parts[0]] {
			seen[parts[0]] = true
			providers = append(providers, parts[0])
		}
	}

	return providers
}

// FormatPrice formats a price for display.
func FormatPrice(price float64) string {
	return fmt.Sprintf("$%.4f", price)
}

// priceInfo holds pricing info with per-token prices.
type priceInfo struct {
	Input  float64 // per 1K token
	Output float64 // per 1K token
}

// modelPriceFromRegistry returns pricing for a given model from the static registry.
func modelPriceFromRegistry(model string) priceInfo {
	// Try to find pricing in registry
	parts := strings.SplitN(model, "/", 2)
	provider := "openai"
	modelName := model

	if len(parts) == 2 {
		provider = parts[0]
		modelName = parts[1]
	}

	pricing, found := GetPricing(provider, modelName)
	if !found {
		// Return default pricing
		return priceInfo{Input: 0.001, Output: 0.002}
	}

	// Convert from per-1M to per-1K
	return priceInfo{
		Input:  pricing.InputPrice / 1000,
		Output: pricing.OutputPrice / 1000,
	}
}

// estimatePromptTokens estimates prompt tokens from request (legacy function for service.go).
func estimatePromptTokens(req interface{}) int {
	switch v := req.(type) {
	case contractsopenai.ChatCompletionRequest:
		textTokens := 0
		imageTokens := 0
		fileTokens := 0

		for _, msg := range v.Messages {
			textTokens += EstimateTokens(contractsopenai.ChatContentText(msg.Content))
			imageTokens += contractsopenai.ChatContentImageCount(msg.Content) * 1024
			fileTokens += contractsopenai.ChatContentFileCount(msg.Content) * 4096
		}

		if textTokens > 0 || imageTokens > 0 || fileTokens > 0 {
			if textTokens < 100 {
				textTokens = 100
			}

			return textTokens + imageTokens + fileTokens
		}
	case *contractsopenai.ChatCompletionRequest:
		if v != nil {
			return estimatePromptTokens(*v)
		}
	}

	return 100
}
