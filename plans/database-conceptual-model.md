# Conceptual Entity Model (from `packages/database/schema.prisma`)

```mermaid
erDiagram
  %% ---------------- Core Tenant / Billing ----------------
  Tenant {
    uuid tenant_id PK
    string slug UK
    string domain UK
    string status
  }

  Wallet {
    uuid wallet_id PK
    uuid tenant_id UK,FK
    float balance_irr
    float pending_irr
  }

  WalletTransaction {
    uuid transaction_id PK
    uuid wallet_id FK
    uuid tenant_id FK
    string tx_type
    float amount_irr
  }

  ModelDeployment {
    uuid id PK
    uuid tenant_id FK
    string model_name
    string provider
    string model
  }

  VerificationToken {
    uuid id PK
    string token_hash UK
    uuid tenant_id FK
    uuid organization_id FK
    uuid team_id FK
    uuid user_id FK
    uuid project_id FK
  }

  RequestLog {
    bigint id PK
    uuid tenant_id FK
    string request_id
    string model
    string status
  }

  SpendLog {
    bigint id PK
    uuid tenant_id FK
    string request_id
    string model
    decimal cost_usd
    decimal cost_irr
  }

  Tenant ||--o| Wallet : has
  Tenant ||--o{ WalletTransaction : owns
  Wallet ||--o{ WalletTransaction : records
  Tenant ||--o{ ModelDeployment : deploys
  Tenant ||--o{ VerificationToken : issues
  Tenant ||--o{ RequestLog : logs
  Tenant ||--o{ SpendLog : tracks

  %% ---------------- Org / Team / User / Budget ----------------
  Budget {
    uuid budget_id PK
    decimal max_budget
    bigint rpm_limit
    bigint tpm_limit
  }

  ObjectPermission {
    uuid object_permission_id PK
    json permissions
  }

  ModelAlias {
    int id PK
    json aliases
  }

  Organization {
    uuid organization_id PK
    string organization_alias
    uuid budget_id FK
    uuid object_permission_id FK
  }

  Team {
    uuid team_id PK
    string team_alias
    uuid organization_id FK
    uuid object_permission_id FK
    int model_id FK
  }

  Project {
    uuid project_id PK
    uuid team_id FK
    uuid budget_id FK
    uuid object_permission_id FK
    string project_alias
  }

  User {
    uuid user_id PK
    string email UK
    uuid team_id FK
    uuid organization_id FK
    uuid object_permission_id FK
  }

  OrganizationMembership {
    uuid id PK
    uuid organization_id FK
    uuid user_id FK
    string role
  }

  EndUser {
    uuid user_id PK
    uuid team_id FK
    uuid budget_id FK
    decimal spend
  }

  Tag {
    string tag_name PK
    uuid budget_id FK
  }

  Budget o|--o{ Organization : assigned_to
  Budget o|--o{ Team : assigned_to
  Budget o|--o{ Project : assigned_to
  Budget o|--o{ EndUser : assigned_to
  Budget o|--o{ Tag : assigned_to

  ObjectPermission o|--o{ Organization : governs
  ObjectPermission o|--o{ Team : governs
  ObjectPermission o|--o{ Project : governs
  ObjectPermission o|--o{ User : governs

  ModelAlias o|--o{ Team : aliases

  Organization o|--o{ Team : contains
  Organization o|--o{ User : has
  Team o|--o{ User : has
  Team o|--o{ Project : owns
  Team o|--o{ EndUser : has

  Organization ||--o{ OrganizationMembership : membership
  User ||--o{ OrganizationMembership : membership

  Organization o|--o{ VerificationToken : scoped
  Team o|--o{ VerificationToken : scoped
  User o|--o{ VerificationToken : scoped
  Project o|--o{ VerificationToken : scoped

  %% ---------------- Tenant Provider Account Extension ----------------
  TenantProviderAccount {
    uuid account_id PK
    uuid tenant_id FK
    string provider
    string name
    string auth_mode
    bool has_credential
    bool allow_virtual_keys_only
  }

  TenantProviderAccountModel {
    uuid account_model_id PK
    uuid account_id FK
    string model
    bool enabled
    string pricing_mode
  }

  TenantProviderAccountAccess {
    uuid account_access_id PK
    uuid account_id FK
    string role
    string principal_type
    string principal_id
  }

  Tenant ||--o{ TenantProviderAccount : configures
  TenantProviderAccount ||--o{ TenantProviderAccountModel : includes
  TenantProviderAccount ||--o{ TenantProviderAccountAccess : controls
```

## Notes

Standalone models in Prisma (no explicit relation edges):
- `Credential`
- `Guardrail`
- `Policy`
- `ExchangeRate`
- `shahrah_sql_migrations`
