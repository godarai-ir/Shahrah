package proxy

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/gofiber/fiber/v2"
)

const (
	openAIBatchResourceKind = "batch"
	openAIFileResourceKind  = "file"
)

func (h *ResponsesHandler) HandleBatchCreate(c *fiber.Ctx) error {
	return h.handleOpenAIJSONResource(c, platform.ProviderOperationOpenAIBatchesCreate, openAIBatchResourceKind, "")
}

func (h *ResponsesHandler) HandleBatchList(c *fiber.Ctx) error {
	return h.handleOpenAIJSONResource(c, platform.ProviderOperationOpenAIBatchesList, openAIBatchResourceKind, "")
}

func (h *ResponsesHandler) HandleBatchGet(c *fiber.Ctx) error {
	return h.handleOpenAIJSONResource(c, platform.ProviderOperationOpenAIBatchesGet, openAIBatchResourceKind, c.Params("batch_id"))
}

func (h *ResponsesHandler) HandleBatchCancel(c *fiber.Ctx) error {
	return h.handleOpenAIJSONResource(c, platform.ProviderOperationOpenAIBatchesCancel, openAIBatchResourceKind, c.Params("batch_id"))
}

func (h *ResponsesHandler) HandleFileCreate(c *fiber.Ctx) error {
	return h.handleOpenAIRawResource(c, platform.ProviderOperationOpenAIFilesCreate, openAIFileResourceKind, "")
}

func (h *ResponsesHandler) HandleFileList(c *fiber.Ctx) error {
	return h.handleOpenAIJSONResource(c, platform.ProviderOperationOpenAIFilesList, openAIFileResourceKind, "")
}

func (h *ResponsesHandler) HandleFileGet(c *fiber.Ctx) error {
	return h.handleOpenAIJSONResource(c, platform.ProviderOperationOpenAIFilesGet, openAIFileResourceKind, c.Params("file_id"))
}

func (h *ResponsesHandler) HandleFileDelete(c *fiber.Ctx) error {
	return h.handleOpenAIJSONResource(c, platform.ProviderOperationOpenAIFilesDelete, openAIFileResourceKind, c.Params("file_id"))
}

func (h *ResponsesHandler) HandleFileContent(c *fiber.Ctx) error {
	return h.handleOpenAIRawResource(c, platform.ProviderOperationOpenAIFilesContent, openAIFileResourceKind, c.Params("file_id"))
}

func (h *ResponsesHandler) handleOpenAIJSONResource(
	c *fiber.Ctx,
	operation platform.ProviderOperation,
	resourceKind string,
	publicID string,
) error {
	_, currentTenant, ok := responsesRequestContext(c)
	if !ok {
		return nil
	}

	input, err := openAIListOperationInput(c)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadRequest, err.Error(), "invalid_request_error")
	}
	input.ResourceID = strings.TrimSpace(publicID)

	accountSelector := strings.TrimSpace(c.Get("x-godarai-provider-account"))
	if input.ResourceID != "" {
		mapping, mappingErr := h.openAIResourceMapping(c, currentTenant.ID, resourceKind, input.ResourceID)
		if mappingErr != nil {
			return mappingErr
		}
		if mapping != nil {
			accountSelector = mapping.ProviderAccountID
			input.ResourceID = mapping.UpstreamID
		}
	}

	body := append([]byte(nil), c.Body()...)
	if operation == platform.ProviderOperationOpenAIBatchesCreate {
		var request struct {
			InputFileID string `json:"input_file_id"`
		}
		if len(body) == 0 || json.Unmarshal(body, &request) != nil || strings.TrimSpace(request.InputFileID) == "" {
			return writeAPIError(c, fiber.StatusBadRequest, "input_file_id is required", "invalid_request_error")
		}
		mapping, mappingErr := h.openAIResourceMapping(c, currentTenant.ID, openAIFileResourceKind, request.InputFileID)
		if mappingErr != nil {
			return mappingErr
		}
		if mapping != nil {
			if accountSelector != "" && !strings.EqualFold(accountSelector, mapping.ProviderAccountID) {
				return writeAPIError(c, fiber.StatusBadRequest, "input_file_id belongs to a different OpenAI provider account", "invalid_request_error")
			}
			accountSelector = mapping.ProviderAccountID
			body, err = rewriteOpenAIJSONStringField(body, "input_file_id", mapping.UpstreamID)
			if err != nil {
				return writeAPIError(c, fiber.StatusBadRequest, "input_file_id is invalid", "invalid_request_error")
			}
		}
	}

	deployment, err := h.resolveResponseDeployment(c.Context(), currentTenant.ID, accountSelector, "")
	if err != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, err.Error(), "routing_error")
	}
	requestID := responseRequestID(c)
	tunnelReq := platform.TunnelRequest{
		RequestID: requestID, TenantID: currentTenant.ID, TraceID: requestID,
		Operation: operation, OperationInput: input, Deployment: deployment,
		Headers: responseProviderHeaders(c), Body: body,
		RequestMetadata: enrichRequestMetadataForDeployment(map[string]interface{}{
			"provider_endpoint_path": c.Path(),
			"resource_kind":          resourceKind,
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

	if operation == platform.ProviderOperationOpenAIBatchesCreate {
		h.recordOpenAIResourceMapping(c, currentTenant.ID, deployment, resourceKind, resp.Body)
	}
	if resourceKind == openAIBatchResourceKind {
		h.recordOpenAIBatchFileMappings(c, currentTenant.ID, deployment, resp.Body)
	}
	if operation == platform.ProviderOperationOpenAIBatchesCancel {
		h.updateOpenAIResourceMappingFromBody(c, currentTenant.ID, resourceKind, publicID, resp.Body)
	}
	if operation == platform.ProviderOperationOpenAIFilesDelete {
		h.updateOpenAIResourceStatus(c, currentTenant.ID, resourceKind, publicID, "deleted")
	}

	forwardProviderHeaders(c, resp.Headers)
	setTraceHeaders(c, requestID)
	if contentType := responseContentType(resp.Headers); contentType != "" {
		c.Set(fiber.HeaderContentType, contentType)
	} else {
		c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	}
	return c.Status(resp.StatusCode).Send(resp.Body)
}

func (h *ResponsesHandler) handleOpenAIRawResource(
	c *fiber.Ctx,
	operation platform.ProviderOperation,
	resourceKind string,
	publicID string,
) error {
	_, currentTenant, ok := responsesRequestContext(c)
	if !ok {
		return nil
	}
	rawTunnel, ok := h.tunnel.(platform.TunnelRawTransport)
	if !ok {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Raw tunnel streaming is unavailable", "tunnel_unavailable")
	}

	input := platform.ProviderOperationInput{ResourceID: strings.TrimSpace(publicID)}
	accountSelector := strings.TrimSpace(c.Get("x-godarai-provider-account"))
	if input.ResourceID != "" {
		mapping, mappingErr := h.openAIResourceMapping(c, currentTenant.ID, resourceKind, input.ResourceID)
		if mappingErr != nil {
			return mappingErr
		}
		if mapping != nil {
			accountSelector = mapping.ProviderAccountID
			input.ResourceID = mapping.UpstreamID
		}
	}
	deployment, err := h.resolveResponseDeployment(c.Context(), currentTenant.ID, accountSelector, "")
	if err != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, err.Error(), "routing_error")
	}
	requestID := responseRequestID(c)
	headers := responseProviderHeaders(c)
	if contentType := strings.TrimSpace(c.Get(fiber.HeaderContentType)); contentType != "" {
		headers["Content-Type"] = contentType
	}
	tunnelReq := platform.TunnelRequest{
		RequestID: requestID, TenantID: currentTenant.ID, TraceID: requestID,
		Operation: operation, OperationInput: input, Deployment: deployment,
		Headers: headers,
		RequestMetadata: enrichRequestMetadataForDeployment(map[string]interface{}{
			"provider_endpoint_path": c.Path(),
			"resource_kind":          resourceKind,
		}, deployment),
	}
	var bodyReader io.Reader = bytes.NewReader(nil)
	contentLength := int64(0)
	if operation == platform.ProviderOperationOpenAIFilesCreate {
		bodyReader = c.Context().RequestBodyStream()
		contentLength = int64(c.Request().Header.ContentLength())
		if bodyReader == nil {
			body := append([]byte(nil), c.Body()...)
			bodyReader = bytes.NewReader(body)
			contentLength = int64(len(body))
		}
	}
	resp, err := rawTunnel.ProxyRaw(c.Context(), tunnelReq, bodyReader, contentLength)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, fmt.Sprintf("Tunnel error: %v", err), "api_error")
	}
	if resp == nil || resp.Body == nil {
		return writeAPIError(c, fiber.StatusBadGateway, "Provider returned an empty response", "api_error")
	}
	if resp.StatusCode >= httpErrorFloor {
		defer resp.Body.Close()
		errorBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
		return writeOpenAIProviderError(c, resp.StatusCode, errorBody, resp.Headers)
	}

	forwardProviderHeaders(c, resp.Headers)
	setTraceHeaders(c, requestID)
	if contentType := responseContentType(resp.Headers); contentType != "" {
		c.Set(fiber.HeaderContentType, contentType)
	}
	if operation == platform.ProviderOperationOpenAIFilesCreate {
		defer resp.Body.Close()
		result, readErr := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
		if readErr != nil {
			return writeAPIError(c, fiber.StatusBadGateway, "Invalid OpenAI Files response", "api_error")
		}
		h.recordOpenAIResourceMapping(c, currentTenant.ID, deployment, resourceKind, result)
		return c.Status(resp.StatusCode).Send(result)
	}

	c.Status(resp.StatusCode)
	c.Context().SetBodyStreamWriter(func(writer *bufio.Writer) {
		defer resp.Body.Close()
		_, _ = io.CopyBuffer(writer, resp.Body, make([]byte, 32*1024))
		_ = writer.Flush()
	})
	return nil
}

func openAIListOperationInput(c *fiber.Ctx) (platform.ProviderOperationInput, error) {
	input := platform.ProviderOperationInput{
		AfterID:  strings.TrimSpace(c.Query("after")),
		BeforeID: strings.TrimSpace(c.Query("before")),
		Order:    strings.ToLower(strings.TrimSpace(c.Query("order"))),
		Purpose:  strings.TrimSpace(c.Query("purpose")),
	}
	if input.Order != "" && input.Order != "asc" && input.Order != "desc" {
		return input, fmt.Errorf("order must be asc or desc")
	}
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 || value > 100 {
			return input, fmt.Errorf("limit must be between 1 and 100")
		}
		input.Limit = value
	}
	return input, nil
}

func (h *ResponsesHandler) openAIResourceMapping(
	c *fiber.Ctx,
	tenantID string,
	resourceKind string,
	publicID string,
) (*platform.ProviderNativeResourceMapping, error) {
	if h.resources == nil {
		return nil, nil
	}
	mapping, err := h.resources.GetProviderNativeResourceMapping(c.Context(), tenantID, "openai", resourceKind, publicID)
	if err == nil {
		return mapping, nil
	}
	if errors.Is(err, storage.ErrProviderNativeResourceNotFound) {
		return nil, writeAPIError(c, fiber.StatusNotFound, "OpenAI resource not found", "invalid_request_error")
	}
	return nil, writeAPIError(c, fiber.StatusServiceUnavailable, "OpenAI resource routing state is unavailable", "routing_error")
}

func (h *ResponsesHandler) recordOpenAIResourceMapping(
	c *fiber.Ctx,
	tenantID string,
	deployment platform.Deployment,
	resourceKind string,
	body []byte,
) {
	if h.resources == nil || strings.TrimSpace(deployment.ProviderAccountID) == "" {
		return
	}
	var result struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if json.Unmarshal(body, &result) != nil || strings.TrimSpace(result.ID) == "" {
		return
	}
	if strings.TrimSpace(result.Status) == "" {
		result.Status = "active"
	}
	_, _ = h.resources.UpsertProviderNativeResourceMapping(c.Context(), platform.ProviderNativeResourceMapping{
		PublicID: result.ID, TenantID: tenantID, Provider: "openai",
		ProviderAccountID: deployment.ProviderAccountID, ResourceKind: resourceKind,
		UpstreamID: result.ID, Status: result.Status,
		Metadata: map[string]interface{}{"deployment_id": deployment.ID},
	})
}

func (h *ResponsesHandler) recordOpenAIBatchFileMappings(
	c *fiber.Ctx,
	tenantID string,
	deployment platform.Deployment,
	body []byte,
) {
	if h.resources == nil || strings.TrimSpace(deployment.ProviderAccountID) == "" {
		return
	}
	var result struct {
		InputFileID  string `json:"input_file_id"`
		OutputFileID string `json:"output_file_id"`
		ErrorFileID  string `json:"error_file_id"`
	}
	if json.Unmarshal(body, &result) != nil {
		return
	}
	for _, fileID := range []string{result.InputFileID, result.OutputFileID, result.ErrorFileID} {
		fileID = strings.TrimSpace(fileID)
		if fileID == "" {
			continue
		}
		_, _ = h.resources.UpsertProviderNativeResourceMapping(c.Context(), platform.ProviderNativeResourceMapping{
			PublicID: fileID, TenantID: tenantID, Provider: "openai",
			ProviderAccountID: deployment.ProviderAccountID, ResourceKind: openAIFileResourceKind,
			UpstreamID: fileID, Status: "active",
			Metadata: map[string]interface{}{"deployment_id": deployment.ID, "source": "batch"},
		})
	}
}

func (h *ResponsesHandler) updateOpenAIResourceMappingFromBody(
	c *fiber.Ctx,
	tenantID string,
	resourceKind string,
	publicID string,
	body []byte,
) {
	var result struct {
		Status string `json:"status"`
	}
	if json.Unmarshal(body, &result) == nil && strings.TrimSpace(result.Status) != "" {
		h.updateOpenAIResourceStatus(c, tenantID, resourceKind, publicID, result.Status)
	}
}

func (h *ResponsesHandler) updateOpenAIResourceStatus(
	c *fiber.Ctx,
	tenantID string,
	resourceKind string,
	publicID string,
	status string,
) {
	if h.resources != nil {
		_ = h.resources.UpdateProviderNativeResourceStatus(c.Context(), tenantID, "openai", resourceKind, publicID, status)
	}
}
