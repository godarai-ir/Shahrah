> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Delete MCP server authentication

> Deletes the current user's OAuth 2.1 authentication for an MCP server integration.



## OpenAPI

````yaml /openapi.json delete /api/svc/v1/mcp-servers/{id}/auth
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
  /api/svc/v1/mcp-servers/{id}/auth:
    delete:
      tags:
        - MCP Servers
      summary: Delete MCP server authentication
      description: >-
        Deletes the current user's OAuth 2.1 authentication for an MCP server
        integration.
      operationId: MCPServer.deleteAuth
      parameters:
        - name: id
          required: true
          in: path
          schema:
            type: string
      responses:
        '200':
          description: OAuth 2.1 authentication data deleted successfully
        '404':
          description: Integration not found or user not authenticated
      security:
        - jwt: []
        - jwt: []
components:
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````