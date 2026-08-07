// Package tracing provides request tracing, privacy, and export configuration.
package tracing

import (
	"time"
)

const (
	defaultSampleRate   = 0.02
	defaultSlowRequest  = 1500 * time.Millisecond
	defaultQueueCap     = 8192
	defaultBatchSize    = 128
	defaultFlushEvery   = 500 * time.Millisecond
	defaultExportTimout = 2 * time.Second
)

// Config controls tracing sampling, export, privacy, and retention behavior.
type Config struct {
	DefaultSampleRate float64
	SlowRequest       time.Duration
	QueueCapacity     int
	BatchSize         int
	FlushInterval     time.Duration
	ExportTimeout     time.Duration
	OTLPEndpoint      string
	OTLPInsecure      bool
	OTLPServiceName   string
	RetryMaxAttempts  int
	RetryBaseDelay    time.Duration
	RetryMaxDelay     time.Duration
	WALEnabled        bool
	WALPath           string
	WALReplayInterval time.Duration
	CapturePolicy     CapturePolicy
	PayloadBlobDir    string
	PayloadEncryptKey string
	PayloadMaxBytes   int
	PIIHashSalt       string
}

// DefaultConfig returns tracing settings without reading process-global state.
func DefaultConfig() Config {
	return Config{
		DefaultSampleRate: defaultSampleRate,
		SlowRequest:       defaultSlowRequest,
		QueueCapacity:     defaultQueueCap,
		BatchSize:         defaultBatchSize,
		FlushInterval:     defaultFlushEvery,
		ExportTimeout:     defaultExportTimout,
		OTLPEndpoint:      "",
		OTLPInsecure:      false,
		OTLPServiceName:   "shahrah-control-plane",
		RetryMaxAttempts:  3,
		RetryBaseDelay:    100 * time.Millisecond,
		RetryMaxDelay:     2 * time.Second,
		WALEnabled:        true,
		WALPath:           "./tmp/trace-wal.jsonl",
		WALReplayInterval: 3 * time.Second,
		CapturePolicy:     CapturePolicyMetadataOnly,
		PayloadBlobDir:    "./tmp/trace-payload-blobs",
		PayloadEncryptKey: "",
		PayloadMaxBytes:   16 * 1024,
		PIIHashSalt:       "",
	}
}
