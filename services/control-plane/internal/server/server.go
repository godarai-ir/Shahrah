package server

import (
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/admin"
	"github.com/Delpak/shahrah/services/control-plane/internal/apiheaders"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/config"
	prommetrics "github.com/Delpak/shahrah/services/control-plane/internal/observability/prometheus"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/Delpak/shahrah/services/control-plane/internal/proxy"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.uber.org/zap"
)

var redactedLogHeaders = map[string]struct{}{
	"authorization":            {},
	"proxy-authorization":      {},
	"cookie":                   {},
	"set-cookie":               {},
	"x-api-key":                {},
	"api-key":                  {},
	"x-auth-token":             {},
	"x-forwarded-access-token": {},
}

// Dependencies contains the concrete services required by the control-plane server.
type Dependencies struct {
	AuthMiddleware          *auth.AuthMiddleware
	TenantResolver          *tenant.Middleware
	TracingMiddleware       fiber.Handler
	TraceHealth             TraceHealthReporter
	AdminHandlers           *admin.Handlers
	AuthHandlers            *admin.AuthHandlers
	ProxyHandler            *proxy.ChatCompletionHandler
	CompletionsHandler      *proxy.CompletionsHandler
	ResponsesHandler        *proxy.ResponsesHandler
	EmbeddingsHandler       *proxy.EmbeddingsHandler
	ModerationsHandler      *proxy.ModerationsHandler
	ImagesHandler           *proxy.ImagesHandler
	ModelsHandler           *proxy.ModelsHandler
	Logger                  *zap.Logger
	PlatformServiceAccounts auth.PlatformServiceAccountStore
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

	httpMetrics := prommetrics.NewHTTPMetrics()

	app.Use(fiberlogger.New(buildHTTPLoggerConfig(nil)))
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET, POST, PUT, PATCH, DELETE, OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, api-key, x-api-key, " + apiheaders.HeaderTenant + ", " + apiheaders.LegacyHeaderTenant + ", " + apiheaders.HeaderGuardrails + ", " + apiheaders.LegacyHeaderGuardrails + ", " + apiheaders.HeaderGuardrailsScope + ", " + apiheaders.LegacyHeaderGuardrailScope + ", x-tfy-guardrails, x-tfy-guardrails-scope",
	}))
	if deps.TracingMiddleware != nil {
		app.Use(deps.TracingMiddleware)
	}
	app.Use(httpMetrics.Middleware())

	registerPublicRoutes(app, cfg, deps, httpMetrics)

	if deps.TenantResolver != nil {
		app.Use(deps.TenantResolver.Handle())
	}
	if deps.AuthMiddleware != nil {
		app.Use(deps.AuthMiddleware.Middleware())
	}

	registerProtectedRoutes(app, deps)
	return app
}

func buildHTTPLoggerConfig(output io.Writer) fiberlogger.Config {
	cfg := fiberlogger.Config{
		TimeFormat:    "2006-01-02T15:04:05.000Z07:00",
		TimeZone:      "UTC",
		DisableColors: true,
		Format:        "ts=${time} status=${status} latency=${latency} ip=${ip} method=${method} path=${path} trace_id=${respHeader:" + tracing.TraceIDHeader + "} ua=\"${safe_ua}\" bytes_in=${bytesReceived} bytes_out=${bytesSent} req_headers=\"${safe_req_headers}\" resp_headers=\"${safe_resp_headers}\" error=\"${safe_error}\"\n",
		CustomTags: map[string]fiberlogger.LogFunc{
			"safe_req_headers": func(output fiberlogger.Buffer, c *fiber.Ctx, _ *fiberlogger.Data, _ string) (int, error) {
				return output.WriteString(formatHeaderLogValue(requestHeaders(c)))
			},
			"safe_resp_headers": func(output fiberlogger.Buffer, c *fiber.Ctx, _ *fiberlogger.Data, _ string) (int, error) {
				return output.WriteString(formatHeaderLogValue(responseHeaders(c)))
			},
			"safe_ua": func(output fiberlogger.Buffer, c *fiber.Ctx, _ *fiberlogger.Data, _ string) (int, error) {
				return output.WriteString(sanitizeLogValue(c.Get(fiber.HeaderUserAgent)))
			},
			"safe_error": func(output fiberlogger.Buffer, c *fiber.Ctx, _ *fiberlogger.Data, _ string) (int, error) {
				if err := c.Context().Err(); err != nil {
					return output.WriteString(sanitizeLogValue(err.Error()))
				}
				return 0, nil
			},
		},
	}
	if output != nil {
		cfg.Output = output
	}
	return cfg
}

func requestHeaders(c *fiber.Ctx) map[string][]string {
	out := map[string][]string{}
	c.Request().Header.VisitAll(func(key, value []byte) {
		k := string(key)
		out[k] = append(out[k], string(value))
	})
	return out
}

func responseHeaders(c *fiber.Ctx) map[string][]string {
	out := map[string][]string{}
	c.Response().Header.VisitAll(func(key, value []byte) {
		k := string(key)
		out[k] = append(out[k], string(value))
	})
	return out
}

func formatHeaderLogValue(headers map[string][]string) string {
	if len(headers) == 0 {
		return ""
	}
	keys := make([]string, 0, len(headers))
	for key := range headers {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		values := headers[key]
		for _, value := range values {
			parts = append(parts, key+"="+sanitizeHeaderValue(key, value))
		}
	}
	return sanitizeLogValue(strings.Join(parts, ";"))
}

func sanitizeHeaderValue(key string, value string) string {
	if _, ok := redactedLogHeaders[strings.ToLower(strings.TrimSpace(key))]; ok {
		return "[REDACTED]"
	}
	return sanitizeLogValue(value)
}

func sanitizeLogValue(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	v = strings.ReplaceAll(v, "\\", "\\\\")
	v = strings.ReplaceAll(v, "\"", "\\\"")
	v = strings.ReplaceAll(v, "\r", "\\r")
	v = strings.ReplaceAll(v, "\n", "\\n")
	return v
}

func registerPublicRoutes(app *fiber.App, cfg *config.Config, deps Dependencies, httpMetrics *prommetrics.HTTPMetrics) {
	app.Get("/health", healthHandler)
	app.Get("/ready", readyHandler(deps))
	app.Get("/metrics", httpMetrics.Handler())
	if deps.AdminHandlers != nil {
		app.Get("/api/v1/login-discovery/tenants", deps.AdminHandlers.DiscoverTenantsByIdentifier)
		app.Get("/api/v1/login-discovery/tenant/:slug", deps.AdminHandlers.DiscoverTenantByName)
		app.Get("/api/v1/commercial-plans", deps.AdminHandlers.ListPublicCommercialPlans)
		app.Get("/api/v1/subscription/payment/callback/:provider", deps.AdminHandlers.SubscriptionPaymentCallback)
		app.Post("/api/v1/subscription/payment/callback/:provider", deps.AdminHandlers.SubscriptionPaymentCallback)
		platformAdmin := app.Group("/api/v1/platform", platformAdminAuthMiddleware(cfg.GeneralSettings.MasterKey, deps.PlatformServiceAccounts))
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
		platformAdmin.Post("/tenants/:tenantID/suspend", deps.AdminHandlers.SuspendPlatformTenant)
		platformAdmin.Post("/tenants/:tenantID/reactivate", deps.AdminHandlers.ReactivatePlatformTenant)
		platformAdmin.Post("/tenants/:tenantID/subscription/cancel-at-period-end", deps.AdminHandlers.SchedulePlatformTenantSubscriptionCancellation)
		platformAdmin.Delete("/tenants/:tenantID/subscription/cancel-at-period-end", deps.AdminHandlers.ReversePlatformTenantSubscriptionCancellation)
		platformAdmin.Get("/tenants/:tenantID/billing/payment-intents", deps.AdminHandlers.ListPlatformTenantSubscriptionPaymentIntents)
	}
}

func registerProtectedRoutes(app *fiber.App, deps Dependencies) {

	v1 := app.Group("/v1")
	if deps.ProxyHandler != nil {
		v1.Post("/chat/completions", deps.ProxyHandler.Handle)
	}
	if deps.CompletionsHandler != nil {
		v1.Post("/completions", deps.CompletionsHandler.Handle)
	}
	if deps.ResponsesHandler != nil {
		v1.Post("/responses", deps.ResponsesHandler.Handle)
	}
	if deps.EmbeddingsHandler != nil {
		v1.Post("/embeddings", deps.EmbeddingsHandler.Handle)
	}
	if deps.ModerationsHandler != nil {
		v1.Post("/moderations", deps.ModerationsHandler.Handle)
	}
	if deps.ImagesHandler != nil {
		v1.Post("/images/generations", deps.ImagesHandler.HandleGenerations)
		v1.Post("/images/edits", deps.ImagesHandler.HandleEdits)
		v1.Post("/images/variations", deps.ImagesHandler.HandleVariations)
	}
	if deps.ModelsHandler != nil {
		v1.Get("/models", deps.ModelsHandler.Handle)
	}

	api := app.Group("/api/v1")
	if deps.AdminHandlers != nil {
		api.Get("/onboarding/tenant-slugs/:slug/availability", deps.AdminHandlers.CheckTenantSlugAvailability)
		api.Post("/onboarding/otp", deps.AdminHandlers.RequestOnboardingOTP)
		api.Post("/onboarding/otp/verify", deps.AdminHandlers.VerifyOnboardingOTP)
		api.Get("/subscription", deps.AdminHandlers.GetTenantSubscription)
		api.Post("/subscription/change", deps.AdminHandlers.ChangeTenantSubscription)
		api.Get("/entitlements", deps.AdminHandlers.GetTenantEntitlements)
		api.Get("/keys", deps.AdminHandlers.ListKeys)
		api.Post("/keys", deps.AdminHandlers.CreateKey)
		api.Post("/keys/vat", deps.AdminHandlers.CreateVATKey)
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
		api.Get("/virtual-users/:virtualUserID/tokens", deps.AdminHandlers.ListVirtualUserTokens)
		api.Post("/virtual-users/:virtualUserID/tokens", deps.AdminHandlers.CreateVirtualUserToken)
		api.Patch("/virtual-users/:virtualUserID/tokens/:tokenID", deps.AdminHandlers.UpdateVirtualUserToken)
		api.Post("/virtual-users/:virtualUserID/tokens/:tokenID/revoke", deps.AdminHandlers.RevokeVirtualUserToken)
		api.Post("/virtual-users/:virtualUserID/tokens/:tokenID/rotate", deps.AdminHandlers.RotateVirtualUserToken)
		api.Post("/models", deps.AdminHandlers.CreateModel)
		api.Get("/models", deps.AdminHandlers.ListModels)
		api.Get("/models/allowed", deps.AdminHandlers.ListAllowedModels)
		api.Post("/virtual-models", deps.AdminHandlers.CreateVirtualModel)
		api.Get("/virtual-models", deps.AdminHandlers.ListVirtualModels)
		api.Get("/virtual-models/:id", deps.AdminHandlers.GetVirtualModel)
		api.Patch("/virtual-models/:id", deps.AdminHandlers.UpdateVirtualModel)
		api.Delete("/virtual-models/:id", deps.AdminHandlers.DeleteVirtualModel)
		api.Get("/gateway-policies/rate-limit-rules", deps.AdminHandlers.ListRateLimitRules)
		api.Post("/gateway-policies/rate-limit-rules", deps.AdminHandlers.CreateRateLimitRule)
		api.Get("/gateway-policies/rate-limit-rules/:id", deps.AdminHandlers.GetRateLimitRule)
		api.Put("/gateway-policies/rate-limit-rules/:id", deps.AdminHandlers.UpdateRateLimitRule)
		api.Delete("/gateway-policies/rate-limit-rules/:id", deps.AdminHandlers.DeleteRateLimitRule)
		api.Patch("/gateway-policies/rate-limit-rules/:id/status", deps.AdminHandlers.UpdateRateLimitRuleStatus)
		api.Get("/gateway-policies/budget-limit-rules", deps.AdminHandlers.ListBudgetLimitRules)
		api.Post("/gateway-policies/budget-limit-rules", deps.AdminHandlers.CreateBudgetLimitRule)
		api.Get("/gateway-policies/budget-limit-rules/:id", deps.AdminHandlers.GetBudgetLimitRule)
		api.Put("/gateway-policies/budget-limit-rules/:id", deps.AdminHandlers.UpdateBudgetLimitRule)
		api.Delete("/gateway-policies/budget-limit-rules/:id", deps.AdminHandlers.DeleteBudgetLimitRule)
		api.Patch("/gateway-policies/budget-limit-rules/:id/status", deps.AdminHandlers.UpdateBudgetLimitRuleStatus)
		api.Get("/guardrails", deps.AdminHandlers.ListTenantGuardrails)
		api.Post("/guardrails", deps.AdminHandlers.CreateTenantGuardrail)
		api.Get("/guardrails/:slug", deps.AdminHandlers.GetTenantGuardrail)
		api.Put("/guardrails/:slug", deps.AdminHandlers.UpdateTenantGuardrail)
		api.Delete("/guardrails/:slug", deps.AdminHandlers.DeleteTenantGuardrail)
		api.Get("/prompts", deps.AdminHandlers.ListPrompts)
		api.Post("/prompts", deps.AdminHandlers.CreatePrompt)
		api.Get("/prompts/:promptID", deps.AdminHandlers.GetPrompt)
		api.Delete("/prompts/:promptID", deps.AdminHandlers.DeletePrompt)
		api.Get("/prompts/:promptID/versions", deps.AdminHandlers.ListPromptVersions)
		api.Post("/prompts/:promptID/versions", deps.AdminHandlers.CreatePromptVersion)
		api.Get("/prompt-versions/:versionID", deps.AdminHandlers.GetPromptVersion)
		api.Post("/prompt-versions/:versionID/render", deps.AdminHandlers.RenderPromptVersion)
		api.Get("/provider-catalog/providers", deps.AdminHandlers.ListProviderCatalog)
		api.Post("/provider-catalog/refresh", deps.AdminHandlers.RefreshProviderCatalogCache)
		api.Get("/provider-catalog/stats", deps.AdminHandlers.GetProviderCatalogStats)
		api.Get("/provider-catalog/display-settings", deps.AdminHandlers.GetProviderCatalogDisplaySettings)
		api.Get("/provider-catalog/assets/:bucket/:fileName", deps.AdminHandlers.GetProviderCatalogAsset)
		api.Get("/provider-catalog/providers/:provider/models", deps.AdminHandlers.ListProviderModelsCatalog)
		api.Get("/access-principals", deps.AdminHandlers.ListAccessPrincipals)
		api.Get("/roles", deps.AdminHandlers.ListRoles)
		api.Post("/roles", deps.AdminHandlers.CreateRole)
		api.Patch("/roles/:roleID", deps.AdminHandlers.UpdateRole)
		api.Get("/me/permissions", deps.AdminHandlers.GetCurrentUserPermissions)
		api.Get("/users", deps.AdminHandlers.ListTenantUsers)
		api.Patch("/users/roles", deps.AdminHandlers.UpdateUserRoles)
		api.Post("/users/invite", deps.AdminHandlers.InviteTenantUser)
		api.Post("/users/invitations", deps.AdminHandlers.InviteTenantUser)
		api.Post("/users/invitations/accept", deps.AdminHandlers.AcceptTenantUserInvitation)
		api.Post("/users/:userID/activate", deps.AdminHandlers.ActivateTenantUser)
		api.Post("/users/:userID/deactivate", deps.AdminHandlers.DeactivateTenantUser)
		api.Post("/tenant-provider-accounts", deps.AdminHandlers.CreateTenantProviderAccount)
		api.Get("/tenant-provider-accounts", deps.AdminHandlers.ListTenantProviderAccounts)
		api.Get("/tenant-provider-accounts/:accountID", deps.AdminHandlers.GetTenantProviderAccount)
		api.Delete("/tenant-provider-accounts/:accountID", deps.AdminHandlers.DeleteTenantProviderAccount)
		api.Get("/tenant-provider-accounts/:accountID/effective-models", deps.AdminHandlers.GetTenantProviderAccountEffectiveModels)
		api.Put("/tenant-provider-accounts/:accountID/credentials", deps.AdminHandlers.UpdateTenantProviderAccountCredentials)
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
		api.Get("/metrics/facets", deps.AdminHandlers.ListModelMetricsFacets)
		api.Post("/metrics/query", deps.AdminHandlers.QueryModelMetrics)
		api.Get("/settings/trace-privacy", deps.AdminHandlers.GetTenantTracePrivacySettings)
		api.Put("/settings/trace-privacy", deps.AdminHandlers.UpdateTenantTracePrivacySettings)
	}

	apiV2 := app.Group("/api/v2")
	if deps.AdminHandlers != nil {
		apiV2.Get("/traces", deps.AdminHandlers.ListRequestTracesV2)
		apiV2.Get("/traces/facets", deps.AdminHandlers.GetTraceFacetsV2)
		apiV2.Get("/traces/:traceID", deps.AdminHandlers.GetRequestTraceV2)
		apiV2.Get("/traces/:traceID/timeline", deps.AdminHandlers.ListRequestTraceTimelineV2)
		apiV2.Get("/traces/:traceID/attempts", deps.AdminHandlers.ListRequestTraceAttemptsV2)
	}

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

func platformAdminAuthMiddleware(masterKey string, serviceAccounts auth.PlatformServiceAccountStore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		apiKey := extractPlatformAdminAPIKey(c)
		if apiKey == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "platform admin authentication required")
		}
		if strings.TrimSpace(masterKey) != "" && secureTokenEqual(apiKey, masterKey) {
			return c.Next()
		}
		if serviceAccounts == nil {
			return fiber.NewError(fiber.StatusUnauthorized, "platform admin authentication required")
		}
		credential, err := serviceAccounts.GetPlatformServiceAccountByTokenHash(c.Context(), auth.HashKey(apiKey))
		if errors.Is(err, auth.ErrInvalidPlatformServiceCredential) {
			return fiber.NewError(fiber.StatusUnauthorized, "platform admin authentication required")
		}
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "platform service authentication lookup failed")
		}
		requiredScope := requiredPlatformScope(c.Method(), c.Path())
		if credential == nil || !auth.HasPlatformScope(credential.Scopes, requiredScope) {
			return fiber.NewError(fiber.StatusForbidden, "platform service account lacks required scope")
		}
		return c.Next()
	}
}

func requiredPlatformScope(method string, path string) string {
	if method == fiber.MethodGet || method == fiber.MethodHead || method == fiber.MethodOptions {
		return auth.PlatformScopeRead
	}
	if strings.Contains(path, "/tenant-subscriptions") || strings.Contains(path, "/tenant-entitlement-overrides") || strings.Contains(path, "/tenants/") {
		return auth.PlatformScopeTenantWrite
	}
	return auth.PlatformScopeCatalogWrite
}

func secureTokenEqual(left string, right string) bool {
	leftHash := sha256.Sum256([]byte(left))
	rightHash := sha256.Sum256([]byte(right))
	return subtle.ConstantTimeCompare(leftHash[:], rightHash[:]) == 1
}

func extractPlatformAdminAPIKey(c *fiber.Ctx) string {
	authHeader := strings.TrimSpace(c.Get("Authorization"))
	if authHeader != "" {
		if strings.HasPrefix(authHeader, "Bearer ") {
			return strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		}
		return authHeader
	}
	if key := strings.TrimSpace(c.Get("api-key")); key != "" {
		return key
	}
	return strings.TrimSpace(c.Get("x-api-key"))
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
