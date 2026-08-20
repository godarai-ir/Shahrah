package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Delpak/shahrah/services/rag/internal/app"
	"github.com/Delpak/shahrah/services/rag/internal/config"
	"github.com/Delpak/shahrah/services/rag/internal/files"
	"github.com/Delpak/shahrah/services/rag/internal/httpapi"
	"github.com/Delpak/shahrah/services/rag/internal/parser"
	"github.com/Delpak/shahrah/services/rag/internal/provider"
	"github.com/Delpak/shahrah/services/rag/internal/store"
	"github.com/Delpak/shahrah/services/rag/internal/tokenization"
	"github.com/Delpak/shahrah/services/rag/internal/worker"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "initialize logger: %v\n", err)

		os.Exit(1)
	}
	defer func() {
		if syncErr := logger.Sync(); syncErr != nil {
			_, _ = fmt.Fprintf(os.Stderr, "flush logger: %v\n", syncErr)
		}
	}()

	if err := run(logger); err != nil {
		logger.Fatal("managed RAG service stopped", zap.Error(err))
	}
}

func run(logger *zap.Logger) error {
	settings, err := config.Load()
	if err != nil {
		return err
	}

	rootContext, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	repository, err := store.New(
		rootContext, settings.MetadataURL, settings.VectorURL, settings.RedisURL, settings.ConversationSecret,
	)
	if err != nil {
		return err
	}
	defer repository.Close()

	activeProfile, err := repository.LoadActiveRAGProfile(rootContext)
	if err != nil {
		return fmt.Errorf("load active RAG profile: %w", err)
	}

	gateway, err := provider.NewGatewayClient(
		settings.GatewayURL, settings.GatewayAPIKey, settings.GatewayRerankModel, settings.GatewayTimeout,
		provider.SemanticEmbeddingConfig{
			BaseURL: settings.EmbeddingBaseURL, APIKey: settings.EmbeddingAPIKey,
			Model: settings.EmbeddingModel, Revision: settings.EmbeddingRevision,
			Dimensions: settings.EmbeddingDimensions, Timeout: settings.EmbeddingTimeout,
			MaxBatchSize: settings.EmbeddingMaxBatchSize,
		},
	)
	if err != nil {
		return err
	}

	fileService, err := files.New(settings.Files)
	if err != nil {
		return err
	}

	application, err := app.New(repository, gateway)
	if err != nil {
		return err
	}

	if settings.Mode == "worker" {
		return runWorker(rootContext, settings, repository, fileService, gateway, logger)
	}

	server, err := httpapi.New(httpapi.Dependencies{
		Store: repository, Application: application, Files: fileService, Logger: logger,
		InternalToken: settings.InternalToken, SessionSecret: settings.SessionSecret,
		Profile: activeProfile,
	})
	if err != nil {
		return err
	}

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.Listen(settings.Address)
	}()

	select {
	case <-rootContext.Done():
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		return server.ShutdownWithContext(shutdownContext)
	case err := <-serverErrors:
		if errors.Is(err, context.Canceled) {
			return nil
		}

		return err
	}
}

func runWorker(
	ctx context.Context,
	settings config.Config,
	repository *store.Store,
	fileService *files.Service,
	gateway *provider.GatewayClient,
	logger *zap.Logger,
) error {
	textParser, err := parser.NewTikaClient(settings.TikaURL, 2*time.Minute)
	if err != nil {
		return err
	}

	tokenizer, err := tokenization.NewCL100K()
	if err != nil {
		return err
	}

	processor, err := worker.New(
		repository, fileService, textParser, gateway, gateway, tokenizer, logger, settings.WorkerID,
	)
	if err != nil {
		return err
	}

	return processor.Run(ctx)
}
