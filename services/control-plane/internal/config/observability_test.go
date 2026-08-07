package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLoadObservabilitySettingsParsesTypedValues(t *testing.T) {
	t.Setenv("TRACE_DEFAULT_SAMPLE_RATE", "0.25")
	t.Setenv("TRACE_CAPTURE_POLICY", "full-secure")
	t.Setenv("METRICS_MODEL_ENABLED", "true")
	t.Setenv("METRICS_MODEL_DLQ_BACKEND", "redis_stream")
	t.Setenv("METRICS_MODEL_DLQ_REDIS_SENTINEL_ADDRS", "redis-0:26379, redis-1:26379")
	t.Setenv("METRICS_MODEL_DLQ_RETENTION", "168h")
	t.Setenv("LOG_RETENTION_CLEANUP_TIMEOUT_MS", "45000")

	settings := LoadObservabilitySettings()

	require.InDelta(t, 0.25, settings.Tracing.DefaultSampleRate, 0.0001)
	require.Equal(t, "full-secure", string(settings.Tracing.CapturePolicy))
	require.True(t, settings.Metrics.Enabled)
	require.Equal(t, "redis_stream", settings.Metrics.DLQ.Backend)
	require.Equal(t, []string{"redis-0:26379", "redis-1:26379"}, settings.Metrics.DLQ.SentinelAddrs)
	require.Equal(t, 7*24*time.Hour, settings.Metrics.DLQ.Retention)
	require.Equal(t, 45*time.Second, settings.LogRetention.CleanupTimeout)
}

func TestLoadObservabilitySettingsRejectsInvalidOverrides(t *testing.T) {
	t.Setenv("TRACE_DEFAULT_SAMPLE_RATE", "2")
	t.Setenv("METRICS_MODEL_QUEUE_CAPACITY", "0")
	t.Setenv("LOG_RETENTION_CLEANUP_INTERVAL_MS", "invalid")

	settings := LoadObservabilitySettings()

	require.InDelta(t, 0.02, settings.Tracing.DefaultSampleRate, 0.0001)
	require.Equal(t, 8192, settings.Metrics.QueueCapacity)
	require.Equal(t, time.Hour, settings.LogRetention.CleanupInterval)
}
