package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/bootstrap"
	"github.com/Delpak/shahrah/services/control-plane/internal/config"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		log.Printf("control-plane stopped with an error: %v", err)
		os.Exit(1)
	}
}

func run() error {
	logger, err := zap.NewProduction()
	if err != nil {
		return fmt.Errorf("initialize logger: %w", err)
	}
	defer func() {
		if syncErr := logger.Sync(); syncErr != nil {
			log.Printf("logger sync failed: %v", syncErr)
		}
	}()

	cfg, err := config.LoadConfig(logger)
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	application, err := bootstrap.Build(context.Background(), cfg, logger)
	if err != nil {
		return fmt.Errorf("build control-plane application: %w", err)
	}

	runCtx, stopSignals := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stopSignals()

	if err := application.Start(runCtx); err != nil {
		return fmt.Errorf("start control-plane workers: %w", err)
	}

	shutdownResult := make(chan error, 1)

	go func() {
		<-runCtx.Done()
		logger.Info("Shutting down gracefully")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		shutdownResult <- application.Shutdown(shutdownCtx)
	}()

	logger.Info(
		"Starting Shahrah control-plane",
		zap.String("addr", application.Address),
		zap.String("version", "0.2.0"),
	)

	listenErr := application.Listen()
	unexpectedStop := runCtx.Err() == nil

	stopSignals()

	shutdownErr := <-shutdownResult

	if !unexpectedStop {
		listenErr = nil
	} else if listenErr == nil {
		listenErr = errors.New("control-plane server stopped unexpectedly")
	} else {
		listenErr = fmt.Errorf("control-plane server stopped unexpectedly: %w", listenErr)
	}

	return errors.Join(listenErr, shutdownErr)
}
