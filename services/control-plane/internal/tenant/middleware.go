// Package tenant resolves tenant context from requests and caches.
package tenant

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/apiheaders"
	"github.com/Delpak/shahrah/services/control-plane/internal/runtimeprofile"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/sync/singleflight"
)

// ContextKey is the local key for tenant context.
type ContextKey struct{}
type RuntimeProfileContextKey struct{}

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
	runtimeProfiles runtimeProfileGetter
	defaultTenant   string
	allowHeaderHint bool
	cacheTTL        time.Duration
	loadGroup       singleflight.Group
}

type runtimeProfileGetter interface {
	Get(ctx context.Context, tenantID string) (*runtimeprofile.Profile, error)
}

// NewMiddleware creates a new tenant middleware without caching.
//
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

func (m *Middleware) SetRuntimeProfiles(runtimeProfiles runtimeProfileGetter) {
	m.runtimeProfiles = runtimeProfiles
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

		host = normalizeTenantHost(host)

		// Check cache first
		if m.cache != nil {
			cacheKey := cacheKeyForHost(host)
			if cached, found := m.cache.Get(c.Context(), cacheKey); found && cached != nil {
				if tenant, ok := tenantFromCache(cached); ok {
					if !isActiveTenant(tenant) {
						return fiber.NewError(fiber.StatusForbidden, fmt.Sprintf("tenant is not active (status: %s)", tenant.Status))
					}

					c.Locals(ContextKey{}, tenant)
					m.attachRuntimeProfile(c, tenant)

					return c.Next()
				}
			}
		}

		tenant, err := m.resolveTenant(c.Context(), host)
		if err != nil {
			return fiber.NewError(fiber.StatusNotFound, "tenant not found")
		}

		// Check tenant status
		if !isActiveTenant(tenant) {
			return fiber.NewError(fiber.StatusForbidden, fmt.Sprintf("tenant is not active (status: %s)", tenant.Status))
		}

		c.Locals(ContextKey{}, tenant)
		m.attachRuntimeProfile(c, tenant)

		return c.Next()
	}
}

func (m *Middleware) attachRuntimeProfile(c *fiber.Ctx, tenant *platform.Tenant) {
	if m == nil || m.runtimeProfiles == nil || tenant == nil {
		return
	}

	profile, err := m.runtimeProfiles.Get(c.Context(), tenant.ID)
	if err != nil || profile == nil {
		return
	}

	c.Locals(RuntimeProfileContextKey{}, profile)
}

func (m *Middleware) resolveTenant(ctx context.Context, host string) (*platform.Tenant, error) {
	value, err, _ := m.loadGroup.Do(cacheKeyForHost(host), func() (interface{}, error) {
		if m.cache != nil {
			if cached, found := m.cache.Get(ctx, cacheKeyForHost(host)); found && cached != nil {
				if tenant, ok := tenantFromCache(cached); ok {
					return tenant, nil
				}
			}
		}

		tenant, err := m.store.GetTenantByHost(ctx, host)
		if err != nil {
			return nil, err
		}

		if !isActiveTenant(tenant) {
			return tenant, nil
		}

		if m.cache != nil {
			cacheTenant(ctx, m.cache, cacheKeyForHost(host), tenant, m.cacheTTL)
		}

		return tenant, nil
	})
	if err != nil {
		return nil, err
	}

	tenant, ok := value.(*platform.Tenant)
	if !ok || tenant == nil {
		return nil, fiber.ErrNotFound
	}

	return tenant, nil
}

func tenantFromCache(value interface{}) (*platform.Tenant, bool) {
	switch typed := value.(type) {
	case *platform.Tenant:
		return cloneTenant(typed), typed != nil
	case platform.Tenant:
		return cloneTenant(&typed), true
	case string:
		return tenantFromJSON([]byte(typed))
	case []byte:
		return tenantFromJSON(typed)
	default:
		return nil, false
	}
}

func tenantFromJSON(raw []byte) (*platform.Tenant, bool) {
	var tenant platform.Tenant
	if err := json.Unmarshal(raw, &tenant); err != nil {
		return nil, false
	}

	return &tenant, true
}

func cacheTenant(ctx context.Context, c Cache, key string, tenant *platform.Tenant, ttl time.Duration) {
	if tenant == nil {
		return
	}

	raw, err := json.Marshal(tenant)
	if err != nil {
		return
	}

	c.Set(ctx, key, string(raw), ttl)
}

func cloneTenant(in *platform.Tenant) *platform.Tenant {
	if in == nil {
		return nil
	}

	out := *in

	return &out
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
func (m *Middleware) InvalidateTenant(ctx context.Context, slug, tenantID string) {
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

func RuntimeProfileFromContext(c *fiber.Ctx) *runtimeprofile.Profile {
	if value := c.Locals(RuntimeProfileContextKey{}); value != nil {
		if profile, ok := value.(*runtimeprofile.Profile); ok {
			return profile
		}
	}

	return nil
}

func cacheKeyForHost(host string) string {
	return "tenant:host:" + normalizeTenantHost(host)
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

func normalizeTenantHost(host string) string {
	normalized := strings.TrimSpace(strings.ToLower(host))
	if normalized == "" {
		return ""
	}

	if parsedHost, _, err := net.SplitHostPort(normalized); err == nil {
		normalized = parsedHost
	}

	return normalized
}
