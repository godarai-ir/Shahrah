> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# AnythingLLM

> Learn how to connect AnythingLLM with the TrueFoundry AI Gateway.

This guide provides instructions for integrating [AnythingLLM](https://anythingllm.com/) with the Truefoundry AI Gateway.

## What is AnythingLLM?

AnythingLLM is an all-in-one Desktop & Docker AI application with built-in RAG, AI agents, and more. It provides a comprehensive platform for running AI applications locally or in the cloud with complete control over your data and models.

### Key Features of AnythingLLM

* **[Custom Model Support](https://docs.anythingllm.com/features/language-models)**: Stay fully local with built-in LLM provider running any model you want, or leverage enterprise models from OpenAI, Azure, AWS, and more with flexible model switching capabilities
* **[Universal Document Support](https://docs.anythingllm.com/chatting-with-documents/introduction)**: Use every type of document including PDFs, Word documents, CSV files, codebases, and much more. Import documents from online locations and make all your business data accessible
* **[Complete Privacy Control](https://docs.anythingllm.com/features/privacy-and-data-handling)**: AnythingLLM comes with sensible and locally running defaults for your LLM, embedder, vector database, and storage. Nothing is shared unless you explicitly allow it

## Prerequisites

Before integrating AnythingLLM with TrueFoundry, ensure you have:

1. **TrueFoundry Account**: Create a [Truefoundry account](https://www.truefoundry.com/register) and follow the instructions in our [Gateway Quick Start Guide](https://docs.truefoundry.com/gateway/quick-start)
2. **AnythingLLM Installation**: Set up AnythingLLM using either the [Desktop application](https://anythingllm.com/download) or [Docker deployment](https://github.com/Mintplex-Labs/anything-llm)

## Integration Steps

This guide assumes you have AnythingLLM installed and running, and have obtained your TrueFoundry AI Gateway base URL.

### Step 1: Access AnythingLLM LLM Settings

1. Launch your AnythingLLM application (Desktop or Docker).

2. Navigate to **Settings** and go to **LLM Preference**:

<Frame>
  <img src="https://mintcdn.com/truefoundry/5ZMNUF311tgGO_yT/images/llm%20provider%20.png?fit=max&auto=format&n=5ZMNUF311tgGO_yT&q=85&s=e98fe40c22451d5a1fd90e0ade7e3e42" alt="AnythingLLM settings page showing LLM provider selection interface" width="1440" height="900" data-path="images/llm provider .png" />
</Frame>

### Step 2: Configure Generic OpenAI Provider

1. In the LLM provider search box, type "Generic OpenAI" and select it from the available options.

2. Configure the TrueFoundry connection with the following settings:
   * **Base URL**: Enter your TrueFoundry Gateway base URL
   * **Chat Model Name**: Enter the model name from the unified code snippet (e.g., `openai-main/gpt-4o`)
   * **Token Context Window**: Set based on your model's limits (e.g., 16000, 128000)
   * **Max Tokens**: Configure according to your needs (e.g., 1024, 2048)

You will get your base URL and model name directly from the unified code snippet:

<Frame>
  <img src="https://mintcdn.com/truefoundry/n3EuZuJ0K8wBFp1G/images/new-code-snippet.png?fit=max&auto=format&n=n3EuZuJ0K8wBFp1G&q=85&s=3634c2dc8c3565fd77ab896d3fd07ed9" alt="TrueFoundry playground showing unified code snippet with base URL and model name" width="2940" height="1664" data-path="images/new-code-snippet.png" />
</Frame>

Copy the base URL and model ID and paste them into AnythingLLM's configuration fields.

### Step 3: Test Your Integration

1. Save your configuration in AnythingLLM.

2. Create a new workspace or open an existing one to test the integration:

<Frame>
  <img src="https://mintcdn.com/truefoundry/2MMcllD7kMlpnaWX/images/test%20anything%20llm.png?fit=max&auto=format&n=2MMcllD7kMlpnaWX&q=85&s=88b88694a653872e5323bb631e0ba780" alt="AnythingLLM chat interface showing successful test message with TrueFoundry integration" width="1920" height="804" data-path="images/test anything llm.png" />
</Frame>

3. Send a test message to verify that AnythingLLM is successfully communicating with TrueFoundry's AI Gateway.

Your AnythingLLM application is now integrated with TrueFoundry's AI Gateway and ready for AI chat, RAG, and agent operations.
