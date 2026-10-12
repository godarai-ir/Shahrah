package admin

import (
	"errors"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/apperrors"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
)

// onBehalfOfProducts are the product labels a service key may report.
var onBehalfOfProducts = map[string]struct{}{
	platform.CreditProductLibreChat: {},
}

// UpdateVirtualUserTokenOnBehalfOf enables or disables x-godarai-on-behalf-of
// for a VAT. A key with this capability can act for any member of the tenant,
// so only tenant admins may change it.
func (h *Handlers) UpdateVirtualUserTokenOnBehalfOf(c *fiber.Ctx) error {
	if !auth.IsTenantAdminOrHigher(auth.GetAuthInfo(c)) {
		return fiber.NewError(fiber.StatusForbidden, "tenant admin role required")
	}

	if h.keys == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "virtual user token store not configured")
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	virtualUserID, tokenID := strings.TrimSpace(c.Params("virtualUserID")), strings.TrimSpace(c.Params("tokenID"))
	if virtualUserID == "" || tokenID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "virtualUserID and tokenID are required")
	}

	var req platform.UpdateVATOnBehalfOfRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	req.Product = strings.ToLower(strings.TrimSpace(req.Product))
	if req.Enabled && req.Product == "" {
		req.Product = platform.CreditProductLibreChat
	}

	if _, ok := onBehalfOfProducts[req.Product]; req.Enabled && !ok {
		return fiber.NewError(fiber.StatusBadRequest, "product is not supported for on-behalf-of keys")
	}

	updated, err := h.keys.SetVATOnBehalfOf(c.Context(), currentTenant.ID, virtualUserID, tokenID, req, delegatedUserID(c))
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "virtual user token not found")
		}

		return err
	}

	return c.JSON(updated)
}
