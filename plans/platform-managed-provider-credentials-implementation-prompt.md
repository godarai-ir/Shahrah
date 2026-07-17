# Executable Prompt: Implement Platform-Managed Provider Credentials

Use the prompt below from the Shahrah repository root. It is intentionally implementation-oriented: complete the code, migrations, UI, tests, and rollout controls rather than returning another plan.

---

You are working in `D:\source\Shahrah`. Implement the Platform-Managed Provider Credentials MVP described in `plans/platform-managed-provider-credentials-product-analysis.md` end to end.

Do not stop after analysis or scaffolding. Inspect the current worktree, preserve unrelated user changes, implement the feature, run the relevant verification, and report the exact files changed, tests run, rollout controls, and any remaining debt. Use `pnpm`, never `npm`.

## Read First

Read these inputs before changing code, in this order:

1. All applicable `AGENTS.md` files, especially the root instructions, `services/control-plane/AGENTS.md`, `services/silk-road/AGENTS.md`, and `apps/support-admin/AGENTS.md`.
2. `.agents/skills/truefoundry-expert/SKILL.md` and `.agents/skills/truefoundry-expert/references/output-templates.md`.
3. Relevant TrueFoundry evidence:
   - `.truefoundry-docs/markdown/docs__ai-gateway__secret-manager-in-ai-gateway.md`
   - `.truefoundry-docs/markdown/docs__ai-gateway__gateway-architecture.md`
   - `.truefoundry-docs/markdown/docs__ai-gateway__model-discovery.md`
   - `.truefoundry-docs/markdown/docs__ai-gateway__azure-openai.md`
4. GodarAI constraints and overlapping plans:
   - `plans/README.md`
   - `plans/10_performance_availability.md`
   - `plans/04_microservices_architecture.md`
   - `plans/phase0-tenant-provider-account-contract.md`
   - `plans/phase1-tenant-provider-account-backend-foundation.md`
   - `plans/provider-account-models-ux-flow.md`
   - `plans/truefoundry-provider-account-management-reference.md`
   - `plans/platform-managed-provider-credentials-product-analysis.md`
5. Before touching UI, read:
   - `docs/design-system/DESIGN.md`
   - `docs/design-system/COMPONENT_USAGE.md`
   - `docs/design-system/RTL_GUIDELINES.md`
   - `docs/design-system/UI_REVIEW_CHECKLIST.md`
   - inspect reusable components in `apps/support-admin/components/ui` and `apps/support-admin/components/app-ui`.

Treat TrueFoundry as product evidence, not an implementation blueprint. Build a focused Go-native GodarAI feature.

## Objective and MVP Cutline

Replace deployment-wide provider secrets from `OPENAI_API_KEY` and `ANTHROPIC_API_KEY` with audited, encrypted, versioned Platform Provider Connections managed in `support-admin` and consumed by the Go control-plane without restart.

The MVP must be production-capable for OpenAI and Anthropic API-key authentication. Make the schema, storage, and typed runtime contract extensible for Azure, Bedrock, and Vertex, but do not pretend their current adapters support Entra, SigV4, AssumeRole, or service-account auth if they do not. Clearly mark those runtime auth modes unsupported until implemented.

Include now:

- immediate containment of secret-like values in `model-discovery`;
- multiple connections in the data model, with one explicit active primary per provider and environment in the MVP;
- provider/auth-mode schema definitions with no connection values in `model-discovery`;
- encrypted and versioned secret storage owned by `control-plane`;
- create, test, activate, rollback, disable, health/status, RBAC, and audit workflows;
- dynamic schema-driven `support-admin` UI;
- explicit BYOK versus platform-managed runtime resolution and billing metadata;
- controlled migration from environment credentials;
- a process-local credential cache whose hot path performs no PostgreSQL query, no Redis call, no decryption, and no network I/O.

Defer weighted multi-credential routing, external secret-manager integrations, auto-rotation, two-person approval, and egress-side secret resolution. Define small interfaces now where they prevent later lock-in.

## Non-Negotiable Architecture Boundaries

1. `model-discovery` owns provider definitions and versioned connection field schemas only. It must not store a real endpoint/account value, API key, client secret, AWS secret, session token, service-account JSON, ciphertext, or secret reference that can resolve to material.
2. `control-plane` owns connection metadata, runtime configuration, lifecycle, secret versions, encryption, testing, activation, rollback, runtime resolution, and audit events.
3. `support-admin` is a server-mediated control surface. The browser must call its App Router handlers, and those handlers must call scoped internal service APIs. The browser must never call `control-plane` or `model-discovery` directly with service credentials.
4. `silk-road` remains stateless. For this MVP, `control-plane` may place resolved provider auth in the existing mTLS tunnel dispatch only for the lifetime of that request. Do not persist the secret in Europe.
5. Never return an existing secret to the browser. Read APIs return only metadata such as `configured`, a safe masked hint, version, timestamps, status, and test result. Empty secret input on edit means unchanged; revocation is a separate confirmed action.
6. Never place plaintext or decrypted provider secrets in Redis, distributed runtime profiles, logs, traces, audit JSON, analytics, metrics labels, error messages, panic output, or generic JSON config columns.
7. Do not silently fall back from a failing tenant BYOK credential to a platform credential. That changes who pays. Only an explicit operator rollout mode may fall back from the new platform store to the legacy environment key.

## First Deliverable: Contain the Current Leak Path

Fix the current `ProviderConnectionProfile.configJson` path before adding the new workflow:

- Remove connection profiles containing values from public provider list/detail responses.
- Replace value-bearing connection profiles with versioned schema definitions. Prefer a `ProviderConnectionSchema` model that supports multiple auth modes and schema versions per provider. Schema field definitions may include `key`, localized label/help text, type, `secret`, required/conditional-required rules, validation, `scope`, runtime mapping, documentation URL, and redaction policy; they must contain no connection values.
- Remove or disable the current support-admin route that sends `configJson` secrets to `model-discovery`.
- Recursively reject secret-shaped keys and secret references on every model-discovery schema write. At minimum cover `api_key`, `apiKey`, `client_secret`, `clientSecret`, `secretAccessKey`, `sessionToken`, `serviceAccountJson`, `credential`, `token`, `password`, and equivalent case/format variants. Return a structured 422 error without echoing the submitted value.
- Add a redacted audit utility that reports only affected row IDs/provider slugs and key names for any existing suspicious `ProviderConnectionProfile.configJson` data. Never print values. Do not delete production data or claim credentials were rotated; document the operator rotation step.
- Regenerate model-discovery Prisma output and OpenAPI artifacts, and update the control-plane contract assertions when the upstream contract changes.

Seed safe versioned schemas for:

- OpenAI `api_key`: secret `api_key`; optional connection-scoped `base_url`, `organization`, and `project`.
- Anthropic `api_key`: secret `api_key`; optional connection-scoped `base_url` and API version.

Model IDs, Azure deployment names, Bedrock model IDs/inference-profile ARNs, and similar deployment-specific fields do not belong in a credential bundle.

## Control-Plane Data Model

Add a new append-only migration after the current highest `db/main/migrations` version and update `packages/database/schema.prisma`. Do not edit applied migrations.

Implement at least these entities, constraints, and indexes:

### `platform_provider_connections`

- UUID `connection_id` primary key;
- normalized `provider_slug`, unique operational `name`, `environment`, `auth_mode`, and positive `schema_version`;
- non-secret `config_json` only;
- lifecycle `status`: `draft`, `testing`, `active`, `degraded`, or `disabled`;
- optional `region`, model/capability scope support, `priority`, `weight`, and explicit `is_primary`;
- nullable `active_credential_version_id`;
- last test status/category/timestamps, last success, last rotation;
- optimistic-lock `revision` incremented on mutation;
- `created_by`, `updated_by`, and timestamps;
- indexes for provider/environment/status and primary selection;
- database enforcement that only one active primary is selected for a provider/environment in the MVP, without preventing multiple draft/disabled connections.

### `platform_provider_credential_versions`

- UUID `credential_version_id`, `connection_id`, and monotonically increasing `version` unique per connection;
- `secret_backend`, initially `encrypted_db`;
- binary ciphertext, data nonce, wrapped DEK, wrap nonce, and `kek_version`; never plaintext JSON;
- keyed, non-reversible fingerprint and safe display hint;
- status `draft`, `validated`, `active`, `retiring`, or `revoked`;
- `valid_from`, optional `expires_at`, optional `retire_at`;
- creator/activator/revoker metadata and timestamps;
- optional idempotency key/request hash with constraints that reject reuse for a different payload.

### `platform_provider_connection_scopes`

- connection ID, model/deployment pattern, capability/region constraints, and enabled state;
- indexes used by connection selection.

### `platform_provider_credential_events`

- append-only actor ID/role, trusted service-account identity where available, request/correlation ID, action, connection/version IDs, result, redacted reason/metadata, timestamp, and bounded source/session metadata;
- idempotency uniqueness where appropriate;
- no secret, ciphertext, fingerprint, auth header, raw upstream body, or raw submitted config in event payloads.

Use transactions and row locks for version allocation, activation, rollback, disable, and competing updates. Activation must atomically retire the previous active version, activate the validated version, set the connection primary/active state, advance `revision`, and append the event.

## Secret Storage and Cryptography

Create small, testable interfaces such as `SecretStore` and `KeyEncryptionKeyProvider`; ship an `encrypted_db` implementation now.

Implement envelope encryption with Go standard cryptography:

- a fresh cryptographically random 256-bit DEK for every credential version;
- AES-256-GCM authenticated encryption with a fresh nonce;
- wrap the DEK with a versioned KEK that lives outside PostgreSQL;
- bind provider slug, connection ID, credential version ID/version, auth mode, and schema version as authenticated additional data so ciphertext cannot be moved between records;
- store only ciphertext and encryption metadata in PostgreSQL;
- compute the duplicate-detection fingerprint with a domain-separated keyed HMAC, never a raw hash of the secret;
- use constant-time comparisons where relevant;
- return stable sanitized errors such as `credential_decrypt_failed`, not raw crypto/provider details.

Provide an environment-backed, versioned KEK provider for the MVP, for example a JSON map of KEK versions plus an active version. Validate key length/encoding at startup. Environment configuration is only the bootstrap trust/key-encryption material; it must not contain per-provider credentials after cutover. Keep the interface ready for KMS/Vault later.

Unit-test round trips, random nonce/DEK uniqueness, AAD tamper detection, wrong KEK version, malformed ciphertext, versioned KEK reads, fingerprint stability, and proof that plaintext does not appear in persisted bytes or marshalled API responses.

## High-Throughput Runtime Cache: Mandatory Design

GodarAI targets 10,000+ requests/second per control-plane instance. Implement a dedicated `PlatformCredentialManager`/`PlatformCredentialResolver`; do not put decrypted secrets into the generic `cache.Cache`/`DualCache`, because it writes values to Redis.

The required behavior is:

- Keep decrypted active credentials only in a bounded process-local L1 snapshot.
- Use an immutable map and atomic snapshot swap (or an equivalently contention-free design) keyed by provider/environment and selected connection/version. Resolution on a warm hit must be a local memory read only.
- `Resolve` must never query PostgreSQL, read Redis, decrypt, or perform provider/network I/O. A missing/expired/disabled entry returns a typed unavailable error and fails the target before upstream dispatch.
- Pre-warm active OpenAI/Anthropic credentials at process startup and reload them in background. A platform-store rollout mode must not be marked ready until its initial credential snapshot load has completed; absence of a credential makes only the affected targets unavailable, not the whole gateway.
- Collapse concurrent loads/reloads with `singleflight`. Never allow a cold burst to cause one DB query/decrypt per request.
- On create-version/test there is no runtime invalidation. On activate/rollback/disable/revoke, update the local snapshot after the database transaction, then publish a redacted invalidation event containing only connection/provider/environment/version/action metadata.
- Extend the Redis coordination abstraction with subscription support if appropriate. All instances must consume invalidation events and reload the affected entry outside the request path. Because Pub/Sub is not durable, also run a jittered reconciliation loop no slower than every 30 seconds that compares non-secret active-version metadata and reloads only changed entries. A missed event must converge within 60 seconds.
- When Redis is unavailable, local mutations still update the current node and the background reconciler may use a bounded database metadata query. Do not move that check into request processing.
- Disable/revoke is fail-closed. Once observed, remove the entry immediately. Do not serve a known-disabled version as stale. For control-plane/cache outages, document and test a bounded stale policy that never exceeds secret expiry and guarantees an unobserved lifecycle change converges in under 60 seconds.
- Keep the previous version stored for rollback/grace, but do not select it for new requests after activation.
- Avoid unnecessary secret copies. Keep cache retention bounded, clear evicted byte buffers on a best-effort basis, and do not claim Go can guarantee memory zeroization.

Add metrics without secret or high-cardinality labels:

- cache resolves by `hit`, `unavailable`, and `expired`;
- snapshot load/reload success/failure and duration;
- invalidation/reconciliation lag;
- active snapshot age and entry count;
- decrypt failure and provider authentication-test category counts.

Add a focused benchmark and tests that prove:

1. after pre-warm, at least 10,000 parallel/sequential `Resolve` calls cause exactly zero repository calls, zero Redis calls, and zero decrypt calls;
2. concurrent reloads collapse to one repository/decrypt operation;
3. activation swaps to the new credential/version atomically;
4. disable removes availability immediately on the writer and within 60 seconds on another simulated node;
5. no plaintext is written to a fake distributed cache or invalidation event;
6. the code passes `go test -race` for the new cache/manager packages;
7. `BenchmarkPlatformCredentialResolve` reports the warm-hit path separately from reload work.

## Typed Runtime Contract and Routing

Introduce a typed `ResolvedProviderAuth` contract instead of expanding untyped maps. For the MVP it must represent API-key auth while leaving explicit variants/interfaces for future Azure Entra, AWS SigV4/AssumeRole, and Google service-account/ADC support. Validate that exactly the fields required for the selected auth mode are present.

Update routing and tunnel metadata so every selected deployment carries non-secret:

- `credential_source`;
- `connection_id` when platform-managed;
- `credential_version_id` when platform-managed;
- `billing_mode`.

Keep secrets out of `RequestMetadata`; only the typed tunnel auth field may carry secret material over the existing mTLS dispatch, and `silk-road` must consume it without logging or persistence. Preserve temporary backward compatibility with the existing `Deployment.APIKey` only as needed for a staged cutover, then centralize all auth injection in one place so handlers cannot accidentally serialize it to logs.

Apply this resolution rule consistently to every gateway endpoint that uses `router.Route`, not only chat completions:

1. A valid tenant provider account with BYOK uses only BYOK, sets `credential_source=user_byok`, and uses `external_billed`.
2. A provider account without BYOK may use the active platform connection only when its account/policy allows platform-managed usage. Set `credential_source=platform_managed` and `wallet_billed`.
3. No active, healthy, in-scope connection/version means the target is unavailable before tunnel/upstream contact.
4. A rejected/expired BYOK key never silently falls through to the platform key.

Add compatibility handling for historical `godar_managed_key` metadata where needed, but emit the new explicit source for new requests. Include source/connection/version in traces, spend/billing records, and operator diagnostics without including secret material.

## Internal APIs

Implement and document the platform-admin APIs below under the existing authenticated `/api/v1/platform` group. Use consistent JSON error envelopes, request IDs, optimistic locking, and `Idempotency-Key` for create/activate operations.

- `GET /provider-connection-schemas/:provider`
- `GET /provider-connections`
- `GET /provider-connections/:id`
- `POST /provider-connections`
- `PATCH /provider-connections/:id` for metadata/non-secret config only
- `POST /provider-connections/:id/credential-versions`
- `POST /provider-connections/:id/credential-versions/:version/test`
- `POST /provider-connections/:id/credential-versions/:version/activate`
- `POST /provider-connections/:id/rollback`
- `POST /provider-connections/:id/disable`
- add an explicit revoke-version endpoint if required to keep revoke separate from edit/disable.

Use status codes deliberately: 400/422 validation, 401 authentication, 403 authorization, 404 unknown connection/version, 409 revision/idempotency/lifecycle conflict, 424 or a documented 422 for provider credential test failure, and 503 for unavailable secret backend/cache initialization. Never echo submitted secrets in an error.

Credential tests must be timeout-bound, rate-limited, and injectable in tests. Reuse and improve the existing provider API-base allowlist/SSRF protection and credential-validation helpers. Prefer a low-cost authentication probe; do not require a paid completion. Map results to stable categories: `invalid_credential`, `expired`, `forbidden`, `quota_exhausted`, `region_mismatch`, `deployment_not_found`, and `network_error`. Persist only the category, sanitized summary, and timestamps.

## RBAC and Audit

Do not reuse broad `provider:write` for secret operations.

Add these human permissions in `support-admin`:

- `platform_credentials:read_metadata`
- `platform_credentials:create_version`
- `platform_credentials:test`
- `platform_credentials:activate`
- `platform_credentials:disable`

Add a `credential_admin` role. `super_admin` and `credential_admin` receive the mutation permissions. `support_agent` may read metadata/health only. No role can reveal an existing secret. Keep catalog/provider permissions separate.

Add scoped control-plane service-account permissions for credential metadata read and credential mutation, update `RequiredPlatformScope`, its tests, the provisioning docs/command examples, and the support-admin BFF path-to-permission mapping. Preserve same-origin checks on mutations and propagate trusted actor/request IDs. Enforce authorization in both the support-admin server route and the control-plane service-account boundary.

Every create-version, test, activate, rollback, disable, revoke, rejected transition, and decrypt/reload failure must produce a redacted audit event where appropriate. Reads should not create noisy per-request audit rows.

## Support-Admin UX

Replace the value-bearing provider connection editor on `/admin/providers` with a dedicated schema-driven Platform Connections workflow backed by the new control-plane endpoints. Keep provider catalog editing distinct from platform connection management.

The workflow must support:

- list/filter by provider, environment, region, auth mode, and status;
- connection name, environment, non-secret config, scope, primary state, status, active version, last rotation, and last test;
- create draft connection;
- create a new credential version without displaying the old value;
- test before activation;
- separate activate, rollback, disable, and revoke confirmations;
- actionable categorized errors and sanitized technical details;
- clear BYOK versus platform-managed billing/source copy;
- loading, empty, error, success, and degraded states;
- RTL, keyboard access, visible focus, and localized strings consistent with the app.

Build the form from the provider/auth-mode schema; do not hardcode provider fields in the page. Reuse existing shadcn/app-ui components and semantic Tailwind tokens. Do not use raw palette utilities or arbitrary values in new/touched UI. Use established `PageContainer`, `SettingsCard`/`AppCard`, `DataListPanel`/table patterns, `StatusBadge`, `OperationNotice`, `ConfirmDialog`, and `DateTimeValue` where appropriate. Add a design-system playground example only if you introduce a genuinely reusable visual pattern.

Secret inputs must use `autocomplete="new-password"`, must not be initialized from API data, and must be cleared after submission. Do not store them in URL/search params, local/session storage, analytics, toasts, React query cache, or long-lived component state. Add browser/server tests proving GET payloads and rendered HTML contain no submitted secret.

## Migration and Rollout

Implement these explicit source modes:

- `env`: current behavior for rollback during initial deployment;
- `platform_store`: strict new source, with no environment fallback;
- `platform_store_with_env_break_glass`: platform store first and an explicit, observable operator-approved environment fallback only when no usable platform connection exists.

Default existing installations to `env` until an operator opts in. Emit non-secret metrics/logs whenever break-glass is used. Never use break-glass after a BYOK authentication failure.

Provide a runbook and configuration documentation covering:

1. create and test draft OpenAI/Anthropic connections;
2. activate and verify cache propagation without restart;
3. switch one provider at a time to shadow validation, then platform store;
4. compare auth failure, billing source, cache, and provider metrics;
5. rollback the credential version or source mode;
6. rotate any credentials found in legacy model-discovery JSON;
7. remove `OPENAI_API_KEY`/`ANTHROPIC_API_KEY` after stabilization while retaining only KEK/service bootstrap credentials;
8. KEK rotation, backup restore, and disaster recovery.

Do not automatically delete legacy environment variables in the first change. Do not silently continue reading `model_deployments.provider_config.api_key` or generic `credentials.credential_values` as a new platform source.

## Verification

Add unit, integration, API contract, migration, security, cache, concurrency, and UI tests. At minimum verify:

- all lifecycle transitions and invalid transitions;
- optimistic locking and idempotency conflicts;
- only validated versions activate;
- atomic rollback and single active primary behavior;
- OpenAI and Anthropic probe classification with fake HTTP servers;
- SSRF rejection for loopback, private/link-local IPs, userinfo, non-HTTPS, redirects to disallowed hosts, and disallowed DNS results;
- read/create/test/activate/disable permissions for every support role and service scope;
- recursive secret-key rejection and redaction in model-discovery;
- no secret in GET responses, OpenAPI examples, logs, traces, audits, metrics, Redis/fake distributed cache, runtime profile L2, error strings, snapshots, or rendered support-admin output;
- explicit BYOK/platform source and billing behavior, including proof that BYOK failure does not fall through;
- all proxy endpoint families receive the same runtime credential behavior;
- cache requirements and benchmark listed above;
- activation becomes visible without restart and converges across simulated instances in under 60 seconds.

Run the narrow tests while iterating, then the broad relevant checks. Use commands compatible with the repository, including at least:

```powershell
gofmt -w <changed-go-files>
go test ./services/control-plane/...
go test ./services/silk-road/...
go test ./packages/contracts/...
go test -race ./services/control-plane/internal/<new-credential-package>/...
go test -run '^$' -bench BenchmarkPlatformCredentialResolve -benchmem ./services/control-plane/internal/<new-credential-package>/...
pnpm --filter model-discovery test
pnpm --filter model-discovery build
pnpm --filter @shahrah/support-admin lint
pnpm --filter @shahrah/support-admin build
```

Also regenerate and validate Prisma/OpenAPI artifacts using existing repository scripts. If a broad command fails because of a pre-existing unrelated issue, prove that with a narrower command and report it precisely; do not hide it.

## Completion Standard

The task is complete only when:

- OpenAI and Anthropic platform credentials can be created, encrypted, tested, activated, resolved, rotated, rolled back, and disabled through support-admin;
- no existing secret can be read back;
- model-discovery no longer accepts or returns connection values;
- request processing uses the process-local snapshot and performs zero DB/Redis/decrypt work on warm credential resolution;
- activation/rollback/disable converge across instances in under 60 seconds without restart;
- billing/tracing identify BYOK versus platform-managed source and version safely;
- rollout modes allow a controlled provider-by-provider migration;
- the relevant tests, race checks, builds, and benchmark pass.

In the final report, separate:

1. documented facts and source files inspected;
2. implementation decisions made for GodarAI;
3. assumptions;
4. files changed and migrations added;
5. verification results and benchmark numbers;
6. rollout/rollback instructions;
7. remaining deferred work and UI/technical debt.

List the TrueFoundry documentation files used, state how `plans/README.md`, the performance plan, and the product analysis changed the implementation, and do not claim TrueFoundry parity.
