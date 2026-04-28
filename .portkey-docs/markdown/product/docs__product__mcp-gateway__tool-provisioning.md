> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Tool Provisioning

> Control which tools, resources, and prompts are available to your organization and workspaces.

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

## Next Steps

<CardGroup cols={2}>
  <Card title="Team Provisioning" icon="users" href="/product/mcp-gateway/access-control">
    Control which teams and users can access MCP servers.
  </Card>

  <Card title="Guardrails" icon="shield" href="/product/mcp-gateway/guardrails">
    Apply rate limits, content filtering, and approval workflows.
  </Card>
</CardGroup>
