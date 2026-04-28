> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Get Application Resources

> Retrieves the ArgoCD resources for the specified application.



## OpenAPI

````yaml /openapi.json get /api/svc/v1/apps/{id}/resources
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
  /api/svc/v1/apps/{id}/resources:
    get:
      tags:
        - Applications
      summary: Get Application Resources
      description: Retrieves the ArgoCD resources for the specified application.
      operationId: Application.getApplicationResources
      parameters:
        - name: id
          required: true
          in: path
          description: Id of the application
          schema:
            type: string
      responses:
        '200':
          description: Successfully retrieved resources
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GetApplicationResourcesResponse'
        '403':
          description: >-
            Forbidden. User does not have read application permission for this
            application
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
        '404':
          description: Not found. Application with provided ID not found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
      security:
        - jwt: []
components:
  schemas:
    GetApplicationResourcesResponse:
      type: object
      properties:
        resources:
          description: Resources
          type: array
          items:
            $ref: '#/components/schemas/ResourceResult'
      required:
        - resources
      title: GetApplicationResourcesResponse
      x-fern-type-name: GetApplicationResourcesResponse
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
    ResourceResult:
      type: object
      properties: {}
      title: ResourceResult
      x-fern-type-name: ResourceResult
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````