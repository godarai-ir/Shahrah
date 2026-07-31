package proxy

import (
	"encoding/json"
	"fmt"
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

func (h *ResponsesHandler) HandleXAIBatchRequestsList(c *fiber.Ctx) error {
	return h.handleXAIBatchResource(c, platform.ProviderOperationXAIBatchRequestsList, c.Params("batch_id"))
}

func (h *ResponsesHandler) HandleXAIBatchRequestsAdd(c *fiber.Ctx) error {
	return h.handleXAIBatchResource(c, platform.ProviderOperationXAIBatchRequestsAdd, c.Params("batch_id"))
}

func (h *ResponsesHandler) HandleXAIBatchResults(c *fiber.Ctx) error {
	return h.handleXAIBatchResource(c, platform.ProviderOperationXAIBatchResults, c.Params("batch_id"))
}

func (h *ResponsesHandler) handleXAIBatchResource(c *fiber.Ctx, operation platform.ProviderOperation, publicID string) error {
	_, currentTenant, ok := responsesRequestContext(c)
	if !ok {
		return nil
	}
	// Native request/result routing is implemented, but creating new billable
	// work remains fail-closed until paginated results are coupled to an
	// exactly-once reservation and settlement state machine. Do not expose a
	// provider job that GodarAI cannot yet reconcile safely.
	if operation == platform.ProviderOperationXAIBatchesCreate {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "xAI Batch billing settlement is not enabled", "billing_unavailable")
	}
	if h.resources == nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "xAI Batch routing state is unavailable", "routing_error")
	}

	input, err := openAIListOperationInput(c)
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
	if operation == platform.ProviderOperationXAIBatchRequestsAdd {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "xAI Batch billing settlement is not enabled", "billing_unavailable")
	}

	body := append([]byte(nil), c.Body()...)
	if operation == platform.ProviderOperationXAIBatchesCreate {
		var request struct {
			Name string `json:"name"`
		}
		if json.Unmarshal(body, &request) != nil || strings.TrimSpace(request.Name) == "" {
			return writeAPIError(c, fiber.StatusBadRequest, "name is required", "invalid_request_error")
		}
	}

	deployment, err := h.resolveResponseDeployment(c.Context(), currentTenant.ID, "xai", accountSelector, "")
	if err != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, err.Error(), "routing_error")
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
	if operation == platform.ProviderOperationXAIBatchRequestsAdd && mapping != nil {
		updated := *mapping
		if updated.Metadata == nil {
			updated.Metadata = map[string]interface{}{}
		}
		updated.Metadata["models"] = models
		updated.Status = "processing"
		_, _ = h.resources.UpsertProviderNativeResourceMapping(c.Context(), updated)
	}
	if operation == platform.ProviderOperationXAIBatchesCancel {
		h.updateProviderResourceStatus(c, currentTenant.ID, "xai", openAIBatchResourceKind, publicID, "cancelled")
	}

	forwardProviderHeaders(c, resp.Headers)
	setTraceHeaders(c, requestID)
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	return c.Status(resp.StatusCode).Send(resp.Body)
}

func (h *ResponsesHandler) resolveXAIBatchRuntimeProfiles(c *fiber.Ctx, tenantID, accountID string, body []byte) ([]platform.ModelRuntimeProfileBinding, []string, error) {
	requests, err := parseXAIBatchRequests(body)
	if err != nil {
		return nil, nil, err
	}
	profiles := make([]platform.ModelRuntimeProfileBinding, 0, len(requests))
	models := make([]string, 0, len(requests))
	seen := map[string]struct{}{}
	for _, request := range requests {
		if _, ok := seen[request.Model]; ok {
			continue
		}
		seen[request.Model] = struct{}{}
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
		binding := *estimated.ModelRuntimeProfile
		if !binding.Profile.Batch.Eligible || !runtimeValueAllowed(binding.Profile.Batch.SupportedEndpoints, request.Endpoint) {
			return nil, nil, fmt.Errorf("model %s does not support xAI Batch %s", request.Model, request.Endpoint)
		}
		if maximum := binding.Profile.Batch.MaxRequestBytes; maximum > 0 && len(body) > maximum {
			return nil, nil, fmt.Errorf("batch request body exceeds model limit of %d bytes", maximum)
		}
		profiles = append(profiles, binding)
		models = append(models, request.Model)
	}
	return profiles, models, nil
}

type xaiBatchModelRequest struct {
	Model    string
	Endpoint string
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
			result = append(result, xaiBatchModelRequest{Model: strings.TrimSpace(modelRequest.Model), Endpoint: endpoint})
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
		BatchID string `json:"batch_id"`
	}
	if json.Unmarshal(body, &result) != nil || strings.TrimSpace(result.BatchID) == "" {
		return
	}
	_, _ = h.resources.UpsertProviderNativeResourceMapping(c.Context(), platform.ProviderNativeResourceMapping{
		PublicID: result.BatchID, TenantID: tenantID, Provider: "xai",
		ProviderAccountID: deployment.ProviderAccountID, ResourceKind: openAIBatchResourceKind,
		UpstreamID: result.BatchID, Status: "created",
		Metadata: map[string]interface{}{"deployment_id": deployment.ID},
	})
}
