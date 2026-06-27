package admin

import (
	"errors"
	"strings"

	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

type updateTracePrivacySettingsRequest struct {
	CapturePolicy string `json:"capture_policy"`
}

func (h *Handlers) GetTenantTracePrivacySettings(c *fiber.Ctx) error {
	if err := requireTeamAdminOrHigher(c); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	if h.tracePrivacy == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "trace privacy store not configured")
	}

	item, err := h.tracePrivacy.GetTenantTracePrivacySettings(c.Context(), currentTenant.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(fiber.Map{
				"tenant_id":      currentTenant.ID,
				"capture_policy": "metadata-only",
			})
		}

		return err
	}

	return c.JSON(item)
}

func (h *Handlers) UpdateTenantTracePrivacySettings(c *fiber.Ctx) error {
	if err := requireTeamAdminOrHigher(c); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	if h.tracePrivacy == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "trace privacy store not configured")
	}

	var req updateTracePrivacySettingsRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	req.CapturePolicy = strings.TrimSpace(strings.ToLower(req.CapturePolicy))
	switch req.CapturePolicy {
	case "none", "metadata-only", "sampled-redacted", "full-secure":
	default:
		return fiber.NewError(fiber.StatusBadRequest, "capture_policy must be one of: none, metadata-only, sampled-redacted, full-secure")
	}

	updatedBy := strings.TrimSpace(c.Get("x-shahrah-user-id"))

	item, err := h.tracePrivacy.UpsertTenantTracePrivacySettings(c.Context(), currentTenant.ID, req.CapturePolicy, updatedBy)
	if err != nil {
		return err
	}

	return c.JSON(item)
}
