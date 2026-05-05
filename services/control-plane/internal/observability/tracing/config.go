package tracing

import (
	"os"
	"strconv"
	"time"
)

const (
	defaultSampleRate   = 0.02
	defaultSlowRequest  = 1500 * time.Millisecond
	defaultQueueCap     = 8192
	defaultBatchSize    = 128
	defaultFlushEvery   = 500 * time.Millisecond
	defaultExportTimout = 2 * time.Second
)

type Config struct {
	DefaultSampleRate float64
	SlowRequest       time.Duration
	QueueCapacity     int
	BatchSize         int
	FlushInterval     time.Duration
	ExportTimeout     time.Duration
}

func LoadConfigFromEnv() Config {
	cfg := Config{
		DefaultSampleRate: defaultSampleRate,
		SlowRequest:       defaultSlowRequest,
		QueueCapacity:     defaultQueueCap,
		BatchSize:         defaultBatchSize,
		FlushInterval:     defaultFlushEvery,
		ExportTimeout:     defaultExportTimout,
	}

	if v := os.Getenv("TRACE_DEFAULT_SAMPLE_RATE"); v != "" {
		if parsed, err := strconv.ParseFloat(v, 64); err == nil && parsed >= 0 && parsed <= 1 {
			cfg.DefaultSampleRate = parsed
		}
	}
	if v := os.Getenv("TRACE_SLOW_REQUEST_MS"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			cfg.SlowRequest = time.Duration(parsed) * time.Millisecond
		}
	}
	if v := os.Getenv("TRACE_QUEUE_CAPACITY"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			cfg.QueueCapacity = parsed
		}
	}
	if v := os.Getenv("TRACE_BATCH_SIZE"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			cfg.BatchSize = parsed
		}
	}
	if v := os.Getenv("TRACE_EXPORT_TIMEOUT_MS"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			cfg.ExportTimeout = time.Duration(parsed) * time.Millisecond
		}
	}
	return cfg
}
