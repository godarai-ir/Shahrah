> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Get File

> Get information about a specific file



## OpenAPI

````yaml /gateway-openapi.json get /files/{id}
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
  /files/{id}:
    get:
      tags:
        - Files
      summary: Get File
      description: Get information about a specific file
      parameters:
        - schema:
            type: string
            description: The ID of the file to retrieve
          required: true
          name: id
          in: path
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
      responses:
        '200':
          description: Files retrieved successfully
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
        '404':
          description: File not found
components:
  securitySchemes:
    AuthorizationBearer:
      type: http
      scheme: bearer
      bearerFormat: JWT

````