// Package prometheus exposes HTTP metrics middleware and handlers.
package prometheus

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// HTTPMetrics records request counts, latency, and in-flight totals.
type HTTPMetrics struct {
	registry        *prom.Registry
	requestsTotal   *prom.CounterVec
	requestDuration *prom.HistogramVec
	inflight        prom.Gauge
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

	registry.MustRegister(requestsTotal, requestDuration, inflight)

	return &HTTPMetrics{
		registry:        registry,
		requestsTotal:   requestsTotal,
		requestDuration: requestDuration,
		inflight:        inflight,
	}
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
