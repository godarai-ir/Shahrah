package billing

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

const spendPurgeTimeout = 2 * time.Minute

type spendPurgeStore interface {
	RunPendingSpendPurges(ctx context.Context) (int, error)
}

// SpendPurgeProcessor runs the one-time ClickHouse spend purges that the
// credit cutover requested. It tries at startup and retries on an interval
// until no purge is pending, then exits.
type SpendPurgeProcessor struct {
	store    spendPurgeStore
	logger   *zap.Logger
	interval time.Duration
	cancel   context.CancelFunc
	done     chan struct{}
	once     sync.Once
}

// NewSpendPurgeProcessor creates the spend purge worker.
func NewSpendPurgeProcessor(store spendPurgeStore, logger *zap.Logger, interval time.Duration) *SpendPurgeProcessor {
	if logger == nil {
		logger = zap.NewNop()
	}

	if interval <= 0 {
		interval = 5 * time.Minute
	}

	return &SpendPurgeProcessor{store: store, logger: logger, interval: interval, done: make(chan struct{})}
}

// Start launches the worker.
func (p *SpendPurgeProcessor) Start() {
	ctx, cancel := context.WithCancel(context.Background())

	p.cancel = cancel
	go p.run(ctx)
}

// Stop waits for the worker to shut down.
func (p *SpendPurgeProcessor) Stop() {
	p.once.Do(func() {
		if p.cancel != nil {
			p.cancel()
		}

		<-p.done
	})
}

func (p *SpendPurgeProcessor) run(ctx context.Context) {
	defer close(p.done)

	if p.store == nil || p.purge(ctx) {
		return
	}

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if p.purge(ctx) {
				return
			}
		}
	}
}

// purge runs the pending purges once and reports whether none is left.
func (p *SpendPurgeProcessor) purge(ctx context.Context) bool {
	workCtx, cancel := context.WithTimeout(ctx, spendPurgeTimeout)
	defer cancel()

	pending, err := p.store.RunPendingSpendPurges(workCtx)
	if err != nil {
		p.logger.Warn("clickhouse spend purge failed; retrying", zap.Int("pending", pending), zap.Error(err))

		return false
	}

	return pending == 0
}
