package billing

import (
	"context"
	"errors"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"go.uber.org/zap"
)

func TestReserveBudgetSeedsWalletOnceThenUsesAllowanceWindow(t *testing.T) {
	mem, err := cache.NewMemoryCache(zap.NewNop())
	if err != nil {
		t.Fatalf("NewMemoryCache returned error: %v", err)
	}
	store := &budgetWindowWalletStore{wallet: &platform.Wallet{TenantID: "tenant-1", BalanceIRR: 100}}
	service := NewService(mem, store, 850000, "test")

	auth := &platform.AuthContext{TenantID: "tenant-1", TokenHash: "hash-1"}
	first, err := service.ReserveBudget(context.Background(), auth, platform.UsageCostBreakdown{TotalCostIRR: 25}, "req-1")
	if err != nil {
		t.Fatalf("first ReserveBudget returned error: %v", err)
	}
	if first == nil || first.ID != "req-1" {
		t.Fatalf("unexpected reservation: %+v", first)
	}
	if store.getWalletCalls != 1 {
		t.Fatalf("expected one wallet seed call, got %d", store.getWalletCalls)
	}

	second, err := service.ReserveBudget(context.Background(), auth, platform.UsageCostBreakdown{TotalCostIRR: 25}, "req-2")
	if err != nil {
		t.Fatalf("second ReserveBudget returned error: %v", err)
	}
	if second == nil || second.ID != "req-2" {
		t.Fatalf("unexpected second reservation: %+v", second)
	}
	if store.getWalletCalls != 1 {
		t.Fatalf("expected hot reserve to avoid wallet store, got %d calls", store.getWalletCalls)
	}

	if err := service.CommitBudget(context.Background(), first, platform.UsageCostBreakdown{TotalCostIRR: 10}); err != nil {
		t.Fatalf("CommitBudget returned error: %v", err)
	}
	if err := service.ReleaseBudget(context.Background(), second, "test"); err != nil {
		t.Fatalf("ReleaseBudget returned error: %v", err)
	}
}

func TestReserveBudgetReturnsBudgetExceededWhenAllowanceExhausted(t *testing.T) {
	mem, err := cache.NewMemoryCache(zap.NewNop())
	if err != nil {
		t.Fatalf("NewMemoryCache returned error: %v", err)
	}
	service := NewService(mem, &budgetWindowWalletStore{wallet: &platform.Wallet{TenantID: "tenant-1", BalanceIRR: 10}}, 850000, "test")

	_, err = service.ReserveBudget(context.Background(), &platform.AuthContext{TenantID: "tenant-1"}, platform.UsageCostBreakdown{TotalCostIRR: 11}, "req-1")
	if !errors.Is(err, ErrBudgetExceeded) {
		t.Fatalf("expected ErrBudgetExceeded, got %v", err)
	}
}

type budgetWindowWalletStore struct {
	wallet         *platform.Wallet
	getWalletCalls int
}

func (s *budgetWindowWalletStore) GetWallet(context.Context, string) (*platform.Wallet, error) {
	s.getWalletCalls++
	out := *s.wallet
	return &out, nil
}

func (s *budgetWindowWalletStore) GetLatestExchangeRate(context.Context) (*platform.ExchangeRate, error) {
	return nil, errors.New("not implemented")
}

func (s *budgetWindowWalletStore) UpsertExchangeRate(context.Context, platform.ExchangeRate) (*platform.ExchangeRate, error) {
	return nil, errors.New("not implemented")
}
