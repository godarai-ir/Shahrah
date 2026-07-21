package billing

import (
	"context"
	"errors"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"go.uber.org/zap"
)

type walletBudgetWindowLister interface {
	ListActiveBudgetWindows(ctx context.Context, batchSize int) ([]cache.BudgetWindowSnapshot, error)
}

type walletBudgetSettlementStore interface {
	ApplyWalletBudgetSettlement(ctx context.Context, snapshot cache.BudgetWindowSnapshot) error
}

// WalletSettlementConfig controls the background ledger settlement worker.
type WalletSettlementConfig struct {
	Enabled   bool
	Interval  time.Duration
	Timeout   time.Duration
	BatchSize int
}

// LoadWalletSettlementConfigFromEnv reads settlement worker config from environment variables.
func LoadWalletSettlementConfigFromEnv() WalletSettlementConfig {
	return WalletSettlementConfig{
		Enabled:   boolEnv("WALLET_BUDGET_SETTLEMENT_ENABLED", true),
		Interval:  secondsEnv("WALLET_BUDGET_SETTLEMENT_INTERVAL_SECONDS", 300*time.Second),
		Timeout:   secondsEnv("WALLET_BUDGET_SETTLEMENT_TIMEOUT_SECONDS", 30*time.Second),
		BatchSize: intEnv("WALLET_BUDGET_SETTLEMENT_BATCH_SIZE", 200),
	}
}

// WalletSettlementProcessor mirrors committed credit-window spend into wallet ledger rows.
type WalletSettlementProcessor struct {
	cfg    WalletSettlementConfig
	cache  walletBudgetWindowLister
	store  walletBudgetSettlementStore
	logger *zap.Logger
	cancel context.CancelFunc
	done   chan struct{}
	once   sync.Once
}

// NewWalletSettlementProcessor creates a wallet settlement worker.
func NewWalletSettlementProcessor(cfg WalletSettlementConfig, liveCache walletBudgetWindowLister, store walletBudgetSettlementStore, logger *zap.Logger) *WalletSettlementProcessor {
	if logger == nil {
		logger = zap.NewNop()
	}

	if cfg.Interval <= 0 {
		cfg.Interval = 300 * time.Second
	}

	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}

	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 200
	}

	return &WalletSettlementProcessor{
		cfg:    cfg,
		cache:  liveCache,
		store:  store,
		logger: logger,
		done:   make(chan struct{}),
	}
}

// Start launches the settlement loop when configuration and dependencies allow it.
func (p *WalletSettlementProcessor) Start() {
	if p == nil || !p.cfg.Enabled || p.cache == nil || p.store == nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	p.cancel = cancel
	go p.run(ctx)
}

// Stop waits for the settlement loop to exit.
func (p *WalletSettlementProcessor) Stop() {
	if p == nil {
		return
	}

	p.once.Do(func() {
		if p.cancel != nil {
			p.cancel()
			<-p.done

			return
		}

		close(p.done)
	})
}

func (p *WalletSettlementProcessor) run(ctx context.Context) {
	defer close(p.done)

	p.processWithTimeout(ctx)

	ticker := time.NewTicker(p.cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.processWithTimeout(ctx)
		}
	}
}

func (p *WalletSettlementProcessor) processWithTimeout(ctx context.Context) {
	workCtx, cancel := context.WithTimeout(ctx, p.cfg.Timeout)
	defer cancel()

	started := time.Now()

	snapshots, err := p.cache.ListActiveBudgetWindows(workCtx, p.cfg.BatchSize)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}

		p.logger.Warn("wallet budget settlement scan failed", zap.Error(err))

		return
	}

	settled := 0

	for _, snapshot := range snapshots {
		// Policy windows mirror the same committed spend for enforcement only.
		// Settling them would charge the tenant wallet once per matching policy.
		if !strings.HasPrefix(strings.TrimSpace(snapshot.WindowKey), "budget:tenant:") {
			continue
		}

		if err := p.store.ApplyWalletBudgetSettlement(workCtx, snapshot); err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}

			p.logger.Warn("wallet budget settlement apply failed", zap.Error(err), zap.String("window_key", snapshot.WindowKey))

			continue
		}

		settled++
	}

	if settled > 0 {
		p.logger.Info("wallet budget settlements applied", zap.Int("windows", settled), zap.Duration("elapsed", time.Since(started)))
	}
}

func boolEnv(key string, fallback bool) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}

	return parsed
}

func secondsEnv(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}

	return time.Duration(value) * time.Second
}

func intEnv(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}

	return value
}
