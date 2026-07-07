package storage

import (
	"context"
	"errors"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
)

var ErrTraceStoreUnavailable = errors.New("trace store unavailable")

// Export persists a batch of trace span events.
func (s *Store) Export(ctx context.Context, batch []tracing.SpanEvent) error {
	if !s.clickHouseTraceReady() {
		return nil
	}

	return s.exportToClickHouse(ctx, batch)
}

func (s *Store) ListRequestTraces(ctx context.Context, params platform.ListRequestTracesParams) ([]platform.RequestTrace, int, error) {
	if !s.clickHouseTraceReady() {
		return nil, 0, ErrTraceStoreUnavailable
	}

	return s.listRequestTracesFromClickHouse(ctx, params)
}

func (s *Store) GetRequestTrace(ctx context.Context, tenantID string, traceID string) (*platform.RequestTrace, error) {
	if !s.clickHouseTraceReady() {
		return nil, ErrTraceStoreUnavailable
	}

	return s.getRequestTraceFromClickHouse(ctx, tenantID, traceID)
}

func (s *Store) ListTraceSpans(ctx context.Context, tenantID string, traceID string) ([]platform.TraceSpan, error) {
	if !s.clickHouseTraceReady() {
		return nil, ErrTraceStoreUnavailable
	}

	return s.listTraceSpansFromClickHouse(ctx, tenantID, traceID)
}

func (s *Store) ListRequestTracesV2(ctx context.Context, params platform.ListRequestTracesV2Params) ([]platform.RequestTraceV2, int, error) {
	if !s.clickHouseTraceReady() {
		return nil, 0, ErrTraceStoreUnavailable
	}

	return s.listRequestTracesV2FromClickHouse(ctx, params)
}

func (s *Store) GetRequestTraceV2(ctx context.Context, tenantID string, traceID string) (*platform.RequestTraceDetailV2, error) {
	if !s.clickHouseTraceReady() {
		return nil, ErrTraceStoreUnavailable
	}

	return s.getRequestTraceV2FromClickHouse(ctx, tenantID, traceID)
}

func (s *Store) ListTraceTimelineV2(ctx context.Context, tenantID string, traceID string) ([]platform.TraceSpan, error) {
	if !s.clickHouseTraceReady() {
		return nil, ErrTraceStoreUnavailable
	}

	return s.listTraceSpansFromClickHouse(ctx, tenantID, traceID)
}

func (s *Store) ListTraceAttemptsV2(ctx context.Context, tenantID string, traceID string) ([]platform.TraceAttemptV2, error) {
	if !s.clickHouseTraceReady() {
		return nil, ErrTraceStoreUnavailable
	}

	return s.listTraceAttemptsV2FromClickHouse(ctx, tenantID, traceID)
}

func (s *Store) GetTraceFacetsV2(ctx context.Context, params platform.ListRequestTracesV2Params) (*platform.TraceFacetsV2, error) {
	if !s.clickHouseTraceReady() {
		return nil, ErrTraceStoreUnavailable
	}

	return s.getTraceFacetsV2FromClickHouse(ctx, params)
}
