# Role-based Access Controls (RBAC) for Shahrah

> **Note**: For the complete hierarchical structure including Tenant subdomains, Organization Units, and Virtual Users, see [tenant-organization-team-design.md](./tenant-organization-team-design.md).

Role-based access control (RBAC) in GodarAI is based on Organizations, Teams, and Internal User Roles.

## Hierarchy Overview

```mermaid
graph TD
    A[Platform Admin] --> B[Organizations]
    B --> C[Teams]
    C --> D[Internal Users]
    D --> E[API Keys]
```

- **Organizations** are the top-level entities that contain Teams.
- **Team** - A Team is a collection of multiple Internal Users.
- **Internal Users** - Users that can create keys, make API calls, and view usage onGodarAI. Users can belong to multiple teams.
- **API Keys** - Keys are used for authentication to the GodarAI API. Each key can optionally be associated with a `user_id`, a `team_id`, or both:
  - **User-only key**: Has a `user_id` but no `team_id`. Tracked individually, deleted when the user is deleted.
  - **Team key (Service Account)**: Has a `team_id` but no `user_id`. Shared by the team, persists even when team members leave.
  - **User + Team key**: Has both `user_id` and `team_id`. Belongs to a specific user within a team context.

### When to Use Each Key Type

| Key Type | Use Case | Spend Tracking | Lifecycle |
|----------|----------|----------------|-----------|
| **User-only** | Personal API keys for individual developers | Tracked to the user | Deleted when user is deleted |
| **Team (Service Account)** | Production applications, CI/CD pipelines, shared services | Tracked to the team only | Persists even when team members leave |
| **User + Team** | User working within a team context | Tracked to both user and team | Deleted when user is deleted |

**Example scenarios:**
- Use **user-only keys** for developers testing locally
- Use **team service account keys** for production applications that should not break when employees leave
- Use **user + team keys** when you want individual accountability within a team budget

---

## User Roles

GodarAI has two types of roles:

1. **Global Platform Roles** - Platform-wide roles that apply across all organizations and teams
2. **Organization/Team Specific Roles** - Roles scoped to specific organizations or teams

### Global Platform Roles

| Role Name | Permissions |
|-----------|-------------|
| `admin` | Full control over the entire platform. Admin over all organizations, teams, and users |
| `viewer` | Can login, view all keys, view all spend across the platform. **Cannot** create keys, delete keys, or add new users |
| `user` | Can login, view/create (when allowed by team-specific permissions)/delete their own keys, view their own spend. **Cannot** add new users |

### Organization/Team Specific Roles

| Role Name | Permissions |
|-----------|-------------|
| `org_admin` | Admin over a specific organization. Can create teams and users within their organization |
| `team_admin` | Admin over a specific team. Can manage team members, update team member permissions, and create keys for their team |

---

## What Can Each Role Do?

Here is what each role can actually do. Think of it like levels of access.

---

## Global Platform Roles

These roles apply across the entire GodarAI platform, regardless of organization or team boundaries.

### Admin - Full Access

The platform admin controls everything. They are like the owner of the whole platform.

**What they can do:**
- Create and manage all organizations
- Create and manage all teams (across all organizations)
- Create and manage all users
- View all spend and usage across the platform
- Create and delete keys for anyone
- Update team budgets, rate limits, and models
- Manage team members and assign roles

**Who should be a platform admin:** Only the people running the GodarAI instance.

---

### Viewer - Platform-Wide Read Access

The platform viewer can see everything across the platform but cannot make changes.

**What they can do:**
- View all organizations, teams, and users
- View all spend and usage across the platform
- View all API keys
- Login to the admin dashboard

**What they cannot do:**
- Create or delete keys
- Add or remove users
- Modify budgets, rate limits, or settings
- Make any changes to the platform

**Who should be a platform viewer:** Finance teams, auditors, or stakeholders who need platform-wide visibility without modification rights.

---

### User

A user can create API keys (when allowed by team-specific permissions) and make calls. They see their own stuff only. They can become a team admin or org admin if they are assigned the respective roles.

**What they can do:**
- Create API keys for themselves
- Delete their own API keys
- View their own spend and usage
- Make API calls using their keys

**Who should be a user:** Anyone who needs UI access for team/org specific operations or for developers you plan to give multiple keys to.

---

## Organization/Team Specific Roles

These roles are scoped to specific organizations or teams. Users with these roles can only manage resources within their assigned organization or team.

### Org Admin - Organization Level Access

An org admin manages one or more organizations. They can create teams within their organization but cannot touch other organizations.

**What they can do:**
- Create teams within their organization
- Add users to teams in their organization
- View spend for their organization
- Create keys for users in their organization

**What they cannot do:**
- Create or manage other organizations
- Modify org budgets or rate limits
- Modify org allowed models (e.g., adding a platform-level model to the org)

**Who should be an org admin:** Department leads or managers who need to manage multiple teams.

---

### Team Admin - Team Level Access

A team admin manages a specific team. They are like a team lead who can add people, update settings, but only for their team.

**What they can do:**
- Add or remove team members from their team
- Update team members budgets and rate limits within the team
- Change team settings (budget, rate limits, models)
- Create and delete keys for team members
- Onboard a team-specific model to GodarAI (e.g., onboarding a team fine-tuned model)
- Configure team member permissions to control what regular team members can do

**What they cannot do:**
- Create new teams
- Modify team budget or rate limits
- Add or remove global platform models to their team

**Who should be a team admin:** Team leads who need to manage their team API access without bothering IT.

**How to create a team admin:**

```shell
curl -X POST 'http://0.0.0.0:4000/team/member_add' \
    -H 'Authorization: Bearer sk-1234' \
    -H 'Content-Type: application/json' \
    -d '{"team_id": "team-123", "member": {"role": "admin", "user_id": "user@company.com"}}'
```

---

## Team Member Permissions

Team member permissions allow you to control what regular team members (with `role=user`) can do with API keys in their team. By default, team members can only view key information, but you can grant them additional permissions to create, update, or delete keys.

### How It Works

- **Applies to**: Team members with `role=user` (not team admins or org admins)
- **Scope**: Permissions only apply to keys belonging to their team
- **Configuration**: Set at the team level using `team_member_permissions`
- **Override**: Team admins and org admins always have full permissions regardless of these settings

### Available Permissions

| Permission | Method | Description |
|-----------|--------|-------------|
| `/key/info` | GET | View information about virtual keys in the team |
| `/key/health` | GET | Check health status of virtual keys in the team |
| `/key/list` | GET | List all virtual keys belonging to the team |
| `/key/generate` | POST | Create new virtual keys for the team |
| `/key/service-account/generate` | POST | Create service account keys (not tied to a specific user) for the team |
| `/key/update` | POST | Modify existing virtual keys in the team |
| `/key/delete` | POST | Delete virtual keys belonging to the team |
| `/key/regenerate` | POST | Regenerate virtual keys in the team |
| `/key/block` | POST | Block virtual keys in the team |
| `/key/unblock` | POST | Unblock virtual keys in the team |

### Default Permissions

By default, team members can only:
- `/key/info` - View key information
- `/key/health` - Check key health

### Common Permission Scenarios

**Read-only access** (default):
```json
["/key/info", "/key/health"]
```

**Allow key creation but not deletion**:
```json
["/key/info", "/key/health", "/key/generate", "/key/update"]
```

**Full key management**:
```json
["/key/info", "/key/health", "/key/generate", "/key/update", "/key/delete", "/key/regenerate", "/key/block", "/key/unblock", "/key/list"]
```

### How to Configure Team Member Permissions

#### View Current Permissions

```shell
curl --location 'http://0.0.0.0:4000/team/permissions_list?team_id=team-123' \
    --header 'Authorization: Bearer sk-1234'
```

Expected Response:
```json
{
  "team_id": "team-123",
  "team_member_permissions": ["/key/info", "/key/health"],
  "all_available_permissions": ["/key/generate", "/key/update", "/key/delete", ...]
}
```

#### Update Team Member Permissions

```shell
curl --location 'http://0.0.0.0:4000/team/update' \
    --header 'Authorization: Bearer sk-1234' \
    --header 'Content-Type: application/json' \
    -d '{
        "team_id": "team-123",
        "team_member_permissions": ["/key/info", "/key/health", "/key/generate", "/key/update"]
    }'
```

This allows team members to:
- View key information
- Create new keys
- Update existing keys
- But NOT delete keys

### Who Can Configure These Permissions?

- **Platform Admin**: Can configure permissions for any team
- **Org Admin**: Can configure permissions for teams in their organization
- **Team Admin**: Can configure permissions for their own team

---

## Quick Comparison

Here is the quick version:

### Global Platform Roles

| Action | Admin | Viewer | User |
|--------|-------|--------|------|
| Create organizations | Yes | No | No |
| Create teams | Yes | No | No |
| Manage all teams | Yes | No | No |
| Create or delete any keys | Yes | No | No |
| Create or delete own keys | Yes | No | Yes |
| View all platform spend | Yes | Yes | No |
| View own spend | Yes | Yes | Yes |
| View all keys | Yes | Yes | No |
| View own keys | Yes | Yes | Yes |
| Add or remove users | Yes | No | No |

### Organization/Team Specific Roles

| Action | Org Admin | Team Admin |
|--------|-----------|------------|
| Create teams (in their org) | Yes | No |
| Manage teams in their org | Yes | No |
| Manage their specific team | Yes | Yes |
| Add or remove team members | Yes (in their org) | Yes (their team only) |
| Update team budgets | Yes (in their org) | Yes (their team only) |
| Create keys for team members | Yes (in their org) | Yes (their team only) |
| View organization spend | Yes (their org) | No |
| View team spend | Yes (in their org) | Yes (their team) |
| Create organizations | No | No |
| View all platform spend | No | No |

---

## Onboarding Organizations

### 1. Creating a new Organization

Any user with `role=admin` can create a new organization.

**Usage**

**API Reference for /organization/new**

```shell
curl --location 'http://0.0.0.0:4000/organization/new' \
    --header 'Authorization: Bearer sk-1234' \
    --header 'Content-Type: application/json' \
    -d '{
        "organization_alias": "marketing_department",
        "models": ["gpt-4"],
        "max_budget": 20
    }'
```

Expected Response:

```json
{
  "organization_id": "ad15e8ca-12ae-46f4-8659-d02debef1b23",
  "organization_alias": "marketing_department",
  "budget_id": "98754244-3a9c-4b31-b2e9-c63edc8fd7eb",
  "metadata": {},
  "models": [
    "gpt-4"
  ],
  "created_by": "109010464461339474872",
  "updated_by": "109010464461339474872",
  "created_at": "2024-10-08T18:30:24.637000Z",
  "updated_at": "2024-10-08T18:30:24.637000Z"
}
```

### 2. Adding an `org_admin` to an Organization

Create a user (`ishaan@company.com`) as an `org_admin` for the `marketing_department` Organization (from [step 1](#1-creating-a-new-organization)).

Users with the following roles can call `/organization/member_add`:
- `admin`
- `org_admin` only within their own organization

```shell
curl -X POST 'http://0.0.0.0:4000/organization/member_add' \
    -H 'Authorization: Bearer sk-1234' \
    -H 'Content-Type: application/json' \
    -d '{"organization_id": "ad15e8ca-12ae-46f4-8659-d02debef1b23", "member": {"role": "org_admin", "user_id": "ishaan@company.com"}}'
```

Now a user with `user_id = ishaan@company.com` and `role = org_admin` has been created in the `marketing_department` Organization.

Create an API Key for `user_id = ishaan@company.com`. The user can then use the API key for their Organization Admin Operations:

```shell
curl --location 'http://0.0.0.0:4000/key/generate' \
    --header 'Authorization: Bearer sk-1234' \
    --header 'Content-Type: application/json' \
    -d '{
        "user_id": "ishaan@company.com"
    }'
```

Expected Response:

```json
{
  "models": [],
  "user_id": "ishaan@company.com",
  "key": "sk-7shH8TGMAofR4zQpAAo6kQ",
  "key_name": "sk-...o6kQ"
}
```

### 3. Organization Admin - Create a Team

The organization admin will use the API key created in [step 2](#2-adding-an-org_admin-to-an-organization) to create a `Team` within the `marketing_department` Organization:

```shell
curl --location 'http://0.0.0.0:4000/team/new' \
    --header 'Authorization: Bearer sk-7shH8TGMAofR4zQpAAo6kQ' \
    --header 'Content-Type: application/json' \
    -d '{
        "team_alias": "engineering_team",
        "organization_id": "ad15e8ca-12ae-46f4-8659-d02debef1b23"
    }'
```

This will create the team `engineering_team` within the `marketing_department` Organization.

Expected Response:

```json
{
  "team_alias": "engineering_team",
  "team_id": "01044ee8-441b-45f4-be7d-c70e002722d8",
  "organization_id": "ad15e8ca-12ae-46f4-8659-d02debef1b23"
}
```

### 4. Organization Admin - Add a Team Admin

The organization admin can now add a team admin who will manage the `engineering_team`:

- We assign `role=admin` to make them a team admin for this specific team
- `team_id` is from [step 3](#3-organization-admin---create-a-team)

```shell
curl -X POST 'http://0.0.0.0:4000/team/member_add' \
    -H 'Authorization: Bearer sk-7shH8TGMAofR4zQpAAo6kQ' \
    -H 'Content-Type: application/json' \
    -d '{"team_id": "01044ee8-441b-45f4-be7d-c70e002722d8", "member": {"role": "admin", "user_id": "john@company.com"}}'
```

Now `john@company.com` is a team admin. They can manage the `engineering_team` - add members, update budgets, create keys - but they cannot touch other teams.

Create an API Key for the team admin:

```shell
curl --location 'http://0.0.0.0:4000/key/generate' \
    --header 'Authorization: Bearer sk-7shH8TGMAofR4zQpAAo6kQ' \
    --header 'Content-Type: application/json' \
    -d '{"user_id": "john@company.com"}'
```

Expected Response:

```json
{
  "models": [],
  "user_id": "john@company.com",
  "key": "sk-TeamAdminKey123",
  "key_name": "sk-...Key123"
}
```

### 5. Team Admin - Add Team Members

Now the team admin can use their key to add team members without needing to ask the org admin:

```shell
curl -X POST 'http://0.0.0.0:4000/team/member_add' \
    -H 'Authorization: Bearer sk-TeamAdminKey123' \
    -H 'Content-Type: application/json' \
    -d '{"team_id": "01044ee8-441b-45f4-be7d-c70e002722d8", "member": {"role": "user", "user_id": "krrish@company.com"}}'
```

The team admin can also create keys for their team members:

```shell
curl --location 'http://0.0.0.0:4000/key/generate' \
    --header 'Authorization: Bearer sk-TeamAdminKey123' \
    --header 'Content-Type: application/json' \
    -d '{
        "user_id": "krrish@company.com",
        "team_id": "01044ee8-441b-45f4-be7d-c70e002722d8"
    }'
```

### 6. Team Admin - Update Team Settings

The team admin can update team budgets and rate limits:

```shell
curl --location 'http://0.0.0.0:4000/team/update' \
    --header 'Authorization: Bearer sk-TeamAdminKey123' \
    --header 'Content-Type: application/json' \
    -d '{
        "team_id": "01044ee8-441b-45f4-be7d-c70e002722d8",
        "max_budget": 100,
        "rpm_limit": 1000
    }'
```

---

## Implementation Guide

This section provides guidance for implementing the access control system in GodarAI based on the patterns from LiteLLM.

### Key Implementation Files

Based on the LiteLLM codebase structure, the following files and patterns should be adapted for Shahrah:

1. **Types and Enums** (`proxy/_types.py`)
   - Define `UserRoles` enum with: `ADMIN`, `VIEWER`, `USER`, `ORG_ADMIN`, `TEAM_ADMIN`
   - Define `TeamMemberRoles` enum with: `TEAM_ADMIN`, `TEAM_MEMBER`
   - Define request/response types for organization, team, and user management

2. **Auth Checks** (`proxy/auth/auth_checks.py`)
   - Implement `_get_user_role()` function to retrieve user role
   - Implement `_is_user_admin()` helper function
   - Implement route access checks based on user role
   - Implement team membership validation

3. **Organization Endpoints** (`proxy/management_endpoints/organization_endpoints.py`)
   - `/organization/new` - Create new organization
   - `/organization/update` - Update organization settings
   - `/organization/delete` - Delete organization
   - `/organization/member_add` - Add member to organization
   - `/organization/member_update` - Update member role
   - `/organization/member_delete` - Remove member from organization
   - `/organization/info` - Get organization information
   - `/organization/list` - List organizations

4. **Team Endpoints** (`proxy/management_endpoints/team_endpoints.py`)
   - `/team/new` - Create new team
   - `/team/update` - Update team settings
   - `/team/delete` - Delete team
   - `/team/info` - Get team information
   - `/team/list` - List teams
   - `/team/member_add` - Add member to team
   - `/team/member_delete` - Remove member from team
   - `/team/permissions_list` - List team member permissions
   - `/team/unblock` - Unblock a team

5. **User Endpoints** (`proxy/management_endpoints/internal_user_endpoints.py`)
   - `/user/new` - Create new user
   - `/user/update` - Update user settings
   - `/user/delete` - Delete user
   - `/user/info` - Get user information
   - `/user/list` - List users

### Database Schema

The database schema should include the following tables (based on LiteLLM Prisma schema):

- `LiteLLM_OrganizationTable` - Organizations
- `LiteLLM_TeamTable` - Teams
- `LiteLLM_UserTable` - Users
- `LiteLLM_OrganizationMembership` - Organization memberships (links users to organizations with roles)
- `LiteLLM_TeamMembership` - Team memberships (links users to teams with roles and budgets)
- `LiteLLM_VerificationToken` - API keys (linked to users and/or teams)
- `LiteLLM_BudgetTable` - Budgets for organizations, teams, and users
- `LiteLLM_ObjectPermissionTable` - Object permissions for resources like vector stores and MCP servers

### Access Verification Pattern

The access verification should follow this pattern:

```python
async def _verify_team_access(team_obj: TeamTable, user_api_key_dict: UserAPIKeyAuth) -> None:
    """
    Verify the caller is authorized to manage the given team.
    
    Access is granted if:
    - Caller is a platform admin, OR
    - Caller is an org admin for the team organization, OR
    - Caller is a team admin of this team
    
    Raises HTTPException(403) otherwise.
    """
    if user_api_key_dict.user_role == UserRoles.ADMIN:
        return
    
    if _is_user_team_admin(user_api_key_dict=user_api_key_dict, team_obj=team_obj):
        return
    
    if await _is_user_org_admin_for_team(user_api_key_dict=user_api_key_dict, team_obj=team_obj):
        return
    
    raise HTTPException(status_code=403, detail="You do not have access to this team")
```

### Team Member Permission Checks

Team member permissions should be checked using a dedicated helper:

```python
class TeamMemberPermissionChecks:
    @staticmethod
    def check_permission(team: TeamTable, permission: str, user_role: str) -> bool:
        """
        Check if a team member has the specified permission.
        Team admins and org admins always have full permissions.
        """
        if user_role in [UserRoles.ADMIN, UserRoles.ORG_ADMIN, UserRoles.TEAM_ADMIN]:
            return True
        
        team_permissions = team.team_member_permissions or []
        return permission in team_permissions
```

### Budget Management

Budget management should be integrated with the access control system:

- Organizations have budgets that limit total spend across all teams
- Teams have budgets that limit total spend across all keys
- Team members can have individual budgets within the team
- Budget checks should be enforced on each API call

### Key Types Implementation

API keys should support three association patterns:

1. **User-only key**: `key.user_id = user_id`, `key.team_id = null`
2. **Team key**: `key.user_id = null`, `key.team_id = team_id`
3. **User + Team key**: `key.user_id = user_id`, `key.team_id = team_id`

The key generation endpoint should accept both `user_id` and `team_id` as optional parameters.
