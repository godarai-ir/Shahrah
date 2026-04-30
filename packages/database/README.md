# @shahrah/database

Centralized database schema for Shahrah platform.

## Source of Truth

This package contains the definitive Prisma schema. All database changes should be reflected here.

## Usage

```bash
# Generate Prisma clients
pnpm prisma:generate

# Apply migrations
pnpm prisma:migrate

# Development migration
pnpm prisma:migrate:dev

# Pull schema from existing database
pnpm prisma:pull

# Open Prisma Studio
pnpm prisma:studio
```

Set `DATABASE_URL` before running commands that talk to Postgres (`prisma:migrate`, `prisma:migrate:dev`, `prisma:pull`, `prisma:studio`).

## Schema Models

- **Budget** - Budget and rate limits
- **Tenant** - Multi-tenant support
- **Organization** - Organization hierarchy
- **VirtualUser** - Service identities for tenant/team automation
- **Team** - Team within organizations
- **TeamMembership** - Explicit scoped user-to-team relation
- **Project** - Project within teams
- **User** - Internal users
- **VerificationToken** - API keys
- **EndUser** - Consumer tracking
- **ModelDeployment** - Provider configurations
- **Credential** - Secure API key storage
- **RequestLog** - Audit logs
- **SpendLog** - Usage tracking
- **Wallet** - Tenant wallets
- **WalletTransaction** - Wallet transactions
- **Policy** - Budget/rate limit policies
- **Guardrail** - Input/output guardrails
