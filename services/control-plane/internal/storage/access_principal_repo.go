package storage

import (
	"context"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

// SearchAccessPrincipals returns platform users/teams for tenant-admin access pickers.
func (s *Store) SearchAccessPrincipals(ctx context.Context, tenantID string, kind string, query string, limit int) ([]platform.AccessPrincipalCandidate, error) {
	normalizedKind := strings.TrimSpace(strings.ToLower(kind))
	normalizedQuery := strings.TrimSpace(query)
	normalizedTenantID := strings.TrimSpace(tenantID)
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	switch normalizedKind {
	case "user":
		return s.searchUserPrincipals(ctx, normalizedTenantID, normalizedQuery, limit)
	case "team":
		return s.searchTeamPrincipals(ctx, normalizedTenantID, normalizedQuery, limit)
	default:
		return []platform.AccessPrincipalCandidate{}, nil
	}
}

func (s *Store) searchUserPrincipals(ctx context.Context, tenantID string, query string, limit int) ([]platform.AccessPrincipalCandidate, error) {
	like := "%" + query + "%"
	rows, err := s.pool.Query(ctx, `
		WITH current_tenant AS (
			SELECT tenant_id, slug
			FROM tenants
			WHERE tenant_id = NULLIF($1, '')::uuid
			LIMIT 1
		),
		candidates AS (
			SELECT user_id::text AS id, COALESCE(user_alias, '') AS alias, COALESCE(email, '') AS email, updated_at, created_at
			FROM users
			WHERE tenant_id = NULLIF($1, '')::uuid
			  AND ($2 = '' OR user_alias ILIKE $3 OR email ILIKE $3)
			UNION ALL
			SELECT tau."id" AS id,
			       COALESCE(NULLIF(tau."email", ''), NULLIF(tau."mobile", ''), tau."id") AS alias,
			       COALESCE(tau."email", '') AS email,
			       tau."updatedAt" AS updated_at,
			       tau."createdAt" AS created_at
			FROM "tenant_admin_users" tau
			JOIN current_tenant ct ON tau."tenantSlug" = ct.slug
			WHERE tau."id" ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
			  AND NOT EXISTS (SELECT 1 FROM users u WHERE u.user_id::text = tau."id")
			  AND ($2 = '' OR tau."email" ILIKE $3 OR tau."mobile" ILIKE $3)
		)
		SELECT id, alias, email
		FROM candidates
		ORDER BY updated_at DESC NULLS LAST, created_at DESC NULLS LAST
		LIMIT $4
	`, tenantID, query, like, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]platform.AccessPrincipalCandidate, 0)
	for rows.Next() {
		var id string
		var alias string
		var email string
		if err := rows.Scan(&id, &alias, &email); err != nil {
			return nil, err
		}
		displayName := strings.TrimSpace(email)
		if displayName == "" {
			displayName = strings.TrimSpace(alias)
		}
		if displayName == "" {
			displayName = id
		}
		secondary := strings.TrimSpace(alias)
		if secondary == displayName {
			secondary = ""
		}

		items = append(items, platform.AccessPrincipalCandidate{
			Type:        platform.PrincipalTypeUser,
			PrincipalID: id,
			DisplayName: displayName,
			Secondary:   secondary,
		})
	}
	return items, rows.Err()
}

func (s *Store) searchTeamPrincipals(ctx context.Context, tenantID string, query string, limit int) ([]platform.AccessPrincipalCandidate, error) {
	like := "%" + query + "%"
	rows, err := s.pool.Query(ctx, `
		SELECT team_id::text, COALESCE(team_alias, ''), COALESCE(organization_alias, '')
		FROM teams
		LEFT JOIN organizations ON teams.organization_id = organizations.organization_id
		WHERE teams.tenant_id = NULLIF($1, '')::uuid
		  AND ($2 = '' OR team_alias ILIKE $3 OR organization_alias ILIKE $3)
		ORDER BY teams.updated_at DESC NULLS LAST, teams.created_at DESC NULLS LAST
		LIMIT $4
	`, tenantID, query, like, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]platform.AccessPrincipalCandidate, 0)
	for rows.Next() {
		var id string
		var alias string
		var orgAlias string
		if err := rows.Scan(&id, &alias, &orgAlias); err != nil {
			return nil, err
		}
		displayName := strings.TrimSpace(alias)
		if displayName == "" {
			displayName = id
		}
		secondary := strings.TrimSpace(orgAlias)

		items = append(items, platform.AccessPrincipalCandidate{
			Type:        platform.PrincipalTypeTeam,
			PrincipalID: id,
			DisplayName: displayName,
			Secondary:   secondary,
		})
	}
	return items, rows.Err()
}
