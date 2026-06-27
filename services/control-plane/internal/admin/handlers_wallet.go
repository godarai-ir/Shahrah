package admin

import (
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
)

func (h *Handlers) GetWallet(c *fiber.Ctx) error {
	if err := requireWalletRead(c); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	wallet, err := h.wallets.GetWallet(c.Context(), currentTenant.ID)
	if err != nil {
		return err
	}

	return c.JSON(wallet)
}

type topUpRequest struct {
	AmountIRR float64                `json:"amount_irr"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// TopUpWallet performs a stubbed immediate top-up.
func (h *Handlers) TopUpWallet(c *fiber.Ctx) error {
	if err := requireWalletManage(c); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	var req topUpRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	txn, err := h.wallets.TopUpWallet(c.Context(), currentTenant.ID, req.AmountIRR, req.Metadata)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(txn)
}
