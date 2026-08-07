package metrics

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/redis/go-redis/v9"
)

const (
	defaultDLQStream    = "shahrah:observability:model-metrics:dlq:v1"
	defaultPoisonStream = "shahrah:observability:model-metrics:poison:v1"
	defaultDLQGroup     = "metrics-dlq-replayer-v1"
)

var (
	dlqEnqueuedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "shahrah_model_metrics_dlq_enqueued_total",
		Help: "Number of failed model metric events added to the Redis Stream DLQ.",
	})
	dlqEnqueueErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "shahrah_model_metrics_dlq_enqueue_errors_total",
		Help: "Number of model metric events that could not be added to the Redis Stream DLQ.",
	})
	dlqDroppedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "shahrah_model_metrics_dlq_dropped_total",
		Help: "Number of model metric events dropped after both ClickHouse and the DLQ were unavailable.",
	})
)

// DLQConfig controls the Redis Stream used for durable model-metric fallback.
type DLQConfig struct {
	Backend          string
	Address          string
	SentinelAddrs    []string
	MasterName       string
	Username         string
	Password         string
	SentinelPassword string
	Stream           string
	PoisonStream     string
	Group            string
	MaxLen           int64
	Retention        time.Duration
	EnqueueTimeout   time.Duration
	ReadBatchSize    int64
	ReadBlock        time.Duration
	ClaimMinIdle     time.Duration
}

// DefaultDLQConfig returns durable queue defaults without reading process-global state.
func DefaultDLQConfig() DLQConfig {
	return DLQConfig{
		Backend:        "disabled",
		MasterName:     "mymaster",
		Stream:         defaultDLQStream,
		PoisonStream:   defaultPoisonStream,
		Group:          defaultDLQGroup,
		MaxLen:         10_000,
		Retention:      7 * 24 * time.Hour,
		EnqueueTimeout: 250 * time.Millisecond,
		ReadBatchSize:  50,
		ReadBlock:      5 * time.Second,
		ClaimMinIdle:   60 * time.Second,
	}
}

// NewRedisStreamDLQ creates and verifies the configured Redis connection.
func NewRedisStreamDLQ(ctx context.Context, cfg DLQConfig) (*RedisStreamDLQ, error) {
	if cfg.Backend != "redis_stream" {
		return nil, fmt.Errorf("unsupported model metrics dlq backend %q", cfg.Backend)
	}

	client, err := newRedisClient(cfg)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx).Err(); err != nil {
		if closeErr := client.Close(); closeErr != nil {
			return nil, fmt.Errorf("ping redis dlq: %w; close failed: %v", err, closeErr)
		}

		return nil, fmt.Errorf("ping redis dlq: %w", err)
	}

	return &RedisStreamDLQ{cfg: cfg, client: client}, nil
}

func newRedisClient(cfg DLQConfig) (redis.UniversalClient, error) {
	if len(cfg.SentinelAddrs) > 0 {
		if strings.TrimSpace(cfg.MasterName) == "" {
			return nil, errors.New("redis sentinel master name is required")
		}

		return redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:            cfg.MasterName,
			SentinelAddrs:         cfg.SentinelAddrs,
			Username:              cfg.Username,
			Password:              cfg.Password,
			SentinelPassword:      cfg.SentinelPassword,
			DialTimeout:           2 * time.Second,
			ReadTimeout:           2 * time.Second,
			WriteTimeout:          250 * time.Millisecond,
			PoolSize:              8,
			MinIdleConns:          1,
			MaxRetries:            1,
			ContextTimeoutEnabled: true,
		}), nil
	}

	if strings.TrimSpace(cfg.Address) == "" {
		return nil, errors.New("redis address or sentinel addresses are required")
	}

	return redis.NewClient(&redis.Options{
		Addr:                  cfg.Address,
		Username:              cfg.Username,
		Password:              cfg.Password,
		DialTimeout:           2 * time.Second,
		ReadTimeout:           2 * time.Second,
		WriteTimeout:          250 * time.Millisecond,
		PoolSize:              8,
		MinIdleConns:          1,
		MaxRetries:            1,
		ContextTimeoutEnabled: true,
	}), nil
}

// RedisStreamDLQ stores failed metric events in a bounded Redis Stream.
type RedisStreamDLQ struct {
	cfg    DLQConfig
	client redis.UniversalClient
}

func (d *RedisStreamDLQ) Enqueue(ctx context.Context, event platform.ModelMetricEvent, cause error) error {
	if strings.TrimSpace(event.EventID) == "" {
		return errors.New("model metric event_id is required")
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal model metric event: %w", err)
	}

	values := map[string]interface{}{
		"event_id":        event.EventID,
		"payload":         string(payload),
		"first_failed_at": time.Now().UTC().Format(time.RFC3339Nano),
	}
	if cause != nil {
		values["last_error"] = truncateError(cause.Error(), 1024)
	}

	if err := d.client.XAdd(ctx, &redis.XAddArgs{
		Stream: d.cfg.Stream,
		MaxLen: d.cfg.MaxLen,
		Approx: true,
		Values: values,
	}).Err(); err != nil {
		dlqEnqueueErrors.Inc()
		return fmt.Errorf("xadd model metrics dlq: %w", err)
	}

	dlqEnqueuedTotal.Inc()

	if d.cfg.Retention > 0 {
		minID := fmt.Sprintf("%d-0", time.Now().Add(-d.cfg.Retention).UnixMilli())
		// The event is already durable at this point. Retention trimming is best
		// effort so a slow trim cannot turn a successful enqueue into a false
		// failure at the request boundary.
		if err := d.client.XTrimMinIDApprox(ctx, d.cfg.Stream, minID, 0).Err(); err != nil {
			dlqEnqueueErrors.Inc()
		}
	}

	return nil
}

func (d *RedisStreamDLQ) Close() error {
	if d == nil || d.client == nil {
		return nil
	}

	return d.client.Close()
}

func truncateError(value string, limit int) string {
	if len(value) <= limit {
		return value
	}

	return value[:limit]
}
