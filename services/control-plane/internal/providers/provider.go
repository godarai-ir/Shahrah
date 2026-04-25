package providers

import (
	"context"

	"github.com/Delpak/shahrah/packages/contracts/openai"
)

// Provider defines the interface for LLM provider implementations.
type Provider interface {
	Name() string
	Complete(ctx context.Context, req openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error)
	StreamComplete(ctx context.Context, req openai.ChatCompletionRequest) (<-chan openai.StreamChunk, <-chan error, error)
}

// ProviderConfig holds provider-specific configuration.
type ProviderConfig struct {
	APIKey     string
	APIBase    string
	APIVersion string
	Region     string
	Timeout    float64
}
