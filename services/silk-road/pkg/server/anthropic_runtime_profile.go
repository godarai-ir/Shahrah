package server

import (
	"encoding/json"
	"fmt"
	"strings"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func anthropicRuntimeProfileForModel(
	payload platform.TunnelRequest,
	model, endpoint string,
) (*platform.ModelRuntimeProfileBinding, error) {
	binding, ok := platform.FindModelRuntimeProfile(
		payload.ModelRuntimeProfiles,
		model,
	)
	if !ok {
		return nil, fmt.Errorf(
			"anthropic model runtime profile is missing for %s",
			model,
		)
	}

	if err := binding.Validate(); err != nil {
		return nil, fmt.Errorf(
			"invalid anthropic model runtime profile: %w",
			err,
		)
	}

	for _, candidate := range payload.ModelRuntimeProfiles {
		if candidate.CatalogRevision != binding.CatalogRevision {
			return nil, fmt.Errorf(
				"anthropic model runtime profiles use mixed revisions",
			)
		}
	}

	if !binding.Profile.AllowsEndpoint(endpoint) {
		return nil, fmt.Errorf(
			"anthropic model %s does not support endpoint %s",
			model,
			endpoint,
		)
	}

	return binding, nil
}

func validateAnthropicOpenAIRequestWithProfile(
	req contractsopenai.ChatCompletionRequest,
	binding *platform.ModelRuntimeProfileBinding,
) error {
	if binding == nil {
		return fmt.Errorf("anthropic model runtime profile is required")
	}

	return validateAnthropicRequestWithRuntimeProfile(req, binding.Profile)
}

func validateAnthropicRequestWithRuntimeProfile(
	req contractsopenai.ChatCompletionRequest,
	profile platform.ModelRuntimeProfile,
) error {
	if req.MaxTokens != nil &&
		*req.MaxTokens > profile.Limits.MaxOutputTokens {
		return fmt.Errorf(
			"max_tokens exceeds model limit %d",
			profile.Limits.MaxOutputTokens,
		)
	}

	if len(req.Messages) > 0 &&
		strings.EqualFold(
			strings.TrimSpace(req.Messages[len(req.Messages)-1].Role),
			"assistant",
		) &&
		!profile.AssistantPrefill {
		return fmt.Errorf("model does not support assistant message prefill")
	}

	imageCount, fileCount := anthropicMessageMediaCounts(req.Messages)
	if imageCount > 0 && !profile.AllowsInputModality("image") {
		return fmt.Errorf("model does not support image input")
	}

	if fileCount > 0 && !profile.AllowsInputModality("pdf") {
		return fmt.Errorf("model does not support PDF input")
	}

	for _, tool := range req.Tools {
		if strings.EqualFold(strings.TrimSpace(tool.Type), "function") {
			if !profile.Tools.Custom {
				return fmt.Errorf("model does not support custom tools")
			}

			continue
		}

		if isSupportedAnthropicServerToolType(
			strings.ToLower(strings.TrimSpace(tool.Type)),
		) {
			if !profile.Tools.Server {
				return fmt.Errorf("model does not support server tools")
			}

			if len(profile.Tools.Supported) > 0 &&
				!profile.AllowsTool(
					anthropicServerToolCapability(tool.Type),
				) {
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

	if (req.ResponseFormat != nil ||
		(req.OutputConfig != nil && req.OutputConfig["format"] != nil)) &&
		!profile.Tools.StructuredOutput {
		return fmt.Errorf("model does not support structured output")
	}

	if rawOutputEffort, present := req.OutputConfig["effort"]; present {
		outputEffort, ok := rawOutputEffort.(string)
		if !ok {
			return fmt.Errorf("output_config.effort must be a string")
		}

		normalized := normalizeAnthropicEffort(outputEffort)
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

	return validateAnthropicThinkingWithRuntimeProfile(req, profile)
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

func validateAnthropicThinkingWithRuntimeProfile(
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
	) || anthropicReasoningEffortEnabled(req.ReasoningEffort) {
		mode = profile.Thinking.DefaultMode
	}

	if mode != "" && !profile.AllowsThinkingMode(mode) {
		return fmt.Errorf("model does not support %s thinking", mode)
	}

	effort := normalizeAnthropicEffort(req.ReasoningEffort)
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

	return nil
}

func validateAnthropicNativeRuntimeProfiles(
	payload platform.TunnelRequest,
) error {
	switch payload.Operation {
	case platform.ProviderOperationAnthropicMessages:
		return validateAnthropicNativeMessageBody(
			payload,
			payload.Body,
			"messages",
		)
	case platform.ProviderOperationAnthropicCountTokens:
		return validateAnthropicNativeMessageBody(
			payload,
			payload.Body,
			"count_tokens",
		)
	case platform.ProviderOperationAnthropicBatchesCreate:
		var batch struct {
			Requests []struct {
				Params json.RawMessage `json:"params"`
			} `json:"requests"`
		}
		if err := json.Unmarshal(payload.Body, &batch); err != nil {
			return fmt.Errorf("invalid Anthropic batch request")
		}

		for index, item := range batch.Requests {
			if err := validateAnthropicNativeMessageBody(
				payload,
				item.Params,
				"message_batches",
			); err != nil {
				return fmt.Errorf("batch item %d: %w", index, err)
			}
		}
	}

	return nil
}

func validateAnthropicNativeMessageBody(
	payload platform.TunnelRequest,
	body []byte,
	endpoint string,
) error {
	var request struct {
		Model        string                          `json:"model"`
		MaxTokens    *int                            `json:"max_tokens"`
		Temperature  *float64                        `json:"temperature"`
		TopP         *float64                        `json:"top_p"`
		TopK         *int                            `json:"top_k"`
		ServiceTier  string                          `json:"service_tier"`
		Speed        string                          `json:"speed"`
		InferenceGeo string                          `json:"inference_geo"`
		Thinking     *contractsopenai.ThinkingConfig `json:"thinking"`
		OutputConfig map[string]interface{}          `json:"output_config"`
		Tools        []json.RawMessage               `json:"tools"`
		Messages     []struct {
			Role    string          `json:"role"`
			Content json.RawMessage `json:"content"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(body, &request); err != nil {
		return fmt.Errorf("invalid Anthropic request")
	}

	binding, err := anthropicRuntimeProfileForModel(
		payload,
		request.Model,
		endpoint,
	)
	if err != nil {
		return err
	}

	effectiveProfile := binding.Profile
	effectiveProfile.Limits.MaxOutputTokens =
		binding.Profile.MaxOutputTokensFor(
			endpoint,
			anthropicRequestedBetaHeaders(payload),
		)

	chatRequest := contractsopenai.ChatCompletionRequest{
		Model:        request.Model,
		MaxTokens:    request.MaxTokens,
		Temperature:  request.Temperature,
		TopP:         request.TopP,
		TopK:         request.TopK,
		ServiceTier:  request.ServiceTier,
		Speed:        request.Speed,
		InferenceGeo: request.InferenceGeo,
		Thinking:     request.Thinking,
		OutputConfig: request.OutputConfig,
	}
	if err := validateAnthropicNativeTools(
		request.Tools,
		binding.Profile,
	); err != nil {
		return err
	}

	if request.OutputConfig["format"] != nil {
		chatRequest.ResponseFormat = request.OutputConfig["format"]
	}

	for _, message := range request.Messages {
		chatRequest.Messages = append(
			chatRequest.Messages,
			contractsopenai.ChatMessage{
				Role:    message.Role,
				Content: string(message.Content),
			},
		)
		if containsAnthropicContentType(message.Content, "image") &&
			!binding.Profile.AllowsInputModality("image") {
			return fmt.Errorf("model does not support image input")
		}

		if containsAnthropicContentType(message.Content, "document") &&
			!binding.Profile.AllowsInputModality("pdf") {
			return fmt.Errorf("model does not support PDF input")
		}
	}

	effectiveBinding := *binding
	effectiveBinding.Profile = effectiveProfile

	return validateAnthropicOpenAIRequestWithProfile(
		chatRequest,
		&effectiveBinding,
	)
}

func validateAnthropicNativeTools(
	tools []json.RawMessage,
	profile platform.ModelRuntimeProfile,
) error {
	for index, rawTool := range tools {
		var tool struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(rawTool, &tool); err != nil {
			return fmt.Errorf("invalid Anthropic tool %d", index)
		}

		toolType := strings.ToLower(strings.TrimSpace(tool.Type))
		if toolType == "" {
			if !profile.Tools.Custom {
				return fmt.Errorf("model does not support custom tools")
			}

			continue
		}

		capability := anthropicNativeToolCapability(toolType)
		// Unknown provider-native tool variants are passed through. This keeps
		// the native tunnel forward compatible while known model/tool
		// incompatibilities still fail closed.
		if capability == "" || len(profile.Tools.Supported) == 0 {
			continue
		}

		if !profile.AllowsTool(capability) {
			return fmt.Errorf(
				"model does not support Anthropic native tool %q",
				tool.Type,
			)
		}
	}

	return nil
}

func anthropicNativeToolCapability(toolType string) string {
	toolType = strings.ToLower(strings.TrimSpace(toolType))
	if toolType == "mcp_toolset" {
		return "mcp"
	}

	for prefix, capability := range map[string]string{
		"web_search_":       "web_search",
		"web_fetch_":        "web_fetch",
		"code_execution_":   "code_execution",
		"tool_search_tool_": "tool_search",
		"memory_":           "memory",
		"advisor_":          "advisor",
		"computer_":         "computer_use",
		"bash_":             "bash",
		"text_editor_":      "text_editor",
	} {
		if strings.HasPrefix(toolType, prefix) {
			return capability
		}
	}

	return ""
}

func anthropicRequestedBetaHeaders(
	payload platform.TunnelRequest,
) []string {
	set := commaSeparatedSet(headerValue(payload.Headers, "anthropic-beta"))

	values := make([]string, 0, len(set))
	for value := range set {
		values = append(values, value)
	}

	return values
}

func containsAnthropicContentType(content []byte, expected string) bool {
	if len(content) == 0 {
		return false
	}

	var blocks []struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(content, &blocks); err != nil {
		return false
	}

	for _, block := range blocks {
		if strings.EqualFold(strings.TrimSpace(block.Type), expected) {
			return true
		}
	}

	return false
}

func anthropicReasoningEffortEnabled(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return value != "" && value != "none"
}

func anthropicRuntimeProfilePayloadLimit(
	payload platform.TunnelRequest,
	fallback int64,
) int64 {
	limit := fallback

	for _, binding := range payload.ModelRuntimeProfiles {
		candidate := 0

		switch payload.Operation {
		case platform.ProviderOperationAnthropicMessages,
			platform.ProviderOperationAnthropicCountTokens:
			candidate = binding.Profile.Limits.MessagesBytes
		case platform.ProviderOperationAnthropicBatchesCreate:
			candidate = binding.Profile.Limits.BatchBytes
		}

		if candidate > 0 &&
			(limit <= 0 || int64(candidate) < limit) {
			limit = int64(candidate)
		}
	}

	return limit
}

func anthropicMessageMediaCounts(
	messages []contractsopenai.ChatMessage,
) (int, int) {
	imageCount := 0
	fileCount := 0

	for _, message := range messages {
		parts, ok, err := contractsopenai.ChatContentParts(message.Content)
		if !ok || err != nil {
			continue
		}

		for _, part := range parts {
			switch part.Type {
			case "image_url":
				imageCount++
			case "file":
				fileCount++
			}
		}
	}

	return imageCount, fileCount
}
