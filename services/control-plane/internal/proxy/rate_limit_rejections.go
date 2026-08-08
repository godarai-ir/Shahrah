package proxy

import (
	"errors"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	metricsobs "github.com/Delpak/shahrah/services/control-plane/internal/observability/metrics"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func recordRateLimitRejection(
	c *fiber.Ctx,
	logger *zap.Logger,
	logs usageLogger,
	emitMetric func(metricsobs.BuildInput),
	authInfo *platform.AuthContext,
	requestedModel string,
	route string,
	requestType string,
	startedAt time.Time,
	quotaStartedAt time.Time,
	result *cache.RateLimitResult,
	err error,
) bool {
	statusCode := rateLimitErrorStatus(err)
	if statusCode != fiber.StatusTooManyRequests || !errors.Is(err, billing.ErrRateLimitExceeded) {
		return false
	}

	requestID := tracing.TraceIDFromContext(c)
	durationMS := int(time.Since(startedAt).Milliseconds())
	metadata := map[string]interface{}{"denial_source": "gateway_rate_limit"}

	if result != nil {
		remaining := result.Limit - result.Current
		if remaining < 0 {
			remaining = 0
		}

		metadata["limit"] = result.Limit
		metadata["current"] = result.Current
		metadata["remaining"] = remaining

		metadata["reset_at"] = result.ResetAt.UTC().Format(time.RFC3339Nano)
		if result.RuleID != "" {
			metadata["rule_id"] = result.RuleID
		}

		if result.Dimension != "" {
			metadata["dimension"] = result.Dimension
		}

		if result.Algorithm != "" {
			metadata["algorithm"] = result.Algorithm
		}
	}

	entry := buildRequestLogEntry(c, authInfo, platform.Deployment{}, requestedModel, route, requestType, "blocked", statusCode, startedAt, durationMS)
	entry.RequestID = requestID
	entry.TraceID = requestID
	entry.ErrorType = "rate_limit_error"
	entry.Metadata = metadata

	entry.ResponseBody = map[string]interface{}{"error": map[string]interface{}{"type": "rate_limit_error", "message": err.Error()}}
	if logs != nil {
		logRequestBestEffort(logger, logs, c.Context(), entry)
	}

	if emitMetric != nil {
		emitMetric(metricsobs.BuildInput{
			TS:                 time.Now().UTC(),
			TenantID:           authInfo.TenantID,
			RequestID:          requestID,
			ModelName:          requestedModel,
			RequestType:        requestType,
			UserID:             authInfo.UserID,
			VirtualAccountID:   authInfo.KeyID,
			VirtualAccountName: authInfo.KeyAlias,
			TeamIDs:            nonEmptySlice([]string{authInfo.TeamID}),
			StatusCode:         statusCode,
			ErrorType:          "rate_limit_error",
			LatencyMS:          durationMS,
			Metadata:           metadata,
		})
	}

	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "quota.check", "internal", quotaStartedAt, statusCode, metadata)

	return true
}
