package admin

import (
	"errors"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
)

func (h *Handlers) ListRequestTracesV2(c *fiber.Ctx) error {
	return listObservabilityItems(
		c, requireTenantAdminOrHigher, h.traces != nil, "trace store not configured",
		func(tenantID string) (platform.ListRequestTracesV2Params, error) {
			return parseTraceListV2Params(c, tenantID)
		},
		func(params platform.ListRequestTracesV2Params) ([]platform.RequestTraceV2, int, error) {
			return h.traces.ListRequestTracesV2(c.Context(), params)
		},
		observability.ErrTraceStoreUnavailable,
		func(params platform.ListRequestTracesV2Params) (int, int) { return params.Limit, params.Offset },
		func(items []platform.RequestTraceV2, total, limit, offset int, nextOffset *int) error {
			return c.JSON(platform.ListRequestTracesV2Response{
				Data: items, Total: total, Limit: limit, Offset: offset, NextOffset: nextOffset,
			})
		},
	)
}

func (h *Handlers) GetTraceFacetsV2(c *fiber.Ctx) error {
	if err := requireTenantAdminOrHigher(c); err != nil {
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
		if errors.Is(err, observability.ErrTraceStoreUnavailable) {
			return fiber.NewError(fiber.StatusNotImplemented, "trace store not configured")
		}

		return err
	}

	return c.JSON(platform.TraceFacetsV2Response{Data: *facets})
}

func (h *Handlers) GetRequestTraceV2(c *fiber.Ctx) error {
	if h.traces == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "trace store not configured")
	}

	return handleTraceByID(c, requireTenantAdminOrHigher, h.traces.GetRequestTraceV2, func(item *platform.RequestTraceDetailV2) interface{} {
		return item
	})
}

func (h *Handlers) ListRequestTraceTimelineV2(c *fiber.Ctx) error {
	if h.traces == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "trace store not configured")
	}

	return handleTraceCollection(c, requireTenantAdminOrHigher, h.traces.ListTraceTimelineV2, func(items []platform.TraceSpan) interface{} {
		return platform.TraceTimelineV2Response{Data: items}
	})
}

func (h *Handlers) ListRequestTraceAttemptsV2(c *fiber.Ctx) error {
	if h.traces == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "trace store not configured")
	}

	return handleTraceCollection(c, requireTenantAdminOrHigher, h.traces.ListTraceAttemptsV2, func(items []platform.TraceAttemptV2) interface{} {
		return platform.TraceAttemptsV2Response{Data: items}
	})
}

func parseTraceListV2Params(c *fiber.Ctx, tenantID string) (platform.ListRequestTracesV2Params, error) {
	limit, err := parsePositiveQueryInt(c, "limit", 50, 200, "limit must be a positive number")
	if err != nil {
		return platform.ListRequestTracesV2Params{}, err
	}

	offset, err := parseNonNegativeQueryInt(c, "offset", 0, "offset must be zero or greater")
	if err != nil {
		return platform.ListRequestTracesV2Params{}, err
	}

	statusCode, err := parseOptionalStatusCode(c)
	if err != nil {
		return platform.ListRequestTracesV2Params{}, err
	}

	from, err := parseOptionalRFC3339Query(c, "from")
	if err != nil {
		return platform.ListRequestTracesV2Params{}, err
	}

	to, err := parseOptionalRFC3339Query(c, "to")
	if err != nil {
		return platform.ListRequestTracesV2Params{}, err
	}

	from, to, err = applyObservabilityTimeBounds(from, to, time.Now().UTC())
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
