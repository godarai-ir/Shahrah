> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Cancel an Ongoing Deployment

> Cancel an ongoing deployment associated with the provided application ID and deployment ID.



## OpenAPI

````yaml /openapi.json post /api/svc/v1/apps/{id}/deployments/{deploymentId}/cancel
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
  /api/svc/v1/apps/{id}/deployments/{deploymentId}/cancel:
    post:
      tags:
        - Applications
      summary: Cancel an Ongoing Deployment
      description: >-
        Cancel an ongoing deployment associated with the provided application ID
        and deployment ID.
      operationId: Application.cancelDeployment
      parameters:
        - name: id
          required: true
          in: path
          description: Application id of the application
          schema:
            type: string
        - name: deploymentId
          required: true
          in: path
          description: Deployment id of the deployment
          schema:
            type: string
      responses:
        '200':
          description: Deployment cancelled successfully.
          content:
            application/json:
              schema:
                type: object
                properties:
                  message:
                    type: string
                    description: Deployment cancelled
        '403':
          description: Forbidden. User does not update permission on this application
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
        '404':
          description: Not Found. Deployment not found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
        '409':
          description: >-
            Conflict. Deployment cannot be cancelled as it has already reached a
            terminal state.
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