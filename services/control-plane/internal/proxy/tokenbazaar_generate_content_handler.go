// Package proxy hosts OpenAI-compatible request handlers for the control plane.
package proxy

import (
	"fmt"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/apiheaders"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	metricsobs "github.com/Delpak/shahrah/services/control-plane/internal/observability/metrics"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// TokenBazaarGenerateContentHandler serves POST /v1beta/models/{model}:generateContent
// and POST /v1beta/models/{model}:streamGenerateContent requests that target a
// TokenBazaar deployment (an explicit "tokenbazaar:<account>:<model-id>" FQN
// in the URL path's {model} segment).
//
// This handler exists as a narrow, TokenBazaar-scoped sibling to
// GeminiNativeHandler rather than as an extension of it. See
// plans/tokenbazaar-production-readiness.md, "Native Gemini
// generateContent/streamGenerateContent delivery", for the full
// architectural reasoning. In short, mirroring the precedent set by
// TokenBazaarMessagesHandler for the direct Anthropic provider:
//
//   - GeminiNativeHandler is built around a single provider check
//     (strings.ToLower(deployment.Provider) == "google-gemini" || "gemini",
//     in resolveDeployment) and resolves deployments via its own
//     anthropicDeploymentLister with no FQN concept at all -- a bare model
//     id such as "gemini-2.5-flash" always resolves to the direct
//     google-gemini provider, never to a FQN'd account. Widening that single
//     check to also accept "tokenbazaar" would not be the narrow, low-risk
//     change it might look like: isAllowlistedGoogleGeminiModel matches on
//     the same bare Google model-name strings TokenBazaar also carries
//     (e.g. "gemini-2.5-flash"), so widening the provider filter alone would
//     silently also open every OTHER Gemini-native operation this handler
//     serves for that model -- context caching (cachedContents),
//     interactions, files, batches, countTokens, embedContent -- none of
//     which have been evidenced, validated, or tested for TokenBazaar. Worse,
//     parseGeminiNativeUsage hardcodes
//     billingusage.NormalizeUsageFromJSON("google-gemini", ...): a
//     TokenBazaar-routed response would be billing-normalized under the
//     wrong provider slug. Both are real correctness/scope risks, not
//     hypothetical ones -- so unlike the Anthropic Messages delivery's own
//     finding (where widening was avoidable but not otherwise unsafe), this
//     one specific single-`if` widening was investigated and rejected as
//     unsafe on its own, independent grounds.
//   - Investigation confirmed POST /v1beta/models/* is wired, in
//     services/control-plane/internal/server/routes_proxy.go, exclusively to
//     GeminiNativeHandler.Handle -- there is no pre-existing
//     provider-agnostic handler for
//     platform.ProviderOperationGeminiGenerateContent /
//     ...GeminiStreamGenerateContent the way there is for chat_completions,
//     responses, images, and embeddings.
//   - This handler is therefore a new, narrow sibling that reuses the
//     standard FQN -> deployment (router.Router) -> providerregistry
//     capability -> Silk Road adapter flow already used for TokenBazaar's
//     other capabilities (chat, images, embeddings, responses, messages). A
//     thin routing shim (NewGeminiGenerateContentRouter, wired in
//     routes_proxy.go in place of the bare "/v1beta/models/*" mount)
//     inspects the URL path's {model} segment before GeminiNativeHandler is
//     ever invoked: a tokenbazaar-FQN'd generateContent/streamGenerateContent
//     request is dispatched here instead; every other request (including
//     every direct-google-gemini request, FQN'd or not, and every other
//     Gemini-native operation) falls through to native
//     (GeminiNativeHandler.Handle, unmodified) unconditionally.
//     GeminiNativeHandler's dispatch/routing logic, its provider filter, and
//     its caching/interactions/files/batches machinery are completely
//     untouched by this delivery. The one exception is a pure,
//     behavior-preserving extract-method refactor (mirroring the Anthropic
//     Messages delivery's own precedent): GeminiNativeHandler.proxyStream's
//     body was moved into the new shared function
//     streamGeminiShapedResponse (gemini_native_handler.go), which this
//     handler also calls, so the SSE streaming-relay logic exists in exactly
//     one place instead of two near-identical copies (this repo's `dupl`
//     linter would otherwise flag the duplication). Likewise,
//     GeminiNativeHandler.estimateCost's chat-style billing-estimate
//     construction was extracted into the shared, pure
//     geminiGenerateContentEstimateRequest function, reused by both
//     handlers.
//
// Evidence-tier caveat (weaker than every other TokenBazaar capability):
// TokenBazaar's own rendered docs page never mentions a Gemini-native API
// style anywhere in its "Text Generation (LLM)" prose (which explicitly
// names exactly three styles: Chat Completions, Responses, Anthropic
// Messages). The only evidence this capability exists at all is the
// authenticated GET /v1/models catalog's supported_endpoint_types column
// listing "generateContent" for the Gemini models. This delivery implements
// against the standard Gemini-native shape gemini.go already builds for the
// direct google-gemini provider, and Silk Road's tokenBazaarAdapter adds a
// defensive response-shape check
// (validateTokenBazaarGenerateContentResponseShape) before returning
// success -- see plans/tokenbazaar-production-readiness.md for the full
// evidence-tier writeup and the release gate this stays behind.
type TokenBazaarGenerateContentHandler struct {
	llmRouter  *router.Router
	pipeline   *RequestPipeline
	billing    budgeter
	logs       usageLogger
	logger     *zap.Logger
	emitMetric func(input metricsobs.BuildInput)
}

// NewTokenBazaarGenerateContentHandlerWithDependencies creates the handler
// from explicit dependencies, mirroring every other *HandlerWithDependencies
// constructor in this package.
func NewTokenBazaarGenerateContentHandlerWithDependencies(deps BaseHandlerDependencies) *TokenBazaarGenerateContentHandler {
	return &TokenBazaarGenerateContentHandler{
		llmRouter:  deps.Router,
		pipeline:   deps.requestPipeline(),
		billing:    deps.Billing,
		logs:       deps.logger(),
		logger:     deps.Logger,
		emitMetric: deps.metricEmitter(),
	}
}

// parseGeminiModelActionPath splits a "/v1beta/models/<model>:<action>" URL
// path into its model and action segments. Unlike
// geminiOperationFromRequest's own path parsing in gemini_native_handler.go
// (which splits on the FIRST colon -- correct for a direct Google Gemini
// model id, which never itself contains a colon), this splits on the LAST
// colon: a TokenBazaar-FQN'd model ("tokenbazaar:<account>:<model-id>")
// itself contains two colons, so the trailing ":<action>" suffix must be
// recovered from the end of the string, not the start.
func parseGeminiModelActionPath(path string) (model string, action string, ok bool) {
	const prefix = "/v1beta/models/"

	if !strings.HasPrefix(path, prefix) {
		return "", "", false
	}

	rest := strings.TrimPrefix(path, prefix)

	separator := strings.LastIndex(rest, ":")
	if separator <= 0 || separator == len(rest)-1 {
		return "", "", false
	}

	return rest[:separator], rest[separator+1:], true
}

// requestTargetsTokenBazaarGenerateContent reports whether a
// "/v1beta/models/<model>:<action>" request path names a
// generateContent/streamGenerateContent action whose <model> is an explicit
// tokenbazaar:<account>:<model> FQN. This is the sole signal used to decide,
// before GeminiNativeHandler is ever invoked, whether a request belongs to
// this handler instead -- a bare model id (e.g. "gemini-2.5-flash") is never
// routed here, exactly like every other TokenBazaar capability's "explicit
// FQN only" rule.
func requestTargetsTokenBazaarGenerateContent(path string) bool {
	model, action, ok := parseGeminiModelActionPath(strings.TrimSuffix(path, "/"))
	if !ok || (action != "generateContent" && action != "streamGenerateContent") {
		return false
	}

	provider, _, _, fqnOK := splitModelFQN(model)

	return fqnOK && isTokenBazaarProvider(provider)
}

// NewGeminiGenerateContentRouter returns the fiber.Handler registered for the
// "/v1beta/models/*" wildcard mount. It dispatches a
// POST /v1beta/models/<model>:generateContent or
// .../<model>:streamGenerateContent request whose <model> is an explicit
// tokenbazaar-FQN to tokenBazaar (this package's narrow, FQN-routed path) and
// falls back to native (GeminiNativeHandler.Handle, unmodified) for every
// other request -- including every direct-google-gemini request and every
// other Gemini-native model operation (countTokens, embedContent,
// batchGenerateContent, ...) -- preserving the direct Gemini provider's
// existing behavior exactly.
func NewGeminiGenerateContentRouter(tokenBazaar *TokenBazaarGenerateContentHandler, native fiber.Handler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if tokenBazaar != nil && c.Method() == fiber.MethodPost && requestTargetsTokenBazaarGenerateContent(c.Path()) {
			return tokenBazaar.Handle(c)
		}

		if native == nil {
			return fiber.ErrNotFound
		}

		return native(c)
	}
}

// supportsGeminiGenerateContentProvider reports whether a routed
// deployment's provider may serve
// platform.ProviderOperationGeminiGenerateContent /
// ...GeminiStreamGenerateContent through this generic handler. Only
// "tokenbazaar" is listed here: the direct "google-gemini" provider
// continues to be served exclusively by GeminiNativeHandler and must never
// reach this generic FQN-routed path (its deployments are never resolved via
// a bare-model FQN in the first place), but this check is kept as an
// explicit, fail-closed defense in depth.
func supportsGeminiGenerateContentProvider(provider string) bool {
	return isTokenBazaarProvider(provider)
}

// Handle processes a TokenBazaar-targeted generateContent/streamGenerateContent
// request.
func (h *TokenBazaarGenerateContentHandler) Handle(c *fiber.Ctx) error {
	startTime := time.Now()

	authInfo := auth.GetAuthInfo(c)
	if authInfo == nil {
		return writeGeminiError(c, fiber.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return writeGeminiError(c, fiber.StatusBadRequest, "INVALID_ARGUMENT", "Tenant resolution required")
	}

	authInfo.TenantID = currentTenant.ID
	tenantSlug := strings.TrimSpace(currentTenant.Slug)

	fqnModel, action, ok := parseGeminiModelActionPath(strings.TrimSuffix(c.Path(), "/"))
	if !ok {
		return writeGeminiError(c, fiber.StatusBadRequest, "INVALID_ARGUMENT", "A fixed model operation is required")
	}

	operation := platform.ProviderOperationGeminiGenerateContent

	stream := false

	if action == "streamGenerateContent" {
		operation = platform.ProviderOperationGeminiStreamGenerateContent
		stream = true
	} else if action != "generateContent" {
		return writeGeminiError(c, fiber.StatusBadRequest, "INVALID_ARGUMENT", "Unsupported Gemini model operation")
	}

	originalBody := append([]byte(nil), c.Body()...)

	requestedModel := normalizeModelIdentifier(fqnModel)
	if requestedModel == "" {
		return writeGeminiError(c, fiber.StatusBadRequest, "INVALID_ARGUMENT", "model is required")
	}

	requestLog := newRequestLogLifecycle(h.logger, h.logs, c, authInfo, "/v1beta/models/"+action, "generate_content", startTime)
	requestLog.SetRequestBody(buildJSONRequestLogPayload(originalBody, nil))
	requestLog.SetModel(requestedModel)

	defer requestLog.FinalizeFromFiber()

	modelAccess := modelAccessForType(authInfo, requestedModel, platform.ModelTypeGenerateContent)
	if !modelAccess.Allowed {
		denyReason := denyReasonForModel(authInfo)
		emitDenyCounter(h.logger, denyReason, requestedModel, currentTenant.ID)
		requestLog.MarkFinalized()

		return writeGeminiError(c, fiber.StatusForbidden, "PERMISSION_DENIED", "Model access denied or unsupported for generateContent")
	}

	estimated, err := h.billing.EstimateCost(c.Context(), geminiGenerateContentEstimateRequest(billingEstimateModel(requestedModel), originalBody))
	if err != nil {
		return writeGeminiError(c, fiber.StatusServiceUnavailable, "PRICING_UNAVAILABLE", "Gemini pricing is unavailable")
	}

	requestID := fmt.Sprintf("req_%d", time.Now().UnixNano())

	if tunnelErr := h.pipeline.EnsureReady(c.Context()); tunnelErr != nil {
		return writeGeminiError(c, fiber.StatusServiceUnavailable, "UNAVAILABLE", "Tunnel path unavailable")
	}

	dep, err := h.llmRouter.Route(c.Context(), currentTenant.ID, requestedModel)
	if err != nil || dep == nil {
		emitDenyCounter(h.logger, denyReasonUnknownModel, requestedModel, currentTenant.ID)

		return writeGeminiError(c, fiber.StatusServiceUnavailable, "NOT_FOUND", fmt.Sprintf("No available deployment for model: %s", requestedModel))
	}

	if !supportsGeminiGenerateContentProvider(dep.Provider) {
		return writeGeminiError(c, fiber.StatusNotFound, "NOT_FOUND", fmt.Sprintf("generateContent is not enabled for provider: %s", dep.Provider))
	}

	estimated, err = applyDeploymentPricingOverride(c.Context(), h.pipeline, estimated, dep)
	if err != nil {
		return writeGeminiError(c, fiber.StatusServiceUnavailable, "PRICING_UNAVAILABLE", err.Error())
	}

	outboundModel := strings.TrimSpace(dep.Model)
	if outboundModel == "" || outboundModel == "*" {
		outboundModel = providerModelID(requestedModel)
	}

	runtimeProfiles, err := runtimeProfilesFromEstimate(dep.Provider, outboundModel, estimated)
	if err != nil {
		return writeGeminiError(c, fiber.StatusServiceUnavailable, "RUNTIME_PROFILE_UNAVAILABLE", err.Error())
	}

	tunnelReq := platform.TunnelRequest{
		RequestID:            requestID,
		TenantID:             currentTenant.ID,
		TraceID:              requestID,
		Stream:               stream,
		Operation:            operation,
		Deployment:           toContractDeployment(dep),
		ModelRuntimeProfiles: runtimeProfiles,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: originalBody,
		RequestMetadata: map[string]interface{}{
			"tenant_slug": apiheaders.GetTenant(c),
			"model":       requestedModel,
		},
	}
	// Gemini's generateContent body carries no "model" field (the model is
	// implicit from the URL path), so -- unlike the OpenAI-compatible and
	// Anthropic Messages paths, which rewrite the body's "model" field --
	// the resolved outbound model must be propagated via
	// tunnelReq.Deployment.Model directly: Silk Road's
	// tokenBazaarRequestModel falls back to exactly this field when the body
	// has no "model" key.
	tunnelReq.Deployment.Model = outboundModel
	tunnelReq.RequestMetadata = enrichRequestMetadataForDeployment(tunnelReq.RequestMetadata, tunnelReq.Deployment)
	requestLog.SetDeployment(tunnelReq.Deployment)
	requestLog.SetRequestIDs(tunnelReq.RequestID, tunnelReq.TraceID)

	reservation, err := h.pipeline.Reserve(c.Context(), budgetRequestFromHTTP(c, authInfo, estimated, requestID, deploymentBillingMode(tunnelReq.Deployment)))
	if err != nil {
		return writeBudgetCheckError(c, err)
	}

	if stream {
		return h.handleStream(c, tunnelReq, reservation, estimated, authInfo, outboundModel)
	}

	return h.handleSync(c, requestedModel, originalBody, authInfo, estimated, tunnelReq, reservation, startTime, tenantSlug, requestLog)
}

func (h *TokenBazaarGenerateContentHandler) handleSync(
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

		return writeGeminiError(c, fiber.StatusBadGateway, "UNAVAILABLE", "Tunnel request failed")
	}

	if response.StatusCode >= fiber.StatusBadRequest {
		h.pipeline.Release(c.Context(), reservation, "provider_http_error")
		forwardProviderHeaders(c, response.Headers)

		return c.Status(response.StatusCode).Send(response.Body)
	}

	usage := parseGeminiNativeUsage(response.Body)

	finalized, finalizeErr := h.billing.FinalizeFromUsage(c.Context(), requestedModel, usage, estimated)
	if finalizeErr != nil {
		h.pipeline.Release(c.Context(), reservation, "usage_finalize_error")

		return writeGeminiError(c, fiber.StatusBadGateway, "BILLING_ERROR", "Failed to finalize Gemini usage")
	}

	if err := h.pipeline.Settle(c.Context(), SettlementRequest{
		Reservation: reservation,
		TenantID:    authInfo.TenantID,
		Actual:      finalized,
		RequestID:   tunnelReq.RequestID,
		BillingMode: tunnelReq.Deployment.BillingMode,
	}); err != nil {
		return writeGeminiError(c, fiber.StatusPaymentRequired, "BILLING_ERROR", err.Error())
	}

	responseTimeMS := int(time.Since(start).Milliseconds())
	logEntry := buildRequestLogEntry(c, authInfo, tunnelReq.Deployment, requestedModel, "/v1beta/models", "generate_content", "success", response.StatusCode, start, responseTimeMS)
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
		RequestType:        "generate_content",
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
			"endpoint":    "/v1beta/models:generateContent",
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

// handleStream delegates to the shared streamGeminiShapedResponse
// (gemini_native_handler.go), which also serves the direct google-gemini
// provider's own streaming dispatch -- see that function's doc comment.
func (h *TokenBazaarGenerateContentHandler) handleStream(
	c *fiber.Ctx,
	request platform.TunnelRequest,
	reservation *platform.BudgetReservation,
	estimated platform.UsageCostBreakdown,
	authInfo *platform.AuthContext,
	model string,
) error {
	return streamGeminiShapedResponse(c, h.pipeline, h.billing, h.logger, request, reservation, estimated, authInfo, model)
}

func (h *TokenBazaarGenerateContentHandler) emitModelMetric(input metricsobs.BuildInput) {
	if h.emitMetric == nil {
		return
	}

	h.emitMetric(input)
}
