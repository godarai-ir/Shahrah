package billing

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"go.uber.org/zap"
)

type walletSettlementListerStub struct {
	snapshots []cache.BudgetWindowSnapshot
}

func (s walletSettlementListerStub) ListActiveBudgetWindows(context.Context, int) ([]cache.BudgetWindowSnapshot, error) {
	return s.snapshots, nil
}

type walletSettlementStoreStub struct {
	mu       sync.Mutex
	applied  []cache.BudgetWindowSnapshot
}

func (s *walletSettlementStoreStub) ApplyWalletBudgetSettlement(_ context.Context, snapshot cache.BudgetWindowSnapshot) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.applied = append(s.applied, snapshot)
	return nil
}

func TestWalletSettlementProcessorProcessesActiveWindows(t *testing.T) {
	t.Parallel()

	liveCache := walletSettlementListerStub{
		snapshots: []cache.BudgetWindowSnapshot{
			{WindowKey: "budget:window:tenant-1:2026-07", CommittedIRR: 4200, LastUpdatedAt: time.Now().UTC()},
			{WindowKey: "budget:window:tenant-2:2026-07", CommittedIRR: 1800, LastUpdatedAt: time.Now().UTC()},
		},
	}
	store := &walletSettlementStoreStub{}
	processor := NewWalletSettlementProcessor(
		WalletSettlementConfig{
			Enabled:      true,
			Interval:     time.Second,
			Timeout:      time.Second,
			BatchSize:    10,
		},
		liveCache,
		store,
		zap.NewNop(),
	)

	processor.processWithTimeout(context.Background())

	if got := len(store.applied); got != 2 {
		t.Fatalf("expected 2 settlement applications, got %d", got)
	}
	if store.applied[0].WindowKey != "budget:window:tenant-1:2026-07" {
		t.Fatalf("unexpected first window key: %s", store.applied[0].WindowKey)
	}
	if store.applied[1].WindowKey != "budget:window:tenant-2:2026-07" {
		t.Fatalf("unexpected second window key: %s", store.applied[1].WindowKey)
	}
}
