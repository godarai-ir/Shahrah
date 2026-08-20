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
	registerPublicAdminRoutes(app, deps)
	registerPlatformAdminRoutes(app, cfg, deps)

	if deps.RAGHandler != nil {
		app.Post("/v1/chatbots/:publicID/responses", deps.RAGHandler.ForwardRuntime)
		app.Delete("/v1/chatbots/:publicID/sessions/:sessionID", deps.RAGHandler.ForwardRuntime)
		app.Get("/v1/chatbots/:publicID/openai/models", deps.RAGHandler.ForwardRuntime)
		app.Post("/v1/chatbots/:publicID/openai/chat/completions", deps.RAGHandler.ForwardRuntime)
	}
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
	if deps.ExchangeRates != nil {
		app.Get("/api/v1/exchange-rates/current", currentExchangeRateHandler(deps.ExchangeRates))
	}

	if deps.ManagementRoutes == nil {
		return
	}

	deps.ManagementRoutes.RegisterPublic(app)
}

func registerPlatformAdminRoutes(app *fiber.App, cfg *config.Config, deps Dependencies) {
	if deps.ManagementRoutes == nil {
		return
	}

	platformAdmin := app.Group("/api/v1/platform", auth.PlatformAdminAuthMiddleware(cfg.GeneralSettings.MasterKey, deps.PlatformServiceAccounts))
	deps.ManagementRoutes.RegisterPlatform(platformAdmin)
}
