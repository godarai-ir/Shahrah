package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Config represents the top-level configuration.
type Config struct {
	ModelList       []ModelDeployment     `mapstructure:"model_list"`
	RouterSettings  RouterSettings        `mapstructure:"router_settings"`
	ResponseCache   ResponseCacheSettings `mapstructure:"response_cache"`
	GeneralSettings GeneralSettings       `mapstructure:"general_settings"`
	TenantSettings  TenantSettings        `mapstructure:"tenant_settings"`
	TunnelSettings  TunnelSettings        `mapstructure:"tunnel_settings"`
	BillingSettings BillingSettings       `mapstructure:"billing_settings"`
}

// ResponseCacheSettings controls opt-in exact and semantic response caching.
type ResponseCacheSettings struct {
	Enabled               bool   `mapstructure:"enabled"`
	KeyPrefix             string `mapstructure:"key_prefix"`
	DefaultTTLSeconds     int    `mapstructure:"default_ttl_seconds"`
	MinTTLSeconds         int    `mapstructure:"min_ttl_seconds"`
	MaxTTLSeconds         int    `mapstructure:"max_ttl_seconds"`
	MaxResponseBytes      int    `mapstructure:"max_response_bytes"`
	MaxSemanticCandidates int    `mapstructure:"max_semantic_candidates"`
	EmbeddingProvider     string `mapstructure:"embedding_provider"`
	EmbeddingModel        string `mapstructure:"embedding_model"`
	EmbeddingRevision     string `mapstructure:"embedding_revision"`
	EmbeddingBaseURL      string `mapstructure:"embedding_base_url"`
	EmbeddingAPIKey       string `mapstructure:"embedding_api_key"`
	EmbeddingDimensions   int    `mapstructure:"embedding_dimensions"`
	EmbeddingTimeoutSec   int    `mapstructure:"embedding_timeout_seconds"`
	EmbeddingMaxBatchSize int    `mapstructure:"embedding_max_batch_size"`
}

// ModelDeployment represents a single model deployment configuration.
type ModelDeployment struct {
	ModelName        string                 `mapstructure:"model_name"`
	ProviderEndpoint ProviderEndpointConfig `mapstructure:"provider_endpoint_config"`
	ModelInfo        map[string]interface{} `mapstructure:"model_info"`
}

// ProviderEndpointConfig represents the connection and routing parameters for an LLM provider endpoint.
type ProviderEndpointConfig struct {
	Model         string            `mapstructure:"model"`
	Provider      string            `mapstructure:"provider"` // Provider name: openai, azure, anthropic, groq, mistral, togetherai, cohere, bedrock, vertexai, gemini
	APIKey        string            `mapstructure:"api_key"`
	APIBase       string            `mapstructure:"api_base"`
	APIVersion    string            `mapstructure:"api_version"`
	AuthMode      string            `mapstructure:"auth_mode"`      // api_key, azure_ad, aws_sigv4, bearer
	ProjectID     string            `mapstructure:"project_id"`     // For Vertex AI
	EndpointID    string            `mapstructure:"endpoint_id"`    // For Azure OpenAI
	CustomHeaders map[string]string `mapstructure:"custom_headers"` // Provider-specific headers
	// AWS-specific
	AWSRegion    string `mapstructure:"aws_region"`
	AWSAccessKey string `mapstructure:"aws_access_key"`
	AWSSecretKey string `mapstructure:"aws_secret_key"`
	// Rate limiting
	RPM           int64   `mapstructure:"rpm"`
	TPM           int64   `mapstructure:"tpm"`
	Weight        float64 `mapstructure:"weight"`
	Timeout       float64 `mapstructure:"timeout"`
	StreamTimeout float64 `mapstructure:"stream_timeout"`
	RegionName    string  `mapstructure:"region_name"`
}

// RouterSettings represents router-specific settings.
type RouterSettings struct {
	RoutingStrategy     string  `mapstructure:"routing_strategy"`
	RedisHost           string  `mapstructure:"redis_host"`
	RedisPort           int     `mapstructure:"redis_port"`
	RedisPassword       string  `mapstructure:"redis_password"`
	EnablePreCallChecks bool    `mapstructure:"enable_pre_call_checks"`
	CooldownTime        float64 `mapstructure:"cooldown_time"`
	NumRetries          int     `mapstructure:"num_retries"`
}

// TenantSettings controls tenant resolution.
type TenantSettings struct {
	BaseDomain      string `mapstructure:"base_domain"`
	AllowHeaderHint bool   `mapstructure:"allow_header_hint"`
}

// TunnelSettings controls the local tunnel client integration.
type TunnelSettings struct {
	ClientURL string `mapstructure:"client_url"`
	Timeout   int    `mapstructure:"timeout_seconds"`
}

// BillingSettings controls exchange rate and pricing behavior.
type BillingSettings struct {
	DefaultUSDtoIRR float64 `mapstructure:"default_usd_to_irr"`
	RateSource      string  `mapstructure:"rate_source"`
}

// GeneralSettings represents general proxy settings.
type GeneralSettings struct {
	MasterKey                   string `mapstructure:"master_key"`
	DatabaseURL                 string `mapstructure:"database_url"`
	Host                        string `mapstructure:"host"`
	Port                        int    `mapstructure:"port"`
	ProxyBatchWriteAt           int    `mapstructure:"proxy_batch_write_at"`
	DatabaseConnectionPoolLimit int    `mapstructure:"database_connection_pool_limit"`
}

// LoadConfig loads the configuration from YAML file and environment variables.
func LoadConfig(logger *zap.Logger) (*Config, error) {
	v := viper.New()
	v.SetConfigName("proxy_config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./configs")
	v.AddConfigPath("/etc/shahrah")

	// Set defaults
	v.SetDefault("general_settings.host", "0.0.0.0")
	v.SetDefault("general_settings.port", 4000)
	v.SetDefault("router_settings.routing_strategy", "simple-shuffle")
	v.SetDefault("router_settings.cooldown_time", 60.0)
	v.SetDefault("router_settings.num_retries", 3)
	v.SetDefault("response_cache.enabled", false)
	v.SetDefault("response_cache.key_prefix", "shahrah:response-cache")
	v.SetDefault("response_cache.default_ttl_seconds", 600)
	v.SetDefault("response_cache.min_ttl_seconds", 1)
	v.SetDefault("response_cache.max_ttl_seconds", 86400)
	v.SetDefault("response_cache.max_response_bytes", 1048576)
	v.SetDefault("response_cache.max_semantic_candidates", 500)
	v.SetDefault("response_cache.embedding_provider", "arvancloud")
	v.SetDefault("response_cache.embedding_model", "Bge-m3")
	v.SetDefault("response_cache.embedding_revision", "v1")
	v.SetDefault("response_cache.embedding_dimensions", 1024)
	v.SetDefault("response_cache.embedding_timeout_seconds", 10)
	v.SetDefault("response_cache.embedding_max_batch_size", 16)
	v.SetDefault("tenant_settings.base_domain", "shahrah.local")
	v.SetDefault("tenant_settings.allow_header_hint", true)
	v.SetDefault("tunnel_settings.client_url", "http://127.0.0.1:8081")
	v.SetDefault("tunnel_settings.timeout_seconds", 300)
	v.SetDefault("billing_settings.default_usd_to_irr", 850000.0)
	v.SetDefault("billing_settings.rate_source", "static")

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		logger.Warn("No config file found, using environment variables only")
	}

	// Bind environment variables
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Resolve environment variable references in config
	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Resolve ${ENV_VAR} references in string fields
	if err := resolveEnvVars(cfg, logger); err != nil {
		return nil, fmt.Errorf("error resolving environment variables: %w", err)
	}

	return cfg, nil
}

// resolveEnvVars resolves ${ENV_VAR} references in configuration strings.
func resolveEnvVars(cfg *Config, logger *zap.Logger) error {
	re := regexp.MustCompile(`\$\{([^}]+)\}`)

	resolveString := func(s string) string {
		return re.ReplaceAllStringFunc(s, func(match string) string {
			key := match[2 : len(match)-1]
			val := os.Getenv(key)
			if val == "" {
				logger.Warn("Environment variable not set", zap.String("key", key))
			}
			return val
		})
	}

	// Resolve model deployment params
	for i := range cfg.ModelList {
		cfg.ModelList[i].ProviderEndpoint.APIKey = resolveString(cfg.ModelList[i].ProviderEndpoint.APIKey)
		cfg.ModelList[i].ProviderEndpoint.APIBase = resolveString(cfg.ModelList[i].ProviderEndpoint.APIBase)
	}

	cfg.RouterSettings.RedisHost = resolveString(cfg.RouterSettings.RedisHost)
	cfg.RouterSettings.RedisPassword = resolveString(cfg.RouterSettings.RedisPassword)
	cfg.ResponseCache.EmbeddingBaseURL = resolveString(cfg.ResponseCache.EmbeddingBaseURL)
	cfg.ResponseCache.EmbeddingAPIKey = resolveString(cfg.ResponseCache.EmbeddingAPIKey)
	cfg.GeneralSettings.MasterKey = resolveString(cfg.GeneralSettings.MasterKey)
	cfg.GeneralSettings.DatabaseURL = resolveString(cfg.GeneralSettings.DatabaseURL)
	cfg.TunnelSettings.ClientURL = resolveString(cfg.TunnelSettings.ClientURL)

	return nil
}
