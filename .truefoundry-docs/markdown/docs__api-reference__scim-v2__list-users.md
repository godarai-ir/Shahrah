> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# List Users



## OpenAPI

````yaml /openapi.json get /api/svc/v1/scim/v2/{tenantName}/{ssoId}/Users
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
    get:
      tags:
        - SCIM v2
      summary: List Users
      operationId: Scim.getUsers
      parameters:
        - name: startIndex
          required: false
          in: query
          description: 1-based index of the first result to return
          schema:
            minimum: 1
            default: 1
            type: integer
        - name: count
          required: false
          in: query
          description: Max resources per page (capped at 200 per ServiceProviderConfig)
          schema:
            minimum: 1
            maximum: 200
            default: 100
            type: integer
        - name: filter
          required: false
          in: query
          description: SCIM filter expression (e.g. userName eq "user@example.com")
          schema:
            type: string
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
      responses:
        '200':
          description: List of users
      security:
        - jwt: []
components:
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````