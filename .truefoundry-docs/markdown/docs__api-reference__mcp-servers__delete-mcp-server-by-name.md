> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Delete MCP server by name

> Deletes an MCP server by its name.



## OpenAPI

````yaml /openapi.json delete /api/svc/v1/mcp-servers/name/{name}
openapi: 3.1.0
info:
  title: TrueFoundry API
  description: API for TrueFoundry
  version: 0.1.0
  contact: {}
servers:
  - url: https://{controlPlaneURL}
    variables:
      controlPlaneURL:
        default: app.truefoundry.com
        description: Control Plane URL
security: []
tags:
  - name: Agent
  - name: Applications
  - name: Apply
  - name: Artifacts
  - name: Audit Logs
  - name: Clusters
  - name: Jobs
  - name: Logs
  - name: MCP Servers
  - name: MLRepos
  - name: Model Deployments
  - name: Models
  - name: Personal Access Tokens
  - name: Prompts
  - name: Provider Integrations
  - name: SCIM v2
  - name: Secret Groups
  - name: Secrets
  - name: Teams
  - name: Traces
  - name: Users
  - name: Virtual Accounts
  - name: Workspaces
paths:
  /api/svc/v1/mcp-servers/name/{name}:
    delete:
      tags:
        - MCP Servers
      summary: Delete MCP server by name
      description: Deletes an MCP server by its name.
      operationId: MCPServer.deleteByName
      parameters:
        - name: name
          required: true
          in: path
          description: MCP Server name
          schema:
            type: string
      responses:
        '200':
          description: MCP server deleted successfully
        '404':
          description: MCP server not found
      security:
        - jwt: []
components:
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````