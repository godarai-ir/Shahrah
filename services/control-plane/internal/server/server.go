package server

import (
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/admin"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/config"
	"github.com/Delpak/shahrah/services/control-plane/internal/proxy"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.uber.org/zap"
)

// Dependencies contains the concrete services required by the control-plane server.
type Dependencies struct {
	AuthMiddleware     *auth.AuthMiddleware
	TenantResolver     *tenant.Middleware
	AdminHandlers      *admin.Handlers
	ProxyHandler       *proxy.ChatCompletionHandler
	CompletionsHandler *proxy.CompletionsHandler
	ModelsHandler      *proxy.ModelsHandler
	Logger             *zap.Logger
}

// NewFiberApp creates and configures a new Fiber application.
func NewFiberApp(cfg *config.Config, deps Dependencies) *fiber.App {
	app := fiber.New(fiber.Config{
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
		ErrorHandler: customErrorHandler,
	})

	app.Use(fiberlogger.New(fiberlogger.Config{
		Format: "${time} | ${status} | ${latency} | ${ip} | ${method} | ${path}\n",
	}))
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET, POST, PUT, OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, api-key, x-api-key, x-shahrah-tenant",
	}))
	app.Use(deps.TenantResolver.Handle())
	app.Use(deps.AuthMiddleware.Middleware())

	registerRoutes(app, deps)
	return app
}

func registerRoutes(app *fiber.App, deps Dependencies) {
	app.Get("/health", healthHandler)
	app.Get("/ready", healthHandler)

	v1 := app.Group("/v1")
	v1.Post("/chat/completions", deps.ProxyHandler.Handle)
	v1.Post("/completions", deps.CompletionsHandler.Handle)
	v1.Get("/models", deps.ModelsHandler.Handle)

	api := app.Group("/api/v1")
	api.Post("/keys", deps.AdminHandlers.CreateKey)
	api.Get("/keys", deps.AdminHandlers.ListKeys)
	api.Post("/models", deps.AdminHandlers.CreateModel)
	api.Get("/models", deps.AdminHandlers.ListModels)
	api.Get("/provider-catalog/providers", deps.AdminHandlers.ListProviderCatalog)
	api.Get("/provider-catalog/providers/:provider/models", deps.AdminHandlers.ListProviderModelsCatalog)
	api.Get("/access-principals", deps.AdminHandlers.ListAccessPrincipals)
	api.Get("/users", deps.AdminHandlers.ListTenantUsers)
	api.Post("/users/invite", deps.AdminHandlers.InviteTenantUser)
	api.Post("/users/invitations/accept", deps.AdminHandlers.AcceptTenantUserInvitation)
	api.Post("/users/:userID/activate", deps.AdminHandlers.ActivateTenantUser)
	api.Post("/users/:userID/deactivate", deps.AdminHandlers.DeactivateTenantUser)
	api.Post("/tenant-provider-accounts", deps.AdminHandlers.CreateTenantProviderAccount)
	api.Get("/tenant-provider-accounts", deps.AdminHandlers.ListTenantProviderAccounts)
	api.Get("/tenant-provider-accounts/:accountID", deps.AdminHandlers.GetTenantProviderAccount)
	api.Put("/tenant-provider-accounts/:accountID/models", deps.AdminHandlers.UpsertTenantProviderAccountModels)
	api.Put("/tenant-provider-accounts/:accountID/access", deps.AdminHandlers.UpdateTenantProviderAccountAccess)
	api.Post("/tenant-provider-accounts/:accountID/test-credentials", deps.AdminHandlers.TestTenantProviderAccountCredentials)
	api.Get("/wallet", deps.AdminHandlers.GetWallet)
	api.Post("/wallet/topup", deps.AdminHandlers.TopUpWallet)
	api.Post("/wallet/topup/intents", deps.AdminHandlers.CreateWalletTopUpIntent)
	api.Get("/wallet/transactions", deps.AdminHandlers.ListWalletTransactions)
	api.Get("/wallet/topup/settings", deps.AdminHandlers.GetWalletTopUpSettings)
	api.Put("/wallet/topup/settings", deps.AdminHandlers.UpdateWalletTopUpSettings)
	api.Get("/wallet/topup/callback/:provider", deps.AdminHandlers.WalletTopUpCallback)
	api.Post("/wallet/topup/callback/:provider", deps.AdminHandlers.WalletTopUpCallback)
}

func healthHandler(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "healthy",
		"version": "0.2.0",
	})
}

func customErrorHandler(ctx *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	msg := "Internal Server Error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		msg = e.Message
	}

	return ctx.Status(code).JSON(fiber.Map{
		"error": map[string]interface{}{
			"message": msg,
			"type":    "server_error",
			"code":    code,
		},
	})
}
