package bedrock

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
	// Default Bedrock runtime endpoint
	defaultAPIBase = "https://bedrock-runtime.%s.amazonaws.com"
	// Default timeout
	defaultTimeout = 600 * time.Second
)

// Client implements providers.Provider for AWS Bedrock.
type Client struct {
	apiKey     string
	apiBase    string
	region     string
	model      string
	httpClient *http.Client
}

// NewClient creates a new Bedrock client.
func NewClient(cfg providers.ProviderConfig) providers.Provider {
	apiBase := cfg.APIBase
	region := cfg.Region

	if apiBase == "" {
		if region == "" {
			region = "us-east-1"
		}
		apiBase = fmt.Sprintf(defaultAPIBase, region)
	}

	timeout := defaultTimeout
	if cfg.Timeout > 0 {
		timeout = time.Duration(cfg.Timeout * float64(time.Second))
	}

	return &Client{
		apiKey:  cfg.APIKey,
		apiBase: apiBase,
		region:  region,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Name returns the provider name.
func (c *Client) Name() string {
	return "bedrock"
}

// BaseURL returns the base URL for the API.
func (c *Client) BaseURL() string {
	return c.apiBase
}

// EndpointPath returns the invoke endpoint path.
func (c *Client) EndpointPath() string {
	return fmt.Sprintf("/model/%s/invoke", c.model)
}

// AuthHeaders returns the auth headers.
func (c *Client) AuthHeaders() map[string]string {
	// Note: For Bedrock, authentication is typically handled via AWS SDK (boto3)
	// This is a simplified version that can use API key if provided
	headers := map[string]string{}
	if c.apiKey != "" {
		headers["x-api-key"] = c.apiKey
	}
	return headers
}

// Bedrock request/response types for Converse API
type bedrockRequest struct {
	Messages        []bedrockMessage       `json:"messages"`
	InferenceConfig bedrockInferenceConfig `json:"inferenceConfig,omitempty"`
	System          []bedrockSystemMessage `json:"system,omitempty"`
}

type bedrockMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type bedrockInferenceConfig struct {
	Temperature *float64 `json:"temperature,omitempty"`
	MaxTokens   *int     `json:"maxTokens,omitempty"`
	TopP        *float64 `json:"topP,omitempty"`
}

type bedrockSystemMessage struct {
	Text string `json:"text"`
}

type bedrockResponse struct {
	Output        bedrockOutput         `json:"output"`
	Metrics       bedrockMetrics        `json:"metrics,omitempty"`
	Conversations []bedrockConversation `json:"conversations,omitempty"`
}

type bedrockOutput struct {
	Message bedrockMessage `json:"message"`
}

type bedrockMetrics struct {
	LatencyMs int `json:"latencyMs,omitempty"`
}

type bedrockConversation struct {
	Messages []bedrockMessage `json:"messages,omitempty"`
}

// Complete sends a non-streaming chat completion request.
func (c *Client) Complete(ctx context.Context, req contractsopenai.ChatCompletionRequest) (*contractsopenai.ChatCompletionResponse, error) {
	// Convert OpenAI request to Bedrock format
	bedrockReq := convertToBedrockRequest(req, c.model)

	body, err := json.Marshal(bedrockReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/model/%s/invoke", c.apiBase, c.model)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("x-api-key", c.apiKey)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Bedrock API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var bedrockResp bedrockResponse
	if err := json.NewDecoder(resp.Body).Decode(&bedrockResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert Bedrock response to OpenAI format
	return convertToOpenAIResponse(bedrockResp, c.model), nil
}

// StreamComplete sends a streaming chat completion request.
func (c *Client) StreamComplete(ctx context.Context, req contractsopenai.ChatCompletionRequest) (<-chan contractsopenai.StreamChunk, <-chan error, error) {
	bedrockReq := convertToBedrockRequest(req, c.model)

	body, err := json.Marshal(bedrockReq)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Note: Bedrock streaming uses different endpoint
	endpoint := fmt.Sprintf("%s/model/%s/invoke", c.apiBase, c.model)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/jsonl")
	if c.apiKey != "" {
		httpReq.Header.Set("x-api-key", c.apiKey)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, nil, fmt.Errorf("Bedrock API error (status %d): %s", resp.StatusCode, string(bodyBytes))
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

			var chunk contractsopenai.StreamChunk
			if err := json.Unmarshal([]byte(line), &chunk); err != nil {
				errorChan <- fmt.Errorf("failed to decode stream chunk: %w", err)
				continue
			}

			streamChan <- chunk
		}
	}()

	return streamChan, errorChan, nil
}

// convertToBedrockRequest converts OpenAI request to Bedrock format
func convertToBedrockRequest(req contractsopenai.ChatCompletionRequest, model string) bedrockRequest {
	bedrockReq := bedrockRequest{
		Messages: make([]bedrockMessage, 0, len(req.Messages)),
	}

	// Convert messages
	for _, msg := range req.Messages {
		if msg.Role == "system" {
			bedrockReq.System = append(bedrockReq.System, bedrockSystemMessage{
				Text: contractsopenai.ChatContentText(msg.Content),
			})
		} else {
			bedrockReq.Messages = append(bedrockReq.Messages, bedrockMessage{
				Role:    msg.Role,
				Content: contractsopenai.ChatContentText(msg.Content),
			})
		}
	}

	// Set inference config
	if req.Temperature != nil || req.MaxTokens != nil {
		bedrockReq.InferenceConfig = bedrockInferenceConfig{
			Temperature: req.Temperature,
			MaxTokens:   req.MaxTokens,
			TopP:        req.TopP,
		}
	}

	// Default max tokens if not set
	if bedrockReq.InferenceConfig.MaxTokens == nil {
		defaultMaxTokens := 4096
		bedrockReq.InferenceConfig.MaxTokens = &defaultMaxTokens
	}

	return bedrockReq
}

// convertToOpenAIResponse converts Bedrock response to OpenAI format
func convertToOpenAIResponse(resp bedrockResponse, model string) *contractsopenai.ChatCompletionResponse {
	openaiResp := &contractsopenai.ChatCompletionResponse{
		ID:      fmt.Sprintf("chatcmpl-%d", time.Now().Unix()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []contractsopenai.Choice{{
			Index: 0,
			Message: &contractsopenai.ChatMessage{
				Role:    resp.Output.Message.Role,
				Content: resp.Output.Message.Content,
			},
			FinishReason: "stop",
		}},
	}

	// Estimate usage (Bedrock doesn't always return usage)
	// Use approximate token count based on content length
	if len(resp.Output.Message.Content) > 0 {
		estimatedTokens := len(resp.Output.Message.Content) / 4 // rough estimate
		openaiResp.Usage = &contractsopenai.Usage{
			PromptTokens:     100, // estimated
			CompletionTokens: estimatedTokens,
			TotalTokens:      100 + estimatedTokens,
		}
	}

	return openaiResp
}

// SetModel sets the Bedrock model name
func (c *Client) SetModel(model string) {
	c.model = model
}

// Register the adapter
func init() {
	providers.Register("bedrock", func(cfg providers.ProviderConfig) providers.Provider {
		return NewClient(cfg)
	})
}
