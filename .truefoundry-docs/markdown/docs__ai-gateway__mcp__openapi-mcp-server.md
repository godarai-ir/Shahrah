> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# OpenAPI to MCP Server

> Automatically generate MCP servers from your existing OpenAPI specifications to expose your APIs as AI-callable tools.

## What is OpenAPI MCP Server?

The OpenAPI MCP Server feature allows you to automatically create an MCP server from any existing OpenAPI (Swagger) specification. Instead of manually coding an MCP server to expose your API endpoints, TrueFoundry parses your OpenAPI spec and generates tools for each endpoint automatically.

This is particularly useful when you want to:

* Quickly expose existing REST APIs as MCP tools for AI agents
* Avoid duplicating API definitions between your backend and MCP server
* Quickly create a POC for development

## How It Works

When you provide an OpenAPI specification, TrueFoundry:

1. **Parses the OpenAPI spec** to extract all available endpoints, their parameters, request bodies, and response schemas
2. **Generates MCP tools** for each endpoint with appropriate input schemas derived from the OpenAPI definition
3. **Creates a OpenAPI MCP server** that proxies requests to your actual API endpoints
4. **Handles authentication** based on the security schemes defined in your OpenAPI spec

| OpenAPI Element              | MCP Mapping       | Notes                                                     |
| ---------------------------- | ----------------- | --------------------------------------------------------- |
| `operationId`                | Tool name         | Uses `operationId` if present; otherwise generates a name |
| `summary` / `description`    | Tool description  | Uses `summary` if available, falls back to `description`  |
| `parameters` + `requestBody` | Tool inputSchema  | Combined into a single input schema                       |
| `responses`                  | Tool outputSchema | Response schema                                           |

## Create an OpenAPI MCP Server

Follow the interactive demo below to create and configure an MCP server from your OpenAPI specification:

<iframe provider="app.supademo.com" href="https://app.supademo.com/demo/cml7ieh7q6zalzsad5fa88vdu" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/cml7ieh7q6zalzsad5fa88vdu" style={{ border:"none",display:"flex",margin:"auto" }} />

<Steps>
  <Step title="Navigate to MCP Servers">
    Go to the **MCP Servers** tab in the TrueFoundry AI Gateway and click **Add Server**.
  </Step>

  <Step title="Select OpenAPI Source">
    Click **From OpenAPI spec**. You can provide your OpenAPI specification in one of the following ways:

    | Method            | Description                                                         |
    | ----------------- | ------------------------------------------------------------------- |
    | **URL**           | Provide a publicly accessible URL to your OpenAPI JSON or YAML file |
    | **Paste Content** | Paste the raw OpenAPI specification content                         |
  </Step>

  <Step title="Select Tools and Configure">
    After providing the OpenAPI spec, select the tools you want to expose. For each tool, you can configure:

    | Field                | Description                                                              |
    | -------------------- | ------------------------------------------------------------------------ |
    | **Tool Name**        | The name used to call the tool (defaults to `operationId` from the spec) |
    | **Tool Description** | A description to help AI agents understand when to use this tool         |
  </Step>

  <Step title="Configure Server Details">
    Provide the following configuration for your OpenAPI MCP Server:

    | Field                  | Description                                                    |
    | ---------------------- | -------------------------------------------------------------- |
    | **Name**               | A descriptive name for the MCP server (e.g., `my-api-server`)  |
    | **Description**        | A brief description of what this MCP server provides           |
    | **OpenAPI Server URL** | The base URL of your API server where requests will be proxied |
  </Step>

  <Step title="Configure Authentication and Access Controls">
    Set up authentication based on how your API is secured. See [MCP Gateway Auth & Security](/docs/ai-gateway/mcp/mcp-gateway-auth-security) for detailed authentication options and [Access Control](/docs/ai-gateway/mcp/mcp-gateway-auth-security#access-control) for managing permissions.
  </Step>

  <Step title="Create and Test">
    Click **Create** to generate your OpenAPI MCP Server. Once created, you can:

    * Test tools in the **Playground**
    * View generated tools in the **Tools** tab
    * Get integration snippets from the **How To Use** tab
  </Step>
</Steps>

## Frequently Asked Questions

<AccordionGroup>
  <Accordion title="What OpenAPI versions are supported?">
    TrueFoundry supports OpenAPI 3.0 and 3.1 specifications.
  </Accordion>

  <Accordion title="Can I update the OpenAPI spec after creation?">
    Yes, you can update the OpenAPI specification at any time. Navigate to the MCP server settings and upload a new spec or update the URL. Tools will be regenerated based on the updated specification.
  </Accordion>

  <Accordion title="What happens if my API returns errors?">
    API errors are passed through to the AI agent as tool execution failures. The error response body and status code are included in the error message, allowing the agent to handle or report the error appropriately.
  </Accordion>

  <Accordion title="Does this work with authenticated APIs?">
    Yes. Configure the appropriate authentication method when creating the MCP server. The gateway handles injecting authentication headers or tokens into every request made to your API.
  </Accordion>

  <Accordion title="Can I combine OpenAPI MCP Server with Virtual MCP Server?">
    Yes. Once your OpenAPI MCP Server is created, you can include its tools in a [Virtual MCP Server](/docs/ai-gateway/mcp/virtual-mcp-server) to create curated tool collections that combine endpoints from multiple APIs.
  </Accordion>
</AccordionGroup>

## Best Practices

<CardGroup cols={2}>
  <Card title="Use Descriptive Operation IDs" icon="tag">
    Ensure your OpenAPI spec has clear `operationId` values as these become tool names that AI agents use.
  </Card>

  <Card title="Document Parameters Well" icon="file-lines">
    Include detailed descriptions for all parameters as this helps AI agents understand when and how to use each tool.
  </Card>

  <Card title="Secure Sensitive Endpoints" icon="shield">
    Use endpoint filtering to exclude administrative or destructive operations from the MCP server.
  </Card>
</CardGroup>
