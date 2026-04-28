> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Invite User

> Invite a user to the tenant



## OpenAPI

````yaml /openapi.json post /api/svc/v1/users/invite
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
  /api/svc/v1/users/invite:
    post:
      tags:
        - Users
      summary: Invite User
      description: Invite a user to the tenant
      operationId: User.inviteUser
      parameters: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/InviteUserRequest'
      responses:
        '200':
          description: User has been successfully invited.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/InviteUserResponse'
        '401':
          description: Unauthorized. Access token not found in request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
        '403':
          description: Forbidden. User invite is not allowed.
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
      security:
        - jwt: []
        - jwt: []
components:
  schemas:
    InviteUserRequest:
      type: object
      properties:
        acceptInviteClientUrl:
          type: string
          description: Url to redirect when invite is accepted
          example: <control plane url>/invite-accept
        email:
          type: string
          description: Email of user
      required:
        - acceptInviteClientUrl
        - email
      title: InviteUserRequest
      x-fern-type-name: InviteUserRequest
    InviteUserResponse:
      type: object
      properties:
        link:
          type: string
      required:
        - link
      title: InviteUserResponse
      x-fern-type-name: InviteUserResponse
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