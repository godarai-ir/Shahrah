package bootstrap

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/logretention"
	metricsobs "github.com/Delpak/shahrah/services/control-plane/internal/observability/metrics"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/Delpak/shahrah/services/control-plane/internal/proxy"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
)

func (b *builder) buildWorkers() error {
	if err := b.buildBillingWorkers(); err != nil {
		return err
	}

	if err := b.buildMetricsWorkers(); err != nil {
		return err
	}

	if err := b.buildAsyncLoggers(); err != nil {
		return err
	}

	if err := b.buildTracingWorkers(); err != nil {
		return err
	}

	retention := logretention.NewRetentionCleaner(b.cfg.Observability.LogRetention, b.store.RetentionRepository, b.logger)
	if err := b.lifecycle.AddWorker("log retention cleaner", func(context.Context) error {
		retention.Start()

		return nil
	}, func(context.Context) error {
		retention.Stop()

		return nil
	}); err != nil {
		return err
	}

	tokenUsage := b.store.TokenUsageRepository

	return b.lifecycle.AddWorker("token usage flusher", func(ctx context.Context) error {
		tokenUsage.StartTokenUsageFlusher(ctx)

		return nil
	}, nil)
}

func (b *builder) buildBillingWorkers() error {
	subscriptions := billing.NewSubscriptionProcessor(b.store.SubscriptionRepository, b.logger, time.Minute)
	if err := b.lifecycle.AddWorker("subscription processor", func(context.Context) error {
		subscriptions.Start()

		return nil
	}, func(context.Context) error {
		subscriptions.Stop()

		return nil
	}); err != nil {
		return err
	}

	nobitexSource := b.cfg.Runtime.NobitexExchangeRateSource
	if nobitexSource == "" {
		nobitexSource = "nobitex:" + b.cfg.Runtime.NobitexExchangeRateSymbol
	}

	nobitexEnabled := b.cfg.Runtime.NobitexExchangeRateEnabled
	if !b.cfg.Runtime.NobitexExchangeRateConfigured {
		nobitexEnabled = strings.EqualFold(b.cfg.BillingSettings.RateSource, "nobitex")
	}

	exchangeRates := billing.NewNobitexExchangeRateProcessor(
		billing.NobitexExchangeRateConfig{
			Enabled:   nobitexEnabled,
			BaseURL:   b.cfg.Runtime.NobitexExchangeRateBaseURL,
			Symbol:    b.cfg.Runtime.NobitexExchangeRateSymbol,
			Source:    nobitexSource,
			Interval:  b.cfg.Runtime.NobitexExchangeRateInterval,
			Timeout:   b.cfg.Runtime.NobitexExchangeRateTimeout,
			MinTrades: b.cfg.Runtime.NobitexExchangeRateMinTrades,
		},
		b.store.ExchangeRateRepository,
		b.logger,
	)
	if err := b.lifecycle.AddWorker("exchange rate processor", func(context.Context) error {
		exchangeRates.Start()

		return nil
	}, func(context.Context) error {
		exchangeRates.Stop()

		return nil
	}); err != nil {
		return err
	}

	if err := b.buildCreditLedgerWorker(); err != nil {
		return err
	}

	spendPurge := billing.NewSpendPurgeProcessor(b.store.SpendHistoryPurgeRepository, b.logger, 5*time.Minute)
	if err := b.lifecycle.AddWorker("clickhouse spend purge", func(context.Context) error {
		spendPurge.Start()

		return nil
	}, func(context.Context) error {
		spendPurge.Stop()

		return nil
	}); err != nil {
		return err
	}

	settlementCfg := billing.WalletSettlementConfig{
		Enabled:   b.cfg.Runtime.WalletSettlementEnabled,
		Interval:  b.cfg.Runtime.WalletSettlementInterval,
		Timeout:   b.cfg.Runtime.WalletSettlementTimeout,
		BatchSize: b.cfg.Runtime.WalletSettlementBatchSize,
	}

	liveBudgetCache, ok := b.cache.(cache.BudgetWindowInspector)
	if !ok {
		if settlementCfg.Enabled {
			b.logger.Warn("Wallet settlement worker is disabled because the cache cannot inspect budget windows")
		}

		return nil
	}

	settlement := billing.NewWalletSettlementProcessor(
		settlementCfg,
		liveBudgetCache,
		b.store.WalletLedgerRepository,
		b.logger,
	)

	return b.lifecycle.AddWorker("wallet settlement processor", func(context.Context) error {
		settlement.Start()

		return nil
	}, func(context.Context) error {
		settlement.Stop()

		return nil
	})
}

func (b *builder) buildCreditLedgerWorker() error {
	settings := b.cfg.Runtime.BillingCreditLedger
	if b.store == nil || b.store.WalletRepository == nil {
		return nil
	}

	processor := billing.NewCreditLedgerProcessor(billing.CreditLedgerConfig{
		SettlementInterval: settings.SettlementInterval,
		SettleDelay:        settings.SettleDelay,
		BatchSize:          settings.BatchSize,
		ExpiryInterval:     settings.ExpiryInterval,
		ReconcileInterval:  settings.ReconcileInterval,
		PriceInterval:      settings.PriceInterval,
		PriceAutoApply:     settings.PriceAutoApply,
		Timeout:            settings.Timeout,
	}, creditLedgerWorkerStore{
		CreditLedgerRepository:   b.store.CreditLedgerRepository,
		CreditPriceRepository:    b.store.CreditPriceRepository,
		CreditRateCardRepository: b.store.CreditRateCardRepository,
	}, b.logger)

	return b.lifecycle.AddWorker("credit ledger processor", func(context.Context) error {
		processor.Start()

		return nil
	}, func(context.Context) error {
		processor.Stop()

		return nil
	})
}

func (b *builder) buildMetricsWorkers() error {
	cfg := b.cfg.Observability.Metrics
	if !cfg.Enabled {
		return nil
	}

	var dlq metricsobs.DeadLetterQueue

	if cfg.DLQ.Backend == "redis_stream" {
		var err error

		dlq, err = newMetricsDLQ(b.ctx, cfg.DLQ)
		if err != nil {
			return fmt.Errorf("initialize model metrics Redis Stream DLQ: %w", err)
		}

		if err := b.lifecycle.AddResource("metrics dead-letter queue", func(context.Context) error {
			return dlq.Close()
		}); err != nil {
			return err
		}
	} else {
		b.logger.Warn("Model metrics DLQ is disabled; failed ClickHouse exports can be dropped")
	}

	queue := metricsobs.NewQueue(cfg.QueueCapacity)

	processor := metricsobs.NewProcessor(cfg, queue, b.store.MetricsRepository, b.logger, dlq)
	if err := b.lifecycle.AddWorker("metrics processor", func(context.Context) error {
		processor.Start()

		return nil
	}, processor.Stop); err != nil {
		return err
	}

	shadowCredits := b.cfg.Runtime.BillingCreditsShadowEnabled
	b.emitMetric = func(input metricsobs.BuildInput) {
		event := processor.BuildEvent(input)
		if shadowCredits {
			metricsobs.ApplyShadowCredits(&event, b.creditRateCards.CurrentRateCard())
		}

		_ = queue.Enqueue(event)
	}

	reconciler := metricsobs.NewCostReconciler(
		b.cfg.Observability.MetricsReconcile,
		&costReconcileAdapter{store: b.store.MetricsRepository},
		b.logger,
	)

	return b.lifecycle.AddWorker("cost reconciler", func(context.Context) error {
		reconciler.Start()

		return nil
	}, func(context.Context) error {
		reconciler.Stop()

		return nil
	})
}

func (b *builder) buildAsyncLoggers() error {
	b.requestLogger = b.store.RequestLogRepository
	if b.cfg.RequestLogs.AsyncEnabled {
		writer := proxy.NewManagedAsyncRequestLogWriter(b.store.RequestLogRepository, proxy.AsyncRequestLogConfig{
			QueueSize:      b.cfg.RequestLogs.QueueSize,
			Workers:        b.cfg.RequestLogs.Workers,
			BatchSize:      b.cfg.RequestLogs.BatchSize,
			FlushInterval:  time.Duration(b.cfg.RequestLogs.FlushIntervalMS) * time.Millisecond,
			EnqueueTimeout: time.Duration(b.cfg.RequestLogs.EnqueueTimeoutMS) * time.Millisecond,
			DropOnFull:     b.cfg.RequestLogs.DropOnFull,
		}, b.logger, b.httpMetrics)
		b.requestLogger = writer

		if err := b.lifecycle.AddWorker("async request logger", func(context.Context) error {
			writer.Start()

			return nil
		}, func(ctx context.Context) error {
			return stopWithConfiguredTimeout(ctx, b.cfg.RequestLogs.ShutdownDrainTimeoutSeconds, writer.Stop)
		}); err != nil {
			return err
		}
	}

	b.spendLogger = b.withCreditShadow(b.store.SpendRepository)
	if !b.cfg.SpendLogs.AsyncEnabled {
		return nil
	}

	writer := proxy.NewManagedAsyncSpendLogWriter(b.store.SpendRepository, proxy.AsyncSpendLogConfig{
		QueueSize:      b.cfg.SpendLogs.QueueSize,
		Workers:        b.cfg.SpendLogs.Workers,
		BatchSize:      b.cfg.SpendLogs.BatchSize,
		FlushInterval:  time.Duration(b.cfg.SpendLogs.FlushIntervalMS) * time.Millisecond,
		EnqueueTimeout: time.Duration(b.cfg.SpendLogs.EnqueueTimeoutMS) * time.Millisecond,
		DropOnFull:     b.cfg.SpendLogs.DropOnFull,
	}, b.logger, b.httpMetrics)
	b.spendLogger = b.withCreditShadow(writer)

	return b.lifecycle.AddWorker("async spend logger", func(context.Context) error {
		writer.Start()

		return nil
	}, func(ctx context.Context) error {
		return stopWithConfiguredTimeout(ctx, b.cfg.SpendLogs.ShutdownDrainTimeoutSeconds, writer.Stop)
	})
}

func (b *builder) buildTracingWorkers() error {
	cfg := b.cfg.Observability.Tracing
	queue := tracing.NewQueue(cfg.QueueCapacity)
	sampler := tracing.NewSampler(cfg.DefaultSampleRate, cfg.SlowRequest.Milliseconds())
	privacy := tracing.NewPrivacyManager(cfg)
	privacy.SetTenantPolicyResolver(newTenantTracePolicyResolver(b.store.ObservabilitySettingsRepository))

	exporter, err := tracing.NewReliableExporter(cfg, b.store.TraceRepository, b.logger)
	if err != nil {
		return fmt.Errorf("initialize tracing exporter: %w", err)
	}

	b.traceHealth = exporter
	b.tracingMiddleware = tracing.Middleware(queue, sampler, cfg.SlowRequest.Milliseconds(), privacy)

	if err := b.lifecycle.AddResource("tracing exporter", func(context.Context) error {
		return exporter.Close()
	}); err != nil {
		return err
	}

	processor := tracing.NewProcessor(cfg, queue, exporter, b.logger)

	return b.lifecycle.AddWorker("trace processor", func(context.Context) error {
		processor.Start()

		return nil
	}, processor.Stop)
}

func newMetricsDLQ(ctx context.Context, cfg metricsobs.DLQConfig) (metricsobs.DeadLetterQueue, error) {
	dlqCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	return metricsobs.NewRedisStreamDLQ(dlqCtx, cfg)
}

func stopWithConfiguredTimeout(
	parent context.Context,
	seconds int,
	stop func(context.Context) error,
) error {
	timeout := time.Duration(seconds) * time.Second
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	return stop(ctx)
}

// creditLedgerWorkerStore joins the repositories the credit ledger worker drives.
type creditLedgerWorkerStore struct {
	*storage.CreditLedgerRepository
	*storage.CreditPriceRepository
	*storage.CreditRateCardRepository
}
