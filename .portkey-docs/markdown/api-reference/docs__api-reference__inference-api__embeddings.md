> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Embeddings



## OpenAPI

````yaml post /embeddings
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
  /embeddings:
    servers:
      - url: https://api.portkey.ai/v1
        description: Portkey API Public Endpoint
      - url: SELF_HOSTED_GATEWAY_URL
        description: Self-Hosted Gateway URL
    post:
      tags:
        - Embeddings
      summary: Embeddings
      operationId: createEmbedding
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateEmbeddingRequest'
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CreateEmbeddingResponse'
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
            curl https://api.portkey.ai/v1/embeddings \
              -H "x-portkey-api-key: $PORTKEY_API_KEY" \
              -H "x-portkey-virtual-key: $PORTKEY_PROVIDER_VIRTUAL_KEY" \
              -H "Content-Type: application/json" \
              -d '{
                "input": "The food was delicious and the waiter...",
                "model": "text-embedding-ada-002",
                "encoding_format": "float"
              }'
        - lang: python
          label: Default
          source: |
            from portkey_ai import Portkey

            client = Portkey(
              api_key = "PORTKEY_API_KEY",
              virtual_key = "PROVIDER_VIRTUAL_KEY"
            )

            client.embeddings.create(
              model="text-embedding-ada-002",
              input="The food was delicious and the waiter...",
              encoding_format="float"
            )
        - lang: javascript
          label: Default
          source: |
            import Portkey from 'portkey-ai';

            const client = new Portkey({
              apiKey: 'PORTKEY_API_KEY',
              virtualKey: 'PROVIDER_VIRTUAL_KEY'
            });

            async function main() {
              const embedding = await client.embeddings.create({
                model: "text-embedding-ada-002",
                input: "The quick brown fox jumped over the lazy dog",
                encoding_format: "float",
              });

              console.log(embedding);
            }

            main();
        - lang: curl
          label: Self-Hosted
          source: |
            curl -X POST "SELF_HOSTED_GATEWAY_URL/embeddings" \
              -H "Content-Type: application/json" \
              -H "x-portkey-api-key: $PORTKEY_API_KEY" \
              -H "x-portkey-virtual-key: $PORTKEY_PROVIDER_VIRTUAL_KEY" \
              -d '{
                "input": "The food was delicious and the waiter...",
                "model": "text-embedding-ada-002",
                "encoding_format": "float"
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

            response = client.embeddings.create(
                model="text-embedding-ada-002",
                input="The food was delicious and the waiter...",
                encoding_format="float"
            )

            print(response.data)
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
              const embedding = await portkey.embeddings.create({
                model: "text-embedding-ada-002",
                input: "The quick brown fox jumped over the lazy dog",
                encoding_format: "float",
              });

              console.log(embedding);
            }

            main();
components:
  schemas:
    CreateEmbeddingRequest:
      type: object
      additionalProperties: false
      properties:
        input:
          description: >
            Input text to embed, encoded as a string or array of tokens. To
            embed multiple inputs in a single request, pass an array of strings
            or array of token arrays. The input must not exceed the max input
            tokens for the model (8192 tokens for `text-embedding-ada-002`),
            cannot be an empty string, and any array must be 2048 dimensions or
            less. [Example Python
            code](https://cookbook.openai.com/examples/how_to_count_tokens_with_tiktoken)
            for counting tokens.
          example: The quick brown fox jumped over the lazy dog
          oneOf:
            - type: string
              title: string
              description: The string that will be turned into an embedding.
              default: ''
              example: This is a test.
            - type: array
              title: array
              description: The array of strings that will be turned into an embedding.
              minItems: 1
              maxItems: 2048
              items:
                type: string
                default: ''
                example: '[''This is a test.'']'
            - type: array
              title: array
              description: The array of integers that will be turned into an embedding.
              minItems: 1
              maxItems: 2048
              items:
                type: integer
              example: '[1212, 318, 257, 1332, 13]'
            - type: array
              title: array
              description: >-
                The array of arrays containing integers that will be turned into
                an embedding.
              minItems: 1
              maxItems: 2048
              items:
                type: array
                minItems: 1
                items:
                  type: integer
              example: '[[1212, 318, 257, 1332, 13]]'
          x-oaiExpandable: true
        model:
          description: >
            ID of the model to use. You can use the [List
            models](https://platform.openai.com/docs/api-reference/models/list)
            API to see all of your available models, or see our [Model
            overview](https://platform.openai.com/docs/models/overview) for
            descriptions of them.
          example: text-embedding-3-small
          anyOf:
            - type: string
            - type: string
              enum:
                - text-embedding-ada-002
                - text-embedding-3-small
                - text-embedding-3-large
          x-oaiTypeLabel: string
        encoding_format:
          description: >-
            The format to return the embeddings in. Can be either `float` or
            [`base64`](https://pypi.org/project/pybase64/).
          example: float
          default: float
          type: string
          enum:
            - float
            - base64
        dimensions:
          description: >
            The number of dimensions the resulting output embeddings should
            have. Only supported in `text-embedding-3` and later models.
          type: integer
          minimum: 1
        user:
          type: string
          example: user-1234
          description: >
            A unique identifier representing your end-user, which can help
            OpenAI to monitor and detect abuse. [Learn
            more](https://platform.openai.com/docs/guides/safety-best-practices/end-user-ids).
      required:
        - model
        - input
    CreateEmbeddingResponse:
      type: object
      properties:
        data:
          type: array
          description: The list of embeddings generated by the model.
          items:
            $ref: '#/components/schemas/Embedding'
        model:
          type: string
          description: The name of the model used to generate the embedding.
        object:
          type: string
          description: The object type, which is always "list".
          enum:
            - list
        usage:
          type: object
          description: The usage information for the request.
          properties:
            prompt_tokens:
              type: integer
              description: The number of tokens used by the prompt.
            total_tokens:
              type: integer
              description: The total number of tokens used by the request.
          required:
            - prompt_tokens
            - total_tokens
      required:
        - object
        - model
        - data
        - usage
    Embedding:
      type: object
      description: |
        Represents an embedding vector returned by embedding endpoint.
      properties:
        index:
          type: integer
          description: The index of the embedding in the list of embeddings.
        embedding:
          type: array
          description: >
            The embedding vector, which is a list of floats. The length of
            vector depends on the model as listed in the [embedding
            guide](https://platform.openai.com/docs/guides/embeddings).
          items:
            type: number
        object:
          type: string
          description: The object type, which is always "embedding".
          enum:
            - embedding
      required:
        - index
        - object
        - embedding
      x-code-samples:
        name: The embedding object
        example: |
          {
            "object": "embedding",
            "embedding": [
              0.0023064255,
              -0.009327292,
              .... (1536 floats total for ada-002)
              -0.0028842222,
            ],
            "index": 0
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