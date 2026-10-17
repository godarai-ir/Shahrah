package integration_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/Delpak/shahrah/services/control-plane/internal/runtimeprofile"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/Delpak/shahrah/services/control-plane/internal/testcache"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type creditProfileStore struct {
	profile *runtimeprofile.Profile
}

func (s creditProfileStore) GetSnapshot(context.Context, string) (*runtimeprofile.Profile, error) {
	return s.profile, nil
}

type creditEnforcementFixture struct {
	store    *storage.Store
	pool     *pgxpool.Pool
	cache    *cache.MemoryCache
	budgets  *billing.Service
	tenantID string
}

// newCreditEnforcementFixture wires a store and a budget service to one
// memory cache, so the store's window re-seeding is visible to reservations.
func newCreditEnforcementFixture(t *testing.T) creditEnforcementFixture {
	t.Helper()

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set")
	}

	ctx := context.Background()

	mem, err := testcache.New(t, zap.NewNop())
	require.NoError(t, err)

	store, err := storage.NewStore(ctx, databaseURL, zap.NewNop(), storage.StoreDependencies{Cache: mem})
	require.NoError(t, err)
	t.Cleanup(store.Close)

	pool, err := pgxpool.New(ctx, databaseURL)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	// A first pass creates the settlement cursor, which starts at the newest
	// spend log; logs inserted afterwards are then always settled.
	settleAllCredits(t, store)

	tenantID := newCreditLedgerTenant(t, pool, credits.BillingUnitCredit)
	profile := &runtimeprofile.Profile{
		TenantID: tenantID,
		Tenant:   platform.Tenant{ID: tenantID, BillingUnit: credits.BillingUnitCredit},
	}

	budgets := billing.NewServiceWithDependencies(billing.Dependencies{
		Cache:           mem,
		BudgetWindows:   mem,
		Wallets:         store.WalletRepository,
		RuntimeProfiles: creditProfileStore{profile: profile},
		CreditPools:     store.CreditLedgerRepository,
	}, 850000, "test")

	return creditEnforcementFixture{store: store, pool: pool, cache: mem, budgets: budgets, tenantID: tenantID}
}

func (f creditEnforcementFixture) auth() *platform.AuthContext {
	return &platform.AuthContext{TenantID: f.tenantID, UserID: "user-1", TokenHash: "hash-1"}
}

// wholeCredits prices a request at a whole number of credits.
func wholeCredits(amount int64) platform.UsageCostBreakdown {
	usd := decimal.NewFromInt(amount).Mul(decimal.RequireFromString("0.002")).String()

	return platform.UsageCostBreakdown{
		Model:        "openai/gpt-test",
		TotalCostIRR: 1e12,
		Components:   []platform.UsageCostComponent{{Quantity: "1", UnitPriceUSD: usd, CostUSD: usd}},
	}
}

func TestIntegration_CreditEnforcementFollowsTheLedger(t *testing.T) {
	f := newCreditEnforcementFixture(t)
	ctx := context.Background()

	grantCredits(t, f.store, f.tenantID, credits.LotSourcePromo, 10_000, nil)

	served, err := f.budgets.ReserveBudget(ctx, f.auth(), wholeCredits(6), "req-1")
	require.NoError(t, err)
	require.NoError(t, f.budgets.CommitBudget(ctx, served, wholeCredits(6)))

	_, err = f.budgets.ReserveBudget(ctx, f.auth(), wholeCredits(5), "req-2")
	require.ErrorIs(t, err, billing.ErrBudgetExceeded, "6 of 10 credits are committed")
	require.Equal(t, billing.BudgetScopeTenantBalance, billing.BudgetErrorScope(err))

	// The ledger settles the same 6 credits. The window is re-seeded with the
	// settled amount added back, so the 4 credits left do not change.
	insertCreditSpendLog(t, f.pool, f.tenantID, "req-1", 6_000, "")
	settleAllCredits(t, f.store)

	var settledMC int64
	require.NoError(t, f.pool.QueryRow(ctx, `
		SELECT settled_mc FROM credit_budget_window_settlements WHERE tenant_id = $1::uuid
	`, f.tenantID).Scan(&settledMC))
	require.Equal(t, int64(6_000), settledMC)

	snapshot, err := f.cache.GetBudgetWindow(ctx, served.WindowKey)
	require.NoError(t, err)
	require.InDelta(t, 10_000, snapshot.AllowanceIRR, 0.001, "4 available plus 6 settled")

	_, err = f.budgets.ReserveBudget(ctx, f.auth(), wholeCredits(5), "req-3")
	require.ErrorIs(t, err, billing.ErrBudgetExceeded, "settlement does not hand back spent credits")

	held, err := f.budgets.ReserveBudget(ctx, f.auth(), wholeCredits(4), "req-4")
	require.NoError(t, err)

	// A grant re-seeds the window at once.
	grantCredits(t, f.store, f.tenantID, credits.LotSourcePromo, 5_000, nil)

	_, err = f.budgets.ReserveBudget(ctx, f.auth(), wholeCredits(5), "req-5")
	require.NoError(t, err, "the new 5 credits are spendable while 4 stay held")

	require.NoError(t, f.budgets.ReleaseBudget(ctx, held, "test"))
	requireTenantReconciles(t, f.store, f.tenantID)
}

func TestIntegration_CreditPoolAllowanceIncludesOverageHeadroom(t *testing.T) {
	f := newCreditEnforcementFixture(t)
	ctx := context.Background()
	windowID, _ := credits.PoolWindow(time.Now())

	// Spend 3 credits with no balance: the ledger records 3 credits of debt.
	insertCreditSpendLog(t, f.pool, f.tenantID, "req-debt", 3_000, "")
	settleAllCredits(t, f.store)

	allowance, err := f.store.CreditPoolAllowanceMC(ctx, f.tenantID, windowID)
	require.NoError(t, err)
	require.Equal(t, int64(0), allowance, "debt only adds back what was settled; nothing is spendable")

	_, err = f.pool.Exec(ctx, `
		INSERT INTO tenant_entitlement_overrides (tenant_id, feature_key, value_json, reason, active)
		VALUES ($1::uuid, $2, 'true'::jsonb, 'test', true)
	`, f.tenantID, credits.FeatureOverageAllowed)
	require.NoError(t, err)

	_, err = f.pool.Exec(ctx, `
		INSERT INTO credit_tenant_settings (tenant_id, overage_enabled, overage_cap_mc) VALUES ($1::uuid, true, 10000)
	`, f.tenantID)
	require.NoError(t, err)

	allowance, err = f.store.CreditPoolAllowanceMC(ctx, f.tenantID, windowID)
	require.NoError(t, err)
	require.Equal(t, int64(7_000), allowance, "-3 owed + 3 settled + (10 cap - 3 owed) headroom: 7 more credits of debt reach the cap")
}

func TestIntegration_BudgetRulesStoreCreditFields(t *testing.T) {
	f := newCreditEnforcementFixture(t)
	ctx := context.Background()

	limit := "12.5"
	alert := platform.BudgetOnExceedAlert

	created, err := f.store.CreateBudgetLimitRule(ctx, f.tenantID, platform.CreateBudgetLimitRuleRequest{
		RuleID: "credit-rule", Unit: platform.BudgetLimitCostPerDay, Limit: &limit, OnExceed: &alert,
		AlertThresholds: []int{100, 50},
	})
	require.NoError(t, err)
	require.Equal(t, int64(12_500), *created.LimitMC)
	require.Equal(t, "12.50", *created.Limit)
	require.Equal(t, platform.BudgetOnExceedAlert, created.OnExceed)
	require.False(t, created.BlockOnBudgetExceed, "the legacy flag follows on_exceed")
	require.Equal(t, []int{50, 100}, created.AlertThresholds)

	notBlocking := false
	legacy, err := f.store.CreateBudgetLimitRule(ctx, f.tenantID, platform.CreateBudgetLimitRuleRequest{
		RuleID: "legacy-rule", Unit: platform.BudgetLimitCostPerMonth, LimitTo: 50_000, BlockOnBudgetExceed: &notBlocking,
	})
	require.NoError(t, err)
	require.Equal(t, platform.BudgetOnExceedAlert, legacy.OnExceed, "a non-blocking legacy rule keeps not blocking")
	require.Nil(t, legacy.LimitMC)
	require.Equal(t, []int{80, 100}, legacy.AlertThresholds)

	blocking := true
	_, err = f.store.UpdateBudgetLimitRuleByRuleID(ctx, f.tenantID, "credit-rule", platform.UpdateBudgetLimitRuleRequest{
		OnExceed: &alert, BlockOnBudgetExceed: &blocking,
	})
	require.ErrorIs(t, err, credits.ErrInconsistentOnExceed)

	downgrade := platform.BudgetOnExceed("downgrade")
	_, err = f.store.UpdateBudgetLimitRuleByRuleID(ctx, f.tenantID, "credit-rule", platform.UpdateBudgetLimitRuleRequest{OnExceed: &downgrade})
	require.ErrorIs(t, err, credits.ErrInvalidOnExceed, "downgrade is not an exceed behavior")

	cleared := ""
	updated, err := f.store.UpdateBudgetLimitRuleByRuleID(ctx, f.tenantID, "credit-rule", platform.UpdateBudgetLimitRuleRequest{Limit: &cleared})
	require.NoError(t, err)
	require.Nil(t, updated.LimitMC, "an empty limit clears it")
	require.Equal(t, platform.BudgetOnExceedAlert, updated.OnExceed, "omitted fields keep their value")
}

func TestIntegration_DirectWalletDebitSkipsCreditTenants(t *testing.T) {
	f := newCreditEnforcementFixture(t)
	ctx := context.Background()

	_, err := f.pool.Exec(ctx, `INSERT INTO wallets (tenant_id, balance_irr) VALUES ($1::uuid, 100000)`, f.tenantID)
	require.NoError(t, err)

	txn, err := f.store.DebitWallet(ctx, f.tenantID, platform.UsageCostBreakdown{Model: "m", TotalCostIRR: 5_000}, "req-direct")
	require.NoError(t, err)
	require.Nil(t, txn, "credit tenants are charged by the ledger, never the IRR wallet")

	var balanceIRR int64
	require.NoError(t, f.pool.QueryRow(ctx, `SELECT ROUND(balance_irr)::bigint FROM wallets WHERE tenant_id = $1::uuid`, f.tenantID).Scan(&balanceIRR))
	require.Equal(t, int64(100_000), balanceIRR)
}

func TestIntegration_GuardrailSideUsageSettlesAsItsOwnProduct(t *testing.T) {
	f := newCreditEnforcementFixture(t)
	ctx := context.Background()

	grantCredits(t, f.store, f.tenantID, credits.LotSourcePromo, 5_000, nil)

	require.NoError(t, f.store.LogSpend(ctx, platform.SpendLog{
		RequestID: "req-judge:guardrail_input", TenantID: f.tenantID, UserID: "", Model: "jev", Status: "success",
		RequestMetadata: map[string]interface{}{"trace_id": "trace-judge"},
		Credits:         &platform.SpendLogCredits{Millicredits: credits.GuardrailJudgeCallMC, Product: platform.CreditProductGuardrail, Source: string(credits.SourceFixedPrice)},
	}))

	_, err := f.pool.Exec(ctx, `UPDATE spend_logs SET created_at = NOW() - INTERVAL '5 minutes' WHERE request_id = 'req-judge:guardrail_input'`)
	require.NoError(t, err)

	settleAllCredits(t, f.store)

	var (
		product string
		traceID string
		amount  int64
	)

	require.NoError(t, f.pool.QueryRow(ctx, `
		SELECT product, COALESCE(trace_id, ''), amount_mc FROM credit_ledger WHERE tenant_id = $1::uuid AND kind = 'consume'
	`, f.tenantID).Scan(&product, &traceID, &amount))
	require.Equal(t, platform.CreditProductGuardrail, product)
	require.Equal(t, "trace-judge", traceID, "the side call links to the request it served")
	require.Equal(t, -credits.GuardrailJudgeCallMC, amount)

	balance, err := f.store.GetCreditBalance(ctx, f.tenantID)
	require.NoError(t, err)
	require.Equal(t, int64(4_000), balance.AvailableMC)
}
