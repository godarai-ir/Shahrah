> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Get signed URLs for artifact version

> Get pre-signed URLs for reading or writing files in an artifact version.



## OpenAPI

````yaml /openapi.json post /api/ml/v1/artifact-versions/signed-urls
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
  /api/ml/v1/artifact-versions/signed-urls:
    post:
      tags:
        - Artifacts
      summary: Get signed URLs for artifact version
      description: Get pre-signed URLs for reading or writing files in an artifact version.
      operationId: >-
        get_signed_urls_for_artifact_version_v1_artifact_versions_signed_urls_post
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/GetSignedURLsRequest'
        required: true
      responses:
        '200':
          description: List of signed URLs for the requested file paths
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GetSignedURLsResponse'
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
    GetSignedURLsRequest:
      properties:
        id:
          type: string
          title: Id
          description: ID of the artifact version to get signed URLs for
        paths:
          items:
            type: string
          type: array
          title: Paths
          description: >-
            List of relative file paths within the artifact version to get
            signed URLs for
        operation:
          allOf:
            - $ref: '#/components/schemas/Operation'
          description: Operation type for the signed URL (e.g., 'READ' or 'WRITE')
      type: object
      required:
        - id
        - paths
        - operation
      title: GetSignedURLsRequest
      x-fern-type-name: GetSignedURLsRequest
    GetSignedURLsResponse:
      properties:
        data:
          items:
            $ref: '#/components/schemas/SignedURL'
          type: array
          title: Data
          description: List of signed URLs for the requested file paths
      type: object
      required:
        - data
      title: GetSignedURLsResponse
      x-fern-type-name: GetSignedURLsResponse
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
    Operation:
      type: string
      enum:
        - READ
        - WRITE
      title: Operation
      description: An enumeration.
      x-fern-type-name: Operation
    SignedURL:
      properties:
        path:
          type: string
          title: Path
          description: Relative path of the file within the artifact storage
        signed_url:
          type: string
          title: Signed Url
          description: Pre-signed URL that can be used to access the file directly
      type: object
      required:
        - path
        - signed_url
      title: SignedURL
      x-fern-type-name: SignedURL
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