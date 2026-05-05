package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/admin"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/config"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/Delpak/shahrah/services/control-plane/internal/proxy"
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
	tunnelClient := tunneltransport.NewHTTPTransport(cfg.TunnelSettings.ClientURL, time.Duration(cfg.TunnelSettings.Timeout)*time.Second)

	adminHandlers := admin.New(store, store, store)
	proxyHandler := proxy.NewChatCompletionHandler(llmRouter, tunnelClient, billingService, store, store, logger)
	completionsHandler := proxy.NewCompletionsHandler(llmRouter, tunnelClient, billingService, store, store, logger)
	modelsHandler := proxy.NewModelsHandler(store)

	traceCfg := tracing.LoadConfigFromEnv()
	traceQueue := tracing.NewQueue(traceCfg.QueueCapacity)
	traceSampler := tracing.NewSampler(traceCfg.DefaultSampleRate, traceCfg.SlowRequest.Milliseconds())
	traceProcessor := tracing.NewProcessor(traceCfg, traceQueue, store, logger)
	traceProcessor.Start()
	defer func() {
		stopCtx, cancelStop := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancelStop()
		_ = traceProcessor.Stop(stopCtx)
	}()

	app := server.NewFiberApp(cfg, server.Dependencies{
		AuthMiddleware:     authMiddleware,
		TenantResolver:     tenantMiddleware,
		TracingMiddleware:  tracing.Middleware(traceQueue, traceSampler, traceCfg.SlowRequest.Milliseconds()),
		AdminHandlers:      adminHandlers,
		ProxyHandler:       proxyHandler,
		CompletionsHandler: completionsHandler,
		ModelsHandler:      modelsHandler,
		Logger:             logger,
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
