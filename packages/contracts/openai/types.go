package openai

// ChatCompletionRequest mirrors OpenAI's chat completion request format.
type ChatCompletionRequest struct {
	Model       string           `json:"model"`
	Messages    []ChatMessage    `json:"messages"`
	Stream      bool             `json:"stream,omitempty"`
	Temperature *float64         `json:"temperature,omitempty"`
	TopP        *float64         `json:"top_p,omitempty"`
	MaxTokens   *int             `json:"max_tokens,omitempty"`
	Tools       []ToolDefinition `json:"tools,omitempty"`
	ToolChoice  interface{}      `json:"tool_choice,omitempty"`
	User        string           `json:"user,omitempty"`
}

// ChatMessage represents a single message in a chat completion request.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content,omitempty"`
	Name    string `json:"name,omitempty"`
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
	Message string `json:"message"`
	Type    string `json:"type"`
	Param   string `json:"param,omitempty"`
	Code    string `json:"code,omitempty"`
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
	Model string      `json:"model"`
	Input interface{} `json:"input"` // string or []string
}

// EmbeddingResponse mirrors OpenAI's embedding response format.
type EmbeddingResponse struct {
	Object string      `json:"object"`
	Data   []Embedding `json:"data"`
	Usage  *Usage      `json:"usage,omitempty"`
}

// Embedding represents a single embedding vector.
type Embedding struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}
