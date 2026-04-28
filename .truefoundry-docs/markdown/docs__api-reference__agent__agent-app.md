> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Agent App

> Generate completions using agent app configuration. The model, MCP servers, system prompt, and guardrails are automatically provided from the agent app configuration.



## OpenAPI

````yaml /gateway-openapi.json post /agent/{agent-app-id}/responses
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
  /agent/{agent-app-id}/responses:
    post:
      tags:
        - Agent
      summary: Agent App
      description: >-
        Generate completions using agent app configuration. The model, MCP
        servers, system prompt, and guardrails are automatically provided from
        the agent app configuration.
      parameters:
        - schema:
            type: string
            description: The ID of the agent app to use for configuration
          required: true
          name: agent-app-id
          in: path
      requestBody:
        required: true
        description: >-
          Chat parameters. The model, MCP servers, system prompt, guardrails,
          and model parameters (temperature, max_tokens, etc.) are autofilled
          from agent app configuration.
        content:
          application/json:
            schema:
              type: object
              properties:
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
                  description: A list of messages comprising the conversation history
                stream:
                  type: boolean
                  nullable: true
                  description: Whether to stream responses
                iteration_limit:
                  type: integer
                  minimum: 0
                  exclusiveMinimum: true
                  maximum: 20
                  default: 5
                  description: >-
                    Maximum number of iterations to perform with tool calls.
                    Must be a positive integer. Defaults to 5.
              required:
                - iteration_limit
      responses:
        '200':
          description: >-
            Chat completions response with tool calls using agent app
            configuration
          content:
            text/event-stream:
              schema:
                type: object
                properties:
                  id:
                    type: string
                  object:
                    type: string
                    enum:
                      - chat.completion.chunk
                  created:
                    type: number
                  model:
                    type: string
                  service_tier:
                    type: string
                  system_fingerprint:
                    type: string
                  choices:
                    type: array
                    items:
                      type: object
                      properties:
                        index:
                          type: number
                        delta:
                          type: object
                          properties:
                            content:
                              type: string
                            role:
                              type: string
                            refusal:
                              nullable: true
                          additionalProperties:
                            nullable: true
                        logprobs:
                          nullable: true
                        finish_reason:
                          type: string
                          nullable: true
                      required:
                        - index
                        - delta
                        - finish_reason
                required:
                  - object
                  - created
                  - model
                  - choices
        '400':
          description: Bad request.
        '404':
          description: Agent app not found or access denied.
components:
  securitySchemes:
    AuthorizationBearer:
      type: http
      scheme: bearer
      bearerFormat: JWT

````