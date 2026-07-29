package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	prommetrics "github.com/Delpak/shahrah/services/control-plane/internal/observability/prometheus"
	"go.uber.org/zap"
)

func (h *AnthropicNativeHandler) RunBatchWorker(ctx context.Context, interval time.Duration) {
	if h == nil || h.resources == nil || h.tunnel == nil || !h.batchesEnabled {
		return
	}
	if interval <= 0 {
		interval = 15 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	h.reconcileAnthropicBatches(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.reconcileAnthropicBatches(ctx)
		}
	}
}

func (h *AnthropicNativeHandler) reconcileAnthropicBatches(ctx context.Context) {
	mappings, err := h.resources.ListPendingAnthropicBatchMappings(ctx, 100)
	if err != nil {
		if h.logger != nil {
			h.logger.Warn("anthropic batch reconciliation failed", zap.Error(err))
		}
		return
	}
	for index := range mappings {
		if err := h.reconcileAnthropicBatch(ctx, &mappings[index]); err != nil && h.logger != nil {
			h.logger.Warn("anthropic batch item reconciliation failed", zap.Error(err))
		}
	}
}

func (h *AnthropicNativeHandler) reconcileAnthropicBatch(ctx context.Context, mapping *platform.AnthropicNativeResourceMapping) error {
	deployment, _, err := h.resolveDeployment(ctx, mapping.TenantID, nil, mapping.ProviderAccountID)
	if err != nil {
		return err
	}
	request := platform.TunnelRequest{
		RequestID:      fmt.Sprintf("batch_reconcile_%d", time.Now().UnixNano()),
		TenantID:       mapping.TenantID,
		Operation:      platform.ProviderOperationAnthropicBatchesGet,
		OperationInput: platform.ProviderOperationInput{ResourceID: mapping.UpstreamID},
		Deployment:     *deployment,
	}
	response, err := h.tunnel.Proxy(ctx, request)
	if err != nil {
		return err
	}
	if response.StatusCode >= 400 {
		return fmt.Errorf("batch status returned HTTP %d", response.StatusCode)
	}
	var status struct {
		ProcessingStatus string `json:"processing_status"`
	}
	if err := json.Unmarshal(response.Body, &status); err != nil {
		return err
	}
	switch status.ProcessingStatus {
	case "ended":
		prommetrics.RecordAnthropicBatchState("ended")
		if mapping.Reservation == nil || mapping.EstimatedUsage == nil {
			return h.resources.UpdateAnthropicBatchStatus(ctx, mapping.TenantID, mapping.PublicID, "ended")
		}
		rawTunnel, ok := h.tunnel.(platform.TunnelRawTransport)
		if !ok {
			return fmt.Errorf("raw tunnel streaming is unavailable for batch reconciliation")
		}
		request.Operation = platform.ProviderOperationAnthropicBatchesResults
		results, err := rawTunnel.ProxyRaw(ctx, request, nil, 0)
		if err != nil {
			return err
		}
		defer results.Body.Close()
		if results.StatusCode >= 400 {
			return fmt.Errorf("batch results returned HTTP %d", results.StatusCode)
		}
		return h.finalizeBatchResultsReader(ctx, results.Body, mapping)
	case "canceling":
		prommetrics.RecordAnthropicBatchState("canceling")
		return h.resources.UpdateAnthropicBatchStatus(ctx, mapping.TenantID, mapping.PublicID, "canceling")
	default:
		prommetrics.RecordAnthropicBatchState("in_progress")
		return nil
	}
}
