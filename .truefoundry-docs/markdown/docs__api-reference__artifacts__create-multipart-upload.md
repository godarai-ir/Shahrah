> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Create multipart upload

> Create a multipart upload for large files in an artifact version.



## OpenAPI

````yaml /openapi.json post /api/ml/v1/artifact-versions/signed-urls/multipart
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
  /api/ml/v1/artifact-versions/signed-urls/multipart:
    post:
      tags:
        - Artifacts
      summary: Create multipart upload
      description: Create a multipart upload for large files in an artifact version.
      operationId: create_multi_part_upload_v1_artifact_versions_signed_urls_multipart_post
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateMultiPartUploadRequest'
        required: true
      responses:
        '200':
          description: Multipart upload information including signed URLs for each part
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/MultiPartUploadResponse'
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
    CreateMultiPartUploadRequest:
      properties:
        id:
          type: string
          title: Id
          description: ID of the artifact version to upload files to
        path:
          type: string
          title: Path
          description: >-
            Relative path within the artifact version where the file should be
            uploaded
        num_parts:
          type: integer
          title: Num Parts
          description: Number of parts to split the upload into for multipart upload
      type: object
      required:
        - id
        - path
        - num_parts
      title: CreateMultiPartUploadRequest
      x-fern-type-name: CreateMultiPartUploadRequest
    MultiPartUploadResponse:
      properties:
        data:
          allOf:
            - $ref: '#/components/schemas/MultiPartUpload'
          title: Data
          description: Multipart upload information including signed URLs for each part
      type: object
      required:
        - data
      title: MultiPartUploadResponse
      x-fern-type-name: MultiPartUploadResponse
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
    MultiPartUpload:
      properties:
        storage_provider:
          allOf:
            - $ref: '#/components/schemas/MultiPartUploadStorageProvider'
          description: >-
            Storage provider being used for the multipart upload (e.g.,
            'S3_COMPATIBLE' or 'AZURE_BLOB')
        part_signed_urls:
          items:
            $ref: '#/components/schemas/SignedURL'
          type: array
          title: Part Signed Urls
          description: List of signed URLs for each part of the multipart upload
        s3_compatible_upload_id:
          type: string
          title: S3 Compatible Upload Id
          description: Upload ID for S3-compatible storage providers
        azure_blob_block_ids:
          items:
            type: string
          type: array
          title: Azure Blob Block Ids
          description: List of block IDs for Azure Blob Storage multipart upload
        finalize_signed_url:
          allOf:
            - $ref: '#/components/schemas/SignedURL'
          title: Finalize Signed Url
          description: >-
            Signed URL to call after all parts are uploaded to finalize the
            multipart upload
      type: object
      required:
        - storage_provider
        - part_signed_urls
        - finalize_signed_url
      title: MultiPartUpload
      x-fern-type-name: MultiPartUpload
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
    MultiPartUploadStorageProvider:
      type: string
      enum:
        - S3_COMPATIBLE
        - AZURE_BLOB
      title: MultiPartUploadStorageProvider
      description: An enumeration.
      x-fern-type-name: MultiPartUploadStorageProvider
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
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````