// Package support owns tenant support and platform support HTTP routes.
package support

import "github.com/gofiber/fiber/v2"

// Handler defines support operations consumed by this feature.
type Handler interface {
	ListTenantSupportTickets(*fiber.Ctx) error
	CreateTenantSupportTicket(*fiber.Ctx) error
	GetTenantSupportTicket(*fiber.Ctx) error
	ListTenantSupportTicketMessages(*fiber.Ctx) error
	AddTenantSupportTicketMessage(*fiber.Ctx) error
	DownloadTenantSupportTicketAttachment(*fiber.Ctx) error
	RateTenantSupportTicket(*fiber.Ctx) error
	ListPlatformTenants(*fiber.Ctx) error
	GetPlatformTenant(*fiber.Ctx) error
	ListPlatformTenantEvents(*fiber.Ctx) error
	SuspendPlatformTenant(*fiber.Ctx) error
	ReactivatePlatformTenant(*fiber.Ctx) error
	ListPlatformSupportTickets(*fiber.Ctx) error
	CountPlatformSupportTickets(*fiber.Ctx) error
	GetPlatformSupportTicket(*fiber.Ctx) error
	ListPlatformSupportTicketMessages(*fiber.Ctx) error
	AddPlatformSupportTicketMessage(*fiber.Ctx) error
	TriagePlatformSupportTicket(*fiber.Ctx) error
	ReopenPlatformSupportTicket(*fiber.Ctx) error
	ListPlatformSupportTicketEvents(*fiber.Ctx) error
	DownloadPlatformSupportTicketAttachment(*fiber.Ctx) error
}

// Routes owns support route registration.
type Routes struct {
	handler Handler
}

// NewRoutes creates support routes.
func NewRoutes(handler Handler) *Routes {
	return &Routes{handler: handler}
}

// RegisterTenant registers tenant support endpoints.
func (r *Routes) RegisterTenant(router fiber.Router) {
	router.Get("/support/tickets", r.handler.ListTenantSupportTickets)
	router.Post("/support/tickets", r.handler.CreateTenantSupportTicket)
	router.Get("/support/tickets/:ticketID", r.handler.GetTenantSupportTicket)
	router.Get("/support/tickets/:ticketID/messages", r.handler.ListTenantSupportTicketMessages)
	router.Post("/support/tickets/:ticketID/messages", r.handler.AddTenantSupportTicketMessage)
	router.Get("/support/ticket-attachments/:attachmentID/download", r.handler.DownloadTenantSupportTicketAttachment)
	router.Put("/support/tickets/:ticketID/rating", r.handler.RateTenantSupportTicket)
}

// RegisterPlatform registers platform support endpoints.
func (r *Routes) RegisterPlatform(router fiber.Router) {
	router.Get("/tenants", r.handler.ListPlatformTenants)
	router.Get("/tenants/:tenantID", r.handler.GetPlatformTenant)
	router.Get("/tenants/:tenantID/events", r.handler.ListPlatformTenantEvents)
	router.Post("/tenants/:tenantID/suspend", r.handler.SuspendPlatformTenant)
	router.Post("/tenants/:tenantID/reactivate", r.handler.ReactivatePlatformTenant)
	router.Get("/support/tickets", r.handler.ListPlatformSupportTickets)
	router.Get("/support/tickets/counts", r.handler.CountPlatformSupportTickets)
	router.Get("/support/tickets/:ticketID", r.handler.GetPlatformSupportTicket)
	router.Get("/support/tickets/:ticketID/messages", r.handler.ListPlatformSupportTicketMessages)
	router.Post("/support/tickets/:ticketID/messages", r.handler.AddPlatformSupportTicketMessage)
	router.Patch("/support/tickets/:ticketID/triage", r.handler.TriagePlatformSupportTicket)
	router.Post("/support/tickets/:ticketID/reopen", r.handler.ReopenPlatformSupportTicket)
	router.Get("/support/tickets/:ticketID/events", r.handler.ListPlatformSupportTicketEvents)
	router.Get("/support/ticket-attachments/:attachmentID/download", r.handler.DownloadPlatformSupportTicketAttachment)
}
