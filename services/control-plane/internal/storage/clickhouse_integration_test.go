package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestClickHouseObservabilityIntegration(t *testing.T) {
	if os.Getenv("SHAHRAH_CLICKHOUSE_INTEGRATION") != "1" {
		t.Skip("set SHAHRAH_CLICKHOUSE_INTEGRATION=1 to run ClickHouse container integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	dsn, cleanup := startClickHouseTestContainer(t, ctx)
	defer cleanup()

	db, err := sql.Open("clickhouse", dsn)
	require.NoError(t, err)

	defer closeBestEffort(db)

	require.NoError(t, waitForClickHouse(ctx, db))
	require.NoError(t, applyClickHouseMigrations(ctx, db))

	store := composeObservabilityRepository(&observabilityState{
		clickhouse:             db,
		clickhouseReady:        true,
		clickhouseTraceReady:   true,
		clickhouseMetricsReady: true,
		clickhouseLogsReady:    true,
		logger:                 zap.NewNop(),
	})

	require.NoError(t, store.ensureClickHouseMetricsSchema(ctx))
	require.NoError(t, store.ensureClickHouseTraceSchema(ctx))
	require.NoError(t, store.ensureClickHouseRequestLogsSchema(ctx))

	start := time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC)
	tenantID := "tenant-integration"

	insertClickHouseMetrics(t, ctx, store, tenantID, start)
	insertClickHouseLogs(t, ctx, db, tenantID, start)
	insertClickHouseTraces(t, ctx, db, tenantID, start)

	assertClickHouseMetricsQueries(t, ctx, store, tenantID, start)
	assertClickHouseLogQueries(t, ctx, store, tenantID, start)
	assertClickHouseTraceQueries(t, ctx, store, tenantID, start)
	assertClickHouseTraceRetentionTTL(t, ctx, db)
}

func startClickHouseTestContainer(t *testing.T, ctx context.Context) (string, func()) {
	t.Helper()

	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker CLI is required for ClickHouse integration test")
	}

	name := fmt.Sprintf("shahrah-clickhouse-test-%d", time.Now().UnixNano())
	args := []string{
		"run", "-d", "--rm",
		"--name", name,
		"-p", "127.0.0.1::9000",
		"-e", "CLICKHOUSE_SKIP_USER_SETUP=1",
		"clickhouse/clickhouse-server:24.8-alpine",
	}

	out, err := exec.CommandContext(ctx, "docker", args...).CombinedOutput()
	if err != nil {
		t.Fatalf("start clickhouse container: %v\n%s", err, string(out))
	}

	cleanup := func() {
		_ = exec.Command("docker", "rm", "-f", name).Run()
	}

	t.Cleanup(cleanup)

	portOut, err := exec.CommandContext(ctx, "docker", "port", name, "9000/tcp").CombinedOutput()
	if err != nil {
		t.Fatalf("inspect clickhouse port: %v\n%s", err, string(portOut))
	}

	port := parseDockerHostPort(t, string(portOut))

	return "clickhouse://127.0.0.1:" + port, cleanup
}

func parseDockerHostPort(t *testing.T, raw string) string {
	t.Helper()

	fields := strings.Fields(raw)
	require.NotEmpty(t, fields)

	endpoint := fields[len(fields)-1]
	idx := strings.LastIndex(endpoint, ":")
	require.NotEqual(t, -1, idx, "unexpected docker port output: %q", raw)

	port := strings.Trim(endpoint[idx+1:], "[] \r\n")
	_, err := strconv.Atoi(port)
	require.NoError(t, err, "unexpected docker port output: %q", raw)

	return port
}

func waitForClickHouse(ctx context.Context, db *sql.DB) error {
	deadline := time.Now().Add(90 * time.Second)

	var lastErr error

	for time.Now().Before(deadline) {
		err := db.PingContext(ctx)
		if err == nil {
			return nil
		}

		lastErr = err

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}

	return fmt.Errorf("clickhouse did not become ready: %w", lastErr)
}

func applyClickHouseMigrations(ctx context.Context, db *sql.DB) error {
	root, err := findRepoRoot()
	if err != nil {
		return err
	}

	for _, name := range []string{
		"001_gateway_model_events.sql",
		"002_gateway_model_metrics_1m.sql",
		"003_mv_gateway_model_metrics_1m.sql",
		"004_observability_runtime_schema.sql",
		"005_gateway_model_event_id.sql",
	} {
		path := filepath.Join(root, "infra", "clickhouse", "migrations", name)

		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		for _, stmt := range splitSQLStatements(string(raw)) {
			if _, err := db.ExecContext(ctx, stmt); err != nil {
				return fmt.Errorf("apply %s: %w", name, err)
			}
		}
	}

	return nil
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "infra", "clickhouse", "migrations")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("repo root not found from %s", dir)
		}

		dir = parent
	}
}

func splitSQLStatements(raw string) []string {
	parts := strings.Split(raw, ";")

	out := make([]string, 0, len(parts))
	for _, part := range parts {
		stmt := strings.TrimSpace(part)
		if stmt != "" {
			out = append(out, stmt)
		}
	}

	return out
}

func insertClickHouseMetrics(t *testing.T, ctx context.Context, store *ObservabilityRepository, tenantID string, start time.Time) {
	t.Helper()

	err := store.InsertModelMetricEvents(ctx, []platform.ModelMetricEvent{
		{
			TS:                start.Add(5 * time.Second),
			TenantID:          tenantID,
			RequestID:         "metric-request-1",
			ModelName:         "gpt-4.1",
			VirtualModelName:  "support-bot",
			ProviderName:      "openai",
			ProviderModelName: "gpt-4.1",
			RequestType:       "chat_completions",
			TeamIDs:           []string{"team-platform"},
			TeamNames:         []string{"platform"},
			StatusCode:        200,
			ErrorType:         "none",
			LatencyMS:         120,
			InputTokens:       10,
			OutputTokens:      20,
			CostUSD:           0.03,
			CacheMode:         "none",
			CacheStatus:       "none",
			CacheMatch:        "none",
			PricingVersion:    "gateway-pricing-v1",
			Metadata:          map[string]interface{}{"environment": "integration"},
		},
		{
			TS:                start.Add(20 * time.Second),
			TenantID:          tenantID,
			RequestID:         "metric-request-2",
			ModelName:         "gpt-4.1",
			VirtualModelName:  "support-bot",
			ProviderName:      "openai",
			ProviderModelName: "gpt-4.1",
			RequestType:       "chat_completions",
			TeamIDs:           []string{"team-growth"},
			TeamNames:         []string{"growth"},
			StatusCode:        500,
			ErrorType:         "provider_error",
			LatencyMS:         250,
			InputTokens:       15,
			OutputTokens:      25,
			CostUSD:           0.05,
			CacheMode:         "none",
			CacheStatus:       "none",
			CacheMatch:        "none",
			PricingVersion:    "gateway-pricing-v1",
			Metadata:          map[string]interface{}{"environment": "integration"},
		},
	})
	require.NoError(t, err)
}

func insertClickHouseLogs(t *testing.T, ctx context.Context, db *sql.DB, tenantID string, start time.Time) {
	t.Helper()

	_, err := db.ExecContext(ctx, `
		INSERT INTO request_logs_ch (
			log_id, tenant_id, request_id, trace_id, token_hash, organization_id, org_unit_id, team_id,
			actor_type, actor_id, provider, provider_model, model, virtual_model, route, request_type,
			status, status_code, error_type, started_at, duration_ms, input_tokens, output_tokens, total_tokens,
			cost_usd, cost_irr, cache_status, cache_mode, retry_count, fallback_count, fallback_status,
			load_balance_status, payload_mode, request_payload_json, response_payload_json,
			request_payload_meta_json, response_payload_meta_json, request_payload_truncated,
			response_payload_truncated, metadata_json
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		"log-integration-1", tenantID, "log-request-1", "trace-integration-1", "token", "org-1", "org-unit-1", "team-platform",
		"user", "user-1", "openai", "gpt-4.1", "gpt-4.1", "support-bot", "/v1/chat/completions", "chat_completions",
		"ok", int32(200), "none", start.Add(10*time.Second), int32(130), uint32(10), uint32(20), uint32(30),
		0.03, 0.0, "none", "none", uint16(0), uint16(0), "none",
		"none", "metadata_only", "{}", "{}", "{}", "{}", uint8(0), uint8(0), `{"environment":"integration"}`,
	)
	require.NoError(t, err)
}

func insertClickHouseTraces(t *testing.T, ctx context.Context, db *sql.DB, tenantID string, start time.Time) {
	t.Helper()

	_, err := db.ExecContext(ctx, `
		INSERT INTO request_traces_ch
		(trace_id, tenant_id, request_id, model, route, status_code, started_at, duration_ms, error_code, sampled, retention_days)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		"trace-integration-1", tenantID, "trace-request-1", "gpt-4.1", "/v1/chat/completions",
		int32(200), start.Add(15*time.Second), int32(150), "", uint8(1), uint16(7),
	)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO request_spans_ch
		(span_id, trace_id, tenant_id, parent_span_id, name, kind, started_at, duration_ms, status, attributes_json, retention_days)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		"span-provider-1", "trace-integration-1", tenantID, "", "provider.call", "client",
		start.Add(16*time.Second), int32(120), "ok", `{"provider":"openai","model":"gpt-4.1"}`, uint16(7),
	)
	require.NoError(t, err)
}

func assertClickHouseMetricsQueries(t *testing.T, ctx context.Context, store *ObservabilityRepository, tenantID string, start time.Time) {
	t.Helper()

	aggregateRows, err := store.QueryModelMetrics(ctx, tenantID, platform.MetricsQueryRequest{
		StartTS:           start,
		EndTS:             start.Add(time.Minute),
		Type:              "timeseries",
		GroupBy:           []string{"modelName"},
		IntervalInSeconds: 60,
		Aggregations:      []platform.MetricsAggregation{{Type: "count", Column: "modelName"}},
	})
	require.NoError(t, err)
	require.Len(t, aggregateRows, 1)
	require.Equal(t, uint64(2), numericMapValue(t, aggregateRows[0], "countModelName"))

	arrayFilterRows, err := store.QueryModelMetrics(ctx, tenantID, platform.MetricsQueryRequest{
		StartTS:      start,
		EndTS:        start.Add(time.Minute),
		Type:         "distribution",
		Aggregations: []platform.MetricsAggregation{{Type: "count", Column: "modelName"}},
		Filters: []platform.MetricsFilter{
			{FieldName: "team", Operator: "EQUAL", Value: "platform"},
		},
	})
	require.NoError(t, err)
	require.Len(t, arrayFilterRows, 1)
	require.Equal(t, uint64(1), numericMapValue(t, arrayFilterRows[0], "countModelName"))
}

func assertClickHouseLogQueries(t *testing.T, ctx context.Context, store *ObservabilityRepository, tenantID string, start time.Time) {
	t.Helper()

	from := start
	to := start.Add(time.Minute)
	logs, total, err := store.ListRequestLogsV2(ctx, platform.ListRequestLogsV2Params{
		TenantID: tenantID,
		From:     &from,
		To:       &to,
		TeamID:   "team-platform",
		Limit:    10,
		SortBy:   "started_at",
	})
	require.NoError(t, err)
	require.Equal(t, 1, total)
	require.Len(t, logs, 1)
	require.Equal(t, "log-integration-1", logs[0].LogID)
	require.Equal(t, "openai", logs[0].Provider)
}

func assertClickHouseTraceQueries(t *testing.T, ctx context.Context, store *ObservabilityRepository, tenantID string, start time.Time) {
	t.Helper()

	from := start
	to := start.Add(time.Minute)
	traces, total, err := store.ListRequestTracesV2(ctx, platform.ListRequestTracesV2Params{
		TenantID: tenantID,
		From:     &from,
		To:       &to,
		Provider: "openai",
		Limit:    10,
		SortBy:   "started_at",
	})
	require.NoError(t, err)
	require.Equal(t, 1, total)
	require.Len(t, traces, 1)
	require.Equal(t, "trace-integration-1", traces[0].TraceID)
	require.Equal(t, "openai", traces[0].Provider)
	require.Equal(t, 1, traces[0].Attempts)
}

func assertClickHouseTraceRetentionTTL(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()

	for _, table := range []string{"request_traces_ch", "request_spans_ch"} {
		var createSQL string

		err := db.QueryRowContext(ctx, "SHOW CREATE TABLE "+table).Scan(&createSQL)
		require.NoError(t, err)
		require.Contains(t, createSQL, "retention_days")
		require.Contains(t, createSQL, "TTL")
	}
}

func numericMapValue(t *testing.T, row map[string]interface{}, key string) uint64 {
	t.Helper()

	value, ok := row[key]
	require.True(t, ok, "missing key %s in row %#v", key, row)

	switch v := value.(type) {
	case uint64:
		return v
	case uint32:
		return uint64(v)
	case uint16:
		return uint64(v)
	case uint8:
		return uint64(v)
	case int64:
		return uint64(v)
	case int:
		return uint64(v)
	default:
		t.Fatalf("unexpected numeric value type %T for %s", value, key)
		return 0
	}
}
