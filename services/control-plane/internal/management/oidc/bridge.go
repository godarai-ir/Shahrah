package oidc

import (
	"errors"
	"strings"

	"github.com/Delpak/shahrah/services/control-plane/internal/oidc"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type loginDecisionBody struct {
	TenantSlug string `json:"tenant_slug"`
	SessionID  string `json:"session_id"`
}

// trustedCaller limits the login bridge to the tenant-admin server's master or platform
// credential, because accepting a request mints an authorization code for a session.
func (r *Routes) trustedCaller(c *fiber.Ctx) bool {
	info := r.authInfo(c)

	return info != nil && r.isPrivilege(info.Role)
}

func writeBridgeForbidden(c *fiber.Ctx) error {
	return writeManagementError(c, fiber.StatusForbidden, "oidc_bridge_forbidden", "the login bridge requires a trusted server credential")
}

func (r *Routes) getLoginRequest(c *fiber.Ctx) error {
	if !r.trustedCaller(c) {
		return writeBridgeForbidden(c)
	}

	request, err := r.service.GetLoginRequest(c.Context(), c.Query("tenant_slug"), c.Params("id"))
	if err != nil {
		return r.mapError(c, err)
	}

	return c.JSON(request)
}

func (r *Routes) acceptLogin(c *fiber.Ctx) error {
	if !r.trustedCaller(c) {
		return writeBridgeForbidden(c)
	}

	var body loginDecisionBody
	if err := c.BodyParser(&body); err != nil {
		return writeManagementError(c, fiber.StatusBadRequest, "invalid_request", "request body is invalid")
	}

	decision, err := r.service.AcceptLogin(c.Context(), body.TenantSlug, c.Params("id"), body.SessionID)
	if err != nil {
		return r.mapError(c, err)
	}

	return c.JSON(decision)
}

func (r *Routes) rejectLogin(c *fiber.Ctx) error {
	if !r.trustedCaller(c) {
		return writeBridgeForbidden(c)
	}

	var body loginDecisionBody
	if err := c.BodyParser(&body); err != nil {
		return writeManagementError(c, fiber.StatusBadRequest, "invalid_request", "request body is invalid")
	}

	decision, err := r.service.RejectLogin(c.Context(), body.TenantSlug, c.Params("id"))
	if err != nil {
		return r.mapError(c, err)
	}

	return c.JSON(decision)
}

func (r *Routes) managedTenantID(c *fiber.Ctx) (string, bool) {
	current := r.tenant(c)
	if current == nil || strings.TrimSpace(current.ID) == "" {
		return "", false
	}

	return current.ID, true
}

func (r *Routes) listClients(c *fiber.Ctx) error {
	tenantID, ok := r.managedTenantID(c)
	if !ok {
		return writeManagementError(c, fiber.StatusForbidden, "tenant_required", "tenant context is required")
	}

	clients, err := r.service.ListClients(c.Context(), tenantID)
	if err != nil {
		return r.mapError(c, err)
	}

	return c.JSON(fiber.Map{"data": clients})
}

func (r *Routes) createClient(c *fiber.Ctx) error {
	tenantID, ok := r.managedTenantID(c)
	if !ok {
		return writeManagementError(c, fiber.StatusForbidden, "tenant_required", "tenant context is required")
	}

	var request oidc.CreateClientRequest
	if err := c.BodyParser(&request); err != nil {
		return writeManagementError(c, fiber.StatusBadRequest, "invalid_request", "request body is invalid")
	}

	created, err := r.service.CreateClient(c.Context(), tenantID, r.actorID(c), request)
	if err != nil {
		return r.mapError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(created)
}

func (r *Routes) updateClient(c *fiber.Ctx) error {
	tenantID, ok := r.managedTenantID(c)
	if !ok {
		return writeManagementError(c, fiber.StatusForbidden, "tenant_required", "tenant context is required")
	}

	var request oidc.UpdateClientRequest
	if err := c.BodyParser(&request); err != nil {
		return writeManagementError(c, fiber.StatusBadRequest, "invalid_request", "request body is invalid")
	}

	updated, err := r.service.UpdateClient(c.Context(), tenantID, c.Params("id"), request)
	if err != nil {
		return r.mapError(c, err)
	}

	return c.JSON(updated)
}

func (r *Routes) rotateClientSecret(c *fiber.Ctx) error {
	tenantID, ok := r.managedTenantID(c)
	if !ok {
		return writeManagementError(c, fiber.StatusForbidden, "tenant_required", "tenant context is required")
	}

	rotated, err := r.service.RotateClientSecret(c.Context(), tenantID, c.Params("id"))
	if err != nil {
		return r.mapError(c, err)
	}

	return c.JSON(rotated)
}

func (r *Routes) deleteClient(c *fiber.Ctx) error {
	tenantID, ok := r.managedTenantID(c)
	if !ok {
		return writeManagementError(c, fiber.StatusForbidden, "tenant_required", "tenant context is required")
	}

	if err := r.service.DeleteClient(c.Context(), tenantID, c.Params("id")); err != nil {
		return r.mapError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (r *Routes) actorID(c *fiber.Ctx) string {
	if actor := strings.TrimSpace(c.Get("x-shahrah-actor-id")); actor != "" {
		return strings.Clone(actor)
	}

	if info := r.authInfo(c); info != nil {
		return info.UserID
	}

	return ""
}

func (r *Routes) mapError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, oidc.ErrClientNotFound):
		return writeManagementError(c, fiber.StatusNotFound, "oidc_client_not_found", "oidc client not found")
	case errors.Is(err, oidc.ErrClientIDTaken):
		return writeManagementError(c, fiber.StatusConflict, "oidc_client_conflict", "oidc client already exists")
	case errors.Is(err, oidc.ErrInvalidClientConfig):
		return writeManagementError(c, fiber.StatusUnprocessableEntity, "oidc_client_invalid", err.Error())
	case errors.Is(err, oidc.ErrRequestNotFound):
		return writeManagementError(c, fiber.StatusNotFound, "oidc_request_not_found", "sign-in request not found")
	case errors.Is(err, oidc.ErrRequestNotPending):
		return writeManagementError(c, fiber.StatusGone, "oidc_request_expired", "sign-in request expired or was already completed")
	case errors.Is(err, oidc.ErrSessionInvalid):
		return writeManagementError(c, fiber.StatusUnauthorized, "oidc_session_invalid", "a fresh sign-in is required")
	}

	r.logger.Error("oidc management request failed", zap.Error(err), zap.String("path", c.Path()))

	return writeManagementError(c, fiber.StatusInternalServerError, "server_error", "the request could not be completed")
}

func writeManagementError(c *fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"error": fiber.Map{
			"code":    code,
			"message": message,
			"type":    "oidc_error",
		},
	})
}
