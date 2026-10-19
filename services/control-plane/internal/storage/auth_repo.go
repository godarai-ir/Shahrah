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
	"github.com/Delpak/shahrah/services/control-plane/internal/apperrors"
	"github.com/Delpak/shahrah/services/control-plane/internal/modelid"
	"github.com/google/uuid"
)

// GetKeyByHash retrieves a key by its hashed token.
func (s *AuthRepository) GetKeyByHash(ctx context.Context, hash string) (*platform.AuthContext, error) {
	var (
		auth        platform.AuthContext
		expiresAt   *time.Time
		keyModels   []string
		rawMetadata []byte
	)

	err := s.pool.QueryRow(ctx, `
		SELECT vt.id::text,
		       vt.tenant_id::text,
		       COALESCE(vt.organization_id::text, ''),
		       COALESCE(vt.team_id::text, ''),
		       COALESCE(vt.user_id::text, ''),
		       COALESCE(vt.virtual_user_id::text, ''),
		       COALESCE(vt.key_alias, ''),
		       COALESCE(vt.token_hash, ''),
		       COALESCE(vt.models, '{}'),
		       COALESCE(vt.spend, 0),
		       COALESCE(vt.expires, NULL),
		       COALESCE(vt.blocked, false),
		       COALESCE(vt.key_type, CASE WHEN vt.team_id IS NULL THEN 'pat' ELSE 'vat' END),
		       COALESCE(vt.metadata, '{}'::jsonb)
		FROM verification_tokens vt
		LEFT JOIN virtual_users vu
		  ON vu.virtual_user_id = vt.virtual_user_id
		 AND vu.tenant_id = vt.tenant_id
		WHERE vt.token_hash = $1
		  AND vt.revoked_at IS NULL
		  AND (vt.expires IS NULL OR vt.expires > NOW())
		  AND (vt.virtual_user_id IS NULL OR COALESCE(vu.active, false) = true)
		LIMIT 1
	`, hash).Scan(
		&auth.KeyID,
		&auth.TenantID,
		&auth.OrganizationID,
		&auth.TeamID,
		&auth.UserID,
		&auth.VirtualUserID,
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

	auth.ExpiresAt = expiresAt
	auth.VirtualModelSnapshots = parseVirtualModelSnapshots(rawMetadata)

	if err := s.applyRuntimeModelPolicy(ctx, &auth, keyModels); err != nil {
		return nil, err
	}

	return &auth, nil
}

func runtimePolicyByModel(runtimeModels []runtimeAllowedModelPolicy, model string) (runtimeAllowedModelPolicy, bool) {
	model = normalizeModelFQN(model)
	for _, item := range runtimeModels {
		if normalizeModelFQN(item.Model) == model {
			return item, true
		}
	}

	return runtimeAllowedModelPolicy{}, false
}

// ResolveTenantUserAuth builds an auth context for a tenant-bound user.
// This is used when tenant-admin calls control-plane with master key + x-godarai-user-id.
func (s *AuthRepository) ResolveTenantUserAuth(ctx context.Context, tenantID, userID string) (*platform.AuthContext, error) {
	tenantID = strings.TrimSpace(tenantID)

	userID = strings.TrimSpace(userID)
	if tenantID == "" || userID == "" {
		return nil, fmt.Errorf("tenantID and userID are required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, apperrors.ErrNotFound
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.ErrNotFound
	}

	tenantID = tenantUUID.String()

	userID = userUUID.String()
	if cached, ok := s.getAuthzCacheEntry(tenantID, userID); ok {
		return cached, nil
	}

	var (
		role   string
		active bool
	)

	queryErr := s.pool.QueryRow(ctx, `
		SELECT COALESCE(role, 'member'), COALESCE(active, true)
		FROM users
		WHERE tenant_id = $1::uuid AND user_id = $2::uuid
		LIMIT 1
	`, tenantID, userID).Scan(&role, &active)
	if queryErr != nil {
		return nil, queryErr
	}

	if !active || strings.EqualFold(role, "deactivated") {
		return nil, apperrors.ErrNotFound
	}

	resolvedRoles, resolvedPermissions, err := s.GetEffectivePermissionsForUser(ctx, tenantID, userID, role)
	if err != nil {
		return nil, err
	}

	effectiveRole := role
	if len(resolvedRoles) > 0 {
		effectiveRole = derivePrimaryRole(resolvedRoles)
	}

	resolved := &platform.AuthContext{
		TenantID:    tenantID,
		UserID:      userID,
		Role:        strings.TrimSpace(strings.ToLower(effectiveRole)),
		Permissions: resolvedPermissions,
	}
	s.setAuthzCacheEntry(tenantID, userID, *resolved)

	return resolved, nil
}

type runtimeAllowedModelPolicy struct {
	Model           string
	Types           []string
	PricingMode     string
	PricingResolved bool
}

func (s *AuthRepository) listRuntimeAllowedModelsByProviderAccount(ctx context.Context, tenantID, teamID, userID string) ([]runtimeAllowedModelPolicy, bool, error) {
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
		SELECT a.provider,
		       a.name,
		       m.model,
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
		 AND ac.role = 'user'
		WHERE a.tenant_id = $1
		  AND (
			   ac.principal_type = 'everyone'
			OR (
				ac.principal_type = 'team'
				AND ac.principal_id IS NOT NULL
				AND (
					ac.principal_id::text = NULLIF($2, '')
					OR (
						NULLIF($3, '') IS NOT NULL
						AND EXISTS (
							SELECT 1
							FROM team_memberships tm
							WHERE tm.tenant_id = $1::uuid
							  AND tm.team_id::text = ac.principal_id::text
							  AND tm.user_id::text = NULLIF($3, '')
						)
					)
				)
			)
			OR (ac.principal_type = 'user' AND ac.principal_id = NULLIF($3, ''))
		  )
	`, tenantID, teamID, userID)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	models := make([]runtimeAllowedModelPolicy, 0)

	for rows.Next() {
		var (
			item         runtimeAllowedModelPolicy
			provider     string
			accountAlias string
			storedModel  string
		)
		if err := rows.Scan(&provider, &accountAlias, &storedModel, &item.Types, &item.PricingMode, &item.PricingResolved); err != nil {
			return nil, false, err
		}

		item.Model = modelid.BuildModelFQN(provider, accountAlias, storedModel)
		if item.Model != "" {
			models = append(models, item)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, false, err
	}

	return models, true, nil
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
func (s *AuthRepository) ListAllowedModelFQNs(ctx context.Context, tenantID, userID, teamID string) ([]string, error) {
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
func (s *AuthRepository) CreateKey(ctx context.Context, params platform.CreateKeyParams) (*platform.CreatedKey, error) {
	params.Normalize()

	if params.KeyType == platform.KeyTypeVAT && params.VirtualUserID != "" {
		teamID, orgID, err := s.resolveVATVirtualUserScope(ctx, params.TenantID, params.VirtualUserID, params.TeamID, params.OrgID)
		if err != nil {
			return nil, err
		}

		params.TeamID = teamID
		params.OrgID = orgID
	}

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
			virtual_user_id, expires, spend, blocked, key_type, created_by_user_id, metadata, created_at, updated_at
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
			NULLIF($10, '')::uuid, $11, 0, false, $12, NULLIF($13, ''), $14, $15, $15
		)
	`, keyID, params.TenantID, tokenHash, params.KeyAlias, masked, params.Models,
		params.TeamID, params.OrgID, params.UserID, params.VirtualUserID, expiresAt, params.KeyType, params.UserID, rawMetadata, now)
	if err != nil {
		return nil, err
	}

	s.invalidateTenantRuntimeProfile(ctx, params.TenantID, "token.created")

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
		UserID:          params.UserID,
		VirtualUserID:   params.VirtualUserID,
		CreatedAt:       &now,
		CreatedByUserID: params.UserID,
		TokenHash:       tokenHash,
	}, nil
}

// ListKeys returns created keys for the tenant.
func (s *AuthRepository) ListKeys(ctx context.Context, tenantID string) ([]platform.CreatedKey, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, COALESCE(key_alias, ''), COALESCE(key_name, ''), tenant_id::text,
		       COALESCE(models, '{}'), COALESCE(team_id::text, ''), COALESCE(organization_id::text, ''),
		       COALESCE(virtual_user_id::text, ''),
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
		var (
			item   platform.CreatedKey
			models []string
		)
		if err := rows.Scan(&item.ID, &item.KeyAlias, &item.MaskedKey, &item.TenantID, &models,
			&item.TeamID, &item.OrgID, &item.VirtualUserID,
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

		keys = append(keys, item)
	}

	return keys, rows.Err()
}

func (s *AuthRepository) TouchKeyLastUsed(ctx context.Context, keyID string, usedAt time.Time) error {
	return s.queueKeyLastUsed(ctx, keyID, usedAt)
}

func (s *AuthRepository) PATExists(ctx context.Context, tenantID string, userID string, name string) (bool, error) {
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

func (s *AuthRepository) ListPATs(ctx context.Context, tenantID string, userID string, query string, limit int, offset int) ([]platform.CreatedKey, int, error) {
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
		var (
			item   platform.CreatedKey
			models []string
		)
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
		items = append(items, item)
	}

	return items, total, rows.Err()
}

func (s *AuthRepository) RevokePAT(ctx context.Context, tenantID string, userID string, keyID string, revokedByUserID string) (bool, error) {
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

	if tag.RowsAffected() > 0 {
		s.invalidateTenantRuntimeProfile(ctx, tenantID, "pat.revoked")
	}

	return tag.RowsAffected() > 0, nil
}

func (s *AuthRepository) RevokeAllPATs(ctx context.Context, tenantID string, userID string, revokedByUserID string) (int64, error) {
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

	if tag.RowsAffected() > 0 {
		s.invalidateTenantRuntimeProfile(ctx, tenantID, "pat.revoked_all")
	}

	return tag.RowsAffected(), nil
}

func (s *AuthRepository) ListVATTokens(ctx context.Context, tenantID string, virtualUserID string, query string, limit int, offset int) ([]platform.CreatedKey, int, error) {
	query = strings.TrimSpace(query)
	virtualUserID = strings.TrimSpace(virtualUserID)

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
		WHERE tenant_id = $1::uuid
		  AND virtual_user_id = $2::uuid
		  AND key_type = 'vat'
		  AND ($3 = '' OR LOWER(COALESCE(key_alias, '')) LIKE $4)
	`, tenantID, virtualUserID, query, like).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id::text, COALESCE(key_alias, ''), COALESCE(key_name, ''), tenant_id::text,
		       COALESCE(models, '{}'), COALESCE(team_id::text, ''), COALESCE(organization_id::text, ''),
		       COALESCE(virtual_user_id::text, ''), COALESCE(spend, 0), COALESCE(blocked, false),
		       COALESCE(expires, NULL), token_hash,
		       COALESCE(key_type, 'vat'), COALESCE(created_by_user_id, ''),
		       COALESCE(created_at, NULL), COALESCE(last_used_at, NULL), COALESCE(revoked_at, NULL)
		FROM verification_tokens
		WHERE tenant_id = $1::uuid
		  AND virtual_user_id = $2::uuid
		  AND key_type = 'vat'
		  AND ($3 = '' OR LOWER(COALESCE(key_alias, '')) LIKE $4)
		ORDER BY created_at DESC
		LIMIT $5 OFFSET $6
	`, tenantID, virtualUserID, query, like, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]platform.CreatedKey, 0)

	for rows.Next() {
		item, err := scanCreatedKey(rows)
		if err != nil {
			return nil, 0, err
		}

		item.Key = ""
		items = append(items, *item)
	}

	return items, total, rows.Err()
}

func (s *AuthRepository) UpdateVATToken(ctx context.Context, tenantID string, virtualUserID string, keyID string, req platform.UpdateVATTokenRequest) (*platform.CreatedKey, error) {
	keyAlias := strings.TrimSpace(req.KeyAlias)

	models := req.AllowedModels
	if len(models) == 0 {
		models = req.Models
	}

	modelsChanged := models != nil
	if models == nil {
		models = []string{}
	}

	var expiresAt *time.Time

	expiresChanged := strings.TrimSpace(req.ExpiresIn) != ""
	if expiresChanged {
		duration, err := parseDuration(strings.TrimSpace(req.ExpiresIn))
		if err != nil {
			return nil, err
		}

		exp := time.Now().UTC().Add(duration)
		expiresAt = &exp
	}

	rawMetadata := []byte("{}")

	metadataChanged := req.Metadata != nil
	if metadataChanged {
		encoded, err := json.Marshal(req.Metadata)
		if err != nil {
			return nil, err
		}

		rawMetadata = encoded
	}

	row := s.pool.QueryRow(ctx, `
		UPDATE verification_tokens
		SET key_alias = CASE WHEN $4 <> '' THEN $4 ELSE key_alias END,
		    models = CASE WHEN $5 THEN $6 ELSE models END,
		    expires = CASE WHEN $7 THEN $8 ELSE expires END,
		    blocked = COALESCE($9, blocked),
		    metadata = CASE WHEN $10 THEN $11::jsonb ELSE metadata END,
		    updated_at = NOW()
		WHERE id = $1::uuid
		  AND tenant_id = $2::uuid
		  AND virtual_user_id = $3::uuid
		  AND key_type = 'vat'
		RETURNING id::text, COALESCE(key_alias, ''), COALESCE(key_name, ''), tenant_id::text,
		       COALESCE(models, '{}'), COALESCE(team_id::text, ''), COALESCE(organization_id::text, ''),
		       COALESCE(virtual_user_id::text, ''), COALESCE(spend, 0), COALESCE(blocked, false),
		       COALESCE(expires, NULL), token_hash,
		       COALESCE(key_type, 'vat'), COALESCE(created_by_user_id, ''),
		       COALESCE(created_at, NULL), COALESCE(last_used_at, NULL), COALESCE(revoked_at, NULL)
	`, keyID, tenantID, virtualUserID, keyAlias, modelsChanged, models, expiresChanged, expiresAt, req.Blocked, metadataChanged, rawMetadata)

	item, err := scanCreatedKey(row)
	if err != nil {
		return nil, err
	}

	s.invalidateTenantRuntimeProfile(ctx, tenantID, "vat.updated")

	return item, nil
}

func (s *AuthRepository) RevokeVATToken(ctx context.Context, tenantID string, virtualUserID string, keyID string, revokedByUserID string, reason string) (bool, error) {
	tag, err := s.pool.Exec(ctx, `
		UPDATE verification_tokens
		SET revoked_at = NOW(),
		    revoked_by_user_id = NULLIF($4, ''),
		    revoked_reason = NULLIF($5, ''),
		    blocked = true,
		    updated_at = NOW()
		WHERE id = $1::uuid
		  AND tenant_id = $2::uuid
		  AND virtual_user_id = $3::uuid
		  AND key_type = 'vat'
		  AND revoked_at IS NULL
	`, keyID, tenantID, virtualUserID, strings.TrimSpace(revokedByUserID), strings.TrimSpace(reason))
	if err != nil {
		return false, err
	}

	if tag.RowsAffected() > 0 {
		s.invalidateTenantRuntimeProfile(ctx, tenantID, "vat.revoked")
	}

	return tag.RowsAffected() > 0, nil
}

func (s *AuthRepository) RotateVATToken(ctx context.Context, tenantID string, virtualUserID string, keyID string, rotatedByUserID string, gracePeriod string) (*platform.CreatedKey, error) {
	current, err := s.getVATToken(ctx, tenantID, virtualUserID, keyID)
	if err != nil {
		return nil, err
	}

	rawKey := "sk-" + strings.ReplaceAll(uuid.NewString(), "-", "")
	tokenHash := hashKey(rawKey)
	newID := uuid.NewString()
	masked := maskKey(rawKey)
	now := time.Now().UTC()
	row := s.pool.QueryRow(ctx, `
		INSERT INTO verification_tokens (
			id, tenant_id, token_hash, key_alias, key_name, models, team_id, organization_id,
			virtual_user_id, expires, spend, blocked, key_type, created_by_user_id,
			metadata, rotated_from_token_id, created_at, updated_at
		)
		SELECT $1, tenant_id, $2, key_alias, $3, models, team_id, organization_id,
		       virtual_user_id, expires, 0, false, 'vat', NULLIF($4, ''),
		       COALESCE(metadata, '{}'::jsonb), id, $5, $5
		FROM verification_tokens
		WHERE id = $6::uuid
		  AND tenant_id = $7::uuid
		  AND virtual_user_id = $8::uuid
		  AND key_type = 'vat'
		RETURNING id::text, COALESCE(key_alias, ''), COALESCE(key_name, ''), tenant_id::text,
		       COALESCE(models, '{}'), COALESCE(team_id::text, ''), COALESCE(organization_id::text, ''),
		       COALESCE(virtual_user_id::text, ''), COALESCE(spend, 0), COALESCE(blocked, false),
		       COALESCE(expires, NULL), token_hash,
		       COALESCE(key_type, 'vat'), COALESCE(created_by_user_id, ''),
		       COALESCE(created_at, NULL), COALESCE(last_used_at, NULL), COALESCE(revoked_at, NULL)
	`, newID, tokenHash, masked, strings.TrimSpace(rotatedByUserID), now, keyID, tenantID, virtualUserID)

	created, err := scanCreatedKey(row)
	if err != nil {
		return nil, err
	}

	s.invalidateTenantRuntimeProfile(ctx, tenantID, "vat.rotated")

	created.Key = rawKey
	created.MaskedKey = masked
	created.Role = platform.LegacyKeyRoleServiceAccount

	created.Models = append([]string(nil), current.Models...)
	created.AllowedModels = append([]string(nil), current.AllowedModels...)

	gracePeriod = strings.TrimSpace(gracePeriod)
	if gracePeriod == "" {
		_, err = s.RevokeVATToken(ctx, tenantID, virtualUserID, keyID, rotatedByUserID, "rotated")
		return created, err
	}

	graceDuration, err := parseDuration(gracePeriod)
	if err != nil {
		return nil, err
	}

	graceExpires := now.Add(graceDuration)

	_, err = s.pool.Exec(ctx, `
		UPDATE verification_tokens
		SET rotation_grace_expires_at = $4,
		    updated_at = NOW()
		WHERE id = $1::uuid
		  AND tenant_id = $2::uuid
		  AND virtual_user_id = $3::uuid
		  AND key_type = 'vat'
	`, keyID, tenantID, virtualUserID, graceExpires)
	if err == nil {
		s.invalidateTenantRuntimeProfile(ctx, tenantID, "vat.rotation_grace")
	}

	return created, err
}

func (s *AuthRepository) getVATToken(ctx context.Context, tenantID, virtualUserID, keyID string) (*platform.CreatedKey, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id::text, COALESCE(key_alias, ''), COALESCE(key_name, ''), tenant_id::text,
		       COALESCE(models, '{}'), COALESCE(team_id::text, ''), COALESCE(organization_id::text, ''),
		       COALESCE(virtual_user_id::text, ''), COALESCE(spend, 0), COALESCE(blocked, false),
		       COALESCE(expires, NULL), token_hash,
		       COALESCE(key_type, 'vat'), COALESCE(created_by_user_id, ''),
		       COALESCE(created_at, NULL), COALESCE(last_used_at, NULL), COALESCE(revoked_at, NULL)
		FROM verification_tokens
		WHERE id = $1::uuid
		  AND tenant_id = $2::uuid
		  AND virtual_user_id = $3::uuid
		  AND key_type = 'vat'
	`, keyID, tenantID, virtualUserID)

	return scanCreatedKey(row)
}

type createdKeyScanner interface {
	Scan(dest ...any) error
}

func scanCreatedKey(scanner createdKeyScanner) (*platform.CreatedKey, error) {
	var (
		item   platform.CreatedKey
		models []string
	)
	if err := scanner.Scan(&item.ID, &item.KeyAlias, &item.MaskedKey, &item.TenantID, &models,
		&item.TeamID, &item.OrgID, &item.VirtualUserID, &item.Spend, &item.Blocked,
		&item.ExpiresAt, &item.TokenHash, &item.KeyType, &item.CreatedByUserID,
		&item.CreatedAt, &item.LastUsedAt, &item.RevokedAt); err != nil {
		return nil, err
	}

	item.Models = models
	item.AllowedModels = append([]string(nil), models...)

	item.Role = deriveKeyRole(item.TeamID)
	if item.KeyType == "" {
		item.KeyType = deriveKeyType(item.TeamID)
	}

	return &item, nil
}

func (s *AuthRepository) resolveVATVirtualUserScope(ctx context.Context, tenantID, virtualUserID, requestedTeamID, requestedOrgID string) (string, string, error) {
	if _, err := uuid.Parse(strings.TrimSpace(virtualUserID)); err != nil {
		return "", "", fmt.Errorf("virtual_user_id must be a UUID")
	}

	var (
		teamID string
		active bool
	)
	if err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(team_id::text, ''), COALESCE(active, true)
		FROM virtual_users
		WHERE tenant_id = $1::uuid
		  AND virtual_user_id = $2::uuid
	`, strings.TrimSpace(tenantID), strings.TrimSpace(virtualUserID)).Scan(&teamID, &active); err != nil {
		return "", "", err
	}

	if !active {
		return "", "", fmt.Errorf("virtual user is inactive")
	}

	requestedTeamID = strings.TrimSpace(requestedTeamID)
	requestedOrgID = strings.TrimSpace(requestedOrgID)

	if teamID != "" && requestedTeamID != "" && requestedTeamID != teamID {
		return "", "", fmt.Errorf("team_id does not match virtual user scope")
	}

	if requestedTeamID != "" {
		teamID = requestedTeamID
	}

	return teamID, requestedOrgID, nil
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
