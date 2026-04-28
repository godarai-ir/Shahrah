> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Moderations



## OpenAPI

````yaml post /moderations
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
  /moderations:
    servers:
      - url: https://api.portkey.ai/v1
        description: Portkey API Public Endpoint
      - url: SELF_HOSTED_GATEWAY_URL
        description: Self-Hosted Gateway URL
    post:
      tags:
        - Moderations
      summary: >
        Identify potentially harmful content in text and images. **Only** works
        with [OpenAI's Moderations
        endpoint](https://platform.openai.com/docs/guides/moderation) currently.
      operationId: createModeration
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateModerationRequest'
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CreateModerationResponse'
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
            curl https://api.portkey.ai/v1/moderations \
              -H "Content-Type: application/json" \
              -H "x-portkey-api-key: $PORTKEY_API_KEY" \
              -H "x-portkey-virtual-key: $PORTKEY_PROVIDER_VIRTUAL_KEY" \
              -d '{
                "input": "I want to kill them."
              }'
        - lang: python
          label: Default
          source: |
            from portkey_ai import Portkey

            client = Portkey(
              api_key = "PORTKEY_API_KEY",
              virtual_key = "PROVIDER_VIRTUAL_KEY"
            )

            moderation = client.moderations.create(input="I want to kill them.")
            print(moderation)
        - lang: javascript
          label: Default
          source: |
            import Portkey from 'portkey-ai';

            const client = new Portkey({
              apiKey: 'PORTKEY_API_KEY',
              virtualKey: 'PROVIDER_VIRTUAL_KEY'
            });

            async function main() {
              const moderation = await client.moderations.create({ input: "I want to kill them." });

              console.log(moderation);
            }
            main();
        - lang: curl
          label: Self-Hosted
          source: |
            curl https://SELF_HOSTED_GATEWAY_URL/moderations \
              -H "Content-Type: application/json" \
              -H "x-portkey-api-key: $PORTKEY_API_KEY" \
              -H "x-portkey-virtual-key: $PORTKEY_PROVIDER_VIRTUAL_KEY" \
              -d '{
                "input": "I want to kill them."
              }'
        - lang: python
          label: Self-Hosted
          source: |
            from portkey_ai import Portkey

            client = Portkey(
              api_key = "PORTKEY_API_KEY",
              base_url="SELF_HOSTED_GATEWAY_URL",
              virtual_key = "PROVIDER_VIRTUAL_KEY"
            )

            moderation = client.moderations.create(input="I want to kill them.")
            print(moderation)
        - lang: javascript
          label: Self-Hosted
          source: |
            import Portkey from 'portkey-ai';

            const client = new Portkey({
              apiKey: 'PORTKEY_API_KEY',
              virtualKey: 'PROVIDER_VIRTUAL_KEY',
              baseURL: 'SELF_HOSTED_GATEWAY_URL'
            });

            async function main() {
              const moderation = await client.moderations.create({ input: "I want to kill them." });

              console.log(moderation);
            }
            main();
components:
  schemas:
    CreateModerationRequest:
      type: object
      properties:
        input:
          description: The input text to classify
          oneOf:
            - type: string
              default: ''
              example: I want to kill them.
            - type: array
              items:
                type: string
                default: ''
                example: I want to kill them.
        model:
          description: >
            Two content moderations models are available:
            `text-moderation-stable` and `text-moderation-latest`.


            The default is `text-moderation-latest` which will be automatically
            upgraded over time. This ensures you are always using our most
            accurate model. If you use `text-moderation-stable`, we will provide
            advanced notice before updating the model. Accuracy of
            `text-moderation-stable` may be slightly lower than for
            `text-moderation-latest`.
          nullable: false
          default: text-moderation-latest
          example: text-moderation-stable
          anyOf:
            - type: string
            - type: string
              enum:
                - text-moderation-latest
                - text-moderation-stable
          x-oaiTypeLabel: string
      required:
        - input
    CreateModerationResponse:
      type: object
      description: Represents if a given text input is potentially harmful.
      properties:
        id:
          type: string
          description: The unique identifier for the moderation request.
        model:
          type: string
          description: The model used to generate the moderation results.
        results:
          type: array
          description: A list of moderation objects.
          items:
            type: object
            properties:
              flagged:
                type: boolean
                description: Whether any of the below categories are flagged.
              categories:
                type: object
                description: A list of the categories, and whether they are flagged or not.
                properties:
                  hate:
                    type: boolean
                    description: >-
                      Content that expresses, incites, or promotes hate based on
                      race, gender, ethnicity, religion, nationality, sexual
                      orientation, disability status, or caste. Hateful content
                      aimed at non-protected groups (e.g., chess players) is
                      harassment.
                  hate/threatening:
                    type: boolean
                    description: >-
                      Hateful content that also includes violence or serious
                      harm towards the targeted group based on race, gender,
                      ethnicity, religion, nationality, sexual orientation,
                      disability status, or caste.
                  harassment:
                    type: boolean
                    description: >-
                      Content that expresses, incites, or promotes harassing
                      language towards any target.
                  harassment/threatening:
                    type: boolean
                    description: >-
                      Harassment content that also includes violence or serious
                      harm towards any target.
                  self-harm:
                    type: boolean
                    description: >-
                      Content that promotes, encourages, or depicts acts of
                      self-harm, such as suicide, cutting, and eating disorders.
                  self-harm/intent:
                    type: boolean
                    description: >-
                      Content where the speaker expresses that they are engaging
                      or intend to engage in acts of self-harm, such as suicide,
                      cutting, and eating disorders.
                  self-harm/instructions:
                    type: boolean
                    description: >-
                      Content that encourages performing acts of self-harm, such
                      as suicide, cutting, and eating disorders, or that gives
                      instructions or advice on how to commit such acts.
                  sexual:
                    type: boolean
                    description: >-
                      Content meant to arouse sexual excitement, such as the
                      description of sexual activity, or that promotes sexual
                      services (excluding sex education and wellness).
                  sexual/minors:
                    type: boolean
                    description: >-
                      Sexual content that includes an individual who is under 18
                      years old.
                  violence:
                    type: boolean
                    description: Content that depicts death, violence, or physical injury.
                  violence/graphic:
                    type: boolean
                    description: >-
                      Content that depicts death, violence, or physical injury
                      in graphic detail.
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
              category_scores:
                type: object
                description: >-
                  A list of the categories along with their scores as predicted
                  by model.
                properties:
                  hate:
                    type: number
                    description: The score for the category 'hate'.
                  hate/threatening:
                    type: number
                    description: The score for the category 'hate/threatening'.
                  harassment:
                    type: number
                    description: The score for the category 'harassment'.
                  harassment/threatening:
                    type: number
                    description: The score for the category 'harassment/threatening'.
                  self-harm:
                    type: number
                    description: The score for the category 'self-harm'.
                  self-harm/intent:
                    type: number
                    description: The score for the category 'self-harm/intent'.
                  self-harm/instructions:
                    type: number
                    description: The score for the category 'self-harm/instructions'.
                  sexual:
                    type: number
                    description: The score for the category 'sexual'.
                  sexual/minors:
                    type: number
                    description: The score for the category 'sexual/minors'.
                  violence:
                    type: number
                    description: The score for the category 'violence'.
                  violence/graphic:
                    type: number
                    description: The score for the category 'violence/graphic'.
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
            required:
              - flagged
              - categories
              - category_scores
      required:
        - id
        - model
        - results
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