package storage

import (
	"strings"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/stretchr/testify/assert"
)

func TestIsRootRequestSpan(t *testing.T) {
	t.Run("matches gateway root", func(t *testing.T) {
		ev := tracing.SpanEvent{Name: "gateway.request", Kind: "server"}
		assert.True(t, isRootRequestSpan(ev))
	})

	t.Run("case-insensitive match", func(t *testing.T) {
		ev := tracing.SpanEvent{Name: "Gateway.Request", Kind: "SERVER"}
		assert.True(t, isRootRequestSpan(ev))
	})

	t.Run("rejects non-root span", func(t *testing.T) {
		ev := tracing.SpanEvent{Name: "provider.call", Kind: "client"}
		assert.False(t, isRootRequestSpan(ev))
	})
}

func TestBuildTraceDedupeSubquery(t *testing.T) {
	sql := buildTraceDedupeSubquery("tenant_id = ? AND route = ?")
	normalized := strings.ToLower(strings.Join(strings.Fields(sql), " "))

	assert.Contains(t, normalized, "from request_traces_ch")
	assert.Contains(t, normalized, "where tenant_id = ? and route = ?")
	assert.Contains(t, normalized, "group by trace_id")
	assert.Contains(t, normalized, "argmin(tenant_id, started_at) as tenant_id_agg")
	assert.Contains(t, normalized, "min(started_at) as started_at_agg")
	assert.Contains(t, normalized, "tenant_id_agg as tenant_id")
}

func TestBuildTraceV2BaseWhere(t *testing.T) {
	status := 500
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	where, args := buildTraceV2BaseWhere(platform.ListRequestTracesV2Params{
		TenantID:   "tenant_a",
		TraceID:    "trace_1",
		RequestID:  "req_1",
		Route:      "/v1/chat/completions",
		Model:      "gpt-4o-mini",
		ErrorCode:  "provider_error",
		StatusCode: &status,
		From:       &from,
		To:         &to,
	})
	normalized := strings.ToLower(strings.Join(strings.Fields(where), " "))
	assert.Contains(t, normalized, "tenant_id = ?")
	assert.Contains(t, normalized, "startswith(route, '/v1/')")
	assert.Contains(t, normalized, "trace_id = ?")
	assert.Contains(t, normalized, "request_id = ?")
	assert.Contains(t, normalized, "route = ?")
	assert.Contains(t, normalized, "model = ?")
	assert.Contains(t, normalized, "status_code = ?")
	assert.Contains(t, normalized, "error_code = ?")
	assert.Contains(t, normalized, "started_at >= ?")
	assert.Contains(t, normalized, "started_at <= ?")
	assert.Len(t, args, 9)
}

func TestBuildTraceFacetQueries(t *testing.T) {
	baseFrom := "FROM (SELECT 1 AS trace_id) t LEFT JOIN (SELECT 1 AS trace_id) p ON t.trace_id = p.trace_id"
	whereSQL := "1 = 1"
	stringQ := buildTraceFacetStringQuery(baseFrom, whereSQL, "coalesce(p.provider, '')")
	stringNorm := strings.ToLower(strings.Join(strings.Fields(stringQ), " "))
	assert.Contains(t, stringNorm, "select coalesce(p.provider, '') as value, count() as count")
	assert.Contains(t, stringNorm, "group by value")
	assert.Contains(t, stringNorm, "order by count desc")
	assert.Contains(t, stringNorm, "limit 20")

	statusQ := buildTraceFacetStatusQuery(baseFrom, whereSQL)
	statusNorm := strings.ToLower(strings.Join(strings.Fields(statusQ), " "))
	assert.Contains(t, statusNorm, "select toint32(t.status_code) as status_code, count() as count")
	assert.Contains(t, statusNorm, "group by status_code")
	assert.Contains(t, statusNorm, "order by count desc")
	assert.Contains(t, statusNorm, "limit 20")
}
