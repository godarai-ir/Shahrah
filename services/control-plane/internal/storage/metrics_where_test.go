package storage

import (
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildMetricsWhere_TeamArrayFiltersUseClickHouseArrayFunctions(t *testing.T) {
	baseReq := platform.MetricsQueryRequest{
		StartTS: time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC),
		EndTS:   time.Date(2026, 7, 8, 11, 0, 0, 0, time.UTC),
	}

	tests := []struct {
		name     string
		filter   platform.MetricsFilter
		contains string
		arg      interface{}
	}{
		{
			name:     "equal",
			filter:   platform.MetricsFilter{FieldName: "team", Operator: "EQUAL", Value: "platform"},
			contains: "has(team_names, ?)",
			arg:      "platform",
		},
		{
			name:     "in",
			filter:   platform.MetricsFilter{FieldName: "team", Operator: "IN", Value: []interface{}{"platform", "growth"}},
			contains: "hasAny(team_names, ?)",
			arg:      []string{"platform", "growth"},
		},
		{
			name:     "not in",
			filter:   platform.MetricsFilter{FieldName: "team", Operator: "NOT_IN", Value: []interface{}{"platform", "growth"}},
			contains: "NOT hasAny(team_names, ?)",
			arg:      []string{"platform", "growth"},
		},
		{
			name:     "is null",
			filter:   platform.MetricsFilter{FieldName: "team", Operator: "IS_NULL", Value: true},
			contains: "empty(team_names)",
		},
		{
			name:     "is not null",
			filter:   platform.MetricsFilter{FieldName: "team", Operator: "IS_NULL", Value: false},
			contains: "NOT empty(team_names)",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := baseReq
			req.Filters = []platform.MetricsFilter{tc.filter}

			where, args, err := buildMetricsWhere("tenant-1", req)
			require.NoError(t, err)
			assert.Contains(t, where, tc.contains)

			if tc.arg != nil {
				require.Len(t, args, 4)
				assert.Equal(t, tc.arg, args[3])
			} else {
				assert.Len(t, args, 3)
			}
		})
	}
}

func TestBuildMetricsWhere_ScalarFiltersStayScalar(t *testing.T) {
	req := platform.MetricsQueryRequest{
		StartTS: time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC),
		EndTS:   time.Date(2026, 7, 8, 11, 0, 0, 0, time.UTC),
		Filters: []platform.MetricsFilter{
			{FieldName: "modelName", Operator: "IN", Value: []interface{}{"gpt-4.1", "gpt-4o"}},
		},
	}

	where, args, err := buildMetricsWhere("tenant-1", req)
	require.NoError(t, err)

	assert.Contains(t, where, "model_name IN (?,?)")
	assert.Equal(t, []interface{}{"tenant-1", req.StartTS.UTC(), req.EndTS.UTC(), "gpt-4.1", "gpt-4o"}, args)
}

func TestBuildMetricsWhere_TeamRejectsNumericComparisons(t *testing.T) {
	req := platform.MetricsQueryRequest{
		StartTS: time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC),
		EndTS:   time.Date(2026, 7, 8, 11, 0, 0, 0, time.UTC),
		Filters: []platform.MetricsFilter{
			{FieldName: "team", Operator: "GREATER_THAN", Value: "platform"},
		},
	}

	_, _, err := buildMetricsWhere("tenant-1", req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported array filter operator")
}
