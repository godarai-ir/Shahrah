# Model Metrics Contract v1

## Purpose
This document freezes the MVP contract for model metrics events, error taxonomy, and metadata whitelist used by Shahrah control-plane and ClickHouse metrics queries.

## Event Schema
All timestamps are UTC and RFC3339-compatible.

### Required fields
- `tenant_id` (string, non-empty)
- `request_id` (string, non-empty)
- `ts` (timestamp)
- `model_name` (string, non-empty)
- `request_type` (enum string)
- `status_code` (int, 100..599)
- `error_type` (enum string)
- `latency_ms` (uint >= 0)
- `input_tokens` (uint >= 0)
- `output_tokens` (uint >= 0)
- `cost_usd` (decimal >= 0)

### Optional fields
- `virtual_model_name` (string|null)
- `provider_name` (string|null)
- `provider_model_name` (string|null)
- `user_id` (string|null)
- `user_email` (string|null)
- `virtual_account_id` (string|null)
- `virtual_account_name` (string|null)
- `team_ids` (string[])
- `team_names` (string[])
- `ttft_ms` (uint|null)
- `itl_ms` (uint|null)
- `tpot_ms` (uint|null)
- `metadata` (object with whitelist keys)

### Request type enum (v1)
- `chat_completions`
- `completions`
- `responses`
- `embeddings`
- `unknown`

## Error Type Taxonomy v1
- `none`: success (`status_code < 400`)
- `client_error`: request-level client error (`4xx`) excluding quota/rate/auth buckets below
- `auth_error`: auth or permission failures (`401/403`)
- `rate_limit_error`: rate limit exceeded (`429`)
- `budget_exceeded`: payment/budget rejection (`402`)
- `provider_error`: upstream/provider failures (`5xx`, provider bad gateway)
- `provider_timeout`: upstream timeout/cancel
- `routing_error`: route/fallback resolution failure
- `tunnel_unavailable`: tunnel path unavailable
- `internal_error`: unexpected internal error before provider call

## Metadata Whitelist Policy v1
Allowed metadata keys:
- `environment`
- `feature`
- `endpoint`
- `key_type`
- `org_unit_id`
- `team_id`

Policy:
- max key length: `64`
- max string value length: `256`
- max keys per event: `16`
- non-whitelisted keys are dropped

## Valid Event Example
```json
{
  "tenant_id": "9b95f190-b0f4-4ea5-b8d8-7dc8d11a4e2d",
  "request_id": "trace_01jv6r1hvhg7d",
  "ts": "2026-05-06T13:30:25.111Z",
  "model_name": "gpt-4o-mini",
  "virtual_model_name": "smart-default",
  "provider_name": "openai",
  "provider_model_name": "gpt-4o-mini",
  "request_type": "chat_completions",
  "user_id": "3dcb26d3-7d18-4d17-98c3-0bda627c2f8f",
  "user_email": "user@example.com",
  "virtual_account_id": null,
  "virtual_account_name": null,
  "team_ids": ["6e2c4ab6-4f4c-40ef-8b36-f75b79d1d1d4"],
  "team_names": ["ml-team-a"],
  "status_code": 200,
  "error_type": "none",
  "latency_ms": 842,
  "ttft_ms": 210,
  "itl_ms": 33,
  "tpot_ms": 29,
  "input_tokens": 510,
  "output_tokens": 230,
  "cost_usd": 0.001928,
  "metadata": {
    "environment": "prod",
    "feature": "assistant-chat",
    "endpoint": "/v1/chat/completions"
  }
}
```

## Invalid Examples
### Invalid #1: missing tenant_id
```json
{
  "request_id": "trace_1",
  "model_name": "gpt-4o-mini",
  "request_type": "chat_completions",
  "status_code": 200,
  "error_type": "none",
  "latency_ms": 1,
  "input_tokens": 1,
  "output_tokens": 1,
  "cost_usd": 0.0001
}
```
Expected validation error:
- `tenant_id is required`

### Invalid #2: status_code out of range
```json
{
  "tenant_id": "t1",
  "request_id": "r1",
  "model_name": "gpt-4o-mini",
  "request_type": "chat_completions",
  "status_code": 700,
  "error_type": "none",
  "latency_ms": 10,
  "input_tokens": 10,
  "output_tokens": 20,
  "cost_usd": 0.001
}
```
Expected validation error:
- `status_code must be between 100 and 599`

### Invalid #3: unknown metadata key
```json
{
  "tenant_id": "t1",
  "request_id": "r1",
  "model_name": "gpt-4o-mini",
  "request_type": "chat_completions",
  "status_code": 200,
  "error_type": "none",
  "latency_ms": 10,
  "input_tokens": 10,
  "output_tokens": 20,
  "cost_usd": 0.001,
  "metadata": {
    "random_unbounded_key": "value"
  }
}
```
Expected validation behavior:
- unknown key dropped from persisted event
- optional warning log emitted by validator

## Decision Log
- Store timezone is UTC only.
- `ttft_ms`, `itl_ms`, and `tpot_ms` are nullable for non-streaming responses.
- Team aggregation on raw data (`arrayJoin`) is accepted in MVP; dedicated team MV is deferred.
- Metrics writes are best-effort and must not fail inference request path.
