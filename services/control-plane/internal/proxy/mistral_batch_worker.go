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
	if !ok || h.tunnel == nil || h.billing == nil || !h.batchEnabled {
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
		claimed, claimErr := resources.ClaimProviderNativeResourceStatus(
			ctx,
			mapping.TenantID,
			mistralProviderName,
			mistralBatchResourceKind,
			mapping.PublicID,
			[]string{mapping.Status},
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
			_ = resources.UpdateProviderNativeResourceStatus(ctx, mapping.TenantID, mistralProviderName, mistralBatchResourceKind, mapping.PublicID, mapping.Status) //nolint:errcheck // A later reconciliation retries the provider request.
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
	response, err := h.tunnel.Proxy(ctx, request)
	if err != nil {
		return err
	}
	if response.StatusCode >= 400 {
		return fmt.Errorf("Mistral batch status returned HTTP %d", response.StatusCode)
	}

	status := mistralResponseStatus(response.Body)
	switch status {
	case "success", "succeeded", "completed":
		return h.finalizeMistralBatch(ctx, resources, request, response.Body, mapping)
	case "failed", "timeout_exceeded", "cancelled", "canceled":
		releaseBudgetBestEffort(ctx, h.logger, h.billing, mapping.Reservation, "batch_"+status)

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
	if mapping.Reservation == nil || mapping.EstimatedUsage == nil {
		return resources.UpdateProviderNativeResourceStatus(ctx, mapping.TenantID, mistralProviderName, mistralBatchResourceKind, mapping.PublicID, "success")
	}

	outputFiles := mistralResponseFileReferences(statusBody)
	if len(outputFiles) == 0 {
		return fmt.Errorf("successful Mistral batch did not include an output file")
	}
	rawTunnel, ok := h.tunnel.(platform.TunnelRawTransport)
	if !ok {
		return fmt.Errorf("raw tunnel streaming is unavailable for Mistral batch settlement")
	}

	request.Operation = platform.ProviderOperationMistralFilesContent
	request.OperationInput.ResourceID = outputFiles[0]
	result, err := rawTunnel.ProxyRaw(ctx, request, nil, 0)
	if err != nil {
		return err
	}
	if result == nil || result.Body == nil {
		return fmt.Errorf("Mistral batch output was empty")
	}
	defer result.Body.Close() //nolint:errcheck // Cleanup cannot alter settlement.

	if result.StatusCode >= 400 {
		return fmt.Errorf("Mistral batch output returned HTTP %d", result.StatusCode)
	}

	metrics, err := parseMistralBatchMetrics(result.Body)
	if err != nil {
		return err
	}

	var finalized platform.UsageCostBreakdown
	switch endpoint, _ := mapping.Metadata["runtime_endpoint"].(string); endpoint {
	case "ocr":
		if metrics.Pages <= 0 {
			return fmt.Errorf("Mistral OCR batch output did not report processed pages")
		}

		finalized, err = priceMeteredUsage(ctx, h.billing, mapping.EstimatedUsage.Model, "ocr_input_pages", fmt.Sprintf("%d", metrics.Pages), *mapping.EstimatedUsage)
	case "audio_transcriptions":
		if metrics.AudioSeconds <= 0 {
			return fmt.Errorf("Mistral transcription batch output did not report audio seconds")
		}

		finalized, err = priceMeteredUsage(ctx, h.billing, mapping.EstimatedUsage.Model, "audio_input_seconds", fmt.Sprintf("%g", metrics.AudioSeconds), *mapping.EstimatedUsage)
	default:
		finalizer, ok := h.billing.(batchUsageFinalizer)
		if !ok {
			return fmt.Errorf("Mistral batch billing finalizer is unavailable")
		}

		finalized, err = finalizer.FinalizeBatchFromUsage(ctx, mapping.EstimatedUsage.Model, metrics.Usage, *mapping.EstimatedUsage)
	}
	if err != nil {
		return err
	}
	if err := commitBudgetOrDebit(ctx, h.logger, h.billing, h.wallets, mapping.Reservation, mapping.TenantID, finalized, mapping.ReservationID); err != nil {
		return err
	}

	return resources.UpdateProviderNativeResourceStatus(ctx, mapping.TenantID, mistralProviderName, mistralBatchResourceKind, mapping.PublicID, "success")
}

func parseMistralBatchUsage(reader io.Reader) (*contractsopenai.Usage, error) {
	metrics, err := parseMistralBatchMetrics(reader)
	if err != nil {
		return nil, err
	}

	return metrics.Usage, nil
}

type mistralBatchMetrics struct {
	Usage        *contractsopenai.Usage
	Pages        int
	AudioSeconds float64
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

		accumulateMistralUsageValue(metrics.Usage, row)
		accumulateMistralMeteredValue(&metrics, row)
	}
	if err := scanner.Err(); err != nil {
		return mistralBatchMetrics{}, err
	}

	return metrics, nil
}

func accumulateMistralUsageValue(total *contractsopenai.Usage, value interface{}) {
	switch typed := value.(type) {
	case map[string]interface{}:
		if rawUsage, ok := typed["usage"].(map[string]interface{}); ok {
			prompt, _, _ := numberFromMap(rawUsage, "prompt_tokens")         //nolint:errcheck // Missing usage fields are zero.
			completion, _, _ := numberFromMap(rawUsage, "completion_tokens") //nolint:errcheck // Missing usage fields are zero.
			all, _, _ := numberFromMap(rawUsage, "total_tokens")             //nolint:errcheck // Missing usage fields are zero.
			total.PromptTokens += int(prompt)
			total.CompletionTokens += int(completion)
			total.TotalTokens += int(all)

			return
		}
		for _, child := range typed {
			accumulateMistralUsageValue(total, child)
		}
	case []interface{}:
		for _, child := range typed {
			accumulateMistralUsageValue(total, child)
		}
	}
}

func accumulateMistralMeteredValue(metrics *mistralBatchMetrics, value interface{}) {
	switch typed := value.(type) {
	case map[string]interface{}:
		reported := false
		for _, usageKey := range []string{"usage", "usage_info"} {
			if usage, ok := typed[usageKey].(map[string]interface{}); ok {
				pages, hasPages, _ := numberFromMap(usage, "pages_processed")          //nolint:errcheck // Missing provider usage fields are zero.
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
