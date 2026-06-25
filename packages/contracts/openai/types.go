package openai

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ChatCompletionRequest mirrors OpenAI's chat completion request format.
type ChatCompletionRequest struct {
	Model            string                 `json:"model"`
	Messages         []ChatMessage          `json:"messages"`
	Stream           bool                   `json:"stream,omitempty"`
	Temperature      *float64               `json:"temperature,omitempty"`
	TopP             *float64               `json:"top_p,omitempty"`
	MaxTokens        *int                   `json:"max_tokens,omitempty"`
	ResponseFormat   interface{}            `json:"response_format,omitempty"`
	Tools            []ToolDefinition       `json:"tools,omitempty"`
	ToolChoice       interface{}            `json:"tool_choice,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	User             string                 `json:"user,omitempty"`
	PromptVersionFQN string                 `json:"prompt_version_fqn,omitempty"`
	PromptVariables  map[string]string      `json:"prompt_variables,omitempty"`
}

// ChatMessage represents a single message in a chat completion request.
type ChatMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content,omitempty"`
	Name    string      `json:"name,omitempty"`
}

// ChatContentPart represents one OpenAI-compatible multimodal message part.
type ChatContentPart struct {
	Type     string        `json:"type"`
	Text     string        `json:"text,omitempty"`
	ImageURL *ChatImageURL `json:"image_url,omitempty"`
}

// ChatImageURL contains an image URL or data URL for multimodal requests.
type ChatImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"`
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
	partType, _ := m["type"].(string)
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
	trimmed := strings.TrimSpace(value)
	if !strings.HasPrefix(strings.ToLower(trimmed), "data:") {
		return "", "", fmt.Errorf("image_url.url must be a data URL")
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
	Type     string       `json:"type"`
	Function *FunctionDef `json:"function,omitempty"`
}

// FunctionDef represents a function definition for tool calling.
type FunctionDef struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// ChatCompletionResponse mirrors OpenAI's chat completion response format.
type ChatCompletionResponse struct {
	ID                string   `json:"id"`
	Object            string   `json:"object"`
	Created           int64    `json:"created"`
	Model             string   `json:"model"`
	Choices           []Choice `json:"choices"`
	Usage             *Usage   `json:"usage,omitempty"`
	SystemFingerprint string   `json:"system_fingerprint,omitempty"`
}

// Choice represents a single choice in a chat completion response.
type Choice struct {
	Index        int          `json:"index"`
	Message      *ChatMessage `json:"message,omitempty"`
	Delta        *ChatMessage `json:"delta,omitempty"`
	FinishReason string       `json:"finish_reason,omitempty"`
}

// Usage represents token usage statistics.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// StreamChunk represents a single chunk in a streaming response.
type StreamChunk struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
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
	ID       string   `json:"id"`
	Object   string   `json:"object"`
	OwnedBy  string   `json:"owned_by"`
	Provider string   `json:"provider,omitempty"`
	Aliases  []string `json:"aliases,omitempty"`
}

// ModelListResponse mirrors OpenAI's models list response.
type ModelListResponse struct {
	Object string  `json:"object"`
	Data   []Model `json:"data"`
}

// CompletionRequest mirrors OpenAI's legacy completion request format.
type CompletionRequest struct {
	Model       string      `json:"model"`
	Prompt      string      `json:"prompt"`
	Stream      bool        `json:"stream,omitempty"`
	MaxTokens   *int        `json:"max_tokens,omitempty"`
	Temperature *float64    `json:"temperature,omitempty"`
	TopP        *float64    `json:"top_p,omitempty"`
	N           int         `json:"n,omitempty"`
	Echo        bool        `json:"echo,omitempty"`
	Stop        interface{} `json:"stop,omitempty"`
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

// ResponseRequest mirrors the subset of OpenAI's responses API used by Shahrah.
type ResponseRequest struct {
	Model              string                 `json:"model"`
	Input              interface{}            `json:"input"`
	Instructions       string                 `json:"instructions,omitempty"`
	MaxOutputTokens    *int                   `json:"max_output_tokens,omitempty"`
	Temperature        *float64               `json:"temperature,omitempty"`
	TopP               *float64               `json:"top_p,omitempty"`
	Tools              []ToolDefinition       `json:"tools,omitempty"`
	ToolChoice         interface{}            `json:"tool_choice,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
	PreviousResponseID string                 `json:"previous_response_id,omitempty"`
	Stream             bool                   `json:"stream,omitempty"`
	User               string                 `json:"user,omitempty"`
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
	InputTokens   int                         `json:"input_tokens"`
	OutputTokens  int                         `json:"output_tokens"`
	TotalTokens   int                         `json:"total_tokens"`
	InputDetails  *ResponseInputTokenDetails  `json:"input_tokens_details,omitempty"`
	OutputDetails *ResponseOutputTokenDetails `json:"output_tokens_details,omitempty"`
}

// ResponseInputTokenDetails provides cached token detail for responses usage.
type ResponseInputTokenDetails struct {
	CachedTokens int `json:"cached_tokens,omitempty"`
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
