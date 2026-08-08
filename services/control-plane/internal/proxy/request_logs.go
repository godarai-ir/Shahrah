package proxy

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/apiheaders"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

const streamResponseLogCaptureLimit = 32 * 1024

type requestLogLifecycle struct {
	logger      *zap.Logger
	logs        usageLogger
	fiberCtx    *fiber.Ctx
	authInfo    *platform.AuthContext
	deployment  platform.Deployment
	model       string
	route       string
	requestType string
	startedAt   time.Time
	requestBody map[string]interface{}
	metadata    map[string]interface{}
	requestID   string
	traceID     string
	snapshot    *platform.RequestLogEntry
	once        sync.Once
	handedOff   atomic.Bool
	finalized   atomic.Bool
}

func newRequestLogLifecycle(
	logger *zap.Logger,
	logs usageLogger,
	c *fiber.Ctx,
	authInfo *platform.AuthContext,
	route string,
	requestType string,
	startedAt time.Time,
) *requestLogLifecycle {
	return &requestLogLifecycle{
		logger:      logger,
		logs:        logs,
		fiberCtx:    c,
		authInfo:    authInfo,
		route:       strings.TrimSpace(route),
		requestType: strings.TrimSpace(requestType),
		startedAt:   startedAt,
	}
}

func (l *requestLogLifecycle) SetModel(model string) {
	if l == nil || l.handedOff.Load() {
		return
	}

	l.model = strings.TrimSpace(model)
}

func (l *requestLogLifecycle) SetDeployment(deployment platform.Deployment) {
	if l == nil || l.handedOff.Load() {
		return
	}

	l.deployment = deployment
}

func (l *requestLogLifecycle) SetRequestBody(body map[string]interface{}) {
	if l == nil || l.handedOff.Load() {
		return
	}

	l.requestBody = body
}

func (l *requestLogLifecycle) SetMetadata(metadata map[string]interface{}) {
	if l == nil || l.handedOff.Load() {
		return
	}

	l.metadata = metadata
}

func (l *requestLogLifecycle) SetRequestIDs(requestID string, traceID string) {
	if l == nil || l.handedOff.Load() {
		return
	}

	l.requestID = strings.TrimSpace(requestID)
	l.traceID = strings.TrimSpace(traceID)
}

func (l *requestLogLifecycle) MarkFinalized() {
	if l == nil {
		return
	}

	l.once.Do(func() {
		l.finalized.Store(true)
	})
}

func (l *requestLogLifecycle) Finalized() bool {
	return l != nil && l.finalized.Load()
}

func (l *requestLogLifecycle) HandOff() {
	if l == nil || l.handedOff.Load() || l.Finalized() {
		return
	}

	entry := l.buildEntry()
	l.snapshot = &entry
	l.fiberCtx = nil
	l.handedOff.Store(true)
}

func (l *requestLogLifecycle) FinalizeFromFiber() {
	if l == nil || l.handedOff.Load() || l.Finalized() || l.fiberCtx == nil {
		return
	}

	statusCode := l.fiberCtx.Response().StatusCode()
	if statusCode <= 0 {
		statusCode = fiber.StatusOK
	}

	status := "success"
	errorType := ""

	if statusCode >= fiber.StatusBadRequest {
		status = "failed"
		errorType = requestLogErrorType(statusCode, l.fiberCtx.Response().Body())
	}

	responseBody := buildJSONResponseLogPayload(l.fiberCtx.Response().Body(), nil)
	l.Finalize(l.fiberCtx.Context(), status, statusCode, errorType, responseBody, nil)
}

func (l *requestLogLifecycle) Finalize(
	ctx context.Context,
	status string,
	statusCode int,
	errorType string,
	responseBody map[string]interface{},
	finalCost *platform.UsageCostBreakdown,
) {
	if l == nil {
		return
	}

	l.once.Do(func() {
		entry := l.buildEntry()
		if l.snapshot != nil {
			entry = *l.snapshot
		}

		entry.Status = strings.TrimSpace(status)
		entry.StatusCode = statusCode
		entry.ErrorType = strings.TrimSpace(errorType)
		entry.ResponseTimeMS = int(time.Since(l.startedAt).Milliseconds())
		entry.ResponseBody = responseBody

		if finalCost != nil {
			entry.CostUSD = finalCost.TotalCostUSD
			entry.CostIRR = finalCost.TotalCostIRR
			entry.InputTokens = finalCost.InputTokens
			entry.OutputTokens = finalCost.OutputTokens
			entry.TotalTokens = finalCost.TotalTokens
		}

		l.finalized.Store(true)
		logRequestBestEffort(l.logger, l.logs, ctx, entry)
	})
}

func (l *requestLogLifecycle) FinalizeEntry(ctx context.Context, entry platform.RequestLogEntry) {
	if l == nil {
		return
	}

	l.once.Do(func() {
		l.finalized.Store(true)
		logRequestBestEffort(l.logger, l.logs, ctx, entry)
	})
}

func (l *requestLogLifecycle) buildEntry() platform.RequestLogEntry {
	entry := buildRequestLogEntry(
		l.fiberCtx,
		l.authInfo,
		l.deployment,
		l.model,
		l.route,
		l.requestType,
		"pending",
		0,
		l.startedAt,
		int(time.Since(l.startedAt).Milliseconds()),
	)
	entry.RequestBody = l.requestBody

	if l.requestID != "" {
		entry.RequestID = l.requestID
	}

	if l.traceID != "" {
		entry.TraceID = l.traceID
	}

	if len(l.metadata) > 0 {
		if entry.Metadata == nil {
			entry.Metadata = make(map[string]interface{}, len(l.metadata))
		}

		for key, value := range l.metadata {
			entry.Metadata[key] = value
		}
	}

	return entry
}

func requestLogErrorType(statusCode int, body []byte) string {
	var envelope struct {
		Error struct {
			Type string `json:"type"`
		} `json:"error"`
		Type string `json:"type"`
	}

	if err := json.Unmarshal(body, &envelope); err == nil {
		if errorType := strings.TrimSpace(envelope.Error.Type); errorType != "" {
			return errorType
		}

		if errorType := strings.TrimSpace(envelope.Type); errorType != "" {
			return errorType
		}
	}

	if errorType := strings.TrimSpace(ErrorTypeMap[statusCode]); errorType != "" {
		return errorType
	}

	return "api_error"
}

type streamResponseLogCapture struct {
	data      []byte
	truncated bool
}

func (capture *streamResponseLogCapture) Append(data []byte) {
	if capture == nil || len(data) == 0 {
		return
	}

	remaining := streamResponseLogCaptureLimit - len(capture.data)
	if remaining <= 0 {
		capture.truncated = true
		return
	}

	if len(data) > remaining {
		capture.data = append(capture.data, data[:remaining]...)
		capture.truncated = true

		return
	}

	capture.data = append(capture.data, data...)
}

func (capture *streamResponseLogCapture) Payload(outcome string) map[string]interface{} {
	payload := map[string]interface{}{
		"stream":  true,
		"outcome": strings.TrimSpace(outcome),
	}

	if capture != nil && len(capture.data) > 0 {
		payload["raw_stream"] = string(capture.data)
	}

	if capture != nil && capture.truncated {
		payload["truncated"] = true
	}

	return payload
}

func logStreamingRequestBestEffort(
	logger *zap.Logger,
	logs usageLogger,
	ctx context.Context,
	base platform.RequestLogEntry,
	status string,
	statusCode int,
	errorType string,
	responseBody map[string]interface{},
	finalCost *platform.UsageCostBreakdown,
) {
	entry := base
	entry.Status = strings.TrimSpace(status)
	entry.StatusCode = statusCode
	entry.ErrorType = strings.TrimSpace(errorType)
	entry.ResponseTimeMS = int(time.Since(entry.StartedAt).Milliseconds())
	entry.ResponseBody = responseBody

	if len(base.Metadata) > 0 {
		entry.Metadata = make(map[string]interface{}, len(base.Metadata)+1)
		for key, value := range base.Metadata {
			entry.Metadata[key] = value
		}
	} else {
		entry.Metadata = make(map[string]interface{}, 1)
	}

	entry.Metadata["stream"] = true

	if finalCost != nil {
		entry.CostUSD = finalCost.TotalCostUSD
		entry.CostIRR = finalCost.TotalCostIRR
		entry.InputTokens = finalCost.InputTokens
		entry.OutputTokens = finalCost.OutputTokens
		entry.TotalTokens = finalCost.TotalTokens
	}

	logRequestBestEffort(logger, logs, ctx, entry)
}

func buildRequestLogEntry(
	c *fiber.Ctx,
	authInfo *platform.AuthContext,
	deployment platform.Deployment,
	requestedModel string,
	route string,
	requestType string,
	status string,
	statusCode int,
	start time.Time,
	durationMS int,
) platform.RequestLogEntry {
	entry := platform.RequestLogEntry{
		RequestID:      traceIDFromFiber(c),
		TraceID:        traceIDFromFiber(c),
		Model:          strings.TrimSpace(requestedModel),
		Provider:       strings.TrimSpace(deployment.Provider),
		ProviderModel:  strings.TrimSpace(deployment.Model),
		Route:          strings.TrimSpace(route),
		RequestType:    strings.TrimSpace(requestType),
		Status:         strings.TrimSpace(status),
		StatusCode:     statusCode,
		ResponseTimeMS: durationMS,
		StartedAt:      start.UTC(),
	}

	if authInfo != nil {
		entry.TokenHash = strings.TrimSpace(authInfo.TokenHash)
		entry.TenantID = strings.TrimSpace(authInfo.TenantID)
		entry.OrganizationID = strings.TrimSpace(authInfo.OrganizationID)
		entry.OrgUnitID = strings.TrimSpace(authInfo.OrgUnitID)
		entry.TeamID = strings.TrimSpace(authInfo.TeamID)
		entry.ActorID = strings.TrimSpace(authInfo.UserID)

		entry.ActorType = "user"
		if entry.ActorID == "" {
			entry.ActorID = strings.TrimSpace(authInfo.VirtualUserID)
			if entry.ActorID != "" {
				entry.ActorType = "virtual_user"
			}
		}

		if snapshot, ok := authInfo.VirtualModelSnapshots[requestedModel]; ok {
			entry.VirtualModel = strings.TrimSpace(snapshot.Alias)
		}
	}

	if c != nil {
		if state := existingRequestContext(c); state != nil {
			if state.virtualModel != nil {
				entry.VirtualModel = strings.TrimSpace(state.virtualModel.Alias)
			}

			entry.FallbackCount = state.fallbackCount
			entry.FallbackStatus = fallbackStatusFromReason(state.failureReason, entry.FallbackCount)
		}

		entry.CacheStatus = responseCacheStatus(c)
		entry.CacheMode = responseCacheType(c)
	}

	entry.LoadBalanceStatus = "inactive"
	if entry.FallbackStatus == "" {
		entry.FallbackStatus = fallbackStatusFromReason("", entry.FallbackCount)
	}

	if hasDeploymentBillingContext(deployment) {
		entry.Metadata = enrichDeploymentBillingMetadata(entry.Metadata, deployment)
	}

	entry.Metadata = enrichPlaygroundObservability(c, entry.Metadata)

	return entry
}

func traceIDFromFiber(c *fiber.Ctx) string {
	if c == nil {
		return ""
	}

	return strings.TrimSpace(c.GetRespHeader(apiheaders.HeaderTraceID))
}

func fallbackStatusFromReason(reason string, count int) string {
	if strings.TrimSpace(reason) != "" || count > 0 {
		return "active"
	}

	return "inactive"
}

func logPreProviderRejection(
	logger *zap.Logger,
	logs usageLogger,
	c *fiber.Ctx,
	authInfo *platform.AuthContext,
	requestedModel string,
	route string,
	requestType string,
	statusCode int,
	errorType string,
	message string,
	reason string,
	start time.Time,
) {
	entry := buildRequestLogEntry(
		c,
		authInfo,
		platform.Deployment{},
		requestedModel,
		route,
		requestType,
		"blocked",
		statusCode,
		start,
		int(time.Since(start).Milliseconds()),
	)
	requestID := tracing.TraceIDFromContext(c)
	entry.RequestID = requestID
	entry.TraceID = requestID
	entry.ErrorType = strings.TrimSpace(errorType)
	entry.RequestBody = map[string]interface{}{"model": strings.TrimSpace(requestedModel)}
	entry.ResponseBody = map[string]interface{}{
		"error": map[string]interface{}{
			"message": strings.TrimSpace(message),
			"type":    strings.TrimSpace(errorType),
		},
	}
	entry.Metadata = map[string]interface{}{
		"denial_source":    "model_access",
		"rejection_stage":  "pre_provider",
		"rejection_reason": strings.TrimSpace(reason),
	}

	logRequestBestEffort(logger, logs, c.Context(), entry)
}

func buildJSONRequestLogPayload(body []byte, overrides map[string]interface{}) map[string]interface{} {
	payload := make(map[string]interface{}, len(overrides)+1)
	payload[observability.LazyJSONPayloadField] = json.RawMessage(body)

	for key, value := range overrides {
		payload[key] = value
	}

	return payload
}

func buildJSONResponseLogPayload(body []byte, fallback interface{}) map[string]interface{} {
	payload := map[string]interface{}{
		observability.LazyJSONPayloadField: json.RawMessage(body),
	}
	if fallback != nil {
		payload[observability.LazyJSONFallbackField] = fallback
	}

	return payload
}

func materializeLazyLogPayload(payload map[string]interface{}) map[string]interface{} {
	if len(payload) == 0 {
		return payload
	}

	rawValue, lazy := payload[observability.LazyJSONPayloadField]
	if !lazy {
		return payload
	}

	var raw []byte

	switch typed := rawValue.(type) {
	case json.RawMessage:
		raw = typed
	case []byte:
		raw = typed
	case string:
		raw = []byte(typed)
	}

	out := make(map[string]interface{}, len(payload))
	if err := json.Unmarshal(raw, &out); err != nil || out == nil {
		out = make(map[string]interface{}, len(payload))
		if fallback, ok := payload[observability.LazyJSONFallbackField]; ok {
			encoded, marshalErr := json.Marshal(fallback)
			if marshalErr == nil {
				if unmarshalErr := json.Unmarshal(encoded, &out); unmarshalErr != nil {
					out = make(map[string]interface{}, len(payload))
				}
			}

			if len(out) == 0 {
				out["value"] = fallback
			}
		} else if len(raw) > 0 {
			out["raw_body"] = string(raw)
		}
	}

	for key, value := range payload {
		if key == observability.LazyJSONPayloadField || key == observability.LazyJSONFallbackField {
			continue
		}

		out[key] = value
	}

	return out
}

// buildProviderHTTPErrorLogPayload preserves the provider response byte-for-byte
// alongside its decoded JSON fields. Error responses to streaming requests are
// not guaranteed to be JSON (some providers return SSE frames, HTML, or plain
// text), so relying on JSON decoding alone can silently discard the only useful
// diagnostic content.
func buildProviderHTTPErrorLogPayload(body []byte, stream bool) map[string]interface{} {
	payload := buildJSONResponseLogPayload(body, nil)
	payload["outcome"] = "provider_http_error"
	payload["stream"] = stream

	if len(body) > 0 {
		payload["raw_body"] = string(body)
	}

	return payload
}
