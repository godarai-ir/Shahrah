> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# List Audit Logs



## OpenAPI

````yaml get /audit-logs
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
  /audit-logs:
    servers:
      - url: https://api.portkey.ai/v1
        description: Portkey API Public Endpoint
      - url: SELF_HOSTED_CONTROL_PLANE_URL
        description: Self-Hosted Control Plane URL
    get:
      tags:
        - Audit Logs
      summary: Get Audit Logs
      parameters:
        - name: start_time
          in: query
          required: true
          description: Start time for filtering logs (ISO8601 format)
          schema:
            type: string
        - name: end_time
          in: query
          required: true
          description: End time for filtering logs (ISO8601 format)
          schema:
            type: string
        - name: organisation_id
          in: query
          required: true
          description: Organisation ID for filtering logs
          schema:
            type: string
        - name: method
          in: query
          required: false
          description: HTTP method for filtering logs
          schema:
            type: string
            enum:
              - POST
              - PUT
              - DELETE
        - name: uri
          in: query
          required: false
          description: URI path for filtering logs
          schema:
            type: string
        - name: request_id
          in: query
          required: false
          description: Request ID for filtering logs
          schema:
            type: string
        - name: user_id
          in: query
          required: false
          description: User ID for filtering logs
          schema:
            type: string
        - name: user_type
          in: query
          required: false
          description: Type of user for filtering logs
          schema:
            type: string
            enum:
              - user
              - api_key
        - name: workspace_id
          in: query
          required: false
          description: Workspace ID for filtering logs
          schema:
            type: string
        - name: response_status_code
          in: query
          required: false
          description: HTTP response status code for filtering logs
          schema:
            type: integer
        - name: resource_type
          in: query
          required: false
          description: Resource type for filtering logs
          schema:
            type: string
        - name: action
          in: query
          required: false
          description: Action type for filtering logs
          schema:
            type: string
        - name: client_ip
          in: query
          required: false
          description: Client IP address for filtering logs
          schema:
            type: string
        - name: country
          in: query
          required: false
          description: Country for filtering logs
          schema:
            type: string
        - name: current_page
          in: query
          required: false
          description: Current page number for pagination
          schema:
            type: integer
            minimum: 0
        - name: page_size
          in: query
          required: false
          description: Number of items per page
          schema:
            type: integer
            minimum: 0
            maximum: 100
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/AuditLogObjectList'
components:
  schemas:
    AuditLogObjectList:
      type: object
      properties:
        records:
          type: array
          items:
            type: object
            properties:
              timestamp:
                type: string
                description: Timestamp of when the action occurred
              method:
                type: string
                enum:
                  - POST
                  - PUT
                  - DELETE
                description: HTTP method used for the request
              uri:
                type: string
                description: URI path that was accessed
              request_id:
                type: string
                description: Unique ID of the request
              request_body:
                type: string
                description: JSON string of the request body
              query_params:
                type: string
                description: JSON string of the query parameters
              request_headers:
                type: string
                description: JSON string of the request headers (partially masked)
              user_id:
                type: string
                format: uuid
                description: ID of the user who made the request
              user_type:
                type: string
                enum:
                  - user
                  - api_key
                description: Type of user who made the request
              organisation_id:
                type: string
                format: uuid
                description: ID of the organisation the user belongs to
              workspace_id:
                type: string
                description: ID of the workspace the resource belongs to
              response_status_code:
                type: integer
                description: HTTP status code of the response
              resource_type:
                type: string
                description: Type of resource that was accessed
              action:
                type: string
                description: Action performed on the resource
              client_ip:
                type: string
                description: IP address of the client
              country:
                type: string
                description: Country of origin based on the IP address
        total:
          type: integer
          description: Total number of records in the response
        object:
          type: string
          description: The type of object being returned
          enum:
            - analytics-graph
  securitySchemes:
    Portkey-Key:
      type: apiKey
      in: header
      name: x-portkey-api-key

````