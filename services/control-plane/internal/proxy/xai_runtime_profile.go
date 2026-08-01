package proxy

import (
	"encoding/json"
	"fmt"
	"strings"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func validateXAIChatRuntimeProfile(
	req contractsopenai.ChatCompletionRequest,
	binding *platform.ModelRuntimeProfileBinding,
) error {
	if binding == nil {
		return fmt.Errorf("xAI model runtime profile is unavailable")
	}

	if err := binding.Validate(); err != nil || !strings.EqualFold(binding.Provider, "xai") {
		return fmt.Errorf("xAI model runtime profile is invalid")
	}

	profile := binding.Profile
	if !profile.AllowsEndpoint("chat_completions") {
		return fmt.Errorf("model does not support chat completions")
	}

	if chatRequestDeferred(req) {
		if req.Stream {
			return fmt.Errorf("deferred chat completions do not support streaming")
		}

		if !profile.AllowsEndpoint("deferred_chat_completions") || !profile.DeferredCompletions.Eligible {
			return fmt.Errorf("model does not support deferred chat completions")
		}
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

	if chatRequestHasFile(req) && !profile.AllowsTool("attachment_search") {
		return fmt.Errorf("model does not support file attachments")
	}

	for _, message := range req.Messages {
		switch strings.ToLower(strings.TrimSpace(message.Role)) {
		case "system", "user", "assistant", "tool":
		case "developer":
			return fmt.Errorf("model developer messages are not enabled without model-specific conformance evidence")
		default:
			return fmt.Errorf("model does not support message role %q", message.Role)
		}
	}

	if req.MaxTokens != nil && profile.Limits.MaxOutputTokens > 0 && *req.MaxTokens > profile.Limits.MaxOutputTokens {
		return fmt.Errorf("max_tokens exceeds model limit %d", profile.Limits.MaxOutputTokens)
	}

	for parameter, present := range map[string]bool{
		"temperature": req.Temperature != nil,
		"top_p":       req.TopP != nil,
		"top_k":       req.TopK != nil,
	} {
		if present && !profile.AllowsSamplingParameter(parameter) {
			return fmt.Errorf("model does not support %s", parameter)
		}
	}

	if effort := strings.ToLower(strings.TrimSpace(req.ReasoningEffort)); effort != "" && !profile.AllowsEffort(effort) {
		return fmt.Errorf("model does not support reasoning_effort=%s", effort)
	}

	if tier := strings.ToLower(strings.TrimSpace(req.ServiceTier)); tier != "" && !profile.AllowsServiceTier(tier) {
		return fmt.Errorf("unsupported service_tier %q", tier)
	}

	if req.ParallelToolCalls != nil && *req.ParallelToolCalls && !profile.Tools.Parallel {
		return fmt.Errorf("model does not support parallel tool calls")
	}

	if chatRequestHasResponseFormat(req) && !profile.Tools.StructuredOutput {
		return fmt.Errorf("model does not support structured output")
	}

	for _, parameter := range profile.UnsupportedParameters {
		if parameter == "stop" && req.Stop != nil {
			return fmt.Errorf("model does not support stop")
		}

		if _, present := req.AdditionalFields[parameter]; present {
			return fmt.Errorf("model does not support %s", parameter)
		}
	}

	for _, tool := range req.Tools {
		family := xaiChatToolFamily(tool.Type)
		if family == "" || !profile.AllowsTool(family) {
			return fmt.Errorf("model does not support xAI tool %q", tool.Type)
		}
	}

	return nil
}

func chatRequestDeferred(req contractsopenai.ChatCompletionRequest) bool {
	raw, ok := req.AdditionalFields["deferred"]
	if !ok || len(raw) == 0 {
		return false
	}

	var deferred bool

	return json.Unmarshal(raw, &deferred) == nil && deferred
}

func xaiChatToolFamily(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "function":
		return "function_calling"
	case "web_search":
		return "web_search"
	case "x_search":
		return "x_search"
	case "code_execution", "code_interpreter":
		return "code_execution"
	case "attachment_search":
		return "attachment_search"
	case "collections_search":
		return "collections_search"
	case "file_search":
		return "file_search"
	case "mcp":
		return "mcp"
	default:
		return ""
	}
}

func validateXAIImageRuntimeProfile(
	operation string,
	contentType string,
	metadata map[string]interface{},
	binding *platform.ModelRuntimeProfileBinding,
) error {
	if binding == nil {
		return fmt.Errorf("xAI model runtime profile is unavailable")
	}

	if err := binding.Validate(); err != nil || !strings.EqualFold(binding.Provider, "xai") {
		return fmt.Errorf("xAI model runtime profile is invalid")
	}

	endpoint := "image_generations"

	switch strings.ToLower(strings.TrimSpace(operation)) {
	case "edits":
		endpoint = "image_edits"
	case "variations":
		return fmt.Errorf("xAI image models do not support image variations")
	}

	profile := binding.Profile
	if !profile.AllowsEndpoint(endpoint) {
		return fmt.Errorf("model does not support %s", strings.ReplaceAll(endpoint, "_", " "))
	}

	if endpoint == "image_edits" && strings.EqualFold(profile.ImageControls.EditTransport, "json") && !isJSONContentType(contentType) {
		return fmt.Errorf("xAI image edits require application/json")
	}

	if maximum := profile.Limits.MaxInputCharacters; maximum > 0 {
		if promptLength, ok := metadataInt(metadata, "prompt_length"); ok && promptLength > maximum {
			return fmt.Errorf("prompt exceeds model limit of %d characters", maximum)
		}
	}

	if maximum := profile.ImageControls.MaximumOutputs; maximum > 0 {
		if count, ok := metadataInt(metadata, "n"); ok && count > maximum {
			return fmt.Errorf("n exceeds model limit of %d", maximum)
		}
	}

	if maximum := profile.ImageControls.MaximumReferenceImages; maximum > 0 {
		if count, ok := metadataInt(metadata, "image_count"); ok && count > maximum {
			return fmt.Errorf("image exceeds model limit of %d reference images", maximum)
		}
	}

	if value, ok := metadataString(metadata, "aspect_ratio"); ok && !runtimeValueAllowed(profile.ImageControls.AspectRatios, value) {
		return fmt.Errorf("unsupported aspect_ratio %q", value)
	}

	if value, ok := metadataString(metadata, "resolution"); ok && !runtimeValueAllowed(profile.ImageControls.Resolutions, value) {
		return fmt.Errorf("unsupported resolution %q", value)
	}

	if value, ok := metadataString(metadata, "response_format"); ok && !runtimeValueAllowed(profile.ImageControls.ResponseFormats, value) {
		return fmt.Errorf("unsupported response_format %q", value)
	}

	if present, _ := metadata["storage_options_present"].(bool); present && !profile.ImageControls.StorageOptions { //nolint:errcheck // Optional metadata uses the zero value when absent or mismatched.
		return fmt.Errorf("model does not support storage_options")
	}

	return nil
}

func metadataInt(metadata map[string]interface{}, key string) (int, bool) {
	if metadata == nil {
		return 0, false
	}

	switch value := metadata[key].(type) {
	case int:
		return value, true
	case float64:
		return int(value), true
	default:
		return 0, false
	}
}

func metadataString(metadata map[string]interface{}, key string) (string, bool) {
	if metadata == nil {
		return "", false
	}

	value, ok := metadata[key].(string)
	value = strings.TrimSpace(value)

	return value, ok && value != ""
}

func runtimeValueAllowed(values []string, requested string) bool {
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), strings.TrimSpace(requested)) {
			return true
		}
	}

	return false
}

func withXAIUsageMetadata(metadata map[string]interface{}, provider string, usage *contractsopenai.Usage) map[string]interface{} {
	if usage == nil || !strings.EqualFold(strings.TrimSpace(provider), "xai") {
		return metadata
	}

	if metadata == nil {
		metadata = map[string]interface{}{}
	}

	metadata["provider_cost_in_usd_ticks"] = usage.CostInUSDTicks
	metadata["provider_cost_exact"] = true
	metadata["num_server_side_tools_used"] = usage.NumServerSideToolsUsed

	return metadata
}

func validateXAIResponseResourceOperation(operation platform.ProviderOperation, profile platform.ModelRuntimeProfile) error {
	switch operation {
	case platform.ProviderOperationOpenAIResponsesGet:
		if profile.ResponseResources.Retrieve {
			return nil
		}
	case platform.ProviderOperationOpenAIResponsesDelete:
		if profile.ResponseResources.Delete {
			return nil
		}
	case platform.ProviderOperationOpenAIResponsesCancel:
		return fmt.Errorf("xAI does not document response cancellation for this model")
	case platform.ProviderOperationOpenAIResponsesInputItems:
		return fmt.Errorf("xAI does not document response input_items retrieval for this model")
	}

	return fmt.Errorf("xAI response operation is not supported by this model")
}
