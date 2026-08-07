package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/observability/logretention"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/metrics"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
)

// ObservabilitySettings contains immutable tracing, metrics, and retention settings.
type ObservabilitySettings struct {
	Tracing          tracing.Config
	Metrics          metrics.Config
	MetricsReconcile metrics.ReconcileConfig
	LogRetention     logretention.Config
}

// LoadObservabilitySettings reads observability environment once at the configuration edge.
func LoadObservabilitySettings() ObservabilitySettings {
	return ObservabilitySettings{
		Tracing:          loadTracingSettings(),
		Metrics:          loadMetricsSettings(),
		MetricsReconcile: loadMetricsReconcileSettings(),
		LogRetention:     loadLogRetentionSettings(),
	}
}

func loadTracingSettings() tracing.Config {
	cfg := tracing.DefaultConfig()
	cfg.DefaultSampleRate = unitFloatEnv("TRACE_DEFAULT_SAMPLE_RATE", cfg.DefaultSampleRate)
	cfg.SlowRequest = millisecondsEnv("TRACE_SLOW_REQUEST_MS", cfg.SlowRequest)
	cfg.QueueCapacity = positiveIntEnv("TRACE_QUEUE_CAPACITY", cfg.QueueCapacity)
	cfg.BatchSize = positiveIntEnv("TRACE_BATCH_SIZE", cfg.BatchSize)
	cfg.ExportTimeout = millisecondsEnv("TRACE_EXPORT_TIMEOUT_MS", cfg.ExportTimeout)
	cfg.OTLPEndpoint = strings.TrimSpace(os.Getenv("TRACE_OTLP_ENDPOINT"))
	cfg.OTLPInsecure = booleanEnv("TRACE_OTLP_INSECURE", cfg.OTLPInsecure)
	cfg.OTLPServiceName = stringEnv("TRACE_OTLP_SERVICE_NAME", cfg.OTLPServiceName)
	cfg.RetryMaxAttempts = positiveIntEnv("TRACE_RETRY_MAX_ATTEMPTS", cfg.RetryMaxAttempts)
	cfg.RetryBaseDelay = millisecondsEnv("TRACE_RETRY_BASE_DELAY_MS", cfg.RetryBaseDelay)
	cfg.RetryMaxDelay = millisecondsEnv("TRACE_RETRY_MAX_DELAY_MS", cfg.RetryMaxDelay)
	cfg.WALEnabled = booleanEnv("TRACE_WAL_ENABLED", cfg.WALEnabled)
	cfg.WALPath = stringEnv("TRACE_WAL_PATH", cfg.WALPath)
	cfg.WALReplayInterval = millisecondsEnv("TRACE_WAL_REPLAY_INTERVAL_MS", cfg.WALReplayInterval)
	cfg.CapturePolicy = tracing.ParseCapturePolicy(stringEnv("TRACE_CAPTURE_POLICY", string(cfg.CapturePolicy)))
	cfg.PayloadBlobDir = stringEnv("TRACE_PAYLOAD_BLOB_DIR", cfg.PayloadBlobDir)
	cfg.PayloadEncryptKey = strings.TrimSpace(os.Getenv("TRACE_PAYLOAD_ENCRYPTION_KEY"))
	cfg.PayloadMaxBytes = intAboveMinimumEnv("TRACE_PAYLOAD_MAX_BYTES", cfg.PayloadMaxBytes, 256)
	cfg.PIIHashSalt = stringEnv("TRACE_PII_HASH_SALT", cfg.PIIHashSalt)

	return cfg
}

func loadMetricsSettings() metrics.Config {
	cfg := metrics.DefaultConfig()
	cfg.Enabled = booleanEnv("METRICS_MODEL_ENABLED", cfg.Enabled)
	cfg.QueueCapacity = positiveIntEnv("METRICS_MODEL_QUEUE_CAPACITY", cfg.QueueCapacity)
	cfg.BatchSize = positiveIntEnv("METRICS_MODEL_BATCH_SIZE", cfg.BatchSize)
	cfg.FlushInterval = millisecondsEnv("METRICS_MODEL_FLUSH_INTERVAL_MS", cfg.FlushInterval)
	cfg.ExportTimeout = millisecondsEnv("METRICS_MODEL_EXPORT_TIMEOUT_MS", cfg.ExportTimeout)
	cfg.RetryMaxAttempts = positiveIntEnv("METRICS_MODEL_RETRY_MAX_ATTEMPTS", cfg.RetryMaxAttempts)
	cfg.RetryBaseDelay = millisecondsEnv("METRICS_MODEL_RETRY_BASE_DELAY_MS", cfg.RetryBaseDelay)
	cfg.RetryMaxDelay = millisecondsEnv("METRICS_MODEL_RETRY_MAX_DELAY_MS", cfg.RetryMaxDelay)
	cfg.DLQ = loadMetricsDLQSettings()

	return cfg
}

func loadMetricsDLQSettings() metrics.DLQConfig {
	cfg := metrics.DefaultDLQConfig()
	cfg.Backend = strings.ToLower(stringEnv("METRICS_MODEL_DLQ_BACKEND", cfg.Backend))
	cfg.Address = strings.TrimSpace(os.Getenv("METRICS_MODEL_DLQ_REDIS_ADDR"))
	cfg.SentinelAddrs = csvEnv("METRICS_MODEL_DLQ_REDIS_SENTINEL_ADDRS")
	cfg.MasterName = stringEnv("METRICS_MODEL_DLQ_REDIS_MASTER_NAME", cfg.MasterName)
	cfg.Username = strings.TrimSpace(os.Getenv("METRICS_MODEL_DLQ_REDIS_USERNAME"))
	cfg.Password = os.Getenv("METRICS_MODEL_DLQ_REDIS_PASSWORD")
	cfg.SentinelPassword = os.Getenv("METRICS_MODEL_DLQ_REDIS_SENTINEL_PASSWORD")
	cfg.Stream = stringEnv("METRICS_MODEL_DLQ_STREAM", cfg.Stream)
	cfg.PoisonStream = stringEnv("METRICS_MODEL_DLQ_POISON_STREAM", cfg.PoisonStream)
	cfg.Group = stringEnv("METRICS_MODEL_DLQ_GROUP", cfg.Group)
	cfg.MaxLen = positiveInt64Env("METRICS_MODEL_DLQ_MAXLEN", cfg.MaxLen)
	cfg.Retention = positiveDurationEnv("METRICS_MODEL_DLQ_RETENTION", cfg.Retention)
	cfg.EnqueueTimeout = millisecondsEnv("METRICS_MODEL_DLQ_ENQUEUE_TIMEOUT_MS", cfg.EnqueueTimeout)
	cfg.ReadBatchSize = positiveInt64Env("METRICS_MODEL_DLQ_READ_BATCH_SIZE", cfg.ReadBatchSize)
	cfg.ReadBlock = positiveDurationEnv("METRICS_MODEL_DLQ_READ_BLOCK", cfg.ReadBlock)
	cfg.ClaimMinIdle = positiveDurationEnv("METRICS_MODEL_DLQ_CLAIM_MIN_IDLE", cfg.ClaimMinIdle)

	return cfg
}

func loadMetricsReconcileSettings() metrics.ReconcileConfig {
	cfg := metrics.DefaultReconcileConfig()
	cfg.Enabled = booleanEnv("METRICS_RECONCILE_ENABLED", cfg.Enabled)

	return cfg
}

func loadLogRetentionSettings() logretention.Config {
	cfg := logretention.DefaultConfig()
	cfg.CleanupEnabled = booleanEnv("LOG_RETENTION_CLEANUP_ENABLED", cfg.CleanupEnabled)
	cfg.CleanupInterval = millisecondsEnv("LOG_RETENTION_CLEANUP_INTERVAL_MS", cfg.CleanupInterval)
	cfg.CleanupTimeout = millisecondsEnv("LOG_RETENTION_CLEANUP_TIMEOUT_MS", cfg.CleanupTimeout)

	return cfg
}

func millisecondsEnv(key string, fallback time.Duration) time.Duration {
	value := positiveIntEnv(key, int(fallback/time.Millisecond))

	return time.Duration(value) * time.Millisecond
}

func unitFloatEnv(key string, fallback float64) float64 {
	raw := strings.TrimSpace(os.Getenv(key))

	value, err := strconv.ParseFloat(raw, 64)
	if err != nil || value < 0 || value > 1 {
		return fallback
	}

	return value
}

func intAboveMinimumEnv(key string, fallback int, minimum int) int {
	value := positiveIntEnv(key, fallback)
	if value <= minimum {
		return fallback
	}

	return value
}

func positiveInt64Env(key string, fallback int64) int64 {
	raw := strings.TrimSpace(os.Getenv(key))

	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value <= 0 {
		return fallback
	}

	return value
}

func positiveDurationEnv(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))

	value, err := time.ParseDuration(raw)
	if err != nil || value <= 0 {
		return fallback
	}

	return value
}
