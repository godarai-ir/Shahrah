package server

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math"
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
var anthropicRetryTotal atomic.Uint64
var anthropicRetryExhausted atomic.Uint64

type metricKey struct {
	route  string
	method string
}

type metricValue struct {
	inFlight       atomic.Uint64
	requests       sync.Map
	durationCount  atomic.Uint64
	durationSum    atomic.Uint64
	durationBucket [len(requestDurationBuckets)]atomic.Uint64
}

type httpMetrics struct {
	service string
	values  sync.Map
}

func newHTTPMetrics(service string) *httpMetrics {
	return &httpMetrics{service: service}
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
	m.value(key).inFlight.Add(1)
}

func (m *httpMetrics) finish(key metricKey, status int, duration float64) {
	value := m.value(key)
	value.inFlight.Add(^uint64(0))
	value.requestCounter(status).Add(1)
	value.durationCount.Add(1)
	addAtomicFloat64(&value.durationSum, duration)
	for i, upperBound := range requestDurationBuckets {
		if duration <= upperBound {
			value.durationBucket[i].Add(1)
		}
	}
}

func (m *httpMetrics) value(key metricKey) *metricValue {
	if value, ok := m.values.Load(key); ok {
		return checkedMetricValue(value)
	}

	value, _ := m.values.LoadOrStore(key, &metricValue{})

	return checkedMetricValue(value)
}

func (v *metricValue) requestCounter(status int) *atomic.Uint64 {
	if counter, ok := v.requests.Load(status); ok {
		return checkedMetricCounter(counter)
	}

	counter, _ := v.requests.LoadOrStore(status, &atomic.Uint64{})

	return checkedMetricCounter(counter)
}

func checkedMetricValue(value any) *metricValue {
	typed, ok := value.(*metricValue)
	if !ok {
		panic("invalid metric value type")
	}

	return typed
}

func checkedMetricCounter(value any) *atomic.Uint64 {
	typed, ok := value.(*atomic.Uint64)
	if !ok {
		panic("invalid metric counter type")
	}

	return typed
}

func addAtomicFloat64(target *atomic.Uint64, delta float64) {
	for {
		current := target.Load()

		updated := math.Float64bits(math.Float64frombits(current) + delta)
		if target.CompareAndSwap(current, updated) {
			return
		}
	}
}

func (m *httpMetrics) writePrometheus(w http.ResponseWriter) {
	var output bytes.Buffer

	service := strconv.Quote(m.service)

	writeMetricText(&output, "# HELP shahrah_transport_build_info Static build information for a Shahrah transport service.")
	writeMetricText(&output, "# TYPE shahrah_transport_build_info gauge")
	writeMetricFormat(&output, "shahrah_transport_build_info{service=%s} 1\n", service)
	writeMetricText(&output, "# HELP shahrah_silk_road_anthropic_wif_exchange_total Anthropic WIF token exchanges by bounded result.")
	writeMetricText(&output, "# TYPE shahrah_silk_road_anthropic_wif_exchange_total counter")
	writeMetricFormat(&output, "shahrah_silk_road_anthropic_wif_exchange_total{result=\"success\"} %d\n", anthropicWIFExchangeSuccess.Load())
	writeMetricFormat(&output, "shahrah_silk_road_anthropic_wif_exchange_total{result=\"failure\"} %d\n", anthropicWIFExchangeFailure.Load())
	writeMetricText(&output, "# HELP shahrah_silk_road_anthropic_wif_refresh_total Anthropic WIF refreshes performed before token expiry.")
	writeMetricText(&output, "# TYPE shahrah_silk_road_anthropic_wif_refresh_total counter")
	writeMetricFormat(&output, "shahrah_silk_road_anthropic_wif_refresh_total %d\n", anthropicWIFRefreshTotal.Load())
	writeMetricFormat(&output, "shahrah_silk_road_anthropic_retry_total %d\n", anthropicRetryTotal.Load())
	writeMetricFormat(&output, "shahrah_silk_road_anthropic_retry_exhausted_total %d\n", anthropicRetryExhausted.Load())
	writeMetricText(&output, "# HELP shahrah_silk_road_anthropic_stream_errors_total Anthropic upstream stream error events.")
	writeMetricText(&output, "# TYPE shahrah_silk_road_anthropic_stream_errors_total counter")
	writeMetricFormat(&output, "shahrah_silk_road_anthropic_stream_errors_total %d\n", anthropicStreamErrors.Load())
	writeMetricText(&output, "# HELP shahrah_transport_http_requests_total Total HTTP requests handled by Shahrah transport services.")
	writeMetricText(&output, "# TYPE shahrah_transport_http_requests_total counter")

	keys := make([]metricKey, 0, 8)

	m.values.Range(func(key, _ any) bool {
		typed, ok := key.(metricKey)
		if ok {
			keys = append(keys, typed)
		}

		return true
	})
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].route == keys[j].route {
			return keys[i].method < keys[j].method
		}
		return keys[i].route < keys[j].route
	})

	for _, key := range keys {
		value := m.value(key)
		statuses := make([]int, 0, 8)

		value.requests.Range(func(status, _ any) bool {
			typed, ok := status.(int)
			if ok {
				statuses = append(statuses, typed)
			}

			return true
		})
		sort.Ints(statuses)
		for _, status := range statuses {
			counter, _ := value.requests.Load(status)
			writeMetricFormat(&output, "shahrah_transport_http_requests_total{service=%s,route=%s,method=%s,status_code=%q} %d\n", service, strconv.Quote(key.route), strconv.Quote(key.method), strconv.Itoa(status), checkedMetricCounter(counter).Load())
		}
	}

	writeMetricText(&output, "# HELP shahrah_transport_http_request_duration_seconds HTTP request duration for Shahrah transport services.")
	writeMetricText(&output, "# TYPE shahrah_transport_http_request_duration_seconds histogram")
	for _, key := range keys {
		value := m.value(key)
		labels := fmt.Sprintf("service=%s,route=%s,method=%s", service, strconv.Quote(key.route), strconv.Quote(key.method))
		for i, upperBound := range requestDurationBuckets {
			writeMetricFormat(&output, "shahrah_transport_http_request_duration_seconds_bucket{%s,le=%q} %d\n", labels, strconv.FormatFloat(upperBound, 'g', -1, 64), value.durationBucket[i].Load())
		}

		writeMetricFormat(&output, "shahrah_transport_http_request_duration_seconds_bucket{%s,le=\"+Inf\"} %d\n", labels, value.durationCount.Load())
		writeMetricFormat(&output, "shahrah_transport_http_request_duration_seconds_sum{%s} %g\n", labels, math.Float64frombits(value.durationSum.Load()))
		writeMetricFormat(&output, "shahrah_transport_http_request_duration_seconds_count{%s} %d\n", labels, value.durationCount.Load())
	}

	writeMetricText(&output, "# HELP shahrah_transport_http_requests_in_flight Current in-flight HTTP requests in Shahrah transport services.")
	writeMetricText(&output, "# TYPE shahrah_transport_http_requests_in_flight gauge")
	for _, key := range keys {
		value := m.value(key)
		writeMetricFormat(&output, "shahrah_transport_http_requests_in_flight{service=%s,route=%s,method=%s} %d\n", service, strconv.Quote(key.route), strconv.Quote(key.method), value.inFlight.Load())
	}

	if _, err := w.Write(output.Bytes()); err != nil {
		return
	}
}

func writeMetricText(output *bytes.Buffer, text string) {
	if _, err := fmt.Fprintln(output, text); err != nil {
		panic("write metric text: " + err.Error())
	}
}

func writeMetricFormat(output *bytes.Buffer, format string, values ...any) {
	if _, err := fmt.Fprintf(output, format, values...); err != nil {
		panic("format metric text: " + err.Error())
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

func (w *metricResponseWriter) ReadFrom(reader io.Reader) (int64, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}

	if readerFrom, ok := w.ResponseWriter.(io.ReaderFrom); ok {
		return readerFrom.ReadFrom(reader)
	}

	return io.Copy(w.ResponseWriter, reader)
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
