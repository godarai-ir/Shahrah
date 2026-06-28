package storage

import (
	"context"
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
		out, err := val.Int64()
		if err != nil {
			return 0
		}

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
		out, err := val.Float64()
		if err != nil {
			return 0
		}

		return out
	default:
		return 0
	}
}

func marshalJSONBestEffort(v interface{}) []byte {
	raw, err := json.Marshal(v)
	if err != nil {
		return []byte("null")
	}

	return raw
}

func rollbackTxBestEffort(ctx context.Context, tx interface{ Rollback(context.Context) error }) {
	if err := tx.Rollback(ctx); err != nil {
		return
	}
}

func closeBestEffort(closer interface{ Close() error }) {
	if err := closer.Close(); err != nil {
		return
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
