> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# LibreChat

> Learn how to use librechat with TrueFoundry AI Gateway, including setup steps, use cases, and production-ready examples.

This guide provides instructions for integrating [LibreChat](https://librechat.ai/) with the Truefoundry AI Gateway.

## What is LibreChat?

LibreChat is a free, open-source platform that allows users to deploy and interact with various AI chat interfaces, like ChatGPT and Gemini, in one place.

### Key Features of LibreChat

1. **Multi-Provider AI Support**: LibreChat integrates with multiple AI services including OpenAI, Anthropic, AWS Bedrock, and Google, allowing users to switch between different models seamlessly within a single interface. This flexibility enables users to choose the most suitable model for their specific tasks.

2. **Code Interpreter API**: Execute code securely in multiple programming languages including Python, JavaScript, TypeScript, Go, C/C++, Java, PHP, Rust, and Fortran without requiring local setup. This feature enables [advanced code analysis and execution](https://librechat.ai/docs/features) capabilities directly within chat conversations.

3. **Custom AI Agents**: Create specialized AI assistants without coding by integrating tools like DALL-E-3, semantic search, calculators, and other APIs. These [no-code agents](https://librechat.ai/docs/features/agents) can be customized for specific workflows and business requirements, enhancing productivity through automated assistance.

## Prerequisites

To use [TrueFoundry's AI Gateway](https://www.truefoundry.com/ai-gateway) follow the quick start guide [here](https://docs.truefoundry.com/gateway/quick-start).

### Getting Base URL and Model Names

Get your TrueFoundry AI Gateway endpoint URL and model name from the unified code snippet:

<Frame>
  <img src="https://mintcdn.com/truefoundry/n3EuZuJ0K8wBFp1G/images/new-code-snippet.png?fit=max&auto=format&n=n3EuZuJ0K8wBFp1G&q=85&s=3634c2dc8c3565fd77ab896d3fd07ed9" alt="TrueFoundry playground showing unified code snippet with base URL and model name" width="2940" height="1664" data-path="images/new-code-snippet.png" />
</Frame>

## Using TrueFoundry AI Gateway with any LLM model

```yaml filename="librechat.yaml" lines theme={"dark"}
      - name: "TrueFoundry"
        apiKey: "${TRUEFOUNDRY_API_KEY}"
        baseURL: "${TRUEFOUNDRY_GATEWAY_URL}"
        models:
            default: ["openai-main/gpt-4o-mini", "openai-main/gpt-4o"]
            fetch: true
        titleConvo: true
        titleModel: "current_model"
        summarize: false
        summaryModel: "current_model"
        forcePrompt: false
        modelDisplayLabel: "TrueFoundry:OpenAI"
```

For more details you can check: [TrueFoundry Docs](https://docs.truefoundry.com/docs/introduction)
