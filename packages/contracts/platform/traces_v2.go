package platform

import "time"

type RequestTraceV2 struct {
	TraceID      string    `json:"trace_id"`
	TenantID     string    `json:"tenant_id"`
	RequestID    string    `json:"request_id,omitempty"`
	Model        string    `json:"model,omitempty"`
	Provider     string    `json:"provider,omitempty"`
	Route        string    `json:"route"`
	StatusCode   int       `json:"status_code"`
	StartedAt    time.Time `json:"started_at"`
	DurationMS   int       `json:"duration_ms"`
	ErrorCode    string    `json:"error_code,omitempty"`
	Sampled      bool      `json:"sampled"`
	Attempts     int       `json:"attempts"`
	TotalTokens  int       `json:"total_tokens,omitempty"`
	CostUSD      float64   `json:"cost_usd,omitempty"`
	ErrorsCount  int       `json:"errors_count,omitempty"`
	RequestType  string    `json:"request_type,omitempty"`
	VirtualModel string    `json:"virtual_model,omitempty"`
}

type RequestTraceDetailV2 struct {
	Trace   RequestTraceV2        `json:"trace"`
	Summary RequestTraceSummaryV2 `json:"summary"`
	Usage   *TraceUsageV2         `json:"usage,omitempty"`
	Metric  *TraceMetricV2        `json:"metric,omitempty"`
}

type RequestTraceSummaryV2 struct {
	SpansCount    int `json:"spans_count"`
	AttemptsCount int `json:"attempts_count"`
	ErrorsCount   int `json:"errors_count"`
}

type TraceAttemptV2 struct {
	AttemptNo      int                    `json:"attempt_no"`
	TraceID        string                 `json:"trace_id"`
	Provider       string                 `json:"provider,omitempty"`
	Model          string                 `json:"model,omitempty"`
	StartedAt      time.Time              `json:"started_at"`
	DurationMS     int                    `json:"duration_ms"`
	Status         string                 `json:"status"`
	RetryReason    string                 `json:"retry_reason,omitempty"`
	FallbackReason string                 `json:"fallback_reason,omitempty"`
	Attributes     map[string]interface{} `json:"attributes,omitempty"`
}

type ListRequestTracesV2Params struct {
	TenantID   string
	TraceID    string
	RequestID  string
	Route      string
	Model      string
	Provider   string
	ErrorCode  string
	StatusCode *int
	From       *time.Time
	To         *time.Time
	Limit      int
	Offset     int
	SortBy     string
	SortOrder  string
}

type ListRequestTracesV2Response struct {
	Data       []RequestTraceV2 `json:"data"`
	Total      int              `json:"total"`
	Limit      int              `json:"limit"`
	Offset     int              `json:"offset"`
	NextOffset *int             `json:"next_offset,omitempty"`
}

type TraceTimelineV2Response struct {
	Data []TraceSpan `json:"data"`
}

type TraceAttemptsV2Response struct {
	Data []TraceAttemptV2 `json:"data"`
}

type TraceUsageV2 struct {
	Model          string                 `json:"model,omitempty"`
	InputTokens    int                    `json:"input_tokens,omitempty"`
	OutputTokens   int                    `json:"output_tokens,omitempty"`
	TotalTokens    int                    `json:"total_tokens,omitempty"`
	CostUSD        float64                `json:"cost_usd,omitempty"`
	CostIRR        float64                `json:"cost_irr,omitempty"`
	ResponseTimeMS int                    `json:"response_time_ms,omitempty"`
	Status         string                 `json:"status,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

type TraceMetricV2 struct {
	RequestType      string  `json:"request_type,omitempty"`
	ProviderName     string  `json:"provider_name,omitempty"`
	ProviderModel    string  `json:"provider_model_name,omitempty"`
	ErrorType        string  `json:"error_type,omitempty"`
	LatencyMS        int     `json:"latency_ms,omitempty"`
	InputTokens      int     `json:"input_tokens,omitempty"`
	OutputTokens     int     `json:"output_tokens,omitempty"`
	CostUSD          float64 `json:"cost_usd,omitempty"`
	TTFTMS           *int    `json:"ttft_ms,omitempty"`
	ITLMS            *int    `json:"itl_ms,omitempty"`
	TPOTMS           *int    `json:"tpot_ms,omitempty"`
	VirtualModelName string  `json:"virtual_model_name,omitempty"`
}

type TraceFacetBucketV2 struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}

type TraceStatusFacetBucketV2 struct {
	StatusCode int `json:"status_code"`
	Count      int `json:"count"`
}

type TraceFacetsV2 struct {
	Providers   []TraceFacetBucketV2       `json:"providers"`
	Models      []TraceFacetBucketV2       `json:"models"`
	Routes      []TraceFacetBucketV2       `json:"routes"`
	ErrorCodes  []TraceFacetBucketV2       `json:"error_codes"`
	StatusCodes []TraceStatusFacetBucketV2 `json:"status_codes"`
}

type TraceFacetsV2Response struct {
	Data TraceFacetsV2 `json:"data"`
}
