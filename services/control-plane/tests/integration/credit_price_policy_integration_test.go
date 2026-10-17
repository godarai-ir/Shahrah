package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

var testActor = platform.SupportActor{ID: "finance-1", Role: "finance"}

// restoreActivePrice puts the launch revision back after a test changed the
// global list price, and removes everything the test's revisions created.
func restoreActivePrice(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	var original string
	require.NoError(t, pool.QueryRow(context.Background(), `SELECT revision_id::text FROM credit_price_revisions WHERE status = 'active'`).Scan(&original))

	// Publishing repriced versions archives the published ones, catalog plans
	// included; they are published again so later tests find their plans.
	var published []string
	require.NoError(t, pool.QueryRow(context.Background(), `
		SELECT COALESCE(array_agg(version_id::text), '{}') FROM plan_versions WHERE status = 'published'
	`).Scan(&published))

	t.Cleanup(func() {
		ctx := context.Background()
		statements := []string{
			`DELETE FROM plan_versions WHERE credit_price_revision_id IS NOT NULL AND credit_price_revision_id <> $1::uuid`,
			// Test revisions become "rejected", never "superseded": the public changelog
			// lists superseded prices as real history.
			`UPDATE credit_price_revisions SET status = 'rejected' WHERE status IN ('active', 'proposed', 'superseded') AND revision_id <> $1::uuid AND proposed_by = 'price_policy'`,
			`UPDATE credit_price_revisions SET status = 'active' WHERE revision_id = $1::uuid`,
			`DELETE FROM exchange_rate_samples WHERE $1::text IS NOT NULL`,
		}

		for _, statement := range statements {
			_, err := pool.Exec(ctx, statement, original)
			require.NoError(t, err, statement)
		}

		_, err := pool.Exec(ctx, `UPDATE plan_versions SET status = 'published' WHERE version_id = ANY($1::uuid[])`, published)
		require.NoError(t, err)
	})
}

// newPublishedCreditPlan creates a throwaway published plan with a credit
// grant, so price revisions have a plan to reprice.
func newPublishedCreditPlan(t *testing.T, pool *pgxpool.Pool) (planID string) {
	t.Helper()

	ctx := context.Background()
	slug := "test-credit-" + uuid.NewString()[:8]

	require.NoError(t, pool.QueryRow(ctx, `
		INSERT INTO commercial_plans (slug, display_name, sort_order, active) VALUES ($1, 'Test credit plan', 900, true)
		RETURNING plan_id::text
	`, slug).Scan(&planID))

	var versionID string
	require.NoError(t, pool.QueryRow(ctx, `
		INSERT INTO plan_versions (plan_id, version, status, published_at) VALUES ($1::uuid, 1, 'published', NOW())
		RETURNING version_id::text
	`, planID).Scan(&versionID))

	_, err := pool.Exec(ctx, `
		INSERT INTO plan_entitlements (version_id, feature_key, value_json) VALUES ($1::uuid, 'credits.monthly_grant', '5000'::jsonb);
	`, versionID)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		INSERT INTO plan_prices (version_id, billing_interval, currency, amount_minor, discount_percent)
		VALUES ($1::uuid, 'monthly', 'IRR', 50000000, 0), ($1::uuid, 'yearly', 'IRR', 540000000, 10)
	`, versionID)
	require.NoError(t, err)

	// Archived, not only inactive: repricing covers every published credit version.
	t.Cleanup(func() {
		_, cleanupErr := pool.Exec(context.Background(), `UPDATE commercial_plans SET active = false WHERE plan_id = $1::uuid`, planID)
		require.NoError(t, cleanupErr)

		_, cleanupErr = pool.Exec(context.Background(), `UPDATE plan_versions SET status = 'archived' WHERE plan_id = $1::uuid`, planID)
		require.NoError(t, cleanupErr)
	})

	return planID
}

func seedRateSamples(t *testing.T, pool *pgxpool.Pool, now time.Time, usdToIRR int64) {
	t.Helper()

	for day := 1; day <= credits.PriceAverageDays; day++ {
		hour := now.UTC().Truncate(24*time.Hour).AddDate(0, 0, -day).Add(12 * time.Hour)
		_, err := pool.Exec(context.Background(), `
			INSERT INTO exchange_rate_samples (sample_hour, source, usd_to_irr, rate_fetched_at)
			VALUES ($1, 'test', $2, $1) ON CONFLICT (sample_hour) DO UPDATE SET usd_to_irr = EXCLUDED.usd_to_irr
		`, hour, usdToIRR)
		require.NoError(t, err)
	}
}

func TestIntegration_CreditPricePolicyProposesApprovesAndReprices(t *testing.T) {
	store, pool := newCreditIntegrationStore(t)
	ctx := context.Background()
	now := time.Now().UTC()

	restoreActivePrice(t, pool)
	planID := newPublishedCreditPlan(t, pool)

	_, err := pool.Exec(ctx, `DELETE FROM exchange_rate_samples`)
	require.NoError(t, err)

	_, due, err := store.ProposeCreditPrice(ctx, now)
	require.NoError(t, err)
	require.False(t, due, "no samples, no proposal")

	seedRateSamples(t, pool, now, 2_900_000) // 290,000 toman: inside the band
	_, due, err = store.ProposeCreditPrice(ctx, now)
	require.NoError(t, err)
	require.False(t, due)

	seedRateSamples(t, pool, now, 3_000_000) // 300,000 toman: outside the band
	proposal, due, err := store.ProposeCreditPrice(ctx, now)
	require.NoError(t, err)
	require.True(t, due)
	require.Equal(t, int64(1_100), proposal.PriceTomanPerCredit, "the PRD example")
	require.Equal(t, "300000.00", proposal.BaseUSDToman)

	again, _, err := store.ProposeCreditPrice(ctx, now)
	require.NoError(t, err)
	require.Equal(t, proposal.ID, again.ID, "an open proposal is refreshed, not duplicated")

	_, err = store.ApproveCreditPriceRevision(ctx, proposal.ID, testActor, " ", now)
	require.ErrorIs(t, err, billing.ErrDecisionReasonRequired)

	approved, err := store.ApproveCreditPriceRevision(ctx, proposal.ID, testActor, "7-day average left the band", now)
	require.NoError(t, err)
	require.Equal(t, "active", approved.Status)
	require.WithinDuration(t, now.AddDate(0, 0, credits.PlanPriceNoticeDays), *approved.PlanEffectiveFrom, time.Minute)

	packs, err := store.ListCreditPacks(ctx)
	require.NoError(t, err)

	for _, pack := range packs {
		if pack.Slug == "credits-1000" {
			require.Equal(t, int64(11_000_000), pack.PriceIRR, "add-ons follow the new price at once")
		}
	}

	var draftMonthly int64
	require.NoError(t, pool.QueryRow(ctx, `
		SELECT pp.amount_minor::bigint FROM plan_versions pv
		JOIN plan_prices pp ON pp.version_id = pv.version_id AND pp.billing_interval = 'monthly'
		WHERE pv.plan_id = $1::uuid AND pv.status = 'draft' AND pv.credit_price_revision_id = $2::uuid
	`, planID, proposal.ID).Scan(&draftMonthly))
	require.Equal(t, int64(55_000_000), draftMonthly, "plan prices scale with the list price")

	published, err := store.PublishDueRevisionPlanVersions(ctx, now)
	require.NoError(t, err)
	require.Zero(t, published, "plan prices wait for the 30-day notice")

	var repriced int
	require.NoError(t, pool.QueryRow(ctx, `
		SELECT COUNT(1) FROM plan_versions WHERE status = 'draft' AND credit_price_revision_id = $1::uuid
	`, proposal.ID).Scan(&repriced))
	require.GreaterOrEqual(t, repriced, 1, "the test plan, and any published credit plan, is repriced")

	published, err = store.PublishDueRevisionPlanVersions(ctx, now.AddDate(0, 0, credits.PlanPriceNoticeDays+1))
	require.NoError(t, err)
	require.Equal(t, repriced, published)

	var publishedVersion int
	require.NoError(t, pool.QueryRow(ctx, `SELECT version FROM plan_versions WHERE plan_id = $1::uuid AND status = 'published'`, planID).Scan(&publishedVersion))
	require.Equal(t, 2, publishedVersion)

	changelog, err := store.ListPricingChangelog(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(changelog.CreditPrices), 2)
	require.Equal(t, int64(1_100), changelog.CreditPrices[0].PriceTomanPerCredit)
	require.NotEmpty(t, changelog.RateCards)

	status, err := store.GetCreditPriceStatus(ctx, now)
	require.NoError(t, err)
	require.NotNil(t, status.Active)
	require.Nil(t, status.Proposed)
	require.NotNil(t, status.SevenDayAverageUSDToman)

	var audits int
	require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(1) FROM financial_audit_events WHERE event_type = 'credit_price.approved' AND source_id = $1`, proposal.ID).Scan(&audits))
	require.Equal(t, 1, audits)
}

func TestIntegration_CreditPriceRejectAndStaleSamples(t *testing.T) {
	store, pool := newCreditIntegrationStore(t)
	ctx := context.Background()
	now := time.Now().UTC()

	restoreActivePrice(t, pool)

	_, err := pool.Exec(ctx, `
		INSERT INTO exchange_rate_current (rate_id, source, usd_to_irr, fetched_at, updated_at)
		VALUES (gen_random_uuid(), 'test:stale', 2500000, NOW() - INTERVAL '2 days', NOW())
		ON CONFLICT (source) DO UPDATE SET usd_to_irr = EXCLUDED.usd_to_irr, fetched_at = EXCLUDED.fetched_at
	`)
	require.NoError(t, err)

	var newest time.Time
	require.NoError(t, pool.QueryRow(ctx, `SELECT MAX(fetched_at) FROM exchange_rate_current`).Scan(&newest))

	recorded, err := store.RecordExchangeRateSample(ctx, now)
	require.NoError(t, err)
	require.Equal(t, now.Sub(newest) < 6*time.Hour, recorded, "only a fresh rate is sampled")

	seedRateSamples(t, pool, now, 2_000_000) // 200,000 toman: price falls
	proposal, due, err := store.ProposeCreditPrice(ctx, now)
	require.NoError(t, err)
	require.True(t, due)
	require.Equal(t, int64(750), proposal.PriceTomanPerCredit)

	rejected, err := store.RejectCreditPriceRevision(ctx, proposal.ID, testActor, "temporary spike")
	require.NoError(t, err)
	require.Equal(t, "rejected", rejected.Status)

	_, err = store.ApproveCreditPriceRevision(ctx, proposal.ID, testActor, "too late", now)
	require.ErrorIs(t, err, billing.ErrCreditPriceRevisionNotFound, "a decided proposal cannot be approved")
}

func TestIntegration_CreditRateCardIncreasesWaitForNotice(t *testing.T) {
	store, pool := newCreditIntegrationStore(t)
	ctx := context.Background()
	now := time.Now().UTC()

	var original int
	require.NoError(t, pool.QueryRow(ctx, `SELECT version FROM credit_rate_card_versions WHERE status = 'published'`).Scan(&original))
	t.Cleanup(func() {
		_, err := pool.Exec(context.Background(), `DELETE FROM credit_rate_card_versions WHERE version > $1`, original)
		require.NoError(t, err)
		_, err = pool.Exec(context.Background(), `UPDATE credit_rate_card_versions SET status = 'published' WHERE version = $1`, original)
		require.NoError(t, err)
	})

	draft, err := store.CreateCreditRateCardDraft(ctx, testActor)
	require.NoError(t, err)

	again, err := store.CreateCreditRateCardDraft(ctx, testActor)
	require.NoError(t, err)
	require.Equal(t, draft, again, "one draft at a time")

	err = store.UpsertCreditRateCardMultiplier(ctx, draft, platform.UpsertCreditMultiplierRequest{Model: "xai/grok-test", Multiplier: "1.25"}, testActor)
	require.ErrorIs(t, err, billing.ErrDecisionReasonRequired)

	err = store.UpsertCreditRateCardMultiplier(ctx, draft, platform.UpsertCreditMultiplierRequest{Model: "xai/grok-test", Multiplier: "1.12345", Reason: "x"}, testActor)
	require.ErrorIs(t, err, billing.ErrInvalidCreditMultiplier)

	require.NoError(t, store.UpsertCreditRateCardMultiplier(ctx, draft, platform.UpsertCreditMultiplierRequest{
		Model: " XAI/Grok-Test ", Multiplier: "1.25", Reason: "tunnel egress cost",
	}, testActor))

	scheduled, err := store.PublishCreditRateCard(ctx, draft, platform.PublishCreditRateCardRequest{ChangelogFA: "افزایش نرخ grok", ChangelogEN: "grok rate increase"}, testActor, now)
	require.NoError(t, err)
	require.Equal(t, "scheduled", scheduled.Status, "an increase waits for the 30-day notice")
	require.Equal(t, "xai/grok-test", scheduled.Multipliers[0].Model)

	card, err := store.GetPublishedCreditRateCard(ctx)
	require.NoError(t, err)
	require.Equal(t, original, card.Version, "nothing changes during the notice")

	count, err := store.PublishDueRateCards(ctx, now.AddDate(0, 0, credits.PlanPriceNoticeDays+1))
	require.NoError(t, err)
	require.Equal(t, 1, count)

	card, err = store.GetPublishedCreditRateCard(ctx)
	require.NoError(t, err)
	require.Equal(t, draft, card.Version)

	cut, err := store.CreateCreditRateCardDraft(ctx, testActor)
	require.NoError(t, err)
	require.NoError(t, store.UpsertCreditRateCardMultiplier(ctx, cut, platform.UpsertCreditMultiplierRequest{Model: "xai/grok-test", Multiplier: "1.10", Reason: "egress renegotiated"}, testActor))

	immediate, err := store.PublishCreditRateCard(ctx, cut, platform.PublishCreditRateCardRequest{ChangelogFA: "کاهش نرخ"}, testActor, now)
	require.NoError(t, err)
	require.Equal(t, "published", immediate.Status, "a cut applies at once")

	_, err = store.PublishCreditRateCard(ctx, cut, platform.PublishCreditRateCardRequest{ChangelogFA: "دوباره"}, testActor, now)
	require.ErrorIs(t, err, billing.ErrCreditRateCardNotDraft)
}

func TestIntegration_CreditPackCatalogVersions(t *testing.T) {
	store, pool := newCreditIntegrationStore(t)
	ctx := context.Background()
	slug := "credits-test-" + uuid.NewString()[:6]

	_, err := store.CreateCreditPackVersion(ctx, platform.CreateCreditPackVersionRequest{Slug: slug, Credits: 0, DiscountPercent: "0", ValidityMonths: 12}, testActor)
	require.ErrorIs(t, err, billing.ErrInvalidCreditPack)

	first, err := store.CreateCreditPackVersion(ctx, platform.CreateCreditPackVersionRequest{Slug: slug, Credits: 2_000, DiscountPercent: "5", ValidityMonths: 12, SortOrder: 999}, testActor)
	require.NoError(t, err)
	require.Equal(t, 1, first.Version)
	require.Equal(t, "draft", first.Status)

	t.Cleanup(func() {
		_, cleanupErr := pool.Exec(context.Background(), `UPDATE credit_packs SET status = 'archived' WHERE slug = $1`, slug)
		require.NoError(t, cleanupErr)
	})

	require.NoError(t, store.PublishCreditPackVersion(ctx, slug, first.Version, testActor))

	second, err := store.CreateCreditPackVersion(ctx, platform.CreateCreditPackVersionRequest{Slug: slug, Credits: 2_000, DiscountPercent: "8", ValidityMonths: 6, SortOrder: 999}, testActor)
	require.NoError(t, err)
	require.Equal(t, 2, second.Version)
	require.NoError(t, store.PublishCreditPackVersion(ctx, slug, second.Version, testActor))

	require.ErrorIs(t, store.PublishCreditPackVersion(ctx, slug, first.Version, testActor), billing.ErrCreditPackVersionNotFound, "archived versions cannot come back")

	packs, err := store.ListCreditPacks(ctx)
	require.NoError(t, err)

	found := 0

	for _, pack := range packs {
		if pack.Slug == slug {
			found++

			require.Equal(t, 2, pack.Version)
			require.Equal(t, 6, pack.ValidityMonths)
		}
	}

	require.Equal(t, 1, found, "one published version per slug")
}

func TestIntegration_CreditAdjustmentsAndLedgerViewer(t *testing.T) {
	store, pool := newCreditIntegrationStore(t)
	ctx := context.Background()

	tenantID := newCreditLedgerTenant(t, pool, credits.BillingUnitIRR)

	_, err := store.AdjustCredits(ctx, tenantID, platform.CreditAdjustmentRequest{AmountCredits: "1.005", Reason: "x", IdempotencyKey: "k"}, testActor)
	require.ErrorIs(t, err, billing.ErrInvalidCreditAdjustment)

	key := uuid.NewString()
	grant := platform.CreditAdjustmentRequest{AmountCredits: "100", Reason: "goodwill for an outage", IdempotencyKey: key}

	overview, err := store.AdjustCredits(ctx, tenantID, grant, testActor)
	require.NoError(t, err)
	require.Equal(t, "100.00", overview.AvailableCredits)

	overview, err = store.AdjustCredits(ctx, tenantID, grant, testActor)
	require.NoError(t, err)
	require.Equal(t, "100.00", overview.AvailableCredits, "a retried adjustment applies once")

	overview, err = store.AdjustCredits(ctx, tenantID, platform.CreditAdjustmentRequest{AmountCredits: "-150.50", Reason: "reversal of a double grant", IdempotencyKey: uuid.NewString()}, testActor)
	require.NoError(t, err)
	require.Equal(t, "0.00", overview.AvailableCredits)
	require.Equal(t, "50.50", overview.DebtCredits)
	require.Equal(t, "irr", overview.BillingUnit)

	entries, err := store.ListTenantCreditLedger(ctx, tenantID, 10, nil)
	require.NoError(t, err)
	require.Len(t, entries, 2)
	require.Equal(t, "-150.50", entries[0].AmountCredits)
	require.Equal(t, "reversal of a double grant", entries[0].Reason)
	require.Equal(t, testActor.ID, entries[0].Actor)

	older, err := store.ListTenantCreditLedger(ctx, tenantID, 10, &entries[0].OccurredAt)
	require.NoError(t, err)
	require.Len(t, older, 1)

	lots, err := store.ListTenantCreditLots(ctx, tenantID)
	require.NoError(t, err)
	require.Len(t, lots, 1)
	require.Equal(t, "adjust", lots[0].Source)
	require.Equal(t, "0.00", lots[0].RemainingCredits)

	allocations, err := store.ListCreditLotAllocations(ctx, tenantID, lots[0].LotID)
	require.NoError(t, err)
	require.Len(t, allocations, 2)
	require.Equal(t, "100.00", allocations[0].AmountCredits)
	require.Equal(t, "-100.00", allocations[1].AmountCredits)

	require.NoError(t, store.SetTenantBillingUnitAudited(ctx, tenantID, platform.SetBillingUnitRequest{BillingUnit: "credit", Reason: "pilot"}, testActor))
	require.ErrorIs(t, store.SetTenantBillingUnitAudited(ctx, tenantID, platform.SetBillingUnitRequest{BillingUnit: "irr"}, testActor), billing.ErrDecisionReasonRequired)

	var audits int
	require.NoError(t, pool.QueryRow(ctx, `
		SELECT COUNT(1) FROM financial_audit_events WHERE tenant_id = $1::uuid AND event_type IN ('credit.adjusted', 'credit.billing_unit_changed')
	`, tenantID).Scan(&audits))
	require.Equal(t, 3, audits)

	requireTenantReconciles(t, store, tenantID)
}
