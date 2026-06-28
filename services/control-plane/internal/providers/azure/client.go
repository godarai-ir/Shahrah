// Package azure implements the Azure OpenAI provider adapter.
package azure

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
	// Default API base for Azure OpenAI.
	defaultAPIBase = "https://%s.openai.azure.com/openai/deployments/%s"
	// Default API version.
	defaultAPIVersion = "2024-02-15"
	defaultTimeout    = 600 * time.Second
)

// Client implements providers.Provider for Azure OpenAI.
type Client struct {
	apiKey     string
	apiBase    string
	apiVersion string
	deployment string
	httpClient *http.Client
}

// NewClient creates a new Azure OpenAI client.
func NewClient(cfg providers.ProviderConfig) providers.Provider {
	return NewClientWithDeployment(cfg, "", "")
}

// NewClientWithDeployment creates a new Azure OpenAI client with deployment specifics.
func NewClientWithDeployment(cfg providers.ProviderConfig, resource, deployment string) providers.Provider {
	apiBase := cfg.APIBase
	if apiBase == "" {
		apiBase = fmt.Sprintf(defaultAPIBase, resource, deployment)
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
		apiBase:    apiBase,
		apiVersion: apiVersion,
		deployment: deployment,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Name returns the provider name.
func (c *Client) Name() string {
	return "azure"
}

// BaseURL returns the base URL for the API.
func (c *Client) BaseURL() string {
	return c.apiBase
}

// EndpointPath returns the chat completions endpoint path.
func (c *Client) EndpointPath() string {
	return "/chat/completions?api-version=" + c.apiVersion
}

// AuthHeaders returns the auth headers.
func (c *Client) AuthHeaders() map[string]string {
	return map[string]string{
		"api-key": c.apiKey,
	}
}

// Complete sends a non-streaming chat completion request.
func (c *Client) Complete(ctx context.Context, req contractsopenai.ChatCompletionRequest) (*contractsopenai.ChatCompletionResponse, error) {
	// Azure uses deployment name, not model name
	azureReq := map[string]interface{}{
		"messages":    req.Messages,
		"temperature": req.Temperature,
		"top_p":       req.TopP,
		"max_tokens":  req.MaxTokens,
		"stream":      false,
	}

	if req.User != "" {
		azureReq["user"] = req.User
	}

	body, err := json.Marshal(azureReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := c.apiBase + "/chat/completions?api-version=" + c.apiVersion

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("api-key", c.apiKey)

	resp, err := c.httpClient.Do(httpReq) //nolint:bodyclose // the streaming goroutine owns and closes resp.Body.
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer providers.Close(resp.Body)

	if resp.StatusCode != http.StatusOK {
		bodyBytes, readErr := providers.ReadAllAndClose(resp)
		if readErr != nil {
			return nil, fmt.Errorf("failed to read azure error response: %w", readErr)
		}

		return nil, fmt.Errorf("azure API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var result contractsopenai.ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Ensure response has correct format
	if result.Model == "" {
		result.Model = c.deployment
	}

	return &result, nil
}

// StreamComplete sends a streaming chat completion request.
func (c *Client) StreamComplete(ctx context.Context, req contractsopenai.ChatCompletionRequest) (<-chan contractsopenai.StreamChunk, <-chan error, error) {
	azureReq := map[string]interface{}{
		"messages":    req.Messages,
		"temperature": req.Temperature,
		"top_p":       req.TopP,
		"max_tokens":  req.MaxTokens,
		"stream":      true,
	}

	body, err := json.Marshal(azureReq)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := c.apiBase + "/chat/completions?api-version=" + c.apiVersion

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("api-key", c.apiKey)

	resp, err := c.httpClient.Do(httpReq) //nolint:bodyclose // the streaming goroutine owns and closes resp.Body.
	if err != nil {
		return nil, nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, readErr := providers.ReadAllAndClose(resp)
		if readErr != nil {
			return nil, nil, fmt.Errorf("failed to read azure error response: %w", readErr)
		}

		return nil, nil, fmt.Errorf("azure API error (status %d): %s", resp.StatusCode, string(bodyBytes))
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

// Register the adapter.
func init() {
	providers.Register("azure", NewClient)
}
