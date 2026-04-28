> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Delete ML Repo

> Delete an ML Repo by its ID.



## OpenAPI

````yaml /openapi.json delete /api/ml/v1/ml-repos/{id}
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
  /api/ml/v1/ml-repos/{id}:
    delete:
      tags:
        - MLRepos
      summary: Delete ML Repo
      description: Delete an ML Repo by its ID.
      operationId: delete_ml_repo_v1_ml_repos__id__delete
      parameters:
        - required: true
          schema:
            type: string
            title: Id
          name: id
          in: path
      responses:
        '200':
          description: Empty response indicating successful deletion
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/EmptyResponse'
        '422':
          description: Validation Error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HTTPValidationError'
      security:
        - jwt: []
components:
  schemas:
    EmptyResponse:
      properties: {}
      type: object
      title: EmptyResponse
      x-fern-type-name: EmptyResponse
    HTTPValidationError:
      properties:
        detail:
          items:
            $ref: '#/components/schemas/ValidationError'
          type: array
          title: Detail
      type: object
      title: HTTPValidationError
      x-fern-type-name: HTTPValidationError
    ValidationError:
      properties:
        loc:
          items:
            anyOf:
              - type: string
              - type: integer
          type: array
          title: Location
        msg:
          type: string
          title: Message
        type:
          type: string
          title: Error Type
      type: object
      required:
        - loc
        - msg
        - type
      title: ValidationError
      x-fern-type-name: ValidationError
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````