package admin

import (
	"context"
	"errors"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	billingdomain "github.com/Delpak/shahrah/services/control-plane/internal/domain/billing"
	"github.com/gofiber/fiber/v2"
)

// creditPackStore sells credit add-on packs and keeps tenant credit settings.
type creditPackStore interface {
	ListCreditPacks(ctx context.Context) ([]platform.CreditPack, error)
	CreateCreditPackOrder(ctx context.Context, tenantID string, req platform.CreateCreditPackOrderRequest, callbackURL, returnURL, actor string) (*platform.CreditPackOrder, error)
	AttachCreditPackOrderGatewayData(ctx context.Context, publicID, authority, paymentURL string, raw map[string]interface{}) error
	GetCreditPackOrder(ctx context.Context, tenantID, publicID string) (*platform.CreditPackOrder, error)
	GetCreditPackOrderByAuthority(ctx context.Context, authority string) (*platform.CreditPackOrder, error)
	ListCreditPackOrders(ctx context.Context, tenantID string, limit int) ([]platform.CreditPackOrder, error)
	CompleteCreditPackOrderByAuthority(ctx context.Context, authority string, verified billingdomain.PaymentVerification) (*platform.CreditPackOrder, bool, error)
	MarkCreditPackOrderFailedByAuthority(ctx context.Context, authority, reason string, raw map[string]interface{}) error
	GetCreditSettings(ctx context.Context, tenantID string) (*platform.CreditSettings, error)
	UpdateCreditSettings(ctx context.Context, tenantID string, req platform.UpdateCreditSettingsRequest, actor platform.SupportActor) (*platform.CreditSettings, error)
}

const creditPackCallbackPath = "/api/v1/credits/packs/payment/callback/"

func (h *Handlers) requireCreditPackStore() (creditPackStore, error) {
	if h.creditPacks == nil {
		return nil, fiber.NewError(fiber.StatusNotImplemented, "credit packs are not configured")
	}

	return h.creditPacks, nil
}

// ListCreditPacks returns the purchasable add-on packs with their current prices.
func (h *Handlers) ListCreditPacks(c *fiber.Ctx) error {
	if err := requireBillingRead(c); err != nil {
		return err
	}

	store, err := h.requireCreditPackStore()
	if err != nil {
		return err
	}

	packs, err := store.ListCreditPacks(c.Context())
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"data": packs})
}

// ListCreditPackOrders returns the tenant's recent pack orders.
func (h *Handlers) ListCreditPackOrders(c *fiber.Ctx) error {
	if err := requireBillingRead(c); err != nil {
		return err
	}

	store, currentTenant, err := h.creditPackContext(c)
	if err != nil {
		return err
	}

	orders, err := store.ListCreditPackOrders(c.Context(), currentTenant.ID, c.QueryInt("limit", 50))
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"data": orders})
}

// GetCreditPackOrder returns one of the tenant's pack orders.
func (h *Handlers) GetCreditPackOrder(c *fiber.Ctx) error {
	if err := requireBillingRead(c); err != nil {
		return err
	}

	store, currentTenant, err := h.creditPackContext(c)
	if err != nil {
		return err
	}

	order, err := store.GetCreditPackOrder(c.Context(), currentTenant.ID, c.Params("orderID"))
	if err != nil {
		return creditPackError(c, err)
	}

	return c.JSON(order)
}

// CreateCreditPackOrder starts a pack purchase. Gateway orders return the
// payment URL; bank-transfer orders wait for a transfer that names the order.
func (h *Handlers) CreateCreditPackOrder(c *fiber.Ctx) error {
	if err := requireBillingManage(c); err != nil {
		return err
	}

	store, currentTenant, err := h.creditPackContext(c)
	if err != nil {
		return err
	}

	var req platform.CreateCreditPackOrderRequest
	if parseErr := c.BodyParser(&req); parseErr != nil || strings.TrimSpace(req.PackSlug) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "pack_slug is required")
	}

	req.Provider = strings.ToLower(strings.TrimSpace(req.Provider))
	if req.Provider == "" {
		req.Provider = h.payment.DefaultSubscriptionProvider
	}

	if !h.creditPackProviderAllowed(req.Provider) {
		return creditPackError(c, billingdomain.ErrCreditPaymentProviderUnavailable)
	}

	baseURL := strings.TrimSuffix(resolveExternalBaseURL(c), "/")

	returnURL := strings.TrimSpace(req.ReturnURL)
	if returnURL == "" {
		returnURL = baseURL + "/billing/credits"
	}

	order, err := store.CreateCreditPackOrder(c.Context(), currentTenant.ID, req, baseURL+creditPackCallbackPath+req.Provider, returnURL, billingActorID(c))
	if err != nil {
		return creditPackError(c, err)
	}

	if order.Status == platform.CreditPackOrderAwaitingTransfer {
		return c.Status(fiber.StatusCreated).JSON(order)
	}

	payment, err := h.selectPaymentProvider(order.Provider).CreatePayment(c, &platform.WalletTopUpIntent{
		PublicID: order.PublicID, TenantID: order.TenantID, Provider: order.Provider,
		PayableIRR: float64(order.PayableIRR), CallbackURL: order.CallbackURL, ReturnURL: order.ReturnURL,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, "payment gateway is unavailable")
	}

	if err := store.AttachCreditPackOrderGatewayData(c.Context(), order.PublicID, payment.Authority, payment.PaymentURL, payment.Raw); err != nil {
		return creditPackError(c, err)
	}

	order.Authority = payment.Authority
	order.PaymentURL = payment.PaymentURL

	return c.Status(fiber.StatusCreated).JSON(order)
}

// creditPackProviderAllowed accepts real gateways and bank transfer. The mock
// gateway approves every payment, so it is never accepted in production, and
// elsewhere only where an operator made it the default provider.
func (h *Handlers) creditPackProviderAllowed(provider string) bool {
	switch provider {
	case "zarinpal", "idpay", platform.CreditPackProviderBankTransfer:
		return true
	case "mock":
		return !h.payment.Production && strings.EqualFold(strings.TrimSpace(h.payment.DefaultSubscriptionProvider), "mock")
	default:
		return false
	}
}

// CreditPackPaymentCallback completes a gateway payment for a pack order. The
// order's own provider verifies the payment; the URL's provider is not trusted.
func (h *Handlers) CreditPackPaymentCallback(c *fiber.Ctx) error {
	store, err := h.requireCreditPackStore()
	if err != nil {
		return err
	}

	authority := strings.TrimSpace(c.Query("authority"))
	if authority == "" {
		authority = strings.TrimSpace(c.FormValue("authority"))
	}

	if authority == "" {
		return fiber.NewError(fiber.StatusBadRequest, "authority is required")
	}

	order, err := store.GetCreditPackOrderByAuthority(c.Context(), authority)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "failed", "reason": "order_not_found"})
	}

	setTopUpIntent(c, &platform.WalletTopUpIntent{
		PublicID: order.PublicID, TenantID: order.TenantID, Provider: order.Provider,
		PayableIRR: float64(order.PayableIRR), CallbackURL: order.CallbackURL, ReturnURL: order.ReturnURL,
	})

	verified, err := h.selectPaymentProvider(order.Provider).VerifyPayment(c, authority)
	if err != nil || !verified.Success {
		reason := "payment failed"

		var raw map[string]interface{}

		if verified != nil {
			raw = verified.Raw
			if verified.Failure != "" {
				reason = verified.Failure
			}
		}

		if markErr := store.MarkCreditPackOrderFailedByAuthority(c.Context(), authority, reason, raw); markErr != nil {
			return markErr
		}

		return c.Redirect(subscriptionPaymentReturnURL(order.ReturnURL, "failed", order.PublicID, authority, "", "payment_failed"), fiber.StatusFound)
	}

	completed, _, err := store.CompleteCreditPackOrderByAuthority(c.Context(), authority, billingdomain.PaymentVerification{
		Raw: verified.Raw, RefID: verified.RefID, GatewayTxnID: verified.GatewayTxnID,
		CardMask: verified.CardMask, PayerPhone: verified.PayerPhone,
	})
	if err != nil {
		return c.Redirect(subscriptionPaymentReturnURL(order.ReturnURL, "failed", order.PublicID, authority, "", "commit_failed"), fiber.StatusFound)
	}

	return c.Redirect(subscriptionPaymentReturnURL(completed.ReturnURL, "success", completed.PublicID, authority, completed.RefID, ""), fiber.StatusFound)
}

// GetCreditSettings returns the tenant's credit billing preferences.
func (h *Handlers) GetCreditSettings(c *fiber.Ctx) error {
	if err := requireBillingRead(c); err != nil {
		return err
	}

	store, currentTenant, err := h.creditPackContext(c)
	if err != nil {
		return err
	}

	settings, err := store.GetCreditSettings(c.Context(), currentTenant.ID)
	if err != nil {
		return creditPackError(c, err)
	}

	return c.JSON(settings)
}

// UpdateCreditSettings changes the tenant's seat preferences. Overage lets a
// tenant run into debt, so only support may change it.
func (h *Handlers) UpdateCreditSettings(c *fiber.Ctx) error {
	if err := requireBillingManage(c); err != nil {
		return err
	}

	store, currentTenant, err := h.creditPackContext(c)
	if err != nil {
		return err
	}

	var req platform.UpdateCreditSettingsRequest
	if parseErr := c.BodyParser(&req); parseErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if req.OverageEnabled != nil || req.OverageCapCredits != nil {
		return creditPackError(c, billingdomain.ErrOverageManagedBySupport)
	}

	settings, err := store.UpdateCreditSettings(c.Context(), currentTenant.ID, req, platform.SupportActor{ID: billingActorID(c), Role: "tenant"})
	if err != nil {
		return creditPackError(c, err)
	}

	return c.JSON(settings)
}

func (h *Handlers) creditPackContext(c *fiber.Ctx) (creditPackStore, *platform.Tenant, error) {
	store, err := h.requireCreditPackStore()
	if err != nil {
		return nil, nil, err
	}

	currentTenant, err := currentTenantOrErr(c)
	if err != nil {
		return nil, nil, err
	}

	return store, currentTenant, nil
}

// creditPackError maps credit pack errors to client-safe responses with a
// stable code.
func creditPackError(c *fiber.Ctx, err error) error {
	status, code := creditPackErrorStatus(err)
	if status == 0 {
		return err
	}

	return c.Status(status).JSON(fiber.Map{"error": fiber.Map{"code": code, "type": "credit_billing", "message": err.Error()}})
}

func creditPackErrorStatus(err error) (int, string) {
	switch {
	case errors.Is(err, billingdomain.ErrCreditBillingNotEnabled):
		return fiber.StatusConflict, "credit_billing_not_enabled"
	case errors.Is(err, billingdomain.ErrCreditPackNotFound):
		return fiber.StatusNotFound, "credit_pack_not_found"
	case errors.Is(err, billingdomain.ErrCreditPackOrderNotFound):
		return fiber.StatusNotFound, "credit_pack_order_not_found"
	case errors.Is(err, billingdomain.ErrCreditPackOrderNotPayable):
		return fiber.StatusConflict, "credit_pack_order_not_payable"
	case errors.Is(err, billingdomain.ErrCreditPackTransferMismatch):
		return fiber.StatusUnprocessableEntity, "credit_pack_transfer_amount_mismatch"
	case errors.Is(err, billingdomain.ErrCreditPaymentProviderUnavailable):
		return fiber.StatusUnprocessableEntity, "payment_provider_unavailable"
	case errors.Is(err, billingdomain.ErrOverageNotAllowed):
		return fiber.StatusForbidden, "overage_not_allowed"
	case errors.Is(err, billingdomain.ErrOverageManagedBySupport):
		return fiber.StatusForbidden, "overage_managed_by_support"
	case errors.Is(err, billingdomain.ErrInvalidOverageCap):
		return fiber.StatusBadRequest, "invalid_overage_cap"
	case errors.Is(err, billingdomain.ErrInvalidSeatPoolLimit):
		return fiber.StatusBadRequest, "invalid_seat_pool_limit"
	case errors.Is(err, billingdomain.ErrWalletTopUpForCreditTenant):
		return fiber.StatusConflict, "credit_billing_uses_packs"
	default:
		return creditAdminErrorStatus(err)
	}
}

func creditAdminErrorStatus(err error) (int, string) {
	switch {
	case errors.Is(err, billingdomain.ErrCreditPriceRevisionNotFound):
		return fiber.StatusNotFound, "credit_price_revision_not_found"
	case errors.Is(err, billingdomain.ErrDecisionReasonRequired):
		return fiber.StatusBadRequest, "reason_required"
	case errors.Is(err, billingdomain.ErrCreditRateCardNotDraft):
		return fiber.StatusConflict, "rate_card_not_draft"
	case errors.Is(err, billingdomain.ErrInvalidCreditMultiplier):
		return fiber.StatusBadRequest, "invalid_multiplier"
	case errors.Is(err, billingdomain.ErrCreditPackVersionNotFound):
		return fiber.StatusNotFound, "credit_pack_version_not_found"
	case errors.Is(err, billingdomain.ErrInvalidCreditPack):
		return fiber.StatusBadRequest, "invalid_credit_pack"
	case errors.Is(err, billingdomain.ErrInvalidCreditAdjustment):
		return fiber.StatusBadRequest, "invalid_credit_adjustment"
	default:
		return 0, ""
	}
}
