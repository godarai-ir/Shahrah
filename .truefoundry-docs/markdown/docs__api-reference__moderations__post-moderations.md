> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Post moderations

> API reference for post moderations, covering request parameters, responses, and examples for integrating with TrueFoundry.



## OpenAPI

````yaml /gateway-openapi.json post /moderations
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
  /moderations:
    post:
      tags:
        - Moderations
      description: >-
        API reference for post moderations, covering request parameters,
        responses, and examples for integrating with TrueFoundry.
      parameters:
        - schema:
            type: string
            description: Optional metadata for the request
          required: false
          name: x-tfy-metadata
          in: header
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                input:
                  anyOf:
                    - type: string
                    - type: array
                      items:
                        type: string
                    - type: array
                      items:
                        nullable: true
                  description: >-
                    Input (or inputs) to classify. Can be a string, array of
                    strings, or array of multi-modal inputs
                model:
                  type: string
                  description: Model to use for moderation
              required:
                - input
      responses:
        '200':
          description: Successful moderation response
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: string
                    description: Unique moderation ID
                  model:
                    type: string
                    description: Model used for moderation
                  results:
                    type: array
                    items:
                      type: object
                      properties:
                        flagged:
                          type: boolean
                          description: True if content violates policies
                        categories:
                          type: object
                          properties:
                            hate:
                              type: boolean
                              description: Contains hate speech
                            hate/threatening:
                              type: boolean
                              description: Contains threatening hate speech
                            harassment:
                              type: boolean
                              description: Contains harassment
                            harassment/threatening:
                              type: boolean
                              description: Contains threatening harassment
                            self-harm:
                              type: boolean
                              description: Contains self-harm content
                            self-harm/intent:
                              type: boolean
                              description: Expresses intent of self-harm
                            self-harm/instructions:
                              type: boolean
                              description: Contains self-harm instructions
                            sexual:
                              type: boolean
                              description: Contains sexual content
                            sexual/minors:
                              type: boolean
                              description: Contains sexual content with minors
                            violence:
                              type: boolean
                              description: Contains violent content
                            violence/graphic:
                              type: boolean
                              description: Contains graphic violence
                          required:
                            - hate
                            - hate/threatening
                            - harassment
                            - harassment/threatening
                            - self-harm
                            - self-harm/intent
                            - self-harm/instructions
                            - sexual
                            - sexual/minors
                            - violence
                            - violence/graphic
                          description: Binary flags for content violations
                        category_scores:
                          type: object
                          properties:
                            hate:
                              type: number
                              description: Hate score
                            hate/threatening:
                              type: number
                              description: Threatening hate score
                            harassment:
                              type: number
                              description: Harassment score
                            harassment/threatening:
                              type: number
                              description: Threatening harassment score
                            self-harm:
                              type: number
                              description: Self-harm score
                            self-harm/intent:
                              type: number
                              description: Self-harm intent score
                            self-harm/instructions:
                              type: number
                              description: Self-harm instructions score
                            sexual:
                              type: number
                              description: Sexual content score
                            sexual/minors:
                              type: number
                              description: Sexual content with minors score
                            violence:
                              type: number
                              description: Violence score
                            violence/graphic:
                              type: number
                              description: Graphic violence score
                          required:
                            - hate
                            - hate/threatening
                            - harassment
                            - harassment/threatening
                            - self-harm
                            - self-harm/intent
                            - self-harm/instructions
                            - sexual
                            - sexual/minors
                            - violence
                            - violence/graphic
                          description: Confidence scores for each category
                      required:
                        - flagged
                        - categories
                        - category_scores
                    description: Moderation results per input
                required:
                  - id
                  - model
                  - results
        '400':
          description: Bad Request
components:
  securitySchemes:
    AuthorizationBearer:
      type: http
      scheme: bearer
      bearerFormat: JWT

````