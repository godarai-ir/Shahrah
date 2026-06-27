package tenant

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/apiheaders"
	"github.com/gofiber/fiber/v2"
)

// ContextKey is the local key for tenant context.
type ContextKey struct{}

// Store is the subset of tenant storage used by middleware.
type Store interface {
	GetTenantByHost(ctx context.Context, host string) (*platform.Tenant, error)
}

// Cache defines the cache operations used by the middleware.
type Cache interface {
	Get(ctx context.Context, key string) (interface{}, bool)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration)
	Delete(ctx context.Context, key string)
}

// MiddlewareConfig holds configuration for the middleware.
type MiddlewareConfig struct {
	AllowHeaderHint bool
	CacheTTL        time.Duration
}

// Middleware resolves tenants from host and optional local-dev header hints.
type Middleware struct {
	store           Store
	cache           Cache
	defaultTenant   string
	allowHeaderHint bool
	cacheTTL        time.Duration
}

// NewMiddleware creates a new tenant middleware without caching.
// Deprecated: Use NewMiddlewareWithCache instead.
func NewMiddleware(store Store, defaultTenant string, allowHeaderHint bool) *Middleware {
	return &Middleware{
		store:           store,
		defaultTenant:   strings.TrimSpace(defaultTenant),
		allowHeaderHint: allowHeaderHint,
		cacheTTL:        5 * time.Minute, // default TTL
	}
}

// NewMiddlewareWithCache creates a new tenant middleware with caching support.
func NewMiddlewareWithCache(store Store, c Cache, cfg MiddlewareConfig) *Middleware {
	cacheTTL := cfg.CacheTTL
	if cacheTTL <= 0 {
		cacheTTL = 5 * time.Minute
	}
	return &Middleware{
		store:           store,
		cache:           c,
		allowHeaderHint: cfg.AllowHeaderHint,
		cacheTTL:        cacheTTL,
	}
}

// Handle resolves and stores the tenant context.
func (m *Middleware) Handle() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if isTenantResolutionBypassed(c.Path()) {
			return c.Next()
		}

		host := c.Hostname()

		// Allow header hint for local development and trusted internal hosts.
		if m.allowHeaderHint && shouldUseTenantHeaderHint(host) {
			if hinted := apiheaders.GetTenant(c); hinted != "" {
				host = hinted + ".localhost"
			}
		}

		if (host == "localhost" || strings.HasPrefix(host, "127.")) && m.defaultTenant != "" {
			host = m.defaultTenant + ".localhost"
		}

		// Check cache first
		if m.cache != nil {
			cacheKey := cacheKeyForHost(host)
			if cached, found := m.cache.Get(c.Context(), cacheKey); found && cached != nil {
				if tenant, ok := cached.(*platform.Tenant); ok {
					c.Locals(ContextKey{}, tenant)
					return c.Next()
				}
			}
		}

		// Cache miss - query database
		tenant, err := m.store.GetTenantByHost(c.Context(), host)
		if err != nil {
			return fiber.NewError(fiber.StatusNotFound, "tenant not found")
		}

		// Check tenant status
		if !isActiveTenant(tenant) {
			return fiber.NewError(fiber.StatusForbidden, fmt.Sprintf("tenant is not active (status: %s)", tenant.Status))
		}

		// Cache the result
		if m.cache != nil {
			cacheKey := cacheKeyForHost(host)
			m.cache.Set(c.Context(), cacheKey, tenant, m.cacheTTL)
		}

		c.Locals(ContextKey{}, tenant)
		return c.Next()
	}
}

func isTenantResolutionBypassed(path string) bool {
	return path == "/health" ||
		path == "/ready" ||
		path == "/metrics" ||
		strings.HasPrefix(path, "/api/v1/auth/") ||
		strings.HasPrefix(path, "/api/v1/onboarding/") ||
		strings.HasPrefix(path, "/api/v1/wallet/topup/callback/")
}

// InvalidateTenant removes the cached tenant entries.
// Call this after updating a tenant via admin API.
func (m *Middleware) InvalidateTenant(ctx context.Context, slug string, tenantID string) {
	if m.cache == nil {
		return
	}
	m.cache.Delete(ctx, cacheKeyForHost(slug))
	m.cache.Delete(ctx, cacheKeyForID(tenantID))
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

func cacheKeyForHost(host string) string {
	return "tenant:host:" + host
}

func cacheKeyForID(tenantID string) string {
	return "tenant:id:" + tenantID
}

func isActiveTenant(t *platform.Tenant) bool {
	if t == nil {
		return false
	}
	return strings.EqualFold(t.Status, "active")
}

func shouldUseTenantHeaderHint(host string) bool {
	normalized := strings.TrimSpace(strings.ToLower(host))
	if normalized == "" {
		return false
	}
	if parsedHost, _, err := net.SplitHostPort(normalized); err == nil {
		normalized = parsedHost
	}

	return normalized == "localhost" ||
		strings.HasPrefix(normalized, "127.") ||
		normalized == "control-plane" ||
		strings.HasPrefix(normalized, "control-plane.") ||
		strings.HasSuffix(normalized, ".svc") ||
		strings.Contains(normalized, ".svc.")
}
