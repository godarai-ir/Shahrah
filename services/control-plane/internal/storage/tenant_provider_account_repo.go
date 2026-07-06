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
	ErrProviderAccountDependencyConflict = errors.New("provider account has active dependencies")
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

	var (
		account        platform.TenantProviderAccount
		providerCfgRaw []byte
		credentialRaw  []byte
	)

	err = s.pool.QueryRow(ctx, `
		INSERT INTO tenant_provider_accounts (
			tenant_id, provider, name, auth_mode, provider_config, credential_data, has_credential, allow_virtual_keys_only
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING account_id::text, tenant_id::text, provider, name, auth_mode, provider_config, COALESCE(credential_data, '{}'::jsonb), has_credential, allow_virtual_keys_only, created_at, updated_at
	`, tenantID, req.Provider, req.Name, string(req.AuthMode), providerCfg, credentialData, hasCredential, req.AllowVirtualKeysOnly).Scan(
		&account.ID,
		&account.TenantID,
		&account.Provider,
		&account.Name,
		&account.AuthMode,
		&providerCfgRaw,
		&credentialRaw,
		&account.HasCredential,
		&account.AllowsVirtualKeysOnly,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	account.ProviderConfig = parseJSONObject(providerCfgRaw)

	account.CredentialHint = credentialHintFromMap(parseJSONObject(credentialRaw))
	if s.logger != nil {
		s.logger.Info("provider_account.created",
			zap.String("tenant_id", tenantID),
			zap.String("account_id", account.ID),
			zap.String("provider", account.Provider),
			zap.String("name", account.Name),
		)
	}

	s.invalidateDeploymentCache(ctx, tenantID)
	s.invalidateTenantRuntimeProfile(ctx, tenantID, "provider_account.created")
	s.warmDeploymentCacheAsync(tenantID)

	return &account, nil
}

// ListTenantProviderAccounts lists provider accounts scoped to one tenant.
func (s *Store) ListTenantProviderAccounts(ctx context.Context, tenantID string) ([]platform.TenantProviderAccount, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT account_id::text, tenant_id::text, provider, name, auth_mode, provider_config, COALESCE(credential_data, '{}'::jsonb),
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
		var (
			account        platform.TenantProviderAccount
			providerCfgRaw []byte
			credentialRaw  []byte
		)
		if err := rows.Scan(
			&account.ID,
			&account.TenantID,
			&account.Provider,
			&account.Name,
			&account.AuthMode,
			&providerCfgRaw,
			&credentialRaw,
			&account.HasCredential,
			&account.AllowsVirtualKeysOnly,
			&account.CreatedAt,
			&account.UpdatedAt,
		); err != nil {
			return nil, err
		}

		account.ProviderConfig = parseJSONObject(providerCfgRaw)
		account.CredentialHint = credentialHintFromMap(parseJSONObject(credentialRaw))
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

// UpdateTenantProviderAccountCredentials updates auth mode, provider config, and credentials for one account.
func (s *Store) UpdateTenantProviderAccountCredentials(ctx context.Context, tenantID string, accountID string, req platform.UpdateTenantProviderAccountCredentialsRequest) (*platform.TenantProviderAccount, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	var (
		currentAuthMode      string
		currentProviderCfg   []byte
		currentCredentialRaw []byte
		currentHasCredential bool
		currentVirtualOnly   bool
	)

	queryErr := tx.QueryRow(ctx, `
		SELECT auth_mode, provider_config, COALESCE(credential_data, '{}'::jsonb), has_credential, allow_virtual_keys_only
		FROM tenant_provider_accounts
		WHERE tenant_id = $1 AND account_id = $2
	`, tenantID, accountID).Scan(&currentAuthMode, &currentProviderCfg, &currentCredentialRaw, &currentHasCredential, &currentVirtualOnly)
	if queryErr != nil {
		return nil, queryErr
	}

	nextAuthMode := strings.TrimSpace(currentAuthMode)
	if req.AuthMode != "" {
		nextAuthMode = strings.TrimSpace(string(req.AuthMode))
	}

	nextProviderCfg := parseJSONObject(currentProviderCfg)
	if req.ProviderConfig != nil {
		nextProviderCfg = req.ProviderConfig
	}

	nextProviderCfgRaw, err := marshalJSONObject(nextProviderCfg)
	if err != nil {
		return nil, err
	}

	nextCredentialMap := parseJSONObject(currentCredentialRaw)

	if req.Credentials != nil {
		if req.Credentials.APIKey != "" {
			nextCredentialMap["api_key"] = strings.TrimSpace(req.Credentials.APIKey)
		}

		if req.Credentials.SecretRef != "" {
			nextCredentialMap["secret_ref"] = strings.TrimSpace(req.Credentials.SecretRef)
		}

		if len(req.Credentials.Raw) > 0 {
			nextCredentialMap["raw"] = req.Credentials.Raw
		}

		if req.Credentials.SkipRequired {
			nextCredentialMap["skip_required"] = true
		}
	}

	nextCredentialRaw, err := marshalJSONObject(nextCredentialMap)
	if err != nil {
		return nil, err
	}

	nextHasCredential := false
	if v, ok := nextCredentialMap["api_key"].(string); ok && strings.TrimSpace(v) != "" {
		nextHasCredential = true
	}

	if v, ok := nextCredentialMap["secret_ref"].(string); ok && strings.TrimSpace(v) != "" {
		nextHasCredential = true
	}

	if _, ok := nextCredentialMap["raw"]; ok {
		nextHasCredential = true
	}

	if req.Credentials == nil {
		nextHasCredential = currentHasCredential
		nextCredentialRaw = currentCredentialRaw
	}

	nextVirtualOnly := currentVirtualOnly
	if req.AllowVirtualKeysOnly != nil {
		nextVirtualOnly = *req.AllowVirtualKeysOnly
	}

	_, execErr := tx.Exec(ctx, `
		UPDATE tenant_provider_accounts
		SET auth_mode = $3,
		    provider_config = $4,
		    credential_data = $5,
		    has_credential = $6,
		    allow_virtual_keys_only = $7,
		    updated_at = NOW()
		WHERE tenant_id = $1 AND account_id = $2
	`, tenantID, accountID, nextAuthMode, nextProviderCfgRaw, nextCredentialRaw, nextHasCredential, nextVirtualOnly)
	if execErr != nil {
		return nil, execErr
	}

	commitErr := tx.Commit(ctx)
	if commitErr != nil {
		return nil, commitErr
	}

	s.invalidateDeploymentCache(ctx, tenantID)
	s.invalidateTenantRuntimeProfile(ctx, tenantID, "provider_account.credentials_updated")
	s.warmDeploymentCacheAsync(tenantID)

	return s.getTenantProviderAccount(ctx, tenantID, accountID)
}

// UpsertTenantProviderAccountModels replaces model selection for one provider account.
func (s *Store) UpsertTenantProviderAccountModels(ctx context.Context, tenantID string, accountID string, models []platform.TenantProviderAccountModel) ([]platform.TenantProviderAccountModel, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	_, existsErr := ensureTenantProviderAccountExists(ctx, tx, tenantID, accountID)
	if existsErr != nil {
		return nil, existsErr
	}

	_, deleteErr := tx.Exec(ctx, `DELETE FROM tenant_provider_account_models WHERE account_id = $1`, accountID)
	if deleteErr != nil {
		return nil, deleteErr
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

		settings, settingsErr := marshalJSONObject(item.Settings)
		if settingsErr != nil {
			return nil, settingsErr
		}

		privatePricing, privatePricingErr := marshalModelPrice(item.PrivatePricing)
		if privatePricingErr != nil {
			return nil, privatePricingErr
		}

		publicPricing, publicPricingErr := marshalModelPrice(item.PublicPricing)
		if publicPricingErr != nil {
			return nil, publicPricingErr
		}

		pricingMode := item.PricingMode
		if pricingMode == "" {
			pricingMode = platform.PricingModePublic
		}

		types := modelTypesToStrings(item.Types)
		inputModalities := normalizeStringList(item.InputModalities)
		outputModalities := normalizeStringList(item.OutputModalities)

		_, insertErr := tx.Exec(ctx, `
			INSERT INTO tenant_provider_account_models (
				account_id, model, enabled, types, input_modalities, output_modalities, pricing_mode, public_pricing, private_pricing, settings
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, accountID, modelName, item.Enabled, types, inputModalities, outputModalities, string(pricingMode), publicPricing, privatePricing, settings)
		if insertErr != nil {
			return nil, insertErr
		}

		modelsChanged++
	}

	commitErr := tx.Commit(ctx)
	if commitErr != nil {
		return nil, commitErr
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

	s.invalidateDeploymentCache(ctx, tenantID)
	s.invalidateTenantRuntimeProfile(ctx, tenantID, "provider_account.models_updated")
	s.warmDeploymentCacheAsync(tenantID)

	return items, nil
}

// UpdateTenantProviderAccountAccess replaces manager/user access assignments for one account.
func (s *Store) UpdateTenantProviderAccountAccess(ctx context.Context, tenantID string, accountID string, access platform.TenantProviderAccountAccessControl) (*platform.TenantProviderAccountAccessControl, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	_, existsErr := ensureTenantProviderAccountExists(ctx, tx, tenantID, accountID)
	if existsErr != nil {
		return nil, existsErr
	}

	_, deleteErr := tx.Exec(ctx, `DELETE FROM tenant_provider_account_access WHERE account_id = $1`, accountID)
	if deleteErr != nil {
		return nil, deleteErr
	}

	seen := make(map[string]struct{})

	managerErr := insertAccessRows(ctx, tx, tenantID, accountID, platform.AccessRoleManager, access.Managers, seen)
	if managerErr != nil {
		return nil, managerErr
	}

	userErr := insertAccessRows(ctx, tx, tenantID, accountID, platform.AccessRoleUser, access.Users, seen)
	if userErr != nil {
		return nil, userErr
	}

	rulesErr := insertAccessRules(ctx, tx, tenantID, accountID, access.Rules, seen)
	if rulesErr != nil {
		return nil, rulesErr
	}

	commitErr := tx.Commit(ctx)
	if commitErr != nil {
		return nil, commitErr
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

	s.invalidateDeploymentCache(ctx, tenantID)
	s.invalidateTenantRuntimeProfile(ctx, tenantID, "provider_account.access_updated")
	s.warmDeploymentCacheAsync(tenantID)

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

// TestTenantProviderAccountCredentials validates whether the account can reach its upstream provider.
func (s *Store) TestTenantProviderAccountCredentials(ctx context.Context, tenantID string, accountID string, req platform.TestTenantProviderAccountCredentialsRequest) (*platform.TestTenantProviderAccountCredentialsResponse, error) {
	var (
		provider         string
		authMode         string
		providerCfgRaw   []byte
		credentialRaw    []byte
		hasCredential    bool
		allowVirtualOnly bool
	)

	err := s.pool.QueryRow(ctx, `
		SELECT provider, auth_mode, provider_config, COALESCE(credential_data, '{}'::jsonb), has_credential, allow_virtual_keys_only
		FROM tenant_provider_accounts
		WHERE tenant_id = $1 AND account_id = $2
	`, tenantID, accountID).Scan(&provider, &authMode, &providerCfgRaw, &credentialRaw, &hasCredential, &allowVirtualOnly)
	if err != nil {
		return nil, err
	}

	if allowVirtualOnly {
		return &platform.TestTenantProviderAccountCredentialsResponse{
			Success: true,
			Message: "virtual-key-only mode is enabled; no provider credential is required",
		}, nil
	}

	if !hasCredential {
		return &platform.TestTenantProviderAccountCredentialsResponse{
			Success: false,
			Message: "credential is missing; provide an API key or enable virtual-key-only mode",
		}, nil
	}

	if platform.ProviderAuthMode(strings.TrimSpace(authMode)) != platform.ProviderAuthModeAPIKey {
		return &platform.TestTenantProviderAccountCredentialsResponse{
			Success: false,
			Message: "connection validation is only implemented for API key provider accounts",
		}, nil
	}

	credentials := parseJSONObject(credentialRaw)

	apiKey := ""
	if value, ok := credentials["api_key"].(string); ok {
		apiKey = value
	}

	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" || isMaskedCredentialValue(apiKey) {
		return &platform.TestTenantProviderAccountCredentialsResponse{
			Success: false,
			Message: "API key is missing or masked; enter a new provider API key before testing",
		}, nil
	}

	result := validateProviderAPIKey(ctx, provider, apiKey, parseJSONObject(providerCfgRaw), nil)

	return &platform.TestTenantProviderAccountCredentialsResponse{
		Success: result.Success,
		Message: result.Message,
	}, nil
}

func (s *Store) getTenantProviderAccount(ctx context.Context, tenantID string, accountID string) (*platform.TenantProviderAccount, error) {
	var (
		account        platform.TenantProviderAccount
		providerCfgRaw []byte
		credentialRaw  []byte
	)

	err := s.pool.QueryRow(ctx, `
		SELECT account_id::text, tenant_id::text, provider, name, auth_mode, provider_config, COALESCE(credential_data, '{}'::jsonb),
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
		&credentialRaw,
		&account.HasCredential,
		&account.AllowsVirtualKeysOnly,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	account.ProviderConfig = parseJSONObject(providerCfgRaw)
	account.CredentialHint = credentialHintFromMap(parseJSONObject(credentialRaw))

	return &account, nil
}

func (s *Store) listTenantProviderAccountModels(ctx context.Context, accountID string) ([]platform.TenantProviderAccountModel, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT model,
		       enabled,
		       COALESCE(types, '{}'),
		       COALESCE(input_modalities, '{}'),
		       COALESCE(output_modalities, '{}'),
		       pricing_mode,
		       public_pricing::text,
		       private_pricing::text,
		       COALESCE(settings, '{}'::jsonb)
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
		var (
			model              platform.TenantProviderAccountModel
			types              []string
			inputModalities    []string
			outputModalities   []string
			publicPricingText  *string
			privatePricingText *string
			settingsRaw        []byte
		)
		if err := rows.Scan(
			&model.Model,
			&model.Enabled,
			&types,
			&inputModalities,
			&outputModalities,
			&model.PricingMode,
			&publicPricingText,
			&privatePricingText,
			&settingsRaw,
		); err != nil {
			return nil, err
		}

		model.Types = stringsToModelTypes(types)
		model.SupportedModes = append([]platform.ModelType(nil), model.Types...)
		model.InputModalities = normalizeStringList(inputModalities)
		model.OutputModalities = normalizeStringList(outputModalities)
		model.Settings = parseJSONObject(settingsRaw)

		if publicPricingText != nil && strings.TrimSpace(*publicPricingText) != "" && strings.TrimSpace(*publicPricingText) != "null" {
			var parsed platform.ModelPrice
			if err := json.Unmarshal([]byte(*publicPricingText), &parsed); err == nil {
				model.PublicPricing = &parsed
			}
		}

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
		var (
			role      platform.AccessRole
			scope     platform.AccessScope
			model     string
			principal platform.AccessPrincipal
		)
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

// IsTenantProviderAccountManager returns true when the given user/team is in manager bucket for the account.
func (s *Store) IsTenantProviderAccountManager(ctx context.Context, tenantID string, accountID string, userID string, teamID string) (bool, error) {
	normalizedUserID := strings.TrimSpace(userID)

	normalizedTeamID := strings.TrimSpace(teamID)
	if normalizedUserID == "" && normalizedTeamID == "" {
		return false, nil
	}

	var allowed bool
	if err := s.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM tenant_provider_account_access ac
			INNER JOIN tenant_provider_accounts a ON a.account_id = ac.account_id
			WHERE a.tenant_id = $1::uuid
			  AND a.account_id = $2::uuid
			  AND ac.role = 'manager'
			  AND (
			    (ac.principal_type = 'user' AND ac.principal_id IS NOT NULL AND ac.principal_id::text = NULLIF($3, ''))
			    OR (ac.principal_type = 'team' AND ac.principal_id IS NOT NULL AND ac.principal_id::text = NULLIF($4, ''))
			    OR (ac.principal_type = 'org_unit' AND ac.principal_id IS NOT NULL AND (
			        (
			          NULLIF($3, '') IS NOT NULL
			          AND EXISTS (
			            SELECT 1
			            FROM organization_memberships om
			            WHERE om.tenant_id = $1::uuid
			              AND om.organization_id::text = ac.principal_id::text
			              AND om.user_id::text = NULLIF($3, '')
			          )
			        )
			        OR (
			          NULLIF($4, '') IS NOT NULL
			          AND EXISTS (
			            SELECT 1
			            FROM teams t
			            WHERE t.tenant_id = $1::uuid
			              AND t.team_id::text = NULLIF($4, '')
			              AND t.org_unit_id::text = ac.principal_id::text
			          )
			        )
			    ))
			  )
		)
	`, tenantID, accountID, normalizedUserID, normalizedTeamID).Scan(&allowed); err != nil {
		return false, err
	}

	return allowed, nil
}

// DeleteTenantProviderAccount deletes one provider account scoped to a tenant.
func (s *Store) DeleteTenantProviderAccount(ctx context.Context, tenantID string, accountID string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer rollbackTxBestEffort(ctx, tx)

	var (
		accountName string
		provider    string
	)

	accountErr := tx.QueryRow(ctx, `
		SELECT name, provider
		FROM tenant_provider_accounts
		WHERE tenant_id = $1::uuid AND account_id = $2::uuid
	`, tenantID, accountID).Scan(&accountName, &provider)
	if accountErr != nil {
		return accountErr
	}

	var hasActiveKeyReference bool

	refErr := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM verification_tokens vt
			INNER JOIN LATERAL unnest(COALESCE(vt.models, ARRAY[]::text[])) AS m(model) ON true
			WHERE vt.tenant_id = $1::uuid
			  AND vt.revoked_at IS NULL
			  AND (
			    m.model ILIKE ($2 || ':' || $3 || ':%')
			    OR m.model ILIKE ($2 || ':' || $3)
			  )
		)
	`, tenantID, strings.TrimSpace(provider), strings.TrimSpace(accountName)).Scan(&hasActiveKeyReference)
	if refErr != nil {
		return refErr
	}

	if hasActiveKeyReference {
		return ErrProviderAccountDependencyConflict
	}

	cmd, err := tx.Exec(ctx, `
		DELETE FROM tenant_provider_accounts
		WHERE tenant_id = $1::uuid AND account_id = $2::uuid
	`, tenantID, accountID)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	if s.logger != nil {
		s.logger.Info("provider_account.deleted",
			zap.String("tenant_id", tenantID),
			zap.String("account_id", accountID),
			zap.String("provider", provider),
			zap.String("name", accountName),
		)
	}

	s.invalidateDeploymentCache(ctx, tenantID)
	s.invalidateTenantRuntimeProfile(ctx, tenantID, "provider_account.deleted")
	s.warmDeploymentCacheAsync(tenantID)

	return nil
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
	// Ensure account exists and belongs to tenant.
	if _, err := s.getTenantProviderAccount(ctx, tenantID, accountID); err != nil {
		return nil, err
	}

	normalizedKeyType := strings.TrimSpace(strings.ToLower(keyType))
	normalizedUserID := strings.TrimSpace(userID)
	normalizedTeamID := strings.TrimSpace(teamID)

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
		           AND ac.role = 'user'
		           AND ac.access_scope = 'account'
		           AND (
		             ac.principal_type = 'everyone'
		             OR (ac.principal_type = 'team' AND ac.principal_id IS NOT NULL AND (
		               ac.principal_id::text = NULLIF($3, '')
		               OR (
		                 NULLIF($2, '') IS NOT NULL
		                 AND EXISTS (
		                   SELECT 1
		                   FROM team_memberships tm
		                   WHERE tm.tenant_id = $5::uuid
		                     AND tm.team_id::text = ac.principal_id::text
		                     AND tm.user_id::text = NULLIF($2, '')
		                 )
		               )
		             ))
		             OR (ac.principal_type = 'user' AND ac.principal_id IS NOT NULL AND ac.principal_id::text = NULLIF($2, ''))
		             OR (ac.principal_type = 'org_unit' AND ac.principal_id IS NOT NULL AND (
		               (
		                 NULLIF($2, '') IS NOT NULL
		                 AND EXISTS (
		                   SELECT 1
		                   FROM organization_memberships om
		                   WHERE om.tenant_id = $5::uuid
		                     AND om.organization_id::text = ac.principal_id::text
		                     AND om.user_id::text = NULLIF($2, '')
		                 )
		               )
		               OR (
		                 NULLIF($2, '') IS NOT NULL
		                 AND EXISTS (
		                   SELECT 1
		                   FROM team_memberships tm
		                   JOIN teams t ON t.team_id = tm.team_id
		                   WHERE tm.tenant_id = $5::uuid
		                     AND tm.user_id::text = NULLIF($2, '')
		                     AND t.org_unit_id::text = ac.principal_id::text
		                 )
		               )
		               OR (
		                 NULLIF($3, '') IS NOT NULL
		                 AND EXISTS (
		                   SELECT 1
		                   FROM teams t
		                   WHERE t.tenant_id = $5::uuid
		                     AND t.team_id::text = NULLIF($3, '')
		                     AND t.org_unit_id::text = ac.principal_id::text
		                 )
		               )
		             ))
		           )
		       ) AS allowed_by_access,
		       COALESCE((
		         SELECT COUNT(*) > 0
		         FROM verification_tokens vt
		         WHERE vt.tenant_id = $5::uuid
		           AND COALESCE(vt.key_type, CASE WHEN vt.team_id IS NULL THEN 'pat' ELSE 'vat' END) = NULLIF($4, '')
		           AND vt.revoked_at IS NULL
		           AND (vt.expires IS NULL OR vt.expires > NOW())
		           AND COALESCE(vt.blocked, false) = false
		           AND (
		             (NULLIF($4, '') = 'pat' AND vt.user_id IS NOT NULL AND vt.user_id::text = NULLIF($2, ''))
		             OR (NULLIF($4, '') = 'vat' AND vt.team_id IS NOT NULL AND vt.team_id::text = NULLIF($3, ''))
		           )
		       ), false) AS has_candidate_keys
		FROM tenant_provider_account_models m
		JOIN tenant_provider_accounts pa ON pa.account_id = m.account_id
		WHERE m.account_id = $1
		ORDER BY m.model ASC
	`, accountID, normalizedUserID, normalizedTeamID, normalizedKeyType, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]platform.EffectiveModelView, 0)

	for rows.Next() {
		var (
			item             platform.EffectiveModelView
			types            []string
			pricingMode      string
			allowedByAccess  bool
			hasCandidateKeys bool
		)
		if err := rows.Scan(
			&item.Model,
			&item.Enabled,
			&types,
			&pricingMode,
			&item.PricingResolved,
			&allowedByAccess,
			&hasCandidateKeys,
		); err != nil {
			return nil, err
		}

		item.PricingMode = platform.PricingMode(pricingMode)
		item.EffectiveCapabilities = stringsToModelTypes(types)

		item.Allowed = allowedByAccess && hasCandidateKeys
		if !item.Allowed {
			switch {
			case !allowedByAccess:
				item.DeniedReason = "denied_by_access_mapping"
			case !hasCandidateKeys:
				item.DeniedReason = "no_active_pat/vat"
			}
		}

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
	case platform.PrincipalTypeUser, platform.PrincipalTypeTeam, platform.PrincipalTypeOrgUnit:
		if principal.PrincipalID == "" {
			return platform.AccessPrincipal{}, fmt.Errorf("principal_id is required for user/team/org_unit entries")
		}
	case platform.PrincipalTypeEveryone:
		if role == platform.AccessRoleManager {
			return platform.AccessPrincipal{}, fmt.Errorf("manager role only supports user/team/org_unit principals")
		}

		principal.PrincipalID = ""
	default:
		return platform.AccessPrincipal{}, fmt.Errorf("unsupported principal type: %s", principal.Type)
	}

	if principal.Type == platform.PrincipalTypeUser || principal.Type == platform.PrincipalTypeTeam || principal.Type == platform.PrincipalTypeOrgUnit {
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
			SELECT EXISTS (
				SELECT 1
				FROM users u
				WHERE u.tenant_id = $1::uuid
				  AND u.user_id::text = $2
			)
		`, tenantID, principalID).Scan(&exists); err != nil {
			return false, err
		}

		return exists, nil
	case platform.PrincipalTypeOrgUnit:
		var exists bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1
				FROM organizations o
				WHERE o.tenant_id = $1::uuid
				  AND o.organization_id::text = $2
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

func normalizeStringList(values []string) []string {
	items := make([]string, 0, len(values))

	seen := make(map[string]struct{}, len(values))
	for _, item := range values {
		value := strings.TrimSpace(item)
		if value == "" {
			continue
		}

		if _, ok := seen[value]; ok {
			continue
		}

		seen[value] = struct{}{}
		items = append(items, value)
	}

	return items
}
