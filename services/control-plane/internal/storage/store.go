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
	requestLoggingModeTTL      time.Duration
	requestLoggingModeGroup    singleflight.Group
	usageCounterCache          cache.UsageCounterCache
	usageCountersEnabled       bool
	usageCountersFailOpen      bool
	usageCounterKeyPrefix      string
	usageCounterFlushInterval  time.Duration
	usageCounterFlushBatchSize int
	usageCounterTTLBuffer      time.Duration
	tokenUsageCache            cache.TokenUsageCache
	tokenUsageMetrics          TokenUsageMetrics
	tokenUsageEnabled          bool
	tokenUsageFailOpen         bool
	tokenUsageKeyPrefix        string
	tokenUsageFlushInterval    time.Duration
	tokenUsageFlushBatchSize   int
	tokenUsageDeltaTTL         time.Duration
	lastUsedFlushInterval      time.Duration
	lastUsedTTL                time.Duration
	lastUsedFallbackMu         sync.Mutex
	lastUsedFallback           map[string]time.Time

	hotPathCacheMu   sync.Mutex
	hotPathCacheTTL  time.Duration
	hotPathCache     map[string]hotPathCacheEntry
	hotPathLoadGroup singleflight.Group

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
	if tokenUsageCache, ok := c.(cache.TokenUsageCache); ok {
		s.tokenUsageCache = tokenUsageCache
	}
}

// SetRequestLoggingModeCacheTTL overrides the tenant request logging mode cache TTL.
func (s *Store) SetRequestLoggingModeCacheTTL(ttl time.Duration) {
	if ttl > 0 {
		s.requestLoggingModeTTL = ttl
	}
}

// SetTokenUsageMetrics attaches token usage aggregation metric hooks.
func (s *Store) SetTokenUsageMetrics(metrics TokenUsageMetrics) {
	s.tokenUsageMetrics = metrics
}

// NewStore creates a new PostgreSQL store.
func NewStore(ctx context.Context, databaseURL string, logger *zap.Logger) (*Store, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database config: %w", err)
	}
	applyPostgresPoolEnv(cfg)

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
		requestLoggingModeTTL:      secondsEnvOrDefault("REQUEST_LOG_MODE_CACHE_TTL_SECONDS", 60*time.Second),
		usageCountersEnabled:       boolEnvOrDefault("USAGE_COUNTER_REDIS_ENABLED", true),
		usageCountersFailOpen:      boolEnvOrDefault("USAGE_COUNTER_REDIS_FAIL_OPEN", false),
		usageCounterKeyPrefix:      stringEnvOrDefault("USAGE_COUNTER_REDIS_KEY_PREFIX", "usage"),
		usageCounterFlushInterval:  secondsEnvOrDefault("USAGE_COUNTER_FLUSH_INTERVAL_SECONDS", 5*time.Second),
		usageCounterFlushBatchSize: intEnvOrDefault("USAGE_COUNTER_FLUSH_BATCH_SIZE", 500),
		usageCounterTTLBuffer:      secondsEnvOrDefault("USAGE_COUNTER_TTL_BUFFER_SECONDS", 3600*time.Second),
		tokenUsageEnabled:          boolEnvOrDefault("TOKEN_USAGE_AGGREGATION_ENABLED", true),
		tokenUsageFailOpen:         boolEnvOrDefault("TOKEN_USAGE_REDIS_FAIL_OPEN", true),
		tokenUsageKeyPrefix:        stringEnvOrDefault("TOKEN_USAGE_REDIS_KEY_PREFIX", "usage"),
		tokenUsageFlushInterval:    secondsEnvOrDefault("TOKEN_USAGE_FLUSH_INTERVAL_SECONDS", 5*time.Second),
		tokenUsageFlushBatchSize:   intEnvOrDefault("TOKEN_USAGE_FLUSH_BATCH_SIZE", 500),
		tokenUsageDeltaTTL:         secondsEnvOrDefault("TOKEN_USAGE_DELTA_TTL_SECONDS", 86400*time.Second),
		lastUsedFlushInterval:      secondsEnvOrDefault("TOKEN_LAST_USED_FLUSH_INTERVAL_SECONDS", 60*time.Second),
		lastUsedTTL:                secondsEnvOrDefault("TOKEN_LAST_USED_TTL_SECONDS", 86400*time.Second),
		lastUsedFallback:           make(map[string]time.Time),
		hotPathCacheTTL:            secondsEnvOrDefault("MODEL_LIST_CACHE_TTL_SECONDS", 30*time.Second),
		hotPathCache:               make(map[string]hotPathCacheEntry),
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

// PoolStats returns a snapshot of PostgreSQL pool state for observability.
func (s *Store) PoolStats() *pgxpool.Stat {
	if s == nil || s.pool == nil {
		return nil
	}

	return s.pool.Stat()
}

func applyPostgresPoolEnv(cfg *pgxpool.Config) {
	if cfg == nil {
		return
	}

	cfg.MaxConns = int32EnvOrDefault("POSTGRES_POOL_MAX_CONNS", 32)
	cfg.MinConns = int32EnvOrDefault("POSTGRES_POOL_MIN_CONNS", 4)
	cfg.MaxConnLifetime = secondsEnvOrDefault("POSTGRES_POOL_MAX_CONN_LIFETIME_SECONDS", 1800*time.Second)
	cfg.MaxConnIdleTime = secondsEnvOrDefault("POSTGRES_POOL_MAX_CONN_IDLE_TIME_SECONDS", 300*time.Second)
	cfg.HealthCheckPeriod = secondsEnvOrDefault("POSTGRES_POOL_HEALTH_CHECK_PERIOD_SECONDS", 30*time.Second)

	if cfg.MinConns > cfg.MaxConns {
		cfg.MinConns = cfg.MaxConns
	}
}
