package providers

import (
	"fmt"
	"sync"

	"github.com/Delpak/shahrah/packages/contracts/openai"
)

// ProviderFactory creates a provider adapter instance
type ProviderFactory func(cfg ProviderConfig) Provider

// ProviderRegistry maps provider names to adapter factories
type ProviderRegistry struct {
	mu        sync.RWMutex
	factories map[string]ProviderFactory
}

// Global registry instance
var global = &ProviderRegistry{
	factories: make(map[string]ProviderFactory),
}

// Register adds a provider adapter factory to the registry
func Register(name string, factory ProviderFactory) {
	global.mu.Lock()
	defer global.mu.Unlock()
	global.factories[name] = factory
}

// Get retrieves a provider adapter by name
func Get(name string, cfg ProviderConfig) (Provider, error) {
	global.mu.RLock()
	factory, ok := global.factories[name]
	global.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unsupported provider: %s", name)
	}
	return factory(cfg), nil
}

// Supported returns list of supported provider names
func Supported() []string {
	global.mu.RLock()
	defer global.mu.RUnlock()

	names := make([]string, 0, len(global.factories))
	for name := range global.factories {
		names = append(names, name)
	}
	return names
}

// IsSupported checks if a provider is registered
func IsSupported(name string) bool {
	global.mu.RLock()
	_, ok := global.factories[name]
	global.mu.RUnlock()
	return ok
}

// NewProviderConfig creates a ProviderConfig from provider endpoint config
func NewProviderConfig(endpointCfg map[string]interface{}) ProviderConfig {
	cfg := ProviderConfig{}

	// Extract common fields
	if v, ok := endpointCfg["api_key"].(string); ok {
		cfg.APIKey = v
	}
	if v, ok := endpointCfg["api_base"].(string); ok {
		cfg.APIBase = v
	}
	if v, ok := endpointCfg["api_version"].(string); ok {
		cfg.APIVersion = v
	}
	if v, ok := endpointCfg["region"].(string); ok {
		cfg.Region = v
	}
	if v, ok := endpointCfg["timeout"].(float64); ok {
		cfg.Timeout = v
	}

	return cfg
}

// ProviderAdapter is the interface all LLM provider adapters must implement
// This extends the base Provider interface with adapter-specific methods
type ProviderAdapter interface {
	Provider
	// BaseURL returns the provider's API base URL
	BaseURL() string
	// EndpointPath returns the chat completions endpoint path
	EndpointPath() string
	// AuthHeaders returns additional auth headers needed
	AuthHeaders() map[string]string
}

// TransformRequest transforms an OpenAI request to provider-specific format
type TransformRequest func(req openai.ChatCompletionRequest) (map[string]interface{}, error)

// TransformResponse transforms a provider response to OpenAI format
type TransformResponse func(resp []byte) (*openai.ChatCompletionResponse, error)

// StreamParser parses streaming responses
type StreamParser func(data []byte) ([]openai.StreamChunk, error)

// AdapterConfig holds provider-specific adapter configuration
type AdapterConfig struct {
	BaseURL         string
	EndpointPath    string
	AuthMode        string // api_key, bearer, aws_sigv4, azure_ad
	AuthHeaders     map[string]string
	TransformReq    TransformRequest
	TransformRes    TransformResponse
	StreamParse     StreamParser
	UseOpenAIFormat bool // If true, request/response use OpenAI format directly
}
