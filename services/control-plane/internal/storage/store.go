package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Store provides PostgreSQL-backed persistence for the control-plane.
type Store struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

// NewStore creates a new PostgreSQL store.
func NewStore(ctx context.Context, databaseURL string, logger *zap.Logger) (*Store, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database config: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return &Store{pool: pool, logger: logger}, nil
}

// Close closes the store.
func (s *Store) Close() {
	s.pool.Close()
}
