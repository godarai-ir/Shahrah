package storage

import (
	"context"
	"net"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

// GetTenantByHost resolves a tenant from the request host.
func (s *Store) GetTenantByHost(ctx context.Context, host string) (*platform.Tenant, error) {
	slug := tenantSlugFromHost(host)
	if slug == "" || slug == "localhost" {
		return s.EnsureTenant(ctx, s.defaultTenant)
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

// EnsureTenant creates the tenant and wallet when missing.
func (s *Store) EnsureTenant(ctx context.Context, slug string) (*platform.Tenant, error) {
	if slug == "" {
		slug = s.defaultTenant
	}

	var tenant platform.Tenant
	err := s.pool.QueryRow(ctx, `
		INSERT INTO tenants (slug, domain, status)
		VALUES ($1, $2, 'active')
		ON CONFLICT (slug)
		DO UPDATE SET slug = EXCLUDED.slug
		RETURNING tenant_id::text, slug, domain, status, created_at
	`, slug, slug).Scan(&tenant.ID, &tenant.Slug, &tenant.Domain, &tenant.Status, &tenant.CreatedAt)
	if err != nil {
		return nil, err
	}

	_, err = s.pool.Exec(ctx, `
		INSERT INTO wallets (tenant_id, balance_irr, pending_irr)
		VALUES ($1, 0, 0)
		ON CONFLICT (tenant_id) DO NOTHING
	`, tenant.ID)
	if err != nil {
		return nil, err
	}

	return &tenant, nil
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
		return "default"
	}
	parts := strings.Split(host, ".")
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}
