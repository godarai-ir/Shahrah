> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# GitHub Copilot

> Route GitHub Copilot requests through TrueFoundry AI Gateway for observability, governance, and custom model access.

This guide provides instructions for integrating [GitHub Copilot](https://github.com/features/copilot) with the TrueFoundry AI Gateway.

## What is GitHub Copilot?

GitHub Copilot is an AI-powered coding assistant that integrates directly into your editor. It provides intelligent code completions, chat-based coding help, and agent-driven workflows across VS Code, JetBrains IDEs, Visual Studio, Eclipse, and Xcode.

### Key Features

1. **Inline Code Suggestions**: Context-aware code completions and multi-line suggestions as you type
2. **Copilot Chat**: Conversational AI assistant for code explanation, generation, debugging, and refactoring
3. **Agent Mode**: Agentic workflows that can autonomously plan, edit files, and run terminal commands to complete complex tasks
4. **Bring Your Own Key (BYOK)**: Use custom models from any OpenAI-compatible provider — including TrueFoundry AI Gateway

## Prerequisites

Before integrating GitHub Copilot with TrueFoundry, ensure you have:

1. **TrueFoundry Account**: Create a [TrueFoundry account](https://www.truefoundry.com/register) with at least one model provider configured and generate a Personal Access Token by following the instructions in [Generating Tokens](/docs/ai-gateway/authentication). For a quick setup guide, see our [Gateway Quick Start](https://docs.truefoundry.com/gateway/quick-start)
2. **GitHub Copilot Subscription**: An active [GitHub Copilot](https://github.com/features/copilot) plan (Free, Pro, Business, or Enterprise)
3. **VS Code**: [Visual Studio Code](https://code.visualstudio.com/) with the GitHub Copilot extension installed

<Note>
  The BYOK feature for custom OpenAI-compatible models is currently available natively in **VS Code Insiders**. For stable VS Code, you can use the community extension [OAI Compatible Provider for Copilot](https://marketplace.visualstudio.com/items?itemName=johnny-zhao.oai-compatible-copilot). See the [Alternative: Stable VS Code](#alternative-stable-vs-code-with-community-extension) section below.
</Note>

## Integration Guide (VS Code Insiders)

### 1. Get Configuration Details

Get the `base URL` and `model name` from your TrueFoundry AI Gateway playground using the unified code snippet:

<Frame caption="Get Base URL and Model Name from Unified Code Snippet">
  <img src="https://mintcdn.com/truefoundry/n3EuZuJ0K8wBFp1G/images/new-code-snippet.png?fit=max&auto=format&n=n3EuZuJ0K8wBFp1G&q=85&s=3634c2dc8c3565fd77ab896d3fd07ed9" alt="TrueFoundry playground showing unified code snippet with base URL and model name" width="2940" height="1664" data-path="images/new-code-snippet.png" />
</Frame>

### 2. Add TrueFoundry as a Model Provider

<Steps>
  <Step title="Open Manage Models">
    Open the Copilot Chat panel (`Ctrl+Alt+I` on Windows/Linux, `Ctrl+Cmd+I` on macOS), click the **model dropdown** at the top, and select **Manage Models...**.
  </Step>

  <Step title="Add OpenAI Compatible Provider">
    In the Language Models editor, click **Add Models** and select **OpenAI Compatible** from the list of providers.
  </Step>

  <Step title="Enter TrueFoundry Gateway Details">
    When prompted, enter the following details:

    * **Base URL**: Your TrueFoundry Gateway URL (e.g., `{GATEWAY_BASE_URL}`)
    * **API Key**: Your TrueFoundry Personal Access Token

    After entering the API key, add the models you want to use. Enter the fully qualified model name from TrueFoundry in `provider-name/model-name` format (e.g., `openai-main/gpt-4o`).
  </Step>

  <Step title="Enable Models">
    Check the models you want to make available in the Copilot Chat model picker, then click **OK**.
  </Step>
</Steps>

### 3. Use TrueFoundry Models in Copilot Chat

Your TrueFoundry models now appear in the model dropdown in Copilot Chat. Select any of them to route requests through the TrueFoundry AI Gateway.

```plain theme={"dark"}
# Ask about your code
> Explain this function and suggest improvements

# Generate code
> Write a REST API endpoint for user authentication

# Debug issues
> Why is this test failing? Suggest a fix
```

<Note>
  For a model to work in **Agent Mode**, it must support tool calling. Most large models (GPT-4o, Claude Sonnet, etc.) support this. If a model doesn't support tool calling, it will only be available in Ask and Edit modes.
</Note>

### Alternative: Settings JSON Configuration

You can also configure custom OpenAI-compatible models directly in your VS Code `settings.json` using the experimental `github.copilot.chat.customOAIModels` setting:

```json lines theme={"dark"}
{
  "github.copilot.chat.customOAIModels": [
    {
      "id": "openai-main/gpt-4o",
      "name": "GPT-4o via TrueFoundry",
      "baseUrl": "{GATEWAY_BASE_URL}",
      "apiKeySettingId": "TrueFoundry"
    }
  ]
}
```

**Replace the placeholders:**

* `{GATEWAY_BASE_URL}` → Your TrueFoundry AI Gateway Base URL ([how to find it](/docs/ai-gateway/quick-start#gateway-base-url))
* `openai-main/gpt-4o` → Your desired model in `provider-name/model-name` format

When you open Copilot Chat after adding this setting, VS Code will prompt you to enter the API key for the "TrueFoundry" provider. Enter your TrueFoundry Personal Access Token.

<Warning>
  The `github.copilot.chat.customOAIModels` setting is currently **experimental** and only functional in **VS Code Insiders**. The setting exists in stable VS Code but has no effect.
</Warning>

## Alternative: Stable VS Code with Community Extension

If you're using stable VS Code (not Insiders), you can use the community extension **OAI Compatible Provider for Copilot** to connect to TrueFoundry:

<Steps>
  <Step title="Install the Extension">
    Search for **"OAI Compatible Provider for Copilot"** (by johnny-zhao) in the VS Code Extensions marketplace (`Ctrl+Shift+X`) and install it.
  </Step>

  <Step title="Configure the Extension">
    Add the following to your `settings.json`:

    ```json lines theme={"dark"}
    {
      "oaicopilot.baseUrl": "{GATEWAY_BASE_URL}",
      "oaicopilot.models": [
        {
          "id": "openai-main/gpt-4o",
          "configId": "gpt-4o-tfy",
          "owned_by": "truefoundry"
        }
      ]
    }
    ```

    Replace `{GATEWAY_BASE_URL}` with your TrueFoundry AI Gateway Base URL and update the model `id` to match your TrueFoundry model.
  </Step>

  <Step title="Activate the Model">
    Open Command Palette (`Ctrl+Shift+P`), run **Chat: Manage Language Models**, select **OAI Compatible**, and enter your TrueFoundry Personal Access Token as the API key. Check the models you want to enable.
  </Step>
</Steps>

## Enterprise / Organization BYOK

For GitHub Copilot Business and Enterprise users, organization and enterprise admins can add TrueFoundry as an **OpenAI-compatible provider** at the org or enterprise level. This makes TrueFoundry-routed models available to all team members through the Copilot Chat model picker.

<Steps>
  <Step title="Navigate to Copilot Settings">
    Go to your GitHub organization or enterprise settings, then navigate to **Copilot** → **AI controls** → **Configure allowed models** → **Custom models** tab.
  </Step>

  <Step title="Add TrueFoundry API Key">
    Click **Add API key** and configure:

    * **Provider**: Select **OpenAI-compatible providers**
    * **Name**: Enter a descriptive name (e.g., "TrueFoundry AI Gateway")
    * **API Key**: Enter your TrueFoundry Personal Access Token
    * **Available models**: Add your desired models using the fully qualified `provider-name/model-name` format
  </Step>

  <Step title="Configure Access">
    Choose which organizations can access the models and save the configuration. The models will appear in the Copilot Chat model picker for all enabled users, listed under your enterprise or organization name.
  </Step>
</Steps>

<Note>
  With Enterprise BYOK, usage is billed directly through your TrueFoundry account and does not count against GitHub Copilot's built-in request quotas. This lets teams leverage existing contracts and credits.
</Note>

## Virtual Model Configuration (Optional)

If your models or Copilot setup requires standard model names (e.g., `gpt-4o` instead of `openai-main/gpt-4o`), create a Virtual Model in TrueFoundry to map standard names to fully qualified model names:

1. Navigate to **AI Gateway → Models → Virtual Model** in the TrueFoundry dashboard.

<Frame>
  <img src="https://mintcdn.com/truefoundry/9sjb7wvpsYV7RE1W/images/Screenshot2026-02-27at4.38.04PM-1.png?fit=max&auto=format&n=9sjb7wvpsYV7RE1W&q=85&s=5815438d18a9ecd74e2c183b7aa8d9a1" alt="TrueFoundry Virtual Models dashboard" className="block dark:hidden" width="3598" height="2008" data-path="images/Screenshot2026-02-27at4.38.04PM-1.png" />

  <img src="https://mintcdn.com/truefoundry/uL7y0Q9ywygbGxPN/images/Screenshot2026-02-27at4.38.04PM.png?fit=max&auto=format&n=uL7y0Q9ywygbGxPN&q=85&s=84550e1c21cc38734bfbabe680a189f1" alt="TrueFoundry Virtual Models dashboard" className="hidden dark:block" width="3598" height="2008" data-path="images/Screenshot2026-02-27at4.38.04PM.png" />
</Frame>

2. Create a new Virtual Model Provider Group with a name (e.g., `copilot-vm`) and configure collaborators for access control.

3. Add your target model (e.g., `openai-main/gpt-4o`) under the provider group. Set the Virtual Model name to the standard model name (e.g., `gpt-4o`), so requests for `gpt-4o` are automatically routed to `openai-main/gpt-4o` through the TrueFoundry Gateway.

For more details, see the [Virtual Models documentation](/docs/ai-gateway/virtual-model).
