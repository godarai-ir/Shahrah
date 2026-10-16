package integration_test

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// newCreditLedgerTenant creates a tenant billed in credits. Ledger rows are
// append-only, so these tenants are never deleted; the test database is disposable.
func newCreditLedgerTenant(t *testing.T, pool *pgxpool.Pool, unit string) string {
	t.Helper()

	tenantID := uuid.NewString()
	slug := "credit-" + strings.ReplaceAll(tenantID, "-", "")[:12]

	_, err := pool.Exec(context.Background(), `
		INSERT INTO tenants (tenant_id, slug, domain, status, billing_unit) VALUES ($1::uuid, $2, $2, 'active', $3)
	`, tenantID, slug, unit)
	require.NoError(t, err)

	return tenantID
}

func insertCreditSpendLog(t *testing.T, pool *pgxpool.Pool, tenantID, requestID string, creditsMC int64, billingMode string) int64 {
	t.Helper()

	var id int64

	err := pool.QueryRow(context.Background(), `
		INSERT INTO spend_logs (request_id, tenant_id, model, cost_usd, status, credits_mc, rate_card_version,
		                        product, billing_mode, credit_source, created_at, pricing_snapshot_json)
		VALUES ($1, $2::uuid, 'openai/gpt-test', 0.002, 'success', $3, 1, 'gateway', NULLIF($4, ''), 'components',
		        NOW() - INTERVAL '5 minutes', '{"revision":"rev-7"}'::jsonb)
		RETURNING id
	`, requestID, tenantID, creditsMC, billingMode).Scan(&id)
	require.NoError(t, err)

	return id
}

func settleAllCredits(t *testing.T, store *storage.Store) {
	t.Helper()

	for range 50 {
		result, err := store.SettleCreditConsumption(context.Background(), 500, time.Second)
		require.NoError(t, err)

		if result.Locked && result.Scanned < 500 {
			return
		}
	}

	t.Fatal("credit settlement did not drain")
}

func grantCredits(t *testing.T, store *storage.Store, tenantID string, source credits.LotSource, amountMC int64, expiresAt *time.Time) storage.CreditGrantResult {
	t.Helper()

	kind := credits.EntryPromoGrant

	switch source {
	case credits.LotSourcePlan:
		kind = credits.EntryPlanGrant
	case credits.LotSourcePurchaseAddon:
		kind = credits.EntryPurchase
	}

	result, err := store.GrantCredits(context.Background(), storage.CreditGrant{
		TenantID: tenantID, Kind: kind, Source: source, AmountMC: amountMC, ExpiresAt: expiresAt,
		IdempotencyKey: "test:" + uuid.NewString(),
	})
	require.NoError(t, err)

	return result
}

func lotRemaining(t *testing.T, pool *pgxpool.Pool, lotID string) int64 {
	t.Helper()

	var remaining int64
	require.NoError(t, pool.QueryRow(context.Background(), `SELECT remaining_mc FROM credit_lot_balances WHERE lot_id = $1::uuid`, lotID).Scan(&remaining))

	return remaining
}

func requireTenantReconciles(t *testing.T, store *storage.Store, tenantID string) {
	t.Helper()

	diffs, err := store.ReconcileCreditLedger(context.Background())
	require.NoError(t, err)

	for _, diff := range diffs {
		require.NotEqual(t, tenantID, diff.TenantID, "reconciliation difference: %+v", diff)
		require.NotEqual(t, credits.CheckDefaultPartition, diff.Check, "rows landed in a DEFAULT partition")
	}
}

func TestIntegration_CreditLedgerConsumesInOrderRepaysDebtAndReconciles(t *testing.T) {
	store, pool := newCreditIntegrationStore(t)
	ctx := context.Background()

	settleAllCredits(t, store) // initializes the cursor before this test's spend logs

	tenantID := newCreditLedgerTenant(t, pool, credits.BillingUnitCredit)
	monthEnd := time.Now().UTC().AddDate(0, 1, 0).Truncate(time.Second)
	yearEnd := time.Now().UTC().AddDate(1, 0, 0).Truncate(time.Second)

	addon := grantCredits(t, store, tenantID, credits.LotSourcePurchaseAddon, 5_000, &monthEnd)
	promo := grantCredits(t, store, tenantID, credits.LotSourcePromo, 300, &monthEnd)
	plan := grantCredits(t, store, tenantID, credits.LotSourcePlan, 1_000, &monthEnd)
	later := grantCredits(t, store, tenantID, credits.LotSourcePlan, 2_000, &yearEnd)

	insertCreditSpendLog(t, pool, tenantID, "req-"+uuid.NewString(), 1_200, "")
	settleAllCredits(t, store)

	require.Zero(t, lotRemaining(t, pool, plan.LotID), "plan credits go first")
	require.Equal(t, int64(100), lotRemaining(t, pool, promo.LotID), "then promo")
	require.Equal(t, int64(5_000), lotRemaining(t, pool, addon.LotID), "add-ons last among equal expiry")
	require.Equal(t, int64(2_000), lotRemaining(t, pool, later.LotID), "later expiry is untouched")

	insertCreditSpendLog(t, pool, tenantID, "req-"+uuid.NewString(), 7_350, "")
	settleAllCredits(t, store)

	balance, err := store.GetCreditBalance(ctx, tenantID)
	require.NoError(t, err)
	require.Zero(t, balance.AvailableMC)
	require.Equal(t, int64(250), balance.DebtMC)
	require.Equal(t, int64(-250), balance.LedgerSumMC)

	repay := grantCredits(t, store, tenantID, credits.LotSourcePlan, 1_000, &yearEnd)
	require.Equal(t, int64(250), repay.RepaidDebtMC)
	require.Equal(t, int64(750), lotRemaining(t, pool, repay.LotID))

	balance, err = store.GetCreditBalance(ctx, tenantID)
	require.NoError(t, err)
	require.Equal(t, int64(750), balance.AvailableMC)
	require.Zero(t, balance.DebtMC)

	requireTenantReconciles(t, store, tenantID)
}

func TestIntegration_CreditLedgerChargesBYOKFeeAndIgnoresIRRTenants(t *testing.T) {
	store, pool := newCreditIntegrationStore(t)
	ctx := context.Background()

	settleAllCredits(t, store)

	tenantID := newCreditLedgerTenant(t, pool, credits.BillingUnitCredit)
	irrTenantID := newCreditLedgerTenant(t, pool, credits.BillingUnitIRR)
	grantCredits(t, store, tenantID, credits.LotSourcePlan, 10_000, nil)

	byokRequest := "req-" + uuid.NewString()
	insertCreditSpendLog(t, pool, tenantID, byokRequest, 4_000, "external_billed")
	insertCreditSpendLog(t, pool, irrTenantID, "req-"+uuid.NewString(), 4_000, "")
	settleAllCredits(t, store)

	var (
		amountMC, equivalentMC int64
		revision, product      string
	)

	require.NoError(t, pool.QueryRow(ctx, `
		SELECT amount_mc, credit_equivalent_mc, pricing_revision, product
		FROM credit_ledger WHERE tenant_id = $1::uuid AND kind = 'consume'
	`, tenantID).Scan(&amountMC, &equivalentMC, &revision, &product))
	require.Equal(t, int64(-200), amountMC, "5% platform fee")
	require.Equal(t, int64(4_000), equivalentMC)
	require.Equal(t, "rev-7", revision)
	require.Equal(t, "gateway", product)

	var irrEntries int
	require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(1) FROM credit_ledger WHERE tenant_id = $1::uuid`, irrTenantID).Scan(&irrEntries))
	require.Zero(t, irrEntries, "IRR tenants never touch the credit ledger")

	requireTenantReconciles(t, store, tenantID)
}

func TestIntegration_CreditLedgerIsIdempotentUnderConcurrency(t *testing.T) {
	store, pool := newCreditIntegrationStore(t)
	ctx := context.Background()

	settleAllCredits(t, store)

	tenantID := newCreditLedgerTenant(t, pool, credits.BillingUnitCredit)
	key := "plan_grant:test:" + uuid.NewString()
	expired := time.Now().UTC().Add(-time.Minute).Truncate(time.Second)

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		lotIDs  = map[string]bool{}
		grants  int
		grantMu sync.Mutex
	)

	for range 8 {
		wg.Go(func() {
			result, err := store.GrantCredits(ctx, storage.CreditGrant{
				TenantID: tenantID, Kind: credits.EntryPlanGrant, Source: credits.LotSourcePlan,
				AmountMC: 900, ExpiresAt: &expired, IdempotencyKey: key,
			})
			if !assertNoError(t, err) {
				return
			}

			mu.Lock()
			lotIDs[result.LotID] = true
			mu.Unlock()

			if !result.Duplicate {
				grantMu.Lock()
				grants++
				grantMu.Unlock()
			}
		})
	}

	wg.Wait()
	require.Equal(t, 1, grants, "one grant wins the key")
	require.Len(t, lotIDs, 1, "duplicates report the winning lot")

	request := "req-" + uuid.NewString()
	insertCreditSpendLog(t, pool, tenantID, request, 100, "")
	insertCreditSpendLog(t, pool, tenantID, request, 100, "") // same request and product: skipped

	for range 4 {
		wg.Go(func() {
			_, err := store.SettleCreditConsumption(ctx, 500, time.Second)
			assertNoError(t, err)
		})
	}

	wg.Wait()
	settleAllCredits(t, store)

	for range 4 {
		wg.Go(func() {
			_, err := store.ExpireCreditLots(ctx, time.Now().UTC(), 0)
			assertNoError(t, err)
		})
	}

	wg.Wait()

	var consumes, expires int
	require.NoError(t, pool.QueryRow(ctx, `
		SELECT COUNT(1) FILTER (WHERE kind = 'consume'), COUNT(1) FILTER (WHERE kind = 'expire')
		FROM credit_ledger WHERE tenant_id = $1::uuid
	`, tenantID).Scan(&consumes, &expires))
	require.Equal(t, 1, consumes)
	require.Equal(t, 1, expires)

	balance, err := store.GetCreditBalance(ctx, tenantID)
	require.NoError(t, err)
	require.Zero(t, balance.AvailableMC)
	require.Zero(t, balance.DebtMC)

	_, err = store.SweepCreditConsumption(ctx, time.Now().Add(-time.Hour))
	require.NoError(t, err)
	requireTenantReconciles(t, store, tenantID)
}

func assertNoError(t *testing.T, err error) bool {
	t.Helper()

	if err != nil {
		t.Errorf("unexpected error: %v", err)

		return false
	}

	return true
}

func TestIntegration_CreditLedgerTablesAreAppendOnly(t *testing.T) {
	store, pool := newCreditIntegrationStore(t)
	ctx := context.Background()

	tenantID := newCreditLedgerTenant(t, pool, credits.BillingUnitCredit)
	grantCredits(t, store, tenantID, credits.LotSourcePromo, 100, nil)

	for _, statement := range []string{
		`UPDATE credit_ledger SET amount_mc = 1 WHERE tenant_id = $1::uuid`,
		`DELETE FROM credit_ledger WHERE tenant_id = $1::uuid`,
		`UPDATE credit_ledger_allocations SET amount_mc = 1 WHERE tenant_id = $1::uuid`,
		`DELETE FROM credit_ledger_allocations WHERE tenant_id = $1::uuid`,
		`DELETE FROM credit_ledger_keys WHERE tenant_id = $1::uuid`,
		`UPDATE credit_lots SET amount_mc = 1 WHERE tenant_id = $1::uuid`,
		`DELETE FROM credit_lots WHERE tenant_id = $1::uuid`,
	} {
		_, err := pool.Exec(ctx, statement, tenantID)
		require.ErrorContains(t, err, "append-only", statement)
	}

	for _, table := range []string{"credit_ledger", "credit_ledger_allocations", "credit_ledger_keys", "credit_lots"} {
		_, err := pool.Exec(ctx, `TRUNCATE `+table+` CASCADE`)
		require.ErrorContains(t, err, "append-only", table)
	}

	_, err := store.GrantCredits(ctx, storage.CreditGrant{
		TenantID: tenantID, Kind: credits.EntryAdjust, Source: credits.LotSourceAdjust, AmountMC: 10, IdempotencyKey: "adjust:" + uuid.NewString(),
	})
	require.ErrorIs(t, err, storage.ErrInvalidCreditGrant, "adjustments need a reason")

	created, err := store.EnsureCreditLedgerPartitions(ctx, time.Now(), 3)
	require.NoError(t, err)
	require.Zero(t, created, "the migration already created the partitions")
}

func TestIntegration_WalletSettlementSkipsCreditTenants(t *testing.T) {
	store, pool := newCreditIntegrationStore(t)
	ctx := context.Background()

	for _, unit := range []string{credits.BillingUnitCredit, credits.BillingUnitIRR} {
		tenantID := newCreditLedgerTenant(t, pool, unit)
		windowKey := "budget:tenant:" + tenantID + ":window:2026-10"

		_, err := pool.Exec(ctx, `INSERT INTO wallets (tenant_id, balance_irr) VALUES ($1::uuid, 100000)`, tenantID)
		require.NoError(t, err)

		require.NoError(t, store.ApplyWalletBudgetSettlement(ctx, cache.BudgetWindowSnapshot{WindowKey: windowKey, CommittedIRR: 12_000}))

		var (
			settledIRR   int64
			transactions int
		)

		require.NoError(t, pool.QueryRow(ctx, `
			SELECT ROUND(settled_irr)::bigint FROM wallet_budget_window_settlements WHERE tenant_id = $1::uuid
		`, tenantID).Scan(&settledIRR))
		require.NoError(t, pool.QueryRow(ctx, `
			SELECT COUNT(1) FROM wallet_transactions WHERE tenant_id = $1::uuid AND source_type = 'budget_window'
		`, tenantID).Scan(&transactions))

		require.Equal(t, int64(12_000), settledIRR, unit)

		if unit == credits.BillingUnitCredit {
			require.Zero(t, transactions, "credit tenants are charged by the credit ledger only")
		} else {
			require.Equal(t, 1, transactions, "IRR tenants keep wallet settlement")
		}
	}
}

func TestIntegration_CreditSweepRecoversSpendLogsBehindTheCursor(t *testing.T) {
	store, pool := newCreditIntegrationStore(t)
	ctx := context.Background()

	settleAllCredits(t, store)

	tenantID := newCreditLedgerTenant(t, pool, credits.BillingUnitCredit)
	grantCredits(t, store, tenantID, credits.LotSourcePlan, 1_000, nil)
	spendLogID := insertCreditSpendLog(t, pool, tenantID, "req-"+uuid.NewString(), 120, "")

	// Simulate a writer whose row committed after the cursor had already moved past it.
	_, err := pool.Exec(ctx, `UPDATE credit_settlement_cursors SET last_spend_log_id = $1 WHERE cursor_id = 'spend_logs'`, spendLogID)
	require.NoError(t, err)
	settleAllCredits(t, store)

	balance, err := store.GetCreditBalance(ctx, tenantID)
	require.NoError(t, err)
	require.Equal(t, int64(1_000), balance.AvailableMC, "the cursor skipped the row")

	sweep, err := store.SweepCreditConsumption(ctx, time.Now().Add(-time.Hour))
	require.NoError(t, err)
	require.GreaterOrEqual(t, sweep.Settled, 1)

	balance, err = store.GetCreditBalance(ctx, tenantID)
	require.NoError(t, err)
	require.Equal(t, int64(880), balance.AvailableMC)

	again, err := store.SweepCreditConsumption(ctx, time.Now().Add(-time.Hour))
	require.NoError(t, err)
	require.Zero(t, again.Settled, "a second sweep finds nothing new")
	requireTenantReconciles(t, store, tenantID)
}
