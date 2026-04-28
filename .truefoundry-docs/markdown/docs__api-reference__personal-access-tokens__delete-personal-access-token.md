> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Delete Personal Access Token

> Delete Personal Access Token associated with the provided serviceAccountId



## OpenAPI

````yaml /openapi.json delete /api/svc/v1/personal-access-tokens/{id}
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
  /api/svc/v1/personal-access-tokens/{id}:
    delete:
      tags:
        - Personal Access Tokens
      summary: Delete Personal Access Token
      description: >-
        Delete Personal Access Token associated with the provided
        serviceAccountId
      operationId: PersonalAccessToken.deletePersonalAccessToken
      parameters:
        - name: id
          required: true
          in: path
          description: serviceaccount id
          schema:
            type: string
      responses:
        '200':
          description: Personal Access Token deleted successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/DeletePersonalAccessTokenResponse'
        '404':
          description: >-
            Not Found. Personal Access Token with the provided serviceAccountId
            not found.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
      security:
        - jwt: []
components:
  schemas:
    DeletePersonalAccessTokenResponse:
      type: object
      properties: {}
      title: DeletePersonalAccessTokenResponse
      x-fern-type-name: DeletePersonalAccessTokenResponse
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
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````