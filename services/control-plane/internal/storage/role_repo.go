package storage

import (
	"context"
	"errors"
	"regexp"
	"sort"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/iam"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrRoleNotFound          = iam.ErrRoleNotFound
	ErrRoleImmutable         = iam.ErrRoleImmutable
	ErrInvalidRoleName       = iam.ErrInvalidRoleName
	ErrInvalidPermission     = iam.ErrInvalidPermission
	ErrInvalidResourceType   = iam.ErrInvalidResourceType
	ErrInvalidUserRoleUpdate = iam.ErrInvalidUserRoleUpdate
	ErrUserNotFound          = iam.ErrUserNotFound
)

var roleNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{1,63}$`)

var systemRolePermissions = map[string][]string{
	"proxy_admin": {
		"policy:list", "policy:manage",
		"role:list", "role:manage",
		"user:list", "user:manage",
		"team:create", "team:read", "team:manage", "team:delete",
		"org_unit:create", "org_unit:read", "org_unit:manage", "org_unit:delete", "org_unit:list",
		"virtual_account:create", "virtual_account:read", "virtual_account:manage", "virtual_account:delete",
		"provider_account:create", "provider_account:read", "provider_account:manage", "provider_account:delete",
		"settings:list", "settings:manage",
		"billing:read", "billing:manage", "billing:profile_manage", "billing:documents_read", "billing:codes_redeem",
		"logs:read_metadata", "logs:read_content", "logs:export", "logs:manage_views", "logs:replay",
		"support_ticket:create", "support_ticket:read_own", "support_ticket:read_all", "support_ticket:reply", "support_ticket:rate",
	},
	"tenant_admin": {
		"policy:list", "policy:manage",
		"role:list", "role:manage",
		"user:list", "user:manage",
		"team:create", "team:read", "team:manage", "team:delete",
		"org_unit:create", "org_unit:read", "org_unit:manage", "org_unit:delete", "org_unit:list",
		"virtual_account:create", "virtual_account:read", "virtual_account:manage", "virtual_account:delete",
		"provider_account:create", "provider_account:read", "provider_account:manage", "provider_account:delete",
		"settings:list", "settings:manage",
		"billing:read", "billing:manage", "billing:profile_manage", "billing:documents_read", "billing:codes_redeem",
		"logs:read_metadata", "logs:read_content", "logs:export", "logs:manage_views", "logs:replay",
		"support_ticket:create", "support_ticket:read_own", "support_ticket:read_all", "support_ticket:reply", "support_ticket:rate",
	},
	"billing_manager": {
		"billing:read", "billing:manage", "billing:profile_manage", "billing:documents_read", "billing:codes_redeem",
		"settings:list",
	},
	"org_admin": {
		"policy:list",
		"role:list",
		"user:list", "user:manage",
		"org_unit:read", "org_unit:manage", "org_unit:list",
		"team:create", "team:read", "team:manage",
		"virtual_account:create", "virtual_account:read", "virtual_account:manage",
		"provider_account:read",
		"settings:list",
		"logs:read_metadata", "logs:read_content", "logs:export", "logs:manage_views", "logs:replay",
		"support_ticket:create", "support_ticket:read_own", "support_ticket:read_all", "support_ticket:reply", "support_ticket:rate",
	},
	"team_admin": {
		"policy:list",
		"role:list",
		"user:list",
		"team:read", "team:manage",
		"virtual_account:create", "virtual_account:read", "virtual_account:manage", "virtual_account:delete",
		"provider_account:read",
		"settings:list",
		"logs:read_metadata",
		"support_ticket:create", "support_ticket:read_own", "support_ticket:reply", "support_ticket:rate",
	},
	"member": {
		"user:list",
		"settings:list",
		"logs:read_metadata",
		"support_ticket:create", "support_ticket:read_own", "support_ticket:reply", "support_ticket:rate",
	},
	"readonly_member": {
		"settings:list",
		"logs:read_metadata",
		"support_ticket:create", "support_ticket:read_own", "support_ticket:reply", "support_ticket:rate",
	},
}

func (s *AuthRepository) ListRoles(ctx context.Context, tenantID string, query string, limit int, offset int) ([]platform.RoleDefinition, int, error) {
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

	if err := s.ensureSystemRoles(ctx, tenantID); err != nil {
		return nil, 0, err
	}

	like := "%" + query + "%"

	var total int
	if err := s.pool.QueryRow(ctx, `
		SELECT COUNT(1)
		FROM roles r
		WHERE r.tenant_id = $1::uuid
		  AND ($2 = '' OR LOWER(r.name) LIKE $3 OR LOWER(COALESCE(r.display_name, '')) LIKE $3)
	`, tenantID, query, like).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := s.pool.Query(ctx, `
		SELECT r.role_id::text,
		       r.tenant_id::text,
		       r.name,
		       COALESCE(r.display_name, ''),
		       COALESCE(r.description, ''),
		       COALESCE(r.is_system, false),
		       COALESCE(ARRAY_AGG(rp.permission ORDER BY rp.permission) FILTER (WHERE rp.permission IS NOT NULL), '{}') AS permissions,
		       r.created_at,
		       r.updated_at
		FROM roles r
		LEFT JOIN role_permissions rp ON rp.role_id = r.role_id
		WHERE r.tenant_id = $1::uuid
		  AND ($2 = '' OR LOWER(r.name) LIKE $3 OR LOWER(COALESCE(r.display_name, '')) LIKE $3)
		GROUP BY r.role_id
		ORDER BY r.is_system DESC, r.updated_at DESC, r.created_at DESC
		LIMIT $4 OFFSET $5
	`, tenantID, query, like, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]platform.RoleDefinition, 0, limit)

	for rows.Next() {
		var item platform.RoleDefinition
		if err := rows.Scan(
			&item.ID,
			&item.TenantID,
			&item.Name,
			&item.DisplayName,
			&item.Description,
			&item.IsSystem,
			&item.Permissions,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}

		items = append(items, item)
	}

	return items, total, rows.Err()
}

func (s *AuthRepository) CreateRole(ctx context.Context, tenantID string, req platform.CreateRoleRequest) (*platform.RoleDefinition, error) {
	if err := s.ensureSystemRoles(ctx, tenantID); err != nil {
		return nil, err
	}

	name := strings.TrimSpace(strings.ToLower(req.Name))
	if !roleNamePattern.MatchString(name) {
		return nil, ErrInvalidRoleName
	}

	displayName := strings.TrimSpace(req.DisplayName)
	if displayName == "" {
		displayName = name
	}

	permissions, err := normalizePermissions(req.Permissions)
	if err != nil {
		return nil, err
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	roleID := uuid.NewString()

	var item platform.RoleDefinition
	if err := tx.QueryRow(ctx, `
		INSERT INTO roles (role_id, tenant_id, name, display_name, description, is_system, created_at, updated_at)
		VALUES ($1::uuid, $2::uuid, $3, $4, NULLIF($5, ''), false, NOW(), NOW())
		RETURNING role_id::text, tenant_id::text, name, display_name, COALESCE(description, ''), is_system, created_at, updated_at
	`, roleID, tenantID, name, displayName, strings.TrimSpace(req.Description)).Scan(
		&item.ID,
		&item.TenantID,
		&item.Name,
		&item.DisplayName,
		&item.Description,
		&item.IsSystem,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrInvalidRoleName
		}

		return nil, err
	}

	for _, permission := range permissions {
		if _, err := tx.Exec(ctx, `
			INSERT INTO role_permissions (role_permission_id, tenant_id, role_id, permission, created_at, updated_at)
			VALUES ($1::uuid, $2::uuid, $3::uuid, $4, NOW(), NOW())
		`, uuid.NewString(), tenantID, roleID, permission); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	s.bumpAuthzTenantEpoch(tenantID)

	item.Permissions = permissions

	return &item, nil
}

func (s *AuthRepository) UpdateRole(ctx context.Context, tenantID string, roleID string, req platform.UpdateRoleRequest) (*platform.RoleDefinition, error) {
	if err := s.ensureSystemRoles(ctx, tenantID); err != nil {
		return nil, err
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	var current platform.RoleDefinition

	queryErr := tx.QueryRow(ctx, `
		SELECT role_id::text, tenant_id::text, name, COALESCE(display_name, ''), COALESCE(description, ''), COALESCE(is_system, false), created_at, updated_at
		FROM roles
		WHERE tenant_id = $1::uuid AND role_id = $2::uuid
		LIMIT 1
		FOR UPDATE
	`, tenantID, roleID).Scan(
		&current.ID,
		&current.TenantID,
		&current.Name,
		&current.DisplayName,
		&current.Description,
		&current.IsSystem,
		&current.CreatedAt,
		&current.UpdatedAt,
	)
	if queryErr != nil {
		if errors.Is(queryErr, pgx.ErrNoRows) {
			return nil, ErrRoleNotFound
		}

		return nil, queryErr
	}

	if current.IsSystem && (current.Name == "proxy_admin" || current.Name == "tenant_admin") {
		return nil, ErrRoleImmutable
	}

	displayName := strings.TrimSpace(req.DisplayName)
	if displayName == "" {
		displayName = current.DisplayName
	}

	description := current.Description
	if req.Description != "" {
		description = strings.TrimSpace(req.Description)
	}

	_, execErr := tx.Exec(ctx, `
		UPDATE roles
		SET display_name = $3,
		    description = NULLIF($4, ''),
		    updated_at = NOW()
		WHERE tenant_id = $1::uuid AND role_id = $2::uuid
	`, tenantID, roleID, displayName, description)
	if execErr != nil {
		return nil, execErr
	}

	var permissions []string

	if req.Permissions != nil {
		normalizedPermissions, normalizeErr := normalizePermissions(req.Permissions)
		if normalizeErr != nil {
			return nil, normalizeErr
		}

		permissions = normalizedPermissions

		if _, err := tx.Exec(ctx, `
			DELETE FROM role_permissions
			WHERE tenant_id = $1::uuid AND role_id = $2::uuid
		`, tenantID, roleID); err != nil {
			return nil, err
		}

		for _, permission := range permissions {
			if _, err := tx.Exec(ctx, `
				INSERT INTO role_permissions (role_permission_id, tenant_id, role_id, permission, created_at, updated_at)
				VALUES ($1::uuid, $2::uuid, $3::uuid, $4, NOW(), NOW())
			`, uuid.NewString(), tenantID, roleID, permission); err != nil {
				return nil, err
			}
		}
	} else {
		rows, err := tx.Query(ctx, `
			SELECT permission
			FROM role_permissions
			WHERE tenant_id = $1::uuid AND role_id = $2::uuid
			ORDER BY permission
		`, tenantID, roleID)
		if err != nil {
			return nil, err
		}

		permissions = nil

		for rows.Next() {
			var permission string
			if err := rows.Scan(&permission); err != nil {
				rows.Close()
				return nil, err
			}

			permissions = append(permissions, permission)
		}

		rows.Close()
	}

	var updated platform.RoleDefinition
	if err := tx.QueryRow(ctx, `
		SELECT role_id::text, tenant_id::text, name, COALESCE(display_name, ''), COALESCE(description, ''), COALESCE(is_system, false), created_at, updated_at
		FROM roles
		WHERE tenant_id = $1::uuid AND role_id = $2::uuid
		LIMIT 1
	`, tenantID, roleID).Scan(
		&updated.ID,
		&updated.TenantID,
		&updated.Name,
		&updated.DisplayName,
		&updated.Description,
		&updated.IsSystem,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	); err != nil {
		return nil, err
	}

	updated.Permissions = permissions

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	s.bumpAuthzTenantEpoch(tenantID)

	return &updated, nil
}

func (s *AuthRepository) UpdateUserRoles(ctx context.Context, tenantID string, req platform.UpdateUserRolesRequest) (*platform.UpdateUserRolesResponse, error) {
	if err := s.ensureSystemRoles(ctx, tenantID); err != nil {
		return nil, err
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" {
		return nil, ErrInvalidUserRoleUpdate
	}

	roles := dedupeNonEmpty(req.Roles)
	if len(roles) == 0 {
		return nil, ErrInvalidUserRoleUpdate
	}

	resourceType := normalizeResourceType(req.ResourceType)
	if resourceType == "" {
		return nil, ErrInvalidResourceType
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	var userID string
	if err := tx.QueryRow(ctx, `
		SELECT user_id::text
		FROM users
		WHERE tenant_id = $1::uuid
		  AND LOWER(COALESCE(email, '')) = $2
		LIMIT 1
		FOR UPDATE
	`, tenantID, email).Scan(&userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	type roleRow struct {
		ID       string
		Name     string
		IsSystem bool
	}

	roleRows := make([]roleRow, 0, len(roles))
	for _, roleName := range roles {
		var row roleRow
		if err := tx.QueryRow(ctx, `
			SELECT role_id::text, name, COALESCE(is_system, false)
			FROM roles
			WHERE tenant_id = $1::uuid AND LOWER(name) = $2
			LIMIT 1
		`, tenantID, strings.ToLower(roleName)).Scan(&row.ID, &row.Name, &row.IsSystem); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrRoleNotFound
			}

			return nil, err
		}

		roleRows = append(roleRows, row)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE user_role_bindings
		SET is_active = false,
		    updated_at = NOW()
		WHERE tenant_id = $1::uuid AND user_id = $2::uuid AND is_active = true
	`, tenantID, userID); err != nil {
		return nil, err
	}

	grantedRoles := make([]string, 0, len(roleRows))
	for _, row := range roleRows {
		if _, err := tx.Exec(ctx, `
			INSERT INTO user_role_bindings (binding_id, tenant_id, user_id, role_id, resource_type, resource_id, is_active, created_at, updated_at)
			VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid, $5, NULL, true, NOW(), NOW())
		`, uuid.NewString(), tenantID, userID, row.ID, resourceType); err != nil {
			return nil, err
		}

		grantedRoles = append(grantedRoles, row.Name)
	}

	sort.Strings(grantedRoles)

	primary := derivePrimaryRole(grantedRoles)
	if _, err := tx.Exec(ctx, `
		UPDATE users
		SET role = $3,
		    active = true,
		    updated_at = NOW()
		WHERE tenant_id = $1::uuid AND user_id = $2::uuid
	`, tenantID, userID, primary); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	s.invalidateAuthzCacheForUser(tenantID, userID)
	s.bumpAuthzTenantEpoch(tenantID)

	return &platform.UpdateUserRolesResponse{
		UserID:       userID,
		Email:        email,
		Roles:        grantedRoles,
		ResourceType: resourceType,
	}, nil
}

func (s *AuthRepository) ListUserRoleNamesByUserIDs(ctx context.Context, tenantID string, userIDs []string) (map[string][]string, error) {
	out := make(map[string][]string, len(userIDs))
	if len(userIDs) == 0 {
		return out, nil
	}

	userUUIDs := make([]uuid.UUID, 0, len(userIDs))
	for _, raw := range userIDs {
		id, err := uuid.Parse(strings.TrimSpace(raw))
		if err != nil {
			continue
		}

		userUUIDs = append(userUUIDs, id)
	}

	if len(userUUIDs) == 0 {
		return out, nil
	}

	rows, err := s.pool.Query(ctx, `
		SELECT urb.user_id::text,
		       COALESCE(ARRAY_AGG(r.name ORDER BY r.name) FILTER (WHERE r.name IS NOT NULL), '{}') AS roles
		FROM user_role_bindings urb
		JOIN roles r ON r.role_id = urb.role_id AND r.tenant_id = urb.tenant_id
		WHERE urb.tenant_id = $1::uuid
		  AND urb.is_active = true
		  AND urb.user_id = ANY($2::uuid[])
		GROUP BY urb.user_id
	`, tenantID, userUUIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			userID string
			roles  []string
		)
		if err := rows.Scan(&userID, &roles); err != nil {
			return nil, err
		}

		out[userID] = roles
	}

	return out, rows.Err()
}

func (s *AuthRepository) GetEffectivePermissionsForUser(ctx context.Context, tenantID string, userID string, fallbackRole string) ([]string, []string, error) {
	tenantID = strings.TrimSpace(tenantID)
	userID = strings.TrimSpace(userID)

	if parsedTenantID, err := uuid.Parse(tenantID); err == nil {
		tenantID = parsedTenantID.String()
	}

	roleNames := make([]string, 0, 2)

	if parsedUserID, err := uuid.Parse(userID); err == nil {
		userID = parsedUserID.String()

		rows, queryErr := s.pool.Query(ctx, `
			SELECT r.name
			FROM user_role_bindings urb
			JOIN roles r ON r.role_id = urb.role_id AND r.tenant_id = urb.tenant_id
			WHERE urb.tenant_id = $1::uuid
			  AND urb.user_id = $2::uuid
			  AND urb.is_active = true
			ORDER BY r.name
		`, tenantID, userID)
		if queryErr != nil {
			return nil, nil, queryErr
		}

		for rows.Next() {
			var roleName string
			if err := rows.Scan(&roleName); err != nil {
				rows.Close()
				return nil, nil, err
			}

			roleNames = append(roleNames, strings.TrimSpace(strings.ToLower(roleName)))
		}

		rows.Close()

		if err := rows.Err(); err != nil {
			return nil, nil, err
		}
	}

	if len(roleNames) == 0 {
		fallback := canonicalizeUserRoleName(fallbackRole)
		if fallback == "" {
			fallback = "member"
		}

		roleNames = []string{fallback}
	}

	roleNames = dedupeNonEmptyCanonicalRoles(roleNames)

	permissionSet := make(map[string]struct{})

	for _, roleName := range roleNames {
		roleName = canonicalizeUserRoleName(roleName)

		perms, ok := systemRolePermissions[roleName]
		if !ok {
			rolePermRows, rolePermErr := s.pool.Query(ctx, `
				SELECT rp.permission
				FROM roles r
				JOIN role_permissions rp ON rp.role_id = r.role_id AND rp.tenant_id = r.tenant_id
				WHERE r.tenant_id = $1::uuid AND r.name = $2
			`, tenantID, roleName)
			if rolePermErr != nil {
				return nil, nil, rolePermErr
			}

			for rolePermRows.Next() {
				var permission string
				if err := rolePermRows.Scan(&permission); err != nil {
					rolePermRows.Close()
					return nil, nil, err
				}

				permissionSet[strings.TrimSpace(strings.ToLower(permission))] = struct{}{}
			}

			rolePermRows.Close()

			if err := rolePermRows.Err(); err != nil {
				return nil, nil, err
			}

			continue
		}

		for _, permission := range perms {
			permissionSet[permission] = struct{}{}
		}
	}

	permissions := make([]string, 0, len(permissionSet))
	for permission := range permissionSet {
		permissions = append(permissions, permission)
	}

	sort.Strings(roleNames)
	sort.Strings(permissions)

	return roleNames, permissions, nil
}

func (s *AuthRepository) ensureSystemRoles(ctx context.Context, tenantID string) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer rollbackTxBestEffort(ctx, tx)

	roleNames := make([]string, 0, len(systemRolePermissions))
	for roleName := range systemRolePermissions {
		roleNames = append(roleNames, roleName)
	}

	sort.Strings(roleNames)

	for _, roleName := range roleNames {
		permissions := systemRolePermissions[roleName]

		var roleID string

		displayName := systemRoleDisplayName(roleName)
		if err := tx.QueryRow(ctx, `
			INSERT INTO roles (role_id, tenant_id, name, display_name, description, is_system, created_at, updated_at)
			VALUES ($1::uuid, $2::uuid, $3, $4, $5, true, NOW(), NOW())
			ON CONFLICT (tenant_id, name)
			DO UPDATE SET
				display_name = EXCLUDED.display_name,
				updated_at = NOW()
			RETURNING role_id::text
		`, uuid.NewString(), tenantID, roleName, displayName, "System role").Scan(&roleID); err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, `
			DELETE FROM role_permissions
			WHERE tenant_id = $1::uuid AND role_id = $2::uuid
		`, tenantID, roleID); err != nil {
			return err
		}

		normalized, err := normalizePermissions(permissions)
		if err != nil {
			return err
		}

		for _, permission := range normalized {
			if _, err := tx.Exec(ctx, `
				INSERT INTO role_permissions (role_permission_id, tenant_id, role_id, permission, created_at, updated_at)
				VALUES ($1::uuid, $2::uuid, $3::uuid, $4, NOW(), NOW())
			`, uuid.NewString(), tenantID, roleID, permission); err != nil {
				return err
			}
		}
	}

	return tx.Commit(ctx)
}

func systemRoleDisplayName(roleName string) string {
	parts := strings.Split(strings.ReplaceAll(strings.TrimSpace(roleName), "_", " "), " ")

	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if len(part) == 1 {
			out = append(out, strings.ToUpper(part))
			continue
		}

		out = append(out, strings.ToUpper(part[:1])+strings.ToLower(part[1:]))
	}

	return strings.Join(out, " ")
}

func normalizePermissions(input []string) ([]string, error) {
	if len(input) == 0 {
		return nil, ErrInvalidPermission
	}

	normalized := dedupeNonEmpty(input)
	if len(normalized) == 0 {
		return nil, ErrInvalidPermission
	}

	for _, permission := range normalized {
		if !iam.IsValidPermission(permission) {
			return nil, ErrInvalidPermission
		}
	}

	sort.Strings(normalized)

	return normalized, nil
}

func dedupeNonEmpty(items []string) []string {
	seen := make(map[string]struct{}, len(items))

	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(strings.ToLower(item))
		if item == "" {
			continue
		}

		if _, ok := seen[item]; ok {
			continue
		}

		seen[item] = struct{}{}
		out = append(out, item)
	}

	return out
}

func normalizeResourceType(resourceType string) string {
	switch strings.TrimSpace(strings.ToLower(resourceType)) {
	case "", "tenant":
		return "tenant"
	case "org_unit", "team", "workspace":
		return strings.TrimSpace(strings.ToLower(resourceType))
	default:
		return ""
	}
}

func derivePrimaryRole(roles []string) string {
	if len(roles) == 0 {
		return "member"
	}

	priority := map[string]int{
		"proxy_admin":     100,
		"tenant_admin":    90,
		"org_admin":       80,
		"team_admin":      70,
		"member":          60,
		"readonly_member": 50,
	}
	best := roles[0]
	bestRank := -1

	for _, role := range roles {
		role = canonicalizeUserRoleName(role)

		rank, ok := priority[role]
		if !ok {
			rank = 10
		}

		if rank > bestRank {
			bestRank = rank
			best = role
		}
	}

	return best
}

func canonicalizeUserRoleName(role string) string {
	switch strings.TrimSpace(strings.ToLower(role)) {
	case "user":
		return "member"
	default:
		return strings.TrimSpace(strings.ToLower(role))
	}
}

func dedupeNonEmptyCanonicalRoles(items []string) []string {
	seen := make(map[string]struct{}, len(items))

	out := make([]string, 0, len(items))
	for _, item := range items {
		item = canonicalizeUserRoleName(item)
		if item == "" {
			continue
		}

		if _, ok := seen[item]; ok {
			continue
		}

		seen[item] = struct{}{}
		out = append(out, item)
	}

	return out
}
