package proxy

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/gofiber/fiber/v2"
)

func isAnthropicFileResourceOperation(operation platform.ProviderOperation) bool {
	switch operation {
	case platform.ProviderOperationAnthropicFilesGet,
		platform.ProviderOperationAnthropicFilesDelete,
		platform.ProviderOperationAnthropicFilesContent:
		return true
	default:
		return false
	}
}

func isAnthropicBatchResourceOperation(operation platform.ProviderOperation) bool {
	switch operation {
	case platform.ProviderOperationAnthropicBatchesGet,
		platform.ProviderOperationAnthropicBatchesCancel,
		platform.ProviderOperationAnthropicBatchesResults:
		return true
	default:
		return false
	}
}

func (h *AnthropicNativeHandler) rewriteNativeResponse(
	ctx context.Context,
	operation platform.ProviderOperation,
	tenantID, accountID string,
	body []byte,
	reservation *platform.BudgetReservation,
	estimated platform.UsageCostBreakdown,
	existing *platform.AnthropicNativeResourceMapping,
) ([]byte, error) {
	if operation == platform.ProviderOperationAnthropicBatchesResults {
		if existing != nil && existing.Reservation != nil && existing.EstimatedUsage != nil {
			if err := h.finalizeBatchResults(ctx, body, existing); err != nil {
				return nil, err
			}
		}

		return body, nil
	}

	if operation != platform.ProviderOperationAnthropicFilesCreate &&
		operation != platform.ProviderOperationAnthropicFilesList &&
		operation != platform.ProviderOperationAnthropicFilesGet &&
		operation != platform.ProviderOperationAnthropicFilesDelete &&
		operation != platform.ProviderOperationAnthropicBatchesCreate &&
		operation != platform.ProviderOperationAnthropicBatchesList &&
		operation != platform.ProviderOperationAnthropicBatchesGet &&
		operation != platform.ProviderOperationAnthropicBatchesCancel {
		return body, nil
	}

	var object map[string]interface{}
	if err := json.Unmarshal(body, &object); err != nil {
		return nil, err
	}

	isFile := operation == platform.ProviderOperationAnthropicFilesCreate ||
		operation == platform.ProviderOperationAnthropicFilesList ||
		operation == platform.ProviderOperationAnthropicFilesGet ||
		operation == platform.ProviderOperationAnthropicFilesDelete

	rewrite := func(item map[string]interface{}) error {
		upstreamID, _ := item["id"].(string) //nolint:errcheck // Optional metadata uses the zero value when absent or mismatched.
		if upstreamID == "" {
			return nil
		}

		if existing != nil && existing.UpstreamID == upstreamID {
			item["id"] = existing.PublicID
			return nil
		}

		status, _ := item["processing_status"].(string) //nolint:errcheck // Optional metadata uses the zero value when absent or mismatched.
		if status == "" {
			status, _ = item["status"].(string) //nolint:errcheck // Optional metadata uses the zero value when absent or mismatched.
		}

		publicID, err := newAnthropicOpaqueID(map[bool]string{true: "file", false: "msgbatch"}[isFile])
		if err != nil {
			return err
		}

		mapping := platform.AnthropicNativeResourceMapping{
			PublicID: publicID,
			TenantID: tenantID, ProviderAccountID: accountID, UpstreamID: upstreamID,
			Status: status, Reservation: reservation, EstimatedUsage: &estimated,
			PricingRevision: estimated.PricingRevision,
		}
		if reservation != nil {
			mapping.ReservationID = reservation.ID
		}

		if isFile {
			persisted, err := h.resources.UpsertAnthropicFileMapping(ctx, mapping)
			if err != nil {
				return err
			}

			item["id"] = persisted.PublicID
		} else {
			persisted, err := h.resources.UpsertAnthropicBatchMapping(ctx, mapping)
			if err != nil {
				return err
			}

			item["id"] = persisted.PublicID
		}

		return nil
	}
	if data, ok := object["data"].([]interface{}); ok {
		for _, raw := range data {
			if item, ok := raw.(map[string]interface{}); ok {
				if err := rewrite(item); err != nil {
					return nil, err
				}
			}
		}
	} else if err := rewrite(object); err != nil {
		return nil, err
	}

	return json.Marshal(object)
}

func newAnthropicOpaqueID(prefix string) (string, error) {
	var value [18]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}

	return prefix + "_gdr_" + base64.RawURLEncoding.EncodeToString(value[:]), nil
}

type batchUsageFinalizer interface {
	FinalizeBatchFromUsage(ctx context.Context, model string, usage *contractsopenai.Usage, estimated platform.UsageCostBreakdown) (platform.UsageCostBreakdown, error)
}

func (h *AnthropicNativeHandler) finalizeBatchResults(ctx context.Context, body []byte, mapping *platform.AnthropicNativeResourceMapping) error {
	return h.finalizeBatchResultsReader(ctx, bytes.NewReader(body), mapping)
}

func (h *AnthropicNativeHandler) finalizeBatchResultsReader(ctx context.Context, reader io.Reader, mapping *platform.AnthropicNativeResourceMapping) error {
	finalizer := h.pipeline.batchUsageFinalizer()
	if finalizer == nil {
		return fmt.Errorf("batch billing finalizer is unavailable")
	}

	usage := &contractsopenai.Usage{ServerToolUse: map[string]int{}}
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), 32*1024*1024)

	for scanner.Scan() {
		var row struct {
			Result struct {
				Type    string          `json:"type"`
				Message json.RawMessage `json:"message"`
			} `json:"result"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &row); err != nil {
			return err
		}

		if row.Result.Type != "succeeded" || len(row.Result.Message) == 0 {
			continue
		}

		item := parseAnthropicNativeUsage(row.Result.Message)
		if item == nil {
			continue
		}

		usage.PromptTokens += item.PromptTokens
		usage.CompletionTokens += item.CompletionTokens
		usage.TotalTokens += item.TotalTokens
		usage.CacheCreationInputTokens += item.CacheCreationInputTokens
		usage.CacheCreation5mInputTokens += item.CacheCreation5mInputTokens
		usage.CacheCreation1hInputTokens += item.CacheCreation1hInputTokens

		usage.CacheReadInputTokens += item.CacheReadInputTokens
		if item.ServiceTier != "" {
			usage.ServiceTier = item.ServiceTier
		}

		if item.Speed != "" {
			usage.Speed = item.Speed
		}

		for dimension, quantity := range item.ServerToolUse {
			usage.ServerToolUse[dimension] += quantity
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	finalized, err := finalizer.FinalizeBatchFromUsage(ctx, mapping.EstimatedUsage.Model, usage, *mapping.EstimatedUsage)
	if err != nil {
		return err
	}

	if err := h.pipeline.Settle(ctx, SettlementRequest{Reservation: mapping.Reservation, TenantID: mapping.TenantID, Actual: finalized, RequestID: mapping.ReservationID}); err != nil {
		return err
	}

	return h.resources.UpdateAnthropicBatchStatus(ctx, mapping.TenantID, mapping.PublicID, "ended")
}

func (h *AnthropicNativeHandler) handleRawOperation(c *fiber.Ctx, tenantID string, operation platform.ProviderOperation, input platform.ProviderOperationInput) error {
	if h.resources == nil {
		return writeAnthropicError(c, fiber.StatusServiceUnavailable, "api_error", "Native resource mapping is unavailable")
	}

	if !h.pipeline.HasRawTransport() {
		return writeAnthropicError(c, fiber.StatusServiceUnavailable, "api_error", "Raw tunnel streaming is unavailable")
	}

	accountSelector := c.Get("x-godarai-provider-account")

	var (
		mapping *platform.AnthropicNativeResourceMapping
		err     error
	)

	switch operation {
	case platform.ProviderOperationAnthropicFilesContent:
		mapping, err = h.resources.GetAnthropicFileMapping(c.Context(), tenantID, input.ResourceID)
	case platform.ProviderOperationAnthropicBatchesResults:
		mapping, err = h.resources.GetAnthropicBatchMapping(c.Context(), tenantID, input.ResourceID)
	}

	if operation != platform.ProviderOperationAnthropicFilesCreate {
		if err != nil || mapping == nil {
			return writeAnthropicError(c, fiber.StatusNotFound, "not_found_error", "Resource not found")
		}

		accountSelector = mapping.ProviderAccountID
		input.ResourceID = mapping.UpstreamID
	}

	deployment, _, err := h.resolveDeployment(c.Context(), tenantID, nil, accountSelector)
	if err != nil {
		return writeAnthropicError(c, fiber.StatusBadRequest, "ambiguous_provider_account", err.Error())
	}

	if deployment == nil || !h.accountEnabled(deployment.ProviderAccountID) {
		return writeAnthropicError(c, fiber.StatusForbidden, "permission_error", "Anthropic native API is not enabled for this provider account")
	}

	request := platform.TunnelRequest{
		RequestID: fmt.Sprintf("req_%d", time.Now().UnixNano()), TenantID: tenantID,
		Operation: operation, OperationInput: input, Deployment: *deployment,
		Headers: map[string]string{
			"content-type":   c.Get(fiber.HeaderContentType),
			"anthropic-beta": c.Get("anthropic-beta"),
		},
	}

	var body io.Reader = httpNoBodyReader{}

	contentLength := int64(0)

	if operation == platform.ProviderOperationAnthropicFilesCreate {
		body = c.Context().RequestBodyStream()
		if body == nil {
			body = bytes.NewReader(c.Body())
		}

		contentLength = int64(c.Request().Header.ContentLength())
	}

	response, err := h.pipeline.DispatchRaw(c.Context(), request, body, contentLength)
	if err != nil {
		return writeAnthropicError(c, fiber.StatusBadGateway, "api_error", "Raw tunnel request failed")
	}

	if response.StatusCode >= fiber.StatusBadRequest {
		defer response.Body.Close() //nolint:errcheck // Cleanup is best-effort and cannot change the completed operation.

		errorBody, _ := io.ReadAll(io.LimitReader(response.Body, 2*1024*1024)) //nolint:errcheck // Failure is intentionally handled by the surrounding best-effort path.
		forwardProviderHeaders(c, response.Headers)

		return c.Status(response.StatusCode).Send(errorBody)
	}

	forwardProviderHeaders(c, response.Headers)

	if contentType := mapHeaderValue(response.Headers, "content-type"); contentType != "" {
		c.Set(fiber.HeaderContentType, contentType)
	}

	if operation == platform.ProviderOperationAnthropicFilesCreate {
		defer response.Body.Close() //nolint:errcheck // Cleanup is best-effort and cannot change the completed operation.

		result, err := io.ReadAll(io.LimitReader(response.Body, 2*1024*1024))
		if err != nil {
			return writeAnthropicError(c, fiber.StatusBadGateway, "api_error", "Invalid Anthropic Files response")
		}

		result, err = h.rewriteNativeResponse(c.Context(), operation, tenantID, deployment.ProviderAccountID, result, nil, platform.UsageCostBreakdown{}, nil)
		if err != nil {
			return writeAnthropicError(c, fiber.StatusBadGateway, "api_error", "Failed to persist file mapping")
		}

		return c.Status(response.StatusCode).Send(result)
	}

	requestContext := c.Context()
	requestContext.SetBodyStreamWriter(func(writer *bufio.Writer) {
		defer response.Body.Close() //nolint:errcheck // Cleanup is best-effort and cannot change the completed operation.

		if operation == platform.ProviderOperationAnthropicFilesContent {
			_, _ = io.CopyBuffer(writer, response.Body, make([]byte, 32*1024)) //nolint:errcheck // A downstream disconnect is handled by terminating the stream.
			_ = writer.Flush()                                                 //nolint:errcheck // A downstream disconnect is handled by terminating the stream.

			return
		}

		usage := &contractsopenai.Usage{ServerToolUse: map[string]int{}}
		scanner := bufio.NewScanner(response.Body)
		scanner.Buffer(make([]byte, 64*1024), 32*1024*1024)

		for scanner.Scan() {
			line := append([]byte(nil), scanner.Bytes()...)
			_, _ = writer.Write(line)       //nolint:errcheck // A downstream disconnect is handled by terminating the stream.
			_, _ = writer.WriteString("\n") //nolint:errcheck // A downstream disconnect is handled by terminating the stream.
			_ = writer.Flush()              //nolint:errcheck // A downstream disconnect is handled by terminating the stream.

			accumulateAnthropicBatchUsage(usage, line)
		}

		if scanner.Err() != nil || mapping == nil || mapping.Reservation == nil || mapping.EstimatedUsage == nil {
			return
		}

		finalizer := h.pipeline.batchUsageFinalizer()
		if finalizer == nil {
			return
		}

		finalized, err := finalizer.FinalizeBatchFromUsage(context.Background(), mapping.EstimatedUsage.Model, usage, *mapping.EstimatedUsage)
		if err != nil {
			return
		}

		if h.pipeline.Settle(context.Background(), SettlementRequest{Reservation: mapping.Reservation, TenantID: tenantID, Actual: finalized, RequestID: mapping.ReservationID}) == nil {
			_ = h.resources.UpdateAnthropicBatchStatus(context.Background(), tenantID, mapping.PublicID, "ended") //nolint:errcheck // Failure is intentionally handled by the surrounding best-effort path.
		}
	})

	return nil
}

type httpNoBodyReader struct{}

func (httpNoBodyReader) Read([]byte) (int, error) { return 0, io.EOF }

func mapHeaderValue(headers map[string]string, name string) string {
	for key, value := range headers {
		if strings.EqualFold(key, name) {
			return value
		}
	}

	return ""
}

func accumulateAnthropicBatchUsage(usage *contractsopenai.Usage, line []byte) {
	var row struct {
		Result struct {
			Type    string          `json:"type"`
			Message json.RawMessage `json:"message"`
		} `json:"result"`
	}
	if json.Unmarshal(line, &row) != nil || row.Result.Type != "succeeded" {
		return
	}

	item := parseAnthropicNativeUsage(row.Result.Message)
	if item == nil {
		return
	}

	usage.PromptTokens += item.PromptTokens
	usage.CompletionTokens += item.CompletionTokens
	usage.TotalTokens += item.TotalTokens
	usage.CacheCreationInputTokens += item.CacheCreationInputTokens
	usage.CacheCreation5mInputTokens += item.CacheCreation5mInputTokens
	usage.CacheCreation1hInputTokens += item.CacheCreation1hInputTokens

	usage.CacheReadInputTokens += item.CacheReadInputTokens
	if item.ServiceTier != "" {
		usage.ServiceTier = item.ServiceTier
	}

	if item.Speed != "" {
		usage.Speed = item.Speed
	}

	for dimension, quantity := range item.ServerToolUse {
		usage.ServerToolUse[dimension] += quantity
	}
}
