import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Budgets & Rate Limits

:::info **Budget Setup Options**

**Personal budgets**: Create virtual keys without `team_id` for individual spending limits

**Team budgets**: Add `team_id` to virtual keys to utilize a team's shared budget

**Team member budgets**: Set individual spending limits within the team's shared budget

**Organization budgets**: Set organization-wide spending caps that apply to all teams within

***If a key belongs to a team, the team budget is applied, not the user's personal budget.***
:::

## Prerequisites

- A PostgreSQL database (e.g., [Supabase](https://supabase.com/), [Neon](https://neon.tech/), etc.)
- GodarAI proxy server configured and running

---

## Set Budgets

### Global Proxy

Apply a budget across all calls on the proxy.

**Step 1. Modify config.yaml**

```yaml
general_settings:
  master_key: sk-1234

shahrah_settings:
  max_budget: 0           # (float) sets max budget as $0 USD
  budget_duration: 30d    # (str) frequency of reset - You can set duration as seconds ("30s"), minutes ("30m"), hours ("30h"), days ("30d").
```

**Step 2. Start proxy**

```bash
GodarAI --config /path/to/config.yaml --port 4000
```

**Step 3. Send test call**

```bash
curl --location 'http://0.0.0.0:4000/chat/completions' \
    --header 'Authorization: Bearer sk-1234' \
    --header 'Content-Type: application/json' \
    --data '{
    "model": "gpt-3.5-turbo",
    "messages": [
        {
        "role": "user",
        "content": "what llm are you"
        }
    ]
}'
```

---

### Organization

You can add budgets and rate limits to Organizations. Organizations represent business units within a tenant.

#### **Add budgets to organizations**

```bash
curl --location 'http://localhost:4000/organization/new' \
--header 'Authorization: Bearer <your-master-key>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "organization_alias": "engineering_dept",
  "max_budget": 5000,
  "rpm_limit": 5000,
  "models": ["gpt-4", "gpt-3.5-turbo", "claude-3-opus"]
}'
```

**Sample Response**

```json
{
  "organization_id": "ad15e8ca-12ae-46f4-8659-d02debef1b23",
  "organization_alias": "engineering_dept",
  "budget_id": "98754244-3a9c-4b31-b2e9-c63edc8fd7eb",
  "models": ["gpt-4", "gpt-3.5-turbo", "claude-3-opus"],
  "max_budget": 5000,
  "rpm_limit": 5000,
  "spend": 0.0,
  "budget_duration": null,
  "budget_reset_at": null
}
```

#### **Add budget duration to organizations**

`budget_duration`: Budget is reset at the end of specified duration. If not set, budget is never reset. You can set duration as seconds (`"30s"`), minutes (`"30m"`), hours (`"30h"`), days (`"30d"`), or months (`"1mo"`).

```bash
curl 'http://0.0.0.0:4000/organization/new' \
--header 'Authorization: Bearer <your-master-key>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "organization_alias": "engineering_dept",
  "max_budget": 5000,
  "budget_duration": "30d"
}'
```

---

### Team

You can add budgets and rate limits to Teams. Teams are working groups within an organization.

#### **Add budgets to teams**

```bash
curl --location 'http://localhost:4000/team/new' \
--header 'Authorization: Bearer <your-master-key>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "team_alias": "platform-team",
  "organization_id": "ad15e8ca-12ae-46f4-8659-d02debef1b23",
  "members_with_roles": [{"role": "admin", "user_id": "5c4a0aa3-a1e1-43dc-bd87-3c2da8382a3a"}],
  "max_budget": 2000,
  "rpm_limit": 1000,
  "models": ["gpt-4", "gpt-3.5-turbo"]
}'
```

**Sample Response**

```json
{
  "team_alias": "platform-team",
  "team_id": "13e83b19-f851-43fe-8e93-f96e21033100",
  "organization_id": "ad15e8ca-12ae-46f4-8659-d02debef1b23",
  "admins": [],
  "members": [],
  "members_with_roles": [
    {
      "role": "admin",
      "user_id": "5c4a0aa3-a1e1-43dc-bd87-3c2da8382a3a"
    }
  ],
  "metadata": {},
  "rpm_limit": 1000,
  "max_budget": 2000,
  "models": ["gpt-4", "gpt-3.5-turbo"],
  "spend": 0.0,
  "max_parallel_requests": null,
  "budget_duration": null,
  "budget_reset_at": null
}
```

#### **Add budget duration to teams**

```bash
curl 'http://0.0.0.0:4000/team/new' \
--header 'Authorization: Bearer <your-master-key>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "team_alias": "platform-team",
  "max_budget": 2000,
  "budget_duration": "30d"
}'
```

#### **Set rate limits per model for a team**

Use `model_rpm_limit` to set rate limits per model for all keys belonging to a team. These limits apply across all keys in the team and are inherited by keys unless overridden at the key level.

```bash
curl --location 'http://0.0.0.0:4000/team/new' \
--header 'Authorization: Bearer sk-1234' \
--header 'Content-Type: application/json' \
--data '{
  "team_id": "my-prod-team",
  "model_rpm_limit": {"gpt-4": 100, "gpt-3.5-turbo": 200}
}'
```

**Update existing team with per-model limits:**

```bash
curl --location 'http://0.0.0.0:4000/team/update' \
--header 'Authorization: Bearer sk-1234' \
--header 'Content-Type: application/json' \
--data '{
  "team_id": "my-prod-team",
  "model_rpm_limit": {"gpt-4": 100, "gpt-3.5-turbo": 200}
}'
```

**Resolution order:** When a key belongs to a team, rate limits are resolved as: **Key metadata > Key model_max_budget > Team metadata**. Keys can override team-level per-model limits with their own `model_rpm_limit`.

**Verify:** Make a `/chat/completions` request and check response header `x-shahrah-key-remaining-requests-{model}` for the model-specific limits.

---

### Team Members

Use this when you want to budget a user's spend within a Team.

#### Step 1. Create User

Create a user with `user_id=ishaan`:

```bash
curl --location 'http://0.0.0.0:4000/user/new' \
    --header 'Authorization: Bearer sk-1234' \
    --header 'Content-Type: application/json' \
    --data '{
        "user_id": "ishaan"
}'
```

#### Step 2. Add User to an existing Team - set `max_budget_in_team`

Set `max_budget_in_team` when adding a User to a team. We use the same `user_id` we set in Step 1:

```bash
curl -X POST 'http://0.0.0.0:4000/team/member_add' \
-H 'Authorization: Bearer sk-1234' \
-H 'Content-Type: application/json' \
-d '{"team_id": "e8d1460f-846c-45d7-9b43-55f3cc52ac32", "max_budget_in_team": 0.000000000001, "member": {"role": "user", "user_id": "ishaan"}}'
```

#### Step 3. Create a Key for Team member from Step 1

Set `user_id=ishaan` from step 1:

```bash
curl --location 'http://0.0.0.0:4000/key/generate' \
    --header 'Authorization: Bearer sk-1234' \
    --header 'Content-Type: application/json' \
    --data '{
        "user_id": "ishaan",
        "team_id": "e8d1460f-846c-45d7-9b43-55f3cc52ac32"
}'
```

Response from `/key/generate`:

```json
{
  "key": "sk-RV-l2BJEZ_LYNChSx2EueQ",
  "models": [],
  "spend": 0.0,
  "max_budget": null,
  "user_id": "ishaan",
  "team_id": "e8d1460f-846c-45d7-9b43-55f3cc52ac32",
  "max_parallel_requests": null,
  "metadata": {},
  "rpm_limit": null,
  "budget_duration": null,
  "soft_budget": null,
  "key_alias": null,
  "duration": null,
  "aliases": {},
  "config": {},
  "permissions": {},
  "model_max_budget": {},
  "key_name": null,
  "expires": null
}
```

#### Step 4. Make `/chat/completions` requests for Team member

Use the key from step 3 for this request. After 2-3 requests expect to see the following error `ExceededBudget: Crossed spend within team`:

```bash
curl --location 'http://localhost:4000/chat/completions' \
    --header 'Authorization: Bearer sk-RV-l2BJEZ_LYNChSx2EueQ' \
    --header 'Content-Type: application/json' \
    --data '{
    "model": "llama3",
    "messages": [
        {
        "role": "user",
        "content": "tes4"
        }
    ]
}'
```

---

### Internal User

Apply a budget across all calls an internal user (key owner) can make on the proxy.

:::info

For keys with a `team_id` set, the team budget is used instead of the user's personal budget.

To apply a budget to a user within a team, use team member budgets.
:::

You can:
- Add budgets to users
- Add budget durations to reset spend

By default the `max_budget` is set to `null` and is not checked for keys.

#### **Add budgets to users**

```bash
curl --location 'http://localhost:4000/user/new' \
--header 'Authorization: Bearer <your-master-key>' \
--header 'Content-Type: application/json' \
--data-raw '{"models": ["azure-models"], "max_budget": 0, "user_id": "user@company.com"}'
```

**Sample Response**

```json
{
  "key": "sk-YF2OxDbrgd1y2KgwxmEA2w",
  "expires": "2023-12-22T09:53:13.861000Z",
  "user_id": "user@company.com",
  "max_budget": 0.0
}
```

#### **Add budget duration to users**

`budget_duration`: Budget is reset at the end of specified duration. If not set, budget is never reset. You can set duration as seconds (`"30s"`), minutes (`"30m"`), hours (`"30h"`), days (`"30d"`), or months (`"1mo"`).

```bash
curl 'http://0.0.0.0:4000/user/new' \
--header 'Authorization: Bearer <your-master-key>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "max_budget": 10,
  "budget_duration": "30s"
}'
```

#### Create new keys for existing user

Now you can call `/key/generate` with that `user_id` and:
- **Budget Check**: The user's budget (e.g., $10) will be checked for this key
- **Spend Tracking**: Spend for this key will update the user's spend as well

```bash
curl --location 'http://0.0.0.0:4000/key/generate' \
--header 'Authorization: Bearer <your-master-key>' \
--header 'Content-Type: application/json' \
--data '{"models": ["azure-models"], "user_id": "user@company.com"}'
```

---

### Virtual Key

Apply a budget on a key.

You can:
- Add budgets to keys
- Add budget durations to reset spend

**Expected Behaviour**
- Costs per key get auto-populated in the verification token table
- After the key crosses its `max_budget`, requests fail
- If duration is set, spend is reset at the end of the duration

By default the `max_budget` is set to `null` and is not checked for keys.

#### **Add budgets to keys**

```bash
curl 'http://0.0.0.0:4000/key/generate' \
--header 'Authorization: Bearer <your-master-key>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "team_id": "core-infra",
  "max_budget": 10
}'
```

Example Request to `/chat/completions` when key has crossed budget:

```bash
curl --location 'http://0.0.0.0:4000/chat/completions' \
  --header 'Content-Type: application/json' \
  --header 'Authorization: Bearer <generated-key>' \
  --data '{
  "model": "azure-gpt-3.5",
  "user": "e09b4da8-ed80-4b05-ac93-e16d9eb56fca",
  "messages": [
      {
      "role": "user",
      "content": "respond in 50 lines"
      }
  ]
}'
```

Expected Response from `/chat/completions` when key has crossed budget:

```json
{
  "detail": "Authentication Error, ExceededTokenBudget: Current spend for token: 7.2e-05; Max Budget for Token: 2e-07"
}
```

#### **Add budget duration to keys**

```bash
curl 'http://0.0.0.0:4000/key/generate' \
--header 'Authorization: Bearer <your-master-key>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "team_id": "core-infra",
  "max_budget": 10,
  "budget_duration": "30d"
}'
```

#### **Set multiple budget windows on a key**

Apply multiple concurrent budget limits at different time scales on the same key — for example, cap a key at **$10/day** AND **$100/month**.

**When is this useful?**

A single `budget_duration` window cannot prevent a bad day from burning your entire month. Multiple budget windows let you:

- Block a runaway usage spike within the day while still allowing normal monthly spend.
- Give daily guardrails and a monthly ceiling so a single heavy session doesn't exhaust the whole month.
- Layer fine-grained hourly limits for bursty workloads on top of a weekly cap.

**Via API**

Pass `budget_limits` as a list of `{budget_duration, max_budget}` objects:

```bash
curl 'http://0.0.0.0:4000/key/generate' \
--header 'Authorization: Bearer <your-master-key>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "budget_limits": [
    {"budget_duration": "24h",  "max_budget": 10},
    {"budget_duration": "30d",  "max_budget": 100}
  ]
}'
```

Each window is tracked independently and resets on its own schedule:

| `budget_duration` | Resets |
|---|---|
| `1h`  | Every hour |
| `24h` | Daily at midnight UTC |
| `7d`  | Every Sunday at midnight UTC |
| `30d` | 1st of every month at midnight UTC |

---

### Model-Specific Budgets on Keys

Apply model-specific budgets on a key. Example:
- Budget for `gpt-4o` is $0.0000001, for time period `1d` for `key = "sk-12345"`
- Budget for `gpt-4o-mini` is $10, for time period `30d` for `key = "sk-12345"`

The spec for `model_max_budget` is a dictionary mapping model names to [`BudgetInfo`](#budgetinfo) objects.

```bash
curl 'http://0.0.0.0:4000/key/generate' \
--header 'Authorization: Bearer <your-master-key>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "model_max_budget": {"gpt-4o": {"budget_limit": 0.0000001, "time_period": "1d"}}
}'
```

#### Make a test request

We expect the first request to succeed, and the second request to fail since we cross the budget for `gpt-4o` on the Virtual Key.

<Tabs>
<TabItem label="Successful Call" value="allowed">

```bash
curl --location 'http://0.0.0.0:4000/chat/completions' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer <sk-generated-key>' \
--data '{
      "model": "gpt-4o",
      "messages": [
        {
          "role": "user",
          "content": "testing request"
        }
      ]
    }'
```

</TabItem>
<TabItem label="Unsuccessful call" value="not-allowed">

Expect this to fail since we cross the budget `model=gpt-4o` on the Virtual Key:

```bash
curl --location 'http://0.0.0.0:4000/chat/completions' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer <sk-generated-key>' \
--data '{
      "model": "gpt-4o",
      "messages": [
        {
          "role": "user",
          "content": "testing request"
        }
      ]
    }'
```

Expected response on failure:

```json
{
    "error": {
        "message": "Virtual Key exceeded budget for model=gpt-4o",
        "type": "budget_exceeded",
        "param": null,
        "code": "400"
    }
}
```

</TabItem>
</Tabs>

---

### End Users (Customers)

Use this to budget the `user` passed to `/chat/completions`, **without needing to create a key for every user**.

**Step 1. Modify config.yaml**

Define `max_end_user_budget`:

```yaml
general_settings:
  master_key: sk-1234

shahrah_settings:
  max_end_user_budget: 0.0001  # budget for 'user' passed to /chat/completions
```

**Step 2. Make a `/chat/completions` call with `user` - First call works**

```bash
curl --location 'http://0.0.0.0:4000/chat/completions' \
        --header 'Content-Type: application/json' \
        --header 'Authorization: Bearer sk-zi5onDRdHGD24v0Zdn7VBA' \
        --data '{
        "model": "azure-gpt-3.5",
        "user": "enduser123",
        "messages": [
            {
            "role": "user",
            "content": "what time is it"
            }
        ]
        }'
```

**Step 3. Make another call - Fails since `enduser123` is over budget**

```bash
curl --location 'http://0.0.0.0:4000/chat/completions' \
        --header 'Content-Type: application/json' \
        --header 'Authorization: Bearer sk-zi5onDRdHGD24v0Zdn7VBA' \
        --data '{
        "model": "azure-gpt-3.5",
        "user": "enduser123",
        "messages": [
            {
            "role": "user",
            "content": "what time is it"
            }
        ]
        }'
```

Error response:

```json
{
  "error": {
    "message": "Budget has been exceeded: User enduser123 has exceeded their budget. Current spend: 0.000887; Max Budget: 0.0001",
    "type": "auth_error",
    "param": null,
    "code": 401
  }
}
```

---

## Reset Budgets

Reset budgets across keys, internal users, teams, and organizations.

`budget_duration`: Budget is reset at the end of specified duration. If not set, budget is never reset. You can set duration as seconds (`"30s"`), minutes (`"30m"`), hours (`"30h"`), days (`"30d"`), or months (`"1mo"`).

<Tabs>
<TabItem value="users" label="Internal Users">

```bash
curl 'http://0.0.0.0:4000/user/new' \
--header 'Authorization: Bearer <your-master-key>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "max_budget": 10,
  "budget_duration": "30d"
}'
```

</TabItem>
<TabItem value="keys" label="Keys">

```bash
curl 'http://0.0.0.0:4000/key/generate' \
--header 'Authorization: Bearer <your-master-key>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "max_budget": 10,
  "budget_duration": "30d"
}'
```

</TabItem>
<TabItem value="teams" label="Teams">

```bash
curl 'http://0.0.0.0:4000/team/new' \
--header 'Authorization: Bearer <your-master-key>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "max_budget": 10,
  "budget_duration": "30d"
}'
```

</TabItem>
<TabItem value="organizations" label="Organizations">

```bash
curl 'http://0.0.0.0:4000/organization/new' \
--header 'Authorization: Bearer <your-master-key>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "max_budget": 10000,
  "budget_duration": "1mo"
}'
```

</TabItem>
</Tabs>

**Note:** By default, the server checks for resets every 10 minutes to minimize database calls.

To change this, set `proxy_budget_rescheduler_min_time` and `proxy_budget_rescheduler_max_time`:

Check every 1 second:

```yaml
general_settings:
  proxy_budget_rescheduler_min_time: 1
  proxy_budget_rescheduler_max_time: 1
```

---

## Set Rate Limits

You can set:
- RPM limits (requests per minute)
- Max parallel requests
- Per-model RPM limits for a given key or team


---

<Tabs>
<TabItem value="per-team" label="Per Team">

Use `/team/new` or `/team/update` to persist rate limits across multiple keys for a team.

```bash
curl --location 'http://0.0.0.0:4000/team/new' \
--header 'Authorization: Bearer sk-1234' \
--header 'Content-Type: application/json' \
--data '{"team_id": "my-prod-team", "max_parallel_requests": 10, "rpm_limit": 4}'
```

**Expected Response**

```json
{
  "key": "sk-sA7VDkyhlQ7m8Gt77Mbt3Q",
  "expires": "2024-01-19T01:21:12.816168",
  "team_id": "my-prod-team"
}
```

</TabItem>
<TabItem value="per-user" label="Per Internal User">

Use `/user/new` or `/user/update` to persist rate limits across multiple keys for internal users.

```bash
curl --location 'http://0.0.0.0:4000/user/new' \
--header 'Authorization: Bearer sk-1234' \
--header 'Content-Type: application/json' \
--data '{"user_id": "user@company.com", "max_parallel_requests": 10, "rpm_limit": 4}'
```

**Expected Response**

```json
{
  "key": "sk-sA7VDkyhlQ7m8Gt77Mbt3Q",
  "expires": "2024-01-19T01:21:12.816168",
  "user_id": "user@company.com"
}
```

</TabItem>
<TabItem value="per-key" label="Per Key">

Use `/key/generate` if you want limits for just that key.

```bash
curl --location 'http://0.0.0.0:4000/key/generate' \
--header 'Authorization: Bearer sk-1234' \
--header 'Content-Type: application/json' \
--data '{"max_parallel_requests": 10, "rpm_limit": 4}'
```

**Expected Response**

```json
{
  "key": "sk-ulGNRXWtv7M0lFnnsQk0wQ",
  "expires": "2024-01-18T20:48:44.297973",
  "user_id": "78c2c8fc-c233-43b9-b0c3-eb931da27b84"
}
```

</TabItem>
<TabItem value="per-key-model" label="Per API Key Per Model">

Set `model_rpm_limit` to set rate limits per model per API key.

```bash
curl --location 'http://0.0.0.0:4000/key/generate' \
--header 'Authorization: Bearer sk-1234' \
--header 'Content-Type: application/json' \
--data '{"model_rpm_limit": {"gpt-4": 2}}'
```

**Expected Response**

```json
{
  "key": "sk-ulGNRXWtv7M0lFnnsQk0wQ",
  "expires": "2024-01-18T20:48:44.297973"
}
```

**Verify Model Rate Limits**

Make a `/chat/completions` request and check for `x-shahrah-key-remaining-requests-gpt-4`:

```bash
curl -i http://localhost:4000/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-ulGNRXWtv7M0lFnnsQk0wQ" \
  -d '{
    "model": "gpt-4",
    "messages": [
      {"role": "user", "content": "Hello!"}
    ]
  }'
```

**Expected headers**

```
x-shahrah-key-remaining-requests-gpt-4: 1
```

These indicate:
- 1 request remaining for GPT-4 for key `sk-ulGNRXWtv7M0lFnnsQk0wQ`

</TabItem>
</Tabs>

---

## Set Default Budget for ALL Internal Users

Use this to set a default budget for users who you give keys to.

This applies when a user has `user_role="internal_user"` (set via `/user/new` or `/user/update`).

This does NOT apply if a key has a `team_id` (team budgets apply then).

**1. Define max budget in config.yaml**

```yaml
model_list:
  - model_name: "gpt-3.5-turbo"
    provider_params:
      model: gpt-3.5-turbo
      api_key: os.environ/OPENAI_API_KEY

shahrah_settings:
  max_internal_user_budget: 0    # amount in USD
  internal_user_budget_duration: "1mo"  # reset every month
```

**2. Create key for user**

```bash
curl -L -X POST 'http://0.0.0.0:4000/key/generate' \
-H 'Authorization: Bearer sk-1234' \
-H 'Content-Type: application/json' \
-d '{}'
```

Expected Response:

```json
{
  "key": "sk-X53RdxnDhzamRwjKXR4IHg"
}
```

**3. Test it**

```bash
curl -L -X POST 'http://0.0.0.0:4000/chat/completions' \
-H 'Content-Type: application/json' \
-H 'Authorization: Bearer sk-X53RdxnDhzamRwjKXR4IHg' \
-d '{
    "model": "gpt-3.5-turbo",
    "messages": [{"role": "user", "content": "Hey, how'\''s it going?"}]
}'
```

Expected Response (when over budget):

```json
{
  "error": {
    "message": "ExceededBudget: User=<user_id> over budget. Spend=3.7e-05, Budget=0.0",
    "type": "budget_exceeded",
    "param": null,
    "code": "400"
  }
}
```

---

## Multi-Instance Rate Limiting

**Important Notes:**
- **Rate limits do not apply to proxy admin users.**
- When testing rate limits, use internal user roles (non-admin) to ensure limits are enforced as expected.

**Implementation details:**
- Uses async increment instead of async set-cache when updating current requests.
- The in-memory cache is synced with Redis every 0.01s to avoid calling Redis for every request.
- In testing, this was found to be 2x faster than the previous implementation, and reduced drift between expected and actual failures to at most 10 requests at high-traffic (100 RPS across 3 instances).

---

## Grant Access to New Model

Use model access groups to give users access to select models, and add new ones over time.

Difference between doing this with `/key/generate` vs. `/user/new`? If you do it on `/user/new`, it persists across multiple keys generated for that user.

**Step 1. Assign model access group in config.yaml**

```yaml
model_list:
  - model_name: text-embedding-ada-002
    provider_params:
      model: azure/azure-embedding-model
      api_base: "os.environ/AZURE_API_BASE"
      api_key: "os.environ/AZURE_API_KEY"
      api_version: "2023-07-01-preview"
    model_info:
      access_groups: ["beta-models"]
```

**Step 2. Create key with access group**

```bash
curl --location 'http://localhost:4000/user/new' \
-H 'Authorization: Bearer <your-master-key>' \
-H 'Content-Type: application/json' \
-d '{"models": ["beta-models"], "max_budget": 0}'
```

---

## Budget Inheritance Hierarchy

Budgets are enforced in a hierarchical manner. When a request arrives, the following checks are performed in order:

```
Organization Budget (top-level)
    └── Team Budget (within Organization)
        └── Key Budget (within Team)
            └── End User Budget (optional)
```

### Inheritance Rules

1. **Team budget cannot exceed Organization budget**: When creating a team within an organization, the team's `max_budget` must be <= the organization's `max_budget`.
2. **Key budget cannot exceed Team budget**: Same validation applies at the key level.
3. **Model access validation**: A team can only access models that its parent organization allows. A key can only access models that its parent team allows.
4. **If a key has its own `budget_id`**, it uses that budget. Otherwise, it inherits the team's budget. If the team has no budget, it inherits the organization's budget.

### Soft Budget vs Hard Budget

- **`max_budget`** (Hard): When reached, all requests with this key/team/org are blocked until budget resets.
- **`soft_budget`**: When reached, alerts are triggered (e.g., Slack notification) but requests continue.

---

## API Specification

### `BudgetInfo`

A model that defines budget information with a time period and limit.

```python
class BudgetInfo(BaseModel):
    budget_limit: float   # The maximum budget amount in USD
    time_period: str      # Duration string like "1d", "30d", etc.
```

#### Fields:
- `budget_limit` (float): The maximum budget amount in USD
- `time_period` (str): Duration string specifying the time period for the budget. Supported formats:
  - Seconds: `"30s"`
  - Minutes: `"30m"`
  - Hours: `"30h"`
  - Days: `"30d"`
  - Months: `"1mo"`

#### Example:

```json
{
  "budget_limit": 0.0001,
  "time_period": "1d"
}
```
