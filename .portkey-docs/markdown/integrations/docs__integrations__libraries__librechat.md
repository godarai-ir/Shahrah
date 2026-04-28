> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# LibreChat

> Cost tracking, observability, and more for LibreChat

Add Portkey to LibreChat to get:

* **Unified access to 1600+ LLMs** through a single API
* **Real-time observability** with 40+ metrics and detailed logs
* **Enterprise governance** with budget limits and RBAC
* **Security guardrails** for PII detection and content filtering

<Note>
  For enterprise governance setup, see [Enterprise Governance](#3-enterprise-governance).
</Note>

## 1. Setup Portkey

<Steps>
  <Step title="Add Provider">
    Go to [Model Catalog → AI Providers](https://app.portkey.ai/model-catalog) and add your provider (OpenAI, Anthropic, etc.) with your API credentials.

    <Frame>
      <img src="https://mintcdn.com/portkey-docs/Hf1XgyjG_b79ym4Q/Screenshot2025-07-21at5.29.57PM.png?fit=max&auto=format&n=Hf1XgyjG_b79ym4Q&q=85&s=ce12039225e568ea407d3c41264f0034" alt="Add Provider" width="1434" height="350" data-path="Screenshot2025-07-21at5.29.57PM.png" />
    </Frame>
  </Step>

  <Step title="Create Config (Optional)">
    Go to [Configs](https://app.portkey.ai/configs) and create a config for routing, fallbacks, or other features:

    ```json theme={"system"}
    {"override_params": {"model": "@openai-prod/gpt-4o"}}
    ```

    <Frame>
      <img src="https://mintcdn.com/portkey-docs/T0lFtdapIPX8YtCI/images/integrations/config.png?fit=max&auto=format&n=T0lFtdapIPX8YtCI&q=85&s=1a7edd4ef3643cae7fec8595b6d315d7" width="500" data-path="images/integrations/config.png" />
    </Frame>
  </Step>

  <Step title="Create Portkey API Key">
    Go to [API Keys](https://app.portkey.ai/api-keys) → **Create New API Key**. Optionally attach your config from Step 2.

    <Frame>
      <img src="https://mintcdn.com/portkey-docs/T0lFtdapIPX8YtCI/images/integrations/api-key.png?fit=max&auto=format&n=T0lFtdapIPX8YtCI&q=85&s=9ac7b990452c59e9bd167761abbd30da" width="500" data-path="images/integrations/api-key.png" />
    </Frame>
  </Step>
</Steps>

## 2. Integrate with LibreChat

### Configure Files

**docker-compose.override.yml** ([docs](https://www.librechat.ai/docs/quick_start/custom_endpoints))

```yaml docker-compose.override.yml theme={"system"}
services:
  api:
    volumes:
    - type: bind
      source: ./librechat.yaml
      target: /app/librechat.yaml
```

**.env**

```env .env theme={"system"}
PORTKEY_API_KEY=YOUR_PORTKEY_API_KEY
PORTKEY_GATEWAY_URL=https://api.portkey.ai/v1
```

**librechat.yaml** - Choose one option:

<CodeGroup>
  ```yaml With Config theme={"system"}
  version: 1.1.4
  cache: true
  endpoints:
    custom:
      - name: "Portkey"
        apiKey: "dummy"
        baseURL: ${PORTKEY_GATEWAY_URL}
        headers:
          x-portkey-api-key: "${PORTKEY_API_KEY}"
          x-portkey-config: "pc-libre-xxx"
        models:
          default: ["@openai-prod/gpt-4o"]
          fetch: true
        titleConvo: true
        titleModel: "current_model"
        modelDisplayLabel: "Portkey"
  ```

  ```yaml With Provider Slug theme={"system"}
  version: 1.1.4
  cache: true
  endpoints:
    custom:
      - name: "Portkey"
        apiKey: "dummy"
        baseURL: ${PORTKEY_GATEWAY_URL}
        headers:
          x-portkey-api-key: "${PORTKEY_API_KEY}"
          x-portkey-provider: "@openai-prod"
        models:
          default: ["gpt-4o"]
          fetch: true
        titleConvo: true
        titleModel: "current_model"
        modelDisplayLabel: "Portkey:OpenAI"
  ```
</CodeGroup>

<Note>
  LibreChat requires an `apiKey` field—use `"dummy"` since auth is via Portkey headers.
</Note>

<Note>
  For per-user cost tracking in centralized deployments, see [this community guide](https://github.com/timmanik/librechat-for-portkey).
</Note>

### Gemini models

Gemini models are not fully compatible with OpenAI chat completions in certain cases like tool calling where thought signature is required for the Gemini models but not supported by OpenAI signature. Here is the recommended configuration:

```yaml librechat.yaml theme={"system"}
# Definition of custom endpoints
endpoints:
  custom:
    - name: "Gemini through Portkey"
      apiKey: "dummy"
      baseURL: 'https://api.portkey.ai/v1'
      headers:
        x-portkey-api-key: "asdT"
        x-portkey-provider: "@gemini-prod"
      models:
        default: ['gemini-3-pro-preview', 'gemini-2.5-flash', 'gemini-2.5-flash-preview']
        fetch: false
      customParams:
        defaultParamsEndpoint: 'google'
      titleConvo: true
      titleModel: 'current_model'
      summarize: false
      summaryModel: 'current_model'
      forcePrompt: false
      modelDisplayLabel: 'Portkey'
      iconURL: https://images.crunchbase.com/image/upload/c_pad,f_auto,q_auto:eco,dpr_1/rjqy7ghvjoiu4cd1xjbf
```

The key difference from the regular configuration is:

```yaml theme={"system"}
customParams:
  defaultParamsEndpoint: 'google'
```

This instructs LibreChat to send the payload in Gemini format.

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
