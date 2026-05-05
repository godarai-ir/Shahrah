package platform

import "time"

type RequestTrace struct {
	TraceID    string    `json:"trace_id"`
	TenantID   string    `json:"tenant_id"`
	RequestID  string    `json:"request_id,omitempty"`
	Model      string    `json:"model,omitempty"`
	Route      string    `json:"route"`
	StatusCode int       `json:"status_code"`
	StartedAt  time.Time `json:"started_at"`
	DurationMS int       `json:"duration_ms"`
	ErrorCode  string    `json:"error_code,omitempty"`
	Sampled    bool      `json:"sampled"`
}

type TraceSpan struct {
	SpanID       string                 `json:"span_id"`
	TraceID      string                 `json:"trace_id"`
	ParentSpanID string                 `json:"parent_span_id,omitempty"`
	Name         string                 `json:"name"`
	Kind         string                 `json:"kind"`
	StartedAt    time.Time              `json:"started_at"`
	DurationMS   int                    `json:"duration_ms"`
	Status       string                 `json:"status"`
	Attributes   map[string]interface{} `json:"attributes"`
}

type ListRequestTracesResponse struct {
	Data       []RequestTrace `json:"data"`
	Total      int            `json:"total"`
	Limit      int            `json:"limit"`
	Offset     int            `json:"offset"`
	NextOffset *int           `json:"next_offset,omitempty"`
}

type TraceSpansResponse struct {
	Data []TraceSpan `json:"data"`
}

type ListRequestTracesParams struct {
	TenantID   string
	TraceID    string
	Route      string
	Model      string
	StatusCode *int
	From       *time.Time
	To         *time.Time
	Limit      int
	Offset     int
}
