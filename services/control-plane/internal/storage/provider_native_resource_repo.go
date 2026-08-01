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
