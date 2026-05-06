package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

var (
	ErrInvalidAccessPrincipal            = errors.New("invalid access principal")
	ErrModelNotAllowedForProviderAccount = errors.New("model is not allowed for provider account")
)

// CreateTenantProviderAccount creates a provider account for the tenant.
func (s *Store) CreateTenantProviderAccount(ctx context.Context, tenantID string, req platform.CreateTenantProviderAccountRequest) (*platform.TenantProviderAccount, error) {
	providerCfg, err := marshalJSONObject(req.ProviderConfig)
	if err != nil {
		return nil, err
	}
	credentialData, hasCredential, err := marshalCredentialData(req.Credentials)
	if err != nil {
		return nil, err
	}

	var account platform.TenantProviderAccount
	var providerCfgRaw []byte
	err = s.pool.QueryRow(ctx, `
		INSERT INTO tenant_provider_accounts (
			tenant_id, provider, name, auth_mode, provider_config, credential_data, has_credential, allow_virtual_keys_only
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING account_id::text, tenant_id::text, provider, name, auth_mode, provider_config, has_credential, allow_virtual_keys_only, created_at, updated_at
	`, tenantID, req.Provider, req.Name, string(req.AuthMode), providerCfg, credentialData, hasCredential, req.AllowVirtualKeysOnly).Scan(
		&account.ID,
		&account.TenantID,
		&account.Provider,
		&account.Name,
		&account.AuthMode,
		&providerCfgRaw,
		&account.HasCredential,
		&account.AllowsVirtualKeysOnly,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	account.ProviderConfig = parseJSONObject(providerCfgRaw)
	if s.logger != nil {
		s.logger.Info("provider_account.created",
			zap.String("tenant_id", tenantID),
			zap.String("account_id", account.ID),
			zap.String("provider", account.Provider),
			zap.String("name", account.Name),
		)
	}
	return &account, nil
}

// ListTenantProviderAccounts lists provider accounts scoped to one tenant.
func (s *Store) ListTenantProviderAccounts(ctx context.Context, tenantID string) ([]platform.TenantProviderAccount, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT account_id::text, tenant_id::text, provider, name, auth_mode, provider_config,
		       has_credential, allow_virtual_keys_only, created_at, updated_at
		FROM tenant_provider_accounts
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]platform.TenantProviderAccount, 0)
	for rows.Next() {
		var account platform.TenantProviderAccount
		var providerCfgRaw []byte
		if err := rows.Scan(
			&account.ID,
			&account.TenantID,
			&account.Provider,
			&account.Name,
			&account.AuthMode,
			&providerCfgRaw,
			&account.HasCredential,
			&account.AllowsVirtualKeysOnly,
			&account.CreatedAt,
			&account.UpdatedAt,
		); err != nil {
			return nil, err
		}
		account.ProviderConfig = parseJSONObject(providerCfgRaw)
		items = append(items, account)
	}
	return items, rows.Err()
}

// GetTenantProviderAccountDetails returns account, selected models, and access control.
func (s *Store) GetTenantProviderAccountDetails(ctx context.Context, tenantID string, accountID string) (*platform.TenantProviderAccountDetailsResponse, error) {
	account, err := s.getTenantProviderAccount(ctx, tenantID, accountID)
	if err != nil {
		return nil, err
	}

	models, err := s.listTenantProviderAccountModels(ctx, accountID)
	if err != nil {
		return nil, err
	}
	access, err := s.listTenantProviderAccountAccess(ctx, accountID)
	if err != nil {
		return nil, err
	}

	return &platform.TenantProviderAccountDetailsResponse{
		Account: *account,
		Models:  models,
		Access:  access,
	}, nil
}

// UpsertTenantProviderAccountModels replaces model selection for one provider account.
func (s *Store) UpsertTenantProviderAccountModels(ctx context.Context, tenantID string, accountID string, models []platform.TenantProviderAccountModel) ([]platform.TenantProviderAccountModel, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if _, err := ensureTenantProviderAccountExists(ctx, tx, tenantID, accountID); err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM tenant_provider_account_models WHERE account_id = $1`, accountID); err != nil {
		return nil, err
	}

	seenModel := make(map[string]struct{}, len(models))
	modelsChanged := 0
	for _, item := range models {
		modelName := strings.TrimSpace(item.Model)
		if modelName == "" {
			return nil, fmt.Errorf("model is required")
		}
		if _, ok := seenModel[modelName]; ok {
			continue
		}
		seenModel[modelName] = struct{}{}

		settings, err := marshalJSONObject(item.Settings)
		if err != nil {
			return nil, err
		}
		privatePricing, err := marshalModelPrice(item.PrivatePricing)
		if err != nil {
			return nil, err
		}
		pricingMode := item.PricingMode
		if pricingMode == "" {
			pricingMode = platform.PricingModePublic
		}
		types := modelTypesToStrings(item.Types)

		if _, err := tx.Exec(ctx, `
			INSERT INTO tenant_provider_account_models (
				account_id, model, enabled, types, pricing_mode, private_pricing, settings
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, accountID, modelName, item.Enabled, types, string(pricingMode), privatePricing, settings); err != nil {
			return nil, err
		}
		modelsChanged++
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	items, err := s.listTenantProviderAccountModels(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if s.logger != nil {
		s.logger.Info("provider_account.models_replaced",
			zap.String("tenant_id", tenantID),
			zap.String("account_id", accountID),
			zap.Int("models_count", modelsChanged),
		)
	}
	return items, nil
}

// UpdateTenantProviderAccountAccess replaces manager/user access assignments for one account.
func (s *Store) UpdateTenantProviderAccountAccess(ctx context.Context, tenantID string, accountID string, access platform.TenantProviderAccountAccessControl) (*platform.TenantProviderAccountAccessControl, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if _, err := ensureTenantProviderAccountExists(ctx, tx, tenantID, accountID); err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM tenant_provider_account_access WHERE account_id = $1`, accountID); err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})
	if err := insertAccessRows(ctx, tx, tenantID, accountID, platform.AccessRoleManager, access.Managers, seen); err != nil {
		return nil, err
	}
	if err := insertAccessRows(ctx, tx, tenantID, accountID, platform.AccessRoleUser, access.Users, seen); err != nil {
		return nil, err
	}
	if err := insertAccessRules(ctx, tx, tenantID, accountID, access.Rules, seen); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	result, err := s.listTenantProviderAccountAccess(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if s.logger != nil {
		s.logger.Info("provider_account.access_replaced",
			zap.String("tenant_id", tenantID),
			zap.String("account_id", accountID),
			zap.Int("managers_count", len(result.Managers)),
			zap.Int("users_count", len(result.Users)),
		)
	}
	return &result, nil
}

func insertAccessRules(ctx context.Context, tx pgx.Tx, tenantID string, accountID string, rules []platform.TenantProviderAccountAccessRule, seen map[string]struct{}) error {
	for _, rule := range rules {
		role := rule.Role
		if role != platform.AccessRoleManager && role != platform.AccessRoleUser {
			continue
		}
		scope := rule.AccessScope
		if scope == "" {
			scope = platform.AccessScopeAccount
		}
		model := strings.TrimSpace(rule.Model)
		if scope == platform.AccessScopeModel && model == "" {
			continue
		}
		if scope == platform.AccessScopeAccount {
			model = ""
		}

		normalized, err := normalizeAccessPrincipal(ctx, tx, tenantID, role, rule.Principal)
		if err != nil {
			return err
		}
		key := strings.Join([]string{string(role), string(scope), model, string(normalized.Type), normalized.PrincipalID}, "|")
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		if _, err := tx.Exec(ctx, `
			INSERT INTO tenant_provider_account_access (
				account_id, role, access_scope, model, principal_type, principal_id, display_name
			)
			VALUES ($1, $2, $3, NULLIF($4, ''), $5, NULLIF($6, ''), NULLIF($7, ''))
		`, accountID, string(role), string(scope), model, string(normalized.Type), normalized.PrincipalID, normalized.DisplayName); err != nil {
			return err
		}
	}
	return nil
}

// TestTenantProviderAccountCredentials validates whether the account can proceed.
func (s *Store) TestTenantProviderAccountCredentials(ctx context.Context, tenantID string, accountID string, req platform.TestTenantProviderAccountCredentialsRequest) (*platform.TestTenantProviderAccountCredentialsResponse, error) {
	var hasCredential bool
	var allowVirtualOnly bool

	err := s.pool.QueryRow(ctx, `
		SELECT has_credential, allow_virtual_keys_only
		FROM tenant_provider_accounts
		WHERE tenant_id = $1 AND account_id = $2
	`, tenantID, accountID).Scan(&hasCredential, &allowVirtualOnly)
	if err != nil {
		return nil, err
	}

	if hasCredential {
		return &platform.TestTenantProviderAccountCredentialsResponse{
			Success: true,
			Message: "credential is present and can be used",
		}, nil
	}
	if allowVirtualOnly {
		return &platform.TestTenantProviderAccountCredentialsResponse{
			Success: true,
			Message: "virtual-key-only mode enabled",
		}, nil
	}

	return &platform.TestTenantProviderAccountCredentialsResponse{
		Success: false,
		Message: "credential is missing; provide BYOK or enable virtual-key-only mode",
	}, nil
}

func (s *Store) getTenantProviderAccount(ctx context.Context, tenantID string, accountID string) (*platform.TenantProviderAccount, error) {
	var account platform.TenantProviderAccount
	var providerCfgRaw []byte
	err := s.pool.QueryRow(ctx, `
		SELECT account_id::text, tenant_id::text, provider, name, auth_mode, provider_config,
		       has_credential, allow_virtual_keys_only, created_at, updated_at
		FROM tenant_provider_accounts
		WHERE tenant_id = $1 AND account_id = $2
	`, tenantID, accountID).Scan(
		&account.ID,
		&account.TenantID,
		&account.Provider,
		&account.Name,
		&account.AuthMode,
		&providerCfgRaw,
		&account.HasCredential,
		&account.AllowsVirtualKeysOnly,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	account.ProviderConfig = parseJSONObject(providerCfgRaw)
	return &account, nil
}

func (s *Store) listTenantProviderAccountModels(ctx context.Context, accountID string) ([]platform.TenantProviderAccountModel, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT model, enabled, COALESCE(types, '{}'), pricing_mode, private_pricing::text, COALESCE(settings, '{}'::jsonb)
		FROM tenant_provider_account_models
		WHERE account_id = $1
		ORDER BY created_at ASC, model ASC
	`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]platform.TenantProviderAccountModel, 0)
	for rows.Next() {
		var model platform.TenantProviderAccountModel
		var types []string
		var privatePricingText *string
		var settingsRaw []byte
		if err := rows.Scan(
			&model.Model,
			&model.Enabled,
			&types,
			&model.PricingMode,
			&privatePricingText,
			&settingsRaw,
		); err != nil {
			return nil, err
		}
		model.Types = stringsToModelTypes(types)
		model.Settings = parseJSONObject(settingsRaw)
		if privatePricingText != nil && strings.TrimSpace(*privatePricingText) != "" && strings.TrimSpace(*privatePricingText) != "null" {
			var parsed platform.ModelPrice
			if err := json.Unmarshal([]byte(*privatePricingText), &parsed); err == nil {
				model.PrivatePricing = &parsed
			}
		}
		items = append(items, model)
	}

	return items, rows.Err()
}

func (s *Store) listTenantProviderAccountAccess(ctx context.Context, accountID string) (platform.TenantProviderAccountAccessControl, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT role, access_scope, COALESCE(model, ''), principal_type, COALESCE(principal_id, ''), COALESCE(display_name, '')
		FROM tenant_provider_account_access
		WHERE account_id = $1
		ORDER BY created_at ASC
	`, accountID)
	if err != nil {
		return platform.TenantProviderAccountAccessControl{}, err
	}
	defer rows.Close()

	access := platform.TenantProviderAccountAccessControl{
		Managers: make([]platform.AccessPrincipal, 0),
		Users:    make([]platform.AccessPrincipal, 0),
		Rules:    make([]platform.TenantProviderAccountAccessRule, 0),
	}
	for rows.Next() {
		var role platform.AccessRole
		var scope platform.AccessScope
		var model string
		var principal platform.AccessPrincipal
		if err := rows.Scan(&role, &scope, &model, &principal.Type, &principal.PrincipalID, &principal.DisplayName); err != nil {
			return platform.TenantProviderAccountAccessControl{}, err
		}
		if principal.Type == platform.PrincipalTypeEveryone {
			principal.PrincipalID = ""
		}
		rule := platform.TenantProviderAccountAccessRule{
			Role:        role,
			Principal:   principal,
			AccessScope: scope,
			Model:       strings.TrimSpace(model),
		}
		if rule.AccessScope == "" {
			rule.AccessScope = platform.AccessScopeAccount
		}
		access.Rules = append(access.Rules, rule)
		switch role {
		case platform.AccessRoleManager:
			access.Managers = append(access.Managers, principal)
		case platform.AccessRoleUser:
			access.Users = append(access.Users, principal)
		}
	}

	return access, rows.Err()
}

func ensureTenantProviderAccountExists(ctx context.Context, tx pgx.Tx, tenantID string, accountID string) (bool, error) {
	var exists bool
	if err := tx.QueryRow(ctx, `
		SELECT true
		FROM tenant_provider_accounts
		WHERE tenant_id = $1 AND account_id = $2
	`, tenantID, accountID).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func insertAccessRows(ctx context.Context, tx pgx.Tx, tenantID string, accountID string, role platform.AccessRole, principals []platform.AccessPrincipal, seen map[string]struct{}) error {
	for _, principal := range principals {
		normalized, err := normalizeAccessPrincipal(ctx, tx, tenantID, role, principal)
		if err != nil {
			return err
		}
		key := string(role) + "|" + string(normalized.Type) + "|" + normalized.PrincipalID
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		if _, err := tx.Exec(ctx, `
			INSERT INTO tenant_provider_account_access (
				account_id, role, access_scope, model, principal_type, principal_id, display_name
			)
			VALUES ($1, $2, 'account', NULL, $3, NULLIF($4, ''), NULLIF($5, ''))
		`, accountID, string(role), string(normalized.Type), normalized.PrincipalID, normalized.DisplayName); err != nil {
			return err
		}
	}
	return nil
}

// ListTenantProviderAccountEffectiveModels resolves model access visibility for debugging/admin observability.
func (s *Store) ListTenantProviderAccountEffectiveModels(ctx context.Context, tenantID string, accountID string, keyType string, userID string, teamID string) ([]platform.EffectiveModelView, error) {
	_ = keyType // reserved for PAT/VAT-specific semantics in later phases
	// Ensure account exists and belongs to tenant.
	if _, err := s.getTenantProviderAccount(ctx, tenantID, accountID); err != nil {
		return nil, err
	}

	rows, err := s.pool.Query(ctx, `
		SELECT m.model,
		       m.enabled,
		       COALESCE(m.types, '{}'),
		       COALESCE(m.pricing_mode, 'public'),
		       CASE
		         WHEN COALESCE(m.pricing_mode, 'public') = 'private' THEN m.private_pricing IS NOT NULL
		         ELSE true
		       END AS pricing_resolved,
		       EXISTS (
		         SELECT 1
		         FROM tenant_provider_account_access ac
		         WHERE ac.account_id = m.account_id
		           AND ac.access_scope = 'account'
		           AND (
		             ac.principal_type = 'everyone'
		             OR (ac.principal_type = 'team' AND ac.principal_id = NULLIF($4, ''))
		             OR (ac.principal_type = 'user' AND ac.principal_id = NULLIF($3, ''))
		           )
		       ) AS allowed_by_access
		FROM tenant_provider_account_models m
		WHERE m.account_id = $2
		ORDER BY m.model ASC
	`, tenantID, accountID, strings.TrimSpace(userID), strings.TrimSpace(teamID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]platform.EffectiveModelView, 0)
	for rows.Next() {
		var item platform.EffectiveModelView
		var types []string
		var pricingMode string
		if err := rows.Scan(&item.Model, &item.Enabled, &types, &pricingMode, &item.PricingResolved, &item.Allowed); err != nil {
			return nil, err
		}
		item.PricingMode = platform.PricingMode(pricingMode)
		item.EffectiveCapabilities = stringsToModelTypes(types)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func normalizeAccessPrincipal(ctx context.Context, tx pgx.Tx, tenantID string, role platform.AccessRole, principal platform.AccessPrincipal) (platform.AccessPrincipal, error) {
	principal.Type = platform.PrincipalType(strings.TrimSpace(string(principal.Type)))
	principal.PrincipalID = strings.TrimSpace(principal.PrincipalID)
	principal.DisplayName = strings.TrimSpace(principal.DisplayName)

	switch principal.Type {
	case platform.PrincipalTypeUser, platform.PrincipalTypeTeam:
		if principal.PrincipalID == "" {
			return platform.AccessPrincipal{}, fmt.Errorf("principal_id is required for user/team entries")
		}
	case platform.PrincipalTypeEveryone:
		if role == platform.AccessRoleManager {
			return platform.AccessPrincipal{}, fmt.Errorf("manager role only supports user/team principals")
		}
		principal.PrincipalID = ""
	default:
		return platform.AccessPrincipal{}, fmt.Errorf("unsupported principal type: %s", principal.Type)
	}
	if principal.Type == platform.PrincipalTypeUser || principal.Type == platform.PrincipalTypeTeam {
		ok, err := principalExistsForTenant(ctx, tx, tenantID, principal.Type, principal.PrincipalID)
		if err != nil {
			return platform.AccessPrincipal{}, err
		}
		if !ok {
			return platform.AccessPrincipal{}, fmt.Errorf("%w: %s principal %q is not found in tenant", ErrInvalidAccessPrincipal, principal.Type, principal.PrincipalID)
		}
	}
	return principal, nil
}

func principalExistsForTenant(ctx context.Context, tx pgx.Tx, tenantID string, principalType platform.PrincipalType, principalID string) (bool, error) {
	switch principalType {
	case platform.PrincipalTypeTeam:
		var exists bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1
				FROM teams
				WHERE tenant_id = $1::uuid AND team_id = $2::uuid
			)
		`, tenantID, principalID).Scan(&exists); err != nil {
			return false, err
		}
		return exists, nil
	case platform.PrincipalTypeUser:
		var exists bool
		if err := tx.QueryRow(ctx, `
			WITH tenant_slug_cte AS (
				SELECT slug
				FROM tenants
				WHERE tenant_id = $1::uuid
				LIMIT 1
			)
			SELECT EXISTS (
				SELECT 1
				FROM users u
				WHERE u.tenant_id = $1::uuid
				  AND u.user_id::text = $2
			) OR EXISTS (
				SELECT 1
				FROM "tenant_admin_users" tau
				JOIN tenant_slug_cte ts ON tau."tenantSlug" = ts.slug
				WHERE tau."id" = $2
			)
		`, tenantID, principalID).Scan(&exists); err != nil {
			return false, err
		}
		return exists, nil
	default:
		return false, nil
	}
}

func marshalCredentialData(input *platform.ProviderAccountCredentials) ([]byte, bool, error) {
	if input == nil {
		return nil, false, nil
	}

	payload := make(map[string]interface{})
	if apiKey := strings.TrimSpace(input.APIKey); apiKey != "" {
		payload["api_key"] = apiKey
	}
	if secretRef := strings.TrimSpace(input.SecretRef); secretRef != "" {
		payload["secret_ref"] = secretRef
	}
	if len(input.Raw) > 0 {
		payload["raw"] = input.Raw
	}
	if input.SkipRequired {
		payload["skip_required"] = true
	}
	if len(payload) == 0 {
		return nil, false, nil
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, false, err
	}
	_, hasAPIKey := payload["api_key"]
	_, hasSecretRef := payload["secret_ref"]
	_, hasRaw := payload["raw"]
	return raw, hasAPIKey || hasSecretRef || hasRaw, nil
}

func marshalJSONObject(input map[string]interface{}) ([]byte, error) {
	if len(input) == 0 {
		return []byte("{}"), nil
	}
	return json.Marshal(input)
}

func parseJSONObject(raw []byte) map[string]interface{} {
	if len(raw) == 0 {
		return map[string]interface{}{}
	}
	var out map[string]interface{}
	if err := json.Unmarshal(raw, &out); err != nil || out == nil {
		return map[string]interface{}{}
	}
	return out
}

func marshalModelPrice(price *platform.ModelPrice) ([]byte, error) {
	if price == nil {
		return nil, nil
	}
	return json.Marshal(price)
}

func modelTypesToStrings(types []platform.ModelType) []string {
	items := make([]string, 0, len(types))
	for _, item := range types {
		value := strings.TrimSpace(string(item))
		if value == "" {
			continue
		}
		items = append(items, value)
	}
	return items
}

func stringsToModelTypes(values []string) []platform.ModelType {
	items := make([]platform.ModelType, 0, len(values))
	for _, item := range values {
		value := strings.TrimSpace(item)
		if value == "" {
			continue
		}
		items = append(items, platform.ModelType(value))
	}
	return items
}
