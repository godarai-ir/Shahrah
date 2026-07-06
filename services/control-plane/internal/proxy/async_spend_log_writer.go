package proxy

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"go.uber.org/zap"
)

var ErrSpendLogQueueFull = errors.New("spend log queue is full")

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
	DropOnFull     bool
}

type AsyncSpendLogWriter struct {
	underlying SpendLogger
	batch      SpendLogBatchWriter
	metrics    SpendLogMetrics
	cfg        AsyncSpendLogConfig
	logger     *zap.Logger
	queue      chan platform.SpendLog
	wg         sync.WaitGroup
	closed     atomic.Bool
}

func NewAsyncSpendLogWriter(underlying SpendLogger, cfg AsyncSpendLogConfig, logger *zap.Logger, metrics SpendLogMetrics) *AsyncSpendLogWriter {
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
	for i := 0; i < cfg.Workers; i++ {
		w.wg.Add(1)
		go w.worker()
	}
	return w
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
		return w.handleQueueFull()
	}

	if w.cfg.EnqueueTimeout > 0 {
		timer := time.NewTimer(w.cfg.EnqueueTimeout)
		defer timer.Stop()
		select {
		case w.queue <- log:
			w.recordQueued()
			return nil
		case <-timer.C:
			return w.handleQueueFull()
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	select {
	case w.queue <- log:
		w.recordQueued()
		return nil
	default:
		return w.handleQueueFull()
	}
}

func (w *AsyncSpendLogWriter) Stop(ctx context.Context) error {
	if w == nil || !w.closed.CompareAndSwap(false, true) {
		return nil
	}
	close(w.queue)

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

func (w *AsyncSpendLogWriter) handleQueueFull() error {
	w.recordDropped()
	w.logger.Warn("spend log queue full; dropping log")
	if w.cfg.DropOnFull {
		return nil
	}
	return ErrSpendLogQueueFull
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
