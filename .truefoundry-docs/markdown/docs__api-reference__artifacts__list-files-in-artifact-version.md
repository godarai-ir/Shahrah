> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# List files in artifact version

> List files and directories in an artifact version.



## OpenAPI

````yaml /openapi.json post /api/ml/v1/artifact-versions/files
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
  /api/ml/v1/artifact-versions/files:
    post:
      tags:
        - Artifacts
      summary: List files in artifact version
      description: List files and directories in an artifact version.
      operationId: list_files_for_artifact_version_v1_artifact_versions_files_post
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ListFilesRequest'
        required: true
      responses:
        '200':
          description: List of files and directories with pagination information
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ListFilesResponse'
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
    ListFilesRequest:
      properties:
        id:
          type: string
          title: Id
          description: ID of the artifact version to list files from
        path:
          type: string
          title: Path
          description: >-
            Relative path within the artifact version to list files from
            (defaults to root)
        limit:
          type: integer
          title: Limit
          description: Maximum number of files/directories to return
        pageToken:
          type: string
          title: Pagetoken
          description: Token to retrieve the next page of results
      type: object
      required:
        - id
      title: ListFilesRequest
      x-fern-type-name: ListFilesRequest
    ListFilesResponse:
      properties:
        data:
          items:
            $ref: '#/components/schemas/FileInfo'
          type: array
          title: Data
          description: List of files and directories in the artifact version
        pagination:
          allOf:
            - $ref: '#/components/schemas/TokenPagination'
          title: Pagination
          description: Pagination information including page tokens for navigation
      type: object
      required:
        - data
        - pagination
      title: ListFilesResponse
      x-fern-type-name: ListFilesResponse
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
    FileInfo:
      properties:
        path:
          type: string
          title: Path
          description: Relative path of the file or directory within the artifact version
        is_dir:
          type: boolean
          title: Is Dir
          description: Whether this path represents a directory (true) or a file (false)
        file_size:
          type: integer
          title: File Size
          description: Size of the file in bytes (only present for files, not directories)
        signed_url:
          type: string
          title: Signed Url
          description: >-
            Pre-signed URL to download the file directly (only present for
            files)
        last_modified:
          type: string
          format: date-time
          title: Last Modified
          description: Timestamp when the file was last modified
      type: object
      required:
        - path
        - is_dir
      title: FileInfo
      x-fern-type-name: FileInfo
    TokenPagination:
      type: object
      properties:
        limit:
          type: integer
          description: Number of items per page
          example: 10
        nextPageToken:
          type: string
          description: Base64 encoded token for the next page
        previousPageToken:
          type: string
          description: Base64 encoded token for the previous page
      title: TokenPagination
      x-fern-type-name: TokenPagination
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