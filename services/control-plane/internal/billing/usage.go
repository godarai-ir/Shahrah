package billing

import (
	"encoding/json"
	"fmt"
	"strings"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
)

// TokenFieldMapper maps provider-specific token fields to OpenAI standard format.
type TokenFieldMapper func(rawUsage map[string]interface{}) *contractsopenai.Usage

// Token Mappers per provider.
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

	// Vertex AI uses its native usage schema.
	"vertexai": vertexTokenMapper,

	// Gemini may return either the native usageMetadata shape or the documented
	// OpenAI-compatible usage shape. The mapper handles both without estimates.
	"gemini":        googleGeminiTokenMapper,
	"google-gemini": googleGeminiTokenMapper,

	// Cohere - uses prompt_tokens, completion_tokens
	"cohere": cohereTokenMapper,
}

func googleGeminiTokenMapper(rawUsage map[string]interface{}) *contractsopenai.Usage {
	if _, compatible := rawUsage["prompt_tokens"]; compatible {
		usage := standardTokenMapper(rawUsage)
		NormalizeGoogleGeminiOpenAIUsage(usage)

		return usage
	}

	metadata := rawUsage
	if nested, ok := rawUsage["usageMetadata"].(map[string]interface{}); ok {
		metadata = nested
	}

	usage := &contractsopenai.Usage{}
	usage.PromptTokens = usageInt(metadata, "promptTokenCount")
	thoughts := usageInt(metadata, "thoughtsTokenCount")
	usage.CompletionTokens = usageInt(metadata, "candidatesTokenCount") + thoughts

	usage.TotalTokens = usageInt(metadata, "totalTokenCount")
	if usage.TotalTokens == 0 && (usage.PromptTokens > 0 || usage.CompletionTokens > 0) {
		usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	}

	cached := usageInt(metadata, "cachedContentTokenCount")
	usage.CacheReadInputTokens = cached
	usage.PromptTokensDetails = &contractsopenai.UsageTokenDetails{
		CachedTokens: cached,
		AudioTokens:  modalityTokenCount(metadata["promptTokensDetails"], "AUDIO"),
		ImageTokens:  modalityTokenCount(metadata["promptTokensDetails"], "IMAGE"),
	}

	usage.CompletionTokensDetails = &contractsopenai.UsageTokenDetails{
		ReasoningTokens: thoughts,
		AudioTokens:     modalityTokenCount(metadata["candidatesTokensDetails"], "AUDIO"),
		ImageTokens:     modalityTokenCount(metadata["candidatesTokensDetails"], "IMAGE"),
	}
	switch strings.ToUpper(strings.TrimSpace(usageString(metadata, "trafficType"))) {
	case "PRIORITY":
		usage.ServiceTier = "priority"
	case "BATCH":
		usage.ServiceTier = "batch"
	case "FLEX":
		usage.ServiceTier = "flex"
	case "ON_DEMAND", "STANDARD":
		usage.ServiceTier = "standard"
	}

	return usage
}

// NormalizeGoogleGeminiOpenAIUsage repairs Gemini's OpenAI-compatible usage
// when hidden thinking tokens are included in total_tokens but omitted from
// completion_tokens. It returns the inferred reasoning-token count and whether
// a repair was applied.
func NormalizeGoogleGeminiOpenAIUsage(usage *contractsopenai.Usage) (int, bool) {
	if usage == nil || usage.PromptTokens < 0 || usage.CompletionTokens < 0 || usage.TotalTokens < 0 {
		return 0, false
	}

	if usage.TotalTokens < usage.PromptTokens {
		return 0, false
	}

	canonicalCompletionTokens := usage.TotalTokens - usage.PromptTokens
	if canonicalCompletionTokens <= usage.CompletionTokens {
		return 0, false
	}

	inferredReasoningTokens := canonicalCompletionTokens - usage.CompletionTokens
	usage.CompletionTokens = canonicalCompletionTokens

	if usage.CompletionTokensDetails == nil {
		usage.CompletionTokensDetails = &contractsopenai.UsageTokenDetails{}
	}

	if usage.CompletionTokensDetails.ReasoningTokens < inferredReasoningTokens {
		usage.CompletionTokensDetails.ReasoningTokens = inferredReasoningTokens
	}

	return inferredReasoningTokens, true
}

func usageInt(values map[string]interface{}, key string) int {
	switch value := values[key].(type) {
	case float64:
		return int(value)
	case int:
		return value
	case json.Number:
		parsed, _ := value.Int64() //nolint:errcheck // Failure is intentionally handled by the surrounding best-effort path.
		return int(parsed)
	default:
		return 0
	}
}

func usageString(values map[string]interface{}, key string) string {
	value, _ := values[key].(string) //nolint:errcheck // Optional metadata uses the zero value when absent or mismatched.
	return value
}

func modalityTokenCount(raw interface{}, modality string) int {
	items, ok := raw.([]interface{})
	if !ok {
		return 0
	}

	total := 0

	for _, item := range items {
		entry, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		if strings.EqualFold(usageString(entry, "modality"), modality) {
			total += usageInt(entry, "tokenCount")
		}
	}

	return total
}

// standardTokenMapper handles providers that use OpenAI format
// Expected fields: prompt_tokens, completion_tokens, total_tokens.
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
// Expected fields: input_tokens, output_tokens.
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
// Expected fields: inputTokens, outputTokens.
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
// Expected fields: promptTokenCount, candidatesTokenCount (or estimate).
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
// Expected fields: prompt_tokens, completion_tokens.
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

// NormalizeUsage converts provider-specific usage to OpenAI standard format.
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

// NormalizeUsageFromJSON parses JSON usage response and normalizes it.
func NormalizeUsageFromJSON(provider string, usageJSON []byte) (*contractsopenai.Usage, error) {
	var rawUsage map[string]interface{}
	if err := json.Unmarshal(usageJSON, &rawUsage); err != nil {
		return nil, fmt.Errorf("failed to parse usage JSON: %w", err)
	}

	return NormalizeUsage(provider, rawUsage), nil
}

// EstimateTokens estimates token count from text (fallback when usage not returned).
func EstimateTokens(text string) int {
	// Rough estimate: ~4 characters per token
	return len(text) / 4
}

// EstimateUsage estimates usage from response content (fallback).
func EstimateUsage(responseContent string) *contractsopenai.Usage {
	return &contractsopenai.Usage{
		PromptTokens:     100, // Estimated
		CompletionTokens: EstimateTokens(responseContent),
		TotalTokens:      100 + EstimateTokens(responseContent),
	}
}

func estimateEmbeddingPromptTokens(input interface{}) int {
	switch v := input.(type) {
	case string:
		return EstimateTokens(v)
	case []string:
		total := 0
		for _, item := range v {
			total += EstimateTokens(item)
		}

		return total
	case []interface{}:
		total := 0

		for _, item := range v {
			switch typed := item.(type) {
			case string:
				total += EstimateTokens(typed)
			case float64:
				// JSON token arrays are decoded as float64 values.
				total++
			case []interface{}:
				for _, token := range typed {
					if _, ok := token.(float64); ok {
						total++
					}
				}
			}
		}

		return total
	default:
		return 0
	}
}
