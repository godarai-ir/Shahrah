package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/admin"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/config"
	"github.com/Delpak/shahrah/services/control-plane/internal/guardrails"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/logretention"
	metricsobs "github.com/Delpak/shahrah/services/control-plane/internal/observability/metrics"
	prommetrics "github.com/Delpak/shahrah/services/control-plane/internal/observability/prometheus"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/Delpak/shahrah/services/control-plane/internal/platformcredentials"
	"github.com/Delpak/shahrah/services/control-plane/internal/providerregistry"
	"github.com/Delpak/shahrah/services/control-plane/internal/proxy"
	"github.com/Delpak/shahrah/services/control-plane/internal/responsecache"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	"github.com/Delpak/shahrah/services/control-plane/internal/runtimeprofile"
	"github.com/Delpak/shahrah/services/control-plane/internal/server"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/Delpak/shahrah/services/control-plane/internal/supportfiles"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenantcredentials"
	tunneltransport "github.com/Delpak/shahrah/services/control-plane/internal/tunnel"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer func() {
		if syncErr := logger.Sync(); syncErr != nil {
			log.Printf("logger sync failed: %v", syncErr)
		}
	}()

	cfg, err := config.LoadConfig(logger)
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	configureRuntimeProfiling(cfg, logger)

	if strings.TrimSpace(cfg.GeneralSettings.DatabaseURL) == "" {
		logger.Fatal("Missing required database configuration", zap.String("field", "general_settings.database_url"))
	}

	ctx := context.Background()

	tenantCredentialStore, tenantCredentialErr :=
		tenantcredentials.NewEncryptedStoreFromEnv()
	if tenantCredentialErr != nil {
		logger.Warn(
			"Tenant provider credential encryption is unavailable; new tenant API keys will be rejected",
			zap.Error(tenantCredentialErr),
		)
	}

	httpMetrics := prommetrics.NewHTTPMetrics()

	var dualCache cache.Cache
	if sentinelAddrs := splitCSVEnv("REDIS_SENTINEL_ADDRS"); len(sentinelAddrs) > 0 {
		dualCache, err = cache.NewDualCacheWithSentinel(
			sentinelAddrs,
			stringEnvOrDefault("REDIS_MASTER_NAME", "mymaster"),
			cfg.RouterSettings.RedisPassword,
			os.Getenv("REDIS_SENTINEL_PASSWORD"),
			logger,
		)
		if err != nil {
			logger.Warn("Failed to initialize Redis Sentinel cache, falling back to memory-only", zap.Error(err))

			dualCache, err = cache.NewMemoryCache(logger)
			if err != nil {
				logger.Fatal("Failed to initialize in-memory cache", zap.Error(err))
			}
		}
	} else if cfg.RouterSettings.RedisHost != "" {
		dualCache, err = cache.NewDualCache(cfg.RouterSettings.RedisHost, cfg.RouterSettings.RedisPort, cfg.RouterSettings.RedisPassword, logger)
		if err != nil {
			logger.Warn("Failed to initialize Redis cache, falling back to memory-only", zap.Error(err))

			dualCache, err = cache.NewMemoryCache(logger)
			if err != nil {
				logger.Fatal("Failed to initialize in-memory cache", zap.Error(err))
			}
		}
	} else {
		dualCache, err = cache.NewMemoryCache(logger)
		if err != nil {
			logger.Fatal("Failed to initialize in-memory cache", zap.Error(err))
		}
	}
	defer func() {
		if closeErr := dualCache.Close(); closeErr != nil {
			logger.Warn("Failed to close cache", zap.Error(closeErr))
		}
	}()

	tunnelClient := tunneltransport.NewHTTPTransport(
		cfg.TunnelSettings.ClientURL,
		time.Duration(cfg.TunnelSettings.Timeout)*time.Second,
	)
	tenantProviderProber := platformcredentials.NewTunnelCredentialProber(tunnelClient)

	var runtimeProfiles *runtimeprofile.Manager

	store, err := storage.NewStore(ctx, cfg.GeneralSettings.DatabaseURL, logger, storage.StoreDependencies{
		Cache:                 dualCache,
		RequestLoggingModeTTL: time.Duration(cfg.RequestLogs.ModeCacheTTLSeconds) * time.Second,
		RuntimeProfileInvalidator: storage.RuntimeProfileInvalidatorFunc(
			func(invalidationCtx context.Context, tenantID string, reason string) {
				if runtimeProfiles != nil {
					runtimeProfiles.Invalidate(invalidationCtx, tenantID, reason)
				}
			},
		),
		TenantCredentialStore: tenantCredentialStore,
		TenantProviderProber:  tenantProviderProber,
		TokenUsageMetrics:     httpMetrics,
	})
	if err != nil {
		logger.Fatal("Failed to initialize storage", zap.Error(err))
	}
	defer store.Close()

	_, sharedResponseCacheAvailable := dualCache.(*cache.DualCache)

	credentialRuntime := initializePlatformCredentialRuntime(ctx, store, dualCache, httpMetrics, logger)
	defer credentialRuntime.cancel()

	credentialMode := credentialRuntime.mode
	platformCredentialManager := credentialRuntime.manager
	platformCredentialService := credentialRuntime.service
	providerSchemaCatalog := credentialRuntime.schemas

	runtimeProfileCfg := runtimeprofile.LoadConfigFromEnv()

	var distributedRuntimeProfileCache runtimeprofile.DistributedCache
	if dc, ok := dualCache.(runtimeprofile.DistributedCache); ok {
		distributedRuntimeProfileCache = dc
	} else {
		logger.Warn("Tenant runtime profile L2 disabled because cache does not support distributed operations")
	}

	runtimeProfiles = runtimeprofile.NewManager(runtimeprofile.NewCompiler(store), store, distributedRuntimeProfileCache, logger, runtimeProfileCfg)
	runtimeDeploymentStore := runtimeprofile.NewDeploymentStore(runtimeProfiles, store)

	llmRouter, err := router.NewRouter(nil, cfg.RouterSettings, dualCache, runtimeDeploymentStore, logger)
	if err != nil {
		logger.Fatal("Failed to initialize router", zap.Error(err))
	}

	llmRouter.SetPlatformCredentialResolver(platformCredentialManager, credentialMode, strings.TrimSpace(os.Getenv("PLATFORM_PROVIDER_CREDENTIAL_ENVIRONMENT")))

	authMiddleware := auth.NewAuthMiddleware(store, cfg.GeneralSettings.MasterKey, logger)
	authMiddleware.SetRuntimeProfiles(runtimeProfiles)

	tenantMiddleware := tenant.NewMiddlewareWithCache(store, dualCache, tenant.MiddlewareConfig{
		AllowHeaderHint: cfg.TenantSettings.AllowHeaderHint,
		CacheTTL:        5 * time.Minute,
	})
	tenantMiddleware.SetRuntimeProfiles(runtimeProfiles)

	billingService := billing.NewServiceWithRepositories(
		dualCache,
		store.WalletRepository,
		store.PlanRepository,
		cfg.BillingSettings.DefaultUSDtoIRR,
		cfg.BillingSettings.RateSource,
	)
	billingService.SetRuntimeProfiles(runtimeProfiles)

	pricingCtx, cancelPricing := context.WithCancel(ctx)
	defer cancelPricing()

	anthropicPricing := configureProviderPricingSnapshots(
		ctx,
		pricingCtx,
		billingService,
		logger,
	)

	subscriptionProcessor := billing.NewSubscriptionProcessor(store.PlanRepository, logger, time.Minute)
	nobitexExchangeRateProcessor := billing.NewNobitexExchangeRateProcessor(
		billing.LoadNobitexExchangeRateConfigFromEnv(cfg.BillingSettings.RateSource),
		store.WalletRepository,
		logger,
	)

	var walletSettlementProcessor *billing.WalletSettlementProcessor
	if liveBudgetCache, ok := dualCache.(cache.BudgetWindowInspector); ok {
		walletSettlementProcessor = billing.NewWalletSettlementProcessor(
			billing.LoadWalletSettlementConfigFromEnv(),
			liveBudgetCache,
			store.WalletRepository,
			logger,
		)
	}

	subscriptionProcessor.Start()
	defer subscriptionProcessor.Stop()

	nobitexExchangeRateProcessor.Start()
	defer nobitexExchangeRateProcessor.Stop()

	if walletSettlementProcessor != nil {
		walletSettlementProcessor.Start()
		defer walletSettlementProcessor.Stop()
	}

	if credentialRuntime.service != nil {
		credentialRuntime.service.SetProber(tenantProviderProber)
	}

	metricsCfg := metricsobs.LoadConfigFromEnv()

	var (
		emitMetric func(input metricsobs.BuildInput)
		reconciler *metricsobs.CostReconciler
	)

	if metricsCfg.Enabled {
		var metricsDLQ metricsobs.DeadLetterQueue

		if metricsCfg.DLQ.Backend == "redis_stream" {
			metricsDLQ, err = newMetricsDLQ(ctx, metricsCfg.DLQ)
			if err != nil {
				logger.Fatal("Failed to initialize model metrics Redis Stream DLQ", zap.Error(err))
			}

			defer func() {
				if closeErr := metricsDLQ.Close(); closeErr != nil {
					logger.Warn("Failed to close model metrics Redis Stream DLQ", zap.Error(closeErr))
				}
			}()
		} else {
			logger.Warn("Model metrics DLQ is disabled; failed ClickHouse exports can be dropped")
		}

		metricsQueue := metricsobs.NewQueue(metricsCfg.QueueCapacity)
		metricsProcessor := metricsobs.NewProcessor(metricsCfg, metricsQueue, store, logger, metricsDLQ)
		metricsProcessor.Start()

		defer func() {
			stopCtx, cancelStop := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancelStop()

			if stopErr := metricsProcessor.Stop(stopCtx); stopErr != nil {
				logger.Warn("Failed to stop metrics processor", zap.Error(stopErr))
			}
		}()

		emitMetric = func(input metricsobs.BuildInput) {
			event := metricsProcessor.BuildEvent(input)
			_ = metricsQueue.Enqueue(event)
		}

		reconciler = metricsobs.NewCostReconciler(metricsobs.LoadReconcileConfigFromEnv(), &costReconcileAdapter{store: store}, logger)

		reconciler.Start()
		defer reconciler.Stop()
	}

	supportFileService, supportFileErr := supportfiles.NewFromEnv()
	if supportFileErr != nil {
		logger.Warn("Support attachment service is unavailable", zap.Error(supportFileErr))
	}

	adminDeps := admin.Dependencies{
		Keys:             store,
		Models:           store,
		Wallets:          store.WalletRepository,
		ProviderAccounts: store.ProviderAccountRepository,
		ProviderSchemas:  providerSchemaCatalog,
		OrgUnits:         store,
		Memberships:      store,
		Users:            store,
		VirtualUsers:     store,
		Teams:            store,
		Onboarding:       store,
		LoginDiscovery:   store,
		Traces:           store,
		Logs:             store,
		Metrics:          store,
		TracePrivacy:     store,
		RequestLogging:   store,
		Prompts:          store,
		Roles:            store,
		Plans:            store.PlanRepository,
		Entitlements:     store.PlanRepository,
		SupportTenants:   store.SupportRepository,
		SupportTickets:   store.SupportRepository,
		SupportFiles:     supportFileService,
	}
	if platformCredentialService != nil {
		adminDeps.ProviderCredentials = platformCredentialService
	}

	adminHandlers := admin.New(adminDeps)
	authHandlers := admin.NewAuthHandlers(store, admin.NewTenantAdminOTPNotifierFromEnv())

	var requestLogger proxy.RequestLogger = store
	if cfg.RequestLogs.AsyncEnabled {
		asyncRequestLogger := proxy.NewAsyncRequestLogWriter(store, proxy.AsyncRequestLogConfig{
			QueueSize:      cfg.RequestLogs.QueueSize,
			Workers:        cfg.RequestLogs.Workers,
			BatchSize:      cfg.RequestLogs.BatchSize,
			FlushInterval:  time.Duration(cfg.RequestLogs.FlushIntervalMS) * time.Millisecond,
			EnqueueTimeout: time.Duration(cfg.RequestLogs.EnqueueTimeoutMS) * time.Millisecond,
			DropOnFull:     cfg.RequestLogs.DropOnFull,
		}, logger)
		requestLogger = asyncRequestLogger

		defer func() {
			timeout := time.Duration(cfg.RequestLogs.ShutdownDrainTimeoutSeconds) * time.Second
			if timeout <= 0 {
				timeout = 5 * time.Second
			}

			shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			if stopErr := asyncRequestLogger.Stop(shutdownCtx); stopErr != nil {
				logger.Warn("Failed to drain async request logs", zap.Error(stopErr))
			}
		}()
	}

	baseProxyDeps := proxy.BaseHandlerDependencies{
		Router:                   llmRouter,
		Tunnel:                   tunnelClient,
		Billing:                  billingService,
		Wallets:                  store,
		AnthropicRuntimeProfiles: anthropicPricing,
		ProviderResources:        store,
		Deployments:              runtimeDeploymentStore,
		RequestLogger:            requestLogger,
		Logger:                   logger,
		EmitMetric:               emitMetric,
	}

	var spendLogger proxy.SpendLogger = store
	if cfg.SpendLogs.AsyncEnabled {
		asyncSpendLogger := proxy.NewAsyncSpendLogWriter(store, proxy.AsyncSpendLogConfig{
			QueueSize:      cfg.SpendLogs.QueueSize,
			Workers:        cfg.SpendLogs.Workers,
			BatchSize:      cfg.SpendLogs.BatchSize,
			FlushInterval:  time.Duration(cfg.SpendLogs.FlushIntervalMS) * time.Millisecond,
			EnqueueTimeout: time.Duration(cfg.SpendLogs.EnqueueTimeoutMS) * time.Millisecond,
			DropOnFull:     cfg.SpendLogs.DropOnFull,
		}, logger, httpMetrics)
		spendLogger = asyncSpendLogger

		defer func() {
			timeout := time.Duration(cfg.SpendLogs.ShutdownDrainTimeoutSeconds) * time.Second
			if timeout <= 0 {
				timeout = 5 * time.Second
			}

			shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			if stopErr := asyncSpendLogger.Stop(shutdownCtx); stopErr != nil {
				logger.Warn("Failed to drain async spend logs", zap.Error(stopErr))
			}
		}()
	}

	baseProxyDeps.SpendLogger = spendLogger
	virtualModelStore := runtimeprofile.NewVirtualModelStore(runtimeProfiles)
	guardrailResolver := runtimeprofile.NewGuardrailResolver(runtimeProfiles)

	proxyHandler := proxy.NewChatCompletionHandlerWithDependencies(proxy.ChatCompletionDependencies{
		Base:              baseProxyDeps,
		VirtualModelStore: virtualModelStore,
		GuardrailResolver: guardrailResolver,
		PromptResolver:    store,
	})
	proxyHandler.SetGuardrails(guardrails.NewDefaultEngine())

	var responseEmbedder responsecache.Embedder

	if cfg.ResponseCache.Enabled && sharedResponseCacheAvailable && strings.TrimSpace(cfg.ResponseCache.EmbeddingBaseURL) != "" && strings.TrimSpace(cfg.ResponseCache.EmbeddingAPIKey) != "" {
		embedder, embedderErr := responsecache.NewArvanCloudBGEEmbedder(&http.Client{Timeout: time.Duration(cfg.ResponseCache.EmbeddingTimeoutSec) * time.Second}, responsecache.ArvanCloudConfig{
			BaseURL: cfg.ResponseCache.EmbeddingBaseURL, APIKey: cfg.ResponseCache.EmbeddingAPIKey,
			Model: cfg.ResponseCache.EmbeddingModel, Revision: cfg.ResponseCache.EmbeddingRevision,
			Dimensions: cfg.ResponseCache.EmbeddingDimensions, MaxBatchSize: cfg.ResponseCache.EmbeddingMaxBatchSize,
		})
		if embedderErr != nil {
			logger.Warn("Smart response cache embedder is unavailable", zap.Error(embedderErr))
		} else {
			responseEmbedder = embedder
		}
	} else if cfg.ResponseCache.Enabled && !sharedResponseCacheAvailable {
		logger.Warn("Smart response cache disabled because shared Redis is unavailable")
	}

	responseCache, responseCacheErr := responsecache.NewService(dualCache, responseEmbedder, responsecache.Config{
		Enabled: cfg.ResponseCache.Enabled, KeyPrefix: cfg.ResponseCache.KeyPrefix,
		DefaultTTL:            time.Duration(cfg.ResponseCache.DefaultTTLSeconds) * time.Second,
		MinTTL:                time.Duration(cfg.ResponseCache.MinTTLSeconds) * time.Second,
		MaxTTL:                time.Duration(cfg.ResponseCache.MaxTTLSeconds) * time.Second,
		MaxResponseBytes:      cfg.ResponseCache.MaxResponseBytes,
		MaxSemanticCandidates: cfg.ResponseCache.MaxSemanticCandidates,
	})
	if responseCacheErr != nil {
		logger.Fatal("Failed to initialize response cache", zap.Error(responseCacheErr))
	}

	proxyHandler.SetResponseCache(responseCache)

	completionsHandler := proxy.NewCompletionsHandlerWithDependencies(baseProxyDeps)
	responsesHandler := proxy.NewResponsesHandlerWithDependencies(baseProxyDeps)
	embeddingsHandler := proxy.NewEmbeddingsHandlerWithDependencies(baseProxyDeps)
	moderationsHandler := proxy.NewModerationsHandlerWithDependencies(baseProxyDeps)
	rerankHandler := proxy.NewRerankHandlerWithDependencies(baseProxyDeps)
	imagesHandler := proxy.NewImagesHandlerWithDependencies(baseProxyDeps)
	audioTranscriptionsHandler := proxy.NewAudioTranscriptionsHandlerWithDependencies(baseProxyDeps)
	audioSpeechHandler := proxy.NewAudioSpeechHandlerWithDependencies(baseProxyDeps)
	mistralOCRHandler := proxy.NewMistralOCRHandlerWithDependencies(baseProxyDeps)
	mistralNativeHandler := proxy.NewMistralNativeHandler(baseProxyDeps, store)
	modelsHandler := proxy.NewModelsHandler(runtimeDeploymentStore)
	anthropicNativeHandler := proxy.NewAnthropicNativeHandler(baseProxyDeps, store)
	geminiNativeHandler := proxy.NewGeminiNativeHandler(baseProxyDeps, store)

	traceCfg := tracing.LoadConfigFromEnv()
	traceQueue := tracing.NewQueue(traceCfg.QueueCapacity)
	traceSampler := tracing.NewSampler(traceCfg.DefaultSampleRate, traceCfg.SlowRequest.Milliseconds())
	tracePrivacy := tracing.NewPrivacyManager(traceCfg)
	tracePrivacy.SetTenantPolicyResolver(newTenantTracePolicyResolver(store))

	reliableExporter, err := tracing.NewReliableExporter(traceCfg, store, logger)
	if err != nil {
		logger.Fatal("Failed to initialize tracing exporter", zap.Error(err))
	}
	defer func() {
		if closeErr := reliableExporter.Close(); closeErr != nil {
			logger.Warn("Failed to close tracing exporter", zap.Error(closeErr))
		}
	}()

	traceProcessor := tracing.NewProcessor(traceCfg, traceQueue, reliableExporter, logger)
	traceProcessor.Start()

	defer func() {
		stopCtx, cancelStop := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancelStop()

		if stopErr := traceProcessor.Stop(stopCtx); stopErr != nil {
			logger.Warn("Failed to stop trace processor", zap.Error(stopErr))
		}
	}()

	logRetentionCleaner := logretention.NewRetentionCleaner(logretention.LoadConfigFromEnv(), store, logger)

	logRetentionCleaner.Start()
	defer logRetentionCleaner.Stop()

	app := server.NewFiberApp(cfg, server.Dependencies{
		AuthMiddleware:             authMiddleware,
		TenantResolver:             tenantMiddleware,
		TracingMiddleware:          tracing.Middleware(traceQueue, traceSampler, traceCfg.SlowRequest.Milliseconds(), tracePrivacy),
		TraceHealth:                reliableExporter,
		ClickHouseHealth:           store,
		AdminHandlers:              adminHandlers,
		AuthHandlers:               authHandlers,
		ProxyHandler:               proxyHandler,
		CompletionsHandler:         completionsHandler,
		ResponsesHandler:           responsesHandler,
		EmbeddingsHandler:          embeddingsHandler,
		ModerationsHandler:         moderationsHandler,
		RerankHandler:              rerankHandler,
		ImagesHandler:              imagesHandler,
		AudioTranscriptionsHandler: audioTranscriptionsHandler,
		AudioSpeechHandler:         audioSpeechHandler,
		MistralOCRHandler:          mistralOCRHandler,
		MistralNativeHandler:       mistralNativeHandler,
		ModelsHandler:              modelsHandler,
		AnthropicNativeHandler:     anthropicNativeHandler,
		GeminiNativeHandler:        geminiNativeHandler,
		PoolStats:                  store,
		HTTPMetrics:                httpMetrics,
		Logger:                     logger,
		PlatformServiceAccounts:    store,
		ExchangeRates:              store,
	})

	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go anthropicNativeHandler.RunBatchWorker(runCtx, 15*time.Second)
	go mistralNativeHandler.RunBatchWorker(runCtx, 15*time.Second)

	store.StartTokenUsageFlusher(runCtx)
	store.StartUsageCounterFlusher(runCtx)

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		select {
		case <-sigCh:
			logger.Info("Shutting down gracefully...")

			if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
				logger.Error("Server shutdown failed", zap.Error(err))
			}

			cancel()
		case <-runCtx.Done():
			return
		}
	}()

	addr := fmt.Sprintf("%s:%d", cfg.GeneralSettings.Host, cfg.GeneralSettings.Port)
	logger.Info("Starting Shahrah control-plane", zap.String("addr", addr), zap.String("version", "0.2.0"))

	if err := app.Listen(addr); err != nil {
		logger.Fatal("Server failed to start", zap.Error(err))
	}
}

func configureProviderPricingSnapshots(
	ctx context.Context,
	runCtx context.Context,
	billingService *billing.Service,
	logger *zap.Logger,
) *billing.AnthropicPricingSnapshotManager {
	var anthropicPricing *billing.AnthropicPricingSnapshotManager

	for _, provider := range providerregistry.PricingProviders() {
		switch provider.PricingSnapshot {
		case providerregistry.PricingSnapshotAnthropic:
			manager, managerErr := billing.NewAnthropicPricingSnapshotManagerFromEnv(nil)
			if managerErr != nil {
				logger.Warn(
					provider.DisplayName+" pricing is unavailable; requests will fail closed until configuration is restored",
					zap.Error(managerErr),
				)

				continue
			}

			billingService.SetAnthropicPricingSnapshotManager(manager)
			anthropicPricing = manager

			reloadErr := manager.Reload(ctx)
			if reloadErr != nil {
				logger.Warn(
					"Initial "+provider.DisplayName+" pricing snapshot is unavailable; background refresh will retry",
					zap.Error(reloadErr),
				)
			}

			go manager.Run(runCtx, 5*time.Minute)
		case providerregistry.PricingSnapshotCatalog:
			manager, managerErr := billing.NewCatalogPricingSnapshotManagerFromEnv(
				provider.Slug,
				nil,
			)
			if managerErr != nil {
				logger.Warn(
					provider.DisplayName+" pricing is unavailable; requests will fail closed until configuration is restored",
					zap.Error(managerErr),
				)

				continue
			}

			billingService.SetCatalogPricingSnapshotManager(provider.Slug, manager)

			reloadErr := manager.Reload(ctx)
			if reloadErr != nil {
				logger.Warn(
					"Initial "+provider.DisplayName+" pricing snapshot is unavailable; background refresh will retry",
					zap.Error(reloadErr),
				)
			}

			go manager.Run(runCtx, 5*time.Minute)
		}
	}

	return anthropicPricing
}

func splitCSVEnv(key string) []string {
	parts := strings.Split(os.Getenv(key), ",")

	values := make([]string, 0, len(parts))
	for _, part := range parts {
		if value := strings.TrimSpace(part); value != "" {
			values = append(values, value)
		}
	}

	return values
}

func stringEnvOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}

	return fallback
}

func newMetricsDLQ(ctx context.Context, cfg metricsobs.DLQConfig) (metricsobs.DeadLetterQueue, error) {
	dlqCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	return metricsobs.NewRedisStreamDLQ(dlqCtx, cfg)
}

type platformCredentialRuntime struct {
	mode    string
	manager *platformcredentials.Manager
	service *platformcredentials.Service
	schemas platformcredentials.SchemaCatalog
	cancel  context.CancelFunc
}

func initializePlatformCredentialRuntime(
	ctx context.Context,
	store *storage.Store,
	sharedCache cache.Cache,
	metrics platformcredentials.Metrics,
	logger *zap.Logger,
) platformCredentialRuntime {
	mode := strings.TrimSpace(os.Getenv("PLATFORM_PROVIDER_CREDENTIAL_MODE"))
	if mode == "" {
		mode = "env"
	}

	credentialRuntime := platformCredentialRuntime{mode: mode, cancel: func() {}}
	schemas, schemaErr := platformcredentials.NewModelDiscoverySchemaCatalogFromEnv(nil)
	credentialRuntime.schemas = schemas

	if schemaErr != nil {
		logger.Warn("Provider connection schema catalog is unavailable", zap.Error(schemaErr))
	}

	credentialsConfigured := mode != "env" || strings.TrimSpace(os.Getenv("PLATFORM_PROVIDER_CREDENTIAL_KEKS_JSON")) != ""
	if !credentialsConfigured {
		return credentialRuntime
	}

	keks, kekErr := platformcredentials.NewEnvironmentKEKProviderFromEnv()

	fingerprintKey, fingerprintErr := platformcredentials.FingerprintKeyFromEnv()
	if kekErr != nil || fingerprintErr != nil {
		if mode == "platform_store" {
			logger.Fatal("Platform provider credential bootstrap keys are invalid")
		}

		logger.Warn("Platform provider credential store is unavailable; environment break-glass remains active")

		return credentialRuntime
	}

	secretStore, err := platformcredentials.NewEncryptedDBSecretStore(keks, fingerprintKey)
	if err != nil {
		logger.Fatal("Failed to initialize platform credential encryption")
	}

	credentialRuntime.manager = platformcredentials.NewManager(store, secretStore, metrics)
	if err := credentialRuntime.manager.Reload(ctx); err != nil {
		if mode == "platform_store" {
			logger.Fatal("Failed to pre-warm platform provider credential snapshot", zap.Error(err))
		}

		logger.Warn("Initial platform credential snapshot load failed; environment break-glass remains active", zap.Error(err))
	}

	var bus *platformcredentials.RedisInvalidationBus
	if redisCache, ok := sharedCache.(*cache.DualCache); ok {
		bus = platformcredentials.NewRedisInvalidationBus(redisCache)
	}

	credentialRuntime.service = platformcredentials.NewService(store, secretStore, credentialRuntime.manager, bus, credentialRuntime.schemas, nil)

	runCtx, cancel := context.WithCancel(ctx)
	credentialRuntime.cancel = cancel

	go credentialRuntime.manager.Reconcile(runCtx, 30*time.Second)

	if bus != nil {
		go func() {
			if err := bus.Consume(runCtx, credentialRuntime.manager); err != nil && !errors.Is(err, context.Canceled) {
				logger.Warn("Platform credential invalidation consumer stopped", zap.Error(err))
			}
		}()
	}

	return credentialRuntime
}

func configureRuntimeProfiling(cfg *config.Config, logger *zap.Logger) {
	if cfg == nil {
		return
	}

	if cfg.Profiling.MutexProfileFraction > 0 {
		runtime.SetMutexProfileFraction(cfg.Profiling.MutexProfileFraction)
		logger.Info("Enabled mutex profiling", zap.Int("fraction", cfg.Profiling.MutexProfileFraction))
	}

	if cfg.Profiling.BlockProfileRate > 0 {
		runtime.SetBlockProfileRate(cfg.Profiling.BlockProfileRate)
		logger.Info("Enabled block profiling", zap.Int("rate", cfg.Profiling.BlockProfileRate))
	}
}

type tenantTracePolicyResolver struct {
	store *storage.Store
	mu    sync.RWMutex
	cache map[string]tenantPolicyCache
	ttl   time.Duration
}

type tenantPolicyCache struct {
	policy    tracing.CapturePolicy
	expiresAt time.Time
}

func newTenantTracePolicyResolver(store *storage.Store) *tenantTracePolicyResolver {
	return &tenantTracePolicyResolver{
		store: store,
		cache: map[string]tenantPolicyCache{},
		ttl:   60 * time.Second,
	}
}

func (r *tenantTracePolicyResolver) ResolveTenantPolicy(tenantID string) (tracing.CapturePolicy, bool) {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return "", false
	}

	now := time.Now()

	r.mu.RLock()
	cached, ok := r.cache[tenantID]
	r.mu.RUnlock()

	if ok && now.Before(cached.expiresAt) {
		return cached.policy, true
	}

	ctx, cancel := context.WithTimeout(context.Background(), 400*time.Millisecond)
	defer cancel()

	policy, exists, err := r.store.ResolveTenantTraceCapturePolicy(ctx, tenantID)
	if err != nil || !exists {
		// Fail open to the global policy when lookup fails or times out.
		return "", false
	}

	out := tracing.ParseCapturePolicy(policy)

	r.mu.Lock()
	r.cache[tenantID] = tenantPolicyCache{
		policy:    out,
		expiresAt: now.Add(r.ttl),
	}
	r.mu.Unlock()

	return out, true
}

type costReconcileAdapter struct {
	store *storage.Store
}

func (a *costReconcileAdapter) ListSpendCostByTenantAndDay(ctx context.Context, day time.Time) ([]metricsobs.ReconcileTenantCostRow, error) {
	rows, err := a.store.ListSpendCostByTenantAndDay(ctx, day)
	if err != nil {
		return nil, err
	}

	out := make([]metricsobs.ReconcileTenantCostRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, metricsobs.ReconcileTenantCostRow{TenantID: r.TenantID, CostUSD: r.CostUSD})
	}

	return out, nil
}

func (a *costReconcileAdapter) ListMetricsCostByTenantAndDay(ctx context.Context, day time.Time) ([]metricsobs.ReconcileTenantCostRow, error) {
	rows, err := a.store.ListMetricsCostByTenantAndDay(ctx, day)
	if err != nil {
		return nil, err
	}

	out := make([]metricsobs.ReconcileTenantCostRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, metricsobs.ReconcileTenantCostRow{TenantID: r.TenantID, CostUSD: r.CostUSD})
	}

	return out, nil
}
