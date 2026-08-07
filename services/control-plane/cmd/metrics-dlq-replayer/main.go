package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/config"
	controlconfig "github.com/Delpak/shahrah/services/control-plane/internal/config/environment"
	metricsobs "github.com/Delpak/shahrah/services/control-plane/internal/observability/metrics"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("initialize logger: %v", err)
	}
	defer func() {
		if syncErr := logger.Sync(); syncErr != nil {
			log.Printf("logger sync failed: %v", syncErr)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	storageSettings := config.LoadStorageSettings()
	storageConfig := storage.DefaultConfig()
	storageConfig.ClickHouse = storage.ClickHouseConfig{
		DSN:        storageSettings.ClickHouseDSN,
		Required:   true,
		Attempts:   storageSettings.ClickHouseAttempts,
		RetryDelay: storageSettings.ClickHouseRetryDelay,
	}

	store, err := storage.NewMetricsStore(ctx, logger, storageConfig)
	if err != nil {
		logger.Fatal("initialize storage", zap.Error(err))
	}
	defer store.Close()

	dlqCfg := config.LoadObservabilitySettings().Metrics.DLQ

	replayer, err := metricsobs.NewReplayer(
		ctx,
		dlqCfg,
		store.ObservabilityRepository,
		logger,
		controlconfig.Environment("HOSTNAME"),
	)
	if err != nil {
		logger.Fatal("initialize Redis Stream DLQ replayer", zap.Error(err))
	}

	defer func() {
		if closeErr := replayer.Close(); closeErr != nil {
			logger.Warn("close Redis Stream DLQ replayer", zap.Error(closeErr))
		}
	}()

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/live", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)

		if _, writeErr := w.Write([]byte("ok")); writeErr != nil {
			logger.Debug("write liveness response", zap.Error(writeErr))
		}
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, req *http.Request) {
		readyCtx, cancel := context.WithTimeout(req.Context(), 500*time.Millisecond)
		defer cancel()

		if err := replayer.Ready(readyCtx); err != nil {
			http.Error(w, "redis unavailable", http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusOK)

		if _, writeErr := w.Write([]byte("ok")); writeErr != nil {
			logger.Debug("write readiness response", zap.Error(writeErr))
		}
	})

	server := &http.Server{
		Addr:              stringEnvOrDefault("METRICS_DLQ_REPLAYER_HTTP_ADDR", ":9091"),
		Handler:           mux,
		ReadHeaderTimeout: 3 * time.Second,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("serve metrics DLQ replayer health endpoint", zap.Error(err))
		}
	}()

	logger.Info("model metrics Redis Stream DLQ replayer started")

	if err := replayer.Run(ctx); err != nil {
		logger.Error("model metrics Redis Stream DLQ replayer stopped", zap.Error(err))
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Warn("shutdown replayer health server", zap.Error(err))
	}
}

func stringEnvOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(controlconfig.Environment(key)); value != "" {
		return value
	}

	return fallback
}
