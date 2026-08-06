package proxy

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	billingusage "github.com/Delpak/shahrah/services/control-plane/internal/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// GeminiNativeHandler exposes only the finite Gemini Developer API operations
// registered in routes_proxy.go. It never accepts an arbitrary upstream URL.
type GeminiNativeHandler struct {
	deployments     anthropicDeploymentLister
	pipeline        *RequestPipeline
	billing         budgeter
	runtimeProfiles anthropicModelRuntimeProfileResolver
	logger          *zap.Logger
}

func NewGeminiNativeHandler(deps BaseHandlerDependencies, deployments anthropicDeploymentLister) *GeminiNativeHandler {
	return &GeminiNativeHandler{
		deployments: deployments, pipeline: deps.requestPipeline(), billing: deps.Billing,
		runtimeProfiles: deps.AnthropicRuntimeProfiles, logger: deps.Logger,
	}
}

func (h *GeminiNativeHandler) Handle(c *fiber.Ctx) error {
	authInfo := auth.GetAuthInfo(c)

	currentTenant := tenant.FromContext(c)
	if authInfo == nil || currentTenant == nil {
		return writeGeminiError(c, fiber.StatusUnauthorized, "UNAUTHENTICATED", "Authentication and tenant resolution are required")
	}

	authInfo.TenantID = currentTenant.ID

	operation, input, model, stream, err := geminiOperationFromRequest(c)
	if err != nil {
		return writeGeminiError(c, fiber.StatusBadRequest, "INVALID_ARGUMENT", err.Error())
	}

	if model != "" && !isAllowlistedGoogleGeminiModel(model) {
		return writeGeminiError(c, fiber.StatusBadRequest, "MODEL_NOT_ALLOWLISTED", "The requested Gemini model is not supported")
	}

	deployment, err := h.resolveDeployment(c.Context(), currentTenant.ID, model, strings.TrimSpace(c.Get("x-godarai-provider-account")))
	if err != nil {
		return writeGeminiError(c, fiber.StatusServiceUnavailable, "ROUTING_ERROR", err.Error())
	}

	requestID := strings.TrimSpace(c.Get("x-client-request-id"))
	if requestID == "" {
		requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
	}

	request := platform.TunnelRequest{
		RequestID: requestID, TenantID: currentTenant.ID, Stream: stream,
		Operation: operation, OperationInput: input, Deployment: *deployment,
		Headers: map[string]string{"content-type": c.Get(fiber.HeaderContentType)}, Body: append([]byte(nil), c.Body()...),
	}
	if operation == platform.ProviderOperationGeminiFilesCreate {
		return h.proxyRawFileUpload(c, request)
	}

	var (
		estimated   platform.UsageCostBreakdown
		reservation *platform.BudgetReservation
	)

	if geminiImmediateBillableOperation(operation) {
		if h.billing == nil {
			return writeGeminiError(c, fiber.StatusServiceUnavailable, "BILLING_UNAVAILABLE", "Gemini billing is unavailable")
		}

		estimated, err = h.estimateCost(c.Context(), operation, model, request.Body)
		if err != nil {
			return writeGeminiError(c, fiber.StatusServiceUnavailable, "PRICING_UNAVAILABLE", "Gemini pricing is unavailable")
		}

		reservation, err = h.pipeline.Reserve(c.Context(), budgetRequestFromHTTP(c, authInfo, estimated, requestID, deployment.BillingMode))
		if err != nil {
			return writeBudgetCheckError(c, err)
		}
	}

	profileModel := model
	if profileModel == "" && geminiOperationNeedsRuntimeProfile(operation) {
		profileModel = strings.TrimPrefix(strings.TrimSpace(deployment.Model), "models/")
	}

	if profileModel != "" && h.runtimeProfiles == nil {
		h.pipeline.Release(c.Context(), reservation, "runtime_profile_unavailable")
		return writeGeminiError(c, fiber.StatusServiceUnavailable, "RUNTIME_PROFILE_UNAVAILABLE", "The Gemini runtime profile is unavailable")
	}

	if profileModel != "" {
		request.ModelRuntimeProfiles, err = h.runtimeProfiles.ResolveModelRuntimeProfiles([]string{profileModel}, "")
		if err != nil {
			h.pipeline.Release(c.Context(), reservation, "runtime_profile_unavailable")
			return writeGeminiError(c, fiber.StatusServiceUnavailable, "RUNTIME_PROFILE_UNAVAILABLE", "The Gemini runtime profile is unavailable")
		}
	}

	if stream {
		return h.proxyStream(c, request, reservation, estimated, authInfo, profileModel)
	}

	response, err := h.pipeline.Dispatch(c.Context(), request)
	if err != nil {
		h.pipeline.Release(c.Context(), reservation, "tunnel_error")
		return writeGeminiError(c, fiber.StatusBadGateway, "UPSTREAM_UNAVAILABLE", "Gemini tunnel request failed")
	}

	if response.StatusCode >= fiber.StatusBadRequest {
		h.pipeline.Release(c.Context(), reservation, "provider_http_error")
	} else if reservation != nil {
		usage := parseGeminiNativeUsage(response.Body)

		finalized, finalizeErr := h.billing.FinalizeFromUsage(c.Context(), profileModel, usage, estimated)
		if finalizeErr != nil {
			h.pipeline.Release(c.Context(), reservation, "usage_finalize_error")
			return writeGeminiError(c, fiber.StatusBadGateway, "BILLING_ERROR", "Failed to finalize Gemini usage")
		}

		if err := h.pipeline.Settle(c.Context(), SettlementRequest{Reservation: reservation, TenantID: currentTenant.ID, Actual: finalized, RequestID: requestID, BillingMode: deployment.BillingMode}); err != nil {
			return writeGeminiError(c, fiber.StatusPaymentRequired, "BILLING_ERROR", err.Error())
		}
	}

	forwardProviderHeaders(c, response.Headers)

	if contentType := response.Headers["Content-Type"]; contentType != "" {
		c.Set(fiber.HeaderContentType, contentType)
	}

	return c.Status(response.StatusCode).Send(response.Body)
}

func geminiOperationNeedsRuntimeProfile(operation platform.ProviderOperation) bool {
	switch operation {
	case platform.ProviderOperationGeminiGenerateContent, platform.ProviderOperationGeminiStreamGenerateContent,
		platform.ProviderOperationGeminiCountTokens, platform.ProviderOperationGeminiEmbedContent,
		platform.ProviderOperationGeminiBatchEmbedContents, platform.ProviderOperationGeminiInteractionsCreate,
		platform.ProviderOperationGeminiInteractionsGet, platform.ProviderOperationGeminiInteractionsDelete,
		platform.ProviderOperationGeminiInteractionsCancel, platform.ProviderOperationGeminiCachesCreate,
		platform.ProviderOperationGeminiCachesList, platform.ProviderOperationGeminiCachesGet,
		platform.ProviderOperationGeminiCachesUpdate, platform.ProviderOperationGeminiCachesDelete,
		platform.ProviderOperationGeminiBatchesCreate:
		return true
	default:
		return false
	}
}

func (h *GeminiNativeHandler) resolveDeployment(ctx context.Context, tenantID, model, account string) (*platform.Deployment, error) {
	if h.deployments == nil {
		return nil, fmt.Errorf("gemini deployments are unavailable")
	}

	deployments, err := h.deployments.ListDeployments(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list Gemini deployments: %w", err)
	}

	var matched *platform.Deployment

	for index := range deployments {
		candidate := &deployments[index]

		provider := strings.ToLower(strings.TrimSpace(candidate.Provider))
		if provider != "google-gemini" && provider != "gemini" {
			continue
		}

		candidateModel := strings.TrimPrefix(strings.TrimSpace(candidate.Model), "models/")
		if !isAllowlistedGoogleGeminiModel(candidateModel) {
			continue
		}

		if model != "" && candidateModel != model {
			continue
		}

		if account != "" && account != candidate.ProviderAccountID && account != candidate.ConnectionID && account != candidate.ID {
			continue
		}

		if matched != nil && account == "" {
			return nil, fmt.Errorf("multiple Gemini provider accounts match; x-godarai-provider-account is required")
		}

		candidateCopy := *candidate
		matched = &candidateCopy
	}

	if matched != nil {
		return matched, nil
	}

	return nil, fmt.Errorf("no allowlisted Gemini deployment matches this operation")
}

func geminiImmediateBillableOperation(operation platform.ProviderOperation) bool {
	switch operation {
	case platform.ProviderOperationGeminiGenerateContent,
		platform.ProviderOperationGeminiStreamGenerateContent,
		platform.ProviderOperationGeminiEmbedContent,
		platform.ProviderOperationGeminiBatchEmbedContents,
		platform.ProviderOperationGeminiInteractionsCreate:
		return true
	default:
		return false
	}
}

func (h *GeminiNativeHandler) estimateCost(ctx context.Context, operation platform.ProviderOperation, model string, body []byte) (platform.UsageCostBreakdown, error) {
	if operation == platform.ProviderOperationGeminiEmbedContent || operation == platform.ProviderOperationGeminiBatchEmbedContents {
		return h.billing.EstimateEmbeddingCost(ctx, contractsopenai.EmbeddingRequest{Model: model, Input: string(body)})
	}

	request := contractsopenai.ChatCompletionRequest{
		Model:    model,
		Messages: []contractsopenai.ChatMessage{{Role: "user", Content: string(body)}},
	}

	var payload struct {
		GenerationConfig struct {
			MaxOutputTokens int `json:"maxOutputTokens"`
		} `json:"generationConfig"`
	}
	if json.Unmarshal(body, &payload) == nil && payload.GenerationConfig.MaxOutputTokens > 0 {
		request.MaxTokens = &payload.GenerationConfig.MaxOutputTokens
	}

	return h.billing.EstimateCost(ctx, request)
}

func (h *GeminiNativeHandler) proxyRawFileUpload(c *fiber.Ctx, request platform.TunnelRequest) error {
	if !h.pipeline.HasRawTransport() {
		return writeGeminiError(c, fiber.StatusServiceUnavailable, "TUNNEL_UNAVAILABLE", "Raw Gemini file upload is unavailable")
	}

	body := append([]byte(nil), c.Body()...)

	response, err := h.pipeline.DispatchRaw(c.Context(), request, bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return writeGeminiError(c, fiber.StatusBadGateway, "UPSTREAM_UNAVAILABLE", "Gemini file upload failed")
	}

	if response == nil || response.Body == nil {
		return writeGeminiError(c, fiber.StatusBadGateway, "UPSTREAM_UNAVAILABLE", "Gemini returned an empty upload response")
	}
	defer response.Body.Close() //nolint:errcheck // Cleanup is best-effort and cannot change the completed operation.

	result, err := io.ReadAll(io.LimitReader(response.Body, 2*1024*1024))
	if err != nil {
		return writeGeminiError(c, fiber.StatusBadGateway, "UPSTREAM_UNAVAILABLE", "Gemini returned an invalid upload response")
	}

	forwardProviderHeaders(c, response.Headers)

	if contentType := responseContentType(response.Headers); contentType != "" {
		c.Set(fiber.HeaderContentType, contentType)
	}

	return c.Status(response.StatusCode).Send(result)
}

func (h *GeminiNativeHandler) proxyStream(c *fiber.Ctx, request platform.TunnelRequest, reservation *platform.BudgetReservation, estimated platform.UsageCostBreakdown, authInfo *platform.AuthContext, model string) error {
	stream, err := h.pipeline.DispatchStream(c.Context(), request)
	if err != nil {
		h.pipeline.Release(c.Context(), reservation, "stream_open_failed")
		return writeGeminiError(c, fiber.StatusBadGateway, "UPSTREAM_UNAVAILABLE", "Gemini stream failed")
	}

	if stream.StatusCode() >= fiber.StatusBadRequest {
		h.pipeline.Release(c.Context(), reservation, "provider_http_error")
		return writeTunnelStreamHTTPError(c, h.logger, stream, nil)
	}

	forwardProviderHeaders(c, stream.Headers())
	c.Set(fiber.HeaderContentType, "text/event-stream")
	c.Context().SetBodyStreamWriter(func(writer *bufio.Writer) {
		defer closeStreamBestEffort(h.logger, stream)
		defer writer.Flush() //nolint:errcheck // A downstream disconnect is handled by terminating the stream.

		var usage *contractsopenai.Usage

		failed := false

		for {
			select {
			case chunk, ok := <-stream.Chunks():
				if !ok {
					if reservation != nil {
						if failed || usage == nil {
							h.pipeline.Release(context.Background(), reservation, "stream_incomplete")
						} else if finalized, finalizeErr := h.billing.FinalizeFromUsage(context.Background(), model, usage, estimated); finalizeErr != nil {
							h.pipeline.Release(context.Background(), reservation, "usage_finalize_error")
						} else {
							settlement := SettlementRequest{
								Reservation: reservation,
								TenantID:    authInfo.TenantID,
								Actual:      finalized,
								RequestID:   request.RequestID,
								BillingMode: request.Deployment.BillingMode,
							}
							_ = h.pipeline.Settle(context.Background(), settlement) //nolint:errcheck // Disconnect-time settlement is reconciled from reservation state.
						}
					}

					return
				}

				if len(chunk.Data) > 0 {
					_, _ = writer.Write(chunk.Data) //nolint:errcheck // A downstream disconnect is handled by terminating the stream.
					_ = writer.Flush()              //nolint:errcheck // A downstream disconnect is handled by terminating the stream.

					if parsed := parseGeminiSSEUsage(chunk.Data); parsed != nil {
						usage = parsed
					}

					if geminiSSEIsError(chunk.Data) {
						failed = true
					}
				}
			case streamErr, ok := <-stream.Errors():
				if ok && streamErr != nil {
					h.pipeline.Release(context.Background(), reservation, "stream_error")
				}

				return
			}
		}
	})

	return nil
}

func parseGeminiNativeUsage(body []byte) *contractsopenai.Usage {
	usage, err := billingusage.NormalizeUsageFromJSON("google-gemini", body)
	if err != nil {
		return nil
	}

	var response struct {
		Candidates []struct {
			GroundingMetadata struct {
				WebSearchQueries []string `json:"webSearchQueries"`
			} `json:"groundingMetadata"`
		} `json:"candidates"`
		Embeddings []json.RawMessage `json:"embeddings"`
	}
	if json.Unmarshal(body, &response) == nil {
		for _, candidate := range response.Candidates {
			if len(candidate.GroundingMetadata.WebSearchQueries) > 0 {
				if usage.ServerToolUse == nil {
					usage.ServerToolUse = map[string]int{}
				}

				usage.ServerToolUse["google_search_queries"] += len(candidate.GroundingMetadata.WebSearchQueries)
			}
		}

		for _, embedding := range response.Embeddings {
			item, itemErr := billingusage.NormalizeUsageFromJSON("google-gemini", embedding)
			if itemErr == nil {
				mergeGeminiUsage(usage, item)
			}
		}
	}

	if usage.TotalTokens == 0 && usage.PromptTokens == 0 && usage.CompletionTokens == 0 && len(usage.ServerToolUse) == 0 {
		return nil
	}

	return usage
}

func mergeGeminiUsage(target, item *contractsopenai.Usage) {
	if target == nil || item == nil {
		return
	}

	target.PromptTokens += item.PromptTokens
	target.CompletionTokens += item.CompletionTokens
	target.TotalTokens += item.TotalTokens

	target.CacheReadInputTokens += item.CacheReadInputTokens
	if target.ServiceTier == "" {
		target.ServiceTier = item.ServiceTier
	}

	if target.PromptTokensDetails == nil {
		target.PromptTokensDetails = &contractsopenai.UsageTokenDetails{}
	}

	if item.PromptTokensDetails != nil {
		target.PromptTokensDetails.CachedTokens += item.PromptTokensDetails.CachedTokens
		target.PromptTokensDetails.AudioTokens += item.PromptTokensDetails.AudioTokens
		target.PromptTokensDetails.ImageTokens += item.PromptTokensDetails.ImageTokens
	}

	if target.CompletionTokensDetails == nil {
		target.CompletionTokensDetails = &contractsopenai.UsageTokenDetails{}
	}

	if item.CompletionTokensDetails != nil {
		target.CompletionTokensDetails.ReasoningTokens += item.CompletionTokensDetails.ReasoningTokens
		target.CompletionTokensDetails.AudioTokens += item.CompletionTokensDetails.AudioTokens
		target.CompletionTokensDetails.ImageTokens += item.CompletionTokensDetails.ImageTokens
	}
}

func parseGeminiSSEUsage(data []byte) *contractsopenai.Usage {
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data:") {
			continue
		}

		if usage := parseGeminiNativeUsage([]byte(strings.TrimSpace(strings.TrimPrefix(line, "data:")))); usage != nil {
			return usage
		}
	}

	return nil
}

func geminiSSEIsError(data []byte) bool {
	text := string(data)
	return strings.Contains(text, `"error":{`) || strings.Contains(text, `"error": {`)
}

func geminiOperationFromRequest(c *fiber.Ctx) (platform.ProviderOperation, platform.ProviderOperationInput, string, bool, error) {
	path := strings.TrimSuffix(c.Path(), "/")
	input := platform.ProviderOperationInput{ResourceID: strings.TrimSpace(c.Params("id")), Limit: c.QueryInt("pageSize", 0)}

	method := c.Method()
	if method == fiber.MethodGet && path == "/v1beta/models" {
		return platform.ProviderOperationGeminiModelsList, input, "", false, nil
	}

	if method == fiber.MethodGet && strings.HasPrefix(path, "/v1beta/models/") && !strings.Contains(strings.TrimPrefix(path, "/v1beta/models/"), ":") {
		input.ResourceID = strings.TrimPrefix(path, "/v1beta/")
		return platform.ProviderOperationGeminiModelsGet, input, strings.TrimPrefix(input.ResourceID, "models/"), false, nil
	}

	if strings.HasPrefix(path, "/v1beta/models/") {
		if method != fiber.MethodPost {
			return "", input, "", false, fmt.Errorf("gemini model actions require POST")
		}

		modelAction := strings.TrimPrefix(path, "/v1beta/models/")

		parts := strings.SplitN(modelAction, ":", 2)
		if len(parts) != 2 {
			return "", input, "", false, fmt.Errorf("a fixed model operation is required")
		}

		model := strings.TrimPrefix(parts[0], "models/")
		switch parts[1] {
		case "generateContent":
			return platform.ProviderOperationGeminiGenerateContent, input, model, false, nil
		case "streamGenerateContent":
			return platform.ProviderOperationGeminiStreamGenerateContent, input, model, true, nil
		case "countTokens":
			return platform.ProviderOperationGeminiCountTokens, input, model, false, nil
		case "embedContent":
			return platform.ProviderOperationGeminiEmbedContent, input, model, false, nil
		case "batchEmbedContents":
			return platform.ProviderOperationGeminiBatchEmbedContents, input, model, false, nil
		case "batchGenerateContent":
			return platform.ProviderOperationGeminiBatchesCreate, input, model, false, nil
		default:
			return "", input, model, false, fmt.Errorf("unsupported Gemini model operation")
		}
	}

	operation, resourceModel, ok := geminiResourceOperation(method, path, &input)
	if ok {
		if operation == platform.ProviderOperationGeminiInteractionsCreate {
			var body struct {
				Model  string `json:"model"`
				Stream bool   `json:"stream"`
			}
			if json.Unmarshal(c.Body(), &body) != nil || strings.TrimSpace(body.Model) == "" {
				return "", input, "", false, fmt.Errorf("interactions create requires model")
			}

			resourceModel = strings.TrimPrefix(strings.TrimSpace(body.Model), "models/")

			return operation, input, resourceModel, body.Stream, nil
		}

		return operation, input, resourceModel, false, nil
	}

	return "", input, "", false, fmt.Errorf("unsupported Gemini operation")
}

func geminiResourceOperation(method, path string, input *platform.ProviderOperationInput) (platform.ProviderOperation, string, bool) {
	type resourceSpec struct {
		prefix                                    string
		create, list, get, update, delete, cancel platform.ProviderOperation
	}

	resources := []resourceSpec{
		{"/v1beta/interactions", platform.ProviderOperationGeminiInteractionsCreate, "", platform.ProviderOperationGeminiInteractionsGet, "", platform.ProviderOperationGeminiInteractionsDelete, platform.ProviderOperationGeminiInteractionsCancel},
		{"/v1beta/files", platform.ProviderOperationGeminiFilesCreate, platform.ProviderOperationGeminiFilesList, platform.ProviderOperationGeminiFilesGet, "", platform.ProviderOperationGeminiFilesDelete, ""},
		{"/upload/v1beta/files", platform.ProviderOperationGeminiFilesCreate, "", "", "", "", ""},
		{"/v1beta/cachedContents", platform.ProviderOperationGeminiCachesCreate, platform.ProviderOperationGeminiCachesList, platform.ProviderOperationGeminiCachesGet, platform.ProviderOperationGeminiCachesUpdate, platform.ProviderOperationGeminiCachesDelete, ""},
		{"/v1beta/batches", "", platform.ProviderOperationGeminiBatchesList, platform.ProviderOperationGeminiBatchesGet, "", "", platform.ProviderOperationGeminiBatchesCancel},
	}
	for _, spec := range resources {
		if path == spec.prefix {
			if method == fiber.MethodPost && spec.create != "" {
				return spec.create, "", true
			}

			if method == fiber.MethodGet && spec.list != "" {
				return spec.list, "", true
			}
		}

		if strings.HasPrefix(path, spec.prefix+"/") {
			id := strings.TrimPrefix(path, spec.prefix+"/")
			cancel := strings.HasSuffix(id, ":cancel")
			id = strings.TrimSuffix(id, ":cancel")
			input.ResourceID = id

			if cancel && method == fiber.MethodPost && spec.cancel != "" {
				return spec.cancel, "", true
			}

			if method == fiber.MethodGet && spec.get != "" {
				return spec.get, "", true
			}

			if method == fiber.MethodPatch && spec.update != "" {
				return spec.update, "", true
			}

			if method == fiber.MethodDelete && spec.delete != "" {
				return spec.delete, "", true
			}
		}
	}

	return "", "", false
}

func isAllowlistedGoogleGeminiModel(model string) bool {
	return platform.IsSupportedGoogleGeminiModel(model)
}

func writeGeminiError(c *fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(fiber.Map{"error": fiber.Map{"code": status, "status": code, "message": message}})
}
