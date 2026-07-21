package storage

import (
	"context"
	"errors"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// GetLatestExchangeRate returns the newest current rate, with legacy history fallback.
func (s *Store) GetLatestExchangeRate(ctx context.Context) (*platform.ExchangeRate, error) {
	rate, err := s.getLatestCurrentExchangeRate(ctx)
	if err == nil {
		return rate, nil
	}

	if !shouldFallbackToLegacyExchangeRates(err) {
		return nil, err
	}

	return s.getLatestLegacyExchangeRate(ctx)
}

func (s *Store) getLatestCurrentExchangeRate(ctx context.Context) (*platform.ExchangeRate, error) {
	var rate platform.ExchangeRate

	err := s.pool.QueryRow(ctx, `
		SELECT rate_id::text, usd_to_irr, source, fetched_at
		FROM exchange_rate_current
		ORDER BY fetched_at DESC
		LIMIT 1
	`).Scan(&rate.ID, &rate.USDtoIRR, &rate.Source, &rate.FetchedAt)
	if err != nil {
		return nil, err
	}

	return &rate, nil
}

func (s *Store) getLatestLegacyExchangeRate(ctx context.Context) (*platform.ExchangeRate, error) {
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

// UpsertExchangeRate stores the latest exchange rate for a source.
func (s *Store) UpsertExchangeRate(ctx context.Context, rate platform.ExchangeRate) (*platform.ExchangeRate, error) {
	if rate.ID == "" {
		rate.ID = uuid.NewString()
	}

	source := strings.TrimSpace(rate.Source)
	if source == "" {
		source = "default"
	}

	err := s.pool.QueryRow(ctx, `
		INSERT INTO exchange_rate_current (rate_id, source, symbol, usd_to_irr, fetched_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT (source) DO UPDATE SET
			symbol = EXCLUDED.symbol,
			usd_to_irr = EXCLUDED.usd_to_irr,
			fetched_at = EXCLUDED.fetched_at,
			updated_at = NOW()
		RETURNING rate_id::text, usd_to_irr, source, fetched_at
	`, rate.ID, source, exchangeRateSymbolFromSource(source), rate.USDtoIRR).Scan(&rate.ID, &rate.USDtoIRR, &rate.Source, &rate.FetchedAt)
	if err != nil {
		if !isUndefinedTableErr(err) {
			return nil, err
		}

		return s.insertLegacyExchangeRate(ctx, rate, source)
	}

	return &rate, nil
}

func (s *Store) insertLegacyExchangeRate(ctx context.Context, rate platform.ExchangeRate, source string) (*platform.ExchangeRate, error) {
	err := s.pool.QueryRow(ctx, `
		INSERT INTO exchange_rates (rate_id, usd_to_irr, source, fetched_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING rate_id::text, usd_to_irr, source, fetched_at
	`, rate.ID, rate.USDtoIRR, source).Scan(&rate.ID, &rate.USDtoIRR, &rate.Source, &rate.FetchedAt)
	if err != nil {
		return nil, err
	}

	return &rate, nil
}

func exchangeRateSymbolFromSource(source string) *string {
	_, symbol, ok := strings.Cut(strings.TrimSpace(source), ":")
	if !ok || strings.TrimSpace(symbol) == "" {
		return nil
	}

	normalized := strings.ToUpper(strings.TrimSpace(symbol))

	return &normalized
}

func shouldFallbackToLegacyExchangeRates(err error) bool {
	return errors.Is(err, pgx.ErrNoRows) || isUndefinedTableErr(err)
}

func isUndefinedTableErr(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "42P01"
}
