package proxy

import (
	"encoding/json"
	"fmt"
	"strings"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
)

const (
	deepSeekProviderName      = "deepseek"
	deepSeekFlashModelID      = "deepseek-v4-flash"
	deepSeekProModelID        = "deepseek-v4-pro"
	deepSeekResponsesEndpoint = "responses"
)

func validateDeepSeekResponsesRuntimeProfile(
	body []byte,
	binding *platform.ModelRuntimeProfileBinding,
) error {
	if binding == nil {
		return fmt.Errorf("deepseek model runtime profile is unavailable")
	}

	if err := binding.Validate(); err != nil || !strings.EqualFold(binding.Provider, deepSeekProviderName) {
		return fmt.Errorf("deepseek model runtime profile is invalid")
	}

	model := strings.ToLower(strings.TrimSpace(binding.Slug))
	if model != deepSeekFlashModelID {
		return fmt.Errorf("DeepSeek Responses API is supported only by %s", deepSeekFlashModelID)
	}

	profile := binding.Profile
	if !profile.AllowsEndpoint(deepSeekResponsesEndpoint) {
		return fmt.Errorf("model does not support Responses create")
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(body, &fields); err != nil {
		return fmt.Errorf("invalid DeepSeek Responses request")
	}

	for parameter := range fields {
		if parameter == "model" {
			continue
		}

		if !profile.AllowsParameter(deepSeekResponsesEndpoint, parameter) {
			return fmt.Errorf("deepseek Responses parameter %s is unsupported or gateway-managed", parameter)
		}
	}

	if err := validateDeepSeekResponseNestedControls(fields); err != nil {
		return err
	}

	input, hasInput := fields["input"]

	instructions, hasInstructions := fields["instructions"]
	if (!hasInput || isDeepSeekJSONNull(input)) && (!hasInstructions || isDeepSeekJSONNull(instructions)) {
		return fmt.Errorf("deepseek Responses requires input or instructions")
	}

	if hasInput {
		if err := validateDeepSeekResponseInput(input); err != nil {
			return err
		}
	}

	var req contractsopenai.ResponseRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return fmt.Errorf("invalid DeepSeek Responses request: %w", err)
	}

	transport := "http"
	if req.Stream {
		transport = "sse"
	}

	if !profile.AllowsTransport(transport) {
		return fmt.Errorf("model does not support %s transport", transport)
	}

	if err := validateDeepSeekResponseNumericControls(req, profile); err != nil {
		return err
	}

	if err := validateDeepSeekResponseReasoning(req); err != nil {
		return err
	}

	if err := validateDeepSeekResponseText(req.Text); err != nil {
		return err
	}

	return validateDeepSeekResponseTools(req, profile)
}

func validateDeepSeekResponseNumericControls(
	req contractsopenai.ResponseRequest,
	profile platform.ModelRuntimeProfile,
) error {
	if req.MaxOutputTokens != nil && (*req.MaxOutputTokens < 1 ||
		profile.Limits.MaxOutputTokens > 0 && *req.MaxOutputTokens > profile.Limits.MaxOutputTokens) {
		return fmt.Errorf("requested output tokens exceed model limit %d", profile.Limits.MaxOutputTokens)
	}

	if req.Temperature != nil && (*req.Temperature < 0 || *req.Temperature > 2) {
		return fmt.Errorf("deepseek Responses temperature must be between 0 and 2")
	}

	if req.TopP != nil && (*req.TopP < 0 || *req.TopP > 1) {
		return fmt.Errorf("deepseek Responses top_p must be between 0 and 1")
	}

	if req.TopLogprobs != nil && (*req.TopLogprobs < 0 || *req.TopLogprobs > 20) {
		return fmt.Errorf("deepseek Responses top_logprobs must be between 0 and 20")
	}

	return nil
}

func validateDeepSeekResponseReasoning(req contractsopenai.ResponseRequest) error {
	effort := ""
	if req.Reasoning != nil {
		effort = strings.ToLower(strings.TrimSpace(req.Reasoning.Effort))
		if effort != "" && !isSupportedDeepSeekResponseEffort(effort) {
			return fmt.Errorf("unsupported deepseek Responses reasoning effort %q", req.Reasoning.Effort)
		}
	}

	if effort != "none" && (req.Temperature != nil || req.TopP != nil) {
		return fmt.Errorf("temperature and top_p are ineffective while deepseek Responses thinking is enabled")
	}

	return nil
}

func isSupportedDeepSeekResponseEffort(effort string) bool {
	switch effort {
	case "none", "minimal", "low", "medium", "high", "xhigh", "max":
		return true
	default:
		return false
	}
}

func validateDeepSeekResponseTools(
	req contractsopenai.ResponseRequest,
	profile platform.ModelRuntimeProfile,
) error {
	for _, tool := range req.Tools {
		if !strings.EqualFold(strings.TrimSpace(tool.Type), "function") ||
			!profile.AllowsTool("function_calling") {
			return fmt.Errorf("deepseek Responses tool %q is outside the enabled capability scope", tool.Type)
		}
	}

	return validateDeepSeekToolChoice(req.ToolChoice)
}

func validateDeepSeekResponseNestedControls(fields map[string]json.RawMessage) error {
	if err := validateDeepSeekResponseObjectFields(fields["reasoning"], "reasoning", "effort"); err != nil {
		return err
	}

	return validateDeepSeekResponseObjectFields(fields["text"], "text", "format")
}

func validateDeepSeekResponseObjectFields(
	raw json.RawMessage,
	objectName string,
	allowedFields ...string,
) error {
	if len(raw) == 0 || isDeepSeekJSONNull(raw) {
		return nil
	}

	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		return fmt.Errorf("invalid deepseek Responses %s control", objectName)
	}

	allowed := make(map[string]struct{}, len(allowedFields))
	for _, field := range allowedFields {
		allowed[field] = struct{}{}
	}

	for field := range object {
		if _, ok := allowed[field]; !ok {
			return fmt.Errorf("deepseek Responses %s.%s has no effect or is unsupported", objectName, field)
		}
	}

	return nil
}

func validateDeepSeekResponseInput(raw json.RawMessage) error {
	if isDeepSeekJSONNull(raw) {
		return nil
	}

	var text string
	if json.Unmarshal(raw, &text) == nil {
		return nil
	}

	var items []map[string]interface{}
	if err := json.Unmarshal(raw, &items); err != nil {
		return fmt.Errorf("deepseek Responses input must be a string or input item array")
	}

	for _, item := range items {
		itemType := ""
		if rawType, exists := item["type"]; exists {
			itemType = strings.ToLower(strings.TrimSpace(fmt.Sprint(rawType)))
		}

		switch itemType {
		case "message":
			if err := validateDeepSeekResponseMessage(item); err != nil {
				return err
			}
		case "function_call", "function_call_output", "reasoning":
		case "":
			if _, hasRole := item["role"]; !hasRole {
				return fmt.Errorf("deepseek Responses input item type is required")
			}

			if err := validateDeepSeekResponseMessage(item); err != nil {
				return err
			}
		default:
			return fmt.Errorf("deepseek Responses input item type %q is outside the enabled capability scope", itemType)
		}
	}

	return nil
}

func validateDeepSeekResponseMessage(item map[string]interface{}) error {
	role := strings.ToLower(strings.TrimSpace(fmt.Sprint(item["role"])))
	switch role {
	case "system", "developer", "user", "assistant":
	default:
		return fmt.Errorf("deepseek Responses message role %q is unsupported", role)
	}

	content, exists := item["content"]
	if !exists {
		return fmt.Errorf("deepseek Responses message content is required")
	}

	if _, ok := content.(string); ok {
		return nil
	}

	blocks, ok := content.([]interface{})
	if !ok {
		return fmt.Errorf("deepseek Responses message content must be text")
	}

	for _, value := range blocks {
		block, ok := value.(map[string]interface{})
		if !ok {
			return fmt.Errorf("deepseek Responses message content block is invalid")
		}

		blockType := strings.ToLower(strings.TrimSpace(fmt.Sprint(block["type"])))
		if blockType != "input_text" && blockType != "output_text" {
			return fmt.Errorf("deepseek Responses content type %q is unsupported", blockType)
		}
	}

	return nil
}

func validateDeepSeekResponseText(text *contractsopenai.ResponseTextConfig) error {
	if text == nil || text.Format == nil {
		return nil
	}

	format, ok := text.Format.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid deepseek Responses text.format")
	}

	rawType, exists := format["type"]
	if !exists {
		return fmt.Errorf("deepseek Responses text.format.type is required")
	}

	formatType := strings.ToLower(strings.TrimSpace(fmt.Sprint(rawType)))
	switch formatType {
	case "text", "json_object", "json_schema":
		return nil
	default:
		return fmt.Errorf("unsupported deepseek Responses text.format type %q", formatType)
	}
}

func isDeepSeekJSONNull(raw json.RawMessage) bool {
	return strings.EqualFold(strings.TrimSpace(string(raw)), "null")
}

func validateDeepSeekChatRuntimeProfile(
	req contractsopenai.ChatCompletionRequest,
	binding *platform.ModelRuntimeProfileBinding,
) error {
	if binding == nil {
		return fmt.Errorf("deepseek model runtime profile is unavailable")
	}

	if err := binding.Validate(); err != nil || !strings.EqualFold(binding.Provider, deepSeekProviderName) {
		return fmt.Errorf("deepseek model runtime profile is invalid")
	}

	model := strings.ToLower(strings.TrimSpace(binding.Slug))
	if !isSupportedDeepSeekChatModel(model) {
		return fmt.Errorf("unsupported deepseek model %q", binding.Slug)
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

	if chatRequestHasImage(req) || chatRequestHasVideo(req) {
		return fmt.Errorf("deepseek direct chat supports text input only")
	}

	for _, message := range req.Messages {
		switch strings.ToLower(strings.TrimSpace(message.Role)) {
		case "system", "user", "assistant", "tool":
		default:
			return fmt.Errorf("model does not support message role %q", message.Role)
		}
	}

	for _, limit := range []*int{req.MaxTokens, req.MaxCompletionTokens} {
		if limit != nil && (*limit < 1 || profile.Limits.MaxOutputTokens > 0 && *limit > profile.Limits.MaxOutputTokens) {
			return fmt.Errorf("requested output tokens exceed model limit %d", profile.Limits.MaxOutputTokens)
		}
	}

	if req.MaxCompletionTokens != nil {
		return fmt.Errorf("deepseek supports max_tokens, not max_completion_tokens")
	}

	if req.TopK != nil || req.ServiceTier != "" || req.Speed != "" || req.InferenceGeo != "" ||
		len(req.OutputConfig) > 0 || len(req.CacheControl) > 0 || len(req.ExtraBody) > 0 {
		return fmt.Errorf("request contains a parameter outside the enabled deepseek chat scope")
	}

	if err := validateDeepSeekThinkingRequest(req, profile); err != nil {
		return err
	}

	if req.ParallelToolCalls != nil {
		return fmt.Errorf("deepseek parallel_tool_calls is outside the enabled capability scope")
	}

	for _, tool := range req.Tools {
		if !strings.EqualFold(strings.TrimSpace(tool.Type), "function") || !profile.AllowsTool("function_calling") {
			return fmt.Errorf("model does not support deepseek tool %q", tool.Type)
		}

		if tool.Strict != nil || tool.Function != nil && tool.Function.Strict != nil {
			return fmt.Errorf("deepseek strict tool mode is outside the enabled capability scope")
		}
	}

	if err := validateDeepSeekToolChoice(req.ToolChoice); err != nil {
		return err
	}

	if err := validateDeepSeekResponseFormat(req.ResponseFormat); err != nil {
		return err
	}

	for _, parameter := range []string{"user_id", "strict", "prefix"} {
		if _, present := req.AdditionalFields[parameter]; present {
			return fmt.Errorf("deepseek parameter %s is reserved or outside the enabled scope", parameter)
		}
	}

	return nil
}

func isSupportedDeepSeekChatModel(model string) bool {
	return model == deepSeekFlashModelID || model == deepSeekProModelID
}

func validateDeepSeekThinkingRequest(
	req contractsopenai.ChatCompletionRequest,
	profile platform.ModelRuntimeProfile,
) error {
	mode := profile.Thinking.DefaultMode
	if mode == "" {
		mode = "adaptive"
	}

	if req.Thinking != nil {
		if req.Thinking.BudgetTokens != 0 || strings.TrimSpace(req.Thinking.Display) != "" {
			return fmt.Errorf("deepseek thinking accepts only its documented type field")
		}

		switch strings.ToLower(strings.TrimSpace(req.Thinking.Type)) {
		case "enabled":
			mode = "adaptive"
		case "disabled":
			mode = "disabled"
		default:
			return fmt.Errorf("unsupported deepseek thinking.type %q", req.Thinking.Type)
		}
	}

	if !profile.AllowsThinkingMode(mode) {
		return fmt.Errorf("model does not support deepseek thinking mode %q", mode)
	}

	enabled := mode != "disabled"

	effort := strings.ToLower(strings.TrimSpace(req.ReasoningEffort))
	if effort != "" && !profile.AllowsEffort(effort) {
		return fmt.Errorf("model does not support reasoning_effort=%s", effort)
	}

	if !enabled && effort != "" {
		return fmt.Errorf("reasoning_effort requires deepseek thinking to be enabled")
	}

	if enabled {
		if req.Temperature != nil || req.TopP != nil {
			return fmt.Errorf("temperature and top_p are ineffective in deepseek thinking mode")
		}

		for _, parameter := range []string{"presence_penalty", "frequency_penalty"} {
			if _, present := req.AdditionalFields[parameter]; present {
				return fmt.Errorf("%s is ineffective in deepseek thinking mode", parameter)
			}
		}
	}

	return nil
}

func validateDeepSeekToolChoice(value interface{}) error {
	if value == nil {
		return nil
	}

	if choice, ok := value.(string); ok {
		switch strings.ToLower(strings.TrimSpace(choice)) {
		case "none", "auto", "required":
			return nil
		default:
			return fmt.Errorf("unsupported deepseek tool_choice %q", choice)
		}
	}

	if choice, ok := value.(map[string]interface{}); ok {
		if strings.EqualFold(strings.TrimSpace(fmt.Sprint(choice["type"])), "function") {
			return nil
		}
	}

	return fmt.Errorf("invalid deepseek tool_choice")
}

func validateDeepSeekResponseFormat(value interface{}) error {
	if value == nil {
		return nil
	}

	raw, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("invalid deepseek response_format")
	}

	var format struct {
		Type string `json:"type"`
	}
	if json.Unmarshal(raw, &format) != nil || format.Type != "json_object" {
		return fmt.Errorf("deepseek response_format supports only json_object")
	}

	return nil
}
