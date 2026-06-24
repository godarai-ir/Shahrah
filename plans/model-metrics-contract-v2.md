# Model Metrics Contract v2: Cache Savings

## Purpose

Extend the v1 model-metrics event with auditable gateway-cache effectiveness and estimated provider cost avoided. All v1 fields and semantics remain compatible.

## Accounting Semantics

- `cost_usd` remains actual provider cost incurred by the request.
- A cache hit has `cost_usd = 0`.
- `avoided_cost_usd` and `avoided_cost_irr` come from the immutable finalized cost snapshot stored when the cached response was created.
- Savings are gross estimates. BGE-M3 embedding and Redis infrastructure costs are excluded until `embedding_cost_usd` is populated from an auditable pricing source.
- `cost_without_cache = cost_usd + avoided_cost_usd` is derived at query/UI time.
- `savings_rate = avoided_cost_usd / cost_without_cache`; it is zero when the denominator is zero.

## Added Event Fields

| Field | Type | Values / meaning |
|---|---|---|
| `cache_mode` | string | `none`, `simple`, `smart` |
| `cache_status` | string | `none`, `hit`, `miss`, `bypass`, `error` |
| `cache_match` | string | `none`, `exact`, `semantic` |
| `cache_similarity` | nullable float | Semantic similarity for semantic hits only |
| `cache_origin_request_id` | nullable string | Trace/request that populated the entry |
| `avoided_input_tokens` | uint | Original cached response input tokens represented by the hit |
| `avoided_output_tokens` | uint | Original cached response output tokens represented by the hit |
| `avoided_cost_usd` | decimal | Gross provider cost avoided at the original pricing snapshot |
| `avoided_cost_irr` | decimal | Same estimate using the original exchange-rate snapshot |
| `embedding_cost_usd` | decimal | Reserved for measured semantic lookup cost; zero until configured |
| `pricing_version` | string | Pricing calculation version stored with the cache entry |

## Hit-Rate Semantics

`cache_hit_rate = hits / (hits + misses)`.

Requests with `cache_status` equal to `none`, `bypass`, or `error` are excluded from the denominator. Smart-mode exact hits are attributed as `cache_mode=smart`, `cache_match=exact`.

## Retention and Isolation

Cache savings use the same ClickHouse tenant scope, filters, and retention as existing model metrics. Raw prompts, cache keys, embeddings, API keys, and namespaces are not persisted in model metrics.
