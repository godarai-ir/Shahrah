> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Mark stage failure

> Mark a staged artifact version as failed.



## OpenAPI

````yaml /openapi.json post /api/ml/v1/artifact-versions/mark-stage-failure
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
  /api/ml/v1/artifact-versions/mark-stage-failure:
    post:
      tags:
        - Artifacts
      summary: Mark stage failure
      description: Mark a staged artifact version as failed.
      operationId: mark_stage_failure_v1_artifact_versions_mark_stage_failure_post
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/MarkStageArtifactFailureRequest'
        required: true
      responses:
        '200':
          description: Empty response indicating successful failure marking
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
    MarkStageArtifactFailureRequest:
      properties:
        id:
          type: string
          title: Id
          description: ID of the staged artifact version to mark as failed
      type: object
      required:
        - id
      title: MarkStageArtifactFailureRequest
      x-fern-type-name: MarkStageArtifactFailureRequest
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