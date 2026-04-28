> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# List Model Response Input Items

> List input items for a specific model response



## OpenAPI

````yaml /gateway-openapi.json get /responses/{id}/input_items
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
  /responses/{id}/input_items:
    get:
      tags:
        - Responses
      summary: List Model Response Input Items
      description: List input items for a specific model response
      parameters:
        - schema:
            type: string
            description: Id of the response
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
          description: Model response input items retrieved successfully
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
                        id:
                          type: string
                          description: Id of the message.
                        type:
                          type: string
                          description: Type of the message, e.g., 'message'.
                        status:
                          type: string
                          description: Status of the message, e.g., 'completed'.
                        content:
                          type: array
                          items:
                            type: object
                            properties:
                              type:
                                type: string
                                description: Type of the content, e.g., 'input_text'.
                              text:
                                type: string
                                description: Text content of the message.
                            required:
                              - type
                              - text
                          description: Array of content objects.
                        role:
                          type: string
                          description: Role of the entity, e.g., 'user' or 'system'.
                      required:
                        - id
                        - type
                        - status
                        - content
                        - role
                    description: Array of message objects.
                  first_id:
                    type: string
                    description: ID of the first message in the list.
                  has_more:
                    type: boolean
                    description: Indicates if there are more messages.
                  last_id:
                    type: string
                    description: ID of the last message in the list.
                required:
                  - object
                  - data
                  - first_id
                  - has_more
                  - last_id
        '404':
          description: Model response not found
components:
  securitySchemes:
    AuthorizationBearer:
      type: http
      scheme: bearer
      bearerFormat: JWT

````