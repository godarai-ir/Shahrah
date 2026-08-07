package tenant

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

type tenantStoreCounter struct {
	mu    sync.Mutex
	calls int
}

func (s *tenantStoreCounter) GetTenantByHost(context.Context, string) (*platform.Tenant, error) {
	s.mu.Lock()
	s.calls++
	s.mu.Unlock()

	return &platform.Tenant{
		ID:     "tenant_a",
		Slug:   "acme",
		Domain: "acme.example",
		Status: "active",
	}, nil
}

type tenantCacheCounter struct {
	mu      sync.Mutex
	reads   int
	deletes int
	deleted []string
}

func (c *tenantCacheCounter) Get(context.Context, string) (interface{}, bool) {
	c.mu.Lock()
	c.reads++
	c.mu.Unlock()

	return nil, false
}

func (*tenantCacheCounter) Set(context.Context, string, interface{}, time.Duration) {}

func (c *tenantCacheCounter) Delete(_ context.Context, key string) {
	c.mu.Lock()
	c.deletes++
	c.deleted = append(c.deleted, key)
	c.mu.Unlock()
}

func TestMiddlewareLocalCacheAvoidsDistributedLookupAndSupportsInvalidation(t *testing.T) {
	store := &tenantStoreCounter{}
	sharedCache := &tenantCacheCounter{}
	middleware := NewMiddlewareWithCache(store, sharedCache, MiddlewareConfig{
		CacheTTL:      time.Minute,
		LocalCacheTTL: time.Minute,
	})
	app := fiber.New()
	app.Use(middleware.Handle())
	app.Get("/v1/models", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	request := func() {
		t.Helper()

		req := httptest.NewRequest("GET", "http://acme.example/v1/models", http.NoBody)
		response, err := app.Test(req)
		require.NoError(t, err)
		require.NoError(t, response.Body.Close())
		require.Equal(t, fiber.StatusNoContent, response.StatusCode)
	}

	request()
	request()

	store.mu.Lock()
	storeCalls := store.calls
	store.mu.Unlock()
	sharedCache.mu.Lock()
	cacheReads := sharedCache.reads
	sharedCache.mu.Unlock()

	require.Equal(t, 1, storeCalls)
	require.Equal(t, 1, cacheReads)

	middleware.InvalidateTenant(t.Context(), "acme.example", "tenant_a")
	request()

	store.mu.Lock()
	storeCalls = store.calls
	store.mu.Unlock()
	require.Equal(t, 2, storeCalls)
}

func TestMiddlewareLocalCacheIsBounded(t *testing.T) {
	middleware := NewMiddlewareWithCache(&tenantStoreCounter{}, nil, MiddlewareConfig{
		CacheTTL:             time.Minute,
		LocalCacheTTL:        time.Minute,
		LocalCacheMaxEntries: 2,
	})

	for _, host := range []string{"one.example", "two.example", "three.example"} {
		middleware.rememberLocalTenant(host, &platform.Tenant{ID: host, Status: "active"})
	}

	middleware.localMu.RLock()
	entries := len(middleware.localTenants)
	middleware.localMu.RUnlock()

	require.Equal(t, 2, entries)
}

func TestMiddlewareInvalidationRemovesEveryCachedAliasForTenant(t *testing.T) {
	sharedCache := &tenantCacheCounter{}
	middleware := NewMiddlewareWithCache(&tenantStoreCounter{}, sharedCache, MiddlewareConfig{
		CacheTTL:      time.Minute,
		LocalCacheTTL: time.Minute,
	})
	tenant := &platform.Tenant{ID: "tenant_a", Status: "active"}
	middleware.rememberLocalTenant("acme.example", tenant)
	middleware.rememberLocalTenant("custom.example", tenant)

	middleware.InvalidateTenant(t.Context(), "acme", tenant.ID)

	middleware.localMu.RLock()
	entries := len(middleware.localTenants)
	middleware.localMu.RUnlock()

	require.Zero(t, entries)

	sharedCache.mu.Lock()
	deleted := append([]string(nil), sharedCache.deleted...)
	sharedCache.mu.Unlock()

	require.ElementsMatch(t, []string{
		cacheKeyForHost("acme"),
		cacheKeyForHost("acme.example"),
		cacheKeyForHost("custom.example"),
		cacheKeyForID(tenant.ID),
	}, deleted)
}
