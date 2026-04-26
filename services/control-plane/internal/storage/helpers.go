package storage

import (
	"encoding/json"
	"fmt"
)

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	default:
		return fmt.Sprintf("%v", v)
	}
}

func toInt64(v interface{}) int64 {
	switch val := v.(type) {
	case float64:
		return int64(val)
	case int64:
		return val
	case int:
		return int64(val)
	case json.Number:
		out, _ := val.Int64()
		return out
	default:
		return 0
	}
}

func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int64:
		return float64(val)
	case int:
		return float64(val)
	case json.Number:
		out, _ := val.Float64()
		return out
	default:
		return 0
	}
}

func emptyToNil(v string) interface{} {
	if v == "" {
		return nil
	}
	return v
}

func nullUUID(v string) interface{} {
	if v == "" {
		return nil
	}
	return v
}
