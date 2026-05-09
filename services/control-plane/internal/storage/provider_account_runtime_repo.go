package storage

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
)

// ResolveProviderAccountAPIKey resolves BYOK api_key for a tenant-scoped provider-account model FQN.
// modelFQN format: provider:account_alias:official_model_id.
func (s *Store) ResolveProviderAccountAPIKey(ctx context.Context, tenantID string, modelFQN string) (string, error) {
	parts := strings.Split(strings.TrimSpace(modelFQN), ":")
	if len(parts) != 3 {
		return "", nil
	}
	provider := strings.TrimSpace(parts[0])
	accountAlias := strings.TrimSpace(parts[1])
	model := strings.TrimSpace(parts[2])
	if provider == "" || accountAlias == "" || model == "" {
		return "", nil
	}

	var apiKey string
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(a.credential_data->>'api_key', '')
		FROM tenant_provider_accounts a
		JOIN tenant_provider_account_models m ON m.account_id = a.account_id
		WHERE a.tenant_id = $1
		  AND LOWER(a.provider) = LOWER($2)
		  AND a.name = $3
		  AND m.model = $4
		  AND m.enabled = true
		ORDER BY a.updated_at DESC
		LIMIT 1
	`, tenantID, provider, accountAlias, model).Scan(&apiKey)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", err
	}

	return strings.TrimSpace(apiKey), nil
}
