> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# OpenAI Codex CLI

> Use OpenAI Codex CLI with TrueFoundry LLM Gateway to interact with AI models from your terminal

This guide provides instructions for using the [Codex CLI](https://github.com/openai/codex) through the Truefoundry LLM Gateway.

<Note>
  **GPT-5.3-Codex is now available!** You can access GPT-5.3-Codex through TrueFoundry AI Gateway. Follow the setup steps below to get started.
</Note>

## What is Codex?

Codex is the official command-line interface (CLI) tool for OpenAI, providing a streamlined way to interact with OpenAI's language models directly from your terminal. With Truefoundry LLM Gateway integration, you can route your Codex requests via Gateway.

### Key Features of OpenAI Codex CLI

1. **Terminal-Native AI Interactions**: Chat with AI models directly from your terminal without switching contexts
2. **Intelligent Code Generation**: Generate code snippets, functions, and programs across multiple programming languages using natural language prompts
3. **Streaming and Interactive Sessions**: Real-time streaming responses enable dynamic, conversation-like interactions for code development

## Prerequisites

Before integrating Codex with TrueFoundry, ensure you have:

1. **TrueFoundry Account**: Create a [Truefoundry account](https://www.truefoundry.com/register) with at least one model provider and generate a Personal Access Token by following the instructions in [Generating Tokens](/docs/ai-gateway/authentication). For a quick setup guide, see our [Gateway Quick Start](https://docs.truefoundry.com/gateway/quick-start)
2. **Codex Installation**: Install the [Codex CLI](https://github.com/openai/codex) on your system
3. **Virtual Model**: Create a Virtual Model for each Codex model you want to use (see [Create a Virtual Model](#2-create-a-virtual-model) below)

## Why You Need a Virtual Model

Codex has internal logic that sends **thinking tokens** to certain models during processing. This works correctly with standard OpenAI-style model names (like `gpt-5`), but causes compatibility issues with TrueFoundry’s **fully qualified** model names (like `openai-main/gpt-5` or `azure-openai/gpt-5`). When Codex sees a fully qualified name directly, it can send thinking tokens incorrectly, which leads to unexpected behavior.

**Virtual Models fix this** by letting you:

1. Use a **slug** as the model name in Codex (e.g. `gpt-5` or `gpt-5.2-codex`) so thinking tokens work as intended.
2. Have the TrueFoundry Gateway map that slug to the fully qualified target (e.g. `openai-main/gpt-5`) and route requests there.

You get correct thinking-token behavior and can still use any model through the Gateway.

## Setup Process

### 1. Configure Environment Variables

To connect Codex with Truefoundry LLM Gateway, set these environment variables:

```bash lines theme={"dark"}
export OPENAI_API_KEY=TFY_API_KEY
export OPENAI_BASE_URL="{GATEWAY_BASE_URL}"
```

Replace `TFY_API_KEY` with your actual TrueFoundry API key and `{GATEWAY_BASE_URL}` with your TrueFoundry AI Gateway Base URL ([how to find it](/docs/ai-gateway/quick-start#gateway-base-url)).

<Frame caption="Get Base URL and Model Name from Unified Code Snippet">
  <img src="https://mintcdn.com/truefoundry/9XxUDrDVg1mtOSWB/images/Screenshot2025-10-06at6.57.08PM.png?fit=max&auto=format&n=9XxUDrDVg1mtOSWB&q=85&s=d5b72d6dd92c100c51be07cc95031c8a" alt="TrueFoundry playground showing unified code snippet with base URL and model name highlighted for Codex CLI integration" width="3600" height="2000" data-path="images/Screenshot2025-10-06at6.57.08PM.png" />
</Frame>

<Note>
  **Tip**: Add these lines to your shell profile (`.bashrc`, `.zshrc`, etc.) to make the configuration persistent across terminal sessions.
</Note>

### 2. Create a Virtual Model

Create a Virtual Model so Codex can use a simple model name that the Gateway maps to your provider. Follow these steps:

1. **Open the Virtual Model / Routing page** in the TrueFoundry AI Gateway dashboard
2. **Create a new Virtual Model** and add a **Slug**. This slug is the name you will use in Codex (e.g. `gpt-5.2-codex` or `gpt-5`).

   <Note>
     Use a slug that matches the **actual model ID** (e.g. `gpt-5.2-codex` for GPT-5.2-Codex). Codex recognizes these IDs and enables the right features (e.g. thinking tokens)
   </Note>
3. **Set the target** to the fully qualified model name (e.g. `openai-main/gpt-5.2-codex`). You can use one target at 100% weight or add multiple targets with weights for load balancing.

<img className="hidden dark:block" src="https://mintcdn.com/truefoundry/9sjb7wvpsYV7RE1W/images/Screenshot2026-02-27at4.38.04PM-1.png?fit=max&auto=format&n=9sjb7wvpsYV7RE1W&q=85&s=5815438d18a9ecd74e2c183b7aa8d9a1" alt="TrueFoundry AI Gateway Routing Configuration setup showing virtual model mapping" width="3598" height="2008" data-path="images/Screenshot2026-02-27at4.38.04PM-1.png" />

<img className="dark:hidden" src="https://mintcdn.com/truefoundry/uL7y0Q9ywygbGxPN/images/Screenshot2026-02-27at4.38.04PM.png?fit=max&auto=format&n=uL7y0Q9ywygbGxPN&q=85&s=84550e1c21cc38734bfbabe680a189f1" alt="TrueFoundry AI Virtual Model Dashboard" width="3598" height="2008" data-path="images/Screenshot2026-02-27at4.38.04PM.png" />

<Frame>
  <img src="https://mintcdn.com/truefoundry/uL7y0Q9ywygbGxPN/images/Screenshot2026-02-27at4.39.44PM.png?fit=max&auto=format&n=uL7y0Q9ywygbGxPN&q=85&s=023c680646cf5f6c28aa1835da2141fa" alt="Screenshot2026 02 27at4 39 44PM" width="1434" height="1982" data-path="images/Screenshot2026-02-27at4.39.44PM.png" />
</Frame>

**Example:** If your slug is `gpt-5.2-codex` and the target is `openai-main/gpt-5.2-codex` at 100% weight, then when you run `codex chat --model gpt-5.2-codex`, the Gateway routes that request to `openai-main/gpt-5.2-codex`.

<Note>
  Use the slug as the model name in Codex (e.g. `--model gpt-5.2-codex`). Do not use the fully qualified name (e.g. `openai-main/gpt-5.2-codex`) in Codex; that can break thinking tokens and other behavior.
</Note>

## Usage Examples

Use the **virtual model slug** you created (e.g. `gpt-5.2-codex` or `gpt-5`) with the `--model` flag so requests go through the TrueFoundry Gateway.

### Basic usage

```bash lines theme={"dark"}
# Chat using your virtual model slug
codex chat --model gpt-5.2-codex "Generate a Python function to calculate the Fibonacci sequence"

codex chat --model gpt-5.2-codex "Explain quantum computing in simple terms"

# With streaming
codex chat --model gpt-5.2-codex --stream "Write a short story about AI"
```

### With options

```bash lines theme={"dark"}
codex chat --model gpt-5.2-codex --temperature 0.7 "Generate creative marketing slogans"

codex chat --model gpt-5.2-codex --max-tokens 500 "Write a detailed explanation of relativity"

# Piping through Codex
cat input.txt | codex chat --model gpt-5.2-codex "Translate to French" | codex chat --model gpt-5.2-codex "Make this more formal"
```

Replace `gpt-5.2-codex` with whatever slug you set in your Virtual Model.

## Understanding Virtual Model Routing

When you use `--model <your-slug>`, the Gateway routes the request according to that Virtual Model’s targets. With a single target at 100% weight, all traffic goes to that model.

You can also add **multiple targets with weights** in the Virtual Model (e.g. 70% to one provider, 30% to another). Example:

```yaml lines theme={"dark"}
load_balance_targets:
  - target: openai-main/gpt-5
    weight: 70
  - target: azure-openai/gpt-5
    weight: 30
```

Then about 70% of requests go to the first provider and 30% to the second.
