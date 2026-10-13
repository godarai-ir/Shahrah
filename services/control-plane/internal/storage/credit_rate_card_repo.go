package storage

import (
	"context"
	"fmt"

	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/shopspring/decimal"
)

// GetPublishedCreditRateCard loads the published credit rate card and its model multipliers.
func (s *CreditRateCardRepository) GetPublishedCreditRateCard(ctx context.Context) (credits.RateCard, error) {
	var card credits.RateCard

	err := s.pool.QueryRow(ctx, `
		SELECT version, effective_from
		FROM credit_rate_card_versions
		WHERE status = 'published'
	`).Scan(&card.Version, &card.EffectiveFrom)
	if err != nil {
		return credits.RateCard{}, fmt.Errorf("load published credit rate card: %w", err)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT model, multiplier::text
		FROM credit_model_multipliers
		WHERE version = $1
	`, card.Version)
	if err != nil {
		return credits.RateCard{}, fmt.Errorf("load credit model multipliers: %w", err)
	}
	defer rows.Close()

	card.Multipliers = map[string]decimal.Decimal{}

	for rows.Next() {
		var model, raw string
		if err := rows.Scan(&model, &raw); err != nil {
			return credits.RateCard{}, fmt.Errorf("scan credit model multiplier: %w", err)
		}

		multiplier, err := decimal.NewFromString(raw)
		if err != nil {
			return credits.RateCard{}, fmt.Errorf("parse credit model multiplier for %s: %w", model, err)
		}

		card.Multipliers[credits.NormalizeModel(model)] = multiplier
	}

	if err := rows.Err(); err != nil {
		return credits.RateCard{}, fmt.Errorf("iterate credit model multipliers: %w", err)
	}

	return card, nil
}
