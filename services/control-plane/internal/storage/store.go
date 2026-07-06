package storage

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
)

// Store provides PostgreSQL-backed persistence for the control-plane.
type Store struct {
	pool                       *pgxpool.Pool
	clickhouse                 *sql.DB
	clickhouseReady            bool
	logger                     *zap.Logger
	entitlementCache           cache.Cache
	entitlementCacheTTL        time.Duration
	entitlementLoadGroup       singleflight.Group
	usageCounterCache          cache.UsageCounterCache
	usageCountersEnabled       bool
	usageCountersFailOpen      bool
	usageCounterKeyPrefix      string
	usageCounterFlushInterval  time.Duration
	usageCounterFlushBatchSize int
	usageCounterTTLBuffer      time.Duration

	authzCacheMu         sync.RWMutex
	authzCache           map[string]authzCacheEntry
	authzTenantEpoch     map[string]uint64
	authzCacheTTL        time.Duration
	authzCacheMaxEntries int
}

// SetEntitlementCache attaches the shared control-plane cache used by entitlement resolution.
func (s *Store) SetEntitlementCache(c cache.Cache) {
	s.entitlementCache = c
	if usageCounterCache, ok := c.(cache.UsageCounterCache); ok {
		s.usageCounterCache = usageCounterCache
	}
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
		pool:                       pool,
		logger:                     logger,
		entitlementCacheTTL:        secondsEnvOrDefault("ENTITLEMENT_CACHE_TTL_SECONDS", 60*time.Second),
		usageCountersEnabled:       boolEnvOrDefault("USAGE_COUNTER_REDIS_ENABLED", true),
		usageCountersFailOpen:      boolEnvOrDefault("USAGE_COUNTER_REDIS_FAIL_OPEN", false),
		usageCounterKeyPrefix:      stringEnvOrDefault("USAGE_COUNTER_REDIS_KEY_PREFIX", "usage"),
		usageCounterFlushInterval:  secondsEnvOrDefault("USAGE_COUNTER_FLUSH_INTERVAL_SECONDS", 5*time.Second),
		usageCounterFlushBatchSize: intEnvOrDefault("USAGE_COUNTER_FLUSH_BATCH_SIZE", 500),
		usageCounterTTLBuffer:      secondsEnvOrDefault("USAGE_COUNTER_TTL_BUFFER_SECONDS", 3600*time.Second),
		authzCache:                 make(map[string]authzCacheEntry),
		authzTenantEpoch:           make(map[string]uint64),
		authzCacheTTL:              60 * time.Second,
		authzCacheMaxEntries:       5000,
	}
	if err := store.initClickHouse(ctx); err != nil {
		logger.Warn("clickhouse trace backend unavailable; continuing without trace storage", zap.Error(err))
	}

	return store, nil
}

// Close closes the store.
func (s *Store) Close() {
	if s.clickhouse != nil {
		closeBestEffort(s.clickhouse)
	}

	s.pool.Close()
}
