> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# List Users

> List all users of tenant filtered by query and showInvalidUsers. Pagination is available based on query parameters.



## OpenAPI

````yaml /openapi.json get /api/svc/v1/users
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
  /api/svc/v1/users:
    get:
      tags:
        - Users
      summary: List Users
      description: >-
        List all users of tenant filtered by query and showInvalidUsers.
        Pagination is available based on query parameters.
      operationId: User.getAllUsersOfTenant
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
        - name: query
          required: false
          in: query
          schema:
            type: string
        - name: showInvalidUsers
          required: false
          in: query
          description: Show Deactivated users
          schema:
            type: boolean
      responses:
        '200':
          description: >-
            Returns all users of tenant and also the response includes paginated
            data.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ListUsersResponse'
      security:
        - jwt: []
components:
  schemas:
    ListUsersResponse:
      type: object
      properties:
        data:
          description: Array of Users
          type: array
          items:
            $ref: '#/components/schemas/User'
        pagination:
          description: Pagination information
          allOf:
            - $ref: '#/components/schemas/Pagination'
      required:
        - data
        - pagination
      title: ListUsersResponse
      x-fern-type-name: ListUsersResponse
    User:
      type: object
      properties:
        id:
          type: string
        email:
          type: string
        tenantName:
          type: string
        metadata:
          $ref: '#/components/schemas/UserMetadata'
        roles:
          type: array
          items:
            type: string
        rolesWithResource:
          type: array
          items:
            $ref: '#/components/schemas/RoleWithResource'
        accounts:
          type: array
          items:
            $ref: '#/components/schemas/AccountInfo'
        active:
          type: boolean
        isEditable:
          type: boolean
        createdAt:
          format: date-time
          type: string
        updatedAt:
          format: date-time
          type: string
      required:
        - id
        - email
        - tenantName
        - metadata
        - active
        - isEditable
        - createdAt
        - updatedAt
      title: User
      x-fern-type-name: User
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
    UserMetadata:
      type: object
      properties:
        sub:
          type: string
        imageURL:
          type: string
        pictureDownloadPath:
          type: string
        displayName:
          type: string
        userObject:
          type: object
          additionalProperties: true
        inviteAccepted:
          type: boolean
        registeredInIdp:
          type: boolean
        preference:
          type: object
          additionalProperties: true
        groups:
          type: array
          items:
            type: string
        tenantRoleManagedBy:
          type: string
          enum:
            - manual
            - group
            - initial
        ssoName:
          type: string
        isPrimarySSO:
          type: boolean
        scimUserObject:
          type: object
          additionalProperties: true
        createdByScim:
          type: boolean
      title: UserMetadata
      x-fern-type-name: UserMetadata
    RoleWithResource:
      type: object
      properties:
        roleId:
          type: string
        resourceType:
          type: string
          enum:
            - role
            - account
            - cluster
            - workspace
            - environment
            - secret-group
            - deployment
            - application
            - secret
            - docker-registry
            - tenant
            - vcs-integration
            - mlf-project
            - repository
            - provider-integration
            - team
            - service-account
            - llm-gateway
            - policy
            - role-binding
            - settings
            - provider-account
            - user
            - alert-config
            - alert-rule
            - gateway-config
            - gateway-controls
            - tracing-project
            - tracing-application
            - agent
            - agent-app
            - gateway-installation
            - virtual-account
            - sso-based-identity
            - external-identity
            - mcp-server
            - external-identity-provider
        resourceId:
          type: string
      required:
        - roleId
        - resourceType
        - resourceId
      title: RoleWithResource
      x-fern-type-name: RoleWithResource
    AccountInfo:
      type: object
      properties:
        accountId:
          type: string
        name:
          type: string
      required:
        - accountId
        - name
      title: AccountInfo
      x-fern-type-name: AccountInfo
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````