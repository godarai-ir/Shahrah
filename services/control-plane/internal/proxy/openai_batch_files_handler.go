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
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/providernative"
	"github.com/gofiber/fiber/v2"
)

const (
	openAIBatchResourceKind = "batch"
	openAIFileResourceKind  = "file"
	xAIFileUploadMaxBytes   = int64(50_000_000)
)

func (h *ResponsesHandler) HandleBatchCreate(c *fiber.Ctx) error {
	if strings.EqualFold(strings.TrimSpace(c.Get("x-godarai-provider")), "xai") {
		return h.HandleXAIBatchCreate(c)
	}

	return h.handleOpenAIJSONResource(c, platform.ProviderOperationOpenAIBatchesCreate, openAIBatchResourceKind, "")
}

func (h *ResponsesHandler) HandleBatchList(c *fiber.Ctx) error {
	if strings.EqualFold(strings.TrimSpace(c.Get("x-godarai-provider")), "xai") {
		return h.HandleXAIBatchList(c)
	}

	return h.handleOpenAIJSONResource(c, platform.ProviderOperationOpenAIBatchesList, openAIBatchResourceKind, "")
}

func (h *ResponsesHandler) HandleBatchGet(c *fiber.Ctx) error {
	if strings.EqualFold(strings.TrimSpace(c.Get("x-godarai-provider")), "xai") {
		return h.HandleXAIBatchGet(c)
	}

	return h.handleOpenAIJSONResource(c, platform.ProviderOperationOpenAIBatchesGet, openAIBatchResourceKind, c.Params("batch_id"))
}

func (h *ResponsesHandler) HandleBatchCancel(c *fiber.Ctx) error {
	if strings.EqualFold(strings.TrimSpace(c.Get("x-godarai-provider")), "xai") {
		return h.HandleXAIBatchCancel(c)
	}

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

func (h *ResponsesHandler) HandleFileUpdate(c *fiber.Ctx) error {
	return h.handleOpenAIJSONResource(c, platform.ProviderOperationXAIFilesUpdate, openAIFileResourceKind, c.Params("file_id"))
}

func (h *ResponsesHandler) HandleFilePublicURLCreate(c *fiber.Ctx) error {
	return h.handleOpenAIJSONResource(c, platform.ProviderOperationXAIFilePublicURLCreate, openAIFileResourceKind, c.Params("file_id"))
}

func (h *ResponsesHandler) HandleFilePublicURLRevoke(c *fiber.Ctx) error {
	return h.handleOpenAIJSONResource(c, platform.ProviderOperationXAIFilePublicURLRevoke, openAIFileResourceKind, c.Params("file_id"))
}

//nolint:govet // Short-lived branch errors are checked immediately and intentionally scoped with short declarations.
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

	providerSelector, err := resourceProviderSelector(c, resourceKind, input.ResourceID == "")
	if err != nil {
		return writeAPIError(c, fiber.StatusBadRequest, err.Error(), "invalid_request_error")
	}

	accountSelector := strings.TrimSpace(c.Get("x-godarai-provider-account"))

	var resourceMapping *platform.ProviderNativeResourceMapping

	if input.ResourceID != "" {
		mapping, mappingErr := h.providerResourceMapping(c, currentTenant.ID, providerSelector, resourceKind, input.ResourceID)
		if mappingErr != nil {
			return mappingErr
		}

		if mapping != nil {
			resourceMapping = mapping
			providerSelector = mapping.Provider
			accountSelector = mapping.ProviderAccountID
			input.ResourceID = mapping.UpstreamID
		}
	}

	body := append([]byte(nil), c.Body()...)
	if operation == platform.ProviderOperationXAIFilePublicURLCreate {
		if len(bytes.TrimSpace(body)) == 0 {
			body = []byte(`{}`)
		}

		var request struct {
			ExpiresAfter *int `json:"expires_after"`
		}
		if json.Unmarshal(body, &request) != nil {
			return writeAPIError(c, fiber.StatusBadRequest, "Invalid public URL request body", "invalid_request_error")
		}

		if request.ExpiresAfter != nil && (*request.ExpiresAfter < 3600 || *request.ExpiresAfter > 30*24*60*60) {
			return writeAPIError(c, fiber.StatusBadRequest, "expires_after must be between 3600 and 2592000 seconds", "invalid_request_error")
		}

		if message := validateXAIFilePublicURLRequest(resourceMapping, request.ExpiresAfter, time.Now()); message != "" {
			return writeAPIError(c, fiber.StatusBadRequest, message, "invalid_request_error")
		}
	}

	if operation == platform.ProviderOperationOpenAIBatchesCreate {
		var request struct {
			InputFileID string `json:"input_file_id"`
		}
		if len(body) == 0 || json.Unmarshal(body, &request) != nil || strings.TrimSpace(request.InputFileID) == "" {
			return writeAPIError(c, fiber.StatusBadRequest, "input_file_id is required", "invalid_request_error")
		}

		mapping, mappingErr := h.providerResourceMapping(c, currentTenant.ID, "openai", openAIFileResourceKind, request.InputFileID)
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

	if operation == platform.ProviderOperationXAIFilePublicURLCreate || operation == platform.ProviderOperationXAIFilePublicURLRevoke || operation == platform.ProviderOperationXAIFilesUpdate {
		if !strings.EqualFold(providerSelector, "xai") {
			return writeAPIError(c, fiber.StatusBadRequest, "file operation is only supported by xAI", "invalid_request_error")
		}
	}

	deployment, err := h.resolveResponseDeployment(c.Context(), currentTenant.ID, providerSelector, accountSelector, "")
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
	if err := h.pipeline.EnsureReady(c.Context()); err != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Tunnel path unavailable", "tunnel_unavailable")
	}

	resp, err := h.pipeline.Dispatch(c.Context(), tunnelReq)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, fmt.Sprintf("Tunnel error: %v", err), "api_error")
	}

	if resp.StatusCode >= httpErrorFloor {
		return writeOpenAIProviderError(c, resp.StatusCode, resp.Body, resp.Headers)
	}

	if operation == platform.ProviderOperationOpenAIBatchesCreate {
		h.recordOpenAIResourceMapping(c, currentTenant.ID, deployment, resourceKind, resp.Body)
	}

	if resourceKind == openAIFileResourceKind && (operation == platform.ProviderOperationOpenAIFilesGet || operation == platform.ProviderOperationXAIFilesUpdate) {
		if resourceMapping != nil {
			h.recordProviderFileMetadata(c, resourceMapping, resp.Body)
		} else {
			h.recordOpenAIResourceMapping(c, currentTenant.ID, deployment, resourceKind, resp.Body)
		}
	}

	if resourceKind == openAIFileResourceKind && operation == platform.ProviderOperationOpenAIFilesList {
		h.recordProviderFileListMappings(c, currentTenant.ID, deployment, resp.Body)
	}

	if operation == platform.ProviderOperationXAIFilePublicURLCreate || operation == platform.ProviderOperationXAIFilePublicURLRevoke {
		h.recordXAIFilePublicURLState(c, resourceMapping, resp.Body, operation == platform.ProviderOperationXAIFilePublicURLRevoke)
	}

	if resourceKind == openAIBatchResourceKind {
		h.recordOpenAIBatchFileMappings(c, currentTenant.ID, deployment, resp.Body)
	}

	if operation == platform.ProviderOperationOpenAIBatchesCancel {
		h.updateOpenAIResourceMappingFromBody(c, currentTenant.ID, resourceKind, publicID, resp.Body)
	}

	if operation == platform.ProviderOperationOpenAIFilesDelete {
		h.updateProviderResourceStatus(c, currentTenant.ID, deployment.Provider, resourceKind, publicID, "deleted")
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

	if !h.pipeline.HasRawTransport() {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Raw tunnel streaming is unavailable", "tunnel_unavailable")
	}

	input := platform.ProviderOperationInput{ResourceID: strings.TrimSpace(publicID)}

	providerSelector, err := resourceProviderSelector(c, resourceKind, input.ResourceID == "")
	if err != nil {
		return writeAPIError(c, fiber.StatusBadRequest, err.Error(), "invalid_request_error")
	}

	accountSelector := strings.TrimSpace(c.Get("x-godarai-provider-account"))
	if input.ResourceID != "" {
		mapping, mappingErr := h.providerResourceMapping(c, currentTenant.ID, providerSelector, resourceKind, input.ResourceID)
		if mappingErr != nil {
			return mappingErr
		}

		if mapping != nil {
			providerSelector = mapping.Provider
			accountSelector = mapping.ProviderAccountID
			input.ResourceID = mapping.UpstreamID
		}
	}

	deployment, err := h.resolveResponseDeployment(c.Context(), currentTenant.ID, providerSelector, accountSelector, "")
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

		if strings.EqualFold(providerSelector, "xai") && contentLength > xAIFileUploadMaxBytes {
			return writeAPIError(c, fiber.StatusRequestEntityTooLarge, "xAI file uploads cannot exceed 50 MB", "invalid_request_error")
		}
	}

	resp, err := h.pipeline.DispatchRaw(c.Context(), tunnelReq, bodyReader, contentLength)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, fmt.Sprintf("Tunnel error: %v", err), "api_error")
	}

	if resp == nil || resp.Body == nil {
		return writeAPIError(c, fiber.StatusBadGateway, "Provider returned an empty response", "api_error")
	}

	if resp.StatusCode >= httpErrorFloor {
		defer resp.Body.Close() //nolint:errcheck // Cleanup is best-effort and cannot change the completed operation.

		errorBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024)) //nolint:errcheck // Failure is intentionally handled by the surrounding best-effort path.

		return writeOpenAIProviderError(c, resp.StatusCode, errorBody, resp.Headers)
	}

	forwardProviderHeaders(c, resp.Headers)
	setTraceHeaders(c, requestID)

	if contentType := responseContentType(resp.Headers); contentType != "" {
		c.Set(fiber.HeaderContentType, contentType)
	}

	if operation == platform.ProviderOperationOpenAIFilesCreate {
		defer resp.Body.Close() //nolint:errcheck // Cleanup is best-effort and cannot change the completed operation.

		result, readErr := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
		if readErr != nil {
			return writeAPIError(c, fiber.StatusBadGateway, "Invalid OpenAI Files response", "api_error")
		}

		h.recordOpenAIResourceMapping(c, currentTenant.ID, deployment, resourceKind, result)

		return c.Status(resp.StatusCode).Send(result)
	}

	c.Status(resp.StatusCode)
	c.Context().SetBodyStreamWriter(func(writer *bufio.Writer) {
		defer resp.Body.Close() //nolint:errcheck // Cleanup is best-effort and cannot change the completed operation.

		_, _ = io.CopyBuffer(writer, resp.Body, make([]byte, 32*1024)) //nolint:errcheck // A downstream disconnect is handled by terminating the stream.
		_ = writer.Flush()                                             //nolint:errcheck // A downstream disconnect is handled by terminating the stream.
	})

	return nil
}

func openAIListOperationInput(c *fiber.Ctx) (platform.ProviderOperationInput, error) {
	input := platform.ProviderOperationInput{
		AfterID:         strings.TrimSpace(c.Query("after")),
		BeforeID:        strings.TrimSpace(c.Query("before")),
		Order:           strings.ToLower(strings.TrimSpace(c.Query("order"))),
		Purpose:         strings.TrimSpace(c.Query("purpose")),
		PaginationToken: strings.TrimSpace(c.Query("pagination_token")),
		SortBy:          strings.TrimSpace(c.Query("sort_by")),
		Filter:          strings.TrimSpace(c.Query("filter")),
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

func (h *ResponsesHandler) providerResourceMapping(
	c *fiber.Ctx,
	tenantID string,
	providerSelector string,
	resourceKind string,
	publicID string,
) (*platform.ProviderNativeResourceMapping, error) {
	if h.resources == nil {
		return nil, nil
	}

	providers := []string{"openai", "xai"}
	if strings.TrimSpace(providerSelector) != "" {
		providers = []string{strings.ToLower(strings.TrimSpace(providerSelector))}
	}

	var found *platform.ProviderNativeResourceMapping

	for _, provider := range providers {
		mapping, err := h.resources.GetProviderNativeResourceMapping(c.Context(), tenantID, provider, resourceKind, publicID)
		if err == nil {
			if found != nil {
				return nil, writeAPIError(c, fiber.StatusConflict, "resource id is ambiguous; specify x-godarai-provider", "invalid_request_error")
			}

			found = mapping

			continue
		}

		if !errors.Is(err, providernative.ErrResourceNotFound) {
			return nil, writeAPIError(c, fiber.StatusServiceUnavailable, "provider resource routing state is unavailable", "routing_error")
		}
	}

	if found == nil {
		return nil, writeAPIError(c, fiber.StatusNotFound, "provider resource not found", "invalid_request_error")
	}

	return found, nil
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
		ID                 string      `json:"id"`
		Status             string      `json:"status"`
		Bytes              json.Number `json:"bytes"`
		Filename           string      `json:"filename"`
		ContentType        string      `json:"content_type"`
		CreatedAt          json.Number `json:"created_at"`
		ExpiresAt          json.Number `json:"expires_at"`
		PublicURL          string      `json:"public_url"`
		PublicURLExpiresAt json.Number `json:"public_url_expires_at"`
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()

	if decoder.Decode(&result) != nil || strings.TrimSpace(result.ID) == "" {
		return
	}

	if strings.TrimSpace(result.Status) == "" {
		result.Status = "active"
	}

	metadata := map[string]interface{}{"deployment_id": deployment.ID}
	copyStringMetadata(metadata, "filename", result.Filename)
	copyStringMetadata(metadata, "content_type", result.ContentType)
	copyStringMetadata(metadata, "public_url", result.PublicURL)
	copyNumberMetadata(metadata, "bytes", result.Bytes)
	copyNumberMetadata(metadata, "created_at", result.CreatedAt)
	copyNumberMetadata(metadata, "expires_at", result.ExpiresAt)
	copyNumberMetadata(metadata, "public_url_expires_at", result.PublicURLExpiresAt)

	var expiresAt *time.Time

	if unix, ok := numberInt64(result.ExpiresAt); ok {
		value := time.Unix(unix, 0).UTC()
		expiresAt = &value
	}

	_, _ = h.resources.UpsertProviderNativeResourceMapping(c.Context(), platform.ProviderNativeResourceMapping{ //nolint:errcheck // Failure is intentionally handled by the surrounding best-effort path.
		PublicID: result.ID, TenantID: tenantID, Provider: strings.ToLower(strings.TrimSpace(deployment.Provider)),
		ProviderAccountID: deployment.ProviderAccountID, ResourceKind: resourceKind,
		UpstreamID: result.ID, Status: result.Status,
		Metadata: metadata, ExpiresAt: expiresAt,
	})
}

func validateXAIFilePublicURLRequest(mapping *platform.ProviderNativeResourceMapping, expiresAfter *int, now time.Time) string {
	if mapping == nil {
		return ""
	}

	if size, ok := metadataInt64(mapping.Metadata, "bytes"); ok && size > 50*1024*1024 {
		return "files larger than 50 MiB cannot have an xAI public URL"
	}

	if contentType, _ := mapping.Metadata["content_type"].(string); strings.TrimSpace(contentType) != "" { //nolint:errcheck // Optional metadata uses the zero value when absent or mismatched.
		switch strings.ToLower(strings.TrimSpace(contentType)) {
		case "image/png", "image/jpeg", "video/mp4", "application/pdf":
		default:
			return "file content_type is not eligible for an xAI public URL"
		}
	}

	if expiresAfter != nil && mapping.ExpiresAt != nil && now.Add(time.Duration(*expiresAfter)*time.Second).After(*mapping.ExpiresAt) {
		return "expires_after cannot exceed the file's remaining lifetime"
	}

	return ""
}

func (h *ResponsesHandler) recordXAIFilePublicURLState(
	c *fiber.Ctx,
	mapping *platform.ProviderNativeResourceMapping,
	body []byte,
	revoked bool,
) {
	if h.resources == nil || mapping == nil {
		return
	}

	updated := *mapping

	updated.Metadata = cloneResourceMetadata(mapping.Metadata)
	if revoked {
		delete(updated.Metadata, "public_url")
		delete(updated.Metadata, "public_url_expires_at")
	} else {
		var result struct {
			PublicURL string      `json:"public_url"`
			ExpiresAt json.Number `json:"expires_at"`
		}

		decoder := json.NewDecoder(bytes.NewReader(body))
		decoder.UseNumber()

		if decoder.Decode(&result) != nil || strings.TrimSpace(result.PublicURL) == "" {
			return
		}

		updated.Metadata["public_url"] = result.PublicURL
		copyNumberMetadata(updated.Metadata, "public_url_expires_at", result.ExpiresAt)
	}

	_, _ = h.resources.UpsertProviderNativeResourceMapping(c.Context(), updated) //nolint:errcheck // Failure is intentionally handled by the surrounding best-effort path.
}

func (h *ResponsesHandler) recordProviderFileMetadata(
	c *fiber.Ctx,
	mapping *platform.ProviderNativeResourceMapping,
	body []byte,
) {
	if h.resources == nil || mapping == nil {
		return
	}

	var result struct {
		Bytes              json.Number `json:"bytes"`
		Filename           string      `json:"filename"`
		ContentType        string      `json:"content_type"`
		CreatedAt          json.Number `json:"created_at"`
		ExpiresAt          json.Number `json:"expires_at"`
		PublicURL          string      `json:"public_url"`
		PublicURLExpiresAt json.Number `json:"public_url_expires_at"`
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()

	if decoder.Decode(&result) != nil {
		return
	}

	updated := *mapping
	updated.Metadata = cloneResourceMetadata(mapping.Metadata)
	copyStringMetadata(updated.Metadata, "filename", result.Filename)
	copyStringMetadata(updated.Metadata, "content_type", result.ContentType)
	copyStringMetadata(updated.Metadata, "public_url", result.PublicURL)
	copyNumberMetadata(updated.Metadata, "bytes", result.Bytes)
	copyNumberMetadata(updated.Metadata, "created_at", result.CreatedAt)
	copyNumberMetadata(updated.Metadata, "expires_at", result.ExpiresAt)
	copyNumberMetadata(updated.Metadata, "public_url_expires_at", result.PublicURLExpiresAt)

	if unix, ok := numberInt64(result.ExpiresAt); ok {
		value := time.Unix(unix, 0).UTC()
		updated.ExpiresAt = &value
	}

	_, _ = h.resources.UpsertProviderNativeResourceMapping(c.Context(), updated) //nolint:errcheck // Failure is intentionally handled by the surrounding best-effort path.
}

func (h *ResponsesHandler) recordProviderFileListMappings(
	c *fiber.Ctx,
	tenantID string,
	deployment platform.Deployment,
	body []byte,
) {
	var result struct {
		Data []json.RawMessage `json:"data"`
	}
	if json.Unmarshal(body, &result) != nil {
		return
	}

	for _, file := range result.Data {
		h.recordOpenAIResourceMapping(c, tenantID, deployment, openAIFileResourceKind, file)
	}
}

func cloneResourceMetadata(source map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(source)+2)
	for key, value := range source {
		result[key] = value
	}

	return result
}

func copyStringMetadata(metadata map[string]interface{}, key, value string) {
	if value = strings.TrimSpace(value); value != "" {
		metadata[key] = value
	}
}

func copyNumberMetadata(metadata map[string]interface{}, key string, value json.Number) {
	if number, ok := numberInt64(value); ok {
		metadata[key] = number
	}
}

func numberInt64(value json.Number) (int64, bool) {
	if strings.TrimSpace(value.String()) == "" {
		return 0, false
	}

	number, err := value.Int64()

	return number, err == nil
}

func metadataInt64(metadata map[string]interface{}, key string) (int64, bool) {
	value, ok := metadata[key]
	if !ok {
		return 0, false
	}

	switch typed := value.(type) {
	case int64:
		return typed, true
	case int:
		return int64(typed), true
	case float64:
		return int64(typed), true
	case json.Number:
		return numberInt64(typed)
	case string:
		number, err := strconv.ParseInt(typed, 10, 64)
		return number, err == nil
	default:
		return 0, false
	}
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

		_, _ = h.resources.UpsertProviderNativeResourceMapping(c.Context(), platform.ProviderNativeResourceMapping{ //nolint:errcheck // Failure is intentionally handled by the surrounding best-effort path.
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
	h.updateProviderResourceStatus(c, tenantID, "openai", resourceKind, publicID, status)
}

func (h *ResponsesHandler) updateProviderResourceStatus(
	c *fiber.Ctx,
	tenantID string,
	provider string,
	resourceKind string,
	publicID string,
	status string,
) {
	if h.resources != nil {
		_ = h.resources.UpdateProviderNativeResourceStatus(c.Context(), tenantID, strings.ToLower(strings.TrimSpace(provider)), resourceKind, publicID, status) //nolint:errcheck // Failure is intentionally handled by the surrounding best-effort path.
	}
}

func resourceProviderSelector(c *fiber.Ctx, resourceKind string, defaultForCollection bool) (string, error) {
	if resourceKind != openAIFileResourceKind {
		return "openai", nil
	}

	provider := strings.ToLower(strings.TrimSpace(c.Get("x-godarai-provider")))
	if provider == "" && defaultForCollection {
		return "openai", nil
	}

	if provider != "" && provider != "openai" && provider != "xai" {
		return "", fmt.Errorf("x-godarai-provider must be openai or xai")
	}

	return provider, nil
}
