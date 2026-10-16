package admin

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/gofiber/fiber/v2"
)

func (h *Handlers) requireEntitlementEnabled(c *fiber.Ctx, tenantID string, featureKey string) error {
	if h.entitlements == nil {
		return nil
	}

	entitlements, err := h.entitlements.ResolveEffectiveEntitlements(c.Context(), tenantID)
	if err != nil {
		return mapPlanStoreError(err)
	}

	if entitlements == nil || entitlements.Subscription == nil || !subscriptionEntitlementsUsable(entitlements.Subscription.Status) {
		return fiber.NewError(fiber.StatusForbidden, "active subscription required")
	}

	enabled, ok := entitlementBool(entitlements.Features[featureKey])
	if !ok || !enabled {
		return fiber.NewError(fiber.StatusForbidden, fmt.Sprintf("feature %s is not enabled for this plan", featureKey))
	}

	return nil
}

// requireActiveSubscription refuses tenant changes that need a current plan.
func (h *Handlers) requireActiveSubscription(c *fiber.Ctx, tenantID string) error {
	if h.entitlements == nil {
		return nil
	}

	entitlements, err := h.entitlements.ResolveEffectiveEntitlements(c.Context(), tenantID)
	if err != nil {
		return mapPlanStoreError(err)
	}

	if entitlements == nil || entitlements.Subscription == nil || !subscriptionEntitlementsUsable(entitlements.Subscription.Status) {
		return fiber.NewError(fiber.StatusForbidden, "active subscription required")
	}

	return nil
}

func subscriptionEntitlementsUsable(status string) bool {
	switch strings.TrimSpace(status) {
	case "active", "trialing":
		return true
	default:
		return false
	}
}

func entitlementBool(value interface{}) (bool, bool) {
	switch v := value.(type) {
	case bool:
		return v, true
	case string:
		parsed, err := strconv.ParseBool(strings.TrimSpace(v))
		return parsed, err == nil
	default:
		return false, false
	}
}

func (h *Handlers) requireVATEnabled(c *fiber.Ctx, tenantID string) error {
	return h.requireEntitlementEnabled(c, tenantID, platform.FeatureVATEnabled)
}
