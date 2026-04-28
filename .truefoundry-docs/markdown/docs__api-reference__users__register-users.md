> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Register Users

> This endpoint allows tenant administrators to register users within their tenant.



## OpenAPI

````yaml /openapi.json post /api/svc/v1/users/register
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
  /api/svc/v1/users/register:
    post:
      tags:
        - Users
      summary: Register Users
      description: >-
        This endpoint allows tenant administrators to register users within
        their tenant.
      operationId: User.registerUsers
      parameters: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/RegisterUsersRequest'
      responses:
        '200':
          description: The users have been successfully registered.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/RegisterUsersResponse'
        '401':
          description: Unauthorized. User does not have permission to register users.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
        '409':
          description: Conflict. Email is reserved for internal use.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
        '422':
          description: Unprocessable Entity. Invalid email.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
      security:
        - jwt: []
components:
  schemas:
    RegisterUsersRequest:
      type: object
      properties:
        email:
          type: string
          description: Email of the user
        sendInviteEmail:
          type: boolean
          description: Send invite email if user does not exist
          default: false
        skipIfUserExists:
          type: boolean
          description: Fail if user exists
          default: false
        dryRun:
          type: boolean
          default: false
          description: Dry run
        acceptInviteClientURL:
          type: string
          description: Url to redirect when invite is accepted
          example: <control plane url>/invite-accept
      required:
        - email
      title: RegisterUsersRequest
      x-fern-type-name: RegisterUsersRequest
    RegisterUsersResponse:
      type: object
      properties: {}
      title: RegisterUsersResponse
      x-fern-type-name: RegisterUsersResponse
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