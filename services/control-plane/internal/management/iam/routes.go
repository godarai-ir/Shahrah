// Package iam owns tenant identity and access management HTTP routes.
package iam

import "github.com/gofiber/fiber/v2"

// Handler defines IAM operations consumed by this feature.
type Handler interface {
	ListKeys(*fiber.Ctx) error
	CreateKey(*fiber.Ctx) error
	CreateVATKey(*fiber.Ctx) error
	ListPersonalAccessTokens(*fiber.Ctx) error
	CreatePersonalAccessToken(*fiber.Ctx) error
	PersonalAccessTokenExists(*fiber.Ctx) error
	RevokeAllPersonalAccessTokens(*fiber.Ctx) error
	RevokePersonalAccessToken(*fiber.Ctx) error
	ListOrgUnits(*fiber.Ctx) error
	CreateOrgUnit(*fiber.Ctx) error
	GetOrgUnit(*fiber.Ctx) error
	UpdateOrgUnit(*fiber.Ctx) error
	DeleteOrgUnit(*fiber.Ctx) error
	ListOrgUnitMembers(*fiber.Ctx) error
	AddOrgUnitMember(*fiber.Ctx) error
	UpdateOrgUnitMember(*fiber.Ctx) error
	DeleteOrgUnitMember(*fiber.Ctx) error
	ListTeams(*fiber.Ctx) error
	CreateTeam(*fiber.Ctx) error
	GetTeam(*fiber.Ctx) error
	UpdateTeam(*fiber.Ctx) error
	DeleteTeam(*fiber.Ctx) error
	ListTeamMembers(*fiber.Ctx) error
	AddTeamMember(*fiber.Ctx) error
	UpdateTeamMember(*fiber.Ctx) error
	DeleteTeamMember(*fiber.Ctx) error
	GetTeamPermissions(*fiber.Ctx) error
	UpdateTeamPermissions(*fiber.Ctx) error
	ListVirtualUsers(*fiber.Ctx) error
	CreateVirtualUser(*fiber.Ctx) error
	GetVirtualUser(*fiber.Ctx) error
	UpdateVirtualUser(*fiber.Ctx) error
	DeleteVirtualUser(*fiber.Ctx) error
	ListVirtualUserTokens(*fiber.Ctx) error
	CreateVirtualUserToken(*fiber.Ctx) error
	UpdateVirtualUserToken(*fiber.Ctx) error
	RevokeVirtualUserToken(*fiber.Ctx) error
	RotateVirtualUserToken(*fiber.Ctx) error
	ListAccessPrincipals(*fiber.Ctx) error
	ListRoles(*fiber.Ctx) error
	CreateRole(*fiber.Ctx) error
	UpdateRole(*fiber.Ctx) error
	GetCurrentUserPermissions(*fiber.Ctx) error
	ListTenantUsers(*fiber.Ctx) error
	UpdateUserRoles(*fiber.Ctx) error
	InviteTenantUser(*fiber.Ctx) error
	AcceptTenantUserInvitation(*fiber.Ctx) error
	ActivateTenantUser(*fiber.Ctx) error
	DeactivateTenantUser(*fiber.Ctx) error
	UpdateTenantUserProfile(*fiber.Ctx) error
}

// Routes owns IAM route registration.
type Routes struct {
	handler Handler
}

// NewRoutes creates IAM routes.
func NewRoutes(handler Handler) *Routes {
	return &Routes{handler: handler}
}

// RegisterTenant registers authenticated IAM endpoints.
func (r *Routes) RegisterTenant(router fiber.Router) {
	router.Get("/keys", r.handler.ListKeys)
	router.Post("/keys", r.handler.CreateKey)
	router.Post("/keys/vat", r.handler.CreateVATKey)
	router.Get("/personal-access-tokens", r.handler.ListPersonalAccessTokens)
	router.Post("/personal-access-tokens", r.handler.CreatePersonalAccessToken)
	router.Get("/personal-access-tokens/exists", r.handler.PersonalAccessTokenExists)
	router.Delete("/personal-access-tokens/revoke/all", r.handler.RevokeAllPersonalAccessTokens)
	router.Delete("/personal-access-tokens/:id", r.handler.RevokePersonalAccessToken)
	router.Get("/org-units", r.handler.ListOrgUnits)
	router.Post("/org-units", r.handler.CreateOrgUnit)
	router.Get("/org-units/:orgUnitID", r.handler.GetOrgUnit)
	router.Patch("/org-units/:orgUnitID", r.handler.UpdateOrgUnit)
	router.Delete("/org-units/:orgUnitID", r.handler.DeleteOrgUnit)
	router.Get("/org-units/:orgUnitID/members", r.handler.ListOrgUnitMembers)
	router.Post("/org-units/:orgUnitID/members", r.handler.AddOrgUnitMember)
	router.Patch("/org-units/:orgUnitID/members/:userID", r.handler.UpdateOrgUnitMember)
	router.Delete("/org-units/:orgUnitID/members/:userID", r.handler.DeleteOrgUnitMember)
	router.Get("/teams", r.handler.ListTeams)
	router.Post("/teams", r.handler.CreateTeam)
	router.Get("/teams/:teamID", r.handler.GetTeam)
	router.Patch("/teams/:teamID", r.handler.UpdateTeam)
	router.Delete("/teams/:teamID", r.handler.DeleteTeam)
	router.Get("/teams/:teamID/members", r.handler.ListTeamMembers)
	router.Post("/teams/:teamID/members", r.handler.AddTeamMember)
	router.Patch("/teams/:teamID/members/:userID", r.handler.UpdateTeamMember)
	router.Delete("/teams/:teamID/members/:userID", r.handler.DeleteTeamMember)
	router.Get("/teams/:teamID/permissions", r.handler.GetTeamPermissions)
	router.Patch("/teams/:teamID/permissions", r.handler.UpdateTeamPermissions)
	router.Get("/virtual-users", r.handler.ListVirtualUsers)
	router.Post("/virtual-users", r.handler.CreateVirtualUser)
	router.Get("/virtual-users/:virtualUserID", r.handler.GetVirtualUser)
	router.Patch("/virtual-users/:virtualUserID", r.handler.UpdateVirtualUser)
	router.Delete("/virtual-users/:virtualUserID", r.handler.DeleteVirtualUser)
	router.Get("/virtual-users/:virtualUserID/tokens", r.handler.ListVirtualUserTokens)
	router.Post("/virtual-users/:virtualUserID/tokens", r.handler.CreateVirtualUserToken)
	router.Patch("/virtual-users/:virtualUserID/tokens/:tokenID", r.handler.UpdateVirtualUserToken)
	router.Post("/virtual-users/:virtualUserID/tokens/:tokenID/revoke", r.handler.RevokeVirtualUserToken)
	router.Post("/virtual-users/:virtualUserID/tokens/:tokenID/rotate", r.handler.RotateVirtualUserToken)
	router.Get("/access-principals", r.handler.ListAccessPrincipals)
	router.Get("/roles", r.handler.ListRoles)
	router.Post("/roles", r.handler.CreateRole)
	router.Patch("/roles/:roleID", r.handler.UpdateRole)
	router.Get("/me/permissions", r.handler.GetCurrentUserPermissions)
	router.Get("/users", r.handler.ListTenantUsers)
	router.Patch("/users/roles", r.handler.UpdateUserRoles)
	router.Post("/users/invite", r.handler.InviteTenantUser)
	router.Post("/users/invitations", r.handler.InviteTenantUser)
	router.Post("/users/invitations/accept", r.handler.AcceptTenantUserInvitation)
	router.Patch("/users/:userID/profile", r.handler.UpdateTenantUserProfile)
	router.Post("/users/:userID/activate", r.handler.ActivateTenantUser)
	router.Post("/users/:userID/deactivate", r.handler.DeactivateTenantUser)
}
