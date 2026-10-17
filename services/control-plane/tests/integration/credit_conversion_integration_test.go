package integration_test

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	billingdomain "github.com/Delpak/shahrah/services/control-plane/internal/domain/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

var conversionActor = platform.SupportActor{ID: "operator-1", Role: "support_admin"}

// testCreditPlan creates a private plan whose old version the tenant is on and
// whose published version grants credits, so conversion tests never touch the
// shared catalog.
func testCreditPlan(t *testing.T, pool *pgxpool.Pool) (planID, oldVersionID, creditVersionID string) {
	t.Helper()

	ctx := context.Background()
	slug := "conv-" + strings.ReplaceAll(uuid.NewString(), "-", "")[:10]

	require.NoError(t, pool.QueryRow(ctx, `
		INSERT INTO commercial_plans (slug, display_name, description, sort_order, active)
		VALUES ($1, $1, 'conversion test', 900, true) RETURNING plan_id::text
	`, slug).Scan(&planID))
	require.NoError(t, pool.QueryRow(ctx, `
		INSERT INTO plan_versions (plan_id, version, status) VALUES ($1::uuid, 1, 'archived') RETURNING version_id::text
	`, planID).Scan(&oldVersionID))
	require.NoError(t, pool.QueryRow(ctx, `
		INSERT INTO plan_versions (plan_id, version, status) VALUES ($1::uuid, 2, 'published') RETURNING version_id::text
	`, planID).Scan(&creditVersionID))

	_, err := pool.Exec(ctx, `
		INSERT INTO plan_entitlements (version_id, feature_key, value_json) VALUES ($1::uuid, 'credits.monthly_grant', '100'::jsonb)
	`, creditVersionID)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		INSERT INTO plan_prices (version_id, billing_interval, currency, amount_minor, discount_percent)
		VALUES ($1::uuid, 'yearly', 'IRR', 540000000, 10)
	`, creditVersionID)
	require.NoError(t, err)

	// A published credit version is repriced by every approved list price, so
	// leaving it published would leak into the price policy test.
	t.Cleanup(func() {
		_, cleanupErr := pool.Exec(context.Background(), `
			UPDATE plan_versions SET status = 'archived' WHERE plan_id = $1::uuid AND status <> 'archived'
		`, planID)
		require.NoError(t, cleanupErr)
	})

	return planID, oldVersionID, creditVersionID
}

// newConvertibleTenant is an IRR tenant with a wallet, an IRR budget rule, and
// a subscription on the old plan version.
func newConvertibleTenant(t *testing.T, f creditEnforcementFixture, walletIRR int64) (tenantID, subscriptionID, creditVersionID string) {
	t.Helper()

	ctx := context.Background()
	tenantID = newCreditLedgerTenant(t, f.pool, credits.BillingUnitIRR)

	_, err := f.pool.Exec(ctx, `INSERT INTO wallets (tenant_id, balance_irr) VALUES ($1::uuid, $2)`, tenantID, walletIRR)
	require.NoError(t, err)

	_, err = f.store.CreateBudgetLimitRule(ctx, tenantID, platform.CreateBudgetLimitRuleRequest{
		RuleID: "irr-rule", Unit: platform.BudgetLimitCostPerMonth, LimitTo: 10_001,
	})
	require.NoError(t, err)

	planID, oldVersionID, creditVersionID := testCreditPlan(t, f.pool)
	now := time.Now().UTC()
	subscriptionID = subscribeTenant(t, f.pool, tenantID, planID, oldVersionID, "active", now.AddDate(0, -1, 0), now.AddDate(0, 11, 0))

	return tenantID, subscriptionID, creditVersionID
}

func walletBalance(t *testing.T, pool *pgxpool.Pool, tenantID string) int64 {
	t.Helper()

	var balance int64
	require.NoError(t, pool.QueryRow(context.Background(), `
		SELECT ROUND(balance_irr)::bigint FROM wallets WHERE tenant_id = $1::uuid
	`, tenantID).Scan(&balance))

	return balance
}

func TestIntegration_CreditConversionMovesWalletRulesAndPlan(t *testing.T) {
	f := newCreditEnforcementFixture(t)
	ctx := context.Background()
	tenantID, oldSubscriptionID, creditVersionID := newConvertibleTenant(t, f, 123_456)

	preview, err := f.store.PreviewCreditConversion(ctx, tenantID)
	require.NoError(t, err)
	require.Equal(t, "12.00", preview.Credits)
	require.Equal(t, int64(3_456), preview.ResidueIRR)
	require.Equal(t, 1, preview.RulesToConvert)
	require.Equal(t, "100.00", preview.NewPlanCredits)
	require.Empty(t, preview.Warnings)

	conversion, err := f.store.ConvertTenantToCredits(ctx, tenantID, "fleet cutover", conversionActor)
	require.NoError(t, err)
	require.Equal(t, credits.ConversionToCredit, conversion.Direction)
	require.Equal(t, "12.00", conversion.Credits)
	require.Equal(t, int64(120_000), conversion.ConvertedIRR)
	require.Equal(t, 1, conversion.RulesConverted)

	require.Equal(t, int64(3_456), walletBalance(t, f.pool, tenantID), "the residue below one credit stays in the wallet")

	balance, err := f.store.GetCreditBalance(ctx, tenantID)
	require.NoError(t, err)
	require.Equal(t, int64(12_000), balance.AvailableMC)

	var (
		lotSource string
		expires   *time.Time
	)
	require.NoError(t, f.pool.QueryRow(ctx, `SELECT source, expires_at FROM credit_lots WHERE tenant_id = $1::uuid`, tenantID).Scan(&lotSource, &expires))
	require.Equal(t, string(credits.LotSourceMigration), lotSource)
	require.Nil(t, expires, "converted value never expires")

	var limitMC int64
	require.NoError(t, f.pool.QueryRow(ctx, `SELECT limit_mc FROM budget_limit_rules WHERE tenant_id = $1::uuid`, tenantID).Scan(&limitMC))
	require.Equal(t, int64(2_000), limitMC, "10,001 IRR rounds up to 2 credits")

	var unit, oldStatus, newVersion, newStatus string
	require.NoError(t, f.pool.QueryRow(ctx, `SELECT billing_unit FROM tenants WHERE tenant_id = $1::uuid`, tenantID).Scan(&unit))
	require.NoError(t, f.pool.QueryRow(ctx, `SELECT status FROM tenant_subscriptions WHERE subscription_id = $1::uuid`, oldSubscriptionID).Scan(&oldStatus))
	require.NoError(t, f.pool.QueryRow(ctx, `
		SELECT version_id::text, status FROM tenant_subscriptions WHERE subscription_id = $1::uuid
	`, conversion.SubscriptionID).Scan(&newVersion, &newStatus))
	require.Equal(t, credits.BillingUnitCredit, unit)
	require.Equal(t, "canceled", oldStatus)
	require.Equal(t, creditVersionID, newVersion)
	require.Equal(t, "active", newStatus)

	var audits int
	require.NoError(t, f.pool.QueryRow(ctx, `
		SELECT COUNT(1) FROM financial_audit_events WHERE tenant_id = $1::uuid AND event_type = 'credit.converted'
	`, tenantID).Scan(&audits))
	require.Equal(t, 1, audits)

	_, err = f.store.ConvertTenantToCredits(ctx, tenantID, "again", conversionActor)
	require.ErrorIs(t, err, billingdomain.ErrCreditConversionNothingToDo, "the residue is not a whole credit")

	requireTenantReconciles(t, f.store, tenantID)
}

func TestIntegration_CreditConversionIsOncePerTenantUnderConcurrency(t *testing.T) {
	f := newCreditEnforcementFixture(t)
	tenantID, _, _ := newConvertibleTenant(t, f, 100_000)

	var (
		wg        sync.WaitGroup
		mu        sync.Mutex
		succeeded int
	)

	for range 6 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			_, err := f.store.ConvertTenantToCredits(context.Background(), tenantID, "parallel", conversionActor)

			mu.Lock()
			defer mu.Unlock()

			if err == nil {
				succeeded++
			} else {
				require.ErrorIs(t, err, billingdomain.ErrCreditConversionNothingToDo)
			}
		}()
	}

	wg.Wait()
	require.Equal(t, 1, succeeded)

	balance, err := f.store.GetCreditBalance(context.Background(), tenantID)
	require.NoError(t, err)
	require.Equal(t, int64(10_000), balance.AvailableMC, "the wallet is converted once")
	require.Zero(t, walletBalance(t, f.pool, tenantID))
}

func TestIntegration_CreditConversionPicksUpLateTopUps(t *testing.T) {
	f := newCreditEnforcementFixture(t)
	ctx := context.Background()
	tenantID, _, _ := newConvertibleTenant(t, f, 20_000)

	_, err := f.store.ConvertTenantToCredits(ctx, tenantID, "cutover", conversionActor)
	require.NoError(t, err)

	// A top-up that was pending at the cutover completes into the IRR wallet.
	_, err = f.store.TopUpWallet(ctx, tenantID, 50_000, map[string]interface{}{"source": "late"})
	require.NoError(t, err)

	late, err := f.store.ConvertTenantToCredits(ctx, tenantID, "late top-up", conversionActor)
	require.NoError(t, err)
	require.Equal(t, 2, late.Sequence)
	require.Equal(t, "5.00", late.Credits)
	require.Empty(t, late.SubscriptionID, "the plan moves only once")

	balance, err := f.store.GetCreditBalance(ctx, tenantID)
	require.NoError(t, err)
	require.Equal(t, int64(7_000), balance.AvailableMC)
}

func TestIntegration_CreditRollbackReturnsCreditsToTheWallet(t *testing.T) {
	f := newCreditEnforcementFixture(t)
	ctx := context.Background()
	tenantID, _, _ := newConvertibleTenant(t, f, 120_000)

	_, err := f.store.ConvertTenantToCredits(ctx, tenantID, "cutover", conversionActor)
	require.NoError(t, err)

	insertCreditSpendLog(t, f.pool, tenantID, "req-rollback", 5_000, "")
	settleAllCredits(t, f.store)

	rollback, err := f.store.RollbackCreditConversion(ctx, tenantID, "incident", conversionActor)
	require.NoError(t, err)
	require.Equal(t, credits.ConversionToIRR, rollback.Direction)
	require.Equal(t, "7.00", rollback.Credits)
	require.Equal(t, int64(70_000), rollback.ConvertedIRR)
	require.Equal(t, int64(70_000), walletBalance(t, f.pool, tenantID))

	balance, err := f.store.GetCreditBalance(ctx, tenantID)
	require.NoError(t, err)
	require.Zero(t, balance.AvailableMC)
	require.Zero(t, balance.DebtMC)

	var unit string
	require.NoError(t, f.pool.QueryRow(ctx, `SELECT billing_unit FROM tenants WHERE tenant_id = $1::uuid`, tenantID).Scan(&unit))
	require.Equal(t, credits.BillingUnitIRR, unit)

	requireTenantReconciles(t, f.store, tenantID)
}

func TestIntegration_CreditRollbackCollectsDebtFromTheWallet(t *testing.T) {
	f := newCreditEnforcementFixture(t)
	ctx := context.Background()
	tenantID, _, _ := newConvertibleTenant(t, f, 23_456)

	_, err := f.store.ConvertTenantToCredits(ctx, tenantID, "cutover", conversionActor)
	require.NoError(t, err)

	// 2 credits converted, 3,456 IRR residue left in the wallet; 10 credits used.
	insertCreditSpendLog(t, f.pool, tenantID, "req-debt", 10_000, "")
	settleAllCredits(t, f.store)

	rollback, err := f.store.RollbackCreditConversion(ctx, tenantID, "incident", conversionActor)
	require.NoError(t, err)
	require.Equal(t, "8.00", rollback.Debt)
	require.Equal(t, int64(3_456), rollback.ConvertedIRR, "the wallet pays what it can")
	require.Equal(t, int64(76_544), rollback.UncollectedIRR)
	require.Zero(t, walletBalance(t, f.pool, tenantID))

	balance, err := f.store.GetCreditBalance(ctx, tenantID)
	require.NoError(t, err)
	require.Zero(t, balance.DebtMC, "the debt is cleared, so a later conversion never repays it twice")

	requireTenantReconciles(t, f.store, tenantID)
}

func TestIntegration_CreditFleetConvertsTheOldestTenantsFirst(t *testing.T) {
	f := newCreditEnforcementFixture(t)
	ctx := context.Background()

	first, _, _ := newConvertibleTenant(t, f, 10_000)
	second, _, _ := newConvertibleTenant(t, f, 30_000)

	_, err := f.pool.Exec(ctx, `UPDATE tenants SET created_at = '2000-01-01' WHERE tenant_id = ANY($1::uuid[])`, []string{first, second})
	require.NoError(t, err)

	result, err := f.store.ConvertCreditFleet(ctx, platform.CreditFleetConversionRequest{Reason: "fleet", Limit: 2}, conversionActor)
	require.NoError(t, err)
	require.Len(t, result.Converted, 2)
	require.Empty(t, result.Failed)

	converted := map[string]string{}
	for _, conversion := range result.Converted {
		converted[conversion.TenantID] = conversion.Credits
	}

	require.Equal(t, map[string]string{first: "1.00", second: "3.00"}, converted)
}

func TestIntegration_CreditTenantsCannotTopUpTheIRRWallet(t *testing.T) {
	f := newCreditEnforcementFixture(t)

	_, err := f.store.CreateBankTransferRequest(context.Background(), f.tenantID, platform.CreateBankTransferRequest{
		AmountIRR: 100_000, TrackingNumber: "T-1", BankTransactionAt: time.Now().UTC(),
	}, "member")
	require.ErrorIs(t, err, billingdomain.ErrWalletTopUpForCreditTenant)
}
