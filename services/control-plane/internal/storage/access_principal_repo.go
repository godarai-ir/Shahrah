package storage

import (
	"context"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

// SearchAccessPrincipals returns platform users/teams/org-units for tenant-admin access pickers.
func (s *ProviderAccountRepository) SearchAccessPrincipals(ctx context.Context, tenantID, kind, query string, limit int) ([]platform.AccessPrincipalCandidate, error) {
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

func (s *ProviderAccountRepository) searchUserPrincipals(ctx context.Context, tenantID, query string, limit int) ([]platform.AccessPrincipalCandidate, error) {
	like := "%" + query + "%"

	rows, err := s.pool.Query(ctx, `
		SELECT user_id::text AS id,
		       COALESCE(NULLIF(user_alias, ''), NULLIF(email, ''), NULLIF(mobile, ''), user_id::text) AS alias,
		       COALESCE(email, '') AS email
		FROM users
		WHERE tenant_id = NULLIF($1, '')::uuid
		  AND ($2 = '' OR user_alias ILIKE $3 OR email ILIKE $3 OR mobile ILIKE $3)
		ORDER BY updated_at DESC NULLS LAST, created_at DESC NULLS LAST
		LIMIT $4
	`, tenantID, query, like, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]platform.AccessPrincipalCandidate, 0)

	for rows.Next() {
		var (
			id    string
			alias string
			email string
		)
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

func (s *ProviderAccountRepository) searchTeamPrincipals(ctx context.Context, tenantID, query string, limit int) ([]platform.AccessPrincipalCandidate, error) {
	like := "%" + query + "%"

	rows, err := s.pool.Query(ctx, `
		SELECT team_id::text, COALESCE(team_alias, '')
		FROM teams
		WHERE teams.tenant_id = NULLIF($1, '')::uuid
		  AND ($2 = '' OR team_alias ILIKE $3)
		ORDER BY teams.updated_at DESC NULLS LAST, teams.created_at DESC NULLS LAST
		LIMIT $4
	`, tenantID, query, like, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]platform.AccessPrincipalCandidate, 0)

	for rows.Next() {
		var (
			id    string
			alias string
		)
		if err := rows.Scan(&id, &alias); err != nil {
			return nil, err
		}

		displayName := strings.TrimSpace(alias)
		if displayName == "" {
			displayName = id
		}

		items = append(items, platform.AccessPrincipalCandidate{
			Type:        platform.PrincipalTypeTeam,
			PrincipalID: id,
			DisplayName: displayName,
		})
	}

	return items, rows.Err()
}
