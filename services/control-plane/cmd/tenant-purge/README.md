# Tenant purge

`tenant-purge` permanently removes one tenant and its tenant-scoped data from:

- the main PostgreSQL database;
- ClickHouse observability tables;
- the support-admin PostgreSQL tenant projection.

The command discovers PostgreSQL and ClickHouse tables containing a
`tenant_id` column at runtime, so newly added tenant-scoped tables are included.
The main PostgreSQL deletion runs in one transaction and retries tables in
foreign-key-safe order.

## Safety requirements

1. Take and verify database backups.
2. Suspend the tenant and stop every workload that can write data for it.
3. Set `DATABASE_URL`, `CLICKHOUSE_DSN`, and
   `SUPPORT_ADMIN_DATABASE_URL` for the target environment.
4. Run a dry-run and inspect every reported table/count.
5. Execute only after confirming the resolved UUID, slug, domain, and status.

Do not run the command while gateway or administrative requests for the tenant
can still be accepted. PostgreSQL and ClickHouse cannot share one distributed
transaction. The command therefore purges ClickHouse and the support projection
first, and deletes the main PostgreSQL tenant last.

## Usage

Run from `services/control-plane`.

Dry-run by exact UUID, slug, or domain:

```powershell
$env:DATABASE_URL = "postgresql://..."
$env:CLICKHOUSE_DSN = "clickhouse://..."
$env:SUPPORT_ADMIN_DATABASE_URL = "postgresql://..."
go run ./cmd/tenant-purge --tenant acme
```

Permanent deletion requires the resolved slug as a second, independent
confirmation:

```powershell
go run ./cmd/tenant-purge --tenant acme --execute --confirm acme
```

An active tenant is rejected by default. If all writers have already been
stopped and deleting it while its database status is still `active` is
intentional:

```powershell
go run ./cmd/tenant-purge --tenant acme --execute --confirm acme --allow-active
```

If an auxiliary database is not deployed, it must be skipped explicitly:

```powershell
go run ./cmd/tenant-purge --tenant acme --execute --confirm acme --skip-clickhouse --skip-support-db
```

The final warning will state that such a purge is not complete across all
Shahrah databases.
