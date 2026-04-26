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

// Anthropic API request/response types
type anthropicRequest struct {
	Model       string                  `json:"model"`
	Messages    []anthropicMessage      `json:"messages"`
	MaxTokens   int                     `json:"max_tokens,omitempty"`
	Temperature *float64                `json:"temperature,omitempty"`
	TopP        *float64                `json:"top_p,omitempty"`
	Stream      bool                    `json:"stream,omitempty"`
	System      []anthropicContentBlock `json:"system,omitempty"`
}

type anthropicMessage struct {
	Role    string                  `json:"role"`
	Content []anthropicContentBlock `json:"content"`
}

type anthropicContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
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

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Anthropic API error (status %d): %s", resp.StatusCode, string(bodyBytes))
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

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, nil, fmt.Errorf("Anthropic API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	streamChan := make(chan contractsopenai.StreamChunk)
	errorChan := make(chan error, 1)

	go func() {
		defer close(streamChan)
		defer resp.Body.Close()

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

// convertToAnthropicRequest converts OpenAI request to Anthropic format
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
				Text: msg.Content,
			})
		} else {
			anthReq.Messages = append(anthReq.Messages, anthropicMessage{
				Role: msg.Role,
				Content: []anthropicContentBlock{{
					Type: "text",
					Text: msg.Content,
				}},
			})
		}
	}

	// Set max_tokens from request if provided
	if req.MaxTokens != nil && *req.MaxTokens > 0 {
		anthReq.MaxTokens = *req.MaxTokens
	}

	return anthReq
}

// convertToOpenAIResponse converts Anthropic response to OpenAI format
func convertToOpenAIResponse(resp anthropicResponse, model string) *contractsopenai.ChatCompletionResponse {
	content := ""
	for _, block := range resp.Content {
		content += block.Text
	}

	openaiResp := &contractsopenai.ChatCompletionResponse{
		ID:      resp.ID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []contractsopenai.Choice{{
			Index: 0,
			Message: &contractsopenai.ChatMessage{
				Role:    "assistant",
				Content: content,
			},
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

// Register the adapter
func init() {
	providers.Register("anthropic", func(cfg providers.ProviderConfig) providers.Provider {
		return NewClient(cfg)
	})
}
