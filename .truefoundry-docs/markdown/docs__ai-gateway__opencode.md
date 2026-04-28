> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# OpenCode

> Learn how to integrate OpenCode with TrueFoundry AI Gateway for secure, governed AI-powered coding in the terminal, desktop, or IDE.

This guide provides instructions for integrating [OpenCode](https://opencode.ai/) with the TrueFoundry AI Gateway.

## What is OpenCode?

OpenCode is an open-source AI coding agent available as a terminal-based interface, desktop app, or IDE extension. It uses the [AI SDK](https://ai-sdk.dev/) and [Models.dev](https://models.dev/) to support 75+ LLM providers, and can also run local models. With TrueFoundry AI Gateway integration, you can route all OpenCode LLM requests through TrueFoundry's Gateway for centralized access control, cost tracking, rate limiting, guardrails, and observability.

### Key Features of OpenCode

* **[AI-Powered Coding Agent](https://opencode.ai/)**: A full-featured coding agent that can read, write, and edit files, run shell commands, and navigate complex codebases directly from your terminal or desktop
* **[Multiple Agent Modes](https://opencode.ai/docs/agents/)**: Built-in Build and Plan agents with the ability to create custom agents for specialized tasks like code review, documentation, or security auditing
* **[75+ Provider Support](https://opencode.ai/docs/providers/)**: Connect to any LLM provider through a unified interface, including custom OpenAI-compatible endpoints like the TrueFoundry AI Gateway

## Prerequisites

Before integrating OpenCode with TrueFoundry, ensure you have:

1. **TrueFoundry Account**: Create a [TrueFoundry account](https://www.truefoundry.com/register) and follow the instructions in our [Gateway Quick Start Guide](https://docs.truefoundry.com/docs/ai-gateway/quick-start)
2. **OpenCode Installation**: Install OpenCode by following the [official documentation](https://opencode.ai/docs/)

## Integration Guide

This guide uses the OpenCode Desktop app for illustration, but the same configuration applies to the terminal-based (TUI) and IDE extension versions.

### Step 1: Open Provider Settings

1. Open the OpenCode Desktop app.
2. Navigate to **Providers** in the left sidebar.
3. Click **+ Connect** next to **Custom provider**.

<Frame caption="OpenCode Providers page showing Custom provider option">
  <img src="https://mintcdn.com/truefoundry/vPP3_T20zeD-LDhH/images/opencode-connect-provider.png?fit=max&auto=format&n=vPP3_T20zeD-LDhH&q=85&s=49f6d99a8cb2356582c279e8be0cfd94" alt="OpenCode Desktop Providers page with Custom provider highlighted at the bottom of the provider list" width="1898" height="1180" data-path="images/opencode-connect-provider.png" />
</Frame>

### Step 2: Configure TrueFoundry as a Custom Provider

Fill in the following details in the Custom provider form:

* **Provider ID**: `tfy-gateway` (or any identifier using lowercase letters, numbers, hyphens, or underscores)
* **Display name**: `truefoundry`
* **Base URL**: Your TrueFoundry AI Gateway URL (e.g., `{GATEWAY_BASE_URL}`). You can get this from the unified code snippet in the TrueFoundry AI Gateway Playground.
* **API key**: Your TrueFoundry API key

<Frame caption="Custom provider form with TrueFoundry configuration">
  <img src="https://mintcdn.com/truefoundry/vPP3_T20zeD-LDhH/images/opencode-tfy-addition.png?fit=max&auto=format&n=vPP3_T20zeD-LDhH&q=85&s=882319597fca0f88604276bab6167dda" alt="OpenCode Custom provider configuration form showing Provider ID as tfy-gateway, Display name as truefoundry, and Base URL pointing to the TrueFoundry AI Gateway endpoint" width="1264" height="1010" data-path="images/opencode-tfy-addition.png" />
</Frame>

<Frame caption="Get Base URL and API key from the unified code snippet">
  <img src="https://mintcdn.com/truefoundry/n3EuZuJ0K8wBFp1G/images/new-code-snippet.png?fit=max&auto=format&n=n3EuZuJ0K8wBFp1G&q=85&s=3634c2dc8c3565fd77ab896d3fd07ed9" alt="TrueFoundry playground showing unified code snippet with base URL and model name" width="2940" height="1664" data-path="images/new-code-snippet.png" />
</Frame>

### Step 3: Add Models

Scroll down in the Custom provider form to add models:

1. In the **Models** section, enter the model ID from TrueFoundry (e.g., `openai-main/gpt-5-codex`) in the first field and a display name (e.g., `tfy-gpt-5-codex`) in the second field.
2. Click **+ Add model** to add more models as needed.
3. Optionally, add custom **Headers** for tracking. For example, set `application` to `opencode` to tag all requests from OpenCode in TrueFoundry's observability dashboard.
4. Click **Submit** to save the configuration.

<Frame caption="Adding models and custom headers in the Custom provider form">
  <img src="https://mintcdn.com/truefoundry/vPP3_T20zeD-LDhH/images/opencode-model-addition.png?fit=max&auto=format&n=vPP3_T20zeD-LDhH&q=85&s=80f56bda84da6d86e293e62edc6bb7eb" alt="OpenCode Custom provider form showing model configuration with openai-main/gpt-5-codex as model ID, tfy-gpt-5-codex as display name, and a custom application header set to opencode" width="1254" height="1000" data-path="images/opencode-model-addition.png" />
</Frame>

<Note>
  **Minimum 128K context window required**

  OpenCode includes a detailed system prompt with tool definitions, agent instructions, and project context that consumes a significant number of input tokens on every request.

  * Models with smaller context windows (e.g., 8K or 32K) **will fail** with `prompt is too long` errors
  * The system prompt combined with conversation history and tool call results quickly exceeds smaller limits
  * Refer to the [OpenCode recommended models](https://opencode.ai/docs/models/#recommended-models) for models that are known to work well

  Make sure the model you configure through the TrueFoundry AI Gateway supports **at least 128K input tokens**.
</Note>

### Step 4: Select a TrueFoundry Model and Start Coding

1. In the OpenCode chat interface, click the model selector at the bottom of the screen.
2. You will see your configured TrueFoundry models listed under the **truefoundry** provider.
3. Select the model you want to use and start coding.

<Frame caption="TrueFoundry models available in OpenCode model selector">
  <img src="https://mintcdn.com/truefoundry/vPP3_T20zeD-LDhH/images/opencode-application-view.png?fit=max&auto=format&n=vPP3_T20zeD-LDhH&q=85&s=5578b2017a6ad7895630e62f482163bf" alt="OpenCode model selector showing TrueFoundry models including tfy-claude-4-5, tfy-gpt-4o, tfy-gpt-5, tfy-gpt-5-codex, and tfy-gpt-5.2 under the truefoundry provider" width="2144" height="1182" data-path="images/opencode-application-view.png" />
</Frame>

## Alternative: Configuration via JSON

If you prefer configuring OpenCode through its JSON config file (useful for TUI or team-wide settings), add the following to your `opencode.json`:

```json theme={"dark"}
{
  "$schema": "https://opencode.ai/config.json",
  "provider": {
    "tfy-gateway": {
      "npm": "@ai-sdk/openai-compatible",
      "name": "truefoundry",
      "options": {
        "baseURL": "{GATEWAY_BASE_URL}",
        "apiKey": "<your-truefoundry-api-key>",
        "headers": {
          "application": "opencode"
        }
      },
      "models": {
        "openai-main/gpt-5-codex": {
          "name": "tfy-gpt-5-codex"
        },
        "anthropic-main/claude-sonnet-4-5": {
          "name": "tfy-claude-4-5"
        }
      }
    }
  }
}
```

Then set the default model to one of your TrueFoundry models:

```json theme={"dark"}
{
  "$schema": "https://opencode.ai/config.json",
  "model": "tfy-gateway/openai-main/gpt-5-codex"
}
```

## Observability and Governance

Monitor your OpenCode usage through TrueFoundry's observability dashboard. With the `application: opencode` header configured, you can filter and analyze:

* **Performance Metrics**: Track request latency, time to first token, and inter-token latency
* **Cost and Token Usage**: Monitor input/output tokens and associated costs per model
* **Usage Patterns**: Understand usage across models, users, and teams
* **Rate Limiting and Virtual Models**: Set up rate limiting and configure [Virtual Models](/docs/ai-gateway/virtual-model) for intelligent routing and fallback across your models
