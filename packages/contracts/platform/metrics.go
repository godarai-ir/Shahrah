package platform

import "time"

type ModelMetricEvent struct {
	TS                   time.Time              `json:"ts"`
	TenantID             string                 `json:"tenant_id"`
	RequestID            string                 `json:"request_id"`
	ModelName            string                 `json:"model_name"`
	VirtualModelName     string                 `json:"virtual_model_name,omitempty"`
	ProviderName         string                 `json:"provider_name,omitempty"`
	ProviderModelName    string                 `json:"provider_model_name,omitempty"`
	RequestType          string                 `json:"request_type"`
	UserID               string                 `json:"user_id,omitempty"`
	UserEmail            string                 `json:"user_email,omitempty"`
	VirtualAccountID     string                 `json:"virtual_account_id,omitempty"`
	VirtualAccountName   string                 `json:"virtual_account_name,omitempty"`
	TeamIDs              []string               `json:"team_ids,omitempty"`
	TeamNames            []string               `json:"team_names,omitempty"`
	StatusCode           int                    `json:"status_code"`
	ErrorType            string                 `json:"error_type"`
	LatencyMS            int                    `json:"latency_ms"`
	TTFTMS               *int                   `json:"ttft_ms,omitempty"`
	ITLMS                *int                   `json:"itl_ms,omitempty"`
	TPOTMS               *int                   `json:"tpot_ms,omitempty"`
	InputTokens          int                    `json:"input_tokens"`
	OutputTokens         int                    `json:"output_tokens"`
	CostUSD              float64                `json:"cost_usd"`
	CacheMode            string                 `json:"cache_mode"`
	CacheStatus          string                 `json:"cache_status"`
	CacheMatch           string                 `json:"cache_match"`
	CacheSimilarity      *float64               `json:"cache_similarity,omitempty"`
	CacheOriginRequestID string                 `json:"cache_origin_request_id,omitempty"`
	AvoidedInputTokens   int                    `json:"avoided_input_tokens"`
	AvoidedOutputTokens  int                    `json:"avoided_output_tokens"`
	AvoidedCostUSD       float64                `json:"avoided_cost_usd"`
	AvoidedCostIRR       float64                `json:"avoided_cost_irr"`
	EmbeddingCostUSD     float64                `json:"embedding_cost_usd"`
	PricingVersion       string                 `json:"pricing_version,omitempty"`
	Metadata             map[string]interface{} `json:"metadata,omitempty"`
}

type MetricsAggregation struct {
	Type   string `json:"type"`
	Column string `json:"column"`
}

type MetricsFilter struct {
	FieldName   string      `json:"fieldName,omitempty"`
	MetadataKey string      `json:"metadataKey,omitempty"`
	Operator    string      `json:"operator"`
	Value       interface{} `json:"value"`
}

type MetricsQueryRequest struct {
	StartTS           time.Time            `json:"startTs"`
	EndTS             time.Time            `json:"endTs"`
	Datasource        string               `json:"datasource"`
	Type              string               `json:"type"`
	Aggregations      []MetricsAggregation `json:"aggregations,omitempty"`
	GroupBy           []string             `json:"groupBy,omitempty"`
	Filters           []MetricsFilter      `json:"filters,omitempty"`
	IntervalInSeconds int                  `json:"intervalInSeconds,omitempty"`
}

type ModelMetricsFacets struct {
	Models    []string `json:"models"`
	Providers []string `json:"providers"`
}
