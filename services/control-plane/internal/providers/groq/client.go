package groq

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
	defaultAPIBase = "https://api.groq.com/openai/v1"
	defaultTimeout = 600 * time.Second
)

// Client implements providers.Provider for Groq.
type Client struct {
	apiKey     string
	apiBase    string
	httpClient *http.Client
}

// Register the adapter
func init() {
	providers.Register("groq", func(cfg providers.ProviderConfig) providers.Provider {
		return NewClient(cfg)
	})
}

// NewClient creates a new Groq client.
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
	return "groq"
}

// BaseURL returns the base URL for the API.
func (c *Client) BaseURL() string {
	return c.apiBase
}

// EndpointPath returns the chat completions endpoint path.
func (c *Client) EndpointPath() string {
	return "/chat/completions"
}

// AuthHeaders returns the auth headers.
func (c *Client) AuthHeaders() map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + c.apiKey,
	}
}

// Complete sends a non-streaming chat completion request.
func (c *Client) Complete(ctx context.Context, req contractsopenai.ChatCompletionRequest) (*contractsopenai.ChatCompletionResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := c.apiBase + "/chat/completions"
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
		return nil, fmt.Errorf("Groq API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var result contractsopenai.ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// StreamComplete sends a streaming chat completion request.
func (c *Client) StreamComplete(ctx context.Context, req contractsopenai.ChatCompletionRequest) (<-chan contractsopenai.StreamChunk, <-chan error, error) {
	// Set stream to true in request
	req.Stream = true

	body, err := json.Marshal(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := c.apiBase + "/chat/completions"
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
		return nil, nil, fmt.Errorf("Groq API error (status %d): %s", resp.StatusCode, string(bodyBytes))
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
