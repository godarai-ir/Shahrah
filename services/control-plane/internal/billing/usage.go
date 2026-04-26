package billing

import (
	"encoding/json"
	"fmt"
	"strings"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
)

// TokenFieldMapper maps provider-specific token fields to OpenAI standard format
type TokenFieldMapper func(rawUsage map[string]interface{}) *contractsopenai.Usage

// Token Mappers per provider
var tokenMappers = map[string]TokenFieldMapper{
	// OpenAI, Azure, Groq, Mistral, Together AI - use standard format
	"openai":     standardTokenMapper,
	"azure":      standardTokenMapper,
	"groq":       standardTokenMapper,
	"mistral":    standardTokenMapper,
	"togetherai": standardTokenMapper,

	// Anthropic uses input_tokens, output_tokens
	"anthropic": anthropicTokenMapper,

	// Bedrock uses inputTokens, outputTokens
	"bedrock": bedrockTokenMapper,

	// Vertex AI / Gemini - estimate from content length
	"vertexai": vertexTokenMapper,
	"gemini":   vertexTokenMapper,

	// Cohere - uses prompt_tokens, completion_tokens
	"cohere": cohereTokenMapper,
}

// standardTokenMapper handles providers that use OpenAI format
// Expected fields: prompt_tokens, completion_tokens, total_tokens
func standardTokenMapper(rawUsage map[string]interface{}) *contractsopenai.Usage {
	usage := &contractsopenai.Usage{}

	if v, ok := rawUsage["prompt_tokens"].(float64); ok {
		usage.PromptTokens = int(v)
	}
	if v, ok := rawUsage["completion_tokens"].(float64); ok {
		usage.CompletionTokens = int(v)
	}
	if v, ok := rawUsage["total_tokens"].(float64); ok {
		usage.TotalTokens = int(v)
	}

	// Calculate total if not provided
	if usage.TotalTokens == 0 && (usage.PromptTokens > 0 || usage.CompletionTokens > 0) {
		usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	}

	return usage
}

// anthropicTokenMapper handles Anthropic's token format
// Expected fields: input_tokens, output_tokens
func anthropicTokenMapper(rawUsage map[string]interface{}) *contractsopenai.Usage {
	usage := &contractsopenai.Usage{}

	if v, ok := rawUsage["input_tokens"].(float64); ok {
		usage.PromptTokens = int(v)
	}
	if v, ok := rawUsage["output_tokens"].(float64); ok {
		usage.CompletionTokens = int(v)
	}

	// Anthropic also returns usage in tokens (may be at different levels)
	if usage.PromptTokens == 0 {
		if v, ok := rawUsage["usage"].(map[string]interface{}); ok {
			if v2, ok := v["input_tokens"].(float64); ok {
				usage.PromptTokens = int(v2)
			}
			if v2, ok := v["output_tokens"].(float64); ok {
				usage.CompletionTokens = int(v2)
			}
		}
	}

	usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	return usage
}

// bedrockTokenMapper handles AWS Bedrock's token format
// Expected fields: inputTokens, outputTokens
func bedrockTokenMapper(rawUsage map[string]interface{}) *contractsopenai.Usage {
	usage := &contractsopenai.Usage{}

	if v, ok := rawUsage["inputTokens"].(float64); ok {
		usage.PromptTokens = int(v)
	}
	if v, ok := rawUsage["outputTokens"].(float64); ok {
		usage.CompletionTokens = int(v)
	}

	// Try nested usage object
	if usage.PromptTokens == 0 {
		if v, ok := rawUsage["usage"].(map[string]interface{}); ok {
			if v2, ok := v["inputTokens"].(float64); ok {
				usage.PromptTokens = int(v2)
			}
			if v2, ok := v["outputTokens"].(float64); ok {
				usage.CompletionTokens = int(v2)
			}
		}
	}

	usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	return usage
}

// vertexTokenMapper handles Google Vertex AI / Gemini token format
// Expected fields: promptTokenCount, candidatesTokenCount (or estimate)
func vertexTokenMapper(rawUsage map[string]interface{}) *contractsopenai.Usage {
	usage := &contractsopenai.Usage{}

	// Try different field names
	if v, ok := rawUsage["promptTokenCount"].(float64); ok {
		usage.PromptTokens = int(v)
	}
	if v, ok := rawUsage["candidatesTokenCount"].(float64); ok {
		usage.CompletionTokens = int(v)
	}

	// Try nested usageMetadata
	if usage.PromptTokens == 0 {
		if v, ok := rawUsage["usageMetadata"].(map[string]interface{}); ok {
			if v2, ok := v["promptTokenCount"].(float64); ok {
				usage.PromptTokens = int(v2)
			}
			if v2, ok := v["candidatesTokenCount"].(float64); ok {
				usage.CompletionTokens = int(v2)
			}
			if v2, ok := v["totalTokenCount"].(float64); ok {
				usage.TotalTokens = int(v2)
			}
		}
	}

	// Fallback: estimate from content length (rough approximation)
	if usage.PromptTokens == 0 && usage.CompletionTokens == 0 {
		usage.PromptTokens = 100 // placeholder estimate
		usage.CompletionTokens = 50
		usage.TotalTokens = 150
	} else if usage.TotalTokens == 0 {
		usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	}

	return usage
}

// cohereTokenMapper handles Cohere's token format
// Expected fields: prompt_tokens, completion_tokens
func cohereTokenMapper(rawUsage map[string]interface{}) *contractsopenai.Usage {
	usage := &contractsopenai.Usage{}

	if v, ok := rawUsage["prompt_tokens"].(float64); ok {
		usage.PromptTokens = int(v)
	}
	if v, ok := rawUsage["completion_tokens"].(float64); ok {
		usage.CompletionTokens = int(v)
	}
	if v, ok := rawUsage["total_tokens"].(float64); ok {
		usage.TotalTokens = int(v)
	}

	if usage.TotalTokens == 0 {
		usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	}

	return usage
}

// NormalizeUsage converts provider-specific usage to OpenAI standard format
func NormalizeUsage(provider string, rawUsage map[string]interface{}) *contractsopenai.Usage {
	if rawUsage == nil {
		return &contractsopenai.Usage{}
	}

	// Try to find provider-specific mapper
	mapper, ok := tokenMappers[provider]
	if !ok {
		// Try with provider prefix
		for p := range tokenMappers {
			if strings.HasPrefix(provider, p) {
				mapper = tokenMappers[p]
				break
			}
		}
	}

	if mapper == nil {
		// Fall back to standard mapper
		mapper = standardTokenMapper
	}

	return mapper(rawUsage)
}

// NormalizeUsageFromJSON parses JSON usage response and normalizes it
func NormalizeUsageFromJSON(provider string, usageJSON []byte) (*contractsopenai.Usage, error) {
	var rawUsage map[string]interface{}
	if err := json.Unmarshal(usageJSON, &rawUsage); err != nil {
		return nil, fmt.Errorf("failed to parse usage JSON: %w", err)
	}

	return NormalizeUsage(provider, rawUsage), nil
}

// EstimateTokens estimates token count from text (fallback when usage not returned)
func EstimateTokens(text string) int {
	// Rough estimate: ~4 characters per token
	return len(text) / 4
}

// EstimateUsage estimates usage from response content (fallback)
func EstimateUsage(responseContent string) *contractsopenai.Usage {
	return &contractsopenai.Usage{
		PromptTokens:     100, // Estimated
		CompletionTokens: EstimateTokens(responseContent),
		TotalTokens:      100 + EstimateTokens(responseContent),
	}
}
