package server

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

const (
	zaiDefaultBaseURL = "https://api.z.ai/api/paas/v4"
	zaiProviderName   = "zai"
)

var zaiModels = map[string]string{
	"glm-5.2":        "chat_completions",
	"glm-4.7-flashx": "chat_completions",
	"glm-5v-turbo":   "chat_completions",
	"glm-5.3-flash":  "chat_completions",
	"glm-5.3":        "chat_completions",
	"glm-ocr":        "layout_parsing",
}

var zaiFunctionNamePattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

type zaiAdapter struct{}

func newZAIAdapter() providerAdapter {
	return &zaiAdapter{}
}

func (a *zaiAdapter) Dispatch(
	ctx context.Context,
	client *http.Client,
	payload platform.TunnelRequest,
) (*adapterResponse, error) {
	endpoint, body, err := validateZAIDispatch(payload)
	if err != nil {
		return nil, err
	}

	base := strings.TrimRight(strings.TrimSpace(payload.Deployment.APIBase), "/")
	if base == "" {
		base = zaiDefaultBaseURL
	}

	operationPath := "/chat/completions"
	if endpoint == "layout_parsing" {
		operationPath = "/layout_parsing"
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		base+operationPath,
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("zai_request_build: %w", err)
	}

	apiKey := deploymentAPIKey(payload.Deployment)
	if apiKey == "" {
		return nil, fmt.Errorf("zai_api_key_unavailable")
	}

	request.Header.Set("Authorization", "Bearer "+apiKey)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	if payload.Stream {
		request.Header.Set("Accept", "text/event-stream")
	}

	if client == nil {
		client = &http.Client{}
	}

	response, err := client.Do(request) //nolint:bodyclose // The server handler owns and closes the transferred body.
	if err != nil {
		return nil, err
	}

	return &adapterResponse{
		StatusCode: response.StatusCode,
		Headers:    response.Header,
		Body:       response.Body,
	}, nil
}

func validateZAIDispatch(payload platform.TunnelRequest) (string, []byte, error) {
	if err := validateZAIBaseURL(payload.Deployment.APIBase); err != nil {
		return "", nil, err
	}

	var body map[string]json.RawMessage
	if err := json.Unmarshal(payload.Body, &body); err != nil || body == nil {
		return "", nil, fmt.Errorf("zai_invalid_json")
	}

	model := strings.ToLower(strings.TrimSpace(payload.Deployment.Model))

	var requestedModel string
	if json.Unmarshal(body["model"], &requestedModel) == nil && strings.TrimSpace(requestedModel) != "" {
		model = strings.ToLower(strings.TrimSpace(requestedModel))
	}

	endpoint, supported := zaiModels[model]
	if !supported {
		return "", nil, fmt.Errorf("zai_unsupported_model: %s", model)
	}

	expectedOperation := platform.ProviderOperationChatCompletions
	if endpoint == "layout_parsing" {
		expectedOperation = platform.ProviderOperationZAILayoutParsing
	}

	if payload.Operation != expectedOperation {
		return "", nil, fmt.Errorf("zai_unsupported_operation: %s", payload.Operation)
	}

	binding, ok := platform.FindModelRuntimeProfile(payload.ModelRuntimeProfiles, model)
	if !ok || binding.Validate() != nil ||
		!strings.EqualFold(binding.Provider, zaiProviderName) ||
		!binding.Profile.AllowsEndpoint(endpoint) {
		return "", nil, fmt.Errorf("zai_runtime_profile_missing: %s", model)
	}

	transport := "http"
	if payload.Stream {
		transport = "sse"
	}

	if !binding.Profile.AllowsTransport(transport) {
		return "", nil, fmt.Errorf("zai_unsupported_transport: %s", transport)
	}

	for parameter := range body {
		if !binding.Profile.AllowsParameter(endpoint, parameter) {
			return "", nil, fmt.Errorf("zai_unsupported_parameter: %s", parameter)
		}
	}

	if err := validateZAIRequestValues(model, endpoint, body, binding.Profile); err != nil {
		return "", nil, err
	}

	if endpoint == "chat_completions" {
		normalizeZAIReasoningEffort(model, binding.Profile, body)
	}

	normalizeZAIUserID(payload, body, binding.Profile.AllowsParameter(endpoint, "user_id"))

	encoded, err := json.Marshal(body)
	if err != nil {
		return "", nil, fmt.Errorf("zai_request_encode: %w", err)
	}

	return endpoint, encoded, nil
}

func normalizeZAIUserID(
	payload platform.TunnelRequest,
	body map[string]json.RawMessage,
	allowed bool,
) {
	tenantID := strings.TrimSpace(payload.TenantID)

	apiKey := deploymentAPIKey(payload.Deployment)
	if !allowed || tenantID == "" || apiKey == "" {
		delete(body, "user_id")

		return
	}

	var clientUserID string
	if raw, exists := body["user_id"]; exists {
		if err := json.Unmarshal(raw, &clientUserID); err != nil {
			clientUserID = ""
		}
	}

	mac := hmac.New(sha256.New, []byte(apiKey))
	_, _ = mac.Write([]byte(tenantID))
	_, _ = mac.Write([]byte{0})
	_, _ = mac.Write([]byte(payload.Deployment.ProviderAccountID))
	_, _ = mac.Write([]byte{0})
	_, _ = mac.Write([]byte(clientUserID))
	body["user_id"] = json.RawMessage(fmt.Sprintf("%q", "gdr_"+fmt.Sprintf("%x", mac.Sum(nil))[:60]))
}

// validateZAIBaseURL allows a BYOK tenant to point z.ai traffic at a
// compatible gateway instead of the official host. It only enforces the
// general BYOK api_base safety rules (see validateBYOKAPIBase); it does not
// pin the host or path to the canonical https://api.z.ai/api/paas/v4.
func validateZAIBaseURL(raw string) error {
	if err := validateBYOKAPIBase(raw); err != nil {
		return fmt.Errorf("zai_invalid_api_base")
	}

	return nil
}

func validateZAIRequestValues(
	model string,
	endpoint string,
	body map[string]json.RawMessage,
	profile platform.ModelRuntimeProfile,
) error {
	if err := validateZAIIdentifierFields(body); err != nil {
		return err
	}

	if endpoint == "layout_parsing" {
		if _, ok := body["file"]; !ok {
			return fmt.Errorf("zai_file_required")
		}

		var start, end int

		startSet := json.Unmarshal(body["start_page_id"], &start) == nil

		endSet := json.Unmarshal(body["end_page_id"], &end) == nil
		if (startSet && start < 1) || (endSet && end < 1) ||
			(startSet && endSet && (end < start || end-start >= 30)) {
			return fmt.Errorf("zai_ocr_page_range_invalid")
		}

		return nil
	}

	if _, ok := body["messages"]; !ok {
		return fmt.Errorf("zai_messages_required")
	}

	if err := validateZAIThinkingValue(profile, body); err != nil {
		return err
	}

	for field, bounds := range map[string][2]float64{
		"temperature": {0, 1},
		"top_p":       {0.01, 1},
	} {
		if raw, ok := body[field]; ok {
			var value float64
			if json.Unmarshal(raw, &value) != nil || value < bounds[0] || value > bounds[1] {
				return fmt.Errorf("zai_%s_invalid", field)
			}
		}
	}

	for _, field := range []string{"do_sample", "tool_stream"} {
		if raw, ok := body[field]; ok {
			var value bool
			if json.Unmarshal(raw, &value) != nil {
				return fmt.Errorf("zai_%s_invalid", field)
			}
		}
	}

	if err := validateZAIToolStream(body); err != nil {
		return err
	}

	if err := validateZAIResponseFormat(body); err != nil {
		return err
	}

	if err := validateZAITools(body); err != nil {
		return err
	}

	var stop []string
	if raw, ok := body["stop"]; ok {
		if err := json.Unmarshal(raw, &stop); err != nil || len(stop) > 1 {
			return fmt.Errorf("zai_stop_limit_exceeded")
		}
	}

	if model == "glm-5v-turbo" {
		if _, ok := body["tools"]; ok {
			return fmt.Errorf("zai_vision_tools_not_verified")
		}
	}

	return nil
}

// validateZAIThinkingValue enforces the model's runtime-profile thinking
// contract at the adapter layer, independently of the control-plane request
// validator. Most Z.ai chat models allow toggling thinking.type between
// "enabled" and "disabled", but some (for example glm-5.3-flash) always run
// with thinking enabled and must reject an explicit "disabled" request
// before it can reach the upstream provider.
func validateZAIThinkingValue(profile platform.ModelRuntimeProfile, body map[string]json.RawMessage) error {
	raw, ok := body["thinking"]
	if !ok {
		return nil
	}

	var thinking struct {
		Type string `json:"type"`
	}
	if json.Unmarshal(raw, &thinking) != nil {
		return fmt.Errorf("zai_thinking_invalid")
	}

	switch strings.ToLower(strings.TrimSpace(thinking.Type)) {
	case "":
		// No explicit directive; nothing to enforce here.
	case "enabled":
		if !profile.AllowsThinkingMode("adaptive") {
			return fmt.Errorf("zai_thinking_enabled_unsupported")
		}
	case "disabled":
		if !profile.AllowsThinkingMode("disabled") {
			return fmt.Errorf("zai_thinking_disabled_unsupported")
		}
	default:
		return fmt.Errorf("zai_thinking_invalid")
	}

	return nil
}

func validateZAIIdentifierFields(body map[string]json.RawMessage) error {
	for field, bounds := range map[string][2]int{"request_id": {6, 64}, "user_id": {6, 128}} {
		raw, ok := body[field]
		if !ok {
			continue
		}

		var value string
		if json.Unmarshal(raw, &value) != nil || len(value) < bounds[0] || len(value) > bounds[1] {
			return fmt.Errorf("zai_%s_invalid", field)
		}
	}

	return nil
}

func validateZAIToolStream(body map[string]json.RawMessage) error {
	raw, ok := body["tool_stream"]
	if !ok {
		return nil
	}

	var enabled bool
	if err := json.Unmarshal(raw, &enabled); err != nil {
		return fmt.Errorf("zai_tool_stream_invalid")
	}

	if !enabled {
		return nil
	}

	var stream bool
	if json.Unmarshal(body["stream"], &stream) != nil || !stream {
		return fmt.Errorf("zai_tool_stream_requires_stream")
	}

	var tools []json.RawMessage
	if json.Unmarshal(body["tools"], &tools) != nil || len(tools) == 0 {
		return fmt.Errorf("zai_tool_stream_requires_tools")
	}

	return nil
}

func validateZAIResponseFormat(body map[string]json.RawMessage) error {
	raw, ok := body["response_format"]
	if !ok {
		return nil
	}

	var format map[string]json.RawMessage
	if json.Unmarshal(raw, &format) != nil || len(format) != 1 {
		return fmt.Errorf("zai_response_format_invalid")
	}

	var formatType string
	if json.Unmarshal(format["type"], &formatType) != nil ||
		(formatType != "text" && formatType != "json_object") {
		return fmt.Errorf("zai_response_format_invalid")
	}

	return nil
}

func validateZAITools(body map[string]json.RawMessage) error {
	raw, ok := body["tools"]
	if !ok {
		return nil
	}

	var tools []map[string]json.RawMessage
	if json.Unmarshal(raw, &tools) != nil || len(tools) > 128 {
		return fmt.Errorf("zai_tools_invalid")
	}

	for index, tool := range tools {
		var toolType string
		if len(tool) != 2 || json.Unmarshal(tool["type"], &toolType) != nil || toolType != "function" {
			return fmt.Errorf("zai_tool_invalid: %d", index)
		}

		var function map[string]json.RawMessage
		if json.Unmarshal(tool["function"], &function) != nil || len(function) != 3 {
			return fmt.Errorf("zai_tool_invalid: %d", index)
		}

		var (
			name, description string
			parameters        map[string]interface{}
		)
		if json.Unmarshal(function["name"], &name) != nil ||
			!zaiFunctionNamePattern.MatchString(name) ||
			json.Unmarshal(function["description"], &description) != nil ||
			strings.TrimSpace(description) == "" ||
			json.Unmarshal(function["parameters"], &parameters) != nil || parameters == nil {
			return fmt.Errorf("zai_tool_invalid: %d", index)
		}
	}

	return nil
}

func normalizeZAIReasoningEffort(model string, profile platform.ModelRuntimeProfile, body map[string]json.RawMessage) {
	raw, exists := body["reasoning_effort"]
	if !exists {
		return
	}

	var effort string
	if json.Unmarshal(raw, &effort) != nil {
		return
	}

	switch strings.ToLower(strings.TrimSpace(effort)) {
	case "none", "minimal":
		if !profile.AllowsThinkingMode("disabled") {
			// This model's thinking can never be disabled; leave any
			// explicit thinking control untouched and drop only the
			// unsupported low-effort alias.
			delete(body, "reasoning_effort")

			return
		}

		body["thinking"] = json.RawMessage(`{"type":"disabled"}`)
		delete(body, "reasoning_effort")
	case "low":
		// glm-5.3's documented reasoning_effort enum is low/high/max, with
		// "low" a genuinely distinct, natively supported tier. Every other
		// zai chat model collapses "low" into "high" instead.
		if model != "glm-5.3" {
			body["reasoning_effort"] = json.RawMessage(`"high"`)
		}
	case "medium":
		body["reasoning_effort"] = json.RawMessage(`"high"`)
	case "xhigh":
		body["reasoning_effort"] = json.RawMessage(`"max"`)
	}
}
