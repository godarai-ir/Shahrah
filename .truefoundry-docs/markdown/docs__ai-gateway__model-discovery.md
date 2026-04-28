> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Model Discovery

> Use the /models API to programmatically list all available models in your TrueFoundry AI Gateway.

The AI Gateway exposes a `/models` endpoint that returns all models your API key has access to. This is useful for building dynamic model selectors, validating model IDs before making requests, and integrating with tools that auto-discover available models.

## List All Models

Send a `GET` request to the `/models` endpoint:

<CodeGroup>
  ```python Python (OpenAI SDK) lines theme={"dark"}
  from openai import OpenAI

  client = OpenAI(
      api_key="your_truefoundry_api_key",
      base_url="https://gateway.truefoundry.ai"
  )

  models = client.models.list()

  for model in models.data:
      print(model.id)
  ```

  ```bash cURL lines theme={"dark"}
  curl https://gateway.truefoundry.ai/models \
    -H "Authorization: Bearer your_truefoundry_api_key"
  ```

  ```typescript Node.js lines theme={"dark"}
  const OpenAI = require('openai');

  const openai = new OpenAI({
      apiKey: 'your_truefoundry_api_key',
      baseURL: 'https://gateway.truefoundry.ai',
  });

  async function main() {
      const models = await openai.models.list();

      for (const model of models.data) {
          console.log(model.id);
      }
  }

  main();
  ```
</CodeGroup>

## Response Format

The endpoint returns an OpenAI-compatible response:

```json lines theme={"dark"}
{
  "object": "list",
  "data": [
    {
      "id": "openai-main/gpt-4o-mini",
      "object": "model",
      "owned_by": "openai-main"
    },
    {
      "id": "anthropic-main/claude-4-sonnet",
      "object": "model",
      "owned_by": "anthropic-main"
    },
    {
      "id": "tfy-ai-bedrock/us-anthropic-claude-sonnet-4-20250514-v1-0",
      "object": "model",
      "owned_by": "tfy-ai-bedrock"
    }
  ]
}
```

<Note>
  The response only includes models that your API key (PAT or VAT) has been granted access to. If a model is missing, check that the relevant provider account has been shared with your user or virtual account. See [Access Control](/docs/ai-gateway/gateway-access-control) for details.
</Note>

## Use Cases

* **Dynamic model selectors**: Build UIs that automatically populate model dropdowns from the API instead of hardcoding model IDs.
* **Validation**: Verify that a model ID exists and is accessible before sending a request.
* **Tool integrations**: Applications like [n8n](/docs/ai-gateway/n8n) use the `/models` endpoint to auto-discover available models.
* **Inventory auditing**: Programmatically audit which models are available across your organization.
