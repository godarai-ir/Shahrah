> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Revoke All Personal Access Tokens

> Revoke All Personal Access Tokens for the user with the given email



## OpenAPI

````yaml /openapi.json delete /api/svc/v1/personal-access-tokens/revoke/all
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
  /api/svc/v1/personal-access-tokens/revoke/all:
    delete:
      tags:
        - Personal Access Tokens
      summary: Revoke All Personal Access Tokens
      description: Revoke All Personal Access Tokens for the user with the given email
      operationId: PersonalAccessToken.revokeAllPersonalAccessTokens
      parameters: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/RevokeAllPersonalAccessTokenRequest'
      responses:
        '200':
          description: All Personal Access Tokens revoked successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/RevokeAllPersonalAccessTokenResponse'
        '404':
          description: No Personal Access Tokens found for the user with the given email
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
      security:
        - jwt: []
components:
  schemas:
    RevokeAllPersonalAccessTokenRequest:
      type: object
      properties:
        email:
          type: string
          description: Email of the user to revoke all Personal Access Tokens for
      required:
        - email
      title: RevokeAllPersonalAccessTokenRequest
      x-fern-type-name: RevokeAllPersonalAccessTokenRequest
    RevokeAllPersonalAccessTokenResponse:
      type: object
      properties: {}
      title: RevokeAllPersonalAccessTokenResponse
      x-fern-type-name: RevokeAllPersonalAccessTokenResponse
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