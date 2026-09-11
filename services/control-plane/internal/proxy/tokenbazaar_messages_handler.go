// Package proxy hosts OpenAI-compatible request handlers for the control plane.
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
	metricsobs "github.com/Delpak/shahrah/services/control-plane/internal/observability/metrics"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// TokenBazaarMessagesHandler serves POST /v1/messages requests that target a
// TokenBazaar deployment (an explicit "tokenbazaar:<account>:<model-id>" FQN
// in the request body's "model" field).
//
// This handler exists as a narrow, TokenBazaar-scoped sibling to
// AnthropicNativeHandler rather than as an extension of it. See
// plans/tokenbazaar-production-readiness.md, "Anthropic Messages endpoint
// delivery", for the full architectural reasoning. In short:
//
//   - AnthropicNativeHandler is built around a single provider: it filters
//     deployments with strings.EqualFold(deployment.Provider, "anthropic"),
//     carries an account allowlist scoped to AnthropicAccountIDs, and bundles
//     Anthropic-only concerns (WIF credentials, native file/batch mappings,
//     filesEnabled/batchesEnabled flags) that have no TokenBazaar equivalent
//     and must never be exercised for a reseller.
//   - Investigation confirmed POST /v1/messages is wired, in
//     services/control-plane/internal/server/routes_proxy.go, exclusively to
//     AnthropicNativeHandler.Handle -- there is no pre-existing
//     provider-agnostic handler for platform.ProviderOperationAnthropicMessages
//     the way there is for chat_completions (ChatCompletionHandler),
//     responses (ResponsesHandler), images (ImagesHandler), and embeddings
//     (EmbeddingsHandler).
//   - Rather than widen AnthropicNativeHandler's provider filter (which would
//     put TokenBazaar traffic inside code paths carrying WIF credential
//     issuance, an account allowlist, and files/batches machinery that do not
//     apply to it), this handler is a new, narrow sibling that reuses the
//     standard FQN -> deployment (router.Router) -> providerregistry
//     capability -> Silk Road adapter flow already used for TokenBazaar's
//     other four capabilities (chat, images, embeddings, responses). A thin
//     routing shim (NewAnthropicMessagesRouter, wired in routes_proxy.go)
//     inspects the request body's "model" field before AnthropicNativeHandler
//     is ever invoked: a tokenbazaar-FQN'd model is dispatched here instead.
//     AnthropicNativeHandler's dispatch/routing logic, account allowlist, and
//     WIF/files/batches machinery are completely untouched by this delivery.
//     The one exception is a pure, behavior-preserving extract-method
//     refactor: AnthropicNativeHandler.proxyStream's body was moved into the
//     new shared function streamAnthropicShapedResponse (below), which this
//     handler also calls, so the streaming-relay logic exists in exactly one
//     place instead of two near-identical copies (required by this repo's
//     `dupl` lint rule once a second, structurally-identical Anthropic SSE
//     relay existed). AnthropicNativeHandler.proxyStream itself is now a
//     one-line wrapper with identical behavior to before.
//
// TokenBazaar's own documented /v1/messages contract is the Anthropic
// Messages envelope verbatim (model, messages, max_tokens required, system,
// stream, thinking) -- not an OpenAI-compatible shape. This handler therefore
// treats the inbound request body as already Anthropic-native JSON and
// forwards it unchanged (after rewriting only the "model" field to the
// deployment's provider-facing model id), exactly as
// AnthropicNativeHandler.Handle does for a direct Anthropic dispatch. Byte-
// level request/response translation, streaming SSE parsing, and usage
// extraction are reused, not reimplemented, from anthropic_native_request.go
// and anthropic_native_handler.go (parseAnthropicNativeUsage,
// parseAnthropicSSEUsage, anthropicSSEIsError, anthropicMaximumOutputTokens,
// writeAnthropicError, closeStreamBestEffort, writeTunnelStreamHTTPError) --
// all package-level, provider-agnostic functions already defined for the
// direct Anthropic path.
type TokenBazaarMessagesHandler struct {
	llmRouter  *router.Router
	pipeline   *RequestPipeline
	billing    budgeter
	logs       usageLogger
	logger     *zap.Logger
	emitMetric func(input metricsobs.BuildInput)
}

// NewTokenBazaarMessagesHandlerWithDependencies creates the handler from
// explicit dependencies, mirroring every other *HandlerWithDependencies
// constructor in this package.
func NewTokenBazaarMessagesHandlerWithDependencies(deps BaseHandlerDependencies) *TokenBazaarMessagesHandler {
	return &TokenBazaarMessagesHandler{
		llmRouter:  deps.Router,
		pipeline:   deps.requestPipeline(),
		billing:    deps.Billing,
		logs:       deps.logger(),
		logger:     deps.Logger,
		emitMetric: deps.metricEmitter(),
	}
}

// requestTargetsTokenBazaarMessages reports whether a raw /v1/messages
// request body's "model" field is an explicit tokenbazaar:<account>:<model>
// FQN. This is the sole signal used to decide, before AnthropicNativeHandler
// is ever invoked, whether a request belongs to this handler instead -- a
// bare model id (e.g. "claude-sonnet-5") is never routed here, exactly like
// every other TokenBazaar capability's "explicit FQN only" rule.
func requestTargetsTokenBazaarMessages(body []byte) bool {
	var request struct {
		Model string `json:"model"`
	}
	if json.Unmarshal(body, &request) != nil {
		return false
	}

	provider, _, _, ok := splitModelFQN(strings.TrimSpace(request.Model))

	return ok && isTokenBazaarProvider(provider)
}

// NewAnthropicMessagesRouter returns the fiber.Handler registered for
// POST /v1/messages. It dispatches a tokenbazaar-FQN'd request to
// tokenBazaar (this package's narrow, non-account-allowlisted path) and
// falls back to native (AnthropicNativeHandler.Handle, unmodified) for every
// other request, preserving the direct Anthropic provider's existing
// behavior exactly.
func NewAnthropicMessagesRouter(tokenBazaar *TokenBazaarMessagesHandler, native fiber.Handler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if tokenBazaar != nil && requestTargetsTokenBazaarMessages(c.Body()) {
			return tokenBazaar.Handle(c)
		}

		if native == nil {
			return fiber.ErrNotFound
		}

		return native(c)
	}
}

// supportsAnthropicMessagesProvider reports whether a routed deployment's
// provider may serve platform.ProviderOperationAnthropicMessages through this
// generic handler. Only "tokenbazaar" is listed here: the direct "anthropic"
// provider continues to be served exclusively by AnthropicNativeHandler and
// must never reach this generic FQN-routed path (its deployments are never
// resolved via a bare-model FQN in the first place -- see
// providerregistry.go's tokenbazaar registry comment -- but this check is
// kept as an explicit, fail-closed defense in depth).
func supportsAnthropicMessagesProvider(provider string) bool {
	return isTokenBazaarProvider(provider)
}

// Handle processes a TokenBazaar-targeted /v1/messages request.
func (h *TokenBazaarMessagesHandler) Handle(c *fiber.Ctx) error {
	startTime := time.Now()

	authInfo := auth.GetAuthInfo(c)
	if authInfo == nil {
		return writeAnthropicError(c, fiber.StatusUnauthorized, "authentication_error", "Authentication required")
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return writeAnthropicError(c, fiber.StatusBadRequest, "invalid_request_error", "Tenant resolution required")
	}

	authInfo.TenantID = currentTenant.ID
	tenantSlug := strings.TrimSpace(currentTenant.Slug)

	originalBody := append([]byte(nil), c.Body()...)

	requestLog := newRequestLogLifecycle(h.logger, h.logs, c, authInfo, "/v1/messages", "messages", startTime)
	requestLog.SetRequestBody(buildJSONRequestLogPayload(originalBody, nil))

	defer requestLog.FinalizeFromFiber()

	fields, err := parseTokenBazaarMessagesRequest(originalBody)
	if err != nil {
		return writeAnthropicError(c, fiber.StatusBadRequest, "invalid_request_error", err.Error())
	}

	requestedModel := normalizeModelIdentifier(fields.Model)
	if requestedModel == "" {
		return writeAnthropicError(c, fiber.StatusBadRequest, "invalid_request_error", "model is required")
	}

	requestLog.SetModel(requestedModel)

	modelAccess := modelAccessForType(authInfo, requestedModel, platform.ModelTypeMessages)
	if !modelAccess.Allowed {
		denyReason := denyReasonForModel(authInfo)
		emitDenyCounter(h.logger, denyReason, requestedModel, currentTenant.ID)
		requestLog.MarkFinalized()

		return writeAnthropicError(c, fiber.StatusForbidden, "permission_error", "Model access denied or unsupported for messages")
	}

	estimateRequest := contractsopenai.ChatCompletionRequest{
		Model:    billingEstimateModel(requestedModel),
		Messages: []contractsopenai.ChatMessage{{Role: "user", Content: string(originalBody)}},
	}
	if maxTokens := anthropicMaximumOutputTokens(originalBody, platform.ProviderOperationAnthropicMessages); maxTokens > 0 {
		estimateRequest.MaxTokens = &maxTokens
	}

	estimated, err := h.billing.EstimateCost(c.Context(), estimateRequest)
	if err != nil {
		return writeAnthropicBillingError(c, err)
	}

	requestID := fmt.Sprintf("req_%d", time.Now().UnixNano())

	if tunnelErr := h.pipeline.EnsureReady(c.Context()); tunnelErr != nil {
		return writeAnthropicError(c, fiber.StatusServiceUnavailable, "api_error", "Tunnel path unavailable")
	}

	dep, err := h.llmRouter.Route(c.Context(), currentTenant.ID, requestedModel)
	if err != nil || dep == nil {
		emitDenyCounter(h.logger, denyReasonUnknownModel, requestedModel, currentTenant.ID)

		return writeAnthropicError(c, fiber.StatusServiceUnavailable, "not_found_error", fmt.Sprintf("No available deployment for model: %s", requestedModel))
	}

	if !supportsAnthropicMessagesProvider(dep.Provider) {
		return writeAnthropicError(c, fiber.StatusNotFound, "not_found_error", fmt.Sprintf("Messages API is not enabled for provider: %s", dep.Provider))
	}

	estimated, err = applyDeploymentPricingOverride(c.Context(), h.pipeline, estimated, dep)
	if err != nil {
		return writeAnthropicBillingError(c, err)
	}

	outboundModel := strings.TrimSpace(dep.Model)
	if outboundModel == "" || outboundModel == "*" {
		outboundModel = providerModelID(requestedModel)
	}

	payload, err := rewriteOpenAIJSONModel(originalBody, providerModelID(outboundModel))
	if err != nil {
		return writeAnthropicError(c, fiber.StatusBadRequest, "invalid_request_error", "Failed to encode request")
	}

	runtimeProfiles, err := runtimeProfilesFromEstimate(dep.Provider, outboundModel, estimated)
	if err != nil {
		return writeAnthropicError(c, fiber.StatusServiceUnavailable, "runtime_profile_unavailable", err.Error())
	}

	tunnelReq := platform.TunnelRequest{
		RequestID:            requestID,
		TenantID:             currentTenant.ID,
		TraceID:              requestID,
		Stream:               fields.Stream,
		Operation:            platform.ProviderOperationAnthropicMessages,
		Deployment:           toContractDeployment(dep),
		ModelRuntimeProfiles: runtimeProfiles,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: payload,
		RequestMetadata: map[string]interface{}{
			"tenant_slug": apiheaders.GetTenant(c),
			"model":       requestedModel,
		},
	}
	tunnelReq.RequestMetadata = enrichRequestMetadataForDeployment(tunnelReq.RequestMetadata, tunnelReq.Deployment)
	requestLog.SetDeployment(tunnelReq.Deployment)
	requestLog.SetRequestIDs(tunnelReq.RequestID, tunnelReq.TraceID)

	reservation, err := h.pipeline.Reserve(c.Context(), budgetRequestFromHTTP(c, authInfo, estimated, requestID, deploymentBillingMode(tunnelReq.Deployment)))
	if err != nil {
		return writeBudgetCheckError(c, err)
	}

	if fields.Stream {
		return h.handleStream(c, tunnelReq, reservation, estimated, authInfo, outboundModel)
	}

	return h.handleSync(c, requestedModel, originalBody, authInfo, estimated, tunnelReq, reservation, startTime, tenantSlug, requestLog)
}

type tokenBazaarMessagesRequestFields struct {
	Model  string `json:"model"`
	Stream bool   `json:"stream"`
}

func parseTokenBazaarMessagesRequest(body []byte) (tokenBazaarMessagesRequestFields, error) {
	var fields tokenBazaarMessagesRequestFields
	if err := json.Unmarshal(body, &fields); err != nil {
		return fields, fmt.Errorf("invalid request body")
	}

	fields.Model = strings.TrimSpace(fields.Model)

	return fields, nil
}

func (h *TokenBazaarMessagesHandler) handleSync(
	c *fiber.Ctx,
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
	response, err := h.pipeline.Dispatch(c.Context(), tunnelReq)
	if err != nil {
		h.pipeline.Release(c.Context(), reservation, "tunnel_error")

		return writeAnthropicError(c, fiber.StatusBadGateway, "api_error", "Tunnel request failed")
	}

	if response.StatusCode >= fiber.StatusBadRequest {
		h.pipeline.Release(c.Context(), reservation, "provider_http_error")
		forwardProviderHeaders(c, response.Headers)

		return c.Status(response.StatusCode).Send(response.Body)
	}

	usage := parseAnthropicNativeUsage(response.Body)

	finalized, finalizeErr := h.billing.FinalizeFromUsage(c.Context(), requestedModel, usage, estimated)
	if finalizeErr != nil {
		h.pipeline.Release(c.Context(), reservation, "usage_finalize_error")

		return writeAnthropicBillingError(c, finalizeErr)
	}

	if err := h.pipeline.Settle(c.Context(), SettlementRequest{
		Reservation: reservation,
		TenantID:    authInfo.TenantID,
		Actual:      finalized,
		RequestID:   tunnelReq.RequestID,
		BillingMode: tunnelReq.Deployment.BillingMode,
	}); err != nil {
		return writeAnthropicError(c, fiber.StatusPaymentRequired, "billing_error", err.Error())
	}

	responseTimeMS := int(time.Since(start).Milliseconds())
	logEntry := buildRequestLogEntry(c, authInfo, tunnelReq.Deployment, requestedModel, "/v1/messages", "messages", "success", response.StatusCode, start, responseTimeMS)
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

	logSpendBestEffort(h.logger, h.logs, c.Context(), platform.SpendLog{
		RequestID:      tunnelReq.RequestID,
		TokenHash:      authInfo.TokenHash,
		TenantID:       authInfo.TenantID,
		OrganizationID: authInfo.OrganizationID,
		TeamID:         authInfo.TeamID,
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

	h.emitModelMetric(metricsobs.BuildInput{
		TS:                 time.Now().UTC(),
		TenantID:           authInfo.TenantID,
		RequestID:          tunnelReq.RequestID,
		ModelName:          requestedModel,
		ProviderName:       tunnelReq.Deployment.Provider,
		ProviderModelName:  tunnelReq.Deployment.Model,
		RequestType:        "messages",
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
			"endpoint":    "/v1/messages",
			"key_type":    authInfo.KeyType,
			"team_id":     authInfo.TeamID,
			"org_unit_id": authInfo.OrgUnitID,
			"tenant_slug": tenantSlug,
		}, tunnelReq.Deployment),
	})

	forwardProviderHeaders(c, response.Headers)
	setCostHeaders(c, finalized)

	if contentType := response.Headers["Content-Type"]; contentType != "" {
		c.Set(fiber.HeaderContentType, contentType)
	} else {
		c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	}

	return c.Status(response.StatusCode).Send(response.Body)
}

// handleStream delegates to the shared streamAnthropicShapedResponse
// (anthropic_native_handler.go), which also serves the direct Anthropic
// provider's own streaming dispatch -- see that function's doc comment.
func (h *TokenBazaarMessagesHandler) handleStream(
	c *fiber.Ctx,
	request platform.TunnelRequest,
	reservation *platform.BudgetReservation,
	estimated platform.UsageCostBreakdown,
	authInfo *platform.AuthContext,
	model string,
) error {
	return streamAnthropicShapedResponse(c, h.pipeline, h.billing, h.logger, request, reservation, estimated, authInfo, model)
}

func (h *TokenBazaarMessagesHandler) emitModelMetric(input metricsobs.BuildInput) {
	if h.emitMetric == nil {
		return
	}

	h.emitMetric(input)
}
