// Package prometheus exposes HTTP metrics middleware and handlers.
package prometheus

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/jackc/pgx/v5/pgxpool"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PoolStatsProvider exposes PostgreSQL pool stats without leaking the pool.
type PoolStatsProvider interface {
	PoolStats() *pgxpool.Stat
}

// PgxPoolStats is the subset of pgxpool.Stat used by metrics.
type PgxPoolStats interface {
	AcquireCount() int64
	AcquireDuration() time.Duration
	AcquiredConns() int32
	CanceledAcquireCount() int64
	ConstructingConns() int32
	EmptyAcquireCount() int64
	IdleConns() int32
	MaxConns() int32
	TotalConns() int32
}

// HTTPMetrics records request counts, latency, and in-flight totals.
type HTTPMetrics struct {
	registry                  *prom.Registry
	requestsTotal             *prom.CounterVec
	requestDuration           *prom.HistogramVec
	inflight                  prom.Gauge
	backpressureInFlight      prom.Gauge
	backpressureRejectedTotal prom.Counter
	tokenSpendQueuedTotal     prom.Counter
	tokenSpendFallbackTotal   prom.Counter
	tokenSpendFlushDuration   prom.Histogram
	tokenSpendFlushBatchSize  prom.Histogram
	tokenSpendFlushErrors     prom.Counter
	spendLogQueuedTotal       prom.Counter
	spendLogDroppedTotal      prom.Counter
	spendLogFlushDuration     prom.Histogram
	spendLogFlushBatchSize    prom.Histogram
	spendLogFlushErrors       prom.Counter
	keyLastUsedQueuedTotal    prom.Counter
	keyLastUsedSkippedTotal   prom.Counter
	keyLastUsedFallbackTotal  prom.Counter
	keyLastUsedFlushDuration  prom.Histogram
	keyLastUsedFlushBatchSize prom.Histogram
	keyLastUsedFlushErrors    prom.Counter
}

// NewHTTPMetrics creates a Prometheus registry and HTTP metrics collectors.
func NewHTTPMetrics() *HTTPMetrics {
	registry := prom.NewRegistry()
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	requestsTotal := prom.NewCounterVec(prom.CounterOpts{
		Namespace: "shahrah",
		Subsystem: "control_plane",
		Name:      "http_requests_total",
		Help:      "Total number of HTTP requests handled by the Shahrah control-plane.",
	}, []string{"method", "route", "status_code"})

	requestDuration := prom.NewHistogramVec(prom.HistogramOpts{
		Namespace: "shahrah",
		Subsystem: "control_plane",
		Name:      "http_request_duration_seconds",
		Help:      "HTTP request duration for the Shahrah control-plane.",
		Buckets:   prom.DefBuckets,
	}, []string{"method", "route", "status_code"})

	inflight := prom.NewGauge(prom.GaugeOpts{
		Namespace: "shahrah",
		Subsystem: "control_plane",
		Name:      "http_requests_in_flight",
		Help:      "Current number of in-flight HTTP requests handled by the Shahrah control-plane.",
	})

	backpressureInFlight := prom.NewGauge(prom.GaugeOpts{
		Namespace: "shahrah",
		Subsystem: "control_plane",
		Name:      "gateway_backpressure_in_flight",
		Help:      "Current number of non-exempt requests admitted by gateway backpressure.",
	})

	backpressureRejectedTotal := prom.NewCounter(prom.CounterOpts{
		Namespace: "shahrah",
		Subsystem: "control_plane",
		Name:      "gateway_backpressure_rejections_total",
		Help:      "Total number of requests rejected by gateway backpressure.",
	})

	tokenSpendQueuedTotal := prom.NewCounter(prom.CounterOpts{
		Namespace: "shahrah", Subsystem: "control_plane", Name: "token_spend_delta_queued_total",
		Help: "Total token spend deltas queued for Redis aggregation.",
	})
	tokenSpendFallbackTotal := prom.NewCounter(prom.CounterOpts{
		Namespace: "shahrah", Subsystem: "control_plane", Name: "token_spend_redis_fallback_total",
		Help: "Total token spend aggregation attempts that fell back because Redis was unavailable.",
	})
	tokenSpendFlushDuration := prom.NewHistogram(prom.HistogramOpts{
		Namespace: "shahrah", Subsystem: "control_plane", Name: "token_spend_flush_duration_seconds",
		Help:    "Duration of token spend aggregate flushes.",
		Buckets: prom.DefBuckets,
	})
	tokenSpendFlushBatchSize := prom.NewHistogram(prom.HistogramOpts{
		Namespace: "shahrah", Subsystem: "control_plane", Name: "token_spend_flush_batch_size",
		Help:    "Number of token spend deltas persisted in each flush.",
		Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000},
	})
	tokenSpendFlushErrors := prom.NewCounter(prom.CounterOpts{
		Namespace: "shahrah", Subsystem: "control_plane", Name: "token_spend_flush_errors_total",
		Help: "Total token spend aggregate flush failures.",
	})
	spendLogQueuedTotal := prom.NewCounter(prom.CounterOpts{
		Namespace: "shahrah", Subsystem: "control_plane", Name: "spend_log_queued_total",
		Help: "Total spend logs queued for asynchronous persistence.",
	})
	spendLogDroppedTotal := prom.NewCounter(prom.CounterOpts{
		Namespace: "shahrah", Subsystem: "control_plane", Name: "spend_log_dropped_total",
		Help: "Total spend logs dropped before persistence because the async queue was full or closed.",
	})
	spendLogFlushDuration := prom.NewHistogram(prom.HistogramOpts{
		Namespace: "shahrah", Subsystem: "control_plane", Name: "spend_log_flush_duration_seconds",
		Help:    "Duration of asynchronous spend log flushes.",
		Buckets: prom.DefBuckets,
	})
	spendLogFlushBatchSize := prom.NewHistogram(prom.HistogramOpts{
		Namespace: "shahrah", Subsystem: "control_plane", Name: "spend_log_flush_batch_size",
		Help:    "Number of spend logs persisted in each asynchronous flush.",
		Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000},
	})
	spendLogFlushErrors := prom.NewCounter(prom.CounterOpts{
		Namespace: "shahrah", Subsystem: "control_plane", Name: "spend_log_flush_errors_total",
		Help: "Total asynchronous spend log flush failures.",
	})
	keyLastUsedQueuedTotal := prom.NewCounter(prom.CounterOpts{
		Namespace: "shahrah", Subsystem: "control_plane", Name: "key_last_used_queued_total",
		Help: "Total key last-used timestamps queued for Redis aggregation.",
	})
	keyLastUsedSkippedTotal := prom.NewCounter(prom.CounterOpts{
		Namespace: "shahrah", Subsystem: "control_plane", Name: "key_last_used_skipped_total",
		Help: "Total key last-used updates skipped because a newer/equivalent timestamp was already queued or a fallback throttle applied.",
	})
	keyLastUsedFallbackTotal := prom.NewCounter(prom.CounterOpts{
		Namespace: "shahrah", Subsystem: "control_plane", Name: "key_last_used_redis_fallback_total",
		Help: "Total key last-used aggregation attempts that fell back because Redis was unavailable.",
	})
	keyLastUsedFlushDuration := prom.NewHistogram(prom.HistogramOpts{
		Namespace: "shahrah", Subsystem: "control_plane", Name: "key_last_used_flush_duration_seconds",
		Help:    "Duration of key last-used aggregate flushes.",
		Buckets: prom.DefBuckets,
	})
	keyLastUsedFlushBatchSize := prom.NewHistogram(prom.HistogramOpts{
		Namespace: "shahrah", Subsystem: "control_plane", Name: "key_last_used_flush_batch_size",
		Help:    "Number of key last-used timestamps persisted in each flush.",
		Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000},
	})
	keyLastUsedFlushErrors := prom.NewCounter(prom.CounterOpts{
		Namespace: "shahrah", Subsystem: "control_plane", Name: "key_last_used_flush_errors_total",
		Help: "Total key last-used aggregate flush failures.",
	})

	registry.MustRegister(
		requestsTotal, requestDuration, inflight, backpressureInFlight, backpressureRejectedTotal,
		tokenSpendQueuedTotal, tokenSpendFallbackTotal, tokenSpendFlushDuration, tokenSpendFlushBatchSize, tokenSpendFlushErrors,
		spendLogQueuedTotal, spendLogDroppedTotal, spendLogFlushDuration, spendLogFlushBatchSize, spendLogFlushErrors,
		keyLastUsedQueuedTotal, keyLastUsedSkippedTotal, keyLastUsedFallbackTotal, keyLastUsedFlushDuration, keyLastUsedFlushBatchSize, keyLastUsedFlushErrors,
	)

	return &HTTPMetrics{
		registry:                  registry,
		requestsTotal:             requestsTotal,
		requestDuration:           requestDuration,
		inflight:                  inflight,
		backpressureInFlight:      backpressureInFlight,
		backpressureRejectedTotal: backpressureRejectedTotal,
		tokenSpendQueuedTotal:     tokenSpendQueuedTotal,
		tokenSpendFallbackTotal:   tokenSpendFallbackTotal,
		tokenSpendFlushDuration:   tokenSpendFlushDuration,
		tokenSpendFlushBatchSize:  tokenSpendFlushBatchSize,
		tokenSpendFlushErrors:     tokenSpendFlushErrors,
		spendLogQueuedTotal:       spendLogQueuedTotal,
		spendLogDroppedTotal:      spendLogDroppedTotal,
		spendLogFlushDuration:     spendLogFlushDuration,
		spendLogFlushBatchSize:    spendLogFlushBatchSize,
		spendLogFlushErrors:       spendLogFlushErrors,
		keyLastUsedQueuedTotal:    keyLastUsedQueuedTotal,
		keyLastUsedSkippedTotal:   keyLastUsedSkippedTotal,
		keyLastUsedFallbackTotal:  keyLastUsedFallbackTotal,
		keyLastUsedFlushDuration:  keyLastUsedFlushDuration,
		keyLastUsedFlushBatchSize: keyLastUsedFlushBatchSize,
		keyLastUsedFlushErrors:    keyLastUsedFlushErrors,
	}
}

// IncBackpressureInFlight records one request admitted by the gateway backpressure gate.
func (m *HTTPMetrics) IncBackpressureInFlight() {
	if m == nil || m.backpressureInFlight == nil {
		return
	}
	m.backpressureInFlight.Inc()
}

// DecBackpressureInFlight records one request leaving the gateway backpressure gate.
func (m *HTTPMetrics) DecBackpressureInFlight() {
	if m == nil || m.backpressureInFlight == nil {
		return
	}
	m.backpressureInFlight.Dec()
}

// IncBackpressureRejected records one request rejected by the gateway backpressure gate.
func (m *HTTPMetrics) IncBackpressureRejected() {
	if m == nil || m.backpressureRejectedTotal == nil {
		return
	}
	m.backpressureRejectedTotal.Inc()
}

// IncTokenSpendQueued records a token spend delta queued in Redis.
func (m *HTTPMetrics) IncTokenSpendQueued() {
	if m != nil && m.tokenSpendQueuedTotal != nil {
		m.tokenSpendQueuedTotal.Inc()
	}
}

// IncTokenSpendRedisFallback records Redis unavailability on the token spend path.
func (m *HTTPMetrics) IncTokenSpendRedisFallback() {
	if m != nil && m.tokenSpendFallbackTotal != nil {
		m.tokenSpendFallbackTotal.Inc()
	}
}

// ObserveTokenSpendFlush records one token spend flush attempt.
func (m *HTTPMetrics) ObserveTokenSpendFlush(batchSize int, duration time.Duration, err error) {
	if m == nil {
		return
	}
	if m.tokenSpendFlushDuration != nil {
		m.tokenSpendFlushDuration.Observe(duration.Seconds())
	}
	if m.tokenSpendFlushBatchSize != nil {
		m.tokenSpendFlushBatchSize.Observe(float64(batchSize))
	}
	if err != nil && m.tokenSpendFlushErrors != nil {
		m.tokenSpendFlushErrors.Inc()
	}
}

// IncSpendLogQueued records one spend log accepted by the async persistence queue.
func (m *HTTPMetrics) IncSpendLogQueued() {
	if m != nil && m.spendLogQueuedTotal != nil {
		m.spendLogQueuedTotal.Inc()
	}
}

// IncSpendLogDropped records one spend log dropped before persistence.
func (m *HTTPMetrics) IncSpendLogDropped() {
	if m != nil && m.spendLogDroppedTotal != nil {
		m.spendLogDroppedTotal.Inc()
	}
}

// ObserveSpendLogFlush records one async spend log flush attempt.
func (m *HTTPMetrics) ObserveSpendLogFlush(batchSize int, duration time.Duration, err error) {
	if m == nil {
		return
	}
	if m.spendLogFlushDuration != nil {
		m.spendLogFlushDuration.Observe(duration.Seconds())
	}
	if m.spendLogFlushBatchSize != nil {
		m.spendLogFlushBatchSize.Observe(float64(batchSize))
	}
	if err != nil && m.spendLogFlushErrors != nil {
		m.spendLogFlushErrors.Inc()
	}
}

// IncKeyLastUsedQueued records a key last-used timestamp queued in Redis.
func (m *HTTPMetrics) IncKeyLastUsedQueued() {
	if m != nil && m.keyLastUsedQueuedTotal != nil {
		m.keyLastUsedQueuedTotal.Inc()
	}
}

// IncKeyLastUsedSkipped records a skipped or throttled key last-used update.
func (m *HTTPMetrics) IncKeyLastUsedSkipped() {
	if m != nil && m.keyLastUsedSkippedTotal != nil {
		m.keyLastUsedSkippedTotal.Inc()
	}
}

// IncKeyLastUsedRedisFallback records Redis unavailability on the key last-used path.
func (m *HTTPMetrics) IncKeyLastUsedRedisFallback() {
	if m != nil && m.keyLastUsedFallbackTotal != nil {
		m.keyLastUsedFallbackTotal.Inc()
	}
}

// ObserveKeyLastUsedFlush records one key last-used flush attempt.
func (m *HTTPMetrics) ObserveKeyLastUsedFlush(batchSize int, duration time.Duration, err error) {
	if m == nil {
		return
	}
	if m.keyLastUsedFlushDuration != nil {
		m.keyLastUsedFlushDuration.Observe(duration.Seconds())
	}
	if m.keyLastUsedFlushBatchSize != nil {
		m.keyLastUsedFlushBatchSize.Observe(float64(batchSize))
	}
	if err != nil && m.keyLastUsedFlushErrors != nil {
		m.keyLastUsedFlushErrors.Inc()
	}
}

// RegisterPgxPool registers PostgreSQL pool gauges on the existing registry.
func (m *HTTPMetrics) RegisterPgxPool(provider PoolStatsProvider) {
	if m == nil || m.registry == nil || provider == nil {
		return
	}

	collectors := []prom.Collector{
		prom.NewGaugeFunc(prom.GaugeOpts{
			Namespace: "shahrah", Subsystem: "control_plane", Name: "postgres_pool_acquired_conns",
			Help: "Current PostgreSQL connections acquired from the pool.",
		}, func() float64 { return int32Stat(provider, func(s PgxPoolStats) int32 { return s.AcquiredConns() }) }),
		prom.NewGaugeFunc(prom.GaugeOpts{
			Namespace: "shahrah", Subsystem: "control_plane", Name: "postgres_pool_idle_conns",
			Help: "Current idle PostgreSQL connections in the pool.",
		}, func() float64 { return int32Stat(provider, func(s PgxPoolStats) int32 { return s.IdleConns() }) }),
		prom.NewGaugeFunc(prom.GaugeOpts{
			Namespace: "shahrah", Subsystem: "control_plane", Name: "postgres_pool_total_conns",
			Help: "Current total PostgreSQL connections in the pool.",
		}, func() float64 { return int32Stat(provider, func(s PgxPoolStats) int32 { return s.TotalConns() }) }),
		prom.NewGaugeFunc(prom.GaugeOpts{
			Namespace: "shahrah", Subsystem: "control_plane", Name: "postgres_pool_constructing_conns",
			Help: "Current PostgreSQL connections being constructed by the pool.",
		}, func() float64 {
			return int32Stat(provider, func(s PgxPoolStats) int32 { return s.ConstructingConns() })
		}),
		prom.NewGaugeFunc(prom.GaugeOpts{
			Namespace: "shahrah", Subsystem: "control_plane", Name: "postgres_pool_max_conns",
			Help: "Maximum PostgreSQL connections configured for the pool.",
		}, func() float64 { return int32Stat(provider, func(s PgxPoolStats) int32 { return s.MaxConns() }) }),
		prom.NewGaugeFunc(prom.GaugeOpts{
			Namespace: "shahrah", Subsystem: "control_plane", Name: "postgres_pool_acquire_count",
			Help: "Cumulative successful PostgreSQL pool acquire operations.",
		}, func() float64 { return int64Stat(provider, func(s PgxPoolStats) int64 { return s.AcquireCount() }) }),
		prom.NewGaugeFunc(prom.GaugeOpts{
			Namespace: "shahrah", Subsystem: "control_plane", Name: "postgres_pool_empty_acquire_count",
			Help: "Cumulative PostgreSQL pool acquires that found no idle connection.",
		}, func() float64 {
			return int64Stat(provider, func(s PgxPoolStats) int64 { return s.EmptyAcquireCount() })
		}),
		prom.NewGaugeFunc(prom.GaugeOpts{
			Namespace: "shahrah", Subsystem: "control_plane", Name: "postgres_pool_canceled_acquire_count",
			Help: "Cumulative canceled PostgreSQL pool acquire operations.",
		}, func() float64 {
			return int64Stat(provider, func(s PgxPoolStats) int64 { return s.CanceledAcquireCount() })
		}),
		prom.NewGaugeFunc(prom.GaugeOpts{
			Namespace: "shahrah", Subsystem: "control_plane", Name: "postgres_pool_acquire_duration_seconds",
			Help: "Cumulative time spent acquiring PostgreSQL pool connections.",
		}, func() float64 {
			stats := provider.PoolStats()
			if stats == nil {
				return 0
			}
			return stats.AcquireDuration().Seconds()
		}),
	}

	m.registry.MustRegister(collectors...)
}

// Middleware returns a Fiber middleware that records request metrics.
func (m *HTTPMetrics) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		m.inflight.Inc()
		defer m.inflight.Dec()

		err := c.Next()

		route := routeLabel(c)
		statusCode := strconv.Itoa(c.Response().StatusCode())
		labels := prom.Labels{
			"method":      strings.Clone(c.Method()),
			"route":       strings.Clone(route),
			"status_code": statusCode,
		}
		m.requestsTotal.With(labels).Inc()
		m.requestDuration.With(labels).Observe(time.Since(start).Seconds())

		return err
	}
}

// Handler returns the Prometheus exposition endpoint.
func (m *HTTPMetrics) Handler() fiber.Handler {
	return adaptor.HTTPHandler(promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{}))
}

func routeLabel(c *fiber.Ctx) string {
	if route := c.Route(); route != nil {
		if path := route.Path; path != "" {
			return path
		}
	}

	if path := c.Path(); path != "" {
		return path
	}

	return "unmatched"
}

func int32Stat(provider PoolStatsProvider, read func(PgxPoolStats) int32) float64 {
	stats := provider.PoolStats()
	if stats == nil {
		return 0
	}
	return float64(read(stats))
}

func int64Stat(provider PoolStatsProvider, read func(PgxPoolStats) int64) float64 {
	stats := provider.PoolStats()
	if stats == nil {
		return 0
	}
	return float64(read(stats))
}
