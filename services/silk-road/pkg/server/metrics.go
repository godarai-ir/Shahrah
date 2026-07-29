package server

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var requestDurationBuckets = [...]float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}

var anthropicWIFExchangeSuccess atomic.Uint64
var anthropicWIFExchangeFailure atomic.Uint64
var anthropicWIFRefreshTotal atomic.Uint64
var anthropicStreamErrors atomic.Uint64

type metricKey struct {
	route  string
	method string
}

type metricValue struct {
	inFlight       uint64
	requests       map[int]uint64
	durationCount  uint64
	durationSum    float64
	durationBucket [len(requestDurationBuckets)]uint64
}

type httpMetrics struct {
	service string
	mu      sync.RWMutex
	values  map[metricKey]*metricValue
}

func newHTTPMetrics(service string) *httpMetrics {
	return &httpMetrics{service: service, values: make(map[metricKey]*metricValue)}
}

func (m *httpMetrics) handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		m.writePrometheus(w)
	})
}

func (m *httpMetrics) wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}

		key := metricKey{route: metricRoute(r.URL.Path), method: r.Method}
		m.start(key)
		start := time.Now()
		recorder := &metricResponseWriter{ResponseWriter: w}
		next.ServeHTTP(recorder, r)

		status := recorder.status
		if status == 0 {
			status = http.StatusOK
		}
		m.finish(key, status, time.Since(start).Seconds())
	})
}

func (m *httpMetrics) start(key metricKey) {
	m.mu.Lock()
	value := m.value(key)
	value.inFlight++
	m.mu.Unlock()
}

func (m *httpMetrics) finish(key metricKey, status int, duration float64) {
	m.mu.Lock()
	value := m.value(key)
	if value.inFlight > 0 {
		value.inFlight--
	}
	value.requests[status]++
	value.durationCount++
	value.durationSum += duration
	for i, upperBound := range requestDurationBuckets {
		if duration <= upperBound {
			value.durationBucket[i]++
		}
	}
	m.mu.Unlock()
}

func (m *httpMetrics) value(key metricKey) *metricValue {
	value := m.values[key]
	if value == nil {
		value = &metricValue{requests: make(map[int]uint64)}
		m.values[key] = value
	}
	return value
}

func (m *httpMetrics) writePrometheus(w http.ResponseWriter) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	service := strconv.Quote(m.service)
	_, _ = fmt.Fprintln(w, "# HELP shahrah_transport_build_info Static build information for a Shahrah transport service.")
	_, _ = fmt.Fprintln(w, "# TYPE shahrah_transport_build_info gauge")
	_, _ = fmt.Fprintf(w, "shahrah_transport_build_info{service=%s} 1\n", service)
	_, _ = fmt.Fprintln(w, "# HELP shahrah_silk_road_anthropic_wif_exchange_total Anthropic WIF token exchanges by bounded result.")
	_, _ = fmt.Fprintln(w, "# TYPE shahrah_silk_road_anthropic_wif_exchange_total counter")
	_, _ = fmt.Fprintf(w, "shahrah_silk_road_anthropic_wif_exchange_total{result=\"success\"} %d\n", anthropicWIFExchangeSuccess.Load())
	_, _ = fmt.Fprintf(w, "shahrah_silk_road_anthropic_wif_exchange_total{result=\"failure\"} %d\n", anthropicWIFExchangeFailure.Load())
	_, _ = fmt.Fprintln(w, "# HELP shahrah_silk_road_anthropic_wif_refresh_total Anthropic WIF refreshes performed before token expiry.")
	_, _ = fmt.Fprintln(w, "# TYPE shahrah_silk_road_anthropic_wif_refresh_total counter")
	_, _ = fmt.Fprintf(w, "shahrah_silk_road_anthropic_wif_refresh_total %d\n", anthropicWIFRefreshTotal.Load())
	_, _ = fmt.Fprintln(w, "# HELP shahrah_silk_road_anthropic_stream_errors_total Anthropic upstream stream error events.")
	_, _ = fmt.Fprintln(w, "# TYPE shahrah_silk_road_anthropic_stream_errors_total counter")
	_, _ = fmt.Fprintf(w, "shahrah_silk_road_anthropic_stream_errors_total %d\n", anthropicStreamErrors.Load())
	_, _ = fmt.Fprintln(w, "# HELP shahrah_transport_http_requests_total Total HTTP requests handled by Shahrah transport services.")
	_, _ = fmt.Fprintln(w, "# TYPE shahrah_transport_http_requests_total counter")

	keys := make([]metricKey, 0, len(m.values))
	for key := range m.values {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].route == keys[j].route {
			return keys[i].method < keys[j].method
		}
		return keys[i].route < keys[j].route
	})

	for _, key := range keys {
		value := m.values[key]
		statuses := make([]int, 0, len(value.requests))
		for status := range value.requests {
			statuses = append(statuses, status)
		}
		sort.Ints(statuses)
		for _, status := range statuses {
			_, _ = fmt.Fprintf(w, "shahrah_transport_http_requests_total{service=%s,route=%s,method=%s,status_code=%q} %d\n", service, strconv.Quote(key.route), strconv.Quote(key.method), strconv.Itoa(status), value.requests[status])
		}
	}

	_, _ = fmt.Fprintln(w, "# HELP shahrah_transport_http_request_duration_seconds HTTP request duration for Shahrah transport services.")
	_, _ = fmt.Fprintln(w, "# TYPE shahrah_transport_http_request_duration_seconds histogram")
	for _, key := range keys {
		value := m.values[key]
		labels := fmt.Sprintf("service=%s,route=%s,method=%s", service, strconv.Quote(key.route), strconv.Quote(key.method))
		for i, upperBound := range requestDurationBuckets {
			_, _ = fmt.Fprintf(w, "shahrah_transport_http_request_duration_seconds_bucket{%s,le=%q} %d\n", labels, strconv.FormatFloat(upperBound, 'g', -1, 64), value.durationBucket[i])
		}
		_, _ = fmt.Fprintf(w, "shahrah_transport_http_request_duration_seconds_bucket{%s,le=\"+Inf\"} %d\n", labels, value.durationCount)
		_, _ = fmt.Fprintf(w, "shahrah_transport_http_request_duration_seconds_sum{%s} %g\n", labels, value.durationSum)
		_, _ = fmt.Fprintf(w, "shahrah_transport_http_request_duration_seconds_count{%s} %d\n", labels, value.durationCount)
	}

	_, _ = fmt.Fprintln(w, "# HELP shahrah_transport_http_requests_in_flight Current in-flight HTTP requests in Shahrah transport services.")
	_, _ = fmt.Fprintln(w, "# TYPE shahrah_transport_http_requests_in_flight gauge")
	for _, key := range keys {
		value := m.values[key]
		_, _ = fmt.Fprintf(w, "shahrah_transport_http_requests_in_flight{service=%s,route=%s,method=%s} %d\n", service, strconv.Quote(key.route), strconv.Quote(key.method), value.inFlight)
	}
}

func metricRoute(path string) string {
	switch path {
	case "/dispatch":
		return "dispatch"
	case "/dispatch/raw":
		return "dispatch_raw"
	case "/health":
		return "health"
	case "/live":
		return "live"
	case "/ready":
		return "ready"
	default:
		return "other"
	}
}

type metricResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *metricResponseWriter) WriteHeader(status int) {
	if w.status != 0 {
		return
	}
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *metricResponseWriter) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(body)
}

func (w *metricResponseWriter) Flush() {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *metricResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	return hijacker.Hijack()
}

func (w *metricResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
