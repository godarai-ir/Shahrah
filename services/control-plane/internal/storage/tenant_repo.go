package storage

import (
	"context"
	"net"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/jackc/pgx/v5"
)

// GetTenantByHost resolves a tenant from the request host.
func (s *Store) GetTenantByHost(ctx context.Context, host string) (*platform.Tenant, error) {
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
