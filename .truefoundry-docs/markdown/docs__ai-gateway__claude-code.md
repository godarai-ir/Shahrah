> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Claude Code

> How to connect Claude Code with the TrueFoundry AI Gateway.

This guide provides instructions for integrating [Claude Code](https://docs.anthropic.com/en/docs/claude-code) with the Truefoundry AI Gateway.

<Warning>
  **Claude Code Desktop is not supported.** The [Claude Code Desktop app](https://code.claude.com/docs/en/desktop) connects directly to Anthropic's API and does not reliably support routing requests through third-party gateways via `ANTHROPIC_BASE_URL`. This integration works with the **Claude Code CLI** and the **VS Code extension** (which inherits the CLI configuration).
</Warning>

## What is Claude Code?

Claude Code is Anthropic's agentic coding tool that lives in your terminal, understands your codebase, and helps you code faster through natural language commands. It serves as an intelligent pair programming partner that can read, analyze, and work with your entire project structure.

## Model Configuration

Claude Code supports model aliases (`opus`, `sonnet`, `haiku`, `opusplan`) that provide convenient shortcuts for different use cases. When integrating with TrueFoundry, you can configure these aliases to use your specific TrueFoundry models through environment variables.

<Note>
  Model aliases allow you to switch between different Claude models during your coding session using simple commands like `/model opus` or `/model sonnet`.
</Note>

## Prerequisites

Before integrating Claude Code with TrueFoundry, ensure you have:

1. **TrueFoundry Account**: Create a [Truefoundry account](https://www.truefoundry.com/register) and follow the instructions in our [Gateway Quick Start Guide](https://docs.truefoundry.com/gateway/quick-start)
2. **Claude Code Installation**: Install Claude Code following the [official documentation](https://docs.anthropic.com/en/docs/claude-code)
3. Configure models in Truefoundry for each provider you want to use

## Integrate Truefoundry with Claude Code

Claude Code uses a settings.json file for configuration. You'll need to modify the environment variables to route requests through Truefoundry.

First, get the base URL and model name from your TrueFoundry AI Gateway playground using the unified code snippet:

<Frame>
  <img src="https://mintcdn.com/truefoundry/n3EuZuJ0K8wBFp1G/images/new-code-snippet.png?fit=max&auto=format&n=n3EuZuJ0K8wBFp1G&q=85&s=3634c2dc8c3565fd77ab896d3fd07ed9" alt="TrueFoundry playground showing unified code snippet with base URL and model name" width="2940" height="1664" data-path="images/new-code-snippet.png" />
</Frame>

Claude Code can be configured either globally or per project by editing the settings file:

* Global: `~/.claude/settings.json`
* Project-specific: `.claude/settings.json` in your project directory

```json lines theme={"dark"}
{
  "env": {
    "ANTHROPIC_BASE_URL": "{GATEWAY_BASE_URL}",
    "ANTHROPIC_AUTH_TOKEN": "your-truefoundry-api-key",
    "ANTHROPIC_MODEL": "anthropic/claude-4-sonnet-20250514",
    "ANTHROPIC_DEFAULT_OPUS_MODEL": "anthropic/claude-4-opus-20250514",
    "ANTHROPIC_DEFAULT_SONNET_MODEL": "anthropic/claude-4-sonnet-20250514",
    "ANTHROPIC_DEFAULT_HAIKU_MODEL": "anthropic/claude-3-5-haiku-20241022",
    "CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS": "1",
    "ANTHROPIC_CUSTOM_HEADERS": "x-tfy-anthropic-beta: context-management-2025-06-27"
  }
}
```

<Note>
  `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS` disables experimental features in Claude Code for stable behavior with TrueFoundry Gateway.
</Note>

### Forwarding Anthropic Beta Features

Anthropic releases beta features (e.g. [context management](https://docs.anthropic.com/en/docs/build-with-claude/context-editing)) that require the `anthropic-beta` header on API requests. To forward a beta flag through the TrueFoundry Gateway, set the `x-tfy-anthropic-beta` header via `ANTHROPIC_CUSTOM_HEADERS` as shown in the config above.

To enable multiple beta features, pass them as a comma-separated list (no spaces):

```json theme={"dark"}
{
  "env": {
    "ANTHROPIC_CUSTOM_HEADERS": "x-tfy-anthropic-beta: context-management-2025-06-27,files-api-2025-04-14"
  }
}
```

<Note>
  Use exact beta feature names as documented by [Anthropic](https://docs.anthropic.com/en/api/beta-headers). Invalid beta names will return a `400` error from the Anthropic API.
</Note>

<Tip>
  `x-tfy-anthropic-beta` is a gateway-specific header that gets mapped to the `anthropic-beta` header on the outbound request to Anthropic. This avoids conflicts with Claude Code's own `anthropic-beta` header.
</Tip>

## Using Claude Code VS Code Extension

The [Claude Code VS Code extension](https://marketplace.visualstudio.com/items?itemName=Anthropic.claude-code) works seamlessly with TrueFoundry once you've configured the CLI.

<Warning>
  The VS Code extension is **not standalone**. It requires the Claude Code CLI to be installed separately and configured first.
</Warning>

**Setup Steps:**

1. Install and configure the Claude Code CLI as described [above](#integrate-truefoundry-with-claude-code)
2. Install the VS Code extension
3. Launch VS Code from your terminal to inherit the configuration:

```bash lines theme={"dark"}
code .
```

<Note>
  The extension automatically uses your CLI configuration (base URL, API keys, model aliases). No separate setup needed.
</Note>

**Important for macOS/Linux:** Launch VS Code from the terminal where you've set up Claude Code. GUI applications don't inherit shell environment variables by default.

### Model Alias Configuration

The following environment variables map Claude Code's built-in aliases to your TrueFoundry models:

* `ANTHROPIC_DEFAULT_OPUS_MODEL`: Maps `opus` alias to your TrueFoundry Opus model
* `ANTHROPIC_DEFAULT_SONNET_MODEL`: Maps `sonnet` alias to your TrueFoundry Sonnet model
* `ANTHROPIC_DEFAULT_HAIKU_MODEL`: Maps `haiku` alias to your TrueFoundry Haiku model

**Benefits:** Model aliases enable seamless integration with Claude Code's built-in commands, flexible model switching with `/model opus` or `/model sonnet`, and optimized workflows by choosing the right model for each task.

### Example Usage

```bash lines theme={"dark"}
# Use different models for different tasks
claude --model opus "Design a scalable microservices architecture"
claude --model sonnet "Implement a user authentication system"
claude --model haiku "Write a function to validate email addresses"

# Switch models during a session
/model opusplan  # Uses opus for planning, sonnet for execution
```

## Claude Agent SDK

The [Claude Agent SDK](https://docs.claude.com/en/api/agent-sdk/overview) is the successor to the Claude Code SDK, providing programmatic access to build custom AI agents with the same capabilities as Claude Code.

<Note>
  The Claude Code SDK has been deprecated. If you're using the old SDK, migrate to the Claude Agent SDK for continued support and new features.
</Note>

### Installation

```bash theme={"dark"}
pip install claude-agent-sdk
```

### Key Changes from Claude Code SDK

**What changed:** The SDK no longer reads from filesystem settings automatically. You must explicitly specify which settings to load.

**Manual setup required:** You now need to explicitly load settings for:

* Settings files (`.claude/settings.json` for project, `~/.claude/settings.json` for user)
* `CLAUDE.md` files (project memory)
* Custom slash commands from `.claude/commands/`

### TrueFoundry Integration

Your existing `.claude/settings.json` configuration works seamlessly with the Agent SDK. Simply specify `setting_sources=["project"]` to load your TrueFoundry gateway configuration:

```python theme={"dark"}
from claude_agent_sdk import query, ClaudeAgentOptions

async for message in query(
    prompt="Analyze my codebase",
    options=ClaudeAgentOptions(
        setting_sources=["project"],  # Loads .claude/settings.json
        max_turns=5,
        allowed_tools=["Read", "Grep", "Glob"]
    )
):
    if message.type == "result":
        print(message.result)
```

All your existing TrueFoundry configurations (`Anthropic Direct`, `AWS Bedrock`, `Google Vertex AI`) work identically with the Agent SDK.

## FAQs

<AccordionGroup>
  <Accordion title="How do I use Claude Opus 4.6 Fast Mode with the Gateway?">
    Claude Opus 4.6 supports a **fast mode** that provides lower-latency responses. To use it through the TrueFoundry Gateway, you need to create a virtual model with the correct additional parameters and headers.

    **Step 1: Create a Virtual Model**

    Create a virtual model in TrueFoundry that routes all traffic to `claude-opus-4-6`. In the virtual model configuration, set it to forward requests to your underlying `claude-opus-4-6` model.

    <Frame>
      <img src="https://mintcdn.com/truefoundry/_Og9Jwu88dKzwpTe/images/Screenshot-2026-04-20-at-4.55.12-PM-1.png?fit=max&auto=format&n=_Og9Jwu88dKzwpTe&q=85&s=bf3c7251cb2c7f058516ac6f59b2bb94" alt="Screenshot 2026 04 20 At 4 55 12 PM 1" width="3596" height="2006" data-path="images/Screenshot-2026-04-20-at-4.55.12-PM-1.png" />
    </Frame>

    **Step 2: Add the `speed` parameter**

    In the virtual model's additional parameters, add the following field:

    ```json theme={"dark"}
    {
      "speed": "fast"
    }
    ```

    This ensures every request sent through this virtual model includes the `speed: fast` parameter.

    <Frame>
      <img src="https://mintcdn.com/truefoundry/_Og9Jwu88dKzwpTe/images/Screenshot-2026-04-20-at-4.54.54-PM-1.png?fit=max&auto=format&n=_Og9Jwu88dKzwpTe&q=85&s=bee9f32c69c36954d0aa2bea122be759" alt="Screenshot 2026 04 20 At 4 54 54 PM 1" width="3590" height="2012" data-path="images/Screenshot-2026-04-20-at-4.54.54-PM-1.png" />
    </Frame>

    **Step 3: Add the required beta header**

    Fast mode requires the `anthropic-beta` header with the value `fast-mode-2026-02-01`. Add this as an additional header in the virtual model configuration:

    ```text theme={"dark"}
    anthropic-beta: fast-mode-2026-02-01
    ```

    <Frame>
      <img src="https://mintcdn.com/truefoundry/_Og9Jwu88dKzwpTe/images/Screenshot-2026-04-20-at-4.54.45-PM-1.png?fit=max&auto=format&n=_Og9Jwu88dKzwpTe&q=85&s=fe9732c2cc9f8e2bd9b9828de6b7b4f5" alt="Screenshot 2026 04 20 At 4 54 45 PM 1" width="3600" height="2006" data-path="images/Screenshot-2026-04-20-at-4.54.45-PM-1.png" />
    </Frame>

    **Step 4: Use the virtual model in Claude Code**

    Once the virtual model is created, use its model name in your `settings.json`. Replace the `ANTHROPIC_MODEL` and the relevant alias (e.g. `ANTHROPIC_DEFAULT_OPUS_MODEL`) with the virtual model name:

    ```json theme={"dark"}
    {
      "env": {
        "ANTHROPIC_BASE_URL": "{GATEWAY_BASE_URL}",
        "ANTHROPIC_AUTH_TOKEN": "your-truefoundry-api-key",
        "ANTHROPIC_MODEL": "your-fast-opus-virtual-model-name",
        "ANTHROPIC_DEFAULT_OPUS_MODEL": "your-fast-opus-virtual-model-name",
        "ANTHROPIC_DEFAULT_SONNET_MODEL": "your-sonnet-model-name",
        "ANTHROPIC_DEFAULT_HAIKU_MODEL": "your-haiku-model-name",
        "CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS": "1"
      }
    }
    ```

    When selecting a model in Claude Code (via `/model` or the `--model` flag), choose the virtual model name you configured above. All requests will automatically include the fast mode parameter and the required beta header.
  </Accordion>

  <Accordion title="How do I control Fast Mode on the Pro plan?">
    Users can enable fast mode in Claude Code — either per-request with `/fast` or globally in their config so all requests use fast mode. Since fast mode uses significantly more credits, you may want to control when it's used. Create two `virtual models` — one that disables fast mode and one that allows it — then hand out the appropriate model ID to users.

    **Virtual Model 1: Fast mode disabled (strip `speed`)**

    Create a virtual model that routes to `claude-opus-4-6` and configure it to delete the `speed` field from the request body. This ensures requests always go through as standard, non-fast requests — even if the user has fast mode enabled.

    <Frame>
      <img src="https://mintcdn.com/truefoundry/e1MT0X46pifx91tO/images/Screenshot-2026-04-20-at-5.27.02-PM.png?fit=max&auto=format&n=e1MT0X46pifx91tO&q=85&s=ad7ac1fc449f228f7a1dfc8501cbea04" alt="Screenshot 2026 04 20 At 5 27 02 PM" width="3592" height="2004" data-path="images/Screenshot-2026-04-20-at-5.27.02-PM.png" />
    </Frame>

    **Virtual Model 2: Fast mode allowed**

    Create a second virtual model that routes to `claude-opus-4-6` with no additional parameters or headers. This model passes requests through as-is, so when a user enables fast mode, it will reach Anthropic.

    <Frame>
      <img src="https://mintcdn.com/truefoundry/e1MT0X46pifx91tO/images/Screenshot-2026-04-20-at-5.27.28-PM.png?fit=max&auto=format&n=e1MT0X46pifx91tO&q=85&s=5597c55fef5bbabf6a05e93f3cd07c8c" alt="Screenshot 2026 04 20 At 5 27 28 PM" width="3596" height="1998" data-path="images/Screenshot-2026-04-20-at-5.27.28-PM.png" />
    </Frame>

    Then use the appropriate virtual model name in `settings.json`:

    ```json theme={"dark"}
    {
      "env": {
        "ANTHROPIC_BASE_URL": "{GATEWAY_BASE_URL}",
        "ANTHROPIC_AUTH_TOKEN": "your-truefoundry-api-key",
        "ANTHROPIC_MODEL": "your-virtual-model-name",
        "ANTHROPIC_DEFAULT_OPUS_MODEL": "your-virtual-model-name",
        "ANTHROPIC_DEFAULT_SONNET_MODEL": "your-sonnet-model-name",
        "ANTHROPIC_DEFAULT_HAIKU_MODEL": "your-haiku-model-name",
        "CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS": "1"
      }
    }
    ```
  </Accordion>
</AccordionGroup>
