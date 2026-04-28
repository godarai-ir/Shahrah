> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Model Responses

> Generate model responses using the specified model.



## OpenAPI

````yaml /gateway-openapi.json post /responses
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
  /responses:
    post:
      tags:
        - Responses
      summary: Model Responses
      description: Generate model responses using the specified model.
      parameters:
        - schema:
            type: string
            description: Optional metadata for the request
          required: false
          name: x-tfy-metadata
          in: header
      requestBody:
        required: true
        description: Parameters for generating model responses.
        content:
          application/json:
            schema:
              type: object
              properties:
                model:
                  type: string
                  description: Model identifier to generate the response
                input:
                  nullable: true
                  description: Text, image, or file inputs to the model
                background:
                  type: boolean
                  nullable: true
                  description: Whether to run the model response in the background
                include:
                  type: array
                  nullable: true
                  items:
                    type: string
                  description: Additional output data to include
                instructions:
                  type: string
                  nullable: true
                  description: System message as first item in context
                max_output_tokens:
                  type: number
                  nullable: true
                  description: Upper bound for tokens generated
                metadata:
                  type: object
                  nullable: true
                  additionalProperties:
                    type: string
                  description: Key-value pairs for additional information
                parallel_tool_calls:
                  type: boolean
                  nullable: true
                  description: Allow parallel tool calls
                previous_response_id:
                  type: string
                  nullable: true
                  description: ID of previous response for multi-turn
                reasoning:
                  type: object
                  nullable: true
                  properties:
                    effort:
                      type: string
                      nullable: true
                      description: Reasoning effort level
                  description: Configuration for reasoning models
                service_tier:
                  type: string
                  nullable: true
                  enum:
                    - auto
                    - default
                    - flex
                    - priority
                  description: Latency tier for processing
                store:
                  type: boolean
                  nullable: true
                  description: Whether to store the response
                stream:
                  type: boolean
                  nullable: true
                  description: Enable streaming response
                temperature:
                  type: number
                  nullable: true
                  description: Sampling temperature between 0 and 2
                text:
                  type: object
                  properties:
                    format:
                      type: object
                      properties:
                        type:
                          type: string
                          enum:
                            - text
                            - json_object
                          description: Response format type
                      required:
                        - type
                      description: Text format configuration
                  description: Text response configuration
                tool_choice:
                  anyOf:
                    - type: string
                      enum:
                        - none
                    - type: string
                      enum:
                        - auto
                    - type: string
                      enum:
                        - required
                    - type: object
                      properties:
                        type:
                          type: string
                          enum:
                            - function
                        function:
                          type: object
                          properties:
                            name:
                              type: string
                          required:
                            - name
                      required:
                        - type
                        - function
                  description: Tool selection behavior
                tools:
                  type: array
                  items:
                    nullable: true
                  description: Available tools for the model
                top_p:
                  type: number
                  nullable: true
                  description: Nucleus sampling parameter
                truncation:
                  type: string
                  nullable: true
                  enum:
                    - auto
                    - disabled
                  description: Truncation strategy
                user:
                  type: string
                  description: End-user identifier
              required:
                - model
      responses:
        '200':
          description: Model Response generated successfully.
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: string
                    description: Response ID.
                  object:
                    type: string
                    description: Object type.
                  created_at:
                    type: number
                    description: Creation timestamp.
                  status:
                    type: string
                    description: Response status.
                  error:
                    nullable: true
                    description: Error details, if any.
                  incomplete_details:
                    nullable: true
                    description: Details about incomplete responses, if any.
                  instructions:
                    nullable: true
                    description: Instructions provided for the response.
                  max_output_tokens:
                    type: number
                    nullable: true
                    description: Maximum output tokens allowed.
                  model:
                    type: string
                    description: Model used for the response.
                  output:
                    type: array
                    items:
                      type: object
                      properties:
                        id:
                          type: string
                          description: Message ID.
                        type:
                          type: string
                          description: Type of the message.
                        status:
                          type: string
                          description: Message status.
                        content:
                          type: array
                          items:
                            type: object
                            properties:
                              type:
                                type: string
                                description: Content type.
                              annotations:
                                type: array
                                items:
                                  nullable: true
                                description: Annotations for the content.
                              text:
                                type: string
                                description: Text content.
                            required:
                              - type
                              - annotations
                              - text
                        role:
                          type: string
                          description: Role of the message sender.
                      required:
                        - id
                        - type
                        - status
                        - content
                        - role
                  parallel_tool_calls:
                    type: boolean
                    description: Indicates if parallel tool calls were used.
                  previous_response_id:
                    type: string
                    nullable: true
                    description: ID of the previous response, if any.
                  reasoning:
                    type: object
                    nullable: true
                    properties:
                      effort:
                        nullable: true
                        description: Effort details.
                      summary:
                        nullable: true
                        description: Reasoning summary.
                    description: Reasoning details.
                  service_tier:
                    type: string
                    description: Service tier.
                  store:
                    type: boolean
                    description: Indicates if the response is stored.
                  temperature:
                    type: number
                    description: Temperature setting for the model.
                  text:
                    type: object
                    properties:
                      format:
                        type: object
                        properties:
                          type:
                            type: string
                            description: Text format type.
                        required:
                          - type
                    required:
                      - format
                  tool_choice:
                    type: string
                    description: Tool choice used.
                  tools:
                    type: array
                    items:
                      nullable: true
                    description: Tools used in the response.
                  top_p:
                    type: number
                    description: Top-p sampling parameter.
                  truncation:
                    type: string
                    description: Truncation setting.
                  usage:
                    type: object
                    properties:
                      input_tokens:
                        type: number
                        description: Number of input tokens.
                      input_tokens_details:
                        type: object
                        properties:
                          cached_tokens:
                            type: number
                            description: Number of cached tokens.
                        required:
                          - cached_tokens
                      output_tokens:
                        type: number
                        description: Number of output tokens.
                      output_tokens_details:
                        type: object
                        properties:
                          reasoning_tokens:
                            type: number
                            description: Number of reasoning tokens.
                        required:
                          - reasoning_tokens
                      total_tokens:
                        type: number
                        description: Total number of tokens.
                    required:
                      - input_tokens
                      - input_tokens_details
                      - output_tokens
                      - output_tokens_details
                      - total_tokens
                  user:
                    nullable: true
                    description: User details, if any.
                  metadata:
                    type: object
                    nullable: true
                    properties: {}
                    description: Additional metadata.
                  provider:
                    type: string
                    description: Provider of the response.
                required:
                  - id
                  - object
                  - created_at
                  - status
                  - max_output_tokens
                  - model
                  - output
                  - parallel_tool_calls
                  - previous_response_id
                  - reasoning
                  - service_tier
                  - store
                  - temperature
                  - text
                  - tool_choice
                  - tools
                  - top_p
                  - truncation
                  - usage
                  - metadata
                  - provider
        '400':
          description: Bad request.
components:
  securitySchemes:
    AuthorizationBearer:
      type: http
      scheme: bearer
      bearerFormat: JWT

````