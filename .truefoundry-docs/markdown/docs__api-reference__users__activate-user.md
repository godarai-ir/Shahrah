> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Activate User

> Activate user associated with the provided email within the tenant.



## OpenAPI

````yaml /openapi.json patch /api/svc/v1/users/activate
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
  /api/svc/v1/users/activate:
    patch:
      tags:
        - Users
      summary: Activate User
      description: Activate user associated with the provided email within the tenant.
      operationId: User.activateUser
      parameters: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ActivateUserRequest'
      responses:
        '200':
          description: User has been successfully activated.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ActivateUserResponse'
        '401':
          description: Unauthorized. User does not have permission to activate user
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
        '404':
          description: Not Found. User with the given email not found.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
      security:
        - jwt: []
        - jwt: []
components:
  schemas:
    ActivateUserRequest:
      type: object
      properties:
        email:
          type: string
          description: Email of the user
        tenantName:
          type: string
          description: Tenant name
      required:
        - email
      title: ActivateUserRequest
      x-fern-type-name: ActivateUserRequest
    ActivateUserResponse:
      type: object
      properties: {}
      title: ActivateUserResponse
      x-fern-type-name: ActivateUserResponse
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