> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Edit images based on a prompt

> Edits an image given the original image and a prompt.



## OpenAPI

````yaml /gateway-openapi.json post /images/edits
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
  /images/edits:
    post:
      tags:
        - Image
      summary: Edit images based on a prompt
      description: Edits an image given the original image and a prompt.
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
                  anyOf:
                    - type: string
                    - type: array
                      items:
                        type: string
                  description: >-
                    The image(s) to edit. Must be a supported image file or an
                    array of images.
                prompt:
                  type: string
                  description: A text description of the desired image(s).
                background:
                  type: string
                  nullable: true
                  description: >-
                    Allows to set transparency for the background of the
                    generated image(s).
                input_fidelity:
                  type: string
                  nullable: true
                  description: >-
                    Control how much effort the model will exert to match the
                    style and features of input images.
                mask:
                  type: string
                  description: >-
                    An additional image whose fully transparent areas indicate
                    where image should be edited.
                'n':
                  type: integer
                  nullable: true
                  description: The number of images to generate. Must be between 1 and 10.
                output_compression:
                  type: integer
                  nullable: true
                  description: The compression level (0-100%) for the generated images.
                output_format:
                  type: string
                  nullable: true
                  description: The format in which the generated images are returned.
                partial_images:
                  type: integer
                  nullable: true
                  description: >-
                    The number of partial images to generate. Used for streaming
                    responses.
                quality:
                  type: string
                  nullable: true
                  description: The quality of the image that will be generated.
                response_format:
                  type: string
                  nullable: true
                  description: >-
                    The format in which generated images are returned. Must be
                    one of url or b64_json.
                size:
                  type: string
                  nullable: true
                  description: The size of the generated images.
                stream:
                  type: boolean
                  nullable: true
                  description: Edit the image in streaming mode.
                user:
                  type: string
                  description: A unique identifier representing your end-user.
              required:
                - model
                - image
                - prompt
      responses:
        '200':
          description: Successfully edited images
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
                        b64_json:
                          type: string
                        revised_prompt:
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