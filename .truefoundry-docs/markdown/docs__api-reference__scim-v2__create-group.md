> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Create Group



## OpenAPI

````yaml /openapi.json post /api/svc/v1/scim/v2/{tenantName}/{ssoId}/Groups
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
  /api/svc/v1/scim/v2/{tenantName}/{ssoId}/Groups:
    post:
      tags:
        - SCIM v2
      summary: Create Group
      operationId: Scim.createGroup
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
              $ref: '#/components/schemas/ScimGroup'
      responses:
        '201':
          description: Group created
        '409':
          description: Group already exists
      security:
        - jwt: []
components:
  schemas:
    ScimGroup:
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
        displayName:
          type: string
        members:
          type: array
          items:
            $ref: '#/components/schemas/ScimGroupMember'
        meta:
          $ref: '#/components/schemas/ScimMeta'
      required:
        - schemas
        - displayName
      title: ScimGroup
      x-fern-type-name: ScimGroup
    ScimGroupMember:
      type: object
      properties:
        value:
          type: string
        display:
          type: string
        ref:
          type: string
          description: URI reference to the member resource
      required:
        - value
      title: ScimGroupMember
      x-fern-type-name: ScimGroupMember
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