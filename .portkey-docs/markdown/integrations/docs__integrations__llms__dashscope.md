> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Dashscope

> Integrate Dashscope with Portkey for seamless completions, embeddings, and advanced features.

## Quick Start

Get started with Dashscope in under 2 minutes:

<CodeGroup>
  ```python Python icon="python" theme={"system"}
  from portkey_ai import Portkey

  # 1. Install: pip install portkey-ai
  # 2. Add @dashscope provider in model catalog
  # 3. Use it:

  portkey = Portkey(api_key="PORTKEY_API_KEY")

  response = portkey.chat.completions.create(
      model="@dashscope/qwen-turbo",
      messages=[{"role": "user", "content": "Hello!"}]
  )

  print(response.choices[0].message.content)
  ```

  ```js Javascript icon="square-js" theme={"system"}
  import Portkey from 'portkey-ai'

  // 1. Install: npm install portkey-ai
  // 2. Add @dashscope provider in model catalog
  // 3. Use it:

  const portkey = new Portkey({
      apiKey: "PORTKEY_API_KEY"
  })

  const response = await portkey.chat.completions.create({
      model: "@dashscope/qwen-turbo",
      messages: [{ role: "user", content: "Hello!" }]
  })

  console.log(response.choices[0].message.content)
  ```

  ```python OpenAI Py icon="python" theme={"system"}
  from openai import OpenAI
  from portkey_ai import PORTKEY_GATEWAY_URL

  # 1. Install: pip install openai portkey-ai
  # 2. Add @dashscope provider in model catalog
  # 3. Use it:

  client = OpenAI(
      api_key="PORTKEY_API_KEY",  # Portkey API key
      base_url=PORTKEY_GATEWAY_URL
  )

  response = client.chat.completions.create(
      model="@dashscope/qwen-turbo",
      messages=[{"role": "user", "content": "Hello!"}]
  )

  print(response.choices[0].message.content)
  ```

  ```js OpenAI JS icon="square-js" theme={"system"}
  import OpenAI from "openai"
  import { PORTKEY_GATEWAY_URL } from "portkey-ai"

  // 1. Install: npm install openai portkey-ai
  // 2. Add @dashscope provider in model catalog
  // 3. Use it:

  const client = new OpenAI({
      apiKey: "PORTKEY_API_KEY",  // Portkey API key
      baseURL: PORTKEY_GATEWAY_URL
  })

  const response = await client.chat.completions.create({
      model: "@dashscope/qwen-turbo",
      messages: [{ role: "user", content: "Hello!" }]
  })

  console.log(response.choices[0].message.content)
  ```

  ```sh cURL icon="square-terminal" theme={"system"}
  # 1. Add @dashscope provider in model catalog
  # 2. Use it:

  curl https://api.portkey.ai/v1/chat/completions \
    -H "Content-Type: application/json" \
    -H "x-portkey-api-key: $PORTKEY_API_KEY" \
    -d '{
      "model": "@dashscope/qwen-turbo",
      "messages": [{"role": "user", "content": "Hello!"}]
    }'
  ```
</CodeGroup>

## Add Provider in Model Catalog

Before making requests, add Dashscope to your Model Catalog:

1. Go to [**Model Catalog → Add Provider**](https://app.portkey.ai/model-catalog/providers)
2. Select **Dashscope**
3. Enter your [Dashscope API key](https://help.aliyun.com/zh/model-studio/developer-reference/get-api-key)
4. Name your provider (e.g., `dashscope`)

<Card title="Complete Setup Guide" icon="book" href="/product/model-catalog">
  See all setup options and detailed configuration instructions
</Card>

<Card title="Dashscope Documentation" icon="book" href="https://help.aliyun.com/zh/model-studio/developer-reference/compatibility-of-openai-with-dashscope">
  Explore the official Dashscope documentation
</Card>

***

## Dashscope Capabilities

### Embeddings

Generate embeddings for text using Dashscope embedding models:

<CodeGroup>
  ```python Python theme={"system"}
  from portkey_ai import Portkey

  portkey = Portkey(api_key="PORTKEY_API_KEY", provider="@dashscope")

  response = portkey.embeddings.create(
      input="Your text string goes here",
      model="text-embedding-v3"
  )

  print(response.data[0].embedding)
  ```

  ```javascript Node.js theme={"system"}
  import Portkey from 'portkey-ai';

  const portkey = new Portkey({
      apiKey: 'PORTKEY_API_KEY',
      provider: '@dashscope'
  });

  const response = await portkey.embeddings.create({
      input: "Your text string goes here",
      model: "text-embedding-v3"
  });

  console.log(response.data[0].embedding);
  ```
</CodeGroup>

***

## Advanced Features

### Track End-User IDs

Monitor user-level costs and requests by passing user IDs:

<CodeGroup>
  ```python Python theme={"system"}
  from portkey_ai import Portkey

  portkey = Portkey(api_key="PORTKEY_API_KEY", provider="@dashscope")

  response = portkey.chat.completions.create(
      model="qwen-turbo",
      messages=[{"role": "user", "content": "Hello!"}],
      user="user_123456"  # Track this user's usage
  )
  ```

  ```javascript Node.js theme={"system"}
  import Portkey from 'portkey-ai';

  const portkey = new Portkey({
      apiKey: 'PORTKEY_API_KEY',
      provider: '@dashscope'
  });

  const response = await portkey.chat.completions.create({
      model: "qwen-turbo",
      messages: [{ role: "user", content: "Hello!" }],
      user: "user_123456"  // Track this user's usage
  });
  ```
</CodeGroup>

<img src="https://mintcdn.com/portkey-docs/wAHXB_jjwLt8bYcN/images/llms/logs.png?fit=max&auto=format&n=wAHXB_jjwLt8bYcN&q=85&s=5bb5f6c00cfbb4510128f772319ea814" alt="Portkey Logs with User ID" width="968" height="668" data-path="images/llms/logs.png" />

<Card title="Learn More About Metadata" icon="tags" href="/product/observability/metadata">
  Explore how to use custom metadata to enhance your request tracking and analysis
</Card>

### Gateway Configurations

Use Portkey's Gateway features for advanced routing and reliability:

**Example: Conditional Routing**

```json theme={"system"}
{
  "strategy": {
    "mode": "conditional",
    "conditions": [
      {
        "query": { "metadata.user_plan": { "$eq": "paid" } },
        "then": "qwen-turbo"
      },
      {
        "query": { "metadata.user_plan": { "$eq": "free" } },
        "then": "gpt-3.5"
      }
    ],
    "default": "gpt-3.5"
  },
  "targets": [
    {
      "name": "qwen-turbo",
      "provider": "@dashscope"
    },
    {
      "name": "gpt-3.5",
      "provider": "@openai"
    }
  ]
}
```

<img src="https://mintcdn.com/portkey-docs/wAHXB_jjwLt8bYcN/images/llms/conditional-routing.png?fit=max&auto=format&n=wAHXB_jjwLt8bYcN&q=85&s=435d0d6163ba594815df7b14d4373828" alt="Conditional Routing Diagram" width="1094" height="726" data-path="images/llms/conditional-routing.png" />

<CardGroup cols={2}>
  <Card title="Load Balancing" icon="link" href="/product/ai-gateway/load-balancing">
    Distribute requests across multiple targets
  </Card>

  <Card title="Fallbacks" icon="life-ring" href="/product/ai-gateway/fallbacks">
    Automatically switch to backup targets
  </Card>

  <Card title="Conditional Routing" icon="route" href="/product/ai-gateway/conditional-routing">
    Route requests based on conditions
  </Card>

  <Card title="Caching" icon="database" href="/product/ai-gateway/cache-simple-and-semantic">
    Cache responses to reduce costs
  </Card>
</CardGroup>

### Guardrails

Enforce input/output checks with custom hooks:

```json theme={"system"}
{
  "provider": "@dashscope",
  "before_request_hooks": [{
    "id": "input-guardrail-id-xx"
  }],
  "after_request_hooks": [{
    "id": "output-guardrail-id-xx"
  }]
}
```

<Card title="Learn More About Guardrails" icon="shield-check" href="/product/guardrails">
  Enhance security and reliability with Portkey Guardrails
</Card>

***

## Next Steps

<CardGroup cols={2}>
  <Card title="Gateway Configs" icon="sliders" href="/product/ai-gateway">
    Add fallbacks, load balancing, and more
  </Card>

  <Card title="Observability" icon="chart-line" href="/product/observability">
    Monitor and trace your Dashscope requests
  </Card>

  <Card title="Prompt Library" icon="book" href="/product/prompt-engineering-studio">
    Manage and version your prompts
  </Card>

  <Card title="Metadata" icon="tag" href="/product/observability/metadata">
    Add custom metadata to requests
  </Card>
</CardGroup>

For complete SDK documentation:

<Card title="SDK Reference" icon="code" href="/api-reference/sdk/list">
  Complete Portkey SDK documentation
</Card>

***

For the most up-to-date information on supported features and endpoints, please refer to our [API Reference](/api-reference/inference-api/introduction).
