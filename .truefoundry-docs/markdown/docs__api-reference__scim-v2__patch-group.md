> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Patch Group



## OpenAPI

````yaml /openapi.json patch /api/svc/v1/scim/v2/{tenantName}/{ssoId}/Groups/{id}
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
  /api/svc/v1/scim/v2/{tenantName}/{ssoId}/Groups/{id}:
    patch:
      tags:
        - SCIM v2
      summary: Patch Group
      operationId: Scim.patchGroup
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
        - name: id
          required: true
          in: path
          description: Resource ID
          schema:
            type: string
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ScimPatchRequest'
      responses:
        '200':
          description: Group patched
        '404':
          description: Group not found
      security:
        - jwt: []
components:
  schemas:
    ScimPatchRequest:
      type: object
      properties:
        schemas:
          type: array
          items:
            type: string
        Operations:
          type: array
          items:
            $ref: '#/components/schemas/ScimPatchOperation'
      required:
        - schemas
        - Operations
      title: ScimPatchRequest
      x-fern-type-name: ScimPatchRequest
    ScimPatchOperation:
      type: object
      properties:
        op:
          type: string
          enum:
            - Add
            - Remove
            - Replace
            - add
            - remove
            - replace
        path:
          type: string
        value:
          type: object
          description: Value for the operation. Type depends on the path being modified.
      required:
        - op
      title: ScimPatchOperation
      x-fern-type-name: ScimPatchOperation
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````