> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Get Team Permissions

> Get all role bindings associated with a team.



## OpenAPI

````yaml /openapi.json get /api/svc/v1/teams/{id}/permissions
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
  /api/svc/v1/teams/{id}/permissions:
    get:
      tags:
        - Teams
      summary: Get Team Permissions
      description: Get all role bindings associated with a team.
      operationId: Teams.getTeamPermissions
      parameters:
        - name: id
          required: true
          in: path
          description: Team Id
          schema:
            type: string
      responses:
        '200':
          description: Returns role bindings for the team.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GetTeamPermissionsResponse'
        '403':
          description: Forbidden. User does not have permission to view this data.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
        '404':
          description: Not Found. No team found for the given team ID.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
      security:
        - jwt: []
        - jwt: []
components:
  schemas:
    GetTeamPermissionsResponse:
      type: object
      properties:
        data:
          description: Role bindings for the team
          type: array
          items:
            $ref: '#/components/schemas/SubjectPermission'
      required:
        - data
      title: GetTeamPermissionsResponse
      x-fern-type-name: GetTeamPermissionsResponse
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
    SubjectPermission:
      type: object
      properties:
        resourceType:
          type: string
          description: Resource Type
        resourceId:
          type: string
          description: Resource ID
        resourceName:
          type: string
          description: Resource Name (if available)
        resourceFqn:
          type: string
          description: Resource FQN (if available)
        roleId:
          type: string
          description: Role ID
        roleName:
          type: string
          description: Role Name
        subjectId:
          type: string
          description: Subject ID (user or team)
        subjectType:
          type: string
          description: Subject Type (user or team)
      required:
        - resourceType
        - resourceId
        - roleId
        - roleName
        - subjectId
        - subjectType
      title: SubjectPermission
      x-fern-type-name: SubjectPermission
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````