// Package tracing provides request tracing, privacy, and export configuration.
package tracing

import (
	"os"
	"strconv"
	"strings"
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

// Config controls tracing sampling, export, privacy, and retention behavior.
type Config struct {
	DefaultSampleRate float64
	SlowRequest       time.Duration
	QueueCapacity     int
	BatchSize         int
	FlushInterval     time.Duration
	ExportTimeout     time.Duration
	OTLPEndpoint      string
	OTLPInsecure      bool
	OTLPServiceName   string
	RetryMaxAttempts  int
	RetryBaseDelay    time.Duration
	RetryMaxDelay     time.Duration
	WALEnabled        bool
	WALPath           string
	WALReplayInterval time.Duration
	CapturePolicy     CapturePolicy
	PayloadBlobDir    string
	PayloadEncryptKey string
	PayloadMaxBytes   int
	PIIHashSalt       string
}

// LoadConfigFromEnv builds the tracing configuration from environment variables.
func LoadConfigFromEnv() Config {
	cfg := Config{
		DefaultSampleRate: defaultSampleRate,
		SlowRequest:       defaultSlowRequest,
		QueueCapacity:     defaultQueueCap,
		BatchSize:         defaultBatchSize,
		FlushInterval:     defaultFlushEvery,
		ExportTimeout:     defaultExportTimout,
		OTLPEndpoint:      strings.TrimSpace(os.Getenv("TRACE_OTLP_ENDPOINT")),
		OTLPInsecure:      strings.EqualFold(strings.TrimSpace(os.Getenv("TRACE_OTLP_INSECURE")), "true"),
		OTLPServiceName:   "shahrah-control-plane",
		RetryMaxAttempts:  3,
		RetryBaseDelay:    100 * time.Millisecond,
		RetryMaxDelay:     2 * time.Second,
		WALEnabled:        true,
		WALPath:           "./tmp/trace-wal.jsonl",
		WALReplayInterval: 3 * time.Second,
		CapturePolicy:     CapturePolicyMetadataOnly,
		PayloadBlobDir:    "./tmp/trace-payload-blobs",
		PayloadEncryptKey: strings.TrimSpace(os.Getenv("TRACE_PAYLOAD_ENCRYPTION_KEY")),
		PayloadMaxBytes:   16 * 1024,
		PIIHashSalt:       strings.TrimSpace(os.Getenv("TRACE_PII_HASH_SALT")),
	}

	cfg.DefaultSampleRate = floatEnvOrDefault("TRACE_DEFAULT_SAMPLE_RATE", cfg.DefaultSampleRate)
	cfg.SlowRequest = durationMillisEnvOrDefault("TRACE_SLOW_REQUEST_MS", cfg.SlowRequest)
	cfg.QueueCapacity = intEnvOrDefault("TRACE_QUEUE_CAPACITY", cfg.QueueCapacity)
	cfg.BatchSize = intEnvOrDefault("TRACE_BATCH_SIZE", cfg.BatchSize)
	cfg.ExportTimeout = durationMillisEnvOrDefault("TRACE_EXPORT_TIMEOUT_MS", cfg.ExportTimeout)
	cfg.OTLPServiceName = stringEnvOrDefault("TRACE_OTLP_SERVICE_NAME", cfg.OTLPServiceName)
	cfg.RetryMaxAttempts = intEnvOrDefault("TRACE_RETRY_MAX_ATTEMPTS", cfg.RetryMaxAttempts)
	cfg.RetryBaseDelay = durationMillisEnvOrDefault("TRACE_RETRY_BASE_DELAY_MS", cfg.RetryBaseDelay)
	cfg.RetryMaxDelay = durationMillisEnvOrDefault("TRACE_RETRY_MAX_DELAY_MS", cfg.RetryMaxDelay)
	cfg.WALPath = stringEnvOrDefault("TRACE_WAL_PATH", cfg.WALPath)
	cfg.WALEnabled = boolEnvOrDefault("TRACE_WAL_ENABLED", cfg.WALEnabled)
	cfg.WALReplayInterval = durationMillisEnvOrDefault("TRACE_WAL_REPLAY_INTERVAL_MS", cfg.WALReplayInterval)
	cfg.CapturePolicy = capturePolicyEnvOrDefault("TRACE_CAPTURE_POLICY", cfg.CapturePolicy)
	cfg.PayloadBlobDir = stringEnvOrDefault("TRACE_PAYLOAD_BLOB_DIR", cfg.PayloadBlobDir)
	cfg.PayloadMaxBytes = intMinEnvOrDefault("TRACE_PAYLOAD_MAX_BYTES", cfg.PayloadMaxBytes, 256)
	cfg.PIIHashSalt = stringEnvOrDefault("TRACE_PII_HASH_SALT", cfg.PIIHashSalt)
	return cfg
}

func floatEnvOrDefault(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if parsed, err := strconv.ParseFloat(v, 64); err == nil && parsed >= 0 && parsed <= 1 {
			return parsed
		}
	}

	return fallback
}

func intEnvOrDefault(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			return parsed
		}
	}

	return fallback
}

func intMinEnvOrDefault(key string, fallback, minimum int) int {
	if v := os.Getenv(key); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > minimum {
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

func boolEnvOrDefault(key string, fallback bool) bool {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return strings.EqualFold(v, "true")
	}

	return fallback
}

func capturePolicyEnvOrDefault(key string, fallback CapturePolicy) CapturePolicy {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return ParseCapturePolicy(v)
	}

	return fallback
}
