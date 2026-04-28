> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Model Catalog

> A single pane to view and manage every AI provider and model in your organization. It provides centralized governance, discovery, and usage controls for all your AI resources.

The Model Catalog is a centralized hub for viewing and managing all AI providers and models within your organization. It abstracts raw API keys and scattered environment variables into governed Provider Integrations and Models, giving you complete control over how your teams access and use AI.

<Note>
  #### [Upgrading from Virtual Keys](/support/upgrade-to-model-catalog)

  The Model Catalog upgrades the Virtual Key experience by introducing a centralized, organization-level management layer, offering advantages like:

  * Centralized provider and model management - no more duplicate configs across workspaces.
  * Fine-grained control: budgets, rate limits, and model allow-lists at both org and workspace level.
  * Inline usage: just pass `model="@provider/model_slug"`

  **Need help?** See our [Migration Guide ➜](/support/upgrade-to-model-catalog)
</Note>

<Frame>
  <img src="https://mintcdn.com/portkey-docs/TEsf28na2t53tvAY/providersandmodels.gif?s=c5cf9cdfc845851cb59f24c1bc6d4df2" alt="Model Catalog - Provider and Models" width="1532" height="1080" data-path="providersandmodels.gif" />
</Frame>

<CardGroup cols={2}>
  <Card title="AI Providers" icon="sparkles" color="#444">
    AI Providers are what you use in your code. Each provider has:

    * ✅ A unique slug (e.g., `@openai-prod`)
    * ✅ Securely stored credentials
    * ✅ Budget and rate limits
    * ✅ Access to specific models

    **To use:** Add a provider, then use `@provider-slug/model-name` in your code.
  </Card>

  <Card title="Models" icon="microchip-ai" iconType="regular" color="#444">
    The Models section is a gallery of all AI models available in your workspace. Each Model entry includes:

    * ✅ Model slug (`@openai-prod/gpt-4o`)
    * ✅ Ready-to-use code snippets
    * ✅ Input/output token limits
    * ✅ Pricing information (where available)

    [View all available models →](https://app.portkey.ai/model-catalog/models)
  </Card>
</CardGroup>

## Adding an AI Provider

Add providers via **UI** (follow the steps below) or [**API**](/api-reference/admin-api/introduction).

<Steps>
  <Step title="Go to AI Providers → Add Provider">
    Navigate to the [Model Catalog](https://app.portkey.ai/model-catalog) in your Portkey dashboard.

    <Frame>
      <img src="https://mintcdn.com/portkey-docs/Hf1XgyjG_b79ym4Q/Screenshot2025-07-21at5.29.57PM.png?fit=max&auto=format&n=Hf1XgyjG_b79ym4Q&q=85&s=ce12039225e568ea407d3c41264f0034" alt="Portkey Model Catalog - Add Provider" width="1434" height="350" data-path="Screenshot2025-07-21at5.29.57PM.png" />
    </Frame>
  </Step>

  <Step title="Select the AI Service">
    Choose from list (OpenAI, Anthropic, etc.) or *Self-hosted / Custom*.

    <Frame>
      <img src="https://mintcdn.com/portkey-docs/Hf1XgyjG_b79ym4Q/Screenshot2025-07-21at5.31.49PM.png?fit=max&auto=format&n=Hf1XgyjG_b79ym4Q&q=85&s=c4e14c1a424afb6ce8376e69090b3209" alt="Portkey Model Catalog - Add Provider - Choose Service" width="1600" height="686" data-path="Screenshot2025-07-21at5.31.49PM.png" />
    </Frame>
  </Step>

  <Step title="Choose or Create Credentials">
    **If credentials already exist:**

    * Select from the dropdown (if your org admin set them up)
    * Skip to step 4 - no API keys needed!

    **If creating new credentials:**

    * Choose "Create new credentials"
    * Enter your API keys here

    <Note>
      Creating new credentials here automatically creates a workspace-linked integration. To share credentials across multiple workspaces, create them in the [Integrations](/product/model-catalog/integrations) page (org admin only).
    </Note>

    <Frame>
      <img src="https://mintcdn.com/portkey-docs/Hf1XgyjG_b79ym4Q/Screenshot2025-07-21at5.34.13PM.png?fit=max&auto=format&n=Hf1XgyjG_b79ym4Q&q=85&s=a090324137da4dde5cf5afca7d2180bb" alt="Model Catalog - Add credentials" width="1440" height="810" data-path="Screenshot2025-07-21at5.34.13PM.png" />
    </Frame>
  </Step>

  <Step title="Name your provider & save">
    Choose a name and slug for this provider. The slug (e.g., `openai-prod`) will be used in your code like `@openai-prod/gpt-4o`.

    <Frame>
      <img src="https://mintcdn.com/portkey-docs/Hf1XgyjG_b79ym4Q/Screenshot2025-07-21at5.36.14PM.png?fit=max&auto=format&n=Hf1XgyjG_b79ym4Q&q=85&s=f51a2b09c4a2925263acb76571230504" alt="Model Catalog - Add Provider Details" width="1254" height="682" data-path="Screenshot2025-07-21at5.36.14PM.png" />
    </Frame>
  </Step>
</Steps>

## Using Provider Models

Once you have AI Providers set up, use their models in your applications. There are three methods - we recommend the model prefix format for clarity.

In Portkey, model strings follow this format:

`@provider_slug/model_name`

<Frame>
  <img src="https://mintcdn.com/portkey-docs/Hf1XgyjG_b79ym4Q/Screenshot2025-07-21at5.39.59PM.png?fit=max&auto=format&n=Hf1XgyjG_b79ym4Q&q=85&s=884eb4c9a28e26166467d1caa461eabc" alt="Model String Format" width="1052" height="252" data-path="Screenshot2025-07-21at5.39.59PM.png" />
</Frame>

For example, `@openai-prod/gpt-4o`, `@anthropic/claude-3-sonnet`, `@bedrock-us/claude-3-sonnet-v1`

<CodeGroup>
  ```javascript Javascript SDK theme={"system"}
  import { Portkey } from 'portkey-ai';
  const client = new Portkey({ apiKey: "PORTKEY_API_KEY" });

  const resp = await client.chat.completions.create({
    model: '@openai-prod/gpt-4o',
    messages: [{ role: 'user', content: 'Hello!' }]
  });
  ```

  ```python Python SDK theme={"system"}
  from portkey_ai import Portkey

  portkey = Portkey(api_key="PORTKEY_API_KEY")

  resp = portkey.chat.completions.create(
    model="@openai-prod/gpt-4o",
    messages=[{"role":"user","content":"Hello"}]
  )
  ```

  ```python OpenAI Python SDK theme={"system"}
  from openai import OpenAI

  client = OpenAI(api_key="PORTKEY_API_KEY", base_url="https://api.portkey.ai/v1")

  client.chat.completions.create(
  	model="@openai-prod/gpt-4o",
  	messages=[{"role": "user", "content": "Hello!"}]
  )
  ```

  ```javascript OpenAI JS SDK theme={"system"}
  import OpenAI from 'openai';

  const openai = new OpenAI({
      apiKey: "PORTKEY_API_KEY",
      baseURL: "https://api.portkey.ai/v1"
  });

  const completion = await openai.chat.completions.create({
      model: "@openai-prod/gpt-4o",
      messages: [{ role: 'user', content: 'Hello!' }]
  });
  ```

  ```bash cURL theme={"system"}
  curl https://api.portkey.ai/v1/chat/completions \
    -H "Content-Type: application/json" \
    -H "x-portkey-api-key: $PORTKEY_API_KEY" \
    -d '{
      "model": "@openai-prod/gpt-4o",
      "messages": [{"role": "user", "content": "Hello!"}]
    }'
  ```
</CodeGroup>

### 2. Using the `provider` header

Specify the provider separately using the `provider` parameter. Remember to add the `@` before your provider slug.

<CodeGroup>
  ```javascript Javascript SDK theme={"system"}
  import { Portkey } from 'portkey-ai';
  const client = new Portkey({
  	apiKey: "PORTKEY_API_KEY",
  	provider: "@openai-prod"
  });

  const resp = await client.chat.completions.create({
    model: 'gpt-4o',
    messages: [{ role: 'user', content: 'Hello!' }]
  });
  ```

  ```python Python SDK theme={"system"}
  from portkey_ai import Portkey

  portkey = Portkey(
  	api_key="PORTKEY_API_KEY",
  	provider="@openai-prod"
  )

  resp = portkey.chat.completions.create(
    model="gpt-4o",
    messages=[{"role":"user","content":"Hello"}]
  )
  ```

  ```python OpenAI Python SDK theme={"system"}
  from openai import OpenAI
  from portkey_ai import createHeaders

  client = OpenAI(
  	api_key="PORTKEY_API_KEY",
  	base_url="https://api.portkey.ai/v1",
  	default_headers=createHeaders(
  		provider="@openai-prod"
  	)
  )

  client.chat.completions.create(
  	model="gpt-4o",
  	messages=[{"role": "user", "content": "Hello!"}]
  )
  ```

  ```javascript OpenAI JS SDK theme={"system"}
  import OpenAI from 'openai';
  import { createHeaders } from 'portkey-ai'

  const openai = new OpenAI({
      apiKey: "PORTKEY_API_KEY",
      baseURL: "https://api.portkey.ai/v1",
  	defaultHeaders: createHeaders({
  		provider: "@openai-prod"
  	})
  });

  const completion = await openai.chat.completions.create({
      model: "gpt-4o",
      messages: [{ role: 'user', content: 'Hello!' }]
  });
  ```

  ```bash cURL theme={"system"}
  curl https://api.portkey.ai/v1/chat/completions \
    -H "Content-Type: application/json" \
    -H "x-portkey-api-key: $PORTKEY_API_KEY" \
    -H "x-portkey-provider: @openai-prod" \
    -d '{
      "model": "gpt-4o",
      "messages": [{"role": "user", "content": "Hello!"}]
    }'
  ```
</CodeGroup>

### 3. Specify `provider` in the config

Portkey's configs are simple JSON structures that help you define routing logic for LLM requests. Learn more [here](/product/ai-gateway/configs).

There are three ways to specify providers in configs:

**Method 1: Model in override\_params (Recommended)**

Specify provider and model together in `override_params`. Works great with multi-provider strategies:

```json theme={"system"}
{
	"strategy": { "mode": "fallback" },
	"targets": [{
		"override_params": { "model": "@openai-prod/gpt-4o" }
	}, {
		"override_params": { "model": "@anthropic/claude-3-sonnet" }
	}]
}
```

**Method 2: Provider in target**

Specify provider directly in the target (remember the `@` symbol):

```json theme={"system"}
{
	"strategy": { "mode": "single" },
	"targets": [{
		"provider": "@openai-prod",
		"override_params": {
			"model": "gpt-4o"
		}
	}]
}
```

**Method 3: Legacy virtual\_key (Backwards Compatible)**

The `virtual_key` field still works:

```json theme={"system"}
{
	"strategy": { "mode": "single" },
	"targets": [{
		"virtual_key": "openai-prod"
	}]
}
```

> **Ordering:** `config` (if provided) defines base; `override_params` merges on top (last write wins for scalars, deep merge for objects like `metadata`).

## Integrations

At the heart of Model Catalog is a simple concept: your AI provider credentials need to be stored securely, governed carefully and managed centrally. In Portkey, these stored credentials are called **Integrations**. Think of an Integration as a secure vault for your API keys - whether it's your OpenAI API key, AWS Bedrock credentials, or Azure OpenAI configuration.

<Frame>
  <img src="https://mintcdn.com/portkey-docs/Buc1Vm2P31GSPm3S/images/product/model-catalog/integrations-page.png?fit=max&auto=format&n=Buc1Vm2P31GSPm3S&q=85&s=df13f2b4eeb68e349b133a53344a543a" alt="Integrations Overview Page" width="2926" height="1767" data-path="images/product/model-catalog/integrations-page.png" />
</Frame>

Once you create an Integration (by storing your credentials), you can use it to create multiple AI Providers. For example, you might have one OpenAI Integration, but create three different AI Providers from it:

* `@openai-dev` for development with strict rate limits
* `@openai-staging` for testing with moderate budgets
* `@openai-prod` for production with higher limits

This separation gives you granular control over how different teams and environments use the same underlying credentials.

<Card title="Integrations" icon="key" href="/product/model-catalog/integrations">
  Learn how to create and manage AI service credentials across your organization
</Card>

## Managing Access and Controls

Each Integration in Portkey acts as a control point where you can configure:

### Budget Limits

Set spending controls at the Integration level to prevent unexpected costs. You can configure:

* **Cost-based limits**: Maximum spend in USD (e.g., \$1000/month)
* **Token-based limits**: Maximum tokens consumed (e.g., 10M tokens/week)
* **Periodic resets**: Weekly or monthly budget refreshes

<Frame>
  <img src="https://mintcdn.com/portkey-docs/Buc1Vm2P31GSPm3S/images/product/model-catalog/budget-and-limits-page-model-catalog.png?fit=max&auto=format&n=Buc1Vm2P31GSPm3S&q=85&s=a7debfad7682a196e68a8e30d6aa5a15" alt="Budget Limits Configuration" width="2523" height="1763" data-path="images/product/model-catalog/budget-and-limits-page-model-catalog.png" />
</Frame>

These limits cascade down to all AI Providers created from that Integration.

<Card title="Budget Management" icon="dollar-sign" href="/product/model-catalog/integrations#3-budget-%26-rate-limits">
  Set up cost controls and spending limits for your AI usage
</Card>

### Rate Limits

Control request rates to manage load and prevent abuse:

* **Requests per minute/hour/day**: Set appropriate throughput limits
* **Concurrent request limits**: Control parallel processing
* **Burst protection**: Prevent sudden spikes in usage

Rate limits help you maintain service quality and prevent any single user or team from monopolizing resources.

<Card title="Rate Limiting" icon="gauge" href="/product/model-catalog/integrations#3-budget-%26-rate-limits">
  Configure request rate controls to ensure fair usage and prevent abuse
</Card>

### Workspace Provisioning

Control which workspaces in your organization can access specific AI Providers:

* **Selective access**: Choose which teams can use production vs development providers
* **Environment isolation**: Keep staging and production resources separate
* **Department-level control**: Give finance different access than engineering

<Frame>
  <img src="https://mintcdn.com/portkey-docs/Buc1Vm2P31GSPm3S/images/product/model-catalog/workspace-provisioning-page.png?fit=max&auto=format&n=Buc1Vm2P31GSPm3S&q=85&s=6e1a86e024e3c03a44d8d30152efbaaa" alt="Workspace Provisioning Interface" width="2525" height="1771" data-path="images/product/model-catalog/workspace-provisioning-page.png" />
</Frame>

This hierarchical approach ensures teams only have access to the resources they need.

<Card title="Workspace Provisioning" icon="building" href="/product/model-catalog/integrations#1-workspace-provisioning">
  Manage workspace access to AI providers and models
</Card>

### Model Provisioning

Fine-tune which models are available through each Integration:

* **Model allowlists**: Only expose specific models (e.g., only GPT-4 for production)
* **Model denylists**: Block access to expensive or experimental models
* **Custom model addition**: Add your fine-tuned or self-hosted models

<Frame>
  <img src="https://mintcdn.com/portkey-docs/Buc1Vm2P31GSPm3S/images/product/model-catalog/model-provisioning-page.png?fit=max&auto=format&n=Buc1Vm2P31GSPm3S&q=85&s=dd617273c3e1cc352c5f1276631dc07c" alt="Model Provisioning Settings" width="2536" height="1774" data-path="images/product/model-catalog/model-provisioning-page.png" />
</Frame>

Model provisioning helps you maintain consistency and control costs across your organization.

<Card title="Model Provisioning" icon="filter" href="/product/model-catalog/integrations#2-model-provisioning">
  Configure which models are available through each integration
</Card>

## Advanced Model Management

### Custom Models

The Model Catalog isn't limited to standard provider models. You can add:

* **Fine-tuned models**: Your custom OpenAI or Anthropic fine-tunes
* **Self-hosted models**: Models running on your infrastructure
* **Private models**: Internal models not publicly available

Each custom model gets the same governance controls as standard models.

<Card title="Custom Models" icon="wrench" href="/product/model-catalog/custom-models">
  Add and manage your fine-tuned, self-hosted, or private models
</Card>

### Overriding Model Details (Custom Pricing)

Override default model pricing for:

* **Negotiated rates**: If you have enterprise agreements with providers
* **Internal chargebacks**: Set custom rates for internal cost allocation
* **Free tier models**: Mark certain models as free for specific teams

Custom pricing ensures your cost tracking accurately reflects your actual spend.

<Card title="Custom Pricing" icon="tag" href="/product/model-catalog/model-overrides">
  Configure custom pricing for models with special rates
</Card>

## Self-hosted AI Providers

<Card title="Coming Soon" />
