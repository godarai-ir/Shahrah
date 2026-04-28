> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Delete Team

> Deletes the Team associated with the provided Id.



## OpenAPI

````yaml /openapi.json delete /api/svc/v1/teams/{id}
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
  /api/svc/v1/teams/{id}:
    delete:
      tags:
        - Teams
      summary: Delete Team
      description: Deletes the Team associated with the provided Id.
      operationId: Teams.deleteTeam
      parameters:
        - name: id
          required: true
          in: path
          description: Team Id
          schema:
            type: string
      responses:
        '200':
          description: Successfully deleted the team.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/DeleteTeamResponse'
        '404':
          description: Not Found. No Team with provided Id found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
        '409':
          description: Conflict. Default team "everyone" cannot be deleted
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
      security:
        - jwt: []
        - jwt: []
components:
  schemas:
    DeleteTeamResponse:
      type: object
      properties: {}
      title: DeleteTeamResponse
      x-fern-type-name: DeleteTeamResponse
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