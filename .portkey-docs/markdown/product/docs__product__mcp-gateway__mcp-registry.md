> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# MCP Registry

> Add, manage, and govern MCP servers across your organization.

The MCP Registry is the central catalog of all MCP server integrations in your organization. Organization admins add servers here, configure authentication, and control which teams can access them.

<Frame>
  <img src="https://mintcdn.com/portkey-docs/SYbJLnYR8wpty2Fv/images/mcp-gateway/mcp-registry-list.png?fit=max&auto=format&n=SYbJLnYR8wpty2Fv&q=85&s=d772598e4180e52530a580c755225c77" alt="MCP Registry showing connected servers" width="2932" height="1478" data-path="images/mcp-gateway/mcp-registry-list.png" />
</Frame>

<Info>
  **Developer note:** Most developers won't add servers here. After an org admin provisions a server, use it from the **workspace MCP page** to copy the connection URL and test tools.
</Info>

The MCP Registry supports:

* Adding MCP servers (from catalog or custom URL)
* Configuring authentication for each server
* Viewing connection status and health metrics
* Controlling which teams access which servers
* Managing capabilities (tools, resources, prompts) per server

<Card title="Quickstart" icon="rocket" href="/product/mcp-gateway/quickstart">
  New to MCP Gateway? Add your first server in 5 minutes.
</Card>

***

## Adding Servers

Click **MCP Registry** in the sidebar, then click **Add MCP Server**.

| Field       | Description                                              |
| ----------- | -------------------------------------------------------- |
| Name        | Display name for the server                              |
| Slug        | URL-friendly identifier (e.g., `linear`, `internal-api`) |
| Server URL  | MCP endpoint (e.g., `https://mcp.linear.app/mcp`)        |
| Server Type | `HTTP` for remote servers                                |
| Auth Type   | Authentication method (see below)                        |

<Note>
  Portkey supports remote MCP servers over HTTP. Local STDIO servers need to be exposed as HTTP endpoints first.
</Note>

***

## Authentication

<Info>
  This section covers how Portkey authenticates to **upstream MCP servers**. For how your agents and MCP clients authenticate to Portkey, see [Gateway Authentication](/product/mcp-gateway/authentication).
</Info>

When adding a server, choose the auth method based on what the MCP server supports:

| Auth Type              | When to Use                                                     |
| ---------------------- | --------------------------------------------------------------- |
| **None**               | Public servers with no authentication required                  |
| **OAuth 2.1**          | Servers with user-level OAuth. Each user gets their own tokens. |
| **Client Credentials** | Servers using OAuth client credentials. Shared across users.    |
| **Headers**            | Servers using API keys or static tokens.                        |

**Quick summary:**

* **OAuth (user-level)**: Each user authorizes their own access. Use for servers with user-specific data (Linear, GitHub, Slack).
* **Client Credentials**: Shared OAuth tokens. Use for service accounts or shared resources.
* **Headers**: Static API keys sent with every request. Use for simple token-based auth.

  <Card title="Authentication Deep Dive" icon="lock" href="/product/mcp-gateway/authentication">
    Gateway authentication, external OAuth, identity forwarding, JWT validation, and more.
  </Card>

***

## Access Control

Access control for MCP servers happens at two levels.

## Organization-level (MCP Registry)

<Info>Requires **Org Admin** or **Org Owner** role.</Info>

Control which workspaces can access an MCP server:

1. Click **MCP Registry** in the sidebar
2. Click on an MCP server
3. Open the **Access Control** tab
4. Toggle workspaces on/off

<Frame>
  <img src="https://mintcdn.com/portkey-docs/SYbJLnYR8wpty2Fv/images/mcp-gateway/provision-server-access.png?fit=max&auto=format&n=SYbJLnYR8wpty2Fv&q=85&s=c068985d563f3b2fec8a439098a8713f" alt="Workspace access control" width="2570" height="1382" data-path="images/mcp-gateway/provision-server-access.png" />
</Frame>

Enable **Automatically provision this integration for new workspaces** to include future workspaces by default.

## Workspace-level (MCP Servers)

<Info>Requires **Workspace Manager** or **Workspace Admin** role.</Info>

Control which users in your workspace can access an MCP server:

1. Go to **MCP** page in your workspace sidebar
2. Click on an MCP server
3. Use the **User Access** tab to toggle user access on/off

<Frame>
  <img src="https://mintcdn.com/portkey-docs/SYbJLnYR8wpty2Fv/images/mcp-gateway/workspave-user-access.png?fit=max&auto=format&n=SYbJLnYR8wpty2Fv&q=85&s=71714d5b233dcb5fc767256fef2a4f75" alt="User access control" width="2558" height="1466" data-path="images/mcp-gateway/workspave-user-access.png" />
</Frame>

***

## Tool Provisioning

The **Capabilities** tab lists all tools, resources, and prompts exposed by an MCP server. Toggle them on or off, or use **Enable All** / **Disable All** for bulk changes.

<Frame>
  <img src="https://mintcdn.com/portkey-docs/SYbJLnYR8wpty2Fv/images/mcp-gateway/tools.png?fit=max&auto=format&n=SYbJLnYR8wpty2Fv&q=85&s=9509e532ac53953f1f908a4a407f71a2" alt="Capabilities tab" width="2568" height="1364" data-path="images/mcp-gateway/tools.png" />
</Frame>

Tool provisioning can be done at two levels:

| Location                                 | Who                       | Scope             |
| ---------------------------------------- | ------------------------- | ----------------- |
| **MCP Registry** → Server → Capabilities | Org Admins/Owners         | Organization-wide |
| **MCP Servers** → Server → Capabilities  | Workspace Managers/Admins | Workspace-only    |

Disabled capabilities are hidden and return errors if called directly. Use this to block dangerous operations, untested tools, or deprecated functionality.

***

## Server Overview

Once added, each server has a detail page with four tabs: **Overview**, **Capabilities**, **Setup & Code**, and **Access Control & Limits**.

<Frame>
  <img src="https://mintcdn.com/portkey-docs/SYbJLnYR8wpty2Fv/images/mcp-gateway/server-overview.png?fit=max&auto=format&n=SYbJLnYR8wpty2Fv&q=85&s=82ea2a5a081050b0c5ef357f847dd507" alt="Server overview page" width="2936" height="1476" data-path="images/mcp-gateway/server-overview.png" />
</Frame>

**Overview tab** shows:

* **Server Status**: Health score, uptime, average response time, total requests
* **Security & Authentication**: Auth type, grant type, transport protocol
* **Configuration**: Server URL, version, connection details

***

## Next Steps

<CardGroup cols={2}>
  <Card title="Add Internal MCP Servers" icon="server" href="/product/mcp-gateway/internal-mcp-servers">
    Connect your own MCP servers with enterprise auth, access control, and logging.
  </Card>

  <Card title="Integrations" icon="plug" href="/product/mcp-gateway/integrations">
    Setup guides for Claude Desktop, Cursor, Python SDK, TypeScript SDK, and more.
  </Card>

  <Card title="Observability" icon="chart-line" href="/product/mcp-gateway/observability">
    Monitor MCP traffic, debug issues, track usage.
  </Card>

  <Card title="Guardrails" icon="shield" href="/product/mcp-gateway/guardrails">
    Apply rate limits, content filtering, and approval workflows.
  </Card>
</CardGroup>
