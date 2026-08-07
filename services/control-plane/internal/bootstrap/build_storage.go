package bootstrap

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/platformcredentials"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	"github.com/Delpak/shahrah/services/control-plane/internal/runtimeprofile"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenantcredentials"
	tunneltransport "github.com/Delpak/shahrah/services/control-plane/internal/tunnel"
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
		b.cfg.TunnelSettings.ClientURL,
		time.Duration(b.cfg.TunnelSettings.Timeout)*time.Second,
	)
	b.tenantProviderProber = platformcredentials.NewTunnelCredentialProber(b.tunnel)

	tenantCredentialStore, credentialErr := tenantcredentials.NewEncryptedStoreFromEnv()
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
	if err := b.lifecycle.AddResource("storage", func(context.Context) error {
		store.Close()

		return nil
	}); err != nil {
		return err
	}

	credentialRuntime, credentialErr := newPlatformCredentialRuntime(
		b.ctx,
		store,
		sharedCache,
		b.httpMetrics,
		b.tenantProviderProber,
		b.logger,
	)
	if credentialErr != nil {
		return credentialErr
	}

	b.credentialRuntime = credentialRuntime

	if err := b.registerPlatformCredentialWorkers(); err != nil {
		return err
	}

	runtimeProfileCfg := runtimeprofile.LoadConfigFromEnv()

	var distributedCache runtimeprofile.DistributedCache
	if candidate, ok := sharedCache.(runtimeprofile.DistributedCache); ok {
		distributedCache = candidate
	} else {
		b.logger.Warn("Tenant runtime profile L2 disabled because cache does not support distributed operations")
	}

	b.runtimeProfiles = runtimeprofile.NewManager(
		runtimeprofile.NewCompiler(store),
		store,
		distributedCache,
		b.logger,
		runtimeProfileCfg,
	)
	b.runtimeDeploymentStore = runtimeprofile.NewDeploymentStore(b.runtimeProfiles, store)

	applicationRouter, routerErr := router.NewRouter(nil, b.cfg.RouterSettings, sharedCache, b.runtimeDeploymentStore, b.logger)
	if routerErr != nil {
		return fmt.Errorf("initialize router: %w", routerErr)
	}

	b.router = applicationRouter

	b.router.SetPlatformCredentialResolver(
		b.credentialRuntime.manager,
		b.credentialRuntime.mode,
		strings.TrimSpace(os.Getenv("PLATFORM_PROVIDER_CREDENTIAL_ENVIRONMENT")),
	)

	b.authMiddleware = auth.NewAuthMiddlewareWithRuntimeProfiles(
		store,
		b.cfg.GeneralSettings.MasterKey,
		b.logger,
		b.runtimeProfiles,
	)
	b.tenantMiddleware = tenant.NewMiddlewareWithCache(store, sharedCache, tenant.MiddlewareConfig{
		AllowHeaderHint: b.cfg.TenantSettings.AllowHeaderHint,
		CacheTTL:        5 * time.Minute,
		RuntimeProfiles: b.runtimeProfiles,
	})
	b.billing = billing.NewServiceWithRepositoriesAndRuntimeProfiles(
		sharedCache,
		store.WalletRepository,
		store.PlanRepository,
		b.runtimeProfiles,
		b.cfg.BillingSettings.DefaultUSDtoIRR,
		b.cfg.BillingSettings.RateSource,
	)

	anthropicPricing, pricingErr := b.configureProviderPricingSnapshots()
	b.anthropicPricing = anthropicPricing

	return pricingErr
}

func (b *builder) buildCache() (cache.Cache, error) {
	var (
		sharedCache cache.Cache
		err         error
	)

	if sentinelAddrs := splitCSVEnv("REDIS_SENTINEL_ADDRS"); len(sentinelAddrs) > 0 {
		sharedCache, err = cache.NewDualCacheWithSentinel(
			sentinelAddrs,
			stringEnvOrDefault("REDIS_MASTER_NAME", "mymaster"),
			b.cfg.RouterSettings.RedisPassword,
			os.Getenv("REDIS_SENTINEL_PASSWORD"),
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
