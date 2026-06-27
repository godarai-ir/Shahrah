package metrics

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Enabled          bool
	QueueCapacity    int
	BatchSize        int
	FlushInterval    time.Duration
	ExportTimeout    time.Duration
	RetryMaxAttempts int
	RetryBaseDelay   time.Duration
	RetryMaxDelay    time.Duration
	DLQPath          string
}

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
		DLQPath:          "./tmp/model-metrics-dlq.jsonl",
	}
	if v := os.Getenv("METRICS_MODEL_QUEUE_CAPACITY"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			cfg.QueueCapacity = parsed
		}
	}

	if v := os.Getenv("METRICS_MODEL_BATCH_SIZE"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			cfg.BatchSize = parsed
		}
	}

	if v := os.Getenv("METRICS_MODEL_FLUSH_INTERVAL_MS"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			cfg.FlushInterval = time.Duration(parsed) * time.Millisecond
		}
	}

	if v := os.Getenv("METRICS_MODEL_EXPORT_TIMEOUT_MS"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			cfg.ExportTimeout = time.Duration(parsed) * time.Millisecond
		}
	}

	if v := os.Getenv("METRICS_MODEL_RETRY_MAX_ATTEMPTS"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			cfg.RetryMaxAttempts = parsed
		}
	}

	if v := os.Getenv("METRICS_MODEL_RETRY_BASE_DELAY_MS"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			cfg.RetryBaseDelay = time.Duration(parsed) * time.Millisecond
		}
	}

	if v := os.Getenv("METRICS_MODEL_RETRY_MAX_DELAY_MS"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			cfg.RetryMaxDelay = time.Duration(parsed) * time.Millisecond
		}
	}

	if v := strings.TrimSpace(os.Getenv("METRICS_MODEL_DLQ_PATH")); v != "" {
		cfg.DLQPath = v
	}

	return cfg
}
