package billing

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/stretchr/testify/require"
)

type fakeCreditLedgerStore struct {
	settleResults []credits.SettlementResult
	settleErr     error
	settleCalls   int
	sweepSince    time.Time
	expireNow     time.Time
	partitions    int
	grantNow      time.Time
	reconciled    int
	diffs         []credits.ReconciliationDiff
}

func (f *fakeCreditLedgerStore) SettleCreditConsumption(context.Context, int, time.Duration) (credits.SettlementResult, error) {
	f.settleCalls++
	if f.settleErr != nil {
		return credits.SettlementResult{Locked: true}, f.settleErr
	}

	if len(f.settleResults) == 0 {
		return credits.SettlementResult{Locked: true}, nil
	}

	next := f.settleResults[0]
	f.settleResults = f.settleResults[1:]

	return next, nil
}

func (f *fakeCreditLedgerStore) SweepCreditConsumption(_ context.Context, since time.Time) (credits.SettlementResult, error) {
	f.sweepSince = since

	return credits.SettlementResult{Locked: true}, nil
}

func (f *fakeCreditLedgerStore) ExpireCreditLots(_ context.Context, now time.Time, _ int) (credits.ExpiryResult, error) {
	f.expireNow = now

	return credits.ExpiryResult{}, nil
}

func (f *fakeCreditLedgerStore) GrantDuePlanCredits(_ context.Context, now time.Time) (credits.GrantResult, error) {
	f.grantNow = now

	return credits.GrantResult{Granted: 1, GrantedMC: 5_000_000}, nil
}

func (f *fakeCreditLedgerStore) EnsureCreditLedgerPartitions(_ context.Context, _ time.Time, monthsAhead int) (int, error) {
	f.partitions = monthsAhead

	return 0, nil
}

func (f *fakeCreditLedgerStore) ReconcileCreditLedger(context.Context) ([]credits.ReconciliationDiff, error) {
	f.reconciled++

	return f.diffs, nil
}

func TestCreditLedgerProcessorDrainsFullBatches(t *testing.T) {
	t.Parallel()

	store := &fakeCreditLedgerStore{settleResults: []credits.SettlementResult{
		{Locked: true, Scanned: 10, Settled: 4},
		{Locked: true, Scanned: 10, Settled: 2},
		{Locked: true, Scanned: 3, Settled: 1},
		{Locked: true, Scanned: 10},
	}}
	processor := NewCreditLedgerProcessor(CreditLedgerConfig{Enabled: true, BatchSize: 10}, store, nil)

	processor.Settle(context.Background())
	require.Equal(t, 3, store.settleCalls, "stops at the first partial batch")
}

func TestCreditLedgerProcessorStopsWhenLockedOrFailing(t *testing.T) {
	t.Parallel()

	locked := &fakeCreditLedgerStore{settleResults: []credits.SettlementResult{{Locked: false}}}
	NewCreditLedgerProcessor(CreditLedgerConfig{Enabled: true, BatchSize: 10}, locked, nil).Settle(context.Background())
	require.Equal(t, 1, locked.settleCalls)

	failing := &fakeCreditLedgerStore{settleErr: errors.New("boom")}
	NewCreditLedgerProcessor(CreditLedgerConfig{Enabled: true, BatchSize: 10}, failing, nil).Settle(context.Background())
	require.Equal(t, 1, failing.settleCalls)
}

func TestCreditLedgerProcessorExpiresAndReconciles(t *testing.T) {
	t.Parallel()

	store := &fakeCreditLedgerStore{diffs: []credits.ReconciliationDiff{{TenantID: "t", Check: credits.CheckDebt, ExpectedMC: 1, ActualMC: 2}}}
	processor := NewCreditLedgerProcessor(CreditLedgerConfig{Enabled: true}, store, nil)
	fixed := time.Date(2026, 10, 12, 12, 0, 0, 0, time.UTC)
	processor.now = func() time.Time { return fixed }

	processor.GrantAndExpire(context.Background())
	require.Equal(t, fixed, store.grantNow)
	require.Equal(t, fixed, store.expireNow)
	require.Equal(t, creditPartitionMonthsAhead, store.partitions)

	processor.Reconcile(context.Background())
	require.Equal(t, fixed.Add(-creditSweepLookback), store.sweepSince)
	require.Equal(t, 1, store.reconciled)
}

func TestCreditLedgerProcessorStartIsANoOpWhenDisabled(t *testing.T) {
	t.Parallel()

	store := &fakeCreditLedgerStore{}
	processor := NewCreditLedgerProcessor(CreditLedgerConfig{}, store, nil)

	processor.Start()
	processor.Stop()
	require.Zero(t, store.settleCalls)
}

func TestCreditLedgerProcessorRunsJobsUntilStopped(t *testing.T) {
	t.Parallel()

	store := &fakeCreditLedgerStore{}
	processor := NewCreditLedgerProcessor(CreditLedgerConfig{Enabled: true, SettlementInterval: time.Hour}, store, nil)

	processor.Start()
	processor.Stop()
	require.GreaterOrEqual(t, store.settleCalls, 1, "each job runs once at start")
}
