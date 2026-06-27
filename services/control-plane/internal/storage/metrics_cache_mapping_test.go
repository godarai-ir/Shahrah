package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheSavingsMetricMappings(t *testing.T) {
	for input, expected := range map[string]string{
		"avoidedCostUSD":      "avoided_cost_usd",
		"avoidedCostIRR":      "avoided_cost_irr",
		"avoidedInputTokens":  "avoided_input_tokens",
		"avoidedOutputTokens": "avoided_output_tokens",
	} {
		actual, err := mapAggregationColumn(input)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	}

	for input, expected := range map[string]string{
		"cacheMode": "cache_mode", "cacheStatus": "cache_status", "cacheMatch": "cache_match",
	} {
		actual, err := mapMetricsGroupBy(input)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	}
}
