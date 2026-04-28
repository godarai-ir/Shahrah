> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Get User Grouped Data



## OpenAPI

````yaml get /analytics/groups/users
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
  /analytics/groups/users:
    servers:
      - url: https://api.portkey.ai/v1
        description: Portkey API Public Endpoint
      - url: SELF_HOSTED_CONTROL_PLANE_URL
        description: Self-Hosted Control Plane URL
    get:
      tags:
        - Analytics > Groups
      summary: Get metadata users grouped data.
      parameters:
        - $ref: '#/components/parameters/WorkspaceSlug'
        - $ref: '#/components/parameters/TimeOfGenerationMin'
        - $ref: '#/components/parameters/TimeOfGenerationMax'
        - $ref: '#/components/parameters/TotalUnitsMin'
        - $ref: '#/components/parameters/TotalUnitsMax'
        - $ref: '#/components/parameters/CostMin'
        - $ref: '#/components/parameters/CostMax'
        - $ref: '#/components/parameters/PromptTokenMin'
        - $ref: '#/components/parameters/PromptTokenMax'
        - $ref: '#/components/parameters/CompletionTokenMin'
        - $ref: '#/components/parameters/CompletionTokenMax'
        - $ref: '#/components/parameters/StatusCode'
        - $ref: '#/components/parameters/WeightedFeedbackMin'
        - $ref: '#/components/parameters/WeightedFeedbackMax'
        - $ref: '#/components/parameters/VirtualKeys'
        - $ref: '#/components/parameters/Configs'
        - $ref: '#/components/parameters/ApiKeyIds'
        - $ref: '#/components/parameters/CurrentPage'
        - $ref: '#/components/parameters/PageSize'
        - $ref: '#/components/parameters/Metadata'
        - $ref: '#/components/parameters/AiOrgModel'
        - $ref: '#/components/parameters/TraceId'
        - $ref: '#/components/parameters/SpanId'
        - $ref: '#/components/parameters/PromptSlug'
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
                type: object
                properties:
                  object:
                    type: string
                    enum:
                      - list
                  total:
                    type: integer
                    description: Total records present across all pages
                  data:
                    type: array
                    items:
                      type: object
                      properties:
                        user:
                          type: string
                          description: The user for which the data is calculated
                        requests:
                          type: string
                          description: Total requests made by this user
                        cost:
                          type: string
                          description: >-
                            Total cost in cents for the requests made by this
                            user
                        object:
                          type: string
                          description: The type of object being returned
                          enum:
                            - analytics-group
                required:
                  - total
                  - object
                  - data
components:
  parameters:
    WorkspaceSlug:
      in: query
      name: workspace_slug
      required: true
      schema:
        type: string
      description: >-
        Workspace slug filter. If a workspace API key is being used, this filter
        will not be taken into consideration. If an organisation API key is used
        and no workspace slug is passed, default workspace will be used.
    TimeOfGenerationMin:
      in: query
      name: time_of_generation_min
      required: true
      schema:
        type: string
        format: date-time
        example: '2026-02-23T14:20:31+05:30'
      description: >-
        Minimum time of generation in ISO8601 format
        (YYYY-MM-DDTHH:MM:SS±HH:MM).
    TimeOfGenerationMax:
      in: query
      name: time_of_generation_max
      required: true
      schema:
        type: string
        format: date-time
        example: '2026-02-24T14:20:31+05:30'
      description: >-
        Maximum time of generation in ISO8601 format
        (YYYY-MM-DDTHH:MM:SS±HH:MM).
    TotalUnitsMin:
      in: query
      name: total_units_min
      schema:
        type: integer
        minimum: 0
      description: Minimum total units (tokens)
    TotalUnitsMax:
      in: query
      name: total_units_max
      schema:
        type: integer
        minimum: 0
      description: Maximum total units (tokens)
    CostMin:
      in: query
      name: cost_min
      schema:
        type: number
        minimum: 0
      description: Minimum cost (in cents)
    CostMax:
      in: query
      name: cost_max
      schema:
        type: number
        minimum: 0
      description: Maximum cost (in cents)
    PromptTokenMin:
      in: query
      name: prompt_token_min
      schema:
        type: integer
        minimum: 0
      description: Minimum number of prompt tokens
    PromptTokenMax:
      in: query
      name: prompt_token_max
      schema:
        type: integer
        minimum: 0
      description: Maximum number of prompt tokens
    CompletionTokenMin:
      in: query
      name: completion_token_min
      schema:
        type: integer
        minimum: 0
      description: Minimum number of completion tokens
    CompletionTokenMax:
      in: query
      name: completion_token_max
      schema:
        type: integer
        minimum: 0
      description: Maximum number of completion tokens
    StatusCode:
      in: query
      name: status_code
      schema:
        type: string
      description: Comma separated response status codes
      example: 401,403
    WeightedFeedbackMin:
      in: query
      name: weighted_feedback_min
      schema:
        type: number
        minimum: -10
        maximum: 10
      description: Minimum weighted feedback score
    WeightedFeedbackMax:
      in: query
      name: weighted_feedback_max
      schema:
        type: number
        minimum: -10
        maximum: 10
      description: Maximum weighted feedback score
    VirtualKeys:
      in: query
      name: virtual_keys
      schema:
        type: string
      description: Comma separated virtual key slugs
      example: vk-slug-1,vk-slug-2
    Configs:
      in: query
      name: configs
      schema:
        type: string
      description: Comma separated config slugs
      example: pc-config-slug-1,pc-config-slug-2
    ApiKeyIds:
      in: query
      name: api_key_ids
      schema:
        type: string
      description: Comma separated API key UUIDs
      example: >-
        765768a9-b4ec-4694-962c-d55f40cdb0dc,7c22af5a-8119-46b8-8d9b-bad3ad382387
    CurrentPage:
      in: query
      name: current_page
      schema:
        type: integer
        minimum: 0
      description: Current page number
    PageSize:
      in: query
      name: page_size
      schema:
        type: integer
        minimum: 0
      description: Number of items per page
    Metadata:
      in: query
      name: metadata
      schema:
        type: string
      description: Stringifed json object with key value metadata pairs
      example: '{"_user":"user_1", "env": "staging"}'
    AiOrgModel:
      in: query
      name: ai_org_model
      schema:
        type: string
      description: >-
        Comma separated ai provider and model combination. Double underscore
        (__) should be used as a separator for each provider and model
        combination
      example: openai__gpt-3.5-turbo,azure-openai__gpt-35-turbo
    TraceId:
      in: query
      name: trace_id
      schema:
        type: string
      description: Comma separated trace IDs
      example: my-unique-trace-1,my-unique-trace-2
    SpanId:
      in: query
      name: span_id
      schema:
        type: string
      description: Comma separated span IDs
      example: my-unique-span-1,my-unique-span-2
    PromptSlug:
      in: query
      name: prompt_slug
      schema:
        type: string
      description: Comma separated prompt slugs
      example: prompt-slug-1,prompt-slug-2
  securitySchemes:
    Portkey-Key:
      type: apiKey
      in: header
      name: x-portkey-api-key

````