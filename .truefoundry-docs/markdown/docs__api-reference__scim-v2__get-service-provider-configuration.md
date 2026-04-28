> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Get Service Provider Configuration



## OpenAPI

````yaml /openapi.json get /api/svc/v1/scim/v2/{tenantName}/{ssoId}/ServiceProviderConfig
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
  /api/svc/v1/scim/v2/{tenantName}/{ssoId}/ServiceProviderConfig:
    get:
      tags:
        - SCIM v2
      summary: Get Service Provider Configuration
      operationId: Scim.getServiceProviderConfig
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
      responses:
        '200':
          description: SCIM capabilities
      security:
        - jwt: []
components:
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````