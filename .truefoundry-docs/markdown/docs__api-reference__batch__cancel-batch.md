> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Cancel Batch

> Cancel a running batch process



## OpenAPI

````yaml /gateway-openapi.json post /batches/{id}/cancel
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
  /batches/{id}/cancel:
    post:
      tags:
        - Batch
      summary: Cancel Batch
      description: Cancel a running batch process
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
          description: Batch cancelled successfully
        '404':
          description: Batch not found
components:
  securitySchemes:
    AuthorizationBearer:
      type: http
      scheme: bearer
      bearerFormat: JWT

````