package admin

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
)

func (h *Handlers) ListRequestTraces(c *fiber.Ctx) error {
	if err := requireTeamAdminOrHigher(c); err != nil {
		return err
	}

	if h.traces == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "trace store not configured")
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	limit, err := parsePositiveQueryInt(c, "limit", 50, 200, "limit must be a positive number")
	if err != nil {
		return err
	}

	offset, err := parseNonNegativeQueryInt(c, "offset", 0, "offset must be zero or greater")
	if err != nil {
		return err
	}

	statusCode, err := parseOptionalStatusCode(c)
	if err != nil {
		return err
	}

	from, err := parseOptionalRFC3339Query(c, "from")
	if err != nil {
		return err
	}

	to, err := parseOptionalRFC3339Query(c, "to")
	if err != nil {
		return err
	}

	items, total, err := h.traces.ListRequestTraces(c.Context(), platform.ListRequestTracesParams{
		TenantID:   currentTenant.ID,
		TraceID:    strings.TrimSpace(c.Query("trace_id")),
		Route:      strings.TrimSpace(c.Query("route")),
		Model:      strings.TrimSpace(c.Query("model")),
		StatusCode: statusCode,
		From:       from,
		To:         to,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		if errors.Is(err, storage.ErrTraceStoreUnavailable) {
			return fiber.NewError(fiber.StatusNotImplemented, "trace store not configured")
		}

		return err
	}

	var nextOffset *int

	if offset+len(items) < total {
		n := offset + len(items)
		nextOffset = &n
	}

	return c.JSON(platform.ListRequestTracesResponse{
		Data:       items,
		Total:      total,
		Limit:      limit,
		Offset:     offset,
		NextOffset: nextOffset,
	})
}

func parsePositiveQueryInt(c *fiber.Ctx, key string, fallback int, max int, message string) (int, error) {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, message)
	}

	if max > 0 && value > max {
		value = max
	}

	return value, nil
}

func parseNonNegativeQueryInt(c *fiber.Ctx, key string, fallback int, message string) (int, error) {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, message)
	}

	return value, nil
}

func parseOptionalStatusCode(c *fiber.Ctx) (*int, error) {
	raw := strings.TrimSpace(c.Query("status_code"))
	if raw == "" {
		return nil, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value < 100 || value > 599 {
		return nil, fiber.NewError(fiber.StatusBadRequest, "status_code must be a valid HTTP status")
	}

	return &value, nil
}

func parseOptionalRFC3339Query(c *fiber.Ctx, key string) (*time.Time, error) {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return nil, nil
	}

	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, key+" must be RFC3339")
	}

	return &parsed, nil
}

func (h *Handlers) GetRequestTrace(c *fiber.Ctx) error {
	if h.traces == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "trace store not configured")
	}

	return handleTraceByID(c, requireTeamAdminOrHigher, h.traces.GetRequestTrace, func(item *platform.RequestTrace) interface{} {
		return item
	})
}

func (h *Handlers) ListRequestTraceSpans(c *fiber.Ctx) error {
	if h.traces == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "trace store not configured")
	}

	return handleTraceCollection(c, requireTeamAdminOrHigher, h.traces.ListTraceSpans, func(items []platform.TraceSpan) interface{} {
		return platform.TraceSpansResponse{Data: items}
	})
}
