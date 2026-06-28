package storage

import (
	"context"
	"strings"

	"github.com/Delpak/shahrah/services/control-plane/internal/modelid"
)

// ResolveProviderAccountAPIKey resolves BYOK api_key for a tenant-scoped provider-account model FQN.
// modelFQN format: provider:account_alias:model_id.
func (s *Store) ResolveProviderAccountAPIKey(ctx context.Context, tenantID, modelFQN string) (string, error) {
	provider, accountAlias, model, ok := modelid.SplitModelFQN(modelFQN)
	if !ok {
		return "", nil
	}

	rows, err := s.pool.Query(ctx, `
		SELECT COALESCE(a.credential_data->>'api_key', ''), m.model
		FROM tenant_provider_accounts a
		JOIN tenant_provider_account_models m ON m.account_id = a.account_id
		WHERE a.tenant_id = $1
		  AND LOWER(a.provider) = LOWER($2)
		  AND a.name = $3
		  AND m.enabled = true
		ORDER BY a.updated_at DESC
	`, tenantID, provider, accountAlias)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			apiKey      string
			storedModel string
		)
		if err := rows.Scan(&apiKey, &storedModel); err != nil {
			return "", err
		}

		if modelid.MatchOfficialModel(provider, storedModel, model) {
			return strings.TrimSpace(apiKey), nil
		}
	}

	if err := rows.Err(); err != nil {
		return "", err
	}

	return "", nil
}
