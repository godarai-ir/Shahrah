package proxy

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/apiheaders"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/guardrails"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/Delpak/shahrah/services/control-plane/internal/responsecache"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
)

type preparedChatRequest struct {
	startedAt          time.Time
	authInfo           *platform.AuthContext
	tenant             *platform.Tenant
	tenantSlug         string
	originalBody       []byte
	request            contractsopenai.ChatCompletionRequest
	requestedModel     string
	resolvedModel      string
	routeCandidates    []string
	guardrailSelection guardrails.RequestSelection
	cacheConfig        *responsecache.RequestConfig
	responseHandled    bool
}

type admittedChatRequest struct {
	estimated            platform.UsageCostBreakdown
	requestID            string
	reservation          *platform.BudgetReservation
	reservationAttempted bool
	cachePending         *responsecache.Pending
	cacheRelease         func()
	responseHandled      bool
}

type routedChatRequest struct {
	tunnelRequest   platform.TunnelRequest
	estimated       platform.UsageCostBreakdown
	reservation     *platform.BudgetReservation
	responseHandled bool
}

func handledPreparedChatRequest(err error) (preparedChatRequest, error) {
	return preparedChatRequest{responseHandled: true}, err
}

func handledAdmittedChatRequest(admitted admittedChatRequest, err error) (admittedChatRequest, error) {
	admitted.responseHandled = true

	return admitted, err
}

func handledRoutedChatRequest(err error) (routedChatRequest, error) {
	return routedChatRequest{responseHandled: true}, err
}

func (h *ChatCompletionHandler) prepareChatRequest(
	c *fiber.Ctx,
	startedAt time.Time,
	requestLog *requestLogLifecycle,
) (preparedChatRequest, error) {
	authInfo := auth.GetAuthInfo(c)
	if authInfo == nil {
		return handledPreparedChatRequest(writeAPIError(c, fiber.StatusUnauthorized, "Authentication required", "authentication_error"))
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return handledPreparedChatRequest(writeAPIError(c, fiber.StatusBadRequest, "Tenant resolution required", "invalid_request_error"))
	}

	authInfo.TenantID = currentTenant.ID
	tracing.SetTenantForTrace(c, currentTenant.ID)

	if err := parsePlaygroundObservability(c); err != nil {
		return handledPreparedChatRequest(writeAPIError(c, fiber.StatusBadRequest, err.Error(), "invalid_request_error"))
	}

	originalBody := append([]byte(nil), c.Body()...)

	var request contractsopenai.ChatCompletionRequest
	if err := json.Unmarshal(originalBody, &request); err != nil {
		return handledPreparedChatRequest(writeAPIError(c, fiber.StatusBadRequest, "Invalid request body", "invalid_request_error"))
	}

	if err := h.applyPromptVersion(c, currentTenant.ID, &request); err != nil {
		return handledPreparedChatRequest(err)
	}

	requestedModel := normalizeModelIdentifier(request.Model)

	routeCandidates := modelRoutingCandidates(requestedModel)
	if len(routeCandidates) == 0 {
		return handledPreparedChatRequest(writeAPIError(c, fiber.StatusBadRequest, "Model is required", "invalid_request_error"))
	}

	resolvedModel := providerModelID(requestedModel)
	tracing.SetModelForTrace(c, requestedModel)

	if handled, err := validateChatCompletionRequest(c, request); handled || err != nil {
		return handledPreparedChatRequest(err)
	}

	modelAccess := modelAccessForType(authInfo, requestedModel, platform.ModelTypeChat)
	if !modelAccess.Allowed {
		if modelAccess.UnsupportedVMMode {
			return handledPreparedChatRequest(writeUnsupportedModelModeError(c, requestedModel, platform.ModelTypeChat, modelAccess.ExposedModes))
		}

		denyReason := denyReasonForModel(authInfo)
		emitDenyCounter(h.logger, denyReason, requestedModel, currentTenant.ID)

		const message = "Model access denied or unsupported for chat completions"
		logPreProviderRejection(h.logger, h.logs, c, authInfo, requestedModel, "/v1/chat/completions", "chat_completions", fiber.StatusForbidden, "authorization_error", message, denyReason, startedAt)
		requestLog.MarkFinalized()

		return handledPreparedChatRequest(writeAPIError(c, fiber.StatusForbidden, message, "authorization_error"))
	}

	guardrailSelection, err := h.parseGuardrails(c)
	if err != nil {
		return handledPreparedChatRequest(writeAPIError(c, fiber.StatusBadRequest, err.Error(), "invalid_request_error"))
	}

	if h.guardrails != nil && !guardrailSelection.Empty() {
		guardrailOverrides, resolveErr := h.resolveTenantGuardrails(c.Context(), currentTenant.ID, guardrailSelection)
		if resolveErr != nil {
			return handledPreparedChatRequest(writeAPIError(c, fiber.StatusBadGateway, "Failed to resolve guardrails", "guardrail_resolution_error"))
		}

		setGuardrailOverrides(c, guardrailOverrides)
		inputResult, inputErr := h.guardrails.ApplyInputWithGuardrails(request, guardrailSelection, guardrailOverrides)
		h.recordGuardrailChecks(c, inputResult.Checks)

		if inputErr != nil {
			return handledPreparedChatRequest(writeGuardrailError(c, fiber.StatusBadRequest, "Request blocked by GodarAI guardrail", inputResult.Checks))
		}

		request = inputResult.Request
	}

	if request.Stream && len(guardrailSelection.LLMOutput) > 0 {
		return handledPreparedChatRequest(writeAPIError(c, fiber.StatusBadRequest, "Streaming with output guardrails is not supported in this MVP", "invalid_request_error"))
	}

	cacheConfig, handled, err := h.chatResponseCacheConfig(c, request)
	if err != nil || handled {
		return handledPreparedChatRequest(err)
	}

	return preparedChatRequest{
		startedAt:          startedAt,
		authInfo:           authInfo,
		tenant:             currentTenant,
		tenantSlug:         strings.TrimSpace(currentTenant.Slug),
		originalBody:       originalBody,
		request:            request,
		requestedModel:     requestedModel,
		resolvedModel:      resolvedModel,
		routeCandidates:    routeCandidates,
		guardrailSelection: guardrailSelection,
		cacheConfig:        cacheConfig,
	}, nil
}

func (h *ChatCompletionHandler) chatResponseCacheConfig(
	c *fiber.Ctx,
	request contractsopenai.ChatCompletionRequest,
) (*responsecache.RequestConfig, bool, error) {
	rawCacheConfig := apiheaders.GetCacheConfig(c)
	if rawCacheConfig == "" || h.responseCache == nil {
		return nil, false, nil
	}

	parsed, err := h.responseCache.ParseHeader(rawCacheConfig)
	if err != nil {
		return nil, true, writeAPIError(c, fiber.StatusBadRequest, err.Error(), "invalid_request_error")
	}

	if !h.responseCache.Enabled() || request.Stream || chatRequestHasExtendedThinking(request) {
		c.Set(apiheaders.HeaderCacheStatus, "bypass")

		return nil, false, nil
	}

	return &parsed, false, nil
}

func (h *ChatCompletionHandler) admitChatRequest(
	c *fiber.Ctx,
	prepared preparedChatRequest,
	requestLog *requestLogLifecycle,
) (admittedChatRequest, error) {
	authStart := time.Now()
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "auth.validate", "internal", authStart, fiber.StatusOK, map[string]interface{}{"model": prepared.requestedModel})

	quotaStart := time.Now()
	isVirtualModel := h.virtualModelForBudgetDecision(
		c.Context(),
		prepared.tenant.ID,
		prepared.authInfo,
		prepared.requestedModel,
	) != nil

	estimateRequest := prepared.request
	estimateRequest.Model = billingEstimateModel(prepared.requestedModel)

	estimated, err := h.billing.EstimateCost(c.Context(), estimateRequest)
	if err != nil {
		return handledAdmittedChatRequest(admittedChatRequest{}, writeBillingEstimateError(c, err))
	}

	estimated.Metadata = prepared.request.Metadata

	if !isVirtualModel {
		var handled bool

		estimated, handled, err = applyAnthropicChatPricingPolicy(c, prepared.request, estimated)
		if err != nil || handled {
			return handledAdmittedChatRequest(admittedChatRequest{}, err)
		}
	}

	rateLimit, err := h.billing.CheckRateLimit(c.Context(), prepared.authInfo, prepared.requestedModel, estimated.TotalTokens)
	setRateLimitHeaders(c, rateLimit)

	if err != nil {
		if recordRateLimitRejection(c, h.logger, h.logs, h.emitMetric, prepared.authInfo, prepared.requestedModel, "/v1/chat/completions", "chat_completions", prepared.startedAt, quotaStart, rateLimit, err) {
			requestLog.MarkFinalized()
		}

		return handledAdmittedChatRequest(admittedChatRequest{}, writeAPIError(c, rateLimitErrorStatus(err), err.Error(), "rate_limit_error"))
	}

	cachePending, cacheRelease, cacheHandled, cacheErr := h.handleResponseCacheLookup(c, prepared.request, prepared.originalBody, prepared.authInfo, prepared.tenant.ID, prepared.requestedModel, prepared.cacheConfig, prepared.startedAt, prepared.tenantSlug, requestLog)

	admitted := admittedChatRequest{
		estimated:       estimated,
		cachePending:    cachePending,
		cacheRelease:    cacheRelease,
		responseHandled: cacheHandled,
	}
	if cacheErr != nil || cacheHandled {
		return admitted, cacheErr
	}

	admitted.requestID = tracing.TraceIDFromContext(c)
	if admitted.requestID == "" {
		admitted.requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
	}

	deferBudgetUntilRoute := shouldDeferBudgetReservationUntilRoute(prepared.requestedModel) || isVirtualModel
	if !deferBudgetUntilRoute {
		admitted.reservationAttempted = true

		admitted.reservation, err = h.pipeline.Reserve(c.Context(), budgetRequestFromHTTP(c, prepared.authInfo, admitted.estimated, admitted.requestID))
		if err != nil {
			return handledAdmittedChatRequest(admitted, writeBudgetCheckError(c, err))
		}
	}

	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "quota.check", "internal", quotaStart, fiber.StatusOK, map[string]interface{}{"estimated_tokens": estimated.TotalTokens})

	return admitted, nil
}

func applyAnthropicChatPricingPolicy(
	c *fiber.Ctx,
	request contractsopenai.ChatCompletionRequest,
	estimated platform.UsageCostBreakdown,
) (platform.UsageCostBreakdown, bool, error) {
	estimated, err := applyAnthropicChatPricingEstimate(request, estimated)
	if err != nil {
		return platform.UsageCostBreakdown{}, true, writeAPIError(c, fiber.StatusBadRequest, err.Error(), "invalid_request_error")
	}

	return estimated, false, nil
}

func applyAnthropicChatPricingEstimate(
	request contractsopenai.ChatCompletionRequest,
	estimated platform.UsageCostBreakdown,
) (platform.UsageCostBreakdown, error) {
	if estimated.ModelRuntimeProfile == nil ||
		!strings.EqualFold(strings.TrimSpace(estimated.ModelRuntimeProfile.Provider), "anthropic") {
		return estimated, nil
	}

	pricingConditions, err := anthropicOpenAIToolPricingPolicy(request)
	if err != nil {
		return platform.UsageCostBreakdown{}, err
	}

	estimated.PricingConditions = mergePricingConditions(estimated.PricingConditions, pricingConditions)
	estimated.Speed = estimated.ModelRuntimeProfile.Profile.EffectiveSpeed(request.Speed)
	estimated.Metadata = withAnthropicEffectiveSpeedMetadata(estimated.Metadata, request.Speed, estimated.Speed)

	return estimated, nil
}

func (h *ChatCompletionHandler) routeChatRequest(
	c *fiber.Ctx,
	prepared preparedChatRequest,
	admitted admittedChatRequest,
) (routedChatRequest, error) {
	if err := h.pipeline.EnsureReady(c.Context()); err != nil {
		return handledRoutedChatRequest(writeAPIError(c, fiber.StatusServiceUnavailable, "Tunnel path unavailable", "tunnel_unavailable"))
	}

	baseRequest := platform.TunnelRequest{
		RequestID: admitted.requestID,
		TenantID:  prepared.tenant.ID,
		TraceID:   admitted.requestID,
		Stream:    prepared.request.Stream,
		Operation: platform.ProviderOperationChatCompletions,
		Headers:   map[string]string{"Content-Type": "application/json"},
		RequestMetadata: enrichResponseCacheMetadata(c, map[string]interface{}{
			"tenant_slug": apiheaders.GetTenant(c),
			"model":       prepared.requestedModel,
		}),
	}
	baseRequest.RequestMetadata = enrichPromptMetadata(c, baseRequest.RequestMetadata)
	baseRequest.RequestMetadata = enrichPlaygroundObservability(c, baseRequest.RequestMetadata)

	routeStart := time.Now()

	deployment, isVirtual, err := h.tryRouteVirtualModel(
		c,
		prepared.tenant.ID,
		prepared.authInfo,
		prepared.requestedModel,
		routeStart,
	)
	if err != nil {
		return handledRoutedChatRequest(err)
	}

	deployment, err = h.routeChatCandidates(c.Context(), prepared.tenant.ID, prepared.routeCandidates, deployment)
	if handled, dependencyErr := writeRoutingDependencyUnavailable(c, h.logger, err, prepared.requestedModel, prepared.tenant.ID); handled {
		return handledRoutedChatRequest(dependencyErr)
	}

	var runtimeProfiles []platform.ModelRuntimeProfileBinding

	var routedEstimate platform.UsageCostBreakdown

	budgetBillingMode := deploymentBillingMode(toContractDeployment(deployment))

	if isVirtual {
		routedEstimate, budgetBillingMode, err = h.preflightVirtualChatTargets(
			c,
			prepared.tenant.ID,
			prepared.request,
		)
		if err != nil {
			return handledRoutedChatRequest(writeAPIError(c, fiber.StatusBadGateway, "Virtual model fallback exhausted", "fallback_exhausted"))
		}
	} else {
		if handled, validationErr := h.validateChatDeploymentRequest(c, prepared.tenant.ID, prepared.requestedModel, prepared.request, deployment); handled || validationErr != nil {
			return handledRoutedChatRequest(validationErr)
		}

		runtimeModel := strings.TrimSpace(deployment.Model)
		if runtimeModel == "" || runtimeModel == "*" {
			runtimeModel = prepared.resolvedModel
		}

		runtimeProfiles, err = runtimeProfilesFromEstimate(deployment.Provider, runtimeModel, admitted.estimated)
		if err != nil {
			return handledRoutedChatRequest(writeAPIError(c, fiber.StatusServiceUnavailable, err.Error(), "runtime_profile_unavailable"))
		}

		var runtimeProfile *platform.ModelRuntimeProfileBinding
		if len(runtimeProfiles) == 1 {
			runtimeProfile = &runtimeProfiles[0]
		}

		if handled, validationErr := h.validateChatRuntimeProfile(c, prepared.request, deployment, runtimeProfile); handled || validationErr != nil {
			return handledRoutedChatRequest(validationErr)
		}

		routedEstimate, err = applyDeploymentPricingOverride(c.Context(), h.pipeline, admitted.estimated, deployment)
		if err != nil {
			return handledRoutedChatRequest(writeBillingEstimateError(c, err))
		}
	}

	estimated := routedEstimate

	if !isVirtual {
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "router.select_model", "internal", routeStart, fiber.StatusOK, map[string]interface{}{"provider": deployment.Provider, "deployment_id": deployment.ID})
	}

	payload, err := buildChatCompletionPayload(prepared.request, deployment.Model, prepared.resolvedModel)
	if err != nil {
		return handledRoutedChatRequest(writeAPIError(c, fiber.StatusBadRequest, "Failed to encode request", "invalid_request_error"))
	}

	tunnelRequest := baseRequest
	tunnelRequest.Body = payload
	tunnelRequest.Deployment = toContractDeployment(deployment)

	if !isVirtual && strings.EqualFold(strings.TrimSpace(deployment.Provider), "xai") {
		payload, err = rewriteXAIFileReferences(c.Context(), h.resources, prepared.tenant.ID, deployment.ProviderAccountID, payload)
		if err != nil {
			return handledRoutedChatRequest(writeXAIFileReferenceError(c, err))
		}

		tunnelRequest.Body = payload
		if conversationID := strings.TrimSpace(c.Get("x-grok-conv-id")); conversationID != "" {
			tunnelRequest.Headers["x-grok-conv-id"] = conversationID
		}
	}

	tunnelRequest.ModelRuntimeProfiles = runtimeProfiles
	tunnelRequest.RequestMetadata = enrichRequestMetadataForDeployment(tunnelRequest.RequestMetadata, tunnelRequest.Deployment)

	reservation := admitted.reservation
	if !admitted.reservationAttempted {
		reservation, err = h.pipeline.Reserve(c.Context(), budgetRequestFromHTTP(c, prepared.authInfo, routedEstimate, admitted.requestID, budgetBillingMode))
		if err != nil {
			return handledRoutedChatRequest(writeBudgetCheckError(c, err))
		}
	}

	return routedChatRequest{
		tunnelRequest: tunnelRequest,
		estimated:     estimated,
		reservation:   reservation,
	}, nil
}
