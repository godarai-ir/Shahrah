# Z.ai Four-Model Production Readiness

Last verified: 2026-08-17 (Asia/Tehran)

## Scope

This plan gates direct, first-class support for exactly `glm-5.2`,
`glm-4.7-flashx`, `glm-5v-turbo`, and `glm-ocr` under the canonical provider
slug `zai`. The existing `xai` provider is unrelated and must remain isolated.

## Source manifest

| Source                                               | Observed checksum (SHA-256)                                        | Authority                                   |
| ---------------------------------------------------- | ------------------------------------------------------------------ | ------------------------------------------- |
| `https://docs.z.ai/openapi.json`                     | `76f83c8525f993e33c79c5483e4cdde7db0924b5c5e5d7480b9b574ce909f4da` | Request/response schemas and endpoint paths |
| `https://docs.z.ai/llms.txt`                         | `60b34406d17fe6804b7c719e96886a8ad30082ada8b2187f85dcbc739cc7c25b` | Official documentation inventory            |
| `https://docs.z.ai/guides/overview/pricing`          | Observed 2026-08-17                                                | Public USD prices                           |
| `https://docs.z.ai/guides/develop/http/introduction` | Observed 2026-08-17                                                | Authentication and general API base URL     |

Catalog model files retain the complete first-party source list used for each
fact. A checksum change is a review signal; it must not silently rewrite an
effective-dated pricing row or an immutable runtime profile revision.

## Approved product contract

| Model            | Endpoint            | Input                        | Output         | Context / max output                | Published price per 1M tokens (input / cached input / output) |
| ---------------- | ------------------- | ---------------------------- | -------------- | ----------------------------------- | ------------------------------------------------------------- |
| `glm-5.2`        | `/chat/completions` | text                         | text           | 1,000,000 / 128,000                 | $1.40 / $0.26 / $4.40                                         |
| `glm-4.7-flashx` | `/chat/completions` | text                         | text           | 200,000 / 128,000                   | $0.07 / $0.01 / $0.40                                         |
| `glm-5v-turbo`   | `/chat/completions` | text, image, video, document | text           | 200,000 / 128,000                   | $1.20 / $0.24 / $4.00                                         |
| `glm-ocr`        | `/layout_parsing`   | image, PDF                   | text, document | provider-defined / provider-defined | $0.03 / not published / $0.03                                 |

The direct general API base URL is `https://api.z.ai/api/paas/v4`. The only
enabled auth mode is a write-only API key mapped to `Authorization: Bearer`.
The Coding Plan API endpoint is out of scope. The upstream API has no general
`GET /models` credential probe, so credential testing is an explicit,
user-triggered, billable one-token `glm-4.7-flashx` request. It must display
that fact before dispatch and must never run while saving an account.

## Conflict register and live gates

| Conflict                                                                                                                                | Safe contract before live evidence                                                                                  | Release gate                                                                                                       |
| --------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------ |
| GLM-5V model prose advertises function calling, while the current OpenAPI tool description limits vision tools to other model families. | Do not advertise or forward function tools for `glm-5v-turbo`.                                                      | Successful bounded tool-call fixture captured from an approved direct account, followed by a profile-version bump. |
| The current GLM-5V endpoint schema does not define `response_format`.                                                                   | Do not advertise or forward structured-output controls for `glm-5v-turbo`; retain JSON Object only for text models. | A first-party endpoint schema revision plus a bounded response-shape fixture.                                      |
| GLM-OCR prose mentions 100 pages, while OpenAPI specifies 30.                                                                           | Reject a single request above 30 pages; no transparent chunking.                                                    | Provider confirmation or bounded 31-page probe plus ordering/accounting tests.                                     |
| Chat OpenAPI permits up to four stop strings, while prose describes one.                                                                | Accept at most one stop string.                                                                                     | Provider confirmation or bounded multi-stop probe and profile revision.                                            |
| Persian OCR quality/language coverage is not guaranteed.                                                                                | Permit Persian documents but make no quality claim.                                                                 | Versioned Persian fixture evaluation reaches the approved threshold.                                               |
| Upstream account tiers and rate limits are account-specific.                                                                            | Do not invent tier names or hard-coded RPM/TPM. Preserve 429 metadata and use GodarAI tenant limits independently.  | Authenticated account evidence may populate observed, non-contractual limits.                                      |
| Cache storage is described as limited-time free without an end date.                                                                    | Bill cache reads at the published cached-token rate; do not create an eternal zero storage rate.                    | A new effective-dated pricing row when Z.ai publishes an end date or price.                                        |

## Release gates

- Exactly four curated YAML models import under provider `zai`; unofficial
  sources, missing prices, zero prices, unknown models, and profile drift fail.
- Provider-account secrets remain write-only and are never present in logs,
  traces, errors, API responses, or readiness artifacts.
- The general API host allowlist is exact (`api.z.ai`), HTTPS-only, and applied
  after redirects/DNS resolution by the existing egress controls.
- Chat, streaming, thinking, custom function loops for the two text models,
  JSON Object output for the text models, multimodal validation, native layout parsing, and the OCR
  alias have contract and adapter tests.
- OCR uses provider-reported input/output tokens. Mistral page accounting must
  never be reused for Z.ai.
- Reservation and settlement use exact decimal rates and fail before dispatch
  for absent/incompatible prices. Cache reads settle from
  `prompt_tokens_details.cached_tokens`.
- Public aliases `/api/paas/v4/chat/completions` and
  `/api/paas/v4/layout_parsing` share the canonical authentication, limits,
  billing, logging, and tracing pipeline.
- Tenant-admin supports provider creation, deployment, catalog discovery,
  model-aware Playground inputs, usage/cost visibility, and Persian/English
  errors without exposing a secret.
- Canary order is `glm-4.7-flashx`, `glm-5.2`, `glm-5v-turbo`, then `glm-ocr`.
  Provider, model, and OCR-specific kill switches are reversible without a
  migration rollback.

## Live probe budget

No live request is authorized by this document. An operator must provide the
account, API key, region confirmation, and a USD ceiling. Each probe records
request ID, model, endpoint, status, finish reason, usage dimensions, actual
provider charge if available, and redacted response shape. Fixtures must not
contain credentials, customer content, or full proprietary responses.

## Local verification status

Completed on 2026-08-17:

- model-discovery tests, typecheck, build, curated Z.ai importer validation,
  and targeted lint;
- contracts, Silk Road, and Control Plane tests, `go vet`, and configured
  `golangci-lint` checks;
- tenant-admin Z.ai provider-account, multimodal Playground, GLM-OCR
  Playground, RTL translations, related node/DOM tests, ESLint, i18n,
  encoding, and production build;
- provider-aware streaming without unsupported `stream_options`, Z.ai
  sampling/thinking controls, remote-only video/document validation, and
  local image/OCR size checks;
- Postman generator coverage for provider discovery and native
  `/v1/layout_parsing` smoke requests;
- documentation typecheck and production build.

Production remains gated. No live Z.ai request, provider-dashboard cost
reconciliation, account-tier observation, Persian OCR evaluation, or
production canary was run because no credential and explicit spend ceiling
were authorized. Browser automation reached the local tenant-admin process,
but an authenticated visual review could not proceed while its required local
Control Plane at `127.0.0.1:4000` was unavailable; automated DOM coverage and
the production build passed, but desktop/narrow screenshots remain a release
review item.

## Rollback

Disable Z.ai routing and new deployments through provider/model feature flags,
then drain in-flight requests. Keep historical catalog, usage, invoice, log,
and trace rows readable. Revoke the affected provider-account secret if the
rollback is security-related. Do not delete or rewrite settled usage.
