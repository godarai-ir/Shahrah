package gemini

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
	// Default Gemini API endpoint
	defaultAPIBase = "https://generativelanguage.googleapis.com/v1"
	defaultTimeout = 600 * time.Second
)

// Client implements providers.Provider for Google Gemini (AI Studio).
type Client struct {
	apiKey     string
	apiBase    string
	model      string
	httpClient *http.Client
}

// NewClient creates a new Gemini client.
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
	return "gemini"
}

// BaseURL returns the base URL for the API.
func (c *Client) BaseURL() string {
	return c.apiBase
}

// EndpointPath returns the generate content endpoint path.
func (c *Client) EndpointPath() string {
	return fmt.Sprintf("/models/%s:generateContent", c.model)
}

// AuthHeaders returns the auth headers.
func (c *Client) AuthHeaders() map[string]string {
	headers := map[string]string{}
	if c.apiKey != "" {
		headers["Authorization"] = "Bearer " + c.apiKey
	}
	return headers
}

// Gemini request/response types
type geminiRequest struct {
	Contents         []geminiContent        `json:"contents"`
	SafetySettings   []geminiSafetySetting  `json:"safetySettings,omitempty"`
	GenerationConfig geminiGenerationConfig `json:"generationConfig,omitempty"`
}

type geminiContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text,omitempty"`
}

type geminiSafetySetting struct {
	Category  string `json:"category"`
	Threshold string `json:"threshold"`
}

type geminiGenerationConfig struct {
	Temperature     *float64 `json:"temperature,omitempty"`
	MaxOutputTokens *int     `json:"maxOutputTokens,omitempty"`
	TopP            *float64 `json:"topP,omitempty"`
	TopK            *int     `json:"topK,omitempty"`
	StopSequences   []string `json:"stopSequences,omitempty"`
}

type geminiResponse struct {
	Candidates     []geminiCandidate    `json:"candidates"`
	PromptFeedback geminiPromptFeedback `json:"promptFeedback,omitempty"`
}

type geminiCandidate struct {
	Content       geminiContent        `json:"content"`
	FinishReason  string               `json:"finishReason"`
	SafetyRatings []geminiSafetyRating `json:"safetyRatings,omitempty"`
}

type geminiSafetyRating struct {
	Category    string `json:"category"`
	Probability string `json:"probability"`
}

type geminiPromptFeedback struct {
	SafetyRatings []geminiSafetyRating `json:"safetyRatings,omitempty"`
}

// Complete sends a non-streaming chat completion request.
func (c *Client) Complete(ctx context.Context, req contractsopenai.ChatCompletionRequest) (*contractsopenai.ChatCompletionResponse, error) {
	// Convert OpenAI request to Gemini format
	geminiReq := convertToGeminiRequest(req, c.model)

	body, err := json.Marshal(geminiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Use API key in query params
	endpoint := fmt.Sprintf("%s/models/%s:generateContent?key=%s", c.apiBase, c.model, c.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Gemini API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var geminiResp geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert Gemini response to OpenAI format
	return convertToOpenAIResponse(geminiResp, c.model), nil
}

// StreamComplete sends a streaming chat completion request.
func (c *Client) StreamComplete(ctx context.Context, req contractsopenai.ChatCompletionRequest) (<-chan contractsopenai.StreamChunk, <-chan error, error) {
	// For Gemini, streaming is handled differently - using server-side streaming
	geminiReq := convertToGeminiRequest(req, c.model)

	body, err := json.Marshal(geminiReq)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/models/%s:streamGenerateContent?key=%s", c.apiBase, c.model, c.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, nil, fmt.Errorf("Gemini API error (status %d): %s", resp.StatusCode, string(bodyBytes))
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

// convertToGeminiRequest converts OpenAI request to Gemini format
func convertToGeminiRequest(req contractsopenai.ChatCompletionRequest, model string) geminiRequest {
	geminiReq := geminiRequest{
		Contents: make([]geminiContent, 0, len(req.Messages)),
	}

	// Convert messages (Gemini uses "model" for assistant, "user" for user)
	for _, msg := range req.Messages {
		role := msg.Role
		if role == "system" {
			role = "user" // Gemini treats system as user with model instruction
		}
		if role == "assistant" {
			role = "model"
		}

		geminiReq.Contents = append(geminiReq.Contents, geminiContent{
			Role:  role,
			Parts: []geminiPart{{Text: contractsopenai.ChatContentText(msg.Content)}},
		})
	}

	// Set generation config
	geminiReq.GenerationConfig = geminiGenerationConfig{
		Temperature: req.Temperature,
		TopP:        req.TopP,
	}

	if req.MaxTokens != nil {
		geminiReq.GenerationConfig.MaxOutputTokens = req.MaxTokens
	}

	return geminiReq
}

// convertToOpenAIResponse converts Gemini response to OpenAI format
func convertToOpenAIResponse(resp geminiResponse, model string) *contractsopenai.ChatCompletionResponse {
	content := ""
	finishReason := "stop"

	for _, candidate := range resp.Candidates {
		for _, part := range candidate.Content.Parts {
			content += part.Text
		}
		finishReason = candidate.FinishReason
	}

	openaiResp := &contractsopenai.ChatCompletionResponse{
		ID:      fmt.Sprintf("chatcmpl-%d", time.Now().Unix()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []contractsopenai.Choice{{
			Index: 0,
			Message: &contractsopenai.ChatMessage{
				Role:    "model",
				Content: content,
			},
			FinishReason: finishReason,
		}},
	}

	// Estimate usage
	if len(content) > 0 {
		estimatedTokens := len(content) / 4
		openaiResp.Usage = &contractsopenai.Usage{
			PromptTokens:     100,
			CompletionTokens: estimatedTokens,
			TotalTokens:      100 + estimatedTokens,
		}
	}

	return openaiResp
}

// SetModel sets the Gemini model name
func (c *Client) SetModel(model string) {
	c.model = model
}

// Register the adapter
func init() {
	providers.Register("gemini", func(cfg providers.ProviderConfig) providers.Provider {
		return NewClient(cfg)
	})
}
