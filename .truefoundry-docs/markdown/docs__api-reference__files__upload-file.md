> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Upload File

> Upload a new file



## OpenAPI

````yaml /gateway-openapi.json post /files
openapi: 3.0.0
info:
  title: LLM Gateway
  description: API for LLM Gateway
  version: 1.0.0
servers:
  - url: https://{controlPlaneURL}/api/llm
    variables:
      controlPlaneURL:
        default: app.truefoundry.com
        description: Control Plane URL
    description: Gateway base URL
security:
  - AuthorizationBearer: []
tags:
  - name: Chat
  - name: Agent
  - name: MCP
  - name: Embeddings
  - name: Rerank
  - name: Responses
  - name: Image
  - name: Audio
  - name: Batch
  - name: Files
  - name: Moderations
  - name: Models
paths:
  /files:
    post:
      tags:
        - Files
      summary: Upload File
      description: Upload a new file
      parameters:
        - schema:
            type: string
            description: Name of the provider
          required: true
          name: x-tfy-provider-name
          in: header
        - schema:
            type: string
            description: Optional metadata for the request
          required: false
          name: x-tfy-metadata
          in: header
      requestBody:
        required: true
        description: File to upload and associated metadata
        content:
          multipart/form-data:
            schema:
              type: object
              properties:
                file:
                  nullable: true
                  description: The file data to upload (not the file name)
      responses:
        '200':
          description: File uploaded successfully
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: string
                    description: Id of the uploaded file.
                  object:
                    type: string
                    description: Type of the object, e.g., 'file'.
                  created_at:
                    type: number
                    description: Timestamp of when the file was created.
                  purpose:
                    type: string
                    description: The purpose of the file, e.g., 'batch'.
                  bytes:
                    type: number
                    description: The size of the file in bytes.
                  status:
                    type: string
                    description: The status of the file, e.g., 'processed'.
                required:
                  - id
                  - object
                  - created_at
                  - purpose
                  - bytes
                  - status
        '400':
          description: Bad Request
components:
  securitySchemes:
    AuthorizationBearer:
      type: http
      scheme: bearer
      bearerFormat: JWT

````