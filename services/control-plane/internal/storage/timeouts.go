package storage

import (
	"context"
	"time"
)

const (
	defaultHotPathDBTimeout               = 3 * time.Second
	defaultSubscriptionLifecycleDBTimeout = 5 * time.Second
)

func withStoreTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}

	if timeout <= 0 {
		return ctx, func() {}
	}

	return context.WithTimeout(ctx, timeout)
}

func rollbackTxWithFreshTimeout(tx interface{ Rollback(context.Context) error }) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	rollbackTxBestEffort(ctx, tx)
}
