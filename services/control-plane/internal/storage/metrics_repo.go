package storage

import (
	"context"
	"errors"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

var ErrMetricsStoreUnavailable = errors.New("metrics store unavailable")

func (s *Store) InsertModelMetricEvent(ctx context.Context, ev platform.ModelMetricEvent) error {
	return s.InsertModelMetricEvents(ctx, []platform.ModelMetricEvent{ev})
}

func (s *Store) InsertModelMetricEvents(ctx context.Context, events []platform.ModelMetricEvent) error {
	if !s.clickHouseMetricsReady() {
		return ErrMetricsStoreUnavailable
	}

	return s.insertModelMetricEvents(ctx, events)
}

func (s *Store) QueryModelMetrics(ctx context.Context, tenantID string, req platform.MetricsQueryRequest) ([]map[string]interface{}, error) {
	if !s.clickHouseMetricsReady() {
		return nil, ErrMetricsStoreUnavailable
	}

	return s.queryModelMetrics(ctx, tenantID, req)
}

func (s *Store) ListModelMetricsFacets(ctx context.Context, tenantID string) (*platform.ModelMetricsFacets, error) {
	if !s.clickHouseMetricsReady() {
		return nil, ErrMetricsStoreUnavailable
	}

	return s.listModelMetricsFacets(ctx, tenantID)
}
