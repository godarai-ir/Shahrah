> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Apply tags to artifact version

> Apply tags to an artifact version.



## OpenAPI

````yaml /openapi.json put /api/ml/v1/artifact-versions/tags
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
  /api/ml/v1/artifact-versions/tags:
    put:
      tags:
        - Artifacts
      summary: Apply tags to artifact version
      description: Apply tags to an artifact version.
      operationId: apply_artifact_version_tags_v1_artifact_versions_tags_put
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ApplyArtifactVersionTagsRequest'
        required: true
      responses:
        '200':
          description: Empty response indicating successful tag application
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
    ApplyArtifactVersionTagsRequest:
      properties:
        artifact_version_id:
          type: string
          title: Artifact Version Id
          description: ID of the artifact version to apply tags to
        tags:
          items:
            type: string
          type: array
          title: Tags
          description: List of tags to apply to the artifact version
        force:
          type: boolean
          title: Force
          description: Whether to overwrite existing tags if they conflict
          default: false
      type: object
      required:
        - artifact_version_id
        - tags
      title: ApplyArtifactVersionTagsRequest
      x-fern-type-name: ApplyArtifactVersionTagsRequest
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