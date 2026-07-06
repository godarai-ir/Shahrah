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
	_, err := s.clickhouse.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS request_logs_ch (
		log_id String,
		tenant_id String,
		request_id String,
		trace_id String,
		token_hash String,
		organization_id String,
		org_unit_id String,
		team_id String,
		actor_type LowCardinality(String),
		actor_id String,
		provider String,
		provider_model String,
		model String,
		virtual_model String,
		route String,
		request_type LowCardinality(String),
		status LowCardinality(String),
		status_code Int32,
		error_type LowCardinality(String),
		started_at DateTime64(3, 'UTC'),
		duration_ms Int32,
		input_tokens UInt32,
		output_tokens UInt32,
		total_tokens UInt32,
		cost_usd Decimal(18, 6),
		cost_irr Decimal(18, 2),
		cache_status LowCardinality(String),
		cache_mode LowCardinality(String),
		retry_count UInt16,
		fallback_count UInt16,
		fallback_status LowCardinality(String),
		load_balance_status LowCardinality(String),
		payload_mode LowCardinality(String),
		request_payload_json String,
		response_payload_json String,
		request_payload_meta_json String,
		response_payload_meta_json String,
		request_payload_truncated UInt8,
		response_payload_truncated UInt8,
		metadata_json String
	) ENGINE = MergeTree
	PARTITION BY toDate(started_at)
	ORDER BY (tenant_id, started_at, log_id)`)

	return err
}

// LogRequest writes the normalized request log entry to ClickHouse.
func (s *Store) LogRequest(ctx context.Context, entry platform.RequestLogEntry) error {
	return s.LogRequests(ctx, []platform.RequestLogEntry{entry})
}

// LogRequests writes normalized request log entries to ClickHouse in one batch.
func (s *Store) LogRequests(ctx context.Context, entries []platform.RequestLogEntry) error {
	if !s.clickhouseReady || s.clickhouse == nil {
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

	var total uint64
	if err := s.clickhouse.QueryRowContext(ctx, fmt.Sprintf(`SELECT count() FROM request_logs_ch WHERE %s`, whereSQL), args...).Scan(&total); err != nil {
		return nil, 0, err
	}

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

	args = append(args, params.Limit, params.Offset)
	rows, err := s.clickhouse.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer closeBestEffort(rows)

	items := make([]platform.RequestLogV2, 0, params.Limit)
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

	return items, int(total), rows.Err()
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
	statusCodes, err := queryLogFacetStatusRows(ctx, s.clickhouse, whereSQL, args...)
	if err != nil {
		return nil, err
	}

	return &platform.LogFacetsV2{
		Providers:        providers,
		Models:           models,
		VirtualModels:    virtualModels,
		Routes:           routes,
		RequestTypes:     requestTypes,
		ErrorTypes:       errorTypes,
		CacheStatuses:    cacheStatuses,
		FallbackStatuses: fallbackStatuses,
		StatusCodes:      statusCodes,
	}, nil
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

	if params.StatusCode != nil {
		where = append(where, "status_code = ?")
		args = append(args, int32(*params.StatusCode))
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
	default:
		return "started_at"
	}
}

func buildRequestLogFacetStringQuery(whereSQL, field string) string {
	return fmt.Sprintf(`
		SELECT nullIf(%s, '') AS value, count() AS count
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
		var value sql.NullString
		var count uint64
		if err := rows.Scan(&value, &count); err != nil {
			return nil, err
		}
		out = append(out, platform.LogFacetBucketV2{Value: strings.TrimSpace(value.String), Count: int(count)})
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
