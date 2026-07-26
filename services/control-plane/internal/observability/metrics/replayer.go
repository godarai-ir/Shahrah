package metrics

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var (
	dlqReplayTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "shahrah_model_metrics_dlq_replayed_total",
		Help: "Number of model metric DLQ events inserted into ClickHouse and acknowledged.",
	})
	dlqReplayErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "shahrah_model_metrics_dlq_replay_errors_total",
		Help: "Number of transient model metric DLQ replay failures.",
	})
	dlqPoisonTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "shahrah_model_metrics_dlq_poison_total",
		Help: "Number of invalid model metric DLQ events moved to the poison stream.",
	})
	dlqBacklog = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "shahrah_model_metrics_dlq_stream_entries",
		Help: "Current number of entries in the model metric DLQ stream.",
	})
	dlqPending = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "shahrah_model_metrics_dlq_pending_entries",
		Help: "Current number of pending model metric DLQ entries for the consumer group.",
	})
)

// DeduplicatingExporter persists one event using event_id as the ClickHouse deduplication token.
type DeduplicatingExporter interface {
	InsertModelMetricEventWithDedupToken(ctx context.Context, event platform.ModelMetricEvent, token string) error
}

// Replayer consumes failed metrics from Redis and acknowledges only successful ClickHouse writes.
type Replayer struct {
	cfg      DLQConfig
	client   redis.UniversalClient
	exporter DeduplicatingExporter
	logger   *zap.Logger
	consumer string
}

// NewReplayer verifies Redis and creates the consumer group if needed.
func NewReplayer(ctx context.Context, cfg DLQConfig, exporter DeduplicatingExporter, logger *zap.Logger) (*Replayer, error) {
	if logger == nil {
		logger = zap.NewNop()
	}

	dlq, err := NewRedisStreamDLQ(ctx, cfg)
	if err != nil {
		return nil, err
	}

	consumer := strings.TrimSpace(os.Getenv("HOSTNAME"))
	if consumer == "" {
		consumer = "metrics-dlq-replayer"
	}

	consumer += "-" + uuid.NewString()[:8]

	replayer := &Replayer{
		cfg:      cfg,
		client:   dlq.client,
		exporter: exporter,
		logger:   logger,
		consumer: consumer,
	}
	if err := replayer.ensureGroup(ctx); err != nil {
		if closeErr := replayer.Close(); closeErr != nil {
			logger.Warn("close Redis client after consumer group initialization failure", zap.Error(closeErr))
		}

		return nil, err
	}

	return replayer, nil
}

func (r *Replayer) ensureGroup(ctx context.Context) error {
	err := r.client.XGroupCreateMkStream(ctx, r.cfg.Stream, r.cfg.Group, "0").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return fmt.Errorf("create redis stream consumer group: %w", err)
	}

	return nil
}

// Run blocks until the context is canceled.
func (r *Replayer) Run(ctx context.Context) error {
	metricsTicker := time.NewTicker(15 * time.Second)
	defer metricsTicker.Stop()

	for {
		if err := r.claimStale(ctx); err != nil && !errors.Is(err, context.Canceled) {
			r.logger.Warn("failed to claim stale model metrics dlq entries", zap.Error(err))
		}

		streams, err := r.client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    r.cfg.Group,
			Consumer: r.consumer,
			Streams:  []string{r.cfg.Stream, ">"},
			Count:    r.cfg.ReadBatchSize,
			Block:    r.cfg.ReadBlock,
		}).Result()
		if err != nil && !errors.Is(err, redis.Nil) && !errors.Is(err, context.Canceled) {
			dlqReplayErrors.Inc()
			r.logger.Warn("failed to read model metrics dlq", zap.Error(err))
		}

		for _, stream := range streams {
			for _, message := range stream.Messages {
				r.processMessage(ctx, message)
			}
		}

		select {
		case <-ctx.Done():
			return nil
		case <-metricsTicker.C:
			r.updateMetrics(ctx)
		default:
		}
	}
}

func (r *Replayer) claimStale(ctx context.Context) error {
	messages, _, err := r.client.XAutoClaim(ctx, &redis.XAutoClaimArgs{
		Stream:   r.cfg.Stream,
		Group:    r.cfg.Group,
		Consumer: r.consumer,
		MinIdle:  r.cfg.ClaimMinIdle,
		Start:    "0-0",
		Count:    r.cfg.ReadBatchSize,
	}).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return err
	}

	for _, message := range messages {
		r.processMessage(ctx, message)
	}

	return nil
}

func (r *Replayer) processMessage(ctx context.Context, message redis.XMessage) {
	payload, ok := stringValue(message.Values["payload"])
	if !ok {
		r.poisonAndAck(ctx, message, "missing payload")
		return
	}

	var event platform.ModelMetricEvent
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		r.poisonAndAck(ctx, message, "invalid json payload: "+err.Error())
		return
	}

	eventID, ok := stringValue(message.Values["event_id"])
	if !ok || strings.TrimSpace(eventID) == "" {
		eventID = event.EventID
	}

	if eventID == "" || event.EventID != eventID {
		r.poisonAndAck(ctx, message, "event_id is missing or inconsistent")
		return
	}

	parsedEventID, err := uuid.Parse(eventID)
	if err != nil || parsedEventID.Version() != uuid.Version(7) {
		r.poisonAndAck(ctx, message, "event_id is not a valid UUIDv7")
		return
	}

	err = r.insertWithTimeout(ctx, event, eventID)
	if err != nil {
		dlqReplayErrors.Inc()
		r.logger.Warn("model metrics dlq replay failed; entry remains pending",
			zap.Error(err),
			zap.String("event_id", eventID),
			zap.String("stream_id", message.ID),
		)

		return
	}

	if err := r.client.XAck(ctx, r.cfg.Stream, r.cfg.Group, message.ID).Err(); err != nil {
		dlqReplayErrors.Inc()
		r.logger.Warn("model metrics dlq ack failed", zap.Error(err), zap.String("stream_id", message.ID))

		return
	}

	if err := r.client.XDel(ctx, r.cfg.Stream, message.ID).Err(); err != nil {
		dlqReplayErrors.Inc()
		r.logger.Warn("model metrics dlq cleanup failed after ack",
			zap.Error(err),
			zap.String("stream_id", message.ID),
		)

		return
	}

	dlqReplayTotal.Inc()
}

func (r *Replayer) insertWithTimeout(ctx context.Context, event platform.ModelMetricEvent, eventID string) error {
	insertCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	return r.exporter.InsertModelMetricEventWithDedupToken(insertCtx, event, eventID)
}

func (r *Replayer) poisonAndAck(ctx context.Context, message redis.XMessage, reason string) {
	values := map[string]interface{}{
		"source_stream_id": message.ID,
		"reason":           truncateError(reason, 1024),
		"poisoned_at":      time.Now().UTC().Format(time.RFC3339Nano),
	}
	for _, key := range []string{"event_id", "payload", "first_failed_at", "last_error"} {
		if value, ok := stringValue(message.Values[key]); ok {
			values[key] = value
		}
	}

	if err := r.client.XAdd(ctx, &redis.XAddArgs{
		Stream: r.cfg.PoisonStream,
		MaxLen: r.cfg.MaxLen,
		Approx: true,
		Values: values,
	}).Err(); err != nil {
		dlqReplayErrors.Inc()
		r.logger.Error("failed to move invalid model metrics event to poison stream", zap.Error(err))

		return
	}

	if err := r.client.XAck(ctx, r.cfg.Stream, r.cfg.Group, message.ID).Err(); err != nil {
		dlqReplayErrors.Inc()
		r.logger.Error("failed to ack poisoned model metrics event", zap.Error(err))

		return
	}

	if err := r.client.XDel(ctx, r.cfg.Stream, message.ID).Err(); err != nil {
		dlqReplayErrors.Inc()
		r.logger.Error("failed to remove poisoned model metrics event after ack", zap.Error(err))

		return
	}

	dlqPoisonTotal.Inc()
	r.logger.Warn("invalid model metrics event moved to poison stream",
		zap.String("stream_id", message.ID),
		zap.String("reason", reason),
	)
}

func (r *Replayer) updateMetrics(ctx context.Context) {
	if length, err := r.client.XLen(ctx, r.cfg.Stream).Result(); err == nil {
		dlqBacklog.Set(float64(length))
	}

	if pending, err := r.client.XPending(ctx, r.cfg.Stream, r.cfg.Group).Result(); err == nil {
		dlqPending.Set(float64(pending.Count))
	}
}

// Ready checks the active Sentinel-resolved Redis master.
func (r *Replayer) Ready(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *Replayer) Close() error {
	if r == nil || r.client == nil {
		return nil
	}

	return r.client.Close()
}

func stringValue(value interface{}) (string, bool) {
	switch typed := value.(type) {
	case string:
		return typed, true
	case []byte:
		return string(typed), true
	default:
		return "", false
	}
}
