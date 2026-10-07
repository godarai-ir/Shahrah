package server

import "github.com/gofiber/fiber/v2"

func registerProtectedRoutes(app *fiber.App, deps Dependencies) {
	if deps.ManagementRoutes != nil {
		deps.ManagementRoutes.RegisterOIDCProvider(app)
	}

	registerProxyRoutes(app, deps)
	registerTenantAdminRoutes(app, deps)
}

func registerTenantAdminRoutes(app *fiber.App, deps Dependencies) {
	if deps.ManagementRoutes == nil {
		return
	}

	deps.ManagementRoutes.RegisterTenant(app.Group("/api/v1"), app.Group("/api/v2"))
	deps.ManagementRoutes.RegisterAuth(app.Group("/api/v1/auth"))
}
