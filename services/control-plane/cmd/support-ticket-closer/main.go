package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"go.uber.org/zap"
)

func main() {
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("initialize logger: %v", err)
	}
	defer logger.Sync()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	store, err := storage.NewStore(ctx, databaseURL, logger)
	if err != nil {
		logger.Fatal("initialize storage", zap.Error(err))
	}
	defer store.Close()

	days := positiveIntEnv("SUPPORT_TICKET_RESOLVED_CLOSE_DAYS", 7)
	batchSize := positiveIntEnv("SUPPORT_TICKET_CLOSE_BATCH_SIZE", 250)
	closed, err := store.CloseResolvedSupportTickets(ctx, time.Now().UTC().Add(-time.Duration(days)*24*time.Hour), batchSize)
	if err != nil {
		logger.Fatal("close resolved support tickets", zap.Error(err))
	}

	logger.Info("Resolved support ticket close run completed", zap.Int("closed_count", closed))
}

func positiveIntEnv(key string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(os.Getenv(key)))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
