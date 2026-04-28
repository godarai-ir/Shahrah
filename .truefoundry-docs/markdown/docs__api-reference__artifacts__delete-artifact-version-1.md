> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Delete artifact version

> Delete an artifact version by manifest.



## OpenAPI

````yaml /openapi.json post /api/ml/v1/artifact-versions
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
  /api/ml/v1/artifact-versions:
    post:
      tags:
        - Artifacts
      summary: Delete artifact version
      description: Delete an artifact version by manifest.
      operationId: delete_artifact_version_v1_artifact_versions_post
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/DeleteArtifactRequest'
        required: true
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
    DeleteArtifactRequest:
      properties:
        manifest:
          allOf:
            - $ref: '#/components/schemas/ArtifactManifest'
          title: Manifest
          description: Manifest identifying the artifact to delete
      type: object
      required:
        - manifest
      title: DeleteArtifactRequest
      x-fern-type-name: DeleteArtifactRequest
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
    ArtifactManifest:
      properties:
        name:
          type: string
          maxLength: 256
          pattern: ^[A-Za-z0-9_\-]+$
          title: Name
          description: >-
            Name of the artifact (alphanumeric characters, hyphens, and
            underscores only, max 256 characters)
        metadata:
          type: object
          title: Metadata
          description: >-
            Key value metadata. Should be valid JSON. For e.g.
            `{"business-unit": "sales", "quality": "good", "rating": 4.5}`
        ml_repo:
          type: string
          pattern: ^[a-zA-Z][a-zA-Z0-9\-]{1,98}[a-zA-Z0-9]$
          title: Ml Repo
          description: >-
            Name of the ML Repo that this artifact belongs to (must start and
            end with alphanumeric, 2-100 characters)
        version:
          type: integer
          minimum: 1
          title: Version
          description: Version of the entity
        type:
          type: string
          enum:
            - artifact-version
          title: Type
          default: artifact-version
        description:
          type: string
          maxLength: 512
          title: Description
        version_alias:
          type: string
          maxLength: 128
          pattern: ^v[a-zA-Z0-9.-]*([a-zA-Z0-9]+)$
          title: Version Alias
          description: >-
            Version alias is alternate, ideally human readable, version string
            to reference an artifact version. It should start with `v` followed
            by alphanumeric and it can include `.` and `-` in between (e.g.
            `v1.0.0`, `v1-prod`, `v3-dev`, etc)
        source:
          oneOf:
            - $ref: '#/components/schemas/TrueFoundryManagedSource'
            - $ref: '#/components/schemas/ExternalBlobStorageSource'
            - $ref: '#/components/schemas/LocalArtifactSource'
          title: Source
          discriminator:
            propertyName: type
            mapping:
              truefoundry:
                $ref: '#/components/schemas/TrueFoundryManagedSource'
              external:
                $ref: '#/components/schemas/ExternalBlobStorageSource'
              local:
                $ref: '#/components/schemas/LocalArtifactSource'
        step:
          type: integer
          minimum: 0
          title: Step
          description: >-
            Step/Epoch number in an iterative training loop the artifact version
            was created. Generally useful when logging a model version from a
            MLRepo Run
          default: 0
        run_id:
          type: string
          title: Run Id
          description: ID of the MLRepo Run that generated the artifact version
      type: object
      required:
        - name
        - metadata
        - ml_repo
        - source
      title: ArtifactManifest
      description: Artifact Version manifest.
      x-fern-type-name: ArtifactManifest
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
    TrueFoundryManagedSource:
      properties:
        type:
          type: string
          enum:
            - truefoundry
          title: Type
          description: Type of the source
        uri:
          type: string
          pattern: ^(s3|gs|wasbs)://.*
          title: Uri
      type: object
      required:
        - type
      title: TrueFoundryManagedSource
      x-fern-type-name: TrueFoundryManagedSource
    ExternalBlobStorageSource:
      properties:
        type:
          type: string
          enum:
            - external
          title: Type
          description: Type of the source
        uri:
          type: string
          pattern: ^(s3|gs|wasbs)://.*
          title: Uri
          description: >-
            URI referencing a path in the blob storage bucket linked to the
            MLRepo
      type: object
      required:
        - type
        - uri
      title: ExternalBlobStorageSource
      x-fern-type-name: ExternalBlobStorageSource
    LocalArtifactSource:
      properties:
        type:
          type: string
          enum:
            - local
          title: Type
          description: Type of the source
        paths:
          items:
            $ref: '#/components/schemas/ArtifactPath'
          type: array
          title: Paths
          description: >-
            Array of ArtifactPath objects representing the source and
            destination paths
      type: object
      required:
        - type
        - paths
      title: LocalArtifactSource
      x-fern-type-name: LocalArtifactSource
    ArtifactPath:
      properties:
        src:
          type: string
          title: Src
          description: Local file or folder path
        dest:
          type: string
          title: Dest
          description: >-
            Relative path where the file or folder will be uploaded to in the
            artifact
      type: object
      required:
        - src
      title: ArtifactPath
      x-fern-type-name: ArtifactPath
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````