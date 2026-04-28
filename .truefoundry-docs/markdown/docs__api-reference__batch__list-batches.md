> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# List batches

> List all batches



## OpenAPI

````yaml /gateway-openapi.json get /batches
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
  /batches:
    get:
      tags:
        - Batch
      summary: List batches
      description: List all batches
      parameters:
        - schema:
            type: string
            description: Optional metadata for the request
          required: false
          name: x-tfy-metadata
          in: header
      responses:
        '200':
          description: Batches retrieved successfully
          content:
            application/json:
              schema:
                type: object
                properties:
                  object:
                    type: string
                    description: Object type.
                  data:
                    type: array
                    items:
                      type: object
                      properties:
                        id:
                          type: string
                          description: Batch ID.
                        object:
                          type: string
                          description: Object type.
                        endpoint:
                          type: string
                          description: API endpoint.
                        errors:
                          type: object
                          nullable: true
                          properties:
                            object:
                              type: string
                              description: Object type for errors.
                            data:
                              type: array
                              items:
                                type: object
                                properties:
                                  code:
                                    type: string
                                    description: Error code.
                                  message:
                                    type: string
                                    description: Error message.
                                  param:
                                    type: string
                                    description: Parameter related to the error.
                                  line:
                                    type: number
                                    description: Line number where the error occurred.
                                required:
                                  - code
                                  - message
                                  - param
                                  - line
                              description: List of errors.
                          required:
                            - object
                            - data
                          description: Errors associated with the batch.
                        input_file_id:
                          type: string
                          description: Input file ID.
                        completion_window:
                          type: string
                          description: Completion window.
                        status:
                          type: string
                          description: Batch status.
                        output_file_id:
                          type: string
                          nullable: true
                          description: Output file ID.
                        error_file_id:
                          type: string
                          nullable: true
                          description: Error file ID.
                        created_at:
                          type: number
                          description: Creation timestamp.
                        in_progress_at:
                          type: number
                          nullable: true
                          description: In-progress timestamp.
                        expires_at:
                          type: number
                          description: Expiration timestamp.
                        finalizing_at:
                          type: number
                          nullable: true
                          description: Finalizing timestamp.
                        completed_at:
                          type: number
                          nullable: true
                          description: Completion timestamp.
                        failed_at:
                          type: number
                          nullable: true
                          description: Failure timestamp.
                        expired_at:
                          type: number
                          nullable: true
                          description: Expiration timestamp.
                        cancelling_at:
                          type: number
                          nullable: true
                          description: Cancelling timestamp.
                        cancelled_at:
                          type: number
                          nullable: true
                          description: Cancelled timestamp.
                        request_counts:
                          type: object
                          properties:
                            total:
                              type: number
                              description: Total number of requests.
                            completed:
                              type: number
                              description: Number of completed requests.
                            failed:
                              type: number
                              description: Number of failed requests.
                          required:
                            - total
                            - completed
                            - failed
                          description: Request counts.
                        metadata:
                          nullable: true
                          description: Additional metadata.
                      required:
                        - id
                        - object
                        - endpoint
                        - errors
                        - input_file_id
                        - completion_window
                        - status
                        - output_file_id
                        - error_file_id
                        - created_at
                        - in_progress_at
                        - expires_at
                        - finalizing_at
                        - completed_at
                        - failed_at
                        - expired_at
                        - cancelling_at
                        - cancelled_at
                        - request_counts
                    description: List of batches.
                  first_id:
                    type: string
                    description: ID of the first batch.
                  last_id:
                    type: string
                    description: ID of the last batch.
                  has_more:
                    type: boolean
                    description: Indicates if there are more batches.
                required:
                  - object
                  - data
                  - first_id
                  - last_id
                  - has_more
components:
  securitySchemes:
    AuthorizationBearer:
      type: http
      scheme: bearer
      bearerFormat: JWT

````