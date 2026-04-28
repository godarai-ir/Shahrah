> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# List available models

> Lists the models that the requesting user is authorized to access.



## OpenAPI

````yaml /gateway-openapi.json get /models
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
  /models:
    get:
      tags:
        - Models
      summary: List available models
      description: Lists the models that the requesting user is authorized to access.
      responses:
        '200':
          description: A list of models the user is authorized to access
          content:
            application/json:
              schema:
                type: object
                properties:
                  object:
                    type: string
                    enum:
                      - list
                    description: The object type, list in this case
                  data:
                    type: array
                    items:
                      type: object
                      properties:
                        id:
                          type: string
                          description: TrueFoundry Model Name
                        object:
                          type: string
                          enum:
                            - model
                          description: The object type, model in this case
                        created:
                          type: number
                          description: >-
                            Unix timestamp (in seconds) when the model was added
                            to the system
                        owned_by:
                          type: string
                          description: >-
                            The organization that owns the model (provider name
                            or "virtual")
                      required:
                        - id
                        - object
                        - created
                        - owned_by
                    description: Array of model objects
                required:
                  - object
                  - data
        '401':
          description: Unauthorized
components:
  securitySchemes:
    AuthorizationBearer:
      type: http
      scheme: bearer
      bearerFormat: JWT

````