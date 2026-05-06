CREATE MATERIALIZED VIEW IF NOT EXISTS mv_gateway_model_metrics_1m
TO gateway_model_metrics_1m
AS
SELECT
  toStartOfMinute(ts) AS bucket_ts,
  tenant_id,
  model_name AS view_model_name,
  virtual_model_name AS view_virtual_model_name,
  user_email AS view_user_email,
  virtual_account_name AS view_virtual_account_name,
  CAST(NULL, 'Nullable(String)') AS view_team_name,
  error_type AS view_error_type,

  countState() AS request_count,
  countIfState(status_code >= 400) AS failed_count,

  sumState(toUInt64(input_tokens)) AS input_tokens_sum,
  sumState(toUInt64(output_tokens)) AS output_tokens_sum,
  sumState(cost_usd) AS cost_usd_sum,

  quantileTDigestState(0.50)(latency_ms) AS latency_p50,
  quantileTDigestState(0.75)(latency_ms) AS latency_p75,
  quantileTDigestState(0.90)(latency_ms) AS latency_p90,
  quantileTDigestState(0.99)(latency_ms) AS latency_p99,

  quantileTDigestStateIf(0.50)(ttft_ms, ttft_ms IS NOT NULL) AS ttft_p50,
  quantileTDigestStateIf(0.75)(ttft_ms, ttft_ms IS NOT NULL) AS ttft_p75,
  quantileTDigestStateIf(0.90)(ttft_ms, ttft_ms IS NOT NULL) AS ttft_p90,
  quantileTDigestStateIf(0.99)(ttft_ms, ttft_ms IS NOT NULL) AS ttft_p99,

  quantileTDigestStateIf(0.50)(itl_ms, itl_ms IS NOT NULL) AS itl_p50,
  quantileTDigestStateIf(0.75)(itl_ms, itl_ms IS NOT NULL) AS itl_p75,
  quantileTDigestStateIf(0.90)(itl_ms, itl_ms IS NOT NULL) AS itl_p90,
  quantileTDigestStateIf(0.99)(itl_ms, itl_ms IS NOT NULL) AS itl_p99,

  quantileTDigestStateIf(0.50)(tpot_ms, tpot_ms IS NOT NULL) AS tpot_p50,
  quantileTDigestStateIf(0.75)(tpot_ms, tpot_ms IS NOT NULL) AS tpot_p75,
  quantileTDigestStateIf(0.90)(tpot_ms, tpot_ms IS NOT NULL) AS tpot_p90,
  quantileTDigestStateIf(0.99)(tpot_ms, tpot_ms IS NOT NULL) AS tpot_p99
FROM gateway_model_events
GROUP BY
  bucket_ts,
  tenant_id,
  view_model_name,
  view_virtual_model_name,
  view_user_email,
  view_virtual_account_name,
  view_team_name,
  view_error_type;
