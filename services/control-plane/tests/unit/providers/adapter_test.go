package providers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/services/control-plane/internal/providers"
	"github.com/Delpak/shahrah/services/control-plane/internal/providers/openai"
)

// TestOpenAIAdapter_Complete tests the OpenAI adapter's Complete method.
func TestOpenAIAdapter_Complete(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Header.Get("Authorization") == "" {
			t.Error("Expected Authorization header")
		}

		// Return mock response
		resp := contractsopenai.ChatCompletionResponse{
			ID:      "chatcmpl-test",
			Object:  "chat.completion",
			Created: 1234567890,
			Model:   "gpt-4",
			Choices: []contractsopenai.Choice{
				{
					Index: 0,
					Message: &contractsopenai.ChatMessage{
						Role:    "assistant",
						Content: "Hello!",
					},
					FinishReason: "stop",
				},
			},
			Usage: &contractsopenai.Usage{
				PromptTokens:     10,
				CompletionTokens: 5,
				TotalTokens:      15,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client
	cfg := providers.ProviderConfig{
		APIKey:  "test-key",
		APIBase: server.URL,
	}
	client := openai.NewClient(cfg)

	// Make request
	req := contractsopenai.ChatCompletionRequest{
		Model: "gpt-4",
		Messages: []contractsopenai.ChatMessage{
			{Role: "user", Content: "Hello"},
		},
	}

	resp, err := client.Complete(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify response
	if resp == nil {
		t.Fatal("Expected response, got nil")
	}

	if resp.ID != "chatcmpl-test" {
		t.Errorf("Expected ID 'chatcmpl-test', got '%s'", resp.ID)
	}

	if len(resp.Choices) != 1 {
		t.Errorf("Expected 1 choice, got %d", len(resp.Choices))
	}
}

// TestOpenAIAdapter_StreamComplete tests the OpenAI adapter's StreamComplete method.
func TestOpenAIAdapter_StreamComplete(t *testing.T) {
	// Create a mock server for streaming
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusOK)

		// Write SSE chunks
		chunks := []string{
			`data: {"id":"chatcmpl-1","object":"chat.completion.chunk","created":1234567890,"model":"gpt-4","choices":[{"index":0,"delta":{"content":"Hello","role":"assistant"},"finish_reason":null}]}`,
			`data: {"id":"chatcmpl-1","object":"chat.completion.chunk","created":1234567890,"model":"gpt-4","choices":[{"index":0,"delta":{"content":" World","role":"assistant"},"finish_reason":null}]}`,
			`data: {"id":"chatcmpl-1","object":"chat.completion.chunk","created":1234567890,"model":"gpt-4","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}`,
			`data: [DONE]`,
		}

		for _, chunk := range chunks {
			_, err := w.Write([]byte(chunk + "\n"))
			if err != nil {
				break
			}

			w.(http.Flusher).Flush()
		}
	}))
	defer server.Close()

	// Create client
	cfg := providers.ProviderConfig{
		APIKey:  "test-key",
		APIBase: server.URL,
	}
	client := openai.NewClient(cfg)

	// Make streaming request
	req := contractsopenai.ChatCompletionRequest{
		Model: "gpt-4",
		Messages: []contractsopenai.ChatMessage{
			{Role: "user", Content: "Hello"},
		},
		Stream: true,
	}

	streamChan, errorChan, err := client.StreamComplete(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Collect chunks
	chunksReceived := 0

	for {
		select {
		case _, ok := <-streamChan:
			if !ok {
				// Stream closed
				return
			}

			chunksReceived++
		case err := <-errorChan:
			if err != nil {
				t.Logf("Stream error: %v", err)
			}
			// Done
			if chunksReceived == 0 {
				t.Error("Expected at least one chunk")
			}

			return
		}
	}
}

// TestProviderRegistry tests the provider registry.
func TestProviderRegistry(t *testing.T) {
	// Test that providers are registered
	supported := providers.Supported()
	if len(supported) == 0 {
		t.Error("Expected at least one registered provider")
	}

	// Test that we can get a provider
	provider, err := providers.Get("openai", providers.ProviderConfig{})
	if err != nil {
		t.Fatalf("Expected to get openai provider, got error: %v", err)
	}

	if provider == nil {
		t.Fatal("Expected provider, got nil")
	}

	if provider.Name() != "openai" {
		t.Errorf("Expected 'openai', got '%s'", provider.Name())
	}

	// Test unsupported provider
	_, err = providers.Get("unsupported", providers.ProviderConfig{})
	if err == nil {
		t.Error("Expected error for unsupported provider")
	}

	// Test IsSupported
	if !providers.IsSupported("openai") {
		t.Error("Expected openai to be supported")
	}

	if providers.IsSupported("unsupported") {
		t.Error("Expected unsupported to not be supported")
	}
}
