package storage

import (
	"context"
	"encoding/json"

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
	return deps, rows.Err()
}
