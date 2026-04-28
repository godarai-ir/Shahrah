> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# OpenAI Swarm

> Learn how to use openai swarm with TrueFoundry AI Gateway, including setup steps, use cases, and production-ready examples.

This guide provides instructions for integrating [OpenAI Swarm](https://github.com/openai/swarm) with the Truefoundry AI Gateway.

## What is OpenAI Swarm?

OpenAI Swarm is an experimental framework for building, orchestrating, and deploying multi-agent systems. It provides a lightweight, scalable, and highly customizable approach to coordinating multiple AI agents that can work together to solve complex tasks through handoffs and collaboration.

### Key Features of OpenAI Swarm

* **[Multi-Agent Coordination](https://github.com/openai/swarm)**: Build teams of specialized agents that can transfer tasks between each other and collaborate on complex workflows
* **[Lightweight Framework](https://github.com/openai/swarm)**: Minimal overhead with a simple API that makes it easy to define agents, their capabilities, and coordination patterns
* **[Function Calling](https://github.com/openai/swarm)**: Agents can be equipped with custom functions and tools to interact with external systems and APIs
* **[Contextual Handoffs](https://github.com/openai/swarm)**: Seamless transfer of context and conversation state between agents based on user needs and agent capabilities

## Prerequisites

Before integrating OpenAI Swarm with TrueFoundry, ensure you have:

1. **TrueFoundry Account**: Create a [Truefoundry account](https://www.truefoundry.com/register) and follow our [Gateway Quick Start](https://docs.truefoundry.com/gateway/quick-start)
2. **OpenAI Swarm Installation**: Install OpenAI Swarm using pip: `pip install git+https://github.com/openai/swarm.git`
3. **Virtual Model**: Create a Virtual Model for your desired models (see [Configure Model Routing](#2-configure-model-routing) below)

## Important: Virtual models

<Warning>
  **Virtual Model Required**: OpenAI Swarm uses standard OpenAI model names (like `gpt-4o`, `gpt-4o-mini`) in its internal logic. To route these requests through TrueFoundry Gateway to your specific model providers, you need to create a Virtual Model on TrueFoundry. This maps standard model names to your TrueFoundry model format (e.g., `openai-main/gpt-4o`).

  For detailed information, see our [Virtual Models documentation](/docs/ai-gateway/virtual-model).
</Warning>

## Setup Process

### 1. Basic Setup with OpenAI Swarm

You will get your 'truefoundry-api-key', 'truefoundry-gateway-url' and model name directly from the unified code snippet

<Frame>
  <img src="https://mintcdn.com/truefoundry/n3EuZuJ0K8wBFp1G/images/new-code-snippet.png?fit=max&auto=format&n=n3EuZuJ0K8wBFp1G&q=85&s=3634c2dc8c3565fd77ab896d3fd07ed9" width="2940" height="1664" data-path="images/new-code-snippet.png" />
</Frame>

Configure the OpenAI client with TrueFoundry gateway settings and use it with Swarm:

```python lines theme={"dark"}
from swarm import Swarm, Agent
import os
from openai import OpenAI

# Get API key from environment variable
api_key = os.getenv("OPENAI_API_KEY")  # Set this to your TrueFoundry API key

# Initialize OpenAI client with TrueFoundry gateway
openai_client = OpenAI(
    api_key="your-truefoundry-api-key",
    base_url="{GATEWAY_BASE_URL}",
)

# Initialize Swarm with the TrueFoundry-configured OpenAI client
client = Swarm(client=openai_client)

def transfer_to_agent_b():
    return agent_b

agent_a = Agent(
    name="Agent A",
    instructions="You are a helpful agent.",
    functions=[transfer_to_agent_b],
)

agent_b = Agent(
    name="Agent B",
    instructions="Only speak in Haikus.",
)

response = client.run(
    agent=agent_a,
    messages=[{"role": "user", "content": "I want to talk to agent B."}],
)

print(response.messages[-1]["content"])
```

Replace:

* `your-truefoundry-api-key` with your actual TrueFoundry API key
* `{GATEWAY_BASE_URL}` with your TrueFoundry Gateway URL

### 2. Configure Model Routing

Create a Virtual Model so OpenAI Swarm can use a simple model name that the Gateway maps to your provider:

1. Navigate to **AI Gateway → Models → Virtual Model** in the TrueFoundry dashboard.

<Frame>
  <img src="https://mintcdn.com/truefoundry/9sjb7wvpsYV7RE1W/images/Screenshot2026-02-27at4.38.04PM-1.png?fit=max&auto=format&n=9sjb7wvpsYV7RE1W&q=85&s=5815438d18a9ecd74e2c183b7aa8d9a1" alt="TrueFoundry Virtual Models dashboard" className="block dark:hidden" width="3598" height="2008" data-path="images/Screenshot2026-02-27at4.38.04PM-1.png" />

  <img src="https://mintcdn.com/truefoundry/uL7y0Q9ywygbGxPN/images/Screenshot2026-02-27at4.38.04PM.png?fit=max&auto=format&n=uL7y0Q9ywygbGxPN&q=85&s=84550e1c21cc38734bfbabe680a189f1" alt="TrueFoundry Virtual Models dashboard" className="hidden dark:block" width="3598" height="2008" data-path="images/Screenshot2026-02-27at4.38.04PM.png" />
</Frame>

2. Create a new Virtual Model Provider Group with a name (e.g., `swarm-vm`) and configure collaborators for access control.

3. Add your target model (e.g., `openai-main/gpt-4o`) under the provider group. Set the Virtual Model name to the model name Swarm expects (e.g., `gpt-4o`), so requests from Swarm are automatically routed to your configured provider.

For more details, see the [Virtual Models documentation](/docs/ai-gateway/virtual-model).

### 3. Environment Variables Configuration

For persistent configuration across your Swarm applications, set these environment variables:

```bash lines theme={"dark"}
export OPENAI_API_KEY="your-truefoundry-api-key"
export OPENAI_BASE_URL="{GATEWAY_BASE_URL}"
```

## Observability and Governance

Monitor your OpenAI Swarm applications through TrueFoundry's metrics tab:

<img src="https://mintcdn.com/truefoundry/yRoKH_fkKi2nPtuV/images/gateway-metrics.png?fit=max&auto=format&n=yRoKH_fkKi2nPtuV&q=85&s=5a442952b4a398bcf6ab277d2392ca2c" alt="TrueFoundry metrics" width="3840" height="1984" data-path="images/gateway-metrics.png" />

With TrueFoundry's AI gateway, you can monitor and analyze:

* **Performance Metrics**: Track key latency metrics like Request Latency, Time to First Token (TTFS), and Inter-Token Latency (ITL) with P99, P90, and P50 percentiles
* **Cost and Token Usage**: Gain visibility into your application's costs with detailed breakdowns of input/output tokens and the associated expenses for each model
* **Usage Patterns**: Understand how your application is being used with detailed analytics on user activity, model distribution, and team-based usage
* **Agent Performance**: Monitor individual agent performance and handoff patterns
* **Rate Limiting and Virtual Models**: Set up rate limiting and configure [Virtual Models](/docs/ai-gateway/virtual-model) for intelligent routing and fallback across your models
