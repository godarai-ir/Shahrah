// Package proxy hosts OpenAI-compatible request handlers for the control plane.
package proxy

import (
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	metricsobs "github.com/Delpak/shahrah/services/control-plane/internal/observability/metrics"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// tokenBazaarSyncConfig carries the pieces that differ between
// TokenBazaarMessagesHandler.handleSync and
// TokenBazaarGenerateContentHandler.handleSync -- everything else in their
// dispatch/settle/log/metric tail is identical and lives in
// tokenBazaarHandleSync below.
type tokenBazaarSyncConfig struct {
	parseUsage              func(body []byte) *contractsopenai.Usage
	writeError              func(c *fiber.Ctx, status int, code, message string) error
	tunnelErrorCode         string
	tunnelErrorMessage      string
	writeUsageFinalizeError func(c *fiber.Ctx, err error) error
	settleErrorCode         string
	logPath                 string
	requestType             string
	metricEndpoint          string
}

// tokenBazaarHandleSync implements the synchronous (non-streaming) dispatch,
// billing settlement, request logging, and metric emission shared by
// TokenBazaarMessagesHandler and TokenBazaarGenerateContentHandler. The two
// callers differ only in usage parsing, error codes/messages, and a handful
// of endpoint/request-type labels, all supplied via cfg.
func tokenBazaarHandleSync(
	c *fiber.Ctx,
	pipeline *RequestPipeline,
	billing budgeter,
	logs usageLogger,
	logger *zap.Logger,
	emitMetric func(input metricsobs.BuildInput),
	cfg tokenBazaarSyncConfig,
	requestedModel string,
	originalBody []byte,
	authInfo *platform.AuthContext,
	estimated platform.UsageCostBreakdown,
	tunnelReq platform.TunnelRequest,
	reservation *platform.BudgetReservation,
	start time.Time,
	tenantSlug string,
	requestLog *requestLogLifecycle,
) error {
	response, err := pipeline.Dispatch(c.Context(), tunnelReq)
	if err != nil {
		pipeline.Release(c.Context(), reservation, "tunnel_error")

		return cfg.writeError(c, fiber.StatusBadGateway, cfg.tunnelErrorCode, cfg.tunnelErrorMessage)
	}

	if response.StatusCode >= fiber.StatusBadRequest {
		pipeline.Release(c.Context(), reservation, "provider_http_error")
		forwardProviderHeaders(c, response.Headers)

		return c.Status(response.StatusCode).Send(response.Body)
	}

	usage := cfg.parseUsage(response.Body)

	finalized, finalizeErr := billing.FinalizeFromUsage(c.Context(), requestedModel, usage, estimated)
	if finalizeErr != nil {
		pipeline.Release(c.Context(), reservation, "usage_finalize_error")

		return cfg.writeUsageFinalizeError(c, finalizeErr)
	}

	if err := pipeline.Settle(c.Context(), SettlementRequest{
		Reservation: reservation,
		TenantID:    authInfo.TenantID,
		Actual:      finalized,
		RequestID:   tunnelReq.RequestID,
		BillingMode: tunnelReq.Deployment.BillingMode,
	}); err != nil {
		return cfg.writeError(c, fiber.StatusPaymentRequired, cfg.settleErrorCode, err.Error())
	}

	responseTimeMS := int(time.Since(start).Milliseconds())
	logEntry := buildRequestLogEntry(c, authInfo, tunnelReq.Deployment, requestedModel, cfg.logPath, cfg.requestType, "success", response.StatusCode, start, responseTimeMS)
	logEntry.RequestID = tunnelReq.RequestID
	logEntry.TraceID = tunnelReq.TraceID
	logEntry.CostUSD = finalized.TotalCostUSD
	logEntry.CostIRR = finalized.TotalCostIRR
	logEntry.InputTokens = finalized.InputTokens
	logEntry.OutputTokens = finalized.OutputTokens
	logEntry.TotalTokens = finalized.TotalTokens
	logEntry.RequestBody = buildJSONRequestLogPayload(originalBody, nil)
	logEntry.ResponseBody = buildJSONResponseLogPayload(response.Body, nil)
	requestLog.FinalizeEntry(c.Context(), logEntry)

	logSpendBestEffort(logger, logs, c.Context(), platform.SpendLog{
		RequestID:      tunnelReq.RequestID,
		TokenHash:      authInfo.TokenHash,
		TenantID:       authInfo.TenantID,
		OrganizationID: authInfo.OrganizationID,
		TeamID:         authInfo.TeamID,
		UserID:         authInfo.UserID,
		OrgUnitID:      authInfo.OrgUnitID,
		Model:          requestedModel,
		InputTokens:    finalized.InputTokens,
		OutputTokens:   finalized.OutputTokens,
		TotalTokens:    finalized.TotalTokens,
		CostUSD:        finalized.TotalCostUSD,
		CostIRR:        finalized.TotalCostIRR,
		ResponseTimeMS: responseTimeMS,
		Status:         "success",
		RequestMetadata: enrichPricingMetadata(tunnelReq.RequestMetadata, authInfo,
			requestedModel),
		CompletionStart: start,
		CompletionEnd:   time.Now(),
	})

	if emitMetric != nil {
		emitMetric(metricsobs.BuildInput{
			TS:                 time.Now().UTC(),
			TenantID:           authInfo.TenantID,
			RequestID:          tunnelReq.RequestID,
			ModelName:          requestedModel,
			ProviderName:       tunnelReq.Deployment.Provider,
			ProviderModelName:  tunnelReq.Deployment.Model,
			RequestType:        cfg.requestType,
			UserID:             authInfo.UserID,
			VirtualAccountID:   authInfo.KeyID,
			VirtualAccountName: authInfo.KeyAlias,
			TeamIDs:            nonEmptySlice([]string{authInfo.TeamID}),
			StatusCode:         response.StatusCode,
			LatencyMS:          responseTimeMS,
			InputTokens:        finalized.InputTokens,
			OutputTokens:       finalized.OutputTokens,
			CostUSD:            finalized.TotalCostUSD,
			Metadata: enrichMetricMetadataForDeployment(map[string]interface{}{
				"endpoint":    cfg.metricEndpoint,
				"key_type":    authInfo.KeyType,
				"team_id":     authInfo.TeamID,
				"org_unit_id": authInfo.OrgUnitID,
				"tenant_slug": tenantSlug,
			}, tunnelReq.Deployment),
		})
	}

	forwardProviderHeaders(c, response.Headers)
	setCostHeaders(c, finalized)

	if contentType := response.Headers["Content-Type"]; contentType != "" {
		c.Set(fiber.HeaderContentType, contentType)
	} else {
		c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	}

	return c.Status(response.StatusCode).Send(response.Body)
}
