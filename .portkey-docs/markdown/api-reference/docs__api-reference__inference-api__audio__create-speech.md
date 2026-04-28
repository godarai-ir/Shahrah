> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Create Speech



## OpenAPI

````yaml post /audio/speech
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
  /audio/speech:
    servers:
      - url: https://api.portkey.ai/v1
        description: Portkey API Public Endpoint
      - url: SELF_HOSTED_GATEWAY_URL
        description: Self-Hosted Gateway URL
    post:
      tags:
        - Audio
      summary: Create Speech
      operationId: createSpeech
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateSpeechRequest'
      responses:
        '200':
          description: OK
          headers:
            Transfer-Encoding:
              schema:
                type: string
              description: chunked
          content:
            application/octet-stream:
              schema:
                type: string
                format: binary
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
            curl https://api.portkey.ai/v1/audio/speech \
              -H "x-portkey-api-key: $PORTKEY_API_KEY" \
              -H "x-portkey-virtual-key: $PORTKEY_PROVIDER_VIRTUAL_KEY" \
              -H "Content-Type: application/json" \
              -d '{
                "model": "tts-1",
                "input": "The quick brown fox jumped over the lazy dog.",
                "voice": "alloy"
              }' \
              --output speech.mp3
        - lang: python
          label: Default
          source: |
            from pathlib import Path
            from portkey_ai import Portkey

            client = Portkey(
              api_key = "PORTKEY_API_KEY",
              virtual_key = "PROVIDER_VIRTUAL_KEY"
            )

            speech_file_path = Path(__file__).parent / "speech.mp3"
            response = client.audio.speech.create(
              model="tts-1",
              voice="alloy",
              input="The quick brown fox jumped over the lazy dog."
            )
            response.stream_to_file(speech_file_path)
        - lang: javascript
          label: Default
          source: |
            import fs from "fs";
            import path from "path";
            import Portkey from 'portkey-ai';

            const client = new Portkey({
              apiKey: 'PORTKEY_API_KEY',
              virtualKey: 'PROVIDER_VIRTUAL_KEY'
            });

            const speechFile = path.resolve("./speech.mp3");

            async function main() {
              const mp3 = await client.audio.speech.create({
                model: "tts-1",
                voice: "alloy",
                input: "Today is a wonderful day to build something people love!",
              });
              console.log(speechFile);
              const buffer = Buffer.from(await mp3.arrayBuffer());
              await fs.promises.writeFile(speechFile, buffer);
            }
            main();
        - lang: curl
          label: Self-Hosted
          source: |
            curl -X POST "SELF_HOSTED_GATEWAY_URL/audio/speech" \
              -H "Content-Type: application/json" \
              -H "x-portkey-api-key: $PORTKEY_API_KEY" \
              -H "x-portkey-virtual-key: $PORTKEY_PROVIDER_VIRTUAL_KEY" \
              -d '{
                "model": "tts-1",
                "input": "The quick brown fox jumped over the lazy dog.",
                "voice": "alloy"
              }' \
              --output speech.mp3
        - lang: python
          label: Self-Hosted
          source: |
            from pathlib import Path
            from portkey_ai import Portkey

            client = Portkey(
              api_key = "PORTKEY_API_KEY",
              virtual_key = "PROVIDER_VIRTUAL_KEY",
              base_url="SELF_HOSTED_GATEWAY_URL"
            )

            speech_file_path = Path(__file__).parent / "speech.mp3"
            response = client.audio.speech.create(
              model="tts-1",
              voice="alloy",
              input="The quick brown fox jumped over the lazy dog."
            )
            response.stream_to_file(speech_file_path)
        - lang: javascript
          label: Self-Hosted
          source: |
            import fs from "fs";
            import path from "path";
            import Portkey from 'portkey-ai';

            const client = new Portkey({
              apiKey: 'PORTKEY_API_KEY',
              virtualKey: 'PROVIDER_VIRTUAL_KEY',
              baseUrl: 'SELF_HOSTED_GATEWAY_URL'
            });

            const speechFile = path.resolve("./speech.mp3");

            async function main() {
              const mp3 = await client.audio.speech.create({
                model: "tts-1",
                voice: "alloy",
                input: "Today is a wonderful day to build something people love!",
              });
              console.log(speechFile);
              const buffer = Buffer.from(await mp3.arrayBuffer());
              await fs.promises.writeFile(speechFile, buffer);
            }
            main();
components:
  schemas:
    CreateSpeechRequest:
      type: object
      additionalProperties: false
      properties:
        model:
          description: >
            One of the available [TTS
            models](https://platform.openai.com/docs/models/tts): `tts-1` or
            `tts-1-hd`
          anyOf:
            - type: string
            - type: string
              enum:
                - tts-1
                - tts-1-hd
          x-oaiTypeLabel: string
        input:
          type: string
          description: >-
            The text to generate audio for. The maximum length is 4096
            characters.
          maxLength: 4096
        voice:
          description: >-
            The voice to use when generating the audio. Supported voices are
            `alloy`, `echo`, `fable`, `onyx`, `nova`, and `shimmer`. Previews of
            the voices are available in the [Text to speech
            guide](https://platform.openai.com/docs/guides/text-to-speech/voice-options).
          type: string
          enum:
            - alloy
            - echo
            - fable
            - onyx
            - nova
            - shimmer
        response_format:
          description: >-
            The format to audio in. Supported formats are `mp3`, `opus`, `aac`,
            `flac`, `wav`, and `pcm`.
          default: mp3
          type: string
          enum:
            - mp3
            - opus
            - aac
            - flac
            - wav
            - pcm
        speed:
          description: >-
            The speed of the generated audio. Select a value from `0.25` to
            `4.0`. `1.0` is the default.
          type: number
          default: 1
          minimum: 0.25
          maximum: 4
      required:
        - model
        - input
        - voice
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