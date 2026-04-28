> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Using the Common Tools MCP Server

> Connect and use tools from the Common Tools MCP Server managed by TrueFoundry

## Overview

The Common Tools MCP Server is a collection of tools that are commonly used by AI applications such as web search, code execution, document extraction, etc. It is managed and deployed by TrueFoundry.
It does not appear in the MCP registry and is available to all users of the platform.

<Callout icon="lightbulb" color="#FFC107" iconType="regular">
  This MCP Server is hosted and managed by TrueFoundry. When you use it (in the Playground or programmatically), your requests are sent to the TrueFoundry‑hosted MCP Server outside your control plane.
</Callout>

## Use Common Tools MCP Server in the AI Gateway Playground

You can use the Common Tools MCP Server in the AI Gateway Playground by selecting the `Common Tools` MCP server from the list of MCP servers.

<img src="https://mintcdn.com/truefoundry/ieLzy-9AkuW1ymqY/images/common-tools-in-playground.png?fit=max&auto=format&n=ieLzy-9AkuW1ymqY&q=85&s=2dfb3d23c8be0e2b0bcd3110421b1a21" alt="Common Tools MCP Server in the AI Gateway Playground" width="3024" height="1582" data-path="images/common-tools-in-playground.png" />

## Programmatically connect to Common Tools MCP Server

You can use the Common Tools MCP Server in your own application by libraries that support the MCP protocol.
It requires a [TrueFoundry API Token](/docs/generating-truefoundry-api-keys) to connect to the MCP server.

<Note>
  Replace the placeholders in the example:

  * `{GATEWAY_BASE_URL}` with your TrueFoundry Gateway Base URL (see [Authentication](/docs/ai-gateway/authentication))
  * `<tfy-api-token>` with your TrueFoundry API token
</Note>

<CodeGroup>
  ```python Python FastMCP lines theme={"dark"}
  # pip install fastmcp
  from fastmcp import Client
  from fastmcp.client.transports import StreamableHttpTransport
  import json

  token = "<tfy-api-token>" # Replace with your actual TFY API token

  async def main():
      url = "{GATEWAY_BASE_URL}/mcp/common-tools/server"
      transport = StreamableHttpTransport(
          url=url,
          auth=token,
      )
      async with Client(transport=transport) as client:
          tools = await client.list_tools()
          for tool in tools:
              print(f"Tool: {tool.name}")
              print(f"Description: {tool.description}")
              print(f"Parameters: {tool.inputSchema}")
          print("--------------------------------")
          tool_name = "web_search"
          tool_args = {"query": "What is the capital of France?"}
          results = await client.call_tool(tool_name, tool_args)
          print(f"Result: {results}")
          print("--------------------------------")


  if __name__ == "__main__":
      import asyncio
      asyncio.run(main())
  ```

  ```typescript TypeScript MCP Client lines theme={"dark"}
  // typescript-mcp-client.ts
  // npm install @modelcontextprotocol/sdk
  import { Client } from "@modelcontextprotocol/sdk/client/index.js";
  import { StreamableHTTPClientTransport } from "@modelcontextprotocol/sdk/client/streamableHttp.js";

  // You can pass any custom headers to MCP servers using the x-tfy-mcp-headers header.
  // This is useful for authentication tokens, metadata, or any headers your MCP server needs.
  //
  // For Remote MCP Servers:
  // - Pass headers as a JSON string: {"Authorization": "Bearer token"}
  // - Headers will be applied to all requests to the MCP server
  //
  // Example usage:
  // const passthroughHeaders = {"Authorization": "Bearer token"};

  // Example usage (runs if this file is executed directly)
  // Run with: npx ts-node typescript-mcp-client.ts
  (async () => {
      const url = '{GATEWAY_BASE_URL}/mcp/common-tools/server';
      const authToken = "<tfy-api-token>"; // Replace with your actual TFY API token
      const toolName = 'web_search';
      const args = {query: 'What is the capital of France?'};

      const requestInit: RequestInit = {
          headers: {
              Authorization: `Bearer ${authToken}`,
              // 'x-tfy-mcp-headers': JSON.stringify(passthroughHeaders)
          }
      };
      const baseUrl = new URL(url);

      const client = new Client({ name: "streamable-http-client", version: "1.0.0" });
      await client.connect(new StreamableHTTPClientTransport(baseUrl, { requestInit }));

      const tools = await client.listTools();
      console.log("Available tools:", tools);

      const result = await client.callTool({ name: toolName, arguments: args });
      console.log("Tool result:", result);
  })();
  ```

  ```json Cursor lines theme={"dark"}
  {
    "mcpServers": {
      "common-tools-custom-domain": {
        "url": "{GATEWAY_BASE_URL}/mcp/common-tools/server",
        "headers": {
          "Authorization": "Bearer <tfy-api-token>"
        }
      }
    }
  }
  ```

  ```json VS Code lines theme={"dark"}
  {
    "servers": {
      "common-tools-custom-domain": {
        "type": "http",
        "url": "{GATEWAY_BASE_URL}/mcp/common-tools/server",
        "headers": {
          "Authorization": "Bearer <tfy-api-token>"
        }
      }
    }
  }
  ```

  ```shell Claude Code lines theme={"dark"}
  claude mcp add --transport http common-tools-custom-domain {GATEWAY_BASE_URL}/mcp/common-tools/server \
    --header "Authorization: Bearer <tfy-api-token>"
  ```

  ```json Windsurf lines theme={"dark"}
  {
    "mcpServers": {
      "common-tools-custom-domain": {
        "serverUrl": "{GATEWAY_BASE_URL}/mcp/common-tools/server",
        "headers": {
          "Authorization": "Bearer <tfy-api-token>"
        }
      }
    }
  }
  ```
</CodeGroup>

## Control what tools are available to your users

Platform administrators can control what tools are available using `Platform > Settings > Common Tools`. It can be completely disabled too.

<img src="https://mintcdn.com/truefoundry/9F9qECwLzhI0scNm/images/common-tools-settings.png?fit=max&auto=format&n=9F9qECwLzhI0scNm&q=85&s=8af79499f6f8b201a487aa3b957ce388" alt="Common Tools MCP Server settings" width="2966" height="1592" data-path="images/common-tools-settings.png" />
