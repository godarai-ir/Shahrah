// Package config loads and validates control-plane configuration.
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
	RouterSettings   RouterSettings           `mapstructure:"router_settings"`
	ResponseCache    ResponseCacheSettings    `mapstructure:"response_cache"`
	SemanticJudge    SemanticJudgeSettings    `mapstructure:"semantic_guardrails"`
	GeneralSettings  GeneralSettings          `mapstructure:"general_settings"`
	Profiling        ProfilingSettings        `mapstructure:"profiling"`
	OperationalLogs  OperationalLogSettings   `mapstructure:"operational_logging"`
	RequestLogs      RequestLogSettings       `mapstructure:"request_logs"`
	SpendLogs        SpendLogSettings         `mapstructure:"spend_logs"`
	TenantSettings   TenantSettings           `mapstructure:"tenant_settings"`
	ProviderEgress   ProviderEgressSettings   `mapstructure:"provider_egress"`
	BillingSettings  BillingSettings          `mapstructure:"billing_settings"`
	Storage          StorageSettings          `mapstructure:"-"`
	Runtime          RuntimeSettings          `mapstructure:"-"`
	Observability    ObservabilitySettings    `mapstructure:"-"`
	Admin            AdminSettings            `mapstructure:"-"`
	Credentials      CredentialSettings       `mapstructure:"-"`
	SupportFiles     SupportFileSettings      `mapstructure:"-"`
	ProviderFeatures ProviderFeatureSettings  `mapstructure:"-"`
	OIDC             OIDCSettings             `mapstructure:"-"`
	ChatProvisioning ChatProvisioningSettings `mapstructure:"-"`
}

// ProfilingSettings controls opt-in runtime profiling endpoints and sampling.
type ProfilingSettings struct {
	EnablePprof          bool `mapstructure:"enable_pprof"`
	MutexProfileFraction int  `mapstructure:"mutex_profile_fraction"`
	BlockProfileRate     int  `mapstructure:"block_profile_rate"`
}

// RequestLogSettings controls tenant-facing product request log ingestion.
type RequestLogSettings struct {
	AsyncEnabled                bool `mapstructure:"async_enabled"`
	QueueSize                   int  `mapstructure:"queue_size"`
	Workers                     int  `mapstructure:"workers"`
	BatchSize                   int  `mapstructure:"batch_size"`
	FlushIntervalMS             int  `mapstructure:"flush_interval_ms"`
	EnqueueTimeoutMS            int  `mapstructure:"enqueue_timeout_ms"`
	DropOnFull                  bool `mapstructure:"drop_on_full"`
	ShutdownDrainTimeoutSeconds int  `mapstructure:"shutdown_drain_timeout_seconds"`
	ModeCacheTTLSeconds         int  `mapstructure:"mode_cache_ttl_seconds"`
}

// SpendLogSettings controls asynchronous spend log persistence.
type SpendLogSettings struct {
	AsyncEnabled                bool `mapstructure:"async_enabled"`
	QueueSize                   int  `mapstructure:"queue_size"`
	Workers                     int  `mapstructure:"workers"`
	BatchSize                   int  `mapstructure:"batch_size"`
	FlushIntervalMS             int  `mapstructure:"flush_interval_ms"`
	EnqueueTimeoutMS            int  `mapstructure:"enqueue_timeout_ms"`
	DropOnFull                  bool `mapstructure:"drop_on_full"`
	ShutdownDrainTimeoutSeconds int  `mapstructure:"shutdown_drain_timeout_seconds"`
}

// OperationalLogSettings controls low-overhead server access/error logging.
type OperationalLogSettings struct {
	Enabled              bool    `mapstructure:"enabled"`
	SuccessSampleRate    float64 `mapstructure:"success_sample_rate"`
	SlowRequestMS        int     `mapstructure:"slow_request_ms"`
	QueueSize            int     `mapstructure:"queue_size"`
	IncludeHeaders       bool    `mapstructure:"include_headers"`
	LogClientErrors      bool    `mapstructure:"log_client_errors"`
	UseFiberAccessLogger bool    `mapstructure:"use_fiber_access_logger"`
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

// SemanticJudgeSettings configures the TypeSafe Jev judge behind semantic_policy guardrails. The judge is disabled
// until an API key is supplied, and semantic guardrails then fail as execution errors.
type SemanticJudgeSettings struct {
	Enabled          bool   `mapstructure:"enabled"`
	BaseURL          string `mapstructure:"base_url"`
	APIKey           string `mapstructure:"api_key"`
	Model            string `mapstructure:"model"`
	TimeoutMS        int    `mapstructure:"timeout_ms"`
	CacheTTLSeconds  int    `mapstructure:"cache_ttl_seconds"`
	CacheKeyPrefix   string `mapstructure:"cache_key_prefix"`
	FailureThreshold int    `mapstructure:"failure_threshold"`
	CooldownSeconds  int    `mapstructure:"cooldown_seconds"`
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
	RecognizerID  string            `mapstructure:"recognizer_id"`  // For Google Speech-to-Text
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
	RoutingStrategy string  `mapstructure:"routing_strategy"`
	RedisHost       string  `mapstructure:"redis_host"`
	RedisPort       int     `mapstructure:"redis_port"`
	RedisPassword   string  `mapstructure:"redis_password"`
	CooldownTime    float64 `mapstructure:"cooldown_time"`
	NumRetries      int     `mapstructure:"num_retries"`
}

// TenantSettings controls tenant resolution.
type TenantSettings struct {
	AllowHeaderHint bool `mapstructure:"allow_header_hint"`
}

// ProviderEgressSettings controls the direct Silk Road integration.
type ProviderEgressSettings struct {
	URL     string `mapstructure:"url"`
	Timeout int    `mapstructure:"timeout_seconds"`
}

// BillingSettings controls exchange rate and pricing behavior.
type BillingSettings struct {
	DefaultUSDtoIRR float64 `mapstructure:"default_usd_to_irr"`
	RateSource      string  `mapstructure:"rate_source"`
}

// GeneralSettings represents general proxy settings.
type GeneralSettings struct {
	MasterKey   string `mapstructure:"master_key"`
	DatabaseURL string `mapstructure:"database_url"`
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
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
	v.SetDefault("profiling.enable_pprof", false)
	v.SetDefault("profiling.mutex_profile_fraction", 0)
	v.SetDefault("profiling.block_profile_rate", 0)
	v.SetDefault("operational_logging.enabled", true)
	v.SetDefault("operational_logging.success_sample_rate", 0.0)
	v.SetDefault("operational_logging.slow_request_ms", 1000)
	v.SetDefault("operational_logging.queue_size", 4096)
	v.SetDefault("operational_logging.include_headers", false)
	v.SetDefault("operational_logging.log_client_errors", false)
	v.SetDefault("operational_logging.use_fiber_access_logger", false)
	v.SetDefault("request_logs.async_enabled", true)
	v.SetDefault("request_logs.queue_size", 10000)
	v.SetDefault("request_logs.workers", 1)
	v.SetDefault("request_logs.batch_size", 500)
	v.SetDefault("request_logs.flush_interval_ms", 500)
	v.SetDefault("request_logs.enqueue_timeout_ms", 0)
	v.SetDefault("request_logs.drop_on_full", true)
	v.SetDefault("request_logs.shutdown_drain_timeout_seconds", 5)
	v.SetDefault("request_logs.mode_cache_ttl_seconds", 60)
	v.SetDefault("spend_logs.async_enabled", true)
	v.SetDefault("spend_logs.queue_size", 10000)
	v.SetDefault("spend_logs.workers", 1)
	v.SetDefault("spend_logs.batch_size", 500)
	v.SetDefault("spend_logs.flush_interval_ms", 500)
	v.SetDefault("spend_logs.enqueue_timeout_ms", 0)
	v.SetDefault("spend_logs.drop_on_full", true)
	v.SetDefault("spend_logs.shutdown_drain_timeout_seconds", 5)
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
	v.SetDefault("semantic_guardrails.enabled", false)
	v.SetDefault("semantic_guardrails.base_url", "https://api.typesafe.ai")
	v.SetDefault("semantic_guardrails.model", "jev-1.13.0")
	v.SetDefault("semantic_guardrails.timeout_ms", 1500)
	v.SetDefault("semantic_guardrails.cache_ttl_seconds", 600)
	v.SetDefault("semantic_guardrails.cache_key_prefix", "shahrah:semantic-judge:")
	v.SetDefault("semantic_guardrails.failure_threshold", 5)
	v.SetDefault("semantic_guardrails.cooldown_seconds", 30)
	v.SetDefault("tenant_settings.allow_header_hint", true)
	v.SetDefault("provider_egress.url", "http://127.0.0.1:8083")
	v.SetDefault("provider_egress.timeout_seconds", 300)
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

	cfg.Storage = LoadStorageSettings()
	cfg.Runtime = LoadRuntimeSettings()
	cfg.Credentials = LoadCredentialSettings()
	cfg.SupportFiles = LoadSupportFileSettings()
	cfg.ProviderFeatures = LoadProviderFeatureSettings()
	cfg.Observability = LoadObservabilitySettings()
	cfg.Admin = LoadAdminSettings()
	cfg.OIDC = LoadOIDCSettings()
	cfg.ChatProvisioning = LoadChatProvisioningSettings()

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

	cfg.RouterSettings.RedisHost = resolveString(cfg.RouterSettings.RedisHost)
	cfg.RouterSettings.RedisPassword = resolveString(cfg.RouterSettings.RedisPassword)
	cfg.ResponseCache.EmbeddingBaseURL = resolveString(cfg.ResponseCache.EmbeddingBaseURL)
	cfg.ResponseCache.EmbeddingAPIKey = resolveString(cfg.ResponseCache.EmbeddingAPIKey)
	cfg.SemanticJudge.APIKey = resolveString(cfg.SemanticJudge.APIKey)
	cfg.GeneralSettings.MasterKey = resolveString(cfg.GeneralSettings.MasterKey)
	cfg.GeneralSettings.DatabaseURL = resolveString(cfg.GeneralSettings.DatabaseURL)
	cfg.ProviderEgress.URL = resolveString(cfg.ProviderEgress.URL)

	return nil
}
