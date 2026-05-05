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
	if !s.clickhouseReady || s.clickhouse == nil {
		return ErrTraceStoreUnavailable
	}
	return s.exportToClickHouse(ctx, batch)
}

func (s *Store) ListRequestTraces(ctx context.Context, params platform.ListRequestTracesParams) ([]platform.RequestTrace, int, error) {
	if !s.clickhouseReady || s.clickhouse == nil {
		return nil, 0, ErrTraceStoreUnavailable
	}
	return s.listRequestTracesFromClickHouse(ctx, params)
}

func (s *Store) GetRequestTrace(ctx context.Context, tenantID string, traceID string) (*platform.RequestTrace, error) {
	if !s.clickhouseReady || s.clickhouse == nil {
		return nil, ErrTraceStoreUnavailable
	}
	return s.getRequestTraceFromClickHouse(ctx, tenantID, traceID)
}

func (s *Store) ListTraceSpans(ctx context.Context, tenantID string, traceID string) ([]platform.TraceSpan, error) {
	if !s.clickhouseReady || s.clickhouse == nil {
		return nil, ErrTraceStoreUnavailable
	}
	return s.listTraceSpansFromClickHouse(ctx, tenantID, traceID)
}
