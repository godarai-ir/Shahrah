> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Create Translation



## OpenAPI

````yaml post /audio/translations
openapi: 3.0.0
info:
  title: Portkey API
  description: >-
    The Portkey REST API. Please see https://portkey.ai/docs/api-reference for
    more details.
  version: 2.0.0
  termsOfService: https://portkey.ai/terms
  contact:
    name: Portkey Developer Forum
    url: https://portkey.wiki/community
  license:
    name: MIT
    url: https://github.com/Portkey-AI/portkey-openapi/blob/master/LICENSE
servers:
  - url: https://api.portkey.ai/v1
    description: Portkey API Public Endpoint
security:
  - Portkey-Key: []
tags:
  - name: Assistants
    description: Build Assistants that can call models and use tools.
  - name: Audio
    description: Turn audio into text or text into audio.
  - name: Chat
    description: >-
      Given a list of messages comprising a conversation, the model will return
      a response.
  - name: Realtime
    description: WebSocket proxy for provider Realtime APIs
  - name: Collections
    description: Create, List, Retrieve, Update, and Delete collections of prompts.
  - name: Labels
    description: Create, List, Retrieve, Update, and Delete labels.
  - name: Prompt Collections
    description: Create, List, Retrieve, Update, and Delete prompt collections.
  - name: PromptPartials
    description: Create, List, Retrieve, Update, and Delete prompt partials.
  - name: Prompts
    description: >-
      Given a prompt template ID and variables, will run the saved prompt
      template and return a response.
  - name: Guardrails
    description: Create, List, Retrieve, Update, and Delete prompt Guardrails.
  - name: Completions
    description: >-
      Given a prompt, the model will return one or more predicted completions,
      and can also return the probabilities of alternative tokens at each
      position.
  - name: Embeddings
    description: >-
      Get a vector representation of a given input that can be easily consumed
      by machine learning models and algorithms.
  - name: Rerank
    description: >-
      Rerank a list of documents based on their relevance to a query. Supported
      providers include Cohere, Voyage, Jina, Pinecone, Bedrock, and Azure AI.
  - name: Fine-tuning
    description: Manage fine-tuning jobs to tailor a model to your specific training data.
  - name: Batch
    description: Create large batches of API requests to run asynchronously.
  - name: Files
    description: >-
      Files are used to upload documents that can be used with features like
      Assistants and Fine-tuning.
  - name: Images
    description: Given a prompt and/or an input image, the model will generate a new image.
  - name: Models
    description: List and describe the various models available in the API.
  - name: Moderations
    description: >-
      Given a input text, outputs if the model classifies it as potentially
      harmful.
  - name: Configs
    description: Create, List, Retrieve, and Update your Portkey Configs.
  - name: Feedback
    description: Send and Update any feedback.
  - name: Logs
    description: Custom Logger to add external logs to Portkey.
  - name: Integrations
    description: Create, List, Retrieve, Update, and Delete your Portkey Integrations.
  - name: Integrations > Workspaces
    description: Manage workspace access for your Portkey Integrations.
  - name: Integrations > Models
    description: Manage model access for your Portkey Integrations.
  - name: Providers
    description: Create, List, Retrieve, Update, and Delete your Portkey Providers.
  - name: Virtual-keys
    description: Create, List, Retrieve, Update, and Delete your Portkey Virtual keys.
  - name: Users
    description: Create and manage users.
  - name: User-invites
    description: Create and manage user invites.
  - name: Workspaces
    description: Create and manage workspaces.
  - name: Workspaces > Members
    description: Create and manage workspace members.
  - name: MCP Integrations
    description: Create, List, Retrieve, Update, and Delete MCP Integrations.
  - name: MCP Integrations > Workspaces
    description: Manage workspace access for MCP Integrations.
  - name: MCP Integrations > Capabilities
    description: List and manage capabilities for MCP Integrations.
  - name: MCP Integrations > Metadata
    description: Get MCP Integration metadata and sync info.
  - name: MCP Servers
    description: >-
      Create, List, Retrieve, Update, and Delete MCP Servers (workspace
      instances of MCP Integrations).
  - name: MCP Servers > Capabilities
    description: List and manage capabilities for MCP Servers.
  - name: MCP Servers > User Access
    description: List and manage user access for MCP Servers.
  - name: MCP Servers > Connections
    description: List and manage user connections for MCP Servers.
  - name: Api-Keys
    description: Create, List, Retrieve, Update, and Delete your Portkey API keys.
  - name: Logs Export
    description: Exports logs service.
  - name: Audit Logs
    description: Get audit logs for your Portkey account.
  - name: Analytics
    description: >-
      Get analytics over different data points like requests, costs, tokens,
      etc.
  - name: Analytics > Graphs
    description: Get data points for graphical representation.
  - name: Analytics > Summary
    description: Get overall summary for the selected time bucket.
  - name: Analytics > Groups
    description: Get grouped metrics for the selected time bucket.
  - name: Usage Limits Policies
    description: Manage usage limits policies to control total usage over time
  - name: Rate Limits Policies
    description: Manage rate limits policies to control request or token rates
  - name: Model Pricing
    description: Model pricing configurations for 2300+ LLMs across 40+ providers
  - name: Secret-References
    description: >-
      Create, List, Retrieve, Update, and Delete secret references to external
      secret managers.
paths:
  /audio/translations:
    servers:
      - url: https://api.portkey.ai/v1
        description: Portkey API Public Endpoint
      - url: SELF_HOSTED_GATEWAY_URL
        description: Self-Hosted Gateway URL
    post:
      tags:
        - Audio
      summary: Create Translation
      operationId: createTranslation
      requestBody:
        required: true
        content:
          multipart/form-data:
            schema:
              $ref: '#/components/schemas/CreateTranslationRequest'
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                oneOf:
                  - $ref: '#/components/schemas/CreateTranslationResponseJson'
                  - $ref: '#/components/schemas/CreateTranslationResponseVerboseJson'
      security:
        - Portkey-Key: []
          Virtual-Key: []
        - Portkey-Key: []
          Provider-Auth: []
          Provider-Name: []
        - Portkey-Key: []
          Config: []
        - Portkey-Key: []
          Provider-Auth: []
          Provider-Name: []
          Custom-Host: []
      x-code-samples:
        - lang: curl
          label: Default
          source: |
            curl https://api.portkey.ai/v1/audio/translations \
              -H "x-portkey-api-key: $PORTKEY_API_KEY" \
              -H "x-portkey-virtual-key: $PORTKEY_PROVIDER_VIRTUAL_KEY" \
              -H "Content-Type: multipart/form-data" \
              -F file="@/path/to/file/german.m4a" \
              -F model="whisper-1"
        - lang: python
          label: Default
          source: |
            from portkey_ai import Portkey

            client = Portkey(
              api_key = "PORTKEY_API_KEY",
              virtual_key = "PROVIDER_VIRTUAL_KEY"
            )

            audio_file = open("speech.mp3", "rb")
            transcript = client.audio.translations.create(
              model="whisper-1",
              file=audio_file
            )
        - lang: javascript
          label: Default
          source: |
            import fs from "fs";
            import Portkey from 'portkey-ai';

            const client = new Portkey({
              apiKey: 'PORTKEY_API_KEY',
              virtualKey: 'PROVIDER_VIRTUAL_KEY'
            });

            async function main() {
                const translation = await client.audio.translations.create({
                    file: fs.createReadStream("speech.mp3"),
                    model: "whisper-1",
                });

                console.log(translation.text);
            }
            main();
        - lang: curl
          label: Self-Hosted
          source: |
            curl -X POST "SELF_HOSTED_GATEWAY_URL/audio/translations" \
              -H "Content-Type: application/json" \
              -H "x-portkey-api-key: $PORTKEY_API_KEY" \
              -H "x-portkey-virtual-key: $PORTKEY_PROVIDER_VIRTUAL_KEY" \
              -d '{
                "model": "whisper-1",
                "file": "@/path/to/file/german.m4a"
              }' \
              --output translation.json
        - lang: python
          label: Self-Hosted
          source: |
            from portkey_ai import Portkey

            client = Portkey(
              api_key = "PORTKEY_API_KEY",
              virtual_key = "PROVIDER_VIRTUAL_KEY",
              base_url="SELF_HOSTED_GATEWAY_URL"
            )

            audio_file = open("speech.mp3", "rb")
            transcript = client.audio.translations.create(
              model="whisper-1",
              file=audio_file
            )
        - lang: javascript
          label: Self-Hosted
          source: |
            import fs from "fs";
            import Portkey from 'portkey-ai';

            const client = new Portkey({
              apiKey: 'PORTKEY_API_KEY',
              virtualKey: 'PROVIDER_VIRTUAL_KEY',
              baseUrl: 'SELF_HOSTED_GATEWAY_URL'
            });

            const audioFile = fs.createReadStream("speech.mp3");

            async function main() {
              const translation = await client.audio.translations.create({
                file: audioFile,
                model: "whisper-1",
              });

              console.log(translation.text);
            }
            main();
components:
  schemas:
    CreateTranslationRequest:
      type: object
      additionalProperties: false
      properties:
        file:
          description: >
            The audio file object (not file name) translate, in one of these
            formats: flac, mp3, mp4, mpeg, mpga, m4a, ogg, wav, or webm.
          type: string
          x-oaiTypeLabel: file
          format: binary
        model:
          description: >
            ID of the model to use. The options are `gpt-4o-transcribe`,
            `gpt-4o-mini-transcribe`, and `whisper-1`.
          example: whisper-1
          anyOf:
            - type: string
            - type: string
              enum:
                - whisper-1
          x-oaiTypeLabel: string
        prompt:
          description: >
            An optional text to guide the model's style or continue a previous
            audio segment. The
            [prompt](https://platform.openai.com/docs/guides/speech-to-text/prompting)
            should be in English.
          type: string
        response_format:
          description: >
            The format of the transcript output, in one of these options:
            `json`, `text`, `srt`, `verbose_json`, or `vtt`.
          type: string
          default: json
        temperature:
          description: >
            The sampling temperature, between 0 and 1. Higher values like 0.8
            will make the output more random, while lower values like 0.2 will
            make it more focused and deterministic. If set to 0, the model will
            use [log probability](https://en.wikipedia.org/wiki/Log_probability)
            to automatically increase the temperature until certain thresholds
            are hit.
          type: number
          default: 0
      required:
        - file
        - model
    CreateTranslationResponseJson:
      type: object
      properties:
        text:
          type: string
      required:
        - text
    CreateTranslationResponseVerboseJson:
      type: object
      properties:
        language:
          type: string
          description: The language of the output translation (always `english`).
        duration:
          type: string
          description: The duration of the input audio.
        text:
          type: string
          description: The translated text.
        segments:
          type: array
          description: Segments of the translated text and their corresponding details.
          items:
            $ref: '#/components/schemas/TranscriptionSegment'
      required:
        - language
        - duration
        - text
    TranscriptionSegment:
      type: object
      properties:
        id:
          type: integer
          description: Unique identifier of the segment.
        seek:
          type: integer
          description: Seek offset of the segment.
        start:
          type: number
          format: float
          description: Start time of the segment in seconds.
        end:
          type: number
          format: float
          description: End time of the segment in seconds.
        text:
          type: string
          description: Text content of the segment.
        tokens:
          type: array
          items:
            type: integer
          description: Array of token IDs for the text content.
        temperature:
          type: number
          format: float
          description: Temperature parameter used for generating the segment.
        avg_logprob:
          type: number
          format: float
          description: >-
            Average logprob of the segment. If the value is lower than -1,
            consider the logprobs failed.
        compression_ratio:
          type: number
          format: float
          description: >-
            Compression ratio of the segment. If the value is greater than 2.4,
            consider the compression failed.
        no_speech_prob:
          type: number
          format: float
          description: >-
            Probability of no speech in the segment. If the value is higher than
            1.0 and the `avg_logprob` is below -1, consider this segment silent.
      required:
        - id
        - seek
        - start
        - end
        - text
        - tokens
        - temperature
        - avg_logprob
        - compression_ratio
        - no_speech_prob
  securitySchemes:
    Portkey-Key:
      type: apiKey
      in: header
      name: x-portkey-api-key
    Virtual-Key:
      type: apiKey
      in: header
      name: x-portkey-virtual-key
    Provider-Auth:
      type: http
      scheme: bearer
    Provider-Name:
      type: apiKey
      in: header
      name: x-portkey-provider
    Config:
      type: apiKey
      in: header
      name: x-portkey-config
    Custom-Host:
      type: apiKey
      in: header
      name: x-portkey-custom-host

````