# DeepSeek two-model production-readiness plan

Status: not production-ready; implemented behind provider and model release gates
Last verified against official provider sources: 2026-08-15

## Scope lock

The direct `deepseek` provider contains exactly:

- `deepseek-v4-flash`
- `deepseek-v4-pro`

The enabled upstream operations are `GET /models`, synchronous or SSE
`POST /chat/completions` for both models, and stateless synchronous or SSE
`POST /responses` for both models. Thinking is a mode of each
model. Legacy aliases and third-party-hosted DeepSeek weights are not catalog
entries for this provider.

## Official-source manifest

All provider facts were observed on 2026-08-15. No third-party model directory
is an API authority for this implementation.

| Fact                                                      | Official source                                                                  | Effective/lifecycle note                            |
| --------------------------------------------------------- | -------------------------------------------------------------------------------- | --------------------------------------------------- |
| Bearer authentication and `https://api.deepseek.com`      | [API overview](https://api-docs.deepseek.com/api/deepseek-api)                   | Current OpenAI-format API                           |
| Exact two-model list                                      | [List Models](https://api-docs.deepseek.com/api/list-models)                     | Current direct API response                         |
| Context, output, features, prices and concurrency         | [Models and Pricing](https://api-docs.deepseek.com/quick_start/pricing)          | Prices observed 2026-08-05                          |
| Chat request, response and usage fields                   | [Chat Completions](https://api-docs.deepseek.com/api/create-chat-completion)     | P0 inference contract                               |
| Thinking defaults, effort mapping and `reasoning_content` | [Thinking Mode](https://api-docs.deepseek.com/guides/thinking_mode)              | Default enabled/high; mappings are model-specific   |
| Function tools                                            | [Tool Calls](https://api-docs.deepseek.com/guides/tool_calls)                    | Caller-defined functions only in P0                 |
| JSON Object output                                        | [JSON Output](https://api-docs.deepseek.com/guides/json_mode)                    | `json_object`, not strict schema                    |
| Automatic cache behavior                                  | [Context Caching](https://api-docs.deepseek.com/guides/kv_cache)                 | Enabled by default                                  |
| Usage interpretation                                      | [Token Usage](https://api-docs.deepseek.com/quick_start/token_usage)             | Provider-returned usage is authoritative            |
| Isolation, keep-alive and concurrency                     | [Rate Limit and Isolation](https://api-docs.deepseek.com/quick_start/rate_limit) | `user_id` is gateway-managed for shared credentials |
| Error status classes                                      | [Error Codes](https://api-docs.deepseek.com/quick_start/error_codes)             | 400/401/402/422/429/500/503                         |
| Lifecycle and retired legacy IDs                          | [Change Log](https://api-docs.deepseek.com/updates)                              | Flash public beta; Pro GA since 2026-08-13          |
| Responses surface for both models                         | [Responses API](https://api-docs.deepseek.com/api/create-response)               | Stateless create for Flash and Pro                  |
| Responses behavior and streaming                          | [Responses guide](https://api-docs.deepseek.com/guides/responses_api)            | No stored resources and no `[DONE]` sentinel        |
| Deferred compatibility surface                            | [Anthropic API](https://api-docs.deepseek.com/guides/anthropic_api)              | Not publicly proxied by GodarAI P0                  |

## Delivered controls

| Area            | Control                                                                                                                                                                                                          |
| --------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Catalog         | Exact allowlist, Flash preview and Pro GA lifecycle, pinned upstream versions, text-only profiles, immutable profile versions, official provenance, three positive USD token rates.                              |
| Credentials     | BYOK and platform-managed API keys, official-host validation, read-only model-list probe and write-only secret schema.                                                                                           |
| Isolation       | Platform-managed dispatch requires `DEEPSEEK_USER_ID_HMAC_SECRET`; a tenant/account-scoped opaque `user_id` is injected only in Silk Road. Missing isolation configuration fails closed; BYOK remains available. |
| Runtime         | Dedicated validation for model, endpoint, transport, roles, limits, thinking/effort, ineffective sampling, function tools and output formats; Responses create is enabled for both current models.               |
| Provider egress | Dedicated adapter with fixed paths, Bearer injection, exact body preservation, keep-alive passthrough, no billable POST retry and one bounded safe-read retry.                                                   |
| Billing         | Exact decimal catalog rates; cache-hit, cache-miss and output usage dimensions; cache mismatch, missing provenance or zero pricing fails closed.                                                                 |
| Privacy         | Caller `user_id`/`user` is rejected; raw tenant identity, keys, prompts, tool data and reasoning are excluded from provider-specific operational fields.                                                         |

## Product decisions and deferred capability gates

Responses create is enabled for Flash and Pro and remains stateless: storage,
retrieval, deletion, cancellation, compaction, input-token counting and
`previous_response_id` are disabled. Anthropic compatibility,
FIM, prefix completion, hosted tools, Files, Batch, embeddings and all non-text
modalities require independent runtime, security, billing and release scopes.

The current catalog rates remain valid through 2026-08-16 16:00 UTC. From that
instant the catalog contains peak/off-peak pricing, with peak windows at
01:00-04:00 and 06:00-10:00 UTC and off-peak prices at half the peak rates. The
selected UTC pricing period is pinned in the request-time billing snapshot so a
request crossing a window boundary retains its reserved rate.

## Live evidence required before release

1. Capture `GET /models` evidence containing both canonical IDs for every approved account.
2. Run bounded sync and SSE canaries for both models with thinking enabled and disabled.
3. Run a multi-turn function-tool loop and prove lossless `reasoning_content` return without trace-content logging.
4. Run Responses sync and SSE canaries for both models and verify terminal events without `[DONE]`.
5. Exercise JSON Object, keep-alive comments, cancellation, disconnect and documented error classes.
6. Reconcile cache-hit, cache-miss and completion tokens against DeepSeek charges and the GodarAI USD/IRR ledger.
7. Prove two tenants sharing one managed credential receive distinct opaque IDs; attach a security/privacy review of the HMAC secret lifecycle.
8. Attach load evidence below the current account concurrency reference, dashboard and alert links, incident owner, rollback owner and sign-offs from product, engineering, security, finance and operations.
9. Re-open all official sources and resolve source drift immediately before enabling a model gate.

The readiness runner must continue to report `production_ready: false`; it is an
evidence collector, not a release authority.

## Rollback

Disable the two model gates first, then disable the `deepseek` provider gate.
Allow in-flight requests to finish; do not retry unknown billable outcomes.
Retain immutable pricing/runtime revisions and spend records for reconciliation.
Restore only after the failed evidence item is corrected and the release owners
approve a new canary.

## Known gaps

- CI does not execute billable DeepSeek calls.
- Live Pro GA canaries and current effort mappings still require account-specific evidence.
- Peak/off-peak selection is implemented, but production billing reconciliation across both UTC windows still requires evidence.
- Production dashboards, alerts, canary reports, provider invoices and formal sign-offs are external release artifacts and are not attached yet.
