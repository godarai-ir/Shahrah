package proxy

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"go.uber.org/zap"
)

var ErrRequestLogQueueFull = errors.New("request log queue is full")

type RequestLogBatchWriter interface {
	LogRequests(ctx context.Context, entries []platform.RequestLogEntry) error
}

type AsyncRequestLogConfig struct {
	Enabled        bool
	QueueSize      int
	Workers        int
	BatchSize      int
	FlushInterval  time.Duration
	EnqueueTimeout time.Duration
	DropOnFull     bool
}

type AsyncRequestLogWriter struct {
	underlying RequestLogger
	batch      RequestLogBatchWriter
	cfg        AsyncRequestLogConfig
	logger     *zap.Logger
	queue      chan platform.RequestLogEntry
	wg         sync.WaitGroup
	lifecycle  sync.Mutex
	started    bool
	closed     atomic.Bool
}

func NewAsyncRequestLogWriter(underlying RequestLogger, cfg AsyncRequestLogConfig, logger *zap.Logger) *AsyncRequestLogWriter {
	w := NewManagedAsyncRequestLogWriter(underlying, cfg, logger)
	w.Start()

	return w
}

// NewManagedAsyncRequestLogWriter creates a writer whose workers are started explicitly.
func NewManagedAsyncRequestLogWriter(underlying RequestLogger, cfg AsyncRequestLogConfig, logger *zap.Logger) *AsyncRequestLogWriter {
	cfg = normalizeAsyncRequestLogConfig(cfg)

	if logger == nil {
		logger = zap.NewNop()
	}

	w := &AsyncRequestLogWriter{
		underlying: underlying,
		cfg:        cfg,
		logger:     logger,
		queue:      make(chan platform.RequestLogEntry, cfg.QueueSize),
	}
	if batch, ok := underlying.(RequestLogBatchWriter); ok {
		w.batch = batch
	}

	return w
}

// Start starts the writer workers once.
func (w *AsyncRequestLogWriter) Start() {
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

func normalizeAsyncRequestLogConfig(cfg AsyncRequestLogConfig) AsyncRequestLogConfig {
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

func (w *AsyncRequestLogWriter) LogRequest(ctx context.Context, entry platform.RequestLogEntry) error {
	if w == nil || w.underlying == nil {
		return nil
	}

	if w.closed.Load() {
		if w.cfg.DropOnFull {
			return nil
		}

		return ErrRequestLogQueueFull
	}

	if w.cfg.EnqueueTimeout > 0 {
		timer := time.NewTimer(w.cfg.EnqueueTimeout)
		defer timer.Stop()

		select {
		case w.queue <- entry:
			return nil
		case <-timer.C:
			return w.handleQueueFull()
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	select {
	case w.queue <- entry:
		return nil
	default:
		return w.handleQueueFull()
	}
}

func (w *AsyncRequestLogWriter) Stop(ctx context.Context) error {
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

func (w *AsyncRequestLogWriter) handleQueueFull() error {
	w.logger.Warn("product request log queue full; dropping log")

	if w.cfg.DropOnFull {
		return nil
	}

	return ErrRequestLogQueueFull
}

func (w *AsyncRequestLogWriter) worker() {
	defer w.wg.Done()

	ticker := time.NewTicker(w.cfg.FlushInterval)
	defer ticker.Stop()

	batch := make([]platform.RequestLogEntry, 0, w.cfg.BatchSize)
	flush := func() {
		if len(batch) == 0 {
			return
		}

		w.flushBatch(context.Background(), batch)
		batch = make([]platform.RequestLogEntry, 0, w.cfg.BatchSize)
	}

	for {
		select {
		case entry, ok := <-w.queue:
			if !ok {
				flush()
				return
			}

			batch = append(batch, entry)
			if len(batch) >= w.cfg.BatchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}

func (w *AsyncRequestLogWriter) flushBatch(ctx context.Context, entries []platform.RequestLogEntry) {
	if len(entries) == 0 || w.underlying == nil {
		return
	}

	var err error
	if w.batch != nil {
		err = w.batch.LogRequests(ctx, entries)
	} else {
		for _, entry := range entries {
			if writeErr := w.underlying.LogRequest(ctx, entry); writeErr != nil {
				err = writeErr
				break
			}
		}
	}

	if err != nil {
		if errors.Is(err, storage.ErrLogStoreUnavailable) {
			return
		}

		w.logger.Warn("failed to flush product request log batch", zap.Error(err), zap.Int("batch_size", len(entries)))
	}
}
