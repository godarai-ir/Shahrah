> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Creates a variation of an image.

> Generates a variation of a given image.



## OpenAPI

````yaml /gateway-openapi.json post /images/variations
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
  /images/variations:
    post:
      tags:
        - Image
      summary: Creates a variation of an image.
      description: Generates a variation of a given image.
      requestBody:
        content:
          multipart/form-data:
            schema:
              type: object
              properties:
                model:
                  type: string
                  description: The model to use for image generation.
                image:
                  type: string
                  description: The image to vary.
                'n':
                  type: integer
                  nullable: true
                  description: >-
                    The number of variations to generate. Must be between 1 and
                    10.
                response_format:
                  type: string
                  nullable: true
                  description: >-
                    The format in which the generated images are returned. Must
                    be one of url or b64_json.
                size:
                  type: string
                  nullable: true
                  description: The size of the generated images.
                user:
                  type: string
                  description: A unique identifier representing your end-user.
              required:
                - image
      responses:
        '200':
          description: Successfully varied images
          content:
            application/json:
              schema:
                type: object
                properties:
                  created:
                    type: number
                  data:
                    type: array
                    items:
                      type: object
                      properties:
                        url:
                          type: string
                required:
                  - created
                  - data
        '400':
          description: Bad Request
        '401':
          description: Unauthorized
        '404':
          description: Not Found
        '500':
          description: Internal Server Error
components:
  securitySchemes:
    AuthorizationBearer:
      type: http
      scheme: bearer
      bearerFormat: JWT

````