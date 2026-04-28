> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# List Files

> List all files



## OpenAPI

````yaml /gateway-openapi.json get /files
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
    get:
      tags:
        - Files
      summary: List Files
      description: List all files
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
      responses:
        '200':
          description: All files retrieved successfully
          content:
            application/json:
              schema:
                type: object
                properties:
                  object:
                    type: string
                    description: Type of the object, e.g., 'list'.
                  data:
                    type: array
                    items:
                      type: object
                      properties:
                        object:
                          type: string
                          description: Type of the object, e.g., 'file'.
                        id:
                          type: string
                          description: Unique identifier for the file.
                        purpose:
                          type: string
                          description: >-
                            The purpose of the file, e.g., 'batch',
                            'assistants'.
                        filename:
                          type: string
                          description: The name of the file.
                        bytes:
                          type: number
                          description: The size of the file in bytes.
                        created_at:
                          type: number
                          description: Timestamp of when the file was created.
                        expires_at:
                          type: number
                          nullable: true
                          description: >-
                            Timestamp of when the file expires, or null if it
                            does not expire.
                        status:
                          type: string
                          description: The status of the file, e.g., 'processed'.
                        status_details:
                          type: string
                          nullable: true
                          description: >-
                            Additional details about the file's status, if
                            available.
                      required:
                        - object
                        - id
                        - purpose
                        - filename
                        - bytes
                        - created_at
                        - expires_at
                        - status
                        - status_details
                    description: Array of file objects.
                  has_more:
                    type: boolean
                    description: Whether there are more files to retrieve.
                  first_id:
                    type: string
                    nullable: true
                    description: The ID of the first file in the list.
                  last_id:
                    type: string
                    nullable: true
                    description: The ID of the last file in the list.
                required:
                  - object
                  - data
                  - has_more
                  - first_id
                  - last_id
components:
  securitySchemes:
    AuthorizationBearer:
      type: http
      scheme: bearer
      bearerFormat: JWT

````