// Package proxy hosts OpenAI-compatible request handlers for the control plane.
package proxy

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/apiheaders"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	metricsobs "github.com/Delpak/shahrah/services/control-plane/internal/observability/metrics"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/Delpak/shahrah/services/control-plane/internal/responsecache"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/Delpak/shahrah/services/control-plane/internal/virtualmodels"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// ResponsesHandler serves the OpenAI-compatible Responses API.
type ResponsesHandler struct {
	llmRouter     *router.Router
	pipeline      *RequestPipeline
	billing       budgeter
	resources     providerNativeResourceStore
	deployments   responseDeploymentLister
	providerAuth  responseProviderAuthResolver
	logs          usageLogger
	logger        *zap.Logger
	emitMetric    func(input metricsobs.BuildInput)
	vmResolver    virtualmodels.VirtualModelResolver
	vmEngine      virtualmodels.RoutingEngine
	vmExec        virtualmodels.FallbackExecutor
	responseCache *responsecache.Service
}

// NewResponsesHandlerWithDependencies creates a responses handler from explicit dependencies.
func NewResponsesHandlerWithDependencies(deps BaseHandlerDependencies) *ResponsesHandler {
	resolver, engine, executor := newVirtualModelDeps(
		deps.VirtualModelStore,
		deps.VirtualModelAliasCacheTTL,
		deps.VirtualModelNegativeTTL,
	)

	return &ResponsesHandler{
		llmRouter:     deps.Router,
		pipeline:      deps.requestPipeline(),
		billing:       deps.Billing,
		resources:     deps.ProviderResources,
		deployments:   deps.Deployments,
		providerAuth:  deps.ResponseProviderAuth,
		logs:          deps.logger(),
		logger:        deps.Logger,
		emitMetric:    deps.metricEmitter(),
		vmResolver:    resolver,
		vmEngine:      engine,
		vmExec:        executor,
		responseCache: deps.ResponseCache,
	}
}

// NewResponsesHandler creates a new responses handler.
func NewResponsesHandler(llmRouter *router.Router, tunnel platform.TunnelTransport, billing budgeter, logs usageLogger, logger *zap.Logger, emitMetric ...func(input metricsobs.BuildInput)) *ResponsesHandler {
	var metricEmitter func(input metricsobs.BuildInput)
	if len(emitMetric) > 0 {
		metricEmitter = emitMetric[0]
	}

	return NewResponsesHandlerWithDependencies(BaseHandlerDependencies{
		Router:              llmRouter,
		Tunnel:              tunnel,
		Billing:             billing,
		BillingCapabilities: compatibilityBillingCapabilities(),
		RequestLogger:       logs,
		SpendLogger:         logs,
		Logger:              logger,
		EmitMetric:          metricEmitter,
	})
}

func (h *ResponsesHandler) routeResponseVirtualModel(
	c *fiber.Ctx,
	authInfo *platform.AuthContext,
	tenantID string,
	requestedModel string,
	routeStart time.Time,
) (*router.Deployment, bool, error) {
	if h.vmEngine == nil || h.vmExec == nil {
		return nil, false, nil
	}

	plan, err := resolveVirtualModelRoutePlan(
		c.Context(),
		authInfo,
		h.vmResolver,
		h.vmEngine,
		tenantID,
		requestedModel,
	)
	if err != nil {
		switch {
		case errors.Is(err, virtualmodels.ErrNotVirtualModel):
			return nil, false, nil
		case errors.Is(err, virtualmodels.ErrDisabledVirtualModel):
			return nil, false, writeAPIError(c, fiber.StatusServiceUnavailable, fmt.Sprintf("Virtual model is disabled: %s", requestedModel), "virtual_model_disabled")
		case errors.Is(err, virtualmodels.ErrNoEligibleTargets):
			return nil, false, writeAPIError(c, fiber.StatusServiceUnavailable, fmt.Sprintf("No eligible targets for virtual model: %s", requestedModel), "virtual_model_no_eligible_targets")
		default:
			return nil, false, writeAPIError(c, fiber.StatusServiceUnavailable, fmt.Sprintf("Virtual model routing is unavailable for model: %s", requestedModel), "virtual_model_routing_unavailable")
		}
	}

	firstDeployment := resolveFirstVirtualDeployment(c.Context(), tenantID, plan.targets, h.llmRouter)

	if firstDeployment == nil {
		return nil, false, writeAPIError(c, fiber.StatusServiceUnavailable, fmt.Sprintf("No available target deployment for virtual model: %s", requestedModel), "virtual_model_target_unavailable")
	}

	setVirtualModelRoutePlan(c, plan.model, plan.targets, plan.fromSnapshot, routeStart)

	return firstDeployment, true, nil
}

func (h *ResponsesHandler) resolveResponseVirtualDeployment(
	ctx context.Context,
	tenantID string,
	deployment *router.Deployment,
) (*router.Deployment, error) {
	return resolveVirtualDeploymentStrict(ctx, h.llmRouter, tenantID, deployment)
}

// Handle processes an OpenAI-compatible Responses request.
func (h *ResponsesHandler) Handle(c *fiber.Ctx) error {
	startTime := time.Now()

	authInfo := auth.GetAuthInfo(c)
	if authInfo == nil {
		return writeAPIError(c, fiber.StatusUnauthorized, "Authentication required", "authentication_error")
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return writeAPIError(c, fiber.StatusBadRequest, "Tenant resolution required", "invalid_request_error")
	}

	authInfo.TenantID = currentTenant.ID
	tracing.SetTenantForTrace(c, currentTenant.ID)
	tenantSlug := strings.TrimSpace(currentTenant.Slug)

	originalBody := append([]byte(nil), c.Body()...)
	requestLog := newRequestLogLifecycle(h.logger, h.logs, c, authInfo, "/v1/responses", "responses", startTime)
	requestLog.SetRequestBody(buildJSONRequestLogPayload(originalBody, nil))

	defer requestLog.FinalizeFromFiber()

	var req contractsopenai.ResponseRequest
	if err := json.Unmarshal(originalBody, &req); err != nil {
		return writeAPIError(c, fiber.StatusBadRequest, "Invalid request body", "invalid_request_error")
	}

	modelWasProvided := strings.TrimSpace(req.Model) != ""
	requestedModel := normalizeModelIdentifier(req.Model)
	requestLog.SetModel(requestedModel)

	resolvedDeployment, effectiveModel, upstreamPreviousResponseID, err := h.resolveResponseCreateDeployment(
		c.Context(),
		currentTenant.ID,
		requestedModel,
		strings.TrimSpace(req.PreviousResponseID),
		strings.TrimSpace(c.Get("x-godarai-provider-account")),
	)
	if err != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, err.Error(), "routing_error")
	}

	requestedModel = effectiveModel
	req.Model = requestedModel
	requestLog.SetModel(requestedModel)

	if requestedModel == "" {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Unable to determine the model for Responses routing and billing", "routing_error")
	}

	tracing.SetModelForTrace(c, requestedModel)

	authorizationErr := h.authorizeResponsesModel(c, authInfo, currentTenant.ID, requestedModel, startTime, requestLog)
	if authorizationErr != nil {
		return authorizationErr
	}

	authStart := time.Now()
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "auth.validate", "internal", authStart, fiber.StatusOK, map[string]interface{}{"model": requestedModel})

	estimateReq := responseRequestToChatCompletion(req, requestedModel)
	estimateReq.Model = billingEstimateModel(requestedModel)
	quotaStart := time.Now()

	estimated, err := h.billing.EstimateCost(c.Context(), estimateReq)
	if err != nil {
		return writeBillingEstimateError(c, err)
	}

	estimated.Metadata = req.Metadata

	rateLimitErr := h.checkResponsesRateLimit(c, authInfo, requestedModel, estimated.TotalTokens, startTime, quotaStart, requestLog)
	if rateLimitErr != nil {
		return rateLimitErr
	}

	cacheAdmission, cacheRelease, cacheErr := h.admitResponsesCache(
		c,
		req,
		originalBody,
		authInfo,
		currentTenant.ID,
		requestedModel,
		startTime,
		tenantSlug,
		requestLog,
	)
	defer cacheRelease()

	if cacheErr != nil || cacheAdmission.handled {
		return cacheErr
	}

	cachePending := cacheAdmission.pending

	requestID := tracing.TraceIDFromContext(c)
	if requestID == "" {
		requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
	}

	var (
		budgetReservation *platform.BudgetReservation
	)

	budgetReservationAttempted := false

	deferBudgetUntilRoute := shouldDeferBudgetReservationUntilRoute(requestedModel) ||
		h.shouldDeferResponsesBudgetForVirtualModel(c.Context(), authInfo, currentTenant.ID, requestedModel)
	if !deferBudgetUntilRoute {
		budgetReservationAttempted = true

		budgetReservation, err = h.pipeline.Reserve(c.Context(), budgetRequestFromHTTP(c, authInfo, estimated, requestID))
		if err != nil {
			return writeBudgetCheckError(c, err)
		}
	}

	releaseBudgetOnExit := true
	defer func() {
		if releaseBudgetOnExit {
			h.pipeline.Release(c.Context(), budgetReservation, "request_not_committed")
		}
	}()

	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "quota.check", "internal", quotaStart, fiber.StatusOK, map[string]interface{}{"estimated_tokens": estimated.TotalTokens})

	tunnelErr := h.pipeline.EnsureReady(c.Context())
	if tunnelErr != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Tunnel path unavailable", "tunnel_unavailable")
	}

	routeStart := time.Now()

	var (
		dep              *router.Deployment
		tunnelDeployment platform.Deployment
		isVirtual        bool
	)
	if resolvedDeployment != nil {
		tunnelDeployment = *resolvedDeployment
	} else {
		dep, isVirtual, err = h.routeResponseVirtualModel(c, authInfo, currentTenant.ID, requestedModel, routeStart)
		if err != nil {
			return err
		}

		if !isVirtual {
			for _, candidate := range modelRoutingCandidates(requestedModel) {
				dep, err = h.llmRouter.Route(c.Context(), currentTenant.ID, candidate)
				if err == nil {
					break
				}
			}

			if err != nil || dep == nil {
				if handled, dependencyErr := writeRoutingDependencyUnavailable(c, h.logger, err, requestedModel, currentTenant.ID); handled {
					return dependencyErr
				}

				emitDenyCounter(h.logger, denyReasonUnknownModel, requestedModel, currentTenant.ID)

				return writeAPIError(c, fiber.StatusServiceUnavailable, fmt.Sprintf("No available deployment for model: %s", requestedModel), "routing_error")
			}
		}

		tunnelDeployment = toContractDeployment(dep)
	}

	if !isVirtual {
		if handled, rejectErr := h.rejectDisabledDirectResponsesProvider(c, tunnelDeployment.Provider); handled {
			return rejectErr
		}
	}

	routedEstimate := estimated
	budgetBillingMode := deploymentBillingMode(tunnelDeployment)

	if isVirtual {
		routedEstimate, budgetBillingMode, err = h.preflightVirtualResponseTargets(
			c,
			currentTenant.ID,
			req,
			originalBody,
		)
		if err != nil {
			return writeAPIError(c, fiber.StatusBadGateway, "Virtual model fallback exhausted", "fallback_exhausted")
		}
	} else if dep != nil {
		routedEstimate, err = applyDeploymentPricingOverride(c.Context(), h.pipeline, estimated, dep)
		if err != nil {
			return writeBillingEstimateError(c, err)
		}
	}

	dispatchEstimate := routedEstimate

	if !isVirtual {
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "router.select_model", "internal", routeStart, fiber.StatusOK, map[string]interface{}{"provider": tunnelDeployment.Provider, "deployment_id": tunnelDeployment.ID})
	}

	payload := originalBody

	if modelWasProvided {
		outboundModel := strings.TrimSpace(tunnelDeployment.Model)
		if outboundModel == "" || outboundModel == "*" {
			outboundModel = providerModelID(requestedModel)
		}

		payload, err = rewriteOpenAIJSONModel(payload, outboundModel)
		if err != nil {
			return writeAPIError(c, fiber.StatusBadRequest, "Failed to encode request", "invalid_request_error")
		}
	}

	if upstreamPreviousResponseID != "" && upstreamPreviousResponseID != strings.TrimSpace(req.PreviousResponseID) {
		payload, err = rewriteOpenAIJSONStringField(payload, "previous_response_id", upstreamPreviousResponseID)
		if err != nil {
			return writeAPIError(c, fiber.StatusBadRequest, "Failed to encode request", "invalid_request_error")
		}
	}

	if !isVirtual && strings.EqualFold(strings.TrimSpace(tunnelDeployment.Provider), "xai") {
		payload, err = rewriteXAIFileReferences(c.Context(), h.resources, currentTenant.ID, tunnelDeployment.ProviderAccountID, payload)
		if err != nil {
			return writeXAIFileReferenceError(c, err)
		}
	}

	tunnelReq := platform.TunnelRequest{
		RequestID:  requestID,
		TenantID:   currentTenant.ID,
		TraceID:    requestID,
		Stream:     req.Stream,
		Operation:  platform.ProviderOperationOpenAIResponsesCreate,
		Deployment: tunnelDeployment,
		Headers:    responseProviderHeaders(c),
		Body:       payload,
		RequestMetadata: enrichResponseCacheMetadata(c, map[string]interface{}{
			"tenant_slug":            apiheaders.GetTenant(c),
			"model":                  requestedModel,
			"provider_endpoint_path": "/responses",
		}),
	}
	tunnelReq.RequestMetadata = enrichRequestMetadataForDeployment(tunnelReq.RequestMetadata, tunnelReq.Deployment)
	requestLog.SetDeployment(tunnelReq.Deployment)
	requestLog.SetRequestIDs(tunnelReq.RequestID, tunnelReq.TraceID)
	requestLog.SetRequestBody(buildProviderResponseRequestLogPayload(tunnelReq.Deployment.Provider, req, originalBody))

	if !isVirtual {
		if handled, validationErr := h.prepareDirectResponseRuntimeProfiles(
			c,
			&tunnelReq,
			requestedModel,
			payload,
			estimated,
		); handled || validationErr != nil {
			return validationErr
		}
	}

	if !budgetReservationAttempted {
		budgetReservation, err = h.pipeline.Reserve(c.Context(), budgetRequestFromHTTP(c, authInfo, routedEstimate, requestID, budgetBillingMode))
		if err != nil {
			return writeBudgetCheckError(c, err)
		}
	}

	if req.Stream {
		releaseBudgetOnExit = false
		return h.handleStream(c, req, authInfo, dispatchEstimate, tunnelReq, budgetReservation, startTime, tenantSlug, requestLog)
	}

	err = h.handleSync(c, req, originalBody, authInfo, dispatchEstimate, tunnelReq, budgetReservation, startTime, tenantSlug, cachePending, requestLog)
	if err == nil {
		releaseBudgetOnExit = false
	}

	return err
}

// rejectDisabledDirectResponsesProvider gates a non-virtual Responses request
// on the provider being in the Responses-capable allowlist at all.
func (h *ResponsesHandler) rejectDisabledDirectResponsesProvider(c *fiber.Ctx, provider string) (bool, error) {
	if !supportsResponsesCreateProvider(provider) {
		return true, writeAPIError(c, fiber.StatusNotImplemented, fmt.Sprintf("Responses API is not enabled for provider: %s", provider), "invalid_request_error")
	}

	return false, nil
}

func (h *ResponsesHandler) authorizeResponsesModel(
	c *fiber.Ctx,
	authInfo *platform.AuthContext,
	tenantID string,
	requestedModel string,
	startTime time.Time,
	requestLog *requestLogLifecycle,
) error {
	modelAccess := modelAccessForType(authInfo, requestedModel, platform.ModelTypeResponses)
	if modelAccess.Allowed {
		return nil
	}

	if modelAccess.UnsupportedVMMode {
		return writeUnsupportedModelModeError(c, requestedModel, platform.ModelTypeResponses, modelAccess.ExposedModes)
	}

	denyReason := denyReasonForModel(authInfo)
	emitDenyCounter(h.logger, denyReason, requestedModel, tenantID)

	const message = "Model access denied or unsupported for responses"
	logPreProviderRejection(
		h.logger,
		h.logs,
		c,
		authInfo,
		requestedModel,
		"/v1/responses",
		"responses",
		fiber.StatusForbidden,
		"authorization_error",
		message,
		denyReason,
		startTime,
	)
	requestLog.MarkFinalized()

	return writeAPIError(c, fiber.StatusForbidden, message, "authorization_error")
}

func (h *ResponsesHandler) checkResponsesRateLimit(
	c *fiber.Ctx,
	authInfo *platform.AuthContext,
	requestedModel string,
	estimatedTokens int,
	startTime time.Time,
	quotaStart time.Time,
	requestLog *requestLogLifecycle,
) error {
	rateLimit, err := h.billing.CheckRateLimit(c.Context(), authInfo, requestedModel, estimatedTokens)
	setRateLimitHeaders(c, rateLimit)

	if err == nil {
		return nil
	}

	if recordRateLimitRejection(
		c,
		h.logger,
		h.logs,
		h.emitMetric,
		authInfo,
		requestedModel,
		"/v1/responses",
		"responses",
		startTime,
		quotaStart,
		rateLimit,
		err,
	) {
		requestLog.MarkFinalized()
	}

	return writeAPIError(c, rateLimitErrorStatus(err), err.Error(), "rate_limit_error")
}

func (h *ResponsesHandler) shouldDeferResponsesBudgetForVirtualModel(
	ctx context.Context,
	authInfo *platform.AuthContext,
	tenantID string,
	requestedModel string,
) bool {
	if snapshotVM, ok := buildVirtualModelFromSnapshot(authInfo, requestedModel); ok {
		return snapshotVM != nil
	}

	if h.vmResolver == nil {
		return false
	}

	virtualModel, err := h.vmResolver.Resolve(ctx, tenantID, requestedModel)

	return err == nil && virtualModel != nil
}

func (h *ResponsesHandler) prepareDirectResponseRuntimeProfiles(
	c *fiber.Ctx,
	tunnelRequest *platform.TunnelRequest,
	requestedModel string,
	payload []byte,
	estimated platform.UsageCostBreakdown,
) (bool, error) {
	runtimeModel := strings.TrimSpace(tunnelRequest.Deployment.Model)
	if runtimeModel == "" || runtimeModel == "*" {
		runtimeModel = providerModelID(requestedModel)
	}

	runtimeProfiles, err := runtimeProfilesFromEstimate(
		tunnelRequest.Deployment.Provider,
		runtimeModel,
		estimated,
	)
	if err != nil {
		return true, writeAPIError(c, fiber.StatusServiceUnavailable, err.Error(), "runtime_profile_unavailable")
	}

	tunnelRequest.ModelRuntimeProfiles = runtimeProfiles
	if !runtimeProfilesAllowResponses(runtimeProfiles) {
		return true, writeAPIError(c, fiber.StatusBadRequest, responsesEndpointUnsupportedMessage(requestedModel), "invalid_request_error")
	}

	if !isDeepSeekProvider(tunnelRequest.Deployment.Provider) {
		return false, nil
	}

	if len(tunnelRequest.ModelRuntimeProfiles) != 1 {
		return true, writeAPIError(c, fiber.StatusServiceUnavailable, "DeepSeek model runtime profile is unavailable", "runtime_profile_unavailable")
	}

	validationErr := validateDeepSeekResponsesRuntimeProfile(payload, &tunnelRequest.ModelRuntimeProfiles[0])
	if validationErr != nil {
		return true, writeAPIError(c, fiber.StatusBadRequest, validationErr.Error(), "invalid_request_error")
	}

	return false, nil
}

func (h *ResponsesHandler) preflightVirtualResponseTargets(
	c *fiber.Ctx,
	tenantID string,
	request contractsopenai.ResponseRequest,
	body []byte,
) (platform.UsageCostBreakdown, string, error) {
	state := existingRequestContext(c)
	if state == nil || len(state.virtualModelTargets) == 0 {
		return platform.UsageCostBreakdown{}, "", virtualmodels.ErrNoEligibleTargets
	}

	preparations := make(map[string]virtualResponseTargetPreparation, len(state.virtualModelTargets))

	var (
		walletEstimate    platform.UsageCostBreakdown
		externalEstimate  platform.UsageCostBreakdown
		lastErr           error
		attemptedTargets  int
		hasWalletTarget   bool
		hasExternalTarget bool
	)

	for targetIndex := range state.virtualModelTargets {
		target := state.virtualModelTargets[targetIndex]
		if targetIndex > 0 && !target.Policy.Fallback.Candidate {
			continue
		}

		attemptedTargets++

		targetDeployment := deploymentFromVirtualTarget(target)
		key := virtualModelTargetKey(targetDeployment)
		preparation := virtualResponseTargetPreparation{}

		resolved, err := resolveVirtualDeploymentStrict(c.Context(), h.llmRouter, tenantID, targetDeployment)
		if err == nil {
			preparation.deployment = resolved
			preparation.estimated, preparation.runtimeProfiles, err = h.estimateVirtualResponseTarget(
				c.Context(),
				request,
				body,
				resolved,
			)
		}

		if err != nil {
			preparation.err = fmt.Errorf("prepare virtual model target %s: %w", target.Model, err)
			preparations[key] = preparation
			lastErr = preparation.err

			continue
		}

		preparations[key] = preparation

		billingMode := deploymentBillingMode(toContractDeployment(resolved))
		if billingMode == platform.BillingModeExternalBilled {
			if !hasExternalTarget {
				externalEstimate = preparation.estimated
				hasExternalTarget = true
			}

			continue
		}

		if !hasWalletTarget || isHigherVirtualModelBudgetEstimate(preparation.estimated, walletEstimate) {
			walletEstimate = preparation.estimated
			hasWalletTarget = true
		}
	}

	state.virtualResponseTargets = preparations

	if hasWalletTarget {
		return walletEstimate, platform.BillingModeWalletBilled, nil
	}

	if hasExternalTarget {
		return externalEstimate, platform.BillingModeExternalBilled, nil
	}

	fallbackCount := attemptedTargets - 1
	if fallbackCount < 0 {
		fallbackCount = 0
	}

	failureReason := "no virtual model target passed preparation"
	if lastErr != nil {
		failureReason = lastErr.Error()
	}

	setVirtualModelFailure(c, fallbackCount, failureReason)

	return platform.UsageCostBreakdown{}, "", fmt.Errorf("%w: %s", virtualmodels.ErrFallbackExhausted, failureReason)
}

func (h *ResponsesHandler) estimateVirtualResponseTarget(
	ctx context.Context,
	request contractsopenai.ResponseRequest,
	body []byte,
	deployment *router.Deployment,
) (platform.UsageCostBreakdown, []platform.ModelRuntimeProfileBinding, error) {
	if deployment == nil {
		return platform.UsageCostBreakdown{}, nil, errors.New("virtual model target deployment is unavailable")
	}

	if !supportsResponsesCreateProvider(deployment.Provider) {
		return platform.UsageCostBreakdown{}, nil, fmt.Errorf("responses API is not enabled for provider: %s", deployment.Provider)
	}

	targetBody, err := buildVirtualResponseTargetPayload(request, body, deployment)
	if err != nil {
		return platform.UsageCostBreakdown{}, nil, err
	}

	estimateRequest := responseRequestToChatCompletion(
		request,
		billingEstimateModel(deployment.ModelName),
	)

	estimated, err := h.billing.EstimateCost(ctx, estimateRequest)
	if err != nil {
		return platform.UsageCostBreakdown{}, nil, err
	}

	estimated.Metadata = request.Metadata

	estimated, err = applyDeploymentPricingOverride(ctx, h.pipeline, estimated, deployment)
	if err != nil {
		return platform.UsageCostBreakdown{}, nil, err
	}

	runtimeModel := strings.TrimSpace(deployment.Model)
	if runtimeModel == "" || runtimeModel == "*" {
		runtimeModel = providerModelID(deployment.ModelName)
	}

	runtimeProfiles, err := runtimeProfilesFromEstimate(deployment.Provider, runtimeModel, estimated)
	if err != nil {
		return platform.UsageCostBreakdown{}, nil, err
	}

	if !runtimeProfilesAllowResponses(runtimeProfiles) {
		return platform.UsageCostBreakdown{}, nil, errors.New(responsesEndpointUnsupportedMessage(deployment.ModelName))
	}

	if isDeepSeekProvider(deployment.Provider) {
		if len(runtimeProfiles) != 1 {
			return platform.UsageCostBreakdown{}, nil, errors.New("deepseek model runtime profile is unavailable")
		}

		if validationErr := validateDeepSeekResponsesRuntimeProfile(
			targetBody,
			&runtimeProfiles[0],
		); validationErr != nil {
			return platform.UsageCostBreakdown{}, nil, validationErr
		}
	}

	return estimated, runtimeProfiles, nil
}

func buildVirtualResponseTargetPayload(
	request contractsopenai.ResponseRequest,
	body []byte,
	deployment *router.Deployment,
) ([]byte, error) {
	payload := append([]byte(nil), body...)
	if strings.TrimSpace(request.Model) == "" {
		return payload, nil
	}

	outboundModel := strings.TrimSpace(deployment.Model)
	if outboundModel == "" || outboundModel == "*" {
		outboundModel = providerModelID(deployment.ModelName)
	}

	payload, err := rewriteOpenAIJSONModel(payload, outboundModel)
	if err != nil {
		return nil, fmt.Errorf("rewrite virtual model response target: %w", err)
	}

	return payload, nil
}

func (h *ResponsesHandler) dispatchResponseSync(
	c *fiber.Ctx,
	tenantID string,
	req contractsopenai.ResponseRequest,
	estimated platform.UsageCostBreakdown,
	baseRequest platform.TunnelRequest,
) (*platform.TunnelResponse, platform.TunnelRequest, platform.UsageCostBreakdown, error) {
	selectedEstimate := estimated
	response, selectedRequest, err := h.dispatchVirtualResponseSync(
		c,
		baseRequest,
		func(callContext context.Context, deployment *router.Deployment) (platform.TunnelRequest, error) {
			candidateRequest, candidateEstimate, prepareErr := h.prepareVirtualResponseRequest(
				callContext,
				c,
				tenantID,
				req,
				baseRequest,
				deployment,
			)
			if prepareErr == nil {
				selectedEstimate = candidateEstimate
			}

			return candidateRequest, prepareErr
		},
	)

	return response, selectedRequest, selectedEstimate, err
}

func (h *ResponsesHandler) dispatchVirtualResponseSync(
	c *fiber.Ctx,
	baseRequest platform.TunnelRequest,
	prepare func(context.Context, *router.Deployment) (platform.TunnelRequest, error),
) (*platform.TunnelResponse, platform.TunnelRequest, error) {
	state := existingRequestContext(c)
	if state == nil || state.virtualModel == nil || len(state.virtualModelTargets) == 0 {
		response, err := h.pipeline.Dispatch(c.Context(), baseRequest)

		return response, baseRequest, err
	}

	selectedRequest := baseRequest

	response, selectedDeployment, fallbackCount, failureReason, err := h.vmExec.ExecuteSync(
		c.Context(),
		state.virtualModelTargets,
		func(callContext context.Context, deployment *router.Deployment) (*platform.TunnelResponse, error) {
			candidateRequest, prepareErr := prepare(callContext, deployment)
			if prepareErr != nil {
				return nil, prepareErr
			}

			selectedRequest = candidateRequest

			return h.pipeline.Dispatch(callContext, candidateRequest)
		},
	)
	if err != nil {
		setVirtualModelFailure(c, fallbackCount, failureReason)

		return nil, selectedRequest, err
	}

	if selectedDeployment == nil {
		return nil, selectedRequest, errors.New("virtual model response selected no deployment")
	}

	recordVirtualModelSelection(
		c,
		state.virtualModel,
		selectedDeployment,
		fallbackCount,
		failureReason,
		state.virtualModelRouteAt,
		state.virtualModelSnapshot,
		false,
	)

	return response, selectedRequest, nil
}

func (h *ResponsesHandler) dispatchResponseStream(
	c *fiber.Ctx,
	tenantID string,
	req contractsopenai.ResponseRequest,
	estimated platform.UsageCostBreakdown,
	baseRequest platform.TunnelRequest,
) (platform.TunnelStream, platform.TunnelRequest, platform.UsageCostBreakdown, error) {
	state := existingRequestContext(c)
	if state == nil || state.virtualModel == nil || len(state.virtualModelTargets) == 0 {
		stream, err := h.pipeline.DispatchStream(c.Context(), baseRequest)

		return stream, baseRequest, estimated, err
	}

	selectedRequest := baseRequest
	selectedEstimate := estimated

	stream, selectedDeployment, fallbackCount, failureReason, err := h.vmExec.ExecuteStream(
		c.Context(),
		state.virtualModelTargets,
		func(callContext context.Context, deployment *router.Deployment) (platform.TunnelStream, error) {
			candidateRequest, candidateEstimate, prepareErr := h.prepareVirtualResponseRequest(
				callContext,
				c,
				tenantID,
				req,
				baseRequest,
				deployment,
			)
			if prepareErr != nil {
				return nil, prepareErr
			}

			selectedRequest = candidateRequest
			selectedEstimate = candidateEstimate

			return h.pipeline.DispatchStream(callContext, candidateRequest)
		},
	)
	if err != nil {
		setVirtualModelFailure(c, fallbackCount, failureReason)

		return nil, selectedRequest, selectedEstimate, err
	}

	if selectedDeployment == nil {
		if stream != nil {
			closeStreamBestEffort(h.logger, stream)
		}

		return nil, selectedRequest, selectedEstimate, errors.New("virtual model response stream selected no deployment")
	}

	recordVirtualModelSelection(
		c,
		state.virtualModel,
		selectedDeployment,
		fallbackCount,
		failureReason,
		state.virtualModelRouteAt,
		state.virtualModelSnapshot,
		true,
	)

	return stream, selectedRequest, selectedEstimate, nil
}

func (h *ResponsesHandler) prepareVirtualResponseRequest(
	ctx context.Context,
	c *fiber.Ctx,
	tenantID string,
	req contractsopenai.ResponseRequest,
	baseRequest platform.TunnelRequest,
	deployment *router.Deployment,
) (platform.TunnelRequest, platform.UsageCostBreakdown, error) {
	targetKey := virtualModelTargetKey(deployment)

	var preparation virtualResponseTargetPreparation

	state := existingRequestContext(c)
	if state != nil && state.virtualResponseTargets != nil {
		preparation = state.virtualResponseTargets[targetKey]
	}

	if preparation.err != nil {
		return platform.TunnelRequest{}, platform.UsageCostBreakdown{}, preparation.err
	}

	resolved := preparation.deployment
	if resolved == nil {
		var err error

		resolved, err = h.resolveResponseVirtualDeployment(ctx, tenantID, deployment)
		if err != nil {
			return platform.TunnelRequest{}, platform.UsageCostBreakdown{}, err
		}

		preparation.estimated, preparation.runtimeProfiles, err = h.estimateVirtualResponseTarget(
			ctx,
			req,
			baseRequest.Body,
			resolved,
		)
		if err != nil {
			return platform.TunnelRequest{}, platform.UsageCostBreakdown{}, err
		}
	}

	*deployment = *resolved

	candidateRequest := baseRequest
	candidateRequest.Deployment = toContractDeployment(resolved)

	candidateRequest.ModelRuntimeProfiles = append(
		[]platform.ModelRuntimeProfileBinding(nil),
		preparation.runtimeProfiles...,
	)
	candidateRequest.RequestMetadata = replaceDeploymentRequestMetadata(candidateRequest.RequestMetadata, candidateRequest.Deployment)

	payload, err := buildVirtualResponseTargetPayload(req, baseRequest.Body, resolved)
	if err != nil {
		return platform.TunnelRequest{}, platform.UsageCostBreakdown{}, err
	}

	if strings.EqualFold(strings.TrimSpace(resolved.Provider), "xai") {
		payload, err = rewriteXAIFileReferences(ctx, h.resources, tenantID, resolved.ProviderAccountID, payload)
		if err != nil {
			return platform.TunnelRequest{}, platform.UsageCostBreakdown{}, err
		}
	}

	candidateRequest.Body = payload

	if isDeepSeekProvider(resolved.Provider) {
		if len(candidateRequest.ModelRuntimeProfiles) != 1 {
			return platform.TunnelRequest{}, platform.UsageCostBreakdown{}, errors.New("deepseek model runtime profile is unavailable")
		}

		if validationErr := validateDeepSeekResponsesRuntimeProfile(
			payload,
			&candidateRequest.ModelRuntimeProfiles[0],
		); validationErr != nil {
			return platform.TunnelRequest{}, platform.UsageCostBreakdown{}, validationErr
		}
	}

	return candidateRequest, preparation.estimated, nil
}

func (h *ResponsesHandler) handleSync(
	c *fiber.Ctx,
	req contractsopenai.ResponseRequest,
	originalBody []byte,
	authInfo *platform.AuthContext,
	estimated platform.UsageCostBreakdown,
	tunnelReq platform.TunnelRequest,
	budgetReservation *platform.BudgetReservation,
	start time.Time,
	tenantSlug string,
	cachePending *responsecache.Pending,
	requestLog *requestLogLifecycle,
) error {
	providerStart := time.Now()
	providerAttrs := map[string]interface{}{
		"provider":        tunnelReq.Deployment.Provider,
		"request_payload": buildProviderTraceResponseRequestPayload(tunnelReq.Deployment.Provider, req),
	}

	resp, tunnelReq, estimated, err := h.dispatchResponseSync(c, authInfo.TenantID, req, estimated, tunnelReq)
	providerAttrs["provider"] = tunnelReq.Deployment.Provider
	providerAttrs["request_payload"] = buildProviderTraceResponseRequestPayload(tunnelReq.Deployment.Provider, req)
	requestLog.SetDeployment(tunnelReq.Deployment)
	requestLog.SetRequestIDs(tunnelReq.RequestID, tunnelReq.TraceID)
	requestLog.SetRequestBody(buildProviderResponseRequestLogPayload(tunnelReq.Deployment.Provider, req, originalBody))

	if err != nil {
		providerAttrs["error"] = err.Error()
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, fiber.StatusBadGateway, providerAttrs)

		if errors.Is(err, virtualmodels.ErrFallbackExhausted) {
			return writeAPIError(c, fiber.StatusBadGateway, "Virtual model fallback exhausted", "fallback_exhausted")
		}

		return writeAPIError(c, fiber.StatusBadGateway, fmt.Sprintf("Tunnel error: %v", err), "api_error")
	}

	if resp.StatusCode >= httpErrorFloor {
		providerAttrs["response_payload"] = buildProviderTraceErrorPayload(tunnelReq.Deployment.Provider, resp.StatusCode, resp.Body)
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, resp.StatusCode, providerAttrs)

		return writeOpenAIProviderError(c, resp.StatusCode, resp.Body, resp.Headers)
	}

	var response contractsopenai.Response

	unmarshalErr := json.Unmarshal(resp.Body, &response)
	if unmarshalErr != nil {
		providerAttrs["error"] = "failed to decode provider response"
		providerAttrs["response_payload"] = buildProviderTraceErrorPayload(tunnelReq.Deployment.Provider, fiber.StatusBadGateway, resp.Body)
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, fiber.StatusBadGateway, providerAttrs)

		return writeAPIError(c, fiber.StatusBadGateway, "Failed to decode provider response", "api_error")
	}

	providerAttrs["response_payload"] = buildProviderTraceResponsePayload(tunnelReq.Deployment.Provider, response)
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, resp.StatusCode, providerAttrs)
	h.recordResponseMapping(c.Context(), authInfo.TenantID, tunnelReq.Deployment, response.ID, response.Status, req.Model)

	finalUsage := responseUsageWithToolsToOpenAI(response)
	if finalUsage != nil {
		finalUsage.ServiceTier = response.ServiceTier
	}

	finalCost, err := h.billing.FinalizeFromUsage(c.Context(), billingModelForDeployment(req.Model, tunnelReq.Deployment), finalUsage, estimated)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, "Failed to finalize usage", "billing_error")
	}

	tunnelReq.RequestMetadata = withXAIUsageMetadata(tunnelReq.RequestMetadata, tunnelReq.Deployment.Provider, finalUsage)

	if debitErr := h.pipeline.Settle(c.Context(), SettlementRequest{Reservation: budgetReservation, TenantID: authInfo.TenantID, Actual: finalCost, RequestID: tunnelReq.RequestID, BillingMode: deploymentBillingMode(tunnelReq.Deployment)}); debitErr != nil {
		return writeBudgetCheckError(c, debitErr)
	}

	if cachePending != nil && h.responseCache != nil && strings.EqualFold(strings.TrimSpace(response.Status), "completed") {
		cacheErr := h.responseCache.Store(
			c.Context(),
			cachePending,
			resp.Body,
			tunnelReq.TraceID,
			tunnelReq.Deployment.Provider,
			tunnelReq.Deployment.Model,
			finalCost,
			"gateway-pricing-v1",
		)
		if cacheErr != nil {
			h.logger.Warn("Response cache store failed", zap.Error(cacheErr))
		}
	}

	responseTimeMS := int(time.Since(start).Milliseconds())
	logEntry := buildRequestLogEntry(c, authInfo, tunnelReq.Deployment, req.Model, "/v1/responses", "responses", "success", resp.StatusCode, start, responseTimeMS)
	logEntry.RequestID = tunnelReq.RequestID
	logEntry.TraceID = tunnelReq.TraceID
	logEntry.CostUSD = finalCost.TotalCostUSD
	logEntry.CostIRR = finalCost.TotalCostIRR
	logEntry.InputTokens = finalCost.InputTokens
	logEntry.OutputTokens = finalCost.OutputTokens
	logEntry.TotalTokens = finalCost.TotalTokens
	logEntry.RequestBody = buildProviderResponseRequestLogPayload(tunnelReq.Deployment.Provider, req, originalBody)
	logEntry.ResponseBody = buildProviderResponseLogPayload(tunnelReq.Deployment.Provider, resp.Body, response)
	requestLog.FinalizeEntry(c.Context(), logEntry)
	logSpendBestEffort(h.logger, h.logs, c.Context(), platform.SpendLog{
		RequestID:       tunnelReq.RequestID,
		TokenHash:       authInfo.TokenHash,
		TenantID:        authInfo.TenantID,
		OrganizationID:  authInfo.OrganizationID,
		TeamID:          authInfo.TeamID,
		UserID:          authInfo.UserID,
		OrgUnitID:       authInfo.OrgUnitID,
		Model:           req.Model,
		InputTokens:     finalCost.InputTokens,
		OutputTokens:    finalCost.OutputTokens,
		TotalTokens:     finalCost.TotalTokens,
		CostUSD:         finalCost.TotalCostUSD,
		CostIRR:         finalCost.TotalCostIRR,
		ResponseTimeMS:  responseTimeMS,
		Status:          "success",
		RequestMetadata: enrichPricingMetadata(tunnelReq.RequestMetadata, authInfo, req.Model),
		CompletionStart: start,
		CompletionEnd:   time.Now(),
	})
	h.emitModelMetric(metricsobs.BuildInput{
		TS:                 time.Now().UTC(),
		TenantID:           authInfo.TenantID,
		RequestID:          tunnelReq.RequestID,
		ModelName:          req.Model,
		VirtualModelName:   modelMetricVirtualName(c, req.Model),
		ProviderName:       tunnelReq.Deployment.Provider,
		ProviderModelName:  tunnelReq.Deployment.Model,
		RequestType:        "responses",
		UserID:             authInfo.UserID,
		VirtualAccountID:   authInfo.KeyID,
		VirtualAccountName: authInfo.KeyAlias,
		TeamIDs:            nonEmptySlice([]string{authInfo.TeamID}),
		StatusCode:         resp.StatusCode,
		LatencyMS:          responseTimeMS,
		InputTokens:        finalCost.InputTokens,
		OutputTokens:       finalCost.OutputTokens,
		CostUSD:            finalCost.TotalCostUSD,
		CacheMode:          responseCacheType(c),
		CacheStatus:        responseCacheStatus(c),
		Metadata: enrichMetricMetadataForDeployment(enrichResponseCacheMetadata(c, map[string]interface{}{
			"endpoint":    "/v1/responses",
			"key_type":    authInfo.KeyType,
			"team_id":     authInfo.TeamID,
			"tenant_slug": tenantSlug,
		}), tunnelReq.Deployment),
	})

	forwardProviderHeaders(c, resp.Headers)
	setTraceHeaders(c, tunnelReq.TraceID)
	setCostHeaders(c, finalCost)

	responseStart := time.Now()

	c.Status(resp.StatusCode)
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	err = c.Send(resp.Body)
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "response.write", "internal", responseStart, resp.StatusCode, map[string]interface{}{
		"response_payload": buildProviderTraceResponsePayload(tunnelReq.Deployment.Provider, response),
	})

	return err
}

// rewriteOpenAIJSONModel changes only the routing-owned model field. Using
// RawMessage values preserves every other official or future OpenAI field
// instead of round-tripping through the gateway's necessarily incomplete DTO.
func rewriteOpenAIJSONModel(body []byte, model string) ([]byte, error) {
	return rewriteOpenAIJSONStringField(body, "model", providerModelID(model))
}

func rewriteOpenAIJSONStringField(body []byte, field string, value string) ([]byte, error) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(body, &object); err != nil {
		return nil, err
	}

	encodedValue, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	object[field] = encodedValue

	return json.Marshal(object)
}

func buildResponseRequestLogPayload(req contractsopenai.ResponseRequest, originalBody []byte) map[string]interface{} {
	return buildJSONRequestLogPayload(originalBody, map[string]interface{}{
		"model": req.Model,
		"input": req.Input,
	})
}

func buildProviderResponseRequestLogPayload(provider string, req contractsopenai.ResponseRequest, originalBody []byte) map[string]interface{} {
	if isDeepSeekProvider(provider) {
		return buildProviderTraceResponseRequestPayload(provider, req)
	}

	return buildResponseRequestLogPayload(req, originalBody)
}

func buildProviderResponseLogPayload(provider string, body []byte, response contractsopenai.Response) map[string]interface{} {
	if isDeepSeekProvider(provider) {
		return buildProviderTraceResponsePayload(provider, response)
	}

	return buildJSONResponseLogPayload(body, response)
}

func (h *ResponsesHandler) handleStream(
	c *fiber.Ctx,
	req contractsopenai.ResponseRequest,
	authInfo *platform.AuthContext,
	estimated platform.UsageCostBreakdown,
	tunnelReq platform.TunnelRequest,
	budgetReservation *platform.BudgetReservation,
	start time.Time,
	tenantSlug string,
	requestLog *requestLogLifecycle,
) error {
	providerStart := time.Now()

	stream, tunnelReq, estimated, err := h.dispatchResponseStream(c, authInfo.TenantID, req, estimated, tunnelReq)
	requestLog.SetDeployment(tunnelReq.Deployment)
	requestLog.SetRequestIDs(tunnelReq.RequestID, tunnelReq.TraceID)
	requestLog.SetRequestBody(buildProviderResponseRequestLogPayload(tunnelReq.Deployment.Provider, req, tunnelReq.Body))

	if err != nil {
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, fiber.StatusBadGateway, map[string]interface{}{
			"provider":        tunnelReq.Deployment.Provider,
			"request_payload": buildProviderTraceResponseRequestPayload(tunnelReq.Deployment.Provider, req),
			"stream":          true,
			"error":           err.Error(),
		})

		h.pipeline.Release(c.Context(), budgetReservation, "stream_open_failed")

		if errors.Is(err, virtualmodels.ErrFallbackExhausted) {
			return writeAPIError(c, fiber.StatusBadGateway, "Virtual model fallback exhausted before streaming began", "fallback_exhausted")
		}

		return writeAPIError(c, fiber.StatusBadGateway, fmt.Sprintf("Tunnel error: %v", err), "api_error")
	}

	if statusCode := stream.StatusCode(); statusCode >= httpErrorFloor {
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, statusCode, map[string]interface{}{
			"provider":        tunnelReq.Deployment.Provider,
			"request_payload": buildProviderTraceResponseRequestPayload(tunnelReq.Deployment.Provider, req),
			"stream":          true,
			"status_code":     statusCode,
		})
		h.pipeline.Release(c.Context(), budgetReservation, "provider_http_error")

		return writeOpenAITunnelStreamHTTPError(c, h.logger, stream)
	}

	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, fiber.StatusOK, map[string]interface{}{
		"provider":        tunnelReq.Deployment.Provider,
		"request_payload": buildProviderTraceResponseRequestPayload(tunnelReq.Deployment.Provider, req),
		"stream":          true,
	})

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	setTraceHeaders(c, tunnelReq.TraceID)
	forwardProviderHeaders(c, stream.Headers())
	virtualModelName := modelMetricVirtualName(c, req.Model)

	requestLog.HandOff()

	requestCtx := c.Context()
	requestCtx.SetBodyStreamWriter(func(w *bufio.Writer) {
		defer closeStreamBestEffort(h.logger, stream)

		responseCapture := &streamResponseLogCapture{}
		streamState := &openAIResponseStreamState{}

		budgetCommitted := false

		var terminalResponse *contractsopenai.Response

		defer func() {
			if !budgetCommitted {
				h.pipeline.Release(requestCtx, budgetReservation, "stream_not_committed")
			}
		}()

		finalizeAndLogSuccess := func() {
			if budgetCommitted {
				return
			}

			if streamState.TerminalEventType() == "" {
				h.logger.Warn(
					"Responses stream closed without a terminal event",
					zap.String("request_id", tunnelReq.RequestID),
					zap.String("provider", tunnelReq.Deployment.Provider),
				)

				framed, frameErr := frameOpenAIResponseStreamError(
					fmt.Errorf("provider stream closed without a terminal Responses event"),
					streamState.NextSequenceNumber(),
				)
				if frameErr != nil {
					h.logger.Warn("failed to frame incomplete Responses stream", zap.Error(frameErr))

					return
				}

				if writeErr := writeBufferedBytes(w, framed); writeErr != nil {
					h.logger.Warn("failed to write incomplete Responses stream error", zap.Error(writeErr))
				}

				responseCapture.Append(framed)
				failurePayload := responseCapture.Payload("failed")
				failurePayload["error"] = map[string]interface{}{
					"message": "provider stream closed without a terminal Responses event",
					"type":    "stream_incomplete",
				}
				requestLog.Finalize(context.Background(), "failed", fiber.StatusOK, "stream_incomplete", failurePayload, nil)

				return
			}

			billingModel := billingModelForDeployment(req.Model, tunnelReq.Deployment)

			var finalUsage *contractsopenai.Usage
			if terminalResponse != nil {
				finalUsage = responseUsageWithToolsToOpenAI(*terminalResponse)
				if finalUsage != nil {
					finalUsage.ServiceTier = terminalResponse.ServiceTier
				}
			}

			finalCost, finalizeErr := h.billing.FinalizeFromUsage(requestCtx, billingModel, finalUsage, estimated)
			if finalizeErr != nil {
				h.logger.Warn("failed to finalize streaming usage", zap.Error(finalizeErr), zap.String("request_id", tunnelReq.RequestID))

				finalCost = estimated
				finalCost.Model = billingModel
			}

			tunnelReq.RequestMetadata = withXAIUsageMetadata(tunnelReq.RequestMetadata, tunnelReq.Deployment.Provider, finalUsage)

			if debitErr := h.pipeline.Settle(requestCtx, SettlementRequest{Reservation: budgetReservation, TenantID: authInfo.TenantID, Actual: finalCost, RequestID: tunnelReq.RequestID, BillingMode: deploymentBillingMode(tunnelReq.Deployment)}); debitErr != nil {
				h.logger.Warn("failed to commit budget for streaming request", zap.Error(debitErr), zap.String("request_id", tunnelReq.RequestID))
			} else {
				budgetCommitted = true
			}

			logSpendBestEffort(h.logger, h.logs, requestCtx, platform.SpendLog{
				RequestID:       tunnelReq.RequestID,
				TokenHash:       authInfo.TokenHash,
				TenantID:        authInfo.TenantID,
				OrganizationID:  authInfo.OrganizationID,
				TeamID:          authInfo.TeamID,
				UserID:          authInfo.UserID,
				OrgUnitID:       authInfo.OrgUnitID,
				Model:           req.Model,
				InputTokens:     finalCost.InputTokens,
				OutputTokens:    finalCost.OutputTokens,
				TotalTokens:     finalCost.TotalTokens,
				CostUSD:         finalCost.TotalCostUSD,
				CostIRR:         finalCost.TotalCostIRR,
				ResponseTimeMS:  int(time.Since(start).Milliseconds()),
				Status:          "success",
				RequestMetadata: enrichPricingMetadata(tunnelReq.RequestMetadata, authInfo, req.Model),
				CompletionStart: start,
				CompletionEnd:   time.Now(),
			})

			responsePayload := responseCapture.Payload("completed")
			requestLog.Finalize(context.Background(), "success", fiber.StatusOK, "", responsePayload, &finalCost)
			h.emitModelMetric(metricsobs.BuildInput{
				TS:                 time.Now().UTC(),
				TenantID:           authInfo.TenantID,
				RequestID:          tunnelReq.RequestID,
				ModelName:          req.Model,
				VirtualModelName:   virtualModelName,
				ProviderName:       tunnelReq.Deployment.Provider,
				ProviderModelName:  tunnelReq.Deployment.Model,
				RequestType:        "responses",
				UserID:             authInfo.UserID,
				VirtualAccountID:   authInfo.KeyID,
				VirtualAccountName: authInfo.KeyAlias,
				TeamIDs:            nonEmptySlice([]string{authInfo.TeamID}),
				StatusCode:         fiber.StatusOK,
				LatencyMS:          int(time.Since(start).Milliseconds()),
				InputTokens:        finalCost.InputTokens,
				OutputTokens:       finalCost.OutputTokens,
				CostUSD:            finalCost.TotalCostUSD,
				Metadata: enrichMetricMetadataForDeployment(map[string]interface{}{
					"endpoint":    "/v1/responses",
					"key_type":    authInfo.KeyType,
					"team_id":     authInfo.TeamID,
					"tenant_slug": tenantSlug,
				}, tunnelReq.Deployment),
			})
		}

		relayTunnelStreamWithErrorFramer(
			stream,
			w,
			finalizeAndLogSuccess,
			h.logger,
			tunnelReq.RequestID,
			"responses",
			func(streamErr error) ([]byte, error) {
				return frameOpenAIResponseStreamError(streamErr, streamState.NextSequenceNumber())
			},
			streamState.Observe,
			func(data []byte) {
				responseCapture.Append(data)
			},
			h.newResponseStreamObserver(
				requestCtx,
				authInfo.TenantID,
				tunnelReq.Deployment,
				req.Model,
				func(response contractsopenai.Response) {
					if isTerminalResponseStatus(response.Status) {
						responseCopy := response
						terminalResponse = &responseCopy
					}
				},
			),
		)

		if !requestLog.Finalized() {
			failurePayload := responseCapture.Payload("failed")
			failurePayload["error"] = map[string]interface{}{
				"message": "provider response stream was interrupted",
				"type":    "streaming_error",
			}
			requestLog.Finalize(context.Background(), "failed", fiber.StatusOK, "streaming_error", failurePayload, nil)
		}
	})

	responseStart := time.Now()
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "response.write", "internal", responseStart, fiber.StatusOK, map[string]interface{}{"stream": true})

	return nil
}

func isTerminalResponseStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "completed", "incomplete", "failed":
		return true
	default:
		return false
	}
}

func (h *ResponsesHandler) emitModelMetric(input metricsobs.BuildInput) {
	if h.emitMetric == nil {
		return
	}

	h.emitMetric(input)
}

func supportsResponsesCreateProvider(provider string) bool {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "openai", "xai", "deepseek", "tokenbazaar", "mock":
		return true
	default:
		return false
	}
}

// runtimeProfilesAllowResponses reports whether every catalog runtime profile
// bound to the request declares the Responses endpoint. A model such as
// Gemini, whose official API has no Responses surface, is rejected here
// instead of failing later inside the provider adapter.
func runtimeProfilesAllowResponses(bindings []platform.ModelRuntimeProfileBinding) bool {
	for i := range bindings {
		if !bindings[i].Profile.AllowsEndpoint("responses") {
			return false
		}
	}

	return true
}

func responsesEndpointUnsupportedMessage(model string) string {
	return fmt.Sprintf("Model %s does not support the Responses API; use /v1/chat/completions instead", model)
}

func supportsResponseResourcesProvider(provider string) bool {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "openai", "xai":
		return true
	default:
		return false
	}
}

func responseRequestToChatCompletion(req contractsopenai.ResponseRequest, model string) contractsopenai.ChatCompletionRequest {
	content := flattenResponseInput(req.Input)
	if content == "" {
		content = strings.TrimSpace(req.Instructions)
	}

	messages := []contractsopenai.ChatMessage{}
	if strings.TrimSpace(req.Instructions) != "" {
		messages = append(messages, contractsopenai.ChatMessage{Role: "system", Content: strings.TrimSpace(req.Instructions)})
	}

	if strings.TrimSpace(content) != "" {
		messages = append(messages, contractsopenai.ChatMessage{Role: "user", Content: content})
	}

	if len(messages) == 0 {
		messages = append(messages, contractsopenai.ChatMessage{Role: "user", Content: " "})
	}

	return contractsopenai.ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		Stream:      req.Stream,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		MaxTokens:   req.MaxOutputTokens,
		Tools:       req.Tools,
		ToolChoice:  req.ToolChoice,
		User:        req.User,
		ServiceTier: req.ServiceTier,
	}
}

func flattenResponseInput(input interface{}) string {
	switch v := input.(type) {
	case string:
		return strings.TrimSpace(v)
	case []interface{}:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			if text := flattenResponseInput(item); text != "" {
				parts = append(parts, text)
			}
		}

		return strings.TrimSpace(strings.Join(parts, "\n"))
	case map[string]interface{}:
		if text, ok := v["text"].(string); ok && strings.TrimSpace(text) != "" {
			return strings.TrimSpace(text)
		}

		if content, ok := v["content"]; ok {
			return flattenResponseInput(content)
		}

		if inputText, ok := v["input_text"].(string); ok && strings.TrimSpace(inputText) != "" {
			return strings.TrimSpace(inputText)
		}

		return ""
	default:
		return ""
	}
}

func responseUsageToOpenAI(usage *contractsopenai.ResponseUsage) *contractsopenai.Usage {
	if usage == nil {
		return nil
	}

	result := &contractsopenai.Usage{
		PromptTokens:           usage.InputTokens,
		CompletionTokens:       usage.OutputTokens,
		TotalTokens:            usage.TotalTokens,
		CostInUSDTicks:         usage.CostInUSDTicks,
		NumServerSideToolsUsed: usage.NumServerSideToolsUsed,
	}
	if usage.InputDetails != nil {
		cachedTokens := usage.InputDetails.CachedTokens

		if cachedTokens < 0 {
			cachedTokens = 0
		}

		if cachedTokens > usage.InputTokens {
			cachedTokens = usage.InputTokens
		}

		result.PromptTokensDetails = &contractsopenai.UsageTokenDetails{
			CachedTokens:     cachedTokens,
			CacheWriteTokens: usage.InputDetails.CacheWriteTokens,
		}
		result.PromptCacheHitTokens = cachedTokens
		result.PromptCacheMissTokens = usage.InputTokens - cachedTokens
	}

	if usage.OutputDetails != nil {
		result.CompletionTokensDetails = &contractsopenai.UsageTokenDetails{
			ReasoningTokens: usage.OutputDetails.ReasoningTokens,
		}
	}

	return result
}

func responseUsageWithToolsToOpenAI(
	response contractsopenai.Response,
) *contractsopenai.Usage {
	usage := responseUsageToOpenAI(response.Usage)
	if usage == nil {
		return nil
	}

	serverTools := map[string]int{}

	for _, item := range response.Output {
		switch strings.ToLower(strings.TrimSpace(item.Type)) {
		case "web_search_call":
			serverTools["web_search_requests"]++
		case "x_search_call":
			serverTools["x_search_requests"]++
		case "code_execution_call", "code_interpreter_call":
			serverTools["code_execution_requests"]++
		case "attachment_search_call":
			serverTools["attachment_search_requests"]++
		case "collections_search_call":
			serverTools["collections_search_requests"]++
		case "file_search_call":
			serverTools["file_search_requests"]++
		}
	}

	if len(serverTools) > 0 {
		usage.ServerToolUse = serverTools
	}

	return usage
}
