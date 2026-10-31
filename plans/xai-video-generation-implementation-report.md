# xAI Video Generation: Implementation Report

Branch: `feat/xai-video-generation`. Executes `plans/xai-video-generation-executable-prompt.md`.
Status: implemented and tested offline. **No live xAI call has been made** and no probe budget has been approved, so every item in the conflict register below is still on its fail-closed default.

## Files changed by area

| Area | Files |
|---|---|
| Contracts | `packages/contracts/platform/model_runtime_video.go` (new): profile `videoControls` block, its validation, and `ValidateVideoRequest`. `packages/contracts/platform/xai_video.go` (new): strict xAI body parser. `packages/contracts/platform/model_runtime_profile.go`: adds the field. `packages/contracts/platform/types.go`: `xai_videos_*` operations; dead `openai_videos_*` operations removed. |
| Silk Road | `services/silk-road/pkg/server/xai_video.go` (new): re-validates requests and does the guarded content fetch. Also `xai.go`, `openai_operations.go` (xAI paths replace the Sora paths; legacy `/videos` passthrough removed) and `provider_contracts_test.go` (Gemini `/videos` row removed). |
| Control plane | `internal/proxy/videos_handler.go`, `videos_request.go`, `videos_provider.go` (codec seam and xAI codec), `videos_worker.go` (reconciliation and settlement). Wiring: `internal/server/routes_proxy.go`, `internal/server/server.go`, `internal/server/dependencies.go`, `internal/bootstrap/build_gateway.go`, `internal/config/provider_features.go`, `internal/proxy/dependencies.go`. Registry: `CapabilityVideos` moved from `openai` to `xai`. |
| Storage | `internal/storage/provider_native_resource_repo.go`: tenant find and list methods, and a status-guarded metadata patch. `db/main/migrations/087_provider_native_resource_tenant_listing.sql`: listing index. |
| Billing | `pricing_engine.go`: `PriceVideoUsage` settles to `cost_in_usd_ticks`. `catalog_pricing_snapshot.go`: a video model must have an unconditional `output_video_seconds` rate. `provider_pricing_strategy.go`: xAI metered usage now fails closed. |
| Catalog | `apps/model-discovery/providers/xai/grok-imagine-video{,-1.5,-1.5-lite}.yaml`, the importer allow-list (`scripts/model-import/providers/xai.ts`), and the runtime-profile Zod schema. No Prisma migration: catalog models ship through the YAML import, as the image models did. |
| Readiness | `cmd/xai-readiness`: a create/poll/download lifecycle on lite (1 s, 480p, ≤ $0.02), a validation rejection, and an unknown-ID 404. |
| UI and docs | `provider-models-list.tsx`: the snippet now shows `POST /v1/videos`. New Persian docs page `apps/docs/docs/supported-apis/video-generation.mdx` plus its sidebar entry. Postman generator: real, capability-gated video request; Sora placeholder removed. |

## Routes

The routes are always on. `SHAHRAH_VIDEOS_CANARY_TENANT_IDS` optionally restricts them to listed tenants; every other tenant gets 404. The reconciliation worker runs for the life of the gateway.

| Route | State |
|---|---|
| `POST /v1/videos` (JSON and multipart), `GET /v1/videos`, `GET /v1/videos/{id}`, `DELETE /v1/videos/{id}`, `GET /v1/videos/{id}/content` | Enabled |
| `POST /v1/videos/generations` (xAI-native), `POST /v1/videos/edits`, `POST /v1/videos/extensions` (both shapes) | Enabled |
| `content?variant=thumbnail\|spritesheet` | Rejected by the profile (`unsupported_variant`); xAI has no such variants |
| `remix` | Not offered (OpenAI deprecated it; xAI has no equivalent) |
| `output.upload_url`, `storage_options`, caller-supplied `reference_audios[].url` | Rejected by the strict parser (gated) |

## Conflict register

| # | Question | Fail-closed choice now | Probe that resolves it |
|---|---|---|---|
| 1 | Is the per-second price flat (Models page) or does it depend on resolution (guides)? | Settle from `cost_in_usd_ticks`. The catalog rate is used only for the reservation, and a drift above the reservation is logged. | Generate 480p and 1080p clips; compare the ticks. |
| 2 | Is `expired` a real poll status? | Mapped to `failed` / `upstream_expired`; the reservation is released, not charged. The 2-hour `upstream_timeout` is treated the same. | Poll an old request ID. |
| 3 | How long does `video.url` stay valid? | The URL is never stored; every download polls for a fresh one. | Download the same job at increasing delays. |
| 4 | Are failed or moderation-blocked jobs billed? | Charge only when the poll reports ticks > 0; otherwise release the reservation. | Run one forced failure and one moderation block; read the ticks. |
| 5 | Is an extension's `video.duration` the new segment or the stitched total? | Bill the requested segment seconds when ticks are missing. | Extend a 4 s clip by 2 s and read the duration. |
| 6 | How are edits priced? | Reserve 9 s (8.7 s source, rounded up); settle from ticks. | One edit; read the ticks. |
| 7 | Rate limits and GA status of the video models | Not modeled; the existing per-key rate limits apply. | Account console / support. |
| 8 | Can the xAI SDK's base URL point at GodarAI? | The REST shape is supported; the SDK is untested. | Point the SDK at staging. |
| 9 | Full preset voice roster | Only the documented `eve`, `ara`, `leo` are allowed. | `GET /v1/tts/voices` (or the equivalent) on a live account. |
| 10 | Which hosts serve video media? | Only `vidgen.x.ai` (from the docs example). Any other host is refused with `xai_video_content_host_not_allowed: <host>`. | First live download; extend `contentHosts` in the YAML if needed. |
| 11 | Inline image size limit | 10 MB per image (`maxInputImageBytes`). xAI does not document a limit. | Upload a large first frame. |
| 12 | The reference-to-video 10 s limit on `grok-imagine-video` | Not expressible in the profile; xAI rejects it upstream and the reservation is released. | n/a |

## Settlement guarantees (tested)

- **Claiming:** each job is claimed (`queued` / `in_progress` → `reconciling`) before it is polled, whether the worker or a client `GET` triggers the poll.
- **At-most-once charging:** the `settling` state is persisted *before* the budget commit and spend log. A settlement interrupted for more than 5 minutes is finished without charging again and logged at error level. This trades a rare lost charge for never double-charging, because `spend_logs.request_id` is not unique.
- **No status regression:** delete and the worker use `PatchProviderNativeResourceState` (a JSONB merge guarded by the expected status), so a delete can never move a settled job back to pending.
- **Deleting a pending job** still settles it; xAI cannot cancel a job.

## Product decision needed

Finished videos are **not** persisted by GodarAI. Downloads stream from xAI's temporary URL, so a tenant who downloads late gets `video_content_unavailable`. If longer retention is wanted, the options are xAI `storage_options` (files kept up to 30 days, recorded through the existing xAI file mapping) or GodarAI object storage.

## Remaining debt

- **No live probe has run.** Every conflict item stays at its default until `xai-readiness` runs against an approved account and budget.
- **No Playground panel.** No existing tenant-admin pattern fits async media.
- **Tenant-admin has no video badge.** Models show as `video_generation` through the existing model-type mapping.
- **Metrics:** a model metric is emitted per settlement; queue-depth and time-to-complete gauges are not added (warnings are logged instead).
- **Docs** exist in Persian only, matching the docs site; there is no English site.
- **`openai-readiness`** will now count `/videos` routes as registered, even though they serve xAI. The OpenAI baseline pin was left unchanged.

## Adding provider N (for example Kling or Veo)

1. **Codec.** In `services/control-plane/internal/proxy/videos_provider.go`, implement `videoProviderCodec` and register it in `NewVideosHandler`'s `codecs` map. The test `TestVideosSecondProviderPlugsInThroughTheCodecOnly` shows the full surface.
2. **Egress.** In `services/silk-road/pkg/server/`, add the provider's video operations to its adapter, plus a content fetch modeled on `xai_video.go`.
3. **Contracts.** In `packages/contracts/platform/types.go`, add the provider's `*_videos_*` operations. If the request shape differs from xAI's, add a parser like `xai_video.go`.
4. **Catalog.** Add `apps/model-discovery/providers/<provider>/<model>.yaml` with a `videoControls` block and `output_cost_per_second`, and add the model to that provider's importer allow-list.
5. **Registry.** Add `CapabilityVideos` to the provider's entry in `providerregistry/registry.go`.

Routes, the job store, the worker, settlement, list/delete, and the docs page do not change.
