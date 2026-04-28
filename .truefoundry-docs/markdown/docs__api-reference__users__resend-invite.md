> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Resend Invite

> Resend the invite to the user



## OpenAPI

````yaml /openapi.json post /api/svc/v1/users/invite/resend
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
  /api/svc/v1/users/invite/resend:
    post:
      tags:
        - Users
      summary: Resend Invite
      description: Resend the invite to the user
      operationId: User.resendInvite
      parameters: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ResendInviteRequest'
      responses:
        '200':
          description: User invite has been successfully resent.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ResendInviteResponse'
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
      security:
        - jwt: []
        - jwt: []
components:
  schemas:
    ResendInviteRequest:
      type: object
      properties:
        acceptInviteClientUrl:
          type: string
          description: Url to redirect when invite is accepted
          example: <control plane url>/invite-accept
        email:
          type: string
          description: User email of user
        tenantName:
          type: string
          description: Tenant name
      required:
        - acceptInviteClientUrl
        - email
      title: ResendInviteRequest
      x-fern-type-name: ResendInviteRequest
    ResendInviteResponse:
      type: object
      properties:
        link:
          type: string
      required:
        - link
      title: ResendInviteResponse
      x-fern-type-name: ResendInviteResponse
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