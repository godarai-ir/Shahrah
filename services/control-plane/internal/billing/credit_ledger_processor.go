package billing

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	metricsobs "github.com/Delpak/shahrah/services/control-plane/internal/observability/metrics"
	"go.uber.org/zap"
)

const (
	creditPartitionMonthsAhead = 3
	creditSweepLookback        = 48 * time.Hour
	// creditSettlementMaxPasses bounds how many full batches one tick drains.
	creditSettlementMaxPasses = 20
)

// creditLedgerStore is the persistence the credit ledger worker drives.
type creditLedgerStore interface {
	SettleCreditConsumption(ctx context.Context, batchSize int, settleDelay time.Duration) (credits.SettlementResult, error)
	SweepCreditConsumption(ctx context.Context, since time.Time) (credits.SettlementResult, error)
	ExpireCreditLots(ctx context.Context, now time.Time, limit int) (credits.ExpiryResult, error)
	GrantDuePlanCredits(ctx context.Context, now time.Time) (credits.GrantResult, error)
	EnsureCreditLedgerPartitions(ctx context.Context, now time.Time, monthsAhead int) (int, error)
	ReconcileCreditLedger(ctx context.Context) ([]credits.ReconciliationDiff, error)
}

// CreditLedgerConfig controls the credit ledger background worker.
type CreditLedgerConfig struct {
	Enabled bool
	// SettlementInterval is how often committed spend logs become consume entries.
	SettlementInterval time.Duration
	// SettleDelay leaves the newest spend logs for a later pass, so a slow writer's
	// lower id is never passed over.
	SettleDelay time.Duration
	BatchSize   int
	// ExpiryInterval is how often plan and trial credits are granted, expired lots
	// are closed, and partitions are kept ahead.
	ExpiryInterval time.Duration
	// ReconcileInterval is how often the ledger is reconciled and recently settled
	// spend logs are swept for gaps.
	ReconcileInterval time.Duration
	Timeout           time.Duration
}

// CreditLedgerProcessor settles credit consumption, expires lots, and
// reconciles the ledger. Every job is idempotent, so overlapping replicas are safe.
type CreditLedgerProcessor struct {
	cfg    CreditLedgerConfig
	store  creditLedgerStore
	logger *zap.Logger
	now    func() time.Time
	cancel context.CancelFunc
	wg     sync.WaitGroup
	once   sync.Once
}

// NewCreditLedgerProcessor creates the credit ledger worker.
func NewCreditLedgerProcessor(cfg CreditLedgerConfig, store creditLedgerStore, logger *zap.Logger) *CreditLedgerProcessor {
	if logger == nil {
		logger = zap.NewNop()
	}

	cfg.SettlementInterval = positiveDuration(cfg.SettlementInterval, 30*time.Second)
	cfg.SettleDelay = positiveDuration(cfg.SettleDelay, 30*time.Second)
	cfg.ExpiryInterval = positiveDuration(cfg.ExpiryInterval, 5*time.Minute)
	cfg.ReconcileInterval = positiveDuration(cfg.ReconcileInterval, 24*time.Hour)
	cfg.Timeout = positiveDuration(cfg.Timeout, 2*time.Minute)

	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 500
	}

	return &CreditLedgerProcessor{cfg: cfg, store: store, logger: logger, now: time.Now}
}

func positiveDuration(value, fallback time.Duration) time.Duration {
	if value <= 0 {
		return fallback
	}

	return value
}

// Start launches the jobs when the ledger is enabled.
func (p *CreditLedgerProcessor) Start() {
	if p == nil || !p.cfg.Enabled || p.store == nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel

	p.loop(ctx, p.cfg.SettlementInterval, p.Settle)
	p.loop(ctx, p.cfg.ExpiryInterval, p.GrantAndExpire)
	p.loop(ctx, p.cfg.ReconcileInterval, p.Reconcile)
}

// Stop waits for every job to exit.
func (p *CreditLedgerProcessor) Stop() {
	if p == nil {
		return
	}

	p.once.Do(func() {
		if p.cancel != nil {
			p.cancel()
		}

		p.wg.Wait()
	})
}

func (p *CreditLedgerProcessor) loop(ctx context.Context, interval time.Duration, job func(context.Context)) {
	p.wg.Add(1)

	go func() {
		defer p.wg.Done()

		p.runJob(ctx, job)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				p.runJob(ctx, job)
			}
		}
	}()
}

func (p *CreditLedgerProcessor) runJob(ctx context.Context, job func(context.Context)) {
	workCtx, cancel := context.WithTimeout(ctx, p.cfg.Timeout)
	defer cancel()

	job(workCtx)
}

// Settle drains committed spend logs into consume entries.
func (p *CreditLedgerProcessor) Settle(ctx context.Context) {
	for range creditSettlementMaxPasses {
		result, err := p.store.SettleCreditConsumption(ctx, p.cfg.BatchSize, p.cfg.SettleDelay)
		metricsobs.ObserveCreditSettlement(result, false)

		if err != nil {
			p.fail("settlement", err)

			return
		}

		if result.Settled > 0 || result.KeyConflicts > 0 {
			p.logger.Info("credit consumption settled",
				zap.Int("entries", result.Settled), zap.Int("key_conflicts", result.KeyConflicts),
				zap.Int64("debt_mc", result.DebtMC), zap.Int64("cursor", result.LastSpendLogID))
		}

		if !result.Locked || result.Scanned < p.cfg.BatchSize {
			return
		}
	}
}

// GrantAndExpire grants due plan and trial credits, then expires lots. Grants
// run first so a renewal's new slice is spendable before the old one expires.
func (p *CreditLedgerProcessor) GrantAndExpire(ctx context.Context) {
	p.Grant(ctx)
	p.Expire(ctx)
}

// Grant gives credit tenants the plan slice or trial credits they are due.
func (p *CreditLedgerProcessor) Grant(ctx context.Context) {
	result, err := p.store.GrantDuePlanCredits(ctx, p.now().UTC())
	metricsobs.ObserveCreditGrants(result)

	if err != nil {
		p.fail("grants", err)

		return
	}

	if result.Granted > 0 {
		p.logger.Info("credit plan grants written", zap.Int("grants", result.Granted), zap.Int64("granted_mc", result.GrantedMC))
	}
}

// Expire closes expired lots and keeps the ledger partitions ahead of time.
func (p *CreditLedgerProcessor) Expire(ctx context.Context) {
	now := p.now().UTC()

	if created, err := p.store.EnsureCreditLedgerPartitions(ctx, now, creditPartitionMonthsAhead); err != nil {
		p.fail("partitions", err)
	} else if created > 0 {
		p.logger.Info("credit ledger partitions created", zap.Int("count", created))
	}

	result, err := p.store.ExpireCreditLots(ctx, now, 0)
	metricsobs.ObserveCreditExpiry(result)

	if err != nil {
		p.fail("expiry", err)

		return
	}

	if result.ExpiredLots > 0 {
		p.logger.Info("credit lots expired",
			zap.Int("lots", result.ExpiredLots), zap.Int64("expired_mc", result.ExpiredMC), zap.Int("tenants", len(result.TenantIDs)))
	}
}

// Reconcile sweeps recent spend logs for gaps, then compares the ledger with
// its projections. Differences are reported, never fixed.
func (p *CreditLedgerProcessor) Reconcile(ctx context.Context) {
	sweep, err := p.store.SweepCreditConsumption(ctx, p.now().UTC().Add(-creditSweepLookback))
	metricsobs.ObserveCreditSettlement(sweep, true)

	switch {
	case err != nil:
		p.fail("sweep", err)
	case sweep.Settled > 0:
		p.logger.Warn("credit settlement sweep recovered spend logs the cursor had passed", zap.Int("entries", sweep.Settled))
	}

	diffs, err := p.store.ReconcileCreditLedger(ctx)
	if err != nil {
		p.fail("reconciliation", err)

		return
	}

	metricsobs.ObserveCreditReconciliation(diffs)

	for _, diff := range diffs {
		p.logger.Error("credit ledger reconciliation difference",
			zap.String("tenant_id", diff.TenantID), zap.String("check", diff.Check),
			zap.Int64("expected_mc", diff.ExpectedMC), zap.Int64("actual_mc", diff.ActualMC))
	}
}

func (p *CreditLedgerProcessor) fail(job string, err error) {
	if errors.Is(err, context.Canceled) {
		return
	}

	metricsobs.ObserveCreditLedgerJobFailure(job)
	p.logger.Warn("credit ledger job failed", zap.String("job", job), zap.Error(err))
}
