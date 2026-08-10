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

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
)

const moonshotDefaultBaseURL = "https://api.moonshot.ai/v1"

type moonshotAdapter struct {
	compat *openAICompatibleAdapter
}

func newMoonshotAdapter() providerAdapter {
	return &moonshotAdapter{compat: &openAICompatibleAdapter{
		defaultBaseURL: moonshotDefaultBaseURL,
		authHeader:     "Authorization",
	}}
}

func (a *moonshotAdapter) Dispatch(
	ctx context.Context,
	client *http.Client,
	payload platform.TunnelRequest,
) (*adapterResponse, error) {
	endpoint := moonshotRuntimeEndpoint(payload.Operation)
	if endpoint == "" {
		return nil, fmt.Errorf("moonshot_unsupported_operation: %s", payload.Operation)
	}

	if endpoint == "chat_completions" {
		body, err := decodeMoonshotChatBody(payload.Body)
		if err != nil {
			return nil, err
		}

		model := moonshotRequestModelFromDecoded(payload, body)

		binding, err := validateMoonshotRuntimeProfile(payload, model, endpoint)
		if err != nil {
			return nil, err
		}

		changed, err := normalizeMoonshotChatBody(body, model)
		if err != nil {
			return nil, err
		}

		if validationErr := validateMoonshotChatBodyDecoded(body, model, binding.Profile); validationErr != nil {
			return nil, validationErr
		}

		if changed {
			payload.Body, err = json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("moonshot_normalized_request_encode: %w", err)
			}
		}
	}

	return a.dispatchWithRetry(ctx, client, payload)
}

func (a *moonshotAdapter) dispatchWithRetry(
	ctx context.Context,
	client *http.Client,
	payload platform.TunnelRequest,
) (*adapterResponse, error) {
	method, _, err := resolveOpenAIOperation(payload)
	if err != nil {
		return nil, err
	}

	attempts := 1
	if method == http.MethodGet {
		attempts = 2
	}

	for attempt := 1; attempt <= attempts; attempt++ {
		response, dispatchErr := a.compat.Dispatch(ctx, client, payload)
		if dispatchErr != nil || response == nil || attempt == attempts || !moonshotRetryableStatus(response.StatusCode) {
			return response, dispatchErr
		}

		if _, copyErr := io.Copy(io.Discard, io.LimitReader(response.Body, 64*1024)); copyErr != nil {
			return nil, fmt.Errorf("moonshot_retry_response_drain: %w", copyErr)
		}

		if closeErr := response.Body.Close(); closeErr != nil {
			return nil, fmt.Errorf("moonshot_retry_response_close: %w", closeErr)
		}

		timer := time.NewTimer(moonshotRetryDelay(response.Headers.Get("Retry-After")))
		select {
		case <-ctx.Done():
			timer.Stop()

			return nil, ctx.Err()
		case <-timer.C:
		}
	}

	return nil, fmt.Errorf("moonshot_retry_exhausted")
}

func moonshotRetryableStatus(status int) bool {
	return status == http.StatusTooManyRequests ||
		status == http.StatusBadGateway ||
		status == http.StatusServiceUnavailable ||
		status == http.StatusGatewayTimeout
}

func moonshotRetryDelay(value string) time.Duration {
	if seconds, err := strconv.Atoi(strings.TrimSpace(value)); err == nil && seconds >= 0 {
		return min(time.Duration(seconds)*time.Second, 5*time.Second)
	}

	return 100 * time.Millisecond
}

func moonshotRuntimeEndpoint(operation platform.ProviderOperation) string {
	switch operation {
	case platform.ProviderOperationChatCompletions:
		return "chat_completions"
	case platform.ProviderOperationOpenAIModelsList,
		platform.ProviderOperationOpenAIModelsGet:
		return "models"
	default:
		return ""
	}
}

func moonshotRequestModelFromDecoded(payload platform.TunnelRequest, body map[string]json.RawMessage) string {
	model := strings.TrimSpace(payload.Deployment.Model)

	var requestedModel string
	if json.Unmarshal(body["model"], &requestedModel) == nil && strings.TrimSpace(requestedModel) != "" {
		model = strings.TrimSpace(requestedModel)
	}

	return model
}

func validateMoonshotRuntimeProfile(
	payload platform.TunnelRequest,
	model string,
	endpoint string,
) (*platform.ModelRuntimeProfileBinding, error) {
	if model == "" {
		return nil, fmt.Errorf("moonshot_model_required")
	}

	binding, ok := platform.FindModelRuntimeProfile(payload.ModelRuntimeProfiles, model)
	if !ok {
		return nil, fmt.Errorf("moonshot_runtime_profile_missing: %s", model)
	}

	if !strings.EqualFold(strings.TrimSpace(binding.Provider), "moonshot") {
		return nil, fmt.Errorf("moonshot_runtime_profile_provider_mismatch")
	}

	if !binding.Profile.AllowsEndpoint(endpoint) {
		return nil, fmt.Errorf("moonshot_unsupported_model_operation: %s", endpoint)
	}

	transport := "http"
	if payload.Stream {
		transport = "sse"
	}

	if !binding.Profile.AllowsTransport(transport) {
		return nil, fmt.Errorf("moonshot_unsupported_transport: %s", transport)
	}

	return binding, nil
}

func decodeMoonshotChatBody(bodyBytes []byte) (map[string]json.RawMessage, error) {
	var body map[string]json.RawMessage
	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		return nil, fmt.Errorf("moonshot_invalid_json: %w", err)
	}

	return body, nil
}

func normalizeMoonshotChatBody(body map[string]json.RawMessage, model string) (bool, error) {
	normalizedModel := strings.ToLower(strings.TrimSpace(model))
	changed := false

	switch normalizedModel {
	case "kimi-k3":
		changed = normalizeMoonshotNullControls(body)
		changed = normalizeMoonshotFixedCommonFields(body) || changed

		thinkingChanged, err := normalizeMoonshotK3Thinking(body)
		if err != nil {
			return false, err
		}

		changed = thinkingChanged || changed
		changed = deleteMoonshotParameter(body, "temperature") || changed
		changed = deleteMoonshotParameter(body, "top_p") || changed
	case "kimi-k2.7-code", "kimi-k2.7-code-highspeed":
		changed = normalizeMoonshotNullControls(body)
		changed = normalizeMoonshotFixedCommonFields(body) || changed

		thinkingChanged, err := normalizeMoonshotAlwaysEnabledThinking(body)
		if err != nil {
			return false, err
		}

		changed = thinkingChanged || changed
		changed = deleteMoonshotParameter(body, "temperature") || changed
		changed = deleteMoonshotParameter(body, "top_p") || changed
	case "kimi-k2.6":
		changed = normalizeMoonshotNullControls(body)
		changed = normalizeMoonshotFixedCommonFields(body) || changed

		thinkingChanged, err := normalizeMoonshotK26Thinking(body)
		if err != nil {
			return false, err
		}

		changed = thinkingChanged || changed

		changed = deleteMoonshotParameter(body, "temperature") || changed
		changed = deleteMoonshotParameter(body, "top_p") || changed
	}

	return changed, nil
}

func normalizeMoonshotNullControls(body map[string]json.RawMessage) bool {
	changed := false

	for _, parameter := range []string{
		"temperature", "top_p", "thinking", "reasoning_effort", "n",
		"presence_penalty", "frequency_penalty", "top_k", "seed", "random_seed",
		"stop", "logit_bias", "logprobs", "top_logprobs", "tool_choice",
		"max_tokens", "max_completion_tokens", "parallel_tool_calls",
	} {
		raw, present := body[parameter]
		if !present || !moonshotJSONNull(raw) {
			continue
		}

		delete(body, parameter)

		changed = true
	}

	return changed
}

func normalizeMoonshotFixedCommonFields(body map[string]json.RawMessage) bool {
	changed := false

	if raw, present := body["n"]; present {
		var value int
		if json.Unmarshal(raw, &value) == nil && value == 1 {
			delete(body, "n")

			changed = true
		}
	}

	for _, parameter := range []string{"presence_penalty", "frequency_penalty"} {
		raw, present := body[parameter]
		if !present {
			continue
		}

		var value float64
		if json.Unmarshal(raw, &value) != nil || value != 0 {
			continue
		}

		delete(body, parameter)

		changed = true
	}

	return changed
}

func normalizeMoonshotK3Thinking(body map[string]json.RawMessage) (bool, error) {
	raw, present := body["thinking"]
	if !present {
		return false, nil
	}

	fields, thinkingType, err := decodeMoonshotThinking(raw)
	if err != nil {
		return false, err
	}

	_, _, controlsErr := validateMoonshotThinkingControls(fields)
	if controlsErr != nil {
		return false, fmt.Errorf("moonshot_k3_thinking_controls_unsupported")
	}

	switch thinkingType {
	case "enabled", "adaptive":
		delete(body, "thinking")

		return true, nil
	case "disabled":
		return false, fmt.Errorf("moonshot_k3_thinking_disable_unsupported")
	default:
		return false, fmt.Errorf("moonshot_invalid_thinking_type: %s", thinkingType)
	}
}

func normalizeMoonshotAlwaysEnabledThinking(body map[string]json.RawMessage) (bool, error) {
	raw, present := body["thinking"]
	if !present {
		return false, nil
	}

	fields, thinkingType, err := decodeMoonshotThinking(raw)
	if err != nil {
		return false, err
	}

	hasKeepAll, hasNullKeep, controlsErr := validateMoonshotThinkingControls(fields)
	if controlsErr != nil {
		return false, fmt.Errorf("moonshot_k2_7_thinking_controls_unsupported")
	}

	switch thinkingType {
	case "enabled":
		if hasNullKeep {
			body["thinking"] = json.RawMessage(`{"type":"enabled"}`)

			return true, nil
		}

		return false, nil
	case "adaptive":
		body["thinking"] = canonicalMoonshotEnabledThinking(hasKeepAll)

		return true, nil
	case "disabled":
		return false, fmt.Errorf("moonshot_k2_7_thinking_always_enabled")
	default:
		return false, fmt.Errorf("moonshot_invalid_thinking_type: %s", thinkingType)
	}
}

func normalizeMoonshotK26Thinking(body map[string]json.RawMessage) (bool, error) {
	raw, present := body["thinking"]
	if !present {
		return false, nil
	}

	fields, thinkingType, err := decodeMoonshotThinking(raw)
	if err != nil {
		return false, err
	}

	hasKeepAll, hasNullKeep, controlsErr := validateMoonshotThinkingControls(fields)
	if controlsErr != nil {
		return false, fmt.Errorf("moonshot_k2_6_thinking_controls_unsupported")
	}

	switch thinkingType {
	case "enabled", "disabled":
		if hasNullKeep {
			body["thinking"] = json.RawMessage(fmt.Sprintf(`{"type":%q}`, thinkingType))

			return true, nil
		}

		return false, nil
	case "adaptive":
		body["thinking"] = canonicalMoonshotEnabledThinking(hasKeepAll)

		return true, nil
	default:
		return false, fmt.Errorf("moonshot_invalid_thinking_type: %s", thinkingType)
	}
}

func validateMoonshotThinkingControls(
	fields map[string]json.RawMessage,
) (bool, bool, error) {
	for field := range fields {
		if field != "type" && field != "keep" {
			return false, false, fmt.Errorf("unsupported thinking control %q", field)
		}
	}

	rawKeep, hasKeep := fields["keep"]
	if !hasKeep {
		return false, false, nil
	}

	if moonshotJSONNull(rawKeep) {
		return false, true, nil
	}

	var keep string
	if json.Unmarshal(rawKeep, &keep) != nil || !strings.EqualFold(strings.TrimSpace(keep), "all") {
		return false, false, fmt.Errorf("unsupported thinking.keep")
	}

	return true, false, nil
}

func canonicalMoonshotEnabledThinking(keepAll bool) json.RawMessage {
	if keepAll {
		return json.RawMessage(`{"type":"enabled","keep":"all"}`)
	}

	return json.RawMessage(`{"type":"enabled"}`)
}

func decodeMoonshotThinking(raw json.RawMessage) (map[string]json.RawMessage, string, error) {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil {
		return nil, "", fmt.Errorf("moonshot_invalid_thinking")
	}

	var thinkingType string
	if json.Unmarshal(fields["type"], &thinkingType) != nil || strings.TrimSpace(thinkingType) == "" {
		return nil, "", fmt.Errorf("moonshot_invalid_thinking")
	}

	return fields, strings.ToLower(strings.TrimSpace(thinkingType)), nil
}

func deleteMoonshotParameter(body map[string]json.RawMessage, parameter string) bool {
	if _, present := body[parameter]; !present {
		return false
	}

	delete(body, parameter)

	return true
}

func moonshotJSONNull(raw json.RawMessage) bool {
	return strings.EqualFold(strings.TrimSpace(string(raw)), "null")
}

func validateMoonshotChatBodyDecoded(
	body map[string]json.RawMessage,
	model string,
	profile platform.ModelRuntimeProfile,
) error {
	for _, parameter := range profile.UnsupportedParameters {
		if contractsopenai.IsModelDecodingPassThroughParameter(parameter) {
			continue
		}

		if raw, present := body[parameter]; present && !moonshotJSONNull(raw) {
			return fmt.Errorf("moonshot_unsupported_parameter: %s", parameter)
		}
	}

	if err := validateMoonshotMessages(body["messages"], profile); err != nil {
		return err
	}

	normalizedModel := strings.ToLower(strings.TrimSpace(model))
	if err := validateMoonshotThinking(body, normalizedModel, profile); err != nil {
		return err
	}

	if err := validateMoonshotFixedCommon(body, normalizedModel); err != nil {
		return err
	}

	if err := validateMoonshotDecodingControls(body, normalizedModel); err != nil {
		return err
	}

	if err := validateMoonshotOutputLimit(body, profile.Limits.MaxOutputTokens); err != nil {
		return err
	}

	return validateMoonshotToolChoice(body, normalizedModel)
}

func validateMoonshotDecodingControls(body map[string]json.RawMessage, model string) error {
	if !moonshotModelUsesFixedControls(model) {
		return nil
	}

	for _, parameter := range []string{
		"top_k", "seed", "random_seed", "logit_bias", "parallel_tool_calls",
	} {
		if raw, present := body[parameter]; present && !moonshotJSONNull(raw) {
			return fmt.Errorf("moonshot_unsupported_parameter: %s", parameter)
		}
	}

	logprobsEnabled := false
	if raw, present := body["logprobs"]; present && !moonshotJSONNull(raw) {
		if json.Unmarshal(raw, &logprobsEnabled) != nil {
			return fmt.Errorf("moonshot_invalid_parameter: logprobs")
		}
	}

	if raw, present := body["top_logprobs"]; present && !moonshotJSONNull(raw) {
		var value int
		if json.Unmarshal(raw, &value) != nil || value < 0 || value > 20 {
			return fmt.Errorf("moonshot_invalid_parameter: top_logprobs")
		}

		if !logprobsEnabled {
			return fmt.Errorf("moonshot_top_logprobs_requires_logprobs")
		}
	}

	if raw, present := body["stop"]; present && !moonshotJSONNull(raw) {
		if err := validateMoonshotStop(raw); err != nil {
			return err
		}
	}

	return nil
}

func validateMoonshotStop(raw json.RawMessage) error {
	var single string
	if json.Unmarshal(raw, &single) == nil {
		if len([]byte(single)) > 32 {
			return fmt.Errorf("moonshot_invalid_parameter: stop")
		}

		return nil
	}

	var multiple []string
	if json.Unmarshal(raw, &multiple) != nil || len(multiple) > 5 {
		return fmt.Errorf("moonshot_invalid_parameter: stop")
	}

	for _, value := range multiple {
		if len([]byte(value)) > 32 {
			return fmt.Errorf("moonshot_invalid_parameter: stop")
		}
	}

	return nil
}

func validateMoonshotToolChoice(body map[string]json.RawMessage, model string) error {
	if !moonshotModelUsesFixedControls(model) {
		return nil
	}

	raw, present := body["tool_choice"]
	if !present || moonshotJSONNull(raw) {
		return nil
	}

	var choice string
	if json.Unmarshal(raw, &choice) == nil {
		switch strings.ToLower(strings.TrimSpace(choice)) {
		case "auto", "none":
			return nil
		case "required":
			if moonshotModelDisallowsRequiredToolChoice(model) {
				return fmt.Errorf("moonshot_tool_choice_required_unsupported")
			}

			return nil
		default:
			return fmt.Errorf("moonshot_invalid_tool_choice")
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
		return fmt.Errorf("moonshot_invalid_tool_choice")
	}

	if moonshotModelAlwaysThinks(model) ||
		(strings.EqualFold(strings.TrimSpace(model), "kimi-k2.6") && !moonshotThinkingDisabled(body)) {
		return fmt.Errorf("moonshot_specific_tool_choice_incompatible_with_thinking")
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

func moonshotThinkingDisabled(body map[string]json.RawMessage) bool {
	raw, present := body["thinking"]
	if !present {
		return false
	}

	_, thinkingType, err := decodeMoonshotThinking(raw)

	return err == nil && thinkingType == "disabled"
}

func validateMoonshotOutputLimit(body map[string]json.RawMessage, limit int) error {
	for _, parameter := range []string{"max_tokens", "max_completion_tokens"} {
		raw := body[parameter]
		if len(raw) == 0 {
			continue
		}

		var value int
		if json.Unmarshal(raw, &value) != nil || value < 1 {
			return fmt.Errorf("moonshot_invalid_parameter: %s", parameter)
		}

		if limit > 0 && value > limit {
			return fmt.Errorf("moonshot_parameter_exceeds_model_limit: %s", parameter)
		}
	}

	return nil
}

func validateMoonshotMessages(raw json.RawMessage, profile platform.ModelRuntimeProfile) error {
	var messages []struct {
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
	}
	if json.Unmarshal(raw, &messages) != nil || len(messages) == 0 {
		return fmt.Errorf("moonshot_invalid_messages")
	}

	for _, message := range messages {
		switch strings.ToLower(strings.TrimSpace(message.Role)) {
		case "system", "user", "assistant", "tool":
		default:
			return fmt.Errorf("moonshot_unsupported_message_role: %s", message.Role)
		}

		var parts []struct {
			Type string `json:"type"`
		}
		if json.Unmarshal(message.Content, &parts) != nil {
			continue
		}

		for _, part := range parts {
			modality := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(part.Type)), "_url")
			if modality != "text" && !profile.AllowsInputModality(modality) {
				return fmt.Errorf("moonshot_unsupported_input_modality: %s", modality)
			}
		}
	}

	return nil
}

func validateMoonshotThinking(
	body map[string]json.RawMessage,
	model string,
	profile platform.ModelRuntimeProfile,
) error {
	if raw := body["reasoning_effort"]; len(raw) > 0 {
		var effort string
		if json.Unmarshal(raw, &effort) != nil || !profile.AllowsEffort(effort) {
			return fmt.Errorf("moonshot_unsupported_reasoning_effort: %s", effort)
		}
	}

	rawThinking, hasThinking := body["thinking"]
	if model == "kimi-k3" {
		if hasThinking {
			return fmt.Errorf("moonshot_k3_thinking_parameter_unsupported")
		}

		return nil
	}

	if !hasThinking {
		return nil
	}

	var thinking struct {
		Type string `json:"type"`
	}
	if json.Unmarshal(rawThinking, &thinking) != nil {
		return fmt.Errorf("moonshot_invalid_thinking")
	}

	typeValue := strings.ToLower(strings.TrimSpace(thinking.Type))
	if model == "kimi-k2.7-code" || model == "kimi-k2.7-code-highspeed" {
		if typeValue != "enabled" && typeValue != "adaptive" {
			return fmt.Errorf("moonshot_k2_7_thinking_always_enabled")
		}

		return nil
	}

	mode := "adaptive"
	if typeValue == "disabled" {
		mode = "disabled"
	} else if typeValue != "enabled" && typeValue != "adaptive" {
		return fmt.Errorf("moonshot_invalid_thinking_type: %s", thinking.Type)
	}

	if !profile.AllowsThinkingMode(mode) {
		return fmt.Errorf("moonshot_unsupported_thinking_mode: %s", thinking.Type)
	}

	return nil
}

func validateMoonshotFixedCommon(body map[string]json.RawMessage, model string) error {
	if !moonshotModelUsesFixedControls(model) {
		return nil
	}

	if raw := body["n"]; len(raw) > 0 {
		if !moonshotJSONNull(raw) {
			var value int
			if json.Unmarshal(raw, &value) != nil || value != 1 {
				return fmt.Errorf("moonshot_n_must_equal_one")
			}
		}
	}

	for _, parameter := range []string{"presence_penalty", "frequency_penalty"} {
		raw := body[parameter]
		if len(raw) == 0 || moonshotJSONNull(raw) {
			continue
		}

		var value float64
		if json.Unmarshal(raw, &value) != nil || value != 0 {
			return fmt.Errorf("moonshot_%s_must_equal_zero", parameter)
		}
	}

	return nil
}

func moonshotModelUsesFixedControls(model string) bool {
	switch strings.ToLower(strings.TrimSpace(model)) {
	case "kimi-k3", "kimi-k2.7-code", "kimi-k2.7-code-highspeed", "kimi-k2.6":
		return true
	default:
		return false
	}
}
