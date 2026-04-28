> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Team Provisioning

> Control which workspaces and users can access MCP servers and their capabilities.

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

## Next Steps

<CardGroup cols={2}>
  <Card title="Tool Provisioning" icon="wrench" href="/product/mcp-gateway/tool-provisioning">
    Enable or disable specific tools at the organization level.
  </Card>

  <Card title="Guardrails" icon="shield" href="/product/mcp-gateway/guardrails">
    Apply rate limits, content filtering, and approval workflows.
  </Card>
</CardGroup>
