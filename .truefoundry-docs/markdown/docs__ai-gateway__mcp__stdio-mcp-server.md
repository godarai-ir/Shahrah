> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Hosted Stdio-based MCP Server

> Run a hosted MCP server via stdio and expose it through the gateway.

## What is Hosted Stdio-based MCP Server?

Run a hosted MCP server via stdio and expose it through the gateway.

<Tip>
  Recommended for local development in IDEs only.
</Tip>

Hosted stdio MCP servers are now available as first-class MCP servers with built-in **per-user** or **global** credential support via environment variables.

A **hosted stdio-based MCP server** speaks the Model Context Protocol over standard input and standard output (typical for CLI-style servers started with `npx`, `uvx`, or your own binary). TrueFoundry registers this pattern as a first-class MCP server: you provide the **command**, **arguments**, and **environment variables** for configuration and credentials, and the MCP Gateway runs and manages the process so clients use the same gateway URL as for [remote MCP servers](/docs/ai-gateway/mcp/mcp-server-getting-started).

This is useful when you already rely on a stdio MCP package or script and want centralized access, auth, guardrails, and observability without standing up a separate HTTP wrapper yourself.

## How It Works

| Element                   | Role                                                                                                                                                                                                                                                                                                |
| ------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Command**               | Executable the gateway runs to start the MCP server (for example `npx` or `uvx`)                                                                                                                                                                                                                    |
| **Arguments**             | Arguments passed to that executable (for example package name and flags)                                                                                                                                                                                                                            |
| **Environment variables** | Injected into the server process; use shared values for **global** credentials, or **per-user** values that include **one** templatized credential placeholder per variable—substituted from [Auth Overrides](/docs/ai-gateway/mcp/mcp-server-auth-overrides), same model as Individual Credentials |
| **Collaborators**         | Same access model as other MCP servers ([Getting Started](/docs/ai-gateway/mcp/mcp-server-getting-started))                                                                                                                                                                                         |

You configure **environment variables** on the server (names and values). For **per-user** credentials, each value may include **one** templatized placeholder that the gateway resolves from that caller’s [Auth Overrides](/docs/ai-gateway/mcp/mcp-server-auth-overrides)—the same substitution behavior as Individual Credentials on remote MCP servers (see [MCP Gateway Auth & Security](/docs/ai-gateway/mcp/mcp-gateway-auth-security)). Arbitrary or additional template syntax is not supported.

## Create a Hosted Stdio-based MCP Server

<Steps>
  <Step title="Navigate to MCP Servers">
    Open the **MCP Servers** section in the TrueFoundry AI Gateway and choose **Add Server**.
  </Step>

  <Step title="Choose hosted stdio registration">
    Select **Hosted stdio-based MCP server**. This path configures a **command**, **arguments**, and **environment variables** instead of a remote MCP URL.
  </Step>

  <Step title="Import from editor-style JSON (optional)">
    Paste MCP configuration JSON as a **single-entry** `mcpServers` map—the same shape as in Cursor or VS Code MCP config: one server id with `command`, `args`, and optional `env`. Only **one** server entry is allowed so the control plane can build one manifest.
  </Step>

  <Step title="Set command and arguments">
    Configure the process the gateway should run:

    | Field         | Description                                               |
    | ------------- | --------------------------------------------------------- |
    | **Command**   | Executable for the MCP server process (for example `npx`) |
    | **Arguments** | List of arguments passed to the command                   |
  </Step>

  <Step title="Configure environment variables and access">
    Add environment variables required by the server (API keys, feature flags, and so on). Choose **global** credentials when everyone shares the same downstream access, or **per-user** when a value uses **one** templatized placeholder filled from each caller’s [Auth Overrides](/docs/ai-gateway/mcp/mcp-server-auth-overrides). Add **collaborators** and roles as for any MCP server.
  </Step>

  <Step title="Create and test">
    Save the server, then use the **Playground** to list and invoke tools, and open **How To Use** for IDE and agent connection snippets.
  </Step>
</Steps>

## Example manifests

The following YAML manifests match the shape used for `type: mcp-server/stdio`. Both examples run [`mcp-remote`](https://www.npmjs.com/package/mcp-remote) under `npx` to bridge a remote MCP HTTPS URL over stdio; credentials are passed through **`auth_data.env`**.

<Note>
  Replace the collaborator `subject` with a real user in your account. Do not commit real API keys—use a secret store or placeholder for **global** env values. For **per-user** env values, use **one** templatized placeholder per value and configure each caller in [Auth Overrides](/docs/ai-gateway/mcp/mcp-server-auth-overrides).
</Note>

<Tabs>
  <Tab title="Per-user (Linear via mcp-remote)">
    ```yaml theme={"dark"}
    name: linear-global
    description: Stdio-based MCP server
    collaborators:
      - role_id: mcp-server-manager
        subject: user:you@example.com
    type: mcp-server/stdio
    command: npx
    args:
      - "-y"
      - mcp-remote
      - https://mcp.linear.app/mcp
      - "--header"
      - "Authorization: Bearer ${AUTH_TOKEN}"
    auth_data:
      type: env
      auth_level: per_user
      env:
        AUTH_TOKEN: "{{API_KEY}}"
    ```

    `mcp-remote` reads the bearer token from the process environment (`${AUTH_TOKEN}`). The gateway injects each user’s secret into `AUTH_TOKEN` when `AUTH_TOKEN` is configured with a templatized value and **per\_user** auth.
  </Tab>

  <Tab title="Global (Exa via mcp-remote)">
    ```yaml theme={"dark"}
    name: exa-user-global-auth
    description: Stdio-based MCP server
    collaborators:
      - role_id: mcp-server-manager
        subject: user:you@example.com
    type: mcp-server/stdio
    command: npx
    args:
      - "-y"
      - mcp-remote
      - https://mcp.exa.ai/mcp
    auth_data:
      type: env
      auth_level: global
      env:
        EXA_API_KEY: your-exa-api-key
    ```

    The same **global** API key is used for every caller. Prefer **per\_user** when each user should use their own Exa key.
  </Tab>
</Tabs>

## Frequently Asked Questions

<AccordionGroup>
  <Accordion title="How is this different from a remote MCP server?">
    Remote servers point at an **HTTP (or streamable HTTP) MCP URL**. Stdio servers describe a **local process** (command + args) the gateway runs. Pick remote when you already have a URL; pick stdio when your integration is packaged as a stdio CLI.
  </Accordion>

  <Accordion title="Can I combine a stdio MCP server with a Virtual MCP server?">
    Yes. After registration, you can include its tools in a [Virtual MCP Server](/docs/ai-gateway/mcp/virtual-mcp-server) alongside tools from OpenAPI or remote servers.
  </Accordion>

  <Accordion title="What templatized values are supported in environment variables?">
    For **per-user** credentials, **one** templatized placeholder per environment variable value is supported; the gateway substitutes it using [Auth Overrides](/docs/ai-gateway/mcp/mcp-server-auth-overrides), consistent with Individual Credentials elsewhere. Multiple placeholders or other template syntax in the same value are not supported.
  </Accordion>

  <Accordion title="Do guardrails and analytics apply to stdio MCP servers?">
    Yes. Traffic through the MCP Gateway is subject to the same controls and observability patterns as other registered MCP server types, consistent with [OpenAPI-backed MCP servers](/docs/ai-gateway/mcp/openapi-mcp-server) and remote servers.
  </Accordion>
</AccordionGroup>

## Best Practices

<CardGroup cols={2}>
  <Card title="Pin versions" icon="tag">
    Prefer explicit package versions in `args` (for example pinned `npx` or `uvx` targets) so tool behavior does not change unexpectedly when upstream releases update.
  </Card>

  <Card title="Least-privilege environment variables" icon="shield">
    Pass only the environment variables the process needs, and use per-user auth when each caller should use their own downstream credentials.
  </Card>

  <Card title="Align with local dev" icon="code">
    Keep command and args aligned with a working local MCP config so debugging stays straightforward.
  </Card>
</CardGroup>
