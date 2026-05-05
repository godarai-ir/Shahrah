package tracing

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Exporter interface {
	Export(ctx context.Context, batch []SpanEvent) error
}

type LogExporter struct {
	logger *zap.Logger
}

func NewLogExporter(logger *zap.Logger) *LogExporter {
	return &LogExporter{logger: logger}
}

func (e *LogExporter) Export(_ context.Context, batch []SpanEvent) error {
	for _, ev := range batch {
		e.logger.Debug("trace span",
			zap.String("trace_id", ev.TraceID),
			zap.String("tenant_id", ev.TenantID),
			zap.String("method", ev.Method),
			zap.String("route", ev.Route),
			zap.Int("status_code", ev.StatusCode),
			zap.Int64("duration_ms", ev.DurationMS),
		)
	}
	return nil
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
