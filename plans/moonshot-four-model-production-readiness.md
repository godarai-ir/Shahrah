# Moonshot four-model production-readiness plan

Status: implemented behind release gates
Last verified against official provider sources: 2026-08-05

## Scope

This delivery adds exactly these Moonshot models to GodarAI:

- `kimi-k3`
- `kimi-k2.7-code`
- `kimi-k2.7-code-highspeed`
- `kimi-k2.6`

The enabled upstream surface is deliberately limited to `GET /v1/models` and
`POST /v1/chat/completions`, including SSE streaming. Files, batches,
embeddings, provider-hosted web search, and any undocumented operation remain
disabled until their runtime, billing, security, and live-readiness contracts
are implemented.

## Source manifest

Provider facts come only from the official Moonshot/Kimi sources below. The
catalog records the model-specific source URLs next to each price and runtime
profile.

- [Machine-readable documentation index](https://platform.kimi.ai/docs/llms.txt)
- [Official OpenAPI document](https://platform.kimi.ai/docs/openapi.json)
- [Model catalog](https://platform.kimi.ai/docs/models)
- [Model capability overview](https://platform.kimi.ai/docs/api/models-overview)
- [Chat Completions API](https://platform.kimi.ai/docs/api/chat)
- [Platform changelog](https://platform.kimi.ai/docs/platform-changelog)
- [Kimi K3 pricing](https://www.kimi.com/resources/kimi-k3-pricing)
- [Kimi K2.7 Code pricing](https://www.kimi.com/resources/kimi-k2-7-code-pricing)
- [Kimi K2.6 pricing](https://www.kimi.com/en/resources/kimi-k2-6-pricing)

## Delivered contracts

| Area          | Delivered behavior                                                                                                                                                     |
| ------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Catalog       | Exact four-model allowlist, capability/runtime profiles, context limits, source URLs, and per-token USD pricing.                                                       |
| Credentials   | API-key schema, canonical `https://api.moonshot.ai/v1` base URL, official-host allowlist, and non-billable model-list probe.                                           |
| Runtime       | Dedicated Moonshot adapter, Bearer auth, HTTP/SSE chat, model-specific thinking/sampling/tool/modality validation, and lossless `reasoning_content`.                   |
| Reliability   | No automatic retry for billable `POST`; one bounded retry for retryable model-list reads with capped `Retry-After`.                                                    |
| Billing       | Separate uncached-input, cache-read, and total-output dimensions; missing catalog pricing fails closed; reasoning output is not double-counted.                        |
| Observability | Provider/model/account routing, runtime-profile revision, pricing revision, reasoning mode, token usage, and categorized errors remain in existing traces and metrics. |
| Readiness     | A bounded gateway-only runner covers model discovery and one minimal request per model without accepting or recording an upstream key.                                 |
| Documentation | Persian setup, synchronous/streaming/tool-loop examples, capability differences, pricing, limits, security, and troubleshooting.                                       |

## Documented facts and product decisions

Documented provider facts include model IDs, global endpoint, Bearer
authentication, message/tool contracts, reasoning controls, multimodal support,
context windows, published prices, and account-dependent rate limits.

GodarAI product decisions include the exact four-model allowlist, disabling
unimplemented Moonshot endpoints, rejecting invalid combinations before
dispatch, fail-closed billing, not retrying billable requests, and requiring
release evidence beyond a smoke test. These choices are conservative gateway
policy, not claims about missing provider capabilities.

## Release gates

The implementation is not production-ready merely because unit and smoke tests
pass. Release owners must attach evidence for all of the following:

1. Run the Moonshot readiness command with an approved spend ceiling and a
   dedicated provider account.
2. Reconcile observed uncached-input, cached-input, and output charges against
   the Moonshot dashboard for every model.
3. Verify non-streaming, streaming, tool loops, structured output, image input,
   and K3 video input in an isolated environment.
4. Confirm rate-limit header forwarding and 429 behavior using the account's
   current official tier rather than a hard-coded limit.
5. Review log/trace redaction, retention, regional policy, alerting, dashboards,
   rollback, and incident ownership.
6. Obtain product, engineering, security, finance, and operations approval.

## Known gaps

- No live or billable Moonshot request is executed by CI.
- Rate limits and pricing tiers are account-dependent and can change; current
  provider data must be rechecked at release time.
- The readiness runner intentionally reports `production_ready: false`; it
  cannot certify load, privacy, finance, or operational gates.
- Moonshot native Files, Batch, Embeddings, and server-side search are outside
  this release.
