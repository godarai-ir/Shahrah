package proxy

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"go.uber.org/zap"
)

type mistralBatchResourceStore interface {
	mistralNativeResourceStore
	ClaimProviderNativeResourceStatus(
		ctx context.Context,
		tenantID string,
		provider string,
		resourceKind string,
		publicID string,
		fromStatuses []string,
		toStatus string,
	) (bool, error)
	ListPendingProviderNativeResourceMappings(
		ctx context.Context,
		provider string,
		resourceKind string,
		statuses []string,
		limit int,
	) ([]platform.ProviderNativeResourceMapping, error)
}

// RunBatchWorker reconciles Mistral jobs and settles each reservation once.
func (h *MistralNativeHandler) RunBatchWorker(ctx context.Context, interval time.Duration) {
	if h == nil {
		return
	}

	resources, ok := h.resources.(mistralBatchResourceStore)
	if !ok || !h.pipeline.HasTransport() || h.billing == nil || !h.batchEnabled {
		return
	}

	if interval <= 0 {
		interval = 15 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	h.reconcileMistralBatches(ctx, resources)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.reconcileMistralBatches(ctx, resources)
		}
	}
}

func (h *MistralNativeHandler) reconcileMistralBatches(ctx context.Context, resources mistralBatchResourceStore) {
	statuses := []string{"queued", "running", "in_progress", "cancellation_requested", "reconciling"}

	mappings, err := resources.ListPendingProviderNativeResourceMappings(ctx, mistralProviderName, mistralBatchResourceKind, statuses, 100)
	if err != nil {
		h.logMistralBatchWarning("Mistral batch reconciliation failed", err)

		return
	}

	for index := range mappings {
		mapping := &mappings[index]

		fromStatuses := []string{mapping.Status}
		if mapping.Status == "reconciling" {
			fromStatuses = nil
		}

		claimed, claimErr := resources.ClaimProviderNativeResourceStatus(
			ctx,
			mapping.TenantID,
			mistralProviderName,
			mistralBatchResourceKind,
			mapping.PublicID,
			fromStatuses,
			"reconciling",
		)
		if claimErr != nil {
			h.logMistralBatchWarning("Mistral batch claim failed", claimErr)

			continue
		}

		if !claimed {
			continue
		}

		if err := h.reconcileMistralBatch(ctx, resources, mapping); err != nil {
			h.logMistralBatchWarning("Mistral batch item reconciliation failed", err)

			retryStatus := mapping.Status
			if retryStatus == "reconciling" {
				retryStatus = "in_progress"
			}

			_ = resources.UpdateProviderNativeResourceStatus(ctx, mapping.TenantID, mistralProviderName, mistralBatchResourceKind, mapping.PublicID, retryStatus) //nolint:errcheck // A later reconciliation retries the provider request.
		}
	}
}

func (h *MistralNativeHandler) reconcileMistralBatch(ctx context.Context, resources mistralBatchResourceStore, mapping *platform.ProviderNativeResourceMapping) error {
	model := ""
	if mapping.EstimatedUsage != nil {
		model = mapping.EstimatedUsage.Model
	}

	deployment, _, err := h.resolveDeployment(ctx, mapping.TenantID, model, mapping.ProviderAccountID)
	if err != nil {
		return err
	}

	request := platform.TunnelRequest{
		RequestID: fmt.Sprintf("mistral_batch_reconcile_%d", time.Now().UnixNano()),
		TenantID:  mapping.TenantID, Operation: platform.ProviderOperationMistralBatchGet,
		OperationInput: platform.ProviderOperationInput{ResourceID: mapping.UpstreamID},
		Deployment:     deployment,
	}

	response, err := h.pipeline.Dispatch(ctx, request)
	if err != nil {
		return err
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("mistral batch status returned HTTP %d", response.StatusCode)
	}

	status := mistralResponseStatus(response.Body)
	switch status {
	case "success", "succeeded", "completed":
		return h.finalizeMistralBatch(ctx, resources, request, response.Body, mapping)
	case "failed", "timeout_exceeded", "cancelled", "canceled": //nolint:misspell // Mistral may return either provider spelling.
		h.pipeline.Release(ctx, mapping.Reservation, "batch_"+status)

		return resources.UpdateProviderNativeResourceStatus(ctx, mapping.TenantID, mistralProviderName, mistralBatchResourceKind, mapping.PublicID, status)
	case "cancellation_requested", "canceling":
		return resources.UpdateProviderNativeResourceStatus(ctx, mapping.TenantID, mistralProviderName, mistralBatchResourceKind, mapping.PublicID, "cancellation_requested")
	default:
		if status == "" {
			status = "in_progress"
		}

		return resources.UpdateProviderNativeResourceStatus(ctx, mapping.TenantID, mistralProviderName, mistralBatchResourceKind, mapping.PublicID, status)
	}
}

func (h *MistralNativeHandler) finalizeMistralBatch(ctx context.Context, resources mistralBatchResourceStore, request platform.TunnelRequest, statusBody []byte, mapping *platform.ProviderNativeResourceMapping) error {
	if mapping.EstimatedUsage == nil {
		return fmt.Errorf("mistral batch pricing snapshot is unavailable")
	}

	outputFile := mistralBatchOutputFileID(statusBody)
	if outputFile == "" {
		return fmt.Errorf("successful Mistral batch did not include an output file")
	}

	if !h.pipeline.HasRawTransport() {
		return fmt.Errorf("raw tunnel streaming is unavailable for Mistral batch settlement")
	}

	request.Operation = platform.ProviderOperationMistralFilesContent
	request.OperationInput.ResourceID = outputFile

	result, err := h.pipeline.DispatchRaw(ctx, request, nil, 0)
	if err != nil {
		return err
	}

	if result == nil || result.Body == nil {
		return fmt.Errorf("mistral batch output was empty")
	}
	defer result.Body.Close() //nolint:errcheck // Cleanup cannot alter settlement.

	if result.StatusCode >= 400 {
		return fmt.Errorf("mistral batch output returned HTTP %d", result.StatusCode)
	}

	metrics, err := parseMistralBatchMetrics(result.Body)
	if err != nil {
		return err
	}

	var finalized platform.UsageCostBreakdown

	switch endpoint := stringFromMap(mapping.Metadata, "runtime_endpoint"); endpoint {
	case "ocr":
		if metrics.Pages <= 0 {
			return fmt.Errorf("mistral OCR batch output did not report processed pages")
		}

		finalized, err = priceMeteredUsage(ctx, h.billing, mapping.EstimatedUsage.Model, "ocr_input_pages", fmt.Sprintf("%d", metrics.Pages), *mapping.EstimatedUsage)
	case "audio_transcriptions":
		if metrics.AudioSeconds <= 0 {
			return fmt.Errorf("mistral transcription batch output did not report audio seconds")
		}

		finalized, err = priceMeteredUsage(ctx, h.billing, mapping.EstimatedUsage.Model, "audio_input_seconds", fmt.Sprintf("%g", metrics.AudioSeconds), *mapping.EstimatedUsage)
	default:
		if !metrics.HasTokenUsage {
			return fmt.Errorf("mistral batch output did not report token usage")
		}

		finalizer, ok := h.billing.(batchUsageFinalizer)
		if !ok {
			return fmt.Errorf("mistral batch billing finalizer is unavailable")
		}

		finalized, err = finalizer.FinalizeBatchFromUsage(ctx, mapping.EstimatedUsage.Model, metrics.Usage, *mapping.EstimatedUsage)
	}

	if err != nil {
		return err
	}

	if err := h.pipeline.Settle(ctx, SettlementRequest{Reservation: mapping.Reservation, TenantID: mapping.TenantID, Actual: finalized, RequestID: mapping.ReservationID, BillingMode: request.Deployment.BillingMode}); err != nil {
		return err
	}

	return resources.UpdateProviderNativeResourceStatus(ctx, mapping.TenantID, mistralProviderName, mistralBatchResourceKind, mapping.PublicID, "success")
}

func mistralBatchOutputFileID(body []byte) string {
	var value interface{}
	if json.Unmarshal(body, &value) != nil {
		return ""
	}

	return findMistralOutputFileID(value)
}

func findMistralOutputFileID(value interface{}) string {
	switch typed := value.(type) {
	case map[string]interface{}:
		for _, key := range []string{"output_file", "output_file_id"} {
			if id := strings.TrimSpace(stringFromMap(typed, key)); id != "" {
				return id
			}
		}

		for _, child := range typed {
			if id := findMistralOutputFileID(child); id != "" {
				return id
			}
		}
	case []interface{}:
		for _, child := range typed {
			if id := findMistralOutputFileID(child); id != "" {
				return id
			}
		}
	}

	return ""
}

func parseMistralBatchUsage(reader io.Reader) (*contractsopenai.Usage, error) {
	metrics, err := parseMistralBatchMetrics(reader)
	if err != nil {
		return nil, err
	}

	return metrics.Usage, nil
}

type mistralBatchMetrics struct {
	Usage         *contractsopenai.Usage
	Pages         int
	AudioSeconds  float64
	HasTokenUsage bool
}

func parseMistralBatchMetrics(reader io.Reader) (mistralBatchMetrics, error) {
	metrics := mistralBatchMetrics{Usage: &contractsopenai.Usage{ServerToolUse: make(map[string]int)}}
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), 32*1024*1024)

	for scanner.Scan() {
		var row interface{}
		if err := json.Unmarshal(scanner.Bytes(), &row); err != nil {
			return mistralBatchMetrics{}, err
		}

		if usage := mistralUsageFromValue(row); usage != nil {
			maxInt := int(^uint(0) >> 1)
			if metrics.Usage.PromptTokens > maxInt-usage.PromptTokens ||
				metrics.Usage.CompletionTokens > maxInt-usage.CompletionTokens ||
				metrics.Usage.TotalTokens > maxInt-usage.TotalTokens {
				return mistralBatchMetrics{}, fmt.Errorf("mistral batch token usage is too large")
			}

			metrics.Usage.PromptTokens += usage.PromptTokens
			metrics.Usage.CompletionTokens += usage.CompletionTokens
			metrics.Usage.TotalTokens += usage.TotalTokens
			metrics.HasTokenUsage = true
		}

		accumulateMistralMeteredValue(&metrics, row)
	}

	if err := scanner.Err(); err != nil {
		return mistralBatchMetrics{}, err
	}

	return metrics, nil
}

func accumulateMistralMeteredValue(metrics *mistralBatchMetrics, value interface{}) {
	switch typed := value.(type) {
	case map[string]interface{}:
		reported := false

		for _, usageKey := range []string{"usage", "usage_info"} {
			usage, ok := typed[usageKey].(map[string]interface{})
			if !ok {
				continue
			}

			pages, hasPages, _ := numberFromMap(usage, "pages_processed") //nolint:errcheck // Missing provider usage fields are zero.

			seconds, hasSeconds, _ := numberFromMap(usage, "prompt_audio_seconds") //nolint:errcheck // Missing provider usage fields are zero.
			if !hasSeconds {
				seconds, hasSeconds, _ = numberFromMap(usage, "audio_seconds") //nolint:errcheck // Missing provider usage fields are zero.
			}

			if hasPages {
				metrics.Pages += int(pages)
				reported = true
			}

			if hasSeconds {
				metrics.AudioSeconds += seconds
				reported = true
			}
		}

		if reported {
			return
		}

		if pages, ok := typed["pages"].([]interface{}); ok {
			metrics.Pages += len(pages)

			return
		}

		for _, child := range typed {
			accumulateMistralMeteredValue(metrics, child)
		}
	case []interface{}:
		for _, child := range typed {
			accumulateMistralMeteredValue(metrics, child)
		}
	}
}

func (h *MistralNativeHandler) logMistralBatchWarning(message string, err error) {
	if h.logger != nil {
		h.logger.Warn(strings.TrimSpace(message), zap.Error(err))
	}
}
