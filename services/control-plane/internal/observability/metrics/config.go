// Package metrics configures model-metrics collection and export.
package metrics

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config controls the metrics queue, flush, and retry behavior.
type Config struct {
	Enabled          bool
	QueueCapacity    int
	BatchSize        int
	FlushInterval    time.Duration
	ExportTimeout    time.Duration
	RetryMaxAttempts int
	RetryBaseDelay   time.Duration
	RetryMaxDelay    time.Duration
	DLQ              DLQConfig
}

// LoadConfigFromEnv builds the metrics configuration from environment variables.
func LoadConfigFromEnv() Config {
	cfg := Config{
		Enabled:          strings.EqualFold(strings.TrimSpace(os.Getenv("METRICS_MODEL_ENABLED")), "true"),
		QueueCapacity:    8192,
		BatchSize:        128,
		FlushInterval:    500 * time.Millisecond,
		ExportTimeout:    2 * time.Second,
		RetryMaxAttempts: 3,
		RetryBaseDelay:   100 * time.Millisecond,
		RetryMaxDelay:    2 * time.Second,
		DLQ:              loadDLQConfigFromEnv(),
	}
	cfg.QueueCapacity = intEnvOrDefault("METRICS_MODEL_QUEUE_CAPACITY", cfg.QueueCapacity)
	cfg.BatchSize = intEnvOrDefault("METRICS_MODEL_BATCH_SIZE", cfg.BatchSize)
	cfg.FlushInterval = durationMillisEnvOrDefault("METRICS_MODEL_FLUSH_INTERVAL_MS", cfg.FlushInterval)
	cfg.ExportTimeout = durationMillisEnvOrDefault("METRICS_MODEL_EXPORT_TIMEOUT_MS", cfg.ExportTimeout)
	cfg.RetryMaxAttempts = intEnvOrDefault("METRICS_MODEL_RETRY_MAX_ATTEMPTS", cfg.RetryMaxAttempts)
	cfg.RetryBaseDelay = durationMillisEnvOrDefault("METRICS_MODEL_RETRY_BASE_DELAY_MS", cfg.RetryBaseDelay)
	cfg.RetryMaxDelay = durationMillisEnvOrDefault("METRICS_MODEL_RETRY_MAX_DELAY_MS", cfg.RetryMaxDelay)

	return cfg
}

func intEnvOrDefault(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			return parsed
		}
	}

	return fallback
}

func durationMillisEnvOrDefault(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			return time.Duration(parsed) * time.Millisecond
		}
	}

	return fallback
}

func stringEnvOrDefault(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}

	return fallback
}
