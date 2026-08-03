# Support ticketing operations

## Runtime components

Support tickets are authoritative in the main PostgreSQL database. Private
attachment bytes are stored in an S3-compatible bucket and are never returned
as public URLs. The control-plane validates file content, stores the object in a
quarantined state, scans it with ClamAV, and only commits clean attachment
metadata with the ticket/message transaction.

Required production settings:

| Setting | Purpose |
| --- | --- |
| `PRIVATE_FILES_S3_ENDPOINT` | Private S3-compatible endpoint |
| `PRIVATE_FILES_S3_REGION` | Bucket region, defaults to `us-east-1` |
| `PRIVATE_FILES_S3_BUCKET` | Private bucket shared by support attachments and billing receipts |
| `PRIVATE_FILES_S3_ACCESS_KEY` / `PRIVATE_FILES_S3_SECRET_KEY` | Storage credential |
| `PRIVATE_FILES_S3_FORCE_PATH_STYLE` | Enable for MinIO |
| `SUPPORT_CLAMAV_ADDRESS` | ClamAV TCP address, for example `clamav:3310` |
| `SUPPORT_CLAMAV_TIMEOUT` | Scan timeout, defaults to `30s` |
| `SUPPORT_TICKETS_PER_USER_PER_HOUR` | Ticket-create rate limit, defaults to `10` |
| `SUPPORT_MESSAGES_PER_USER_PER_TICKET_PER_HOUR` | Reply rate limit, defaults to `30` |

The legacy `SUPPORT_S3_*` variables remain supported as fallbacks. Support
attachments use the `support/` object prefix. Bank-transfer receipt images use
the isolated `billing/receipts/` prefix; PostgreSQL stores only their private
application URL and related metadata, not the file bytes.

`SUPPORT_ATTACHMENT_ALLOW_UNSCANNED_DEV=true` is an explicit local/test bypass.
It must not be set in production. If ClamAV is absent or fails, attachment
operations fail closed.

Local Compose starts MinIO, creates a non-public `godarai-support` bucket, and
starts ClamAV. Start it with:

```powershell
docker compose -f infra/compose/local.yml up -d minio minio-init clamav
```

Production GitOps deploys the same dependency model in Kubernetes: a persistent
MinIO StatefulSet, an idempotent bucket bootstrap Job, and an internal ClamAV
service. The `PRIVATE_FILES_S3_ACCESS_KEY` and
`PRIVATE_FILES_S3_SECRET_KEY` repository secrets are synchronized into the
production application Secret before the GitOps change is rolled out.

## Multipart contract

Ticket creation and message creation use `multipart/form-data`:

- One `payload` text part containing JSON matching the typed request contract.
- Zero to five repeated `files` file parts.
- An `Idempotency-Key` header, at most 128 characters. The same key and payload
  replay the operation; a different payload with the same scoped key returns
  `409 idempotency_conflict`.

The request body limit is 30 MiB. Per message, each file is limited to 10 MiB
and all files together to 25 MiB.

## Periodic jobs

The repository has no shared scheduler, so both maintenance operations are
idempotent one-shot commands suitable for a Kubernetes CronJob, Nomad periodic
job, or system cron.

Run the resolved-ticket closer at least hourly:

```powershell
go -C services/control-plane run ./cmd/support-ticket-closer
```

Settings:

- `DATABASE_URL` (required)
- `SUPPORT_TICKET_RESOLVED_CLOSE_DAYS` (default `7`)
- `SUPPORT_TICKET_CLOSE_BATCH_SIZE` (default `250`)

Run orphan cleanup daily:

```powershell
go -C services/control-plane run ./cmd/support-ticket-orphan-cleaner
```

The cleaner lists private objects under `support/`, compares them with
PostgreSQL attachment metadata, and deletes only unreferenced objects older
than `SUPPORT_ORPHAN_GRACE_PERIOD` (default `24h`). It uses `DATABASE_URL` and
the S3 settings above. Re-running either command is safe.

## Access and download behavior

Tenant routes are scoped by the authenticated tenant and requester. Cross-tenant
lookups return `404`. Support-admin routes require both the local
`AdminPermission` and the control-plane service-account support-ticket scope.
Downloads are authorized for every request and include:

- `Content-Disposition: attachment`
- `Cache-Control: private, no-store`
- `X-Content-Type-Options: nosniff`

Quarantined, rejected, missing, or scanner-failed files are not downloadable.
Download attempts and outcomes are written to the support-ticket event log.

## Retention

MVP performs no automatic deletion of tickets, messages, ratings, events, or
referenced attachments. This preserves the audit record until legal and
operational retention requirements are approved. The orphan cleaner only
removes unreferenced object-store data and is not a ticket-retention policy.
