package storage

import (
	"context"
	"encoding/json"
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

	var pricing []byte
	if dep.EffectivePricing != nil {
		pricing, err = json.Marshal(dep.EffectivePricing)
		if err != nil {
			return nil, err
		}
	}

	_, err = s.pool.Exec(ctx, `
		INSERT INTO model_deployments (
			id, tenant_id, model_name, provider, model, provider_config, model_info,
			rpm, tpm, timeout, credential_source, billing_mode, effective_pricing, pricing_source
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`, dep.ID, nullUUID(dep.TenantID), dep.ModelName, dep.Provider, dep.Model, params, info,
		dep.RPM, dep.TPM, dep.TimeoutSec, emptyToNil(dep.CredentialSource),
		emptyToNil(dep.BillingMode), pricing, emptyToNil(dep.PricingSource))
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
		       provider_config, COALESCE(model_info, '{}'),
		       COALESCE(credential_source, ''), COALESCE(billing_mode, ''),
		       effective_pricing, COALESCE(pricing_source, '')
		FROM model_deployments
		WHERE (tenant_id IS NULL OR tenant_id = $1)
		  AND (catalog_managed = false OR catalog_available = true)
		ORDER BY created_at ASC
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deps []platform.Deployment

	for rows.Next() {
		var (
			dep        platform.Deployment
			paramsRaw  []byte
			infoRaw    []byte
			pricingRaw []byte
		)

		scanErr := rows.Scan(
			&dep.ID, &dep.TenantID, &dep.ModelName, &dep.Provider, &dep.Model,
			&paramsRaw, &infoRaw, &dep.CredentialSource, &dep.BillingMode,
			&pricingRaw, &dep.PricingSource,
		)
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

		if len(pricingRaw) > 0 {
			if unmarshalErr := json.Unmarshal(pricingRaw, &dep.EffectivePricing); unmarshalErr != nil {
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
		       COALESCE(a.credential_data, '{}'::jsonb),
		       a.has_credential,
		       a.use_platform_managed_credentials,
		       m.model,
		       COALESCE(m.types, '{}'),
		       COALESCE(m.settings, '{}'::jsonb),
		       m.pricing_mode,
		       m.public_pricing,
		       m.private_pricing
		FROM tenant_provider_accounts a
		JOIN tenant_provider_account_models m
		  ON m.account_id = a.account_id
		 AND m.enabled = true
		 AND m.catalog_available = true
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
			credentialRaw              []byte
			hasCredential              bool
			platformManagedCredentials bool
			model                      string
			modelTypes                 []string
			settingsRaw                []byte
			pricingMode                string
			publicPricingRaw           []byte
			privatePricingRaw          []byte
		)
		if err := rows.Scan(&accountID, &accountTenant, &provider, &accountName, &authMode, &providerRaw, &credentialRaw, &hasCredential, &platformManagedCredentials, &model, &modelTypes, &settingsRaw, &pricingMode, &publicPricingRaw, &privatePricingRaw); err != nil {
			return nil, err
		}

		provider = strings.TrimSpace(provider)
		accountName = strings.TrimSpace(accountName)

		model = strings.TrimSpace(model)
		if provider == "" || accountName == "" || model == "" {
			continue
		}

		providerConfig := parseJSONObject(providerRaw)
		credentials := parseJSONObject(credentialRaw)
		settings := parseJSONObject(settingsRaw)
		apiKey := strings.TrimSpace(toString(credentials["api_key"]))
		rawCredentials := parseNestedJSONObject(credentials["raw"])
		runtimeAPIKey := ""
		credentialSource := platform.CredentialSourceGodarManagedKey
		billingMode := platform.BillingModeWalletBilled
		noCredential := providerAccountUsesNoCredential(provider, authMode)

		var providerAuth *platform.ResolvedProviderAuth

		if apiKey != "" && hasCredential && !platformManagedCredentials && !noCredential {
			runtimeAPIKey = apiKey
			credentialSource = platform.CredentialSourceUserBYOK
			billingMode = platform.BillingModeExternalBilled
			providerAuth = &platform.ResolvedProviderAuth{Mode: string(platform.ProviderAuthModeAPIKey), APIKey: apiKey}
		} else if strings.EqualFold(authMode, string(platform.ProviderAuthModeAWSSigV4)) && hasCredential && !platformManagedCredentials {
			credentialSource = platform.CredentialSourceUserBYOK
			billingMode = platform.BillingModeExternalBilled
			providerAuth = &platform.ResolvedProviderAuth{
				Mode:               string(platform.ProviderAuthModeAWSSigV4),
				AWSAccessKeyID:     firstCredential(rawCredentials, "access_key_id", "aws_access_key_id", "access_key"),
				AWSSecretAccessKey: firstCredential(rawCredentials, "secret_access_key", "aws_secret_access_key", "secret_key"),
				AWSSessionToken:    firstCredential(rawCredentials, "session_token", "aws_session_token"),
			}
		} else if strings.EqualFold(authMode, string(platform.ProviderAuthModeGoogleServiceAccount)) && hasCredential && !platformManagedCredentials {
			credentialSource = platform.CredentialSourceUserBYOK
			billingMode = platform.BillingModeExternalBilled
			providerAuth = &platform.ResolvedProviderAuth{
				Mode:              string(platform.ProviderAuthModeGoogleServiceAccount),
				GoogleClientEmail: firstCredential(rawCredentials, "client_email"),
				GooglePrivateKey:  firstCredential(rawCredentials, "private_key"),
				GoogleTokenURI:    firstCredential(rawCredentials, "token_uri"),
			}
		} else if noCredential {
			providerAuth = &platform.ResolvedProviderAuth{Mode: string(platform.ProviderAuthModeNoCredential)}

			billingMode = platform.BillingModeExternalBilled
		}

		apiBase := safeProviderAPIBase(provider, firstString(providerConfig, settings, "api_base", "base_url"))
		effectivePricing, pricingSource, err := effectiveDeploymentPricing(pricingMode, publicPricingRaw, privatePricingRaw)
		if err != nil {
			return nil, err
		}

		deps = append(deps, platform.Deployment{
			ID:                  "provider-account:" + accountID + ":" + model,
			TenantID:            accountTenant,
			ModelName:           modelid.BuildModelFQN(provider, accountName, model),
			Provider:            provider,
			Model:               model,
			APIBase:             apiBase,
			APIKey:              runtimeAPIKey,
			APIVersion:          firstString(providerConfig, settings, "api_version"),
			ProjectID:           firstString(providerConfig, settings, "project_id", "google_project_id"),
			RegionName:          firstString(providerConfig, settings, "region_name", "region", "aws_region"),
			RPM:                 firstInt64(settings, providerConfig, "rpm"),
			TPM:                 firstInt64(settings, providerConfig, "tpm"),
			Weight:              firstFloat64(settings, providerConfig, "weight"),
			TimeoutSec:          firstFloat64(settings, providerConfig, "timeout", "timeout_sec"),
			CredentialSource:    credentialSource,
			BillingMode:         billingMode,
			ProviderAccountID:   accountID,
			ProviderAccountName: accountName,
			ProviderAuth:        providerAuth,
			EffectivePricing:    effectivePricing,
			PricingSource:       pricingSource,
			ModelInfo: map[string]interface{}{
				"source":                "tenant_provider_account",
				"provider_account_id":   accountID,
				"provider_account_name": accountName,
				"credential_source":     credentialSource,
				"billing_mode":          billingMode,
				"supported_modes":       normalizeStringList(modelTypes),
				"pricing_source":        pricingSource,
				"video_output_s3_uri":   firstString(providerConfig, settings, "video_output_s3_uri", "output_s3_uri"),
				"video_output_uri":      firstString(providerConfig, settings, "video_output_uri", "output_gcs_uri", "storage_uri"),
			},
		})
	}

	return deps, rows.Err()
}

func effectiveDeploymentPricing(mode string, publicRaw, privateRaw []byte) (*platform.ModelPrice, string, error) {
	source := "public"
	raw := publicRaw
	if strings.EqualFold(strings.TrimSpace(mode), string(platform.PricingModePrivate)) {
		source = "private"
		raw = privateRaw
	}
	if len(raw) == 0 || string(raw) == "null" {
		return nil, source, nil
	}

	var pricing platform.ModelPrice
	if err := json.Unmarshal(raw, &pricing); err != nil {
		return nil, source, err
	}
	return &pricing, source, nil
}

func parseNestedJSONObject(value interface{}) map[string]interface{} {
	if typed, ok := value.(map[string]interface{}); ok {
		return typed
	}
	return map[string]interface{}{}
}

func firstCredential(credentials map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(toString(credentials[key])); value != "" {
			return value
		}
	}
	return ""
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
