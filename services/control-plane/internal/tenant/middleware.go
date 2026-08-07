// Package tenant resolves tenant context from requests and caches.
package tenant

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"
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

const defaultLocalCacheMaxEntries = 4096

// MiddlewareConfig holds configuration for the middleware.
type MiddlewareConfig struct {
	AllowHeaderHint      bool
	CacheTTL             time.Duration
	LocalCacheTTL        time.Duration
	LocalCacheMaxEntries int
	RuntimeProfiles      RuntimeProfileGetter
}

type localTenantEntry struct {
	tenant    platform.Tenant
	expiresAt time.Time
}

// Middleware resolves tenants from host and optional local-dev header hints.
type Middleware struct {
	store                Store
	cache                Cache
	runtimeProfiles      RuntimeProfileGetter
	defaultTenant        string
	allowHeaderHint      bool
	cacheTTL             time.Duration
	localCacheTTL        time.Duration
	localCacheMaxEntries int
	localMu              sync.RWMutex
	localTenants         map[string]localTenantEntry
	loadGroup            singleflight.Group
}

// RuntimeProfileGetter resolves compiled tenant state.
type RuntimeProfileGetter interface {
	GetSnapshot(ctx context.Context, tenantID string) (*runtimeprofile.Profile, error)
}

// NewMiddleware creates a new tenant middleware without caching.
//
// Deprecated: Use NewMiddlewareWithCache instead.
func NewMiddleware(store Store, defaultTenant string, allowHeaderHint bool) *Middleware {
	return &Middleware{
		store:                store,
		defaultTenant:        strings.TrimSpace(defaultTenant),
		allowHeaderHint:      allowHeaderHint,
		cacheTTL:             5 * time.Minute, // default TTL
		localCacheTTL:        time.Second,
		localCacheMaxEntries: defaultLocalCacheMaxEntries,
		localTenants:         make(map[string]localTenantEntry),
	}
}

// NewMiddlewareWithCache creates a new tenant middleware with caching support.
func NewMiddlewareWithCache(store Store, c Cache, cfg MiddlewareConfig) *Middleware {
	cacheTTL := cfg.CacheTTL
	if cacheTTL <= 0 {
		cacheTTL = 5 * time.Minute
	}

	localCacheTTL := cfg.LocalCacheTTL
	if localCacheTTL <= 0 {
		localCacheTTL = time.Second
	}

	if localCacheTTL > cacheTTL {
		localCacheTTL = cacheTTL
	}

	localCacheMaxEntries := cfg.LocalCacheMaxEntries
	if localCacheMaxEntries <= 0 {
		localCacheMaxEntries = defaultLocalCacheMaxEntries
	}

	return &Middleware{
		store:                store,
		cache:                c,
		runtimeProfiles:      cfg.RuntimeProfiles,
		allowHeaderHint:      cfg.AllowHeaderHint,
		cacheTTL:             cacheTTL,
		localCacheTTL:        localCacheTTL,
		localCacheMaxEntries: localCacheMaxEntries,
		localTenants:         make(map[string]localTenantEntry),
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

		host = normalizeTenantHost(host)

		if tenant, found := m.localTenant(host); found {
			return m.continueWithTenant(c, tenant)
		}

		tenant, err := m.resolveTenant(c.Context(), host)
		if err != nil {
			return fiber.NewError(fiber.StatusNotFound, "tenant not found")
		}

		m.rememberLocalTenant(host, tenant)

		return m.continueWithTenant(c, tenant)
	}
}

func (m *Middleware) continueWithTenant(c *fiber.Ctx, resolved *platform.Tenant) error {
	tenant := resolved

	profile := m.runtimeProfile(c, tenant)
	if profile != nil && profile.Tenant.ID == tenant.ID {
		tenant = cloneTenant(&profile.Tenant)
	}

	if !isActiveTenant(tenant) {
		return fiber.NewError(fiber.StatusForbidden, fmt.Sprintf("tenant is not active (status: %s)", tenant.Status))
	}

	c.Locals(ContextKey{}, tenant)

	if profile != nil {
		c.Locals(RuntimeProfileContextKey{}, profile)
	}

	return c.Next()
}

func (m *Middleware) runtimeProfile(c *fiber.Ctx, tenant *platform.Tenant) *runtimeprofile.Profile {
	if m == nil || m.runtimeProfiles == nil || tenant == nil {
		return nil
	}

	profile, err := m.runtimeProfiles.GetSnapshot(c.Context(), tenant.ID)
	if err != nil || profile == nil {
		return nil
	}

	return profile
}

func (m *Middleware) localTenant(host string) (*platform.Tenant, bool) {
	if m == nil || m.localCacheTTL <= 0 {
		return nil, false
	}

	key := cacheKeyForHost(host)

	m.localMu.RLock()
	entry, ok := m.localTenants[key]
	m.localMu.RUnlock()

	now := time.Now()
	if !ok || now.After(entry.expiresAt) {
		if ok {
			m.localMu.Lock()
			if current, exists := m.localTenants[key]; exists && current.expiresAt.Equal(entry.expiresAt) && now.After(current.expiresAt) {
				delete(m.localTenants, key)
			}
			m.localMu.Unlock()
		}

		return nil, false
	}

	return cloneTenant(&entry.tenant), true
}

func (m *Middleware) rememberLocalTenant(host string, tenant *platform.Tenant) {
	if m == nil || tenant == nil || m.localCacheTTL <= 0 || !isActiveTenant(tenant) {
		return
	}

	entry := localTenantEntry{
		tenant:    *tenant,
		expiresAt: time.Now().Add(m.localCacheTTL),
	}

	key := cacheKeyForHost(host)

	maxEntries := m.localCacheMaxEntries
	if maxEntries <= 0 {
		maxEntries = defaultLocalCacheMaxEntries
	}

	m.localMu.Lock()
	if _, exists := m.localTenants[key]; !exists && len(m.localTenants) >= maxEntries {
		m.evictOneLocalTenantLocked()
	}

	m.localTenants[key] = entry
	m.localMu.Unlock()
}

func (m *Middleware) evictOneLocalTenantLocked() {
	for key := range m.localTenants {
		delete(m.localTenants, key)

		return
	}
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
	hostKey := cacheKeyForHost(slug)
	hostKeys := map[string]struct{}{hostKey: {}}

	m.localMu.Lock()
	for key, entry := range m.localTenants {
		if key == hostKey || strings.EqualFold(strings.TrimSpace(entry.tenant.ID), strings.TrimSpace(tenantID)) {
			hostKeys[key] = struct{}{}
			delete(m.localTenants, key)
		}
	}
	m.localMu.Unlock()

	if m.cache != nil {
		for key := range hostKeys {
			m.cache.Delete(ctx, key)
		}

		m.cache.Delete(ctx, cacheKeyForID(tenantID))
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
