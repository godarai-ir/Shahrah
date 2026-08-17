package openai

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ChatCompletionRequest mirrors OpenAI's chat completion request format.
type ChatCompletionRequest struct {
	Model               string                     `json:"model"`
	Messages            []ChatMessage              `json:"messages"`
	Stream              bool                       `json:"stream,omitempty"`
	StreamOptions       *StreamOptions             `json:"stream_options,omitempty"`
	Temperature         *float64                   `json:"temperature,omitempty"`
	TopP                *float64                   `json:"top_p,omitempty"`
	TopK                *int                       `json:"top_k,omitempty"`
	MaxTokens           *int                       `json:"max_tokens,omitempty"`
	MaxCompletionTokens *int                       `json:"max_completion_tokens,omitempty"`
	Stop                interface{}                `json:"stop,omitempty"`
	ReasoningEffort     string                     `json:"reasoning_effort,omitempty"`
	Thinking            *ThinkingConfig            `json:"thinking,omitempty"`
	ParallelToolCalls   *bool                      `json:"parallel_tool_calls,omitempty"`
	ResponseFormat      interface{}                `json:"response_format,omitempty"`
	OutputConfig        map[string]interface{}     `json:"output_config,omitempty"`
	Tools               []ToolDefinition           `json:"tools,omitempty"`
	ToolChoice          interface{}                `json:"tool_choice,omitempty"`
	CacheControl        map[string]interface{}     `json:"cache_control,omitempty"`
	ExtraBody           map[string]interface{}     `json:"extra_body,omitempty"`
	ServiceTier         string                     `json:"service_tier,omitempty"`
	Speed               string                     `json:"speed,omitempty"`
	InferenceGeo        string                     `json:"inference_geo,omitempty"`
	Metadata            map[string]interface{}     `json:"metadata,omitempty"`
	User                string                     `json:"user,omitempty"`
	PromptVersionFQN    string                     `json:"prompt_version_fqn,omitempty"`
	PromptVariables     map[string]string          `json:"prompt_variables,omitempty"`
	AdditionalFields    map[string]json.RawMessage `json:"-"`
}

// UnmarshalJSON retains fields added by OpenAI after this contract was
// compiled. Known fields remain strongly typed for policy and billing logic,
// while unknown fields are relayed losslessly by MarshalJSON.
func (r *ChatCompletionRequest) UnmarshalJSON(data []byte) error {
	type requestAlias ChatCompletionRequest

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}

	// Decode pass-through controls on a best-effort basis. Their original JSON
	// remains in AdditionalFields so provider-specific types and values can be
	// relayed without Shahrah becoming their semantic validator.
	decodedFields := make(map[string]json.RawMessage, len(fields))
	for key, value := range fields {
		if IsModelDecodingPassThroughParameter(key) {
			continue
		}

		decodedFields[key] = value
	}

	decodedJSON, err := json.Marshal(decodedFields)
	if err != nil {
		return err
	}

	var decoded requestAlias
	if err := json.Unmarshal(decodedJSON, &decoded); err != nil {
		return err
	}

	decodeOptionalFloat64(fields["temperature"], &decoded.Temperature)
	decodeOptionalFloat64(fields["top_p"], &decoded.TopP)
	decodeOptionalInt(fields["top_k"], &decoded.TopK)

	if raw := fields["stop"]; len(raw) > 0 {
		if err := json.Unmarshal(raw, &decoded.Stop); err != nil {
			decoded.Stop = nil
		}
	}

	*r = ChatCompletionRequest(decoded)
	r.AdditionalFields = fields

	return nil
}

// MarshalJSON overlays the current typed request on the original raw fields.
// This preserves future OpenAI parameters while still allowing GodarAI to
// rewrite routing fields and apply policy transformations.
func (r ChatCompletionRequest) MarshalJSON() ([]byte, error) {
	type requestAlias ChatCompletionRequest

	knownJSON, err := json.Marshal(requestAlias(r))
	if err != nil {
		return nil, err
	}

	var known map[string]json.RawMessage
	if err := json.Unmarshal(knownJSON, &known); err != nil {
		return nil, err
	}

	merged := make(map[string]json.RawMessage, len(r.AdditionalFields)+len(known))
	for key, value := range r.AdditionalFields {
		merged[key] = append(json.RawMessage(nil), value...)
	}

	for key, value := range known {
		if IsModelDecodingPassThroughParameter(key) {
			if _, present := merged[key]; present {
				continue
			}
		}

		merged[key] = value
	}

	delete(merged, "AdditionalFields")

	return json.Marshal(merged)
}

// IsModelDecodingPassThroughParameter reports whether a model-owned decoding
// control must be relayed without gateway-level semantic validation.
func IsModelDecodingPassThroughParameter(name string) bool {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "temperature",
		"top_p",
		"top_k",
		"presence_penalty",
		"frequency_penalty",
		"seed",
		"random_seed",
		"stop",
		"logit_bias",
		"logprobs",
		"top_logprobs":
		return true
	default:
		return false
	}
}

func decodeOptionalFloat64(raw json.RawMessage, target **float64) {
	if len(raw) == 0 {
		return
	}

	var value *float64
	if err := json.Unmarshal(raw, &value); err != nil {
		return
	}

	*target = value
}

func decodeOptionalInt(raw json.RawMessage, target **int) {
	if len(raw) == 0 {
		return
	}

	var value *int
	if err := json.Unmarshal(raw, &value); err != nil {
		return
	}

	*target = value
}

// StreamOptions configures optional fields emitted by compatible streaming providers.
type StreamOptions struct {
	IncludeUsage bool `json:"include_usage,omitempty"`
}

// ChatMessage represents a single message in a chat completion request.
type ChatMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content,omitempty"`
	Name    string      `json:"name,omitempty"`
	Refusal string      `json:"refusal,omitempty"`
	// ReasoningContent preserves OpenAI-compatible reasoning text returned by
	// providers that expose it outside the final answer.
	ReasoningContent string          `json:"reasoning_content,omitempty"`
	ThinkingBlocks   []ThinkingBlock `json:"thinking_blocks,omitempty"`
	ToolCalls        []ToolCall      `json:"tool_calls,omitempty"`
	ToolCallID       string          `json:"tool_call_id,omitempty"`
	Citations        []interface{}   `json:"citations,omitempty"`
	// ProviderContent preserves provider-native content blocks that do not
	// have an OpenAI equivalent, including Anthropic server-tool results.
	ProviderContent []map[string]interface{} `json:"provider_content,omitempty"`
	// FunctionCall is deprecated by OpenAI in favor of tool_calls, but retained
	// so OpenAI-compatible providers can round-trip older responses.
	FunctionCall *FunctionCall `json:"function_call,omitempty"`
}

// ThinkingConfig preserves provider thinking configuration. OpenAI-compatible
// callers typically use reasoning_effort; Anthropic-native routing maps either
// reasoning_effort or this object to the Anthropic thinking request field.
type ThinkingConfig struct {
	Type          string `json:"type,omitempty"`
	BudgetTokens  int    `json:"budget_tokens,omitempty"`
	Display       string `json:"display,omitempty"`
	ClearThinking *bool  `json:"clear_thinking,omitempty"`
}

// ThinkingBlock preserves provider reasoning blocks, including Anthropic
// signatures and redacted_thinking payloads, without normalizing unknown fields.
type ThinkingBlock map[string]interface{}

// ChatContentPart represents one OpenAI-compatible multimodal message part.
type ChatContentPart struct {
	Type       string          `json:"type"`
	Text       string          `json:"text,omitempty"`
	ImageURL   *ChatImageURL   `json:"image_url,omitempty"`
	VideoURL   *ChatVideoURL   `json:"video_url,omitempty"`
	File       *ChatFile       `json:"file,omitempty"`
	FileURL    *ChatFileURL    `json:"file_url,omitempty"`
	InputAudio *ChatInputAudio `json:"input_audio,omitempty"`
}

// ChatImageURL contains an image URL or data URL for multimodal requests.
type ChatImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

// ChatVideoURL contains a video URL or data URL for multimodal requests.
type ChatVideoURL struct {
	URL string `json:"url"`
}

// ChatFile contains an OpenAI-compatible file reference or inline file data.
type ChatFile struct {
	Filename string `json:"filename,omitempty"`
	FileData string `json:"file_data,omitempty"`
	FileID   string `json:"file_id,omitempty"`
}

// ChatFileURL contains a provider-retrievable document URL or data URL.
type ChatFileURL struct {
	URL string `json:"url"`
}

// ChatInputAudio contains base64 audio and its documented OpenAI-compatible
// format (for example wav or mp3).
type ChatInputAudio struct {
	Data   string `json:"data"`
	Format string `json:"format"`
}

// ValidateChatMessageContent checks the supported Shahrah chat content contract.
func ValidateChatMessageContent(content interface{}) error {
	_, _, err := ChatContentParts(content)
	return err
}

// ChatContentText extracts text from either string content or text parts.
func ChatContentText(content interface{}) string {
	switch v := content.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(v)
	case []ChatContentPart:
		parts := make([]string, 0, len(v))
		for _, part := range v {
			if part.Type == "text" && strings.TrimSpace(part.Text) != "" {
				parts = append(parts, strings.TrimSpace(part.Text))
			}
		}

		return strings.Join(parts, "\n")
	default:
		parts, ok, err := ChatContentParts(v)
		if !ok || err != nil {
			return ""
		}

		texts := make([]string, 0, len(parts))
		for _, part := range parts {
			if part.Type == "text" && strings.TrimSpace(part.Text) != "" {
				texts = append(texts, strings.TrimSpace(part.Text))
			}
		}

		return strings.Join(texts, "\n")
	}
}

// ChatContentImageCount returns the number of image_url parts in content.
func ChatContentImageCount(content interface{}) int {
	parts, ok, err := ChatContentParts(content)
	if !ok || err != nil {
		return 0
	}

	count := 0

	for _, part := range parts {
		if part.Type == "image_url" {
			count++
		}
	}

	return count
}

// ChatContentHasImage reports whether content contains any image_url part.
func ChatContentHasImage(content interface{}) bool {
	return ChatContentImageCount(content) > 0
}

// ChatContentVideoCount returns the number of video_url parts in content.
func ChatContentVideoCount(content interface{}) int {
	parts, ok, err := ChatContentParts(content)
	if !ok || err != nil {
		return 0
	}

	count := 0

	for _, part := range parts {
		if part.Type == "video_url" {
			count++
		}
	}

	return count
}

// ChatContentHasVideo reports whether content contains any video_url part.
func ChatContentHasVideo(content interface{}) bool {
	return ChatContentVideoCount(content) > 0
}

// ChatContentFileCount returns the number of file parts in content.
func ChatContentFileCount(content interface{}) int {
	parts, ok, err := ChatContentParts(content)
	if !ok || err != nil {
		return 0
	}

	count := 0

	for _, part := range parts {
		if part.Type == "file" || part.Type == "file_url" {
			count++
		}
	}

	return count
}

// ChatContentHasFile reports whether content contains any file part.
func ChatContentHasFile(content interface{}) bool {
	return ChatContentFileCount(content) > 0
}

// ChatContentInputAudioCount returns the number of input_audio parts in content.
func ChatContentInputAudioCount(content interface{}) int {
	parts, ok, err := ChatContentParts(content)
	if !ok || err != nil {
		return 0
	}

	count := 0

	for _, part := range parts {
		if part.Type == "input_audio" {
			count++
		}
	}

	return count
}

// ChatContentHasInputAudio reports whether content contains input audio.
func ChatContentHasInputAudio(content interface{}) bool {
	return ChatContentInputAudioCount(content) > 0
}

// ChatContentParts normalizes array content parts. The bool is false for plain string content.
func ChatContentParts(content interface{}) ([]ChatContentPart, bool, error) {
	switch v := content.(type) {
	case nil:
		return nil, false, nil
	case string:
		return nil, false, nil
	case []ChatContentPart:
		return validateChatContentParts(v)
	case []interface{}:
		parts := make([]ChatContentPart, 0, len(v))
		for i, raw := range v {
			part, err := chatContentPartFromValue(raw)
			if err != nil {
				return nil, true, fmt.Errorf("content part[%d]: %w", i, err)
			}

			parts = append(parts, part)
		}

		return validateChatContentParts(parts)
	default:
		return nil, false, fmt.Errorf("content must be a string or an array of content parts")
	}
}

func validateChatContentParts(parts []ChatContentPart) ([]ChatContentPart, bool, error) {
	for i, part := range parts {
		part.Type = strings.TrimSpace(part.Type)
		switch part.Type {
		case "text":
			if strings.TrimSpace(part.Text) == "" {
				return nil, true, fmt.Errorf("content part[%d].text is required", i)
			}
		case "image_url":
			if part.ImageURL == nil || strings.TrimSpace(part.ImageURL.URL) == "" {
				return nil, true, fmt.Errorf("content part[%d].image_url.url is required", i)
			}
		case "video_url":
			if part.VideoURL == nil || strings.TrimSpace(part.VideoURL.URL) == "" {
				return nil, true, fmt.Errorf("content part[%d].video_url.url is required", i)
			}
		case "file":
			if part.File == nil {
				return nil, true, fmt.Errorf("content part[%d].file is required", i)
			}

			hasFileData := strings.TrimSpace(part.File.FileData) != ""

			hasFileID := strings.TrimSpace(part.File.FileID) != ""
			if !hasFileData && !hasFileID {
				return nil, true, fmt.Errorf("content part[%d].file.file_data or file.file_id is required", i)
			}
		case "file_url":
			if part.FileURL == nil || strings.TrimSpace(part.FileURL.URL) == "" {
				return nil, true, fmt.Errorf("content part[%d].file_url.url is required", i)
			}
		case "input_audio":
			if part.InputAudio == nil || strings.TrimSpace(part.InputAudio.Data) == "" {
				return nil, true, fmt.Errorf("content part[%d].input_audio.data is required", i)
			}

			if strings.TrimSpace(part.InputAudio.Format) == "" {
				return nil, true, fmt.Errorf("content part[%d].input_audio.format is required", i)
			}
		default:
			return nil, true, fmt.Errorf("content part[%d].type %q is not supported", i, part.Type)
		}

		parts[i] = part
	}

	return parts, true, nil
}

func chatContentPartFromValue(raw interface{}) (ChatContentPart, error) {
	if part, ok := raw.(ChatContentPart); ok {
		return part, nil
	}

	m, ok := raw.(map[string]interface{})
	if !ok {
		return ChatContentPart{}, fmt.Errorf("must be an object")
	}

	partType, ok := m["type"].(string)
	if !ok {
		return ChatContentPart{}, fmt.Errorf("type must be a string")
	}

	part := ChatContentPart{Type: strings.TrimSpace(partType)}
	switch part.Type {
	case "text":
		if text, ok := m["text"].(string); ok {
			part.Text = text
		}
	case "image_url":
		imageURL, ok := m["image_url"].(map[string]interface{})
		if !ok {
			return ChatContentPart{}, fmt.Errorf("image_url must be an object")
		}

		part.ImageURL = &ChatImageURL{}
		if url, ok := imageURL["url"].(string); ok {
			part.ImageURL.URL = url
		}

		if detail, ok := imageURL["detail"].(string); ok {
			part.ImageURL.Detail = detail
		}
	case "video_url":
		videoValue, ok := m["video_url"]
		if !ok {
			return ChatContentPart{}, fmt.Errorf("video_url is required")
		}

		part.VideoURL = &ChatVideoURL{}

		switch videoURL := videoValue.(type) {
		case string:
			part.VideoURL.URL = videoURL
		case map[string]interface{}:
			if rawURL, ok := videoURL["url"].(string); ok {
				part.VideoURL.URL = rawURL
			}
		default:
			return ChatContentPart{}, fmt.Errorf("video_url must be a string or object")
		}
	case "file":
		file, ok := m["file"].(map[string]interface{})
		if !ok {
			return ChatContentPart{}, fmt.Errorf("file must be an object")
		}

		part.File = &ChatFile{}
		if filename, ok := file["filename"].(string); ok {
			part.File.Filename = filename
		}

		if fileData, ok := file["file_data"].(string); ok {
			part.File.FileData = fileData
		}

		if fileID, ok := file["file_id"].(string); ok {
			part.File.FileID = fileID
		}
	case "file_url":
		fileValue, ok := m["file_url"]
		if !ok {
			return ChatContentPart{}, fmt.Errorf("file_url is required")
		}

		part.FileURL = &ChatFileURL{}

		switch fileURL := fileValue.(type) {
		case string:
			part.FileURL.URL = fileURL
		case map[string]interface{}:
			if rawURL, ok := fileURL["url"].(string); ok {
				part.FileURL.URL = rawURL
			}
		default:
			return ChatContentPart{}, fmt.Errorf("file_url must be a string or object")
		}
	case "input_audio":
		inputAudio, ok := m["input_audio"].(map[string]interface{})
		if !ok {
			return ChatContentPart{}, fmt.Errorf("input_audio must be an object")
		}

		part.InputAudio = &ChatInputAudio{}
		if data, ok := inputAudio["data"].(string); ok {
			part.InputAudio.Data = data
		}

		if format, ok := inputAudio["format"].(string); ok {
			part.InputAudio.Format = format
		}
	}

	return part, nil
}

// RedactChatContent returns trace-safe content with image data URLs redacted.
func RedactChatContent(content interface{}) interface{} {
	switch v := content.(type) {
	case string:
		return RedactDataURL(v)
	case []ChatContentPart:
		return redactChatContentParts(v)
	default:
		parts, ok, err := ChatContentParts(v)
		if ok && err == nil {
			return redactChatContentParts(parts)
		}

		return content
	}
}

func redactChatContentParts(parts []ChatContentPart) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(parts))
	for _, part := range parts {
		item := map[string]interface{}{"type": part.Type}
		if part.Type == "text" {
			item["text"] = part.Text
		}

		if part.Type == "image_url" && part.ImageURL != nil {
			imageURL := map[string]interface{}{"url": RedactDataURL(part.ImageURL.URL)}
			if part.ImageURL.Detail != "" {
				imageURL["detail"] = part.ImageURL.Detail
			}

			item["image_url"] = imageURL
		}

		if part.Type == "file" && part.File != nil {
			file := map[string]interface{}{}
			if part.File.Filename != "" {
				file["filename"] = part.File.Filename
			}

			if part.File.FileData != "" {
				file["file_data"] = RedactDataURL(part.File.FileData)
			}

			if part.File.FileID != "" {
				file["file_id"] = part.File.FileID
			}

			item["file"] = file
		}

		if part.Type == "file_url" && part.FileURL != nil {
			item["file_url"] = map[string]interface{}{
				"url": RedactDataURL(part.FileURL.URL),
			}
		}

		if part.Type == "input_audio" && part.InputAudio != nil {
			item["input_audio"] = map[string]interface{}{
				"data":   "[REDACTED]",
				"format": part.InputAudio.Format,
			}
		}

		out = append(out, item)
	}

	return out
}

// RedactDataURL removes inline base64 payloads from data URLs.
func RedactDataURL(value string) string {
	trimmed := strings.TrimSpace(value)

	lower := strings.ToLower(trimmed)
	if !strings.HasPrefix(lower, "data:") {
		return value
	}

	comma := strings.Index(trimmed, ",")
	if comma < 0 {
		return "data:[redacted]"
	}

	prefix := trimmed[:comma]
	if strings.Contains(strings.ToLower(prefix), ";base64") {
		return prefix + ",[redacted]"
	}

	return "data:[redacted]"
}

// ParseImageDataURL extracts media type and base64 data from an image data URL.
func ParseImageDataURL(value string) (string, string, error) {
	return parseBase64DataURL(value, "image_url.url")
}

// ParseFileDataURL extracts media type and base64 data from a file data URL.
func ParseFileDataURL(value string) (string, string, error) {
	return parseBase64DataURL(value, "file.file_data")
}

func parseBase64DataURL(value string, field string) (string, string, error) {
	trimmed := strings.TrimSpace(value)
	if !strings.HasPrefix(strings.ToLower(trimmed), "data:") {
		return "", "", fmt.Errorf("%s must be a data URL", field)
	}

	comma := strings.Index(trimmed, ",")
	if comma < 0 {
		return "", "", fmt.Errorf("data URL is missing payload")
	}

	meta := trimmed[len("data:"):comma]
	data := strings.TrimSpace(trimmed[comma+1:])
	metaParts := strings.Split(meta, ";")

	mediaType := strings.TrimSpace(metaParts[0])
	if mediaType == "" {
		mediaType = "application/octet-stream"
	}

	base64Encoded := false

	for _, part := range metaParts[1:] {
		if strings.EqualFold(strings.TrimSpace(part), "base64") {
			base64Encoded = true
			break
		}
	}

	if !base64Encoded {
		return "", "", fmt.Errorf("data URL must be base64 encoded")
	}

	if data == "" {
		return "", "", fmt.Errorf("data URL payload is empty")
	}

	return mediaType, data, nil
}

// ToolDefinition represents a function/tool definition.
type ToolDefinition struct {
	Type           string                 `json:"type"`
	Function       *FunctionDef           `json:"function,omitempty"`
	Custom         *CustomTool            `json:"custom,omitempty"`
	Name           string                 `json:"name,omitempty"`
	AllowedCallers []string               `json:"allowed_callers,omitempty"`
	AllowedDomains []string               `json:"allowed_domains,omitempty"`
	BlockedDomains []string               `json:"blocked_domains,omitempty"`
	MaxUses        *int                   `json:"max_uses,omitempty"`
	UserLocation   map[string]interface{} `json:"user_location,omitempty"`
	CacheControl   map[string]interface{} `json:"cache_control,omitempty"`
	Strict         *bool                  `json:"strict,omitempty"`
	DeferLoading   *bool                  `json:"defer_loading,omitempty"`
}

// FunctionDef represents a function definition for tool calling.
type FunctionDef struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Strict      *bool                  `json:"strict,omitempty"`
}

// CustomTool represents OpenAI-compatible custom tool metadata.
type CustomTool struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Format      interface{} `json:"format,omitempty"`
}

// ToolCall represents an OpenAI-compatible tool call returned by a chat model.
type ToolCall struct {
	ID       string        `json:"id,omitempty"`
	Type     string        `json:"type"`
	Function *FunctionCall `json:"function,omitempty"`
	Custom   *CustomCall   `json:"custom,omitempty"`
	Index    *int          `json:"index,omitempty"`
}

// FunctionCall contains the name and JSON arguments for a function tool call.
type FunctionCall struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

// CustomCall contains the name and free-form input for a custom tool call.
type CustomCall struct {
	Name  string `json:"name,omitempty"`
	Input string `json:"input,omitempty"`
}

// ChatCompletionResponse mirrors OpenAI's chat completion response format.
type ChatCompletionResponse struct {
	ID                string   `json:"id"`
	Object            string   `json:"object"`
	Created           int64    `json:"created"`
	Model             string   `json:"model"`
	Choices           []Choice `json:"choices"`
	Usage             *Usage   `json:"usage,omitempty"`
	ServiceTier       string   `json:"service_tier,omitempty"`
	SystemFingerprint string   `json:"system_fingerprint,omitempty"`
}

// Choice represents a single choice in a chat completion response.
type Choice struct {
	Index        int             `json:"index"`
	Message      *ChatMessage    `json:"message,omitempty"`
	Delta        *ChatMessage    `json:"delta,omitempty"`
	FinishReason string          `json:"finish_reason,omitempty"`
	Logprobs     json.RawMessage `json:"logprobs,omitempty"`
}

// Usage represents token usage statistics.
type Usage struct {
	PromptTokens               int                `json:"prompt_tokens"`
	CompletionTokens           int                `json:"completion_tokens"`
	TotalTokens                int                `json:"total_tokens"`
	PromptTokensDetails        *UsageTokenDetails `json:"prompt_tokens_details,omitempty"`
	CompletionTokensDetails    *UsageTokenDetails `json:"completion_tokens_details,omitempty"`
	CacheCreationInputTokens   int                `json:"cache_creation_input_tokens,omitempty"`
	CacheCreation5mInputTokens int                `json:"cache_creation_5m_input_tokens,omitempty"`
	CacheCreation1hInputTokens int                `json:"cache_creation_1h_input_tokens,omitempty"`
	CacheReadInputTokens       int                `json:"cache_read_input_tokens,omitempty"`
	PromptCacheHitTokens       int                `json:"prompt_cache_hit_tokens,omitempty"`
	PromptCacheMissTokens      int                `json:"prompt_cache_miss_tokens,omitempty"`
	ServerToolUse              map[string]int     `json:"server_tool_use,omitempty"`
	NumServerSideToolsUsed     int                `json:"num_server_side_tools_used,omitempty"`
	CostInUSDTicks             int64              `json:"cost_in_usd_ticks,omitempty"`
	ServiceTier                string             `json:"service_tier,omitempty"`
	Speed                      string             `json:"speed,omitempty"`
}

// UsageTokenDetails carries OpenAI-compatible token subcategories needed for
// modality- and cache-aware billing.
type UsageTokenDetails struct {
	AudioTokens       int `json:"audio_tokens,omitempty"`
	TextTokens        int `json:"text_tokens,omitempty"`
	ImageTokens       int `json:"image_tokens,omitempty"`
	CachedTokens      int `json:"cached_tokens,omitempty"`
	CachedTextTokens  int `json:"cached_text_tokens,omitempty"`
	CachedImageTokens int `json:"cached_image_tokens,omitempty"`
	CacheWriteTokens  int `json:"cache_write_tokens,omitempty"`
	ReasoningTokens   int `json:"reasoning_tokens,omitempty"`
}

// StreamChunk represents a single chunk in a streaming response.
type StreamChunk struct {
	ID          string   `json:"id"`
	Object      string   `json:"object"`
	Created     int64    `json:"created"`
	Model       string   `json:"model"`
	Choices     []Choice `json:"choices"`
	Usage       *Usage   `json:"usage,omitempty"`
	ServiceTier string   `json:"service_tier,omitempty"`
}

// APIError represents an API error response.
type APIError struct {
	Error ErrorInfo `json:"error"`
}

// ErrorInfo represents the error details.
type ErrorInfo struct {
	Message string                 `json:"message"`
	Type    string                 `json:"type"`
	Param   string                 `json:"param,omitempty"`
	Code    string                 `json:"code,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// Model describes a tenant-visible model alias.
type Model struct {
	ID             string   `json:"id"`
	Object         string   `json:"object"`
	OwnedBy        string   `json:"owned_by"`
	Provider       string   `json:"provider,omitempty"`
	Aliases        []string `json:"aliases,omitempty"`
	SupportedModes []string `json:"supported_modes,omitempty"`
}

// ModelListResponse mirrors OpenAI's models list response.
type ModelListResponse struct {
	Object string  `json:"object"`
	Data   []Model `json:"data"`
}

// CompletionRequest mirrors OpenAI's legacy completion request format.
type CompletionRequest struct {
	Model            string                     `json:"model"`
	Prompt           string                     `json:"prompt"`
	Suffix           string                     `json:"suffix,omitempty"`
	Stream           bool                       `json:"stream,omitempty"`
	MaxTokens        *int                       `json:"max_tokens,omitempty"`
	MinTokens        *int                       `json:"min_tokens,omitempty"`
	Temperature      *float64                   `json:"temperature,omitempty"`
	TopP             *float64                   `json:"top_p,omitempty"`
	RandomSeed       *int                       `json:"random_seed,omitempty"`
	N                int                        `json:"n,omitempty"`
	Echo             bool                       `json:"echo,omitempty"`
	Stop             interface{}                `json:"stop,omitempty"`
	Metadata         map[string]interface{}     `json:"metadata,omitempty"`
	PromptCacheKey   string                     `json:"prompt_cache_key,omitempty"`
	AdditionalFields map[string]json.RawMessage `json:"-"`
}

func (r *CompletionRequest) UnmarshalJSON(data []byte) error {
	type requestAlias CompletionRequest

	var decoded requestAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}

	*r = CompletionRequest(decoded)
	r.AdditionalFields = fields

	return nil
}

func (r CompletionRequest) MarshalJSON() ([]byte, error) {
	type requestAlias CompletionRequest

	knownJSON, err := json.Marshal(requestAlias(r))
	if err != nil {
		return nil, err
	}

	var known map[string]json.RawMessage
	if err := json.Unmarshal(knownJSON, &known); err != nil {
		return nil, err
	}

	merged := make(map[string]json.RawMessage, len(r.AdditionalFields)+len(known))
	for key, value := range r.AdditionalFields {
		merged[key] = append(json.RawMessage(nil), value...)
	}

	for key, value := range known {
		merged[key] = value
	}

	delete(merged, "AdditionalFields")

	return json.Marshal(merged)
}

// CompletionResponse mirrors OpenAI's legacy completion response format.
type CompletionResponse struct {
	ID      string             `json:"id"`
	Object  string             `json:"object"`
	Created int64              `json:"created"`
	Model   string             `json:"model"`
	Choices []CompletionChoice `json:"choices"`
	Usage   *Usage             `json:"usage,omitempty"`
}

// CompletionChoice represents a single choice in completion response.
type CompletionChoice struct {
	Text         string `json:"text"`
	Index        int    `json:"index"`
	FinishReason string `json:"finish_reason,omitempty"`
}

// EmbeddingRequest mirrors OpenAI's embedding request format.
type EmbeddingRequest struct {
	Model          string      `json:"model"`
	Input          interface{} `json:"input"`
	EncodingFormat string      `json:"encoding_format,omitempty"`
	Dimensions     *int        `json:"dimensions,omitempty"`
	User           string      `json:"user,omitempty"`
}

// EmbeddingResponse mirrors OpenAI's embedding response format.
type EmbeddingResponse struct {
	Object string      `json:"object"`
	Data   []Embedding `json:"data"`
	Model  string      `json:"model,omitempty"`
	Usage  *Usage      `json:"usage,omitempty"`
}

// Embedding represents a single embedding vector.
type Embedding struct {
	Object    string          `json:"object"`
	Embedding json.RawMessage `json:"embedding"`
	Index     int             `json:"index"`
}

// ModerationRequest mirrors OpenAI's moderation request format.
type ModerationRequest struct {
	Model string      `json:"model"`
	Input interface{} `json:"input"`
	User  string      `json:"user,omitempty"`
}

// ModerationResponse mirrors OpenAI's moderation response format.
type ModerationResponse struct {
	ID      string             `json:"id,omitempty"`
	Model   string             `json:"model"`
	Results []ModerationResult `json:"results"`
}

// ModerationResult contains category decisions and confidence scores.
type ModerationResult struct {
	Flagged        bool               `json:"flagged"`
	Categories     map[string]bool    `json:"categories"`
	CategoryScores map[string]float64 `json:"category_scores"`
}

// RerankRequest mirrors the Cohere-compatible rerank request format used by
// gateway providers such as Cohere and other OpenAI-compatible rerank backends.
type RerankRequest struct {
	Model           string   `json:"model"`
	Query           string   `json:"query"`
	Documents       []string `json:"documents"`
	TopN            *int     `json:"top_n,omitempty"`
	MaxTokensPerDoc *int     `json:"max_tokens_per_doc,omitempty"`
	Priority        *int     `json:"priority,omitempty"`
	ReturnDocuments *bool    `json:"return_documents,omitempty"`
}

// RerankResponse preserves the Cohere-compatible rerank response envelope.
type RerankResponse struct {
	ID      string         `json:"id,omitempty"`
	Results []RerankResult `json:"results"`
	Meta    interface{}    `json:"meta,omitempty"`
	Usage   *Usage         `json:"usage,omitempty"`
}

// RerankResult is one ranked document result.
type RerankResult struct {
	Index          int         `json:"index"`
	RelevanceScore float64     `json:"relevance_score"`
	Document       interface{} `json:"document,omitempty"`
}

// ResponseRequest mirrors the subset of OpenAI's responses API used by Shahrah.
type ResponseRequest struct {
	Model              string                 `json:"model"`
	Input              interface{}            `json:"input"`
	Instructions       string                 `json:"instructions,omitempty"`
	MaxOutputTokens    *int                   `json:"max_output_tokens,omitempty"`
	Reasoning          *ResponseReasoning     `json:"reasoning,omitempty"`
	Temperature        *float64               `json:"temperature,omitempty"`
	TopP               *float64               `json:"top_p,omitempty"`
	TopLogprobs        *int                   `json:"top_logprobs,omitempty"`
	Text               *ResponseTextConfig    `json:"text,omitempty"`
	Tools              []ToolDefinition       `json:"tools,omitempty"`
	ToolChoice         interface{}            `json:"tool_choice,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
	PreviousResponseID string                 `json:"previous_response_id,omitempty"`
	Stream             bool                   `json:"stream,omitempty"`
	User               string                 `json:"user,omitempty"`
	ServiceTier        string                 `json:"service_tier,omitempty"`
}

// ResponseReasoning contains Responses API reasoning controls.
type ResponseReasoning struct {
	Effort  string `json:"effort,omitempty"`
	Summary string `json:"summary,omitempty"`
}

// ResponseTextConfig contains text output formatting options for the Responses API.
type ResponseTextConfig struct {
	Format interface{} `json:"format,omitempty"`
}

// Response represents the normalized OpenAI responses payload returned by providers.
type Response struct {
	ID                string                 `json:"id"`
	Object            string                 `json:"object"`
	CreatedAt         int64                  `json:"created_at,omitempty"`
	Model             string                 `json:"model"`
	Status            string                 `json:"status,omitempty"`
	Output            []ResponseOutput       `json:"output,omitempty"`
	Usage             *ResponseUsage         `json:"usage,omitempty"`
	Error             interface{}            `json:"error,omitempty"`
	IncompleteDetails interface{}            `json:"incomplete_details,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	ServiceTier       string                 `json:"service_tier,omitempty"`
}

// ResponseOutput captures one output item returned by the responses API.
type ResponseOutput struct {
	ID      string                  `json:"id,omitempty"`
	Type    string                  `json:"type,omitempty"`
	Role    string                  `json:"role,omitempty"`
	Status  string                  `json:"status,omitempty"`
	Content []ResponseOutputContent `json:"content,omitempty"`
}

// ResponseOutputContent captures one content block inside an output item.
type ResponseOutputContent struct {
	Type        string        `json:"type,omitempty"`
	Text        string        `json:"text,omitempty"`
	Annotations []interface{} `json:"annotations,omitempty"`
}

// ResponseUsage mirrors token accounting in the OpenAI responses API.
type ResponseUsage struct {
	InputTokens            int                         `json:"input_tokens"`
	OutputTokens           int                         `json:"output_tokens"`
	TotalTokens            int                         `json:"total_tokens"`
	InputDetails           *ResponseInputTokenDetails  `json:"input_tokens_details,omitempty"`
	OutputDetails          *ResponseOutputTokenDetails `json:"output_tokens_details,omitempty"`
	NumServerSideToolsUsed int                         `json:"num_server_side_tools_used,omitempty"`
	CostInUSDTicks         int64                       `json:"cost_in_usd_ticks,omitempty"`
}

// ResponseInputTokenDetails provides cached token detail for responses usage.
type ResponseInputTokenDetails struct {
	CachedTokens     int `json:"cached_tokens,omitempty"`
	CacheWriteTokens int `json:"cache_write_tokens,omitempty"`
}

// ResponseOutputTokenDetails provides reasoning token detail for responses usage.
type ResponseOutputTokenDetails struct {
	ReasoningTokens int `json:"reasoning_tokens,omitempty"`
}

// ImageRequest mirrors the common OpenAI-compatible image request envelope while
// leaving provider-specific fields in Extra for passthrough callers.
type ImageRequest struct {
	Model          string                 `json:"model"`
	Prompt         string                 `json:"prompt,omitempty"`
	Image          interface{}            `json:"image,omitempty"`
	Mask           interface{}            `json:"mask,omitempty"`
	N              *int                   `json:"n,omitempty"`
	Size           string                 `json:"size,omitempty"`
	Quality        string                 `json:"quality,omitempty"`
	ResponseFormat string                 `json:"response_format,omitempty"`
	User           string                 `json:"user,omitempty"`
	Extra          map[string]interface{} `json:"-"`
}

func (r *ImageRequest) UnmarshalJSON(data []byte) error {
	type alias ImageRequest

	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	for _, key := range []string{"model", "prompt", "image", "mask", "n", "size", "quality", "response_format", "user"} {
		delete(raw, key)
	}

	*r = ImageRequest(a)
	r.Extra = raw

	return nil
}

func (r ImageRequest) MarshalJSON() ([]byte, error) {
	type alias ImageRequest

	base, err := structToMap(alias(r))
	if err != nil {
		return nil, err
	}

	for k, v := range r.Extra {
		base[k] = v
	}

	return json.Marshal(base)
}

// ImageResponse preserves standard OpenAI image response fields and accepts
// provider-specific usage fields such as xAI cost_in_usd_ticks.
type ImageResponse struct {
	Created int64                  `json:"created,omitempty"`
	Data    []ImageData            `json:"data,omitempty"`
	Usage   map[string]interface{} `json:"usage,omitempty"`
	Extra   map[string]interface{} `json:"-"`
}

func (r *ImageResponse) UnmarshalJSON(data []byte) error {
	type alias ImageResponse

	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	for _, key := range []string{"created", "data", "usage"} {
		delete(raw, key)
	}

	*r = ImageResponse(a)
	r.Extra = raw

	return nil
}

func (r ImageResponse) MarshalJSON() ([]byte, error) {
	type alias ImageResponse

	base, err := structToMap(alias(r))
	if err != nil {
		return nil, err
	}

	for k, v := range r.Extra {
		base[k] = v
	}

	return json.Marshal(base)
}

// ImageData is one generated/edited/variant image result.
type ImageData struct {
	URL           string                 `json:"url,omitempty"`
	B64JSON       string                 `json:"b64_json,omitempty"`
	RevisedPrompt string                 `json:"revised_prompt,omitempty"`
	MimeType      string                 `json:"mime_type,omitempty"`
	Extra         map[string]interface{} `json:"-"`
}

func (d *ImageData) UnmarshalJSON(data []byte) error {
	type alias ImageData

	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	for _, key := range []string{"url", "b64_json", "revised_prompt", "mime_type"} {
		delete(raw, key)
	}

	*d = ImageData(a)
	d.Extra = raw

	return nil
}

func (d ImageData) MarshalJSON() ([]byte, error) {
	type alias ImageData

	base, err := structToMap(alias(d))
	if err != nil {
		return nil, err
	}

	for k, v := range d.Extra {
		base[k] = v
	}

	return json.Marshal(base)
}

func structToMap(v interface{}) (map[string]interface{}, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	out := map[string]interface{}{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}

	return out, nil
}
