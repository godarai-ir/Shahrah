package storage

import (
	"context"
	"errors"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
)

var ErrTraceStoreUnavailable = errors.New("trace store unavailable")

// Export persists a batch of trace span events.
func (s *ObservabilityRepository) Export(ctx context.Context, batch []tracing.SpanEvent) error {
	if !s.clickHouseTraceReady() {
		return nil
	}

	return s.exportToClickHouse(ctx, batch)
}

func (s *ObservabilityRepository) ListRequestTraces(ctx context.Context, params platform.ListRequestTracesParams) ([]platform.RequestTrace, int, error) {
	if !s.clickHouseTraceReady() {
		return nil, 0, ErrTraceStoreUnavailable
	}

	return s.listRequestTracesFromClickHouse(ctx, params)
}

func (s *ObservabilityRepository) GetRequestTrace(ctx context.Context, tenantID string, traceID string) (*platform.RequestTrace, error) {
	if !s.clickHouseTraceReady() {
		return nil, ErrTraceStoreUnavailable
	}

	return s.getRequestTraceFromClickHouse(ctx, tenantID, traceID)
}

func (s *ObservabilityRepository) ListTraceSpans(ctx context.Context, tenantID string, traceID string) ([]platform.TraceSpan, error) {
	if !s.clickHouseTraceReady() {
		return nil, ErrTraceStoreUnavailable
	}

	return s.listTraceSpansFromClickHouse(ctx, tenantID, traceID)
}

func (s *ObservabilityRepository) ListRequestTracesV2(ctx context.Context, params platform.ListRequestTracesV2Params) ([]platform.RequestTraceV2, int, error) {
	if !s.clickHouseTraceReady() {
		return nil, 0, ErrTraceStoreUnavailable
	}

	return s.listRequestTracesV2FromClickHouse(ctx, params)
}

func (s *ObservabilityRepository) GetRequestTraceV2(ctx context.Context, tenantID string, traceID string) (*platform.RequestTraceDetailV2, error) {
	if !s.clickHouseTraceReady() {
		return nil, ErrTraceStoreUnavailable
	}

	return s.getRequestTraceV2FromClickHouse(ctx, tenantID, traceID)
}

func (s *ObservabilityRepository) ListTraceTimelineV2(ctx context.Context, tenantID string, traceID string) ([]platform.TraceSpan, error) {
	if !s.clickHouseTraceReady() {
		return nil, ErrTraceStoreUnavailable
	}

	return s.listTraceSpansFromClickHouse(ctx, tenantID, traceID)
}

func (s *ObservabilityRepository) ListTraceAttemptsV2(ctx context.Context, tenantID string, traceID string) ([]platform.TraceAttemptV2, error) {
	if !s.clickHouseTraceReady() {
		return nil, ErrTraceStoreUnavailable
	}

	return s.listTraceAttemptsV2FromClickHouse(ctx, tenantID, traceID)
}

func (s *ObservabilityRepository) GetTraceFacetsV2(ctx context.Context, params platform.ListRequestTracesV2Params) (*platform.TraceFacetsV2, error) {
	if !s.clickHouseTraceReady() {
		return nil, ErrTraceStoreUnavailable
	}

	return s.getTraceFacetsV2FromClickHouse(ctx, params)
}
