package proxy

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	"github.com/gofiber/fiber/v2"
)

func (h *ResponsesHandler) HandleXAIBatchCreate(c *fiber.Ctx) error {
	return h.handleXAIBatchResource(c, platform.ProviderOperationXAIBatchesCreate, "")
}

func (h *ResponsesHandler) HandleXAIBatchList(c *fiber.Ctx) error {
	return h.handleXAIBatchResource(c, platform.ProviderOperationXAIBatchesList, "")
}

func (h *ResponsesHandler) HandleXAIBatchGet(c *fiber.Ctx) error {
	return h.handleXAIBatchResource(c, platform.ProviderOperationXAIBatchesGet, c.Params("batch_id"))
}

func (h *ResponsesHandler) HandleXAIBatchCancel(c *fiber.Ctx) error {
	return h.handleXAIBatchResource(c, platform.ProviderOperationXAIBatchesCancel, c.Params("batch_id"))
}

// HandleXAIBatchCancelColon accepts xAI's canonical /batches/{id}:cancel path.
func (h *ResponsesHandler) HandleXAIBatchCancelColon(c *fiber.Ctx) error {
	value := strings.TrimSpace(c.Params("*"))
	if !strings.HasSuffix(value, ":cancel") {
		return fiber.ErrNotFound
	}

	batchID := strings.TrimSpace(strings.TrimSuffix(value, ":cancel"))
	if batchID == "" || strings.Contains(batchID, "/") {
		return fiber.ErrNotFound
	}

	return h.handleXAIBatchResource(c, platform.ProviderOperationXAIBatchesCancel, batchID)
}

func (h *ResponsesHandler) HandleXAIBatchRequestsList(c *fiber.Ctx) error {
	return h.handleXAIBatchResource(c, platform.ProviderOperationXAIBatchRequestsList, c.Params("batch_id"))
}

func (h *ResponsesHandler) HandleXAIBatchRequestsAdd(c *fiber.Ctx) error {
	return h.handleXAIBatchResource(c, platform.ProviderOperationXAIBatchRequestsAdd, c.Params("batch_id"))
}

func (h *ResponsesHandler) HandleXAIBatchResults(c *fiber.Ctx) error {
	return h.handleXAIBatchResource(c, platform.ProviderOperationXAIBatchResults, c.Params("batch_id"))
}

//nolint:govet // Short-lived branch errors are checked immediately and intentionally scoped with short declarations.
func (h *ResponsesHandler) handleXAIBatchResource(c *fiber.Ctx, operation platform.ProviderOperation, publicID string) error {
	_, currentTenant, ok := responsesRequestContext(c)
	if !ok {
		return nil
	}
	if h.resources == nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "xAI Batch routing state is unavailable", "routing_error")
	}

	input, err := xaiBatchListOperationInput(c, operation)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadRequest, err.Error(), "invalid_request_error")
	}

	input.ResourceID = strings.TrimSpace(publicID)
	accountSelector := strings.TrimSpace(c.Get("x-godarai-provider-account"))

	var mapping *platform.ProviderNativeResourceMapping
	if input.ResourceID != "" {
		mapping, err = h.providerResourceMapping(c, currentTenant.ID, "xai", openAIBatchResourceKind, input.ResourceID)
		if err != nil {
			return err
		}

		accountSelector = mapping.ProviderAccountID
		input.ResourceID = mapping.UpstreamID
	}

	body := append([]byte(nil), c.Body()...)
	inputFileID := ""

	if operation == platform.ProviderOperationXAIBatchesCreate {
		var request struct {
			Name        string `json:"name"`
			InputFileID string `json:"input_file_id"`
		}
		if json.Unmarshal(body, &request) != nil || strings.TrimSpace(request.Name) == "" {
			return writeAPIError(c, fiber.StatusBadRequest, "name is required", "invalid_request_error")
		}

		inputFileID = strings.TrimSpace(request.InputFileID)
		if inputFileID != "" {
			fileMapping, mappingErr := h.providerResourceMapping(c, currentTenant.ID, "xai", openAIFileResourceKind, inputFileID)
			if mappingErr != nil {
				return mappingErr
			}

			if accountSelector != "" && !strings.EqualFold(accountSelector, fileMapping.ProviderAccountID) {
				return writeAPIError(c, fiber.StatusBadRequest, "input_file_id belongs to a different xAI provider account", "invalid_request_error")
			}

			accountSelector = fileMapping.ProviderAccountID

			body, err = rewriteOpenAIJSONStringField(body, "input_file_id", fileMapping.UpstreamID)
			if err != nil {
				return writeAPIError(c, fiber.StatusBadRequest, "input_file_id is invalid", "invalid_request_error")
			}
		}
	}

	deployment, err := h.resolveResponseDeployment(c.Context(), currentTenant.ID, "xai", accountSelector, "")
	if err != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, err.Error(), "routing_error")
	}

	if (operation == platform.ProviderOperationXAIBatchRequestsAdd || inputFileID != "") &&
		deploymentBillingMode(deployment) != platform.BillingModeExternalBilled {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "xAI Batch wallet billing settlement is not enabled; use an external-billed provider account", "billing_unavailable")
	}

	profiles := []platform.ModelRuntimeProfileBinding(nil)

	models := []string(nil)
	if operation == platform.ProviderOperationXAIBatchRequestsAdd {
		profiles, models, err = h.resolveXAIBatchRuntimeProfiles(c, currentTenant.ID, deployment.ProviderAccountID, body)
		if err != nil {
			return writeAPIError(c, fiber.StatusBadRequest, err.Error(), "invalid_request_error")
		}
	}

	requestID := responseRequestID(c)

	tunnelReq := platform.TunnelRequest{
		RequestID: requestID, TenantID: currentTenant.ID, TraceID: requestID,
		Operation: operation, OperationInput: input, Deployment: deployment,
		ModelRuntimeProfiles: profiles, Headers: responseProviderHeaders(c), Body: body,
		RequestMetadata: enrichRequestMetadataForDeployment(map[string]interface{}{
			"provider_endpoint_path": c.Path(),
			"resource_kind":          openAIBatchResourceKind,
			"models":                 models,
		}, deployment),
	}
	if err := h.tunnel.Healthy(c.Context()); err != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Tunnel path unavailable", "tunnel_unavailable")
	}

	resp, err := h.tunnel.Proxy(c.Context(), tunnelReq)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, fmt.Sprintf("Tunnel error: %v", err), "api_error")
	}

	if resp.StatusCode >= httpErrorFloor {
		return writeOpenAIProviderError(c, resp.StatusCode, resp.Body, resp.Headers)
	}

	if operation == platform.ProviderOperationXAIBatchesCreate {
		h.recordXAIBatchMapping(c, currentTenant.ID, deployment, resp.Body)
	}

	if operation == platform.ProviderOperationXAIBatchesList {
		h.recordXAIBatchListMappings(c, currentTenant.ID, deployment, resp.Body)
	}

	if operation == platform.ProviderOperationXAIBatchRequestsAdd && mapping != nil {
		updated := *mapping
		if updated.Metadata == nil {
			updated.Metadata = map[string]interface{}{}
		}

		updated.Metadata["models"] = models
		updated.Status = "processing"
		_, _ = h.resources.UpsertProviderNativeResourceMapping(c.Context(), updated) //nolint:errcheck // Failure is intentionally handled by the surrounding best-effort path.
	}

	if operation == platform.ProviderOperationXAIBatchesCancel {
		h.updateProviderResourceStatus(c, currentTenant.ID, "xai", openAIBatchResourceKind, publicID, "canceled")
	}

	forwardProviderHeaders(c, resp.Headers)
	setTraceHeaders(c, requestID)
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	return c.Status(resp.StatusCode).Send(resp.Body)
}

func xaiBatchListOperationInput(c *fiber.Ctx, operation platform.ProviderOperation) (platform.ProviderOperationInput, error) {
	input := platform.ProviderOperationInput{
		AfterID:         strings.TrimSpace(c.Query("after")),
		BeforeID:        strings.TrimSpace(c.Query("before")),
		Order:           strings.ToLower(strings.TrimSpace(c.Query("order"))),
		PaginationToken: strings.TrimSpace(c.Query("pagination_token")),
	}
	if input.Order != "" && input.Order != "asc" && input.Order != "desc" {
		return input, fmt.Errorf("order must be asc or desc")
	}

	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		maximum := 100
		if operation == platform.ProviderOperationXAIBatchRequestsList || operation == platform.ProviderOperationXAIBatchResults {
			maximum = 1000
		}

		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 || value > maximum {
			return input, fmt.Errorf("limit must be between 1 and %d", maximum)
		}

		input.Limit = value
	}

	return input, nil
}

func (h *ResponsesHandler) resolveXAIBatchRuntimeProfiles(c *fiber.Ctx, tenantID, accountID string, body []byte) ([]platform.ModelRuntimeProfileBinding, []string, error) {
	requests, err := parseXAIBatchRequests(body)
	if err != nil {
		return nil, nil, err
	}

	profiles := make([]platform.ModelRuntimeProfileBinding, 0, len(requests))
	models := make([]string, 0, len(requests))

	bindings := make(map[string]platform.ModelRuntimeProfileBinding, len(requests))
	for _, request := range requests {
		binding, found := bindings[request.Model]
		if !found {
			var deployment *router.Deployment
			for _, candidate := range modelRoutingCandidates(request.Model) {
				deployment, err = h.llmRouter.Route(c.Context(), tenantID, candidate)
				if err == nil && deployment != nil {
					break
				}
			}

			if err != nil || deployment == nil || !strings.EqualFold(deployment.Provider, "xai") || !strings.EqualFold(deployment.ProviderAccountID, accountID) {
				return nil, nil, fmt.Errorf("batch model %s is not available on the batch provider account", request.Model)
			}

			var estimated platform.UsageCostBreakdown
			if strings.HasPrefix(strings.ToLower(providerModelID(request.Model)), "grok-imagine-image") {
				estimated = estimateImageCostFallback(request.Model, map[string]interface{}{"n": 1, "resolution": "1k"})
				estimated.Metadata = map[string]interface{}{"n": 1, "resolution": "1k"}
				estimated, err = h.billing.FinalizeImageUsage(c.Context(), billingModelForDeployment(request.Model, toContractDeployment(deployment)), &contractsopenai.Usage{}, nil, estimated)
			} else {
				estimated, err = h.billing.EstimateCost(c.Context(), contractsopenai.ChatCompletionRequest{Model: billingModelForDeployment(request.Model, toContractDeployment(deployment))})
			}

			if err != nil || estimated.ModelRuntimeProfile == nil {
				return nil, nil, fmt.Errorf("batch runtime profile is unavailable for %s", request.Model)
			}

			binding = *estimated.ModelRuntimeProfile
			bindings[request.Model] = binding
			profiles = append(profiles, binding)
			models = append(models, request.Model)
		}

		if !binding.Profile.Batch.Eligible || !runtimeValueAllowed(binding.Profile.Batch.SupportedEndpoints, request.Endpoint) {
			return nil, nil, fmt.Errorf("model %s does not support xAI Batch %s", request.Model, request.Endpoint)
		}

		if maximum := binding.Profile.Batch.MaxRequestBytes; maximum > 0 && request.PayloadBytes > maximum {
			return nil, nil, fmt.Errorf("batch request for model %s exceeds limit of %d bytes", request.Model, maximum)
		}
	}

	return profiles, models, nil
}

type xaiBatchModelRequest struct {
	Model        string
	Endpoint     string
	PayloadBytes int
}

func parseXAIBatchRequests(body []byte) ([]xaiBatchModelRequest, error) {
	var request struct {
		BatchRequests []struct {
			BatchRequest map[string]json.RawMessage `json:"batch_request"`
		} `json:"batch_requests"`
	}
	if json.Unmarshal(body, &request) != nil || len(request.BatchRequests) == 0 {
		return nil, fmt.Errorf("batch_requests is required")
	}

	result := make([]xaiBatchModelRequest, 0, len(request.BatchRequests))
	for _, item := range request.BatchRequests {
		if len(item.BatchRequest) != 1 {
			return nil, fmt.Errorf("each batch_request must contain exactly one request type")
		}

		for requestType, raw := range item.BatchRequest {
			var modelRequest struct {
				Model string `json:"model"`
			}
			if json.Unmarshal(raw, &modelRequest) != nil || strings.TrimSpace(modelRequest.Model) == "" {
				return nil, fmt.Errorf("each batch request model is required")
			}

			endpoint := xaiBatchEndpointForRequestType(requestType)
			if endpoint == "" {
				return nil, fmt.Errorf("unsupported xAI batch request type %s", requestType)
			}

			result = append(result, xaiBatchModelRequest{
				Model:        strings.TrimSpace(modelRequest.Model),
				Endpoint:     endpoint,
				PayloadBytes: len(raw),
			})
		}
	}

	return result, nil
}

func xaiBatchEndpointForRequestType(requestType string) string {
	switch strings.ToLower(strings.TrimSpace(requestType)) {
	case "chat_get_completion":
		return "chat_completions"
	case "responses":
		return "responses"
	case "image_generation":
		return "image_generations"
	case "image_edit":
		return "image_edits"
	default:
		return ""
	}
}

func (h *ResponsesHandler) recordXAIBatchMapping(c *fiber.Ctx, tenantID string, deployment platform.Deployment, body []byte) {
	var result struct {
		BatchID    string  `json:"batch_id"`
		CancelTime *string `json:"cancel_time"`
		State      struct {
			NumRequests  int `json:"num_requests"`
			NumPending   int `json:"num_pending"`
			NumSuccess   int `json:"num_success"`
			NumError     int `json:"num_error"`
			NumCancelled int `json:"num_cancelled"` //nolint:misspell // xAI's official response field uses British spelling.
		} `json:"state"`
	}
	if json.Unmarshal(body, &result) != nil || strings.TrimSpace(result.BatchID) == "" {
		return
	}

	status := "created"
	if result.CancelTime != nil || result.State.NumCancelled > 0 {
		status = "canceled"
	} else if result.State.NumPending > 0 {
		status = "processing"
	} else if result.State.NumRequests > 0 && result.State.NumSuccess+result.State.NumError >= result.State.NumRequests {
		status = "completed"
	}

	mapping := platform.ProviderNativeResourceMapping{
		PublicID: result.BatchID, TenantID: tenantID, Provider: "xai",
		ProviderAccountID: deployment.ProviderAccountID, ResourceKind: openAIBatchResourceKind,
		UpstreamID: result.BatchID, Status: status,
		Metadata: map[string]interface{}{"deployment_id": deployment.ID},
	}
	if existing, err := h.resources.GetProviderNativeResourceMapping(c.Context(), tenantID, "xai", openAIBatchResourceKind, result.BatchID); err == nil && existing != nil {
		mapping = *existing
		mapping.Status = status
	}

	_, _ = h.resources.UpsertProviderNativeResourceMapping(c.Context(), mapping) //nolint:errcheck // Failure is intentionally handled by the surrounding best-effort path.
}

func (h *ResponsesHandler) recordXAIBatchListMappings(c *fiber.Ctx, tenantID string, deployment platform.Deployment, body []byte) {
	var result struct {
		Batches []json.RawMessage `json:"batches"`
	}
	if json.Unmarshal(body, &result) != nil {
		return
	}

	for _, batch := range result.Batches {
		h.recordXAIBatchMapping(c, tenantID, deployment, batch)
	}
}
