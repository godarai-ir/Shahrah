> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Get token for a virtual account

> Get token for a virtual account by id



## OpenAPI

````yaml /openapi.json get /api/svc/v1/virtual-accounts/{id}/token
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
  /api/svc/v1/virtual-accounts/{id}/token:
    get:
      tags:
        - Virtual Accounts
      summary: Get token for a virtual account
      description: Get token for a virtual account by id
      operationId: VirtualAccount.getTokenForVirtualAccount
      parameters:
        - name: id
          required: true
          in: path
          description: serviceaccount id
          schema:
            type: string
      responses:
        '200':
          description: Token for the virtual account
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GetTokenForVirtualAccountResponse'
      security:
        - jwt: []
components:
  schemas:
    GetTokenForVirtualAccountResponse:
      type: object
      properties:
        token:
          type: string
          description: Token for the virtual account
      required:
        - token
      title: GetTokenForVirtualAccountResponse
      x-fern-type-name: GetTokenForVirtualAccountResponse
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````