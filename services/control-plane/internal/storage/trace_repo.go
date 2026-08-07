package storage

import (
	"context"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
)

// ErrTraceStoreUnavailable is retained for storage-adapter compatibility.
var ErrTraceStoreUnavailable = observability.ErrTraceStoreUnavailable

// Export persists a batch of trace span events.
func (s *TraceRepository) Export(ctx context.Context, batch []tracing.SpanEvent) error {
	if !s.clickHouseTraceReady() {
		return nil
	}

	return s.exportToClickHouse(ctx, batch)
}

func (s *TraceRepository) ListRequestTraces(ctx context.Context, params platform.ListRequestTracesParams) ([]platform.RequestTrace, int, error) {
	if !s.clickHouseTraceReady() {
		return nil, 0, ErrTraceStoreUnavailable
	}

	return s.listRequestTracesFromClickHouse(ctx, params)
}

func (s *TraceRepository) GetRequestTrace(ctx context.Context, tenantID string, traceID string) (*platform.RequestTrace, error) {
	if !s.clickHouseTraceReady() {
		return nil, ErrTraceStoreUnavailable
	}

	return s.getRequestTraceFromClickHouse(ctx, tenantID, traceID)
}

func (s *TraceRepository) ListTraceSpans(ctx context.Context, tenantID string, traceID string) ([]platform.TraceSpan, error) {
	if !s.clickHouseTraceReady() {
		return nil, ErrTraceStoreUnavailable
	}

	return s.listTraceSpansFromClickHouse(ctx, tenantID, traceID)
}

func (s *TraceRepository) ListRequestTracesV2(ctx context.Context, params platform.ListRequestTracesV2Params) ([]platform.RequestTraceV2, int, error) {
	if !s.clickHouseTraceReady() {
		return nil, 0, ErrTraceStoreUnavailable
	}

	return s.listRequestTracesV2FromClickHouse(ctx, params)
}

func (s *TraceRepository) GetRequestTraceV2(ctx context.Context, tenantID string, traceID string) (*platform.RequestTraceDetailV2, error) {
	if !s.clickHouseTraceReady() {
		return nil, ErrTraceStoreUnavailable
	}

	return s.getRequestTraceV2FromClickHouse(ctx, tenantID, traceID)
}

func (s *TraceRepository) ListTraceTimelineV2(ctx context.Context, tenantID string, traceID string) ([]platform.TraceSpan, error) {
	if !s.clickHouseTraceReady() {
		return nil, ErrTraceStoreUnavailable
	}

	return s.listTraceSpansFromClickHouse(ctx, tenantID, traceID)
}

func (s *TraceRepository) ListTraceAttemptsV2(ctx context.Context, tenantID string, traceID string) ([]platform.TraceAttemptV2, error) {
	if !s.clickHouseTraceReady() {
		return nil, ErrTraceStoreUnavailable
	}

	return s.listTraceAttemptsV2FromClickHouse(ctx, tenantID, traceID)
}

func (s *TraceRepository) GetTraceFacetsV2(ctx context.Context, params platform.ListRequestTracesV2Params) (*platform.TraceFacetsV2, error) {
	if !s.clickHouseTraceReady() {
		return nil, ErrTraceStoreUnavailable
	}

	return s.getTraceFacetsV2FromClickHouse(ctx, params)
}
