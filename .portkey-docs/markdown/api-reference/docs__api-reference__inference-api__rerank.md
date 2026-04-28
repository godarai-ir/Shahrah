> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Rerank

> Reranks a list of documents based on their relevance to a query. This endpoint provides a unified interface to reranking models from multiple providers including Cohere, Voyage, Jina, Pinecone, Bedrock, and Azure AI.

Reranking is useful for improving search results by scoring and sorting documents based on semantic relevance to a query, rather than just keyword matching.




## OpenAPI

````yaml post /rerank
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
  /rerank:
    servers:
      - url: https://api.portkey.ai/v1
        description: Portkey API Public Endpoint
      - url: SELF_HOSTED_GATEWAY_URL
        description: Self-Hosted Gateway URL
    post:
      tags:
        - Rerank
      summary: Rerank
      description: >
        Reranks a list of documents based on their relevance to a query. This
        endpoint provides a unified interface to reranking models from multiple
        providers including Cohere, Voyage, Jina, Pinecone, Bedrock, and Azure
        AI.


        Reranking is useful for improving search results by scoring and sorting
        documents based on semantic relevance to a query, rather than just
        keyword matching.
      operationId: createRerank
      parameters:
        - $ref: '#/components/parameters/PortkeyTraceId'
        - $ref: '#/components/parameters/PortkeySpanId'
        - $ref: '#/components/parameters/PortkeyParentSpanId'
        - $ref: '#/components/parameters/PortkeySpanName'
        - $ref: '#/components/parameters/PortkeyMetadata'
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateRerankRequest'
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CreateRerankResponse'
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
            curl https://api.portkey.ai/v1/rerank \
              -H "x-portkey-api-key: $PORTKEY_API_KEY" \
              -H "x-portkey-virtual-key: $PORTKEY_PROVIDER_VIRTUAL_KEY" \
              -H "Content-Type: application/json" \
              -d '{
                "model": "rerank-v3.5",
                "query": "What is the capital of France?",
                "documents": [
                  "Paris is the capital of France.",
                  "Berlin is the capital of Germany.",
                  "Madrid is the capital of Spain."
                ],
                "top_n": 2
              }'
        - lang: python
          label: Default
          source: |
            from portkey_ai import Portkey

            client = Portkey(
              api_key = "PORTKEY_API_KEY",
              virtual_key = "PROVIDER_VIRTUAL_KEY"
            )

            response = client.post(
              "/rerank",
              model="rerank-v3.5",
              query="What is the capital of France?",
              documents=[
                "Paris is the capital of France.",
                "Berlin is the capital of Germany.",
                "Madrid is the capital of Spain.",
              ],
              top_n=2,
            )

            print(response)
        - lang: javascript
          label: Default
          source: |
            import Portkey from 'portkey-ai';

            const client = new Portkey({
              apiKey: 'PORTKEY_API_KEY',
              virtualKey: 'PROVIDER_VIRTUAL_KEY'
            });

            async function main() {
              const response = await client.post('/rerank', {
                model: 'rerank-v3.5',
                query: 'What is the capital of France?',
                documents: [
                  'Paris is the capital of France.',
                  'Berlin is the capital of Germany.',
                  'Madrid is the capital of Spain.'
                ],
                top_n: 2
              });

              console.log(response);
            }

            main();
        - lang: curl
          label: Self-Hosted
          source: |
            curl -X POST "SELF_HOSTED_GATEWAY_URL/rerank" \
              -H "Content-Type: application/json" \
              -H "x-portkey-api-key: $PORTKEY_API_KEY" \
              -H "x-portkey-virtual-key: $PORTKEY_PROVIDER_VIRTUAL_KEY" \
              -d '{
                "model": "rerank-v3.5",
                "query": "What is the capital of France?",
                "documents": [
                  "Paris is the capital of France.",
                  "Berlin is the capital of Germany.",
                  "Madrid is the capital of Spain."
                ],
                "top_n": 2
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

            response = client.post(
              "/rerank",
              model="rerank-v3.5",
              query="What is the capital of France?",
              documents=[
                "Paris is the capital of France.",
                "Berlin is the capital of Germany.",
                "Madrid is the capital of Spain.",
              ],
              top_n=2,
            )

            print(response)
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
              const response = await client.post('/rerank', {
                model: 'rerank-v3.5',
                query: 'What is the capital of France?',
                documents: [
                  'Paris is the capital of France.',
                  'Berlin is the capital of Germany.',
                  'Madrid is the capital of Spain.'
                ],
                top_n: 2
              });

              console.log(response);
            }

            main();
components:
  parameters:
    PortkeyTraceId:
      in: header
      name: x-portkey-trace-id
      schema:
        type: string
      description: >-
        An ID you can pass to refer to one or more requests later on. If not
        provided, Portkey generates a trace ID automatically for each request.
        [Docs](https://portkey.ai/docs/product/observability/traces)
      required: false
    PortkeySpanId:
      in: header
      name: x-portkey-span-id
      schema:
        type: string
      description: An ID you can pass to refer to a span under a trace.
      required: false
    PortkeyParentSpanId:
      in: header
      name: x-portkey-parent-span-id
      schema:
        type: string
      description: Link a child span to a parent span
      required: false
    PortkeySpanName:
      in: header
      name: x-portkey-span-name
      schema:
        type: string
      description: Name for the Span ID
      required: false
    PortkeyMetadata:
      in: header
      name: x-portkey-metadata
      schema:
        type: object
      description: Pass any arbitrary metadata along with your request
      required: false
  schemas:
    CreateRerankRequest:
      type: object
      description: >
        Request body for reranking documents. The unified API supports multiple
        providers including Cohere, Voyage, Jina, Pinecone, Bedrock, and Azure
        AI.
      properties:
        model:
          description: >
            ID of the model to use for reranking. Model availability depends on
            the provider:

            - **Cohere**: `rerank-v3.5`, `rerank-english-v3.0`,
            `rerank-multilingual-v3.0`, `rerank-english-v2.0`,
            `rerank-multilingual-v2.0`

            - **Voyage**: `rerank-2`, `rerank-2-lite`

            - **Jina**: `jina-reranker-v2-base-multilingual`,
            `jina-reranker-v1-base-en`, `jina-reranker-v1-turbo-en`,
            `jina-reranker-v1-tiny-en`

            - **Pinecone**: `bge-reranker-v2-m3`, `pinecone-rerank-v0`

            - **Bedrock**: Model ARN (e.g.,
            `arn:aws:bedrock:us-west-2::foundation-model/cohere.rerank-v3-5:0`)

            - **Azure AI**: Cohere rerank deployments on Azure AI Inference; use
            the model name from your deployment, typically prefixed with
            `cohere.` (the gateway strips that prefix for the upstream request)
          type: string
          example: rerank-v3.5
        query:
          description: The search query to compare against the documents.
          type: string
          example: What is the capital of France?
        documents:
          description: >
            The list of documents to rerank. Each document can be a string or an
            object with a `text` field.

            The documents will be scored based on their relevance to the query.
          type: array
          items:
            $ref: '#/components/schemas/RerankDocument'
          minItems: 1
          example:
            - Paris is the capital of France.
            - Berlin is the capital of Germany.
            - Madrid is the capital of Spain.
        top_n:
          description: >
            The number of top results to return. If not specified, all documents
            are returned sorted by relevance.

            For Voyage, the gateway maps this field to the provider's `top_k`
            parameter.
          type: integer
          minimum: 1
          example: 3
        return_documents:
          description: >
            Whether to return the document text in the response. Supported by
            Voyage, Jina, and Pinecone.
          type: boolean
          default: false
        max_tokens_per_doc:
          description: >
            Maximum number of tokens per document. Documents exceeding this
            limit will be truncated. Cohere-specific parameter.
          type: integer
          minimum: 1
        priority:
          description: |
            Request priority hint. Cohere-specific parameter.
          type: number
        rank_fields:
          description: >
            The fields to use for ranking when documents are objects with
            multiple fields. Pinecone-specific parameter.
          type: array
          items:
            type: string
          example:
            - text
            - title
        truncation:
          description: >
            Whether to truncate documents that exceed the model's maximum
            context length. Voyage-specific parameter.
          type: boolean
        parameters:
          description: >
            Additional provider-specific parameters. Pinecone-specific
            parameter.
          type: object
          additionalProperties: true
      required:
        - model
        - query
        - documents
    CreateRerankResponse:
      type: object
      description: Response from the rerank endpoint.
      properties:
        id:
          type: string
          description: A unique identifier for the rerank request.
          example: rerank-abc123
        object:
          type: string
          description: The object type, which is always "list".
          enum:
            - list
          example: list
        results:
          type: array
          description: |
            The reranked results sorted by relevance score in descending order.
          items:
            $ref: '#/components/schemas/RerankResult'
        model:
          type: string
          description: The model used for reranking.
          example: rerank-v3.5
        usage:
          $ref: '#/components/schemas/RerankUsage'
        provider:
          type: string
          description: The provider that processed the request.
          example: cohere
      required:
        - object
        - results
        - model
    RerankDocument:
      description: >
        A document to be reranked. Can be a simple string or an object with a
        text field and optional metadata.
      oneOf:
        - type: string
          title: string
          description: A simple text string to be reranked.
          example: Paris is the capital of France.
        - type: object
          title: object
          description: An object containing the document text and optional metadata.
          properties:
            text:
              type: string
              description: The text content of the document.
              example: Paris is the capital of France.
          required:
            - text
          additionalProperties: true
    RerankResult:
      type: object
      description: A single reranked document result.
      properties:
        index:
          type: integer
          description: The index of the document in the original input array.
          example: 0
        relevance_score:
          type: number
          format: float
          description: >
            The relevance score of the document to the query. Higher scores
            indicate greater relevance.

            Score ranges vary by provider but are typically between 0 and 1.
          example: 0.98
        document:
          type: object
          description: >-
            The original document text. Only present if `return_documents` is
            true.
          properties:
            text:
              type: string
              description: The text content of the document.
          additionalProperties: true
      required:
        - index
        - relevance_score
    RerankUsage:
      type: object
      description: Usage information for the rerank request.
      properties:
        search_units:
          type: integer
          description: >
            The number of search units consumed by the request. Billing varies
            by provider.
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