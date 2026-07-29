package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenantcredentials"
	"go.uber.org/zap"
)

func main() {
	batchSize := flag.Int("batch-size", 100, "maximum rows per transaction")
	once := flag.Bool("once", false, "process one batch and exit")
	flag.Parse()

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = logger.Sync() }()

	ctx := context.Background()
	store, err := storage.NewStore(ctx, databaseURL, logger)
	if err != nil {
		logger.Fatal("initialize storage", zap.Error(err))
	}
	defer store.Close()

	secretStore, err := tenantcredentials.NewEncryptedStoreFromEnv()
	if err != nil {
		logger.Fatal("initialize tenant credential encryption", zap.Error(err))
	}
	store.SetTenantCredentialStore(secretStore)

	for {
		result, err := store.MigrateLegacyTenantProviderCredentials(
			ctx,
			*batchSize,
		)
		if err != nil {
			logger.Fatal("migrate tenant credentials", zap.Error(err))
		}
		logger.Info(
			"tenant credential migration batch complete",
			zap.Int("migrated", result.Migrated),
			zap.Int("remaining", result.Remaining),
		)
		if *once || result.Remaining == 0 || result.Migrated == 0 {
			break
		}
	}
}
