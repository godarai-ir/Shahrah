> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Langflow

> Learn how to use langflow with TrueFoundry AI Gateway, including setup steps, use cases, and production-ready examples.

This guide provides instructions for integrating [Langflow](https://www.langflow.org/) with the Truefoundry AI Gateway.

## What is Langflow?

Langflow is a visual framework for building multi-agent and RAG applications. It's a low-code app builder for RAG and multi-agent AI applications, providing a Python-based platform that allows users to create flows using a drag-and-drop interface or through code.

### Key Features of Langflow

1. **Visual Drag-and-Drop Interface**: Langflow provides an intuitive visual interface where users can build complex AI workflows by simply dragging and dropping components. This eliminates the need for extensive coding and makes AI application development accessible to non-technical users and developers alike.

2. **Multi-Agent and RAG Support**: Built-in support for Retrieval-Augmented Generation (RAG) and multi-agent architectures allows users to create sophisticated AI applications that can access external knowledge bases and coordinate multiple AI agents to solve complex tasks collaboratively.

3. **Code and No-Code Flexibility**: Langflow offers both visual workflow creation and [Python SDK](https://github.com/langflow-ai/langflow) integration, allowing users to switch between drag-and-drop interfaces and programmatic control. This flexibility enables both rapid prototyping and production-ready deployments with custom logic.

## Prerequisites

Before integrating Langflow with TrueFoundry, ensure you have:

1. **TrueFoundry Account**: Create a [Truefoundry account](https://www.truefoundry.com/register) and follow the instructions in our [Gateway Quick Start Guide](https://docs.truefoundry.com/gateway/quick-start)
2. **Langflow Installation**: Install Langflow using either the [Python package](https://github.com/langflow-ai/langflow) or [Docker deployment](https://docs.langflow.org/getting-started-installation)
3. **Virtual Model**: Create a Virtual Model for your desired models (see [Create a Virtual Model](#1-create-a-virtual-model) below)

## Why You Need a Virtual Model

Langflow works optimally with standard OpenAI model names (like `gpt-4`, `gpt-4o-mini`), but may experience compatibility issues with TrueFoundry's fully qualified model names (like `openai-main/gpt-4` or `azure-openai/gpt-4`).

When Langflow encounters these fully qualified model names directly, it may not function as expected due to internal processing differences.

**The Solution**: A Virtual Model allows you to:

1. Use standard model names in your Langflow configurations (e.g., `gpt-4`)
2. Have TrueFoundry Gateway automatically route the request to the fully qualified target model (e.g., `openai-main/gpt-4`)

This approach ensures seamless compatibility while still allowing you to access any model through the TrueFoundry Gateway.

## Setup Process

### 1. Create a Virtual Model

Create a Virtual Model so Langflow can use a simple model name that the Gateway maps to your provider:

1. Navigate to **AI Gateway → Models → Virtual Model** in the TrueFoundry dashboard.

<Frame>
  <img src="https://mintcdn.com/truefoundry/9sjb7wvpsYV7RE1W/images/Screenshot2026-02-27at4.38.04PM-1.png?fit=max&auto=format&n=9sjb7wvpsYV7RE1W&q=85&s=5815438d18a9ecd74e2c183b7aa8d9a1" alt="TrueFoundry Virtual Models dashboard" className="block dark:hidden" width="3598" height="2008" data-path="images/Screenshot2026-02-27at4.38.04PM-1.png" />

  <img src="https://mintcdn.com/truefoundry/uL7y0Q9ywygbGxPN/images/Screenshot2026-02-27at4.38.04PM.png?fit=max&auto=format&n=uL7y0Q9ywygbGxPN&q=85&s=84550e1c21cc38734bfbabe680a189f1" alt="TrueFoundry Virtual Models dashboard" className="hidden dark:block" width="3598" height="2008" data-path="images/Screenshot2026-02-27at4.38.04PM.png" />
</Frame>

2. Create a new Virtual Model Provider Group with a name (e.g., `langflow-vm`) and configure collaborators for access control.

3. Add your target model (e.g., `openai-main/gpt-4o`) under the provider group. Set the Virtual Model name to the model name Langflow expects (e.g., `gpt-4o`), so requests from Langflow are automatically routed to your configured provider.

For more details, see the [Virtual Models documentation](/docs/ai-gateway/virtual-model).

### 2. Configure Langflow Language Model Component

In your Langflow interface, configure the Language Model component with TrueFoundry Gateway settings:

<Frame>
  <img src="https://mintcdn.com/truefoundry/5ZMNUF311tgGO_yT/images/langflow-1.png?fit=max&auto=format&n=5ZMNUF311tgGO_yT&q=85&s=bbb7427be6e918d8be09d7a1c392af7e" alt="Langflow interface showing Language Model component configuration panel" width="2056" height="1328" data-path="images/langflow-1.png" />
</Frame>

You will get your base URL and model name directly from the unified code snippet:

<Frame>
  <img src="https://mintcdn.com/truefoundry/n3EuZuJ0K8wBFp1G/images/new-code-snippet.png?fit=max&auto=format&n=n3EuZuJ0K8wBFp1G&q=85&s=3634c2dc8c3565fd77ab896d3fd07ed9" alt="TrueFoundry playground showing unified code snippet with base URL and model name" width="2940" height="1664" data-path="images/new-code-snippet.png" />
</Frame>

1. **Model Provider**: Select "OpenAI" from the dropdown
2. **Model Name**: Use the Virtual Model name you configured (e.g., `gpt-4o-mini`)
3. **OpenAI API Base**: Set to `{GATEWAY_BASE_URL}`

<Frame>
  <img src="https://mintcdn.com/truefoundry/5ZMNUF311tgGO_yT/images/langflow-2.png?fit=max&auto=format&n=5ZMNUF311tgGO_yT&q=85&s=29221f2df883643f4ff5e6f5a0bcca52" alt="Langflow OpenAI component configuration with TrueFoundry API settings" width="3840" height="1982" data-path="images/langflow-2.png" />
</Frame>

Replace `{GATEWAY_BASE_URL}` with your actual TrueFoundry Gateway Base URL.

### 3. Advanced Configuration in Agent Settings

For more advanced flows using agents, configure the agent component with TrueFoundry settings:

Ensure the following settings are configured:

* **Model Name**: Use the Virtual Model name (e.g., `gpt-4`)
* **OpenAI API Base**: `{GATEWAY_BASE_URL}`

## Usage Examples

### Basic Chat Flow

Create a simple chat flow using the configured Language Model:

<Frame>
  <img src="https://mintcdn.com/truefoundry/5ZMNUF311tgGO_yT/images/langflow-3.png?fit=max&auto=format&n=5ZMNUF311tgGO_yT&q=85&s=1e20aa7bf5c9c294818890ad5a6d6285" alt="Langflow canvas showing a basic chat flow with Chat Input, OpenAI model, and Chat Output components connected" width="3840" height="1976" data-path="images/langflow-3.png" />
</Frame>

1. Drag and drop a "Chat Input" component
2. Connect it to your configured "OpenAI" Language Model component
3. Connect the output to a "Chat Output" component
4. Run the flow to test the integration

### Multi-Agent RAG Application

For more complex applications involving RAG and multiple agents:

```python lines theme={"dark"}
# Example of using Langflow with TrueFoundry Gateway programmatically
from langflow import load_flow_from_json

# Load your flow configuration
flow = load_flow_from_json("path/to/your/flow.json")

# The flow will automatically use the configured TrueFoundry Gateway
# for all OpenAI model calls
result = flow.run(
    input_value="What is the capital of Brazil?",
    tweaks={
        "OpenAI-model_name": "gpt-4",  # Routed via Virtual Model
        "OpenAI-openai_api_base": "{GATEWAY_BASE_URL}",
        "OpenAI-openai_api_key": "your-truefoundry-api-key"
    }
)
```

### Environment Variables Configuration

Alternatively, you can set environment variables for easier configuration across multiple flows:

```bash lines theme={"dark"}
export OPENAI_API_KEY="your-truefoundry-api-key"
export OPENAI_API_BASE="{GATEWAY_BASE_URL}"
```

## Understanding Virtual Model Routing

When you use Langflow with standard model names like `gpt-4`, your requests are routed through the Virtual Model you configured. The Virtual Model maps the standard name to your actual provider model (e.g., `openai-main/gpt-4`).

You can configure more advanced routing within a Virtual Model, including multiple target models with different weights for load distribution and automatic failover. For full details, see the [Virtual Models documentation](/docs/ai-gateway/virtual-model).

## Benefits of Using TrueFoundry Gateway with Langflow

1. **Cost Tracking**: Monitor and track costs across all your Langflow applications
2. **Security**: Enhanced security with centralized API key management
3. **Access Controls**: Implement fine-grained access controls for different teams
4. **Rate Limiting**: Prevent API quota exhaustion with intelligent rate limiting
5. **Fallback Support**: Automatic failover to alternative providers when needed
6. **Analytics**: Detailed analytics and monitoring for all LLM calls
