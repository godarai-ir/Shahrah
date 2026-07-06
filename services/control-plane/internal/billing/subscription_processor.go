package billing

import (
	"context"
	"errors"
	"sync"
	"time"

	"go.uber.org/zap"
)

const (
	subscriptionProcessBatchSize = 100
	subscriptionProcessTimeout   = 30 * time.Second
)

type subscriptionTransitionStore interface {
	ProcessSubscriptionTransitions(ctx context.Context, now time.Time, limit int) (int, error)
}

// SubscriptionProcessor applies period-end subscription transitions.
type SubscriptionProcessor struct {
	store    subscriptionTransitionStore
	logger   *zap.Logger
	interval time.Duration
	cancel   context.CancelFunc
	done     chan struct{}
	once     sync.Once
}

// NewSubscriptionProcessor creates a background worker for subscription transitions.
func NewSubscriptionProcessor(store subscriptionTransitionStore, logger *zap.Logger, interval time.Duration) *SubscriptionProcessor {
	if interval <= 0 {
		interval = time.Minute
	}

	return &SubscriptionProcessor{store: store, logger: logger, interval: interval, done: make(chan struct{})}
}

// Start launches the subscription processor loop.
func (p *SubscriptionProcessor) Start() {
	ctx, cancel := context.WithCancel(context.Background())

	p.cancel = cancel
	go p.run(ctx)
}

// Stop waits for the processor to shut down.
func (p *SubscriptionProcessor) Stop() {
	p.once.Do(func() {
		if p.cancel != nil {
			p.cancel()
		}

		<-p.done
	})
}

func (p *SubscriptionProcessor) run(ctx context.Context) {
	defer close(p.done)

	p.processWithTimeout(ctx, time.Now())

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			p.processWithTimeout(ctx, now)
		}
	}
}

func (p *SubscriptionProcessor) processWithTimeout(ctx context.Context, now time.Time) {
	workCtx, cancel := context.WithTimeout(ctx, subscriptionProcessTimeout)
	defer cancel()

	p.process(workCtx, now)
}

func (p *SubscriptionProcessor) process(ctx context.Context, now ...time.Time) {
	current := time.Now().UTC()
	if len(now) > 0 {
		current = now[0].UTC()
	}

	started := time.Now()
	processed, err := p.store.ProcessSubscriptionTransitions(ctx, current, subscriptionProcessBatchSize)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}

		p.logger.Error(
			"subscription lifecycle processing failed",
			zap.Error(err),
			zap.Duration("elapsed", time.Since(started)),
			zap.Int("batch_size", subscriptionProcessBatchSize),
			zap.Time("now", current),
		)
		return
	}

	if processed > 0 {
		p.logger.Info(
			"subscription lifecycle transitions applied",
			zap.Int("count", processed),
			zap.Duration("elapsed", time.Since(started)),
		)
	}
}
