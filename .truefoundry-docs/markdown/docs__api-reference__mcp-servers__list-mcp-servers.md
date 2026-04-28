> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# List MCP servers

> Lists all MCP servers for the current tenant with pagination.



## OpenAPI

````yaml /openapi.json get /api/svc/v1/mcp-servers
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
  /api/svc/v1/mcp-servers:
    get:
      tags:
        - MCP Servers
      summary: List MCP servers
      description: Lists all MCP servers for the current tenant with pagination.
      operationId: MCPServer.list
      parameters:
        - name: limit
          required: false
          in: query
          description: Number of items per page
          schema:
            minimum: 1
            maximum: 1000
            default: 100
            example: 10
            type: integer
        - name: offset
          required: false
          in: query
          description: Number of items to skip
          schema:
            minimum: 0
            default: 0
            example: 0
            type: integer
        - name: id
          required: false
          in: query
          description: Filter by MCP Server ID
          schema:
            type: string
        - name: name
          required: false
          in: query
          description: Filter by MCP Server name
          schema:
            type: string
        - name: fqn
          required: false
          in: query
          description: Filter by MCP Server FQN
          schema:
            type: string
      responses:
        '200':
          description: Returns list of MCP servers
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ListMCPServersResponse'
      security:
        - jwt: []
components:
  schemas:
    ListMCPServersResponse:
      type: object
      properties:
        data:
          description: Array of MCP Servers
          type: array
          items:
            $ref: '#/components/schemas/MCPServer'
        pagination:
          description: Pagination Information
          allOf:
            - $ref: '#/components/schemas/Pagination'
      required:
        - data
        - pagination
      title: ListMCPServersResponse
      x-fern-type-name: ListMCPServersResponse
    MCPServer:
      type: object
      properties:
        id:
          type: string
        fqn:
          type: string
        manifest:
          type: object
          additionalProperties: true
        createdBySubject:
          $ref: '#/components/schemas/Subject'
        accountId:
          type: string
        tools:
          type: array
          items:
            type: object
        specEndpoints:
          type: array
          items:
            type: object
        createdAt:
          format: date-time
          type: string
        updatedAt:
          format: date-time
          type: string
        proxyUrl:
          type: string
          description: Proxy URL for accessing the MCP server
        authStatus:
          description: Authentication status for the MCP server
          allOf:
            - $ref: '#/components/schemas/MCPServerAuthStatusInfo'
        isAuthOverrideSupported:
          type: boolean
      required:
        - fqn
        - manifest
        - createdBySubject
        - accountId
      title: MCPServer
      x-fern-type-name: MCPServer
    Pagination:
      type: object
      properties:
        total:
          type: integer
          description: Total number of items
          example: 100
        offset:
          type: integer
          description: Number of items skipped
          example: 0
        limit:
          type: integer
          description: Number of items per page
          example: 10
      required:
        - total
      title: Pagination
      x-fern-type-name: Pagination
    Subject:
      type: object
      properties:
        subjectId:
          type: string
          description: Subject ID
        subjectType:
          description: Subject type
          allOf:
            - $ref: '#/components/schemas/SubjectType'
        subjectSlug:
          type: string
          description: Subject slug
        subjectDisplayName:
          type: string
          description: Subject display name
        subjectPatName:
          type: string
          description: Subject PAT name
        subjectControllerName:
          type: string
        subjectExternalIdentitySlug:
          type: string
          description: External identity slug (external_identity_id:sub:email)
      required:
        - subjectId
        - subjectType
      title: Subject
      x-fern-type-name: Subject
    MCPServerAuthStatusInfo:
      type: object
      properties:
        status:
          type: string
          enum:
            - authenticated
            - unauthenticated
          description: >-
            Whether the user is authenticated or unauthenticated for this MCP
            server
        method:
          type: string
          enum:
            - oauth
            - auth-override
          description: Authentication method used (e.g. OAuth, auth-override)
        message:
          type: string
          description: >-
            Human-readable message providing additional context about the
            authentication status
      required:
        - status
      title: MCPServerAuthStatusInfo
      x-fern-type-name: MCPServerAuthStatusInfo
    SubjectType:
      type: string
      enum:
        - user
        - team
        - serviceaccount
        - virtualaccount
        - external-identity
        - role
      title: SubjectType
      x-fern-type-name: SubjectType
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````