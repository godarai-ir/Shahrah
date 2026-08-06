package storage

import (
	"context"
	"database/sql"
	"strings"
	"sync"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/platformcredentials"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
)

// RuntimeProfileInvalidator invalidates compiled tenant runtime profiles after mutations.
type RuntimeProfileInvalidator interface {
	Invalidate(ctx context.Context, tenantID string, reason string)
}

// RuntimeProfileInvalidatorFunc adapts a function to RuntimeProfileInvalidator.
type RuntimeProfileInvalidatorFunc func(ctx context.Context, tenantID string, reason string)

// Invalidate calls f when it is non-nil.
func (f RuntimeProfileInvalidatorFunc) Invalidate(ctx context.Context, tenantID string, reason string) {
	if f != nil {
		f(ctx, tenantID, reason)
	}
}

// PlanRepository persists commercial plans, subscriptions, and entitlement sources.
type PlanRepository struct {
	pool                       *pgxpool.Pool
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
	runtimeProfileInvalidator  RuntimeProfileInvalidator
}

// WalletRepository persists wallet state and payment intents.
type WalletRepository struct {
	pool                   *pgxpool.Pool
	clickhouse             *sql.DB
	clickhouseMetricsReady bool
	logger                 *zap.Logger
	entitlementCache       cache.Cache
}

// SupportRepository persists support tickets and support-only tenant operations.
type SupportRepository struct {
	pool    *pgxpool.Pool
	plans   *PlanRepository
	wallets *WalletRepository
}

// ProviderAccountRepository persists tenant provider accounts and encrypted credential versions.
type ProviderAccountRepository struct {
	pool                          *pgxpool.Pool
	logger                        *zap.Logger
	runtimeProfileInvalidator     RuntimeProfileInvalidator
	invalidateDeploymentCacheFunc func(context.Context, string)
	warmDeploymentCacheAsyncFunc  func(string)
	tenantCredentialStore         *platformcredentials.EncryptedDBSecretStore
	tenantCredentialCacheMu       sync.Mutex
	tenantCredentialCache         map[string]tenantCredentialCacheEntry
	tenantCredentialCacheTTL      time.Duration
	tenantCredentialLoadGroup     singleflight.Group
	tenantProviderProber          platformcredentials.CredentialProber
}

func newPlanRepository(
	pool *pgxpool.Pool,
	logger *zap.Logger,
	sharedCache cache.Cache,
	invalidator RuntimeProfileInvalidator,
) *PlanRepository {
	repository := &PlanRepository{
		pool:                       pool,
		logger:                     logger,
		entitlementCache:           sharedCache,
		entitlementCacheTTL:        secondsEnvOrDefault("ENTITLEMENT_CACHE_TTL_SECONDS", 60*time.Second),
		usageCountersEnabled:       boolEnvOrDefault("USAGE_COUNTER_REDIS_ENABLED", true),
		usageCountersFailOpen:      boolEnvOrDefault("USAGE_COUNTER_REDIS_FAIL_OPEN", false),
		usageCounterKeyPrefix:      stringEnvOrDefault("USAGE_COUNTER_REDIS_KEY_PREFIX", "usage"),
		usageCounterFlushInterval:  secondsEnvOrDefault("USAGE_COUNTER_FLUSH_INTERVAL_SECONDS", 5*time.Second),
		usageCounterFlushBatchSize: intEnvOrDefault("USAGE_COUNTER_FLUSH_BATCH_SIZE", 500),
		usageCounterTTLBuffer:      secondsEnvOrDefault("USAGE_COUNTER_TTL_BUFFER_SECONDS", 3600*time.Second),
		runtimeProfileInvalidator:  invalidator,
	}

	if usageCounterCache, ok := sharedCache.(cache.UsageCounterCache); ok {
		repository.usageCounterCache = usageCounterCache
	}

	return repository
}

func newWalletRepository(
	pool *pgxpool.Pool,
	clickhouse *sql.DB,
	clickhouseMetricsReady bool,
	logger *zap.Logger,
	sharedCache cache.Cache,
) *WalletRepository {
	return &WalletRepository{
		pool:                   pool,
		clickhouse:             clickhouse,
		clickhouseMetricsReady: clickhouseMetricsReady,
		logger:                 logger,
		entitlementCache:       sharedCache,
	}
}

func newSupportRepository(
	pool *pgxpool.Pool,
	plans *PlanRepository,
	wallets *WalletRepository,
) *SupportRepository {
	return &SupportRepository{
		pool:    pool,
		plans:   plans,
		wallets: wallets,
	}
}

func newProviderAccountRepository(
	pool *pgxpool.Pool,
	logger *zap.Logger,
	invalidator RuntimeProfileInvalidator,
	invalidateDeploymentCache func(context.Context, string),
	warmDeploymentCacheAsync func(string),
	credentialStore *platformcredentials.EncryptedDBSecretStore,
	prober platformcredentials.CredentialProber,
) *ProviderAccountRepository {
	return &ProviderAccountRepository{
		pool:                          pool,
		logger:                        logger,
		runtimeProfileInvalidator:     invalidator,
		invalidateDeploymentCacheFunc: invalidateDeploymentCache,
		warmDeploymentCacheAsyncFunc:  warmDeploymentCacheAsync,
		tenantCredentialStore:         credentialStore,
		tenantCredentialCache:         make(map[string]tenantCredentialCacheEntry),
		tenantCredentialCacheTTL:      secondsEnvOrDefault("TENANT_PROVIDER_CREDENTIAL_CACHE_TTL_SECONDS", 60*time.Second),
		tenantProviderProber:          prober,
	}
}

func (r *ProviderAccountRepository) invalidateDeploymentCache(ctx context.Context, tenantID string) {
	if r != nil && r.invalidateDeploymentCacheFunc != nil {
		r.invalidateDeploymentCacheFunc(ctx, tenantID)
	}
}

func (r *ProviderAccountRepository) warmDeploymentCacheAsync(tenantID string) {
	if r != nil && r.warmDeploymentCacheAsyncFunc != nil {
		r.warmDeploymentCacheAsyncFunc(tenantID)
	}
}

func (r *PlanRepository) invalidateTenantRuntimeProfile(
	ctx context.Context,
	tenantID string,
	reason string,
) {
	if r == nil || r.runtimeProfileInvalidator == nil {
		return
	}

	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return
	}

	r.runtimeProfileInvalidator.Invalidate(ctx, tenantID, reason)
}

func (r *ProviderAccountRepository) invalidateTenantRuntimeProfile(
	ctx context.Context,
	tenantID string,
	reason string,
) {
	if r == nil || r.runtimeProfileInvalidator == nil {
		return
	}

	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return
	}

	r.runtimeProfileInvalidator.Invalidate(ctx, tenantID, reason)
}

func (r *WalletRepository) invalidateWalletCache(ctx context.Context, tenantID string) {
	if r == nil || r.entitlementCache == nil {
		return
	}

	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return
	}

	r.entitlementCache.Delete(ctx, walletCacheKey(tenantID))
	r.syncCurrentBudgetWindowAllowance(ctx, tenantID)
}
