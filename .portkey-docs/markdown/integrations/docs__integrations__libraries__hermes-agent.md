> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Hermes Agent

> Route the Hermes autonomous agent through Portkey for observability, cost tracking, fallbacks, and access to 200+ LLMs

[Hermes Agent](https://hermes-agent.nousresearch.com/) is an open-source autonomous agent from [Nous Research](https://nousresearch.com/) that runs on your server — not tethered to an IDE. It lives on CLI, Telegram, Discord, Slack, WhatsApp, and 10+ other platforms, with persistent memory, self-improving skills, and full terminal/browser control.

Hermes works with any OpenAI-compatible endpoint, which makes Portkey a drop-in gateway. Route Hermes through Portkey to get full observability, cost tracking, budget controls, fallbacks, and access to 200+ models — without changing how you use Hermes.

## Quick start

<Steps>
  <Step title="Install Hermes">
    ```bash theme={"system"}
    curl -fsSL https://hermes-agent.nousresearch.com/install.sh | bash
    ```

    Supports Linux, macOS, and WSL2. See the [Hermes installation docs](https://hermes-agent.nousresearch.com/docs/getting-started/installation) for details.
  </Step>

  <Step title="Add a provider in Portkey">
    Go to [AI Providers](https://app.portkey.ai/ai-providers) → **Add Provider** → select your LLM provider (OpenAI, Anthropic, Google, etc.).

    Enter your provider API key and create a slug like `openai-prod` or `anthropic-prod`.

    <Frame>
      <img src="https://mintcdn.com/portkey-docs/Buc1Vm2P31GSPm3S/images/product/model-catalog/create-provider-page.png?fit=max&auto=format&n=Buc1Vm2P31GSPm3S&q=85&s=621794dd554eb06ec3d5d567b60ea1a9" width="500" data-path="images/product/model-catalog/create-provider-page.png" />
    </Frame>
  </Step>

  <Step title="Get your Portkey API key">
    Go to [API Keys](https://app.portkey.ai/api-keys) → **Generate a new key**.
  </Step>

  <Step title="Configure Hermes to use Portkey">
    Edit `~/.hermes/config.yaml` to point the main model at Portkey's OpenAI-compatible endpoint:

    ```yaml theme={"system"}
    model:
      provider: custom
      default: "@openai-prod/gpt-4o"
      base_url: https://api.portkey.ai/v1
      api_key: YOUR_PORTKEY_API_KEY
    ```

    Replace:

    * `YOUR_PORTKEY_API_KEY` with your Portkey API key
    * `@openai-prod` with your provider slug
    * `gpt-4o` with any model that provider supports

    The model format is `@<provider-slug>/<model-name>` — this maps directly to a provider integration in your Portkey workspace.

    <Tip>
      Use `hermes config set` to save values without opening the file:

      ```bash theme={"system"}
      hermes config set model.provider custom
      hermes config set model.base_url https://api.portkey.ai/v1
      hermes config set model.default "@openai-prod/gpt-4o"
      hermes config set OPENAI_API_KEY YOUR_PORTKEY_API_KEY
      ```

      API keys are saved to `~/.hermes/.env`; everything else goes to `config.yaml`.
    </Tip>
  </Step>

  <Step title="Start chatting">
    ```bash theme={"system"}
    hermes chat
    ```

    All Hermes requests now route through Portkey. Monitor usage, costs, and traces in the [Portkey Dashboard](https://app.portkey.ai/logs).
  </Step>
</Steps>

## Reliability

All reliability controls are configured through [Portkey Configs](/product/ai-gateway/configs) and attached to a scoped API key. Hermes sends the key as a Bearer token — Portkey applies the logic server-side, so Hermes config stays simple.

Create a Config at [Configs](https://app.portkey.ai/configs), then attach the Config ID to a scoped API key at [API Keys](https://app.portkey.ai/api-keys) → **Create Key** → **Advanced Settings**. Use that scoped key as `OPENAI_API_KEY` in Hermes.

### Fallbacks

Route to backup providers when the primary fails — critical for long-running autonomous Hermes sessions where a provider outage shouldn't break the task:

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

Distribute requests across multiple keys or providers:

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

Reduce costs and latency for repeated queries (common with Hermes's scheduled cron jobs):

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

### Metadata for tracing

Attach metadata in the same Config to group and filter logs by session, host, environment, or user:

```json theme={"system"}
{
    "provider": "@openai-prod",
    "metadata": {
        "agent": "hermes",
        "environment": "production",
        "host": "vps-01"
    }
}
```

## Budget limits

Hermes runs unattended via cron, messaging gateways, and subagents — costs can spiral without controls. Set provider-level limits in Portkey:

1. Go to [AI Providers](https://app.portkey.ai/ai-providers) → select your provider
2. Click **Budget & Limits**
3. Configure:
   * **Cost limit**: e.g., \$200/month
   * **Token limit**: e.g., 10M tokens/week
   * **Rate limit**: requests per minute

<Frame>
  <img src="https://mintcdn.com/portkey-docs/Buc1Vm2P31GSPm3S/images/product/model-catalog/budget-and-limits-page-model-catalog.png?fit=max&auto=format&n=Buc1Vm2P31GSPm3S&q=85&s=a7debfad7682a196e68a8e30d6aa5a15" width="500" data-path="images/product/model-catalog/budget-and-limits-page-model-catalog.png" />
</Frame>

Budget limits automatically block further requests when thresholds are hit, preventing runaway costs from autonomous sessions.

## Switch models mid-session

Inside a Hermes chat, use `/model` to switch to any model available in your Portkey workspace:

```text theme={"system"}
/model custom:@anthropic-prod/claude-sonnet-4-6
/model custom:@gemini-prod/gemini-2.5-pro
```

## Multiple Portkey providers

To keep multiple Portkey-backed routes distinct — say, a production key for your main agent and a separate key for subagents — define them as [named custom providers](https://hermes-agent.nousresearch.com/docs/integrations/providers#named-custom-providers) in `~/.hermes/config.yaml`:

```yaml theme={"system"}
custom_providers:
  - name: portkey-prod
    base_url: https://api.portkey.ai/v1
    key_env: PORTKEY_API_KEY
    api_mode: chat_completions
  - name: portkey-dev
    base_url: https://api.portkey.ai/v1
    key_env: PORTKEY_DEV_KEY
    api_mode: chat_completions

model:
  provider: custom
  default: "@openai-prod/gpt-4o"
  base_url: https://api.portkey.ai/v1
  api_key: ${PORTKEY_API_KEY}
```

Set both keys in `~/.hermes/.env`:

```bash theme={"system"}
PORTKEY_API_KEY=pk-...
PORTKEY_DEV_KEY=pk-...
```

Switch between them mid-session:

```text theme={"system"}
/model custom:portkey-prod:@openai-prod/gpt-4o
/model custom:portkey-dev:@anthropic-prod/claude-sonnet-4-6
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
