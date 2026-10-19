package proxy

import (
	"context"
	"fmt"
	"strings"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/apiheaders"
	metricsobs "github.com/Delpak/shahrah/services/control-plane/internal/observability/metrics"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/Delpak/shahrah/services/control-plane/internal/responsecache"
	"github.com/Delpak/shahrah/services/control-plane/internal/virtualmodels"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type responsesCacheAdmission struct {
	pending *responsecache.Pending
	handled bool
}

func (h *ResponsesHandler) admitResponsesCache(
	c *fiber.Ctx,
	req contractsopenai.ResponseRequest,
	originalBody []byte,
	authInfo *platform.AuthContext,
	tenantID string,
	requestedModel string,
	startTime time.Time,
	tenantSlug string,
	requestLog *requestLogLifecycle,
) (responsesCacheAdmission, func(), error) {
	cacheConfig, handled, err := h.responsesCacheConfig(c, req)
	if err != nil || handled {
		return responsesCacheAdmission{handled: handled}, func() {}, err
	}

	pending, release, cacheHandled, err := h.handleResponsesCacheLookup(
		c,
		req,
		originalBody,
		authInfo,
		tenantID,
		requestedModel,
		cacheConfig,
		startTime,
		tenantSlug,
		requestLog,
	)
	if release == nil {
		release = func() {}
	}

	return responsesCacheAdmission{pending: pending, handled: cacheHandled}, release, err
}

func (h *ResponsesHandler) responsesCacheConfig(
	c *fiber.Ctx,
	request contractsopenai.ResponseRequest,
) (*responsecache.RequestConfig, bool, error) {
	rawCacheConfig := apiheaders.GetCacheConfig(c)
	if rawCacheConfig == "" || h.responseCache == nil {
		return nil, false, nil
	}

	parsed, err := h.responseCache.ParseHeader(rawCacheConfig)
	if err != nil {
		return nil, true, writeAPIError(c, fiber.StatusBadRequest, err.Error(), "invalid_request_error")
	}

	if !h.responseCache.Enabled() || request.Stream {
		c.Set(apiheaders.HeaderCacheStatus, "bypass")
		setResponseCacheRequestMetadata(c, responseCacheRequestMetadata{
			status:    "bypass",
			cacheType: parsed.Type,
		})

		return nil, false, nil
	}

	return &parsed, false, nil
}

func (h *ResponsesHandler) handleResponsesCacheLookup(
	c *fiber.Ctx,
	req contractsopenai.ResponseRequest,
	originalBody []byte,
	authInfo *platform.AuthContext,
	tenantID string,
	requestedModel string,
	cacheConfig *responsecache.RequestConfig,
	startTime time.Time,
	tenantSlug string,
	requestLog *requestLogLifecycle,
) (*responsecache.Pending, func(), bool, error) {
	if cacheConfig == nil {
		return nil, nil, false, nil
	}

	keyID := strings.TrimSpace(authInfo.KeyID)
	if keyID == "" {
		keyID = strings.TrimSpace(authInfo.TokenHash)
	}

	cacheScope := responsecache.Scope{
		TenantID:     tenantID,
		KeyID:        keyID,
		Model:        requestedModel,
		ModelVersion: virtualModelCacheVersion(c.Context(), authInfo, tenantID, requestedModel, h.vmResolver),
	}

	cacheResult, cachePending, lookupErr := h.responseCache.LookupResponse(
		c.Context(),
		cacheScope,
		req,
		originalBody,
		*cacheConfig,
	)
	if lookupErr != nil && cacheResult.Status == "" {
		cacheResult.Status = "error"
		cacheResult.CacheType = cacheConfig.Type
	}

	h.setResponsesCacheResult(c, *cacheConfig, cacheResult)

	if lookupErr != nil {
		h.logger.Warn("Response cache lookup failed", zap.String("cache_type", cacheConfig.Type), zap.Error(lookupErr))
	} else if cacheResult.Entry != nil {
		return cachePending, nil, true, h.handleResponsesCacheHit(
			c,
			req,
			originalBody,
			authInfo,
			cacheResult,
			startTime,
			tenantSlug,
			requestLog,
		)
	}

	if cacheResult.Status == "bypass" {
		return cachePending, nil, false, nil
	}

	release, waited, acquireErr := h.responseCache.AcquireMiss(c.Context(), cachePending)
	if acquireErr != nil {
		return cachePending, nil, false, writeAPIError(
			c,
			fiber.StatusRequestTimeout,
			"Request canceled while waiting for response cache",
			"request_canceled",
		)
	}

	if !waited {
		return cachePending, release, false, nil
	}

	secondResult, secondPending, secondErr := h.responseCache.LookupResponse(
		c.Context(),
		cacheScope,
		req,
		originalBody,
		*cacheConfig,
	)
	cachePending = secondPending

	if secondErr != nil {
		if secondResult.Status == "" {
			secondResult.Status = "error"
			secondResult.CacheType = cacheConfig.Type
		}

		h.logger.Warn("Response cache lookup after wait failed", zap.String("cache_type", cacheConfig.Type), zap.Error(secondErr))
	}

	h.setResponsesCacheResult(c, *cacheConfig, secondResult)

	if secondErr == nil && secondResult.Entry != nil {
		return cachePending, release, true, h.handleResponsesCacheHit(
			c,
			req,
			originalBody,
			authInfo,
			secondResult,
			startTime,
			tenantSlug,
			requestLog,
		)
	}

	return cachePending, release, false, nil
}

func (h *ResponsesHandler) setResponsesCacheResult(
	c *fiber.Ctx,
	cacheConfig responsecache.RequestConfig,
	result responsecache.Result,
) {
	c.Set(apiheaders.HeaderCacheStatus, result.Status)
	setResponseCacheRequestMetadata(c, responseCacheRequestMetadata{
		status:             result.Status,
		cacheType:          cacheConfig.Type,
		embeddingLatencyMS: result.EmbeddingLatencyMS,
	})
}

func (h *ResponsesHandler) handleResponsesCacheHit(
	c *fiber.Ctx,
	req contractsopenai.ResponseRequest,
	originalBody []byte,
	authInfo *platform.AuthContext,
	result responsecache.Result,
	start time.Time,
	tenantSlug string,
	requestLog *requestLogLifecycle,
) error {
	entry := result.Entry
	if entry == nil {
		return writeAPIError(c, fiber.StatusInternalServerError, "Cached response is unavailable", "api_error")
	}

	markVirtualModelCacheHit(c, authInfo, req.Model, h.vmResolver)

	requestID := tracing.TraceIDFromContext(c)
	if requestID == "" {
		requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
	}

	responseTimeMS := int(time.Since(start).Milliseconds())
	metadata := enrichPlaygroundObservability(c, map[string]interface{}{
		"endpoint":                   "/v1/responses",
		"cache_status":               "hit",
		"cache_similarity":           result.Similarity,
		"cache_type":                 result.CacheType,
		"cache_embedding_latency_ms": result.EmbeddingLatencyMS,
		"cached_trace_id":            entry.TraceID,
		"tenant_slug":                tenantSlug,
		"key_type":                   authInfo.KeyType,
	})
	deployment := platform.Deployment{Provider: entry.Provider, Model: entry.ProviderModel}
	logEntry := buildRequestLogEntry(
		c,
		authInfo,
		deployment,
		req.Model,
		"/v1/responses",
		"responses",
		"success",
		fiber.StatusOK,
		start,
		responseTimeMS,
	)
	logEntry.RequestID = requestID
	logEntry.TraceID = entry.TraceID
	logEntry.RequestBody = buildProviderResponseRequestLogPayload(entry.Provider, req, originalBody)
	logEntry.ResponseBody = buildProviderResponseLogPayload(entry.Provider, entry.Response, contractsopenai.Response{})
	logEntry.CacheStatus = "hit"
	requestLog.FinalizeEntry(c.Context(), logEntry)
	logSpendBestEffort(h.logger, h.logs, c.Context(), platform.SpendLog{
		RequestID:       requestID,
		TokenHash:       authInfo.TokenHash,
		TenantID:        authInfo.TenantID,
		TeamID:          authInfo.TeamID,
		UserID:          authInfo.UserID,
		Model:           req.Model,
		ResponseTimeMS:  responseTimeMS,
		Status:          "success",
		RequestMetadata: metadata,
		CompletionStart: start,
		CompletionEnd:   time.Now(),
	})
	h.emitModelMetric(metricsobs.BuildInput{
		TS:                   time.Now().UTC(),
		TenantID:             authInfo.TenantID,
		RequestID:            requestID,
		ModelName:            req.Model,
		VirtualModelName:     modelMetricVirtualName(c, req.Model),
		ProviderName:         entry.Provider,
		ProviderModelName:    entry.ProviderModel,
		RequestType:          "responses",
		UserID:               authInfo.UserID,
		VirtualAccountID:     authInfo.KeyID,
		VirtualAccountName:   authInfo.KeyAlias,
		TeamIDs:              nonEmptySlice([]string{authInfo.TeamID}),
		StatusCode:           fiber.StatusOK,
		LatencyMS:            responseTimeMS,
		CacheMode:            result.CacheType,
		CacheStatus:          "hit",
		CacheMatch:           cacheMatchForResult(result),
		CacheSimilarity:      cacheSimilarityForResult(result),
		CacheOriginRequestID: entry.TraceID,
		AvoidedInputTokens:   entry.CostSnapshot.InputTokens,
		AvoidedOutputTokens:  entry.CostSnapshot.OutputTokens,
		AvoidedCostUSD:       entry.CostSnapshot.CostUSD,
		AvoidedCostIRR:       entry.CostSnapshot.CostIRR,
		PricingVersion:       entry.CostSnapshot.PricingVersion,
		Metadata:             metadata,
	})

	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	c.Set(apiheaders.HeaderCacheStatus, "hit")
	c.Set(apiheaders.HeaderCachedTraceID, entry.TraceID)

	if result.Semantic {
		c.Set(apiheaders.HeaderCacheSimilarity, fmt.Sprintf("%.6f", result.Similarity))
	}

	setTraceHeaders(c, requestID)
	setCostHeaders(c, platform.UsageCostBreakdown{})

	return c.Status(fiber.StatusOK).Send(entry.Response)
}

func markVirtualModelCacheHit(
	c *fiber.Ctx,
	authInfo *platform.AuthContext,
	requestedModel string,
	resolver virtualmodels.VirtualModelResolver,
) {
	virtualModel, _ := buildVirtualModelFromSnapshot(authInfo, requestedModel)
	if virtualModel == nil && resolver != nil {
		resolved, err := resolver.Resolve(c.Context(), authInfo.TenantID, requestedModel)
		if err == nil {
			virtualModel = resolved
		}
	}

	if virtualModel == nil {
		return
	}

	state := requestContext(c)
	state.virtualModel = virtualModel
	state.virtualModelActive = true
}

func virtualModelCacheVersion(
	ctx context.Context,
	authInfo *platform.AuthContext,
	tenantID string,
	model string,
	resolver virtualmodels.VirtualModelResolver,
) string {
	if authInfo == nil {
		return ""
	}

	if snapshot, ok := authInfo.VirtualModelSnapshots[model]; ok {
		return fmt.Sprintf("%s:%d", snapshot.VirtualModelID, snapshot.Version)
	}

	if resolver != nil {
		if virtualModel, err := resolver.Resolve(ctx, tenantID, model); err == nil && virtualModel != nil {
			return fmt.Sprintf("%s:%d", virtualModel.ID, virtualModel.Version)
		}
	}

	return ""
}
