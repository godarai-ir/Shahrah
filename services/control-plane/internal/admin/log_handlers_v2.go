package admin

import (
	"errors"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

func (h *Handlers) ListRequestLogsV2(c *fiber.Ctx) error {
	if err := requireLogsMetadataRead(c); err != nil {
		return err
	}
	if h.logs == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "log store not configured")
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	params, err := parseLogListV2Params(c, currentTenant.ID)
	if err != nil {
		return err
	}

	items, total, err := h.logs.ListRequestLogsV2(c.Context(), params)
	if err != nil {
		if errors.Is(err, storage.ErrLogStoreUnavailable) {
			return fiber.NewError(fiber.StatusNotImplemented, "log store not configured")
		}
		return err
	}

	var nextOffset *int
	if params.Offset+len(items) < total {
		n := params.Offset + len(items)
		nextOffset = &n
	}

	return c.JSON(platform.ListRequestLogsV2Response{
		Data:       items,
		Total:      total,
		Limit:      params.Limit,
		Offset:     params.Offset,
		NextOffset: nextOffset,
	})
}

func (h *Handlers) GetRequestLogV2(c *fiber.Ctx) error {
	if err := requireLogsMetadataRead(c); err != nil {
		return err
	}
	if h.logs == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "log store not configured")
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	logID := strings.TrimSpace(c.Params("logID"))
	if logID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "logID is required")
	}

	includeContent := requireLogsContentRead(c) == nil

	item, err := h.logs.GetRequestLogV2(c.Context(), currentTenant.ID, logID, includeContent)
	if err != nil {
		if errors.Is(err, storage.ErrLogStoreUnavailable) {
			return fiber.NewError(fiber.StatusNotImplemented, "log store not configured")
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "log not found")
		}
		return err
	}

	return c.JSON(item)
}

func (h *Handlers) GetRequestLogFacetsV2(c *fiber.Ctx) error {
	if err := requireLogsMetadataRead(c); err != nil {
		return err
	}
	if h.logs == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "log store not configured")
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	params, err := parseLogListV2Params(c, currentTenant.ID)
	if err != nil {
		return err
	}
	params.Limit = 0
	params.Offset = 0

	facets, err := h.logs.GetRequestLogFacetsV2(c.Context(), params)
	if err != nil {
		if errors.Is(err, storage.ErrLogStoreUnavailable) {
			return fiber.NewError(fiber.StatusNotImplemented, "log store not configured")
		}
		return err
	}

	return c.JSON(platform.LogFacetsV2Response{Data: *facets})
}

func parseLogListV2Params(c *fiber.Ctx, tenantID string) (platform.ListRequestLogsV2Params, error) {
	limit, err := parsePositiveQueryInt(c, "limit", 50, 200, "limit must be a positive number")
	if err != nil {
		return platform.ListRequestLogsV2Params{}, err
	}
	offset, err := parseNonNegativeQueryInt(c, "offset", 0, "offset must be zero or greater")
	if err != nil {
		return platform.ListRequestLogsV2Params{}, err
	}
	statusCode, err := parseOptionalStatusCode(c)
	if err != nil {
		return platform.ListRequestLogsV2Params{}, err
	}
	from, err := parseOptionalRFC3339Query(c, "from")
	if err != nil {
		return platform.ListRequestLogsV2Params{}, err
	}
	to, err := parseOptionalRFC3339Query(c, "to")
	if err != nil {
		return platform.ListRequestLogsV2Params{}, err
	}

	return platform.ListRequestLogsV2Params{
		TenantID:       tenantID,
		LogID:          strings.TrimSpace(c.Query("log_id")),
		RequestID:      strings.TrimSpace(c.Query("request_id")),
		TraceID:        strings.TrimSpace(c.Query("trace_id")),
		Provider:       strings.TrimSpace(c.Query("provider")),
		Model:          strings.TrimSpace(c.Query("model")),
		VirtualModel:   strings.TrimSpace(c.Query("virtual_model")),
		Route:          strings.TrimSpace(c.Query("route")),
		RequestType:    strings.TrimSpace(c.Query("request_type")),
		ErrorType:      strings.TrimSpace(c.Query("error_type")),
		CacheStatus:    strings.TrimSpace(c.Query("cache_status")),
		FallbackStatus: strings.TrimSpace(c.Query("fallback_status")),
		StatusCode:     statusCode,
		From:           from,
		To:             to,
		Limit:          limit,
		Offset:         offset,
		SortBy:         strings.TrimSpace(c.Query("sort_by")),
		SortOrder:      strings.TrimSpace(c.Query("sort_order")),
	}, nil
}
