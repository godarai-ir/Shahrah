package tracing

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"

	tracev1 "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	resourcev1 "go.opentelemetry.io/proto/otlp/resource/v1"
	otlptracev1 "go.opentelemetry.io/proto/otlp/trace/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

type Exporter interface {
	Export(ctx context.Context, batch []SpanEvent) error
}

type DurableBatch struct {
	Events    []SpanEvent `json:"events"`
	CreatedAt time.Time   `json:"created_at"`
}

type OTLPExporter struct {
	cfg    Config
	client tracev1.TraceServiceClient
	conn   *grpc.ClientConn
	logger *zap.Logger
}

func NewOTLPExporter(cfg Config, logger *zap.Logger) (*OTLPExporter, error) {
	if cfg.OTLPEndpoint == "" {
		return nil, errors.New("TRACE_OTLP_ENDPOINT is required")
	}

	opts := []grpc.DialOption{}
	if cfg.OTLPInsecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.NewClient(cfg.OTLPEndpoint, opts...)
	if err != nil {
		return nil, err
	}

	return &OTLPExporter{
		cfg:    cfg,
		client: tracev1.NewTraceServiceClient(conn),
		conn:   conn,
		logger: logger,
	}, nil
}

func (e *OTLPExporter) Close() error {
	if e.conn != nil {
		return e.conn.Close()
	}

	return nil
}

func (e *OTLPExporter) Export(ctx context.Context, batch []SpanEvent) error {
	if len(batch) == 0 {
		return nil
	}

	req := toOTLPRequest(e.cfg.OTLPServiceName, batch)
	_, err := e.client.Export(ctx, req)

	return err
}

type StoreExporter struct {
	store Exporter
}

func NewStoreExporter(store Exporter) *StoreExporter {
	return &StoreExporter{store: store}
}

func (e *StoreExporter) Export(ctx context.Context, batch []SpanEvent) error {
	return e.store.Export(ctx, batch)
}

type WAL struct {
	path string
	mu   sync.Mutex
}

func NewWAL(path string) *WAL {
	return &WAL{path: path}
}

func (w *WAL) Append(batch DurableBatch) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(w.path), 0o755); err != nil {
		return err
	}

	f, err := os.OpenFile(w.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer closeFile(f)

	enc := json.NewEncoder(f)

	return enc.Encode(batch)
}

func (w *WAL) Drain() ([]DurableBatch, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	f, err := os.Open(w.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, err
	}
	defer closeFile(f)

	out := make([]DurableBatch, 0, 64)

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}

		var b DurableBatch
		if err := json.Unmarshal(line, &b); err != nil {
			continue
		}

		if len(b.Events) > 0 {
			out = append(out, b)
		}
	}

	if err := sc.Err(); err != nil {
		return nil, err
	}

	if err := os.Truncate(w.path, 0); err != nil {
		return nil, err
	}

	if _, err := os.OpenFile(w.path, os.O_WRONLY, 0o644); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return out, nil
}

func (w *WAL) BacklogSize() (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	f, err := os.Open(w.path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}

		return 0, err
	}
	defer closeFile(f)

	count := 0

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if len(sc.Bytes()) > 0 {
			count++
		}
	}

	if err := sc.Err(); err != nil {
		return 0, err
	}

	return count, nil
}

func closeFile(f *os.File) {
	if f == nil {
		return
	}

	if err := f.Close(); err != nil {
		return
	}
}

type ReliableExporter struct {
	cfg          Config
	primary      *OTLPExporter
	secondary    Exporter
	wal          *WAL
	logger       *zap.Logger
	stopReplayCh chan struct{}
	wg           sync.WaitGroup
}

func NewReliableExporter(cfg Config, secondary Exporter, logger *zap.Logger) (*ReliableExporter, error) {
	var (
		primary *OTLPExporter
		err     error
	)
	if cfg.OTLPEndpoint != "" {
		primary, err = NewOTLPExporter(cfg, logger)
		if err != nil {
			logger.Warn("otlp exporter init failed; using fallback only", zap.Error(err))
		}
	}

	var wal *WAL
	if cfg.WALEnabled {
		wal = NewWAL(cfg.WALPath)
	}

	re := &ReliableExporter{
		cfg:          cfg,
		primary:      primary,
		secondary:    secondary,
		wal:          wal,
		logger:       logger,
		stopReplayCh: make(chan struct{}),
	}
	re.startReplayLoop()

	return re, nil
}

func (e *ReliableExporter) Close() error {
	close(e.stopReplayCh)
	e.wg.Wait()

	if e.primary != nil {
		return e.primary.Close()
	}

	return nil
}

func (e *ReliableExporter) Export(ctx context.Context, batch []SpanEvent) error {
	if len(batch) == 0 {
		return nil
	}

	if err := e.exportPrimaryWithRetry(ctx, batch); err != nil {
		e.logger.Warn("otlp export failed after retries; writing to wal", zap.Error(err), zap.Int("batch", len(batch)))

		if e.wal != nil {
			_ = e.wal.Append(DurableBatch{Events: batch, CreatedAt: time.Now().UTC()})
		}
	}

	// Keep local queryability independent from OTLP health.
	if e.secondary != nil {
		if err := e.secondary.Export(ctx, batch); err != nil {
			e.logger.Warn("secondary trace export failed", zap.Error(err), zap.Int("batch", len(batch)))
			return err
		}
	}

	return nil
}

func (e *ReliableExporter) exportPrimaryWithRetry(ctx context.Context, batch []SpanEvent) error {
	if e.primary == nil {
		return nil
	}

	var lastErr error

	for attempt := 1; attempt <= e.cfg.RetryMaxAttempts; attempt++ {
		err := e.primary.Export(ctx, batch)
		if err == nil {
			return nil
		}

		lastErr = err

		sleep := backoffDelay(attempt, e.cfg.RetryBaseDelay, e.cfg.RetryMaxDelay)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(sleep):
		}
	}

	return lastErr
}

func (e *ReliableExporter) startReplayLoop() {
	if e.wal == nil || e.primary == nil {
		return
	}

	e.wg.Add(1)
	go func() {
		defer e.wg.Done()

		ticker := time.NewTicker(e.cfg.WALReplayInterval)
		defer ticker.Stop()

		for {
			select {
			case <-e.stopReplayCh:
				return
			case <-ticker.C:
				e.replayOnce()
			}
		}
	}()
}

func (e *ReliableExporter) replayOnce() {
	batches, err := e.wal.Drain()
	if err != nil || len(batches) == 0 {
		return
	}

	failed := make([]DurableBatch, 0, len(batches))
	for _, b := range batches {
		ctx, cancel := context.WithTimeout(context.Background(), e.cfg.ExportTimeout)
		err := e.exportPrimaryWithRetry(ctx, b.Events)

		cancel()

		if err != nil {
			failed = append(failed, b)
		}
	}

	for _, b := range failed {
		_ = e.wal.Append(b)
	}

	if len(failed) > 0 {
		e.logger.Warn("wal replay incomplete", zap.Int("failed_batches", len(failed)))
	}
}

func (e *ReliableExporter) ReadinessReport() map[string]interface{} {
	report := map[string]interface{}{
		"otlp_configured":    e.primary != nil,
		"otlp_reachable":     false,
		"wal_enabled":        e.wal != nil,
		"wal_backlog_size":   0,
		"telemetry_degraded": false,
	}

	if e.primary != nil && e.primary.conn != nil {
		st := e.primary.conn.GetState()
		reachable := st == connectivity.Ready

		report["otlp_reachable"] = reachable
		if !reachable {
			report["telemetry_degraded"] = true
			report["otlp_state"] = st.String()
		}
	}

	if e.wal != nil {
		if n, err := e.wal.BacklogSize(); err == nil {
			report["wal_backlog_size"] = n
			if n > 0 {
				report["telemetry_degraded"] = true
			}
		} else {
			report["telemetry_degraded"] = true
			report["wal_error"] = err.Error()
		}
	}

	return report
}

func backoffDelay(attempt int, base, max time.Duration) time.Duration {
	if attempt < 1 {
		attempt = 1
	}

	exp := float64(base) * math.Pow(2, float64(attempt-1))

	d := time.Duration(exp)
	if d > max {
		d = max
	}

	jitter := time.Duration(rand.Int63n(int64(d / 5)))

	return d + jitter
}

type Processor struct {
	cfg      Config
	queue    *Queue
	exporter Exporter
	logger   *zap.Logger
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

func NewProcessor(cfg Config, queue *Queue, exporter Exporter, logger *zap.Logger) *Processor {
	return &Processor{
		cfg:      cfg,
		queue:    queue,
		exporter: exporter,
		logger:   logger,
		stopCh:   make(chan struct{}),
	}
}

func (p *Processor) Start() {
	p.wg.Add(1)
	go p.loop()
}

func (p *Processor) Stop(ctx context.Context) error {
	close(p.stopCh)

	done := make(chan struct{})
	go func() {
		defer close(done)

		p.wg.Wait()
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *Processor) loop() {
	defer p.wg.Done()

	ticker := time.NewTicker(p.cfg.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopCh:
			p.flush()
			return
		case <-ticker.C:
			p.flush()
		}
	}
}

func (p *Processor) flush() {
	batch := p.queue.DequeueBatch(p.cfg.BatchSize)
	if len(batch) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), p.cfg.ExportTimeout)
	defer cancel()

	if err := p.exporter.Export(ctx, batch); err != nil {
		p.logger.Warn("trace export failed", zap.Error(err), zap.Int("batch", len(batch)))
		return
	}

	p.logger.Debug("trace batch exported", zap.Int("batch", len(batch)))
}

func (p *Processor) Health() string {
	m := p.queue.Metrics()
	return fmt.Sprintf("depth=%d dropped=%d enqueued=%d dequeued=%d", m["depth"], m["dropped"], m["enqueued"], m["dequeued"])
}

func toOTLPRequest(serviceName string, batch []SpanEvent) *tracev1.ExportTraceServiceRequest {
	scopeSpans := make([]*otlptracev1.ScopeSpans, 0, len(batch))
	for i, ev := range batch {
		spanID := fmt.Sprintf("%016x", i+1)

		statusCode := otlptracev1.Status_STATUS_CODE_OK
		if ev.StatusCode >= 500 {
			statusCode = otlptracev1.Status_STATUS_CODE_ERROR
		}

		span := &otlptracev1.Span{
			TraceId:           traceIDBytes(ev.TraceID),
			SpanId:            spanIDBytes(spanID),
			Name:              "gateway.request",
			Kind:              otlptracev1.Span_SPAN_KIND_SERVER,
			StartTimeUnixNano: uint64(ev.StartedAt.UnixNano()),
			EndTimeUnixNano:   uint64(ev.StartedAt.Add(time.Duration(ev.DurationMS) * time.Millisecond).UnixNano()),
			Attributes: []*commonv1.KeyValue{
				{Key: "tenant.id", Value: strVal(ev.TenantID)},
				{Key: "http.method", Value: strVal(ev.Method)},
				{Key: "http.route", Value: strVal(ev.Route)},
				{Key: "http.status_code", Value: intVal(int64(ev.StatusCode))},
				{Key: "shahrah.sampled", Value: boolVal(ev.Sampled)},
			},
			Status: &otlptracev1.Status{Code: statusCode},
		}
		scopeSpans = append(scopeSpans, &otlptracev1.ScopeSpans{
			Spans: []*otlptracev1.Span{span},
		})
	}

	return &tracev1.ExportTraceServiceRequest{
		ResourceSpans: []*otlptracev1.ResourceSpans{{
			Resource: &resourcev1.Resource{
				Attributes: []*commonv1.KeyValue{
					{Key: "service.name", Value: strVal(serviceName)},
				},
			},
			ScopeSpans: scopeSpans,
		}},
	}
}

func strVal(v string) *commonv1.AnyValue {
	return &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: v}}
}
func intVal(v int64) *commonv1.AnyValue {
	return &commonv1.AnyValue{Value: &commonv1.AnyValue_IntValue{IntValue: v}}
}
func boolVal(v bool) *commonv1.AnyValue {
	return &commonv1.AnyValue{Value: &commonv1.AnyValue_BoolValue{BoolValue: v}}
}

func traceIDBytes(traceID string) []byte {
	b := make([]byte, 16)
	copy(b, []byte(traceID))

	return b
}
func spanIDBytes(spanID string) []byte {
	b := make([]byte, 8)
	copy(b, []byte(spanID))

	return b
}
