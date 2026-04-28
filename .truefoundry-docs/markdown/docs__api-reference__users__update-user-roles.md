> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Update User Roles

> This endpoint allows tenant administrators to update the roles of a user within their tenant.



## OpenAPI

````yaml /openapi.json patch /api/svc/v1/users/roles
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
  /api/svc/v1/users/roles:
    patch:
      tags:
        - Users
      summary: Update User Roles
      description: >-
        This endpoint allows tenant administrators to update the roles of a user
        within their tenant.
      operationId: User.editUserRole
      parameters: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/UpdateUserRolesRequest'
      responses:
        '200':
          description: The user roles have been successfully updated.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/UpdateUserRolesResponse'
        '401':
          description: Unauthorized. User does not have permission to update roles.
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
        '422':
          description: Unprocessable Entity. Missing required fields.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
      security:
        - jwt: []
components:
  schemas:
    UpdateUserRolesRequest:
      type: object
      properties:
        email:
          type: string
          description: Email of the user
        roles:
          description: Role names for the user
          type: array
          items:
            type: string
        resourceType:
          type: string
          description: Resource Type
      required:
        - email
        - roles
      title: UpdateUserRolesRequest
      x-fern-type-name: UpdateUserRolesRequest
    UpdateUserRolesResponse:
      type: object
      properties: {}
      title: UpdateUserRolesResponse
      x-fern-type-name: UpdateUserRolesResponse
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