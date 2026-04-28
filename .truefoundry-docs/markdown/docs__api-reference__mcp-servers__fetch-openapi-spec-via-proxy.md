> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Fetch OpenAPI spec via proxy

> Returns the JSON document from a source URL.



## OpenAPI

````yaml /openapi.json post /api/svc/v1/mcp-servers/proxy/openapi-spec
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
  /api/svc/v1/mcp-servers/proxy/openapi-spec:
    post:
      tags:
        - MCP Servers
      summary: Fetch OpenAPI spec via proxy
      description: Returns the JSON document from a source URL.
      operationId: MCPServer.fetchOpenAPISpecProxy
      parameters: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/FetchOpenAPISpecProxyDto'
      responses:
        '200':
          description: OpenAPI spec fetched successfully
          content:
            application/json:
              schema:
                type: object
      security:
        - jwt: []
components:
  schemas:
    FetchOpenAPISpecProxyDto:
      type: object
      properties:
        sourceURL:
          type: string
          description: URL to fetch the OpenAPI spec from
      required:
        - sourceURL
      title: FetchOpenAPISpecProxyDto
      x-fern-type-name: FetchOpenAPISpecProxyDto
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````