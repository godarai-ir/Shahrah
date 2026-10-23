package proxy

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"go.uber.org/zap"
)

// spendLogDirectWriteTimeout bounds the synchronous write used when the queue cannot take a log.
const spendLogDirectWriteTimeout = 5 * time.Second

type SpendLogBatchWriter interface {
	LogSpends(ctx context.Context, logs []platform.SpendLog) error
}

type SpendLogMetrics interface {
	IncSpendLogQueued()
	IncSpendLogDropped()
	ObserveSpendLogFlush(batchSize int, duration time.Duration, err error)
}

type AsyncSpendLogConfig struct {
	Enabled        bool
	QueueSize      int
	Workers        int
	BatchSize      int
	FlushInterval  time.Duration
	EnqueueTimeout time.Duration
}

type AsyncSpendLogWriter struct {
	underlying SpendLogger
	batch      SpendLogBatchWriter
	metrics    SpendLogMetrics
	cfg        AsyncSpendLogConfig
	logger     *zap.Logger
	queue      chan platform.SpendLog
	wg         sync.WaitGroup
	lifecycle  sync.Mutex
	started    bool
	closed     atomic.Bool
}

func NewAsyncSpendLogWriter(underlying SpendLogger, cfg AsyncSpendLogConfig, logger *zap.Logger, metrics SpendLogMetrics) *AsyncSpendLogWriter {
	w := NewManagedAsyncSpendLogWriter(underlying, cfg, logger, metrics)
	w.Start()

	return w
}

// NewManagedAsyncSpendLogWriter creates a writer whose workers are started explicitly.
func NewManagedAsyncSpendLogWriter(underlying SpendLogger, cfg AsyncSpendLogConfig, logger *zap.Logger, metrics SpendLogMetrics) *AsyncSpendLogWriter {
	cfg = normalizeAsyncSpendLogConfig(cfg)

	if logger == nil {
		logger = zap.NewNop()
	}

	w := &AsyncSpendLogWriter{
		underlying: underlying,
		cfg:        cfg,
		logger:     logger,
		metrics:    metrics,
		queue:      make(chan platform.SpendLog, cfg.QueueSize),
	}
	if batch, ok := underlying.(SpendLogBatchWriter); ok {
		w.batch = batch
	}

	return w
}

// Start starts the writer workers once.
func (w *AsyncSpendLogWriter) Start() {
	if w == nil {
		return
	}

	w.lifecycle.Lock()
	defer w.lifecycle.Unlock()

	if w.started || w.closed.Load() {
		return
	}

	w.started = true
	for i := 0; i < w.cfg.Workers; i++ {
		w.wg.Add(1)
		go w.worker()
	}
}

func normalizeAsyncSpendLogConfig(cfg AsyncSpendLogConfig) AsyncSpendLogConfig {
	if cfg.QueueSize <= 0 {
		cfg.QueueSize = 10000
	}

	if cfg.Workers <= 0 {
		cfg.Workers = 1
	}

	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 500
	}

	if cfg.FlushInterval <= 0 {
		cfg.FlushInterval = 500 * time.Millisecond
	}

	return cfg
}

func (w *AsyncSpendLogWriter) LogSpend(ctx context.Context, log platform.SpendLog) error {
	if w == nil || w.underlying == nil {
		return nil
	}

	if w.closed.Load() {
		return w.writeOverflow(log)
	}

	if w.cfg.EnqueueTimeout > 0 {
		timer := time.NewTimer(w.cfg.EnqueueTimeout)
		defer timer.Stop()

		select {
		case w.queue <- log:
			w.recordQueued()
			return nil
		case <-timer.C:
			return w.writeOverflow(log)
		case <-ctx.Done():
			return w.writeOverflow(log)
		}
	}

	select {
	case w.queue <- log:
		w.recordQueued()
		return nil
	default:
		return w.writeOverflow(log)
	}
}

func (w *AsyncSpendLogWriter) Stop(ctx context.Context) error {
	if w == nil {
		return nil
	}

	w.lifecycle.Lock()
	if !w.closed.CompareAndSwap(false, true) {
		w.lifecycle.Unlock()

		return nil
	}

	close(w.queue)
	w.lifecycle.Unlock()

	done := make(chan struct{})

	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// writeOverflow writes a log that the queue could not take, so usage is still billed.
// The write uses its own timeout, so a client disconnect cannot cancel it.
func (w *AsyncSpendLogWriter) writeOverflow(log platform.SpendLog) error {
	w.recordDropped()

	ctx, cancel := context.WithTimeout(context.Background(), spendLogDirectWriteTimeout)
	defer cancel()

	if err := w.underlying.LogSpend(ctx, log); err != nil {
		w.logger.Error("spend log queue full and direct write failed", zap.Error(err), zap.String("request_id", log.RequestID))

		return err
	}

	w.logger.Warn("spend log queue full; wrote log directly", zap.String("request_id", log.RequestID))

	return nil
}

func (w *AsyncSpendLogWriter) recordQueued() {
	if w.metrics != nil {
		w.metrics.IncSpendLogQueued()
	}
}

func (w *AsyncSpendLogWriter) recordDropped() {
	if w.metrics != nil {
		w.metrics.IncSpendLogDropped()
	}
}

func (w *AsyncSpendLogWriter) worker() {
	defer w.wg.Done()

	ticker := time.NewTicker(w.cfg.FlushInterval)
	defer ticker.Stop()

	batch := make([]platform.SpendLog, 0, w.cfg.BatchSize)
	flush := func() {
		if len(batch) == 0 {
			return
		}

		w.flushBatch(context.Background(), batch)
		batch = make([]platform.SpendLog, 0, w.cfg.BatchSize)
	}

	for {
		select {
		case log, ok := <-w.queue:
			if !ok {
				flush()
				return
			}

			batch = append(batch, log)
			if len(batch) >= w.cfg.BatchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}

func (w *AsyncSpendLogWriter) flushBatch(ctx context.Context, logs []platform.SpendLog) {
	if len(logs) == 0 || w.underlying == nil {
		return
	}

	start := time.Now()

	var err error
	if w.batch != nil {
		err = w.batch.LogSpends(ctx, logs)
	} else {
		for _, log := range logs {
			if writeErr := w.underlying.LogSpend(ctx, log); writeErr != nil {
				err = writeErr
				break
			}
		}
	}

	if w.metrics != nil {
		w.metrics.ObserveSpendLogFlush(len(logs), time.Since(start), err)
	}

	if err != nil {
		w.logger.Warn("failed to flush spend log batch", zap.Error(err), zap.Int("batch_size", len(logs)))
	}
}
