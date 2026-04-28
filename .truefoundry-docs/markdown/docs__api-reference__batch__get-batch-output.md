> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Get Batch Output

> Get output of a specific batch process



## OpenAPI

````yaml /gateway-openapi.json get /batches/{id}/output
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
  /batches/{id}/output:
    get:
      tags:
        - Batch
      summary: Get Batch Output
      description: Get output of a specific batch process
      parameters:
        - schema:
            type: string
            description: Batch Id
          required: true
          name: id
          in: path
        - schema:
            type: string
            description: Optional metadata for the request
          required: false
          name: x-tfy-metadata
          in: header
      responses:
        '200':
          description: Batch output retrieved successfully
          content:
            application/json:
              schema:
                type: array
                items:
                  type: object
                  properties:
                    id:
                      type: string
                      description: Batch request ID.
                    custom_id:
                      type: string
                      description: Custom ID for the request.
                    response:
                      type: object
                      properties:
                        status_code:
                          type: number
                          description: HTTP status code of the response.
                        request_id:
                          type: string
                          description: Unique request ID.
                        body:
                          type: object
                          properties:
                            id:
                              type: string
                              description: ID of the chat completion.
                            object:
                              type: string
                              description: Object type.
                            created:
                              type: number
                              description: Creation timestamp.
                            model:
                              type: string
                              description: Model used.
                            choices:
                              type: array
                              items:
                                type: object
                                properties:
                                  index:
                                    type: number
                                    description: Index of the choice.
                                  message:
                                    type: object
                                    properties:
                                      role:
                                        type: string
                                        description: Role of the message sender.
                                      content:
                                        type: string
                                        description: Message content.
                                      refusal:
                                        nullable: true
                                        description: Refusal details, if any.
                                      annotations:
                                        type: array
                                        items:
                                          nullable: true
                                        description: Annotations for the message.
                                    required:
                                      - role
                                      - content
                                      - annotations
                                  logprobs:
                                    nullable: true
                                    description: Log probabilities.
                                  finish_reason:
                                    type: string
                                    description: Reason for finishing the response.
                                required:
                                  - index
                                  - message
                                  - finish_reason
                            usage:
                              type: object
                              properties:
                                prompt_tokens:
                                  type: number
                                  description: Number of prompt tokens.
                                completion_tokens:
                                  type: number
                                  description: Number of completion tokens.
                                total_tokens:
                                  type: number
                                  description: Total number of tokens.
                                prompt_tokens_details:
                                  type: object
                                  properties:
                                    cached_tokens:
                                      type: number
                                      description: Number of cached tokens.
                                    audio_tokens:
                                      type: number
                                      description: Number of audio tokens.
                                  required:
                                    - cached_tokens
                                    - audio_tokens
                                completion_tokens_details:
                                  type: object
                                  properties:
                                    reasoning_tokens:
                                      type: number
                                      description: Number of reasoning tokens.
                                    audio_tokens:
                                      type: number
                                      description: Number of audio tokens.
                                    accepted_prediction_tokens:
                                      type: number
                                      description: Number of accepted prediction tokens.
                                    rejected_prediction_tokens:
                                      type: number
                                      description: Number of rejected prediction tokens.
                                  required:
                                    - reasoning_tokens
                                    - audio_tokens
                                    - accepted_prediction_tokens
                                    - rejected_prediction_tokens
                              required:
                                - prompt_tokens
                                - completion_tokens
                                - total_tokens
                                - prompt_tokens_details
                                - completion_tokens_details
                            service_tier:
                              type: string
                              description: Service tier.
                            system_fingerprint:
                              nullable: true
                              description: System fingerprint.
                          required:
                            - id
                            - object
                            - created
                            - model
                            - choices
                            - usage
                            - service_tier
                        error:
                          nullable: true
                          description: Error details, if any.
                      required:
                        - status_code
                        - request_id
                        - body
                      description: Response details.
                    error:
                      nullable: true
                      description: Error details, if any.
                  required:
                    - id
                    - custom_id
                    - response
        '404':
          description: Batch not found
components:
  securitySchemes:
    AuthorizationBearer:
      type: http
      scheme: bearer
      bearerFormat: JWT

````