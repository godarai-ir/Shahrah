package storage

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/platformcredentials"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
)

// Store provides PostgreSQL-backed persistence for the control-plane.
type Store struct {
	*PlanRepository
	*WalletRepository
	*SupportRepository
	*ProviderAccountRepository

	pool                     *pgxpool.Pool
	clickhouse               *sql.DB
	clickhouseReady          bool
	clickhouseTraceReady     bool
	clickhouseMetricsReady   bool
	clickhouseLogsReady      bool
	logger                   *zap.Logger
	entitlementCache         cache.Cache
	requestLoggingModeTTL    time.Duration
	requestLoggingModeGroup  singleflight.Group
	tokenUsageCache          cache.TokenUsageCache
	tokenUsageMetrics        TokenUsageMetrics
	tokenUsageEnabled        bool
	tokenUsageFailOpen       bool
	tokenUsageKeyPrefix      string
	tokenUsageFlushInterval  time.Duration
	tokenUsageFlushBatchSize int
	tokenUsageDeltaTTL       time.Duration
	lastUsedFlushInterval    time.Duration
	lastUsedTTL              time.Duration
	lastUsedFallbackMu       sync.Mutex
	lastUsedFallback         map[string]time.Time

	hotPathCacheMu       sync.Mutex
	hotPathCacheTTL      time.Duration
	hotPathCacheStaleTTL time.Duration
	hotPathCache         map[string]hotPathCacheEntry
	hotPathLoadGroup     singleflight.Group
	hotPathRefreshGroup  singleflight.Group

	authzCacheMu         sync.RWMutex
	authzCache           map[string]authzCacheEntry
	authzTenantEpoch     map[string]uint64
	authzCacheTTL        time.Duration
	authzCacheMaxEntries int

	runtimeProfileInvalidator RuntimeProfileInvalidator
}

// StoreDependencies contains the non-database collaborators required by repositories.
type StoreDependencies struct {
	Cache                     cache.Cache
	RequestLoggingModeTTL     time.Duration
	RuntimeProfileInvalidator RuntimeProfileInvalidator
	TenantCredentialStore     *platformcredentials.EncryptedDBSecretStore
	TenantProviderProber      platformcredentials.CredentialProber
	TokenUsageMetrics         TokenUsageMetrics
}

// NewStore creates a new PostgreSQL store.
func NewStore(
	ctx context.Context,
	databaseURL string,
	logger *zap.Logger,
	dependencies ...StoreDependencies,
) (*Store, error) {
	if len(dependencies) > 1 {
		return nil, fmt.Errorf("initialize store: expected at most one dependency set, got %d", len(dependencies))
	}

	deps := StoreDependencies{}
	if len(dependencies) == 1 {
		deps = dependencies[0]
	}

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

	requestLoggingModeTTL := secondsEnvOrDefault("REQUEST_LOG_MODE_CACHE_TTL_SECONDS", 60*time.Second)
	if deps.RequestLoggingModeTTL > 0 {
		requestLoggingModeTTL = deps.RequestLoggingModeTTL
	}

	store := &Store{
		pool:                      pool,
		logger:                    logger,
		entitlementCache:          deps.Cache,
		requestLoggingModeTTL:     requestLoggingModeTTL,
		tokenUsageEnabled:         boolEnvOrDefault("TOKEN_USAGE_AGGREGATION_ENABLED", true),
		tokenUsageFailOpen:        boolEnvOrDefault("TOKEN_USAGE_REDIS_FAIL_OPEN", true),
		tokenUsageKeyPrefix:       stringEnvOrDefault("TOKEN_USAGE_REDIS_KEY_PREFIX", "usage"),
		tokenUsageFlushInterval:   secondsEnvOrDefault("TOKEN_USAGE_FLUSH_INTERVAL_SECONDS", 5*time.Second),
		tokenUsageFlushBatchSize:  intEnvOrDefault("TOKEN_USAGE_FLUSH_BATCH_SIZE", 500),
		tokenUsageDeltaTTL:        secondsEnvOrDefault("TOKEN_USAGE_DELTA_TTL_SECONDS", 86400*time.Second),
		lastUsedFlushInterval:     secondsEnvOrDefault("TOKEN_LAST_USED_FLUSH_INTERVAL_SECONDS", 60*time.Second),
		lastUsedTTL:               secondsEnvOrDefault("TOKEN_LAST_USED_TTL_SECONDS", 86400*time.Second),
		lastUsedFallback:          make(map[string]time.Time),
		hotPathCacheTTL:           secondsEnvOrDefault("MODEL_LIST_CACHE_TTL_SECONDS", 30*time.Second),
		hotPathCacheStaleTTL:      secondsEnvOrDefault("MODEL_LIST_CACHE_STALE_SECONDS", 10*time.Second),
		hotPathCache:              make(map[string]hotPathCacheEntry),
		authzCache:                make(map[string]authzCacheEntry),
		authzTenantEpoch:          make(map[string]uint64),
		authzCacheTTL:             60 * time.Second,
		authzCacheMaxEntries:      5000,
		runtimeProfileInvalidator: deps.RuntimeProfileInvalidator,
		tokenUsageMetrics:         deps.TokenUsageMetrics,
	}
	if tokenUsageCache, ok := deps.Cache.(cache.TokenUsageCache); ok {
		store.tokenUsageCache = tokenUsageCache
	}

	if err := store.initClickHouseWithRetry(ctx); err != nil {
		if boolEnvOrDefault("CLICKHOUSE_REQUIRED", false) {
			pool.Close()
			return nil, fmt.Errorf("initialize clickhouse observability backend: %w", err)
		}

		logger.Warn("clickhouse observability backend unavailable; continuing without clickhouse-backed traces, metrics, and logs", zap.Error(err))
	}

	store.PlanRepository = newPlanRepository(pool, logger, deps.Cache, deps.RuntimeProfileInvalidator)
	store.WalletRepository = newWalletRepository(
		pool,
		store.clickhouse,
		store.clickHouseMetricsReady(),
		logger,
		deps.Cache,
	)
	store.SupportRepository = newSupportRepository(pool, store.PlanRepository, store.WalletRepository)
	store.ProviderAccountRepository = newProviderAccountRepository(
		pool,
		logger,
		deps.RuntimeProfileInvalidator,
		store.invalidateDeploymentCache,
		store.warmDeploymentCacheAsync,
		deps.TenantCredentialStore,
		deps.TenantProviderProber,
	)

	return store, nil
}

// NewMetricsStore creates the ClickHouse-only store used by background metrics
// workers. It intentionally does not initialize PostgreSQL.
func NewMetricsStore(ctx context.Context, logger *zap.Logger) (*Store, error) {
	store := &Store{logger: logger}
	if err := store.initClickHouseWithRetry(ctx); err != nil {
		return nil, fmt.Errorf("initialize clickhouse metrics backend: %w", err)
	}

	return store, nil
}

// Close closes the store.
func (s *Store) Close() {
	if s.clickhouse != nil {
		closeBestEffort(s.clickhouse)
	}

	if s.pool != nil {
		s.pool.Close()
	}
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
