// Package catalog owns model, policy, prompt, and provider catalog HTTP routes.
package catalog

import "github.com/gofiber/fiber/v2"

// Handler defines catalog operations consumed by this feature.
type Handler interface {
	CreateModel(*fiber.Ctx) error
	ListModels(*fiber.Ctx) error
	ListAllowedModels(*fiber.Ctx) error
	CreateVirtualModel(*fiber.Ctx) error
	ListVirtualModels(*fiber.Ctx) error
	GetVirtualModel(*fiber.Ctx) error
	UpdateVirtualModel(*fiber.Ctx) error
	DeleteVirtualModel(*fiber.Ctx) error
	ListRateLimitRules(*fiber.Ctx) error
	CreateRateLimitRule(*fiber.Ctx) error
	GetRateLimitRule(*fiber.Ctx) error
	UpdateRateLimitRule(*fiber.Ctx) error
	DeleteRateLimitRule(*fiber.Ctx) error
	UpdateRateLimitRuleStatus(*fiber.Ctx) error
	ListBudgetLimitRules(*fiber.Ctx) error
	CreateBudgetLimitRule(*fiber.Ctx) error
	GetBudgetLimitRule(*fiber.Ctx) error
	UpdateBudgetLimitRule(*fiber.Ctx) error
	DeleteBudgetLimitRule(*fiber.Ctx) error
	UpdateBudgetLimitRuleStatus(*fiber.Ctx) error
	ListTenantGuardrails(*fiber.Ctx) error
	CreateTenantGuardrail(*fiber.Ctx) error
	GetTenantGuardrail(*fiber.Ctx) error
	UpdateTenantGuardrail(*fiber.Ctx) error
	DeleteTenantGuardrail(*fiber.Ctx) error
	ListPrompts(*fiber.Ctx) error
	CreatePrompt(*fiber.Ctx) error
	GetPrompt(*fiber.Ctx) error
	DeletePrompt(*fiber.Ctx) error
	ListPromptVersions(*fiber.Ctx) error
	CreatePromptVersion(*fiber.Ctx) error
	GetPromptVersion(*fiber.Ctx) error
	RenderPromptVersion(*fiber.Ctx) error
	ListProviderCatalog(*fiber.Ctx) error
	RefreshProviderCatalogCache(*fiber.Ctx) error
	GetProviderCatalogStats(*fiber.Ctx) error
	GetProviderCatalogDisplaySettings(*fiber.Ctx) error
	GetProviderCatalogAsset(*fiber.Ctx) error
	ListProviderModelsCatalog(*fiber.Ctx) error
	ListTenantProviderConnectionSchemas(*fiber.Ctx) error
	TestProviderConnection(*fiber.Ctx) error
	ListTenantProviderAccountModels(*fiber.Ctx) error
	CreateTenantProviderAccount(*fiber.Ctx) error
	ListTenantProviderAccounts(*fiber.Ctx) error
	GetTenantProviderAccount(*fiber.Ctx) error
	DeleteTenantProviderAccount(*fiber.Ctx) error
	GetTenantProviderAccountEffectiveModels(*fiber.Ctx) error
	UpdateTenantProviderAccountCredentials(*fiber.Ctx) error
	UpsertTenantProviderAccountModels(*fiber.Ctx) error
	UpdateTenantProviderAccountAccess(*fiber.Ctx) error
	TestTenantProviderAccountCredentials(*fiber.Ctx) error
	GetPlatformProviderConnectionSchemas(*fiber.Ctx) error
	ListPlatformProviderConnections(*fiber.Ctx) error
	GetPlatformProviderConnection(*fiber.Ctx) error
	CreatePlatformProviderConnection(*fiber.Ctx) error
	UpdatePlatformProviderConnection(*fiber.Ctx) error
	CreatePlatformProviderCredentialVersion(*fiber.Ctx) error
	TestPlatformProviderCredentialVersion(*fiber.Ctx) error
	ActivatePlatformProviderCredentialVersion(*fiber.Ctx) error
	RollbackPlatformProviderCredentialVersion(*fiber.Ctx) error
	DisablePlatformProviderConnection(*fiber.Ctx) error
	RevokePlatformProviderCredentialVersion(*fiber.Ctx) error
}

// Routes owns catalog route registration.
type Routes struct {
	handler Handler
}

// NewRoutes creates catalog routes.
func NewRoutes(handler Handler) *Routes {
	return &Routes{handler: handler}
}

// RegisterTenant registers authenticated catalog endpoints.
func (r *Routes) RegisterTenant(router fiber.Router) {
	router.Post("/models", r.handler.CreateModel)
	router.Get("/models", r.handler.ListModels)
	router.Get("/models/allowed", r.handler.ListAllowedModels)
	router.Post("/virtual-models", r.handler.CreateVirtualModel)
	router.Get("/virtual-models", r.handler.ListVirtualModels)
	router.Get("/virtual-models/:id", r.handler.GetVirtualModel)
	router.Patch("/virtual-models/:id", r.handler.UpdateVirtualModel)
	router.Delete("/virtual-models/:id", r.handler.DeleteVirtualModel)
	router.Get("/gateway-policies/rate-limit-rules", r.handler.ListRateLimitRules)
	router.Post("/gateway-policies/rate-limit-rules", r.handler.CreateRateLimitRule)
	router.Get("/gateway-policies/rate-limit-rules/:id", r.handler.GetRateLimitRule)
	router.Put("/gateway-policies/rate-limit-rules/:id", r.handler.UpdateRateLimitRule)
	router.Delete("/gateway-policies/rate-limit-rules/:id", r.handler.DeleteRateLimitRule)
	router.Patch("/gateway-policies/rate-limit-rules/:id/status", r.handler.UpdateRateLimitRuleStatus)
	router.Get("/gateway-policies/budget-limit-rules", r.handler.ListBudgetLimitRules)
	router.Post("/gateway-policies/budget-limit-rules", r.handler.CreateBudgetLimitRule)
	router.Get("/gateway-policies/budget-limit-rules/:id", r.handler.GetBudgetLimitRule)
	router.Put("/gateway-policies/budget-limit-rules/:id", r.handler.UpdateBudgetLimitRule)
	router.Delete("/gateway-policies/budget-limit-rules/:id", r.handler.DeleteBudgetLimitRule)
	router.Patch("/gateway-policies/budget-limit-rules/:id/status", r.handler.UpdateBudgetLimitRuleStatus)
	router.Get("/guardrails", r.handler.ListTenantGuardrails)
	router.Post("/guardrails", r.handler.CreateTenantGuardrail)
	router.Get("/guardrails/:slug", r.handler.GetTenantGuardrail)
	router.Put("/guardrails/:slug", r.handler.UpdateTenantGuardrail)
	router.Delete("/guardrails/:slug", r.handler.DeleteTenantGuardrail)
	router.Get("/prompts", r.handler.ListPrompts)
	router.Post("/prompts", r.handler.CreatePrompt)
	router.Get("/prompts/:promptID", r.handler.GetPrompt)
	router.Delete("/prompts/:promptID", r.handler.DeletePrompt)
	router.Get("/prompts/:promptID/versions", r.handler.ListPromptVersions)
	router.Post("/prompts/:promptID/versions", r.handler.CreatePromptVersion)
	router.Get("/prompt-versions/:versionID", r.handler.GetPromptVersion)
	router.Post("/prompt-versions/:versionID/render", r.handler.RenderPromptVersion)
	router.Get("/provider-catalog/providers", r.handler.ListProviderCatalog)
	router.Post("/provider-catalog/refresh", r.handler.RefreshProviderCatalogCache)
	router.Get("/provider-catalog/stats", r.handler.GetProviderCatalogStats)
	router.Get("/provider-catalog/display-settings", r.handler.GetProviderCatalogDisplaySettings)
	router.Get("/provider-catalog/assets/:bucket/:fileName", r.handler.GetProviderCatalogAsset)
	router.Get("/provider-catalog/providers/:provider/models", r.handler.ListProviderModelsCatalog)
	router.Get("/provider-catalog/providers/:provider/connection-schemas", r.handler.ListTenantProviderConnectionSchemas)
	router.Post("/provider-connections/test", r.handler.TestProviderConnection)
	router.Get("/tenant-provider-account-models", r.handler.ListTenantProviderAccountModels)
	router.Post("/tenant-provider-accounts", r.handler.CreateTenantProviderAccount)
	router.Get("/tenant-provider-accounts", r.handler.ListTenantProviderAccounts)
	router.Get("/tenant-provider-accounts/:accountID", r.handler.GetTenantProviderAccount)
	router.Delete("/tenant-provider-accounts/:accountID", r.handler.DeleteTenantProviderAccount)
	router.Get("/tenant-provider-accounts/:accountID/effective-models", r.handler.GetTenantProviderAccountEffectiveModels)
	router.Put("/tenant-provider-accounts/:accountID/credentials", r.handler.UpdateTenantProviderAccountCredentials)
	router.Put("/tenant-provider-accounts/:accountID/models", r.handler.UpsertTenantProviderAccountModels)
	router.Put("/tenant-provider-accounts/:accountID/access", r.handler.UpdateTenantProviderAccountAccess)
	router.Post("/tenant-provider-accounts/:accountID/test-credentials", r.handler.TestTenantProviderAccountCredentials)
}

// RegisterPlatform registers platform provider credential endpoints.
func (r *Routes) RegisterPlatform(router fiber.Router) {
	router.Get("/provider-connection-schemas/:provider", r.handler.GetPlatformProviderConnectionSchemas)
	router.Get("/provider-connections", r.handler.ListPlatformProviderConnections)
	router.Get("/provider-connections/:id", r.handler.GetPlatformProviderConnection)
	router.Post("/provider-connections", r.handler.CreatePlatformProviderConnection)
	router.Patch("/provider-connections/:id", r.handler.UpdatePlatformProviderConnection)
	router.Post("/provider-connections/:id/credential-versions", r.handler.CreatePlatformProviderCredentialVersion)
	router.Post("/provider-connections/:id/credential-versions/:version/test", r.handler.TestPlatformProviderCredentialVersion)
	router.Post("/provider-connections/:id/credential-versions/:version/activate", r.handler.ActivatePlatformProviderCredentialVersion)
	router.Post("/provider-connections/:id/credential-versions/:version/rollback", r.handler.RollbackPlatformProviderCredentialVersion)
	router.Post("/provider-connections/:id/disable", r.handler.DisablePlatformProviderConnection)
	router.Post("/provider-connections/:id/credential-versions/:version/revoke", r.handler.RevokePlatformProviderCredentialVersion)
}
