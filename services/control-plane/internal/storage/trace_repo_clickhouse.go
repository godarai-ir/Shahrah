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
)

func (s *Store) initClickHouse(ctx context.Context) error {
	dsn := strings.TrimSpace(os.Getenv("CLICKHOUSE_DSN"))
	if dsn == "" {
		return fmt.Errorf("CLICKHOUSE_DSN is required for trace storage")
	}
	db, err := sql.Open("clickhouse", dsn)
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(16)
	db.SetMaxIdleConns(8)
	db.SetConnMaxLifetime(15 * time.Minute)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return err
	}
	s.clickhouse = db
	if err := s.ensureClickHouseTraceSchema(ctx); err != nil {
		_ = db.Close()
		s.clickhouse = nil
		return err
	}
	if err := s.ensureClickHouseMetricsSchema(ctx); err != nil {
		_ = db.Close()
		s.clickhouse = nil
		return err
	}
	s.clickhouseReady = true
	s.logger.Info("clickhouse trace backend enabled")
	return nil
}

func (s *Store) ensureClickHouseTraceSchema(ctx context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS request_traces_ch (
			trace_id String,
			tenant_id String,
			request_id String,
			model String,
			route String,
			status_code Int32,
			started_at DateTime64(3, 'UTC'),
			duration_ms Int32,
			error_code String,
			sampled UInt8
		) ENGINE = MergeTree
		PARTITION BY toDate(started_at)
		ORDER BY (tenant_id, started_at, trace_id)
		TTL toDateTime(started_at) + INTERVAL 7 DAY`,
		`CREATE TABLE IF NOT EXISTS request_spans_ch (
			span_id String,
			trace_id String,
			tenant_id String,
			parent_span_id String,
			name String,
			kind String,
			started_at DateTime64(3, 'UTC'),
			duration_ms Int32,
			status String,
			attributes_json String
		) ENGINE = MergeTree
		PARTITION BY toDate(started_at)
		ORDER BY (tenant_id, trace_id, started_at, span_id)
		TTL toDateTime(started_at) + INTERVAL 7 DAY`,
	}
	for _, stmt := range stmts {
		if _, err := s.clickhouse.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) exportToClickHouse(ctx context.Context, batch []tracing.SpanEvent) error {
	if len(batch) == 0 {
		return nil
	}
	const traceInsertSQL = `
		INSERT INTO request_traces_ch
		(trace_id, tenant_id, request_id, model, route, status_code, started_at, duration_ms, error_code, sampled)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	const spanInsertSQL = `
		INSERT INTO request_spans_ch
		(span_id, trace_id, tenant_id, parent_span_id, name, kind, started_at, duration_ms, status, attributes_json)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	for _, ev := range batch {
		if strings.TrimSpace(ev.TraceID) == "" || strings.TrimSpace(ev.TenantID) == "" {
			continue
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
		requestID := ev.TraceID
		_, err := s.clickhouse.ExecContext(
			ctx,
			traceInsertSQL,
			ev.TraceID,
			ev.TenantID,
			requestID,
			model,
			ev.Route,
			int32(ev.StatusCode),
			ev.StartedAt.UTC(),
			int32(ev.DurationMS),
			errorCode,
			boolToUint8(ev.Sampled),
		)
		if err != nil {
			return err
		}

		attrs := map[string]interface{}{
			"http.method": ev.Method,
			"http.route":  ev.Route,
			"sampled":     ev.Sampled,
		}
		for k, v := range ev.Attributes {
			attrs[k] = v
		}
		attrsRaw, _ := json.Marshal(attrs)
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
		_, err = s.clickhouse.ExecContext(
			ctx,
			spanInsertSQL,
			spanID,
			ev.TraceID,
			ev.TenantID,
			parentSpanID,
			spanName,
			spanKind,
			ev.StartedAt.UTC(),
			int32(ev.DurationMS),
			status,
			string(attrsRaw),
		)
		if err != nil {
			return err
		}
	}
	return nil
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
	countQ := fmt.Sprintf(`SELECT count() FROM request_traces_ch WHERE %s`, whereSQL)
	var total uint64
	if err := s.clickhouse.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT trace_id, tenant_id, request_id, model, route, status_code, started_at, duration_ms, error_code, sampled
		FROM request_traces_ch
		WHERE %s
		ORDER BY started_at DESC
		LIMIT ? OFFSET ?
	`, whereSQL)
	args = append(args, params.Limit, params.Offset)
	rows, err := s.clickhouse.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]platform.RequestTrace, 0, params.Limit)
	for rows.Next() {
		var item platform.RequestTrace
		var sampled uint8
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
	return items, int(total), rows.Err()
}

func (s *Store) getRequestTraceFromClickHouse(ctx context.Context, tenantID, traceID string) (*platform.RequestTrace, error) {
	var item platform.RequestTrace
	var sampled uint8
	err := s.clickhouse.QueryRowContext(ctx, `
		SELECT trace_id, tenant_id, request_id, model, route, status_code, started_at, duration_ms, error_code, sampled
		FROM request_traces_ch
		WHERE tenant_id = ? AND trace_id = ?
		ORDER BY started_at DESC
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
	defer rows.Close()

	out := make([]platform.TraceSpan, 0, 8)
	for rows.Next() {
		var item platform.TraceSpan
		var attrsRaw string
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
			_ = json.Unmarshal([]byte(attrsRaw), &item.Attributes)
		}
		if item.Attributes == nil {
			item.Attributes = map[string]interface{}{}
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func boolToUint8(v bool) uint8 {
	if v {
		return 1
	}
	return 0
}

func errorsIsNoRows(err error) bool {
	return err == sql.ErrNoRows
}
