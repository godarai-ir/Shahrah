package server

import (
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/admin"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/config"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
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
	TracingMiddleware  fiber.Handler
	TraceHealth        TraceHealthReporter
	AdminHandlers      *admin.Handlers
	AuthHandlers       *admin.AuthHandlers
	ProxyHandler       *proxy.ChatCompletionHandler
	CompletionsHandler *proxy.CompletionsHandler
	ModelsHandler      *proxy.ModelsHandler
	Logger             *zap.Logger
}

type TraceHealthReporter interface {
	ReadinessReport() map[string]interface{}
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
		TimeFormat: "2006-01-02T15:04:05.000Z07:00",
		TimeZone:   "UTC",
		Format:     "ts=${time} status=${status} latency=${latency} ip=${ip} method=${method} path=${path} trace_id=${respHeader:" + tracing.TraceIDHeader + "} ua=\"${ua}\" bytes_in=${bytesReceived} bytes_out=${bytesSent} req_headers=\"${reqHeaders}\" req_body=\"${body}\" resp_headers=\"${respHeaders}\" resp_body=\"${resBody}\" error=\"${error}\"\n",
	}))
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET, POST, PUT, PATCH, DELETE, OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, api-key, x-api-key, x-shahrah-tenant",
	}))
	if deps.TracingMiddleware != nil {
		app.Use(deps.TracingMiddleware)
	}
	app.Use(deps.TenantResolver.Handle())
	app.Use(deps.AuthMiddleware.Middleware())

	registerRoutes(app, deps)
	return app
}

func registerRoutes(app *fiber.App, deps Dependencies) {
	app.Get("/health", healthHandler)
	app.Get("/ready", readyHandler(deps))

	v1 := app.Group("/v1")
	v1.Post("/chat/completions", deps.ProxyHandler.Handle)
	v1.Post("/completions", deps.CompletionsHandler.Handle)
	v1.Get("/models", deps.ModelsHandler.Handle)

	api := app.Group("/api/v1")
	api.Get("/onboarding/tenant-slugs/:slug/availability", deps.AdminHandlers.CheckTenantSlugAvailability)
	api.Post("/onboarding/otp", deps.AdminHandlers.RequestOnboardingOTP)
	api.Post("/onboarding/otp/verify", deps.AdminHandlers.VerifyOnboardingOTP)
	api.Post("/keys", deps.AdminHandlers.CreateKey)
	api.Post("/keys/pat", deps.AdminHandlers.CreatePATKey)
	api.Post("/keys/vat", deps.AdminHandlers.CreateVATKey)
	api.Get("/keys", deps.AdminHandlers.ListKeys)
	api.Get("/personal-access-tokens", deps.AdminHandlers.ListPersonalAccessTokens)
	api.Post("/personal-access-tokens", deps.AdminHandlers.CreatePersonalAccessToken)
	api.Get("/personal-access-tokens/exists", deps.AdminHandlers.PersonalAccessTokenExists)
	api.Delete("/personal-access-tokens/:id", deps.AdminHandlers.RevokePersonalAccessToken)
	api.Delete("/personal-access-tokens/revoke/all", deps.AdminHandlers.RevokeAllPersonalAccessTokens)
	api.Get("/org-units", deps.AdminHandlers.ListOrgUnits)
	api.Post("/org-units", deps.AdminHandlers.CreateOrgUnit)
	api.Get("/org-units/:orgUnitID", deps.AdminHandlers.GetOrgUnit)
	api.Patch("/org-units/:orgUnitID", deps.AdminHandlers.UpdateOrgUnit)
	api.Delete("/org-units/:orgUnitID", deps.AdminHandlers.DeleteOrgUnit)
	api.Get("/org-units/:orgUnitID/members", deps.AdminHandlers.ListOrgUnitMembers)
	api.Post("/org-units/:orgUnitID/members", deps.AdminHandlers.AddOrgUnitMember)
	api.Patch("/org-units/:orgUnitID/members/:userID", deps.AdminHandlers.UpdateOrgUnitMember)
	api.Delete("/org-units/:orgUnitID/members/:userID", deps.AdminHandlers.DeleteOrgUnitMember)
	api.Get("/teams", deps.AdminHandlers.ListTeams)
	api.Post("/teams", deps.AdminHandlers.CreateTeam)
	api.Get("/teams/:teamID", deps.AdminHandlers.GetTeam)
	api.Patch("/teams/:teamID", deps.AdminHandlers.UpdateTeam)
	api.Delete("/teams/:teamID", deps.AdminHandlers.DeleteTeam)
	api.Get("/teams/:teamID/members", deps.AdminHandlers.ListTeamMembers)
	api.Post("/teams/:teamID/members", deps.AdminHandlers.AddTeamMember)
	api.Patch("/teams/:teamID/members/:userID", deps.AdminHandlers.UpdateTeamMember)
	api.Delete("/teams/:teamID/members/:userID", deps.AdminHandlers.DeleteTeamMember)
	api.Get("/teams/:teamID/permissions", deps.AdminHandlers.GetTeamPermissions)
	api.Patch("/teams/:teamID/permissions", deps.AdminHandlers.UpdateTeamPermissions)
	api.Get("/virtual-users", deps.AdminHandlers.ListVirtualUsers)
	api.Post("/virtual-users", deps.AdminHandlers.CreateVirtualUser)
	api.Get("/virtual-users/:virtualUserID", deps.AdminHandlers.GetVirtualUser)
	api.Patch("/virtual-users/:virtualUserID", deps.AdminHandlers.UpdateVirtualUser)
	api.Delete("/virtual-users/:virtualUserID", deps.AdminHandlers.DeleteVirtualUser)
	api.Post("/models", deps.AdminHandlers.CreateModel)
	api.Get("/models", deps.AdminHandlers.ListModels)
	api.Post("/virtual-models", deps.AdminHandlers.CreateVirtualModel)
	api.Get("/virtual-models", deps.AdminHandlers.ListVirtualModels)
	api.Get("/virtual-models/:id", deps.AdminHandlers.GetVirtualModel)
	api.Patch("/virtual-models/:id", deps.AdminHandlers.UpdateVirtualModel)
	api.Get("/provider-catalog/providers", deps.AdminHandlers.ListProviderCatalog)
	api.Post("/provider-catalog/refresh", deps.AdminHandlers.RefreshProviderCatalogCache)
	api.Get("/provider-catalog/stats", deps.AdminHandlers.GetProviderCatalogStats)
	api.Get("/provider-catalog/providers/:provider/models", deps.AdminHandlers.ListProviderModelsCatalog)
	api.Get("/access-principals", deps.AdminHandlers.ListAccessPrincipals)
	api.Get("/users", deps.AdminHandlers.ListTenantUsers)
	api.Post("/users/invite", deps.AdminHandlers.InviteTenantUser)
	api.Post("/users/invitations", deps.AdminHandlers.InviteTenantUser)
	api.Post("/users/invitations/accept", deps.AdminHandlers.AcceptTenantUserInvitation)
	api.Post("/users/:userID/activate", deps.AdminHandlers.ActivateTenantUser)
	api.Post("/users/:userID/deactivate", deps.AdminHandlers.DeactivateTenantUser)
	api.Post("/tenant-provider-accounts", deps.AdminHandlers.CreateTenantProviderAccount)
	api.Get("/tenant-provider-accounts", deps.AdminHandlers.ListTenantProviderAccounts)
	api.Get("/tenant-provider-accounts/:accountID", deps.AdminHandlers.GetTenantProviderAccount)
	api.Get("/tenant-provider-accounts/:accountID/effective-models", deps.AdminHandlers.GetTenantProviderAccountEffectiveModels)
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
	api.Get("/traces", deps.AdminHandlers.ListRequestTraces)
	api.Get("/traces/:traceID", deps.AdminHandlers.GetRequestTrace)
	api.Get("/traces/:traceID/spans", deps.AdminHandlers.ListRequestTraceSpans)
	api.Post("/metrics/query", deps.AdminHandlers.QueryModelMetrics)

	// Tenant-admin auth decoupling endpoints
	if deps.AuthHandlers != nil {
		authGroup := api.Group("/auth")
		authGroup.Post("/captcha/create", deps.AuthHandlers.CreateCaptcha)
		authGroup.Post("/captcha/consume", deps.AuthHandlers.ConsumeCaptcha)
		authGroup.Post("/otp/request", deps.AuthHandlers.CreateOTPRequest)
		authGroup.Post("/otp/verify", deps.AuthHandlers.VerifyOTPRequest)
		authGroup.Post("/session", deps.AuthHandlers.CreateSession)
		authGroup.Get("/session/:id", deps.AuthHandlers.GetSession)
		authGroup.Delete("/session/:id", deps.AuthHandlers.RevokeSession)
		authGroup.Post("/user/find-or-create", deps.AuthHandlers.FindOrCreateUser)
		authGroup.Get("/user/:id", deps.AuthHandlers.GetUser)
		authGroup.Post("/user/:id/update-contact", deps.AuthHandlers.UpdateUserContact)
		authGroup.Post("/user/:id/touch-last-login", deps.AuthHandlers.TouchLastLogin)
	}
}

func healthHandler(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "healthy",
		"version": "0.2.0",
	})
}

func readyHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		resp := fiber.Map{
			"status":  "ready",
			"version": "0.2.0",
		}
		if deps.TraceHealth != nil {
			resp["tracing"] = deps.TraceHealth.ReadinessReport()
		}
		return c.JSON(resp)
	}
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
