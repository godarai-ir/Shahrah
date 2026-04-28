> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Request Headers & Metadata

> Configure request headers for authentication, metadata tagging, logging, retries, and timeouts in the AI Gateway.

The AI Gateway uses HTTP headers to control authentication, request routing, logging, retries, and metadata tagging. This page covers all available request and response headers, along with custom metadata and logging configuration.

## Request Headers

| Name                    | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                  | Example                                                                   |
| ----------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------- |
| `Authorization`         | Your TrueFoundry API key as bearer token                                                                                                                                                                                                                                                                                                                                                                                                                                     | `Authorization: Bearer TFY_API_KEY`                                       |
| `x-tfy-metadata`        | Stringified JSON where both keys and values must be strings. Used for request routing and metrics filtering                                                                                                                                                                                                                                                                                                                                                                  | `x-tfy-metadata: {"custom_field":"value"}`                                |
| `x-tfy-provider-name`   | Required for responses API, file upload API, and batch APIs to route requests to the correct provider account                                                                                                                                                                                                                                                                                                                                                                | `x-tfy-provider-name: openai`                                             |
| `x-tfy-strict-openai`   | Boolean flag to enable strict OpenAI compatibility (set to `false` for Claude reasoning model responses with thinking tokens)                                                                                                                                                                                                                                                                                                                                                | `x-tfy-strict-openai: true`                                               |
| `x-tfy-retry-config`    | JSON object to configure retry behavior for failed requests                                                                                                                                                                                                                                                                                                                                                                                                                  | `x-tfy-retry-config: {"attempts": 3, "onStatusCodes": [429, 500, 503]}`   |
| `x-tfy-request-timeout` | Number in milliseconds specifying the maximum time to wait for a response from a single model. If fallbacks or retries are configured, the timeout is applied per model request (i.e., each attempt, including fallbacks, will have its own timeout).                                                                                                                                                                                                                        | `x-tfy-request-timeout: 60000`                                            |
| `x-tfy-ttft-timeout-ms` | Number in milliseconds specifying the maximum time to wait for the first token in a streaming response (time-to-first-token). If no token is received within this window, the request is considered timed out and the gateway returns **408**. For [virtual models](/docs/ai-gateway/virtual-model) or [routing config](/docs/ai-gateway/load-balancing-overview), the gateway falls back to the next model on 408 even if 408 is not included in the fallback status codes. | `x-tfy-ttft-timeout-ms: 30000`                                            |
| `x-tfy-logging-config`  | Configuration for request logging                                                                                                                                                                                                                                                                                                                                                                                                                                            | `x-tfy-logging-config: {"enabled": true}`                                 |
| `x-tfy-mcp-headers`     | Stringified JSON to pass custom headers to MCP servers. Format varies by API — see [MCP Gateway](/docs/ai-gateway/mcp/mcp-gateway-auth-security#passing-custom-headers) and [Agent API](/docs/ai-gateway/agents/use-mcp-server-in-code-agent#method-2:-using-x-tfy-mcp-headers-header-registered-servers-only) docs. Agent API only supports registered servers.                                                                                                             | `x-tfy-mcp-headers: {"truefoundry:...":{"Authorization":"Bearer TOKEN"}}` |

## Custom Metadata

You can tag requests with custom metadata using the `X-TFY-METADATA` header. Metadata is a JSON object where both keys and values must be strings, with a maximum value length of 128 characters.

With metadata, you can:

* **Enhance Observability**: Filter request logs and create custom metrics dashboards grouped by metadata keys.
* **Apply Conditional Configurations**: Use metadata in the `when` block of gateway configurations to selectively apply rate limiting, model fallbacks, load balancing, and more.

### Sending Metadata

<CodeGroup>
  ```python OpenAI SDK highlight={12-14} wrap lines theme={"dark"}
  from openai import OpenAI

  USE_STREAM = True
  client = OpenAI(api_key="your_truefoundry_api_key", base_url="https://gateway.truefoundry.ai")
  stream = client.chat.completions.create(
      messages=[
          {"role": "system", "content": "You are an AI bot."},
          {"role": "user", "content": "Enter your prompt here"},
      ],
      model="openai-main/gpt-4",
      stream=USE_STREAM,
      extra_headers={
          "X-TFY-METADATA": '{"application":"booking-bot", "environment":"staging", "customer_id":"123456"}',
      }
  )

  if USE_STREAM:
      for chunk in stream:
          if (
              chunk.choices
              and len(chunk.choices) > 0
              and chunk.choices[0].delta.content is not None
          ):
              print(chunk.choices[0].delta.content, end="")
  else:
      print(stream.choices[0].message.content)
  ```

  ```typescript Node.js highlight={6-8} wrap lines theme={"dark"}
  const OpenAI = require('openai');

  const openai = new OpenAI({
      apiKey: 'your_truefoundry_api_key',
      baseURL: 'https://gateway.truefoundry.ai',
      defaultHeaders: {
          "X-TFY-METADATA":
          '{"application":"booking-bot", "environment":"staging", "customer_id":"123456"}',
      },
  });

  async function main() {
      try {
          const stream = await openai.chat.completions.create({
              model: 'openai-main/gpt-4',
              messages: [
                  { role: 'system', content: 'You are an AI bot.' },
                  { role: 'user', content: 'Enter your prompt here' }
              ],
              stream: true
          });

          for await (const chunk of stream) {
              if (chunk.choices[0]?.delta?.content) {
                  process.stdout.write(chunk.choices[0].delta.content);
              }
          }
      } catch (error) {
          console.error('An error occurred:', error.message);
      }
  }

  main();
  ```
</CodeGroup>

### Using Metadata for Observability

#### Filter Logs

Filter your request logs using one or more metadata keys to isolate specific requests. This is useful for debugging or analyzing usage patterns for a particular feature, environment, or user.

<Frame caption="Filtering requests using custom Metadata Keys">
  <img src="https://mintcdn.com/truefoundry/FeKcq2n1MMm83Par/images/Screenshot2025-07-21at6.57.29PM-min.png?fit=max&auto=format&n=FeKcq2n1MMm83Par&q=85&s=0992cf1c5ae5c79d1cecc4c7f15fe315" alt="TrueFoundry dashboard showing how to filter requests using custom metadata keys" width="3588" height="2004" data-path="images/Screenshot2025-07-21at6.57.29PM-min.png" />
</Frame>

#### Create Custom Metrics

Group metrics by metadata keys to create custom visualizations. For example, monitor cost and usage per customer by grouping with a `customer_id` key.

<Frame caption="Grouping metrics by a custom 'customer_id' metadata key">
  <img src="https://mintcdn.com/truefoundry/FeKcq2n1MMm83Par/images/Screenshot2025-07-21at6.29.09PM-min.png?fit=max&auto=format&n=FeKcq2n1MMm83Par&q=85&s=b45ab05b7fe4eaba48dbc662516b01de" alt="Dashboard view showing metrics grouped by customer_id metadata key with usage statistics per customer" width="3600" height="2004" data-path="images/Screenshot2025-07-21at6.29.09PM-min.png" />
</Frame>

### Using Metadata in Gateway Configurations

Metadata can be used in the `when` block of gateway configurations to selectively apply rules. For example, to rate limit requests from the `dev` environment:

```yaml lines theme={"dark"}
name: ratelimiting-config
type: gateway-rate-limiting-config
rules:
  - id: 'openai-gpt4-dev-env'
    when:
      models: ['openai-main/gpt4']
      metadata:
        env: dev
    limit_to: 1000
    unit: requests_per_day
```

You can also use metadata to configure **Load Balancing** and **Fallbacks**. Learn more in [Virtual Models](/docs/ai-gateway/virtual-model#retries-and-fallbacks) and [Routing Config](/docs/ai-gateway/load-balancing-overview).

## Logging Configuration

Control whether individual requests are logged using the `X-TFY-LOGGING-CONFIG` header.

### Enable Logging

```python lines theme={"dark"}
from openai import OpenAI

client = OpenAI(
    api_key="your_truefoundry_api_key",
    base_url="https://gateway.truefoundry.ai",
    default_headers={
        "X-TFY-LOGGING-CONFIG": '{"enabled": true}'
    }
)
```

### Disable Logging

To prevent a request from being logged, set `enabled` to `false`:

```python lines theme={"dark"}
client = OpenAI(
    api_key="your_truefoundry_api_key",
    base_url="https://gateway.truefoundry.ai",
    default_headers={
        "X-TFY-LOGGING-CONFIG": '{"enabled": false}'
    }
)
```

### Server-Side Logging Mode (Self-Hosted Only)

<Note>
  Environment variable configuration is only available when running a self-hosted instance of TrueFoundry AI Gateway.
</Note>

You can control logging behavior globally by setting the `REQUEST_LOGGING_MODE` environment variable:

| Mode                | Description                                                                                                                                                                           |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `HEADER_CONTROLLED` | Logging depends on the `enabled` value in the `X-TFY-LOGGING-CONFIG` header. If the header is absent or set to `true`, logging will occur. If set to `false`, no logging will happen. |
| `ALWAYS`            | All requests are logged regardless of the `enabled` value.                                                                                                                            |
| `NEVER`             | No requests are logged regardless of the `enabled` value.                                                                                                                             |

To view logged requests, go to **AI Gateway > Monitor > Requests** in the TrueFoundry UI. See [View Traces](/docs/ai-gateway/request-logging) for details.

## Response Headers

| Name                           | Description                                                                                                                                     |
| ------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| `x-tfy-resolved-model`         | The final TrueFoundry model ID used to process the request (may differ from requested model due to load balancing or fallbacks)                 |
| `x-tfy-applied-configurations` | Dictionary of applied configurations including load balancing, fallback, model config, applied guardrails, and rate limiting                    |
| `server-timing`                | For non-streaming requests only. Contains timing information for different processing stages including middlewares, guardrails, and model calls |

### Server-Timing Breakdown

When inspecting network requests in your browser's developer tools, the `server-timing` header provides a detailed performance breakdown:

| Processing Stage  | Duration    | Description                                         |
| ----------------- | ----------- | --------------------------------------------------- |
| Authentication    | 0.9 ms      | Authenticating User                                 |
| Input guardrails  | 0.7 ms      | Input validation and content filtering              |
| Model call        | 1350 ms     | AI model response generation (bulk of the time)     |
| Output guardrails | 722.3 ms    | Output validation and filtering                     |
| Logging           | 1.1 ms      | Logging request                                     |
| **Total**         | **2080 ms** | **Complete request processing time (2.08 seconds)** |

Metrics like `load balancing (0 ms)`, `rate limiting (0 ms)`, and `cost budget (0 ms)` show zero duration because these configs weren't triggered for this particular request.

<Frame caption="Server-timing header in browser developer tools">
  <img src="https://mintcdn.com/truefoundry/QPyNP8qMPAoy2O9d/images/gateway-server-timing.png?fit=max&auto=format&n=QPyNP8qMPAoy2O9d&q=85&s=676cbae17c2be4809a7f6e5cfc915b2b" alt="Server-timing header in browser developer tools showing request processing breakdown" width="3024" height="1588" data-path="images/gateway-server-timing.png" />
</Frame>
