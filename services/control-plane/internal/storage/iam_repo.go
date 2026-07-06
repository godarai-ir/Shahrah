package storage

import (
	"context"
	"encoding/json"
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

	var (
		item          platform.OrgUnit
		metadataBytes []byte
	)

	err = s.pool.QueryRow(ctx, `
		INSERT INTO organizations (organization_id, tenant_id, organization_alias, metadata, created_at, updated_at)
		VALUES ($1, $2::uuid, $3, $4, NOW(), NOW())
		RETURNING organization_id::text, tenant_id::text, organization_alias, COALESCE(metadata, '{}'::jsonb), created_at, updated_at
	`, uuid.NewString(), tenantID, strings.TrimSpace(req.Alias), rawMetadata).Scan(
		&item.ID,
		&item.TenantID,
		&item.Alias,
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
	q, like, limit, offset := normalizePagedQuery(query, limit, offset)

	return listSearchResults(ctx, s.pool, `
		SELECT COUNT(1)
		FROM organizations
		WHERE tenant_id = $1::uuid
		  AND ($2 = '' OR LOWER(COALESCE(organization_alias, '')) LIKE $3)
	`, []any{tenantID, q, like}, `
		SELECT organization_id::text, tenant_id::text, organization_alias, COALESCE(metadata, '{}'::jsonb), created_at, updated_at
		FROM organizations
		WHERE tenant_id = $1::uuid
		  AND ($2 = '' OR LOWER(COALESCE(organization_alias, '')) LIKE $3)
		ORDER BY updated_at DESC, created_at DESC
		LIMIT $4 OFFSET $5
	`, []any{tenantID, q, like, limit, offset}, limit, scanOrgUnit)
}

func (s *Store) GetOrgUnit(ctx context.Context, tenantID string, orgUnitID string) (*platform.OrgUnit, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT organization_id::text, tenant_id::text, organization_alias, COALESCE(metadata, '{}'::jsonb), created_at, updated_at
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
		    metadata = $4,
		    updated_at = NOW()
		WHERE tenant_id = $1::uuid AND organization_id = $2::uuid
		RETURNING organization_id::text, tenant_id::text, organization_alias, COALESCE(metadata, '{}'::jsonb), created_at, updated_at
	`, tenantID, orgUnitID, alias, rawMetadata)
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

func (s *Store) CreateTeam(ctx context.Context, tenantID string, req platform.CreateTeamRequest) (*platform.Team, error) {
	metadata, err := mergeDisplayNameMetadata(req.Metadata, req.DisplayName)
	if err != nil {
		return nil, err
	}

	rawMetadata, err := marshalJSONObject(metadata)
	if err != nil {
		return nil, err
	}

	workspaceType := strings.TrimSpace(req.WorkspaceType)
	if workspaceType == "" {
		workspaceType = "general"
	}

	rows, err := s.pool.Query(ctx, `
		INSERT INTO teams (team_id, tenant_id, team_alias, org_unit_id, workspace_type, metadata, models, team_member_permissions, created_at, updated_at)
		VALUES ($1, $2::uuid, $3, NULLIF($4, '')::uuid, $5, $6, $7, $8, NOW(), NOW())
		RETURNING team_id::text, tenant_id::text, COALESCE(team_alias, ''), COALESCE(org_unit_id::text, ''), workspace_type,
		          COALESCE(metadata, '{}'::jsonb), COALESCE(models, '{}'), COALESCE(team_member_permissions, '{}'),
		          COALESCE(spend::text, ''), created_at, updated_at
	`, uuid.NewString(), tenantID, strings.TrimSpace(req.Alias), strings.TrimSpace(req.OrgUnitID), workspaceType, rawMetadata, req.AllowedModels, []string{})
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, pgx.ErrNoRows
	}

	return scanTeam(rows)
}

func (s *Store) ListTeams(ctx context.Context, tenantID string, query string, limit int, offset int) ([]platform.Team, int, error) {
	q, like, limit, offset := normalizePagedQuery(query, limit, offset)

	return listSearchResults(ctx, s.pool, `
		SELECT COUNT(1)
		FROM teams
		WHERE tenant_id = $1::uuid
		  AND ($2 = '' OR LOWER(COALESCE(team_alias, '')) LIKE $3 OR LOWER(COALESCE(metadata->>'display_name', '')) LIKE $3)
	`, []any{tenantID, q, like}, `
		SELECT team_id::text, tenant_id::text, COALESCE(team_alias, ''), COALESCE(org_unit_id::text, ''), workspace_type,
		       COALESCE(metadata, '{}'::jsonb), COALESCE(models, '{}'), COALESCE(team_member_permissions, '{}'),
		       COALESCE(spend::text, ''), created_at, updated_at
		FROM teams
		WHERE tenant_id = $1::uuid
		  AND ($2 = '' OR LOWER(COALESCE(team_alias, '')) LIKE $3 OR LOWER(COALESCE(metadata->>'display_name', '')) LIKE $3)
		ORDER BY updated_at DESC, created_at DESC
		LIMIT $4 OFFSET $5
	`, []any{tenantID, q, like, limit, offset}, limit, scanTeam)
}

func (s *Store) GetTeam(ctx context.Context, tenantID string, teamID string) (*platform.Team, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT team_id::text, tenant_id::text, COALESCE(team_alias, ''), COALESCE(org_unit_id::text, ''), workspace_type,
		       COALESCE(metadata, '{}'::jsonb), COALESCE(models, '{}'), COALESCE(team_member_permissions, '{}'),
		       COALESCE(spend::text, ''), created_at, updated_at
		FROM teams
		WHERE tenant_id = $1::uuid AND team_id = $2::uuid
		LIMIT 1
	`, tenantID, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, pgx.ErrNoRows
	}

	return scanTeam(rows)
}

func (s *Store) UpdateTeam(ctx context.Context, tenantID string, teamID string, req platform.UpdateTeamRequest) (*platform.Team, error) {
	current, err := s.GetTeam(ctx, tenantID, teamID)
	if err != nil {
		return nil, err
	}

	alias := current.Alias
	if trimmed := strings.TrimSpace(req.Alias); trimmed != "" {
		alias = trimmed
	}

	displayName := current.DisplayName
	if trimmed := strings.TrimSpace(req.DisplayName); trimmed != "" {
		displayName = trimmed
	}

	workspaceType := current.WorkspaceType
	if trimmed := strings.TrimSpace(req.WorkspaceType); trimmed != "" {
		workspaceType = trimmed
	}

	orgUnitID := current.OrgUnitID
	if req.OrgUnitID != "" {
		orgUnitID = strings.TrimSpace(req.OrgUnitID)
	}

	allowedModels := current.AllowedModels
	if req.AllowedModels != nil {
		allowedModels = req.AllowedModels
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
		UPDATE teams
		SET team_alias = $3,
		    org_unit_id = NULLIF($4, '')::uuid,
		    workspace_type = $5,
		    metadata = $6,
		    models = $7,
		    updated_at = NOW()
		WHERE tenant_id = $1::uuid AND team_id = $2::uuid
		RETURNING team_id::text, tenant_id::text, COALESCE(team_alias, ''), COALESCE(org_unit_id::text, ''), workspace_type,
		          COALESCE(metadata, '{}'::jsonb), COALESCE(models, '{}'), COALESCE(team_member_permissions, '{}'),
		          COALESCE(spend::text, ''), created_at, updated_at
	`, tenantID, teamID, alias, orgUnitID, workspaceType, rawMetadata, allowedModels)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, pgx.ErrNoRows
	}

	return scanTeam(rows)
}

func (s *Store) DeleteTeam(ctx context.Context, tenantID string, teamID string) error {
	command, err := s.pool.Exec(ctx, `
		DELETE FROM teams
		WHERE tenant_id = $1::uuid AND team_id = $2::uuid
	`, tenantID, teamID)
	if err != nil {
		return err
	}

	if command.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (s *Store) GetTeamPermissions(ctx context.Context, tenantID string, teamID string) (*platform.TeamPermissionsResponse, error) {
	var permissions []string
	if err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(team_member_permissions, '{}')
		FROM teams
		WHERE tenant_id = $1::uuid AND team_id = $2::uuid
		LIMIT 1
	`, tenantID, teamID).Scan(&permissions); err != nil {
		return nil, err
	}

	return &platform.TeamPermissionsResponse{
		TeamID:                teamID,
		TeamMemberPermissions: permissions,
	}, nil
}

func (s *Store) UpdateTeamPermissions(ctx context.Context, tenantID string, teamID string, req platform.UpdateTeamPermissionsRequest) (*platform.TeamPermissionsResponse, error) {
	var permissions []string
	if err := s.pool.QueryRow(ctx, `
		UPDATE teams
		SET team_member_permissions = $3,
		    updated_at = NOW()
		WHERE tenant_id = $1::uuid AND team_id = $2::uuid
		RETURNING COALESCE(team_member_permissions, '{}')
	`, tenantID, teamID, req.TeamMemberPermissions).Scan(&permissions); err != nil {
		return nil, err
	}

	return &platform.TeamPermissionsResponse{
		TeamID:                teamID,
		TeamMemberPermissions: permissions,
	}, nil
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

	item, err := scanVirtualUser(rows)
	if err == nil {
		s.invalidateTenantRuntimeProfile(ctx, tenantID, "virtual_user.created")
	}
	return item, err
}

func (s *Store) ListVirtualUsers(ctx context.Context, tenantID string, query string, limit int, offset int) ([]platform.VirtualUser, int, error) {
	q, like, limit, offset := normalizePagedQuery(query, limit, offset)

	return listSearchResults(ctx, s.pool, `
		SELECT COUNT(1)
		FROM virtual_users
		WHERE tenant_id = $1::uuid
		  AND ($2 = '' OR LOWER(COALESCE(alias, '')) LIKE $3 OR LOWER(COALESCE(display_name, '')) LIKE $3)
	`, []any{tenantID, q, like}, `
		SELECT virtual_user_id::text, tenant_id::text, COALESCE(team_id::text, ''), COALESCE(org_unit_id::text, ''), alias, COALESCE(display_name, ''), active, COALESCE(metadata, '{}'::jsonb), created_at, updated_at
		FROM virtual_users
		WHERE tenant_id = $1::uuid
		  AND ($2 = '' OR LOWER(COALESCE(alias, '')) LIKE $3 OR LOWER(COALESCE(display_name, '')) LIKE $3)
		ORDER BY updated_at DESC, created_at DESC
		LIMIT $4 OFFSET $5
	`, []any{tenantID, q, like, limit, offset}, limit, scanVirtualUser)
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

	item, err := scanVirtualUser(rows)
	if err == nil {
		s.invalidateTenantRuntimeProfile(ctx, tenantID, "virtual_user.updated")
	}
	return item, err
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
	s.invalidateTenantRuntimeProfile(ctx, tenantID, "virtual_user.deleted")

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

func (s *Store) ensurePlatformUserForMembership(ctx context.Context, tenantID, userID string) error {
	if _, err := uuid.Parse(strings.TrimSpace(userID)); err != nil {
		return err
	}

	var exists bool
	if err := s.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM users
			WHERE user_id = $2::uuid
			  AND tenant_id = $1::uuid
		)
	`, tenantID, userID).Scan(&exists); err != nil {
		return err
	}

	if !exists {
		return pgx.ErrNoRows
	}

	return nil
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
		       COALESCE(spend, 0.0), created_at, updated_at
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
		if err := rows.Scan(&item.ID, &item.TenantID, &item.TeamID, &item.UserID, &item.Role, &item.TeamMemberPermissions, &item.Spend, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, rows.Err()
}

func (s *Store) AddTeamMembership(ctx context.Context, tenantID string, teamID string, req platform.CreateTeamMembershipRequest) (*platform.TeamMembership, error) {
	rows, err := s.pool.Query(ctx, `
		INSERT INTO team_memberships (team_membership_id, tenant_id, team_id, user_id, role, team_member_permissions, created_at, updated_at)
		VALUES ($1, $2::uuid, $3::uuid, $4::uuid, $5, $6, NOW(), NOW())
		ON CONFLICT (team_id, user_id)
		DO UPDATE SET role = EXCLUDED.role, team_member_permissions = EXCLUDED.team_member_permissions, updated_at = NOW()
		RETURNING team_membership_id::text, tenant_id::text, team_id::text, user_id::text, role, COALESCE(team_member_permissions, '{}'),
		          COALESCE(spend, 0.0), created_at, updated_at
	`, uuid.NewString(), tenantID, teamID, req.UserID, req.Role, req.TeamMemberPermissions)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, pgx.ErrNoRows
	}

	var item platform.TeamMembership
	if err := rows.Scan(&item.ID, &item.TenantID, &item.TeamID, &item.UserID, &item.Role, &item.TeamMemberPermissions, &item.Spend, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, err
	}

	return &item, nil
}

func (s *Store) UpdateTeamMembership(ctx context.Context, tenantID string, teamID string, userID string, req platform.UpdateTeamMembershipRequest) (*platform.TeamMembership, error) {
	currentRows, err := s.pool.Query(ctx, `
		SELECT team_membership_id::text, tenant_id::text, team_id::text, user_id::text, role, COALESCE(team_member_permissions, '{}'),
		       COALESCE(spend, 0.0), created_at, updated_at
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

	scanErr := currentRows.Scan(&current.ID, &current.TenantID, &current.TeamID, &current.UserID, &current.Role, &current.TeamMemberPermissions, &current.Spend, &current.CreatedAt, &current.UpdatedAt)
	if scanErr != nil {
		return nil, scanErr
	}

	role := current.Role
	if trimmed := strings.TrimSpace(req.Role); trimmed != "" {
		role = trimmed
	}

	perms := current.TeamMemberPermissions
	if req.TeamMemberPermissions != nil {
		perms = req.TeamMemberPermissions
	}

	rows, err := s.pool.Query(ctx, `
		UPDATE team_memberships
		SET role = $4, team_member_permissions = $5, updated_at = NOW()
		WHERE tenant_id = $1::uuid AND team_id = $2::uuid AND user_id = $3::uuid
		RETURNING team_membership_id::text, tenant_id::text, team_id::text, user_id::text, role, COALESCE(team_member_permissions, '{}'),
		          COALESCE(spend, 0.0), created_at, updated_at
	`, tenantID, teamID, userID, role, perms)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, pgx.ErrNoRows
	}

	var item platform.TeamMembership
	if err := rows.Scan(&item.ID, &item.TenantID, &item.TeamID, &item.UserID, &item.Role, &item.TeamMemberPermissions, &item.Spend, &item.CreatedAt, &item.UpdatedAt); err != nil {
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
	var (
		item          platform.OrgUnit
		metadataBytes []byte
	)
	if err := rows.Scan(&item.ID, &item.TenantID, &item.Alias, &metadataBytes, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, err
	}

	item.Metadata = parseJSONObject(metadataBytes)
	item.DisplayName = extractDisplayName(item.Metadata, item.Alias)

	return &item, nil
}

func scanVirtualUser(rows pgx.Rows) (*platform.VirtualUser, error) {
	var (
		item          platform.VirtualUser
		displayName   string
		metadataBytes []byte
	)
	if err := rows.Scan(&item.ID, &item.TenantID, &item.TeamID, &item.OrgUnitID, &item.Alias, &displayName, &item.Active, &metadataBytes, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, err
	}

	item.DisplayName = strings.TrimSpace(displayName)
	item.Metadata = parseJSONObject(metadataBytes)

	return &item, nil
}

func scanTeam(rows pgx.Rows) (*platform.Team, error) {
	var (
		item                  platform.Team
		metadataBytes         []byte
		teamMemberPermissions []string
		spend                 string
	)
	if err := rows.Scan(
		&item.ID,
		&item.TenantID,
		&item.Alias,
		&item.OrgUnitID,
		&item.WorkspaceType,
		&metadataBytes,
		&item.AllowedModels,
		&teamMemberPermissions,
		&spend,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return nil, err
	}

	item.Metadata = parseJSONObject(metadataBytes)

	item.DisplayName = extractDisplayName(item.Metadata, item.Alias)
	if item.Metadata == nil {
		item.Metadata = map[string]interface{}{}
	}

	if spend != "" {
		item.Metadata["spend"] = spend
	}

	item.Metadata["team_member_permissions"] = teamMemberPermissions
	item.Metadata["allowed_models"] = item.AllowedModels

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
