package tracing

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/apiheaders"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type privacyResolverCounter struct {
	calls atomic.Int64
}

func (r *privacyResolverCounter) ResolveTenantPolicy(string) (CapturePolicy, bool) {
	r.calls.Add(1)

	return CapturePolicySampledRedacted, true
}

func TestMiddleware_UsesResolvedTenantContextForRootSpan(t *testing.T) {
	app := fiber.New()
	queue := NewQueue(16)
	sampler := NewSampler(1.0, 1500)

	app.Use(Middleware(queue, sampler, 1500))
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(tenant.ContextKey{}, &platform.Tenant{
			ID:     "10f8c4d0-b67d-47bc-bb6c-88172b7c6ec6",
			Slug:   "tenant1",
			Status: "active",
		})

		return c.Next()
	})
	app.Get("/admin", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/admin", http.NoBody)
	req.Header.Set(apiheaders.HeaderTenant, "tenant1")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)

	batch := queue.DequeueBatch(10)
	require.Len(t, batch, 1)
	assert.Equal(t, "10f8c4d0-b67d-47bc-bb6c-88172b7c6ec6", batch[0].TenantID)
}

func TestRecordStepSpan_UsesResolvedTenantContext(t *testing.T) {
	app := fiber.New()
	queue := NewQueue(16)
	sampler := NewSampler(1.0, 1500)

	app.Use(Middleware(queue, sampler, 1500))
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(tenant.ContextKey{}, &platform.Tenant{
			ID:     "10f8c4d0-b67d-47bc-bb6c-88172b7c6ec6",
			Slug:   "tenant1",
			Status: "active",
		})

		return c.Next()
	})
	app.Get("/admin", func(c *fiber.Ctx) error {
		start := time.Now().Add(-5 * time.Millisecond)
		RecordStepSpan(c, QueueFromContext(c), "", "admin.query", "internal", start, fiber.StatusOK, map[string]interface{}{
			"component": "tenant-admin",
		})

		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/admin", http.NoBody)
	req.Header.Set(apiheaders.HeaderTenant, "tenant1")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)

	batch := queue.DequeueBatch(10)
	require.Len(t, batch, 2)
	assert.Equal(t, "10f8c4d0-b67d-47bc-bb6c-88172b7c6ec6", batch[0].TenantID)
	assert.Equal(t, "10f8c4d0-b67d-47bc-bb6c-88172b7c6ec6", batch[1].TenantID)
	assert.Equal(t, "gateway.request", batch[0].Name)
	assert.Equal(t, "admin.query", batch[1].Name)
}

func TestRecordStepSpanDefersPrivacyWorkForUnsampledRequest(t *testing.T) {
	app := fiber.New()
	queue := NewQueue(16)
	privacy := NewPrivacyManager(DefaultConfig())
	resolver := &privacyResolverCounter{}
	privacy.SetTenantPolicyResolver(resolver)

	app.Use(Middleware(queue, NewSampler(0, 60_000), 60_000, privacy))
	app.Get("/v1/models", func(c *fiber.Ctx) error {
		SetTenantForTrace(c, "tenant-1")
		RecordStepSpan(
			c,
			QueueFromContext(c),
			"",
			"models.list",
			"internal",
			time.Now(),
			fiber.StatusOK,
			map[string]interface{}{payloadRequestKey: map[string]interface{}{"api_key": "secret"}},
		)

		return c.SendStatus(fiber.StatusOK)
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/v1/models", http.NoBody))
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Zero(t, resolver.calls.Load())
	assert.Zero(t, queue.Len())
}
