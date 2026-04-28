> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Making Your First Request

> Get your Gateway Base URL, Model ID, and API key to start calling LLMs through the TrueFoundry AI Gateway.

To call models through the TrueFoundry AI Gateway, you need three things:

1. **Gateway Base URL** — the endpoint for the AI Gateway service
2. **Model ID** — the identifier of the model you want to call
3. **API Key** — a TrueFoundry authentication token (Personal Access Token or Virtual Account Token)

All three are available in the **Code Snippets** tab of the [Playground](/docs/ai-gateway/playground-overview). Open the Playground, select a model, and click the **Code** button to see them pre-filled in a ready-to-use code snippet.

<Columns cols={2}>
  <Frame caption="Code Snippets tab in the Playground showing Base URL, API Key, and Model ID">
    <img src="https://mintcdn.com/truefoundry/Pv3zXte3KtZeCZSK/images/playground-new.png?fit=max&auto=format&n=Pv3zXte3KtZeCZSK&q=85&s=86c9cb57fd8294257617ee71f05e56e0" alt="Gateway Code Snippets Page with code generation options" width="2686" height="1720" data-path="images/playground-new.png" />
  </Frame>

  <Frame caption="Base URL, Model ID, and API Key fields in the Playground">
    <img src="https://mintcdn.com/truefoundry/yR_clVDeJDlQkXKY/images/docs/ai-gateway/llm-gateway-base-url-model-id.png?fit=max&auto=format&n=yR_clVDeJDlQkXKY&q=85&s=d79789c1dc618aaf2d8c96eaee93660f" alt="Playground header showing Base URL, Model ID, and API Key fields" width="1024" height="494" data-path="images/docs/ai-gateway/llm-gateway-base-url-model-id.png" />
  </Frame>
</Columns>

## Gateway Base URL

The Gateway Base URL is the single endpoint your application sends all LLM requests to. Instead of calling each provider (OpenAI, Anthropic, Azure, etc.) at their own endpoints, you point every request at this one URL. The Gateway routes it to the right provider behind the scenes.

| Deployment           | Base URL                                             |
| -------------------- | ---------------------------------------------------- |
| **TrueFoundry SaaS** | `https://gateway.truefoundry.ai`                     |
| **Self-hosted**      | Found in the Code Snippets section of the Playground |

The Gateway Base URL is compatible with the OpenAI API format, so any SDK or tool that supports the OpenAI API can work with TrueFoundry by simply changing the base URL.

## Model ID

Every model accessible through the Gateway has a unique identifier in the format `provider_account/model_name`.

For example, `openai-main/gpt-4o-mini` means the model `gpt-4o-mini` available through the provider account named `openai-main`. The provider account represents a configured connection to an LLM provider — your organization may have multiple accounts for the same provider (e.g., `openai-main` for production and `openai-dev` for testing).

## API Key (Authentication)

To access models through the AI Gateway, you authenticate with a **TrueFoundry API key** — not the original provider keys (OpenAI, Anthropic, etc.). The Gateway manages provider credentials centrally, so developers only need a single TrueFoundry token.

TrueFoundry supports two types of API keys:

| Token type                                                                                   | Best for                |
| -------------------------------------------------------------------------------------------- | ----------------------- |
| **[Personal Access Token (PAT)](/docs/ai-gateway/authentication#personal-access-token-pat)** | Development and testing |
| **[Virtual Account Token (VAT)](/docs/ai-gateway/authentication#virtual-account-token-vat)** | Production applications |

For detailed instructions on creating and managing tokens, see the [Authentication](/docs/ai-gateway/authentication) documentation.

## Getting Code Snippets from the Playground

The fastest way to start making requests is to copy a code snippet directly from the Playground. The generated snippets come pre-filled with the correct Base URL, Model ID, and your API key — so you can paste and run immediately.

1. Open the **Playground** and select a model
2. Click the **Code** button in the top bar
3. Choose your language or framework
4. Copy the generated code into your application

Available snippet formats: `OpenAI SDK`, `Langchain`, `Langgraph`, `Google ADK`, `Stream API`, `Rest API`, `Go-OpenAI`, `Rust-OpenAI`, `cURL`, `Node.js`, `LlamaIndex`, `Langchain4j`

<Frame caption="Code snippet generation with language options">
  <img src="https://mintcdn.com/truefoundry/yR_clVDeJDlQkXKY/images/docs/ai-gateway/llm-gateway-code-snippets.png?fit=max&auto=format&n=yR_clVDeJDlQkXKY&q=85&s=bc6285331c363d6b42e659f9953ba158" alt="Code snippet generation interface with language options" width="3840" height="1858" data-path="images/docs/ai-gateway/llm-gateway-code-snippets.png" />
</Frame>
