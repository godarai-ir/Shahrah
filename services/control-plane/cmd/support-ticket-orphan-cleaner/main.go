package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/Delpak/shahrah/services/control-plane/internal/supportfiles"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		log.Printf("support attachment orphan cleanup failed: %v", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	logger, err := zap.NewProduction()
	if err != nil {
		return fmt.Errorf("initialize logger: %w", err)
	}
	defer logger.Sync() //nolint:errcheck // Best-effort flush during command shutdown.

	store, err := storage.NewStore(ctx, databaseURL, logger, storage.StoreDependencies{})
	if err != nil {
		return fmt.Errorf("initialize storage: %w", err)
	}
	defer store.Close()

	files, err := supportfiles.NewFromEnv()
	if err != nil {
		return fmt.Errorf("initialize attachment storage: %w", err)
	}

	if files == nil {
		return fmt.Errorf("support attachment storage is not configured")
	}

	referenced, err := store.ListSupportTicketAttachmentObjectKeys(ctx)
	if err != nil {
		return fmt.Errorf("list referenced attachment objects: %w", err)
	}

	grace := durationEnv("SUPPORT_ORPHAN_GRACE_PERIOD", 24*time.Hour)

	deleted, err := files.CleanupOrphans(ctx, referenced, time.Now().UTC().Add(-grace))
	if err != nil {
		return fmt.Errorf("clean orphan attachment objects: %w", err)
	}

	log.Printf("support attachment orphan cleanup complete: deleted=%d referenced=%d", deleted, len(referenced))

	return nil
}

func durationEnv(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	if seconds, err := strconv.Atoi(value); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}

	if duration, err := time.ParseDuration(value); err == nil && duration > 0 {
		return duration
	}

	return fallback
}
