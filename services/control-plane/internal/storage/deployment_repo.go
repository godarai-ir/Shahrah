package storage

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/google/uuid"
)

// CreateDeployment stores a model deployment for a tenant.
func (s *Store) CreateDeployment(ctx context.Context, dep platform.Deployment) (*platform.Deployment, error) {
	if dep.ID == "" {
		dep.ID = uuid.NewString()
	}
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

	return &dep, nil
}

// ListDeployments returns tenant-specific deployments.
func (s *Store) ListDeployments(ctx context.Context, tenantID string) ([]platform.Deployment, error) {
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
		var dep platform.Deployment
		var paramsRaw []byte
		var infoRaw []byte
		if err := rows.Scan(&dep.ID, &dep.TenantID, &dep.ModelName, &dep.Provider, &dep.Model, &paramsRaw, &infoRaw); err != nil {
			return nil, err
		}
		var params map[string]interface{}
		if err := json.Unmarshal(paramsRaw, &params); err == nil {
			dep.APIBase = toString(params["api_base"])
			dep.APIKey = toString(params["api_key"])
			dep.APIVersion = toString(params["api_version"])
			dep.RegionName = toString(params["region_name"])
			dep.RPM = toInt64(params["rpm"])
			dep.TPM = toInt64(params["tpm"])
			dep.Weight = toFloat64(params["weight"])
			dep.TimeoutSec = toFloat64(params["timeout"])
		}
		if len(infoRaw) > 0 {
			_ = json.Unmarshal(infoRaw, &dep.ModelInfo)
		}
		deps = append(deps, dep)
	}
	if err := rows.Err(); err != nil {
		return nil, err
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
		       COALESCE(a.provider_config, '{}'::jsonb),
		       COALESCE(a.credential_data, '{}'::jsonb),
		       m.model,
		       COALESCE(m.settings, '{}'::jsonb)
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
			accountID     string
			accountTenant string
			provider      string
			accountName   string
			providerRaw   []byte
			credentialRaw []byte
			model         string
			settingsRaw   []byte
		)
		if err := rows.Scan(&accountID, &accountTenant, &provider, &accountName, &providerRaw, &credentialRaw, &model, &settingsRaw); err != nil {
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
		deps = append(deps, platform.Deployment{
			ID:         "provider-account:" + accountID + ":" + model,
			TenantID:   accountTenant,
			ModelName:  provider + ":" + accountName + ":" + model,
			Provider:   provider,
			Model:      model,
			APIBase:    firstString(providerConfig, settings, "api_base", "base_url"),
			APIKey:     toString(credentials["api_key"]),
			APIVersion: firstString(providerConfig, settings, "api_version"),
			RegionName: firstString(providerConfig, settings, "region_name", "region", "aws_region"),
			RPM:        firstInt64(settings, providerConfig, "rpm"),
			TPM:        firstInt64(settings, providerConfig, "tpm"),
			Weight:     firstFloat64(settings, providerConfig, "weight"),
			TimeoutSec: firstFloat64(settings, providerConfig, "timeout", "timeout_sec"),
			ModelInfo: map[string]interface{}{
				"source":                "tenant_provider_account",
				"provider_account_id":   accountID,
				"provider_account_name": accountName,
			},
		})
	}
	return deps, rows.Err()
}

func firstString(primary map[string]interface{}, secondary map[string]interface{}, keys ...string) string {
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

func firstInt64(primary map[string]interface{}, secondary map[string]interface{}, keys ...string) int64 {
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

func firstFloat64(primary map[string]interface{}, secondary map[string]interface{}, keys ...string) float64 {
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
