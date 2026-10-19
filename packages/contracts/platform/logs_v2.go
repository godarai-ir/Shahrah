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
	TenantID          string
	LogID             string
	RequestID         string
	TraceID           string
	Provider          string
	Model             string
	VirtualModel      string
	Route             string
	RequestType       string
	ErrorType         string
	CacheStatus       string
	FallbackStatus    string
	LoadBalanceStatus string
	PayloadMode       string
	ActorType         string
	ActorID           string
	TeamID            string
	StatusCode        *int
	From              *time.Time
	To                *time.Time
	DurationMin       *int
	DurationMax       *int
	TokensMin         *int
	TokensMax         *int
	CostUSDMin        *float64
	CostUSDMax        *float64
	CostIRRMin        *float64
	CostIRRMax        *float64
	Limit             int
	Offset            int
	SortBy            string
	SortOrder         string
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
	Providers           []LogFacetBucketV2       `json:"providers"`
	Models              []LogFacetBucketV2       `json:"models"`
	VirtualModels       []LogFacetBucketV2       `json:"virtual_models"`
	Routes              []LogFacetBucketV2       `json:"routes"`
	RequestTypes        []LogFacetBucketV2       `json:"request_types"`
	ErrorTypes          []LogFacetBucketV2       `json:"error_types"`
	CacheStatuses       []LogFacetBucketV2       `json:"cache_statuses"`
	FallbackStatuses    []LogFacetBucketV2       `json:"fallback_statuses"`
	LoadBalanceStatuses []LogFacetBucketV2       `json:"load_balance_statuses"`
	PayloadModes        []LogFacetBucketV2       `json:"payload_modes"`
	StatusCodes         []LogStatusFacetBucketV2 `json:"status_codes"`
}

type LogFacetsV2Response struct {
	Data LogFacetsV2 `json:"data"`
}

type LogSummaryV2 struct {
	RequestCount     int                      `json:"request_count"`
	ErrorCount       int                      `json:"error_count"`
	ErrorRate        float64                  `json:"error_rate"`
	LatencyAvgMS     float64                  `json:"latency_avg_ms"`
	LatencyP50MS     float64                  `json:"latency_p50_ms"`
	LatencyP95MS     float64                  `json:"latency_p95_ms"`
	LatencyP99MS     float64                  `json:"latency_p99_ms"`
	InputTokens      int                      `json:"input_tokens"`
	OutputTokens     int                      `json:"output_tokens"`
	TotalTokens      int                      `json:"total_tokens"`
	CostUSD          float64                  `json:"cost_usd"`
	CostIRR          float64                  `json:"cost_irr"`
	CacheHitCount    int                      `json:"cache_hit_count"`
	CacheHitRate     float64                  `json:"cache_hit_rate"`
	RetryCount       int                      `json:"retry_count"`
	FallbackCount    int                      `json:"fallback_count"`
	LoadBalanceCount int                      `json:"load_balance_count"`
	PayloadModes     []LogFacetBucketV2       `json:"payload_modes"`
	StatusCodes      []LogStatusFacetBucketV2 `json:"status_codes"`
	Providers        []LogFacetBucketV2       `json:"providers"`
	Models           []LogFacetBucketV2       `json:"models"`
	Routes           []LogFacetBucketV2       `json:"routes"`
}

type LogSummaryV2Response struct {
	Data LogSummaryV2 `json:"data"`
}

type LogRangeV2 struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

type LogRangeStatsV2 struct {
	DurationMS  LogRangeV2 `json:"duration_ms"`
	TotalTokens LogRangeV2 `json:"total_tokens"`
	CostUSD     LogRangeV2 `json:"cost_usd"`
	CostIRR     LogRangeV2 `json:"cost_irr"`
}

type LogRangeStatsV2Response struct {
	Data LogRangeStatsV2 `json:"data"`
}

type LogAnalyticsV2Params struct {
	ListRequestLogsV2Params
	Type            string
	GroupBy         string
	IntervalSeconds int
}

type LogAnalyticsPointV2 struct {
	Timestamp        *time.Time `json:"timestamp,omitempty"`
	Group            string     `json:"group,omitempty"`
	RequestCount     int        `json:"request_count"`
	ErrorCount       int        `json:"error_count"`
	ErrorRate        float64    `json:"error_rate"`
	LatencyP95MS     float64    `json:"latency_p95_ms"`
	TotalTokens      int        `json:"total_tokens"`
	CostUSD          float64    `json:"cost_usd"`
	CacheHitCount    int        `json:"cache_hit_count"`
	RetryCount       int        `json:"retry_count"`
	FallbackCount    int        `json:"fallback_count"`
	LoadBalanceCount int        `json:"load_balance_count"`
}

type LogAnalyticsV2Response struct {
	Data []LogAnalyticsPointV2 `json:"data"`
}

type SavedLogViewV2 struct {
	SavedViewID     string                 `json:"saved_view_id"`
	TenantID        string                 `json:"tenant_id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description,omitempty"`
	Filters         map[string]interface{} `json:"filters"`
	VisibleColumns  []string               `json:"visible_columns"`
	SharedScope     string                 `json:"shared_scope"`
	CreatedByUserID string                 `json:"created_by_user_id,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

type CreateSavedLogViewV2Request struct {
	Name           string                 `json:"name"`
	Description    string                 `json:"description,omitempty"`
	Filters        map[string]interface{} `json:"filters"`
	VisibleColumns []string               `json:"visible_columns"`
	SharedScope    string                 `json:"shared_scope,omitempty"`
}

type UpdateSavedLogViewV2Request struct {
	Name           *string                `json:"name,omitempty"`
	Description    *string                `json:"description,omitempty"`
	Filters        map[string]interface{} `json:"filters,omitempty"`
	VisibleColumns []string               `json:"visible_columns,omitempty"`
	SharedScope    *string                `json:"shared_scope,omitempty"`
}

type ListSavedLogViewsV2Response struct {
	Data []SavedLogViewV2 `json:"data"`
}

type LogExportJobV2 struct {
	ExportID        string                 `json:"export_id"`
	TenantID        string                 `json:"tenant_id"`
	Filters         map[string]interface{} `json:"filters"`
	Format          string                 `json:"format"`
	IncludeContent  bool                   `json:"include_content"`
	Status          string                 `json:"status"`
	RowCount        int                    `json:"row_count"`
	CreatedByUserID string                 `json:"created_by_user_id,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	CompletedAt     *time.Time             `json:"completed_at,omitempty"`
}

type CreateLogExportV2Request struct {
	Filters        map[string]interface{} `json:"filters"`
	Format         string                 `json:"format"`
	IncludeContent bool                   `json:"include_content"`
}

type ListLogExportsV2Response struct {
	Data []LogExportJobV2 `json:"data"`
}

type LogReplayDraftV2 struct {
	LogID          string                 `json:"log_id"`
	Supported      bool                   `json:"supported"`
	Reason         string                 `json:"reason,omitempty"`
	Route          string                 `json:"route,omitempty"`
	RequestType    string                 `json:"request_type,omitempty"`
	Model          string                 `json:"model,omitempty"`
	Provider       string                 `json:"provider,omitempty"`
	RequestPayload map[string]interface{} `json:"request_payload,omitempty"`
	SafeHeaders    map[string]string      `json:"safe_headers,omitempty"`
	Warnings       []string               `json:"warnings,omitempty"`
}
