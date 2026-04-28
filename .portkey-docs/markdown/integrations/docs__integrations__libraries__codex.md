> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# OpenAI Codex

> Add usage tracking, cost controls, and security guardrails to Codex with Portkey

**Codex** is OpenAI’s coding agent for terminal, IDE extension, and CLI. It uses a shared config (user-level `~/.codex/config.toml` and optional project-level `.codex/config.toml`) for default model, approval policies, sandbox settings, and provider details. Add Portkey to get:

* **1600+ LLMs** through one interface — switch providers by updating `model` in config
* **Observability** — track costs, tokens, and latency for every request
* **Reliability** — automatic fallbacks, retries, and caching
* **Governance** — budget limits, usage tracking, and team access controls

Configure Codex with Portkey in a few minutes.

<Note>
  For enterprise deployments across teams, see [Enterprise Governance](#3-enterprise-governance).
</Note>

## Quick setup with CLI

One command configures gateway routing, MCP servers, and team skills:

```bash theme={"system"}
npx portkey
```

<Card title="Portkey CLI reference" icon="terminal" href="/guides/coding-agents/agent-cli">
  Full CLI reference — flags, CI usage, and all options
</Card>

***

## 1. Manual setup

<Steps>
  <Step title="Add Provider">
    Go to [Model Catalog](https://app.portkey.ai/model-catalog) → **Add Provider**.

    <Frame>
      <img src="https://mintcdn.com/portkey-docs/Hf1XgyjG_b79ym4Q/Screenshot2025-07-21at5.29.57PM.png?fit=max&auto=format&n=Hf1XgyjG_b79ym4Q&q=85&s=ce12039225e568ea407d3c41264f0034" width="500" data-path="Screenshot2025-07-21at5.29.57PM.png" />
    </Frame>
  </Step>

  <Step title="Configure Credentials">
    Select provider (OpenAI, Anthropic, etc.), enter API key, and create a slug like `openai-prod`.

    <Frame>
      <img src="https://mintcdn.com/portkey-docs/Buc1Vm2P31GSPm3S/images/product/model-catalog/create-provider-page.png?fit=max&auto=format&n=Buc1Vm2P31GSPm3S&q=85&s=621794dd554eb06ec3d5d567b60ea1a9" width="500" data-path="images/product/model-catalog/create-provider-page.png" />
    </Frame>
  </Step>

  <Step title="Get Portkey API Key">
    Go to [API Keys](https://app.portkey.ai/api-keys) and generate your Portkey API key.
  </Step>
</Steps>

## 2. Configure Portkey in Codex

Codex loads config from `~/.codex/config.toml` (overridable with `.codex/config.toml` in a repo; see [Config basics](https://developers.openai.com/codex/config-basic) for precedence).

Add Portkey as the provider by setting `model_provider` and defining `[model_providers.portkey]` with `base_url` and `env_key` (see [Config reference](https://developers.openai.com/codex/config-reference)):

```toml theme={"system"}
model_provider = "portkey"
model = "@openai-prod/gpt-4o"

[model_providers.portkey]
name = "Portkey"
base_url = "https://api.portkey.ai/v1"
env_key = "PORTKEY_API_KEY"
# wire_api = "chat"   # optional: "chat" (default) or "responses"
```

### Multiple model providers

Define multiple entries under `model_providers` to switch between environments or backends by changing `model_provider`:

```toml theme={"system"}
model_provider = "portkey-prod"
model = "@openai-prod/gpt-4o"

[model_providers.portkey-prod]
name = "Portkey (prod)"
base_url = "https://api.portkey.ai/v1"
env_key = "PORTKEY_API_KEY"

[model_providers.portkey-dev]
name = "Portkey (dev)"
base_url = "https://api.portkey.ai/v1"
env_key = "PORTKEY_API_KEY_DEV"
```

Use `.codex/config.toml` in a repository to override `model_provider` and `model` for that project while keeping the shared `~/.codex/config.toml` as the default.

Set `PORTKEY_API_KEY` in the environment:

```shell theme={"system"}
export PORTKEY_API_KEY="<portkey-api-key>"
```

<Note>
  Add to `~/.zshrc` or `~/.bashrc` for persistence.
</Note>

Test the integration:

```shell theme={"system"}
codex "explain this repository to me"
```

Monitor usage in the [Portkey Dashboard](https://app.portkey.ai/dashboard).

## 3. Using Codex with 1600+ Models

Codex uses the `model` value in `config.toml` to decide which model to call. With Portkey, set `model` to a [Model Catalog](/product/model-catalog) slug in the form `@<provider-slug>/<model-name>`. Change providers or models by updating `model` in the config.

Example:

```toml theme={"system"}
model = "@openai-prod/gpt-4o"
```

### Using Responses API

Codex supports OpenAI's Responses API natively. Configure `config.toml` with keys from the [Codex config reference](https://developers.openai.com/codex/config-reference).

**Provider protocol (`wire_api`)**\
Under `[model_providers.portkey]`, set `wire_api` to choose which API protocol Codex uses when talking to the provider:

| Value         | Description                                                                                                                                                |
| ------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `"chat"`      | Chat Completions API (default if omitted). Use for standard chat/completion models.                                                                        |
| `"responses"` | [Responses API](https://developers.openai.com/docs/guides/responses). Use for models that support structured reasoning and tool use via the Responses API. |

Example with protocol and optional retry/timeout tuning:

```toml theme={"system"}
[model_providers.portkey]
name = "Portkey"
base_url = "https://api.portkey.ai/v1"
env_key = "PORTKEY_API_KEY"
wire_api = "responses"
# request_max_retries = 4
# stream_idle_timeout_ms = 300000
```

### Adding Model Capabilities

**Reasoning, output, and tools (top-level)**\
These top-level keys apply to the current session model and control reasoning, output, and tool behavior:

| Key                       | Values                                      | Description                                                                                                              |
| ------------------------- | ------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------ |
| `model_reasoning_effort`  | `minimal`, `low`, `medium`, `high`, `xhigh` | How much reasoning effort the model uses (Responses API). Higher values can improve quality; `xhigh` is model-dependent. |
| `model_reasoning_summary` | `auto`, `concise`, `detailed`, `none`       | How much reasoning summary to include or whether to disable summaries.                                                   |
| `personality`             | `none`, `friendly`, `pragmatic`             | Default communication style for models that support it. Overridable per thread or via `/personality` in-session.         |
| `temperature`             | `0`–`2` (for example `0.1`)                 | Sampling temperature. Lower values make outputs more deterministic; `0.1` is a good default for coding and tooling.      |
| `max_output_tokens`       | Integer (for example `8192`)                | Maximum number of tokens in the response. Prevents run-away output; upper bound is model-dependent.                      |
| `parallel_tool_calls`     | `true` / `false`                            | Allow the model to call multiple tools in parallel when tools are configured.                                            |
| `tool_choice`             | `"auto"`, `"required"`, `"none"`            | Control whether the model decides when to call tools (`"auto"`), must call tools, or never uses tools.                   |

Example:

```toml theme={"system"}
model_provider = "portkey"
model = "@openai-prod/gpt-4o"

model_reasoning_effort = "high"
model_reasoning_summary = "concise"
personality = "pragmatic"
temperature = 0.1
max_output_tokens = 8192
parallel_tool_calls = true
tool_choice = "auto"

[model_providers.portkey]
name = "Portkey"
base_url = "https://api.portkey.ai/v1"
env_key = "PORTKEY_API_KEY"
wire_api = "chat"
```

<Note>
  **Fallbacks, load balancing, or caching?** Create a [Portkey Config](/product/ai-gateway/configs), attach it to your API key, and set `model` to the config’s virtual model. See [Enterprise Governance](#3-enterprise-governance) for examples.
</Note>

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
