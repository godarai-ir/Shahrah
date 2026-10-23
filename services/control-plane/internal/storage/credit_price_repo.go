package storage

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	billingdomain "github.com/Delpak/shahrah/services/control-plane/internal/domain/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

// maxExchangeRateSampleAge keeps a stalled rate feed from repeating one rate
// into many samples: a stale rate is skipped, and the missing day then blocks
// the 7-day average instead of skewing it.
const maxExchangeRateSampleAge = 6 * time.Hour

const creditPriceRevisionColumns = `
	r.revision_id::text, r.price_toman_per_credit, r.previous_price_toman, r.base_usd_toman::text,
	r.observed_avg_usd_toman::text, r.band_pct::text, r.status, r.effective_from, r.plan_effective_from,
	COALESCE(r.proposed_by, ''), COALESCE(r.decided_by, ''), r.decided_at, COALESCE(r.decision_reason, ''), r.created_at`

func scanCreditPriceRevision(row pgx.Row) (*platform.CreditPriceRevision, error) {
	var (
		rev      platform.CreditPriceRevision
		previous *int64
	)

	err := row.Scan(&rev.ID, &rev.PriceTomanPerCredit, &previous, &rev.BaseUSDToman,
		&rev.ObservedAvgUSDToman, &rev.BandPercent, &rev.Status, &rev.EffectiveFrom, &rev.PlanEffectiveFrom,
		&rev.ProposedBy, &rev.DecidedBy, &rev.DecidedAt, &rev.DecisionReason, &rev.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, billingdomain.ErrCreditPriceRevisionNotFound
	}

	if err != nil {
		return nil, err
	}

	rev.PreviousPriceToman = previous

	return &rev, nil
}

// creditListPriceIRR is the active list price of one credit in IRR. Before the
// price policy has any revision it falls back to the launch price.
func creditListPriceIRR(ctx context.Context, q rowQuerier) (int64, error) {
	var priceToman int64

	err := q.QueryRow(ctx, `SELECT price_toman_per_credit FROM credit_price_revisions WHERE status = 'active'`).Scan(&priceToman)
	if errors.Is(err, pgx.ErrNoRows) {
		return credits.ListPriceIRRPerCredit, nil
	}

	if err != nil {
		return 0, err
	}

	return priceToman * credits.IRRPerToman, nil
}

// RecordExchangeRateSample stores the rate the pricing engine currently uses as
// this hour's sample. A rate older than maxExchangeRateSampleAge is skipped.
func (s *CreditPriceRepository) RecordExchangeRateSample(ctx context.Context, now time.Time) (bool, error) {
	tag, err := s.pool.Exec(ctx, `
		INSERT INTO exchange_rate_samples (sample_hour, source, usd_to_irr, rate_fetched_at)
		SELECT $1, latest.source, latest.usd_to_irr, latest.fetched_at
		FROM (
			SELECT source, usd_to_irr, fetched_at FROM exchange_rate_current
			ORDER BY fetched_at DESC LIMIT 1
		) latest
		WHERE latest.usd_to_irr > 0 AND latest.fetched_at >= $2
		ON CONFLICT (sample_hour) DO NOTHING
	`, now.UTC().Truncate(time.Hour), now.UTC().Add(-maxExchangeRateSampleAge))
	if err != nil {
		return false, err
	}

	return tag.RowsAffected() == 1, nil
}

// sevenDayAverageUSDToman averages the stored samples per the price policy.
func (s *CreditPriceRepository) sevenDayAverageUSDToman(ctx context.Context, now time.Time) (decimal.Decimal, bool, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT sample_hour, usd_to_irr::text FROM exchange_rate_samples
		WHERE sample_hour >= $1 AND sample_hour < $2
	`, now.UTC().AddDate(0, 0, -(credits.PriceAverageDays+1)), now.UTC())
	if err != nil {
		return decimal.Zero, false, err
	}

	samples, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (credits.RateSample, error) {
		var (
			sample credits.RateSample
			raw    string
		)

		if scanErr := row.Scan(&sample.At, &raw); scanErr != nil {
			return sample, scanErr
		}

		irr, parseErr := decimal.NewFromString(raw)
		sample.USDToman = irr.Div(decimal.NewFromInt(credits.IRRPerToman))

		return sample, parseErr
	})
	if err != nil {
		return decimal.Zero, false, err
	}

	avg, ok := credits.SevenDayAverage(samples, now)

	return avg, ok, nil
}

// ProposeCreditPrice proposes a new list price when the 7-day average left the
// band. An open proposal is refreshed rather than duplicated. The bool reports
// whether a change is due.
func (s *CreditPriceRepository) ProposeCreditPrice(ctx context.Context, now time.Time) (*platform.CreditPriceRevision, bool, error) {
	avg, ok, err := s.sevenDayAverageUSDToman(ctx, now)
	if err != nil || !ok {
		return nil, false, err
	}

	active, err := scanCreditPriceRevision(s.pool.QueryRow(ctx, `SELECT `+creditPriceRevisionColumns+` FROM credit_price_revisions r WHERE r.status = 'active'`))
	if err != nil {
		return nil, false, err
	}

	base, err := decimal.NewFromString(active.BaseUSDToman)
	if err != nil {
		return nil, false, err
	}

	proposal, due := credits.ProposePrice(active.PriceTomanPerCredit, base, avg)
	if !due {
		return nil, false, nil
	}

	rev, err := scanCreditPriceRevision(s.pool.QueryRow(ctx, `
		WITH refreshed AS (
			UPDATE credit_price_revisions
			SET price_toman_per_credit = $1, previous_price_toman = $2, base_usd_toman = $3,
			    observed_avg_usd_toman = $3, updated_at = NOW()
			WHERE status = 'proposed'
			RETURNING *
		), inserted AS (
			INSERT INTO credit_price_revisions (
				price_toman_per_credit, previous_price_toman, base_usd_toman, observed_avg_usd_toman,
				band_pct, status, proposed_by
			)
			SELECT $1, $2, $3, $3, $4, 'proposed', 'price_policy'
			WHERE NOT EXISTS (SELECT 1 FROM refreshed)
			RETURNING *
		)
		SELECT `+creditPriceRevisionColumns+` FROM (SELECT * FROM refreshed UNION ALL SELECT * FROM inserted) r
	`, proposal.PriceToman, active.PriceTomanPerCredit, proposal.BaseUSDToman.String(), credits.PriceBandPercent))

	return rev, err == nil, err
}

// GetCreditPriceStatus returns the active price, any open proposal, the
// current 7-day average, and recent revisions.
func (s *CreditPriceRepository) GetCreditPriceStatus(ctx context.Context, now time.Time) (*platform.CreditPriceStatus, error) {
	rows, err := s.pool.Query(ctx, `SELECT `+creditPriceRevisionColumns+` FROM credit_price_revisions r ORDER BY r.created_at DESC LIMIT 50`)
	if err != nil {
		return nil, err
	}

	revisions, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (platform.CreditPriceRevision, error) {
		rev, scanErr := scanCreditPriceRevision(row)
		if scanErr != nil {
			return platform.CreditPriceRevision{}, scanErr
		}

		return *rev, nil
	})
	if err != nil {
		return nil, err
	}

	status := &platform.CreditPriceStatus{Revisions: revisions}

	for i := range revisions {
		switch revisions[i].Status {
		case "active":
			status.Active = &revisions[i]
		case "proposed":
			status.Proposed = &revisions[i]
		}
	}

	avg, ok, err := s.sevenDayAverageUSDToman(ctx, now)
	if err != nil {
		return nil, err
	}

	if ok {
		formatted := avg.StringFixed(2)
		status.SevenDayAverageUSDToman = &formatted
	}

	return status, nil
}

// ApproveCreditPriceRevision makes a proposal the active list price. Add-on
// packs follow it at once; plan prices get draft versions that publish after
// the notice period, so subscribers see the change from their next period.
func (s *CreditPriceRepository) ApproveCreditPriceRevision(ctx context.Context, revisionID string, actor platform.SupportActor, reason string, now time.Time) (*platform.CreditPriceRevision, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return nil, billingdomain.ErrDecisionReasonRequired
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	proposal, err := scanCreditPriceRevision(tx.QueryRow(ctx, `
		SELECT `+creditPriceRevisionColumns+` FROM credit_price_revisions r
		WHERE r.revision_id::text = $1 AND r.status = 'proposed' FOR UPDATE
	`, revisionID))
	if err != nil {
		return nil, err
	}

	var activePrice int64
	if activeErr := tx.QueryRow(ctx, `
		UPDATE credit_price_revisions SET status = 'superseded', updated_at = NOW()
		WHERE status = 'active'
		RETURNING price_toman_per_credit
	`).Scan(&activePrice); activeErr != nil && !errors.Is(activeErr, pgx.ErrNoRows) {
		return nil, activeErr
	}

	planEffective := now.UTC().AddDate(0, 0, credits.PlanPriceNoticeDays)

	if _, execErr := tx.Exec(ctx, `
		UPDATE credit_price_revisions
		SET status = 'active', effective_from = $2, plan_effective_from = $3,
		    decided_by = NULLIF($4, ''), decided_at = $2, decision_reason = $5, updated_at = NOW()
		WHERE revision_id = $1::uuid
	`, proposal.ID, now.UTC(), planEffective, actor.ID, reason); execErr != nil {
		return nil, execErr
	}

	scheduled, err := scheduleRevisionPlanVersionsTx(ctx, tx, proposal.ID, proposal.PriceTomanPerCredit, planEffective)
	if err != nil {
		return nil, err
	}

	if auditErr := insertFinancialAuditTx(ctx, tx, "", "credit_price.approved", actor, "credit_price_revision", proposal.ID, "credit_price:"+proposal.ID,
		map[string]interface{}{
			"price_toman_per_credit": proposal.PriceTomanPerCredit, "previous_price_toman": activePrice,
			"plan_versions_scheduled": scheduled, "plan_effective_from": planEffective, "reason": reason,
		}); auditErr != nil {
		return nil, auditErr
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		return nil, commitErr
	}

	return scanCreditPriceRevision(s.pool.QueryRow(ctx, `SELECT `+creditPriceRevisionColumns+` FROM credit_price_revisions r WHERE r.revision_id = $1::uuid`, proposal.ID))
}

// RejectCreditPriceRevision closes a proposal without changing the price.
func (s *CreditPriceRepository) RejectCreditPriceRevision(ctx context.Context, revisionID string, actor platform.SupportActor, reason string) (*platform.CreditPriceRevision, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return nil, billingdomain.ErrDecisionReasonRequired
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	rejected, err := scanCreditPriceRevision(tx.QueryRow(ctx, `
		WITH rejected AS (
			UPDATE credit_price_revisions
			SET status = 'rejected', decided_by = NULLIF($2, ''), decided_at = NOW(), decision_reason = $3, updated_at = NOW()
			WHERE revision_id::text = $1 AND status = 'proposed'
			RETURNING *
		)
		SELECT `+creditPriceRevisionColumns+` FROM rejected r
	`, revisionID, actor.ID, reason))
	if err != nil {
		return nil, err
	}

	if auditErr := insertFinancialAuditTx(ctx, tx, "", "credit_price.rejected", actor, "credit_price_revision", rejected.ID, "",
		map[string]interface{}{"price_toman_per_credit": rejected.PriceTomanPerCredit, "reason": reason}); auditErr != nil {
		return nil, auditErr
	}

	return rejected, tx.Commit(ctx)
}

// scheduleRevisionPlanVersionsTx drafts a new version of every published
// credit plan, with prices scaled from the list price that version was priced
// at to the new one. The drafts publish at planEffective.
func scheduleRevisionPlanVersionsTx(ctx context.Context, tx pgx.Tx, revisionID string, newPriceToman int64, planEffective time.Time) (int, error) {
	rows, err := tx.Query(ctx, `
		SELECT pv.version_id::text, pv.plan_id::text,
		       COALESCE(
		           (SELECT r.price_toman_per_credit FROM credit_price_revisions r WHERE r.revision_id = pv.credit_price_revision_id),
		           (SELECT r.price_toman_per_credit FROM credit_price_revisions r ORDER BY r.created_at ASC LIMIT 1))
		FROM plan_versions pv
		WHERE pv.status = 'published'
		  AND EXISTS (SELECT 1 FROM plan_entitlements pe WHERE pe.version_id = pv.version_id AND pe.feature_key = $1)
	`, credits.FeatureMonthlyGrant)
	if err != nil {
		return 0, err
	}

	type publishedPlan struct {
		versionID, planID string
		priceToman        int64
	}

	plans, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (publishedPlan, error) {
		var plan publishedPlan

		scanErr := row.Scan(&plan.versionID, &plan.planID, &plan.priceToman)

		return plan, scanErr
	})
	if err != nil {
		return 0, err
	}

	for _, plan := range plans {
		draftID := uuid.NewString()

		if _, execErr := tx.Exec(ctx, `
			INSERT INTO plan_versions (version_id, plan_id, version, status, effective_from, credit_price_revision_id)
			SELECT $1::uuid, $2::uuid, COALESCE(MAX(version), 0) + 1, 'draft', $3, $4::uuid
			FROM plan_versions WHERE plan_id = $2::uuid
		`, draftID, plan.planID, planEffective, revisionID); execErr != nil {
			return 0, execErr
		}

		if _, execErr := tx.Exec(ctx, `
			INSERT INTO plan_entitlements (version_id, feature_key, value_json)
			SELECT $1::uuid, feature_key, value_json FROM plan_entitlements WHERE version_id = $2::uuid
		`, draftID, plan.versionID); execErr != nil {
			return 0, execErr
		}

		if _, execErr := tx.Exec(ctx, `
			INSERT INTO plan_prices (version_id, billing_interval, currency, amount_minor, discount_percent)
			SELECT $1::uuid, billing_interval, currency,
			       CASE WHEN amount_minor IS NULL THEN NULL ELSE ROUND(amount_minor * $3::numeric / $4::numeric) END,
			       discount_percent
			FROM plan_prices WHERE version_id = $2::uuid
		`, draftID, plan.versionID, newPriceToman, plan.priceToman); execErr != nil {
			return 0, execErr
		}
	}

	return len(plans), nil
}

// PublishDueRevisionPlanVersions publishes plan versions drafted by the active
// price revision once their notice period ends. Drafts of a superseded
// revision never publish; the newer revision has its own.
func (s *CreditPriceRepository) PublishDueRevisionPlanVersions(ctx context.Context, now time.Time) (int, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	rows, err := tx.Query(ctx, `
		SELECT pv.version_id::text, pv.plan_id::text
		FROM plan_versions pv
		JOIN credit_price_revisions r ON r.revision_id = pv.credit_price_revision_id AND r.status = 'active'
		WHERE pv.status = 'draft' AND pv.effective_from <= $1
		FOR UPDATE OF pv
	`, now.UTC())
	if err != nil {
		return 0, err
	}

	type dueVersion struct{ versionID, planID string }

	due, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (dueVersion, error) {
		var version dueVersion

		scanErr := row.Scan(&version.versionID, &version.planID)

		return version, scanErr
	})
	if err != nil {
		return 0, err
	}

	for _, version := range due {
		if _, execErr := tx.Exec(ctx, `
			UPDATE plan_versions SET status = 'archived', updated_at = NOW()
			WHERE plan_id = $1::uuid AND status = 'published'
		`, version.planID); execErr != nil {
			return 0, execErr
		}

		if _, execErr := tx.Exec(ctx, `
			UPDATE plan_versions SET status = 'published', published_at = NOW(), updated_at = NOW()
			WHERE version_id = $1::uuid
		`, version.versionID); execErr != nil {
			return 0, execErr
		}
	}

	return len(due), tx.Commit(ctx)
}

// ListPricingChangelog returns every public rate-card version and decided
// credit price, newest first. It never exposes the USD conversion factor.
func (s *CreditPriceRepository) ListPricingChangelog(ctx context.Context) (*platform.PricingChangelog, error) {
	changelog := &platform.PricingChangelog{RateCards: []platform.PricingChangelogRateCard{}, CreditPrices: []platform.PricingChangelogPrice{}}

	rows, err := s.pool.Query(ctx, `
		SELECT version, status, effective_from, changelog_fa, changelog_en
		FROM credit_rate_card_versions
		WHERE status IN ('scheduled', 'published', 'superseded')
		ORDER BY effective_from DESC, version DESC
	`)
	if err != nil {
		return nil, err
	}

	changelog.RateCards, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (platform.PricingChangelogRateCard, error) {
		var card platform.PricingChangelogRateCard

		scanErr := row.Scan(&card.Version, &card.Status, &card.EffectiveFrom, &card.ChangelogFA, &card.ChangelogEN)
		card.Label = credits.RateCard{Version: card.Version}.Label()

		return card, scanErr
	})
	if err != nil {
		return nil, err
	}

	rows, err = s.pool.Query(ctx, `
		SELECT price_toman_per_credit, status, effective_from, COALESCE(plan_effective_from, effective_from)
		FROM credit_price_revisions
		WHERE status IN ('active', 'superseded') AND effective_from IS NOT NULL
		ORDER BY effective_from DESC
	`)
	if err != nil {
		return nil, err
	}

	changelog.CreditPrices, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (platform.PricingChangelogPrice, error) {
		var price platform.PricingChangelogPrice

		scanErr := row.Scan(&price.PriceTomanPerCredit, &price.Status, &price.EffectiveFrom, &price.PlanEffectiveFrom)

		return price, scanErr
	})

	return changelog, err
}
