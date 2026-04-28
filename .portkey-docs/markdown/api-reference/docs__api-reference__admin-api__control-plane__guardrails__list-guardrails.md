> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# List Guardrails

> Retrieves a paginated list of guardrails for the specified workspace or organisation



## OpenAPI

````yaml get /guardrails
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
  /guardrails:
    get:
      tags:
        - Guardrails
      summary: List guardrails
      description: >-
        Retrieves a paginated list of guardrails for the specified workspace or
        organisation
      operationId: listGuardrails
      parameters:
        - name: workspace_id
          in: query
          description: Workspace UUID to filter guardrails
          schema:
            type: string
            format: uuid
        - name: organisation_id
          in: query
          description: Organisation UUID to filter guardrails
          schema:
            type: string
            format: uuid
        - name: page_size
          in: query
          description: Number of items per page
          schema:
            type: integer
            minimum: 1
            maximum: 1000
            default: 100
        - name: current_page
          in: query
          description: Current page number (0-indexed)
          schema:
            type: integer
            minimum: 0
            default: 0
      responses:
        '200':
          description: List of guardrails retrieved successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ListGuardrailsResponse'
        '400':
          description: Bad request - invalid parameters
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '403':
          description: Forbidden - insufficient permissions
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
components:
  schemas:
    ListGuardrailsResponse:
      type: object
      required:
        - data
        - total
      properties:
        data:
          type: array
          description: Array of guardrail summaries
          items:
            $ref: '#/components/schemas/GuardrailSummary'
        total:
          type: integer
          description: Total number of guardrails available
          minimum: 0
    ErrorResponse:
      type: object
      properties:
        error:
          $ref: '#/components/schemas/Error'
      required:
        - error
    GuardrailSummary:
      type: object
      required:
        - id
        - name
        - slug
        - created_at
        - last_updated_at
        - owner_id
      properties:
        id:
          type: string
          description: Unique identifier of the guardrail
        name:
          type: string
          description: Name of the guardrail
        slug:
          type: string
          description: URL-friendly slug
        organisation_id:
          type: string
          format: uuid
          description: Organisation UUID
        workspace_id:
          type: string
          format: uuid
          nullable: true
          description: Workspace UUID (null for organisation-level guardrails)
        status:
          type: string
          enum:
            - active
            - archived
          description: Current status of the guardrail
        created_at:
          type: string
          format: date-time
          description: Creation timestamp
        last_updated_at:
          type: string
          format: date-time
          description: Last update timestamp
        owner_id:
          type: string
          format: uuid
          description: UUID of the user who created the guardrail
        updated_by:
          type: string
          format: uuid
          nullable: true
          description: UUID of the user who last updated the guardrail
    Error:
      type: object
      properties:
        code:
          type: string
          nullable: true
        message:
          type: string
          nullable: false
        param:
          type: string
          nullable: true
        type:
          type: string
          nullable: false
      required:
        - type
        - message
        - param
        - code
  securitySchemes:
    Portkey-Key:
      type: apiKey
      in: header
      name: x-portkey-api-key

````