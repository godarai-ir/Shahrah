package storage

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	billingdomain "github.com/Delpak/shahrah/services/control-plane/internal/domain/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

// ListCreditRateCardVersions returns every rate-card version with its multipliers.
func (s *CreditRateCardRepository) ListCreditRateCardVersions(ctx context.Context) ([]platform.CreditRateCardVersion, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT version, status, effective_from, notice_until, changelog_fa, changelog_en,
		       COALESCE(published_by, ''), published_at
		FROM credit_rate_card_versions
		ORDER BY version DESC
	`)
	if err != nil {
		return nil, err
	}

	versions, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (platform.CreditRateCardVersion, error) {
		var version platform.CreditRateCardVersion

		scanErr := row.Scan(&version.Version, &version.Status, &version.EffectiveFrom, &version.NoticeUntil,
			&version.ChangelogFA, &version.ChangelogEN, &version.PublishedBy, &version.PublishedAt)
		version.Multipliers = []platform.CreditModelMultiplier{}

		return version, scanErr
	})
	if err != nil {
		return nil, err
	}

	byVersion := make(map[int]*platform.CreditRateCardVersion, len(versions))
	for i := range versions {
		byVersion[versions[i].Version] = &versions[i]
	}

	rows, err = s.pool.Query(ctx, `
		SELECT version, model, multiplier::text, reason, COALESCE(created_by, ''), created_at
		FROM credit_model_multipliers ORDER BY version DESC, model
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			version    int
			multiplier platform.CreditModelMultiplier
		)

		if err := rows.Scan(&version, &multiplier.Model, &multiplier.Multiplier, &multiplier.Reason, &multiplier.CreatedBy, &multiplier.CreatedAt); err != nil {
			return nil, err
		}

		if target, ok := byVersion[version]; ok {
			target.Multipliers = append(target.Multipliers, multiplier)
		}
	}

	return versions, rows.Err()
}

// CreateCreditRateCardDraft starts a draft from the published version's
// multipliers. An existing draft is returned instead of starting a second one.
func (s *CreditRateCardRepository) CreateCreditRateCardDraft(ctx context.Context, actor platform.SupportActor) (int, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	if _, lockErr := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext('credit_rate_card_draft'))`); lockErr != nil {
		return 0, lockErr
	}

	var draft int

	err = tx.QueryRow(ctx, `SELECT version FROM credit_rate_card_versions WHERE status = 'draft' ORDER BY version DESC LIMIT 1`).Scan(&draft)
	if err == nil {
		return draft, tx.Commit(ctx)
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, err
	}

	if scanErr := tx.QueryRow(ctx, `
		INSERT INTO credit_rate_card_versions (version, status)
		SELECT COALESCE(MAX(version), 0) + 1, 'draft' FROM credit_rate_card_versions
		RETURNING version
	`).Scan(&draft); scanErr != nil {
		return 0, scanErr
	}

	if _, copyErr := tx.Exec(ctx, `
		INSERT INTO credit_model_multipliers (version, model, multiplier, reason, created_by)
		SELECT $1, m.model, m.multiplier, m.reason, m.created_by
		FROM credit_model_multipliers m
		JOIN credit_rate_card_versions v ON v.version = m.version AND v.status = 'published'
	`, draft); copyErr != nil {
		return 0, copyErr
	}

	if auditErr := insertFinancialAuditTx(ctx, tx, "", "credit_rate_card.draft_created", actor, "credit_rate_card", "", "", map[string]interface{}{"version": draft}); auditErr != nil {
		return 0, auditErr
	}

	return draft, tx.Commit(ctx)
}

// UpsertCreditRateCardMultiplier sets one model's multiplier on a draft. A
// reason is required, because a multiplier means GodarAI's cost differs from
// the provider's list price.
func (s *CreditRateCardRepository) UpsertCreditRateCardMultiplier(ctx context.Context, version int, req platform.UpsertCreditMultiplierRequest, actor platform.SupportActor) error {
	model := credits.NormalizeModel(req.Model)
	reason := strings.TrimSpace(req.Reason)

	multiplier, err := decimal.NewFromString(strings.TrimSpace(req.Multiplier))
	if err != nil || !multiplier.IsPositive() || !multiplier.Equal(multiplier.Round(4)) || model == "" {
		return billingdomain.ErrInvalidCreditMultiplier
	}

	if reason == "" {
		return billingdomain.ErrDecisionReasonRequired
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer rollbackTxBestEffort(ctx, tx)

	if draftErr := lockDraftRateCardTx(ctx, tx, version); draftErr != nil {
		return draftErr
	}

	if _, execErr := tx.Exec(ctx, `
		INSERT INTO credit_model_multipliers (version, model, multiplier, reason, created_by)
		VALUES ($1, $2, $3::numeric, $4, NULLIF($5, ''))
		ON CONFLICT (version, model) DO UPDATE
		SET multiplier = EXCLUDED.multiplier, reason = EXCLUDED.reason, created_by = EXCLUDED.created_by, created_at = NOW()
	`, version, model, multiplier.String(), reason, actor.ID); execErr != nil {
		return execErr
	}

	if auditErr := insertFinancialAuditTx(ctx, tx, "", "credit_rate_card.multiplier_set", actor, "credit_rate_card", "", "",
		map[string]interface{}{"version": version, "model": model, "multiplier": multiplier.String(), "reason": reason}); auditErr != nil {
		return auditErr
	}

	return tx.Commit(ctx)
}

func lockDraftRateCardTx(ctx context.Context, tx pgx.Tx, version int) error {
	var status string

	err := tx.QueryRow(ctx, `SELECT status FROM credit_rate_card_versions WHERE version = $1 FOR UPDATE`, version).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && status != "draft") {
		return billingdomain.ErrCreditRateCardNotDraft
	}

	return err
}

// PublishCreditRateCard publishes a draft. If any model costs more than in the
// published version, the draft is scheduled after the 30-day notice instead,
// as the PRD requires for provider price increases.
func (s *CreditRateCardRepository) PublishCreditRateCard(ctx context.Context, version int, req platform.PublishCreditRateCardRequest, actor platform.SupportActor, now time.Time) (*platform.CreditRateCardVersion, error) {
	if strings.TrimSpace(req.ChangelogFA) == "" {
		return nil, billingdomain.ErrDecisionReasonRequired
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	if draftErr := lockDraftRateCardTx(ctx, tx, version); draftErr != nil {
		return nil, draftErr
	}

	var increases bool
	if scanErr := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM credit_model_multipliers d
			LEFT JOIN credit_model_multipliers p
			  ON p.model = d.model AND p.version = (SELECT version FROM credit_rate_card_versions WHERE status = 'published')
			WHERE d.version = $1 AND d.multiplier > COALESCE(p.multiplier, 1)
		)
	`, version).Scan(&increases); scanErr != nil {
		return nil, scanErr
	}

	status, effective := "published", now.UTC()
	if increases {
		status, effective = "scheduled", now.UTC().AddDate(0, 0, credits.PlanPriceNoticeDays)
	}

	if _, execErr := tx.Exec(ctx, `
		UPDATE credit_rate_card_versions
		SET changelog_fa = $2, changelog_en = $3, effective_from = $4::timestamptz,
		    notice_until = CASE WHEN $5::text = 'scheduled' THEN $4::timestamptz ELSE NULL END,
		    published_by = NULLIF($6, ''), updated_at = NOW()
		WHERE version = $1
	`, version, strings.TrimSpace(req.ChangelogFA), strings.TrimSpace(req.ChangelogEN), effective, status, actor.ID); execErr != nil {
		return nil, execErr
	}

	if status == "published" {
		if publishErr := activateRateCardTx(ctx, tx, version); publishErr != nil {
			return nil, publishErr
		}
	} else if _, execErr := tx.Exec(ctx, `UPDATE credit_rate_card_versions SET status = 'scheduled' WHERE version = $1`, version); execErr != nil {
		return nil, execErr
	}

	if auditErr := insertFinancialAuditTx(ctx, tx, "", "credit_rate_card."+status, actor, "credit_rate_card", "", "",
		map[string]interface{}{"version": version, "effective_from": effective, "price_increase": increases}); auditErr != nil {
		return nil, auditErr
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		return nil, commitErr
	}

	versions, err := s.ListCreditRateCardVersions(ctx)
	if err != nil {
		return nil, err
	}

	for i := range versions {
		if versions[i].Version == version {
			return &versions[i], nil
		}
	}

	return nil, billingdomain.ErrCreditRateCardNotDraft
}

// activateRateCardTx makes one version the published rate card.
func activateRateCardTx(ctx context.Context, tx pgx.Tx, version int) error {
	if _, err := tx.Exec(ctx, `
		UPDATE credit_rate_card_versions SET status = 'superseded', updated_at = NOW() WHERE status = 'published'
	`); err != nil {
		return err
	}

	_, err := tx.Exec(ctx, `
		UPDATE credit_rate_card_versions SET status = 'published', published_at = NOW(), updated_at = NOW() WHERE version = $1
	`, version)

	return err
}

// PublishDueRateCards publishes scheduled rate cards whose notice has ended.
func (s *CreditRateCardRepository) PublishDueRateCards(ctx context.Context, now time.Time) (int, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	var version int

	err = tx.QueryRow(ctx, `
		SELECT version FROM credit_rate_card_versions
		WHERE status = 'scheduled' AND effective_from <= $1
		ORDER BY effective_from DESC, version DESC LIMIT 1
		FOR UPDATE
	`, now.UTC()).Scan(&version)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}

	if err != nil {
		return 0, err
	}

	if activateErr := activateRateCardTx(ctx, tx, version); activateErr != nil {
		return 0, activateErr
	}

	// An older schedule that came due in the same window is replaced by the newest.
	if _, execErr := tx.Exec(ctx, `
		UPDATE credit_rate_card_versions SET status = 'superseded', updated_at = NOW()
		WHERE status = 'scheduled' AND effective_from <= $1
	`, now.UTC()); execErr != nil {
		return 0, execErr
	}

	return 1, tx.Commit(ctx)
}
