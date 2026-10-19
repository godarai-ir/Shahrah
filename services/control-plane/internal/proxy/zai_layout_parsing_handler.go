package proxy

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	metricsobs "github.com/Delpak/shahrah/services/control-plane/internal/observability/metrics"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

const zaiOCRReservationTokens = 2_000_000

// ZAILayoutParsingHandler implements Z.ai's native layout-parsing contract.
type ZAILayoutParsingHandler struct {
	llmRouter        *router.Router
	pipeline         *RequestPipeline
	billing          budgeter
	logger           *zap.Logger
	logs             usageLogger
	emitMetric       func(input metricsobs.BuildInput)
	providerFeatures ProviderFeatureConfig
}

func NewZAILayoutParsingHandlerWithDependencies(deps BaseHandlerDependencies) *ZAILayoutParsingHandler {
	return &ZAILayoutParsingHandler{
		llmRouter: deps.Router, pipeline: deps.requestPipeline(), billing: deps.Billing,
		logger: deps.Logger, logs: deps.logger(), emitMetric: deps.metricEmitter(),
		providerFeatures: deps.ProviderFeatures,
	}
}

func (h *ZAILayoutParsingHandler) Handle(c *fiber.Ctx) error {
	startedAt := time.Now()

	if h.providerFeatures.Configured &&
		(!h.providerFeatures.ZAIEnabled || !h.providerFeatures.ZAILayoutParsingEnabled) {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Z.ai layout parsing is disabled", "provider_disabled")
	}

	authInfo := auth.GetAuthInfo(c)

	currentTenant := tenant.FromContext(c)
	if authInfo == nil || currentTenant == nil {
		return writeAPIError(c, fiber.StatusUnauthorized, "Authentication and tenant resolution are required", "authentication_error")
	}

	authInfo.TenantID = currentTenant.ID

	if !isJSONContentType(c.Get(fiber.HeaderContentType)) {
		return writeAPIErrorDetail(c, fiber.StatusBadRequest, "application/json content type is required", "invalid_request_error", "request", "unsupported_content_type")
	}

	body := append([]byte(nil), c.Body()...)

	var request map[string]interface{}
	if err := json.Unmarshal(body, &request); err != nil || request == nil {
		return writeAPIErrorDetail(c, fiber.StatusBadRequest, "invalid JSON request body", "invalid_request_error", "request", "invalid_json")
	}

	requestedModel := normalizeModelIdentifier(stringFromMap(request, "model"))
	if requestedModel == "" {
		return writeAPIErrorDetail(c, fiber.StatusBadRequest, "model is required", "invalid_request_error", "model", "invalid_model")
	}

	if !strings.EqualFold(providerModelID(requestedModel), "glm-ocr") {
		return writeAPIErrorDetail(c, fiber.StatusBadRequest, "layout parsing requires glm-ocr", "invalid_request_error", "model", "unsupported_model")
	}

	if err := validateZAILayoutParsingRequest(request); err != nil {
		return writeAPIErrorDetail(c, fiber.StatusBadRequest, err.Error(), "invalid_request_error", "request", "invalid_request")
	}

	if access := modelAccessForType(authInfo, requestedModel, platform.ModelTypeOCR); !access.Allowed {
		return writeAPIError(c, fiber.StatusForbidden, "Model access denied or unsupported for OCR", "authorization_error")
	}

	estimated := platform.UsageCostBreakdown{
		Model: requestedModel, Endpoint: c.Path(), RequestType: "ocr",
		Metadata: map[string]interface{}{"reservation_tokens": zaiOCRReservationTokens, "pricing_dimensions": []string{"input_tokens", "output_tokens"}},
	}

	var err error

	estimated, err = priceMeteredUsage(c.Context(), h.pipeline, billingEstimateModel(requestedModel), "input_tokens", strconv.Itoa(zaiOCRReservationTokens), estimated)
	if err != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Z.ai OCR pricing is unavailable", "billing_error")
	}

	if estimated.ModelRuntimeProfile == nil || !estimated.ModelRuntimeProfile.Profile.AllowsEndpoint("layout_parsing") {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Z.ai OCR runtime profile is unavailable", "runtime_profile_unavailable")
	}

	for parameter := range request {
		if !estimated.ModelRuntimeProfile.Profile.AllowsParameter("layout_parsing", parameter) {
			return writeAPIErrorDetail(c, fiber.StatusBadRequest, fmt.Sprintf("parameter %q is not supported by GLM-OCR", parameter), "invalid_request_error", parameter, "unsupported_parameter")
		}
	}

	rateLimit, err := h.billing.CheckRateLimit(c.Context(), authInfo, requestedModel, zaiOCRReservationTokens)
	setRateLimitHeaders(c, rateLimit)

	if err != nil {
		return writeAPIError(c, rateLimitErrorStatus(err), err.Error(), "rate_limit_error")
	}

	var deployment *router.Deployment
	for _, candidate := range modelRoutingCandidates(requestedModel) {
		deployment, err = h.llmRouter.Route(c.Context(), currentTenant.ID, candidate)
		if err == nil {
			break
		}
	}

	if err != nil || deployment == nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "No available Z.ai GLM-OCR deployment", "routing_error")
	}

	if !isZAIProvider(deployment.Provider) {
		return writeAPIError(c, fiber.StatusNotImplemented, "GLM-OCR requires the direct Z.ai provider", "invalid_request_error")
	}

	accountSelector := strings.TrimSpace(c.Get("x-godarai-provider-account"))
	if accountSelector != "" && !strings.EqualFold(accountSelector, deployment.ProviderAccountID) &&
		!strings.EqualFold(accountSelector, deployment.ProviderAccountName) {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Z.ai provider-account affinity could not be preserved", "routing_error")
	}

	estimated, err = applyDeploymentPricingOverride(c.Context(), h.pipeline, estimated, deployment)
	if err != nil {
		return writeBillingEstimateError(c, err)
	}

	requestID := strings.TrimSpace(c.Get("x-client-request-id"))
	if requestID == "" {
		requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
	}

	reservation, err := h.pipeline.Reserve(c.Context(), budgetRequestFromHTTP(c, authInfo, estimated, requestID, deployment.BillingMode))
	if err != nil {
		return writeBudgetCheckError(c, err)
	}

	committed := false
	defer func() {
		if !committed {
			h.pipeline.Release(c.Context(), reservation, "zai_ocr_request_not_committed")
		}
	}()

	outboundModel := strings.TrimSpace(deployment.Model)
	if outboundModel == "" || outboundModel == "*" {
		outboundModel = providerModelID(requestedModel)
	}

	payload, err := rewriteOpenAIJSONModel(body, providerModelID(outboundModel))
	if err != nil {
		return writeAPIError(c, fiber.StatusBadRequest, "Failed to encode layout-parsing request", "invalid_request_error")
	}

	tunnelRequest := platform.TunnelRequest{
		RequestID: requestID, TenantID: currentTenant.ID, TraceID: requestID,
		Operation: platform.ProviderOperationZAILayoutParsing, Deployment: toContractDeployment(deployment),
		Headers: map[string]string{"Content-Type": "application/json"}, Body: payload,
		RequestMetadata:      map[string]interface{}{"model": requestedModel, "provider_endpoint_path": "/layout_parsing"},
		ModelRuntimeProfiles: []platform.ModelRuntimeProfileBinding{*estimated.ModelRuntimeProfile},
	}
	tunnelRequest.RequestMetadata = enrichRequestMetadataForDeployment(tunnelRequest.RequestMetadata, tunnelRequest.Deployment)

	response, err := h.pipeline.Dispatch(c.Context(), tunnelRequest)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, "Z.ai layout-parsing tunnel request failed", "api_error")
	}

	if response.StatusCode >= fiber.StatusBadRequest {
		return writeCompatibleProviderError(c, deployment.Provider, response.StatusCode, response.Body, response.Headers)
	}

	usage, usageErr := zaiOCRUsage(response.Body)
	if usageErr != nil {
		return writeAPIError(c, fiber.StatusBadGateway, usageErr.Error(), "billing_error")
	}

	finalCost, err := h.billing.FinalizeFromUsage(c.Context(), billingModelForDeployment(requestedModel, tunnelRequest.Deployment), usage, estimated)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, "Failed to finalize Z.ai OCR usage", "billing_error")
	}

	if err := h.pipeline.Settle(c.Context(), SettlementRequest{Reservation: reservation, TenantID: currentTenant.ID, Actual: finalCost, RequestID: requestID, BillingMode: deployment.BillingMode}); err != nil {
		return writeBudgetCheckError(c, err)
	}

	committed = true

	responseTimeMS := int(time.Since(startedAt).Milliseconds())
	logEntry := buildRequestLogEntry(c, authInfo, tunnelRequest.Deployment, requestedModel, c.Path(), "ocr", "success", response.StatusCode, startedAt, responseTimeMS)
	logEntry.RequestID, logEntry.TraceID = requestID, requestID
	logEntry.CostUSD, logEntry.CostIRR = finalCost.TotalCostUSD, finalCost.TotalCostIRR
	logEntry.InputTokens, logEntry.OutputTokens, logEntry.TotalTokens = usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens
	logEntry.RequestBody = map[string]interface{}{"model": requestedModel, "file_redacted": true}
	logEntry.ResponseBody = map[string]interface{}{"usage_only": true}
	logRequestBestEffort(h.logger, h.logs, c.Context(), logEntry)
	logSpendBestEffort(h.logger, h.logs, c.Context(), platform.SpendLog{
		RequestID: requestID, TokenHash: authInfo.TokenHash, TenantID: authInfo.TenantID,
		TeamID: authInfo.TeamID, UserID: authInfo.UserID, Model: requestedModel,
		CostUSD: finalCost.TotalCostUSD, CostIRR: finalCost.TotalCostIRR, ResponseTimeMS: responseTimeMS,
		Status: "success", RequestMetadata: enrichPricingMetadata(tunnelRequest.RequestMetadata, authInfo, requestedModel),
		CompletionStart: startedAt, CompletionEnd: time.Now(),
	})

	if h.emitMetric != nil {
		h.emitMetric(metricsobs.BuildInput{
			TS: time.Now().UTC(), TenantID: authInfo.TenantID, RequestID: requestID,
			ModelName: requestedModel, ProviderName: deployment.Provider, ProviderModelName: deployment.Model,
			RequestType: "ocr", StatusCode: response.StatusCode, LatencyMS: responseTimeMS,
			CostUSD: finalCost.TotalCostUSD, Metadata: map[string]interface{}{"endpoint": c.Path(), "input_tokens": usage.PromptTokens, "output_tokens": usage.CompletionTokens},
		})
	}

	forwardProviderHeaders(c, response.Headers)
	setTraceHeaders(c, requestID)
	setCostHeaders(c, finalCost)

	return c.Status(response.StatusCode).Send(response.Body)
}

func validateZAILayoutParsingRequest(request map[string]interface{}) error {
	file, ok := request["file"].(string)
	if !ok || strings.TrimSpace(file) == "" {
		return fmt.Errorf("file is required and must be a URL or data URL")
	}

	if err := validateZAIOCRFile(file); err != nil {
		return err
	}

	for _, key := range []string{"return_crop_images", "need_layout_visualization"} {
		if value, exists := request[key]; exists {
			if _, valid := value.(bool); !valid {
				return fmt.Errorf("%s must be a boolean", key)
			}
		}
	}

	start, startSet, err := nonnegativeJSONInteger(request, "start_page_id")
	if err != nil {
		return err
	}

	end, endSet, err := nonnegativeJSONInteger(request, "end_page_id")
	if err != nil {
		return err
	}

	if startSet && endSet && (end < start || end-start >= 30) {
		return fmt.Errorf("page range must be ordered and contain at most 30 pages")
	}

	for _, key := range []string{"request_id", "user_id"} {
		if value, exists := request[key]; exists {
			text, valid := value.(string)
			if !valid || len(text) < 6 || (key == "request_id" && len(text) > 64) || (key == "user_id" && len(text) > 128) {
				return fmt.Errorf("%s length is invalid", key)
			}
		}
	}

	return nil
}

func validateZAIOCRFile(file string) error {
	file = strings.TrimSpace(file)
	if strings.HasPrefix(strings.ToLower(file), "data:") {
		comma := strings.IndexByte(file, ',')
		if comma < 0 || !strings.Contains(strings.ToLower(file[:comma]), ";base64") {
			return fmt.Errorf("file data URL must contain base64 data")
		}

		mediaType := strings.ToLower(strings.TrimSpace(strings.SplitN(file[5:comma], ";", 2)[0]))

		limit := 10 * 1024 * 1024
		if mediaType == "application/pdf" {
			limit = 50 * 1024 * 1024
		} else if mediaType != "image/jpeg" && mediaType != "image/jpg" && mediaType != "image/png" {
			return fmt.Errorf("file data URL must be PDF, JPG, JPEG, or PNG")
		}

		decoded, err := base64.StdEncoding.DecodeString(file[comma+1:])
		if err != nil {
			return fmt.Errorf("file data URL contains invalid base64")
		}

		if len(decoded) > limit {
			return fmt.Errorf("file exceeds the Z.ai size limit")
		}

		return nil
	}

	parsed, err := url.Parse(file)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil {
		return fmt.Errorf("file must be an HTTPS URL or supported data URL")
	}

	return nil
}

func nonnegativeJSONInteger(request map[string]interface{}, key string) (int, bool, error) {
	value, exists := request[key]
	if !exists {
		return 0, false, nil
	}

	number, ok := value.(float64)
	if !ok || number < 1 || number != float64(int(number)) {
		return 0, false, fmt.Errorf("%s must be a positive integer", key)
	}

	return int(number), true, nil
}

func zaiOCRUsage(body []byte) (*contractsopenai.Usage, error) {
	var response struct {
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("z.ai OCR response usage is invalid")
	}

	if response.Usage.PromptTokens < 0 || response.Usage.CompletionTokens < 0 ||
		response.Usage.PromptTokens+response.Usage.CompletionTokens == 0 {
		return nil, fmt.Errorf("z.ai OCR response usage is missing")
	}

	total := response.Usage.TotalTokens
	if total == 0 {
		total = response.Usage.PromptTokens + response.Usage.CompletionTokens
	}

	return &contractsopenai.Usage{
		PromptTokens: response.Usage.PromptTokens, CompletionTokens: response.Usage.CompletionTokens,
		TotalTokens: total,
	}, nil
}
