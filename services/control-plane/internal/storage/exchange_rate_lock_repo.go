package storage

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const exchangeRateUpdateAdvisoryLockKey = "exchange_rate_update:nobitex"

// TryExchangeRateUpdateLock acquires a PostgreSQL advisory lock for exchange-rate refresh work.
func (s *ExchangeRateRepository) TryExchangeRateUpdateLock(ctx context.Context) (bool, func(), error) {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return false, nil, err
	}

	var acquired bool
	if err := conn.QueryRow(ctx, `SELECT pg_try_advisory_lock(hashtext($1))`, exchangeRateUpdateAdvisoryLockKey).Scan(&acquired); err != nil {
		conn.Release()
		return false, nil, err
	}

	if !acquired {
		conn.Release()
		return false, nil, nil
	}

	var released bool

	release := func() {
		if released {
			return
		}

		released = true

		releaseAdvisoryLock(conn, exchangeRateUpdateAdvisoryLockKey)
	}

	return true, release, nil
}

func releaseAdvisoryLock(conn *pgxpool.Conn, key string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var unlocked bool

	err := conn.QueryRow(ctx, `SELECT pg_advisory_unlock(hashtext($1))`, key).Scan(&unlocked)
	if err != nil || !unlocked {
		raw := conn.Hijack()
		if closeErr := raw.Close(context.Background()); closeErr != nil {
			return
		}

		return
	}

	conn.Release()
}
