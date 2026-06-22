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

	RetentionCleanupEnabled  bool
	RetentionCleanupInterval time.Duration
	RetentionCleanupTimeout  time.Duration
}

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

		RetentionCleanupEnabled:  true,
		RetentionCleanupInterval: time.Hour,
		RetentionCleanupTimeout:  30 * time.Second,
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
	if v := strings.TrimSpace(os.Getenv("TRACE_OTLP_SERVICE_NAME")); v != "" {
		cfg.OTLPServiceName = v
	}
	if v := os.Getenv("TRACE_RETRY_MAX_ATTEMPTS"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			cfg.RetryMaxAttempts = parsed
		}
	}
	if v := os.Getenv("TRACE_RETRY_BASE_DELAY_MS"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			cfg.RetryBaseDelay = time.Duration(parsed) * time.Millisecond
		}
	}
	if v := os.Getenv("TRACE_RETRY_MAX_DELAY_MS"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			cfg.RetryMaxDelay = time.Duration(parsed) * time.Millisecond
		}
	}
	if v := strings.TrimSpace(os.Getenv("TRACE_WAL_PATH")); v != "" {
		cfg.WALPath = v
	}
	if v := os.Getenv("TRACE_WAL_ENABLED"); v != "" {
		cfg.WALEnabled = strings.EqualFold(strings.TrimSpace(v), "true")
	}
	if v := os.Getenv("TRACE_WAL_REPLAY_INTERVAL_MS"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			cfg.WALReplayInterval = time.Duration(parsed) * time.Millisecond
		}
	}
	if v := strings.TrimSpace(os.Getenv("TRACE_CAPTURE_POLICY")); v != "" {
		cfg.CapturePolicy = ParseCapturePolicy(v)
	}
	if v := strings.TrimSpace(os.Getenv("TRACE_PAYLOAD_BLOB_DIR")); v != "" {
		cfg.PayloadBlobDir = v
	}
	if v := os.Getenv("TRACE_PAYLOAD_MAX_BYTES"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 256 {
			cfg.PayloadMaxBytes = parsed
		}
	}
	if v := strings.TrimSpace(os.Getenv("TRACE_RETENTION_CLEANUP_ENABLED")); v != "" {
		cfg.RetentionCleanupEnabled = parseEnvBool(v)
	}
	if v := os.Getenv("TRACE_RETENTION_CLEANUP_INTERVAL_MS"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			cfg.RetentionCleanupInterval = time.Duration(parsed) * time.Millisecond
		}
	}
	if v := os.Getenv("TRACE_RETENTION_CLEANUP_TIMEOUT_MS"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			cfg.RetentionCleanupTimeout = time.Duration(parsed) * time.Millisecond
		}
	}
	return cfg
}

func parseEnvBool(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}
