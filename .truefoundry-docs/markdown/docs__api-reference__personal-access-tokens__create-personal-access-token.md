> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Create Personal Access Token

> Create Personal Access Token



## OpenAPI

````yaml /openapi.json post /api/svc/v1/personal-access-tokens
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
  /api/svc/v1/personal-access-tokens:
    post:
      tags:
        - Personal Access Tokens
      summary: Create Personal Access Token
      description: Create Personal Access Token
      operationId: PersonalAccessToken.createPersonalAccessToken
      parameters: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreatePersonalAccessTokenRequest'
      responses:
        '200':
          description: Personal Access Token created successfully and returned
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CreatePersonalAccessTokenResponse'
        '400':
          description: Bad Request. Invalid name provided.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
        '409':
          description: Conflict. Personal Access Token with the same name already exists.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
      security:
        - jwt: []
components:
  schemas:
    CreatePersonalAccessTokenRequest:
      type: object
      properties:
        name:
          type: string
          description: serviceaccount name
        expirationDate:
          type: string
          description: Expiration date in ISO format (e.g. 2025-08-01T12:00)
          example: 2025-09-01T12:00
        accountName:
          type: string
          description: Account name that owns this PAT
      required:
        - name
      title: CreatePersonalAccessTokenRequest
      x-fern-type-name: CreatePersonalAccessTokenRequest
    CreatePersonalAccessTokenResponse:
      type: object
      properties:
        token:
          type: string
          description: Personal Access Token
      required:
        - token
      title: CreatePersonalAccessTokenResponse
      x-fern-type-name: CreatePersonalAccessTokenResponse
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