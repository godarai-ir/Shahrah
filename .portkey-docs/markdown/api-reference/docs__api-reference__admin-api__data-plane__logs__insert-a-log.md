> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Insert a Log

> Submit one or more log entries

The log object comprises of 3 parts:

| Part       | Accepted Values                                                                                                                    |
| :--------- | :--------------------------------------------------------------------------------------------------------------------------------- |
| `request`  | `url`, `provider`, `headers`, `method` (defaults to `post`), and `body`                                                            |
| `response` | `status` (defaults to 200), `headers`, `body`, `time` (response latency), and `streamingMode` (defaults to false), `response_time` |
| `metadata` | `organization`, `user`, tracing info (`traceId`, `spanId`, `spanName`, `parentSpanId`), and any `key:value` pair                   |


## OpenAPI

````yaml post /logs
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
  /logs:
    servers:
      - url: https://api.portkey.ai/v1
        description: Portkey API Public Endpoint
      - url: SELF_HOSTED_GATEWAY_URL
        description: Self-Hosted Gateway URL
    post:
      tags:
        - Logs
      summary: Insert New logs
      description: Submit one or more log entries
      requestBody:
        required: true
        content:
          application/json:
            schema:
              oneOf:
                - $ref: '#/components/schemas/CustomLog'
                - type: array
                  items:
                    $ref: '#/components/schemas/CustomLog'
      responses:
        '200':
          description: Successful response
      x-code-samples:
        - lang: python
          label: Default
          source: >
            from portkey_ai import Portkey


            portkey = Portkey(
              api_key="PORTKEY_API_KEY",
            )


            request = {
              "url": "https://api.someprovider.com/model/generate",
              "method": "POST",
              "headers": {"Content-Type": "application/json"},
              "body": {"prompt": "What is AI?"},
            }

            response = {
              "status": 200,
              "headers": {"Content-Type": "application/json"},
              "body": {"response": "AI stands for Artificial Intelligence..."},
              "response_time": 123,
            }

            metadata = {
              "user_id": "123",
              "user_name": "John Doe",
            }


            result = portkey.logs.create(request=request, response=response,
            metadata=metadata)


            print(result)
        - lang: javascript
          label: Default
          source: |
            import Portkey from "portkey-ai";

            const portkey = new Portkey({
              apiKey:"PORTKEY_API_KEY"
            })

            async function main() {
              const request = {
                url: "https://api.someprovider.com/model/generate",
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: { prompt: "What is AI?" },
              };
              const response = {
                status: 200,
                headers: { "Content-Type": "application/json" },
                body: { response: "AI stands for Artificial Intelligence..." },
                response_time: 123,
              };
              const metadata = {
                user_id: "123",
                user_name: "John Doe",
              };
              const result = await portkey.logs.create({
                request: request,
                response: response,
                metadata: metadata,
              });
              console.log(result);
            }

            main();
        - lang: curl
          label: Default
          source: |
            curl -X POST "https://api.portkey.ai/v1/logs" \
            -H "x-portkey-api-key: PORTKEY_API_KEY" \
            -H "Content-Type: application/json" \
            -d '{
              "request": {
                "url": "https://api.someprovider.com/model/generate",
                "method": "POST",
                "headers": { "Content-Type": "application/json" },
                "body": { "prompt": "What is AI?" }
              },
              "response": {
                "status": 200,
                "headers": { "Content-Type": "application/json" },
                "body": { "response": "AI stands for Artificial Intelligence..." },
                "response_time": 123
              },
              "metadata": {
                "user_id": "123",
                "user_name": "John Doe"
              }
            }'
        - lang: curl
          label: Self-Hosted
          source: |
            curl -X POST "SELF_HOSTED_GATEWAY_URL/logs" \
            -H "x-portkey-api-key: PORTKEY_API_KEY" \
            -H "Content-Type: application/json" \
            -d '{
              "request": {
                "url": "https://api.someprovider.com/model/generate",
                "method": "POST",
                "headers": { "Content-Type": "application/json" },
                "body": { "prompt": "What is AI?" }
              },
              "response": {
                "status": 200,
                "headers": { "Content-Type": "application/json" },
                "body": { "response": "AI stands for Artificial Intelligence..." },
                "response_time": 123
              },
              "metadata": {
                "user_id": "123",
                "user_name": "John Doe"
              }
            }'
        - lang: python
          label: Self-Hosted
          source: >
            from portkey_ai import Portkey


            portkey = Portkey(
              api_key="PORTKEY_API_KEY",
              base_url="SELF_HOSTED_GATEWAY_URL"
            )


            request = {
              "url": "https://api.someprovider.com/model/generate",
              "method": "POST",
              "headers": {"Content-Type": "application/json"},
              "body": {"prompt": "What is AI?"},
            }

            response = {
              "status": 200,
              "headers": {"Content-Type": "application/json"},
              "body": {"response": "AI stands for Artificial Intelligence..."},
              "response_time": 123,
            }

            metadata = {
              "user_id": "123",
              "user_name": "John Doe",
            }


            result = portkey.logs.create(request=request, response=response,
            metadata=metadata)


            print(result)
        - lang: javascript
          label: Self-Hosted
          source: |
            import Portkey from "portkey-ai";

            const portkey = new Portkey({
              apiKey:"PORTKEY_API_KEY",
              baseUrl: "SELF_HOSTED_GATEWAY_URL"
            })

            async function main() {
              const request = {
                url: "https://api.someprovider.com/model/generate",
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: { prompt: "What is AI?" },
              };
              const response = {
                status: 200,
                headers: { "Content-Type": "application/json" },
                body: { response: "AI stands for Artificial Intelligence..." },
                response_time: 123,
              };
              const metadata = {
                user_id: "123",
                user_name: "John Doe",
              };
              const result = await portkey.logs.create({
                request: request,
                response: response,
                metadata: metadata,
              });
              console.log(result);
            }

            main();
components:
  schemas:
    CustomLog:
      type: object
      properties:
        request:
          type: object
          properties:
            url:
              type: string
            method:
              type: string
            headers:
              type: object
              additionalProperties:
                type: string
            body:
              type: object
          required:
            - url
            - body
        response:
          type: object
          properties:
            status:
              type: integer
            headers:
              type: object
              additionalProperties:
                type: string
            body:
              type: object
            response_time:
              type: integer
          required:
            - body
        metadata:
          type: object
          properties:
            trace_id:
              type: string
            span_id:
              type: string
            span_name:
              type: string
            additionalProperties:
              type: string
      required:
        - request
        - response
  securitySchemes:
    Portkey-Key:
      type: apiKey
      in: header
      name: x-portkey-api-key

````