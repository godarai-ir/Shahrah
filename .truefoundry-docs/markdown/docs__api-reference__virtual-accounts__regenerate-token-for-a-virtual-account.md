> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Regenerate token for a virtual account

> Regenerate token for a virtual account by id. The old token will remain valid for the specified grace period.



## OpenAPI

````yaml /openapi.json post /api/svc/v1/virtual-accounts/{id}/regenerate-token
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
  /api/svc/v1/virtual-accounts/{id}/regenerate-token:
    post:
      tags:
        - Virtual Accounts
      summary: Regenerate token for a virtual account
      description: >-
        Regenerate token for a virtual account by id. The old token will remain
        valid for the specified grace period.
      operationId: VirtualAccount.regenerateTokenForVirtualAccount
      parameters:
        - name: id
          required: true
          in: path
          description: serviceaccount id
          schema:
            type: string
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/RegenerateVirtualAccountTokenRequest'
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
    RegenerateVirtualAccountTokenRequest:
      type: object
      properties:
        gracePeriodInDays:
          type: number
          description: >-
            Grace period in days for which the old token will remain valid after
            regeneration
          minimum: 0
          maximum: 365
          example: 30
      required:
        - gracePeriodInDays
      title: RegenerateVirtualAccountTokenRequest
      x-fern-type-name: RegenerateVirtualAccountTokenRequest
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