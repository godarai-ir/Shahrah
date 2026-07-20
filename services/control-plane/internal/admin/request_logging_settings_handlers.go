package admin

import (
	"errors"

	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

type updateRequestLoggingSettingsRequest struct {
	LoggingMode string `json:"logging_mode"`
}

func (h *Handlers) GetTenantRequestLoggingSettings(c *fiber.Ctx) error {
	if err := requireTeamAdminOrHigher(c); err != nil {
		return err
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	if h.requestLogging == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "request logging store not configured")
	}

	item, err := h.requestLogging.GetTenantRequestLoggingSettings(c.Context(), currentTenant.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(fiber.Map{
				"tenant_id":    currentTenant.ID,
				"logging_mode": storage.RequestLoggingModeMetadataOnly,
			})
		}
		return err
	}

	return c.JSON(item)
}

func (h *Handlers) UpdateTenantRequestLoggingSettings(c *fiber.Ctx) error {
	if err := requireTeamAdminOrHigher(c); err != nil {
		return err
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	if h.requestLogging == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "request logging store not configured")
	}

	var req updateRequestLoggingSettingsRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	loggingMode := storage.NormalizeRequestLoggingMode(req.LoggingMode)
	switch loggingMode {
	case storage.RequestLoggingModeNone, storage.RequestLoggingModeMetadataOnly, storage.RequestLoggingModeRedactedContent, storage.RequestLoggingModeFullContent:
	default:
		return fiber.NewError(fiber.StatusBadRequest, "invalid logging_mode")
	}

	item, err := h.requestLogging.UpsertTenantRequestLoggingSettings(
		c.Context(),
		currentTenant.ID,
		loggingMode,
		delegatedUserID(c),
	)
	if err != nil {
		return err
	}

	return c.JSON(item)
}
