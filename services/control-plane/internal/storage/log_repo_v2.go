package storage

import (
	"context"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability"
)

// ErrLogStoreUnavailable is retained for storage-adapter compatibility.
var ErrLogStoreUnavailable = observability.ErrLogStoreUnavailable

func (s *RequestLogRepository) ListRequestLogsV2(ctx context.Context, params platform.ListRequestLogsV2Params) ([]platform.RequestLogV2, int, error) {
	if !s.clickHouseLogsReady() {
		return nil, 0, ErrLogStoreUnavailable
	}

	return s.listRequestLogsV2FromClickHouse(ctx, params)
}

func (s *RequestLogRepository) GetRequestLogV2(ctx context.Context, tenantID, logID string, includeContent bool) (*platform.RequestLogDetailV2, error) {
	if !s.clickHouseLogsReady() {
		return nil, ErrLogStoreUnavailable
	}

	return s.getRequestLogV2FromClickHouse(ctx, tenantID, logID, includeContent)
}

func (s *RequestLogRepository) GetRequestLogFacetsV2(ctx context.Context, params platform.ListRequestLogsV2Params) (*platform.LogFacetsV2, error) {
	if !s.clickHouseLogsReady() {
		return nil, ErrLogStoreUnavailable
	}

	return s.getRequestLogFacetsV2FromClickHouse(ctx, params)
}

func (s *RequestLogRepository) GetRequestLogSummaryV2(ctx context.Context, params platform.ListRequestLogsV2Params) (*platform.LogSummaryV2, error) {
	if !s.clickHouseLogsReady() {
		return nil, ErrLogStoreUnavailable
	}

	return s.getRequestLogSummaryV2FromClickHouse(ctx, params)
}

func (s *RequestLogRepository) GetRequestLogRangeStatsV2(ctx context.Context, params platform.ListRequestLogsV2Params) (*platform.LogRangeStatsV2, error) {
	if !s.clickHouseLogsReady() {
		return nil, ErrLogStoreUnavailable
	}

	return s.getRequestLogRangeStatsV2FromClickHouse(ctx, params)
}

func (s *RequestLogRepository) GetRequestLogAnalyticsV2(ctx context.Context, params platform.LogAnalyticsV2Params) ([]platform.LogAnalyticsPointV2, error) {
	if !s.clickHouseLogsReady() {
		return nil, ErrLogStoreUnavailable
	}

	return s.getRequestLogAnalyticsV2FromClickHouse(ctx, params)
}
