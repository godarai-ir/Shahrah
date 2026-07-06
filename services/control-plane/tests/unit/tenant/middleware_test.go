package tenant_test

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/apiheaders"
	tenantmw "github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// mockTenantStore implements tenant.Store for testing.
type mockTenantStore struct {
	mu            sync.Mutex
	tenantsByHost map[string]*platform.Tenant
	calls         int
	delay         time.Duration
}

func (s *mockTenantStore) GetTenantByHost(_ context.Context, host string) (*platform.Tenant, error) {
	s.mu.Lock()
	s.calls++
	delay := s.delay
	tenant, ok := s.tenantsByHost[host]
	s.mu.Unlock()

	if delay > 0 {
		time.Sleep(delay)
	}

	if ok {
		return tenant, nil
	}

	return nil, errors.New("tenant not found")
}

func (s *mockTenantStore) callCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.calls
}

func (s *mockTenantStore) addTenant(host string, t *platform.Tenant) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tenantsByHost[host] = t
}

// mockCache implements tenant.Cache for testing.
type mockCache struct {
	mu      sync.Mutex
	data    map[string]interface{}
	ttls    map[string]time.Duration
	gets    int
	sets    int
	deletes int
}

func newMockCache() *mockCache {
	return &mockCache{
		data: make(map[string]interface{}),
		ttls: make(map[string]time.Duration),
	}
}

func (c *mockCache) Get(_ context.Context, key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.gets++
	val, ok := c.data[key]

	return val, ok
}

func (c *mockCache) Set(_ context.Context, key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.sets++
	c.data[key] = value
	c.ttls[key] = ttl
}

func (c *mockCache) Delete(_ context.Context, key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.deletes++
	delete(c.data, key)
}

func newActiveTenant(id, slug, domain, status string) *platform.Tenant {
	return &platform.Tenant{
		ID:        id,
		Slug:      slug,
		Domain:    domain,
		Status:    status,
		CreatedAt: time.Now(),
	}
}

func TestTenantMiddleware_CacheHit(t *testing.T) {
	app := fiber.New()
	cache := newMockCache()
	store := &mockTenantStore{
		tenantsByHost: make(map[string]*platform.Tenant),
	}

	// Pre-populate cache
	cache.data["tenant:host:acme.localhost"] = newActiveTenant("tenant-1", "acme", "acme.localhost", "active")

	mw := tenantmw.NewMiddlewareWithCache(store, cache, tenantmw.MiddlewareConfig{
		AllowHeaderHint: false,
		CacheTTL:        5 * time.Minute,
	})

	app.Use(mw.Handle())
	app.Get("/", func(c *fiber.Ctx) error {
		current := tenantmw.FromContext(c)
		assert.NotNil(t, current)
		assert.Equal(t, "tenant-1", current.ID)

		return c.SendStatus(fiber.StatusNoContent)
	})

	req, err := http.NewRequest(http.MethodGet, "http://acme.localhost/", http.NoBody)
	assert.NoError(t, err)

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)

	// Store should not be called on cache hit
	assert.Equal(t, 0, store.callCount())
}

func TestTenantMiddleware_CacheMiss(t *testing.T) {
	app := fiber.New()
	cache := newMockCache()
	store := &mockTenantStore{
		tenantsByHost: make(map[string]*platform.Tenant),
	}
	store.addTenant("acme.localhost", newActiveTenant("tenant-1", "acme", "acme.localhost", "active"))

	mw := tenantmw.NewMiddlewareWithCache(store, cache, tenantmw.MiddlewareConfig{
		AllowHeaderHint: false,
		CacheTTL:        5 * time.Minute,
	})

	app.Use(mw.Handle())
	app.Get("/", func(c *fiber.Ctx) error {
		current := tenantmw.FromContext(c)
		assert.NotNil(t, current)
		assert.Equal(t, "tenant-1", current.ID)

		return c.SendStatus(fiber.StatusNoContent)
	})

	req, err := http.NewRequest(http.MethodGet, "http://acme.localhost/", http.NoBody)
	assert.NoError(t, err)

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)

	// Cache should have been populated
	assert.Equal(t, 1, cache.sets)
	_, found := cache.Get(context.Background(), "tenant:host:acme.localhost")
	assert.True(t, found)
}

func TestTenantMiddleware_ConcurrentCacheMissSingleflightsStoreLookup(t *testing.T) {
	app := fiber.New()
	cache := newMockCache()
	store := &mockTenantStore{
		tenantsByHost: make(map[string]*platform.Tenant),
		delay:         25 * time.Millisecond,
	}
	store.addTenant("acme.localhost", newActiveTenant("tenant-1", "acme", "acme.localhost", "active"))

	mw := tenantmw.NewMiddlewareWithCache(store, cache, tenantmw.MiddlewareConfig{
		AllowHeaderHint: false,
		CacheTTL:        5 * time.Minute,
	})

	app.Use(mw.Handle())
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	const workers = 16
	var wg sync.WaitGroup
	wg.Add(workers)
	errs := make(chan error, workers)
	statuses := make(chan int, workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()

			req, err := http.NewRequest(http.MethodGet, "http://acme.localhost/", http.NoBody)
			if err != nil {
				errs <- err
				return
			}

			resp, err := app.Test(req)
			if resp != nil && resp.Body != nil {
				resp.Body.Close()
			}
			if err != nil {
				errs <- err
				return
			}

			statuses <- resp.StatusCode
		}()
	}
	wg.Wait()
	close(errs)
	close(statuses)

	for err := range errs {
		assert.NoError(t, err)
	}
	for status := range statuses {
		assert.Equal(t, fiber.StatusNoContent, status)
	}
	assert.Equal(t, 1, store.callCount())
	assert.Equal(t, 1, cache.sets)
}

func TestTenantMiddleware_TenantNotFound_ReturnsNotFound(t *testing.T) {
	app := fiber.New()
	cache := newMockCache()
	store := &mockTenantStore{
		tenantsByHost: make(map[string]*platform.Tenant),
	}

	mw := tenantmw.NewMiddlewareWithCache(store, cache, tenantmw.MiddlewareConfig{
		AllowHeaderHint: false,
		CacheTTL:        5 * time.Minute,
	})

	app.Use(mw.Handle())
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	req, err := http.NewRequest(http.MethodGet, "http://unknown.localhost/", http.NoBody)
	assert.NoError(t, err)

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	assert.Empty(t, store.tenantsByHost)
}

func TestTenantMiddleware_TenantSuspended(t *testing.T) {
	app := fiber.New()
	cache := newMockCache()
	store := &mockTenantStore{
		tenantsByHost: make(map[string]*platform.Tenant),
	}
	store.addTenant("suspended.localhost", newActiveTenant("tenant-1", "suspended", "suspended.localhost", "suspended"))

	mw := tenantmw.NewMiddlewareWithCache(store, cache, tenantmw.MiddlewareConfig{
		AllowHeaderHint: false,
		CacheTTL:        5 * time.Minute,
	})

	app.Use(mw.Handle())
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	req, err := http.NewRequest(http.MethodGet, "http://suspended.localhost/", http.NoBody)
	assert.NoError(t, err)

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestTenantMiddleware_InactiveTenantIsNotCached(t *testing.T) {
	app := fiber.New()
	cache := newMockCache()
	store := &mockTenantStore{
		tenantsByHost: make(map[string]*platform.Tenant),
	}
	store.addTenant("suspended.localhost", newActiveTenant("tenant-1", "suspended", "suspended.localhost", "suspended"))

	mw := tenantmw.NewMiddlewareWithCache(store, cache, tenantmw.MiddlewareConfig{
		AllowHeaderHint: false,
		CacheTTL:        5 * time.Minute,
	})

	app.Use(mw.Handle())
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	for i := 0; i < 2; i++ {
		req, err := http.NewRequest(http.MethodGet, "http://suspended.localhost/", http.NoBody)
		assert.NoError(t, err)

		resp, err := app.Test(req)
		defer closeResponseBody(t, resp)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	}

	assert.Equal(t, 2, store.callCount())
	assert.Equal(t, 0, cache.sets)
}

func TestTenantMiddleware_LocalDevHint(t *testing.T) {
	app := fiber.New()
	cache := newMockCache()
	store := &mockTenantStore{
		tenantsByHost: make(map[string]*platform.Tenant),
	}
	store.addTenant("mytenant.localhost", newActiveTenant("tenant-1", "mytenant", "mytenant.localhost", "active"))

	mw := tenantmw.NewMiddlewareWithCache(store, cache, tenantmw.MiddlewareConfig{
		AllowHeaderHint: true,
		CacheTTL:        5 * time.Minute,
	})

	app.Use(mw.Handle())
	app.Get("/", func(c *fiber.Ctx) error {
		current := tenantmw.FromContext(c)
		assert.NotNil(t, current)
		assert.Equal(t, "tenant-1", current.ID)

		return c.SendStatus(fiber.StatusNoContent)
	})

	req, err := http.NewRequest(http.MethodGet, "http://localhost/", http.NoBody)
	assert.NoError(t, err)
	req.Header.Set(apiheaders.HeaderTenant, "mytenant")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
}

func TestTenantMiddleware_InternalServiceHint(t *testing.T) {
	app := fiber.New()
	cache := newMockCache()
	store := &mockTenantStore{
		tenantsByHost: make(map[string]*platform.Tenant),
	}
	store.addTenant("mytenant.localhost", newActiveTenant("tenant-1", "mytenant", "mytenant.localhost", "active"))

	mw := tenantmw.NewMiddlewareWithCache(store, cache, tenantmw.MiddlewareConfig{
		AllowHeaderHint: true,
		CacheTTL:        5 * time.Minute,
	})

	app.Use(mw.Handle())
	app.Get("/", func(c *fiber.Ctx) error {
		current := tenantmw.FromContext(c)
		assert.NotNil(t, current)
		assert.Equal(t, "tenant-1", current.ID)

		return c.SendStatus(fiber.StatusNoContent)
	})

	req, err := http.NewRequest(http.MethodGet, "http://control-plane:4000/", http.NoBody)
	assert.NoError(t, err)
	req.Header.Set(apiheaders.LegacyHeaderTenant, "mytenant")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
}

func TestTenantMiddleware_BypassesWalletTopUpCallback(t *testing.T) {
	app := fiber.New()
	cache := newMockCache()
	store := &mockTenantStore{
		tenantsByHost: make(map[string]*platform.Tenant),
	}

	mw := tenantmw.NewMiddlewareWithCache(store, cache, tenantmw.MiddlewareConfig{
		AllowHeaderHint: true,
		CacheTTL:        5 * time.Minute,
	})

	app.Use(mw.Handle())
	app.Get("/api/v1/wallet/topup/callback/:provider", func(c *fiber.Ctx) error {
		assert.Nil(t, tenantmw.FromContext(c))
		return c.SendStatus(fiber.StatusNoContent)
	})

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:4000/api/v1/wallet/topup/callback/mock?authority=mock_123&status=ok", http.NoBody)
	assert.NoError(t, err)

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
	assert.Empty(t, store.tenantsByHost)
	assert.Equal(t, 0, cache.gets)
	assert.Equal(t, 0, cache.sets)
}

func TestTenantMiddleware_NoCache_StillWorks(t *testing.T) {
	app := fiber.New()
	store := &mockTenantStore{
		tenantsByHost: make(map[string]*platform.Tenant),
	}
	store.addTenant("acme.localhost", newActiveTenant("tenant-1", "acme", "acme.localhost", "active"))

	mw := tenantmw.NewMiddlewareWithCache(store, nil, tenantmw.MiddlewareConfig{})

	app.Use(mw.Handle())
	app.Get("/", func(c *fiber.Ctx) error {
		current := tenantmw.FromContext(c)
		assert.NotNil(t, current)
		assert.Equal(t, "tenant-1", current.ID)

		return c.SendStatus(fiber.StatusNoContent)
	})

	req, err := http.NewRequest(http.MethodGet, "http://acme.localhost/", http.NoBody)
	assert.NoError(t, err)

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
}

func TestTenantMiddleware_InvalidateCache(t *testing.T) {
	cache := newMockCache()
	store := &mockTenantStore{
		tenantsByHost: make(map[string]*platform.Tenant),
	}
	store.addTenant("acme.localhost", newActiveTenant("tenant-1", "acme", "acme.localhost", "active"))

	mw := tenantmw.NewMiddlewareWithCache(store, cache, tenantmw.MiddlewareConfig{
		AllowHeaderHint: false,
		CacheTTL:        5 * time.Minute,
	})

	// Simulate caching with keys matching what the middleware uses
	ctx := context.Background()
	cache.Set(ctx, "tenant:host:acme", newActiveTenant("tenant-1", "acme", "acme.localhost", "active"), 5*time.Minute)
	cache.Set(ctx, "tenant:id:tenant-1", newActiveTenant("tenant-1", "acme", "acme.localhost", "active"), 5*time.Minute)

	// Invalidate
	mw.InvalidateTenant(ctx, "acme", "tenant-1")

	// Verify deleted by checking internal data directly
	cache.mu.Lock()
	_, foundHost := cache.data["tenant:host:acme"]
	_, foundID := cache.data["tenant:id:tenant-1"]
	cache.mu.Unlock()
	assert.Equal(t, 2, cache.deletes)
	assert.False(t, foundHost, "host cache should be deleted")
	assert.False(t, foundID, "id cache should be deleted")
}

func TestExtractSubdomain_WithSubdomain(t *testing.T) {
	assert.Equal(t, "acme", tenantmw.ExtractSubdomain("acme.shahrah.ir"))
	assert.Equal(t, "my-company", tenantmw.ExtractSubdomain("my-company.shahrah.ir"))
	assert.Equal(t, "test", tenantmw.ExtractSubdomain("test.api.example.com"))
}

func TestExtractSubdomain_WithPort(t *testing.T) {
	assert.Equal(t, "acme", tenantmw.ExtractSubdomain("acme.shahrah.ir:8080"))
	assert.Equal(t, "acme", tenantmw.ExtractSubdomain("acme.shahrah.ir:443"))
}

func TestExtractSubdomain_NoSubdomain(t *testing.T) {
	assert.Equal(t, "", tenantmw.ExtractSubdomain("localhost"))
	assert.Equal(t, "", tenantmw.ExtractSubdomain("127.0.0.1"))
	assert.Equal(t, "", tenantmw.ExtractSubdomain("shahrah.ir"))
	assert.Equal(t, "", tenantmw.ExtractSubdomain(""))
}

func TestExtractSubdomain_IPAddress(t *testing.T) {
	assert.Equal(t, "", tenantmw.ExtractSubdomain("192.168.1.1"))
	assert.Equal(t, "", tenantmw.ExtractSubdomain("10.0.0.1:3000"))
}

func TestValidateTenantSlug_Valid(t *testing.T) {
	assert.True(t, tenantmw.ValidateTenantSlug("acme"))
	assert.True(t, tenantmw.ValidateTenantSlug("my-company"))
	assert.True(t, tenantmw.ValidateTenantSlug("test123"))
	assert.True(t, tenantmw.ValidateTenantSlug("a"))
}

func TestValidateTenantSlug_Invalid(t *testing.T) {
	assert.False(t, tenantmw.ValidateTenantSlug(""))                                                                                   // Empty
	assert.False(t, tenantmw.ValidateTenantSlug("-acme"))                                                                              // Starts with hyphen
	assert.False(t, tenantmw.ValidateTenantSlug("acme-"))                                                                              // Ends with hyphen
	assert.False(t, tenantmw.ValidateTenantSlug("ACME"))                                                                               // Uppercase
	assert.False(t, tenantmw.ValidateTenantSlug("acme_company"))                                                                       // Underscore
	assert.False(t, tenantmw.ValidateTenantSlug("default"))                                                                            // Reserved
	assert.False(t, tenantmw.ValidateTenantSlug("www"))                                                                                // Reserved
	assert.False(t, tenantmw.ValidateTenantSlug("this-is-a-very-long-slug-that-exceeds-the-maximum-length-of-sixty-three-characters")) // Too long
}

func TestFromContext_NilLocals(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		assert.Nil(t, tenantmw.FromContext(c))
		return c.SendStatus(fiber.StatusNoContent)
	})

	req, err := http.NewRequest(http.MethodGet, "/", http.NoBody)
	assert.NoError(t, err)

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
}

func TestFromContext_WrongType(t *testing.T) {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(tenantmw.ContextKey{}, "not-a-tenant")
		return c.Next()
	})
	app.Get("/", func(c *fiber.Ctx) error {
		assert.Nil(t, tenantmw.FromContext(c))
		return c.SendStatus(fiber.StatusNoContent)
	})

	req, err := http.NewRequest(http.MethodGet, "/", http.NoBody)
	assert.NoError(t, err)

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
}
