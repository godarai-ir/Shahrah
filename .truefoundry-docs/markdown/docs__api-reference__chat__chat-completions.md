> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Chat Completions

> Generate chat-based completions using the specified model.



## OpenAPI

````yaml /gateway-openapi.json post /chat/completions
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
  /chat/completions:
    post:
      tags:
        - Chat
      summary: Chat Completions
      description: Generate chat-based completions using the specified model.
      parameters:
        - schema:
            type: string
            description: Optional metadata for the request
          required: false
          name: x-tfy-metadata
          in: header
      requestBody:
        required: true
        description: Parameters for generating chat completions.
        content:
          application/json:
            schema:
              type: object
              properties:
                model:
                  type: string
                  description: Identifier for the language model to be used for generation.
                messages:
                  type: array
                  items:
                    oneOf:
                      - type: object
                        properties:
                          role:
                            type: string
                            enum:
                              - system
                            description: >-
                              System message providing context or instructions
                              to the model.
                          content:
                            anyOf:
                              - type: string
                              - type: array
                                items:
                                  type: object
                                  properties:
                                    type:
                                      type: string
                                      description: >-
                                        The type of content being transmitted.
                                        Supported values include "text" and
                                        "image".
                                    text:
                                      type: string
                                      description: >-
                                        The text content when type is "text".
                                        UTF-8 encoded string.
                                    image_url:
                                      type: object
                                      properties:
                                        url:
                                          type: string
                                          description: >-
                                            The URL pointing to the image resource.
                                            Must be a valid, accessible URL.
                                      required:
                                        - url
                                      description: >-
                                        Image URL configuration when type is
                                        "image".
                                  required:
                                    - type
                            description: >-
                              System instructions or context. Can be a string or
                              an array.
                          name:
                            type: string
                            description: Optional identifier for the system.
                        required:
                          - role
                          - content
                        title: System Message
                      - type: object
                        properties:
                          role:
                            type: string
                            enum:
                              - user
                            description: Message from the end user making the request.
                          content:
                            anyOf:
                              - type: string
                              - type: array
                                items:
                                  type: object
                                  properties:
                                    type:
                                      type: string
                                      description: >-
                                        The type of content being transmitted.
                                        Supported values include "text" and
                                        "image".
                                    text:
                                      type: string
                                      description: >-
                                        The text content when type is "text".
                                        UTF-8 encoded string.
                                    image_url:
                                      type: object
                                      properties:
                                        url:
                                          type: string
                                          description: >-
                                            The URL pointing to the image resource.
                                            Must be a valid, accessible URL.
                                      required:
                                        - url
                                      description: >-
                                        Image URL configuration when type is
                                        "image".
                                  required:
                                    - type
                            description: User input as text or structured content array.
                          name:
                            type: string
                            description: >-
                              Optional identifier for the user sending the
                              message.
                        required:
                          - role
                          - content
                        title: User Message
                      - type: object
                        properties:
                          role:
                            type: string
                            enum:
                              - assistant
                            description: Response message from the AI assistant.
                          content:
                            anyOf:
                              - type: string
                              - type: array
                                items:
                                  type: object
                                  properties:
                                    type:
                                      type: string
                                      description: >-
                                        The type of content being transmitted.
                                        Supported values include "text" and
                                        "image".
                                    text:
                                      type: string
                                      description: >-
                                        The text content when type is "text".
                                        UTF-8 encoded string.
                                    image_url:
                                      type: object
                                      properties:
                                        url:
                                          type: string
                                          description: >-
                                            The URL pointing to the image resource.
                                            Must be a valid, accessible URL.
                                      required:
                                        - url
                                      description: >-
                                        Image URL configuration when type is
                                        "image".
                                  required:
                                    - type
                              - nullable: true
                              - nullable: true
                            description: >-
                              Assistant's response content. Can be text,
                              structured content, or null when using
                              tools/functions.
                          function_call:
                            nullable: true
                            description: >-
                              Function call details when the assistant invokes a
                              function.
                          tool_calls:
                            nullable: true
                            description: >-
                              Tool call details when the assistant uses
                              available tools.
                          name:
                            type: string
                            description: Optional identifier for the assistant instance.
                          citationMetadata:
                            type: object
                            properties:
                              citationSources:
                                type: array
                                items:
                                  type: object
                                  properties:
                                    startIndex:
                                      type: number
                                      description: >-
                                        Starting character index in the content
                                        where this citation applies.
                                    endIndex:
                                      type: number
                                      description: >-
                                        Ending character index in the content
                                        where this citation applies.
                                    uri:
                                      type: string
                                      description: >-
                                        URI reference to the source material
                                        being cited.
                                    license:
                                      type: string
                                      description: >-
                                        License information for the cited
                                        content.
                                description: >-
                                  Array of citation sources referenced in the
                                  content.
                            description: Metadata about citations used in the response.
                        required:
                          - role
                        title: Assistant Message
                      - type: object
                        properties:
                          role:
                            type: string
                            enum:
                              - function
                            description: Message containing function execution results.
                          content:
                            type: string
                            description: The result returned from the function execution.
                          name:
                            type: string
                            description: Name of the function that was executed.
                        required:
                          - role
                          - content
                          - name
                        title: Function Message
                      - type: object
                        properties:
                          role:
                            type: string
                            enum:
                              - tool
                            description: Message containing tool execution results.
                          content:
                            type: string
                            description: The result returned from the tool execution.
                          tool_call_id:
                            type: string
                            description: >-
                              Unique identifier for the tool call this message
                              responds to.
                        required:
                          - role
                          - content
                          - tool_call_id
                        title: Tool Message
                      - type: object
                        properties:
                          role:
                            type: string
                            enum:
                              - developer
                            description: Message from a developer or system administrator.
                          content:
                            type: string
                            description: >-
                              Developer message content or system configuration
                              instructions.
                          name:
                            type: string
                            description: Optional identifier for the developer.
                        required:
                          - role
                          - content
                        title: Developer Message
                  description: >-
                    Conversation history as an array of messages for contextual
                    generation.
                tools:
                  type: array
                  nullable: true
                  items:
                    type: object
                    properties:
                      type:
                        type: string
                        description: >-
                          The type of tool being specified. Currently supports
                          "function".
                      function:
                        type: object
                        properties:
                          name:
                            type: string
                            description: >-
                              The name of the function to be called. Must be
                              unique within the context.
                          description:
                            type: string
                            description: >-
                              Detailed description of the function's purpose and
                              behavior.
                          parameters:
                            type: object
                            additionalProperties:
                              nullable: true
                            description: >-
                              JSON Schema specification defining the expected
                              parameters for this function.
                        required:
                          - name
                        description: Function specification when type is "function".
                    required:
                      - type
                  description: >-
                    Array of tool definitions available for the model to use
                    during generation.
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
                    - nullable: true
                  description: >-
                    Controls tool usage behavior. Use "none" to disable, "auto"
                    for model choice, "required" to force tool use, or specify
                    tool for forced usage.
                temperature:
                  type: number
                  nullable: true
                  description: >-
                    Sampling temperature between 0 and 2. Higher values make
                    output more random, lower values more deterministic.
                top_p:
                  type: number
                  nullable: true
                  description: >-
                    Nucleus sampling parameter. Limits cumulative probability of
                    tokens considered for sampling.
                top_k:
                  type: number
                  nullable: true
                  description: Top-k sampling parameter
                'n':
                  type: number
                  nullable: true
                  description: Number of response alternatives to generate.
                stream:
                  type: boolean
                  nullable: true
                  description: >-
                    Enable streaming of partial response chunks as they are
                    generated.
                logprobs:
                  anyOf:
                    - type: boolean
                    - type: number
                    - nullable: true
                  description: >-
                    Include log probabilities of tokens: set to boolean for Chat
                    Completions; integer for legacy Completions (number of
                    top-token logprobs).
                stop:
                  anyOf:
                    - type: string
                    - type: array
                      items:
                        type: string
                    - nullable: true
                  description: >-
                    Sequence(s) at which to stop generation. Can be a single
                    string or array of strings.
                max_tokens:
                  type: number
                  nullable: true
                  description: Maximum number of tokens to generate in the response.
                presence_penalty:
                  type: number
                  nullable: true
                  description: >-
                    Penalty factor for new tokens based on their presence in
                    existing text. Range: -2.0 to 2.0
                frequency_penalty:
                  type: number
                  nullable: true
                  description: >-
                    Penalty factor for new tokens based on their frequency in
                    existing text. Range: -2.0 to 2.0
                logit_bias:
                  type: object
                  nullable: true
                  additionalProperties:
                    type: number
                  description: >-
                    Modify the likelihood of specified tokens appearing in the
                    completion.
                user:
                  type: string
                  nullable: true
                  description: >-
                    Unique identifier for the end-user making the request, for
                    monitoring and detecting abuse.
              required:
                - model
                - messages
      responses:
        '200':
          description: Chat completions generated successfully.
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: string
                    description: Id of the response.
                  object:
                    type: string
                    description: Type of the object, e.g., 'chat.completion'.
                  created:
                    type: number
                    description: Timestamp of when the response was created.
                  model:
                    type: string
                    description: The model used to generate the response.
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
                              description: Role of the message sender, e.g., 'assistant'.
                            content:
                              type: string
                              description: Content of the message.
                          required:
                            - role
                            - content
                        logprobs:
                          nullable: true
                          description: Log probabilities of the tokens.
                        finish_reason:
                          type: string
                          description: Reason why the generation stopped.
                      required:
                        - index
                        - message
                        - finish_reason
                    description: Array of choices returned by the model.
                  usage:
                    type: object
                    properties:
                      prompt_tokens:
                        type: number
                        description: Number of tokens in the prompt.
                      completion_tokens:
                        type: number
                        description: Number of tokens in the completion.
                      total_tokens:
                        type: number
                        description: Total number of tokens used.
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
                    description: Details about token usage.
                  service_tier:
                    type: string
                    description: Service tier used for the request.
                  system_fingerprint:
                    type: string
                    nullable: true
                    description: System fingerprint, if available.
                required:
                  - id
                  - object
                  - created
                  - model
                  - choices
                  - usage
                  - service_tier
                  - system_fingerprint
        '400':
          description: Bad request.
components:
  securitySchemes:
    AuthorizationBearer:
      type: http
      scheme: bearer
      bearerFormat: JWT

````