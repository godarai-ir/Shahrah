package proxy

import (
	"encoding/json"
	"fmt"
	"strings"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func runtimeProfilesFromEstimate(
	provider, model string,
	estimated platform.UsageCostBreakdown,
) ([]platform.ModelRuntimeProfileBinding, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider != "anthropic" && !isGoogleGeminiProvider(provider) && !isMistralProvider(provider) {
		return nil, nil
	}
	providerLabel := "Anthropic"
	if isGoogleGeminiProvider(provider) {
		providerLabel = "Google Gemini"
	} else if isMistralProvider(provider) {
		providerLabel = "Mistral AI"
	}
	if estimated.ModelRuntimeProfile == nil {
		return nil, fmt.Errorf("%s model runtime profile is unavailable", providerLabel)
	}
	binding := *estimated.ModelRuntimeProfile
	if err := binding.Validate(); err != nil ||
		binding.CatalogRevision != estimated.PricingRevision ||
		!binding.MatchesModel(model) ||
		!runtimeProfileProviderMatches(provider, binding.Provider) {
		return nil, fmt.Errorf("%s model runtime profile is invalid", providerLabel)
	}
	return []platform.ModelRuntimeProfileBinding{binding}, nil
}

func runtimeProfileProviderMatches(deploymentProvider, bindingProvider string) bool {
	if strings.EqualFold(
		strings.TrimSpace(deploymentProvider),
		strings.TrimSpace(bindingProvider),
	) {
		return true
	}
	return isGoogleGeminiProvider(deploymentProvider) &&
		isGoogleGeminiProvider(bindingProvider) ||
		isMistralProvider(deploymentProvider) && isMistralProvider(bindingProvider)
}

func validateAnthropicChatRuntimeProfile(
	req contractsopenai.ChatCompletionRequest,
	binding *platform.ModelRuntimeProfileBinding,
) error {
	if binding == nil {
		return fmt.Errorf("Anthropic model runtime profile is unavailable")
	}
	if err := binding.Validate(); err != nil {
		return fmt.Errorf("Anthropic model runtime profile is invalid")
	}
	profile := binding.Profile
	if !profile.AllowsEndpoint("messages") {
		return fmt.Errorf("model does not support the messages endpoint")
	}
	if req.MaxTokens != nil &&
		*req.MaxTokens > profile.Limits.MaxOutputTokens {
		return fmt.Errorf(
			"max_tokens exceeds model limit %d",
			profile.Limits.MaxOutputTokens,
		)
	}
	if req.Temperature != nil &&
		!profile.AllowsSamplingParameter("temperature") {
		return fmt.Errorf("model does not support temperature")
	}
	if req.TopP != nil && !profile.AllowsSamplingParameter("top_p") {
		return fmt.Errorf("model does not support top_p")
	}
	if req.TopK != nil && !profile.AllowsSamplingParameter("top_k") {
		return fmt.Errorf("model does not support top_k")
	}
	if len(req.Messages) > 0 &&
		strings.EqualFold(
			strings.TrimSpace(req.Messages[len(req.Messages)-1].Role),
			"assistant",
		) &&
		!profile.AssistantPrefill {
		return fmt.Errorf("model does not support assistant message prefill")
	}
	if chatRequestHasImage(req) && !profile.AllowsInputModality("image") {
		return fmt.Errorf("model does not support image input")
	}
	if chatRequestHasFile(req) && !profile.AllowsInputModality("pdf") {
		return fmt.Errorf("model does not support PDF input")
	}
	for _, tool := range req.Tools {
		if strings.EqualFold(strings.TrimSpace(tool.Type), "function") {
			if !profile.Tools.Custom {
				return fmt.Errorf("model does not support custom tools")
			}
			continue
		}
		if isAnthropicServerToolType(tool.Type) {
			if !profile.Tools.Server {
				return fmt.Errorf("model does not support server tools")
			}
			if len(profile.Tools.Supported) > 0 &&
				!profile.AllowsTool(anthropicServerToolCapability(tool.Type)) {
				return fmt.Errorf(
					"model does not support Anthropic server tool %q",
					tool.Type,
				)
			}
			continue
		}
		return fmt.Errorf("unsupported Anthropic tool type %q", tool.Type)
	}
	if req.ParallelToolCalls != nil &&
		*req.ParallelToolCalls &&
		!profile.Tools.Parallel {
		return fmt.Errorf("model does not support parallel tool calls")
	}
	if chatRequestHasResponseFormat(req) &&
		!profile.Tools.StructuredOutput {
		return fmt.Errorf("model does not support structured output")
	}
	if rawOutputEffort, present := req.OutputConfig["effort"]; present {
		outputEffort, ok := rawOutputEffort.(string)
		if !ok {
			return fmt.Errorf("output_config.effort must be a string")
		}
		normalized := normalizeProfileEffort(outputEffort)
		if normalized == "" || !profile.AllowsEffort(normalized) {
			return fmt.Errorf(
				"model does not support output_config.effort=%s",
				outputEffort,
			)
		}
		if req.Thinking != nil &&
			strings.EqualFold(strings.TrimSpace(req.Thinking.Type), "disabled") &&
			!profile.AllowsDisabledEffort(normalized) {
			return fmt.Errorf(
				"model does not support output_config.effort=%s with disabled thinking",
				outputEffort,
			)
		}
	}
	if value := strings.ToLower(strings.TrimSpace(req.ServiceTier)); value != "" &&
		!profile.AllowsServiceTier(value) {
		return fmt.Errorf("unsupported service_tier %q", value)
	}
	if value := strings.ToLower(strings.TrimSpace(req.InferenceGeo)); value != "" &&
		!profile.AllowsInferenceGeo(value) {
		return fmt.Errorf("unsupported inference_geo %q", value)
	}
	if value := strings.ToLower(strings.TrimSpace(req.Speed)); value != "" &&
		!profile.AllowsSpeed(value) {
		return fmt.Errorf("unsupported speed %q", value)
	}
	return validateAnthropicThinkingAgainstProfile(req, profile)
}

func validateAnthropicThinkingAgainstProfile(
	req contractsopenai.ChatCompletionRequest,
	profile platform.ModelRuntimeProfile,
) error {
	mode := ""
	if req.Thinking != nil {
		switch strings.ToLower(strings.TrimSpace(req.Thinking.Type)) {
		case "", "enabled":
			mode = "manual"
		case "adaptive":
			mode = "adaptive"
		case "disabled":
			mode = "disabled"
		default:
			return fmt.Errorf(
				"unsupported thinking type %q",
				req.Thinking.Type,
			)
		}
	} else if strings.EqualFold(
		profile.Thinking.DefaultMode,
		"adaptive",
	) || reasoningEffortEnabled(req.ReasoningEffort) {
		mode = profile.Thinking.DefaultMode
	}
	if mode != "" && !profile.AllowsThinkingMode(mode) {
		return fmt.Errorf("model does not support %s thinking", mode)
	}

	effort := normalizeProfileEffort(req.ReasoningEffort)
	if effort != "" && !profile.AllowsEffort(effort) {
		return fmt.Errorf("model does not support reasoning_effort=%s", effort)
	}
	if mode == "" || mode == "disabled" {
		if mode == "disabled" && !profile.AllowsDisabledEffort(effort) {
			return fmt.Errorf(
				"model does not support reasoning_effort=%s with disabled thinking",
				effort,
			)
		}
		return nil
	}
	for _, parameter := range profile.Sampling.DisallowedWithThinking {
		switch strings.ToLower(strings.TrimSpace(parameter)) {
		case "temperature":
			if req.Temperature != nil {
				return fmt.Errorf(
					"temperature cannot be combined with %s thinking",
					mode,
				)
			}
		case "top_p":
			if req.TopP != nil {
				return fmt.Errorf(
					"top_p cannot be combined with %s thinking",
					mode,
				)
			}
		case "top_k":
			if req.TopK != nil {
				return fmt.Errorf(
					"top_k cannot be combined with %s thinking",
					mode,
				)
			}
		}
	}
	return nil
}

func reasoningEffortEnabled(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return value != "" && value != "none"
}

func normalizeProfileEffort(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "minimal", "low":
		return "low"
	case "medium":
		return "medium"
	case "high":
		return "high"
	case "xhigh":
		return "xhigh"
	case "max":
		return "max"
	default:
		return ""
	}
}

func isAnthropicServerToolType(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	_, ok := map[string]struct{}{
		"web_search_20260318":     {},
		"web_search_20260209":     {},
		"web_search_20250305":     {},
		"web_fetch_20260318":      {},
		"web_fetch_20260309":      {},
		"web_fetch_20260209":      {},
		"web_fetch_20250910":      {},
		"code_execution_20260521": {},
		"code_execution_20260120": {},
		"code_execution_20250825": {},
	}[value]
	return ok
}

func anthropicServerToolCapability(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	for _, capability := range []string{
		"web_search",
		"web_fetch",
		"code_execution",
	} {
		if strings.HasPrefix(value, capability+"_") {
			return capability
		}
	}
	return ""
}

func anthropicOpenAIToolPricingPolicy(
	req contractsopenai.ChatCompletionRequest,
) (map[string]string, error) {
	body, err := json.Marshal(struct {
		Tools []contractsopenai.ToolDefinition `json:"tools,omitempty"`
	}{
		Tools: req.Tools,
	})
	if err != nil {
		return nil, fmt.Errorf("invalid Anthropic tools")
	}
	return anthropicCodeExecutionPricingPolicy(
		body,
		platform.ProviderOperationAnthropicMessages,
	)
}

func mergePricingConditions(
	base map[string]string,
	additional map[string]string,
) map[string]string {
	if len(additional) == 0 {
		return base
	}
	if base == nil {
		base = make(map[string]string, len(additional))
	}
	for key, value := range additional {
		base[key] = value
	}
	return base
}

func withAnthropicEffectiveSpeedMetadata(
	metadata map[string]interface{},
	requested, effective string,
) map[string]interface{} {
	requested = strings.ToLower(strings.TrimSpace(requested))
	effective = strings.ToLower(strings.TrimSpace(effective))
	if requested == "" || effective == "" {
		return metadata
	}
	result := make(map[string]interface{}, len(metadata)+2)
	for key, value := range metadata {
		result[key] = value
	}
	result["requested_speed"] = requested
	result["effective_speed"] = effective
	return result
}
