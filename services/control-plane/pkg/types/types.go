// Package types contains shared top-level control-plane types.
package types

import "time"

// ChatCompletionRequest mirrors OpenAI's chat completion request format.
type ChatCompletionRequest struct {
	Model             string           `json:"model"`
	Messages          []ChatMessage    `json:"messages"`
	Stream            bool             `json:"stream,omitempty"`
	Temperature       *float64         `json:"temperature,omitempty"`
	TopP              *float64         `json:"top_p,omitempty"`
	MaxTokens         *int             `json:"max_tokens,omitempty"`
	ReasoningEffort   string           `json:"reasoning_effort,omitempty"`
	Thinking          *ThinkingConfig  `json:"thinking,omitempty"`
	ParallelToolCalls *bool            `json:"parallel_tool_calls,omitempty"`
	ResponseFormat    interface{}      `json:"response_format,omitempty"`
	Tools             []ToolDefinition `json:"tools,omitempty"`
	ToolChoice        interface{}      `json:"tool_choice,omitempty"`
}

// ChatMessage represents a single message in a chat completion request.
type ChatMessage struct {
	Role             string          `json:"role"`
	Content          string          `json:"content,omitempty"`
	Name             string          `json:"name,omitempty"`
	Refusal          string          `json:"refusal,omitempty"`
	ReasoningContent string          `json:"reasoning_content,omitempty"`
	ThinkingBlocks   []ThinkingBlock `json:"thinking_blocks,omitempty"`
	ToolCalls        []ToolCall      `json:"tool_calls,omitempty"`
	ToolCallID       string          `json:"tool_call_id,omitempty"`
	FunctionCall     *FunctionCall   `json:"function_call,omitempty"`
}

// ThinkingConfig preserves provider thinking configuration.
type ThinkingConfig struct {
	Type         string `json:"type,omitempty"`
	BudgetTokens int    `json:"budget_tokens,omitempty"`
	Display      string `json:"display,omitempty"`
}

// ThinkingBlock preserves provider reasoning block fields.
type ThinkingBlock map[string]interface{}

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
	Strict      *bool                  `json:"strict,omitempty"`
}

// ToolCall represents an OpenAI-compatible tool call returned by a chat model.
type ToolCall struct {
	ID       string        `json:"id,omitempty"`
	Type     string        `json:"type"`
	Function *FunctionCall `json:"function,omitempty"`
	Index    *int          `json:"index,omitempty"`
}

// FunctionCall contains the name and JSON arguments for a function tool call.
type FunctionCall struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
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

// RerankRequest mirrors the Cohere-compatible rerank request format.
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

// UserAPIKeyAuth represents the authenticated user context.
type UserAPIKeyAuth struct {
	TokenHash string     `json:"token_hash"`
	TeamID    string     `json:"team_id,omitempty"`
	UserID    string     `json:"user_id,omitempty"`
	Role      string     `json:"role"`
	Models    []string   `json:"models"`
	Spend     float64    `json:"spend"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}
