package admin

import (
	"context"
	"errors"
	"mime/multipart"
	"strconv"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	billingdomain "github.com/Delpak/shahrah/services/control-plane/internal/domain/billing"
	supportdomain "github.com/Delpak/shahrah/services/control-plane/internal/domain/support"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/Delpak/shahrah/services/control-plane/internal/supportfiles"
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

	var (
		req           platform.CreateBankTransferRequest
		receiptHeader *multipart.FileHeader
	)
	if strings.HasPrefix(strings.ToLower(c.Get(fiber.HeaderContentType)), "multipart/form-data") {
		form, parseErr := c.MultipartForm()
		if parseErr != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid multipart request body")
		}

		amountIRR, parseErr := strconv.ParseInt(strings.TrimSpace(firstFormValue(form.Value, "amount_irr")), 10, 64)
		if parseErr != nil {
			return fiber.NewError(fiber.StatusBadRequest, "amount_irr must be an integer")
		}

		req.AmountIRR = amountIRR
		req.TrackingNumber = strings.TrimSpace(firstFormValue(form.Value, "tracking_number"))
		req.BankTransactionDateJalali = strings.TrimSpace(firstFormValue(form.Value, "bank_transaction_date_jalali"))

		receipts := form.File["receipt_image"]
		if len(receipts) > 1 {
			return fiber.NewError(fiber.StatusBadRequest, "only one receipt_image is allowed")
		}

		if len(receipts) == 1 && receipts[0] != nil && receipts[0].Size > 0 {
			receiptHeader = receipts[0]
		}
	} else if parseErr := c.BodyParser(&req); parseErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	req.TrackingNumber = strings.TrimSpace(req.TrackingNumber)

	req.ReceiptFileURL = strings.TrimSpace(req.ReceiptFileURL)
	if req.ReceiptFileURL != "" {
		return fiber.NewError(fiber.StatusBadRequest, "receipt_file_url must be created by uploading receipt_image")
	}

	req.BankTransactionDateJalali = strings.TrimSpace(req.BankTransactionDateJalali)
	if req.AmountIRR <= 0 || req.TrackingNumber == "" || req.BankTransactionDateJalali == "" {
		return fiber.NewError(fiber.StatusBadRequest, "amount_irr, tracking_number, and bank_transaction_date_jalali are required")
	}

	bankTransactionAt, err := storage.ParseJalaliDateAtTehranMidday(req.BankTransactionDateJalali)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "bank_transaction_date_jalali must be a valid Jalali date in YYYY-MM-DD format")
	}

	req.BankTransactionAt = bankTransactionAt

	var receipt *supportfiles.ReceiptUpload

	if receiptHeader != nil {
		if h.supportFiles == nil {
			return fiber.NewError(fiber.StatusServiceUnavailable, "private file storage is not configured")
		}

		receipt, err = h.supportFiles.UploadReceiptImage(c.Context(), currentTenant.ID, receiptHeader)
		if err != nil {
			return mapBankTransferReceiptError(err)
		}

		req.ReceiptFileURL = "/api/wallet/bank-transfer-receipts/" + receipt.FileName
	}

	item, err := h.wallets.CreateBankTransferRequest(c.Context(), currentTenant.ID, req, billingActorID(c))
	if err != nil {
		if receipt != nil {
			h.supportFiles.DeleteObject(context.WithoutCancel(c.Context()), receipt.ObjectKey)
		}

		return err
	}

	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handlers) DownloadBankTransferReceipt(c *fiber.Ctx) error {
	if err := requireBillingRead(c); err != nil {
		return err
	}

	if h.wallets == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "wallet store not configured")
	}

	currentTenant, err := currentTenantOrErr(c)
	if err != nil {
		return err
	}

	fileName := strings.TrimSpace(c.Params("fileName"))

	contentType, ok := supportfiles.ValidateBankTransferReceiptFileName(fileName)
	if !ok {
		return fiber.NewError(fiber.StatusNotFound, "bank transfer receipt not found")
	}

	if h.supportFiles == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "private file storage is not configured")
	}

	receiptURL := "/api/wallet/bank-transfer-receipts/" + fileName

	exists, err := h.wallets.BankTransferReceiptExists(c.Context(), currentTenant.ID, receiptURL)
	if err != nil {
		return err
	}

	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "bank transfer receipt not found")
	}

	objectKey := supportfiles.BankTransferReceiptObjectKey(currentTenant.ID, fileName)

	object, err := h.supportFiles.Get(c.Context(), objectKey)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "bank transfer receipt not found")
	}

	c.Set(fiber.HeaderContentType, contentType)
	c.Set(fiber.HeaderCacheControl, "private, max-age=3600")
	c.Set("X-Content-Type-Options", "nosniff")
	c.Context().SetBodyStream(object.Body, -1)

	return nil
}

func firstFormValue(values map[string][]string, key string) string {
	items := values[key]
	if len(items) == 0 {
		return ""
	}

	return items[0]
}

func mapBankTransferReceiptError(err error) error {
	switch {
	case errors.Is(err, supportfiles.ErrFileTooLarge):
		return fiber.NewError(fiber.StatusRequestEntityTooLarge, "receipt image must be at most 2 MiB")
	case errors.Is(err, supportfiles.ErrInvalidReceipt), errors.Is(err, supportfiles.ErrTypeNotAllowed), errors.Is(err, supportfiles.ErrInfected):
		return fiber.NewError(fiber.StatusUnprocessableEntity, "receipt image is invalid")
	case errors.Is(err, supportfiles.ErrNotConfigured), errors.Is(err, supportfiles.ErrScanFailed):
		return fiber.NewError(fiber.StatusServiceUnavailable, "private file storage is unavailable")
	default:
		return fiber.NewError(fiber.StatusBadGateway, "receipt image could not be stored")
	}
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
	if parseErr := c.BodyParser(&req); parseErr != nil {
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
	if parseErr := c.BodyParser(&req); parseErr != nil {
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
	if parseErr := c.BodyParser(&req); parseErr != nil {
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
	if parseErr := c.BodyParser(&req); parseErr != nil {
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

	if strings.TrimSpace(c.Query("range")) != "" {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "range query parameter is no longer supported; use preset")
	}

	preset := strings.TrimSpace(c.Query("preset"))
	if preset == "" {
		return fiber.NewError(fiber.StatusBadRequest, "preset is required")
	}

	item, err := h.wallets.GetBillingUsageSummary(c.Context(), currentTenant.ID, preset)
	if err != nil {
		if errors.Is(err, storage.ErrBillingReportPresetInvalid) {
			return fiber.NewError(fiber.StatusBadRequest, "preset is invalid")
		}

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

	if strings.TrimSpace(c.Query("month")) != "" {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "month query parameter is no longer supported; use jalali_year and jalali_month")
	}

	jalaliYear, err := strconv.Atoi(strings.TrimSpace(c.Query("jalali_year")))
	if err != nil || jalaliYear <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "jalali_year must be a positive integer")
	}

	jalaliMonth, err := strconv.Atoi(strings.TrimSpace(c.Query("jalali_month")))
	if err != nil || jalaliMonth <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "jalali_month must be a positive integer")
	}

	item, err := h.wallets.GetMonthlyUsageDetail(c.Context(), currentTenant.ID, jalaliYear, jalaliMonth)
	if err != nil {
		if errors.Is(err, storage.ErrBillingJalaliMonthInvalid) {
			return fiber.NewError(fiber.StatusBadRequest, "jalali_year and jalali_month must describe a valid Jalali month")
		}

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
	case errors.Is(err, billingdomain.ErrInsufficientWalletBalance):
		return fiber.NewError(fiber.StatusConflict, "insufficient wallet balance")
	case errors.Is(err, supportdomain.ErrInvalidTenantStatusChange):
		return fiber.NewError(fiber.StatusConflict, err.Error())
	default:
		return err
	}
}
