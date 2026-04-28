package storage

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/google/uuid"
)

// GetKeyByHash retrieves a key by its hashed token.
func (s *Store) GetKeyByHash(ctx context.Context, hash string) (*platform.AuthContext, error) {
	var auth platform.AuthContext
	var maxBudget *float64
	var rpmLimit *int64
	var tpmLimit *int64
	var expiresAt *time.Time
	var keyModels []string

	err := s.pool.QueryRow(ctx, `
		SELECT id::text,
		       tenant_id::text,
		       COALESCE(organization_id::text, ''),
		       COALESCE(team_id::text, ''),
		       COALESCE(user_id::text, ''),
		       COALESCE(key_alias, ''),
		       COALESCE(token_hash, ''),
		       COALESCE(models, '{}'),
		       COALESCE(spend, 0),
		       max_budget,
		       rpm_limit,
		       tpm_limit,
		       COALESCE(expires, NULL),
		       COALESCE(blocked, false)
		FROM verification_tokens
		WHERE token_hash = $1
		LIMIT 1
	`, hash).Scan(
		&auth.KeyID,
		&auth.TenantID,
		&auth.OrganizationID,
		&auth.TeamID,
		&auth.UserID,
		&auth.KeyAlias,
		&auth.TokenHash,
		&keyModels,
		&auth.Spend,
		&maxBudget,
		&rpmLimit,
		&tpmLimit,
		&expiresAt,
		&auth.Blocked,
	)
	if err != nil {
		return nil, err
	}

	auth.Role = deriveKeyRole(auth.TeamID)
	auth.MaxBudget = maxBudget
	auth.RPMLimit = rpmLimit
	auth.TPMLimit = tpmLimit
	auth.ExpiresAt = expiresAt

	runtimeModels, enforceProviderPolicy, err := s.listRuntimeAllowedModelsByProviderAccount(ctx, auth.TenantID, auth.TeamID, auth.UserID)
	if err != nil {
		return nil, err
	}
	if enforceProviderPolicy {
		auth.AllowedModels = intersectAllowedModels(keyModels, runtimeModels)
	} else {
		auth.AllowedModels = keyModels
	}

	return &auth, nil
}

func (s *Store) listRuntimeAllowedModelsByProviderAccount(ctx context.Context, tenantID string, teamID string, userID string) ([]string, bool, error) {
	var hasEnabledModels bool
	if err := s.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM tenant_provider_accounts a
			JOIN tenant_provider_account_models m ON m.account_id = a.account_id
			WHERE a.tenant_id = $1
			  AND m.enabled = true
		)
	`, tenantID).Scan(&hasEnabledModels); err != nil {
		return nil, false, err
	}
	if !hasEnabledModels {
		return nil, false, nil
	}

	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT m.model
		FROM tenant_provider_accounts a
		JOIN tenant_provider_account_models m
		  ON m.account_id = a.account_id
		 AND m.enabled = true
		JOIN tenant_provider_account_access ac
		  ON ac.account_id = a.account_id
		 AND ac.role IN ('manager', 'user')
		WHERE a.tenant_id = $1
		  AND (
			   ac.principal_type = 'everyone'
			OR (ac.principal_type = 'team' AND ac.principal_id = NULLIF($2, ''))
			OR (ac.principal_type = 'user' AND ac.principal_id = NULLIF($3, ''))
		  )
	`, tenantID, teamID, userID)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	models := make([]string, 0)
	for rows.Next() {
		var model string
		if err := rows.Scan(&model); err != nil {
			return nil, false, err
		}
		model = strings.TrimSpace(model)
		if model != "" {
			models = append(models, model)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
	}

	return models, true, nil
}

func intersectAllowedModels(keyModels []string, runtimeModels []string) []string {
	if len(runtimeModels) == 0 {
		return []string{}
	}
	if len(keyModels) == 0 || containsWildcard(keyModels) {
		return dedupeModels(runtimeModels)
	}

	runtimeSet := make(map[string]struct{}, len(runtimeModels))
	for _, model := range runtimeModels {
		model = strings.TrimSpace(model)
		if model == "" {
			continue
		}
		runtimeSet[model] = struct{}{}
	}

	out := make([]string, 0, len(keyModels))
	seen := make(map[string]struct{}, len(keyModels))
	for _, model := range keyModels {
		model = strings.TrimSpace(model)
		if model == "" {
			continue
		}
		if _, ok := runtimeSet[model]; !ok {
			continue
		}
		if _, ok := seen[model]; ok {
			continue
		}
		seen[model] = struct{}{}
		out = append(out, model)
	}
	return out
}

func dedupeModels(models []string) []string {
	out := make([]string, 0, len(models))
	seen := make(map[string]struct{}, len(models))
	for _, model := range models {
		model = strings.TrimSpace(model)
		if model == "" {
			continue
		}
		if _, ok := seen[model]; ok {
			continue
		}
		seen[model] = struct{}{}
		out = append(out, model)
	}
	return out
}

func containsWildcard(models []string) bool {
	for _, model := range models {
		if strings.TrimSpace(model) == "*" {
			return true
		}
	}
	return false
}

// CreateKey creates a tenant-scoped virtual key.
func (s *Store) CreateKey(ctx context.Context, params platform.CreateKeyParams) (*platform.CreatedKey, error) {
	rawKey := "sk-" + strings.ReplaceAll(uuid.NewString(), "-", "")
	tokenHash := hashKey(rawKey)
	keyID := uuid.NewString()
	masked := maskKey(rawKey)

	var expiresAt *time.Time
	if params.ExpiresIn != "" {
		duration, err := parseDuration(params.ExpiresIn)
		if err != nil {
			return nil, err
		}
		exp := time.Now().Add(duration)
		expiresAt = &exp
	}

	_, err := s.pool.Exec(ctx, `
		INSERT INTO verification_tokens (
			id, tenant_id, token_hash, key_alias, key_name, models, max_budget,
			rpm_limit, tpm_limit, team_id, organization_id, expires, spend, blocked
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NULLIF($10, '')::uuid, NULLIF($11, '')::uuid, $12, 0, false)
	`, keyID, params.TenantID, tokenHash, params.KeyAlias, masked, params.Models, params.MaxBudget,
		params.RPMLimit, params.TPMLimit, params.TeamID, params.OrgID, expiresAt)
	if err != nil {
		return nil, err
	}

	return &platform.CreatedKey{
		ID:        keyID,
		Key:       rawKey,
		KeyAlias:  params.KeyAlias,
		MaskedKey: masked,
		TenantID:  params.TenantID,
		Role:      params.Role,
		Models:    params.Models,
		MaxBudget: params.MaxBudget,
		RPMLimit:  params.RPMLimit,
		TPMLimit:  params.TPMLimit,
		TeamID:    params.TeamID,
		OrgID:     params.OrgID,
		TokenHash: tokenHash,
	}, nil
}

// ListKeys returns created keys for the tenant.
func (s *Store) ListKeys(ctx context.Context, tenantID string) ([]platform.CreatedKey, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, COALESCE(key_alias, ''), COALESCE(key_name, ''), tenant_id::text,
		       COALESCE(models, '{}'), COALESCE(max_budget, NULL), COALESCE(rpm_limit, NULL),
		       COALESCE(tpm_limit, NULL), COALESCE(team_id::text, ''), COALESCE(organization_id::text, ''),
		       COALESCE(spend, 0), COALESCE(blocked, false), COALESCE(expires, NULL), token_hash
		FROM verification_tokens
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []platform.CreatedKey
	for rows.Next() {
		var item platform.CreatedKey
		var models []string
		if err := rows.Scan(&item.ID, &item.KeyAlias, &item.MaskedKey, &item.TenantID, &models,
			&item.MaxBudget, &item.RPMLimit, &item.TPMLimit, &item.TeamID, &item.OrgID,
			&item.Spend, &item.Blocked, &item.ExpiresAt, &item.TokenHash); err != nil {
			return nil, err
		}
		item.Models = models
		item.Role = deriveKeyRole(item.TeamID)
		keys = append(keys, item)
	}
	return keys, rows.Err()
}

func deriveKeyRole(teamID string) string {
	if strings.TrimSpace(teamID) != "" {
		return "service_account"
	}
	return "virtual_key"
}

func maskKey(key string) string {
	if len(key) <= 8 {
		return key
	}
	return key[:5] + "..." + key[len(key)-4:]
}

func parseDuration(input string) (time.Duration, error) {
	if strings.HasSuffix(input, "d") {
		n, err := strconv.Atoi(strings.TrimSuffix(input, "d"))
		if err != nil {
			return 0, err
		}
		return time.Duration(n) * 24 * time.Hour, nil
	}
	return time.ParseDuration(input)
}

func hashKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return fmt.Sprintf("%x", hash[:])
}
