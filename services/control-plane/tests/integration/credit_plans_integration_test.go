package integration_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// draftCreditVersion returns the seeded draft version of a credit plan.
func draftCreditVersion(t *testing.T, pool *pgxpool.Pool, slug string) (planID, versionID string) {
	t.Helper()

	require.NoError(t, pool.QueryRow(context.Background(), `
		SELECT p.plan_id::text, pv.version_id::text
		FROM plan_versions pv
		JOIN commercial_plans p ON p.plan_id = pv.plan_id
		JOIN plan_entitlements pe ON pe.version_id = pv.version_id AND pe.feature_key = 'credits.monthly_grant'
		WHERE p.slug = $1
		ORDER BY pv.version DESC
		LIMIT 1
	`, slug).Scan(&planID, &versionID))

	return planID, versionID
}

func subscribeTenant(t *testing.T, pool *pgxpool.Pool, tenantID, planID, versionID, status string, start, end time.Time) string {
	t.Helper()

	var subscriptionID string

	var trialStart, trialEnd *time.Time
	if status == "trialing" {
		trialStart, trialEnd = &start, &end
	}

	require.NoError(t, pool.QueryRow(context.Background(), `
		INSERT INTO tenant_subscriptions (tenant_id, plan_id, version_id, status, billing_interval,
		                                  current_period_start, current_period_end, trial_started_at, trial_ends_at)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4, 'yearly', $5, $6, $7, $8)
		RETURNING subscription_id::text
	`, tenantID, planID, versionID, status, start, end, trialStart, trialEnd).Scan(&subscriptionID))

	return subscriptionID
}

func addTenantUser(t *testing.T, pool *pgxpool.Pool, tenantID string, active bool) string {
	t.Helper()

	userID := uuid.NewString()
	email := "member-" + userID[:8] + "@credits.example"

	_, err := pool.Exec(context.Background(), `
		INSERT INTO users (user_id, tenant_id, user_alias, email, role, active, created_at, updated_at)
		VALUES ($1::uuid, $2::uuid, 'Member', $3, 'member', $4, NOW(), NOW())
	`, userID, tenantID, email, active)
	require.NoError(t, err)

	return userID
}

func tenantSlug(t *testing.T, pool *pgxpool.Pool, tenantID string) string {
	t.Helper()

	var slug string
	require.NoError(t, pool.QueryRow(context.Background(), `SELECT slug FROM tenants WHERE tenant_id = $1::uuid`, tenantID).Scan(&slug))

	return slug
}

func TestIntegration_CreditPlanDraftsMatchThePRDInIRR(t *testing.T) {
	_, pool := newCreditIntegrationStore(t)
	ctx := context.Background()

	for _, offer := range credits.PlanOffers {
		_, versionID := draftCreditVersion(t, pool, offer.Slug)

		var (
			status                  string
			monthlyIRR, yearlyIRR   int64
			monthlyGrant, usersMax  int64
			copiedEntitlementsCount int
		)

		require.NoError(t, pool.QueryRow(ctx, `
			SELECT pv.status,
			       (SELECT amount_minor::bigint FROM plan_prices WHERE version_id = pv.version_id AND billing_interval = 'monthly'),
			       (SELECT amount_minor::bigint FROM plan_prices WHERE version_id = pv.version_id AND billing_interval = 'yearly'),
			       (SELECT value_json::text::bigint FROM plan_entitlements WHERE version_id = pv.version_id AND feature_key = 'credits.monthly_grant'),
			       (SELECT value_json::text::bigint FROM plan_entitlements WHERE version_id = pv.version_id AND feature_key = 'users.max'),
			       (SELECT COUNT(1) FROM plan_entitlements WHERE version_id = pv.version_id)
			FROM plan_versions pv WHERE pv.version_id = $1::uuid
		`, versionID).Scan(&status, &monthlyIRR, &yearlyIRR, &monthlyGrant, &usersMax, &copiedEntitlementsCount))

		// 067 seeds drafts; the fleet cutover (072) publishes them.
		require.Contains(t, []string{"draft", "published"}, status, offer.Slug)
		require.Equal(t, offer.MonthlyPriceIRR(), monthlyIRR, "%s monthly IRR", offer.Slug)
		require.Equal(t, offer.YearlyPriceIRR(), yearlyIRR, "%s yearly IRR", offer.Slug)
		require.Equal(t, offer.MonthlyCredits, monthlyGrant, offer.Slug)
		require.Equal(t, int64(offer.MaxUsers), usersMax, offer.Slug)
		require.Greater(t, copiedEntitlementsCount, 20, "%s keeps its other entitlements", offer.Slug)
	}

	var quarterly int
	require.NoError(t, pool.QueryRow(ctx, `
		SELECT COUNT(1) FROM plan_prices pp
		JOIN plan_entitlements pe ON pe.version_id = pp.version_id AND pe.feature_key = 'credits.monthly_grant'
		WHERE pp.billing_interval = 'quarterly'
	`).Scan(&quarterly))
	require.Zero(t, quarterly, "credit plans are monthly or yearly")
}

func TestIntegration_CreditPlanGrantsMonthlySlicesAndTrials(t *testing.T) {
	store, pool := newCreditIntegrationStore(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)
	planID, versionID := draftCreditVersion(t, pool, "team")

	annual := newCreditLedgerTenant(t, pool, credits.BillingUnitCredit)
	start := now.AddDate(0, -1, -10)
	subscriptionID := subscribeTenant(t, pool, annual, planID, versionID, "active", start, start.AddDate(1, 0, 0))

	trial := newCreditLedgerTenant(t, pool, credits.BillingUnitCredit)
	trialStart := now.AddDate(0, 0, -2)
	subscribeTenant(t, pool, trial, planID, versionID, "trialing", trialStart, trialStart.AddDate(0, 0, 15))

	irr := newCreditLedgerTenant(t, pool, credits.BillingUnitIRR)
	subscribeTenant(t, pool, irr, planID, versionID, "active", start, start.AddDate(1, 0, 0))

	_, err := store.GrantDuePlanCredits(ctx, now)
	require.NoError(t, err)

	again, err := store.GrantDuePlanCredits(ctx, now)
	require.NoError(t, err)
	require.Zero(t, again.Granted, "a repeated pass grants nothing")

	var (
		amountMC  int64
		expiresAt time.Time
		ref       string
	)

	require.NoError(t, pool.QueryRow(ctx, `
		SELECT amount_mc, expires_at, source_ref->>'slice_start' FROM credit_lots WHERE tenant_id = $1::uuid AND source = 'plan'
	`, annual).Scan(&amountMC, &expiresAt, &ref))
	require.Equal(t, int64(16_500_000), amountMC, "team grants 16,500 credits a month")
	require.WithinDuration(t, start.AddDate(0, 2, 0), expiresAt, time.Second, "the second monthly slice of the year")
	require.Contains(t, ref, start.AddDate(0, 1, 0).Format("2006-01-02"))

	var lots int
	require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(1) FROM credit_lots WHERE tenant_id = $1::uuid`, annual).Scan(&lots))
	require.Equal(t, 1, lots, "only the current slice is granted")

	require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(1) FROM credit_lots WHERE tenant_id = $1::uuid`, trial).Scan(&lots))
	require.Zero(t, lots, "trialing tenants get no plan or trial grant; free seats are their trial credits")

	require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(1) FROM credit_lots WHERE tenant_id = $1::uuid`, irr).Scan(&lots))
	require.Zero(t, lots, "IRR tenants get no credit grants")

	balance, err := store.GetCreditBalance(ctx, annual)
	require.NoError(t, err)
	require.Equal(t, int64(16_500_000), balance.AvailableMC)
	require.NotEmpty(t, subscriptionID)

	requireTenantReconciles(t, store, annual)
}

func TestIntegration_UserCapCountsInvitationsAndIsAtomic(t *testing.T) {
	store, pool := newCreditIntegrationStore(t)
	ctx := context.Background()
	now := time.Now().UTC()
	planID, versionID := draftCreditVersion(t, pool, "developer") // 3 users

	tenantID := newCreditLedgerTenant(t, pool, credits.BillingUnitIRR)
	subscribeTenant(t, pool, tenantID, planID, versionID, "active", now.AddDate(0, 0, -1), now.AddDate(0, 1, 0))
	slug := tenantSlug(t, pool, tenantID)
	addTenantUser(t, pool, tenantID, true)

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		invited int
		capped  int
	)

	for i := range 6 {
		wg.Go(func() {
			_, _, err := store.InviteTenantAdminUser(ctx, tenantID, slug, platform.InviteTenantAdminUserRequest{
				Email: fmt.Sprintf("invitee-%d-%s@credits.example", i, uuid.NewString()[:6]), Role: "member",
			})

			mu.Lock()
			defer mu.Unlock()

			switch {
			case err == nil:
				invited++
			case errors.Is(err, billing.ErrPlanUserLimitReached):
				capped++
			default:
				t.Errorf("unexpected invite error: %v", err)
			}
		})
	}

	wg.Wait()
	require.Equal(t, 2, invited, "one member plus two invitations fill three seats")
	require.Equal(t, 4, capped)

	seats, err := store.GetTenantSeatUsage(ctx, tenantID)
	require.NoError(t, err)
	require.Equal(t, 3, seats.Used)
	require.Equal(t, 1, seats.ActiveMembers)
	require.Equal(t, 2, seats.PendingInvitations)
	require.NotNil(t, seats.Limit)
	require.Equal(t, 3, *seats.Limit)

	// A deactivated member cannot come back while the plan is full, by either path.
	sleeperID := addTenantUser(t, pool, tenantID, false)

	_, err = store.SetTenantAdminUserActive(ctx, slug, sleeperID, true)
	require.ErrorIs(t, err, billing.ErrPlanUserLimitReached)

	var sleeperEmail string
	require.NoError(t, pool.QueryRow(ctx, `SELECT email FROM users WHERE user_id = $1::uuid`, sleeperID).Scan(&sleeperEmail))

	_, err = store.UpdateUserRoles(ctx, tenantID, platform.UpdateUserRolesRequest{Email: sleeperEmail, Roles: []string{"member"}, ResourceType: "tenant"})
	require.ErrorIs(t, err, billing.ErrPlanUserLimitReached)

	var active bool
	require.NoError(t, pool.QueryRow(ctx, `SELECT active FROM users WHERE user_id = $1::uuid`, sleeperID).Scan(&active))
	require.False(t, active)

	// Freeing a seat lets the member back.
	_, err = pool.Exec(ctx, `UPDATE tenant_admin_invitations SET status = 'revoked' WHERE tenant_id = $1::uuid AND status = 'pending' AND email LIKE 'invitee-%'`, tenantID)
	require.NoError(t, err)

	_, err = store.SetTenantAdminUserActive(ctx, slug, sleeperID, true)
	require.NoError(t, err)
}

func TestIntegration_UserCapGrandfathersAndBlocksDowngrades(t *testing.T) {
	store, pool := newCreditIntegrationStore(t)
	ctx := context.Background()
	now := time.Now().UTC()
	planID, versionID := draftCreditVersion(t, pool, "team") // 10 users

	tenantID := newCreditLedgerTenant(t, pool, credits.BillingUnitIRR)
	subscribeTenant(t, pool, tenantID, planID, versionID, "active", now.AddDate(0, 0, -1), now.AddDate(0, 1, 0))
	slug := tenantSlug(t, pool, tenantID)

	for range 4 {
		addTenantUser(t, pool, tenantID, true)
	}

	// The published developer plan allows 3 users; 4 active members do not fit.
	_, err := store.ChangeTenantSubscription(ctx, tenantID, platform.ChangeTenantSubscriptionRequest{PlanSlug: "developer", BillingInterval: "monthly"})
	require.ErrorIs(t, err, billing.ErrPlanUserLimitReached)

	_, err = store.PrepareTenantSubscriptionChange(ctx, tenantID, platform.ChangeTenantSubscriptionRequest{PlanSlug: "developer", BillingInterval: "monthly"}, "https://cb.example", "https://return.example")
	require.ErrorIs(t, err, billing.ErrPlanUserLimitReached)

	// A tenant already over its cap keeps its members but cannot add more.
	_, err = pool.Exec(ctx, `
		INSERT INTO tenant_entitlement_overrides (tenant_id, feature_key, value_json, reason, active)
		VALUES ($1::uuid, 'users.max', '2'::jsonb, 'integration test', true)
	`, tenantID)
	require.NoError(t, err)

	_, _, err = store.InviteTenantAdminUser(ctx, tenantID, slug, platform.InviteTenantAdminUserRequest{Email: "late-" + strings.ToLower(uuid.NewString()[:6]) + "@credits.example"})
	require.ErrorIs(t, err, billing.ErrPlanUserLimitReached)

	var activeMembers int
	require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(1) FROM users WHERE tenant_id = $1::uuid AND active`, tenantID).Scan(&activeMembers))
	require.Equal(t, 4, activeMembers, "nobody is removed")
}

func TestIntegration_CreditTrialCapsUsersForCreditTenantsOnly(t *testing.T) {
	store, pool := newCreditIntegrationStore(t)
	ctx := context.Background()
	now := time.Now().UTC()
	planID, versionID := draftCreditVersion(t, pool, "growth") // 30 users

	for unit, wantLimit := range map[string]int{credits.BillingUnitCredit: credits.TrialMaxUsers, credits.BillingUnitIRR: 30} {
		tenantID := newCreditLedgerTenant(t, pool, unit)
		subscribeTenant(t, pool, tenantID, planID, versionID, "trialing", now.AddDate(0, 0, -1), now.AddDate(0, 0, 13))

		seats, err := store.GetTenantSeatUsage(ctx, tenantID)
		require.NoError(t, err)
		require.NotNil(t, seats.Limit)
		require.Equal(t, wantLimit, *seats.Limit, unit)
	}
}
