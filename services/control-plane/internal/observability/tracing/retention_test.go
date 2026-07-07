package tracing

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTraceRetentionDays(t *testing.T) {
	t.Run("positive integers are limited", func(t *testing.T) {
		days, ok := ParseTraceRetentionDays(30)
		require.True(t, ok)
		require.NotNil(t, days)
		assert.Equal(t, 30, *days)
	})

	t.Run("json numbers are supported", func(t *testing.T) {
		days, ok := ParseTraceRetentionDays(json.Number("90"))
		require.True(t, ok)
		require.NotNil(t, days)
		assert.Equal(t, 90, *days)
	})

	t.Run("nil and unlimited strings are unlimited", func(t *testing.T) {
		for _, value := range []interface{}{nil, "", "null", " unlimited "} {
			days, ok := ParseTraceRetentionDays(value)
			assert.True(t, ok, "value=%v", value)
			assert.Nil(t, days, "value=%v", value)
		}
	})

	t.Run("invalid and non-positive values are ignored", func(t *testing.T) {
		for _, value := range []interface{}{0, -1, "0", "-3", "abc", 1.5, map[string]interface{}{"days": 7}} {
			days, ok := ParseTraceRetentionDays(value)
			assert.False(t, ok, "value=%v", value)
			assert.Nil(t, days, "value=%v", value)
		}
	})
}
