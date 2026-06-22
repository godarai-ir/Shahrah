package billing

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
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

func NewSubscriptionProcessor(store subscriptionTransitionStore, logger *zap.Logger, interval time.Duration) *SubscriptionProcessor {
	if interval <= 0 {
		interval = time.Minute
	}
	return &SubscriptionProcessor{store: store, logger: logger, interval: interval, done: make(chan struct{})}
}

func (p *SubscriptionProcessor) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel
	go p.run(ctx)
}

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
	p.process(ctx)
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			workCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			p.process(workCtx, now)
			cancel()
		}
	}
}

func (p *SubscriptionProcessor) process(ctx context.Context, now ...time.Time) {
	current := time.Now().UTC()
	if len(now) > 0 {
		current = now[0].UTC()
	}
	processed, err := p.store.ProcessSubscriptionTransitions(ctx, current, 100)
	if err != nil {
		p.logger.Error("subscription lifecycle processing failed", zap.Error(err))
		return
	}
	if processed > 0 {
		p.logger.Info("subscription lifecycle transitions applied", zap.Int("count", processed))
	}
}
