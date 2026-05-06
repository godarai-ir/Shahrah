CREATE TABLE IF NOT EXISTS gateway_model_events
(
  ts DateTime64(3, 'UTC'),
  tenant_id String,
  request_id String,

  model_name LowCardinality(String),
  virtual_model_name Nullable(LowCardinality(String)),
  provider_name LowCardinality(String),
  provider_model_name LowCardinality(String),
  request_type LowCardinality(String),

  user_id Nullable(String),
  user_email Nullable(String),
  virtual_account_id Nullable(String),
  virtual_account_name Nullable(String),

  team_ids Array(String),
  team_names Array(String),

  status_code UInt16,
  error_type LowCardinality(String),

  latency_ms UInt32,
  ttft_ms Nullable(UInt32),
  itl_ms Nullable(UInt32),
  tpot_ms Nullable(UInt32),

  input_tokens UInt32,
  output_tokens UInt32,
  cost_usd Decimal(18, 8),

  metadata_json String
)
ENGINE = MergeTree
PARTITION BY toDate(ts)
ORDER BY (tenant_id, ts, model_name, virtual_model_name)
TTL ts + INTERVAL 90 DAY
SETTINGS index_granularity = 8192;
