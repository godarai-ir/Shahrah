CREATE TABLE IF NOT EXISTS gateway_model_metrics_1m
(
  bucket_ts DateTime('UTC'),
  tenant_id String,

  view_model_name LowCardinality(String),
  view_virtual_model_name LowCardinality(Nullable(String)),
  view_user_email Nullable(String),
  view_virtual_account_name Nullable(String),
  view_team_name Nullable(String),
  view_error_type LowCardinality(String),

  request_count AggregateFunction(count),
  failed_count AggregateFunction(countIf, UInt8),

  input_tokens_sum AggregateFunction(sum, UInt64),
  output_tokens_sum AggregateFunction(sum, UInt64),
  cost_usd_sum AggregateFunction(sum, Decimal(18, 8)),

  latency_p50 AggregateFunction(quantileTDigest(0.50), UInt32),
  latency_p75 AggregateFunction(quantileTDigest(0.75), UInt32),
  latency_p90 AggregateFunction(quantileTDigest(0.90), UInt32),
  latency_p99 AggregateFunction(quantileTDigest(0.99), UInt32),

  ttft_p50 AggregateFunction(quantileTDigest(0.50), UInt32),
  ttft_p75 AggregateFunction(quantileTDigest(0.75), UInt32),
  ttft_p90 AggregateFunction(quantileTDigest(0.90), UInt32),
  ttft_p99 AggregateFunction(quantileTDigest(0.99), UInt32),

  itl_p50 AggregateFunction(quantileTDigest(0.50), UInt32),
  itl_p75 AggregateFunction(quantileTDigest(0.75), UInt32),
  itl_p90 AggregateFunction(quantileTDigest(0.90), UInt32),
  itl_p99 AggregateFunction(quantileTDigest(0.99), UInt32),

  tpot_p50 AggregateFunction(quantileTDigest(0.50), UInt32),
  tpot_p75 AggregateFunction(quantileTDigest(0.75), UInt32),
  tpot_p90 AggregateFunction(quantileTDigest(0.90), UInt32),
  tpot_p99 AggregateFunction(quantileTDigest(0.99), UInt32)
)
ENGINE = ReplicatedAggregatingMergeTree
PARTITION BY toDate(bucket_ts)
ORDER BY (tenant_id, bucket_ts, view_model_name, ifNull(view_virtual_model_name, ''))
TTL toDate(bucket_ts) + INTERVAL 540 DAY
SETTINGS index_granularity = 8192, allow_nullable_key = 1;
