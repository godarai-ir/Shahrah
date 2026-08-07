// Package observability owns control-plane observability application contracts.
package observability

import (
	"errors"
	"strings"
)

var (
	// ErrTraceStoreUnavailable indicates that trace persistence is unavailable.
	ErrTraceStoreUnavailable = errors.New("trace store unavailable")
	// ErrLogStoreUnavailable indicates that request-log persistence is unavailable.
	ErrLogStoreUnavailable = errors.New("log store unavailable")
	// ErrMetricsStoreUnavailable indicates that metrics persistence is unavailable.
	ErrMetricsStoreUnavailable = errors.New("metrics store unavailable")
)

const (
	// RequestLoggingModeNone disables request logging.
	RequestLoggingModeNone = "none"
	// RequestLoggingModeMetadataOnly records request metadata only.
	RequestLoggingModeMetadataOnly = "metadata-only"
	// RequestLoggingModeRedactedContent records redacted request content.
	RequestLoggingModeRedactedContent = "redacted-content"
	// RequestLoggingModeFullContent records full request content.
	RequestLoggingModeFullContent = "full-content"
	// LazyJSONPayloadField carries undecoded JSON to the asynchronous log worker.
	LazyJSONPayloadField = "__shahrah_lazy_json_payload"
	// LazyJSONFallbackField carries a fallback used when lazy JSON is invalid.
	LazyJSONFallbackField = "__shahrah_lazy_json_fallback"
)

// NormalizeRequestLoggingMode normalizes a request-logging mode to a supported value.
func NormalizeRequestLoggingMode(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case RequestLoggingModeNone:
		return RequestLoggingModeNone
	case RequestLoggingModeRedactedContent:
		return RequestLoggingModeRedactedContent
	case RequestLoggingModeFullContent:
		return RequestLoggingModeFullContent
	default:
		return RequestLoggingModeMetadataOnly
	}
}
