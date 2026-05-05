package storage

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Store provides PostgreSQL-backed persistence for the control-plane.
type Store struct {
	pool            *pgxpool.Pool
	clickhouse      *sql.DB
	clickhouseReady bool
	logger          *zap.Logger
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

	store := &Store{
		pool:   pool,
		logger: logger,
	}
	if err := store.initClickHouse(ctx); err != nil {
		logger.Warn("clickhouse trace backend unavailable; continuing without trace storage", zap.Error(err))
	}
	return store, nil
}

// Close closes the store.
func (s *Store) Close() {
	if s.clickhouse != nil {
		_ = s.clickhouse.Close()
	}
	s.pool.Close()
}
