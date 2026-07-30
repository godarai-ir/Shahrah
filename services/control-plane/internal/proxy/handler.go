// Package proxy hosts OpenAI-compatible request handlers for the control plane.
package proxy

import (
	"bufio"
	"bytes"
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
	"github.com/Delpak/shahrah/services/control-plane/internal/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/guardrails"
	metricsobs "github.com/Delpak/shahrah/services/control-plane/internal/observability/metrics"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/Delpak/shahrah/services/control-plane/internal/prompts"
	"github.com/Delpak/shahrah/services/control-plane/internal/responsecache"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/Delpak/shahrah/services/control-plane/internal/virtualmodels"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// HTTP status constant used by sync.go for error floor checks.
const httpErrorFloor = 400

type budgeter interface {
	EstimateCost(ctx context.Context, req contractsopenai.ChatCompletionRequest) (platform.UsageCostBreakdown, error)
	EstimateEmbeddingCost(ctx context.Context, req contractsopenai.EmbeddingRequest) (platform.UsageCostBreakdown, error)
	CheckRateLimit(ctx context.Context, auth *platform.AuthContext, model string, estimatedTokens int) (*cache.RateLimitResult, error)
	CheckBudget(ctx context.Context, auth *platform.AuthContext, breakdown platform.UsageCostBreakdown) error
	FinalizeFromUsage(ctx context.Context, model string, usage *contractsopenai.Usage, estimated platform.UsageCostBreakdown) (platform.UsageCostBreakdown, error)
	FinalizeImageUsage(ctx context.Context, model string, usage *contractsopenai.Usage, providerCostUSDTicks *int64, estimated platform.UsageCostBreakdown) (platform.UsageCostBreakdown, error)
}

type deploymentPricingBudgeter interface {
	ApplyPricingOverride(
		ctx context.Context,
		estimated platform.UsageCostBreakdown,
		effective *platform.ModelPrice,
	) (platform.UsageCostBreakdown, error)
}

type meteredUsageBudgeter interface {
	PriceMeteredUsage(
		ctx context.Context,
		model string,
		dimension string,
		quantity string,
		estimated platform.UsageCostBreakdown,
	) (platform.UsageCostBreakdown, error)
}

func priceMeteredUsage(
	ctx context.Context,
	billing budgeter,
	model string,
	dimension string,
	quantity string,
	estimated platform.UsageCostBreakdown,
) (platform.UsageCostBreakdown, error) {
	metered, ok := billing.(meteredUsageBudgeter)
	if !ok {
		return estimated, nil
	}

	return metered.PriceMeteredUsage(
		ctx,
		model,
		dimension,
		quantity,
		estimated,
	)
}

func applyDeploymentPricingOverride(
	ctx context.Context,
	billing budgeter,
	estimated platform.UsageCostBreakdown,
	deployment *router.Deployment,
) (platform.UsageCostBreakdown, error) {
	if deployment == nil || deployment.PricingMode != platform.PricingModePrivate {
		return estimated, nil
	}
	overrider, ok := billing.(deploymentPricingBudgeter)
	if !ok {
		return platform.UsageCostBreakdown{}, fmt.Errorf("private pricing override is unavailable")
	}
	return overrider.ApplyPricingOverride(ctx, estimated, deployment.EffectivePricing)
}

type walletStore interface {
	DebitWallet(ctx context.Context, tenantID string, breakdown platform.UsageCostBreakdown, requestID string) (*platform.Transaction, error)
}

type usageLogger interface {
	LogRequest(ctx context.Context, entry platform.RequestLogEntry) error
	LogSpend(ctx context.Context, log platform.SpendLog) error
}

type guardrailResolver interface {
	ListEnabledTenantGuardrailsBySlugs(ctx context.Context, tenantID string, slugs []string) ([]platform.TenantGuardrail, error)
}

type promptResolver interface {
	GetPromptVersionByFQN(ctx context.Context, tenantID string, fqn string) (*platform.PromptVersion, error)
}

// ChatCompletionHandler handles chat completion requests.
// ChatCompletionHandler serves chat completion requests.
type ChatCompletionHandler struct {
	llmRouter     *router.Router
	tunnel        platform.TunnelTransport
	billing       budgeter
	wallets       walletStore
	logs          usageLogger
	logger        *zap.Logger
	vmResolver    virtualmodels.VirtualModelResolver
	vmEngine      virtualmodels.RoutingEngine
	vmExec        virtualmodels.FallbackExecutor
	emitMetric    func(input metricsobs.BuildInput)
	responseCache *responsecache.Service
	guardrails    *guardrails.Engine
	guardrailRepo guardrailResolver
	prompts       promptResolver
}

// SetResponseCache enables opt-in response caching without changing existing constructor callers.
func (h *ChatCompletionHandler) SetResponseCache(service *responsecache.Service) {
	h.responseCache = service
}

// SetGuardrails enables per-request LLM guardrail evaluation.
func (h *ChatCompletionHandler) SetGuardrails(engine *guardrails.Engine) {
	h.guardrails = engine
}

// NewChatCompletionHandlerWithDependencies creates a chat handler from explicit dependencies.
func NewChatCompletionHandlerWithDependencies(deps ChatCompletionDependencies) *ChatCompletionHandler {
	resolver, engine, exec := newChatVirtualModelDeps(deps.VirtualModelStore)

	return &ChatCompletionHandler{
		llmRouter:     deps.Base.Router,
		tunnel:        deps.Base.Tunnel,
		billing:       deps.Base.Billing,
		wallets:       deps.Base.Wallets,
		logs:          deps.Base.logger(),
		logger:        deps.Base.Logger,
		vmResolver:    resolver,
		vmEngine:      engine,
		vmExec:        exec,
		emitMetric:    deps.Base.metricEmitter(),
		guardrailRepo: deps.GuardrailResolver,
		prompts:       deps.PromptResolver,
	}
}

// NewChatCompletionHandler creates a new chat completion handler.
func NewChatCompletionHandler(llmRouter *router.Router, tunnel platform.TunnelTransport, billing budgeter, wallets walletStore, logs usageLogger, logger *zap.Logger, emitMetric ...func(input metricsobs.BuildInput)) *ChatCompletionHandler {
	var metricEmitter func(input metricsobs.BuildInput)
	if len(emitMetric) > 0 {
		metricEmitter = emitMetric[0]
	}

	return NewChatCompletionHandlerWithDependencies(ChatCompletionDependencies{
		Base: BaseHandlerDependencies{
			Router:        llmRouter,
			Tunnel:        tunnel,
			Billing:       billing,
			Wallets:       wallets,
			RequestLogger: logs,
			SpendLogger:   logs,
			Logger:        logger,
			EmitMetric:    metricEmitter,
		},
	})
}

// Handle processes a chat completion request.
func (h *ChatCompletionHandler) Handle(c *fiber.Ctx) error {
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

	var req contractsopenai.ChatCompletionRequest
	if err := json.Unmarshal(originalBody, &req); err != nil {
		return writeAPIError(c, fiber.StatusBadRequest, "Invalid request body", "invalid_request_error")
	}

	if err := h.applyPromptVersion(c, currentTenant.ID, &req); err != nil {
		return err
	}

	if req.Model == "" {
		return writeAPIError(c, fiber.StatusBadRequest, "Model is required", "invalid_request_error")
	}

	requestedModel := normalizeModelIdentifier(req.Model)

	routeCandidates := modelRoutingCandidates(requestedModel)
	if len(routeCandidates) == 0 {
		return writeAPIError(c, fiber.StatusBadRequest, "Model is required", "invalid_request_error")
	}

	resolvedModel := providerModelID(requestedModel)
	tracing.SetModelForTrace(c, requestedModel)

	if handled, err := validateChatCompletionRequest(c, req); handled || err != nil {
		return err
	}

	modelAccess := modelAccessForType(authInfo, requestedModel, platform.ModelTypeChat)
	if !modelAccess.Allowed {
		if modelAccess.UnsupportedVMMode {
			return writeUnsupportedModelModeError(c, requestedModel, platform.ModelTypeChat, modelAccess.ExposedModes)
		}

		denyReason := denyReasonForModel(authInfo)
		emitDenyCounter(h.logger, denyReason, requestedModel, currentTenant.ID)

		const message = "Model access denied or unsupported for chat completions"
		logPreProviderRejection(h.logger, h.logs, c, authInfo, requestedModel, "/v1/chat/completions", "chat_completions", fiber.StatusForbidden, "authorization_error", message, denyReason, startTime)

		return writeAPIError(c, fiber.StatusForbidden, message, "authorization_error")
	}

	guardrailSelection, err := h.parseGuardrails(c)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadRequest, err.Error(), "invalid_request_error")
	}

	if h.guardrails != nil && !guardrailSelection.Empty() {
		guardrailOverrides, resolveErr := h.resolveTenantGuardrails(c.Context(), currentTenant.ID, guardrailSelection)
		if resolveErr != nil {
			return writeAPIError(c, fiber.StatusBadGateway, "Failed to resolve guardrails", "guardrail_resolution_error")
		}

		c.Locals("guardrail_overrides", guardrailOverrides)
		inputResult, inputErr := h.guardrails.ApplyInputWithGuardrails(req, guardrailSelection, guardrailOverrides)
		h.recordGuardrailChecks(c, inputResult.Checks)

		if inputErr != nil {
			return writeGuardrailError(c, fiber.StatusBadRequest, "Request blocked by GodarAI guardrail", inputResult.Checks)
		}

		req = inputResult.Request
	}

	if req.Stream && len(guardrailSelection.LLMOutput) > 0 {
		return writeAPIError(c, fiber.StatusBadRequest, "Streaming with output guardrails is not supported in this MVP", "invalid_request_error")
	}

	var cacheConfig *responsecache.RequestConfig

	if rawCacheConfig := apiheaders.GetCacheConfig(c); rawCacheConfig != "" && h.responseCache != nil {
		parsed, parseErr := h.responseCache.ParseHeader(rawCacheConfig)
		if parseErr != nil {
			return writeAPIError(c, fiber.StatusBadRequest, parseErr.Error(), "invalid_request_error")
		}

		cacheConfig = &parsed

		if !h.responseCache.Enabled() || req.Stream || chatRequestHasExtendedThinking(req) {
			c.Set(apiheaders.HeaderCacheStatus, "bypass")

			cacheConfig = nil
		}
	}

	authStart := time.Now()
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "auth.validate", "internal", authStart, fiber.StatusOK, map[string]interface{}{"model": requestedModel})

	quotaStart := time.Now()
	estimateReq := req
	estimateReq.Model = billingEstimateModel(requestedModel)

	estimated, err := h.billing.EstimateCost(c.Context(), estimateReq)
	if err != nil {
		return writeBillingEstimateError(c, err)
	}

	estimated.Metadata = req.Metadata
	if estimated.ModelRuntimeProfile != nil &&
		strings.EqualFold(
			strings.TrimSpace(estimated.ModelRuntimeProfile.Provider),
			"anthropic",
		) {
		pricingConditions, policyErr :=
			anthropicOpenAIToolPricingPolicy(req)
		if policyErr != nil {
			return writeAPIError(
				c,
				fiber.StatusBadRequest,
				policyErr.Error(),
				"invalid_request_error",
			)
		}
		estimated.PricingConditions = mergePricingConditions(
			estimated.PricingConditions,
			pricingConditions,
		)
		estimated.Speed = estimated.ModelRuntimeProfile.Profile.EffectiveSpeed(
			req.Speed,
		)
		estimated.Metadata = withAnthropicEffectiveSpeedMetadata(
			estimated.Metadata,
			req.Speed,
			estimated.Speed,
		)
	}

	rateLimit, err := h.billing.CheckRateLimit(c.Context(), authInfo, requestedModel, estimated.TotalTokens)
	setRateLimitHeaders(c, rateLimit)

	if err != nil {
		recordRateLimitRejection(c, h.logger, h.logs, h.emitMetric, authInfo, requestedModel, "/v1/chat/completions", "chat_completions", startTime, quotaStart, rateLimit, err)
		return writeAPIError(c, rateLimitErrorStatus(err), err.Error(), "rate_limit_error")
	}

	cachePending, cacheRelease, cacheHandled, cacheErr := h.handleResponseCacheLookup(c, req, originalBody, authInfo, currentTenant.ID, requestedModel, cacheConfig, startTime, tenantSlug)
	if cacheRelease != nil {
		defer cacheRelease()
	}

	if cacheErr != nil || cacheHandled {
		return cacheErr
	}

	requestID := tracing.TraceIDFromContext(c)
	if requestID == "" {
		requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
	}

	var (
		budgetReservation     *platform.BudgetReservation
		budgetResponseHandled bool
	)

	budgetReservationAttempted := false

	deferBudgetUntilRoute := shouldDeferBudgetReservationUntilRoute(requestedModel) || h.shouldDeferBudgetForVirtualModel(c.Context(), currentTenant.ID, authInfo, requestedModel)
	if !deferBudgetUntilRoute {
		budgetReservationAttempted = true

		budgetReservation, budgetResponseHandled, err = reserveBudgetForRequest(c, h.billing, authInfo, estimated, requestID)
		if err != nil || budgetResponseHandled {
			return err
		}
	}

	releaseBudgetOnExit := true
	defer func() {
		if releaseBudgetOnExit {
			releaseBudgetBestEffort(c.Context(), h.logger, h.billing, budgetReservation, "request_not_committed")
		}
	}()

	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "quota.check", "internal", quotaStart, fiber.StatusOK, map[string]interface{}{"estimated_tokens": estimated.TotalTokens})

	tunnelErr := h.tunnel.Healthy(c.Context())
	if tunnelErr != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Tunnel path unavailable", "tunnel_unavailable")
	}

	baseReq := platform.TunnelRequest{
		RequestID: requestID,
		TenantID:  currentTenant.ID,
		TraceID:   requestID,
		Stream:    req.Stream,
		Operation: platform.ProviderOperationChatCompletions,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		RequestMetadata: enrichResponseCacheMetadata(c, map[string]interface{}{
			"tenant_slug": apiheaders.GetTenant(c),
			"model":       requestedModel,
		}),
	}
	baseReq.RequestMetadata = enrichPromptMetadata(c, baseReq.RequestMetadata)

	routeStart := time.Now()

	deployment, isVirtual, routeErr := h.tryRouteVirtualModel(c, currentTenant.ID, authInfo, requestedModel, req, resolvedModel, baseReq, routeStart)
	if routeErr != nil {
		return routeErr
	}

	deployment, routeErr = h.routeChatCandidates(c.Context(), currentTenant.ID, routeCandidates, deployment)
	if handled, dependencyErr := writeRoutingDependencyUnavailable(c, h.logger, routeErr, requestedModel, currentTenant.ID); handled {
		return dependencyErr
	}

	if handled, validationErr := h.validateChatDeployment(
		c,
		currentTenant.ID,
		requestedModel,
		req,
		deployment,
		estimated.ModelRuntimeProfile,
	); handled || validationErr != nil {
		return validationErr
	}
	estimated, err = applyDeploymentPricingOverride(c.Context(), h.billing, estimated, deployment)
	if err != nil {
		return writeBillingEstimateError(c, err)
	}

	if !isVirtual {
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "router.select_model", "internal", routeStart, fiber.StatusOK, map[string]interface{}{"provider": deployment.Provider, "deployment_id": deployment.ID})
	}

	payload, err := buildChatCompletionPayload(req, deployment.Model, resolvedModel)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadRequest, "Failed to encode request", "invalid_request_error")
	}

	tunnelReq := baseReq
	tunnelReq.Body = payload
	tunnelReq.Deployment = toContractDeployment(deployment)
	runtimeModel := strings.TrimSpace(deployment.Model)
	if runtimeModel == "" || runtimeModel == "*" {
		runtimeModel = resolvedModel
	}
	tunnelReq.ModelRuntimeProfiles, err = runtimeProfilesFromEstimate(
		deployment.Provider,
		runtimeModel,
		estimated,
	)
	if err != nil {
		return writeAPIError(
			c,
			fiber.StatusServiceUnavailable,
			err.Error(),
			"runtime_profile_unavailable",
		)
	}
	tunnelReq.RequestMetadata = enrichRequestMetadataForDeployment(tunnelReq.RequestMetadata, tunnelReq.Deployment)

	if !budgetReservationAttempted {
		budgetReservation, budgetResponseHandled, err = reserveBudgetForRequest(c, h.billing, authInfo, estimated, requestID, deploymentBillingMode(tunnelReq.Deployment))
		if err != nil || budgetResponseHandled {
			return err
		}
	}

	if req.Stream {
		releaseBudgetOnExit = false
		return h.handleStream(c, req, originalBody, authInfo, estimated, tunnelReq, budgetReservation, startTime, tenantSlug)
	}

	err = h.handleSync(c, req, originalBody, authInfo, estimated, tunnelReq, budgetReservation, startTime, tenantSlug, cachePending, guardrailSelection)
	if err == nil {
		releaseBudgetOnExit = false
	}

	return err
}

func writeBillingEstimateError(c *fiber.Ctx, err error) error {
	if errors.Is(err, billing.ErrPricingUnavailable) {
		return writeAPIError(
			c,
			fiber.StatusServiceUnavailable,
			"Provider pricing snapshot is unavailable or stale",
			"pricing_unavailable",
		)
	}
	return writeAPIError(
		c,
		fiber.StatusBadGateway,
		"Failed to estimate request cost",
		"billing_error",
	)
}

func (h *ChatCompletionHandler) routeChatCandidates(
	ctx context.Context,
	tenantID string,
	candidates []string,
	deployment *router.Deployment,
) (*router.Deployment, error) {
	if deployment != nil {
		return deployment, nil
	}

	var routeErr error

	for _, candidate := range candidates {
		resolved, err := h.llmRouter.Route(ctx, tenantID, candidate)
		if err == nil {
			return resolved, nil
		}

		routeErr = err
	}

	return nil, routeErr
}

func (h *ChatCompletionHandler) validateChatDeployment(
	c *fiber.Ctx,
	tenantID string,
	requestedModel string,
	req contractsopenai.ChatCompletionRequest,
	deployment *router.Deployment,
	runtimeProfile *platform.ModelRuntimeProfileBinding,
) (bool, error) {
	if deployment == nil {
		emitDenyCounter(h.logger, denyReasonUnknownModel, requestedModel, tenantID)

		return true, writeAPIError(c, fiber.StatusServiceUnavailable, fmt.Sprintf("No available deployment for model: %s", requestedModel), "routing_error")
	}

	provider := deployment.Provider
	if chatRequestHasImage(req) && !supportsChatImageProvider(provider) {
		return true, writeAPIError(c, fiber.StatusBadRequest, fmt.Sprintf("Provider %s does not support image input through GodarAI chat completions", provider), "invalid_request_error")
	}

	if chatRequestHasFile(req) && !supportsChatFileProvider(provider) {
		return true, writeAPIError(c, fiber.StatusBadRequest, fmt.Sprintf("Provider %s does not support file input through GodarAI chat completions", provider), "invalid_request_error")
	}

	if chatRequestHasInputAudio(req) && !supportsChatInputAudioProvider(provider) {
		return true, writeAPIError(c, fiber.StatusBadRequest, fmt.Sprintf("Provider %s does not support input_audio through GodarAI chat completions", provider), "invalid_request_error")
	}

	if chatRequestHasTools(req) && !supportsChatToolsProvider(provider) {
		return true, writeAPIError(c, fiber.StatusBadRequest, fmt.Sprintf("Provider %s does not support tools through GodarAI chat completions", provider), "invalid_request_error")
	}

	if chatRequestHasResponseFormat(req) && !supportsChatResponseFormatProvider(provider) {
		return true, writeAPIError(c, fiber.StatusBadRequest, fmt.Sprintf("Provider %s does not support response_format through GodarAI chat completions", provider), "invalid_request_error")
	}

	if chatRequestHasExtendedThinking(req) && !supportsChatExtendedThinkingProvider(provider) {
		return true, writeAPIError(c, fiber.StatusBadRequest, fmt.Sprintf("Provider %s does not support extended thinking through GodarAI chat completions", provider), "invalid_request_error")
	}

	if err := validateProviderMediaInput(req, provider); err != nil {
		return true, writeAPIError(c, fiber.StatusBadRequest, err.Error(), "invalid_request_error")
	}

	if strings.EqualFold(strings.TrimSpace(provider), "anthropic") {
		if err := validateAnthropicChatRuntimeProfile(
			req,
			runtimeProfile,
		); err != nil {
			return true, writeAPIError(
				c,
				fiber.StatusBadRequest,
				err.Error(),
				"invalid_request_error",
			)
		}
	}
	if isGoogleGeminiProvider(provider) {
		if err := validateGoogleGeminiChatRuntimeProfile(
			req,
			runtimeProfile,
		); err != nil {
			return true, writeAPIError(
				c,
				fiber.StatusBadRequest,
				err.Error(),
				"invalid_request_error",
			)
		}
	}

	return false, nil
}

func buildChatCompletionPayload(req contractsopenai.ChatCompletionRequest, deploymentModel string, fallbackModel string) ([]byte, error) {
	outboundReq := req
	if req.AdditionalFields != nil {
		outboundReq.AdditionalFields = make(map[string]json.RawMessage, len(req.AdditionalFields))
		for key, value := range req.AdditionalFields {
			outboundReq.AdditionalFields[key] = append(json.RawMessage(nil), value...)
		}
	}

	outboundModel := strings.TrimSpace(deploymentModel)
	if outboundModel == "" || outboundModel == "*" {
		outboundModel = fallbackModel
	}

	outboundReq.Model = outboundModel
	outboundReq.PromptVersionFQN = ""
	outboundReq.PromptVariables = nil
	delete(outboundReq.AdditionalFields, "prompt_version_fqn")
	delete(outboundReq.AdditionalFields, "prompt_variables")

	return json.Marshal(outboundReq)
}

func buildChatRequestLogPayload(req contractsopenai.ChatCompletionRequest, originalBody []byte) map[string]interface{} {
	return buildJSONRequestLogPayload(originalBody, map[string]interface{}{
		"model":    req.Model,
		"messages": req.Messages,
	})
}

func validateChatCompletionRequest(c *fiber.Ctx, req contractsopenai.ChatCompletionRequest) (bool, error) {
	if len(req.Messages) == 0 {
		return true, writeAPIError(c, fiber.StatusBadRequest, "messages array must not be empty", "invalid_request_error")
	}

	for i, msg := range req.Messages {
		if msg.Role == "" {
			return true, writeAPIError(c, fiber.StatusBadRequest, fmt.Sprintf("message[%d].role is required", i), "invalid_request_error")
		}

		if strings.EqualFold(strings.TrimSpace(msg.Role), "tool") && strings.TrimSpace(msg.ToolCallID) == "" {
			return true, writeAPIError(c, fiber.StatusBadRequest, fmt.Sprintf("message[%d].tool_call_id is required for tool messages", i), "invalid_request_error")
		}

		if err := contractsopenai.ValidateChatMessageContent(msg.Content); err != nil {
			return true, writeAPIError(c, fiber.StatusBadRequest, fmt.Sprintf("message[%d].content: %v", i, err), "invalid_request_error")
		}
	}

	if req.Temperature != nil && (*req.Temperature < 0 || *req.Temperature > 2) {
		return true, writeAPIError(c, fiber.StatusBadRequest, "temperature must be between 0 and 2", "invalid_request_error")
	}

	if req.TopP != nil && (*req.TopP < 0 || *req.TopP > 1) {
		return true, writeAPIError(c, fiber.StatusBadRequest, "top_p must be between 0 and 1", "invalid_request_error")
	}

	if req.MaxTokens != nil && *req.MaxTokens < 1 {
		return true, writeAPIError(c, fiber.StatusBadRequest, "max_tokens must be at least 1", "invalid_request_error")
	}

	if err := validateReasoningEffort(req.ReasoningEffort); err != nil {
		return true, writeAPIError(c, fiber.StatusBadRequest, err.Error(), "invalid_request_error")
	}

	if err := validateThinkingConfig(req.Thinking); err != nil {
		return true, writeAPIError(c, fiber.StatusBadRequest, err.Error(), "invalid_request_error")
	}

	if err := validateChatResponseFormat(req.ResponseFormat); err != nil {
		return true, writeAPIError(c, fiber.StatusBadRequest, err.Error(), "invalid_request_error")
	}

	return false, nil
}

func (h *ChatCompletionHandler) applyPromptVersion(c *fiber.Ctx, tenantID string, req *contractsopenai.ChatCompletionRequest) error {
	if strings.TrimSpace(req.PromptVersionFQN) == "" {
		return nil
	}

	if h.prompts == nil {
		return writeAPIError(c, fiber.StatusNotImplemented, "Prompt store is not configured", "invalid_request_error")
	}

	promptVersion, err := h.prompts.GetPromptVersionByFQN(c.Context(), tenantID, req.PromptVersionFQN)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadRequest, err.Error(), "invalid_request_error")
	}

	rendered, err := prompts.Render(*promptVersion, req.PromptVariables, req.Messages, req.Model)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadRequest, err.Error(), "invalid_request_error")
	}

	req.Model = rendered.Model
	req.Messages = rendered.Messages
	c.Locals("prompt_metadata", map[string]interface{}{
		"prompt_id":         rendered.PromptID,
		"prompt_version_id": rendered.PromptVersionID,
		"prompt_fqn":        rendered.PromptFQN,
		"prompt_version":    promptVersion.Version,
	})

	return nil
}

func (h *ChatCompletionHandler) handleResponseCacheLookup(
	c *fiber.Ctx,
	req contractsopenai.ChatCompletionRequest,
	originalBody []byte,
	authInfo *platform.AuthContext,
	tenantID string,
	requestedModel string,
	cacheConfig *responsecache.RequestConfig,
	startTime time.Time,
	tenantSlug string,
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
		ModelVersion: h.virtualModelCacheVersion(c.Context(), authInfo, tenantID, requestedModel),
	}
	cacheResult, cachePending, lookupErr := h.responseCache.Lookup(c.Context(), cacheScope, req, *cacheConfig)

	c.Set(apiheaders.HeaderCacheStatus, cacheResult.Status)
	c.Locals("response_cache_status", cacheResult.Status)
	c.Locals("response_cache_type", cacheConfig.Type)
	c.Locals("response_cache_embedding_latency_ms", cacheResult.EmbeddingLatencyMS)

	if lookupErr != nil {
		h.logger.Warn("Response cache lookup failed", zap.String("cache_type", cacheConfig.Type), zap.Error(lookupErr))
	} else if cacheResult.Entry != nil {
		return cachePending, nil, true, h.handleCacheHit(c, req, originalBody, authInfo, cacheResult, startTime, tenantSlug)
	}

	if cacheResult.Status == "bypass" {
		return cachePending, nil, false, nil
	}

	release, waited, acquireErr := h.responseCache.AcquireMiss(c.Context(), cachePending)
	if acquireErr != nil {
		return cachePending, nil, false, writeAPIError(c, fiber.StatusRequestTimeout, "Request canceled while waiting for response cache", "request_canceled")
	}

	if !waited {
		return cachePending, release, false, nil
	}

	secondResult, secondPending, secondErr := h.responseCache.Lookup(c.Context(), cacheScope, req, *cacheConfig)
	cachePending = secondPending

	if secondErr == nil && secondResult.Entry != nil {
		c.Set(apiheaders.HeaderCacheStatus, "hit")
		return cachePending, release, true, h.handleCacheHit(c, req, originalBody, authInfo, secondResult, startTime, tenantSlug)
	}

	return cachePending, release, false, nil
}

func (h *ChatCompletionHandler) tryRouteVirtualModel(
	c *fiber.Ctx,
	tenantID string,
	authInfo *platform.AuthContext,
	requestedModel string,
	req contractsopenai.ChatCompletionRequest,
	resolvedModel string,
	baseReq platform.TunnelRequest,
	routeStart time.Time,
) (*router.Deployment, bool, error) {
	if req.Stream || h.vmEngine == nil || h.vmExec == nil {
		return nil, false, nil
	}

	if snapshotVM, ok := buildVirtualModelFromSnapshot(authInfo, requestedModel); ok {
		deployment, err := h.executeVirtualModelRoute(c, tenantID, snapshotVM, req, resolvedModel, baseReq, routeStart, true)
		if err != nil || deployment != nil {
			return deployment, deployment != nil, err
		}
	}

	if h.vmResolver == nil {
		return nil, false, nil
	}

	vm, resolveErr := h.vmResolver.Resolve(c.Context(), tenantID, requestedModel)
	if resolveErr == nil {
		deployment, err := h.executeVirtualModelRoute(c, tenantID, vm, req, resolvedModel, baseReq, routeStart, false)

		return deployment, deployment != nil, err
	}

	return nil, false, nil
}

func (h *ChatCompletionHandler) executeVirtualModelRoute(
	c *fiber.Ctx,
	tenantID string,
	vm *platform.VirtualModel,
	req contractsopenai.ChatCompletionRequest,
	resolvedModel string,
	baseReq platform.TunnelRequest,
	routeStart time.Time,
	fromSnapshot bool,
) (*router.Deployment, error) {
	targets, selErr := h.vmEngine.SelectTargets(vm)
	if selErr != nil {
		targets = nil
	}

	if targets == nil {
		return nil, nil
	}

	resp, selectedDep, fallbackCount, failureReason, execErr := h.vmExec.ExecuteSync(c.Context(), targets, func(callCtx context.Context, dep *router.Deployment) (*platform.TunnelResponse, error) {
		dep = h.resolveVirtualDeployment(callCtx, tenantID, dep)
		tunnelReq := baseReq
		tunnelReq.Deployment = toContractDeployment(dep)
		tunnelReq.RequestMetadata = enrichRequestMetadataForDeployment(tunnelReq.RequestMetadata, tunnelReq.Deployment)

		payload, payloadErr := buildChatCompletionPayload(req, dep.Model, resolvedModel)
		if payloadErr != nil {
			return nil, payloadErr
		}

		tunnelReq.Body = payload

		return h.tunnel.Proxy(callCtx, tunnelReq)
	})
	if execErr == nil && resp != nil && selectedDep != nil {
		h.recordVirtualModelRoute(c, vm, selectedDep, resp, fallbackCount, failureReason, routeStart, fromSnapshot)
		return selectedDep, nil
	}

	if errors.Is(execErr, virtualmodels.ErrFallbackExhausted) {
		return nil, writeAPIError(c, fiber.StatusBadGateway, "Virtual model fallback exhausted", "fallback_exhausted")
	}

	return nil, nil
}

func (h *ChatCompletionHandler) shouldDeferBudgetForVirtualModel(ctx context.Context, tenantID string, authInfo *platform.AuthContext, requestedModel string) bool {
	vm := h.virtualModelForBudgetDecision(ctx, tenantID, authInfo, requestedModel)
	if vm == nil || h.vmEngine == nil {
		return false
	}

	targets, err := h.vmEngine.SelectTargets(vm)
	if err != nil || len(targets) == 0 {
		return false
	}

	for _, target := range targets {
		dep := h.resolveVirtualDeployment(ctx, tenantID, deploymentFromVirtualTarget(target))
		if dep == nil || deploymentBillingMode(toContractDeployment(dep)) != platform.BillingModeExternalBilled {
			return false
		}
	}

	return true
}

func (h *ChatCompletionHandler) virtualModelForBudgetDecision(ctx context.Context, tenantID string, authInfo *platform.AuthContext, requestedModel string) *platform.VirtualModel {
	if snapshotVM, ok := buildVirtualModelFromSnapshot(authInfo, requestedModel); ok {
		return snapshotVM
	}

	if h.vmResolver == nil {
		return nil
	}

	vm, err := h.vmResolver.Resolve(ctx, tenantID, requestedModel)
	if err != nil {
		return nil
	}

	return vm
}

func deploymentFromVirtualTarget(target platform.VirtualModelTarget) *router.Deployment {
	provider := strings.TrimSpace(target.Provider)

	modelForProvider := strings.TrimSpace(target.Model)
	if p, _, official, ok := splitModelFQN(target.Model); ok {
		if provider == "" {
			provider = p
		}

		modelForProvider = official
	}

	return &router.Deployment{
		ID:        target.ID,
		ModelName: target.Model,
		Provider:  provider,
		Model:     modelForProvider,
		Healthy:   true,
		Weight:    target.Weight,
	}
}

func (h *ChatCompletionHandler) recordVirtualModelRoute(
	c *fiber.Ctx,
	vm *platform.VirtualModel,
	selectedDep *router.Deployment,
	resp *platform.TunnelResponse,
	fallbackCount int,
	failureReason string,
	routeStart time.Time,
	fromSnapshot bool,
) {
	c.Locals("vm_resp", resp)
	c.Locals("vm_fallback_count", fallbackCount)
	c.Locals("vm_failure_reason", failureReason)

	metadata := map[string]interface{}{
		"virtual_model":     vm.Alias,
		"routing_strategy":  vm.RoutingStrategy,
		"selected_provider": selectedDep.Provider,
		"selected_model":    selectedDep.Model,
		"fallback_count":    fallbackCount,
		"failure_reason":    failureReason,
	}
	if fromSnapshot {
		metadata["snapshot"] = true
	}

	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "router.select_virtual_model", "internal", routeStart, fiber.StatusOK, metadata)
}

func (h *ChatCompletionHandler) resolveVirtualDeployment(ctx context.Context, tenantID string, dep *router.Deployment) *router.Deployment {
	if h == nil || h.llmRouter == nil || dep == nil {
		return dep
	}

	modelName := strings.TrimSpace(dep.ModelName)
	if modelName == "" {
		return dep
	}

	resolved, err := h.llmRouter.Route(ctx, tenantID, modelName)
	if err != nil || resolved == nil {
		return dep
	}

	*dep = *resolved

	return dep
}

func (h *ChatCompletionHandler) handleSync(
	c *fiber.Ctx,
	req contractsopenai.ChatCompletionRequest,
	originalBody []byte,
	authInfo *platform.AuthContext,
	estimated platform.UsageCostBreakdown,
	tunnelReq platform.TunnelRequest,
	budgetReservation *platform.BudgetReservation,
	start time.Time,
	tenantSlug string,
	cachePending *responsecache.Pending,
	guardrailSelection guardrails.RequestSelection,
) error {
	providerStart := time.Now()
	providerAttrs := map[string]interface{}{
		"provider":        tunnelReq.Deployment.Provider,
		"request_payload": buildTraceChatRequestPayload(req),
	}

	var resp *platform.TunnelResponse
	if vmResp, ok := c.Locals("vm_resp").(*platform.TunnelResponse); ok && vmResp != nil {
		resp = vmResp
	} else {
		var err error

		resp, err = h.tunnel.Proxy(c.Context(), tunnelReq)
		if err != nil {
			providerAttrs["error"] = err.Error()
			tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, fiber.StatusBadGateway, providerAttrs)
			h.llmRouter.MarkFailed(c.Context(), &router.Deployment{ID: tunnelReq.Deployment.ID}, err.Error())

			return writeAPIError(c, fiber.StatusBadGateway, fmt.Sprintf("Tunnel error: %v", err), "api_error")
		}
	}

	if resp == nil {
		providerAttrs["error"] = "tunnel returned an empty response"
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, fiber.StatusBadGateway, providerAttrs)

		return writeAPIError(c, fiber.StatusBadGateway, "Tunnel returned an empty response", "api_error")
	}

	if resp.StatusCode >= httpErrorFloor {
		providerAttrs["response_payload"] = buildTraceProviderErrorPayload(resp.StatusCode, resp.Body)
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, resp.StatusCode, providerAttrs)
		logProviderFailureRequest(h.logger, h.logs, c, authInfo, tunnelReq, req.Model, "/v1/chat/completions", "chat_completions", resp.StatusCode, resp.Body, start)

		return transformProviderError(c, resp.StatusCode, resp.Body, resp.Headers)
	}

	var completion contractsopenai.ChatCompletionResponse
	if err := json.Unmarshal(resp.Body, &completion); err != nil {
		providerAttrs["error"] = "failed to decode provider response"
		providerAttrs["response_payload"] = buildTraceProviderErrorPayload(fiber.StatusBadGateway, resp.Body)
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, fiber.StatusBadGateway, providerAttrs)

		return writeAPIError(c, fiber.StatusBadGateway, "Failed to decode provider response", "api_error")
	}
	if completion.Usage != nil {
		if tier := responseHeaderValue(resp.Headers, "x-gemini-service-tier"); tier != "" {
			completion.Usage.ServiceTier = tier
		} else if completion.Usage.ServiceTier == "" {
			completion.Usage.ServiceTier = completion.ServiceTier
		}
	}

	providerAttrs["response_payload"] = buildTraceChatResponsePayload(completion)
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, resp.StatusCode, providerAttrs)

	billingModel := billingModelForDeployment(req.Model, tunnelReq.Deployment)

	finalCost, err := h.billing.FinalizeFromUsage(c.Context(), billingModel, completion.Usage, estimated)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, "Failed to finalize usage", "billing_error")
	}

	if debitErr := commitBudgetOrDebit(c.Context(), h.logger, h.billing, h.wallets, budgetReservation, authInfo.TenantID, finalCost, tunnelReq.RequestID, deploymentBillingMode(tunnelReq.Deployment)); debitErr != nil {
		return writeBudgetCheckError(c, debitErr)
	}

	if h.guardrails != nil && len(guardrailSelection.LLMOutput) > 0 {
		guardrailOverrides := map[string]guardrails.Guardrail(nil)
		if overrides, ok := c.Locals("guardrail_overrides").(map[string]guardrails.Guardrail); ok {
			guardrailOverrides = overrides
		}

		outputResult, outputErr := h.guardrails.ApplyOutputWithGuardrails(completion, guardrailSelection, guardrailOverrides)
		h.recordGuardrailChecks(c, outputResult.Checks)

		if outputErr != nil {
			responseTimeMS := int(time.Since(start).Milliseconds())
			logEntry := buildRequestLogEntry(c, authInfo, tunnelReq.Deployment, req.Model, "/v1/chat/completions", "chat_completions", "blocked", fiber.StatusBadRequest, start, responseTimeMS)
			logEntry.RequestID = tunnelReq.RequestID
			logEntry.TraceID = tunnelReq.TraceID
			logEntry.CostUSD = finalCost.TotalCostUSD
			logEntry.CostIRR = finalCost.TotalCostIRR
			logEntry.InputTokens = finalCost.InputTokens
			logEntry.OutputTokens = finalCost.OutputTokens
			logEntry.TotalTokens = finalCost.TotalTokens
			logEntry.RequestBody = buildChatRequestLogPayload(req, originalBody)
			logEntry.ResponseBody = map[string]interface{}{"guardrail_status": "blocked"}
			logEntry.ErrorType = "guardrail_blocked"
			logRequestBestEffort(h.logger, h.logs, c.Context(), logEntry)
			logSpendBestEffort(h.logger, h.logs, c.Context(), platform.SpendLog{
				RequestID:       tunnelReq.RequestID,
				TokenHash:       authInfo.TokenHash,
				TenantID:        authInfo.TenantID,
				OrganizationID:  authInfo.OrganizationID,
				TeamID:          authInfo.TeamID,
				Model:           req.Model,
				InputTokens:     finalCost.InputTokens,
				OutputTokens:    finalCost.OutputTokens,
				TotalTokens:     finalCost.TotalTokens,
				CostUSD:         finalCost.TotalCostUSD,
				CostIRR:         finalCost.TotalCostIRR,
				ResponseTimeMS:  responseTimeMS,
				Status:          "blocked",
				RequestMetadata: enrichPricingMetadata(tunnelReq.RequestMetadata, authInfo, req.Model),
				UsageBreakdown:  &finalCost,
				PricingSnapshot: finalCost.PricingSnapshot,
				CompletionStart: start,
				CompletionEnd:   time.Now(),
			})

			return writeGuardrailError(c, fiber.StatusBadRequest, "Response blocked by GodarAI guardrail", outputResult.Checks)
		}

		if outputResult.Response != nil {
			completion = *outputResult.Response
		}
	}

	if cachePending != nil && h.responseCache != nil {
		cachedBody, marshalErr := json.Marshal(completion)
		if marshalErr == nil {
			if cacheErr := h.responseCache.Store(c.Context(), cachePending, cachedBody, tunnelReq.TraceID, tunnelReq.Deployment.Provider, tunnelReq.Deployment.Model, finalCost, "gateway-pricing-v1"); cacheErr != nil {
				h.logger.Warn("Response cache store failed", zap.Error(cacheErr))
			}
		}
	}

	responseTimeMS := int(time.Since(start).Milliseconds())
	logEntry := buildRequestLogEntry(c, authInfo, tunnelReq.Deployment, req.Model, "/v1/chat/completions", "chat_completions", "success", resp.StatusCode, start, responseTimeMS)
	logEntry.RequestID = tunnelReq.RequestID
	logEntry.TraceID = tunnelReq.TraceID
	logEntry.CostUSD = finalCost.TotalCostUSD
	logEntry.CostIRR = finalCost.TotalCostIRR
	logEntry.InputTokens = finalCost.InputTokens
	logEntry.OutputTokens = finalCost.OutputTokens
	logEntry.TotalTokens = finalCost.TotalTokens
	logEntry.RequestBody = buildChatRequestLogPayload(req, originalBody)
	logEntry.ResponseBody = buildJSONResponseLogPayload(resp.Body, completion)
	logRequestBestEffort(h.logger, h.logs, c.Context(), logEntry)
	logSpendBestEffort(h.logger, h.logs, c.Context(), platform.SpendLog{
		RequestID:       tunnelReq.RequestID,
		TokenHash:       authInfo.TokenHash,
		TenantID:        authInfo.TenantID,
		OrganizationID:  authInfo.OrganizationID,
		TeamID:          authInfo.TeamID,
		Model:           req.Model,
		InputTokens:     finalCost.InputTokens,
		OutputTokens:    finalCost.OutputTokens,
		TotalTokens:     finalCost.TotalTokens,
		CostUSD:         finalCost.TotalCostUSD,
		CostIRR:         finalCost.TotalCostIRR,
		ResponseTimeMS:  responseTimeMS,
		Status:          "success",
		RequestMetadata: enrichPricingMetadata(tunnelReq.RequestMetadata, authInfo, req.Model),
		UsageBreakdown:  &finalCost,
		PricingSnapshot: finalCost.PricingSnapshot,
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
		RequestType:        "chat_completions",
		UserID:             authInfo.UserID,
		VirtualAccountID:   authInfo.KeyID,
		VirtualAccountName: authInfo.KeyAlias,
		TeamIDs:            nonEmptySlice([]string{authInfo.TeamID}),
		TeamNames:          []string{},
		StatusCode:         resp.StatusCode,
		LatencyMS:          responseTimeMS,
		InputTokens:        finalCost.InputTokens,
		OutputTokens:       finalCost.OutputTokens,
		CostUSD:            finalCost.TotalCostUSD,
		CacheMode:          cacheMetricString(c, "response_cache_type"),
		CacheStatus:        cacheMetricString(c, "response_cache_status"),
		Metadata: enrichMetricMetadataForDeployment(enrichResponseCacheMetadata(c, map[string]interface{}{
			"endpoint":    "/v1/chat/completions",
			"key_type":    authInfo.KeyType,
			"team_id":     authInfo.TeamID,
			"org_unit_id": authInfo.OrgUnitID,
			"tenant_slug": tenantSlug,
		}), tunnelReq.Deployment),
	})

	// Forward allowed provider response headers
	forwardProviderHeaders(c, resp.Headers)

	setTraceHeaders(c, tunnelReq.TraceID)
	setCostHeaders(c, finalCost)

	responseStart := time.Now()
	err = c.Status(resp.StatusCode).JSON(completion)
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "response.write", "internal", responseStart, resp.StatusCode, map[string]interface{}{
		"response_payload": buildTraceChatResponsePayload(completion),
	})

	return err
}

func (h *ChatCompletionHandler) handleCacheHit(c *fiber.Ctx, req contractsopenai.ChatCompletionRequest, originalBody []byte, authInfo *platform.AuthContext, result responsecache.Result, start time.Time, tenantSlug string) error {
	entry := result.Entry

	requestID := tracing.TraceIDFromContext(c)
	if requestID == "" {
		requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
	}

	responseTimeMS := int(time.Since(start).Milliseconds())
	metadata := map[string]interface{}{
		"endpoint": "/v1/chat/completions", "cache_status": "hit", "cache_similarity": result.Similarity,
		"cache_type": result.CacheType, "cache_embedding_latency_ms": result.EmbeddingLatencyMS,
		"cached_trace_id": entry.TraceID, "tenant_slug": tenantSlug, "key_type": authInfo.KeyType,
	}
	logEntry := buildRequestLogEntry(c, authInfo, platform.Deployment{Provider: entry.Provider, Model: entry.ProviderModel}, req.Model, "/v1/chat/completions", "chat_completions", "success", fiber.StatusOK, start, responseTimeMS)
	logEntry.RequestID = requestID
	logEntry.TraceID = entry.TraceID
	logEntry.RequestBody = buildChatRequestLogPayload(req, originalBody)
	logEntry.ResponseBody = buildJSONResponseLogPayload(entry.Response, map[string]interface{}{"cache_status": "hit"})
	logEntry.CacheStatus = "hit"
	logRequestBestEffort(h.logger, h.logs, c.Context(), logEntry)
	logSpendBestEffort(h.logger, h.logs, c.Context(), platform.SpendLog{
		RequestID: requestID, TokenHash: authInfo.TokenHash, TenantID: authInfo.TenantID,
		OrganizationID: authInfo.OrganizationID, TeamID: authInfo.TeamID, Model: req.Model,
		CostUSD: 0, CostIRR: 0, ResponseTimeMS: responseTimeMS, Status: "success",
		RequestMetadata: metadata, CompletionStart: start, CompletionEnd: time.Now(),
	})
	h.emitModelMetric(metricsobs.BuildInput{
		TS: time.Now().UTC(), TenantID: authInfo.TenantID, RequestID: requestID, ModelName: req.Model,
		ProviderName: entry.Provider, ProviderModelName: entry.ProviderModel, RequestType: "chat_completions",
		UserID: authInfo.UserID, VirtualAccountID: authInfo.KeyID, VirtualAccountName: authInfo.KeyAlias,
		TeamIDs: nonEmptySlice([]string{authInfo.TeamID}), StatusCode: fiber.StatusOK, LatencyMS: responseTimeMS,
		CostUSD:   0,
		CacheMode: result.CacheType, CacheStatus: "hit", CacheMatch: cacheMatchForResult(result),
		CacheSimilarity: cacheSimilarityForResult(result), CacheOriginRequestID: entry.TraceID,
		AvoidedInputTokens: entry.CostSnapshot.InputTokens, AvoidedOutputTokens: entry.CostSnapshot.OutputTokens,
		AvoidedCostUSD: entry.CostSnapshot.CostUSD, AvoidedCostIRR: entry.CostSnapshot.CostIRR,
		PricingVersion: entry.CostSnapshot.PricingVersion,
		Metadata:       metadata,
	})
	c.Set("Content-Type", fiber.MIMEApplicationJSON)
	c.Set(apiheaders.HeaderCacheStatus, "hit")
	c.Set(apiheaders.HeaderCachedTraceID, entry.TraceID)

	if result.Semantic {
		c.Set(apiheaders.HeaderCacheSimilarity, fmt.Sprintf("%.6f", result.Similarity))
	}

	setTraceHeaders(c, requestID)
	setCostHeaders(c, platform.UsageCostBreakdown{})

	return c.Status(fiber.StatusOK).Send(entry.Response)
}

func (h *ChatCompletionHandler) handleStream(
	c *fiber.Ctx,
	req contractsopenai.ChatCompletionRequest,
	originalBody []byte,
	authInfo *platform.AuthContext,
	estimated platform.UsageCostBreakdown,
	tunnelReq platform.TunnelRequest,
	budgetReservation *platform.BudgetReservation,
	start time.Time,
	tenantSlug string,
) error {
	includeStreamTelemetry := strings.EqualFold(strings.TrimSpace(c.Get(apiheaders.HeaderStreamTelemetry)), "true")
	requestLogBase := buildRequestLogEntry(c, authInfo, tunnelReq.Deployment, req.Model, "/v1/chat/completions", "chat_completions", "failed", 0, start, 0)
	requestLogBase.RequestID = tunnelReq.RequestID
	requestLogBase.TraceID = tunnelReq.TraceID
	requestLogBase.RequestBody = buildChatRequestLogPayload(req, originalBody)

	providerStart := time.Now()

	stream, err := h.tunnel.Stream(c.Context(), tunnelReq)
	if err != nil {
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, fiber.StatusBadGateway, map[string]interface{}{
			"provider":        tunnelReq.Deployment.Provider,
			"request_payload": buildTraceChatRequestPayload(req),
			"stream":          true,
			"error":           err.Error(),
		})

		releaseBudgetBestEffort(c.Context(), h.logger, h.billing, budgetReservation, "stream_open_failed")
		logStreamingRequestBestEffort(h.logger, h.logs, context.Background(), requestLogBase, "failed", fiber.StatusBadGateway, "stream_open_error", map[string]interface{}{
			"stream":  true,
			"outcome": "open_failed",
			"error":   map[string]interface{}{"message": err.Error(), "type": "stream_open_error"},
		}, nil)

		return writeAPIError(c, fiber.StatusBadGateway, fmt.Sprintf("Tunnel error: %v", err), "api_error")
	}

	if statusCode := stream.StatusCode(); statusCode >= httpErrorFloor {
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, statusCode, map[string]interface{}{
			"provider":        tunnelReq.Deployment.Provider,
			"request_payload": buildTraceChatRequestPayload(req),
			"stream":          true,
			"status_code":     statusCode,
		})
		releaseBudgetBestEffort(c.Context(), h.logger, h.billing, budgetReservation, "provider_http_error")

		return writeTunnelStreamHTTPError(c, h.logger, stream, func(statusCode int, body []byte) {
			responsePayload := buildJSONResponseLogPayload(body, map[string]interface{}{"stream": true, "outcome": "provider_http_error"})
			responsePayload["stream"] = true
			responsePayload["outcome"] = "provider_http_error"

			errorType := ErrorTypeMap[statusCode]
			if strings.TrimSpace(errorType) == "" {
				errorType = "api_error"
			}

			logStreamingRequestBestEffort(h.logger, h.logs, context.Background(), requestLogBase, "failed", statusCode, errorType, responsePayload, nil)
		})
	}

	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, fiber.StatusOK, map[string]interface{}{
		"provider":        tunnelReq.Deployment.Provider,
		"request_payload": buildTraceChatRequestPayload(req),
		"stream":          true,
	})

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	setTraceHeaders(c, tunnelReq.TraceID)

	// Forward provider headers from stream
	streamHeaders := stream.Headers()
	forwardProviderHeaders(c, streamHeaders)
	virtualModelName := modelMetricVirtualName(c, req.Model)

	requestCtx := c.Context()
	requestCtx.SetBodyStreamWriter(func(w *bufio.Writer) {
		defer closeStreamBestEffort(h.logger, stream)

		responseCapture := &streamResponseLogCapture{}
		logStreamFailure := func(errorType string, message string) {
			payload := responseCapture.Payload("failed")
			payload["error"] = map[string]interface{}{"message": message, "type": errorType}
			logStreamingRequestBestEffort(h.logger, h.logs, context.Background(), requestLogBase, "failed", fiber.StatusOK, errorType, payload, nil)
		}

		budgetCommitted := false
		finalized := false

		var finalizedCost platform.UsageCostBreakdown

		usageCapture := &streamUsageCapture{
			tier: responseHeaderValue(streamHeaders, "x-gemini-service-tier"),
		}

		defer func() {
			if !budgetCommitted {
				releaseBudgetBestEffort(requestCtx, h.logger, h.billing, budgetReservation, "stream_not_committed")
			}
		}()

		finalizeAndLogSuccess := func() platform.UsageCostBreakdown {
			if finalized {
				return finalizedCost
			}

			finalized = true

			billingModel := billingModelForDeployment(req.Model, tunnelReq.Deployment)

			finalCost, finalizeErr := h.billing.FinalizeFromUsage(requestCtx, billingModel, usageCapture.Usage(), estimated)
			if finalizeErr != nil {
				h.logger.Warn("failed to finalize streaming usage", zap.Error(finalizeErr), zap.String("request_id", tunnelReq.RequestID))

				finalCost = estimated
				finalCost.Model = billingModel
			}

			if debitErr := commitBudgetOrDebit(requestCtx, h.logger, h.billing, h.wallets, budgetReservation, authInfo.TenantID, finalCost, tunnelReq.RequestID, deploymentBillingMode(tunnelReq.Deployment)); debitErr != nil {
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
				Model:           req.Model,
				InputTokens:     finalCost.InputTokens,
				OutputTokens:    finalCost.OutputTokens,
				TotalTokens:     finalCost.TotalTokens,
				CostUSD:         finalCost.TotalCostUSD,
				CostIRR:         finalCost.TotalCostIRR,
				ResponseTimeMS:  int(time.Since(start).Milliseconds()),
				Status:          "success",
				RequestMetadata: enrichPricingMetadata(tunnelReq.RequestMetadata, authInfo, req.Model),
				UsageBreakdown:  &finalCost,
				PricingSnapshot: finalCost.PricingSnapshot,
				CompletionStart: start,
				CompletionEnd:   time.Now(),
			})
			logStreamingRequestBestEffort(h.logger, h.logs, context.Background(), requestLogBase, "success", fiber.StatusOK, "", responseCapture.Payload("completed"), &finalCost)
			h.emitModelMetric(metricsobs.BuildInput{
				TS:                 time.Now().UTC(),
				TenantID:           authInfo.TenantID,
				RequestID:          tunnelReq.RequestID,
				ModelName:          req.Model,
				VirtualModelName:   virtualModelName,
				ProviderName:       tunnelReq.Deployment.Provider,
				ProviderModelName:  tunnelReq.Deployment.Model,
				RequestType:        "chat_completions",
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
					"endpoint":    "/v1/chat/completions",
					"key_type":    authInfo.KeyType,
					"team_id":     authInfo.TeamID,
					"org_unit_id": authInfo.OrgUnitID,
					"tenant_slug": tenantSlug,
				}, tunnelReq.Deployment),
			})
			finalizedCost = finalCost

			return finalizedCost
		}

		for {
			select {
			case chunk, ok := <-stream.Chunks():
				if !ok {
					finalizeAndLogSuccess()
					return
				}

				if len(chunk.Data) > 0 {
					responseCapture.Append(chunk.Data)
					usageCapture.Add(chunk.Data)

					if chunk.IsDone && includeStreamTelemetry && isStreamDoneMarker(chunk.Data) {
						finalCost := finalizeAndLogSuccess()
						if writeErr := writeStreamTelemetryEvent(w, finalCost); writeErr != nil {
							h.logger.Warn("failed to write streaming telemetry", zap.Error(writeErr), zap.String("request_id", tunnelReq.RequestID))
							return
						}

						if writeErr := writeBufferedBytes(w, chunk.Data); writeErr != nil {
							h.logger.Warn("failed to write streaming done chunk", zap.Error(writeErr), zap.String("request_id", tunnelReq.RequestID))
						}

						return
					}

					if writeErr := writeBufferedBytes(w, chunk.Data); writeErr != nil {
						h.logger.Warn("failed to write streaming chunk", zap.Error(writeErr), zap.String("request_id", tunnelReq.RequestID))
						releaseBudgetBestEffort(requestCtx, h.logger, h.billing, budgetReservation, "stream_write_failed")
						logStreamFailure("stream_write_error", writeErr.Error())

						return
					}
				}

				if chunk.IsDone {
					finalCost := finalizeAndLogSuccess()
					if includeStreamTelemetry {
						if writeErr := writeStreamTelemetryEvent(w, finalCost); writeErr != nil {
							h.logger.Warn("failed to write streaming telemetry", zap.Error(writeErr), zap.String("request_id", tunnelReq.RequestID))
						}
					}

					return
				}
			case err, ok := <-stream.Errors():
				if ok && err != nil {
					releaseBudgetBestEffort(requestCtx, h.logger, h.billing, budgetReservation, "stream_error")

					errorPayload, marshalErr := json.Marshal(contractsopenai.APIError{
						Error: contractsopenai.ErrorInfo{
							Message: err.Error(),
							Type:    "streaming_error",
						},
					})
					if marshalErr != nil {
						h.logger.Warn("failed to encode streaming error payload", zap.Error(marshalErr), zap.String("request_id", tunnelReq.RequestID))
						logStreamFailure("streaming_error", err.Error())

						return
					}

					streamErrorEvent := bytes.Join([][]byte{[]byte("data: "), errorPayload, []byte("\n\n")}, nil)
					responseCapture.Append(streamErrorEvent)

					if writeErr := writeBufferedBytes(w, streamErrorEvent); writeErr != nil {
						h.logger.Warn("failed to write streaming error payload", zap.Error(writeErr), zap.String("request_id", tunnelReq.RequestID))
					}

					logStreamFailure("streaming_error", err.Error())
				} else {
					logStreamFailure("stream_closed_without_completion", "provider stream closed without a completion marker")
				}

				return
			}
		}
	})

	responseStart := time.Now()
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "response.write", "internal", responseStart, fiber.StatusOK, map[string]interface{}{"stream": true})

	return nil
}

type streamUsageCapture struct {
	buffer string
	usage  *contractsopenai.Usage
	tier   string
}

func (capture *streamUsageCapture) Add(data []byte) {
	capture.buffer += string(data)
	for {
		frame, rest, ok := takeSSEFrame(capture.buffer)
		if !ok {
			return
		}

		capture.buffer = rest

		payload := streamFrameData(frame)
		if payload == "" || payload == "[DONE]" {
			continue
		}

		var chunk contractsopenai.StreamChunk
		if err := json.Unmarshal([]byte(payload), &chunk); err == nil && chunk.Usage != nil {
			usage := *chunk.Usage
			if capture.tier != "" {
				usage.ServiceTier = capture.tier
			} else if usage.ServiceTier == "" {
				usage.ServiceTier = chunk.ServiceTier
			}
			capture.usage = &usage
		}
	}
}

func (capture *streamUsageCapture) Usage() *contractsopenai.Usage {
	if capture == nil || capture.usage == nil {
		return nil
	}

	usage := *capture.usage

	return &usage
}

func responseHeaderValue(headers map[string]string, name string) string {
	for key, value := range headers {
		if strings.EqualFold(strings.TrimSpace(key), name) {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func takeSSEFrame(buffer string) (string, string, bool) {
	newlineIndex := strings.Index(buffer, "\n\n")
	carriageReturnIndex := strings.Index(buffer, "\r\n\r\n")
	index := newlineIndex
	delimiterLength := 2

	if carriageReturnIndex >= 0 && (index < 0 || carriageReturnIndex < index) {
		index = carriageReturnIndex
		delimiterLength = 4
	}

	if index < 0 {
		return "", buffer, false
	}

	return buffer[:index], buffer[index+delimiterLength:], true
}

func streamFrameData(frame string) string {
	lines := strings.Split(strings.ReplaceAll(frame, "\r\n", "\n"), "\n")

	data := make([]string, 0, len(lines))
	for _, line := range lines {
		if !strings.HasPrefix(line, "data:") {
			continue
		}

		data = append(data, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
	}

	if len(data) > 0 {
		return strings.Join(data, "\n")
	}

	trimmed := strings.TrimSpace(frame)
	if strings.HasPrefix(trimmed, "{") || trimmed == "[DONE]" {
		return trimmed
	}

	return ""
}

func isStreamDoneMarker(data []byte) bool {
	trimmed := strings.TrimSpace(string(data))
	return trimmed == "[DONE]" || trimmed == "data: [DONE]"
}

func writeStreamTelemetryEvent(w *bufio.Writer, cost platform.UsageCostBreakdown) error {
	payload, err := json.Marshal(struct {
		Object  string                `json:"object"`
		Usage   contractsopenai.Usage `json:"usage"`
		CostUSD float64               `json:"cost_usd"`
		CostIRR float64               `json:"cost_irr"`
	}{
		Object: "godarai.telemetry",
		Usage: contractsopenai.Usage{
			PromptTokens:     cost.InputTokens,
			CompletionTokens: cost.OutputTokens,
			TotalTokens:      cost.TotalTokens,
		},
		CostUSD: cost.TotalCostUSD,
		CostIRR: cost.TotalCostIRR,
	})
	if err != nil {
		return err
	}

	return writeBufferedBytes(w, bytes.Join([][]byte{[]byte("event: godarai.telemetry\ndata: "), payload, []byte("\n\n")}, nil))
}

func (h *ChatCompletionHandler) emitModelMetric(input metricsobs.BuildInput) {
	if h.emitMetric == nil {
		return
	}

	h.emitMetric(input)
}

func modelMetricVirtualName(c *fiber.Ctx, fallback string) string {
	if c == nil {
		return ""
	}

	if v, ok := c.Locals("vm_failure_reason").(string); ok && v != "" {
		_ = v
	}

	if v, ok := c.Locals("vm_resp").(*platform.TunnelResponse); ok && v != nil {
		return fallback
	}

	return ""
}

func nonEmptySlice(in []string) []string {
	out := make([]string, 0, len(in))
	for _, v := range in {
		v = strings.TrimSpace(v)
		if v != "" {
			out = append(out, v)
		}
	}

	return out
}

func enrichResponseCacheMetadata(c *fiber.Ctx, metadata map[string]interface{}) map[string]interface{} {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	if c == nil {
		return metadata
	}

	if value := c.Locals("response_cache_status"); value != nil {
		metadata["cache_status"] = value
	}

	if value := c.Locals("response_cache_type"); value != nil {
		metadata["cache_type"] = value
	}

	if value := c.Locals("response_cache_embedding_latency_ms"); value != nil {
		metadata["cache_embedding_latency_ms"] = value
	}

	return metadata
}

func enrichPromptMetadata(c *fiber.Ctx, metadata map[string]interface{}) map[string]interface{} {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	if c == nil {
		return metadata
	}

	if raw, ok := c.Locals("prompt_metadata").(map[string]interface{}); ok {
		for key, value := range raw {
			metadata[key] = value
		}
	}

	return metadata
}

func cacheMetricString(c *fiber.Ctx, key string) string {
	if c == nil {
		return ""
	}

	value := ""
	if localValue, ok := c.Locals(key).(string); ok {
		value = localValue
	}

	return value
}

func cacheMatchForResult(result responsecache.Result) string {
	if result.Semantic {
		return "semantic"
	}

	return "exact"
}

func cacheSimilarityForResult(result responsecache.Result) *float64 {
	if !result.Semantic {
		return nil
	}

	score := result.Similarity

	return &score
}

func billingModelForDeployment(requestedModel string, deployment platform.Deployment) string {
	model := strings.TrimSpace(deployment.Model)
	if model == "" || model == "*" {
		model = providerModelID(requestedModel)
	}

	if model == "" {
		model = strings.TrimSpace(requestedModel)
	}

	provider := strings.TrimSpace(deployment.Provider)
	if provider != "" && !strings.Contains(model, "/") {
		return provider + "/" + model
	}

	return model
}

func (h *ChatCompletionHandler) virtualModelCacheVersion(ctx context.Context, authInfo *platform.AuthContext, tenantID, model string) string {
	if authInfo == nil {
		return ""
	}

	if snapshot, ok := authInfo.VirtualModelSnapshots[model]; ok {
		return fmt.Sprintf("%s:%d", snapshot.VirtualModelID, snapshot.Version)
	}

	if h.vmResolver != nil {
		if virtualModel, err := h.vmResolver.Resolve(ctx, tenantID, model); err == nil && virtualModel != nil {
			return fmt.Sprintf("%s:%d", virtualModel.ID, virtualModel.Version)
		}
	}

	return ""
}

func toContractDeployment(dep *router.Deployment) platform.Deployment {
	legacyAPIKey := dep.Params.APIKey
	if dep.ProviderAuth != nil {
		legacyAPIKey = ""
	}

	return platform.Deployment{
		ID:                  dep.ID,
		TenantID:            dep.TenantID,
		ModelName:           dep.ModelName,
		Provider:            dep.Provider,
		Model:               dep.Model,
		APIBase:             dep.Params.APIBase,
		APIKey:              legacyAPIKey,
		APIVersion:          dep.Params.APIVersion,
		ProjectID:           dep.Params.ProjectID,
		EndpointID:          dep.Params.EndpointID,
		RecognizerID:        dep.Params.RecognizerID,
		ProviderHeaders:     dep.Params.CustomHeaders,
		RegionName:          dep.Params.RegionName,
		RPM:                 dep.Params.RPM,
		TPM:                 dep.Params.TPM,
		Weight:              dep.Weight,
		TimeoutSec:          dep.Params.Timeout,
		CredentialSource:    dep.CredentialSource,
		BillingMode:         dep.BillingMode,
		PricingMode:         dep.PricingMode,
		ProviderAccountID:   dep.ProviderAccountID,
		ProviderAccountName: dep.ProviderAccountName,
		ConnectionID:        dep.ConnectionID,
		CredentialVersionID: dep.CredentialVersionID,
		ProviderAuth:        dep.ProviderAuth,
		EffectivePricing:    dep.EffectivePricing,
	}
}

func logRequestBestEffort(logger *zap.Logger, logs usageLogger, ctx context.Context, entry platform.RequestLogEntry) {
	if err := logs.LogRequest(ctx, entry); err != nil && logger != nil && !errors.Is(err, storage.ErrLogStoreUnavailable) {
		logger.Warn("failed to log request", zap.Error(err), zap.String("request_id", entry.RequestID))
	}
}

func logProviderFailureRequest(
	logger *zap.Logger,
	logs usageLogger,
	c *fiber.Ctx,
	authInfo *platform.AuthContext,
	tunnelReq platform.TunnelRequest,
	requestedModel string,
	route string,
	requestType string,
	statusCode int,
	responseBody []byte,
	start time.Time,
) {
	responseTimeMS := int(time.Since(start).Milliseconds())
	logEntry := buildRequestLogEntry(c, authInfo, tunnelReq.Deployment, requestedModel, route, requestType, "failed", statusCode, start, responseTimeMS)
	logEntry.RequestID = tunnelReq.RequestID
	logEntry.TraceID = tunnelReq.TraceID
	logEntry.RequestBody = map[string]interface{}{"model": requestedModel}
	logEntry.ResponseBody = map[string]interface{}{"error": parseGenericError(responseBody)}

	if errType, ok := ErrorTypeMap[statusCode]; ok && strings.TrimSpace(errType) != "" {
		logEntry.ErrorType = errType
	} else {
		logEntry.ErrorType = "api_error"
	}

	logRequestBestEffort(logger, logs, c.Context(), logEntry)
}

func logSpendBestEffort(logger *zap.Logger, logs usageLogger, ctx context.Context, spendLog platform.SpendLog) {
	if err := logs.LogSpend(ctx, spendLog); err != nil && logger != nil {
		logger.Warn("failed to log spend", zap.Error(err), zap.String("request_id", spendLog.RequestID))
	}
}

func closeStreamBestEffort(logger *zap.Logger, stream interface{ Close() error }) {
	if err := stream.Close(); err != nil && logger != nil {
		logger.Warn("failed to close stream", zap.Error(err))
	}
}

func writeBufferedBytes(w *bufio.Writer, data []byte) error {
	if _, err := w.Write(data); err != nil {
		return err
	}

	return w.Flush()
}

func setRateLimitHeaders(c *fiber.Ctx, result *cache.RateLimitResult) {
	if result == nil {
		return
	}

	remaining := result.Limit - result.Current
	if remaining < 0 {
		remaining = 0
	}

	c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", result.Limit))
	c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
	c.Set("X-RateLimit-Reset", fmt.Sprintf("%d", result.ResetAt.Unix()))

	if result.RetryAfter {
		retryAfter := time.Until(result.ResetAt)
		if retryAfter < 0 {
			retryAfter = 0
		}

		c.Set("Retry-After", fmt.Sprintf("%d", int64(retryAfter.Seconds())))
	}
}

func chatRequestHasImage(req contractsopenai.ChatCompletionRequest) bool {
	for _, msg := range req.Messages {
		if contractsopenai.ChatContentHasImage(msg.Content) {
			return true
		}
	}

	return false
}

func chatRequestHasFile(req contractsopenai.ChatCompletionRequest) bool {
	for _, msg := range req.Messages {
		if contractsopenai.ChatContentHasFile(msg.Content) {
			return true
		}
	}

	return false
}

func chatRequestHasInputAudio(req contractsopenai.ChatCompletionRequest) bool {
	for _, msg := range req.Messages {
		if contractsopenai.ChatContentHasInputAudio(msg.Content) {
			return true
		}
	}
	return false
}

func chatRequestHasTools(req contractsopenai.ChatCompletionRequest) bool {
	return len(req.Tools) > 0 || req.ToolChoice != nil || req.ParallelToolCalls != nil
}

func chatRequestHasResponseFormat(req contractsopenai.ChatCompletionRequest) bool {
	return req.ResponseFormat != nil ||
		(req.OutputConfig != nil && req.OutputConfig["format"] != nil)
}

func chatRequestHasExtendedThinking(req contractsopenai.ChatCompletionRequest) bool {
	effort := strings.ToLower(strings.TrimSpace(req.ReasoningEffort))
	if effort != "" && effort != "none" {
		return true
	}

	if thinkingConfigEnabled(req.Thinking) {
		return true
	}

	for _, msg := range req.Messages {
		if strings.TrimSpace(msg.ReasoningContent) != "" || len(msg.ThinkingBlocks) > 0 {
			return true
		}
	}

	return false
}

func thinkingConfigEnabled(value *contractsopenai.ThinkingConfig) bool {
	if value == nil {
		return false
	}

	thinkingType := strings.ToLower(strings.TrimSpace(value.Type))

	return thinkingType == "enabled" || thinkingType == "adaptive"
}

func supportsChatImageProvider(provider string) bool {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "", "openai", "azure", "groq", "mistral", "mistral-ai", "openrouter", "perplexity", "perplexityai", "sambanova", "togetherai", "together-ai", "xai", "anthropic", "gemini", "google-gemini":
		return true
	default:
		return false
	}
}

func supportsChatToolsProvider(provider string) bool {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "", "openai", "azure", "groq", "mistral", "mistral-ai", "openrouter", "perplexity", "perplexityai", "sambanova", "togetherai", "together-ai", "xai", "gemini", "google-gemini":
		return true
	default:
		return false
	}
}

func supportsChatResponseFormatProvider(provider string) bool {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "", "openai", "azure", "groq", "mistral", "mistral-ai", "openrouter", "perplexity", "perplexityai", "sambanova", "togetherai", "together-ai", "xai", "gemini", "google-gemini":
		return true
	default:
		return false
	}
}

func supportsChatExtendedThinkingProvider(provider string) bool {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "", "openai", "azure", "groq", "mistral", "mistral-ai", "openrouter", "perplexity", "perplexityai", "sambanova", "togetherai", "together-ai", "xai", "anthropic", "gemini", "google-gemini":
		return true
	default:
		return false
	}
}

func supportsChatFileProvider(provider string) bool {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "", "openai", "azure", "openrouter", "anthropic":
		return true
	default:
		return false
	}
}

func supportsChatInputAudioProvider(provider string) bool {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "", "openai", "azure", "gemini", "google-gemini":
		return true
	default:
		return false
	}
}

func validateReasoningEffort(value string) error {
	effort := strings.ToLower(strings.TrimSpace(value))
	if effort == "" {
		return nil
	}

	switch effort {
	case "none", "minimal", "low", "medium", "high", "xhigh", "max":
		return nil
	default:
		return fmt.Errorf("reasoning_effort %q is not supported", value)
	}
}

func validateThinkingConfig(value *contractsopenai.ThinkingConfig) error {
	if value == nil {
		return nil
	}

	thinkingType := strings.ToLower(strings.TrimSpace(value.Type))
	if thinkingType == "" {
		return fmt.Errorf("thinking.type is required")
	}

	switch thinkingType {
	case "enabled", "adaptive", "disabled":
	default:
		return fmt.Errorf("thinking.type %q is not supported", value.Type)
	}

	if value.BudgetTokens < 0 {
		return fmt.Errorf("thinking.budget_tokens must be greater than or equal to 0")
	}

	return nil
}

func validateChatResponseFormat(value interface{}) error {
	if value == nil {
		return nil
	}

	format, ok := value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("response_format must be an object")
	}

	formatType, ok := format["type"].(string)

	formatType = strings.TrimSpace(formatType)
	if !ok || formatType == "" {
		return fmt.Errorf("response_format.type is required")
	}

	switch formatType {
	case "json_object":
		return nil
	case "json_schema":
		rawSchema, ok := format["json_schema"].(map[string]interface{})
		if !ok || rawSchema == nil {
			return fmt.Errorf("response_format.json_schema is required for json_schema")
		}

		if _, ok := rawSchema["schema"].(map[string]interface{}); !ok {
			return fmt.Errorf("response_format.json_schema.schema is required for json_schema")
		}

		return nil
	default:
		return fmt.Errorf("response_format.type %q is not supported", formatType)
	}
}

func validateProviderMediaInput(req contractsopenai.ChatCompletionRequest, provider string) error {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "anthropic", "gemini", "google-gemini":
	default:
		return nil
	}

	for msgIndex, msg := range req.Messages {
		parts, ok, err := contractsopenai.ChatContentParts(msg.Content)
		if err != nil {
			return fmt.Errorf("message[%d].content: %v", msgIndex, err)
		}

		if !ok {
			continue
		}

		for partIndex, part := range parts {
			if part.Type != "image_url" {
				if part.Type != "file" {
					continue
				}

				if strings.TrimSpace(part.File.FileID) != "" {
					return fmt.Errorf("message[%d].content part[%d].file.file_id is not supported for provider %s", msgIndex, partIndex, provider)
				}

				if _, _, err := contractsopenai.ParseFileDataURL(part.File.FileData); err != nil {
					return fmt.Errorf("message[%d].content part[%d].file.file_data: %v", msgIndex, partIndex, err)
				}

				continue
			}

			if _, _, err := contractsopenai.ParseImageDataURL(part.ImageURL.URL); err != nil {
				return fmt.Errorf("message[%d].content part[%d].image_url.url: %v", msgIndex, partIndex, err)
			}
		}
	}

	return nil
}
