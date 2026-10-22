package integration_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	billingdomain "github.com/Delpak/shahrah/services/control-plane/internal/domain/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/seats"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/Delpak/shahrah/services/control-plane/internal/testcache"
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
	mem, err := testcache.New(t, logger)
	require.NoError(t, err)

	store, err := storage.NewStore(ctx, databaseURL, logger, storage.StoreDependencies{Cache: mem})
	require.NoError(t, err)

	defer store.Close()

	pool, err := pgxpool.New(ctx, databaseURL)
	require.NoError(t, err)

	defer pool.Close()

	resetPersistentState(t, pool)

	plans, err := store.ListPublicCommercialPlans(ctx)
	require.NoError(t, err)

	// Plans are retired: only the free plan and the seat subscription remain.
	assert.ElementsMatch(t, []string{seats.FreePlanSlug, seats.SeatsPlanSlug}, planSlugs(plans))

	features, err := store.ListFeatureCatalog(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(features), 20)

	tenant := createIntegrationTenant(t, pool, "entitlement-"+uuid.NewString()[:8])
	_, err = store.UpsertTenantSubscription(ctx, platform.UpsertTenantSubscriptionRequest{TenantID: tenant.ID, PlanSlug: seats.FreePlanSlug}, "test")
	require.NoError(t, err)
	base, err := store.ResolveEffectiveEntitlements(ctx, tenant.ID)
	require.NoError(t, err)
	assert.Nil(t, base.Features[platform.FeatureUsersMax], "members come from seats, not users.max")
	assert.Equal(t, true, base.Features["rbac.enabled"], "every feature is on for every tenant")

	_, err = store.UpsertTenantSubscription(ctx, platform.UpsertTenantSubscriptionRequest{TenantID: tenant.ID, PlanSlug: "developer"}, "test")
	require.ErrorIs(t, err, billingdomain.ErrPlansRetired)

	_, err = store.UpsertTenantEntitlementOverride(ctx, platform.UpsertTenantEntitlementOverrideRequest{
		TenantID: tenant.ID, FeatureKey: platform.FeatureUsersMax, Value: 10, Reason: "contract", CreatedBy: "test",
	})
	require.NoError(t, err)
	effective, err := store.ResolveEffectiveEntitlements(ctx, tenant.ID)
	require.NoError(t, err)
	assert.EqualValues(t, 10, effective.Features[platform.FeatureUsersMax])
	assert.EqualValues(t, 10, effective.Overrides[platform.FeatureUsersMax])
}

func TestIntegration_PublishedVersionAndRetiredPlanChanges(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if strings.TrimSpace(databaseURL) == "" {
		t.Skip("DATABASE_URL not set")
	}

	ctx := context.Background()
	logger := zap.NewNop()
	store, err := storage.NewStore(ctx, databaseURL, logger, storage.StoreDependencies{})
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
	assert.ErrorIs(t, err, billingdomain.ErrPlanVersionPublished)

	// Tenants can no longer buy or change plans: paid seats start the seat
	// subscription instead.
	tenant := createIntegrationTenant(t, pool, "retired-"+uuid.NewString()[:8])
	_, err = store.PrepareTenantSubscriptionChange(ctx, tenant.ID, platform.ChangeTenantSubscriptionRequest{
		PlanSlug: planSlug, BillingInterval: "yearly", Provider: "mock",
	}, "http://localhost/callback", "http://localhost/return")
	require.ErrorIs(t, err, billingdomain.ErrPlansRetired)

	_, err = store.ChangeTenantSubscription(ctx, tenant.ID, platform.ChangeTenantSubscriptionRequest{PlanSlug: "developer", BillingInterval: "monthly"})
	require.ErrorIs(t, err, billingdomain.ErrPlansRetired)
}

func planSlugs(plans []platform.CommercialPlan) []string {
	out := make([]string, 0, len(plans))
	for i := range plans {
		out = append(out, plans[i].Slug)
	}

	return out
}
