> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Gemini CLI

> This guide provides instructions for integrating Gemini CLI with TrueFoundry's AI Gateway

## What is Gemini CLI?

The [Gemini command line interface (CLI)](https://github.com/google-gemini/gemini-cli) is Google's official AI coding assistant that runs directly in your terminal. It understands your codebase and helps you code faster through natural language commands.

<Note>
  **Rate and budget limits:** TrueFoundry AI Gateway now supports Rate and Budget Limits for Gemini CLI. You can set configurations to cap usage and costs and avoid unexpected overuse. [Learn More](/docs/ai-gateway/budgetlimiting)
</Note>

## Prerequisites

Before integrating Gemini CLI with TrueFoundry, ensure you have:

1. **TrueFoundry Account**: A [TrueFoundry account](https://www.truefoundry.com/register) with at least one Google Gemini or Vertex AI Provider Account configured and a Personal Access Token ([Generate here](/docs/ai-gateway/authentication))
2. **Gemini CLI Installation**: Install Gemini CLI by running:
   ```bash lines theme={"dark"}
   npm install -g @google/gemini-cli
   ```

## Setup Configuration

### 1. Create Provider Account

Create a provider account for either Google Gemini or Vertex AI:

* **Google Gemini**: Follow the [Google Gemini setup guide](/docs/ai-gateway/google-gemini)
* **Vertex AI**: Follow the [Vertex AI setup guide](/docs/ai-gateway/google-vertex)

<Note>
  **Provider Support**: Both Google Gemini and Vertex AI providers are supported. The gateway automatically handles the necessary transformations for Vertex AI when using the Gemini CLI.
</Note>

### 2. Get Configuration Details

Get the `base URL` and `Provider Name` from your TrueFoundry AI Gateway playground using the unified code snippet:

<img src="https://mintcdn.com/truefoundry/dnMutGJug9ieMa1o/images/Screenshot2025-09-24at7.39.04PM.png?fit=max&auto=format&n=dnMutGJug9ieMa1o&q=85&s=8cd01b779e7dc3853a42bcf2d4b8a097" alt="Screenshot2025 09 24at7 39 04PM Pn" width="3600" height="2006" data-path="images/Screenshot2025-09-24at7.39.04PM.png" />

### 3. Configure Environment Variables

Choose the configuration that matches your provider. **Google Gemini** and **Vertex AI** use different environment variables and auth mechanisms.

<Accordion title="Google Gemini">
  ```bash lines theme={"dark"}
  export GOOGLE_GEMINI_BASE_URL="{GATEWAY_BASE_URL}/gemini/{providerAccountName}/proxy"
  export GEMINI_API_KEY="your-truefoundry-api-key"
  ```
</Accordion>

<Accordion title="Vertex AI">
  ```bash lines theme={"dark"}
  export GOOGLE_VERTEX_BASE_URL="{GATEWAY_BASE_URL}/gemini/{providerAccountName}/proxy"
  export GOOGLE_API_KEY="your-truefoundry-api-key"
  export GEMINI_API_KEY_AUTH_MECHANISM="bearer"
  ```
</Accordion>

**Replace the placeholders:**

| Placeholder                | Description                                                                                            |
| -------------------------- | ------------------------------------------------------------------------------------------------------ |
| `GATEWAY_BASE_URL`         | Your TrueFoundry AI Gateway Base URL ([how to find it](/docs/ai-gateway/quick-start#gateway-base-url)) |
| `providerAccountName`      | Your Provider Account name (either `Google Gemini or Vertex AI`)                                       |
| `your-truefoundry-api-key` | Your TrueFoundry Personal Access Token                                                                 |

<Note>
  **Vertex AI users:** You must use the **Vertex AI** configuration above (with `GOOGLE_VERTEX_BASE_URL`, `GOOGLE_API_KEY`, and `GEMINI_API_KEY_AUTH_MECHANISM="bearer"`). The older Gemini-style config is no longer supported for Vertex AI with Gemini CLI due to upstream changes affecting models such as **gemini-3.1-pro**. If you were previously using Vertex AI with the Gemini config, see [Migrating from Gemini config to Vertex AI config](#migrating-from-gemini-config-to-vertex-ai-config) below.
</Note>

## Migrating from Gemini config to Vertex AI config

If you were using **Vertex AI** with the previous Gemini-style configuration (`GOOGLE_GEMINI_BASE_URL` + `GEMINI_API_KEY`), you need to switch to the **Vertex AI–specific configuration**. Gemini CLI has made changes that make Vertex AI Gemini models (when used via Gemini CLI in gemini mode) **unusable for gemini-3.1-pro** unless you use the new Vertex AI config and auth type.

<Warning>
  **Action required for Vertex AI users:** Migrate to the new Vertex AI configuration and re-select **Vertex AI** as the auth type in Gemini CLI. Until you do, Vertex AI models such as **gemini-3.1-pro** may not work correctly with Gemini CLI.
</Warning>

Follow **one** of the migration options below.

### Option 1: Reset auth via CLI (/auth and logout)

1. Set the [Vertex AI environment variables](#3-configure-environment-variables) (see the **Vertex AI** accordion above) in your shell profile or environment.
2. Start the CLI with `gemini`, then run the `/auth` command.
3. Use the logout option in the auth flow to clear the current auth type.
4. When prompted, select **Vertex AI** as the auth type.

### Option 2: Reset auth via settings file

1. Open the Gemini CLI settings file (e.g. `~/.gemini/settings.json` .
2. Delete the `selectedType` field from the auth object so that the CLI no longer uses the old auth selection.
3. Set the [Vertex AI environment variables](#3-configure-environment-variables) (see the **Vertex AI** accordion above) in your shell profile or environment.
4. Start the CLI with `gemini`. When prompted for authentication, choose **Vertex AI**.

<Note>
  After removing `selectedType`, the next time you run `gemini` you will be asked to pick an auth type again. Select **Vertex AI** to use the new configuration.
</Note>

## Using Gemini CLI

1. **Start the CLI:**
   ```bash lines theme={"dark"}
   gemini
   ```
2. **Setup Authentication:**
   * **Google Gemini users:** If you see authentication options, choose **Use Gemini API Key**. If not prompted, type `/auth` and select **Use Gemini API Key**.
   * **Vertex AI users:** Choose **Vertex AI** (or the Vertex AI auth type). If not prompted, type `/auth` and select **Vertex AI**. Ensure you have set the [Vertex AI environment variables](#3-configure-environment-variables) (see the **Vertex AI** accordion above) before starting.

<img src="https://mintcdn.com/truefoundry/dnMutGJug9ieMa1o/images/Screenshot2025-09-24at7.46.24PM.png?fit=max&auto=format&n=dnMutGJug9ieMa1o&q=85&s=d2674d28cf521581d9b66579657d34e3" alt="Screenshot2025 09 24at7 46 24PM Pn" width="2682" height="1260" data-path="images/Screenshot2025-09-24at7.46.24PM.png" />

3. **Start using Gemini:**
   ```plain theme={"dark"}
   # Ask about your code
   > Explain this function
   # Get code suggestions
   > Write a function to validate email addresses
   # Debug issues
   > Why is this giving me a syntax error?
   ```

You can now use Gemini CLI with all requests going through your TrueFoundry AI Gateway.
