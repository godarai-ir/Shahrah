package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func orderCreditPack(t *testing.T, store *storage.Store, tenantID, slug, provider string) *platform.CreditPackOrder {
	t.Helper()

	order, err := store.CreateCreditPackOrder(context.Background(), tenantID, platform.CreateCreditPackOrderRequest{PackSlug: slug, Provider: provider},
		"https://cp.example/callback", "https://admin.example/billing", "user-1")
	require.NoError(t, err)

	return order
}

func creditWalletTransactions(t *testing.T, pool *pgxpool.Pool, tenantID string) int {
	t.Helper()

	var count int
	require.NoError(t, pool.QueryRow(context.Background(), `SELECT COUNT(1) FROM wallet_transactions WHERE tenant_id = $1::uuid`, tenantID).Scan(&count))

	return count
}

func TestIntegration_CreditPacksArePricedAtTheListPrice(t *testing.T) {
	store, _ := newCreditIntegrationStore(t)

	packs, err := store.ListCreditPacks(context.Background())
	require.NoError(t, err)

	prices := map[string]int64{}
	for _, pack := range packs {
		prices[pack.Slug] = pack.PriceIRR
		require.Equal(t, 12, pack.ValidityMonths, pack.Slug)
		require.Equal(t, pack.PriceIRR+pack.TaxIRR, pack.PayableIRR, pack.Slug)
	}

	require.Equal(t, map[string]int64{"credits-1000": 10_000_000, "credits-5000": 45_000_000, "credits-20000": 160_000_000}, prices)
}

func TestIntegration_CreditPackGatewayPurchaseGrantsOnce(t *testing.T) {
	store, pool := newCreditIntegrationStore(t)
	ctx := context.Background()

	irrTenant := newCreditLedgerTenant(t, pool, credits.BillingUnitIRR)
	_, err := store.CreateCreditPackOrder(ctx, irrTenant, platform.CreateCreditPackOrderRequest{PackSlug: "credits-1000", Provider: "zarinpal"}, "", "", "")
	require.ErrorIs(t, err, billing.ErrCreditBillingNotEnabled)

	tenantID := newCreditLedgerTenant(t, pool, credits.BillingUnitCredit)
	_, err = store.CreateCreditPackOrder(ctx, tenantID, platform.CreateCreditPackOrderRequest{PackSlug: "credits-7", Provider: "zarinpal"}, "", "", "")
	require.ErrorIs(t, err, billing.ErrCreditPackNotFound)

	order := orderCreditPack(t, store, tenantID, "credits-5000", "zarinpal")
	require.Equal(t, platform.CreditPackOrderPending, order.Status)
	require.Equal(t, int64(45_000_000), order.PriceIRR)

	authority := "auth-" + uuid.NewString()
	require.NoError(t, store.AttachCreditPackOrderGatewayData(ctx, order.PublicID, authority, "https://pay.example", nil))

	verified := billing.PaymentVerification{RefID: "ref-1", GatewayTxnID: "txn-1"}

	completed, duplicate, err := store.CompleteCreditPackOrderByAuthority(ctx, authority, verified)
	require.NoError(t, err)
	require.False(t, duplicate)
	require.Equal(t, platform.CreditPackOrderCompleted, completed.Status)
	require.Equal(t, "ref-1", completed.RefID)

	_, duplicate, err = store.CompleteCreditPackOrderByAuthority(ctx, authority, verified)
	require.NoError(t, err)
	require.True(t, duplicate, "a repeated callback grants nothing")

	require.NoError(t, store.MarkCreditPackOrderFailedByAuthority(ctx, authority, "late failure", nil))

	var (
		lots       int
		amountMC   int64
		expiresAt  time.Time
		perCredit  string
		purchaseIR int64
	)

	require.NoError(t, pool.QueryRow(ctx, `
		SELECT COUNT(1) OVER (), amount_mc, expires_at, price_per_credit_irr::text, purchase_price_irr::bigint
		FROM credit_lots WHERE tenant_id = $1::uuid
	`, tenantID).Scan(&lots, &amountMC, &expiresAt, &perCredit, &purchaseIR))
	require.Equal(t, 1, lots)
	require.Equal(t, int64(5_000_000), amountMC)
	require.WithinDuration(t, time.Now().AddDate(1, 0, 0), expiresAt, time.Minute, "add-ons last 12 months")
	require.Equal(t, "9000.0000", perCredit)
	require.Equal(t, int64(45_000_000), purchaseIR)

	again, err := store.GetCreditPackOrder(ctx, tenantID, order.PublicID)
	require.NoError(t, err)
	require.Equal(t, platform.CreditPackOrderCompleted, again.Status, "a completed order is never marked failed")

	failed := orderCreditPack(t, store, tenantID, "credits-1000", "zarinpal")
	failedAuthority := "auth-" + uuid.NewString()
	require.NoError(t, store.AttachCreditPackOrderGatewayData(ctx, failed.PublicID, failedAuthority, "", nil))
	require.NoError(t, store.MarkCreditPackOrderFailedByAuthority(ctx, failedAuthority, "canceled", nil))

	_, _, err = store.CompleteCreditPackOrderByAuthority(ctx, failedAuthority, verified)
	require.ErrorIs(t, err, billing.ErrCreditPackOrderNotPayable)

	requireTenantReconciles(t, store, tenantID)
}

func TestIntegration_CreditPackBankTransferCompletesTheOrder(t *testing.T) {
	store, pool := newCreditIntegrationStore(t)
	ctx := context.Background()

	tenantID := newCreditLedgerTenant(t, pool, credits.BillingUnitCredit)
	order := orderCreditPack(t, store, tenantID, "credits-1000", platform.CreditPackProviderBankTransfer)
	require.Equal(t, platform.CreditPackOrderAwaitingTransfer, order.Status)

	transfer := platform.CreateBankTransferRequest{
		AmountIRR: order.PayableIRR - 1, TrackingNumber: "trk-" + uuid.NewString()[:8],
		BankTransactionAt: time.Now().UTC(), CreditPackOrderID: order.PublicID,
	}

	_, err := store.CreateBankTransferRequest(ctx, tenantID, transfer, "user-1")
	require.ErrorIs(t, err, billing.ErrCreditPackTransferMismatch)

	transfer.AmountIRR = order.PayableIRR
	request, err := store.CreateBankTransferRequest(ctx, tenantID, transfer, "user-1")
	require.NoError(t, err)
	require.Equal(t, order.PublicID, request.Metadata["credit_pack_order_id"])

	actor := platform.SupportActor{ID: "finance-1", Role: "finance"}

	approved, err := store.ApproveBankTransferRequest(ctx, tenantID, request.ID, platform.ReviewBankTransferRequest{}, actor)
	require.NoError(t, err)
	require.Equal(t, "approved", approved.Status)

	_, err = store.ApproveBankTransferRequest(ctx, tenantID, request.ID, platform.ReviewBankTransferRequest{}, actor)
	require.NoError(t, err, "approving twice is a no-op")

	completed, err := store.GetCreditPackOrder(ctx, tenantID, order.PublicID)
	require.NoError(t, err)
	require.Equal(t, platform.CreditPackOrderCompleted, completed.Status)
	require.Equal(t, transfer.TrackingNumber, completed.RefID)

	balance, err := store.GetCreditBalance(ctx, tenantID)
	require.NoError(t, err)
	require.Equal(t, int64(1_000_000), balance.AvailableMC)
	require.Zero(t, creditWalletTransactions(t, pool, tenantID), "a pack transfer never credits the IRR wallet")

	transfer.TrackingNumber = "trk-" + uuid.NewString()[:8]
	_, err = store.CreateBankTransferRequest(ctx, tenantID, transfer, "user-1")
	require.ErrorIs(t, err, billing.ErrCreditPackOrderNotPayable, "a paid order takes no second transfer")

	var audits int
	require.NoError(t, pool.QueryRow(ctx, `
		SELECT COUNT(1) FROM financial_audit_events WHERE tenant_id = $1::uuid AND event_type IN ('credit_pack.purchased', 'bank_transfer.approved')
	`, tenantID).Scan(&audits))
	require.Equal(t, 2, audits)

	requireTenantReconciles(t, store, tenantID)
}

func TestIntegration_CreditOverageIsInvoicedBeforeTheNextSlice(t *testing.T) {
	store, pool := newCreditIntegrationStore(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)
	planID, versionID := draftCreditVersion(t, pool, "developer") // 5,000 credits a month

	settleAllCredits(t, store)

	results := map[bool]string{}

	for _, overage := range []bool{true, false} {
		tenantID := newCreditLedgerTenant(t, pool, credits.BillingUnitCredit)
		results[overage] = tenantID
		start := now.AddDate(0, 0, -3)
		subscriptionID := subscribeTenant(t, pool, tenantID, planID, versionID, "active", start, start.AddDate(1, 0, 0))

		if overage {
			_, err := store.UpdateCreditSettings(ctx, tenantID, platform.UpdateCreditSettingsRequest{OverageEnabled: &overage}, platform.SupportActor{ID: "admin"})
			require.NoError(t, err)
		}

		_, err := store.GrantDuePlanCredits(ctx, now)
		require.NoError(t, err)

		insertCreditSpendLog(t, pool, tenantID, "req-"+uuid.NewString(), 5_250_000, "")
		settleAllCredits(t, store)

		balance, err := store.GetCreditBalance(ctx, tenantID)
		require.NoError(t, err)
		require.Equal(t, int64(250_000), balance.DebtMC)

		// Move the subscription back so a new monthly slice (with a new start) is due.
		_, err = pool.Exec(ctx, `UPDATE tenant_subscriptions SET current_period_start = current_period_start - INTERVAL '1 month 1 day' WHERE subscription_id = $1::uuid`, subscriptionID)
		require.NoError(t, err)

		grants, err := store.GrantDuePlanCredits(ctx, now)
		require.NoError(t, err)

		if overage {
			require.GreaterOrEqual(t, grants.OverageMC, int64(250_000))
		}
	}

	var (
		amountIRR int64
		itemType  string
		quantity  string
	)

	require.NoError(t, pool.QueryRow(ctx, `
		SELECT i.final_amount_irr::bigint, li.item_type, li.quantity::text
		FROM invoices i JOIN invoice_line_items li ON li.invoice_id = i.invoice_id
		WHERE i.tenant_id = $1::uuid
	`, results[true]).Scan(&amountIRR, &itemType, &quantity))
	require.Equal(t, int64(2_500_000), amountIRR, "250 credits at 10,000 IRR")
	require.Equal(t, "credit_overage", itemType)
	require.Equal(t, "250.000000", quantity)

	for overage, tenantID := range results {
		balance, err := store.GetCreditBalance(ctx, tenantID)
		require.NoError(t, err)
		require.Zero(t, balance.DebtMC)

		want := int64(4_750_000) // the new slice repays the debt
		if overage {
			want = 5_000_000 // the invoice settled it, so the new slice is untouched
		}

		require.Equal(t, want, balance.AvailableMC, "overage=%v", overage)
		requireTenantReconciles(t, store, tenantID)
	}

	var invoices int
	require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(1) FROM invoices WHERE tenant_id = $1::uuid`, results[false]).Scan(&invoices))
	require.Zero(t, invoices, "without overage nothing is invoiced")

	_, err := store.GrantDuePlanCredits(ctx, now)
	require.NoError(t, err)
	require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(1) FROM invoices WHERE tenant_id = $1::uuid`, results[true]).Scan(&invoices))
	require.Equal(t, 1, invoices, "a repeated pass invoices nothing")
}

func TestIntegration_CreditSettingsRespectThePlan(t *testing.T) {
	store, pool := newCreditIntegrationStore(t)
	ctx := context.Background()
	now := time.Now().UTC()
	planID, versionID := draftCreditVersion(t, pool, "team")
	actor := platform.SupportActor{ID: "admin"}
	enable := true

	irrTenant := newCreditLedgerTenant(t, pool, credits.BillingUnitIRR)
	_, err := store.UpdateCreditSettings(ctx, irrTenant, platform.UpdateCreditSettingsRequest{OverageEnabled: &enable}, actor)
	require.ErrorIs(t, err, billing.ErrCreditBillingNotEnabled)

	tenantID := newCreditLedgerTenant(t, pool, credits.BillingUnitCredit)
	subscribeTenant(t, pool, tenantID, planID, versionID, "active", now.AddDate(0, 0, -1), now.AddDate(0, 1, 0))

	for _, bad := range []string{"12.505", "0", "-3", "abc"} {
		value := bad
		_, err = store.UpdateCreditSettings(ctx, tenantID, platform.UpdateCreditSettingsRequest{OverageCapCredits: &value}, actor)
		require.ErrorIs(t, err, billing.ErrInvalidOverageCap, bad)
	}

	capCredits := "12.5"
	settings, err := store.UpdateCreditSettings(ctx, tenantID, platform.UpdateCreditSettingsRequest{OverageEnabled: &enable, OverageCapCredits: &capCredits}, actor)
	require.NoError(t, err)
	require.True(t, settings.OverageEnabled)
	require.True(t, settings.OverageAllowed)
	require.NotNil(t, settings.OverageCapCredits)
	require.Equal(t, "12.50", *settings.OverageCapCredits)

	_, err = pool.Exec(ctx, `
		INSERT INTO tenant_entitlement_overrides (tenant_id, feature_key, value_json, reason, active)
		VALUES ($1::uuid, 'credits.overage_allowed', 'false'::jsonb, 'integration test', true)
	`, tenantID)
	require.NoError(t, err)

	_, err = store.UpdateCreditSettings(ctx, tenantID, platform.UpdateCreditSettingsRequest{OverageEnabled: &enable}, actor)
	require.ErrorIs(t, err, billing.ErrOverageNotAllowed)

	disable := false
	empty := ""
	settings, err = store.UpdateCreditSettings(ctx, tenantID, platform.UpdateCreditSettingsRequest{OverageEnabled: &disable, OverageCapCredits: &empty}, actor)
	require.NoError(t, err)
	require.False(t, settings.OverageEnabled)
	require.Nil(t, settings.OverageCapCredits)
}
