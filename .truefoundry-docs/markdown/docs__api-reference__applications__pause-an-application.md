> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Pause an Application

> Pause a running application by scaling to 0 replicas



## OpenAPI

````yaml /openapi.json patch /api/svc/v1/apps/{id}/scale-to-zero
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
  /api/svc/v1/apps/{id}/scale-to-zero:
    patch:
      tags:
        - Applications
      summary: Pause an Application
      description: Pause a running application by scaling to 0 replicas
      operationId: Application.scaleToZero
      parameters:
        - name: id
          required: true
          in: path
          description: Id of the application
          schema:
            type: string
      responses:
        '200':
          description: Scales the number of replicas of the application deployment to 0
        '403':
          description: Forbidden. User does not have update permission on this application
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
        '404':
          description: Not Found. Application with provided Application Id not found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
        '405':
          description: Pausing/Scale to Zero is not applicable for this application
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
        '501':
          description: Pausing/Scale to Zero is not applicable for the application type
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
      security:
        - jwt: []
components:
  schemas:
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