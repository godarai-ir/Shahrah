package storage

import (
	"context"
	"fmt"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/platformcredentials"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Store provides PostgreSQL-backed persistence for the control-plane.
type Store struct {
	*PlanRepository
	*WalletRepository
	*SupportRepository
	*ProviderAccountRepository
	*PromptRepository
	*IAMRepository
	*TenantRepository
	*AuthRepository
	*TenantAdminRepository
	*PolicyRepository
	*RetentionRepository
	*ModelRepository
	*ObservabilityRepository
	*PlatformCredentialRepository
	*ProviderResourceRepository
	*RuntimeProfileRepository
	*PlatformServiceAccountRepository
	*DatabaseHealthRepository

	pool *pgxpool.Pool
}

// StoreDependencies contains the optional non-database collaborators used by repositories.
type StoreDependencies struct {
	Config                    *Config
	Cache                     cache.Cache
	RequestLoggingModeTTL     time.Duration
	RuntimeProfileInvalidator RuntimeProfileInvalidator
	TenantCredentialStore     *platformcredentials.EncryptedDBSecretStore
	TenantProviderProber      platformcredentials.CredentialProber
	TokenUsageMetrics         TokenUsageMetrics
}

// NewStore creates a new PostgreSQL store with an explicit dependency set.
func NewStore(
	ctx context.Context,
	databaseURL string,
	logger *zap.Logger,
	dependencies StoreDependencies,
) (*Store, error) {
	storageConfig := DefaultConfig()
	if dependencies.Config != nil {
		storageConfig = *dependencies.Config
	}

	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database config: %w", err)
	}

	applyPostgresPoolConfig(cfg, storageConfig.Postgres)

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	requestLoggingModeTTL := storageConfig.HotPath.RequestLoggingModeTTL
	if dependencies.RequestLoggingModeTTL > 0 {
		requestLoggingModeTTL = dependencies.RequestLoggingModeTTL
	}

	observability := newObservabilityRepository(
		pool,
		logger,
		dependencies.Cache,
		dependencies.TokenUsageMetrics,
		requestLoggingModeTTL,
		storageConfig,
	)
	store := &Store{pool: pool, ObservabilityRepository: observability}

	if err := store.initClickHouseWithRetry(ctx); err != nil {
		if storageConfig.ClickHouse.Required {
			pool.Close()
			return nil, fmt.Errorf("initialize clickhouse observability backend: %w", err)
		}

		logger.Warn("clickhouse observability backend unavailable; continuing without clickhouse-backed traces, metrics, and logs", zap.Error(err))
	}

	store.PlanRepository = newPlanRepository(
		pool,
		logger,
		dependencies.Cache,
		dependencies.RuntimeProfileInvalidator,
		storageConfig,
	)
	store.WalletRepository = newWalletRepository(
		pool,
		observability.clickhouse,
		observability.clickHouseMetricsReady(),
		logger,
		dependencies.Cache,
		storageConfig,
	)
	store.SupportRepository = newSupportRepository(pool, store.PlanRepository, store.WalletRepository)
	store.ModelRepository = newModelRepository(
		pool,
		dependencies.Cache,
		dependencies.RuntimeProfileInvalidator,
		storageConfig,
	)
	store.ProviderAccountRepository = newProviderAccountRepository(
		pool,
		logger,
		dependencies.RuntimeProfileInvalidator,
		store.ModelRepository.invalidateDeploymentCache,
		store.ModelRepository.warmDeploymentCacheAsync,
		dependencies.TenantCredentialStore,
		dependencies.TenantProviderProber,
		storageConfig,
	)
	store.PromptRepository = newPromptRepository(pool)
	store.IAMRepository = newIAMRepository(pool, dependencies.RuntimeProfileInvalidator)
	store.TenantRepository = newTenantRepository(pool)
	store.AuthRepository = newAuthRepository(
		pool,
		dependencies.RuntimeProfileInvalidator,
		store.ModelRepository.ListVirtualModelsForPrincipal,
		observability.queueKeyLastUsed,
		storageConfig,
	)
	store.TenantAdminRepository = newTenantAdminRepository(pool, logger, store.AuthRepository)
	store.PolicyRepository = newPolicyRepository(pool, dependencies.RuntimeProfileInvalidator)
	store.RetentionRepository = newRetentionRepository(
		pool,
		observability.clickhouse,
		observability.clickHouseLogsReady(),
		logger,
	)
	observability.retention = store.RetentionRepository
	store.PlatformCredentialRepository = newPlatformCredentialRepository(pool, logger)
	store.ProviderResourceRepository = newProviderResourceRepository(pool)
	store.RuntimeProfileRepository = newRuntimeProfileRepository(
		pool,
		store.AuthRepository,
		dependencies.RuntimeProfileInvalidator,
	)
	store.PlatformServiceAccountRepository = newPlatformServiceAccountRepository(pool)
	store.DatabaseHealthRepository = newDatabaseHealthRepository(pool)

	return store, nil
}

// NewMetricsStore creates the ClickHouse-only store used by background metrics
// workers. It intentionally does not initialize PostgreSQL.
func NewMetricsStore(ctx context.Context, logger *zap.Logger, storageConfig Config) (*Store, error) {
	observability := newObservabilityRepository(nil, logger, nil, nil, 0, storageConfig)

	store := &Store{ObservabilityRepository: observability}
	if err := store.initClickHouseWithRetry(ctx); err != nil {
		return nil, fmt.Errorf("initialize clickhouse metrics backend: %w", err)
	}

	return store, nil
}

// Close closes the store.
func (s *Store) Close() {
	if s.ObservabilityRepository != nil && s.ObservabilityRepository.clickhouse != nil {
		closeBestEffort(s.ObservabilityRepository.clickhouse)
	}

	if s.pool != nil {
		s.pool.Close()
	}
}

// PoolStats returns a snapshot of PostgreSQL pool state for observability.
func (s *DatabaseHealthRepository) PoolStats() *pgxpool.Stat {
	if s == nil || s.pool == nil {
		return nil
	}

	return s.pool.Stat()
}

func applyPostgresPoolConfig(cfg *pgxpool.Config, settings PostgresConfig) {
	if cfg == nil {
		return
	}

	cfg.MaxConns = settings.MaxConnections
	cfg.MinConns = settings.MinConnections
	cfg.MaxConnLifetime = settings.MaxLifetime
	cfg.MaxConnIdleTime = settings.MaxIdleTime
	cfg.HealthCheckPeriod = settings.HealthCheckPeriod

	if cfg.MinConns > cfg.MaxConns {
		cfg.MinConns = cfg.MaxConns
	}
}
