package proxy

import (
	"encoding/json"
	"fmt"
	"strings"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func validateMoonshotChatRuntimeProfile(
	req contractsopenai.ChatCompletionRequest,
	binding *platform.ModelRuntimeProfileBinding,
) error {
	if binding == nil {
		return fmt.Errorf("moonshot model runtime profile is unavailable")
	}

	if err := binding.Validate(); err != nil || !strings.EqualFold(binding.Provider, "moonshot") {
		return fmt.Errorf("moonshot model runtime profile is invalid")
	}

	profile := binding.Profile
	if !profile.AllowsEndpoint("chat_completions") {
		return fmt.Errorf("model does not support chat completions")
	}

	transport := "http"
	if req.Stream {
		transport = "sse"
	}

	if !profile.AllowsTransport(transport) {
		return fmt.Errorf("model does not support %s transport", transport)
	}

	if chatRequestHasImage(req) && !profile.AllowsInputModality("image") {
		return fmt.Errorf("model does not support image input")
	}

	if chatRequestHasVideo(req) && !profile.AllowsInputModality("video") {
		return fmt.Errorf("model does not support video input")
	}

	for _, message := range req.Messages {
		switch strings.ToLower(strings.TrimSpace(message.Role)) {
		case "system", "user", "assistant", "tool":
		default:
			return fmt.Errorf("model does not support message role %q", message.Role)
		}
	}

	if req.MaxTokens != nil && profile.Limits.MaxOutputTokens > 0 && *req.MaxTokens > profile.Limits.MaxOutputTokens {
		return fmt.Errorf("max_tokens exceeds model limit %d", profile.Limits.MaxOutputTokens)
	}

	if req.MaxCompletionTokens != nil && profile.Limits.MaxOutputTokens > 0 &&
		*req.MaxCompletionTokens > profile.Limits.MaxOutputTokens {
		return fmt.Errorf("max_completion_tokens exceeds model limit %d", profile.Limits.MaxOutputTokens)
	}

	model := strings.ToLower(strings.TrimSpace(binding.Slug))
	if err := validateMoonshotThinkingRequest(req, model, profile); err != nil {
		return err
	}

	if err := validateMoonshotFixedAdditionalFields(req.AdditionalFields, model); err != nil {
		return err
	}

	if err := validateMoonshotDecodingAdditionalFields(req.AdditionalFields, model); err != nil {
		return err
	}

	if moonshotModelUsesFixedControls(model) && req.ParallelToolCalls != nil {
		return fmt.Errorf("model does not support parallel_tool_calls")
	}

	if err := validateMoonshotToolChoiceRequest(req, model); err != nil {
		return err
	}

	if req.ParallelToolCalls != nil && *req.ParallelToolCalls && !profile.Tools.Parallel {
		return fmt.Errorf("model does not support parallel tool calls")
	}

	if chatRequestHasResponseFormat(req) && !profile.Tools.StructuredOutput {
		return fmt.Errorf("model does not support structured output")
	}

	for _, tool := range req.Tools {
		if !strings.EqualFold(strings.TrimSpace(tool.Type), "function") || !profile.AllowsTool("function_calling") {
			return fmt.Errorf("model does not support Moonshot tool %q", tool.Type)
		}
	}

	return nil
}

func validateMoonshotThinkingRequest(
	req contractsopenai.ChatCompletionRequest,
	model string,
	profile platform.ModelRuntimeProfile,
) error {
	effort := strings.ToLower(strings.TrimSpace(req.ReasoningEffort))
	if effort != "" && !profile.AllowsEffort(effort) {
		return fmt.Errorf("model does not support reasoning_effort=%s", effort)
	}

	if model == "kimi-k3" {
		if req.Thinking == nil {
			return nil
		}

		if req.Thinking.BudgetTokens != 0 || strings.TrimSpace(req.Thinking.Display) != "" {
			return fmt.Errorf("kimi-k3 does not accept thinking budget or display controls")
		}

		switch strings.ToLower(strings.TrimSpace(req.Thinking.Type)) {
		case "enabled", "adaptive":
			return nil
		case "disabled":
			return fmt.Errorf("kimi-k3 does not support disabling thinking")
		default:
			return fmt.Errorf("unsupported thinking.type %q", req.Thinking.Type)
		}
	}

	if req.Thinking == nil {
		return nil
	}

	if req.Thinking.BudgetTokens != 0 || strings.TrimSpace(req.Thinking.Display) != "" {
		return fmt.Errorf("moonshot thinking accepts only its documented type field")
	}

	thinkingType := strings.ToLower(strings.TrimSpace(req.Thinking.Type))
	if model == "kimi-k2.7-code" || model == "kimi-k2.7-code-highspeed" {
		if thinkingType != "enabled" && thinkingType != "adaptive" {
			return fmt.Errorf("%s always has thinking enabled", model)
		}

		return nil
	}

	mode := "adaptive"
	if thinkingType == "disabled" {
		mode = "disabled"
	} else if thinkingType != "enabled" && thinkingType != "adaptive" {
		return fmt.Errorf("unsupported thinking.type %q", req.Thinking.Type)
	}

	if !profile.AllowsThinkingMode(mode) {
		return fmt.Errorf("model does not support thinking.type=%s", req.Thinking.Type)
	}

	return nil
}

func validateMoonshotFixedAdditionalFields(fields map[string]json.RawMessage, model string) error {
	if !moonshotModelUsesFixedControls(model) {
		return nil
	}

	if raw := fields["n"]; len(raw) > 0 {
		if !moonshotJSONNull(raw) {
			var value int
			if json.Unmarshal(raw, &value) != nil || value != 1 {
				return fmt.Errorf("n must equal 1 for Moonshot models")
			}
		}
	}

	for _, parameter := range []string{"presence_penalty", "frequency_penalty"} {
		raw := fields[parameter]
		if len(raw) == 0 || moonshotJSONNull(raw) {
			continue
		}

		var value float64
		if json.Unmarshal(raw, &value) != nil || value != 0 {
			return fmt.Errorf("%s must equal 0 for Moonshot models", parameter)
		}
	}

	return nil
}

func validateMoonshotDecodingAdditionalFields(fields map[string]json.RawMessage, model string) error {
	if !moonshotModelUsesFixedControls(model) {
		return nil
	}

	for _, parameter := range []string{
		"top_k", "seed", "random_seed", "logit_bias", "parallel_tool_calls",
	} {
		if raw, present := fields[parameter]; present && !moonshotJSONNull(raw) {
			return fmt.Errorf("model does not support %s", parameter)
		}
	}

	logprobsEnabled := false
	if raw, present := fields["logprobs"]; present && !moonshotJSONNull(raw) {
		if json.Unmarshal(raw, &logprobsEnabled) != nil {
			return fmt.Errorf("logprobs must be a boolean")
		}
	}

	if raw, present := fields["top_logprobs"]; present && !moonshotJSONNull(raw) {
		var value int
		if json.Unmarshal(raw, &value) != nil || value < 0 || value > 20 {
			return fmt.Errorf("top_logprobs must be an integer between 0 and 20")
		}

		if !logprobsEnabled {
			return fmt.Errorf("top_logprobs requires logprobs=true")
		}
	}

	if raw, present := fields["stop"]; present && !moonshotJSONNull(raw) {
		if err := validateMoonshotStopRequest(raw); err != nil {
			return err
		}
	}

	return nil
}

func validateMoonshotStopRequest(raw json.RawMessage) error {
	var single string
	if json.Unmarshal(raw, &single) == nil {
		if len([]byte(single)) > 32 {
			return fmt.Errorf("stop strings must not exceed 32 bytes")
		}

		return nil
	}

	var multiple []string
	if json.Unmarshal(raw, &multiple) != nil || len(multiple) > 5 {
		return fmt.Errorf("stop must be a string or an array of at most 5 strings")
	}

	for _, value := range multiple {
		if len([]byte(value)) > 32 {
			return fmt.Errorf("stop strings must not exceed 32 bytes")
		}
	}

	return nil
}

func validateMoonshotToolChoiceRequest(req contractsopenai.ChatCompletionRequest, model string) error {
	if !moonshotModelUsesFixedControls(model) || req.ToolChoice == nil {
		return nil
	}

	raw := req.AdditionalFields["tool_choice"]
	if len(raw) == 0 {
		encoded, err := json.Marshal(req.ToolChoice)
		if err != nil {
			return fmt.Errorf("invalid tool_choice")
		}

		raw = encoded
	}

	if moonshotJSONNull(raw) {
		return nil
	}

	var choice string
	if json.Unmarshal(raw, &choice) == nil {
		switch strings.ToLower(strings.TrimSpace(choice)) {
		case "auto", "none":
			return nil
		case "required":
			if moonshotModelDisallowsRequiredToolChoice(model) {
				return fmt.Errorf("model does not support tool_choice=required")
			}

			return nil
		default:
			return fmt.Errorf("invalid tool_choice")
		}
	}

	var specified struct {
		Type     string `json:"type"`
		Function struct {
			Name string `json:"name"`
		} `json:"function"`
	}
	if json.Unmarshal(raw, &specified) != nil ||
		!strings.EqualFold(strings.TrimSpace(specified.Type), "function") ||
		strings.TrimSpace(specified.Function.Name) == "" {
		return fmt.Errorf("invalid tool_choice")
	}

	thinkingDisabled := req.Thinking != nil &&
		strings.EqualFold(strings.TrimSpace(req.Thinking.Type), "disabled")
	if moonshotModelAlwaysThinks(model) ||
		(strings.EqualFold(strings.TrimSpace(model), "kimi-k2.6") && !thinkingDisabled) {
		return fmt.Errorf("specific tool_choice is incompatible with thinking enabled")
	}

	return nil
}

func moonshotModelDisallowsRequiredToolChoice(model string) bool {
	switch strings.ToLower(strings.TrimSpace(model)) {
	case "kimi-k2.7-code", "kimi-k2.7-code-highspeed", "kimi-k2.6":
		return true
	default:
		return false
	}
}

func moonshotModelAlwaysThinks(model string) bool {
	switch strings.ToLower(strings.TrimSpace(model)) {
	case "kimi-k3", "kimi-k2.7-code", "kimi-k2.7-code-highspeed":
		return true
	default:
		return false
	}
}

func moonshotModelUsesFixedControls(model string) bool {
	switch strings.ToLower(strings.TrimSpace(model)) {
	case "kimi-k3", "kimi-k2.7-code", "kimi-k2.7-code-highspeed", "kimi-k2.6":
		return true
	default:
		return false
	}
}

func moonshotJSONNull(raw json.RawMessage) bool {
	return strings.EqualFold(strings.TrimSpace(string(raw)), "null")
}
