// Package observability owns management logs, traces, and metrics HTTP routes.
package observability

import "github.com/gofiber/fiber/v2"

// Handler defines observability operations consumed by this feature.
type Handler interface {
	ListRequestTraces(*fiber.Ctx) error
	GetRequestTrace(*fiber.Ctx) error
	ListRequestTraceSpans(*fiber.Ctx) error
	ListModelMetricsFacets(*fiber.Ctx) error
	QueryModelMetrics(*fiber.Ctx) error
	GetTenantTracePrivacySettings(*fiber.Ctx) error
	UpdateTenantTracePrivacySettings(*fiber.Ctx) error
	GetTenantRequestLoggingSettings(*fiber.Ctx) error
	UpdateTenantRequestLoggingSettings(*fiber.Ctx) error
	ListRequestLogsV2(*fiber.Ctx) error
	GetRequestLogFacetsV2(*fiber.Ctx) error
	GetRequestLogSummaryV2(*fiber.Ctx) error
	GetRequestLogRangeStatsV2(*fiber.Ctx) error
	GetRequestLogAnalyticsV2(*fiber.Ctx) error
	ListSavedLogViewsV2(*fiber.Ctx) error
	CreateSavedLogViewV2(*fiber.Ctx) error
	UpdateSavedLogViewV2(*fiber.Ctx) error
	DeleteSavedLogViewV2(*fiber.Ctx) error
	ListLogExportsV2(*fiber.Ctx) error
	CreateLogExportV2(*fiber.Ctx) error
	GetLogExportV2(*fiber.Ctx) error
	DownloadLogExportV2(*fiber.Ctx) error
	CancelLogExportV2(*fiber.Ctx) error
	CreateRequestLogReplayDraftV2(*fiber.Ctx) error
	GetRequestLogV2(*fiber.Ctx) error
	ListRequestTracesV2(*fiber.Ctx) error
	GetTraceFacetsV2(*fiber.Ctx) error
	GetRequestTraceV2(*fiber.Ctx) error
	ListRequestTraceTimelineV2(*fiber.Ctx) error
	ListRequestTraceAttemptsV2(*fiber.Ctx) error
}

// Routes owns observability route registration.
type Routes struct {
	handler Handler
}

// NewRoutes creates observability routes.
func NewRoutes(handler Handler) *Routes {
	return &Routes{handler: handler}
}

// RegisterTenant registers tenant observability endpoints.
func (r *Routes) RegisterTenant(v1 fiber.Router, v2 fiber.Router) {
	v1.Get("/traces", r.handler.ListRequestTraces)
	v1.Get("/traces/:traceID", r.handler.GetRequestTrace)
	v1.Get("/traces/:traceID/spans", r.handler.ListRequestTraceSpans)
	v1.Get("/metrics/facets", r.handler.ListModelMetricsFacets)
	v1.Post("/metrics/query", r.handler.QueryModelMetrics)
	v1.Get("/settings/trace-privacy", r.handler.GetTenantTracePrivacySettings)
	v1.Put("/settings/trace-privacy", r.handler.UpdateTenantTracePrivacySettings)
	v1.Get("/settings/request-logging", r.handler.GetTenantRequestLoggingSettings)
	v1.Put("/settings/request-logging", r.handler.UpdateTenantRequestLoggingSettings)
	v2.Get("/logs", r.handler.ListRequestLogsV2)
	v2.Get("/logs/facets", r.handler.GetRequestLogFacetsV2)
	v2.Get("/logs/summary", r.handler.GetRequestLogSummaryV2)
	v2.Get("/logs/range-stats", r.handler.GetRequestLogRangeStatsV2)
	v2.Get("/logs/analytics", r.handler.GetRequestLogAnalyticsV2)
	v2.Get("/logs/saved-views", r.handler.ListSavedLogViewsV2)
	v2.Post("/logs/saved-views", r.handler.CreateSavedLogViewV2)
	v2.Patch("/logs/saved-views/:viewID", r.handler.UpdateSavedLogViewV2)
	v2.Delete("/logs/saved-views/:viewID", r.handler.DeleteSavedLogViewV2)
	v2.Get("/logs/exports", r.handler.ListLogExportsV2)
	v2.Post("/logs/exports", r.handler.CreateLogExportV2)
	v2.Get("/logs/exports/:exportID", r.handler.GetLogExportV2)
	v2.Get("/logs/exports/:exportID/download", r.handler.DownloadLogExportV2)
	v2.Post("/logs/exports/:exportID/cancel", r.handler.CancelLogExportV2)
	v2.Post("/logs/:logID/replay-draft", r.handler.CreateRequestLogReplayDraftV2)
	v2.Get("/logs/:logID", r.handler.GetRequestLogV2)
	v2.Get("/traces", r.handler.ListRequestTracesV2)
	v2.Get("/traces/facets", r.handler.GetTraceFacetsV2)
	v2.Get("/traces/:traceID", r.handler.GetRequestTraceV2)
	v2.Get("/traces/:traceID/timeline", r.handler.ListRequestTraceTimelineV2)
	v2.Get("/traces/:traceID/attempts", r.handler.ListRequestTraceAttemptsV2)
}
