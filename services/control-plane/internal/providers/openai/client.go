package openai

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
	defaultAPIBase = "https://api.openai.com/v1"
	defaultTimeout = 600 * time.Second
)

// Client implements providers.Provider for OpenAI-compatible APIs.
type Client struct {
	apiKey     string
	apiBase    string
	httpClient *http.Client
}

// Register the adapter.
func init() {
	providers.Register("openai", func(cfg providers.ProviderConfig) providers.Provider {
		return NewClient(cfg)
	})
}

// NewClient creates a new OpenAI client.
func NewClient(cfg providers.ProviderConfig) *Client {
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
	return "openai"
}

// Complete sends a non-streaming chat completion request.
func (c *Client) Complete(ctx context.Context, req contractsopenai.ChatCompletionRequest) (*contractsopenai.ChatCompletionResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiBase+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq) //nolint:bodyclose // the streaming goroutine owns and closes resp.Body.
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer providers.Close(resp.Body)

	if resp.StatusCode != http.StatusOK {
		bodyBytes, readErr := providers.ReadAllAndClose(resp)
		if readErr != nil {
			return nil, fmt.Errorf("failed to read provider error response: %w", readErr)
		}

		return nil, fmt.Errorf("api error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var result contractsopenai.ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// StreamComplete sends a streaming chat completion request.
func (c *Client) StreamComplete(ctx context.Context, req contractsopenai.ChatCompletionRequest) (<-chan contractsopenai.StreamChunk, <-chan error, error) {
	req.Stream = true

	body, err := json.Marshal(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiBase+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := c.httpClient.Do(httpReq) //nolint:bodyclose // the streaming goroutine owns and closes resp.Body.
	if err != nil {
		return nil, nil, fmt.Errorf("request failed: %w", err)
	}

	chunkCh := make(chan contractsopenai.StreamChunk)
	errCh := make(chan error, 1)

	go func() {
		defer providers.Close(resp.Body)
		defer close(chunkCh)
		defer close(errCh)

		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					return
				}

				errCh <- fmt.Errorf("read error: %w", err)

				return
			}

			line = strings.TrimSpace(line)
			if line == "" || !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				return
			}

			var chunk contractsopenai.StreamChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				errCh <- fmt.Errorf("decode error: %w", err)
				return
			}

			select {
			case chunkCh <- chunk:
			case <-ctx.Done():
				return
			}
		}
	}()

	return chunkCh, errCh, nil
}
