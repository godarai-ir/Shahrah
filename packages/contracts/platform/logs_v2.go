package platform

import "time"

type RequestLogV2 struct {
	LogID             string    `json:"log_id"`
	RequestID         string    `json:"request_id"`
	TraceID           string    `json:"trace_id,omitempty"`
	TenantID          string    `json:"tenant_id"`
	Provider          string    `json:"provider,omitempty"`
	ProviderModel     string    `json:"provider_model,omitempty"`
	Model             string    `json:"model,omitempty"`
	VirtualModel      string    `json:"virtual_model,omitempty"`
	Route             string    `json:"route,omitempty"`
	RequestType       string    `json:"request_type,omitempty"`
	Status            string    `json:"status,omitempty"`
	StatusCode        int       `json:"status_code"`
	ErrorType         string    `json:"error_type,omitempty"`
	StartedAt         time.Time `json:"started_at"`
	DurationMS        int       `json:"duration_ms"`
	InputTokens       int       `json:"input_tokens,omitempty"`
	OutputTokens      int       `json:"output_tokens,omitempty"`
	TotalTokens       int       `json:"total_tokens,omitempty"`
	CostUSD           float64   `json:"cost_usd,omitempty"`
	CostIRR           float64   `json:"cost_irr,omitempty"`
	CacheStatus       string    `json:"cache_status,omitempty"`
	RetryCount        int       `json:"retry_count,omitempty"`
	FallbackCount     int       `json:"fallback_count,omitempty"`
	FallbackStatus    string    `json:"fallback_status,omitempty"`
	LoadBalanceStatus string    `json:"load_balance_status,omitempty"`
	PayloadMode       string    `json:"payload_mode,omitempty"`
}

type RequestLogContentVisibilityV2 struct {
	CanViewContent    bool   `json:"can_view_content"`
	RequestReason     string `json:"request_reason,omitempty"`
	ResponseReason    string `json:"response_reason,omitempty"`
	RequestAvailable  bool   `json:"request_available"`
	ResponseAvailable bool   `json:"response_available"`
}

type RequestLogDetailV2 struct {
	Log                 RequestLogV2                  `json:"log"`
	ContentVisibility   RequestLogContentVisibilityV2 `json:"content_visibility"`
	RequestPayload      map[string]interface{}        `json:"request_payload,omitempty"`
	ResponsePayload     map[string]interface{}        `json:"response_payload,omitempty"`
	RequestPayloadMeta  map[string]interface{}        `json:"request_payload_meta,omitempty"`
	ResponsePayloadMeta map[string]interface{}        `json:"response_payload_meta,omitempty"`
	Metadata            map[string]interface{}        `json:"metadata,omitempty"`
	Actor               map[string]string             `json:"actor,omitempty"`
	Scope               map[string]string             `json:"scope,omitempty"`
}

type ListRequestLogsV2Params struct {
	TenantID       string
	LogID          string
	RequestID      string
	TraceID        string
	Provider       string
	Model          string
	VirtualModel   string
	Route          string
	RequestType    string
	ErrorType      string
	CacheStatus    string
	FallbackStatus string
	StatusCode     *int
	From           *time.Time
	To             *time.Time
	Limit          int
	Offset         int
	SortBy         string
	SortOrder      string
}

type ListRequestLogsV2Response struct {
	Data       []RequestLogV2 `json:"data"`
	Total      int            `json:"total"`
	Limit      int            `json:"limit"`
	Offset     int            `json:"offset"`
	NextOffset *int           `json:"next_offset,omitempty"`
}

type LogFacetBucketV2 struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}

type LogStatusFacetBucketV2 struct {
	StatusCode int `json:"status_code"`
	Count      int `json:"count"`
}

type LogFacetsV2 struct {
	Providers        []LogFacetBucketV2       `json:"providers"`
	Models           []LogFacetBucketV2       `json:"models"`
	VirtualModels    []LogFacetBucketV2       `json:"virtual_models"`
	Routes           []LogFacetBucketV2       `json:"routes"`
	RequestTypes     []LogFacetBucketV2       `json:"request_types"`
	ErrorTypes       []LogFacetBucketV2       `json:"error_types"`
	CacheStatuses    []LogFacetBucketV2       `json:"cache_statuses"`
	FallbackStatuses []LogFacetBucketV2       `json:"fallback_statuses"`
	StatusCodes      []LogStatusFacetBucketV2 `json:"status_codes"`
}

type LogFacetsV2Response struct {
	Data LogFacetsV2 `json:"data"`
}
