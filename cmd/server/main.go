package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Delpak/shahrah/internal/cache"
	"github.com/Delpak/shahrah/internal/config"
	"github.com/Delpak/shahrah/internal/router"
	"github.com/Delpak/shahrah/internal/server"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Load configuration
	cfg, err := config.LoadConfig(logger)
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Initialize dual cache
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

	// Initialize router
	llmRouter, err := router.NewRouter(cfg.ModelList, cfg.RouterSettings, dualCache, logger)
	if err != nil {
		logger.Fatal("Failed to initialize router", zap.Error(err))
	}

	// Create and start Fiber server
	app := server.NewFiberApp(cfg, llmRouter, dualCache, logger)

	// Graceful shutdown handling
	ctx, cancel := context.WithCancel(context.Background())
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
		case <-ctx.Done():
			return
		}
	}()

	addr := fmt.Sprintf("%s:%d", cfg.GeneralSettings.Host, cfg.GeneralSettings.Port)
	logger.Info("Starting Shahrah", zap.String("addr", addr), zap.String("version", "0.1.0"))

	if err := app.Listen(addr); err != nil {
		logger.Fatal("Server failed to start", zap.Error(err))
	}
}

