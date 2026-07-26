package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadDLQConfigFromEnvUsesProductionStreamDefaults(t *testing.T) {
	t.Setenv("METRICS_MODEL_DLQ_BACKEND", "redis_stream")
	t.Setenv("METRICS_MODEL_DLQ_REDIS_SENTINEL_ADDRS", "redis-0:26379, redis-1:26379")
	t.Setenv("METRICS_MODEL_DLQ_MAXLEN", "10000")
	t.Setenv("METRICS_MODEL_DLQ_RETENTION", "168h")
	t.Setenv("METRICS_MODEL_DLQ_ENQUEUE_TIMEOUT_MS", "250")

	cfg := LoadConfigFromEnv().DLQ

	assert.Equal(t, "redis_stream", cfg.Backend)
	assert.Equal(t, []string{"redis-0:26379", "redis-1:26379"}, cfg.SentinelAddrs)
	assert.Equal(t, defaultDLQStream, cfg.Stream)
	assert.Equal(t, defaultPoisonStream, cfg.PoisonStream)
	assert.Equal(t, int64(10_000), cfg.MaxLen)
	assert.Equal(t, 7*24*time.Hour, cfg.Retention)
	assert.Equal(t, 250*time.Millisecond, cfg.EnqueueTimeout)
}
