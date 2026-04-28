> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Caching (Exact and Semantic)

> Speed up responses and cut costs with gateway-level exact-match and semantic caching.

TrueFoundry AI Gateway caches LLM responses so that repeated or similar queries are served instantly without calling the model provider. The gateway supports two caching strategies:

* **Exact-Match** — returns cached responses only when the request is identical
* **Semantic** — returns cached responses when requests are semantically similar, even if worded differently

## Prompt Caching vs Gateway Caching

LLM providers and the TrueFoundry AI Gateway offer caching at different layers. Understanding the difference helps you pick the right approach — or combine both for maximum savings.

|                      | Provider Prompt Caching                                                               | Gateway Caching (Exact-Match & Semantic)                                                          |
| -------------------- | ------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------- |
| **Where it runs**    | Inside the model provider (e.g. Anthropic)                                            | In the TrueFoundry AI Gateway, before the request reaches any provider                            |
| **What is cached**   | The prompt prefix — the provider skips re-processing tokens it has already seen       | The complete LLM response — a cache hit returns the response instantly with zero model invocation |
| **How it matches**   | Exact token-prefix match on the prompt                                                | Exact request hash (exact-match) or embedding cosine similarity (semantic)                        |
| **Latency savings**  | Reduces time-to-first-token; the model still generates a new completion               | Eliminates model call entirely — response is returned from cache in milliseconds                  |
| **Cost savings**     | Cached input tokens are billed at a reduced rate (varies by provider)                 | Cache hit = zero model cost for that request                                                      |
| **Provider support** | Provider-specific (see [Prompt Caching](/ai-gateway/chat-completions-prompt-caching)) | Works with every provider routed through the gateway                                              |

## Cache Types

### Exact-Match Cache

Exact-match caching stores responses keyed by a hash of the complete request — messages, model, and all parameters. A cached response is returned only when every part of the request matches exactly.

**Best for:**

* API calls with identical parameters
* Deterministic queries that always need the same response
* Development and testing environments
* Applications with predictable, repetitive queries

### Semantic Cache

Semantic caching uses embeddings and cosine similarity to match requests that express the same intent, even when worded differently. For example, *"How do I reset my password?"* and *"What's the password reset process?"* would match.

The gateway extracts the last message, generates an embedding, and compares it against cached embeddings. All other request parameters (model, prior messages, temperature, etc.) are hashed and must still match exactly — only the last message is compared semantically.

**Best for:**

* Customer support chatbots
* FAQ systems where users phrase questions differently
* Conversational AI applications
* Any scenario where query variations express the same intent

<Note>
  Semantic cache is a superset of exact-match cache. Setting the cache type to `semantic` will also return results for exact-match hits, so you don't need to configure both.
</Note>

<AccordionGroup>
  <Accordion title="How to control the similarity required for a cache hit?">
    The `similarity_threshold` parameter (0 – 1.0) controls how close two queries must be for the cached response to be returned. A higher value demands closer matches; a lower value allows broader matching.

    | Range           | Behaviour                                                  | Recommended for                                                |
    | --------------- | ---------------------------------------------------------- | -------------------------------------------------------------- |
    | **0.95 – 1.0**  | Very strict — only nearly identical queries match          | High-precision use cases where incorrect cache hits are costly |
    | **0.85 – 0.95** | Balanced — works well for most conversational applications | General-purpose chatbots and FAQ systems                       |
    | **\< 0.85**     | Broad — may return results for loosely related queries     | Exploratory or low-risk scenarios                              |

    <Tip>
      Start with a threshold of **0.9** and monitor cache hit rates. Adjust up for precision or down for coverage.
    </Tip>
  </Accordion>

  <Accordion title="How is similarity computed?">
    The gateway converts the last message of each request into a vector embedding using an embedding model. When a new request arrives, its embedding is compared against cached embeddings using **cosine similarity** — a score between 0 and 1.0 that measures how close two vectors are in meaning. If the score meets or exceeds the configured `similarity_threshold`, the cached response is returned.

    All other request parameters (model, prior messages, temperature, etc.) are hashed separately and must match exactly — only the last message is compared semantically.
  </Accordion>

  <Accordion title="Which embedding model is used?">
    The embedding model that generates these vectors depends on your deployment type:

    * **SaaS** — Uses **OpenAI text-embedding-3-small** by default. The embedding model is not configurable in the SaaS version.
    * **On-Premise** — Configure the embedding model through **Controls > Settings > Semantic Cache** in the gateway dashboard. The selected model applies to all semantic cache operations gateway-wide.

    <Frame>
      <img src="https://mintcdn.com/truefoundry/bng6_67kuXVrAlwQ/images/semantic-cache-settings.png?fit=max&auto=format&n=bng6_67kuXVrAlwQ&q=85&s=8fffe2cf463217a63083ac0403f6118e" alt="Semantic Cache Embedding Model Settings showing the Configure Semantic Caching dialog with an embedding model selector" width="2524" height="1212" data-path="images/semantic-cache-settings.png" />
    </Frame>
  </Accordion>
</AccordionGroup>

***

## Namespacing and Cache Isolation

The gateway isolates cache entries at two levels to prevent data leaking across boundaries.

### Level 1 — User / Virtual Account (automatic)

Every cache entry is scoped to the **user or virtual account** that created it. This happens automatically — you don't need to configure anything. A cache entry created by User A is never visible to User B, even if they send the exact same request.

### Level 2 — Custom Namespace (optional)

Within a user's or virtual account's cache, you can further partition entries by providing a `namespace` string. Entries in one namespace are invisible to requests with a different namespace (or no namespace).

This is useful when a single virtual account serves multiple downstream end-users or application contexts and you want per-context cache isolation. For example, an application that serves many tenants through a single virtual account can set `namespace` to each tenant's ID so that cached responses never cross tenant boundaries.

| Scenario                         | Namespace value                                                | Effect                                                                     |
| -------------------------------- | -------------------------------------------------------------- | -------------------------------------------------------------------------- |
| Single application, shared cache | *omit or set to `"default"`*                                   | All requests under the same user/virtual account share a single cache pool |
| Multi-tenant application         | Set to the tenant or end-user ID (e.g. `"tenant-123"`)         | Each tenant gets its own isolated cache within the same virtual account    |
| Multiple environments            | Set to the environment name (e.g. `"staging"`, `"production"`) | Prevents staging queries from returning production-cached responses        |

```json theme={"dark"}
{
  "type": "semantic",
  "similarity_threshold": 0.9,
  "ttl": 600,
  "namespace": "tenant-123"
}
```

***

## Configuration

Enable caching by adding the `x-tfy-cache-config` header to your requests.

### Examples

<Note>
  Replace `{GATEWAY_BASE_URL}` with the base URL of the TrueFoundry AI Gateway and `your-truefoundry-api-key` with your API key.
</Note>

<CodeGroup>
  ```bash Exact-Match Cache theme={"dark"}
  curl {GATEWAY_BASE_URL}/chat/completions \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer your-truefoundry-api-key" \
    -H "x-tfy-cache-config: {\"type\": \"exact-match\", \"ttl\": 600}" \
    -d '{
      "model": "openai-main/gpt-4o",
      "messages": [
        {
          "role": "user",
          "content": "What is the capital of France?"
        }
      ]
    }'
  ```

  ```bash Semantic Cache theme={"dark"}
  curl {GATEWAY_BASE_URL}/chat/completions \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer your-truefoundry-api-key" \
    -H "x-tfy-cache-config: {\"type\": \"semantic\", \"similarity_threshold\": 0.9, \"ttl\": 600}" \
    -d '{
      "model": "openai-main/gpt-4o",
      "messages": [
        {
          "role": "user",
          "content": "How do I reset my password?"
        }
      ]
    }'
  ```

  ```bash Semantic Cache with Namespace theme={"dark"}
  curl {GATEWAY_BASE_URL}/chat/completions \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer your-truefoundry-api-key" \
    -H "x-tfy-cache-config: {\"type\": \"semantic\", \"similarity_threshold\": 0.9, \"ttl\": 600, \"namespace\": \"tenant-123\"}" \
    -d '{
      "model": "openai-main/gpt-4o",
      "messages": [
        {
          "role": "user",
          "content": "How do I reset my password?"
        }
      ]
    }'
  ```
</CodeGroup>

***

## Response Headers

The gateway returns headers indicating cache status on every response:

| Header                         | Description                                                              | Example Values         |
| ------------------------------ | ------------------------------------------------------------------------ | ---------------------- |
| `x-tfy-cache-status`           | Whether the cache was hit                                                | `hit`, `miss`, `error` |
| `x-tfy-cached-trace-id`        | Trace ID of the original request that populated the cache (only on hits) | `trace_abc123`         |
| `x-tfy-cache-similarity-score` | Cosine similarity score (semantic cache hits only)                       | `0.95`                 |

***

## How It Works

<Tabs>
  <Tab title="Exact-Match">
    <Steps>
      <Step title="Hash the request">
        The gateway generates a hash of the complete request (messages, model, parameters).
      </Step>

      <Step title="Look up the cache">
        If a cached response exists for this hash, it is returned immediately.
      </Step>

      <Step title="Forward on miss">
        On a cache miss, the request is forwarded to the model provider. The response is cached before being returned.
      </Step>
    </Steps>
  </Tab>

  <Tab title="Semantic">
    <Steps>
      <Step title="Extract and embed">
        The gateway extracts the last message and generates an embedding vector. All other request parameters are hashed.
      </Step>

      <Step title="Compare cached embeddings">
        The embedding is compared against cached embeddings using cosine similarity. The parameter hash must also match exactly.
      </Step>

      <Step title="Return or forward">
        If the similarity score exceeds the configured threshold and the parameter hash matches, the cached response is returned. Otherwise, the request is forwarded to the model provider and both the response and embedding are cached.
      </Step>
    </Steps>
  </Tab>
</Tabs>

***

## Infrastructure Setup

For **SaaS** customers, caching infrastructure is fully managed — no additional setup is required.

For **On-Premise** deployments, caching requires two infrastructure components:

1. **Redis** — cache store for both exact-match and semantic caching
2. **Embedding Model** — required for semantic caching to generate vector embeddings (see [Embedding Model](#embedding-model))

### Redis Setup

You can either enable the bundled Redis instance in the `tfy-llm-gateway` Helm chart, or connect your own Redis (or Redis-compatible store such as [Valkey](https://valkey.io/)).

<Tabs>
  <Tab title="Bundled Redis">
    Enable Redis directly in the `tfy-llm-gateway` chart values:

    ```yaml theme={"dark"}
    tfy-llm-gateway:
      redis:
        enabled: true
    ```
  </Tab>

  <Tab title="Bring Your Own Redis">
    Disable the bundled Redis and configure connection details via environment variables:

    ```yaml theme={"dark"}
    tfy-llm-gateway:
      redis:
        enabled: false
      env:
        REDIS_HOST: "<your-redis-host>"
        REDIS_PORT: 6379
        REDIS_USERNAME: "<your-redis-username>"
        REDIS_PASSWORD: "<your-redis-password>"
    ```

    | Variable         | Required | Type     | Description                           |
    | ---------------- | -------- | -------- | ------------------------------------- |
    | `REDIS_HOST`     | Yes      | `string` | Hostname or IP of your Redis instance |
    | `REDIS_PORT`     | Yes      | `number` | Port number (default: `6379`)         |
    | `REDIS_USERNAME` | No       | `string` | Username for Redis authentication     |
    | `REDIS_PASSWORD` | No       | `string` | Password for Redis authentication     |
  </Tab>
</Tabs>
