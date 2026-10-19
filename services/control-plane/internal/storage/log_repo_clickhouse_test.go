package storage

import (
	"strings"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func TestBuildRequestLogsWhereIncludesEnterpriseFilters(t *testing.T) {
	statusCode := 429
	durationMin := 100
	durationMax := 2000
	tokensMin := 10
	tokensMax := 20000
	costUSDMin := 0.01
	costUSDMax := 2.5
	from := time.Date(2026, 7, 7, 8, 0, 0, 0, time.UTC)
	to := from.Add(time.Hour)

	whereSQL, args := buildRequestLogsWhere(platform.ListRequestLogsV2Params{
		TenantID:          "tenant-1",
		LogID:             "log-1",
		RequestID:         "req-1",
		TraceID:           "trace-1",
		Provider:          "openai",
		Model:             "gpt-4.1",
		VirtualModel:      "vm-prod",
		Route:             "/v1/chat/completions",
		RequestType:       "chat",
		ErrorType:         "rate_limit",
		CacheStatus:       "hit",
		FallbackStatus:    "active",
		LoadBalanceStatus: "weighted",
		PayloadMode:       "metadata_only",
		ActorID:           "user-1",
		TeamID:            "team-1",
		StatusCode:        &statusCode,
		DurationMin:       &durationMin,
		DurationMax:       &durationMax,
		TokensMin:         &tokensMin,
		TokensMax:         &tokensMax,
		CostUSDMin:        &costUSDMin,
		CostUSDMax:        &costUSDMax,
		From:              &from,
		To:                &to,
	})

	for _, clause := range []string{
		"tenant_id = ?",
		"log_id = ?",
		"request_id = ?",
		"trace_id = ?",
		"provider = ?",
		"virtual_model = ?",
		"load_balance_status = ?",
		"payload_mode = ?",
		"actor_id = ?",
		"team_id = ?",
		"duration_ms >= ?",
		"duration_ms <= ?",
		"total_tokens >= ?",
		"total_tokens <= ?",
		"cost_usd >= ?",
		"cost_usd <= ?",
		"started_at >= ?",
		"started_at <= ?",
	} {
		if !strings.Contains(whereSQL, clause) {
			t.Fatalf("expected where clause %q in %q", clause, whereSQL)
		}
	}

	if got, want := len(args), 25; got != want {
		t.Fatalf("unexpected arg count: got %d want %d", got, want)
	}
}
