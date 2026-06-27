package proxy_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/proxy"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockDeploymentLister struct {
	deployments   []platform.Deployment
	virtualModels []platform.VirtualModel
}

func (m mockDeploymentLister) ListDeployments(ctx context.Context, tenantID string) ([]platform.Deployment, error) {
	return m.deployments, nil
}

func (m mockDeploymentLister) ListVirtualModels(ctx context.Context, tenantID string) ([]platform.VirtualModel, error) {
	return m.virtualModels, nil
}

func TestModelsHandler_FiltersByAllowedModels(t *testing.T) {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(auth.AuthContextKey{}, &platform.AuthContext{
			TenantID:      "tenant-1",
			AllowedModels: []string{"gpt-4o-mini"},
		})

		return c.Next()
	})
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(tenant.ContextKey{}, &platform.Tenant{ID: "tenant-1", Slug: "acme", Status: "active"})
		return c.Next()
	})

	handler := proxy.NewModelsHandler(mockDeploymentLister{
		deployments: []platform.Deployment{
			{ID: "dep-1", ModelName: "gpt-4o-mini", Provider: "openai", Model: "gpt-4o-mini"},
			{ID: "dep-2", ModelName: "claude-sonnet", Provider: "anthropic", Model: "claude-sonnet"},
		},
		virtualModels: []platform.VirtualModel{
			{ID: "vm-1", Alias: "smart-router"},
		},
	})
	app.Get("/v1/models", handler.Handle)

	req := httptest.NewRequest("GET", "/v1/models", nil)

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var out contractsopenai.ModelListResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
	require.Len(t, out.Data, 1)
	assert.Equal(t, "gpt-4o-mini", out.Data[0].ID)
}

func TestModelsHandler_EnforcedAllowlistEmptyReturnsNoModels(t *testing.T) {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(auth.AuthContextKey{}, &platform.AuthContext{
			TenantID:              "tenant-1",
			AllowedModels:         []string{},
			EnforceModelAllowlist: true,
		})

		return c.Next()
	})
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(tenant.ContextKey{}, &platform.Tenant{ID: "tenant-1", Slug: "acme", Status: "active"})
		return c.Next()
	})

	handler := proxy.NewModelsHandler(mockDeploymentLister{
		deployments: []platform.Deployment{
			{ID: "dep-1", ModelName: "gpt-4o-mini", Provider: "openai", Model: "gpt-4o-mini"},
		},
	})
	app.Get("/v1/models", handler.Handle)

	req := httptest.NewRequest("GET", "/v1/models", nil)

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var out contractsopenai.ModelListResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
	assert.Len(t, out.Data, 0)
}
