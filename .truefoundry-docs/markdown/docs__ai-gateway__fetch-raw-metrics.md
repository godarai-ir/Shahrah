> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# API Access to Metrics

> Query Gateway model metrics for usage, cost, and performance analytics via API.

The **Gateway Model Metrics Query API** provides a flexible way to query gateway model metrics on model usage, performance, cost, and user activity. You can retrieve either **distribution** (aggregated) or **timeseries** metrics with powerful filtering and grouping capabilities.

### Access control

* **Tenant admins:** Can query metrics for the entire organization (tenant-wide).
* **Users:** Can query their own data and their teams' data.
* **Virtual accounts:** Can query their own data and their teams' data; with tenant-admin permissions, they can access tenant-wide data.

## Authentication

You need to authenticate with your TrueFoundry API key. You can use either a Personal Access Token **(PAT)** or Virtual Account Token **(VAT)**.

<Accordion title="Get your API key">
  To generate an API key:

  1. **Personal Access Token (PAT)**: Go to Access → Personal Access Tokens in your TrueFoundry dashboard
  2. **Virtual Account Token (VAT)**: Go to Access → Virtual Account Tokens (requires admin permissions)

  For detailed authentication setup, see our [Authentication guide](/docs/ai-gateway/authentication).
</Accordion>

## Quick Start

<Warning>
  By default, the API returns metrics for **both models and virtual models**. If you want metrics for only models or only virtual models, you must explicitly filter using the `IS_NULL` operator on the virtual model name field in your request filters.
</Warning>

### Distribution Query

Get aggregated metrics distribution with multiple aggregations including count, sum, and percentiles:

```python theme={"dark"}
import requests

response = requests.post(
    "https://{your_control_plane_url}/api/svc/v1/llm-gateway/metrics/query",
    headers={
        "Authorization": "Bearer <your_api_key>",
        "Content-Type": "application/json"
    },
    json={
        "startTs": "2025-01-21T00:00:00.000Z",
        "endTs": "2025-01-22T00:00:00.000Z",
        "datasource": "modelMetrics",
        "type": "distribution",
        "aggregations": [
            {"type": "count", "column": "modelName"},
            {"type": "sum", "column": "inputTokens"},
            {"type": "sum", "column": "outputTokens"},
            {"type": "p99", "column": "latencyMs"},
            {"type": "sum", "column": "costInUSD"}
        ],
        "groupBy": ["modelName"],
        "filters": [
            {
                "fieldName": "virtualModelName",
                "operator": "IS_NULL",
                "value": true
            }
        ]
    }
)

print(response.json())
```

### Timeseries Query

Get metrics over time with hourly intervals, including latency percentiles:

```python theme={"dark"}
import requests

response = requests.post(
    "https://{your_control_plane_url}/api/svc/v1/llm-gateway/metrics/query",
    headers={
        "Authorization": "Bearer <your_api_key>",
        "Content-Type": "application/json"
    },
    json={
        "startTs": "2025-01-21T00:00:00.000Z",
        "endTs": "2025-01-22T00:00:00.000Z",
        "datasource": "modelMetrics",
        "type": "timeseries",
        "aggregations": [
            {"type": "count", "column": "modelName"},
            {"type": "sum", "column": "inputTokens"},
            {"type": "p99", "column": "latencyMs"}
        ],
        "groupBy": ["modelName"],
        "intervalInSeconds": 3600,
        "filters": [
            {
                "fieldName": "virtualModelName",
                "operator": "IS_NULL",
                "value": true
            }
        ],
    }
)

print(response.json())
```

## API Reference

### Endpoint

```
POST https://{your_control_plane_url}/api/svc/v1/llm-gateway/metrics/query
```

### Request Parameters

<ParamField path="startTs" type="string" required>
  ISO 8601 timestamp for the start of the data range (e.g., `"2025-01-21T00:00:00.000Z"`)
</ParamField>

<ParamField path="endTs" type="string" required>
  ISO 8601 timestamp for the end of the data range (e.g., `"2025-01-22T00:00:00.000Z"`)
</ParamField>

<ParamField path="datasource" type="string" required>
  The data source to query. Use `"modelMetrics"` for gateway model metrics.
</ParamField>

<ParamField path="type" type="string" required>
  The type of query to execute:

  * `"distribution"` - Returns aggregated metrics
  * `"timeseries"` - Returns metrics over time intervals
</ParamField>

<ParamField path="aggregations" type="array">
  Array of aggregation objects. Each aggregation specifies:

  * `type` - The aggregation type
  * `column` - The column to aggregate on

  **Supported aggregation types:**

  | Type    | Description              |
  | ------- | ------------------------ |
  | `count` | Count of records         |
  | `sum`   | Sum of values            |
  | `p50`   | 50th percentile (median) |
  | `p75`   | 75th percentile          |
  | `p90`   | 90th percentile          |
  | `p99`   | 99th percentile          |

  ```json theme={"dark"}
  "aggregations": [
      {"type": "count", "column": "modelName"},
      {"type": "sum", "column": "inputTokens"},
      {"type": "p90", "column": "latencyMs"}
  ]
  ```

  Supported columns for aggregation:

  * `costInUSD` - Cost incurred in USD
  * `inputTokens` - Number of input tokens
  * `outputTokens` - Number of output tokens
  * `latencyMs` - Total request latency (ms)
  * `interTokenLatencyMs` -  Latency between the generation of consecutive tokens (ms)
  * `timeToFirstTokenMs` - Time to first token (ms)
  * `timePerOutputTokenLatencyMs` - Latency per output token (ms)
</ParamField>

<ParamField path="groupBy" type="array">
  Array of fields to group the metrics by. Available options:

  * `modelName` - Group by model name
  * `userEmail` - Group by user email
  * `virtualaccount` - Group by virtual account
  * `team` - Group by team (unnests the Teams array)
  * `virtualModel` - Group by virtual model
  * `errorCode` - HTTP error code returned
  * `requestType` - Type of model request (e.g. `ChatCompletion`, `Embedding` `etc`)
  * `providerAccountType` - Account type of provider (e.g. `model`, `mcp-server`, `guardrail-config`)
  * `providerModelName` - Underlying provider model name
  * `createdBySubjectType` - Subject type (e.g. `user`, `virtualaccount`)
  * `metadata.<key>` - Group by a custom metadata key (e.g., `metadata.environment`)

  ```json theme={"dark"}
  "groupBy": ["modelName", "team", "metadata.environment"]
  ```
</ParamField>

<ParamField path="filters" type="array">
  Array of filter objects to narrow down the results. See [Filtering](#filtering) for details.
</ParamField>

<ParamField path="intervalInSeconds" type="number">
  **Required for timeseries queries.** The time interval in seconds for grouping data points.

  Common values:

  * `60` - 1 minute intervals
  * `300` - 5 minute intervals
  * `1800` - 30 minute intervals
  * `3600` - 1 hour intervals
  * `86400` - 1 day intervals
</ParamField>

## Filtering

Filters allow you to narrow down your query results. The API supports different filter operators depending on the field type.

### Filter Structure

<Tabs>
  <Tab title="Field Filters">
    For standard fields, use `fieldName`:

    ```json theme={"dark"}
    {
        "fieldName": "modelName",
        "operator": "IN",
        "value": ["gpt-4", "gpt-3.5-turbo"]
    }
    ```
  </Tab>

  <Tab title="Metadata Filters">
    For metadata fields, use `metadataKey`:

    ```json theme={"dark"}
    {
        "metadataKey": "environment",
        "operator": "IN",
        "value": ["production"]
    }
    ```
  </Tab>
</Tabs>

### Filterable Fields

| Field              | Type   | Description                                  |
| ------------------ | ------ | -------------------------------------------- |
| `modelName`        | string | The name of the LLM model                    |
| `requestType`      | string | Type of request (e.g., "chat", "completion") |
| `userEmail`        | string | The email of the user making requests        |
| `virtualAccount`   | string | The virtual account name                     |
| `team`             | array  | Teams associated with the request            |
| `latencyMs`        | number | Request latency in milliseconds              |
| `conversationID`   | string | The conversation identifier                  |
| `virtualModelName` | string | The virtual model name                       |
| `inputTokens`      | number | Number of input tokens                       |
| `outputTokens`     | number | Number of output tokens                      |

### Filter Operators

#### String Field Operators

| Operator             | Description                             | Example Value                |
| -------------------- | --------------------------------------- | ---------------------------- |
| `EQUAL`              | Exact match                             | `"gpt-4"`                    |
| `IN`                 | Match any value in the list             | `["gpt-4", "gpt-3.5-turbo"]` |
| `NOT_IN`             | Exclude values in the list              | `["deprecated-model"]`       |
| `STRING_CONTAINS`    | Contains substring                      | `"gpt"`                      |
| `STRING_STARTS_WITH` | Starts with prefix                      | `"gpt-"`                     |
| `STRING_ENDS_WITH`   | Ends with suffix                        | `"-turbo"`                   |
| `GREATER_THAN`       | Lexicographically greater than          | `"gpt-3"`                    |
| `LESS_THAN`          | Lexicographically less than             | `"gpt-5"`                    |
| `GREATER_THAN_EQUAL` | Lexicographically greater than or equal | `"gpt-3"`                    |
| `LESS_THAN_EQUAL`    | Lexicographically less than or equal    | `"gpt-5"`                    |
| `BETWEEN`            | Lexicographically between two values    | `["a", "z"]`                 |

```python theme={"dark"}
# Filter for specific models
{
    "fieldName": "modelName",
    "operator": "IN",
    "value": ["gpt-4", "gpt-3.5-turbo"]
}

# Exclude specific users
{
    "fieldName": "userEmail",
    "operator": "NOT_IN",
    "value": ["excluded@example.com"]
}

# Filter models containing "gpt"
{
    "fieldName": "modelName",
    "operator": "STRING_CONTAINS",
    "value": "gpt"
}

# Filter models starting with "claude"
{
    "fieldName": "modelName",
    "operator": "STRING_STARTS_WITH",
    "value": "claude"
}
```

#### Numeric Field Operators

| Operator             | Description                    | Example Value     |
| -------------------- | ------------------------------ | ----------------- |
| `EQUAL`              | Exact match                    | `1000`            |
| `GREATER_THAN`       | Greater than value             | `1000`            |
| `LESS_THAN`          | Less than value                | `5000`            |
| `GREATER_THAN_EQUAL` | Greater than or equal to       | `100`             |
| `LESS_THAN_EQUAL`    | Less than or equal to          | `1000`            |
| `BETWEEN`            | Between two values (inclusive) | `[500, 5000]`     |
| `IN`                 | Match any value in the list    | `[100, 200, 300]` |
| `NOT_IN`             | Exclude values in the list     | `[0]`             |

```python theme={"dark"}
# Filter for high latency requests
{
    "fieldName": "latencyMs",
    "operator": "GREATER_THAN",
    "value": 1000
}

# Filter for latency within a range
{
    "fieldName": "latencyMs",
    "operator": "BETWEEN",
    "value": [500, 5000]
}

# Filter for requests with significant input tokens
{
    "fieldName": "inputTokens",
    "operator": "GREATER_THAN",
    "value": 100
}
```

#### Array Field Operators (Teams)

| Operator         | Description                                | Example Value                 |
| ---------------- | ------------------------------------------ | ----------------------------- |
| `ARRAY_HAS_ANY`  | Match if array contains any of the values  | `["team-alpha", "team-beta"]` |
| `ARRAY_HAS_NONE` | Match if array contains none of the values | `["excluded-team"]`           |

```python theme={"dark"}
# Filter for specific teams
{
    "fieldName": "team",
    "operator": "ARRAY_HAS_ANY",
    "value": ["team-alpha", "team-beta"]
}

# Exclude specific teams
{
    "fieldName": "team",
    "operator": "ARRAY_HAS_NONE",
    "value": ["excluded-team"]
}
```

### Combining Multiple Filters

You can combine multiple filters in a single query. All filters are applied with AND logic:

```python theme={"dark"}
{
    "startTs": "2025-01-21T00:00:00.000Z",
    "endTs": "2025-01-22T00:00:00.000Z",
    "datasource": "modelMetrics",
    "type": "distribution",
    "aggregations": [],
    "filters": [
        {
            "fieldName": "virtualModelName",
            "operator": "IS_NULL",
            "value": true
        },
        {
            "fieldName": "modelName",
            "operator": "IN",
            "value": ["gpt-4", "gpt-3.5-turbo"]
        },
        {
            "fieldName": "latencyMs",
            "operator": "LESS_THAN",
            "value": 5000
        },
        {
            "fieldName": "team",
            "operator": "ARRAY_HAS_ANY",
            "value": ["team-alpha"]
        },
        {
            "metadataKey": "environment",
            "operator": "IN",
            "value": ["production"]
        }
    ],
    "groupBy": ["modelName", "team"]
}
```

## Query Examples

### Distribution Queries

<AccordionGroup>
  <Accordion title="Count by model name">
    Get request counts grouped by model:

    ```python theme={"dark"}
    response = requests.post(
        "https://{your_control_plane_url}/api/svc/v1/llm-gateway/metrics/query",
        headers={
            "Authorization": "Bearer <your_api_key>",
            "Content-Type": "application/json"
        },
        json={
            "startTs": "2025-01-21T00:00:00.000Z",
            "endTs": "2025-01-22T00:00:00.000Z",
            "datasource": "modelMetrics",
            "type": "distribution",
            "aggregations": [
                {"type": "count", "column": "modelName"}
            ],
            "filters": [
                {
                    "fieldName": "virtualModelName",
                    "operator": "IS_NULL",
                    "value": true
                }
            ],
            "groupBy": ["modelName"]
        }
    )
    ```
  </Accordion>

  <Accordion title="Count by team">
    Get request counts grouped by team:

    ```python theme={"dark"}
    response = requests.post(
        "https://{your_control_plane_url}/api/svc/v1/llm-gateway/metrics/query",
        headers={
            "Authorization": "Bearer <your_api_key>",
            "Content-Type": "application/json"
        },
        json={
            "startTs": "2025-01-21T00:00:00.000Z",
            "endTs": "2025-01-22T00:00:00.000Z",
            "datasource": "modelMetrics",
            "type": "distribution",
            "aggregations": [
                {"type": "count", "column": "team"}
            ],
            "filters": [
                {
                    "fieldName": "virtualModelName",
                    "operator": "IS_NULL",
                    "value": true
                }
            ],
            "groupBy": ["team"]
        }
    )
    ```
  </Accordion>

  <Accordion title="Sum tokens by model">
    Get total input and output tokens grouped by model:

    ```python theme={"dark"}
    response = requests.post(
        "https://{your_control_plane_url}/api/svc/v1/llm-gateway/metrics/query",
        headers={
            "Authorization": "Bearer <your_api_key>",
            "Content-Type": "application/json"
        },
        json={
            "startTs": "2025-01-21T00:00:00.000Z",
            "endTs": "2025-01-22T00:00:00.000Z",
            "datasource": "modelMetrics",
            "type": "distribution",
            "aggregations": [
                {"type": "sum", "column": "inputTokens"},
                {"type": "sum", "column": "outputTokens"}
            ],
            "filters": [
                {
                    "fieldName": "virtualModelName",
                    "operator": "IS_NULL",
                    "value": true
                }
            ],
            "groupBy": ["modelName"]
        }
    )
    ```
  </Accordion>

  <Accordion title="Latency percentiles by model">
    Get p50, p90, and p99 latency percentiles grouped by model:

    ```python theme={"dark"}
    response = requests.post(
        "https://{your_control_plane_url}/api/svc/v1/llm-gateway/metrics/query",
        headers={
            "Authorization": "Bearer <your_api_key>",
            "Content-Type": "application/json"
        },
        json={
            "startTs": "2025-01-21T00:00:00.000Z",
            "endTs": "2025-01-22T00:00:00.000Z",
            "datasource": "modelMetrics",
            "type": "distribution",
            "aggregations": [
                {"type": "p50", "column": "latencyMs"},
                {"type": "p90", "column": "latencyMs"},
                {"type": "p99", "column": "latencyMs"}
            ],
            "filters": [
                {
                    "fieldName": "virtualModelName",
                    "operator": "IS_NULL",
                    "value": true
                }
            ],
            "groupBy": ["modelName"]
        }
    )
    ```
  </Accordion>

  <Accordion title="Multi-dimensional grouping">
    Group by multiple dimensions:

    ```python theme={"dark"}
    response = requests.post(
        "https://{your_control_plane_url}/api/svc/v1/llm-gateway/metrics/query",
        headers={
            "Authorization": "Bearer <your_api_key>",
            "Content-Type": "application/json"
        },
        json={
            "startTs": "2025-01-21T00:00:00.000Z",
            "endTs": "2025-01-22T00:00:00.000Z",
            "datasource": "modelMetrics",
            "type": "distribution",
            "aggregations": [],
            "filters": [
                {
                    "fieldName": "virtualModelName",
                    "operator": "IS_NULL",
                    "value": true
                }
            ],
            "groupBy": ["modelName", "userEmail", "virtualaccount"]
        }
    )
    ```
  </Accordion>

  <Accordion title="Group by metadata">
    Group by a custom metadata key:

    ```python theme={"dark"}
    response = requests.post(
        "https://{your_control_plane_url}/api/svc/v1/llm-gateway/metrics/query",
        headers={
            "Authorization": "Bearer <your_api_key>",
            "Content-Type": "application/json"
        },
        json={
            "startTs": "2025-01-21T00:00:00.000Z",
            "endTs": "2025-01-22T00:00:00.000Z",
            "datasource": "modelMetrics",
            "type": "distribution",
            "aggregations": [],
            "filters": [
                {
                    "fieldName": "virtualModelName",
                    "operator": "IS_NULL",
                    "value": true
                }
            ],
            "groupBy": ["modelName", "metadata.environment"]
        }
    )
    ```
  </Accordion>

  <Accordion title="Filter by model name">
    Filter results to specific models:

    ```python theme={"dark"}
    response = requests.post(
        "https://{your_control_plane_url}/api/svc/v1/llm-gateway/metrics/query",
        headers={
            "Authorization": "Bearer <your_api_key>",
            "Content-Type": "application/json"
        },
        json={
            "startTs": "2025-01-21T00:00:00.000Z",
            "endTs": "2025-01-22T00:00:00.000Z",
            "datasource": "modelMetrics",
            "type": "distribution",
            "aggregations": [],
            "filters": [
                {
                    "fieldName": "virtualModelName",
                    "operator": "IS_NULL",
                    "value": true
                },
                {
                    "fieldName": "modelName",
                    "operator": "IN",
                    "value": ["gpt-4", "gpt-3.5-turbo", "claude-2"]
                }
            ],
            "groupBy": ["modelName"]
        }
    )
    ```
  </Accordion>

  <Accordion title="Filter high latency requests">
    Find requests with latency above a threshold:

    ```python theme={"dark"}
    response = requests.post(
        "https://{your_control_plane_url}/api/svc/v1/llm-gateway/metrics/query",
        headers={
            "Authorization": "Bearer <your_api_key>",
            "Content-Type": "application/json"
        },
        json={
            "startTs": "2025-01-21T00:00:00.000Z",
            "endTs": "2025-01-22T00:00:00.000Z",
            "datasource": "modelMetrics",
            "type": "distribution",
            "aggregations": [],
            "filters": [
                {
                    "fieldName": "virtualModelName",
                    "operator": "IS_NULL",
                    "value": true
                },
                {
                    "fieldName": "latencyMs",
                    "operator": "GREATER_THAN",
                    "value": 1000
                }
            ],
            "groupBy": ["modelName"]
        }
    )
    ```
  </Accordion>

  <Accordion title="Filter by latency range">
    Find requests within a latency range:

    ```python theme={"dark"}
    response = requests.post(
        "https://{your_control_plane_url}/api/svc/v1/llm-gateway/metrics/query",
        headers={
            "Authorization": "Bearer <your_api_key>",
            "Content-Type": "application/json"
        },
        json={
            "startTs": "2025-01-21T00:00:00.000Z",
            "endTs": "2025-01-22T00:00:00.000Z",
            "datasource": "modelMetrics",
            "type": "distribution",
            "aggregations": [],
            "filters": [
                {
                    "fieldName": "virtualModelName",
                    "operator": "IS_NULL",
                    "value": true
                },
                {
                    "fieldName": "latencyMs",
                    "operator": "BETWEEN",
                    "value": [500, 5000]
                }
            ],
            "groupBy": ["modelName"]
        }
    )
    ```
  </Accordion>

  <Accordion title="Filter by token counts">
    Filter by input and output token thresholds:

    ```python theme={"dark"}
    response = requests.post(
        "https://{your_control_plane_url}/api/svc/v1/llm-gateway/metrics/query",
        headers={
            "Authorization": "Bearer <your_api_key>",
            "Content-Type": "application/json"
        },
        json={
            "startTs": "2025-01-21T00:00:00.000Z",
            "endTs": "2025-01-22T00:00:00.000Z",
            "datasource": "modelMetrics",
            "type": "distribution",
            "aggregations": [],
            "filters": [
                {
                    "fieldName": "virtualModelName",
                    "operator": "IS_NULL",
                    "value": true
                },
                {
                    "fieldName": "inputTokens",
                    "operator": "GREATER_THAN",
                    "value": 100
                },
                {
                    "fieldName": "outputTokens",
                    "operator": "LESS_THAN_EQUAL",
                    "value": 1000
                }
            ],
            "groupBy": ["modelName"]
        }
    )
    ```
  </Accordion>

  <Accordion title="Filter by team">
    Filter to specific teams using array operators:

    ```python theme={"dark"}
    response = requests.post(
        "https://{your_control_plane_url}/api/svc/v1/llm-gateway/metrics/query",
        headers={
            "Authorization": "Bearer <your_api_key>",
            "Content-Type": "application/json"
        },
        json={
            "startTs": "2025-01-21T00:00:00.000Z",
            "endTs": "2025-01-22T00:00:00.000Z",
            "datasource": "modelMetrics",
            "type": "distribution",
            "aggregations": [],
            "filters": [
                {
                    "fieldName": "virtualModelName",
                    "operator": "IS_NULL",
                    "value": true
                },
                {
                    "fieldName": "team",
                    "operator": "ARRAY_HAS_ANY",
                    "value": ["team-alpha", "team-beta"]
                }
            ],
            "groupBy": ["team", "modelName"]
        }
    )
    ```
  </Accordion>

  <Accordion title="Filter by metadata">
    Filter by custom metadata values:

    ```python theme={"dark"}
    response = requests.post(
        "https://{your_control_plane_url}/api/svc/v1/llm-gateway/metrics/query",
        headers={
            "Authorization": "Bearer <your_api_key>",
            "Content-Type": "application/json"
        },
        json={
            "startTs": "2025-01-21T00:00:00.000Z",
            "endTs": "2025-01-22T00:00:00.000Z",
            "datasource": "modelMetrics",
            "type": "distribution",
            "aggregations": [],
            "filters": [
                {
                    "fieldName": "virtualModelName",
                    "operator": "IS_NULL",
                    "value": true
                },
                {
                    "metadataKey": "environment",
                    "operator": "IN",
                    "value": ["production"]
                }
            ],
            "groupBy": ["modelName"]
        }
    )
    ```
  </Accordion>

  <Accordion title="Complex filter combination">
    Combine multiple filter types:

    ```python theme={"dark"}
    response = requests.post(
        "https://{your_control_plane_url}/api/svc/v1/llm-gateway/metrics/query",
        headers={
            "Authorization": "Bearer <your_api_key>",
            "Content-Type": "application/json"
        },
        json={
            "startTs": "2025-01-21T00:00:00.000Z",
            "endTs": "2025-01-22T00:00:00.000Z",
            "datasource": "modelMetrics",
            "type": "distribution",
            "aggregations": [],
            "filters": [
                {
                    "fieldName": "virtualModelName",
                    "operator": "IS_NULL",
                    "value": true
                },
                {
                    "fieldName": "modelName",
                    "operator": "IN",
                    "value": ["gpt-4", "gpt-3.5-turbo"]
                },
                {
                    "fieldName": "latencyMs",
                    "operator": "BETWEEN",
                    "value": [100, 10000]
                },
                {
                    "fieldName": "inputTokens",
                    "operator": "GREATER_THAN",
                    "value": 50
                },
                {
                    "fieldName": "outputTokens",
                    "operator": "LESS_THAN",
                    "value": 2000
                }
            ],
            "groupBy": ["modelName"]
        }
    )
    ```
  </Accordion>

  <Accordion title="Virtual model metrics only">
    Query metrics for virtual models only. Use `IS_NULL` with value `false` to return only requests routed through a virtual model related metrics:

    ```python theme={"dark"}
    response = requests.post(
        "https://{your_control_plane_url}/api/svc/v1/llm-gateway/metrics/query",
        headers={
            "Authorization": "Bearer <your_api_key>",
            "Content-Type": "application/json"
        },
        json={
            "startTs": "2025-01-21T00:00:00.000Z",
            "endTs": "2025-01-22T00:00:00.000Z",
            "datasource": "modelMetrics",
            "type": "distribution",
            "aggregations": [
                {"type": "count", "column": "virtualModelName"},
                {"type": "sum", "column": "inputTokens"},
                {"type": "sum", "column": "outputTokens"}
            ],
            "filters": [
                {
                    "fieldName": "virtualModelName",
                    "operator": "IS_NULL",
                    "value": false
                }
            ],
            "groupBy": ["virtualModelName"]
        }
    )
    ```
  </Accordion>
</AccordionGroup>

### Timeseries Queries

<AccordionGroup>
  <Accordion title="Basic timeseries (hourly)">
    Get hourly request counts:

    ```python theme={"dark"}
    response = requests.post(
        "https://{your_control_plane_url}/api/svc/v1/llm-gateway/metrics/query",
        headers={
            "Authorization": "Bearer <your_api_key>",
            "Content-Type": "application/json"
        },
        json={
            "startTs": "2025-01-21T00:00:00.000Z",
            "endTs": "2025-01-22T00:00:00.000Z",
            "datasource": "modelMetrics",
            "type": "timeseries",
            "aggregations": [
                {"type": "count", "column": "modelName"}
            ],
            "filters": [
                {
                    "fieldName": "virtualModelName",
                    "operator": "IS_NULL",
                    "value": true
                }
            ],
            "intervalInSeconds": 3600
        }
    )
    ```
  </Accordion>

  <Accordion title="5-minute intervals">
    Get fine-grained metrics with 5-minute intervals:

    ```python theme={"dark"}
    response = requests.post(
        "https://{your_control_plane_url}/api/svc/v1/llm-gateway/metrics/query",
        headers={
            "Authorization": "Bearer <your_api_key>",
            "Content-Type": "application/json"
        },
        json={
            "startTs": "2025-01-21T00:00:00.000Z",
            "endTs": "2025-01-21T06:00:00.000Z",
            "datasource": "modelMetrics",
            "type": "timeseries",
            "aggregations": [
                {"type": "count", "column": "modelName"}
            ],
            "filters": [
                {
                    "fieldName": "virtualModelName",
                    "operator": "IS_NULL",
                    "value": true
                }
            ],
            "intervalInSeconds": 300
        }
    )
    ```
  </Accordion>

  <Accordion title="Timeseries by model">
    Get hourly counts grouped by model:

    ```python theme={"dark"}
    response = requests.post(
        "https://{your_control_plane_url}/api/svc/v1/llm-gateway/metrics/query",
        headers={
            "Authorization": "Bearer <your_api_key>",
            "Content-Type": "application/json"
        },
        json={
            "startTs": "2025-01-21T00:00:00.000Z",
            "endTs": "2025-01-22T00:00:00.000Z",
            "datasource": "modelMetrics",
            "type": "timeseries",
            "aggregations": [
                {"type": "count", "column": "modelName"}
            ],
            "filters": [
                {
                    "fieldName": "virtualModelName",
                    "operator": "IS_NULL",
                    "value": true
                }
            ],
            "groupBy": ["modelName"],
            "intervalInSeconds": 3600
        }
    )
    ```
  </Accordion>

  <Accordion title="Timeseries by team">
    Get hourly counts grouped by team:

    ```python theme={"dark"}
    response = requests.post(
        "https://{your_control_plane_url}/api/svc/v1/llm-gateway/metrics/query",
        headers={
            "Authorization": "Bearer <your_api_key>",
            "Content-Type": "application/json"
        },
        json={
            "startTs": "2025-01-21T00:00:00.000Z",
            "endTs": "2025-01-22T00:00:00.000Z",
            "datasource": "modelMetrics",
            "type": "timeseries",
            "aggregations": [
                {"type": "count", "column": "team"}
            ],
            "filters": [
                {
                    "fieldName": "virtualModelName",
                    "operator": "IS_NULL",
                    "value": true
                }
            ],
            "groupBy": ["team"],
            "intervalInSeconds": 3600
        }
    )
    ```
  </Accordion>

  <Accordion title="Timeseries with filters">
    Apply filters to timeseries data:

    ```python theme={"dark"}
    response = requests.post(
        "https://{your_control_plane_url}/api/svc/v1/llm-gateway/metrics/query",
        headers={
            "Authorization": "Bearer <your_api_key>",
            "Content-Type": "application/json"
        },
        json={
            "startTs": "2025-01-21T00:00:00.000Z",
            "endTs": "2025-01-22T00:00:00.000Z",
            "datasource": "modelMetrics",
            "type": "timeseries",
            "aggregations": [
                {"type": "count", "column": "modelName"}
            ],
            "filters": [
                {
                    "fieldName": "virtualModelName",
                    "operator": "IS_NULL",
                    "value": true
                },
                {
                    "fieldName": "modelName",
                    "operator": "IN",
                    "value": ["gpt-4", "gpt-3.5-turbo"]
                }
            ],
            "groupBy": ["modelName"],
            "intervalInSeconds": 3600
        }
    )
    ```
  </Accordion>

  <Accordion title="Timeseries with latency filter">
    Filter timeseries by latency threshold:

    ```python theme={"dark"}
    response = requests.post(
        "https://{your_control_plane_url}/api/svc/v1/llm-gateway/metrics/query",
        headers={
            "Authorization": "Bearer <your_api_key>",
            "Content-Type": "application/json"
        },
        json={
            "startTs": "2025-01-21T00:00:00.000Z",
            "endTs": "2025-01-22T00:00:00.000Z",
            "datasource": "modelMetrics",
            "type": "timeseries",
            "aggregations": [
                {"type": "count", "column": "modelName"}
            ],
            "filters": [
                {
                    "fieldName": "virtualModelName",
                    "operator": "IS_NULL",
                    "value": true
                },
                {
                    "fieldName": "latencyMs",
                    "operator": "GREATER_THAN",
                    "value": 500
                }
            ],
            "groupBy": ["modelName"],
            "intervalInSeconds": 3600
        }
    )
    ```
  </Accordion>

  <Accordion title="Timeseries with team filter">
    Filter timeseries by team:

    ```python theme={"dark"}
    response = requests.post(
        "https://{your_control_plane_url}/api/svc/v1/llm-gateway/metrics/query",
        headers={
            "Authorization": "Bearer <your_api_key>",
            "Content-Type": "application/json"
        },
        json={
            "startTs": "2025-01-21T00:00:00.000Z",
            "endTs": "2025-01-22T00:00:00.000Z",
            "datasource": "modelMetrics",
            "type": "timeseries",
            "aggregations": [
                {"type": "count", "column": "modelName"}
            ],
            "filters": [
                {
                    "fieldName": "virtualModelName",
                    "operator": "IS_NULL",
                    "value": true
                },
                {
                    "fieldName": "team",
                    "operator": "ARRAY_HAS_ANY",
                    "value": ["team-alpha", "team-beta"]
                }
            ],
            "groupBy": ["team"],
            "intervalInSeconds": 3600
        }
    )
    ```
  </Accordion>

  <Accordion title="Timeseries by metadata">
    Group timeseries by metadata:

    ```python theme={"dark"}
    response = requests.post(
        "https://{your_control_plane_url}/api/svc/v1/llm-gateway/metrics/query",
        headers={
            "Authorization": "Bearer <your_api_key>",
            "Content-Type": "application/json"
        },
        json={
            "startTs": "2025-01-21T00:00:00.000Z",
            "endTs": "2025-01-22T00:00:00.000Z",
            "datasource": "modelMetrics",
            "type": "timeseries",
            "aggregations": [
                {"type": "count", "column": "modelName"}
            ],
            "filters": [
                {
                    "fieldName": "virtualModelName",
                    "operator": "IS_NULL",
                    "value": true
                }
            ],
            "groupBy": ["metadata.environment"],
            "intervalInSeconds": 3600
        }
    )
    ```
  </Accordion>

  <Accordion title="Weekly timeseries">
    Get daily data for a week:

    ```python theme={"dark"}
    response = requests.post(
        "https://{your_control_plane_url}/api/svc/v1/llm-gateway/metrics/query",
        headers={
            "Authorization": "Bearer <your_api_key>",
            "Content-Type": "application/json"
        },
        json={
            "startTs": "2025-01-14T00:00:00.000Z",
            "endTs": "2025-01-21T00:00:00.000Z",
            "datasource": "modelMetrics",
            "type": "timeseries",
            "aggregations": [
                {"type": "count", "column": "modelName"}
            ],
            "filters": [
                {
                    "fieldName": "virtualModelName",
                    "operator": "IS_NULL",
                    "value": true
                }
            ],
            "intervalInSeconds": 86400
        }
    )
    ```
  </Accordion>

  <Accordion title="Complex timeseries query">
    Combine filters, groupBy, and metadata:

    ```python theme={"dark"}
    response = requests.post(
        "https://{your_control_plane_url}/api/svc/v1/llm-gateway/metrics/query",
        headers={
            "Authorization": "Bearer <your_api_key>",
            "Content-Type": "application/json"
        },
        json={
            "startTs": "2025-01-21T00:00:00.000Z",
            "endTs": "2025-01-22T00:00:00.000Z",
            "datasource": "modelMetrics",
            "type": "timeseries",
            "aggregations": [
                {"type": "count", "column": "modelName"}
            ],
            "filters": [
                {
                    "fieldName": "virtualModelName",
                    "operator": "IS_NULL",
                    "value": true
                },
                {
                    "fieldName": "modelName",
                    "operator": "IN",
                    "value": ["gpt-4", "gpt-3.5-turbo"]
                },
                {
                    "metadataKey": "environment",
                    "operator": "IN",
                    "value": ["production"]
                }
            ],
            "groupBy": ["modelName"],
            "intervalInSeconds": 3600
        }
    )
    ```
  </Accordion>
</AccordionGroup>

## Response Format

The API returns metrics data in JSON format. Aggregation results are returned with keys in camelCase format: `{aggregationType}{ColumnName}` where the column name is capitalized (e.g., `countModelName`, `sumInputTokens`, `p99LatencyMs`).

### Distribution Response

```json theme={"dark"}
{
    "data": [
        {
            "modelName": "gpt-4",
            "countModelName": 150,
            "sumInputTokens": 125000,
            "sumOutputTokens": 45000,
            "p99LatencyMs": 2450.5
        },
        {
            "modelName": "gpt-3.5-turbo",
            "countModelName": 320,
            "sumInputTokens": 89000,
            "sumOutputTokens": 32000,
            "p99LatencyMs": 1820.3
        }
    ]
}
```

### Timeseries Response

```json theme={"dark"}
{
    "data": [
        {
            "timestamp": "2025-01-21T00:00:00.000Z",
            "modelName": "gpt-4",
            "countModelName": 25,
            "sumInputTokens": 15000,
            "p99LatencyMs": 2100.5
        },
        {
            "timestamp": "2025-01-21T01:00:00.000Z",
            "modelName": "gpt-4",
            "countModelName": 30,
            "sumInputTokens": 18500,
            "p99LatencyMs": 2350.2
        }
    ]
}
```

<Info>
  If the `groupBy` array is empty, the API returns a summarized overview of all requests within the specified time range.
</Info>
