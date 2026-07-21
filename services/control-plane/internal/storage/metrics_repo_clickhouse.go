package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func (s *Store) ensureClickHouseMetricsSchema(ctx context.Context) error {
	if err := s.validateClickHouseColumns(ctx, "gateway_model_events", []string{
		"ts", "tenant_id", "request_id", "model_name", "virtual_model_name", "provider_name", "provider_model_name",
		"request_type", "user_id", "user_email", "virtual_account_id", "virtual_account_name", "team_ids", "team_names",
		"status_code", "error_type", "latency_ms", "ttft_ms", "itl_ms", "tpot_ms", "input_tokens", "output_tokens", "cost_usd",
		"cache_mode", "cache_status", "cache_match", "cache_similarity", "cache_origin_request_id",
		"avoided_input_tokens", "avoided_output_tokens", "avoided_cost_usd", "avoided_cost_irr",
		"embedding_cost_usd", "pricing_version", "metadata_json",
	}); err != nil {
		return err
	}

	if err := s.validateClickHouseColumns(ctx, "gateway_model_metrics_1m", []string{
		"bucket_ts", "tenant_id", "view_model_name", "view_virtual_model_name", "view_user_email",
		"view_virtual_account_name", "view_team_name", "view_error_type", "request_count", "failed_count",
		"input_tokens_sum", "output_tokens_sum", "cost_usd_sum", "latency_p50", "latency_p75",
		"latency_p90", "latency_p99", "ttft_p50", "ttft_p75", "ttft_p90", "ttft_p99",
		"itl_p50", "itl_p75", "itl_p90", "itl_p99", "tpot_p50", "tpot_p75", "tpot_p90", "tpot_p99",
	}); err != nil {
		return err
	}

	return s.validateClickHouseObject(ctx, "mv_gateway_model_metrics_1m")
}

const modelMetricInsertColumns = `
	ts, tenant_id, request_id, model_name, virtual_model_name, provider_name, provider_model_name,
	request_type, user_id, user_email, virtual_account_id, virtual_account_name, team_ids, team_names,
	status_code, error_type, latency_ms, ttft_ms, itl_ms, tpot_ms, input_tokens, output_tokens, cost_usd,
	cache_mode, cache_status, cache_match, cache_similarity, cache_origin_request_id,
	avoided_input_tokens, avoided_output_tokens, avoided_cost_usd, avoided_cost_irr,
	embedding_cost_usd, pricing_version, metadata_json
`

const modelMetricInsertPlaceholders = "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
const modelMetricInsertArgCount = 35

func (s *Store) insertModelMetricEvents(ctx context.Context, events []platform.ModelMetricEvent) error {
	if len(events) == 0 {
		return nil
	}

	values := make([]string, 0, len(events))

	args := make([]interface{}, 0, len(events)*modelMetricInsertArgCount)
	for _, ev := range events {
		values = append(values, modelMetricInsertPlaceholders)
		args = append(args, modelMetricInsertArgs(ev)...)
	}

	insertSQL := fmt.Sprintf(
		"INSERT INTO gateway_model_events (%s) VALUES %s",
		modelMetricInsertColumns,
		strings.Join(values, ", "),
	)

	_, err := s.clickhouse.ExecContext(ctx, insertSQL, args...)

	return err
}

func modelMetricInsertArgs(ev platform.ModelMetricEvent) []interface{} {
	metaRaw := "{}"

	if ev.Metadata != nil {
		if b, err := json.Marshal(ev.Metadata); err == nil {
			metaRaw = string(b)
		}
	}

	return []interface{}{
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
		ev.CacheMode,
		ev.CacheStatus,
		ev.CacheMatch,
		floatPtrToNull(ev.CacheSimilarity),
		nullIfEmpty(ev.CacheOriginRequestID),
		ev.AvoidedInputTokens,
		ev.AvoidedOutputTokens,
		ev.AvoidedCostUSD,
		ev.AvoidedCostIRR,
		ev.EmbeddingCostUSD,
		ev.PricingVersion,
		metaRaw,
	}
}

func (s *Store) queryModelMetrics(ctx context.Context, tenantID string, req platform.MetricsQueryRequest) ([]map[string]interface{}, error) {
	if req.Type == "timeseries" {
		return s.queryModelMetricsTimeseries(ctx, tenantID, req)
	}

	return s.queryModelMetricsDistribution(ctx, tenantID, req)
}

func (s *Store) listModelMetricsFacets(ctx context.Context, tenantID string) (*platform.ModelMetricsFacets, error) {
	models, err := queryDistinctMetricValues(ctx, s.clickhouse, "model_name", tenantID)
	if err != nil {
		return nil, err
	}

	providers, err := queryDistinctMetricValues(ctx, s.clickhouse, "provider_name", tenantID)
	if err != nil {
		return nil, err
	}

	return &platform.ModelMetricsFacets{
		Models:    models,
		Providers: providers,
	}, nil
}

func queryDistinctMetricValues(ctx context.Context, db *sql.DB, column, tenantID string) ([]string, error) {
	query := fmt.Sprintf(`
		SELECT DISTINCT %s
		FROM gateway_model_events
		WHERE tenant_id = ? AND %s != ''
		ORDER BY %s ASC
		LIMIT 5000
	`, column, column, column)

	rows, err := db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer closeBestEffort(rows)

	values := make([]string, 0, 128)

	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}

		value = strings.TrimSpace(value)
		if value != "" {
			values = append(values, value)
		}
	}

	return values, rows.Err()
}

func (s *Store) queryModelMetricsDistribution(ctx context.Context, tenantID string, req platform.MetricsQueryRequest) ([]map[string]interface{}, error) {
	if query, args, ok, err := buildModelMetricsAggregateQuery(tenantID, req, false); err != nil {
		return nil, err
	} else if ok {
		return queryRowsAsMaps(ctx, s.clickhouse, query, args...)
	}

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
	if query, args, ok, err := buildModelMetricsAggregateQuery(tenantID, req, true); err != nil {
		return nil, err
	} else if ok {
		return queryRowsAsMaps(ctx, s.clickhouse, query, args...)
	}

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

type metricAggregateDimension struct {
	Alias     string
	RawExpr   string
	StateExpr string
}

type metricAggregateSpec struct {
	Alias          string
	StateAlias     string
	RawStateExpr   string
	TableStateExpr string
	MergeExpr      string
}

func buildModelMetricsAggregateQuery(tenantID string, req platform.MetricsQueryRequest, timeseries bool) (string, []interface{}, bool, error) {
	aggStart, aggEnd := fullMinuteWindow(req.StartTS, req.EndTS)
	if !aggStart.Before(aggEnd) {
		return "", nil, false, nil
	}

	dimensions, ok, err := buildMetricsAggregateDimensions(req.GroupBy)
	if err != nil || !ok {
		return "", nil, ok, err
	}

	aggregations, ok, err := buildMetricsAggregateSpecs(req.Aggregations)
	if err != nil || !ok {
		return "", nil, ok, err
	}

	rawFilter, aggregateFilter, filterArgs, ok, err := buildMetricsAggregateFilters(req.Filters)
	if err != nil || !ok {
		return "", nil, ok, err
	}

	outerSelects := make([]string, 0, len(dimensions)+len(aggregations)+1)
	outerGroups := make([]string, 0, len(dimensions)+1)
	rawSelects := make([]string, 0, len(dimensions)+len(aggregations)+2)
	tableSelects := make([]string, 0, len(dimensions)+len(aggregations)+2)
	rawGroups := make([]string, 0, len(dimensions)+1)

	if timeseries {
		interval := req.IntervalInSeconds
		if interval <= 0 {
			interval = 300
		}

		if interval < 60 || interval%60 != 0 {
			return "", nil, false, nil
		}

		outerSelects = append(outerSelects, "timestamp")
		outerGroups = append(outerGroups, "timestamp")
		rawSelects = append(rawSelects, fmt.Sprintf("toStartOfInterval(ts, INTERVAL %d second) AS timestamp", interval))
		tableSelects = append(tableSelects, fmt.Sprintf("toStartOfInterval(bucket_ts, INTERVAL %d second) AS timestamp", interval))
		rawGroups = append(rawGroups, "timestamp")
	}

	for _, dim := range dimensions {
		outerSelects = append(outerSelects, dim.Alias)
		outerGroups = append(outerGroups, dim.Alias)
		rawSelects = append(rawSelects, fmt.Sprintf("%s AS %s", dim.RawExpr, dim.Alias))
		tableSelects = append(tableSelects, fmt.Sprintf("%s AS %s", dim.StateExpr, dim.Alias))
		rawGroups = append(rawGroups, dim.Alias)
	}

	for _, agg := range aggregations {
		outerSelects = append(outerSelects, fmt.Sprintf("%s AS %s", agg.MergeExpr, agg.Alias))
		rawSelects = append(rawSelects, fmt.Sprintf("%s AS %s", agg.RawStateExpr, agg.StateAlias))
		tableSelects = append(tableSelects, fmt.Sprintf("%s AS %s", agg.TableStateExpr, agg.StateAlias))
	}

	rawSelects = append(rawSelects, "countState() AS __order_count_state")
	tableSelects = append(tableSelects, "request_count AS __order_count_state")

	rawWhere := "tenant_id = ? AND ts >= ? AND ts <= ? AND NOT (ts >= ? AND ts < ?)"
	tableWhere := "tenant_id = ? AND bucket_ts >= ? AND bucket_ts < ?"

	if rawFilter != "" {
		rawWhere += " AND " + rawFilter
		tableWhere += " AND " + aggregateFilter
	}

	rawQuery := fmt.Sprintf(
		"SELECT %s FROM gateway_model_events WHERE %s",
		strings.Join(rawSelects, ", "),
		rawWhere,
	)
	if len(rawGroups) > 0 {
		rawQuery += " GROUP BY " + strings.Join(rawGroups, ", ")
	}

	tableQuery := fmt.Sprintf(
		"SELECT %s FROM gateway_model_metrics_1m WHERE %s",
		strings.Join(tableSelects, ", "),
		tableWhere,
	)

	unionQuery := tableQuery + " UNION ALL " + rawQuery
	args := make([]interface{}, 0, 10+len(filterArgs)*2)
	args = append(args, tenantID, aggStart, aggEnd)
	args = append(args, filterArgs...)
	args = append(args, tenantID, req.StartTS.UTC(), req.EndTS.UTC(), aggStart, aggEnd)
	args = append(args, filterArgs...)

	if len(outerGroups) == 0 {
		query := fmt.Sprintf(
			"SELECT %s FROM (%s)",
			strings.Join(outerSelects, ", "),
			unionQuery,
		)
		if !timeseries {
			query += " LIMIT 1000"
		}

		return query, args, true, nil
	}

	query := fmt.Sprintf(
		"SELECT %s FROM (%s) GROUP BY %s",
		strings.Join(outerSelects, ", "),
		unionQuery,
		strings.Join(outerGroups, ", "),
	)

	if timeseries {
		query += " ORDER BY timestamp ASC LIMIT 5000"
	} else {
		query += " ORDER BY countMerge(__order_count_state) DESC LIMIT 1000"
	}

	return query, args, true, nil
}

func buildMetricsAggregateDimensions(groupBy []string) ([]metricAggregateDimension, bool, error) {
	out := make([]metricAggregateDimension, 0, len(groupBy))

	for _, g := range groupBy {
		switch g {
		case "modelName":
			out = append(out, metricAggregateDimension{
				Alias:     sanitizeAlias(g),
				RawExpr:   "model_name",
				StateExpr: "view_model_name",
			})
		case "virtualModelName":
			out = append(out, metricAggregateDimension{
				Alias:     sanitizeAlias(g),
				RawExpr:   "virtual_model_name",
				StateExpr: "view_virtual_model_name",
			})
		case "userEmail", "virtualAccount", "team", "errorCode":
			return nil, false, nil
		default:
			if strings.HasPrefix(g, "metadata.") {
				return nil, false, nil
			}

			if _, err := mapMetricsGroupBy(g); err != nil {
				return nil, false, err
			}

			return nil, false, nil
		}
	}

	return out, true, nil
}

func buildMetricsAggregateSpecs(aggs []platform.MetricsAggregation) ([]metricAggregateSpec, bool, error) {
	if len(aggs) == 0 {
		return []metricAggregateSpec{{
			Alias:          "countAll",
			StateAlias:     "__countAll_state",
			RawStateExpr:   "countState()",
			TableStateExpr: "request_count",
			MergeExpr:      "countMerge(__countAll_state)",
		}}, true, nil
	}

	out := make([]metricAggregateSpec, 0, len(aggs))
	for _, agg := range aggs {
		spec, ok, err := buildMetricsAggregateSpec(agg)
		if err != nil || !ok {
			return nil, ok, err
		}

		out = append(out, spec)
	}

	return out, true, nil
}

func buildMetricsAggregateSpec(agg platform.MetricsAggregation) (metricAggregateSpec, bool, error) {
	pascal := toPascal(agg.Column)
	switch agg.Type {
	case "count":
		if !isAggregateCountColumn(agg.Column) {
			if _, err := mapAggregationColumn(agg.Column); err != nil {
				return metricAggregateSpec{}, false, err
			}

			return metricAggregateSpec{}, false, nil
		}

		alias := "count" + pascal

		return metricAggregateSpec{
			Alias:          alias,
			StateAlias:     "__" + alias + "_state",
			RawStateExpr:   "countState()",
			TableStateExpr: "request_count",
			MergeExpr:      "countMerge(__" + alias + "_state)",
		}, true, nil
	case "sum":
		rawExpr, tableExpr, ok, err := mapAggregateSumState(agg.Column)
		if err != nil || !ok {
			return metricAggregateSpec{}, ok, err
		}

		alias := "sum" + pascal

		return metricAggregateSpec{
			Alias:          alias,
			StateAlias:     "__" + alias + "_state",
			RawStateExpr:   rawExpr,
			TableStateExpr: tableExpr,
			MergeExpr:      "sumMerge(__" + alias + "_state)",
		}, true, nil
	case "p50", "p75", "p90", "p99":
		rawExpr, tableExpr, q, ok, err := mapAggregateQuantileState(agg.Type, agg.Column)
		if err != nil || !ok {
			return metricAggregateSpec{}, ok, err
		}

		alias := agg.Type + pascal

		return metricAggregateSpec{
			Alias:          alias,
			StateAlias:     "__" + alias + "_state",
			RawStateExpr:   rawExpr,
			TableStateExpr: tableExpr,
			MergeExpr:      fmt.Sprintf("quantileTDigestMerge(%s)(__%s_state)", q, alias),
		}, true, nil
	default:
		return metricAggregateSpec{}, false, fmt.Errorf("unsupported aggregation type: %s", agg.Type)
	}
}

func isAggregateCountColumn(column string) bool {
	switch column {
	case "", "modelName", "costInUSD", "inputTokens", "outputTokens", "latencyMs":
		return true
	default:
		return false
	}
}

func mapAggregateSumState(column string) (string, string, bool, error) {
	switch column {
	case "inputTokens":
		return "sumState(toUInt64(input_tokens))", "input_tokens_sum", true, nil
	case "outputTokens":
		return "sumState(toUInt64(output_tokens))", "output_tokens_sum", true, nil
	case "costInUSD":
		return "sumState(cost_usd)", "cost_usd_sum", true, nil
	default:
		if _, err := mapAggregationColumn(column); err != nil {
			return "", "", false, err
		}

		return "", "", false, nil
	}
}

func mapAggregateQuantileState(aggType, column string) (string, string, string, bool, error) {
	q, ok := map[string]string{"p50": "0.50", "p75": "0.75", "p90": "0.90", "p99": "0.99"}[aggType]
	if !ok {
		return "", "", "", false, fmt.Errorf("unsupported aggregation type: %s", aggType)
	}

	switch column {
	case "latencyMs":
		suffix := strings.TrimPrefix(aggType, "p")
		return fmt.Sprintf("quantileTDigestState(%s)(latency_ms)", q), "latency_p" + suffix, q, true, nil
	case "timeToFirstTokenMs":
		suffix := strings.TrimPrefix(aggType, "p")
		return fmt.Sprintf("quantileTDigestStateIf(%s)(ttft_ms, ttft_ms IS NOT NULL)", q), "ttft_p" + suffix, q, true, nil
	case "interTokenLatencyMs":
		suffix := strings.TrimPrefix(aggType, "p")
		return fmt.Sprintf("quantileTDigestStateIf(%s)(itl_ms, itl_ms IS NOT NULL)", q), "itl_p" + suffix, q, true, nil
	case "timePerOutputTokenLatencyMs":
		suffix := strings.TrimPrefix(aggType, "p")
		return fmt.Sprintf("quantileTDigestStateIf(%s)(tpot_ms, tpot_ms IS NOT NULL)", q), "tpot_p" + suffix, q, true, nil
	default:
		if _, err := mapAggregationColumn(column); err != nil {
			return "", "", "", false, err
		}

		return "", "", "", false, nil
	}
}

func buildMetricsAggregateFilters(filters []platform.MetricsFilter) (string, string, []interface{}, bool, error) {
	rawParts := make([]string, 0, len(filters))
	aggregateParts := make([]string, 0, len(filters))
	args := make([]interface{}, 0, len(filters))

	for _, f := range filters {
		rawCol, aggregateCol, ok, err := mapAggregateFilterField(f.FieldName, f.MetadataKey)
		if err != nil || !ok {
			return "", "", nil, ok, err
		}

		op := strings.ToUpper(strings.TrimSpace(f.Operator))
		switch op {
		case "EQUAL":
			rawParts = append(rawParts, rawCol+" = ?")
			aggregateParts = append(aggregateParts, aggregateCol+" = ?")
			args = append(args, f.Value)
		case "IN", "NOT_IN":
			values, ok := interfaceSlice(f.Value)
			if !ok || len(values) == 0 {
				return "", "", nil, false, fmt.Errorf("%s filter requires non-empty array value", op)
			}

			operator := "IN"
			if op == "NOT_IN" {
				operator = "NOT IN"
			}

			placeholders := strings.TrimRight(strings.Repeat("?,", len(values)), ",")
			rawParts = append(rawParts, rawCol+" "+operator+" ("+placeholders+")")
			aggregateParts = append(aggregateParts, aggregateCol+" "+operator+" ("+placeholders+")")
			args = append(args, values...)
		case "STRING_CONTAINS":
			rawParts = append(rawParts, rawCol+" LIKE ?")
			aggregateParts = append(aggregateParts, aggregateCol+" LIKE ?")
			args = append(args, "%"+fmt.Sprintf("%v", f.Value)+"%")
		case "STRING_STARTS_WITH":
			rawParts = append(rawParts, rawCol+" LIKE ?")
			aggregateParts = append(aggregateParts, aggregateCol+" LIKE ?")
			args = append(args, fmt.Sprintf("%v", f.Value)+"%")
		case "STRING_ENDS_WITH":
			rawParts = append(rawParts, rawCol+" LIKE ?")
			aggregateParts = append(aggregateParts, aggregateCol+" LIKE ?")
			args = append(args, "%"+fmt.Sprintf("%v", f.Value))
		case "IS_NULL":
			isNull := false

			switch v := f.Value.(type) {
			case bool:
				isNull = v
			case string:
				isNull = strings.EqualFold(v, "true")
			}

			if isNull {
				rawParts = append(rawParts, "("+rawCol+" IS NULL OR "+rawCol+" = '')")
				aggregateParts = append(aggregateParts, "("+aggregateCol+" IS NULL OR "+aggregateCol+" = '')")
			} else {
				rawParts = append(rawParts, "("+rawCol+" IS NOT NULL AND "+rawCol+" != '')")
				aggregateParts = append(aggregateParts, "("+aggregateCol+" IS NOT NULL AND "+aggregateCol+" != '')")
			}
		default:
			if _, _, err := buildMetricsWhere("tenant", platform.MetricsQueryRequest{
				StartTS: time.Unix(1, 0),
				EndTS:   time.Unix(2, 0),
				Filters: []platform.MetricsFilter{f},
			}); err != nil {
				return "", "", nil, false, err
			}

			return "", "", nil, false, nil
		}
	}

	return strings.Join(rawParts, " AND "), strings.Join(aggregateParts, " AND "), args, true, nil
}

func mapAggregateFilterField(fieldName, metadataKey string) (string, string, bool, error) {
	if strings.TrimSpace(metadataKey) != "" {
		return "", "", false, nil
	}

	switch fieldName {
	case "modelName":
		return "model_name", "view_model_name", true, nil
	case "virtualModelName":
		return "virtual_model_name", "view_virtual_model_name", true, nil
	default:
		if _, err := mapFilterField(fieldName); err != nil {
			return "", "", false, err
		}

		return "", "", false, nil
	}
}

func fullMinuteWindow(start, end time.Time) (time.Time, time.Time) {
	start = start.UTC()

	end = end.UTC()
	if !start.Before(end) {
		return time.Time{}, time.Time{}
	}

	aggStart := start.Truncate(time.Minute)
	if !start.Equal(aggStart) {
		aggStart = aggStart.Add(time.Minute)
	}

	aggEnd := end.Truncate(time.Minute)

	return aggStart, aggEnd
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
		var col string
		if f.MetadataKey != "" {
			col = fmt.Sprintf("JSONExtractString(metadata_json, '%s')", sanitizeMetadataKey(f.MetadataKey))
		} else {
			mapped, err := mapFilterField(f.FieldName)
			if err != nil {
				return "", nil, err
			}

			col = mapped
		}

		arrayFilter := isMetricsArrayFilterField(f.FieldName, f.MetadataKey)

		op := strings.ToUpper(strings.TrimSpace(f.Operator))
		switch op {
		case "EQUAL":
			if arrayFilter {
				where = append(where, fmt.Sprintf("has(%s, ?)", col))
			} else {
				where = append(where, col+" = ?")
			}

			args = append(args, f.Value)
		case "IN":
			if arrayFilter {
				values, ok := stringSlice(f.Value)
				if !ok || len(values) == 0 {
					return "", nil, fmt.Errorf("IN filter requires non-empty array value")
				}

				where = append(where, fmt.Sprintf("hasAny(%s, ?)", col))
				args = append(args, values)

				continue
			}

			values, ok := interfaceSlice(f.Value)
			if !ok || len(values) == 0 {
				return "", nil, fmt.Errorf("IN filter requires non-empty array value")
			}

			where = append(where, col+" IN ("+strings.TrimRight(strings.Repeat("?,", len(values)), ",")+")")
			args = append(args, values...)
		case "NOT_IN":
			if arrayFilter {
				values, ok := stringSlice(f.Value)
				if !ok || len(values) == 0 {
					return "", nil, fmt.Errorf("NOT_IN filter requires non-empty array value")
				}

				where = append(where, fmt.Sprintf("NOT hasAny(%s, ?)", col))
				args = append(args, values)

				continue
			}

			values, ok := interfaceSlice(f.Value)
			if !ok || len(values) == 0 {
				return "", nil, fmt.Errorf("NOT_IN filter requires non-empty array value")
			}

			where = append(where, col+" NOT IN ("+strings.TrimRight(strings.Repeat("?,", len(values)), ",")+")")
			args = append(args, values...)
		case "STRING_CONTAINS":
			if arrayFilter {
				where = append(where, fmt.Sprintf("arrayExists(x -> x LIKE ?, %s)", col))
			} else {
				where = append(where, col+" LIKE ?")
			}

			args = append(args, "%"+fmt.Sprintf("%v", f.Value)+"%")
		case "STRING_STARTS_WITH":
			if arrayFilter {
				where = append(where, fmt.Sprintf("arrayExists(x -> x LIKE ?, %s)", col))
			} else {
				where = append(where, col+" LIKE ?")
			}

			args = append(args, fmt.Sprintf("%v", f.Value)+"%")
		case "STRING_ENDS_WITH":
			if arrayFilter {
				where = append(where, fmt.Sprintf("arrayExists(x -> x LIKE ?, %s)", col))
			} else {
				where = append(where, col+" LIKE ?")
			}

			args = append(args, "%"+fmt.Sprintf("%v", f.Value))
		case "GREATER_THAN":
			if arrayFilter {
				return "", nil, fmt.Errorf("unsupported array filter operator: %s", op)
			}

			where = append(where, col+" > ?")
			args = append(args, f.Value)
		case "LESS_THAN":
			if arrayFilter {
				return "", nil, fmt.Errorf("unsupported array filter operator: %s", op)
			}

			where = append(where, col+" < ?")
			args = append(args, f.Value)
		case "GREATER_THAN_EQUAL":
			if arrayFilter {
				return "", nil, fmt.Errorf("unsupported array filter operator: %s", op)
			}

			where = append(where, col+" >= ?")
			args = append(args, f.Value)
		case "LESS_THAN_EQUAL":
			if arrayFilter {
				return "", nil, fmt.Errorf("unsupported array filter operator: %s", op)
			}

			where = append(where, col+" <= ?")
			args = append(args, f.Value)
		case "BETWEEN":
			if arrayFilter {
				return "", nil, fmt.Errorf("unsupported array filter operator: %s", op)
			}

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
				if arrayFilter {
					where = append(where, fmt.Sprintf("empty(%s)", col))
				} else {
					where = append(where, "("+col+" IS NULL OR "+col+" = '')")
				}
			} else {
				if arrayFilter {
					where = append(where, fmt.Sprintf("NOT empty(%s)", col))
				} else {
					where = append(where, "("+col+" IS NOT NULL AND "+col+" != '')")
				}
			}
		default:
			return "", nil, fmt.Errorf("unsupported filter operator: %s", op)
		}
	}

	return strings.Join(where, " AND "), args, nil
}

func isMetricsArrayFilterField(fieldName, metadataKey string) bool {
	return strings.TrimSpace(metadataKey) == "" && fieldName == "team"
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
	case "avoidedInputTokens":
		return "avoided_input_tokens", nil
	case "avoidedOutputTokens":
		return "avoided_output_tokens", nil
	case "avoidedCostUSD":
		return "avoided_cost_usd", nil
	case "avoidedCostIRR":
		return "avoided_cost_irr", nil
	case "embeddingCostUSD":
		return "embedding_cost_usd", nil
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
	case g == "cacheMode":
		return "cache_mode", nil
	case g == "cacheStatus":
		return "cache_status", nil
	case g == "cacheMatch":
		return "cache_match", nil
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
	case "requestID":
		return "request_id", nil
	case "userEmail":
		return "user_email", nil
	case "virtualAccount":
		return "virtual_account_name", nil
	case "providerName":
		return "provider_name", nil
	case "providerModelName":
		return "provider_model_name", nil
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
	case "cacheMode":
		return "cache_mode", nil
	case "cacheStatus":
		return "cache_status", nil
	case "cacheMatch":
		return "cache_match", nil
	case "cacheSimilarity":
		return "cache_similarity", nil
	default:
		return "", fmt.Errorf("unsupported filter field: %s", f)
	}
}

func queryRowsAsMaps(ctx context.Context, db *sql.DB, query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer closeBestEffort(rows)

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

func floatPtrToNull(v *float64) interface{} {
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
		return cases.Title(language.Und).String(input)
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
