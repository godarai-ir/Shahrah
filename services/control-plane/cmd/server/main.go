package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
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
	metricsobs "github.com/Delpak/shahrah/services/control-plane/internal/observability/metrics"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/Delpak/shahrah/services/control-plane/internal/proxy"
	"github.com/Delpak/shahrah/services/control-plane/internal/responsecache"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	"github.com/Delpak/shahrah/services/control-plane/internal/server"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	tunneltransport "github.com/Delpak/shahrah/services/control-plane/internal/tunnel"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	cfg, err := config.LoadConfig(logger)
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}
	if strings.TrimSpace(cfg.GeneralSettings.DatabaseURL) == "" {
		logger.Fatal("Missing required database configuration", zap.String("field", "general_settings.database_url"))
	}

	ctx := context.Background()
	store, err := storage.NewStore(ctx, cfg.GeneralSettings.DatabaseURL, logger)
	if err != nil {
		logger.Fatal("Failed to initialize storage", zap.Error(err))
	}
	defer store.Close()

	var dualCache cache.Cache
	if cfg.RouterSettings.RedisHost != "" {
		dualCache, err = cache.NewDualCache(cfg.RouterSettings.RedisHost, cfg.RouterSettings.RedisPort, cfg.RouterSettings.RedisPassword, logger)
		if err != nil {
			logger.Warn("Failed to initialize Redis cache, falling back to memory-only", zap.Error(err))
			dualCache, _ = cache.NewMemoryCache(logger)
		}
	} else {
		dualCache, _ = cache.NewMemoryCache(logger)
	}
	defer dualCache.Close()
	_, sharedResponseCacheAvailable := dualCache.(*cache.DualCache)
	store.SetEntitlementCache(dualCache)

	llmRouter, err := router.NewRouter(cfg.ModelList, cfg.RouterSettings, dualCache, store, logger)
	if err != nil {
		logger.Fatal("Failed to initialize router", zap.Error(err))
	}

	authMiddleware := auth.NewAuthMiddleware(store, cfg.GeneralSettings.MasterKey, logger)
	tenantMiddleware := tenant.NewMiddlewareWithCache(store, dualCache, tenant.MiddlewareConfig{
		AllowHeaderHint: cfg.TenantSettings.AllowHeaderHint,
		CacheTTL:        5 * time.Minute,
	})
	billingService := billing.NewService(dualCache, store, cfg.BillingSettings.DefaultUSDtoIRR, cfg.BillingSettings.RateSource)
	subscriptionProcessor := billing.NewSubscriptionProcessor(store, logger, time.Minute)
	subscriptionProcessor.Start()
	defer subscriptionProcessor.Stop()
	tunnelClient := tunneltransport.NewHTTPTransport(cfg.TunnelSettings.ClientURL, time.Duration(cfg.TunnelSettings.Timeout)*time.Second)

	metricsCfg := metricsobs.LoadConfigFromEnv()
	var emitMetric func(input metricsobs.BuildInput)
	var reconciler *metricsobs.CostReconciler
	if metricsCfg.Enabled {
		metricsQueue := metricsobs.NewQueue(metricsCfg.QueueCapacity)
		metricsProcessor := metricsobs.NewProcessor(metricsCfg, metricsQueue, store, logger)
		metricsProcessor.Start()
		defer func() {
			stopCtx, cancelStop := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancelStop()
			_ = metricsProcessor.Stop(stopCtx)
		}()
		emitMetric = func(input metricsobs.BuildInput) {
			event := metricsProcessor.BuildEvent(input)
			_ = metricsQueue.Enqueue(event)
		}

		reconciler = metricsobs.NewCostReconciler(metricsobs.LoadReconcileConfigFromEnv(), &costReconcileAdapter{store: store}, logger)
		reconciler.Start()
		defer reconciler.Stop()
	}

	adminHandlers := admin.New(store, store, store)
	authHandlers := admin.NewAuthHandlers(store, admin.NewOTPEmailSenderFromEnv())
	proxyHandler := proxy.NewChatCompletionHandler(llmRouter, tunnelClient, billingService, store, store, logger, emitMetric)
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
	completionsHandler := proxy.NewCompletionsHandler(llmRouter, tunnelClient, billingService, store, store, logger, emitMetric)
	responsesHandler := proxy.NewResponsesHandler(llmRouter, tunnelClient, billingService, store, store, logger, emitMetric)
	embeddingsHandler := proxy.NewEmbeddingsHandler(llmRouter, tunnelClient, billingService, store, store, logger, emitMetric)
	moderationsHandler := proxy.NewModerationsHandler(llmRouter, tunnelClient, billingService, store, store, logger, emitMetric)
	imagesHandler := proxy.NewImagesHandler(llmRouter, tunnelClient, billingService, store, store, logger, emitMetric)
	modelsHandler := proxy.NewModelsHandler(store)

	traceCfg := tracing.LoadConfigFromEnv()
	traceQueue := tracing.NewQueue(traceCfg.QueueCapacity)
	traceSampler := tracing.NewSampler(traceCfg.DefaultSampleRate, traceCfg.SlowRequest.Milliseconds())
	tracePrivacy := tracing.NewPrivacyManager(traceCfg)
	tracePrivacy.SetTenantPolicyResolver(newTenantTracePolicyResolver(store))
	reliableExporter, err := tracing.NewReliableExporter(traceCfg, store, logger)
	if err != nil {
		logger.Fatal("Failed to initialize tracing exporter", zap.Error(err))
	}
	defer reliableExporter.Close()
	traceProcessor := tracing.NewProcessor(traceCfg, traceQueue, reliableExporter, logger)
	traceProcessor.Start()
	defer func() {
		stopCtx, cancelStop := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancelStop()
		_ = traceProcessor.Stop(stopCtx)
	}()
	traceRetentionCleaner := tracing.NewRetentionCleaner(traceCfg, store, logger)
	traceRetentionCleaner.Start()
	defer traceRetentionCleaner.Stop()

	app := server.NewFiberApp(cfg, server.Dependencies{
		AuthMiddleware:          authMiddleware,
		TenantResolver:          tenantMiddleware,
		TracingMiddleware:       tracing.Middleware(traceQueue, traceSampler, traceCfg.SlowRequest.Milliseconds(), tracePrivacy),
		TraceHealth:             reliableExporter,
		AdminHandlers:           adminHandlers,
		AuthHandlers:            authHandlers,
		ProxyHandler:            proxyHandler,
		CompletionsHandler:      completionsHandler,
		ResponsesHandler:        responsesHandler,
		EmbeddingsHandler:       embeddingsHandler,
		ModerationsHandler:      moderationsHandler,
		ImagesHandler:           imagesHandler,
		ModelsHandler:           modelsHandler,
		Logger:                  logger,
		PlatformServiceAccounts: store,
	})

	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
		if err != nil && !errors.Is(err, context.DeadlineExceeded) {
			// Fail open to global policy.
		}
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
