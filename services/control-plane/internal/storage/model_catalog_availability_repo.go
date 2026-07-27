package storage

import (
	"context"
	"strings"

	"github.com/Delpak/shahrah/services/control-plane/internal/modelid"
)

type CatalogModelPublication struct {
	Provider  string
	Model     string
	Published bool
}

// ListCatalogAvailabilityProviders returns providers with persisted tenant model selections.
func (s *Store) ListCatalogAvailabilityProviders(ctx context.Context) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT lower(a.provider)
		FROM tenant_provider_accounts a
		JOIN tenant_provider_account_models m ON m.account_id = a.account_id
		ORDER BY 1
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]string, 0)
	for rows.Next() {
		var provider string
		if err := rows.Scan(&provider); err != nil {
			return nil, err
		}
		out = append(out, strings.TrimSpace(provider))
	}
	return out, rows.Err()
}

func (s *Store) ReconcileLegacyDeploymentCatalogAvailability(ctx context.Context, catalog []CatalogModelPublication) (int64, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, COALESCE(tenant_id::text, ''), provider, model,
		       catalog_managed, catalog_available
		FROM model_deployments
	`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	type deployment struct {
		id, tenantID, provider, model string
		managed, available            bool
	}
	items := make([]deployment, 0)
	for rows.Next() {
		var item deployment
		if err := rows.Scan(&item.id, &item.tenantID, &item.provider, &item.model, &item.managed, &item.available); err != nil {
			return 0, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	var changed int64
	affected := make(map[string]struct{})
	for _, item := range items {
		known, published := false, false
		for _, candidate := range catalog {
			if strings.EqualFold(candidate.Provider, item.provider) &&
				modelid.MatchOfficialModel(item.provider, candidate.Model, item.model) {
				known, published = true, candidate.Published
				break
			}
		}
		if !known {
			continue
		}
		if _, err := s.pool.Exec(ctx, `
			UPDATE model_deployments
			SET catalog_managed = true,
			    catalog_available = $2,
			    catalog_checked_at = now(),
			    catalog_block_reason = CASE WHEN $2 THEN NULL ELSE 'catalog_unpublished' END
			WHERE id = $1::uuid
		`, item.id, published); err != nil {
			return changed, err
		}
		if !item.managed || item.available != published {
			changed++
			if item.tenantID != "" {
				affected[item.tenantID] = struct{}{}
			}
		}
	}
	for tenantID := range affected {
		s.invalidateDeploymentCache(ctx, tenantID)
		s.invalidateTenantRuntimeProfile(ctx, tenantID, "catalog.legacy_availability")
		s.warmDeploymentCacheAsync(tenantID)
	}
	return changed, nil
}

// ReconcileProviderCatalogAvailability applies the published catalog snapshot locally.
// Runtime reads never call model-discovery; they consume this durable availability bit.
func (s *Store) ReconcileProviderCatalogAvailability(ctx context.Context, provider string, publishedModels []string) (int64, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT m.account_model_id::text, a.tenant_id::text, m.model, m.catalog_available
		FROM tenant_provider_account_models m
		JOIN tenant_provider_accounts a ON a.account_id = m.account_id
		WHERE lower(a.provider) = lower($1)
	`, provider)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	type candidate struct {
		id, tenantID, model string
		current             bool
	}
	items := make([]candidate, 0)
	for rows.Next() {
		var item candidate
		if err := rows.Scan(&item.id, &item.tenantID, &item.model, &item.current); err != nil {
			return 0, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	var changed int64
	affectedTenants := make(map[string]struct{})
	for _, item := range items {
		available := false
		for _, published := range publishedModels {
			if modelid.MatchOfficialModel(provider, item.model, published) {
				available = true
				break
			}
		}
		tag, err := s.pool.Exec(ctx, `
			UPDATE tenant_provider_account_models
			SET catalog_available = $2,
			    catalog_checked_at = now(),
			    catalog_block_reason = CASE WHEN $2 THEN NULL ELSE 'catalog_unpublished' END,
			    updated_at = CASE WHEN catalog_available IS DISTINCT FROM $2 THEN now() ELSE updated_at END
			WHERE account_model_id = $1::uuid
		`, item.id, available)
		if err != nil {
			return changed, err
		}
		if item.current != available && tag.RowsAffected() > 0 {
			changed++
			affectedTenants[item.tenantID] = struct{}{}
		}
	}
	for tenantID := range affectedTenants {
		s.invalidateDeploymentCache(ctx, tenantID)
		s.invalidateTenantRuntimeProfile(ctx, tenantID, "catalog.availability")
		s.warmDeploymentCacheAsync(tenantID)
	}
	return changed, nil
}
