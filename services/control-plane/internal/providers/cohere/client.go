package cohere

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
	defaultAPIBase = "https://api.cohere.ai/v1"
	defaultTimeout = 600 * time.Second
)

// Client implements providers.Provider for Cohere.
type Client struct {
	apiKey     string
	apiBase    string
	httpClient *http.Client
}

// NewClient creates a new Cohere client.
func NewClient(cfg providers.ProviderConfig) providers.Provider {
	apiBase := cfg.APIBase
	if apiBase == "" {
		apiBase = defaultAPIBase
	}

	timeout := defaultTimeout
	if cfg.Timeout > 0 {
		timeout = time.Duration(cfg.Timeout * float64(time.Second))
	}

	return &Client{
		apiKey:  cfg.APIKey,
		apiBase: apiBase,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Name returns the provider name.
func (c *Client) Name() string {
	return "cohere"
}

// BaseURL returns the base URL for the API.
func (c *Client) BaseURL() string {
	return c.apiBase
}

// EndpointPath returns the chat endpoint path.
func (c *Client) EndpointPath() string {
	return "/chat"
}

// AuthHeaders returns the auth headers.
func (c *Client) AuthHeaders() map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + c.apiKey,
	}
}

// Cohere request/response types
type cohereRequest struct {
	Model       string          `json:"model"`
	Message     string          `json:"message"`
	ChatHistory []cohereMessage `json:"chat_history,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
	Temperature *float64        `json:"temperature,omitempty"`
	MaxTokens   *int            `json:"max_tokens,omitempty"`
	TopP        *float64        `json:"top_p,omitempty"`
}

type cohereMessage struct {
	Role    string `json:"role"`
	Message string `json:"message"`
}

type cohereResponse struct {
	ChatID       string       `json:"chat_id"`
	Text         string       `json:"text"`
	FinishReason string       `json:"finish_reason"`
	Tokens       cohereTokens `json:"tokens"`
}

type cohereTokens struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Complete sends a non-streaming chat completion request.
func (c *Client) Complete(ctx context.Context, req contractsopenai.ChatCompletionRequest) (*contractsopenai.ChatCompletionResponse, error) {
	// Convert OpenAI request to Cohere format
	cohereReq := convertToCohereRequest(req)

	body, err := json.Marshal(cohereReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := c.apiBase + "/chat"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Cohere API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var cohereResp cohereResponse
	if err := json.NewDecoder(resp.Body).Decode(&cohereResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert Cohere response to OpenAI format
	return convertToOpenAIResponse(cohereResp, req.Model), nil
}

// StreamComplete sends a streaming chat completion request.
func (c *Client) StreamComplete(ctx context.Context, req contractsopenai.ChatCompletionRequest) (<-chan contractsopenai.StreamChunk, <-chan error, error) {
	cohereReq := convertToCohereRequest(req)
	cohereReq.Stream = true

	body, err := json.Marshal(cohereReq)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := c.apiBase + "/chat"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, nil, fmt.Errorf("Cohere API error (status %d): %s", resp.StatusCode, string(bodyBytes))
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

			// Skip event: lines
			if !strings.HasPrefix(line, "event:") && !strings.HasPrefix(line, "data:") {
				continue
			}

			// Parse just the text content from event stream
			// Format: "event: text-generation\n\ndata: {...text...}"
			if strings.HasPrefix(line, "data:") {
				data := strings.TrimPrefix(line, "data: ")
				if data == "[DONE]" || data == "" {
					close(errorChan)
					return
				}

				var chunk contractsopenai.StreamChunk
				if err := json.Unmarshal([]byte(data), &chunk); err != nil {
					continue
				}
				streamChan <- chunk
			}
		}
	}()

	return streamChan, errorChan, nil
}

// convertToCohereRequest converts OpenAI request to Cohere format
func convertToCohereRequest(req contractsopenai.ChatCompletionRequest) cohereRequest {
	// Build message from last user message
	message := ""
	chatHistory := make([]cohereMessage, 0, len(req.Messages))

	for i, msg := range req.Messages {
		if msg.Role == "user" {
			if i == len(req.Messages)-1 {
				// Last user message is the current prompt
				message = contractsopenai.ChatContentText(msg.Content)
			} else {
				chatHistory = append(chatHistory, cohereMessage{
					Role:    "User",
					Message: contractsopenai.ChatContentText(msg.Content),
				})
			}
		} else if msg.Role == "assistant" {
			chatHistory = append(chatHistory, cohereMessage{
				Role:    "Chatbot",
				Message: contractsopenai.ChatContentText(msg.Content),
			})
		}
	}

	cohereReq := cohereRequest{
		Model:       req.Model,
		Message:     message,
		ChatHistory: chatHistory,
		Stream:      false,
		Temperature: req.Temperature,
		TopP:        req.TopP,
	}

	if req.MaxTokens != nil {
		cohereReq.MaxTokens = req.MaxTokens
	}

	return cohereReq
}

// convertToOpenAIResponse converts Cohere response to OpenAI format
func convertToOpenAIResponse(resp cohereResponse, model string) *contractsopenai.ChatCompletionResponse {
	finishReason := resp.FinishReason
	if finishReason == "" {
		finishReason = "stop"
	}

	openaiResp := &contractsopenai.ChatCompletionResponse{
		ID:      fmt.Sprintf("chatcmpl-%s", resp.ChatID),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []contractsopenai.Choice{{
			Index: 0,
			Message: &contractsopenai.ChatMessage{
				Role:    "assistant",
				Content: resp.Text,
			},
			FinishReason: finishReason,
		}},
		Usage: &contractsopenai.Usage{
			PromptTokens:     resp.Tokens.PromptTokens,
			CompletionTokens: resp.Tokens.CompletionTokens,
			TotalTokens:      resp.Tokens.TotalTokens,
		},
	}

	return openaiResp
}

// Register the adapter
func init() {
	providers.Register("cohere", func(cfg providers.ProviderConfig) providers.Provider {
		return NewClient(cfg)
	})
}
