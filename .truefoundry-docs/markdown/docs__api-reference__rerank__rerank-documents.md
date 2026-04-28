> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Rerank Documents

> Rerank documents based on the given query and parameters.



## OpenAPI

````yaml /gateway-openapi.json post /v2/rerank
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
  /v2/rerank:
    post:
      tags:
        - Rerank
      summary: Rerank Documents
      description: Rerank documents based on the given query and parameters.
      parameters:
        - schema:
            type: string
            description: Optional metadata for the request
          required: false
          name: x-tfy-metadata
          in: header
      requestBody:
        required: true
        description: Parameters for reranking documents.
        content:
          application/json:
            schema:
              type: object
              properties:
                model:
                  type: string
                  description: Model to generate the response
                query:
                  type: string
                  nullable: true
                  description: The query string for reranking.
                documents:
                  type: array
                  nullable: true
                  items:
                    type: string
                  description: An array of documents to rerank.
                top_n:
                  type: number
                  nullable: true
                  description: The number of top documents to return.
                max_tokens_per_doc:
                  type: number
                  nullable: true
                  description: The maximum number of tokens per document.
              required:
                - model
      responses:
        '200':
          description: Documents reranked successfully.
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: string
                    description: Id of the rerank response.
                  results:
                    type: array
                    items:
                      type: object
                      properties:
                        index:
                          type: number
                          description: Index of the result.
                        relevance_score:
                          type: number
                          description: Relevance score of the result.
                      required:
                        - index
                        - relevance_score
                    description: Array of results returned by the model.
                required:
                  - id
                  - results
        '400':
          description: Bad request.
components:
  securitySchemes:
    AuthorizationBearer:
      type: http
      scheme: bearer
      bearerFormat: JWT

````