package tenant_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	tenantmw "github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

type tenantStoreStub struct {
	tenant *platform.Tenant
}

func (s tenantStoreStub) GetTenantByHost(ctx context.Context, host string) (*platform.Tenant, error) {
	return s.tenant, nil
}

func (s tenantStoreStub) EnsureTenant(ctx context.Context, slug string) (*platform.Tenant, error) {
	return s.tenant, nil
}

func TestTenantMiddleware_ShouldAttachTenant(t *testing.T) {
	app := fiber.New()
	mw := tenantmw.NewMiddleware(tenantStoreStub{
		tenant: &platform.Tenant{
			ID:        "tenant-1",
			Slug:      "default",
			Status:    "active",
			CreatedAt: time.Now(),
		},
	}, "default", true)

	app.Use(mw.Handle())
	app.Get("/", func(c *fiber.Ctx) error {
		current := tenantmw.FromContext(c)
		assert.NotNil(t, current)
		assert.Equal(t, "tenant-1", current.ID)
		return c.SendStatus(fiber.StatusNoContent)
	})

	req, err := http.NewRequest(http.MethodGet, "http://localhost/", nil)
	assert.NoError(t, err)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
}
