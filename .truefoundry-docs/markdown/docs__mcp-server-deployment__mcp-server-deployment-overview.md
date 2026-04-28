> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Overview

> Learn how to deploy MCP servers on TrueFoundry. Covers HTTP and stdio-based servers from source code, npx/uvx packages, and custom implementations.

MCP (Model Context Protocol) servers expose data and capabilities to LLMs through a standardized interface. This guide helps you deploy MCP servers on TrueFoundry as services that can be accessed by the AI Gateway or other applications. MCP servers can be either HTTP/SSE-based or stdio-based.

**HTTP/SSE-based servers:** Expose HTTP endpoints directly. Deploy without additional wrappers.

**Stdio-based servers:** Communicate via stdin/stdout. Wrap with `mcp-proxy` to convert to HTTP.

## Choose Your Deployment Path

### Deploy from Code

**Use this when:**

* You have MCP server source code (GitHub repository or local)
* You've written your own MCP server
* You need to customize the server

**Supports:** HTTP/SSE-based servers (deploy directly) and stdio-based servers (wrapped with mcp-proxy)

[Deploy MCP Server from Code →](/docs/mcp-server-deployment/deploy-mcp-server-from-code)

### Deploy from npx/uvx

**Use this when:**

* You have instructions for using an MCP server in VSCode/Cursor/Claude
* The MCP server is available as an npm or Python package
* You don't need to modify the server code

**Supports:** npm packages (via `npx`) and Python packages (via `uvx`)

[Deploy MCP Server from npx/uvx →](/docs/mcp-server-deployment/deploy-mcp-server-from-npx-uvx)

## Prerequisites

* A TrueFoundry workspace ([Create workspace](/docs/key-concepts#creating-a-workspace) if needed)
* Access to deploy services
* Your MCP server code or package name

<Note>
  If you're new to deploying services on TrueFoundry, we recommend first going through the [Deploy Your First Service](/docs/deploy-first-service) guide.
</Note>

## After Deployment

1. Note the service endpoint URL from the deployment dashboard
2. Test the endpoint to verify it's responding
3. Connect to AI Gateway using the endpoint URL (see [MCP Server Getting Started](/docs/ai-gateway/mcp/mcp-server-getting-started))
