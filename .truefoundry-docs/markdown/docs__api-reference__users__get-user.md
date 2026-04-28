> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Get User

> Get User associated with provided User id



## OpenAPI

````yaml /openapi.json get /api/svc/v1/users/{id}
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
  /api/svc/v1/users/{id}:
    get:
      tags:
        - Users
      summary: Get User
      description: Get User associated with provided User id
      operationId: User.getUserById
      parameters:
        - name: id
          required: true
          in: path
          description: User Id
          schema:
            type: string
      responses:
        '200':
          description: Returns the User associated with provided User id
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GetUserResponse'
        '404':
          description: Not Found. No user found for the given user ID
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
      security:
        - jwt: []
components:
  schemas:
    GetUserResponse:
      type: object
      properties:
        data:
          description: User
          allOf:
            - $ref: '#/components/schemas/User'
      required:
        - data
      title: GetUserResponse
      x-fern-type-name: GetUserResponse
    HttpError:
      type: object
      properties:
        statusCode:
          type: integer
          description: HTTP Status Code
        message:
          type: string
          description: Error Message
        code:
          oneOf:
            - type: integer
            - type: string
          description: Error code
        details:
          type: array
          description: Error details
          items:
            type: object
      required:
        - statusCode
        - message
      title: HttpError
      x-fern-type-name: HttpError
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