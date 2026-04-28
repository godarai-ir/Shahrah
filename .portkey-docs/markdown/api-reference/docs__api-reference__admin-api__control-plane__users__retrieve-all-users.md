> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Retrieve all users



## OpenAPI

````yaml get /admin/users
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
  /admin/users:
    servers:
      - url: https://api.portkey.ai/v1
      - url: https://SELF_HOSTED_CONTROL_PLANE_URL
    get:
      tags:
        - Users
      summary: Get users
      parameters:
        - name: x-portkey-api-key
          in: header
          schema:
            type: string
          example: '{{PORTKEY_API_KEY}}'
        - name: pageSize
          in: query
          schema:
            type: integer
          example: '1'
        - name: currentPage
          in: query
          schema:
            type: integer
          example: '0'
        - name: role
          in: query
          schema:
            type: string
            enum:
              - admin
              - member
              - owner
          example: admin
        - name: email
          in: query
          schema:
            type: string
            format: email
          example: foo@bar.com
      responses:
        '200':
          description: OK
          headers:
            Content-Type:
              schema:
                type: string
                example: application/json
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/UserList'
              example:
                total: 2
                object: list
                data:
                  - object: user
                    id: 61e08f60-4822-465e-ba23-39f85cd741cb
                    first_name: horace
                    last_name: slughorn
                    role: member
                    email: horace.slughorn@example.com
                    created_at: '2024-01-25 11:35:07'
                    last_updated_at: '2024-01-25 11:35:07'
      x-code-samples:
        - lang: python
          label: Default
          source: |
            from portkey_ai import Portkey

            # Initialize the Portkey client
            portkey = Portkey(
                api_key="PORTKEY_API_KEY",
            )

            # List users
            users = portkey.admin.users.list()

            print(users)
        - lang: javascript
          label: Default
          source: |
            import { Portkey } from "portkey-ai";

            const portkey = new Portkey({
                apiKey: "PORTKEY_API_KEY",
            })

            const users=await portkey.admin.users.list({})

            console.log(users);
        - lang: curl
          label: Default
          source: |
            curl -X GET "https://api.portkey.ai/v1/admin/users"
              -H "x-portkey-api-key: PORTKEY_API_KEY"
        - lang: curl
          label: Self-Hosted
          source: |
            curl -X GET "SELF_HOSTED_CONTROL_PLANE_URL/admin/users"
              -H "x-portkey-api-key: PORTKEY_API_KEY"
        - lang: python
          label: Self-Hosted
          source: |
            from portkey_ai import Portkey

            # Initialize the Portkey client
            portkey = Portkey(
                api_key="PORTKEY_API_KEY",
                base_url="SELF_HOSTED_CONTROL_PLANE_URL"
            )

            # List users
            users = portkey.admin.users.list()

            print(users)
        - lang: javascript
          label: Self-Hosted
          source: |
            import { Portkey } from "portkey-ai";

            const portkey = new Portkey({
                apiKey: "PORTKEY_API_KEY",
                baseUrl: "SELF_HOSTED_CONTROL_PLANE_URL"
            })

            const users=await portkey.admin.users.list({})

            console.log(users);
components:
  schemas:
    UserList:
      type: object
      properties:
        total:
          type: integer
          example: 2
        object:
          type: string
          enum:
            - list
        data:
          type: array
          items:
            $ref: '#/components/schemas/User'
    User:
      type: object
      properties:
        object:
          type: string
          enum:
            - user
        id:
          type: string
          format: uuid
        first_name:
          type: string
        last_name:
          type: string
        role:
          type: string
          enum:
            - admin
            - member
            - owner
        email:
          type: string
          format: email
        created_at:
          type: string
          format: date-time
        last_updated_at:
          type: string
          format: date-time
        workspace_ids:
          type: array
          items:
            type: string
  securitySchemes:
    Portkey-Key:
      type: apiKey
      in: header
      name: x-portkey-api-key

````