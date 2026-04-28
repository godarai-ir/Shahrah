> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Change Password

> Change password for the authenticated user. Requires clientId and loginId in the request body.



## OpenAPI

````yaml /openapi.json post /api/svc/v1/users/change-password
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
  /api/svc/v1/users/change-password:
    post:
      tags:
        - Users
      summary: Change Password
      description: >-
        Change password for the authenticated user. Requires clientId and
        loginId in the request body.
      operationId: User.changePassword
      parameters: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ChangePasswordRequest'
      responses:
        '200':
          description: Password has been changed successfully.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ChangePasswordResponse'
      security:
        - jwt: []
components:
  schemas:
    ChangePasswordRequest:
      type: object
      properties:
        loginId:
          type: string
          description: login id of the user(email)
        newPassword:
          type: string
          description: New password
        oldPassword:
          type: string
          description: Old password
      required:
        - loginId
        - newPassword
        - oldPassword
      title: ChangePasswordRequest
      x-fern-type-name: ChangePasswordRequest
    ChangePasswordResponse:
      type: object
      properties: {}
      title: ChangePasswordResponse
      x-fern-type-name: ChangePasswordResponse
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````