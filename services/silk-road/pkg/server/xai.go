package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

const xaiDefaultBaseURL = "https://api.x.ai/v1"

type xaiAdapter struct {
	compat *openAICompatibleAdapter
}

func newXAIAdapter() providerAdapter {
	return &xaiAdapter{compat: &openAICompatibleAdapter{
		defaultBaseURL:        xaiDefaultBaseURL,
		authHeader:            "Authorization",
		allowedRequestHeaders: []string{"x-grok-conv-id"},
	}}
}

func (a *xaiAdapter) Dispatch(ctx context.Context, client *http.Client, payload platform.TunnelRequest) (*adapterResponse, error) {
	endpoint := xaiRuntimeEndpoint(payload.Operation)
	if endpoint == "" {
		return nil, fmt.Errorf("xai_unsupported_operation: %s", payload.Operation)
	}
	if payload.Operation == platform.ProviderOperationOpenAIImagesVariations {
		return nil, fmt.Errorf("xai_unsupported_operation: image variations")
	}
	if payload.Operation == platform.ProviderOperationXAIBatchRequestsAdd {
		if err := validateXAIBatchRequestProfiles(payload); err != nil {
			return nil, err
		}
	}
	model := xaiRequestModel(payload)
	binding, err := validateXAIRuntimeProfile(payload, model, endpoint)
	if err != nil {
		return nil, err
	}
	if err := validateXAIConversationHeader(payload, binding, endpoint); err != nil {
		return nil, err
	}
	if payload.Operation == platform.ProviderOperationOpenAIImagesEdits &&
		!strings.HasPrefix(strings.ToLower(strings.TrimSpace(headerValue(payload.Headers, "content-type"))), "application/json") {
		return nil, fmt.Errorf("xai_image_edits_require_json")
	}
	if binding != nil {
		if err := validateXAIRequestBody(payload, binding.Profile); err != nil {
			return nil, err
		}
	}
	return a.dispatchWithRetry(ctx, client, payload)
}

func (a *xaiAdapter) DispatchRaw(ctx context.Context, client *http.Client, payload platform.TunnelRequest, body io.Reader, contentLength int64) (*adapterResponse, error) {
	switch payload.Operation {
	case platform.ProviderOperationOpenAIFilesCreate,
		platform.ProviderOperationOpenAIFilesContent:
		// These are the only xAI resource operations whose payload or response
		// must stay streamed through the raw tunnel.
	default:
		return nil, fmt.Errorf("xai_unsupported_raw_operation: %s", payload.Operation)
	}
	if xaiRuntimeEndpoint(payload.Operation) != "files" {
		return nil, fmt.Errorf("xai_unsupported_operation: %s", payload.Operation)
	}
	return a.compat.DispatchRaw(ctx, client, payload, body, contentLength)
}

func (a *xaiAdapter) dispatchWithRetry(ctx context.Context, client *http.Client, payload platform.TunnelRequest) (*adapterResponse, error) {
	method, _, err := resolveOpenAIOperation(payload)
	if err != nil {
		return nil, err
	}
	maximumAttempts := 1
	if method == http.MethodGet {
		maximumAttempts = 2
	}
	for attempt := 1; attempt <= maximumAttempts; attempt++ {
		response, dispatchErr := a.compat.Dispatch(ctx, client, payload)
		if dispatchErr != nil || response == nil || attempt == maximumAttempts || !xaiRetryableStatus(response.StatusCode) {
			return response, dispatchErr
		}
		wait := xaiRetryDelay(response.Headers.Get("Retry-After"))
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 64*1024))
		_ = response.Body.Close()
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
	return nil, fmt.Errorf("xai_retry_exhausted")
}

func xaiRetryableStatus(status int) bool {
	return status == http.StatusTooManyRequests || status == http.StatusBadGateway || status == http.StatusServiceUnavailable || status == http.StatusGatewayTimeout
}

func xaiRetryDelay(value string) time.Duration {
	value = strings.TrimSpace(value)
	if seconds, err := strconv.Atoi(value); err == nil && seconds >= 0 {
		wait := time.Duration(seconds) * time.Second
		if wait > 5*time.Second {
			return 5 * time.Second
		}
		return wait
	}
	if at, err := http.ParseTime(value); err == nil {
		wait := time.Until(at)
		if wait < 0 {
			return 0
		}
		if wait > 5*time.Second {
			return 5 * time.Second
		}
		return wait
	}
	return 100 * time.Millisecond
}

func xaiRequestModel(payload platform.TunnelRequest) string {
	model := strings.TrimSpace(payload.Deployment.Model)
	if len(payload.Body) == 0 {
		return model
	}
	var body struct {
		Model string `json:"model"`
	}
	if json.Unmarshal(payload.Body, &body) == nil && strings.TrimSpace(body.Model) != "" {
		return strings.TrimSpace(body.Model)
	}
	return model
}

func validateXAIRuntimeProfile(payload platform.TunnelRequest, model, endpoint string) (*platform.ModelRuntimeProfileBinding, error) {
	if !xaiOperationRequiresModel(payload.Operation) {
		return nil, nil
	}
	if strings.TrimSpace(model) == "" {
		return nil, fmt.Errorf("xai_model_required")
	}
	binding, ok := platform.FindModelRuntimeProfile(payload.ModelRuntimeProfiles, model)
	if !ok {
		return nil, fmt.Errorf("xai_runtime_profile_missing: %s", model)
	}
	if !strings.EqualFold(strings.TrimSpace(binding.Provider), "xai") {
		return nil, fmt.Errorf("xai_runtime_profile_provider_mismatch")
	}
	if !binding.Profile.AllowsEndpoint(endpoint) {
		return nil, fmt.Errorf("xai_unsupported_model_operation: %s does not allow %s", model, endpoint)
	}
	transport := "http"
	if payload.Stream {
		transport = "sse"
	}
	if !binding.Profile.AllowsTransport(transport) {
		return nil, fmt.Errorf("xai_unsupported_transport: %s", transport)
	}
	if isXAIBatchOperation(payload.Operation) && !binding.Profile.Batch.Eligible {
		return nil, fmt.Errorf("xai_batch_unsupported_model: %s", model)
	}
	return binding, nil
}

func validateXAIConversationHeader(payload platform.TunnelRequest, binding *platform.ModelRuntimeProfileBinding, endpoint string) error {
	value := strings.TrimSpace(headerValue(payload.Headers, "x-grok-conv-id"))
	if value == "" {
		return nil
	}
	if endpoint != "chat_completions" || binding == nil {
		return fmt.Errorf("xai_conversation_header_not_allowed")
	}
	if !containsFold(binding.Profile.RequestHeaders.Allowed, "x-grok-conv-id") ||
		!containsFold(binding.Profile.RequestHeaders.ByEndpoint[endpoint], "x-grok-conv-id") {
		return fmt.Errorf("xai_conversation_header_not_allowed")
	}
	if len(value) > 256 || strings.IndexFunc(value, func(r rune) bool {
		return unicode.IsControl(r) || r > unicode.MaxASCII
	}) >= 0 {
		return fmt.Errorf("xai_conversation_header_invalid")
	}
	return nil
}

func validateXAIRequestBody(payload platform.TunnelRequest, profile platform.ModelRuntimeProfile) error {
	if len(payload.Body) == 0 {
		return nil
	}
	switch payload.Operation {
	case platform.ProviderOperationChatCompletions:
		return validateXAIChatRequestBody(payload.Body, profile)
	case platform.ProviderOperationOpenAIResponsesCreate,
		platform.ProviderOperationOpenAIResponsesCompact,
		platform.ProviderOperationOpenAIResponsesInputTokens:
		return validateXAIResponsesRequestBody(payload.Body, profile)
	case platform.ProviderOperationOpenAIImagesGenerations,
		platform.ProviderOperationOpenAIImagesEdits:
		return validateXAIImageRequestBody(payload.Body, payload.Operation, profile)
	default:
		return nil
	}
}

func validateXAIChatRequestBody(bodyBytes []byte, profile platform.ModelRuntimeProfile) error {
	var body map[string]json.RawMessage
	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		return fmt.Errorf("xai_invalid_json: %w", err)
	}
	for _, parameter := range profile.UnsupportedParameters {
		if _, present := body[parameter]; present {
			return fmt.Errorf("xai_unsupported_parameter: %s", parameter)
		}
	}
	if raw := body["messages"]; len(raw) > 0 {
		var messages []struct {
			Role string `json:"role"`
		}
		if json.Unmarshal(raw, &messages) != nil {
			return fmt.Errorf("xai_invalid_messages")
		}
		for _, message := range messages {
			switch strings.ToLower(strings.TrimSpace(message.Role)) {
			case "system", "user", "assistant", "tool":
			default:
				return fmt.Errorf("xai_unsupported_message_role: %s", message.Role)
			}
		}
	}
	if raw := body["reasoning_effort"]; len(raw) > 0 {
		var value string
		if json.Unmarshal(raw, &value) != nil || !profile.AllowsEffort(value) {
			return fmt.Errorf("xai_unsupported_reasoning_effort: %s", value)
		}
	}
	if raw := body["service_tier"]; len(raw) > 0 {
		var value string
		if json.Unmarshal(raw, &value) != nil || !profile.AllowsServiceTier(value) {
			return fmt.Errorf("xai_unsupported_service_tier: %s", value)
		}
	}
	for _, parameter := range []string{"temperature", "top_p"} {
		if _, present := body[parameter]; present && !profile.AllowsSamplingParameter(parameter) {
			return fmt.Errorf("xai_unsupported_parameter: %s", parameter)
		}
	}
	if raw := body["tools"]; len(raw) > 0 {
		var tools []struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(raw, &tools); err != nil {
			return fmt.Errorf("xai_invalid_tools")
		}
		for _, tool := range tools {
			family := xaiToolFamily(tool.Type)
			if family == "" || !profile.AllowsTool(family) {
				return fmt.Errorf("xai_unsupported_tool: %s", tool.Type)
			}
		}
	}
	return nil
}

func validateXAIResponsesRequestBody(bodyBytes []byte, profile platform.ModelRuntimeProfile) error {
	var body map[string]json.RawMessage
	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		return fmt.Errorf("xai_invalid_json: %w", err)
	}
	for _, parameter := range profile.UnsupportedParameters {
		if _, present := body[parameter]; present {
			return fmt.Errorf("xai_unsupported_parameter: %s", parameter)
		}
	}
	for _, parameter := range []string{"temperature", "top_p"} {
		if _, present := body[parameter]; present && !profile.AllowsSamplingParameter(parameter) {
			return fmt.Errorf("xai_unsupported_parameter: %s", parameter)
		}
	}
	if raw := body["service_tier"]; len(raw) > 0 {
		var value string
		if json.Unmarshal(raw, &value) != nil || !profile.AllowsServiceTier(value) {
			return fmt.Errorf("xai_unsupported_service_tier: %s", value)
		}
	}
	if raw := body["reasoning"]; len(raw) > 0 {
		var reasoning struct {
			Effort string `json:"effort"`
		}
		if json.Unmarshal(raw, &reasoning) != nil || (reasoning.Effort != "" && !profile.AllowsEffort(reasoning.Effort)) {
			return fmt.Errorf("xai_unsupported_reasoning_effort: %s", reasoning.Effort)
		}
	}
	if raw := body["parallel_tool_calls"]; len(raw) > 0 {
		var enabled bool
		if json.Unmarshal(raw, &enabled) != nil {
			return fmt.Errorf("xai_invalid_parallel_tool_calls")
		}
		if enabled && !profile.Tools.Parallel {
			return fmt.Errorf("xai_unsupported_parallel_tool_calls")
		}
	}
	if raw := body["tools"]; len(raw) > 0 {
		var tools []struct {
			Type string `json:"type"`
		}
		if json.Unmarshal(raw, &tools) != nil {
			return fmt.Errorf("xai_invalid_tools")
		}
		for _, tool := range tools {
			family := xaiToolFamily(tool.Type)
			if family == "" || !profile.AllowsTool(family) {
				return fmt.Errorf("xai_unsupported_tool: %s", tool.Type)
			}
		}
	}
	if raw := body["input"]; len(raw) > 0 {
		var input interface{}
		if json.Unmarshal(raw, &input) != nil {
			return fmt.Errorf("xai_invalid_responses_input")
		}
		if err := validateXAIResponsesInput(input, profile); err != nil {
			return err
		}
	}
	if raw := body["text"]; len(raw) > 0 && !profile.Tools.StructuredOutput {
		var textConfig map[string]interface{}
		if json.Unmarshal(raw, &textConfig) == nil {
			if _, present := textConfig["format"]; present {
				return fmt.Errorf("xai_unsupported_structured_output")
			}
		}
	}
	return nil
}

func validateXAIResponsesInput(value interface{}, profile platform.ModelRuntimeProfile) error {
	switch item := value.(type) {
	case []interface{}:
		for _, child := range item {
			if err := validateXAIResponsesInput(child, profile); err != nil {
				return err
			}
		}
	case map[string]interface{}:
		if role, ok := item["role"].(string); ok {
			switch strings.ToLower(strings.TrimSpace(role)) {
			case "system", "user", "assistant", "tool":
			case "developer":
				return fmt.Errorf("xai_unsupported_message_role: developer")
			default:
				return fmt.Errorf("xai_unsupported_message_role: %s", role)
			}
		}
		if kind, ok := item["type"].(string); ok && strings.EqualFold(strings.TrimSpace(kind), "input_image") && !profile.AllowsInputModality("image") {
			return fmt.Errorf("xai_unsupported_input_modality: image")
		}
		for key, child := range item {
			if key == "schema" || key == "parameters" {
				continue
			}
			if err := validateXAIResponsesInput(child, profile); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateXAIImageRequestBody(bodyBytes []byte, operation platform.ProviderOperation, profile platform.ModelRuntimeProfile) error {
	var body map[string]json.RawMessage
	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		return fmt.Errorf("xai_invalid_json: %w", err)
	}
	allowedFields := map[string]struct{}{
		"model": {}, "prompt": {}, "n": {}, "aspect_ratio": {}, "resolution": {},
		"response_format": {}, "storage_options": {}, "stream": {},
	}
	if operation == platform.ProviderOperationOpenAIImagesEdits {
		allowedFields["image"] = struct{}{}
		allowedFields["images"] = struct{}{}
	}
	for key := range body {
		if _, allowed := allowedFields[key]; !allowed {
			return fmt.Errorf("xai_unsupported_image_parameter: %s", key)
		}
	}
	if raw := body["prompt"]; len(raw) > 0 && profile.Limits.MaxInputCharacters > 0 {
		var prompt string
		if json.Unmarshal(raw, &prompt) != nil {
			return fmt.Errorf("xai_invalid_prompt")
		}
		if len([]rune(prompt)) > profile.Limits.MaxInputCharacters {
			return fmt.Errorf("xai_prompt_too_long")
		}
	}
	if raw := body["n"]; len(raw) > 0 {
		var count int
		if json.Unmarshal(raw, &count) != nil || count < 1 || (profile.ImageControls.MaximumOutputs > 0 && count > profile.ImageControls.MaximumOutputs) {
			return fmt.Errorf("xai_invalid_image_count")
		}
	}
	for key, allowed := range map[string][]string{
		"aspect_ratio":    profile.ImageControls.AspectRatios,
		"resolution":      profile.ImageControls.Resolutions,
		"response_format": profile.ImageControls.ResponseFormats,
	} {
		if raw := body[key]; len(raw) > 0 {
			var value string
			if json.Unmarshal(raw, &value) != nil || !containsFold(allowed, value) {
				return fmt.Errorf("xai_unsupported_%s: %s", key, value)
			}
		}
	}
	if raw, present := body["storage_options"]; present {
		if !profile.ImageControls.StorageOptions {
			return fmt.Errorf("xai_unsupported_storage_options")
		}
		if err := validateXAIImageStorageOptions(raw); err != nil {
			return err
		}
	}
	if raw, present := body["stream"]; present {
		if !profile.ImageControls.Streaming {
			return fmt.Errorf("xai_image_streaming_unsupported")
		}
		var enabled bool
		if err := json.Unmarshal(raw, &enabled); err != nil {
			return fmt.Errorf("xai_image_stream_invalid")
		}
	}
	if operation == platform.ProviderOperationOpenAIImagesEdits {
		count, err := xaiImageReferenceCount(body)
		if err != nil {
			return err
		}
		if count == 0 {
			return fmt.Errorf("xai_image_reference_required")
		}
		if maximum := profile.ImageControls.MaximumReferenceImages; maximum > 0 && count > maximum {
			return fmt.Errorf("xai_too_many_image_references")
		}
	}
	return nil
}

func validateXAIImageStorageOptions(raw json.RawMessage) error {
	var options map[string]json.RawMessage
	if json.Unmarshal(raw, &options) != nil || options == nil {
		return fmt.Errorf("xai_invalid_storage_options")
	}
	for key := range options {
		switch key {
		case "filename", "expires_after", "public_url":
		default:
			return fmt.Errorf("xai_unsupported_storage_option: %s", key)
		}
	}
	var filename string
	if json.Unmarshal(options["filename"], &filename) != nil || strings.TrimSpace(filename) == "" {
		return fmt.Errorf("xai_storage_filename_required")
	}
	fileExpiry, hasFileExpiry, err := xaiExpirySeconds(options["expires_after"])
	if err != nil {
		return fmt.Errorf("xai_invalid_storage_expiry")
	}
	publicRaw, hasPublicURL := options["public_url"]
	if !hasPublicURL {
		return nil
	}
	var enabled bool
	if json.Unmarshal(publicRaw, &enabled) == nil {
		return nil
	}
	var publicOptions map[string]json.RawMessage
	if json.Unmarshal(publicRaw, &publicOptions) != nil || publicOptions == nil {
		return fmt.Errorf("xai_invalid_storage_public_url")
	}
	for key := range publicOptions {
		if key != "expires_after" {
			return fmt.Errorf("xai_unsupported_public_url_option: %s", key)
		}
	}
	publicExpiry, hasPublicExpiry, err := xaiExpirySeconds(publicOptions["expires_after"])
	if err != nil || !hasPublicExpiry {
		return fmt.Errorf("xai_invalid_public_url_expiry")
	}
	if hasFileExpiry && publicExpiry > fileExpiry {
		return fmt.Errorf("xai_public_url_outlives_file")
	}
	return nil
}

func xaiExpirySeconds(raw json.RawMessage) (int, bool, error) {
	if len(raw) == 0 {
		return 0, false, nil
	}
	var value int
	if json.Unmarshal(raw, &value) != nil || value < 3600 || value > 30*24*60*60 {
		return 0, true, fmt.Errorf("invalid expiry")
	}
	return value, true, nil
}

func xaiImageReferenceCount(body map[string]json.RawMessage) (int, error) {
	count := 0
	if raw := body["image"]; len(raw) > 0 && string(raw) != "null" {
		if err := validateXAIImageReference(raw); err != nil {
			return 0, err
		}
		count++
	}
	if raw := body["images"]; len(raw) > 0 {
		var images []json.RawMessage
		if json.Unmarshal(raw, &images) != nil {
			return 0, fmt.Errorf("xai_invalid_image_references")
		}
		for _, image := range images {
			if err := validateXAIImageReference(image); err != nil {
				return 0, err
			}
		}
		count += len(images)
	}
	return count, nil
}

func validateXAIImageReference(raw json.RawMessage) error {
	var reference map[string]json.RawMessage
	if json.Unmarshal(raw, &reference) != nil || reference == nil {
		return fmt.Errorf("xai_invalid_image_reference")
	}
	for key := range reference {
		switch key {
		case "url", "file_id", "type":
		default:
			return fmt.Errorf("xai_unsupported_image_reference_field: %s", key)
		}
	}
	values := 0
	for _, key := range []string{"url", "file_id"} {
		if rawValue := reference[key]; len(rawValue) > 0 {
			var value string
			if json.Unmarshal(rawValue, &value) != nil || strings.TrimSpace(value) == "" {
				return fmt.Errorf("xai_invalid_image_reference")
			}
			values++
		}
	}
	if values != 1 {
		return fmt.Errorf("xai_invalid_image_reference")
	}
	if rawType := reference["type"]; len(rawType) > 0 {
		var imageType string
		if json.Unmarshal(rawType, &imageType) != nil || !strings.EqualFold(strings.TrimSpace(imageType), "image_url") {
			return fmt.Errorf("xai_invalid_image_reference_type")
		}
	}
	return nil
}

func xaiToolFamily(value string) string {
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

func xaiRuntimeEndpoint(operation platform.ProviderOperation) string {
	switch operation {
	case platform.ProviderOperationChatCompletions:
		return "chat_completions"
	case platform.ProviderOperationXAIDeferredChatGet:
		return "deferred_chat_completions"
	case platform.ProviderOperationOpenAIResponsesCreate,
		platform.ProviderOperationOpenAIResponsesGet,
		platform.ProviderOperationOpenAIResponsesDelete,
		platform.ProviderOperationOpenAIResponsesInputItems:
		return "responses"
	case platform.ProviderOperationOpenAIResponsesCompact:
		return "responses_compact"
	case platform.ProviderOperationOpenAIResponsesInputTokens:
		return "responses_input_tokens"
	case platform.ProviderOperationOpenAIImagesGenerations:
		return "image_generations"
	case platform.ProviderOperationOpenAIImagesEdits:
		return "image_edits"
	case platform.ProviderOperationOpenAIImagesVariations:
		return "image_variations"
	case platform.ProviderOperationXAIBatchesCreate,
		platform.ProviderOperationXAIBatchesList,
		platform.ProviderOperationXAIBatchesGet:
		return "batches"
	case platform.ProviderOperationXAIBatchRequestsList,
		platform.ProviderOperationXAIBatchRequestsAdd:
		return "batch_requests"
	case platform.ProviderOperationXAIBatchResults:
		return "batch_results"
	case platform.ProviderOperationXAIBatchesCancel:
		return "batches"
	case platform.ProviderOperationOpenAIFilesCreate,
		platform.ProviderOperationOpenAIFilesList,
		platform.ProviderOperationOpenAIFilesGet,
		platform.ProviderOperationOpenAIFilesDelete,
		platform.ProviderOperationOpenAIFilesContent:
		return "files"
	case platform.ProviderOperationXAIFilesInitialize,
		platform.ProviderOperationXAIFilesUploadChunks:
		return "files"
	case platform.ProviderOperationXAIFilesUpdate:
		return "files"
	case platform.ProviderOperationXAIFilePublicURLCreate,
		platform.ProviderOperationXAIFilePublicURLRevoke:
		return "file_public_urls"
	case platform.ProviderOperationOpenAIModelsList, platform.ProviderOperationOpenAIModelsGet:
		return "models"
	case platform.ProviderOperationXAILanguageModelsList, platform.ProviderOperationXAILanguageModelsGet:
		return "language_models"
	case platform.ProviderOperationXAIImageGenerationModelsList, platform.ProviderOperationXAIImageGenerationModelsGet:
		return "image_generation_models"
	default:
		return ""
	}
}

func xaiOperationRequiresModel(operation platform.ProviderOperation) bool {
	switch operation {
	case platform.ProviderOperationChatCompletions,
		platform.ProviderOperationOpenAIResponsesCreate,
		platform.ProviderOperationOpenAIResponsesCompact,
		platform.ProviderOperationOpenAIResponsesInputTokens,
		platform.ProviderOperationOpenAIImagesGenerations,
		platform.ProviderOperationOpenAIImagesEdits:
		return true
	default:
		return false
	}
}

func isXAIBatchOperation(operation platform.ProviderOperation) bool {
	switch operation {
	case platform.ProviderOperationXAIBatchesCreate,
		platform.ProviderOperationXAIBatchesList,
		platform.ProviderOperationXAIBatchesGet,
		platform.ProviderOperationXAIBatchRequestsList,
		platform.ProviderOperationXAIBatchRequestsAdd,
		platform.ProviderOperationXAIBatchResults,
		platform.ProviderOperationXAIBatchesCancel:
		return true
	default:
		return false
	}
}

func validateXAIBatchRequestProfiles(payload platform.TunnelRequest) error {
	if len(payload.Body) == 0 {
		return fmt.Errorf("xai_batch_requests_required")
	}
	var request struct {
		BatchRequests []struct {
			BatchRequest map[string]json.RawMessage `json:"batch_request"`
		} `json:"batch_requests"`
	}
	if json.Unmarshal(payload.Body, &request) != nil || len(request.BatchRequests) == 0 {
		return fmt.Errorf("xai_batch_requests_required")
	}
	for _, item := range request.BatchRequests {
		if len(item.BatchRequest) != 1 {
			return fmt.Errorf("xai_batch_request_invalid")
		}
		for requestType, raw := range item.BatchRequest {
			var modelRequest struct {
				Model string `json:"model"`
			}
			if json.Unmarshal(raw, &modelRequest) != nil || strings.TrimSpace(modelRequest.Model) == "" {
				return fmt.Errorf("xai_batch_request_model_required")
			}
			binding, ok := platform.FindModelRuntimeProfile(payload.ModelRuntimeProfiles, modelRequest.Model)
			if !ok || !strings.EqualFold(binding.Provider, "xai") {
				return fmt.Errorf("xai_runtime_profile_missing: %s", modelRequest.Model)
			}
			if !binding.Profile.Batch.Eligible {
				return fmt.Errorf("xai_batch_unsupported_model: %s", modelRequest.Model)
			}
			endpoint := xaiBatchRequestEndpoint(requestType)
			if endpoint == "" || !containsFold(binding.Profile.Batch.SupportedEndpoints, endpoint) {
				return fmt.Errorf("xai_batch_unsupported_request_type: %s", requestType)
			}
		}
	}
	return nil
}

func xaiBatchRequestEndpoint(requestType string) string {
	switch strings.ToLower(strings.TrimSpace(requestType)) {
	case "chat_get_completion":
		return "chat_completions"
	case "image_generate", "image_generation", "image_get_generation":
		return "image_generations"
	case "image_edit", "image_edit_image":
		return "image_edits"
	default:
		return ""
	}
}

func containsFold(values []string, wanted string) bool {
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), wanted) {
			return true
		}
	}
	return false
}
