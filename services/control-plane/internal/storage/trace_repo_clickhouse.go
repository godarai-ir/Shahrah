package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

func (s *Store) initClickHouse(ctx context.Context) error {
	dsn := strings.TrimSpace(os.Getenv("CLICKHOUSE_DSN"))
	if dsn == "" {
		return fmt.Errorf("CLICKHOUSE_DSN is required for clickhouse observability storage")
	}

	db, err := sql.Open("clickhouse", dsn)
	if err != nil {
		return err
	}

	db.SetMaxOpenConns(16)
	db.SetMaxIdleConns(8)
	db.SetConnMaxLifetime(15 * time.Minute)

	if err := db.PingContext(ctx); err != nil {
		closeBestEffort(db)
		return err
	}

	s.clickhouse = db
	s.clickhouseReady = true
	s.clickhouseTraceReady = s.enableClickHouseSubsystem(ctx, "traces", s.ensureClickHouseTraceSchema)
	s.clickhouseMetricsReady = s.enableClickHouseSubsystem(ctx, "metrics", s.ensureClickHouseMetricsSchema)
	s.clickhouseLogsReady = s.enableClickHouseSubsystem(ctx, "logs", s.ensureClickHouseRequestLogsSchema)

	logger := s.logger
	if logger == nil {
		logger = zap.NewNop()
	}

	if s.clickhouseTraceReady || s.clickhouseMetricsReady || s.clickhouseLogsReady {
		logger.Info(
			"clickhouse observability backend enabled",
			zap.Bool("traces_ready", s.clickhouseTraceReady),
			zap.Bool("metrics_ready", s.clickhouseMetricsReady),
			zap.Bool("logs_ready", s.clickhouseLogsReady),
		)
	} else {
		logger.Warn("clickhouse connected but no observability subsystem is ready")
	}

	return nil
}

func (s *Store) enableClickHouseSubsystem(ctx context.Context, name string, validate func(context.Context) error) bool {
	if err := validate(ctx); err != nil {
		logger := s.logger
		if logger == nil {
			logger = zap.NewNop()
		}
		logger.Warn("clickhouse observability subsystem unavailable", zap.String("subsystem", name), zap.Error(err))
		return false
	}

	return true
}

func (s *Store) clickHouseTraceReady() bool {
	return s != nil && s.clickhouse != nil && s.clickhouseTraceReady
}

func (s *Store) clickHouseMetricsReady() bool {
	return s != nil && s.clickhouse != nil && s.clickhouseMetricsReady
}

func (s *Store) clickHouseLogsReady() bool {
	return s != nil && s.clickhouse != nil && s.clickhouseLogsReady
}

func (s *Store) ClickHouseReadinessReport() map[string]interface{} {
	if s == nil {
		return map[string]interface{}{
			"connected": false,
			"traces":    false,
			"metrics":   false,
			"logs":      false,
		}
	}

	return map[string]interface{}{
		"connected": s.clickhouse != nil && s.clickhouseReady,
		"traces":    s.clickHouseTraceReady(),
		"metrics":   s.clickHouseMetricsReady(),
		"logs":      s.clickHouseLogsReady(),
	}
}

func (s *Store) ensureClickHouseTraceSchema(ctx context.Context) error {
	if err := s.validateClickHouseColumns(ctx, "request_traces_ch", []string{
		"trace_id", "tenant_id", "request_id", "model", "route", "status_code", "started_at",
		"duration_ms", "error_code", "sampled", "retention_days",
	}); err != nil {
		return err
	}

	if err := s.validateClickHouseColumns(ctx, "request_spans_ch", []string{
		"span_id", "trace_id", "tenant_id", "parent_span_id", "name", "kind", "started_at",
		"duration_ms", "status", "attributes_json", "retention_days",
	}); err != nil {
		return err
	}

	return nil
}

func (s *Store) exportToClickHouse(ctx context.Context, batch []tracing.SpanEvent) error {
	if len(batch) == 0 {
		return nil
	}

	const (
		traceInsertSQL = `
		INSERT INTO request_traces_ch
		(trace_id, tenant_id, request_id, model, route, status_code, started_at, duration_ms, error_code, sampled, retention_days)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
		spanInsertSQL = `
		INSERT INTO request_spans_ch
		(span_id, trace_id, tenant_id, parent_span_id, name, kind, started_at, duration_ms, status, attributes_json, retention_days)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	)

	retentionDaysByTenant := map[string]uint16{}
	for _, ev := range batch {
		tenantID := strings.TrimSpace(ev.TenantID)
		if strings.TrimSpace(ev.TraceID) == "" || tenantID == "" {
			continue
		}
		retentionDays, ok := retentionDaysByTenant[tenantID]
		if !ok {
			retentionDays = s.traceRetentionDaysForInsert(ctx, tenantID)
			retentionDaysByTenant[tenantID] = retentionDays
		}

		errorCode := ""
		status := "ok"

		if ev.StatusCode >= 500 {
			errorCode = "upstream_error"
			status = "error"
		}

		model := strings.TrimSpace(ev.Model)
		if model == "" && ev.Attributes != nil {
			if raw, ok := ev.Attributes["model"]; ok {
				if s, ok := raw.(string); ok {
					model = strings.TrimSpace(s)
				}
			}
		}

		if isRootRequestSpan(ev) {
			requestID := ev.TraceID

			_, err := s.clickhouse.ExecContext(
				ctx,
				traceInsertSQL,
				ev.TraceID,
				tenantID,
				requestID,
				model,
				ev.Route,
				int32(ev.StatusCode),
				ev.StartedAt.UTC(),
				int32(ev.DurationMS),
				errorCode,
				boolToUint8(ev.Sampled),
				retentionDays,
			)
			if err != nil {
				return err
			}
		}

		attrs := map[string]interface{}{
			"http.method": ev.Method,
			"http.route":  ev.Route,
			"sampled":     ev.Sampled,
		}
		for k, v := range ev.Attributes {
			attrs[k] = v
		}

		attrsRaw := marshalJSONBestEffort(attrs)

		spanID := ev.SpanID
		if strings.TrimSpace(spanID) == "" {
			spanID = fmt.Sprintf("%s:%d", ev.TraceID, ev.StartedAt.UnixNano())
		}

		parentSpanID := ev.ParentSpanID

		spanName := ev.Name
		if strings.TrimSpace(spanName) == "" {
			spanName = "gateway.request"
		}

		spanKind := ev.Kind
		if strings.TrimSpace(spanKind) == "" {
			spanKind = "server"
		}

		_, err := s.clickhouse.ExecContext(
			ctx,
			spanInsertSQL,
			spanID,
			ev.TraceID,
			tenantID,
			parentSpanID,
			spanName,
			spanKind,
			ev.StartedAt.UTC(),
			int32(ev.DurationMS),
			status,
			string(attrsRaw),
			retentionDays,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) traceRetentionDaysForInsert(ctx context.Context, tenantID string) uint16 {
	days, err := s.ResolveTraceRetentionDays(ctx, tenantID)
	if err != nil {
		logger := s.logger
		if logger == nil {
			logger = zap.NewNop()
		}
		logger.Warn("failed to resolve trace retention; writing trace with unlimited ttl", zap.String("tenant_id", tenantID), zap.Error(err))

		return 0
	}

	if days == nil {
		return clickHouseTraceRetentionDaysValue(nil)
	}

	return clickHouseTraceRetentionDaysValue(days)
}

func clickHouseTraceRetentionDaysValue(days *int) uint16 {
	if days == nil {
		return 0
	}

	if *days > maxClickHouseRetentionDays {
		return maxClickHouseRetentionDays
	}

	if *days <= 0 {
		return 0
	}

	return uint16(*days)
}

func (s *Store) listRequestTracesFromClickHouse(ctx context.Context, params platform.ListRequestTracesParams) ([]platform.RequestTrace, int, error) {
	if params.Limit <= 0 {
		params.Limit = 50
	}

	if params.Limit > 200 {
		params.Limit = 200
	}

	if params.Offset < 0 {
		params.Offset = 0
	}

	where := []string{"tenant_id = ?"}

	args := []interface{}{params.TenantID}
	if params.TraceID != "" {
		where = append(where, "trace_id = ?")
		args = append(args, params.TraceID)
	}

	if params.Route != "" {
		where = append(where, "route = ?")
		args = append(args, params.Route)
	}

	if params.Model != "" {
		where = append(where, "model = ?")
		args = append(args, params.Model)
	}

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

	whereSQL := strings.Join(where, " AND ")
	dedupSubquery := buildTraceDedupeSubquery(whereSQL)

	query := fmt.Sprintf(`
		SELECT trace_id, tenant_id, request_id, model, route, status_code, started_at, duration_ms, error_code, sampled
		FROM (%s)
		ORDER BY started_at DESC
		LIMIT ? OFFSET ?
	`, dedupSubquery)

	pageLimit := params.Limit + 1
	args = append(args, pageLimit, params.Offset)

	rows, err := s.clickhouse.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer closeBestEffort(rows)

	items := make([]platform.RequestTrace, 0, pageLimit)

	for rows.Next() {
		var (
			item    platform.RequestTrace
			sampled uint8
		)
		if err := rows.Scan(
			&item.TraceID,
			&item.TenantID,
			&item.RequestID,
			&item.Model,
			&item.Route,
			&item.StatusCode,
			&item.StartedAt,
			&item.DurationMS,
			&item.ErrorCode,
			&sampled,
		); err != nil {
			return nil, 0, err
		}

		item.Sampled = sampled == 1
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

func (s *Store) getRequestTraceFromClickHouse(ctx context.Context, tenantID, traceID string) (*platform.RequestTrace, error) {
	var (
		item    platform.RequestTrace
		sampled uint8
	)

	err := s.clickhouse.QueryRowContext(ctx, `
		SELECT trace_id, tenant_id, request_id, model, route, status_code, started_at, duration_ms, error_code, sampled
		FROM (`+buildTraceDedupeSubquery("tenant_id = ? AND trace_id = ?")+`)
		LIMIT 1
	`, tenantID, traceID).Scan(
		&item.TraceID,
		&item.TenantID,
		&item.RequestID,
		&item.Model,
		&item.Route,
		&item.StatusCode,
		&item.StartedAt,
		&item.DurationMS,
		&item.ErrorCode,
		&sampled,
	)
	if err != nil {
		if errorsIsNoRows(err) {
			return nil, pgx.ErrNoRows
		}

		return nil, err
	}

	item.Sampled = sampled == 1

	return &item, nil
}

func (s *Store) listTraceSpansFromClickHouse(ctx context.Context, tenantID, traceID string) ([]platform.TraceSpan, error) {
	if _, err := s.getRequestTraceFromClickHouse(ctx, tenantID, traceID); err != nil {
		return nil, err
	}

	rows, err := s.clickhouse.QueryContext(ctx, `
		SELECT span_id, trace_id, parent_span_id, name, kind, started_at, duration_ms, status, attributes_json
		FROM request_spans_ch
		WHERE tenant_id = ? AND trace_id = ?
		ORDER BY started_at ASC
	`, tenantID, traceID)
	if err != nil {
		return nil, err
	}
	defer closeBestEffort(rows)

	out := make([]platform.TraceSpan, 0, 8)

	for rows.Next() {
		var (
			item     platform.TraceSpan
			attrsRaw string
		)
		if err := rows.Scan(
			&item.SpanID,
			&item.TraceID,
			&item.ParentSpanID,
			&item.Name,
			&item.Kind,
			&item.StartedAt,
			&item.DurationMS,
			&item.Status,
			&attrsRaw,
		); err != nil {
			return nil, err
		}

		if strings.TrimSpace(attrsRaw) != "" {
			if err := json.Unmarshal([]byte(attrsRaw), &item.Attributes); err != nil {
				return nil, err
			}
		}

		if item.Attributes == nil {
			item.Attributes = map[string]interface{}{}
		}

		out = append(out, item)
	}

	return out, rows.Err()
}

func (s *Store) listRequestTracesV2FromClickHouse(ctx context.Context, params platform.ListRequestTracesV2Params) ([]platform.RequestTraceV2, int, error) {
	if params.Limit <= 0 {
		params.Limit = 50
	}

	if params.Limit > 200 {
		params.Limit = 200
	}

	if params.Offset < 0 {
		params.Offset = 0
	}

	sortBy := normalizeTraceSortBy(params.SortBy)
	sortOrder := normalizeSortOrder(params.SortOrder)

	baseWhereSQL, baseArgs := buildTraceV2BaseWhere(params)
	dedupSubquery := buildTraceDedupeSubquery(baseWhereSQL)
	spanWhereSQL, spanArgs := buildTraceProviderSpanWhere(params)
	providerSubquery := `
		SELECT
			trace_id,
			argMin(JSONExtractString(attributes_json, 'provider'), started_at) AS provider
		FROM request_spans_ch
		WHERE ` + spanWhereSQL + `
		GROUP BY trace_id
	`
	attemptsSubquery := `
		SELECT
			trace_id,
			count() AS attempts
		FROM request_spans_ch
		WHERE ` + spanWhereSQL + `
		GROUP BY trace_id
	`
	outerWhere := []string{"1 = 1"}
	outerArgs := []interface{}{}

	if strings.TrimSpace(params.Provider) != "" {
		outerWhere = append(outerWhere, "p.provider = ?")
		outerArgs = append(outerArgs, strings.TrimSpace(params.Provider))
	}

	baseFrom := fmt.Sprintf(`
		FROM (%s) t
		LEFT JOIN (%s) p ON t.trace_id = p.trace_id
		LEFT JOIN (%s) a ON t.trace_id = a.trace_id
	`, dedupSubquery, providerSubquery, attemptsSubquery)

	query := fmt.Sprintf(`
		SELECT
			t.trace_id,
			t.tenant_id,
			t.request_id,
			t.model,
			coalesce(p.provider, '') AS provider,
			t.route,
			t.status_code,
			t.started_at,
			t.duration_ms,
			t.error_code,
			t.sampled,
			toInt32(coalesce(a.attempts, 0)) AS attempts
		%s
		WHERE %s
		ORDER BY %s %s
		LIMIT ? OFFSET ?
	`, baseFrom, strings.Join(outerWhere, " AND "), sortBy, sortOrder)

	args := append([]interface{}{}, baseArgs...)
	args = append(args, spanArgs...)
	args = append(args, spanArgs...)
	args = append(args, outerArgs...)
	countQuery := fmt.Sprintf(`
		SELECT count()
		%s
		WHERE %s
	`, baseFrom, strings.Join(outerWhere, " AND "))
	countArgs := append([]interface{}{}, args...)
	var total uint64
	if err := s.clickhouse.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}
	pageLimit := params.Limit + 1
	args = append(args, pageLimit, params.Offset)

	rows, err := s.clickhouse.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer closeBestEffort(rows)

	items := make([]platform.RequestTraceV2, 0, pageLimit)

	for rows.Next() {
		var (
			item     platform.RequestTraceV2
			sampled  uint8
			attempts int32
		)
		if err := rows.Scan(
			&item.TraceID,
			&item.TenantID,
			&item.RequestID,
			&item.Model,
			&item.Provider,
			&item.Route,
			&item.StatusCode,
			&item.StartedAt,
			&item.DurationMS,
			&item.ErrorCode,
			&sampled,
			&attempts,
		); err != nil {
			return nil, 0, err
		}

		item.Sampled = sampled == 1
		item.Attempts = int(attempts)
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	hasMore := len(items) > params.Limit
	if hasMore {
		items = items[:params.Limit]
	}

	return items, int(total), nil
}

func (s *Store) initClickHouseWithRetry(ctx context.Context) error {
	attempts := intEnvOrDefault("CLICKHOUSE_INIT_ATTEMPTS", 10)
	if attempts < 1 {
		attempts = 1
	}
	delay := secondsEnvOrDefault("CLICKHOUSE_INIT_RETRY_SECONDS", time.Second)

	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		if err := s.initClickHouse(ctx); err == nil {
			return nil
		} else {
			lastErr = err
		}

		if attempt == attempts {
			break
		}
		if s.logger != nil {
			s.logger.Warn("clickhouse initialization failed; retrying", zap.Int("attempt", attempt), zap.Int("max_attempts", attempts), zap.Duration("retry_in", delay), zap.Error(lastErr))
		}

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return fmt.Errorf("wait to retry clickhouse initialization: %w", ctx.Err())
		case <-timer.C:
		}
	}

	return lastErr
}

func (s *Store) getRequestTraceV2FromClickHouse(ctx context.Context, tenantID, traceID string) (*platform.RequestTraceDetailV2, error) {
	trace, err := s.getRequestTraceFromClickHouse(ctx, tenantID, traceID)
	if err != nil {
		return nil, err
	}

	spans, err := s.listTraceSpansFromClickHouse(ctx, tenantID, traceID)
	if err != nil {
		return nil, err
	}

	attempts, err := s.listTraceAttemptsV2FromClickHouse(ctx, tenantID, traceID)
	if err != nil {
		return nil, err
	}

	errorsCount := 0

	for _, span := range spans {
		if strings.EqualFold(span.Status, "error") {
			errorsCount++
		}
	}

	provider := ""
	if len(attempts) > 0 {
		provider = attempts[0].Provider
	}

	detail := &platform.RequestTraceDetailV2{
		Trace: platform.RequestTraceV2{
			TraceID:    trace.TraceID,
			TenantID:   trace.TenantID,
			RequestID:  trace.RequestID,
			Model:      trace.Model,
			Provider:   provider,
			Route:      trace.Route,
			StatusCode: trace.StatusCode,
			StartedAt:  trace.StartedAt,
			DurationMS: trace.DurationMS,
			ErrorCode:  trace.ErrorCode,
			Sampled:    trace.Sampled,
			Attempts:   len(attempts),
		},
		Summary: platform.RequestTraceSummaryV2{
			SpansCount:    len(spans),
			AttemptsCount: len(attempts),
			ErrorsCount:   errorsCount,
		},
	}

	usage, err := s.getTraceUsageFromPostgres(ctx, tenantID, trace.RequestID)
	if err != nil {
		return nil, err
	}

	detail.Usage = usage

	metric, err := s.getTraceMetricFromClickHouse(ctx, tenantID, trace.RequestID)
	if err != nil {
		return nil, err
	}

	detail.Metric = metric
	if detail.Trace.Provider == "" && metric != nil {
		detail.Trace.Provider = metric.ProviderName
	}

	return detail, nil
}

func (s *Store) listTraceAttemptsV2FromClickHouse(ctx context.Context, tenantID, traceID string) ([]platform.TraceAttemptV2, error) {
	if _, err := s.getRequestTraceFromClickHouse(ctx, tenantID, traceID); err != nil {
		return nil, err
	}

	rows, err := s.clickhouse.QueryContext(ctx, `
		SELECT started_at, duration_ms, status, attributes_json
		FROM request_spans_ch
		WHERE tenant_id = ? AND trace_id = ? AND name = 'provider.call'
		ORDER BY started_at ASC
	`, tenantID, traceID)
	if err != nil {
		return nil, err
	}
	defer closeBestEffort(rows)

	out := make([]platform.TraceAttemptV2, 0, 4)

	for rows.Next() {
		var (
			startedAt time.Time
			duration  int32
			status    string
			attrsRaw  string
		)
		if err := rows.Scan(&startedAt, &duration, &status, &attrsRaw); err != nil {
			return nil, err
		}

		attrs := map[string]interface{}{}
		if strings.TrimSpace(attrsRaw) != "" {
			if err := json.Unmarshal([]byte(attrsRaw), &attrs); err != nil {
				return nil, err
			}
		}

		item := platform.TraceAttemptV2{
			AttemptNo:  len(out) + 1,
			TraceID:    traceID,
			StartedAt:  startedAt,
			DurationMS: int(duration),
			Status:     status,
			Attributes: attrs,
		}
		if v, ok := attrs["provider"].(string); ok {
			item.Provider = strings.TrimSpace(v)
		}

		if v, ok := attrs["model"].(string); ok {
			item.Model = strings.TrimSpace(v)
		} else if v, ok := attrs["selected_model"].(string); ok {
			item.Model = strings.TrimSpace(v)
		}

		if v, ok := attrs["retry_reason"].(string); ok {
			item.RetryReason = strings.TrimSpace(v)
		}

		if v, ok := attrs["failure_reason"].(string); ok {
			item.FallbackReason = strings.TrimSpace(v)
		}

		out = append(out, item)
	}

	return out, rows.Err()
}

func (s *Store) getTraceFacetsV2FromClickHouse(ctx context.Context, params platform.ListRequestTracesV2Params) (*platform.TraceFacetsV2, error) {
	baseWhereSQL, baseArgs := buildTraceV2BaseWhere(params)
	dedupSubquery := buildTraceDedupeSubquery(baseWhereSQL)
	providerSubquery := `
		SELECT
			trace_id,
			argMin(JSONExtractString(attributes_json, 'provider'), started_at) AS provider
		FROM request_spans_ch
		WHERE tenant_id = ? AND name = 'provider.call'
		GROUP BY trace_id
	`
	attemptsSubquery := `
		SELECT
			trace_id,
			count() AS attempts
		FROM request_spans_ch
		WHERE tenant_id = ? AND name = 'provider.call'
		GROUP BY trace_id
	`
	outerWhere := []string{"1 = 1"}
	outerArgs := []interface{}{}

	if strings.TrimSpace(params.Provider) != "" {
		outerWhere = append(outerWhere, "p.provider = ?")
		outerArgs = append(outerArgs, strings.TrimSpace(params.Provider))
	}

	baseFrom := fmt.Sprintf(`
		FROM (%s) t
		LEFT JOIN (%s) p ON t.trace_id = p.trace_id
		LEFT JOIN (%s) a ON t.trace_id = a.trace_id
	`, dedupSubquery, providerSubquery, attemptsSubquery)

	sharedArgs := append([]interface{}{}, baseArgs...)
	sharedArgs = append(sharedArgs, params.TenantID, params.TenantID)
	sharedArgs = append(sharedArgs, outerArgs...)
	whereSQL := strings.Join(outerWhere, " AND ")

	providers, err := queryFacetStringRows(ctx, s.clickhouse, buildTraceFacetStringQuery(baseFrom, whereSQL, "coalesce(p.provider, '')"), sharedArgs...)
	if err != nil {
		return nil, err
	}

	models, err := queryFacetStringRows(ctx, s.clickhouse, buildTraceFacetStringQuery(baseFrom, whereSQL, "coalesce(t.model, '')"), sharedArgs...)
	if err != nil {
		return nil, err
	}

	routes, err := queryFacetStringRows(ctx, s.clickhouse, buildTraceFacetStringQuery(baseFrom, whereSQL, "coalesce(t.route, '')"), sharedArgs...)
	if err != nil {
		return nil, err
	}

	errorCodes, err := queryFacetStringRows(ctx, s.clickhouse, buildTraceFacetStringQuery(baseFrom, whereSQL, "coalesce(t.error_code, '')"), sharedArgs...)
	if err != nil {
		return nil, err
	}

	statuses, err := queryFacetStatusRows(ctx, s.clickhouse, buildTraceFacetStatusQuery(baseFrom, whereSQL), sharedArgs...)
	if err != nil {
		return nil, err
	}

	return &platform.TraceFacetsV2{
		Providers:   providers,
		Models:      models,
		Routes:      routes,
		ErrorCodes:  errorCodes,
		StatusCodes: statuses,
	}, nil
}

func (s *Store) getTraceUsageFromPostgres(ctx context.Context, tenantID, requestID string) (*platform.TraceUsageV2, error) {
	if strings.TrimSpace(requestID) == "" {
		return nil, nil
	}

	var (
		model          sql.NullString
		inputTokens    sql.NullInt64
		outputTokens   sql.NullInt64
		totalTokens    sql.NullInt64
		costUSD        sql.NullFloat64
		costIRR        sql.NullFloat64
		responseTimeMS sql.NullInt64
		status         sql.NullString
		metaRaw        []byte
	)

	err := s.pool.QueryRow(ctx, `
		SELECT
			model,
			input_tokens,
			output_tokens,
			total_tokens,
			cost_usd::float8,
			cost_irr::float8,
			response_time_ms,
			status,
			request_metadata
		FROM spend_logs
		WHERE tenant_id = NULLIF($1, '')::uuid AND request_id = $2
		ORDER BY completion_end_time DESC NULLS LAST, created_at DESC
		LIMIT 1
	`, tenantID, requestID).Scan(
		&model,
		&inputTokens,
		&outputTokens,
		&totalTokens,
		&costUSD,
		&costIRR,
		&responseTimeMS,
		&status,
		&metaRaw,
	)
	if err != nil {
		if errorsIsNoRows(err) {
			return nil, nil
		}

		return nil, err
	}

	item := &platform.TraceUsageV2{
		Model:          model.String,
		InputTokens:    int(inputTokens.Int64),
		OutputTokens:   int(outputTokens.Int64),
		TotalTokens:    int(totalTokens.Int64),
		CostUSD:        costUSD.Float64,
		CostIRR:        costIRR.Float64,
		ResponseTimeMS: int(responseTimeMS.Int64),
		Status:         status.String,
	}

	if len(metaRaw) > 0 {
		meta := map[string]interface{}{}
		if err := json.Unmarshal(metaRaw, &meta); err != nil {
			return nil, err
		}

		item.Metadata = meta
	}

	return item, nil
}

func (s *Store) getTraceMetricFromClickHouse(ctx context.Context, tenantID, requestID string) (*platform.TraceMetricV2, error) {
	if strings.TrimSpace(requestID) == "" {
		return nil, nil
	}

	var (
		requestType   sql.NullString
		providerName  sql.NullString
		providerModel sql.NullString
		errorType     sql.NullString
		latencyMS     uint32
		inputTokens   uint32
		outputTokens  uint32
		costUSD       float64
		ttftMS        sql.NullInt64
		itlMS         sql.NullInt64
		tpotMS        sql.NullInt64
		virtualModel  sql.NullString
	)

	err := s.clickhouse.QueryRowContext(ctx, `
		SELECT
			request_type,
			provider_name,
			provider_model_name,
			error_type,
			latency_ms,
			input_tokens,
			output_tokens,
			toFloat64(cost_usd) AS cost_usd,
			ttft_ms,
			itl_ms,
			tpot_ms,
			virtual_model_name
		FROM gateway_model_events
		WHERE tenant_id = ? AND request_id = ?
		ORDER BY ts DESC
		LIMIT 1
	`, tenantID, requestID).Scan(
		&requestType,
		&providerName,
		&providerModel,
		&errorType,
		&latencyMS,
		&inputTokens,
		&outputTokens,
		&costUSD,
		&ttftMS,
		&itlMS,
		&tpotMS,
		&virtualModel,
	)
	if err != nil {
		if errorsIsNoRows(err) {
			return nil, nil
		}

		return nil, err
	}

	item := &platform.TraceMetricV2{
		RequestType:      requestType.String,
		ProviderName:     providerName.String,
		ProviderModel:    providerModel.String,
		ErrorType:        errorType.String,
		LatencyMS:        int(latencyMS),
		InputTokens:      int(inputTokens),
		OutputTokens:     int(outputTokens),
		CostUSD:          costUSD,
		VirtualModelName: virtualModel.String,
	}

	if ttftMS.Valid {
		v := int(ttftMS.Int64)
		item.TTFTMS = &v
	}

	if itlMS.Valid {
		v := int(itlMS.Int64)
		item.ITLMS = &v
	}

	if tpotMS.Valid {
		v := int(tpotMS.Int64)
		item.TPOTMS = &v
	}

	return item, nil
}

func queryFacetStringRows(ctx context.Context, db *sql.DB, query string, args ...interface{}) ([]platform.TraceFacetBucketV2, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer closeBestEffort(rows)

	out := make([]platform.TraceFacetBucketV2, 0, 8)

	for rows.Next() {
		var (
			value sql.NullString
			count uint64
		)
		if err := rows.Scan(&value, &count); err != nil {
			return nil, err
		}

		out = append(out, platform.TraceFacetBucketV2{
			Value: strings.TrimSpace(value.String),
			Count: int(count),
		})
	}

	return out, rows.Err()
}

func queryFacetStatusRows(ctx context.Context, db *sql.DB, query string, args ...interface{}) ([]platform.TraceStatusFacetBucketV2, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer closeBestEffort(rows)

	out := make([]platform.TraceStatusFacetBucketV2, 0, 8)

	for rows.Next() {
		var (
			statusCode int32
			count      uint64
		)
		if err := rows.Scan(&statusCode, &count); err != nil {
			return nil, err
		}

		out = append(out, platform.TraceStatusFacetBucketV2{
			StatusCode: int(statusCode),
			Count:      int(count),
		})
	}

	return out, rows.Err()
}

func buildTraceV2BaseWhere(params platform.ListRequestTracesV2Params) (string, []interface{}) {
	where := []string{"tenant_id = ?"}

	args := []interface{}{params.TenantID}
	if params.TraceID != "" {
		where = append(where, "trace_id = ?")
		args = append(args, params.TraceID)
	}

	if params.RequestID != "" {
		where = append(where, "request_id = ?")
		args = append(args, params.RequestID)
	}

	if params.Route != "" {
		where = append(where, "route = ?")
		args = append(args, params.Route)
	}

	if params.Model != "" {
		where = append(where, "model = ?")
		args = append(args, params.Model)
	}

	if params.StatusCode != nil {
		where = append(where, "status_code = ?")
		args = append(args, int32(*params.StatusCode))
	}

	if params.ErrorCode != "" {
		where = append(where, "error_code = ?")
		args = append(args, params.ErrorCode)
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

func buildTraceProviderSpanWhere(params platform.ListRequestTracesV2Params) (string, []interface{}) {
	where := []string{"tenant_id = ?", "name = 'provider.call'"}
	args := []interface{}{params.TenantID}

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

func buildTraceFacetStringQuery(baseFrom, whereSQL, valueExpr string) string {
	return fmt.Sprintf(`
		SELECT %s AS value, count() AS count
		%s
		WHERE %s
		GROUP BY value
		ORDER BY count DESC
		LIMIT 20
	`, valueExpr, baseFrom, whereSQL)
}

func buildTraceFacetStatusQuery(baseFrom, whereSQL string) string {
	return fmt.Sprintf(`
		SELECT toInt32(t.status_code) AS status_code, count() AS count
		%s
		WHERE %s
		GROUP BY status_code
		ORDER BY count DESC
		LIMIT 20
	`, baseFrom, whereSQL)
}

func normalizeTraceSortBy(sortBy string) string {
	switch strings.ToLower(strings.TrimSpace(sortBy)) {
	case "request_id":
		return "t.request_id"
	case "duration_ms":
		return "t.duration_ms"
	case "status_code":
		return "t.status_code"
	default:
		return "t.started_at"
	}
}

func normalizeSortOrder(order string) string {
	if strings.EqualFold(strings.TrimSpace(order), "asc") {
		return "ASC"
	}

	return "DESC"
}

func boolToUint8(v bool) uint8 {
	if v {
		return 1
	}

	return 0
}

func isRootRequestSpan(ev tracing.SpanEvent) bool {
	return strings.EqualFold(strings.TrimSpace(ev.Name), "gateway.request") &&
		strings.EqualFold(strings.TrimSpace(ev.Kind), "server")
}

func buildTraceDedupeSubquery(whereSQL string) string {
	return fmt.Sprintf(`
		SELECT
			trace_id,
			tenant_id_agg AS tenant_id,
			request_id_agg AS request_id,
			model_agg AS model,
			route_agg AS route,
			status_code_agg AS status_code,
			started_at_agg AS started_at,
			duration_ms_agg AS duration_ms,
			error_code_agg AS error_code,
			sampled_agg AS sampled
		FROM (
			SELECT
				trace_id,
				argMin(tenant_id, started_at) AS tenant_id_agg,
				argMin(request_id, started_at) AS request_id_agg,
				argMin(model, started_at) AS model_agg,
				argMin(route, started_at) AS route_agg,
				argMin(status_code, started_at) AS status_code_agg,
				min(started_at) AS started_at_agg,
				argMin(duration_ms, started_at) AS duration_ms_agg,
				argMin(error_code, started_at) AS error_code_agg,
				argMin(sampled, started_at) AS sampled_agg
			FROM request_traces_ch
			WHERE %s
			GROUP BY trace_id
		)
	`, whereSQL)
}

func errorsIsNoRows(err error) bool {
	return err == sql.ErrNoRows || err == pgx.ErrNoRows
}
