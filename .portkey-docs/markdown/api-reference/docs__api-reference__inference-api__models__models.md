> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Models

> Lists the currently available models that can be used through Portkey, and provides basic information about each one.



## OpenAPI

````yaml get /models
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
  /models:
    servers:
      - url: https://api.portkey.ai/v1
        description: Portkey API Public Endpoint
      - url: SELF_HOSTED_GATEWAY_URL
        description: Self-Hosted Gateway URL
    get:
      tags:
        - Models
      summary: List Available Models
      description: >-
        Lists the currently available models that can be used through Portkey,
        and provides basic information about each one.
      operationId: listModels
      parameters:
        - in: query
          name: ai_service
          required: false
          description: Filter models by the AI service (e.g., 'openai', 'anthropic').
          schema:
            type: string
        - in: query
          name: provider
          required: false
          description: Filter models by the provider.
          schema:
            type: string
        - in: query
          name: limit
          required: false
          description: The maximum number of models to return.
          schema:
            type: integer
        - in: query
          name: offset
          required: false
          description: >-
            The number of models to skip before starting to collect the result
            set.
          schema:
            type: integer
        - in: query
          name: sort
          required: false
          description: The field to sort the results by.
          schema:
            type: string
            enum:
              - name
              - provider
              - ai_service
            default: name
        - in: query
          name: order
          required: false
          description: The order to sort the results in.
          schema:
            type: string
            enum:
              - asc
              - desc
            default: asc
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ListModelsResponse'
              example:
                object: list
                total: 500
                data:
                  - id: '@ai-provider-slug/gpt-5'
                    slug: gpt-5
                    canonical_slug: gpt-5
                    object: model
      security:
        - Portkey-Key: []
      x-code-samples:
        - lang: curl
          label: Default
          source: |
            # Example of sending a query parameter in the URL
            curl 'https://api.portkey.ai/v1/models?provider=openai' \
              -H "x-portkey-api-key: $PORTKEY_API_KEY"
        - lang: curl
          label: Self-Hosted
          source: |
            # Example of sending a query parameter in the URL
            curl 'https://YOUR_SELF_HOSTED_URL/models?provider=openai' \
              -H "x-portkey-api-key: $PORTKEY_API_KEY"
        - lang: python
          label: Default
          source: |
            from portkey_ai import Portkey

            client = Portkey(
              api_key = "PORTKEY_API_KEY"
            )

            # Example of sending query parameters via extra_query
            models = client.models.list(
              extra_query={"provider": "openai"}
            )
            print(models)
        - lang: python
          label: Self-Hosted
          source: |
            from portkey_ai import Portkey

            client = Portkey(
              api_key = "PORTKEY_API_KEY",
              base_url = "https://YOUR_SELF_HOSTED_URL"
            )

            # Example of sending query parameters via extra_query
            models = client.models.list(
              extra_query={"provider": "openai"}
            )
            print(models)
        - lang: javascript
          label: Default
          source: |
            import Portkey from 'portkey-ai';

            const client = new Portkey({
              apiKey: 'PORTKEY_API_KEY'
            });

            async function main() {
              // Example of sending query parameters in the list method
              const list = await client.models.list({
                provider: "openai"
              });
              console.log(list);
            }
            main();
        - lang: javascript
          label: Self-Hosted
          source: |
            import Portkey from 'portkey-ai';

            const client = new Portkey({
              apiKey: 'PORTKEY_API_KEY',
              baseUrl: 'https://YOUR_SELF_HOSTED_URL'
            });

            async function main() {
              // Example of sending query parameters in the list method
              const list = await client.models.list({
                provider: "openai"
              });
              console.log(list);
            }
            main();     
components:
  schemas:
    ListModelsResponse:
      type: object
      properties:
        object:
          type: string
          enum:
            - list
        data:
          type: array
          items:
            $ref: '#/components/schemas/Model'
      required:
        - object
        - data
    Model:
      title: Model
      description: Describes an OpenAI model offering that can be used with the API.
      properties:
        id:
          type: string
          description: The model identifier, which can be referenced in the API endpoints.
        created:
          type: integer
          description: The Unix timestamp (in seconds) when the model was created.
        object:
          type: string
          description: The object type, which is always "model".
          enum:
            - model
        owned_by:
          type: string
          description: The organization that owns the model.
      required:
        - id
        - object
        - created
        - owned_by
  securitySchemes:
    Portkey-Key:
      type: apiKey
      in: header
      name: x-portkey-api-key

````