package storage

import (
	"strings"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildModelMetricsAggregateQuery_TimeseriesUsesAggregateAndRawBoundaries(t *testing.T) {
	req := platform.MetricsQueryRequest{
		StartTS:           time.Date(2026, 7, 8, 10, 0, 15, 0, time.UTC),
		EndTS:             time.Date(2026, 7, 8, 11, 0, 30, 0, time.UTC),
		Type:              "timeseries",
		GroupBy:           []string{"modelName"},
		IntervalInSeconds: 300,
		Aggregations: []platform.MetricsAggregation{
			{Type: "count", Column: "modelName"},
			{Type: "sum", Column: "inputTokens"},
			{Type: "p99", Column: "latencyMs"},
		},
		Filters: []platform.MetricsFilter{
			{FieldName: "modelName", Operator: "STRING_CONTAINS", Value: "gpt"},
		},
	}

	query, args, ok, err := buildModelMetricsAggregateQuery("tenant-1", req, true)
	require.NoError(t, err)
	require.True(t, ok)

	assert.Contains(t, query, "FROM gateway_model_metrics_1m")
	assert.Contains(t, query, "UNION ALL")
	assert.Contains(t, query, "FROM gateway_model_events")
	assert.Contains(t, query, "countMerge(__countModelName_state) AS countModelName")
	assert.Contains(t, query, "sumMerge(__sumInputTokens_state) AS sumInputTokens")
	assert.Contains(t, query, "quantileTDigestMerge(0.99)(__p99LatencyMs_state) AS p99LatencyMs")
	assert.Contains(t, query, "ORDER BY timestamp ASC LIMIT 5000")
	assert.Len(t, args, 10)
	assert.Equal(t, time.Date(2026, 7, 8, 10, 1, 0, 0, time.UTC), args[1])
	assert.Equal(t, time.Date(2026, 7, 8, 11, 0, 0, 0, time.UTC), args[2])
}

func TestBuildModelMetricsAggregateQuery_DistributionWithoutGroupUsesAggregate(t *testing.T) {
	req := platform.MetricsQueryRequest{
		StartTS: time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC),
		EndTS:   time.Date(2026, 7, 8, 11, 0, 0, 0, time.UTC),
		Type:    "distribution",
	}

	query, args, ok, err := buildModelMetricsAggregateQuery("tenant-1", req, false)
	require.NoError(t, err)
	require.True(t, ok)

	assert.Contains(t, query, "FROM gateway_model_metrics_1m")
	assert.Contains(t, query, "countMerge(__countAll_state) AS countAll")
	assert.NotContains(t, query, "GROUP BY")
	assert.Len(t, args, 8)
}

func TestBuildModelMetricsAggregateQuery_FallsBackForUnsupportedAggregateShape(t *testing.T) {
	baseReq := platform.MetricsQueryRequest{
		StartTS: time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC),
		EndTS:   time.Date(2026, 7, 8, 11, 0, 0, 0, time.UTC),
		Type:    "timeseries",
	}

	tests := []platform.MetricsQueryRequest{
		{
			StartTS:      baseReq.StartTS,
			EndTS:        baseReq.EndTS,
			Type:         baseReq.Type,
			GroupBy:      []string{"cacheMode"},
			Aggregations: []platform.MetricsAggregation{{Type: "sum", Column: "costInUSD"}},
		},
		{
			StartTS:      baseReq.StartTS,
			EndTS:        baseReq.EndTS,
			Type:         baseReq.Type,
			GroupBy:      []string{"modelName"},
			Aggregations: []platform.MetricsAggregation{{Type: "sum", Column: "avoidedCostUSD"}},
		},
		{
			StartTS:      baseReq.StartTS,
			EndTS:        baseReq.EndTS,
			Type:         baseReq.Type,
			GroupBy:      []string{"modelName"},
			Aggregations: []platform.MetricsAggregation{{Type: "sum", Column: "costInUSD"}},
			Filters:      []platform.MetricsFilter{{FieldName: "providerName", Operator: "EQUAL", Value: "openai"}},
		},
		{
			StartTS:           baseReq.StartTS,
			EndTS:             baseReq.EndTS,
			Type:              baseReq.Type,
			GroupBy:           []string{"modelName"},
			IntervalInSeconds: 90,
			Aggregations:      []platform.MetricsAggregation{{Type: "sum", Column: "costInUSD"}},
		},
	}

	for _, req := range tests {
		query, args, ok, err := buildModelMetricsAggregateQuery("tenant-1", req, true)
		require.NoError(t, err)
		assert.False(t, ok)
		assert.Empty(t, strings.TrimSpace(query))
		assert.Nil(t, args)
	}
}
