package storage

import (
	"context"
	"errors"

	clickhouse "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/Delpak/shahrah/packages/contracts/platform"
)

var ErrMetricsStoreUnavailable = errors.New("metrics store unavailable")

func (s *ObservabilityRepository) InsertModelMetricEvent(ctx context.Context, ev platform.ModelMetricEvent) error {
	return s.InsertModelMetricEvents(ctx, []platform.ModelMetricEvent{ev})
}

func (s *ObservabilityRepository) InsertModelMetricEvents(ctx context.Context, events []platform.ModelMetricEvent) error {
	if !s.clickHouseMetricsReady() {
		return ErrMetricsStoreUnavailable
	}

	return s.insertModelMetricEvents(ctx, events)
}

// InsertModelMetricEventWithDedupToken writes one replayed event idempotently.
// ClickHouse uses this token to ignore retries of the same insert block.
func (s *ObservabilityRepository) InsertModelMetricEventWithDedupToken(ctx context.Context, ev platform.ModelMetricEvent, token string) error {
	if !s.clickHouseMetricsReady() {
		return ErrMetricsStoreUnavailable
	}

	if token == "" {
		return errors.New("clickhouse insert deduplication token is required")
	}

	ctx = clickhouse.Context(ctx, clickhouse.WithSettings(clickhouse.Settings{
		"insert_deduplication_token": token,
	}))

	return s.insertModelMetricEvents(ctx, []platform.ModelMetricEvent{ev})
}

func (s *ObservabilityRepository) QueryModelMetrics(ctx context.Context, tenantID string, req platform.MetricsQueryRequest) ([]map[string]interface{}, error) {
	if !s.clickHouseMetricsReady() {
		return nil, ErrMetricsStoreUnavailable
	}

	return s.queryModelMetrics(ctx, tenantID, req)
}

func (s *ObservabilityRepository) ListModelMetricsFacets(ctx context.Context, tenantID string) (*platform.ModelMetricsFacets, error) {
	if !s.clickHouseMetricsReady() {
		return nil, ErrMetricsStoreUnavailable
	}

	return s.listModelMetricsFacets(ctx, tenantID)
}
