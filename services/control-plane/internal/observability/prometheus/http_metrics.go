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

	registry.MustRegister(requestsTotal, requestDuration, inflight, backpressureInFlight, backpressureRejectedTotal)

	return &HTTPMetrics{
		registry:                  registry,
		requestsTotal:             requestsTotal,
		requestDuration:           requestDuration,
		inflight:                  inflight,
		backpressureInFlight:      backpressureInFlight,
		backpressureRejectedTotal: backpressureRejectedTotal,
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
