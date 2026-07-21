// Package proxy hosts OpenAI-compatible request handlers for the control plane.
package proxy

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"mime"
	"mime/multipart"
	"net/textproto"
	"strconv"
	"strings"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
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

// ImagesHandler serves image generation and edit requests.
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

// NewImagesHandler creates a new images handler.
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

// NewImagesHandlerWithDependencies creates an images handler from explicit dependencies.
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

// HandleGenerations serves image generation requests.
func (h *ImagesHandler) HandleGenerations(c *fiber.Ctx) error {
	return h.handle(c, imageOperation{Name: "generations", Path: "/images/generations", RequestType: "image_generations"})
}

// HandleEdits serves image edit requests.
func (h *ImagesHandler) HandleEdits(c *fiber.Ctx) error {
	return h.handle(c, imageOperation{Name: "edits", Path: "/images/edits", RequestType: "image_edits"})
}

// HandleVariations serves image variation requests.
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
		param, code := imageValidationErrorDetail(err)
		return writeAPIErrorDetail(c, fiber.StatusBadRequest, err.Error(), "invalid_request_error", param, code)
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

		denyReason := denyReasonForModel(authInfo)
		emitDenyCounter(h.logger, denyReason, requestedModel, currentTenant.ID)

		const message = "Model access denied or unsupported for image generation"
		logPreProviderRejection(h.logger, h.logs, c, authInfo, requestedModel, c.Path(), "images", fiber.StatusForbidden, "authorization_error", message, denyReason, startTime)

		return writeAPIError(c, fiber.StatusForbidden, message, "authorization_error")
	}

	authStart := time.Now()
	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "auth.validate", "internal", authStart, fiber.StatusOK, map[string]interface{}{"model": requestedModel})

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

	if validationErr := validateImageOperationDeployment(op, requestedModel, dep); validationErr != nil {
		return writeAPIErrorDetail(c, fiber.StatusBadRequest, validationErr.Error(), "invalid_request_error", "model", "unsupported_model")
	}

	tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "router.select_model", "internal", routeStart, fiber.StatusOK, map[string]interface{}{"provider": dep.Provider, "deployment_id": dep.ID})

	contractDeployment := toContractDeployment(dep)
	outboundModel := providerOutboundModel(requestedModel, dep)

	estimated := estimateImageCostFallback(requestedModel, metadata)
	if usage := estimatedImageUsage(outboundModel, metadata); usage != nil {
		finalizedEstimate, finalizeErr := h.billing.FinalizeImageUsage(c.Context(), outboundModel, usage, nil, estimated)
		if finalizeErr != nil {
			return writeAPIError(c, fiber.StatusBadGateway, "Failed to estimate image request cost", "billing_error")
		}

		estimated = finalizedEstimate
	}

	estimated.Metadata = metadata
	quotaStart := time.Now()
	rateLimit, err := h.billing.CheckRateLimit(c.Context(), authInfo, requestedModel, estimated.TotalTokens)
	setRateLimitHeaders(c, rateLimit)

	if err != nil {
		recordRateLimitRejection(c, h.logger, h.logs, h.emitMetric, authInfo, requestedModel, "/v1"+op.Path, op.RequestType, startTime, quotaStart, rateLimit, err)
		return writeAPIError(c, rateLimitErrorStatus(err), err.Error(), "rate_limit_error")
	}

	requestID := tracing.TraceIDFromContext(c)
	if requestID == "" {
		requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
	}

	budgetReservation, budgetResponseHandled, err := reserveBudgetForRequest(c, h.billing, authInfo, estimated, requestID, deploymentBillingMode(contractDeployment))
	if err != nil || budgetResponseHandled {
		return err
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

	outboundBody := body
	outboundContentType := contentType

	if isJSONContentType(contentType) {
		outboundBody, err = rewriteJSONModel(body, outboundModel)
		if err != nil {
			return writeAPIError(c, fiber.StatusBadRequest, "Invalid request body", "invalid_request_error")
		}
	} else if isMultipartContentType(contentType) {
		outboundBody, outboundContentType, err = rewriteMultipartModel(c, contentType, outboundModel)
		if err != nil {
			return writeAPIError(c, fiber.StatusBadRequest, "Invalid multipart request body", "invalid_request_error")
		}
	}

	if outboundContentType == "" {
		outboundContentType = "application/json"
	}

	requestMetadata := map[string]interface{}{
		"tenant_slug":               apiheaders.GetTenant(c),
		"model":                     requestedModel,
		"provider_endpoint_path":    op.Path,
		"operation":                 op.Name,
		"usage_reconciliation_mode": "image_usage_reconciliation",
	}
	for k, v := range metadata {
		requestMetadata[k] = v
	}

	tunnelReq := platform.TunnelRequest{
		RequestID:  requestID,
		TenantID:   currentTenant.ID,
		TraceID:    requestID,
		Stream:     false,
		Deployment: contractDeployment,
		Headers: map[string]string{
			"Content-Type": outboundContentType,
		},
		Body:            outboundBody,
		RequestMetadata: requestMetadata,
	}
	tunnelReq.RequestMetadata = enrichRequestMetadataForDeployment(tunnelReq.RequestMetadata, tunnelReq.Deployment)

	err = h.handleSync(c, requestedModel, authInfo, estimated, tunnelReq, budgetReservation, op, startTime, tenantSlug, outboundContentType)
	if err == nil {
		releaseBudgetOnExit = false
	}

	return err
}

func (h *ImagesHandler) handleSync(
	c *fiber.Ctx,
	requestedModel string,
	authInfo *platform.AuthContext,
	estimated platform.UsageCostBreakdown,
	tunnelReq platform.TunnelRequest,
	budgetReservation *platform.BudgetReservation,
	op imageOperation,
	start time.Time,
	tenantSlug string,
	contentType string,
) error {
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

	providerCostUSDTicks := imageUsageCostTicksPtr(resp.Body)

	finalCost, err := h.billing.FinalizeImageUsage(c.Context(), billingModelForDeployment(requestedModel, tunnelReq.Deployment), imageUsageToOpenAI(resp.Body), providerCostUSDTicks, estimated)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, "Failed to finalize image usage", "billing_error")
	}

	metadata := enrichPricingMetadata(withImageResponseMetadata(tunnelReq.RequestMetadata, resp.Body), auth.GetAuthInfo(c), requestedModel)

	if debitErr := commitBudgetOrDebit(c.Context(), h.logger, h.billing, h.wallets, budgetReservation, authInfo.TenantID, finalCost, tunnelReq.RequestID, deploymentBillingMode(tunnelReq.Deployment)); debitErr != nil {
		return writeBudgetCheckError(c, debitErr)
	}

	responseTimeMS := int(time.Since(start).Milliseconds())
	logEntry := buildRequestLogEntry(c, authInfo, tunnelReq.Deployment, requestedModel, "/v1"+op.Path, op.RequestType, "success", resp.StatusCode, start, responseTimeMS)
	logEntry.RequestID = tunnelReq.RequestID
	logEntry.TraceID = tunnelReq.TraceID
	logEntry.CostUSD = finalCost.TotalCostUSD
	logEntry.CostIRR = finalCost.TotalCostIRR
	logEntry.InputTokens = finalCost.InputTokens
	logEntry.OutputTokens = finalCost.OutputTokens
	logEntry.TotalTokens = finalCost.TotalTokens
	logEntry.RequestBody = map[string]interface{}{"model": requestedModel, "operation": op.Name, "n": tunnelReq.RequestMetadata["n"], "size": tunnelReq.RequestMetadata["size"], "quality": tunnelReq.RequestMetadata["quality"]}
	logEntry.ResponseBody = buildTraceImageResponsePayload(resp.Body)
	logRequestBestEffort(h.logger, h.logs, c.Context(), logEntry)
	logSpendBestEffort(h.logger, h.logs, c.Context(), platform.SpendLog{
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
		Metadata: enrichMetricMetadataForDeployment(map[string]interface{}{
			"endpoint":    "/v1" + op.Path,
			"key_type":    authInfo.KeyType,
			"team_id":     authInfo.TeamID,
			"org_unit_id": authInfo.OrgUnitID,
			"tenant_slug": tenantSlug,
			"operation":   op.Name,
		}, tunnelReq.Deployment),
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

		model := ""
		if rawModel, ok := raw["model"].(string); ok {
			model = rawModel
		}

		fillImageJSONMetadata(metadata, raw, op)

		if strings.TrimSpace(model) == "" {
			return "", metadata, fmt.Errorf("model is required")
		}

		if err := validateImageOperationMetadata(op, raw["prompt"], raw["image"]); err != nil {
			return "", metadata, err
		}

		return strings.TrimSpace(model), metadata, nil
	}

	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(contentType)), "multipart/form-data") {
		model := strings.TrimSpace(c.FormValue("model"))
		fillImageMultipartMetadata(c, metadata, op)

		if model == "" {
			return "", metadata, fmt.Errorf("model is required")
		}

		if err := validateImageOperationMetadata(op, c.FormValue("prompt"), multipartImagePresence(c)); err != nil {
			return "", metadata, err
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

	for _, key := range []string{"size", "quality", "response_format", "output_format", "moderation", "background", "input_fidelity"} {
		if v, ok := raw[key].(string); ok && strings.TrimSpace(v) != "" {
			metadata[key] = strings.TrimSpace(v)
		}
	}

	for _, key := range []string{"output_compression", "partial_images"} {
		if n := imageNumber(raw[key]); n > 0 {
			metadata[key] = n
		}
	}

	if n := imageNumber(raw["n"]); n > 0 {
		metadata["n"] = n
	}

	if _, ok := raw["image"]; ok {
		metadata["image_source_type"] = traceImageSourceType(raw["image"])
		if count := jsonImageCount(raw["image"]); count > 0 {
			metadata["image_count"] = count
		}
	}

	if _, ok := raw["mask"]; ok {
		metadata["mask_present"] = true
	}

	if op.Name == "edits" || op.Name == "variations" {
		metadata["file_count"] = 0
	}
}

func validateImageOperationMetadata(op imageOperation, prompt interface{}, image interface{}) error {
	if op.Name == "generations" {
		if promptText, ok := prompt.(string); !ok || strings.TrimSpace(promptText) == "" {
			return fmt.Errorf("prompt is required")
		}

		return nil
	}

	if op.Name == "edits" {
		if promptText, ok := prompt.(string); !ok || strings.TrimSpace(promptText) == "" {
			return fmt.Errorf("prompt is required")
		}
	}

	if !imageProvided(image) {
		return fmt.Errorf("image is required")
	}

	return nil
}

func validateImageOperationDeployment(op imageOperation, requestedModel string, dep *router.Deployment) error {
	if op.Name != "variations" || dep == nil || !strings.EqualFold(strings.TrimSpace(dep.Provider), "openai") {
		return nil
	}

	model := strings.TrimSpace(dep.Model)
	if model == "" || model == "*" {
		model = providerModelID(requestedModel)
	}

	model = strings.ToLower(providerModelID(model))
	if model != "dall-e-2" {
		return fmt.Errorf("OpenAI image variations only support dall-e-2; use /v1/images/edits or image references for GPT Image models")
	}

	return nil
}

func imageProvided(image interface{}) bool {
	switch v := image.(type) {
	case nil:
		return false
	case string:
		return strings.TrimSpace(v) != ""
	case []interface{}:
		return len(v) > 0
	case map[string]interface{}:
		return len(v) > 0
	default:
		return true
	}
}

func jsonImageCount(image interface{}) int {
	switch v := image.(type) {
	case nil:
		return 0
	case string:
		if strings.TrimSpace(v) == "" {
			return 0
		}

		return 1
	case []interface{}:
		count := 0

		for _, item := range v {
			if imageProvided(item) {
				count++
			}
		}

		return count
	case map[string]interface{}:
		if len(v) == 0 {
			return 0
		}

		return 1
	default:
		return 1
	}
}

func imageValidationErrorDetail(err error) (string, string) {
	message := ""
	if err != nil {
		message = strings.ToLower(strings.TrimSpace(err.Error()))
	}

	switch {
	case strings.Contains(message, "prompt"):
		return "prompt", "invalid_prompt"
	case strings.Contains(message, "image"):
		return "image", "invalid_image"
	case strings.Contains(message, "content type"):
		return "content_type", "unsupported_content_type"
	default:
		return "model", "invalid_model"
	}
}

func multipartImagePresence(c *fiber.Ctx) interface{} {
	if c == nil {
		return nil
	}

	form, err := c.MultipartForm()
	if err != nil || form == nil {
		return nil
	}

	if count := multipartImageFileCount(c); count > 0 {
		return count
	}

	return nil
}

func fillImageMultipartMetadata(c *fiber.Ctx, metadata map[string]interface{}, op imageOperation) {
	for _, key := range []string{"size", "quality", "response_format", "output_format", "moderation", "background", "output_compression", "partial_images", "input_fidelity"} {
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

	imageFileCount := multipartImageFileCount(c)

	metadata["file_count"] = fileCount
	if imageFileCount > 0 {
		metadata["image_count"] = imageFileCount
	}

	if len(form.File["mask"]) > 0 {
		metadata["mask_present"] = true
	}

	if fileCount > 0 {
		metadata["image_source_type"] = "multipart_file"
	}

	_ = op
}

func multipartImageFileCount(c *fiber.Ctx) int {
	if c == nil {
		return 0
	}

	form, err := c.MultipartForm()
	if err != nil || form == nil {
		return 0
	}

	return len(form.File["image"]) + len(form.File["image[]"])
}

func isJSONContentType(contentType string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(contentType)), "application/json")
}

func isMultipartContentType(contentType string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(contentType)), "multipart/form-data")
}

func providerOutboundModel(requestedModel string, dep *router.Deployment) string {
	outboundModel := ""
	if dep != nil {
		outboundModel = strings.TrimSpace(dep.Model)
	}

	if outboundModel == "" || outboundModel == "*" {
		outboundModel = providerModelID(requestedModel)
	}

	return providerModelID(outboundModel)
}

func rewriteJSONModel(body []byte, model string) ([]byte, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	raw["model"] = model

	return json.Marshal(raw)
}

func rewriteMultipartModel(c *fiber.Ctx, contentType string, model string) ([]byte, string, error) {
	if c == nil {
		return nil, "", fmt.Errorf("request context is required")
	}

	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil, "", err
	}

	if !strings.EqualFold(mediaType, "multipart/form-data") || strings.TrimSpace(params["boundary"]) == "" {
		return nil, "", fmt.Errorf("multipart boundary is required")
	}

	form, err := c.MultipartForm()
	if err != nil {
		return nil, "", err
	}

	var buf bytes.Buffer

	writer := multipart.NewWriter(&buf)

	for key, values := range form.Value {
		if key == "model" {
			continue
		}

		for _, value := range values {
			if err := writer.WriteField(key, value); err != nil {
				return nil, "", errors.Join(err, writer.Close())
			}
		}
	}

	if err := writer.WriteField("model", model); err != nil {
		return nil, "", errors.Join(err, writer.Close())
	}

	for key, files := range form.File {
		for _, header := range files {
			if err := copyMultipartFile(writer, key, header); err != nil {
				return nil, "", errors.Join(err, writer.Close())
			}
		}
	}

	if err := writer.Close(); err != nil {
		return nil, "", err
	}

	return buf.Bytes(), writer.FormDataContentType(), nil
}

func copyMultipartFile(writer *multipart.Writer, field string, header *multipart.FileHeader) (returnErr error) {
	if writer == nil || header == nil {
		return fmt.Errorf("multipart file is required")
	}

	src, err := header.Open()
	if err != nil {
		return err
	}
	defer func() {
		returnErr = errors.Join(returnErr, src.Close())
	}()

	partHeader := make(textproto.MIMEHeader, len(header.Header)+1)
	for key, values := range header.Header {
		copied := make([]string, len(values))
		copy(copied, values)
		partHeader[key] = copied
	}

	if partHeader.Get("Content-Disposition") == "" {
		partHeader.Set("Content-Disposition", fmt.Sprintf("form-data; name=%q; filename=%q", escapeMultipartQuote(field), escapeMultipartQuote(header.Filename)))
	}

	dst, err := writer.CreatePart(partHeader)
	if err != nil {
		return err
	}

	_, err = io.Copy(dst, src)

	return err
}

func escapeMultipartQuote(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, `"`, `\"`)

	return value
}

func estimateImageCostFallback(model string, _ map[string]interface{}) platform.UsageCostBreakdown {
	return platform.UsageCostBreakdown{
		Model:       model,
		TotalTokens: 0,
	}
}

func estimatedImageUsage(model string, metadata map[string]interface{}) *contractsopenai.Usage {
	if len(metadata) == 0 {
		return nil
	}

	inputTokens := 0
	if rawPromptLength, ok := metadata["prompt_length"]; ok {
		inputTokens = estimateTextTokensFromLength(rawPromptLength)
	}

	if inputTokens == 0 {
		inputTokens = 100
	}

	inputTokens += estimatedImageInputTokens(model, metadata)

	imageCount := imageCountFromMetadata(metadata)
	outputTokens := estimatedImageOutputTokens(model, metadata) * imageCount
	outputTokens += partialImageCountFromMetadata(metadata) * 100

	return &contractsopenai.Usage{
		PromptTokens:     inputTokens,
		CompletionTokens: outputTokens,
		TotalTokens:      inputTokens + outputTokens,
	}
}

func imageUsageToOpenAI(body []byte) *contractsopenai.Usage {
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil
	}

	usage, ok := raw["usage"].(map[string]interface{})
	if !ok {
		return nil
	}

	inputTokens := intFromImageUsage(usage, "input_tokens")
	if inputTokens == 0 {
		inputTokens = intFromImageUsage(usage, "prompt_tokens")
	}

	outputTokens := intFromImageUsage(usage, "output_tokens")
	if outputTokens == 0 {
		outputTokens = intFromImageUsage(usage, "completion_tokens")
	}

	totalTokens := intFromImageUsage(usage, "total_tokens")
	if totalTokens == 0 {
		totalTokens = inputTokens + outputTokens
	}

	if inputTokens == 0 && outputTokens == 0 && totalTokens == 0 {
		return nil
	}

	return &contractsopenai.Usage{
		PromptTokens:     inputTokens,
		CompletionTokens: outputTokens,
		TotalTokens:      totalTokens,
	}
}

func estimateTextTokensFromLength(raw interface{}) int {
	switch v := raw.(type) {
	case int:
		return int(math.Ceil(float64(v) / 4.0))
	case int64:
		return int(math.Ceil(float64(v) / 4.0))
	case float64:
		return int(math.Ceil(v / 4.0))
	case json.Number:
		i, err := v.Int64()
		if err != nil {
			return 0
		}

		return int(math.Ceil(float64(i) / 4.0))
	default:
		return 0
	}
}

func imageCountFromMetadata(metadata map[string]interface{}) int {
	count := 1

	if n, ok := metadata["n"]; ok {
		switch v := n.(type) {
		case int:
			count = v
		case int64:
			count = int(v)
		case float64:
			count = int(v)
		case json.Number:
			if parsed, err := v.Int64(); err == nil {
				count = int(parsed)
			}
		}
	}

	if count < 1 {
		return 1
	}

	return count
}

func estimatedImageInputTokens(model string, metadata map[string]interface{}) int {
	imageCount := imageInputCountFromMetadata(metadata)
	if imageCount == 0 {
		return 0
	}

	tokensPerImage := 1024
	if strings.Contains(strings.ToLower(providerModelID(model)), "gpt-image-2") {
		tokensPerImage = 1536
	}

	return imageCount * tokensPerImage
}

func imageInputCountFromMetadata(metadata map[string]interface{}) int {
	count := imageNumber(metadata["image_count"])
	if count < 0 {
		return 0
	}

	return count
}

func estimatedImageOutputTokens(model string, metadata map[string]interface{}) int {
	model = strings.ToLower(providerModelID(model))
	quality := strings.ToLower(strings.TrimSpace(fmt.Sprint(metadata["quality"])))
	size := strings.ToLower(strings.TrimSpace(fmt.Sprint(metadata["size"])))

	if quality == "" || quality == "<nil>" || quality == "auto" {
		quality = "medium"
	}

	if size == "" || size == "<nil>" || size == "auto" {
		size = "1024x1024"
	}

	if strings.Contains(model, "gpt-image-2") {
		return estimatedGPTImage2OutputTokens(quality, size)
	}

	if model == "dall-e-2" {
		return estimatedDallE2OutputUnits(size)
	}

	switch quality {
	case "low":
		if isPortraitSize(size) {
			return 408
		}

		if isLandscapeSize(size) {
			return 400
		}

		return 272
	case "high", "hd":
		if isPortraitSize(size) {
			return 6240
		}

		if isLandscapeSize(size) {
			return 6208
		}

		return 4160
	default:
		if isPortraitSize(size) {
			return 1584
		}

		if isLandscapeSize(size) {
			return 1568
		}

		return 1056
	}
}

func estimatedDallE2OutputUnits(size string) int {
	switch size {
	case "256x256":
		return 800
	case "512x512":
		return 900
	default:
		return 1000
	}
}

func estimatedGPTImage2OutputTokens(quality string, size string) int {
	switch quality {
	case "low":
		if isPortraitSize(size) || isLandscapeSize(size) {
			return 167
		}

		return 196
	case "high":
		if isPortraitSize(size) || isLandscapeSize(size) {
			return 5500
		}

		return 7033
	default:
		if isPortraitSize(size) || isLandscapeSize(size) {
			return 1367
		}

		return 1767
	}
}

func isPortraitSize(size string) bool {
	parts := strings.Split(size, "x")
	if len(parts) != 2 {
		return false
	}

	width, wErr := strconv.Atoi(strings.TrimSpace(parts[0]))
	height, hErr := strconv.Atoi(strings.TrimSpace(parts[1]))

	return wErr == nil && hErr == nil && height > width
}

func isLandscapeSize(size string) bool {
	parts := strings.Split(size, "x")
	if len(parts) != 2 {
		return false
	}

	width, wErr := strconv.Atoi(strings.TrimSpace(parts[0]))
	height, hErr := strconv.Atoi(strings.TrimSpace(parts[1]))

	return wErr == nil && hErr == nil && width > height
}

func partialImageCountFromMetadata(metadata map[string]interface{}) int {
	if metadata == nil {
		return 0
	}

	count := imageNumber(metadata["partial_images"])
	if count < 0 {
		return 0
	}

	return count
}

func intFromImageUsage(usage map[string]interface{}, key string) int {
	switch v := usage[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	case json.Number:
		n, err := v.Int64()
		if err != nil {
			return 0
		}

		return int(n)
	default:
		return 0
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

func imageUsageCostTicksPtr(body []byte) *int64 {
	ticks, ok := imageUsageCostTicks(body)
	if !ok {
		return nil
	}

	return &ticks
}

func imageNumber(v interface{}) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case json.Number:
		i, err := n.Int64()
		if err != nil {
			return 0
		}

		return int(i)
	default:
		return 0
	}
}
