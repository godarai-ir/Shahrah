// Package anthropic implements the Anthropic provider adapter.
package anthropic

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/services/control-plane/internal/providers"
)

const (
	defaultAPIBase    = "https://api.anthropic.com"
	defaultAPIVersion = "2023-06-01"
	defaultTimeout    = 600 * time.Second
)

// Client implements providers.Provider for Anthropic.
type Client struct {
	apiKey     string
	apiVersion string
	apiBase    string
	httpClient *http.Client
}

// NewClient creates a new Anthropic client.
func NewClient(cfg providers.ProviderConfig) providers.Provider {
	apiBase := cfg.APIBase
	if apiBase == "" {
		apiBase = defaultAPIBase
	}

	apiVersion := cfg.APIVersion
	if apiVersion == "" {
		apiVersion = defaultAPIVersion
	}

	timeout := defaultTimeout
	if cfg.Timeout > 0 {
		timeout = time.Duration(cfg.Timeout * float64(time.Second))
	}

	return &Client{
		apiKey:     cfg.APIKey,
		apiVersion: apiVersion,
		apiBase:    apiBase,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Name returns the provider name.
func (c *Client) Name() string {
	return "anthropic"
}

// BaseURL returns the base URL for the API.
func (c *Client) BaseURL() string {
	return c.apiBase
}

// EndpointPath returns the messages endpoint path.
func (c *Client) EndpointPath() string {
	return "/v1/messages"
}

// AuthHeaders returns the auth headers.
func (c *Client) AuthHeaders() map[string]string {
	return map[string]string{
		"x-api-key":         c.apiKey,
		"anthropic-version": c.apiVersion,
	}
}

// Anthropic API request/response types.
type anthropicRequest struct {
	Model       string                          `json:"model"`
	Messages    []anthropicMessage              `json:"messages"`
	MaxTokens   int                             `json:"max_tokens,omitempty"`
	Temperature *float64                        `json:"temperature,omitempty"`
	TopP        *float64                        `json:"top_p,omitempty"`
	Stream      bool                            `json:"stream,omitempty"`
	System      []anthropicContentBlock         `json:"system,omitempty"`
	Thinking    *contractsopenai.ThinkingConfig `json:"thinking,omitempty"`
}

type anthropicMessage struct {
	Role    string                  `json:"role"`
	Content []anthropicContentBlock `json:"content"`
}

type anthropicContentBlock struct {
	Type      string `json:"type"`
	Text      string `json:"text,omitempty"`
	Thinking  string `json:"thinking,omitempty"`
	Signature string `json:"signature,omitempty"`
	Data      string `json:"data,omitempty"`
}

type anthropicResponse struct {
	ID         string                  `json:"id"`
	Type       string                  `json:"type"`
	Role       string                  `json:"role"`
	Content    []anthropicContentBlock `json:"content"`
	Model      string                  `json:"model"`
	Usage      anthropicUsage          `json:"usage"`
	StopReason string                  `json:"stop_reason,omitempty"`
}

type anthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// Complete sends a non-streaming chat completion request.
func (c *Client) Complete(ctx context.Context, req contractsopenai.ChatCompletionRequest) (*contractsopenai.ChatCompletionResponse, error) {
	// Convert OpenAI request to Anthropic format
	anthReq := convertToAnthropicRequest(req)

	body, err := json.Marshal(anthReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := c.apiBase + "/v1/messages"

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", c.apiVersion)

	resp, err := c.httpClient.Do(httpReq) //nolint:bodyclose // the streaming goroutine owns and closes resp.Body.
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer providers.Close(resp.Body)

	if resp.StatusCode != http.StatusOK {
		bodyBytes, readErr := providers.ReadAllAndClose(resp)
		if readErr != nil {
			return nil, fmt.Errorf("failed to read anthropic error response: %w", readErr)
		}

		return nil, fmt.Errorf("anthropic API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var anthResp anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&anthResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert Anthropic response to OpenAI format
	return convertToOpenAIResponse(anthResp, req.Model), nil
}

// StreamComplete sends a streaming chat completion request.
func (c *Client) StreamComplete(ctx context.Context, req contractsopenai.ChatCompletionRequest) (<-chan contractsopenai.StreamChunk, <-chan error, error) {
	anthReq := convertToAnthropicRequest(req)
	anthReq.Stream = true

	body, err := json.Marshal(anthReq)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := c.apiBase + "/v1/messages"

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", c.apiVersion)

	resp, err := c.httpClient.Do(httpReq) //nolint:bodyclose // the streaming goroutine owns and closes resp.Body.
	if err != nil {
		return nil, nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, readErr := providers.ReadAllAndClose(resp)
		if readErr != nil {
			return nil, nil, fmt.Errorf("failed to read anthropic error response: %w", readErr)
		}

		return nil, nil, fmt.Errorf("anthropic API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	streamChan := make(chan contractsopenai.StreamChunk)
	errorChan := make(chan error, 1)

	go func() {
		defer close(streamChan)
		defer providers.Close(resp.Body)

		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					errorChan <- fmt.Errorf("stream read error: %w", err)
				}

				close(errorChan)

				return
			}

			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// Skip SSE comments
			if strings.HasPrefix(line, ":") {
				continue
			}

			// Remove "data: " prefix
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")

			// Check for [DONE]
			if data == "[DONE]" {
				close(errorChan)
				return
			}

			var chunk contractsopenai.StreamChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				errorChan <- fmt.Errorf("failed to decode stream chunk: %w", err)
				continue
			}

			streamChan <- chunk
		}
	}()

	return streamChan, errorChan, nil
}

// convertToAnthropicRequest converts OpenAI request to Anthropic format.
func convertToAnthropicRequest(req contractsopenai.ChatCompletionRequest) anthropicRequest {
	anthReq := anthropicRequest{
		Model:       req.Model,
		MaxTokens:   4096, // default
		Temperature: req.Temperature,
		TopP:        req.TopP,
	}

	// Convert messages
	for _, msg := range req.Messages {
		if msg.Role == "system" {
			anthReq.System = append(anthReq.System, anthropicContentBlock{
				Type: "text",
				Text: contractsopenai.ChatContentText(msg.Content),
			})
		} else {
			blocks := []anthropicContentBlock{{
				Type: "text",
				Text: contractsopenai.ChatContentText(msg.Content),
			}}
			if strings.EqualFold(strings.TrimSpace(msg.Role), "assistant") && len(msg.ThinkingBlocks) > 0 {
				blocks = append(anthropicBlocksFromThinkingBlocks(msg.ThinkingBlocks), blocks...)
			}
			anthReq.Messages = append(anthReq.Messages, anthropicMessage{
				Role:    msg.Role,
				Content: blocks,
			})
		}
	}

	// Set max_tokens from request if provided
	if req.MaxTokens != nil && *req.MaxTokens > 0 {
		anthReq.MaxTokens = *req.MaxTokens
	}

	anthReq.Thinking = anthropicThinkingConfig(req, anthReq.MaxTokens)

	return anthReq
}

func anthropicBlocksFromThinkingBlocks(blocks []contractsopenai.ThinkingBlock) []anthropicContentBlock {
	out := make([]anthropicContentBlock, 0, len(blocks))
	for _, block := range blocks {
		blockType, _ := block["type"].(string)
		blockType = strings.TrimSpace(blockType)
		switch blockType {
		case "thinking":
			item := anthropicContentBlock{Type: "thinking"}
			if thinking, ok := block["thinking"].(string); ok {
				item.Thinking = thinking
			}
			if signature, ok := block["signature"].(string); ok {
				item.Signature = signature
			}
			out = append(out, item)
		case "redacted_thinking":
			item := anthropicContentBlock{Type: "redacted_thinking"}
			if data, ok := block["data"].(string); ok {
				item.Data = data
			}
			out = append(out, item)
		}
	}
	return out
}

func anthropicThinkingConfig(req contractsopenai.ChatCompletionRequest, maxTokens int) *contractsopenai.ThinkingConfig {
	if req.Thinking != nil {
		if strings.EqualFold(strings.TrimSpace(req.Thinking.Type), "disabled") {
			return nil
		}
		cfg := *req.Thinking
		if strings.TrimSpace(cfg.Type) == "" {
			cfg.Type = "enabled"
		}
		if cfg.BudgetTokens <= 0 {
			cfg.BudgetTokens = defaultAnthropicThinkingBudget(req.ReasoningEffort, maxTokens)
		}
		return &cfg
	}

	effort := strings.ToLower(strings.TrimSpace(req.ReasoningEffort))
	if effort == "" || effort == "none" {
		return nil
	}

	return &contractsopenai.ThinkingConfig{
		Type:         "enabled",
		BudgetTokens: defaultAnthropicThinkingBudget(effort, maxTokens),
	}
}

func defaultAnthropicThinkingBudget(effort string, maxTokens int) int {
	budget := 2048
	switch strings.ToLower(strings.TrimSpace(effort)) {
	case "minimal", "low":
		budget = 1024
	case "medium", "":
		budget = 2048
	case "high":
		budget = 3200
	case "xhigh", "max":
		budget = 8192
	}
	if maxTokens > 1024 && budget >= maxTokens {
		return maxTokens - 1
	}
	return budget
}

// convertToOpenAIResponse converts Anthropic response to OpenAI format.
func convertToOpenAIResponse(resp anthropicResponse, model string) *contractsopenai.ChatCompletionResponse {
	content := ""
	reasoningContent := ""
	thinkingBlocks := make([]contractsopenai.ThinkingBlock, 0)
	for _, block := range resp.Content {
		switch block.Type {
		case "text":
			content += block.Text
		case "thinking":
			thinkingBlock := contractsopenai.ThinkingBlock{"type": block.Type}
			if block.Thinking != "" {
				thinkingBlock["thinking"] = block.Thinking
				reasoningContent += block.Thinking
			}
			if block.Signature != "" {
				thinkingBlock["signature"] = block.Signature
			}
			thinkingBlocks = append(thinkingBlocks, thinkingBlock)
		case "redacted_thinking":
			thinkingBlock := contractsopenai.ThinkingBlock{"type": block.Type}
			if block.Data != "" {
				thinkingBlock["data"] = block.Data
			}
			thinkingBlocks = append(thinkingBlocks, thinkingBlock)
		default:
			content += block.Text
		}
	}

	message := &contractsopenai.ChatMessage{
		Role:    "assistant",
		Content: content,
	}
	if reasoningContent != "" {
		message.ReasoningContent = reasoningContent
	}
	if len(thinkingBlocks) > 0 {
		message.ThinkingBlocks = thinkingBlocks
	}

	openaiResp := &contractsopenai.ChatCompletionResponse{
		ID:      resp.ID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []contractsopenai.Choice{{
			Index:        0,
			Message:      message,
			FinishReason: resp.StopReason,
		}},
		Usage: &contractsopenai.Usage{
			PromptTokens:     resp.Usage.InputTokens,
			CompletionTokens: resp.Usage.OutputTokens,
			TotalTokens:      resp.Usage.InputTokens + resp.Usage.OutputTokens,
		},
	}

	return openaiResp
}

// Register the adapter.
func init() {
	providers.Register("anthropic", NewClient)
}
