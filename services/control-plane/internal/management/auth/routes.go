// Package auth owns management authentication HTTP routes.
package auth

import "github.com/gofiber/fiber/v2"

// Handler defines authentication operations consumed by this feature.
type Handler interface {
	CreateCaptcha(*fiber.Ctx) error
	ConsumeCaptcha(*fiber.Ctx) error
	CreateOTPRequest(*fiber.Ctx) error
	VerifyOTPRequest(*fiber.Ctx) error
	CreateSession(*fiber.Ctx) error
	GetSession(*fiber.Ctx) error
	RevokeSession(*fiber.Ctx) error
	FindOrCreateUser(*fiber.Ctx) error
	GetUser(*fiber.Ctx) error
	UpdateUserContact(*fiber.Ctx) error
	TouchLastLogin(*fiber.Ctx) error
	ProvisionSocialTenant(*fiber.Ctx) error
	CreateSocialHandoff(*fiber.Ctx) error
	RedeemSocialHandoff(*fiber.Ctx) error
	CreateSocialLinkTicket(*fiber.Ctx) error
	CompleteSocialLink(*fiber.Ctx) error
}

// Routes owns authentication route registration.
type Routes struct {
	handler Handler
}

// NewRoutes creates authentication routes.
func NewRoutes(handler Handler) *Routes {
	return &Routes{handler: handler}
}

// Register registers authentication endpoints on the supplied router.
func (r *Routes) Register(router fiber.Router) {
	router.Post("/captcha/create", r.handler.CreateCaptcha)
	router.Post("/captcha/consume", r.handler.ConsumeCaptcha)
	router.Post("/otp/request", r.handler.CreateOTPRequest)
	router.Post("/otp/verify", r.handler.VerifyOTPRequest)
	router.Post("/session", r.handler.CreateSession)
	router.Get("/session/:id", r.handler.GetSession)
	router.Delete("/session/:id", r.handler.RevokeSession)
	router.Post("/user/find-or-create", r.handler.FindOrCreateUser)
	router.Get("/user/:id", r.handler.GetUser)
	router.Post("/user/:id/update-contact", r.handler.UpdateUserContact)
	router.Post("/user/:id/touch-last-login", r.handler.TouchLastLogin)
	router.Post("/social/provision", r.handler.ProvisionSocialTenant)
	router.Post("/social/handoff", r.handler.CreateSocialHandoff)
	router.Post("/social/handoff/redeem", r.handler.RedeemSocialHandoff)
	router.Post("/social/link-ticket", r.handler.CreateSocialLinkTicket)
	router.Post("/social/link", r.handler.CompleteSocialLink)
}
