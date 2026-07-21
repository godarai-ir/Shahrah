package storage

import (
	"context"
	"time"

	"go.uber.org/zap"
)

func (s *Store) startPeriodicFlusher(
	ctx context.Context,
	enabled bool,
	interval time.Duration,
	name string,
	flush func(context.Context) error,
) {
	if !enabled || interval <= 0 {
		return
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				flushCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				err := flush(flushCtx)

				cancel()

				if err != nil && s.logger != nil {
					s.logger.Warn(name+" shutdown flush failed", zap.Error(err))
				}

				return
			case <-ticker.C:
				if err := flush(ctx); err != nil && s.logger != nil {
					s.logger.Warn(name+" flush failed", zap.Error(err))
				}
			}
		}
	}()
}
