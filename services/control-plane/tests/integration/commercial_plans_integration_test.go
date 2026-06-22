package integration_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestIntegration_CommercialPlanSeedOverrideAndMonthlyQuota(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if strings.TrimSpace(databaseURL) == "" {
		t.Skip("DATABASE_URL not set")
	}
	ctx := context.Background()
	logger := zap.NewNop()
	store, err := storage.NewStore(ctx, databaseURL, logger)
	require.NoError(t, err)
	defer store.Close()
	pool, err := pgxpool.New(ctx, databaseURL)
	require.NoError(t, err)
	defer pool.Close()
	mem, err := cache.NewMemoryCache(logger)
	require.NoError(t, err)
	defer mem.Close()
	store.SetEntitlementCache(mem)
	resetPersistentState(t, pool)

	plans, err := store.ListPublicCommercialPlans(ctx)
	require.NoError(t, err)
	slugs := planSlugs(plans)
	for _, expected := range []string{"developer", "team", "growth", "enterprise"} {
		assert.Contains(t, slugs, expected)
	}
	features, err := store.ListFeatureCatalog(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(features), 20)

	tenant := createIntegrationTenant(t, pool, "entitlement-"+uuid.NewString()[:8])
	_, err = store.UpsertTenantSubscription(ctx, platform.UpsertTenantSubscriptionRequest{TenantID: tenant.ID, PlanSlug: "developer", BillingInterval: "yearly"}, "test")
	require.NoError(t, err)
	base, err := store.ResolveEffectiveEntitlements(ctx, tenant.ID)
	require.NoError(t, err)
	assert.EqualValues(t, 3, base.Features[platform.FeatureUsersMax])

	_, err = store.UpsertTenantEntitlementOverride(ctx, platform.UpsertTenantEntitlementOverrideRequest{
		TenantID: tenant.ID, FeatureKey: platform.FeatureUsersMax, Value: 10, Reason: "contract", CreatedBy: "test",
	})
	require.NoError(t, err)
	effective, err := store.ResolveEffectiveEntitlements(ctx, tenant.ID)
	require.NoError(t, err)
	assert.EqualValues(t, 10, effective.Features[platform.FeatureUsersMax])
	assert.EqualValues(t, 10, effective.Overrides[platform.FeatureUsersMax])

	january := time.Date(2026, time.January, 15, 12, 0, 0, 0, time.UTC)
	_, allowed, err := store.CheckAndIncrementTenantUsageCounter(ctx, tenant.ID, platform.FeatureGatewayRequestsMonthly, 1, 1, january)
	require.NoError(t, err)
	assert.True(t, allowed)
	_, allowed, err = store.CheckAndIncrementTenantUsageCounter(ctx, tenant.ID, platform.FeatureGatewayRequestsMonthly, 1, 1, january)
	require.NoError(t, err)
	assert.False(t, allowed)
	february := time.Date(2026, time.February, 1, 0, 0, 1, 0, time.UTC)
	_, allowed, err = store.CheckAndIncrementTenantUsageCounter(ctx, tenant.ID, platform.FeatureGatewayRequestsMonthly, 1, 1, february)
	require.NoError(t, err)
	assert.True(t, allowed, "yearly subscriptions must still use monthly quota windows")
}

func TestIntegration_PublishedVersionAndPeriodEndDowngrade(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if strings.TrimSpace(databaseURL) == "" {
		t.Skip("DATABASE_URL not set")
	}
	ctx := context.Background()
	logger := zap.NewNop()
	store, err := storage.NewStore(ctx, databaseURL, logger)
	require.NoError(t, err)
	defer store.Close()
	pool, err := pgxpool.New(ctx, databaseURL)
	require.NoError(t, err)
	defer pool.Close()
	resetPersistentState(t, pool)

	planSlug := "publish-test-" + uuid.NewString()[:8]
	_, err = store.CreateCommercialPlan(ctx, platform.CreateCommercialPlanRequest{Slug: planSlug, DisplayName: "Publish Test", SortOrder: 50})
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM subscription_payment_intents WHERE plan_id IN (SELECT plan_id FROM commercial_plans WHERE slug = $1)`, planSlug)
		_, _ = pool.Exec(context.Background(), `DELETE FROM commercial_plans WHERE slug = $1`, planSlug)
	})
	version, err := store.CreatePlanVersion(ctx, planSlug, platform.CreatePlanVersionRequest{Version: 1})
	require.NoError(t, err)
	_, err = store.UpsertPlanEntitlement(ctx, planSlug, version.Version, platform.UpsertPlanEntitlementRequest{FeatureKey: platform.FeatureUsersMax, Value: 5})
	require.NoError(t, err)
	amount := float64(120000)
	_, err = store.UpsertPlanPrice(ctx, planSlug, version.Version, platform.UpsertPlanPriceRequest{BillingInterval: "yearly", Currency: "IRR", AmountMinor: &amount, DiscountPercent: 20})
	require.NoError(t, err)
	_, err = store.PublishPlanVersion(ctx, planSlug, version.Version)
	require.NoError(t, err)
	_, err = store.UpsertPlanEntitlement(ctx, planSlug, version.Version, platform.UpsertPlanEntitlementRequest{FeatureKey: platform.FeatureUsersMax, Value: 6})
	assert.ErrorIs(t, err, storage.ErrPlanVersionPublished)

	tenant := createIntegrationTenant(t, pool, "downgrade-"+uuid.NewString()[:8])
	paymentChange, err := store.PrepareTenantSubscriptionChange(ctx, tenant.ID, platform.ChangeTenantSubscriptionRequest{
		PlanSlug: planSlug, BillingInterval: "yearly", Provider: "mock",
	}, "http://localhost/callback", "http://localhost/return")
	require.NoError(t, err)
	require.NotNil(t, paymentChange.PaymentIntent)
	assert.Equal(t, "payment_required", paymentChange.Mode)
	assert.Equal(t, float64(120000), paymentChange.PaymentIntent.AmountMinor)
	assert.Equal(t, float64(20), paymentChange.PaymentIntent.DiscountPercent)
	authority := "test_" + uuid.NewString()
	require.NoError(t, store.AttachSubscriptionPaymentGatewayData(ctx, paymentChange.PaymentIntent.PublicID, authority, "http://localhost/pay", map[string]interface{}{"test": true}))
	_, paidSubscription, alreadyProcessed, err := store.CompleteSubscriptionPaymentByAuthority(ctx, authority, map[string]interface{}{"status": "ok"}, "ref-1", "txn-1", "", "")
	require.NoError(t, err)
	assert.False(t, alreadyProcessed)
	require.NotNil(t, paidSubscription)
	assert.Equal(t, planSlug, paidSubscription.PlanSlug)
	_, _, alreadyProcessed, err = store.CompleteSubscriptionPaymentByAuthority(ctx, authority, map[string]interface{}{"status": "ok"}, "ref-1", "txn-1", "", "")
	require.NoError(t, err)
	assert.True(t, alreadyProcessed)
	_, err = store.UpsertTenantSubscription(ctx, platform.UpsertTenantSubscriptionRequest{TenantID: tenant.ID, PlanSlug: "growth", BillingInterval: "monthly"}, "test")
	require.NoError(t, err)
	scheduled, err := store.ChangeTenantSubscription(ctx, tenant.ID, platform.ChangeTenantSubscriptionRequest{PlanSlug: "developer", BillingInterval: "monthly"})
	require.NoError(t, err)
	assert.True(t, scheduled.CancelAtPeriodEnd)
	assert.NotEmpty(t, scheduled.PendingVersionID)
	_, err = pool.Exec(ctx, `UPDATE tenant_subscriptions SET current_period_end = NOW() - INTERVAL '1 minute' WHERE subscription_id = $1::uuid`, scheduled.ID)
	require.NoError(t, err)
	processed, err := store.ProcessSubscriptionTransitions(ctx, time.Now().UTC(), 10)
	require.NoError(t, err)
	assert.Equal(t, 1, processed)
	current, err := store.GetTenantSubscription(ctx, tenant.ID)
	require.NoError(t, err)
	assert.Equal(t, "developer", current.PlanSlug)
}

func planSlugs(plans []platform.CommercialPlan) []string {
	out := make([]string, 0, len(plans))
	for _, plan := range plans {
		out = append(out, plan.Slug)
	}
	return out
}
