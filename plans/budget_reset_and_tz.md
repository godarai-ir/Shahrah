# Budget Reset and Timezones

GodarAI supports predictable budget reset times that align with natural calendar boundaries, ensuring consistent and transparent spend management across your organization.

## How Budget Resets Work

All budgets reset at midnight (00:00:00) in the configured timezone. The reset behavior depends on the duration you configure:

| Duration | Reset Behavior |
|----------|----------------|
| Daily (`24h` or `1d`) | Resets at midnight every day |
| Weekly (`7d`) | Resets on Monday at midnight |
| Monthly (`30d` or `1mo`) | Resets on the 1st of each month at midnight |
| Custom (`Nh`, `Nm`, `Ns`) | Resets every N hours/minutes/seconds, aligned to time boundaries |

## Configuring the Timezone

Specify the timezone for all budget resets in your configuration file:

```yaml
general_settings:
  master_key: sk-1234

shahrah_settings:
  max_budget: 100         # (float) sets max budget as $100 USD
  budget_duration: 30d    # (string) reset frequency - seconds ("30s"), minutes ("30m"), hours ("30h"), days ("30d"), months ("1mo")
  timezone: "Asia/Tehran" # Any valid IANA timezone string
```

This ensures that all budget resets happen at midnight in your specified timezone rather than in UTC. If no timezone is specified, UTC will be used by default.

## Supported Timezones

Any valid [IANA timezone string](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones) is supported (powered by Python's `zoneinfo` module). Daylight Saving Time (DST) transitions are handled automatically.

**Common timezone values:**

| Timezone | Description |
|----------|-------------|
| `UTC` | Coordinated Universal Time |
| `US/Eastern` | Eastern Time |
| `US/Pacific` | Pacific Time |
| `Europe/London` | UK Time |
| `Europe/Berlin` | Central European Time |
| `Asia/Tehran` | Iran Standard Time (IRST) |
| `Asia/Kolkata` | Indian Standard Time (IST) |
| `Asia/Bangkok` | Indochina Time (ICT) |
| `Asia/Tokyo` | Japan Standard Time |
| `Asia/Dubai` | Gulf Standard Time |
| `Australia/Sydney` | Australian Eastern Time |

## Budget Duration Formats

The following formats are supported for `budget_duration`:

| Format | Example | Description |
|--------|---------|-------------|
| Seconds | `30s` | Reset every 30 seconds |
| Minutes | `30m` | Reset every 30 minutes |
| Hours | `24h` | Reset every 24 hours at aligned hour boundaries |
| Days | `30d` | Reset every 30 days (or daily with `1d`) |
| Weeks | `7d` | Reset every Monday at midnight |
| Months | `1mo` | Reset on the 1st of each month at midnight |

## Budget Inheritance and Reset Priority

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
3. **If a key has its own `budget_id`**, it uses that budget. Otherwise, it inherits the team's budget. If the team has no budget, it inherits the organization's budget.

## Soft Budget vs Hard Budget

- **`max_budget` (Hard)**: When reached, all requests with this key/team/org are blocked until budget resets.
- **`soft_budget`**: When reached, alerts are triggered (e.g., Slack notification) but requests continue to be processed.

## Reset Timing Behavior

By default, the server checks for budget resets every 10 minutes to minimize database calls. To adjust this frequency, configure the following settings:

```yaml
general_settings:
  proxy_budget_rescheduler_min_time: 1
  proxy_budget_rescheduler_max_time: 1
```

Setting both values to `1` will check every second.

## Multiple Budget Windows

You can apply multiple concurrent budget limits at different time scales on the same key — for example, cap a key at **$10/day** AND **$100/month**.

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
|-------------------|--------|
| `1h` | Every hour at aligned boundaries |
| `24h` | Daily at midnight in configured timezone |
| `7d` | Every Monday at midnight in configured timezone |
| `30d` | 1st of every month at midnight in configured timezone |

## API Specification

### `BudgetInfo`

A model that defines budget information with a time period and limit.

```python
class BudgetInfo(BaseModel):
    budget_limit: float   # The maximum budget amount in USD
    time_period: str      # Duration string like "1d", "30d", "1mo", etc.
```

#### Fields:
- `budget_limit` (float): The maximum budget amount in USD
- `time_period` (str): Duration string specifying the time period for the budget

#### Example:

```json
{
  "budget_limit": 0.0001,
  "time_period": "1d"
}
```
