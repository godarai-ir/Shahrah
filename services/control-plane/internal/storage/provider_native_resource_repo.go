package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/jackc/pgx/v5"
)

var ErrProviderNativeResourceNotFound = errors.New("provider native resource not found")

func (s *Store) UpsertProviderNativeResourceMapping(
	ctx context.Context,
	mapping platform.ProviderNativeResourceMapping,
) (*platform.ProviderNativeResourceMapping, error) {
	mapping.Provider = strings.ToLower(strings.TrimSpace(mapping.Provider))

	mapping.ResourceKind = strings.ToLower(strings.TrimSpace(mapping.ResourceKind))
	if mapping.PublicID == "" || mapping.TenantID == "" ||
		mapping.Provider == "" || mapping.ProviderAccountID == "" ||
		mapping.ResourceKind == "" || mapping.UpstreamID == "" {
		return nil, fmt.Errorf("provider native resource mapping is incomplete")
	}

	metadata, err := json.Marshal(mapping.Metadata)
	if err != nil {
		return nil, err
	}

	reservation, err := json.Marshal(mapping.Reservation)
	if err != nil {
		return nil, err
	}

	estimated, err := json.Marshal(mapping.EstimatedUsage)
	if err != nil {
		return nil, err
	}

	err = s.pool.QueryRow(ctx, `
		INSERT INTO provider_native_resource_mappings (
			public_id, tenant_id, provider, provider_account_id, resource_kind,
			upstream_id, status, metadata_json, pricing_revision,
			reservation_id, reservation_json, estimated_usage_json, expires_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, COALESCE(NULLIF($7, ''), 'active'),
			$8::jsonb, NULLIF($9, ''), NULLIF($10, ''), $11::jsonb, $12::jsonb, $13
		)
		ON CONFLICT (tenant_id, provider, provider_account_id, resource_kind, upstream_id)
		DO UPDATE SET
			status = EXCLUDED.status,
			metadata_json = EXCLUDED.metadata_json,
			pricing_revision = COALESCE(EXCLUDED.pricing_revision, provider_native_resource_mappings.pricing_revision),
			reservation_id = COALESCE(EXCLUDED.reservation_id, provider_native_resource_mappings.reservation_id),
			reservation_json = COALESCE(EXCLUDED.reservation_json, provider_native_resource_mappings.reservation_json),
			estimated_usage_json = COALESCE(EXCLUDED.estimated_usage_json, provider_native_resource_mappings.estimated_usage_json),
			expires_at = EXCLUDED.expires_at,
			updated_at = now()
		RETURNING public_id
	`, mapping.PublicID, mapping.TenantID, mapping.Provider,
		mapping.ProviderAccountID, mapping.ResourceKind, mapping.UpstreamID,
		mapping.Status, metadata, mapping.PricingRevision, mapping.ReservationID,
		reservation, estimated, mapping.ExpiresAt,
	).Scan(&mapping.PublicID)
	if err != nil {
		return nil, err
	}

	return &mapping, nil
}

func (s *Store) GetProviderNativeResourceMapping(
	ctx context.Context,
	tenantID string,
	provider string,
	resourceKind string,
	publicID string,
) (*platform.ProviderNativeResourceMapping, error) {
	var (
		mapping                          platform.ProviderNativeResourceMapping
		metadata, reservation, estimated []byte
	)

	err := s.pool.QueryRow(ctx, `
		SELECT public_id, tenant_id::text, provider, provider_account_id::text,
		       resource_kind, upstream_id, status, metadata_json,
		       COALESCE(pricing_revision, ''), COALESCE(reservation_id, ''),
		       reservation_json, estimated_usage_json, expires_at
		FROM provider_native_resource_mappings
		WHERE tenant_id = $1 AND provider = $2 AND resource_kind = $3 AND public_id = $4
	`, tenantID, strings.ToLower(strings.TrimSpace(provider)),
		strings.ToLower(strings.TrimSpace(resourceKind)), publicID,
	).Scan(
		&mapping.PublicID, &mapping.TenantID, &mapping.Provider,
		&mapping.ProviderAccountID, &mapping.ResourceKind, &mapping.UpstreamID,
		&mapping.Status, &metadata, &mapping.PricingRevision,
		&mapping.ReservationID, &reservation, &estimated, &mapping.ExpiresAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrProviderNativeResourceNotFound
	}

	if err != nil {
		return nil, err
	}

	if err := decodeOptionalJSON(metadata, &mapping.Metadata); err != nil {
		return nil, err
	}

	if err := decodeOptionalJSON(reservation, &mapping.Reservation); err != nil {
		return nil, err
	}

	if err := decodeOptionalJSON(estimated, &mapping.EstimatedUsage); err != nil {
		return nil, err
	}

	return &mapping, nil
}

func (s *Store) GetProviderNativeResourceMappingByUpstream(
	ctx context.Context,
	tenantID string,
	provider string,
	providerAccountID string,
	resourceKind string,
	upstreamID string,
) (*platform.ProviderNativeResourceMapping, error) {
	var (
		mapping                          platform.ProviderNativeResourceMapping
		metadata, reservation, estimated []byte
	)

	err := s.pool.QueryRow(ctx, `
		SELECT public_id, tenant_id::text, provider, provider_account_id::text,
		       resource_kind, upstream_id, status, metadata_json,
		       COALESCE(pricing_revision, ''), COALESCE(reservation_id, ''),
		       reservation_json, estimated_usage_json, expires_at
		FROM provider_native_resource_mappings
		WHERE tenant_id = $1 AND provider = $2 AND provider_account_id = $3
		  AND resource_kind = $4 AND upstream_id = $5
	`, tenantID, strings.ToLower(strings.TrimSpace(provider)), providerAccountID,
		strings.ToLower(strings.TrimSpace(resourceKind)), upstreamID,
	).Scan(
		&mapping.PublicID, &mapping.TenantID, &mapping.Provider,
		&mapping.ProviderAccountID, &mapping.ResourceKind, &mapping.UpstreamID,
		&mapping.Status, &metadata, &mapping.PricingRevision,
		&mapping.ReservationID, &reservation, &estimated, &mapping.ExpiresAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrProviderNativeResourceNotFound
	}

	if err != nil {
		return nil, err
	}

	if err := decodeOptionalJSON(metadata, &mapping.Metadata); err != nil {
		return nil, err
	}

	if err := decodeOptionalJSON(reservation, &mapping.Reservation); err != nil {
		return nil, err
	}

	if err := decodeOptionalJSON(estimated, &mapping.EstimatedUsage); err != nil {
		return nil, err
	}

	return &mapping, nil
}

func (s *Store) ListProviderNativeResourceMappings(
	ctx context.Context,
	tenantID string,
	provider string,
	providerAccountID string,
	resourceKind string,
	statuses []string,
	limit int,
) ([]platform.ProviderNativeResourceMapping, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	rows, err := s.pool.Query(ctx, `
		SELECT public_id, tenant_id::text, provider, provider_account_id::text,
		       resource_kind, upstream_id, status, metadata_json,
		       COALESCE(pricing_revision, ''), COALESCE(reservation_id, ''),
		       reservation_json, estimated_usage_json, expires_at
		FROM provider_native_resource_mappings
		WHERE tenant_id = $1 AND provider = $2 AND provider_account_id = $3
		  AND resource_kind = $4
		  AND (cardinality($5::text[]) = 0 OR status = ANY($5::text[]))
		ORDER BY updated_at ASC
		LIMIT $6
	`, tenantID, strings.ToLower(strings.TrimSpace(provider)), providerAccountID,
		strings.ToLower(strings.TrimSpace(resourceKind)), statuses, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	mappings := make([]platform.ProviderNativeResourceMapping, 0)

	for rows.Next() {
		var (
			mapping                          platform.ProviderNativeResourceMapping
			metadata, reservation, estimated []byte
		)

		if err := rows.Scan(
			&mapping.PublicID, &mapping.TenantID, &mapping.Provider,
			&mapping.ProviderAccountID, &mapping.ResourceKind, &mapping.UpstreamID,
			&mapping.Status, &metadata, &mapping.PricingRevision,
			&mapping.ReservationID, &reservation, &estimated, &mapping.ExpiresAt,
		); err != nil {
			return nil, err
		}

		if err := decodeOptionalJSON(metadata, &mapping.Metadata); err != nil {
			return nil, err
		}

		if err := decodeOptionalJSON(reservation, &mapping.Reservation); err != nil {
			return nil, err
		}

		if err := decodeOptionalJSON(estimated, &mapping.EstimatedUsage); err != nil {
			return nil, err
		}

		mappings = append(mappings, mapping)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return mappings, nil
}

func (s *Store) ClaimProviderNativeResourceStatus(
	ctx context.Context,
	tenantID string,
	provider string,
	resourceKind string,
	publicID string,
	fromStatuses []string,
	toStatus string,
) (bool, error) {
	tag, err := s.pool.Exec(ctx, `
		UPDATE provider_native_resource_mappings
		SET status = $6, updated_at = now()
		WHERE tenant_id = $1 AND provider = $2 AND resource_kind = $3
		  AND public_id = $4
		  AND (status = ANY($5::text[]) OR (status = $6 AND updated_at < now() - interval '5 minutes'))
	`, tenantID, strings.ToLower(strings.TrimSpace(provider)),
		strings.ToLower(strings.TrimSpace(resourceKind)), publicID, fromStatuses,
		strings.TrimSpace(toStatus),
	)
	if err != nil {
		return false, err
	}

	return tag.RowsAffected() == 1, nil
}

func (s *Store) ListPendingProviderNativeResourceMappings(
	ctx context.Context,
	provider string,
	resourceKind string,
	statuses []string,
	limit int,
) ([]platform.ProviderNativeResourceMapping, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	rows, err := s.pool.Query(ctx, `
		SELECT public_id, tenant_id::text, provider, provider_account_id::text,
		       resource_kind, upstream_id, status, metadata_json,
		       COALESCE(pricing_revision, ''), COALESCE(reservation_id, ''),
		       reservation_json, estimated_usage_json, expires_at
		FROM provider_native_resource_mappings
		WHERE provider = $1 AND resource_kind = $2 AND status = ANY($3::text[])
		ORDER BY updated_at ASC
		LIMIT $4
	`, strings.ToLower(strings.TrimSpace(provider)),
		strings.ToLower(strings.TrimSpace(resourceKind)), statuses, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	mappings := make([]platform.ProviderNativeResourceMapping, 0)

	for rows.Next() {
		var (
			mapping                          platform.ProviderNativeResourceMapping
			metadata, reservation, estimated []byte
		)

		if err := rows.Scan(
			&mapping.PublicID, &mapping.TenantID, &mapping.Provider,
			&mapping.ProviderAccountID, &mapping.ResourceKind, &mapping.UpstreamID,
			&mapping.Status, &metadata, &mapping.PricingRevision,
			&mapping.ReservationID, &reservation, &estimated, &mapping.ExpiresAt,
		); err != nil {
			return nil, err
		}

		if err := decodeOptionalJSON(metadata, &mapping.Metadata); err != nil {
			return nil, err
		}

		if err := decodeOptionalJSON(reservation, &mapping.Reservation); err != nil {
			return nil, err
		}

		if err := decodeOptionalJSON(estimated, &mapping.EstimatedUsage); err != nil {
			return nil, err
		}

		mappings = append(mappings, mapping)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return mappings, nil
}

func (s *Store) UpdateProviderNativeResourceStatus(
	ctx context.Context,
	tenantID string,
	provider string,
	resourceKind string,
	publicID string,
	status string,
) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE provider_native_resource_mappings
		SET status = $5, updated_at = now()
		WHERE tenant_id = $1 AND provider = $2 AND resource_kind = $3 AND public_id = $4
	`, tenantID, strings.ToLower(strings.TrimSpace(provider)),
		strings.ToLower(strings.TrimSpace(resourceKind)), publicID, status)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrProviderNativeResourceNotFound
	}

	return nil
}

func decodeOptionalJSON(raw []byte, target interface{}) error {
	if len(raw) == 0 || string(raw) == "null" || string(raw) == "{}" {
		return nil
	}

	return json.Unmarshal(raw, target)
}
