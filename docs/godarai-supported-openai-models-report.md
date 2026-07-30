# OpenAI Responses support in GodarAI

**Audit date:** 2026-07-30  
**Provider:** OpenAI only  
**Private pricing:** Out of scope  
**Official REST baseline:** `openai/openai-openapi@db14b6e1712aaf5265cf5a6871adff7a9c61d31c`

## Result

GodarAI exposes all seven public REST operations in OpenAI's `Responses`
resource:

| Method | Public route | Status |
|---|---|---|
| `POST` | `/v1/responses` | Implemented |
| `POST` | `/v1/responses/compact` | Implemented |
| `POST` | `/v1/responses/input_tokens` | Implemented |
| `GET` | `/v1/responses/{response_id}` | Implemented |
| `DELETE` | `/v1/responses/{response_id}` | Implemented |
| `POST` | `/v1/responses/{response_id}/cancel` | Implemented |
| `GET` | `/v1/responses/{response_id}/input_items` | Implemented |

The gateway preserves unknown request and response fields, OpenAI error bodies,
SSE event bodies, query parameters, `X-Client-Request-Id`, `X-Request-Id`,
`Retry-After`, and the documented OpenAI request/token rate-limit headers.
Stored response IDs are bound to the tenant and provider account that created
them. A continuation may omit `model`; GodarAI derives its routing and billing
model from the stored `previous_response_id` mapping without adding a `model`
field to the provider request.

## Supported models

All requested models now declare `responses` support in the GodarAI catalog.

| Model | Context | Max output | REST streaming | Modalities |
|---|---:|---:|:---:|---|
| [`gpt-5.6-sol`](https://developers.openai.com/api/docs/models/gpt-5.6-sol) | 1.05M | 128K | Yes | Text input/output; image input |
| [`gpt-5.6-terra`](https://developers.openai.com/api/docs/models/gpt-5.6-terra) | 1.05M | 128K | Yes | Text input/output; image input |
| [`gpt-5.6-luna`](https://developers.openai.com/api/docs/models/gpt-5.6-luna) | 1.05M | 128K | Yes | Text input/output; image input |
| [`gpt-5.5`](https://developers.openai.com/api/docs/models/gpt-5.5) | 1.05M | 128K | Yes | Text input/output; image input |
| [`gpt-5.5-pro`](https://developers.openai.com/api/docs/models/gpt-5.5-pro) | 1.05M | 128K | No | Text input/output; image input |
| [`gpt-5.4`](https://developers.openai.com/api/docs/models/gpt-5.4) | 1.05M | 128K | Yes | Text input/output; image input |
| [`gpt-5.4-pro`](https://developers.openai.com/api/docs/models/gpt-5.4-pro) | 1.05M | 128K | Yes | Text input/output; image input |
| [`gpt-5.4-mini`](https://developers.openai.com/api/docs/models/gpt-5.4-mini) | 400K | 128K | Yes | Text input/output; image input |
| [`gpt-5.4-nano`](https://developers.openai.com/api/docs/models/gpt-5.4-nano) | 400K | 128K | Yes | Text input/output; image input |

Audio and video are not supported modalities for these models. Model-specific
reasoning-effort values, structured-output support, and the exact supported
Responses tools are imported into the runtime profile rather than inferred
from model size.

## Responses tools

`FC` means function calling. The remaining abbreviations are `WS` web search,
`FS` file search, `TS` tool search, `IG` image generation, `CI` code
interpreter, `HS` hosted shell, `AP` apply patch, `SK` skills, `CU` computer
use, and `MCP`.

| Model | FC | WS | FS | TS | IG | CI | HS | AP | SK | CU | MCP |
|---|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| GPT-5.6 Sol/Terra/Luna | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| GPT-5.5 | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| GPT-5.5 Pro | ✓ | ✓ | ✓ | — | ✓ | ✓ | ✓ | — | — | — | ✓ |
| GPT-5.4 | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| GPT-5.4 Pro | ✓ | ✓ | ✓ | ✓ | ✓ | — | — | ✓ | — | ✓ | ✓ |
| GPT-5.4 mini | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| GPT-5.4 nano | ✓ | ✓ | ✓ | — | ✓ | ✓ | ✓ | ✓ | ✓ | — | ✓ |

The gateway relays tool definitions and future tool fields losslessly. Tool
availability is still enforced by OpenAI for the selected model.

## Public token pricing

Prices are USD per one million tokens and come from the current
[OpenAI pricing table](https://developers.openai.com/api/docs/pricing). A dash
means OpenAI does not publish that discount.

| Model | Standard input / cached / output | Batch and Flex input / cached / output | Fast input / cached / output |
|---|---|---|---|
| `gpt-5.6-sol` | 5 / 0.5 / 30 | 2.5 / 0.25 / 15 | 10 / 1 / 60 |
| `gpt-5.6-terra` | 2 / 0.2 / 12 | 1 / 0.1 / 6 | 4 / 0.4 / 24 |
| `gpt-5.6-luna` | 0.2 / 0.02 / 1.2 | 0.1 / 0.01 / 0.6 | 0.4 / 0.04 / 2.4 |
| `gpt-5.5` | 5 / 0.5 / 30 | 2.5 / 0.25 / 15 | 12.5 / 1.25 / 75 |
| `gpt-5.5-pro` | 30 / — / 180 | 15 / — / 90 | — |
| `gpt-5.4` | 2.5 / 0.25 / 15 | 1.25 / 0.13 / 7.5 | 5 / 0.5 / 30 |
| `gpt-5.4-pro` | 30 / — / 180 | 15 / — / 90 | — |
| `gpt-5.4-mini` | 0.75 / 0.075 / 4.5 | 0.375 / 0.0375 / 2.25 | 1.5 / 0.15 / 9 |
| `gpt-5.4-nano` | 0.2 / 0.02 / 1.25 | 0.1 / 0.01 / 0.625 | — |

For 1.05M-context models, requests above 272K input tokens use the published
long-context rates for the full request. GodarAI selects separate Standard,
Batch, and Flex long-context dimensions. Fast/Priority is rejected for a
long-context request because OpenAI does not support that combination.
`service_tier: "fast"` and `"priority"` share the Fast pricing dimensions.

GPT-5.6 cache-write tokens are preserved from
`usage.input_tokens_details.cache_write_tokens` and billed separately. For Pro
models, cached tokens fall back to the normal input price because OpenAI offers
no cached-input discount.

The catalog also records current public tool rates: web search
($10/1,000 calls), file search ($2.50/1,000 calls), file-search storage
($0.10/GB/day after the free allowance), and container session prices for
1 GB, 4 GB, 16 GB, and 64 GB.

## Provider rate limits

GodarAI forwards OpenAI's live rate-limit headers and does not replace the
provider's usage-tier decision. This report records the official Tier 1–5 RPM,
TPM, and Batch queue limits from each linked model page.

Common Standard profile:

| Tier | RPM | TPM | Batch queue |
|---:|---:|---:|---:|
| 1 | 500 | 500,000 | 1,500,000 |
| 2 | 5,000 | 1,000,000 | 3,000,000 |
| 3 | 5,000 | 2,000,000 | 100,000,000 |
| 4 | 10,000 | 4,000,000 | 200,000,000 |
| 5 | 15,000 | 40,000,000 | 15,000,000,000 |

This common profile applies to GPT-5.6 Sol/Terra and GPT-5.5/GPT-5.4 standard
context. Luna and the mini model have higher upper-tier TPM; nano, Pro, and
long-context requests have their own tables. Consumers should use the linked
model page and the live response headers as the authoritative values.

## Known incompatibilities

The REST implementation is contract-ready, but the complete OpenAI Responses
surface is **not yet production-ready** for these reasons:

1. **Responses WebSocket mode is not public in GodarAI.** OpenAI documents a
   persistent WebSocket connection at `/v1/responses`; GodarAI currently
   exposes HTTP and SSE only. See the
   [OpenAI WebSocket guide](https://developers.openai.com/api/docs/guides/websocket-mode).
2. **Tool-cost reconciliation is incomplete.** Public tool prices are stored
   in the catalog, but file-search storage and container duration/memory are
   not fully attributable from a single Responses body. They are not yet
   automatically debited by a provider-usage reconciliation worker.
3. **No credentialed live canary was run.** The audit is based on official
   schemas, pinned route data, catalog import, and automated contract tests.

Because of these gaps, the static readiness command reports all seven REST
routes as ready while leaving `websocket_ready`, `production_ready`, and
`live_validated` false.

## Verification

- OpenAI route baseline:
  [`packages/contracts/openapi/openai-scoped-routes-db14b6e.json`](../packages/contracts/openapi/openai-scoped-routes-db14b6e.json)
- REST routing and account affinity:
  [`services/control-plane/internal/proxy/responses_operations_handler.go`](../services/control-plane/internal/proxy/responses_operations_handler.go)
- Request, response, error, and SSE relay:
  [`services/control-plane/internal/proxy/responses_handler.go`](../services/control-plane/internal/proxy/responses_handler.go)
- Catalog importer and runtime profile:
  [`apps/model-discovery/scripts/import-models-from-yaml.ts`](../apps/model-discovery/scripts/import-models-from-yaml.ts)
- Readiness audit:
  [`services/control-plane/cmd/openai-readiness`](../services/control-plane/cmd/openai-readiness)
