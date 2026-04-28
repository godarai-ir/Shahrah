> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Check Virtual Account Exists

> Check whether a Virtual Account with the given name exists in the current tenant.



## OpenAPI

````yaml /openapi.json get /api/svc/v1/virtual-accounts/exists
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
  /api/svc/v1/virtual-accounts/exists:
    get:
      tags:
        - Virtual Accounts
      summary: Check Virtual Account Exists
      description: >-
        Check whether a Virtual Account with the given name exists in the
        current tenant.
      operationId: VirtualAccount.virtualAccountExists
      parameters:
        - name: name
          required: true
          in: query
          schema:
            type: string
      responses:
        '200':
          description: Whether Virtual Account exists
          content:
            application/json:
              schema:
                type: object
                properties:
                  exists:
                    type: boolean
                required:
                  - exists
      security:
        - jwt: []
components:
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````