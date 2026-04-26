package storage

import (
	"context"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/google/uuid"
)

// GetLatestExchangeRate returns the newest rate snapshot.
func (s *Store) GetLatestExchangeRate(ctx context.Context) (*platform.ExchangeRate, error) {
	var rate platform.ExchangeRate
	err := s.pool.QueryRow(ctx, `
		SELECT rate_id::text, usd_to_irr, source, fetched_at
		FROM exchange_rates
		ORDER BY fetched_at DESC
		LIMIT 1
	`).Scan(&rate.ID, &rate.USDtoIRR, &rate.Source, &rate.FetchedAt)
	if err != nil {
		return nil, err
	}
	return &rate, nil
}

// UpsertExchangeRate stores a rate snapshot.
func (s *Store) UpsertExchangeRate(ctx context.Context, rate platform.ExchangeRate) (*platform.ExchangeRate, error) {
	if rate.ID == "" {
		rate.ID = uuid.NewString()
	}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO exchange_rates (rate_id, usd_to_irr, source, fetched_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING rate_id::text, usd_to_irr, source, fetched_at
	`, rate.ID, rate.USDtoIRR, rate.Source).Scan(&rate.ID, &rate.USDtoIRR, &rate.Source, &rate.FetchedAt)
	if err != nil {
		return nil, err
	}
	return &rate, nil
}
