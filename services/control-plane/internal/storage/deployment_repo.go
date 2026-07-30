package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/modelid"
	"github.com/google/uuid"
)

// CreateDeployment stores a model deployment for a tenant.
func (s *Store) CreateDeployment(ctx context.Context, dep platform.Deployment) (*platform.Deployment, error) {
	if dep.ID == "" {
		dep.ID = uuid.NewString()
	}

	apiBase, err := ValidateProviderAPIBase(dep.Provider, dep.APIBase)
	if err != nil {
		return nil, err
	}

	dep.APIBase = apiBase

	params, err := json.Marshal(map[string]interface{}{
		"model":       dep.Model,
		"api_key":     dep.APIKey,
		"api_base":    dep.APIBase,
		"api_version": dep.APIVersion,
		"region_name": dep.RegionName,
		"rpm":         dep.RPM,
		"tpm":         dep.TPM,
		"weight":      dep.Weight,
		"timeout":     dep.TimeoutSec,
	})
	if err != nil {
		return nil, err
	}

	info, err := json.Marshal(dep.ModelInfo)
	if err != nil {
		return nil, err
	}

	_, err = s.pool.Exec(ctx, `
		INSERT INTO model_deployments (
			id, tenant_id, model_name, provider, model, provider_config, model_info, rpm, tpm, timeout
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, dep.ID, nullUUID(dep.TenantID), dep.ModelName, dep.Provider, dep.Model, params, info, dep.RPM, dep.TPM, dep.TimeoutSec)
	if err != nil {
		return nil, err
	}

	s.invalidateDeploymentCache(ctx, dep.TenantID)
	s.invalidateTenantRuntimeProfile(ctx, dep.TenantID, "deployment.created")
	s.warmDeploymentCacheAsync(dep.TenantID)

	return &dep, nil
}

// ListDeployments returns tenant-specific deployments.
func (s *Store) ListDeployments(ctx context.Context, tenantID string) ([]platform.Deployment, error) {
	cacheKey := deploymentCacheKey(tenantID)

	return loadHotPathList(
		&s.hotPathLoadGroup, cacheKey, "deployment",
		func() ([]platform.Deployment, hotPathCacheState) { return s.getCachedDeployments(cacheKey) },
		func() { s.refreshDeploymentsCacheAsync(tenantID, cacheKey) },
		func() ([]platform.Deployment, error) { return s.listDeploymentsFromDB(ctx, tenantID) },
		func(items []platform.Deployment) { s.cacheDeployments(cacheKey, items) },
		cloneDeployments,
	)
}

func (s *Store) listDeploymentsFromDB(ctx context.Context, tenantID string) ([]platform.Deployment, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, COALESCE(tenant_id::text, ''), model_name, provider, model,
		       provider_config, COALESCE(model_info, '{}')
		FROM model_deployments
		WHERE tenant_id IS NULL OR tenant_id = $1
		ORDER BY created_at ASC
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deps []platform.Deployment

	for rows.Next() {
		var (
			dep       platform.Deployment
			paramsRaw []byte
			infoRaw   []byte
		)

		scanErr := rows.Scan(&dep.ID, &dep.TenantID, &dep.ModelName, &dep.Provider, &dep.Model, &paramsRaw, &infoRaw)
		if scanErr != nil {
			return nil, scanErr
		}

		var params map[string]interface{}
		if jsonErr := json.Unmarshal(paramsRaw, &params); jsonErr == nil {
			dep.APIBase = safeProviderAPIBase(dep.Provider, toString(params["api_base"]))
			dep.APIKey = toString(params["api_key"])
			dep.APIVersion = toString(params["api_version"])
			dep.RegionName = toString(params["region_name"])
			dep.RPM = toInt64(params["rpm"])
			dep.TPM = toInt64(params["tpm"])
			dep.Weight = toFloat64(params["weight"])
			dep.TimeoutSec = toFloat64(params["timeout"])
		}

		if len(infoRaw) > 0 {
			if unmarshalErr := json.Unmarshal(infoRaw, &dep.ModelInfo); unmarshalErr != nil {
				return nil, unmarshalErr
			}
		}

		deps = append(deps, dep)
	}

	rowsErr := rows.Err()
	if rowsErr != nil {
		return nil, rowsErr
	}

	providerAccountDeps, err := s.listProviderAccountDeployments(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	deps = append(deps, providerAccountDeps...)

	return deps, nil
}

func (s *Store) listProviderAccountDeployments(ctx context.Context, tenantID string) ([]platform.Deployment, error) {
	if strings.TrimSpace(tenantID) == "" {
		return nil, nil
	}

	rows, err := s.pool.Query(ctx, `
		SELECT a.account_id::text,
		       a.tenant_id::text,
		       a.provider,
		       a.name,
		       a.auth_mode,
		       COALESCE(a.provider_config, '{}'::jsonb),
		       a.has_credential,
		       a.use_platform_managed_credentials,
		       m.model,
		       COALESCE(m.types, '{}'),
		       COALESCE(m.settings, '{}'::jsonb),
		       COALESCE(m.pricing_mode, 'public'),
		       m.public_pricing,
		       m.private_pricing
		FROM tenant_provider_accounts a
		JOIN tenant_provider_account_models m
		  ON m.account_id = a.account_id
		 AND m.enabled = true
		WHERE a.tenant_id = $1
		ORDER BY a.created_at ASC, m.created_at ASC, m.model ASC
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	deps := make([]platform.Deployment, 0)

	for rows.Next() {
		var (
			accountID                  string
			accountTenant              string
			provider                   string
			accountName                string
			authMode                   string
			providerRaw                []byte
			hasCredential              bool
			platformManagedCredentials bool
			model                      string
			modelTypes                 []string
			settingsRaw                []byte
			pricingMode                string
			publicPricingRaw           []byte
			privatePricingRaw          []byte
		)
		if err := rows.Scan(
			&accountID,
			&accountTenant,
			&provider,
			&accountName,
			&authMode,
			&providerRaw,
			&hasCredential,
			&platformManagedCredentials,
			&model,
			&modelTypes,
			&settingsRaw,
			&pricingMode,
			&publicPricingRaw,
			&privatePricingRaw,
		); err != nil {
			return nil, err
		}

		provider = strings.TrimSpace(provider)
		accountName = strings.TrimSpace(accountName)

		model = strings.TrimSpace(model)
		if provider == "" || accountName == "" || model == "" {
			continue
		}

		providerConfig := parseJSONObject(providerRaw)
		settings := parseJSONObject(settingsRaw)
		effectivePricing, pricingErr := effectiveDeploymentPricing(
			pricingMode,
			publicPricingRaw,
			privatePricingRaw,
		)
		if pricingErr != nil {
			return nil, pricingErr
		}
		credentialSource := platform.CredentialSourceGodarManagedKey
		billingMode := platform.BillingModeWalletBilled
		noCredential := providerAccountUsesNoCredential(provider, authMode)

		var providerAuth *platform.ResolvedProviderAuth

		if hasCredential && !platformManagedCredentials && !noCredential {
			credentialSource = platform.CredentialSourceUserBYOK
			billingMode = platform.BillingModeExternalBilled
			if strings.EqualFold(authMode, string(platform.ProviderAuthModeAnthropicWIF)) {
				providerAuth = &platform.ResolvedProviderAuth{
					Mode: string(platform.ProviderAuthModeAnthropicWIF),
					AnthropicWIF: &platform.ResolvedAnthropicWIFAuth{
						FederationRuleID: firstString(providerConfig, nil, "federation_rule_id"),
						OrganizationID:   firstString(providerConfig, nil, "organization_id"),
						ServiceAccountID: firstString(providerConfig, nil, "service_account_id"),
						WorkspaceID:      firstString(providerConfig, nil, "workspace_id"),
						TokenSourceID:    firstString(providerConfig, nil, "token_source_id"),
					},
				}
			}
		} else if noCredential {
			providerAuth = &platform.ResolvedProviderAuth{Mode: string(platform.ProviderAuthModeNoCredential)}

			billingMode = platform.BillingModeExternalBilled
		}

		apiBase := safeProviderAPIBase(provider, firstString(providerConfig, settings, "api_base", "base_url"))

		deps = append(deps, platform.Deployment{
			ID:                  "provider-account:" + accountID + ":" + model,
			TenantID:            accountTenant,
			ModelName:           modelid.BuildModelFQN(provider, accountName, model),
			Provider:            provider,
			Model:               model,
			APIBase:             apiBase,
			APIKey:              "",
			APIVersion:          firstString(providerConfig, settings, "api_version"),
			RegionName:          firstString(providerConfig, settings, "region_name", "region", "aws_region"),
			RPM:                 firstInt64(settings, providerConfig, "rpm"),
			TPM:                 firstInt64(settings, providerConfig, "tpm"),
			Weight:              firstFloat64(settings, providerConfig, "weight"),
			TimeoutSec:          firstFloat64(settings, providerConfig, "timeout", "timeout_sec"),
			CredentialSource:    credentialSource,
			BillingMode:         billingMode,
			PricingMode:         platform.PricingMode(pricingMode),
			ProviderAccountID:   accountID,
			ProviderAccountName: accountName,
			ProviderAuth:        providerAuth,
			ProviderHeaders:     documentedProviderHeaders(provider, providerConfig),
			EffectivePricing:    effectivePricing,
			ModelInfo: map[string]interface{}{
				"source":                "tenant_provider_account",
				"provider_account_id":   accountID,
				"provider_account_name": accountName,
				"credential_source":     credentialSource,
				"billing_mode":          billingMode,
				"supported_modes":       normalizeStringList(modelTypes),
			},
		})
	}

	return deps, rows.Err()
}

func documentedProviderHeaders(
	provider string,
	config map[string]interface{},
) map[string]string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "openai":
		headers := make(map[string]string, 2)
		if value := firstString(config, nil, "organization", "openai_organization"); value != "" {
			headers["OpenAI-Organization"] = value
		}
		if value := firstString(config, nil, "project", "openai_project"); value != "" {
			headers["OpenAI-Project"] = value
		}
		if len(headers) == 0 {
			return nil
		}
		return headers
	case "anthropic":
	default:
		return nil
	}
	values := providerConfigStringList(config["beta_allowlist"])
	if len(values) == 0 {
		return nil
	}
	return map[string]string{
		"Anthropic-Beta-Allowlist": strings.Join(values, ","),
	}
}

func effectiveDeploymentPricing(
	mode string,
	publicRaw []byte,
	privateRaw []byte,
) (*platform.ModelPrice, error) {
	decode := func(raw []byte) (*platform.ModelPrice, error) {
		if len(raw) == 0 || string(raw) == "null" {
			return nil, nil
		}
		var value platform.ModelPrice
		if err := json.Unmarshal(raw, &value); err != nil {
			return nil, err
		}
		return &value, nil
	}
	publicPrice, err := decode(publicRaw)
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(strings.TrimSpace(mode), string(platform.PricingModePrivate)) {
		return publicPrice, nil
	}
	privatePrice, err := decode(privateRaw)
	if err != nil {
		return nil, err
	}
	if privatePrice == nil {
		return nil, nil
	}
	if publicPrice == nil {
		return privatePrice, nil
	}
	merged := *publicPrice
	merged.Rates = append([]platform.ModelPricingRate(nil), publicPrice.Rates...)
	index := make(map[string]int, len(merged.Rates))
	for i, rate := range merged.Rates {
		index[strings.TrimSpace(rate.Key)] = i
	}
	for _, override := range privatePrice.Rates {
		key := strings.TrimSpace(override.Key)
		position, ok := index[key]
		if key == "" || !ok {
			return nil, fmt.Errorf("private pricing rate %q does not match the public rate card", key)
		}
		publicRate := merged.Rates[position]
		if override.Dimension != publicRate.Dimension ||
			override.Unit != publicRate.Unit {
			return nil, fmt.Errorf("private pricing rate %q changes immutable rate identity", key)
		}
		publicRate.UnitPriceUSD = override.UnitPriceUSD
		merged.Rates[position] = publicRate
	}
	if privatePrice.InputUSDPer1M != nil {
		merged.InputUSDPer1M = privatePrice.InputUSDPer1M
	}
	if privatePrice.OutputUSDPer1M != nil {
		merged.OutputUSDPer1M = privatePrice.OutputUSDPer1M
	}
	return &merged, nil
}

// anthropicProviderHeaders is kept for focused compatibility tests and callers
// compiled during rolling deployments.
func anthropicProviderHeaders(provider string, config map[string]interface{}) map[string]string {
	return documentedProviderHeaders(provider, config)
}

func providerConfigStringList(value interface{}) []string {
	var input []string
	switch typed := value.(type) {
	case string:
		input = strings.Split(typed, ",")
	case []string:
		input = typed
	case []interface{}:
		input = make([]string, 0, len(typed))
		for _, item := range typed {
			if text, ok := item.(string); ok {
				input = append(input, text)
			}
		}
	}
	seen := make(map[string]struct{}, len(input))
	result := make([]string, 0, len(input))
	for _, item := range input {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}

func providerAccountUsesNoCredential(provider, authMode string) bool {
	return strings.EqualFold(strings.TrimSpace(provider), "mock") ||
		strings.EqualFold(strings.TrimSpace(authMode), string(platform.ProviderAuthModeNoCredential))
}

func firstString(primary, secondary map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if v := strings.TrimSpace(toString(primary[key])); v != "" {
			return v
		}

		if v := strings.TrimSpace(toString(secondary[key])); v != "" {
			return v
		}
	}

	return ""
}

func firstInt64(primary, secondary map[string]interface{}, keys ...string) int64 {
	for _, key := range keys {
		if v := toInt64(primary[key]); v != 0 {
			return v
		}

		if v := toInt64(secondary[key]); v != 0 {
			return v
		}
	}

	return 0
}

func firstFloat64(primary, secondary map[string]interface{}, keys ...string) float64 {
	for _, key := range keys {
		if v := toFloat64(primary[key]); v != 0 {
			return v
		}

		if v := toFloat64(secondary[key]); v != 0 {
			return v
		}
	}

	return 0
}
