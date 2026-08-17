package proxy

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
)

var zaiFunctionNamePattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

func validateZAIChatRuntimeProfile(
	req contractsopenai.ChatCompletionRequest,
	binding *platform.ModelRuntimeProfileBinding,
) error {
	if binding == nil || binding.Validate() != nil || !isZAIProvider(binding.Provider) {
		return fmt.Errorf("z.ai runtime profile is unavailable")
	}

	profile := binding.Profile
	if !profile.AllowsEndpoint("chat_completions") {
		return fmt.Errorf("z.ai model does not support chat completions")
	}

	transport := "http"
	if req.Stream {
		transport = "sse"
	}

	if !profile.AllowsTransport(transport) {
		return fmt.Errorf("z.ai model does not support %s transport", transport)
	}

	if req.MaxTokens != nil && profile.Limits.MaxOutputTokens > 0 &&
		*req.MaxTokens > profile.Limits.MaxOutputTokens {
		return fmt.Errorf("max_tokens exceeds model limit %d", profile.Limits.MaxOutputTokens)
	}

	if req.Temperature != nil && (*req.Temperature < 0 || *req.Temperature > 1) {
		return fmt.Errorf("temperature must be between 0 and 1 for Z.ai")
	}

	if req.TopP != nil && (*req.TopP < 0.01 || *req.TopP > 1) {
		return fmt.Errorf("top_p must be between 0.01 and 1 for Z.ai")
	}

	if req.Thinking != nil && req.Thinking.Type != "" &&
		!strings.EqualFold(req.Thinking.Type, "enabled") &&
		!strings.EqualFold(req.Thinking.Type, "disabled") {
		return fmt.Errorf("thinking.type must be enabled or disabled for Z.ai")
	}

	if strings.TrimSpace(req.ReasoningEffort) != "" && !profile.AllowsEffort(req.ReasoningEffort) {
		return fmt.Errorf("reasoning_effort %q is not supported by this Z.ai model", req.ReasoningEffort)
	}

	if err := validateZAIToolsAndResponseFormat(req, profile); err != nil {
		return err
	}

	if err := validateZAIRawChatControls(req); err != nil {
		return err
	}

	if err := validateZAIMessageContract(req, profile); err != nil {
		return err
	}

	if chatRequestHasImage(req) && !profile.AllowsInputModality("image") {
		return fmt.Errorf("this Z.ai model does not support image input")
	}

	if chatRequestHasVideo(req) && !profile.AllowsInputModality("video") {
		return fmt.Errorf("this Z.ai model does not support video input")
	}

	if chatRequestHasFile(req) && !profile.AllowsInputModality("document") {
		return fmt.Errorf("this Z.ai model does not support document input")
	}

	if chatRequestHasFile(req) && (chatRequestHasImage(req) || chatRequestHasVideo(req)) {
		return fmt.Errorf("z.ai file_url cannot be combined with image or video input")
	}

	images, videos, files := 0, 0, 0
	for _, message := range req.Messages {
		images += contractsopenai.ChatContentImageCount(message.Content)
		videos += contractsopenai.ChatContentVideoCount(message.Content)
		files += contractsopenai.ChatContentFileCount(message.Content)
	}

	if images > 150 {
		return fmt.Errorf("z.ai supports at most 150 images per request")
	}

	if videos > 2 {
		return fmt.Errorf("z.ai supports at most 2 videos per request")
	}

	if files > 50 {
		return fmt.Errorf("z.ai supports at most 50 documents per request")
	}

	if zaiStopSequenceCount(req.Stop) > 1 {
		return fmt.Errorf("z.ai stop accepts at most one sequence until upstream multi-stop support is verified")
	}

	for parameter := range req.AdditionalFields {
		if !profile.AllowsParameter("chat_completions", parameter) {
			return fmt.Errorf("parameter %q is not supported by this Z.ai model", parameter)
		}
	}

	return nil
}

func validateZAIToolsAndResponseFormat(
	req contractsopenai.ChatCompletionRequest,
	profile platform.ModelRuntimeProfile,
) error {
	if len(req.Tools) > 0 && !profile.Tools.Custom {
		return fmt.Errorf("function tools are not verified for this Z.ai model")
	}

	if len(req.Tools) > 128 {
		return fmt.Errorf("z.ai supports at most 128 function tools")
	}

	for index, tool := range req.Tools {
		if tool.Type != "function" || tool.Function == nil {
			return fmt.Errorf("tools[%d] must be a function tool for Z.ai", index)
		}

		if !zaiFunctionNamePattern.MatchString(tool.Function.Name) {
			return fmt.Errorf("tools[%d].function.name is invalid for Z.ai", index)
		}

		if strings.TrimSpace(tool.Function.Description) == "" {
			return fmt.Errorf("tools[%d].function.description is required for Z.ai", index)
		}

		if tool.Function.Parameters == nil {
			return fmt.Errorf("tools[%d].function.parameters is required for Z.ai", index)
		}
	}

	if req.ToolChoice != nil {
		choice, ok := req.ToolChoice.(string)
		if !ok || !strings.EqualFold(choice, "auto") {
			return fmt.Errorf("z.ai tool_choice currently supports only auto")
		}
	}

	if req.ResponseFormat == nil {
		return nil
	}

	format, ok := req.ResponseFormat.(map[string]interface{})
	if !ok {
		return fmt.Errorf("response_format must be an object for Z.ai")
	}

	formatType, ok := format["type"].(string)
	if !ok || (formatType != "text" && formatType != "json_object") || len(format) != 1 {
		return fmt.Errorf("z.ai response_format supports only type text or json_object")
	}

	return nil
}

func validateZAIMessageContract(
	req contractsopenai.ChatCompletionRequest,
	profile platform.ModelRuntimeProfile,
) error {
	vision := profile.AllowsInputModality("image") || profile.AllowsInputModality("video") ||
		profile.AllowsInputModality("document")
	hasUser := false

	for messageIndex, message := range req.Messages {
		role := strings.ToLower(strings.TrimSpace(message.Role))
		if role == "user" {
			hasUser = true
		}

		if vision && role != "system" && role != "user" && role != "assistant" {
			return fmt.Errorf("message[%d].role is not supported by this Z.ai vision model", messageIndex)
		}

		parts, multipart, err := contractsopenai.ChatContentParts(message.Content)
		if err != nil {
			return fmt.Errorf("message[%d].content: %w", messageIndex, err)
		}

		if !multipart {
			continue
		}

		if !vision {
			return fmt.Errorf("message[%d].content must be text for this Z.ai model", messageIndex)
		}

		if role != "user" {
			return fmt.Errorf("multimodal content is allowed only in Z.ai user messages")
		}

		for partIndex, part := range parts {
			if err := validateZAIContentPart(part); err != nil {
				return fmt.Errorf("message[%d].content[%d]: %w", messageIndex, partIndex, err)
			}
		}
	}

	if !hasUser {
		return fmt.Errorf("z.ai requests require at least one user message")
	}

	return nil
}

func validateZAIContentPart(part contractsopenai.ChatContentPart) error {
	switch part.Type {
	case "text":
		return nil
	case "image_url":
		if part.ImageURL == nil || strings.TrimSpace(part.ImageURL.Detail) != "" {
			return fmt.Errorf("image_url.detail is not supported by Z.ai")
		}

		rawURL := strings.TrimSpace(part.ImageURL.URL)
		if strings.HasPrefix(strings.ToLower(rawURL), "data:") {
			mediaType, encoded, err := contractsopenai.ParseImageDataURL(rawURL)
			if err != nil ||
				(mediaType != "image/png" && mediaType != "image/jpeg" && mediaType != "image/jpg") {
				return fmt.Errorf("image must be PNG or JPEG base64 data")
			}

			decoded, err := base64.StdEncoding.DecodeString(encoded)
			if err != nil || len(decoded) > 5*1024*1024 {
				return fmt.Errorf("image base64 data exceeds the 5 MB Z.ai limit or is invalid")
			}

			return nil
		}

		return validateZAIHTTPSURL(rawURL, "image_url")
	case "video_url":
		if part.VideoURL == nil {
			return fmt.Errorf("video_url is required")
		}

		return validateZAIHTTPSURL(part.VideoURL.URL, "video_url")
	case "file_url":
		if part.FileURL == nil {
			return fmt.Errorf("file_url is required")
		}

		return validateZAIHTTPSURL(part.FileURL.URL, "file_url")
	default:
		return fmt.Errorf("content type %q is not supported by Z.ai", part.Type)
	}
}

func validateZAIHTTPSURL(value string, field string) error {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil {
		return fmt.Errorf("%s must be an HTTPS URL for Z.ai", field)
	}

	return nil
}

func validateZAIRawChatControls(req contractsopenai.ChatCompletionRequest) error {
	for _, field := range []string{"do_sample", "tool_stream"} {
		raw, ok := req.AdditionalFields[field]
		if !ok {
			continue
		}

		var value bool
		if err := json.Unmarshal(raw, &value); err != nil {
			return fmt.Errorf("%s must be a boolean for Z.ai", field)
		}

		if field == "tool_stream" && value && (!req.Stream || len(req.Tools) == 0) {
			return fmt.Errorf("z.ai tool_stream requires stream=true and at least one tool")
		}
	}

	for field, bounds := range map[string][2]int{"request_id": {6, 64}, "user_id": {6, 128}} {
		raw, ok := req.AdditionalFields[field]
		if !ok {
			continue
		}

		var value string
		if err := json.Unmarshal(raw, &value); err != nil || len(value) < bounds[0] || len(value) > bounds[1] {
			return fmt.Errorf("%s must be a string between %d and %d characters for Z.ai", field, bounds[0], bounds[1])
		}
	}

	return nil
}

func zaiStopSequenceCount(value interface{}) int {
	switch typed := value.(type) {
	case []string:
		return len(typed)
	case []interface{}:
		return len(typed)
	case string:
		if typed != "" {
			return 1
		}
	}

	return 0
}
