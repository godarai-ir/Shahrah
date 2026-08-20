package server

import "github.com/gofiber/fiber/v2"

func registerProtectedRoutes(app *fiber.App, deps Dependencies) {
	registerProxyRoutes(app, deps)
	registerTenantAdminRoutes(app, deps)
}

func registerTenantAdminRoutes(app *fiber.App, deps Dependencies) {
	if deps.ManagementRoutes == nil {
		return
	}

	deps.ManagementRoutes.RegisterTenant(app.Group("/api/v1"), app.Group("/api/v2"))
	deps.ManagementRoutes.RegisterAuth(app.Group("/api/v1/auth"))

	if deps.RAGHandler != nil {
		api := app.Group("/api/v1")
		api.All("/knowledge-bases", deps.RAGHandler.ForwardManagement)
		api.All("/knowledge-bases/*", deps.RAGHandler.ForwardManagement)
		api.All("/chatbots", deps.RAGHandler.ForwardManagement)
		api.All("/chatbots/*", deps.RAGHandler.ForwardManagement)
	}
}
