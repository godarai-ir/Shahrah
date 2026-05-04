package storage

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Store) CreateOrgUnit(ctx context.Context, tenantID string, req platform.CreateOrgUnitRequest) (*platform.OrgUnit, error) {
	metadata, err := mergeDisplayNameMetadata(req.Metadata, req.DisplayName)
	if err != nil {
		return nil, err
	}
	rawMetadata, err := marshalJSONObject(metadata)
	if err != nil {
		return nil, err
	}

	var item platform.OrgUnit
	var metadataBytes []byte
	err = s.pool.QueryRow(ctx, `
		INSERT INTO organizations (organization_id, tenant_id, organization_alias, budget_id, metadata, created_at, updated_at)
		VALUES ($1, $2::uuid, $3, NULLIF($4, '')::uuid, $5, NOW(), NOW())
		RETURNING organization_id::text, tenant_id::text, organization_alias, COALESCE(budget_id::text, ''), COALESCE(metadata, '{}'::jsonb), created_at, updated_at
	`, uuid.NewString(), tenantID, strings.TrimSpace(req.Alias), strings.TrimSpace(req.BudgetID), rawMetadata).Scan(
		&item.ID,
		&item.TenantID,
		&item.Alias,
		&item.BudgetID,
		&metadataBytes,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	item.Metadata = parseJSONObject(metadataBytes)
	item.DisplayName = extractDisplayName(item.Metadata, item.Alias)
	return &item, nil
}

func (s *Store) ListOrgUnits(ctx context.Context, tenantID string, query string, limit int, offset int) ([]platform.OrgUnit, int, error) {
	query = strings.TrimSpace(strings.ToLower(query))
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}
	like := "%" + query + "%"

	var total int
	if err := s.pool.QueryRow(ctx, `
		SELECT COUNT(1)
		FROM organizations
		WHERE tenant_id = $1::uuid
		  AND ($2 = '' OR LOWER(COALESCE(organization_alias, '')) LIKE $3)
	`, tenantID, query, like).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := s.pool.Query(ctx, `
		SELECT organization_id::text, tenant_id::text, organization_alias, COALESCE(budget_id::text, ''), COALESCE(metadata, '{}'::jsonb), created_at, updated_at
		FROM organizations
		WHERE tenant_id = $1::uuid
		  AND ($2 = '' OR LOWER(COALESCE(organization_alias, '')) LIKE $3)
		ORDER BY updated_at DESC, created_at DESC
		LIMIT $4 OFFSET $5
	`, tenantID, query, like, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]platform.OrgUnit, 0, limit)
	for rows.Next() {
		item, err := scanOrgUnit(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, *item)
	}
	return items, total, rows.Err()
}

func (s *Store) GetOrgUnit(ctx context.Context, tenantID string, orgUnitID string) (*platform.OrgUnit, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT organization_id::text, tenant_id::text, organization_alias, COALESCE(budget_id::text, ''), COALESCE(metadata, '{}'::jsonb), created_at, updated_at
		FROM organizations
		WHERE tenant_id = $1::uuid AND organization_id = $2::uuid
		LIMIT 1
	`, tenantID, orgUnitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, pgx.ErrNoRows
	}
	return scanOrgUnit(rows)
}

func (s *Store) UpdateOrgUnit(ctx context.Context, tenantID string, orgUnitID string, req platform.UpdateOrgUnitRequest) (*platform.OrgUnit, error) {
	current, err := s.GetOrgUnit(ctx, tenantID, orgUnitID)
	if err != nil {
		return nil, err
	}

	alias := current.Alias
	if trimmed := strings.TrimSpace(req.Alias); trimmed != "" {
		alias = trimmed
	}
	displayName := current.DisplayName
	if strings.TrimSpace(req.DisplayName) != "" {
		displayName = strings.TrimSpace(req.DisplayName)
	}
	budgetID := current.BudgetID
	if req.BudgetID != "" {
		budgetID = strings.TrimSpace(req.BudgetID)
	}
	metadata := current.Metadata
	if req.Metadata != nil {
		metadata = req.Metadata
	}
	mergedMetadata, err := mergeDisplayNameMetadata(metadata, displayName)
	if err != nil {
		return nil, err
	}
	rawMetadata, err := marshalJSONObject(mergedMetadata)
	if err != nil {
		return nil, err
	}

	rows, err := s.pool.Query(ctx, `
		UPDATE organizations
		SET organization_alias = $3,
		    budget_id = NULLIF($4, '')::uuid,
		    metadata = $5,
		    updated_at = NOW()
		WHERE tenant_id = $1::uuid AND organization_id = $2::uuid
		RETURNING organization_id::text, tenant_id::text, organization_alias, COALESCE(budget_id::text, ''), COALESCE(metadata, '{}'::jsonb), created_at, updated_at
	`, tenantID, orgUnitID, alias, budgetID, rawMetadata)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, pgx.ErrNoRows
	}
	return scanOrgUnit(rows)
}

func (s *Store) DeleteOrgUnit(ctx context.Context, tenantID string, orgUnitID string) error {
	command, err := s.pool.Exec(ctx, `
		DELETE FROM organizations
		WHERE tenant_id = $1::uuid AND organization_id = $2::uuid
	`, tenantID, orgUnitID)
	if err != nil {
		return err
	}
	if command.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) CreateVirtualUser(ctx context.Context, tenantID string, req platform.CreateVirtualUserRequest) (*platform.VirtualUser, error) {
	rawMetadata, err := marshalJSONObject(req.Metadata)
	if err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `
		INSERT INTO virtual_users (virtual_user_id, tenant_id, team_id, org_unit_id, alias, display_name, metadata, active, created_at, updated_at)
		VALUES ($1, $2::uuid, NULLIF($3, '')::uuid, NULLIF($4, '')::uuid, $5, NULLIF($6, ''), $7, true, NOW(), NOW())
		RETURNING virtual_user_id::text, tenant_id::text, COALESCE(team_id::text, ''), COALESCE(org_unit_id::text, ''), alias, COALESCE(display_name, ''), active, COALESCE(metadata, '{}'::jsonb), created_at, updated_at
	`, uuid.NewString(), tenantID, strings.TrimSpace(req.TeamID), strings.TrimSpace(req.OrgUnitID), strings.TrimSpace(req.Alias), strings.TrimSpace(req.DisplayName), rawMetadata)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, pgx.ErrNoRows
	}
	return scanVirtualUser(rows)
}

func (s *Store) ListVirtualUsers(ctx context.Context, tenantID string, query string, limit int, offset int) ([]platform.VirtualUser, int, error) {
	query = strings.TrimSpace(strings.ToLower(query))
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}
	like := "%" + query + "%"

	var total int
	if err := s.pool.QueryRow(ctx, `
		SELECT COUNT(1)
		FROM virtual_users
		WHERE tenant_id = $1::uuid
		  AND ($2 = '' OR LOWER(COALESCE(alias, '')) LIKE $3 OR LOWER(COALESCE(display_name, '')) LIKE $3)
	`, tenantID, query, like).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := s.pool.Query(ctx, `
		SELECT virtual_user_id::text, tenant_id::text, COALESCE(team_id::text, ''), COALESCE(org_unit_id::text, ''), alias, COALESCE(display_name, ''), active, COALESCE(metadata, '{}'::jsonb), created_at, updated_at
		FROM virtual_users
		WHERE tenant_id = $1::uuid
		  AND ($2 = '' OR LOWER(COALESCE(alias, '')) LIKE $3 OR LOWER(COALESCE(display_name, '')) LIKE $3)
		ORDER BY updated_at DESC, created_at DESC
		LIMIT $4 OFFSET $5
	`, tenantID, query, like, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]platform.VirtualUser, 0, limit)
	for rows.Next() {
		item, err := scanVirtualUser(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, *item)
	}
	return items, total, rows.Err()
}

func (s *Store) GetVirtualUser(ctx context.Context, tenantID string, virtualUserID string) (*platform.VirtualUser, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT virtual_user_id::text, tenant_id::text, COALESCE(team_id::text, ''), COALESCE(org_unit_id::text, ''), alias, COALESCE(display_name, ''), active, COALESCE(metadata, '{}'::jsonb), created_at, updated_at
		FROM virtual_users
		WHERE tenant_id = $1::uuid AND virtual_user_id = $2::uuid
		LIMIT 1
	`, tenantID, virtualUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, pgx.ErrNoRows
	}
	return scanVirtualUser(rows)
}

func (s *Store) UpdateVirtualUser(ctx context.Context, tenantID string, virtualUserID string, req platform.UpdateVirtualUserRequest) (*platform.VirtualUser, error) {
	current, err := s.GetVirtualUser(ctx, tenantID, virtualUserID)
	if err != nil {
		return nil, err
	}
	alias := current.Alias
	if trimmed := strings.TrimSpace(req.Alias); trimmed != "" {
		alias = trimmed
	}
	displayName := current.DisplayName
	if req.DisplayName != "" {
		displayName = strings.TrimSpace(req.DisplayName)
	}
	teamID := current.TeamID
	if req.TeamID != "" {
		teamID = strings.TrimSpace(req.TeamID)
	}
	orgUnitID := current.OrgUnitID
	if req.OrgUnitID != "" {
		orgUnitID = strings.TrimSpace(req.OrgUnitID)
	}
	active := current.Active
	if req.Active != nil {
		active = *req.Active
	}
	metadata := current.Metadata
	if req.Metadata != nil {
		metadata = req.Metadata
	}
	rawMetadata, err := marshalJSONObject(metadata)
	if err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `
		UPDATE virtual_users
		SET alias = $3,
		    display_name = NULLIF($4, ''),
		    team_id = NULLIF($5, '')::uuid,
		    org_unit_id = NULLIF($6, '')::uuid,
		    active = $7,
		    metadata = $8,
		    updated_at = NOW()
		WHERE tenant_id = $1::uuid AND virtual_user_id = $2::uuid
		RETURNING virtual_user_id::text, tenant_id::text, COALESCE(team_id::text, ''), COALESCE(org_unit_id::text, ''), alias, COALESCE(display_name, ''), active, COALESCE(metadata, '{}'::jsonb), created_at, updated_at
	`, tenantID, virtualUserID, alias, displayName, teamID, orgUnitID, active, rawMetadata)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, pgx.ErrNoRows
	}
	return scanVirtualUser(rows)
}

func (s *Store) DeleteVirtualUser(ctx context.Context, tenantID string, virtualUserID string) error {
	command, err := s.pool.Exec(ctx, `
		DELETE FROM virtual_users
		WHERE tenant_id = $1::uuid AND virtual_user_id = $2::uuid
	`, tenantID, virtualUserID)
	if err != nil {
		return err
	}
	if command.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) ListOrgUnitMemberships(ctx context.Context, tenantID string, orgUnitID string) ([]platform.OrgUnitMembership, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, tenant_id::text, organization_id::text, user_id::text, COALESCE(role, 'member'), created_at, updated_at
		FROM organization_memberships
		WHERE tenant_id = $1::uuid AND organization_id = $2::uuid
		ORDER BY updated_at DESC, created_at DESC
	`, tenantID, orgUnitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]platform.OrgUnitMembership, 0)
	for rows.Next() {
		var item platform.OrgUnitMembership
		if err := rows.Scan(&item.ID, &item.TenantID, &item.OrgUnitID, &item.UserID, &item.Role, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) AddOrgUnitMembership(ctx context.Context, tenantID string, orgUnitID string, req platform.CreateOrgUnitMembershipRequest) (*platform.OrgUnitMembership, error) {
	if err := s.ensurePlatformUserForMembership(ctx, tenantID, req.UserID); err != nil {
		return nil, err
	}

	rows, err := s.pool.Query(ctx, `
		INSERT INTO organization_memberships (id, tenant_id, organization_id, user_id, role, created_at, updated_at)
		VALUES ($1, $2::uuid, $3::uuid, $4::uuid, $5, NOW(), NOW())
		ON CONFLICT (organization_id, user_id)
		DO UPDATE SET role = EXCLUDED.role, tenant_id = EXCLUDED.tenant_id, updated_at = NOW()
		RETURNING id::text, tenant_id::text, organization_id::text, user_id::text, COALESCE(role, 'member'), created_at, updated_at
	`, uuid.NewString(), tenantID, orgUnitID, req.UserID, req.Role)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, pgx.ErrNoRows
	}
	var item platform.OrgUnitMembership
	if err := rows.Scan(&item.ID, &item.TenantID, &item.OrgUnitID, &item.UserID, &item.Role, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *Store) ensurePlatformUserForMembership(ctx context.Context, tenantID string, userID string) error {
	if _, err := uuid.Parse(strings.TrimSpace(userID)); err != nil {
		return err
	}

	command, err := s.pool.Exec(ctx, `
		UPDATE users
		SET tenant_id = COALESCE(tenant_id, $1::uuid),
		    updated_at = NOW()
		WHERE user_id = $2::uuid
		  AND (tenant_id IS NULL OR tenant_id = $1::uuid)
	`, tenantID, userID)
	if err != nil {
		return err
	}
	if command.RowsAffected() > 0 {
		return nil
	}

	var tenantSlug string
	if err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(slug, '')
		FROM tenants
		WHERE tenant_id = $1::uuid
		LIMIT 1
	`, tenantID).Scan(&tenantSlug); err != nil {
		return err
	}

	var email string
	var mobile string
	var role string
	if err := s.pool.QueryRow(ctx, `
		SELECT COALESCE("email", ''), COALESCE("mobile", ''), COALESCE("role", 'tenant_member')
		FROM "tenant_admin_users"
		WHERE "id" = $1 AND "tenantSlug" = $2
		LIMIT 1
	`, userID, tenantSlug).Scan(&email, &mobile, &role); err != nil {
		return err
	}

	emailForInsert := strings.TrimSpace(strings.ToLower(email))
	if emailForInsert != "" {
		var existingEmailUserID string
		err := s.pool.QueryRow(ctx, `
			SELECT user_id::text
			FROM users
			WHERE LOWER(COALESCE(email, '')) = LOWER($1)
			LIMIT 1
		`, emailForInsert).Scan(&existingEmailUserID)
		if err == nil && existingEmailUserID != userID {
			emailForInsert = ""
		} else if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
	}

	alias := emailForInsert
	if alias == "" {
		alias = strings.TrimSpace(mobile)
	}
	if alias == "" {
		alias = userID
	}

	_, err = s.pool.Exec(ctx, `
		INSERT INTO users (user_id, tenant_id, user_alias, email, mobile, role, active, metadata, created_at, updated_at)
		VALUES ($1::uuid, $2::uuid, $3, NULLIF($4, ''), NULLIF($5, ''), $6, true, '{}'::jsonb, NOW(), NOW())
		ON CONFLICT (user_id) DO UPDATE
		SET tenant_id = EXCLUDED.tenant_id,
		    user_alias = COALESCE(NULLIF(users.user_alias, ''), EXCLUDED.user_alias),
		    updated_at = NOW()
	`, userID, tenantID, alias, emailForInsert, strings.TrimSpace(mobile), normalizePlatformUserRole(role))
	return err
}

func (s *Store) UpdateOrgUnitMembership(ctx context.Context, tenantID string, orgUnitID string, userID string, req platform.UpdateOrgUnitMembershipRequest) (*platform.OrgUnitMembership, error) {
	rows, err := s.pool.Query(ctx, `
		UPDATE organization_memberships
		SET role = $4, updated_at = NOW()
		WHERE tenant_id = $1::uuid AND organization_id = $2::uuid AND user_id = $3::uuid
		RETURNING id::text, tenant_id::text, organization_id::text, user_id::text, COALESCE(role, 'member'), created_at, updated_at
	`, tenantID, orgUnitID, userID, req.Role)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, pgx.ErrNoRows
	}
	var item platform.OrgUnitMembership
	if err := rows.Scan(&item.ID, &item.TenantID, &item.OrgUnitID, &item.UserID, &item.Role, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *Store) DeleteOrgUnitMembership(ctx context.Context, tenantID string, orgUnitID string, userID string) error {
	command, err := s.pool.Exec(ctx, `
		DELETE FROM organization_memberships
		WHERE tenant_id = $1::uuid AND organization_id = $2::uuid AND user_id = $3::uuid
	`, tenantID, orgUnitID, userID)
	if err != nil {
		return err
	}
	if command.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) ListTeamMemberships(ctx context.Context, tenantID string, teamID string) ([]platform.TeamMembership, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT team_membership_id::text, tenant_id::text, team_id::text, user_id::text, role, COALESCE(team_member_permissions, '{}'),
		       COALESCE(budget_id::text, ''), COALESCE(spend, 0.0), created_at, updated_at
		FROM team_memberships
		WHERE tenant_id = $1::uuid AND team_id = $2::uuid
		ORDER BY updated_at DESC, created_at DESC
	`, tenantID, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]platform.TeamMembership, 0)
	for rows.Next() {
		var item platform.TeamMembership
		if err := rows.Scan(&item.ID, &item.TenantID, &item.TeamID, &item.UserID, &item.Role, &item.TeamMemberPermissions, &item.BudgetID, &item.Spend, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) AddTeamMembership(ctx context.Context, tenantID string, teamID string, req platform.CreateTeamMembershipRequest) (*platform.TeamMembership, error) {
	rows, err := s.pool.Query(ctx, `
		INSERT INTO team_memberships (team_membership_id, tenant_id, team_id, user_id, role, team_member_permissions, budget_id, created_at, updated_at)
		VALUES ($1, $2::uuid, $3::uuid, $4::uuid, $5, $6, NULLIF($7, '')::uuid, NOW(), NOW())
		ON CONFLICT (team_id, user_id)
		DO UPDATE SET role = EXCLUDED.role, team_member_permissions = EXCLUDED.team_member_permissions, budget_id = EXCLUDED.budget_id, updated_at = NOW()
		RETURNING team_membership_id::text, tenant_id::text, team_id::text, user_id::text, role, COALESCE(team_member_permissions, '{}'),
		          COALESCE(budget_id::text, ''), COALESCE(spend, 0.0), created_at, updated_at
	`, uuid.NewString(), tenantID, teamID, req.UserID, req.Role, req.TeamMemberPermissions, strings.TrimSpace(req.BudgetID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, pgx.ErrNoRows
	}
	var item platform.TeamMembership
	if err := rows.Scan(&item.ID, &item.TenantID, &item.TeamID, &item.UserID, &item.Role, &item.TeamMemberPermissions, &item.BudgetID, &item.Spend, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *Store) UpdateTeamMembership(ctx context.Context, tenantID string, teamID string, userID string, req platform.UpdateTeamMembershipRequest) (*platform.TeamMembership, error) {
	currentRows, err := s.pool.Query(ctx, `
		SELECT team_membership_id::text, tenant_id::text, team_id::text, user_id::text, role, COALESCE(team_member_permissions, '{}'),
		       COALESCE(budget_id::text, ''), COALESCE(spend, 0.0), created_at, updated_at
		FROM team_memberships
		WHERE tenant_id = $1::uuid AND team_id = $2::uuid AND user_id = $3::uuid
		LIMIT 1
	`, tenantID, teamID, userID)
	if err != nil {
		return nil, err
	}
	defer currentRows.Close()
	if !currentRows.Next() {
		return nil, pgx.ErrNoRows
	}
	var current platform.TeamMembership
	if err := currentRows.Scan(&current.ID, &current.TenantID, &current.TeamID, &current.UserID, &current.Role, &current.TeamMemberPermissions, &current.BudgetID, &current.Spend, &current.CreatedAt, &current.UpdatedAt); err != nil {
		return nil, err
	}
	role := current.Role
	if trimmed := strings.TrimSpace(req.Role); trimmed != "" {
		role = trimmed
	}
	perms := current.TeamMemberPermissions
	if req.TeamMemberPermissions != nil {
		perms = req.TeamMemberPermissions
	}
	budgetID := current.BudgetID
	if req.BudgetID != "" {
		budgetID = strings.TrimSpace(req.BudgetID)
	}
	rows, err := s.pool.Query(ctx, `
		UPDATE team_memberships
		SET role = $4, team_member_permissions = $5, budget_id = NULLIF($6, '')::uuid, updated_at = NOW()
		WHERE tenant_id = $1::uuid AND team_id = $2::uuid AND user_id = $3::uuid
		RETURNING team_membership_id::text, tenant_id::text, team_id::text, user_id::text, role, COALESCE(team_member_permissions, '{}'),
		          COALESCE(budget_id::text, ''), COALESCE(spend, 0.0), created_at, updated_at
	`, tenantID, teamID, userID, role, perms, budgetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, pgx.ErrNoRows
	}
	var item platform.TeamMembership
	if err := rows.Scan(&item.ID, &item.TenantID, &item.TeamID, &item.UserID, &item.Role, &item.TeamMemberPermissions, &item.BudgetID, &item.Spend, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *Store) DeleteTeamMembership(ctx context.Context, tenantID string, teamID string, userID string) error {
	command, err := s.pool.Exec(ctx, `
		DELETE FROM team_memberships
		WHERE tenant_id = $1::uuid AND team_id = $2::uuid AND user_id = $3::uuid
	`, tenantID, teamID, userID)
	if err != nil {
		return err
	}
	if command.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func scanOrgUnit(rows pgx.Rows) (*platform.OrgUnit, error) {
	var item platform.OrgUnit
	var metadataBytes []byte
	if err := rows.Scan(&item.ID, &item.TenantID, &item.Alias, &item.BudgetID, &metadataBytes, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, err
	}
	item.Metadata = parseJSONObject(metadataBytes)
	item.DisplayName = extractDisplayName(item.Metadata, item.Alias)
	return &item, nil
}

func scanVirtualUser(rows pgx.Rows) (*platform.VirtualUser, error) {
	var item platform.VirtualUser
	var displayName string
	var metadataBytes []byte
	if err := rows.Scan(&item.ID, &item.TenantID, &item.TeamID, &item.OrgUnitID, &item.Alias, &displayName, &item.Active, &metadataBytes, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, err
	}
	item.DisplayName = strings.TrimSpace(displayName)
	item.Metadata = parseJSONObject(metadataBytes)
	return &item, nil
}

func mergeDisplayNameMetadata(metadata map[string]interface{}, displayName string) (map[string]interface{}, error) {
	cloned := make(map[string]interface{})
	if metadata != nil {
		raw, err := json.Marshal(metadata)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(raw, &cloned); err != nil {
			return nil, err
		}
	}
	if trimmed := strings.TrimSpace(displayName); trimmed != "" {
		cloned["display_name"] = trimmed
	}
	return cloned, nil
}

func extractDisplayName(metadata map[string]interface{}, fallback string) string {
	if metadata != nil {
		if value, ok := metadata["display_name"]; ok {
			if name := strings.TrimSpace(toString(value)); name != "" {
				return name
			}
		}
	}
	return strings.TrimSpace(fallback)
}

func normalizePlatformUserRole(role string) string {
	switch strings.TrimSpace(strings.ToLower(role)) {
	case "tenant_admin", "proxy_admin", "org_admin":
		return "org_admin"
	case "team_admin":
		return "team_admin"
	default:
		return "internal_user"
	}
}
