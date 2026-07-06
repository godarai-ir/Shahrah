package storage

import (
	"context"
	"time"
)

const (
	defaultHotPathDBTimeout               = 3 * time.Second
	defaultSubscriptionLifecycleDBTimeout = 5 * time.Second
)

func withHotPathDBTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return withStoreTimeout(ctx, "GATEWAY_DB_TIMEOUT_SECONDS", defaultHotPathDBTimeout)
}

func withSubscriptionLifecycleDBTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return withStoreTimeout(ctx, "SUBSCRIPTION_LIFECYCLE_DB_TIMEOUT_SECONDS", defaultSubscriptionLifecycleDBTimeout)
}

func withStoreTimeout(ctx context.Context, envKey string, fallback time.Duration) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}

	timeout := secondsEnvOrDefault(envKey, fallback)
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
