package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func (s *Store) ensureClickHouseMetricsSchema(ctx context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS gateway_model_events
		(
		  ts DateTime64(3, 'UTC'),
		  tenant_id String,
		  request_id String,
		  model_name LowCardinality(String),
		  virtual_model_name LowCardinality(Nullable(String)),
		  provider_name LowCardinality(String),
		  provider_model_name LowCardinality(String),
		  request_type LowCardinality(String),
		  user_id Nullable(String),
		  user_email Nullable(String),
		  virtual_account_id Nullable(String),
		  virtual_account_name Nullable(String),
		  team_ids Array(String),
		  team_names Array(String),
		  status_code UInt16,
		  error_type LowCardinality(String),
		  latency_ms UInt32,
		  ttft_ms Nullable(UInt32),
		  itl_ms Nullable(UInt32),
		  tpot_ms Nullable(UInt32),
		  input_tokens UInt32,
		  output_tokens UInt32,
		  cost_usd Decimal(18, 8),
		  metadata_json String
		)
		ENGINE = MergeTree
		PARTITION BY toDate(ts)
		ORDER BY (tenant_id, ts, model_name, ifNull(virtual_model_name, ''))
		TTL toDate(ts) + INTERVAL 90 DAY
		SETTINGS index_granularity = 8192, allow_nullable_key = 1`,
		`CREATE TABLE IF NOT EXISTS gateway_model_metrics_1m
		(
		  bucket_ts DateTime('UTC'),
		  tenant_id String,
		  view_model_name LowCardinality(String),
		  view_virtual_model_name LowCardinality(Nullable(String)),
		  view_user_email Nullable(String),
		  view_virtual_account_name Nullable(String),
		  view_team_name Nullable(String),
		  view_error_type LowCardinality(String),
		  request_count AggregateFunction(count),
		  failed_count AggregateFunction(countIf, UInt8),
		  input_tokens_sum AggregateFunction(sum, UInt64),
		  output_tokens_sum AggregateFunction(sum, UInt64),
		  cost_usd_sum AggregateFunction(sum, Decimal(18, 8)),
		  latency_p50 AggregateFunction(quantileTDigest(0.50), UInt32),
		  latency_p75 AggregateFunction(quantileTDigest(0.75), UInt32),
		  latency_p90 AggregateFunction(quantileTDigest(0.90), UInt32),
		  latency_p99 AggregateFunction(quantileTDigest(0.99), UInt32),
		  ttft_p50 AggregateFunction(quantileTDigest(0.50), UInt32),
		  ttft_p75 AggregateFunction(quantileTDigest(0.75), UInt32),
		  ttft_p90 AggregateFunction(quantileTDigest(0.90), UInt32),
		  ttft_p99 AggregateFunction(quantileTDigest(0.99), UInt32),
		  itl_p50 AggregateFunction(quantileTDigest(0.50), UInt32),
		  itl_p75 AggregateFunction(quantileTDigest(0.75), UInt32),
		  itl_p90 AggregateFunction(quantileTDigest(0.90), UInt32),
		  itl_p99 AggregateFunction(quantileTDigest(0.99), UInt32),
		  tpot_p50 AggregateFunction(quantileTDigest(0.50), UInt32),
		  tpot_p75 AggregateFunction(quantileTDigest(0.75), UInt32),
		  tpot_p90 AggregateFunction(quantileTDigest(0.90), UInt32),
		  tpot_p99 AggregateFunction(quantileTDigest(0.99), UInt32)
		)
		ENGINE = AggregatingMergeTree
		PARTITION BY toDate(bucket_ts)
		ORDER BY (tenant_id, bucket_ts, view_model_name, ifNull(view_virtual_model_name, ''))
		TTL toDate(bucket_ts) + INTERVAL 540 DAY
		SETTINGS index_granularity = 8192, allow_nullable_key = 1`,
		`CREATE MATERIALIZED VIEW IF NOT EXISTS mv_gateway_model_metrics_1m
		TO gateway_model_metrics_1m
		AS
		SELECT
		  toStartOfMinute(ts) AS bucket_ts,
		  tenant_id,
		  model_name AS view_model_name,
		  virtual_model_name AS view_virtual_model_name,
		  user_email AS view_user_email,
		  virtual_account_name AS view_virtual_account_name,
		  CAST(NULL, 'Nullable(String)') AS view_team_name,
		  error_type AS view_error_type,
		  countState() AS request_count,
		  countIfState(status_code >= 400) AS failed_count,
		  sumState(toUInt64(input_tokens)) AS input_tokens_sum,
		  sumState(toUInt64(output_tokens)) AS output_tokens_sum,
		  sumState(cost_usd) AS cost_usd_sum,
		  quantileTDigestState(0.50)(latency_ms) AS latency_p50,
		  quantileTDigestState(0.75)(latency_ms) AS latency_p75,
		  quantileTDigestState(0.90)(latency_ms) AS latency_p90,
		  quantileTDigestState(0.99)(latency_ms) AS latency_p99,
		  quantileTDigestStateIf(0.50)(ttft_ms, ttft_ms IS NOT NULL) AS ttft_p50,
		  quantileTDigestStateIf(0.75)(ttft_ms, ttft_ms IS NOT NULL) AS ttft_p75,
		  quantileTDigestStateIf(0.90)(ttft_ms, ttft_ms IS NOT NULL) AS ttft_p90,
		  quantileTDigestStateIf(0.99)(ttft_ms, ttft_ms IS NOT NULL) AS ttft_p99,
		  quantileTDigestStateIf(0.50)(itl_ms, itl_ms IS NOT NULL) AS itl_p50,
		  quantileTDigestStateIf(0.75)(itl_ms, itl_ms IS NOT NULL) AS itl_p75,
		  quantileTDigestStateIf(0.90)(itl_ms, itl_ms IS NOT NULL) AS itl_p90,
		  quantileTDigestStateIf(0.99)(itl_ms, itl_ms IS NOT NULL) AS itl_p99,
		  quantileTDigestStateIf(0.50)(tpot_ms, tpot_ms IS NOT NULL) AS tpot_p50,
		  quantileTDigestStateIf(0.75)(tpot_ms, tpot_ms IS NOT NULL) AS tpot_p75,
		  quantileTDigestStateIf(0.90)(tpot_ms, tpot_ms IS NOT NULL) AS tpot_p90,
		  quantileTDigestStateIf(0.99)(tpot_ms, tpot_ms IS NOT NULL) AS tpot_p99
		FROM gateway_model_events
		GROUP BY
		  bucket_ts,
		  tenant_id,
		  view_model_name,
		  view_virtual_model_name,
		  view_user_email,
		  view_virtual_account_name,
		  view_team_name,
		  view_error_type`,
	}
	for _, stmt := range stmts {
		if _, err := s.clickhouse.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) insertModelMetricEvent(ctx context.Context, ev platform.ModelMetricEvent) error {
	const insertSQL = `
		INSERT INTO gateway_model_events
		(
			ts, tenant_id, request_id, model_name, virtual_model_name, provider_name, provider_model_name,
			request_type, user_id, user_email, virtual_account_id, virtual_account_name, team_ids, team_names,
			status_code, error_type, latency_ms, ttft_ms, itl_ms, tpot_ms, input_tokens, output_tokens, cost_usd, metadata_json
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	metaRaw := "{}"
	if ev.Metadata != nil {
		if b, err := json.Marshal(ev.Metadata); err == nil {
			metaRaw = string(b)
		}
	}
	_, err := s.clickhouse.ExecContext(
		ctx,
		insertSQL,
		ev.TS.UTC(),
		ev.TenantID,
		ev.RequestID,
		ev.ModelName,
		nullIfEmpty(ev.VirtualModelName),
		nullIfEmpty(ev.ProviderName),
		nullIfEmpty(ev.ProviderModelName),
		ev.RequestType,
		nullIfEmpty(ev.UserID),
		nullIfEmpty(ev.UserEmail),
		nullIfEmpty(ev.VirtualAccountID),
		nullIfEmpty(ev.VirtualAccountName),
		ev.TeamIDs,
		ev.TeamNames,
		ev.StatusCode,
		ev.ErrorType,
		ev.LatencyMS,
		intPtrToNull(ev.TTFTMS),
		intPtrToNull(ev.ITLMS),
		intPtrToNull(ev.TPOTMS),
		ev.InputTokens,
		ev.OutputTokens,
		ev.CostUSD,
		metaRaw,
	)
	return err
}

func (s *Store) queryModelMetrics(ctx context.Context, tenantID string, req platform.MetricsQueryRequest) ([]map[string]interface{}, error) {
	if req.Type == "timeseries" {
		return s.queryModelMetricsTimeseries(ctx, tenantID, req)
	}
	return s.queryModelMetricsDistribution(ctx, tenantID, req)
}

func (s *Store) queryModelMetricsDistribution(ctx context.Context, tenantID string, req platform.MetricsQueryRequest) ([]map[string]interface{}, error) {
	selects := make([]string, 0, len(req.GroupBy)+len(req.Aggregations))
	groupBy := make([]string, 0, len(req.GroupBy))
	for _, g := range req.GroupBy {
		col, err := mapMetricsGroupBy(g)
		if err != nil {
			return nil, err
		}
		selects = append(selects, fmt.Sprintf("%s AS %s", col, sanitizeAlias(g)))
		groupBy = append(groupBy, col)
	}
	aggs, err := buildAggregations(req.Aggregations)
	if err != nil {
		return nil, err
	}
	selects = append(selects, aggs...)
	if len(selects) == 0 {
		selects = append(selects, "count() AS countAll")
	}

	where, args, err := buildMetricsWhere(tenantID, req)
	if err != nil {
		return nil, err
	}
	query := fmt.Sprintf("SELECT %s FROM gateway_model_events WHERE %s", strings.Join(selects, ", "), where)
	if len(groupBy) > 0 {
		query += " GROUP BY " + strings.Join(groupBy, ", ")
	}
	query += " ORDER BY count() DESC LIMIT 1000"

	return queryRowsAsMaps(ctx, s.clickhouse, query, args...)
}

func (s *Store) queryModelMetricsTimeseries(ctx context.Context, tenantID string, req platform.MetricsQueryRequest) ([]map[string]interface{}, error) {
	interval := req.IntervalInSeconds
	if interval <= 0 {
		interval = 300
	}
	selects := []string{fmt.Sprintf("toStartOfInterval(ts, INTERVAL %d second) AS timestamp", interval)}
	groupBy := []string{"timestamp"}
	for _, g := range req.GroupBy {
		col, err := mapMetricsGroupBy(g)
		if err != nil {
			return nil, err
		}
		selects = append(selects, fmt.Sprintf("%s AS %s", col, sanitizeAlias(g)))
		groupBy = append(groupBy, col)
	}
	aggs, err := buildAggregations(req.Aggregations)
	if err != nil {
		return nil, err
	}
	selects = append(selects, aggs...)
	where, args, err := buildMetricsWhere(tenantID, req)
	if err != nil {
		return nil, err
	}
	query := fmt.Sprintf(
		"SELECT %s FROM gateway_model_events WHERE %s GROUP BY %s ORDER BY timestamp ASC LIMIT 5000",
		strings.Join(selects, ", "),
		where,
		strings.Join(groupBy, ", "),
	)
	return queryRowsAsMaps(ctx, s.clickhouse, query, args...)
}

func buildAggregations(aggs []platform.MetricsAggregation) ([]string, error) {
	out := make([]string, 0, len(aggs))
	if len(aggs) == 0 {
		return []string{"count() AS countAll"}, nil
	}
	for _, agg := range aggs {
		col, err := mapAggregationColumn(agg.Column)
		if err != nil {
			return nil, err
		}
		switch agg.Type {
		case "count":
			out = append(out, fmt.Sprintf("count(%s) AS count%s", col, toPascal(agg.Column)))
		case "sum":
			out = append(out, fmt.Sprintf("sum(%s) AS sum%s", col, toPascal(agg.Column)))
		case "p50", "p75", "p90", "p99":
			q := map[string]string{"p50": "0.50", "p75": "0.75", "p90": "0.90", "p99": "0.99"}[agg.Type]
			out = append(out, fmt.Sprintf("quantileTDigest(%s)(%s) AS %s%s", q, col, agg.Type, toPascal(agg.Column)))
		default:
			return nil, fmt.Errorf("unsupported aggregation type: %s", agg.Type)
		}
	}
	return out, nil
}

func buildMetricsWhere(tenantID string, req platform.MetricsQueryRequest) (string, []interface{}, error) {
	where := []string{"tenant_id = ?", "ts >= ?", "ts <= ?"}
	args := []interface{}{tenantID, req.StartTS.UTC(), req.EndTS.UTC()}
	for _, f := range req.Filters {
		col := ""
		if f.MetadataKey != "" {
			col = fmt.Sprintf("JSONExtractString(metadata_json, '%s')", sanitizeMetadataKey(f.MetadataKey))
		} else {
			mapped, err := mapFilterField(f.FieldName)
			if err != nil {
				return "", nil, err
			}
			col = mapped
		}
		op := strings.ToUpper(strings.TrimSpace(f.Operator))
		switch op {
		case "EQUAL":
			where = append(where, col+" = ?")
			args = append(args, f.Value)
		case "IN":
			values, ok := interfaceSlice(f.Value)
			if !ok || len(values) == 0 {
				return "", nil, fmt.Errorf("IN filter requires non-empty array value")
			}
			where = append(where, col+" IN ("+strings.TrimRight(strings.Repeat("?,", len(values)), ",")+")")
			args = append(args, values...)
		case "NOT_IN":
			values, ok := interfaceSlice(f.Value)
			if !ok || len(values) == 0 {
				return "", nil, fmt.Errorf("NOT_IN filter requires non-empty array value")
			}
			where = append(where, col+" NOT IN ("+strings.TrimRight(strings.Repeat("?,", len(values)), ",")+")")
			args = append(args, values...)
		case "STRING_CONTAINS":
			where = append(where, col+" LIKE ?")
			args = append(args, "%"+fmt.Sprintf("%v", f.Value)+"%")
		case "STRING_STARTS_WITH":
			where = append(where, col+" LIKE ?")
			args = append(args, fmt.Sprintf("%v", f.Value)+"%")
		case "STRING_ENDS_WITH":
			where = append(where, col+" LIKE ?")
			args = append(args, "%"+fmt.Sprintf("%v", f.Value))
		case "GREATER_THAN":
			where = append(where, col+" > ?")
			args = append(args, f.Value)
		case "LESS_THAN":
			where = append(where, col+" < ?")
			args = append(args, f.Value)
		case "GREATER_THAN_EQUAL":
			where = append(where, col+" >= ?")
			args = append(args, f.Value)
		case "LESS_THAN_EQUAL":
			where = append(where, col+" <= ?")
			args = append(args, f.Value)
		case "BETWEEN":
			values, ok := interfaceSlice(f.Value)
			if !ok || len(values) != 2 {
				return "", nil, fmt.Errorf("BETWEEN filter requires two values")
			}
			where = append(where, col+" BETWEEN ? AND ?")
			args = append(args, values[0], values[1])
		case "ARRAY_HAS_ANY":
			values, ok := stringSlice(f.Value)
			if !ok || len(values) == 0 {
				return "", nil, fmt.Errorf("ARRAY_HAS_ANY requires non-empty string array")
			}
			where = append(where, fmt.Sprintf("hasAny(%s, ?)", col))
			args = append(args, values)
		case "ARRAY_HAS_NONE":
			values, ok := stringSlice(f.Value)
			if !ok || len(values) == 0 {
				return "", nil, fmt.Errorf("ARRAY_HAS_NONE requires non-empty string array")
			}
			where = append(where, fmt.Sprintf("NOT hasAny(%s, ?)", col))
			args = append(args, values)
		case "IS_NULL":
			isNull := false
			switch v := f.Value.(type) {
			case bool:
				isNull = v
			case string:
				isNull = strings.EqualFold(v, "true")
			}
			if isNull {
				where = append(where, "("+col+" IS NULL OR "+col+" = '')")
			} else {
				where = append(where, "("+col+" IS NOT NULL AND "+col+" != '')")
			}
		default:
			return "", nil, fmt.Errorf("unsupported filter operator: %s", op)
		}
	}
	return strings.Join(where, " AND "), args, nil
}

func mapAggregationColumn(col string) (string, error) {
	switch col {
	case "modelName":
		return "model_name", nil
	case "costInUSD":
		return "cost_usd", nil
	case "inputTokens":
		return "input_tokens", nil
	case "outputTokens":
		return "output_tokens", nil
	case "latencyMs":
		return "latency_ms", nil
	case "timeToFirstTokenMs":
		return "ttft_ms", nil
	case "interTokenLatencyMs":
		return "itl_ms", nil
	case "timePerOutputTokenLatencyMs":
		return "tpot_ms", nil
	default:
		return "", fmt.Errorf("unsupported aggregation column: %s", col)
	}
}

func mapMetricsGroupBy(g string) (string, error) {
	switch {
	case strings.HasPrefix(g, "metadata."):
		key := strings.TrimPrefix(g, "metadata.")
		return fmt.Sprintf("JSONExtractString(metadata_json, '%s')", sanitizeMetadataKey(key)), nil
	case g == "modelName":
		return "model_name", nil
	case g == "virtualModelName":
		return "virtual_model_name", nil
	case g == "userEmail":
		return "user_email", nil
	case g == "virtualAccount":
		return "virtual_account_name", nil
	case g == "team":
		return "arrayJoin(team_names)", nil
	case g == "errorCode":
		return "toString(status_code)", nil
	case g == "requestType":
		return "request_type", nil
	case g == "providerModelName":
		return "provider_model_name", nil
	default:
		return "", fmt.Errorf("unsupported groupBy field: %s", g)
	}
}

func mapFilterField(f string) (string, error) {
	switch f {
	case "modelName":
		return "model_name", nil
	case "requestType":
		return "request_type", nil
	case "userEmail":
		return "user_email", nil
	case "virtualAccount":
		return "virtual_account_name", nil
	case "team":
		return "team_names", nil
	case "latencyMs":
		return "latency_ms", nil
	case "virtualModelName":
		return "virtual_model_name", nil
	case "inputTokens":
		return "input_tokens", nil
	case "outputTokens":
		return "output_tokens", nil
	case "statusCode":
		return "status_code", nil
	case "errorType":
		return "error_type", nil
	default:
		return "", fmt.Errorf("unsupported filter field: %s", f)
	}
}

func queryRowsAsMaps(ctx context.Context, db *sql.DB, query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, 0, 128)
	for rows.Next() {
		values := make([]interface{}, len(cols))
		scanTargets := make([]interface{}, len(cols))
		for i := range values {
			scanTargets[i] = &values[i]
		}
		if err := rows.Scan(scanTargets...); err != nil {
			return nil, err
		}
		item := make(map[string]interface{}, len(cols))
		for i, col := range cols {
			switch v := values[i].(type) {
			case []byte:
				item[col] = string(v)
			default:
				item[col] = v
			}
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func nullIfEmpty(v string) interface{} {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	return v
}

func intPtrToNull(v *int) interface{} {
	if v == nil {
		return nil
	}
	return *v
}

func toPascal(input string) string {
	parts := strings.FieldsFunc(input, func(r rune) bool {
		return r == '_' || r == '-' || r == '.'
	})
	if len(parts) == 0 {
		return strings.Title(input)
	}
	for i := range parts {
		if parts[i] == "" {
			continue
		}
		parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
	}
	return strings.Join(parts, "")
}

func sanitizeAlias(v string) string {
	v = strings.TrimSpace(v)
	v = strings.ReplaceAll(v, ".", "_")
	v = strings.ReplaceAll(v, "-", "_")
	if v == "" {
		return "group"
	}
	return v
}

func sanitizeMetadataKey(v string) string {
	v = strings.TrimSpace(v)
	allowed := map[string]struct{}{
		"environment": {},
		"feature":     {},
		"endpoint":    {},
		"key_type":    {},
		"org_unit_id": {},
		"team_id":     {},
	}
	if _, ok := allowed[v]; ok {
		return v
	}
	return "environment"
}

func interfaceSlice(v interface{}) ([]interface{}, bool) {
	switch t := v.(type) {
	case []interface{}:
		return t, true
	case []string:
		out := make([]interface{}, 0, len(t))
		for _, s := range t {
			out = append(out, s)
		}
		return out, true
	default:
		return nil, false
	}
}

func stringSlice(v interface{}) ([]string, bool) {
	switch t := v.(type) {
	case []string:
		return t, true
	case []interface{}:
		out := make([]string, 0, len(t))
		for _, i := range t {
			out = append(out, fmt.Sprintf("%v", i))
		}
		return out, true
	default:
		return nil, false
	}
}

func parseRFC3339(v string) (time.Time, error) {
	return time.Parse(time.RFC3339, strings.TrimSpace(v))
}

func intFromAny(v interface{}) (int, error) {
	switch t := v.(type) {
	case int:
		return t, nil
	case int64:
		return int(t), nil
	case float64:
		return int(t), nil
	case string:
		return strconv.Atoi(strings.TrimSpace(t))
	default:
		return 0, fmt.Errorf("cannot parse integer from %T", v)
	}
}
