> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Delete Application

> Delete Application associated with the provided application ID.



## OpenAPI

````yaml /openapi.json delete /api/svc/v1/apps/{id}
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
  /api/svc/v1/apps/{id}:
    delete:
      tags:
        - Applications
      summary: Delete Application
      description: Delete Application associated with the provided application ID.
      operationId: Application.deleteApplication
      parameters:
        - name: id
          required: true
          in: path
          description: Id of the application
          schema:
            type: string
      responses:
        '200':
          description: Delete application response.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/DeleteApplicationResponse'
        '403':
          description: >-
            Forbidden. User does not delete application permission on the
            workspace
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
        '404':
          description: Not found. No application found for the given application ID
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
      security:
        - jwt: []
components:
  schemas:
    DeleteApplicationResponse:
      type: object
      properties:
        message:
          type: string
          description: Application deletion message
      required:
        - message
      title: DeleteApplicationResponse
      x-fern-type-name: DeleteApplicationResponse
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