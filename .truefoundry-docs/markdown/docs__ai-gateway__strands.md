> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Strands Agents

> Learn how to use strands with TrueFoundry AI Gateway, including setup steps, use cases, and production-ready examples.

This guide provides instructions for integrating [Strands Agents](https://strandsagents.com/) with the Truefoundry AI Gateway.

## What is Strands Agents?

Strands Agents is an open-source framework developed by AWS for building production-ready, multi-agent AI systems. It leverages model reasoning to plan, orchestrate tasks, and reflect on goals, making it ideal for enterprise-grade agentic applications.

### Key Features of Strands Agents

* **Model-Driven Orchestration**: Leverages model reasoning to plan, orchestrate tasks, and reflect on goals autonomously
* **Model & Provider Agnostic**: Work with any LLM provider - Amazon Bedrock, OpenAI, Anthropic, local models - without changing your code
* **Multi-Agent Primitives**: Simple primitives for handoffs, swarms, and graph workflows with built-in support for Agent-to-Agent (A2A) communication
* **Native AWS Integration**: Best-in-class AWS integrations with easy deployment to EKS, Lambda, EC2, and native MCP tool integration

## How TrueFoundry Integrates with Strands Agents

TrueFoundry enhances Strands Agents with production-grade observability, cost management, and multi-provider support through its LLM Gateway.

### Installation & Setup

<Steps>
  <Step title="Install Strands Agents">
    ```bash theme={"dark"}
    pip install -U strands-agents strands-agents-tools
    ```
  </Step>

  <Step title="Get TrueFoundry Access Token">
    1. Sign up for a [TrueFoundry account](https://www.truefoundry.com/register)
    2. Follow the steps here in [Quick start](https://docs.truefoundry.com/gateway/quick-start)
  </Step>

  <Step title="Configure Strands with TrueFoundry">
    <img src="https://mintcdn.com/truefoundry/n3EuZuJ0K8wBFp1G/images/new-code-snippet.png?fit=max&auto=format&n=n3EuZuJ0K8wBFp1G&q=85&s=3634c2dc8c3565fd77ab896d3fd07ed9" alt="TrueFoundry Code Configuration" width="2940" height="1664" data-path="images/new-code-snippet.png" />

    ```python theme={"dark"}
    from strands.models.openai import OpenAIModel

    # Create a model instance with TrueFoundry AI Gateway
    truefoundry_model = OpenAIModel(
        client_args={
            "api_key": "your_truefoundry_api_key",
            "base_url": "{GATEWAY_BASE_URL}"
        },
        model_id="openai-main/gpt-4o",  
        params={"temperature": 0.7}
    )

    # Use in your Strands agent
    from strands import Agent

    agent = Agent(model=truefoundry_model)
    response = agent("What is 2+2?")
    ```
  </Step>
</Steps>

### Multi-Provider Support

TrueFoundry's LLM Gateway provides an OpenAI-compatible API that works with all model providers:

```python theme={"dark"}
# OpenAI
openai_model = OpenAIModel(
    client_args={
        "api_key": "your_truefoundry_api_key",
        "base_url": "{GATEWAY_BASE_URL}"
    },
    model_id="llm-gateway-prod/gpt-4o",
    params={"temperature": 0.7}
)

# Anthropic Claude
claude_model = OpenAIModel(
    client_args={
        "api_key": "your_truefoundry_api_key",
        "base_url": "{GATEWAY_BASE_URL}"
    },
    model_id="llm-gateway-prod/claude-sonnet-4-5",
    params={"temperature": 0.7}
)

# Google Gemini
gemini_model = OpenAIModel(
    client_args={
        "api_key": "your_truefoundry_api_key",
        "base_url": "{GATEWAY_BASE_URL}"
    },
    model_id="llm-gateway-prod/gemini-2-0-flash-lite-001",
    params={"temperature": 0.7}
)

# Use different models for different agents
researcher = Agent(model=claude_model, tools=[web_search])
calculator_agent = Agent(model=openai_model, tools=[calculator])
```

### Observability and Governance

Monitor your Strands agents through TrueFoundry's metrics tab:

<img src="https://mintcdn.com/truefoundry/yRoKH_fkKi2nPtuV/images/gateway-metrics.png?fit=max&auto=format&n=yRoKH_fkKi2nPtuV&q=85&s=5a442952b4a398bcf6ab277d2392ca2c" alt="TrueFoundry metrics dashboard showing usage statistics, costs, and performance metrics for Strands agents" width="3840" height="1984" data-path="images/gateway-metrics.png" />

With Truefoundry's AI gateway, you can monitor and analyze:

* **Performance Metrics**: Track key latency metrics like Request Latency, Time to First Token (TTFS), and Inter-Token Latency (ITL) with P99, P90, and P50 percentiles
* **Cost and Token Usage**: Gain visibility into your application's costs with detailed breakdowns of input/output tokens and the associated expenses for each model
* **Usage Patterns**: Understand how your application is being used with detailed analytics on user activity, model distribution, and team-based usage
* **Rate Limiting and Virtual Models**: Set up rate limiting and configure [Virtual Models](/docs/ai-gateway/virtual-model) for intelligent routing and fallback across your models
