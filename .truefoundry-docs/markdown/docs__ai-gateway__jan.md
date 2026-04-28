> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Jan AI

> Learn how to use jan with TrueFoundry AI Gateway, including setup steps, use cases, and production-ready examples.

This guide provides instructions for integrating [Jan AI](https://jan.ai/) with the Truefoundry AI Gateway.

## What is Jan AI?

Jan is an open source ChatGPT-alternative that runs 100% offline on your computer. It's a desktop application designed for privacy-conscious users who want to chat with AI without sending their data to external servers. Jan can run powerful local models like Llama3, Gemma, and Mistral directly on your device, while also providing the flexibility to connect to cloud AI providers when needed.

### Key Features of Jan AI

* **[100% Offline Operation](https://jan.ai/)**: Run AI models locally on your computer without internet dependency, ensuring complete privacy and data security
* **[Model Hub](https://jan.ai/)**: Download and run powerful open-source models like Llama3, Gemma, and Mistral directly on your device
* **[Cloud AI Integration](https://jan.ai/)**: Connect to cloud models from providers like OpenAI, Groq, and Cohere when you need more computational power
* **[Local API Server](https://jan.ai/)**: Set up and run your own OpenAI-compatible API server using local models with just one click
* **[Chat with Files](https://jan.ai/)**: Talk to PDFs, notes, and other documents directly to get summaries, answers, or insights
* **[Fully Customizable](https://jan.ai/)**: Customize Jan with extensions and adjust alignment, moderation, and censorship levels to your needs

## Prerequisites

Before integrating Jan AI with TrueFoundry, ensure you have:

1. **TrueFoundry Account**: Create a [Truefoundry account](https://www.truefoundry.com/register) and follow our [Gateway Quick Start](https://docs.truefoundry.com/gateway/quick-start)
2. **Jan AI Installation**: Download and install Jan from [jan.ai](https://jan.ai/) (3.9M+ downloads, free & open source)

## Setup Process

### 1. Configure Jan AI with TrueFoundry

You will get your 'truefoundry-api-key', 'truefoundry-gateway-url' and model name directly from the unified code snippet

<Frame>
  <img src="https://mintcdn.com/truefoundry/n3EuZuJ0K8wBFp1G/images/new-code-snippet.png?fit=max&auto=format&n=n3EuZuJ0K8wBFp1G&q=85&s=3634c2dc8c3565fd77ab896d3fd07ed9" width="2940" height="1664" data-path="images/new-code-snippet.png" />
</Frame>

### 2. Configure OpenAI Provider in Jan Settings

1. **Open Jan AI Application**
2. **Navigate to Settings** → **Model Providers** → **OpenAI**

<Frame caption="Jan AI Settings - OpenAI Provider Configuration">
  <img src="https://mintcdn.com/truefoundry/5ZMNUF311tgGO_yT/images/jan-api-settings.png?fit=max&auto=format&n=5ZMNUF311tgGO_yT&q=85&s=90c8fc8dc3cab8ca356144a8f7def82b" alt="Jan AI settings showing OpenAI provider configuration" width="2272" height="1824" data-path="images/jan-api-settings.png" />
</Frame>

3. **Update the configuration**:
   * **API Key**: Enter your TrueFoundry API key
   * **Base URL**: Enter your TrueFoundry gateway URL (e.g., `{GATEWAY_BASE_URL}`)

### 3. Add TrueFoundry Models

1. **In the Models section**, click the **"+"** button to add new models
2. **Enter model IDs** in TrueFoundry format (e.g., `openai-main/gpt-4o`, `anthropic-main/claude-3-5-sonnet-20241022`)

<Frame caption="Add New Model Dialog">
  <img src="https://mintcdn.com/truefoundry/5ZMNUF311tgGO_yT/images/jan-add-model-name.png?fit=max&auto=format&n=5ZMNUF311tgGO_yT&q=85&s=d0d954771383b5fb28253b436b4ed7cd" alt="Jan AI add new model dialog showing TrueFoundry model format" width="2272" height="1824" data-path="images/jan-add-model-name.png" />
</Frame>

## Usage Examples

### Start Chatting with TrueFoundry Models

Once configured, you can start using TrueFoundry models in Jan's chat interface:

<Frame caption="Jan AI Chat Interface with TrueFoundry Models">
  <img src="https://mintcdn.com/truefoundry/5ZMNUF311tgGO_yT/images/jan-response.png?fit=max&auto=format&n=5ZMNUF311tgGO_yT&q=85&s=c07bec20e9e1cb89f57202df1954c92b" alt="Jan AI chat interface showing TrueFoundry models in use" width="2272" height="1824" data-path="images/jan-response.png" />
</Frame>

1. **Select Model**: Choose your configured TrueFoundry model from the model dropdown (e.g., `openai-main/gpt-4o`)
2. **Start Chatting**: Begin conversations with the AI - requests will be routed through TrueFoundry's gateway
3. **Switch Models**: Easily switch between local models and TrueFoundry cloud models as needed

### Observability and Governance

Monitor your Jan AI usage through TrueFoundry's metrics tab:

<img src="https://mintcdn.com/truefoundry/yRoKH_fkKi2nPtuV/images/gateway-metrics.png?fit=max&auto=format&n=yRoKH_fkKi2nPtuV&q=85&s=5a442952b4a398bcf6ab277d2392ca2c" alt="TrueFoundry metrics" width="3840" height="1984" data-path="images/gateway-metrics.png" />

With TrueFoundry's AI gateway, you can monitor and analyze:

* **Performance Metrics**: Track key latency metrics like Request Latency, Time to First Token (TTFS), and Inter-Token Latency (ITL) with P99, P90, and P50 percentiles
* **Cost and Token Usage**: Gain visibility into your application's costs with detailed breakdowns of input/output tokens and the associated expenses for each model
* **Usage Patterns**: Understand how your application is being used with detailed analytics on user activity, model distribution, and team-based usage
* **Hybrid Usage**: Monitor the balance between local model usage and cloud model requests through TrueFoundry
* **Rate Limiting and Virtual Models**: Set up rate limiting and configure [Virtual Models](/docs/ai-gateway/virtual-model) for intelligent routing and fallback across your models
