# GodarAI xAI/Grok support

GodarAI uses `xai` as the canonical provider slug. The supported catalog is
limited to these four models and their catalog aliases:

| Canonical model | Aliases |
|---|---|
| `grok-4.3` | `grok-4.3-latest`, `grok-latest` |
| `grok-4.5` | `grok-4.5-latest`, `grok-build-latest` |
| `grok-imagine-image` | `grok-imagine-image-2026-03-02` |
| `grok-imagine-image-quality` | `grok-imagine-image-quality-20260403`, `grok-imagine-image-quality-latest`, `grok-imagine-image-pro` |

Every request is checked against the effective-dated model runtime profile
from Model Discovery. If that profile or its pricing revision is missing,
stale, invalid, or does not match the routed model, the request fails closed.

## Credentials

Tenant BYOK accounts use provider `xai`, Bearer authentication, and default API
base `https://api.x.ai/v1`. Credential validation probes `GET /v1/models`.
Custom API bases remain subject to GodarAI's provider-host and SSRF controls.

## Public API

Text models support the applicable model-profile subset of:

- `POST /v1/chat/completions`, including SSE and `x-grok-conv-id`;
- `GET /v1/chat/deferred-completion/{request_id}`;
- `POST /v1/responses`, stored response retrieve/delete, and Grok 4.5
  `POST /v1/responses/compact`;
- Files upload/list/get/update/delete/content and public-URL lifecycle operations.

For text inference, `service_tier` accepts only the currently documented xAI
values `default` and `priority`. Historical OpenAI-style values such as `auto`
or `standard` are rejected instead of being silently translated.

Image models support JSON `POST /v1/images/generations` and JSON
`POST /v1/images/edits`. GodarAI validates `n`, `aspect_ratio`, `resolution`,
`response_format`, prompt length, reference-image count, and storage options.
xAI multipart edits, image variations, and image streaming are rejected before
provider dispatch.

Files default to OpenAI for backward compatibility. Select xAI for a create or
list operation with:

```http
x-godarai-provider: xai
x-godarai-provider-account: <tenant-provider-account-id>
```

Once a file, response, deferred completion, or batch ID is recorded, GodarAI
routes later operations to its owning tenant, provider, and provider account.
Cross-tenant IDs return not found; ambiguous cross-provider IDs require the
provider header.

xAI file public URLs use:

- `POST /v1/files/{file_id}/public-url`
- `POST /v1/files/{file_id}/public-url/revoke`

`expires_after`, when supplied, must be between 3,600 and 2,592,000 seconds and
cannot outlive the file. GodarAI also rejects a known-ineligible file before
dispatch when recorded metadata shows a size over 50 MiB or a MIME type other
than PNG, JPEG, MP4, or PDF. The provider remains authoritative when older
resource mappings do not contain enough metadata for local prevalidation.

## xAI Batch (release-gated)

Use `x-godarai-provider: xai` on the shared create/list/get/cancel routes.
GodarAI translates cancellation to xAI's native `:cancel` route. Additional
native subresources are:

- `GET|POST /v1/batches/{batch_id}/requests`
- `GET /v1/batches/{batch_id}/results`

The native route compiler, resource ownership, `batch_requests` validation,
4.5 rejection policy, and model pricing rules are implemented. New Batch
creation and request submission currently return `503 billing_unavailable`.
They remain fail-closed until paginated and partially failed results are tied
to an exactly-once reservation and settlement state machine. This prevents
provider work from running without safe tenant charging and reconciliation.
Batch endpoints are therefore not advertised in the effective model runtime
profiles while this gate is open.

Once that release gate is closed, `grok-4.3` uses the catalog's discounted
Batch rates, `grok-4.5` is rejected, and both Imagine models use standard image
prices. Each nested request must resolve to the same xAI provider account that
owns the batch.

## Accounting and observability

Catalog estimates are pinned to a revision. Final successful text, Responses,
deferred, and image calls preserve and reconcile `cost_in_usd_ticks` as an
exact signed 64-bit value. Logs retain the exact ticks and
`num_server_side_tools_used`; token, cache, reasoning, image, long-context,
priority, and applicable Batch dimensions remain visible in the pricing
snapshot and usage breakdown.

Only idempotent xAI `GET` operations receive one bounded retry for `429`,
`502`, `503`, or `504`. `Retry-After` is honored up to five seconds. Mutating
and generation `POST` operations are never automatically replayed.

xAI's asynchronous file-storage and download charges are not yet settled to a
tenant ledger. Files remain a production release blocker until account-level
storage/download reconciliation and commercial charging policy are implemented;
these costs must not be inferred from an individual inference response.

All four profiles currently set `zdrEligible=false`. GodarAI therefore does
not advertise these models for a tenant/account policy that requires ZDR;
stateful Responses, deferred work, Files, URL image output, and public URLs are
not treated as ZDR-compatible by inference.

## Release status

The local adapter and contract implementation is not a production-readiness
certificate. xAI Batch settlement, Files storage/download reconciliation,
credentialed platform/BYOK conformance, live cost reconciliation, sustained
fault/load tests, dashboards, alerts, and release approvals remain required.
The readiness command and Postman
collection intentionally report failures or blocked cases instead of turning
missing evidence into support claims.

## Source of truth

The implementation follows current official xAI model, pricing, REST, Files,
Public URL, Deferred Completion, and Batch documentation. The historical
[Apidog reference](https://grok-api.apidog.io/llms.txt) is retained as a
secondary source where it does not conflict with current official xAI docs.
See [the execution plan](../plans/xai-four-model-production-readiness.md) for
the evidence policy, pricing tables, rollout gates, and unresolved external
ambiguities.
