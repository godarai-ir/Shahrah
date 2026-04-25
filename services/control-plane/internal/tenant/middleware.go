package tenant

import (
	"context"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/gofiber/fiber/v2"
)

// ContextKey is the local key for tenant context.
type ContextKey struct{}

// Store is the subset of tenant storage used by middleware.
type Store interface {
	GetTenantByHost(ctx context.Context, host string) (*platform.Tenant, error)
	EnsureTenant(ctx context.Context, slug string) (*platform.Tenant, error)
}

// Middleware resolves tenants from host and optional local-dev header hints.
type Middleware struct {
	store           Store
	defaultTenant   string
	allowHeaderHint bool
}

// NewMiddleware creates a new tenant middleware.
func NewMiddleware(store Store, defaultTenant string, allowHeaderHint bool) *Middleware {
	return &Middleware{store: store, defaultTenant: defaultTenant, allowHeaderHint: allowHeaderHint}
}

// Handle resolves and stores the tenant context.
func (m *Middleware) Handle() fiber.Handler {
	return func(c *fiber.Ctx) error {
		host := c.Hostname()
		if m.allowHeaderHint && (host == "localhost" || strings.HasPrefix(host, "127.")) {
			if hinted := c.Get("x-shahrah-tenant"); hinted != "" {
				host = hinted
			}
		}

		tenant, err := m.store.GetTenantByHost(c.Context(), host)
		if err != nil {
			tenant, err = m.store.EnsureTenant(c.Context(), m.defaultTenant)
			if err != nil {
				return fiber.NewError(fiber.StatusNotFound, "tenant not found")
			}
		}

		c.Locals(ContextKey{}, tenant)
		return c.Next()
	}
}

// FromContext returns the current tenant from locals.
func FromContext(c *fiber.Ctx) *platform.Tenant {
	if value := c.Locals(ContextKey{}); value != nil {
		if tenant, ok := value.(*platform.Tenant); ok {
			return tenant
		}
	}
	return nil
}
