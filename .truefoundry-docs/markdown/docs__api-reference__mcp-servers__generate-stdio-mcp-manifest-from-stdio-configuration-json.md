> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Generate stdio MCP manifest from stdio configuration JSON

> Parses a single-entry mcpServers object and returns a stdio MCP server manifest.



## OpenAPI

````yaml /openapi.json post /api/svc/v1/mcp-servers/stdio/generate-manifest-from-stdio-config-json
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
  /api/svc/v1/mcp-servers/stdio/generate-manifest-from-stdio-config-json:
    post:
      tags:
        - MCP Servers
      summary: Generate stdio MCP manifest from stdio configuration JSON
      description: >-
        Parses a single-entry mcpServers object and returns a stdio MCP server
        manifest.
      operationId: MCPServer.generateStdioManifestFromStdioConfigurationJson
      parameters: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/StdioConfigurationJsonToManifestRequestDto'
      responses:
        '200':
          description: Stdio MCP manifest generated successfully
          content:
            application/json:
              schema:
                $ref: >-
                  #/components/schemas/StdioConfigurationJsonToManifestResponseDto
        '400':
          description: Invalid json payload or derived name
      security:
        - jwt: []
components:
  schemas:
    StdioConfigurationJsonToManifestRequestDto:
      type: object
      properties:
        mcpServers:
          type: object
          description: >-
            Stdio configuration (JSON): map of MCP server id to stdio config
            (exactly one entry). Includes command, args, env.
          additionalProperties: true
        description:
          type: string
        authLevel:
          type: string
          enum:
            - global
            - per_user
      required:
        - mcpServers
      title: StdioConfigurationJsonToManifestRequestDto
      x-fern-type-name: StdioConfigurationJsonToManifestRequestDto
    StdioConfigurationJsonToManifestResponseDto:
      type: object
      properties:
        manifest:
          type: object
          description: Stdio MCP server manifest for create/update
      required:
        - manifest
      title: StdioConfigurationJsonToManifestResponseDto
      x-fern-type-name: StdioConfigurationJsonToManifestResponseDto
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````