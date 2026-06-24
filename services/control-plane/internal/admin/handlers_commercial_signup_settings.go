package admin

import (
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/gofiber/fiber/v2"
)

func (h *Handlers) GetPlatformCommercialSignupSettings(c *fiber.Ctx) error {
	store, err := h.requirePlanStore()
	if err != nil {
		return err
	}
	item, err := store.GetCommercialSignupSettings(c.Context())
	if err != nil {
		return err
	}
	return c.JSON(item)
}

func (h *Handlers) UpdatePlatformCommercialSignupSettings(c *fiber.Ctx) error {
	store, err := h.requirePlanStore()
	if err != nil {
		return err
	}
	var req platform.UpdateCommercialSignupSettingsRequest
	if err := c.BodyParser(&req); err != nil || req.SignupCreditIRR == nil {
		return fiber.NewError(fiber.StatusBadRequest, "signup_credit_irr is required")
	}
	item, err := store.UpdateCommercialSignupSettings(c.Context(), *req.SignupCreditIRR, strings.TrimSpace(c.Get("x-shahrah-actor-id")))
	if err != nil {
		if err == storage.ErrInvalidSignupCredit {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		return err
	}
	return c.JSON(item)
}
