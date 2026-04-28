> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Promptfoo

> Run Promptfoo evals on 1600+ LLMs with observability and cost tracking via Portkey.

[Promptfoo](https://promptfoo.dev/docs/intro) is an open source library for evaluating LLM output quality.

With Portkey, you can:

* **Manage prompts** with versioning and call them in Promptfoo
* **Run evals on 1600+ LLMs** including private/local models
* **Track costs and metrics** for all eval runs
* **Avoid rate limits** with load balancing and caching

## 1. Reference Portkey Prompts

Use prompts stored in Portkey directly in Promptfoo:

1. Set `PORTKEY_API_KEY` environment variable
2. Use `portkey://` prefix with your prompt ID:

```yaml theme={"system"}
prompts:
  - portkey://pp-test-prompt-669f48
providers:
  - openai:gpt-3.5-turbo-0613
tests:
  - vars:
      topic: ...
```

Variables from test cases are automatically plugged into the Portkey prompt.

<Note>
  Promptfoo doesn't follow temperature, model, and other parameters set in Portkey. Set them in the providers configuration.
</Note>

## 2. Route to Any Provider

Set `PORTKEY_API_KEY` and configure providers with the `portkey:` prefix:

<CodeGroup>
  ```yaml OpenAI theme={"system"}
  providers:
    id: portkey:gpt-4o
    config:
      portkeyProvider: openai
  ```

  ```yaml Anthropic theme={"system"}
  providers:
    id: portkey:claude-3-opus-20240229
    config:
      portkeyProvider: anthropic
      portkeyVirtualKey: "@anthropic-prod"
  ```

  ```yaml Google Gemini theme={"system"}
  providers:
    id: portkey:gemini-1.5-flash-latest
    config:
      portkeyProvider: google
      portkeyVirtualKey: "@google-prod"
  ```

  ```yaml Groq theme={"system"}
  providers:
    id: portkey:llama3-8b-8192
    config:
      portkeyProvider: groq
      portkeyVirtualKey: "@groq-prod"
  ```

  ```yaml Ollama (Local) theme={"system"}
  providers:
    id: portkey:llama3
    config:
      portkeyProvider: ollama
      portkeyCustomHost: YOUR_OLLAMA_NGROK_URL
  ```
</CodeGroup>

### Cloud Providers (Azure, Bedrock, Vertex)

<CodeGroup>
  ```yaml Azure OpenAI theme={"system"}
  providers:
    id: portkey:xxx
    config:
      portkeyVirtualKey: "@azure-openai-prod"
  ```

  ```yaml AWS Bedrock theme={"system"}
  providers:
    id: portkey:anthropic.claude-3-sonnet-20240229-v1:0
    config:
      portkeyVirtualKey: "@bedrock-prod"
  ```

  ```yaml Google Vertex AI theme={"system"}
  providers:
    id: portkey:gemini-1.5-flash-latest
    config:
      portkeyVirtualKey: "@vertex-prod"
  ```
</CodeGroup>

## 3. Track Costs & Metrics

Add metadata to segment requests and track costs per team/project:

```yaml theme={"system"}
providers:
  id: portkey:claude-3-opus-20240229
  config:
    portkeyVirtualKey: "@anthropic-prod"
    portkeyMetadata:
      team: alpha9
      prompt: classification
    portkeyTraceId: run_1
```

Filter and group by these keys in Portkey dashboards.

<Frame>
  <img src="https://mintcdn.com/portkey-docs/wAHXB_jjwLt8bYcN/images/libraries/promptfoo.gif?s=d673a48902e326f5bab883a1b664725f" width="1648" height="1080" data-path="images/libraries/promptfoo.gif" />
</Frame>

## 4. Avoid Rate Limits with Caching

Create a config with load balancing and caching for high-volume evals:

```json theme={"system"}
{
  "cache": { "mode": "simple" },
  "strategy": { "mode": "loadbalance" },
  "targets": [
    { "provider": "@openai-key-1" },
    { "provider": "@openai-key-2" },
    { "provider": "@openai-key-3" }
  ]
}
```

Reference the config in your YAML:

```yaml theme={"system"}
providers:
  id: portkey:claude-3-opus-20240229
  config:
    portkeyConfig: pc-your-config-id
```

***

## Next Steps

<CardGroup cols={2}>
  <Card title="Model Catalog" icon="list" href="/product/model-catalog">
    Set up providers
  </Card>

  <Card title="Prompt Management" icon="message" href="/product/prompt-library">
    Version and manage prompts
  </Card>
</CardGroup>
