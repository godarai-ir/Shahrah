package admin

import (
	"encoding/json"
	"fmt"
	"math"
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

func (h *Handlers) requireEntitlementLimitAvailable(c *fiber.Ctx, tenantID string, featureKey string, used int, requested int) error {
	if h.entitlements == nil {
		return nil
	}

	if requested < 1 {
		requested = 1
	}

	entitlements, err := h.entitlements.ResolveEffectiveEntitlements(c.Context(), tenantID)
	if err != nil {
		return mapPlanStoreError(err)
	}

	if entitlements == nil || entitlements.Subscription == nil || !subscriptionEntitlementsUsable(entitlements.Subscription.Status) {
		return fiber.NewError(fiber.StatusForbidden, "active subscription required")
	}

	limit, limited := entitlementNumber(entitlements.Features[featureKey])
	if !limited {
		return nil
	}

	if float64(used+requested) > limit {
		return fiber.NewError(fiber.StatusForbidden, fmt.Sprintf("feature %s limit exceeded for this plan", featureKey))
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

func entitlementNumber(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case nil:
		return 0, false
	case int:
		return normalizeEntitlementLimit(float64(v))
	case int64:
		return normalizeEntitlementLimit(float64(v))
	case float64:
		return normalizeEntitlementLimit(v)
	case json.Number:
		parsed, err := v.Float64()
		if err != nil {
			return 0, false
		}

		return normalizeEntitlementLimit(parsed)
	case string:
		raw := strings.TrimSpace(strings.ToLower(v))
		if raw == "" || raw == "custom" || raw == "unlimited" || raw == "null" {
			return 0, false
		}

		parsed, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return 0, false
		}

		return normalizeEntitlementLimit(parsed)
	default:
		return 0, false
	}
}

func normalizeEntitlementLimit(value float64) (float64, bool) {
	if value < 0 || math.IsInf(value, 0) || math.IsNaN(value) {
		return 0, false
	}

	return value, true
}

func (h *Handlers) requireVATEnabled(c *fiber.Ctx, tenantID string) error {
	return h.requireEntitlementEnabled(c, tenantID, platform.FeatureVATEnabled)
}
