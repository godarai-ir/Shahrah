> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Get Cluster status

> Get the status of provided cluster



## OpenAPI

````yaml /openapi.json get /api/svc/v1/clusters/{id}/is-connected
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
  /api/svc/v1/clusters/{id}/is-connected:
    get:
      tags:
        - Clusters
      summary: Get Cluster status
      description: Get the status of provided cluster
      operationId: Cluster.getClusterStatus
      parameters:
        - name: id
          required: true
          in: path
          description: Cluster id of the cluster
          schema:
            type: string
      responses:
        '200':
          description: Returns the status of provided cluster
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/IsClusterConnectedResponse'
        '401':
          description: 'Error: Unauthorized'
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
      security:
        - jwt: []
components:
  schemas:
    IsClusterConnectedResponse:
      type: object
      properties:
        isConnected:
          type: boolean
          description: Whether the cluster is connected
      required:
        - isConnected
      title: IsClusterConnectedResponse
      x-fern-type-name: IsClusterConnectedResponse
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