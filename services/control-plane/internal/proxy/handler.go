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
	"github.com/Delpak/shahrah/services/control-plane/internal/observability"
	metricsobs "github.com/Delpak/shahrah/services/control-plane/internal/observability/metrics"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/Delpak/shahrah/services/control-plane/internal/prompts"
	"github.com/Delpak/shahrah/services/control-plane/internal/responsecache"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
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
	CheckRateLimit(ctx context.Context, auth *platform.AuthContext, model string) (*cache.RateLimitResult, error)
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
	pipeline *RequestPipeline,
	model string,
	dimension string,
	quantity string,
	estimated platform.UsageCostBreakdown,
) (platform.UsageCostBreakdown, error) {
	return pipeline.priceMeteredUsage(
		ctx,
		model,
		dimension,
		quantity,
		estimated,
	)
}

func applyDeploymentPricingOverride(
	ctx context.Context,
	pipeline *RequestPipeline,
	estimated platform.UsageCostBreakdown,
	deployment *router.Deployment,
) (platform.UsageCostBreakdown, error) {
	return pipeline.applyDeploymentPricingOverride(ctx, estimated, deployment)
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

// ChatCompletionHandler serves chat completion requests.
type ChatCompletionHandler struct {
	llmRouter        *router.Router
	pipeline         *RequestPipeline
	billing          budgeter
	logs             usageLogger
	logger           *zap.Logger
	vmResolver       virtualmodels.VirtualModelResolver
	vmEngine         virtualmodels.RoutingEngine
	vmExec           virtualmodels.FallbackExecutor
	emitMetric       func(input metricsobs.BuildInput)
	responseCache    *responsecache.Service
	guardrails       *guardrails.Engine
	guardrailRepo    guardrailResolver
	prompts          promptResolver
	resources        providerNativeResourceStore
	deployments      responseDeploymentLister
	providerFeatures ProviderFeatureConfig
}

// NewChatCompletionHandlerWithDependencies creates a chat handler from explicit dependencies.
func NewChatCompletionHandlerWithDependencies(deps ChatCompletionDependencies) *ChatCompletionHandler {
	virtualModelStore := deps.VirtualModelStore
	if virtualModelStore == nil {
		virtualModelStore = deps.Base.VirtualModelStore
	}

	aliasCacheTTL := deps.VirtualModelAliasCacheTTL
	if aliasCacheTTL <= 0 {
		aliasCacheTTL = deps.Base.VirtualModelAliasCacheTTL
	}

	negativeCacheTTL := deps.VirtualModelNegativeTTL
	if negativeCacheTTL <= 0 {
		negativeCacheTTL = deps.Base.VirtualModelNegativeTTL
	}

	resolver, engine, exec := newVirtualModelDeps(
		virtualModelStore,
		aliasCacheTTL,
		negativeCacheTTL,
	)

	return &ChatCompletionHandler{
		llmRouter:        deps.Base.Router,
		pipeline:         deps.Base.requestPipeline(),
		billing:          deps.Base.Billing,
		logs:             deps.Base.logger(),
		logger:           deps.Base.Logger,
		vmResolver:       resolver,
		vmEngine:         engine,
		vmExec:           exec,
		emitMetric:       deps.Base.metricEmitter(),
		guardrailRepo:    deps.GuardrailResolver,
		prompts:          deps.PromptResolver,
		guardrails:       deps.Guardrails,
		responseCache:    deps.ResponseCache,
		resources:        deps.Base.ProviderResources,
		deployments:      deps.Base.Deployments,
		providerFeatures: deps.Base.ProviderFeatures,
	}
}

// NewChatCompletionHandler creates a new chat completion handler.
func NewChatCompletionHandler(llmRouter *router.Router, tunnel platform.TunnelTransport, budget budgeter, logs usageLogger, logger *zap.Logger, emitMetric ...func(input metricsobs.BuildInput)) *ChatCompletionHandler {
	var metricEmitter func(input metricsobs.BuildInput)
	if len(emitMetric) > 0 {
		metricEmitter = emitMetric[0]
	}

	return NewChatCompletionHandlerWithDependencies(ChatCompletionDependencies{
		Base: BaseHandlerDependencies{
			Router:              llmRouter,
			Tunnel:              tunnel,
			Billing:             budget,
			BillingCapabilities: compatibilityBillingCapabilities(),
			RequestLogger:       logs,
			SpendLogger:         logs,
			Logger:              logger,
			EmitMetric:          metricEmitter,
		},
	})
}

// Handle executes the explicit chat request pipeline.
func (h *ChatCompletionHandler) Handle(c *fiber.Ctx) error {
	startedAt := time.Now()

	var requestLog *requestLogLifecycle

	if authInfo := auth.GetAuthInfo(c); authInfo != nil {
		if currentTenant := tenant.FromContext(c); currentTenant != nil {
			authInfo.TenantID = currentTenant.ID
			requestLog = newRequestLogLifecycle(h.logger, h.logs, c, authInfo, "/v1/chat/completions", "chat_completions", startedAt)
			requestLog.SetRequestBody(buildJSONRequestLogPayload(append([]byte(nil), c.Body()...), nil))

			defer requestLog.FinalizeFromFiber()
		}
	}

	prepared, err := h.prepareChatRequest(c, startedAt, requestLog)
	if err != nil || prepared.responseHandled {
		return err
	}

	requestLog.SetModel(prepared.requestedModel)
	requestLog.SetRequestBody(buildChatRequestLogPayload(prepared.request, prepared.originalBody))

	admitted, err := h.admitChatRequest(c, prepared, requestLog)
	if admitted.cacheRelease != nil {
		defer admitted.cacheRelease()
	}

	if err != nil || admitted.responseHandled {
		return err
	}

	reservationToRelease := admitted.reservation

	releaseBudgetOnExit := true
	defer func() {
		if releaseBudgetOnExit {
			h.pipeline.Release(c.Context(), reservationToRelease, "request_not_committed")
		}
	}()

	routed, err := h.routeChatRequest(c, prepared, admitted)
	if err != nil || routed.responseHandled {
		return err
	}

	reservationToRelease = routed.reservation
	requestLog.SetDeployment(routed.tunnelRequest.Deployment)
	requestLog.SetRequestIDs(routed.tunnelRequest.RequestID, routed.tunnelRequest.TraceID)
	requestLog.SetRequestBody(buildProviderChatRequestLogPayload(routed.tunnelRequest.Deployment.Provider, prepared.request, prepared.originalBody))

	if prepared.request.Stream {
		releaseBudgetOnExit = false

		requestLog.MarkFinalized()

		return h.handleStream(
			c,
			prepared.request,
			prepared.originalBody,
			prepared.authInfo,
			routed.estimated,
			routed.tunnelRequest,
			routed.reservation,
			prepared.startedAt,
			prepared.tenantSlug,
		)
	}

	err = h.handleSync(
		c,
		prepared.request,
		prepared.originalBody,
		prepared.authInfo,
		routed.estimated,
		routed.tunnelRequest,
		routed.reservation,
		prepared.startedAt,
		prepared.tenantSlug,
		admitted.cachePending,
		prepared.guardrailSelection,
		requestLog,
	)
	if err == nil {
		releaseBudgetOnExit = false
	}

	return err
}

func writeBillingEstimateError(c *fiber.Ctx, err error) error {
	var unsupportedTier *billing.UnsupportedServiceTierError
	if errors.As(err, &unsupportedTier) {
		return writeAPIError(
			c,
			fiber.StatusBadRequest,
			unsupportedTier.PublicMessage(),
			"invalid_request_error",
		)
	}

	if errors.Is(err, billing.ErrPricingModelUnavailable) {
		return writeAPIError(
			c,
			fiber.StatusServiceUnavailable,
			"Provider pricing is unavailable for the requested model",
			"pricing_unavailable",
		)
	}

	if errors.Is(err, billing.ErrPricingRateUnavailable) {
		return writeAPIError(
			c,
			fiber.StatusServiceUnavailable,
			"Provider pricing is incomplete for the requested configuration",
			"pricing_unavailable",
		)
	}

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

func (h *ChatCompletionHandler) validateChatDeploymentRequest(
	c *fiber.Ctx,
	tenantID string,
	requestedModel string,
	req contractsopenai.ChatCompletionRequest,
	deployment *router.Deployment,
) (bool, error) {
	if deployment == nil {
		emitDenyCounter(h.logger, denyReasonUnknownModel, requestedModel, tenantID)

		return true, writeAPIError(c, fiber.StatusServiceUnavailable, fmt.Sprintf("No available deployment for model: %s", requestedModel), "routing_error")
	}

	statusCode, errorType, validationErr := h.chatDeploymentRequestValidationError(req, deployment)
	if validationErr != nil {
		message := validationErr.Error()
		if message != "" {
			message = strings.ToUpper(message[:1]) + message[1:]
		}

		return true, writeAPIError(c, statusCode, message, errorType)
	}

	return false, nil
}

func (h *ChatCompletionHandler) chatDeploymentRequestValidationError(
	req contractsopenai.ChatCompletionRequest,
	deployment *router.Deployment,
) (int, string, error) {
	provider := deployment.Provider
	if isMistralProvider(provider) && !mistralEndpointEnabled(h.providerFeatures, "chat_completions") {
		return fiber.StatusServiceUnavailable, "provider_disabled", errors.New("mistral AI chat completions are disabled")
	}

	if isZAIProvider(provider) && h.providerFeatures.Configured && !h.providerFeatures.ZAIEnabled {
		return fiber.StatusServiceUnavailable, "provider_disabled", errors.New("z.ai chat completions are disabled")
	}

	if err := validateProviderMediaInput(req, provider); err != nil {
		return fiber.StatusBadRequest, "invalid_request_error", err
	}

	if strings.EqualFold(strings.TrimSpace(provider), "xai") &&
		chatRequestDeferred(req) &&
		(h.resources == nil || h.deployments == nil || strings.TrimSpace(deployment.ProviderAccountID) == "") {
		return fiber.StatusServiceUnavailable, "routing_error", errors.New("deferred chat completion routing state is unavailable")
	}

	return 0, "", nil
}

func (h *ChatCompletionHandler) validateChatRuntimeProfile(
	c *fiber.Ctx,
	req contractsopenai.ChatCompletionRequest,
	deployment *router.Deployment,
	runtimeProfile *platform.ModelRuntimeProfileBinding,
) (bool, error) {
	if err := chatRuntimeProfileValidationError(req, deployment, runtimeProfile); err != nil {
		return true, writeAPIError(c, fiber.StatusBadRequest, err.Error(), "invalid_request_error")
	}

	return false, nil
}

func chatRuntimeProfileValidationError(
	req contractsopenai.ChatCompletionRequest,
	deployment *router.Deployment,
	runtimeProfile *platform.ModelRuntimeProfileBinding,
) error {
	if deployment == nil {
		return nil
	}

	provider := deployment.Provider

	if strings.EqualFold(strings.TrimSpace(provider), "anthropic") {
		return validateAnthropicChatRuntimeProfile(
			req,
			runtimeProfile,
		)
	}

	if isGoogleGeminiProvider(provider) {
		return validateGoogleGeminiChatRuntimeProfile(
			req,
			runtimeProfile,
		)
	}

	if strings.EqualFold(strings.TrimSpace(provider), "xai") {
		return validateXAIChatRuntimeProfile(req, runtimeProfile)
	}

	if strings.EqualFold(strings.TrimSpace(provider), "moonshot") {
		return validateMoonshotChatRuntimeProfile(req, runtimeProfile)
	}

	if strings.EqualFold(strings.TrimSpace(provider), "deepseek") {
		return validateDeepSeekChatRuntimeProfile(req, runtimeProfile)
	}

	if isMistralProvider(provider) {
		return validateMistralChatRuntimeProfile(req, runtimeProfile)
	}

	if isZAIProvider(provider) {
		return validateZAIChatRuntimeProfile(req, runtimeProfile)
	}

	return nil
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

func buildProviderChatRequestLogPayload(provider string, req contractsopenai.ChatCompletionRequest, originalBody []byte) map[string]interface{} {
	if isDeepSeekProvider(provider) {
		return buildProviderTraceChatRequestPayload(provider, req)
	}

	return buildChatRequestLogPayload(req, originalBody)
}

func buildProviderChatResponseLogPayload(
	provider string,
	body []byte,
	fallback contractsopenai.ChatCompletionResponse,
) map[string]interface{} {
	if !isDeepSeekProvider(provider) {
		return buildJSONResponseLogPayload(body, fallback)
	}

	completion := fallback
	if completion.Model == "" && len(completion.Choices) == 0 && completion.Usage == nil {
		if err := json.Unmarshal(body, &completion); err != nil {
			return map[string]interface{}{
				"body_bytes": len(body),
				"redacted":   true,
			}
		}
	}

	return buildProviderTraceChatResponsePayload(provider, completion)
}

func buildProviderHTTPErrorLogPayloadForProvider(provider string, statusCode int, body []byte, stream bool) map[string]interface{} {
	if isDeepSeekProvider(provider) {
		return map[string]interface{}{
			"outcome":     "provider_http_error",
			"status_code": statusCode,
			"body_bytes":  len(body),
			"stream":      stream,
			"redacted":    true,
		}
	}

	return buildProviderHTTPErrorLogPayload(body, stream)
}

func buildProviderStreamLogPayload(provider string, status string, payload map[string]interface{}) map[string]interface{} {
	if !isDeepSeekProvider(provider) {
		return payload
	}

	return map[string]interface{}{
		"stream":   true,
		"outcome":  status,
		"redacted": true,
	}
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

	if req.MaxTokens != nil && *req.MaxTokens < 1 {
		return true, writeAPIError(c, fiber.StatusBadRequest, "max_tokens must be at least 1", "invalid_request_error")
	}

	if req.MaxCompletionTokens != nil && *req.MaxCompletionTokens < 1 {
		return true, writeAPIError(c, fiber.StatusBadRequest, "max_completion_tokens must be at least 1", "invalid_request_error")
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
	setPromptRequestMetadata(c, promptRequestMetadata{
		promptID:        rendered.PromptID,
		promptVersionID: rendered.PromptVersionID,
		promptFQN:       rendered.PromptFQN,
		promptVersion:   promptVersion.Version,
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
		ModelVersion: h.virtualModelCacheVersion(c.Context(), authInfo, tenantID, requestedModel),
	}
	cacheResult, cachePending, lookupErr := h.responseCache.Lookup(c.Context(), cacheScope, req, *cacheConfig)

	c.Set(apiheaders.HeaderCacheStatus, cacheResult.Status)
	setResponseCacheRequestMetadata(c, responseCacheRequestMetadata{
		status:             cacheResult.Status,
		cacheType:          cacheConfig.Type,
		embeddingLatencyMS: cacheResult.EmbeddingLatencyMS,
	})

	if lookupErr != nil {
		h.logger.Warn("Response cache lookup failed", zap.String("cache_type", cacheConfig.Type), zap.Error(lookupErr))
	} else if cacheResult.Entry != nil {
		return cachePending, nil, true, h.handleCacheHit(c, req, originalBody, authInfo, cacheResult, startTime, tenantSlug, requestLog)
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
		return cachePending, release, true, h.handleCacheHit(c, req, originalBody, authInfo, secondResult, startTime, tenantSlug, requestLog)
	}

	return cachePending, release, false, nil
}

func (h *ChatCompletionHandler) tryRouteVirtualModel(
	c *fiber.Ctx,
	tenantID string,
	authInfo *platform.AuthContext,
	requestedModel string,
	routeStart time.Time,
) (*router.Deployment, bool, error) {
	if h.vmEngine == nil || h.vmExec == nil {
		return nil, false, nil
	}

	plan, resolveErr := resolveVirtualModelRoutePlan(
		c.Context(),
		authInfo,
		h.vmResolver,
		h.vmEngine,
		tenantID,
		requestedModel,
	)
	if errors.Is(resolveErr, virtualmodels.ErrNotVirtualModel) {
		return nil, false, nil
	}

	if errors.Is(resolveErr, virtualmodels.ErrDisabledVirtualModel) {
		return nil, false, writeAPIError(c, fiber.StatusServiceUnavailable, fmt.Sprintf("Virtual model is disabled: %s", requestedModel), "virtual_model_disabled")
	}

	if resolveErr != nil {
		if errors.Is(resolveErr, virtualmodels.ErrNoEligibleTargets) {
			return nil, false, writeAPIError(c, fiber.StatusServiceUnavailable, fmt.Sprintf("No eligible targets for virtual model: %s", requestedModel), "virtual_model_no_eligible_targets")
		}

		return nil, false, writeAPIError(c, fiber.StatusServiceUnavailable, fmt.Sprintf("Virtual model routing is unavailable for model: %s", requestedModel), "virtual_model_routing_unavailable")
	}

	deployment, err := h.prepareVirtualModelRoute(c, tenantID, plan, routeStart)

	return deployment, deployment != nil, err
}

func (h *ChatCompletionHandler) prepareVirtualModelRoute(
	c *fiber.Ctx,
	tenantID string,
	plan *virtualModelRoutePlan,
	routeStart time.Time,
) (*router.Deployment, error) {
	firstDeployment := resolveFirstVirtualDeployment(c.Context(), tenantID, plan.targets, h.llmRouter)

	if firstDeployment == nil {
		return nil, writeAPIError(c, fiber.StatusServiceUnavailable, fmt.Sprintf("No available target deployment for virtual model: %s", plan.model.Alias), "virtual_model_target_unavailable")
	}

	setVirtualModelRoutePlan(c, plan.model, plan.targets, plan.fromSnapshot, routeStart)

	return firstDeployment, nil
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

func (h *ChatCompletionHandler) preflightVirtualChatTargets(
	c *fiber.Ctx,
	tenantID string,
	request contractsopenai.ChatCompletionRequest,
) (platform.UsageCostBreakdown, string, error) {
	state := existingRequestContext(c)
	if state == nil || len(state.virtualModelTargets) == 0 {
		return platform.UsageCostBreakdown{}, "", virtualmodels.ErrNoEligibleTargets
	}

	preparations := make(map[string]virtualChatTargetPreparation, len(state.virtualModelTargets))

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
		preparation := virtualChatTargetPreparation{}

		resolved, err := resolveVirtualDeploymentStrict(c.Context(), h.llmRouter, tenantID, targetDeployment)
		if err == nil {
			preparation.deployment = resolved
			preparation.estimated, preparation.runtimeProfiles, err = h.estimateVirtualChatTarget(
				c.Context(),
				request,
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

	state.virtualChatTargets = preparations

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

func (h *ChatCompletionHandler) estimateVirtualChatTarget(
	ctx context.Context,
	request contractsopenai.ChatCompletionRequest,
	deployment *router.Deployment,
) (platform.UsageCostBreakdown, []platform.ModelRuntimeProfileBinding, error) {
	if deployment == nil {
		return platform.UsageCostBreakdown{}, nil, errors.New("virtual model target deployment is unavailable")
	}

	if _, _, err := h.chatDeploymentRequestValidationError(request, deployment); err != nil {
		return platform.UsageCostBreakdown{}, nil, err
	}

	estimateRequest := request
	estimateRequest.Model = billingEstimateModel(deployment.ModelName)

	estimated, err := h.billing.EstimateCost(ctx, estimateRequest)
	if err != nil {
		return platform.UsageCostBreakdown{}, nil, err
	}

	estimated.Metadata = request.Metadata

	estimated, err = applyAnthropicChatPricingEstimate(request, estimated)
	if err != nil {
		return platform.UsageCostBreakdown{}, nil, err
	}

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

	var runtimeProfile *platform.ModelRuntimeProfileBinding
	if len(runtimeProfiles) == 1 {
		runtimeProfile = &runtimeProfiles[0]
	}

	if err := chatRuntimeProfileValidationError(request, deployment, runtimeProfile); err != nil {
		return platform.UsageCostBreakdown{}, nil, err
	}

	return estimated, runtimeProfiles, nil
}

func isHigherVirtualModelBudgetEstimate(candidate, current platform.UsageCostBreakdown) bool {
	if candidate.TotalCostIRR != current.TotalCostIRR {
		return candidate.TotalCostIRR > current.TotalCostIRR
	}

	if candidate.TotalCostUSD != current.TotalCostUSD {
		return candidate.TotalCostUSD > current.TotalCostUSD
	}

	return candidate.TotalTokens > current.TotalTokens
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

	deployment := &router.Deployment{
		ID:        target.ID,
		ModelName: target.Model,
		Provider:  provider,
		Model:     modelForProvider,
		Healthy:   true,
		Weight:    target.Weight,
	}
	deployment.Params.Timeout = target.Policy.TimeoutSec

	return deployment
}

func recordVirtualModelSelection(
	c *fiber.Ctx,
	vm *platform.VirtualModel,
	selectedDep *router.Deployment,
	fallbackCount int,
	failureReason string,
	routeStart time.Time,
	fromSnapshot bool,
	stream bool,
) {
	state := requestContext(c)
	state.virtualModel = vm
	state.virtualModelActive = true
	state.fallbackCount = fallbackCount
	state.failureReason = failureReason

	metadata := map[string]interface{}{
		"virtual_model":          vm.Alias,
		"virtual_model_id":       vm.ID,
		"virtual_model_version":  vm.Version,
		"routing_strategy":       vm.RoutingStrategy,
		"selected_provider":      selectedDep.Provider,
		"selected_model":         selectedDep.Model,
		"selected_deployment_id": selectedDep.ID,
		"fallback_count":         fallbackCount,
		"failure_reason":         failureReason,
		"stream":                 stream,
	}
	if fromSnapshot {
		metadata["snapshot"] = true
	}

	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "router.select_virtual_model", "internal", routeStart, fiber.StatusOK, metadata)
}

func (h *ChatCompletionHandler) handleSync(
	c *fiber.Ctx,
	request contractsopenai.ChatCompletionRequest,
	originalBody []byte,
	authInfo *platform.AuthContext,
	estimated platform.UsageCostBreakdown,
	tunnelRequest platform.TunnelRequest,
	reservation *platform.BudgetReservation,
	startedAt time.Time,
	tenantSlug string,
	cachePending *responsecache.Pending,
	guardrailSelection guardrails.RequestSelection,
	requestLog *requestLogLifecycle,
) error {
	result, err := h.executeChatProvider(
		c,
		request,
		originalBody,
		authInfo,
		estimated,
		tunnelRequest,
		reservation,
		startedAt,
		requestLog,
	)
	if err != nil || result.handled {
		return err
	}

	var responseHandled bool

	result.completion, responseHandled, err = h.applyChatOutputGuardrails(
		c,
		request,
		originalBody,
		authInfo,
		result.tunnelRequest,
		guardrailSelection,
		result.completion,
		result.finalCost,
		startedAt,
		requestLog,
	)
	if err != nil || responseHandled {
		return err
	}

	if cachePending != nil && h.responseCache != nil {
		cachedBody, marshalErr := json.Marshal(result.completion)
		if marshalErr == nil {
			cacheErr := h.responseCache.Store(
				c.Context(),
				cachePending,
				cachedBody,
				result.tunnelRequest.TraceID,
				result.tunnelRequest.Deployment.Provider,
				result.tunnelRequest.Deployment.Model,
				result.finalCost,
				"gateway-pricing-v1",
			)
			if cacheErr != nil {
				h.logger.Warn("Response cache store failed", zap.Error(cacheErr))
			}
		}
	}

	h.observeSuccessfulChat(
		c,
		request,
		originalBody,
		authInfo,
		result.tunnelRequest,
		result,
		startedAt,
		tenantSlug,
		requestLog,
	)

	forwardProviderHeaders(c, result.response.Headers)
	setTraceHeaders(c, result.tunnelRequest.TraceID)
	setCostHeaders(c, result.finalCost)

	responseStart := time.Now()
	err = c.Status(result.response.StatusCode).JSON(result.completion)
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "response.write", "internal", responseStart, result.response.StatusCode, map[string]interface{}{
		"response_payload": buildProviderTraceChatResponsePayload(result.tunnelRequest.Deployment.Provider, result.completion),
	})

	return err
}
func (h *ChatCompletionHandler) handleCacheHit(c *fiber.Ctx, req contractsopenai.ChatCompletionRequest, originalBody []byte, authInfo *platform.AuthContext, result responsecache.Result, start time.Time, tenantSlug string, requestLog *requestLogLifecycle) error {
	entry := result.Entry

	h.markVirtualModelCacheHit(c, authInfo, req.Model)

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
	metadata = enrichPlaygroundObservability(c, metadata)
	logEntry := buildRequestLogEntry(c, authInfo, platform.Deployment{Provider: entry.Provider, Model: entry.ProviderModel}, req.Model, "/v1/chat/completions", "chat_completions", "success", fiber.StatusOK, start, responseTimeMS)
	logEntry.RequestID = requestID
	logEntry.TraceID = entry.TraceID
	logEntry.RequestBody = buildProviderChatRequestLogPayload(entry.Provider, req, originalBody)
	logEntry.ResponseBody = buildProviderChatResponseLogPayload(entry.Provider, entry.Response, contractsopenai.ChatCompletionResponse{})
	logEntry.CacheStatus = "hit"
	requestLog.FinalizeEntry(c.Context(), logEntry)
	logSpendBestEffort(h.logger, h.logs, c.Context(), platform.SpendLog{
		RequestID: requestID, TokenHash: authInfo.TokenHash, TenantID: authInfo.TenantID,
		TeamID: authInfo.TeamID, UserID: authInfo.UserID, Model: req.Model,
		CostUSD: 0, CostIRR: 0, ResponseTimeMS: responseTimeMS, Status: "success",
		RequestMetadata: metadata, CompletionStart: start, CompletionEnd: time.Now(),
	})
	h.emitModelMetric(metricsobs.BuildInput{
		TS: time.Now().UTC(), TenantID: authInfo.TenantID, RequestID: requestID, ModelName: req.Model,
		VirtualModelName: modelMetricVirtualName(c, req.Model),
		ProviderName:     entry.Provider, ProviderModelName: entry.ProviderModel, RequestType: "chat_completions",
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

func (h *ChatCompletionHandler) markVirtualModelCacheHit(
	c *fiber.Ctx,
	authInfo *platform.AuthContext,
	requestedModel string,
) {
	markVirtualModelCacheHit(c, authInfo, requestedModel, h.vmResolver)
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
	streamCreditRateCards := creditSurface(c)
	providerStart := time.Now()

	stream, tunnelReq, estimated, err := h.dispatchChatStream(c, authInfo.TenantID, req, estimated, tunnelReq)
	requestLogBase := buildRequestLogEntry(c, authInfo, tunnelReq.Deployment, req.Model, "/v1/chat/completions", "chat_completions", "failed", 0, start, 0)
	requestLogBase.RequestID = tunnelReq.RequestID
	requestLogBase.TraceID = tunnelReq.TraceID
	requestLogBase.RequestBody = buildProviderChatRequestLogPayload(tunnelReq.Deployment.Provider, req, originalBody)

	if err != nil {
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, fiber.StatusBadGateway, enrichPlaygroundObservability(c, map[string]interface{}{
			"provider":        tunnelReq.Deployment.Provider,
			"request_payload": buildProviderTraceChatRequestPayload(tunnelReq.Deployment.Provider, req),
			"stream":          true,
			"error":           err.Error(),
		}))

		h.pipeline.Release(c.Context(), budgetReservation, "stream_open_failed")
		logStreamingRequestBestEffort(h.logger, h.logs, context.Background(), requestLogBase, "failed", fiber.StatusBadGateway, "stream_open_error", map[string]interface{}{
			"stream":  true,
			"outcome": "open_failed",
			"error":   map[string]interface{}{"message": err.Error(), "type": "stream_open_error"},
		}, nil)

		if errors.Is(err, virtualmodels.ErrFallbackExhausted) {
			return writeAPIError(c, fiber.StatusBadGateway, "Virtual model fallback exhausted before streaming began", "fallback_exhausted")
		}

		return writeAPIError(c, fiber.StatusBadGateway, fmt.Sprintf("Tunnel error: %v", err), "api_error")
	}

	if statusCode := stream.StatusCode(); statusCode >= httpErrorFloor {
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, statusCode, enrichPlaygroundObservability(c, map[string]interface{}{
			"provider":        tunnelReq.Deployment.Provider,
			"request_payload": buildProviderTraceChatRequestPayload(tunnelReq.Deployment.Provider, req),
			"stream":          true,
			"status_code":     statusCode,
		}))
		h.pipeline.Release(c.Context(), budgetReservation, "provider_http_error")

		return writeOpenAITunnelStreamHTTPError(c, h.logger, stream, func(statusCode int, body []byte) {
			responsePayload := buildProviderHTTPErrorLogPayloadForProvider(tunnelReq.Deployment.Provider, statusCode, body, true)

			errorType := ErrorTypeMap[statusCode]
			if strings.TrimSpace(errorType) == "" {
				errorType = "api_error"
			}

			logStreamingRequestBestEffort(h.logger, h.logs, context.Background(), requestLogBase, "failed", statusCode, errorType, responsePayload, nil)
		})
	}

	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, fiber.StatusOK, enrichPlaygroundObservability(c, map[string]interface{}{
		"provider":        tunnelReq.Deployment.Provider,
		"request_payload": buildProviderTraceChatRequestPayload(tunnelReq.Deployment.Provider, req),
		"stream":          true,
	}))

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
		budgetCommitted := false
		budgetReleaseReason := "stream_not_committed"
		finalized := false

		var finalizedCost platform.UsageCostBreakdown

		logStreamFailure := func(errorType string, message string) {
			payload := responseCapture.Payload("failed")
			payload["error"] = map[string]interface{}{"message": message, "type": errorType}
			payload = buildProviderStreamLogPayload(tunnelReq.Deployment.Provider, "failed", payload)
			logStreamingRequestBestEffort(h.logger, h.logs, context.Background(), requestLogBase, "failed", fiber.StatusOK, errorType, payload, nil)
		}
		writeStreamFailure := func(errorType string, code string, publicMessage string, logMessage string) {
			budgetReleaseReason = "stream_error"
			if code == "stream_incomplete" {
				budgetReleaseReason = "stream_incomplete"
			}

			errorPayload, marshalErr := json.Marshal(contractsopenai.APIError{
				Error: contractsopenai.ErrorInfo{
					Message: publicMessage,
					Type:    "streaming_error",
					Code:    code,
				},
			})
			if marshalErr != nil {
				h.logger.Warn("failed to encode streaming error payload", zap.Error(marshalErr), zap.String("request_id", tunnelReq.RequestID))
				logStreamFailure(errorType, logMessage)

				return
			}

			streamErrorEvent := bytes.Join([][]byte{[]byte("data: "), errorPayload, []byte("\n\n")}, nil)
			responseCapture.Append(streamErrorEvent)

			if writeErr := writeBufferedBytes(w, streamErrorEvent); writeErr != nil {
				h.logger.Warn("failed to write streaming error payload", zap.Error(writeErr), zap.String("request_id", tunnelReq.RequestID))
			}

			logStreamFailure(errorType, logMessage)
		}

		usageCapture := &streamUsageCapture{
			tier: responseHeaderValue(streamHeaders, "x-gemini-service-tier"),
		}
		completionCapture := &streamCompletionCapture{}
		streamUsageNormalizer := newGoogleGeminiStreamUsageNormalizer(tunnelReq.Deployment.Provider)

		defer func() {
			if !budgetCommitted {
				h.pipeline.Release(requestCtx, budgetReservation, budgetReleaseReason)
			}
		}()

		finalizeAndLogSuccess := func() platform.UsageCostBreakdown {
			if finalized {
				return finalizedCost
			}

			finalized = true

			billingModel := billingModelForDeployment(req.Model, tunnelReq.Deployment)

			capturedUsage := usageCapture.Usage()

			finalCost, finalizeErr := h.billing.FinalizeFromUsage(requestCtx, billingModel, capturedUsage, estimated)
			if finalizeErr != nil {
				h.logger.Warn("failed to finalize streaming usage", zap.Error(finalizeErr), zap.String("request_id", tunnelReq.RequestID))

				finalCost = estimated
				finalCost.Model = billingModel
			}

			tunnelReq.RequestMetadata = withXAIUsageMetadata(tunnelReq.RequestMetadata, tunnelReq.Deployment.Provider, capturedUsage)

			if debitErr := h.pipeline.Settle(requestCtx, SettlementRequest{Reservation: budgetReservation, TenantID: authInfo.TenantID, Actual: finalCost, RequestID: tunnelReq.RequestID, BillingMode: deploymentBillingMode(tunnelReq.Deployment)}); debitErr != nil {
				h.logger.Warn("failed to commit budget for streaming request", zap.Error(debitErr), zap.String("request_id", tunnelReq.RequestID))
			} else {
				budgetCommitted = true
			}

			logSpendBestEffort(h.logger, h.logs, requestCtx, platform.SpendLog{
				RequestID:       tunnelReq.RequestID,
				TokenHash:       authInfo.TokenHash,
				TenantID:        authInfo.TenantID,
				TeamID:          authInfo.TeamID,
				UserID:          authInfo.UserID,
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
			responsePayload := buildProviderStreamLogPayload(tunnelReq.Deployment.Provider, "completed", responseCapture.Payload("completed"))
			logStreamingRequestBestEffort(h.logger, h.logs, context.Background(), requestLogBase, "success", fiber.StatusOK, "", responsePayload, &finalCost)
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
				Metadata: enrichMetricMetadataForDeployment(copyPlaygroundObservability(map[string]interface{}{
					"endpoint":    "/v1/chat/completions",
					"key_type":    authInfo.KeyType,
					"team_id":     authInfo.TeamID,
					"tenant_slug": tenantSlug,
				}, tunnelReq.RequestMetadata), tunnelReq.Deployment),
			})
			finalizedCost = finalCost

			return finalizedCost
		}

		streamChunks := stream.Chunks()
		streamErrors := stream.Errors()

		finishStreamAfterTransportEnd := func() {
			completionCapture.Finish()

			if completionCapture.Completed() {
				finalCost := finalizeAndLogSuccess()
				if includeStreamTelemetry {
					if writeErr := writeStreamTelemetryEvent(w, finalCost, creditsUsed(streamCreditRateCards, finalCost)); writeErr != nil {
						h.logger.Warn("failed to write streaming telemetry", zap.Error(writeErr), zap.String("request_id", tunnelReq.RequestID))
					}
				}

				return
			}

			const logMessage = "provider stream closed without a completion marker"
			writeStreamFailure(
				"stream_closed_without_completion",
				"stream_incomplete",
				"The upstream response stream ended before completion.",
				logMessage,
			)
		}

		for streamChunks != nil {
			select {
			case chunk, ok := <-streamChunks:
				if !ok {
					pending := streamUsageNormalizer.Flush()
					if len(pending) > 0 {
						responseCapture.Append(pending)
						usageCapture.Add(pending)
						completionCapture.Add(pending)

						if writeErr := writeBufferedBytes(w, pending); writeErr != nil {
							h.logger.Warn("failed to write pending streaming data", zap.Error(writeErr), zap.String("request_id", tunnelReq.RequestID))

							budgetReleaseReason = "stream_write_failed"

							logStreamFailure("stream_write_error", writeErr.Error())

							return
						}
					}

					if streamErrors != nil {
						select {
						case streamErr, errorsOpen := <-streamErrors:
							if errorsOpen && streamErr != nil {
								writeStreamFailure(
									"streaming_error",
									"stream_interrupted",
									"The upstream response stream was interrupted.",
									streamErr.Error(),
								)

								return
							}
						default:
						}
					}

					finishStreamAfterTransportEnd()

					return
				}

				if len(chunk.Data) > 0 {
					streamData, repairs := streamUsageNormalizer.Add(chunk.Data)
					for _, repair := range repairs {
						logGoogleGeminiUsageRepair(h.logger, tunnelReq.RequestID, repair)
					}

					if chunk.IsDone {
						streamData = append(streamData, streamUsageNormalizer.Flush()...)
					}

					if len(streamData) == 0 && !chunk.IsDone {
						continue
					}

					responseCapture.Append(streamData)
					usageCapture.Add(streamData)
					completionCapture.Add(streamData)

					if chunk.IsDone && includeStreamTelemetry && isStreamDoneMarker(streamData) {
						finalCost := finalizeAndLogSuccess()
						if writeErr := writeStreamTelemetryEvent(w, finalCost, creditsUsed(streamCreditRateCards, finalCost)); writeErr != nil {
							h.logger.Warn("failed to write streaming telemetry", zap.Error(writeErr), zap.String("request_id", tunnelReq.RequestID))
							return
						}

						if writeErr := writeBufferedBytes(w, streamData); writeErr != nil {
							h.logger.Warn("failed to write streaming done chunk", zap.Error(writeErr), zap.String("request_id", tunnelReq.RequestID))
						}

						return
					}

					if writeErr := writeBufferedBytes(w, streamData); writeErr != nil {
						h.logger.Warn("failed to write streaming chunk", zap.Error(writeErr), zap.String("request_id", tunnelReq.RequestID))

						budgetReleaseReason = "stream_write_failed"

						logStreamFailure("stream_write_error", writeErr.Error())

						return
					}
				}

				if chunk.IsDone {
					finishStreamAfterTransportEnd()

					return
				}
			case streamErr, ok := <-streamErrors:
				if !ok {
					streamErrors = nil

					continue
				}

				if streamErr == nil {
					continue
				}

				completionCapture.Finish()

				if completionCapture.Completed() {
					finishStreamAfterTransportEnd()

					return
				}

				writeStreamFailure(
					"streaming_error",
					"stream_interrupted",
					"The upstream response stream was interrupted.",
					streamErr.Error(),
				)

				return
			}
		}
	})

	responseStart := time.Now()
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "response.write", "internal", responseStart, fiber.StatusOK, map[string]interface{}{"stream": true})

	return nil
}

func (h *ChatCompletionHandler) dispatchChatStream(
	c *fiber.Ctx,
	tenantID string,
	req contractsopenai.ChatCompletionRequest,
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
			candidateRequest, candidateEstimate, prepareErr := h.prepareVirtualChatRequest(
				callContext,
				c,
				tenantID,
				req,
				baseRequest,
				deployment,
			)
			if prepareErr != nil {
				return nil, virtualmodels.WithoutRetry(prepareErr)
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

		return nil, selectedRequest, selectedEstimate, errors.New("virtual model stream selected no deployment")
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

func (h *ChatCompletionHandler) prepareVirtualChatRequest(
	ctx context.Context,
	c *fiber.Ctx,
	tenantID string,
	req contractsopenai.ChatCompletionRequest,
	baseRequest platform.TunnelRequest,
	deployment *router.Deployment,
) (platform.TunnelRequest, platform.UsageCostBreakdown, error) {
	targetKey := virtualModelTargetKey(deployment)

	var preparation virtualChatTargetPreparation

	state := existingRequestContext(c)
	if state != nil && state.virtualChatTargets != nil {
		preparation = state.virtualChatTargets[targetKey]
	}

	if preparation.err != nil {
		return platform.TunnelRequest{}, platform.UsageCostBreakdown{}, preparation.err
	}

	resolved := preparation.deployment
	if resolved == nil {
		var err error

		resolved, err = resolveVirtualDeploymentStrict(ctx, h.llmRouter, tenantID, deployment)
		if err != nil {
			return platform.TunnelRequest{}, platform.UsageCostBreakdown{}, err
		}

		preparation.estimated, preparation.runtimeProfiles, err = h.estimateVirtualChatTarget(ctx, req, resolved)
		if err != nil {
			return platform.TunnelRequest{}, platform.UsageCostBreakdown{}, err
		}
	}

	*deployment = *resolved

	if _, _, validationErr := h.chatDeploymentRequestValidationError(req, resolved); validationErr != nil {
		return platform.TunnelRequest{}, platform.UsageCostBreakdown{}, validationErr
	}

	candidateRequest := baseRequest
	candidateRequest.Deployment = toContractDeployment(resolved)
	candidateRequest.Headers = cloneStringMap(baseRequest.Headers)
	candidateRequest.RequestMetadata = replaceDeploymentRequestMetadata(
		baseRequest.RequestMetadata,
		candidateRequest.Deployment,
	)

	payload, err := buildChatCompletionPayload(req, resolved.Model, providerModelID(resolved.ModelName))
	if err != nil {
		return platform.TunnelRequest{}, platform.UsageCostBreakdown{}, fmt.Errorf("build virtual model chat target payload: %w", err)
	}

	deleteHeaderFold(candidateRequest.Headers, "x-grok-conv-id")

	if strings.EqualFold(strings.TrimSpace(resolved.Provider), "xai") {
		payload, err = rewriteXAIFileReferences(ctx, h.resources, tenantID, resolved.ProviderAccountID, payload)
		if err != nil {
			return platform.TunnelRequest{}, platform.UsageCostBreakdown{}, err
		}

		if conversationID := strings.TrimSpace(c.Get("x-grok-conv-id")); conversationID != "" {
			candidateRequest.Headers["x-grok-conv-id"] = conversationID
		}
	}

	candidateRequest.Body = payload

	candidateEstimate := preparation.estimated
	candidateRequest.ModelRuntimeProfiles = append(
		[]platform.ModelRuntimeProfileBinding(nil),
		preparation.runtimeProfiles...,
	)

	var runtimeProfile *platform.ModelRuntimeProfileBinding
	if len(candidateRequest.ModelRuntimeProfiles) == 1 {
		runtimeProfile = &candidateRequest.ModelRuntimeProfiles[0]
	}

	if err := chatRuntimeProfileValidationError(req, resolved, runtimeProfile); err != nil {
		return platform.TunnelRequest{}, platform.UsageCostBreakdown{}, err
	}

	return candidateRequest, candidateEstimate, nil
}

type streamUsageCapture struct {
	buffer []byte
	usage  *contractsopenai.Usage
	tier   string
}

type streamCompletionCapture struct {
	buffer    []byte
	completed bool
}

func (capture *streamCompletionCapture) Add(data []byte) {
	if capture == nil || capture.completed || len(data) == 0 {
		return
	}

	capture.buffer = append(capture.buffer, data...)
	consumed := 0

	for {
		frame, frameBytes, ok := takeSSEFrameBytes(capture.buffer[consumed:])
		if !ok {
			break
		}

		consumed += frameBytes

		if bytes.Equal(streamFrameDataBytes(frame), []byte("[DONE]")) {
			capture.completed = true

			break
		}
	}

	if consumed > 0 {
		remaining := copy(capture.buffer, capture.buffer[consumed:])
		capture.buffer = capture.buffer[:remaining]
	}
}

func (capture *streamCompletionCapture) Finish() {
	if capture == nil || capture.completed || len(capture.buffer) == 0 {
		return
	}

	payload := streamFrameDataBytes(capture.buffer)
	capture.completed = bytes.Equal(payload, []byte("[DONE]"))
	capture.buffer = capture.buffer[:0]
}

func (capture *streamCompletionCapture) Completed() bool {
	return capture != nil && capture.completed
}

func (capture *streamUsageCapture) Add(data []byte) {
	capture.buffer = append(capture.buffer, data...)
	consumed := 0

	for {
		frame, frameBytes, ok := takeSSEFrameBytes(capture.buffer[consumed:])
		if !ok {
			break
		}

		consumed += frameBytes

		payload := streamFrameDataBytes(frame)
		if len(payload) == 0 || bytes.Equal(payload, []byte("[DONE]")) {
			continue
		}

		var chunk contractsopenai.StreamChunk
		if err := json.Unmarshal(payload, &chunk); err == nil && chunk.Usage != nil {
			usage := *chunk.Usage
			if capture.tier != "" {
				usage.ServiceTier = capture.tier
			} else if usage.ServiceTier == "" {
				usage.ServiceTier = chunk.ServiceTier
			}

			capture.usage = &usage
		}
	}

	if consumed > 0 {
		remaining := copy(capture.buffer, capture.buffer[consumed:])
		capture.buffer = capture.buffer[:remaining]
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

func takeSSEFrameBytes(buffer []byte) ([]byte, int, bool) {
	newlineIndex := bytes.Index(buffer, []byte("\n\n"))
	carriageReturnIndex := bytes.Index(buffer, []byte("\r\n\r\n"))
	index := newlineIndex
	delimiterLength := 2

	if carriageReturnIndex >= 0 && (index < 0 || carriageReturnIndex < index) {
		index = carriageReturnIndex
		delimiterLength = 4
	}

	if index < 0 {
		return nil, 0, false
	}

	return buffer[:index], index + delimiterLength, true
}

func streamFrameDataBytes(frame []byte) []byte {
	lines := bytes.Split(frame, []byte{'\n'})
	data := make([][]byte, 0, len(lines))
	totalBytes := 0

	for _, line := range lines {
		line = bytes.TrimSpace(line)
		if !bytes.HasPrefix(line, []byte("data:")) {
			continue
		}

		value := bytes.TrimSpace(bytes.TrimPrefix(line, []byte("data:")))
		data = append(data, value)
		totalBytes += len(value)
	}

	if len(data) == 1 {
		return data[0]
	}

	if len(data) > 1 {
		joined := make([]byte, 0, totalBytes+len(data)-1)
		for index, value := range data {
			if index > 0 {
				joined = append(joined, '\n')
			}

			joined = append(joined, value...)
		}

		return joined
	}

	trimmed := bytes.TrimSpace(frame)
	if bytes.HasPrefix(trimmed, []byte{'{'}) || bytes.Equal(trimmed, []byte("[DONE]")) {
		return trimmed
	}

	return nil
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

func writeStreamTelemetryEvent(w *bufio.Writer, cost platform.UsageCostBreakdown, creditsText string) error {
	payload, err := json.Marshal(struct {
		Object  string                `json:"object"`
		Usage   contractsopenai.Usage `json:"usage"`
		CostUSD float64               `json:"cost_usd"`
		Credits string                `json:"credits,omitempty"`
	}{
		Object: "godarai.telemetry",
		Usage: contractsopenai.Usage{
			PromptTokens:     cost.InputTokens,
			CompletionTokens: cost.OutputTokens,
			TotalTokens:      cost.TotalTokens,
		},
		CostUSD: cost.TotalCostUSD,
		Credits: creditsText,
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

	if state := existingRequestContext(c); state != nil && state.virtualModelActive {
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

	if state := existingRequestContext(c); state != nil && state.responseCacheSet {
		metadata["cache_status"] = state.responseCache.status
		metadata["cache_type"] = state.responseCache.cacheType
		metadata["cache_embedding_latency_ms"] = state.responseCache.embeddingLatencyMS
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

	if state := existingRequestContext(c); state != nil && state.promptSet {
		metadata["prompt_id"] = state.prompt.promptID
		metadata["prompt_version_id"] = state.prompt.promptVersionID
		metadata["prompt_fqn"] = state.prompt.promptFQN
		metadata["prompt_version"] = state.prompt.promptVersion
	}

	return metadata
}

func responseCacheStatus(c *fiber.Ctx) string {
	state := existingRequestContext(c)
	if state == nil {
		return ""
	}

	return state.responseCache.status
}

func responseCacheType(c *fiber.Ctx) string {
	state := existingRequestContext(c)
	if state == nil {
		return ""
	}

	return state.responseCache.cacheType
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
	return virtualModelCacheVersion(ctx, authInfo, tenantID, model, h.vmResolver)
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
	if sink, ok := logs.(interface {
		SupportsLazyRequestLogPayload() bool
	}); !ok || !sink.SupportsLazyRequestLogPayload() {
		entry.RequestBody = materializeLazyLogPayload(entry.RequestBody)
		entry.ResponseBody = materializeLazyLogPayload(entry.ResponseBody)
	}

	if err := logs.LogRequest(ctx, entry); err != nil && logger != nil && !errors.Is(err, observability.ErrLogStoreUnavailable) {
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
	requestLogs ...*requestLogLifecycle,
) {
	responseTimeMS := int(time.Since(start).Milliseconds())
	logEntry := buildRequestLogEntry(c, authInfo, tunnelReq.Deployment, requestedModel, route, requestType, "failed", statusCode, start, responseTimeMS)
	logEntry.RequestID = tunnelReq.RequestID
	logEntry.TraceID = tunnelReq.TraceID
	logEntry.RequestBody = map[string]interface{}{"model": requestedModel}
	logEntry.ResponseBody = buildProviderHTTPErrorLogPayloadForProvider(tunnelReq.Deployment.Provider, statusCode, responseBody, false)

	if errType, ok := ErrorTypeMap[statusCode]; ok && strings.TrimSpace(errType) != "" {
		logEntry.ErrorType = errType
	} else {
		logEntry.ErrorType = "api_error"
	}

	if len(requestLogs) > 0 && requestLogs[0] != nil {
		requestLogs[0].FinalizeEntry(c.Context(), logEntry)

		return
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

func chatRequestHasVideo(req contractsopenai.ChatCompletionRequest) bool {
	for _, msg := range req.Messages {
		if contractsopenai.ChatContentHasVideo(msg.Content) {
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
	case "text", "json_object":
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
