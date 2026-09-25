# Shahrah single-host demo

A low-cost replacement for the K3s production cluster, intended **only for demos**.
It runs every product workload on one VPS with Docker Compose and pulls the same
images that production deploys. There is no high availability: if the host goes
down, the demo goes down.

## What runs

| Service | Purpose | Memory limit |
|---|---|---|
| caddy | TLS termination and routing (mirrors the production ingress) | 128 MB |
| postgres | `sharah_db`, `model_discovery`, `support_admin` | 1 GB |
| redis | runtime cache, usage counters | 320 MB |
| clickhouse | usage dashboards and request traces (single node + embedded Keeper) | 1.5 GB |
| typesense | model catalog search | 512 MB |
| minio | private support attachments | 384 MB |
| control-plane, silk-road | gateway API and provider egress | 768 MB + 256 MB |
| model-discovery, model-discovery-web | catalog API and `/catalog` UI | 512 MB + 384 MB |
| tenant-admin | tenant workspaces on `<tenant>.DOMAIN` | 512 MB |
| website | landing page and `auth.DOMAIN` | 512 MB |
| support-admin | `admin.DOMAIN` | 384 MB |

Optional profiles:

- `marketing`: `blog` and `docs` (+512 MB). Without it, `/blog` and `/docs` return 502.
- `clamav`: malware scanning for support attachments (+1.5 GB).

What production has that this stack drops: K3s, Argo CD, Longhorn, the replicated
ClickHouse cluster and standalone Keeper, Redis Sentinel and the metrics DLQ,
HAProxy/keepalived, the observability stack, and the managed Postgres instance.

## Host sizing

- **Recommended:** 4 vCPU, 8 GB RAM, 80 GB SSD. The limits above add up to about 7 GB.
- **With ClamAV:** 12 GB RAM.
- Ubuntu 24.04 LTS with Docker Engine and the Compose plugin. Enable a 2 GB swap file as a safety net.
- Open only ports 22, 80 and 443 in the firewall.

## First start

```bash
git clone <repo> /opt/shahrah && cd /opt/shahrah/infra/compose/demo
scripts/init-env.sh              # creates .env with fresh secrets
$EDITOR .env                     # set DEMO_DOMAIN, email, OAuth, egress proxy, and similar settings
docker login docker.io           # the godar-ai repository is private
```

### TLS certificate

Tenants live on arbitrary subdomains, so the stack uses a wildcard certificate
for `DOMAIN` and `*.DOMAIN`. Put it in `certs/`:

```text
certs/fullchain.pem
certs/privkey.pem
```

The production wildcard certificate (`WILDCARD_GODARAI_IR_TLS_CRT` /
`WILDCARD_GODARAI_IR_TLS_KEY` in the GitHub secrets) can be reused. It must be
renewed manually before it expires; run `docker compose restart caddy` after
replacing the files.

### DNS

Point `DOMAIN` and `*.DOMAIN` (A records) at the VPS.

### Start

```bash
docker compose pull
docker compose up -d
docker compose ps
```

The one-shot `migrate-*` and `minio-init` containers must exit with code 0.
Inspect them with `docker compose logs migrate-clickhouse`, for example.

### Service tokens for support-admin

support-admin needs two service credentials. Create them once the stack is up,
put them in `.env`, and run `docker compose up -d support-admin`.

```bash
# MODEL_DISCOVERY_SERVICE_TOKEN
docker compose run --rm migrate-model-discovery \
  pnpm run service-account:create -- --name=support-admin \
  --scopes=catalog:read,catalog:write,provider:write,import:execute,indexing:execute,display-settings:write

# CONTROL_PLANE_SERVICE_TOKEN (the tool is not shipped in the image; run it from source)
set -a; source .env; set +a
docker run --rm --network shahrah-demo_internal \
  -v /opt/shahrah:/src -w /src/services/control-plane \
  -e GOPROXY=https://goproxy.io,direct \
  -e DATABASE_URL="postgresql://postgres:${POSTGRES_PASSWORD}@postgres:5432/sharah_db?sslmode=disable" \
  golang:1.26 go run ./cmd/platform-service-account -name support-admin -profile support-admin
```

## Moving data from production (optional)

For a clean demo, skip this step and create demo tenants by hand. To keep
existing tenants, restore dumps of `sharah_db`, `model_discovery` and
`support_admin` **before** the first `docker compose up`:

```bash
docker compose up -d postgres
docker compose exec -T postgres pg_restore -U postgres -d sharah_db --no-owner < sharah_db.dump
# repeat for model_discovery and support_admin
docker compose up -d
```

Encrypted provider credentials only decrypt if the `*_PROVIDER_CREDENTIAL_KEKS_JSON`,
`*_FINGERPRINT_KEY` and `*_ACTIVE_KEK_VERSION` values are copied from production
instead of generated. ClickHouse history is not migrated.

## Provider egress

If the VPS cannot reach provider APIs directly (for example, a host in Iran),
set `SILK_ROAD_EGRESS_PROXY_URL` to an HTTP CONNECT proxy and
`SILK_ROAD_EGRESS_PROXY_REQUIRED=true`. See `infra/eu-egress-proxy`. A VPS
outside Iran needs neither setting.

## Support attachments

Uploads are stored in MinIO. With ClamAV disabled (the default),
`SUPPORT_ATTACHMENT_ALLOW_UNSCANNED=true` accepts files without a malware scan.
To enable scanning:

```bash
# .env: SUPPORT_CLAMAV_ADDRESS=clamav:3310, SUPPORT_ATTACHMENT_ALLOW_UNSCANNED=false
docker compose --profile clamav up -d
```

## Backups

`scripts/backup.sh` dumps the three Postgres databases and copies the
attachment bucket to `backups/<timestamp>/`. It uploads to S3-compatible
storage when `BACKUP_S3_*` is set, and prunes local copies after
`BACKUP_LOCAL_RETENTION_DAYS`.

```bash
crontab -e
# 30 3 * * * /opt/shahrah/infra/compose/demo/scripts/backup.sh >> /var/log/shahrah-backup.log 2>&1
```

## Upgrading

Change the `*_TAG` values in `.env`, then run:

```bash
docker compose pull && docker compose up -d
```

The migration containers re-run automatically before the workloads that depend on them.
