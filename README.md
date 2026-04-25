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
- `services/tunnel-client`: Iran-side tunnel client scaffold
- `services/tunnel-server`: Europe-side tunnel server scaffold
- `services/silk-road`: provider proxy scaffold
- `services/sync-worker`: background worker scaffold

## Current Apps

- `apps/tenant-admin`: tenant-facing admin panel scaffold
- `apps/support-admin`: internal operations panel scaffold
- `apps/docs`: Docusaurus docs scaffold
- `apps/blog`: Astro blog scaffold

## Quick Start

### Control Plane

```bash
make setup
make docker-up
make docker-migrate
make test-unit
make run
```

PowerShell:

```powershell
go -C services/control-plane mod download
go -C services/control-plane mod tidy
docker compose -f infra/compose/local.yml up -d postgres redis
docker compose -f infra/compose/local.yml exec postgres psql -U postgres -d llmproxy -f /docker-entrypoint-initdb.d/001_initial.sql
go -C services/control-plane test -v -race ./tests/unit/...
go -C services/control-plane run ./cmd/server
```

### Workspace Tooling

- Go services are managed via `go.work`
- JS apps/packages are prepared for `pnpm` workspaces + `turbo`
- Database assets live under `db/main` and `db/admin`

## Important Paths

- `services/control-plane/configs/proxy_config.yaml`
- `db/main/migrations/001_initial.sql`
- `infra/compose/local.yml`
- `plans/README.md`

## Next Migration Steps

1. Build `apps/tenant-admin` and `apps/support-admin` as real Next.js apps.
2. Add `proto` and OpenAPI sources under `packages/contracts/`.
3. Split shared Go utilities from `services/control-plane/internal` into `packages/go/` only when another service actually needs them.
4. Add per-service CI once new services become real modules.

## License

MIT License. See [LICENSE](LICENSE) for details.
