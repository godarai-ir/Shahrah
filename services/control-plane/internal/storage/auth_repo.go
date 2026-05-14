package storage

import (
	"context"
	"crypto/sha256"
	"encoding/json"
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
	var expiresAt *time.Time
	var keyModels []string
	var rawMetadata []byte

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
		       COALESCE(expires, NULL),
		       COALESCE(blocked, false),
		       COALESCE(key_type, CASE WHEN team_id IS NULL THEN 'pat' ELSE 'vat' END),
		       COALESCE(metadata, '{}'::jsonb)
		FROM verification_tokens
		WHERE token_hash = $1
		  AND revoked_at IS NULL
		  AND (expires IS NULL OR expires > NOW())
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
		&expiresAt,
		&auth.Blocked,
		&auth.KeyType,
		&rawMetadata,
	)
	if err != nil {
		return nil, err
	}

	auth.Role = deriveKeyRole(auth.TeamID)
	if auth.KeyType == "" {
		auth.KeyType = deriveKeyType(auth.TeamID)
	}
	auth.OrgUnitID = auth.OrganizationID
	auth.ExpiresAt = expiresAt
	auth.VirtualModelSnapshots = parseVirtualModelSnapshots(rawMetadata)

	runtimeModels, enforceProviderPolicy, err := s.listRuntimeAllowedModelsByProviderAccount(ctx, auth.TenantID, auth.TeamID, auth.UserID)
	if err != nil {
		return nil, err
	}
	auth.EnforceModelAllowlist = enforceProviderPolicy
	auth.AllowedModelTypes = make(map[string][]string)
	auth.AllowedModelPricingMode = make(map[string]string)
	auth.AllowedModelPricingResolved = make(map[string]bool)
	if enforceProviderPolicy {
		intersected, aliasTypes := intersectAllowedModelPoliciesWithSnapshots(keyModels, runtimeModels, auth.VirtualModelSnapshots)
		auth.AllowedModels = make([]string, 0, len(intersected)+len(aliasTypes))
		for _, item := range intersected {
			auth.AllowedModels = append(auth.AllowedModels, item.Model)
			auth.AllowedModelTypes[item.Model] = append([]string(nil), item.Types...)
			auth.AllowedModelPricingMode[item.Model] = item.PricingMode
			auth.AllowedModelPricingResolved[item.Model] = item.PricingResolved
		}
		for alias, types := range aliasTypes {
			if strings.TrimSpace(alias) == "" {
				continue
			}
			auth.AllowedModels = append(auth.AllowedModels, alias)
			auth.AllowedModelTypes[alias] = append([]string(nil), types...)
			auth.AllowedModelPricingMode[alias] = "virtual"
			auth.AllowedModelPricingResolved[alias] = true
		}
	} else {
		auth.AllowedModels = keyModels
	}

	return &auth, nil
}

type runtimeAllowedModelPolicy struct {
	Model           string
	Types           []string
	PricingMode     string
	PricingResolved bool
}

func (s *Store) listRuntimeAllowedModelsByProviderAccount(ctx context.Context, tenantID string, teamID string, userID string) ([]runtimeAllowedModelPolicy, bool, error) {
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
		SELECT DISTINCT CONCAT(a.provider, ':', a.name, ':', m.model) AS model_fqn,
		       COALESCE(m.types, '{}'),
		       COALESCE(m.pricing_mode, 'public'),
		       CASE
		         WHEN COALESCE(m.pricing_mode, 'public') = 'private' THEN m.private_pricing IS NOT NULL
		         ELSE true
		       END AS pricing_resolved
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
			OR (
				ac.principal_type = 'org_unit'
				AND ac.principal_id IS NOT NULL
				AND (
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
						NULLIF($3, '') IS NOT NULL
						AND EXISTS (
							SELECT 1
							FROM team_memberships tm
							JOIN teams t ON t.team_id = tm.team_id
							WHERE tm.tenant_id = $1::uuid
							  AND tm.user_id::text = NULLIF($3, '')
							  AND t.org_unit_id::text = ac.principal_id::text
						)
					)
					OR (
						NULLIF($2, '') IS NOT NULL
						AND EXISTS (
							SELECT 1
							FROM teams t
							WHERE t.tenant_id = $1::uuid
							  AND t.team_id::text = NULLIF($2, '')
							  AND t.org_unit_id::text = ac.principal_id::text
						)
					)
				)
			)
		  )
	`, tenantID, teamID, userID)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	models := make([]runtimeAllowedModelPolicy, 0)
	for rows.Next() {
		var item runtimeAllowedModelPolicy
		if err := rows.Scan(&item.Model, &item.Types, &item.PricingMode, &item.PricingResolved); err != nil {
			return nil, false, err
		}
		item.Model = strings.TrimSpace(item.Model)
		if item.Model != "" {
			models = append(models, item)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
	}

	return models, true, nil
}

func intersectAllowedModelPolicies(keyModels []string, runtimeModels []runtimeAllowedModelPolicy) []runtimeAllowedModelPolicy {
	if len(runtimeModels) == 0 {
		return []runtimeAllowedModelPolicy{}
	}
	if len(keyModels) == 0 || containsWildcard(keyModels) {
		return dedupeRuntimePolicies(runtimeModels)
	}

	runtimeSet := make(map[string]runtimeAllowedModelPolicy, len(runtimeModels))
	for _, item := range runtimeModels {
		model := strings.TrimSpace(item.Model)
		if model == "" {
			continue
		}
		runtimeSet[model] = item
	}

	out := make([]runtimeAllowedModelPolicy, 0, len(keyModels))
	seen := make(map[string]struct{}, len(keyModels))
	for _, model := range keyModels {
		model = strings.TrimSpace(model)
		if model == "" {
			continue
		}
		item, ok := runtimeSet[model]
		if !ok {
			continue
		}
		if _, ok := seen[model]; ok {
			continue
		}
		seen[model] = struct{}{}
		out = append(out, item)
	}
	return out
}

func dedupeRuntimePolicies(models []runtimeAllowedModelPolicy) []runtimeAllowedModelPolicy {
	out := make([]runtimeAllowedModelPolicy, 0, len(models))
	seen := make(map[string]struct{}, len(models))
	for _, item := range models {
		model := strings.TrimSpace(item.Model)
		if model == "" {
			continue
		}
		if _, ok := seen[model]; ok {
			continue
		}
		seen[model] = struct{}{}
		out = append(out, item)
	}
	return out
}

func intersectAllowedModelPoliciesWithSnapshots(keyModels []string, runtimeModels []runtimeAllowedModelPolicy, snapshots map[string]platform.VirtualModelSnapshot) ([]runtimeAllowedModelPolicy, map[string][]string) {
	base := intersectAllowedModelPolicies(keyModels, runtimeModels)
	aliasTypes := make(map[string][]string)
	if len(snapshots) == 0 || len(runtimeModels) == 0 || len(keyModels) == 0 {
		return base, aliasTypes
	}

	runtimeSet := make(map[string]runtimeAllowedModelPolicy, len(runtimeModels))
	for _, item := range runtimeModels {
		model := strings.TrimSpace(item.Model)
		if model == "" {
			continue
		}
		runtimeSet[model] = item
	}

	requested := make(map[string]struct{}, len(keyModels))
	for _, model := range keyModels {
		model = strings.TrimSpace(model)
		if model == "" {
			continue
		}
		requested[model] = struct{}{}
	}

	for alias, snapshot := range snapshots {
		alias = strings.TrimSpace(alias)
		if alias == "" {
			continue
		}
		if _, ok := requested[alias]; !ok {
			continue
		}
		if len(snapshot.TargetFQNs) == 0 {
			continue
		}
		typeSet := make(map[string]struct{})
		valid := true
		for _, fqn := range snapshot.TargetFQNs {
			fqn = strings.TrimSpace(fqn)
			policy, ok := runtimeSet[fqn]
			if !ok {
				valid = false
				break
			}
			for _, t := range policy.Types {
				t = strings.TrimSpace(t)
				if t == "" {
					continue
				}
				typeSet[t] = struct{}{}
			}
		}
		if !valid {
			continue
		}
		types := make([]string, 0, len(typeSet))
		for t := range typeSet {
			types = append(types, t)
		}
		aliasTypes[alias] = types
	}
	return base, aliasTypes
}

func containsWildcard(models []string) bool {
	for _, model := range models {
		if strings.TrimSpace(model) == "*" {
			return true
		}
	}
	return false
}

func parseVirtualModelSnapshots(raw []byte) map[string]platform.VirtualModelSnapshot {
	snapshots := make(map[string]platform.VirtualModelSnapshot)
	if len(raw) == 0 {
		return snapshots
	}
	var metadata map[string]interface{}
	if err := json.Unmarshal(raw, &metadata); err != nil {
		return snapshots
	}
	rawSnapshots, ok := metadata["virtual_model_snapshots"]
	if !ok || rawSnapshots == nil {
		return snapshots
	}
	rawJSON, err := json.Marshal(rawSnapshots)
	if err != nil {
		return snapshots
	}
	var decoded map[string]platform.VirtualModelSnapshot
	if err := json.Unmarshal(rawJSON, &decoded); err != nil {
		return snapshots
	}
	for alias, item := range decoded {
		alias = strings.TrimSpace(alias)
		if alias == "" {
			continue
		}
		if strings.TrimSpace(item.Alias) == "" {
			item.Alias = alias
		}
		cleanTargets := make([]string, 0, len(item.TargetFQNs))
		seen := make(map[string]struct{})
		for _, target := range item.TargetFQNs {
			target = strings.TrimSpace(target)
			if target == "" {
				continue
			}
			if _, exists := seen[target]; exists {
				continue
			}
			seen[target] = struct{}{}
			cleanTargets = append(cleanTargets, target)
		}
		item.TargetFQNs = cleanTargets
		snapshots[alias] = item
	}
	return snapshots
}

// ListAllowedModelFQNs returns tenant-scoped model FQNs allowed for a principal.
// Exactly one principal dimension should be used: PAT(userID) or VAT(teamID).
func (s *Store) ListAllowedModelFQNs(ctx context.Context, tenantID string, userID string, teamID string) ([]string, error) {
	policies, _, err := s.listRuntimeAllowedModelsByProviderAccount(ctx, tenantID, strings.TrimSpace(teamID), strings.TrimSpace(userID))
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(policies))
	seen := make(map[string]struct{}, len(policies))
	for _, item := range policies {
		model := strings.TrimSpace(item.Model)
		if model == "" {
			continue
		}
		if _, ok := seen[model]; ok {
			continue
		}
		seen[model] = struct{}{}
		out = append(out, model)
	}
	return out, nil
}

// CreateKey creates a tenant-scoped virtual key.
func (s *Store) CreateKey(ctx context.Context, params platform.CreateKeyParams) (*platform.CreatedKey, error) {
	params.Normalize()

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

	now := time.Now().UTC()
	rawMetadata, err := json.Marshal(params.Metadata)
	if err != nil {
		return nil, err
	}
	if len(rawMetadata) == 0 || string(rawMetadata) == "null" {
		rawMetadata = []byte("{}")
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO verification_tokens (
			id, tenant_id, token_hash, key_alias, key_name, models, team_id, organization_id, user_id,
			expires, spend, blocked, key_type, created_by_user_id, metadata, created_at, updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, NULLIF($7, '')::uuid, NULLIF($8, '')::uuid,
			CASE
				WHEN NULLIF($9, '') IS NULL THEN NULL
				WHEN $9 ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$'
					AND EXISTS (SELECT 1 FROM users u WHERE u.user_id = $9::uuid)
				THEN $9::uuid
				ELSE NULL
			END,
			$10, 0, false, $11, NULLIF($12, ''), $13, $14, $14
		)
	`, keyID, params.TenantID, tokenHash, params.KeyAlias, masked, params.Models,
		params.TeamID, params.OrgID, params.UserID, expiresAt, params.KeyType, params.UserID, rawMetadata, now)
	if err != nil {
		return nil, err
	}

	return &platform.CreatedKey{
		ID:              keyID,
		Key:             rawKey,
		KeyAlias:        params.KeyAlias,
		MaskedKey:       masked,
		TenantID:        params.TenantID,
		Role:            params.Role,
		KeyType:         params.KeyType,
		Models:          params.Models,
		AllowedModels:   append([]string(nil), params.Models...),
		TeamID:          params.TeamID,
		OrgID:           params.OrgID,
		OrgUnitID:       params.OrgUnitID,
		UserID:          params.UserID,
		VirtualUserID:   params.VirtualUserID,
		CreatedAt:       &now,
		CreatedByUserID: params.UserID,
		TokenHash:       tokenHash,
	}, nil
}

// ListKeys returns created keys for the tenant.
func (s *Store) ListKeys(ctx context.Context, tenantID string) ([]platform.CreatedKey, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, COALESCE(key_alias, ''), COALESCE(key_name, ''), tenant_id::text,
		       COALESCE(models, '{}'), COALESCE(team_id::text, ''), COALESCE(organization_id::text, ''),
		       COALESCE(spend, 0), COALESCE(blocked, false), COALESCE(expires, NULL), token_hash,
		       COALESCE(key_type, CASE WHEN team_id IS NULL THEN 'pat' ELSE 'vat' END),
		       COALESCE(created_by_user_id, ''), COALESCE(created_at, NULL), COALESCE(last_used_at, NULL), COALESCE(revoked_at, NULL)
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
			&item.TeamID, &item.OrgID,
			&item.Spend, &item.Blocked, &item.ExpiresAt, &item.TokenHash, &item.KeyType,
			&item.CreatedByUserID, &item.CreatedAt, &item.LastUsedAt, &item.RevokedAt); err != nil {
			return nil, err
		}
		item.Models = models
		item.AllowedModels = append([]string(nil), models...)
		item.Role = deriveKeyRole(item.TeamID)
		if item.KeyType == "" {
			item.KeyType = deriveKeyType(item.TeamID)
		}
		item.OrgUnitID = item.OrgID
		keys = append(keys, item)
	}
	return keys, rows.Err()
}

func (s *Store) TouchKeyLastUsed(ctx context.Context, keyID string, usedAt time.Time) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE verification_tokens
		SET last_used_at = $2, updated_at = NOW()
		WHERE id = $1
	`, keyID, usedAt)
	return err
}

func (s *Store) PATExists(ctx context.Context, tenantID string, userID string, name string) (bool, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return false, nil
	}
	var exists bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM verification_tokens
			WHERE tenant_id = $1
			  AND key_type = 'pat'
			  AND COALESCE(created_by_user_id, '') = $2
			  AND key_alias = $3
			  AND revoked_at IS NULL
		)
	`, tenantID, strings.TrimSpace(userID), name).Scan(&exists)
	return exists, err
}

func (s *Store) ListPATs(ctx context.Context, tenantID string, userID string, query string, limit int, offset int) ([]platform.CreatedKey, int, error) {
	query = strings.TrimSpace(query)
	if limit <= 0 {
		limit = 25
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	like := "%"
	if query != "" {
		like = "%" + strings.ToLower(query) + "%"
	}

	var total int
	if err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM verification_tokens
		WHERE tenant_id = $1
		  AND key_type = 'pat'
		  AND COALESCE(created_by_user_id, '') = $2
		  AND ($3 = '' OR LOWER(COALESCE(key_alias, '')) LIKE $4)
		  AND revoked_at IS NULL
	`, tenantID, strings.TrimSpace(userID), query, like).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id::text, COALESCE(key_alias, ''), COALESCE(key_name, ''), tenant_id::text,
		       COALESCE(models, '{}'), COALESCE(team_id::text, ''), COALESCE(organization_id::text, ''),
		       COALESCE(spend, 0), COALESCE(blocked, false), COALESCE(expires, NULL), token_hash,
		       COALESCE(created_by_user_id, ''), COALESCE(created_at, NULL), COALESCE(last_used_at, NULL), COALESCE(revoked_at, NULL)
		FROM verification_tokens
		WHERE tenant_id = $1
		  AND key_type = 'pat'
		  AND COALESCE(created_by_user_id, '') = $2
		  AND ($3 = '' OR LOWER(COALESCE(key_alias, '')) LIKE $4)
		  AND revoked_at IS NULL
		ORDER BY created_at DESC
		LIMIT $5 OFFSET $6
	`, tenantID, strings.TrimSpace(userID), query, like, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]platform.CreatedKey, 0)
	for rows.Next() {
		var item platform.CreatedKey
		var models []string
		if err := rows.Scan(&item.ID, &item.KeyAlias, &item.MaskedKey, &item.TenantID, &models,
			&item.TeamID, &item.OrgID,
			&item.Spend, &item.Blocked, &item.ExpiresAt, &item.TokenHash, &item.CreatedByUserID,
			&item.CreatedAt, &item.LastUsedAt, &item.RevokedAt); err != nil {
			return nil, 0, err
		}
		item.Models = models
		item.AllowedModels = append([]string(nil), models...)
		item.Role = platform.LegacyKeyRoleVirtualKey
		item.KeyType = platform.KeyTypePAT
		item.OrgUnitID = item.OrgID
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (s *Store) RevokePAT(ctx context.Context, tenantID string, userID string, keyID string, revokedByUserID string) (bool, error) {
	tag, err := s.pool.Exec(ctx, `
		UPDATE verification_tokens
		SET revoked_at = NOW(),
		    revoked_by_user_id = NULLIF($4, ''),
		    blocked = true,
		    updated_at = NOW()
		WHERE id = $1
		  AND tenant_id = $2
		  AND key_type = 'pat'
		  AND COALESCE(created_by_user_id, '') = $3
		  AND revoked_at IS NULL
	`, keyID, tenantID, strings.TrimSpace(userID), strings.TrimSpace(revokedByUserID))
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (s *Store) RevokeAllPATs(ctx context.Context, tenantID string, userID string, revokedByUserID string) (int64, error) {
	tag, err := s.pool.Exec(ctx, `
		UPDATE verification_tokens
		SET revoked_at = NOW(),
		    revoked_by_user_id = NULLIF($3, ''),
		    blocked = true,
		    updated_at = NOW()
		WHERE tenant_id = $1
		  AND key_type = 'pat'
		  AND COALESCE(created_by_user_id, '') = $2
		  AND revoked_at IS NULL
	`, tenantID, strings.TrimSpace(userID), strings.TrimSpace(revokedByUserID))
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func deriveKeyRole(teamID string) string {
	if strings.TrimSpace(teamID) != "" {
		return platform.LegacyKeyRoleServiceAccount
	}
	return platform.LegacyKeyRoleVirtualKey
}

func deriveKeyType(teamID string) string {
	if strings.TrimSpace(teamID) != "" {
		return platform.KeyTypeVAT
	}
	return platform.KeyTypePAT
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
