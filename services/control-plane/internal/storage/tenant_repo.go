package storage

import (
	"context"
	"net"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/jackc/pgx/v5"
)

// GetTenantByHost resolves a tenant from the request host.
func (s *TenantRepository) GetTenantByHost(ctx context.Context, host string) (*platform.Tenant, error) {
	slug := tenantSlugFromHost(host)
	if slug == "" {
		return nil, pgx.ErrNoRows
	}

	var tenant platform.Tenant

	err := s.pool.QueryRow(ctx, `
		SELECT tenant_id::text, slug, domain, status, created_at
		FROM tenants
		WHERE slug = $1 OR domain = $2
		LIMIT 1
	`, slug, host).Scan(&tenant.ID, &tenant.Slug, &tenant.Domain, &tenant.Status, &tenant.CreatedAt)
	if err == nil {
		return &tenant, nil
	}

	return nil, err
}

func (s *TenantRepository) GetTenantBySlug(ctx context.Context, slug string) (*platform.Tenant, error) {
	slug = strings.TrimSpace(strings.ToLower(slug))
	if slug == "" {
		return nil, pgx.ErrNoRows
	}

	var tenant platform.Tenant

	err := s.pool.QueryRow(ctx, `
		SELECT tenant_id::text, slug, domain, status, created_at
		FROM tenants
		WHERE slug = $1
		LIMIT 1
	`, slug).Scan(&tenant.ID, &tenant.Slug, &tenant.Domain, &tenant.Status, &tenant.CreatedAt)
	if err == nil {
		return &tenant, nil
	}

	return nil, err
}

func (s *TenantRepository) ListTenantsByAdminIdentifier(ctx context.Context, channel string, identifier string) ([]platform.Tenant, error) {
	channel = strings.TrimSpace(strings.ToLower(channel))

	identifier = strings.TrimSpace(identifier)
	if channel == "" || identifier == "" {
		return []platform.Tenant{}, nil
	}

	var condition string

	switch channel {
	case "email":
		condition = "LOWER(COALESCE(u.email, '')) = LOWER($1)"
	case "mobile":
		condition = "COALESCE(u.mobile, '') = $1"
	default:
		return []platform.Tenant{}, nil
	}

	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT t.tenant_id::text, t.slug, t.domain, t.status, t.created_at
		FROM users u
		JOIN tenants t ON t.tenant_id = u.tenant_id
		WHERE `+condition+`
		  AND COALESCE(u.active, true) = true
		  AND COALESCE(t.status, '') <> 'deleted'
		ORDER BY t.slug ASC
	`, identifier)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]platform.Tenant, 0)

	for rows.Next() {
		var tenant platform.Tenant
		if err := rows.Scan(&tenant.ID, &tenant.Slug, &tenant.Domain, &tenant.Status, &tenant.CreatedAt); err != nil {
			return nil, err
		}

		items = append(items, tenant)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func tenantSlugFromHost(host string) string {
	host = strings.TrimSpace(host)
	if host == "" {
		return ""
	}

	if parsedHost, _, err := net.SplitHostPort(host); err == nil {
		host = parsedHost
	}

	if ip := net.ParseIP(host); ip != nil {
		return ""
	}

	parts := strings.Split(host, ".")
	if len(parts) == 0 {
		return ""
	}

	return parts[0]
}
