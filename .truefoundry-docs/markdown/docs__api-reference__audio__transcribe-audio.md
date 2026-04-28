> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Transcribe Audio

> Transcribes audio into the input language.



## OpenAPI

````yaml /gateway-openapi.json post /audio/transcriptions
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
  /audio/transcriptions:
    post:
      tags:
        - Audio
      summary: Transcribe Audio
      description: Transcribes audio into the input language.
      parameters:
        - schema:
            type: string
            description: Optional metadata for the request
          required: false
          name: x-tfy-metadata
          in: header
      requestBody:
        required: true
        description: Audio file to transcribe and transcription options
        content:
          multipart/form-data:
            schema:
              type: object
              properties:
                file:
                  nullable: true
                  description: >-
                    Audio file to transcribe: flac, mp3, mp4, mpeg, mpga, m4a,
                    ogg, wav, or webm.
                model:
                  type: string
                  description: Identifier of the model to be used for generation.
                chunking_strategy:
                  anyOf:
                    - type: string
                      enum:
                        - auto
                    - type: object
                      properties: {}
                  description: >-
                    Controls audio chunking. "auto" uses VAD to select
                    boundaries.
                include:
                  type: array
                  items:
                    type: string
                  description: >-
                    Additional info to include (e.g. logprobs for model
                    confidence).
                language:
                  type: string
                  description: Input audio language in ISO-639-1 format (e.g. en).
                prompt:
                  type: string
                  description: Text to guide model style or continue a previous segment.
                response_format:
                  type: string
                  description: >-
                    Output format: json, text, srt, verbose_json, or vtt.
                    Default: json.
                temperature:
                  type: number
                  description: >-
                    Sampling temperature (0-1). Higher values increase
                    randomness. Default: 0.
                timestamp_granularities:
                  type: array
                  items:
                    type: string
                    enum:
                      - word
                      - segment
                  description: >-
                    Timestamp detail level: word or segment. Default:
                    ["segment"].
              required:
                - model
      responses:
        '200':
          description: Audio transcribed successfully
          content:
            application/json:
              schema:
                type: object
                properties:
                  text:
                    type: string
                    description: The transcribed text.
                required:
                  - text
        '400':
          description: Bad Request
components:
  securitySchemes:
    AuthorizationBearer:
      type: http
      scheme: bearer
      bearerFormat: JWT

````