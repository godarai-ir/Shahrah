# Shahrah

Shahrah is now organized as a monorepo for the full platform: control plane, tunnel services, admin panels, docs, blog, SDKs, shared contracts, and infrastructure.

## Repository Layout

```text
shahrah/
├─ apps/                  # Next.js, Docusaurus, Astro applications
├─ services/              # Backend deployables
├─ sdks/                  # Client SDKs and generators
├─ packages/              # Shared contracts and reusable packages
├─ db/                    # Main/admin schema, migrations, seed data
├─ infra/                 # Compose, Docker, Kubernetes, Terraform, monitoring
├─ plans/                 # Architecture and product planning docs
├─ go.work                # Go workspace
├─ pnpm-workspace.yaml    # JS workspace
└─ turbo.json             # JS task orchestration
```

## Current Services

- `services/control-plane`: existing Go gateway/control-plane service
- `services/tunnel-client`: Iran-side tunnel client
- `services/tunnel-server`: Europe-side tunnel server
- `services/silk-road`: Europe-side provider proxy
- `services/sync-worker`: background worker scaffold

## Current Apps

- `apps/tenant-admin`: tenant-facing Next.js admin panel
- `apps/support-admin`: internal Next.js support panel
- `apps/docs`: Docusaurus documentation site
- `apps/blog`: Astro blog site

## Quick Start

### Backend Stack

```bash
make setup
make docker-up
make docker-migrate
make test-integration
make test-unit
make run
```

PowerShell:

```powershell
go -C services/control-plane mod download
go -C services/control-plane mod tidy
putty.exe -load node-4-eu -l ubuntu -N -L 1443:127.0.0.1:443
docker compose -f infra/compose/local.yml up -d postgres redis clickhouse silk-road tunnel-server tunnel-client control-plane
docker compose -f infra/compose/local.yml exec postgres psql -U postgres -d llmproxy -f /docker-entrypoint-initdb.d/001_initial.sql
docker compose -f infra/compose/local.yml exec postgres psql -U postgres -d llmproxy -f /docker-entrypoint-initdb.d/002_platform_foundations.sql
go -C services/control-plane test -v -race ./tests/integration/...
go -C services/control-plane test -v -race ./tests/unit/...
go -C services/control-plane run ./cmd/server
```

The local Silk Road service is fail-closed and sends provider traffic, including
credential probes, through `http://host.docker.internal:1443`. Keep the PuTTY
forward above running while using the local stack. It forwards the host listener
to Squid on `node-4-eu`. Required mode prevents Silk Road from silently falling
back to direct egress if a launch configuration does not provide a proxy URL.
Override `SILK_ROAD_EGRESS_PROXY_URL` only when using another trusted development
proxy.

### Persistent MVP Verification

The control-plane now expects a reachable PostgreSQL database and applies tenant, key, deployment, wallet, and usage persistence through `internal/storage.Store`.

Migration order:

1. `db/main/migrations/001_initial.sql`
2. `db/main/migrations/002_platform_foundations.sql`
3. `db/main/migrations/005_wallet_topup_intents.sql`

Required runtime configuration:

- `DATABASE_URL`
- `MASTER_KEY`
- `TUNNEL_CLIENT_URL`
- `OPENAI_API_KEY` only if using the sample static model entries in `services/control-plane/configs/proxy_config.yaml`

Minimal onboarding flow is documented in `services/control-plane/README.md`.

### Workspace Tooling

- Go services are managed via `go.work`
- JS apps/packages are prepared for `pnpm` workspaces + `turbo`
- Database assets live under `db/main` and `db/admin`

## Important Paths

- `services/control-plane/configs/proxy_config.yaml`
- `db/main/migrations/001_initial.sql`
- `db/main/migrations/002_platform_foundations.sql`
- `infra/compose/local.yml`
- `plans/README.md`

## Database Schema Management

Shahrah uses Prisma as the single source of truth for database schema. All three services (`tenant-admin`, `support-admin`, and `control-plane`) share the same PostgreSQL database schema.

### Architecture

```text
db/main/migrations/*.sql     -->  packages/database/schema.prisma  -->  Prisma Clients
        (source of SQL)              (source of truth)              (for apps)
                                                                  |
                                                        control-plane (Go: raw SQL)
```

### Workflow (Goose-First)

1. **Add new migration**: Create a new Goose SQL migration in `db/main/migrations/` (`-- +goose Up/Down`)
2. **Apply migrations locally**: Run `make migrate-up` (or run `node scripts/sync-prisma.js`, which runs Goose first)
3. **Sync Prisma schema/clients**: Run `node scripts/sync-prisma.js`
4. **Review changes**: Check `packages/database/schema.prisma` and generated clients
5. **Commit**: Push migration + schema/client updates together
6. **Deploy**: Apply Goose migrations in target environments (`make migrate-up` or equivalent Goose command)

### Packages

- `packages/database/` - Shared Prisma schema and generated client
- `apps/tenant-admin/` - Tenant admin panel (Next.js + Prisma)
- `apps/support-admin/` - Support admin panel (Next.js + Prisma)

### Key Files

- `packages/database/schema.prisma` - Single source of truth
- `scripts/sync-prisma.js` - Goose-first migration + Prisma sync automation
- `db/main/migrations/` - Goose SQL migrations (source of DB changes)
- `apps/tenant-admin/prisma/schema.prisma` - Tenant admin schema
- `apps/support-admin/prisma/schema.prisma` - Support admin schema

### Troubleshooting

```bash
# Apply pending Goose migrations
make migrate-up

# Regenerate all Prisma clients
node scripts/sync-prisma.js

# Format schema
npx prisma@7.8.0 format --schema=packages/database/schema.prisma

# Generate client
npx prisma@7.8.0 generate --schema=packages/database/schema.prisma
```

## ClickHouse Migrations (Observability)

Use the Windows helper script to apply ClickHouse migrations reliably without container path issues:

```cmd
scripts\clickhouse-migrate.cmd
```

Check status/tables:

```cmd
scripts\clickhouse-migrate.cmd status
```

Source files are under [`infra/clickhouse/migrations`](infra/clickhouse/migrations).

## Next Migration Steps

1. Replace remaining framework starter assets and sample content with Shahrah-specific UI and docs.
2. Add generated protobuf artifacts and stronger tunnel transport semantics on top of the current HTTP-backed local chain.
3. Split shared Go utilities from `services/control-plane/internal` into `packages/go/` only when another service actually needs them.
4. Add per-service CI once new services become real modules.

## License

MIT License. See [LICENSE](LICENSE) for details.
