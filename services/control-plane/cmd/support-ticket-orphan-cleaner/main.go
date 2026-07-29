package main

import (
	"context"
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("initialize logger: %v", err)
	}
	defer logger.Sync()

	store, err := storage.NewStore(ctx, databaseURL, logger)
	if err != nil {
		log.Fatalf("initialize storage: %v", err)
	}
	defer store.Close()

	files, err := supportfiles.NewFromEnv()
	if err != nil {
		log.Fatalf("initialize attachment storage: %v", err)
	}
	if files == nil {
		log.Fatal("support attachment storage is not configured")
	}

	referenced, err := store.ListSupportTicketAttachmentObjectKeys(ctx)
	if err != nil {
		log.Fatalf("list referenced attachment objects: %v", err)
	}
	grace := durationEnv("SUPPORT_ORPHAN_GRACE_PERIOD", 24*time.Hour)
	deleted, err := files.CleanupOrphans(ctx, referenced, time.Now().UTC().Add(-grace))
	if err != nil {
		log.Fatalf("clean orphan attachment objects: %v", err)
	}
	log.Printf("support attachment orphan cleanup complete: deleted=%d referenced=%d", deleted, len(referenced))
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
