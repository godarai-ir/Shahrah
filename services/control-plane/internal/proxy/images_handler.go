package proxy

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/apiheaders"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	metricsobs "github.com/Delpak/shahrah/services/control-plane/internal/observability/metrics"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type ImagesHandler struct {
	llmRouter  *router.Router
	tunnel     platform.TunnelTransport
	billing    budgeter
	wallets    walletStore
	logs       usageLogger
	logger     *zap.Logger
	emitMetric func(input metricsobs.BuildInput)
}

type imageOperation struct {
	Name        string
	Path        string
	RequestType string
}

func NewImagesHandler(llmRouter *router.Router, tunnel platform.TunnelTransport, billing budgeter, wallets walletStore, logs usageLogger, logger *zap.Logger, emitMetric ...func(input metricsobs.BuildInput)) *ImagesHandler {
	var metricEmitter func(input metricsobs.BuildInput)
	if len(emitMetric) > 0 {
		metricEmitter = emitMetric[0]
	}

	return NewImagesHandlerWithDependencies(BaseHandlerDependencies{
		Router:        llmRouter,
		Tunnel:        tunnel,
		Billing:       billing,
		Wallets:       wallets,
		RequestLogger: logs,
		SpendLogger:   logs,
		Logger:        logger,
		EmitMetric:    metricEmitter,
	})
}

func NewImagesHandlerWithDependencies(deps BaseHandlerDependencies) *ImagesHandler {
	return &ImagesHandler{
		llmRouter:  deps.Router,
		tunnel:     deps.Tunnel,
		billing:    deps.Billing,
		wallets:    deps.Wallets,
		logs:       deps.logger(),
		logger:     deps.Logger,
		emitMetric: deps.metricEmitter(),
	}
}

func (h *ImagesHandler) HandleGenerations(c *fiber.Ctx) error {
	return h.handle(c, imageOperation{Name: "generations", Path: "/images/generations", RequestType: "image_generations"})
}

func (h *ImagesHandler) HandleEdits(c *fiber.Ctx) error {
	return h.handle(c, imageOperation{Name: "edits", Path: "/images/edits", RequestType: "image_edits"})
}

func (h *ImagesHandler) HandleVariations(c *fiber.Ctx) error {
	return h.handle(c, imageOperation{Name: "variations", Path: "/images/variations", RequestType: "image_variations"})
}

func (h *ImagesHandler) handle(c *fiber.Ctx, op imageOperation) error {
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

	contentType := strings.TrimSpace(c.Get(fiber.HeaderContentType))
	body := append([]byte(nil), c.Body()...)

	requestedModel, metadata, err := extractImageRequestMetadata(c, body, contentType, op)
	if err != nil {
		return writeAPIErrorDetail(c, fiber.StatusBadRequest, err.Error(), "invalid_request_error", "model", "invalid_model")
	}

	requestedModel = normalizeModelIdentifier(requestedModel)

	routeCandidates := modelRoutingCandidates(requestedModel)
	if len(routeCandidates) == 0 {
		return writeAPIErrorDetail(c, fiber.StatusBadRequest, "model is required", "invalid_request_error", "model", "invalid_model")
	}

	tracing.SetModelForTrace(c, requestedModel)

	modelAccess := modelAccessForType(authInfo, requestedModel, platform.ModelTypeImageGeneration)
	if !modelAccess.Allowed {
		if modelAccess.UnsupportedVMMode {
			return writeUnsupportedModelModeError(c, requestedModel, platform.ModelTypeImageGeneration, modelAccess.ExposedModes)
		}

		emitDenyCounter(h.logger, denyReasonForModel(authInfo), requestedModel, currentTenant.ID)

		return writeAPIError(c, fiber.StatusForbidden, "Model access denied or unsupported for image generation", "authorization_error")
	}

	authStart := time.Now()
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "auth.validate", "internal", authStart, fiber.StatusOK, map[string]interface{}{"model": requestedModel})

	estimated := estimateImageCostFallback(requestedModel, metadata)
	quotaStart := time.Now()
	rateLimit, err := h.billing.CheckRateLimit(c.Context(), authInfo, requestedModel, estimated.TotalTokens)
	setRateLimitHeaders(c, rateLimit)

	if err != nil {
		return writeAPIError(c, rateLimitErrorStatus(err), err.Error(), "rate_limit_error")
	}

	budgetErr := h.billing.CheckBudget(c.Context(), authInfo, estimated)
	if budgetErr != nil {
		return writeAPIError(c, fiber.StatusPaymentRequired, budgetErr.Error(), "budget_exceeded")
	}

	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "quota.check", "internal", quotaStart, fiber.StatusOK, map[string]interface{}{"estimated_tokens": estimated.TotalTokens})

	tunnelErr := h.tunnel.Healthy(c.Context())
	if tunnelErr != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Tunnel path unavailable", "tunnel_unavailable")
	}

	routeStart := time.Now()

	var dep *router.Deployment
	for _, candidate := range routeCandidates {
		dep, err = h.llmRouter.Route(c.Context(), currentTenant.ID, candidate)
		if err == nil {
			break
		}
	}

	if err != nil || dep == nil {
		emitDenyCounter(h.logger, denyReasonUnknownModel, requestedModel, currentTenant.ID)
		return writeAPIError(c, fiber.StatusServiceUnavailable, fmt.Sprintf("No available deployment for model: %s", requestedModel), "routing_error")
	}

	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "router.select_model", "internal", routeStart, fiber.StatusOK, map[string]interface{}{"provider": dep.Provider, "deployment_id": dep.ID})

	outboundBody := body

	if isJSONContentType(contentType) {
		outboundModel := strings.TrimSpace(dep.Model)
		if outboundModel == "" || outboundModel == "*" {
			outboundModel = providerModelID(requestedModel)
		}

		outboundBody, err = rewriteJSONModel(body, providerModelID(outboundModel))
		if err != nil {
			return writeAPIError(c, fiber.StatusBadRequest, "Invalid request body", "invalid_request_error")
		}
	}

	if contentType == "" {
		contentType = "application/json"
	}

	requestID := tracing.TraceIDFromContext(c)
	if requestID == "" {
		requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
	}

	requestMetadata := map[string]interface{}{
		"tenant_slug":            apiheaders.GetTenant(c),
		"model":                  requestedModel,
		"provider_endpoint_path": op.Path,
		"operation":              op.Name,
		"billing_mode":           "image_mvp_reconciliation",
	}
	for k, v := range metadata {
		requestMetadata[k] = v
	}

	tunnelReq := platform.TunnelRequest{
		RequestID:  requestID,
		TenantID:   currentTenant.ID,
		TraceID:    requestID,
		Stream:     false,
		Deployment: toContractDeployment(dep),
		Headers: map[string]string{
			"Content-Type": contentType,
		},
		Body:            outboundBody,
		RequestMetadata: requestMetadata,
	}

	return h.handleSync(c, requestedModel, estimated, tunnelReq, op, startTime, tenantSlug, contentType)
}

func (h *ImagesHandler) handleSync(c *fiber.Ctx, requestedModel string, estimated platform.UsageCostBreakdown, tunnelReq platform.TunnelRequest, op imageOperation, start time.Time, tenantSlug string, contentType string) error {
	providerStart := time.Now()
	providerAttrs := map[string]interface{}{
		"provider":        tunnelReq.Deployment.Provider,
		"request_payload": buildTraceImageRequestPayload(op.Name, tunnelReq.Body, contentType, tunnelReq.RequestMetadata),
	}

	resp, err := h.tunnel.Proxy(c.Context(), tunnelReq)
	if err != nil {
		providerAttrs["error"] = err.Error()
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, fiber.StatusBadGateway, providerAttrs)

		return writeAPIError(c, fiber.StatusBadGateway, fmt.Sprintf("Tunnel error: %v", err), "api_error")
	}

	if resp.StatusCode >= httpErrorFloor {
		providerAttrs["response_payload"] = buildTraceProviderErrorPayload(resp.StatusCode, resp.Body)
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, resp.StatusCode, providerAttrs)

		return transformProviderError(c, resp.StatusCode, resp.Body, resp.Headers)
	}

	providerAttrs["response_payload"] = buildTraceImageResponsePayload(resp.Body)
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", providerStart, resp.StatusCode, providerAttrs)

	finalCost := estimated
	finalCost.Model = billingModelForDeployment(requestedModel, tunnelReq.Deployment)

	metadata := enrichPricingMetadata(withImageResponseMetadata(tunnelReq.RequestMetadata, resp.Body), auth.GetAuthInfo(c), requestedModel)

	_, debitErr := h.wallets.DebitWallet(c.Context(), auth.GetAuthInfo(c).TenantID, finalCost, tunnelReq.RequestID)
	if debitErr != nil {
		return writeAPIError(c, fiber.StatusPaymentRequired, debitErr.Error(), "budget_exceeded")
	}

	authInfo := auth.GetAuthInfo(c)
	responseTimeMS := int(time.Since(start).Milliseconds())
	_ = h.logs.LogRequest(c.Context(), platform.RequestLogEntry{
		RequestID:      tunnelReq.RequestID,
		TokenHash:      authInfo.TokenHash,
		TenantID:       authInfo.TenantID,
		Model:          requestedModel,
		Status:         "success",
		CostUSD:        finalCost.TotalCostUSD,
		ResponseTimeMS: responseTimeMS,
		RequestBody:    map[string]interface{}{"model": requestedModel, "operation": op.Name, "n": tunnelReq.RequestMetadata["n"], "size": tunnelReq.RequestMetadata["size"], "quality": tunnelReq.RequestMetadata["quality"]},
		ResponseBody:   buildTraceImageResponsePayload(resp.Body),
	})
	_ = h.logs.LogSpend(c.Context(), platform.SpendLog{
		RequestID:       tunnelReq.RequestID,
		TokenHash:       authInfo.TokenHash,
		TenantID:        authInfo.TenantID,
		OrganizationID:  authInfo.OrganizationID,
		TeamID:          authInfo.TeamID,
		Model:           requestedModel,
		InputTokens:     finalCost.InputTokens,
		OutputTokens:    finalCost.OutputTokens,
		TotalTokens:     finalCost.TotalTokens,
		CostUSD:         finalCost.TotalCostUSD,
		CostIRR:         finalCost.TotalCostIRR,
		ResponseTimeMS:  responseTimeMS,
		Status:          "success",
		RequestMetadata: metadata,
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
		RequestType:        op.RequestType,
		UserID:             authInfo.UserID,
		VirtualAccountID:   authInfo.KeyID,
		VirtualAccountName: authInfo.KeyAlias,
		TeamIDs:            nonEmptySlice([]string{authInfo.TeamID}),
		StatusCode:         resp.StatusCode,
		LatencyMS:          responseTimeMS,
		CostUSD:            finalCost.TotalCostUSD,
		Metadata: map[string]interface{}{
			"endpoint":    "/v1" + op.Path,
			"key_type":    authInfo.KeyType,
			"team_id":     authInfo.TeamID,
			"org_unit_id": authInfo.OrgUnitID,
			"tenant_slug": tenantSlug,
			"operation":   op.Name,
		},
	})

	forwardProviderHeaders(c, resp.Headers)
	setTraceHeaders(c, tunnelReq.TraceID)
	setCostHeaders(c, finalCost)

	responseStart := time.Now()
	err = c.Status(resp.StatusCode).Send(resp.Body)
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "response.write", "internal", responseStart, resp.StatusCode, map[string]interface{}{
		"response_payload": buildTraceImageResponsePayload(resp.Body),
	})

	return err
}

func (h *ImagesHandler) emitModelMetric(input metricsobs.BuildInput) {
	if h.emitMetric == nil {
		return
	}

	h.emitMetric(input)
}

func extractImageRequestMetadata(c *fiber.Ctx, body []byte, contentType string, op imageOperation) (string, map[string]interface{}, error) {
	metadata := map[string]interface{}{"content_type": safeContentTypeFamily(contentType)}
	if isJSONContentType(contentType) || strings.TrimSpace(contentType) == "" {
		var raw map[string]interface{}
		if err := json.Unmarshal(body, &raw); err != nil {
			return "", metadata, fmt.Errorf("invalid request body")
		}

		model, _ := raw["model"].(string)
		fillImageJSONMetadata(metadata, raw, op)

		if strings.TrimSpace(model) == "" {
			return "", metadata, fmt.Errorf("model is required")
		}

		return strings.TrimSpace(model), metadata, nil
	}

	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(contentType)), "multipart/form-data") {
		model := strings.TrimSpace(c.FormValue("model"))
		fillImageMultipartMetadata(c, metadata, op)

		if model == "" {
			return "", metadata, fmt.Errorf("model is required")
		}

		return model, metadata, nil
	}

	return "", metadata, fmt.Errorf("unsupported content type")
}

func fillImageJSONMetadata(metadata map[string]interface{}, raw map[string]interface{}, op imageOperation) {
	if prompt, ok := raw["prompt"].(string); ok {
		metadata["prompt_length"] = len(prompt)
		metadata["prompt_hash"] = hashTracePreview(prompt)
	}

	for _, key := range []string{"size", "quality", "response_format"} {
		if v, ok := raw[key].(string); ok && strings.TrimSpace(v) != "" {
			metadata[key] = strings.TrimSpace(v)
		}
	}

	if n := imageNumber(raw["n"]); n > 0 {
		metadata["n"] = n
	}

	if _, ok := raw["image"]; ok {
		metadata["image_source_type"] = traceImageSourceType(raw["image"])
	}

	if op.Name == "edits" || op.Name == "variations" {
		metadata["file_count"] = 0
	}
}

func fillImageMultipartMetadata(c *fiber.Ctx, metadata map[string]interface{}, op imageOperation) {
	for _, key := range []string{"size", "quality", "response_format"} {
		if v := strings.TrimSpace(c.FormValue(key)); v != "" {
			metadata[key] = v
		}
	}

	if prompt := c.FormValue("prompt"); prompt != "" {
		metadata["prompt_length"] = len(prompt)
		metadata["prompt_hash"] = hashTracePreview(prompt)
	}

	if n, err := strconv.Atoi(strings.TrimSpace(c.FormValue("n"))); err == nil && n > 0 {
		metadata["n"] = n
	}

	form, err := c.MultipartForm()
	if err != nil || form == nil {
		return
	}

	fileCount := 0
	for _, files := range form.File {
		fileCount += len(files)
	}

	metadata["file_count"] = fileCount
	if fileCount > 0 {
		metadata["image_source_type"] = "multipart_file"
	}

	_ = op
}

func isJSONContentType(contentType string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(contentType)), "application/json")
}

func rewriteJSONModel(body []byte, model string) ([]byte, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	raw["model"] = model

	return json.Marshal(raw)
}

func estimateImageCostFallback(model string, metadata map[string]interface{}) platform.UsageCostBreakdown {
	return platform.UsageCostBreakdown{
		Model:       model,
		TotalTokens: 0,
	}
}

func withImageResponseMetadata(input map[string]interface{}, body []byte) map[string]interface{} {
	out := map[string]interface{}{}
	for k, v := range input {
		out[k] = v
	}

	out["response_summary"] = buildTraceImageResponsePayload(body)
	if ticks, ok := imageUsageCostTicks(body); ok {
		out["provider_cost_in_usd_ticks"] = ticks
	}

	return out
}

func imageUsageCostTicks(body []byte) (int64, bool) {
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return 0, false
	}

	usage, ok := raw["usage"].(map[string]interface{})
	if !ok {
		return 0, false
	}

	switch v := usage["cost_in_usd_ticks"].(type) {
	case float64:
		return int64(v), true
	case int64:
		return v, true
	case json.Number:
		n, err := v.Int64()
		return n, err == nil
	default:
		return 0, false
	}
}

func imageNumber(v interface{}) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	default:
		return 0
	}
}
