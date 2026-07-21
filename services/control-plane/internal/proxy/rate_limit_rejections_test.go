package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	metricsobs "github.com/Delpak/shahrah/services/control-plane/internal/observability/metrics"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestRecordRateLimitRejectionWritesProductLogMetricAndSpan(t *testing.T) {
	app := fiber.New()
	queue := tracing.NewQueue(10)
	app.Use(tracing.Middleware(queue, tracing.NewSampler(1, 1000), 1000))

	logs := &testLogs{}

	var metric metricsobs.BuildInput

	app.Post("/v1/chat/completions", func(c *fiber.Ctx) error {
		authInfo := &platform.AuthContext{
			TenantID: "tenant-1",
			UserID:   "user-1",
			TeamID:   "team-1",
			KeyID:    "key-1",
		}
		startedAt := time.Now().Add(-20 * time.Millisecond)
		quotaStartedAt := time.Now().Add(-5 * time.Millisecond)
		recordRateLimitRejection(
			c,
			zap.NewNop(),
			logs,
			func(input metricsobs.BuildInput) { metric = input },
			authInfo,
			"gpt-4o-mini",
			"/v1/chat/completions",
			"chat_completions",
			startedAt,
			quotaStartedAt,
			&cache.RateLimitResult{
				Allowed:   false,
				Current:   10,
				Limit:     10,
				ResetAt:   time.Now().Add(time.Minute),
				RuleID:    "tenant-rpm",
				Dimension: "rpm",
				Algorithm: "fixed_window",
			},
			billing.ErrRateLimitExceeded,
		)

		return writeAPIError(c, fiber.StatusTooManyRequests, billing.ErrRateLimitExceeded.Error(), "rate_limit_error")
	})

	request := httptest.NewRequest(fiber.MethodPost, "/v1/chat/completions", http.NoBody)
	response, err := app.Test(request)

	require.NoError(t, err)
	defer closeResponseBody(t, response)

	require.Equal(t, fiber.StatusTooManyRequests, response.StatusCode)

	require.Len(t, logs.loggedEntries, 1)
	entry := logs.loggedEntries[0]
	assert.NotEmpty(t, entry.RequestID)
	assert.Equal(t, entry.RequestID, entry.TraceID)
	assert.Equal(t, "blocked", entry.Status)
	assert.Equal(t, fiber.StatusTooManyRequests, entry.StatusCode)
	assert.Equal(t, "rate_limit_error", entry.ErrorType)
	assert.Zero(t, entry.CostUSD)
	assert.Zero(t, entry.CostIRR)
	assert.Equal(t, "tenant-rpm", entry.Metadata["rule_id"])
	assert.Equal(t, "rpm", entry.Metadata["dimension"])
	assert.Empty(t, logs.loggedSpends)

	assert.Equal(t, entry.RequestID, metric.RequestID)
	assert.Equal(t, fiber.StatusTooManyRequests, metric.StatusCode)
	assert.Equal(t, "rate_limit_error", metric.ErrorType)
	assert.Equal(t, "tenant-rpm", metric.Metadata["rule_id"])

	spans := queue.DequeueBatch(10)
	require.NotEmpty(t, spans)
	assert.Condition(t, func() bool {
		for _, span := range spans {
			if span.Name == "quota.check" && span.StatusCode == fiber.StatusTooManyRequests {
				return true
			}
		}

		return false
	}, "expected a 429 quota.check span")
}
