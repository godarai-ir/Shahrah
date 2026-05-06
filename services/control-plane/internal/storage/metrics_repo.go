package storage

import (
	"context"
	"errors"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

var ErrMetricsStoreUnavailable = errors.New("metrics store unavailable")

func (s *Store) InsertModelMetricEvent(ctx context.Context, ev platform.ModelMetricEvent) error {
	if !s.clickhouseReady || s.clickhouse == nil {
		return ErrMetricsStoreUnavailable
	}
	return s.insertModelMetricEvent(ctx, ev)
}

func (s *Store) QueryModelMetrics(ctx context.Context, tenantID string, req platform.MetricsQueryRequest) ([]map[string]interface{}, error) {
	if !s.clickhouseReady || s.clickhouse == nil {
		return nil, ErrMetricsStoreUnavailable
	}
	return s.queryModelMetrics(ctx, tenantID, req)
}
