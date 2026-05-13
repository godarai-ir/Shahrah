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
	"github.com/jackc/pgx/v5"
)

func (h *Handlers) ListRequestTracesV2(c *fiber.Ctx) error {
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

	params, err := parseTraceListV2Params(c, currentTenant.ID)
	if err != nil {
		return err
	}
	items, total, err := h.traces.ListRequestTracesV2(c.Context(), params)
	if err != nil {
		if errors.Is(err, storage.ErrTraceStoreUnavailable) {
			return fiber.NewError(fiber.StatusNotImplemented, "trace store not configured")
		}
		return err
	}

	var nextOffset *int
	if params.Offset+len(items) < total {
		n := params.Offset + len(items)
		nextOffset = &n
	}
	return c.JSON(platform.ListRequestTracesV2Response{
		Data:       items,
		Total:      total,
		Limit:      params.Limit,
		Offset:     params.Offset,
		NextOffset: nextOffset,
	})
}

func (h *Handlers) GetTraceFacetsV2(c *fiber.Ctx) error {
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
	params, err := parseTraceListV2Params(c, currentTenant.ID)
	if err != nil {
		return err
	}
	// Facets should reflect the active filter set, not pagination.
	params.Limit = 0
	params.Offset = 0
	facets, err := h.traces.GetTraceFacetsV2(c.Context(), params)
	if err != nil {
		if errors.Is(err, storage.ErrTraceStoreUnavailable) {
			return fiber.NewError(fiber.StatusNotImplemented, "trace store not configured")
		}
		return err
	}
	return c.JSON(platform.TraceFacetsV2Response{Data: *facets})
}

func (h *Handlers) GetRequestTraceV2(c *fiber.Ctx) error {
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
	traceID := strings.TrimSpace(c.Params("traceID"))
	if traceID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "traceID is required")
	}
	item, err := h.traces.GetRequestTraceV2(c.Context(), currentTenant.ID, traceID)
	if err != nil {
		if errors.Is(err, storage.ErrTraceStoreUnavailable) {
			return fiber.NewError(fiber.StatusNotImplemented, "trace store not configured")
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "trace not found")
		}
		return err
	}
	return c.JSON(item)
}

func (h *Handlers) ListRequestTraceTimelineV2(c *fiber.Ctx) error {
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
	traceID := strings.TrimSpace(c.Params("traceID"))
	if traceID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "traceID is required")
	}
	items, err := h.traces.ListTraceTimelineV2(c.Context(), currentTenant.ID, traceID)
	if err != nil {
		if errors.Is(err, storage.ErrTraceStoreUnavailable) {
			return fiber.NewError(fiber.StatusNotImplemented, "trace store not configured")
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "trace not found")
		}
		return err
	}
	return c.JSON(platform.TraceTimelineV2Response{Data: items})
}

func (h *Handlers) ListRequestTraceAttemptsV2(c *fiber.Ctx) error {
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
	traceID := strings.TrimSpace(c.Params("traceID"))
	if traceID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "traceID is required")
	}
	items, err := h.traces.ListTraceAttemptsV2(c.Context(), currentTenant.ID, traceID)
	if err != nil {
		if errors.Is(err, storage.ErrTraceStoreUnavailable) {
			return fiber.NewError(fiber.StatusNotImplemented, "trace store not configured")
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "trace not found")
		}
		return err
	}
	return c.JSON(platform.TraceAttemptsV2Response{Data: items})
}

func parseTraceListV2Params(c *fiber.Ctx, tenantID string) (platform.ListRequestTracesV2Params, error) {
	limit := 50
	offset := 0
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v <= 0 {
			return platform.ListRequestTracesV2Params{}, fiber.NewError(fiber.StatusBadRequest, "limit must be a positive number")
		}
		if v > 200 {
			v = 200
		}
		limit = v
	}
	if raw := strings.TrimSpace(c.Query("offset")); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 0 {
			return platform.ListRequestTracesV2Params{}, fiber.NewError(fiber.StatusBadRequest, "offset must be zero or greater")
		}
		offset = v
	}
	var statusCode *int
	if raw := strings.TrimSpace(c.Query("status_code")); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 100 || v > 599 {
			return platform.ListRequestTracesV2Params{}, fiber.NewError(fiber.StatusBadRequest, "status_code must be a valid HTTP status")
		}
		statusCode = &v
	}
	parseTime := func(key string) (*time.Time, error) {
		raw := strings.TrimSpace(c.Query(key))
		if raw == "" {
			return nil, nil
		}
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, key+" must be RFC3339")
		}
		return &t, nil
	}
	from, err := parseTime("from")
	if err != nil {
		return platform.ListRequestTracesV2Params{}, err
	}
	to, err := parseTime("to")
	if err != nil {
		return platform.ListRequestTracesV2Params{}, err
	}
	return platform.ListRequestTracesV2Params{
		TenantID:   tenantID,
		TraceID:    strings.TrimSpace(c.Query("trace_id")),
		RequestID:  strings.TrimSpace(c.Query("request_id")),
		Route:      strings.TrimSpace(c.Query("route")),
		Model:      strings.TrimSpace(c.Query("model")),
		Provider:   strings.TrimSpace(c.Query("provider")),
		ErrorCode:  strings.TrimSpace(c.Query("error_code")),
		StatusCode: statusCode,
		From:       from,
		To:         to,
		Limit:      limit,
		Offset:     offset,
		SortBy:     strings.TrimSpace(c.Query("sort_by")),
		SortOrder:  strings.TrimSpace(c.Query("sort_order")),
	}, nil
}
