# Main Database

Schema, migrations, and seed data for the main Shahrah platform database.

## Migration Tooling

This project uses [pressly/goose](https://github.com/pressly/goose) for production-ready database migrations.

### Directory Structure

```
db/main/migrations-new/
├── 001_initial.sql              # Create + drop base schema
├── 002_platform_foundations.sql # Add tenants, wallets + rollback
└── ...
```

### Quick Start

#### Install CLI

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

#### Run Migrations

```bash
# Using Makefile (recommended)
make migrate-up           # Apply all pending migrations
make migrate-down         # Rollback last migration
make migrate-status       # Show current version + status
make migrate-reset        # Drop all migrations
make migrate-create name=add_column  # Create new migration

# Using the CI/CD script
./scripts/migrate.sh up
./scripts/migrate.sh down
./scripts/migrate.sh status

# Using Docker Compose
docker compose -f infra/compose/local.yml run --rm migrate
```

#### Manual CLI

```bash
cd db/main/migrations-new
goose postgres "postgres://postgres:postgres@localhost:5432/sharah_db?sslmode=disable" up
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | `localhost` | Database hostname |
| `DB_PORT` | `5432` | Database port |
| `DB_NAME` | `sharah_db` | Database name |
| `DB_USER` | `postgres` | Database username |
| `DB_PASSWORD` | `postgres` | Database password |
| `DB_SSLMODE` | `disable` | SSL mode (`disable`, `require`, `verify-full`) |

### Creating New Migrations

```bash
# Makefile helper
make migrate-create name=add_user_avatar

# Manual
cd db/main/migrations-new
goose postgres "postgres://..." create add_user_avatar sql
```

This creates a file like `015_add_user_avatar.sql` with:

```sql
-- +goose Up
CREATE TABLE ...

-- +goose Down
DROP TABLE ...
```

### CI/CD Integration

In CI/CD pipelines, run migrations before starting the application:

```yaml
# Example GitHub Actions
- name: Run migrations
  env:
    DB_HOST: ${{ secrets.DB_HOST }}
    DB_PASSWORD: ${{ secrets.DB_PASSWORD }}
  run: ./scripts/migrate.sh up
```

### Migration Status

Check the current migration state:

```bash
make migrate-status
```

### How Goose Tracks Migrations

Goose automatically manages a `goose_db_version` table in the database:

```sql
CREATE TABLE goose_db_version (
    id SERIAL PRIMARY KEY,
    version_id BIGINT NOT NULL,
    is_applied BOOLEAN NOT NULL,
    timestamptz TIMESTAMP NOT NULL
);
```

### Migration File Format

Each `.sql` file contains both up and down sections:

```sql
-- +goose Up
-- SQL to apply migration

-- +goose Down
-- SQL to rollback migration
```
