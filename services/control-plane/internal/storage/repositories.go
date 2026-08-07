package storage

import (
	"context"
	"database/sql"
	"strings"
	"sync"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
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
	hotPathDBTimeout           time.Duration
	subscriptionDBTimeout      time.Duration
}

// WalletRepository persists wallet state and payment intents.
type WalletRepository struct {
	pool                   *pgxpool.Pool
	clickhouse             *sql.DB
	clickhouseMetricsReady bool
	logger                 *zap.Logger
	entitlementCache       cache.Cache
	budgetWindowTTLGrace   time.Duration
	hotPathDBTimeout       time.Duration
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

// PromptRepository persists prompts and their immutable versions.
type PromptRepository struct {
	pool *pgxpool.Pool
}

// IAMRepository persists organizational units, teams, virtual users, and memberships.
type IAMRepository struct {
	pool                      *pgxpool.Pool
	runtimeProfileInvalidator RuntimeProfileInvalidator
}

// TenantRepository resolves tenant identity and administration scope.
type TenantRepository struct {
	pool *pgxpool.Pool
}

// AuthRepository persists API keys, role bindings, and authorization state.
type AuthRepository struct {
	pool                              *pgxpool.Pool
	authzCacheMu                      sync.RWMutex
	authzCache                        map[string]authzCacheEntry
	authzTenantEpoch                  map[string]uint64
	authzCacheTTL                     time.Duration
	authzCacheMaxEntries              int
	runtimeProfileInvalidator         RuntimeProfileInvalidator
	listVirtualModelsForPrincipalFunc func(context.Context, string, string, string) ([]platform.VirtualModel, error)
	queueKeyLastUsedFunc              func(context.Context, string, time.Time) error
}

// TenantAdminRepository persists tenant-admin authentication, onboarding, and membership state.
type TenantAdminRepository struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
	auth   *AuthRepository
}

// PolicyRepository persists tenant rate, budget, and guardrail policies.
type PolicyRepository struct {
	pool                      *pgxpool.Pool
	runtimeProfileInvalidator RuntimeProfileInvalidator
}

// RetentionRepository resolves observability retention and deletes expired data.
type RetentionRepository struct {
	pool                *pgxpool.Pool
	clickhouse          *sql.DB
	clickhouseLogsReady bool
	logger              *zap.Logger
}

// ModelRepository persists deployments and virtual models and owns their hot-path cache.
type ModelRepository struct {
	pool                      *pgxpool.Pool
	sharedCache               cache.Cache
	hotPathCacheMu            sync.Mutex
	hotPathCacheTTL           time.Duration
	hotPathCacheStaleTTL      time.Duration
	hotPathCache              map[string]hotPathCacheEntry
	hotPathLoadGroup          singleflight.Group
	hotPathRefreshGroup       singleflight.Group
	runtimeProfileInvalidator RuntimeProfileInvalidator
}

// ObservabilityRepository persists logs, traces, metrics, and usage aggregates.
type ObservabilityRepository struct {
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
	config                   Config
	retention                *RetentionRepository
}

// PlatformCredentialRepository persists platform-managed provider credentials.
type PlatformCredentialRepository struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

// ProviderResourceRepository persists provider-native asynchronous resources.
type ProviderResourceRepository struct {
	pool *pgxpool.Pool
}

// RuntimeProfileRepository persists compiled profile versions and source data.
type RuntimeProfileRepository struct {
	pool                      *pgxpool.Pool
	auth                      *AuthRepository
	runtimeProfileInvalidator RuntimeProfileInvalidator
}

// PlatformServiceAccountRepository persists platform service-account credentials.
type PlatformServiceAccountRepository struct {
	pool *pgxpool.Pool
}

// DatabaseHealthRepository exposes read-only PostgreSQL pool health statistics.
type DatabaseHealthRepository struct {
	pool *pgxpool.Pool
}

func newPlanRepository(
	pool *pgxpool.Pool,
	logger *zap.Logger,
	sharedCache cache.Cache,
	invalidator RuntimeProfileInvalidator,
	storageConfig Config,
) *PlanRepository {
	repository := &PlanRepository{
		pool:                       pool,
		logger:                     logger,
		entitlementCache:           sharedCache,
		entitlementCacheTTL:        storageConfig.Plans.EntitlementCacheTTL,
		usageCountersEnabled:       storageConfig.TokenUsage.UsageCounters,
		usageCountersFailOpen:      storageConfig.TokenUsage.CountersFailOpen,
		usageCounterKeyPrefix:      storageConfig.TokenUsage.CounterKeyPrefix,
		usageCounterFlushInterval:  storageConfig.TokenUsage.CounterInterval,
		usageCounterFlushBatchSize: storageConfig.TokenUsage.CounterBatchSize,
		usageCounterTTLBuffer:      storageConfig.TokenUsage.CounterTTLBuffer,
		runtimeProfileInvalidator:  invalidator,
		hotPathDBTimeout:           storageConfig.HotPath.DatabaseTimeout,
		subscriptionDBTimeout:      storageConfig.HotPath.SubscriptionLifecycleTimeout,
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
	storageConfig Config,
) *WalletRepository {
	return &WalletRepository{
		pool:                   pool,
		clickhouse:             clickhouse,
		clickhouseMetricsReady: clickhouseMetricsReady,
		logger:                 logger,
		entitlementCache:       sharedCache,
		budgetWindowTTLGrace:   storageConfig.Wallets.BudgetWindowTTLGrace,
		hotPathDBTimeout:       storageConfig.HotPath.DatabaseTimeout,
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
	storageConfig Config,
) *ProviderAccountRepository {
	return &ProviderAccountRepository{
		pool:                          pool,
		logger:                        logger,
		runtimeProfileInvalidator:     invalidator,
		invalidateDeploymentCacheFunc: invalidateDeploymentCache,
		warmDeploymentCacheAsyncFunc:  warmDeploymentCacheAsync,
		tenantCredentialStore:         credentialStore,
		tenantCredentialCache:         make(map[string]tenantCredentialCacheEntry),
		tenantCredentialCacheTTL:      storageConfig.ProviderAccount.CredentialCacheTTL,
		tenantProviderProber:          prober,
	}
}

func newPromptRepository(pool *pgxpool.Pool) *PromptRepository {
	return &PromptRepository{pool: pool}
}

func newIAMRepository(pool *pgxpool.Pool, invalidator RuntimeProfileInvalidator) *IAMRepository {
	return &IAMRepository{
		pool:                      pool,
		runtimeProfileInvalidator: invalidator,
	}
}

func newTenantRepository(pool *pgxpool.Pool) *TenantRepository {
	return &TenantRepository{pool: pool}
}

func newAuthRepository(
	pool *pgxpool.Pool,
	invalidator RuntimeProfileInvalidator,
	listVirtualModelsForPrincipal func(context.Context, string, string, string) ([]platform.VirtualModel, error),
	queueKeyLastUsed func(context.Context, string, time.Time) error,
	storageConfig Config,
) *AuthRepository {
	return &AuthRepository{
		pool:                              pool,
		authzCache:                        make(map[string]authzCacheEntry),
		authzTenantEpoch:                  make(map[string]uint64),
		authzCacheTTL:                     storageConfig.HotPath.AuthorizationCacheTTL,
		authzCacheMaxEntries:              storageConfig.HotPath.AuthorizationCacheMaxEntries,
		runtimeProfileInvalidator:         invalidator,
		listVirtualModelsForPrincipalFunc: listVirtualModelsForPrincipal,
		queueKeyLastUsedFunc:              queueKeyLastUsed,
	}
}

func newTenantAdminRepository(pool *pgxpool.Pool, logger *zap.Logger, auth *AuthRepository) *TenantAdminRepository {
	return &TenantAdminRepository{
		pool:   pool,
		logger: logger,
		auth:   auth,
	}
}

func newPolicyRepository(pool *pgxpool.Pool, invalidator RuntimeProfileInvalidator) *PolicyRepository {
	return &PolicyRepository{
		pool:                      pool,
		runtimeProfileInvalidator: invalidator,
	}
}

func newRetentionRepository(
	pool *pgxpool.Pool,
	clickhouse *sql.DB,
	clickhouseLogsReady bool,
	logger *zap.Logger,
) *RetentionRepository {
	return &RetentionRepository{
		pool:                pool,
		clickhouse:          clickhouse,
		clickhouseLogsReady: clickhouseLogsReady,
		logger:              logger,
	}
}

func newModelRepository(
	pool *pgxpool.Pool,
	sharedCache cache.Cache,
	invalidator RuntimeProfileInvalidator,
	storageConfig Config,
) *ModelRepository {
	return &ModelRepository{
		pool:                      pool,
		sharedCache:               sharedCache,
		hotPathCacheTTL:           storageConfig.HotPath.ModelCacheTTL,
		hotPathCacheStaleTTL:      storageConfig.HotPath.ModelCacheStaleTTL,
		hotPathCache:              make(map[string]hotPathCacheEntry),
		runtimeProfileInvalidator: invalidator,
	}
}

func newObservabilityRepository(
	pool *pgxpool.Pool,
	logger *zap.Logger,
	sharedCache cache.Cache,
	tokenUsageMetrics TokenUsageMetrics,
	requestLoggingModeTTL time.Duration,
	storageConfig Config,
) *ObservabilityRepository {
	repository := &ObservabilityRepository{
		pool:                     pool,
		logger:                   logger,
		entitlementCache:         sharedCache,
		requestLoggingModeTTL:    requestLoggingModeTTL,
		tokenUsageMetrics:        tokenUsageMetrics,
		tokenUsageEnabled:        storageConfig.TokenUsage.Enabled,
		tokenUsageFailOpen:       storageConfig.TokenUsage.FailOpen,
		tokenUsageKeyPrefix:      storageConfig.TokenUsage.KeyPrefix,
		tokenUsageFlushInterval:  storageConfig.TokenUsage.FlushInterval,
		tokenUsageFlushBatchSize: storageConfig.TokenUsage.FlushBatchSize,
		tokenUsageDeltaTTL:       storageConfig.TokenUsage.DeltaTTL,
		lastUsedFlushInterval:    storageConfig.TokenUsage.LastUsedInterval,
		lastUsedTTL:              storageConfig.TokenUsage.LastUsedTTL,
		lastUsedFallback:         make(map[string]time.Time),
		config:                   storageConfig,
	}

	if tokenUsageCache, ok := sharedCache.(cache.TokenUsageCache); ok {
		repository.tokenUsageCache = tokenUsageCache
	}

	return repository
}

func newPlatformCredentialRepository(pool *pgxpool.Pool, logger *zap.Logger) *PlatformCredentialRepository {
	return &PlatformCredentialRepository{pool: pool, logger: logger}
}

func newProviderResourceRepository(pool *pgxpool.Pool) *ProviderResourceRepository {
	return &ProviderResourceRepository{pool: pool}
}

func newRuntimeProfileRepository(
	pool *pgxpool.Pool,
	auth *AuthRepository,
	invalidator RuntimeProfileInvalidator,
) *RuntimeProfileRepository {
	return &RuntimeProfileRepository{
		pool:                      pool,
		auth:                      auth,
		runtimeProfileInvalidator: invalidator,
	}
}

func newPlatformServiceAccountRepository(pool *pgxpool.Pool) *PlatformServiceAccountRepository {
	return &PlatformServiceAccountRepository{pool: pool}
}

func newDatabaseHealthRepository(pool *pgxpool.Pool) *DatabaseHealthRepository {
	return &DatabaseHealthRepository{pool: pool}
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

func (r *IAMRepository) invalidateTenantRuntimeProfile(
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

func (r *AuthRepository) invalidateTenantRuntimeProfile(
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

func (r *AuthRepository) ListVirtualModelsForPrincipal(
	ctx context.Context,
	tenantID string,
	userID string,
	teamID string,
) ([]platform.VirtualModel, error) {
	return r.listVirtualModelsForPrincipalFunc(ctx, tenantID, userID, teamID)
}

func (r *AuthRepository) queueKeyLastUsed(ctx context.Context, keyID string, usedAt time.Time) error {
	return r.queueKeyLastUsedFunc(ctx, keyID, usedAt)
}

func (r *TenantAdminRepository) invalidateAuthzCacheForUser(tenantID string, userID string) {
	r.auth.invalidateAuthzCacheForUser(tenantID, userID)
}

func (r *TenantAdminRepository) bumpAuthzTenantEpoch(tenantID string) {
	r.auth.bumpAuthzTenantEpoch(tenantID)
}

func (r *PolicyRepository) invalidateTenantRuntimeProfile(
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

func (r *ModelRepository) invalidateTenantRuntimeProfile(
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
