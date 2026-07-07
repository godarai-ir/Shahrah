package admin

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
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

func (h *Handlers) GetRequestLogSummaryV2(c *fiber.Ctx) error {
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

	summary, err := h.logs.GetRequestLogSummaryV2(c.Context(), params)
	if err != nil {
		if errors.Is(err, storage.ErrLogStoreUnavailable) {
			return fiber.NewError(fiber.StatusNotImplemented, "log store not configured")
		}
		return err
	}

	return c.JSON(platform.LogSummaryV2Response{Data: *summary})
}

func (h *Handlers) GetRequestLogRangeStatsV2(c *fiber.Ctx) error {
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

	stats, err := h.logs.GetRequestLogRangeStatsV2(c.Context(), params)
	if err != nil {
		if errors.Is(err, storage.ErrLogStoreUnavailable) {
			return fiber.NewError(fiber.StatusNotImplemented, "log store not configured")
		}
		return err
	}

	return c.JSON(platform.LogRangeStatsV2Response{Data: *stats})
}

func (h *Handlers) GetRequestLogAnalyticsV2(c *fiber.Ctx) error {
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

	listParams, err := parseLogListV2Params(c, currentTenant.ID)
	if err != nil {
		return err
	}
	listParams.Limit = 0
	listParams.Offset = 0
	interval, err := parsePositiveQueryInt(c, "interval_seconds", 3600, 86400, "interval_seconds must be a positive number")
	if err != nil {
		return err
	}

	points, err := h.logs.GetRequestLogAnalyticsV2(c.Context(), platform.LogAnalyticsV2Params{
		ListRequestLogsV2Params: listParams,
		Type:                    strings.TrimSpace(c.Query("type")),
		GroupBy:                 strings.TrimSpace(c.Query("group_by")),
		IntervalSeconds:         interval,
	})
	if err != nil {
		if errors.Is(err, storage.ErrLogStoreUnavailable) {
			return fiber.NewError(fiber.StatusNotImplemented, "log store not configured")
		}
		return err
	}

	return c.JSON(platform.LogAnalyticsV2Response{Data: points})
}

func (h *Handlers) ListSavedLogViewsV2(c *fiber.Ctx) error {
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
	items, err := h.logs.ListSavedLogViewsV2(c.Context(), currentTenant.ID, currentUserID(c))
	if err != nil {
		return err
	}
	return c.JSON(platform.ListSavedLogViewsV2Response{Data: items})
}

func (h *Handlers) CreateSavedLogViewV2(c *fiber.Ctx) error {
	if err := requireLogsManageViews(c); err != nil {
		return err
	}
	if h.logs == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "log store not configured")
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	var req platform.CreateSavedLogViewV2Request
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid saved view payload")
	}
	if strings.TrimSpace(req.Name) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "saved view name is required")
	}
	item, err := h.logs.CreateSavedLogViewV2(c.Context(), currentTenant.ID, currentUserID(c), req)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handlers) UpdateSavedLogViewV2(c *fiber.Ctx) error {
	if err := requireLogsManageViews(c); err != nil {
		return err
	}
	if h.logs == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "log store not configured")
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	var req platform.UpdateSavedLogViewV2Request
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid saved view payload")
	}
	item, err := h.logs.UpdateSavedLogViewV2(c.Context(), currentTenant.ID, strings.TrimSpace(c.Params("viewID")), req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "saved view not found")
		}
		return err
	}
	return c.JSON(item)
}

func (h *Handlers) DeleteSavedLogViewV2(c *fiber.Ctx) error {
	if err := requireLogsManageViews(c); err != nil {
		return err
	}
	if h.logs == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "log store not configured")
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	err := h.logs.DeleteSavedLogViewV2(c.Context(), currentTenant.ID, strings.TrimSpace(c.Params("viewID")))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "saved view not found")
		}
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handlers) ListLogExportsV2(c *fiber.Ctx) error {
	if err := requireLogsExport(c); err != nil {
		return err
	}
	if h.logs == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "log store not configured")
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	limit, err := parsePositiveQueryInt(c, "limit", 25, 100, "limit must be a positive number")
	if err != nil {
		return err
	}
	items, err := h.logs.ListLogExportsV2(c.Context(), currentTenant.ID, limit)
	if err != nil {
		return err
	}
	return c.JSON(platform.ListLogExportsV2Response{Data: items})
}

func (h *Handlers) CreateLogExportV2(c *fiber.Ctx) error {
	if err := requireLogsExport(c); err != nil {
		return err
	}
	if h.logs == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "log store not configured")
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	var req platform.CreateLogExportV2Request
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid export payload")
	}
	if req.IncludeContent {
		if err := requireLogsContentRead(c); err != nil {
			return err
		}
	}
	item, err := h.logs.CreateLogExportV2(c.Context(), currentTenant.ID, currentUserID(c), req)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handlers) GetLogExportV2(c *fiber.Ctx) error {
	if err := requireLogsExport(c); err != nil {
		return err
	}
	item, err := h.getLogExportJob(c)
	if err != nil {
		return err
	}
	return c.JSON(item)
}

func (h *Handlers) CancelLogExportV2(c *fiber.Ctx) error {
	if err := requireLogsExport(c); err != nil {
		return err
	}
	if h.logs == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "log store not configured")
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	item, err := h.logs.CancelLogExportV2(c.Context(), currentTenant.ID, strings.TrimSpace(c.Params("exportID")))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "export not found")
		}
		return err
	}
	return c.JSON(item)
}

func (h *Handlers) DownloadLogExportV2(c *fiber.Ctx) error {
	if err := requireLogsExport(c); err != nil {
		return err
	}
	if h.logs == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "log store not configured")
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	job, err := h.getLogExportJob(c)
	if err != nil {
		return err
	}
	if job.IncludeContent {
		if err := requireLogsContentRead(c); err != nil {
			return err
		}
	}
	params := logParamsFromFilterMap(currentTenant.ID, job.Filters)
	if params.Limit <= 0 {
		params.Limit = 10000
	}
	if params.Limit > 10000 {
		params.Limit = 10000
	}
	items, _, err := h.logs.ListRequestLogsV2(c.Context(), params)
	if err != nil {
		if errors.Is(err, storage.ErrLogStoreUnavailable) {
			return fiber.NewError(fiber.StatusNotImplemented, "log store not configured")
		}
		return err
	}

	fileName := fmt.Sprintf("logs-%s.%s", job.ExportID, job.Format)
	c.Set(fiber.HeaderContentDisposition, fmt.Sprintf(`attachment; filename="%s"`, fileName))
	if job.Format == "jsonl" {
		c.Set(fiber.HeaderContentType, "application/x-ndjson; charset=utf-8")
		return c.Send(buildLogExportJSONL(items))
	}

	c.Set(fiber.HeaderContentType, "text/csv; charset=utf-8")
	return c.Send(buildLogExportCSV(items))
}

func (h *Handlers) CreateRequestLogReplayDraftV2(c *fiber.Ctx) error {
	if err := requireLogsReplay(c); err != nil {
		return err
	}
	if err := requireLogsContentRead(c); err != nil {
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
	detail, err := h.logs.GetRequestLogV2(c.Context(), currentTenant.ID, logID, true)
	if err != nil {
		if errors.Is(err, storage.ErrLogStoreUnavailable) {
			return fiber.NewError(fiber.StatusNotImplemented, "log store not configured")
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "log not found")
		}
		return err
	}

	draft := platform.LogReplayDraftV2{
		LogID:          detail.Log.LogID,
		Supported:      true,
		Route:          detail.Log.Route,
		RequestType:    detail.Log.RequestType,
		Model:          detail.Log.Model,
		Provider:       detail.Log.Provider,
		RequestPayload: detail.RequestPayload,
		SafeHeaders:    map[string]string{},
		Warnings:       []string{"sensitive headers and credentials are intentionally omitted", "draft only; no request is executed automatically"},
	}
	if !isReplaySupportedRoute(detail.Log.Route) {
		draft.Supported = false
		draft.Reason = "unsupported_route"
		draft.RequestPayload = nil
	}
	if !detail.ContentVisibility.RequestAvailable || len(detail.RequestPayload) == 0 {
		draft.Supported = false
		draft.Reason = "payload_hidden_or_unavailable"
	}

	return c.JSON(draft)
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
	from, to, err = applyObservabilityTimeBounds(from, to, time.Now().UTC())
	if err != nil {
		return platform.ListRequestLogsV2Params{}, err
	}
	durationMin, err := parseOptionalIntQuery(c, "duration_min", "duration_min must be a number")
	if err != nil {
		return platform.ListRequestLogsV2Params{}, err
	}
	durationMax, err := parseOptionalIntQuery(c, "duration_max", "duration_max must be a number")
	if err != nil {
		return platform.ListRequestLogsV2Params{}, err
	}
	tokensMin, err := parseOptionalIntQuery(c, "tokens_min", "tokens_min must be a number")
	if err != nil {
		return platform.ListRequestLogsV2Params{}, err
	}
	tokensMax, err := parseOptionalIntQuery(c, "tokens_max", "tokens_max must be a number")
	if err != nil {
		return platform.ListRequestLogsV2Params{}, err
	}
	costUSDMin, err := parseOptionalFloatQuery(c, "cost_usd_min", "cost_usd_min must be a number")
	if err != nil {
		return platform.ListRequestLogsV2Params{}, err
	}
	costUSDMax, err := parseOptionalFloatQuery(c, "cost_usd_max", "cost_usd_max must be a number")
	if err != nil {
		return platform.ListRequestLogsV2Params{}, err
	}
	costIRRMin, err := parseOptionalFloatQuery(c, "cost_irr_min", "cost_irr_min must be a number")
	if err != nil {
		return platform.ListRequestLogsV2Params{}, err
	}
	costIRRMax, err := parseOptionalFloatQuery(c, "cost_irr_max", "cost_irr_max must be a number")
	if err != nil {
		return platform.ListRequestLogsV2Params{}, err
	}

	return platform.ListRequestLogsV2Params{
		TenantID:          tenantID,
		LogID:             strings.TrimSpace(c.Query("log_id")),
		RequestID:         strings.TrimSpace(c.Query("request_id")),
		TraceID:           strings.TrimSpace(c.Query("trace_id")),
		Provider:          strings.TrimSpace(c.Query("provider")),
		Model:             strings.TrimSpace(c.Query("model")),
		VirtualModel:      strings.TrimSpace(c.Query("virtual_model")),
		Route:             strings.TrimSpace(c.Query("route")),
		RequestType:       strings.TrimSpace(c.Query("request_type")),
		ErrorType:         strings.TrimSpace(c.Query("error_type")),
		CacheStatus:       strings.TrimSpace(c.Query("cache_status")),
		FallbackStatus:    strings.TrimSpace(c.Query("fallback_status")),
		LoadBalanceStatus: strings.TrimSpace(c.Query("load_balance_status")),
		PayloadMode:       strings.TrimSpace(c.Query("payload_mode")),
		ActorType:         strings.TrimSpace(c.Query("actor_type")),
		ActorID:           strings.TrimSpace(c.Query("actor_id")),
		OrgUnitID:         strings.TrimSpace(c.Query("org_unit_id")),
		TeamID:            strings.TrimSpace(c.Query("team_id")),
		StatusCode:        statusCode,
		From:              from,
		To:                to,
		DurationMin:       durationMin,
		DurationMax:       durationMax,
		TokensMin:         tokensMin,
		TokensMax:         tokensMax,
		CostUSDMin:        costUSDMin,
		CostUSDMax:        costUSDMax,
		CostIRRMin:        costIRRMin,
		CostIRRMax:        costIRRMax,
		Limit:             limit,
		Offset:            offset,
		SortBy:            strings.TrimSpace(c.Query("sort_by")),
		SortOrder:         strings.TrimSpace(c.Query("sort_order")),
	}, nil
}

func parseOptionalIntQuery(c *fiber.Ctx, key string, message string) (*int, error) {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return nil, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, message)
	}
	return &value, nil
}

func parseOptionalFloatQuery(c *fiber.Ctx, key string, message string) (*float64, error) {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return nil, nil
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, message)
	}
	return &value, nil
}

func (h *Handlers) getLogExportJob(c *fiber.Ctx) (*platform.LogExportJobV2, error) {
	if h.logs == nil {
		return nil, fiber.NewError(fiber.StatusNotImplemented, "log store not configured")
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	item, err := h.logs.GetLogExportV2(c.Context(), currentTenant.ID, strings.TrimSpace(c.Params("exportID")))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fiber.NewError(fiber.StatusNotFound, "export not found")
		}
		return nil, err
	}
	return item, nil
}

func currentUserID(c *fiber.Ctx) string {
	if info := auth.GetAuthInfo(c); info != nil {
		return strings.TrimSpace(info.UserID)
	}
	return ""
}

func isReplaySupportedRoute(route string) bool {
	switch strings.TrimSpace(route) {
	case "/v1/chat/completions", "/v1/completions", "/v1/embeddings", "/v1/responses":
		return true
	default:
		return false
	}
}

func buildLogExportCSV(items []platform.RequestLogV2) []byte {
	buf := bytes.NewBuffer(nil)
	writer := csv.NewWriter(buf)
	_ = writer.Write([]string{
		"log_id", "request_id", "trace_id", "provider", "model", "virtual_model", "route", "request_type",
		"status_code", "error_type", "started_at", "duration_ms", "total_tokens", "cost_usd", "cost_irr",
		"cache_status", "retry_count", "fallback_count", "fallback_status", "load_balance_status", "payload_mode",
	})
	for _, item := range items {
		_ = writer.Write([]string{
			item.LogID, item.RequestID, item.TraceID, item.Provider, item.Model, item.VirtualModel, item.Route,
			item.RequestType, strconv.Itoa(item.StatusCode), item.ErrorType, item.StartedAt.Format(time.RFC3339Nano),
			strconv.Itoa(item.DurationMS), strconv.Itoa(item.TotalTokens), strconv.FormatFloat(item.CostUSD, 'f', -1, 64),
			strconv.FormatFloat(item.CostIRR, 'f', -1, 64), item.CacheStatus, strconv.Itoa(item.RetryCount),
			strconv.Itoa(item.FallbackCount), item.FallbackStatus, item.LoadBalanceStatus, item.PayloadMode,
		})
	}
	writer.Flush()
	return buf.Bytes()
}

func buildLogExportJSONL(items []platform.RequestLogV2) []byte {
	buf := bytes.NewBuffer(nil)
	encoder := json.NewEncoder(buf)
	for _, item := range items {
		_ = encoder.Encode(item)
	}
	return buf.Bytes()
}

func logParamsFromFilterMap(tenantID string, filters map[string]interface{}) platform.ListRequestLogsV2Params {
	out := platform.ListRequestLogsV2Params{
		TenantID:  tenantID,
		Limit:     10000,
		SortBy:    "started_at",
		SortOrder: "desc",
	}
	if filters == nil {
		return out
	}
	out.LogID = filterString(filters, "log_id")
	out.RequestID = filterString(filters, "request_id")
	out.TraceID = filterString(filters, "trace_id")
	out.Provider = filterString(filters, "provider")
	out.Model = filterString(filters, "model")
	out.VirtualModel = filterString(filters, "virtual_model")
	out.Route = filterString(filters, "route")
	out.RequestType = filterString(filters, "request_type")
	out.ErrorType = filterString(filters, "error_type")
	out.CacheStatus = filterString(filters, "cache_status")
	out.FallbackStatus = filterString(filters, "fallback_status")
	out.LoadBalanceStatus = filterString(filters, "load_balance_status")
	out.PayloadMode = filterString(filters, "payload_mode")
	out.ActorType = filterString(filters, "actor_type")
	out.ActorID = filterString(filters, "actor_id")
	out.OrgUnitID = filterString(filters, "org_unit_id")
	out.TeamID = filterString(filters, "team_id")
	out.SortBy = defaultString(filterString(filters, "sort_by"), out.SortBy)
	out.SortOrder = defaultString(filterString(filters, "sort_order"), out.SortOrder)
	if limit := filterInt(filters, "limit"); limit != nil {
		out.Limit = *limit
	}
	if offset := filterInt(filters, "offset"); offset != nil {
		out.Offset = *offset
	}
	out.StatusCode = filterInt(filters, "status_code")
	out.DurationMin = filterInt(filters, "duration_min")
	out.DurationMax = filterInt(filters, "duration_max")
	out.TokensMin = filterInt(filters, "tokens_min")
	out.TokensMax = filterInt(filters, "tokens_max")
	out.CostUSDMin = filterFloat(filters, "cost_usd_min")
	out.CostUSDMax = filterFloat(filters, "cost_usd_max")
	out.CostIRRMin = filterFloat(filters, "cost_irr_min")
	out.CostIRRMax = filterFloat(filters, "cost_irr_max")
	out.From = filterTime(filters, "from")
	out.To = filterTime(filters, "to")
	return out
}

func filterString(values map[string]interface{}, key string) string {
	if value, ok := values[key]; ok {
		return strings.TrimSpace(fmt.Sprint(value))
	}
	return ""
}

func filterInt(values map[string]interface{}, key string) *int {
	raw := filterString(values, key)
	if raw == "" {
		return nil
	}
	if value, err := strconv.Atoi(raw); err == nil {
		return &value
	}
	return nil
}

func filterFloat(values map[string]interface{}, key string) *float64 {
	raw := filterString(values, key)
	if raw == "" {
		return nil
	}
	if value, err := strconv.ParseFloat(raw, 64); err == nil {
		return &value
	}
	return nil
}

func filterTime(values map[string]interface{}, key string) *time.Time {
	raw := filterString(values, key)
	if raw == "" {
		return nil
	}
	if value, err := time.Parse(time.RFC3339, raw); err == nil {
		return &value
	}
	return nil
}

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
