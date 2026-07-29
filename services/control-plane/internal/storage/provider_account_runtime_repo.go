package storage

import (
	"context"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/modelid"
)

type providerAccountRuntimeRecord struct {
	accountID      string
	authMode       string
	providerConfig map[string]interface{}
}

// ResolveProviderAccountAPIKey preserves the legacy resolver contract while
// reading the encrypted tenant credential version first and plaintext only for
// not-yet-migrated accounts.
func (s *Store) ResolveProviderAccountAPIKey(
	ctx context.Context,
	tenantID string,
	modelFQN string,
) (string, error) {
	record, ok, err := s.resolveProviderAccountRuntimeRecord(
		ctx,
		tenantID,
		modelFQN,
	)
	if err != nil || !ok ||
		!strings.EqualFold(record.authMode, string(platform.ProviderAuthModeAPIKey)) {
		return "", err
	}
	return s.resolveTenantProviderAccountAPIKey(ctx, tenantID, record.accountID)
}

func (s *Store) ResolveProviderAccountAuth(
	ctx context.Context,
	tenantID string,
	modelFQN string,
) (*platform.ResolvedProviderAuth, error) {
	record, ok, err := s.resolveProviderAccountRuntimeRecord(
		ctx,
		tenantID,
		modelFQN,
	)
	if err != nil || !ok {
		return nil, err
	}

	switch platform.ProviderAuthMode(strings.TrimSpace(record.authMode)) {
	case platform.ProviderAuthModeAPIKey:
		apiKey, resolveErr := s.resolveTenantProviderAccountAPIKey(
			ctx,
			tenantID,
			record.accountID,
		)
		if resolveErr != nil || apiKey == "" {
			return nil, resolveErr
		}
		return &platform.ResolvedProviderAuth{
			Mode:   string(platform.ProviderAuthModeAPIKey),
			APIKey: apiKey,
		}, nil
	case platform.ProviderAuthModeAnthropicWIF:
		return &platform.ResolvedProviderAuth{
			Mode: string(platform.ProviderAuthModeAnthropicWIF),
			AnthropicWIF: &platform.ResolvedAnthropicWIFAuth{
				FederationRuleID: tenantProviderStringConfig(
					record.providerConfig,
					"federation_rule_id",
				),
				OrganizationID: tenantProviderStringConfig(
					record.providerConfig,
					"organization_id",
				),
				ServiceAccountID: tenantProviderStringConfig(
					record.providerConfig,
					"service_account_id",
				),
				WorkspaceID: tenantProviderStringConfig(record.providerConfig, "workspace_id"),
				TokenSourceID: tenantProviderStringConfig(
					record.providerConfig,
					"token_source_id",
				),
			},
		}, nil
	default:
		return nil, nil
	}
}

func (s *Store) resolveProviderAccountRuntimeRecord(
	ctx context.Context,
	tenantID string,
	modelFQN string,
) (providerAccountRuntimeRecord, bool, error) {
	provider, accountAlias, model, ok := modelid.SplitModelFQN(modelFQN)
	if !ok {
		return providerAccountRuntimeRecord{}, false, nil
	}

	rows, err := s.pool.Query(ctx, `
		SELECT a.account_id::text, a.auth_mode,
		       COALESCE(a.provider_config, '{}'::jsonb), m.model
		FROM tenant_provider_accounts a
		JOIN tenant_provider_account_models m ON m.account_id = a.account_id
		WHERE a.tenant_id = $1
		  AND LOWER(a.provider) = LOWER($2)
		  AND a.name = $3
		  AND a.has_credential = true
		  AND a.use_platform_managed_credentials = false
		  AND m.enabled = true
		ORDER BY a.updated_at DESC
	`, tenantID, provider, accountAlias)
	if err != nil {
		return providerAccountRuntimeRecord{}, false, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			record    providerAccountRuntimeRecord
			configRaw []byte
			stored    string
		)
		if err := rows.Scan(
			&record.accountID,
			&record.authMode,
			&configRaw,
			&stored,
		); err != nil {
			return providerAccountRuntimeRecord{}, false, err
		}
		if modelid.MatchOfficialModel(provider, stored, model) {
			record.providerConfig = parseJSONObject(configRaw)
			return record, true, nil
		}
	}
	if err := rows.Err(); err != nil {
		return providerAccountRuntimeRecord{}, false, err
	}
	return providerAccountRuntimeRecord{}, false, nil
}

func tenantProviderStringConfig(config map[string]interface{}, key string) string {
	value, _ := config[key].(string)
	return strings.TrimSpace(value)
}
