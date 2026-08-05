package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

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
		model := moonshotRequestModel(payload)

		binding, err := validateMoonshotRuntimeProfile(payload, model, endpoint)
		if err != nil {
			return nil, err
		}

		if err := validateMoonshotChatBody(payload.Body, model, binding.Profile); err != nil {
			return nil, err
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

func moonshotRequestModel(payload platform.TunnelRequest) string {
	model := strings.TrimSpace(payload.Deployment.Model)

	var request struct {
		Model string `json:"model"`
	}
	if json.Unmarshal(payload.Body, &request) == nil && strings.TrimSpace(request.Model) != "" {
		model = strings.TrimSpace(request.Model)
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

func validateMoonshotChatBody(
	bodyBytes []byte,
	model string,
	profile platform.ModelRuntimeProfile,
) error {
	var body map[string]json.RawMessage
	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		return fmt.Errorf("moonshot_invalid_json: %w", err)
	}

	for _, parameter := range profile.UnsupportedParameters {
		if _, present := body[parameter]; present {
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

	if err := validateMoonshotSampling(body, normalizedModel); err != nil {
		return err
	}

	if err := validateMoonshotOutputLimit(body, profile.Limits.MaxOutputTokens); err != nil {
		return err
	}

	if raw := body["tool_choice"]; len(raw) > 0 && normalizedModel != "kimi-k3" {
		var value string
		if json.Unmarshal(raw, &value) == nil && strings.EqualFold(value, "required") {
			return fmt.Errorf("moonshot_tool_choice_required_unsupported")
		}
	}

	return nil
}

func validateMoonshotOutputLimit(body map[string]json.RawMessage, limit int) error {
	if limit <= 0 {
		return nil
	}

	for _, parameter := range []string{"max_tokens", "max_completion_tokens"} {
		raw := body[parameter]
		if len(raw) == 0 {
			continue
		}

		var value int
		if json.Unmarshal(raw, &value) != nil || value < 1 {
			return fmt.Errorf("moonshot_invalid_parameter: %s", parameter)
		}

		if value > limit {
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
		if typeValue != "enabled" {
			return fmt.Errorf("moonshot_k2_7_thinking_always_enabled")
		}

		return nil
	}

	mode := "adaptive"
	if typeValue == "disabled" {
		mode = "disabled"
	} else if typeValue != "enabled" {
		return fmt.Errorf("moonshot_invalid_thinking_type: %s", thinking.Type)
	}

	if !profile.AllowsThinkingMode(mode) {
		return fmt.Errorf("moonshot_unsupported_thinking_mode: %s", thinking.Type)
	}

	return nil
}

func validateMoonshotSampling(body map[string]json.RawMessage, model string) error {
	if model == "kimi-k3" {
		for _, key := range []string{"temperature", "top_p"} {
			if _, present := body[key]; present {
				return fmt.Errorf("moonshot_unsupported_parameter: %s", key)
			}
		}

		return validateMoonshotFixedCommon(body)
	}

	expectedTemperature := 1.0

	if model == "kimi-k2.6" {
		var thinking struct {
			Type string `json:"type"`
		}
		if json.Unmarshal(body["thinking"], &thinking) == nil && strings.EqualFold(thinking.Type, "disabled") {
			expectedTemperature = 0.6
		}
	}

	if err := requireMoonshotFloat(body, "temperature", expectedTemperature); err != nil {
		return err
	}

	if err := requireMoonshotFloat(body, "top_p", 0.95); err != nil {
		return err
	}

	return validateMoonshotFixedCommon(body)
}

func validateMoonshotFixedCommon(body map[string]json.RawMessage) error {
	if raw := body["n"]; len(raw) > 0 {
		var value int
		if json.Unmarshal(raw, &value) != nil || value != 1 {
			return fmt.Errorf("moonshot_n_must_equal_one")
		}
	}

	for _, key := range []string{"presence_penalty", "frequency_penalty"} {
		if raw := body[key]; len(raw) > 0 {
			var value float64
			if json.Unmarshal(raw, &value) != nil || value != 0 {
				return fmt.Errorf("moonshot_%s_must_equal_zero", key)
			}
		}
	}

	return nil
}

func requireMoonshotFloat(
	body map[string]json.RawMessage,
	key string,
	expected float64,
) error {
	raw := body[key]
	if len(raw) == 0 {
		return nil
	}

	var value float64
	if json.Unmarshal(raw, &value) != nil || math.Abs(value-expected) > 1e-9 {
		return fmt.Errorf("moonshot_%s_must_equal_%g", key, expected)
	}

	return nil
}
