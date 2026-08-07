package admin

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/apperrors"
	billingdomain "github.com/Delpak/shahrah/services/control-plane/internal/domain/billing"
	supportdomain "github.com/Delpak/shahrah/services/control-plane/internal/domain/support"
	"github.com/gofiber/fiber/v2"
)

func (h *Handlers) ListPlatformTenants(c *fiber.Ctx) error {
	store, err := h.requireSupportTenantStore()
	if err != nil {
		return err
	}

	limit := parseSupportLimit(c.QueryInt("limit", 25))

	offset, err := decodeSupportCursor(c.Query("cursor"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid cursor")
	}

	items, err := store.ListSupportTenants(c.Context(), c.Query("query"), c.Query("status"), c.Query("plan"), limit+1, offset)
	if err != nil {
		return err
	}

	response := platform.SupportTenantListResponse{Data: items}
	if len(items) > limit {
		response.Data = items[:limit]
		response.NextCursor = encodeSupportCursor(offset + limit)
	}

	return c.JSON(response)
}

func (h *Handlers) GetPlatformTenant(c *fiber.Ctx) error {
	store, err := h.requireSupportTenantStore()
	if err != nil {
		return err
	}

	tenantID := strings.TrimSpace(c.Params("tenantID"))

	summary, err := store.GetSupportTenantSummary(c.Context(), tenantID)
	if err != nil {
		return mapSupportTenantError(err)
	}

	detail := platform.SupportTenantDetail{Tenant: *summary}

	detail.Owner, err = store.GetTenantOwnerContact(c.Context(), tenantID)
	if err != nil {
		return err
	}

	if h.wallets != nil {
		wallet, walletErr := h.wallets.GetWallet(c.Context(), tenantID)
		if walletErr == nil {
			detail.Wallet = wallet
		}
	}

	if h.plans != nil {
		subscription, subscriptionErr := h.plans.GetTenantSubscription(c.Context(), tenantID)
		if subscriptionErr == nil {
			detail.Subscription = subscription
		}

		entitlements, entitlementsErr := h.plans.ResolveEffectiveEntitlements(c.Context(), tenantID)
		if entitlementsErr == nil {
			detail.Entitlements = entitlements
		}

		if detail.Entitlements != nil {
			for _, counter := range detail.Entitlements.UsageCounters {
				if counter.FeatureKey == platform.FeatureGatewayRequestsMonthly {
					detail.Tenant.UsageUsed = counter.Used
					detail.Tenant.UsageLimit = counter.Limit
				}
			}
		}
	}

	return c.JSON(detail)
}

func (h *Handlers) ListPlatformTenantEvents(c *fiber.Ctx) error {
	store, err := h.requireSupportTenantStore()
	if err != nil {
		return err
	}

	limit := parseSupportLimit(c.QueryInt("limit", 50))

	offset, err := decodeSupportCursor(c.Query("cursor"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid cursor")
	}

	items, err := store.ListTenantEvents(c.Context(), c.Params("tenantID"), limit+1, offset)
	if err != nil {
		return err
	}

	next := ""

	if len(items) > limit {
		items = items[:limit]
		next = encodeSupportCursor(offset + limit)
	}

	return c.JSON(fiber.Map{"data": items, "next_cursor": next})
}

func (h *Handlers) GetPlatformTenantWallet(c *fiber.Ctx) error {
	if h.wallets == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "wallet store not configured")
	}

	item, err := h.wallets.GetWallet(c.Context(), c.Params("tenantID"))
	if err != nil {
		return err
	}

	return c.JSON(item)
}

func (h *Handlers) ListPlatformTenantWalletTransactions(c *fiber.Ctx) error {
	if h.wallets == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "wallet store not configured")
	}

	items, err := h.wallets.ListWalletTransactions(c.Context(), c.Params("tenantID"), parseSupportLimit(c.QueryInt("limit", 50)))
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"data": items})
}

func (h *Handlers) AdjustPlatformTenantWallet(c *fiber.Ctx) error {
	store, err := h.requireSupportTenantStore()
	if err != nil {
		return err
	}

	var req platform.WalletAdjustmentRequest

	parseErr := c.BodyParser(&req)
	if parseErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	req.Direction = strings.ToLower(strings.TrimSpace(req.Direction))
	req.Reason = strings.TrimSpace(req.Reason)

	req.IdempotencyKey = strings.TrimSpace(req.IdempotencyKey)
	if (req.Direction != "credit" && req.Direction != "debit") || req.AmountMinor <= 0 || req.Reason == "" || req.IdempotencyKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "direction, positive amount_minor, reason, and idempotency_key are required")
	}

	if len(req.Reason) > 1000 || len(req.IdempotencyKey) > 128 || len(req.ExternalReference) > 255 {
		return fiber.NewError(fiber.StatusBadRequest, "request field is too long")
	}

	result, err := store.AdjustTenantWallet(c.Context(), c.Params("tenantID"), req, supportActor(c))
	if err != nil {
		return mapSupportTenantError(err)
	}

	return c.JSON(result)
}

func (h *Handlers) SuspendPlatformTenant(c *fiber.Ctx) error {
	return h.changePlatformTenantStatus(c, "suspended")
}

func (h *Handlers) ReactivatePlatformTenant(c *fiber.Ctx) error {
	return h.changePlatformTenantStatus(c, "active")
}

func (h *Handlers) SchedulePlatformTenantSubscriptionCancellation(c *fiber.Ctx) error {
	return h.setPlatformTenantSubscriptionCancellation(c, true)
}

func (h *Handlers) ReversePlatformTenantSubscriptionCancellation(c *fiber.Ctx) error {
	return h.setPlatformTenantSubscriptionCancellation(c, false)
}

func (h *Handlers) setPlatformTenantSubscriptionCancellation(c *fiber.Ctx, cancel bool) error {
	store, err := h.requireSupportTenantStore()
	if err != nil {
		return err
	}

	var req platform.SubscriptionOperatorActionRequest

	parseErr := c.BodyParser(&req)
	if parseErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	req.Reason = strings.TrimSpace(req.Reason)
	if req.Reason == "" || len(req.Reason) > 1000 {
		return fiber.NewError(fiber.StatusBadRequest, "reason is required and must not exceed 1000 characters")
	}

	item, err := store.SetTenantSubscriptionCancellation(c.Context(), c.Params("tenantID"), cancel, req.Reason, supportActor(c))
	if err != nil {
		return mapSupportTenantError(err)
	}

	return c.JSON(item)
}

func (h *Handlers) ListPlatformTenantSubscriptionPaymentIntents(c *fiber.Ctx) error {
	store, err := h.requireSupportTenantStore()
	if err != nil {
		return err
	}

	limit := parseSupportLimit(c.QueryInt("limit", 50))

	offset, err := decodeSupportCursor(c.Query("cursor"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid cursor")
	}

	items, err := store.ListTenantSubscriptionPaymentIntents(c.Context(), c.Params("tenantID"), limit, offset)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"data": items})
}

func (h *Handlers) changePlatformTenantStatus(c *fiber.Ctx, target string) error {
	store, err := h.requireSupportTenantStore()
	if err != nil {
		return err
	}

	var req platform.TenantStatusChangeRequest

	parseErr := c.BodyParser(&req)
	if parseErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	req.Reason = strings.TrimSpace(req.Reason)
	if req.Reason == "" || len(req.Reason) > 1000 {
		return fiber.NewError(fiber.StatusBadRequest, "reason is required and must not exceed 1000 characters")
	}

	item, err := store.ChangeTenantStatus(c.Context(), c.Params("tenantID"), target, req.Reason, supportActor(c))
	if err != nil {
		return mapSupportTenantError(err)
	}

	return c.JSON(item)
}

func (h *Handlers) requireSupportTenantStore() (supportTenantOperationsStore, error) {
	if h.supportTenants == nil {
		return nil, fiber.NewError(fiber.StatusNotImplemented, "support tenant operations store not configured")
	}

	return h.supportTenants, nil
}

func supportActor(c *fiber.Ctx) platform.SupportActor {
	return platform.SupportActor{
		ID: strings.TrimSpace(c.Get("x-shahrah-actor-id")), Role: strings.TrimSpace(c.Get("x-shahrah-actor-role")),
		RequestID: strings.TrimSpace(c.Get("x-request-id")),
	}
}

func mapSupportTenantError(err error) error {
	switch {
	case errors.Is(err, supportdomain.ErrTenantNotFound), errors.Is(err, apperrors.ErrNotFound):
		return fiber.NewError(fiber.StatusNotFound, "tenant not found")
	case errors.Is(err, billingdomain.ErrInsufficientWalletBalance):
		return fiber.NewError(fiber.StatusConflict, "insufficient wallet balance")
	case errors.Is(err, supportdomain.ErrIdempotencyPayloadConflict), errors.Is(err, supportdomain.ErrInvalidTenantStatusChange):
		return fiber.NewError(fiber.StatusConflict, err.Error())
	default:
		return err
	}
}

func parseSupportLimit(value int) int {
	if value < 1 {
		return 25
	}

	if value > 100 {
		return 100
	}

	return value
}

func encodeSupportCursor(offset int) string {
	return base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf("offset:%d", offset)))
}

func decodeSupportCursor(cursor string) (int, error) {
	if strings.TrimSpace(cursor) == "" {
		return 0, nil
	}

	raw, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil || !strings.HasPrefix(string(raw), "offset:") {
		return 0, errors.New("invalid cursor")
	}

	offset, err := strconv.Atoi(strings.TrimPrefix(string(raw), "offset:"))
	if err != nil || offset < 0 {
		return 0, errors.New("invalid cursor")
	}

	return offset, nil
}
