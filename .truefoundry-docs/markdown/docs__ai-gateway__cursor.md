> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Cursor

> Set up Cursor with the TrueFoundry AI Gateway for AI-powered development.

This guide provides instructions for integrating [Cursor](https://cursor.sh/) with the Truefoundry AI Gateway.

## What is Cursor?

Cursor is an AI-powered code editor that enhances developer productivity with intelligent code completion, generation, and debugging capabilities. Built upon Visual Studio Code, it integrates artificial intelligence features directly into the coding environment to streamline the development process.

### Key Features of Cursor

* **[AI-Powered Code Completion](https://cursor.com/features)**: Provides real-time, context-aware code suggestions and multi-line completions that accelerate the coding process with intelligent autocomplete capabilities
* **[Intelligent Code Understanding](https://docs.cursor.com/)**: Offers code explanation, documentation generation, and smart navigation to help developers comprehend and manage complex codebases effectively
* **[Integrated AI Chat Assistant](https://cursor.com/features)**: Includes built-in AI chat for instant programming help, supports multiple programming languages and frameworks, and integrates with version control systems for seamless development workflow

## Prerequisites

Before integrating Cursor with TrueFoundry, ensure you have:

1. **TrueFoundry Account**: Create a [Truefoundry account](https://www.truefoundry.com/register) and follow the instructions in our [Gateway Quick Start Guide](https://docs.truefoundry.com/gateway/quick-start)
2. **Cursor Installation**: Download and install Cursor from the [official website](https://cursor.sh/)
3. **Virtual Model**: Create a Virtual Model for each model you want to use in Cursor (see [Create a Virtual Model](#step-1-create-a-virtual-model) below)

## Important Network Requirements

<Note>
  **Gateway URL Public Exposure Required**: Your TrueFoundry Gateway URL must be exposed to public internet for Cursor integration to work properly. If your gateway is behind a VPN, the request flow is: Cursor → Cursor Server → Service Provider Server. This means Cursor's servers need to be able to reach your TrueFoundry Gateway endpoint. For more information about Cursor's security practices, visit [Cursor's security page](https://cursor.sh/security).
</Note>

## Why You Need a Virtual Model

Cursor has internal logic that works optimally with standard OpenAI model names (like `gpt-4o`, `claude-3-sonnet`), but may experience compatibility issues with TrueFoundry's fully qualified model names (like `openai-main/gpt-4o` or `azure-openai/gpt-4o`). When Cursor encounters these fully qualified names directly, it may not function as expected due to internal processing differences.

**Virtual Models fix this** by letting you:

1. Use a standard model name as the Virtual Model name in Cursor (e.g., `gpt-4o`) that Cursor recognizes.
2. Have the TrueFoundry Gateway map that name to the fully qualified target model (e.g., `openai-main/gpt-4o`) and route requests there.

This gives you seamless compatibility while still routing all requests through the TrueFoundry Gateway. To learn more, see the [Virtual Models documentation](/docs/ai-gateway/virtual-model).

## Integration Guide

### Step 1: Create a Virtual Model

Create a Virtual Model so Cursor can use a simple model name that the Gateway maps to your provider. Follow these steps:

1. **Open the Virtual Model / Routing page** in the TrueFoundry AI Gateway dashboard

<img className="hidden dark:block" src="https://mintcdn.com/truefoundry/9sjb7wvpsYV7RE1W/images/Screenshot2026-02-27at4.38.04PM-1.png?fit=max&auto=format&n=9sjb7wvpsYV7RE1W&q=85&s=5815438d18a9ecd74e2c183b7aa8d9a1" alt="TrueFoundry AI Gateway Virtual Models page in dark mode" width="3598" height="2008" data-path="images/Screenshot2026-02-27at4.38.04PM-1.png" />

<img className="dark:hidden" src="https://mintcdn.com/truefoundry/uL7y0Q9ywygbGxPN/images/Screenshot2026-02-27at4.38.04PM.png?fit=max&auto=format&n=uL7y0Q9ywygbGxPN&q=85&s=84550e1c21cc38734bfbabe680a189f1" alt="TrueFoundry AI Gateway Virtual Models page" width="3598" height="2008" data-path="images/Screenshot2026-02-27at4.38.04PM.png" />

2. **Create a new Virtual Model** and give it a name. This name (in the format `{group-name}/{model-name}`) is what you will use in Cursor (e.g., `cursor-vm/gpt-5-3-codex`).

<Frame caption="Create a Virtual Model Provider Group with models for Cursor">
  <img src="https://mintcdn.com/truefoundry/w7u-ph19zuQ3z4Sl/images/2026-03-11_09.56.35.png?fit=max&auto=format&n=w7u-ph19zuQ3z4Sl&q=85&s=25a849c549d701229eb9d60178159e7a" alt="TrueFoundry Virtual Model Provider Group page showing cursor-vm group with collaborators and virtual models" width="1336" height="536" data-path="images/2026-03-11_09.56.35.png" />
</Frame>

3. **Set the target** to the fully qualified model name (e.g., `openai-main/gpt-4o-mini`). You can use one target at 100% weight or add multiple targets with weights for routing across providers.

<Frame caption="Configure Virtual Model with target model and routing strategy">
  <img src="https://mintcdn.com/truefoundry/w7u-ph19zuQ3z4Sl/images/2026-03-11_09.57.25.png?fit=max&auto=format&n=w7u-ph19zuQ3z4Sl&q=85&s=73e7c5c656ca931cda6adc63836c0d06" alt="TrueFoundry Virtual Model configuration showing target model openai-main/gpt-4o-mini with weight-based routing" width="746" height="608" data-path="images/2026-03-11_09.57.25.png" />
</Frame>

**Example:** If your Virtual Model is `cursor-vm/dummy` and the target is `openai-main/gpt-4o-mini` at 100% weight, then when you use `cursor-vm/dummy` in Cursor, the Gateway routes that request to `openai-main/gpt-4o-mini`.

<Note>
  Use the Virtual Model name (e.g., `cursor-vm/dummy`) in Cursor. Do not use the fully qualified name (e.g., `openai-main/gpt-4o-mini`) directly in Cursor, as that can cause compatibility issues.
</Note>

### Step 2: Get TrueFoundry Gateway Configuration

Gather your TrueFoundry Gateway details:

1. **Navigate to AI Gateway Playground**: Go to your TrueFoundry AI Gateway playground
2. **Access Unified Code Snippet**: Use the unified code snippet
3. **Copy Base URL**: You will get the base URL from the unified code snippet

<Frame>
  <img src="https://mintcdn.com/truefoundry/n3EuZuJ0K8wBFp1G/images/new-code-snippet.png?fit=max&auto=format&n=n3EuZuJ0K8wBFp1G&q=85&s=3634c2dc8c3565fd77ab896d3fd07ed9" alt="TrueFoundry playground showing unified code snippet with base URL and model name" width="2940" height="1664" data-path="images/new-code-snippet.png" />
</Frame>

### Step 3: Configure Cursor with TrueFoundry Gateway

1. Open **Cursor Settings**
2. In the **Models** section, add your Virtual Model names as custom models (e.g., `cursor-vm/gpt-5-3-codex`, `cursor-vm/dummy`)
3. Under **API Keys**, enter your TrueFoundry API key in the **OpenAI API Key** field
4. Enable **Override OpenAI Base URL** and enter your TrueFoundry Gateway base URL

<Frame caption="Cursor settings showing Virtual Models, API key, and Gateway base URL configuration">
  <img src="https://mintcdn.com/truefoundry/oyTyqm78duGfISb7/images/cursor-settings.png?fit=max&auto=format&n=oyTyqm78duGfISb7&q=85&s=2664b7e52f1cde00242408753fde4feb" width="500" alt="Cursor settings page showing virtual models added as custom models, TrueFoundry API key in OpenAI API Key field, and Gateway base URL in Override OpenAI Base URL field" data-path="images/cursor-settings.png" />
</Frame>

### Step 4: Start Using Cursor

Select your Virtual Model from the model dropdown in Cursor and start sending requests. All your calls from Cursor will now go through the TrueFoundry Gateway, with automatic routing to the target models configured in your Virtual Model.

You can add multiple targets with different weights in your Virtual Model for intelligent routing and fallback across providers. To learn more about routing strategies, see the [Virtual Models documentation](/docs/ai-gateway/virtual-model).

### Request Traces

You can monitor and inspect your Cursor requests in the TrueFoundry AI Gateway dashboard. Navigate to the request traces view to see request details, latency, token usage, and routing information for each call.

<Frame caption="Request traces view in the TrueFoundry AI Gateway">
  <img src="https://mintcdn.com/truefoundry/se5cTTB7B6tyoRdV/images/2026-03-11_10.26.59.png?fit=max&auto=format&n=se5cTTB7B6tyoRdV&q=85&s=84bbacba8961b94c010c2b0c73f93126" alt="TrueFoundry AI Gateway request traces view showing Cursor requests, latency, and routing details" width="778" height="111" data-path="images/2026-03-11_10.26.59.png" />
</Frame>
