package proxy

import (
	"fmt"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/gofiber/fiber/v2"
)

// HandleXAIModelsList relays the xAI all-models catalog.
func (h *ResponsesHandler) HandleXAIModelsList(c *fiber.Ctx) error {
	return h.handleXAIModelResource(c, platform.ProviderOperationOpenAIModelsList, "")
}

// HandleXAIModelsGet relays one xAI model record.
func (h *ResponsesHandler) HandleXAIModelsGet(c *fiber.Ctx) error {
	modelID := strings.TrimSpace(c.Params("model_id"))
	if modelID == "" {
		modelID = c.Params("id")
	}

	return h.handleXAIModelResource(c, platform.ProviderOperationOpenAIModelsGet, modelID)
}

// HandleXAILanguageModelsList relays the xAI language-model catalog.
func (h *ResponsesHandler) HandleXAILanguageModelsList(c *fiber.Ctx) error {
	return h.handleXAIModelResource(c, platform.ProviderOperationXAILanguageModelsList, "")
}

// HandleXAILanguageModelsGet relays one xAI language-model record.
func (h *ResponsesHandler) HandleXAILanguageModelsGet(c *fiber.Ctx) error {
	return h.handleXAIModelResource(c, platform.ProviderOperationXAILanguageModelsGet, c.Params("model_id"))
}

// HandleXAIImageGenerationModelsList relays the xAI image-model catalog.
func (h *ResponsesHandler) HandleXAIImageGenerationModelsList(c *fiber.Ctx) error {
	return h.handleXAIModelResource(c, platform.ProviderOperationXAIImageGenerationModelsList, "")
}

// HandleXAIImageGenerationModelsGet relays one xAI image-model record.
func (h *ResponsesHandler) HandleXAIImageGenerationModelsGet(c *fiber.Ctx) error {
	return h.handleXAIModelResource(c, platform.ProviderOperationXAIImageGenerationModelsGet, c.Params("model_id"))
}

// HandleXAIFileInitialize relays xAI's fixed resumable-upload initializer.
func (h *ResponsesHandler) HandleXAIFileInitialize(c *fiber.Ctx) error {
	return h.handleXAIModelResource(c, platform.ProviderOperationXAIFilesInitialize, "")
}

// HandleXAIFileUploadChunks relays xAI's fixed chunk-upload operation.
func (h *ResponsesHandler) HandleXAIFileUploadChunks(c *fiber.Ctx) error {
	return h.handleXAIModelResource(c, platform.ProviderOperationXAIFilesUploadChunks, "")
}

func (h *ResponsesHandler) handleXAIModelResource(
	c *fiber.Ctx,
	operation platform.ProviderOperation,
	modelID string,
) error {
	_, currentTenant, ok := responsesRequestContext(c)
	if !ok {
		return nil
	}

	deployment, err := h.resolveResponseDeployment(
		c.Context(),
		currentTenant.ID,
		"xai",
		strings.TrimSpace(c.Get("x-godarai-provider-account")),
		"",
	)
	if err != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, err.Error(), "routing_error")
	}

	requestID := responseRequestID(c)
	request := platform.TunnelRequest{
		RequestID: requestID,
		TenantID:  currentTenant.ID,
		TraceID:   requestID,
		Operation: operation,
		OperationInput: platform.ProviderOperationInput{
			ResourceID: strings.TrimSpace(modelID),
		},
		Deployment: deployment,
		Headers:    responseProviderHeaders(c),
		Body:       append([]byte(nil), c.Body()...),
		RequestMetadata: enrichRequestMetadataForDeployment(map[string]interface{}{
			"provider_endpoint_path": c.Path(),
		}, deployment),
	}

	if healthyErr := h.pipeline.EnsureReady(c.Context()); healthyErr != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Tunnel path unavailable", "tunnel_unavailable")
	}

	response, err := h.pipeline.Dispatch(c.Context(), request)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, fmt.Sprintf("Tunnel error: %v", err), "api_error")
	}

	if response.StatusCode >= httpErrorFloor {
		return writeOpenAIProviderError(c, response.StatusCode, response.Body, response.Headers)
	}

	forwardProviderHeaders(c, response.Headers)
	setTraceHeaders(c, requestID)
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	return c.Status(response.StatusCode).Send(response.Body)
}
