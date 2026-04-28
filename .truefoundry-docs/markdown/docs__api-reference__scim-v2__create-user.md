> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Create User



## OpenAPI

````yaml /openapi.json post /api/svc/v1/scim/v2/{tenantName}/{ssoId}/Users
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
  /api/svc/v1/scim/v2/{tenantName}/{ssoId}/Users:
    post:
      tags:
        - SCIM v2
      summary: Create User
      operationId: Scim.createUser
      parameters:
        - name: tenantName
          required: true
          in: path
          description: Tenant name
          schema:
            type: string
        - name: ssoId
          required: true
          in: path
          description: SSO configuration ID
          schema:
            type: string
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ScimUser'
      responses:
        '201':
          description: User created
        '409':
          description: User already exists
      security:
        - jwt: []
components:
  schemas:
    ScimUser:
      type: object
      properties:
        schemas:
          type: array
          items:
            type: string
        id:
          type: string
        externalId:
          type: string
        userName:
          type: string
        name:
          $ref: '#/components/schemas/ScimUserName'
        displayName:
          type: string
        emails:
          type: array
          items:
            $ref: '#/components/schemas/ScimUserEmail'
        active:
          type: boolean
        meta:
          $ref: '#/components/schemas/ScimMeta'
      required:
        - schemas
        - userName
        - emails
        - active
      title: ScimUser
      x-fern-type-name: ScimUser
    ScimUserName:
      type: object
      properties:
        formatted:
          type: string
        familyName:
          type: string
        givenName:
          type: string
        middleName:
          type: string
      title: ScimUserName
      x-fern-type-name: ScimUserName
    ScimUserEmail:
      type: object
      properties:
        value:
          type: string
        type:
          type: string
        primary:
          type: boolean
      required:
        - value
      title: ScimUserEmail
      x-fern-type-name: ScimUserEmail
    ScimMeta:
      type: object
      properties:
        resourceType:
          type: string
        created:
          type: string
        lastModified:
          type: string
        location:
          type: string
        version:
          type: string
      title: ScimMeta
      x-fern-type-name: ScimMeta
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````