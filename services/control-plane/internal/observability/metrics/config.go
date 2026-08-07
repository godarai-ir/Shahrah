// Package metrics configures model-metrics collection and export.
package metrics

import (
	"time"
)

// Config controls the metrics queue, flush, and retry behavior.
type Config struct {
	Enabled          bool
	QueueCapacity    int
	BatchSize        int
	FlushInterval    time.Duration
	ExportTimeout    time.Duration
	RetryMaxAttempts int
	RetryBaseDelay   time.Duration
	RetryMaxDelay    time.Duration
	DLQ              DLQConfig
}

// DefaultConfig returns metrics settings without reading process-global state.
func DefaultConfig() Config {
	return Config{
		Enabled:          false,
		QueueCapacity:    8192,
		BatchSize:        128,
		FlushInterval:    500 * time.Millisecond,
		ExportTimeout:    2 * time.Second,
		RetryMaxAttempts: 3,
		RetryBaseDelay:   100 * time.Millisecond,
		RetryMaxDelay:    2 * time.Second,
		DLQ:              DefaultDLQConfig(),
	}
}
