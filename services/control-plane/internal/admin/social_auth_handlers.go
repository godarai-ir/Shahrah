package admin

import (
	"errors"
	"strings"

	"github.com/Delpak/shahrah/services/control-plane/internal/tenantadmin"
	"github.com/gofiber/fiber/v2"
)

func (h *AuthHandlers) ProvisionSocialTenant(c *fiber.Ctx) error {
	if h.socialStore == nil {
		return fiber.ErrNotImplemented
	}

	var body tenantadmin.ProvisionSocialTenantParams
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	result, err := h.socialStore.ProvisionSocialTenant(c.Context(), body)
	if err != nil {
		return socialAuthHTTPError(err)
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}

func (h *AuthHandlers) CreateSocialHandoff(c *fiber.Ctx) error {
	if h.socialStore == nil {
		return fiber.ErrNotImplemented
	}

	var body struct {
		TenantSlug string `json:"tenant_slug"`
		AccountID  string `json:"account_id"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	result, err := h.socialStore.CreateSocialLoginHandoff(c.Context(), body.TenantSlug, body.AccountID)
	if err != nil {
		return socialAuthHTTPError(err)
	}

	return c.JSON(result)
}

func (h *AuthHandlers) RedeemSocialHandoff(c *fiber.Ctx) error {
	if h.socialStore == nil {
		return fiber.ErrNotImplemented
	}

	var body struct {
		TenantSlug string `json:"tenant_slug"`
		Code       string `json:"code"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	result, err := h.socialStore.RedeemSocialLoginHandoff(c.Context(), body.TenantSlug, body.Code)
	if err != nil {
		return socialAuthHTTPError(err)
	}

	return c.JSON(result)
}

func (h *AuthHandlers) CreateSocialLinkTicket(c *fiber.Ctx) error {
	if h.socialStore == nil {
		return fiber.ErrNotImplemented
	}

	var body struct {
		TenantSlug string `json:"tenant_slug"`
		UserID     string `json:"user_id"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	result, err := h.socialStore.CreateSocialLinkTicket(c.Context(), body.TenantSlug, body.UserID)
	if err != nil {
		return socialAuthHTTPError(err)
	}

	return c.JSON(result)
}

func (h *AuthHandlers) CompleteSocialLink(c *fiber.Ctx) error {
	if h.socialStore == nil {
		return fiber.ErrNotImplemented
	}

	var body struct {
		TicketCode string `json:"ticket_code"`
		AccountID  string `json:"account_id"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	result, err := h.socialStore.CompleteSocialLink(c.Context(), body.TicketCode, body.AccountID)
	if err != nil {
		return socialAuthHTTPError(err)
	}

	return c.JSON(result)
}

func socialAuthHTTPError(err error) error {
	switch {
	case errors.Is(err, tenantadmin.ErrTenantSlugInvalid), errors.Is(err, tenantadmin.ErrSocialAccountInvalid):
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	case errors.Is(err, tenantadmin.ErrTenantSlugTaken), errors.Is(err, tenantadmin.ErrSocialLinkConflict):
		return fiber.NewError(fiber.StatusConflict, err.Error())
	case errors.Is(err, tenantadmin.ErrSocialAccessDenied):
		return fiber.NewError(fiber.StatusForbidden, err.Error())
	case errors.Is(err, tenantadmin.ErrSocialHandoffInvalid):
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	default:
		if strings.Contains(strings.ToLower(err.Error()), "plan") {
			return fiber.NewError(fiber.StatusBadRequest, "Plan not found")
		}

		return fiber.NewError(fiber.StatusInternalServerError, "Social authentication failed")
	}
}
