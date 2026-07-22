ALTER TABLE gateway_model_events ADD COLUMN IF NOT EXISTS cache_mode LowCardinality(String) DEFAULT 'none';
ALTER TABLE gateway_model_events ADD COLUMN IF NOT EXISTS cache_status LowCardinality(String) DEFAULT 'none';
ALTER TABLE gateway_model_events ADD COLUMN IF NOT EXISTS cache_match LowCardinality(String) DEFAULT 'none';
ALTER TABLE gateway_model_events ADD COLUMN IF NOT EXISTS cache_similarity Nullable(Float32);
ALTER TABLE gateway_model_events ADD COLUMN IF NOT EXISTS cache_origin_request_id Nullable(String);
ALTER TABLE gateway_model_events ADD COLUMN IF NOT EXISTS avoided_input_tokens UInt32 DEFAULT 0;
ALTER TABLE gateway_model_events ADD COLUMN IF NOT EXISTS avoided_output_tokens UInt32 DEFAULT 0;
ALTER TABLE gateway_model_events ADD COLUMN IF NOT EXISTS avoided_cost_usd Decimal(18, 8) DEFAULT 0;
ALTER TABLE gateway_model_events ADD COLUMN IF NOT EXISTS avoided_cost_irr Decimal(24, 2) DEFAULT 0;
ALTER TABLE gateway_model_events ADD COLUMN IF NOT EXISTS embedding_cost_usd Decimal(18, 8) DEFAULT 0;
ALTER TABLE gateway_model_events ADD COLUMN IF NOT EXISTS pricing_version String DEFAULT '';

CREATE TABLE IF NOT EXISTS request_traces_ch
(
  trace_id String,
  tenant_id String,
  request_id String,
  model String,
  route String,
  status_code Int32,
  started_at DateTime64(3, 'UTC'),
  duration_ms Int32,
  error_code String,
  sampled UInt8,
  retention_days UInt16 DEFAULT 0
)
ENGINE = ReplicatedMergeTree
PARTITION BY toDate(started_at)
ORDER BY (tenant_id, started_at, trace_id);

CREATE TABLE IF NOT EXISTS request_spans_ch
(
  span_id String,
  trace_id String,
  tenant_id String,
  parent_span_id String,
  name String,
  kind String,
  started_at DateTime64(3, 'UTC'),
  duration_ms Int32,
  status String,
  attributes_json String,
  retention_days UInt16 DEFAULT 0
)
ENGINE = ReplicatedMergeTree
PARTITION BY toDate(started_at)
ORDER BY (tenant_id, trace_id, started_at, span_id);

ALTER TABLE request_traces_ch ADD COLUMN IF NOT EXISTS retention_days UInt16 DEFAULT 0;
ALTER TABLE request_spans_ch ADD COLUMN IF NOT EXISTS retention_days UInt16 DEFAULT 0;
ALTER TABLE request_traces_ch MODIFY TTL toDateTime(started_at) + toIntervalDay(retention_days) DELETE WHERE retention_days > 0;
ALTER TABLE request_spans_ch MODIFY TTL toDateTime(started_at) + toIntervalDay(retention_days) DELETE WHERE retention_days > 0;

CREATE TABLE IF NOT EXISTS request_logs_ch
(
  log_id String,
  tenant_id String,
  request_id String,
  trace_id String,
  token_hash String,
  organization_id String,
  org_unit_id String,
  team_id String,
  actor_type LowCardinality(String),
  actor_id String,
  provider String,
  provider_model String,
  model String,
  virtual_model String,
  route String,
  request_type LowCardinality(String),
  status LowCardinality(String),
  status_code Int32,
  error_type LowCardinality(String),
  started_at DateTime64(3, 'UTC'),
  duration_ms Int32,
  input_tokens UInt32,
  output_tokens UInt32,
  total_tokens UInt32,
  cost_usd Decimal(18, 6),
  cost_irr Decimal(18, 2),
  cache_status LowCardinality(String),
  cache_mode LowCardinality(String),
  retry_count UInt16,
  fallback_count UInt16,
  fallback_status LowCardinality(String),
  load_balance_status LowCardinality(String),
  payload_mode LowCardinality(String),
  request_payload_json String,
  response_payload_json String,
  request_payload_meta_json String,
  response_payload_meta_json String,
  request_payload_truncated UInt8,
  response_payload_truncated UInt8,
  metadata_json String
)
ENGINE = ReplicatedMergeTree
PARTITION BY toDate(started_at)
ORDER BY (tenant_id, started_at, log_id);
