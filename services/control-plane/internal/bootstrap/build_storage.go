package bootstrap

import (
	"context"
	"fmt"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/config"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/Delpak/shahrah/services/control-plane/internal/iam"
	"github.com/Delpak/shahrah/services/control-plane/internal/platformcredentials"
	"github.com/Delpak/shahrah/services/control-plane/internal/policies"
	"github.com/Delpak/shahrah/services/control-plane/internal/prompts"
	"github.com/Delpak/shahrah/services/control-plane/internal/provideraccounts"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	"github.com/Delpak/shahrah/services/control-plane/internal/runtimeprofile"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenantadmin"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenantcredentials"
	tunneltransport "github.com/Delpak/shahrah/services/control-plane/internal/tunnel"
	"github.com/Delpak/shahrah/services/control-plane/internal/virtualmodels"
	"go.uber.org/zap"
)

func (b *builder) buildStorage() error {
	sharedCache, cacheErr := b.buildCache()
	if cacheErr != nil {
		return cacheErr
	}

	b.cache = sharedCache
	if err := b.lifecycle.AddResource("cache", func(context.Context) error {
		return sharedCache.Close()
	}); err != nil {
		return err
	}

	b.tunnel = tunneltransport.NewHTTPTransport(
		b.cfg.ProviderEgress.URL,
		time.Duration(b.cfg.ProviderEgress.Timeout)*time.Second,
	)
	b.tenantProviderProber = platformcredentials.NewTunnelCredentialProber(b.tunnel)

	tenantCredentialStore, credentialErr := tenantcredentials.NewEncryptedStore(
		b.cfg.Credentials.TenantActiveVersion,
		b.cfg.Credentials.TenantKEKsJSON,
		b.cfg.Credentials.TenantFingerprintKey,
	)
	if credentialErr != nil {
		b.logger.Warn(
			"Tenant provider credential encryption is unavailable; new tenant API keys will be rejected",
			zap.Error(credentialErr),
		)
	}

	store, storeErr := storage.NewStore(
		b.ctx,
		b.cfg.GeneralSettings.DatabaseURL,
		b.logger,
		storage.StoreDependencies{
			Config:                storageConfig(b.cfg.Storage),
			Cache:                 sharedCache,
			RequestLoggingModeTTL: time.Duration(b.cfg.RequestLogs.ModeCacheTTLSeconds) * time.Second,
			RuntimeProfileInvalidator: storage.RuntimeProfileInvalidatorFunc(
				func(ctx context.Context, tenantID string, reason string) {
					if b.runtimeProfiles != nil {
						b.runtimeProfiles.Invalidate(ctx, tenantID, reason)
					}
				},
			),
			TenantCredentialStore: tenantCredentialStore,
			TenantProviderProber:  b.tenantProviderProber,
			TokenUsageMetrics:     b.httpMetrics,
		},
	)
	if storeErr != nil {
		return fmt.Errorf("initialize storage: %w", storeErr)
	}

	b.store = store

	modelApplication, modelErr := virtualmodels.NewService(store.ModelRepository)
	if modelErr != nil {
		return fmt.Errorf("initialize model application: %w", modelErr)
	}

	b.modelApplication = modelApplication

	providerAccountApplication, err := provideraccounts.NewService(store.ProviderAccountRepository)
	if err != nil {
		return fmt.Errorf("initialize provider-account application service: %w", err)
	}

	b.providerAccountApplication = providerAccountApplication

	iamApplication, iamErr := iam.NewService(store.IAMRepository, store.AuthRepository)
	if iamErr != nil {
		return fmt.Errorf("initialize IAM application: %w", iamErr)
	}

	b.iamApplication = iamApplication

	tenantAdminApplication, tenantAdminErr := tenantadmin.NewService(store.TenantAdminRepository)
	if tenantAdminErr != nil {
		return fmt.Errorf("initialize tenant-admin application: %w", tenantAdminErr)
	}

	b.tenantAdminApplication = tenantAdminApplication

	policyApplication, policyErr := policies.NewService(store.PolicyRepository)
	if policyErr != nil {
		return fmt.Errorf("initialize policy application: %w", policyErr)
	}

	b.policyApplication = policyApplication

	promptApplication, promptErr := prompts.NewService(store.PromptRepository)
	if promptErr != nil {
		return fmt.Errorf("initialize prompt application: %w", promptErr)
	}

	b.promptApplication = promptApplication
	if err := b.lifecycle.AddResource("storage", func(context.Context) error {
		store.Close()

		return nil
	}); err != nil {
		return err
	}

	credentialRuntime, credentialErr := newPlatformCredentialRuntime(
		b.ctx,
		store.PlatformCredentialRepository,
		sharedCache,
		b.httpMetrics,
		b.tenantProviderProber,
		b.logger,
		b.cfg.Credentials,
		b.cfg.Admin.Discovery,
	)
	if credentialErr != nil {
		return credentialErr
	}

	b.credentialRuntime = credentialRuntime

	if err := b.registerPlatformCredentialWorkers(); err != nil {
		return err
	}

	runtimeProfileCfg := runtimeprofile.Config{
		VersionCheckInterval: b.cfg.Runtime.RuntimeProfileVersionCheck,
		StaleGrace:           b.cfg.Runtime.RuntimeProfileStaleGrace,
		L2TTL:                b.cfg.Runtime.RuntimeProfileL2TTL,
	}

	var distributedCache runtimeprofile.DistributedCache
	if candidate, ok := sharedCache.(runtimeprofile.DistributedCache); ok {
		distributedCache = candidate
	} else {
		b.logger.Warn("Tenant runtime profile L2 disabled because cache does not support distributed operations")
	}

	b.runtimeProfiles = runtimeprofile.NewManager(
		runtimeprofile.NewCompiler(newRuntimeProfileSources(store, b.modelApplication, b.policyApplication)),
		store.RuntimeProfileRepository,
		distributedCache,
		b.logger,
		runtimeProfileCfg,
	)
	b.runtimeDeploymentStore = runtimeprofile.NewDeploymentStore(
		b.runtimeProfiles,
		store.ProviderAccountRepository,
	)

	applicationRouter, routerErr := router.NewRouterWithDependencies(
		nil,
		b.cfg.RouterSettings,
		sharedCache,
		router.Dependencies{
			DeploymentStore:                   b.runtimeDeploymentStore,
			DeploymentSnapshots:               b.runtimeDeploymentStore,
			ProviderAccountAuthResolver:       b.runtimeDeploymentStore,
			ProviderAccountCredentialPolicy:   router.ProviderAccountCredentialsFailOpen,
			RequireProviderAccountCredentials: true,
			PlatformCredentials:               b.credentialRuntime.manager,
			PlatformCredentialMode:            b.credentialRuntime.mode,
			PlatformCredentialEnvironment:     b.cfg.Runtime.PlatformCredentialEnv,
			DeploymentCacheTTL:                &b.cfg.Runtime.RouterDeploymentCacheTTL,
		},
		b.logger,
	)
	if routerErr != nil {
		return fmt.Errorf("initialize router: %w", routerErr)
	}

	b.router = applicationRouter

	if err := b.buildOIDC(); err != nil {
		return err
	}

	b.authMiddleware = auth.NewAuthMiddlewareWithOptions(
		b.iamApplication,
		b.cfg.GeneralSettings.MasterKey,
		b.logger,
		auth.AuthMiddlewareOptions{
			RuntimeProfiles: b.runtimeProfiles,
			AccessTokens:    b.oidcAccessTokens(),
		},
	)
	b.tenantMiddleware = tenant.NewMiddlewareWithCache(store.TenantRepository, sharedCache, tenant.MiddlewareConfig{
		AllowHeaderHint: b.cfg.TenantSettings.AllowHeaderHint,
		CacheTTL:        5 * time.Minute,
		RuntimeProfiles: b.runtimeProfiles,
	})

	// Credit enforcement prices reservations with the published rate card.
	b.buildCreditRateCards()

	b.billing = billing.NewServiceWithDependencies(
		billing.Dependencies{
			Cache:           sharedCache,
			BudgetWindows:   sharedCache,
			Wallets:         store.WalletRepository,
			Entitlements:    store.PlanRepository,
			RuntimeProfiles: b.runtimeProfiles,
			CreditPools:     store.CreditLedgerRepository,
			CreditRateCards: b.creditRateCards,
			Logger:          b.logger,
			Config: &billing.ServiceConfig{
				ExchangeRateCache: billing.ExchangeRateCacheConfig{
					FreshTTL: b.cfg.Runtime.BillingExchangeRateTTL,
					MaxStale: b.cfg.Runtime.BillingExchangeRateStale,
				},
				PricingCache: billing.PricingCacheConfig{
					Enabled: b.cfg.Runtime.BillingPricingCacheEnabled,
					TTL:     b.cfg.Runtime.BillingPricingCacheTTL,
				},
				WalletCacheTTL:    b.cfg.Runtime.BillingWalletCacheTTL,
				BudgetWindowGrace: b.cfg.Storage.BudgetWindowTTLGrace,
				RateLimitFailOpen: b.cfg.Runtime.RateLimitPolicyFailOpen,
				CreditEnforcementEnabled: b.cfg.Runtime.BillingCreditLedger.Enabled &&
					b.cfg.Runtime.BillingCreditLedger.EnforcementEnabled,
			},
		},
		b.cfg.BillingSettings.DefaultUSDtoIRR,
		b.cfg.BillingSettings.RateSource,
	)
	if err := b.lifecycle.AddResource("billing", func(ctx context.Context) error {
		return b.billing.Shutdown(ctx)
	}); err != nil {
		return err
	}

	anthropicPricing, pricingErr := b.configureProviderPricingSnapshots()
	b.anthropicPricing = anthropicPricing

	return pricingErr
}

type controlPlaneCache interface {
	cache.Cache
	billing.BudgetWindowCache
}

func (b *builder) buildCache() (controlPlaneCache, error) {
	var (
		sharedCache controlPlaneCache
		err         error
	)

	if sentinelAddrs := b.cfg.Runtime.RedisSentinelAddresses; len(sentinelAddrs) > 0 {
		sharedCache, err = cache.NewDualCacheWithSentinel(
			sentinelAddrs,
			b.cfg.Runtime.RedisMasterName,
			b.cfg.RouterSettings.RedisPassword,
			b.cfg.Runtime.RedisSentinelPassword,
			b.logger,
		)
	} else if b.cfg.RouterSettings.RedisHost != "" {
		sharedCache, err = cache.NewDualCache(
			b.cfg.RouterSettings.RedisHost,
			b.cfg.RouterSettings.RedisPort,
			b.cfg.RouterSettings.RedisPassword,
			b.logger,
		)
	} else {
		return cache.NewMemoryCache(b.logger)
	}

	if err == nil {
		return sharedCache, nil
	}

	b.logger.Warn("Failed to initialize shared Redis cache, falling back to memory-only", zap.Error(err))

	return cache.NewMemoryCache(b.logger)
}

func storageConfig(settings config.StorageSettings) *storage.Config {
	return &storage.Config{
		Postgres: storage.PostgresConfig{
			MaxConnections:    settings.PostgresMaxConnections,
			MinConnections:    settings.PostgresMinConnections,
			MaxLifetime:       settings.PostgresMaxLifetime,
			MaxIdleTime:       settings.PostgresMaxIdleTime,
			HealthCheckPeriod: settings.PostgresHealthCheckPeriod,
		},
		ClickHouse: storage.ClickHouseConfig{
			DSN:        settings.ClickHouseDSN,
			Required:   settings.ClickHouseRequired,
			Attempts:   settings.ClickHouseAttempts,
			RetryDelay: settings.ClickHouseRetryDelay,
		},
		TokenUsage: storage.TokenUsageConfig{
			Enabled:          settings.TokenUsageEnabled,
			FailOpen:         settings.TokenUsageFailOpen,
			KeyPrefix:        settings.TokenUsageKeyPrefix,
			FlushInterval:    settings.TokenUsageFlushInterval,
			FlushBatchSize:   settings.TokenUsageFlushBatchSize,
			DeltaTTL:         settings.TokenUsageDeltaTTL,
			LastUsedInterval: settings.LastUsedFlushInterval,
			LastUsedTTL:      settings.LastUsedTTL,
		},
		HotPath: storage.HotPathConfig{
			RequestLoggingModeTTL:        settings.RequestLoggingModeTTL,
			ModelCacheTTL:                settings.ModelCacheTTL,
			ModelCacheStaleTTL:           settings.ModelCacheStaleTTL,
			DatabaseTimeout:              settings.GatewayDatabaseTimeout,
			SubscriptionLifecycleTimeout: settings.SubscriptionDatabaseTimeout,
			AuthorizationCacheTTL:        settings.AuthorizationCacheTTL,
			AuthorizationCacheMaxEntries: settings.AuthorizationCacheMaxEntries,
		},
		Plans: storage.PlanRepositoryConfig{
			EntitlementCacheTTL: settings.EntitlementCacheTTL,
		},
		Wallets: storage.WalletRepositoryConfig{
			BudgetWindowTTLGrace:   settings.BudgetWindowTTLGrace,
			CreditOverageCeilingMC: settings.CreditOverageCeilingCredits * credits.MillicreditsPerCredit,
		},
		ProviderAccount: storage.ProviderAccountRepositoryConfig{
			CredentialCacheTTL: settings.TenantCredentialCacheTTL,
		},
	}
}
