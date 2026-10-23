package proxy

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	metricsobs "github.com/Delpak/shahrah/services/control-plane/internal/observability/metrics"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

const mistralOCRReservationPages = 100

// MistralOCRHandler implements the native Mistral OCR contract without
// exposing an arbitrary provider path.
type MistralOCRHandler struct {
	llmRouter  *router.Router
	pipeline   *RequestPipeline
	billing    budgeter
	logger     *zap.Logger
	resources  providerNativeResourceStore
	logs       usageLogger
	emitMetric func(input metricsobs.BuildInput)
}

func NewMistralOCRHandlerWithDependencies(deps BaseHandlerDependencies) *MistralOCRHandler {
	return &MistralOCRHandler{llmRouter: deps.Router, pipeline: deps.requestPipeline(), billing: deps.Billing, logger: deps.Logger, resources: deps.ProviderResources, logs: deps.logger(), emitMetric: deps.metricEmitter()}
}

func (h *MistralOCRHandler) Handle(c *fiber.Ctx) error {
	startedAt := time.Now()

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

	if err := validateMistralOCRRequest(request); err != nil {
		return writeAPIErrorDetail(c, fiber.StatusBadRequest, err.Error(), "invalid_request_error", "request", "invalid_request")
	}

	accountSelector := strings.TrimSpace(c.Get("x-godarai-provider-account"))

	if document, ok := request["document"].(map[string]interface{}); ok &&
		strings.EqualFold(stringFromMap(document, "type"), "file") {
		if h.resources == nil {
			return writeAPIError(c, fiber.StatusServiceUnavailable, "Mistral file mapping is unavailable", "routing_error")
		}

		publicID := strings.TrimSpace(stringFromMap(document, "file_id"))

		mapping, mappingErr := h.resources.GetProviderNativeResourceMapping(c.Context(), currentTenant.ID, mistralProviderName, mistralFileResourceKind, publicID)
		if mappingErr != nil || mapping == nil {
			return writeAPIError(c, fiber.StatusNotFound, "Referenced Mistral file was not found", "not_found_error")
		}

		if accountSelector != "" && !strings.EqualFold(accountSelector, mapping.ProviderAccountID) {
			return writeAPIError(c, fiber.StatusBadRequest, "Referenced file belongs to a different provider account", "invalid_request_error")
		}

		accountSelector = mapping.ProviderAccountID
		document["file_id"] = mapping.UpstreamID

		encodedBody, marshalErr := json.Marshal(request)
		if marshalErr != nil {
			return writeAPIError(c, fiber.StatusBadRequest, "Failed to encode OCR request", "invalid_request_error")
		}

		body = encodedBody
	}

	if access := modelAccessForType(authInfo, requestedModel, platform.ModelTypeOCR); !access.Allowed {
		return writeAPIError(c, fiber.StatusForbidden, "Model access denied or unsupported for OCR", "authorization_error")
	}

	pageEstimate := mistralOCRPageEstimate(request)

	dimension := "ocr_input_pages"
	if mistralOCRUsesAnnotations(request) {
		dimension = "ocr_annotated_pages"
	}

	estimated := platform.UsageCostBreakdown{Model: requestedModel, Endpoint: "/v1/ocr", RequestType: "ocr", Metadata: map[string]interface{}{"estimated_pages": pageEstimate, "pricing_dimension": dimension}}

	var err error

	estimated, err = priceMeteredUsage(c.Context(), h.pipeline, billingEstimateModel(requestedModel), dimension, strconv.Itoa(pageEstimate), estimated)
	if err != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Mistral OCR pricing is unavailable", "billing_error")
	}

	if estimated.ModelRuntimeProfile == nil || !estimated.ModelRuntimeProfile.Profile.AllowsEndpoint("ocr") {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Mistral OCR runtime profile is unavailable", "runtime_profile_unavailable")
	}

	for parameter := range request {
		if !estimated.ModelRuntimeProfile.Profile.AllowsParameter("ocr", parameter) {
			return writeAPIErrorDetail(c, fiber.StatusBadRequest, fmt.Sprintf("parameter %q is not supported by this Mistral OCR model", parameter), "invalid_request_error", parameter, "unsupported_parameter")
		}
	}

	rateLimit, err := h.billing.CheckRateLimit(c.Context(), authInfo, requestedModel)
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
		return writeAPIError(c, fiber.StatusServiceUnavailable, "No available deployment for Mistral OCR", "routing_error")
	}

	if !isMistralProvider(deployment.Provider) {
		return writeAPIError(c, fiber.StatusNotImplemented, "OCR is only enabled for the Mistral AI provider", "invalid_request_error")
	}

	if accountSelector != "" && !strings.EqualFold(accountSelector, deployment.ProviderAccountID) &&
		!strings.EqualFold(accountSelector, deployment.ProviderAccountName) {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Mistral provider-account affinity could not be preserved", "routing_error")
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
			h.pipeline.Release(c.Context(), reservation, "ocr_request_not_committed")
		}
	}()

	outboundModel := strings.TrimSpace(deployment.Model)
	if outboundModel == "" || outboundModel == "*" {
		outboundModel = providerModelID(requestedModel)
	}

	payload, err := rewriteOpenAIJSONModel(body, providerModelID(outboundModel))
	if err != nil {
		return writeAPIError(c, fiber.StatusBadRequest, "Failed to encode OCR request", "invalid_request_error")
	}

	tunnelRequest := platform.TunnelRequest{
		RequestID: requestID, TenantID: currentTenant.ID, TraceID: requestID,
		Operation: platform.ProviderOperationMistralOCR, Deployment: toContractDeployment(deployment),
		Headers: map[string]string{"Content-Type": "application/json"}, Body: payload,
		RequestMetadata:      map[string]interface{}{"model": requestedModel, "provider_endpoint_path": "/ocr", "estimated_pages": pageEstimate},
		ModelRuntimeProfiles: []platform.ModelRuntimeProfileBinding{*estimated.ModelRuntimeProfile},
	}
	tunnelRequest.RequestMetadata = enrichRequestMetadataForDeployment(tunnelRequest.RequestMetadata, tunnelRequest.Deployment)

	response, err := h.pipeline.Dispatch(c.Context(), tunnelRequest)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, "Mistral OCR tunnel request failed", "api_error")
	}

	if response.StatusCode >= fiber.StatusBadRequest {
		return writeCompatibleProviderError(c, deployment.Provider, response.StatusCode, response.Body, response.Headers)
	}

	actualPages := mistralOCRResponsePages(response.Body)
	if actualPages <= 0 {
		actualPages = pageEstimate
	}

	finalCost, err := priceMeteredUsage(c.Context(), h.pipeline, billingModelForDeployment(requestedModel, tunnelRequest.Deployment), dimension, strconv.Itoa(actualPages), estimated)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, "Failed to finalize Mistral OCR usage", "billing_error")
	}

	if err := h.pipeline.Settle(c.Context(), SettlementRequest{Reservation: reservation, TenantID: currentTenant.ID, Actual: finalCost, RequestID: requestID, BillingMode: deployment.BillingMode}); err != nil {
		return writeBudgetCheckError(c, err)
	}

	committed = true
	responseTimeMS := int(time.Since(startedAt).Milliseconds())
	logEntry := buildRequestLogEntry(c, authInfo, tunnelRequest.Deployment, requestedModel, "/v1/ocr", "ocr", "success", response.StatusCode, startedAt, responseTimeMS)
	logEntry.RequestID = requestID
	logEntry.TraceID = requestID
	logEntry.CostUSD = finalCost.TotalCostUSD
	logEntry.CostIRR = finalCost.TotalCostIRR
	logEntry.RequestBody = map[string]interface{}{"model": requestedModel, "estimated_pages": pageEstimate, "annotated": dimension == "ocr_annotated_pages"}
	logEntry.ResponseBody = map[string]interface{}{"processed_pages": actualPages}
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
			CostUSD: finalCost.TotalCostUSD, Metadata: map[string]interface{}{"endpoint": "/v1/ocr", "pages": actualPages, "pricing_dimension": dimension},
		})
	}

	forwardProviderHeaders(c, response.Headers)
	setTraceHeaders(c, requestID)
	setCostHeaders(c, finalCost)

	return c.Status(response.StatusCode).Send(response.Body)
}

func mistralOCRPageEstimate(request map[string]interface{}) int {
	if pages, ok := request["pages"].([]interface{}); ok && len(pages) > 0 {
		return len(pages)
	}

	if pages, ok := request["pages"].(string); ok {
		if count, err := mistralOCRPageSelectionCount(pages); err == nil && count > 0 {
			return count
		}
	}

	return mistralOCRReservationPages
}

func validateMistralOCRRequest(request map[string]interface{}) error {
	document, ok := request["document"].(map[string]interface{})
	if !ok || len(document) == 0 {
		return fmt.Errorf("document is required")
	}

	documentType := strings.ToLower(strings.TrimSpace(stringFromMap(document, "type")))

	var documentField string

	switch documentType {
	case "file":
		documentField = "file_id"
	case "document_url":
		documentField = "document_url"
	case "image_url":
		documentField = "image_url"
	default:
		return fmt.Errorf("document.type must be one of file, document_url, or image_url")
	}

	if strings.TrimSpace(stringFromMap(document, documentField)) == "" {
		return fmt.Errorf("document.%s is required for document.type=%s", documentField, documentType)
	}

	if prompt, exists := request["document_annotation_prompt"]; exists && prompt != nil {
		if _, ok := prompt.(string); !ok {
			return fmt.Errorf("document_annotation_prompt must be a string")
		}

		if format, exists := request["document_annotation_format"]; !exists || format == nil {
			return fmt.Errorf("document_annotation_format is required when document_annotation_prompt is provided")
		}
	}

	for _, key := range []string{"bbox_annotation_format", "document_annotation_format"} {
		if value, exists := request[key]; exists && value != nil {
			if _, ok := value.(map[string]interface{}); !ok {
				return fmt.Errorf("%s must be an object", key)
			}
		}
	}

	for _, key := range []string{"include_image_base64", "extract_header", "extract_footer", "include_blocks"} {
		if value, exists := request[key]; exists && value != nil {
			if _, ok := value.(bool); !ok {
				return fmt.Errorf("%s must be a boolean", key)
			}
		}
	}

	for _, key := range []string{"image_limit", "image_min_size"} {
		if value, exists := request[key]; exists && value != nil {
			number, ok := value.(float64)
			if !ok || number < 0 || number != float64(int64(number)) {
				return fmt.Errorf("%s must be a non-negative integer", key)
			}
		}
	}

	if value, exists := request["table_format"]; exists && value != nil {
		format, ok := value.(string)
		if !ok || (format != "markdown" && format != "html") {
			return fmt.Errorf("table_format must be markdown or html")
		}
	}

	if value, exists := request["confidence_scores_granularity"]; exists && value != nil {
		granularity, ok := value.(string)
		if !ok || (granularity != "word" && granularity != "page") {
			return fmt.Errorf("confidence_scores_granularity must be word or page")
		}
	}

	if pages, exists := request["pages"]; exists && pages != nil {
		switch value := pages.(type) {
		case string:
			if _, err := mistralOCRPageSelectionCount(value); err != nil {
				return err
			}
		case []interface{}:
			if len(value) == 0 {
				return fmt.Errorf("pages must not be empty")
			}

			for _, page := range value {
				number, ok := page.(float64)
				if !ok || number < 0 || number != float64(int64(number)) {
					return fmt.Errorf("pages must contain non-negative integers")
				}
			}
		default:
			return fmt.Errorf("pages must be a page selection string or an array of integers")
		}
	}

	return nil
}

func mistralOCRPageSelectionCount(selection string) (int, error) {
	selection = strings.TrimSpace(selection)
	if selection == "" {
		return 0, fmt.Errorf("pages must not be empty")
	}

	count := 0

	for _, part := range strings.Split(selection, ",") {
		part = strings.TrimSpace(part)

		bounds := strings.Split(part, "-")
		if len(bounds) < 1 || len(bounds) > 2 || strings.TrimSpace(bounds[0]) == "" {
			return 0, fmt.Errorf("pages must contain comma-separated page numbers or ranges")
		}

		start, err := strconv.Atoi(strings.TrimSpace(bounds[0]))
		if err != nil || start < 0 {
			return 0, fmt.Errorf("pages must contain non-negative page numbers")
		}

		end := start
		if len(bounds) == 2 {
			end, err = strconv.Atoi(strings.TrimSpace(bounds[1]))
			if err != nil || end < start {
				return 0, fmt.Errorf("pages contains an invalid range")
			}
		}

		if end-start > 100000 || count > 100000-(end-start+1) {
			return 0, fmt.Errorf("pages selection is too large")
		}

		count += end - start + 1
	}

	return count, nil
}

func mistralOCRUsesAnnotations(request map[string]interface{}) bool {
	for _, key := range []string{"bbox_annotation_format", "document_annotation_format", "document_annotation_prompt"} {
		if value, exists := request[key]; exists && value != nil {
			return true
		}
	}

	return false
}

func mistralOCRResponsePages(body []byte) int {
	var response map[string]interface{}
	if json.Unmarshal(body, &response) != nil {
		return 0
	}

	if pages, ok := response["pages"].([]interface{}); ok {
		return len(pages)
	}

	for _, usageKey := range []string{"usage_info", "usage"} {
		if usage, ok := response[usageKey].(map[string]interface{}); ok {
			if pages, ok, _ := numberFromMap(usage, "pages_processed"); ok && pages > 0 { //nolint:errcheck // Optional metadata uses the zero value when absent or mismatched.
				return int(pages)
			}
		}
	}

	return 0
}
