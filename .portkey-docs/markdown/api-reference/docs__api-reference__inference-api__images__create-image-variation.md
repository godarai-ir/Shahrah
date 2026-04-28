> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Create Image Variation



## OpenAPI

````yaml post /images/variations
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
  /images/variations:
    servers:
      - url: https://api.portkey.ai/v1
        description: Portkey API Public Endpoint
      - url: SELF_HOSTED_GATEWAY_URL
        description: Self-Hosted Gateway URL
    post:
      tags:
        - Images
      summary: Creates Image Variation
      operationId: createImageVariation
      requestBody:
        required: true
        content:
          multipart/form-data:
            schema:
              $ref: '#/components/schemas/CreateImageVariationRequest'
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ImagesResponse'
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
            curl https://api.portkey.ai/v1/images/variations \
              -H "x-portkey-api-key: $PORTKEY_API_KEY" \
              -H "x-portkey-virtual-key: $PORTKEY_PROVIDER_VIRTUAL_KEY" \
              -F image="@otter.png" \
              -F n=2 \
              -F size="1024x1024"
        - lang: python
          label: Default
          source: |
            from portkey_ai import Portkey

            client = Portkey(
              api_key = "PORTKEY_API_KEY",
              virtual_key = "PROVIDER_VIRTUAL_KEY"
            )

            response = client.images.create_variation(
              image=open("image_edit_original.png", "rb"),
              n=2,
              size="1024x1024"
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
              const image = await client.images.createVariation({
                image: fs.createReadStream("otter.png"),
              });

              console.log(image.data);
            }
            main();
        - lang: curl
          label: Self-Hosted
          source: |
            curl -X POST "SELF_HOSTED_GATEWAY_URL/images/variations" \
              -H "Content-Type: application/json" \
              -H "x-portkey-api-key: $PORTKEY_API_KEY" \
              -H "x-portkey-virtual-key: $PORTKEY_PROVIDER_VIRTUAL_KEY" \
              -d '{
                "image": "@otter.png",
                "n": 2,
                "size": "1024x1024"
              }'
        - lang: python
          label: Self-Hosted
          source: |
            from portkey_ai import Portkey

            client = Portkey(
                api_key="PORTKEY_API_KEY",
                virtual_key="PROVIDER_VIRTUAL_KEY",
                base_url="SELF_HOSTED_GATEWAY_URL"
            )

            image = client.images.create_variation(
                image=open("otter.png", "rb"),
                n=2,
                size="1024x1024"
            )

            print(image.data)
        - lang: javascript
          label: Self-Hosted
          source: |
            import fs from "fs";
            import Portkey from 'portkey-ai';

            const portkey = new Portkey({
              apiKey: 'PORTKEY_API_KEY',
              virtualKey: 'PROVIDER_VIRTUAL_KEY',
              baseURL: 'SELF_HOSTED_GATEWAY_URL'
            });

            async function main() {
              const image = await portkey.images.createVariation({
                image: fs.createReadStream("otter.png"),
              });

              console.log(image.data);
            }
            main();
components:
  schemas:
    CreateImageVariationRequest:
      type: object
      properties:
        image:
          description: >-
            The image to use as the basis for the variation(s). Must be a valid
            PNG file, less than 4MB, and square.
          type: string
          format: binary
        model:
          anyOf:
            - type: string
            - type: string
              enum:
                - dall-e-2
          x-oaiTypeLabel: string
          default: dall-e-2
          example: dall-e-2
          nullable: true
          description: >-
            The model to use for image generation. Only `dall-e-2` is supported
            at this time.
        'n':
          type: integer
          minimum: 1
          maximum: 10
          default: 1
          example: 1
          nullable: true
          description: >-
            The number of images to generate. Must be between 1 and 10. For
            `dall-e-3`, only `n=1` is supported.
        response_format:
          type: string
          enum:
            - url
            - b64_json
          default: url
          example: url
          nullable: true
          description: >-
            The format in which the generated images are returned. Must be one
            of `url` or `b64_json`. URLs are only valid for 60 minutes after the
            image has been generated.
        size:
          type: string
          enum:
            - 256x256
            - 512x512
            - 1024x1024
          default: 1024x1024
          example: 1024x1024
          nullable: true
          description: >-
            The size of the generated images. Must be one of `256x256`,
            `512x512`, or `1024x1024`.
        user:
          type: string
          example: user-1234
          description: >
            A unique identifier representing your end-user, which can help
            OpenAI to monitor and detect abuse. [Learn
            more](https://platform.openai.com/docs/guides/safety-best-practices/end-user-ids).
      required:
        - image
    ImagesResponse:
      properties:
        created:
          type: integer
        data:
          type: array
          items:
            $ref: '#/components/schemas/Image'
      required:
        - created
        - data
    Image:
      type: object
      description: >-
        Represents the url or the content of an image generated by the Portkey
        API.
      properties:
        b64_json:
          type: string
          description: >-
            The base64-encoded JSON of the generated image, if `response_format`
            is `b64_json`.
        url:
          type: string
          description: >-
            The URL of the generated image, if `response_format` is `url`
            (default).
        revised_prompt:
          type: string
          description: >-
            The prompt that was used to generate the image, if there was any
            revision to the prompt.
      x-code-samples:
        name: The image object
        example: |
          {
            "url": "...",
            "revised_prompt": "..."
          }
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