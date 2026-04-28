> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Generate Embeddings

> Generate embeddings for the given input using the specified model.



## OpenAPI

````yaml /gateway-openapi.json post /embeddings
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
  /embeddings:
    post:
      tags:
        - Embeddings
      summary: Generate Embeddings
      description: Generate embeddings for the given input using the specified model.
      parameters:
        - schema:
            type: string
            description: Optional metadata for the request
          required: false
          name: x-tfy-metadata
          in: header
      requestBody:
        required: true
        description: Parameters for generating embeddings.
        content:
          application/json:
            schema:
              type: object
              properties:
                model:
                  type: string
                  description: Model Name
                input:
                  anyOf:
                    - type: string
                    - type: array
                      items:
                        nullable: true
                  description: Input text to embed, encoded as a string or array of tokens
                user:
                  type: string
                  nullable: true
                  description: Unique identifier representing end-user
              required:
                - model
                - input
      responses:
        '200':
          description: Embeddings generated successfully.
        '400':
          description: Bad request.
components:
  securitySchemes:
    AuthorizationBearer:
      type: http
      scheme: bearer
      bearerFormat: JWT

````