package admin

import (
	"errors"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

func (h *Handlers) CreateBankTransfer(c *fiber.Ctx) error {
	if err := requireBillingManage(c); err != nil {
		return err
	}
	currentTenant, err := currentTenantOrErr(c)
	if err != nil {
		return err
	}
	if h.wallets == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "wallet store not configured")
	}

	var req platform.CreateBankTransferRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	req.TrackingNumber = strings.TrimSpace(req.TrackingNumber)
	req.ReceiptFileURL = strings.TrimSpace(req.ReceiptFileURL)
	if req.AmountIRR <= 0 || req.TrackingNumber == "" || req.BankTransactionAt.IsZero() {
		return fiber.NewError(fiber.StatusBadRequest, "amount_irr, tracking_number, and bank_transaction_at are required")
	}

	item, err := h.wallets.CreateBankTransferRequest(c.Context(), currentTenant.ID, req, billingActorID(c))
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handlers) ListBankTransfers(c *fiber.Ctx) error {
	if err := requireBillingRead(c); err != nil {
		return err
	}
	currentTenant, err := currentTenantOrErr(c)
	if err != nil {
		return err
	}
	if h.wallets == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "wallet store not configured")
	}

	items, err := h.wallets.ListBankTransferRequests(c.Context(), currentTenant.ID, c.QueryInt("limit", 50))
	if err != nil {
		return err
	}

	return c.JSON(platform.BankTransferListResponse{Data: items})
}

func (h *Handlers) ApprovePlatformBankTransfer(c *fiber.Ctx) error {
	return h.reviewPlatformBankTransfer(c, true)
}

func (h *Handlers) RejectPlatformBankTransfer(c *fiber.Ctx) error {
	return h.reviewPlatformBankTransfer(c, false)
}

func (h *Handlers) reviewPlatformBankTransfer(c *fiber.Ctx, approve bool) error {
	if h.wallets == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "wallet store not configured")
	}

	var req platform.ReviewBankTransferRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	var (
		item *platform.BankTransferRequest
		err  error
	)
	if approve {
		item, err = h.wallets.ApproveBankTransferRequest(c.Context(), c.Params("tenantID"), c.Params("requestID"), req, supportActor(c))
	} else {
		item, err = h.wallets.RejectBankTransferRequest(c.Context(), c.Params("tenantID"), c.Params("requestID"), req, supportActor(c))
	}
	if err != nil {
		return mapBillingError(err)
	}

	return c.JSON(item)
}

func (h *Handlers) RedeemGiftCode(c *fiber.Ctx) error {
	if err := requireBillingCodesRedeem(c); err != nil {
		return err
	}
	currentTenant, err := currentTenantOrErr(c)
	if err != nil {
		return err
	}
	var req platform.RedeemCodeRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	item, err := h.wallets.RedeemGiftCode(c.Context(), currentTenant.ID, req.Code, billingActorID(c))
	if err != nil {
		return mapBillingError(err)
	}

	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handlers) RedeemDiscountCode(c *fiber.Ctx) error {
	if err := requireBillingCodesRedeem(c); err != nil {
		return err
	}
	currentTenant, err := currentTenantOrErr(c)
	if err != nil {
		return err
	}
	var req platform.RedeemCodeRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	item, err := h.wallets.RedeemDiscountCode(c.Context(), currentTenant.ID, req.Code, billingActorID(c))
	if err != nil {
		return mapBillingError(err)
	}

	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handlers) GetBillingProfile(c *fiber.Ctx) error {
	if err := requireBillingRead(c); err != nil {
		return err
	}
	currentTenant, err := currentTenantOrErr(c)
	if err != nil {
		return err
	}
	item, err := h.wallets.GetBillingProfile(c.Context(), currentTenant.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(fiber.Map{"data": nil})
		}
		return err
	}

	return c.JSON(item)
}

func (h *Handlers) UpsertBillingProfile(c *fiber.Ctx) error {
	if err := requireBillingProfileManage(c); err != nil {
		return err
	}
	currentTenant, err := currentTenantOrErr(c)
	if err != nil {
		return err
	}

	var req platform.UpsertBillingProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	item, err := h.wallets.UpsertBillingProfile(c.Context(), currentTenant.ID, req, billingActorID(c))
	if err != nil {
		return mapBillingError(err)
	}

	return c.JSON(item)
}

func (h *Handlers) GetBillingUsageSummary(c *fiber.Ctx) error {
	if err := requireBillingRead(c); err != nil {
		return err
	}
	currentTenant, err := currentTenantOrErr(c)
	if err != nil {
		return err
	}
	item, err := h.wallets.GetBillingUsageSummary(c.Context(), currentTenant.ID, c.Query("range", "current_month"))
	if err != nil {
		return err
	}

	return c.JSON(item)
}

func (h *Handlers) GetBillingMonthlyUsage(c *fiber.Ctx) error {
	if err := requireBillingRead(c); err != nil {
		return err
	}
	currentTenant, err := currentTenantOrErr(c)
	if err != nil {
		return err
	}
	item, err := h.wallets.GetMonthlyUsageDetail(c.Context(), currentTenant.ID, c.Query("month", time.Now().UTC().Format("2006-01")))
	if err != nil {
		return err
	}

	return c.JSON(item)
}

func (h *Handlers) ListBillingInvoices(c *fiber.Ctx) error {
	if err := requireBillingDocumentsRead(c); err != nil {
		return err
	}
	currentTenant, err := currentTenantOrErr(c)
	if err != nil {
		return err
	}
	items, err := h.wallets.ListInvoices(c.Context(), currentTenant.ID, c.QueryInt("limit", 50))
	if err != nil {
		return err
	}

	return c.JSON(platform.InvoiceListResponse{Data: items})
}

func (h *Handlers) ListBillingTransactions(c *fiber.Ctx) error {
	if err := requireBillingDocumentsRead(c); err != nil {
		return err
	}
	currentTenant, err := currentTenantOrErr(c)
	if err != nil {
		return err
	}
	items, err := h.wallets.ListWalletTransactions(c.Context(), currentTenant.ID, c.QueryInt("limit", 50))
	if err != nil {
		return err
	}

	return c.JSON(platform.TransactionListResponse{Data: items})
}

func (h *Handlers) ListBillingDiscounts(c *fiber.Ctx) error {
	if err := requireBillingDocumentsRead(c); err != nil {
		return err
	}
	currentTenant, err := currentTenantOrErr(c)
	if err != nil {
		return err
	}
	items, err := h.wallets.ListDiscountRedemptions(c.Context(), currentTenant.ID, c.QueryInt("limit", 50))
	if err != nil {
		return err
	}

	return c.JSON(platform.DiscountRedemptionListResponse{Data: items})
}

func billingActorID(c *fiber.Ctx) string {
	authInfo := auth.GetAuthInfo(c)
	if authInfo == nil {
		return ""
	}

	return strings.TrimSpace(authInfo.UserID)
}

func mapBillingError(err error) error {
	switch {
	case errors.Is(err, storage.ErrFinancialCodeInvalid):
		return fiber.NewError(fiber.StatusUnprocessableEntity, "code is invalid or not redeemable")
	case errors.Is(err, storage.ErrBillingProfileInvalid):
		return fiber.NewError(fiber.StatusUnprocessableEntity, "billing profile is invalid")
	case errors.Is(err, storage.ErrBankTransferNotFound):
		return fiber.NewError(fiber.StatusNotFound, "bank transfer request not found")
	case errors.Is(err, storage.ErrInsufficientWalletBalance):
		return fiber.NewError(fiber.StatusConflict, "insufficient wallet balance")
	case errors.Is(err, storage.ErrInvalidTenantStatusChange):
		return fiber.NewError(fiber.StatusConflict, err.Error())
	default:
		return err
	}
}
