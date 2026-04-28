> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Mastra Agents

> Use Portkey with Mastra to take your AI Agents to production

## Introduction

Mastra is a TypeScript framework for building AI agents with tools, workflows, memory, and evaluation scoring. Portkey enhances Mastra agents with observability, reliability, and production-readiness features.

Portkey turns your experimental Mastra agents into production-ready systems by providing:

* **Complete observability** of every agent step, tool use, and interaction
* **Built-in reliability** with fallbacks, retries, and load balancing
* **Cost tracking and optimization** to manage your AI spend
* **Access to 1600+ LLMs** through a single integration
* **Guardrails** to keep agent behavior safe and compliant
* **Version-controlled prompts** for consistent agent performance

<Card title="Mastra Official Documentation" icon="arrow-up-right-from-square" href="https://mastra.ai/docs">
  Learn more about Mastra's core concepts and features
</Card>

### Installation & Setup

<Steps>
  <Step title="Install the required packages">
    ```bash theme={"system"}
    npm install @mastra/core
    npm install --save-dev mastra
    ```
  </Step>

  <Step title="Generate API Key">
    Create a Portkey API key from the [Portkey dashboard](https://app.portkey.ai/). You can attach optional budget/rate limits and configurations.
  </Step>

  <Step title="Set Up Provider Integration">
    In Portkey, set up your LLM provider integration:

    1. Go to [Integrations](https://app.portkey.ai/integrations) in Portkey
    2. Connect your LLM provider (OpenAI, Anthropic, etc.)
    3. Note your provider slug (e.g., `openai-dev`, `anthropic-prod`)

    You'll use this slug in your Mastra model configuration.
  </Step>

  <Step title="Configure Mastra Agent with Portkey">
    Configure your Mastra agent's model to use Portkey as the gateway:

    ```typescript theme={"system"}
    import { Agent } from '@mastra/core/agent';

    export const agent = new Agent({
      name: 'Assistant',
      instructions: 'You are a helpful assistant.',
      model: {
        id: 'openai/@YOUR_PROVIDER_SLUG@gpt-4o',  // Format: openai/@provider-slug@model-name
        url: 'https://api.portkey.ai/v1',
        apiKey: 'YOUR_PORTKEY_API_KEY',
        headers: {
          // Optional: Add Portkey configuration
          'x-portkey-trace-id': 'agent-session-123',
          'x-portkey-metadata': JSON.stringify({ 
            agent: 'assistant',
            env: 'production' 
          })
        }
      }
    });
    ```

    <Note>
      **Model ID Format**: Use `openai/@provider-slug@model-name` because Mastra uses OpenAI-compatible interfaces under the hood. The `@provider-slug` should match the slug from your Portkey integration.
    </Note>
  </Step>
</Steps>

## Production Features

### 1. Enhanced Observability

Portkey provides comprehensive observability for your Mastra agents, helping you understand exactly what's happening during each execution.

<Tabs>
  <Tab title="Traces">
    <Frame>
      <img src="https://mintcdn.com/portkey-docs/_Cb_bj7tVjxcfwsu/images/product/product-11-1.webp?fit=max&auto=format&n=_Cb_bj7tVjxcfwsu&q=85&s=d64ccfdd125ef6ca39bc3e344d7411a4" width="2224" height="1166" data-path="images/product/product-11-1.webp" />
    </Frame>

    Traces provide a hierarchical view of your agent's execution, showing the sequence of LLM calls, tool invocations, and state transitions.

    ```typescript theme={"system"}
    // Add tracing to your Mastra agents
    export const agent = new Agent({
      name: 'Research Assistant',
      instructions: 'You are a helpful research assistant.',
      model: {
        id: 'openai/@YOUR_PROVIDER_SLUG@gpt-4o',
        url: 'https://api.portkey.ai/v1',
        apiKey: process.env.PORTKEY_API_KEY,
        headers: {
          'x-portkey-trace-id': 'unique_execution_trace_id', // Add unique trace ID
          'x-portkey-metadata': JSON.stringify({ 
            agent_type: 'research_agent' 
          })
        }
      }
    });
    ```
  </Tab>

  <Tab title="Logs">
    <Frame>
      <img src="https://mintcdn.com/portkey-docs/_Cb_bj7tVjxcfwsu/images/product/product-2.avif?fit=max&auto=format&n=_Cb_bj7tVjxcfwsu&q=85&s=831f1cb98919f6a6d9e2a467db031653" width="800" height="492" data-path="images/product/product-2.avif" />
    </Frame>

    Portkey logs every interaction with LLMs, including:

    * Complete request and response payloads
    * Latency and token usage metrics
    * Cost calculations
    * Tool calls and function executions

    All logs can be filtered by metadata, trace IDs, models, and more, making it easy to debug specific agent runs.
  </Tab>

  <Tab title="Metrics & Dashboards">
    <Frame>
      <img src="https://mintcdn.com/portkey-docs/Buc1Vm2P31GSPm3S/images/product/dashboard.png?fit=max&auto=format&n=Buc1Vm2P31GSPm3S&q=85&s=6c880d5ae17f22884436ad3c7b3347d9" width="600" height="348" data-path="images/product/dashboard.png" />
    </Frame>

    Portkey provides built-in dashboards that help you:

    * Track cost and token usage across all agent runs
    * Analyze performance metrics like latency and success rates
    * Identify bottlenecks in your agent workflows
    * Compare different agent configurations and LLMs

    You can filter and segment all metrics by custom metadata to analyze specific agent types, user groups, or use cases.
  </Tab>

  <Tab title="Metadata Filtering">
    <Frame>
      <img src="https://mintcdn.com/portkey-docs/wAHXB_jjwLt8bYcN/images/metadata.png?fit=max&auto=format&n=wAHXB_jjwLt8bYcN&q=85&s=4191ca4b6fa6f583343aeba6c5f5cbb7" alt="Analytics with metadata filters" width="3156" height="1876" data-path="images/metadata.png" />
    </Frame>

    Add custom metadata to your Mastra agent calls to enable powerful filtering and segmentation:

    ```typescript theme={"system"}
    export const agent = new Agent({
      name: 'Assistant',
      instructions: 'You are a helpful assistant.',
      model: {
        id: 'openai/@YOUR_PROVIDER_SLUG@gpt-4o',
        url: 'https://api.portkey.ai/v1',
        apiKey: process.env.PORTKEY_API_KEY,
        headers: {
          'x-portkey-metadata': JSON.stringify({
            agent_type: 'research_agent',
            environment: 'production',
            _user: 'user_123',  // Special _user field for user analytics
            team: 'engineering'
          })
        }
      }
    });
    ```

    This metadata can be used to filter logs, traces, and metrics on the Portkey dashboard, allowing you to analyze specific agent runs, users, or environments.
  </Tab>
</Tabs>

### 2. Reliability - Keep Your Agents Running Smoothly

When running agents in production, things can go wrong - API rate limits, network issues, or provider outages. Portkey's reliability features ensure your agents keep running smoothly even when problems occur.

It's this simple to enable fallback in your Mastra agents:

```typescript theme={"system"}
import { Agent } from '@mastra/core/agent';

// Create a config with fallbacks
// It's recommended that you create the Config in Portkey App rather than hard-code the config JSON directly
const portkeyConfig = {
  strategy: {
    mode: 'fallback'
  },
  targets: [
    {
      provider: '@YOUR_OPENAI_PROVIDER',
      override_params: { model: 'gpt-4o' }
    },
    {
      provider: '@YOUR_ANTHROPIC_PROVIDER',
      override_params: { model: 'claude-3-opus-20240229' }
    }
  ]
};

export const agent = new Agent({
  name: 'Resilient Agent',
  instructions: 'You are a helpful assistant.',
  model: {
    id: 'openai/@YOUR_OPENAI_PROVIDER@gpt-4o',
    url: 'https://api.portkey.ai/v1',
    apiKey: process.env.PORTKEY_API_KEY,
    headers: {
      'x-portkey-config': JSON.stringify(portkeyConfig)
    }
  }
});
```

This configuration will automatically try Claude if the GPT-4o request fails, ensuring your agent can continue operating.

<CardGroup cols="2">
  <Card title="Automatic Retries" icon="rotate" href="../../product/ai-gateway/automatic-retries">
    Handles temporary failures automatically. If an LLM call fails, Portkey will retry the same request for the specified number of times - perfect for rate limits or network blips.
  </Card>

  <Card title="Request Timeouts" icon="clock" href="../../product/ai-gateway/request-timeouts">
    Prevent your agents from hanging. Set timeouts to ensure you get responses (or can fail gracefully) within your required timeframes.
  </Card>

  <Card title="Conditional Routing" icon="route" href="../../product/ai-gateway/conditional-routing">
    Send different requests to different providers. Route complex reasoning to GPT-4, creative tasks to Claude, and quick responses to Gemini based on your needs.
  </Card>

  <Card title="Fallbacks" icon="shield" href="../../product/ai-gateway/fallbacks">
    Keep running even if your primary provider fails. Automatically switch to backup providers to maintain availability.
  </Card>

  <Card title="Load Balancing" icon="scale-balanced" href="../../product/ai-gateway/load-balancing">
    Spread requests across multiple API keys or providers. Great for high-volume agent operations and staying within rate limits.
  </Card>
</CardGroup>

### 3. Prompting in Mastra Agents

Portkey's Prompt Engineering Studio helps you create, manage, and optimize the prompts used in your Mastra agents. Instead of hardcoding prompts or instructions, use Portkey's prompt rendering API to dynamically fetch and apply your versioned prompts.

<Frame caption="Manage prompts in Portkey's Prompt Library">
  <img src="https://mintcdn.com/portkey-docs/VWP2Y8zxPP5N4jE6/images/product/ai-gateway/ai-20.webp?fit=max&auto=format&n=VWP2Y8zxPP5N4jE6&q=85&s=1460d52a53097dc02c6423ec94bf349c" alt="Prompt Playground Interface" width="2304" height="1302" data-path="images/product/ai-gateway/ai-20.webp" />
</Frame>

<Tabs>
  <Tab title="Prompt Playground">
    Prompt Playground is a place to compare, test and deploy perfect prompts for your AI application. It's where you experiment with different models, test variables, compare outputs, and refine your prompt engineering strategy before deploying to production. It allows you to:

    1. Iteratively develop prompts before using them in your agents
    2. Test prompts with different variables and models
    3. Compare outputs between different prompt versions
    4. Collaborate with team members on prompt development

    This visual environment makes it easier to craft effective prompts for each step in your Mastra agent's workflow.
  </Tab>

  <Tab title="Using Prompt Templates">
    The Prompt Render API retrieves your prompt templates with all parameters configured:

    ```typescript theme={"system"}
    import { Portkey } from 'portkey-ai';
    import { Agent } from '@mastra/core/agent';

    // Initialize Portkey client
    const portkey = new Portkey({
      apiKey: process.env.PORTKEY_API_KEY
    });

    // Retrieve prompt using the render API
    const promptData = await portkey.prompts.render({
      promptID: 'YOUR_PROMPT_ID',
      variables: {
        user_input: 'Tell me about artificial intelligence'
      }
    });

    // Use the rendered prompt in your Mastra agent
    export const agent = new Agent({
      name: 'Assistant',
      instructions: promptData.data.messages[0].content, // Use the rendered prompt as instructions
      model: {
        id: 'openai/@YOUR_PROVIDER_SLUG@gpt-4o',
        url: 'https://api.portkey.ai/v1',
        apiKey: process.env.PORTKEY_API_KEY
      }
    });

    const result = await agent.generate('Tell me about artificial intelligence');
    console.log(result.text);
    ```
  </Tab>

  <Tab title="Prompt Versioning">
    You can:

    * Create multiple versions of the same prompt
    * Compare performance between versions
    * Roll back to previous versions if needed
    * Specify which version to use in your code:

    ```typescript theme={"system"}
    // Use a specific prompt version
    const promptData = await portkey.prompts.render({
      promptID: 'YOUR_PROMPT_ID@version_number',
      variables: {
        user_input: 'Tell me about quantum computing'
      }
    });
    ```
  </Tab>

  <Tab title="Mustache Templating for variables">
    Portkey prompts use Mustache-style templating for easy variable substitution:

    ```
    You are an AI assistant helping with {{task_type}}.

    User question: {{user_input}}

    Please respond in a {{tone}} tone and include {{required_elements}}.
    ```

    When rendering, simply pass the variables:

    ```typescript theme={"system"}
    const promptData = await portkey.prompts.render({
      promptID: 'YOUR_PROMPT_ID',
      variables: {
        task_type: 'research',
        user_input: 'Tell me about quantum computing',
        tone: 'professional',
        required_elements: 'recent academic references'
      }
    });
    ```
  </Tab>
</Tabs>

<Card title="Prompt Engineering Studio" icon="wand-magic-sparkles" href="/product/prompt-library">
  Learn more about Portkey's prompt management features
</Card>

### 4. Guardrails for Safe Agents

Guardrails ensure your Mastra agents operate safely and respond appropriately in all situations.

**Why Use Guardrails?**

Mastra agents can experience various failure modes:

* Generating harmful or inappropriate content
* Leaking sensitive information like PII
* Hallucinating incorrect information
* Generating outputs in incorrect formats

Portkey's guardrails protect against these issues by validating both inputs and outputs.

**Implementing Guardrails**

```typescript theme={"system"}
import { Agent } from '@mastra/core/agent';

// Create a config with input and output guardrails
// It's recommended you create Config in Portkey App and pass the config ID in the headers
const guardrailConfig = {
  provider: '@YOUR_PROVIDER',
  input_guardrails: ['guardrails-id-xxx', 'guardrails-id-yyy'],
  output_guardrails: ['guardrails-id-xxx']
};

export const agent = new Agent({
  name: 'Safe Agent',
  instructions: 'You are a helpful assistant that provides safe responses.',
  model: {
    id: 'openai/@YOUR_PROVIDER_SLUG@gpt-4o',
    url: 'https://api.portkey.ai/v1',
    apiKey: process.env.PORTKEY_API_KEY,
    headers: {
      'x-portkey-config': JSON.stringify(guardrailConfig)
    }
  }
});
```

Portkey's guardrails can:

* Detect and redact PII in both inputs and outputs
* Filter harmful or inappropriate content
* Validate response formats against schemas
* Check for hallucinations against ground truth
* Apply custom business logic and rules

<Card title="Learn More About Guardrails" icon="shield-check" href="/product/guardrails">
  Explore Portkey's guardrail features to enhance agent safety
</Card>

### 5. User Tracking with Metadata

Track individual users through your Mastra agents using Portkey's metadata system.

**What is Metadata in Portkey?**

Metadata allows you to associate custom data with each request, enabling filtering, segmentation, and analytics. The special `_user` field is specifically designed for user tracking.

```typescript theme={"system"}
import { Agent } from '@mastra/core/agent';

export const agent = new Agent({
  name: 'Personalized Agent',
  instructions: 'You are a personalized assistant.',
  model: {
    id: 'openai/@YOUR_PROVIDER_SLUG@gpt-4o',
    url: 'https://api.portkey.ai/v1',
    apiKey: process.env.PORTKEY_API_KEY,
    headers: {
      'x-portkey-metadata': JSON.stringify({
        _user: 'user_123',  // Special _user field for user analytics
        user_name: 'John Doe',
        user_tier: 'premium',
        user_company: 'Acme Corp'
      })
    }
  }
});
```

**Filter Analytics by User**

With metadata in place, you can filter analytics by user and analyze performance metrics on a per-user basis:

<Frame caption="Filter analytics by user">
  <img src="https://mintcdn.com/portkey-docs/wAHXB_jjwLt8bYcN/images/metadata-filters.png?fit=max&auto=format&n=wAHXB_jjwLt8bYcN&q=85&s=69ecfc9ee8e12723441941a28c60f682" width="1158" height="732" data-path="images/metadata-filters.png" />
</Frame>

This enables:

* Per-user cost tracking and budgeting
* Personalized user analytics
* Team or organization-level metrics
* Environment-specific monitoring (staging vs. production)

<Card title="Learn More About Metadata" icon="tags" href="/product/observability/metadata">
  Explore how to use custom metadata to enhance your analytics
</Card>

### 6. Caching for Efficient Agents

Implement caching to make your Mastra agents more efficient and cost-effective:

<Tabs>
  <Tab title="Simple Caching">
    ```typescript theme={"system"}
    import { Agent } from '@mastra/core/agent';

    const cacheConfig = {
      provider: '@YOUR_PROVIDER',
      cache: {
        mode: 'simple'
      }
    };

    export const agent = new Agent({
      name: 'Cached Agent',
      instructions: 'You are a helpful assistant.',
      model: {
        id: 'openai/@YOUR_PROVIDER_SLUG@gpt-4o',
        url: 'https://api.portkey.ai/v1',
        apiKey: process.env.PORTKEY_API_KEY,
        headers: {
          'x-portkey-config': JSON.stringify(cacheConfig)
        }
      }
    });
    ```

    Simple caching performs exact matches on input prompts, caching identical requests to avoid redundant model executions.
  </Tab>

  <Tab title="Semantic Caching">
    ```typescript theme={"system"}
    import { Agent } from '@mastra/core/agent';

    const semanticCacheConfig = {
      provider: '@YOUR_PROVIDER',
      cache: {
        mode: 'semantic'
      }
    };

    export const agent = new Agent({
      name: 'Semantically Cached Agent',
      instructions: 'You are a helpful assistant.',
      model: {
        id: 'openai/@YOUR_PROVIDER_SLUG@gpt-4o',
        url: 'https://api.portkey.ai/v1',
        apiKey: process.env.PORTKEY_API_KEY,
        headers: {
          'x-portkey-config': JSON.stringify(semanticCacheConfig)
        }
      }
    });
    ```

    Semantic caching considers the contextual similarity between input requests, caching responses for semantically similar inputs.
  </Tab>
</Tabs>

### 7. Model Interoperability

With Portkey, you can easily switch between different LLMs in your Mastra agents without changing your core agent logic.

```typescript theme={"system"}
import { Agent } from '@mastra/core/agent';

// Using OpenAI
const openaiAgent = new Agent({
  name: 'OpenAI Agent',
  instructions: 'You are a helpful assistant.',
  model: {
    id: 'openai/@YOUR_OPENAI_PROVIDER@gpt-4o',
    url: 'https://api.portkey.ai/v1',
    apiKey: process.env.PORTKEY_API_KEY
  }
});

// Using Anthropic
const anthropicAgent = new Agent({
  name: 'Anthropic Agent',
  instructions: 'You are a helpful assistant.',
  model: {
    id: 'openai/@YOUR_ANTHROPIC_PROVIDER@claude-3-opus-20240229',
    url: 'https://api.portkey.ai/v1',
    apiKey: process.env.PORTKEY_API_KEY
  }
});

// Using Google Gemini
const geminiAgent = new Agent({
  name: 'Gemini Agent',
  instructions: 'You are a helpful assistant.',
  model: {
    id: 'openai/@YOUR_GOOGLE_PROVIDER@gemini-2.0-flash-exp',
    url: 'https://api.portkey.ai/v1',
    apiKey: process.env.PORTKEY_API_KEY
  }
});
```

Portkey provides access to over 200 LLMs through a unified interface, including:

* OpenAI (GPT-4o, GPT-4 Turbo, etc.)
* Anthropic (Claude 3.5 Sonnet, Claude 3 Opus, etc.)
* Mistral AI (Mistral Large, Mistral Medium, etc.)
* Google Vertex AI (Gemini 1.5 Pro, etc.)
* Cohere (Command, Command-R, etc.)
* AWS Bedrock (Claude, Titan, etc.)
* Local/Private Models

<Card title="Supported Providers" icon="server" href="/integrations/llms">
  See the full list of LLM providers supported by Portkey
</Card>

## Set Up Enterprise Governance for Mastra Agents

**Why Enterprise Governance?**

If you are using Mastra agents inside your organization, you need to consider several governance aspects:

* **Cost Management**: Controlling and tracking AI spending across teams
* **Access Control**: Managing which teams can use specific models
* **Usage Analytics**: Understanding how AI is being used across the organization
* **Security & Compliance**: Maintaining enterprise security standards
* **Reliability**: Ensuring consistent service across all users

Portkey adds a comprehensive governance layer to address these enterprise needs. Let's implement these controls step by step.

**Enterprise Implementation Guide**

Portkey allows you to use 1600+ LLMs with your Mastra agents setup, with minimal configuration required. Let's set up the core components in Portkey that you'll need for integration.

<Steps>
  <Step title="Create Integration">
    To create a new LLM integration:

    Go to [Integrations](https://app.portkey.ai/integrations) in the Portkey App. Set budget / rate limits, model access if required and save the integration.

    This creates a "Portkey Provider" that you can then use in any of your Portkey requests without having to send auth details for that LLM provider again.

    <Frame>
      <img src="https://mintcdn.com/portkey-docs/Buc1Vm2P31GSPm3S/images/product/model-catalog/integrations-page.png?fit=max&auto=format&n=Buc1Vm2P31GSPm3S&q=85&s=df13f2b4eeb68e349b133a53344a543a" width="500" data-path="images/product/model-catalog/integrations-page.png" />
    </Frame>
  </Step>

  <Step title="Create Config">
    Configs in Portkey define how your requests are routed, with features like advanced routing, fallbacks, and retries.

    To create your config:

    1. Go to [Configs](https://app.portkey.ai/configs) in Portkey dashboard
    2. Create new config with:

    ```json theme={"system"}
    {
      "provider": "@YOUR_PROVIDER_FROM_STEP1",
      "override_params": {
        "model": "gpt-4o" // Your preferred model name
      }
    }
    ```

    3. Save and note the Config ID for the next step

    <Frame>
      <img src="https://mintcdn.com/portkey-docs/T0lFtdapIPX8YtCI/images/integrations/config.png?fit=max&auto=format&n=T0lFtdapIPX8YtCI&q=85&s=1a7edd4ef3643cae7fec8595b6d315d7" width="500" data-path="images/integrations/config.png" />
    </Frame>
  </Step>

  <Step title="Configure Portkey API Key">
    Now create a Portkey API key and attach the config you created in Step 2:

    1. Go to [API Keys](https://app.portkey.ai/api-keys) in Portkey and Create new API key
    2. Select your config from Step 2
    3. Generate and save your API key

    <Frame>
      <img src="https://mintcdn.com/portkey-docs/T0lFtdapIPX8YtCI/images/integrations/api-key.png?fit=max&auto=format&n=T0lFtdapIPX8YtCI&q=85&s=9ac7b990452c59e9bd167761abbd30da" width="500" data-path="images/integrations/api-key.png" />
    </Frame>
  </Step>

  <Step title="Connect to Mastra">
    After setting up your Portkey API key with the attached config, connect it to your Mastra agents:

    ```typescript theme={"system"}
    import { Agent } from '@mastra/core/agent';

    export const agent = new Agent({
      name: 'Enterprise Agent',
      instructions: 'You are a helpful assistant.',
      model: {
        id: 'openai/@YOUR_PROVIDER_SLUG@gpt-4o',
        url: 'https://api.portkey.ai/v1',
        apiKey: 'YOUR_PORTKEY_API_KEY'  // The API key with attached config from step 3
      }
    });
    ```
  </Step>
</Steps>

<AccordionGroup>
  <Accordion title="Step 1: Implement Budget Controls & Rate Limits">
    ### Step 1: Implement Budget Controls & Rate Limits

    Integrations enable granular control over LLM access at the team/department level. This helps you:

    * Set up [budget limits](/product/model-catalog/budget-limits)
    * Prevent unexpected usage spikes using Rate limits
    * Track departmental spending

    #### Setting Up Department-Specific Controls:

    1. Navigate to [Integrations](https://app.portkey.ai/integrations) in Portkey dashboard and create a new Integration
    2. Provision this Integration for each department with their budget limits and rate limits
    3. Configure model access if required

    <Frame>
      <img src="https://mintcdn.com/portkey-docs/Buc1Vm2P31GSPm3S/images/product/model-catalog/create-provider-page.png?fit=max&auto=format&n=Buc1Vm2P31GSPm3S&q=85&s=621794dd554eb06ec3d5d567b60ea1a9" width="500" data-path="images/product/model-catalog/create-provider-page.png" />
    </Frame>
  </Accordion>

  <Accordion title="Step 2: Define Model Access Rules">
    ### Step 2: Define Model Access Rules

    As your AI usage scales, controlling which teams can access specific models becomes crucial. Portkey Configs provide this control layer with features like:

    #### Access Control Features:

    * **Model Restrictions**: Limit access to specific models
    * **Data Protection**: Implement guardrails for sensitive data
    * **Reliability Controls**: Add fallbacks and retry logic

    #### Example Configuration:

    Here's a basic configuration to route requests to OpenAI, specifically using GPT-4o:

    ```json theme={"system"}
    {
      "strategy": {
        "mode": "single"
      },
      "targets": [
        {
          "provider": "@YOUR_OPENAI_PROVIDER",
          "override_params": {
            "model": "gpt-4o"
          }
        }
      ]
    }
    ```

    Create your config on the [Configs page](https://app.portkey.ai/configs) in your Portkey dashboard.

    <Note>
      Configs can be updated anytime to adjust controls without affecting running applications.
    </Note>
  </Accordion>

  <Accordion title="Step 3: Implement Access Controls">
    ### Step 3: Implement Access Controls

    Create User-specific API keys that automatically:

    * Track usage per user/team with the help of metadata
    * Apply appropriate configs to route requests
    * Collect relevant metadata to filter logs
    * Enforce access permissions

    Create API keys through:

    * [Portkey App](https://app.portkey.ai/)
    * [API Key Management API](/api-reference/admin-api/control-plane/api-keys/create-api-key)

    Example using Node.js SDK:

    ```typescript theme={"system"}
    import { Portkey } from 'portkey-ai';

    const portkey = new Portkey({
      apiKey: 'YOUR_ADMIN_API_KEY'
    });

    const apiKey = await portkey.apiKeys.create({
      name: 'engineering-team',
      type: 'organisation',
      workspaceId: 'YOUR_WORKSPACE_ID',
      defaults: {
        configId: 'your-config-id',
        metadata: {
          environment: 'production',
          department: 'engineering'
        }
      },
      scopes: ['logs.view', 'configs.read']
    });
    ```

    For detailed key management instructions, see our [API Keys documentation](/api-reference/admin-api/control-plane/api-keys/create-api-key).
  </Accordion>

  <Accordion title="Step 4: Deploy & Monitor">
    ### Step 4: Deploy & Monitor

    After distributing API keys to your team members, your enterprise-ready Mastra setup is ready to go. Each team member can now use their designated API keys with appropriate access levels and budget controls.

    Apply your governance setup using the integration steps from earlier sections. Monitor usage in Portkey dashboard:

    * Cost tracking by department
    * Model usage patterns
    * Request volumes
    * Error rates
  </Accordion>
</AccordionGroup>

<Check>
  ### Enterprise Features Now Available

  **Mastra agents now has:**

  * Departmental budget controls
  * Model access governance
  * Usage tracking & attribution
  * Security guardrails
  * Reliability features
</Check>

## Frequently Asked Questions

<AccordionGroup>
  <Accordion title="How does Portkey enhance Mastra agents?">
    Portkey adds production-readiness to Mastra agents through comprehensive observability (traces, logs, metrics), reliability features (fallbacks, retries, caching), and access to 1600+ LLMs through a unified interface. This makes it easier to debug, optimize, and scale your agent applications.
  </Accordion>

  <Accordion title="Can I use Portkey with existing Mastra agents?">
    Yes! Portkey integrates seamlessly with existing Mastra agents. You only need to update your agent's model configuration to point to Portkey. The rest of your agent code remains unchanged.
  </Accordion>

  <Accordion title="Does Portkey work with all Mastra features?">
    Portkey supports all Mastra features, including tools, workflows, memory, and scoring. It adds observability and reliability without limiting any of the framework's functionality.
  </Accordion>

  <Accordion title="Why does the model ID use 'openai/' prefix?">
    Mastra uses OpenAI-compatible interfaces under the hood, so the model ID format is `openai/@provider-slug@model-name`. This allows Mastra to work with any LLM provider through Portkey, not just OpenAI. The `@provider-slug` corresponds to your Portkey integration slug.
  </Accordion>

  <Accordion title="How do I filter logs and traces for specific agent runs?">
    Portkey allows you to add custom metadata and trace IDs to your agent runs through the model headers. Add fields like `agent_name`, `agent_type`, or `session_id` to easily find and analyze specific agent executions.
  </Accordion>

  <Accordion title="Can I use different LLM providers for different agents?">
    Yes! Simply change the provider slug in the model ID. For example, use `openai/@openai-provider@gpt-4o` for one agent and `openai/@anthropic-provider@claude-3-opus-20240229` for another. All agents route through Portkey.
  </Accordion>
</AccordionGroup>

## Resources

<CardGroup cols="3">
  <Card title="Mastra Docs" icon="book" href="https://mastra.ai/docs">
    <p>Official Mastra documentation</p>
  </Card>

  <Card title="Mastra Examples" icon="code" href="https://github.com/mastra-ai/mastra/tree/main/examples">
    <p>Example implementations for various use cases</p>
  </Card>

  <Card title="Book a Demo" icon="calendar" href="https://portkey.sh/demo">
    <p>Get personalized guidance on implementing this integration</p>
  </Card>
</CardGroup>
