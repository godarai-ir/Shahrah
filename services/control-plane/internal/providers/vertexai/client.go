package vertexai

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
	// Default Vertex AI endpoint.
	defaultAPIBase = "https://%s-aiplatform.googleapis.com/v1"
	defaultTimeout = 600 * time.Second
)

// Client implements providers.Provider for Google Vertex AI.
type Client struct {
	apiKey     string
	apiBase    string
	project    string
	location   string
	model      string
	httpClient *http.Client
}

// NewClient creates a new Vertex AI client.
func NewClient(cfg providers.ProviderConfig) providers.Provider {
	apiBase := cfg.APIBase
	region := cfg.Region

	if apiBase == "" {
		if region == "" {
			region = "us-central1"
		}

		apiBase = fmt.Sprintf(defaultAPIBase, region)
	}

	timeout := defaultTimeout
	if cfg.Timeout > 0 {
		timeout = time.Duration(cfg.Timeout * float64(time.Second))
	}

	return &Client{
		apiKey:   cfg.APIKey,
		apiBase:  apiBase,
		project:  cfg.ProjectID,
		location: region,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Name returns the provider name.
func (c *Client) Name() string {
	return "vertexai"
}

// BaseURL returns the base URL for the API.
func (c *Client) BaseURL() string {
	return c.apiBase
}

// EndpointPath returns the predict endpoint path.
func (c *Client) EndpointPath() string {
	return fmt.Sprintf("/projects/%s/locations/%s/publishers/google/models/%s:predict", c.project, c.location, c.model)
}

// AuthHeaders returns the auth headers.
func (c *Client) AuthHeaders() map[string]string {
	headers := map[string]string{}
	if c.apiKey != "" {
		headers["Authorization"] = "Bearer " + c.apiKey
	}

	return headers
}

// Vertex AI request/response types.
type vertexRequest struct {
	Instances  []vertexInstance `json:"instances"`
	Parameters vertexParameters `json:"parameters,omitempty"`
}

type vertexInstance struct {
	Content string `json:"content,omitempty"`
}

type vertexParameters struct {
	Temperature     *float64 `json:"temperature,omitempty"`
	MaxOutputTokens *int     `json:"maxOutputTokens,omitempty"`
	TopP            *float64 `json:"topP,omitempty"`
	TopK            *int     `json:"topK,omitempty"`
	Stream          bool     `json:"stream,omitempty"`
}

type vertexResponse struct {
	Predictions []vertexPrediction `json:"predictions"`
}

type vertexPrediction struct {
	Content          string       `json:"content"`
	SafetyAttributes vertexSafety `json:"safetyAttributes,omitempty"`
}

type vertexSafety struct {
	Scores     []float64 `json:"scores,omitempty"`
	Categories []string  `json:"categories,omitempty"`
}

// Complete sends a non-streaming chat completion request.
func (c *Client) Complete(ctx context.Context, req contractsopenai.ChatCompletionRequest) (*contractsopenai.ChatCompletionResponse, error) {
	// Convert OpenAI request to Vertex AI format
	vertexReq := convertToVertexRequest(req)

	body, err := json.Marshal(vertexReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/projects/%s/locations/%s/publishers/google/models/%s:predict",
		c.apiBase, c.project, c.location, c.model)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(httpReq) //nolint:bodyclose // the streaming goroutine owns and closes resp.Body.
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer providers.Close(resp.Body)

	if resp.StatusCode != http.StatusOK {
		bodyBytes, readErr := providers.ReadAllAndClose(resp)
		if readErr != nil {
			return nil, fmt.Errorf("failed to read vertex AI error response: %w", readErr)
		}

		return nil, fmt.Errorf("vertex AI API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var vertexResp vertexResponse
	if err := json.NewDecoder(resp.Body).Decode(&vertexResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert Vertex response to OpenAI format
	return convertToOpenAIResponse(vertexResp, c.model), nil
}

// StreamComplete sends a streaming chat completion request.
func (c *Client) StreamComplete(ctx context.Context, req contractsopenai.ChatCompletionRequest) (<-chan contractsopenai.StreamChunk, <-chan error, error) {
	vertexReq := convertToVertexRequest(req)
	vertexReq.Parameters.Stream = true

	body, err := json.Marshal(vertexReq)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/projects/%s/locations/%s/publishers/google/models/%s:streamPredict",
		c.apiBase, c.project, c.location, c.model)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(httpReq) //nolint:bodyclose // the streaming goroutine owns and closes resp.Body.
	if err != nil {
		return nil, nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, readErr := providers.ReadAllAndClose(resp)
		if readErr != nil {
			return nil, nil, fmt.Errorf("failed to read vertex AI error response: %w", readErr)
		}

		return nil, nil, fmt.Errorf("vertex AI API error (status %d): %s", resp.StatusCode, string(bodyBytes))
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

// convertToVertexRequest converts OpenAI request to Vertex AI format.
func convertToVertexRequest(req contractsopenai.ChatCompletionRequest) vertexRequest {
	// Build messages content
	content := ""

	for _, msg := range req.Messages {
		role := msg.Role
		if role == "system" {
			role = "system"
		}

		content += fmt.Sprintf("%s: %s\n", role, contractsopenai.ChatContentText(msg.Content))
	}

	vertexReq := vertexRequest{
		Instances: []vertexInstance{{
			Content: content,
		}},
		Parameters: vertexParameters{
			Temperature: req.Temperature,
			TopP:        req.TopP,
			Stream:      false,
		},
	}

	if req.MaxTokens != nil {
		vertexReq.Parameters.MaxOutputTokens = req.MaxTokens
	} else {
		defaultMaxTokens := 2048
		vertexReq.Parameters.MaxOutputTokens = &defaultMaxTokens
	}

	return vertexReq
}

// convertToOpenAIResponse converts Vertex response to OpenAI format.
func convertToOpenAIResponse(resp vertexResponse, model string) *contractsopenai.ChatCompletionResponse {
	content := ""
	for _, pred := range resp.Predictions {
		content += pred.Content
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
			FinishReason: "stop",
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

// SetModel sets the Vertex AI model name.
func (c *Client) SetModel(model string) {
	c.model = model
}

// SetProject sets the Google Cloud project.
func (c *Client) SetProject(project string) {
	c.project = project
}

// Register the adapter.
func init() {
	providers.Register("vertexai", func(cfg providers.ProviderConfig) providers.Provider {
		return NewClient(cfg)
	})
}
