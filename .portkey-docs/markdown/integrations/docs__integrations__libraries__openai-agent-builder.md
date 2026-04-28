> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# OpenAI Agent Builder (TypeScript)

> Add Portkey to visual agent workflows exported from OpenAI Agent Builder.

OpenAI Agent Builder is a visual canvas for creating multi-step agent workflows. Export production-ready TypeScript code and add Portkey for:

* Complete observability of agent workflows
* Cost tracking and optimization
* Reliability features (fallbacks, retries)
* Access to 1600+ LLMs
* Guardrails for safe agent behavior

## Quick Start

<Steps>
  <Step title="Design in Agent Builder">
    Open [OpenAI Agent Builder](https://platform.openai.com/playground/agent-builder) and create your workflow using the visual canvas.

    <Frame>
      <img src="https://mintcdn.com/portkey-docs/yp2Ps0cNckQLzo91/images/libraries/openai-agent-builder.webp?fit=max&auto=format&n=yp2Ps0cNckQLzo91&q=85&s=dd2ae7746dd8a5d4332c90e6a886ce07" width="3840" height="2160" data-path="images/libraries/openai-agent-builder.webp" />
    </Frame>
  </Step>

  <Step title="Export Code">
    Click **Code** → **Agents SDK** to get the TypeScript implementation.

    <Frame>
      <img src="https://mintcdn.com/portkey-docs/yp2Ps0cNckQLzo91/images/libraries/agent-builder-2.png?fit=max&auto=format&n=yp2Ps0cNckQLzo91&q=85&s=c6a4890ddca39435ac2417686e18faee" width="2449" height="1787" data-path="images/libraries/agent-builder-2.png" />
    </Frame>
  </Step>

  <Step title="Install Packages">
    ```bash theme={"system"}
    npm install @openai/agents openai portkey-ai
    ```
  </Step>

  <Step title="Add Portkey">
    Replace the OpenAI client with Portkey:

    ```typescript theme={"system"}
    import { Agent, run, setDefaultOpenAIClient, setOpenAIAPI } from '@openai/agents';
    import { OpenAI } from 'openai';
    import { PORTKEY_GATEWAY_URL, createHeaders } from 'portkey-ai';

    const client = new OpenAI({
        baseURL: PORTKEY_GATEWAY_URL,
        apiKey: 'YOUR_PORTKEY_API_KEY',
        defaultHeaders: createHeaders({ provider: '@openai-prod' })
    });
    setDefaultOpenAIClient(client);
    setOpenAIAPI('chat_completions');

    // Your Agent Builder workflow code continues as exported...
    ```

    <Note>
      Update the `model` field to use Portkey's format: `@openai-prod/gpt-4o`
    </Note>
  </Step>
</Steps>

## Setup

<Steps>
  <Step title="Add Provider in Model Catalog">
    Go to [Model Catalog → Add Provider](https://app.portkey.ai/model-catalog). Select your provider (OpenAI, Anthropic, etc.), enter API keys, and name it (e.g., `openai-prod`).

    Your provider slug is `@openai-prod`.
  </Step>

  <Step title="Get Portkey API Key">
    Create an API key at [app.portkey.ai/api-keys](https://app.portkey.ai/api-keys).

    **Pro tip:** Attach a [config](https://app.portkey.ai/configs) for fallbacks, caching, and guardrails—applies automatically.
  </Step>
</Steps>

## Production Features

### Observability

All workflow executions are logged:

<Frame>
  <img src="https://mintcdn.com/portkey-docs/_Cb_bj7tVjxcfwsu/images/product/product-11-1.webp?fit=max&auto=format&n=_Cb_bj7tVjxcfwsu&q=85&s=d64ccfdd125ef6ca39bc3e344d7411a4" width="2224" height="1166" data-path="images/product/product-11-1.webp" />
</Frame>

Add trace IDs and metadata for filtering:

```typescript theme={"system"}
defaultHeaders: createHeaders({
    provider: '@openai-prod',
    traceId: 'workflow-session-123',
    metadata: {
        workflow_type: 'research',
        _user: 'user_123',
        environment: 'production'
    }
})
```

<Frame>
  <img src="https://mintcdn.com/portkey-docs/wAHXB_jjwLt8bYcN/images/metadata.png?fit=max&auto=format&n=wAHXB_jjwLt8bYcN&q=85&s=4191ca4b6fa6f583343aeba6c5f5cbb7" alt="Analytics with metadata filters" width="3156" height="1876" data-path="images/metadata.png" />
</Frame>

### Reliability

Enable fallbacks via [Configs](https://app.portkey.ai/configs):

```typescript theme={"system"}
const client = new OpenAI({
    baseURL: PORTKEY_GATEWAY_URL,
    apiKey: 'YOUR_PORTKEY_API_KEY',
    defaultHeaders: createHeaders({
        config: {
            strategy: { mode: 'fallback' },
            targets: [
                { override_params: { model: '@openai-prod/gpt-4o' } },
                { override_params: { model: '@anthropic-prod/claude-sonnet-4' } }
            ]
        }
    })
});
```

<CardGroup cols="2">
  <Card title="Automatic Retries" icon="rotate" href="/product/ai-gateway/automatic-retries">
    Handles temporary failures automatically
  </Card>

  <Card title="Load Balancing" icon="scale-balanced" href="/product/ai-gateway/load-balancing">
    Distribute across multiple keys
  </Card>

  <Card title="Conditional Routing" icon="route" href="/product/ai-gateway/conditional-routing">
    Route based on request attributes
  </Card>

  <Card title="Fallbacks" icon="shield" href="/product/ai-gateway/fallbacks">
    Switch to backup providers automatically
  </Card>
</CardGroup>

### Guardrails

Add input/output validation:

```typescript theme={"system"}
defaultHeaders: createHeaders({
    provider: '@openai-prod',
    config: {
        input_guardrails: ['guardrail-id-xxx'],
        output_guardrails: ['guardrail-id-yyy']
    }
})
```

Guardrails can:

* Detect and redact PII
* Filter harmful content
* Validate response formats
* Apply custom business rules

<Card title="Guardrails Guide" icon="shield-check" href="/product/guardrails">
  PII detection, content filtering, and custom rules
</Card>

### Caching

Reduce costs with response caching:

```typescript theme={"system"}
defaultHeaders: createHeaders({
    provider: '@openai-prod',
    config: { cache: { mode: 'semantic' } }
})
```

### Prompt Templates

Use Portkey's prompt management for versioned prompts:

```typescript theme={"system"}
import { Portkey } from 'portkey-ai';

const portkey = new Portkey({ apiKey: 'YOUR_PORTKEY_API_KEY' });

const promptData = await portkey.prompts.render({
    promptId: 'YOUR_PROMPT_ID',
    variables: { task: 'research' }
});

const agent = new Agent({
    name: 'Assistant',
    instructions: promptData.data.messages[0].content,
    model: '@openai-prod/gpt-4o'
});
```

<Card title="Prompt Engineering Studio" icon="wand-magic-sparkles" href="/product/prompt-library">
  Prompt versioning and collaboration
</Card>

## Switching Providers

Use any of 1600+ models:

```typescript theme={"system"}
// OpenAI
createHeaders({ provider: '@openai-prod' })

// Anthropic
createHeaders({ provider: '@anthropic-prod' })

// Google
createHeaders({ provider: '@google-prod' })
```

<Card title="Supported Providers" icon="server" href="/integrations/llms">
  See all 1600+ supported models
</Card>

## Enterprise Governance

Set up centralized control for your workflows.

<Steps>
  <Step title="Add Provider with Budget">
    Go to [Model Catalog](https://app.portkey.ai/model-catalog) → Add Provider. Set budget limits and rate limits.
  </Step>

  <Step title="Create Config">
    Go to [Configs](https://app.portkey.ai/configs):

    ```json theme={"system"}
    {
      "override_params": { "model": "@openai-prod/gpt-4o" }
    }
    ```
  </Step>

  <Step title="Create Team API Keys">
    Go to [API Keys](https://app.portkey.ai/api-keys). Create keys per team, attach configs.
  </Step>

  <Step title="Distribute to Teams">
    Teams use their Portkey API key:

    ```typescript theme={"system"}
    const client = new OpenAI({
        baseURL: PORTKEY_GATEWAY_URL,
        apiKey: 'TEAM_PORTKEY_API_KEY'  // Config attached to key
    });
    ```
  </Step>
</Steps>

<Card title="Enterprise Features" icon="building" href="/product/enterprise-offering">
  Governance, security, and compliance
</Card>

## FAQ

<AccordionGroup>
  <Accordion title="Can I use Portkey with existing Agent Builder workflows?">
    Yes. Export your workflow, add Portkey client initialization, and your code works unchanged.
  </Accordion>

  <Accordion title="Does Portkey work with all Agent Builder features?">
    Yes. Handoffs, tools, guardrails—all work with Portkey observability and reliability.
  </Accordion>

  <Accordion title="How do I track workflow costs?">
    Add metadata to your requests. Filter by workflow type, user, or environment in the dashboard.
  </Accordion>

  <Accordion title="Can I use my own API keys?">
    Yes. Portkey stores your provider keys securely. Rotate keys without code changes.
  </Accordion>
</AccordionGroup>

## Resources

<CardGroup cols="2">
  <Card title="OpenAI Agents SDK" icon="robot" href="/integrations/agents/openai-agents-ts">
    Full SDK integration guide
  </Card>

  <Card title="Configs" icon="gear" href="/product/ai-gateway/configs">
    Fallbacks, caching, and routing
  </Card>

  <Card title="OpenAI Agents Docs" icon="book" href="https://openai.github.io/openai-agents-js/">
    Official documentation
  </Card>

  <Card title="Book a Demo" icon="calendar" href="https://portkey.sh/openai-agents">
    Get implementation guidance
  </Card>
</CardGroup>
