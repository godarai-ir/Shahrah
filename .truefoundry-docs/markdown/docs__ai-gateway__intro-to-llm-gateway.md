> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Introduction to AI Gateway

> TrueFoundry AI Gateway: A unified interface for accessing 1000+ LLMs with enterprise-grade security, observability, and governance

TrueFoundry AI Gateway is the proxy layer that sits between your applications and the LLM providers and MCP Servers. It is an enterprise-grade platform that enables users to access 1000+ LLMs using a unified interface while taking care of observability and governance.

<img src="https://mintcdn.com/truefoundry/yOHcIuJN9I1uqUfe/images/new-ai-copy.png?fit=max&auto=format&n=yOHcIuJN9I1uqUfe&q=85&s=60d0d6629a9d197ac25041b8f708a80b" alt="TrueFoundry AI Gateway architecture diagram showing the gateway as a proxy between applications and multiple LLM providers" width="2518" height="1468" data-path="images/new-ai-copy.png" />

## Key Features

<CardGroup cols={3}>
  <Card title="Unified API Interface" icon="plug" href="/docs/ai-gateway/chat-completions-overview">
    Call **1000+ LLMs** using a single endpoint with unified API interface
  </Card>

  <Card title="API Keys Management" icon="key" href="/docs/ai-gateway/authentication">
    Generate and manage API keys for users/applications
  </Card>

  <Card title="Multimodal Inputs" icon="image" href="/docs/ai-gateway/chat-completions-overview#working-with-multi-modal">
    Support for text, image, and audio inputs across compatible models
  </Card>

  <Card title="Access Control" icon="lock" href="/docs/ai-gateway/gateway-access-control">
    Fine-grained access control and permissions management
  </Card>

  <Card title="Rate Limiting" icon="gauge-high" href="/docs/ai-gateway/ratelimiting">
    Control Models Usage with flexible rate limiting policies per user/model/application
  </Card>

  <Card title="Load balancing" icon="scale-balanced" href="/docs/ai-gateway/virtual-model">
    Use virtual models to spread traffic across targets by weight, latency, or priority, with retries and fallbacks.
  </Card>

  <Card title="Budget Limiting" icon="wallet" href="/docs/ai-gateway/budgetlimiting">
    Control spending and enforce cost limits for users, teams, and models
  </Card>

  <Card title="Guardrails" icon="shield" href="/docs/ai-gateway/guardrails-overview">
    Content filtering and safety checks to ensure
  </Card>

  <Card title="Observability & Metrics" icon="chart-line" href="/docs/ai-gateway/analytics">
    Opentelemetry compliant metrics and logging for all requests.
  </Card>

  <Card title="Prompt Playground" icon="message" href="/docs/ai-gateway/prompt-management">
    Centralized prompt playground with versioning and management system
  </Card>

  <Card title="Batch Predictions" icon="layer-group" href="/docs/ai-gateway/batch-predictions-with-truefoundry-llm-gateway">
    Process multiple requests efficiently with batch processing
  </Card>

  <Card title="MCP Registry" icon="server" href="/docs/ai-gateway/mcp/mcp-server-getting-started">
    Deploy and manage your own MCP servers with TrueFoundry AI Gateway.
  </Card>

  <Card title="Centralized Authn/Authz for all MCP Servers" icon="key" href="/docs/ai-gateway/mcp/mcp-gateway-auth-security">
    One API key to access all MCP servers and their tools.
  </Card>

  <Card title="Virtual MCP Servers" icon="key" href="/docs/ai-gateway/mcp/virtual-mcp-server">
    Create virtual MCP servers combining specific tools from multiple MCP servers.
  </Card>

  <Card title="Agent Playground" icon="gauge-high" href="/docs/ai-gateway/agents/use-mcp-server-in-code-agent">
    Test Agents by adding tools and models from Playground
  </Card>

  <Card title="Build Agents with unified API for all MCP servers" icon="route-interstate" href="/docs/ai-gateway/mcp-gateway-sdk-usage">
    Connect to MCP Servers with a single API in the gateway.
  </Card>

  <Card title="Rate Limiting and Observability for Tools" icon="route-interstate" href="/docs/ai-gateway/mcp-gateway-sdk-usage">
    Coming Soon
  </Card>
</CardGroup>

## Supported Model Providers

We integrate with 1000+ LLMs through the following providers.

<Tip>
  If you don't see the provider you need, there is a high change it will just work as self hosted models or OpenAI provider. Please reach out to us at [support@truefoundry.com](mailto:support@truefoundry.com) and we will be happy to guide you.
</Tip>

<CardGroup cols={4}>
  <Card className="!py-2 !px-3" href="/docs/ai-gateway/google-vertex">
    <div className="flex items-center gap-1.5 w-full">
      <img src="https://mintcdn.com/truefoundry/FrY4JbiyZud2He3p/images/0530d732-852cf2ae041019e3322649a2731ef87fe038bb09ce2271e7114f4fb2ed9da415-file.png?fit=max&auto=format&n=FrY4JbiyZud2He3p&q=85&s=99bf56cda35ee8fe027ec9ccbef71a6f" width="24" height="24" style={{ maxWidth:"24px",maxHeight:"24px",objectFit:"contain" }} data-path="images/0530d732-852cf2ae041019e3322649a2731ef87fe038bb09ce2271e7114f4fb2ed9da415-file.png" />

      <span className="text-sm font-semibold leading-tight">
        Gemini & Vertex AI
      </span>
    </div>
  </Card>

  <Card className="!py-2 !px-3" href="/docs/ai-gateway/google-gemini">
    <div className="flex items-center gap-1.5 w-full">
      <img src="https://mintcdn.com/truefoundry/gonlEy8fl_2fR_5Y/images/icons/gemini-logo.png?fit=max&auto=format&n=gonlEy8fl_2fR_5Y&q=85&s=21e5ae33d0b2b3ef879b326fd2bdff65" alt="Google Gemini logo" width="24" height="24" style={{ maxWidth:"24px",maxHeight:"24px",objectFit:"contain" }} data-path="images/icons/gemini-logo.png" />

      <span className="text-sm font-semibold leading-tight">
        Google Gemini
      </span>
    </div>
  </Card>

  <Card className="!py-2 !px-3" href="/docs/ai-gateway/aws-bedrock">
    <div className="flex items-center gap-1.5 w-full">
      <img src="https://mintcdn.com/truefoundry/yRoKH_fkKi2nPtuV/images/fa5bae4c-3d9a209788691caf01ac023fd388edcb5eb0bebddc7a6cbcfdfe5749d7167efa-file_1.png?fit=max&auto=format&n=yRoKH_fkKi2nPtuV&q=85&s=a4bee47a494a35f975d194c421b3a183" width="24" height="24" style={{ maxWidth:"24px",maxHeight:"24px",objectFit:"contain" }} data-path="images/fa5bae4c-3d9a209788691caf01ac023fd388edcb5eb0bebddc7a6cbcfdfe5749d7167efa-file_1.png" />

      <span className="text-sm font-semibold leading-tight">
        AWS Bedrock
      </span>
    </div>
  </Card>

  <Card className="!py-2 !px-3" href="/docs/ai-gateway/aws-sagemaker">
    <div className="flex items-center gap-1.5 w-full">
      <img src="https://assets.production.truefoundry.com/aws-sagemaker.svg" alt="AWS SageMaker logo" width="24" height="24" style={{ maxWidth:"24px",maxHeight:"24px",objectFit:"contain" }} />

      <span className="text-sm font-semibold leading-tight">
        AWS SageMaker
      </span>
    </div>
  </Card>

  <Card className="!py-2 !px-3" href="/docs/ai-gateway/azure-openai">
    <div className="flex items-center gap-1.5 w-full">
      <img src="https://mintcdn.com/truefoundry/DdP_2rhue4AQQlob/images/4ba5199c-3e375f637423c83621df291e00c7fcd27ed05a0d3a5b90794e6a85d4702662a1-file_3.jpg?fit=max&auto=format&n=DdP_2rhue4AQQlob&q=85&s=4b0067ff5ecf8cf568e4c551fc602fbb" width="24" height="24" style={{ maxWidth:"24px",maxHeight:"24px",objectFit:"contain" }} data-path="images/4ba5199c-3e375f637423c83621df291e00c7fcd27ed05a0d3a5b90794e6a85d4702662a1-file_3.jpg" />

      <span className="text-sm font-semibold leading-tight">
        Azure OpenAI
      </span>
    </div>
  </Card>

  <Card className="!py-2 !px-3" href="/docs/ai-gateway/azure-ai-foundry">
    <div className="flex items-center gap-1.5 w-full">
      <img src="https://mintcdn.com/truefoundry/DdP_2rhue4AQQlob/images/4ba5199c-3e375f637423c83621df291e00c7fcd27ed05a0d3a5b90794e6a85d4702662a1-file_3.jpg?fit=max&auto=format&n=DdP_2rhue4AQQlob&q=85&s=4b0067ff5ecf8cf568e4c551fc602fbb" width="24" height="24" style={{ maxWidth:"24px",maxHeight:"24px",objectFit:"contain" }} data-path="images/4ba5199c-3e375f637423c83621df291e00c7fcd27ed05a0d3a5b90794e6a85d4702662a1-file_3.jpg" />

      <span className="text-sm font-semibold leading-tight">
        Azure AI Foundry
      </span>
    </div>
  </Card>

  <Card className="!py-2 !px-3" href="/docs/ai-gateway/openai">
    <div className="flex items-center gap-1.5 w-full">
      <img src="https://mintcdn.com/truefoundry/yRoKH_fkKi2nPtuV/images/f863f110-89598b0f5a65b24a53d4e81077408fb7201ec861bfd2aace4665afc6a7e46d24-file_14.png?fit=max&auto=format&n=yRoKH_fkKi2nPtuV&q=85&s=8a9a714cb2c486f252e95948b5a5e5fb" alt="OpenAI logo" width="24" height="24" style={{ maxWidth:"24px",maxHeight:"24px",objectFit:"contain" }} data-path="images/f863f110-89598b0f5a65b24a53d4e81077408fb7201ec861bfd2aace4665afc6a7e46d24-file_14.png" />

      <span className="text-sm font-semibold leading-tight">
        OpenAI
      </span>
    </div>
  </Card>

  <Card className="!py-2 !px-3" href="/docs/ai-gateway/cohere">
    <div className="flex items-center gap-1.5 w-full">
      <img src="https://mintcdn.com/truefoundry/s4Aj2_qGCrSP-zc8/images/889ac4a7-fe5113cbdecff51ea30fd7a2f9de3dabc76a656274bdb5cf789bb0829dc1226f-file_2.png?fit=max&auto=format&n=s4Aj2_qGCrSP-zc8&q=85&s=c16d6313f920d64c5c46be5db297a299" width="24" height="24" style={{ maxWidth:"24px",maxHeight:"24px",objectFit:"contain" }} data-path="images/889ac4a7-fe5113cbdecff51ea30fd7a2f9de3dabc76a656274bdb5cf789bb0829dc1226f-file_2.png" />

      <span className="text-sm font-semibold leading-tight">
        Cohere
      </span>
    </div>
  </Card>

  <Card className="!py-2 !px-3" href="/docs/ai-gateway/databricks-models">
    <div className="flex items-center gap-1.5 w-full">
      <img src="https://mintcdn.com/truefoundry/_wMKf6ALFfIRnUSQ/images/databricks.svg?fit=max&auto=format&n=_wMKf6ALFfIRnUSQ&q=85&s=d1af0b862115159b3473e89df853c12d" width="24" height="24" style={{ maxWidth:"24px",maxHeight:"24px",objectFit:"contain" }} data-path="images/databricks.svg" />

      <span className="text-sm font-semibold leading-tight">
        Databricks
      </span>
    </div>
  </Card>

  <Card className="!py-2 !px-3" href="/docs/ai-gateway/ai21">
    <div className="flex items-center gap-1.5 w-full">
      <img src="https://mintcdn.com/truefoundry/DdP_2rhue4AQQlob/images/47429280-27551f097d32d959f0c4c61a42ba6bf529ce637f85658742edce85f0d5347d29-file_5.png?fit=max&auto=format&n=DdP_2rhue4AQQlob&q=85&s=7dec74b7b9d9eb403afb2dff8c795514" width="24" height="24" style={{ maxWidth:"24px",maxHeight:"24px",objectFit:"contain" }} data-path="images/47429280-27551f097d32d959f0c4c61a42ba6bf529ce637f85658742edce85f0d5347d29-file_5.png" />

      <span className="text-sm font-semibold leading-tight">
        AI21
      </span>
    </div>
  </Card>

  <Card className="!py-2 !px-3" href="/docs/ai-gateway/anthropic">
    <div className="flex items-center gap-1.5 w-full">
      <img src="https://mintcdn.com/truefoundry/4MAaF__cLD4iud16/images/5f73a51c-d876ed2262e9a83e2e08daaf7386c2d47fca02d2ac74d15166d5163c0f8cf386-file_6.png?fit=max&auto=format&n=4MAaF__cLD4iud16&q=85&s=455f52ce6f6237e43faa06531790cad5" width="24" height="24" style={{ maxWidth:"24px",maxHeight:"24px",objectFit:"contain" }} data-path="images/5f73a51c-d876ed2262e9a83e2e08daaf7386c2d47fca02d2ac74d15166d5163c0f8cf386-file_6.png" />

      <span className="text-sm font-semibold leading-tight">
        Anthropic
      </span>
    </div>
  </Card>

  <Card className="!py-2 !px-3" href="/docs/ai-gateway/together-ai">
    <div className="flex items-center gap-1.5 w-full">
      <img src="https://mintcdn.com/truefoundry/OHzlp6GY5G-JfKle/images/d3067631-a4374539855336dc486639bd82d7440b841c19f537d15e65e0195d5ab4c912a4-file_13.png?fit=max&auto=format&n=OHzlp6GY5G-JfKle&q=85&s=394b88f2bb9609d8f75dceee3251f961" width="24" height="24" style={{ maxWidth:"24px",maxHeight:"24px",objectFit:"contain" }} data-path="images/d3067631-a4374539855336dc486639bd82d7440b841c19f537d15e65e0195d5ab4c912a4-file_13.png" />

      <span className="text-sm font-semibold leading-tight">
        Together AI
      </span>
    </div>
  </Card>

  <Card className="!py-2 !px-3" href="/docs/ai-gateway/xai">
    <div className="flex items-center gap-1.5 w-full">
      <img src="https://mintcdn.com/truefoundry/1XZGPKeV4QNYX7dI/images/xai.png?fit=max&auto=format&n=1XZGPKeV4QNYX7dI&q=85&s=36c5e1282975cd58476a589d54c23c0c" width="24" height="24" style={{ maxWidth:"24px",maxHeight:"24px",objectFit:"contain" }} data-path="images/xai.png" />

      <span className="text-sm font-semibold leading-tight">
        xAI
      </span>
    </div>
  </Card>

  <Card className="!py-2 !px-3" href="/docs/ai-gateway/deepinfra">
    <div className="flex items-center gap-1.5 w-full">
      <img src="https://mintcdn.com/truefoundry/4MAaF__cLD4iud16/images/59c01ead-636bf9baad4503876e179f1047874d50dccb083a8c4e8c385e76d7c360c2f1e6-file_3.png?fit=max&auto=format&n=4MAaF__cLD4iud16&q=85&s=03ceff9e1d841caf2e9b12b00630c437" width="24" height="24" style={{ maxWidth:"24px",maxHeight:"24px",objectFit:"contain" }} data-path="images/59c01ead-636bf9baad4503876e179f1047874d50dccb083a8c4e8c385e76d7c360c2f1e6-file_3.png" />

      <span className="text-sm font-semibold leading-tight">
        DeepInfra
      </span>
    </div>
  </Card>

  <Card className="!py-2 !px-3" href="/docs/ai-gateway/perplexity-ai">
    <div className="flex items-center gap-1.5 w-full">
      <img src="https://mintcdn.com/truefoundry/s4Aj2_qGCrSP-zc8/images/8d4ef381-9e235c1c05e84fed849a0c378f60c9a9ed408d950c2c4fd016ca74c4b7ef2fd1-file_12.png?fit=max&auto=format&n=s4Aj2_qGCrSP-zc8&q=85&s=87f18eb81cd7bc56f9b1be95ca6f5d88" width="24" height="24" style={{ maxWidth:"24px",maxHeight:"24px",objectFit:"contain" }} data-path="images/8d4ef381-9e235c1c05e84fed849a0c378f60c9a9ed408d950c2c4fd016ca74c4b7ef2fd1-file_12.png" />

      <span className="text-sm font-semibold leading-tight">
        Perplexity AI
      </span>
    </div>
  </Card>

  <Card className="!py-2 !px-3" href="/docs/ai-gateway/mistral">
    <div className="flex items-center gap-1.5 w-full">
      <img src="https://mintcdn.com/truefoundry/4MAaF__cLD4iud16/images/504a8e0a-2f6ad398a7cae8352f0ac97224304abfb07789c9405369e68f07494a259d6340-file_7.png?fit=max&auto=format&n=4MAaF__cLD4iud16&q=85&s=e61c0c8a7b762af5f5b90d416148a884" width="24" height="24" style={{ maxWidth:"24px",maxHeight:"24px",objectFit:"contain" }} data-path="images/504a8e0a-2f6ad398a7cae8352f0ac97224304abfb07789c9405369e68f07494a259d6340-file_7.png" />

      <span className="text-sm font-semibold leading-tight">
        Mistral AI
      </span>
    </div>
  </Card>

  <Card className="!py-2 !px-3" href="/docs/ai-gateway/cloudera">
    <div className="flex items-center gap-1.5 w-full">
      <img src="https://assets.production.truefoundry.com/cloudera.svg" alt="Cloudera logo" width="24" height="24" style={{ maxWidth:"24px",maxHeight:"24px",objectFit:"contain" }} />

      <span className="text-sm font-semibold leading-tight">
        Cloudera
      </span>
    </div>
  </Card>

  <Card className="!py-2 !px-3" href="/docs/ai-gateway/groq">
    <div className="flex items-center gap-1.5 w-full">
      <img src="https://mintcdn.com/truefoundry/FrY4JbiyZud2He3p/images/115edf07-91692a7eb9b4b5670fd7bb682b0870a098663889cafa4475df06496d3f8073ac-file_4.png?fit=max&auto=format&n=FrY4JbiyZud2He3p&q=85&s=58e9a3c2e65e5784c396e18b0afcc464" width="24" height="24" style={{ maxWidth:"24px",maxHeight:"24px",objectFit:"contain" }} data-path="images/115edf07-91692a7eb9b4b5670fd7bb682b0870a098663889cafa4475df06496d3f8073ac-file_4.png" />

      <span className="text-sm font-semibold leading-tight">
        Groq
      </span>
    </div>
  </Card>

  <Card className="!py-2 !px-3" href="/docs/ai-gateway/elevenlabs">
    <div className="flex items-center gap-1.5 w-full">
      <img src="https://assets.production.truefoundry.com/elevenlabs.svg" alt="ElevenLabs logo" width="24" height="24" style={{ maxWidth:"24px",maxHeight:"24px",objectFit:"contain" }} />

      <span className="text-sm font-semibold leading-tight">
        ElevenLabs
      </span>
    </div>
  </Card>

  <Card className="!py-2 !px-3" href="/docs/ai-gateway/deepgram">
    <div className="flex items-center gap-1.5 w-full">
      <img src="https://assets.production.truefoundry.com/deepgram.svg" alt="Deepgram logo" width="24" height="24" style={{ maxWidth:"24px",maxHeight:"24px",objectFit:"contain" }} />

      <span className="text-sm font-semibold leading-tight">
        Deepgram
      </span>
    </div>
  </Card>

  <Card className="!py-2 !px-3" href="/docs/ai-gateway/cartesia">
    <div className="flex items-center gap-1.5 w-full">
      <img src="https://assets.production.truefoundry.com/cartesia.svg" alt="Cartesia logo" width="24" height="24" style={{ maxWidth:"24px",maxHeight:"24px",objectFit:"contain" }} />

      <span className="text-sm font-semibold leading-tight">
        Cartesia
      </span>
    </div>
  </Card>

  <Card className="!py-2 !px-3" href="/docs/ai-gateway/snowflake-cortex">
    <div className="flex items-center gap-1.5 w-full">
      <img src="https://assets.production.truefoundry.com/snowflake-cortex.svg" alt="Snowflake Cortex logo" width="24" height="24" style={{ maxWidth:"24px",maxHeight:"24px",objectFit:"contain" }} />

      <span className="text-sm font-semibold leading-tight">
        Snowflake Cortex
      </span>
    </div>
  </Card>

  <Card className="!py-2 !px-3" href="/docs/ai-gateway/self-hosted-models">
    <div className="flex items-center gap-1.5 w-full">
      <img src="https://mintcdn.com/truefoundry/6bbtTt95c1iCcyph/favicon.svg?fit=max&auto=format&n=6bbtTt95c1iCcyph&q=85&s=f55bde34214b38034126d05b08ea16c7" width="24" height="24" style={{ maxWidth:"24px",maxHeight:"24px",objectFit:"contain" }} data-path="favicon.svg" />

      <span className="text-sm font-semibold leading-tight">
        Self Hosted
      </span>
    </div>
  </Card>

  <Card className="!py-2 !px-3" href="/docs/ai-gateway/openrouter">
    <div className="flex items-center gap-1.5 w-full">
      <img src="https://assets.production.truefoundry.com/openrouter.svg" width="24" height="24" style={{ maxWidth:"24px",maxHeight:"24px",objectFit:"contain" }} />

      <span className="text-sm font-semibold leading-tight">
        OpenRouter
      </span>
    </div>
  </Card>

  <Card className="!py-2 !px-3" href="/docs/ai-gateway/sambanova">
    <div className="flex items-center gap-1.5 w-full">
      <img src="https://mintcdn.com/truefoundry/6SYERHht8PcdzWk0/images/SambaNova%20Favicon.webp?fit=max&auto=format&n=6SYERHht8PcdzWk0&q=85&s=36fffc3387a8312ea9958e206b2a7e48" width="24" height="24" style={{ maxWidth:"24px",maxHeight:"24px",objectFit:"contain" }} data-path="images/SambaNova Favicon.webp" />

      <span className="text-sm font-semibold leading-tight">
        SambaNova
      </span>
    </div>
  </Card>

  <Card className="!py-2 !px-3" href="/docs/ai-gateway/cerebras">
    <div className="flex items-center gap-1.5 w-full">
      <img src="https://mintcdn.com/truefoundry/qsnIjfIPg_BZ5d3X/images/cerebras-logo.jpeg?fit=max&auto=format&n=qsnIjfIPg_BZ5d3X&q=85&s=07d0b1589ba23c45c06695ca98319176" width="24" height="24" style={{ maxWidth:"24px",maxHeight:"24px",objectFit:"contain" }} data-path="images/cerebras-logo.jpeg" />

      <span className="text-sm font-semibold leading-tight">
        Cerebras
      </span>
    </div>
  </Card>
</CardGroup>

## Supported APIs

The following accordions summarize provider support for each gateway endpoint. Each section links to the full guide for that API (same order as **Supported APIs** in the sidebar).

<Info>
  Legend:

  * **✅** Supported by Provider and Truefoundry
  * <Icon icon="circle-xmark" iconType="regular" color="red" /> Provided by provider, but not by Truefoundry
  * <Icon icon="circle-minus" iconType="regular" /> Provider does not support this feature
</Info>

<AccordionGroup>
  <Accordion title="Chat Completion (/chat/completions)" defaultOpen="true">
    **Documentation:** [Chat Completions API](/docs/ai-gateway/chat-completions-overview)

    | Provider      | Stream | Non Stream | Tools                                           | JSON Mode                                       | Schema Mode                                     | Prompt Caching                                  | Reasoning                                       | Structured Output                               |
    | ------------- | ------ | ---------- | ----------------------------------------------- | ----------------------------------------------- | ----------------------------------------------- | ----------------------------------------------- | ----------------------------------------------- | ----------------------------------------------- |
    | OpenAI        | ✅      | ✅          | ✅                                               | ✅                                               | ✅                                               | ✅                                               | ✅                                               | <Icon icon="circle-minus" iconType="regular" /> |
    | Azure OpenAI  | ✅      | ✅          | ✅                                               | ✅                                               | ✅                                               | ✅                                               | ✅                                               | <Icon icon="circle-minus" iconType="regular" /> |
    | Anthropic     | ✅      | ✅          | ✅                                               | <Icon icon="circle-minus" iconType="regular" /> | ✅                                               | ✅                                               | ✅                                               | <Icon icon="circle-minus" iconType="regular" /> |
    | Bedrock       | ✅      | ✅          | ✅                                               | <Icon icon="circle-minus" iconType="regular" /> | ✅                                               | ✅                                               | ✅                                               | <Icon icon="circle-minus" iconType="regular" /> |
    | Vertex        | ✅      | ✅          | ✅                                               | <Icon icon="circle-minus" iconType="regular" /> | ✅                                               | ✅                                               | ✅                                               | <Icon icon="circle-minus" iconType="regular" /> |
    | Cohere        | ✅      | ✅          | ✅                                               | ✅                                               | ✅                                               | <Icon icon="circle-minus" iconType="regular" /> | ✅                                               | <Icon icon="circle-minus" iconType="regular" /> |
    | Gemini        | ✅      | ✅          | ✅                                               | ✅                                               | ✅                                               | ✅                                               | ✅                                               | <Icon icon="circle-minus" iconType="regular" /> |
    | Groq          | ✅      | ✅          | ✅                                               | ✅                                               | ✅                                               | ✅                                               | ✅                                               | <Icon icon="circle-minus" iconType="regular" /> |
    | AI21          | ✅      | ✅          | <Icon icon="circle-minus" iconType="regular" /> | ✅                                               | <Icon icon="circle-minus" iconType="regular" /> | <Icon icon="circle-minus" iconType="regular" /> | <Icon icon="circle-minus" iconType="regular" /> | <Icon icon="circle-minus" iconType="regular" /> |
    | Cerebras      | ✅      | ✅          | <Icon icon="circle-minus" iconType="regular" /> | ✅                                               | <Icon icon="circle-minus" iconType="regular" /> | <Icon icon="circle-minus" iconType="regular" /> | ✅                                               | <Icon icon="circle-minus" iconType="regular" /> |
    | SambaNova     | ✅      | ✅          | <Icon icon="circle-minus" iconType="regular" /> | ✅                                               | <Icon icon="circle-minus" iconType="regular" /> | <Icon icon="circle-minus" iconType="regular" /> | ✅                                               | <Icon icon="circle-minus" iconType="regular" /> |
    | Perplexity-AI | ✅      | ✅          | <Icon icon="circle-minus" iconType="regular" /> | ✅                                               | <Icon icon="circle-minus" iconType="regular" /> | <Icon icon="circle-minus" iconType="regular" /> | ✅                                               | ✅                                               |
    | Together-AI   | ✅      | ✅          | ✅                                               | ✅                                               | <Icon icon="circle-minus" iconType="regular" /> | ✅                                               | ✅                                               | ✅                                               |
    | xAI           | ✅      | ✅          | ✅                                               | ✅                                               | ✅                                               | ✅                                               | ✅                                               | ✅                                               |
    | DeepInfra     | ✅      | ✅          | ✅                                               | ✅                                               | <Icon icon="circle-minus" iconType="regular" /> | ✅                                               | ✅                                               | <Icon icon="circle-minus" iconType="regular" /> |
  </Accordion>

  <Accordion title="Embedding (/embeddings)">
    **Documentation:** [Embeddings API](/docs/ai-gateway/embed)

    | Provider     | String                                                      | List of String                                              |
    | ------------ | ----------------------------------------------------------- | ----------------------------------------------------------- |
    | OpenAI       | ✅                                                           | ✅                                                           |
    | Azure OpenAI | ✅                                                           | ✅                                                           |
    | Anthropic    | <Icon icon="circle-minus" iconType="regular" />             | <Icon icon="circle-minus" iconType="regular" />             |
    | Bedrock      | ✅                                                           | ✅                                                           |
    | Vertex       | ✅                                                           | ✅                                                           |
    | Cohere       | ✅                                                           | ✅                                                           |
    | Gemini       | <Icon icon="circle-minus" iconType="regular" />             | <Icon icon="circle-minus" iconType="regular" />             |
    | Groq         | <Icon icon="circle-minus" iconType="regular" />             | <Icon icon="circle-minus" iconType="regular" />             |
    | SambaNova    | <Icon icon="circle-xmark" iconType="regular" color="red" /> | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
    | Together-AI  | ✅                                                           | ✅                                                           |
    | xAI          | <Icon icon="circle-minus" iconType="regular" />             | <Icon icon="circle-minus" iconType="regular" />             |
    | DeepInfra    | <Icon icon="circle-xmark" iconType="regular" color="red" /> | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
  </Accordion>

  <Accordion title="Batch (/batches)">
    **Documentation:** [Batch API](/docs/ai-gateway/batch-predictions-with-truefoundry-llm-gateway)

    | Provider     | Batch                                                       |
    | ------------ | ----------------------------------------------------------- |
    | OpenAI       | ✅                                                           |
    | Azure OpenAI | ✅                                                           |
    | Anthropic    | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
    | Bedrock      | ✅                                                           |
    | Vertex       | ✅                                                           |
    | Cohere       | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
    | Gemini       | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
    | Groq         | <Icon icon="circle-minus" iconType="regular" />             |
    | Cerebras     | <Icon icon="circle-minus" iconType="regular" />             |
    | Together-AI  | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
    | xAI          | <Icon icon="circle-minus" iconType="regular" />             |
    | DeepInfra    | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
  </Accordion>

  <Accordion title="Fine Tune">
    **Documentation:** [Finetune API](/docs/ai-gateway/finetune)

    | Provider     | Fine Tune                                                   |
    | ------------ | ----------------------------------------------------------- |
    | OpenAI       | ✅                                                           |
    | Azure OpenAI | <Icon icon="circle-minus" iconType="regular" />             |
    | Anthropic    | <Icon icon="circle-minus" iconType="regular" />             |
    | Bedrock      | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
    | Vertex       | ✅                                                           |
    | Cohere       | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
    | Gemini       | <Icon icon="circle-minus" iconType="regular" />             |
    | Groq         | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
    | Cerebras     | <Icon icon="circle-minus" iconType="regular" />             |
    | Together-AI  | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
    | xAI          | <Icon icon="circle-minus" iconType="regular" />             |
    | DeepInfra    | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
  </Accordion>

  <Accordion title="Model Response (/responses)">
    **Documentation:** [Responses API](/docs/ai-gateway/responses-api)

    | Provider     | Model Response                                              |
    | ------------ | ----------------------------------------------------------- |
    | OpenAI       | ✅                                                           |
    | Azure OpenAI | ✅                                                           |
    | Anthropic    | <Icon icon="circle-minus" iconType="regular" />             |
    | Bedrock      | <Icon icon="circle-minus" iconType="regular" />             |
    | Vertex       | <Icon icon="circle-minus" iconType="regular" />             |
    | Cohere       | <Icon icon="circle-minus" iconType="regular" />             |
    | Gemini       | <Icon icon="circle-minus" iconType="regular" />             |
    | Groq         | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
    | Cerebras     | <Icon icon="circle-minus" iconType="regular" />             |
    | Together-AI  | <Icon icon="circle-minus" iconType="regular" />             |
    | xAI          | <Icon icon="circle-minus" iconType="regular" />             |
    | DeepInfra    | <Icon icon="circle-minus" iconType="regular" />             |
  </Accordion>

  <Accordion title="Image Generation (/images/generations)">
    **Documentation:** [Image Generation API](/docs/ai-gateway/image-generation)

    | Provider     | Generate                                                    |
    | ------------ | ----------------------------------------------------------- |
    | OpenAI       | ✅                                                           |
    | Azure OpenAI | ✅                                                           |
    | Bedrock      | ✅                                                           |
    | Vertex       | ✅                                                           |
    | Anthropic    | <Icon icon="circle-minus" iconType="regular" />             |
    | Cohere       | <Icon icon="circle-minus" iconType="regular" />             |
    | Gemini       | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
    | Groq         | <Icon icon="circle-minus" iconType="regular" />             |
    | Together-AI  | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
    | xAI          | <Icon icon="circle-minus" iconType="regular" />             |
    | DeepInfra    | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
  </Accordion>

  <Accordion title="Image Edit (/images/edits)">
    **Documentation:** [Image Edit API](/docs/ai-gateway/image-edit)

    | Provider     | Edit                                                        |
    | ------------ | ----------------------------------------------------------- |
    | OpenAI       | ✅                                                           |
    | Azure OpenAI | ✅                                                           |
    | Bedrock      | ✅                                                           |
    | Vertex       | ✅                                                           |
    | Anthropic    | <Icon icon="circle-minus" iconType="regular" />             |
    | Cohere       | <Icon icon="circle-minus" iconType="regular" />             |
    | Gemini       | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
    | Groq         | <Icon icon="circle-minus" iconType="regular" />             |
    | Together-AI  | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
    | xAI          | <Icon icon="circle-minus" iconType="regular" />             |
    | DeepInfra    | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
  </Accordion>

  <Accordion title="Image Variation (/images/variations)">
    **Documentation:** [Image Variation API](/docs/ai-gateway/image-variation)

    | Provider     | Variation                                                   |
    | ------------ | ----------------------------------------------------------- |
    | OpenAI       | ✅                                                           |
    | Azure OpenAI | <Icon icon="circle-minus" iconType="regular" />             |
    | Bedrock      | ✅                                                           |
    | Vertex       | <Icon icon="circle-minus" iconType="regular" />             |
    | Anthropic    | <Icon icon="circle-minus" iconType="regular" />             |
    | Cohere       | <Icon icon="circle-minus" iconType="regular" />             |
    | Gemini       | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
    | Groq         | <Icon icon="circle-minus" iconType="regular" />             |
    | Together-AI  | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
    | xAI          | <Icon icon="circle-minus" iconType="regular" />             |
    | DeepInfra    | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
  </Accordion>

  <Accordion title="Text To Speech">
    **Documentation:** [Text to Speech API](/docs/ai-gateway/text-to-speech)

    | Provider         | Text To Speech                                              |
    | ---------------- | ----------------------------------------------------------- |
    | OpenAI           | ✅                                                           |
    | Azure OpenAI     | ✅                                                           |
    | Azure AI Foundry | ✅                                                           |
    | Anthropic        | <Icon icon="circle-minus" iconType="regular" />             |
    | Bedrock          | <Icon icon="circle-minus" iconType="regular" />             |
    | Vertex           | ✅                                                           |
    | Cohere           | <Icon icon="circle-minus" iconType="regular" />             |
    | Gemini           | ✅                                                           |
    | Groq             | ✅                                                           |
    | Together-AI      | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
    | xAI              | <Icon icon="circle-minus" iconType="regular" />             |
    | DeepInfra        | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
    | DeepGram         | ✅                                                           |
    | Cartesia         | ✅                                                           |
    | ElevenLabs       | ✅                                                           |
  </Accordion>

  <Accordion title="Audio Translation">
    **Documentation:** [Audio Translation API](/docs/ai-gateway/audio-translation)

    | Provider         | Translation                                                 |
    | ---------------- | ----------------------------------------------------------- |
    | OpenAI           | ✅                                                           |
    | Azure OpenAI     | ✅                                                           |
    | Azure AI Foundry | ✅                                                           |
    | Anthropic        | <Icon icon="circle-minus" iconType="regular" />             |
    | Bedrock          | <Icon icon="circle-minus" iconType="regular" />             |
    | Vertex           | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
    | Cohere           | <Icon icon="circle-minus" iconType="regular" />             |
    | Gemini           | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
    | Groq             | ✅                                                           |
    | Together-AI      | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
    | xAI              | <Icon icon="circle-minus" iconType="regular" />             |
    | DeepInfra        | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
  </Accordion>

  <Accordion title="Speech to Text">
    **Documentation:** [Speech to Text API](/docs/ai-gateway/audio-transcription)

    | Provider         | Transcription                                               |
    | ---------------- | ----------------------------------------------------------- |
    | OpenAI           | ✅                                                           |
    | Azure OpenAI     | ✅                                                           |
    | Azure AI Foundry | ✅                                                           |
    | Anthropic        | <Icon icon="circle-minus" iconType="regular" />             |
    | Bedrock          | <Icon icon="circle-minus" iconType="regular" />             |
    | Vertex           | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
    | Cohere           | <Icon icon="circle-minus" iconType="regular" />             |
    | Gemini           | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
    | Groq             | ✅                                                           |
    | Together-AI      | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
    | xAI              | <Icon icon="circle-minus" iconType="regular" />             |
    | DeepInfra        | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
    | DeepGram         | ✅                                                           |
    | Cartesia         | ✅                                                           |
    | ElevenLabs       | ✅                                                           |
  </Accordion>

  <Accordion title="Live / Realtime API">
    **Documentation:** [Live / Realtime API](/docs/ai-gateway/live-api)

    | Provider         | Live / Realtime API |
    | ---------------- | ------------------- |
    | Gemini           | ✅                   |
    | Vertex           | ✅                   |
    | OpenAI           | ✅                   |
    | Azure AI Foundry | ✅                   |
  </Accordion>

  <Accordion title="Files (/files)">
    **Documentation:** [Files API](/docs/ai-gateway/file-endpoints)

    | Provider     | Files                                                       |
    | ------------ | ----------------------------------------------------------- |
    | OpenAI       | ✅                                                           |
    | Azure OpenAI | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
    | Anthropic    | ✅                                                           |
    | Bedrock      | ✅                                                           |
    | Vertex       | ✅                                                           |
    | Cohere       | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
    | Gemini       | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
    | Groq         | ✅                                                           |
    | Cerebras     | <Icon icon="circle-minus" iconType="regular" />             |
    | Together-AI  | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
    | xAI          | <Icon icon="circle-minus" iconType="regular" />             |
    | DeepInfra    | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
  </Accordion>

  <Accordion title="Rerank (/rerank)">
    **Documentation:** [Rerank API](/docs/ai-gateway/rerank)

    | Provider     | Rerank                                                      |
    | ------------ | ----------------------------------------------------------- |
    | OpenAI       | <Icon icon="circle-minus" iconType="regular" />             |
    | Azure OpenAI | <Icon icon="circle-minus" iconType="regular" />             |
    | Anthropic    | <Icon icon="circle-minus" iconType="regular" />             |
    | Bedrock      | ✅                                                           |
    | Vertex       | <Icon icon="circle-minus" iconType="regular" />             |
    | Cohere       | ✅                                                           |
    | Gemini       | <Icon icon="circle-minus" iconType="regular" />             |
    | Groq         | <Icon icon="circle-minus" iconType="regular" />             |
    | Together-AI  | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
    | xAI          | <Icon icon="circle-minus" iconType="regular" />             |
    | DeepInfra    | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
  </Accordion>

  <Accordion title="Moderation (/moderations)">
    **Documentation:** [Moderation API](/docs/ai-gateway/moderation)

    | Provider     | Moderation                                                  |
    | ------------ | ----------------------------------------------------------- |
    | OpenAI       | ✅                                                           |
    | Azure OpenAI | <Icon icon="circle-minus" iconType="regular" />             |
    | Anthropic    | <Icon icon="circle-minus" iconType="regular" />             |
    | Bedrock      | <Icon icon="circle-minus" iconType="regular" />             |
    | Vertex       | <Icon icon="circle-minus" iconType="regular" />             |
    | Cohere       | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
    | Gemini       | <Icon icon="circle-minus" iconType="regular" />             |
    | Groq         | <Icon icon="circle-minus" iconType="regular" />             |
    | Cerebras     | <Icon icon="circle-minus" iconType="regular" />             |
    | Together-AI  | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
    | xAI          | <Icon icon="circle-minus" iconType="regular" />             |
    | DeepInfra    | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
  </Accordion>

  <Accordion title="Compaction API">
    **Documentation:** [Compaction API](/docs/ai-gateway/compaction)

    | Provider | Compaction API |
    | -------- | -------------- |
    | OpenAI   | ✅              |
  </Accordion>

  <Accordion title="Messages API">
    **Documentation:** [Messages API](/docs/ai-gateway/messages-overview)

    | Provider  | Messages API |
    | --------- | ------------ |
    | Anthropic | ✅            |
  </Accordion>

  <Accordion title="Proxy API (/proxy)">
    **Documentation:** [Proxy API](/docs/ai-gateway/proxy-api)

    Forward provider-native requests through the gateway while keeping logging, rate limiting, and budget controls. See the guide for setup, headers, and examples by provider.
  </Accordion>
</AccordionGroup>

## Deployment Options

You can run the AI Gateway as fully managed SaaS, keep LLM request–response data in your own object storage while Truefoundry operates the gateway, or host the gateway plane (and optionally more of the stack) in your cloud or on-prem for stricter data residency and control. Each option differs in who hosts infrastructure, where traffic flows, and pricing tier.

Read the full comparison—including a scenario table, diagrams, and operational notes—in [AI Gateway deployment options](/docs/ai-gateway/modes-of-deployment). For background on how the gateway fits the platform, see [gateway plane architecture](/docs/platform/gateway-plane-architecture). To start on managed SaaS, follow the [quick start](/docs/ai-gateway/quick-start).

## Frequently Asked Questions

<AccordionGroup>
  <Accordion title="What's the performance impact of using the gateway?">
    The latency overhead is minimal, typically less than 5ms. Our benchmarks show enterprise-grade performance that scales with your needs. Our SaaS offering is hosted in multiple regions across the world to ensure low latency and high availability. You can also deploy the gateway on-premise or on any cloud provider in your region which \
    is closer to your users.

    <Frame caption="AI Gateway on the edge, close to your applications for optimal performance">
      <img src="https://mintcdn.com/truefoundry/FrY4JbiyZud2He3p/images/081e6057-686a04943a81ee773a3a94dcddff05647776f82f7a2f32d4155d4df0f61f0032-image.png?fit=max&auto=format&n=FrY4JbiyZud2He3p&q=85&s=c96dc61ac031c4c4dae8edf115f6f0f4" alt="" width="2438" height="1544" data-path="images/081e6057-686a04943a81ee773a3a94dcddff05647776f82f7a2f32d4155d4df0f61f0032-image.png" />
    </Frame>
  </Accordion>

  <Accordion title="Can I deploy the gateway on-premise?">
    Yes, the AI Gateway supports on-premise deployments on any infrastructure or cloud provider, giving you complete control over your AI operations.
  </Accordion>

  <Accordion title="How do I integrate my self-hosted models?">
    You can easily integrate any OpenAI-compatible self-hosted model. Check our [self-hosted models guide](/docs/ai-gateway/self-hosted-models) for detailed instructions.
  </Accordion>

  <Accordion title="Can I use the gateway without the full MLOps platform?">
    Yes, The AI Gateway can be used as a standalone solution. You can use the full MLOps platform if you're using features like model deployment(traditional models and LLMs), model training, llm fine-tuning or training/data-processing workflows.
  </Accordion>
</AccordionGroup>
