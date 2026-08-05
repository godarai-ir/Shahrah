package proxy

import (
	"encoding/json"
	"fmt"
	"math"
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

	if err := validateMoonshotSamplingRequest(req, model); err != nil {
		return err
	}

	if choice, ok := req.ToolChoice.(string); ok && strings.EqualFold(strings.TrimSpace(choice), "required") && model != "kimi-k3" {
		return fmt.Errorf("model does not support tool_choice=required")
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
		if req.Thinking != nil {
			return fmt.Errorf("kimi-k3 does not accept the thinking parameter")
		}

		return nil
	}

	if req.Thinking == nil {
		return nil
	}

	if req.Thinking.BudgetTokens != 0 || strings.TrimSpace(req.Thinking.Display) != "" {
		return fmt.Errorf("moonshot thinking accepts only its documented type field")
	}

	thinkingType := strings.ToLower(strings.TrimSpace(req.Thinking.Type))
	if model == "kimi-k2.7-code" || model == "kimi-k2.7-code-highspeed" {
		if thinkingType != "enabled" {
			return fmt.Errorf("%s always has thinking enabled", model)
		}

		return nil
	}

	mode := "adaptive"
	if thinkingType == "disabled" {
		mode = "disabled"
	} else if thinkingType != "enabled" {
		return fmt.Errorf("unsupported thinking.type %q", req.Thinking.Type)
	}

	if !profile.AllowsThinkingMode(mode) {
		return fmt.Errorf("model does not support thinking.type=%s", req.Thinking.Type)
	}

	return nil
}

func validateMoonshotSamplingRequest(
	req contractsopenai.ChatCompletionRequest,
	model string,
) error {
	if model == "kimi-k3" {
		if req.Temperature != nil {
			return fmt.Errorf("kimi-k3 does not accept temperature")
		}

		if req.TopP != nil {
			return fmt.Errorf("kimi-k3 does not accept top_p")
		}

		return validateMoonshotFixedAdditionalFields(req.AdditionalFields)
	}

	expectedTemperature := 1.0
	if model == "kimi-k2.6" && req.Thinking != nil && strings.EqualFold(req.Thinking.Type, "disabled") {
		expectedTemperature = 0.6
	}

	if req.Temperature != nil && math.Abs(*req.Temperature-expectedTemperature) > 1e-9 {
		return fmt.Errorf("temperature must equal %g for the selected Moonshot mode", expectedTemperature)
	}

	if req.TopP != nil && math.Abs(*req.TopP-0.95) > 1e-9 {
		return fmt.Errorf("top_p must equal 0.95 for the selected Moonshot model")
	}

	return validateMoonshotFixedAdditionalFields(req.AdditionalFields)
}

func validateMoonshotFixedAdditionalFields(fields map[string]json.RawMessage) error {
	if raw := fields["n"]; len(raw) > 0 {
		var value int
		if json.Unmarshal(raw, &value) != nil || value != 1 {
			return fmt.Errorf("n must equal 1 for Moonshot models")
		}
	}

	for _, parameter := range []string{"presence_penalty", "frequency_penalty"} {
		if raw := fields[parameter]; len(raw) > 0 {
			var value float64
			if json.Unmarshal(raw, &value) != nil || value != 0 {
				return fmt.Errorf("%s must equal 0 for Moonshot models", parameter)
			}
		}
	}

	return nil
}
