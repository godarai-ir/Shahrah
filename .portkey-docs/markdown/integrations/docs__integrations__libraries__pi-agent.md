> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Pi Agent

> Route Pi coding agent through Portkey for observability, cost tracking, and access to 200+ LLMs

[Pi](https://github.com/badlogic/pi-mono/tree/main/packages/coding-agent) is a minimal, extensible terminal coding agent. It supports custom providers via `~/.pi/agent/models.json`, making it straightforward to route through Portkey's AI Gateway.

With Portkey, you get full observability, cost tracking, budget controls, and access to any model from 200+ providers — all without changing how you use Pi.

## Quick start

<Steps>
  <Step title="Add a provider in Portkey">
    Go to [AI Providers](https://app.portkey.ai/ai-providers) → **Add Provider** → select your LLM provider (e.g., OpenAI, Anthropic, Google).

    Enter your provider API key and create a slug like `openai-prod` or `anthropic-prod`.

    <Frame>
      <img src="https://mintcdn.com/portkey-docs/Buc1Vm2P31GSPm3S/images/product/model-catalog/create-provider-page.png?fit=max&auto=format&n=Buc1Vm2P31GSPm3S&q=85&s=621794dd554eb06ec3d5d567b60ea1a9" width="500" data-path="images/product/model-catalog/create-provider-page.png" />
    </Frame>
  </Step>

  <Step title="Get your Portkey API key">
    Go to [API Keys](https://app.portkey.ai/api-keys) → **Generate a new key**.
  </Step>

  <Step title="Configure Pi">
    Edit `~/.pi/agent/models.json` (create it if it doesn't exist):

    ```json theme={"system"}
    {
      "providers": {
        "portkey": {
          "api": "openai-completions",
          "baseUrl": "https://api.portkey.ai/v1",
          "apiKey": "YOUR_PORTKEY_API_KEY",
          "models": [
            {
              "id": "@openai-prod/gpt-4o",
              "name": "GPT-4o via Portkey"
            }
          ]
        }
      }
    }
    ```

    Replace:

    * `YOUR_PORTKEY_API_KEY` with your Portkey API key
    * `@openai-prod` with your provider slug
    * `gpt-4o` with the model you want to use

    The model `id` format is `@<provider-slug>/<model-name>` — this maps directly to a provider integration in your Portkey workspace.
  </Step>

  <Step title="Start Pi with Portkey">
    ```bash theme={"system"}
    pi --provider portkey --model gpt-4o
    ```

    Or use the model selector inside Pi with <kbd>Ctrl</kbd>+<kbd>L</kbd> and pick your Portkey model.
  </Step>
</Steps>

All Pi requests now route through Portkey. Monitor usage, costs, and traces in the [Portkey Dashboard](https://app.portkey.ai/logs).

## Add multiple models

Add as many models as you need under the `models` array. Pi lets you cycle through them with <kbd>Ctrl</kbd>+<kbd>P</kbd>:

```json theme={"system"}
{
  "providers": {
    "portkey": {
      "api": "openai-completions",
      "baseUrl": "https://api.portkey.ai/v1",
      "apiKey": "YOUR_PORTKEY_API_KEY",
      "models": [
        {
          "id": "@openai-prod/gpt-4o",
          "name": "GPT-4o"
        },
        {
          "id": "@anthropic-prod/claude-sonnet-4-20250514",
          "name": "Claude Sonnet 4"
        },
        {
          "id": "@gemini-prod/gemini-2.5-pro",
          "name": "Gemini 2.5 Pro"
        }
      ]
    }
  }
}
```

Each model `id` maps to a different provider integration in Portkey. Switch providers without managing separate API keys in Pi.

## Trace requests

Add trace IDs to group and debug requests in the Portkey Dashboard. Create a [Portkey Config](/product/ai-gateway/configs) with metadata and attach it to a scoped API key:

<Steps>
  <Step title="Create a Config with metadata">
    Go to [Configs](https://app.portkey.ai/configs) → **Create Config**:

    ```json theme={"system"}
    {
      "provider": "@openai-prod",
      "metadata": {
        "environment": "development",
        "agent": "pi"
      }
    }
    ```

    Save and copy the Config ID.
  </Step>

  <Step title="Create a scoped API key with the Config">
    Go to [API Keys](https://app.portkey.ai/api-keys) → **Create Key** → attach your Config ID under **Advanced Settings**.
  </Step>

  <Step title="Use the scoped key in Pi">
    ```json theme={"system"}
    {
      "providers": {
        "portkey": {
          "api": "openai-completions",
          "baseUrl": "https://api.portkey.ai/v1",
          "apiKey": "YOUR_SCOPED_PORTKEY_KEY",
          "models": [
            {
              "id": "gpt-4o",
              "name": "GPT-4o"
            }
          ]
        }
      }
    }
    ```
  </Step>
</Steps>

Use trace metadata to:

* Group all requests from a coding session
* Debug issues by filtering logs by environment or agent
* Track usage per project or task

## Advanced configuration

All advanced routing is configured through [Portkey Configs](/product/ai-gateway/configs), then attached to a scoped API key used as `apiKey` in `models.json`. Pi passes the key as a Bearer token — Portkey applies the routing logic server-side.

### Fallbacks

Route to backup providers when the primary fails. Create a config with fallback targets:

```json theme={"system"}
{
  "strategy": { "mode": "fallback" },
  "targets": [
    { "provider": "@openai-prod" },
    { "provider": "@anthropic-prod" },
    { "provider": "@bedrock-prod" }
  ]
}
```

### Load balancing

Distribute requests across multiple API keys or providers:

```json theme={"system"}
{
  "strategy": { "mode": "loadbalance" },
  "targets": [
    { "provider": "@openai-key-1", "weight": 0.5 },
    { "provider": "@openai-key-2", "weight": 0.5 }
  ]
}
```

### Caching

Reduce costs and latency for repeated queries:

```json theme={"system"}
{
  "provider": "@openai-prod",
  "cache": { "mode": "simple" }
}
```

### Retries

Automatically retry failed requests:

```json theme={"system"}
{
  "provider": "@openai-prod",
  "retry": { "attempts": 3, "on_status_codes": [429, 500, 502, 503] }
}
```

## Budget limits

Set spending controls at the provider level to prevent runaway costs from long coding sessions:

1. Go to [AI Providers](https://app.portkey.ai/ai-providers) → select your provider
2. Click **Budget & Limits**
3. Configure:
   * **Cost limit**: Maximum spend (e.g., \$100/month)
   * **Token limit**: Maximum tokens (e.g., 5M tokens/week)
   * **Rate limit**: Requests per minute

<Frame>
  <img src="https://mintcdn.com/portkey-docs/Buc1Vm2P31GSPm3S/images/product/model-catalog/budget-and-limits-page-model-catalog.png?fit=max&auto=format&n=Buc1Vm2P31GSPm3S&q=85&s=a7debfad7682a196e68a8e30d6aa5a15" width="500" data-path="images/product/model-catalog/budget-and-limits-page-model-catalog.png" />
</Frame>

Budget limits prevent runaway costs from agentic coding sessions.

## Complete example

Full `~/.pi/agent/models.json` with multiple providers and a scoped key that applies fallback + retry logic:

```json theme={"system"}
{
  "providers": {
    "portkey": {
      "api": "openai-completions",
      "baseUrl": "https://api.portkey.ai/v1",
      "apiKey": "YOUR_SCOPED_PORTKEY_KEY",
      "models": [
        {
          "id": "gpt-4o",
          "name": "GPT-4o (with fallback)"
        },
        {
          "id": "@anthropic-prod/claude-sonnet-4-20250514",
          "name": "Claude Sonnet 4"
        },
        {
          "id": "@gemini-prod/gemini-2.5-pro",
          "name": "Gemini 2.5 Pro"
        }
      ]
    }
  }
}
```

Corresponding Portkey Config attached to the scoped key:

```json theme={"system"}
{
  "strategy": { "mode": "fallback" },
  "targets": [
    { "provider": "@openai-prod" },
    { "provider": "@anthropic-prod" }
  ],
  "cache": { "mode": "simple" },
  "retry": { "attempts": 3, "on_status_codes": [429, 500, 502, 503] },
  "metadata": {
    "agent": "pi",
    "environment": "production"
  }
}
```

## Troubleshooting

### Requests not appearing in Portkey logs

**Cause:** Pi may not be using the configured provider.

**Fix:**

1. Verify `~/.pi/agent/models.json` is valid JSON — run `cat ~/.pi/agent/models.json` to check
2. Start Pi with `--provider portkey` explicitly
3. Confirm `baseUrl` is `https://api.portkey.ai/v1` (with `/v1`)
4. Check that `apiKey` matches your Portkey API key

### Model not found error

**Cause:** The model `id` doesn't match a valid provider slug and model in your Portkey workspace.

**Fix:** Go to [AI Providers](https://app.portkey.ai/ai-providers), confirm the provider slug, and verify the model name is supported by that provider.

### Pi version issues

**Cause:** Older versions of Pi may not support the `models.json` custom provider format.

**Fix:** Update to the latest version:

```bash theme={"system"}
npm update -g @mariozechner/pi-coding-agent
```

# 3. Set Up Enterprise Governance

**Why Enterprise Governance?**

* **Cost Management**: Controlling and tracking AI spending across teams
* **Access Control**: Managing team access and workspaces
* **Usage Analytics**: Understanding how AI is being used across the organization
* **Security & Compliance**: Maintaining enterprise security standards
* **Reliability**: Ensuring consistent service across all users
* **Model Management**: Managing what models are being used in your setup

Portkey adds a comprehensive governance layer to address these enterprise needs.

**Enterprise Implementation Guide**

<AccordionGroup>
  <Accordion title="Step 1: Implement Budget Controls & Rate Limits">
    ### Step 1: Implement Budget Controls & Rate Limits

    Model Catalog enables you to have granular control over LLM access at the team/department level. This helps you:

    * Set up [budget limits](/product/model-catalog/budget-limits)
    * Prevent unexpected usage spikes using Rate limits
    * Track departmental spending

    #### Setting Up Department-Specific Controls:

    1. Navigate to [Model Catalog](https://app.portkey.ai/model-catalog) in Portkey dashboard
    2. Create new Provider for each engineering team with budget limits and rate limits
    3. Configure department-specific limits

    <Frame>
      <img src="https://mintcdn.com/portkey-docs/Buc1Vm2P31GSPm3S/images/product/model-catalog/create-provider-page.png?fit=max&auto=format&n=Buc1Vm2P31GSPm3S&q=85&s=621794dd554eb06ec3d5d567b60ea1a9" width="500" data-path="images/product/model-catalog/create-provider-page.png" />
    </Frame>
  </Accordion>

  <Accordion title="Step 2: Define Model Access Rules">
    ### Step 2: Define Model Access Rules

    As your AI usage scales, controlling which teams can access specific models becomes crucial. You can simply manage AI models in your org by provisioning model at the top integration level.

    <Frame>
      <img src="https://mintcdn.com/portkey-docs/Buc1Vm2P31GSPm3S/images/product/model-catalog/model-provisioning-page.png?fit=max&auto=format&n=Buc1Vm2P31GSPm3S&q=85&s=dd617273c3e1cc352c5f1276631dc07c" width="500" data-path="images/product/model-catalog/model-provisioning-page.png" />
    </Frame>
  </Accordion>

  <Accordion title="Step 4: Set Routing Configuration">
    Portkey allows you to control your routing logic very simply with it's Configs feature. Portkey Configs provide this control layer with things like:

    * **Data Protection**: Implement guardrails for sensitive code and data
    * **Reliability Controls**: Add fallbacks, load-balance, retry and smart conditional routing logic
    * **Caching**: Implement Simple and Semantic Caching. and more....

    #### Example Configuration:

    Here's a basic configuration to load-balance requests to OpenAI and Anthropic:

    ```json theme={"system"}
    {
    	"strategy": {
    		"mode": "load-balance"
    	},
    	"targets": [
    		{
    			"override_params": {
    				"model": "@YOUR_OPENAI_PROVIDER_SLUG/gpt-model"
    			}
    		},
    		{
    			"override_params": {
    				"model": "@YOUR_ANTHROPIC_PROVIDER/claude-sonnet-model"
    			}
    		}
    	]
    }
    ```

    Create your config on the [Configs page](https://app.portkey.ai/configs) in your Portkey dashboard. You'll need the config ID for connecting.

    <Note>
      Configs can be updated anytime to adjust controls without affecting running applications.
    </Note>
  </Accordion>

  <Accordion title="Step 4: Implement Access Controls">
    ### Step 3: Implement Access Controls

    Create User-specific API keys that automatically:

    * Track usage per developer/team with the help of metadata
    * Apply appropriate configs to route requests
    * Collect relevant metadata to filter logs
    * Enforce access permissions

    Create API keys through:

    * [Portkey App](https://app.portkey.ai/)
    * [API Key Management API](/api-reference/admin-api/control-plane/api-keys/create-api-key)

    Example using Python SDK:

    ```python theme={"system"}
    from portkey_ai import Portkey

    portkey = Portkey(api_key="YOUR_ADMIN_API_KEY")

    api_key = portkey.api_keys.create(
        name="frontend-engineering",
        type="organisation",
        workspace_id="YOUR_WORKSPACE_ID",
        defaults={
            "config_id": "your-config-id",
            "metadata": {
                "environment": "development",
                "department": "engineering",
                "team": "frontend"
            }
        },
        scopes=["logs.view", "configs.read"]
    )
    ```

    For detailed key management instructions, see our [API Keys documentation](/api-reference/admin-api/control-plane/api-keys/create-api-key).
  </Accordion>

  <Accordion title="Step 5: Deploy & Monitor">
    ### Step 4: Deploy & Monitor

    After distributing API keys to your engineering teams, your enterprise-ready setup is ready to go. Each developer can now use their designated API keys with appropriate access levels and budget controls.
    Apply your governance setup using the integration steps from earlier sections
    Monitor usage in Portkey dashboard:

    * Cost tracking by engineering team
    * Model usage patterns for AI agent tasks
    * Request volumes
    * Error rates and debugging logs
  </Accordion>
</AccordionGroup>

<Check>
  ### Enterprise Features Now Available

  **You now have:**

  * Departmental budget controls
  * Model access governance
  * Usage tracking & attribution
  * Security guardrails
  * Reliability features
</Check>

# Portkey Features

Now that you have an enterprise-grade setup, let's explore the comprehensive features Portkey provides to ensure secure, efficient, and cost-effective AI operations.

### 1. Comprehensive Metrics

Using Portkey you can track 40+ key metrics including cost, token usage, response time, and performance across all your LLM providers in real time. You can also filter these metrics based on custom metadata that you can set in your configs. Learn more about metadata here.

<Frame>
  <img src="https://mintcdn.com/portkey-docs/T0lFtdapIPX8YtCI/images/integrations/observability.png?fit=max&auto=format&n=T0lFtdapIPX8YtCI&q=85&s=2ae78f4fa0c682ce65125ce6bc0d0d55" width="600" data-path="images/integrations/observability.png" />
</Frame>

### 2. Advanced Logs

Portkey's logging dashboard provides detailed logs for every request made to your LLMs. These logs include:

* Complete request and response tracking
* Metadata tags for filtering
* Cost attribution and much more...

<Frame>
  <img src="https://mintcdn.com/portkey-docs/wAHXB_jjwLt8bYcN/images/llms/openai/logs.png?fit=max&auto=format&n=wAHXB_jjwLt8bYcN&q=85&s=bc96d99ebbd97ce31224877650cbee8b" width="3040" height="1764" data-path="images/llms/openai/logs.png" />
</Frame>

### 3. Unified Access to 1600+ LLMs

You can easily switch between 1600+ LLMs. Call various LLMs such as Anthropic, Gemini, Mistral, Azure OpenAI, Google Vertex AI, AWS Bedrock, and many more by simply changing the `provider` slug in your default `config` object.

### 4. Advanced Metadata Tracking

Using Portkey, you can add custom metadata to your LLM requests for detailed tracking and analytics. Use metadata tags to filter logs, track usage, and attribute costs across departments and teams.

<Card title="Custom Metata" icon="coins" href="/product/observability/metadata" />

### 5. Enterprise Access Management

<CardGroup cols={2}>
  <Card title="Budget Controls" icon="coins" href="/product/model-catalog/budget-limits">
    Set and manage spending limits across teams and departments. Control costs with granular budget limits and usage tracking.
  </Card>

  <Card title="Single Sign-On (SSO)" icon="key" href="/product/enterprise-offering/org-management/sso">
    Enterprise-grade SSO integration with support for SAML 2.0, Okta, Azure AD, and custom providers for secure authentication.
  </Card>

  <Card title="Organization Management" icon="building" href="/product/enterprise-offering/org-management">
    Hierarchical organization structure with workspaces, teams, and role-based access control for enterprise-scale deployments.
  </Card>

  <Card title="Access Rules & Audit Logs" icon="shield-check" href="/product/enterprise-offering/access-control-management#audit-logs">
    Comprehensive access control rules and detailed audit logging for security compliance and usage tracking.
  </Card>
</CardGroup>

### 6. Reliability Features

<CardGroup cols={3}>
  <Card title="Fallbacks" icon="life-ring" href="/product/ai-gateway/fallbacks">
    Automatically switch to backup targets if the primary target fails.
  </Card>

  <Card title="Conditional Routing" icon="route" href="/product/ai-gateway/conditional-routing">
    Route requests to different targets based on specified conditions.
  </Card>

  <Card title="Load Balancing" icon="key" href="/product/ai-gateway/load-balancing">
    Distribute requests across multiple targets based on defined weights.
  </Card>

  <Card title="Caching" icon="database" href="/product/ai-gateway/cache-simple-and-semantic">
    Enable caching of responses to improve performance and reduce costs.
  </Card>

  <Card title="Smart Retries" icon="database" href="/product/ai-gateway/automatic-retries">
    Automatic retry handling with exponential backoff for failed requests
  </Card>

  <Card title="Budget Limits" icon="shield-check" href="/product/model-catalog/budget-limits">
    Set and manage budget limits across teams and departments. Control costs with granular budget limits and usage tracking.
  </Card>
</CardGroup>

### 7. Advanced Guardrails

Protect your Project's data and enhance reliability with real-time checks on LLM inputs and outputs. Leverage guardrails to:

* Prevent sensitive data leaks
* Enforce compliance with organizational policies
* PII detection and masking
* Content filtering
* Custom security rules
* Data compliance checks

<Card title="Guardrails" icon="shield-check" href="/product/guardrails">
  Implement real-time protection for your LLM interactions with automatic detection and filtering of sensitive content, PII, and custom security rules. Enable comprehensive data protection while maintaining compliance with organizational policies.
</Card>

# FAQs

<AccordionGroup>
  <Accordion title="How do I update my AI Provider limits after creation?">
    Update AI Provider limits at any time from [Model Catalog](https://app.portkey.ai/model-catalog): 1. Open the provider you want to modify. 2. Update the budget or rate limits. 3. Save your changes.
  </Accordion>

  <Accordion title="Can I use multiple LLM providers with the same API key?">
    Yes! Add multiple AI Providers to Model Catalog (one for each provider) and attach them to a single config. This config can then be connected to your API key, allowing you to use multiple providers through a single API key.
  </Accordion>

  <Accordion title="How do I track costs for different teams?">
    Portkey provides several ways to track team costs:

    * Create separate AI Providers for each team
    * Use metadata tags in your configs
    * Set up team-specific API keys
    * Monitor usage in the analytics dashboard
  </Accordion>

  <Accordion title="What happens if a team exceeds their budget limit?">
    When a team reaches their budget limit:

    1. Further requests will be blocked
    2. Team admins receive notifications
    3. Usage statistics remain available in dashboard
    4. Limits can be adjusted if needed
  </Accordion>
</AccordionGroup>

# Next Steps

**Join our Community**

* [Discord Community](https://portkey.sh/discord-report)
* [GitHub Repository](https://github.com/Portkey-AI)

<Note>
  For enterprise support and custom features, contact our [enterprise team](https://calendly.com/portkey-ai).
</Note>
