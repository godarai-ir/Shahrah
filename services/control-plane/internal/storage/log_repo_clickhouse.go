package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

const logPayloadMaxBytes = 32 * 1024
const requestLogInsertColumnCount = 40

func (s *Store) ensureClickHouseRequestLogsSchema(ctx context.Context) error {
	return s.validateClickHouseColumns(ctx, "request_logs_ch", []string{
		"log_id", "tenant_id", "request_id", "trace_id", "token_hash", "organization_id", "org_unit_id", "team_id",
		"actor_type", "actor_id", "provider", "provider_model", "model", "virtual_model", "route", "request_type",
		"status", "status_code", "error_type", "started_at", "duration_ms", "input_tokens", "output_tokens",
		"total_tokens", "cost_usd", "cost_irr", "cache_status", "cache_mode", "retry_count", "fallback_count",
		"fallback_status", "load_balance_status", "payload_mode", "request_payload_json", "response_payload_json",
		"request_payload_meta_json", "response_payload_meta_json", "request_payload_truncated",
		"response_payload_truncated", "metadata_json",
	})
}

// LogRequest writes the normalized request log entry to ClickHouse.
func (s *Store) LogRequest(ctx context.Context, entry platform.RequestLogEntry) error {
	return s.LogRequests(ctx, []platform.RequestLogEntry{entry})
}

// LogRequests writes normalized request log entries to ClickHouse in one batch.
func (s *Store) LogRequests(ctx context.Context, entries []platform.RequestLogEntry) error {
	if !s.clickHouseLogsReady() {
		return ErrLogStoreUnavailable
	}
	if len(entries) == 0 {
		return nil
	}

	values := make([]string, 0, len(entries))
	args := make([]interface{}, 0, len(entries)*requestLogInsertColumnCount)
	for _, entry := range entries {
		rowArgs, err := s.requestLogInsertArgs(ctx, entry)
		if err != nil {
			return err
		}
		values = append(values, requestLogInsertPlaceholders(requestLogInsertColumnCount))
		args = append(args, rowArgs...)
	}

	_, err := s.clickhouse.ExecContext(ctx, `
		INSERT INTO request_logs_ch (
			log_id, tenant_id, request_id, trace_id, token_hash, organization_id, org_unit_id, team_id,
			actor_type, actor_id, provider, provider_model, model, virtual_model, route, request_type,
			status, status_code, error_type, started_at, duration_ms, input_tokens, output_tokens, total_tokens,
			cost_usd, cost_irr, cache_status, cache_mode, retry_count, fallback_count, fallback_status,
			load_balance_status, payload_mode, request_payload_json, response_payload_json,
			request_payload_meta_json, response_payload_meta_json, request_payload_truncated,
			response_payload_truncated, metadata_json
		) VALUES `+strings.Join(values, ", "), args...)

	return err
}

func (s *Store) requestLogInsertArgs(ctx context.Context, entry platform.RequestLogEntry) ([]interface{}, error) {
	logID := strings.TrimSpace(entry.LogID)
	if logID == "" {
		logID = "log_" + uuid.NewString()
	}

	startedAt := entry.StartedAt.UTC()
	if startedAt.IsZero() {
		startedAt = time.Now().UTC()
	}

	loggingMode := RequestLoggingModeMetadataOnly
	if mode, ok, err := s.ResolveTenantRequestLoggingMode(ctx, entry.TenantID); err != nil {
		if s.logger != nil {
			s.logger.Warn("failed to resolve tenant request logging mode; using metadata-only", zap.Error(err), zap.String("tenant_id", entry.TenantID))
		}
	} else if ok {
		loggingMode = mode
	}

	requestPayload, requestMeta, requestTruncated := prepareRequestLogPayload(loggingMode, entry.RequestBody)
	responsePayload, responseMeta, responseTruncated := prepareRequestLogPayload(loggingMode, entry.ResponseBody)

	if loggingMode == RequestLoggingModeNone {
		requestMeta = map[string]interface{}{}
		responseMeta = map[string]interface{}{}
	}

	if entry.Metadata == nil {
		entry.Metadata = map[string]interface{}{}
	}
	entry.Metadata["payload_mode"] = loggingMode

	return []interface{}{
		logID,
		strings.TrimSpace(entry.TenantID),
		strings.TrimSpace(entry.RequestID),
		strings.TrimSpace(entry.TraceID),
		strings.TrimSpace(entry.TokenHash),
		strings.TrimSpace(entry.OrganizationID),
		strings.TrimSpace(entry.OrgUnitID),
		strings.TrimSpace(entry.TeamID),
		strings.TrimSpace(entry.ActorType),
		strings.TrimSpace(entry.ActorID),
		strings.TrimSpace(entry.Provider),
		strings.TrimSpace(entry.ProviderModel),
		strings.TrimSpace(entry.Model),
		strings.TrimSpace(entry.VirtualModel),
		strings.TrimSpace(entry.Route),
		strings.TrimSpace(entry.RequestType),
		strings.TrimSpace(entry.Status),
		int32(entry.StatusCode),
		strings.TrimSpace(entry.ErrorType),
		startedAt,
		int32(entry.ResponseTimeMS),
		uint32(maxInt(entry.InputTokens, 0)),
		uint32(maxInt(entry.OutputTokens, 0)),
		uint32(maxInt(entry.TotalTokens, 0)),
		entry.CostUSD,
		entry.CostIRR,
		strings.TrimSpace(entry.CacheStatus),
		strings.TrimSpace(entry.CacheMode),
		uint16(maxInt(entry.RetryCount, 0)),
		uint16(maxInt(entry.FallbackCount, 0)),
		deriveFallbackStatus(entry.FallbackStatus, entry.FallbackCount),
		deriveLoadBalanceStatus(entry.LoadBalanceStatus),
		loggingMode,
		string(marshalJSONBestEffort(requestPayload)),
		string(marshalJSONBestEffort(responsePayload)),
		string(marshalJSONBestEffort(requestMeta)),
		string(marshalJSONBestEffort(responseMeta)),
		boolToUint8(requestTruncated || entry.RequestPayloadTruncated),
		boolToUint8(responseTruncated || entry.ResponsePayloadTruncated),
		string(marshalJSONBestEffort(entry.Metadata)),
	}, nil
}

func requestLogInsertPlaceholders(columns int) string {
	parts := make([]string, columns)
	for i := range parts {
		parts[i] = "?"
	}
	return "(" + strings.Join(parts, ", ") + ")"
}

func (s *Store) listRequestLogsV2FromClickHouse(ctx context.Context, params platform.ListRequestLogsV2Params) ([]platform.RequestLogV2, int, error) {
	if params.Limit <= 0 {
		params.Limit = 50
	}
	if params.Limit > 200 {
		params.Limit = 200
	}
	if params.Offset < 0 {
		params.Offset = 0
	}

	whereSQL, args := buildRequestLogsWhere(params)

	query := fmt.Sprintf(`
		SELECT
			log_id, request_id, trace_id, tenant_id, provider, provider_model, model, virtual_model, route,
			request_type, status, status_code, error_type, started_at, duration_ms, input_tokens, output_tokens,
			total_tokens, toFloat64(cost_usd), toFloat64(cost_irr), cache_status, retry_count, fallback_count,
			fallback_status, load_balance_status, payload_mode
		FROM request_logs_ch
		WHERE %s
		ORDER BY %s %s
		LIMIT ? OFFSET ?
	`, whereSQL, normalizeLogSortBy(params.SortBy), normalizeSortOrder(params.SortOrder))

	pageLimit := params.Limit + 1
	args = append(args, pageLimit, params.Offset)
	rows, err := s.clickhouse.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer closeBestEffort(rows)

	items := make([]platform.RequestLogV2, 0, pageLimit)
	for rows.Next() {
		var (
			item          platform.RequestLogV2
			statusCode    int32
			durationMS    int32
			inputTokens   uint32
			outputTokens  uint32
			totalTokens   uint32
			retryCount    uint16
			fallbackCount uint16
		)
		if err := rows.Scan(
			&item.LogID, &item.RequestID, &item.TraceID, &item.TenantID, &item.Provider, &item.ProviderModel,
			&item.Model, &item.VirtualModel, &item.Route, &item.RequestType, &item.Status, &statusCode,
			&item.ErrorType, &item.StartedAt, &durationMS, &inputTokens, &outputTokens, &totalTokens,
			&item.CostUSD, &item.CostIRR, &item.CacheStatus, &retryCount, &fallbackCount,
			&item.FallbackStatus, &item.LoadBalanceStatus, &item.PayloadMode,
		); err != nil {
			return nil, 0, err
		}

		item.StatusCode = int(statusCode)
		item.DurationMS = int(durationMS)
		item.InputTokens = int(inputTokens)
		item.OutputTokens = int(outputTokens)
		item.TotalTokens = int(totalTokens)
		item.RetryCount = int(retryCount)
		item.FallbackCount = int(fallbackCount)
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	hasMore := len(items) > params.Limit
	if hasMore {
		items = items[:params.Limit]
	}

	return items, approximateWindowedTotal(params.Offset, len(items), hasMore), nil
}

func (s *Store) getRequestLogV2FromClickHouse(ctx context.Context, tenantID, logID string, includeContent bool) (*platform.RequestLogDetailV2, error) {
	var (
		item                platform.RequestLogDetailV2
		statusCode          int32
		durationMS          int32
		inputTokens         uint32
		outputTokens        uint32
		totalTokens         uint32
		retryCount          uint16
		fallbackCount       uint16
		requestPayloadJSON  string
		responsePayloadJSON string
		requestMetaJSON     string
		responseMetaJSON    string
		requestTruncated    uint8
		responseTruncated   uint8
		metadataJSON        string
		actorType           string
		actorID             string
		orgUnitID           string
		teamID              string
	)

	err := s.clickhouse.QueryRowContext(ctx, `
		SELECT
			log_id, request_id, trace_id, tenant_id, provider, provider_model, model, virtual_model, route,
			request_type, status, status_code, error_type, started_at, duration_ms, input_tokens, output_tokens,
			total_tokens, toFloat64(cost_usd), toFloat64(cost_irr), cache_status, retry_count, fallback_count,
			fallback_status, load_balance_status, payload_mode, request_payload_json, response_payload_json,
			request_payload_meta_json, response_payload_meta_json, request_payload_truncated, response_payload_truncated,
			metadata_json, actor_type, actor_id, org_unit_id, team_id
		FROM request_logs_ch
		WHERE tenant_id = ? AND log_id = ?
		ORDER BY started_at DESC
		LIMIT 1
	`, tenantID, logID).Scan(
		&item.Log.LogID, &item.Log.RequestID, &item.Log.TraceID, &item.Log.TenantID, &item.Log.Provider, &item.Log.ProviderModel,
		&item.Log.Model, &item.Log.VirtualModel, &item.Log.Route, &item.Log.RequestType, &item.Log.Status,
		&statusCode, &item.Log.ErrorType, &item.Log.StartedAt, &durationMS, &inputTokens, &outputTokens, &totalTokens,
		&item.Log.CostUSD, &item.Log.CostIRR, &item.Log.CacheStatus, &retryCount, &fallbackCount,
		&item.Log.FallbackStatus, &item.Log.LoadBalanceStatus, &item.Log.PayloadMode, &requestPayloadJSON,
		&responsePayloadJSON, &requestMetaJSON, &responseMetaJSON, &requestTruncated, &responseTruncated,
		&metadataJSON, &actorType, &actorID, &orgUnitID, &teamID,
	)
	if err != nil {
		if errorsIsNoRows(err) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}

	item.Log.StatusCode = int(statusCode)
	item.Log.DurationMS = int(durationMS)
	item.Log.InputTokens = int(inputTokens)
	item.Log.OutputTokens = int(outputTokens)
	item.Log.TotalTokens = int(totalTokens)
	item.Log.RetryCount = int(retryCount)
	item.Log.FallbackCount = int(fallbackCount)
	item.Actor = map[string]string{"type": actorType, "id": actorID}
	item.Scope = map[string]string{"org_unit_id": orgUnitID, "team_id": teamID}

	item.ContentVisibility = platform.RequestLogContentVisibilityV2{
		CanViewContent:    includeContent,
		RequestAvailable:  strings.TrimSpace(requestPayloadJSON) != "" && requestPayloadJSON != "null",
		ResponseAvailable: strings.TrimSpace(responsePayloadJSON) != "" && responsePayloadJSON != "null",
	}
	if requestTruncated == 1 {
		item.ContentVisibility.RequestReason = "payload_truncated"
	}
	if responseTruncated == 1 {
		item.ContentVisibility.ResponseReason = "payload_truncated"
	}
	if !includeContent {
		if item.ContentVisibility.RequestAvailable {
			item.ContentVisibility.RequestReason = "content_permission_required"
		}
		if item.ContentVisibility.ResponseAvailable {
			item.ContentVisibility.ResponseReason = "content_permission_required"
		}
	}

	decodeJSONMap(requestMetaJSON, &item.RequestPayloadMeta)
	decodeJSONMap(responseMetaJSON, &item.ResponsePayloadMeta)
	decodeJSONMap(metadataJSON, &item.Metadata)

	if includeContent {
		decodeJSONMap(requestPayloadJSON, &item.RequestPayload)
		decodeJSONMap(responsePayloadJSON, &item.ResponsePayload)
	}

	return &item, nil
}

func (s *Store) getRequestLogFacetsV2FromClickHouse(ctx context.Context, params platform.ListRequestLogsV2Params) (*platform.LogFacetsV2, error) {
	whereSQL, args := buildRequestLogsWhere(params)

	providers, err := queryLogFacetStringRows(ctx, s.clickhouse, buildRequestLogFacetStringQuery(whereSQL, "provider"), args...)
	if err != nil {
		return nil, err
	}
	models, err := queryLogFacetStringRows(ctx, s.clickhouse, buildRequestLogFacetStringQuery(whereSQL, "model"), args...)
	if err != nil {
		return nil, err
	}
	virtualModels, err := queryLogFacetStringRows(ctx, s.clickhouse, buildRequestLogFacetStringQuery(whereSQL, "virtual_model"), args...)
	if err != nil {
		return nil, err
	}
	routes, err := queryLogFacetStringRows(ctx, s.clickhouse, buildRequestLogFacetStringQuery(whereSQL, "route"), args...)
	if err != nil {
		return nil, err
	}
	requestTypes, err := queryLogFacetStringRows(ctx, s.clickhouse, buildRequestLogFacetStringQuery(whereSQL, "request_type"), args...)
	if err != nil {
		return nil, err
	}
	errorTypes, err := queryLogFacetStringRows(ctx, s.clickhouse, buildRequestLogFacetStringQuery(whereSQL, "error_type"), args...)
	if err != nil {
		return nil, err
	}
	cacheStatuses, err := queryLogFacetStringRows(ctx, s.clickhouse, buildRequestLogFacetStringQuery(whereSQL, "cache_status"), args...)
	if err != nil {
		return nil, err
	}
	fallbackStatuses, err := queryLogFacetStringRows(ctx, s.clickhouse, buildRequestLogFacetStringQuery(whereSQL, "fallback_status"), args...)
	if err != nil {
		return nil, err
	}
	loadBalanceStatuses, err := queryLogFacetStringRows(ctx, s.clickhouse, buildRequestLogFacetStringQuery(whereSQL, "load_balance_status"), args...)
	if err != nil {
		return nil, err
	}
	payloadModes, err := queryLogFacetStringRows(ctx, s.clickhouse, buildRequestLogFacetStringQuery(whereSQL, "payload_mode"), args...)
	if err != nil {
		return nil, err
	}
	statusCodes, err := queryLogFacetStatusRows(ctx, s.clickhouse, whereSQL, args...)
	if err != nil {
		return nil, err
	}

	return &platform.LogFacetsV2{
		Providers:           providers,
		Models:              models,
		VirtualModels:       virtualModels,
		Routes:              routes,
		RequestTypes:        requestTypes,
		ErrorTypes:          errorTypes,
		CacheStatuses:       cacheStatuses,
		FallbackStatuses:    fallbackStatuses,
		LoadBalanceStatuses: loadBalanceStatuses,
		PayloadModes:        payloadModes,
		StatusCodes:         statusCodes,
	}, nil
}

func (s *Store) getRequestLogSummaryV2FromClickHouse(ctx context.Context, params platform.ListRequestLogsV2Params) (*platform.LogSummaryV2, error) {
	whereSQL, args := buildRequestLogsWhere(params)

	var (
		requestCount     uint64
		errorCount       uint64
		latencyAvg       float64
		latencyP50       float64
		latencyP95       float64
		latencyP99       float64
		inputTokens      uint64
		outputTokens     uint64
		totalTokens      uint64
		costUSD          float64
		costIRR          float64
		cacheHitCount    uint64
		retryCount       uint64
		fallbackCount    uint64
		loadBalanceCount uint64
	)

	err := s.clickhouse.QueryRowContext(ctx, fmt.Sprintf(`
		SELECT
			count(),
			countIf(status_code >= 400),
			if(count() = 0, 0, avg(duration_ms)),
			if(count() = 0, 0, quantileTDigest(0.50)(duration_ms)),
			if(count() = 0, 0, quantileTDigest(0.95)(duration_ms)),
			if(count() = 0, 0, quantileTDigest(0.99)(duration_ms)),
			sum(input_tokens),
			sum(output_tokens),
			sum(total_tokens),
			toFloat64(sum(cost_usd)),
			toFloat64(sum(cost_irr)),
			countIf(positionCaseInsensitive(cache_status, 'hit') > 0),
			sum(retry_count),
			countIf(fallback_count > 0 OR lower(fallback_status) NOT IN ('', 'inactive', 'disabled')),
			countIf(lower(load_balance_status) NOT IN ('', 'inactive', 'disabled'))
		FROM request_logs_ch
		WHERE %s
	`, whereSQL), args...).Scan(
		&requestCount, &errorCount, &latencyAvg, &latencyP50, &latencyP95, &latencyP99,
		&inputTokens, &outputTokens, &totalTokens, &costUSD, &costIRR, &cacheHitCount,
		&retryCount, &fallbackCount, &loadBalanceCount,
	)
	if err != nil {
		return nil, err
	}

	payloadModes, err := queryLogFacetStringRows(ctx, s.clickhouse, buildRequestLogFacetStringQuery(whereSQL, "payload_mode"), args...)
	if err != nil {
		return nil, err
	}
	statusCodes, err := queryLogFacetStatusRows(ctx, s.clickhouse, whereSQL, args...)
	if err != nil {
		return nil, err
	}
	providers, err := queryLogFacetStringRows(ctx, s.clickhouse, buildRequestLogFacetStringQuery(whereSQL, "provider"), args...)
	if err != nil {
		return nil, err
	}
	models, err := queryLogFacetStringRows(ctx, s.clickhouse, buildRequestLogFacetStringQuery(whereSQL, "model"), args...)
	if err != nil {
		return nil, err
	}
	routes, err := queryLogFacetStringRows(ctx, s.clickhouse, buildRequestLogFacetStringQuery(whereSQL, "route"), args...)
	if err != nil {
		return nil, err
	}

	summary := &platform.LogSummaryV2{
		RequestCount:     int(requestCount),
		ErrorCount:       int(errorCount),
		LatencyAvgMS:     latencyAvg,
		LatencyP50MS:     latencyP50,
		LatencyP95MS:     latencyP95,
		LatencyP99MS:     latencyP99,
		InputTokens:      int(inputTokens),
		OutputTokens:     int(outputTokens),
		TotalTokens:      int(totalTokens),
		CostUSD:          costUSD,
		CostIRR:          costIRR,
		CacheHitCount:    int(cacheHitCount),
		RetryCount:       int(retryCount),
		FallbackCount:    int(fallbackCount),
		LoadBalanceCount: int(loadBalanceCount),
		PayloadModes:     payloadModes,
		StatusCodes:      statusCodes,
		Providers:        providers,
		Models:           models,
		Routes:           routes,
	}
	if requestCount > 0 {
		summary.ErrorRate = float64(errorCount) / float64(requestCount)
		summary.CacheHitRate = float64(cacheHitCount) / float64(requestCount)
	}

	return summary, nil
}

func (s *Store) getRequestLogRangeStatsV2FromClickHouse(ctx context.Context, params platform.ListRequestLogsV2Params) (*platform.LogRangeStatsV2, error) {
	whereSQL, args := buildRequestLogsWhere(params)

	var (
		rowCount    uint64
		durationMin float64
		durationMax float64
		tokensMin   float64
		tokensMax   float64
		costUSDMin  float64
		costUSDMax  float64
		costIRRMin  float64
		costIRRMax  float64
	)

	err := s.clickhouse.QueryRowContext(ctx, fmt.Sprintf(`
		SELECT
			count(),
			if(count() = 0, 0, toFloat64(min(duration_ms))),
			if(count() = 0, 0, toFloat64(max(duration_ms))),
			if(count() = 0, 0, toFloat64(min(total_tokens))),
			if(count() = 0, 0, toFloat64(max(total_tokens))),
			if(count() = 0, 0, min(cost_usd)),
			if(count() = 0, 0, max(cost_usd)),
			if(count() = 0, 0, min(cost_irr)),
			if(count() = 0, 0, max(cost_irr))
		FROM request_logs_ch
		WHERE %s
	`, whereSQL), args...).Scan(
		&rowCount,
		&durationMin,
		&durationMax,
		&tokensMin,
		&tokensMax,
		&costUSDMin,
		&costUSDMax,
		&costIRRMin,
		&costIRRMax,
	)
	if err != nil {
		return nil, err
	}
	if rowCount == 0 {
		return &platform.LogRangeStatsV2{}, nil
	}

	return &platform.LogRangeStatsV2{
		DurationMS:  platform.LogRangeV2{Min: durationMin, Max: durationMax},
		TotalTokens: platform.LogRangeV2{Min: tokensMin, Max: tokensMax},
		CostUSD:     platform.LogRangeV2{Min: costUSDMin, Max: costUSDMax},
		CostIRR:     platform.LogRangeV2{Min: costIRRMin, Max: costIRRMax},
	}, nil
}

func (s *Store) getRequestLogAnalyticsV2FromClickHouse(ctx context.Context, params platform.LogAnalyticsV2Params) ([]platform.LogAnalyticsPointV2, error) {
	whereSQL, args := buildRequestLogsWhere(params.ListRequestLogsV2Params)
	groupColumn := normalizeLogAnalyticsGroupBy(params.GroupBy)
	intervalSeconds := params.IntervalSeconds
	if intervalSeconds <= 0 {
		intervalSeconds = 3600
	}
	if intervalSeconds < 60 {
		intervalSeconds = 60
	}
	if intervalSeconds > 86400 {
		intervalSeconds = 86400
	}

	if strings.EqualFold(strings.TrimSpace(params.Type), "breakdown") {
		if groupColumn == "" {
			groupColumn = "provider"
		}
		return s.queryRequestLogAnalyticsBreakdown(ctx, whereSQL, groupColumn, args...)
	}

	return s.queryRequestLogAnalyticsTimeseries(ctx, whereSQL, groupColumn, intervalSeconds, args...)
}

func (s *Store) queryRequestLogAnalyticsTimeseries(ctx context.Context, whereSQL, groupColumn string, intervalSeconds int, args ...interface{}) ([]platform.LogAnalyticsPointV2, error) {
	groupSelect := "''"
	groupBy := "bucket"
	if groupColumn != "" {
		groupSelect = fmt.Sprintf("ifNull(nullIf(%s, ''), '')", groupColumn)
		groupBy = "bucket, analytics_group"
	}

	query := fmt.Sprintf(`
		SELECT
			toStartOfInterval(started_at, INTERVAL %d SECOND) AS bucket,
			%s AS analytics_group,
			count() AS request_count,
			countIf(status_code >= 400) AS error_count,
			if(count() = 0, 0, quantileTDigest(0.95)(duration_ms)) AS latency_p95_ms,
			sum(total_tokens) AS total_tokens_sum,
			toFloat64(sum(cost_usd)) AS cost_usd,
			countIf(positionCaseInsensitive(cache_status, 'hit') > 0) AS cache_hit_count,
			sum(retry_count) AS retry_count,
			countIf(fallback_count > 0 OR lower(fallback_status) NOT IN ('', 'inactive', 'disabled')) AS fallback_count,
			countIf(lower(load_balance_status) NOT IN ('', 'inactive', 'disabled')) AS load_balance_count
		FROM request_logs_ch
		WHERE %s
		GROUP BY %s
		ORDER BY bucket ASC, request_count DESC
		LIMIT 500
	`, intervalSeconds, groupSelect, whereSQL, groupBy)

	rows, err := s.clickhouse.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer closeBestEffort(rows)

	out := make([]platform.LogAnalyticsPointV2, 0, 64)
	for rows.Next() {
		var (
			point            platform.LogAnalyticsPointV2
			ts               time.Time
			group            string
			requestCount     uint64
			errorCount       uint64
			totalTokens      uint64
			cacheHitCount    uint64
			retryCount       uint64
			fallbackCount    uint64
			loadBalanceCount uint64
		)
		if err := rows.Scan(
			&ts, &group, &requestCount, &errorCount, &point.LatencyP95MS, &totalTokens, &point.CostUSD,
			&cacheHitCount, &retryCount, &fallbackCount, &loadBalanceCount,
		); err != nil {
			return nil, err
		}
		point.Timestamp = &ts
		point.Group = strings.TrimSpace(group)
		point.RequestCount = int(requestCount)
		point.ErrorCount = int(errorCount)
		point.TotalTokens = int(totalTokens)
		point.CacheHitCount = int(cacheHitCount)
		point.RetryCount = int(retryCount)
		point.FallbackCount = int(fallbackCount)
		point.LoadBalanceCount = int(loadBalanceCount)
		if requestCount > 0 {
			point.ErrorRate = float64(errorCount) / float64(requestCount)
		}
		out = append(out, point)
	}

	return out, rows.Err()
}

func (s *Store) queryRequestLogAnalyticsBreakdown(ctx context.Context, whereSQL, groupColumn string, args ...interface{}) ([]platform.LogAnalyticsPointV2, error) {
	query := fmt.Sprintf(`
		SELECT
			ifNull(nullIf(%s, ''), '') AS analytics_group,
			count() AS request_count,
			countIf(status_code >= 400) AS error_count,
			if(count() = 0, 0, quantileTDigest(0.95)(duration_ms)) AS latency_p95_ms,
			sum(total_tokens) AS total_tokens_sum,
			toFloat64(sum(cost_usd)) AS cost_usd,
			countIf(positionCaseInsensitive(cache_status, 'hit') > 0) AS cache_hit_count,
			sum(retry_count) AS retry_count,
			countIf(fallback_count > 0 OR lower(fallback_status) NOT IN ('', 'inactive', 'disabled')) AS fallback_count,
			countIf(lower(load_balance_status) NOT IN ('', 'inactive', 'disabled')) AS load_balance_count
		FROM request_logs_ch
		WHERE %s
		GROUP BY analytics_group
		ORDER BY request_count DESC
		LIMIT 50
	`, groupColumn, whereSQL)

	rows, err := s.clickhouse.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer closeBestEffort(rows)

	out := make([]platform.LogAnalyticsPointV2, 0, 16)
	for rows.Next() {
		var (
			point            platform.LogAnalyticsPointV2
			group            string
			requestCount     uint64
			errorCount       uint64
			totalTokens      uint64
			cacheHitCount    uint64
			retryCount       uint64
			fallbackCount    uint64
			loadBalanceCount uint64
		)
		if err := rows.Scan(
			&group, &requestCount, &errorCount, &point.LatencyP95MS, &totalTokens, &point.CostUSD,
			&cacheHitCount, &retryCount, &fallbackCount, &loadBalanceCount,
		); err != nil {
			return nil, err
		}
		point.Group = strings.TrimSpace(group)
		point.RequestCount = int(requestCount)
		point.ErrorCount = int(errorCount)
		point.TotalTokens = int(totalTokens)
		point.CacheHitCount = int(cacheHitCount)
		point.RetryCount = int(retryCount)
		point.FallbackCount = int(fallbackCount)
		point.LoadBalanceCount = int(loadBalanceCount)
		if requestCount > 0 {
			point.ErrorRate = float64(errorCount) / float64(requestCount)
		}
		out = append(out, point)
	}

	return out, rows.Err()
}

func prepareRequestLogPayload(mode string, payload map[string]interface{}) (map[string]interface{}, map[string]interface{}, bool) {
	if len(payload) == 0 {
		return nil, map[string]interface{}{"present": false}, false
	}

	raw := marshalJSONBestEffort(payload)
	meta := map[string]interface{}{
		"present":   true,
		"bytes":     len(raw),
		"truncated": len(raw) > logPayloadMaxBytes,
	}

	switch mode {
	case RequestLoggingModeNone, RequestLoggingModeMetadataOnly:
		return nil, meta, false
	case RequestLoggingModeRedactedContent:
		redacted := redactPayloadMap(payload)
		return capPayloadMap(redacted, logPayloadMaxBytes), meta, len(raw) > logPayloadMaxBytes
	default:
		return capPayloadMap(payload, logPayloadMaxBytes), meta, len(raw) > logPayloadMaxBytes
	}
}

func capPayloadMap(payload map[string]interface{}, maxBytes int) map[string]interface{} {
	if len(payload) == 0 {
		return nil
	}

	raw := marshalJSONBestEffort(payload)
	if len(raw) <= maxBytes {
		return payload
	}

	out := map[string]interface{}{}
	if err := json.Unmarshal(raw[:maxBytes], &out); err == nil {
		return out
	}

	return map[string]interface{}{"truncated": true, "preview": string(raw[:maxBytes])}
}

func redactPayloadMap(payload map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(payload))
	for k, v := range payload {
		switch value := v.(type) {
		case string:
			out[k] = "[redacted]"
		case map[string]interface{}:
			out[k] = redactPayloadMap(value)
		case []interface{}:
			items := make([]interface{}, 0, len(value))
			for _, item := range value {
				if child, ok := item.(map[string]interface{}); ok {
					items = append(items, redactPayloadMap(child))
					continue
				}
				if _, ok := item.(string); ok {
					items = append(items, "[redacted]")
					continue
				}
				items = append(items, item)
			}
			out[k] = items
		default:
			out[k] = value
		}
	}
	return out
}

func buildRequestLogsWhere(params platform.ListRequestLogsV2Params) (string, []interface{}) {
	where := []string{"tenant_id = ?"}
	args := []interface{}{params.TenantID}

	appendFilter := func(value, clause string) {
		if strings.TrimSpace(value) == "" {
			return
		}
		where = append(where, clause)
		args = append(args, strings.TrimSpace(value))
	}
	appendMultiFilter := func(value, field string) {
		values := splitLogFilterValues(value)
		if len(values) == 0 {
			return
		}
		if len(values) == 1 {
			where = append(where, field+" = ?")
			args = append(args, values[0])
			return
		}
		placeholders := make([]string, 0, len(values))
		for _, item := range values {
			placeholders = append(placeholders, "?")
			args = append(args, item)
		}
		where = append(where, fmt.Sprintf("%s IN (%s)", field, strings.Join(placeholders, ", ")))
	}

	appendFilter(params.LogID, "log_id = ?")
	appendFilter(params.RequestID, "request_id = ?")
	appendFilter(params.TraceID, "trace_id = ?")
	appendFilter(params.Provider, "provider = ?")
	appendFilter(params.Model, "model = ?")
	appendFilter(params.VirtualModel, "virtual_model = ?")
	appendFilter(params.Route, "route = ?")
	appendFilter(params.RequestType, "request_type = ?")
	appendFilter(params.ErrorType, "error_type = ?")
	appendFilter(params.CacheStatus, "cache_status = ?")
	appendFilter(params.FallbackStatus, "fallback_status = ?")
	appendFilter(params.LoadBalanceStatus, "load_balance_status = ?")
	appendFilter(params.PayloadMode, "payload_mode = ?")
	appendMultiFilter(params.ActorType, "actor_type")
	appendMultiFilter(params.ActorID, "actor_id")
	appendMultiFilter(params.OrgUnitID, "org_unit_id")
	appendMultiFilter(params.TeamID, "team_id")

	if params.StatusCode != nil {
		where = append(where, "status_code = ?")
		args = append(args, int32(*params.StatusCode))
	}
	if params.DurationMin != nil {
		where = append(where, "duration_ms >= ?")
		args = append(args, int32(*params.DurationMin))
	}
	if params.DurationMax != nil {
		where = append(where, "duration_ms <= ?")
		args = append(args, int32(*params.DurationMax))
	}
	if params.TokensMin != nil {
		where = append(where, "total_tokens >= ?")
		args = append(args, uint32(*params.TokensMin))
	}
	if params.TokensMax != nil {
		where = append(where, "total_tokens <= ?")
		args = append(args, uint32(*params.TokensMax))
	}
	if params.CostUSDMin != nil {
		where = append(where, "cost_usd >= ?")
		args = append(args, *params.CostUSDMin)
	}
	if params.CostUSDMax != nil {
		where = append(where, "cost_usd <= ?")
		args = append(args, *params.CostUSDMax)
	}
	if params.CostIRRMin != nil {
		where = append(where, "cost_irr >= ?")
		args = append(args, *params.CostIRRMin)
	}
	if params.CostIRRMax != nil {
		where = append(where, "cost_irr <= ?")
		args = append(args, *params.CostIRRMax)
	}
	if params.From != nil {
		where = append(where, "started_at >= ?")
		args = append(args, params.From.UTC())
	}
	if params.To != nil {
		where = append(where, "started_at <= ?")
		args = append(args, params.To.UTC())
	}

	return strings.Join(where, " AND "), args
}

func approximateWindowedTotal(offset, itemCount int, hasMore bool) int {
	total := offset + itemCount
	if hasMore {
		total++
	}

	return total
}

func splitLogFilterValues(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		normalized := strings.TrimSpace(part)
		if normalized == "" {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	return out
}

func normalizeLogSortBy(sortBy string) string {
	switch strings.TrimSpace(strings.ToLower(sortBy)) {
	case "duration_ms":
		return "duration_ms"
	case "status_code":
		return "status_code"
	case "cost_usd":
		return "cost_usd"
	case "total_tokens":
		return "total_tokens"
	case "cost_irr":
		return "cost_irr"
	case "retry_count":
		return "retry_count"
	default:
		return "started_at"
	}
}

func normalizeLogAnalyticsGroupBy(groupBy string) string {
	switch strings.TrimSpace(strings.ToLower(groupBy)) {
	case "provider":
		return "provider"
	case "model":
		return "model"
	case "virtual_model":
		return "virtual_model"
	case "route":
		return "route"
	case "request_type":
		return "request_type"
	case "status_code":
		return "toString(status_code)"
	case "error_type":
		return "error_type"
	case "cache_status":
		return "cache_status"
	case "fallback_status":
		return "fallback_status"
	case "load_balance_status":
		return "load_balance_status"
	case "payload_mode":
		return "payload_mode"
	default:
		return ""
	}
}

func buildRequestLogFacetStringQuery(whereSQL, field string) string {
	return fmt.Sprintf(`
		SELECT ifNull(nullIf(%s, ''), '') AS value, count() AS count
		FROM request_logs_ch
		WHERE %s
		GROUP BY value
		ORDER BY count DESC
		LIMIT 20
	`, field, whereSQL)
}

func queryLogFacetStringRows(ctx context.Context, db *sql.DB, query string, args ...interface{}) ([]platform.LogFacetBucketV2, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer closeBestEffort(rows)

	out := make([]platform.LogFacetBucketV2, 0, 8)
	for rows.Next() {
		var value string
		var count uint64
		if err := rows.Scan(&value, &count); err != nil {
			return nil, err
		}
		out = append(out, platform.LogFacetBucketV2{Value: strings.TrimSpace(value), Count: int(count)})
	}

	return out, rows.Err()
}

func queryLogFacetStatusRows(ctx context.Context, db *sql.DB, whereSQL string, args ...interface{}) ([]platform.LogStatusFacetBucketV2, error) {
	rows, err := db.QueryContext(ctx, fmt.Sprintf(`
		SELECT status_code, count() AS count
		FROM request_logs_ch
		WHERE %s
		GROUP BY status_code
		ORDER BY count DESC
		LIMIT 20
	`, whereSQL), args...)
	if err != nil {
		return nil, err
	}
	defer closeBestEffort(rows)

	out := make([]platform.LogStatusFacetBucketV2, 0, 8)
	for rows.Next() {
		var statusCode int32
		var count uint64
		if err := rows.Scan(&statusCode, &count); err != nil {
			return nil, err
		}
		out = append(out, platform.LogStatusFacetBucketV2{StatusCode: int(statusCode), Count: int(count)})
	}

	return out, rows.Err()
}

func decodeJSONMap(raw string, target *map[string]interface{}) {
	if strings.TrimSpace(raw) == "" || raw == "null" {
		return
	}
	out := map[string]interface{}{}
	if err := json.Unmarshal([]byte(raw), &out); err == nil {
		*target = out
	}
}

func deriveFallbackStatus(current string, count int) string {
	current = strings.TrimSpace(current)
	if current != "" {
		return current
	}
	if count > 0 {
		return "active"
	}
	return "inactive"
}

func deriveLoadBalanceStatus(current string) string {
	current = strings.TrimSpace(current)
	if current == "" {
		return "inactive"
	}
	return current
}

func maxInt(value int, floor int) int {
	if value < floor {
		return floor
	}
	return value
}
