package billing

import (
	"context"
	"strconv"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/Delpak/shahrah/services/control-plane/internal/runtimeprofile"
	"github.com/Delpak/shahrah/services/control-plane/internal/testcache"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

type staticCreditPools struct {
	allowanceMC int64
	calls       int
}

func (s *staticCreditPools) CreditPoolAllowanceMC(context.Context, string, string) (int64, error) {
	s.calls++

	return s.allowanceMC, nil
}

type creditTestHarness struct {
	service *Service
	cache   *cache.MemoryCache
	pools   *staticCreditPools
	logs    *observer.ObservedLogs
}

func newCreditTestHarness(t *testing.T, allowanceMC int64, rules ...platform.BudgetLimitRule) creditTestHarness {
	t.Helper()

	mem, err := testcache.New(t, zap.NewNop())
	require.NoError(t, err)

	core, logs := observer.New(zap.WarnLevel)
	pools := &staticCreditPools{allowanceMC: allowanceMC}
	profile := &runtimeprofile.Profile{
		TenantID:    "tenant-1",
		Tenant:      platform.Tenant{ID: "tenant-1", BillingUnit: credits.BillingUnitCredit},
		BudgetRules: rules,
	}

	service := NewServiceWithDependencies(Dependencies{
		Cache:           mem,
		BudgetWindows:   mem,
		Wallets:         &budgetWindowWalletStore{wallet: &platform.Wallet{TenantID: "tenant-1", BalanceIRR: 1_000_000_000}},
		RuntimeProfiles: staticRuntimeProfiles{profile: profile},
		CreditPools:     pools,
		Logger:          zap.New(core),
	}, 850000, "test")

	return creditTestHarness{service: service, cache: mem, pools: pools, logs: logs}
}

// creditBreakdown prices a request at the given number of whole credits
// (1 credit = 0.002 USD at the fixed factor of 500).
func creditBreakdown(creditsAmount int) platform.UsageCostBreakdown {
	return platform.UsageCostBreakdown{
		Model: "gpt-test",
		// The IRR estimate is huge on purpose: credit tenants must never be
		// enforced in IRR.
		TotalCostIRR: 1e12,
		Components: []platform.UsageCostComponent{{
			Quantity: "1", UnitPriceUSD: "0.002", CostUSD: creditUSD(creditsAmount),
		}},
	}
}

// creditUSD is the provider USD cost of a whole number of credits.
func creditUSD(creditsAmount int) string {
	return decimal.NewFromInt(int64(creditsAmount)).Mul(decimal.RequireFromString("0.002")).String()
}

func creditUserRule(onExceed platform.BudgetOnExceed, limitCredits int64) platform.BudgetLimitRule {
	limitMC := limitCredits * credits.MillicreditsPerCredit

	return platform.BudgetLimitRule{
		RuleID:           "user-daily",
		Enabled:          true,
		Unit:             platform.BudgetLimitCostPerDay,
		BudgetAppliesPer: stringPtr(string(platform.BudgetAppliesPerUser)),
		LimitMC:          &limitMC,
		OnExceed:         onExceed,
		AlertThresholds:  []int{80, 100},
	}
}

var creditAuth = &platform.AuthContext{TenantID: "tenant-1", UserID: "user-1", TokenHash: "hash-1"}

func TestCreditTenantWithoutCreditPoolsIsRefused(t *testing.T) {
	mem, err := testcache.New(t, zap.NewNop())
	require.NoError(t, err)

	profile := &runtimeprofile.Profile{
		TenantID: "tenant-1",
		Tenant:   platform.Tenant{ID: "tenant-1", BillingUnit: credits.BillingUnitCredit},
	}
	service := NewServiceWithDependencies(Dependencies{
		Cache: mem, BudgetWindows: mem, RuntimeProfiles: staticRuntimeProfiles{profile: profile},
	}, 850000, "test")

	_, err = service.ReserveBudget(context.Background(), creditAuth, creditBreakdown(1), "req-1")
	require.ErrorIs(t, err, ErrBudgetUnavailable, "a credit tenant is never served unenforced")
}

func TestCreditTenantReservesMillicreditsFromItsBalance(t *testing.T) {
	h := newCreditTestHarness(t, 10_000)
	ctx := context.Background()

	reservation, err := h.service.ReserveBudget(ctx, creditAuth, creditBreakdown(6), "req-1")
	require.NoError(t, err)
	require.Equal(t, credits.BillingUnitCredit, reservation.Unit)
	require.Equal(t, int64(6_000), reservation.Targets[0].EstimatedMC)
	require.Equal(t, credits.PoolWindowKey("tenant-1", reservation.WindowID), reservation.WindowKey)
	require.Equal(t, 1, h.pools.calls, "the pool window is seeded from the credit balance")

	_, err = h.service.ReserveBudget(ctx, creditAuth, creditBreakdown(5), "req-2")
	require.ErrorIs(t, err, ErrBudgetExceeded)
	require.Equal(t, BudgetScopeTenantBalance, BudgetErrorScope(err))

	require.NoError(t, h.service.CommitBudget(ctx, reservation, creditBreakdown(4)))

	_, err = h.service.ReserveBudget(ctx, creditAuth, creditBreakdown(5), "req-3")
	require.NoError(t, err, "a lower actual cost frees the rest of the estimate")
}

func TestCreditCommitIsForcedForServedUsage(t *testing.T) {
	h := newCreditTestHarness(t, 10_000)
	ctx := context.Background()

	reservation, err := h.service.ReserveBudget(ctx, creditAuth, creditBreakdown(8), "req-1")
	require.NoError(t, err)

	require.NoError(t, h.service.CommitBudget(ctx, reservation, creditBreakdown(15)),
		"usage already served is committed even beyond the balance; the ledger records it as debt")

	snapshot, err := h.cache.GetBudgetWindow(ctx, reservation.WindowKey)
	require.NoError(t, err)
	require.InDelta(t, 15_000, snapshot.CommittedIRR, 0.001)
}

func TestCreditBlockRuleRefusesWithItsScope(t *testing.T) {
	h := newCreditTestHarness(t, 1_000_000, creditUserRule(platform.BudgetOnExceedBlock, 10))
	ctx := context.Background()

	reservation, err := h.service.ReserveBudget(ctx, creditAuth, creditBreakdown(7), "req-1")
	require.NoError(t, err)
	require.Len(t, reservation.Targets, 2)
	require.Equal(t, "credit_policy:", reservation.Targets[1].WindowKey[:len("credit_policy:")])

	_, err = h.service.ReserveBudget(ctx, creditAuth, creditBreakdown(4), "req-2")
	require.ErrorIs(t, err, ErrBudgetExceeded)
	require.Equal(t, string(platform.BudgetAppliesPerUser), BudgetErrorScope(err))

	other := &platform.AuthContext{TenantID: "tenant-1", UserID: "user-2", TokenHash: "hash-2"}
	_, err = h.service.ReserveBudget(ctx, other, creditBreakdown(4), "req-3")
	require.NoError(t, err, "the limit applies per user")
}

func TestCreditAlertRuleNeverBlocksAndAlertsOnce(t *testing.T) {
	h := newCreditTestHarness(t, 1_000_000, creditUserRule(platform.BudgetOnExceedAlert, 10))
	ctx := context.Background()

	for i, amount := range []int{9, 2, 3} {
		reservation, err := h.service.ReserveBudget(ctx, creditAuth, creditBreakdown(amount), "req-"+strconv.Itoa(i))
		require.NoError(t, err, "alert rules never refuse")
		require.NoError(t, h.service.CommitBudget(ctx, reservation, creditBreakdown(amount)))
	}

	alerts := h.logs.FilterMessage("budget_alert").AllUntimed()
	require.Len(t, alerts, 2, "80%% on the first request, 100%% on the second, nothing on the third")
	require.Equal(t, int64(80), alerts[0].ContextMap()["threshold_percent"])
	require.Equal(t, int64(100), alerts[1].ContextMap()["threshold_percent"])
}

func TestCreditExternalBilledReservesOnlyTheFeeFromTheBalance(t *testing.T) {
	h := newCreditTestHarness(t, 1_000, creditUserRule(platform.BudgetOnExceedBlock, 30))
	ctx := context.Background()

	reservation, err := h.service.ReserveExternalBilledBudget(ctx, creditAuth, creditBreakdown(20), "req-1")
	require.NoError(t, err)
	require.Equal(t, platform.BillingModeExternalBilled, reservation.BillingMode)
	require.Equal(t, int64(1_000), reservation.Targets[0].EstimatedMC, "5% of 20 credits is 1 credit against the balance")
	require.Equal(t, int64(20_000), reservation.Targets[1].EstimatedMC, "the rule counts the full equivalent")

	_, err = h.service.ReserveExternalBilledBudget(ctx, creditAuth, creditBreakdown(20), "req-2")
	require.ErrorIs(t, err, ErrBudgetExceeded)
}

func TestIRRTenantExternalBilledIsNotReserved(t *testing.T) {
	mem, err := testcache.New(t, zap.NewNop())
	require.NoError(t, err)

	service := newPolicyTestService(mem, staticRuntimeProfiles{profile: &runtimeprofile.Profile{TenantID: "tenant-1"}})

	reservation, err := service.ReserveExternalBilledBudget(context.Background(), creditAuth, creditBreakdown(20), "req-1")
	require.NoError(t, err)
	require.Nil(t, reservation)
}
