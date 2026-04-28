> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Create Batch

> Creates a new batch process



## OpenAPI

````yaml /gateway-openapi.json post /batches
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
    post:
      tags:
        - Batch
      summary: Create Batch
      description: Creates a new batch process
      parameters:
        - schema:
            type: string
            description: Optional metadata for the request
          required: false
          name: x-tfy-metadata
          in: header
      requestBody:
        required: true
        description: Batch processing parameters
        content:
          application/json:
            schema:
              anyOf:
                - type: object
                  properties:
                    input_file_id:
                      type: string
                      description: The ID of the input file to process.
                    completion_window:
                      type: string
                      nullable: true
                      description: Time window for completion.
                    endpoint:
                      type: string
                      nullable: true
                      description: The API endpoint to be used.
                  required:
                    - input_file_id
                  title: OpenAI
                - type: object
                  properties:
                    input_file_id:
                      type: string
                      description: The ID of the input file to process.
                    completion_window:
                      type: string
                      nullable: true
                      description: Time window for completion.
                    endpoint:
                      type: string
                      nullable: true
                      description: The API endpoint to be used.
                    job_name:
                      type: string
                      nullable: true
                      description: Name for the job.
                  required:
                    - input_file_id
                  title: Vertex AI
                - type: object
                  properties:
                    model:
                      type: string
                      description: The model to use for the batch job.
                    role_arn:
                      type: string
                      description: AWS IAM role ARN for the batch job.
                    input_file_id:
                      type: string
                      description: The ID of the input file to process.
                    endpoint:
                      type: string
                      nullable: true
                      description: The API endpoint to be used.
                    job_name:
                      type: string
                      nullable: true
                      description: Name for the job.
                    output_data_config:
                      nullable: true
                      description: >-
                        Batch job output storage location, will be constructed
                        based on input_file_id if not provided.
                  required:
                    - model
                    - role_arn
                    - input_file_id
                  title: Bedrock
      responses:
        '200':
          description: Batch created successfully
          content:
            application/json:
              schema:
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
                    nullable: true
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
        '400':
          description: Bad request
components:
  securitySchemes:
    AuthorizationBearer:
      type: http
      scheme: bearer
      bearerFormat: JWT

````