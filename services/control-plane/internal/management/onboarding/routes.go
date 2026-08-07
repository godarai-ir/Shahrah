// Package onboarding owns tenant onboarding HTTP routes.
package onboarding

import "github.com/gofiber/fiber/v2"

// Handler defines onboarding operations consumed by this feature.
type Handler interface {
	CheckTenantSlugAvailability(*fiber.Ctx) error
	RequestOnboardingOTP(*fiber.Ctx) error
	VerifyOnboardingOTP(*fiber.Ctx) error
	DiscoverTenantsByIdentifier(*fiber.Ctx) error
	DiscoverTenantByName(*fiber.Ctx) error
}

// Routes owns onboarding route registration.
type Routes struct {
	handler Handler
}

// NewRoutes creates onboarding routes.
func NewRoutes(handler Handler) *Routes {
	return &Routes{handler: handler}
}

// RegisterPublic registers tenant discovery endpoints.
func (r *Routes) RegisterPublic(router fiber.Router) {
	router.Get("/api/v1/login-discovery/tenants", r.handler.DiscoverTenantsByIdentifier)
	router.Get("/api/v1/login-discovery/tenant/:slug", r.handler.DiscoverTenantByName)
}

// RegisterTenant registers authenticated onboarding endpoints.
func (r *Routes) RegisterTenant(router fiber.Router) {
	router.Get("/onboarding/tenant-slugs/:slug/availability", r.handler.CheckTenantSlugAvailability)
	router.Post("/onboarding/otp", r.handler.RequestOnboardingOTP)
	router.Post("/onboarding/otp/verify", r.handler.VerifyOnboardingOTP)
}
