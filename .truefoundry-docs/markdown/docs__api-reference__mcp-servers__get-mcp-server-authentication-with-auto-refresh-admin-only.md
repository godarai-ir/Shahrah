> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Get MCP server authentication with auto-refresh (Admin only)

> Retrieves authentication details for the specified user. Automatically refreshes expired tokens for OAuth based servers. Requires admin access.



## OpenAPI

````yaml /openapi.json get /api/svc/v1/mcp-servers/{id}/auth
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
    get:
      tags:
        - MCP Servers
      summary: Get MCP server authentication with auto-refresh (Admin only)
      description: >-
        Retrieves authentication details for the specified user. Automatically
        refreshes expired tokens for OAuth based servers. Requires admin access.
      operationId: MCPServer.getAuth
      parameters:
        - name: id
          required: true
          in: path
          schema:
            type: string
        - name: userId
          required: true
          in: query
          description: User ID
          schema:
            type: string
        - name: tenantName
          required: true
          in: query
          description: Tenant name
          schema:
            type: string
        - name: subjectType
          required: false
          in: query
          description: Subject type
          schema:
            type: string
            enum:
              - user
              - virtualaccount
              - external-identity
      responses:
        '200':
          description: Authentication details with refresh metadata
          content:
            application/json:
              schema:
                type: object
                properties:
                  status:
                    type: string
                    enum:
                      - authenticated
                      - unauthenticated
                  data:
                    type: object
                    description: Authentication data (only present if authenticated)
                    properties:
                      type:
                        type: string
                        enum:
                          - oauth2
                      access_token:
                        type: string
                      refresh_token:
                        type: string
                      id_token:
                        type: string
                      expires_at:
                        type: number
                  timestamp:
                    type: number
                    description: >-
                      Timestamp (unix ms) of when the auth record was last
                      updated
                  mcpAuthRevision:
                    type: number
                    description: Auth revision timestamp for MCP server auth configuration
                  reason:
                    type: string
                    description: >-
                      Reason for unauthenticated status (e.g., network_error,
                      upstream_error, invalid_grant, refresh_failed)
        '400':
          description: Missing required parameters or tenant mismatch
        '403':
          description: Admin access required
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