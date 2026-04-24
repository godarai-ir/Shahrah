package providers

import (
	"context"

	"github.com/Delpak/shahrah/pkg/types"
)

// Provider defines the interface for LLM provider implementations.
type Provider interface {
	Name() string
	Complete(ctx context.Context, req types.ChatCompletionRequest) (*types.ChatCompletionResponse, error)
	StreamComplete(ctx context.Context, req types.ChatCompletionRequest) (<-chan types.StreamChunk, <-chan error, error)
}

// ProviderConfig holds provider-specific configuration.
type ProviderConfig struct {
	APIKey     string
	APIBase    string
	APIVersion string
	Region     string
	Timeout    float64
}

