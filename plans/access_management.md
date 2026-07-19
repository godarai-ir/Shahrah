# Access Management Guide

This guide covers role-based access control (RBAC), user management, team management, and organization management in the GodarAI proxy server.

## Overview

Shahrah's access control is hierarchical, based on three main entity types:

- **Organizations** - Top-level entities representing business units or tenants
- **Teams** - Working groups within an organization
- **Internal Users** - Individual users who can create keys, make LLM API calls, and view usage
- **Virtual Keys** - API keys used for authentication, optionally associated with a user or team

## User Hierarchy

```
Organization
    ├── Team 1
    │   ├── Internal User A
    │   ├── Internal User B
    │   └── Virtual Keys
    ├── Team 2
    │   ├── Internal User C
    │   └── Virtual Keys
    └── Organization-level settings
```

### Key Type Classification

| Key Type | Description | Use Case |
|----------|-------------|----------|
| **User-only key** | Has a `user_id` but no `team_id` | Personal API keys for individual developers |
| **Team key (Service Account)** | Has a `team_id` but no `user_id` | Production apps, CI/CD pipelines, shared services |
| **User + Team key** | Has both `user_id` and `team_id` | Individual accountability within a team budget |

---

## Roles and Permissions

### Global Proxy Roles

These roles apply across the entire platform, regardless of organization or team boundaries.

| Role | Permissions |
|------|-------------|
| `proxy_admin` | Full platform control. Create/manage organizations, teams, users, keys, budgets |
| `proxy_admin_viewer` | Read-only access to all platform resources. Cannot create/delete keys or modify settings |
| `internal_user` | Create/delete own keys, view own spend. Can become team/org admin if assigned |

### Organization/Team Specific Roles

These roles are scoped to specific organizations or teams.

| Role | Scope | Permissions |
|------|-------|-------------|
| `org_admin` | Organization | Create teams, add users, view org spend, create keys within their org |
| `team_admin` | Team | Add/remove team members, update member budgets, manage team settings, create/delete team keys |

---

## Managing Organizations

### Creating an Organization

```bash
curl --location 'http://localhost:4000/organization/new' \
--header 'Authorization: Bearer <your-master-key>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "organization_alias": "engineering_dept",
  "models": ["gpt-4", "gpt-3.5-turbo"],
  "max_budget": 5000,
  "tpm_limit": 1000000,
  "rpm_limit": 5000
}'
```

**Sample Response:**

```json
{
  "organization_id": "ad15e8ca-12ae-46f4-8659-d02debef1b23",
  "organization_alias": "engineering_dept",
  "budget_id": "98754244-3a9c-4b31-b2e9-c63edc8fd7eb",
  "models": ["gpt-4", "gpt-3.5-turbo"],
  "max_budget": 5000,
  "tpm_limit": 1000000,
  "rpm_limit": 5000,
  "spend": 0.0
}
```

### Adding an Organization Admin

```bash
curl -X POST 'http://localhost:4000/organization/member_add' \
-H 'Authorization: Bearer <your-master-key>' \
-H 'Content-Type: application/json' \
-d '{
  "organization_id": "ad15e8ca-12ae-46f4-8659-d02debef1b23",
  "member": {"role": "org_admin", "user_id": "admin@company.com"}
}'
```

### Generating a Key for an Org Admin

```bash
curl --location 'http://localhost:4000/key/generate' \
--header 'Authorization: Bearer <your-master-key>' \
--header 'Content-Type: application/json' \
--data '{"user_id": "admin@company.com"}'
```

---

## Managing Teams

### Creating a Team

```bash
curl --location 'http://localhost:4000/team/new' \
--header 'Authorization: Bearer <your-master-key>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "team_alias": "platform-team",
  "organization_id": "ad15e8ca-12ae-46f4-8659-d02debef1b23",
  "members_with_roles": [{"role": "admin", "user_id": "admin@company.com"}],
  "max_budget": 2000,
  "tpm_limit": 200000,
  "rpm_limit": 1000,
  "models": ["gpt-4", "gpt-3.5-turbo"]
}'
```

**Sample Response:**

```json
{
  "team_alias": "platform-team",
  "team_id": "13e83b19-f851-43fe-8e93-f96e21033100",
  "organization_id": "ad15e8ca-12ae-46f4-8659-d02debef1b23",
  "members_with_roles": [
    {"role": "admin", "user_id": "admin@company.com"}
  ],
  "max_budget": 2000,
  "tpm_limit": 200000,
  "rpm_limit": 1000,
  "models": ["gpt-4", "gpt-3.5-turbo"],
  "spend": 0.0,
  "budget_duration": null,
  "budget_reset_at": null
}
```

### Adding Budget Duration to Teams

The budget is automatically reset at the end of the specified duration. Supported formats: `30s`, `30m`, `30h`, `30d`, `1mo`.

```bash
curl 'http://localhost:4000/team/new' \
--header 'Authorization: Bearer <your-master-key>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "team_alias": "platform-team",
  "max_budget": 2000,
  "budget_duration": "30d"
}'
```

### Setting Per-Model Rate Limits for a Team

```bash
curl --location 'http://localhost:4000/team/new' \
--header 'Authorization: Bearer <your-master-key>' \
--header 'Content-Type: application/json' \
--data '{
  "team_id": "my-prod-team",
  "model_rpm_limit": {"gpt-4": 100, "gpt-3.5-turbo": 200},
  "model_tpm_limit": {"gpt-4": 10000, "gpt-3.5-turbo": 20000}
}'
```

**Resolution order:** Key metadata > Key model_max_budget > Team metadata

### Updating Team Settings

```bash
curl --location 'http://localhost:4000/team/update' \
--header 'Authorization: Bearer <your-master-key>' \
--header 'Content-Type: application/json' \
--data '{
  "team_id": "my-prod-team",
  "max_budget": 3000,
  "rpm_limit": 2000
}'
```

---

## Managing Team Members

### Adding a User to a Team

```bash
curl -X POST 'http://localhost:4000/team/member_add' \
-H 'Authorization: Bearer <your-master-key>' \
-H 'Content-Type: application/json' \
-d '{
  "team_id": "13e83b19-f851-43fe-8e93-f96e21033100",
  "member": {"role": "user", "user_id": "developer@company.com"}
}'
```

### Setting Individual Budget Within a Team

```bash
curl -X POST 'http://localhost:4000/team/member_add' \
-H 'Authorization: Bearer <your-master-key>' \
-H 'Content-Type: application/json' \
-d '{
  "team_id": "13e83b19-f851-43fe-8e93-f96e21033100",
  "max_budget_in_team": 100.00,
  "member": {"role": "user", "user_id": "developer@company.com"}
}'
```

### Updating Team Member Permissions

Configure what regular team members can do with API keys:

```bash
curl --location 'http://localhost:4000/team/update' \
--header 'Authorization: Bearer <your-master-key>' \
--header 'Content-Type: application/json' \
--data '{
  "team_id": "team-123",
  "team_member_permissions": ["/key/info", "/key/health", "/key/generate", "/key/update"]
}'
```

**Available permissions:**

| Permission | Method | Description |
|------------|--------|-------------|
| `/key/info` | GET | View virtual key information |
| `/key/health` | GET | Check key health status |
| `/key/list` | GET | List all team virtual keys |
| `/key/generate` | POST | Create new virtual keys |
| `/key/update` | POST | Modify existing virtual keys |
| `/key/delete` | POST | Delete virtual keys |
| `/key/regenerate` | POST | Regenerate virtual keys |
| `/key/block` | POST | Block virtual keys |
| `/key/unblock` | POST | Unblock virtual keys |

### Viewing Team Member Permissions

```bash
curl --location 'http://localhost:4000/team/permissions_list?team_id=team-123' \
--header 'Authorization: Bearer <your-master-key>'
```

**Expected Response:**

```json
{
  "team_id": "team-123",
  "team_member_permissions": ["/key/info", "/key/health"],
  "all_available_permissions": ["/key/generate", "/key/update", "/key/delete", "/key/regenerate", "/key/block", "/key/unblock", "/key/list", "/key/health", "/key/info"]
}
```

---

## Managing Internal Users

### Creating a User

```bash
curl --location 'http://localhost:4000/user/new' \
--header 'Authorization: Bearer <your-master-key>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "user_id": "user@company.com",
  "user_role": "internal_user",
  "max_budget": 100,
  "models": ["gpt-4", "gpt-3.5-turbo"]
}'
```

### Setting Budget Duration for Users

```bash
curl 'http://localhost:4000/user/new' \
--header 'Authorization: Bearer <your-master-key>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "user_id": "user@company.com",
  "max_budget": 10,
  "budget_duration": "30d"
}'
```

### Setting Default Budget for All Internal Users

Configure in `config.yaml`:

```yaml
shahrah_settings:
  max_internal_user_budget: 100     # amount in USD
  internal_user_budget_duration: "1mo"  # reset frequency
```

### Updating a User

```bash
curl --location 'http://localhost:4000/user/update' \
--header 'Authorization: Bearer <your-master-key>' \
--header 'Content-Type: application/json' \
--data '{
  "user_id": "user@company.com",
  "max_budget": 200,
  "rpm_limit": 100,
  "tpm_limit": 50000
}'
```

---

## Managing Virtual Keys

### Generating a Key

```bash
curl --location 'http://localhost:4000/key/generate' \
--header 'Authorization: Bearer <your-master-key>' \
--header 'Content-Type: application/json' \
--data '{
  "user_id": "user@company.com",
  "team_id": "team-123",
  "models": ["gpt-4"],
  "max_budget": 50,
  "budget_duration": "30d",
  "rpm_limit": 100,
  "tpm_limit": 50000
}'
```

**Sample Response:**

```json
{
  "key": "sk-xxxxxxxxxxxxxxxxxxxxxxxx",
  "key_name": "sk-...xxxx",
  "models": ["gpt-4"],
  "spend": 0.0,
  "max_budget": 50.0,
  "user_id": "user@company.com",
  "team_id": "team-123",
  "budget_duration": "30d",
  "rpm_limit": 100,
  "tpm_limit": 50000,
  "expires": null
}
```

### Setting Multiple Budget Windows on a Key

```bash
curl 'http://localhost:4000/key/generate' \
--header 'Authorization: Bearer <your-master-key>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "budget_limits": [
    {"budget_duration": "24h",  "max_budget": 10},
    {"budget_duration": "30d",  "max_budget": 100}
  ]
}'
```

### Setting Model-Specific Budgets on a Key

```bash
curl 'http://localhost:4000/key/generate' \
--header 'Authorization: Bearer <your-master-key>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "model_max_budget": {
    "gpt-4": {"budget_limit": 0.0000001, "time_period": "1d"},
    "gpt-3.5-turbo": {"budget_limit": 10, "time_period": "30d"}
  }
}'
```

### Updating a Key

```bash
curl --location 'http://localhost:4000/key/update' \
--header 'Authorization: Bearer <your-master-key>' \
--header 'Content-Type: application/json' \
--data '{
  "key": "sk-xxxxxxxxxxxxxxxxxxxxxxxx",
  "max_budget": 100,
  "rpm_limit": 200
}'
```

### Deleting a Key

```bash
curl --location 'http://localhost:4000/key/delete' \
--header 'Authorization: Bearer <your-master-key>' \
--header 'Content-Type: application/json' \
--data '{"keys": ["sk-xxxxxxxxxxxxxxxxxxxxxxxx"]}'
```

---

## Service Account Keys

Service account keys are team keys that are not tied to a specific user. They persist even when team members leave.

```bash
curl --location 'http://localhost:4000/key/service-account/generate' \
--header 'Authorization: Bearer <your-master-key>' \
--header 'Content-Type: application/json' \
--data '{
  "team_id": "team-123",
  "key_alias": "production-api-key",
  "models": ["gpt-4"],
  "max_budget": 500
}'
```

---

## End Users (Customers)

Manage budget and rate limits for end users (customers) without creating individual keys.

### Creating a Customer Budget

```bash
curl --location 'http://localhost:4000/customer/new' \
--header 'Authorization: Bearer <your-master-key>' \
--header 'Content-Type: application/json' \
--data '{
  "user_id": "customer-123",
  "max_budget": 50,
  "budget_duration": "30d"
}'
```

### Setting Rate Limits for End Users

Create a budget and assign to a customer:

```bash
# Step 1: Create budget
curl --location 'http://localhost:4000/budget/new' \
--header 'Authorization: Bearer <your-master-key>' \
--header 'Content-Type: application/json' \
--data '{
  "budget_id": "free-tier",
  "tpm_limit": 5000,
  "rpm_limit": 10
}'

# Step 2: Assign to customer
curl --location 'http://localhost:4000/customer/new' \
--header 'Authorization: Bearer <your-master-key>' \
--header 'Content-Type: application/json' \
--data '{
  "user_id": "customer-123",
  "budget_id": "free-tier"
}'
```

---

## Model Access Control

### Using Model Access Groups

Define model access groups in `config.yaml`:

```yaml
model_list:
  - model_name: gpt-4
    litellm_params:
      model: gpt-4
      api_key: os.environ/OPENAI_API_KEY
    model_info:
      access_groups: ["premium-models"]
```

Then create a user or key with access to the group:

```bash
curl --location 'http://localhost:4000/user/new' \
--header 'Authorization: Bearer <your-master-key>' \
--header 'Content-Type: application/json' \
--data '{"models": ["premium-models"], "max_budget": 100}'
```

---

## Authentication Response Headers

After making API calls, the following headers provide rate limit and budget information:

| Header | Description |
|--------|-------------|
| `x-shahrah-key-remaining-requests` | Remaining requests for the key |
| `x-shahrah-key-remaining-tokens` | Remaining tokens for the key |
| `x-shahrah-key-remaining-requests-{model}` | Remaining requests for a specific model |
| `x-shahrah-key-remaining-tokens-{model}` | Remaining tokens for a specific model |
| `x-shahrah-key-max-budget` | Key's maximum budget |
| `x-shahrah-key-spend` | Key's current spend |

---

## Implementation Guidelines

### When Implementing Access Management Features:

1. **Follow the hierarchy**: Organization > Team > Key > End User
2. **Budget validation**: Always validate that child budgets don't exceed parent budgets
3. **Role checks**: Verify user has appropriate role before allowing operations
4. **Rate limit resolution**: Key metadata > Key model_max_budget > Team metadata
5. **Service account handling**: Service account keys should not be deleted when users leave
6. **Timezone handling**: Use IANA timezone strings, default to UTC if not specified

### Database Schema Considerations:

- Organizations map to budget tables via `budget_id`
- Teams link to organizations via `organization_id`
- Users can belong to multiple teams via membership tables
- Keys optionally link to users and/or teams
- All spend tracking flows through a unified spend logging mechanism
