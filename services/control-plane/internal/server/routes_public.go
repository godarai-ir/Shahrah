package server

import (
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/config"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/profiling"
	prommetrics "github.com/Delpak/shahrah/services/control-plane/internal/observability/prometheus"
	"github.com/gofiber/fiber/v2"
)

func registerPublicRoutes(app *fiber.App, cfg *config.Config, deps Dependencies, httpMetrics *prommetrics.HTTPMetrics) {
	registerObservabilityRoutes(app, cfg, deps, httpMetrics)
	if deps.VideoHandler != nil {
		app.Post("/webhooks/videos/:provider", deps.VideoHandler.Webhook)
	}
	registerPublicAdminRoutes(app, deps)
	registerPlatformAdminRoutes(app, cfg, deps)
}

func registerObservabilityRoutes(app *fiber.App, cfg *config.Config, deps Dependencies, httpMetrics *prommetrics.HTTPMetrics) {
	app.Get("/health", healthHandler)
	app.Get("/ready", readyHandler(deps))
	app.Get("/metrics", httpMetrics.Handler())

	if cfg != nil && cfg.Profiling.EnablePprof {
		profiling.RegisterPprofRoutes(app)
	}
}

func registerPublicAdminRoutes(app *fiber.App, deps Dependencies) {
	if deps.AdminHandlers == nil {
		return
	}

	app.Get("/api/v1/login-discovery/tenants", deps.AdminHandlers.DiscoverTenantsByIdentifier)
	app.Get("/api/v1/login-discovery/tenant/:slug", deps.AdminHandlers.DiscoverTenantByName)
	app.Get("/api/v1/commercial-plans", deps.AdminHandlers.ListPublicCommercialPlans)
	app.Get("/api/v1/subscription/payment/callback/:provider", deps.AdminHandlers.SubscriptionPaymentCallback)
	app.Post("/api/v1/subscription/payment/callback/:provider", deps.AdminHandlers.SubscriptionPaymentCallback)
}

func registerPlatformAdminRoutes(app *fiber.App, cfg *config.Config, deps Dependencies) {
	if deps.AdminHandlers == nil {
		return
	}

	platformAdmin := app.Group("/api/v1/platform", auth.PlatformAdminAuthMiddleware(cfg.GeneralSettings.MasterKey, deps.PlatformServiceAccounts))
	platformAdmin.Get("/features", deps.AdminHandlers.ListPlatformFeatureCatalog)
	platformAdmin.Get("/commercial-plans", deps.AdminHandlers.ListPlatformCommercialPlans)
	platformAdmin.Get("/commercial-signup-settings", deps.AdminHandlers.GetPlatformCommercialSignupSettings)
	platformAdmin.Put("/commercial-signup-settings", deps.AdminHandlers.UpdatePlatformCommercialSignupSettings)
	platformAdmin.Post("/commercial-plans", deps.AdminHandlers.CreatePlatformCommercialPlan)
	platformAdmin.Patch("/commercial-plans/:slug", deps.AdminHandlers.UpdatePlatformCommercialPlan)
	platformAdmin.Delete("/commercial-plans/:slug", deps.AdminHandlers.DeactivatePlatformCommercialPlan)
	platformAdmin.Post("/commercial-plans/:slug/versions", deps.AdminHandlers.CreatePlatformPlanVersion)
	platformAdmin.Post("/commercial-plans/:slug/versions/:version/publish", deps.AdminHandlers.PublishPlatformPlanVersion)
	platformAdmin.Post("/commercial-plans/:slug/versions/:version/entitlements", deps.AdminHandlers.UpsertPlatformPlanEntitlement)
	platformAdmin.Put("/commercial-plans/:slug/versions/:version/entitlements/:featureKey", deps.AdminHandlers.UpsertPlatformPlanEntitlement)
	platformAdmin.Post("/commercial-plans/:slug/versions/:version/prices", deps.AdminHandlers.UpsertPlatformPlanPrice)
	platformAdmin.Put("/commercial-plans/:slug/versions/:version/prices/:interval", deps.AdminHandlers.UpsertPlatformPlanPrice)
	platformAdmin.Put("/tenant-subscriptions", deps.AdminHandlers.UpsertPlatformTenantSubscription)
	platformAdmin.Put("/tenant-entitlement-overrides", deps.AdminHandlers.UpsertPlatformTenantEntitlementOverride)
	platformAdmin.Delete("/tenant-entitlement-overrides/:tenantID/:featureKey", deps.AdminHandlers.DeletePlatformTenantEntitlementOverride)
	platformAdmin.Get("/tenants", deps.AdminHandlers.ListPlatformTenants)
	platformAdmin.Get("/tenants/:tenantID", deps.AdminHandlers.GetPlatformTenant)
	platformAdmin.Get("/tenants/:tenantID/events", deps.AdminHandlers.ListPlatformTenantEvents)
	platformAdmin.Get("/tenants/:tenantID/wallet", deps.AdminHandlers.GetPlatformTenantWallet)
	platformAdmin.Get("/tenants/:tenantID/wallet/transactions", deps.AdminHandlers.ListPlatformTenantWalletTransactions)
	platformAdmin.Post("/tenants/:tenantID/wallet/adjustments", deps.AdminHandlers.AdjustPlatformTenantWallet)
	platformAdmin.Post("/tenants/:tenantID/wallet/bank-transfers/:requestID/approve", deps.AdminHandlers.ApprovePlatformBankTransfer)
	platformAdmin.Post("/tenants/:tenantID/wallet/bank-transfers/:requestID/reject", deps.AdminHandlers.RejectPlatformBankTransfer)
	platformAdmin.Post("/tenants/:tenantID/suspend", deps.AdminHandlers.SuspendPlatformTenant)
	platformAdmin.Post("/tenants/:tenantID/reactivate", deps.AdminHandlers.ReactivatePlatformTenant)
	platformAdmin.Post("/tenants/:tenantID/subscription/cancel-at-period-end", deps.AdminHandlers.SchedulePlatformTenantSubscriptionCancellation)
	platformAdmin.Delete("/tenants/:tenantID/subscription/cancel-at-period-end", deps.AdminHandlers.ReversePlatformTenantSubscriptionCancellation)
	platformAdmin.Get("/tenants/:tenantID/billing/payment-intents", deps.AdminHandlers.ListPlatformTenantSubscriptionPaymentIntents)
	platformAdmin.Get("/provider-connection-schemas/:provider", deps.AdminHandlers.GetPlatformProviderConnectionSchemas)
	platformAdmin.Get("/provider-connections", deps.AdminHandlers.ListPlatformProviderConnections)
	platformAdmin.Get("/provider-connections/:id", deps.AdminHandlers.GetPlatformProviderConnection)
	platformAdmin.Post("/provider-connections", deps.AdminHandlers.CreatePlatformProviderConnection)
	platformAdmin.Patch("/provider-connections/:id", deps.AdminHandlers.UpdatePlatformProviderConnection)
	platformAdmin.Post("/provider-connections/:id/credential-versions", deps.AdminHandlers.CreatePlatformProviderCredentialVersion)
	platformAdmin.Post("/provider-connections/:id/credential-versions/:version/test", deps.AdminHandlers.TestPlatformProviderCredentialVersion)
	platformAdmin.Post("/provider-connections/:id/credential-versions/:version/activate", deps.AdminHandlers.ActivatePlatformProviderCredentialVersion)
	platformAdmin.Post("/provider-connections/:id/credential-versions/:version/rollback", deps.AdminHandlers.RollbackPlatformProviderCredentialVersion)
	platformAdmin.Post("/provider-connections/:id/disable", deps.AdminHandlers.DisablePlatformProviderConnection)
	platformAdmin.Post("/provider-connections/:id/credential-versions/:version/revoke", deps.AdminHandlers.RevokePlatformProviderCredentialVersion)
}
