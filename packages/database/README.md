# @shahrah/database

Centralized database schema for Shahrah platform.

## Source of Truth

This package contains the definitive Prisma schema. All database changes should be reflected here.

## Usage

```bash
# Generate Prisma clients
npm run prisma:generate

# Apply migrations
npm run prisma:migrate

# Development migration
npm run prisma:migrate:dev

# Pull schema from existing database
npm run prisma:pull

# Open Prisma Studio
npm run prisma:studio
```

## Schema Models

- **Budget** - Budget and rate limits
- **Tenant** - Multi-tenant support
- **Organization** - Organization hierarchy
- **Team** - Team within organizations
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