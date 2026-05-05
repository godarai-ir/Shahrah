package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/jackc/pgx/v5"
)

// Export persists a batch of trace span events.
func (s *Store) Export(ctx context.Context, batch []tracing.SpanEvent) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, ev := range batch {
		if strings.TrimSpace(ev.TraceID) == "" {
			continue
		}
		var errorCode interface{}
		if ev.StatusCode >= 500 {
			errorCode = "upstream_error"
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO request_traces (
				trace_id, tenant_id, request_id, route, status_code, started_at, duration_ms, error_code, sampled
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (trace_id) DO UPDATE
			SET tenant_id = EXCLUDED.tenant_id,
			    request_id = EXCLUDED.request_id,
			    route = EXCLUDED.route,
			    status_code = EXCLUDED.status_code,
			    started_at = EXCLUDED.started_at,
			    duration_ms = EXCLUDED.duration_ms,
			    error_code = EXCLUDED.error_code,
			    sampled = EXCLUDED.sampled
		`, ev.TraceID, nullUUID(ev.TenantID), ev.TraceID, ev.Route, ev.StatusCode, ev.StartedAt, int(ev.DurationMS), errorCode, ev.Sampled)
		if err != nil {
			return err
		}

		attrsRaw, _ := json.Marshal(map[string]interface{}{
			"http.method": ev.Method,
			"http.route":  ev.Route,
			"sampled":     ev.Sampled,
		})
		spanID := fmt.Sprintf("%s:%d", ev.TraceID, ev.StartedAt.UnixNano())
		status := "ok"
		if ev.StatusCode >= 500 {
			status = "error"
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO request_spans (
				span_id, trace_id, parent_span_id, name, kind, started_at, duration_ms, status, attributes_json
			)
			VALUES ($1, $2, NULL, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (span_id) DO NOTHING
		`, spanID, ev.TraceID, "gateway.request", "server", ev.StartedAt, int(ev.DurationMS), status, attrsRaw)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (s *Store) ListRequestTraces(ctx context.Context, params platform.ListRequestTracesParams) ([]platform.RequestTrace, int, error) {
	if params.Limit <= 0 {
		params.Limit = 50
	}
	if params.Limit > 200 {
		params.Limit = 200
	}
	if params.Offset < 0 {
		params.Offset = 0
	}

	args := make([]interface{}, 0, 10)
	where := []string{"tenant_id = $1::uuid"}
	args = append(args, params.TenantID)
	argPos := 2

	if params.TraceID != "" {
		where = append(where, fmt.Sprintf("trace_id = $%d", argPos))
		args = append(args, params.TraceID)
		argPos++
	}
	if params.Route != "" {
		where = append(where, fmt.Sprintf("route = $%d", argPos))
		args = append(args, params.Route)
		argPos++
	}
	if params.Model != "" {
		where = append(where, fmt.Sprintf("model = $%d", argPos))
		args = append(args, params.Model)
		argPos++
	}
	if params.StatusCode != nil {
		where = append(where, fmt.Sprintf("status_code = $%d", argPos))
		args = append(args, *params.StatusCode)
		argPos++
	}
	if params.From != nil {
		where = append(where, fmt.Sprintf("started_at >= $%d", argPos))
		args = append(args, *params.From)
		argPos++
	}
	if params.To != nil {
		where = append(where, fmt.Sprintf("started_at <= $%d", argPos))
		args = append(args, *params.To)
		argPos++
	}
	whereSQL := strings.Join(where, " AND ")

	var total int
	countSQL := fmt.Sprintf(`SELECT COUNT(*) FROM request_traces WHERE %s`, whereSQL)
	if err := s.pool.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listSQL := fmt.Sprintf(`
		SELECT trace_id, COALESCE(tenant_id::text, ''), COALESCE(request_id, ''), COALESCE(model, ''), route, status_code,
		       started_at, duration_ms, COALESCE(error_code, ''), sampled
		FROM request_traces
		WHERE %s
		ORDER BY started_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argPos, argPos+1)
	args = append(args, params.Limit, params.Offset)

	rows, err := s.pool.Query(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]platform.RequestTrace, 0, params.Limit)
	for rows.Next() {
		var item platform.RequestTrace
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
			&item.Sampled,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (s *Store) GetRequestTrace(ctx context.Context, tenantID string, traceID string) (*platform.RequestTrace, error) {
	var item platform.RequestTrace
	err := s.pool.QueryRow(ctx, `
		SELECT trace_id, COALESCE(tenant_id::text, ''), COALESCE(request_id, ''), COALESCE(model, ''), route, status_code,
		       started_at, duration_ms, COALESCE(error_code, ''), sampled
		FROM request_traces
		WHERE tenant_id = $1::uuid AND trace_id = $2
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
		&item.Sampled,
	)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *Store) ListTraceSpans(ctx context.Context, tenantID string, traceID string) ([]platform.TraceSpan, error) {
	if _, err := s.GetRequestTrace(ctx, tenantID, traceID); err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `
		SELECT span_id, trace_id, COALESCE(parent_span_id, ''), name, kind, started_at, duration_ms, status, attributes_json
		FROM request_spans
		WHERE trace_id = $1
		ORDER BY started_at ASC
	`, traceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]platform.TraceSpan, 0, 8)
	for rows.Next() {
		var item platform.TraceSpan
		var raw []byte
		if err := rows.Scan(
			&item.SpanID,
			&item.TraceID,
			&item.ParentSpanID,
			&item.Name,
			&item.Kind,
			&item.StartedAt,
			&item.DurationMS,
			&item.Status,
			&raw,
		); err != nil {
			return nil, err
		}
		if len(raw) > 0 {
			_ = json.Unmarshal(raw, &item.Attributes)
		}
		if item.Attributes == nil {
			item.Attributes = map[string]interface{}{}
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func parseRFC3339Time(raw string) (*time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	v, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func isNotFound(err error) bool {
	return err == pgx.ErrNoRows
}
