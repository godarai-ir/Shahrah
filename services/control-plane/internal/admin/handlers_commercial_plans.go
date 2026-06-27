package admin

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

func (h *Handlers) ListPublicCommercialPlans(c *fiber.Ctx) error {
	store, err := h.requirePlanStore()
	if err != nil {
		return err
	}

	items, err := store.ListPublicCommercialPlans(c.Context())
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"data": items})
}

func (h *Handlers) ListPlatformFeatureCatalog(c *fiber.Ctx) error {
	store, err := h.requirePlanStore()
	if err != nil {
		return err
	}

	items, err := store.ListFeatureCatalog(c.Context())
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"data": items})
}

func (h *Handlers) ListPlatformCommercialPlans(c *fiber.Ctx) error {
	store, err := h.requirePlanStore()
	if err != nil {
		return err
	}

	items, err := store.ListPlatformCommercialPlans(c.Context())
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"data": items})
}

func (h *Handlers) CreatePlatformCommercialPlan(c *fiber.Ctx) error {
	store, err := h.requirePlanStore()
	if err != nil {
		return err
	}

	var req platform.CreateCommercialPlanRequest

	parseErr := c.BodyParser(&req)
	if parseErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if strings.TrimSpace(req.Slug) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "slug is required")
	}

	item, err := store.CreateCommercialPlan(c.Context(), req)
	if err != nil {
		return mapPlanStoreError(err)
	}

	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handlers) UpdatePlatformCommercialPlan(c *fiber.Ctx) error {
	store, err := h.requirePlanStore()
	if err != nil {
		return err
	}

	var req platform.UpdateCommercialPlanRequest

	parseErr := c.BodyParser(&req)
	if parseErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	item, err := store.UpdateCommercialPlan(c.Context(), c.Params("slug"), req)
	if err != nil {
		return mapPlanStoreError(err)
	}

	return c.JSON(item)
}

func (h *Handlers) DeactivatePlatformCommercialPlan(c *fiber.Ctx) error {
	store, err := h.requirePlanStore()
	if err != nil {
		return err
	}

	if err := store.DeactivateCommercialPlan(c.Context(), c.Params("slug")); err != nil {
		return mapPlanStoreError(err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handlers) CreatePlatformPlanVersion(c *fiber.Ctx) error {
	store, err := h.requirePlanStore()
	if err != nil {
		return err
	}

	var req platform.CreatePlanVersionRequest
	if len(c.Body()) > 0 {
		parseErr := c.BodyParser(&req)
		if parseErr != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
		}
	}

	item, err := store.CreatePlanVersion(c.Context(), c.Params("slug"), req)
	if err != nil {
		return mapPlanStoreError(err)
	}

	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handlers) PublishPlatformPlanVersion(c *fiber.Ctx) error {
	store, err := h.requirePlanStore()
	if err != nil {
		return err
	}

	version, err := parsePlanVersionParam(c)
	if err != nil {
		return err
	}

	item, err := store.PublishPlanVersion(c.Context(), c.Params("slug"), version)
	if err != nil {
		return mapPlanStoreError(err)
	}

	return c.JSON(item)
}

func (h *Handlers) UpsertPlatformPlanEntitlement(c *fiber.Ctx) error {
	store, err := h.requirePlanStore()
	if err != nil {
		return err
	}

	version, err := parsePlanVersionParam(c)
	if err != nil {
		return err
	}

	var req platform.UpsertPlanEntitlementRequest

	parseErr := c.BodyParser(&req)
	if parseErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if req.FeatureKey == "" {
		req.FeatureKey = strings.TrimSpace(c.Params("featureKey"))
	}

	if strings.TrimSpace(req.FeatureKey) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "feature_key is required")
	}

	item, err := store.UpsertPlanEntitlement(c.Context(), c.Params("slug"), version, req)
	if err != nil {
		return mapPlanStoreError(err)
	}

	return c.JSON(item)
}

func (h *Handlers) UpsertPlatformPlanPrice(c *fiber.Ctx) error {
	store, err := h.requirePlanStore()
	if err != nil {
		return err
	}

	version, err := parsePlanVersionParam(c)
	if err != nil {
		return err
	}

	var req platform.UpsertPlanPriceRequest

	parseErr := c.BodyParser(&req)
	if parseErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if req.BillingInterval == "" {
		req.BillingInterval = strings.TrimSpace(c.Params("interval"))
	}

	item, err := store.UpsertPlanPrice(c.Context(), c.Params("slug"), version, req)
	if err != nil {
		return mapPlanStoreError(err)
	}

	return c.JSON(item)
}

func (h *Handlers) GetTenantSubscription(c *fiber.Ctx) error {
	store, err := h.requirePlanStore()
	if err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	item, err := store.GetTenantSubscription(c.Context(), currentTenant.ID)
	if err != nil {
		return mapPlanStoreError(err)
	}

	return c.JSON(item)
}

func (h *Handlers) GetTenantEntitlements(c *fiber.Ctx) error {
	store, err := h.requirePlanStore()
	if err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	item, err := store.ResolveEffectiveEntitlements(c.Context(), currentTenant.ID)
	if err != nil {
		return mapPlanStoreError(err)
	}

	return c.JSON(item)
}

func (h *Handlers) ChangeTenantSubscription(c *fiber.Ctx) error {
	store, err := h.requirePlanStore()
	if err != nil {
		return err
	}

	permissionErr := requirePermissionAnyOrTenantAdmin(c, "subscription management permission required", []string{"settings:manage"})
	if permissionErr != nil {
		return permissionErr
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	var req platform.ChangeTenantSubscriptionRequest

	parseErr := c.BodyParser(&req)
	if parseErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	providerName := strings.ToLower(strings.TrimSpace(req.Provider))
	if providerName == "" {
		providerName = defaultSubscriptionPaymentProvider()
	}

	req.Provider = providerName

	returnURL := strings.TrimSpace(req.ReturnURL)
	if returnURL == "" {
		returnURL = strings.TrimSuffix(resolveExternalBaseURL(c), "/") + "/billing/plan"
	}

	callbackURL := strings.TrimSuffix(resolveExternalBaseURL(c), "/") + "/api/v1/subscription/payment/callback/" + providerName

	result, err := store.PrepareTenantSubscriptionChange(c.Context(), currentTenant.ID, req, callbackURL, returnURL)
	if err != nil {
		return mapPlanStoreError(err)
	}

	if result.PaymentIntent == nil {
		return c.JSON(result)
	}

	provider := selectProvider(providerName)
	payable := result.PaymentIntent.AmountMinor
	paymentInput := &platform.WalletTopUpIntent{
		PublicID: result.PaymentIntent.PublicID, TenantID: result.PaymentIntent.TenantID,
		Provider: providerName, PayableIRR: payable,
		CallbackURL: result.PaymentIntent.CallbackURL, ReturnURL: result.PaymentIntent.ReturnURL,
	}

	payRes, err := provider.CreatePayment(c, paymentInput)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}

	if err := store.AttachSubscriptionPaymentGatewayData(c.Context(), result.PaymentIntent.PublicID, payRes.Authority, payRes.PaymentURL, payRes.Raw); err != nil {
		return mapPlanStoreError(err)
	}

	result.PaymentIntent.Authority = payRes.Authority
	result.PaymentIntent.PaymentURL = payRes.PaymentURL
	result.PaymentIntent.RawPayload = payRes.Raw

	return c.Status(fiber.StatusCreated).JSON(result)
}

func (h *Handlers) SubscriptionPaymentCallback(c *fiber.Ctx) error {
	store, err := h.requirePlanStore()
	if err != nil {
		return err
	}

	providerName := strings.ToLower(strings.TrimSpace(c.Params("provider")))

	authority := strings.TrimSpace(c.Query("authority"))
	if authority == "" {
		authority = strings.TrimSpace(c.FormValue("authority"))
	}

	if authority == "" {
		return fiber.NewError(fiber.StatusBadRequest, "authority is required")
	}

	intent, err := store.GetSubscriptionPaymentIntentByAuthority(c.Context(), authority)
	if err != nil {
		return h.redirectSubscriptionPaymentResult(c, store, authority, "failed", "intent_not_found")
	}

	c.Locals("topup_intent", &platform.WalletTopUpIntent{
		PublicID: intent.PublicID, TenantID: intent.TenantID, Provider: intent.Provider,
		PayableIRR: intent.AmountMinor, CallbackURL: intent.CallbackURL, ReturnURL: intent.ReturnURL,
	})

	verifyRes, err := selectProvider(providerName).VerifyPayment(c, authority)
	if err != nil {
		_ = store.MarkSubscriptionPaymentFailedByAuthority(c.Context(), authority, err.Error(), map[string]interface{}{"provider": providerName})
		return h.redirectSubscriptionPaymentResult(c, store, authority, "failed", "verify_error")
	}

	if !verifyRes.Success {
		reason := verifyRes.Failure
		if reason == "" {
			reason = "payment failed"
		}

		_ = store.MarkSubscriptionPaymentFailedByAuthority(c.Context(), authority, reason, verifyRes.Raw)

		return h.redirectSubscriptionPaymentResult(c, store, authority, "failed", "payment_failed")
	}

	completed, _, _, err := store.CompleteSubscriptionPaymentByAuthority(c.Context(), authority, verifyRes.Raw, verifyRes.RefID, verifyRes.GatewayTxnID, verifyRes.CardMask, verifyRes.PayerPhone)
	if err != nil {
		return h.redirectSubscriptionPaymentResult(c, store, authority, "failed", "commit_failed")
	}

	return c.Redirect(subscriptionPaymentReturnURL(completed.ReturnURL, "success", completed.PublicID, authority, completed.RefID, ""), fiber.StatusFound)
}

func (h *Handlers) redirectSubscriptionPaymentResult(c *fiber.Ctx, store planStore, authority string, status string, reason string) error {
	intent, err := store.GetSubscriptionPaymentIntentByAuthority(c.Context(), authority)
	if err != nil || intent == nil || strings.TrimSpace(intent.ReturnURL) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": status, "reason": reason, "authority": authority})
	}

	return c.Redirect(subscriptionPaymentReturnURL(intent.ReturnURL, status, intent.PublicID, authority, intent.RefID, reason), fiber.StatusFound)
}

func subscriptionPaymentReturnURL(returnURL string, status string, intentID string, authority string, refID string, reason string) string {
	separator := "?"
	if strings.Contains(returnURL, "?") {
		separator = "&"
	}

	return fmt.Sprintf("%s%spayment_status=%s&intent_id=%s&authority=%s&ref_id=%s&reason=%s", returnURL, separator,
		url.QueryEscape(status), url.QueryEscape(intentID), url.QueryEscape(authority), url.QueryEscape(refID), url.QueryEscape(reason))
}

func (h *Handlers) UpsertPlatformTenantSubscription(c *fiber.Ctx) error {
	store, err := h.requirePlanStore()
	if err != nil {
		return err
	}

	var req platform.UpsertTenantSubscriptionRequest

	parseErr := c.BodyParser(&req)
	if parseErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if strings.TrimSpace(req.TenantID) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "tenant_id is required")
	}

	item, err := store.UpsertTenantSubscription(c.Context(), req, platformActor(c))
	if err != nil {
		return mapPlanStoreError(err)
	}

	return c.JSON(item)
}

func (h *Handlers) UpsertPlatformTenantEntitlementOverride(c *fiber.Ctx) error {
	store, err := h.requirePlanStore()
	if err != nil {
		return err
	}

	var req platform.UpsertTenantEntitlementOverrideRequest

	parseErr := c.BodyParser(&req)
	if parseErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if strings.TrimSpace(req.TenantID) == "" || strings.TrimSpace(req.FeatureKey) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "tenant_id and feature_key are required")
	}

	if req.CreatedBy == "" {
		req.CreatedBy = platformActor(c)
	}

	item, err := store.UpsertTenantEntitlementOverride(c.Context(), req)
	if err != nil {
		return mapPlanStoreError(err)
	}

	return c.JSON(item)
}

func (h *Handlers) DeletePlatformTenantEntitlementOverride(c *fiber.Ctx) error {
	store, err := h.requirePlanStore()
	if err != nil {
		return err
	}

	tenantID := strings.TrimSpace(c.Params("tenantID"))

	featureKey := strings.TrimSpace(c.Params("featureKey"))
	if tenantID == "" || featureKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "tenantID and featureKey are required")
	}

	if err := store.DeleteTenantEntitlementOverride(c.Context(), tenantID, featureKey, platformActor(c)); err != nil {
		return mapPlanStoreError(err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handlers) requirePlanStore() (planStore, error) {
	if h.plans == nil {
		return nil, fiber.NewError(fiber.StatusNotImplemented, "commercial plan store not configured")
	}

	return h.plans, nil
}

func parsePlanVersionParam(c *fiber.Ctx) (int, error) {
	raw := strings.TrimSpace(c.Params("version"))

	version, err := strconv.Atoi(raw)
	if err != nil || version <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "version must be a positive integer")
	}

	return version, nil
}

func platformActor(c *fiber.Ctx) string {
	if authInfo := auth.GetAuthInfo(c); authInfo != nil && authInfo.UserID != "" {
		return authInfo.UserID
	}

	if actor := strings.TrimSpace(c.Get("x-shahrah-user-id")); actor != "" {
		return actor
	}

	return "platform_admin"
}

func mapPlanStoreError(err error) error {
	switch {
	case errors.Is(err, storage.ErrPlanNotFound), errors.Is(err, storage.ErrPlanVersionNotFound), errors.Is(err, storage.ErrSubscriptionNotFound), errors.Is(err, pgx.ErrNoRows):
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	case errors.Is(err, storage.ErrPlanVersionPublished), errors.Is(err, storage.ErrFeatureOverrideForbidden):
		return fiber.NewError(fiber.StatusForbidden, err.Error())
	case errors.Is(err, storage.ErrFeatureNotFound), errors.Is(err, storage.ErrPlanPriceUnavailable):
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, storage.ErrSubscriptionPaymentIntentNotFound):
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	default:
		return err
	}
}
