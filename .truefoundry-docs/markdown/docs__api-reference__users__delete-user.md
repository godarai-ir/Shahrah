> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Delete User

> Delete user if they are not a collaborator in any resource and not part of any team other than everyone.



## OpenAPI

````yaml /openapi.json delete /api/svc/v1/users/{id}
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
  /api/svc/v1/users/{id}:
    delete:
      tags:
        - Users
      summary: Delete User
      description: >-
        Delete user if they are not a collaborator in any resource and not part
        of any team other than everyone.
      operationId: User.deleteUser
      parameters:
        - name: id
          required: true
          in: path
          description: User Id
          schema:
            type: string
        - name: tenantName
          required: false
          in: query
          description: Tenant name
          schema:
            type: string
      responses:
        '200':
          description: User has been successfully deleted.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/DeleteUserResponse'
        '400':
          description: >-
            Bad Request. User cannot be deleted due to active collaborations or
            team memberships.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
        '401':
          description: Unauthorized. User does not have permission to delete user.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
        '404':
          description: Not Found. User with the given ID not found.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
      security:
        - jwt: []
components:
  schemas:
    DeleteUserResponse:
      type: object
      properties: {}
      title: DeleteUserResponse
      x-fern-type-name: DeleteUserResponse
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