package storage

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/jackc/pgx/v5"
)

var ErrAnthropicNativeResourceNotFound = errors.New("anthropic native resource not found")

func (s *Store) UpsertAnthropicFileMapping(ctx context.Context, mapping platform.AnthropicNativeResourceMapping) (*platform.AnthropicNativeResourceMapping, error) {
	err := s.pool.QueryRow(ctx, `
		INSERT INTO anthropic_file_mappings (
			public_id, tenant_id, provider_account_id, upstream_id, status
		) VALUES ($1, $2, $3, $4, COALESCE(NULLIF($5, ''), 'active'))
		ON CONFLICT (tenant_id, provider_account_id, upstream_id) DO UPDATE
		SET status = EXCLUDED.status, updated_at = now()
		RETURNING public_id, tenant_id::text, provider_account_id::text, upstream_id, status
	`, mapping.PublicID, mapping.TenantID, mapping.ProviderAccountID, mapping.UpstreamID, mapping.Status).Scan(
		&mapping.PublicID, &mapping.TenantID, &mapping.ProviderAccountID,
		&mapping.UpstreamID, &mapping.Status,
	)

	return &mapping, err
}

func (s *Store) GetAnthropicFileMapping(ctx context.Context, tenantID, publicID string) (*platform.AnthropicNativeResourceMapping, error) {
	var mapping platform.AnthropicNativeResourceMapping

	err := s.pool.QueryRow(ctx, `
		SELECT public_id, tenant_id::text, provider_account_id::text, upstream_id, status
		FROM anthropic_file_mappings
		WHERE tenant_id = $1 AND public_id = $2
	`, tenantID, publicID).Scan(
		&mapping.PublicID, &mapping.TenantID, &mapping.ProviderAccountID,
		&mapping.UpstreamID, &mapping.Status,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrAnthropicNativeResourceNotFound
	}

	return &mapping, err
}

func (s *Store) UpsertAnthropicBatchMapping(ctx context.Context, mapping platform.AnthropicNativeResourceMapping) (*platform.AnthropicNativeResourceMapping, error) {
	estimated, err := json.Marshal(mapping.EstimatedUsage)
	if err != nil {
		return nil, err
	}

	reservation, err := json.Marshal(mapping.Reservation)
	if err != nil {
		return nil, err
	}

	err = s.pool.QueryRow(ctx, `
		INSERT INTO anthropic_batch_mappings (
			public_id, tenant_id, provider_account_id, upstream_id, status,
			reservation_id, reservation_json, pricing_revision, estimated_usage_json
		) VALUES (
			$1, $2, $3, $4, COALESCE(NULLIF($5, ''), 'in_progress'),
			NULLIF($6, ''), $7::jsonb, NULLIF($8, ''), $9::jsonb
		)
		ON CONFLICT (tenant_id, provider_account_id, upstream_id) DO UPDATE
		SET status = EXCLUDED.status,
		    reservation_id = COALESCE(EXCLUDED.reservation_id, anthropic_batch_mappings.reservation_id),
		    reservation_json = COALESCE(EXCLUDED.reservation_json, anthropic_batch_mappings.reservation_json),
		    pricing_revision = COALESCE(EXCLUDED.pricing_revision, anthropic_batch_mappings.pricing_revision),
		    estimated_usage_json = CASE
		        WHEN EXCLUDED.estimated_usage_json = 'null'::jsonb
		        THEN anthropic_batch_mappings.estimated_usage_json
		        ELSE EXCLUDED.estimated_usage_json
		    END,
		    updated_at = now()
		RETURNING public_id
	`, mapping.PublicID, mapping.TenantID, mapping.ProviderAccountID, mapping.UpstreamID,
		mapping.Status, mapping.ReservationID, reservation, mapping.PricingRevision, estimated).Scan(&mapping.PublicID)

	return &mapping, err
}

func (s *Store) GetAnthropicBatchMapping(ctx context.Context, tenantID, publicID string) (*platform.AnthropicNativeResourceMapping, error) {
	var (
		mapping                platform.AnthropicNativeResourceMapping
		estimated, reservation []byte
	)

	err := s.pool.QueryRow(ctx, `
		SELECT public_id, tenant_id::text, provider_account_id::text, upstream_id, status,
		       COALESCE(reservation_id, ''), reservation_json, COALESCE(pricing_revision, ''),
		       estimated_usage_json
		FROM anthropic_batch_mappings
		WHERE tenant_id = $1 AND public_id = $2
	`, tenantID, publicID).Scan(
		&mapping.PublicID, &mapping.TenantID, &mapping.ProviderAccountID,
		&mapping.UpstreamID, &mapping.Status, &mapping.ReservationID,
		&reservation, &mapping.PricingRevision, &estimated,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrAnthropicNativeResourceNotFound
	}

	if err != nil {
		return nil, err
	}

	if len(estimated) > 0 && string(estimated) != "null" && string(estimated) != "{}" {
		var breakdown platform.UsageCostBreakdown
		if err := json.Unmarshal(estimated, &breakdown); err != nil {
			return nil, err
		}

		mapping.EstimatedUsage = &breakdown
	}

	if len(reservation) > 0 && string(reservation) != "null" {
		var value platform.BudgetReservation
		if err := json.Unmarshal(reservation, &value); err != nil {
			return nil, err
		}

		mapping.Reservation = &value
	}

	return &mapping, nil
}

func (s *Store) UpdateAnthropicBatchStatus(ctx context.Context, tenantID, publicID, status string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE anthropic_batch_mappings
		SET status = $3,
		    updated_at = now(),
		    ended_at = CASE
		        WHEN $3 IN ('ended', 'canceled', 'expired', 'failed') THEN now()
		        ELSE ended_at
		    END
		WHERE tenant_id = $1 AND public_id = $2
	`, tenantID, publicID, status)

	return err
}

func (s *Store) ListPendingAnthropicBatchMappings(ctx context.Context, limit int) ([]platform.AnthropicNativeResourceMapping, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	rows, err := s.pool.Query(ctx, `
		SELECT public_id, tenant_id::text, provider_account_id::text, upstream_id, status,
		       COALESCE(reservation_id, ''), reservation_json,
		       COALESCE(pricing_revision, ''), estimated_usage_json
		FROM anthropic_batch_mappings
		WHERE status IN ('in_progress', 'canceling')
		ORDER BY updated_at ASC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]platform.AnthropicNativeResourceMapping, 0)

	for rows.Next() {
		var (
			mapping                platform.AnthropicNativeResourceMapping
			reservation, estimated []byte
		)
		if err := rows.Scan(
			&mapping.PublicID, &mapping.TenantID, &mapping.ProviderAccountID,
			&mapping.UpstreamID, &mapping.Status, &mapping.ReservationID,
			&reservation, &mapping.PricingRevision, &estimated,
		); err != nil {
			return nil, err
		}

		if len(reservation) > 0 && string(reservation) != "null" {
			var value platform.BudgetReservation
			if err := json.Unmarshal(reservation, &value); err != nil {
				return nil, err
			}

			mapping.Reservation = &value
		}

		if len(estimated) > 0 && string(estimated) != "null" && string(estimated) != "{}" {
			var value platform.UsageCostBreakdown
			if err := json.Unmarshal(estimated, &value); err != nil {
				return nil, err
			}

			mapping.EstimatedUsage = &value
		}

		out = append(out, mapping)
	}

	return out, rows.Err()
}
