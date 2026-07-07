package tracing

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"
)

type TraceRetentionPolicy struct {
	TenantID      string
	RetentionDays *int
}

func ParseTraceRetentionDays(value interface{}) (*int, bool) {
	switch v := value.(type) {
	case nil:
		return nil, true
	case int:
		return positiveRetentionDays(v)
	case int64:
		return parseTraceRetentionInt64(v)
	case float64:
		return parseTraceRetentionFloat64(v)
	case json.Number:
		return parseTraceRetentionJSONNumber(v)
	case string:
		return parseTraceRetentionString(v)
	default:
		return nil, false
	}
}

func parseTraceRetentionInt64(value int64) (*int, bool) {
	if value > maxIntValue() {
		return nil, false
	}

	return positiveRetentionDays(int(value))
}

func parseTraceRetentionFloat64(value float64) (*int, bool) {
	if value != math.Trunc(value) || value > float64(maxIntValue()) {
		return nil, false
	}

	return positiveRetentionDays(int(value))
}

func parseTraceRetentionJSONNumber(value json.Number) (*int, bool) {
	parsed, err := value.Int64()
	if err != nil {
		return nil, false
	}

	return parseTraceRetentionInt64(parsed)
}

func parseTraceRetentionString(value string) (*int, bool) {
	raw := strings.TrimSpace(strings.ToLower(value))
	if raw == "" || raw == "null" || raw == "unlimited" {
		return nil, true
	}

	parsed, err := strconv.Atoi(raw)
	if err != nil {
		return nil, false
	}

	return positiveRetentionDays(parsed)
}

func positiveRetentionDays(value int) (*int, bool) {
	if value <= 0 {
		return nil, false
	}

	return &value, true
}

func maxIntValue() int64 {
	return int64(^uint(0) >> 1)
}
