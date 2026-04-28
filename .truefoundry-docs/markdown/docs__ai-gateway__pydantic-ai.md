> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Pydantic AI

> Learn how to use pydantic ai with TrueFoundry AI Gateway, including setup steps, use cases, and production-ready examples.

This guide provides instructions for integrating [Pydantic AI](https://pydantic.dev/) with the Truefoundry AI Gateway.

## What is Pydantic AI?

Pydantic AI is a Python agent framework designed to make it less painful to build production-grade applications with Generative AI.

### Key Features of Pydantic AI

1. **Type Safe and Production Ready**: Pydantic AI is built to be type safe and ready for production-grade applications. It uses [Pydantic models](https://ai.pydantic.dev/) to ensure your AI agents work reliably in real business environments, with proper validation and error handling built-in from the start.

2. **Model Agnostic Support**: Works with all popular LLM providers including OpenAI, Anthropic, Google, and more. You can easily switch between different AI models without changing your code, giving you flexibility to choose the best model for each task.

3. **Built-in OpenTelemetry Integration**: Includes comprehensive observability and monitoring through [OpenTelemetry integration](https://logfire.pydantic.dev/docs/). Track performance, debug issues, and monitor your AI agents in production with detailed traces, logs, and metrics automatically collected and structured for analysis.

## Prerequisites

Before integrating Pydantic AI with TrueFoundry, ensure you have:

1. **TrueFoundry Account**: Create a [TrueFoundry account](https://www.truefoundry.com/register) and follow the instructions in our [Gateway Quick Start Guide](https://docs.truefoundry.com/gateway/quick-start)
2. **Pydantic AI Installation**: Install Pydantic AI in your Python environment:
   ```bash lines theme={"dark"}
   pip install pydantic-ai
   ```

You will get your base URL and model name directly from the unified code snippet:

<Frame>
  <img src="https://mintcdn.com/truefoundry/n3EuZuJ0K8wBFp1G/images/new-code-snippet.png?fit=max&auto=format&n=n3EuZuJ0K8wBFp1G&q=85&s=3634c2dc8c3565fd77ab896d3fd07ed9" alt="TrueFoundry playground showing unified code snippet with base URL and model name" width="2940" height="1664" data-path="images/new-code-snippet.png" />
</Frame>

## Integration Guide

### Step 1: Initialize Pydantic AI with TrueFoundry

Create an OpenAI model instance configured to use TrueFoundry's gateway:

```python lines theme={"dark"}
from pydantic_ai import Agent
from pydantic_ai.models.openai import OpenAIModel
from pydantic_ai.providers.openai import OpenAIProvider

model = OpenAIModel(
    model_name='openai-main/gpt-4o',  # Ensure you pick the correct model name
    provider=OpenAIProvider(
        base_url='{GATEWAY_BASE_URL}',
        api_key='YOUR_TRUEFOUNDRY_PAT'
    )
)
```

### Step 2: Basic Agent Example

Here's a simple example of creating a Pydantic AI agent with TrueFoundry:

```python lines theme={"dark"}
from pydantic_ai import Agent
from pydantic import BaseModel

# Define response structure
class SummaryResponse(BaseModel):
    title: str
    summary: str
    key_points: list[str]

# Create agent with TrueFoundry model
summary_agent = Agent(
    model=model,
    result_type=SummaryResponse,
    system_prompt='You are a helpful assistant that creates structured summaries.'
)

# Use the agent
async def summarize_text(text: str):
    result = await summary_agent.run(
        f"Summarize this text: {text}"
    )
    return result.data
```

By integrating Pydantic AI with TrueFoundry's AI Gateway, you can build type-safe, production-ready AI applications with comprehensive monitoring and cost control.
