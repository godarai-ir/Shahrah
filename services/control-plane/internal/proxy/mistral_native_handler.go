package proxy

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type mistralNativeResourceStore interface {
	providerNativeResourceStore
	GetProviderNativeResourceMappingByUpstream(
		ctx context.Context,
		tenantID string,
		provider string,
		providerAccountID string,
		resourceKind string,
		upstreamID string,
	) (*platform.ProviderNativeResourceMapping, error)
	ListProviderNativeResourceMappings(
		ctx context.Context,
		tenantID string,
		provider string,
		providerAccountID string,
		resourceKind string,
		statuses []string,
		limit int,
	) ([]platform.ProviderNativeResourceMapping, error)
}

// MistralNativeHandler exposes fixed Mistral-native resource APIs while
// preserving tenant ownership, provider-account affinity, and billing.
type MistralNativeHandler struct {
	router        *router.Router
	deployments   responseDeploymentLister
	resources     mistralNativeResourceStore
	tunnel        platform.TunnelTransport
	billing       budgeter
	wallets       walletStore
	logger        *zap.Logger
	filesEnabled  bool
	batchEnabled  bool
	agentsEnabled bool
	convsEnabled  bool
}

func NewMistralNativeHandler(deps BaseHandlerDependencies, deployments responseDeploymentLister) *MistralNativeHandler {
	resources, _ := deps.ProviderResources.(mistralNativeResourceStore) //nolint:errcheck // Capability detection intentionally fails closed in Handle.

	return &MistralNativeHandler{
		router: deps.Router, deployments: deployments, resources: resources,
		tunnel: deps.Tunnel, billing: deps.Billing, wallets: deps.Wallets,
		logger:        deps.Logger,
		filesEnabled:  mistralNativeFeatureEnabled("FILES"),
		batchEnabled:  mistralNativeFeatureEnabled("BATCH"),
		agentsEnabled: mistralNativeFeatureEnabled("AGENTS"),
		convsEnabled:  mistralNativeFeatureEnabled("CONVERSATIONS"),
	}
}

func mistralNativeFeatureEnabled(surface string) bool {
	return mistralBooleanEnvironment("SHAHRAH_MISTRAL_ENABLED", true) &&
		mistralBooleanEnvironment("SHAHRAH_MISTRAL_"+surface+"_ENABLED", false)
}

//nolint:gocyclo,cyclop,funlen // This is an operation dispatcher; helpers own validation and mutations.
func (h *MistralNativeHandler) Handle(c *fiber.Ctx) error {
	authInfo := auth.GetAuthInfo(c)
	currentTenant := tenant.FromContext(c)
	if authInfo == nil || currentTenant == nil {
		return writeAPIError(c, fiber.StatusUnauthorized, "Authentication and tenant resolution are required", "authentication_error")
	}

	authInfo.TenantID = currentTenant.ID

	operation, input, err := mistralOperationFromRequest(c)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadRequest, err.Error(), "invalid_request_error")
	}
	if !h.operationEnabled(operation) {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Mistral native operation is disabled", "provider_disabled")
	}
	if h.resources == nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Mistral native resource mapping is unavailable", "routing_error")
	}

	if operation == platform.ProviderOperationMistralFilesCreate ||
		operation == platform.ProviderOperationMistralFilesContent {
		return h.handleRaw(c, currentTenant.ID, operation, input)
	}

	body := append([]byte(nil), c.Body()...)
	model, stream, err := mistralRequestModel(operation, body)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadRequest, err.Error(), "invalid_request_error")
	}
	if stream && !mistralBooleanEnvironment("SHAHRAH_MISTRAL_CONVERSATIONS_STREAMING_ENABLED", false) {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Mistral conversation streaming settlement is disabled", "billing_unavailable")
	}
	if model == "" {
		model = input.Model
	}

	resourceKind := mistralOperationResourceKind(operation)
	accountSelector := strings.TrimSpace(c.Get("x-godarai-provider-account"))
	publicID := input.ResourceID

	var resource *platform.ProviderNativeResourceMapping
	if publicID != "" {
		resource, err = h.resources.GetProviderNativeResourceMapping(c.Context(), currentTenant.ID, mistralProviderName, resourceKind, publicID)
		if err != nil || resource == nil {
			return writeAPIError(c, fiber.StatusNotFound, "Resource not found", "not_found_error")
		}

		accountSelector = resource.ProviderAccountID
		input.ResourceID = resource.UpstreamID
		if model == "" {
			model, _ = resource.Metadata["model"].(string)
		}
	}

	body, model, accountSelector, err = h.rewriteRequestReferences(c.Context(), currentTenant.ID, accountSelector, model, body)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadRequest, err.Error(), "invalid_request_error")
	}

	deployment, officialModel, err := h.resolveDeployment(c.Context(), currentTenant.ID, model, accountSelector)
	if err != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, err.Error(), "routing_error")
	}
	if model != "" {
		expectedType := platform.ModelTypeChat
		if operation == platform.ProviderOperationMistralBatchCreate {
			expectedType = platform.ModelTypeBatch
		}
		if access := modelAccessForType(authInfo, model, expectedType); !access.Allowed {
			return writeAPIError(c, fiber.StatusForbidden, "Model access denied or unsupported for this Mistral operation", "authorization_error")
		}

		body, err = rewriteMistralModel(body, officialModel)
		if err != nil {
			return writeAPIError(c, fiber.StatusBadRequest, "Invalid Mistral request body", "invalid_request_error")
		}
	}

	var (
		estimated   platform.UsageCostBreakdown
		reservation *platform.BudgetReservation
	)
	if model != "" && (mistralOperationIsInference(operation) || operation == platform.ProviderOperationMistralBatchCreate || operation == platform.ProviderOperationMistralAgentsCreate) {
		estimated, err = h.estimateAndValidate(c.Context(), operation, model, officialModel, body, deployment)
		if err != nil {
			return writeAPIError(c, fiber.StatusBadRequest, err.Error(), "invalid_request_error")
		}
	}
	if operation == platform.ProviderOperationMistralBatchCreate || mistralOperationIsInference(operation) {
		if h.billing == nil {
			return writeAPIError(c, fiber.StatusServiceUnavailable, "Billing is unavailable", "billing_error")
		}

		requestID := responseRequestID(c)
		var handled bool
		reservation, handled, err = reserveBudgetForRequest(c, h.billing, authInfo, estimated, requestID, deployment.BillingMode)
		if err != nil || handled {
			return err
		}
	}

	requestID := responseRequestID(c)
	tunnelRequest := platform.TunnelRequest{
		RequestID: requestID, TenantID: currentTenant.ID, TraceID: requestID,
		Operation: operation, OperationInput: input, Deployment: deployment,
		Headers: map[string]string{"Content-Type": fiber.MIMEApplicationJSON}, Body: body,
		RequestMetadata: enrichRequestMetadataForDeployment(map[string]interface{}{
			"provider_endpoint_path": c.Path(), "resource_kind": resourceKind,
		}, deployment),
	}
	if estimated.ModelRuntimeProfile != nil {
		tunnelRequest.ModelRuntimeProfiles = []platform.ModelRuntimeProfileBinding{*estimated.ModelRuntimeProfile}
	}

	if stream {
		return h.handleStream(c, authInfo, model, tunnelRequest, reservation, estimated)
	}

	response, err := h.tunnel.Proxy(c.Context(), tunnelRequest)
	if err != nil {
		releaseBudgetBestEffort(c.Context(), h.logger, h.billing, reservation, "tunnel_error")

		return writeAPIError(c, fiber.StatusBadGateway, "Mistral tunnel request failed", "api_error")
	}
	if response.StatusCode >= httpErrorFloor {
		releaseBudgetBestEffort(c.Context(), h.logger, h.billing, reservation, "provider_http_error")

		return writeCompatibleProviderError(c, deployment.Provider, response.StatusCode, response.Body, response.Headers)
	}

	response.Body, err = h.persistAndRewriteResponse(c.Context(), currentTenant.ID, operation, deployment, response.Body, resource, reservation, estimated, model)
	if err != nil {
		releaseBudgetBestEffort(c.Context(), h.logger, h.billing, reservation, "resource_mapping_error")

		return writeAPIError(c, fiber.StatusBadGateway, "Failed to persist Mistral resource mapping", "api_error")
	}

	if mistralOperationIsInference(operation) && reservation != nil {
		usage := parseMistralNativeUsage(response.Body)
		finalized, finalizeErr := h.billing.FinalizeFromUsage(c.Context(), billingModelForDeployment(model, deployment), usage, estimated)
		if finalizeErr != nil {
			releaseBudgetBestEffort(c.Context(), h.logger, h.billing, reservation, "usage_finalize_error")

			return writeAPIError(c, fiber.StatusBadGateway, "Failed to finalize Mistral usage", "billing_error")
		}
		if err := commitBudgetOrDebit(c.Context(), h.logger, h.billing, h.wallets, reservation, currentTenant.ID, finalized, requestID, deployment.BillingMode); err != nil {
			return writeBudgetCheckError(c, err)
		}
	}

	if resource != nil && mistralOperationDeletesResource(operation) {
		status := "deleted"
		if operation == platform.ProviderOperationMistralBatchCancel {
			status = "canceled"
			releaseBudgetBestEffort(c.Context(), h.logger, h.billing, resource.Reservation, "batch_canceled")
		}

		_ = h.resources.UpdateProviderNativeResourceStatus(c.Context(), currentTenant.ID, mistralProviderName, resourceKind, publicID, status) //nolint:errcheck // Provider succeeded; reconciliation can repair local status.
	}

	forwardProviderHeaders(c, response.Headers)
	setTraceHeaders(c, requestID)
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	return c.Status(response.StatusCode).Send(response.Body)
}

func (h *MistralNativeHandler) operationEnabled(operation platform.ProviderOperation) bool {
	switch mistralOperationResourceKind(operation) {
	case mistralFileResourceKind:
		return h.filesEnabled
	case mistralBatchResourceKind:
		return h.batchEnabled
	case mistralAgentResourceKind:
		return h.agentsEnabled
	case mistralConversationResourceKind:
		return h.convsEnabled
	default:
		return false
	}
}

func (h *MistralNativeHandler) resolveDeployment(ctx context.Context, tenantID, requestedModel, selector string) (platform.Deployment, string, error) {
	if h.router == nil || h.deployments == nil {
		return platform.Deployment{}, "", fmt.Errorf("Mistral routing is unavailable")
	}

	configured, err := h.deployments.ListDeployments(ctx, tenantID)
	if err != nil {
		return platform.Deployment{}, "", err
	}

	matches := make([]platform.Deployment, 0)
	for _, candidate := range configured {
		if !isMistralProvider(candidate.Provider) {
			continue
		}
		if selector != "" && !strings.EqualFold(selector, candidate.ProviderAccountID) &&
			!strings.EqualFold(selector, candidate.ProviderAccountName) {
			continue
		}
		if requestedModel != "" && !mistralDeploymentMatchesModel(candidate, requestedModel) {
			continue
		}

		matches = append(matches, candidate)
	}
	if len(matches) == 0 {
		return platform.Deployment{}, "", fmt.Errorf("no matching Mistral provider account is available")
	}

	accountIDs := make(map[string]struct{})
	for _, match := range matches {
		accountIDs[match.ProviderAccountID] = struct{}{}
	}
	if len(accountIDs) > 1 {
		return platform.Deployment{}, "", fmt.Errorf("ambiguous Mistral provider account; x-godarai-provider-account is required")
	}

	selected := matches[0]
	routed, err := h.router.Route(ctx, tenantID, selected.ModelName)
	if err != nil || routed == nil {
		return platform.Deployment{}, "", fmt.Errorf("Mistral deployment routing failed")
	}
	if !isMistralProvider(routed.Provider) || !strings.EqualFold(routed.ProviderAccountID, selected.ProviderAccountID) {
		return platform.Deployment{}, "", fmt.Errorf("Mistral provider-account affinity could not be preserved")
	}

	officialModel := strings.TrimSpace(routed.Model)
	if officialModel == "" || officialModel == "*" {
		officialModel = providerModelID(requestedModel)
	}

	return toContractDeployment(routed), officialModel, nil
}

func mistralDeploymentMatchesModel(deployment platform.Deployment, requested string) bool {
	requested = normalizeModelIdentifier(requested)

	return strings.EqualFold(requested, normalizeModelIdentifier(deployment.Model)) ||
		strings.EqualFold(requested, normalizeModelIdentifier(deployment.ModelName)) ||
		strings.EqualFold(providerModelID(requested), providerModelID(deployment.Model))
}

func (h *MistralNativeHandler) estimateAndValidate(ctx context.Context, operation platform.ProviderOperation, requestedModel, officialModel string, body []byte, deployment platform.Deployment) (platform.UsageCostBreakdown, error) {
	if err := validateMistralNativePolicy(operation, body); err != nil {
		return platform.UsageCostBreakdown{}, err
	}

	billingModel := billingModelForDeployment(requestedModel, deployment)
	target := ""
	batchEndpoint := ""
	if operation == platform.ProviderOperationMistralBatchCreate {
		var request struct {
			Endpoint string `json:"endpoint"`
		}
		if err := json.Unmarshal(body, &request); err != nil {
			return platform.UsageCostBreakdown{}, fmt.Errorf("invalid Mistral batch request")
		}

		batchEndpoint = request.Endpoint
		target = mistralBatchRuntimeEndpoint(request.Endpoint)
		if target == "" {
			return platform.UsageCostBreakdown{}, fmt.Errorf("unsupported Mistral batch endpoint %q", request.Endpoint)
		}
	}

	var (
		estimated platform.UsageCostBreakdown
		err       error
	)
	switch target {
	case "ocr":
		estimated, err = priceMeteredUsage(ctx, h.billing, billingModel, "ocr_input_pages", "100", platform.UsageCostBreakdown{Model: billingModel, Endpoint: "/v1/batch/jobs", RequestType: "batch"})
	case "audio_transcriptions":
		estimated, err = priceMeteredUsage(ctx, h.billing, billingModel, "audio_input_seconds", "3600", platform.UsageCostBreakdown{Model: billingModel, Endpoint: "/v1/batch/jobs", RequestType: "batch"})
	default:
		estimated, err = h.billing.EstimateCost(ctx, contractsopenai.ChatCompletionRequest{
			Model:    billingModel,
			Messages: []contractsopenai.ChatMessage{{Role: "user", Content: string(body)}},
		})
	}
	if err != nil {
		return platform.UsageCostBreakdown{}, fmt.Errorf("Mistral pricing is unavailable")
	}

	endpoint := mistralRuntimeEndpoint(operation)
	if err := validateMistralJSONParameters(estimated.ModelRuntimeProfile, endpoint, body); err != nil {
		return platform.UsageCostBreakdown{}, err
	}
	if operation == platform.ProviderOperationMistralBatchCreate {
		if estimated.ModelRuntimeProfile == nil ||
			!estimated.ModelRuntimeProfile.Profile.Batch.Eligible ||
			!runtimeValueAllowed(estimated.ModelRuntimeProfile.Profile.Batch.SupportedEndpoints, target) {
			return platform.UsageCostBreakdown{}, fmt.Errorf("model %s does not support batch endpoint %q", officialModel, batchEndpoint)
		}

		estimated.Batch = true
		estimated.ServiceTier = "batch"
		if estimated.PricingConditions == nil {
			estimated.PricingConditions = make(map[string]string)
		}
		estimated.PricingConditions["api"] = "batches"
		if estimated.Metadata == nil {
			estimated.Metadata = make(map[string]interface{})
		}
		estimated.Metadata["batch_endpoint"] = batchEndpoint
		estimated.Metadata["batch_runtime_endpoint"] = target
	}

	return estimated, nil
}

func mistralRuntimeEndpoint(operation platform.ProviderOperation) string {
	switch mistralOperationResourceKind(operation) {
	case mistralFileResourceKind:
		return "files"
	case mistralBatchResourceKind:
		return "batches"
	case mistralAgentResourceKind:
		return "agents"
	case mistralConversationResourceKind:
		return "conversations"
	default:
		return ""
	}
}

func mistralBatchRuntimeEndpoint(endpoint string) string {
	switch strings.TrimSpace(endpoint) {
	case "/v1/chat/completions":
		return "chat_completions"
	case "/v1/embeddings":
		return "embeddings"
	case "/v1/fim/completions":
		return "fim_completions"
	case "/v1/moderations":
		return "moderations"
	case "/v1/chat/moderations":
		return "chat_moderations"
	case "/v1/ocr":
		return "ocr"
	case "/v1/classifications":
		return "classifications"
	case "/v1/conversations":
		return "conversations"
	case "/v1/audio/transcriptions":
		return "audio_transcriptions"
	default:
		return ""
	}
}

func validateMistralNativePolicy(operation platform.ProviderOperation, body []byte) error {
	if operation != platform.ProviderOperationMistralAgentsCreate &&
		operation != platform.ProviderOperationMistralAgentsUpdate &&
		!mistralOperationIsInference(operation) {
		return nil
	}

	var request struct {
		Tools []struct {
			Type string `json:"type"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(body, &request); err != nil {
		return fmt.Errorf("invalid Mistral request body")
	}
	for _, tool := range request.Tools {
		if !strings.EqualFold(strings.TrimSpace(tool.Type), "function") {
			return fmt.Errorf("billable Mistral built-in tool %q is disabled until trustworthy per-tool usage is available", tool.Type)
		}
	}

	return nil
}

func rewriteMistralModel(body []byte, model string) ([]byte, error) {
	if len(body) == 0 || model == "" {
		return body, nil
	}

	var request map[string]interface{}
	if err := json.Unmarshal(body, &request); err != nil {
		return nil, err
	}
	request["model"] = model

	return json.Marshal(request)
}

func mistralOperationDeletesResource(operation platform.ProviderOperation) bool {
	return operation == platform.ProviderOperationMistralFilesDelete ||
		operation == platform.ProviderOperationMistralBatchCancel ||
		operation == platform.ProviderOperationMistralBatchDelete ||
		operation == platform.ProviderOperationMistralAgentsDelete ||
		operation == platform.ProviderOperationMistralConversationsDelete
}

func newMistralPublicID(kind string) string {
	prefix := "mres_gdr_"
	switch kind {
	case mistralFileResourceKind:
		prefix = "mfile_gdr_"
	case mistralBatchResourceKind:
		prefix = "mbatch_gdr_"
	case mistralAgentResourceKind:
		prefix = "magent_gdr_"
	case mistralConversationResourceKind:
		prefix = "mconv_gdr_"
	}

	return prefix + strings.ReplaceAll(uuid.NewString(), "-", "")
}

func parseMistralNativeUsage(body []byte) *contractsopenai.Usage {
	var response struct {
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}
	if json.Unmarshal(body, &response) != nil {
		return nil
	}

	return &contractsopenai.Usage{
		PromptTokens: response.Usage.PromptTokens, CompletionTokens: response.Usage.CompletionTokens,
		TotalTokens: response.Usage.TotalTokens,
	}
}

func (h *MistralNativeHandler) handleRaw(c *fiber.Ctx, tenantID string, operation platform.ProviderOperation, input platform.ProviderOperationInput) error {
	rawTunnel, ok := h.tunnel.(platform.TunnelRawTransport)
	if !ok {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Raw tunnel streaming is unavailable", "tunnel_unavailable")
	}

	accountSelector := strings.TrimSpace(c.Get("x-godarai-provider-account"))
	publicID := input.ResourceID
	if publicID != "" {
		mapping, err := h.resources.GetProviderNativeResourceMapping(c.Context(), tenantID, mistralProviderName, mistralFileResourceKind, publicID)
		if err != nil || mapping == nil {
			return writeAPIError(c, fiber.StatusNotFound, "Resource not found", "not_found_error")
		}

		accountSelector = mapping.ProviderAccountID
		input.ResourceID = mapping.UpstreamID
	}

	deployment, _, err := h.resolveDeployment(c.Context(), tenantID, "", accountSelector)
	if err != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, err.Error(), "routing_error")
	}

	var reader io.Reader = bytes.NewReader(nil)
	contentLength := int64(0)
	if operation == platform.ProviderOperationMistralFilesCreate {
		if err := validateMistralFileUpload(c); err != nil {
			return writeAPIError(c, fiber.StatusBadRequest, err.Error(), "invalid_request_error")
		}

		body := append([]byte(nil), c.Body()...)
		reader = bytes.NewReader(body)
		contentLength = int64(len(body))
	}

	requestID := responseRequestID(c)
	request := platform.TunnelRequest{
		RequestID: requestID, TenantID: tenantID, TraceID: requestID,
		Operation: operation, OperationInput: input, Deployment: deployment,
		Headers: map[string]string{"Content-Type": c.Get(fiber.HeaderContentType)},
		RequestMetadata: enrichRequestMetadataForDeployment(map[string]interface{}{
			"provider_endpoint_path": c.Path(), "resource_kind": mistralFileResourceKind,
		}, deployment),
	}
	response, err := rawTunnel.ProxyRaw(c.Context(), request, reader, contentLength)
	if err != nil || response == nil || response.Body == nil {
		return writeAPIError(c, fiber.StatusBadGateway, "Mistral raw tunnel request failed", "api_error")
	}
	if response.StatusCode >= httpErrorFloor {
		defer response.Body.Close() //nolint:errcheck // Cleanup cannot alter the completed provider response.

		responseBody, _ := io.ReadAll(io.LimitReader(response.Body, 2*1024*1024)) //nolint:errcheck // Provider error reading is best-effort.

		return writeCompatibleProviderError(c, deployment.Provider, response.StatusCode, responseBody, response.Headers)
	}

	forwardProviderHeaders(c, response.Headers)
	setTraceHeaders(c, requestID)
	if contentType := responseContentType(response.Headers); contentType != "" {
		c.Set(fiber.HeaderContentType, contentType)
	}

	if operation == platform.ProviderOperationMistralFilesCreate {
		defer response.Body.Close() //nolint:errcheck // Cleanup cannot alter the completed provider response.

		body, readErr := io.ReadAll(io.LimitReader(response.Body, 4*1024*1024))
		if readErr != nil {
			return writeAPIError(c, fiber.StatusBadGateway, "Invalid Mistral Files response", "api_error")
		}

		body, err = h.persistAndRewriteResponse(c.Context(), tenantID, operation, deployment, body, nil, nil, platform.UsageCostBreakdown{}, "")
		if err != nil {
			return writeAPIError(c, fiber.StatusBadGateway, "Failed to persist Mistral file mapping", "api_error")
		}

		return c.Status(response.StatusCode).Send(body)
	}

	c.Status(response.StatusCode)
	c.Context().SetBodyStreamWriter(func(writer *bufio.Writer) {
		defer response.Body.Close() //nolint:errcheck // Cleanup is best-effort after stream completion.

		_, _ = io.CopyBuffer(writer, response.Body, make([]byte, 32*1024)) //nolint:errcheck // A downstream disconnect ends the stream.
		_ = writer.Flush()                                                 //nolint:errcheck // A downstream disconnect ends the stream.
	})

	return nil
}

func validateMistralFileUpload(c *fiber.Ctx) error {
	if !strings.HasPrefix(strings.ToLower(c.Get(fiber.HeaderContentType)), "multipart/form-data") {
		return fmt.Errorf("multipart/form-data content type is required")
	}
	form, err := c.MultipartForm()
	if err != nil || form == nil || len(form.File["file"]) != 1 {
		return fmt.Errorf("exactly one file is required")
	}
	if form.File["file"][0].Size > 512*1024*1024 {
		return fmt.Errorf("Mistral file uploads cannot exceed 512 MiB")
	}
	purpose := ""
	if values := form.Value["purpose"]; len(values) > 0 {
		purpose = strings.ToLower(strings.TrimSpace(values[0]))
	}
	if purpose != "batch" && purpose != "ocr" {
		return fmt.Errorf("purpose must be batch or ocr")
	}

	return nil
}

func (h *MistralNativeHandler) handleStream(c *fiber.Ctx, authInfo *platform.AuthContext, model string, request platform.TunnelRequest, reservation *platform.BudgetReservation, estimated platform.UsageCostBreakdown) error {
	stream, err := h.tunnel.Stream(c.Context(), request)
	if err != nil {
		releaseBudgetBestEffort(c.Context(), h.logger, h.billing, reservation, "stream_open_failed")

		return writeAPIError(c, fiber.StatusBadGateway, "Mistral tunnel stream failed", "api_error")
	}
	if stream.StatusCode() >= httpErrorFloor {
		releaseBudgetBestEffort(c.Context(), h.logger, h.billing, reservation, "provider_http_error")

		return writeTunnelStreamHTTPError(c, h.logger, stream, nil)
	}

	forwardProviderHeaders(c, stream.Headers())
	c.Set(fiber.HeaderContentType, "text/event-stream")
	c.Context().SetBodyStreamWriter(func(writer *bufio.Writer) {
		defer closeStreamBestEffort(h.logger, stream)
		defer writer.Flush() //nolint:errcheck // A downstream disconnect terminates the stream.

		var usage *contractsopenai.Usage
		failed := false
		for {
			select {
			case chunk, ok := <-stream.Chunks():
				if !ok {
					if failed || usage == nil {
						releaseBudgetBestEffort(context.Background(), h.logger, h.billing, reservation, "stream_incomplete")

						return
					}

					finalized, finalizeErr := h.billing.FinalizeFromUsage(context.Background(), billingModelForDeployment(model, request.Deployment), usage, estimated)
					if finalizeErr != nil {
						releaseBudgetBestEffort(context.Background(), h.logger, h.billing, reservation, "usage_finalize_error")

						return
					}

					_ = commitBudgetOrDebit(context.Background(), h.logger, h.billing, h.wallets, reservation, authInfo.TenantID, finalized, request.RequestID, request.Deployment.BillingMode) //nolint:errcheck // Disconnect-time settlement is reconciled from reservation state.

					return
				}

				data, parsedUsage, rewriteErr := h.rewriteMistralSSEChunk(context.Background(), authInfo.TenantID, request, chunk.Data, model)
				if rewriteErr != nil {
					failed = true
					releaseBudgetBestEffort(context.Background(), h.logger, h.billing, reservation, "stream_resource_mapping_error")

					return
				}
				if parsedUsage != nil {
					usage = parsedUsage
				}
				if mistralSSEIsError(data) {
					failed = true
				}

				_, _ = writer.Write(data) //nolint:errcheck // A downstream disconnect terminates the stream.
				_ = writer.Flush()        //nolint:errcheck // A downstream disconnect terminates the stream.
			case streamErr, ok := <-stream.Errors():
				if ok && streamErr != nil {
					releaseBudgetBestEffort(context.Background(), h.logger, h.billing, reservation, "stream_error")
				}

				return
			}
		}
	})

	return nil
}
