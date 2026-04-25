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
	ModelList       []ModelDeployment `mapstructure:"model_list"`
	RouterSettings  RouterSettings    `mapstructure:"router_settings"`
	GeneralSettings GeneralSettings   `mapstructure:"general_settings"`
}

// ModelDeployment represents a single model deployment configuration.
type ModelDeployment struct {
	ModelName     string                 `mapstructure:"model_name"`
	LiteLLMParams LiteLLMParams          `mapstructure:"litellm_params"`
	ModelInfo     map[string]interface{} `mapstructure:"model_info"`
}

// LiteLLMParams represents the parameters for a LiteLLM call.
type LiteLLMParams struct {
	Model         string  `mapstructure:"model"`
	APIKey        string  `mapstructure:"api_key"`
	APIBase       string  `mapstructure:"api_base"`
	APIVersion    string  `mapstructure:"api_version"`
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
		cfg.ModelList[i].LiteLLMParams.APIKey = resolveString(cfg.ModelList[i].LiteLLMParams.APIKey)
		cfg.ModelList[i].LiteLLMParams.APIBase = resolveString(cfg.ModelList[i].LiteLLMParams.APIBase)
	}

	cfg.RouterSettings.RedisHost = resolveString(cfg.RouterSettings.RedisHost)
	cfg.RouterSettings.RedisPassword = resolveString(cfg.RouterSettings.RedisPassword)
	cfg.GeneralSettings.MasterKey = resolveString(cfg.GeneralSettings.MasterKey)
	cfg.GeneralSettings.DatabaseURL = resolveString(cfg.GeneralSettings.DatabaseURL)

	return nil
}
